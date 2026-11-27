package combat

import (
	"testing"
	"time"
)

func TestHealth_PetalsNeverGoBelowZero(t *testing.T) {
	now := time.Duration(0)
	clock := func() time.Duration { return now } // fake for time, the only boundary here
	h := NewHealth(clock)

	if !h.Hit(5) {
		t.Fatalf("Hit(5) = false, want true")
	}
	if got := h.Petals(); got != 0 {
		t.Errorf("Petals() after overkill hit = %d, want 0", got)
	}
}

func TestHealth_AHitDuringInvulnerabilityDoesNothingUntilOneSecondPasses(t *testing.T) {
	now := 10 * time.Second
	clock := func() time.Duration { return now } // fake for time, the only boundary here
	h := NewHealth(clock)

	if !h.Hit(1) {
		t.Fatalf("first Hit(1) = false, want true")
	}
	if got := h.Petals(); got != 2 {
		t.Fatalf("Petals() = %d, want 2", got)
	}

	now += 999 * time.Millisecond
	if !h.Invulnerable() {
		t.Errorf("Invulnerable() at +999ms = false, want true")
	}
	if h.Hit(1) {
		t.Errorf("Hit(1) at +999ms = true, want false")
	}
	if got := h.Petals(); got != 2 {
		t.Errorf("Petals() at +999ms = %d, want 2", got)
	}

	now += time.Millisecond
	if h.Invulnerable() {
		t.Errorf("Invulnerable() at +1s = true, want false")
	}
	if !h.Hit(1) {
		t.Errorf("Hit(1) at +1s = false, want true")
	}
	if got := h.Petals(); got != 1 {
		t.Errorf("Petals() at +1s = %d, want 1", got)
	}
}

func TestHealth_ARejectedHitDoesNotStartInvulnerability(t *testing.T) {
	now := time.Duration(0)
	clock := func() time.Duration { return now } // fake for time, the only boundary here
	h := NewHealth(clock)

	for _, damage := range []int{0, -2} {
		if h.Hit(damage) {
			t.Errorf("Hit(%d) = true, want false", damage)
		}
		if h.Invulnerable() {
			t.Errorf("Invulnerable() after Hit(%d) = true, want false", damage)
		}
	}
	if !h.Hit(1) {
		t.Errorf("Hit(1) right after rejected hits = false, want true")
	}
	if got := h.Petals(); got != 2 {
		t.Errorf("Petals() = %d, want 2", got)
	}
}

func TestHealth_ReachingZeroPetalsStartsANapAndWakingRestoresTheCurrentMaximum(t *testing.T) {
	now := time.Duration(0)
	clock := func() time.Duration { return now } // fake for time, the only boundary here
	h := NewHealth(clock)
	h.GrowMax()

	if !h.Hit(3) {
		t.Fatalf("Hit(3) = false, want true")
	}
	if !h.Napping() {
		t.Fatalf("Napping() at 0 petals = false, want true")
	}

	now += 2 * time.Second
	if h.Hit(1) {
		t.Errorf("Hit(1) while napping = true, want false")
	}
	h.Heal(2)
	if got := h.Petals(); got != 0 {
		t.Errorf("Petals() while napping after Heal = %d, want 0", got)
	}

	h.Wake()
	if h.Napping() {
		t.Errorf("Napping() after Wake = true, want false")
	}
	if got := h.Petals(); got != 4 {
		t.Errorf("Petals() after Wake = %d, want 4", got)
	}
}

func TestHealth_HealingStopsAtTheMaximumAndTheMaximumGrowsFromThreeToFive(t *testing.T) {
	now := time.Duration(0)
	clock := func() time.Duration { return now } // fake for time, the only boundary here
	h := NewHealth(clock)

	if got := h.Petals(); got != 3 {
		t.Fatalf("initial Petals() = %d, want 3", got)
	}
	if got := h.Max(); got != 3 {
		t.Fatalf("initial Max() = %d, want 3", got)
	}

	h.Hit(2)
	h.Heal(0)
	h.Heal(-1)
	if got := h.Petals(); got != 1 {
		t.Errorf("Petals() after non-positive Heal = %d, want 1", got)
	}
	h.Heal(10)
	if got := h.Petals(); got != 3 {
		t.Errorf("Petals() after Heal(10) at max 3 = %d, want 3", got)
	}

	if !h.GrowMax() || h.Max() != 4 || h.Petals() != 3 {
		t.Errorf("first GrowMax(): max=%d petals=%d, want max=4 petals=3", h.Max(), h.Petals())
	}
	if !h.GrowMax() || h.Max() != 5 || h.Petals() != 3 {
		t.Errorf("second GrowMax(): max=%d petals=%d, want max=5 petals=3", h.Max(), h.Petals())
	}
	if h.GrowMax() || h.Max() != 5 {
		t.Errorf("third GrowMax() should fail at 5, got max=%d", h.Max())
	}

	h.Heal(10)
	if got := h.Petals(); got != 5 {
		t.Errorf("Petals() after Heal(10) at max 5 = %d, want 5", got)
	}
}
