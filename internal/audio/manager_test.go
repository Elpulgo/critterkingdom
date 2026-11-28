package audio_test

import (
	"errors"
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/audio"
)

// fakeBackend stands in for the sound device (a process boundary): it records
// what the Manager asks it to play.
type fakeBackend struct {
	ready bool
	// failMusic names tracks PlayMusic refuses, as for a missing file.
	failMusic map[string]bool

	musicPlays  []play // every successful PlayMusic, in order
	music       string // the track looping now
	musicVolume float64
	sounds      []play
}

type play struct {
	name   string
	volume float64
}

func (b *fakeBackend) Ready() bool { return b.ready }

func (b *fakeBackend) PlayMusic(track string, volume float64) error {
	if b.failMusic[track] {
		return errors.New("no such file")
	}
	b.musicPlays = append(b.musicPlays, play{track, volume})
	b.music = track
	b.musicVolume = volume
	return nil
}

func (b *fakeBackend) StopMusic() { b.music = "" }

func (b *fakeBackend) SetMusicVolume(volume float64) { b.musicVolume = volume }

func (b *fakeBackend) PlaySound(name string, volume float64) error {
	b.sounds = append(b.sounds, play{name, volume})
	return nil
}

const frame = 1.0 / 60

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestVolumes_MasterIsCappedAndEveryChannelClampsToItsRange(t *testing.T) {
	tests := []struct {
		name     string
		set      func(m *audio.Manager, v float64)
		get      func(m *audio.Manager) float64
		in, want float64
	}{
		{"master above the cap stops at 0.8", (*audio.Manager).SetMasterVolume, (*audio.Manager).MasterVolume, 1, 0.8},
		{"master inside the range is kept", (*audio.Manager).SetMasterVolume, (*audio.Manager).MasterVolume, 0.3, 0.3},
		{"negative master is silence", (*audio.Manager).SetMasterVolume, (*audio.Manager).MasterVolume, -0.5, 0},
		{"music above 1 stops at 1", (*audio.Manager).SetMusicVolume, (*audio.Manager).MusicVolume, 1.7, 1},
		{"negative music is silence", (*audio.Manager).SetMusicVolume, (*audio.Manager).MusicVolume, -1, 0},
		{"sound above 1 stops at 1", (*audio.Manager).SetSoundVolume, (*audio.Manager).SoundVolume, 2, 1},
		{"negative sound is silence", (*audio.Manager).SetSoundVolume, (*audio.Manager).SoundVolume, -0.1, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := audio.NewManager(&fakeBackend{ready: true})
			tc.set(m, tc.in)
			if got := tc.get(m); got != tc.want {
				t.Errorf("volume = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMusic_AskedForBeforeUnlock_StartsOnceWhenBackendBecomesReady(t *testing.T) {
	b := &fakeBackend{}
	m := audio.NewManager(b)

	m.SetMusic("village")
	for range 10 {
		m.Tick(frame, false)
	}
	if len(b.musicPlays) != 0 {
		t.Fatalf("locked backend played %v, want nothing", b.musicPlays)
	}

	b.ready = true
	for range 10 {
		m.Tick(frame, false)
	}
	if len(b.musicPlays) != 1 || b.musicPlays[0].name != "village" {
		t.Fatalf("music plays = %v, want village once", b.musicPlays)
	}
	if v := b.musicPlays[0].volume; !near(v, 0.4) {
		t.Errorf("music volume = %v, want 0.4 (master 0.8 × music 0.5)", v)
	}
}

func TestMusic_VolumeChangesReachThePlayingTrack(t *testing.T) {
	b := &fakeBackend{ready: true}
	m := audio.NewManager(b)
	m.SetMusic("village")

	m.SetMasterVolume(0.4)
	if !near(b.musicVolume, 0.2) {
		t.Errorf("after master 0.4: music volume = %v, want 0.2", b.musicVolume)
	}
	m.SetMusicVolume(1)
	if !near(b.musicVolume, 0.4) {
		t.Errorf("after music 1: music volume = %v, want 0.4", b.musicVolume)
	}
}

func TestMusic_SameTrackKeepsPlayingAndAnotherTrackReplacesIt(t *testing.T) {
	b := &fakeBackend{ready: true}
	m := audio.NewManager(b)

	m.SetMusic("village")
	m.Tick(frame, false)
	m.SetMusic("village")
	m.Tick(frame, false)
	if len(b.musicPlays) != 1 {
		t.Fatalf("music plays = %v, want village started once", b.musicPlays)
	}

	m.SetMusic("meadow")
	m.Tick(frame, false)
	if len(b.musicPlays) != 2 || b.music != "meadow" {
		t.Errorf("music plays = %v, now %q; want meadow to replace village", b.musicPlays, b.music)
	}
}

func TestSound_BeforeUnlockIsDroppedAndAfterPlaysAtMasterTimesSound(t *testing.T) {
	b := &fakeBackend{}
	m := audio.NewManager(b)

	m.PlaySound("step")
	b.ready = true
	m.Tick(frame, false)
	if len(b.sounds) != 0 {
		t.Fatalf("sound asked for while locked was played later: %v", b.sounds)
	}

	m.PlaySound("step")
	if len(b.sounds) != 1 || b.sounds[0].name != "step" {
		t.Fatalf("sounds = %v, want step once", b.sounds)
	}
	if v := b.sounds[0].volume; !near(v, 0.64) {
		t.Errorf("sound volume = %v, want 0.64 (master 0.8 × sound 0.8)", v)
	}
}

func TestMusic_UnplayableOrEmptyTrackKeepsTheCurrentMusic(t *testing.T) {
	tests := []struct {
		name  string
		track string
	}{
		{"track the backend cannot play", "home"},
		{"map without a music track", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &fakeBackend{ready: true, failMusic: map[string]bool{"home": true}}
			m := audio.NewManager(b)
			m.SetMusic("village")

			m.SetMusic(tc.track)
			for range 3 {
				m.Tick(frame, false)
			}
			if b.music != "village" || len(b.musicPlays) != 1 {
				t.Errorf("music plays = %v, now %q; want village to keep playing", b.musicPlays, b.music)
			}
		})
	}
}

func TestMusic_EmptyTrackStartsNothingWhenSilent(t *testing.T) {
	b := &fakeBackend{ready: true}
	m := audio.NewManager(b)

	m.SetMusic("")
	m.Tick(frame, false)
	if len(b.musicPlays) != 0 {
		t.Errorf("music plays = %v, want nothing", b.musicPlays)
	}
}
