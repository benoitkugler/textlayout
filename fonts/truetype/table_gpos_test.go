package truetype

import (
	"bytes"
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestBits(t *testing.T) {
	var buf8 [8]int8
	uint16As2Bits(buf8[:], 0x123F)
	if exp := [8]int8{0, 1, 0, -2, 0, -1, -1, -1}; buf8 != exp {
		t.Fatalf("expected %v, got %v", exp, buf8)
	}

	var buf4 [4]int8
	uint16As4Bits(buf4[:], 0x123F)
	if exp := [4]int8{1, 2, 3, -1}; buf4 != exp {
		t.Fatalf("expected %v, got %v", exp, buf4)
	}

	var buf2 [2]int8
	uint16As8Bits(buf2[:], 0x123F)
	if exp := [2]int8{18, 63}; buf2 != exp {
		t.Fatalf("expected %v, got %v", exp, buf2)
	}
}

func TestGPOS(t *testing.T) {
	filenames := []string{
		"Raleway-v4020-Regular.otf",
		"Estedad-VF.ttf",
		"Mada-VF.ttf",
	}

	filenames = append(filenames, dirFiles(t, "layout_fonts/gpos")...)

	for _, filename := range filenames {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		gpos, err := font.GPOSTable()
		if err != nil {
			t.Fatal(filename, err)
		}
		for _, l := range gpos.Lookups {
			for _, s := range l.Subtables {
				if pair1, ok := s.Data.(GPOSPair1); ok {
					for _, set := range pair1.Values {
						for _, v := range set {
							if set.FindGlyph(v.SecondGlyph) == nil {
								t.Fatal("invalid binary search")
							}
						}
					}
				}
			}
		}
	}
}

func TestGPOSCursive1(t *testing.T) {
	filename := "ToyGPOSCursive.ttf"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := NewFontParser(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	gpos, err := font.GPOSTable()
	if err != nil {
		t.Fatal(filename, err)
	}

	if len(gpos.Lookups) != 4 || len(gpos.Lookups[0].Subtables) != 1 {
		t.Fatalf("invalid gpos lookups: %v", gpos.Lookups)
	}
	cursive, ok := gpos.Lookups[0].Subtables[0].Data.(GPOSCursive1)
	if !ok {
		t.Fatalf("unexpected type for lookup %T", gpos.Lookups[0].Subtables[0].Data)
	}

	expected := GPOSCursive1{
		[2]GPOSAnchor{GPOSAnchorFormat1{405, 45}, GPOSAnchorFormat1{0, 0}},
		[2]GPOSAnchor{GPOSAnchorFormat1{452, 500}, GPOSAnchorFormat1{0, 0}},
	}
	if !reflect.DeepEqual(expected, cursive) {
		t.Fatalf("expected %v, got %v", expected, cursive)
	}
}
