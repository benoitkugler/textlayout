package type1

import (
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestPsi(t *testing.T) {
	file := "test/CalligrapherRegular.pfb"
	b, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	font, err := Parse(b)
	if err != nil {
		t.Error(err)
	}
	for i := range font.charstrings {
		_, err := font.GetAdvance(fonts.GlyphIndex(i))
		if err != nil {
			t.Fatal(err)
		}
	}
}
