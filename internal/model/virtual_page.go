package model

// VirtualPage represents a generated page that does not correspond to a
// Markdown source file. It is produced by SitePlugins during the build
// pipeline and rendered by the HTML generator using a named template.
//
// Example: the bookshelf plugin emits one VirtualPage per locale containing
// all book entries aggregated from every article's front-matter.
type VirtualPage struct {
	// OutputPath is the file path to write relative to the output directory.
	// e.g. "bookshelf/index.html" (default locale) or "ja/bookshelf/index.html"
	OutputPath string

	// URL is the canonical URL path for this page, including trailing slash.
	// e.g. "/bookshelf/" or "/ja/bookshelf/"
	URL string

	// Template is the theme template filename used to render this page.
	// e.g. "bookshelf.html"
	Template string

	// Locale is the locale code for this page (e.g. "en", "ja").
	// Empty when i18n is not configured.
	Locale string

	// Data holds arbitrary page-specific data injected by the SitePlugin.
	// Accessible in templates via .VirtualPageData.
	Data map[string]interface{}
}
