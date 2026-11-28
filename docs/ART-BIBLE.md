# Art Bible

This is the one-page rulebook for Critter Kingdom's art, expanded from section
6.1 of [PLAN.md](PLAN.md). Read it before drawing a tile, a prop, a critter, an
enemy or a UI piece, and keep the palette and the style sheet open next to
Inkscape while you work.

## Mood

Five words set the feel of every drawing: **sunny, soft, round, friendly,
storybook**. The world is a warm afternoon in a picture book for
5-to-9-year-olds. Even the caves and the grumpy enemies belong to the same
gentle world.

## Look

Smooth vector shapes drawn at full screen resolution, never scaled-up pixels.
Real art is drawn in Inkscape at **8 px per world pixel** and exported to PNG,
which is 1:1 at scale 8 (a 2048 by 1536 iPad) and a little softer at other
scales. UI and text are drawn at full screen resolution too, with UI art exported
at **2 px per CSS pixel**.

## Palette

Every colour in the game comes from **Endesga 32** by ENDESGA, stored in the
repo at [`assets-src/palettes/endesga_32.gpl`](../assets-src/palettes/endesga_32.gpl).
Never eyedrop colours from outside the palette, never type an off-palette hex
code, and never use gradients or blur filters. Soft shadows use palette ink
`#181425` with lowered opacity (`opacity` or `fill-opacity`).

### Loading the palette into Inkscape as swatches

1. Find your user palettes folder in Inkscape under **Edit > Preferences >
   System > User palettes**:
   - Linux: `~/.config/inkscape/palettes/`
   - macOS: `~/Library/Application Support/org.inkscape.Inkscape/config/inkscape/palettes/`
   - Windows: `%APPDATA%\inkscape\palettes\`
2. Copy `assets-src/palettes/endesga_32.gpl` into that folder.
3. Restart Inkscape.
4. Open **View > Swatches** (or click the tiny arrow at the right end of the
   bottom palette bar) and pick **Endesga 32**.

### All 32 colours, their roles and their outlines

Characters and props get an 8 px outline (1 world pixel) in a darker palette
colour from the same family, never `#181425` ink and never black. Ground tiles
get no outline. The table below lists all 32 colours in palette order, gives
each colour its role (or marks it as spare for later biomes and costumes), and
pairs each fill with its darker outline colour.

