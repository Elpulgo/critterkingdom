package synth

import (
	"math"
	"testing"
)

// syllableNames is the 20 critter-voice syllables from the spec's Decisions.
var syllableNames = []string{
	"ba", "bi", "bo", "da", "di", "do", "ka", "ki", "ko",
	"la", "li", "lo", "ma", "mi", "mo", "na", "ni", "no", "pa", "pi",
}

// fundamental estimates the pitch in Hz of the middle of a voiced sound by
// autocorrelation: the shortest lag between 1 ms and 1/60 s at which the
// sound nearly repeats itself.
func fundamental(t *testing.T, samples []float64) float64 {
	t.Helper()
	const minLag, maxLag, window = SampleRate / 1000, SampleRate / 60, SampleRate / 25
	start := len(samples) / 2
	if start+window+maxLag > len(samples) {
		t.Fatalf("%d samples is too short to measure a pitch", len(samples))
	}
	corr := make([]float64, maxLag+2)
	for lag := minLag; lag <= maxLag+1; lag++ {
		var xy, xx, yy float64
		for i := start; i < start+window; i++ {
			x, y := samples[i], samples[i+lag]
			xy += x * y
			xx += x * x
			yy += y * y
		}
		if xx > 0 && yy > 0 {
			corr[lag] = xy / math.Sqrt(xx*yy)
		}
	}
	best := 0.0
	for lag := minLag; lag <= maxLag; lag++ {
		best = max(best, corr[lag])
	}
	if best < 0.5 {
		t.Fatalf("the middle of the sound does not repeat (best correlation %.2f): no pitch", best)
	}
	// The first peak nearly as strong as the strongest is the period; later
	// peaks at two or three periods are as strong and would halve the pitch.
	for lag := minLag + 1; lag <= maxLag; lag++ {
		if corr[lag] >= 0.9*best && corr[lag] >= corr[lag-1] && corr[lag] >= corr[lag+1] {
			return float64(SampleRate) / float64(lag)
		}
	}
	t.Fatal("found no pitch")
	return 0
}

func TestVoices_HighVoiceIsPitchedAboveTheLowVoice(t *testing.T) {
	render := map[string]func() []float64{}
	for _, s := range Catalog() {
		render[s.Path] = s.Render
	}
	for _, syl := range syllableNames {
		t.Run(syl, func(t *testing.T) {
			high, okH := render["voices/high/"+syl+".ogg"]
			low, okL := render["voices/low/"+syl+".ogg"]
			if !okH || !okL {
				t.Fatal("the syllable is missing from a voice")
			}
			fh, fl := fundamental(t, high()), fundamental(t, low())
			if fh <= fl {
				t.Errorf("high voice at %.0f Hz is not above the low voice at %.0f Hz", fh, fl)
			}
		})
	}
}
