// Package assets embeds the game's data files, so the browser build is a
// single file. Exported art, sounds and maps join the fonts here.
package assets

import "embed"

// FS holds the embedded files: fonts/Fredoka.ttf, and the Tiled maps
// maps/*.tmj (village, house, meadow), which internal/assets.MapNames lists.
//
//go:embed fonts maps
var FS embed.FS
