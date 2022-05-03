package harfbuzz

import (
	"reflect"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from harfbuzz/test/api/test-font.c Copyright © 2011  Google, Inc. Behdad Esfahbod

var _ fonts.FaceMetrics = dummyFace{}

// implements fonts.FontMetrics  with no-ops
type dummyFace struct {
	Face
}

func (dummyFace) LoadMetrics() fonts.FaceMetrics { return dummyFace{} }

func (dummyFace) Upem() uint16               { return 1000 }
func (dummyFace) GlyphName(fonts.GID) string { return "" }
func (dummyFace) FontHExtents() (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}

func (dummyFace) FontVExtents() (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}
func (dummyFace) LineMetric(fonts.LineMetric) (float32, bool)           { return 0, false }
func (dummyFace) NominalGlyph(ch rune) (fonts.GID, bool)                { return 0, false }
func (dummyFace) VariationGlyph(ch, varSelector rune) (fonts.GID, bool) { return 0, false }
func (dummyFace) HorizontalAdvance(gid fonts.GID) float32               { return 0 }
func (dummyFace) VerticalAdvance(gid fonts.GID) float32                 { return 0 }
func (dummyFace) GlyphHOrigin(fonts.GID) (x, y Position, found bool) {
	return 0, 0, false
}

func (dummyFace) GlyphVOrigin(fonts.GID) (x, y Position, found bool) {
	return 0, 0, false
}

func (dummyFace) GlyphExtents(fonts.GID, uint16, uint16) (fonts.GlyphExtents, bool) {
	return fonts.GlyphExtents{}, false
}

func (dummyFace) GetGlyphContourPoint(glyph fonts.GID, pointIndex uint16) (x, y Position, ok bool) {
	return 0, 0, false
}

func TestFontProperties(t *testing.T) {
	font := NewFont(dummyFace{})

	/* Check scale */

	upem := int(font.face.Upem())
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
	face := openFontFile("fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
	font := NewFont(face)

	extents, result := font.GlyphExtents(2)
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 10)
	assertEqualInt32(t, extents.YBearing, 846)
	assertEqualInt32(t, extents.Width, 500)
	assertEqualInt32(t, extents.Height, -846)

	coords := [1]float32{500.0}
	font.SetVarCoordsDesign(coords[:])

	extents, result = font.GlyphExtents(2)
	assert(t, result)
	assertEqualInt32(t, extents.XBearing, 0)
	assertEqualInt32(t, extents.YBearing, 874)
	assertEqualInt32(t, extents.Width, 550)
	assertEqualInt32(t, extents.Height, -874)
}

func TestAdvanceTtVarNohvar(t *testing.T) {
	face := openFontFile("fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
	font := NewFont(face)

	x, y := font.GlyphAdvanceForDirection(2, LeftToRight)

	assertEqualInt32(t, x, 520)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(2, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1000)

	coords := []float32{500.0}
	font.SetVarCoordsDesign(coords)
	x, y = font.GlyphAdvanceForDirection(2, LeftToRight)

	assertEqualInt32(t, x, 551)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(2, TopToBottom)
	assertEqualInt32(t, x, 0)
	// https://lorp.github.io/samsa/src/samsa-gui.html disagree with harfbuzz here
	assertEqualInt32(t, y, -995)
}

func TestAdvanceTtVarHvarvvar(t *testing.T) {
	face := openFontFile("fonts/SourceSerifVariable-Roman-VVAR.abc.ttf")
	font := NewFont(face)

	x, y := font.GlyphAdvanceForDirection(1, LeftToRight)

	assertEqualInt32(t, x, 508)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(1, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1000)

	coords := []float32{700.0}
	font.SetVarCoordsDesign(coords)
	x, y = font.GlyphAdvanceForDirection(1, LeftToRight)

	assertEqualInt32(t, x, 531)
	assertEqualInt32(t, y, 0)

	x, y = font.GlyphAdvanceForDirection(1, TopToBottom)

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -1012)
}

func TestAdvanceTtVarAnchor(t *testing.T) {
	face := openFontFile("fonts/SourceSansVariable-Roman.anchor.ttf")
	font := NewFont(face)

	extents, result := font.GlyphExtents(2)
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 56)
	assertEqualInt32(t, extents.YBearing, 672)
	assertEqualInt32(t, extents.Width, 556)
	assertEqualInt32(t, extents.Height, -684)

	coords := []float32{500.0}
	font.SetVarCoordsDesign(coords)
	extents, result = font.GlyphExtents(2)
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 50)
	assertEqualInt32(t, extents.YBearing, 667)
	assertEqualInt32(t, extents.Width, 592)
	assertEqualInt32(t, extents.Height, -679)
}

