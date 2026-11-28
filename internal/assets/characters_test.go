package assets_test

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/anim"
	"github.com/Elpulgo/critterkingdom/internal/assets"
)

func TestLoadCharacter_TheFoxIsBuiltFromPartsInEveryViewAndUsesAllFiveMotions(t *testing.T) {
	fox, err := assets.LoadCharacter("fox")
	if err != nil {
		t.Fatalf(`LoadCharacter("fox"): %v`, err)
	}

	for _, view := range []string{anim.ViewDown, anim.ViewUp, anim.ViewSide} {
		v := fox.Layout.View(view)
		if v.Name != view {
			t.Errorf("the fox has no %s view (got the %s view instead)", view, v.Name)
			continue
		}
		has := make(map[string]bool, len(v.Parts))
		for _, p := range v.Parts {
			has[p.Name] = true
		}
		for _, part := range []string{"head", "body", "ear_left", "ear_right", "tail", "foot_left", "foot_right"} {
			if !has[part] {
				t.Errorf("the fox's %s view has no %s", view, part)
			}
		}
	}

	used := make(map[anim.MotionKind]bool)
	for _, a := range []anim.Animation{fox.Idle, fox.Walk} {
		for _, m := range a.Motions {
			used[m.Kind] = true
		}
	}
	for _, kind := range []anim.MotionKind{anim.Breathe, anim.Blink, anim.Bob, anim.Patter, anim.Wiggle} {
		if !used[kind] {
			t.Errorf("neither the fox's idle nor its walk uses %s", kind)
		}
	}
}

func TestLoadCharacter_ACharacterWithNoRigFileIsAnErrorNamingTheFile(t *testing.T) {
	_, err := assets.LoadCharacter("unicorn")
	if err == nil {
		t.Fatal(`LoadCharacter("unicorn") returned no error`)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error %v does not wrap fs.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), "data/rigs/unicorn.json") {
		t.Errorf("error %q does not name data/rigs/unicorn.json", err)
	}
}
