// Package assets embeds the game's data files, so the browser build is a
// single file. Exported art, sounds and maps join the fonts here.
package assets

import "embed"

// FS holds the embedded files, such as fonts/Fredoka.ttf, maps/meadow.tmj,
// the item catalog data/items.json, the string tables data/strings/sv.json
// and data/strings/en.json, the quest definitions data/quests/*.json, and
// critters.json.
//
//go:embed fonts maps critters.json
//go:embed fonts maps data
var FS embed.FS
