package synth

// Kind says what a sound is for: it decides its length limits and how it is
// encoded.
type Kind int

const (
	// SFX is a short sound effect.
	SFX Kind = iota
	// Music is a looping music track.
	Music
	// Voice is one critter-voice syllable.
	Voice
)

// String returns the kind's name: "sfx", "music" or "voice".
func (k Kind) String() string {
	switch k {
	case SFX:
		return "sfx"
	case Music:
		return "music"
	case Voice:
		return "voice"
	}
	return "unknown"
}

// Sound is one entry in the catalog.
type Sound struct {
	// Path is the OGG file's slash-separated path under assets/audio/,
	// for example "sfx/pickup.ogg".
	Path string
	// Kind is what the sound is for.
	Kind Kind
	// Render synthesises the sound: mono samples in -1..1 at SampleRate.
	Render func() []float64
}

// Catalog returns every sound the game ships, in a fixed order.
func Catalog() []Sound {
	var c []Sound
	for _, e := range sfxList {
		c = append(c, Sound{Path: "sfx/" + e.name + ".ogg", Kind: SFX, Render: e.render})
	}
	for _, s := range Songs() {
		c = append(c, Sound{Path: "music/" + s.Name + ".ogg", Kind: Music, Render: s.Render})
	}
	for _, v := range Speakers() {
		for _, syl := range Syllables {
			render := func() []float64 { return v.Render(syl) }
			c = append(c, Sound{Path: "voices/" + v.Name + "/" + syl + ".ogg", Kind: Voice, Render: render})
		}
	}
	return c
}
