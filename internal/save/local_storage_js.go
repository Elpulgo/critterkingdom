//go:build js

package save

import (
	"errors"
	"fmt"
	"syscall/js"
	"unicode/utf8"
)

// localStorageKeyPrefix namespaces the game's saves among other data on the
// same origin.
const localStorageKeyPrefix = "critterkingdom.save."

// LocalStorageStore keeps each save as a string under
// "critterkingdom.save.<key>" in a browser Storage object such as
// window.localStorage.
type LocalStorageStore struct {
	storage js.Value
}

// NewLocalStorageStore returns a store backed by storage, which must be a JS
// object with getItem and setItem methods like the browser's Storage.
func NewLocalStorageStore(storage js.Value) *LocalStorageStore {
	return &LocalStorageStore{storage: storage}
}

// Load reads the save stored under key. It returns an error wrapping
// ErrInvalidKey for a bad key and ErrNotFound when nothing is stored there.
func (s *LocalStorageStore) Load(key string) ([]byte, error) {
	if !validKey(key) {
		return nil, fmt.Errorf("load %q: %w", key, ErrInvalidKey)
	}
	v, err := callJS(s.storage, "getItem", localStorageKeyPrefix+key)
	if err != nil {
		return nil, fmt.Errorf("load %q: %w", key, err)
	}
	if v.IsNull() || v.IsUndefined() {
		return nil, fmt.Errorf("load %q: %w", key, ErrNotFound)
	}
	if v.Type() != js.TypeString {
		return nil, fmt.Errorf("load %q: stored value is a %s, not a string", key, v.Type())
	}
	return []byte(v.String()), nil
}

// Save stores data under key, replacing any earlier save. data must be valid
// UTF-8, because localStorage holds text. It returns an error wrapping
// ErrInvalidKey for a bad key, and an error when the storage refuses the
// write, for example because it is full.
func (s *LocalStorageStore) Save(key string, data []byte) error {
	if !validKey(key) {
		return fmt.Errorf("save %q: %w", key, ErrInvalidKey)
	}
	if !utf8.Valid(data) {
		return fmt.Errorf("save %q: data is not valid UTF-8 text", key)
	}
	if _, err := callJS(s.storage, "setItem", localStorageKeyPrefix+key, string(data)); err != nil {
		return fmt.Errorf("save %q: %w", key, err)
	}
	return nil
}

// callJS calls method on v and turns a JS exception, which syscall/js raises
// as a Go panic, into a returned error.
func callJS(v js.Value, method string, args ...any) (result js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoveredJSError(r)
		}
	}()
	return v.Call(method, args...), nil
}

// recoveredJSError converts a value recovered from a syscall/js panic into an
// error.
func recoveredJSError(r any) error {
	if e, ok := r.(error); ok {
		return e
	}
	return errors.New(fmt.Sprint(r))
}
