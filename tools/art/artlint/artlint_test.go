package artlint_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/tools/art/artlint"
)

const palettePath = "../../../assets-src/palettes/endesga_32.gpl"

// page is the root of a clean 64 by 64 drawing.
const page = `width="64" height="64" viewBox="0 0 64 64"`

// cleanBody follows the art bible: palette fills, a darker-fill outline with
// round joins, a 4 px stroke, and a #181425 shadow with opacity.
const cleanBody = `
  <ellipse cx="32" cy="56" rx="20" ry="4" fill="#181425" opacity="0.3"/>
  <circle cx="32" cy="30" r="22" fill="#3e8948" stroke="#265c42" stroke-width="4"
          stroke-linejoin="round" stroke-linecap="round"/>
  <circle cx="24" cy="22" r="6" style="fill:#63c74d;stroke:none"/>`

func svg(root, body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" ` + root + `>` + body + `
</svg>`
}

func loadPalette(t *testing.T) map[string]bool {
	t.Helper()
	palette, err := artlint.LoadPalette(palettePath)
	if err != nil {
		t.Fatalf("LoadPalette(%s): %v", palettePath, err)
	}
	return palette
}

func TestLint_DrawingThatFollowsTheArtBible_PassesWithItsPageSize(t *testing.T) {
	rep, err := artlint.Lint(strings.NewReader(svg(page, cleanBody)), loadPalette(t))
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if rep.Width != 64 || rep.Height != 64 {
		t.Errorf("page size = %d by %d, want 64 by 64", rep.Width, rep.Height)
	}
	if len(rep.Problems) != 0 {
		t.Errorf("problems = %q, want none", rep.Problems)
	}
}

func TestLint_EachArtBibleViolation_IsReportedOnce(t *testing.T) {
	palette := loadPalette(t)
	tests := []struct {
		name string
		root string
		body string
		want string // a fragment the single problem must mention
	}{
		{
			name: "off-palette fill attribute",
			root: page,
			body: `<rect x="8" y="8" width="48" height="48" fill="#ff0000"/>`,
			want: `#ff0000`,
		},
		{
			name: "off-palette colour inside a style attribute",
			root: page,
			body: `<rect x="8" y="8" width="48" height="48" style="fill:#3e8948;stroke:#123456"/>`,
			want: `#123456`,
		},
		{
			name: "linear gradient",
			root: page,
			body: `<defs><linearGradient id="g"><stop offset="0" stop-color="#3e8948"/></linearGradient></defs>`,
			want: `<linearGradient>`,
		},
		{
			name: "blur filter",
			root: page,
			body: `<defs><filter id="f"><feGaussianBlur stdDeviation="2"/></filter></defs>`,
			want: `<filter>`,
		},
		{
			name: "text",
			root: page,
			body: `<text x="8" y="40" fill="#ffffff">Hi</text>`,
			want: `<text>`,
		},
		{
			name: "embedded image",
			root: page,
			body: `<image x="0" y="0" width="64" height="64" href="leaf.png"/>`,
			want: `<image>`,
		},
		{
			name: "stroke thinner than 4 px",
			root: page,
			body: `<circle cx="32" cy="32" r="20" fill="#3e8948" stroke="#265c42" stroke-width="2"/>`,
			want: `stroke-width="2"`,
		},
		{
			name: "viewBox that is not 0 0 width height",
			root: `width="64" height="64" viewBox="0 0 32 32"`,
			body: cleanBody,
			want: `viewBox="0 0 32 32"`,
		},
		{
			name: "width that is not a whole number",
			root: `width="64.5" height="64" viewBox="0 0 64 64"`,
			body: cleanBody,
			want: `width="64.5"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rep, err := artlint.Lint(strings.NewReader(svg(tc.root, tc.body)), palette)
			if err != nil {
				t.Fatalf("Lint: %v", err)
			}
			if len(rep.Problems) != 1 {
				t.Fatalf("problems = %q, want exactly one", rep.Problems)
			}
			if !strings.Contains(rep.Problems[0], tc.want) {
				t.Errorf("problem = %q, want it to mention %q", rep.Problems[0], tc.want)
			}
		})
	}
}

func TestLoadPalette_Endesga32_LoadsAs32LowerCaseHexColours(t *testing.T) {
	palette := loadPalette(t)
	if len(palette) != 32 {
		t.Errorf("palette has %d colours, want 32", len(palette))
	}
	lowerHex := regexp.MustCompile(`^#[0-9a-f]{6}$`)
	for key := range palette {
		if !lowerHex.MatchString(key) {
			t.Errorf("palette key %q is not lower-case #rrggbb", key)
		}
	}
	// Spot-check two entries worked out by hand from the file's first and
	// last lines (190 74 47 and 194 133 105).
	for _, want := range []string{"#be4a2f", "#c28569"} {
		if !palette[want] {
			t.Errorf("palette is missing %s", want)
		}
	}
}

func TestLoadPalette_FileWithoutGIMPHeader_IsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not_a_palette.gpl")
	if err := os.WriteFile(path, []byte("#Palette Name: Endesga 32\n190\t74\t47\tbe4a2f\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := artlint.LoadPalette(path); err == nil {
		t.Error("LoadPalette accepted a file whose first line is not \"GIMP Palette\"")
	}
}
