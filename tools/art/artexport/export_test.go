package artexport_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Elpulgo/critterkingdom/tools/art/artexport"
)

// fakeInkscape stands in for the Inkscape process (an external program).
type fakeInkscape struct {
	queries map[string]string // SVG path -> canned --query-all stdout
	calls   []string          // SVG path of every call, in order
}

func (f *fakeInkscape) QueryAll(svg string) ([]byte, error) {
	f.calls = append(f.calls, svg)
	out, ok := f.queries[svg]
	if !ok {
		return nil, errors.New("fake inkscape: no query output for " + svg)
	}
	return []byte(out), nil
}

func (f *fakeInkscape) ExportID(svg, id, png string) error {
	f.calls = append(f.calls, svg)
	return os.WriteFile(png, []byte("png of "+id), 0o644)
}

func (f *fakeInkscape) ExportPage(svg, png string) error {
	f.calls = append(f.calls, svg)
	return os.WriteFile(png, []byte("png of page"), 0o644)
}

// foxQuery is the worked example from the spec: two parts in the down view.
const foxQuery = "svg1,0,0,256,256\n" +
	"layer1,0,0,256,256\n" +
	"down_feet,124,236,8,8\n" +
	"down_body,80.5,150.2,95,90.3\n" +
	"down_head,64,40,128,120\n" +
	"down_head_pivot,126,146,4,4\n"

var (
	sourceTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	laterTime  = time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
)

