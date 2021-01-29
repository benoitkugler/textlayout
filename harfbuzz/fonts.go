package harfbuzz

import (
	"math"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from src/hb-font.hh, src/hb-font.cc  Copyright © 2009  Red Hat, Inc., 2012  Google, Inc.  Behdad Esfahbod

const faceUpem = 1000

var emptyFont = hb_font_t{
	//    const_cast<hb_face_t *> (&_hb_Null_hb_face_t), // TODO: "empty face"
	x_scale: 1000,    // x_scale
	y_scale: 1000,    // y_scale
	x_mult:  1 << 16, // x_mult
	y_mult:  1 << 16, // y_mult
}

// hb_font_extents_t exposes font-wide extent values, measured in font units.
// Note that typically ascender is positive and descender negative in coordinate systems that grow up.
// TODO: use plain ints if possible
type hb_font_extents_t struct {
	Ascender  hb_position_t // typographic ascender.
	Descender hb_position_t // typographic descender.
	LineGap   hb_position_t // suggested line spacing gap.
}

type hb_face_t interface {
	// common

	// Return the glyph used to represent the given rune,
	// or false if not found.
	GetNominalGlyph(ch rune) (fonts.GlyphIndex, bool)

	// Returns the horizontal advance, or false if no
	// advance is found an a defaut value should be used.
	// `coords` is used by variable fonts, and specified in normalized coordinates.
	GetHorizontalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool)

	// Same as `GetHorizontalAdvance`, but for vertical advance
	GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool)

	// Returns the extents of the font for horizontal text, or false
	// it not available.
	GetFontHExtents() (hb_font_extents_t, bool)

	// Fetches the (X,Y) coordinates of the origin (in font units) for a glyph ID,
	// for horizontal text segments.
	// Returns `false` if not available.
	GetGlyphHOrigin(fonts.GlyphIndex) (x, y hb_position_t, found bool)

	// Same as `GetGlyphHOrigin`, but for vertical text segments.
	GetGlyphVOrigin(fonts.GlyphIndex) (x, y hb_position_t, found bool)

	// specialized
	get_gsubgpos_table() (gsub, gpos *truetype.TableLayout) // optional
	getGDEF() truetype.TableGDEF                            // optional
	getKerx() (interface{}, bool)                           // optional
	getKerns() truetype.Kerns                               // optional
	hasMachineKerning() bool                                // TODO
	hasCrossKerning() bool                                  // TODO:
	hasTrackTable() bool                                    // TODO:
	getFeatTable() bool
	// return the variations_index
	hb_ot_layout_table_find_feature_variations(table_tag hb_tag_t, coords []float32) int
	Normalize(coords []float32) []float32
}

// hb_font_t represents a font face at a specific size and with
// certain other parameters (pixels-per-em, points-per-em, variation
// settings) specified. Font objects are created from font face
// objects, and are used as input to hb_shape(), among other things.
//
// Client programs can optionally pass in their own functions that
// implement the basic, lower-level queries of font objects. This set
// of font functions is defined by the virtual methods in
// #hb_font_funcs_t.
//
// HarfBuzz provides a built-in set of lightweight default
// functions for each method in #hb_font_funcs_t.
type hb_font_t struct {
	face hb_face_t

	x_scale, y_scale int32
	x_mult, y_mult   int64 // cached value of  (x_scale << 16) / faceUpem

	x_ppem, y_ppem uint

	ptem float32

	// font variation coordinates (optionnal)
	coords        []float32 // length num_coords, normalized
	design_coords []float32 // length num_coords, in design units
}

/* Convert from font-space to user-space */
//    int64 dir_mult (hb_direction_t direction) { return HB_DIRECTION_IS_VERTICAL(direction) ? y_mult : x_mult; }
func (f hb_font_t) em_scale_x(v int16) hb_position_t    { return em_mult(v, f.x_mult) }
func (f hb_font_t) em_scale_y(v int16) hb_position_t    { return em_mult(v, f.y_mult) }
func (f hb_font_t) em_scalef_x(v float32) hb_position_t { return em_scalef(v, f.x_scale) }
func (f hb_font_t) em_scalef_y(v float32) hb_position_t { return em_scalef(v, f.y_scale) }
func (f hb_font_t) em_fscale_x(v int16) float32         { return em_fscale(v, f.x_scale) }
func (f hb_font_t) em_fscale_y(v int16) float32         { return em_fscale(v, f.y_scale) }

func (f *hb_font_t) mults_changed() {
	f.x_mult = (int64(f.x_scale) << 16) / faceUpem
	f.y_mult = (int64(f.y_scale) << 16) / faceUpem
}

func em_mult(v int16, mult int64) hb_position_t {
	return hb_position_t((int64(v) * mult) >> 16)
}
func em_scalef(v float32, scale int32) hb_position_t {
	return hb_position_t(math.Round(float64(v * float32(scale) / faceUpem)))
}
func em_fscale(v int16, scale int32) float32 {
	return float32(v) * float32(scale) / faceUpem
}

// Fetches the advance for a glyph ID from the specified font,
// in a text segment of the specified direction.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f hb_font_t) get_glyph_advance_for_direction(glyph fonts.GlyphIndex, dir hb_direction_t) (x, y hb_position_t) {
	if dir.isHorizontal() {
		return f.get_glyph_h_advance(glyph), 0
	}
	return 0, f.get_glyph_v_advance(glyph)
}

// Fetches the advance for a glyph ID in the specified font,
// for horizontal text segments.
func (f *hb_font_t) get_glyph_h_advance(glyph fonts.GlyphIndex) hb_position_t {
	adv, has := f.face.GetHorizontalAdvance(glyph)
	if !has {
		adv = faceUpem
	}
	return f.em_scale_x(adv)
}

// Fetches the advance for a glyph ID in the specified font,
// for vertical text segments.
func (f *hb_font_t) get_glyph_v_advance(glyph fonts.GlyphIndex) hb_position_t {
	adv, has := f.face.GetVerticalAdvance(glyph)
	if !has {
		adv = faceUpem
	}
	return f.em_scale_y(adv)
}

// Subtracts the origin coordinates from an (X,Y) point coordinate,
// in the specified glyph ID in the specified font.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f *hb_font_t) subtract_glyph_origin_for_direction(glyph fonts.GlyphIndex, direction hb_direction_t,
	x, y hb_position_t) (hb_position_t, hb_position_t) {

	origin_x, origin_y := f.get_glyph_origin_for_direction(glyph, direction)

	return x - origin_x, y - origin_y
}

// Fetches the (X,Y) coordinates of the origin for a glyph in
// the specified font.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f *hb_font_t) get_glyph_origin_for_direction(glyph fonts.GlyphIndex, direction hb_direction_t) (x, y hb_position_t) {
	if direction.isHorizontal() {
		return f.get_glyph_h_origin_with_fallback(glyph)
	}
	return f.get_glyph_v_origin_with_fallback(glyph)
}

func (f *hb_font_t) get_glyph_h_origin_with_fallback(glyph fonts.GlyphIndex) (hb_position_t, hb_position_t) {
	x, y, ok := f.face.GetGlyphHOrigin(glyph)
	if !ok {
		x, y, ok = f.face.GetGlyphVOrigin(glyph)
		if ok {
			dx, dy := f.guess_v_origin_minus_h_origin(glyph)
			return x - dx, y - dy
		}
	}
	return x, y
}

func (f *hb_font_t) get_glyph_v_origin_with_fallback(glyph fonts.GlyphIndex) (hb_position_t, hb_position_t) {
	x, y, ok := f.face.GetGlyphVOrigin(glyph)
	if !ok {
		x, y, ok = f.face.GetGlyphHOrigin(glyph)
		if ok {
			dx, dy := f.guess_v_origin_minus_h_origin(glyph)
			return x + dx, y + dy
		}
	}
	return x, y
}

func (f *hb_font_t) guess_v_origin_minus_h_origin(glyph fonts.GlyphIndex) (x, y hb_position_t) {
	x = f.get_glyph_h_advance(glyph) / 2
	extents := f.get_h_extents_with_fallback()
	y = extents.Ascender
	return x, y
}

func (f *hb_font_t) get_h_extents_with_fallback() hb_font_extents_t {
	extents, ok := f.face.GetFontHExtents()
	if !ok {
		extents.Ascender = f.y_scale * 4 / 5
		extents.Descender = extents.Ascender - f.y_scale
		extents.LineGap = 0
	}
	return extents
}

//    hb_position_t em_scale_dir (int16 v, hb_direction_t direction)
//    { return em_mult (v, dir_mult (direction)); }

