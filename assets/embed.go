// Package assets embeds the game's data files, so the browser build is a
// single file. Exported art, sounds and maps join the fonts here.
package assets

import "embed"

// FS holds the embedded files, such as fonts/Fredoka.ttf, the Tiled maps
// maps/*.tmj (village, house, meadow, dark_test, and testroom with one of each
// enemy), which internal/assets.MapNames lists, the item catalog
// data/items.json, the enemy definitions data/enemies.json, the string tables
// data/strings/sv.json and data/strings/en.json, critters.json, and the fox's
// character data data/rigs/fox.json with its placeholder part sheet
// sprites/critters/fox/fox.json, which internal/assets.LoadCharacter joins.
//
//go:embed fonts maps critters.json
//go:embed fonts maps data sprites
var FS embed.FS
