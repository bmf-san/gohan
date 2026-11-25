// Package glossary is a gohan built-in SitePlugin that generates one glossary
// page per locale by aggregating term definitions from every article's
// front-matter.
//
// Each term links back to the articles that define it, so the glossary works
// as an index into in-depth articles instead of a set of thin per-term pages.
//
// # Configuration (config.yaml)
//
//	plugins:
//	  glossary:
//	    enabled: true
//	    path: "glossary"   # optional URL segment (default "glossary")
//
// # Front-matter (article .md)
//
//	glossary:
//	  - term: "PEP"
//	    reading: "ペップ"   # optional; pronunciation, also the sort key for non-Latin terms
//	    definition: "A component that enforces authorization decisions."
//	    id: "pep"          # optional; anchor id on the glossary page
//
// # Template usage (glossary.html)
//
//	{{range index .VirtualPageData "groups"}}
//	  <h2 id="{{.ID}}">{{.Key}}</h2>
//	  <dl>
//	    {{range .Terms}}
//	      <dt id="{{.ID}}">{{.Term}}{{if .Reading}} ({{.Reading}}){{end}}</dt>
//	      <dd>
//	        {{.Definition}}
//	        {{range .Articles}}<a href="{{.URL}}">{{.Title}}</a>{{end}}
//	      </dd>
//	    {{end}}
//	  </dl>
//	{{end}}
//
// # Navigation link (any template)
//
// The plugin also exposes the glossary URL per locale on every page via
// Site.SiteData (see SiteDataProvider), so themes can show a link only when a
// glossary exists for the current locale:
//
//	{{with index .SiteData "glossary"}}{{with index (index . "urls") $.CurrentLocale}}
//	  <a href="{{.}}">Glossary</a>
//	{{end}}{{end}}
package glossary

import (
	"fmt"
	"hash/fnv"
	"path"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/bmf-san/gohan/internal/model"
)

const (
	// Name is the plugin identifier used as the key in config.yaml and registry.
	Name = "glossary"

	// frontMatterKey is the article front-matter key that holds glossary entries.
	frontMatterKey = "glossary"

	defaultPath  = "glossary"
	templateName = "glossary.html"

	// derivedIDPrefix keeps generated anchors from colliding with ids that the
	// theme itself uses on the page.
	derivedIDPrefix = "term-"
	groupIDPrefix   = "glossary-"
)

// ArticleRef identifies an article that defines a term.
type ArticleRef struct {
	Title string
	URL   string
	Date  time.Time
}

// Term is a single glossary entry exposed to templates.
type Term struct {
	// ID is the anchor id of the entry on the glossary page.
	ID string
	// Term is the display name of the term.
	Term string
	// Reading is the optional pronunciation, such as kana for a Japanese term.
	Reading string
	// Definition is a short plain-text definition.
	Definition string
	// Articles lists the articles that define the term, newest first.
	Articles []ArticleRef
}

// Group is a set of terms that share an index heading.
type Group struct {
	// Key is the heading label: "0-9", "A" to "Z", a kana row such as "あ" or
	// "か", or "#" for terms that start with anything else.
	Key string
	// ID is an anchor-safe id for the heading, e.g. "glossary-a" or
	// "glossary-kana-ka".
	ID string
	// Terms are the entries in this group, in display order.
	Terms []Term
}

// Glossary implements SitePlugin and SiteDataProvider.
type Glossary struct{}

// New returns a new Glossary plugin instance.
func New() *Glossary { return &Glossary{} }

// compile-time interface check
var _ interface {
	Name() string
	Enabled(map[string]interface{}) bool
	VirtualPages(*model.Site, map[string]interface{}) ([]*model.VirtualPage, error)
	SiteData(*model.Site, map[string]interface{}) (interface{}, error)
} = (*Glossary)(nil)

// Name returns the plugin identifier.
func (g *Glossary) Name() string { return Name }

// Enabled returns true when the plugin config has enabled: true.
func (g *Glossary) Enabled(cfg map[string]interface{}) bool {
	v, ok := cfg["enabled"]
	if !ok {
		return false
	}
	bv, ok := v.(bool)
	return ok && bv
}

