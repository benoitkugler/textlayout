package graphite

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
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

		ta, err = font.GetRawTable(tagGloc)
		if err != nil {
			t.Fatal(err)
		}

		locations, _, err := parseTableGloc(ta, int(font.NumGlyphs))
		if err != nil {
			t.Fatal(err)
		}

		ta, err = font.GetRawTable(tagGlat)
		if err != nil {
			t.Fatal(err)
		}

		attrs, err := parseTableGlat(ta, locations)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(len(attrs))

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
		parseTableGlat(input, []uint32{1, 45, 78, 896, 4566})
	}
}

// load a .ttx file, produced by fonttools
func readExpectedGlat() []map[uint16]int16 {
	data, err := ioutil.ReadFile("testdata/Annapurnarc2.ttx")
	if err != nil {
		log.Fatal(err)
	}

	type xmlDoc struct {
		Glyphs []struct {
			Name  string `xml:"name,attr"`
			Attrs []struct {
				Index uint16 `xml:"index,attr"`
				Value int16  `xml:"value,attr"`
			} `xml:"attribute"`
		} `xml:"Glat>glyph"`
	}
	var doc xmlDoc
	err = xml.Unmarshal(data, &doc)
	if err != nil {
		log.Fatal(err)
	}

	out := make([]map[uint16]int16, len(doc.Glyphs))
	for i, glyph := range doc.Glyphs {
		m := make(map[uint16]int16)
		for _, attr := range glyph.Attrs {
			m[attr.Index] = attr.Value
		}
		out[i] = m
	}

	return out
}

func TestGlat(t *testing.T) {
	filename := "testdata/Annapurnarc2.ttf"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := truetype.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	ta, err := font.GetRawTable(tagGloc)
	if err != nil {
		t.Fatal(err)
	}

	locations, _, err := parseTableGloc(ta, int(font.NumGlyphs))
	if err != nil {
		t.Fatal(err)
	}

	ta, err = font.GetRawTable(tagGlat)
	if err != nil {
		t.Fatal(err)
	}

	attrs, err := parseTableGlat(ta, locations)
	if err != nil {
		t.Fatal(err)
	}
	expected := readExpectedGlat()
	if len(attrs) != len(expected) {
		t.Errorf("wrong length")
	}

	for i, m := range expected {
		attrSet := attrs[i]
		for k, v := range m {
			if g := attrSet.get(k); g != v {
				t.Errorf("expected %d, got %d", v, g)
			}
			if _, in := m[k+1]; !in {
				if attrSet.get(k+1) != 0 {
					t.Errorf("expected not found")
				}
			}
		}
	}
}
