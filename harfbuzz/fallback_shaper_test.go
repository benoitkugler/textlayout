package harfbuzz

import (
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from harfbuzz/test/api/test-shape.c  Copyright © 2011  Google, Inc. Behdad Esfahbod

func testFont(t *testing.T, font *Font) {
	buffer := NewBuffer()
	buffer.Props.Direction = LeftToRight
	buffer.AddRunes([]rune("TesT"), 0, 4)

	buffer.Shape(font, nil)

	glyphs := buffer.Info
	positions := buffer.Pos

	var (
		outputGlyphs    = []int{1, 2, 3, 1}
		outputXAdvances = []int{10, 6, 5, 10}
	)
	assertEqualInt(t, len(glyphs), 4)
	assertEqualInt(t, len(glyphs), len(positions))
	for i, info := range glyphs {
		assertEqualInt(t, outputGlyphs[i], int(info.Glyph))
		assertEqualInt(t, i, info.Cluster)
	}
	for i, pos := range positions {
		assertEqualInt(t, outputXAdvances[i], int(pos.XAdvance))
		assertEqualInt(t, 0, int(pos.XOffset))
		assertEqualInt(t, 0, int(pos.YAdvance))
		assertEqualInt(t, 0, int(pos.YOffset))
	}
}

type dummyFaceShape struct {
	dummyFace
	xScale int32
}

func (f dummyFaceShape) LoadMetrics() fonts.FaceMetrics { return f }

// the result should be in font units
func (f dummyFaceShape) HorizontalAdvance(gid fonts.GID) float32 {
	switch gid {
	case 1:
		return float32(10 * 1000 / f.xScale)
	case 2:
		return float32(6 * 1000 / f.xScale)
	case 3:
		return float32(5 * 1000 / f.xScale)
	}
	return 0
}

func (dummyFaceShape) NominalGlyph(ch rune) (fonts.GID, bool) {
	switch ch {
	case 'T':
		return 1, true
	case 'e':
		return 2, true
	case 's':
		return 3, true
	}
	return 0, false
}

func TestShape(t *testing.T) {
	font := NewFont(dummyFaceShape{xScale: 100})
	font.XScale = 100
	testFont(t, font)
}