// VirtualPages collects the glossary entries from article front-matter and
// returns one VirtualPage per locale that has at least one entry. Each page's
// Data map contains:
//
//	"terms"        []Term             — all entries in display order
//	"groups"       []Group            — entries grouped by index heading
//	"url"          string             — URL of this glossary page
//	"translations" map[string]string  — locale → URL of the glossary in other locales
func (g *Glossary) VirtualPages(site *model.Site, cfg map[string]interface{}) ([]*model.VirtualPage, error) {
	byLocale, err := collect(site)
	if err != nil {
		return nil, err
	}
	if len(byLocale) == 0 {
		return nil, nil
	}

	urls := pageURLs(site, cfg, byLocale)
	locales := make([]string, 0, len(byLocale))
	for locale := range byLocale {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	pages := make([]*model.VirtualPage, 0, len(locales))
	for _, locale := range locales {
		entries := byLocale[locale]
		pageURL := urls[locale]

		translations := make(map[string]string, len(urls)-1)
		for l, u := range urls {
			if l != locale {
				translations[l] = u
			}
		}

		pages = append(pages, &model.VirtualPage{
			OutputPath: path.Join(strings.TrimPrefix(pageURL, "/"), "index.html"),
			URL:        pageURL,
			Template:   templateName,
			Locale:     locale,
			Data: map[string]interface{}{
				"terms":        termsOf(entries),
				"groups":       groupsOf(entries),
				"url":          pageURL,
				"translations": translations,
			},
		})
	}
	return pages, nil
}

// SiteData implements plugin.SiteDataProvider. It exposes the glossary page
// URL and the number of terms per locale on every rendered page, so themes can
// link to the glossary only when one exists for the current locale.
//
// The returned value has the shape:
//
//	{ "urls": { "<locale>": "/glossary/" }, "counts": { "<locale>": 42 } }
//
// Returns nil when no article declares glossary entries.
func (g *Glossary) SiteData(site *model.Site, cfg map[string]interface{}) (interface{}, error) {
	byLocale, err := collect(site)
	if err != nil {
		return nil, err
	}
	if len(byLocale) == 0 {
		return nil, nil
	}

	counts := make(map[string]int, len(byLocale))
	for locale, entries := range byLocale {
		counts[locale] = len(entries)
	}
	return map[string]interface{}{
		"urls":   pageURLs(site, cfg, byLocale),
		"counts": counts,
	}, nil
}

// entry is a Term plus the keys used to order and group it.
type entry struct {
	term       Term
	explicitID bool
	sortKey    string
	group      int
}

// collect aggregates glossary entries from every article's front-matter,
// grouped by locale and sorted into display order. Entries whose terms match
// after normalization are merged: the newest article supplies the definition,
// and every defining article is kept in Articles (newest first). Locales
// without any entry are omitted.
func collect(site *model.Site) (map[string][]*entry, error) {
	articles := make([]*model.ProcessedArticle, len(site.Articles))
	copy(articles, site.Articles)
	sort.SliceStable(articles, func(i, j int) bool {
		di, dj := articles[i].FrontMatter.Date, articles[j].FrontMatter.Date
		if !di.Equal(dj) {
			return di.After(dj)
		}
		return articles[i].URL < articles[j].URL
	})

	type localeSet struct {
		byKey map[string]*entry
		order []*entry
	}
	sets := map[string]*localeSet{}

	for _, a := range articles {
		raw, ok := a.FrontMatter.Extra[frontMatterKey]
		if !ok || raw == nil {
			continue
		}
		items, ok := raw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("glossary: %q in article %q must be a YAML sequence", frontMatterKey, a.FrontMatter.Title)
		}

		ls, ok := sets[a.Locale]
		if !ok {
			ls = &localeSet{byKey: map[string]*entry{}}
			sets[a.Locale] = ls
		}
		ref := ArticleRef{Title: a.FrontMatter.Title, URL: a.URL, Date: a.FrontMatter.Date}

		for i, item := range items {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("glossary: entry %d in article %q must be a mapping", i+1, a.FrontMatter.Title)
			}
			name := scalar(m, "term")
			def := scalar(m, "definition")
			if name == "" || def == "" {
				return nil, fmt.Errorf("glossary: entry %d in article %q needs both term and definition", i+1, a.FrontMatter.Title)
			}
			reading := scalar(m, "reading")
			id := scalar(m, "id")

			key := matchKey(name)
			if e, ok := ls.byKey[key]; ok {
				if e.term.Reading == "" {
					e.term.Reading = reading
				}
				if !e.explicitID && id != "" {
					e.term.ID = id
					e.explicitID = true
				}
				if !hasArticle(e.term.Articles, ref) {
					e.term.Articles = append(e.term.Articles, ref)
				}
				continue
			}
			e := &entry{
				term:       Term{ID: id, Term: name, Reading: reading, Definition: def, Articles: []ArticleRef{ref}},
				explicitID: id != "",
			}
			ls.byKey[key] = e
			ls.order = append(ls.order, e)
		}
	}

	out := make(map[string][]*entry, len(sets))
	for locale, ls := range sets {
		if len(ls.order) == 0 {
			continue
		}
		for _, e := range ls.order {
			e.sortKey, e.group = sortKeyAndGroup(e.term.Term, e.term.Reading)
		}
		sort.SliceStable(ls.order, func(i, j int) bool {
			a, b := ls.order[i], ls.order[j]
			if a.group != b.group {
				return a.group < b.group
			}
			if a.sortKey != b.sortKey {
				return a.sortKey < b.sortKey
			}
			return a.term.Term < b.term.Term
		})
		assignIDs(ls.order)
		out[locale] = ls.order
	}
	return out, nil
}

