// Package ogp is a gohan built-in asset plugin that generates Open Graph (OGP)
// thumbnail images for every article at build time.
//
// # Configuration (config.yaml)
//
//	plugins:
//	  ogp:
//	    enabled: true
//	    logo_file: "assets/logo.png"   # optional; top-left logo overlay
//	    width: 1200                    # optional; default 1200
//	    height: 630                    # optional; default 630
//
// Each image is written to <output>/ogp/{slug}.png and is intended to be
// referenced from a theme template via `<meta property="og:image">`. The design
// (a slug-seeded diagonal gradient with geometric decorations) is deterministic
// per article, so on incremental builds unchanged articles are skipped.
package ogp

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg" // register JPEG decoder for logo loading
	"image/png"    // register PNG encoder/decoder
	"math/rand"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"

	"github.com/bmf-san/gohan/internal/model"
)

const (
	// Name is the plugin identifier, used as the key under config.yaml plugins.<name>.
	Name = "ogp"

	defaultWidth  = 1200
	defaultHeight = 630
)

// OGP implements plugin.AssetPlugin: it renders one OGP PNG per article.
type OGP struct{}

// New returns a new OGP asset plugin instance.
func New() *OGP { return &OGP{} }

// compile-time interface check (structural, to avoid importing the plugin
// package and creating an import cycle).
var _ interface {
	Name() string
	Enabled(map[string]interface{}) bool
	GenerateAssets(*model.Site, string, *model.ChangeSet, map[string]interface{}) error
} = (*OGP)(nil)

// Name returns the plugin identifier.
func (o *OGP) Name() string { return Name }

// Enabled returns true when the plugin config has enabled: true.
func (o *OGP) Enabled(cfg map[string]interface{}) bool {
	return boolVal(cfg, "enabled")
}

// GenerateAssets writes <outDir>/ogp/{slug}.png for every article. On incremental
// builds (changeSet non-nil) an article whose image already exists and is not in
// the changed set is skipped.
func (o *OGP) GenerateAssets(site *model.Site, outDir string, changeSet *model.ChangeSet, cfg map[string]interface{}) error {
	w := intVal(cfg, "width", defaultWidth)
	h := intVal(cfg, "height", defaultHeight)
	if w <= 0 {
		w = defaultWidth
	}
	if h <= 0 {
		h = defaultHeight
	}
	logoFile := strVal(cfg, "logo_file", "")

	// contentDir lets us convert absolute FilePaths to changeSet-relative paths.
	contentDir := site.Config.Build.ContentDir

	var logoImg image.Image
	if logoFile != "" {
		var err error
		logoImg, err = loadImage(logoFile)
		if err != nil {
			return fmt.Errorf("ogp: load logo %q: %w", logoFile, err)
		}
	}

	ogpDir := filepath.Join(outDir, "ogp")
	if err := os.MkdirAll(ogpDir, 0o755); err != nil {
		return fmt.Errorf("ogp: mkdir: %w", err)
	}

	changed := changedSet(changeSet)

	for _, a := range site.Articles {
		// Sanitize slug to prevent path traversal (slugify strips dots, slashes, etc.).
		slug := slugify(a.FrontMatter.Slug)
		if slug == "untitled" {
			slug = slugify(a.FrontMatter.Title)
		}
		outPath := filepath.Join(ogpDir, slug+".png")

		// Skip if already exists and article not in change set. changeSet entries
		// are relative to contentDir, but a.FilePath is absolute.
		if _, statErr := os.Stat(outPath); statErr == nil && changeSet != nil {
			lookupPath := a.FilePath
			if contentDir != "" {
				if rel, relErr := filepath.Rel(contentDir, a.FilePath); relErr == nil {
					lookupPath = rel
				}
			}
			if !changed[lookupPath] {
				continue
			}
		}

		if err := renderImage(outPath, slug, w, h, logoImg); err != nil {
			return fmt.Errorf("ogp: render %q: %w", slug, err)
		}
	}
	return nil
}

func renderImage(outPath, slug string, w, h int, logo image.Image) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	seed := ogpHash(slug)
	drawGradientBackground(img, seed, w, h)
	drawGeometricShapes(img, seed, w, h)

	// Draw logo (top-left, with padding).
	if logo != nil {
		const logoPad = 40
		bounds := logo.Bounds()
		dstRect := image.Rect(logoPad, logoPad, logoPad+bounds.Dx(), logoPad+bounds.Dy())
		xdraw.BiLinear.Scale(img, dstRect, logo, bounds, draw.Over, nil)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("ogp: encode %q: %w", slug, err)
	}
	return writeFileAtomic(outPath, buf.Bytes(), 0o644)
}

