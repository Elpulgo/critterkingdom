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
	"sprites/decor/clover.svg":           {16, 16},
	"sprites/decor/flower_pink.svg":      {16, 16},
	"sprites/decor/flower_red.svg":       {16, 16},
	"sprites/decor/flower_white.svg":     {16, 16},
	"sprites/decor/flower_yellow.svg":    {16, 16},
	"sprites/decor/leaf_1.svg":           {16, 16},
	"sprites/decor/leaf_2.svg":           {16, 16},
	"sprites/decor/leaf_3.svg":           {16, 16},
	"sprites/decor/pebble_1.svg":         {16, 16},
	"sprites/decor/pebble_2.svg":         {16, 16},
	"sprites/decor/pebble_3.svg":         {16, 16},
	"sprites/decor/tuft_1.svg":           {16, 16},
	"sprites/decor/tuft_2.svg":           {16, 16},
	"sprites/decor/tuft_3.svg":           {16, 16},
	"sprites/decor/twig.svg":             {16, 16},
	"sprites/props/berry_bush_empty.svg": {16, 16},
	"sprites/props/berry_bush_full.svg":  {16, 16},
	"sprites/props/bush.svg":             {16, 16},
	"sprites/props/fence_gate.svg":       {16, 16},
	"sprites/props/fence_h.svg":          {16, 16},
	"sprites/props/fence_post.svg":       {16, 16},
	"sprites/props/fence_v.svg":          {16, 16},
	"sprites/props/house.svg":            {64, 64},
	"sprites/props/rock.svg":             {16, 16},
	"sprites/props/rock_small.svg":       {16, 16},
	"sprites/props/sign.svg":             {16, 16},
	"sprites/props/tree.svg":             {32, 48},
	"sprites/props/tree_small.svg":       {32, 32},
	"tiles/tileset_cave.svg":             {128, 16},
	"tiles/tileset_indoor.svg":           {128, 16},
	"tiles/tileset_outdoor.svg":          {128, 128},
}

// artFolders are the folders under assets-src whose drawings `make art` exports.
var artFolders = []string{"tiles", "sprites"}

// minStrokeWidth is half a world pixel in drawing px: nothing may be thinner.
const minStrokeWidth = 4

// propsFolder holds the prop drawings, which need an outline of outlineWidth
// (one world pixel) with round joins and caps, never in ink or black.
const (
	propsFolder  = "sprites/props/"
	outlineWidth = 1 * drawingPxPerWorldPx
)

