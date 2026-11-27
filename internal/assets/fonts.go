// Package assets loads the embedded game data.
package assets

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	embedded "github.com/Elpulgo/critterkingdom/assets"
)

// LoadFont loads Fredoka, the game's font. It is a variable font, so one file
// gives every weight; see engine.NewFace.
func LoadFont() (*text.GoTextFaceSource, error) {
	b, err := fs.ReadFile(embedded.FS, "fonts/Fredoka.ttf")
	if err != nil {
		return nil, fmt.Errorf("read font: %w", err)
	}
	src, err := text.NewGoTextFaceSource(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	return src, nil
}
