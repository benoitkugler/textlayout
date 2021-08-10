package truetype

import (
	"math"

	"github.com/benoitkugler/textlayout/fonts"
	type1c "github.com/benoitkugler/textlayout/fonts/type1C"
)

var _ fonts.FaceMetrics = (*metrics)(nil)

// metrics provides support for the 'fonts.FaceMetrics' interface
// by querying various open type tables.
type metrics struct {
	cmap        Cmap
	hvar, vvar  *tableHVvar // optionnel
	hhea, vhea  *TableHVhea
	vorg        *tableVorg // optionnel
	cff         *type1c.Font
	post        PostTable // optionnel
	mvar        TableMvar
	gvar        tableGvar
	fvar        TableFvar
	glyphs      TableGlyf
	colorBitmap bitmapTable // TODO: support for gray ?
	cmapVar     unicodeVariations
	vmtx, hmtx  TableHVmtx
	sbix        tableSbix

	head TableHead
	os2  TableOS2

	upem uint16
}

func (font *Font) loadMetrics() {
	font.metrics.head = font.Head
	if font.metrics.head.UnitsPerEm < 16 || font.metrics.head.UnitsPerEm > 16384 {
		font.metrics.upem = 1000
	} else {
		font.metrics.upem = font.metrics.head.UnitsPerEm
	}

	if os2, err := font.OS2Table(); err == nil {
		font.metrics.os2 = *os2
	}

	font.metrics.glyphs, _ = font.GlyfTable()
	font.metrics.colorBitmap, _ = font.colorBitmapTable()
	font.metrics.sbix, _ = font.sbixTable()
	font.metrics.cff, _ = font.cffTable()
	font.metrics.post, _ = font.PostTable()

	font.metrics.hhea, _ = font.HheaTable()
	font.metrics.vhea, _ = font.VheaTable()
	font.metrics.hmtx, _ = font.HtmxTable()
	font.metrics.vmtx, _ = font.VtmxTable()

	if len(font.fvar.Axis) != 0 {
		font.metrics.mvar, _ = font.mvarTable()
		font.metrics.gvar, _ = font.gvarTable(font.metrics.glyphs)
		if v, err := font.hvarTable(); err == nil {
			font.metrics.hvar = &v
		}
		if v, err := font.vvarTable(); err == nil {
			font.metrics.vvar = &v
		}
	}

	font.metrics.cmap, _ = font.cmaps.BestEncoding()
	font.metrics.cmapVar = font.cmaps.unicodeVariation

	if vorg, err := font.vorgTable(); err == nil {
		font.metrics.vorg = &vorg
	}
}

// Returns true if the font has Graphite capabilities,
// but does not check if the tables are actually valid.
func (font *Font) IsGraphite() (*Font, bool) {
	return font, font.HasTable(TagSilf)
}

func (f *metrics) GetGlyphContourPoint(glyph fonts.GID, pointIndex uint16) (x, y int32, ok bool) {
	// harfbuzz seems not to implement this feature
	return 0, 0, false
}

func (f *metrics) GlyphName(glyph GID) string {
	if postNames := f.post.Names; postNames != nil {
		if name := postNames.GlyphName(glyph); name != "" {
			return name
		}
	}
	if f.cff != nil {
		return f.cff.GlyphName(glyph)
	}
	return ""
}

func (f *metrics) Upem() uint16 { return f.upem }

