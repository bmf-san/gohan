// Package generator writes rendered HTML, assets, sitemap, and Atom feed.
package generator

import (
	"github.com/bmf-san/gohan/internal/model"
)

// OutputGenerator takes the fully-rendered site data and writes all output
// files to the configured output directory.
//
// Sitemap and feed generation are handled by the package-level GenerateSitemap
// and GenerateFeeds functions, which are i18n-aware and kept separate from
// the HTML generation step.
type OutputGenerator interface {
	// Generate writes all HTML pages, copies static assets, and generates OGP
	// images into outDir.  Only files in changeSet (or all files when changeSet
	// is nil) are written.
	Generate(site *model.Site, changeSet *model.ChangeSet) error
}
