package synth

import (
	"fmt"
	"strconv"
	"strings"
)

// Mode is the scale a song's notes come from.
type Mode int

const (
	// Major is the seven-note major scale.
	Major Mode = iota
	// Pentatonic is the five-note major pentatonic scale.
	Pentatonic
)

// steps returns the scale's notes as semitones above its root.
func (m Mode) steps() []int {
	if m == Pentatonic {
		return []int{0, 2, 4, 7, 9}
	}
	return []int{0, 2, 4, 5, 7, 9, 11}
}

// Note is one pitched note of a song.
type Note struct {
	// Beat is when the note starts, in beats from the start of the loop.
	Beat float64
	// Beats is the note's written length in beats.
	Beats float64
	// MIDI is the note's pitch as a MIDI note number.
	MIDI int
}

// Song is one music loop: a few instrument parts over a repeating chord
// progression, a whole number of bars long. Notes that ring past the end of
// the loop wrap round to its start, so the loop plays seamlessly with no
// fade-out.
type Song struct {
	// Name is the file name without extension, for example "village".
	Name string
	// BPM is the tempo in beats per minute.
	BPM float64
	// BeatsPerBar is the time signature: 3 or 4 beats to a bar.
	BeatsPerBar int
	// Bars is the loop's length in bars.
	Bars int
	// Root is the MIDI note of the scale's first degree, for example 60 for C.
	Root int
	// Mode is the scale every pitched note comes from.
	Mode Mode

	// chords is the scale degree each bar's chord is built on, cycled.
	chords []int
	parts  []part
	// drums is the unpitched percussion, if any.
	drums *drumPart
}

// instrument is how a part's notes sound.
type instrument struct {
	wave                   Wave
	amp                    float64
	attack, decay, release float64
	// ring is how long, in seconds, a note keeps sounding past its written
	// length.
	ring                      float64
	vibratoRate, vibratoDepth float64
}

// echo repeats a part's notes later and quieter, like a cave's echo.
type echo struct {
	beats   float64
	gain    float64
	repeats int
}

// part is one instrument line. Each string in bars is one bar of
// space-separated steps, step beats apart, cycled through the song:
//
//   - a scale degree (0 is the root, 7 the octave in Major and 5 in
//     Pentatonic; negative degrees go down) starts a note;
//   - degrees joined by "+" start a chord;
//   - "-" holds the notes before it for one more step, across bar lines too;
//   - "." is a rest.
type part struct {
	inst instrument
	// base shifts the part by whole octaves, in semitones (a multiple of 12).
	base int
	step float64
	// follow adds the bar's chord degree to every degree, so the line
	// follows the chord progression.
	follow bool
	bars   []string
	echo   echo
}

// drumPart is unpitched percussion: each bar is space-separated steps of
// "k" (a soft kick), "s" (a shaker), "ks" (both) or "." (nothing).
type drumPart struct {
	step float64
	bars []string
}

// Seconds returns the loop's length in seconds.
func (s Song) Seconds() float64 {
	return float64(s.Bars*s.BeatsPerBar) * 60 / s.BPM
}

// Notes returns every pitched note of the song, part by part. It panics on a
// malformed bar, which is a mistake in the song's code.
func (s Song) Notes() []Note {
	var all []Note
	for _, p := range s.parts {
		all = append(all, s.partNotes(p)...)
	}
	return all
}

