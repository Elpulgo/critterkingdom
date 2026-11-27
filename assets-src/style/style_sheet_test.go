package style_test

import (
	"bufio"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

var hexColour = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

var colourProperties = map[string]bool{
	"fill":        true,
	"stroke":      true,
	"stop-color":  true,
	"flood-color": true,
	"color":       true,
}

func TestStyleSheet_UsesOnlyPaletteColours(t *testing.T) {
	palette := loadPalette(t, "../palettes/endesga_32.gpl")

	f, err := os.Open("style_sheet.svg")
	if err != nil {
		t.Fatalf("open style_sheet.svg: %v", err)
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	checked := 0
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("decode style_sheet.svg: %v", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		for _, attr := range start.Attr {
			name := strings.ToLower(attr.Name.Local)
			if colourProperties[name] {
				checked++
				checkColour(t, palette, start.Name.Local, name, attr.Value)
			}
			if name == "style" {
				for _, decl := range strings.Split(attr.Value, ";") {
					prop, val, ok := strings.Cut(decl, ":")
					if !ok {
						continue
					}
					prop = strings.ToLower(strings.TrimSpace(prop))
					if colourProperties[prop] {
						checked++
						checkColour(t, palette, start.Name.Local, "style."+prop, strings.TrimSpace(val))
					}
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("found no colour attributes in style_sheet.svg")
	}
}

func TestStyleSheet_PageIsTheMockScreenAtEightPxPerWorldPixel(t *testing.T) {
	f, err := os.Open("style_sheet.svg")
	if err != nil {
		t.Fatalf("open style_sheet.svg: %v", err)
	}
	defer f.Close()

	var root struct {
		XMLName xml.Name
		Width   string `xml:"width,attr"`
		Height  string `xml:"height,attr"`
		ViewBox string `xml:"viewBox,attr"`
	}
	if err := xml.NewDecoder(f).Decode(&root); err != nil {
		t.Fatalf("decode style_sheet.svg root: %v", err)
	}
	if root.XMLName.Local != "svg" {
		t.Fatalf("root element = <%s>, want <svg>", root.XMLName.Local)
	}
	if root.Width != "2560" || root.Height != "1440" || root.ViewBox != "0 0 2560 1440" {
		t.Errorf("root size = width=%q height=%q viewBox=%q, want width=\"2560\" height=\"1440\" viewBox=\"0 0 2560 1440\"",
			root.Width, root.Height, root.ViewBox)
	}
}

func loadPalette(t *testing.T, path string) map[string]bool {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open palette %s: %v", path, err)
	}
	defer f.Close()

	colours := make(map[string]bool)
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if first {
			first = false
			if line != "GIMP Palette" {
				t.Fatalf("palette %s first line = %q, want \"GIMP Palette\"", path, line)
			}
			continue
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			t.Fatalf("palette line %q has fewer than 4 tab-separated fields", line)
		}
		hex := "#" + strings.ToLower(strings.TrimSpace(fields[3]))
		colours[hex] = true
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read palette %s: %v", path, err)
	}
	if len(colours) != 32 {
		t.Fatalf("palette %s has %d colours, want 32", path, len(colours))
	}
	return colours
}

func checkColour(t *testing.T, palette map[string]bool, elem, prop, value string) {
	t.Helper()
	v := strings.TrimSpace(value)
	if v == "none" || (strings.HasPrefix(v, "url(#") && strings.HasSuffix(v, ")")) {
		return
	}
	if !hexColour.MatchString(v) {
		t.Errorf("<%s> %s=%q: want a #rrggbb palette colour, \"none\" or \"url(#...)\"", elem, prop, value)
		return
	}
	if !palette[strings.ToLower(v)] {
		t.Errorf("<%s> %s=%q is not in the Endesga 32 palette", elem, prop, value)
	}
}
