package truetype

import (
	"os"
	"testing"
)

func TestTrak(t *testing.T) {
	f, err := os.Open("testdata/ToyTrak.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	track, err := font.TableTrak()
	if err != nil {
		t.Fatal(err)
	}
	if len(track.Horizontal.Sizes) != 4 {
		t.Error()
	}
	if len(track.Vertical.Sizes) != 4 {
		t.Error()
	}
}
