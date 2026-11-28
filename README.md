# Critter Kingdom

A cosy top-down adventure for children aged 5 to 9, written in Go with
[Ebitengine](https://ebitengine.org/). It runs in a browser on a tablet and as
a desktop game. The plan, the roadmap and the reasons behind them are in
[docs/PLAN.md](docs/PLAN.md).

Status: milestone M1, "A place to walk", in progress. A placeholder critter,
built from parts animated in code, starts in a small village, walks into a
house and out to the meadow, hops and says hello.

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
make run ARGS='-map dark_test'  # start on the dark test map, lit by the lantern
make run ARGS='-map testroom'  # start in the test room with one of each enemy
make run ARGS='-preset little-explorer'  # gentler, fewer enemies (default explorer)
make run ARGS=-peaceful  # every enemy sleeps and never attacks
make serve            # play in a browser at http://localhost:8080/web/
make test             # run the tests
make check            # what CI runs: gofmt, go mod tidy, vet and tests
make web              # build the browser bundle into build/web
make art              # export changed drawings to PNG (needs Inkscape)
make audio            # render every sound to OGG in assets/audio (needs ffmpeg)
make                  # list every command
```

Without make, use `go run ./cmd/game` and `go tool wasmserve ./cmd/game`.
In the browser, add `?preset=little-explorer` or `?peaceful=1` to the
address for the same as `-preset little-explorer` and `-peaceful`.

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
6. Add `?preset=little-explorer` for the gentler Little Explorer enemies,
   and `?peaceful=1` to put every enemy to sleep, for example
   `?preset=little-explorer&peaceful=1`. An unknown preset name keeps
   Explorer.

Every reload rebuilds the game, so edit, save and reload the tablet.

## Controls

| Input | Move | Paw, say hello |
|---|---|---|
| Touch | Tap where to walk, or hold and lead the critter. It walks around anything in the way, and if you tap something solid or out of reach it walks as near as it can get. With `?pad=1`, drag on the left half. | The paw button at the bottom right, shown after the first touch |
| Keyboard | Arrows or WASD | Space, Enter or E |
| Gamepad | Left stick or d-pad | A, the bottom face button |
| Mouse | Click where to walk, or hold and lead, the same as touch | |

Esc, Tab, Start and Y are the menu buttons. Nothing opens yet. In a debug
build, F3 shows and hides the overlay.

## Maps

The maps are made in [Tiled](https://www.mapeditor.org/) and saved as JSON
in `assets/maps/`, one `.tmj` file per map. Section 6 of
[docs/PLAN.md](docs/PLAN.md) lists the layers they use.

Wild critters appear in `spawn_zone` objects. A spawn zone is a rectangle in
the `objects` layer with these custom properties:

| Property | Type | Needed | Meaning |
|---|---|---|---|
| `habitat` | string | yes | Which critters appear here: those in `assets/critters.json` with the same `habitat`, for example `pond` or `meadow` |
| `max` | int | no, 3 | How many critters the zone holds at once, at least 1 |
| `respawn_s` | float | no, 60 | Seconds a freed place waits before a new critter fills it, at least 0 |

The object's name is the zone's id. A zone without a name is called
`zone-<object id>`, and two zones on one map may not share a name. Critters
appear on open ground inside the zone and stay in it as they hop about. A
zone without area or without `habitat`, or with a bad `max` or `respawn_s`,
stops the map from loading. `make test` also checks the zones on the real
maps: each zone has to lie inside its map, keep clear of warps and player
spawns, have open ground you can walk to, and have a habitat that at least
one species in `critters.json` lives in by day.

Enemies stand on `enemy` point objects in the `objects` layer, with these
custom properties:

| Property | Type | Needed | Meaning |
|---|---|---|---|
| `enemy` | string | yes | Which enemy stands here: an `id` from `assets/data/enemies.json`, for example `sulky_slime` |
| `explorer_only` | bool | no, false | Leave this enemy out on Little Explorer, so rooms hold fewer enemies for younger players |

An enemy object without `enemy` stops the map from loading. On Explorer
every enemy object appears; on Little Explorer those with `explorer_only`
set to true are left out, at the start, after a warp and after a nap.

In a debug build (`make run TAGS=debug`) the overlay outlines every spawn
zone of the current map in red, next to the outline of where you may walk,
and counts the wild critters present.

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
  Draw the `side` view facing right; the game mirrors it about the feet for
  a character facing left.
- Every other drawing under `assets-src/sprites/` (props, effects, ui,
  emotes) and `assets-src/tiles/` is exported whole, as one PNG of its page at
  the same path under `assets/`.

Characters are built from parts that the game moves in code, not from drawn
frames. The player critter is the fox: its part sheet
`assets/sprites/critters/fox/fox.json` is hand-written in the format `make art`
writes, with placeholder boxes and pivots, until `fox.svg` is drawn and
exported over it. The game reads where each part sits and turns from that
file but still draws every part as a shape in code; the exported part PNGs are
not loaded yet.

A character's own data file, `assets/data/rigs/<id>.json`, picks its motions,
sizes and speeds, so a new critter is two data files and no code. Its
`sprite` names the part sheet (`"critters/fox"` reads
`sprites/critters/fox/fox.json`), and its `animations` have an `idle` and a
`walk` list. Each entry is a shared motion with `parts` (the part names it
moves, in every view that has them; none means the whole character),
`period` (seconds), `phase` (a fraction of a period) and:

- `breathe`: squashes its parts about their pivots, taller and thinner, then
  shorter and wider; `size` is the scale fraction.
- `blink`: closes its parts for the last `duration` seconds of each period.
- `bob`: lifts its parts smoothly by up to `size` world pixels once per
  period.
- `patter`: lifts its parts by up to `size` world pixels for the first half
  of each period and rests them for the second half.
- `wiggle`: turns its parts about their pivots by up to `size` degrees each
  way.

The game plays idle while the critter stands and walk while it moves, and
blends between the two over a tenth of a second.

A drawing is exported again when its SVG is newer than its JSON or PNG;
`make art ART_ARGS=-force` exports everything. If Inkscape is not on your
`PATH`, give its path, for example on a Mac
`make art INKSCAPE=/Applications/Inkscape.app/Contents/MacOS/inkscape`.

## Licence

The code is under the MIT licence in [LICENSE](LICENSE). Fonts and other
assets keep their own licences, listed in [docs/CREDITS.md](docs/CREDITS.md).
