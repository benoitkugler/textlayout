package truetype

import (
	"os"
	"testing"
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
	names, err := font.NameTable()
	if err != nil {
		t.Fatal(err)
	}
	vari, err := font.VarTable(names)
	if err != nil {
		t.Fatal(err)
	}
	if len(vari.Axis) != len(expected) {
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
