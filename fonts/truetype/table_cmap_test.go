package truetype

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"log"
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
	"github.com/benoitkugler/textlayout/fonts"
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
	case cmap6or10:
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
			chars[rune(c)] = GID(c - cm.start + cm.value)
		}
	}
	return chars
}

func (s cmap6or10) Compile() map[rune]GID {
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
				out[rune(c)] = GID(c + entry.delta)
				if c == 0xFFFF { // avoid overflow
					break
				}
			}
		} else {
			for i, gi := range entry.indexes {
				out[rune(i)+rune(entry.start)] = GID(gi)
			}
		}
	}
	return out
}

func testCm(t *testing.T, cmap Cmap) {
	all := compileCmap(cmap)
	fmt.Println("	cmap:", len(all))
	for r, c := range all {
		c2, _ := cmap.Lookup(r)
		if c2 != c {
			t.Errorf("inconsistent lookup for rune %d : got %d and %d", r, c, c2)
		}
	}

	all2 := compileNativeCmap(cmap)
	if !reflect.DeepEqual(all, all2) {
		t.Errorf("inconsistant compile functions for type %T", cmap)
	}
}

func TestCmap(t *testing.T) {
	for _, file := range []string{
		"Roboto-BoldItalic.ttf",
		"Raleway-v4020-Regular.otf",
		"Castoro-Regular.ttf",
		"Castoro-Italic.ttf",
		"FreeSerif.ttf",
		"AnjaliOldLipi-Regular.ttf",
		"04B_30.ttf",
	} {
		f, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := NewFontParser(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}

		cmaps, err := font.CmapTable()
		if err != nil {
			t.Fatal(err)
		}

		for _, cmap := range cmaps.Cmaps {
			testCm(t, cmap.Cmap)
		}
	}
}

func TestCmap6(t *testing.T) {
	cm := cmap6or10{
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
		got, _ := cmap.Lookup(r)
		if exp := glyphs[i]; got != exp {
			t.Fatalf("expected %d, got %d", exp, got)
		}
	}
}

// load a .ttx file, produced by fonttools
func readExpectedCmap2() map[rune]GID {
	data, err := testdata.Files.ReadFile("cmap2_expected.ttx")
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
	data, err := testdata.Files.ReadFile("cmap2.bin")
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
	filename := "ToyTTC.ttc"
	f, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	fs, err := NewFontParsers(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}

	font := fs[0]
	cmaps, err := font.CmapTable()
	if err != nil {
		t.Fatal(err)
	}

	if L := len(cmaps.Cmaps); L != 3 {
		t.Fatalf("expected 3 subtables, got %d", L)
	}
	cmap, _ := cmaps.BestEncoding()
	if _, ok := cmap.Lookup(0x2026); !ok {
		t.Fatalf("rune 0x2026 not supported")
	}
}

func TestVariationSelector(t *testing.T) {
	font := loadFont(t, "ToyCMAP14.otf")

	gid, ok := font.VariationGlyph(33446, 917761)
	if !ok || gid != 2 {
		t.Fatalf("expected 2, true ; got %d, %v", gid, ok)
	}
}

func TestCmap12(t *testing.T) {
	font := loadFont(t, "ToyCMAP12.otf")

	cmap, _ := font.Cmap()
	fmt.Printf("%T", cmap)

	runes := [...]rune{
		0x0011, 0x0012, 0x0013, 0x0014, 0x0015, 0x0016, 0x0017, 0x0018,
	}
	gids := [...]fonts.GID{
		17, 18, 19, 20, 21, 22, 23, 24,
	}

	for i, r := range runes {
		got, _ := cmap.Lookup(r)
		if exp := gids[i]; exp != got {
			t.Fatalf("for rune 0x%x expected %d, got %d", r, exp, got)
		}
	}
}
