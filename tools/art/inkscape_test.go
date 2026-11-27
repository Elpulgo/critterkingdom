//go:build inkscape

// This test runs the real Inkscape (an external program not installed in
// CI), so it only builds with the tag: go test -tags inkscape ./tools/art/...

package main_test

import (
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/Elpulgo/critterkingdom/tools/art/artexport"
)

func TestInkscapeExport_PartPNGsHaveTheSizesTheJSONDescribes(t *testing.T) {
	out := t.TempDir()
	err := artexport.Run(artexport.Options{
		SrcDir:   filepath.Join("testdata", "src"),
		OutDir:   out,
		Inkscape: artexport.NewInkscape("inkscape"),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	dir := filepath.Join(out, "sprites", "critters", "blob")
	data, err := os.ReadFile(filepath.Join(dir, "blob.json"))
	if err != nil {
		t.Fatal(err)
	}
	var sheet artexport.Sheet
	if err := json.Unmarshal(data, &sheet); err != nil {
		t.Fatalf("decode blob.json: %v", err)
	}
	parts := sheet.Views["down"].Parts
	if len(parts) != 2 || parts[0].Name != "body" || parts[1].Name != "head" {
		t.Fatalf("down parts = %+v, want body then head", parts)
	}
	// The body's fractional box snaps outward to 96 by 91 px.
	if parts[0].W != 12 || parts[0].H != 11.375 {
		t.Errorf("body w, h = %v, %v; want 12, 11.375", parts[0].W, parts[0].H)
	}

	for _, p := range parts {
		w, h := pngSize(t, filepath.Join(dir, p.Image))
		if float64(w) != p.W*8 || float64(h) != p.H*8 {
			t.Errorf("%s is %d by %d px; the JSON says %v by %v", p.Image, w, h, p.W*8, p.H*8)
		}
	}

	for _, marker := range []string{"down_head_pivot.png", "down_feet.png"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); !os.IsNotExist(err) {
			t.Errorf("%s was exported; markers must not become PNGs (stat error: %v)", marker, err)
		}
	}
}

func pngSize(t *testing.T, path string) (int, int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return cfg.Width, cfg.Height
}
