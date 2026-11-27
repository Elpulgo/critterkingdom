package critter_test

import (
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/critter"
)

func TestSnackTime_RightSnackBefriendsAndThreeWrongMakeItHopOff(t *testing.T) {
	firefly := critter.Species{ID: "firefly", Likes: "berry"}

	type offer struct {
		snack     string
		want      critter.Outcome
		triesLeft int
	}
	tests := []struct {
		name   string
		offers []offer
	}{
		{
			name:   "right snack first befriends",
			offers: []offer{{"berry", critter.Befriended, 3}},
		},
		{
			name: "wrong snack giggles, then right snack befriends",
			offers: []offer{
				{"acorn", critter.Giggled, 2},
				{"berry", critter.Befriended, 2},
			},
		},
		{
			name: "right snack on the last try still befriends",
			offers: []offer{
				{"acorn", critter.Giggled, 2},
				{"carrot", critter.Giggled, 1},
				{"berry", critter.Befriended, 1},
			},
		},
		{
			name: "third wrong snack makes it hop off",
			offers: []offer{
				{"acorn", critter.Giggled, 2},
				{"carrot", critter.Giggled, 1},
				{"acorn", critter.HoppedOff, 0},
			},
		},
		{
			name: "right snack after it hopped off is too late",
			offers: []offer{
				{"acorn", critter.Giggled, 2},
				{"carrot", critter.Giggled, 1},
				{"acorn", critter.HoppedOff, 0},
				{"berry", critter.HoppedOff, 0},
			},
		},
		{
			name: "wrong snack after befriending changes nothing",
			offers: []offer{
				{"berry", critter.Befriended, 3},
				{"acorn", critter.Befriended, 3},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			round := critter.NewSnackTime(&firefly)
			for i, o := range tc.offers {
				if got := round.Offer(o.snack); got != o.want {
					t.Fatalf("offer %d (%s): outcome %d, want %d", i+1, o.snack, got, o.want)
				}
				if got := round.TriesLeft(); got != o.triesLeft {
					t.Fatalf("offer %d (%s): %d tries left, want %d", i+1, o.snack, got, o.triesLeft)
				}
			}
		})
	}
}