// bannedOutlines are the stroke colours a prop is never outlined in: the
// palette's ink and black.
var bannedOutlines = map[string]bool{"#181425": true, "#000000": true, "#000": true, "black": true}

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
	sprite := "sprites/decor/decor.svg"
	prop := "sprites/props/prop.svg"
	tileOnly := map[string]worldSize{tile: {16, 16}}
	spriteOnly := map[string]worldSize{sprite: {16, 16}}
	propOnly := map[string]worldSize{prop: {16, 16}}
	// round sets the round joins and caps every prop stroke needs.
	const round = `stroke-linejoin="round" stroke-linecap="round"`

	tests := []struct {
		name     string
		manifest map[string]worldSize
		files    map[string]string
		want     string // a substring of the single problem expected; "" for none
	}{
		{
			name:     "a tile, a decoration and a prop that follow the rules pass",
			manifest: map[string]worldSize{tile: {16, 16}, sprite: {16, 16}, prop: {16, 16}},
			files: map[string]string{
				tile:   page(128, 128, `<rect width="128" height="128" style="fill:#63c74d;stroke:none"/>`),
				sprite: page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#265c42" stroke-width="8"/>`),
				prop:   page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#265c42" stroke-width="8" `+round+`/>`),
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
			name:     "a stroked drawing with no stroke-width anywhere gets the 1 px default and is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#265c42"/>`)},
			want:     `<circle> stroke="#265c42" has no stroke-width on it or a parent, so it is SVG's default 1 px`,
		},
		{
			name:     "a stroke inherited from a parent group with no stroke-width anywhere is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<g style="stroke:#265c42"><path d="M0 0 L8 8" fill="none"/></g>`)},
			want:     `<path> stroke="#265c42" has no stroke-width on it or a parent`,
		},
		{
			name:     "a stroked drawing whose stroke-width is set on a parent group passes",
			manifest: spriteOnly,
			files: map[string]string{sprite: page(128, 128,
				`<g stroke-width="8"><circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#265c42"/></g>`)},
		},
		{
			name:     "a stroke-width under 4 px in a style declaration is rejected",
			manifest: spriteOnly,
			files:    map[string]string{sprite: page(128, 128, `<circle cx="64" cy="64" r="40" style="fill:#3e8948;stroke:#265c42;stroke-width:3.5px"/>`)},
			want:     `stroke-width="3.5px" is thinner than 4 px`,
		},
		{
			name:     "a prop with no outline at all is rejected",
			manifest: propOnly,
			files:    map[string]string{prop: page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948"/>`)},
			want:     "no shape has an 8 px outline",
		},
		{
			name:     "a prop whose only stroke is thinner than one world pixel is rejected",
			manifest: propOnly,
			files:    map[string]string{prop: page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#265c42" stroke-width="4" `+round+`/>`)},
			want:     "no shape has an 8 px outline",
		},
		{
			name:     "a prop outlined in ink is rejected",
			manifest: propOnly,
			files:    map[string]string{prop: page(128, 128, `<circle cx="64" cy="64" r="40" fill="#3e8948" stroke="#181425" stroke-width="8" `+round+`/>`)},
			want:     `stroke="#181425": props are never outlined in ink or black`,
		},
		{
			name:     "a prop outlined in black in a style declaration is rejected",
			manifest: propOnly,
			files: map[string]string{prop: page(128, 128,
				`<circle cx="64" cy="64" r="40" fill="#3e8948" style="stroke:#000000;stroke-width:8px;stroke-linejoin:round;stroke-linecap:round"/>`)},
			want: `stroke="#000000": props are never outlined in ink or black`,
		},
		{
			name:     "a prop stroke with mitred joins is rejected",
			manifest: propOnly,
			files: map[string]string{prop: page(128, 128,
				`<rect width="64" height="64" fill="#b86f50" stroke="#733e39" stroke-width="8" stroke-linejoin="miter" stroke-linecap="round"/>`)},
			want: `<rect> stroke="#733e39" has stroke-linejoin="miter" stroke-linecap="round", want round joins and caps`,
		},
		{
			name:     "a prop stroke with no caps set gets SVG's butt caps and is rejected",
			manifest: propOnly,
			files: map[string]string{prop: page(128, 128,
				`<path d="M8 8 L64 64" fill="none" stroke="#733e39" stroke-width="8" stroke-linejoin="round"/>`)},
			want: `<path> stroke="#733e39" has stroke-linejoin="round" stroke-linecap="butt", want round joins and caps`,
		},
		{
			name:     "a prop whose outline, joins and caps come from a parent's style passes",
			manifest: propOnly,
			files: map[string]string{prop: page(128, 128,
				`<g style="stroke:#265c42;stroke-width:8;stroke-linejoin:round;stroke-linecap:round"><circle cx="64" cy="64" r="40" fill="#3e8948"/></g>`)},
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
	prop := strings.HasPrefix(rel, propsFolder)
	outlined := false
	var problems []string
	dec := xml.NewDecoder(f)
	sawRoot := false
	// strokes holds the stroke inherited by each open element; the top is the
	// innermost. The root's parent draws no stroke and has no width set.
	strokes := []inheritedStroke{{stroke: "none", width: 1, linejoin: "miter", linecap: "butt"}}
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return append(problems, "not well-formed XML: "+err.Error())
		}
		if _, ok := tok.(xml.EndElement); ok {
			strokes = strokes[:len(strokes)-1]
			continue
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		props := properties(start)
		stroke := strokes[len(strokes)-1].with(props)
		strokes = append(strokes, stroke)
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
		for _, p := range props {
			problems = append(problems, checkProperty(elem, p.name, p.value, ground, prop, palette)...)
		}
		// A shape that strokes with no stroke-width on it or any parent gets
		// SVG's default 1 px hairline. Ground tiles already report the stroke.
		if !ground && shapeElements[elem] && stroke.stroke != "none" && !stroke.widthSet {
			problems = append(problems, fmt.Sprintf(
				"<%s> stroke=%q has no stroke-width on it or a parent, so it is SVG's default 1 px, thinner than %d px (half a world pixel)",
				elem, stroke.stroke, minStrokeWidth))
		}
		if prop && shapeElements[elem] && stroke.stroke != "none" {
			if stroke.width == outlineWidth {
				outlined = true
			}
			if stroke.linejoin != "round" || stroke.linecap != "round" {
				problems = append(problems, fmt.Sprintf(
					"<%s> stroke=%q has stroke-linejoin=%q stroke-linecap=%q, want round joins and caps",
					elem, stroke.stroke, stroke.linejoin, stroke.linecap))
			}
		}
	}
	if !sawRoot {
		return append(problems, "no <svg> root element")
	}
	if prop && !outlined {
		problems = append(problems, fmt.Sprintf(
			"no shape has an %d px outline (one world pixel); props need one in the darker palette colour paired with the fill",
			outlineWidth))
	}
	return problems
}

// shapeElements are the SVG elements that draw a stroke themselves.
var shapeElements = map[string]bool{
	"path":     true,
	"rect":     true,
	"circle":   true,
	"ellipse":  true,
	"line":     true,
	"polyline": true,
	"polygon":  true,
}

// inheritedStroke is the stroke, its width (and whether one is set), joins
// and caps, as an element inherits them from its parents. Unset, SVG draws a
// 1 px stroke with mitred joins and butt caps.
type inheritedStroke struct {
	stroke            string
	widthSet          bool
	width             float64 // -1 when the width is not a number of px
	linejoin, linecap string
}

// with returns the stroke an element with props ends up with, given that its
// parent's is s. The style attribute's declarations come after the
// presentation attributes in props, so they win, as in SVG.
func (s inheritedStroke) with(props []property) inheritedStroke {
	for _, p := range props {
		if p.value == "inherit" {
			continue
		}
		switch p.name {
		case "stroke":
			s.stroke = strings.ToLower(p.value)
		case "stroke-width":
			s.widthSet = true
			s.width = pagePx(p.value)
		case "stroke-linejoin":
			s.linejoin = strings.ToLower(p.value)
		case "stroke-linecap":
			s.linecap = strings.ToLower(p.value)
		}
	}
	return s
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

// properties returns an element's presentation attributes and then the
// declarations of its style attribute, with lower-case names. Style comes last
// because it overrides the attributes.
func properties(start xml.StartElement) []property {
	var props, styles []property
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
			styles = append(styles, property{strings.ToLower(strings.TrimSpace(prop)), strings.TrimSpace(val)})
		}
	}
	return append(props, styles...)
}

// checkProperty returns the problems with one presentation property of a
// drawing; ground and prop say whether it is a ground tileset or a prop.
func checkProperty(elem, name, value string, ground, prop bool, palette map[string]bool) []string {
	var problems []string
	at := fmt.Sprintf("<%s> %s=%q", elem, name, value)
	if prop && name == "stroke" && bannedOutlines[strings.ToLower(value)] {
		return []string{at + ": props are never outlined in ink or black; use the darker palette colour paired with the fill"}
	}
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

// tileSize is the size of one tileset tile in drawing px (16 world px).
const tileSize = 16 * drawingPxPerWorldPx

// terrainPairs are the transitions the outdoor ground terrain set must paint:
// each pair's 16 corner combinations need a tile.
var terrainPairs = [][2]string{{"grass", "path"}, {"grass", "water"}, {"sand", "water"}}

func TestTilesets_EveryTileSitsInItsDrawingsGridCell(t *testing.T) {
	sheets, err := filepath.Glob("tiles/*.svg")
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) == 0 {
		t.Fatal("no tileset drawings in tiles/")
	}
	for _, sheet := range sheets {
		name := strings.TrimSuffix(filepath.Base(sheet), ".svg")
		t.Run(name, func(t *testing.T) {
			ts := loadTSX(t, filepath.Join("tiles", name+".tsx"))
			for _, p := range checkTileset(name, loadSheet(t, sheet), ts) {
				t.Error(p)
			}
		})
	}
}

func TestOutdoorTileset_GroundTerrainCoversEveryCornerCombination(t *testing.T) {
	ts := loadTSX(t, "tiles/tileset_outdoor.tsx")
	for _, p := range checkTerrain(ts, "ground", terrainPairs) {
		t.Error(p)
	}
}

func TestOutdoorTileset_WaterAnimatesThroughItsThreeFrames(t *testing.T) {
	ts := loadTSX(t, "tiles/tileset_outdoor.tsx")
	types := make(map[int]string)
	var water *tsxTile
	for i, tile := range ts.Tiles {
		types[tile.ID] = tile.Type
		if tile.Type == "water_1" {
			water = &ts.Tiles[i]
		}
	}
	if water == nil {
		t.Fatal("no water_1 tile")
	}
	var got []string
	for _, f := range water.Frames {
		got = append(got, fmt.Sprintf("%s %dms", types[f.TileID], f.Duration))
	}
	want := []string{"water_1 300ms", "water_2 300ms", "water_3 300ms"}
	if !slices.Equal(got, want) {
		t.Errorf("water_1 animation = %q, want %q", got, want)
	}
}

func TestTilesetCheck_RejectsTilesThatDisagreeWithTheDrawing(t *testing.T) {
	sheet := loadSheet(t, "tiles/tileset_outdoor.svg")
	tests := []struct {
		name   string
		change func(ts *tsxTileset)
		want   string // a substring of a problem expected; "" for none
	}{
		{
			name:   "the outdoor tileset as committed passes",
			change: func(*tsxTileset) {},
		},
		{
			name:   "a tile whose id is not its drawing's grid cell is rejected",
			change: func(ts *tsxTileset) { tileOfType(t, ts, "cliff_face").ID = 63 },
			want:   `tile 63 type "cliff_face": drawn in cell 59 (row 7, column 3), want id 59`,
		},
		{
			name: "a drawn tile with no tileset entry is rejected",
			change: func(ts *tsxTileset) {
				ts.Tiles = slices.DeleteFunc(ts.Tiles, func(x tsxTile) bool { return x.Type == "sand" })
			},
			want: `"sand" is drawn in cell 5 but has no tile in the tileset`,
		},
		{
			name:   "a type on two tiles is rejected",
			change: func(ts *tsxTileset) { tileOfType(t, ts, "grass_2").Type = "grass" },
			want:   `tile 1 type "grass": type repeats tile 0`,
		},
		{
			name:   "a tile with no type is rejected",
			change: func(ts *tsxTileset) { tileOfType(t, ts, "path").Type = "" },
			want:   `tile 4 has no type`,
		},
		{
			name:   "a type with no drawing is rejected",
			change: func(ts *tsxTileset) { tileOfType(t, ts, "cliff_face").Type = "cliff_cave" },
			want:   `tile 59 type "cliff_cave": no <g id="cliff_cave"> tile in the drawing`,
		},
		{
			name:   "an image that is not the drawing's exported PNG is rejected",
			change: func(ts *tsxTileset) { ts.Image.Source = "../../assets/tiles/tileset_placeholder.png" },
			want:   `image source "../../assets/tiles/tileset_placeholder.png", want "../../assets/tiles/tileset_outdoor.png"`,
		},
		{
			name:   "an image size other than the drawing's page is rejected",
			change: func(ts *tsxTileset) { ts.Image.Height = 1152 },
			want:   `image is 1024 by 1152 px, want the drawing's page 1024 by 1024`,
		},
		{
			name:   "a column count other than the drawing's is rejected",
			change: func(ts *tsxTileset) { ts.Columns = 4 },
			want:   `columns="4", want 8`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := loadTSX(t, "tiles/tileset_outdoor.tsx")
			tc.change(&ts)

			problems := checkTileset("tileset_outdoor", sheet, ts)

			checkProblems(t, problems, tc.want)
		})
	}
}

func TestTerrainCheck_RejectsIncompleteOrMisnamedTransitions(t *testing.T) {
	tests := []struct {
		name   string
		change func(ts *tsxTileset)
		want   string // a substring of a problem expected; "" for none
	}{
		{
			name:   "the outdoor ground terrain as committed passes",
			change: func(*tsxTileset) {},
		},
		{
			name:   "a missing mixed corner combination is rejected",
			change: func(ts *tsxTileset) { dropWangTile(t, ts, "grass_water_0110") },
			want:   `grass/water: no tile has corners 0110 (nw ne sw se, 0 grass, 1 water)`,
		},
		{
			name:   "a missing plain corner combination is rejected",
			change: func(ts *tsxTileset) { dropWangTile(t, ts, "sand") },
			want:   `sand/water: no tile has corners 0000 (nw ne sw se, 0 sand, 1 water)`,
		},
		{
			name: "a mixed corner combination on two tiles is rejected",
			change: func(ts *tsxTileset) {
				w := tileOfType(t, ts, "cliff_face").ID
				ts.WangSets[0].Tiles = append(ts.WangSets[0].Tiles, tsxWangTile{TileID: w, WangID: "0,1,0,2,0,1,0,1"})
			},
			want: `grass/path: corners 0001 are on 2 tiles`,
		},
		{
			name: "a transition tile whose corners differ from its name is rejected",
			change: func(ts *tsxTileset) {
				setWangID(t, ts, "sand_water_0001", "0,4,0,4,0,3,0,4")
			},
			want: `tile 37 "sand_water_0001": corners nw sand, ne sand, sw water, se sand, want nw sand, ne sand, sw sand, se water`,
		},
		{
			name:   "a transition tile left out of the terrain set is rejected",
			change: func(ts *tsxTileset) { dropWangTile(t, ts, "grass_path_1110") },
			want:   `tile 22 "grass_path_1110": not in wangset "ground"`,
		},
		{
			name:   "a plain tile painted as another terrain is rejected",
			change: func(ts *tsxTileset) { setWangID(t, ts, "grass_3", "0,2,0,2,0,2,0,2") },
			want:   `tile 2 "grass_3": corners are all path, want a path tile`,
		},
		{
			name:   "an edge-type terrain set is rejected",
			change: func(ts *tsxTileset) { ts.WangSets[0].Type = "edge" },
			want:   `wangset "ground" is type "edge", want "corner"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := loadTSX(t, "tiles/tileset_outdoor.tsx")
			tc.change(&ts)

			problems := checkTerrain(ts, "ground", terrainPairs)

			checkProblems(t, problems, tc.want)
		})
	}
}

// imageCollections are the image-collection tilesets, by slash path relative to
// assets-src. Each lists every drawing in its own folder, one tile per drawing.
var imageCollections = []string{"sprites/decor/decor.tsx"}

func TestImageCollections_EveryTileMatchesADrawingAndNoDrawingIsLeftOut(t *testing.T) {
	for _, path := range imageCollections {
		t.Run(path, func(t *testing.T) {
			dir := filepath.ToSlash(filepath.Dir(path))
			for _, p := range checkImageCollection(dir, loadPages(t, dir), loadTSX(t, path)) {
				t.Error(p)
			}
		})
	}
}

func TestImageCollectionCheck_RejectsTilesThatDisagreeWithTheDrawings(t *testing.T) {
	pages := loadPages(t, "sprites/decor")
	tests := []struct {
		name   string
		change func(ts *tsxTileset)
		want   string // a substring of a problem expected; "" for none
	}{
		{
			name:   "the decor tileset as committed passes",
			change: func(*tsxTileset) {},
		},
		{
			name: "a tile naming a drawing that does not exist is rejected",
			change: func(ts *tsxTileset) {
				tile := tileOfType(t, ts, "twig")
				tile.Type = "stick"
				tile.Image.Source = "../../../assets/sprites/decor/stick.png"
			},
			want: `tile 14 type "stick": no drawing sprites/decor/stick.svg`,
		},
		{
			name: "a tile whose image is not its own drawing's exported PNG is rejected",
			change: func(ts *tsxTileset) {
				tileOfType(t, ts, "clover").Image.Source = "../../../assets/sprites/decor/twig.png"
			},
			want: `tile 13 type "clover": image source "../../../assets/sprites/decor/twig.png", want "../../../assets/sprites/decor/clover.png"`,
		},
		{
			name:   "a tile whose image size is not its drawing's page is rejected",
			change: func(ts *tsxTileset) { tileOfType(t, ts, "pebble_1").Image.Width = 256 },
			want:   `tile 7 type "pebble_1": image is 256 by 128 px, want the drawing's page 128 by 128`,
		},
		{
			name: "a drawing in the folder with no tile is rejected",
			change: func(ts *tsxTileset) {
				ts.Tiles = slices.DeleteFunc(ts.Tiles, func(x tsxTile) bool { return x.Type == "leaf_2" })
			},
			want: `sprites/decor/leaf_2.svg has no tile in the tileset`,
		},
		{
			name:   "a drawing on two tiles is rejected",
			change: func(ts *tsxTileset) { tileOfType(t, ts, "tuft_2").Type = "tuft_1" },
			want:   `tile 1 type "tuft_1": type repeats tile 0`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := loadTSX(t, "sprites/decor/decor.tsx")
			tc.change(&ts)

			problems := checkImageCollection("sprites/decor", pages, ts)

			checkProblems(t, problems, tc.want)
		})
	}
}

