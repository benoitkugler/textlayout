// Package fcfonts is an implementation of
// the font tooling required by Pango, using textlayout/fontconfig
// and textlayout/fonts.
package fcfonts

import (
	"strings"

	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/pango"
)

// ported from pangofc-font.c:

var (
	_ pango.Font     = (*Font)(nil)
	_ pango.Coverage = (*coverage)(nil)
)

var (
	tagStrikeoutSize   = truetype.MustNewTag("strs")
	tagStrikeoutOffset = truetype.MustNewTag("stro")
	tagUnderlineSize   = truetype.MustNewTag("unds")
	tagUnderlineOffset = truetype.MustNewTag("undo")
)

type fontPrivate struct {
	decoder Decoder
	key     *PangoFcFontKey
}

type Font struct {
	fcFont

	//   FT_Face face;
	//   int load_flags;
	size int

	//   GSList *metrics_by_lang;

	glyphInfo map[pango.Glyph]*ft2GlyphInfo
	//   GDestroyNotify glyph_cache_destroy;
}

func newFont(pattern fc.Pattern, fontmap *FontMap) *Font {
	var ft2font Font
	if ds := pattern.GetFloats(fc.FC_PIXEL_SIZE); len(ds) != 0 {
		ft2font.size = int(ds[0] * float64(pango.PangoScale))
	}
	ft2font.fontPattern = pattern
	ft2font.fontmap = fontmap

	ft2font.glyphInfo = make(map[pango.Glyph]*ft2GlyphInfo)

	return &ft2font
}

type ft2GlyphInfo struct {
	logicalRect, inkRect pango.Rectangle
	cached_glyph         interface{}
}

func (font *Font) getGlyphInfo(glyph pango.Glyph, create bool) *ft2GlyphInfo {
	info := font.glyphInfo[glyph]

	if info == nil && create {
		info = new(ft2GlyphInfo)
		info.inkRect, info.logicalRect = font.getRawExtents(glyph)
		font.glyphInfo[glyph] = info
	}

	return info
}

func (font *Font) GetGlyphExtents(glyph pango.Glyph, inkRect, logicalRect *pango.Rectangle) {
	empty := false

	if glyph == pango.PANGO_GLYPH_EMPTY {
		glyph = font.getGlyph(' ')
		empty = true
	}

	if glyph&pango.PANGO_GLYPH_UNKNOWN_FLAG != 0 {
		metrics := pango.FontGetMetrics(font, "")
		if inkRect != nil {
			inkRect.X = pango.PangoScale
			inkRect.Width = metrics.ApproximateCharWidth - 2*pango.PangoScale
			inkRect.Y = -(metrics.Ascent - pango.PangoScale)
			inkRect.Height = metrics.Ascent + metrics.Descent - 2*pango.PangoScale
		}
		if logicalRect != nil {
			logicalRect.X = 0
			logicalRect.Width = metrics.ApproximateCharWidth
			logicalRect.Y = -metrics.Ascent
			logicalRect.Height = metrics.Ascent + metrics.Descent
		}
		return
	}

	info := font.getGlyphInfo(glyph, true)

	if inkRect != nil {
		*inkRect = info.inkRect
	}
	if logicalRect != nil {
		*logicalRect = info.logicalRect
	}

	if empty {
		if inkRect != nil {
			*inkRect = pango.Rectangle{}
		}
		if logicalRect != nil {
			logicalRect.X, logicalRect.Width = 0, 0
		}
	}
}

type PangoFcMetricsInfo struct {
	sampleStr string
	metrics   pango.FontMetrics
}

// fcFont is a base class for font implementations
// using the Fontconfig and FreeType libraries and is used in
// conjunction with `FontMap`.
type fcFont struct {
	// parent_instance pango.Font
	hbFont *pango.Hb_font_t // cached result of createHBFont

	fontPattern fc.Pattern   // fully resolved pattern
	fontmap     *FontMap     // associated map
	priv        fontPrivate  // used internally
	matrix      pango.Matrix // used internally
	description pango.FontDescription

	metricsByLang []PangoFcMetricsInfo

	isHinted      bool //  = 1;
	isTransformed bool //  = 1;
}

func (font *fcFont) Describe(absolute bool) pango.FontDescription {
	if !absolute {
		return font.description
	}

	desc := font.description

	size, ok := font.fontPattern.GetFloat(fc.FC_PIXEL_SIZE)
	if ok {
		desc.SetAbsoluteSize(int(size * float64(pango.PangoScale)))
	}

	return desc
}

func (font *fcFont) GetCoverage(_ pango.Language) pango.Coverage {
	if font.priv.decoder != nil {
		charset := font.priv.decoder.GetCharset(font)
		return fromCharset(charset)
	}

	if font.fontmap == nil {
		return &coverage{}
	}

	_, data := font.fontmap.getFontFaceData(font.fontPattern)
	if data == nil {
		return &coverage{}
	}

	if data.coverage == nil {
		// Pull the coverage out of the pattern, this doesn't require loading the font
		charset, _ := font.fontPattern.GetCharset(fc.FC_CHARSET)
		data.coverage = fromCharset(charset) // stores it into the map
	}

	return data.coverage
}

func (font *fcFont) GetFontMap() pango.FontMap { return font.fontmap }

