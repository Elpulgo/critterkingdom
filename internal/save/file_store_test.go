package save

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore_ASaveGameWrittenReadsBackIdentical(t *testing.T) {
	dir := t.TempDir()
	store := FileStore{Dir: dir}
	sg := SaveGame{Version: 2, Critter: Critter{Name: "Pip", Species: "fox", Colour: "red"}}

	if err := Write(store, "slot1", sg); err != nil {
		t.Fatalf("Write(slot1) unexpected error: %v", err)
	}
	got, err := Read(store, "slot1")
	if err != nil {
		t.Fatalf("Read(slot1) unexpected error: %v", err)
	}

	if got != sg {
		t.Errorf("Read(slot1) = %+v, want %+v", got, sg)
	}
	if _, err := os.Stat(filepath.Join(dir, "slot1.json")); err != nil {
		t.Errorf("save file slot1.json: %v, want it to exist", err)
	}
}

func TestFileStore_AMissingSaveIsReportedAsNotFound(t *testing.T) {
	store := FileStore{Dir: t.TempDir()}

	_, err := Read(store, "slot1")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Read(slot1) error = %v, want errors.Is(_, ErrNotFound)", err)
	}
}

func TestFileStore_ACorruptSaveFileIsReportedAndLeftUntouchedOnDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "slot1.json")
	garbage := []byte("{\"version\":2,\"crit\x00ter")
	if err := os.WriteFile(path, garbage, 0o644); err != nil {
		t.Fatalf("write corrupt save: %v", err)
	}

	_, err := Read(FileStore{Dir: dir}, "slot1")

	if !errors.Is(err, ErrCorrupt) {
		t.Errorf("Read(slot1) error = %v, want errors.Is(_, ErrCorrupt)", err)
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("corrupt save file after Read: %v, want it still there", readErr)
	}
	if !bytes.Equal(after, garbage) {
		t.Errorf("corrupt save file after Read = %q, want it unchanged as %q", after, garbage)
	}
}

func TestFileStore_SavingAgainReplacesTheEarlierSaveAndLeavesOnlyOneFile(t *testing.T) {
	// The save directory does not exist yet: Save must create it.
	dir := filepath.Join(t.TempDir(), "saves", "nested")
	store := FileStore{Dir: dir}
	first := SaveGame{Version: 2, Critter: Critter{Name: "Pip", Species: "fox", Colour: "red"}}
	second := SaveGame{Version: 2, Critter: Critter{Name: "Bramble", Species: "hedgehog", Colour: "brown"}}

	if err := Write(store, "slot1", first); err != nil {
		t.Fatalf("first Write(slot1) unexpected error: %v", err)
	}
	if err := Write(store, "slot1", second); err != nil {
		t.Fatalf("second Write(slot1) unexpected error: %v", err)
	}

	got, err := Read(store, "slot1")
	if err != nil {
		t.Fatalf("Read(slot1) unexpected error: %v", err)
	}
	if got != second {
		t.Errorf("Read(slot1) = %+v, want the second save %+v", got, second)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read save directory: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 1 || names[0] != "slot1.json" {
		t.Errorf("save directory holds %q, want only [slot1.json]", names)
	}
}

func TestFileStore_KeysThatCouldEscapeTheSaveDirectoryAreRejected(t *testing.T) {
	keys := []struct {
		name string
		key  string
	}{
		{"empty key", ""},
		{"parent directory", "../x"},
		{"subdirectory", "a/b"},
		{"capital letter and space", "Slot 1"},
	}
	data := []byte(`{"version":2,"critter":{"name":"Pip","species":"fox","colour":"red"}}`)

	for _, tt := range keys {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			store := FileStore{Dir: filepath.Join(root, "saves")}

			if err := store.Save(tt.key, data); !errors.Is(err, ErrInvalidKey) {
				t.Errorf("Save(%q) error = %v, want errors.Is(_, ErrInvalidKey)", tt.key, err)
			}
			if _, err := store.Load(tt.key); !errors.Is(err, ErrInvalidKey) {
				t.Errorf("Load(%q) error = %v, want errors.Is(_, ErrInvalidKey)", tt.key, err)
			}

			var created []string
			err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if path != root {
					created = append(created, path)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walk %s: %v", root, err)
			}
			if len(created) != 0 {
				t.Errorf("Save(%q) created %q, want nothing created", tt.key, created)
			}
		})
	}
}
