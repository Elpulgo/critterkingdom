package synth

import (
	"math"
	"slices"
	"testing"
)

func TestCatalog_HoldsExactlyTheSpecifiedSoundsOfEachKind(t *testing.T) {
	var voices []string
	for _, voice := range []string{"high", "low"} {
		for _, syl := range syllableNames {
			voices = append(voices, "voices/"+voice+"/"+syl+".ogg")
		}
	}
	tests := []struct {
		name string
		kind Kind
		want []string
	}{
		{"twenty-one sound effects", SFX, []string{
			"sfx/step.ogg", "sfx/talk_blip.ogg", "sfx/pickup.ogg", "sfx/door.ogg",
			"sfx/chest.ogg", "sfx/quest_start.ogg", "sfx/quest_done.ogg",
			"sfx/befriend.ogg", "sfx/wrong_snack.ogg", "sfx/menu_tap.ogg",
			"sfx/hearts.ogg", "sfx/sparkle.ogg", "sfx/zzz.ogg", "sfx/swing.ogg",
			"sfx/hit_pop.ogg", "sfx/hurt.ogg", "sfx/poof.ogg", "sfx/heart_pickup.ogg",
			"sfx/room_lock.ogg", "sfx/room_unlock.ogg", "sfx/grub_grumble.ogg",
		}},
		{"four music loops", Music, []string{
			"music/village.ogg", "music/meadow.ogg", "music/cave.ogg", "music/boss.ogg",
		}},
		{"twenty syllables in a high and a low voice", Voice, voices},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, s := range Catalog() {
				if s.Kind != tt.kind {
					continue
				}
				if s.Render == nil {
					t.Errorf("%s has no render function", s.Path)
				}
				got = append(got, s.Path)
			}
			want := slices.Clone(tt.want)
			slices.Sort(want)
			slices.Sort(got)
			if !slices.Equal(got, want) {
				t.Errorf("catalog:\n got %q\nwant %q", got, want)
			}
		})
	}
}

func TestShortSounds_AreSoftClickFreeAndShort(t *testing.T) {
	// The longest each kind of short sound may last, in seconds.
	maxSeconds := map[Kind]float64{SFX: 2.5, Voice: 0.3}
	for _, s := range Catalog() {
		limit, ok := maxSeconds[s.Kind]
		if !ok {
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
			if sec := float64(len(samples)) / SampleRate; sec > limit {
				t.Errorf("lasts %.2f s, more than %.1f s", sec, limit)
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
