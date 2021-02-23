package truetype

import (
	"math"

	"github.com/benoitkugler/textlayout/fonts"
)

var _ fonts.FontMetrics = (*fontMetrics)(nil)

type fontMetrics struct {
	cmap       Cmap
	hvar, vvar *tableHVvar // optionnel
	hhea       *TableHVhea
	vhea       *TableHVhea
	vorg       *tableVorg // optionnel
	mvar       TableMvar
	gvar       tableGvar
	fvar       TableFvar
	cmapVar    unicodeVariations
	glyphs     TableGlyf
	vmtx, hmtx tableHVmtx
	avar       tableAvar
	head       TableHead
	os2        TableOS2
	upem       uint16
}

func (font *Font) LoadMetrics() fonts.FontMetrics {
	var out fontMetrics

	out.head = font.Head
	if out.head.UnitsPerEm < 16 || out.head.UnitsPerEm > 16384 {
		out.upem = 1000
	} else {
		out.upem = out.head.UnitsPerEm
	}

	if os2, err := font.OS2Table(); err == nil {
		out.os2 = *os2
	}

	out.glyphs, _ = font.glyfTable()
	out.hhea, _ = font.HheaTable()
	out.vhea, _ = font.VheaTable()
	out.hmtx, _ = font.HtmxTable()

	if font.Fvar != nil {
		out.fvar = *font.Fvar
		out.mvar, _ = font.mvarTable()
		out.gvar, _ = font.gvarTable(out.glyphs)
		if v, err := font.hvarTable(); err == nil {
			out.hvar = &v
		}
		if v, err := font.vvarTable(); err == nil {
			out.vvar = &v
		}
		out.avar, _ = font.avarTable()
	}

	out.cmap, _ = font.Cmap.BestEncoding()
	out.cmapVar = font.Cmap.unicodeVariation

	if vorg, err := font.vorgTable(); err != nil {
		out.vorg = &vorg
	}

	return &out
}

func (f *fontMetrics) GetUpem() uint16 { return f.upem }

var (
	metricsTagHorizontalAscender  = MustNewTag("hasc")
	metricsTagHorizontalDescender = MustNewTag("hdsc")
	metricsTagHorizontalLineGap   = MustNewTag("hlgp")
	// metricsTagHorizontalClippingAscent  = MustNewTag("hcla")
	// metricsTagHorizontalClippingDescent = MustNewTag("hcld")
	metricsTagVerticalAscender  = MustNewTag("vasc")
	metricsTagVerticalDescender = MustNewTag("vdsc")
	metricsTagVerticalLineGap   = MustNewTag("vlgp")
	// metricsTagHorizontalCaretRise       = MustNewTag("hcrs")
	// metricsTagHorizontalCaretRun        = MustNewTag("hcrn")
	// metricsTagHorizontalCaretOffset     = MustNewTag("hcof")
	// metricsTagVerticalCaretRise         = MustNewTag("vcrs")
	// metricsTagVerticalCaretRun          = MustNewTag("vcrn")
	// metricsTagVerticalCaretOffset       = MustNewTag("vcof")
	// metricsTagXHeight                   = MustNewTag("xhgt")
	// metricsTagCapHeight                 = MustNewTag("cpht")
	// metricsTagSubscriptEmXSize          = MustNewTag("sbxs")
	// metricsTagSubscriptEmYSize          = MustNewTag("sbys")
	// metricsTagSubscriptEmXOffset        = MustNewTag("sbxo")
	// metricsTagSubscriptEmYOffset        = MustNewTag("sbyo")
	// metricsTagSuperscriptEmXSize        = MustNewTag("spxs")
	// metricsTagSuperscriptEmYSize        = MustNewTag("spys")
	// metricsTagSuperscriptEmXOffset      = MustNewTag("spxo")
	// metricsTagSuperscriptEmYOffset      = MustNewTag("spyo")
	// metricsTagStrikeoutSize             = MustNewTag("strs")
	// metricsTagStrikeoutOffset           = MustNewTag("stro")
	// metricsTagUnderlineSize             = MustNewTag("unds")
	// metricsTagUnderlineOffset           = MustNewTag("undo")
)

func fixAscenderDescender(value float32, metricsTag Tag) float32 {
	if metricsTag == metricsTagHorizontalAscender || metricsTag == metricsTagVerticalAscender {
		return float32(math.Abs(float64(value)))
	}
	if metricsTag == metricsTagHorizontalDescender || metricsTag == metricsTagVerticalDescender {
		return float32(-math.Abs(float64(value)))
	}
	return value
}