| # | Hex | Swatch name | Role | Paired darker outline |
|---|---|---|---|---|
| 1 | `#be4a2f` | Rust red | Fox fur outline, terracotta roof shade, warm dark accent | `#733e39` (when used as a fill) |
| 2 | `#d77643` | Warm ochre | Fox fur shade, dirt path shade, terracotta pot fill | `#be4a2f` |
| 3 | `#ead4aa` | Cream | Fox belly and muzzle, bunny fur highlight, UI panel fill, dialogue card | `#c28569` (UI panel) or `#be4a2f` (on fox) |
| 4 | `#e4a672` | Light tan | Dirt path fill, wooden plank fill, fox fur highlight, sand shade | `#b86f50` |
| 5 | `#b86f50` | Wood brown | Tree trunk fill, fence and sign fill, bridge planks, UI frame accent | `#733e39` |
| 6 | `#733e39` | Bark brown | Tree trunk outline and deep shade, dark wood fill, hedgehog quills | `#3e2731` (when used as a fill) |
| 7 | `#3e2731` | Cocoa | Dialogue text, nose and mouth lines, darkest warm outline for bark | — (line and text colour) |
| 8 | `#a22633` | Berry crimson | Berry outline, apple and red mushroom cap outline, heart icon outline | `#3e2731` (when used as a fill) |
| 9 | `#e43b44` | Poppy red | Berry fill, red mushroom cap, heart icon fill | `#a22633` |
| 10 | `#f77622` | Fox orange | Fox critter primary fur fill, autumn leaf fill | `#be4a2f` |
| 11 | `#feae34` | Marigold | Flower heart fill, fox ear/cheek warm highlight, star and coin shade | `#d77643` |
| 12 | `#fee761` | Sunny yellow | Star fill, firefly glow, sunflower petals, tree canopy highlight dab | `#feae34` |
| 13 | `#63c74d` | Meadow green | Base grass tile fill, tree canopy highlight, frog fill | `#265c42` (on props/critters; none on ground) |
| 14 | `#3e8948` | Leaf green | Tree canopy fill, bush fill, grass tuft and blade detail on ground | `#265c42` (none on ground tufts) |
| 15 | `#265c42` | Pine green | Tree and bush outline, canopy shade, hedge fill | `#193c3e` (when used as a hedge fill) |
| 16 | `#193c3e` | Deep teal | Deep forest and hedge outline, cave moss shade | — (dark outline colour) |
| 17 | `#124e89` | Deep water blue | Pond water outline and deep water shade, night tint | `#262b44` (when used as a fill) |
| 18 | `#0099db` | Pond blue | Water tile fill, Splash Boots fill, bluebird feathers | `#124e89` |
| 19 | `#2ce8f5` | Sky cyan | Water ripple highlight, crystal and magic wand sparkle | `#0099db` |
| 20 | `#ffffff` | White | Eye highlight, daisy petals, sticker border, speech bubble fill, cloud | `#c0cbdc` (or `#c28569` on warm UI) |
| 21 | `#c0cbdc` | Pale slate | Stone and pebble highlight, white fur shade, metal blade fill | `#5a6988` |
| 22 | `#8b9bb4` | Cool grey | Rock and boulder fill, pebble fill, stone wall fill | `#5a6988` |
| 23 | `#5a6988` | Slate blue | Rock outline and shade, cave floor fill | `#3a4466` (when used as a fill) |
| 24 | `#3a4466` | Dusk indigo | Cave wall fill, slate rock dark outline | `#262b44` (when used as a fill) |
| 25 | `#262b44` | Midnight blue | Cave wall outline and deep shadow fill, night sky | — (dark outline colour) |
| 26 | `#181425` | Ink | Pupils, eye shapes, soft ground and UI drop shadows (with opacity) | — (never used as a character/prop outline) |
| 27 | `#ff0044` | Bright magenta | Spare (rare flower accent, magic berry) | `#a22633` |
| 28 | `#68386c` | Plum purple | Grumpy enemy (Sulky Slime / Thistle Beast) outline and shade | `#3e2731` (when used as a fill) |
| 29 | `#b55088` | Orchid purple | Grumpy enemy (Sulky Slime) body fill, thistle flower fill | `#68386c` |
| 30 | `#f6757a` | Blossom pink | Pink flower petals, inner ear fill, rosy cheeks, slime highlight | `#b55088` (on enemy) or `#a22633` (on props) |
| 31 | `#e8b796` | Peach | Inner ear highlight, portrait background card fill, sandy beach fill | `#c28569` |
| 32 | `#c28569` | Warm taupe | UI panel and portrait frame outline, path pebble, bunny fur shade | `#733e39` (when used as a fill) |

## Sizes

World art is measured in **world pixels** and drawn in Inkscape at **8 px per
world pixel** (set Inkscape's Document Properties to `px` units with scale `1`,
so 1 px in the drawing is 1 pixel in the exported PNG). UI art is measured in
**CSS pixels** (like the 64 CSS px minimum touch target in PLAN §3) and drawn at
**2 px per CSS pixel**.

### World sizes (at 8 px per world pixel)

| Item | World pixels | Drawing pixels (px) | Notes |
|---|---|---|---|
| Terrain tile | 16 by 16 | 128 by 128 px | Seamless wrap across all four edges |
| Player and villagers | 32 by 32 box | 256 by 256 px | Feet at bottom centre; collision box about 12 by 10 world px (96 by 80 px) at the feet |
| Small critters and enemies | 16 by 16 or 24 by 24 | 128 by 128 px or 192 by 192 px | Sulky Slime fits 16 or 24 world px; birds and frogs use 16, larger critters 24 |
| Boss (Grumbling Grub) | 48 by 48 | 384 by 384 px | Glow Cave mini-dungeon boss |
| Tree | 32 by 48 | 256 by 384 px | Canopy in the upper two thirds, trunk base at bottom centre |
| House | 64 by 64 | 512 by 512 px | Village exterior prop |
| Mock screen (16:9 view) | 320 by 180 | 2560 by 1440 px | Standard 16:9 world view used by the style sheet |