//    /* Convert from parent-font user-space to our user-space */
//    hb_position_t parent_scale_x_distance (hb_position_t v)
//    {
// 	 if (unlikely (parent && parent.x_scale != x_scale))
// 	   return (hb_position_t) (v * (int64) this.x_scale / this.parent.x_scale);
// 	 return v;
//    }
//    hb_position_t parent_scale_y_distance (hb_position_t v)
//    {
// 	 if (unlikely (parent && parent.y_scale != y_scale))
// 	   return (hb_position_t) (v * (int64) this.y_scale / this.parent.y_scale);
// 	 return v;
//    }
//    hb_position_t parent_scale_x_position (hb_position_t v)
//    { return parent_scale_x_distance (v); }
//    hb_position_t parent_scale_y_position (hb_position_t v)
//    { return parent_scale_y_distance (v); }

//    void parent_scale_distance (hb_position_t *x, hb_position_t *y)
//    {
// 	 *x = parent_scale_x_distance (*x);
// 	 *y = parent_scale_y_distance (*y);
//    }
//    void parent_scale_position (hb_position_t *x, hb_position_t *y)
//    {
// 	 *x = parent_scale_x_position (*x);
// 	 *y = parent_scale_y_position (*y);
//    }

/* Public getters */

//    hb_bool_t get_font_h_extents (hb_font_extents_t *extents)
//    {
// 	 memset (extents, 0, sizeof (*extents));
// 	 return klass.get.f.font_h_extents (this, user_data,
// 					 extents,
// 					 klass.user_data.font_h_extents);
//    }
//    hb_bool_t get_font_v_extents (hb_font_extents_t *extents)
//    {
// 	 memset (extents, 0, sizeof (*extents));
// 	 return klass.get.f.font_v_extents (this, user_data,
// 					 extents,
// 					 klass.user_data.font_v_extents);
//    }

//    bool has_glyph (hb_codepoint_t unicode)
//    {
// 	 hb_codepoint_t glyph;
// 	 return get_nominal_glyph (unicode, &glyph);
//    }

//    hb_bool_t get_nominal_glyph (hb_codepoint_t unicode,
// 					hb_codepoint_t *glyph)
//    {
// 	 *glyph = 0;
// 	 return klass.get.f.nominal_glyph (this, user_data,
// 						unicode, glyph,
// 						klass.user_data.nominal_glyph);
//    }
//    unsigned int get_nominal_glyphs (unsigned int count,
// 					const hb_codepoint_t *first_unicode,
// 					unsigned int unicode_stride,
// 					hb_codepoint_t *first_glyph,
// 					unsigned int glyph_stride)
//    {
// 	 return klass.get.f.nominal_glyphs (this, user_data,
// 					 count,
// 					 first_unicode, unicode_stride,
// 					 first_glyph, glyph_stride,
// 					 klass.user_data.nominal_glyphs);
//    }

//    hb_bool_t get_variation_glyph (hb_codepoint_t unicode, hb_codepoint_t variation_selector,
// 				  hb_codepoint_t *glyph)
//    {
// 	 *glyph = 0;
// 	 return klass.get.f.variation_glyph (this, user_data,
// 					  unicode, variation_selector, glyph,
// 					  klass.user_data.variation_glyph);
//    }

//    hb_position_t get_glyph_h_advance (hb_codepoint_t glyph)
//    {
// 	 return klass.get.f.glyph_h_advance (this, user_data,
// 					  glyph,
// 					  klass.user_data.glyph_h_advance);
//    }

//    hb_position_t get_glyph_v_advance (hb_codepoint_t glyph)
//    {
// 	 return klass.get.f.glyph_v_advance (this, user_data,
// 					  glyph,
// 					  klass.user_data.glyph_v_advance);
//    }

//    void get_glyph_h_advances (unsigned int count,
// 				  const hb_codepoint_t *first_glyph,
// 				  unsigned int glyph_stride,
// 				  hb_position_t *first_advance,
// 				  unsigned int advance_stride)
//    {
// 	 return klass.get.f.glyph_h_advances (this, user_data,
// 					   count,
// 					   first_glyph, glyph_stride,
// 					   first_advance, advance_stride,
// 					   klass.user_data.glyph_h_advances);
//    }

//    void get_glyph_v_advances (unsigned int count,
// 				  const hb_codepoint_t *first_glyph,
// 				  unsigned int glyph_stride,
// 				  hb_position_t *first_advance,
// 				  unsigned int advance_stride)
//    {
// 	 return klass.get.f.glyph_v_advances (this, user_data,
// 					   count,
// 					   first_glyph, glyph_stride,
// 					   first_advance, advance_stride,
// 					   klass.user_data.glyph_v_advances);
//    }

//    hb_bool_t get_glyph_h_origin (hb_codepoint_t glyph,
// 				 hb_position_t *x, hb_position_t *y)
//    {
// 	 *x = *y = 0;
// 	 return klass.get.f.glyph_h_origin (this, user_data,
// 					 glyph, x, y,
// 					 klass.user_data.glyph_h_origin);
//    }

//    hb_bool_t get_glyph_v_origin (hb_codepoint_t glyph,
// 				 hb_position_t *x, hb_position_t *y)
//    {
// 	 *x = *y = 0;
// 	 return klass.get.f.glyph_v_origin (this, user_data,
// 					 glyph, x, y,
// 					 klass.user_data.glyph_v_origin);
//    }

//    hb_position_t get_glyph_h_kerning (hb_codepoint_t left_glyph,
// 					  hb_codepoint_t right_glyph)
//    {
//  #ifdef HB_DISABLE_DEPRECATED
// 	 return 0;
//  #else
// 	 return klass.get.f.glyph_h_kerning (this, user_data,
// 					  left_glyph, right_glyph,
// 					  klass.user_data.glyph_h_kerning);
//  #endif
//    }

//    hb_position_t get_glyph_v_kerning (hb_codepoint_t top_glyph,
// 					  hb_codepoint_t bottom_glyph)
//    {
//  #ifdef HB_DISABLE_DEPRECATED
// 	 return 0;
//  #else
// 	 return klass.get.f.glyph_v_kerning (this, user_data,
// 					  top_glyph, bottom_glyph,
// 					  klass.user_data.glyph_v_kerning);
//  #endif
//    }

//    hb_bool_t get_glyph_extents (hb_codepoint_t glyph,
// 					hb_glyph_extents_t *extents)
//    {
// 	 memset (extents, 0, sizeof (*extents));
// 	 return klass.get.f.glyph_extents (this, user_data,
// 						glyph,
// 						extents,
// 						klass.user_data.glyph_extents);
//    }

//    hb_bool_t get_glyph_contour_point (hb_codepoint_t glyph, unsigned int point_index,
// 					  hb_position_t *x, hb_position_t *y)
//    {
// 	 *x = *y = 0;
// 	 return klass.get.f.glyph_contour_point (this, user_data,
// 						  glyph, point_index,
// 						  x, y,
// 						  klass.user_data.glyph_contour_point);
//    }

//    hb_bool_t get_glyph_name (hb_codepoint_t glyph,
// 				 char *name, unsigned int size)
//    {
// 	 if (size) *name = '\0';
// 	 return klass.get.f.glyph_name (this, user_data,
// 					 glyph,
// 					 name, size,
// 					 klass.user_data.glyph_name);
//    }

//    hb_bool_t get_glyph_from_name (const char *name, int len, /* -1 means nul-terminated */
// 				  hb_codepoint_t *glyph)
//    {
// 	 *glyph = 0;
// 	 if (len == -1) len = strlen (name);
// 	 return klass.get.f.glyph_from_name (this, user_data,
// 					  name, len,
// 					  glyph,
// 					  klass.user_data.glyph_from_name);
//    }

//    /* A bit higher-level, and with fallback */

//    void get_v_extents_with_fallback (hb_font_extents_t *extents)
//    {
// 	 if (!get_font_v_extents (extents))
// 	 {
// 	   extents.Ascender = x_scale / 2;
// 	   extents.descender = extents.Ascender - x_scale;
// 	   extents.line_gap = 0;
// 	 }
//    }

//    void get_extents_for_direction (hb_direction_t direction,
// 				   hb_font_extents_t *extents)
//    {
// 	 if (likely (HB_DIRECTION_IS_HORIZONTAL (direction)))
// 	   get_h_extents_with_fallback (extents);
// 	 else
// 	   get_v_extents_with_fallback (extents);
//    }

//    void get_glyph_advances_for_direction (hb_direction_t direction,
// 					  unsigned int count,
// 					  const hb_codepoint_t *first_glyph,
// 					  unsigned glyph_stride,
// 					  hb_position_t *first_advance,
// 					  unsigned advance_stride)
//    {
// 	 if (likely (HB_DIRECTION_IS_HORIZONTAL (direction)))
// 	   get_glyph_h_advances (count, first_glyph, glyph_stride, first_advance, advance_stride);
// 	 else
// 	   get_glyph_v_advances (count, first_glyph, glyph_stride, first_advance, advance_stride);
//    }

//    void add_glyph_h_origin (hb_codepoint_t glyph,
// 				hb_position_t *x, hb_position_t *y)
//    {
// 	 hb_position_t origin_x, origin_y;

