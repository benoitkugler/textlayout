package graphite

import (
	"math/rand"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

func TestTableSilf(t *testing.T) {
	filenames := []string{
		"testdata/Annapurnarc2.ttf",
		"testdata/Scheherazadegr.ttf",
	}
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}

		font, err := truetype.Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		ta, err := font.GetRawTable(tagSilf)
		if err != nil {
			t.Fatal(err)
		}

		_, err = parseTableSilf(ta)
		if err != nil {
			t.Fatal(err)
		}

		ta, err = font.GetRawTable(tagSill)
		if err != nil {
			t.Fatal(err)
		}

		_, err = parseTableSill(ta)
		if err != nil {
			t.Fatal(err)
		}

		ta, err = font.GetRawTable(tagFeat)
		if err != nil {
			t.Fatal(err)
		}

		_, err = parseTableFeat(ta)
		if err != nil {
			t.Fatal(err)
		}

		f.Close()
	}
}

func TestCrash(t *testing.T) {
	for range [100]int{} {
		input := make([]byte, rand.Intn(1000))
		rand.Read(input)
		parseTableSilf(input)
		parseTableSill(input)
		parseTableFeat(input)
	}
}
