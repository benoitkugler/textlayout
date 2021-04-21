package truetype

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

// using the iterator
func compileCmap(c Cmap) map[rune]GID {
	out := make(map[rune]GID)
	iter := c.Iter()
	for iter.Next() {
		r, g := iter.Char()
		out[r] = g
	}
	return out
}

func compileNativeCmap(c Cmap) map[rune]GID {
	switch c := c.(type) {
	case cmap0:
		return c
	case cmap4:
		return c.Compile()
	case cmap6:
		return c.Compile()
	case cmap12:
		return c.Compile()
	default:
		panic("should not happen")
	}
}

func (s cmap12) Compile() map[rune]GID {
	chars := map[rune]GID{}
	for _, cm := range s {
		for c := cm.start; c <= cm.end; c++ {
			chars[rune(c)] = GID(c - cm.start + cm.delta)
		}
	}
	return chars
}

func (s cmap6) Compile() map[rune]GID {
	chars := make(map[rune]GID, len(s.entries))
	for i, entry := range s.entries {
		chars[rune(i)+s.firstCode] = GID(entry)
	}
	return chars
}

func (s cmap4) Compile() map[rune]GID {
	out := make(map[rune]GID, len(s))
	for _, entry := range s {
		if entry.indexes == nil {
			for c := entry.start; c <= entry.end; c++ {
				out[rune(c)] = GID(int(c) + int(entry.delta))
				if c == 65535 { // avoid overflow
					break
				}
			}
		} else {
			for i, gi := range entry.indexes {
				out[rune(i)+rune(entry.start)] = gi
			}
		}
	}
	return out
}

func testCm(t *testing.T, cmap Cmap) {
	all := compileCmap(cmap)
	fmt.Println("	cmap:", len(all))
	for r, c := range all {
		c2 := cmap.Lookup(r)
		if c2 != c {
			t.Errorf("inconsistent lookup for rune %d : got %d and %d", r, c, c2)
		}
	}

	all2 := compileNativeCmap(cmap)
	if !reflect.DeepEqual(all, all2) {
		t.Error("inconsistant compile functions")
	}
}

func TestCmap(t *testing.T) {
	for _, file := range []string{
		"testdata/Roboto-BoldItalic.ttf",
		"testdata/Raleway-v4020-Regular.otf",
		"testdata/Castoro-Regular.ttf",
		"testdata/Castoro-Italic.ttf",
		"testdata/FreeSerif.ttf",
		"testdata/AnjaliOldLipi-Regular.ttf",
		"testdata/04B_30.ttf",
	} {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		for _, cmap := range font.Cmap.Cmaps {
			testCm(t, cmap.Cmap)
		}

		f.Close()
	}
}

func TestCmap6(t *testing.T) {
	cm := cmap6{
		firstCode: 45684,
		entries:   []uint16{1, 2, 5, 9, 489, 8231, 84},
	}
	testCm(t, cm)
}

func TestCmap4(t *testing.T) {
	d1, d2, d3 := int16(-9), int16(-18), int16(-80)
	input := []uint16{
		0, 0, 0,
		8,
		8, 4, 0,
		20, 90, 480, 0xffff,
		0, // reserved pad
		10, 30, 153, 0xffff,
		uint16(d1), uint16(d2), uint16(d3), 1,
		0, 0, 0, 0,
	}
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, input)
	if err != nil {
		t.Fatal(err)
	}

	cmap, err := parseCmapFormat4(buf.Bytes(), 0)
	if err != nil {
		t.Fatal(err)
	}

	runes := [...]rune{10, 20, 30, 90, 153, 480, 0xFFFF}
	glyphs := [...]GID{1, 11, 12, 72, 73, 400, 0}
	for i, r := range runes {
		got := cmap.Lookup(r)
		if exp := glyphs[i]; got != exp {
			t.Fatalf("expected %d, got %d", exp, got)
		}
	}
}

// load a .ttx file, produced by fonttools
func readExpectedCmap2() map[rune]GID {
	data, err := ioutil.ReadFile("testdata/cmap2_expected.ttx")
	if err != nil {
		log.Fatal(err)
	}

	type xmlDoc struct {
		Maps []struct {
			Code string `xml:"code,attr"`
			Name string `xml:"name,attr"`
		} `xml:"cmap>cmap_format_2>map"`
	}
	var doc xmlDoc
	err = xml.Unmarshal(data, &doc)
	if err != nil {
		log.Fatal(err)
	}

	out := make(map[rune]GID)
	for _, m := range doc.Maps {
		var (
			r   rune
			gid GID
		)
		fmt.Sscanf(m.Code, "0x%x", &r)
		fmt.Sscanf(m.Name, "cid%05d", &gid)
		out[r] = gid
	}

	return out
}

func TestCmap2(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/cmap2.bin")
	if err != nil {
		t.Fatal(err)
	}
	cmap, err := parseCmapFormat2(data, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, header := range cmap.subHeaders {
		_ = cmap.glyphIndexArray[header.rangeIndex : header.rangeIndex+int(header.entryCount)]
	}

	// expected := readExpectedCmap2()
	// fmt.Println(len(expected))
}

func TestBestEncoding(t *testing.T) {
	filename := "testdata/ToyTTC.ttc"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fs, err := Loader.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	font := fs[0].(*Font)
	if L := len(font.Cmap.Cmaps); L != 3 {
		t.Fatalf("expected 3 subtables, got %d", L)
	}
	cmap, _ := font.Cmap.BestEncoding()
	if cmap.Lookup(0x2026) == 0 {
		t.Fatalf("run 0x2026 not supported")
	}
}

func TestVariationSelector(t *testing.T) {
	filename := "testdata/ToyCMAP14.otf"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	gid, ok := font.LoadMetrics().GetVariationGlyph(33446, 917761)
	if !ok || gid != 2 {
		t.Fatalf("expected 2, true ; got %d, %v", gid, ok)
	}
}