// 	 get_glyph_h_origin_with_fallback (glyph, &origin_x, &origin_y);

// 	 *x += origin_x;
// 	 *y += origin_y;
//    }
//    void add_glyph_v_origin (hb_codepoint_t glyph,
// 				hb_position_t *x, hb_position_t *y)
//    {
// 	 hb_position_t origin_x, origin_y;

// 	 get_glyph_v_origin_with_fallback (glyph, &origin_x, &origin_y);

// 	 *x += origin_x;
// 	 *y += origin_y;
//    }
//    void add_glyph_origin_for_direction (hb_codepoint_t glyph,
// 						hb_direction_t direction,
// 						hb_position_t *x, hb_position_t *y)
//    {
// 	 hb_position_t origin_x, origin_y;

// 	 get_glyph_origin_for_direction (glyph, direction, &origin_x, &origin_y);

// 	 *x += origin_x;
// 	 *y += origin_y;
//    }

//    void subtract_glyph_h_origin (hb_codepoint_t glyph,
// 				 hb_position_t *x, hb_position_t *y)
//    {
// 	 hb_position_t origin_x, origin_y;

// 	 get_glyph_h_origin_with_fallback (glyph, &origin_x, &origin_y);

// 	 *x -= origin_x;
// 	 *y -= origin_y;
//    }
//    void subtract_glyph_v_origin (hb_codepoint_t glyph,
// 				 hb_position_t *x, hb_position_t *y)
//    {
// 	 hb_position_t origin_x, origin_y;

// 	 get_glyph_v_origin_with_fallback (glyph, &origin_x, &origin_y);

// 	 *x -= origin_x;
// 	 *y -= origin_y;
//    }

//    void get_glyph_kerning_for_direction (hb_codepoint_t first_glyph, hb_codepoint_t second_glyph,
// 					 hb_direction_t direction,
// 					 hb_position_t *x, hb_position_t *y)
//    {
// 	 if (likely (HB_DIRECTION_IS_HORIZONTAL (direction))) {
// 	   *y = 0;
// 	   *x = get_glyph_h_kerning (first_glyph, second_glyph);
// 	 } else {
// 	   *x = 0;
// 	   *y = get_glyph_v_kerning (first_glyph, second_glyph);
// 	 }
//    }

//    hb_bool_t get_glyph_extents_for_origin (hb_codepoint_t glyph,
// 					   hb_direction_t direction,
// 					   hb_glyph_extents_t *extents)
//    {
// 	 hb_bool_t ret = get_glyph_extents (glyph, extents);

// 	 if (ret)
// 	   subtract_glyph_origin_for_direction (glyph, direction, &extents.x_bearing, &extents.y_bearing);

// 	 return ret;
//    }

//    hb_bool_t get_glyph_contour_point_for_origin (hb_codepoint_t glyph, unsigned int point_index,
// 						 hb_direction_t direction,
// 						 hb_position_t *x, hb_position_t *y)
//    {
// 	 hb_bool_t ret = get_glyph_contour_point (glyph, point_index, x, y);

// 	 if (ret)
// 	   subtract_glyph_origin_for_direction (glyph, direction, x, y);

// 	 return ret;
//    }

//    /* Generates gidDDD if glyph has no name. */
//    void
//    glyph_to_string (hb_codepoint_t glyph,
// 			char *s, unsigned int size)
//    {
// 	 if (get_glyph_name (glyph, s, size)) return;

// 	 if (size && snprintf (s, size, "gid%u", glyph) < 0)
// 	   *s = '\0';
//    }

//    /* Parses gidDDD and uniUUUU strings automatically. */
//    hb_bool_t
//    glyph_from_string (const char *s, int len, /* -1 means nul-terminated */
// 			  hb_codepoint_t *glyph)
//    {
// 	 if (get_glyph_from_name (s, len, glyph)) return true;

// 	 if (len == -1) len = strlen (s);

// 	 /* Straight glyph index. */
// 	 if (hb_codepoint_parse (s, len, 10, glyph))
// 	   return true;

// 	 if (len > 3)
// 	 {
// 	   /* gidDDD syntax for glyph indices. */
// 	   if (0 == strncmp (s, "gid", 3) &&
// 	   hb_codepoint_parse (s + 3, len - 3, 10, glyph))
// 	 return true;

// 	   /* uniUUUU and other Unicode character indices. */
// 	   hb_codepoint_t unichar;
// 	   if (0 == strncmp (s, "uni", 3) &&
// 	   hb_codepoint_parse (s + 3, len - 3, 16, &unichar) &&
// 	   get_nominal_glyph (unichar, glyph))
// 	 return true;
// 	 }

// 	 return false;
//    }

// Constructs a new font object from the specified face.
func hb_font_create(face hb_face_t) *hb_font_t {
	var font hb_font_t

	font.parent = &emptyFont
	font.face = face
	font.x_scale = faceUpem
	font.y_scale = faceUpem
	font.x_mult = 1 << 16
	font.y_mult = 1 << 16

	return &font
}

/**
 * hb_font_create_sub_font:
 * @parent: The parent font object
 *
 * Constructs a sub-font font object from the specified @parent font,
 * replicating the parent's properties.
 *
 * Return value: (transfer full): The new sub-font font object
 *
 * Since: 0.9.2
 **/

func (parent *hb_font_t) hb_font_create_sub_font() *hb_font_t {
	font := hb_font_create(parent.face)

	font.parent = parent

	font.x_scale = parent.x_scale
	font.y_scale = parent.y_scale
	font.mults_changed()
	font.x_ppem = parent.x_ppem
	font.y_ppem = parent.y_ppem
	font.ptem = parent.ptem

	// deepcopy
	font.coords = append(font.coords, parent.coords...)
	font.design_coords = append(font.design_coords, parent.design_coords...)

	return font
}

//  static hb_bool_t
//  hb_font_get_font_h_extents_nil (font *hb_font_t HB_UNUSED,
// 				 void              *font_data HB_UNUSED,
// 				 extents *hb_font_extents_t,
// 				 void              *user_data HB_UNUSED)
//  {
//    memset (extents, 0, sizeof (*extents));
//    return false;
//  }

// Fetches the extents for a specified font, for horizontal
// text segments.
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_font_h_extents_default(extents *hb_font_extents_t) bool {
	ret := font.parent.get_font_h_extents(extents)
	if ret {
		extents.Ascender = font.parent_scale_y_distance(extents.Ascender)
		extents.descender = font.parent_scale_y_distance(extents.descender)
		extents.line_gap = font.parent_scale_y_distance(extents.line_gap)
	}
	return ret
}

//  static hb_bool_t
//  hb_font_get_font_v_extents_nil (font *hb_font_t HB_UNUSED,
// 				 void              *font_data HB_UNUSED,
// 				 extents *hb_font_extents_t,
// 				 void              *user_data HB_UNUSED)
//  {
//    memset (extents, 0, sizeof (*extents));
//    return false;
//  }

// Fetches the extents for a specified font, for vertical
// text segments.
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_font_v_extents_default(extents *hb_font_extents_t) bool {
	ret := font.parent.get_font_v_extents(extents)
	if ret {
		extents.Ascender = font.parent_scale_x_distance(extents.Ascender)
		extents.descender = font.parent_scale_x_distance(extents.descender)
		extents.line_gap = font.parent_scale_x_distance(extents.line_gap)
	}
	return ret
}

//  static hb_bool_t
//  hb_font_get_nominal_glyph_nil (font *hb_font_t       HB_UNUSED,
// 					void           *font_data HB_UNUSED,
// 					rune  unicode HB_UNUSED,
// 					rune *glyph,
// 					void           *user_data HB_UNUSED)
//  {
//    *glyph = 0;
//    return false;
//  }

// Fetches the nominal glyph ID for a Unicode code point in the
// specified font.
//
// This version of the function should not be used to fetch glyph IDs
// for code points modified by variation selectors. For variation-selector
// support, user hb_font_get_variation_glyph() or use hb_font_get_glyph().
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_nominal_glyph_default(unicode rune) (glyphIndex, bool) {
	if font.has_nominal_glyphs_func_set() {
		return font.get_nominal_glyphs(1, &unicode, 0, glyph, 0)
	}
	return font.parent.get_nominal_glyph(unicode, glyph)
}

//  #define hb_font_get_nominal_glyphs_nil hb_font_get_nominal_glyphs_default

