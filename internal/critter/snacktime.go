package critter

// MaxTries is how many snacks the player may offer in one Snack Time.
const MaxTries = 3

// Outcome is how a critter reacts to an offered snack.
type Outcome int

// The reactions to an offered snack.
const (
	// Giggled means the snack was wrong and the player may try again.
	Giggled Outcome = iota + 1
	// Befriended means the snack was the one the critter likes.
	Befriended
	// HoppedOff means the third wrong snack was offered; the critter leaves
	// and comes back later.
	HoppedOff
)

// SnackTime is one round of offering snacks to a critter. A round ends when
// the critter is befriended or hops off; after that every offer returns the
// same outcome and changes nothing.
type SnackTime struct {
	likes     string
	triesLeft int
	done      Outcome
}

// NewSnackTime starts a round of Snack Time with a critter of species sp.
func NewSnackTime(sp *Species) *SnackTime {
	return &SnackTime{likes: sp.Likes, triesLeft: MaxTries}
}

// Offer gives the critter the snack with the given item id and returns how
// it reacts.
func (s *SnackTime) Offer(snack string) Outcome {
	if s.done != 0 {
		return s.done
	}
	if snack == s.likes {
		s.done = Befriended
		return s.done
	}
	s.triesLeft--
	if s.triesLeft == 0 {
		s.done = HoppedOff
		return s.done
	}
	return Giggled
}

// TriesLeft returns how many wrong snacks the player may still offer before
// the critter hops off.
func (s *SnackTime) TriesLeft() int {
	return s.triesLeft
}