func (f *fontMetrics) getPositionCommon(metricTag Tag, coords []float32) (float32, bool) {
	deltaVar := f.mvar.getVar(metricTag, coords)
	switch metricTag {
	case metricsTagHorizontalAscender:
		if f.os2.useTypoMetrics() && f.os2.hasData() {
			return fixAscenderDescender(float32(f.os2.STypoAscender)+deltaVar, metricTag), true
		} else if f.hhea != nil {
			return fixAscenderDescender(float32(f.hhea.Ascent)+deltaVar, metricTag), true
		}

	case metricsTagHorizontalDescender:
		if f.os2.useTypoMetrics() && f.os2.hasData() {
			return fixAscenderDescender(float32(f.os2.STypoDescender)+deltaVar, metricTag), true
		} else if f.hhea != nil {
			return fixAscenderDescender(float32(f.hhea.Descent)+deltaVar, metricTag), true
		}
	case metricsTagHorizontalLineGap:
		if f.os2.useTypoMetrics() && f.os2.hasData() {
			return fixAscenderDescender(float32(f.os2.STypoLineGap)+deltaVar, metricTag), true
		} else if f.hhea != nil {
			return fixAscenderDescender(float32(f.hhea.LineGap)+deltaVar, metricTag), true
		}
	case metricsTagVerticalAscender:
		if f.vhea != nil {
			return fixAscenderDescender(float32(f.vhea.Ascent)+deltaVar, metricTag), true
		}
	case metricsTagVerticalDescender:
		if f.vhea != nil {
			return fixAscenderDescender(float32(f.vhea.Descent)+deltaVar, metricTag), true
		}
	case metricsTagVerticalLineGap:
		if f.vhea != nil {
			return fixAscenderDescender(float32(f.vhea.LineGap)+deltaVar, metricTag), true
		}
	}
	return 0, false
}

func (f *fontMetrics) GetFontHExtents(coords []float32) (fonts.FontExtents, bool) {
	var (
		out           fonts.FontExtents
		ok1, ok2, ok3 bool
	)
	out.Ascender, ok1 = f.getPositionCommon(metricsTagHorizontalAscender, coords)
	out.Descender, ok2 = f.getPositionCommon(metricsTagHorizontalDescender, coords)
	out.LineGap, ok3 = f.getPositionCommon(metricsTagHorizontalLineGap, coords)
	return out, ok1 && ok2 && ok3
}

func (f *fontMetrics) GetNominalGlyph(ch rune) (fonts.GlyphIndex, bool) {
	gid := f.cmap.Lookup(ch)
	return gid, gid != 0
}

func (f *fontMetrics) GetVariationGlyph(ch, varSelector rune) (fonts.GlyphIndex, bool) {
	gid, kind := f.cmapVar.getGlyphVariant(ch, varSelector)
	switch kind {
	case variantNotFound:
		return 0, false
	case variantFound:
		return gid, true
	default: // variantUseDefault
		return f.GetNominalGlyph(ch)
	}
}

// do not take into account variations
func (f *fontMetrics) getBaseAdvance(gid fonts.GlyphIndex, table tableHVmtx) int16 {
	if int(gid) >= len(table) {
		/* If `table` is empty, it means we don't have the metrics table
		 * for this direction: return default advance.  Otherwise, it means that the
		 * glyph index is out of bound: return zero. */
		if len(table) == 0 {
			return int16(f.upem)
		}
		return 0
	}
	return table[gid].Advance
}

const (
	phantomLeft = iota
	phantomRight
	phantomTop
	phantomBottom
	phantomCount
)