// @count: number of code points to query
// @first_unicode: The first Unicode code point to query
// @unicode_stride: The stride between successive code points
// @first_glyph: (out): The first glyph ID retrieved
// @glyph_stride: The stride between successive glyph IDs
//
// Fetches the nominal glyph IDs for a sequence of Unicode code points. Glyph
// IDs must be returned in a #rune output parameter.
//
// Return value: the number of code points processed
func (font *hb_font_t) hb_font_get_nominal_glyphs_default(count uint,
	first_unicode []rune, unicode_stride uint,
	first_glyph []rune, glyph_stride uint) uint {
	if font.has_nominal_glyph_func_set() {
		for i := 0; i < count; i++ {
			if !font.get_nominal_glyph(*first_unicode, first_glyph) {
				return i
			}

			//    first_unicode = &StructAtOffsetUnaligned<rune> (first_unicode, unicode_stride);
			//    first_glyph = &StructAtOffsetUnaligned<rune> (first_glyph, glyph_stride);
		}
		return count
	}

	return font.parent.get_nominal_glyphs(count,
		first_unicode, unicode_stride,
		first_glyph, glyph_stride)
}

//  static hb_bool_t
//  hb_font_get_variation_glyph_nil (font *hb_font_t       HB_UNUSED,
// 				  void           *font_data HB_UNUSED,
// 				  rune  unicode HB_UNUSED,
// 				  rune  variation_selector HB_UNUSED,
// 				  rune *glyph,
// 				  void           *user_data HB_UNUSED)
//  {
//    *glyph = 0;
//    return false;
//  }

// Fetches the glyph ID for a Unicode code point when followed by
// by the specified variation-selector code point, in the specified
// font.
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_variation_glyph_default(
	unicode, variation_selector rune) (glyphIndex, bool) {
	return font.parent.get_variation_glyph(unicode, variation_selector, glyph)
}

//  static hb_position_t
//  hb_font_get_glyph_h_advance_nil (font *hb_font_t      ,
// 				  void           *font_data HB_UNUSED,
// 				  rune  glyph HB_UNUSED,
// 				  void           *user_data HB_UNUSED)
//  {
//    return font.x_scale;
//  }

//  static hb_position_t
//  hb_font_get_glyph_v_advance_nil (font *hb_font_t      ,
// 				  void           *font_data HB_UNUSED,
// 				  rune  glyph HB_UNUSED,
// 				  void           *user_data HB_UNUSED)
//  {
//    /* TODO use font_extents.Ascender+descender */
//    return font.y_scale;
//  }

// Fetches the advance for a glyph ID in the specified font,
// for vertical text segments.
//
// Return value: The advance of @glyph within @font
func (font *hb_font_t) hb_font_get_glyph_v_advance_default(glyph glyphIndex) hb_position_t {
	if font.has_glyph_v_advances_func_set() {
		//  hb_position_t ret;
		font.get_glyph_v_advances(1, &glyph, 0, &ret, 0)
		return ret
	}
	return font.parent_scale_y_distance(font.parent.get_glyph_v_advance(glyph))
}

//  #define hb_font_get_glyph_h_advances_nil hb_font_get_glyph_h_advances_default

// @font: #hb_font_t to work upon
// @count: The number of glyph IDs in the sequence queried
// @first_glyph: The first glyph ID to query
// @glyph_stride: The stride between successive glyph IDs
// @first_advance: (out): The first advance retrieved
// @advance_stride: The stride between successive advances
//
// Fetches the advances for a sequence of glyph IDs in the specified
// font, for horizontal text segments.
func (font *hb_font_t) hb_font_get_glyph_h_advances_default(count uint,
	first_glyph []glyphIndex, glyph_stride uint,
	first_advance []hb_glyph_position_t, advance_stride uint) {
	if font.has_glyph_h_advance_func_set() {
		for i := 0; i < count; i++ {
			*first_advance = font.get_glyph_h_advance(*first_glyph)
			//    first_glyph = &StructAtOffsetUnaligned<rune> (first_glyph, glyph_stride);
			//    first_advance = &StructAtOffsetUnaligned<hb_position_t> (first_advance, advance_stride);
		}
		return
	}

	font.parent.get_glyph_h_advances(count,
		first_glyph, glyph_stride,
		first_advance, advance_stride)
	for i := 0; i < count; i++ {
		*first_advance = font.parent_scale_x_distance(*first_advance)
		//  first_advance = &StructAtOffsetUnaligned<hb_position_t> (first_advance, advance_stride);
	}
}

//  #define hb_font_get_glyph_v_advances_nil hb_font_get_glyph_v_advances_default

// @font: #hb_font_t to work upon
// @count: The number of glyph IDs in the sequence queried
// @first_glyph: The first glyph ID to query
// @glyph_stride: The stride between successive glyph IDs
// @first_advance: (out): The first advance retrieved
// @advance_stride: (out): The stride between successive advances
//
// Fetches the advances for a sequence of glyph IDs in the specified
// font, for vertical text segments.
func (font *hb_font_t) hb_font_get_glyph_v_advances_default(
	uint count,
	first_glyph []glyphIndex,
	uint glyph_stride,
	first_advance []hb_position_t,
	uint advance_stride) {
	if font.has_glyph_v_advance_func_set() {
		for i := 0; i < count; i++ {
			*first_advance = font.get_glyph_v_advance(*first_glyph)
			//    first_glyph = &StructAtOffsetUnaligned<rune> (first_glyph, glyph_stride);
			//    first_advance = &StructAtOffsetUnaligned<hb_position_t> (first_advance, advance_stride);
		}
		return
	}

	font.parent.get_glyph_v_advances(count,
		first_glyph, glyph_stride,
		first_advance, advance_stride)
	for i := 0; i < count; i++ {
		*first_advance = font.parent_scale_y_distance(*first_advance)
		//  first_advance = &StructAtOffsetUnaligned<hb_position_t> (first_advance, advance_stride);
	}
}

//  static hb_bool_t
//  hb_font_get_glyph_h_origin_nil (font *hb_font_t       HB_UNUSED,
// 				 void           *font_data HB_UNUSED,
// 				 rune  glyph HB_UNUSED,
// 				 hb_position_t  *x,
// 				 hb_position_t  *y,
// 				 void           *user_data HB_UNUSED)
//  {
//    *x = *y = 0;
//    return true;
//  }

// @font: #hb_font_t to work upon
// @glyph: The glyph ID to query
// @x: (out): The X coordinate of the origin
// @y: (out): The Y coordinate of the origin
//
// Fetches the (X,Y) coordinates of the origin for a glyph ID
// in the specified font, for horizontal text segments.
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_glyph_h_origin_default(
	glyph glyphIndex, x, y *hb_position_t) bool {
	ret := font.parent.get_glyph_h_origin(glyph, x, y)
	if ret {
		font.parent_scale_position(x, y)
	}
	return ret
}

//  static hb_bool_t
//  hb_font_get_glyph_v_origin_nil (font *hb_font_t       HB_UNUSED,
// 				 void           *font_data HB_UNUSED,
// 				 rune  glyph HB_UNUSED,
// 				 hb_position_t  *x,
// 				 hb_position_t  *y,
// 				 void           *user_data HB_UNUSED)
//  {
//    *x = *y = 0;
//    return false;
//  }

// @glyph: The glyph ID to query
// @x: (out): The X coordinate of the origin
// @y: (out): The Y coordinate of the origin
//
// Fetches the (X,Y) coordinates of the origin for a glyph ID
// in the specified font, for vertical text segments.
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_glyph_v_origin_default(
	glyph glyphIndex, x, y *hb_position_t) bool {
	ret := font.parent.get_glyph_v_origin(glyph, x, y)
	if ret {
		font.parent_scale_position(x, y)
	}
	return ret
}

//  static hb_position_t
//  hb_font_get_glyph_h_kerning_nil (font *hb_font_t       HB_UNUSED,
// 				  void           *font_data HB_UNUSED,
// 				  rune  left_glyph HB_UNUSED,
// 				  rune  right_glyph HB_UNUSED,
// 				  void           *user_data HB_UNUSED)
//  {
//    return 0;
//  }

// @left_glyph: The glyph ID of the left glyph in the glyph pair
// @right_glyph: The glyph ID of the right glyph in the glyph pair
//
// Fetches the kerning-adjustment value for a glyph-pair in
// the specified font, for horizontal text segments.
//
// <note>It handles legacy kerning only (as returned by the corresponding
// #hb_font_funcs_t function).</note>
//
// Return value: The kerning adjustment value
func (font *hb_font_t) hb_font_get_glyph_h_kerning_default(
	left_glyph, right_glyph glyphIndex) hb_position_t {
	return font.parent_scale_x_distance(font.parent.get_glyph_h_kerning(left_glyph, right_glyph))
}

//  static hb_bool_t
//  hb_font_get_glyph_extents_nil (font *hb_font_t HB_UNUSED,
// 					void               *font_data HB_UNUSED,
// 					rune      glyph HB_UNUSED,
// 					hb_glyph_extents_t *extents,
// 					void               *user_data HB_UNUSED)
//  {
//    memset (extents, 0, sizeof (*extents));
//    return false;
//  }

