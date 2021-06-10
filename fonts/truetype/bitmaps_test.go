package truetype

import (
	"fmt"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestSbix(t *testing.T) {
	for _, filename := range []string{
		"testdata/ToyFeat.ttf",
		"testdata/ToySbix.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		gs, err := font.sbixTable()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of strkes:", len(gs.strikes))
		met := font.LoadMetrics().(*FontMetrics)

		for gid := GID(0); gid < fonts.GID(font.NumGlyphs); gid++ {
			met.getExtentsFromSbix(gid, nil, 94, 94)
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

		file.Close()
	}
}

func TestCblc(t *testing.T) {
	for _, filename := range []string{
		"testdata/ToyCBLC1.ttf",
		"testdata/ToyCBLC2.ttf",
		"testdata/NotoColorEmoji.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		gs, err := font.colorBitmapTable()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of strikes:", len(gs))
		for _, strike := range gs {
			fmt.Println("number of subtables:", len(strike.subTables))
		}
		met := font.LoadMetrics().(*FontMetrics)
		file.Close()

		cmap, _ := font.cmaps.BestEncoding()
		iter := cmap.Iter()
		for iter.Next() {
			_, gid := iter.Char()
			met.getExtentsFromCBDT(gid, 94, 94)
		}
	}
}

func TestEblc(t *testing.T) {
	for _, filename := range []string{
		"testdata/mry_KacstQurn.ttf",
		"testdata/IBM3161-bitmap.otb",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatal(filename, err)
		}

		font, err := Parse(file)
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
		file.Close()
	}
}

func TestAppleBitmap(t *testing.T) {
	filename := "testdata/Gacha_9.dfont"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(filename, err)
	}
	defer file.Close()

	fonts, err := Loader.(loader).Load(file)
	if err != nil {
		t.Fatal(err)
	}

	font := fonts[0].(*Font)

	gs, err := font.appleBitmapTable()
	if err != nil {
		t.Fatal(err)
	}

	for _, strike := range gs {
		fmt.Println(len(strike.subTables))
	}
}

func TestSize(t *testing.T) {
	// expectedSizes := [][]Size{
	// 	{
	// 		{Height: 300, Width: 300, XPpem: 300, YPpem: 300},
	// 	},
	// 	{
	// 		{Height: 127, Width: 136, XPpem: 109, YPpem: 109},
	// 	},
	// 	{
	// 		{Height: 128, Width: 136, XPpem: 109, YPpem: 109},
	// 	},
	// 	{
	// 		{Height: 128, Width: 117, XPpem: 94, YPpem: 94},
	// 	},
	// 	{
	// 		{Height: 128, Width: 136, XPpem: 109, YPpem: 109},
	// 	},
	// 	{
	// 		{Height: 33, Width: 8, XPpem: 16, YPpem: 16},
	// 		{Height: 44, Width: 10, XPpem: 21, YPpem: 21},
	// 	},
	// 	{
	// 		{Height: 16, Width: 15, XPpem: 16, YPpem: 16},
	// 	},
	// 	{
	// 		{Height: 9, Width: 0, XPpem: 9, YPpem: 9},
	// 	},
	// }
	for _, filename := range []string{
		"testdata/ToyFeat.ttf",
		"testdata/ToySbix.ttf",
		"testdata/ToyCBLC1.ttf",
		"testdata/ToyCBLC2.ttf",
		"testdata/NotoColorEmoji.ttf",
		"testdata/mry_KacstQurn.ttf",
		"testdata/IBM3161-bitmap.otb",
		"testdata/Gacha_9.dfont",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatal(filename, err)
		}

		fonts, err := Loader.(loader).Load(file)
		if err != nil {
			t.Fatal(filename, err)
		}

		font := fonts[0].(*Font)
		fmt.Println(font.loadBitmaps())

		file.Close()
	}
}
