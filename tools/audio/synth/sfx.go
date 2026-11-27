package synth

// MIDI note numbers used by the sound effects.
const (
	noteC4 = 60
	noteC5 = 72
	noteE5 = 76
	noteG5 = 79
	noteA5 = 81
	noteB5 = 83
	noteC6 = 84
	noteD6 = 86
	noteE6 = 88
	noteG6 = 91
	noteA6 = 93
	noteC7 = 96
)

// sfxList is every sound effect, in catalog order. The names follow
// docs/PLAN.md §6.5; every sound is soft and short, with no harsh waves.
var sfxList = []struct {
	name   string
	render func() []float64
}{
	{"step", step},
	{"talk_blip", talkBlip},
	{"pickup", pickup},
	{"door", door},
	{"chest", chest},
	{"quest_start", questStart},
	{"quest_done", questDone},
	{"befriend", befriend},
	{"wrong_snack", wrongSnack},
	{"menu_tap", menuTap},
	{"hearts", hearts},
	{"sparkle", sparkle},
	{"zzz", zzz},
	{"swing", swing},
	{"hit_pop", hitPop},
	{"hurt", hurt},
	{"poof", poof},
	{"heart_pickup", heartPickup},
	{"room_lock", roomLock},
	{"room_unlock", roomUnlock},
	{"grub_grumble", grubGrumble},
}

// pling is a short plucked note: quick attack, ringing decay.
func pling(wave Wave, midi, dur, decay float64) []float64 {
	return Tone{Wave: wave, Freq: Hz(midi), Dur: dur, Amp: 1, Attack: 0.005, Decay: decay, Release: 0.03}.Render()
}

// arpeggio plays the notes one after another, each gap seconds after the last.
func arpeggio(wave Wave, notes []float64, gap, dur, decay float64) []float64 {
	var out []float64
	for i, n := range notes {
		out = Mix(out, pling(wave, n, dur, decay), float64(i)*gap)
	}
	return out
}

// filteredNoise is noise through a low-pass whose cutoff follows cutoff,
// shaped by attack, decay and release.
func filteredNoise(seed uint32, dur float64, cutoff func(float64) float64, attack, decay, release float64) []float64 {
	return Shape(LowPass(Noise(seed, Samples(dur)), cutoff), attack, decay, release)
}

// step is a soft footstep: a muffled pad of noise over a low thump.
func step() []float64 {
	out := filteredNoise(11, 0.09, Fixed(900), 0.004, 0.025, 0.02)
	thump := Tone{Freq: 140, EndFreq: 90, Dur: 0.09, Amp: 0.4, Attack: 0.003, Decay: 0.03, Release: 0.02}.Render()
	return Finish(Mix(out, thump, 0))
}

// talkBlip is a tiny triangle blip for text appearing.
func talkBlip() []float64 {
	return Finish(Tone{Wave: Triangle, Freq: Hz(noteE5), Dur: 0.06, Amp: 1, Attack: 0.005, Decay: 0.04, Release: 0.02}.Render())
}

// pickup is two quick notes up: something nice went into the backpack.
func pickup() []float64 {
	return Finish(arpeggio(Triangle, []float64{noteC6, noteG6}, 0.07, 0.18, 0.08))
}

// door is a low wooden knock with a breathy swing of air.
func door() []float64 {
	air := filteredNoise(23, 0.45, func(s float64) float64 { return 300 + 900*s }, 0.12, 0, 0.2)
	for i := range air {
		air[i] *= 0.5
	}
	thud := Tone{Freq: 110, EndFreq: 75, Dur: 0.3, Amp: 0.8, Attack: 0.004, Decay: 0.08, Release: 0.05}.Render()
	return Finish(Mix(air, thud, 0.3))
}

// chest is a soft lid thump followed by a rising sparkle of four notes.
func chest() []float64 {
	thud := Tone{Freq: 130, EndFreq: 90, Dur: 0.2, Amp: 0.7, Attack: 0.004, Decay: 0.06, Release: 0.05}.Render()
	notes := arpeggio(Bell, []float64{noteC5, noteE5, noteG5, noteC6}, 0.08, 0.5, 0.18)
	return Finish(Mix(thud, notes, 0.12))
}

// questStart is three notes up: something new to do.
func questStart() []float64 {
	return Finish(arpeggio(Triangle, []float64{noteC5, noteE5, noteG5}, 0.14, 0.35, 0.15))
}

// questDone is a little fanfare: three notes up, then a held major chord.
func questDone() []float64 {
	out := arpeggio(Triangle, []float64{noteC5, noteE5, noteG5}, 0.12, 0.25, 0.12)
	for _, n := range []float64{noteC5, noteE5, noteG5, noteC6} {
		chord := Tone{Wave: Triangle, Freq: Hz(n), Dur: 1.0, Amp: 0.5, Attack: 0.02, Decay: 0.5, Release: 0.35, VibratoRate: 5, VibratoDepth: 0.08}.Render()
		out = Mix(out, chord, 0.38)
	}
	return Finish(out)
}

// befriend is a bright, happy sparkle running up the pentatonic scale.
func befriend() []float64 {
	return Finish(arpeggio(Bell, []float64{noteC6, noteD6, noteE6, noteG6, noteA6, noteC7}, 0.06, 0.55, 0.15))
}

