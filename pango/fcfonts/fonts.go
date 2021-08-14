// Package fcfonts is an implementation of
// the font tooling required by Pango, using textlayout/fontconfig
// and textlayout/fonts.
//
// The entry point of the package is the `NewFontMap` constructor.
package fcfonts

import (
	"log"
	"strings"

	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
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
	decoder decoder
	key     *fcFontKey
}

// Font implements the pango.Font interface.
type Font struct {
	glyphInfo map[pango.Glyph]*ft2GlyphInfo

	fcFont

	size int
}

func newFont(pattern fc.Pattern, fontmap *FontMap) *Font {
	var ft2font Font

	ft2font.Pattern = pattern
	ft2font.description = newFontDescriptionFromPattern(pattern, true)

	ft2font.fontmap = fontmap

	if ds := pattern.GetFloats(fc.PIXEL_SIZE); len(ds) != 0 {
		ft2font.size = int(ds[0] * float32(pango.Scale))
	}
	ft2font.glyphInfo = make(map[pango.Glyph]*ft2GlyphInfo)

	return &ft2font
}

func slantToPango(fc_style int32) pango.Style {
	switch fc_style {
	case fc.SLANT_ROMAN:
		return pango.STYLE_NORMAL
	case fc.SLANT_ITALIC:
		return pango.STYLE_ITALIC
	case fc.SLANT_OBLIQUE:
		return pango.STYLE_OBLIQUE
	default:
		return pango.STYLE_NORMAL
	}
}

func widthToPango(fc_stretch int32) pango.Stretch {
	switch fc_stretch {
	case fc.WIDTH_NORMAL:
		return pango.STRETCH_NORMAL
	case fc.WIDTH_ULTRACONDENSED:
		return pango.STRETCH_ULTRA_CONDENSED
	case fc.WIDTH_EXTRACONDENSED:
		return pango.STRETCH_EXTRA_CONDENSED
	case fc.WIDTH_CONDENSED:
		return pango.STRETCH_CONDENSED
	case fc.WIDTH_SEMICONDENSED:
		return pango.STRETCH_SEMI_CONDENSED
	case fc.WIDTH_SEMIEXPANDED:
		return pango.STRETCH_SEMI_EXPANDED
	case fc.WIDTH_EXPANDED:
		return pango.STRETCH_EXPANDED
	case fc.WIDTH_EXTRAEXPANDED:
		return pango.STRETCH_EXTRA_EXPANDED
	case fc.WIDTH_ULTRAEXPANDED:
		return pango.STRETCH_ULTRA_EXPANDED
	default:
		return pango.STRETCH_NORMAL
	}
}

// newFontDescriptionFromPattern creates a FontDescription that matches the specified
// Fontconfig pattern as closely as possible. Many possible Fontconfig
// pattern values, such as RASTERIZER or DPI, don't make sense in
// the context of FontDescription, so will be ignored.
// If `includeSize` is TRUE, the description will include the size from
//   `pattern`; otherwise the resulting description will be unsized.
//   (only SIZE is examined, not PIXEL_SIZE)
func newFontDescriptionFromPattern(pattern fc.Pattern, includeSize bool) pango.FontDescription {
	desc := pango.NewFontDescription()

	fam, _ := pattern.GetString(fc.FAMILY)
	desc.SetFamily(fam)

	style := pango.STYLE_NORMAL
	if i, ok := pattern.GetInt(fc.SLANT); ok {
		style = slantToPango(i)
	}
	desc.SetStyle(style)

	weight := pango.WEIGHT_NORMAL
	if d, ok := pattern.GetFloat(fc.WEIGHT); ok {
		weight = pango.Weight(fc.WeightToOT(d))
	}
	desc.SetWeight(weight)

	stretch := pango.STRETCH_NORMAL
	if i, ok := pattern.GetInt(fc.WIDTH); ok {
		stretch = widthToPango(i)
	}
	desc.SetStretch(stretch)

	desc.SetVariant(pango.VARIANT_NORMAL)

	if size, ok := pattern.GetFloat(fc.SIZE); includeSize && ok {
		var scale_factor float32 = 1

		if fcMatrix, ok := pattern.GetMatrix(fc.MATRIX); ok {
			mat := pango.Identity

			mat.Xx = fcMatrix.Xx
			mat.Xy = fcMatrix.Xy
			mat.Yx = fcMatrix.Yx
			mat.Yy = fcMatrix.Yy
			_, scale_factor = mat.GetFontScaleFactors()
		}

		desc.SetSize(int(scale_factor * size * float32(pango.Scale)))
	}

	/* gravity is a bit different.  we don't want to set it if it was not set on
	* the pattern */
	if s, ok := pattern.GetString(fcGravity); ok {
		gravity, _ := pango.GravityMap.FromString(s)
		desc.SetGravity(pango.Gravity(gravity))
	}

	if s, ok := pattern.GetString(fcFontVariations); includeSize && ok {
		if s != "" {
			desc.SetVariations(s)
		}
	}

	return desc
}

