package harfbuzz

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from src/hb-font.hh, src/hb-font.cc  Copyright © 2009  Red Hat, Inc., 2012  Google, Inc.  Behdad Esfahbod

// type Face interface {
// 	// Returns the units per em of the font file.
// 	// If not found, should return 1000 as fallback value.
// 	GetUpem() uint16

// 	// Returns the extents of the font for horizontal text, or false
// 	// it not available.
// 	GetFontHExtents() (fonts.FontExtents, bool)

// 	// Return the glyph used to represent the given rune,
// 	// or false if not found.
// 	GetNominalGlyph(ch rune) (fonts.GlyphIndex, bool)

// 	// Retrieves the glyph ID for a specified Unicode code point
// 	// followed by a specified Variation Selector code point, or false if not found
// 	GetVariationGlyph(ch, varSelector rune) (fonts.GlyphIndex, bool)

// 	// Returns the horizontal advance in font units, or false if no
// 	// advance is found an a defaut value should be used.
// 	// `coords` is used by variable fonts, and is specified in normalized coordinates.
// 	GetHorizontalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool)

// 	// Same as `GetHorizontalAdvance`, but for vertical advance.
// 	GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) (int16, bool)

// 	// Fetches the (X,Y) coordinates of the origin (in font units) for a glyph ID,
// 	// for horizontal text segments.
// 	// Returns `false` if not available.
// 	GetGlyphHOrigin(fonts.GlyphIndex) (x, y Position, found bool)

// 	// Same as `GetGlyphHOrigin`, but for vertical text segments.
// 	GetGlyphVOrigin(fonts.GlyphIndex) (x, y Position, found bool)

// 	// Retrieve the extents for a specified glyph, of false, if not available.
// 	GetGlyphExtents(fonts.GlyphIndex) (fonts.GlyphExtents, bool)

// 	// NormalizeVariations should normalize the given design-space coordinates. The minimum and maximum
// 	// values for the axis are mapped to the interval [-1,1], with the default
// 	// axis value mapped to 0.
// 	// This should be a no-op for non-variable fonts.
// 	NormalizeVariations(coords []float32) []float32
// }

type Face = fonts.FontMetrics

// FaceOpentype add support for adavanced layout features
// found in Opentype/Truetype font files.
type FaceOpentype interface {
	Face

	// LayoutTables fetch the opentype layout tables of the font.
	LayoutTables() truetype.LayoutTables

	// Retrieve the (X,Y) coordinates (in font units) for a
	// specified contour point in a glyph, or false if not found.
	GetGlyphContourPoint(glyph fonts.GlyphIndex, pointIndex uint16) (x, y Position, ok bool)
}

// FaceGraphite add support for Graphite layout tables
type FaceGraphite interface {
	Face

	IsGraphite() // TODO:
}

// Font is used internally as a light wrapper around the provided Face.
//
// While a font face is generally the in-memory representation of a static font file,
// `Font` handles dynamic attributes like size, with and
// other parameters (pixels-per-em, points-per-em, variation
// settings).
//
// Font are constructed with `NewFont` and adjusted by accessing the fields
// XPpem, YPpem, Ptem,XScale, YScale and with the method `SetVarCoordsDesign` for
// variable fonts.
type Font struct {
	face Face

	otTables               *truetype.LayoutTables      // opentype fields, initialized from a FaceOpentype
	coords                 []float32                   // // font variation coordinates (optionnal), normalized
	gsubAccels, gposAccels []otLayoutLookupAccelerator // accelators for lookup
	faceUpem               int32                       // cached value of Face.GetUpem()

	// Point size of the font. Set to zero to unset.
	// This is used in AAT layout, when applying 'trak' table.
	Ptem float32

	// Horizontal and vertical scale of the font.
	// The resulting positions are computed with: fontUnit * Scale / faceUpem,
	// where faceUpem is given by the face
	XScale, YScale int32

	// Horizontal and vertical pixels-per-em (ppem) of the font.
	XPpem, YPpem uint16
}

