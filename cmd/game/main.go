// Command game runs Critter Kingdom, on the desktop or, built for js/wasm, in
// the browser.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/i18n"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// options are the settings given at start: flags on the desktop, the page's
// query string in the browser.
type options struct {
	// touchPad turns on the floating touch pad.
	touchPad bool
	// lang is the language to start in, "sv" or "en". An empty or unknown
	// value keeps Swedish.
	lang string
}

func main() {
	opts := parseOptions()
	font, err := assets.LoadFont()
	if err != nil {
		log.Fatal(err)
	}
	strs, err := assets.LoadStrings()
	if err != nil {
		log.Fatal(err)
	}
	i18n.Use(strs)
	if opts.lang != "" {
		if err := i18n.SetLanguage(i18n.Lang(opts.lang)); err != nil {
			log.Printf("lang %q: %v; starting in %s", opts.lang, err, i18n.Language())
		}
	}
	m, err := assets.LoadMap("meadow")
	if err != nil {
		log.Fatal(err)
	}
	scene, err := world.NewScene(font, m)
	if err != nil {
		log.Fatal(err)
	}
	game := engine.NewGame(scene, engine.Options{
		Font:     font,
		TouchPad: opts.touchPad,
	})

	ebiten.SetWindowTitle("Critter Kingdom")
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