func TestExtentsTtVarComp(t *testing.T) {
	face := openFontFile("fonts/SourceSansVariable-Roman.modcomp.ttf")
	font := NewFont(face)

	coords := []float32{800.0}
	font.SetVarCoordsDesign(coords)

	extents, result := font.GlyphExtents(2) /* Ccedilla, cedilla y-scaled by 0.8, with unscaled component offset */
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 663)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -894)

	extents, result = font.GlyphExtents(3) /* Cacute, acute y-scaled by 0.8, with unscaled component offset (default) */
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 909)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -921)

	extents, result = font.GlyphExtents(4) /* Ccaron, caron y-scaled by 0.8, with scaled component offset */
	assert(t, result)

	assertEqualInt32(t, extents.XBearing, 19)
	assertEqualInt32(t, extents.YBearing, 866)
	assertEqualInt32(t, extents.Width, 519)
	assertEqualInt32(t, extents.Height, -878)
}

func TestAdvanceTtVarCompV(t *testing.T) {
	face := openFontFile("fonts/SourceSansVariable-Roman.modcomp.ttf")
	font := NewFont(face)

	coords := []float32{800.0}
	font.SetVarCoordsDesign(coords)

	x, y := font.GlyphAdvanceForDirection(2, TopToBottom) /* No VVAR; 'C' in composite Ccedilla determines metrics */

	assertEqualInt32(t, x, 0)
	assertEqualInt32(t, y, -991)

	x, y = font.getGlyphOriginForDirection(2, TopToBottom)

	assertEqualInt32(t, x, 291)
	assertEqualInt32(t, y, 1012)
}

func TestAdvanceTtVarGvarInfer(t *testing.T) {
	face := openFontFile("fonts/TestGVAREight.ttf")
	coords := []float32{float32(100) / (1 << 14)}
	face.SetVarCoordinates(coords)

	font := NewFont(face)
	_, ok := font.GlyphExtents(4)
	assert(t, ok)
}

func TestLoadGraphite(t *testing.T) {
	face := openFontFile("fonts/Simple-Graphite-Font.ttf")
	font := NewFont(face)
	if font.gr == nil {
		t.Fatal("missing graphite tables")
	}
}

func TestLigCarets(t *testing.T) {
	face := openFontFile("fonts/NotoNastaliqUrdu-Regular.ttf")
	font := NewFont(face)
	font.XScale, font.YScale = int32(face.Upem())*2, int32(face.Upem())*4

	/* call with no result */
	if L := len(font.GetOTLigatureCarets(LeftToRight, 188)); L != 0 {
		t.Fatalf("for glyph %d, expected %d, got %d", 188, 0, L)
	}
	if L := len(font.GetOTLigatureCarets(LeftToRight, 1021)); L != 0 {
		t.Fatalf("for glyph %d, expected %d, got %d", 1021, 0, L)
	}

	/* a glyph with 3 ligature carets */
	carets := font.GetOTLigatureCarets(LeftToRight, 1020)
	expected := []Position{2718, 5438, 8156}
	if !reflect.DeepEqual(expected, carets) {
		t.Fatalf("for glyph %d, expected %v, got %v", 1020, expected, carets)
	}

	/* a glyph with 1 ligature caret */
	carets = font.GetOTLigatureCarets(LeftToRight, 1022)
	expected = []Position{3530}
	if !reflect.DeepEqual(expected, carets) {
		t.Fatalf("for glyph %d, expected %v, got %v", 1022, expected, carets)
	}

	/* a glyph with 2 ligature carets */
	carets = font.GetOTLigatureCarets(LeftToRight, 1023)
	expected = []Position{2352, 4706}
	if !reflect.DeepEqual(expected, carets) {
		t.Fatalf("for glyph %d, expected %v, got %v", 1023, expected, carets)
	}
}
