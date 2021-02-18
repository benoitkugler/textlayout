package harfbuzz

import (
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from harfbuzz/test/api/test-font.c Copyright © 2011  Google, Inc. Behdad Esfahbod

var _ Face = dummyFace{}

// implements Face with no-ops
type dummyFace struct{}

func (dummyFace) GetUpem() uint16                                                 { return 1000 }
func (dummyFace) GetFontHExtents() (hb_font_extents_t, bool)                      { return hb_font_extents_t{}, false }
func (dummyFace) GetNominalGlyph(ch rune) (fonts.GlyphIndex, bool)                { return 0, false }
func (dummyFace) GetVariationGlyph(ch, varSelector rune) (fonts.GlyphIndex, bool) { return 0, false }
func (dummyFace) GetHorizontalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool) {
	return 0, false
}

func (dummyFace) GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool) {
	return 0, false
}
func (dummyFace) GetGlyphHOrigin(fonts.GlyphIndex) (x, y Position, found bool) { return 0, 0, false }
func (dummyFace) GetGlyphVOrigin(fonts.GlyphIndex) (x, y Position, found bool) { return 0, 0, false }
func (dummyFace) GetGlyphExtents(fonts.GlyphIndex) (GlyphExtents, bool)        { return GlyphExtents{}, false }
func (dummyFace) NormalizeVariations(coords []float32) []float32               { return coords }
func (dummyFace) GetGlyphContourPoint(glyph fonts.GlyphIndex, pointIndex uint16) (x, y Position, ok bool) {
	return 0, 0, false
}

func TestFontProperties(t *testing.T) {
	font := NewFont(dummyFace{})

	/* Check scale */

	upem := int(font.face.GetUpem())
	xScale, yScale := font.XScale, font.YScale
	assertEqualInt(t, int(xScale), upem)
	assertEqualInt(t, int(yScale), upem)

	assertEqualInt(t, int(font.XPpem), 0)
	assertEqualInt(t, int(font.YPpem), 0)
	assertEqualInt(t, int(font.Ptem), 0)
}
