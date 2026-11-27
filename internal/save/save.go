// Package save encodes, migrates and stores versioned game saves.
package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// CurrentVersion is the save format version written by Encode.
const CurrentVersion = 2

var (
	// ErrCorrupt reports that a save could not be decoded or failed validation.
	ErrCorrupt = errors.New("save: corrupt save data")
	// ErrNewerVersion reports that a save was written by a newer version of the game.
	ErrNewerVersion = errors.New("save: save is from a newer version")
)

// Critter holds the player critter's chosen identity.
type Critter struct {
	Name    string `json:"name"`
	Species string `json:"species"`
	Colour  string `json:"colour"`
}

// SaveGame is the persisted state of a playthrough at CurrentVersion.
type SaveGame struct {
	Version int     `json:"version"`
	Critter Critter `json:"critter"`
}

// Encode serialises sg as JSON with Version set to CurrentVersion.
func Encode(sg SaveGame) ([]byte, error) {
	sg.Version = CurrentVersion
	return json.Marshal(sg)
}

// Decode parses a JSON save, migrating older versions up to CurrentVersion.
// It returns an error wrapping ErrCorrupt when the data is not a valid save,
// or wrapping ErrNewerVersion when the save's version is above CurrentVersion.
func Decode(data []byte) (SaveGame, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return SaveGame{}, fmt.Errorf("unmarshal save object: %w (%w)", ErrCorrupt, err)
	}
	if raw == nil {
		return SaveGame{}, fmt.Errorf("save is null: %w", ErrCorrupt)
	}

	rawVersion, ok := raw["version"]
	if !ok || len(rawVersion) == 0 || string(rawVersion) == "null" {
		return SaveGame{}, fmt.Errorf("missing version: %w", ErrCorrupt)
	}
	var version int
	if err := json.Unmarshal(rawVersion, &version); err != nil {
		return SaveGame{}, fmt.Errorf("invalid version %s: %w (%w)", string(rawVersion), ErrCorrupt, err)
	}
	if version < 1 {
		return SaveGame{}, fmt.Errorf("version %d below 1: %w", version, ErrCorrupt)
	}
	if version > CurrentVersion {
		return SaveGame{}, fmt.Errorf("version %d > current %d: %w", version, CurrentVersion, ErrNewerVersion)
	}

	for version < CurrentVersion {
		step, ok := migrations[version]
		if !ok {
			return SaveGame{}, fmt.Errorf("no migration from version %d: %w", version, ErrCorrupt)
		}
		if err := step(raw); err != nil {
			return SaveGame{}, fmt.Errorf("migrate from version %d: %w (%w)", version, ErrCorrupt, err)
		}
		version++
		raw["version"] = json.RawMessage(strconv.Itoa(version))
	}

	migrated, err := json.Marshal(raw)
	if err != nil {
		return SaveGame{}, fmt.Errorf("marshal migrated save: %w (%w)", ErrCorrupt, err)
	}
	var sg SaveGame
	if err := json.Unmarshal(migrated, &sg); err != nil {
		return SaveGame{}, fmt.Errorf("unmarshal save game: %w (%w)", ErrCorrupt, err)
	}
	return sg, nil
}
