//go:build js

package save

import (
	"errors"
	"fmt"
	"syscall/js"
)

// DefaultStore returns the store the game saves to: the browser's
// window.localStorage. It returns an error, rather than panicking, when
// there is no window or localStorage is missing or blocked (for example in
// some private browsing modes).
func DefaultStore() (Store, error) {
	storage, err := browserLocalStorage()
	if err != nil {
		return nil, fmt.Errorf("default save store: %w", err)
	}
	return NewLocalStorageStore(storage), nil
}

// browserLocalStorage reads window.localStorage, turning a thrown
// SecurityError or a missing window into an error.
func browserLocalStorage() (storage js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoveredJSError(r)
		}
	}()
	window := js.Global().Get("window")
	if window.Type() != js.TypeObject {
		return js.Value{}, errors.New("no browser window")
	}
	// js.Value.Get does not catch a throwing getter and would crash the
	// runtime; Reflect.get through Call turns the throw into a panic that
	// is recovered above.
	storage = js.Global().Get("Reflect").Call("get", window, "localStorage")
	if storage.Type() != js.TypeObject {
		return js.Value{}, errors.New("window.localStorage is unavailable")
	}
	return storage, nil
}
