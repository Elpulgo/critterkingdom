// Package audio plays the game's music and sound effects. The Manager holds
// the rules (channel volumes under a capped master volume, music held back
// until the device is unlocked, no restart of the track already playing) and
// talks to the sound device through the small Backend interface, so it has no
// dependency on Ebitengine and is tested with an in-memory backend.
package audio

import "log"

// MaxMaster caps the master volume: everything in the game stays soft.
const MaxMaster = 0.8

// Default volumes of a new Manager.
const (
	DefaultMaster = MaxMaster
	DefaultMusic  = 0.5
	DefaultSound  = 0.8
)

// Backend is the sound device the Manager drives. Volumes are final, in
// [0, 1]: the Manager has already applied master × channel.
type Backend interface {
	// Ready reports whether the device may play. In the browser it is false
	// until the first user gesture unlocks audio.
	Ready() bool
	// PlayMusic starts track as a loop, replacing the current one. On error
	// the current track keeps playing.
	PlayMusic(track string, volume float64) error
	// StopMusic stops the current track, if any.
	StopMusic()
	// SetMusicVolume changes the volume of the playing track.
	SetMusicVolume(volume float64)
	// PlaySound plays the named sound effect once.
	PlaySound(name string, volume float64) error
}

// Manager decides what the Backend plays and how loud. Its zero value is not
// usable; build one with NewManager.
type Manager struct {
	backend Backend

	master, music, sound float64

	// wanted is the track asked for with SetMusic; playing is the track the
	// backend is looping. They differ while the device is locked.
	wanted, playing string

	// failed holds names that already failed, so each is logged once.
	failed map[string]bool
}

// NewManager returns a Manager over b with the default volumes and no music.
func NewManager(b Backend) *Manager {
	return &Manager{
		backend: b,
		master:  DefaultMaster,
		music:   DefaultMusic,
		sound:   DefaultSound,
		failed:  map[string]bool{},
	}
}

// MasterVolume returns the master volume, in [0, MaxMaster].
func (m *Manager) MasterVolume() float64 { return m.master }

// MusicVolume returns the music channel volume, in [0, 1].
func (m *Manager) MusicVolume() float64 { return m.music }

// SoundVolume returns the sound effect channel volume, in [0, 1].
func (m *Manager) SoundVolume() float64 { return m.sound }

// SetMasterVolume sets the master volume, clamped to [0, MaxMaster].
func (m *Manager) SetMasterVolume(v float64) {
	m.master = clamp(v, MaxMaster)
	m.updateMusicVolume()
}

// SetMusicVolume sets the music channel volume, clamped to [0, 1].
func (m *Manager) SetMusicVolume(v float64) {
	m.music = clamp(v, 1)
	m.updateMusicVolume()
}

// SetSoundVolume sets the sound effect channel volume, clamped to [0, 1].
func (m *Manager) SetSoundVolume(v float64) {
	m.sound = clamp(v, 1)
}

// SetMusic asks for track to loop. The track already playing carries on
// without a restart, and an empty name changes nothing. While the backend is
// not ready the track is remembered and starts on the first Tick after it
// becomes ready.
func (m *Manager) SetMusic(track string) {
	if track == "" {
		return
	}
	m.wanted = track
	m.startWantedMusic()
}

// PlaySound plays the named sound effect at master × sound volume. While the
// backend is not ready the sound is dropped, not queued.
func (m *Manager) PlaySound(name string) {
	if !m.backend.Ready() {
		return
	}
	if err := m.backend.PlaySound(name, m.master*m.sound); err != nil {
		m.logOnce("sound", name, err)
	}
}

// Tick advances the manager by one frame of dt seconds; walking reports
// whether the player moved this frame. It starts the music asked for while
// the backend was locked as soon as the backend is ready.
func (m *Manager) Tick(dt float64, walking bool) {
	m.startWantedMusic()
}

// startWantedMusic starts the wanted track when it is not the one playing and
// the backend is ready. A track that fails is given up on, so the previous
// one keeps playing and the failure is not retried every tick.
func (m *Manager) startWantedMusic() {
	if m.wanted == m.playing || !m.backend.Ready() {
		return
	}
	if err := m.backend.PlayMusic(m.wanted, m.master*m.music); err != nil {
		m.logOnce("music", m.wanted, err)
		m.wanted = m.playing
		return
	}
	m.playing = m.wanted
}

func (m *Manager) updateMusicVolume() {
	if m.playing != "" {
		m.backend.SetMusicVolume(m.master * m.music)
	}
}

func (m *Manager) logOnce(kind, name string, err error) {
	key := kind + "/" + name
	if m.failed[key] {
		return
	}
	m.failed[key] = true
	log.Printf("audio: %s %q: %v", kind, name, err)
}

// clamp limits v to [0, hi]. NaN becomes 0.
func clamp(v, hi float64) float64 {
	switch {
	case v > hi:
		return hi
	case v >= 0:
		return v
	default:
		return 0
	}
}
