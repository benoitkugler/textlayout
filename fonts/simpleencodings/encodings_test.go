package simpleencodings

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/glyphsnames"
)

var encs = [...]*Encoding{
	&MacExpert, &MacRoman, &AdobeStandard, &Symbol, &WinAnsi, &ZapfDingbats,
}

func TestNames(t *testing.T) {
	for _, e := range encs {
		fmt.Println(len(e.NameToRune()))
	}
}

func TestRuneNames(t *testing.T) {
	for i, e := range encs {
		for b, name := range e {
			if name == "" {
				continue
			}
			_, ok := glyphsnames.GlyphToRune(name)
			if !ok {
				t.Errorf("encoding %d missing glyph name for %x %s", i, b, name)
			}
		}
	}
}
