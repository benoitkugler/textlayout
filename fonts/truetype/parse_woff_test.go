package truetype

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/rand"
	"testing"
)

func readWOFFHeaderReflect(r io.Reader) (woffHeader, error) {
	var header woffHeader
	err := binary.Read(r, binary.BigEndian, &header)
	return header, err
}

func readWOFFEntryReflect(r io.Reader) (woffEntry, error) {
	var entry woffEntry
	err := binary.Read(r, binary.BigEndian, &entry)
	return entry, err
}

func TestParseWOFFBinary(t *testing.T) {
	for range [200]int{} {
		var (
			header [woffHeaderSize]byte
			entry  [woffEntrySize]byte
		)
		rand.Read(header[:])
		rand.Read(entry[:])

		h1, err := readWOFFHeader(bytes.NewReader(header[:]))
		if err != nil {
			t.Fatal(err)
		}
		h2, err := readWOFFHeaderReflect(bytes.NewReader(header[:]))
		if err != nil {
			t.Fatal(err)
		}
		if h1 != h2 {
			t.Errorf("expected %v, got %v", h2, h1)
		}

		e1, err := readWOFFEntry(bytes.NewReader(entry[:]))
		if err != nil {
			t.Fatal(err)
		}
		e2, err := readWOFFEntryReflect(bytes.NewReader(entry[:]))
		if err != nil {
			t.Fatal(err)
		}
		if e1 != e2 {
			t.Errorf("expected %v, got %v", h2, h1)
		}
	}
}