// for composite, recursively calls itself; allPoints includes phantom points and will be at least of length 4
func (f *fontMetrics) getPointsForGlyph(gid fonts.GlyphIndex, coords []float32, phantomOnly bool,
	depth int, allPoints *[]contourPoint /* OUT */) {
	// adapted from harfbuzz/src/hb-ot-glyf-table.hh

	if depth > maxCompositeNesting || int(gid) >= len(f.glyphs) {
		return
	}
	g := f.glyphs[gid]

	var points []contourPoint
	if data, ok := g.data.(simpleGlyphData); !phantomOnly && ok {
		points = data.getContourPoints() // fetch the "real" points
	} else { // zeros values are enough
		points = make([]contourPoint, g.pointNumbersCount())
	}

	// init phantom point
	points = append(points, make([]contourPoint, phantomCount)...)
	phantoms := points[len(points)-phantomCount:]

	hDelta := float32(g.Xmin - f.hmtx[gid].SideBearing)
	vOrig := float32(g.Ymax + f.vmtx[gid].SideBearing)
	hAdv := float32(f.getBaseAdvance(gid, f.hmtx))
	vAdv := float32(f.getBaseAdvance(gid, f.vmtx))
	phantoms[phantomLeft].x = hDelta
	phantoms[phantomRight].x = hAdv + hDelta
	phantoms[phantomTop].y = vOrig
	phantoms[phantomBottom].y = vOrig - vAdv

	f.gvar.applyDeltasToPoints(gid, coords, points)

	switch data := g.data.(type) {
	case simpleGlyphData:
		*allPoints = append(*allPoints, points...)
	case compositeGlyphData:
		for compIndex, item := range data.glyphs {
			// recurse on component
			var compPoints []contourPoint

			f.getPointsForGlyph(item.glyphIndex, coords, phantomOnly, depth+1, &compPoints)

			LC := len(compPoints)
			if LC < phantomCount { // in case of max depth reached
				return
			}

			/* Copy phantom points from component if USE_MY_METRICS flag set */
			if item.hasUseMyMetrics() {
				for i := range phantoms {
					phantoms[i] = compPoints[LC-phantomCount+i]
				}
			}

			/* Apply component transformation & translation */
			item.transformPoints(compPoints)

			/* Apply translation from gvar */
			tx, ty := points[compIndex].x, points[compIndex].y
			for i := range compPoints {
				compPoints[i].translate(tx, ty)
			}

			if item.isAnchored() {
				p1, p2 := int(item.arg1), int(item.arg2)
				if p1 < len(*allPoints) && p2 < LC {
					tx, ty := (*allPoints)[p1].x-compPoints[p2].x, (*allPoints)[p1].y-compPoints[p2].y
					for i := range compPoints {
						compPoints[i].translate(tx, ty)
					}
				}
			}

			*allPoints = append(*allPoints, compPoints[0:LC-phantomCount]...)

		}

		*allPoints = append(*allPoints, phantoms...)
	default:
		*allPoints = append(*allPoints, phantoms...)
	}

	// apply at top level
	if depth == 0 {
		/* Undocumented rasterizer behavior:
		 * Shift points horizontally by the updated left side bearing */
		tx := -phantoms[phantomLeft].x
		for i := range *allPoints {
			(*allPoints)[i].translate(tx, 0)
		}
	}
}

// walk through the contour points of the given glyph to compute its extends and its phantom points
// As an optimization, if `computeExtents` is false, the extents computation is skipped (a zero value is returned).
func (f *fontMetrics) getPoints(gid fonts.GlyphIndex, coords []float32, computeExtents bool) (ext fonts.GlyphExtents, ph [phantomCount]contourPoint) {
	if int(gid) >= len(f.glyphs) {
		return
	}
	var allPoints []contourPoint
	f.getPointsForGlyph(gid, coords, !computeExtents, 0, &allPoints)

	copy(ph[:], allPoints[len(allPoints)-phantomCount:])

	if computeExtents {
		truePoints := allPoints[:len(allPoints)-phantomCount]
		var minX, minY, maxX, maxY float32
		for _, p := range truePoints {
			minX = minF(minX, p.x)
			minY = minF(minY, p.y)
			maxX = maxF(maxX, p.x)
			maxY = maxF(maxY, p.y)
		}
		ext.XBearing = minX
		ext.Width = maxX - minX
		ext.YBearing = maxY
		ext.Height = minY - maxY
	}

	return ext, ph
}

func roundAndClamp(v float32) int16 {
	out := int16(v)
	if out < 0 {
		out = 0
	}
	return out
}

func (f *fontMetrics) getGlyphAdvanceVar(gid fonts.GlyphIndex, coords []float32, isVertical bool) int16 {
	_, phantoms := f.getPoints(gid, coords, false)
	if isVertical {
		return roundAndClamp(phantoms[phantomTop].y - phantoms[phantomBottom].y)
	}
	return roundAndClamp(phantoms[phantomRight].x - phantoms[phantomLeft].x)
}

func (f *fontMetrics) GetHorizontalAdvance(gid fonts.GlyphIndex, coords []float32) int16 {
	advance := f.getBaseAdvance(gid, f.hmtx)
	if len(coords) == 0 || len(coords) != len(f.fvar.Axis) {
		return advance
	}
	if f.hvar != nil {
		return advance + int16(f.hvar.getAdvanceVar(gid, coords))
	}
	return f.getGlyphAdvanceVar(gid, coords, false)
}

