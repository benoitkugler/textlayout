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

func (f *Font) LoadMetrics() fonts.FontMetrics {
	var out fontMetrics

	out.head = f.Head
	if out.head.UnitsPerEm < 16 || out.head.UnitsPerEm > 16384 {
		out.upem = 1000
	} else {
		out.upem = out.head.UnitsPerEm
	}

	if os2, err := f.OS2Table(); err == nil {
		out.os2 = *os2
	}

	out.glyphs, _ = f.glyfTable()
	out.hhea, _ = f.HheaTable()
	out.vhea, _ = f.VheaTable()
	out.hmtx, _ = f.HtmxTable()

	if f.Fvar != nil {
		out.fvar = *f.Fvar
		out.mvar, _ = f.mvarTable()
		out.gvar, _ = f.gvarTable(out.glyphs)
		if v, err := f.hvarTable(); err == nil {
			out.hvar = &v
		}
		if v, err := f.vvarTable(); err == nil {
			out.vvar = &v
		}
		out.avar, _ = f.avarTable()
	}

	out.cmap, _ = f.Cmap.BestEncoding()
	out.cmapVar = f.Cmap.unicodeVariation

	return &out
}

func (f *fontMetrics) GetUpem() uint16 { return f.upem }

var (
	HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER         = MustNewTag("hasc")
	HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER        = MustNewTag("hdsc")
	HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP         = MustNewTag("hlgp")
	HB_OT_METRICS_TAG_HORIZONTAL_CLIPPING_ASCENT  = MustNewTag("hcla")
	HB_OT_METRICS_TAG_HORIZONTAL_CLIPPING_DESCENT = MustNewTag("hcld")
	HB_OT_METRICS_TAG_VERTICAL_ASCENDER           = MustNewTag("vasc")
	HB_OT_METRICS_TAG_VERTICAL_DESCENDER          = MustNewTag("vdsc")
	HB_OT_METRICS_TAG_VERTICAL_LINE_GAP           = MustNewTag("vlgp")
	HB_OT_METRICS_TAG_HORIZONTAL_CARET_RISE       = MustNewTag("hcrs")
	HB_OT_METRICS_TAG_HORIZONTAL_CARET_RUN        = MustNewTag("hcrn")
	HB_OT_METRICS_TAG_HORIZONTAL_CARET_OFFSET     = MustNewTag("hcof")
	HB_OT_METRICS_TAG_VERTICAL_CARET_RISE         = MustNewTag("vcrs")
	HB_OT_METRICS_TAG_VERTICAL_CARET_RUN          = MustNewTag("vcrn")
	HB_OT_METRICS_TAG_VERTICAL_CARET_OFFSET       = MustNewTag("vcof")
	HB_OT_METRICS_TAG_X_HEIGHT                    = MustNewTag("xhgt")
	HB_OT_METRICS_TAG_CAP_HEIGHT                  = MustNewTag("cpht")
	HB_OT_METRICS_TAG_SUBSCRIPT_EM_X_SIZE         = MustNewTag("sbxs")
	HB_OT_METRICS_TAG_SUBSCRIPT_EM_Y_SIZE         = MustNewTag("sbys")
	HB_OT_METRICS_TAG_SUBSCRIPT_EM_X_OFFSET       = MustNewTag("sbxo")
	HB_OT_METRICS_TAG_SUBSCRIPT_EM_Y_OFFSET       = MustNewTag("sbyo")
	HB_OT_METRICS_TAG_SUPERSCRIPT_EM_X_SIZE       = MustNewTag("spxs")
	HB_OT_METRICS_TAG_SUPERSCRIPT_EM_Y_SIZE       = MustNewTag("spys")
	HB_OT_METRICS_TAG_SUPERSCRIPT_EM_X_OFFSET     = MustNewTag("spxo")
	HB_OT_METRICS_TAG_SUPERSCRIPT_EM_Y_OFFSET     = MustNewTag("spyo")
	HB_OT_METRICS_TAG_STRIKEOUT_SIZE              = MustNewTag("strs")
	HB_OT_METRICS_TAG_STRIKEOUT_OFFSET            = MustNewTag("stro")
	HB_OT_METRICS_TAG_UNDERLINE_SIZE              = MustNewTag("unds")
	HB_OT_METRICS_TAG_UNDERLINE_OFFSET            = MustNewTag("undo")
)

