// Package synth makes every sound in the game from code: soft oscillators,
// envelopes, a gentle low-pass and a peak cap. Every sound is a mono slice of
// samples in -1..1 at SampleRate, rendered without randomness (noise comes
// from a fixed seed), so the same code always gives the same file.
package synth

import "math"

// SampleRate is the sample rate of every rendered sound, in Hz.
const SampleRate = 44100

// PeakCap is the loudest any rendered sound may get, as a fraction of full
// scale. It stands in for the master volume cap: nothing is ever harsh.
const PeakCap = 0.7

// targetPeak is the peak Finish scales a sound to, safely under PeakCap.
const targetPeak = 0.6

// edgeFade is how long Finish fades a sound in and out, in seconds, so it
// never starts or ends with a click.
const edgeFade = 0.004

// Wave is one cycle of an oscillator: it maps a phase in cycles (any real
// number, one cycle per 1.0) to an amplitude in -1..1.
type Wave func(phase float64) float64

// Sine is the softest wave.
func Sine(phase float64) float64 {
	return math.Sin(2 * math.Pi * phase)
}

// Triangle is a sine-phased triangle: a little brighter than Sine, still soft.
func Triangle(phase float64) float64 {
	p := phase + 0.25
	p -= math.Floor(p)
	return 1 - 4*math.Abs(p-0.5)
}

// RoundSquare is a square wave with rounded corners, for a warm buzz
// without the harsh edges of a true square.
func RoundSquare(phase float64) float64 {
	return math.Tanh(3*Sine(phase)) / math.Tanh(3)
}

// Bell is a sine with a faint inharmonic overtone, for chimes and sparkles.
func Bell(phase float64) float64 {
	return (Sine(phase) + 0.25*Sine(2.76*phase) + 0.1*Sine(5.4*phase)) / 1.35
}

// Hz returns the frequency of a MIDI note number (69 is A4 at 440 Hz).
func Hz(midi float64) float64 {
	return 440 * math.Pow(2, (midi-69)/12)
}

// Samples returns the number of samples in the given number of seconds.
func Samples(seconds float64) int {
	return int(math.Round(seconds * SampleRate))
}

// Tone is one note: an oscillator with an optional pitch glide and vibrato,
// shaped by an attack, an exponential decay and a release.
type Tone struct {
	Wave Wave
	// Freq is the starting frequency in Hz.
	Freq float64
	// EndFreq, when set, glides the pitch exponentially from Freq to EndFreq
	// over the whole note.
	EndFreq float64
	// Dur is the note length in seconds, release included.
	Dur float64
	// Amp is the peak amplitude before Finish scales the sound.
	Amp float64
	// Attack is the fade-in time in seconds.
	Attack float64
	// Decay, when set, is the time in seconds for the level to fall to 1/e.
	Decay float64
	// Release is the fade-out time in seconds at the end of the note.
	Release float64
	// VibratoRate in Hz and VibratoDepth in semitones wobble the pitch.
	VibratoRate, VibratoDepth float64
}

// Render returns the note's samples. Every note starts and ends at silence.
func (t Tone) Render() []float64 {
	n := Samples(t.Dur)
	out := make([]float64, n)
	wave := t.Wave
	if wave == nil {
		wave = Sine
	}
	phase := 0.0
	for i := range out {
		sec := float64(i) / SampleRate
		frac := 0.0
		if n > 1 {
			frac = float64(i) / float64(n-1)
		}
		f := t.Freq
		if t.EndFreq > 0 {
			f = t.Freq * math.Pow(t.EndFreq/t.Freq, frac)
		}
		if t.VibratoDepth != 0 {
			f *= math.Pow(2, t.VibratoDepth*math.Sin(2*math.Pi*t.VibratoRate*sec)/12)
		}
		out[i] = t.Amp * envelope(sec, t.Dur, t.Attack, t.Decay, t.Release) * wave(phase)
		phase += f / SampleRate
	}
	return out
}

// envelope is the level at time sec of a note of length dur.
func envelope(sec, dur, attack, decay, release float64) float64 {
	level := 1.0
	if attack > 0 {
		level = math.Min(level, sec/attack)
	}
	if release > 0 {
		level = math.Min(level, (dur-sec)/release)
	}
	if level < 0 {
		level = 0
	}
	if decay > 0 {
		level *= math.Exp(-sec / decay)
	}
	return level
}

// Noise returns n samples of white noise in -1..1 from a fixed seed, so the
// same seed always gives the same noise.
func Noise(seed uint32, n int) []float64 {
	if seed == 0 {
		seed = 1
	}
	out := make([]float64, n)
	x := seed
	for i := range out {
		// xorshift32
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		out[i] = float64(x)/float64(math.MaxUint32)*2 - 1
	}
	return out
}

// LowPass smooths samples in place with a one-pole low-pass filter whose
// cutoff in Hz may change over time: cutoff gets the time in seconds.
func LowPass(samples []float64, cutoff func(sec float64) float64) []float64 {
	y := 0.0
	for i, x := range samples {
		fc := cutoff(float64(i) / SampleRate)
		a := 1 - math.Exp(-2*math.Pi*fc/SampleRate)
		y += a * (x - y)
		samples[i] = y
	}
	return samples
}

// Fixed returns a cutoff function for LowPass that stays at hz.
func Fixed(hz float64) func(float64) float64 {
	return func(float64) float64 { return hz }
}

// Shape multiplies samples in place by an envelope with the given attack,
// decay and release in seconds, as Tone does.
func Shape(samples []float64, attack, decay, release float64) []float64 {
	dur := float64(len(samples)) / SampleRate
	for i := range samples {
		samples[i] *= envelope(float64(i)/SampleRate, dur, attack, decay, release)
	}
	return samples
}

// Mix adds src into dst starting at the given time in seconds, growing dst
// when src runs past its end, and returns dst.
func Mix(dst, src []float64, at float64) []float64 {
	start := Samples(at)
	if need := start + len(src); need > len(dst) {
		dst = append(dst, make([]float64, need-len(dst))...)
	}
	for i, s := range src {
		dst[start+i] += s
	}
	return dst
}

// Peak returns the largest absolute sample value.
func Peak(samples []float64) float64 {
	p := 0.0
	for _, s := range samples {
		p = math.Max(p, math.Abs(s))
	}
	return p
}

// Finish makes a sound safe to ship: it removes any DC offset, fades the
// first and last few milliseconds so the sound starts and ends at exactly
// zero, and scales it to a peak a little under PeakCap.
func Finish(samples []float64) []float64 {
	if len(samples) == 0 {
		return samples
	}
	mean := 0.0
	for _, s := range samples {
		mean += s
	}
	mean /= float64(len(samples))
	for i := range samples {
		samples[i] -= mean
	}

	fade := min(Samples(edgeFade), len(samples)/2)
	for i := range fade {
		g := 0.5 - 0.5*math.Cos(math.Pi*float64(i)/float64(fade))
		samples[i] *= g
		samples[len(samples)-1-i] *= g
	}

	if p := Peak(samples); p > 0 {
		scale := targetPeak / p
		for i := range samples {
			samples[i] *= scale
		}
	}
	return samples
}
