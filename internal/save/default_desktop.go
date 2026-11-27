//go:build !js

package save

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultStore returns the store the game saves to: a FileStore in
// critterkingdom/saves under the user's config directory. It returns an error
// when the config directory cannot be found.
func DefaultStore() (Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("default save store: %w", err)
	}
	return FileStore{Dir: filepath.Join(dir, "critterkingdom", "saves")}, nil
}
