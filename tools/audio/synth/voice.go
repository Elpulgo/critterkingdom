package synth

import "math"

// Syllables are the critter-voice syllables, in catalog order: open
// consonant-vowel syllables that read the same in Swedish and English.
var Syllables = []string{
	"ba", "bi", "bo", "da", "di", "do", "ka", "ki", "ko",
	"la", "li", "lo", "ma", "mi", "mo", "na", "ni", "no", "pa", "pi",
}

// Speaker is one critter voice that speaks every syllable.
type Speaker struct {
	// Name is the voice's folder under voices/, for example "high".
	Name string
	// Pitch is the voice's fundamental frequency in Hz in the middle of a
	// syllable.
	Pitch float64
	// FormantScale shifts every formant: a smaller critter has a smaller
	// mouth, so its vowels sit higher.
	FormantScale float64
}

// Speakers returns the two critter voices, high then low. The high voice is an
// octave above the low one.
func Speakers() []Speaker {
	return []Speaker{
		{Name: "high", Pitch: 360, FormantScale: 1.15},
		{Name: "low", Pitch: 180, FormantScale: 1},
	}
}

// vowel holds the first three formant frequencies of a vowel, in Hz.
type vowel struct{ f1, f2, f3 float64 }

var vowels = map[byte]vowel{
	'a': {800, 1250, 2600},
	'i': {320, 2300, 3000},
	'o': {480, 850, 2600},
}

// consonant describes a syllable's onset: an optional lead before the
// release (a soft voiced hum or silence), a puff of noise at the release,
// and where the formants start before they glide to the vowel.
type consonant struct {
	// lead is the time in seconds before the release.
	lead float64
	// leadAmp is the voicing level during the lead: 0 for silence (p, k),
	// low for a voiced closure (b, d), higher for a hum (m, n, l).
	leadAmp float64
	// leadF1 and leadF2 are the formants during the lead.
	leadF1, leadF2 float64
	// locusF1 and locusF2 are the formants at the release, from where they
	// glide to the vowel's.
	locusF1, locusF2 float64
	// burstFreq is the centre of the release noise in Hz, burstAmp its level
	// against the voiced sound (0 for none) and burstDur its length.
	burstFreq, burstAmp, burstDur float64
	// aspiration is the breathy gap in seconds between the release and the
	// start of the vowel's voicing (p, k).
	aspiration float64
	// seed makes the consonant's noise.
	seed uint32
}

var consonants = map[byte]consonant{
	'b': {lead: 0.02, leadAmp: 0.15, leadF1: 200, leadF2: 900, locusF1: 250, locusF2: 900, burstFreq: 700, burstAmp: 0.15, burstDur: 0.01, seed: 101},
	'p': {lead: 0.015, leadF1: 250, leadF2: 900, locusF1: 250, locusF2: 900, burstFreq: 900, burstAmp: 0.3, burstDur: 0.035, aspiration: 0.03, seed: 103},
	'd': {lead: 0.02, leadAmp: 0.15, leadF1: 200, leadF2: 1700, locusF1: 250, locusF2: 1700, burstFreq: 3000, burstAmp: 0.15, burstDur: 0.012, seed: 107},
	'k': {lead: 0.015, leadF1: 250, leadF2: 1900, locusF1: 250, locusF2: 1900, burstFreq: 1800, burstAmp: 0.3, burstDur: 0.04, aspiration: 0.035, seed: 109},
	'm': {lead: 0.06, leadAmp: 0.45, leadF1: 250, leadF2: 1000, locusF1: 300, locusF2: 1000, seed: 113},
	'n': {lead: 0.06, leadAmp: 0.45, leadF1: 250, leadF2: 1600, locusF1: 300, locusF2: 1600, seed: 127},
	'l': {lead: 0.05, leadAmp: 0.7, leadF1: 350, leadF2: 1100, locusF1: 350, locusF2: 1100, seed: 131},
}

// Syllable timing in seconds after the vowel's voicing starts.
const (
	voiceRise    = 0.012 // the vowel's voicing fades in
	formantGlide = 0.05  // the formants move from the consonant to the vowel
	vowelHold    = 0.13  // the vowel at full level
	vowelFall    = 0.045 // the vowel fades out
)

