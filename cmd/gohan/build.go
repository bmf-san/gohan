package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bmf-san/gohan/internal/config"
	"github.com/bmf-san/gohan/internal/diff"
	"github.com/bmf-san/gohan/internal/generator"
	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/parser"
	"github.com/bmf-san/gohan/internal/plugin"
	"github.com/bmf-san/gohan/internal/processor"
	gohantemplate "github.com/bmf-san/gohan/internal/template"
)

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	full := fs.Bool("full", false, "force full build (bypass diff detection)")
	configPath := fs.String("config", "config.yaml", "path to config file")
	outputDir := fs.String("output", "", "override output directory")
	parallel := fs.Int("parallel", 0, "override parallelism (0 = use config value)")
	dryRun := fs.Bool("dry-run", false, "simulate build without writing files")
	logFmt := fs.String("log-format", "text", "log format: text or json")
	draft := fs.Bool("draft", false, "include draft articles in the build")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_ = logFmt // reserved for structured logging

	start := time.Now()

	// Determine project root from config file location.
	cfgAbs, err := filepath.Abs(*configPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}
	rootDir := filepath.Dir(cfgAbs)

	// Prevent concurrent builds from multiple processes (e.g. `gohan serve`
	// watchLoop rebuild racing with a manual `gohan build`).  We use an
	// exclusive non-blocking flock on .gohan/build.lock.  If the lock is
	// already held by another process we print a notice and skip this run
	// so that public/ is never written by two processes at once.
	gohanDir := filepath.Join(rootDir, ".gohan")
	// Print .gitignore hint on first run (before creating the directory).
	if _, statErr := os.Stat(gohanDir); os.IsNotExist(statErr) {
		fmt.Println("hint: add '.gohan/' to your .gitignore to exclude build cache")
	}
	_ = os.MkdirAll(gohanDir, 0o755)
	lockPath := filepath.Join(gohanDir, "build.lock")
	unlock, acquired := tryLockBuildFile(lockPath)
	if !acquired {
		fmt.Println("build: another build is already running — skipping")
		return nil
	}
	defer unlock()

	loader := config.New(rootDir)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if *outputDir != "" {
		cfg.Build.OutputDir = *outputDir
	}
	if *parallel > 0 {
		cfg.Build.Parallelism = *parallel
	}

	cacheDir := filepath.Join(rootDir, ".gohan", "cache")

	// Hash config for cache invalidation.
	cfgHasher := diff.NewGitDiffEngine(rootDir)
	configHash, configHashErr := cfgHasher.Hash(cfgAbs)

	// Load manifest.
	manifest, err := diff.ReadManifest(cacheDir)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}

	// Full build when: --full flag, config hashing failed, config changed, or no manifest yet.
	// If we cannot hash the config, we must assume it has changed to avoid stale output.
	forceFullBuild := *full || configHashErr != nil || diff.CheckConfigChange(manifest, configHash)
	if forceFullBuild && manifest != nil {
		if clearErr := diff.ClearCache(cacheDir); clearErr != nil {
			return fmt.Errorf("clear cache: %w", clearErr)
		}
		manifest = nil
	}

	// Parse content.
	p := parser.NewFileParser(cfg.Build.ExcludeFiles...)
	contentDir := filepath.Join(rootDir, cfg.Build.ContentDir)
	// Resolve path fields to absolute so that processor functions that call
	// filepath.Rel(cfg.Build.ContentDir, a.FilePath) work correctly when
	// article FilePaths are absolute (as set by the file parser).
	cfg.Build.ContentDir = contentDir
	cfg.Build.OutputDir = filepath.Join(rootDir, cfg.Build.OutputDir)
	cfg.Build.AssetsDir = filepath.Join(rootDir, cfg.Build.AssetsDir)
	if cfg.Build.StaticDir != "" {
		cfg.Build.StaticDir = filepath.Join(rootDir, cfg.Build.StaticDir)
	}
	articles, err := p.ParseAll(contentDir)
	if err != nil {
		return fmt.Errorf("parse content: %w", err)
	}

	// Filter draft articles unless --draft flag is set.
	if !*draft {
		filtered := articles[:0]
		for _, a := range articles {
			if !a.FrontMatter.Draft {
				filtered = append(filtered, a)
			}
		}
		articles = filtered
	}

	// Detect diff.
	var changeSet *model.ChangeSet
	if !forceFullBuild {
		engine := diff.NewGitDiffEngine(contentDir)
		changeSet, err = engine.Detect(manifest)
		if err != nil {
			return fmt.Errorf("detect changes: %w", err)
		}
	}

	// Process articles.
	proc := processor.NewSiteProcessor()
	processed, err := proc.Process(articles, *cfg)
	if err != nil {
		return fmt.Errorf("process articles: %w", err)
	}

	// Link translations across locales (no-op when i18n is not configured).
	proc.BuildTranslationMap(processed)

	// Validate that no two articles resolve to the same output path.
	// Duplicate output paths cause silent page overwrites during HTML generation.
	if errs := processor.ValidateOutputPaths(processed); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "warn: output path: %v\n", e)
		}
	}

	// Build taxonomy registry.
	// When i18n is active, locale-specific files are preferred:
	//   {contentDir}/{locale}/tags.yaml        (falls back to {contentDir}/tags.yaml)
	//   {contentDir}/{locale}/categories.yaml  (falls back to {contentDir}/categories.yaml)
	// When no registry files exist, the registry is derived from article frontmatter
	// (no validation is performed).
	var taxo *model.TaxonomyRegistry
	{
		regs, loadErr := processor.LoadLocaleAwareTaxonomyRegistries(contentDir, cfg.I18n.Locales)
		if loadErr != nil {
			return fmt.Errorf("load taxonomy registries: %w", loadErr)
		}
		merged := processor.MergeTaxonomyRegistries(regs)
		if len(merged.Tags) > 0 || len(merged.Categories) > 0 {
			taxo = merged
			if errs := processor.ValidateArticleTaxonomiesLocale(processed, regs); len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "warn: taxonomy: %v\n", e)
				}
			}
		} else {
			computed, err := proc.BuildTaxonomyRegistry(processed, *cfg)
			if err != nil {
				return fmt.Errorf("build taxonomy: %w", err)
			}
			taxo = computed
		}
	}

	site := &model.Site{
		Config:     *cfg,
		Articles:   processed,
		Tags:       taxo.Tags,
		Categories: taxo.Categories,
	}

	// Run plugins.
	if err := plugin.DefaultRegistry().Enrich(site); err != nil {
		return fmt.Errorf("plugin enrichment: %w", err)
	}

	if *dryRun {
		elapsed := time.Since(start)
		fmt.Printf("dry-run: %d articles, %s\n", len(processed), elapsed.Round(time.Millisecond))
		return nil
	}

	// Render HTML.
	outDir := cfg.Build.OutputDir
	templateDir := filepath.Join(rootDir, cfg.Theme.Dir, "templates")
	tmpl := gohantemplate.NewEngine()
	if loadErr := tmpl.Load(templateDir, nil); loadErr != nil {
		fmt.Fprintf(os.Stderr, "warn: load templates: %v\n", loadErr)
	}
	gen := generator.NewHTMLGenerator(outDir, tmpl, *cfg)
	if err := gen.Generate(site, changeSet); err != nil {
		return fmt.Errorf("generate HTML: %w", err)
	}

	// Sitemap + feeds.
	if err := generator.GenerateSitemap(outDir, cfg.Site.BaseURL, processed, *cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warn: sitemap: %v\n", err)
	}
	if err := generator.GenerateFeeds(outDir, cfg.Site.BaseURL, cfg.Site.Title, processed, *cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warn: feeds: %v\n", err)
	}

	// Update manifest.
	newManifest := diff.NewManifest(configHash)
	hashEngine := diff.NewGitDiffEngine(contentDir)
	for _, a := range articles {
		rel, _ := filepath.Rel(contentDir, a.FilePath)
		if h, herr := hashEngine.Hash(a.FilePath); herr == nil {
			newManifest.FileHashes[rel] = h
		}
	}
	if err := diff.WriteManifest(cacheDir, newManifest); err != nil {
		fmt.Fprintf(os.Stderr, "warn: write manifest: %v\n", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("build: %d articles, 0 errors, %s\n", len(processed), elapsed.Round(time.Millisecond))
	return nil
}
