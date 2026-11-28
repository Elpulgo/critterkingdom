// Command game runs Critter Kingdom, on the desktop or, built for js/wasm, in
// the browser.
package main

import (
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Elpulgo/critterkingdom/internal/assets"
	"github.com/Elpulgo/critterkingdom/internal/combat"
	"github.com/Elpulgo/critterkingdom/internal/engine"
	"github.com/Elpulgo/critterkingdom/internal/i18n"
	"github.com/Elpulgo/critterkingdom/internal/world"
)

// options are the settings given at start: flags on the desktop, the page's
// query string in the browser.
type options struct {
	// touchPad turns on the floating touch pad.
	touchPad bool
	// startMap names the map the game starts on, such as "village".
	startMap string
	// lang is the language to start in, "sv" or "en". An empty or unknown
	// value keeps Swedish.
	lang string
	// preset names the age preset, "explorer" or "little-explorer". An empty
	// or unknown value keeps Explorer.
	preset string
	// peaceful puts every enemy to sleep.
	peaceful bool
}

// defaultStartMap is the map the game starts on unless -map names another.
const defaultStartMap = "village"

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
	species, err := assets.LoadCritters()
	if err != nil {
		log.Fatal(err)
	}
	enemies, err := assets.LoadEnemies()
	if err != nil {
		log.Fatal(err)
	}
	preset := combat.PresetExplorer
	if opts.preset != "" {
		if preset, err = combat.ParsePreset(opts.preset); err != nil {
			log.Printf("preset: %v; starting on explorer", err)
		}
	}
	sceneOpts := []world.Option{world.WithEnemies(enemies), world.WithPreset(preset)}
	if opts.peaceful {
		sceneOpts = append(sceneOpts, world.WithPeaceful())
	}
	scene, err := world.NewScene(font, assets.LoadMap, opts.startMap, sceneOpts...)
	if err != nil {
		log.Fatal(err)
	}
	// One generator, seeded once from the runtime's random source, drives
	// every wild critter for the whole run.
	scene.AddWildCritters(species, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())))
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
