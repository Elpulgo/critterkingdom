package save

import (
	"encoding/json"
	"fmt"
)

// migrations maps a save version v to the step that upgrades the raw JSON
// object in place from v to v+1. Decode walks this table until the save
// reaches CurrentVersion, so a new version needs only a new step here.
var migrations = map[int]func(map[string]json.RawMessage) error{
	1: migrateV1,
}

// migrateV1 upgrades a version 1 save, which holds the critter's name,
// species and colour as flat top-level fields, to version 2, which nests
// them under "critter". An absent field becomes "", a non-string field is an
// error. The object is only changed once every field has been read.
func migrateV1(raw map[string]json.RawMessage) error {
	var critter Critter
	fields := []struct {
		key string
		dst *string
	}{
		{"name", &critter.Name},
		{"species", &critter.Species},
		{"colour", &critter.Colour},
	}
	for _, f := range fields {
		value, ok := raw[f.key]
		if !ok {
			continue
		}
		if err := json.Unmarshal(value, f.dst); err != nil || string(value) == "null" {
			return fmt.Errorf("version 1 field %q is not a string: %s", f.key, value)
		}
	}

	nested, err := json.Marshal(critter)
	if err != nil {
		return fmt.Errorf("marshal version 2 critter: %w", err)
	}
	for _, f := range fields {
		delete(raw, f.key)
	}
	raw["critter"] = nested
	return nil
}
