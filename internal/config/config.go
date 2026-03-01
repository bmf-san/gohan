// Package config loads and validates gohan configuration from YAML files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/bmf-san/gohan/internal/model"
)

const (
	configFileName        = "config.yaml"
	defaultContentDir     = "content"
	defaultOutputDir      = "public"
	defaultAssetsDir      = "assets"
	defaultParallelism    = 4
	defaultThemeName      = "default"
	defaultLanguage       = "en"
	defaultHighlightTheme = "github"
)

// Loader reads and validates the gohan project configuration.
type Loader struct {
	rootDir string
}

// New returns a Loader that reads config.yaml from rootDir.
func New(rootDir string) *Loader {
	return &Loader{rootDir: rootDir}
}

// Load reads config.yaml, applies default values, validates required fields,
// and returns the resolved configuration.
func (l *Loader) Load() (*model.Config, error) {
	path := filepath.Join(l.rootDir, configFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	applyDefaults(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// applyDefaults fills optional fields with sensible default values.
func applyDefaults(cfg *model.Config) {
	if cfg.Build.ContentDir == "" {
		cfg.Build.ContentDir = defaultContentDir
	}
	if cfg.Build.OutputDir == "" {
		cfg.Build.OutputDir = defaultOutputDir
	}
	if cfg.Build.AssetsDir == "" {
		cfg.Build.AssetsDir = defaultAssetsDir
	}
	if cfg.Build.Parallelism <= 0 {
		cfg.Build.Parallelism = defaultParallelism
	}
	if cfg.Theme.Name == "" {
		cfg.Theme.Name = defaultThemeName
	}
	if cfg.Theme.Dir == "" {
		cfg.Theme.Dir = filepath.Join("themes", cfg.Theme.Name)
	}
	if cfg.Site.Language == "" {
		cfg.Site.Language = defaultLanguage
	}
	if cfg.SyntaxHighlight.Theme == "" {
		cfg.SyntaxHighlight.Theme = defaultHighlightTheme
	}
}

// validate checks required fields and returns an error if any are missing.
func validate(cfg *model.Config) error {
	if cfg.Site.Title == "" {
		return errors.New("config: site.title is required")
	}
	if cfg.Site.BaseURL == "" {
		return errors.New("config: site.base_url is required")
	}
	return nil
}
