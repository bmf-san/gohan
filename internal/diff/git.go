package diff

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bmf-san/gohan/internal/model"
)

// GitDiffEngine implements DiffEngine using file hashes with an optional
// git-accelerated change list.
type GitDiffEngine struct {
	rootDir string
}

// NewGitDiffEngine returns a GitDiffEngine rooted at rootDir.
func NewGitDiffEngine(rootDir string) *GitDiffEngine {
	return &GitDiffEngine{rootDir: rootDir}
}

// Detect compares the current working tree against manifest.
// When manifest is nil, every file under rootDir is returned as Added
// (full-build signal). Otherwise files whose SHA-256 hash has changed are
// returned as Modified, newly added files as Added, and missing files as Deleted.
func (g *GitDiffEngine) Detect(manifest *model.BuildManifest) (*model.ChangeSet, error) {
	current, err := hashAllFiles(g.rootDir)
	if err != nil {
		return nil, err
	}

	if manifest == nil {
		cs := &model.ChangeSet{}
		for path := range current {
			cs.AddedFiles = append(cs.AddedFiles, path)
		}
		return cs, nil
	}

	cs := &model.ChangeSet{}
	for path, hash := range current {
		if prev, ok := manifest.FileHashes[path]; !ok {
			cs.AddedFiles = append(cs.AddedFiles, path)
		} else if prev != hash {
			cs.ModifiedFiles = append(cs.ModifiedFiles, path)
		}
	}
	for path := range manifest.FileHashes {
		if path == configHashKey {
			continue // sentinel key — not a real content file
		}
		if _, ok := current[path]; !ok {
			cs.DeletedFiles = append(cs.DeletedFiles, path)
		}
	}
	return cs, nil
}

// Hash returns the SHA-256 hex digest of the file at filePath.
func (g *GitDiffEngine) Hash(filePath string) (string, error) {
	return hashFile(filePath)
}

// hashAllFiles walks rootDir and returns a map of relative-path → SHA-256 hex.
func hashAllFiles(rootDir string) (map[string]string, error) {
	result := make(map[string]string)
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(rootDir, path)
		h, err := hashFile(path)
		if err != nil {
			return err
		}
		result[rel] = h
		return nil
	})
	return result, err
}

// hashFile returns the SHA-256 hex digest of the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