// create a new font, which be cached
func (font *fcFont) createHBFont() *pango.Hb_font_t {
	xScaleInv, yScaleInv := 1.0, 1.
	size := 1.0

	key := font.priv.key
	if key != nil {
		pattern := key.pattern

		xScaleInv, yScaleInv = key.matrix.GetFontScaleFactors()

		fcMatrix := fc.Identity
		for _, fcMatrixVal := range pattern.GetMatrices(fc.FC_MATRIX) {
			fcMatrix = fcMatrix.Multiply(fcMatrixVal)
		}

		matrix2 := pango.Matrix{
			Xx: fcMatrix.Xx,
			Yx: fcMatrix.Yx,
			Xy: fcMatrix.Xy,
			Yy: fcMatrix.Yy,
		}
		x, y := matrix2.GetFontScaleFactors()

		xScaleInv /= x
		yScaleInv /= y

		gravity := key.pango_fc_font_key_get_gravity()
		if gravity.IsImproper() {
			xScaleInv = -xScaleInv
			yScaleInv = -yScaleInv
		}
		size = key.get_font_size()
	}

	xScale := 1. / xScaleInv
	yScale := 1. / yScaleInv

	hb_face := font.fontmap.getHBFace(font)

	hbFont := pango.HB_font_create(hb_face)
	pango.HB_font_set_scale(hbFont, size*pango.PangoScale*xScale, size*pango.PangoScale*yScale)

	if key != nil {
		axes := pango.HB_ot_var_get_axis_infos(hb_face)
		if len(axes) == 0 {
			return hbFont
		}
		nAxes := len(axes)

		coords := make([]float64, len(axes))

		for i, axe := range axes {
			coords[i] = axe.Def
		}

		if index, ok := key.pattern.GetInt(fc.FC_INDEX); ok && index != 0 {
			instance := (index >> 16) - 1
			pango.HB_ot_var_named_instance_get_design_coords(hb_face, instance, &nAxes, coords)
		}

		if variations, ok := key.pattern.GetString(fcFontVariations); ok {
			parseVariations(variations, axes, coords)
		}

		if key.variations != "" {
			parseVariations(key.variations, axes, coords)
		}

		pango.HB_font_set_var_coords_design(hbFont, coords)

	}

	return hbFont
}

// len(axes) == len(coords)
func parseVariations(variations string, axes []truetype.VarAxis, coords []float64) {
	varis := strings.Split(variations, ",")
	for _, varia := range varis {
		if vari, err := truetype.NewVariation(varia); err == nil {
			for i, axe := range axes {
				if axe.Tag == vari.Tag {
					coords[i] = vari.Value
					break
				}
			}
		}
	}
}

func (font *fcFont) GetHBFont() *pango.Hb_font_t {
	if font.hbFont != nil {
		return font.hbFont
	}
	font.hbFont = font.createHBFont()
	return font.hbFont
}

// getGlyph gets the glyph index for a given Unicode character
// for `font`. If you only want to determine
// whether the font has the glyph, use pango_fc_font_has_char().
// It returns 0 if the Unicode character doesn't exist in the font.
func (font *fcFont) getGlyph(wc rune) pango.Glyph {
	/* Replace NBSP with a normal space; it should be invariant that
	* they shape the same other than breaking properties. */
	if wc == 0xA0 {
		wc = 0x20
	}

	if font.priv.decoder != nil {
		return font.priv.decoder.GetGlyph(font, wc)
	}

	hbFont := font.GetHBFont()
	glyph := pango.AsUnknownGlyph(wc)

	glyph, _ = pango.HbFontGetNominalGlyph(hbFont, wc)

	return glyph

}

func (font *fcFont) GetMetrics(language pango.Language) pango.FontMetrics {
	sampleStr := language.GetSampleString()

	for _, info := range font.metricsByLang {
		if info.sampleStr == sampleStr {
			return info.metrics
		}
	}

	fontmap := font.fontmap
	if fontmap == nil {
		return pango.FontMetrics{}
	}

	/* Note: we need to add info to the list before calling
	* into PangoLayout below, to prevent recursion */
	font.metricsByLang = append(font.metricsByLang, PangoFcMetricsInfo{})
	info := &font.metricsByLang[len(font.metricsByLang)-1]
	info.sampleStr = sampleStr

	context := pango.NewContext(fontmap)
	context.SetLanguage(language)

	info.metrics = font.getFaceMetrics()

	// Compute derived metrics
	desc := font.Describe(true)
	//    gulong sampleStrWidth;

	layout := pango.NewLayout(context)
	layout.SetFontDescription(&desc)

	layout.SetText(sampleStr)
	var extents pango.Rectangle
	layout.GetExtents(nil, &extents)

	sampleStrWidth := len([]rune(sampleStr))
	info.metrics.ApproximateCharWidth = extents.Width / sampleStrWidth

	layout.SetText("0123456789")
	info.metrics.ApproximateDigitWidth = maxGlyphWidth(layout)

	return info.metrics
}

