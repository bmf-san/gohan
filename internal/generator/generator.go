// Package generator writes rendered HTML, assets, sitemap, and Atom feed.
package generator

import (
	"github.com/bmf-san/gohan/internal/model"
)

// OutputGenerator takes the fully-rendered site data and writes all output
// files to the configured output directory.
type OutputGenerator interface {
	// Generate writes all HTML pages and copies static assets into outDir.
	// Only files in changeSet (or all files when changeSet is nil) are written.
	Generate(site *model.Site, changeSet *model.ChangeSet) error

	// GenerateSitemap creates sitemap.xml inside outDir.
	GenerateSitemap(site *model.Site) error

	// GenerateFeed creates atom.xml (Atom feed) inside outDir.
	GenerateFeed(site *model.Site) error
}
