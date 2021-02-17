package harfbuzz

import (
	"math"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from src/hb-font.hh, src/hb-font.cc  Copyright © 2009  Red Hat, Inc., 2012  Google, Inc.  Behdad Esfahbod

const faceUpem = 1000

var emptyFont = Font{
	//    const_cast<Face *> (&_hb_Null_hb_face_t), // TODO: "empty face"
	XScale: 1000,    // x_scale
	YScale: 1000,    // y_scale
	x_mult: 1 << 16, // x_mult
	y_mult: 1 << 16, // y_mult
}

// hb_font_extents_t exposes font-wide extent values, measured in font units.
// Note that typically ascender is positive and descender negative in coordinate systems that grow up.
// TODO: use plain ints if possible
type hb_font_extents_t struct {
	Ascender  Position // typographic ascender.
	Descender Position // typographic descender.
	LineGap   Position // suggested line spacing gap.
}

// Glyph extent values, measured in font units.
// Note that height is negative in coordinate systems that grow up.
type GlyphExtents struct {
	XBearing Position // left side of glyph from origin
	YBearing Position // top side of glyph from origin
	Width    Position // distance from left to right side
	Height   Position // distance from top to bottom side
}

type Face interface {
	// common

	// // Returns the number of glyphs found in the font.
	// GetNumGlyphs() int

	// Returns the extents of the font for horizontal text, or false
	// it not available.
	GetFontHExtents() (hb_font_extents_t, bool)

	// Return the glyph used to represent the given rune,
	// or false if not found.
	GetNominalGlyph(ch rune) (fonts.GlyphIndex, bool)

	// Retrieves the glyph ID for a specified Unicode code point
	// followed by a specified Variation Selector code point, or false if not found
	GetVariationGlyph(ch, varSelector rune) (fonts.GlyphIndex, bool)

	// Returns the horizontal advance, or false if no
	// advance is found an a defaut value should be used.
	// `coords` is used by variable fonts, and specified in normalized coordinates.
	GetHorizontalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool)

	// Same as `GetHorizontalAdvance`, but for vertical advance
	GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool)

	// Fetches the (X,Y) coordinates of the origin (in font units) for a glyph ID,
	// for horizontal text segments.
	// Returns `false` if not available.
	GetGlyphHOrigin(fonts.GlyphIndex) (x, y Position, found bool)

	// Same as `GetGlyphHOrigin`, but for vertical text segments.
	GetGlyphVOrigin(fonts.GlyphIndex) (x, y Position, found bool)

	// Retrieve the extents for a specified glyph, of false, if not available.
	GetGlyphExtents(fonts.GlyphIndex) (GlyphExtents, bool)

	// specialized

	// Retrieve the (X,Y) coordinates (in font units) for a
	// specified contour point in a glyph, or false if not found.
	GetGlyphContourPoint(glyph fonts.GlyphIndex, pointIndex uint16) (x, y Position, ok bool)
	get_gsubgpos_table() (gsub *truetype.TableGSUB, gpos *truetype.TableGPOS) // optional
	getGDEF() truetype.TableGDEF                                              // optional
	getMorxTable() truetype.TableMorx
	getKerxTable() truetype.TableKernx
	getKernTable() truetype.TableKernx
	getAnkrTable() truetype.TableAnkr
	getTrakTable() truetype.TableTrak
	getFeatTable() truetype.TableFeat
	// return the variations_index
	hb_ot_layout_table_find_feature_variations(table_tag hb_tag_t, coords []float32) int
	Normalize(coords []float32) []float32
}

