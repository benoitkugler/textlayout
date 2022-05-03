package truetype

import (
	"bytes"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestFeat(t *testing.T) {
	f, err := testdata.Files.ReadFile("ToyFeat.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := NewFontParser(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}
	_, err = font.FeatTable()
	if err != nil {
		t.Fatal(err)
	}
}