func fixAscenderDescender(value float32, metricsTag Tag) float32 {
	if metricsTag == HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER || metricsTag == HB_OT_METRICS_TAG_VERTICAL_ASCENDER {
		return float32(math.Abs(float64(value)))
	}
	if metricsTag == HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER || metricsTag == HB_OT_METRICS_TAG_VERTICAL_DESCENDER {
		return float32(-math.Abs(float64(value)))
	}
	return value
}

func (f *fontMetrics) getPositionCommon(metricTag Tag, coords []float32) (float32, bool) {
	deltaVar := f.mvar.getVar(metricTag, coords)
	switch metricTag {
	case HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER:
		if f.os2.useTypoMetrics() && f.os2.hasData() {
			return fixAscenderDescender(float32(f.os2.STypoAscender)+deltaVar, metricTag), true
		} else if f.hhea != nil {
			return fixAscenderDescender(float32(f.hhea.Ascent)+deltaVar, metricTag), true
		}

	case HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER:
		if f.os2.useTypoMetrics() && f.os2.hasData() {
			return fixAscenderDescender(float32(f.os2.STypoDescender)+deltaVar, metricTag), true
		} else if f.hhea != nil {
			return fixAscenderDescender(float32(f.hhea.Descent)+deltaVar, metricTag), true
		}
	case HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP:
		if f.os2.useTypoMetrics() && f.os2.hasData() {
			return fixAscenderDescender(float32(f.os2.STypoLineGap)+deltaVar, metricTag), true
		} else if f.hhea != nil {
			return fixAscenderDescender(float32(f.hhea.LineGap)+deltaVar, metricTag), true
		}
	case HB_OT_METRICS_TAG_VERTICAL_ASCENDER:
		if f.vhea != nil {
			return fixAscenderDescender(float32(f.vhea.Ascent)+deltaVar, metricTag), true
		}
	case HB_OT_METRICS_TAG_VERTICAL_DESCENDER:
		if f.vhea != nil {
			return fixAscenderDescender(float32(f.vhea.Descent)+deltaVar, metricTag), true
		}
	case HB_OT_METRICS_TAG_VERTICAL_LINE_GAP:
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
	out.Ascender, ok1 = f.getPositionCommon(HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER, coords)
	out.Descender, ok2 = f.getPositionCommon(HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER, coords)
	out.LineGap, ok3 = f.getPositionCommon(HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP, coords)
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
		/* If num_metrics is zero, it means we don't have the metrics table
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
func (f *fontMetrics) getPoints(gid fonts.GlyphIndex, coords []float32, depth int, allPoints *[]contourPoint /* OUT */) {
	// adapted from harfbuzz/src/hb-ot-glyf-table.hh

	if depth > maxCompositeNesting || int(gid) >= len(f.glyphs) {
		return
	}
	g := f.glyphs[gid]
	points := make([]contourPoint, g.pointNumbersCount())
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

			f.getPoints(item.glyphIndex, coords, depth+1, &compPoints)

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

	if depth == 0 /* Apply at top level */ {
		/* Undocumented rasterizer behavior:
		 * Shift points horizontally by the updated left side bearing
		 */
		// contour_point_t delta;
		tx := -phantoms[phantomLeft].x
		for i := range *allPoints {
			(*allPoints)[i].translate(tx, 0)
		}
	}
}

func roundAndClamp(v float32) int16 {
	out := int16(v)
	if out < 0 {
		out = 0
	}
	return out
}

func (f *fontMetrics) getGlyphAdvanceVar(gid fonts.GlyphIndex, coords []float32, isVertical bool) int16 {
	var allPoints []contourPoint
	f.getPoints(gid, coords, 0, &allPoints)
	phantoms := allPoints[len(allPoints)-4:]
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

func (f *fontMetrics) GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) int16 {
	// return the opposite of the advance from the font

	advance := f.getBaseAdvance(gid, f.vmtx)
	if len(coords) == 0 || len(coords) != len(f.fvar.Axis) {
		return -advance
	}
	if f.vvar != nil {
		return -advance - int16(f.vvar.getAdvanceVar(gid, coords))
	}
	return -f.getGlyphAdvanceVar(gid, coords, true)
}

func (f *fontMetrics) GetGlyphHOrigin(fonts.GlyphIndex) (x, y int32, found bool) { return }

func (f *fontMetrics) GetGlyphVOrigin(fonts.GlyphIndex) (x, y int32, found bool) { return }

func (f *fontMetrics) GetGlyphExtents(fonts.GlyphIndex) (fonts.GlyphExtents, bool) {
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
