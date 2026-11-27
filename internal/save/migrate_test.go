package save

import (
	"errors"
	"testing"
)

func TestDecode_LoadsAVersion1SaveAsACurrentSaveKeepingTheCritter(t *testing.T) {
	data := []byte(`{"version":1,"name":"Pip","species":"fox","colour":"red"}`)

	got, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode(%s) unexpected error: %v", data, err)
	}
	want := SaveGame{Version: 2, Critter: Critter{Name: "Pip", Species: "fox", Colour: "red"}}
	if got != want {
		t.Errorf("Decode(%s) = %+v, want %+v", data, got, want)
	}
}

func TestDecode_ReportsABrokenVersion1SaveAsCorruptNotHalfMigrated(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"name is a number", `{"version":1,"name":42,"species":"fox","colour":"red"}`},
		{"species is an object", `{"version":1,"name":"Pip","species":{"kind":"fox"},"colour":"red"}`},
		{"colour is null", `{"version":1,"name":"Pip","species":"fox","colour":null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode([]byte(tt.data))
			if !errors.Is(err, ErrCorrupt) {
				t.Errorf("Decode(%s) error = %v, want errors.Is(_, ErrCorrupt)", tt.data, err)
			}
			if got != (SaveGame{}) {
				t.Errorf("Decode(%s) = %+v, want the zero SaveGame alongside the error", tt.data, got)
			}
		})
	}
}
