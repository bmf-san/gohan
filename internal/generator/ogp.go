package generator

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg" // register JPEG decoder
	"image/png"    // also self-registers PNG decoder
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/bmf-san/gohan/internal/model"
)

const (
	ogpDefaultWidth  = 1200
	ogpDefaultHeight = 630
)

// OGPGenerator generates OGP thumbnail images for articles at build time.
type OGPGenerator struct {
	outDir string
	cfg    model.OGPConfig
}

// NewOGPGenerator returns an OGPGenerator configured from cfg.
func NewOGPGenerator(outDir string, cfg model.OGPConfig) *OGPGenerator {
	return &OGPGenerator{outDir: outDir, cfg: cfg}
}

// Generate creates one PNG per article in public/ogp/{slug}.png.
// Articles whose output file already exists are skipped when changeSet is non-nil
// and the article is not in the changed set.
func (g *OGPGenerator) Generate(site *model.Site, changeSet *model.ChangeSet) error {
	if !g.cfg.Enabled {
		return nil
	}

	w := g.cfg.Width
	h := g.cfg.Height
	if w == 0 {
		w = ogpDefaultWidth
	}
	if h == 0 {
		h = ogpDefaultHeight
	}

	bgColor, err := parseHexColor(g.cfg.BackgroundColor)
	if err != nil {
		bgColor = color.RGBA{R: 30, G: 30, B: 46, A: 255}
	}
	textColor, err := parseHexColor(g.cfg.TextColor)
	if err != nil {
		textColor = color.RGBA{R: 205, G: 214, B: 244, A: 255}
	}

	var face font.Face
	if g.cfg.FontFile != "" {
		face, err = loadFontFace(g.cfg.FontFile, 64)
		if err != nil {
			return fmt.Errorf("ogp: load font %q: %w", g.cfg.FontFile, err)
		}
	}

	var logoImg image.Image
	if g.cfg.LogoFile != "" {
		logoImg, err = loadImage(g.cfg.LogoFile)
		if err != nil {
			return fmt.Errorf("ogp: load logo %q: %w", g.cfg.LogoFile, err)
		}
	}

	ogpDir := filepath.Join(g.outDir, "ogp")
	if err := os.MkdirAll(ogpDir, 0o755); err != nil {
		return fmt.Errorf("ogp: mkdir: %w", err)
	}

	changed := changedSet(changeSet)

	for _, a := range site.Articles {
		slug := a.FrontMatter.Slug
		if slug == "" {
			slug = slugify(a.FrontMatter.Title)
		}
		outPath := filepath.Join(ogpDir, slug+".png")

		// Skip if already exists and article not in change set
		if _, statErr := os.Stat(outPath); statErr == nil && changeSet != nil {
			if !changed[a.FilePath] {
				continue
			}
		}

		if err := g.renderImage(outPath, a.FrontMatter.Title, w, h, bgColor, textColor, face, logoImg); err != nil {
			return fmt.Errorf("ogp: render %q: %w", slug, err)
		}
	}
	return nil
}

func (g *OGPGenerator) renderImage(
	outPath, title string,
	w, h int,
	bg, fg color.Color,
	face font.Face,
	logo image.Image,
) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Draw logo (top-left, with padding)
	if logo != nil {
		const logoPad = 40
		bounds := logo.Bounds()
		dstRect := image.Rect(logoPad, logoPad, logoPad+bounds.Dx(), logoPad+bounds.Dy())
		xdraw.BiLinear.Scale(img, dstRect, logo, bounds, draw.Over, nil)
	}

	// Draw title text (centered, word-wrapped) only when a font face is available
	if face != nil {
		drawCenteredText(img, title, face, fg, w, h)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

// drawCenteredText renders word-wrapped text centered in the image.
func drawCenteredText(img *image.RGBA, text string, face font.Face, fg color.Color, w, h int) {
	metrics := face.Metrics()
	lineHeight := metrics.Height.Ceil()
	maxWidth := w - 120 // horizontal padding

	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if measureText(face, candidate) > maxWidth && current != "" {
			lines = append(lines, current)
			current = word
		} else {
			current = candidate
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	totalHeight := len(lines) * lineHeight
	startY := (h - totalHeight) / 2

	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{fg},
		Face: face,
	}
	for i, line := range lines {
		lineW := measureText(face, line)
		x := (w - lineW) / 2
		y := startY + (i+1)*lineHeight
		d.Dot = fixed.P(x, y)
		d.DrawString(line)
	}
}

func measureText(face font.Face, s string) int {
	advance := font.MeasureString(face, s)
	return advance.Ceil()
}

// loadFontFace loads a TTF/OTF file and returns a font.Face at the given size.
func loadFontFace(path string, size float64) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tt, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

// loadImage loads a PNG/JPEG image from path.
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	return img, err
}

// parseHexColor parses a "#rrggbb" string into color.RGBA.
func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color: %q", s)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

// changedSet converts a ChangeSet slice into a map for O(1) lookup.
func changedSet(cs *model.ChangeSet) map[string]bool {
	if cs == nil {
		return nil
	}
	m := make(map[string]bool, len(cs.ModifiedFiles)+len(cs.AddedFiles))
	for _, f := range cs.ModifiedFiles {
		m[f] = true
	}
	for _, f := range cs.AddedFiles {
		m[f] = true
	}
	return m
}