type ft2GlyphInfo struct {
	cached_glyph         interface{}
	logicalRect, inkRect pango.Rectangle
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

func (font *Font) GlyphExtents(glyph pango.Glyph, inkRect, logicalRect *pango.Rectangle) {
	empty := false

	if glyph == pango.GLYPH_EMPTY {
		glyph = font.getGlyph(' ')
		empty = true
	}

	if glyph&pango.GLYPH_UNKNOWN_FLAG != 0 {
		metrics := pango.FontGetMetrics(font, "")
		if inkRect != nil {
			inkRect.X = pango.Scale
			inkRect.Width = metrics.ApproximateCharWidth - 2*pango.Scale
			inkRect.Y = -(metrics.Ascent - pango.Scale)
			inkRect.Height = metrics.Ascent + metrics.Descent - 2*pango.Scale
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

type fcMetricsInfo struct {
	sampleStr string
	metrics   pango.FontMetrics
}

// fcFont is a base class for font implementations
// using the Fontconfig and FreeType libraries and is used in
// conjunction with `FontMap`.
type fcFont struct {
	priv   fontPrivate    // used internally
	hbFont *harfbuzz.Font // cached result of createHBFont

	fontmap       *FontMap   // associated map
	Pattern       fc.Pattern // fully resolved pattern
	metricsByLang []fcMetricsInfo
	description   pango.FontDescription
	matrix        pango.Matrix // used internally
}

func (font *fcFont) Describe(absolute bool) pango.FontDescription {
	if !absolute {
		return font.description
	}

	desc := font.description

	size, ok := font.Pattern.GetFloat(fc.PIXEL_SIZE)
	if ok {
		desc.SetAbsoluteSize(int(size * float32(pango.Scale)))
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

	_, data := font.fontmap.getFontFaceData(font.Pattern)
	if data == nil {
		return &coverage{}
	}

	if data.coverage == nil {
		// Pull the coverage out of the pattern, this doesn't require loading the font
		charset, _ := font.Pattern.GetCharset(fc.CHARSET)
		data.coverage = fromCharset(charset) // stores it into the map
	}

	return data.coverage
}

func (font *fcFont) GetFontMap() pango.FontMap { return font.fontmap }

// create a new font, which will be cached
func (font *fcFont) createHBFont() (*harfbuzz.Font, error) {
	var (
		xScaleInv, yScaleInv float32 = 1.0, 1.
		size                 float32 = 1.0
	)

	key := font.priv.key
	if key != nil {
		pattern := key.pattern

		xScaleInv, yScaleInv = key.matrix.GetFontScaleFactors()

		fcMatrix := fc.Identity
		for _, fcMatrixVal := range pattern.GetMatrices(fc.MATRIX) {
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

		gravity := key.pango_font_key_get_gravity()
		if gravity.IsImproper() {
			xScaleInv = -xScaleInv
			yScaleInv = -yScaleInv
		}
		size = key.getFontSize()
	}

	xScale := 1. / xScaleInv
	yScale := 1. / yScaleInv

	hb_face, err := font.fontmap.getHBFace(font)
	if err != nil {
		return nil, err
	}

	hbFont := harfbuzz.NewFont(hb_face)
	hbFont.XScale, hbFont.YScale = int32(size*pango.Scale*xScale), int32(size*pango.Scale*yScale)
	if varFont, isVariable := hb_face.(truetype.FaceVariable); key != nil && isVariable {
		fvar := varFont.Variations()
		if len(fvar.Axis) == 0 {
			return hbFont, nil
		}

		coords := fvar.GetDesignCoordsDefault(nil)

		if index, ok := key.pattern.GetInt(fc.INDEX); ok && index != 0 {
			if instance := (index >> 16) - 1; int(instance) < len(fvar.Instances) {
				coords = fvar.Instances[instance].Coords
			}
		}

		if variations, ok := key.pattern.GetString(fcFontVariations); ok {
			vars := parseVariations(variations)
			fvar.GetDesignCoords(vars, coords)
		}

		if key.variations != "" {
			vars := parseVariations(key.variations)
			fvar.GetDesignCoords(vars, coords)
		}

		hbFont.SetVarCoordsDesign(coords)
	}

	return hbFont, nil
}

func (f *fcFont) isHinted() bool {
	hinting, ok := f.Pattern.GetBool(fc.HINTING)
	if !ok {
		return true
	}
	return hinting != fc.False
}

func (f *fcFont) isTransformed() bool {
	mat, ok := f.Pattern.GetMatrix(fc.MATRIX)
	if !ok {
		return false
	}
	return mat != fc.Identity
}

// len(axes) == len(coords)
func parseVariations(variations string) (parsedVars []truetype.Variation) {
	varis := strings.Split(variations, ",")
	for _, varia := range varis {
		vari, err := harfbuzz.ParseVariation(varia)
		if err != nil {
			continue
		}
		parsedVars = append(parsedVars, vari)
	}
	return parsedVars
}

func (font *fcFont) GetHarfbuzzFont() *harfbuzz.Font {
	if font.hbFont != nil {
		return font.hbFont
	}
	var err error
	font.hbFont, err = font.createHBFont() // TODO: add proper error handling
	if err != nil {
		log.Println("creating HB font:", err)
	}
	return font.hbFont
}

// getGlyph gets the glyph index for a given Unicode character
// for `font`. If you only want to determine
// whether the font has the glyph, use pango_font_has_char().
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

	hbFont := font.GetHarfbuzzFont()
	if glyph, ok := hbFont.Face().NominalGlyph(wc); ok {
		return pango.Glyph(glyph)
	}

	return pango.AsUnknownGlyph(wc)
}

func (font *fcFont) GetMetrics(lang pango.Language) pango.FontMetrics {
	sampleStr := pango.SampleString(lang)

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
	font.metricsByLang = append(font.metricsByLang, fcMetricsInfo{})
	info := &font.metricsByLang[len(font.metricsByLang)-1]
	info.sampleStr = sampleStr

	context := pango.NewContext(fontmap)
	context.SetLanguage(lang)

	info.metrics = font.getFaceMetrics()

	// Compute derived metrics
	desc := font.Describe(true)

	layout := pango.NewLayout(context)
	layout.SetFontDescription(&desc)
	layout.SetText(sampleStr)
	var extents pango.Rectangle
	layout.GetExtents(nil, &extents)

	sampleStrWidth := pango.GlyphUnit(len([]rune(sampleStr)))
	info.metrics.ApproximateCharWidth = extents.Width / sampleStrWidth

	layout.SetText("0123456789")
	info.metrics.ApproximateDigitWidth = pango.GlyphUnit(maxGlyphWidth(layout))

	return info.metrics
}

// The code in this function is partly based on code from Xft,
// Copyright 2000 Keith Packard
func (font *fcFont) getFaceMetrics() pango.FontMetrics {
	hbFont := font.GetHarfbuzzFont()

	extents := hbFont.ExtentsForDirection(harfbuzz.LeftToRight)
	var metrics pango.FontMetrics
	if fcMatrix, haveTransform := font.Pattern.GetMatrix(fc.MATRIX); haveTransform {
		metrics.Descent = -pango.GlyphUnit(float32(extents.Descender) * fcMatrix.Yy)
		metrics.Ascent = pango.GlyphUnit(float32(extents.Ascender) * fcMatrix.Yy)
		metrics.Height = pango.GlyphUnit(float32(extents.Ascender-extents.Descender+extents.LineGap) * fcMatrix.Yy)
	} else {
		metrics.Descent = -pango.GlyphUnit(extents.Descender)
		metrics.Ascent = pango.GlyphUnit(extents.Ascender)
		metrics.Height = pango.GlyphUnit(extents.Ascender - extents.Descender + extents.LineGap)
	}

	metrics.UnderlineThickness = pango.Scale
	metrics.UnderlinePosition = -pango.Scale
	metrics.StrikethroughThickness = pango.Scale
	metrics.StrikethroughPosition = metrics.Ascent / 2

	if position, ok := hbFont.LineMetric(fonts.UnderlineThickness); ok {
		metrics.UnderlineThickness = pango.GlyphUnit(position)
	}

	if position, ok := hbFont.LineMetric(fonts.UnderlinePosition); ok {
		metrics.UnderlinePosition = pango.GlyphUnit(position)
	}

	if position, ok := hbFont.LineMetric(fonts.StrikethroughThickness); ok {
		metrics.StrikethroughThickness = pango.GlyphUnit(position)
	}

	if position, ok := hbFont.LineMetric(fonts.StrikethroughPosition); ok {
		metrics.StrikethroughPosition = pango.GlyphUnit(position)
	}

	return metrics
}

func maxGlyphWidth(layout *pango.Layout) int32 {
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
	return int32(maxWidth)
}

// Gets the extents of a single glyph from a font. The extents are in
// user space; that is, they are not transformed by any matrix in effect
// for the font.
func (font *fcFont) getRawExtents(glyph pango.Glyph) (inkRect, logicalRect pango.Rectangle) {
	if glyph == pango.GLYPH_EMPTY {
		return pango.Rectangle{}, pango.Rectangle{}
	}

	hbFont := font.GetHarfbuzzFont()

	extents, _ := hbFont.GlyphExtents(glyph.GID())
	font_extents := hbFont.ExtentsForDirection(harfbuzz.LeftToRight)

	inkRect.X = pango.GlyphUnit(extents.XBearing)
	inkRect.Width = pango.GlyphUnit(extents.Width)
	inkRect.Y = pango.GlyphUnit(-extents.YBearing)
	inkRect.Height = pango.GlyphUnit(-extents.Height)

	x, _ := hbFont.GlyphAdvanceForDirection(glyph.GID(), harfbuzz.LeftToRight)

	logicalRect.X = 0
	logicalRect.Width = pango.GlyphUnit(x)
	logicalRect.Y = -pango.GlyphUnit(font_extents.Ascender)
	logicalRect.Height = pango.GlyphUnit(font_extents.Ascender - font_extents.Descender)

	return
}

//  func load_fallback_face (PangoFT2Font *ft2font,
// 			 const char   *original_file) {
//    PangoFcFont *fcfont = PANGO_FONT (ft2font);
//    Pattern *sans;
//    Pattern *matched;
//    Result result;
//    FT_Error error;
//    FcChar8 *filename2 = nil;
//    gchar *name;
//    int id;

//    sans = PatternBuild (nil,
// 			  FAMILY,     FcTypeString, "sans",
// 			  PIXEL_SIZE, FcTypeDouble, (double)ft2font.size / pango.Scale,
// 			  nil);

//    _pango_ft2_font_map_default_substitute ((PangoFcFontMap *)fcfont.fontmap, sans);

//    matched = FcFontMatch (pango_font_map_get_config ((PangoFcFontMap *)fcfont.fontmap), sans, &result);

//    if (PatternGetString (matched, FILE, 0, &filename2) != ResultMatch)
// 	 goto bail1;

//    if (PatternGetInteger (matched, INDEX, 0, &id) != ResultMatch)
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

//    PatternDestroy (sans);
//    PatternDestroy (matched);
//  }

// func set_transform (PangoFT2Font *ft2font) {
//    PangoFcFont *fcfont = (PangoFcFont *)ft2font;
//    FcMatrix *fcMatrix;

//    if (PatternGetMatrix (fcfont.font_pattern, MATRIX, 0, &fcMatrix) == ResultMatch)
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
//   * Use pango_font_lock_face() instead; when you are done with a
//   * face from pango_font_lock_face() you must call
//   * pango_font_unlock_face().
//   *
//   * Return value: (nullable): a pointer to a `FT_Face` structure, with the
//   *   size set correctly, or %nil if @font is %nil.
//   **/
// func  pango_ft2_font_get_face (PangoFont *font)  FT_Face  {
//    PangoFT2Font *ft2font = (PangoFT2Font *)font;
//    PangoFcFont *fcfont = (PangoFcFont *)font;
//    FT_Error error;
//    Pattern *pattern;
//    FcChar8 *filename;
//    Bool antialias, hinting, autohint;
//    int hintstyle;
//    int id;

//    if (G_UNLIKELY (!font))
// 	 return nil;

//    pattern = fcfont.font_pattern;

//    if (!ft2font.face)
// 	 {
// 	   ft2font.load_flags = 0;

// 	   /* disable antialiasing if requested */
// 	   if (PatternGetBool (pattern,
// 				 ANTIALIAS, 0, &antialias) != ResultMatch)
// 	 antialias = FcTrue;

// 	   if (antialias)
// 	 ft2font.load_flags |= FT_LOAD_NO_BITMAP;
// 	   else
// 	 ft2font.load_flags |= FT_LOAD_TARGET_MONO;

// 	   /* disable hinting if requested */
// 	   if (PatternGetBool (pattern,
// 				 HINTING, 0, &hinting) != ResultMatch)
// 	 hinting = FcTrue;

//  #ifdef HINT_STYLE
// 	   if (PatternGetInteger (pattern, HINT_STYLE, 0, &hintstyle) != ResultMatch)
// 	 hintstyle = HINT_FULL;

// 	   if (!hinting || hintstyle == HINT_NONE)
// 		   ft2font.load_flags |= FT_LOAD_NO_HINTING;

// 	   switch (hintstyle) {
// 	   case HINT_SLIGHT:
// 	   case HINT_MEDIUM:
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
// 	   if (PatternGetBool (pattern,
// 				 AUTOHINT, 0, &autohint) != ResultMatch)
// 	 autohint = FcFalse;

// 	   if (autohint)
// 	 ft2font.load_flags |= FT_LOAD_FORCE_AUTOHINT;

// 	   if (PatternGetString (pattern, FILE, 0, &filename) != ResultMatch)
// 		   goto bail0;

// 	   if (PatternGetInteger (pattern, INDEX, 0, &id) != ResultMatch)
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
//    PangoFcFontClass *font_class = PANGO_FONT_CLASS (class);

//    object_class.finalize = pango_ft2_font_finalize;

//    font_class.get_glyph_extents = pango_ft2_font_get_glyph_extents;

//    font_class.lock_face = pango_ft2_font_real_lock_face;
//    font_class.unlock_face = pango_ft2_font_real_unlock_face;
//  }

//  /**
//   * pango_ft2_font_get_kerning:
//   * @font: a #PangoFont
//   * @left: the left #PangoGlyph
//   * @right: the right #PangoGlyph
//   *
//   * Retrieves kerning information for a combination of two glyphs.
//   *
//   * Use pango_font_kern_glyphs() instead.
//   *
//   * Return value: The amount of kerning (in Pango units) to apply for
//   * the given combination of glyphs.
//   **/
//  int
//  pango_ft2_font_get_kerning (PangoFont *font,
// 				 PangoGlyph left,
// 				 PangoGlyph right)
//  {
//    PangoFcFont *font = PANGO_FONT (font);

//    FT_Face face;
//    FT_Error error;
//    FT_Vector kerning;

//    face = pango_font_lock_face (font);
//    if (!face)
// 	 return 0;

//    if (!FT_HAS_KERNING (face))
// 	 {
// 	   pango_font_unlock_face (font);
// 	   return 0;
// 	 }

//    error = FT_Get_Kerning (face, left, right, ft_kerning_default, &kerning);
//    if (error != FT_Err_Ok)
// 	 {
// 	   pango_font_unlock_face (font);
// 	   return 0;
// 	 }

//    pango_font_unlock_face (font);
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
//   * @font, or %GLYPH_EMPTY if no suitable glyph found.
//   *
//   * If you want to draw an unknown-box for a character that is not covered
//   * by the font,
//   * use AsUnknownGlyph() instead.
//   *
//   * Return value: a glyph index into @font, or %GLYPH_EMPTY
//   **/
//  PangoGlyph
//  pango_ft2_get_unknown_glyph (PangoFont *font)
//  {
//    FT_Face face = pango_ft2_font_get_face (font);
//    if (face && FT_IS_SFNT (face))
// 	 /* TrueType fonts have an 'unknown glyph' box on glyph index 0 */
// 	 return 0;
//    else
// 	 return GLYPH_EMPTY;
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

//  #define PANGO_TYPE_FAMILY            (pango_family_get_type ())
//  #define PANGO_FAMILY(object)         (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_TYPE_FAMILY, PangoFcFamily))
//  #define PANGO_IS_FAMILY(object)      (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_TYPE_FAMILY))

//  #define PANGO_TYPE_FACE              (pango_face_get_type ())
//  #define PANGO_FACE(object)           (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_TYPE_FACE, PangoFcFace))
//  #define PANGO_IS_FACE(object)        (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_TYPE_FACE))

//  #define PANGO_TYPE_Fontset           (pango_Fontset_get_type ())
//  #define PANGO_Fontset(object)        (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_TYPE_Fontset, PangoFontset))
//  #define PANGO_IS_Fontset(object)     (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_TYPE_Fontset))

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

//  #define PANGO_FONT_LOCK_FACE(font)	(PANGO_FONT_GET_CLASS (font).lock_face (font))
//  #define PANGO_FONT_UNLOCK_FACE(font)	(PANGO_FONT_GET_CLASS (font).unlock_face (font))

//  G_DEFINE_ABSTRACT_TYPE_WITH_CODE (PangoFcFont, pango_font, PANGO_TYPE_FONT,
// 								   G_ADD_PRIVATE (PangoFcFont))

//  static void
//  pango_font_class_init (PangoFcFontClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoFontClass *font_class = PANGO_FONT_CLASS (class);

//    class.has_char = pango_font_real_has_char;
//    class.get_glyph = pango_font_real_get_glyph;
//    class.get_unknown_glyph = nil;

//    object_class.finalize = pango_font_finalize;
//    object_class.set_property = pango_font_set_property;
//    object_class.get_property = pango_font_get_property;
//    font_class.describe = pango_font_describe;
//    font_class.describe_absolute = pango_font_describe_absolute;
//    font_class.GetCoverage = GetCoverage;
//    font_class.GetMetrics = GetMetrics;
//    font_class.GetFontMap = pango_font_get_font_map;
//    font_class.GetFeatures = pango_font_get_features;
//    font_class.CreateHBFont = createHBFont;
//    font_class.GetFeatures = pango_font_get_features;

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
// 							 PANGO_TYPE_FONT_MAP,
// 							 G_PARAM_READWRITE |
// 							 G_PARAM_STATIC_STRINGS));
//  }

//  static void
//  pango_font_init (PangoFcFont *font)
//  {
//    font.priv = pango_font_get_instance_private (font);
//  }

//  static void
//  free_metrics_info (PangoFcMetricsInfo *info)
//  {
//    pango_font_metrics_unref (info.metrics);
//    g_slice_free (PangoFcMetricsInfo, info);
//  }

//  static void
//  pango_font_finalize (GObject *object)
//  {
//    PangoFcFont *fcfont = PANGO_FONT (object);
//    PangoFcFontPrivate *priv = fcfont.priv;
//    PangoFcFontMap *fontmap;

//    g_slist_foreach (fcfont.metrics_by_lang, (GFunc)free_metrics_info, nil);
//    g_slist_free (fcfont.metrics_by_lang);

//    fontmap = g_weak_ref_get ((GWeakRef *) &fcfont.fontmap);
//    if (fontmap)
// 	 {
// 	   _pango_font_map_remove (PANGO_FONT_MAP (fcfont.fontmap), fcfont);
// 	   g_weak_ref_clear ((GWeakRef *) &fcfont.fontmap);
// 	   g_object_unref (fontmap);
// 	 }

//    PatternDestroy (fcfont.font_pattern);
//    pango_font_description_free (fcfont.description);

//    if (priv.decoder)
// 	 _pango_font_set_decoder (fcfont, nil);

//    G_OBJECT_CLASS (pango_font_parent_class).finalize (object);
//  }

//  static gboolean
//  pattern_is_hinted (Pattern *pattern)
//  {
//    Bool hinting;

//    if (PatternGetBool (pattern,
// 			 HINTING, 0, &hinting) != ResultMatch)
// 	 hinting = FcTrue;

//    return hinting;
//  }

//  static gboolean
//  pattern_is_transformed (Pattern *pattern)
//  {
//    FcMatrix *fcMatrix;

//    if (PatternGetMatrix (pattern, MATRIX, 0, &fcMatrix) == ResultMatch)
// 	 {
// 	   return fcMatrix.xx != 1 || fcMatrix.xy != 0 ||
// 			  fcMatrix.yx != 0 || fcMatrix.yy != 1;
// 	 }
//    else
// 	 return false;
//  }

//  static void
//  pango_font_set_property (GObject       *object,
// 				 guint          prop_id,
// 				 const GValue  *value,
// 				 GParamSpec    *pspec)
//  {
//    PangoFcFont *fcfont = PANGO_FONT (object);

//    switch (prop_id)
// 	 {
// 	 case PROP_PATTERN:
// 	   {
// 	 Pattern *pattern = g_value_get_pointer (value);

// 	 g_return_if_fail (pattern != nil);
// 	 g_return_if_fail (fcfont.font_pattern == nil);

// 	 PatternReference (pattern);
// 	 fcfont.font_pattern = pattern;
// 	 fcfont.description = pango_font_description_from_pattern (pattern, true);
// 	 fcfont.is_hinted = pattern_is_hinted (pattern);
// 	 fcfont.is_transformed = pattern_is_transformed (pattern);
// 	   }
// 	   goto set_decoder;

// 	 case PROP_FONTMAP:
// 	   {
// 	 PangoFcFontMap *fcfontmap = PANGO_FONT_MAP (g_value_get_object (value));

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
// 	 _pango_font_set_decoder (fcfont,
// 				 pango_font_map_find_decoder  ((PangoFcFontMap *) fcfont.fontmap,
// 								  fcfont.font_pattern));
//  }

//  static void
//  pango_font_get_property (GObject       *object,
// 				 guint          prop_id,
// 				 GValue        *value,
// 				 GParamSpec    *pspec)
//  {
//    switch (prop_id)
// 	 {
// 	 case PROP_PATTERN:
// 	   {
// 	 PangoFcFont *fcfont = PANGO_FONT (object);
// 	 g_value_set_pointer (value, fcfont.font_pattern);
// 	   }
// 	   break;
// 	 case PROP_FONTMAP:
// 	   {
// 	 PangoFcFont *fcfont = PANGO_FONT (object);
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
//  pango_font_create_base_metrics_for_context (PangoFcFont   *fcfont,
// 							PangoContext  *context)
//  {
//    PangoFontMetrics *metrics;
//    metrics = pango_font_metrics_new ();

//    getFaceMetrics (fcfont, metrics);

//    return metrics;
//  }

//  static PangoFontMap *
//  pango_font_get_font_map (font *PangoFcFont)
//  {
//    PangoFcFont *fcfont = PANGO_FONT (font);

//    /* MT-unsafe.  Oh well...  The API is unsafe. */
//    return fcfont.fontmap;
//  }

//  static gboolean
//  pango_font_real_has_char (font *PangoFcFont,
// 				  gunichar     wc)
//  {
//    FcCharSet *charset;

//    if (PatternGetCharSet (font.font_pattern,
// 				CHARSET, 0, &charset) != ResultMatch)
// 	 return false;

//    return FcCharSetHasChar (charset, wc);
//  }

//  /**
//   * pango_font_lock_face: (skip)
//   * @font: a #PangoFcFont.
//   *
//   * Gets the FreeType `FT_Face` associated with a font,
//   * This face will be kept around until you call
//   * pango_font_unlock_face().
//   *
//   * Return value: the FreeType `FT_Face` associated with @font.
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use GetHarfbuzzFont() instead
//   **/
//  FT_Face
//  pango_font_lock_face (font *PangoFcFont)
//  {
//    g_return_val_if_fail (PANGO_IS_FONT (font), nil);

//    return PANGO_FONT_LOCK_FACE (font);
//  }

//  /**
//   * pango_font_unlock_face:
//   * @font: a #PangoFcFont.
//   *
//   * Releases a font previously obtained with
//   * pango_font_lock_face().
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use GetHarfbuzzFont() instead
//   **/
//  void
//  pango_font_unlock_face (font *PangoFcFont)
//  {
//    g_return_if_fail (PANGO_IS_FONT (font));

//    PANGO_FONT_UNLOCK_FACE (font);
//  }

//  /**
//   * pango_font_has_char:
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
//  pango_font_has_char (font *PangoFcFont,
// 			 gunichar     wc)
//  {
//    PangoFcFontPrivate *priv = font.priv;
//    FcCharSet *charset;

//    g_return_val_if_fail (PANGO_IS_FONT (font), false);

//    if (priv.decoder)
// 	 {
// 	   charset = pango_decoder_get_charset (priv.decoder, font);
// 	   return FcCharSetHasChar (charset, wc);
// 	 }

//    return PANGO_FONT_GET_CLASS (font).has_char (font, wc);
//  }

//  /**
//   * pango_font_get_unknown_glyph:
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
//  pango_font_get_unknown_glyph (font *PangoFcFont,
// 				  gunichar     wc)
//  {
//    if (font && PANGO_FONT_GET_CLASS (font).get_unknown_glyph)
// 	 return PANGO_FONT_GET_CLASS (font).get_unknown_glyph (font, wc);

//    return AsUnknownGlyph (wc);
//  }

//  void
//  _pango_font_shutdown (font *PangoFcFont)
//  {
//    g_return_if_fail (PANGO_IS_FONT (font));

//    if (PANGO_FONT_GET_CLASS (font).shutdown)
// 	 PANGO_FONT_GET_CLASS (font).shutdown (font);
//  }

//  /**
//   * pango_font_kern_glyphs:
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
//  pango_font_kern_glyphs (PangoFcFont      *font,
// 				PangoGlyphString *glyphs)
//  {
//  }

//  /**
//   * _pango_font_get_decoder:
//   * @font: a #PangoFcFont
//   *
//   * This will return any custom decoder set on this font.
//   *
//   * Return value: The custom decoder
//   *
//   * Since: 1.6
//   **/

//  PangoFcDecoder *
//  _pango_font_get_decoder (font *PangoFcFont)
//  {
//    PangoFcFontPrivate *priv = font.priv;

//    return priv.decoder;
//  }

//  /**
//   * _pango_font_set_decoder:
//   * @font: a #PangoFcFont
//   * @decoder: a #PangoFcDecoder to set for this font
//   *
//   * This sets a custom decoder for this font.  Any previous decoder
//   * will be released before this one is set.
//   *
//   * Since: 1.6
//   **/

//  void
//  _pango_font_set_decoder (PangoFcFont    *font,
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
//  _pango_font_get_font_key (PangoFcFont *fcfont)
//  {
//    PangoFcFontPrivate *priv = fcfont.priv;

//    return priv.key;
//  }

//  void
//  _pango_font_set_font_key (fcfont *PangoFcFont,
// 				  PangoFcFontKey *key)
//  {
//    PangoFcFontPrivate *priv = fcfont.priv;

//    priv.key = key;
//  }

func (font *fcFont) GetFeatures() []harfbuzz.Feature {
	/* Setup features from fontconfig pattern. */
	features := font.Pattern.GetStrings(fc.FONT_FEATURES)
	var out []harfbuzz.Feature
	for _, feature := range features {
		feat, err := harfbuzz.ParseFeature(feature)
		if err != nil {
			continue
		}
		feat.Start = 0
		feat.End = harfbuzz.FeatureGlobalEnd
		out = append(out, feat)
	}
	return out
}

//  extern gpointer get_gravity_class (void);

//  /**
//   * pango_font_get_languages:
//   * @font: a #PangoFcFont
//   *
//   * Returns the languages that are supported by @font.
//   *
//   * This corresponds to the LANG member of the Pattern.
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
//  pango_font_get_languages (font *PangoFcFont)
//  {
//    PangoFcFontMap *fontmap;
//    PangoLanguage **languages;

//    fontmap = g_weak_ref_get ((GWeakRef *) &font.fontmap);
//    if (!fontmap)
// 	 return nil;

//    languages  = _pango_font_map_get_languages (fontmap, font);
//    g_object_unref (fontmap);

//    return languages;
//  }

//  /**
//   * pango_font_get_pattern: (skip)
//   * @font: a #PangoFcFont
//   *
//   * Returns the Pattern that @font is based on.
//   *
//   * Returns: the fontconfig pattern for this font
//   *
//   * Since: 1.48
//   */
//  Pattern *
//  pango_font_get_pattern (font *PangoFcFont)
//  {
//    return font.font_pattern;
//  }