// @glyph: The glyph ID to query
// @extents: (out): The #hb_glyph_extents_t retrieved
//
// Fetches the #hb_glyph_extents_t data for a glyph ID
// in the specified font.
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_glyph_extents_default(glyph glyphIndex,
	hb_glyph_extents_t *extents) bool {
	ret := font.parent.get_glyph_extents(glyph, extents)
	if ret {
		font.parent_scale_position(&extents.x_bearing, &extents.y_bearing)
		font.parent_scale_distance(&extents.width, &extents.height)
	}
	return ret
}

//  static hb_bool_t
//  hb_font_get_glyph_contour_point_nil (font *hb_font_t       HB_UNUSED,
// 					  void           *font_data HB_UNUSED,
// 					  rune  glyph HB_UNUSED,
// 					  uint    point_index HB_UNUSED,
// 					  hb_position_t  *x,
// 					  hb_position_t  *y,
// 					  void           *user_data HB_UNUSED)
//  {
//    *x = *y = 0;
//    return false;
//  }

// @glyph: The glyph ID to query
// @point_index: The contour-point index to query
// @x: (out): The X value retrieved for the contour point
// @y: (out): The Y value retrieved for the contour point
//
// Fetches the (x,y) coordinates of a specified contour-point index
// in the specified glyph, within the specified font.
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_glyph_contour_point_default(glyph glyphIndex,
	point_index uint, x, y *hb_position_t) bool {
	ret := font.parent.get_glyph_contour_point(glyph, point_index, x, y)
	if ret {
		font.parent_scale_position(x, y)
	}
	return ret
}

//  static hb_bool_t
//  hb_font_get_glyph_name_nil (font *hb_font_t       HB_UNUSED,
// 				 void           *font_data HB_UNUSED,
// 				 rune  glyph HB_UNUSED,
// 				 char           *name,
// 				 uint    size,
// 				 void           *user_data HB_UNUSED)
//  {
//    if (size) *name = '\0';
//    return false;
//  }

// * @glyph: The glyph ID to query
// * @name: (out) (array length=size): Name string retrieved for the glyph ID
// * @size: Length of the glyph-name string retrieved
// *
// * Fetches the glyph-name string for a glyph ID in the specified @font.
// *
// * Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_glyph_name_default(glyph glyphIndex,
	name string, size uint) bool {
	return font.parent.get_glyph_name(glyph, name, size)
}

//  static hb_bool_t
//  hb_font_get_glyph_from_name_nil (font *hb_font_t       HB_UNUSED,
// 				  void           *font_data HB_UNUSED,
// 				  const char     *name HB_UNUSED,
// 				  int             len HB_UNUSED, /* -1 means nul-terminated */
// 				  rune *glyph,
// 				  void           *user_data HB_UNUSED)
//  {
//    *glyph = 0;
//    return false;
//  }

// @name: (array length=len): The name string to query
// @len: The length of the name queried
// @glyph: (out): The glyph ID retrieved
//
// Fetches the glyph ID that corresponds to a name string in the specified @font.
//
// <note>Note: @len == -1 means the name string is null-terminated.</note>
//
// Return value: `true` if data found, `false` otherwise
func (font *hb_font_t) hb_font_get_glyph_from_name_default(name string) (glyphIndex, bool) {
	return font.parent.get_glyph_from_name(name, len, glyph)
}

//  HB_FONT_FUNCS_IMPLEMENT_CALLBACKS
//  #undef HB_FONT_FUNC_IMPLEMENT

//  bool
//  hb_font_t::has_func_set (uint i)
//  {
//    return this.klass.get.array[i] != _hb_font_funcs_default.get.array[i];
//  }

//  bool
//  hb_font_t::has_func (uint i)
//  {
//    return has_func_set (i) ||
// 	  (parent && parent != &_hb_Null_hb_font_t && parent.has_func (i));
//  }

/* Public getters */

/**
 * hb_font_get_h_extents:
 * @font: #hb_font_t to work upon
 * @extents: (out): The font extents retrieved
 *
 * Fetches the extents for a specified font, for horizontal
 * text segments.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 1.1.3
 **/
//  hb_bool_t
//  hb_font_get_h_extents (font *hb_font_t,
// 				extents *hb_font_extents_t)
//  {
//    return font.get_font_h_extents (extents);
//  }

//  /**
//   * hb_font_get_v_extents:
//   * @font: #hb_font_t to work upon
//   * @extents: (out): The font extents retrieved
//   *
//   * Fetches the extents for a specified font, for vertical
//   * text segments.
//   *
//   * Return value: `true` if data found, `false` otherwise
//   *
//   * Since: 1.1.3
//   **/
//  hb_bool_t
//  hb_font_get_v_extents (font *hb_font_t,
// 				extents *hb_font_extents_t)
//  {
//    return font.get_font_v_extents (extents);
//  }

/**
 * hb_font_get_glyph:
 * @font: #hb_font_t to work upon
 * @unicode: The Unicode code point to query
 * @variation_selector: A variation-selector code point
 *
 * Fetches the glyph ID for a Unicode code point in the specified
 * font, with an optional variation selector.
 *
 * If @variation_selector is 0, calls hb_font_get_nominal_glyph();
 * otherwise calls hb_font_get_variation_glyph().
 *
 * Return value: `true` if data found, `false` otherwise
 **/
func hb_font_get_glyph(font *hb_font_t, unicode, variation_selector rune) (glyphIndex, bool) {
	if variation_selector {
		return font.get_variation_glyph(unicode, variation_selector, glyph)
	}
	return font.get_nominal_glyph(unicode, glyph)
}

//  /**
//   * hb_font_get_nominal_glyph:
//   * @font: #hb_font_t to work upon
//   * @unicode: The Unicode code point to query
//   * @glyph: (out): The glyph ID retrieved
//   *
//   * Fetches the nominal glyph ID for a Unicode code point in the
//   * specified font.
//   *
//   * This version of the function should not be used to fetch glyph IDs
//   * for code points modified by variation selectors. For variation-selector
//   * support, user hb_font_get_variation_glyph() or use hb_font_get_glyph().
//   *
//   * Return value: `true` if data found, `false` otherwise
//   *
//   * Since: 1.2.3
//   **/
//  hb_bool_t
//  hb_font_get_nominal_glyph (font *hb_font_t      ,
// 				rune  unicode,
// 				rune *glyph)
//  {
//    return font.get_nominal_glyph (unicode, glyph);
//  }

/**
 * hb_font_get_nominal_glyphs:
 * @font: #hb_font_t to work upon
 * @count: number of code points to query
 * @first_unicode: The first Unicode code point to query
 * @unicode_stride: The stride between successive code points
 * @first_glyph: (out): The first glyph ID retrieved
 * @glyph_stride: The stride between successive glyph IDs
 *
 * Fetches the nominal glyph IDs for a sequence of Unicode code points. Glyph
 * IDs must be returned in a #rune output parameter.
 *
 * Return value: the number of code points processed
 *
 * Since: 2.6.3
 **/
//  uint
//  hb_font_get_nominal_glyphs (hb_font_t *font,
// 				 uint count,
// 				 const rune *first_unicode,
// 				 uint unicode_stride,
// 				 rune *first_glyph,
// 				 uint glyph_stride)
//  {
//    return font.get_nominal_glyphs (count,
// 					first_unicode, unicode_stride,
// 					first_glyph, glyph_stride);
//  }

/**
 * hb_font_get_variation_glyph:
 * @font: #hb_font_t to work upon
 * @unicode: The Unicode code point to query
 * @variation_selector: The  variation-selector code point to query
 * @glyph: (out): The glyph ID retrieved
 *
 * Fetches the glyph ID for a Unicode code point when followed by
 * by the specified variation-selector code point, in the specified
 * font.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 1.2.3
 **/
//  hb_bool_t
//  hb_font_get_variation_glyph (font *hb_font_t      ,
// 				  rune  unicode,
// 				  rune  variation_selector,
// 				  rune *glyph)
//  {
//    return font.get_variation_glyph (unicode, variation_selector, glyph);
//  }

/**
 * hb_font_get_glyph_h_advance:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 *
 * Fetches the advance for a glyph ID in the specified font,
 * for horizontal text segments.
 *
 * Return value: The advance of @glyph within @font
 *
 * Since: 0.9.2
 **/
//  hb_position_t
//  hb_font_get_glyph_h_advance (font *hb_font_t      ,
// 				  rune  glyph)
//  {
//    return font.get_glyph_h_advance (glyph);
//  }

/**
 * hb_font_get_glyph_v_advance:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 *
 * Fetches the advance for a glyph ID in the specified font,
 * for vertical text segments.
 *
 * Return value: The advance of @glyph within @font
 *
 * Since: 0.9.2
 **/
//  hb_position_t
//  hb_font_get_glyph_v_advance (font *hb_font_t      ,
// 				  rune  glyph)
//  {
//    return font.get_glyph_v_advance (glyph);
//  }