// pxSize is a drawing's page size in drawing px.
type pxSize struct{ w, h int }

// loadPages returns the page size of each drawing in dir, keyed by its file
// name without .svg.
func loadPages(t *testing.T, dir string) map[string]pxSize {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(filepath.FromSlash(dir), "*.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no drawings in %s", dir)
	}
	pages := make(map[string]pxSize)
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		var root struct {
			Width  string `xml:"width,attr"`
			Height string `xml:"height,attr"`
		}
		err = xml.NewDecoder(f).Decode(&root)
		f.Close()
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		pages[strings.TrimSuffix(filepath.Base(p), ".svg")] = pxSize{int(pagePx(root.Width)), int(pagePx(root.Height))}
	}
	return pages
}

// checkImageCollection returns the ways the image-collection tileset ts
// disagrees with the drawings in dir, whose page sizes are pages: each tile's
// type names a drawing, its image is that drawing's exported PNG at the
// drawing's page size, and every drawing has exactly one tile.
func checkImageCollection(dir string, pages map[string]pxSize, ts tsxTileset) []string {
	var problems []string
	// The .tsx sits in assets-src/<dir>; the PNG at assets/<dir>/<name>.png.
	toRoot := strings.Repeat("../", strings.Count(dir, "/")+2)
	seen := make(map[string]int)
	for _, tile := range ts.Tiles {
		at := fmt.Sprintf("tile %d type %q", tile.ID, tile.Type)
		if first, dup := seen[tile.Type]; dup {
			problems = append(problems, fmt.Sprintf("%s: type repeats tile %d", at, first))
			continue
		}
		seen[tile.Type] = tile.ID
		page, drawn := pages[tile.Type]
		if !drawn {
			problems = append(problems, fmt.Sprintf("%s: no drawing %s/%s.svg", at, dir, tile.Type))
			continue
		}
		if want := toRoot + "assets/" + dir + "/" + tile.Type + ".png"; tile.Image.Source != want {
			problems = append(problems, fmt.Sprintf("%s: image source %q, want %q", at, tile.Image.Source, want))
		}
		if tile.Image.Width != page.w || tile.Image.Height != page.h {
			problems = append(problems, fmt.Sprintf("%s: image is %d by %d px, want the drawing's page %d by %d",
				at, tile.Image.Width, tile.Image.Height, page.w, page.h))
		}
	}
	for name := range pages {
		if _, ok := seen[name]; !ok {
			problems = append(problems, fmt.Sprintf("%s/%s.svg has no tile in the tileset", dir, name))
		}
	}
	slices.Sort(problems)
	return problems
}

