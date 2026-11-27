package save

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound reports that no save exists under the requested key.
	ErrNotFound = errors.New("save: no save under this key")
	// ErrInvalidKey reports a key that is empty or holds anything other than
	// a-z, 0-9, '-' or '_'.
	ErrInvalidKey = errors.New("save: invalid save key")
)

// Store saves and loads raw save bytes by key. It knows nothing about the
// save format; Write and Read join it to Encode and Decode.
type Store interface {
	// Load returns the bytes saved under key, or an error wrapping
	// ErrNotFound when nothing is saved there. Load never writes.
	Load(key string) ([]byte, error)
	// Save stores data under key, replacing any earlier save.
	Save(key string, data []byte) error
}

// Write encodes sg at CurrentVersion and saves it under key in s.
func Write(s Store, key string, sg SaveGame) error {
	data, err := Encode(sg)
	if err != nil {
		return fmt.Errorf("encode save %q: %w", key, err)
	}
	if err := s.Save(key, data); err != nil {
		return fmt.Errorf("write save %q: %w", key, err)
	}
	return nil
}

// Read loads the save under key from s and decodes it, migrating an older
// version. It returns an error wrapping ErrNotFound when no save exists,
// ErrCorrupt when the stored data is not a valid save, or ErrNewerVersion
// when it was written by a newer game. Read never changes the stored data.
func Read(s Store, key string) (SaveGame, error) {
	data, err := s.Load(key)
	if err != nil {
		return SaveGame{}, fmt.Errorf("read save %q: %w", key, err)
	}
	sg, err := Decode(data)
	if err != nil {
		return SaveGame{}, fmt.Errorf("read save %q: %w", key, err)
	}
	return sg, nil
}

// validKey reports whether key is non-empty and holds only a-z, 0-9, '-' and
// '_', so it is safe to use as a file name or a storage key suffix.
func validKey(key string) bool {
	if key == "" {
		return false
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return false
		}
	}
	return true
}