// Font represents a font face at a specific size and with
// certain other parameters (pixels-per-em, points-per-em, variation
// settings) specified. Font objects are created from font face
// objects, and are used as input to Shape(), among other things.
//
// Client programs can optionally pass in their own functions that
// implement the basic, lower-level queries of font objects. This set
// of font functions is defined by the virtual methods in
// #hb_font_funcs_t.
//
// HarfBuzz provides a built-in set of lightweight default
// functions for each method in #hb_font_funcs_t.
type Font struct {
	Face Face

	XScale, YScale int32
	x_mult, y_mult int64 // cached value of  (x_scale << 16) / faceUpem

	x_ppem, y_ppem uint16

	ptem float32

	// font variation coordinates (optionnal)
	coords        []float32 // length num_coords, normalized
	design_coords []float32 // length num_coords, in design units

	// accelators for lookup // TODO: à initialiszer dans le constructor
	gsubAccels, gsposAccels []hb_ot_layout_lookup_accelerator_t
}

/* Convert from font-space to user-space */
//    int64 dir_mult (Direction direction) { return HB_DIRECTION_IS_VERTICAL(direction) ? y_mult : x_mult; }
func (f Font) em_scale_x(v int16) Position    { return em_mult(v, f.x_mult) }
func (f Font) em_scale_y(v int16) Position    { return em_mult(v, f.y_mult) }
func (f Font) em_scalef_x(v float32) Position { return em_scalef(v, f.XScale) }
func (f Font) em_scalef_y(v float32) Position { return em_scalef(v, f.YScale) }
func (f Font) em_fscale_x(v int16) float32    { return em_fscale(v, f.XScale) }
func (f Font) em_fscale_y(v int16) float32    { return em_fscale(v, f.YScale) }

func (f *Font) mults_changed() {
	f.x_mult = (int64(f.XScale) << 16) / faceUpem
	f.y_mult = (int64(f.YScale) << 16) / faceUpem
}

func em_mult(v int16, mult int64) Position {
	return Position((int64(v) * mult) >> 16)
}

func em_scalef(v float32, scale int32) Position {
	return Position(math.Round(float64(v * float32(scale) / faceUpem)))
}

func em_fscale(v int16, scale int32) float32 {
	return float32(v) * float32(scale) / faceUpem
}

// Fetches the advance for a glyph ID from the specified font,
// in a text segment of the specified direction.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f Font) GetGlyphAdvanceForDirection(glyph fonts.GlyphIndex, dir Direction) (x, y Position) {
	if dir.IsHorizontal() {
		return f.GetGlyphHAdvance(glyph), 0
	}
	return 0, f.GetGlyphVAdvance(glyph)
}

// Fetches the advance for a glyph ID in the specified font,
// for horizontal text segments.
func (f *Font) GetGlyphHAdvance(glyph fonts.GlyphIndex) Position {
	adv, has := f.Face.GetHorizontalAdvance(glyph)
	if !has {
		adv = faceUpem
	}
	return f.em_scale_x(adv)
}

// Fetches the advance for a glyph ID in the specified font,
// for vertical text segments.
func (f *Font) GetGlyphVAdvance(glyph fonts.GlyphIndex) Position {
	adv, has := f.Face.GetVerticalAdvance(glyph)
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
func (f *Font) subtractGlyphOriginForDirection(glyph fonts.GlyphIndex, direction Direction,
	x, y Position) (Position, Position) {
	origin_x, origin_y := f.get_glyph_origin_for_direction(glyph, direction)

	return x - origin_x, y - origin_y
}

// Fetches the (X,Y) coordinates of the origin for a glyph in
// the specified font.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f *Font) get_glyph_origin_for_direction(glyph fonts.GlyphIndex, direction Direction) (x, y Position) {
	if direction.IsHorizontal() {
		return f.get_glyph_h_origin_with_fallback(glyph)
	}
	return f.get_glyph_v_origin_with_fallback(glyph)
}

func (f *Font) get_glyph_h_origin_with_fallback(glyph fonts.GlyphIndex) (Position, Position) {
	x, y, ok := f.Face.GetGlyphHOrigin(glyph)
	if !ok {
		x, y, ok = f.Face.GetGlyphVOrigin(glyph)
		if ok {
			dx, dy := f.guess_v_origin_minus_h_origin(glyph)
			return x - dx, y - dy
		}
	}
	return x, y
}