/**
 * hb_font_get_glyph_h_advances:
 * @font: #hb_font_t to work upon
 * @count: The number of glyph IDs in the sequence queried
 * @first_glyph: The first glyph ID to query
 * @glyph_stride: The stride between successive glyph IDs
 * @first_advance: (out): The first advance retrieved
 * @advance_stride: The stride between successive advances
 *
 * Fetches the advances for a sequence of glyph IDs in the specified
 * font, for horizontal text segments.
 *
 * Since: 1.8.6
 **/
//  void
//  hb_font_get_glyph_h_advances (font *hb_font_t,
// 				   uint          count,
// 				   const rune *first_glyph,
// 				   unsigned              glyph_stride,
// 				   hb_position_t        *first_advance,
// 				   unsigned              advance_stride)
//  {
//    font.get_glyph_h_advances (count, first_glyph, glyph_stride, first_advance, advance_stride);
//  }
/**
 * hb_font_get_glyph_v_advances:
 * @font: #hb_font_t to work upon
 * @count: The number of glyph IDs in the sequence queried
 * @first_glyph: The first glyph ID to query
 * @glyph_stride: The stride between successive glyph IDs
 * @first_advance: (out): The first advance retrieved
 * @advance_stride: (out): The stride between successive advances
 *
 * Fetches the advances for a sequence of glyph IDs in the specified
 * font, for vertical text segments.
 *
 * Since: 1.8.6
 **/
//  void
//  hb_font_get_glyph_v_advances (font *hb_font_t,
// 				   uint          count,
// 				   const rune *first_glyph,
// 				   unsigned              glyph_stride,
// 				   hb_position_t        *first_advance,
// 				   unsigned              advance_stride)
//  {
//    font.get_glyph_v_advances (count, first_glyph, glyph_stride, first_advance, advance_stride);
//  }

/**
 * hb_font_get_glyph_h_origin:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @x: (out): The X coordinate of the origin
 * @y: (out): The Y coordinate of the origin
 *
 * Fetches the (X,Y) coordinates of the origin for a glyph ID
 * in the specified font, for horizontal text segments.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
//  hb_bool_t
//  hb_font_get_glyph_h_origin (font *hb_font_t      ,
// 				 rune  glyph,
// 				 hb_position_t  *x,
// 				 hb_position_t  *y)
//  {
//    return font.get_glyph_h_origin (glyph, x, y);
//  }

/**
 * hb_font_get_glyph_v_origin:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @x: (out): The X coordinate of the origin
 * @y: (out): The Y coordinate of the origin
 *
 * Fetches the (X,Y) coordinates of the origin for a glyph ID
 * in the specified font, for vertical text segments.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
//  hb_bool_t
//  hb_font_get_glyph_v_origin (font *hb_font_t      ,
// 				 rune  glyph,
// 				 hb_position_t  *x,
// 				 hb_position_t  *y)
//  {
//    return font.get_glyph_v_origin (glyph, x, y);
//  }

/**
 * hb_font_get_glyph_h_kerning:
 * @font: #hb_font_t to work upon
 * @left_glyph: The glyph ID of the left glyph in the glyph pair
 * @right_glyph: The glyph ID of the right glyph in the glyph pair
 *
 * Fetches the kerning-adjustment value for a glyph-pair in
 * the specified font, for horizontal text segments.
 *
 * <note>It handles legacy kerning only (as returned by the corresponding
 * #hb_font_funcs_t function).</note>
 *
 * Return value: The kerning adjustment value
 *
 * Since: 0.9.2
 **/
//  hb_position_t
//  hb_font_get_glyph_h_kerning (font *hb_font_t      ,
// 				  rune  left_glyph,
// 				  rune  right_glyph)
//  {
//    return font.get_glyph_h_kerning (left_glyph, right_glyph);
//  }

//  #ifndef HB_DISABLE_DEPRECATED
//  /**
//   * hb_font_get_glyph_v_kerning:
//   * @font: #hb_font_t to work upon
//   * @top_glyph: The glyph ID of the top glyph in the glyph pair
//   * @bottom_glyph: The glyph ID of the bottom glyph in the glyph pair
//   *
//   * Fetches the kerning-adjustment value for a glyph-pair in
//   * the specified font, for vertical text segments.
//   *
//   * <note>It handles legacy kerning only (as returned by the corresponding
//   * #hb_font_funcs_t function).</note>
//   *
//   * Return value: The kerning adjustment value
//   *
//   * Since: 0.9.2
//   * Deprecated: 2.0.0
//   **/
//  hb_position_t
//  hb_font_get_glyph_v_kerning (font *hb_font_t      ,
// 				  rune  top_glyph,
// 				  rune  bottom_glyph)
//  {
//    return font.get_glyph_v_kerning (top_glyph, bottom_glyph);
//  }
//  #endif

/**
 * hb_font_get_glyph_extents:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @extents: (out): The #hb_glyph_extents_t retrieved
 *
 * Fetches the #hb_glyph_extents_t data for a glyph ID
 * in the specified font.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
//  hb_bool_t
//  hb_font_get_glyph_extents (font *hb_font_t,
// 				rune      glyph,
// 				hb_glyph_extents_t *extents)
//  {
//    return font.get_glyph_extents (glyph, extents);
//  }

/**
 * hb_font_get_glyph_contour_point:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @point_index: The contour-point index to query
 * @x: (out): The X value retrieved for the contour point
 * @y: (out): The Y value retrieved for the contour point
 *
 * Fetches the (x,y) coordinates of a specified contour-point index
 * in the specified glyph, within the specified font.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
//  hb_bool_t
//  hb_font_get_glyph_contour_point (font *hb_font_t      ,
// 				  rune  glyph,
// 				  uint    point_index,
// 				  hb_position_t  *x,
// 				  hb_position_t  *y)
//  {
//    return font.get_glyph_contour_point (glyph, point_index, x, y);
//  }

/**
 * hb_font_get_glyph_name:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @name: (out) (array length=size): Name string retrieved for the glyph ID
 * @size: Length of the glyph-name string retrieved
 *
 * Fetches the glyph-name string for a glyph ID in the specified @font.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
//  hb_bool_t
//  hb_font_get_glyph_name (font *hb_font_t      ,
// 			 rune  glyph,
// 			 char           *name,
// 			 uint    size)
//  {
//    return font.get_glyph_name (glyph, name, size);
//  }

/**
 * hb_font_get_glyph_from_name:
 * @font: #hb_font_t to work upon
 * @name: (array length=len): The name string to query
 * @len: The length of the name queried
 * @glyph: (out): The glyph ID retrieved
 *
 * Fetches the glyph ID that corresponds to a name string in the specified @font.
 *
 * <note>Note: @len == -1 means the name string is null-terminated.</note>
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
//  hb_bool_t
//  hb_font_get_glyph_from_name (font *hb_font_t      ,
// 				  const char     *name,
// 				  int             len, /* -1 means nul-terminated */
// 				  rune *glyph)
//  {
//    return font.get_glyph_from_name (name, len, glyph);
//  }

/* A bit higher-level, and with fallback */

/**
 * hb_font_get_extents_for_direction:
 * @font: #hb_font_t to work upon
 * @direction: The direction of the text segment
 * @extents: (out): The #hb_font_extents_t retrieved
 *
 * Fetches the extents for a font in a text segment of the
 * specified direction.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Since: 1.1.3
 **/
func hb_font_get_extents_for_direction(font *hb_font_t,
	hb_direction_t direction,
	extents *hb_font_extents_t) {
	return font.get_extents_for_direction(direction, extents)
}

/**
 * hb_font_get_glyph_advance_for_direction:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @direction: The direction of the text segment
 * @x: (out): The horizontal advance retrieved
 * @y: (out):  The vertical advance retrieved
 *
 * Fetches the advance for a glyph ID from the specified font,
 * in a text segment of the specified direction.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Since: 0.9.2
 **/
func hb_font_get_glyph_advance_for_direction(font *hb_font_t,
	rune glyph,
	hb_direction_t direction,
	hb_position_t *x,
	hb_position_t *y) {
	return font.get_glyph_advance_for_direction(glyph, direction, x, y)
}

/**
 * hb_font_get_glyph_advances_for_direction:
 * @font: #hb_font_t to work upon
 * @direction: The direction of the text segment
 * @count: The number of glyph IDs in the sequence queried
 * @first_glyph: The first glyph ID to query
 * @glyph_stride: The stride between successive glyph IDs
 * @first_advance: (out): The first advance retrieved
 * @advance_stride: (out): The stride between successive advances
 *
 * Fetches the advances for a sequence of glyph IDs in the specified
 * font, in a text segment of the specified direction.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Since: 1.8.6
 **/
func hb_font_get_glyph_advances_for_direction(font *hb_font_t,
	hb_direction_t direction,
	uint count,
	first_glyph []rune,
	unsigned glyph_stride,
	first_advance []hb_position_t,
	unsigned advance_stride) {
	font.get_glyph_advances_for_direction(direction, count, first_glyph, glyph_stride, first_advance, advance_stride)
}

