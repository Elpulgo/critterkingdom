//go:build ffmpeg

// This test runs the real ffmpeg (an external program not installed in CI),
// so it only builds with the tag: go test -tags ffmpeg ./tools/audio/...

package audioexport_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Elpulgo/critterkingdom/tools/audio/audioexport"
	"github.com/Elpulgo/critterkingdom/tools/audio/synth"
)

func TestFFmpegExport_RenderedSoundBecomesAnOggFile(t *testing.T) {
	var pickup []synth.Sound
	for _, s := range synth.Catalog() {
		if s.Path == "sfx/pickup.ogg" {
			pickup = append(pickup, s)
		}
	}
	if len(pickup) != 1 {
		t.Fatalf("catalog has %d sfx/pickup.ogg entries, want 1", len(pickup))
	}
	out := t.TempDir()

	err := audioexport.Run(audioexport.Options{
		Sounds:  pickup,
		OutDir:  out,
		WAVDir:  t.TempDir(),
		Encoder: audioexport.NewFFmpeg("ffmpeg"),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(out, "sfx", "pickup.ogg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("OggS")) {
		t.Errorf("sfx/pickup.ogg starts with %q, want OggS", data[:min(len(data), 4)])
	}
	if !bytes.Contains(data, []byte("\x01vorbis")) {
		t.Error("sfx/pickup.ogg has no Vorbis identification header")
	}
}