func (f *Font) get_glyph_v_origin_with_fallback(glyph fonts.GlyphIndex) (Position, Position) {
	x, y, ok := f.Face.GetGlyphVOrigin(glyph)
	if !ok {
		x, y, ok = f.Face.GetGlyphHOrigin(glyph)
		if ok {
			dx, dy := f.guess_v_origin_minus_h_origin(glyph)
			return x + dx, y + dy
		}
	}
	return x, y
}

func (f *Font) guess_v_origin_minus_h_origin(glyph fonts.GlyphIndex) (x, y Position) {
	x = f.GetGlyphHAdvance(glyph) / 2
	extents := f.get_h_extents_with_fallback()
	y = extents.Ascender
	return x, y
}

func (f *Font) get_h_extents_with_fallback() hb_font_extents_t {
	extents, ok := f.Face.GetFontHExtents()
	if !ok {
		extents.Ascender = f.YScale * 4 / 5
		extents.Descender = extents.Ascender - f.YScale
		extents.LineGap = 0
	}
	return extents
}

func (f *Font) HasGlyph(ch rune) bool {
	_, ok := f.Face.GetNominalGlyph(ch)
	return ok
}

func (f *Font) subtract_glyph_h_origin(glyph fonts.GlyphIndex, x, y Position) (Position, Position) {
	origin_x, origin_y := f.get_glyph_h_origin_with_fallback(glyph)
	return x - origin_x, y - origin_y
}

func (f *Font) subtract_glyph_v_origin(glyph fonts.GlyphIndex, x, y Position) (Position, Position) {
	origin_x, origin_y := f.get_glyph_v_origin_with_fallback(glyph)
	return x - origin_x, y - origin_y
}

func (f *Font) add_glyph_h_origin(glyph fonts.GlyphIndex, x, y Position) (Position, Position) {
	origin_x, origin_y := f.get_glyph_h_origin_with_fallback(glyph)
	return x + origin_x, y + origin_y
}

func (f *Font) get_glyph_contour_point_for_origin(glyph fonts.GlyphIndex, pointIndex uint16, direction Direction) (x, y Position, ok bool) {
	x, y, ok = f.Face.GetGlyphContourPoint(glyph, pointIndex)

	if ok {
		x, y = f.subtractGlyphOriginForDirection(glyph, direction, x, y)
	}

	return x, y, ok
}

//    Position em_scale_dir (int16 v, Direction direction)
//    { return em_mult (v, dir_mult (direction)); }

//    /* Convert from parent-font user-space to our user-space */
//    Position parent_scale_x_distance (Position v)
//    {
// 	 if (unlikely (parent && parent.x_scale != x_scale))
// 	   return (Position) (v * (int64) this.x_scale / this.parent.x_scale);
// 	 return v;
//    }
//    Position parent_scale_y_distance (Position v)
//    {
// 	 if (unlikely (parent && parent.y_scale != y_scale))
// 	   return (Position) (v * (int64) this.y_scale / this.parent.y_scale);
// 	 return v;
//    }
//    Position parent_scale_x_position (Position v)
//    { return parent_scale_x_distance (v); }
//    Position parent_scale_y_position (Position v)
//    { return parent_scale_y_distance (v); }

//    void parent_scale_distance (Position *x, Position *y)
//    {
// 	 *x = parent_scale_x_distance (*x);
// 	 *y = parent_scale_y_distance (*y);
//    }
//    void parent_scale_position (Position *x, Position *y)
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

//    bool HasGlyph (hb_codepoint_t unicode)
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

//    Position GetGlyphHAdvance (hb_codepoint_t glyph)
//    {
// 	 return klass.get.f.glyph_h_advance (this, user_data,
// 					  glyph,
// 					  klass.user_data.glyph_h_advance);
//    }

//    Position GetGlyphVAdvance (hb_codepoint_t glyph)
//    {
// 	 return klass.get.f.glyph_v_advance (this, user_data,
// 					  glyph,
// 					  klass.user_data.glyph_v_advance);
//    }