// drawGradientBackground fills img with a diagonal two-color gradient whose hues
// are derived deterministically from the article slug hash.
func drawGradientBackground(img *image.RGBA, seed uint64, w, h int) {
	h1 := float64(seed % 360)
	offset := float64(80 + (seed>>16)%100) // 80–180 degree hue offset for contrast
	h2 := h1 + offset
	if h2 >= 360 {
		h2 -= 360
	}
	c1 := hsvToRGBA(h1, 0.60, 0.22, 255)
	c2 := hsvToRGBA(h2, 0.55, 0.14, 255)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t := (float64(x)/float64(w) + float64(y)/float64(h)) / 2
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
				G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
				B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
				A: 255,
			})
		}
	}
}

// drawGeometricShapes overlays semi-transparent filled circles and accent rings
// whose positions, sizes and hues are deterministically derived from seed.
func drawGeometricShapes(img *image.RGBA, seed uint64, w, h int) {
	rng := rand.New(rand.NewSource(int64(seed))) //nolint:gosec

	// Large filled circles (very low alpha).
	for i := 0; i < 5; i++ {
		cx := rng.Intn(w)
		cy := rng.Intn(h)
		r := w/7 + rng.Intn(w/4)
		hue := float64((seed + uint64(i)*73) % 360)
		c := hsvToRGBA(hue, 0.75, 0.90, 28)
		drawFilledCircle(img, cx, cy, r, c)
	}

	// Accent rings (higher alpha).
	for i := 0; i < 3; i++ {
		cx := rng.Intn(w)
		cy := rng.Intn(h)
		r := w/9 + rng.Intn(w/5)
		hue := float64((seed + uint64(i+5)*97) % 360)
		c := hsvToRGBA(hue, 0.80, 0.95, 65)
		drawRing(img, cx, cy, r, 4, c)
	}
}

// ogpHash returns a deterministic 64-bit FNV-1a hash of s.
func ogpHash(s string) uint64 {
	hh := fnv.New64a()
	_, _ = hh.Write([]byte(s))
	return hh.Sum64()
}

// hsvToRGBA converts HSV (h∈[0,360), s∈[0,1], v∈[0,1]) to color.RGBA with the
// given alpha.
func hsvToRGBA(h, s, v float64, a uint8) color.RGBA {
	if h >= 360 {
		h -= 360
	}
	i := int(h / 60)
	f := h/60 - float64(i)
	p := v * (1 - s)
	q := v * (1 - s*f)
	tt := v * (1 - s*(1-f))
	var r, g, b float64
	switch i {
	case 0:
		r, g, b = v, tt, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, tt
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = tt, p, v
	default:
		r, g, b = v, p, q
	}
	return color.RGBA{R: uint8(r * 255), G: uint8(g * 255), B: uint8(b * 255), A: a}
}

// blendPixel alpha-composites c over the pixel at (x, y). Out-of-bounds pixels
// are silently ignored.
func blendPixel(img *image.RGBA, x, y int, c color.RGBA) {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}
	dst := img.RGBAAt(x, y)
	a := float64(c.A) / 255.0
	ia := 1 - a
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(c.R)*a + float64(dst.R)*ia),
		G: uint8(float64(c.G)*a + float64(dst.G)*ia),
		B: uint8(float64(c.B)*a + float64(dst.B)*ia),
		A: 255,
	})
}

// drawFilledCircle draws a solid filled circle using alpha compositing.
func drawFilledCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r {
				blendPixel(img, x, y, c)
			}
		}
	}
}

// drawRing draws a ring (annulus) of the given pixel thickness using alpha
// compositing.
func drawRing(img *image.RGBA, cx, cy, r, thickness int, c color.RGBA) {
	inner := r - thickness
	if inner < 0 {
		inner = 0
	}
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			dist2 := dx*dx + dy*dy
			if dist2 <= r*r && dist2 >= inner*inner {
				blendPixel(img, x, y, c)
			}
		}
	}
}

// loadImage loads a PNG/JPEG image from path.
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path) //nolint:gosec // path comes from trusted config
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	return img, err
}

// changedSet converts a ChangeSet into a set for O(1) lookup.
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

// slugify converts s to a lowercase hyphen-separated URL slug, keeping only
// ASCII letters, digits and hyphens. Returns "untitled" for an empty result.
func slugify(s string) string {
	var out []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			out = append(out, c+32)
		case c == ' ' || c == '_':
			out = append(out, '-')
		case (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-':
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return "untitled"
	}
	return string(out)
}

// writeFileAtomic writes data to path atomically via a temp file + rename.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".gohan-tmp-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil {
		_ = os.Remove(tmpName)
		return werr
	}
	if cerr != nil {
		_ = os.Remove(tmpName)
		return cerr
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// --- config helpers ---

func boolVal(cfg map[string]interface{}, key string) bool {
	if v, ok := cfg[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func strVal(cfg map[string]interface{}, key, def string) string {
	if v, ok := cfg[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func intVal(cfg map[string]interface{}, key string, def int) int {
	v, ok := cfg[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return def
	}
}
