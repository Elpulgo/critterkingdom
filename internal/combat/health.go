package combat

import "time"

const (
	// StartPetals is how many petals a new critter begins with.
	StartPetals = 3
	// MaxPetals is the most petals a critter can ever grow to hold.
	MaxPetals = 5
	// InvulnerableFor is how long a critter cannot be hurt again after a hit.
	InvulnerableFor = time.Second
)

// Clock returns the game time elapsed since the adventure began.
type Clock func() time.Duration

// Health tracks the critter's petals, its one-second invulnerability window
// after taking a hit, and the nap when its petals run out.
type Health struct {
	petals  int
	max     int
	clock   Clock
	hit     bool
	lastHit time.Duration
	napping bool
}

// NewHealth returns a Health starting at StartPetals and reading game time
// from clock.
func NewHealth(clock Clock) *Health {
	return &Health{
		petals: StartPetals,
		max:    StartPetals,
		clock:  clock,
	}
}

// Petals returns how many petals the critter currently has.
func (h *Health) Petals() int {
	return h.petals
}

// Max returns the maximum number of petals the critter can currently hold.
func (h *Health) Max() int {
	return h.max
}

// Invulnerable reports whether the critter is still within the one-second
// invulnerability window from its last hit.
func (h *Health) Invulnerable() bool {
	return h.hit && h.clock()-h.lastHit < InvulnerableFor
}

// Napping reports whether the critter is taking a nap after losing all its
// petals.
func (h *Health) Napping() bool {
	return h.napping
}

// Hit removes damage petals and starts the invulnerability window, clamping
// petals at zero and starting a nap if none remain. It does nothing and
// returns false when damage is not positive, while invulnerable, or during a
// nap.
func (h *Health) Hit(damage int) bool {
	if damage <= 0 || h.napping || h.Invulnerable() {
		return false
	}
	h.petals -= damage
	if h.petals <= 0 {
		h.petals = 0
		h.napping = true
	}
	h.lastHit = h.clock()
	h.hit = true
	return true
}

// Heal restores n petals, up to Max. It does nothing when n is not positive or
// while the critter is napping.
func (h *Health) Heal(n int) {
	if n <= 0 || h.napping {
		return
	}
	h.petals += n
	if h.petals > h.max {
		h.petals = h.max
	}
}

// GrowMax adds one to the petal maximum, up to MaxPetals, without changing the
// current petal count. It returns false when already at MaxPetals.
func (h *Health) GrowMax() bool {
	if h.max >= MaxPetals {
		return false
	}
	h.max++
	return true
}

// Wake ends a nap and restores petals to the current maximum.
func (h *Health) Wake() {
	h.petals = h.max
	h.napping = false
}
