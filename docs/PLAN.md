# Critter Kingdom: Project Plan

Status: draft v0.4, 2026-11-27. Changes since v0.3: the art is smooth instead
of pixel art, after comparing the two with the same placeholder art. Sections
0, 4, 5, 6, 7 and 8 follow from that, and section 9 says why. v0.3 marked M0
as implemented and made sections 4.2 to 4.5, 6.6 and Appendix B describe the
code as built; the tablet playtest and the first deploy are still open. v0.2
added real-time fights (section 2.7) and milestone M5.

How to use it: each milestone in section 5 is meant to become its own spec and
branch. Tick the boxes as work lands. When a decision changes, update the
decision log in section 9 instead of rewriting history.

## 0. In one paragraph

Critter Kingdom is a cozy 2D top-down adventure for children aged 5 to 9. You
play a small critter (fox, bunny or hedgehog) in a kingdom of talking animals.
You explore a storybook world, chat with villagers, take on small quests,
befriend wild critters and collect them in a sticker book, bonk grumpy critters
with your stick, and solve gentle puzzles in tiny dungeons. It is written in Go
with Ebitengine, compiled to WebAssembly and played in the browser on a tablet
or laptop. No accounts, no ads, no purchases, no chat with strangers.

## 1. Vision

### 1.1 What we borrow

