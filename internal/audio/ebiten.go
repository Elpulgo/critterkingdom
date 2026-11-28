package audio

import (
	"bytes"
	"fmt"
	"io"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

// SampleRate is the sample rate of the Ebitengine audio context, in Hz. OGG
// files at another rate are resampled when they are decoded.
const SampleRate = 44100

// EbitenBackend is the Backend that plays through Ebitengine's audio device.
// It decodes the OGG Vorbis data its loaders return: music streams as an
// endless loop, sound effects are decoded once and kept as PCM.
//
// It is not safe for concurrent use; the game calls it from its update loop.
type EbitenBackend struct {
	ctx                  *eaudio.Context
	loadMusic, loadSound func(string) ([]byte, error)

	// music is the player looping the current track, nil when none plays.
	music *eaudio.Player
	// sounds caches each sound effect's decoded PCM by name.
	sounds map[string][]byte
}

// NewEbitenBackend opens Ebitengine's audio context at SampleRate and returns
// a backend that reads music tracks with loadMusic and sound effects with
// loadSound; both return OGG Vorbis data for a name. Ebitengine allows one
// audio context per process, so call it once.
func NewEbitenBackend(loadMusic, loadSound func(string) ([]byte, error)) *EbitenBackend {
	return &EbitenBackend{
		ctx:       eaudio.NewContext(SampleRate),
		loadMusic: loadMusic,
		loadSound: loadSound,
		sounds:    map[string][]byte{},
	}
}

// Ready reports whether the audio device may play. In the browser it turns
// true once the first user gesture has resumed the audio context.
func (b *EbitenBackend) Ready() bool { return b.ctx.IsReady() }

// PlayMusic loads, decodes and loops track at volume, stopping the player of
// the previous track before the new one starts. A load or decode error is
// returned before anything is stopped, so the previous track keeps playing.
func (b *EbitenBackend) PlayMusic(track string, volume float64) error {
	data, err := b.loadMusic(track)
	if err != nil {
		return err
	}
	stream, err := vorbis.DecodeWithSampleRate(SampleRate, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decode music %q: %w", track, err)
	}
	player, err := b.ctx.NewPlayer(eaudio.NewInfiniteLoop(stream, stream.Length()))
	if err != nil {
		return fmt.Errorf("play music %q: %w", track, err)
	}
	player.SetVolume(volume)
	b.StopMusic()
	b.music = player
	player.Play()
	return nil
}

// StopMusic stops the current track, if any, and lets its player go.
// PauseAndStopReading is how Ebitengine 2.10 closes a player (Close is
// deprecated); once the player is unreachable Ebitengine releases it.
func (b *EbitenBackend) StopMusic() {
	if b.music == nil {
		return
	}
	b.music.PauseAndStopReading()
	b.music = nil
}

// SetMusicVolume changes the volume of the playing track.
func (b *EbitenBackend) SetMusicVolume(volume float64) {
	if b.music != nil {
		b.music.SetVolume(volume)
	}
}

// PlaySound plays the named sound effect once at volume. The first call for a
// name loads and decodes it; later calls reuse the decoded PCM. Several
// plays of the same sound may overlap.
func (b *EbitenBackend) PlaySound(name string, volume float64) error {
	pcm, err := b.soundPCM(name)
	if err != nil {
		return err
	}
	player := b.ctx.NewPlayerFromBytes(pcm)
	player.SetVolume(volume)
	player.Play()
	return nil
}

// soundPCM returns the decoded PCM of a sound effect, from the cache when it
// was decoded before. Failures are not cached.
func (b *EbitenBackend) soundPCM(name string) ([]byte, error) {
	if pcm, ok := b.sounds[name]; ok {
		return pcm, nil
	}
	data, err := b.loadSound(name)
	if err != nil {
		return nil, err
	}
	stream, err := vorbis.DecodeWithSampleRate(SampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode sound %q: %w", name, err)
	}
	pcm, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("decode sound %q: %w", name, err)
	}
	b.sounds[name] = pcm
	return pcm, nil
}