//    void get_glyph_h_advances (unsigned int count,
// 				  const hb_codepoint_t *first_glyph,
// 				  unsigned int glyph_stride,
// 				  Position *first_advance,
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
// 				  Position *first_advance,
// 				  unsigned int advance_stride)
//    {
// 	 return klass.get.f.glyph_v_advances (this, user_data,
// 					   count,
// 					   first_glyph, glyph_stride,
// 					   first_advance, advance_stride,
// 					   klass.user_data.glyph_v_advances);
//    }

//    hb_bool_t get_glyph_h_origin (hb_codepoint_t glyph,
// 				 Position *x, Position *y)
//    {
// 	 *x = *y = 0;
// 	 return klass.get.f.glyph_h_origin (this, user_data,
// 					 glyph, x, y,
// 					 klass.user_data.glyph_h_origin);
//    }

//    hb_bool_t get_glyph_v_origin (hb_codepoint_t glyph,
// 				 Position *x, Position *y)
//    {
// 	 *x = *y = 0;
// 	 return klass.get.f.glyph_v_origin (this, user_data,
// 					 glyph, x, y,
// 					 klass.user_data.glyph_v_origin);
//    }

//    Position get_glyph_h_kerning (hb_codepoint_t left_glyph,
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

//    Position get_glyph_v_kerning (hb_codepoint_t top_glyph,
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
// 					GlyphExtents *extents)
//    {
// 	 memset (extents, 0, sizeof (*extents));
// 	 return klass.get.f.glyph_extents (this, user_data,
// 						glyph,
// 						extents,
// 						klass.user_data.glyph_extents);
//    }

//    hb_bool_t get_glyph_contour_point (hb_codepoint_t glyph, unsigned int point_index,
// 					  Position *x, Position *y)
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

//    void get_extents_for_direction (Direction direction,
// 				   hb_font_extents_t *extents)
//    {
// 	 if (likely (HB_DIRECTION_IS_HORIZONTAL (direction)))
// 	   get_h_extents_with_fallback (extents);
// 	 else
// 	   get_v_extents_with_fallback (extents);
//    }

//    void get_glyph_advances_for_direction (Direction direction,
// 					  unsigned int count,
// 					  const hb_codepoint_t *first_glyph,
// 					  unsigned glyph_stride,
// 					  Position *first_advance,
// 					  unsigned advance_stride)
//    {
// 	 if (likely (HB_DIRECTION_IS_HORIZONTAL (direction)))
// 	   get_glyph_h_advances (count, first_glyph, glyph_stride, first_advance, advance_stride);
// 	 else
// 	   get_glyph_v_advances (count, first_glyph, glyph_stride, first_advance, advance_stride);
//    }

//    void add_glyph_v_origin (hb_codepoint_t glyph,
// 				Position *x, Position *y)
//    {
// 	 Position origin_x, origin_y;

// 	 get_glyph_v_origin_with_fallback (glyph, &origin_x, &origin_y);

// 	 *x += origin_x;
// 	 *y += origin_y;
//    }
//    void add_glyph_origin_for_direction (hb_codepoint_t glyph,
// 						Direction direction,
// 						Position *x, Position *y)
//    {
// 	 Position origin_x, origin_y;

// 	 get_glyph_origin_for_direction (glyph, direction, &origin_x, &origin_y);

// 	 *x += origin_x;
// 	 *y += origin_y;
//    }

//    void get_glyph_kerning_for_direction (hb_codepoint_t first_glyph, hb_codepoint_t second_glyph,
// 					 Direction direction,
// 					 Position *x, Position *y)
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
// 					   Direction direction,
// 					   GlyphExtents *extents)
//    {
// 	 hb_bool_t ret = get_glyph_extents (glyph, extents);

// 	 if (ret)
// 	   subtractGlyphOriginForDirection (glyph, direction, &extents.x_bearing, &extents.y_bearing);

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
func hb_font_create(face Face) *Font {
	var font Font

	font.Face = face
	font.XScale = faceUpem
	font.YScale = faceUpem
	font.x_mult = 1 << 16
	font.y_mult = 1 << 16

	return &font
}

