// Package assets embeds the game's data files, so the browser build is a
// single file. Exported art and sounds join the fonts and maps here.
package assets

import "embed"

// FS holds the embedded files, such as fonts/Fredoka.ttf, the Tiled maps
// maps/*.tmj (village, house, meadow, dark_test, and testroom with one of each
// enemy), which internal/assets.MapNames lists, the item catalog
// data/items.json, the enemy definitions data/enemies.json, the string tables
// data/strings/sv.json and data/strings/en.json, critters.json, the music
// loops audio/music/*.ogg (internal/assets.LoadMusic) and the sound effects
// audio/sfx/*.ogg (internal/assets.LoadSound). The critter voices in
// audio/voices are not embedded.
//
//go:embed fonts maps data critters.json audio/music audio/sfx
var FS embed.FS
