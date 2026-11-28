package synth

import (
	"math"
	"slices"
	"testing"
)

// renderLoop renders a song the way the catalog ships it.
func renderLoop(t *testing.T, song Song) []float64 {
	t.Helper()
	path := "music/" + song.Name + ".ogg"
	for _, s := range Catalog() {
		if s.Path == path {
			return s.Render()
		}
	}
	t.Fatalf("%s is not in the catalog", path)
	return nil
}

func TestMusicLoops_AreWholeBarsAtAGentleTempoAndCapped(t *testing.T) {
	for _, song := range Songs() {
		t.Run(song.Name, func(t *testing.T) {
			if song.BPM < 90 || song.BPM > 120 {
				t.Errorf("tempo %.0f BPM is outside 90 to 120", song.BPM)
			}

			samples := renderLoop(t, song)
			sec := float64(len(samples)) / SampleRate
			if sec < 45 || sec > 60 {
				t.Errorf("lasts %.2f s, want 45 to 60 s", sec)
			}

			samplesPerBar := float64(SampleRate) * 60 / song.BPM * float64(song.BeatsPerBar)
			bars := float64(len(samples)) / samplesPerBar
			if off := math.Abs(float64(len(samples)) - math.Round(bars)*samplesPerBar); off > 1 {
				t.Errorf("%d samples is %.3f bars of %d beats, not a whole number", len(samples), bars, song.BeatsPerBar)
			}

			if p := Peak(samples); p > PeakCap {
				t.Errorf("peak %.3f is over the cap of %.2f", p, PeakCap)
			} else if p < 0.05 {
				t.Errorf("peak %.3f: the loop is silent", p)
			}
		})
	}
}

func TestMusicLoops_UseOnlyNotesOfTheirScale(t *testing.T) {
	// Semitones above the root that each scale allows.
	scales := map[Mode][]int{
		Major:      {0, 2, 4, 5, 7, 9, 11},
		Pentatonic: {0, 2, 4, 7, 9},
	}
	for _, song := range Songs() {
		t.Run(song.Name, func(t *testing.T) {
			allowed, ok := scales[song.Mode]
			if !ok {
				t.Fatalf("mode %d is neither major nor pentatonic", song.Mode)
			}
			notes := song.Notes()
			if len(notes) == 0 {
				t.Fatal("has no notes")
			}
			for _, n := range notes {
				if pc := ((n.MIDI-song.Root)%12 + 12) % 12; !slices.Contains(allowed, pc) {
					t.Errorf("note %d at beat %.1f is %d semitones above the root, outside the scale", n.MIDI, n.Beat, pc)
				}
			}
		})
	}
}

func TestMusicLoops_EndMeetsStartWithoutAJump(t *testing.T) {
	const maxJump = 0.05
	for _, song := range Songs() {
		t.Run(song.Name, func(t *testing.T) {
			samples := renderLoop(t, song)
			first, last := samples[0], samples[len(samples)-1]
			if jump := math.Abs(last - first); jump > maxJump {
				t.Errorf("last sample %.4f and first sample %.4f differ by %.4f, more than %.2f: the seam clicks", last, first, jump, maxJump)
			}
		})
	}
}

func TestMusicLoop_NoteRingingPastTheEnd_CarriesOnAtTheStart(t *testing.T) {
	// One bar whose only note starts on the last beat and rings 0.3 s past
	// the end of the loop.
	song := Song{
		Name: "tail", BPM: 100, BeatsPerBar: 4, Bars: 1, Root: 60, Mode: Major,
		chords: []int{0},
		parts: []part{{
			inst: instrument{wave: Sine, amp: 1, attack: 0.01, release: 0.05, ring: 0.3},
			step: 1, bars: []string{". . . 0"},
		}},
	}

	samples := song.Render()

	if p := Peak(samples[:Samples(0.2)]); p < 0.1 {
		t.Errorf("the first 0.2 s peak at %.3f: the ringing note was cut at the end instead of wrapping to the start", p)
	}
	if jump := math.Abs(samples[len(samples)-1] - samples[0]); jump > 0.05 {
		t.Errorf("end and start differ by %.4f: the seam clicks", jump)
	}
}
