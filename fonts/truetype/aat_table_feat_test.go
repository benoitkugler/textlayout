package truetype

import (
	"os"
	"testing"
)

func TestFeat(t *testing.T) {
	f, err := os.Open("testdata/ToyFeat.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	_, err = font.FeatTable()
	if err != nil {
		t.Fatal(err)
	}
}
