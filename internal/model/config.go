package model

// Config is the top-level structure of config.yaml.
type Config struct {
	Site            SiteConfig             `yaml:"site"`
	Build           BuildConfig            `yaml:"build"`
	Theme           ThemeConfig            `yaml:"theme"`
	SyntaxHighlight SyntaxHighlightConfig  `yaml:"syntax_highlight"`
	OGP             OGPConfig              `yaml:"ogp"`
	Plugins         map[string]interface{} `yaml:"plugins"`
	I18n            I18nConfig             `yaml:"i18n"`
}

// SiteConfig holds site-wide metadata.
type SiteConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	BaseURL     string `yaml:"base_url"`
	Language    string `yaml:"language"`
	// GitHubRepo is the base URL of the GitHub repository that holds the site
	// source (e.g. "https://github.com/owner/repo"). When set, templates can
	// render an "Edit this page" link using .ContentPath.
	GitHubRepo string `yaml:"github_repo"`
	// GitHubBranch is the branch used to build the edit URL. Defaults to "main".
	GitHubBranch string `yaml:"github_branch"`
}

// BuildConfig holds build-time directory and parallelism settings.
type BuildConfig struct {
	ContentDir   string   `yaml:"content_dir"`
	OutputDir    string   `yaml:"output_dir"`
	AssetsDir    string   `yaml:"assets_dir"`
	ExcludeFiles []string `yaml:"exclude_files"`
	Parallelism  int      `yaml:"parallelism"`
	PerPage      int      `yaml:"per_page"`
}

// ThemeConfig holds theme name, directory, and custom parameters.
type ThemeConfig struct {
	Name   string            `yaml:"name"`
	Dir    string            `yaml:"dir"`
	Params map[string]string `yaml:"params"`
}

// SyntaxHighlightConfig holds settings for code-block syntax highlighting.
type SyntaxHighlightConfig struct {
	// Theme is a chroma style name (e.g. "github", "monokai", "dracula").
	Theme string `yaml:"theme"`
	// LineNumbers enables line number display when true.
	LineNumbers bool `yaml:"line_numbers"`
}

// OGPConfig holds settings for build-time OGP image generation.
type OGPConfig struct {
	Enabled         bool   `yaml:"enabled"`
	BackgroundColor string `yaml:"background_color"`
	TextColor       string `yaml:"text_color"`
	FontFile        string `yaml:"font_file"`
	LogoFile        string `yaml:"logo_file"` // empty means no logo
	Width           int    `yaml:"width"`
	Height          int    `yaml:"height"`
}

// I18nConfig holds multi-language content configuration.
type I18nConfig struct {
	// Locales is the ordered list of locale codes present under the content
	// directory (e.g. ["en", "ja"]). When empty, i18n is disabled and gohan
	// behaves as a single-language site with no URL changes.
	Locales []string `yaml:"locales"`
	// DefaultLocale is the locale code served at the root URL (without a
	// language prefix). Defaults to Site.Language when Locales is non-empty.
	// Example: if DefaultLocale is "en", /posts/hello/ is English and
	// /ja/posts/hello/ is Japanese.
	DefaultLocale string `yaml:"default_locale"`
}
