package audioexport_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/tools/audio/audioexport"
	"github.com/Elpulgo/critterkingdom/tools/audio/synth"
)

// fakeEncoder stands in for the ffmpeg process (an external program). It
// writes a stub OGG and records the quality asked for per final OGG path
// under out.
type fakeEncoder struct {
	out       string
	qualities map[string]int
	// failFor makes Encode write a few bytes and then fail for a WAV whose
	// name ends in this suffix.
	failFor string
}

func (f *fakeEncoder) Encode(wav, ogg string, quality int) error {
	data, err := os.ReadFile(wav)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(data, []byte("RIFF")) {
		return errors.New("fake encoder: input is not a WAV: " + wav)
	}
	if f.failFor != "" && strings.HasSuffix(wav, f.failFor) {
		os.WriteFile(ogg, []byte("Ogg"), 0o644)
		return errors.New("fake encoder: broken input")
	}
	if f.qualities == nil {
		f.qualities = map[string]int{}
	}
	rel, err := filepath.Rel(f.out, ogg)
	if err != nil {
		return err
	}
	f.qualities[strings.Replace(filepath.ToSlash(rel), ".partial", "", 1)] = quality
	return os.WriteFile(ogg, []byte("OggS"), 0o644)
}

func blip() []float64 { return []float64{0, 0.3, -0.3, 0} }

// testSounds is a small catalog with one sound of each kind.
func testSounds() []synth.Sound {
	return []synth.Sound{
		{Path: "sfx/pickup.ogg", Kind: synth.SFX, Render: blip},
		{Path: "music/village.ogg", Kind: synth.Music, Render: blip},
		{Path: "voices/high/ba.ogg", Kind: synth.Voice, Render: blip},
	}
}

// filesUnder lists every file under dir as slash-separated relative paths.
func filesUnder(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		files = append(files, filepath.ToSlash(rel))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return files
}

func TestRun_WritesOneOGGPerCatalogEntry_MusicAtQuality3AndTheRestAt4(t *testing.T) {
	out, wav := t.TempDir(), t.TempDir()
	enc := &fakeEncoder{out: out}

	err := audioexport.Run(audioexport.Options{Sounds: testSounds(), OutDir: out, WAVDir: wav, Encoder: enc})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []string{"music/village.ogg", "sfx/pickup.ogg", "voices/high/ba.ogg"}
	if got := filesUnder(t, out); !reflect.DeepEqual(got, want) {
		t.Errorf("files under the output folder = %v, want %v", got, want)
	}
	wantQ := map[string]int{"sfx/pickup.ogg": 4, "music/village.ogg": 3, "voices/high/ba.ogg": 4}
	if !reflect.DeepEqual(enc.qualities, wantQ) {
		t.Errorf("qualities asked for = %v, want %v", enc.qualities, wantQ)
	}
}

func TestRun_OGGNoLongerInTheCatalog_IsRemoved_AndOtherFilesAreKept(t *testing.T) {
	out := t.TempDir()
	for _, rel := range []string{"sfx/old_beep.ogg", "sfx/notes.txt", "voices/mid/ba.ogg", "README.md"} {
		p := filepath.Join(out, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	err := audioexport.Run(audioexport.Options{Sounds: testSounds(), OutDir: out, WAVDir: t.TempDir(), Encoder: &fakeEncoder{out: out}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []string{"README.md", "music/village.ogg", "sfx/notes.txt", "sfx/pickup.ogg", "voices/high/ba.ogg"}
	if got := filesUnder(t, out); !reflect.DeepEqual(got, want) {
		t.Errorf("files under the output folder = %v, want %v", got, want)
	}
}

func TestRun_FailingEncoder_StopsWithAnErrorNamingTheSound_AndLeavesNoPartialOGG(t *testing.T) {
	out := t.TempDir()
	enc := &fakeEncoder{out: out, failFor: "village.wav"}

	err := audioexport.Run(audioexport.Options{Sounds: testSounds(), OutDir: out, WAVDir: t.TempDir(), Encoder: enc})
	if err == nil {
		t.Fatal("Run succeeded; want the encoder's error")
	}
	if !strings.Contains(err.Error(), "music/village.ogg") {
		t.Errorf("error %q does not name music/village.ogg", err)
	}

	// The sound before the failure is written; the failed one leaves nothing
	// and the export stops there.
	want := []string{"sfx/pickup.ogg"}
	if got := filesUnder(t, out); !reflect.DeepEqual(got, want) {
		t.Errorf("files under the output folder = %v, want %v", got, want)
	}
}

func TestRun_MissingFFmpeg_FailsWithAnErrorNamingTheSoundAndTheFlag(t *testing.T) {
	out := t.TempDir()
	missing := filepath.Join(t.TempDir(), "no-ffmpeg-here")

	err := audioexport.Run(audioexport.Options{Sounds: testSounds(), OutDir: out, WAVDir: t.TempDir(), Encoder: audioexport.NewFFmpeg(missing)})
	if err == nil {
		t.Fatal("Run succeeded without ffmpeg")
	}
	for _, part := range []string{"sfx/pickup.ogg", "-ffmpeg"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error %q does not mention %s", err, part)
		}
	}
	if got := filesUnder(t, out); len(got) != 0 {
		t.Errorf("files under the output folder = %v, want none", got)
	}
}
