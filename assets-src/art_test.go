package assetssrc_test

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// drawingPxPerWorldPx is the Art Bible's drawing scale: 8 px in a drawing per
// world pixel.
const drawingPxPerWorldPx = 8

// worldSize is the size of a drawing's page in world pixels.
type worldSize struct{ w, h int }

// manifest lists every drawing under assets-src/tiles and assets-src/sprites,
// keyed by its slash path relative to assets-src, with its page size in world
// pixels. A drawing missing from the folders, or a drawing in the folders but
// not listed here, fails TestArt_EveryDrawingFollowsTheArtBible.
var manifest = map[string]worldSize{
	"tiles/tileset_outdoor.svg": {128, 128},
}

// artFolders are the folders under assets-src whose drawings `make art` exports.
var artFolders = []string{"tiles", "sprites"}

// minStrokeWidth is half a world pixel in drawing px: nothing may be thinner.
const minStrokeWidth = 4

var colourProperties = map[string]bool{
	"fill":           true,
	"stroke":         true,
	"stop-color":     true,
	"flood-color":    true,
	"lighting-color": true,
	"color":          true,
}

// bannedElements are SVG elements the Art Bible rules out, with the reason.
var bannedElements = map[string]string{
	"lineargradient": "gradients are not allowed",
	"radialgradient": "gradients are not allowed",
	"meshgradient":   "gradients are not allowed",
	"filter":         "filters and blur are not allowed",
	"image":          "raster <image> elements are not allowed",
	"text":           "<text> elements are not allowed; convert text to paths",
}

var hexColour = regexp.MustCompile(`^#(?:[0-9a-f]{6}|[0-9a-f]{3})$`)

func TestArt_EveryDrawingFollowsTheArtBible(t *testing.T) {
	palette := loadPalette(t, "palettes/endesga_32.gpl")
	for _, p := range lintArt(t, ".", manifest, palette) {
		t.Error(p)
	}
}

func TestArt_LintRejectsDrawingsThatBreakTheArtBible(t *testing.T) {
	palette := loadPalette(t, "palettes/endesga_32.gpl")
	tile := "tiles/tiles.svg"
	sprite := "sprites/props/prop.svg"
	tileOnly := map[string]worldSize{tile: {16, 16}}
	spriteOnly := map[string]worldSize{sprite: {16, 16}}

	tests := []struct {
		name     string
		manifest map[string]worldSize
		files    map[string]string
		want     string // a substring of the single problem expected; "" for none
	}{
		{
			name:     "a tile and a prop that follow the rules pass",
			manifest: map[string]worldSize{tile: {16, 16}, sprite: {16, 16}},
			files: map[string]string{
				tile:   page(128, 128, `<rect width="128" height="128" style="fill:#63c74d;stroke:none"/>`),
				sprite: page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#265c42" stroke-width="8"/>`),
			},
		},
		{
			name:     "an off-palette fill attribute is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<rect width="8" height="8" fill="#123456"/>`)},
			want:     `fill="#123456": not in the Endesga 32 palette`,
		},
		{
			name:     "an off-palette fill in a style declaration is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<rect width="8" height="8" style="stroke-width:8;fill:#00ff00"/>`)},
			want:     `fill="#00ff00": not in the Endesga 32 palette`,
		},
		{
			name:     "a named colour is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<rect width="8" height="8" fill="red"/>`)},
			want:     `fill="red": want a #rrggbb palette colour`,
		},
		{
			name:     "a linear gradient is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<defs><linearGradient id="g"/></defs>`)},
			want:     "<linearGradient>: gradients are not allowed",
		},
		{
			name:     "a radial gradient is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<defs><radialGradient id="g"/></defs>`)},
			want:     "<radialGradient>: gradients are not allowed",
		},
		{
			name:     "a filter is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<defs><filter id="blur"/></defs>`)},
			want:     "<filter>: filters and blur are not allowed",
		},
		{
			name:     "a CSS blur in a style declaration is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<rect width="8" height="8" style="filter:blur(4px)"/>`)},
			want:     `filter="blur(4px)": filters and blur are not allowed`,
		},
		{
			name:     "an embedded raster image is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<image href="photo.png" width="8" height="8"/>`)},
			want:     "<image>: raster <image> elements are not allowed",
		},
		{
			name:     "a text element is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<text>hi</text>`)},
			want:     "<text>: <text> elements are not allowed",
		},
		{
			name:     "a page whose width is not its manifest size at 8 px per world pixel is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(16, 16, ``)},
			want:     `page is width="16" height="16" viewBox="0 0 16 16", want width="128" height="128" viewBox="0 0 128 128"`,
		},
		{
			name:     "a page whose viewBox differs from its size is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: `<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128" viewBox="0 0 64 64"></svg>`},
			want:     `viewBox="0 0 64 64", want width="128" height="128" viewBox="0 0 128 128"`,
		},
		{
			name:     "a drawing not in the manifest is rejected",
			manifest: map[string]worldSize{},
			files:    map[string]string{"sprites/decor/stray.svg": page(128, 128, ``)},
			want:     "sprites/decor/stray.svg: drawing is not in the manifest",
		},
		{
			name:     "a manifest entry with no drawing is rejected",
			manifest: map[string]worldSize{"sprites/props/ghost.svg": {16, 16}},
			files:    map[string]string{},
			want:     "sprites/props/ghost.svg: in the manifest but no drawing exists",
		},
		{
			name:     "a stroke attribute on a ground tileset is rejected",
			manifest: tileOnly,
			files:    map[string]string{tile: page(128, 128, `<rect width="128" height="128" fill="#63c74d" stroke="#3e8948" stroke-width="8"/>`)},
			want:     `stroke="#3e8948": ground tiles have no outline`,
		},
		{
			name:     "a stroke in a style declaration on a ground tileset is rejected",
			manifest: tileOnly,
			files:    map[string]string{tile: page(128, 128, `<rect width="128" height="128" style="fill:#63c74d;stroke:#3e8948;stroke-width:8"/>`)},
			want:     `stroke="#3e8948": ground tiles have no outline`,
		},
		{
			name:     "a stroke-width attribute under 4 px is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#265c42" stroke-width="2"/>`)},
			want:     `stroke-width="2" is thinner than 4 px`,
		},
		{
			name:     "a stroke-width under 4 px in a style declaration is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<circle cx="64" cy="64" r="40" style="fill:#3e8948;stroke:#265c42;stroke-width:3.5px"/>`)},
			want:     `stroke-width="3.5px" is thinner than 4 px`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for rel, content := range tc.files {
				p := filepath.Join(root, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			problems := lintArt(t, root, tc.manifest, palette)

			if tc.want == "" {
				if len(problems) != 0 {
					t.Fatalf("problems = %q, want none", problems)
				}
				return
			}
			if len(problems) != 1 || !strings.Contains(problems[0], tc.want) {
				t.Fatalf("problems = %q, want exactly one containing %q", problems, tc.want)
			}
		})
	}
}

