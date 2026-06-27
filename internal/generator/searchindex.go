package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

// searchIndex is the top-level JSON document written to search-index.json.
// It is consumed by client-side search implemented in the theme.
type searchIndex struct {
	// Generated is the RFC3339 timestamp the index was written.
	Generated string `json:"generated"`
	// Count is the number of article entries in this index.
	Count int `json:"count"`
	// Articles holds one searchable record per article, newest-first.
	Articles []searchIndexEntry `json:"articles"`
}

// searchIndexEntry is a single searchable article record.
// Only metadata is included (no full body text) to keep the index small.
type searchIndexEntry struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Date        string   `json:"date,omitempty"`
	Locale      string   `json:"locale,omitempty"`
}

// GenerateSearchIndex writes search-index.json to outDir for client-side search.
//
// Each entry holds article metadata only (title, URL, description, summary,
// tags, categories, date, locale) — no full body text — so the index stays
// small even for large sites. Articles are sorted newest-first and the URL is
// built the same way as feed/sitemap links (absolute when baseURL is set).
// baseURL must not have a trailing slash.
//
// When cfg has I18n.Locales configured, a per-locale index is written for each
// non-default locale at {locale}/search-index.json, and the root
// search-index.json contains only default-locale articles. Without i18n the
// root index contains every article.
func GenerateSearchIndex(outDir, baseURL string, articles []*model.ProcessedArticle, cfg model.Config) error {
	sorted := make([]*model.ProcessedArticle, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FrontMatter.Date.After(sorted[j].FrontMatter.Date)
	})

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// When i18n is active, filter the root index to the default locale only and
	// write per-locale indexes under their locale subdirectory.
	if len(cfg.I18n.Locales) > 0 {
		rootArticles := filterFeedArticles(sorted, cfg.I18n.DefaultLocale)
		if err := writeSearchIndex(filepath.Join(outDir, "search-index.json"), baseURL, rootArticles); err != nil {
			return err
		}
		for _, loc := range cfg.I18n.Locales {
			if loc == cfg.I18n.DefaultLocale {
				continue // already written at root
			}
			locDir := filepath.Join(outDir, loc)
			if err := os.MkdirAll(locDir, 0o755); err != nil {
				return err
			}
			locArticles := filterFeedArticles(sorted, loc)
			if err := writeSearchIndex(filepath.Join(locDir, "search-index.json"), baseURL, locArticles); err != nil {
				return err
			}
		}
		return nil
	}

	return writeSearchIndex(filepath.Join(outDir, "search-index.json"), baseURL, sorted)
}

// writeSearchIndex marshals articles into the search-index.json document at path.
func writeSearchIndex(path, baseURL string, articles []*model.ProcessedArticle) error {
	idx := searchIndex{
		Generated: time.Now().UTC().Format(time.RFC3339),
		Count:     len(articles),
		Articles:  make([]searchIndexEntry, 0, len(articles)),
	}
	for _, a := range articles {
		entry := searchIndexEntry{
			Title:       a.FrontMatter.Title,
			URL:         articleLink(baseURL, a),
			Description: a.FrontMatter.Description,
			Summary:     a.Summary,
			Tags:        a.FrontMatter.Tags,
			Categories:  a.FrontMatter.Categories,
			Locale:      a.Locale,
		}
		if !a.FrontMatter.Date.IsZero() {
			entry.Date = a.FrontMatter.Date.UTC().Format(time.RFC3339)
		}
		idx.Articles = append(idx.Articles, entry)
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data, 0o644)
}
