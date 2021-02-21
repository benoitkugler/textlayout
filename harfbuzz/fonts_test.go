package harfbuzz

import (
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from harfbuzz/test/api/test-font.c Copyright © 2011  Google, Inc. Behdad Esfahbod

var _ Face = dummyFace{}

// implements Face with no-ops
type dummyFace struct{}

func (dummyFace) GetUpem() uint16 { return 1000 }
func (dummyFace) GetFontHExtents([]float32) (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}
func (dummyFace) GetNominalGlyph(ch rune) (fonts.GlyphIndex, bool)                  { return 0, false }
func (dummyFace) GetVariationGlyph(ch, varSelector rune) (fonts.GlyphIndex, bool)   { return 0, false }
func (dummyFace) GetHorizontalAdvance(gid fonts.GlyphIndex, coords []float32) int16 { return 0 }
func (dummyFace) GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) int16   { return 0 }
func (dummyFace) GetGlyphHOrigin(fonts.GlyphIndex) (x, y Position, found bool)      { return 0, 0, false }
func (dummyFace) GetGlyphVOrigin(fonts.GlyphIndex) (x, y Position, found bool)      { return 0, 0, false }
func (dummyFace) GetGlyphExtents(fonts.GlyphIndex) (fonts.GlyphExtents, bool) {
	return fonts.GlyphExtents{}, false
}
func (dummyFace) NormalizeVariations(coords []float32) []float32 { return coords }
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

// Unit tests for glyph advance widths and extents of TrueType variable fonts
// ported from harfbuzz/test/api/test-ot-metrics-tt-var.c Copyright © 2019 Adobe Inc. Michiharu Ariza

// func TestExtentsTtVar(t *testing.T) {
// 	face := hb_test_open_font_file("fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
// 	font := NewFont(face)
// 	//    hb_ot_font_set_funcs (font);

// 	//    hb_glyph_extents_t  extents;
// 	extents, result := hb_font_get_glyph_extents(font, 2)
// 	assert(t, result)

// 	assertEqualInt(t, extents.x_bearing, 10)
// 	assertEqualInt(t, extents.y_bearing, 846)
// 	assertEqualInt(t, extents.width, 500)
// 	assertEqualInt(t, extents.height, -846)

// 	coords := [1]float32{500.0}
// 	hb_font_set_var_coords_design(font, coords, 1)
// 	result = hb_font_get_glyph_extents(font, 2, &extents)
// 	assert(t, result)

// 	assertEqualInt(t, extents.x_bearing, 0)
// 	assertEqualInt(t, extents.y_bearing, 874)
// 	assertEqualInt(t, extents.width, 551)
// 	assertEqualInt(t, extents.height, -874)
// }

// func TestadvanceTtVarNohvar(t *testing.T) {
// 	hb_face_t * face = hb_test_open_font_file("fonts/SourceSansVariable-Roman-nohvar-41,C1.ttf")
// 	assert(t, face)
// 	hb_font_t * font = hb_font_create(face)
// 	hb_face_destroy(face)
// 	assert(t, font)
// 	hb_ot_font_set_funcs(font)

// 	x, y := hb_font_get_glyph_advance_for_direction(font, 2, HB_DIRECTION_LTR, &x, &y)

// 	assertEqualInt(t, x, 520)
// 	assertEqualInt(t, y, 0)

// 	x, y = hb_font_get_glyph_advance_for_direction(font, 2, HB_DIRECTION_TTB, &x, &y)

// 	assertEqualInt(t, x, 0)
// 	assertEqualInt(t, y, -1000)

// 	coords = [1]float32{500.0}
// 	hb_font_set_var_coords_design(font, coords, 1)
// 	hb_font_get_glyph_advance_for_direction(font, 2, HB_DIRECTION_LTR, &x, &y)

// 	assertEqualInt(t, x, 551)
// 	assertEqualInt(t, y, 0)

// 	hb_font_get_glyph_advance_for_direction(font, 2, HB_DIRECTION_TTB, &x, &y)

// 	assertEqualInt(t, x, 0)
// 	assertEqualInt(t, y, -1000)
// }

// func TestadvanceTtVarHvarvvar(t *testing.T) {
// 	hb_face_t * face = hb_test_open_font_file("fonts/SourceSerifVariable-Roman-VVAR.abc.ttf")
// 	assert(t, face)
// 	hb_font_t * font = hb_font_create(face)
// 	hb_face_destroy(face)
// 	assert(t, font)
// 	hb_ot_font_set_funcs(font)

// 	x, y = hb_font_get_glyph_advance_for_direction(font, 1, HB_DIRECTION_LTR, &x, &y)

// 	assertEqualInt(t, x, 508)
// 	assertEqualInt(t, y, 0)

// 	hb_font_get_glyph_advance_for_direction(font, 1, HB_DIRECTION_TTB, &x, &y)

// 	assertEqualInt(t, x, 0)
// 	assertEqualInt(t, y, -1000)