// termsOf returns the Terms of entries in their current order.
func termsOf(entries []*entry) []Term {
	terms := make([]Term, len(entries))
	for i, e := range entries {
		terms[i] = e.term
	}
	return terms
}

// groupsOf splits entries, already in display order, into index groups.
func groupsOf(entries []*entry) []Group {
	var groups []Group
	for _, e := range entries {
		if n := len(groups); n == 0 || groups[n-1].Key != groupKey(e.group) {
			groups = append(groups, Group{Key: groupKey(e.group), ID: groupID(e.group)})
		}
		last := &groups[len(groups)-1]
		last.Terms = append(last.Terms, e.term)
	}
	return groups
}

// pageURLs returns the glossary page URL for every locale in byLocale.
func pageURLs(site *model.Site, cfg map[string]interface{}, byLocale map[string][]*entry) map[string]string {
	segment := strings.Trim(path.Clean("/"+scalar(cfg, "path")), "/")
	if segment == "" {
		segment = defaultPath
	}
	defaultLocale := site.Config.I18n.DefaultLocale
	if defaultLocale == "" {
		defaultLocale = site.Config.Site.Language
	}

	urls := make(map[string]string, len(byLocale))
	for locale := range byLocale {
		if locale == defaultLocale || locale == "" {
			urls[locale] = "/" + segment + "/"
		} else {
			urls[locale] = "/" + locale + "/" + segment + "/"
		}
	}
	return urls
}

// assignIDs gives every entry a unique anchor id. Explicit ids are applied
// first so they keep their exact value; derived ids get a "-2", "-3", …
// suffix when they collide.
func assignIDs(entries []*entry) {
	used := map[string]bool{}
	unique := func(base string) string {
		id := base
		for n := 2; used[id]; n++ {
			id = fmt.Sprintf("%s-%d", base, n)
		}
		used[id] = true
		return id
	}
	for _, e := range entries {
		if e.explicitID {
			e.term.ID = unique(e.term.ID)
		}
	}
	for _, e := range entries {
		if !e.explicitID {
			e.term.ID = unique(derivedIDPrefix + slug(e.term.Term))
		}
	}
}

