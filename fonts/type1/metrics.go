package type1

import (
	"math"

	"github.com/benoitkugler/textlayout/fonts"
)

// font metrics

var _ fonts.FaceMetrics = (*Font)(nil)

// Upem reads the FontMatrix to extract the scaling factor (the maximum between x and y coordinates)
func (f *Font) Upem() uint16 {
	if len(f.FontMatrix) < 4 {
		return 1000 // typical value for Type1 fonts
	}
	xx, yy := math.Abs(float64(f.FontMatrix[0])), math.Abs(float64(f.FontMatrix[3]))
	var (
		upemX uint16 = 1000
		upemY        = upemX
	)
	if xx != 0 {
		upemX = uint16(math.Round(1 / xx))
	}
	if yy != 0 {
		upemY = uint16(math.Round(1 / yy))
	}
	if upemX > upemY {
		return upemX
	}
	return upemY
}

func (f *Font) GlyphName(gid fonts.GID) string {
	if int(gid) >= len(f.charstrings) {
		return ""
	}
	return f.charstrings[gid].name
}

func (f *Font) LineMetric(metric fonts.LineMetric) (float32, bool) {
	switch metric {
	case fonts.UnderlinePosition:
		return float32(f.PSInfo.UnderlinePosition), true
	case fonts.UnderlineThickness:
		return float32(f.PSInfo.UnderlineThickness), true
	default:
		// CapHeight and XHeight are stored in .afm files
		return 0, false
	}
}

func (f *Font) FontHExtents() (fonts.FontExtents, bool) {
	var extents fonts.FontExtents
	if len(f.FontBBox) < 4 {
		return extents, false
	}
	yMin, yMax := f.FontBBox[1], f.FontBBox[3]
	// following freetype here
	extents.Ascender = float32(yMax)
	extents.Descender = float32(yMin)

	extents.LineGap = float32(f.Upem()) * 1.2
	if extents.LineGap < extents.Ascender-extents.Descender {
		extents.LineGap = extents.Ascender - extents.Descender
	}
	return extents, true
}

// FontVExtents returns zero values.
func (f *Font) FontVExtents() (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}

func (f *Font) Cmap() (fonts.Cmap, fonts.CmapEncoding) {
	return f.cmap, fonts.EncUnicode
}

func (f *Font) NominalGlyph(ch rune) (fonts.GID, bool) {
	out, ok := f.cmap[ch]
	return out, ok
}

// HorizontalAdvance returns the advance of the glyph with index `index`
// The return value is expressed in font units.
// 0 is returned for invalid index values and for invalid
// charstring glyph data.
func (f *Font) HorizontalAdvance(gid fonts.GID) float32 {
	_, _, adv, err := f.loadGlyph(gid, false)
	if err != nil {
		return 0
	}
	return float32(adv)
}

func (f *Font) VerticalAdvance(gid fonts.GID) float32 { return 0 }

// GlyphHOrigin always return 0,0,true
func (Font) GlyphHOrigin(fonts.GID) (x, y int32, found bool) {
	return 0, 0, true
}

// GlyphVOrigin always return 0,0,false
func (Font) GlyphVOrigin(fonts.GID) (x, y int32, found bool) {
	return 0, 0, false
}

func (f *Font) GlyphExtents(glyph fonts.GID, _, _ uint16) (fonts.GlyphExtents, bool) {
	_, bbox, _, err := f.loadGlyph(glyph, false)
	if err != nil {
		return fonts.GlyphExtents{}, false
	}
	return bbox.ToExtents(), true
}

func (Font) NormalizeVariations(coords []float32) []float32 { return coords }

var _ fonts.FaceRenderer = (*Font)(nil)

// GlyphData returns the outlines of the given glyph.
// The returned value is either a fonts.GlyphOutline or nil if an error
// occured.
func (f *Font) GlyphData(gid fonts.GID, _, _ uint16) fonts.GlyphData {
	segments, _, _, err := f.loadGlyph(gid, false)
	if err != nil {
		return nil
	}
	return fonts.GlyphOutline{Segments: segments}
}
