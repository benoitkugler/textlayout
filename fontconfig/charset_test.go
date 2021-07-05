package fontconfig

import (
	"fmt"
	"testing"
	"time"
)

func TestCharset(t *testing.T) {
	var cs1, cs2, cs3 Charset

	for i := 100; i < 20000; i += 100 {
		cs1.AddChar(rune(i))
	}

	if c := cs1.Len(); c != 199 {
		t.Errorf("expected 199, got %d", c)
	}

	for i := 5; i < 90; i += 5 {
		cs2.AddChar(rune(i))
	}

	cs4 := charsetUnion(cs1, cs2)

	for i := 100; i < 20000; i += 100 {
		cs3.AddChar(rune(i))
	}
	for i := 5; i < 90; i += 5 {
		cs3.AddChar(rune(i))
	}

	if !charsetEqual(cs3, cs4) {
		t.Errorf("wrong union, got %v", cs3)
	}

	if cs5 := charsetSubtract(cs4, cs2); !charsetEqual(cs5, cs1) {
		t.Errorf("wrong difference, got %v", cs5)
	}

	if count := charsetSubtractCount(cs4, cs2); count != 199 {
		t.Errorf("expected 199, got %d", count)
	}

	if cs5 := charsetSubtract(cs4, cs1); !charsetEqual(cs5, cs2) {
		t.Errorf("wrong difference, got %v", cs5)
	}

	if cs5 := charsetSubtract(cs4, cs4); !charsetEqual(cs5, Charset{}) {
		t.Errorf("wrong difference, got %v", cs5)
	}

	if cs5 := charsetSubtract(cs2, cs4); !charsetEqual(cs5, Charset{}) {
		t.Errorf("wrong difference, got %v", cs5)
	}
}

func TestCharsetHash(t *testing.T) {
	var cs Charset

	for i := 100; i < 20000; i += 100 {
		cs.AddChar(rune(i))
	}
	for i := 0xFFFF; i < 0x2FFFF; i += 1 {
		cs.AddChar(rune(i))
	}

	fmt.Println(cs.Len())
	fmt.Println(len(cs.hash()))
	fmt.Println(len(fmt.Sprintf("%v", cs)))
}

func TestCharsetMerge(t *testing.T) {
	var cs1, cs2 Charset

	for i := 100; i < 20000; i += 100 {
		cs1.AddChar(rune(i))
	}
	for i := 5; i < 90; i += 5 {
		cs2.AddChar(rune(i))
	}
	L1, L2 := cs1.Len(), cs2.Len()

	cs3 := cs1.Copy()
	added := cs3.merge(cs2)
	if !added {
		t.Fatal("expected added")
	}

	if cs3.Len() != L1+L2 {
		t.Fatalf("invalid length, expected %d, got %d", L1+L2, cs3.Len())
	}

	for i := 100; i < 20000; i += 100 {
		if !cs3.HasChar(rune(i)) {
			t.Fatalf("missing rune %d", i)
		}
	}
	for i := 5; i < 90; i += 5 {
		if !cs3.HasChar(rune(i)) {
			t.Fatalf("missing rune %d", i)
		}
	}
}

func TestMergeMany(t *testing.T) {
	fs := cachedFS()
	var charsets []Charset
	for _, f := range fs {
		cs, ok := f.GetCharset(CHARSET)
		if !ok {
			continue
		}
		charsets = append(charsets, cs)
	}
	var cs Charset
	ti := time.Now()
	for _, c := range charsets {
		cs.merge(c)
	}
	fmt.Printf("Merging sequentially %d charsets: %s\n", len(charsets), time.Since(ti))
}
