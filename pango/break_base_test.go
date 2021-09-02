package pango

import (
	"testing"
	"unicode"

	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

func TestResolveSentenceBreak(t *testing.T) {
	if unicode.Is(ucd.STerm, 46) {
		t.Fatalf("expected not STerm for rune 46")
	}

	sb := resolveSentenceBreakType(46, sb_Other, unicode.Po, ucd.BreakIS)
	if sb != sb_ATerm {
		t.Fatalf("expected %d, got %d", sb_ATerm, sb)
	}
}
