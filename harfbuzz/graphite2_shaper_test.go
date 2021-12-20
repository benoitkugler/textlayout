package harfbuzz

import (
	"math"
	"testing"

	tt "github.com/benoitkugler/textlayout/fonts/truetype"
)

func TestTagFromString(t *testing.T) {
	assertEqualTag(t, tt.NewTag('a', 'B', 'c', 'D'), 0x61426344)

	assertEqualTag(t, tagFromString("aBcDe"), 0x61426344)
	assertEqualTag(t, tagFromString("aBcD"), 0x61426344)
	assertEqualTag(t, tagFromString("aBc"), 0x61426320)
	assertEqualTag(t, tagFromString("aB"), 0x61422020)
	assertEqualTag(t, tagFromString("a"), 0x61202020)
	assertEqualTag(t, tagFromString("aBcDe"[:1]), 0x61202020)
	assertEqualTag(t, tagFromString("aBcDe"[:2]), 0x61422020)
	assertEqualTag(t, tagFromString("aBcDe"[:3]), 0x61426320)
	assertEqualTag(t, tagFromString("aBcDe"[:4]), 0x61426344)
	assertEqualTag(t, tagFromString("aBcDe"[:4]), 0x61426344)

	assertEqualTag(t, tagFromString(""), 0)
}

func TestGlyphStoreInt32(t *testing.T) {
	var gl GlyphInfo
	for i := int32(math.MinInt32); i <= int32(math.MaxInt32)-1000; i += 100 {
		gl.setInt32(i)
		if gl.getInt32() != i {
			t.Fatalf("failed to store int32: %d", i)
		}
	}
}
