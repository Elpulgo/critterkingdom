//go:build js

package save

import (
	"errors"
	"syscall/js"
	"testing"
)

// newFakeStorage returns a JS object with getItem and setItem backed by the
// returned Go map. It stands in for the browser's localStorage, which Node
// does not have.
func newFakeStorage(t *testing.T) (js.Value, map[string]string) {
	t.Helper()
	items := map[string]string{}
	getItem := js.FuncOf(func(_ js.Value, args []js.Value) any {
		v, ok := items[args[0].String()]
		if !ok {
			return nil
		}
		return v
	})
	setItem := js.FuncOf(func(_ js.Value, args []js.Value) any {
		items[args[0].String()] = args[1].String()
		return nil
	})
	t.Cleanup(func() {
		getItem.Release()
		setItem.Release()
	})
	storage := js.Global().Get("Object").New()
	storage.Set("getItem", getItem)
	storage.Set("setItem", setItem)
	return storage, items
}

func TestLocalStorageStore_ASaveGameWrittenReadsBackIdenticalUnderTheGamesKey(t *testing.T) {
	storage, items := newFakeStorage(t)
	store := NewLocalStorageStore(storage)
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
	if _, ok := items["critterkingdom.save.slot1"]; !ok || len(items) != 1 {
		t.Errorf("stored keys = %v, want only critterkingdom.save.slot1", keysOf(items))
	}
}

func TestLocalStorageStore_AMissingSaveIsReportedAsNotFound(t *testing.T) {
	storage, _ := newFakeStorage(t)

	_, err := Read(NewLocalStorageStore(storage), "slot1")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Read(slot1) error = %v, want errors.Is(_, ErrNotFound)", err)
	}
}

func TestLocalStorageStore_ACorruptStoredSaveIsReportedAndLeftUnchanged(t *testing.T) {
	storage, items := newFakeStorage(t)
	const garbage = `{"version":2,"crit`
	items["critterkingdom.save.slot1"] = garbage

	_, err := Read(NewLocalStorageStore(storage), "slot1")

	if !errors.Is(err, ErrCorrupt) {
		t.Errorf("Read(slot1) error = %v, want errors.Is(_, ErrCorrupt)", err)
	}
	if after := items["critterkingdom.save.slot1"]; after != garbage {
		t.Errorf("stored save after Read = %q, want it unchanged as %q", after, garbage)
	}
}

func TestLocalStorageStore_KeysOutsideTheAllowedCharactersAreRejected(t *testing.T) {
	for _, key := range []string{"", "../x", "a/b", "Slot 1"} {
		t.Run(key, func(t *testing.T) {
			storage, items := newFakeStorage(t)
			store := NewLocalStorageStore(storage)

			saveErr := store.Save(key, []byte(`{"version":2}`))
			_, loadErr := store.Load(key)

			if !errors.Is(saveErr, ErrInvalidKey) {
				t.Errorf("Save(%q) error = %v, want errors.Is(_, ErrInvalidKey)", key, saveErr)
			}
			if !errors.Is(loadErr, ErrInvalidKey) {
				t.Errorf("Load(%q) error = %v, want errors.Is(_, ErrInvalidKey)", key, loadErr)
			}
			if len(items) != 0 {
				t.Errorf("stored keys = %v, want none", keysOf(items))
			}
		})
	}
}

func TestLocalStorageStore_AFullStorageIsReportedAsAnError(t *testing.T) {
	// Stands in for a browser localStorage whose quota is used up.
	full := js.Global().Call("eval", "({getItem(){return null}, setItem(){throw new Error('QuotaExceededError')}})")
	sg := SaveGame{Version: 2, Critter: Critter{Name: "Pip", Species: "fox", Colour: "red"}}

	err := Write(NewLocalStorageStore(full), "slot1", sg)

	if err == nil {
		t.Fatal("Write(slot1) to a full storage returned nil, want an error")
	}
}

func TestDefaultStore_WithoutAUsableLocalStorageReportsAnError(t *testing.T) {
	cases := []struct {
		name   string
		window string
	}{
		{name: "no window", window: "undefined"},
		{name: "localStorage access throws", window: "({get localStorage(){throw new Error('SecurityError')}})"},
		{name: "localStorage is null", window: "({localStorage: null})"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original := js.Global().Get("window")
			t.Cleanup(func() { js.Global().Set("window", original) })
			js.Global().Set("window", js.Global().Call("eval", tc.window))

			store, err := DefaultStore()

			if err == nil || store != nil {
				t.Errorf("DefaultStore() = (%v, %v), want (nil, an error)", store, err)
			}
		})
	}
}

func keysOf(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
