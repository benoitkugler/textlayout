package truetype

import (
	"bytes"
	"fmt"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestSbixGlyph(t *testing.T) {
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

		cmaps, err := font.CmapTable()
		if err != nil {
			t.Fatal(err)
		}

		cmap, _ := cmaps.BestEncoding()
		iter := cmap.Iter()

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := font.sbixTable(ng)
		if err != nil {
			t.Fatal(err)
		}

		for iter.Next() {
			_, gid := iter.Char()
			if gid == 0 {
				continue
			}

			data, err := gs.glyphData(gid, 100, 100)
			if err != nil {
				continue
			}
			if data.Format != fonts.PNG {
				t.Fatalf("unexpected format %d", data.Format)
			}
			fmt.Println(data.Width, data.Height)
		}
	}
}

func TestCblcGlyph(t *testing.T) {
	for _, filename := range []string{
		"ToyCBLC1.ttf",
		"ToyCBLC2.ttf",
		"NotoColorEmoji.ttf",
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		cmaps, err := font.CmapTable()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := font.colorBitmapTable()
		if err != nil {
			t.Fatal(err)
		}

		cmap, _ := cmaps.BestEncoding()
		iter := cmap.Iter()
		for iter.Next() {
			_, gid := iter.Char()
			data, err := gs.glyphData(gid, 94, 94)
			if err != nil {
				// t.Logf("unsupported rune %d", gid)
				continue
			}
			if data.Format != fonts.PNG {
				t.Fatalf("unexpected format %d", data.Format)
			}
			if data.Width != 136 || data.Height != 128 {
				t.Fatalf("unexpected width and height %d %d", data.Width, data.Height)
			}
		}
	}
}

func TestEblcGlyph(t *testing.T) {
	runess := [][]rune{
		[]rune("The quick brown fox jumps over the lazy dog"),
		{1569, 1570, 1571, 1572, 1573, 1574, 1575, 1576, 1577, 1578, 1579},
	}
	for i, filename := range []string{
		"IBM3161-bitmap.otb",
		"mry_KacstQurn.ttf",
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatal(filename, err)
		}

		cmaps, err := font.CmapTable()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := font.grayBitmapTable()
		if err != nil {
			t.Fatal(err)
		}

		cmap, _ := cmaps.BestEncoding()
		runes := runess[i]
		for _, r := range runes {
			gid, ok := cmap.Lookup(r)
			if !ok {
				t.Fatalf("unsupported rune %d", r)
			}
			data, err := gs.glyphData(gid, 94, 94)
			if err != nil {
				t.Fatal(err)
			}
			if data.Format != fonts.BlackAndWhite {
				t.Fatalf("unexpected format %d", data.Format)
			}
		}
	}
}

func TestAppleBitmapGlyph(t *testing.T) {
	filename := "Gacha_9.dfont"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatal(filename, err)
	}

	fs, err := NewFontParsers(bytes.NewReader(file))
	if err != nil {
		t.Fatal(err)
	}

	font := fs[0]

	cmaps, err := font.CmapTable()
	if err != nil {
		t.Fatal(err)
	}

	gs, err := font.appleBitmapTable()
	if err != nil {
		t.Fatal(err)
	}

	cmap, _ := cmaps.BestEncoding()
	runes := []rune("The quick brown fox jumps over the lazy dog")
	for _, r := range runes {
		gid, ok := cmap.Lookup(r)
		if !ok {
			t.Fatalf("unsupported rune %d", r)
		}
		data, err := gs.glyphData(gid, 94, 94)
		if err != nil {
			t.Fatal(err)
		}
		if data.Format != fonts.BlackAndWhite {
			t.Fatalf("unexpected format %d", data.Format)
		}
	}
}

func TestMixedGlyphs(t *testing.T) {
	for _, filename := range []string{
		"Roboto-BoldItalic.ttf",
		"Raleway-v4020-Regular.otf",
		"open-sans-v15-latin-regular.woff",
		"OldaniaADFStd-Bold.otf", // duplicate tables
	} {
		font := loadFont(t, filename)
		space, ok := font.NominalGlyph(' ')
		if !ok {
			t.Fatal(filename)
		}
		if font.GlyphData(space, 94, 94) == nil {
			t.Fatal(filename)
		}
	}
}
