package assets_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
)

func TestAudio_ShippedVillageLoopAndStepSoundLoadAsOggData(t *testing.T) {
	tests := []struct {
		name string
		load func(string) ([]byte, error)
		file string
	}{
		{"village music loop", assets.LoadMusic, "village"},
		{"step sound", assets.LoadSound, "step"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.load(tc.file)
			if err != nil {
				t.Fatalf("load %q: %v", tc.file, err)
			}
			if !bytes.HasPrefix(b, []byte("OggS")) {
				t.Errorf("load %q: data does not start with OggS", tc.file)
			}
		})
	}
}

func TestLoadMusic_UnknownTrackIsAnErrorThatNamesIt(t *testing.T) {
	_, err := assets.LoadMusic("nope")
	if err == nil {
		t.Fatal(`LoadMusic("nope") returned no error`)
	}
	if !strings.Contains(err.Error(), `"nope"`) {
		t.Errorf(`LoadMusic("nope") error = %q; want it to name the track "nope"`, err)
	}
}
