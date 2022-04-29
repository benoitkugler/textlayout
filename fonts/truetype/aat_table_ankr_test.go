package truetype

import (
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestParseTableAnkr(t *testing.T) {
	data, err := testdata.Files.ReadFile("ankr.bin")
	if err != nil {
		t.Fatal(err)
	}

	ankr, err := parseTableAnkr(data, 1409)
	if err != nil {
		t.Fatal(err)
	}
	expecteds := []struct {
		anchor AATAnchor
		glyph  GID
		index  int
	}{
		{AATAnchor{1043, 11}, 160, 1},
		{AATAnchor{0, 0}, 1324, 0},
	}
	for _, exp := range expecteds {
		got := ankr.GetAnchor(exp.glyph, exp.index)
		if got != exp.anchor {
			t.Fatalf("invalid anchor for (%d, %d): expected %v, got %v", exp.glyph, exp.index, exp.anchor, got)
		}
	}
}