### UI sizes (at 2 px per CSS pixel)

| Item | CSS pixels | Drawing pixels (px) | Notes |
|---|---|---|---|
| Minimum touch target | 64 by 64 CSS px | 128 by 128 px | About 1 cm on a tablet |
| Portrait | 160 by 160 CSS px | 320 by 320 px | Three expressions (`neutral`, `happy`, `sad`) by swapping eyes and mouth |
| Sticker | 96 by 96 CSS px | 192 by 192 px | Includes a white (`#ffffff`) sticker border |
| HUD icon | 32 by 32 CSS px | 64 by 64 px | Quest tracker, petals, counters |
| Menu icon | 64 by 64 CSS px | 128 by 128 px | Backpack, Snack Time choices, dialogue choices |

## Shapes

- **Build everything from simple shapes**: circles, ellipses and rounded
  rectangles or rounded triangles, just like the M0 placeholder critter in
  `internal/world/art.go`.
- **Proportions**: big heads (about half the character's total height, roughly
  16 world pixels / 128 px across in a 32 by 32 box), compact round bodies
  (about 10 world pixels / 80 px across), small paws and feet at the bottom
  centre of the box.
- **Faces**: big dark eyes (`#181425`) with a crisp white (`#ffffff`) highlight
  dot near the top-left. Rounded noses and ears; no sharp angles anywhere.
- **Enemies look grumpy, never scary**: furrowed eyebrows tilted inward and a
  little pout. No fangs, no claws, no blood, no glowing red eyes. When an enemy
  is defeated, its grumpy face swaps to a happy smile as it poofs into flowers
  and runs off.

## Lines and shading

- **Outlines**: characters and props get an outline **1 world pixel thick (8 px
  in the drawing)** in a darker version of their fill colour from the palette
  (for example, `#f77622` fox fur outlined in `#be4a2f`, `#3e8948` tree canopy
  outlined in `#265c42`, `#b55088` enemy outlined in `#68386c`). Never outline
  characters or props in `#181425` ink or black.
- **Ground tiles get no outline**: grass, dirt, sand and water meet with soft,
  flat shapes so the ground stays calm underneath the characters.
- **Stroke joins and caps**: always round (`stroke-linejoin="round"` and
  `stroke-linecap="round"`).
- **Minimum thickness**: nothing thinner than **half a world pixel (4 px in the
  drawing)**, so eyebrows, mouths, grass blades and small details survive on a
  small laptop screen.
- **Shading**: flat fills with at most **one shade and one highlight** per
  material. Light comes from the **top-left**: highlights sit on the upper-left
  curve, shades sit on the lower-right curve, and a soft ellipse shadow
  (`#181425` at low opacity) sits under the feet. No textures, gradients or
  airbrushing, so young children get an instant, clean read.

## Contrast

Ground tiles are lighter and less saturated than characters, so characters pop
at a glance.

In Endesga 32, `#63c74d` is the lightest green available for meadow grass. Do
not invent an off-palette pale green tint. Instead, keep the `#63c74d` ground
completely flat and outline-free, keep `#3e8948` grass tufts small and sparse,
and make characters and props pop against the meadow through:

1. **Warm and contrasting hues** (orange `#f77622`, cream `#ead4aa`, orchid
   purple `#b55088`) against the green ground.
2. **Their 8 px darker outlines**, which separate every character and prop from
   the un-outlined ground.
3. **A soft dark ground shadow** (`#181425` with `fill-opacity`) beneath their
   feet or base.

## Animation

Characters are **sets of vector parts** (head, body, ears, tail, feet, eyes,
mouth, stick) that the game engine moves, turns and squashes in code instead of
frame-by-frame sprite sheets:

- **Views**: draw `down`, `up` and `side` views for movable characters; the
  engine mirrors the `side` view horizontally (`flipX`).
- **Part naming and pivots**: each part is an Inkscape object with an id named
  `<view>_<part>` (such as `down_head`, `down_body`, `down_ear_left`,
  `down_tail`), plus spare expression parts (`down_eyes_closed`,
  `down_eyes_happy`). Mark the point each part turns around with a tiny circle
  named `<view>_<part>_pivot`, and the feet anchor with `<view>_feet`.
- **Player and villager motions**:
  - *Idle*: a slow breathing squash and a blink every few seconds.
  - *Walk*: a gentle vertical bob with pattering feet and bouncing ears.
  - *Swing*: a forward lean with the stick swept through an arc.
  - *Hurt*: a white flash and a brief squash.
  - *Small critters*: hop in an arc and squash on landing.
- **Enemy motions**: `idle`, `move`, `telegraph`, `attack`, `hurt` and `poof`,
  built from parts the same way. The **telegraph** is the most important
  animation in the game (at least 0.6 seconds of puffing up, shaking or
  flashing) so a 5-year-old clearly sees the attack coming.
- **Shared emotes**: floating icons above any character (heart, exclamation,
  zzz, sweat drop, music note).
- **Part budget**: playable critter about 30 parts across 3 views (roughly 8
  parts per view plus spare eyes, mouths and the stick); villagers about 20 (or
  8 if they only face down); small critters 5 to 10; enemies about 12; the boss
  about 20.

## References

- **Cozy Grove** and **Spiritfarer**: soft, hand-drawn, warm storybook coziness.
- **Sago Mini** and **Toca Boca**: bold, simple rounded shapes and big friendly
  faces that young children read at a glance.
- **Animal Crossing**: toy-like proportions (big heads, stubby limbs, expressive
  eyes) and cheerful village warmth.
- **The Legend of Zelda: Link's Awakening (Switch)** and **Tunic**: clean
  top-down readability, bright saturated colour against calm ground, and clear
  silhouettes.

When gathering visual references locally, drop screenshots into a PureRef board
next to Inkscape (PureRef files stay local and are not committed to the repo).

## Style sheet

Before drawing individual character part files or tilesets, we test the look on
one mock screen at [`assets-src/style/style_sheet.svg`](../assets-src/style/style_sheet.svg).
It places all six core elements on a **320 by 180 world pixel mock screen**
(**2560 by 1440 px** at 8 px per world pixel):

1. `grass_tile`: one 128 by 128 px seamless grass tile repeated across the ground.
2. `tree`: one 256 by 384 px (32 by 48 world px) rounded tree.
3. `critter`: the fox player critter inside a 256 by 256 px (32 by 32 world px)
   box, with big head, cream belly and muzzle, 8 px `#be4a2f` outline and big
   highlighted eyes.
4. `enemy`: a grumpy Sulky Slime inside a 192 by 192 px (24 by 24 world px) box,
   in orchid purple `#b55088` with a `#68386c` outline, furrowed eyebrows and a
   pout (no fangs).
5. `ui_panel`: a rounded dialogue box across the bottom with a soft offset
   shadow and two short lines of Fredoka dialogue text.
6. `portrait`: a 320 by 320 px (160 by 160 CSS px) character portrait framed on
   the left side of the dialogue panel.

### Viewing the style sheet on the tablet

The SVG references the game's embedded font at `../../assets/fonts/Fredoka.ttf`,
so serve it from the repository root with any static file server (note:
`make serve` runs `wasmserve` for the Go game, not static files under
`assets-src/`):

1. In a terminal at the repository root, run:
   ```sh
   python3 -m http.server 8080
   ```
2. Find your computer's LAN IP address (for example `ipconfig getifaddr en0` on
   macOS or `hostname -I` on Linux).
3. On the tablet's browser on the same Wi-Fi, open:
   ```text
   http://<computer-ip>:8080/assets-src/style/style_sheet.svg
   ```

