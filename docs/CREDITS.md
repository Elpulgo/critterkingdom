# Credits

## Fonts

**Fredoka** by Milena Brandão and Hafontia sets all text in the game.

- Source: the Google Fonts repository, `ofl/fredoka`, from the upstream
  project at https://github.com/hafontia/Fredoka-One.
- Copyright 2016 The Fredoka Project Authors.
- Licence: SIL Open Font License 1.1, with no Reserved Font Name. The licence
  text is in [assets/fonts/OFL.txt](../assets/fonts/OFL.txt) and ships inside
  the game next to the font.

## Code

The game's code is under the MIT licence in [LICENSE](../LICENSE). Go
modules it depends on keep their own licences.

## Art and audio

**Endesga 32** by ENDESGA is the colour palette for the game's art, stored as
an Inkscape swatch palette at
[assets-src/palettes/endesga_32.gpl](../assets-src/palettes/endesga_32.gpl).

- Source: https://lospec.com/palette-list/endesga-32.
- Licence: free to use (Lospec palette list; credit appreciated).

**Sound effects, music and critter voices** in [assets/audio/](../assets/audio/)
are made by the Critter Kingdom project. Every file, the 21 sound effects in
`sfx/`, the four music loops in `music/` and the 20 syllables in each voice in
`voices/high/` and `voices/low/`, is synthesised by our own Go code in
[tools/audio/](../tools/audio/) and encoded to OGG Vorbis with ffmpeg by
`make audio`. No samples or recordings by others are used.

- Source: [tools/audio/synth/](../tools/audio/synth/) is the source of truth;
  the OGG files are generated from it.
- Licence: CC0 1.0 Universal (public domain dedication),
  https://creativecommons.org/publicdomain/zero/1.0/. The code in
  `tools/audio/` itself is MIT like the rest of the code.

The art on screen in M0 is drawn in code. Art or audio made by others gets an
entry here with its author, source and licence before it is added.
