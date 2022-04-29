package truetype

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestGlyf(t *testing.T) {
	for _, filename := range []string{
		"Roboto-BoldItalic.ttf",
		"open-sans-v15-latin-regular.woff",
		"Commissioner-VF.ttf",
		"FreeSerif.ttf",
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		head, err := font.loadHeadTable()
		if err != nil {
			t.Fatal(err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := font.GlyfTable(ng, head.indexToLocFormat)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of glyphs:", len(gs))
	}
}

func assertGlyphHeaderEqual(t *testing.T, exp, got GlyphData) {
	if exp.Xmin != got.Xmin || exp.Ymin != got.Ymin || exp.Xmax != got.Xmax || exp.Ymax != got.Ymax {
		t.Errorf("expected glyph size (%d, %d, %d, %d), got (%d, %d, %d, %d)", exp.Xmin, exp.Ymin, exp.Xmax, exp.Ymax,
			got.Xmin, got.Ymin, got.Xmax, got.Ymax)
	}
}

func assertPointEqual(t *testing.T, exp, got glyphContourPoint) {
	if exp.x != got.x || exp.y != got.y {
		t.Errorf("expected contour point (%d,%d), got (%d,%d)", exp.x, exp.y, got.x, got.y)
	}
}

func assertCompositeEqual(t *testing.T, exp, got compositeGlyphPart) {
	exp.flags, got.flags = 0, 0
	if exp != got {
		t.Errorf("expected composite part %v, got %v", exp, got)
	}
}

func TestCoordinatesGlyph(t *testing.T) {
	// imported from fonttools
	g := contourPoint{SegmentPoint: fonts.SegmentPoint{X: 1, Y: 2}}
	g.translate(.5, 0)
	if g.X != 1.5 || g.Y != 2.0 {
		t.Errorf("expected (1.5, 2.0), got (%f, %f)", g.X, g.Y)
	}
	g = contourPoint{SegmentPoint: fonts.SegmentPoint{X: 1, Y: 2}}
	g.transform([4]float32{0.5, 0, 0, 0})
	if g.X != 0.5 || g.Y != 0. {
		t.Errorf("expected (0.5, 0.), got (%f, %f)", g.X, g.Y)
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
	assertGlyphHeaderEqual(t, GlyphData{Xmin: 12, Ymin: 0, Xmax: 1172, Ymax: 1430}, glyph)

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

func TestGlyphsRoman(t *testing.T) {
	// references values are from https://fontdrop.info/
	expecteds := []TableGlyf{
		{
			{
				Xmin: 96, Xmax: 528, Ymin: 0, Ymax: 660,
				data: simpleGlyphData{
					endPtsOfContours: []uint16{3, 9, 15, 18, 21},
					points: []glyphContourPoint{
						{x: 96, y: 0},
						{x: 96, y: 660},
						{x: 528, y: 660},
						{x: 528, y: 0},
						{x: 144, y: 32},
						{x: 476, y: 32},
						{x: 376, y: 208},
						{x: 314, y: 314},
						{x: 310, y: 314},
						{x: 246, y: 208},
						{x: 310, y: 366},
						{x: 314, y: 366},
						{x: 368, y: 458},
						{x: 462, y: 626},
						{x: 160, y: 626},
						{x: 254, y: 458},
						{x: 134, y: 74},
						{x: 288, y: 340},
						{x: 134, y: 610},
						{x: 488, y: 74},
						{x: 488, y: 610},
						{x: 336, y: 340},
					},
				},
			},
			{
				Xmin: 56, Xmax: 334, Ymin: -12, Ymax: 672,
				data: simpleGlyphData{
					endPtsOfContours: []uint16{18},
					points: []glyphContourPoint{
						{x: 334, y: -12},
						{x: 247, y: -12},
						{x: 124, y: 73},
						{x: 56, y: 228},
						{x: 56, y: 332},
						{x: 56, y: 410},
						{x: 95, y: 535},
						{x: 169, y: 625},
						{x: 271, y: 672},
						{x: 334, y: 672},
						{x: 334, y: 642},
						{x: 258, y: 642},
						{x: 149, y: 566},
						{x: 90, y: 427},
						{x: 90, y: 332},
						{x: 90, y: 237},
						{x: 148, y: 96},
						{x: 256, y: 18},
						{x: 334, y: 18},
					},
				},
			},
			{
				Xmin: 56, Xmax: 612, Ymin: -12, Ymax: 672,
				data: compositeGlyphData{
					glyphs: []compositeGlyphPart{
						{glyphIndex: 1, arg1: 0, arg2: 0, scale: [4]float32{1, 0, 0, 1}},
						{glyphIndex: 1, arg1: 0, arg2: 9, scale: [4]float32{-1, 0, 0, -1}},
					},
				},
			},
		},
		{
			{
				Xmin: 96, Xmax: 528, Ymin: 0, Ymax: 660,
				data: simpleGlyphData{
					endPtsOfContours: []uint16{3, 9, 15, 18, 21},
					points: []glyphContourPoint{
						{x: 96, y: 0},
						{x: 96, y: 660},
						{x: 528, y: 660},
						{x: 528, y: 0},
						{x: 144, y: 32},
						{x: 476, y: 32},
						{x: 376, y: 208},
						{x: 314, y: 314},
						{x: 310, y: 314},
						{x: 246, y: 208},
						{x: 310, y: 366},
						{x: 314, y: 366},
						{x: 368, y: 458},
						{x: 462, y: 626},
						{x: 160, y: 626},
						{x: 254, y: 458},
						{x: 134, y: 74},
						{x: 288, y: 340},
						{x: 134, y: 610},
						{x: 488, y: 74},
						{x: 488, y: 610},
						{x: 336, y: 340},
					},
				},
			},
			{
				Xmin: 10, Xmax: 510, Ymin: 0, Ymax: 660,
				data: simpleGlyphData{
					endPtsOfContours: []uint16{13, 17},
					points: []glyphContourPoint{
						{x: 10, y: 0},
						{x: 246, y: 660},
						{x: 274, y: 660},
						{x: 510, y: 0},
						{x: 476, y: 0},
						{x: 338, y: 396},
						{x: 317, y: 456},
						{x: 280, y: 565},
						{x: 262, y: 626},
						{x: 258, y: 626},
						{x: 240, y: 565},
						{x: 203, y: 456},
						{x: 182, y: 396},
						{x: 42, y: 0},
						{x: 112, y: 236},
						{x: 112, y: 264},
						{x: 405, y: 264},
						{x: 405, y: 236},
					},
				},
			},
			{
				Xmin: 10, Xmax: 510, Ymin: 0, Ymax: 846,
				data: compositeGlyphData{
					glyphs: []compositeGlyphPart{
						{glyphIndex: 1, arg1: 0, arg2: 0, scale: [4]float32{1, 0, 0, 1}},
						{glyphIndex: 3, arg1: 260, arg2: 0, scale: [4]float32{1, 0, 0, 1}},
					},
				},
			},
			{
				Xmin: -36, Xmax: 104, Ymin: 710, Ymax: 846,
				data: simpleGlyphData{
					endPtsOfContours: []uint16{3},
					points: []glyphContourPoint{
						{x: -22, y: 710},
						{x: -36, y: 726},
						{x: 82, y: 846},
						{x: 104, y: 822},
					},
				},
			},
		},
	}

	filenames := []string{
		"SourceSansVariable-Roman.anchor.ttf",
		"SourceSansVariable-Roman-nohvar-41,C1.ttf",
	}
	for i, filename := range filenames {
		expected := expecteds[i]

		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		head, err := font.loadHeadTable()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := font.GlyfTable(ng, head.indexToLocFormat)
		if err != nil {
			t.Fatal(err)
		}

		if len(gs) != len(expected) {
			t.Fatalf("expected %d glyphs, got %d", len(expected), len(gs))
		}
		for i, exp := range expected {
			got := gs[i]
			assertGlyphHeaderEqual(t, exp, got)

			switch d := exp.data.(type) {
			case simpleGlyphData:
				gd, ok := got.data.(simpleGlyphData)
				if !ok {
					t.Errorf("invalid type %T", got.data)
				}
				if !reflect.DeepEqual(d.endPtsOfContours, gd.endPtsOfContours) {
					t.Errorf("expected %v, got %v", d.endPtsOfContours, gd.endPtsOfContours)
				}
				if len(d.points) != len(gd.points) {
					t.Errorf("expected %d contour points, got %d", len(d.points), len(gd.points))
				}
				for i, p := range d.points {
					assertPointEqual(t, p, gd.points[i])
				}
			case compositeGlyphData:
				gd, ok := got.data.(compositeGlyphData)
				if !ok {
					t.Errorf("invalid type %T", got.data)
				}
				if len(d.glyphs) != len(gd.glyphs) {
					t.Errorf("expected %d glyphs, got %d", len(d.glyphs), len(gd.glyphs))
				}
				for i, comp := range d.glyphs {
					compGot := gd.glyphs[i]
					assertCompositeEqual(t, comp, compGot)
				}
			}
		}
	}
}

func TestGlyphExtentsFromPoints(t *testing.T) {
	font := loadFont(t, "SourceSansVariable-Roman.anchor.ttf")

	for i := 0; i < int(font.NumGlyphs); i++ {
		ext1, _ := font.GlyphExtents(fonts.GID(i), 0, 0)

		var out1 []contourPoint
		font.getPointsForGlyph(fonts.GID(i), 0, &out1)
		ext1bis := extentsFromPoints(out1)

		if ext1 != ext1bis {
			t.Errorf("invalid extents from points: expected %v, got %v", ext1, ext1bis)
		}
	}
}

func TestGlyphPhantoms(t *testing.T) {
	font := loadFont(t, "SourceSansVariable-Roman.modcomp.ttf")

	fmt.Println(font.vmtx)

	_, phantoms := font.getGlyfPoints(1, false)
	fmt.Println(phantoms)
}

func TestByteArg1Arg2(t *testing.T) {
	// Comfortaa font stripped to contain the single composite glyph "i" using
	// byte offsets (arg1And2AreWords is not set).
	font := loadFont(t, "Comfortaa-i.ttf")
	g := font.GlyphData(1, 100, 100).(fonts.GlyphOutline)
	for _, seg := range g.Segments {
		for _, p := range seg.Args {
			if p.X > 200 {
				t.Fatalf("a contour point is out of bounds: %v", p)
			}
		}
	}
}

func BenchmarkParseContour(b *testing.B) {
	data := []byte{
		0x2, 0xb4, 0x7d, 0x91, 0xfe, 0x6e, 0x1, 0x98, 0x8b, 0x7d, 0xfe, 0xf2, 0x6, 0x5, 0x18, 0x1e, 0x1e,
		0xb, 0xa8, 0x3, 0x14, 0xca, 0xca, 0x65, 0x2, 0x43, 0xfd, 0xcd, 0xc3, 0x86, 0x4b, 0xc, 0x27, 0x2d, 0x2d, 0x11,
		0xf6, 0xff, 0xff, 0x0, 0x1d, 0x0, 0x0, 0x6, 0xd3, 0x6, 0x1f, 0x0, 0x27, 0x0, 0x48, 0x2, 0xb0, 0x0, 0x0, 0x0,
		0x26, 0x0, 0x48, 0x0, 0x0, 0x0, 0x7, 0x0, 0x4b, 0x5, 0x6d, 0x0, 0x0, 0xff, 0xff, 0x0, 0x1d, 0x0, 0x0, 0x6,
		0xc3, 0x6, 0x1f, 0x0, 0x27, 0x0, 0x48, 0x2, 0xb0, 0x0, 0x0, 0x0, 0x26, 0x0, 0x48, 0x0, 0x0, 0x0, 0x7, 0x0, 0x4e,
		0x5, 0x6d, 0x0, 0x0, 0xff, 0xff, 0x0, 0x1d, 0x0, 0x0, 0x5, 0xc4, 0x6, 0x1f, 0x0, 0x27, 0x0, 0x48, 0x2, 0xb6, 0x0,
		0x0, 0x0, 0x6, 0x0, 0x48, 0x0, 0x0, 0x0, 0x1, 0x0, 0xc9, 0x0, 0x0, 0x1, 0x73, 0x5, 0xb6, 0x0, 0x3, 0x0, 0x11,
		0xb6, 0x0, 0x4, 0x5, 0x1, 0x3, 0x0, 0x12, 0x0, 0x3f, 0x3f, 0x11, 0x12, 0x1, 0x39, 0x31, 0x30, 0x33, 0x11, 0x33,
		0x11, 0xc9, 0xaa, 0x5, 0xb6, 0xfa, 0x4a, 0x0, 0xff, 0xff, 0x0, 0x5, 0x0, 0x0, 0x1, 0x8e,
		0x7, 0x73, 0x2, 0x26, 0x0, 0xd8, 0x0, 0x0, 0x1, 0x7, 0x0, 0x42, 0xfe, 0x7c, 0x1, 0x52, 0x0, 0x8, 0xb3, 0x1,
		0x5, 0x5, 0x26, 0x0, 0x2b, 0x35, 0xff, 0xff, 0x0, 0xb3, 0x0, 0x0, 0x2, 0x3c, 0x7, 0x73, 0x2, 0x26, 0x0, 0xd8,
		0x0, 0x0, 0x1, 0x7, 0x0, 0x75, 0xff, 0x2a, 0x1, 0x52, 0x0, 0x8, 0xb3, 0x1, 0xd, 0x5, 0x26, 0x0, 0x2b, 0x35, 0xff,
		0xff, 0xff, 0xc7, 0x0, 0x0, 0x2, 0x69, 0x7, 0x73, 0x2, 0x26, 0x0, 0xd8, 0x0, 0x0, 0x1, 0x7, 0x0, 0xc0, 0xfe, 0xbb,
		0x1, 0x52, 0x0, 0x8, 0xb3, 0x1, 0x12, 0x5, 0x26, 0x0, 0x2b, 0x35, 0xff, 0xff, 0x0, 0x5, 0x0, 0x0, 0x2, 0x38, 0x7,
		0x25, 0x2, 0x26, 0x0, 0xd8, 0x0, 0x0, 0x1, 0x7, 0x0, 0x69, 0xfe, 0xd0, 0x1, 0x52, 0x0, 0xa, 0xb4, 0x2, 0x1, 0x19,
		0x5, 0x26, 0x0, 0x2b, 0x35, 0x35,
	}
	points := []glyphContourPoint{{flag: 0x1, x: 0, y: 0}, {flag: 0x23, x: 0, y: 0}, {flag: 0x15, x: 0, y: 0}, {flag: 0x23, x: 0, y: 0}, {flag: 0x35, x: 0, y: 0}, {flag: 0x21, x: 0, y: 0}, {flag: 0x35, x: 0, y: 0}, {flag: 0x1, x: 0, y: 0}, {flag: 0x33, x: 0, y: 0}, {flag: 0x11, x: 0, y: 0}, {flag: 0x33, x: 0, y: 0}, {flag: 0x21, x: 0, y: 0}, {flag: 0x35, x: 0, y: 0}, {flag: 0x34, x: 0, y: 0}, {flag: 0x37, x: 0, y: 0}, {flag: 0xe, x: 0, y: 0}, {flag: 0xe, x: 0, y: 0}, {flag: 0xe, x: 0, y: 0}, {flag: 0xe, x: 0, y: 0}, {flag: 0x7, x: 0, y: 0}, {flag: 0x7, x: 0, y: 0}}

	for i := 0; i < b.N; i++ {
		parseGlyphContourPoints(data[:19], data[19:19+18], points)
	}
}
