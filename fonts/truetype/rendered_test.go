package truetype

import (
	"fmt"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestSbixGlyph(t *testing.T) {
	for _, filename := range []string{
		"testdata/ToyFeat.ttf",
		"testdata/ToySbix.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file, true)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		cmap, _ := font.Cmap()
		iter := cmap.Iter()

		gs, err := font.sbixTable()
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

		file.Close()
	}
}

func TestCblcGlyph(t *testing.T) {
	for _, filename := range []string{
		"testdata/ToyCBLC1.ttf",
		"testdata/ToyCBLC2.ttf",
		"testdata/NotoColorEmoji.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file, true)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		gs, err := font.colorBitmapTable()
		if err != nil {
			t.Fatal(err)
		}

		file.Close()

		cmap, _ := font.cmaps.BestEncoding()
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
		"testdata/IBM3161-bitmap.otb",
		"testdata/mry_KacstQurn.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatal(filename, err)
		}

		font, err := Parse(file, false)
		if err != nil {
			t.Fatal(filename, err)
		}

		gs, err := font.grayBitmapTable()
		if err != nil {
			t.Fatal(err)
		}

		cmap, _ := font.cmaps.BestEncoding()
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

		file.Close()
	}
}

func TestAppleBitmapGlyph(t *testing.T) {
	filename := "testdata/Gacha_9.dfont"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(filename, err)
	}
	defer file.Close()

	fs, err := Loader.(loader).Load(file)
	if err != nil {
		t.Fatal(err)
	}

	font := fs[0].(*Font)

	gs, err := font.appleBitmapTable()
	if err != nil {
		t.Fatal(err)
	}

	cmap, _ := font.cmaps.BestEncoding()
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
