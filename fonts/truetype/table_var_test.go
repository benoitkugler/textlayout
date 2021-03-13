package truetype

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestVariations(t *testing.T) {
	expected := []VarAxis{
		{Tag: MustNewTag("wght"), Minimum: 100, Maximum: 900, Default: 100},
		{Tag: MustNewTag("slnt"), Minimum: -12, Maximum: 0, Default: 0},
		{Tag: MustNewTag("FLAR"), Minimum: 0, Maximum: 100, Default: 0},
		{Tag: MustNewTag("VOLM"), Minimum: 0, Maximum: 100, Default: 0},
	}
	f, err := os.Open("testdata/Commissioner-VF.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	vari := font.Fvar
	if vari == nil || len(vari.Axis) != len(expected) {
		t.Errorf("invalid number of axis")
	}
	for i, axe := range vari.Axis {
		// ignore flags and strid
		axe.flags = expected[i].flags
		axe.strid = expected[i].strid
		if axe != expected[i] {
			t.Errorf("expected %v, got %v", expected[i], axe)
		}
	}

	_, err = font.avarTable()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGvar(t *testing.T) {
	f, err := os.Open("testdata/Commissioner-VF.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	glyphs, err := font.glyfTable()
	if err != nil {
		t.Fatal(err)
	}

	ta, err := font.gvarTable(glyphs)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(ta.variations))
}

func TestHvar(t *testing.T) {
	f, err := os.Open("testdata/Commissioner-VF.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	ta, err := font.hvarTable()
	if err != nil {
		t.Fatal(err)
	}

	coords := []float32{-0.4, 0, 0.8, 1}
	for gid := GID(0); gid < fonts.GlyphIndex(font.NumGlyphs); gid++ {
		ta.getAdvanceVar(gid, coords)
	}
}

func TestPackedPointCount(t *testing.T) {
	inputs := [][]byte{
		{0, 0},
		{0x32, 0},
		{0x81, 0x22},
	}
	expecteds := []uint16{0, 50, 290}
	for i, input := range inputs {
		if got, _, _ := getPackedPointCount(input); got != expecteds[i] {
			t.Fatalf("expected %d, got %d", expecteds[i], got)
		}
	}
}

func TestPackedDeltas(t *testing.T) {
	in := []byte{0x03, 0x0A, 0x97, 0x00, 0xC6, 0x87, 0x41, 0x10, 0x22, 0xFB, 0x34}
	out, err := unpackDeltas(in, 14)
	if err != nil {
		t.Fatal(err)
	}
	if exp := []int16{10, -105, 0, -58, 0, 0, 0, 0, 0, 0, 0, 0, 4130, -1228}; !reflect.DeepEqual(out, exp) {
		t.Fatalf("expected %v, got %v", exp, out)
	}
}