### Human sign-off checklist

- [ ] Copy `assets-src/palettes/endesga_32.gpl` into Inkscape's user `palettes`
      folder and select **Endesga 32** in the Swatches panel.
- [ ] Open `assets-src/style/style_sheet.svg` on the tablet over Wi-Fi and check
      that the outlines, eyes and text read cleanly at arm's length.
- [ ] Show the style sheet on the tablet to the kids and ask which parts they
      like best.
- [ ] Iterate on the shapes and proportions in Inkscape until the sheet feels
      cute, keeping `go test ./assets-src/...` green so every colour stays on
      the palette.

## UI, effects and emotes

The UI pieces, icons, petals, stars, touch pad, logo, effects and shared emotes
from PLAN §6.5 are hand-written SVGs under `assets-src/sprites/`, one drawing
per piece. `make art` exports each one as one PNG of its page at the mirrored
path (for example `assets-src/sprites/ui/icons/berry_32.svg` becomes
`assets/sprites/ui/icons/berry_32.png`) at 1 drawing px per PNG px. So each
SVG's page is the piece's exact size: UI art at 2 px per CSS pixel, effects
and emotes at 8 px per world pixel, with a transparent background.

### Pieces

Paths are under `assets-src/sprites/`. This table lists the same rows, in the
same order, as the `uiArt` manifest in
[`assets-src/sprites/ui_art_test.go`](../assets-src/sprites/ui_art_test.go);
change both together.

