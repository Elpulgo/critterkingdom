// Package assets embeds the game's data files, so the browser build is a
// single file. Exported art, sounds and maps join the fonts here.
package assets

import "embed"

// FS holds the embedded files, such as fonts/Fredoka.ttf, maps/meadow.tmj
// and critters.json.
//
//go:embed fonts maps critters.json
var FS embed.FS