// return `true` is the font is variable and `coords` is valid
func (f *fontMetrics) isVar(coords []float32) bool {
	return len(coords) != 0 && len(coords) == len(f.fvar.Axis)
}

func (f *fontMetrics) GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) int16 {
	// return the opposite of the advance from the font

	advance := f.getBaseAdvance(gid, f.vmtx)
	if !f.isVar(coords) {
		return -advance
	}
	if f.vvar != nil {
		return -advance - int16(f.vvar.getAdvanceVar(gid, coords))
	}
	return -f.getGlyphAdvanceVar(gid, coords, true)
}

func (f *fontMetrics) GetGlyphHOrigin(fonts.GlyphIndex, []float32) (x, y int32, found bool) {
	return 0, 0, true
}

func (f *fontMetrics) GetGlyphVOrigin(glyph fonts.GlyphIndex, coords []float32) (x, y int32, found bool) {
	x = int32(f.GetHorizontalAdvance(glyph, coords) / 2)

	if f.vorg != nil {
		y = int32(f.vorg.getYOrigin(glyph))
		return x, y, true
	}

	// extents, ok := ot_face.glyf.get_extents(font, glyph, &extents)
	// if ok {
	// 	//   const OT::vmtx_accelerator_t &vmtx = *ot_face.vmtx;
	// 	tsb := vmtx.get_side_bearing(font, glyph)
	// 	*y = extents.y_bearing + font.em_scale_y(tsb)
	// 	return true
	// }

	// font_extents := font.get_h_extents_with_fallback(&font_extents)
	// *y = font_extents.ascender

	return
}

func (f *fontMetrics) getExtentsFromGlyf(glyph fonts.GlyphIndex, coords []float32) (fonts.GlyphExtents, bool) {
	if int(glyph) >= len(f.glyphs) {
		return fonts.GlyphExtents{}, false
	}
	g := f.glyphs[glyph]
	if g.data == nil {
		return fonts.GlyphExtents{}, false
	}
	if f.isVar(coords) {
		extents, _ := f.getPoints(glyph, coords, true)
		return extents, true
	}
	return g.getExtents(f.hmtx, glyph), true
}

// func (f *fontMetrics) getExtentsFromSbix(glyph fonts.GlyphIndex, coords []float32) (fonts.GlyphExtents, bool) {
// }

// func (f *fontMetrics) getExtentsFromCff1(glyph fonts.GlyphIndex, coords []float32) (fonts.GlyphExtents, bool) {
// }

// func (f *fontMetrics) getExtentsFromCff2(glyph fonts.GlyphIndex, coords []float32) (fonts.GlyphExtents, bool) {
// }

// func (f *fontMetrics) getExtentsFromCBDT(glyph fonts.GlyphIndex, coords []float32) (fonts.GlyphExtents, bool) {
// }

func (f *fontMetrics) GetGlyphExtents(glyph fonts.GlyphIndex, xPpem, yPpem uint16) (fonts.GlyphExtents, bool) {
	return fonts.GlyphExtents{}, false
}

// Normalizes the given design-space coordinates. The minimum and maximum
// values for the axis are mapped to the interval [-1,1], with the default
// axis value mapped to 0.
// Any additional scaling defined in the face's `avar` table is also
// applied, as described at https://docs.microsoft.com/en-us/typography/opentype/spec/avar
func (f *fontMetrics) NormalizeVariations(coords []float32) []float32 {
	// ported from freetype2

	normalized := make([]float32, len(coords))
	// Axis normalization is a two-stage process.  First we normalize
	// based on the [min,def,max] values for the axis to be [-1,0,1].
	// Then, if there's an `avar' table, we renormalize this range.

	for i, a := range f.fvar.Axis {
		coord := coords[i]

		if coord > a.Maximum || coord < a.Minimum { // out of range: clamping
			if coord > a.Maximum {
				coord = a.Maximum
			} else {
				coord = a.Minimum
			}
		}

		if coord < a.Default {
			normalized[i] = -(coord - a.Default) / (a.Minimum - a.Default)
		} else if coord > a.Default {
			normalized[i] = (coord - a.Default) / (a.Maximum - a.Default)
		} else {
			normalized[i] = 0
		}
	}

	// now applying 'avar'
	for i, av := range f.avar {
		for j := 1; j < len(av); j++ {
			previous, pair := av[j-1], av[j]
			if normalized[i] < pair.from {
				normalized[i] =
					previous.to + (normalized[i]-previous.from)*
						(pair.to-previous.to)/(pair.from-previous.from)
				break
			}
		}
	}

	return normalized
}
