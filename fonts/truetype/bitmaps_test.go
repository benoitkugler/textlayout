package truetype

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestSbix(t *testing.T) {
	for _, filename := range []string{
		"ToyFeat.ttf",
		"ToySbix.ttf",
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := font.sbixTable(ng)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of strkes:", len(gs.strikes))

		for gid := GID(0); gid < fonts.GID(ng); gid++ {
			for _, strike := range gs.strikes {
				g := strike.getGlyph(gid, 0)
				if g.isNil() {
					continue
				}
				if _, ok := g.glyphExtents(); !ok {
					t.Error(filename, gid)
				}
			}
		}
	}
}

func TestCblc(t *testing.T) {
	for _, filename := range []string{
		"ToyCBLC1.ttf",
		"ToyCBLC2.ttf",
		"NotoColorEmoji.ttf",
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		pr, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		cmaps, err := pr.CmapTable()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := pr.colorBitmapTable()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of strikes:", len(gs))
		for _, strike := range gs {
			fmt.Println("number of subtables:", len(strike.subTables))
		}

		head, err := pr.loadHeadTable()
		if err != nil {
			t.Fatal(err)
		}

		font := Font{bitmap: gs, upem: head.Upem()}
		cmap, _ := cmaps.BestEncoding()
		iter := cmap.Iter()
		for iter.Next() {
			_, gid := iter.Char()
			font.getExtentsFromCBDT(gid, 94, 94)
		}
	}
}

func TestEblc(t *testing.T) {
	for _, filename := range []string{
		"mry_KacstQurn.ttf",
		"IBM3161-bitmap.otb",
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatal(filename, err)
		}

		gs, err := font.grayBitmapTable()
		if err != nil {
			t.Fatal(err)
		}

		for _, strike := range gs {
			fmt.Println(len(strike.subTables))
			strike.subTables = nil // not to flood the terminal
			fmt.Println(strike)
		}
	}
}

func TestAppleBitmap(t *testing.T) {
	filename := "Gacha_9.dfont"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatal(filename, err)
	}

	fonts, err := NewFontParsers(bytes.NewReader(file))
	if err != nil {
		t.Fatal(err)
	}

	font := fonts[0]

	gs, err := font.appleBitmapTable()
	if err != nil {
		t.Fatal(err)
	}

	for _, strike := range gs {
		fmt.Println(len(strike.subTables))
	}
}

func TestSize(t *testing.T) {
	expectedSizes := [][]fonts.BitmapSize{
		{
			{Height: 300, Width: 300, XPpem: 300, YPpem: 300},
		},
		{
			{Height: 127, Width: 136, XPpem: 109, YPpem: 109},
		},
		{
			{Height: 128, Width: 136, XPpem: 109, YPpem: 109},
		},
		{
			{Height: 128, Width: 117, XPpem: 94, YPpem: 94},
		},
		{
			{Height: 128, Width: 136, XPpem: 109, YPpem: 109},
		},
		{
			{Height: 33, Width: 8, XPpem: 16, YPpem: 16},
			{Height: 44, Width: 10, XPpem: 21, YPpem: 21},
		},
		{
			{Height: 16, Width: 15, XPpem: 16, YPpem: 16},
		},
		{
			{Height: 9, Width: 6, XPpem: 9, YPpem: 9}, // freetype actually gives a width of 0, which is suspicious
		},
	}
	for i, filename := range []string{
		"ToyFeat.ttf",
		"ToySbix.ttf",
		"ToyCBLC1.ttf",
		"ToyCBLC2.ttf",
		"NotoColorEmoji.ttf",
		"mry_KacstQurn.ttf",
		"IBM3161-bitmap.otb",
		"Gacha_9.dfont",
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(filename, err)
		}

		fonts, err := Load(bytes.NewReader(file))
		if err != nil {
			t.Fatal(filename, err)
		}

		font := fonts[0].(*Font)
		got := font.LoadBitmaps()
		if !reflect.DeepEqual(got, expectedSizes[i]) {
			t.Fatalf("font %s, expected %v got %v", filename, expectedSizes[i], got)
		}
	}
}
