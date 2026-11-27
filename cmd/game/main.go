// Command game runs Critter Kingdom, on the desktop or, built for js/wasm, in
// the browser.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// options are the settings given at start: flags on the desktop, the page's
// query string in the browser.
type options struct {
	// touchPad turns on the floating touch pad.
	touchPad bool
}

func main() {
	opts := parseOptions()
	font, err := assets.LoadFont()
	if err != nil {
		log.Fatal(err)
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