| Drawing | Drawing size | CSS or world size | Notes |
|---|---|---|---|
| `ui/panel.svg` | 96 by 96 px | 48 by 48 CSS px | 9-slice panel |
| `ui/button.svg` | 96 by 96 px | 48 by 48 CSS px | 9-slice button, normal |
| `ui/button_pressed.svg` | 96 by 96 px | 48 by 48 CSS px | 9-slice button, pressed: the same shape 4 px lower, no drop shadow, the shade colour as fill |
| `ui/arrow_next.svg` | 64 by 64 px | 32 by 32 CSS px | "Next" arrow, HUD size |
| `ui/checkmark.svg` | 64 by 64 px | 32 by 32 CSS px | Checkmark, HUD size |
| `ui/petal_full.svg` | 64 by 64 px | 32 by 32 CSS px | Petals HUD |
| `ui/petal_empty.svg` | 64 by 64 px | 32 by 32 CSS px | Same outline as the full petal, muted fill |
| `ui/star_full.svg` | 96 by 96 px | 48 by 48 CSS px | The boss's row of stars |
| `ui/star_empty.svg` | 96 by 96 px | 48 by 48 CSS px | Same outline as the full star, muted fill |
| `ui/joystick_ring.svg` | 256 by 256 px | 128 by 128 CSS px | `padRadius` 64 CSS px; transparent centre so the world shows through |
| `ui/joystick_knob.svg` | 112 by 112 px | 56 by 56 CSS px | `padKnobRadius` 28 CSS px |
| `ui/paw_button.svg` | 192 by 192 px | 96 by 96 CSS px | `pawButtonRadius` 48 CSS px; no pressed drawing |
| `ui/logo.svg` | 256 by 256 px | 128 by 128 CSS px | Emblem only: a crowned paw print in a leafy badge; the engine sets the title in Fredoka |
| `ui/icons/speech_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/footsteps_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/berry_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/hat_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/heart_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/star_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/sticker_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/backpack_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/gear_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/speaker_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/flag_sv_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/flag_en_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/snacks_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/key_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/lantern_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/zzz_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/stick_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/sturdy_stick_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/wand_32.svg` | 64 by 64 px | 32 by 32 CSS px | HUD icon, a simplified redraw of the `_64` drawing |
| `ui/icons/speech_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/footsteps_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/berry_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/hat_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/heart_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/star_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/sticker_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/backpack_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/gear_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/speaker_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/flag_sv_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/flag_en_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/snacks_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/key_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/lantern_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/zzz_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/stick_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/sturdy_stick_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `ui/icons/wand_64.svg` | 128 by 128 px | 64 by 64 CSS px | Menu icon |
| `effects/swoosh_arc.svg` | 256 by 256 px | 32 by 32 world px | The player's box; drawn for a swing to the right |
| `effects/poof_cloud.svg` | 192 by 192 px | 24 by 24 world px | The largest normal enemy |
| `effects/flower_burst.svg` | 192 by 192 px | 24 by 24 world px | The largest normal enemy |
| `effects/spore_cloud.svg` | 192 by 192 px | 24 by 24 world px | The largest normal enemy |
| `effects/hit_spark.svg` | 128 by 128 px | 16 by 16 world px | One tile |
| `effects/dust_puff.svg` | 128 by 128 px | 16 by 16 world px | One tile |
| `emotes/heart.svg` | 128 by 128 px | 16 by 16 world px | Floats above a 32 world px character |
| `emotes/exclamation.svg` | 128 by 128 px | 16 by 16 world px | Floats above a 32 world px character |
| `emotes/zzz.svg` | 128 by 128 px | 16 by 16 world px | Floats above a 32 world px character |
| `emotes/sweat.svg` | 128 by 128 px | 16 by 16 world px | Floats above a 32 world px character |
| `emotes/music_note.svg` | 128 by 128 px | 16 by 16 world px | Floats above a 32 world px character |

### 9-slice insets

`ui/panel.svg`, `ui/button.svg` and `ui/button_pressed.svg` are 9-slice
sources with **32 px (16 CSS px) insets on all four sides**. The corners,
outline, highlight and shade strips and drop shadow stay inside the insets, so
the middle 32 by 32 px is plain fill. A 9-slice draw keeps the four 32 by
32 px corners as they are, stretches the edges along their length and the
middle both ways, so one drawing makes a panel or button of any size. Nothing
in the game draws these images yet; the UI code that loads them must use
these insets.

### Effects are single drawings

Each effect is **one drawing, not a set of frames**, centred on its page. The
engine animates it in code: it scales, turns and fades the drawing about the
page centre (the same idea as characters, which are parts animated in code).
`effects/swoosh_arc.svg` is drawn for a swing to the right; the engine turns
or mirrors it for the other directions. Emotes are single drawings too,
centred on their pages, and float above a character.

### Lint

`go test ./assets-src/...` (part of `make test`) lints every drawing in the
manifest with `tools/art/artlint`. A test fails when:

- a listed drawing is missing, or its page is not its listed size (`width`
  and `height` must be whole numbers matching the row, with
  `viewBox="0 0 w h"`);
- a `fill`, `stroke`, `stop-color`, `flood-color` or `color` (as an attribute
  or inside `style`) is not `none` or an Endesga 32 colour;
- it has a `linearGradient`, `radialGradient`, `filter`, `image`, `text` or
  `pattern` element;
- a `stroke-width` is thinner than 4 px;
- an `.svg` under `ui/`, `effects/` or `emotes/` is not in the manifest.

So when you add or resize a piece, add or change its manifest row and its row
in the table above. The lint can't judge how a drawing looks; the checklist
below is for that.

### Human sign-off checklist for UI, effects and emotes

- [ ] Install Inkscape 1.x and run `make art ART_ARGS=-force` (add
      `INKSCAPE=/path/to/inkscape` when it is not on `PATH`).
- [ ] Check the PNGs under `assets/sprites/ui/`, `assets/sprites/ui/icons/`,
      `assets/sprites/effects/` and `assets/sprites/emotes/`: one for every row
      in the table above, each at its drawing size, with a transparent
      background and nothing cut off at the page edge.
- [ ] View the icons at 32 CSS px on the tablet. A 64 by 64 px `_32` PNG shows
      at 32 CSS px with `width="32"`, so put a throwaway page in the ignored
      `build/` folder, for example `build/icons.html` with one
      `<img src="../assets/sprites/ui/icons/berry_32.png" width="32">` per
      icon. Serve the repository root as in
      [Viewing the style sheet on the tablet](#viewing-the-style-sheet-on-the-tablet)
      and open `http://<computer-ip>:8080/build/icons.html`. Check that every
      icon reads at arm's length and that the outlines don't fill in.