// page returns an SVG document with the given page size around body.
func page(w, h int, body string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">%s</svg>`, w, h, w, h, body)
}

// lintArt checks every drawing under root/tiles and root/sprites against the
// manifest and the machine-checkable Art Bible rules, and returns one line per
// problem found.
func lintArt(t *testing.T, root string, manifest map[string]worldSize, palette map[string]bool) []string {
	t.Helper()
	var problems []string
	found := make(map[string]bool)
	for _, folder := range artFolders {
		dir := filepath.Join(root, folder)
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if p == dir && errors.Is(err, fs.ErrNotExist) {
					return filepath.SkipDir
				}
				return err
			}
			if d.IsDir() || !strings.EqualFold(filepath.Ext(p), ".svg") {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			found[rel] = true
			size, ok := manifest[rel]
			if !ok {
				problems = append(problems, rel+": drawing is not in the manifest")
				return nil
			}
			for _, prob := range lintDrawing(p, rel, size, palette) {
				problems = append(problems, rel+": "+prob)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	for rel := range manifest {
		if !found[rel] {
			problems = append(problems, rel+": in the manifest but no drawing exists")
		}
	}
	slices.Sort(problems)
	return problems
}

// lintDrawing checks one SVG drawing and returns its problems.
func lintDrawing(path, rel string, size worldSize, palette map[string]bool) []string {
	f, err := os.Open(path)
	if err != nil {
		return []string{err.Error()}
	}
	defer f.Close()

	ground := strings.HasPrefix(rel, "tiles/")
	var problems []string
	dec := xml.NewDecoder(f)
	sawRoot := false
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return append(problems, "not well-formed XML: "+err.Error())
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		elem := start.Name.Local
		if !sawRoot {
			sawRoot = true
			if elem != "svg" {
				return append(problems, fmt.Sprintf("root element is <%s>, want <svg>", elem))
			}
			if prob := checkPage(start, size); prob != "" {
				problems = append(problems, prob)
			}
		}
		if reason, banned := bannedElements[strings.ToLower(elem)]; banned {
			problems = append(problems, fmt.Sprintf("<%s>: %s", elem, reason))
		}
		for _, prop := range properties(start) {
			problems = append(problems, checkProperty(elem, prop.name, prop.value, ground, palette)...)
		}
	}
	if !sawRoot {
		problems = append(problems, "no <svg> root element")
	}
	return problems
}

// checkPage returns a problem when the root element's page is not size at
// drawingPxPerWorldPx, or "" when it is.
func checkPage(root xml.StartElement, size worldSize) string {
	w, h := size.w*drawingPxPerWorldPx, size.h*drawingPxPerWorldPx
	var width, height, viewBox string
	for _, a := range root.Attr {
		switch a.Name.Local {
		case "width":
			width = a.Value
		case "height":
			height = a.Value
		case "viewBox":
			viewBox = a.Value
		}
	}
	wantBox := []float64{0, 0, float64(w), float64(h)}
	if pagePx(width) == float64(w) && pagePx(height) == float64(h) && slices.Equal(numbers(viewBox), wantBox) {
		return ""
	}
	return fmt.Sprintf(`page is width=%q height=%q viewBox=%q, want width="%d" height="%d" viewBox="0 0 %d %d"`,
		width, height, viewBox, w, h, w, h)
}

// pagePx parses a page dimension in px ("128" or "128px"), or returns -1.
func pagePx(v string) float64 {
	n, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(v), "px"), 64)
	if err != nil {
		return -1
	}
	return n
}

