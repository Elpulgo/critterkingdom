package anim

import (
	"encoding/json"
	"fmt"
	"sort"
)

// The animation names a character's data must have.
const (
	AnimIdle = "idle"
	AnimWalk = "walk"
)

// Animation is a named list of shared motions that play together.
type Animation struct {
	Name    string
	Motions []Motion
}

// Character is everything needed to pose a critter: its parts (from the part
// JSON) and its animations (from its character data). A new critter is a new
// pair of data files, not new code.
type Character struct {
	// Sprite is the part sheet the data names, for example "critters/fox".
	Sprite string
	Layout Layout
	Idle   Animation
	Walk   Animation
}

// characterJSON and motionJSON decode a character's data.
type characterJSON struct {
	Sprite     string                  `json:"sprite"`
	Animations map[string][]motionJSON `json:"animations"`
}

type motionJSON struct {
	Motion   string   `json:"motion"`
	Parts    []string `json:"parts"`
	Size     float64  `json:"size"`
	Period   float64  `json:"period"`
	Phase    float64  `json:"phase"`
	Duration float64  `json:"duration"`
}

// ParseCharacter reads a character's data, which picks the motions, sizes and
// speeds of its idle and walk animations, and joins it with the character's
// part layout. It returns an error for malformed JSON, a missing idle or walk
// animation, an animation it does not know, an unknown motion, a period that
// is not positive, a blink whose duration is not positive and shorter than
// its period, and a part name found in no view of the layout.
func ParseCharacter(layout Layout, data []byte) (Character, error) {
	var c characterJSON
	if err := json.Unmarshal(data, &c); err != nil {
		return Character{}, fmt.Errorf("anim: character data: %w", err)
	}

	names := make([]string, 0, len(c.Animations))
	for name := range c.Animations {
		names = append(names, name)
	}
	sort.Strings(names) // report the same error every time

	anims := make(map[string]Animation, len(c.Animations))
	for _, name := range names {
		if name != AnimIdle && name != AnimWalk {
			return Character{}, fmt.Errorf("anim: character data: unknown animation %q", name)
		}
		a, err := parseAnimation(layout, name, c.Animations[name])
		if err != nil {
			return Character{}, err
		}
		anims[name] = a
	}
	for _, name := range []string{AnimIdle, AnimWalk} {
		if _, ok := anims[name]; !ok {
			return Character{}, fmt.Errorf("anim: character data has no %s animation", name)
		}
	}
	return Character{
		Sprite: c.Sprite,
		Layout: layout,
		Idle:   anims[AnimIdle],
		Walk:   anims[AnimWalk],
	}, nil
}

func parseAnimation(layout Layout, name string, motions []motionJSON) (Animation, error) {
	a := Animation{Name: name, Motions: make([]Motion, 0, len(motions))}
	for i, mj := range motions {
		m := Motion{
			Kind:     MotionKind(mj.Motion),
			Parts:    mj.Parts,
			Size:     mj.Size,
			Period:   mj.Period,
			Phase:    mj.Phase,
			Duration: mj.Duration,
		}
		if err := checkMotion(layout, m); err != nil {
			return Animation{}, fmt.Errorf("anim: character data: %s motion %d: %w", name, i+1, err)
		}
		a.Motions = append(a.Motions, m)
	}
	return a, nil
}

func checkMotion(layout Layout, m Motion) error {
	if !knownMotion(m.Kind) {
		return fmt.Errorf("unknown motion %q", m.Kind)
	}
	if !(m.Period > 0) {
		return fmt.Errorf("%s period %v is not positive", m.Kind, m.Period)
	}
	if m.Kind == Blink && !(m.Duration > 0 && m.Duration < m.Period) {
		return fmt.Errorf("blink duration %v is not positive and shorter than its period %v", m.Duration, m.Period)
	}
	for _, part := range m.Parts {
		if !layout.hasPart(part) {
			return fmt.Errorf("%s names part %q, which no view has", m.Kind, part)
		}
	}
	return nil
}

// hasPart reports whether any view of the layout has the named part.
func (l Layout) hasPart(name string) bool {
	for _, v := range l.views {
		for _, p := range v.Parts {
			if p.Name == name {
				return true
			}
		}
	}
	return false
}

// Pose returns where every part of the view sits at time t (seconds into the
// animation), in the view's drawing order. Each motion moves the parts it
// names that the view has, or every part when it names none. The same t
// always gives the same pose.
func (a Animation) Pose(v View, t float64) []PartPose {
	poses := restPoses(v)
	a.addTo(poses, t, 1)
	return poses
}

// restPoses returns every part of the view where it is drawn, unmoved.
func restPoses(v View) []PartPose {
	poses := make([]PartPose, len(v.Parts))
	for i, p := range v.Parts {
		poses[i] = restPose(p)
	}
	return poses
}

// addTo adds the animation at time t, scaled by weight, to poses.
func (a Animation) addTo(poses []PartPose, t, weight float64) {
	for i := range poses {
		for _, m := range a.Motions {
			if m.moves(poses[i].Name) {
				m.applyTo(&poses[i], t, weight)
			}
		}
	}
}
