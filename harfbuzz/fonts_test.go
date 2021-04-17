package harfbuzz

import (
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from harfbuzz/test/api/test-font.c Copyright © 2011  Google, Inc. Behdad Esfahbod

var _ Face = dummyFace{}

// implements Face with no-ops
type dummyFace struct{}

func (dummyFace) GetUpem() uint16               { return 1000 }
func (dummyFace) GetGlyphName(fonts.GID) string { return "" }
func (dummyFace) GetFontHExtents([]float32) (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}
func (dummyFace) GetNominalGlyph(ch rune) (fonts.GID, bool)                    { return 0, false }
func (dummyFace) GetVariationGlyph(ch, varSelector rune) (fonts.GID, bool)     { return 0, false }
func (dummyFace) GetHorizontalAdvance(gid fonts.GID, coords []float32) float32 { return 0 }
func (dummyFace) GetVerticalAdvance(gid fonts.GID, coords []float32) float32   { return 0 }
func (dummyFace) GetGlyphHOrigin(fonts.GID, []float32) (x, y Position, found bool) {
	return 0, 0, false
}

func (dummyFace) GetGlyphVOrigin(fonts.GID, []float32) (x, y Position, found bool) {
	return 0, 0, false
}

func (dummyFace) GetGlyphExtents(fonts.GID, []float32, uint16, uint16) (fonts.GlyphExtents, bool) {
	return fonts.GlyphExtents{}, false
}
func (dummyFace) NormalizeVariations(coords []float32) []float32 { return coords }
func (dummyFace) GetGlyphContourPoint(glyph fonts.GID, pointIndex uint16) (x, y Position, ok bool) {
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

// Unit tests for glyph advance Widths and extents of TrueType variable fonts
// ported from harfbuzz/test/api/test-ot-metrics-tt-var.c Copyright © 2019 Adobe Inc. Michiharu Ariza

func TestExtentsTtVar(t *testing.T) {
	face := openFontFile("testdata/fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
	font := NewFont(face.LoadMetrics())

	extents, result := font.getGlyphExtents(2)
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 10)
	assertEqualInt32(t, extents.YBearing, 846)
	assertEqualInt32(t, extents.Width, 500)
	assertEqualInt32(t, extents.Height, -846)

	coords := [1]float32{500.0}
	font.SetVarCoordsDesign(coords[:])

	extents, result = font.getGlyphExtents(2)
	assert(t, result)
	assertEqualInt32(t, extents.XBearing, 0)
	assertEqualInt32(t, extents.YBearing, 874)
	assertEqualInt32(t, extents.Width, 550)
	assertEqualInt32(t, extents.Height, -874)
}

func TestAdvanceTtVarNohvar(t *testing.T) {
	face := openFontFile("testdata/fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
	font := NewFont(face.LoadMetrics())

	x, y := font.getGlyphAdvanceForDirection(2, LeftToRight)

	assertEqualInt32(t, x, 520)
	assertEqualInt32(t, y, 0)

	x, y = font.getGlyphAdvanceForDirection(2, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1000)

	coords := []float32{500.0}
	font.SetVarCoordsDesign(coords)
	x, y = font.getGlyphAdvanceForDirection(2, LeftToRight)

	assertEqualInt32(t, x, 551)
	assertEqualInt32(t, y, 0)

	x, y = font.getGlyphAdvanceForDirection(2, TopToBottom)
	assertEqualInt32(t, x, 0)
	// https://lorp.github.io/samsa/src/samsa-gui.html disagree with harfbuzz here
	assertEqualInt32(t, y, -995)
}

func TestAdvanceTtVarHvarvvar(t *testing.T) {
	face := openFontFile("testdata/fonts/SourceSerifVariable-Roman-VVAR.abc.ttf")
	font := NewFont(face.LoadMetrics())

	x, y := font.getGlyphAdvanceForDirection(1, LeftToRight)

	assertEqualInt32(t, x, 508)
	assertEqualInt32(t, y, 0)

	x, y = font.getGlyphAdvanceForDirection(1, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1000)

	coords := []float32{700.0}
	font.SetVarCoordsDesign(coords)
	x, y = font.getGlyphAdvanceForDirection(1, LeftToRight)

	assertEqualInt32(t, x, 531)
	assertEqualInt32(t, y, 0)

	x, y = font.getGlyphAdvanceForDirection(1, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1012)
}

func TestAdvanceTtVarAnchor(t *testing.T) {
	face := openFontFile("testdata/fonts/SourceSansVariable-Roman.anchor.ttf")
	font := NewFont(face.LoadMetrics())

	extents, result := font.getGlyphExtents(2)
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 56)
	assertEqualInt32(t, extents.YBearing, 672)
	assertEqualInt32(t, extents.Width, 556)
	assertEqualInt32(t, extents.Height, -684)

	coords := []float32{500.0}
	font.SetVarCoordsDesign(coords)
	extents, result = font.getGlyphExtents(2)
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 50)
	assertEqualInt32(t, extents.YBearing, 667)
	assertEqualInt32(t, extents.Width, 592)
	assertEqualInt32(t, extents.Height, -679)
}

func TestExtentsTtVarComp(t *testing.T) {
	face := openFontFile("testdata/fonts/SourceSansVariable-Roman.modcomp.ttf")
	font := NewFont(face.LoadMetrics())

	coords := []float32{800.0}
	font.SetVarCoordsDesign(coords)

	extents, result := font.getGlyphExtents(2) /* Ccedilla, cedilla y-scaled by 0.8, with unscaled component offset */
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 663)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -894)

	extents, result = font.getGlyphExtents(3) /* Cacute, acute y-scaled by 0.8, with unscaled component offset (default) */
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 909)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -921)

	extents, result = font.getGlyphExtents(4) /* Ccaron, caron y-scaled by 0.8, with scaled component offset */
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 866)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -878)
}

func TestAdvanceTtVarCompV(t *testing.T) {
	face := openFontFile("testdata/fonts/SourceSansVariable-Roman.modcomp.ttf")
	font := NewFont(face.LoadMetrics())

	coords := []float32{800.0}
	font.SetVarCoordsDesign(coords)

	x, y := font.getGlyphAdvanceForDirection(2, TopToBottom) /* No VVAR; 'C' in composite Ccedilla determines metrics */

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -991)

	x, y = font.getGlyphOriginForDirection(2, TopToBottom)

	assertEqualInt32(t, x, 291)
	assertEqualInt32(t, y, 1012)
}

func TestAdvanceTtVarGvarInfer(t *testing.T) {
	face := openFontFile("testdata/fonts/TestGVAREight.ttf")
	font := NewFont(face.LoadMetrics())

	coords := []float32{float32(100) / (1 << 14)}
	font.coords = coords
	_, ok := font.getGlyphExtents(4)
	assert(t, ok)
}