/**
 * hb_font_get_glyph_origin_for_direction:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @direction: The direction of the text segment
 * @x: (out): The X coordinate retrieved for the origin
 * @y: (out): The Y coordinate retrieved for the origin
 *
 * Fetches the (X,Y) coordinates of the origin for a glyph in
 * the specified font.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Since: 0.9.2
 **/
func hb_font_get_glyph_origin_for_direction(font *hb_font_t,
	rune glyph,
	hb_direction_t direction,
	hb_position_t *x,
	hb_position_t *y) {
	return font.get_glyph_origin_for_direction(glyph, direction, x, y)
}

/**
 * hb_font_add_glyph_origin_for_direction:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @direction: The direction of the text segment
 * @x: (inout): Input = The original X coordinate
 *     Output = The X coordinate plus the X-coordinate of the origin
 * @y: (inout): Input = The original Y coordinate
 *     Output = The Y coordinate plus the Y-coordinate of the origin
 *
 * Adds the origin coordinates to an (X,Y) point coordinate, in
 * the specified glyph ID in the specified font.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Since: 0.9.2
 **/
func hb_font_add_glyph_origin_for_direction(font *hb_font_t,
	rune glyph,
	hb_direction_t direction,
	hb_position_t *x,
	hb_position_t *y) {
	return font.add_glyph_origin_for_direction(glyph, direction, x, y)
}

/**
 * hb_font_subtract_glyph_origin_for_direction:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @direction: The direction of the text segment
 * @x: (inout): Input = The original X coordinate
 *     Output = The X coordinate minus the X-coordinate of the origin
 * @y: (inout): Input = The original Y coordinate
 *     Output = The Y coordinate minus the Y-coordinate of the origin
 *
 * Subtracts the origin coordinates from an (X,Y) point coordinate,
 * in the specified glyph ID in the specified font.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Since: 0.9.2
 **/
func hb_font_subtract_glyph_origin_for_direction(font *hb_font_t,
	rune glyph,
	hb_direction_t direction,
	hb_position_t *x,
	hb_position_t *y) {
	return font.subtract_glyph_origin_for_direction(glyph, direction, x, y)
}

/**
 * hb_font_get_glyph_kerning_for_direction:
 * @font: #hb_font_t to work upon
 * @first_glyph: The glyph ID of the first glyph in the glyph pair to query
 * @second_glyph: The glyph ID of the second glyph in the glyph pair to query
 * @direction: The direction of the text segment
 * @x: (out): The horizontal kerning-adjustment value retrieved
 * @y: (out): The vertical kerning-adjustment value retrieved
 *
 * Fetches the kerning-adjustment value for a glyph-pair in the specified font.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Since: 0.9.2
 **/
func hb_font_get_glyph_kerning_for_direction(font *hb_font_t,
	rune first_glyph,
	rune second_glyph,
	hb_direction_t direction,
	hb_position_t *x,
	hb_position_t *y) {
	return font.get_glyph_kerning_for_direction(first_glyph, second_glyph, direction, x, y)
}

/**
 * hb_font_get_glyph_extents_for_origin:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @direction: The direction of the text segment
 * @extents: (out): The #hb_glyph_extents_t retrieved
 *
 * Fetches the #hb_glyph_extents_t data for a glyph ID
 * in the specified font, with respect to the origin in
 * a text segment in the specified direction.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
func hb_font_get_glyph_extents_for_origin(font *hb_font_t,
	rune glyph,
	hb_direction_t direction,
	hb_glyph_extents_t *extents) bool {
	return font.get_glyph_extents_for_origin(glyph, direction, extents)
}

/**
 * hb_font_get_glyph_contour_point_for_origin:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @point_index: The contour-point index to query
 * @direction: The direction of the text segment
 * @x: (out): The X value retrieved for the contour point
 * @y: (out): The Y value retrieved for the contour point
 *
 * Fetches the (X,Y) coordinates of a specified contour-point index
 * in the specified glyph ID in the specified font, with respect
 * to the origin in a text segment in the specified direction.
 *
 * Calls the appropriate direction-specific variant (horizontal
 * or vertical) depending on the value of @direction.
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
func hb_font_get_glyph_contour_point_for_origin(font *hb_font_t,
	rune glyph,
	uint point_index,
	hb_direction_t direction,
	hb_position_t *x,
	hb_position_t *y) bool {
	return font.get_glyph_contour_point_for_origin(glyph, point_index, direction, x, y)
}

/**
 * hb_font_glyph_to_string:
 * @font: #hb_font_t to work upon
 * @glyph: The glyph ID to query
 * @s: (out) (array length=size): The string containing the glyph name
 * @size: Length of string @s
 *
 * Fetches the name of the specified glyph ID in @font and returns
 * it in string @s.
 *
 * If the glyph ID has no name in @font, a string of the form `gidDDD` is
 * generated, with `DDD` being the glyph ID.
 *
 * Since: 0.9.2
 **/
func hb_font_glyph_to_string(font *hb_font_t,
	rune glyph,
	char *s,
	uint size) {
	font.glyph_to_string(glyph, s, size)
}

/**
 * hb_font_glyph_from_string:
 * @font: #hb_font_t to work upon
 * @s: (array length=len) (element-type uint8_t): string to query
 * @len: The length of the string @s
 * @glyph: (out): The glyph ID corresponding to the string requested
 *
 * Fetches the glyph ID from @font that matches the specified string.
 * Strings of the format `gidDDD` or `uniUUUU` are parsed automatically.
 *
 * <note>Note: @len == -1 means the string is null-terminated.</note>
 *
 * Return value: `true` if data found, `false` otherwise
 *
 * Since: 0.9.2
 **/
func hb_font_glyph_from_string(font *hb_font_t,
	s string, int len, rune *glyph) bool {
	return font.glyph_from_string(s, len, glyph)
}

//  static void
//  _hb_font_adopt_var_coords (hb_font_t *font,
// 				int *coords, /* 2.14 normalized */
// 				float32 *design_coords,
// 				uint coords_length)
//  {
//    free (font.coords);
//    free (font.design_coords);

//    font.coords = coords;
//    font.design_coords = design_coords;
//    font.num_coords = coords_length;
//  }

/**
 * hb_font_get_empty:
 *
 * Fetches the empty font object.
 *
 * Return value: (transfer full): The empty font object
 *
 * Since: 0.9.2
 **/
//  hb_font_t *
//  hb_font_get_empty ()
//  {
//    return const_cast<hb_font_t *> (&Null (hb_font_t));
//  }

/**
 * hb_font_set_parent:
 * @font: #hb_font_t to work upon
 * @parent: The parent font object to assign
 *
 * Sets the parent font of @font.
 *
 * Since: 1.0.5
 **/
//  void
//  hb_font_set_parent (hb_font_t *font,
// 			 hb_font_t *parent)
//  {
//    if (hb_object_is_immutable (font))
// 	 return;

//    if (!parent)
// 	 parent = hb_font_get_empty ();

//    hb_font_t *old = font.parent;

//    font.parent = hb_font_reference (parent);

//    hb_font_destroy (old);
//  }

/**
 * hb_font_get_parent:
 * @font: #hb_font_t to work upon
 *
 * Fetches the parent font of @font.
 *
 * Return value: (transfer none): The parent font object
 *
 * Since: 0.9.2
 **/
//  hb_font_t *
//  hb_font_get_parent (hb_font_t *font)
//  {
//    return font.parent;
//  }

/**
 * hb_font_set_face:
 * @font: #hb_font_t to work upon
 * @face: The #hb_face_t to assign
 *
 * Sets @face as the font-face value of @font.
 *
 * Since: 1.4.3
 **/
//  void
//  hb_font_set_face (hb_font_t *font,
// 		   face *hb_face_t)  {
//    if (hb_object_is_immutable (font))
// 	 return;

//    if (unlikely (!face))
// 	 face = hb_face_get_empty ();

//    hb_face_t *old = font.face;

//    hb_face_make_immutable (face);
//    font.face = hb_face_reference (face);
//    font.mults_changed ();

//    hb_face_destroy (old);
//  }

/**
 * hb_font_get_face:
 * @font: #hb_font_t to work upon
 *
 * Fetches the face associated with the specified font object.
 *
 * Return value: (transfer none): The #hb_face_t value
 *
 * Since: 0.9.2
 **/
//  hb_face_t *
//  hb_font_get_face (hb_font_t *font)
//  {
//    return font.face;
//  }

/**
 * hb_font_set_scale:
 * @font: #hb_font_t to work upon
 * @x_scale: Horizontal scale value to assign
 * @y_scale: Vertical scale value to assign
 *
 * Sets the horizontal and vertical scale of a font.
 *
 * Since: 0.9.2
 **/