// partNotes turns a part's bars into notes.
func (s Song) partNotes(p part) []Note {
	steps := int(float64(s.BeatsPerBar)/p.step + 0.5)
	var out []Note
	var held []int // indices in out of the notes a "-" extends
	for bar := range s.Bars {
		tokens := strings.Fields(p.bars[bar%len(p.bars)])
		if len(tokens) != steps {
			panic(fmt.Sprintf("synth: song %s bar %q has %d steps, want %d", s.Name, p.bars[bar%len(p.bars)], len(tokens), steps))
		}
		shift := 0
		if p.follow {
			shift = s.chords[bar%len(s.chords)]
		}
		for i, tok := range tokens {
			beat := float64(bar*s.BeatsPerBar) + float64(i)*p.step
			switch tok {
			case ".":
				held = nil
			case "-":
				for _, h := range held {
					out[h].Beats += p.step
				}
			default:
				held = nil
				for _, d := range strings.Split(tok, "+") {
					deg, err := strconv.Atoi(d)
					if err != nil {
						panic(fmt.Sprintf("synth: song %s has a bad step %q", s.Name, tok))
					}
					out = append(out, Note{Beat: beat, Beats: p.step, MIDI: s.pitch(deg+shift) + p.base})
					held = append(held, len(out)-1)
				}
			}
		}
	}
	return out
}

// pitch returns the MIDI note of a scale degree of the song.
func (s Song) pitch(deg int) int {
	steps := s.Mode.steps()
	n := len(steps)
	oct := deg / n
	if deg%n < 0 {
		oct--
	}
	return s.Root + 12*oct + steps[deg-oct*n]
}

// Render synthesises the loop: mono samples in -1..1 at SampleRate, exactly
// Seconds long, with note tails wrapped round to the start.
func (s Song) Render() []float64 {
	secPerBeat := 60 / s.BPM
	out := make([]float64, Samples(s.Seconds()))
	for _, p := range s.parts {
		for _, n := range s.partNotes(p) {
			in := p.inst
			tone := Tone{
				Wave: in.wave, Freq: Hz(float64(n.MIDI)), Dur: n.Beats*secPerBeat + in.ring,
				Amp: in.amp, Attack: in.attack, Decay: in.decay, Release: in.release,
				VibratoRate: in.vibratoRate, VibratoDepth: in.vibratoDepth,
			}.Render()
			start := n.Beat * secPerBeat
			mixLoop(out, tone, start, 1)
			gain := 1.0
			for r := 1; r <= p.echo.repeats; r++ {
				gain *= p.echo.gain
				mixLoop(out, tone, start+float64(r)*p.echo.beats*secPerBeat, gain)
			}
		}
	}
	if s.drums != nil {
		kick, shaker := kickHit(), shakerHit()
		for bar := range s.Bars {
			for i, tok := range strings.Fields(s.drums.bars[bar%len(s.drums.bars)]) {
				at := (float64(bar*s.BeatsPerBar) + float64(i)*s.drums.step) * secPerBeat
				if strings.Contains(tok, "k") {
					mixLoop(out, kick, at, 1)
				}
				if strings.Contains(tok, "s") {
					mixLoop(out, shaker, at, 1)
				}
			}
		}
	}
	return finishLoop(out)
}

// mixLoop adds src times gain into the loop dst starting at the given time in
// seconds, wrapping anything past the end of the loop round to its start.
func mixLoop(dst, src []float64, at, gain float64) {
	start := Samples(at)
	for i, v := range src {
		dst[(start+i)%len(dst)] += gain * v
	}
}

// finishLoop is Finish for a loop: it removes any DC offset and scales the
// loop to a peak a little under PeakCap, but does not fade the ends, which
// already meet seamlessly.
func finishLoop(samples []float64) []float64 {
	if len(samples) == 0 {
		return samples
	}
	mean := 0.0
	for _, v := range samples {
		mean += v
	}
	mean /= float64(len(samples))
	for i := range samples {
		samples[i] -= mean
	}
	if p := Peak(samples); p > 0 {
		for i := range samples {
			samples[i] *= targetPeak / p
		}
	}
	return samples
}

// kickHit is a soft, round kick drum: a low sine thump dropping in pitch.
func kickHit() []float64 {
	return Tone{Freq: 120, EndFreq: 55, Dur: 0.16, Amp: 0.5, Attack: 0.003, Decay: 0.06, Release: 0.04}.Render()
}

