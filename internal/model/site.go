package model

// Site holds the full rendering context passed to templates.
type Site struct {
	Config        Config
	Articles      []*ProcessedArticle
	Tags          []Taxonomy
	Categories    []Taxonomy
	Pagination    *Pagination // nil when pagination is disabled or not a listing page
	CurrentLocale string      // locale for the current page; empty when i18n is not configured
}

// Pagination holds computed paging metadata for listing pages.
type Pagination struct {
	CurrentPage int
	TotalPages  int
	PerPage     int
	TotalItems  int
	PrevURL     string // empty string if no previous page
	NextURL     string // empty string if no next page
	BaseURL     string // URL path prefix used to construct PrevURL/NextURL (e.g. "/tags/go")
}

// FileWatcher is the interface for watching file system changes.
type FileWatcher interface {
	Add(path string) error
	Events() <-chan string
	Close() error
}

// DevServer is the local HTTP development server configuration.
type DevServer struct {
	Host    string
	Port    int
	OutDir  string
	Watcher FileWatcher
}