var (
	metricsTagHorizontalAscender  = MustNewTag("hasc")
	metricsTagHorizontalDescender = MustNewTag("hdsc")
	metricsTagHorizontalLineGap   = MustNewTag("hlgp")
	metricsTagVerticalAscender    = MustNewTag("vasc")
	metricsTagVerticalDescender   = MustNewTag("vdsc")
	metricsTagVerticalLineGap     = MustNewTag("vlgp")
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

func (f *metrics) getPositionCommon(metricTag Tag, coords []float32) (float32, bool) {
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

func (f *metrics) FontHExtents(coords []float32) (fonts.FontExtents, bool) {
	var (
		out           fonts.FontExtents
		ok1, ok2, ok3 bool
	)
	out.Ascender, ok1 = f.getPositionCommon(metricsTagHorizontalAscender, coords)
	out.Descender, ok2 = f.getPositionCommon(metricsTagHorizontalDescender, coords)
	out.LineGap, ok3 = f.getPositionCommon(metricsTagHorizontalLineGap, coords)
	return out, ok1 && ok2 && ok3
}

func (f *metrics) FontVExtents(coords []float32) (fonts.FontExtents, bool) {
	var (
		out           fonts.FontExtents
		ok1, ok2, ok3 bool
	)
	out.Ascender, ok1 = f.getPositionCommon(metricsTagVerticalAscender, coords)
	out.Descender, ok2 = f.getPositionCommon(metricsTagVerticalDescender, coords)
	out.LineGap, ok3 = f.getPositionCommon(metricsTagVerticalLineGap, coords)
	return out, ok1 && ok2 && ok3
}

var (
	tagStrikeoutSize   = MustNewTag("strs")
	tagStrikeoutOffset = MustNewTag("stro")
	tagUnderlineSize   = MustNewTag("unds")
	tagUnderlineOffset = MustNewTag("undo")
)

func (f *metrics) LineMetric(metric fonts.LineMetric, coords []float32) (float32, bool) {
	switch metric {
	case fonts.UnderlinePosition:
		return float32(f.post.UnderlinePosition) + f.mvar.getVar(tagUnderlineOffset, coords), true
	case fonts.UnderlineThickness:
		return float32(f.post.UnderlineThickness) + f.mvar.getVar(tagUnderlineSize, coords), true
	case fonts.StrikethroughPosition:
		return float32(f.os2.YStrikeoutPosition) + f.mvar.getVar(tagStrikeoutOffset, coords), true
	case fonts.StrikethroughThickness:
		return float32(f.os2.YStrikeoutSize) + f.mvar.getVar(tagStrikeoutSize, coords), true
	}
	return 0, false
}

func (f *metrics) NominalGlyph(ch rune) (GID, bool) {
	return f.cmap.Lookup(ch)
}

func (f *metrics) VariationGlyph(ch, varSelector rune) (GID, bool) {
	gid, kind := f.cmapVar.getGlyphVariant(ch, varSelector)
	switch kind {
	case variantNotFound:
		return 0, false
	case variantFound:
		return gid, true
	default: // variantUseDefault
		return f.NominalGlyph(ch)
	}
}

// do not take into account variations
func (f *metrics) getBaseAdvance(gid GID, table TableHVmtx) int16 {
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
func (f *metrics) getPointsForGlyph(gid GID, coords []float32, depth int, allPoints *[]contourPoint /* OUT */) {
	// adapted from harfbuzz/src/hb-ot-glyf-table.hh

	if depth > maxCompositeNesting || int(gid) >= len(f.glyphs) {
		return
	}
	g := f.glyphs[gid]

	var points []contourPoint
	if data, ok := g.data.(simpleGlyphData); ok {
		points = data.getContourPoints() // fetch the "real" points
	} else { // zeros values are enough
		points = make([]contourPoint, g.pointNumbersCount())
	}

	// init phantom point
	points = append(points, make([]contourPoint, phantomCount)...)
	phantoms := points[len(points)-phantomCount:]

	hDelta := float32(g.Xmin - f.hmtx.getSideBearing(gid))
	vOrig := float32(g.Ymax + f.vmtx.getSideBearing(gid))
	hAdv := float32(f.getBaseAdvance(gid, f.hmtx))
	vAdv := float32(f.getBaseAdvance(gid, f.vmtx))
	phantoms[phantomLeft].x = hDelta
	phantoms[phantomRight].x = hAdv + hDelta
	phantoms[phantomTop].y = vOrig
	phantoms[phantomBottom].y = vOrig - vAdv

	if f.isVar(coords) {
		f.gvar.applyDeltasToPoints(gid, coords, points)
	}

	switch data := g.data.(type) {
	case simpleGlyphData:
		*allPoints = append(*allPoints, points...)
	case compositeGlyphData:
		for compIndex, item := range data.glyphs {
			// recurse on component
			var compPoints []contourPoint

			f.getPointsForGlyph(item.glyphIndex, coords, depth+1, &compPoints)

			LC := len(compPoints)
			if LC < phantomCount { // in case of max depth reached
				return
			}

			/* Copy phantom points from component if USE_MY_METRICS flag set */
			if item.hasUseMyMetrics() {
				copy(phantoms, compPoints[LC-phantomCount:])
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
	default: // no data for the glyph
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

func extentsFromPoints(allPoints []contourPoint) (ext fonts.GlyphExtents) {
	truePoints := allPoints[:len(allPoints)-phantomCount]
	if len(truePoints) == 0 {
		// zero extent for the empty glyph
		return ext
	}
	minX, minY := truePoints[0].x, truePoints[0].y
	maxX, maxY := minX, minY
	for _, p := range truePoints {
		minX = minF(minX, p.x)
		minY = minF(minY, p.y)
		maxX = maxF(maxX, p.x)
		maxY = maxF(maxY, p.y)
	}
	ext.XBearing = minX
	ext.YBearing = maxY
	ext.Width = maxX - minX
	ext.Height = minY - maxY
	return ext
}

// walk through the contour points of the given glyph to compute its extends and its phantom points
// As an optimization, if `computeExtents` is false, the extents computation is skipped (a zero value is returned).
func (f *metrics) getPoints(gid GID, coords []float32, computeExtents bool) (ext fonts.GlyphExtents, ph [phantomCount]contourPoint) {
	if int(gid) >= len(f.glyphs) {
		return
	}
	var allPoints []contourPoint
	f.getPointsForGlyph(gid, coords, 0, &allPoints)

	copy(ph[:], allPoints[len(allPoints)-phantomCount:])

	if computeExtents {
		ext = extentsFromPoints(allPoints)
	}

	return ext, ph
}

func clamp(v float32) float32 {
	if v < 0 {
		v = 0
	}
	return v
}

func ceil(v float32) int16 {
	return int16(math.Ceil(float64(v)))
}

func (f *metrics) getGlyphAdvanceVar(gid GID, coords []float32, isVertical bool) float32 {
	_, phantoms := f.getPoints(gid, coords, false)
	if isVertical {
		return clamp(phantoms[phantomTop].y - phantoms[phantomBottom].y)
	}
	return clamp(phantoms[phantomRight].x - phantoms[phantomLeft].x)
}

func (f *metrics) HorizontalAdvance(gid GID, coords []float32) float32 {
	advance := f.getBaseAdvance(gid, f.hmtx)
	if !f.isVar(coords) {
		return float32(advance)
	}
	if f.hvar != nil {
		return float32(advance) + f.hvar.getAdvanceVar(gid, coords)
	}
	return f.getGlyphAdvanceVar(gid, coords, false)
}

// return `true` is the font is variable and `coords` is valid
func (f *metrics) isVar(coords []float32) bool {
	return len(coords) != 0 && len(coords) == len(f.fvar.Axis)
}

func (f *metrics) VerticalAdvance(gid GID, coords []float32) float32 {
	// return the opposite of the advance from the font
	advance := f.getBaseAdvance(gid, f.vmtx)
	if !f.isVar(coords) {
		return -float32(advance)
	}
	if f.vvar != nil {
		return -float32(advance) - f.vvar.getAdvanceVar(gid, coords)
	}
	return -f.getGlyphAdvanceVar(gid, coords, true)
}

func (f *metrics) getGlyphSideBearingVar(gid GID, coords []float32, isVertical bool) int16 {
	extents, phantoms := f.getPoints(gid, coords, true)
	if isVertical {
		return ceil(phantoms[phantomTop].y - extents.YBearing)
	}
	return int16(phantoms[phantomLeft].x)
}

// take variations into account
func (f *metrics) getHorizontalSideBearing(glyph GID, coords []float32) int16 {
	// base side bearing
	sideBearing := f.hmtx.getSideBearing(glyph)
	if !f.isVar(coords) {
		return sideBearing
	}
	if f.hvar != nil {
		return sideBearing + int16(f.hvar.getSideBearingVar(glyph, coords))
	}
	return f.getGlyphSideBearingVar(glyph, coords, false)
}

// take variations into account
func (f *metrics) getVerticalSideBearing(glyph GID, coords []float32) int16 {
	// base side bearing
	sideBearing := f.vmtx.getSideBearing(glyph)
	if !f.isVar(coords) {
		return sideBearing
	}
	if f.vvar != nil {
		return sideBearing + int16(f.vvar.getSideBearingVar(glyph, coords))
	}
	return f.getGlyphSideBearingVar(glyph, coords, true)
}

func (f *metrics) GlyphHOrigin(GID, []float32) (x, y int32, found bool) {
	// zero is the right value here
	return 0, 0, true
}

func (f *metrics) GlyphVOrigin(glyph GID, coords []float32) (x, y int32, found bool) {
	x = int32(f.HorizontalAdvance(glyph, coords) / 2)

	if f.vorg != nil {
		y = int32(f.vorg.getYOrigin(glyph))
		return x, y, true
	}

	if extents, ok := f.getExtentsFromGlyf(glyph, coords); ok {
		tsb := f.getVerticalSideBearing(glyph, coords)
		y = int32(extents.YBearing) + int32(tsb)
		return x, y, true
	}

	fontExtents, ok := f.FontHExtents(coords)
	y = int32(fontExtents.Ascender)

	return x, y, ok
}

func (f *metrics) getExtentsFromGlyf(glyph GID, coords []float32) (fonts.GlyphExtents, bool) {
	if int(glyph) >= len(f.glyphs) {
		return fonts.GlyphExtents{}, false
	}
	g := f.glyphs[glyph]
	if f.isVar(coords) {
		extents, _ := f.getPoints(glyph, coords, true)
		return extents, true
	}
	return g.getExtents(f.hmtx, glyph), true
}

func (f *metrics) getExtentsFromCBDT(glyph GID, xPpem, yPpem uint16) (fonts.GlyphExtents, bool) {
	strike := f.colorBitmap.chooseStrike(xPpem, yPpem)
	if strike == nil || strike.ppemX == 0 || strike.ppemY == 0 {
		return fonts.GlyphExtents{}, false
	}
	subtable := strike.findTable(glyph)
	if subtable == nil {
		return fonts.GlyphExtents{}, false
	}
	image := subtable.getImage(glyph)
	if image == nil {
		return fonts.GlyphExtents{}, false
	}
	extents := image.metrics.glyphExtents()

	/* convert to font units. */
	xScale := float32(f.upem) / float32(strike.ppemX)
	yScale := float32(f.upem) / float32(strike.ppemY)
	extents.XBearing *= xScale
	extents.YBearing *= yScale
	extents.Width *= xScale
	extents.Height *= yScale
	return extents, true
}

func (f *metrics) getExtentsFromSbix(glyph GID, coords []float32, xPpem, yPpem uint16) (fonts.GlyphExtents, bool) {
	strike := f.sbix.chooseStrike(xPpem, yPpem)
	if strike == nil || strike.ppem == 0 {
		return fonts.GlyphExtents{}, false
	}
	data := strike.getGlyph(glyph, 0)
	if data.isNil() {
		return fonts.GlyphExtents{}, false
	}
	extents, ok := data.glyphExtents()

	/* convert to font units. */
	scale := float32(f.upem) / float32(strike.ppem)
	extents.XBearing *= scale
	extents.YBearing *= scale
	extents.Width *= scale
	extents.Height *= scale
	return extents, ok
}

func (f *metrics) getExtentsFromCff1(glyph GID) (fonts.GlyphExtents, bool) {
	if f.cff == nil {
		return fonts.GlyphExtents{}, false
	}
	return f.cff.GetExtents(glyph)
}

// func (f *fontMetrics) getExtentsFromCff2(glyph , coords []float32) (fonts.GlyphExtents, bool) {
// }

func (f *metrics) GlyphExtents(glyph GID, coords []float32, xPpem, yPpem uint16) (fonts.GlyphExtents, bool) {
	out, ok := f.getExtentsFromSbix(glyph, coords, xPpem, yPpem)
	if ok {
		return out, ok
	}
	out, ok = f.getExtentsFromGlyf(glyph, coords)
	if ok {
		return out, ok
	}
	out, ok = f.getExtentsFromCff1(glyph)
	if ok {
		return out, ok
	}
	out, ok = f.getExtentsFromCBDT(glyph, xPpem, yPpem)
	return out, ok
}
