package assets_test

import (
	"io/fs"
	"testing"

	"golang.org/x/image/font/sfnt"

	embedded "github.com/Elpulgo/critterkingdom/assets"
)

// A rune the font has no glyph for maps to glyph index 0, which draws as a
// blank box, so a non-zero index means the font can draw the letter.
func TestFont_DrawsTheSwedishLettersInBothCases(t *testing.T) {
	b, err := fs.ReadFile(embedded.FS, "fonts/Fredoka.ttf")
	if err != nil {
		t.Fatalf("read font: %v", err)
	}
	f, err := sfnt.Parse(b)
	if err != nil {
		t.Fatalf("parse font: %v", err)
	}
	var buf sfnt.Buffer
	glyph := func(r rune) sfnt.GlyphIndex {
		t.Helper()
		g, err := f.GlyphIndex(&buf, r)
		if err != nil {
			t.Fatalf("GlyphIndex(%q): %v", r, err)
		}
		return g
	}

	for _, r := range "åäöÅÄÖ" {
		if g := glyph(r); g == 0 {
			t.Errorf("font has no glyph for %q (U+%04X)", r, r)
		}
	}

	// The check can fail: the font has no CJK letters, so U+4E2D 中 is missing.
	if g := glyph('中'); g != 0 {
		t.Errorf("glyph index for U+4E2D = %d; want 0, as Fredoka has no CJK letters", g)
	}
}
