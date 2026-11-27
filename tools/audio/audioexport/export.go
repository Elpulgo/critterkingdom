// Package audioexport renders the sound catalog of tools/audio/synth to WAV
// files and encodes each one to an OGG Vorbis file under the output folder.
package audioexport

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Elpulgo/critterkingdom/tools/audio/synth"
)

// Encoder turns a WAV file into an OGG Vorbis file.
type Encoder interface {
	// Encode reads the WAV at wav and writes an OGG Vorbis file at ogg with
	// the given Vorbis quality (-q:a). The folder of ogg already exists.
	Encode(wav, ogg string, quality int) error
}

// Options configures Run.
type Options struct {
	// Sounds are the sounds to export; nil means synth.Catalog().
	Sounds []synth.Sound
	// OutDir is the folder the OGG files go to; defaults to "assets/audio".
	OutDir string
	// WAVDir is the folder the rendered WAV files go to; defaults to
	// "build/audio-src".
	WAVDir string
	// Encoder does the WAV to OGG conversion.
	Encoder Encoder
	// Log receives one line per written or removed file; nil discards it.
	Log io.Writer
}

// Quality returns the Vorbis quality a sound of the given kind is encoded
// with: 3 for music, the biggest files, and 4 for everything else.
func Quality(k synth.Kind) int {
	if k == synth.Music {
		return 3
	}
	return 4
}

// partialSuffix marks the file an encoder writes to before it is renamed to
// its final name, so a failed encode never leaves a partial OGG behind. It
// ends in .ogg so the encoder still sees which format to write.
const partialSuffix = ".partial.ogg"

// Run renders every sound to a WAV under WAVDir and encodes it to its path
// under OutDir, always re-encoding all of them. It stops at the first sound
// that fails, with an error naming its path. After every sound is written,
// it deletes the .ogg files under OutDir that are not in the catalog; any
// other file is left alone.
func Run(o Options) error {
	if o.Sounds == nil {
		o.Sounds = synth.Catalog()
	}
	if o.OutDir == "" {
		o.OutDir = filepath.Join("assets", "audio")
	}
	if o.WAVDir == "" {
		o.WAVDir = filepath.Join("build", "audio-src")
	}
	if o.Log == nil {
		o.Log = io.Discard
	}
	if o.Encoder == nil {
		return errors.New("audioexport: no encoder given")
	}

	keep := make(map[string]bool, len(o.Sounds))
	for _, s := range o.Sounds {
		if err := export(o, s); err != nil {
			return fmt.Errorf("%s: %w", s.Path, err)
		}
		keep[s.Path] = true
	}
	return removeStale(o, keep)
}

// export renders one sound to its WAV and encodes it to its OGG.
func export(o Options, s synth.Sound) error {
	rel := filepath.FromSlash(s.Path)
	wav := filepath.Join(o.WAVDir, strings.TrimSuffix(rel, ".ogg")+".wav")
	ogg := filepath.Join(o.OutDir, rel)

	if err := writeWAV(wav, s.Render()); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ogg), 0o755); err != nil {
		return err
	}
	partial := strings.TrimSuffix(ogg, ".ogg") + partialSuffix
	if err := o.Encoder.Encode(wav, partial, Quality(s.Kind)); err != nil {
		os.Remove(partial)
		return err
	}
	if err := os.Rename(partial, ogg); err != nil {
		os.Remove(partial)
		return err
	}
	fmt.Fprintf(o.Log, "wrote %s\n", ogg)
	return nil
}

// writeWAV writes samples to a WAV file at p, creating its folder.
func writeWAV(p string, samples []float64) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	if err := synth.WriteWAV(f, samples); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// removeStale deletes every .ogg file under OutDir whose slash-separated
// path is not in keep.
func removeStale(o Options, keep map[string]bool) error {
	return filepath.WalkDir(o.OutDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Ext(d.Name()) != ".ogg" {
			return nil
		}
		rel, err := filepath.Rel(o.OutDir, p)
		if err != nil {
			return err
		}
		if keep[filepath.ToSlash(rel)] {
			return nil
		}
		if err := os.Remove(p); err != nil {
			return err
		}
		fmt.Fprintf(o.Log, "removed %s\n", p)
		return nil
	})
}
