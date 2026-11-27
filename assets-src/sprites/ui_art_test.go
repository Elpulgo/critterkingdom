package sprites_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/tools/art/artlint"
)

// uiArt is the manifest of every UI, effect and emote drawing: its path under
// assets-src/sprites and its exact page size in drawing px. UI art is drawn at
// 2 px per CSS px, effects and emotes at 8 px per world px.
var uiArt = []struct {
	path          string
	width, height int
}{
	// 9-slice sources, 48 CSS px with 16 CSS px insets.
	{"ui/panel.svg", 96, 96},
	{"ui/button.svg", 96, 96},
	{"ui/button_pressed.svg", 96, 96},
	// HUD size, 32 CSS px.
	{"ui/arrow_next.svg", 64, 64},
	{"ui/checkmark.svg", 64, 64},
}

// uiArtFolders are the folders whose every .svg must be a manifest row.
var uiArtFolders = []string{"ui", "effects", "emotes"}

func TestUIArt_EveryListedDrawingExistsAtItsSizeAndFollowsTheArtBible(t *testing.T) {
	palette, err := artlint.LoadPalette("../palettes/endesga_32.gpl")
	if err != nil {
		t.Fatalf("load palette: %v", err)
	}

	for _, art := range uiArt {
		t.Run(art.path, func(t *testing.T) {
			f, err := os.Open(filepath.FromSlash(art.path))
			if err != nil {
				t.Fatalf("listed drawing is missing: %v", err)
			}
			defer f.Close()

			rep, err := artlint.Lint(f, palette)
			if err != nil {
				t.Fatalf("lint: %v", err)
			}
			for _, p := range rep.Problems {
				t.Errorf("art lint: %s", p)
			}
			if rep.Width != art.width || rep.Height != art.height {
				t.Errorf("page is %d by %d px, want %d by %d px", rep.Width, rep.Height, art.width, art.height)
			}
		})
	}
}

func TestUIArt_UnlistedDrawingsAreRejected(t *testing.T) {
	listed := make(map[string]bool, len(uiArt))
	for _, art := range uiArt {
		listed[art.path] = true
	}

	for _, folder := range uiArtFolders {
		err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if path == folder && errors.Is(err, fs.ErrNotExist) {
					return filepath.SkipDir // a folder with no drawings yet
				}
				return err
			}
			if d.IsDir() || !strings.EqualFold(filepath.Ext(path), ".svg") {
				return nil
			}
			if rel := filepath.ToSlash(path); !listed[rel] {
				t.Errorf("%s is not in the uiArt manifest; add a row with its page size", rel)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", folder, err)
		}
	}
}
