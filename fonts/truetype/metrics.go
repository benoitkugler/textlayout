package truetype

import (
	"math"

	"github.com/benoitkugler/textlayout/fonts"
)

var _ fonts.FontMetrics = (*fontMetrics)(nil)

type fontMetrics struct {
	upem uint16

	head       TableHead
	os2        TableOS2
	mvar       TableMvar
	hhea, vhea *TableHVhea
	hmtx, vmtx tableHVmtx
	hvar, vvar *tableHVvar // optionnel
	cmap       Cmap
	cmapVar    unicodeVariations
}

func (f *Font) LoadMetrics() fonts.FontMetrics {
	var out fontMetrics

	if head, err := f.HeadTable(); err != nil {
		out.head = *head
	}
	if out.head.UnitsPerEm < 16 || out.head.UnitsPerEm > 16384 {
		out.upem = 1000
	} else {
		out.upem = out.head.UnitsPerEm
	}

	if os2, err := f.OS2Table(); err != nil {
		out.os2 = *os2
	}

	out.mvar, _ = f.mvarTable()
	out.hhea, _ = f.HheaTable()
	out.vhea, _ = f.VheaTable()
	out.hmtx, _ = f.HtmxTable()

	if v, err := f.hvarTable(); err != nil {
		out.hvar = &v
	}
	if v, err := f.vvarTable(); err != nil {
		out.vvar = &v
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
func (f *fontMetrics) getBaseHorizontalAdvance(gid fonts.GlyphIndex) int16 {
	if int(gid) >= len(f.hmtx) {
		/* If num_metrics is zero, it means we don't have the metrics table
		 * for this direction: return default advance.  Otherwise, it means that the
		 * glyph index is out of bound: return zero. */
		if len(f.hmtx) == 0 {
			return int16(f.upem)
		}
		return 0
	}
	return f.hmtx[gid].Advance
}

func (f *fontMetrics) GetHorizontalAdvance(gid fonts.GlyphIndex, coords []float32) int16 {
	advance := f.getBaseHorizontalAdvance(gid)
	if len(coords) == 0 {
		return advance
	}
	if f.hvar != nil {
		return advance + int16(f.hvar.getAdvanceVar(gid, coords))
	}
	// TODO:
	return 0
}

func (f *fontMetrics) GetVerticalAdvance(gid fonts.GlyphIndex, coords []float32) int16 { return 0 }

func (f *fontMetrics) GetGlyphHOrigin(fonts.GlyphIndex) (x, y int32, found bool) { return }

func (f *fontMetrics) GetGlyphVOrigin(fonts.GlyphIndex) (x, y int32, found bool) { return }

func (f *fontMetrics) GetGlyphExtents(fonts.GlyphIndex) (fonts.GlyphExtents, bool) {
	return fonts.GlyphExtents{}, false
}

func (f *fontMetrics) NormalizeVariations(coords []float32) []float32 { return nil }