// 	coords := [1]float32{700.0}
// 	hb_font_set_var_coords_design(font, coords, 1)
// 	x, y = hb_font_get_glyph_advance_for_direction(font, 1, HB_DIRECTION_LTR, &x, &y)

// 	assertEqualInt(t, x, 531)
// 	assertEqualInt(t, y, 0)

// 	x, y = hb_font_get_glyph_advance_for_direction(font, 1, HB_DIRECTION_TTB, &x, &y)

// 	assertEqualInt(t, x, 0)
// 	assertEqualInt(t, y, -1012)
// }

// func TestadvanceTtVarAnchor(t *testing.T) {
// 	hb_face_t * face = hb_test_open_font_file("fonts/SourceSansVariable-Roman.anchor.ttf")
// 	assert(t, face)
// 	hb_font_t * font = hb_font_create(face)
// 	hb_face_destroy(face)
// 	assert(t, font)
// 	hb_ot_font_set_funcs(font)

// 	extents, result = hb_font_get_glyph_extents(font, 2, &extents)
// 	assert(t, result)

// 	assertEqualInt(t, extents.x_bearing, 56)
// 	assertEqualInt(t, extents.y_bearing, 672)
// 	assertEqualInt(t, extents.width, 556)
// 	assertEqualInt(t, extents.height, -684)

// 	coords := [1]float32{500.0}
// 	hb_font_set_var_coords_design(font, coords, 1)
// 	result = hb_font_get_glyph_extents(font, 2, &extents)
// 	assert(t, result)

// 	assertEqualInt(t, extents.x_bearing, 50)
// 	assertEqualInt(t, extents.y_bearing, 667)
// 	assertEqualInt(t, extents.width, 593)
// 	assertEqualInt(t, extents.height, -679)
// }

// func TestextentsTtVarComp(t *testing.T) {
// 	hb_face_t * face = hb_test_open_font_file("fonts/SourceSansVariable-Roman.modcomp.ttf")
// 	assert(t, face)
// 	hb_font_t * font = hb_font_create(face)
// 	hb_face_destroy(face)
// 	assert(t, font)
// 	hb_ot_font_set_funcs(font)

// coords:
// 	[1]float32{800.0}
// 	hb_font_set_var_coords_design(font, coords, 1)

// 	//    hb_bool_t result;
// 	extents, result = hb_font_get_glyph_extents(font, 2, &extents) /* Ccedilla, cedilla y-scaled by 0.8, with unscaled component offset */
// 	assert(t, result)

// 	assertEqualInt(t, extents.x_bearing, 19)
// 	assertEqualInt(t, extents.y_bearing, 663)
// 	assertEqualInt(t, extents.width, 519)
// 	assertEqualInt(t, extents.height, -895)

// 	result = hb_font_get_glyph_extents(font, 3, &extents) /* Cacute, acute y-scaled by 0.8, with unscaled component offset (default) */
// 	assert(t, result)

// 	assertEqualInt(t, extents.x_bearing, 19)
// 	assertEqualInt(t, extents.y_bearing, 909)
// 	assertEqualInt(t, extents.width, 519)
// 	assertEqualInt(t, extents.height, -921)

// 	result = hb_font_get_glyph_extents(font, 4, &extents) /* Ccaron, caron y-scaled by 0.8, with scaled component offset */
// 	assert(t, result)

// 	assertEqualInt(t, extents.x_bearing, 19)
// 	assertEqualInt(t, extents.y_bearing, 866)
// 	assertEqualInt(t, extents.width, 519)
// 	assertEqualInt(t, extents.height, -878)
// }

// func TestadvanceTtVarCompV(t *testing.T) {
// 	hb_face_t * face = hb_test_open_font_file("fonts/SourceSansVariable-Roman.modcomp.ttf")
// 	assert(t, face)
// 	hb_font_t * font = hb_font_create(face)
// 	hb_face_destroy(face)
// 	assert(t, font)
// 	hb_ot_font_set_funcs(font)

// 	coords := [1]float32{800.0}
// 	hb_font_set_var_coords_design(font, coords, 1)

// 	x, y := hb_font_get_glyph_advance_for_direction(font, 2, HB_DIRECTION_TTB, &x, &y) /* No VVAR; 'C' in composite Ccedilla determines metrics */

// 	assertEqualInt(t, x, 0)
// 	assertEqualInt(t, y, -991)

// 	hb_font_get_glyph_origin_for_direction(font, 2, HB_DIRECTION_TTB, &x, &y)

// 	assertEqualInt(t, x, 292)
// 	assertEqualInt(t, y, 1013)
// }

// func TestadvanceTtVarGvarInfer(t *testing.T) {
// 	hb_face_t * face = hb_test_open_font_file("fonts/TestGVAREight.ttf")
// 	hb_font_t * font = hb_font_create(face)
// 	hb_ot_font_set_funcs(font)
// 	hb_face_destroy(face)

// 	coords := [6]int{100}
// 	hb_font_set_var_coords_normalized(font, coords, 6)

// 	extents := hb_glyph_extents_t{0}
// 	assert(t, hb_font_get_glyph_extents(font, 4, &extents))
// }
