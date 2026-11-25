package glossary_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bmf-san/gohan/internal/model"
	"github.com/bmf-san/gohan/internal/plugin/glossary"
)

func enabled() map[string]interface{} {
	return map[string]interface{}{"enabled": true}
}

func site(articles ...*model.ProcessedArticle) *model.Site {
	return &model.Site{
		Config: model.Config{
			Site: model.SiteConfig{Language: "en"},
			I18n: model.I18nConfig{DefaultLocale: "en", Locales: []string{"en", "ja"}},
		},
		Articles: articles,
	}
}

func day(d int) time.Time { return time.Date(2026, 1, d, 0, 0, 0, 0, time.UTC) }

// article builds an article whose front-matter declares the given glossary entries.
func article(locale, title, url string, date time.Time, entries ...map[string]interface{}) *model.ProcessedArticle {
	items := make([]interface{}, len(entries))
	for i, e := range entries {
		items[i] = e
	}
	return &model.ProcessedArticle{
		Article: model.Article{FrontMatter: model.FrontMatter{
			Title: title,
			Date:  date,
			Extra: map[string]interface{}{"glossary": items},
		}},
		Locale: locale,
		URL:    url,
	}
}

func term(name, definition string) map[string]interface{} {
	return map[string]interface{}{"term": name, "definition": definition}
}

func termWith(name, reading, definition string) map[string]interface{} {
	return map[string]interface{}{"term": name, "reading": reading, "definition": definition}
}

func pageFor(t *testing.T, pages []*model.VirtualPage, locale string) *model.VirtualPage {
	t.Helper()
	for _, p := range pages {
		if p.Locale == locale {
			return p
		}
	}
	t.Fatalf("no page for locale %q", locale)
	return nil
}

func termsOf(t *testing.T, p *model.VirtualPage) []glossary.Term {
	t.Helper()
	terms, ok := p.Data["terms"].([]glossary.Term)
	if !ok {
		t.Fatalf("Data[\"terms\"] has type %T", p.Data["terms"])
	}
	return terms
}

func groupsOf(t *testing.T, p *model.VirtualPage) []glossary.Group {
	t.Helper()
	groups, ok := p.Data["groups"].([]glossary.Group)
	if !ok {
		t.Fatalf("Data[\"groups\"] has type %T", p.Data["groups"])
	}
	return groups
}

func names(terms []glossary.Term) []string {
	out := make([]string, len(terms))
	for i, tm := range terms {
		out[i] = tm.Term
	}
	return out
}

func TestGlossary_Name(t *testing.T) {
	if got := glossary.New().Name(); got != "glossary" {
		t.Errorf("Name() = %q, want %q", got, "glossary")
	}
}

func TestGlossary_Enabled(t *testing.T) {
	g := glossary.New()
	if !g.Enabled(enabled()) {
		t.Error("Enabled should return true when enabled: true")
	}
	if g.Enabled(map[string]interface{}{"enabled": false}) {
		t.Error("Enabled should return false when enabled: false")
	}
	if g.Enabled(map[string]interface{}{"enabled": "yes"}) {
		t.Error("Enabled should return false for a non-bool value")
	}
	if g.Enabled(map[string]interface{}{}) {
		t.Error("Enabled should return false when the key is absent")
	}
}

func TestGlossary_NoEntries(t *testing.T) {
	g := glossary.New()
	empty := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Empty list", Extra: map[string]interface{}{"glossary": []interface{}{}},
	}}, Locale: "en"}
	null := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Null", Extra: map[string]interface{}{"glossary": nil},
	}}, Locale: "en"}
	none := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{Title: "None"}}, Locale: "en"}
	s := site(empty, null, none)

	pages, err := g.VirtualPages(s, enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	if len(pages) != 0 {
		t.Errorf("expected no pages, got %d", len(pages))
	}
	data, err := g.SiteData(s, enabled())
	if err != nil {
		t.Fatalf("SiteData: %v", err)
	}
	if data != nil {
		t.Errorf("expected nil SiteData, got %#v", data)
	}
}

