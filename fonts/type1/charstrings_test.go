package type1

import (
	"bytes"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/type1"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestParseMetrics(t *testing.T) {
	for j, filename := range filenamesBounds {
		b, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		font, err := Parse(bytes.NewReader(b))
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
			_, bounds, adv, err := font.loadGlyph(fonts.GID(i), false)
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