func numbers(v string) []float64 {
	var out []float64
	for _, field := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' }) {
		n, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil
		}
		out = append(out, n)
	}
	return out
}

type property struct{ name, value string }

// properties returns an element's presentation attributes and the
// declarations of its style attribute, with lower-case names.
func properties(start xml.StartElement) []property {
	var props []property
	for _, a := range start.Attr {
		name := strings.ToLower(a.Name.Local)
		if a.Name.Space != "" && name != "style" {
			// Namespaced attributes such as inkscape:label are not presentation attributes.
			continue
		}
		if name != "style" {
			props = append(props, property{name, strings.TrimSpace(a.Value)})
			continue
		}
		for _, decl := range strings.Split(a.Value, ";") {
			prop, val, ok := strings.Cut(decl, ":")
			if !ok {
				continue
			}
			props = append(props, property{strings.ToLower(strings.TrimSpace(prop)), strings.TrimSpace(val)})
		}
	}
	return props
}

// checkProperty returns the problems with one presentation property.
func checkProperty(elem, name, value string, ground bool, palette map[string]bool) []string {
	var problems []string
	at := fmt.Sprintf("<%s> %s=%q", elem, name, value)
	if colourProperties[name] {
		if prob := checkColour(value, palette); prob != "" {
			problems = append(problems, at+": "+prob)
		}
	}
	switch name {
	case "stroke":
		if ground && value != "none" {
			problems = append(problems, at+": ground tiles have no outline; use stroke none")
		}
	case "stroke-width":
		n, err := strconv.ParseFloat(strings.TrimSuffix(value, "px"), 64)
		switch {
		case err != nil:
			problems = append(problems, at+": want a width in px")
		case n < minStrokeWidth:
			problems = append(problems, fmt.Sprintf("%s is thinner than %d px (half a world pixel)", at, minStrokeWidth))
		}
	case "filter":
		if value != "none" {
			problems = append(problems, at+": filters and blur are not allowed")
		}
	}
	return problems
}

// checkColour returns a problem with a colour value, or "" when it is "none",
// a url(#...) reference or a palette colour.
func checkColour(value string, palette map[string]bool) string {
	v := strings.ToLower(value)
	if v == "none" || (strings.HasPrefix(v, "url(#") && strings.HasSuffix(v, ")")) {
		return ""
	}
	if !hexColour.MatchString(v) {
		return `want a #rrggbb palette colour, "none" or "url(#...)"`
	}
	if len(v) == 4 {
		v = string([]byte{'#', v[1], v[1], v[2], v[2], v[3], v[3]})
	}
	if !palette[v] {
		return "not in the Endesga 32 palette"
	}
	return ""
}

// loadPalette reads the hex colours of a GIMP palette file.
func loadPalette(t *testing.T, path string) map[string]bool {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open palette: %v", err)
	}
	defer f.Close()

	colours := make(map[string]bool)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Split(strings.TrimSpace(sc.Text()), "\t")
		if len(fields) < 4 {
			continue
		}
		colours["#"+strings.ToLower(strings.TrimSpace(fields[3]))] = true
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read palette: %v", err)
	}
	if len(colours) != 32 {
		t.Fatalf("palette %s has %d colours, want 32", path, len(colours))
	}
	return colours
}