func TestGlossary_PagesPerLocale(t *testing.T) {
	g := glossary.New()
	s := site(
		article("en", "Authorization", "/posts/authz/", day(2), term("PEP", "Policy enforcement point.")),
		article("ja", "認可とは", "/ja/posts/authz/", day(2), term("PEP", "認可の判定を強制する構成要素。")),
	)

	pages, err := g.VirtualPages(s, enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	if len(pages) != 2 || pages[0].Locale != "en" || pages[1].Locale != "ja" {
		t.Fatalf("expected pages for en and ja in that order, got %d pages", len(pages))
	}

	en, ja := pages[0], pages[1]
	if en.URL != "/glossary/" || en.OutputPath != "glossary/index.html" {
		t.Errorf("en page: URL %q, OutputPath %q", en.URL, en.OutputPath)
	}
	if ja.URL != "/ja/glossary/" || ja.OutputPath != "ja/glossary/index.html" {
		t.Errorf("ja page: URL %q, OutputPath %q", ja.URL, ja.OutputPath)
	}
	for _, p := range pages {
		if p.Template != "glossary.html" {
			t.Errorf("%s page: Template %q", p.Locale, p.Template)
		}
		if p.Data["url"] != p.URL {
			t.Errorf("%s page: Data[url] %v, want %q", p.Locale, p.Data["url"], p.URL)
		}
	}
	if got := en.Data["translations"]; !reflect.DeepEqual(got, map[string]string{"ja": "/ja/glossary/"}) {
		t.Errorf("en translations = %#v", got)
	}
	if got := ja.Data["translations"]; !reflect.DeepEqual(got, map[string]string{"en": "/glossary/"}) {
		t.Errorf("ja translations = %#v", got)
	}

	jaTerms := termsOf(t, ja)
	if len(jaTerms) != 1 || jaTerms[0].Definition != "認可の判定を強制する構成要素。" {
		t.Fatalf("ja terms = %#v", jaTerms)
	}
	if got := jaTerms[0].Articles; len(got) != 1 || got[0].URL != "/ja/posts/authz/" || got[0].Title != "認可とは" {
		t.Errorf("ja term articles = %#v", got)
	}
}

func TestGlossary_SingleLanguageSite(t *testing.T) {
	g := glossary.New()
	s := &model.Site{
		Config:   model.Config{Site: model.SiteConfig{Language: "en"}},
		Articles: []*model.ProcessedArticle{article("", "Post", "/posts/p/", day(1), term("API", "Application programming interface."))},
	}
	pages, err := g.VirtualPages(s, enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	if len(pages) != 1 || pages[0].URL != "/glossary/" || pages[0].OutputPath != "glossary/index.html" {
		t.Fatalf("unexpected pages: %#v", pages)
	}
	if got := pages[0].Data["translations"]; !reflect.DeepEqual(got, map[string]string{}) {
		t.Errorf("translations = %#v, want empty", got)
	}
}

func TestGlossary_CustomPath(t *testing.T) {
	g := glossary.New()
	s := site(
		article("en", "A", "/posts/a/", day(1), term("API", "x")),
		article("ja", "A", "/ja/posts/a/", day(1), term("API", "x")),
	)
	for _, p := range []string{"terms", "/terms/", " terms "} {
		cfg := map[string]interface{}{"enabled": true, "path": p}
		pages, err := g.VirtualPages(s, cfg)
		if err != nil {
			t.Fatalf("VirtualPages: %v", err)
		}
		if en := pageFor(t, pages, "en"); en.URL != "/terms/" || en.OutputPath != "terms/index.html" {
			t.Errorf("path %q: en URL %q, OutputPath %q", p, en.URL, en.OutputPath)
		}
		if ja := pageFor(t, pages, "ja"); ja.URL != "/ja/terms/" || ja.OutputPath != "ja/terms/index.html" {
			t.Errorf("path %q: ja URL %q, OutputPath %q", p, ja.URL, ja.OutputPath)
		}
	}
}

func TestGlossary_SortingAndGroups(t *testing.T) {
	g := glossary.New()
	s := site(article("ja", "用語", "/ja/posts/terms/", day(1),
		termWith("認可", "にんか", "アクセスの可否を決めること。"),
		term("API", "アプリケーションの接続口。"),
		termWith("OAuth", "オーオース", "認可の委譲の仕組み。"),
		term("ロードアベレージ", "実行待ちのプロセス数の平均。"),
		term("ガベージコレクション", "不要なメモリの自動解放。"),
		term("２FA", "二要素認証。"),
		term("認証", "本人であることの確認。"),
		term("Ｇｏ", "プログラミング言語。"),
		term(".NET", "開発プラットフォーム。"),
		term("api gateway", "APIの入り口。"),
		term("あいまい検索", "曖昧な語での検索。"),
	))

	pages, err := g.VirtualPages(s, enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	ja := pageFor(t, pages, "ja")

	wantOrder := []string{"２FA", "API", "api gateway", "Ｇｏ", ".NET", "OAuth", "あいまい検索", "ガベージコレクション", "認可", "ロードアベレージ", "認証"}
	if got := names(termsOf(t, ja)); !reflect.DeepEqual(got, wantOrder) {
		t.Errorf("term order\n got: %v\nwant: %v", got, wantOrder)
	}

	type grp struct {
		key, id string
		terms   []string
	}
	want := []grp{
		{"0-9", "glossary-0-9", []string{"２FA"}},
		{"A", "glossary-a", []string{"API", "api gateway"}},
		{"G", "glossary-g", []string{"Ｇｏ"}},
		{"N", "glossary-n", []string{".NET"}},
		{"O", "glossary-o", []string{"OAuth"}},
		{"あ", "glossary-kana-a", []string{"あいまい検索"}},
		{"か", "glossary-kana-ka", []string{"ガベージコレクション"}},
		{"な", "glossary-kana-na", []string{"認可"}},
		{"ら", "glossary-kana-ra", []string{"ロードアベレージ"}},
		{"#", "glossary-other", []string{"認証"}},
	}
	groups := groupsOf(t, ja)
	if len(groups) != len(want) {
		t.Fatalf("got %d groups, want %d: %#v", len(groups), len(want), groups)
	}
	for i, w := range want {
		if groups[i].Key != w.key || groups[i].ID != w.id || !reflect.DeepEqual(names(groups[i].Terms), w.terms) {
			t.Errorf("group %d = {%q %q %v}, want {%q %q %v}", i, groups[i].Key, groups[i].ID, names(groups[i].Terms), w.key, w.id, w.terms)
		}
	}
}

func TestGlossary_KanaRows(t *testing.T) {
	g := glossary.New()
	rows := map[string]string{
		"ウィンドウ": "あ", "キャッシュ": "か", "ジッター": "さ", "ツリー": "た", "ヌル": "な",
		"パーティション": "は", "ミドルウェア": "ま", "ユニーク制約": "や", "リトライ": "ら", "ワーカー": "わ",
	}
	var entries []map[string]interface{}
	for name := range rows {
		entries = append(entries, term(name, "定義。"))
	}
	pages, err := g.VirtualPages(site(article("ja", "用語", "/ja/posts/t/", day(1), entries...)), enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	for _, grp := range groupsOf(t, pageFor(t, pages, "ja")) {
		for _, tm := range grp.Terms {
			if rows[tm.Term] != grp.Key {
				t.Errorf("%s is in group %q, want %q", tm.Term, grp.Key, rows[tm.Term])
			}
		}
	}
}

func TestGlossary_MergesDuplicates(t *testing.T) {
	g := glossary.New()
	older := article("ja", "古い記事", "/ja/posts/old/", day(1),
		map[string]interface{}{"term": "pep", "reading": "ぺっぷ", "definition": "古い定義。", "id": "pep"})
	newer := article("ja", "新しい記事", "/ja/posts/new/", day(5),
		term("PEP", "新しい定義。"),
		term("ＰＥＰ", "同じ記事での重複。"))

	pages, err := g.VirtualPages(site(older, newer), enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	terms := termsOf(t, pageFor(t, pages, "ja"))
	if len(terms) != 1 {
		t.Fatalf("expected one merged term, got %v", names(terms))
	}
	got := terms[0]
	if got.Term != "PEP" || got.Definition != "新しい定義。" {
		t.Errorf("merged term = %q / %q, want the newest article's term and definition", got.Term, got.Definition)
	}
	if got.Reading != "ぺっぷ" || got.ID != "pep" {
		t.Errorf("reading %q, id %q: missing fields should be filled from older articles", got.Reading, got.ID)
	}
	if len(got.Articles) != 2 || got.Articles[0].URL != "/ja/posts/new/" || got.Articles[1].URL != "/ja/posts/old/" {
		t.Errorf("articles = %#v, want newest first without duplicates", got.Articles)
	}
}

func TestGlossary_IDs(t *testing.T) {
	g := glossary.New()
	s := site(article("ja", "用語", "/ja/posts/t/", day(1),
		term("OAuth 2.0", "x"),
		term("認可", "x"),
		term("Sagaパターン", "x"),
		term("C++", "x"),
		term("C", "x"),
		map[string]interface{}{"term": "Custom", "definition": "x", "id": "term-c"},
		term("#", "x"),
	))
	pages, err := g.VirtualPages(s, enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	ids := map[string]string{}
	seen := map[string]bool{}
	for _, tm := range termsOf(t, pageFor(t, pages, "ja")) {
		ids[tm.Term] = tm.ID
		if seen[tm.ID] {
			t.Errorf("duplicate id %q", tm.ID)
		}
		seen[tm.ID] = true
	}
	for name, want := range map[string]string{
		"OAuth 2.0": "term-oauth-2-0",
		"認可":        "term-認可",
		"Sagaパターン":  "term-sagaパターン",
		"Custom":    "term-c",
		"C":         "term-c-2",
		"C++":       "term-c-3",
	} {
		if ids[name] != want {
			t.Errorf("id of %q = %q, want %q", name, ids[name], want)
		}
	}
	if id := ids["#"]; !strings.HasPrefix(id, "term-") || len(id) != len("term-")+8 {
		t.Errorf("id of %q = %q, want a hashed fallback", "#", id)
	}
}

func TestGlossary_ScalarValues(t *testing.T) {
	g := glossary.New()
	s := site(article("en", "Errors", "/posts/errors/", day(1),
		map[string]interface{}{"term": 404, "definition": "Not Found."}))
	pages, err := g.VirtualPages(s, enabled())
	if err != nil {
		t.Fatalf("VirtualPages: %v", err)
	}
	if got := names(termsOf(t, pageFor(t, pages, "en"))); !reflect.DeepEqual(got, []string{"404"}) {
		t.Errorf("terms = %v", got)
	}
}

func TestGlossary_InvalidFrontMatter(t *testing.T) {
	g := glossary.New()
	notList := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Bad list", Extra: map[string]interface{}{"glossary": "PEP"},
	}}, Locale: "en"}
	notMap := &model.ProcessedArticle{Article: model.Article{FrontMatter: model.FrontMatter{
		Title: "Bad entry", Extra: map[string]interface{}{"glossary": []interface{}{"PEP"}},
	}}, Locale: "en"}

	cases := []struct {
		name    string
		article *model.ProcessedArticle
		want    string
	}{
		{"not a sequence", notList, `"glossary" in article "Bad list" must be a YAML sequence`},
		{"entry not a mapping", notMap, `entry 1 in article "Bad entry" must be a mapping`},
		{"missing definition", article("en", "No definition", "/posts/x/", day(1), map[string]interface{}{"term": "PEP"}), `entry 1 in article "No definition" needs both term and definition`},
		{"missing term", article("en", "No term", "/posts/y/", day(1), term("ok", "fine"), map[string]interface{}{"definition": "x"}), `entry 2 in article "No term" needs both term and definition`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := g.VirtualPages(site(tc.article), enabled())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("VirtualPages error = %v, want it to contain %q", err, tc.want)
			}
			if _, err := g.SiteData(site(tc.article), enabled()); err == nil {
				t.Error("SiteData should return the same error")
			}
		})
	}
}

func TestGlossary_SiteData(t *testing.T) {
	g := glossary.New()
	s := site(
		article("en", "A", "/posts/a/", day(1), term("API", "x"), term("PEP", "y")),
		article("ja", "A", "/ja/posts/a/", day(1), term("API", "x")),
	)
	data, err := g.SiteData(s, map[string]interface{}{"enabled": true, "path": "terms"})
	if err != nil {
		t.Fatalf("SiteData: %v", err)
	}
	m, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("SiteData type %T", data)
	}
	if got := m["urls"]; !reflect.DeepEqual(got, map[string]string{"en": "/terms/", "ja": "/ja/terms/"}) {
		t.Errorf("urls = %#v", got)
	}
	if got := m["counts"]; !reflect.DeepEqual(got, map[string]int{"en": 2, "ja": 1}) {
		t.Errorf("counts = %#v", got)
	}
}

func TestGlossary_SiteDataArticles(t *testing.T) {
	g := glossary.New()
	s := site(
		article("en", "Old", "/posts/old/", day(1), term("PEP", "old"), term("API", "x")),
		article("en", "New", "/posts/new/", day(2), term("pep", "new")),
		article("ja", "A", "/ja/posts/a/", day(1), term("認可", "z")),
	)
	data, err := g.SiteData(s, enabled())
	if err != nil {
		t.Fatalf("SiteData: %v", err)
	}
	got := data.(map[string]interface{})["articles"]
	want := map[string][]glossary.TermLink{
		"/posts/old/":  {{Term: "API", URL: "/glossary/#term-api"}, {Term: "pep", URL: "/glossary/#term-pep"}},
		"/posts/new/":  {{Term: "pep", URL: "/glossary/#term-pep"}},
		"/ja/posts/a/": {{Term: "認可", URL: "/ja/glossary/#term-認可"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("articles = %#v", got)
	}
}
