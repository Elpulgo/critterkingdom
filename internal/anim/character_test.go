package anim_test

import (
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/anim"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// twoViewSheet has eyes in the down view only, so a motion on the eyes has
// nothing to move in the side view.
const twoViewSheet = `{"pxPerWorldPixel": 8, "views": {
	"down": {"parts": [
		{"name": "body", "image": "down_body.png", "x": -6, "y": -10, "w": 12, "h": 10, "pivotX": 0, "pivotY": -5},
		{"name": "eyes", "image": "down_eyes.png", "x": -3, "y": -9, "w": 6, "h": 2, "pivotX": 0, "pivotY": -8}
	]},
	"side": {"parts": [
		{"name": "body", "image": "side_body.png", "x": -7, "y": -10, "w": 14, "h": 10, "pivotX": 0, "pivotY": -5}
	]}
}}`

func parseTwoView(t *testing.T, data string) anim.Character {
	t.Helper()
	layout, err := anim.ParseLayout([]byte(twoViewSheet))
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	c, err := anim.ParseCharacter(layout, []byte(data))
	if err != nil {
		t.Fatalf("ParseCharacter: %v", err)
	}
	return c
}

func TestAnimationPose_AMotionWithoutPartsMovesTheWholeCharacter(t *testing.T) {
	c := parseTwoView(t, `{"sprite": "test/two", "animations": {"idle": [],
		"walk": [{"motion": "bob", "size": 1, "period": 0.3}]}}`)

	for _, view := range []string{anim.ViewDown, anim.ViewSide} {
		for _, p := range c.Walk.Pose(c.Layout.View(view), 0.15) {
			if !near(p.Offset.Y, -1) {
				t.Errorf("%s %s lift = %v, want -1", view, p.Name, p.Offset.Y)
			}
		}
	}
}

func TestAnimationPose_AMotionOnAPartTheViewLacksLeavesTheOtherPartsAlone(t *testing.T) {
	c := parseTwoView(t, `{"sprite": "test/two", "animations": {"idle": [],
		"walk": [{"motion": "bob", "parts": ["eyes"], "size": 1, "period": 0.3}]}}`)

	down := c.Walk.Pose(c.Layout.View(anim.ViewDown), 0.15)
	if !near(down[0].Offset.Y, 0) || !near(down[1].Offset.Y, -1) {
		t.Errorf("down lifts body %v eyes %v, want 0 and -1", down[0].Offset.Y, down[1].Offset.Y)
	}
	side := c.Walk.Pose(c.Layout.View(anim.ViewSide), 0.15)
	if len(side) != 1 || side[0].Offset != (geom.Vec{}) || side[0].ScaleX != 1 || side[0].ScaleY != 1 || side[0].Turn != 0 {
		t.Errorf("side pose = %+v, want only the body, unmoved", side)
	}
}

func TestParseCharacter_ANewCritterIsOnlyData(t *testing.T) {
	// A snail: one shell part, unlike any of the fox's, that wiggles while
	// idle and bobs while walking. No code knows about shells.
	layout, err := anim.ParseLayout([]byte(`{"pxPerWorldPixel": 8, "views": {"down": {"parts": [
		{"name": "shell", "image": "down_shell.png", "x": -4, "y": -8, "w": 8, "h": 8, "pivotX": 0, "pivotY": -4}
	]}}}`))
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	c, err := anim.ParseCharacter(layout, []byte(`{"sprite": "critters/snail", "animations": {
		"idle": [{"motion": "wiggle", "parts": ["shell"], "size": 20, "period": 2}],
		"walk": [{"motion": "bob", "parts": ["shell"], "size": 0.5, "period": 1}]
	}}`))
	if err != nil {
		t.Fatalf("ParseCharacter: %v", err)
	}
	if c.Sprite != "critters/snail" {
		t.Errorf("Sprite = %q, want critters/snail", c.Sprite)
	}

	// The snail has no side view, so it shows its down view sideways too.
	idle := c.Idle.Pose(c.Layout.View(anim.ViewSide), 0.5)
	if len(idle) != 1 || idle[0].Name != "shell" || !near(idle[0].Turn, 0.3490658503988659) {
		t.Errorf("idle pose at 0.5s = %+v, want the shell turned by 20 degrees", idle)
	}
	// Turned 20 degrees about its pivot (0,-4), the shell's top centre (0,-8)
	// swings right: (4·sin 20°, -4 - 4·cos 20°).
	top := idle[0].Apply(geom.Vec{X: 0, Y: -8})
	if !near(top.X, 1.3680805733026749) || !near(top.Y, -7.758770483143634) {
		t.Errorf("shell top at 0.5s = %v, want (1.368, -7.759)", top)
	}

	walk := c.Walk.Pose(c.Layout.View(anim.ViewDown), 0.5)
	if !near(walk[0].Offset.Y, -0.5) || walk[0].Turn != 0 {
		t.Errorf("walk pose at 0.5s = %+v, want the shell lifted by 0.5 and not turned", walk[0])
	}
}

func TestParseCharacter_RejectsDataItCannotAnimate(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		mention string // the error names the problem
	}{
		{"broken JSON", `{"sprite": "test/two", "animations": {`, "character data"},
		{"unknown motion", `{"animations": {"idle": [{"motion": "jump", "period": 1}], "walk": []}}`, `"jump"`},
		{"zero period", `{"animations": {"idle": [], "walk": [{"motion": "bob", "size": 1, "period": 0}]}}`, "period"},
		{"negative period", `{"animations": {"idle": [], "walk": [{"motion": "wiggle", "size": 5, "period": -1}]}}`, "period"},
		{"blink without a duration", `{"animations": {"idle": [{"motion": "blink", "parts": ["eyes"], "period": 4}], "walk": []}}`, "duration"},
		{"blink as long as its period", `{"animations": {"idle": [{"motion": "blink", "parts": ["eyes"], "period": 4, "duration": 4}], "walk": []}}`, "duration"},
		{"part in no view", `{"animations": {"idle": [{"motion": "wiggle", "parts": ["wing"], "size": 5, "period": 1}], "walk": []}}`, `"wing"`},
		{"no walk", `{"animations": {"idle": []}}`, "walk"},
		{"no idle", `{"animations": {"walk": []}}`, "idle"},
		{"unknown animation", `{"animations": {"idle": [], "walk": [], "wlak": []}}`, `"wlak"`},
	}
	layout, err := anim.ParseLayout([]byte(twoViewSheet))
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := anim.ParseCharacter(layout, []byte(tt.data))
			if err == nil {
				t.Fatalf("ParseCharacter accepted data with %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.mention) {
				t.Errorf("error %q does not mention %s", err, tt.mention)
			}
		})
	}
}
