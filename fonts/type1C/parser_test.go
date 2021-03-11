package type1c

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

func TestParseCFF(t *testing.T) {
	files := []string{
		"test/AAAPKB+SourceSansPro-Bold.cff",
		"test/AdobeMingStd-Light-Identity-H.cff",
		"test/YPTQCA+CMR17.cff",
	}
	ttfs, err := ioutil.ReadDir("test/ttf")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range ttfs {
		files = append(files, filepath.Join("test/ttf", f.Name()))
	}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err, "in", file)
		}
		fmt.Println("num glyphs:", len(font.charstrings))

		if font.fdSelect != nil {
			for i := 0; i < len(font.charstrings); i++ {
				_, err = font.fdSelect.fontDictIndex(fonts.GlyphIndex(i))
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		for glyphIndex, chars := range font.charstrings {
			var (
				psi     psinterpreter.Inter
				metrics type2CharstringHandler
			)
			var index byte = 0
			if font.fdSelect != nil {
				index, err = font.fdSelect.fontDictIndex(fonts.GlyphIndex(glyphIndex))
				if err != nil {
					t.Fatal(err)
				}
			}
			subrs := font.localSubrs[index]
			if err := psi.Run(chars, subrs, font.globalSubrs, &metrics); err != nil {
				t.Fatal(err, "in", file, chars, glyphIndex)
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

func TestLoader(t *testing.T) {
	for _, file := range []string{
		"test/AAAPKB+SourceSansPro-Bold.cff",
		"test/AdobeMingStd-Light-Identity-H.cff",
		"test/YPTQCA+CMR17.cff",
	} {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		fonts, err := Loader.Load(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, font := range fonts {
			font.PoscriptName()
			_, has := font.PostscriptInfo()
			if !has {
				t.Error("expected PS info")
			}
			font.Style()
		}
	}
}

func TestCIDFont(t *testing.T) {
	file := "test/AdobeMingStd-Light-Identity-H.cff"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	font, err := Parse(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(font.localSubrs))
}