// checkProblems fails unless problems is empty when want is "", or holds a
// problem containing want.
func checkProblems(t *testing.T, problems []string, want string) {
	t.Helper()
	if want == "" {
		if len(problems) != 0 {
			t.Fatalf("problems = %q, want none", problems)
		}
		return
	}
	for _, p := range problems {
		if strings.Contains(p, want) {
			return
		}
	}
	t.Fatalf("problems = %q, want one containing %q", problems, want)
}

func tileOfType(t *testing.T, ts *tsxTileset, typ string) *tsxTile {
	t.Helper()
	for i := range ts.Tiles {
		if ts.Tiles[i].Type == typ {
			return &ts.Tiles[i]
		}
	}
	t.Fatalf("no tile of type %q", typ)
	return nil
}

func dropWangTile(t *testing.T, ts *tsxTileset, typ string) {
	t.Helper()
	id := tileOfType(t, ts, typ).ID
	ws := &ts.WangSets[0]
	n := len(ws.Tiles)
	ws.Tiles = slices.DeleteFunc(ws.Tiles, func(w tsxWangTile) bool { return w.TileID == id })
	if len(ws.Tiles) == n {
		t.Fatalf("%q has no wangtile to drop", typ)
	}
}

func setWangID(t *testing.T, ts *tsxTileset, typ, wangID string) {
	t.Helper()
	id := tileOfType(t, ts, typ).ID
	for i := range ts.WangSets[0].Tiles {
		if ts.WangSets[0].Tiles[i].TileID == id {
			ts.WangSets[0].Tiles[i].WangID = wangID
			return
		}
	}
	t.Fatalf("%q has no wangtile", typ)
}