// shakerHit is a quiet, muffled shaker tick of filtered noise.
func shakerHit() []float64 {
	out := filteredNoise(37, 0.05, Fixed(5000), 0.003, 0.012, 0.015)
	for i := range out {
		out[i] *= 0.35
	}
	return out
}

// The instruments of the music loops: all soft waves with gentle attacks.
var (
	musicBox   = instrument{wave: Bell, amp: 0.35, attack: 0.004, decay: 0.4, release: 0.1, ring: 0.4}
	waterDrop  = instrument{wave: Bell, amp: 0.3, attack: 0.004, decay: 0.6, release: 0.2, ring: 1.0}
	softBass   = instrument{wave: Triangle, amp: 0.5, attack: 0.01, decay: 0.6, release: 0.06}
	bounceBass = instrument{wave: RoundSquare, amp: 0.22, attack: 0.005, decay: 0.12, release: 0.03}
	pluck      = instrument{wave: Triangle, amp: 0.16, attack: 0.004, decay: 0.15, release: 0.04}
	lead       = instrument{wave: Triangle, amp: 0.32, attack: 0.006, decay: 0.22, release: 0.04}
	flute      = instrument{wave: Sine, amp: 0.4, attack: 0.04, decay: 1.5, release: 0.08, vibratoRate: 5, vibratoDepth: 0.12}
	pad        = instrument{wave: Sine, amp: 0.12, attack: 0.4, release: 0.5, ring: 0.3}
)

// Songs returns the four music loops in catalog order: village, meadow, cave
// and boss.
func Songs() []Song {
	return []Song{village(), meadow(), cave(), boss()}
}

