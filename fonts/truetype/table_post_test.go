package truetype

import (
	"bytes"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestPost(t *testing.T) {
	for _, file := range []string{
		"Castoro-Regular.ttf",
		"Castoro-Italic.ttf",
		"FreeSerif.ttf",
	} {
		f, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := NewFontParser(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}

		cmaps, err := font.CmapTable()
		if err != nil {
			t.Fatal(err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		ps, err := font.PostTable(ng)
		if err != nil {
			t.Fatal(err)
		}
		if ps.Names == nil {
			t.Fatalf("expected post names for font %s", file)
		}

		cmap, _ := cmaps.BestEncoding()

		for _, gi := range compileCmap(cmap) {
			name := ps.Names.GlyphName(gi)
			if name == "" {
				t.Error("empty name")
			}
		}
	}
}