// wrongSnack is a soft, round "bwomp" sliding down, never harsh.
func wrongSnack() []float64 {
	t := Tone{Wave: RoundSquare, Freq: 300, EndFreq: 150, Dur: 0.45, Amp: 1, Attack: 0.02, Decay: 0.25, Release: 0.12}.Render()
	return Finish(LowPass(t, Fixed(700)))
}

// menuTap is a very short, quiet tick of a sine.
func menuTap() []float64 {
	return Finish(Tone{Freq: Hz(noteA5), Dur: 0.05, Amp: 1, Attack: 0.003, Decay: 0.02, Release: 0.02}.Render())
}

// hearts is two warm notes with a gentle wobble, a little "aww".
func hearts() []float64 {
	var out []float64
	for i, n := range []float64{noteE5, noteA5} {
		note := Tone{Wave: Triangle, Freq: Hz(n), Dur: 0.4, Amp: 1, Attack: 0.02, Decay: 0.25, Release: 0.12, VibratoRate: 6, VibratoDepth: 0.15}.Render()
		out = Mix(out, note, float64(i)*0.18)
	}
	return Finish(out)
}

// sparkle is a scatter of high bell pings at fixed, irregular times.
func sparkle() []float64 {
	pings := []struct{ at, midi float64 }{
		{0, noteE6}, {0.07, noteA6}, {0.12, noteG6}, {0.2, noteC7}, {0.27, noteE6}, {0.33, noteA6},
	}
	var out []float64
	for _, p := range pings {
		out = Mix(out, Tone{Wave: Bell, Freq: Hz(p.midi), Dur: 0.3, Amp: 0.7, Attack: 0.003, Decay: 0.08, Release: 0.05}.Render(), p.at)
	}
	return Finish(out)
}

// zzz is three slow, sleepy hums, each a little lower than the last.
func zzz() []float64 {
	var out []float64
	for i, f := range []float64{220, 196, 175} {
		hum := Tone{Wave: RoundSquare, Freq: f, EndFreq: f * 0.94, Dur: 0.4, Amp: 1, Attack: 0.12, Release: 0.2, VibratoRate: 7, VibratoDepth: 0.2}.Render()
		out = Mix(out, LowPass(hum, Fixed(500)), float64(i)*0.45)
	}
	return Finish(out)
}

// swing is a short whoosh: noise whose muffling opens and closes again.
func swing() []float64 {
	const dur = 0.28
	cutoff := func(s float64) float64 {
		x := s / dur
		return 300 + 2200*4*x*(1-x)
	}
	return Finish(filteredNoise(37, dur, cutoff, 0.1, 0, 0.15))
}

// hitPop is a bubbly pop: a quick downward blip with a tiny soft tap.
func hitPop() []float64 {
	pop := Tone{Freq: 650, EndFreq: 220, Dur: 0.12, Amp: 1, Attack: 0.003, Decay: 0.04, Release: 0.03}.Render()
	tap := filteredNoise(41, 0.04, Fixed(1500), 0.002, 0.01, 0.02)
	for i := range tap {
		tap[i] *= 0.4
	}
	return Finish(Mix(pop, tap, 0))
}

// hurt is a short, round "oof" sliding down, more surprised than pained.
func hurt() []float64 {
	t := Tone{Wave: RoundSquare, Freq: 280, EndFreq: 190, Dur: 0.26, Amp: 1, Attack: 0.015, Decay: 0.15, Release: 0.08}.Render()
	return Finish(LowPass(t, Fixed(900)))
}

// poof is a soft puff of smoke: noise that dims as it fades.
func poof() []float64 {
	return Finish(filteredNoise(53, 0.45, func(s float64) float64 { return 1800 - 3000*s }, 0.01, 0.12, 0.15))
}

// heartPickup is three bright notes up with a ringing top note.
func heartPickup() []float64 {
	return Finish(arpeggio(Bell, []float64{noteG5, noteB5, noteD6}, 0.07, 0.45, 0.16))
}

// roomLock is a chime falling down a major triad: the room closes.
func roomLock() []float64 {
	return Finish(arpeggio(Bell, []float64{noteG5, noteE5, noteC5}, 0.16, 0.7, 0.25))
}

// roomUnlock is a chime rising up a major triad to the octave: the room opens.
func roomUnlock() []float64 {
	return Finish(arpeggio(Bell, []float64{noteC5, noteE5, noteG5, noteC6}, 0.14, 0.7, 0.25))
}

// grubGrumble is the Grub's friendly low grumble, rolling and rising at the
// end like a question.
func grubGrumble() []float64 {
	var out []float64
	parts := []struct{ at, from, to, dur float64 }{
		{0, Hz(noteC4 - 24), Hz(noteC4 - 22), 0.35},
		{0.3, Hz(noteC4 - 22), Hz(noteC4 - 25), 0.3},
		{0.55, Hz(noteC4 - 24), Hz(noteC4 - 17), 0.35},
	}
	for _, p := range parts {
		g := Tone{Wave: RoundSquare, Freq: p.from, EndFreq: p.to, Dur: p.dur, Amp: 1, Attack: 0.05, Release: 0.12, VibratoRate: 11, VibratoDepth: 0.5}.Render()
		out = Mix(out, LowPass(g, Fixed(450)), p.at)
	}
	return Finish(out)
}