// tsxTileset is the part of a Tiled .tsx tileset the tests read.
type tsxTileset struct {
	TileWidth  int          `xml:"tilewidth,attr"`
	TileHeight int          `xml:"tileheight,attr"`
	Columns    int          `xml:"columns,attr"`
	Image      tsxImage     `xml:"image"`
	Tiles      []tsxTile    `xml:"tile"`
	WangSets   []tsxWangSet `xml:"wangsets>wangset"`
}

type tsxImage struct {
	Source string `xml:"source,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

type tsxTile struct {
	ID     int        `xml:"id,attr"`
	Type   string     `xml:"type,attr"`
	Frames []tsxFrame `xml:"animation>frame"`
	// Image is the tile's own image in an image-collection tileset.
	Image tsxImage `xml:"image"`
}

type tsxFrame struct {
	TileID   int `xml:"tileid,attr"`
	Duration int `xml:"duration,attr"`
}

type tsxWangSet struct {
	Name   string `xml:"name,attr"`
	Type   string `xml:"type,attr"`
	Colors []struct {
		Name string `xml:"name,attr"`
	} `xml:"wangcolor"`
	Tiles []tsxWangTile `xml:"wangtile"`
}

type tsxWangTile struct {
	TileID int    `xml:"tileid,attr"`
	WangID string `xml:"wangid,attr"`
}

func loadTSX(t *testing.T, path string) tsxTileset {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tileset: %v", err)
	}
	var ts tsxTileset
	if err := xml.Unmarshal(data, &ts); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return ts
}

// sheet is a tileset drawing: its page size in px and the grid cell
// (row*columns + column) of each tile, keyed by the tile group's id.
type sheet struct {
	width, height int
	cells         map[string]int
}

// loadSheet reads a tileset drawing. Its tiles are the <g id="..."> children
// of the root <svg>, each moved to its cell by transform="translate(x,y)".
func loadSheet(t *testing.T, path string) sheet {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open sheet: %v", err)
	}
	defer f.Close()

	s := sheet{cells: make(map[string]int)}
	dec := xml.NewDecoder(f)
	depth := 0
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		switch tok := tok.(type) {
		case xml.EndElement:
			depth--
		case xml.StartElement:
			depth++
			attrs := make(map[string]string)
			for _, a := range tok.Attr {
				if a.Name.Space == "" {
					attrs[a.Name.Local] = a.Value
				}
			}
			if depth == 1 {
				s.width, s.height = int(pagePx(attrs["width"])), int(pagePx(attrs["height"]))
				if s.width <= 0 || s.width%tileSize != 0 || s.height <= 0 || s.height%tileSize != 0 {
					t.Fatalf("%s: page %q by %q is not a whole number of %d px tiles", path, attrs["width"], attrs["height"], tileSize)
				}
			}
			if depth != 2 || tok.Name.Local != "g" || attrs["id"] == "" {
				continue
			}
			id := attrs["id"]
			var xy []float64
			if inner, ok := strings.CutPrefix(attrs["transform"], "translate("); ok {
				xy = numbers(strings.TrimSuffix(inner, ")"))
			}
			if len(xy) != 2 || int(xy[0])%tileSize != 0 || int(xy[1])%tileSize != 0 ||
				xy[0] < 0 || xy[1] < 0 || int(xy[0]) >= s.width || int(xy[1]) >= s.height {
				t.Fatalf("%s: tile %q has transform %q, want translate(column*%d,row*%d) inside the page", path, id, attrs["transform"], tileSize, tileSize)
			}
			if _, dup := s.cells[id]; dup {
				t.Fatalf("%s: tile %q is drawn twice", path, id)
			}
			s.cells[id] = int(xy[1])/tileSize*(s.width/tileSize) + int(xy[0])/tileSize
		}
	}
	return s
}

// checkTileset returns the ways the Tiled tileset ts disagrees with its
// drawing s, the tiles/<name>.svg it is exported from.
func checkTileset(name string, s sheet, ts tsxTileset) []string {
	var problems []string
	columns := s.width / tileSize
	if ts.TileWidth != tileSize || ts.TileHeight != tileSize {
		problems = append(problems, fmt.Sprintf("tiles are %d by %d px, want %d by %d", ts.TileWidth, ts.TileHeight, tileSize, tileSize))
	}
	if ts.Columns != columns {
		problems = append(problems, fmt.Sprintf("columns=\"%d\", want %d", ts.Columns, columns))
	}
	if want := "../../assets/tiles/" + name + ".png"; ts.Image.Source != want {
		problems = append(problems, fmt.Sprintf("image source %q, want %q", ts.Image.Source, want))
	}
	if ts.Image.Width != s.width || ts.Image.Height != s.height {
		problems = append(problems, fmt.Sprintf("image is %d by %d px, want the drawing's page %d by %d",
			ts.Image.Width, ts.Image.Height, s.width, s.height))
	}
	seen := make(map[string]int)
	for _, tile := range ts.Tiles {
		if tile.Type == "" {
			problems = append(problems, fmt.Sprintf("tile %d has no type", tile.ID))
			continue
		}
		at := fmt.Sprintf("tile %d type %q", tile.ID, tile.Type)
		if first, dup := seen[tile.Type]; dup {
			problems = append(problems, fmt.Sprintf("%s: type repeats tile %d", at, first))
			continue
		}
		seen[tile.Type] = tile.ID
		cell, drawn := s.cells[tile.Type]
		switch {
		case !drawn:
			problems = append(problems, fmt.Sprintf("%s: no <g id=%q> tile in the drawing", at, tile.Type))
		case cell != tile.ID:
			problems = append(problems, fmt.Sprintf("%s: drawn in cell %d (row %d, column %d), want id %d",
				at, cell, cell/columns, cell%columns, cell))
		}
	}
	for typ, cell := range s.cells {
		if _, ok := seen[typ]; !ok {
			problems = append(problems, fmt.Sprintf("%q is drawn in cell %d but has no tile in the tileset", typ, cell))
		}
	}
	slices.Sort(problems)
	return problems
}

// transitionName matches a transition tile type, <a>_<b>_<nw><ne><sw><se>,
// with 0 for terrain a and 1 for terrain b at each corner.
var transitionName = regexp.MustCompile(`^([a-z]+)_([a-z]+)_([01]{4})$`)

// plainName matches a single-terrain tile type and its variants: grass, grass_2.
var plainName = regexp.MustCompile(`^([a-z]+)(?:_[0-9]+)?$`)

// checkTerrain returns the problems with ts's corner terrain set setName: each
// pair needs all 16 corner combinations, its mixed ones on exactly one tile,
// and every tile's corners must be the ones its name states.
func checkTerrain(ts tsxTileset, setName string, pairs [][2]string) []string {
	var ws *tsxWangSet
	for i := range ts.WangSets {
		if ts.WangSets[i].Name == setName {
			ws = &ts.WangSets[i]
		}
	}
	if ws == nil {
		return []string{fmt.Sprintf("no wangset %q", setName)}
	}
	var problems []string
	if ws.Type != "corner" {
		problems = append(problems, fmt.Sprintf("wangset %q is type %q, want \"corner\"", setName, ws.Type))
	}
	types := make(map[int]string)
	for _, tile := range ts.Tiles {
		types[tile.ID] = tile.Type
	}
	colour := func(i int) string {
		if i < 1 || i > len(ws.Colors) {
			return fmt.Sprintf("colour %d", i)
		}
		return ws.Colors[i-1].Name
	}

	// corners maps each tile in the set to its terrain at nw, ne, sw, se.
	corners := make(map[int][4]string)
	for _, wt := range ws.Tiles {
		at := fmt.Sprintf("tile %d %q", wt.TileID, types[wt.TileID])
		ids := numbers(wt.WangID)
		if len(ids) != 8 {
			problems = append(problems, fmt.Sprintf("%s: wangid %q, want 8 numbers", at, wt.WangID))
			continue
		}
		if ids[0] != 0 || ids[2] != 0 || ids[4] != 0 || ids[6] != 0 {
			problems = append(problems, fmt.Sprintf("%s: wangid %q sets an edge in a corner set", at, wt.WangID))
		}
		if _, dup := corners[wt.TileID]; dup {
			problems = append(problems, at+": in the terrain set twice")
		}
		// Tiled lists a wangid clockwise from the top edge: top, ne, right,
		// se, bottom, sw, left, nw.
		c := [4]string{colour(int(ids[7])), colour(int(ids[1])), colour(int(ids[5])), colour(int(ids[3]))}
		corners[wt.TileID] = c

		typ := types[wt.TileID]
		if m := transitionName.FindStringSubmatch(typ); m != nil {
			var want [4]string
			for i, d := range m[3] {
				want[i] = m[1+int(d-'0')]
			}
			if c != want {
				problems = append(problems, fmt.Sprintf("%s: corners %s, want %s", at, describeCorners(c), describeCorners(want)))
			}
			continue
		}
		if c[0] != c[1] || c[1] != c[2] || c[2] != c[3] {
			problems = append(problems, fmt.Sprintf("%s: corners %s, but the name states no corners", at, describeCorners(c)))
			continue
		}
		if m := plainName.FindStringSubmatch(typ); m == nil || m[1] != c[0] {
			problems = append(problems, fmt.Sprintf("%s: corners are all %s, want a %s tile", at, c[0], c[0]))
		}
	}

	for _, tile := range ts.Tiles {
		if m := transitionName.FindStringSubmatch(tile.Type); m != nil {
			if _, ok := corners[tile.ID]; !ok {
				problems = append(problems, fmt.Sprintf("tile %d %q: not in wangset %q", tile.ID, tile.Type, setName))
			}
		}
	}

	for _, pair := range pairs {
		count := make(map[string]int)
		for _, c := range corners {
			code := ""
			for _, terrain := range c {
				switch terrain {
				case pair[0]:
					code += "0"
				case pair[1]:
					code += "1"
				}
			}
			if len(code) == 4 {
				count[code]++
			}
		}
		for n := range 16 {
			code := fmt.Sprintf("%04b", n)
			switch {
			case count[code] == 0:
				problems = append(problems, fmt.Sprintf("%s/%s: no tile has corners %s (nw ne sw se, 0 %s, 1 %s)",
					pair[0], pair[1], code, pair[0], pair[1]))
			case count[code] > 1 && n != 0 && n != 15:
				problems = append(problems, fmt.Sprintf("%s/%s: corners %s are on %d tiles, want one",
					pair[0], pair[1], code, count[code]))
			}
		}
	}
	slices.Sort(problems)
	return problems
}

func describeCorners(c [4]string) string {
	return fmt.Sprintf("nw %s, ne %s, sw %s, se %s", c[0], c[1], c[2], c[3])
}