- [ ] Show the icons, petals, stars, pad, logo, effects and emotes on the
      tablet to the kids and ask what each one means. Redraw any they can't
      name, keeping `go test ./assets-src/...` green.
## World art set

The terrain tiles, ground decorations, props and interior pieces of PLAN §6.4
and §6.7 are drawn as SVG sources under `assets-src/`, following every rule
above. Sizes below are in world pixels; the drawings are 8 times larger.

### Tilesets

Each tileset is one SVG sheet, 8 tiles wide, of 16 by 16 world px tiles
(128 by 128 px). Each tile is a `<g>` whose `id` is the tile name, placed in
its own grid cell, and its Tiled tile id is `row * 8 + column`.

| Sheet | Size (px) | Tiles |
|---|---|---|
| `assets-src/tiles/tileset_outdoor.svg` | 1024 by 1024 (8 by 8) | 60: `grass`, `grass_2`, `grass_3`, `grass_4`, `path`, `sand`, `water_1`, `water_2`, `water_3`; 14 transition tiles each for grass/path (`grass_path_XXXX`), grass/water (`grass_water_XXXX`) and sand/water (`sand_water_XXXX`); cliffs `cliff_nw`, `cliff_n`, `cliff_ne`, `cliff_w`, `cliff_e`, `cliff_sw`, `cliff_s`, `cliff_se`, `cliff_face` |
| `assets-src/tiles/tileset_cave.svg` | 1024 by 128 (8 by 1) | `cave_floor`, `cave_floor_2`, `cave_floor_3`, `cave_wall` |
| `assets-src/tiles/tileset_indoor.svg` | 1024 by 128 (8 by 1) | `floor`, `floor_2`, `wall`, `wall_top`, `window` |

**Transition naming**: a transition tile is named `<a>_<b>_<nw><ne><sw><se>`,
with one digit per corner in the order north-west, north-east, south-west,
south-east: `0` for terrain `a` and `1` for terrain `b`. For example
`grass_path_0111` is grass in the north-west corner and path in the other
three. Each pair has the 14 mixed codes `0001` to `1110`; the all-`0` and
all-`1` tiles are the shared plain tiles (`grass`, `path`, `sand`, `water_1`).
All 14 are drawn out in the sheet, from one edge and one corner shape per pair
placed with `<use>`, so the light stays top-left; never use Tiled's flip or
rotate on them.

### Ground decorations

15 decorations, each 16 by 16 world px (128 by 128 px), one SVG each in
`assets-src/sprites/decor/`: `tuft_1`, `tuft_2`, `tuft_3`, `flower_white`,
`flower_pink`, `flower_yellow`, `flower_red`, `pebble_1`, `pebble_2`,
`pebble_3`, `leaf_1`, `leaf_2`, `leaf_3`, `clover`, `twig`.

### Props

30 props, one SVG each in `assets-src/sprites/props/`, every one with the 8 px
outline:

| Group | Prop | World pixels | Drawing pixels (px) |
|---|---|---|---|
| Outdoor | `tree` | 32 by 48 | 256 by 384 |
| Outdoor | `tree_small` | 32 by 32 | 256 by 256 |
| Outdoor | `bush`, `berry_bush_full`, `berry_bush_empty`, `rock`, `rock_small`, `fence_h`, `fence_v`, `fence_post`, `fence_gate`, `sign` | 16 by 16 | 128 by 128 |
| Outdoor | `house` (door drawn in) | 64 by 64 | 512 by 512 |
| Puzzle | `chest_closed`, `chest_open`, `pot`, `push_block`, `flower_switch_off`, `flower_switch_on`, `mushroom_off`, `mushroom_on`, `bramble`, `room_door_open`, `room_door_locked` | 16 by 16 | 128 by 128 |
| Puzzle | `plank_bridge` | 32 by 16 | 256 by 128 |
| Puzzle | `sunflower_checkpoint` | 16 by 32 | 128 by 256 |
| Interior | `bed` | 16 by 32 | 128 by 256 |
| Interior | `table` | 32 by 16 | 256 by 128 |
| Interior | `rug` | 32 by 32 | 256 by 256 |
| Interior | `shelf` | 16 by 32 | 128 by 256 |

