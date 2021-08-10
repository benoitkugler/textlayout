package truetype

import (
	"os"
	"reflect"
	"testing"
)

func TestTrak(t *testing.T) {
	f, err := os.Open("testdata/ToyTrak.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(f, false)
	if err != nil {
		t.Fatal(err)
	}

	track, err := font.TrakTable()
	if err != nil {
		t.Fatal(err)
	}
	if len(track.Horizontal.Sizes) != 4 {
		t.Error()
	}
	if len(track.Vertical.Sizes) != 4 {
		t.Error()
	}

	exp := TrakData{
		Entries: []TrackEntry{
			{
				PerSizeTracking: []int16{200, 200, 0, -100},
				Track:           0,
				NameIndex:       2,
			},
		},
		Sizes: []float32{1, 2, 12, 96},
	}
	if got := track.Horizontal; !reflect.DeepEqual(got, exp) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}
