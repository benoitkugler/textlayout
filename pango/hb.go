package pango

// TODO:
func pango_hb_shape(font Font, item_text []rune, analysis *Analysis, glyphs *GlyphString, paragraph_text []rune) {
}

type hb_font_t struct {
	//   hb_object_header_t header;

	parent *hb_font_t
	// face *hb_face_t

	x_scale int32
	y_scale int32
	x_mult  int64
	y_mult  int64

	x_ppem uint
	y_ppem uint

	ptem float32

	/* Font variation coordinates. */
	//   unsigned int num_coords;
	coords        []int
	design_coords []float32
}

// pango_font_get_hb_font: (skip)
// @font: a #PangoFont
//
// Get a hb_font_t object backing this font.
//
// Note that the objects returned by this function
// are cached and immutable. If you need to make
// changes to the hb_font_t, use hb_font_create_sub_font().
//
// Returns: (transfer none) (nullable): the hb_font_t object backing the
//          font, or %NULL if the font does not have one
// TODO:
func pango_font_get_hb_font(font Font) *hb_font_t {
	return nil
	// PangoFontPrivate * priv = pango_font_get_instance_private(font)

	// g_return_val_if_fail(PANGO_IS_FONT(font), NULL)

	// if priv.hb_font {
	// 	return priv.hb_font
	// }

	// priv.hb_font = PANGO_FONT_GET_CLASS(font).create_hb_font(font)

	// hb_font_make_immutable(priv.hb_font)

	// return priv.hb_font
}

// TODO:
func hb_font_get_nominal_glyph(font *hb_font_t, u rune) (rune, bool) {
	// return font.get_nominal_glyph(unicode, glyph)
	return 0, false
}

// TODO:
func hb_font_get_glyph_h_advance(font *hb_font_t, glyph rune) int32 {
	//    return font->get_glyph_h_advance (glyph);
	return 0
}