// Render synthesises one syllable, such as "ba", in this voice: a soft buzz
// at the voice's pitch through three vowel formants, after a short consonant
// onset. It panics on a syllable that is not a known consonant and vowel,
// which is a mistake in the static syllable list.
func (v Speaker) Render(syllable string) []float64 {
	if len(syllable) != 2 {
		panic("synth: syllable " + syllable + " is not a consonant and a vowel")
	}
	c, okC := consonants[syllable[0]]
	vw, okV := vowels[syllable[1]]
	if !okC || !okV {
		panic("synth: syllable " + syllable + " is not a consonant and a vowel")
	}

	release := c.lead
	voiceStart := release + c.aspiration
	end := voiceStart + vowelHold + vowelFall
	n := Samples(end)

	// Voicing level and formant tracks over time.
	amp := func(sec float64) float64 {
		switch {
		case sec < release:
			return c.leadAmp * math.Min(1, sec/0.01)
		case sec < voiceStart:
			return 0
		case sec < voiceStart+voiceRise:
			from := 0.0
			if c.aspiration == 0 {
				from = c.leadAmp
			}
			return from + (1-from)*smooth((sec-voiceStart)/voiceRise)
		case sec < voiceStart+vowelHold:
			return 1
		default:
			return 1 - smooth((sec-voiceStart-vowelHold)/vowelFall)
		}
	}
	formants := func(sec float64) (f1, f2, f3 float64) {
		if sec < release {
			return c.leadF1, c.leadF2, vw.f3
		}
		g := smooth((sec - release) / formantGlide)
		return c.locusF1 + (vw.f1-c.locusF1)*g, c.locusF2 + (vw.f2-c.locusF2)*g, vw.f3
	}

	// The voiced part: a buzz whose harmonics fall off gently, through the
	// three formants in series.
	voiced := make([]float64, n)
	r1, r2, r3 := resonator{}, resonator{}, resonator{}
	phase := 0.0
	for i := range voiced {
		sec := float64(i) / SampleRate
		// A little fall in pitch across the syllable, like a cheerful "ba!".
		f0 := v.Pitch * (1.05 - 0.1*sec/end)
		x := amp(sec) * buzz(phase, f0)
		phase += f0 / SampleRate
		f1, f2, f3 := formants(sec)
		x = r1.next(x, f1*v.FormantScale, 90)
		x = r2.next(x, f2*v.FormantScale, 110)
		voiced[i] = r3.next(x, f3*v.FormantScale, 170)
	}

	// The release: a short, soft puff of noise centred on burstFreq.
	if c.burstAmp > 0 {
		noise := Noise(c.seed, Samples(c.burstDur))
		prev := 0.0
		var r resonator
		for i, s := range noise {
			noise[i] = r.next(s-prev, c.burstFreq*v.FormantScale, 900)
			prev = s
		}
		Shape(noise, 0.001, c.burstDur/3, c.burstDur/3)
		scale := c.burstAmp * Peak(voiced) / Peak(noise)
		for i := range noise {
			noise[i] *= scale
		}
		voiced = Mix(voiced, noise, release)
	}

	return Finish(LowPass(voiced, Fixed(5000)))
}

// buzz is the voice source at phase (in cycles of f0): every harmonic below
// 5 kHz, each a little quieter than the one before, for a warm hum.
func buzz(phase, f0 float64) float64 {
	sum := 0.0
	for k := 1.0; k*f0 < 5000; k++ {
		sum += math.Sin(2*math.Pi*k*phase) / math.Pow(k, 1.3)
	}
	return sum
}

// smooth eases x from 0 to 1, clamped to that range.
func smooth(x float64) float64 {
	x = math.Max(0, math.Min(1, x))
	return x * x * (3 - 2*x)
}

// resonator is a two-pole filter that rings at one formant frequency, with
// unity gain at 0 Hz. Its frequency may change from sample to sample.
type resonator struct{ y1, y2 float64 }

// next filters one sample through a resonance at freq Hz with the given
// bandwidth in Hz.
func (r *resonator) next(x, freq, bandwidth float64) float64 {
	c := -math.Exp(-2 * math.Pi * bandwidth / SampleRate)
	b := 2 * math.Exp(-math.Pi*bandwidth/SampleRate) * math.Cos(2*math.Pi*freq/SampleRate)
	a := 1 - b - c
	y := a*x + b*r.y1 + c*r.y2
	r.y2, r.y1 = r.y1, y
	return y
}
