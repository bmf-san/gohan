package diff

import (
	"github.com/bmf-san/gohan/internal/model"
)

// DiffEngine detects which source files have changed since the last build,
// returning a ChangeSet used by dependent packages to skip unchanged work.
type DiffEngine interface {
	// Detect compares the current working tree against the provided manifest
	// and returns the set of files that are new, modified, or deleted.
	Detect(manifest *model.BuildManifest) (*model.ChangeSet, error)

	// Hash returns the SHA-256 hex digest for the file at filePath.
	Hash(filePath string) (string, error)
}