// The code in this function is partly based on code from Xft,
// Copyright 2000 Keith Packard
func (font *fcFont) getFaceMetrics() pango.FontMetrics {
	hbFont := font.GetHBFont()

	extents := pango.HBFontGetExtentsForDirection(hbFont, pango.HB_DIRECTION_LTR)

	var metrics pango.FontMetrics
	if fcMatrix, haveTransform := font.fontPattern.GetMatrix(fc.FC_MATRIX); haveTransform {
		metrics.Descent = -int(float64(extents.Descender) * fcMatrix.Yy)
		metrics.Ascent = int(float64(extents.Ascender) * fcMatrix.Yy)
		metrics.Height = int(float64(extents.Ascender-extents.Descender+extents.LineGap) * fcMatrix.Yy)
	} else {
		metrics.Descent = -int(extents.Descender)
		metrics.Ascent = int(extents.Ascender)
		metrics.Height = int(extents.Ascender) - int(extents.Descender) + int(extents.LineGap)
	}

	metrics.UnderlineThickness = pango.PangoScale
	metrics.UnderlinePosition = -pango.PangoScale
	metrics.StrikethroughThickness = pango.PangoScale
	metrics.StrikethroughPosition = metrics.Ascent / 2

	if position, ok := pango.HbOtMetricsGetPosition(hbFont, tagUnderlineSize); ok {
		metrics.UnderlineThickness = int(position)
	}

	if position, ok := pango.HbOtMetricsGetPosition(hbFont, tagUnderlineOffset); ok {
		metrics.UnderlinePosition = int(position)
	}

	if position, ok := pango.HbOtMetricsGetPosition(hbFont, tagStrikeoutSize); ok {
		metrics.StrikethroughThickness = int(position)
	}

	if position, ok := pango.HbOtMetricsGetPosition(hbFont, tagStrikeoutOffset); ok {
		metrics.StrikethroughPosition = int(position)
	}

	return metrics
}

func maxGlyphWidth(layout *pango.Layout) int {
	var maxWidth pango.GlyphUnit
	for _, line := range layout.GetLinesReadonly() {
		for r := line.Runs; r != nil; r = r.Next {
			glyphs := r.Data.Glyphs.Glyphs
			for _, g := range glyphs {
				if g.Geometry.Width > maxWidth {
					maxWidth = g.Geometry.Width
				}
			}
		}
	}
	return int(maxWidth)
}

// Gets the extents of a single glyph from a font. The extents are in
// user space; that is, they are not transformed by any matrix in effect
// for the font.
func (font *fcFont) getRawExtents(glyph pango.Glyph) (inkRect, logicalRect pango.Rectangle) {
	if glyph == pango.PANGO_GLYPH_EMPTY {
		return pango.Rectangle{}, pango.Rectangle{}
	}

	hbFont := font.GetHBFont()

	extents := pango.HB_font_get_glyph_extents(hbFont, glyph)
	font_extents := pango.HBFontGetExtentsForDirection(hbFont, pango.HB_DIRECTION_LTR)

	inkRect.X = int(extents.XBearing)
	inkRect.Width = int(extents.Width)
	inkRect.Y = -int(extents.YBearing)
	inkRect.Height = -int(extents.Height)

	x, _ := pango.HB_font_get_glyph_advance_for_direction(hbFont, glyph, pango.HB_DIRECTION_LTR)

	logicalRect.X = 0
	logicalRect.Width = int(x)
	logicalRect.Y = -int(font_extents.Ascender)
	logicalRect.Height = int(font_extents.Ascender - font_extents.Descender)

	return
}

//  func load_fallback_face (PangoFT2Font *ft2font,
// 			 const char   *original_file) {
//    PangoFcFont *fcfont = PANGO_FC_FONT (ft2font);
//    FcPattern *sans;
//    FcPattern *matched;
//    FcResult result;
//    FT_Error error;
//    FcChar8 *filename2 = nil;
//    gchar *name;
//    int id;

//    sans = FcPatternBuild (nil,
// 			  FC_FAMILY,     FcTypeString, "sans",
// 			  FC_PIXEL_SIZE, FcTypeDouble, (double)ft2font.size / pango.PangoScale,
// 			  nil);

//    _pango_ft2_font_map_default_substitute ((PangoFcFontMap *)fcfont.fontmap, sans);

//    matched = FcFontMatch (pango_fc_font_map_get_config ((PangoFcFontMap *)fcfont.fontmap), sans, &result);

//    if (FcPatternGetString (matched, FC_FILE, 0, &filename2) != FcResultMatch)
// 	 goto bail1;

//    if (FcPatternGetInteger (matched, FC_INDEX, 0, &id) != FcResultMatch)
// 	 goto bail1;

//    error = FT_New_Face (_pango_ft2_font_map_get_library (fcfont.fontmap),
// 				(char *) filename2, id, &ft2font.face);

//    if (error)
// 	 {
// 	 bail1:
// 	   name = pango_font_description_to_string (fcfont.description);
// 	   g_error ("Unable to open font file %s for font %s, exiting\n", filename2, name);
// 	 }
//    else
// 	 {
// 	   name = pango_font_description_to_string (fcfont.description);
// 	   g_warning ("Unable to open font file %s for font %s, falling back to %s\n", original_file, name, filename2);
// 	   g_free (name);
// 	 }

//    FcPatternDestroy (sans);
//    FcPatternDestroy (matched);
//  }

// func set_transform (PangoFT2Font *ft2font) {
//    PangoFcFont *fcfont = (PangoFcFont *)ft2font;
//    FcMatrix *fcMatrix;

