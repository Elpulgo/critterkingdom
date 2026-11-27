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
