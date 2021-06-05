package harfbuzz

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	tt "github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/graphite"
)

// ported from src/hb-font.hh, src/hb-font.cc  Copyright © 2009  Red Hat, Inc., 2012  Google, Inc.  Behdad Esfahbod

// Face is the interface providing font metrics and layout information.
// Harfbuzz is mostly useful when used with fonts providing advanced layout capabilities :
// see the extension interface `FaceOpentype`.
type Face = fonts.FaceLayout

var (
	_ FaceOpentype        = (*truetype.Font)(nil)
	_ FaceMetricsOpentype = (*truetype.FontMetrics)(nil)
)

// FaceMetricsOpentype provides additional metrics exposed by
// Opentype fonts.
type FaceMetricsOpentype interface {
	// GetGlyphContourPoint retrieves the (X,Y) coordinates (in font units) for a
	// specified contour point in a glyph, or false if not found.
	GetGlyphContourPoint(glyph fonts.GID, pointIndex uint16) (x, y Position, ok bool)
}

// FaceOpentype adds support for advanced layout features
// found in Opentype/Truetype font files.
// LoadMetrics should return an oject implementing `FaceMetricsOpentype`.
// See the package fonts/truetype for more details.
type FaceOpentype interface {
	Face

	// Returns true if the font has Graphite capabilities.
	// Note that tables validity will still be checked in `NewFont`,
	// using the table from the returned `truetype.Font`.
	// Overide this method to disable Graphite functionalities.
	IsGraphite() (bool, *truetype.Font)

	// LayoutTables fetchs the Opentype layout tables of the font.
	LayoutTables() truetype.LayoutTables
}

// Font is used internally as a light wrapper around the provided Face.
//
// While a font face is generally the in-memory representation of a static font file,
// `Font` handles dynamic attributes like size, width and
// other parameters (pixels-per-em, points-per-em, variation
// settings).
//
// Font are constructed with `NewFont` and adjusted by accessing the fields
// XPpem, YPpem, Ptem,XScale, YScale and with the method `SetVarCoordsDesign` for
// variable fonts.
type Font struct {
	origin Face
	face   fonts.FaceMetrics

	// only non nil for valid graphite fonts
	gr *graphite.GraphiteFace

	// opentype fields, initialized from a FaceOpentype
	otTables               *truetype.LayoutTables
	coords                 []float32                   // font variation coordinates (optional), normalized
	gsubAccels, gposAccels []otLayoutLookupAccelerator // accelators for lookup
	faceUpem               int32                       // cached value of Face.Upem()

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
// In particular, when appropriate, it will load the additional information
// required for Opentype and Graphite layout, which will influence
// the shaping plan used in `Buffer.Shape`.
// The `face` object should not be modified after this call.
func NewFont(face Face) *Font {
	var font Font

	font.origin = face
	font.face = face.LoadMetrics()
	font.faceUpem = Position(font.face.Upem())
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

		if is, tables := opentypeFace.IsGraphite(); is {
			font.gr, _ = graphite.LoadGraphite(tables)
		}
	}

	return &font
}

// Applies a list of font-variation settings to a font.
func (f *Font) setVariations(variations []tt.Variation) {
	if len(variations) == 0 {
		f.coords = nil
		return
	}

	varFont, isVar := f.face.(truetype.VariableFont)
	if !isVar {
		f.coords = nil
		return
	}
	fvar := varFont.Variations()

	designCoords := fvar.GetDesignCoordsDefault(variations)

	f.SetVarCoordsDesign(designCoords)
}

// Face returns the underlying face.
// Note that field is readonly, since some caching may happen
// in the `NewFont` constructor.
func (f *Font) Face() fonts.FaceMetrics { return f.face }

// SetVarCoordsDesign applies a list of variation coordinates, in design-space units,
// to the font.
func (f *Font) SetVarCoordsDesign(coords []float32) {
	f.coords = f.face.NormalizeVariations(coords)
}

// ---- Convert from font-space to user-space ----

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

// GlyphExtents is the same as fonts.GlyphExtents but with int type
type GlyphExtents struct {
	XBearing int32
	YBearing int32
	Width    int32
	Height   int32
}

// GlyphExtents fetches the GlyphExtents data for a glyph ID
// in the specified font, or false if not found
func (f *Font) GlyphExtents(glyph fonts.GID) (out GlyphExtents, ok bool) {
	ext, ok := f.face.GlyphExtents(glyph, f.coords, f.XPpem, f.YPpem)
	if !ok {
		return out, false
	}
	out.XBearing = f.emScalefX(ext.XBearing)
	out.Width = f.emScalefX(ext.Width)
	out.YBearing = f.emScalefY(ext.YBearing)
	out.Height = f.emScalefY(ext.Height)
	return out, true
}

// GlyphAdvanceForDirection fetches the advance for a glyph ID from the specified font,
// in a text segment of the specified direction.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of `dir`.
func (f *Font) GlyphAdvanceForDirection(glyph fonts.GID, dir Direction) (x, y Position) {
	if dir.isHorizontal() {
		return f.GlyphHAdvance(glyph), 0
	}
	return 0, f.getGlyphVAdvance(glyph)
}

