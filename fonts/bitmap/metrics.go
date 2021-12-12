package bitmap

import "github.com/benoitkugler/textlayout/fonts"

var _ fonts.FaceMetrics = (*Font)(nil)

func (Font) Upem() uint16 { return 1000 }

func (f *Font) GlyphName(gid fonts.GID) string {
	if int(gid) >= len(f.names) {
		return ""
	}
	return f.names[gid]
}

func (Font) LineMetric(fonts.LineMetric) (float32, bool) {
	return 0, false
}

func (f *Font) FontHExtents() (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}

func (f *Font) FontVExtents() (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}

func (f *Font) NominalGlyph(r rune) (fonts.GID, bool) {
	return f.cmap.Lookup(r)
}

func (f *Font) HorizontalAdvance(gid fonts.GID) float32 {
	if int(gid) >= len(f.metrics) {
		return 0
	}
	return float32(f.metrics[gid].characterWidth)
}

// adapted from freetype ft_synthesize_vertical_metrics
func (m *metric) synthesizeVerticalMetrics(accelerator *acceleratorTable) (bearingX, bearingY, vAdvance int32) {
	vAdvance = accelerator.fontAscent + accelerator.fontDescent
	height := m.characterAscent + m.characterDescent

	if vAdvance == 0 {
		horiBearingY := m.characterAscent
		/* compensate for glyph with bbox above/below the baseline */
		if horiBearingY < 0 {
			if height < horiBearingY {
				height = horiBearingY
			}
		} else if horiBearingY > 0 {
			height -= horiBearingY
		}

		/* the factor 1.2 is a heuristical value */
		vAdvance = int32(height) * 12 / 10
	}

	bearingX = int32(m.leftSideBearing - m.characterWidth/2)
	bearingY = (vAdvance - int32(height)) / 2
	return
}

func (f *Font) VerticalAdvance(gid fonts.GID) float32 {
	if int(gid) >= len(f.metrics) {
		return 0
	}

	_, _, advance := f.metrics[gid].synthesizeVerticalMetrics(f.accelerator)
	return float32(advance)
}

// GlyphHOrigin fetches the (X,Y) coordinates of the origin (in font units) for a glyph ID,
// for horizontal text segments.
// Returns `false` if not available.
func (f *Font) GlyphHOrigin(fonts.GID) (x, y int32, found bool) {
	return 0, 0, true
}

// GlyphVOrigin is the same as `GlyphHOrigin`, but for vertical text segments.
func (f *Font) GlyphVOrigin(gid fonts.GID) (x, y int32, found bool) {
	if int(gid) >= len(f.metrics) {
		return 0, 0, false
	}

	// adapted from harfbuzz
	m := f.metrics[gid]
	vertBearingX, vertBearingY, _ := m.synthesizeVerticalMetrics(f.accelerator)

	/* Note: FreeType's vertical metrics grows downward while other FreeType coordinates
	 * have a Y growing upward.  Hence the extra negation. */
	x = int32(m.leftSideBearing) - vertBearingX
	y = int32(m.characterAscent) - (-vertBearingY)
	return x, y, true
}

func (m metric) extents() (ext fonts.GlyphExtents) {
	ext.XBearing = float32(m.leftSideBearing)
	ext.YBearing = float32(m.characterAscent)
	ext.Width = float32(m.rightSideBearing - m.leftSideBearing)
	ext.Height = -(float32(m.characterAscent - m.characterDescent))
	return ext
}

// GlyphExtents retrieve the extents for a specified glyph, of false, if not available.
// `coords` is used by variable fonts, and is specified in normalized coordinates.
// For bitmap glyphs, the closest resolution to `xPpem` and `yPpem` is selected.
func (f *Font) GlyphExtents(gid fonts.GID, _, _ uint16) (fonts.GlyphExtents, bool) {
	if int(gid) >= len(f.metrics) {
		return fonts.GlyphExtents{}, false
	}

	// adapted from harfbuzz
	return f.metrics[gid].extents(), true
}