// writeSource writes a placeholder SVG at rel under src with the given mtime
// and returns its path.
func writeSource(t *testing.T, src, rel string, mtime time.Time) string {
	t.Helper()
	p := filepath.Join(src, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	setMtime(t, p, mtime)
	return p
}

func setMtime(t *testing.T, p string, mtime time.Time) {
	t.Helper()
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

// outputFiles lists every file under out as sorted slash paths relative to out.
func outputFiles(t *testing.T, out string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(out, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(out, p)
		files = append(files, filepath.ToSlash(rel))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return files
}

// calledSVGs returns the distinct SVG paths Inkscape was called with, sorted.
func calledSVGs(ink *fakeInkscape) []string {
	seen := map[string]bool{}
	var svgs []string
	for _, c := range ink.calls {
		if !seen[c] {
			seen[c] = true
			svgs = append(svgs, c)
		}
	}
	sort.Strings(svgs)
	return svgs
}

func TestRun_CharacterDrawing_WritesOnePNGPerPartAndThePartJSON(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	fox := writeSource(t, src, "sprites/critters/fox.svg", sourceTime)
	ink := &fakeInkscape{queries: map[string]string{fox: foxQuery}}

	if err := artexport.Run(artexport.Options{SrcDir: src, OutDir: out, Inkscape: ink}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	wantFiles := []string{
		"sprites/critters/fox/down_body.png",
		"sprites/critters/fox/down_head.png",
		"sprites/critters/fox/fox.json",
	}
	if got := outputFiles(t, out); !reflect.DeepEqual(got, wantFiles) {
		t.Errorf("output files = %v, want %v", got, wantFiles)
	}

	wantJSON := `{
  "pxPerWorldPixel": 8,
  "views": {
    "down": {
      "parts": [
        {
          "name": "body",
          "image": "down_body.png",
          "x": -6,
          "y": -11.25,
          "w": 12,
          "h": 11.375,
          "pivotX": 0,
          "pivotY": -5.5625
        },
        {
          "name": "head",
          "image": "down_head.png",
          "x": -8,
          "y": -25,
          "w": 16,
          "h": 15,
          "pivotX": 0,
          "pivotY": -11.5
        }
      ]
    }
  }
}
`
	got, err := os.ReadFile(filepath.Join(out, "sprites/critters/fox/fox.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != wantJSON {
		t.Errorf("fox.json =\n%s\nwant\n%s", got, wantJSON)
	}
}

func TestRun_PropsAndTiles_ExportOnePagePNGAtTheMirroredPath_AndOtherFoldersAreIgnored(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	writeSource(t, src, "sprites/props/tree.svg", sourceTime)
	writeSource(t, src, "tiles/grass.svg", sourceTime)
	writeSource(t, src, "style/style_sheet.svg", sourceTime)
	ink := &fakeInkscape{}

	if err := artexport.Run(artexport.Options{SrcDir: src, OutDir: out, Inkscape: ink}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []string{"sprites/props/tree.png", "tiles/grass.png"}
	if got := outputFiles(t, out); !reflect.DeepEqual(got, want) {
		t.Errorf("output files = %v, want %v", got, want)
	}
}

func TestRun_SourceRootWithoutSpritesOrTiles_IsNotAnError(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	ink := &fakeInkscape{}

	if err := artexport.Run(artexport.Options{SrcDir: src, OutDir: out, Inkscape: ink}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(ink.calls) != 0 {
		t.Errorf("Inkscape calls = %v, want none", ink.calls)
	}
}

func TestRun_OnlyChangedSourcesAreExported_UnlessForced(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	fox := writeSource(t, src, "sprites/critters/fox.svg", sourceTime)
	tree := writeSource(t, src, "sprites/props/tree.svg", sourceTime)
	ink := &fakeInkscape{queries: map[string]string{fox: foxQuery}}
	opts := artexport.Options{SrcDir: src, OutDir: out, Inkscape: ink}

	if err := artexport.Run(opts); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	// Outputs as old as their sources still count as up to date.
	setMtime(t, filepath.Join(out, "sprites/critters/fox/fox.json"), sourceTime)
	setMtime(t, filepath.Join(out, "sprites/props/tree.png"), sourceTime)

	ink.calls = nil
	if err := artexport.Run(opts); err != nil {
		t.Fatalf("unchanged Run: %v", err)
	}
	if len(ink.calls) != 0 {
		t.Errorf("unchanged sources: Inkscape calls = %v, want none", ink.calls)
	}

	setMtime(t, fox, laterTime)
	ink.calls = nil
	if err := artexport.Run(opts); err != nil {
		t.Fatalf("Run after saving fox: %v", err)
	}
	if got, want := calledSVGs(ink), []string{fox}; !reflect.DeepEqual(got, want) {
		t.Errorf("after saving fox: exported %v, want %v", got, want)
	}

	opts.Force = true
	ink.calls = nil
	if err := artexport.Run(opts); err != nil {
		t.Fatalf("forced Run: %v", err)
	}
	want := []string{fox, tree}
	sort.Strings(want)
	if got := calledSVGs(ink); !reflect.DeepEqual(got, want) {
		t.Errorf("forced: exported %v, want %v", got, want)
	}
}

func TestRun_PartRemovedFromDrawing_HasItsStalePNGDeleted(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	fox := writeSource(t, src, "sprites/critters/fox.svg", sourceTime)
	ink := &fakeInkscape{queries: map[string]string{fox: foxQuery + "down_tail,100,200,20,20\n"}}
	opts := artexport.Options{SrcDir: src, OutDir: out, Inkscape: ink}
	if err := artexport.Run(opts); err != nil {
		t.Fatalf("first Run: %v", err)
	}

	ink.queries[fox] = foxQuery
	opts.Force = true
	if err := artexport.Run(opts); err != nil {
		t.Fatalf("Run after removing the tail: %v", err)
	}

	want := []string{
		"sprites/critters/fox/down_body.png",
		"sprites/critters/fox/down_head.png",
		"sprites/critters/fox/fox.json",
	}
	if got := outputFiles(t, out); !reflect.DeepEqual(got, want) {
		t.Errorf("output files = %v, want %v", got, want)
	}
}

func TestRun_DrawingThatFailsValidation_WritesNoJSONAndTheErrorNamesTheSVG(t *testing.T) {
	src, out := t.TempDir(), t.TempDir()
	fox := writeSource(t, src, "sprites/critters/fox.svg", sourceTime)
	// The down view has parts but no down_feet anchor.
	ink := &fakeInkscape{queries: map[string]string{fox: "down_body,80,150,96,88\n"}}

	err := artexport.Run(artexport.Options{SrcDir: src, OutDir: out, Inkscape: ink})
	if err == nil {
		t.Fatal("Run returned nil, want an error naming fox.svg")
	}
	if !strings.Contains(err.Error(), fox) {
		t.Errorf("error = %q, want it to name %s", err, fox)
	}
	if got := outputFiles(t, out); len(got) != 0 {
		t.Errorf("output files = %v, want none", got)
	}
}
