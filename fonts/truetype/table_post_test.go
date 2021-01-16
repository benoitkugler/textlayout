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
		font, err := Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		ps, err := font.PostTable()
		if err != nil {
			t.Fatal(err)
		}
		if ps.Names == nil {
			t.Fatalf("expected post names for font %s", file)
		}

		cmap, err := font.CmapTable()
		if err != nil {
			t.Fatal(err)
		}
		for _, gi := range compileCmap(cmap) {
			name := ps.Names.GlyphName(gi)
			if name == "" {
				t.Error("empty name")
			}
		}
		f.Close()
	}
}
