package anim_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/anim"
)

// newTestAnimator returns an animator for a small critter. Idle breathes the
// body (size 0.1, a quarter period at 0.25 s) and blinks the eyes, closed for
// the last 0.75 s of each second. Walk bobs everything by 1 (highest at
// 0.25 s), patters the foot by 2 and wiggles the tail by 10 degrees (both at
// their top at 0.25 s). The times used below are sums of powers of two, so
// they add up exactly.
func newTestAnimator(t *testing.T) *anim.Animator {
	t.Helper()
	layout, err := anim.ParseLayout([]byte(`{"pxPerWorldPixel": 8, "views": {"down": {"parts": [
		{"name": "body", "image": "down_body.png", "x": -6, "y": -10, "w": 12, "h": 10, "pivotX": 0, "pivotY": -5},
		{"name": "eyes", "image": "down_eyes.png", "x": -3, "y": -9, "w": 6, "h": 2, "pivotX": 0, "pivotY": -8},
		{"name": "foot", "image": "down_foot.png", "x": -2, "y": -2, "w": 4, "h": 2, "pivotX": 0, "pivotY": -1},
		{"name": "tail", "image": "down_tail.png", "x": 4, "y": -6, "w": 4, "h": 3, "pivotX": 4, "pivotY": -4}
	]}}}`))
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	c, err := anim.ParseCharacter(layout, []byte(`{"sprite": "test/critter", "animations": {
		"idle": [
			{"motion": "breathe", "parts": ["body"], "size": 0.1, "period": 1},
			{"motion": "blink", "parts": ["eyes"], "period": 1, "duration": 0.75}
		],
		"walk": [
			{"motion": "bob", "size": 1, "period": 0.5},
			{"motion": "patter", "parts": ["foot"], "size": 2, "period": 1},
			{"motion": "wiggle", "parts": ["tail"], "size": 10, "period": 1}
		]
	}}`))
	if err != nil {
		t.Fatalf("ParseCharacter: %v", err)
	}
	return anim.NewAnimator(c)
}

// byName indexes a pose by part name.
func byName(poses []anim.PartPose) map[string]anim.PartPose {
	m := make(map[string]anim.PartPose, len(poses))
	for _, p := range poses {
		m[p.Name] = p
	}
	return m
}

func TestAnimator_ACritterThatHasNeverWalkedShowsOnlyIdleMotion(t *testing.T) {
	a := newTestAnimator(t)
	for range 2 {
		a.Update(0.125, false) // 0.25 s of idle: breathing at its widest
	}

	poses := a.Pose(anim.ViewDown)
	for _, p := range poses {
		if p.Offset.X != 0 || p.Offset.Y != 0 || p.Turn != 0 {
			t.Errorf("%s offset %v turn %v, want no bob, patter or wiggle", p.Name, p.Offset, p.Turn)
		}
	}
	parts := byName(poses)
	if body := parts["body"]; !near(body.ScaleX, 0.9) || !near(body.ScaleY, 1.1) {
		t.Errorf("body scale %v x %v, want breathing at full size 0.9 x 1.1", body.ScaleX, body.ScaleY)
	}
	if eyes := parts["eyes"]; !near(eyes.ScaleY, 0.1) {
		t.Errorf("eyes scaleY %v, want closed (0.1) by the idle blink", eyes.ScaleY)
	}
}

func TestAnimator_AfterWalkingLongerThanTheBlendTimeThePoseIsTheWalkAlone(t *testing.T) {
	a := newTestAnimator(t)
	for range 4 {
		a.Update(0.0625, true) // 0.25 s of walking, past the 0.1 s blend
	}

	want := map[string]struct{ lift, turn float64 }{
		"body": {lift: -1},
		"eyes": {lift: -1},
		"foot": {lift: -3}, // bob 1 plus patter 2
		"tail": {lift: -1, turn: math.Pi / 18},
	}
	for _, p := range a.Pose(anim.ViewDown) {
		w := want[p.Name]
		if p.Offset.X != 0 || !near(p.Offset.Y, w.lift) || !near(p.Turn, w.turn) {
			t.Errorf("%s offset %v turn %v, want lift %v turn %v", p.Name, p.Offset, p.Turn, w.lift, w.turn)
		}
		// Idle is also 0.25 s in, where breathing and the blink would show.
		if p.ScaleX != 1 || p.ScaleY != 1 {
			t.Errorf("%s scale %v x %v, want 1 x 1: no breathing or idle blink while walking", p.Name, p.ScaleX, p.ScaleY)
		}
	}
}

func TestAnimator_AfterStoppingTheWalkLiftShrinksToZeroWithinTheBlendTime(t *testing.T) {
	a := newTestAnimator(t)
	for range 4 {
		a.Update(0.0625, true) // walk is at its highest: body lifted by 1
	}

	a.Update(1.0/60, false)
	if body := byName(a.Pose(anim.ViewDown))["body"]; !near(body.Offset.Y, -5.0/6) {
		t.Fatalf("body lift one frame after stopping = %v, want -5/6", body.Offset.Y)
	}
	for range 5 {
		a.Update(1.0/60, false) // six frames of 1/60 s: the whole blend time
	}
	for frame := range 60 {
		for _, p := range a.Pose(anim.ViewDown) {
			if p.Offset.Y != 0 || p.Turn != 0 {
				t.Fatalf("%d frames after the blend: %s offset %v turn %v, want exactly 0", frame, p.Name, p.Offset, p.Turn)
			}
		}
		a.Update(1.0/60, false)
	}
}

func TestAnimator_TheSameFrameTimesAlwaysGiveTheSamePose(t *testing.T) {
	frames := []struct {
		dt      float64
		walking bool
	}{
		{0.016, false}, {0.033, true}, {0.017, true}, {0.05, true},
		{0.02, false}, {0.031, false}, {0.016, true}, {0.4, false},
	}
	first, second := newTestAnimator(t), newTestAnimator(t)
	for i, f := range frames {
		first.Update(f.dt, f.walking)
		second.Update(f.dt, f.walking)
		if a, b := first.Pose(anim.ViewDown), second.Pose(anim.ViewDown); !reflect.DeepEqual(a, b) {
			t.Fatalf("frame %d: poses differ:\n%+v\n%+v", i, a, b)
		}
	}
}
