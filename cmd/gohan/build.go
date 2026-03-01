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

	// Load config.
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

	// Print .gitignore hint on first run.
	gohanDir := filepath.Join(rootDir, ".gohan")
	if _, statErr := os.Stat(gohanDir); os.IsNotExist(statErr) {
		fmt.Println("hint: add '.gohan/' to your .gitignore to exclude build cache")
	}

	// Hash config for cache invalidation.
	cfgHasher := diff.NewGitDiffEngine(rootDir)
	configHash, _ := cfgHasher.Hash(cfgAbs)

	// Load manifest.
	manifest, err := diff.ReadManifest(cacheDir)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}

	// Full build when: --full flag, config changed, or no manifest yet.
	forceFullBuild := *full || diff.CheckConfigChange(manifest, configHash)
	if forceFullBuild && manifest != nil {
		if clearErr := diff.ClearCache(cacheDir); clearErr != nil {
			return fmt.Errorf("clear cache: %w", clearErr)
		}
		manifest = nil
	}

	// Parse content.
	p := parser.NewFileParser()
	contentDir := filepath.Join(rootDir, cfg.Build.ContentDir)
	articles, err := p.ParseAll(contentDir)
	if err != nil {
		return fmt.Errorf("parse content: %w", err)
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

	// Build taxonomy.
	taxo, err := proc.BuildTaxonomyRegistry(processed, *cfg)
	if err != nil {
		return fmt.Errorf("build taxonomy: %w", err)
	}

	site := &model.Site{
		Config:     *cfg,
		Articles:   processed,
		Tags:       taxo.Tags,
		Categories: taxo.Categories,
	}

	if *dryRun {
		elapsed := time.Since(start)
		fmt.Printf("dry-run: %d articles, %s\n", len(processed), elapsed.Round(time.Millisecond))
		return nil
	}

	// Render HTML.
	outDir := filepath.Join(rootDir, cfg.Build.OutputDir)
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
	if err := generator.GenerateSitemap(outDir, cfg.Site.BaseURL, processed); err != nil {
		fmt.Fprintf(os.Stderr, "warn: sitemap: %v\n", err)
	}
	if err := generator.GenerateFeeds(outDir, cfg.Site.BaseURL, cfg.Site.Title, processed); err != nil {
		fmt.Fprintf(os.Stderr, "warn: feeds: %v\n", err)
	}

	// Copy assets.
	assetsDir := filepath.Join(rootDir, cfg.Build.AssetsDir)
	if err := generator.CopyAssets(assetsDir, outDir); err != nil {
		fmt.Fprintf(os.Stderr, "warn: copy assets: %v\n", err)
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
