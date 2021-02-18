package truetype

import (
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestKern(t *testing.T) {
	for _, file := range []string{
		"testdata/Castoro-Regular.ttf",
		"testdata/Castoro-Italic.ttf",
		"testdata/FreeSerif.ttf",
	} {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		font, err := Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		var kern SimpleKerns
		if font.tables[tagKern] != nil {
			_, err = font.KernTable()
			if err != nil {
				t.Fatal(err)
			}
			// fmt.Println("	kerns (kern):", kern.Size())
		}

		if font.tables[TagGpos] != nil {
			gpos, err := font.GPOSTable()
			if err != nil {
				t.Fatal(err)
			}
			kern, err = gpos.horizontalKerning()
			if err != nil {
				t.Fatal(err)
			}
			// fmt.Println("	kerns (GPOS):", kern.Size())
		}

		widths, err := font.HtmxTable()
		if err != nil {
			t.Fatal(err)
		}
		for gid := range widths {
			a, b := fonts.GlyphIndex(gid), fonts.GlyphIndex(gid+1)
			_, _ = kern.KernPair(a, b)
		}

		f.Close()
	}
}

func TestKern1(t *testing.T) {
	f, err := os.Open("testdata/ToyKern1.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	ng := font.NumGlyphs

	kerns, err := font.KernTable()
	if err != nil {
		t.Fatal(err)
	}

	for _, k := range kerns {
		if simple, ok := k.Data.(SimpleKerns); ok {
			for i := fonts.GlyphIndex(0); i < fonts.GlyphIndex(ng); i++ {
				for j := fonts.GlyphIndex(0); j < fonts.GlyphIndex(ng); j++ {
					simple.KernPair(i, j)
				}
			}
		}
	}
}
