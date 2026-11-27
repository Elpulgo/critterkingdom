// Package assets embeds the game's data files, so the browser build is a
// single file. Exported art, sounds and maps join the fonts here.
package assets

import "embed"

// FS holds the embedded files, such as fonts/Fredoka.ttf, maps/meadow.tmj
// and data/items.json.
//
//go:embed fonts maps data
var FS embed.FS
