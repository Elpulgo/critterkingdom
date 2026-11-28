package entity_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/entity"
	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestPickInteractable_PicksTheNearestInFrontWithinReachAndNothingBehindBesideOrFar(t *testing.T) {
	// The critter's feet are at (100,100) in every case.
	pos := geom.Vec{X: 100, Y: 100}
	tests := []struct {
		name       string
		facing     entity.Facing
		candidates []geom.Vec
		want       int
		wantOK     bool
	}{
		{"facing down picks the one below", entity.FacingDown,
			[]geom.Vec{{X: 100, Y: 116}}, 0, true},
		{"facing up picks the one above", entity.FacingUp,
			[]geom.Vec{{X: 100, Y: 84}}, 0, true},
		{"facing left picks the one to the left", entity.FacingLeft,
			[]geom.Vec{{X: 84, Y: 100}}, 0, true},
		{"facing right picks the one to the right", entity.FacingRight,
			[]geom.Vec{{X: 116, Y: 100}}, 0, true},
		{"facing down ignores a nearer one behind", entity.FacingDown,
			[]geom.Vec{{X: 100, Y: 96}, {X: 100, Y: 120}}, 1, true},
		{"facing up ignores a nearer one behind", entity.FacingUp,
			[]geom.Vec{{X: 100, Y: 104}, {X: 100, Y: 80}}, 1, true},
		{"facing left ignores a nearer one behind", entity.FacingLeft,
			[]geom.Vec{{X: 104, Y: 100}, {X: 80, Y: 100}}, 1, true},
		{"facing right ignores a nearer one behind", entity.FacingRight,
			[]geom.Vec{{X: 96, Y: 100}, {X: 120, Y: 100}}, 1, true},
		{"only one behind picks nothing", entity.FacingDown,
			[]geom.Vec{{X: 100, Y: 90}}, 0, false},
		{"one straight to the side picks nothing", entity.FacingUp,
			[]geom.Vec{{X: 110, Y: 100}, {X: 90, Y: 100}}, 0, false},
		{"one just outside the cone picks nothing", entity.FacingRight,
			// 10 along the facing, 18 across: 10 is less than half of 20.59.
			[]geom.Vec{{X: 110, Y: 118}}, 0, false},
		{"one just inside the cone is picked", entity.FacingRight,
			// 10 along the facing, 17 across: 10 is more than half of 19.72.
			[]geom.Vec{{X: 110, Y: 117}}, 0, true},
		{"the nearest of two in front wins", entity.FacingLeft,
			[]geom.Vec{{X: 80, Y: 100}, {X: 88, Y: 104}}, 1, true},
		{"the lower index wins a tie", entity.FacingDown,
			[]geom.Vec{{X: 106, Y: 112}, {X: 94, Y: 112}}, 0, true},
		{"one exactly at reach is picked", entity.FacingDown,
			[]geom.Vec{{X: 100, Y: 124}}, 0, true},
		{"one beyond reach picks nothing", entity.FacingDown,
			[]geom.Vec{{X: 100, Y: 124.5}}, 0, false},
		{"one on the critter's feet picks nothing", entity.FacingDown,
			[]geom.Vec{{X: 100, Y: 100}}, 0, false},
		{"no candidates picks nothing", entity.FacingDown, nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := entity.PickInteractable(pos, tt.facing, tt.candidates)
			if ok != tt.wantOK || (ok && got != tt.want) {
				t.Errorf("PickInteractable(%v, %v, %v) = %d, %v; want %d, %v",
					pos, tt.facing, tt.candidates, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
