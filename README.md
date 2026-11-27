# Critter Kingdom

A cosy top-down adventure for children aged 5 to 9, written in Go with
[Ebitengine](https://ebitengine.org/). It runs in a browser on a tablet and as
a desktop game. The plan, the roadmap and the reasons behind them are in
[docs/PLAN.md](docs/PLAN.md).

Status: milestone M0, "Hello Critter". A placeholder critter walks around a
meadow, hops and says hello.

## Getting started

You need Go 1.25 or later and nothing else: no C compiler and no Node.

```sh
make run              # play on the desktop
make run TAGS=debug   # with the debug overlay, F3 hides it
make run ARGS=-pad    # with the touch pad instead of tap-to-walk
make serve            # play in a browser at http://localhost:8080/web/
make test             # run the tests
make check            # what CI runs: gofmt, go mod tidy, vet and tests
make web              # build the browser bundle into build/web
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
resolution. Real art will be drawn in [Inkscape](https://inkscape.org/) and
exported to PNG; [docs/ART-BIBLE.md](docs/ART-BIBLE.md) sets the palette, sizes
and shape rules, and section 6 of [docs/PLAN.md](docs/PLAN.md) has the asset
pipeline plan.

## Licence

The code is under the MIT licence in [LICENSE](LICENSE). Fonts and other
assets keep their own licences, listed in [docs/CREDITS.md](docs/CREDITS.md).
