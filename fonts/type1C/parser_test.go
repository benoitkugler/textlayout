package type1c

import (
	"bytes"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/type1C"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestParseCFF(t *testing.T) {
	files := []string{
		"AAAPKB+SourceSansPro-Bold.cff",
		"AdobeMingStd-Light-Identity-H.cff",
		"YPTQCA+CMR17.cff",
	}
	ttfs, err := testdata.Files.ReadDir("ttf")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range ttfs {
		files = append(files, filepath.Join("ttf", f.Name()))
	}

	for _, file := range files {
		b, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err, "in", file)
		}
		fmt.Println(file, "num glyphs:", len(font.charstrings))

		if font.fdSelect != nil {
			for i := 0; i < len(font.charstrings); i++ {
				_, err = font.fdSelect.fontDictIndex(fonts.GID(i))
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		for glyphIndex := range font.charstrings {
			_, _, err := font.LoadGlyph(fonts.GID(glyphIndex))
			if err != nil {
				t.Fatalf("can't get extents for %s GID %d: %s", file, glyphIndex, err)
			}
		}
	}
}

func TestBulk(t *testing.T) {
	for _, file := range []string{
		"AAAPKB+SourceSansPro-Bold.cff",
		"AdobeMingStd-Light-Identity-H.cff",
		"YPTQCA+CMR17.cff",
	} {
		b, err := testdata.Files.ReadFile(file)
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
		"AAAPKB+SourceSansPro-Bold.cff",
		"AdobeMingStd-Light-Identity-H.cff",
		"YPTQCA+CMR17.cff",
	} {
		f, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		fonts, err := Load(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}
		for _, font := range fonts {
			font.PoscriptName()
			_, has := font.PostscriptInfo()
			if !has {
				t.Error("expected PS info")
			}
			font.LoadSummary()
		}
	}
}

func TestCIDFont(t *testing.T) {
	file := "AdobeMingStd-Light-Identity-H.cff"
	b, err := testdata.Files.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	font, err := Parse(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(font.localSubrs))
}