// GlyphHAdvance fetches the advance for a glyph ID in the font,
// for horizontal text segments.
func (f *Font) GlyphHAdvance(glyph fonts.GID) Position {
	adv := f.face.HorizontalAdvance(glyph, f.coords)
	return f.emScalefX(adv)
}

// Fetches the advance for a glyph ID in the font,
// for vertical text segments.
func (f *Font) getGlyphVAdvance(glyph fonts.GID) Position {
	adv := f.face.VerticalAdvance(glyph, f.coords)
	return f.emScalefY(adv)
}

// Subtracts the origin coordinates from an (X,Y) point coordinate,
// in the specified glyph ID in the specified font.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f *Font) subtractGlyphOriginForDirection(glyph fonts.GID, direction Direction,
	x, y Position) (Position, Position) {
	originX, originY := f.getGlyphOriginForDirection(glyph, direction)

	return x - originX, y - originY
}

// Fetches the (X,Y) coordinates of the origin for a glyph in
// the specified font.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of @direction.
func (f *Font) getGlyphOriginForDirection(glyph fonts.GID, direction Direction) (x, y Position) {
	if direction.isHorizontal() {
		return f.getGlyphHOriginWithFallback(glyph)
	}
	return f.getGlyphVOriginWithFallback(glyph)
}

func (f *Font) getGlyphHOriginWithFallback(glyph fonts.GID) (Position, Position) {
	x, y, ok := f.face.GlyphHOrigin(glyph, f.coords)
	if !ok {
		x, y, ok = f.face.GlyphVOrigin(glyph, f.coords)
		if ok {
			dx, dy := f.guessVOriginMinusHOrigin(glyph)
			return x - dx, y - dy
		}
	}
	return x, y
}

func (f *Font) getGlyphVOriginWithFallback(glyph fonts.GID) (Position, Position) {
	x, y, ok := f.face.GlyphVOrigin(glyph, f.coords)
	if !ok {
		x, y, ok = f.face.GlyphHOrigin(glyph, f.coords)
		if ok {
			dx, dy := f.guessVOriginMinusHOrigin(glyph)
			return x + dx, y + dy
		}
	}
	return x, y
}

func (f *Font) guessVOriginMinusHOrigin(glyph fonts.GID) (x, y Position) {
	x = f.GlyphHAdvance(glyph) / 2
	y = f.getHExtendsAscender()
	return x, y
}

func (f *Font) getHExtendsAscender() Position {
	extents, ok := f.face.FontHExtents(f.coords)
	if !ok {
		return f.YScale * 4 / 5
	}
	return f.emScalefY(extents.Ascender)
}

func (f *Font) hasGlyph(ch rune) bool {
	_, ok := f.face.NominalGlyph(ch)
	return ok
}

func (f *Font) subtractGlyphHOrigin(glyph fonts.GID, x, y Position) (Position, Position) {
	originX, originY := f.getGlyphHOriginWithFallback(glyph)
	return x - originX, y - originY
}

func (f *Font) subtractGlyphVOrigin(glyph fonts.GID, x, y Position) (Position, Position) {
	originX, originY := f.getGlyphVOriginWithFallback(glyph)
	return x - originX, y - originY
}

func (f *Font) addGlyphHOrigin(glyph fonts.GID, x, y Position) (Position, Position) {
	originX, originY := f.getGlyphHOriginWithFallback(glyph)
	return x + originX, y + originY
}

func (f *Font) getGlyphContourPointForOrigin(glyph fonts.GID, pointIndex uint16, direction Direction) (x, y Position, ok bool) {
	met, ok := f.face.(FaceMetricsOpentype)
	if !ok {
		return
	}

	x, y, ok = met.GetGlyphContourPoint(glyph, pointIndex)
	if ok {
		x, y = f.subtractGlyphOriginForDirection(glyph, direction, x, y)
	}

	return x, y, ok
}

// Generates gidDDD if glyph has no name.
func (f *Font) glyphToString(glyph fonts.GID) string {
	if name := f.face.GlyphName(glyph); name != "" {
		return name
	}

	return fmt.Sprintf("gid%d", glyph)
}

// ExtentsForDirection fetches the extents for a font in a text segment of the
// specified direction.
//
// Calls the appropriate direction-specific variant (horizontal
// or vertical) depending on the value of `direction`.
func (f *Font) ExtentsForDirection(direction Direction) fonts.FontExtents {
	var (
		extents fonts.FontExtents
		ok      bool
	)
	if direction.isHorizontal() {
		extents, ok = f.face.FontHExtents(f.coords)
		if !ok {
			extents.Ascender = float32(f.YScale) * 0.8
			extents.Descender = extents.Ascender - float32(f.YScale)
			extents.LineGap = 0
		}
	} else {
		extents, ok = f.face.FontVExtents(f.coords)
		if !ok {
			extents.Ascender = float32(f.XScale) * 0.5
			extents.Descender = extents.Ascender - float32(f.XScale)
			extents.LineGap = 0
		}
	}
	return extents
}

// LineMetric fetches the given metric, applying potential variations
// and scaling.
func (f *Font) LineMetric(metric fonts.LineMetric) (int32, bool) {
	m, ok := f.face.LineMetric(metric, f.coords)
	return f.emScalefY(m), ok
}
