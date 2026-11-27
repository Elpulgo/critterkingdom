package critter_test

import (
	"strings"
	"testing"

	"github.com/Elpulgo/critterkingdom/internal/critter"
)

func TestLoadSpecies_RejectsInvalidSpeciesNamingTheCritter(t *testing.T) {
	tests := []struct {
		name string
		data string
		// wantErr lists substrings the error must contain: the critter it
		// names and the problem.
		wantErr []string
	}{
		{
			name: "duplicate id",
			data: `[
				{"id": "frog", "habitat": "pond", "likes": "berry", "rarity": "common"},
				{"id": "frog", "habitat": "meadow", "likes": "seed", "rarity": "rare"}
			]`,
			wantErr: []string{`"frog"`, "duplicate id"},
		},
		{
			name:    "unknown rarity",
			data:    `[{"id": "frog", "habitat": "pond", "likes": "berry", "rarity": "legendary"}]`,
			wantErr: []string{`"frog"`, "rarity", `"legendary"`},
		},
		{
			name:    "unknown activeAt",
			data:    `[{"id": "owl", "habitat": "forest", "likes": "seed", "rarity": "rare", "activeAt": ["dusk"]}]`,
			wantErr: []string{`"owl"`, "activeAt", `"dusk"`},
		},
		{
			name: "missing id",
			data: `[
				{"id": "frog", "habitat": "pond", "likes": "berry", "rarity": "common"},
				{"habitat": "pond", "likes": "berry", "rarity": "common"}
			]`,
			wantErr: []string{"critter 1", "missing id"},
		},
		{
			name:    "missing habitat",
			data:    `[{"id": "frog", "likes": "berry", "rarity": "common"}]`,
			wantErr: []string{`"frog"`, "missing habitat"},
		},
		{
			name:    "missing likes",
			data:    `[{"id": "frog", "habitat": "pond", "rarity": "common"}]`,
			wantErr: []string{`"frog"`, "missing likes"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			species, err := critter.LoadSpecies([]byte(tc.data))
			if err == nil {
				t.Fatalf("LoadSpecies succeeded with %d species; want error", len(species))
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("LoadSpecies error = %q; want it to contain %q", err, want)
				}
			}
		})
	}
}
