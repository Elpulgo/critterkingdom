package assets_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/assets"
)

func TestLoadSound_ShippedStepSoundLoadsAsOggData(t *testing.T) {
	b, err := assets.LoadSound("step")
	if err != nil {
		t.Fatalf(`LoadSound("step"): %v`, err)
	}
	if !bytes.HasPrefix(b, []byte("OggS")) {
		t.Error(`LoadSound("step"): data does not start with OggS`)
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
