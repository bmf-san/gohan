package diff

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

const (
	cacheManifestFile = "manifest.json"
	cacheHTMLDir      = "html"
	configHashKey     = "__config__"
	manifestVersion   = "1"
)

// ReadManifest loads the BuildManifest from cacheDir/manifest.json.
// Returns nil (no error) when the file does not exist yet.
func ReadManifest(cacheDir string) (*model.BuildManifest, error) {
	path := filepath.Join(cacheDir, cacheManifestFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m model.BuildManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %w", err)
	}
	return &m, nil
}

// WriteManifest persists m to cacheDir/manifest.json, creating directories as
// needed.
func WriteManifest(cacheDir string, m *model.BuildManifest) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	path := filepath.Join(cacheDir, cacheManifestFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

// ReadCachedHTML returns the cached HTML for slug from cacheDir/html/<slug>.html.
// Returns ("", nil) when not present.
func ReadCachedHTML(cacheDir, slug string) (string, error) {
	path := filepath.Join(cacheDir, cacheHTMLDir, slug+".html")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read cached html: %w", err)
	}
	return string(data), nil
}

// WriteCachedHTML stores html under cacheDir/html/<slug>.html.
func WriteCachedHTML(cacheDir, slug, html string) error {
	dir := filepath.Join(cacheDir, cacheHTMLDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir html cache: %w", err)
	}
	path := filepath.Join(dir, slug+".html")
	if err := os.WriteFile(path, []byte(html), 0644); err != nil {
		return fmt.Errorf("write cached html: %w", err)
	}
	return nil
}

// ClearCache removes all files under cacheDir.
func ClearCache(cacheDir string) error {
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("clear cache: %w", err)
	}
	return nil
}

// CheckConfigChange returns true when the hash stored for configHashKey in
// manifest differs from currentConfigHash.  A nil manifest is treated as
// changed (first build).
func CheckConfigChange(manifest *model.BuildManifest, currentConfigHash string) bool {
	if manifest == nil || manifest.FileHashes == nil {
		return true
	}
	stored, ok := manifest.FileHashes[configHashKey]
	return !ok || stored != currentConfigHash
}

// NewManifest returns a fresh BuildManifest stamped with currentConfigHash.
func NewManifest(configHash string) *model.BuildManifest {
	return &model.BuildManifest{
		Version:    manifestVersion,
		BuildTime:  time.Now().UTC(),
		FileHashes: map[string]string{configHashKey: configHash},
	}
}
