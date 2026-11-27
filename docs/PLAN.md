# Critter Kingdom: Project Plan

Status: draft v0.1, 2026-11-27. This is the living plan for the game.

How to use it: each milestone in section 5 is meant to become its own spec and
branch. Tick the boxes as work lands. When a decision changes, update the
decision log in section 9 instead of rewriting history.

## 0. In one paragraph

Critter Kingdom is a cozy 2D top-down adventure for children aged 5 to 9. You
play a small critter (fox, bunny or hedgehog) in a kingdom of talking animals.
You explore a pixel-art world, chat with villagers, take on small quests,
befriend wild critters and collect them in a sticker book, and solve gentle
puzzles in tiny dungeons. It is written in Go with Ebitengine, compiled to
WebAssembly and played in the browser on a tablet or laptop. No accounts, no
ads, no purchases, no chat with strangers.

## 1. Vision

### 1.1 What we borrow

| Inspiration | What we take | How it becomes kid-sized |
|---|---|---|
| Zelda (Link's Awakening, A Link to the Past) | Top-down exploration, items that open new areas, small dungeons with puzzles, villagers to talk to | One puzzle idea per room. No game over. "Enemies" are grumpy critters you boop into friends. |
| Pokémon Go | Creatures hidden in the world, finding them in their habitats, a collection to fill, rarity, "only comes out at night" | No GPS. Critters live in the game world. Befriend by picking the right snack, not by throwing balls. The collection is a sticker book. |
| Kitaria Fables | Cast of cute animal villagers, village-and-quests loop, cozy tone, light action, co-op | Quests take 3 to 8 minutes and are shown with pictures. Co-op is a later stretch goal. |

### 1.2 Design pillars for ages 5 to 9

1. **Never stuck, never punished.** No game over, no timers, nothing is ever lost. A stuck player always has a visible hint.
2. **Readable without reading.** Every quest step has an icon. Every speaker has a portrait with an expression. Every line can be read aloud. Text is short (max two lines) and large.
3. **One stick, one button.** Move and "paw" (interact). Menus are big picture buttons. Nothing needs double-tap, long-press or button combos.
4. **Kindness is the mechanic.** You help, befriend and calm. Winning looks like hearts and stickers.
5. **Small loops, frequent smiles.** Something rewarding happens every minute. A quest fits in one sitting.
6. **Safe by design.** Offline-first, no accounts, no network needed, saves live on the device.
7. **Playtest with real kids** at the end of every milestone. Watch, don't explain.

### 1.3 Deliberately out of scope for v1

Real-world GPS play, online multiplayer, chat, accounts, monetisation, farming
and crafting systems, inventory-management puzzles, reflex-based combat, 3D.

## 2. The game

### 2.1 World

A small kingdom seen from above. The hub village **Acorn Hollow** has five or
six villagers and your own little house. Around it, biomes unlock over time:

| Biome | Unlocked by | Critters | Hook |
|---|---|---|---|
| Sunny Meadow | start | bird, snail, squirrel | tutorial quests, berry bushes |
| Lily Pond | start | frog, firefly (night only) | first companion with an ability |
| Whispering Woods | "Fix the Bridge" quest | owl, fox kit, mushroom sprite | brambles need the Snippers |
| Glow Cave | firefly companion or lantern | glow-worm, friendly bat | first mini-dungeon, five rooms |
| Pebble Beach | Splash Boots | crab, seal pup | later content |

Each biome is one or two maps of roughly 40 by 30 tiles, with two or three
critters, one or two quests and one secret.

Time of day is **not** a real-time clock. It changes when you sleep in your bed
(or through quest events). The firefly only comes out at night? Go home,
sleep, and it is night. Kids understand this, and it keeps the game
deterministic.

### 2.2 Your critter

- Pick a species (fox, bunny, hedgehog) and a colour. Pick a name from a list
  of presets, or type one (pre-readers cannot type).
- Three save slots, so siblings each have their own critter.
- Cosmetic rewards: hats, scarves, stickers for your house wall.
- A **companion**: one befriended critter follows you. You can pet it
  (hearts). A few companions have an ability used by puzzles and quests: the
  firefly lights dark places, the bird points at hidden things, the frog
  fetches from shallow water.

### 2.3 Talking

- Walk near a villager and a "…" bubble appears above them. Press paw: a
  dialogue box with portrait, name and text that types out. Tap anywhere to
  continue.
- Choices: at most two, as big buttons with icons (check and cross, or
  pictures).
- A speaker button (or auto mode) reads the line aloud using the browser's
  built-in speech voices. Zero asset cost, works in Swedish and English.
- Villagers have a few lines of daily chatter plus quest lines. They face you
  when talking and wander in small loops otherwise.

### 2.4 Quests

Step types, enough for everything in v1: **talk to**, **go to**, **collect
N**, **deliver X to Y**, **befriend critter**, **find hidden thing**
(companion gives hot/cold hints), **solve puzzle room**.

Presentation:

- Quests come from talking to villagers. One main quest plus at most two side
  quests active at a time.
- HUD tracker: the current step as icon plus counter (berry 2/5).
- Helper arrow at the screen edge pointing to the target (on by default in
  "Little Explorer" mode).
- Quest page in the backpack: one row per quest, pictures and checkmarks.
- Completion: fanfare, confetti, happy portrait, reward (sticker, item, hat or
  key item).

Example quests for the vertical slice:

| Quest | Giver | Teaches |
|---|---|---|
| The Missing Hat | Grandma Hedgehog | walk, talk, pick up, return (the hat is on the scarecrow) |
| Berry Pie | Baker Badger | collect five berries, counter UI |
| Scared Little Mouse | Pip the mouse | befriend the firefly at night, companion ability (light) |
| Fix the Bridge | Beaver Woodcutter | deliver three planks, area gating |
| The Sleepy Tree (main) | Owl Teacher | collect four Glow Seeds, one per mini-dungeon |

### 2.5 Wild critters and the Sticker Book (the Pokémon Go part)

- Each map has **spawn zones** per habitat with a weighted list of critters. A
  handful are present at once and respawn after a while. Rare ones sparkle.
- Some appear only under conditions: at night, after a quest, with a certain
  companion, or as the **daily visitor** (one special critter per real
  calendar day, a gentle reason to come back).
- Press paw near a critter to start **Snack Time**: three big snack icons, and
  the sticker book hints what the critter likes. Wrong: it giggles, try again
  (three tries, then it hops off and comes back later). Right: hearts,
  befriended, sticker unlocked. No reflexes needed.
- Optional second mini-game later: **Follow the Bounce**, a three-step
  Simon-says with coloured flowers.
- **Sticker Book**: a grid of silhouettes that fill in. Tap a sticker for its
  habitat icon, favourite snack and a fun fact read aloud. Progress "5/20".
- Any befriended critter can be chosen as companion.

### 2.6 Exploring and puzzles (the Zelda part)

- **Gating items** open regions in classic Zelda fashion: Snippers (cut
  brambles), Splash Boots (shallow water), lantern or firefly (dark), bridge
  planks (quest).
- **Puzzle pieces**, one idea per room: push a block onto a flower switch;
  light three mushrooms in the order a painting shows; find the key under a
  pot; step on tiles in order; bring a specific companion to wake a sleepy
  guardian.
- **Glow Cave** is the first mini-dungeon: five rooms, dark, ends with a Glow
  Seed and a hat.
- **Grumpy critters** (thistle-beasts, sulky slimes) wander some areas.
  Touching one bumps you back and costs a petal (you have three). Boop them
  with your stick and they turn into flowers or run off. At zero petals you
  take a nap and wake at the area entrance with full petals, nothing lost. A
  "No grumpies" toggle puts them all to sleep.

### 2.7 Session shape and saving

- Autosave at every quest step, map change and new sticker. The title screen
  has a huge "Continue" button.
- Natural stopping points every 3 to 8 minutes (quest done, sticker earned).
- Saves live in the browser (localStorage) and as a file on desktop builds.
  Later: export and import a save as a short "magic word" so it can move
  between devices.

## 3. Controls

| Input | Move | Paw (interact) | Backpack | Notes |
|---|---|---|---|---|
| Touch (default) | **Tap where to walk** (pathfinding). Tap a critter, villager or item to walk there and interact. Optional floating joystick. | Big paw-print button bottom-right (or just tap the thing) | Big backpack button top-right | Point-and-click is the easiest thing a 5-year-old can do |
| Keyboard | Arrows or WASD | Space, Enter or E | Esc or Tab | |
| Gamepad | Left stick or d-pad | A (bottom face button) | Start or Y | Standard layout via Ebitengine |
| Mouse | Click to walk | Click the thing | Click the backpack | |

Rules: eight-direction movement, four-direction facing. No run button (walk
speed is brisk). Dialogue advances with any tap or paw. Everything the player
can interact with shows a bubble when they are near. Touch targets are at least
64 CSS pixels (about 1 cm on a tablet).

## 4. Technical plan

### 4.1 Engine choice: Ebitengine compiled to WebAssembly

Ebitengine is the mature 2D game library for Go. One codebase gives you a
browser build (WASM) plus Windows, macOS and Linux, and later Android and iOS
through `ebitenmobile`. It covers rendering, sprites, input (keyboard, mouse,
touch, gamepad), audio (ogg, mp3, wav) and text. The current release is v2.10
(September 2026). It requires Go 1.25 or later, and since v2.10 desktop builds
on macOS and Linux are pure Go, so no C compiler is needed anywhere.

Alternatives considered:

| Option | Verdict |
|---|---|
| Go server for game logic plus a TypeScript client (Phaser or PixiJS) | Two languages, a network round-trip for every input, hosting cost. Only worth it for a multiplayer-first game. Not now. |
| Go plus a hand-written canvas renderer through `syscall/js` | Reinventing Ebitengine. No. |
| Godot (drop the Go requirement) | Best 2D tooling, but you would write GDScript or C#. Listed for honesty. Ebitengine is a good and fun path. |

Trade-offs accepted with Ebitengine and WASM: binary size (10 to 20 MB
uncompressed, 3 to 6 MB with Brotli), no DOM UI (all UI is drawn in-game,
which is what we want anyway), single-threaded in the browser, and browser
audio that only starts after a tap.

### 4.2 Architecture

Keep it simple: structs and small interfaces, no ECS framework. Logic packages
never import Ebitengine, so they can be unit tested.

```
critterkingdom/
  cmd/game/            main.go: window setup, scene stack, build tags for js vs desktop
  internal/
    engine/            game loop glue, scene stack, camera, input abstraction (Move vector, Paw, Menu, TapAt)
    assets/            embed.FS + loaders (images, Aseprite JSON, audio, fonts), caching
    world/             Tiled map loading, layers, collision grid, warps, triggers, y-sorting, A* pathfinding
    entity/            player, npc, critter, item, prop; animation controller
    dialogue/          data model + runner (conditions, choices, actions)   [pure Go]
    quest/             definitions, runtime state, event bus, step types    [pure Go]
    critter/           species data, spawn zones, befriend rules            [pure Go]
    inventory/         items, stickers, companion                           [pure Go]
    save/              SaveGame struct, versioning, Store interface (localStorage on js, file on desktop)
    i18n/              string tables sv/en, T(key)
    tts/               read-aloud: speechSynthesis on js, no-op elsewhere
    ui/                dialogue box, HUD, backpack, sticker book, virtual pad, 9-slice, buttons
    minigame/          snack time, follow the bounce
    audio/             music and sfx manager, volume, unlock
    content/           content lint: validates all data cross-references    [test]
  assets/              exported, embedded game data (see 6.6)
  assets-src/          Aseprite, Tiled and BeepBox sources (not embedded)
  web/                 index.html, wasm_exec.js, manifest.webmanifest, sw.js, loading screen
  tools/               asset export helpers
  docs/                PLAN.md (this), later ART-BIBLE.md, CREDITS.md, PLAYTEST-LOG.md
```

Flow: a `Scene` stack (Boot, Title, World) where Dialogue, Backpack and
Minigame are pushed as overlays that pause the world underneath. The World
scene owns the map, entities, camera and systems: movement and collision,
interaction (nearest interactable in facing direction, or the tapped one),
animation, simple AI (idle, wander, follow), triggers (warps, area enter) and
the spawner. Systems emit events (`ItemCollected`, `TalkedTo`, `EnteredArea`,
`Befriended`) and the quest system listens.

### 4.3 Key technical decisions

| Topic | Decision | Why |
|---|---|---|
| Resolution | The world is rendered to an offscreen image 180 px tall whose width follows the screen's aspect ratio (320 on 16:9, 240 on a 4:3 iPad), then drawn at the largest integer scale that fits. UI and text are drawn at full screen resolution. | Chunky pixel art plus crisp, large, readable text for early readers. No letterboxing on tablets. |
| Tiles | 16 by 16 grid. Player and villagers drawn in a 32 by 32 box (big heads), small critters 16 or 24 px. | Tiles are fast to produce; characters stay cute and readable. |
| Maps | Tiled, exported as JSON (`.tmj`), read by a small loader of our own (the format is plain JSON: layers, tile ids, objects with properties). `go-tiled` reads the TMX format if you would rather not write one. | Mature, free, terrain brushes, object layers for spawns, warps and triggers. LDtk is the alternative. |
| Content | Data-driven JSON: critters, quests, dialogue, strings, items. Adding a quest means adding a file, not code. | Fast iteration, testable, and later the kids' ideas go in as data. |
| Assets | `embed.FS` into the binary to start. Move large audio to fetched-at-runtime if the compressed WASM passes about 10 MB. | Simplest first; optimise when measured. |
| Save | A versioned JSON `SaveGame`; a `Store` interface with a localStorage implementation (js build tag) and a file implementation (desktop). Autosave. | Offline, no accounts. The version number allows migrations. |
| Languages | Swedish and English from day one. All text goes through `i18n.T(key)`. The font must include å, ä and ö. | Retrofitting i18n is painful; both are needed at home. |
| Read-aloud | Web Speech API (`speechSynthesis`) through `syscall/js`, with a settings toggle. | Free, instant, supports sv and en. Voice quality varies by device; recorded voices can replace it later. |
| Time | In-game time advances by sleeping, not by the real clock. Only the daily visitor uses the real date. | Deterministic, kid-controlled, testable. |
| Combat | A light "boop" with a stick, petals instead of hearts, a nap instead of death, "No grumpies" toggle. | Zelda feel without fear or frustration. |
| UI framework | Hand-rolled: 9-slice panels, big buttons. | The UI is small and unusual (icons, huge targets). `ebitenui` exists if it grows. |
| Dependencies | Ebitengine, possibly `go-tiled`, stdlib for JSON. | Keep the dependency list short. |

### 4.4 Browser specifics

- `web/index.html`: full-screen canvas, a "Tap to start" overlay (unlocks
  audio, requests fullscreen), `touch-action: none`, no pinch zoom, no text
  selection, a loading bar.
- PWA: `manifest.webmanifest` plus a service worker, so the game can be added
  to the tablet home screen and works offline.
- Portrait on a phone: show a "turn your device" picture (iOS cannot lock
  orientation from a web page).
- Serve `.wasm` as `application/wasm` with Brotli or gzip. Cloudflare Pages,
  Netlify and GitHub Pages all work as free static hosts; verify compression.
- Build with `-trimpath -ldflags="-s -w"`; optionally try `wasm-opt -Oz` for a
  further size cut.

### 4.5 Developer loop

- Daily work on desktop: `go run ./cmd/game` (fast compile, debugger works). A
  `debug` build tag adds an overlay with FPS, position and collision boxes,
  plus a `-map` flag to jump straight to a map.
- Browser check: `go run github.com/hajimehoshi/wasmserve@latest ./cmd/game`
  and open http://localhost:8080. Test on a real tablet over the LAN from
  milestone 0 onwards.
- CI (GitHub Actions): `go vet`, `go test ./...`, build desktop and wasm,
  deploy `web/` on main.

### 4.6 Testing

Follow the house rules: real implementations, test business rules, no tests
for rendering or wiring.

- Unit tests in the pure packages: quest state machine (every step type,
  counters, rewards), dialogue runner (conditions, choices, actions),
  inventory, Snack Time rules, save round-trip and version migration, movement
  and collision maths, spawner weights with a seeded RNG, i18n fallback.
- **Content lint test**: loads every real data file and asserts that every
  referenced NPC, item, critter, map, sprite and string key exists in both
  languages, every map has a spawn point, and every warp resolves. This test
  is what catches a typo in a quest file.
- Determinism: inject a clock and a seeded `rand.Rand`; no sleeps.

## 5. Roadmap: the steps

Estimates assume evenings and weekends, with placeholder art until the art
work in section 6 kicks in. Each milestone ends with a kid playtest.

### M0 Foundation: "Hello Critter" (1 to 2 weeks)

- [ ] `go mod init`, the folder layout above, a Makefile (`run`, `web`, `test`, `art`, `check`), GitHub Actions
- [ ] Ebitengine window: a placeholder critter moves with keyboard, touch (tap-to-walk and pad), gamepad and mouse, through the input abstraction in `engine`
- [ ] Offscreen world render with integer scaling and full-resolution text; adaptive logical width
- [ ] The same build served by wasmserve and tested on an iPad or Android tablet
- [ ] `web/index.html` with tap-to-start; deployed to a URL from CI
- [ ] Debug overlay behind a build tag

**Done when** a child moves a square around on a tablet from a link.

### M1 A place to walk: "Acorn Hollow" (2 to 3 weeks)

- [ ] Tiled JSON loading, layers, collision grid, animated tiles (water)
- [ ] Camera follow and clamp; y-sorted props; "above" layer
- [ ] Warps: village to meadow, village to house interior
- [ ] Player sprite with four-direction idle and walk (a placeholder pack is fine)
- [ ] A* tap-to-walk, joystick alternative
- [ ] Audio manager: one music loop, footsteps, unlock on first tap

**Done when** you can walk around the village, enter a house and reach the meadow.

### M2 Talking: "Hello, neighbour" (2 weeks)

- [ ] NPC entities: idle and wander, face the player, "…" bubble
- [ ] Dialogue box: portrait, name, typewriter text, at most two big choice buttons
- [ ] Dialogue data format, runner and tests
- [ ] i18n with sv and en string tables; a font with å, ä and ö
- [ ] Read-aloud via speechSynthesis with a toggle

**Done when** three villagers chat in two languages and are read aloud.

### M3 Quests: "Can you help me?" (2 to 3 weeks)

- [ ] Quest definitions, runtime state, event bus, tests for every step type
- [ ] World items (berries, hat), pickup, inventory with icons
- [ ] HUD tracker with icon and counter; helper arrow; quest page in the backpack
- [ ] Rewards and completion fanfare
- [ ] Save and load with three slots, autosave, localStorage and file stores

**Done when** "The Missing Hat" and "Berry Pie" play start to finish and survive a page reload.

### M4 Wild critters: "New friends" (2 to 3 weeks)

- [ ] Spawn zones with weights, conditions (night, flag, companion), respawn timers, daily visitor
- [ ] Critter AI: idle, hop, notice the player
- [ ] Snack Time mini-game
- [ ] Sticker Book UI; companion follows (trail) and can be petted
- [ ] Five critters with sheets and stickers; the "Scared Little Mouse" quest
- [ ] Sleep in bed to change time of day; night tint

**Done when** all five critters can be befriended, show up in the book, and one follows you.

### M5 Exploring and puzzles: "The Glow Cave" (3 to 4 weeks)

- [ ] Gating items: Snippers (brambles), Splash Boots (shallow water)
- [ ] Puzzle pieces: push block, flower switch, mushrooms in order, key and door, pots
- [ ] Darkness with a light radius (firefly companion or lantern) using a light-mask image
- [ ] Grumpy critters: patrol, bump, boop, petals, nap; "No grumpies" toggle
- [ ] Glow Cave map (five rooms), Glow Seed, main quest skeleton; "Fix the Bridge" opens the Woods

**Done when** a 6-year-old finishes the cave with at most one hint from an adult.

### M6 Kid-proofing and polish: "Ready for the sofa" (2 to 3 weeks)

- [ ] Title: three slots, species and colour picker, preset names, language flags
- [ ] "Little Explorer" (5 to 6) and "Explorer" (7 to 9) presets: arrows, hints, grumpies, pad or tap
- [ ] Settings: sound, read-aloud, controls
- [ ] Juice: sparkles, hearts, confetti, soft screen transitions; no harsh flashes
- [ ] Loading screen with progress; PWA manifest and offline; orientation hint
- [ ] Performance and size pass on a tablet (60 fps, under about 10 MB over the wire)
- [ ] Three playtests with target-age kids; fix the top five confusions
- [ ] `CREDITS.md`, licences checked

**Done when** a parent can send a link and a 5-year-old plays for 10 minutes unassisted.

### M7 Content expansion (ongoing)

Whispering Woods and Pebble Beach, ten more critters, six more quests, two more
mini-dungeons, the full "Sleepy Tree" main quest, hats and house decoration,
the Follow the Bounce mini-game.

### Stretch goals (pick later)

Local co-op (a second player steers the companion with a gamepad),
draw-your-own-critter pixel editor, a tiny garden (plant a seed, berries after
you sleep), a birthday party event on a date set by a parent, photo mode
(saves a PNG), desktop and mobile store builds, optional cloud save through a
small Go server using a six-letter family code (no accounts).

Rough total to a playable vertical slice (M0 to M6): 4 to 5 months part-time,
plus the art time in 6.7.

## 6. Making the assets

### 6.1 Art direction on one page (write it down as `docs/ART-BIBLE.md` before drawing)

- **Mood words**: sunny, soft, round, friendly, storybook.
- **Palette**: pick one from Lospec and stick to it. Recommended: *Endesga 32*
  (bright, great for cute games) or *Resurrect 64* (more range). Import it
  into Aseprite and never eyedrop colours from elsewhere.
- **Sizes**: tiles 16 by 16; player and villagers in a 32 by 32 box (feet at
  bottom centre, collision box about 12 by 10 at the feet); small critters 16
  or 24 px; portraits 64 by 64 with three expressions; stickers 48 by 48 with
  a white "sticker" border; icons 16 px (HUD) and 32 px (menus); 9-slice panel
  48 by 48 with 16 px borders.
- **Shapes**: big heads (about half the height), big eyes with a highlight,
  rounded everything, no sharp angles.
- **Lines and shading**: a 1 px outline in a darker version of the fill colour
  (not black); two or three shades per material; light from top-left; no
  dithering, so kids get a clean read.
- **Contrast rule**: ground tiles are lighter and less saturated than
  characters, so characters pop.
- **Animation**: idle 2 to 4 frames at 4 to 6 fps (a small bob or squash);
  walk 4 frames at 8 to 10 fps; hop 3 frames for small critters. Draw down, up
  and side; the engine mirrors the side view. Shared emote sprites (heart,
  exclamation, zzz, sweat, music note) that float above anyone.
- **References**: Kitaria Fables, Animal Crossing, Stardew Valley, Link's
  Awakening (Switch), Tunic (for colour and softness). Collect screenshots in
  a PureRef board.

### 6.2 Tools

| Need | Recommended | Free alternative |
|---|---|---|
| Pixel art and animation | Aseprite (about 20 USD; can also be built from source) | Libresprite, Pixelorama, Piskel (browser) |
| Maps | Tiled | LDtk |
| Sound effects | jsfxr or ChipTone (browser, instant) | Bfxr |
| Music | Bosca Ceoil Blue or BeepBox (browser) | LMMS, FamiStudio |
| Audio editing and conversion | Audacity, ffmpeg | |
| Fonts | Google Fonts: Fredoka, Baloo 2 or Nunito (rounded, readable, OFL licence, include å ä ö) | Pixelify Sans (pixel look), m5x7 |
| Palettes | Lospec | |
| Placeholder art while you build | Kenney.nl (CC0), Ninja Adventure pack (CC0), Sprout Lands (check licence) | OpenGameArt (check each licence) |
| Reference boards | PureRef | |

AI image tools are fine for mood boards and quick concept drafts. Final sprites
still need hand clean-up for palette and consistency, and each tool's licence
terms need checking.

### 6.3 Step by step: characters

1. Make the **style sheet first**: one grass tile, one tree, one critter, one
   UI panel and one portrait, placed on a 320 by 180 mock screen and viewed at
   4x. Iterate until it feels cute. Show it to the kids and ask which one they
   like.
2. New 32 by 32 file in Aseprite with the palette loaded. Block the
   silhouette: head circle about 16 px, body about 10 px, feet on row 30.
3. Flat colours, then shading, then outline, then eyes and highlight.
4. Duplicate for the up and side views. The side view is mirrored by the
   engine (`flipX`).
5. Animate with tags: `idle_down`, `walk_down`, `idle_up`, and so on. Use
   onion skin. Idle is a 1 px squash; walk is a bounce with alternating feet.
6. Export a sheet plus JSON:
   ```
   aseprite -b fox.aseprite --sheet fox.png --data fox.json --format json-array --list-tags --sheet-type packed
   ```
   The Go loader reads `frames` and `frameTags`, so there are no hand-written
   frame tables.
7. Portrait: a 64 by 64 file, same palette, tags `neutral`, `happy`, `sad`.
8. Cheap variants: palette swaps (bunny to brown bunny), or a shared body with
   swappable head, ear and tail layers.

Frame budget per playable critter: 3 directions times (2 idle + 4 walk), plus
boop, sleep and happy, about 23 frames. Villagers: 18 (or 6 if they stand
still). Small critters: 5 to 10.

### 6.4 Step by step: tiles, props and maps

1. Base terrain in Aseprite's **Tiled Mode** (View, Tiled Mode) so edges wrap
   seamlessly: grass plus two or three variants with tiny flowers, dirt path,
   sand, water (4-frame animation), cave floor, cliff edges.
2. Transitions: a 16-tile corner set per pair (grass to dirt, grass to water,
   sand to water). In Tiled, define a **Terrain Set** so you paint transitions
   with one brush.
3. Props as multi-tile images: tree (32 by 48), bush, berry bush (full and
   empty), rock, fence set, sign, house (64 by 64) with door, chest (closed and
   open), pot, push block, flower switch (off and on), mushroom (off and on),
   bramble, plank bridge.
4. Interior set: floor, walls, window, bed, table, rug, shelf.
5. Tilesets in Tiled: `tileset_outdoor.png`, `tileset_indoor.png` and
   `tileset_cave.png` as regular tilesets; props as an *image collection*
   tileset so each prop keeps its own size.
6. Map layout in Tiled (orthogonal, 16 by 16, about 40 by 30). Layers:
   `ground`, `decor`, `props` (object layer with tile objects, y-sorted),
   `above` (tree tops, roofs), `collision` (a one-tile "solid" tileset, or
   rectangles), `objects` (`player_spawn`, `npc` with `id`, `warp` with
   `target_map` and `target_spawn`, `spawn_zone` with `habitat`, `item` with
   `item_id`, `trigger` with `event`, puzzle objects). Map properties:
   `music`, `biome`, `dark`.
7. Turn on "Repeat last export on save" in Tiled so the `.tmj` in
   `assets/maps/` is always current. Test with `go run ./cmd/game -map meadow`.

### 6.5 Step by step: UI, fonts, audio and voice

**UI**

- 9-slice panel (48 by 48, 16 px borders), rounded with a soft shadow; a
  button in normal and pressed states; a "next" arrow; a checkmark.
- Icon set at 16 and 32 px: speech, footsteps, berry, hat, heart, star,
  sticker, backpack, gear, speaker, flags, snacks, key, lantern, zzz.
- Virtual joystick (ring plus knob) and a big **paw-print** action button.
- Text at full resolution using Fredoka or Baloo 2 through Ebitengine's
  `text/v2`. Dialogue text about 1 cm tall on a tablet, at most two lines of
  about 40 characters.

**Sound effects** (jsfxr or ChipTone to WAV, then
`ffmpeg -i x.wav -c:a libvorbis -q:a 4 x.ogg`): step, talk blip, pickup, door,
chest, quest start (three notes up), quest done (fanfare), befriend success
(sparkle), wrong snack (a soft "bwomp", never harsh), menu tap, boop, hearts,
sparkle, zzz. Keep everything soft and cap the master volume.

**Music** (BeepBox or Bosca Ceoil to WAV, then OGG at `-q:a 3`): three loops
of 45 to 60 seconds for village, meadow and cave; 90 to 120 BPM, major or
pentatonic, soft instruments. Music is the biggest asset, so keep loops short.

**Critter voices**: record about 20 syllables ("ba", "mi", "ko", ...) in a
high and a low voice and play one per couple of letters while the text types
out (the Animal Crossing trick). Cheap and charming. Read-aloud of the actual
words comes from the browser's speech voices; later, family-recorded lines can
replace them.

**Fonts** must include Latin Extended for å, ä and ö. Fredoka, Baloo 2, Nunito
and Pixelify Sans do, all under the SIL Open Font License (keep the licence
text in `CREDITS.md`).

### 6.6 Pipeline and folders

```
assets-src/   .aseprite, .tmx/.tsx, BeepBox .json, raw .wav   (sources of truth; consider git LFS)
assets/       exported and embedded
  sprites/    critters/fox.png + fox.json, npcs/, props/, ui/, emotes/
  tiles/      tileset_outdoor.png ...
  maps/       village.tmj, meadow.tmj, ...
  audio/      music/*.ogg, sfx/*.ogg|wav, voices/
  fonts/      Fredoka-Medium.ttf
  data/       critters.json, items.json, quests/*.json, dialogue/*.json, strings/sv.json, strings/en.json
```

- `make art` runs Aseprite's CLI export on changed sources; `make audio` runs
  ffmpeg; `make check` runs the content lint test; `make web` builds the wasm
  bundle into `web/`.
- Naming: lower_snake_case, one sheet per character, tags named
  `<anim>_<dir>`.
- Licences: prefer CC0; CC-BY needs an entry in `CREDITS.md`; avoid NC and ND
  material in case the game is ever sold. Decide a licence for your own art
  too (the repo is MIT, which covers code; art can be the same or CC-BY).

### 6.7 Asset list for the vertical slice (rough first-timer hours)

| Asset | Count | Hours |
|---|---|---|
| Terrain tiles and transitions | about 60 | 12 to 15 |
| Props and interiors | about 30 | 10 |
| Playable critters (fox, bunny, hedgehog via swaps) | 3 | 8 to 10 |
| Villagers | 5 | 12 to 15 |
| Wild critters | 5 | 8 to 10 |
| Portraits (three expressions each) | 6 | 6 |
| UI, icons, pad, logo | about 40 pieces | 8 |
| Sound effects | about 15 | 3 |
| Music loops | 3 | 6 |
| **Total** | | **about 75 to 85** |

Strategy for a non-artist: build the whole game on CC0 placeholders, replace
one category at a time, keep frame counts low, reuse a base body with palette
swaps. Buying an itch.io pack (5 to 30 EUR) or commissioning a few key sprites
is a legitimate shortcut.

## 7. Suggestions beyond the brief

1. **Read-aloud plus icons** is the single biggest accessibility win for 5 and
   6-year-olds. Build it in M2, not later.
2. **Tap-to-walk as the default touch control** turns the game into
   point-and-click for the youngest.
3. **Two languages from day one**; never hard-code a string.
4. **Low-res world, full-res text**: crisp, readable text without giving up
   the pixel look.
5. **Age presets** ("Little Explorer" and "Explorer") instead of a difficulty
   menu.
6. **PWA**: an icon on the tablet home screen that works offline, including in
   the car.
7. **Draw-your-own-critter**: a 16 by 16 in-game pixel editor with the game
   palette. Cheap in Ebitengine, huge for kids.
8. **Birthday party event**: a parent sets a date in settings and the village
   celebrates that day.
9. **Photo mode**: one button saves a screenshot for showing grandparents.
10. **Skip real-world GPS**: browsers have no step counter, and outdoor play
    with a tablet is a poor fit for ages 5 to 9. Take the collecting loop,
    leave the walking.
11. **Content as data**, so adding a quest is adding a JSON file, and the
    content lint test catches mistakes.
12. **Playtest protocol**: 10-minute sessions, sit behind, say nothing, write
    down every place they look lost or ask "what now?" in
    `docs/PLAYTEST-LOG.md`. Ask "what would you do next?" rather than
    explaining.
13. Later, a **tiny Go server** for family cloud saves via a short code (no
    accounts, no email). The save types are the same Go structs on both sides.

## 8. Risks

| Risk | Mitigation |
|---|---|
| WASM size and first load on tablets | Brotli hosting, short music loops, a size budget checked in CI, a loading screen with progress |
| Browser audio blocked until a gesture | "Tap to start" overlay; the audio manager unlocks on first input |
| iOS Safari quirks (memory, touch, speech, no orientation lock) | Test on an iPad from M0; orientation hint; keep memory under about 300 MB |
| Art is the long pole | Placeholder packs first; art bible; low frame counts; palette swaps |
| Scope creep | Vertical slice M0 to M6 before any content expansion; the stretch list stays a list |
| Save data lost when browser storage is cleared | Autosave often; export and import "magic word"; later an optional cloud save |
| Text too hard for 5-year-olds | Icons everywhere, read-aloud, max two lines, playtests |
| Go GC pauses in WASM | Avoid per-frame allocations, reuse `DrawImageOptions`, draw only visible tiles |

## 9. Decision log

| Date | Decision | Alternatives | Reason |
|---|---|---|---|
| 2026-11-27 | Ebitengine v2 compiled to WebAssembly | Go server plus TS client; Godot | One language, one codebase for web and desktop |
| 2026-11-27 | Tiled JSON maps with our own loader | LDtk; go-tiled with TMX | Mature editor, terrain brushes, object layers; the JSON is simple |
| 2026-11-27 | 16 px tiles, 32 px characters, 180 px logical height, adaptive width | 32 px tiles; fixed 16:9 | Cheap tiles, cute characters, no letterboxing on iPad |
| 2026-11-27 | UI and text at full resolution | Pixel fonts at world resolution | Readability for early readers |
| 2026-11-27 | Tap-to-walk as the default on touch | Virtual pad only | Easiest for 5-year-olds |
| 2026-11-27 | Day and night change by sleeping | Real-time clock | Deterministic and kid-controlled |
| 2026-11-27 | Boop, petals and nap; no game over | Classic hearts and game over | Pillar 1 |
| 2026-11-27 | localStorage and file saves, no accounts | Backend with logins | Pillar 6 |
| 2026-11-27 | sv and en from day one | English only | Needed at home; retrofitting is painful |

## Appendix A: data format sketches

A critter:

```json
{
  "id": "firefly",
  "name": {"en": "Firefly", "sv": "Eldfluga"},
  "habitat": "pond",
  "activeAt": ["night"],
  "likes": "berry",
  "rarity": "uncommon",
  "sprite": "critters/firefly",
  "ability": "light",
  "fact": {"en": "Fireflies glow to say hello.", "sv": "Eldflugor lyser för att säga hej."}
}
```

A quest:

```json
{
  "id": "missing_hat",
  "giver": "grandma_hedgehog",
  "title": {"en": "The Missing Hat", "sv": "Den försvunna hatten"},
  "steps": [
    {"type": "goto",    "target": "meadow",                   "icon": "footsteps"},
    {"type": "collect", "item": "grandmas_hat", "count": 1,   "icon": "hat"},
    {"type": "talk",    "npc": "grandma_hedgehog",            "icon": "speech"}
  ],
  "reward": {"sticker": "hat_hero", "item": "cookie"}
}
```

A dialogue file:

```json
{
  "npc": "grandma_hedgehog",
  "nodes": {
    "start": {
      "if": [{"quest": "missing_hat", "is": "not_started"}],
      "lines": [
        {"text": "grandma.hat.1", "face": "sad"},
        {"text": "grandma.hat.2", "face": "neutral"}
      ],
      "choices": [
        {"text": "ui.yes",   "icon": "check", "do": [{"startQuest": "missing_hat"}], "goto": "thanks"},
        {"text": "ui.later", "icon": "cross", "goto": "end"}
      ]
    }
  }
}
```

A string table (`strings/sv.json`):

```json
{"grandma.hat.1": "Åh nej! Min hatt blåste bort.", "grandma.hat.2": "Kan du titta på ängen?", "ui.yes": "Ja!", "ui.later": "Senare"}
```

## Appendix B: commands

```bash
# project
go mod init github.com/<you>/critterkingdom
go get github.com/hajimehoshi/ebiten/v2@latest          # needs Go 1.25 or later

# run on desktop
go run ./cmd/game

# run in the browser (rebuilds on every reload)
go run github.com/hajimehoshi/wasmserve@latest ./cmd/game   # then open http://localhost:8080

# build the web bundle
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o web/game.wasm ./cmd/game
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/            # Go 1.23 and older: misc/wasm/wasm_exec.js

# assets
aseprite -b assets-src/fox.aseprite --sheet assets/sprites/critters/fox.png \
  --data assets/sprites/critters/fox.json --format json-array --list-tags --sheet-type packed
ffmpeg -i in.wav -c:a libvorbis -q:a 4 out.ogg

# tests
go test ./...
```

## Appendix C: links

- Ebitengine: https://ebitengine.org/ and the WebAssembly guide https://ebitengine.org/en/documents/webassembly.html
- Ebitengine 2.10 release notes: https://ebitengine.org/en/documents/2.10.html
- wasmserve: https://github.com/hajimehoshi/wasmserve
- Ebitengine text/v2: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/text/v2
- ebitenui (if the UI grows): https://github.com/ebitenui/ebitenui
- go-tiled (TMX loader): https://github.com/lafriks/go-tiled
- Tiled: https://www.mapeditor.org/ and terrain sets https://doc.mapeditor.org/en/stable/manual/terrain/
- LDtk: https://ldtk.io/
- Aseprite: https://www.aseprite.org/ and its CLI https://www.aseprite.org/docs/cli/
- Lospec palettes: https://lospec.com/palette-list
- Kenney (CC0 assets): https://kenney.nl/assets
- Ninja Adventure pack (CC0): https://pixel-boy.itch.io/ninja-adventure-asset-pack
- Sprout Lands: https://cupnooble.itch.io/sprout-lands-asset-pack
- jsfxr: https://sfxr.me/ and ChipTone: https://sfbgames.itch.io/chiptone
- BeepBox: https://www.beepbox.co/ and Bosca Ceoil Blue: https://github.com/YuriSizov/boscaceoil-blue
- Google Fonts: https://fonts.google.com/
- Web Speech API: https://developer.mozilla.org/en-US/docs/Web/API/SpeechSynthesis