| Inspiration | What we take | How it becomes kid-sized |
|---|---|---|
| Zelda (Link's Awakening, A Link to the Past) | Top-down exploration, items that open new areas, small dungeons with puzzles, real-time stick fights and bosses, villagers to talk to | One puzzle idea per room. Every enemy telegraphs its attack. No game over: at zero petals you nap and wake at the door. |
| Pokémon Go | Creatures hidden in the world, finding them in their habitats, a collection to fill, rarity, "only comes out at night" | No GPS. Critters live in the game world. Befriend by picking the right snack, not by throwing balls. The collection is a sticker book. |
| Kitaria Fables | Cast of cute animal villagers, village-and-quests loop, cozy tone, action combat with cute animals, co-op | Quests take 3 to 8 minutes and are shown with pictures. Fights are short and forgiving. Co-op is a later stretch goal. |

### 1.2 Design pillars for ages 5 to 9

1. **Never stuck, never punished.** No game over, no timers, nothing is ever lost. A stuck player always has a visible hint.
2. **Readable without reading.** Every quest step has an icon. Every speaker has a portrait with an expression. Every line can be read aloud. Text is short (max two lines) and large.
3. **One stick, one button.** Move and "paw" (interact or swing). Menus are big picture buttons. Nothing needs double-tap, long-press or button combos.
4. **Brave, not brutal.** You can fight grumpy critters with your stick, and they poof into flowers and run off smiling. Nobody bleeds, nobody dies, and helping and befriending are still how you win.
5. **Small loops, frequent smiles.** Something rewarding happens every minute. A quest fits in one sitting.
6. **Safe by design.** Offline-first, no accounts, no network needed, saves live on the device.
7. **Playtest with real kids** at the end of every milestone. Watch, don't explain.

### 1.3 Deliberately out of scope for v1

Real-world GPS play, online multiplayer, chat, accounts, monetisation, farming
and crafting systems, inventory-management puzzles, twitch combat (parries,
combos, dodge rolls), 3D.

## 2. The game

### 2.1 World

A small kingdom seen from above. The hub village **Acorn Hollow** has five or
six villagers and your own little house. Around it, biomes unlock over time:

| Biome | Unlocked by | Critters | Hook |
|---|---|---|---|
| Sunny Meadow | start | bird, snail, squirrel | tutorial quests, berry bushes, first slimes at the edges |
| Lily Pond | start | frog, firefly (night only) | first companion with an ability |
| Whispering Woods | "Fix the Bridge" quest | owl, fox kit, mushroom sprite | brambles need the Snippers, thistle beasts |
| Glow Cave | firefly companion or lantern | glow-worm, friendly bat | first mini-dungeon, five rooms, first boss |
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
- Your stick. Later a Sturdy Stick and a Sparkle Wand. The newest one is
  always used, so there is no weapon menu.
- Cosmetic rewards: hats, scarves, stickers for your house wall.
- A **companion**: one befriended critter follows you. You can pet it
  (hearts). A few companions have an ability used by puzzles and quests: the
  firefly lights dark places, the bird points at hidden things, the frog
  fetches from shallow water. Companions also help in fights (see 2.7).

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
(companion gives hot/cold hints), **solve puzzle room**, **defeat N
grumpies**, **clear a room**.

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
| Slimes in the Berry Patch | Baker Badger | first fights: defeat three slimes at the meadow's edge |
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
- **Glow Cave** is the first mini-dungeon: five rooms, dark, two of them locked
  until the grumpies inside are cleared, a sunflower checkpoint, and the
  Grumbling Grub boss at the end. Reward: a Glow Seed, the Sturdy Stick and a
  hat.

### 2.7 Fights

Real-time action in the overworld, in the spirit of Zelda and Kitaria Fables,
tuned for small hands.

**How a fight works**

- The paw button swings your stick in the direction you face. Interacting has
  priority: if a villager, item or sign is in front of you, paw talks or picks
  up instead. Villagers can never be hit.
- Aim assist: when an enemy is within a short cone in front of you, the swing
  turns toward it. On touch, tapping an enemy walks you up to it and swings
  until it is gone.
- Enemies always telegraph. Every attack starts with a visible wind-up of at
  least 0.6 seconds (shaking, puffing up, a flash) before anything can hurt
  you. Projectiles are slow and big.
- Hitboxes are generous for you and forgiving for them: your swing is wider
  than the stick looks, and an enemy's hurt-box is smaller than its sprite.
- Taking a hit: lose one petal, a short knockback, one second of blinking
  invulnerability, a soft "oof". No red flashes and no screen shake by
  default.
- Defeating an enemy: it poofs into flowers and hearts, sometimes drops a
  snack, and its grumpy face turns happy as it runs off. Nobody dies, nothing
  is scary.
- Nap rule: at zero petals you fall asleep and wake at the room or map
  entrance with full petals. Enemies in that room reset. Nothing is lost.

**Petals** are your health: three to start, up to five through Big Petal
rewards. Hearts dropped by enemies and pots restore one. Outside fights petals
regrow slowly (about one every 20 seconds), so nobody gets stuck weak.

**Enemies for v1** (grumpy, never monstrous; each has exactly one trick):

| Enemy | Where | Trick | Hits to defeat |
|---|---|---|---|
| Sulky Slime | meadow edges, cave | hops toward you slowly | 1 |
| Thistle Beast | woods, cave | shakes, then charges in a straight line; step sideways | 2 |
| Grumpy Mushroom | cave | stands still, puffs a slow spore cloud in four directions; hit it between puffs | 1 |
| Buzzy Bee (M8) | meadow, woods | flies in a lazy wave | 1 |
| Rolling Pinecone (M8) | woods | rolls at you, then stops dizzy; hit it while dizzy | 2 |

**Boss: the Grumbling Grub**, at the end of Glow Cave. Three telegraphed
phases (roll across the room, spit slow glow-blobs, sulk when tired). It only
takes hits while sulking. Its health is a row of five big stars, not a number.
When beaten, it turns out it was only hungry: you give it a berry, it becomes a
friend and a sticker.

**Weapons**: the Stick (start), then the Sturdy Stick (Glow Cave reward, two
damage), then the Sparkle Wand (M8, shoots a slow sparkle for enemies that are
hard to reach). The newest one is always used.

**Companion in fights**: your companion follows a step behind, cannot be hurt,
and helps on its own: it nips at nearby enemies for small damage every few
seconds, and its ability still applies (the firefly lights the room). It
cheers when you win.

**Where fights happen**: never inside the village. A few slimes at the
meadow's edges, more in the woods, and rooms in the cave. Some cave rooms lock
their doors until cleared (classic Zelda), always with a sunflower checkpoint
before the boss.

**Age presets**: Little Explorer gives enemies one hit point, one petal of
damage, slower speed and fewer per room. Explorer uses the values above.
Peaceful mode puts every enemy to sleep for kids who just want to explore.

**Data and code**: enemy stats, speed, telegraph time, drops and which
behaviour they use live in `enemies.json`. The behaviours themselves (hop
toward, charge, stationary puff, roll, fly in a wave) and each boss are code.

### 2.8 Session shape and saving

- Autosave at every quest step, map change and new sticker. The title screen
  has a huge "Continue" button.
- Natural stopping points every 3 to 8 minutes (quest done, sticker earned).
- Saves live in the browser (localStorage) and as a file on desktop builds.
  Later: export and import a save as a short "magic word" so it can move
  between devices.

## 3. Controls

| Input | Move | Paw (interact or swing) | Backpack | Notes |
|---|---|---|---|---|
| Touch (default) | **Tap where to walk** (pathfinding). Tap a critter, villager or item to walk there and interact; tap an enemy to walk there and swing. Optional floating joystick. | Big paw-print button bottom-right (or just tap the thing) | Big backpack button top-right | Point-and-click is the easiest thing a 5-year-old can do |
| Keyboard | Arrows or WASD | Space, Enter or E | Esc or Tab | |
| Gamepad | Left stick or d-pad | A (bottom face button) | Start or Y | Standard layout via Ebitengine |
| Mouse | Click to walk | Click the thing | Click the backpack | |

Rules: eight-direction movement, four-direction facing. No run button (walk
speed is brisk). The paw button interacts when something is in front of you
and swings the stick otherwise; villagers can never be hit. There is no dodge
or block button; enemies telegraph instead. Dialogue advances with any tap or
paw. Everything the player can interact with shows a bubble when they are
near. Touch targets are at least 64 CSS pixels (about 1 cm on a tablet).

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
    engine/            game loop glue, scene stack, camera, input abstraction (Move vector, Paw, Menu, Pointing),
                       touch controls (paw button, pad) until ui/ exists
    assets/            embed.FS + loaders (images, part layouts, audio, fonts), caching
    world/             Tiled map loading, layers, collision grid, warps, triggers, y-sorting, A* pathfinding
    entity/            player, npc, critter, enemy, item, prop; animation controller
    combat/            hitboxes, petals and damage rules, invulnerability, knockback, aim assist,
                       enemy behaviour state machines, boss scripts                 [pure Go]
    dialogue/          data model + runner (conditions, choices, actions)   [pure Go]
    quest/             definitions, runtime state, event bus, step types    [pure Go]
    critter/           species data, spawn zones, befriend rules            [pure Go]
    inventory/         items, stickers, companion, weapons                  [pure Go]
    save/              SaveGame struct, versioning, Store interface (localStorage on js, file on desktop)
    i18n/              string tables sv/en, T(key)
    tts/               read-aloud: speechSynthesis on js, no-op elsewhere
    ui/                dialogue box, HUD, backpack, sticker book, virtual pad and paw button (from engine/), 9-slice, buttons
    minigame/          snack time, follow the bounce
    audio/             music and sfx manager, volume, unlock
    content/           content lint: validates all data cross-references    [test]
  assets/              exported, embedded game data (see 6.6)
  assets-src/          Inkscape, Tiled and BeepBox sources (not embedded)
  web/                 index.html and loader.js (loading bar, tap-to-start), manifest.webmanifest, sw.js, icons/
  tools/               asset export helpers
  docs/                PLAN.md (this), later ART-BIBLE.md, CREDITS.md, PLAYTEST-LOG.md
```

Flow: a `Scene` stack (Boot, Title, World) where Dialogue, Backpack and
Minigame are pushed as overlays that pause the world underneath. The World
scene owns the map, entities, camera and systems: movement and collision,
interaction (nearest interactable in facing direction, or the tapped one),
combat (swings, hurt-boxes, enemy behaviours), animation, simple AI (idle,
wander, follow), triggers (warps, area enter, room doors) and the spawner.
Systems emit events (`ItemCollected`, `TalkedTo`, `EnteredArea`,
`Befriended`, `EnemyDefeated`, `RoomCleared`) and the quest system listens.

### 4.3 Key technical decisions

| Topic | Decision | Why |
|---|---|---|
| Resolution | The world is measured in world pixels and drawn straight onto the screen at full resolution, with the largest whole-number scale that still shows at least 240 by 180 world pixels. At that scale the world fills the screen, up to 480 by 240 world pixels: 320 by 180 on 16:9, 256 by 192 on a 2048 by 1536 iPad. Only screens beyond that cap get a border. UI and text are drawn at full screen resolution too. | Smooth, soft art plus crisp, large, readable text for early readers. The whole-number scale puts tile edges on whole screen pixels, so tiles meet without seams. No letterboxing on tablets, and no very wide or very tall screen sees too much of a map. |
| Tiles | 16 by 16 world pixel grid. Player and villagers drawn in a 32 by 32 box (big heads), small critters 16 or 24. | Tiles are fast to produce; characters stay cute and readable. Sizes are in world pixels, so maps, collision and the camera do not depend on how detailed the art is. |
| Art | Smooth shapes at full screen resolution, never scaled-up pixels. Placeholders are drawn in code. Real art is drawn in Inkscape and exported to PNG at 8 px per world pixel, which is 1:1 at scale 8 (a 2048 by 1536 iPad) and a little softer at other scales. Characters are sets of parts that the engine moves, turns and squashes (section 6.3). | Chosen over pixel art after comparing the two. Animating parts in code avoids drawing every frame, which is the slow part of smooth art. |
| Maps | Tiled, exported as JSON (`.tmj`), read by a small loader of our own (the format is plain JSON: layers, tile ids, objects with properties). `go-tiled` reads the TMX format if you would rather not write one. | Mature, free, terrain brushes, object layers for spawns, warps and triggers. LDtk is the alternative. |
| Content | Data-driven JSON: critters, enemies, quests, dialogue, strings, items. Adding a quest means adding a file, not code. | Fast iteration, testable, and later the kids' ideas go in as data. |
| Assets | `embed.FS` into the binary to start. Move large audio, then art, to fetched-at-runtime if the compressed WASM passes about 10 MB. | Simplest first; optimise when measured. Smooth art at 8 px per world pixel takes more space than pixel art would. |
| Save | A versioned JSON `SaveGame`; a `Store` interface with a localStorage implementation (js build tag) and a file implementation (desktop). Autosave. | Offline, no accounts. The version number allows migrations. |
| Languages | Swedish and English from day one. All text goes through `i18n.T(key)`. The font must include å, ä and ö. | Retrofitting i18n is painful; both are needed at home. |
| Read-aloud | Web Speech API (`speechSynthesis`) through `syscall/js`, with a settings toggle. | Free, instant, supports sv and en. Voice quality varies by device; recorded voices can replace it later. |
| Time | In-game time advances by sleeping, not by the real clock. Only the daily visitor uses the real date. | Deterministic, kid-controlled, testable. |
| Combat | Real-time action on the existing paw button (interact first, swing otherwise). Enemies telegraph for at least 0.6 s, hitboxes are generous, health is petals, zero petals means a nap, age presets scale enemies, Peaceful mode sleeps them. Behaviours are code, stats are data. | The Zelda and Kitaria feel, kept to one stick and one button, and never a game over. |
| UI framework | Hand-rolled: 9-slice panels, big buttons. | The UI is small and unusual (icons, huge targets). `ebitenui` exists if it grows. |
| Dependencies | Ebitengine, possibly `go-tiled`, stdlib for JSON. | Keep the dependency list short. |

### 4.4 Browser specifics

- `web/index.html`: full-screen canvas, a "Tap to start" overlay (unlocks
  audio, requests fullscreen on touch devices), `touch-action: none`, no pinch
  zoom, no text selection, a loading bar.
- `web/loader.js` downloads `main.wasm` with progress. The size comes from the
  page, where `make web` writes it, or from `Content-Length` under wasmserve.
  The game starts on the tap. Ebitengine adds its own viewport tag at start,
  so the loader keeps ours in force.
- PWA: `manifest.webmanifest`, app icons in `web/icons/` and a service worker
  (`web/sw.js`), so the game can be added to the tablet home screen and works
  offline. The worker stores the whole bundle when it installs, then answers
  network-first with a cache fallback: an online visit always gets the newest
  build, and the stored copy is used only when the network fails. `loader.js`
  registers it only where service workers exist (not on plain http over the
  LAN), and a failed registration never stops the game.
- Portrait on a phone: show a "turn your device" picture (iOS cannot lock
  orientation from a web page). The hint is CSS-only: a media query shows it
  on coarse-pointer screens up to 600 CSS px wide held in portrait, with no
  JavaScript, so it also covers the loading screen.
- Serve `.wasm` as `application/wasm` with Brotli or gzip. Cloudflare Pages,
  Netlify and GitHub Pages all work as free static hosts; verify compression.
- Build with `-trimpath -ldflags="-s -w"`; optionally try `wasm-opt -Oz` for a
  further size cut. M0 measured 18.9 MB raw and 4.4 MB with gzip -9.

### 4.5 Developer loop

- Daily work on desktop: `make run` (fast compile, debugger works). The
  `debug` build tag (`make run TAGS=debug`) adds an overlay, toggled with F3,
  with FPS, screen and world sizes, the input state and the critter's
  position. Later it gains collision boxes and hitboxes. The `-map` flag
  (`make run ARGS='-map meadow'`) jumps straight to a map; it exists today in
  every desktop build, not only `debug` ones.
- Browser check: `make serve` runs wasmserve, pinned as a Go tool in
  `go.mod`, and serves http://localhost:8080/web/, rebuilding the game on
  every reload. Test on a real tablet over the LAN from milestone 0 onwards;
  `?pad=1` turns on the touch pad.
- CI (GitHub Actions): `make check`, desktop builds for Linux, Windows and
  macOS, `make web`, and on main a deploy of `build/web` to GitHub Pages.

### 4.6 Testing

Follow the house rules: real implementations, test business rules, no tests
for rendering or wiring.

- Unit tests in the pure packages: quest state machine (every step type,
  counters, rewards), dialogue runner (conditions, choices, actions),
  inventory, Snack Time rules, save round-trip and version migration, movement
  and collision maths, spawner weights with a seeded RNG, i18n fallback.
- Combat tests: petals never go below zero, the invulnerability window, the
  knockback direction, the aim-assist cone, hitbox overlap, enemy behaviour
  state machines stepped with a fake clock (telegraph, attack, recover), the
  boss phase script, room-clear door logic, age-preset scaling, drop tables
  with a seeded RNG.
- **Content lint test**: loads every real data file and asserts that every
  referenced NPC, item, critter, enemy, map, sprite and string key exists in
  both languages, every map has a spawn point, and every warp resolves. This
  test is what catches a typo in a quest file.
- Determinism: inject a clock and a seeded `rand.Rand`; no sleeps.

## 5. Roadmap: the steps

Estimates assume evenings and weekends, with placeholder art until the art
work in section 6 kicks in. Each milestone ends with a kid playtest.

### M0 Foundation: "Hello Critter" (1 to 2 weeks)

- [x] `go mod init`, the folder layout above, a Makefile (`run`, `web`, `test`, `check`)
- [x] GitHub Actions workflow in `.github/workflows/ci.yml`
- [x] Ebitengine window: a placeholder critter moves with keyboard, touch (tap-to-walk and pad), gamepad and mouse, through the input abstraction in `engine`
- [x] World drawn at full screen resolution with a whole-number scale, full-resolution text and an adaptive logical width
- [x] The same build served by wasmserve (`make serve`)
- [ ] Tested on an iPad or Android tablet
- [x] `web/index.html` with tap-to-start
- [ ] Deployed to a URL from CI
- [x] Debug overlay behind a build tag

**Done when** a child moves a square around on a tablet from a link.

### M1 A place to walk: "Acorn Hollow" (2 to 3 weeks)

- [x] Tiled JSON loading, layers, collision grid, animated tiles (water)
- [ ] Camera follow and clamp; y-sorted props; "above" layer
- [x] Warps: village to meadow, village to house interior
- [ ] Player critter as parts, with four-direction idle and walk animated in code; `make art` exports the parts from Inkscape (the code-drawn placeholder is fine until the drawing is ready)
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
- [ ] Five critters with their parts and stickers; the "Scared Little Mouse" quest
- [ ] Sleep in bed to change time of day; night tint

**Done when** all five critters can be befriended, show up in the book, and one follows you.

### M5 Fights: "Brave little critter" (2 to 3 weeks)

- [ ] `combat` package with tests: hitboxes, petals, damage, invulnerability, knockback, aim-assist cone
- [ ] Player swing (three directions, mirrored), hurt and nap animations; paw logic (interact first, swing otherwise); tap-an-enemy on touch
- [ ] Enemy framework: `enemies.json`, behaviour state machines with telegraphs; Sulky Slime, Thistle Beast, Grumpy Mushroom
- [ ] Drops (hearts, snacks), slow petal regrowth, Big Petal item
- [ ] Companion assist: nips at enemies, cannot be hurt, cheers
- [ ] Age-preset scaling and Peaceful mode
- [ ] Hit feedback: white flash, pop particles, poof into flowers, soft sounds, a tiny hitstop
- [ ] Quest step types "defeat N" and "clear a room"; the "Slimes in the Berry Patch" quest at the meadow's edge

**Done when** a 7-year-old clears the meadow slimes and asks for more, and a 5-year-old on Little Explorer naps at most once.

### M6 Exploring and puzzles: "The Glow Cave" (3 to 4 weeks)

- [ ] Gating items: Snippers (brambles), Splash Boots (shallow water)
- [ ] Puzzle pieces: push block, flower switch, mushrooms in order, key and door, pots
- [ ] Darkness with a light radius (firefly companion or lantern) using a light-mask image
- [ ] Rooms that lock until cleared; sunflower checkpoint
- [ ] Boss: the Grumbling Grub with three telegraphed phases, star health bar, berry ending and sticker
- [ ] Glow Cave map (five rooms), Glow Seed, Sturdy Stick, main quest skeleton; "Fix the Bridge" opens the Woods

**Done when** a 6-year-old finishes the cave with at most one hint from an adult and beats the Grub within three tries.

### M7 Kid-proofing and polish: "Ready for the sofa" (2 to 3 weeks)

- [ ] Title: three slots, species and colour picker, preset names, language flags
- [ ] "Little Explorer" (5 to 6) and "Explorer" (7 to 9) presets: arrows, hints, enemy scaling, pad or tap
- [ ] Settings: sound, read-aloud, controls, Peaceful mode
- [ ] Juice: sparkles, hearts, confetti, soft screen transitions; no harsh flashes
- [ ] Loading screen with progress; PWA manifest and offline; orientation hint
- [ ] Performance and size pass on a tablet (60 fps, under about 10 MB over the wire)
- [ ] Three playtests with target-age kids; fix the top five confusions; tune enemy numbers per preset from the notes
- [ ] `CREDITS.md`, licences checked

**Done when** a parent can send a link and a 5-year-old plays for 10 minutes unassisted.

### M8 Content expansion (ongoing)

Whispering Woods and Pebble Beach, ten more critters, six more quests, two more
mini-dungeons with their own bosses, the full "Sleepy Tree" main quest, Buzzy
Bee and Rolling Pinecone, the Sparkle Wand, hats and house decoration, the
Follow the Bounce mini-game.

### Stretch goals (pick later)

Local co-op (a second player steers the companion with a gamepad and can
fight too), a Leaf Shield that blocks on its own while you stand still,
make-your-own-critter from parts, a tiny garden (plant a seed, berries after
you sleep), a birthday party event on a date set by a parent, photo mode
(saves a PNG), desktop and mobile store builds, optional cloud save through a
small Go server using a six-letter family code (no accounts).

Rough total to a playable vertical slice (M0 to M7): 5 to 6 months part-time,
plus the art time in 6.7.

## 6. Making the assets

### 6.1 Art direction on one page (write it down as `docs/ART-BIBLE.md` before drawing)

- **Mood words**: sunny, soft, round, friendly, storybook.
- **Look**: smooth vector shapes drawn at full screen resolution, never
  scaled-up pixels.
- **Palette**: pick one from Lospec and stick to it. Recommended: *Endesga 32*
  (bright, great for cute games), which the placeholder art uses, or
  *Resurrect 64* (more range). Load it into Inkscape as a swatch palette and
  never eyedrop colours from elsewhere.
- **Sizes**, in world pixels: tiles 16 by 16; player and villagers in a 32 by
  32 box (feet at bottom centre, collision box about 12 by 10 at the feet);
  small critters and enemies 16 or 24, the boss 48 by 48. Drawings are made at
  8 px per world pixel, so a tile is 128 by 128 px and a villager's box 256 by
  256. The UI is sized in CSS pixels, like the touch targets in section 3, and
  drawn at 2 px per CSS pixel: portraits about 160 by 160 with three
  expressions, stickers about 96 by 96 with a white "sticker" border, icons 32
  (HUD) and 64 (menus).
- **Shapes**: build everything from simple shapes (circles, ellipses, rounded
  triangles), as the M0 critter is. Big heads (about half the height), big
  eyes with a highlight, rounded everything, no sharp angles.
- **Enemies look grumpy, never scary**: furrowed eyebrows and a pout, no
  fangs, no blood, no glowing red eyes. A defeated enemy smiles.
- **Lines and shading**: characters and props get an outline about 1 world
  pixel thick (8 px in the drawing) in a darker version of the fill colour
  (not black); the ground gets none. Flat fills with one shade and one
  highlight per material; light from top-left; no textures, so kids get a
  clean read. Nothing thinner than half a world pixel, so details survive on
  a small laptop screen.
- **Contrast rule**: ground tiles are lighter and less saturated than
  characters, so characters pop.
- **Animation**: characters are sets of parts (head, body, ears, tail, feet,
  eyes) that the engine moves, turns and squashes, instead of drawn frames.
  Idle is a slow breathing squash and a blink every few seconds; walk is a
  bob with pattering feet and bouncing ears (the M0 critter already bobs);
  swing is a lean with the stick swept through an arc; hurt is a flash and a
  squash; small critters hop and squash on landing. Enemies: idle, move,
  telegraph, attack, hurt and poof, made the same way. Draw down, up and side
  views; the engine mirrors the side view. Shared emote sprites (heart,
  exclamation, zzz, sweat, music note) float above anyone.
- **References**: Cozy Grove and Spiritfarer (soft, hand-drawn, cozy), Sago
  Mini and Toca Boca (simple shapes young children read at a glance), Animal
  Crossing, Link's Awakening (Switch), Tunic (for colour and softness).
  Collect screenshots in a PureRef board.

### 6.2 Tools

| Need | Recommended | Free alternative |
|---|---|---|
| Drawing | Inkscape (free; its command line does the export for `make art`) | Affinity (free since late 2025, with a Canva account) or Figma, exporting PNGs by hand |
| Animation | In code: the engine moves, turns and squashes each character's parts | |
| Maps | Tiled | LDtk |
| Sound effects | jsfxr or ChipTone (browser, instant) | Bfxr |
| Music | Bosca Ceoil Blue or BeepBox (browser) | LMMS, FamiStudio |
| Audio editing and conversion | Audacity, ffmpeg | |
| Fonts | Google Fonts: Fredoka, Baloo 2 or Nunito (rounded, readable, OFL licence, include å ä ö) | Andika (made for beginning readers) |
| Palettes | Lospec (download as a GIMP `.gpl` file, which Inkscape reads as swatches) | |
| Placeholder art while you build | Shapes drawn in code, as in M0; Kenney.nl (CC0, much of it in a flat vector style) | OpenGameArt (check each licence) |
| Reference boards | PureRef | |

AI image tools are fine for mood boards and quick concept drafts. Final art is
still redrawn as vector parts in the palette, for consistency, and each tool's
licence terms need checking.

### 6.3 Step by step: characters and enemies

1. Make the **style sheet first**: one grass tile, one tree, one critter, one
   enemy, one UI panel and one portrait, placed on a 320 by 180 world pixel
   mock screen (2560 by 1440 px in the drawing) and viewed on the tablet.
   Iterate until it feels cute. Show it to the kids and ask which one they
   like.
2. One Inkscape file per character, at 8 px per world pixel, with the palette
   loaded: a villager's 32 by 32 box is 256 by 256 px. Set the document's
   units to px with a scale of 1 (Document Properties), so a px in the
   drawing is a pixel in the PNG. Block the silhouette with circles and
   ellipses: head about 16 world pixels across, body about 10, feet at the
   bottom centre.
3. Flat colours, then the shade and highlight, then the outline, then eyes
   and highlight.
4. Split it into parts, each its own object with an id named `<view>_<part>`:
   `down_head`, `down_body`, `down_ear_left`, `down_tail`, and so on, plus
   spare parts for expressions (`down_eyes_closed`, `down_eyes_happy`). Draw
   the down, up and side views; the engine mirrors the side view (`flipX`).
5. Mark the point each part turns around (the base of an ear, the root of the
   tail) with a tiny circle named `<view>_<part>_pivot`, and the feet with one
   named `<view>_feet`.
6. `make art` exports every part to its own PNG and writes a JSON file with
   where each part sits and turns, read from Inkscape's `--query-all`, so
   there are no hand-written offset tables.
7. Animate in code with a few motions every character shares: breathe, blink,
   bob, patter, wiggle, lean, squash and flash. A character's data picks the
   motions and their sizes and speeds, so a new critter needs no new code.
8. Portrait: the same parts approach at 2 px per CSS pixel, with `neutral`,
   `happy` and `sad` made by swapping eyes and mouth.
9. Cheap variants: recolour a copy (Edit, Select Same, Fill Color), or share
   a body and swap head, ear and tail parts.
10. Enemies: the same process in a 16 by 16 or 24 by 24 box, with motions for
    `idle`, `move`, `telegraph`, `attack`, `hurt` and `poof`. The telegraph is
    the most important animation in the game: make it big and obvious (puff
    up, shake, flash) so a 5-year-old sees the attack coming.

Part budget per playable critter: 3 views of about 8 parts, plus spare eyes
and mouths and the stick, about 30 parts. Villagers: about 20 (or 8 if they
only face down). Small critters: 5 to 10. Enemies: about 12. The boss: about
20.

### 6.4 Step by step: tiles, props and maps

1. Base terrain on a 128 by 128 px page, which is one 16 by 16 world pixel
   tile: grass plus two or three variants, dirt path, sand, water (a few
   frames of moving ripples), cave floor, cliff edges. Keep the ground plain
   and soft. Anything that crosses an edge gets a copy 128 px over on the
   opposite side, so the tile wraps seamlessly.
2. Transitions: soft, wavy edges as a 16-tile corner set per pair (grass to
   dirt, grass to water, sand to water). Draw one edge and one corner, then
   duplicate, turn and mirror them. In Tiled, define a **Terrain Set** so you
   paint transitions with one brush.
3. Ground decorations to break up the plain ground: tufts, flowers, pebbles
   and leaves, placed freely as tile objects or scattered over a zone with a
   seeded random generator, as the M0 meadow does.
4. Props as separate images: tree (32 by 48 world pixels), bush, berry bush
   (full and empty), rock, fence set, sign, house (64 by 64) with door, chest
   (closed and open), pot, push block, flower switch (off and on), mushroom
   (off and on), bramble, plank bridge, room door (open and locked),
   sunflower checkpoint.
5. Interior set: floor, walls, window, bed, table, rug, shelf.
6. Tilesets in Tiled: `tileset_outdoor.png`, `tileset_indoor.png` and
   `tileset_cave.png` as regular tilesets; props and decorations as an
   *image collection* tileset so each keeps its own size.
7. Map layout in Tiled (orthogonal, 128 px tiles, about 40 by 30). Layers:
   `ground`, `decor`, `props` (object layer with tile objects, y-sorted),
   `above` (tree tops, roofs), `collision` (a one-tile "solid" tileset, or
   rectangles), `objects` (`player_spawn`, `npc` with `id`, `warp` with
   `target_map` and `target_spawn`, `spawn_zone` with `habitat`, `enemy` with
   `id`, `room` rectangles with `locks_doors`, `item` with `item_id`,
   `trigger` with `event`, puzzle objects). Map properties: `music`, `biome`,
   `dark`. The loader divides Tiled's positions by 8, so the game works in
   world pixels.
8. Turn on "Repeat last export on save" in Tiled so the `.tmj` in
   `assets/maps/` is always current. Test with `go run ./cmd/game -map meadow`.

### 6.5 Step by step: UI, fonts, audio and voice

**UI**

- Panels drawn in code as rounded shapes with a soft shadow, like the M0
  speech bubble, or as 9-slice images where they need decoration; a button in
  normal and pressed states; a "next" arrow; a checkmark.
- Icon set at 32 and 64 CSS px: speech, footsteps, berry, hat, heart, star,
  sticker, backpack, gear, speaker, flags, snacks, key, lantern, zzz, stick,
  sturdy stick, wand.
- Petals HUD (full and empty petal), the boss's row of stars.
- Effects: swoosh arc, hit spark, poof cloud, flowers burst, spore cloud,
  dust puff.
- Virtual joystick (ring plus knob) and a big **paw-print** action button.
- Text at full resolution using Fredoka or Baloo 2 through Ebitengine's
  `text/v2`. Dialogue text about 1 cm tall on a tablet, at most two lines of
  about 40 characters.

**Sound effects** (jsfxr or ChipTone to WAV, then
`ffmpeg -i x.wav -c:a libvorbis -q:a 4 x.ogg`): step, talk blip, pickup, door,
chest, quest start (three notes up), quest done (fanfare), befriend success
(sparkle), wrong snack (a soft "bwomp", never harsh), menu tap, hearts,
sparkle, zzz, swing whoosh, hit pop, hurt "oof", poof, heart pickup, room lock
and unlock chime, the Grub's friendly grumble. Keep everything soft and cap the
master volume.

**Music** (BeepBox or Bosca Ceoil to WAV, then OGG at `-q:a 3`): four loops of
45 to 60 seconds for village, meadow, cave and boss; 90 to 120 BPM, major or
pentatonic, soft instruments. The boss track is bouncy, not scary. Music is
the biggest asset, so keep loops short.

**Critter voices**: record about 20 syllables ("ba", "mi", "ko", ...) in a
high and a low voice and play one per couple of letters while the text types
out (the Animal Crossing trick). Cheap and charming. Read-aloud of the actual
words comes from the browser's speech voices; later, family-recorded lines can
replace them.

**Fonts** must include Latin Extended for å, ä and ö. Fredoka, Baloo 2, Nunito
and Andika do, all under the SIL Open Font License (keep the licence
text in `CREDITS.md`).

### 6.6 Pipeline and folders

```
assets-src/   Inkscape .svg, .tmx/.tsx, BeepBox .json, raw .wav   (sources of truth; consider git LFS for audio)
assets/       exported and embedded
  sprites/    critters/fox/*.png + fox.json (where parts sit and turn), npcs/, enemies/, props/, effects/, ui/, emotes/
  tiles/      tileset_outdoor.png ...
  maps/       village.tmj, meadow.tmj, ...
  audio/      music/*.ogg, sfx/*.ogg|wav, voices/
  fonts/      Fredoka-Medium.ttf
  data/       critters.json, enemies.json, items.json, quests/*.json, dialogue/*.json, strings/sv.json, strings/en.json
```

- `make art` runs `go run ./tools/art`, a Go command (logic in
  `tools/art/artexport`) that drives Inkscape 1.x's command line over the
  SVGs in `assets-src/sprites/` and `assets-src/tiles/` that are newer than
  their export (`make art ART_ARGS=-force` exports all of them;
  `INKSCAPE=/path/to/inkscape` when it is not on `PATH`). The folder decides
  the kind of drawing. A character in `assets-src/sprites/critters/`,
  `npcs/` or `enemies/`, for example `critters/fox.svg`, becomes one PNG per
  `<view>_<part>` object plus `fox.json` in `assets/sprites/critters/fox/`;
  the JSON gives each part's box and pivot in world pixels from the
  `<view>_feet` marker, read from `--query-all`, and PNGs of parts no longer
  in the drawing are deleted. Every other drawing (props, effects, ui,
  emotes, tiles) becomes one PNG of its page at the mirrored path, for
  example `assets/tiles/tileset_outdoor.png`. Everything is exported at
  8 px per world pixel. Only whole parts get a `<view>_` id: a named object
  inside a part, such as `down_head_shine`, is exported as a part of its own.
  Until the first drawing lands the placeholder art is drawn in code. The
  exporter's tests use a fake Inkscape; `go test -tags inkscape ./tools/art/...`
  runs the real one on a machine that has it. Later `make audio` runs ffmpeg;
  `make check` runs gofmt, `go mod tidy -diff`, vet for desktop and browser
  and every test, which will include the content lint test; `make web` builds
  the wasm bundle into `build/web/`.
- Naming: lower_snake_case, one drawing per character, parts named
  `<view>_<part>`.
- Licences: prefer CC0; CC-BY needs an entry in `CREDITS.md`; avoid NC and ND
  material in case the game is ever sold. Decide a licence for your own art
  too (the repo is MIT, which covers code; art can be the same or CC-BY).

### 6.7 Asset list for the vertical slice (rough first-timer hours)

| Asset | Count | Hours |
|---|---|---|
| Terrain tiles and transitions, with plain ground | about 40 | 8 to 10 |
| Ground decorations (tufts, flowers, pebbles, leaves) | about 15 | 3 |
| Props and interiors | about 30 | 10 |
| Playable critters (fox, bunny, hedgehog via swaps), as parts in three views | 3 | 8 to 10 |
| Villagers | 5 | 10 to 12 |
| Wild critters | 5 | 6 to 8 |
| Enemies (slime, thistle beast, mushroom) | 3 | 8 to 10 |
| Boss (the Grumbling Grub) | 1 | 5 to 6 |
| Portraits (three expressions each, by swapping eyes and mouths) | 6 | 6 |
| UI, icons, pad, logo, petals, effects | about 55 pieces | 12 |
| Sound effects | about 22 | 4 |
| Music loops | 4 | 8 |
| **Total** | | **about 90 to 100** |

Strategy for a non-artist: build the whole game on placeholders (shapes drawn
in code, CC0 packs), replace one category at a time, build characters from
circles and ellipses, let the code animate parts instead of drawing frames,
and reuse a base body with colour swaps. The animation code (breathe, blink,
bob, lean, squash) is engine work in M1 and M5, not art time. Buying a pack in
a matching smooth style (5 to 30 EUR) or commissioning a few key characters is
a legitimate shortcut.

## 7. Suggestions beyond the brief

1. **Read-aloud plus icons** is the single biggest accessibility win for 5 and
   6-year-olds. Build it in M2, not later.
2. **Tap-to-walk as the default touch control** turns the game into
   point-and-click for the youngest.
3. **Two languages from day one**; never hard-code a string.
4. **Smooth art, full-res text**: the world is drawn at full screen
   resolution, so shapes stay soft and text stays crisp and readable.
5. **Age presets** ("Little Explorer" and "Explorer") instead of a difficulty
   menu.
6. **Grumpies become friends.** Every enemy is a grumpy version of a critter.
   Calm five of a kind and it appears in the sticker book; every boss turns
   into a friend when beaten. Fights feed the collecting loop instead of
   fighting it.
7. **PWA**: an icon on the tablet home screen that works offline, including in
   the car.
8. **Make-your-own-critter**: pick ears, tails, patterns and colours from the
   game's parts and palette, and paint on the body with a finger. Cheap with
   critters made of parts, huge for kids.
9. **Birthday party event**: a parent sets a date in settings and the village
   celebrates that day.
10. **Photo mode**: one button saves a screenshot for showing grandparents.
11. **Skip real-world GPS**: browsers have no step counter, and outdoor play
    with a tablet is a poor fit for ages 5 to 9. Take the collecting loop,
    leave the walking.
12. **Content as data**, so adding a quest is adding a JSON file, and the
    content lint test catches mistakes.
13. **Playtest protocol**: 10-minute sessions, sit behind, say nothing, write
    down every place they look lost or ask "what now?" in
    `docs/PLAYTEST-LOG.md`. Ask "what would you do next?" rather than
    explaining.
14. Later, a **tiny Go server** for family cloud saves via a short code (no
    accounts, no email). The save types are the same Go structs on both sides.

## 8. Risks

| Risk | Mitigation |
|---|---|
| WASM size and first load on tablets | Brotli hosting, short music loops, a size budget checked in CI, a loading screen with progress |
| Browser audio blocked until a gesture | "Tap to start" overlay; the audio manager unlocks on first input |
| iOS Safari quirks (memory, touch, speech, no orientation lock) | Test on an iPad from M0; orientation hint; keep memory under about 300 MB |
| Fights too hard or too scary for the youngest | Telegraphs of at least 0.6 s, generous hitboxes, slow petal regrowth, age presets, Peaceful mode, grumpy-not-monstrous art rule, playtests with real 5-year-olds |
| Accidental swings at villagers or during dialogue | Interact has priority over swing; villagers have no hurt-box; overlays swallow the paw button |
| Art is the long pole | Placeholders drawn in code and CC0 packs first; art bible; parts animated in code instead of drawn frames; colour swaps |
| Smooth art is 1:1 at only one scale, and larger than pixel art | Export at 8 px per world pixel, 1:1 on a 2048 by 1536 iPad and a little softer elsewhere; vector sources make another density one more export; fetch art at runtime if the bundle passes the size budget |
| Scope creep | Vertical slice M0 to M7 before any content expansion; the stretch list stays a list |
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
| 2026-11-27 | Real-time action fights on the paw button, telegraphed enemies, petals and nap, age presets, Peaceful mode | Boop-only grumpies; turn-based critter battles; both | The Zelda and Kitaria feel the game is modelled on, kept to one stick and one button, never a game over. Turn-based would be a second game mode and more reading. |
| 2026-11-27 | localStorage and file saves, no accounts | Backend with logins | Pillar 6 |
| 2026-11-27 | sv and en from day one | English only | Needed at home; retrofitting is painful |
| 2026-11-27 | View: the largest whole-number scale that shows at least 240 by 180 world px; the world fills the screen at that scale, up to 480 by 240 | A fixed 180 px height with adaptive width | A 2048 by 1536 iPad fills the screen at scale 8 instead of getting a border, and very wide or very tall screens cannot see too much of a map |
| 2026-11-27 | Input reports Pointing, a finger or mouse button held at a point, instead of a one-off TapAt | TapAt events | One rule covers tap-to-walk and leading the critter with a held finger; a tap is a short hold |
| 2026-11-27 | Smooth art drawn at full screen resolution, replacing pixel art drawn small and scaled up; sizes stay in world pixels at a whole-number scale | Pixel art; finer pixel art at a higher logical resolution | Pixel art and smooth art were compared with the same placeholder art, and the smooth version looked better. Finer pixels would still look pixelated, with more pixels to draw. Keeping world pixels and the whole-number scale left maps, collision, the camera and their tests as they were, and tiles still meet on whole screen pixels |
| 2026-11-27 | Art drawn in Inkscape and exported by `make art` to PNG at 8 px per world pixel; characters are parts animated in code | Aseprite frames; SVG drawn at run time (for example with oksvg); Spine | Free, with a command line for the export, and 1:1 at scale 8. Parts and code motions avoid drawing every frame, the slow part of smooth art. SVG drawn at run time would be sharp at every scale, but Go's SVG renderers support only part of SVG, and drawing it costs load time in the browser |
| 2026-11-27 | Endesga 32 palette for all game art (`assets-src/palettes/endesga_32.gpl`) | Resurrect 64; custom palette | Bright and warm for a cute storybook look, small enough to keep fills and darker outlines consistent across drawings, and already used by the M0 placeholder art |

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

An enemy (behaviours are code; everything else is data):

```json
{
  "id": "sulky_slime",
  "name": {"en": "Sulky Slime", "sv": "Surt Slem"},
  "sprite": "enemies/sulky_slime",
  "behaviour": "hop_toward",
  "hp": 1,
  "damage": 1,
  "speed": 30,
  "telegraphSeconds": 0.6,
  "drops": [{"item": "heart", "chance": 0.5}, {"item": "berry", "chance": 0.2}],
  "littleExplorer": {"hp": 1, "damage": 1, "speed": 20},
  "sticker": "calm_slime",
  "stickerAfter": 5
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

A fight step inside a quest: `{"type": "defeat", "enemy": "sulky_slime", "count": 3, "icon": "stick"}`

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
go mod init github.com/Elpulgo/critterkingdom
go get github.com/hajimehoshi/ebiten/v2@latest          # needs Go 1.25 or later

# run on desktop
go run ./cmd/game

# run in the browser (rebuilds on every reload)
go get -tool github.com/hajimehoshi/wasmserve@latest        # once; pins it in go.mod
go tool wasmserve ./cmd/game                                # then open http://localhost:8080/web/

# build the web bundle
make web    # or by hand:
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o build/web/main.wasm ./cmd/game
cp -R web/. build/web/ && rm -f build/web/*_test.go         # page, loader, manifest, icons, worker
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" build/web/      # Go 1.23 and older: misc/wasm/wasm_exec.js

# assets
make art    # export changed drawings; ART_ARGS=-force for all, INKSCAPE=/path/to/inkscape if not on PATH
# what make art runs for each part of a character, by hand:
inkscape assets-src/sprites/critters/fox.svg --query-all    # id,x,y,width,height of every object
inkscape assets-src/sprites/critters/fox.svg --export-type=png --export-id=down_head --export-id-only \
  --export-area-snap --export-dpi=96 \
  --export-filename=assets/sprites/critters/fox/down_head.png   # one part, at 8 px per world pixel
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
- Inkscape: https://inkscape.org/ and its command line https://inkscape.org/doc/inkscape-man.html
- Lospec palettes: https://lospec.com/palette-list
- Kenney (CC0 assets): https://kenney.nl/assets
- jsfxr: https://sfxr.me/ and ChipTone: https://sfbgames.itch.io/chiptone
- BeepBox: https://www.beepbox.co/ and Bosca Ceoil Blue: https://github.com/YuriSizov/boscaceoil-blue
- Google Fonts: https://fonts.google.com/
- Web Speech API: https://developer.mozilla.org/en-US/docs/Web/API/SpeechSynthesis