//    if (FcPatternGetMatrix (fcfont.font_pattern, FC_MATRIX, 0, &fcMatrix) == FcResultMatch)
// 	 {
// 	   FT_Matrix ft_matrix;

// 	   ft_matrix.xx = 0x10000L * fcMatrix.xx;
// 	   ft_matrix.yy = 0x10000L * fcMatrix.yy;
// 	   ft_matrix.xy = 0x10000L * fcMatrix.xy;
// 	   ft_matrix.yx = 0x10000L * fcMatrix.yx;

// 	   FT_Set_Transform (ft2font.face, &ft_matrix, nil);
// 	 }
//  }

//  /**
//   * pango_ft2_font_get_face: (skip)
//   * @font: a #PangoFont
//   *
//   * Returns the native FreeType2 `FT_Face` structure used for this #PangoFont.
//   * This may be useful if you want to use FreeType2 functions directly.
//   *
//   * Use pango_fc_font_lock_face() instead; when you are done with a
//   * face from pango_fc_font_lock_face() you must call
//   * pango_fc_font_unlock_face().
//   *
//   * Return value: (nullable): a pointer to a `FT_Face` structure, with the
//   *   size set correctly, or %nil if @font is %nil.
//   **/
// func  pango_ft2_font_get_face (PangoFont *font)  FT_Face  {
//    PangoFT2Font *ft2font = (PangoFT2Font *)font;
//    PangoFcFont *fcfont = (PangoFcFont *)font;
//    FT_Error error;
//    FcPattern *pattern;
//    FcChar8 *filename;
//    FcBool antialias, hinting, autohint;
//    int hintstyle;
//    int id;

//    if (G_UNLIKELY (!font))
// 	 return nil;

//    pattern = fcfont.font_pattern;

//    if (!ft2font.face)
// 	 {
// 	   ft2font.load_flags = 0;

// 	   /* disable antialiasing if requested */
// 	   if (FcPatternGetBool (pattern,
// 				 FC_ANTIALIAS, 0, &antialias) != FcResultMatch)
// 	 antialias = FcTrue;

// 	   if (antialias)
// 	 ft2font.load_flags |= FT_LOAD_NO_BITMAP;
// 	   else
// 	 ft2font.load_flags |= FT_LOAD_TARGET_MONO;

// 	   /* disable hinting if requested */
// 	   if (FcPatternGetBool (pattern,
// 				 FC_HINTING, 0, &hinting) != FcResultMatch)
// 	 hinting = FcTrue;

//  #ifdef FC_HINT_STYLE
// 	   if (FcPatternGetInteger (pattern, FC_HINT_STYLE, 0, &hintstyle) != FcResultMatch)
// 	 hintstyle = FC_HINT_FULL;

// 	   if (!hinting || hintstyle == FC_HINT_NONE)
// 		   ft2font.load_flags |= FT_LOAD_NO_HINTING;

// 	   switch (hintstyle) {
// 	   case FC_HINT_SLIGHT:
// 	   case FC_HINT_MEDIUM:
// 	 ft2font.load_flags |= FT_LOAD_TARGET_LIGHT;
// 	 break;
// 	   default:
// 	 ft2font.load_flags |= FT_LOAD_TARGET_NORMAL;
// 	 break;
// 	   }
//  #else
// 	   if (!hinting)
// 		   ft2font.load_flags |= FT_LOAD_NO_HINTING;
//  #endif

// 	   /* force autohinting if requested */
// 	   if (FcPatternGetBool (pattern,
// 				 FC_AUTOHINT, 0, &autohint) != FcResultMatch)
// 	 autohint = FcFalse;

// 	   if (autohint)
// 	 ft2font.load_flags |= FT_LOAD_FORCE_AUTOHINT;

// 	   if (FcPatternGetString (pattern, FC_FILE, 0, &filename) != FcResultMatch)
// 		   goto bail0;

// 	   if (FcPatternGetInteger (pattern, FC_INDEX, 0, &id) != FcResultMatch)
// 		   goto bail0;

// 	   error = FT_New_Face (_pango_ft2_font_map_get_library (fcfont.fontmap),
// 				(char *) filename, id, &ft2font.face);
// 	   if (error != FT_Err_Ok)
// 	 {
// 	 bail0:
// 	   load_fallback_face (ft2font, (char *) filename);
// 	 }

// 	   g_assert (ft2font.face);

// 	   set_transform (ft2font);

// 	   error = FT_Set_Char_Size (ft2font.face,
// 				 PANGO_PIXELS_26_6 (ft2font.size),
// 				 PANGO_PIXELS_26_6 (ft2font.size),
// 				 0, 0);
// 	   if (error)
// 	 g_warning ("Error in FT_Set_Char_Size: %d", error);
// 	 }

//    return ft2font.face;
//  }

//  static void
//  pango_ft2_font_class_init (PangoFT2FontClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoFontClass *font_class = PANGO_FONT_CLASS (class);
//    PangoFcFontClass *fc_font_class = PANGO_FC_FONT_CLASS (class);

//    object_class.finalize = pango_ft2_font_finalize;

//    font_class.get_glyph_extents = pango_ft2_font_get_glyph_extents;

