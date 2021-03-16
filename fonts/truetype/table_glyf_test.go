package truetype

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestGlyf(t *testing.T) {
	for _, filename := range []string{
		"testdata/Roboto-BoldItalic.ttf",
		"testdata/open-sans-v15-latin-regular.woff",
		"testdata/Commissioner-VF.ttf",
		"testdata/FreeSerif.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		gs, err := font.glyfTable()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of glyphs:", len(gs))

		file.Close()
	}
}

func TestCoordinatesGlyph(t *testing.T) {
	// imported from fonttools
	g := contourPoint{x: 1, y: 2}
	g.translate(.5, 0)
	if g.x != 1.5 || g.y != 2.0 {
		t.Errorf("expected (1.5, 2.0), got (%f, %f)", g.x, g.y)
	}
	g = contourPoint{x: 1, y: 2}
	g.transform([4]float32{0.5, 0, 0, 0})
	if g.x != 0.5 || g.y != 0. {
		t.Errorf("expected (0.5, 0.), got (%f, %f)", g.x, g.y)
	}

	glyfBin := []byte{0x0, 0x2, 0x0, 0xc, 0x0, 0x0, 0x4, 0x94, 0x5, 0x96, 0x0, 0x6, 0x0, 0xa, 0x0, 0x0, 0x41, 0x33, 0x1, 0x23, 0x1, 0x1, 0x23, 0x13, 0x21, 0x15, 0x21, 0x1, 0xf5, 0xb6, 0x1, 0xe9, 0xbd, 0xfe, 0x78, 0xfe, 0x78, 0xbb, 0xed, 0x2, 0xae, 0xfd, 0x52, 0x5, 0x96, 0xfa, 0x6a, 0x4, 0xa9, 0xfb, 0x57, 0x2, 0x2, 0xa2}
	headBin := []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x6e, 0x4f, 0x1c, 0xcf, 0x5f, 0xf, 0x3c, 0xf5, 0x20, 0x1b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0xd7, 0xc0, 0x23, 0x1c, 0x0, 0x0, 0x0, 0x0, 0xd7, 0xc1, 0x2e, 0xf9, 0x0, 0xc, 0x0, 0x0, 0x4, 0x94, 0x5, 0x96, 0x0, 0x0, 0x0, 0x9, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0}
	locaBin := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1b}
	maxpBin := []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x4, 0x0, 0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0x0, 0x2, 0x0, 0x1, 0x61, 0x0, 0x0, 0x0, 0x0}

	num, err := parseTableMaxp(maxpBin)
	if err != nil {
		t.Fatal(err)
	}
	head, err := parseTableHead(headBin)
	if err != nil {
		t.Fatal(err)
	}
	loca, err := parseTableLoca(locaBin, int(num), head.indexToLocFormat == 1)
	if err != nil {
		t.Fatal(err)
	}
	glyphs, err := parseTableGlyf(glyfBin, loca)
	if err != nil {
		t.Fatal(err)
	}

	if num != 4 {
		t.Errorf("expected 4 glyphs, got %d", num)
	}
	if glyphs[0].data != nil {
		t.Errorf("expected no glyph data for glyph 0")
	}
	if glyphs[1].data != nil {
		t.Errorf("expected no glyph data for glyph 1")
	}
	if glyphs[2].data != nil {
		t.Errorf("expected no glyph data for glyph 2")
	}

	glyph := glyphs[3]
	if glyph.Xmin != 12 || glyph.Ymin != 0 || glyph.Xmax != 1172 || glyph.Ymax != 1430 {
		t.Errorf("expected (12,0,1172,1430), got (%d, %d, %d, %d)", glyph.Xmin, glyph.Ymin, glyph.Xmax, glyph.Ymax)
	}
	glyphData, ok := glyph.data.(simpleGlyphData)
	if !ok {
		t.Errorf("expected simple glyph, got %T", glyph.data)
	}
	if len(glyphData.instructions) != 0 {
		t.Errorf("expected empty instructions, got %v", glyphData.instructions)
	}
	if !reflect.DeepEqual(glyphData.endPtsOfContours, []uint16{6, 10}) {
		t.Errorf("expected 2 contours, got %v", glyphData.endPtsOfContours)
	}
	type expected struct {
		x, y    int16
		overlap bool
	}
	exp := []expected{
		{x: 501, y: 1430, overlap: true},
		{x: 683, y: 1430},
		{x: 1172, y: 0},
		{x: 983, y: 0},
		{x: 591, y: 1193},
		{x: 199, y: 0},
		{x: 12, y: 0},
		{x: 249, y: 514},
		{x: 935, y: 514},
		{x: 935, y: 352},
		{x: 249, y: 352},
	}
	if len(glyphData.points) != 11 {
		t.Errorf("expected 11 points, got %v", glyphData.points)
	}
	for i, v := range glyphData.points {
		e := exp[i]
		if v.x != e.x || v.y != e.y || (v.flag&overlapSimple != 0) != e.overlap {
			t.Errorf("expected %v, got %v", e, v)
		}
	}
}