// concat joins lists of bars into one.
func concat(lists ...[]string) []string {
	var out []string
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// village is a cosy music-box tune in C major: 24 bars of 4/4 at 100 BPM,
// 57.6 s. The melody plays A, B, A over I vi IV V I vi ii V.
func village() Song {
	a := []string{
		"4 - 2 - 0 - 2 4",
		"5 - - 4 5 - 7 -",
		"3 - 5 - 7 - 5 -",
		"6 - 4 - 1 - . .",
		"7 - 4 - 2 - 4 -",
		"5 - 4 - 2 - 0 -",
		"1 - 3 - 5 - 3 -",
		"4 - - - - - . .",
	}
	b := []string{
		"2 - 4 - 7 - - -",
		"7 - 5 - 2 - - -",
		"5 - 3 - 0 - 3 -",
		"4 - - 6 4 - - -",
		"0 - 2 - 4 - 7 -",
		"5 - 4 - 2 - 4 -",
		"3 - 1 - 3 - 5 -",
		"4 - 1 - 4 - . .",
	}
	return Song{
		Name: "village", BPM: 100, BeatsPerBar: 4, Bars: 24, Root: 60, Mode: Major,
		chords: []int{0, 5, 3, 4, 0, 5, 1, 4},
		parts: []part{
			{inst: softBass, base: -12, step: 1, follow: true, bars: []string{"0 . 4 .", "0 . 4 .", "0 . 4 .", "0 2 4 2"}},
			{inst: pluck, base: 0, step: 0.5, follow: true, bars: []string{"0 4 2 4 7 4 2 4"}},
			{inst: musicBox, base: 12, step: 0.5, bars: concat(a, b)},
		},
		drums: &drumPart{step: 0.5, bars: []string{"k . s . k . s s"}},
	}
}

// meadow is a lilting flute waltz in F major pentatonic: 32 bars of 3/4 at
// 105 BPM, 54.9 s. The melody plays A, B, A, B.
func meadow() Song {
	a := []string{
		"2 - 3 - 5 -",
		"3 - - - 2 -",
		"4 - 3 - 4 -",
		"5 - - - - -",
		"6 - 4 - 3 -",
		"4 - - - 1 -",
		"3 - 1 - 0 -",
		"1 - - - . .",
	}
	b := []string{
		"5 - 7 - 5 -",
		"3 - 2 - 3 -",
		"4 - 5 - 7 -",
		"6 - 5 - 4 -",
		"3 - 4 - 6 -",
		"4 - - - 3 -",
		"1 - 2 - 3 -",
		"3 - - - . .",
	}
	return Song{
		Name: "meadow", BPM: 105, BeatsPerBar: 3, Bars: 32, Root: 65, Mode: Pentatonic,
		chords: []int{0, 0, 4, 4, 1, 1, 3, 3},
		parts: []part{
			{inst: softBass, base: -24, step: 1, follow: true, bars: []string{"0 . .", "3 . ."}},
			{inst: pluck, base: -12, step: 1, follow: true, bars: []string{". 2+3 2+3"}},
			{inst: flute, base: 0, step: 0.5, bars: concat(a, b)},
		},
		drums: &drumPart{step: 1, bars: []string{". s s"}},
	}
}

// cave is slow and echoing, water drops over open fifths in D major
// pentatonic: 20 bars of 4/4 at 90 BPM, 53.3 s. The last four bars are
// only the pad and bass, so the drops can breathe.
func cave() Song {
	a := []string{
		"4 - - - . . 3 -",
		"2 - - - . . . .",
		"1 - - - . . 4 -",
		"3 - - - . . . .",
	}
	b := []string{
		"5 - 4 - 3 - . .",
		"2 - 3 - 0 - . .",
		"1 - - - 4 - - -",
		"3 - - - . . . .",
	}
	c := []string{
		"4 - - - 2 - - -",
		"3 - - - 5 - - -",
		"4 - 1 - - - . .",
		"0 - - - . . . .",
	}
	rest := []string{". . . . . . . ."}
	return Song{
		Name: "cave", BPM: 90, BeatsPerBar: 4, Bars: 20, Root: 62, Mode: Pentatonic,
		chords: []int{4, 0, 1, 3},
		parts: []part{
			{inst: pad, base: -12, step: 4, follow: true, bars: []string{"0+3"}},
			{inst: softBass, base: -24, step: 1, follow: true, bars: []string{"0 - . 3"}},
			{inst: waterDrop, base: 12, step: 0.5, bars: concat(a, b, a, c, rest, rest, rest, rest),
				echo: echo{beats: 0.75, gain: 0.4, repeats: 3}},
		},
	}
}

// boss is a bouncy, cheerful romp in G major, never scary: 24 bars of 4/4
// at 120 BPM, 48 s, with an octave-hopping bass and off-beat chords. The
// melody plays A, B, A.
func boss() Song {
	a := []string{
		"0 2 4 7 4 2 4 .",
		"3 5 7 5 3 . 2 .",
		"4 - 6 - 8 - 6 4",
		"7 - 4 - 7 - . .",
		"5 4 5 7 5 . 4 .",
		"3 2 3 5 3 . 2 .",
		"1 3 5 3 1 - . .",
		"4 - 6 - 4 - . .",
	}
	b := []string{
		"7 - 7 4 7 - 9 -",
		"8 7 5 3 5 - . .",
		"4 6 8 6 4 6 8 .",
		"7 - - - 4 - . .",
		"2 4 5 4 2 . 4 .",
		"3 5 7 5 3 . 5 .",
		"8 7 5 3 1 . 3 .",
		"4 - 6 - 8 - - .",
	}
	return Song{
		Name: "boss", BPM: 120, BeatsPerBar: 4, Bars: 24, Root: 67, Mode: Major,
		chords: []int{0, 3, 4, 0, 5, 3, 1, 4},
		parts: []part{
			{inst: bounceBass, base: -24, step: 0.5, follow: true, bars: []string{"0 7 4 7 0 7 4 7"}},
			{inst: pluck, base: -12, step: 0.5, follow: true, bars: []string{". 0+2+4 . 0+2+4 . 0+2+4 . 0+2+4"}},
			{inst: lead, base: 0, step: 0.5, bars: concat(a, b)},
		},
		drums: &drumPart{step: 0.5, bars: []string{"k . s . k . s s"}},
	}
}
