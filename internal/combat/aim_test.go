package combat

import (
	"math"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/geom"
)

func TestAim_TurnsTowardAnEnemyInsideTheConeAndKeepsFacingOtherwise(t *testing.T) {
	origin := geom.Vec{X: 100, Y: 100}
	facingRight := geom.Vec{X: 1, Y: 0}
	facingDiag := geom.Vec{X: 4, Y: 3}

	rad44 := 44 * math.Pi / 180
	rad46 := 46 * math.Pi / 180
	cos44, sin44 := math.Cos(rad44), math.Sin(rad44)
	cos46, sin46 := math.Cos(rad46), math.Sin(rad46)

	tests := []struct {
		name    string
		facing  geom.Vec
		enemies []geom.Vec
		want    geom.Vec
		wantLen float64
	}{
		{
			"no enemies at all keeps the normalized facing",
			facingDiag,
			nil,
			geom.Vec{X: 0.8, Y: 0.6},
			1,
		},
		{
			"zero facing has no cone and returns the zero vector",
			geom.Vec{},
			[]geom.Vec{{X: 120, Y: 100}},
			geom.Vec{},
			0,
		},
		{
			"enemy straight ahead on an axis at the 40 px range limit is aimed at",
			facingDiag,
			[]geom.Vec{{X: 140, Y: 100}},
			geom.Vec{X: 1, Y: 0},
			1,
		},
		{
			"enemy straight ahead on an axis just beyond the range limit at 41 px is ignored",
			facingDiag,
			[]geom.Vec{{X: 141, Y: 100}},
			geom.Vec{X: 0.8, Y: 0.6},
			1,
		},
		{
			"enemy 44 degrees to one side of facing is aimed at",
			facingRight,
			[]geom.Vec{{X: 100 + 20*cos44, Y: 100 + 20*sin44}},
			geom.Vec{X: cos44, Y: sin44},
			1,
		},
		{
			"enemy 44 degrees to the other side of facing is aimed at",
			facingRight,
			[]geom.Vec{{X: 100 + 20*cos44, Y: 100 - 20*sin44}},
			geom.Vec{X: cos44, Y: -sin44},
			1,
		},
		{
			"enemy 46 degrees to one side of facing is ignored",
			facingRight,
			[]geom.Vec{{X: 100 + 20*cos46, Y: 100 + 20*sin46}},
			geom.Vec{X: 1, Y: 0},
			1,
		},
		{
			"enemy 46 degrees to the other side of facing is ignored",
			facingRight,
			[]geom.Vec{{X: 100 + 20*cos46, Y: 100 - 20*sin46}},
			geom.Vec{X: 1, Y: 0},
			1,
		},
		{
			"enemy directly behind the player is ignored",
			facingRight,
			[]geom.Vec{{X: 80, Y: 100}},
			geom.Vec{X: 1, Y: 0},
			1,
		},
		{
			"enemy on top of the player is ignored",
			facingRight,
			[]geom.Vec{{X: 100, Y: 100}},
			geom.Vec{X: 1, Y: 0},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Aim(origin, tt.facing, tt.enemies)
			if math.Abs(got.X-tt.want.X) > 1e-9 || math.Abs(got.Y-tt.want.Y) > 1e-9 {
				t.Errorf("Aim(%v, %v, %v) = %v, want %v", origin, tt.facing, tt.enemies, got, tt.want)
			}
			if math.Abs(got.Len()-tt.wantLen) > 1e-9 {
				t.Errorf("Aim(%v, %v, %v).Len() = %v, want %v", origin, tt.facing, tt.enemies, got.Len(), tt.wantLen)
			}
		})
	}
}

func TestAim_PicksTheNearestEnemyInTheConeAndIgnoresNearerOnesOutside(t *testing.T) {
	origin := geom.Vec{X: 0, Y: 0}
	facing := geom.Vec{X: 1, Y: 0}
	enemies := []geom.Vec{
		{X: -5, Y: 0},   // behind: closer (dist 5), but outside the cone
		{X: 0, Y: 10},   // 90 degrees off: closer (dist 10), but outside the cone
		{X: 30, Y: 0},   // straight ahead in the cone (dist 30)
		{X: 16, Y: 12},  // in the cone and nearer (dist 20, 3-4-5 triangle)
		{X: 16, Y: -12}, // in the cone at the same distance (dist 20); first tied enemy wins
	}

	got := Aim(origin, facing, enemies)
	want := geom.Vec{X: 0.8, Y: 0.6}
	if math.Abs(got.X-want.X) > 1e-9 || math.Abs(got.Y-want.Y) > 1e-9 {
		t.Errorf("Aim = %v, want %v", got, want)
	}
	if math.Abs(got.Len()-1) > 1e-9 {
		t.Errorf("Aim.Len() = %v, want 1", got.Len())
	}
}