//    fc_font_class.lock_face = pango_ft2_font_real_lock_face;
//    fc_font_class.unlock_face = pango_ft2_font_real_unlock_face;
//  }

//  /**
//   * pango_ft2_font_get_kerning:
//   * @font: a #PangoFont
//   * @left: the left #PangoGlyph
//   * @right: the right #PangoGlyph
//   *
//   * Retrieves kerning information for a combination of two glyphs.
//   *
//   * Use pango_fc_font_kern_glyphs() instead.
//   *
//   * Return value: The amount of kerning (in Pango units) to apply for
//   * the given combination of glyphs.
//   **/
//  int
//  pango_ft2_font_get_kerning (PangoFont *font,
// 				 PangoGlyph left,
// 				 PangoGlyph right)
//  {
//    PangoFcFont *fc_font = PANGO_FC_FONT (font);

//    FT_Face face;
//    FT_Error error;
//    FT_Vector kerning;

//    face = pango_fc_font_lock_face (fc_font);
//    if (!face)
// 	 return 0;

//    if (!FT_HAS_KERNING (face))
// 	 {
// 	   pango_fc_font_unlock_face (fc_font);
// 	   return 0;
// 	 }

//    error = FT_Get_Kerning (face, left, right, ft_kerning_default, &kerning);
//    if (error != FT_Err_Ok)
// 	 {
// 	   pango_fc_font_unlock_face (fc_font);
// 	   return 0;
// 	 }

//    pango_fc_font_unlock_face (fc_font);
//    return PANGO_UNITS_26_6 (kerning.x);
//  }

//  static FT_Face
//  pango_ft2_font_real_lock_face (font *PangoFcFont)
//  {
//    return pango_ft2_font_get_face ((PangoFont *)font);
//  }

//  static void
//  pango_ft2_font_real_unlock_face (font *PangoFcFont G_GNUC_UNUSED)
//  {
//  }

//  /* Utility functions */

//  /**
//   * pango_ft2_get_unknown_glyph:
//   * @font: a #PangoFont
//   *
//   * Return the index of a glyph suitable for drawing unknown characters with
//   * @font, or %PANGO_GLYPH_EMPTY if no suitable glyph found.
//   *
//   * If you want to draw an unknown-box for a character that is not covered
//   * by the font,
//   * use AsUnknownGlyph() instead.
//   *
//   * Return value: a glyph index into @font, or %PANGO_GLYPH_EMPTY
//   **/
//  PangoGlyph
//  pango_ft2_get_unknown_glyph (PangoFont *font)
//  {
//    FT_Face face = pango_ft2_font_get_face (font);
//    if (face && FT_IS_SFNT (face))
// 	 /* TrueType fonts have an 'unknown glyph' box on glyph index 0 */
// 	 return 0;
//    else
// 	 return PANGO_GLYPH_EMPTY;
//  }

//  void *
//  _pango_ft2_font_get_cache_glyph_data (PangoFont *font,
// 					  int        glyph_index)
//  {
//    ft2GlyphInfo *info;

//    if (!PANGO_FT2_IS_FONT (font))
// 	 return nil;

//    info = getGlyphInfo (font, glyph_index, false);

//    if (info == nil)
// 	 return nil;

//    return info.cached_glyph;
//  }

//  void
//  _pango_ft2_font_set_cache_glyph_data (PangoFont     *font,
// 					  int            glyph_index,
// 					  void          *cached_glyph)
//  {
//    ft2GlyphInfo *info;

//    if (!PANGO_FT2_IS_FONT (font))
// 	 return;

//    info = getGlyphInfo (font, glyph_index, true);

//    info.cached_glyph = cached_glyph;

//    /* TODO: Implement limiting of the number of cached glyphs */
//  }

//  void
//  _pango_ft2_font_set_glyph_cache_destroy (PangoFont      *font,
// 					  GDestroyNotify  destroy_notify)
//  {
//    if (!PANGO_FT2_IS_FONT (font))
// 	 return;

//    PANGO_FT2_FONT (font).glyph_cache_destroy = destroy_notify;
//  }

//  #define PANGO_FC_TYPE_FAMILY            (pango_fc_family_get_type ())
//  #define PANGO_FC_FAMILY(object)         (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FAMILY, PangoFcFamily))
//  #define PANGO_FC_IS_FAMILY(object)      (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FAMILY))

//  #define PANGO_FC_TYPE_FACE              (pango_fc_face_get_type ())
//  #define PANGO_FC_FACE(object)           (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FACE, PangoFcFace))
//  #define PANGO_FC_IS_FACE(object)        (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FACE))

//  #define PANGO_FC_TYPE_FONTSET           (pango_fc_fontset_get_type ())
//  #define PANGO_FC_FONTSET(object)        (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FONTSET, PangoFcFontset))
//  #define PANGO_FC_IS_FONTSET(object)     (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FONTSET))

//  enum {
//    PROP_0,
//    PROP_PATTERN,
//    PROP_FONTMAP
//  };

//  typedef struct _PangoFcFontPrivate PangoFcFontPrivate;

//  struct _PangoFcFontPrivate
//  {
//    PangoFcDecoder *decoder;
//    PangoFcFontKey *key;
//  };

