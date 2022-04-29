package truetype

import (
	"bytes"
	"fmt"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestHtmx(t *testing.T) {
	for _, file := range []string{
		"Roboto-BoldItalic.ttf",
		"Raleway-v4020-Regular.otf",
		"Castoro-Regular.ttf",
		"Castoro-Italic.ttf",
		"FreeSerif.ttf",
		"AnjaliOldLipi-Regular.ttf",
	} {
		f, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := NewFontParser(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		widths, err := font.HtmxTable(ng)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("	widths:", len(widths))
	}
}