// NewFont constructs a new font object from the specified face.
// It will cache some internal values and set a default size.
// The `face` object should not be modified after this call.
func NewFont(face Face) *Font {
	var font Font

	font.face = face
	font.faceUpem = Position(face.GetUpem())
	font.XScale = font.faceUpem
	font.YScale = font.faceUpem

	if opentypeFace, ok := face.(FaceOpentype); ok {
		lt := opentypeFace.LayoutTables()
		font.otTables = &lt

		// accelerators
		font.gsubAccels = make([]otLayoutLookupAccelerator, len(lt.GSUB.Lookups))
		for i, l := range lt.GSUB.Lookups {
			font.gsubAccels[i].init(lookupGSUB(l))
		}
		font.gposAccels = make([]otLayoutLookupAccelerator, len(lt.GPOS.Lookups))
		for i, l := range lt.GPOS.Lookups {
			font.gposAccels[i].init(lookupGPOS(l))
		}
	}

	return &font
}

// SetVarCoordsDesign applies a list of variation coordinates, in design-space units,
// to the font.
func (f *Font) SetVarCoordsDesign(coords []float32) {
	f.coords = f.face.NormalizeVariations(coords)
	// f.designCoords = append([]float32(nil), coords...)
}

/* Convert from font-space to user-space */

func (f *Font) emScaleX(v int16) Position    { return Position(v) * f.XScale / f.faceUpem }
func (f *Font) emScaleY(v int16) Position    { return Position(v) * f.YScale / f.faceUpem }
func (f *Font) emScalefX(v float32) Position { return emScalef(v, f.XScale, f.faceUpem) }
func (f *Font) emScalefY(v float32) Position { return emScalef(v, f.YScale, f.faceUpem) }
func (f *Font) emFscaleX(v int16) float32    { return emFscale(v, f.XScale, f.faceUpem) }
func (f *Font) emFscaleY(v int16) float32    { return emFscale(v, f.YScale, f.faceUpem) }

func emScalef(v float32, scale, faceUpem int32) Position {
	return roundf(v * float32(scale) / float32(faceUpem))
}

func emFscale(v int16, scale, faceUpem int32) float32 {
	return float32(v) * float32(scale) / float32(faceUpem)
}

// same as fonts.GlyphExtents but with int type
type glyphExtents struct {
	XBearing int32
	YBearing int32
	Width    int32
	Height   int32
}

func (f *Font) getGlyphExtents(glyph fonts.GlyphIndex) (out glyphExtents, ok bool) {
	ext, ok := f.face.GetGlyphExtents(glyph, f.coords, f.XPpem, f.YPpem)
	if !ok {
		return out, false
	}
	out.XBearing = f.emScalefX(ext.XBearing)
	out.Width = f.emScalefX(ext.Width)
	out.YBearing = f.emScalefY(ext.YBearing)
	out.Height = f.emScalefY(ext.Height)
	return out, true
}

// Fetches the advance for a glyph ID from the specified font,
// in a text segment of the specified direction.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of `dir`.
func (f Font) getGlyphAdvanceForDirection(glyph fonts.GlyphIndex, dir Direction) (x, y Position) {
	if dir.isHorizontal() {
		return f.getGlyphHAdvance(glyph), 0
	}
	return 0, f.getGlyphVAdvance(glyph)
}

// Fetches the advance for a glyph ID in the specified font,
// for horizontal text segments.
func (f *Font) getGlyphHAdvance(glyph fonts.GlyphIndex) Position {
	adv := f.face.GetHorizontalAdvance(glyph, f.coords)
	return f.emScalefX(adv)
}

// Fetches the advance for a glyph ID in the specified font,
// for vertical text segments.
func (f *Font) getGlyphVAdvance(glyph fonts.GlyphIndex) Position {
	adv := f.face.GetVerticalAdvance(glyph, f.coords)
	return f.emScalefY(adv)
}