//  #define PANGO_FC_FONT_LOCK_FACE(font)	(PANGO_FC_FONT_GET_CLASS (font).lock_face (font))
//  #define PANGO_FC_FONT_UNLOCK_FACE(font)	(PANGO_FC_FONT_GET_CLASS (font).unlock_face (font))

//  G_DEFINE_ABSTRACT_TYPE_WITH_CODE (PangoFcFont, pango_fc_font, PANGO_TYPE_FONT,
// 								   G_ADD_PRIVATE (PangoFcFont))

//  static void
//  pango_fc_font_class_init (PangoFcFontClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoFontClass *font_class = PANGO_FONT_CLASS (class);

//    class.has_char = pango_fc_font_real_has_char;
//    class.get_glyph = pango_fc_font_real_get_glyph;
//    class.get_unknown_glyph = nil;

//    object_class.finalize = pango_fc_font_finalize;
//    object_class.set_property = pango_fc_font_set_property;
//    object_class.get_property = pango_fc_font_get_property;
//    font_class.describe = pango_fc_font_describe;
//    font_class.describe_absolute = pango_fc_font_describe_absolute;
//    font_class.GetCoverage = GetCoverage;
//    font_class.GetMetrics = GetMetrics;
//    font_class.GetFontMap = pango_fc_font_get_font_map;
//    font_class.GetFeatures = pango_fc_font_get_features;
//    font_class.CreateHBFont = createHBFont;
//    font_class.GetFeatures = pango_fc_font_get_features;

//    g_object_class_install_property (object_class, PROP_PATTERN,
// 					g_param_spec_pointer ("pattern",
// 							  "Pattern",
// 							  "The fontconfig pattern for this font",
// 							  G_PARAM_READWRITE | G_PARAM_CONSTRUCT_ONLY |
// 							  G_PARAM_STATIC_STRINGS));
//    g_object_class_install_property (object_class, PROP_FONTMAP,
// 					g_param_spec_object ("fontmap",
// 							 "Font Map",
// 							 "The PangoFc font map this font is associated with (Since: 1.26)",
// 							 PANGO_TYPE_FC_FONT_MAP,
// 							 G_PARAM_READWRITE |
// 							 G_PARAM_STATIC_STRINGS));
//  }

//  static void
//  pango_fc_font_init (PangoFcFont *font)
//  {
//    font.priv = pango_fc_font_get_instance_private (font);
//  }

//  static void
//  free_metrics_info (PangoFcMetricsInfo *info)
//  {
//    pango_font_metrics_unref (info.metrics);
//    g_slice_free (PangoFcMetricsInfo, info);
//  }

//  static void
//  pango_fc_font_finalize (GObject *object)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (object);
//    PangoFcFontPrivate *priv = fcfont.priv;
//    PangoFcFontMap *fontmap;

//    g_slist_foreach (fcfont.metrics_by_lang, (GFunc)free_metrics_info, nil);
//    g_slist_free (fcfont.metrics_by_lang);

//    fontmap = g_weak_ref_get ((GWeakRef *) &fcfont.fontmap);
//    if (fontmap)
// 	 {
// 	   _pango_fc_font_map_remove (PANGO_FC_FONT_MAP (fcfont.fontmap), fcfont);
// 	   g_weak_ref_clear ((GWeakRef *) &fcfont.fontmap);
// 	   g_object_unref (fontmap);
// 	 }

//    FcPatternDestroy (fcfont.font_pattern);
//    pango_font_description_free (fcfont.description);

//    if (priv.decoder)
// 	 _pango_fc_font_set_decoder (fcfont, nil);

//    G_OBJECT_CLASS (pango_fc_font_parent_class).finalize (object);
//  }

//  static gboolean
//  pattern_is_hinted (FcPattern *pattern)
//  {
//    FcBool hinting;

//    if (FcPatternGetBool (pattern,
// 			 FC_HINTING, 0, &hinting) != FcResultMatch)
// 	 hinting = FcTrue;

//    return hinting;
//  }

//  static gboolean
//  pattern_is_transformed (FcPattern *pattern)
//  {
//    FcMatrix *fcMatrix;

//    if (FcPatternGetMatrix (pattern, FC_MATRIX, 0, &fcMatrix) == FcResultMatch)
// 	 {
// 	   return fcMatrix.xx != 1 || fcMatrix.xy != 0 ||
// 			  fcMatrix.yx != 0 || fcMatrix.yy != 1;
// 	 }
//    else
// 	 return false;
//  }

//  static void
//  pango_fc_font_set_property (GObject       *object,
// 				 guint          prop_id,
// 				 const GValue  *value,
// 				 GParamSpec    *pspec)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (object);

//    switch (prop_id)
// 	 {
// 	 case PROP_PATTERN:
// 	   {
// 	 FcPattern *pattern = g_value_get_pointer (value);

// 	 g_return_if_fail (pattern != nil);
// 	 g_return_if_fail (fcfont.font_pattern == nil);

// 	 FcPatternReference (pattern);
// 	 fcfont.font_pattern = pattern;
// 	 fcfont.description = pango_fc_font_description_from_pattern (pattern, true);
// 	 fcfont.is_hinted = pattern_is_hinted (pattern);
// 	 fcfont.is_transformed = pattern_is_transformed (pattern);
// 	   }
// 	   goto set_decoder;

