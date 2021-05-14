package graphite

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

func TestTableSilf(t *testing.T) {
	filenames := []string{
		"testdata/Annapurnarc2.ttf",
		"testdata/Scheherazadegr.ttf",
		"testdata/Awami_test.ttf",
	}
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}

		ft, err := truetype.Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		font, err := LoadGraphite(ft)
		if err != nil {
			t.Fatalf("font %s: %s", filename, err)
		}
		fmt.Println(len(font.glyphs))
		// fmt.Println(font.glyphs[10].boxes)
		f.Close()
	}
}

func TestCrash(t *testing.T) {
	for range [100]int{} {
		input := make([]byte, rand.Intn(1000))
		rand.Read(input)
		parseTableSilf(input, 5, 5)
		parseTableSill(input)
		parseTableFeat(input)
		parseTableGlat(input, []uint32{1, 45, 78, 896, 4566})
	}
}
