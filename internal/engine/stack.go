package engine

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

// Overlay is a scene shown over another, such as a dialogue box. While it is
// on top of a Stack it gets all the input and the scenes underneath are
// paused.
type Overlay interface {
	Scene
	// Done reports whether the overlay has finished. The stack removes it
	// right after the Update in which it first reports true.
	Done() bool
}

// Stack is a scene made of a base scene, such as the world, with overlays
// pushed on top of it. Every scene in the stack is drawn, bottom first, but
// only the top one is updated, so an overlay pauses the scenes underneath and
// takes all their input.
type Stack struct {
	base     Scene
	overlays []Overlay
	// awaitRelease hides Pointing from the top scene until the pointer has
	// been let go, after an overlay closed, so a press that closed it does
	// not carry on into the scene underneath.
	awaitRelease bool
}

// NewStack returns a stack with base at the bottom and no overlays.
func NewStack(base Scene) *Stack {
	return &Stack{base: base}
}

// Push puts o on top of the stack. From the next Update on it gets the
// input and the scenes under it are paused.
func (s *Stack) Push(o Overlay) {
	s.overlays = append(s.overlays, o)
}

// Overlaid reports whether an overlay is on the stack. The Game asks for it
// to take the touch controls away while an overlay is up.
func (s *Stack) Overlaid() bool {
	return len(s.overlays) > 0
}

// Update updates the top scene only. When the top is an overlay that reports
// Done after its update, it is removed, and the scene underneath sees no
// pointer until the pointer has been released.
func (s *Stack) Update(f Frame) error {
	if s.awaitRelease {
		if f.Input.Pointing {
			f.Input.Pointing = false
			f.Input.Point = geom.Vec{}
		} else {
			s.awaitRelease = false
		}
	}
	n := len(s.overlays)
	if n == 0 {
		return s.base.Update(f)
	}
	top := s.overlays[n-1]
	if err := top.Update(f); err != nil {
		return err
	}
	if top.Done() {
		// The overlay may have pushed another on top of itself; remove only
		// the one that finished.
		s.overlays = slices.Delete(s.overlays, n-1, n)
		s.awaitRelease = true
	}
	return nil
}

// DrawWorld draws the world of the base scene and then of each overlay, in
// the order they were pushed.
func (s *Stack) DrawWorld(c Canvas) {
	s.base.DrawWorld(c)
	for _, o := range s.overlays {
		o.DrawWorld(c)
	}
}

// DrawUI draws the interface of the base scene and then of each overlay, in
// the order they were pushed.
func (s *Stack) DrawUI(screen *ebiten.Image, ui UI) {
	s.base.DrawUI(screen, ui)
	for _, o := range s.overlays {
		o.DrawUI(screen, ui)
	}
}

// DebugLines returns the base scene's debug lines, if it has any.
func (s *Stack) DebugLines() []string {
	if d, ok := s.base.(debugScene); ok {
		return d.DebugLines()
	}
	return nil
}

// DrawDebug lets the base scene draw its debug marks, if it has any.
func (s *Stack) DrawDebug(c Canvas) {
	if d, ok := s.base.(debugScene); ok {
		d.DrawDebug(c)
	}
}
