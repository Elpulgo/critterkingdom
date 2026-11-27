package synth

import (
	"math"
	"slices"
	"testing"
)

func TestCatalog_HoldsExactlyTheTwentyOneSoundEffects(t *testing.T) {
	want := []string{
		"sfx/step.ogg", "sfx/talk_blip.ogg", "sfx/pickup.ogg", "sfx/door.ogg",
		"sfx/chest.ogg", "sfx/quest_start.ogg", "sfx/quest_done.ogg",
		"sfx/befriend.ogg", "sfx/wrong_snack.ogg", "sfx/menu_tap.ogg",
		"sfx/hearts.ogg", "sfx/sparkle.ogg", "sfx/zzz.ogg", "sfx/swing.ogg",
		"sfx/hit_pop.ogg", "sfx/hurt.ogg", "sfx/poof.ogg", "sfx/heart_pickup.ogg",
		"sfx/room_lock.ogg", "sfx/room_unlock.ogg", "sfx/grub_grumble.ogg",
	}

	var got []string
	for _, s := range Catalog() {
		if s.Kind != SFX {
			continue
		}
		if s.Render == nil {
			t.Errorf("%s has no render function", s.Path)
		}
		got = append(got, s.Path)
	}

	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("sound effects in the catalog:\n got %q\nwant %q", got, want)
	}
}

func TestSoundEffects_AreSoftClickFreeAndShort(t *testing.T) {
	const maxSeconds = 2.5
	for _, s := range Catalog() {
		if s.Kind != SFX {
			continue
		}
		t.Run(s.Path, func(t *testing.T) {
			samples := s.Render()
			if len(samples) == 0 {
				t.Fatal("renders no samples")
			}
			if p := Peak(samples); p > PeakCap {
				t.Errorf("peak %.3f is over the cap of %.2f", p, PeakCap)
			} else if p < 0.05 {
				t.Errorf("peak %.3f: the sound is silent", p)
			}
			if first := samples[0]; math.Abs(first) > 0.01 {
				t.Errorf("first sample %.4f is not silent, it would click", first)
			}
			if last := samples[len(samples)-1]; math.Abs(last) > 0.01 {
				t.Errorf("last sample %.4f is not silent, it would click", last)
			}
			if sec := float64(len(samples)) / SampleRate; sec > maxSeconds {
				t.Errorf("lasts %.2f s, more than %.1f s", sec, maxSeconds)
			}
		})
	}
}

func TestSoundEffects_RenderTheSameSamplesEveryTime(t *testing.T) {
	for _, s := range Catalog() {
		if s.Kind != SFX {
			continue
		}
		t.Run(s.Path, func(t *testing.T) {
			if !slices.Equal(s.Render(), s.Render()) {
				t.Error("two renders differ")
			}
		})
	}
}