func hb_font_set_scale(hb_font_t *font,
	int x_scale,
	int y_scale) {
	if hb_object_is_immutable(font) {
		return
	}

	font.x_scale = x_scale
	font.y_scale = y_scale
	font.mults_changed()
}

/**
 * hb_font_get_scale:
 * @font: #hb_font_t to work upon
 * @x_scale: (out): Horizontal scale value
 * @y_scale: (out): Vertical scale value
 *
 * Fetches the horizontal and vertical scale of a font.
 *
 * Since: 0.9.2
 **/
//  void
//  hb_font_get_scale (hb_font_t *font,
// 			int       *x_scale,
// 			int       *y_scale)
//  {
//    if (x_scale) *x_scale = font.x_scale;
//    if (y_scale) *y_scale = font.y_scale;
//  }

/**
 * hb_font_set_ppem:
 * @font: #hb_font_t to work upon
 * @x_ppem: Horizontal ppem value to assign
 * @y_ppem: Vertical ppem value to assign
 *
 * Sets the horizontal and vertical pixels-per-em (ppem) of a font.
 *
 * Since: 0.9.2
 **/
//  void
//  hb_font_set_ppem (font *hb_font_t,
// 		   uint  x_ppem,
// 		   uint  y_ppem)
//  {
//    if (hb_object_is_immutable (font))
// 	 return;

//    font.x_ppem = x_ppem;
//    font.y_ppem = y_ppem;
//  }

/**
 * hb_font_get_ppem:
 * @font: #hb_font_t to work upon
 * @x_ppem: (out): Horizontal ppem value
 * @y_ppem: (out): Vertical ppem value
 *
 * Fetches the horizontal and vertical points-per-em (ppem) of a font.
 *
 * Since: 0.9.2
 **/
//  void
//  hb_font_get_ppem (font *hb_font_t,
// 		   uint *x_ppem,
// 		   uint *y_ppem)
//  {
//    if (x_ppem) *x_ppem = font.x_ppem;
//    if (y_ppem) *y_ppem = font.y_ppem;
//  }

/**
 * hb_font_set_ptem:
 * @font: #hb_font_t to work upon
 * @ptem: font size in points.
 *
 * Sets the "point size" of a font. Set to zero to unset.
 * Used in CoreText to implement optical sizing.
 *
 * <note>Note: There are 72 points in an inch.</note>
 *
 * Since: 1.6.0
 **/
//  void
//  hb_font_set_ptem (hb_font_t *font,
// 		   float32      ptem)
//  {
//    if (hb_object_is_immutable (font))
// 	 return;

//    font.ptem = ptem;
//  }

/**
 * hb_font_get_ptem:
 * @font: #hb_font_t to work upon
 *
 * Fetches the "point size" of a font. Used in CoreText to
 * implement optical sizing.
 *
 * Return value: Point size.  A value of zero means "not set."
 *
 * Since: 0.9.2
 **/
//  float32
//  hb_font_get_ptem (hb_font_t *font)
//  {
//    return font.ptem;
//  }

/*
 * Variations
 */

/**
 * hb_font_set_variations:
 * @font: #hb_font_t to work upon
 * @variations: (array length=variations_length): Array of variation settings to apply
 * @variations_length: Number of variations to apply
 *
 * Applies a list of font-variation settings to a font.
 *
 * Since: 1.4.2
 */
// func hb_font_set_variations(font *hb_font_t, variations []hb_variation_t) {
// 	if hb_object_is_immutable(font) {
// 		return
// 	}

// 	if !variations_length {
// 		hb_font_set_var_coords_normalized(font, nullptr, 0)
// 		return
// 	}

// 	coords_length := hb_ot_var_get_axis_count(font.face)

// 	//    int *normalized = coords_length ? (int *) calloc (coords_length, sizeof (int)) : nullptr;
// 	//    float32 *design_coords = coords_length ? (float32 *) calloc (coords_length, sizeof (float32)) : nullptr;

// 	if unlikely(coords_length && !(normalized && design_coords)) {
// 		return
// 	}

// 	fvar := *font.face.table.fvar
// 	for i = 0; i < variations_length; i++ {
// 		//  hb_ot_var_axis_info_t info;
// 		if hb_ot_var_find_axis_info(font.face, variations[i].tag, &info) &&
// 			info.axis_index < coords_length {
// 			//    float32 v = variations[i].value;
// 			design_coords[info.axis_index] = v
// 			normalized[info.axis_index] = fvar.normalize_axis_value(info.axis_index, v)
// 		}
// 	}
// 	font.face.table.avar.map_coords(normalized, coords_length)

// 	_hb_font_adopt_var_coords(font, normalized, design_coords, coords_length)
// }

/**
 * hb_font_set_var_coords_design:
 * @font: #hb_font_t to work upon
 * @coords: (array length=coords_length): Array of variation coordinates to apply
 * @coords_length: Number of coordinates to apply
 *
 * Applies a list of variation coordinates (in design-space units)
 * to a font.
 *
 * Since: 1.4.2
 */
func (font *hb_font_t) hb_font_set_var_coords_design(coords []float32) {
	font.coords = font.face.Normalize(coords)
	font.design_coords = append([]float32(nil), coords...)
}

/**
 * hb_font_set_var_named_instance:
 * @font: a font.
 * @instance_index: named instance index.
 *
 * Sets design coords of a font from a named instance index.
 *
 * Since: 2.6.0
 */
// func hb_font_set_var_named_instance(hb_font_t *font, instance_index int) {
// 	if hb_object_is_immutable(font) {
// 		return
// 	}

// 	coords_length := hb_ot_var_named_instance_get_design_coords(font.face, instance_index, nullptr, nullptr)

// 	//    float32 *coords = coords_length ? (float32 *) calloc (coords_length, sizeof (float32)) : nullptr;
// 	if unlikely(coords_length && !coords) {
// 		return
// 	}

// 	hb_ot_var_named_instance_get_design_coords(font.face, instance_index, &coords_length, coords)
// 	hb_font_set_var_coords_design(font, coords, coords_length)
// 	free(coords)
// }

/**
 * hb_font_set_var_coords_normalized:
 * @font: #hb_font_t to work upon
 * @coords: (array length=coords_length): Array of variation coordinates to apply
 * @coords_length: Number of coordinates to apply
 *
 * Applies a list of variation coordinates (in normalized units)
 * to a font.
 *
 * <note>Note: Coordinates should be normalized to 2.14.</note>
 *
 * Since: 1.4.2
 */
// func hb_font_set_var_coords_normalized(font *hb_font_t,
// 	coords []uint16 /* 2.14 normalized */) {
// 	if hb_object_is_immutable(font) {
// 		return
// 	}

// 	//    int *copy = coords_length ? (int *) calloc (coords_length, sizeof (coords[0])) : nullptr;
// 	//    int *unmapped = coords_length ? (int *) calloc (coords_length, sizeof (coords[0])) : nullptr;
// 	//    float32 *design_coords = coords_length ? (float32 *) calloc (coords_length, sizeof (design_coords[0])) : nullptr;

// 	if unlikely(coords_length && !(copy && unmapped && design_coords)) {
// 		free(copy)
// 		free(unmapped)
// 		free(design_coords)
// 		return
// 	}

// 	if coords_length {
// 		memcpy(copy, coords, coords_length*sizeof(coords[0]))
// 		memcpy(unmapped, coords, coords_length*sizeof(coords[0]))
// 	}

// 	/* Best effort design coords simulation */
// 	font.face.table.avar.unmap_coords(unmapped, coords_length)
// 	for i = 0; i < coords_length; i++ {
// 		design_coords[i] = font.face.table.fvar.unnormalize_axis_value(i, unmapped[i])
// 	}
// 	free(unmapped)

// 	_hb_font_adopt_var_coords(font, copy, design_coords, coords_length)
// }

/**
 * hb_font_get_var_coords_normalized:
 * @font: #hb_font_t to work upon
 * @length: Number of coordinates retrieved
 *
 * Fetches the list of normalized variation coordinates currently
 * set on a font.
 *
 * Return value is valid as long as variation coordinates of the font
 * are not modified.
 *
 * Since: 1.4.2
 */
//  const int *
//  hb_font_get_var_coords_normalized (font *hb_font_t,
// 					uint *length)
//  {
//    if (length)
// 	 *length = font.num_coords;

//    return font.coords;
//  }

//  #ifdef HB_EXPERIMENTAL_API
//  /**
//   * hb_font_get_var_coords_design:
//   * @font: #hb_font_t to work upon
//   * @length: (out): number of coordinates
//   *
//   * Return value is valid as long as variation coordinates of the font
//   * are not modified.
//   *
//   * Return value: coordinates array
//   *
//   * Since: EXPERIMENTAL
//   */
//  const float32 *
//  hb_font_get_var_coords_design (hb_font_t *font,
// 					uint *length)
//  {
//    if (length)
// 	 *length = font.num_coords;

//    return font.design_coords;
//  }
//  #endif
