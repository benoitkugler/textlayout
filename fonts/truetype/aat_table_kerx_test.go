package truetype

import (
	"bytes"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestKerx(t *testing.T) {
	filename := "Bangla Sangam MN.ttc"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	fonts, err := NewFontParsers(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	for _, font := range fonts {
		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}
		_, err = font.KerxTable(ng)
		if err != nil {
			t.Fatal(err)
		}

	}
}

func TestKern0(t *testing.T) {
	data, err := testdata.Files.ReadFile("kernSubtable0.bin")
	if err != nil {
		t.Fatal(err)
	}
	kerx, err := parseKernxSubtable0(data, 8, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	expecteds := []struct { // value extracted from harfbuzz run
		left, right GID
		kerning     int16
	}{
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 104, 0},
		{104, 504, -42},
		{504, 1108, 0},
		{1108, 65535, 0},
		{65535, 552, 0},
		{552, 573, 0},
		{573, 1059, 0},
		{1059, 65535, 0},
	}

	for _, exp := range expecteds {
		got := kerx.KernPair(exp.left, exp.right)
		if got != exp.kerning {
			t.Fatalf("kerx6 - for (%d,%d), expected %d, got %d", exp.left, exp.right, exp.kerning, got)
		}
	}
}

func TestKerx6(t *testing.T) {
	data, err := testdata.Files.ReadFile("kerxSubtable6.bin")
	if err != nil {
		t.Fatal(err)
	}
	kerx, err := parseKerxSubtable6(data, 2775, 3)
	if err != nil {
		t.Fatal(err)
	}
	expecteds := []struct { // value extracted from harfbuzz run
		left, right GID
		kerning     int16
	}{
		{283, 659, -270},
		{659, 3, 0},
		{3, 4, 0},
		{4, 333, -130},
		{333, 3, 0},
		{3, 283, 0},
		{283, 815, -230},
		{815, 3, 0},
		{3, 333, 0},
		{333, 573, -150},
		{573, 3, 0},
		{3, 815, 0},
		{815, 283, -170},
		{283, 3, 0},
		{3, 659, 0},
		{659, 283, -270},
		{283, 3, 0},
		{3, 283, 0},
		{283, 650, -270},
	}

	for _, exp := range expecteds {
		got := kerx.KernPair(exp.left, exp.right)
		if got != exp.kerning {
			t.Fatalf("kerx6 - for (%d,%d), expected %d, got %d", exp.left, exp.right, exp.kerning, got)
		}
	}
}