// Subtracts the origin coordinates from an (X,Y) point coordinate,
// in the specified glyph ID in the specified font.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f *Font) subtractGlyphOriginForDirection(glyph fonts.GlyphIndex, direction Direction,
	x, y Position) (Position, Position) {
	originX, originY := f.getGlyphOriginForDirection(glyph, direction)

	return x - originX, y - originY
}

// Fetches the (X,Y) coordinates of the origin for a glyph in
// the specified font.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f *Font) getGlyphOriginForDirection(glyph fonts.GlyphIndex, direction Direction) (x, y Position) {
	if direction.isHorizontal() {
		return f.getGlyphHOriginWithFallback(glyph)
	}
	return f.getGlyphVOriginWithFallback(glyph)
}

func (f *Font) getGlyphHOriginWithFallback(glyph fonts.GlyphIndex) (Position, Position) {
	x, y, ok := f.face.GetGlyphHOrigin(glyph, f.coords)
	if !ok {
		x, y, ok = f.face.GetGlyphVOrigin(glyph, f.coords)
		if ok {
			dx, dy := f.guessVOriginMinusHOrigin(glyph)
			return x - dx, y - dy
		}
	}
	return x, y
}

func (f *Font) getGlyphVOriginWithFallback(glyph fonts.GlyphIndex) (Position, Position) {
	x, y, ok := f.face.GetGlyphVOrigin(glyph, f.coords)
	if !ok {
		x, y, ok = f.face.GetGlyphHOrigin(glyph, f.coords)
		if ok {
			dx, dy := f.guessVOriginMinusHOrigin(glyph)
			return x + dx, y + dy
		}
	}
	return x, y
}

func (f *Font) guessVOriginMinusHOrigin(glyph fonts.GlyphIndex) (x, y Position) {
	x = f.getGlyphHAdvance(glyph) / 2
	y = f.getHExtendsAscender()
	return x, y
}

func (f *Font) getHExtendsAscender() Position {
	extents, ok := f.face.GetFontHExtents(f.coords)
	if !ok {
		return f.YScale * 4 / 5
	}
	return f.emScalefY(extents.Ascender)
}

func (f *Font) hasGlyph(ch rune) bool {
	_, ok := f.face.GetNominalGlyph(ch)
	return ok
}

func (f *Font) subtractGlyphHOrigin(glyph fonts.GlyphIndex, x, y Position) (Position, Position) {
	originX, originY := f.getGlyphHOriginWithFallback(glyph)
	return x - originX, y - originY
}

func (f *Font) subtractGlyphVOrigin(glyph fonts.GlyphIndex, x, y Position) (Position, Position) {
	originX, originY := f.getGlyphVOriginWithFallback(glyph)
	return x - originX, y - originY
}

func (f *Font) addGlyphHOrigin(glyph fonts.GlyphIndex, x, y Position) (Position, Position) {
	originX, originY := f.getGlyphHOriginWithFallback(glyph)
	return x + originX, y + originY
}

// will crash if face if not FaceOpentype
func (f *Font) getGlyphContourPointForOrigin(glyph fonts.GlyphIndex, pointIndex uint16, direction Direction) (x, y Position, ok bool) {
	x, y, ok = f.face.(FaceOpentype).GetGlyphContourPoint(glyph, pointIndex)

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
// 	 Position originX, originY;

// 	 get_glyph_v_origin_with_fallback (glyph, &originX, &originY);

// 	 *x += originX;
// 	 *y += originY;
//    }
//    void add_glyph_origin_for_direction (hb_codepoint_t glyph,
// 						Direction direction,
// 						Position *x, Position *y)
//    {
// 	 Position originX, originY;

// 	 get_glyph_origin_for_direction (glyph, direction, &originX, &originY);

// 	 *x += originX;
// 	 *y += originY;
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
