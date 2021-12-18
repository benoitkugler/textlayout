package truetype

import (
	"os"
	"testing"
)

func TestPost(t *testing.T) {
	for _, file := range []string{
		"testdata/Castoro-Regular.ttf",
		"testdata/Castoro-Italic.ttf",
		"testdata/FreeSerif.ttf",
	} {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := NewFontParser(f)
		if err != nil {
			t.Fatal(err)
		}

		if err = font.loadCmapTable(); err != nil {
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

		cmap, _ := font.Cmaps.BestEncoding()

		for _, gi := range compileCmap(cmap) {
			name := ps.Names.GlyphName(gi)
			if name == "" {
				t.Error("empty name")
			}
		}
		f.Close()
	}
}