The interior floor, walls and window are tiles in `tileset_indoor.svg`, not
props.

### Tiled tilesets

The `.tsx` files sit next to their drawings and point at the PNGs that
`make art` writes under `assets/`:

- `assets-src/tiles/tileset_outdoor.tsx`, `tileset_cave.tsx` and
  `tileset_indoor.tsx`: regular tilesets on `assets/tiles/<sheet>.png`, with a
  `type` on every drawn tile. The outdoor one also animates `water_1` through
  `water_1`, `water_2`, `water_3` at 300 ms each, and holds the corner
  terrain set `ground` (colours grass, path, water and sand) that paints all
  three transitions with one brush.
- `assets-src/sprites/decor/decor.tsx` and
  `assets-src/sprites/props/props.tsx`: image-collection tilesets, one tile per
  drawing, with the drawing's name as its `type` and its own image
  `assets/sprites/decor/<name>.png` or `assets/sprites/props/<name>.png` at
  the drawing's page size.

### Export, tests and the game

- **The PNGs are not committed.** Only the SVG and `.tsx` sources are in the
  repo. A human with Inkscape 1.x produces the PNGs with `make art`
  (`make art ART_ARGS=-force` to export everything again), which writes
  `assets/tiles/*.png`, `assets/sprites/decor/*.png` and
  `assets/sprites/props/*.png`.
- **The game still draws placeholders.** `internal/world/art.go` keeps drawing
  the shapes in code until the follow-up item "Load exported tiles and props
  in the world scene" loads the exported PNGs. That item comes after issue #2
  (maps) and issue #16 (y-sorting and the `above` layer), and after the PNGs
  are exported and committed.
- **`go test ./assets-src/...` checks the rules a machine can check** (it runs
  as part of `make test`): every colour is in Endesga 32; no gradients,
  filters, `<image>` or `<text>`; every drawing has its listed page size and
  every file is listed in the test's manifest; tilesets have no strokes;
  nothing is thinner than 4 px; every prop has an 8 px outline with round
  joins and caps, never in ink or black; every `.tsx` tile matches its
  drawing's grid cell or image size; the `ground` terrain set covers all 16
  corner combinations of each pair, matching the tile names; and the water
  animation runs through its three frames. A new drawing needs a row in the
  manifest in `assets-src/art_test.go`.

### Human sign-off checklist

How the art looks is not tested. Before the follow-up item loads it into the
game, a human checks:

- [ ] Open each sheet in Inkscape (`tileset_outdoor.svg`, `tileset_cave.svg`,
      `tileset_indoor.svg`) and each decoration and prop SVG, and check the
      shapes, colours, outlines and top-left light look right and feel
      sunny, soft, round, friendly and storybook.
- [ ] Run `make art`, open `tileset_outdoor.tsx` in Tiled, and paint a test
      map with the `ground` terrain brush: check that grass/path, grass/water
      and sand/water edges join without seams, that plain tiles wrap
      seamlessly, and that the water animates.
- [ ] Look at the props and decorations on the tablet: serve the repository
      root as in "Viewing the style sheet on the tablet" and open
      `http://<computer-ip>:8080/assets-src/sprites/props/<name>.svg` (and
      `.../decor/<name>.svg`), or place them from `decor.tsx` and `props.tsx`
      on the Tiled test map and view a screenshot of it there. Check that
      they pop against the ground, read clearly at arm's length, and look
      right at the sizes in the tables above.
- [ ] Fix anything that looks wrong in the SVG sources, keeping
      `go test ./assets-src/...` green, and run `make art` again.
