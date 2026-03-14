package model

// Site holds the full rendering context passed to templates.
type Site struct {
	Config             Config
	Articles           []*ProcessedArticle
	Tags               []Taxonomy
	Categories         []Taxonomy
	ArchiveYears       []int               // unique years that have articles, sorted newest-first
	Pagination         *Pagination         // nil when pagination is disabled or not a listing page
	CurrentLocale      string              // locale for the current page; empty when i18n is not configured
	RelatedArticles    []*ProcessedArticle // articles sharing at least one category with the current article (article pages only)
	CurrentTaxonomy      *Taxonomy           // set on tag and category listing pages; nil elsewhere
	CurrentArchivePath   string              // set on archive pages; locale-aware path e.g. "/archives/2024/01/" or "/ja/archives/2024/01/"
	CurrentArchiveIsMonth bool               // true for month archives (/archives/2024/01/), false for year archives (/archives/2024/)
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
