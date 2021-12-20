package type1

import (
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestParseMetrics(t *testing.T) {
	for j, filename := range filenamesBounds {
		b, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}
		defer b.Close()
		font, err := Parse(b)
		if err != nil {
			t.Error(err)
		}
		exps := expectedBounds[j]
		if len(font.charstrings) != len(exps) {
			t.Fatalf("invalid num glyphs for %s", filename)
		}
		for i := range font.charstrings {
			if i == 0 {
				continue
			}
			_, bounds, adv, err := font.parseGlyphMetrics(fonts.GID(i), false)
			if err != nil {
				t.Fatal(err)
			}

			if bounds != exps[i] {
				t.Fatalf("invalid path bounds for glyph %d in %s: expected %v, got %v", i, filename, exps[i], bounds)
			}

			if exp := expectedAdvances[j][i]; adv != exp {
				t.Fatalf("invalid advance for glyph %d in %s: expected %v, got %v", i, filename, exp, adv)
			}
		}
	}
}
