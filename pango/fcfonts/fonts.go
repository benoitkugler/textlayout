// Package fcfonts is an implementation of
// the font tooling required by Pango, using textlayout/fontconfig
// and textlayout/fonts.
//
// The entry point of the package is the `NewFontMap` constructor.
package fcfonts

import (
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

		desc.SetSize(int32(scale_factor * size * float32(pango.Scale)))
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
	decoder  decoder
	key      *fcFontKey
	hbFont   *harfbuzz.Font // cached result of loadHBFont
	coverage pango.Coverage // cached result of loadCoverage

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
		desc.SetAbsoluteSize(int32(size * float32(pango.Scale)))
	}

	return desc
}

func (font *fcFont) GetCoverage(_ pango.Language) pango.Coverage {
	if font.decoder != nil {
		charset := font.decoder.GetCharset(font)
		return fromCharset(charset)
	}

	if font.coverage == nil {
		// Pull the coverage out of the pattern, this doesn't require loading the font
		charset, _ := font.Pattern.GetCharset(fc.CHARSET)
		font.coverage = fromCharset(charset) // stores it into the map
	}

	return font.coverage
}

func (font *fcFont) GetFontMap() pango.FontMap { return font.fontmap }

// create a new font, which will be cached
func (font *fcFont) loadHBFont() error {
	var (
		xScaleInv, yScaleInv float32 = 1.0, 1.
		size                 float32 = 1.0
	)

	key := font.key
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
		return err
	}

	font.hbFont = harfbuzz.NewFont(hb_face)
	font.hbFont.XScale, font.hbFont.YScale = int32(size*pango.Scale*xScale), int32(size*pango.Scale*yScale)
	if varFont, isVariable := hb_face.(truetype.FaceVariable); key != nil && isVariable {
		fvar := varFont.Variations()
		if len(fvar.Axis) == 0 {
			return nil
		}

		coords := fvar.GetDesignCoordsDefault(nil)

		if index, ok := key.pattern.GetInt(fc.INDEX); ok && index != 0 {
			if instance := (index >> 16) - 1; instance >= 0 && int(instance) < len(fvar.Instances) {
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

		font.hbFont.SetVarCoordsDesign(coords)
	}

	return nil
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

func (font *fcFont) GetHarfbuzzFont() *harfbuzz.Font { return font.hbFont }

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

	if font.decoder != nil {
		return font.decoder.GetGlyph(font, wc)
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
