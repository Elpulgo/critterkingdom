# Critter Kingdom

A cosy top-down adventure for children aged 5 to 9, written in Go with
[Ebitengine](https://ebitengine.org/). It runs in a browser on a tablet and as
a desktop game. The plan, the roadmap and the reasons behind them are in
[docs/PLAN.md](docs/PLAN.md).

Status: milestone M1, "A place to walk", in progress. A placeholder critter
starts in a small village, walks into a house and out to the meadow, hops and
says hello.

## Getting started

You need Go 1.25 or later and nothing else: no C compiler and no Node.
The exceptions are `make art`, which also needs Inkscape (see [Art](#art)),
and `make audio`, which also needs ffmpeg with libvorbis.

```sh
make run              # play on the desktop
make run TAGS=debug   # with the debug overlay, F3 hides it
make run ARGS=-pad    # with the touch pad instead of tap-to-walk
make run ARGS=-lang=en # in English instead of Swedish
make run ARGS='-map meadow'  # start on another map from assets/maps
make serve            # play in a browser at http://localhost:8080/web/
make test             # run the tests
make check            # what CI runs: gofmt, go mod tidy, vet and tests
make web              # build the browser bundle into build/web
make art              # export changed drawings to PNG (needs Inkscape)
make audio            # render every sound to OGG in assets/audio (needs ffmpeg)
make                  # list every command
```

Without make, use `go run ./cmd/game` and `go tool wasmserve ./cmd/game`.

## Playing on a tablet during development

1. Run `make serve` on your computer.
2. Find the computer's address on the local network. On a Mac,
   `ipconfig getifaddr en0` prints it.
3. On a tablet on the same Wi-Fi, open `http://<address>:8080/web/`. Keep the
   trailing slash, because `/web/index.html` gets wasmserve's plain page
   instead of the game's.
4. Add `?pad=1` to the address to steer with a pad on the left half of the
   screen instead of tapping where to walk.
5. Add `?lang=en` to play in English instead of Swedish, or both at once:
   `?pad=1&lang=en`.

Every reload rebuilds the game, so edit, save and reload the tablet.

## Controls

| Input | Move | Paw, say hello |
|---|---|---|
| Touch | Tap where to walk, or hold and lead the critter. With `?pad=1`, drag on the left half. | The paw button at the bottom right, shown after the first touch |
| Keyboard | Arrows or WASD | Space, Enter or E |
| Gamepad | Left stick or d-pad | A, the bottom face button |
| Mouse | Click where to walk, or hold and lead | |

Esc, Tab, Start and Y are the menu buttons. Nothing opens yet. In a debug
build, F3 shows and hides the overlay.

## Art

The placeholder art is drawn in code with smooth shapes at full screen
resolution. Real art is drawn in [Inkscape](https://inkscape.org/) and
exported to PNG at 8 px per world pixel by `make art`;
[docs/ART-BIBLE.md](docs/ART-BIBLE.md) sets the palette, sizes and shape rules,
and section 6 of [docs/PLAN.md](docs/PLAN.md) has the asset pipeline plan.

`make art` runs Inkscape 1.x's command line on every drawing that changed
since its last export:

- Characters go in `assets-src/sprites/critters/`, `npcs/` or `enemies/`, one
  file per character, for example `assets-src/sprites/critters/fox.svg`. Each
  part is an object with an id named `<view>_<part>`, where the view is
  `down`, `up` or `side` (`down_head`, `down_ear_left`). A tiny circle named
  `<view>_<part>_pivot` marks where a part turns, and one named `<view>_feet`
  marks the feet. Every part becomes its own PNG in
  `assets/sprites/critters/fox/`, next to `fox.json`, which says where each
  part sits and turns. Only give whole parts a `<view>_` id: a named object
  inside a part, such as `down_head_shine`, is exported as a part of its own.
- Every other drawing under `assets-src/sprites/` (props, effects, ui,
  emotes) and `assets-src/tiles/` is exported whole, as one PNG of its page at
  the same path under `assets/`.

A drawing is exported again when its SVG is newer than its JSON or PNG;
`make art ART_ARGS=-force` exports everything. If Inkscape is not on your
`PATH`, give its path, for example on a Mac
`make art INKSCAPE=/Applications/Inkscape.app/Contents/MacOS/inkscape`.

## Licence

The code is under the MIT licence in [LICENSE](LICENSE). Fonts and other
assets keep their own licences, listed in [docs/CREDITS.md](docs/CREDITS.md).