// slug turns a term into an anchor-friendly string. Latin letters are
// lowercased, other letters and digits (e.g. kana and kanji) are kept, and
// everything else collapses into single hyphens. Terms with no letters or
// digits fall back to a short hash.
func slug(term string) string {
	var b strings.Builder
	pendingDash := false
	for _, r := range term {
		r = toHalfWidth(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingDash && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingDash = false
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		pendingDash = true
	}
	if b.Len() == 0 {
		h := fnv.New32a()
		_, _ = h.Write([]byte(term))
		return fmt.Sprintf("%08x", h.Sum32())
	}
	return b.String()
}

// Group indexes, in display order: digits, A–Z, the kana rows, then others.
const (
	groupDigits    = 0
	groupFirstLtr  = 1
	groupFirstKana = groupFirstLtr + 26
)

// kanaRows lists the gojuon rows used as index headings for Japanese terms.
// Small kana and voiced or semi-voiced kana belong to their base row.
var kanaRows = []struct{ key, id, chars string }{
	{"あ", "a", "ぁあぃいぅうぇえぉおゔ"},
	{"か", "ka", "かがきぎくぐけげこごゕゖ"},
	{"さ", "sa", "さざしじすずせぜそぞ"},
	{"た", "ta", "ただちぢっつづてでとど"},
	{"な", "na", "なにぬねの"},
	{"は", "ha", "はばぱひびぴふぶぷへべぺほぼぽ"},
	{"ま", "ma", "まみむめも"},
	{"や", "ya", "ゃやゅゆょよ"},
	{"ら", "ra", "らりるれろ"},
	{"わ", "wa", "ゎわゐゑをん"},
}

var groupOther = groupFirstKana + len(kanaRows)

// sortKeyAndGroup returns the key a term sorts by and its index group.
// Terms that start with a Latin letter, a digit, or kana sort by the term
// itself, so "OAuth" stays under "O" even when it has a kana reading. Other
// terms, such as kanji, sort by their reading when one is given.
func sortKeyAndGroup(term, reading string) (string, int) {
	src := term
	r, ok := firstSignificant(normalizeSort(term))
	startsLatinOrKana := ok && (isASCIIAlnum(r) || isHiragana(r))
	if !startsLatinOrKana && reading != "" {
		src = reading
	}
	key := normalizeSort(src)
	r, ok = firstSignificant(key)
	if !ok {
		return key, groupOther
	}
	key = key[strings.IndexRune(key, r):]

	switch {
	case r >= '0' && r <= '9':
		return key, groupDigits
	case r >= 'a' && r <= 'z':
		return key, groupFirstLtr + int(r-'a')
	case isHiragana(r):
		for i, row := range kanaRows {
			if strings.ContainsRune(row.chars, r) {
				return key, groupFirstKana + i
			}
		}
	}
	return key, groupOther
}

// groupKey returns the heading label for a group index.
func groupKey(group int) string {
	switch {
	case group == groupDigits:
		return "0-9"
	case group >= groupFirstLtr && group < groupFirstKana:
		return string(rune('A' + group - groupFirstLtr))
	case group >= groupFirstKana && group < groupOther:
		return kanaRows[group-groupFirstKana].key
	default:
		return "#"
	}
}

// groupID returns the anchor id for a group index.
func groupID(group int) string {
	switch {
	case group == groupDigits:
		return groupIDPrefix + "0-9"
	case group >= groupFirstLtr && group < groupFirstKana:
		return groupIDPrefix + string(rune('a'+group-groupFirstLtr))
	case group >= groupFirstKana && group < groupOther:
		return groupIDPrefix + "kana-" + kanaRows[group-groupFirstKana].id
	default:
		return groupIDPrefix + "other"
	}
}

// matchKey normalizes a term for merging duplicates: full-width ASCII becomes
// half-width, letters are lowercased, and runs of white space collapse.
func matchKey(term string) string {
	var b strings.Builder
	for _, r := range term {
		b.WriteRune(unicode.ToLower(toHalfWidth(r)))
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// normalizeSort normalizes text for ordering: full-width ASCII becomes
// half-width, katakana becomes hiragana, and letters are lowercased.
func normalizeSort(s string) string {
	var b strings.Builder
	for _, r := range s {
		r = toHalfWidth(r)
		if r >= 0x30A1 && r <= 0x30F6 { // katakana → hiragana
			r -= 0x60
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return strings.TrimSpace(b.String())
}

// firstSignificant returns the first letter or digit in s.
func firstSignificant(s string) (rune, bool) {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r, true
		}
	}
	return 0, false
}

func toHalfWidth(r rune) rune {
	switch {
	case r >= 0xFF01 && r <= 0xFF5E:
		return r - 0xFEE0
	case r == 0x3000:
		return ' '
	}
	return r
}

func isASCIIAlnum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func isHiragana(r rune) bool { return r >= 0x3041 && r <= 0x3096 }

func hasArticle(refs []ArticleRef, ref ArticleRef) bool {
	for _, r := range refs {
		if r.URL == ref.URL && r.Title == ref.Title {
			return true
		}
	}
	return false
}

// scalar returns m[key] as trimmed text. Strings are returned as-is; numbers
// and booleans (e.g. a term written as 404) are formatted; anything else, or a
// missing key, yields "".
func scalar(m map[string]interface{}, key string) string {
	switch v := m[key].(type) {
	case string:
		return strings.TrimSpace(v)
	case int, int64, float64, bool:
		return fmt.Sprint(v)
	default:
		return ""
	}
}
