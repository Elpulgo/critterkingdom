// Package artlint checks hand-drawn SVG sprites against the machine-checkable
// rules of docs/ART-BIBLE.md: an exact integer page size, Endesga 32 colours
// only, no gradients, patterns, filters, images or text, and no stroke thinner
// than 4 px.
package artlint

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// MinStrokeWidth is the thinnest stroke the art bible allows in a drawing, in px.
const MinStrokeWidth = 4

const svgNamespace = "http://www.w3.org/2000/svg"

var hexColour = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// colourProperties are the paint properties whose value must be "none" or a
// palette colour, whether set as an attribute or inside a style attribute.
var colourProperties = map[string]bool{
	"fill":        true,
	"stroke":      true,
	"stop-color":  true,
	"flood-color": true,
	"color":       true,
}

// forbiddenElements are the SVG elements the art bible bans in sprites.
var forbiddenElements = map[string]bool{
	"linearGradient": true,
	"radialGradient": true,
	"pattern":        true,
	"filter":         true,
	"image":          true,
	"text":           true,
}

// Report is the result of linting one SVG.
type Report struct {
	// Width and Height are the root page size in px, or 0 when the root
	// attribute is not an integer.
	Width, Height int
	// Problems lists every rule the drawing breaks, one entry per finding,
	// in document order. An empty list means the drawing is clean.
	Problems []string
}

// LoadPalette reads a GIMP palette (.gpl) file and returns its colours as a
// set of lower-case "#rrggbb" keys. The first line must be "GIMP Palette";
// comment lines, blank lines and the Name: and Columns: header lines are
// skipped, and every other line must start with three 0-255 channel values.
func LoadPalette(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open palette: %w", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return nil, fmt.Errorf("read palette %s: %w", path, err)
		}
		return nil, fmt.Errorf("palette %s is empty, want a \"GIMP Palette\" header", path)
	}
	if first := strings.TrimSpace(sc.Text()); first != "GIMP Palette" {
		return nil, fmt.Errorf("palette %s first line = %q, want \"GIMP Palette\"", path, first)
	}

	colours := make(map[string]bool)
	lineNo := 1
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "Name:") || strings.HasPrefix(line, "Columns:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("palette %s line %d %q: want \"R G B [name]\"", path, lineNo, line)
		}
		var rgb [3]uint64
		for i := range rgb {
			v, err := strconv.ParseUint(fields[i], 10, 8)
			if err != nil {
				return nil, fmt.Errorf("palette %s line %d %q: channel %q is not 0-255", path, lineNo, line, fields[i])
			}
			rgb[i] = v
		}
		colours[fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])] = true
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read palette %s: %w", path, err)
	}
	if len(colours) == 0 {
		return nil, fmt.Errorf("palette %s has no colours", path)
	}
	return colours, nil
}

// Lint reads one SVG document and reports every art-bible rule it breaks.
// palette is the set of allowed lower-case "#rrggbb" colours, as returned by
// LoadPalette. The error is non-nil only when the input is not well-formed
// XML or has no root element; rule violations go into Report.Problems.
func Lint(r io.Reader, palette map[string]bool) (Report, error) {
	var rep Report
	dec := xml.NewDecoder(r)
	root := true
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return rep, fmt.Errorf("decode svg: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		line, _ := dec.InputPos()
		at := fmt.Sprintf("line %d: <%s>", line, start.Name.Local)

		if root {
			root = false
			rep.checkPage(at, start)
		}
		if isSVGElement(start.Name) && forbiddenElements[start.Name.Local] {
			rep.addf("%s is not allowed in sprites (no gradients, patterns, filters, images or text)", at)
		}
		for _, attr := range start.Attr {
			if attr.Name.Space != "" {
				continue // inkscape:, sodipodi:, xlink: and other namespaced attributes
			}
			name := strings.ToLower(attr.Name.Local)
			if name == "style" {
				for _, decl := range strings.Split(attr.Value, ";") {
					prop, val, ok := strings.Cut(decl, ":")
					if !ok {
						continue
					}
					rep.checkProperty(palette, at, "style "+strings.TrimSpace(prop),
						strings.ToLower(strings.TrimSpace(prop)), strings.TrimSpace(val))
				}
				continue
			}
			rep.checkProperty(palette, at, name, name, strings.TrimSpace(attr.Value))
		}
	}
	if root {
		return rep, errors.New("decode svg: no root element")
	}
	return rep, nil
}

// checkPage checks that the root element is an <svg> whose integer width and
// height match a viewBox of "0 0 width height".
func (rep *Report) checkPage(at string, start xml.StartElement) {
	if start.Name.Local != "svg" {
		rep.addf("%s: root element must be <svg>", at)
		return
	}
	var width, height, viewBox string
	var hasViewBox bool
	for _, attr := range start.Attr {
		if attr.Name.Space != "" {
			continue
		}
		switch attr.Name.Local {
		case "width":
			width = attr.Value
		case "height":
			height = attr.Value
		case "viewBox":
			viewBox, hasViewBox = attr.Value, true
		}
	}

	w, wErr := strconv.Atoi(strings.TrimSpace(width))
	if wErr != nil || w <= 0 {
		rep.addf("%s width=%q: want a positive whole number of px", at, width)
	} else {
		rep.Width = w
	}
	h, hErr := strconv.Atoi(strings.TrimSpace(height))
	if hErr != nil || h <= 0 {
		rep.addf("%s height=%q: want a positive whole number of px", at, height)
	} else {
		rep.Height = h
	}
	if rep.Width == 0 || rep.Height == 0 {
		return // the viewBox cannot be compared against a broken page size
	}

	want := fmt.Sprintf("0 0 %d %d", rep.Width, rep.Height)
	if !hasViewBox || !viewBoxEquals(viewBox, rep.Width, rep.Height) {
		rep.addf("%s viewBox=%q: want %q so one drawing px is one PNG px", at, viewBox, want)
	}
}

// viewBoxEquals reports whether vb is "0 0 w h", allowing comma or extra
// whitespace separators as SVG does.
func viewBoxEquals(vb string, w, h int) bool {
	fields := strings.FieldsFunc(vb, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	if len(fields) != 4 {
		return false
	}
	want := [4]int{0, 0, w, h}
	for i, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil || v != float64(want[i]) {
			return false
		}
	}
	return true
}

// checkProperty checks one presentation property, named prop (lower case),
// from an attribute or a style declaration labelled label.
func (rep *Report) checkProperty(palette map[string]bool, at, label, prop, value string) {
	switch {
	case colourProperties[prop]:
		if value == "none" {
			return
		}
		if !hexColour.MatchString(value) {
			rep.addf("%s %s=%q: want \"none\" or a #rrggbb Endesga 32 colour", at, label, value)
			return
		}
		if !palette[strings.ToLower(value)] {
			rep.addf("%s %s=%q is not in the Endesga 32 palette", at, label, value)
		}
	case prop == "stroke-width":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil || v < MinStrokeWidth {
			rep.addf("%s %s=%q: want a number of at least %d px", at, label, value, MinStrokeWidth)
		}
	}
}

func (rep *Report) addf(format string, args ...any) {
	rep.Problems = append(rep.Problems, fmt.Sprintf(format, args...))
}

func isSVGElement(name xml.Name) bool {
	return name.Space == "" || name.Space == svgNamespace
}
