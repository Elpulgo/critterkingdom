package anim

import "math"

// BlendTime is how many seconds the animator takes to blend fully from idle
// to walk or back, so starting and stopping are soft.
const BlendTime = 0.1

// blendSnap is how close the walk weight must come to 0 or 1 to be taken as
// there, so floating-point steps like 1/60 s end the blend exactly.
const blendSnap = 1e-9

// Animator plays one character's idle and walk animations from frame times.
// It reads no clock: the same sequence of Update calls always gives the same
// poses.
type Animator struct {
	character Character
	idleTime  float64 // seconds of idle played
	walkTime  float64 // seconds of walking played
	weight    float64 // walk's share of the pose, from 0 (idle) to 1 (walk)
}

// NewAnimator returns an animator for the character, standing still at the
// start of its idle animation.
func NewAnimator(c Character) *Animator {
	return &Animator{character: c}
}

// Update advances the animations by dt seconds. Idle time always advances,
// walk time only while walking, and walk's weight moves linearly toward 1
// while walking and toward 0 while standing, taking BlendTime to cross.
func (a *Animator) Update(dt float64, walking bool) {
	a.idleTime += dt
	target := 0.0
	step := -dt / BlendTime
	if walking {
		a.walkTime += dt
		target = 1
		step = dt / BlendTime
	}
	a.weight = math.Min(1, math.Max(0, a.weight+step))
	if math.Abs(a.weight-target) < blendSnap {
		a.weight = target
	}
}

// Pose returns where every part of the named view sits now, in drawing order
// (a view the character lacks falls back to down). Idle motions are scaled by
// 1 minus walk's weight and walk motions by walk's weight; each blink plays
// only while its animation's weight is at least half. Mirroring is left to
// the caller.
func (a *Animator) Pose(view string) []PartPose {
	poses := restPoses(a.character.Layout.View(view))
	a.character.Idle.addTo(poses, a.idleTime, 1-a.weight)
	a.character.Walk.addTo(poses, a.walkTime, a.weight)
	return poses
}
