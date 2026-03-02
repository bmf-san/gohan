package model

import "time"

// BuildManifest is the build history persisted to .gohan/cache/manifest.json.
type BuildManifest struct {
	Version      string              `json:"version"`
	BuildTime    time.Time           `json:"build_time"`
	LastCommit   string              `json:"last_commit"`
	FileHashes   map[string]string   `json:"file_hashes"`
	Dependencies map[string][]string `json:"dependencies"`
	OutputFiles  []OutputFile        `json:"output_files"`
}

// OutputFile records metadata for a single generated file.
type OutputFile struct {
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ContentType  string    `json:"content_type"`
}
