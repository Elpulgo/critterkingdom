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
	{"ui/petal_full.svg", 64, 64},
	{"ui/petal_empty.svg", 64, 64},
	// The boss's big health stars, 48 CSS px.
	{"ui/star_full.svg", 96, 96},
	{"ui/star_empty.svg", 96, 96},
	// Touch pad and paw button, sized to padRadius 64, padKnobRadius 28 and
	// pawButtonRadius 48 CSS px in internal/engine/touch.go.
	{"ui/joystick_ring.svg", 256, 256},
	{"ui/joystick_knob.svg", 112, 112},
	{"ui/paw_button.svg", 192, 192},
	// HUD icons, 32 CSS px: simplified redraws of the menu icons.
	{"ui/icons/speech_32.svg", 64, 64},
	{"ui/icons/footsteps_32.svg", 64, 64},
	{"ui/icons/berry_32.svg", 64, 64},
	{"ui/icons/hat_32.svg", 64, 64},
	{"ui/icons/heart_32.svg", 64, 64},
	{"ui/icons/star_32.svg", 64, 64},
	{"ui/icons/sticker_32.svg", 64, 64},
	{"ui/icons/backpack_32.svg", 64, 64},
	{"ui/icons/gear_32.svg", 64, 64},
	{"ui/icons/speaker_32.svg", 64, 64},
	{"ui/icons/flag_sv_32.svg", 64, 64},
	{"ui/icons/flag_en_32.svg", 64, 64},
	{"ui/icons/snacks_32.svg", 64, 64},
	{"ui/icons/key_32.svg", 64, 64},
	{"ui/icons/lantern_32.svg", 64, 64},
	{"ui/icons/zzz_32.svg", 64, 64},
	{"ui/icons/stick_32.svg", 64, 64},
	{"ui/icons/sturdy_stick_32.svg", 64, 64},
	{"ui/icons/wand_32.svg", 64, 64},
	// Menu icons, 64 CSS px.
	{"ui/icons/speech_64.svg", 128, 128},
	{"ui/icons/footsteps_64.svg", 128, 128},
	{"ui/icons/berry_64.svg", 128, 128},
	{"ui/icons/hat_64.svg", 128, 128},
	{"ui/icons/heart_64.svg", 128, 128},
	{"ui/icons/star_64.svg", 128, 128},
	{"ui/icons/sticker_64.svg", 128, 128},
	{"ui/icons/backpack_64.svg", 128, 128},
	{"ui/icons/gear_64.svg", 128, 128},
	{"ui/icons/speaker_64.svg", 128, 128},
	{"ui/icons/flag_sv_64.svg", 128, 128},
	{"ui/icons/flag_en_64.svg", 128, 128},
	{"ui/icons/snacks_64.svg", 128, 128},
	{"ui/icons/key_64.svg", 128, 128},
	{"ui/icons/lantern_64.svg", 128, 128},
	{"ui/icons/zzz_64.svg", 128, 128},
	{"ui/icons/stick_64.svg", 128, 128},
	{"ui/icons/sturdy_stick_64.svg", 128, 128},
	{"ui/icons/wand_64.svg", 128, 128},
	// Effects, one drawing each that the engine scales, turns and fades: the
	// swoosh is the player's 32 world px box, the clouds and flowers the
	// largest normal enemy (24), the spark and dust one tile (16).
	{"effects/swoosh_arc.svg", 256, 256},
	{"effects/poof_cloud.svg", 192, 192},
	{"effects/flower_burst.svg", 192, 192},
	{"effects/spore_cloud.svg", 192, 192},
	{"effects/hit_spark.svg", 128, 128},
	{"effects/dust_puff.svg", 128, 128},
	// Shared emotes, 16 world px: half the width of the 32 world px character
	// they float above.
	{"emotes/heart.svg", 128, 128},
	{"emotes/exclamation.svg", 128, 128},
	{"emotes/zzz.svg", 128, 128},
	{"emotes/sweat.svg", 128, 128},
	{"emotes/music_note.svg", 128, 128},
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
