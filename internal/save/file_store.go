package save

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FileStore keeps each save as <Dir>/<key>.json on the local file system.
type FileStore struct {
	// Dir is the directory the save files live in; Save creates it if needed.
	Dir string
}

// Load reads the save file for key. It returns an error wrapping ErrInvalidKey
// for a bad key and ErrNotFound when the file does not exist.
func (s FileStore) Load(key string) ([]byte, error) {
	if !validKey(key) {
		return nil, fmt.Errorf("load %q: %w", key, ErrInvalidKey)
	}
	data, err := os.ReadFile(s.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("load %q: %w", key, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("load %q: %w", key, err)
	}
	return data, nil
}

// Save writes data to the save file for key, creating Dir if missing. It
// writes a temporary file in Dir and renames it over the target, so a crash
// mid-write never leaves a half-written save. It returns an error wrapping
// ErrInvalidKey for a bad key.
func (s FileStore) Save(key string, data []byte) (err error) {
	if !validKey(key) {
		return fmt.Errorf("save %q: %w", key, ErrInvalidKey)
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return fmt.Errorf("save %q: create directory: %w", key, err)
	}
	tmp, err := os.CreateTemp(s.Dir, key+".*.tmp")
	if err != nil {
		return fmt.Errorf("save %q: create temp file: %w", key, err)
	}
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("save %q: write temp file: %w", key, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("save %q: close temp file: %w", key, err)
	}
	if err := os.Rename(tmp.Name(), s.path(key)); err != nil {
		return fmt.Errorf("save %q: replace save file: %w", key, err)
	}
	return nil
}

func (s FileStore) path(key string) string {
	return filepath.Join(s.Dir, key+".json")
}
