package main

import (
	"bytes"
	"io"

	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "bestpicture"
	pathkey := CASPathTransformFunc(key)
	expectedOriginalKey := "71056ad8aa24742ea41ea36fa2e3452a31636e82"
	expectedPathname := "71056/ad8aa/24742/ea41e/a36fa/2e345/2a316/36e82"
	if pathkey.Pathname != expectedPathname {
		t.Errorf("have %s want %s", pathkey.Pathname, expectedPathname)
	}
	if pathkey.Filename != expectedOriginalKey {
		t.Errorf("have %s want %s", pathkey.Filename, expectedOriginalKey)
	}

}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "goodpicture"
	data := []byte("Some jpg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)
	key := "goodpicture"
	data := []byte("Some jpg bytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if ok := s.Has(key); !ok {
		t.Errorf("expected to have key %s", key)
	}
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, err := io.ReadAll(r)

	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}
	// s.Delete(key)
}
