package assets

import (
	"fmt"
	"io/fs"

	embedded "github.com/Elpulgo/critterkingdom/assets"
)

// LoadMusic returns the OGG Vorbis data of the embedded music loop at
// audio/music/<track>.ogg.
func LoadMusic(track string) ([]byte, error) {
	b, err := fs.ReadFile(embedded.FS, "audio/music/"+track+".ogg")
	if err != nil {
		return nil, fmt.Errorf("read music %q: %w", track, err)
	}
	return b, nil
}

// LoadSound returns the OGG Vorbis data of the embedded sound effect at
// audio/sfx/<name>.ogg.
func LoadSound(name string) ([]byte, error) {
	b, err := fs.ReadFile(embedded.FS, "audio/sfx/"+name+".ogg")
	if err != nil {
		return nil, fmt.Errorf("read sound %q: %w", name, err)
	}
	return b, nil
}
