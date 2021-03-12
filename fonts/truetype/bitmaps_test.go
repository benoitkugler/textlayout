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
		met := font.LoadMetrics().(*fontMetrics)

		for gid := GID(0); gid < fonts.GlyphIndex(font.NumGlyphs); gid++ {
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
			fmt.Println(len(strike.subTables))
		}
		met := font.LoadMetrics().(*fontMetrics)
		file.Close()

		cmap, _ := font.Cmap.BestEncoding()
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