// 	 case PROP_FONTMAP:
// 	   {
// 	 PangoFcFontMap *fcfontmap = PANGO_FC_FONT_MAP (g_value_get_object (value));

// 	 g_return_if_fail (fcfont.fontmap == nil);
// 	 g_weak_ref_set ((GWeakRef *) &fcfont.fontmap, fcfontmap);
// 	   }
// 	   goto set_decoder;

// 	 default:
// 	   G_OBJECT_WARN_INVALID_PROPERTY_ID (object, prop_id, pspec);
// 	   return;
// 	 }

//  set_decoder:
//    /* set decoder if both pattern and fontmap are set now */
//    if (fcfont.font_pattern && fcfont.fontmap)
// 	 _pango_fc_font_set_decoder (fcfont,
// 				 pango_fc_font_map_find_decoder  ((PangoFcFontMap *) fcfont.fontmap,
// 								  fcfont.font_pattern));
//  }

//  static void
//  pango_fc_font_get_property (GObject       *object,
// 				 guint          prop_id,
// 				 GValue        *value,
// 				 GParamSpec    *pspec)
//  {
//    switch (prop_id)
// 	 {
// 	 case PROP_PATTERN:
// 	   {
// 	 PangoFcFont *fcfont = PANGO_FC_FONT (object);
// 	 g_value_set_pointer (value, fcfont.font_pattern);
// 	   }
// 	   break;
// 	 case PROP_FONTMAP:
// 	   {
// 	 PangoFcFont *fcfont = PANGO_FC_FONT (object);
// 	 PangoFontMap *fontmap = g_weak_ref_get ((GWeakRef *) &fcfont.fontmap);
// 	 g_value_take_object (value, fontmap);
// 	   }
// 	   break;
// 	 default:
// 	   G_OBJECT_WARN_INVALID_PROPERTY_ID (object, prop_id, pspec);
// 	   break;
// 	 }
//  }

//  PangoFontMetrics *
//  pango_fc_font_create_base_metrics_for_context (PangoFcFont   *fcfont,
// 							PangoContext  *context)
//  {
//    PangoFontMetrics *metrics;
//    metrics = pango_font_metrics_new ();

//    getFaceMetrics (fcfont, metrics);

//    return metrics;
//  }

//  static PangoFontMap *
//  pango_fc_font_get_font_map (font *PangoFcFont)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (font);

//    /* MT-unsafe.  Oh well...  The API is unsafe. */
//    return fcfont.fontmap;
//  }

//  static gboolean
//  pango_fc_font_real_has_char (font *PangoFcFont,
// 				  gunichar     wc)
//  {
//    FcCharSet *charset;

//    if (FcPatternGetCharSet (font.font_pattern,
// 				FC_CHARSET, 0, &charset) != FcResultMatch)
// 	 return false;

//    return FcCharSetHasChar (charset, wc);
//  }

//  /**
//   * pango_fc_font_lock_face: (skip)
//   * @font: a #PangoFcFont.
//   *
//   * Gets the FreeType `FT_Face` associated with a font,
//   * This face will be kept around until you call
//   * pango_fc_font_unlock_face().
//   *
//   * Return value: the FreeType `FT_Face` associated with @font.
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use GetHBFont() instead
//   **/
//  FT_Face
//  pango_fc_font_lock_face (font *PangoFcFont)
//  {
//    g_return_val_if_fail (PANGO_IS_FC_FONT (font), nil);

//    return PANGO_FC_FONT_LOCK_FACE (font);
//  }

//  /**
//   * pango_fc_font_unlock_face:
//   * @font: a #PangoFcFont.
//   *
//   * Releases a font previously obtained with
//   * pango_fc_font_lock_face().
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use GetHBFont() instead
//   **/
//  void
//  pango_fc_font_unlock_face (font *PangoFcFont)
//  {
//    g_return_if_fail (PANGO_IS_FC_FONT (font));

//    PANGO_FC_FONT_UNLOCK_FACE (font);
//  }

//  /**
//   * pango_fc_font_has_char:
//   * @font: a #PangoFcFont
//   * @wc: Unicode codepoint to look up
//   *
//   * Determines whether @font has a glyph for the codepoint @wc.
//   *
//   * Return value: %true if @font has the requested codepoint.
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use pango_font_has_char()
//   **/
//  gboolean
//  pango_fc_font_has_char (font *PangoFcFont,
// 			 gunichar     wc)
//  {
//    PangoFcFontPrivate *priv = font.priv;
//    FcCharSet *charset;

//    g_return_val_if_fail (PANGO_IS_FC_FONT (font), false);

//    if (priv.decoder)
// 	 {
// 	   charset = pango_fc_decoder_get_charset (priv.decoder, font);
// 	   return FcCharSetHasChar (charset, wc);
// 	 }

//    return PANGO_FC_FONT_GET_CLASS (font).has_char (font, wc);
//  }

//  /**
//   * pango_fc_font_get_unknown_glyph:
//   * @font: a #PangoFcFont
//   * @wc: the Unicode character for which a glyph is needed.
//   *
//   * Returns the index of a glyph suitable for drawing @wc as an
//   * unknown character.
//   *
//   * Use AsUnknownGlyph() instead.
//   *
//   * Return value: a glyph index into @font.
//   *
//   * Since: 1.4
//   **/
//  PangoGlyph
//  pango_fc_font_get_unknown_glyph (font *PangoFcFont,
// 				  gunichar     wc)
//  {
//    if (font && PANGO_FC_FONT_GET_CLASS (font).get_unknown_glyph)
// 	 return PANGO_FC_FONT_GET_CLASS (font).get_unknown_glyph (font, wc);

