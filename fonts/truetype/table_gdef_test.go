package truetype

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestParseGdef(t *testing.T) {
	filename := "Commissioner-VF.ttf"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := NewFontParser(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	names, err := font.tryAndLoadNameTable()
	if err != nil {
		t.Fatal(err)
	}

	fvar, err := font.tryAndLoadFvarTable(names)
	if err != nil {
		t.Fatal(err)
	}

	gdef, err := font.GDEFTable(len(fvar.Axis))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gdef.Class.GlyphSize())
	fmt.Println(len(gdef.VariationStore.Regions))
	fmt.Println(len(gdef.VariationStore.Datas))
}

func TestParseGDEFCaretList(t *testing.T) {
	filename := "GDEFCaretList3.ttf"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := NewFontParser(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	names, err := font.tryAndLoadNameTable()
	if err != nil {
		t.Fatal(err)
	}

	fvar, err := font.tryAndLoadFvarTable(names)
	if err != nil {
		t.Fatal(err)
	}

	gdef, err := font.GDEFTable(len(fvar.Axis))
	if err != nil {
		t.Fatal(err)
	}

	// reference values are taken from fonttools
	coverage := []GID{380, 381, 382, 383}
	if s := gdef.LigatureCaretList.Coverage.Size(); s != 4 {
		t.Fatalf("expected 4, got %d", s)
	}
	for i, g := range coverage {
		v, _ := gdef.LigatureCaretList.Coverage.Index(g)
		if v != i {
			t.Fatalf("expected %d, got %d", i, v)
		}
	}

	expectedLigGlyphs := [][]CaretValue{ //  LigGlyphCount=4
		// CaretCount=1
		{
			CaretValueFormat3{Coordinate: 620, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 205}},
		},
		// CaretCount=1
		{
			CaretValueFormat3{Coordinate: 675, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 193}},
		},
		// CaretCount=2
		{
			CaretValueFormat3{Coordinate: 696, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 173}},
			CaretValueFormat3{Coordinate: 1351, Device: DeviceVariation{DeltaSetOuter: 6, DeltaSetInner: 14}},
		},
		// CaretCount=2
		{
			CaretValueFormat3{Coordinate: 702, Device: DeviceVariation{DeltaSetOuter: 3, DeltaSetInner: 179}},
			CaretValueFormat3{Coordinate: 1392, Device: DeviceVariation{DeltaSetOuter: 6, DeltaSetInner: 11}},
		},
	}

	if !reflect.DeepEqual(expectedLigGlyphs, gdef.LigatureCaretList.LigCarets) {
		t.Fatalf("expected %v, got %v", expectedLigGlyphs, gdef.LigatureCaretList.LigCarets)
	}
}
