package type1C

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

func TestParseCFF(t *testing.T) {
	for _, file := range []string{
		"test/AAAPKB+SourceSansPro-Bold.cff",
		"test/AdobeMingStd-Light-Identity-H.cff",
		"test/YPTQCA+CMR17.cff",
	} {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("num glyphs:", len(font.charstrings))

		if font.fdSelect != nil {
			for i := 0; i < len(font.charstrings); i++ {
				_, err := font.fdSelect.fontDictIndex(fonts.GlyphIndex(i))
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		for _, chars := range font.charstrings {
			var (
				psi     psinterpreter.Inter
				metrics type2Metrics
			)
			if err := psi.Run(chars, nil, &metrics); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestBulk(t *testing.T) {
	for _, file := range []string{
		"test/AAAPKB+SourceSansPro-Bold.cff",
		"test/AdobeMingStd-Light-Identity-H.cff",
		"test/YPTQCA+CMR17.cff",
	} {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for range [100]int{} {
			for range [500]int{} { // random mutation
				i := rand.Intn(len(b))
				b[i] = byte(rand.Intn(256))
			}
			Parse(bytes.NewReader(b)) // we just check for crashes
		}
	}
}