//    return AsUnknownGlyph (wc);
//  }

//  void
//  _pango_fc_font_shutdown (font *PangoFcFont)
//  {
//    g_return_if_fail (PANGO_IS_FC_FONT (font));

//    if (PANGO_FC_FONT_GET_CLASS (font).shutdown)
// 	 PANGO_FC_FONT_GET_CLASS (font).shutdown (font);
//  }

//  /**
//   * pango_fc_font_kern_glyphs:
//   * @font: a #PangoFcFont
//   * @glyphs: a #PangoGlyphString
//   *
//   * This function used to adjust each adjacent pair of glyphs
//   * in @glyphs according to kerning information in @font.
//   *
//   * Since 1.44, it does nothing.
//   *
//   *
//   * Since: 1.4
//   * Deprecated: 1.32
//   **/
//  void
//  pango_fc_font_kern_glyphs (PangoFcFont      *font,
// 				PangoGlyphString *glyphs)
//  {
//  }

//  /**
//   * _pango_fc_font_get_decoder:
//   * @font: a #PangoFcFont
//   *
//   * This will return any custom decoder set on this font.
//   *
//   * Return value: The custom decoder
//   *
//   * Since: 1.6
//   **/

//  PangoFcDecoder *
//  _pango_fc_font_get_decoder (font *PangoFcFont)
//  {
//    PangoFcFontPrivate *priv = font.priv;

//    return priv.decoder;
//  }

//  /**
//   * _pango_fc_font_set_decoder:
//   * @font: a #PangoFcFont
//   * @decoder: a #PangoFcDecoder to set for this font
//   *
//   * This sets a custom decoder for this font.  Any previous decoder
//   * will be released before this one is set.
//   *
//   * Since: 1.6
//   **/

//  void
//  _pango_fc_font_set_decoder (PangoFcFont    *font,
// 				 PangoFcDecoder *decoder)
//  {
//    PangoFcFontPrivate *priv = font.priv;

//    if (priv.decoder)
// 	 g_object_unref (priv.decoder);

//    priv.decoder = decoder;

//    if (priv.decoder)
// 	 g_object_ref (priv.decoder);
//  }

//  PangoFcFontKey *
//  _pango_fc_font_get_font_key (PangoFcFont *fcfont)
//  {
//    PangoFcFontPrivate *priv = fcfont.priv;

//    return priv.key;
//  }

//  void
//  _pango_fc_font_set_font_key (fcfont *PangoFcFont,
// 				  PangoFcFontKey *key)
//  {
//    PangoFcFontPrivate *priv = fcfont.priv;

//    priv.key = key;
//  }

//  static void
//  pango_fc_font_get_features (PangoFont    *font,
// 							 hb_feature_t *features,
// 							 guint         len,
// 							 guint        *num_features)
//  {
//    /* Setup features from fontconfig pattern. */
//    PangoFcFont *fc_font = PANGO_FC_FONT (font);
//    if (fc_font.font_pattern)
// 	 {
// 	   char *s;
// 	   while (*num_features < len &&
// 			  FcResultMatch == FcPatternGetString (fc_font.font_pattern,
// 												   PANGO_FC_FONT_FEATURES,
// 												   *num_features,
// 												   (FcChar8 **) &s))
// 		 {
// 		   gboolean ret = hb_feature_from_string (s, -1, &features[*num_features]);
// 		   features[*num_features].start = 0;
// 		   features[*num_features].end   = (unsigned int) -1;
// 		   if (ret)
// 			 (*num_features)++;
// 		 }
// 	 }
//  }

//  extern gpointer get_gravity_class (void);

//  /**
//   * pango_fc_font_get_languages:
//   * @font: a #PangoFcFont
//   *
//   * Returns the languages that are supported by @font.
//   *
//   * This corresponds to the FC_LANG member of the FcPattern.
//   *
//   * The returned array is only valid as long as the font
//   * and its fontmap are valid.
//   *
//   * Returns: (transfer none) (nullable): a %nil-terminated
//   *    array of PangoLanguage*
//   *
//   * Since: 1.48
//   */
//  PangoLanguage **
//  pango_fc_font_get_languages (font *PangoFcFont)
//  {
//    PangoFcFontMap *fontmap;
//    PangoLanguage **languages;

//    fontmap = g_weak_ref_get ((GWeakRef *) &font.fontmap);
//    if (!fontmap)
// 	 return nil;

//    languages  = _pango_fc_font_map_get_languages (fontmap, font);
//    g_object_unref (fontmap);

//    return languages;
//  }

//  /**
//   * pango_fc_font_get_pattern: (skip)
//   * @font: a #PangoFcFont
//   *
//   * Returns the FcPattern that @font is based on.
//   *
//   * Returns: the fontconfig pattern for this font
//   *
//   * Since: 1.48
//   */
//  FcPattern *
//  pango_fc_font_get_pattern (font *PangoFcFont)
//  {
//    return font.font_pattern;
//  }
