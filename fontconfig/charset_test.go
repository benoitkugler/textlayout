package fontconfig

import (
	"testing"
)

func TestCharset(t *testing.T) {
	var cs1, cs2, cs3 Charset

	for i := 100; i < 20000; i += 100 {
		cs1.AddChar(rune(i))
	}

	if c := cs1.count(); c != 199 {
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

	if !FcCharsetEqual(cs3, cs4) {
		t.Errorf("wrong union, got %v", cs3)
	}

	if cs5 := charsetSubtract(cs4, cs2); !FcCharsetEqual(cs5, cs1) {
		t.Errorf("wrong difference, got %v", cs5)
	}

	if count := charsetSubtractCount(cs4, cs2); count != 199 {
		t.Errorf("expected 199, got %d", count)
	}

	if cs5 := charsetSubtract(cs4, cs1); !FcCharsetEqual(cs5, cs2) {
		t.Errorf("wrong difference, got %v", cs5)
	}

	if cs5 := charsetSubtract(cs4, cs4); !FcCharsetEqual(cs5, Charset{}) {
		t.Errorf("wrong difference, got %v", cs5)
	}

	if cs5 := charsetSubtract(cs2, cs4); !FcCharsetEqual(cs5, Charset{}) {
		t.Errorf("wrong difference, got %v", cs5)
	}

}
