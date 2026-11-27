package save

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestEncodeAndDecode_RoundTripsASaveGameAtTheCurrentVersion(t *testing.T) {
	input := SaveGame{
		Version: 0,
		Critter: Critter{
			Name:    "Pip",
			Species: "fox",
			Colour:  "red",
		},
	}

	data, err := Encode(input)
	if err != nil {
		t.Fatalf("Encode(%+v) unexpected error: %v", input, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("encoded save is not valid JSON: %v", err)
	}
	if gotVersion, ok := raw["version"].(float64); !ok || gotVersion != 2 {
		t.Errorf("encoded top-level version = %v, want 2", raw["version"])
	}
	critterObj, ok := raw["critter"].(map[string]any)
	if !ok {
		t.Fatalf("encoded top-level critter = %T, want JSON object", raw["critter"])
	}
	if critterObj["name"] != "Pip" || critterObj["species"] != "fox" || critterObj["colour"] != "red" {
		t.Errorf("encoded critter object = %+v, want name=Pip species=fox colour=red", critterObj)
	}

	got, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode(%s) unexpected error: %v", data, err)
	}
	want := SaveGame{
		Version: CurrentVersion,
		Critter: Critter{
			Name:    "Pip",
			Species: "fox",
			Colour:  "red",
		},
	}
	if got != want {
		t.Errorf("Decode(Encode(sg)) = %+v, want %+v", got, want)
	}

	staleVersion := SaveGame{
		Version: 99,
		Critter: input.Critter,
	}
	staleData, err := Encode(staleVersion)
	if err != nil {
		t.Fatalf("Encode(%+v) unexpected error: %v", staleVersion, err)
	}
	gotFromStale, err := Decode(staleData)
	if err != nil {
		t.Fatalf("Decode(%s) unexpected error: %v", staleData, err)
	}
	if gotFromStale != want {
		t.Errorf("Decode(Encode(sg with Version 99)) = %+v, want %+v", gotFromStale, want)
	}
}

func TestDecode_ReportsCorruptInputsAsErrCorrupt(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty input", ``},
		{"truncated JSON", `{"version":2`},
		{"JSON null", `null`},
		{"JSON array instead of object", `[1,2,3]`},
		{"missing version field", `{"critter":{"name":"Pip","species":"fox","colour":"red"}}`},
		{"version is null", `{"version":null,"critter":{"name":"Pip","species":"fox","colour":"red"}}`},
		{"version is a string", `{"version":"2","critter":{"name":"Pip","species":"fox","colour":"red"}}`},
		{"version is a fractional number", `{"version":1.5,"critter":{"name":"Pip","species":"fox","colour":"red"}}`},
		{"version is zero", `{"version":0,"critter":{"name":"Pip","species":"fox","colour":"red"}}`},
		{"critter is a number instead of an object", `{"version":2,"critter":123}`},
		{"critter name is a number instead of a string", `{"version":2,"critter":{"name":42,"species":"fox","colour":"red"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode([]byte(tt.data))
			if !errors.Is(err, ErrCorrupt) {
				t.Errorf("Decode(%q) error = %v, want errors.Is(_, ErrCorrupt)", tt.data, err)
			}
			if errors.Is(err, ErrNewerVersion) {
				t.Errorf("Decode(%q) error = %v, unexpectedly matched ErrNewerVersion", tt.data, err)
			}
		})
	}
}

func TestDecode_ReportsASaveFromANewerVersionAsErrNewerVersionAndNotCorrupt(t *testing.T) {
	data := []byte(`{"version":3,"critter":{"name":"Pip","species":"fox","colour":"red"}}`)

	_, err := Decode(data)
	if !errors.Is(err, ErrNewerVersion) {
		t.Errorf("Decode(%s) error = %v, want errors.Is(_, ErrNewerVersion)", data, err)
	}
	if errors.Is(err, ErrCorrupt) {
		t.Errorf("Decode(%s) error = %v, must not match ErrCorrupt", data, err)
	}
}
