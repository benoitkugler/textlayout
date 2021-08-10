package graphite

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

func loadGraphite(t *testing.T, filename string) *GraphiteFace {
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("can't open font file: %s", err)
	}
	defer f.Close()

	ft, err := truetype.Parse(f, true)
	if err != nil {
		t.Fatalf("can't parse truetype font %s: %s", filename, err)
	}

	font, err := LoadGraphite(ft)
	if err != nil {
		t.Fatalf("can't load graphite tables (font %s): %s", filename, err)
	}

	return font
}

func TestLoadGraphite(t *testing.T) {
	filenames := []string{
		"testdata/Annapurnarc2.ttf",
		"testdata/Scheherazadegr.ttf",
		"testdata/Awami_test.ttf",
		"testdata/charis.ttf",
		"testdata/Padauk.ttf",
		"testdata/MagyarLinLibertineG.ttf",
		"testdata/AwamiNastaliq-Regular.ttf",
		"testdata/Awami_compressed_test.ttf",
	}
	for _, filename := range filenames {
		font := loadGraphite(t, filename)
		fmt.Println(len(font.glyphs))
		// fmt.Println(font.glyphs[10].boxes
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