/* A bit higher-level, and with fallback */

// /**
//  * hb_font_set_scale:
//  * @font: #Font to work upon
//  * @x_scale: Horizontal scale value to assign
//  * @y_scale: Vertical scale value to assign
//  *
//  * Sets the horizontal and vertical scale of a font.
//  *
//  * Since: 0.9.2
//  **/
// func hb_font_set_scale(Font *font,
// 	int x_scale,
// 	int y_scale) {
// 	if hb_object_is_immutable(font) {
// 		return
// 	}

// 	font.x_scale = x_scale
// 	font.y_scale = y_scale
// 	font.mults_changed()
// }

/**
 * hb_font_set_ppem:
 * @font: #Font to work upon
 * @x_ppem: Horizontal ppem value to assign
 * @y_ppem: Vertical ppem value to assign
 *
 * Sets the horizontal and vertical pixels-per-em (ppem) of a font.
 *
 * Since: 0.9.2
 **/
//  void
//  hb_font_set_ppem (font *Font,
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
 * @font: #Font to work upon
 * @x_ppem: (out): Horizontal ppem value
 * @y_ppem: (out): Vertical ppem value
 *
 * Fetches the horizontal and vertical points-per-em (ppem) of a font.
 *
 * Since: 0.9.2
 **/
//  void
//  hb_font_get_ppem (font *Font,
// 		   uint *x_ppem,
// 		   uint *y_ppem)
//  {
//    if (x_ppem) *x_ppem = font.x_ppem;
//    if (y_ppem) *y_ppem = font.y_ppem;
//  }

/**
 * hb_font_set_ptem:
 * @font: #Font to work upon
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
//  hb_font_set_ptem (Font *font,
// 		   float32      ptem)
//  {
//    if (hb_object_is_immutable (font))
// 	 return;

//    font.ptem = ptem;
//  }

/**
 * hb_font_get_ptem:
 * @font: #Font to work upon
 *
 * Fetches the "point size" of a font. Used in CoreText to
 * implement optical sizing.
 *
 * Return value: Point size.  A value of zero means "not set."
 *
 * Since: 0.9.2
 **/
//  float32
//  hb_font_get_ptem (Font *font)
//  {
//    return font.ptem;
//  }

/*
 * Variations
 */

/**
 * hb_font_set_variations:
 * @font: #Font to work upon
 * @variations: (array length=variations_length): Array of variation settings to apply
 * @variations_length: Number of variations to apply
 *
 * Applies a list of font-variation settings to a font.
 *
 * Since: 1.4.2
 */
// func hb_font_set_variations(font *Font, variations []hb_variation_t) {
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
 * @font: #Font to work upon
 * @coords: (array length=coords_length): Array of variation coordinates to apply
 * @coords_length: Number of coordinates to apply
 *
 * Applies a list of variation coordinates (in design-space units)
 * to a font.
 *
 * Since: 1.4.2
 */
func (font *Font) hb_font_set_var_coords_design(coords []float32) {
	font.coords = font.Face.Normalize(coords)
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
// func hb_font_set_var_named_instance(Font *font, instance_index int) {
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
 * @font: #Font to work upon
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
// func hb_font_set_var_coords_normalized(font *Font,
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
 * @font: #Font to work upon
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
//  hb_font_get_var_coords_normalized (font *Font,
// 					uint *length)
//  {
//    if (length)
// 	 *length = font.num_coords;

//    return font.coords;
//  }

//  #ifdef HB_EXPERIMENTAL_API
//  /**
//   * hb_font_get_var_coords_design:
//   * @font: #Font to work upon
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
//  hb_font_get_var_coords_design (Font *font,
// 					uint *length)
//  {
//    if (length)
// 	 *length = font.num_coords;

//    return font.design_coords;
//  }
//  #endif
