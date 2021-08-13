package pango

import (
	"fmt"
	"log"
	"unicode"

	"github.com/benoitkugler/textlayout/fonts"
)

// Glyph represents a single glyph in the output form of a str.
// The low bytes stores the glyph index.
type Glyph uint32

const (
	// `GLYPH_EMPTY` represents a `Glyph` value that has a
	// special meaning, which is a zero-width empty glyph. This is useful for
	// example in shaper modules, to use as the glyph for various zero-width
	// Unicode characters (those passing isZeroWidth()).
	GLYPH_EMPTY Glyph = 0x0FFFFFFF

	// `GLYPH_INVALID_INPUT` represents a `Glyph` value that has a
	// special meaning of invalid input. `Layout` produces one such glyph
	// per invalid input UTF-8 byte and such a glyph is rendered as a crossed
	// box.
	//
	// Note that this value is defined such that it has the `GLYPH_UNKNOWN_FLAG` on.
	GLYPH_INVALID_INPUT Glyph = 0xFFFFFFFF

	// `GLYPH_UNKNOWN_FLAG` is a flag value that can be added to
	// a rune value of a valid Unicode character, to produce a `Glyph`
	// value, representing an unknown-character glyph for the respective rune.
	GLYPH_UNKNOWN_FLAG = 0x10000000
)

// AsUnknownGlyph returns a `Glyph` value that means no glyph was found for `wc`.
//
// The way this unknown glyphs are rendered is backend specific. For example,
// a box with the hexadecimal Unicode code-point of the character written in it
// is what is done in the most common backends.
func AsUnknownGlyph(wc rune) Glyph {
	return Glyph(wc | GLYPH_UNKNOWN_FLAG)
}

func (g Glyph) GID() fonts.GID {
	return fonts.GID(g)
}

// Scale represents the scale between dimensions used
// for Pango distances and device units. (The definition of device
// units is dependent on the output device; it will typically be pixels
// for a screen, and points for a printer.) Scale is currently
// 1024, but this may be changed in the future.
//
// When setting font sizes, device units are always considered to be
// points (as in "12 point font"), rather than pixels.
const Scale = 1024

const (
	unknownGlyphWidth  = 10
	unknownGlyphHeight = 14
)

// GlyphUnit is used to store dimensions within
// Pango. Dimensions are stored in 1/Scale of a device unit.
// (A device unit might be a pixel for screen display, or
// a point on a printer.) Scale is currently 1024, and
// may change in the future (unlikely though), but you should not
// depend on its exact value. .
type GlyphUnit int32

// Pixels converts from glyph units into device units with correct rounding.
func (g GlyphUnit) Pixels() int32 {
	return (int32(g) + 512) >> 10
}

// PixelsFloor converts from glyph units into device units by flooring.
func (g GlyphUnit) PixelsFloor() int32 {
	return int32(g) >> 10
}

// PixelsCeil converts from glyph units into device units by ceiling
func (g GlyphUnit) PixelsCeil() int32 {
	return (int32(g) + 1023) >> 10
}

// Round rounds a dimension to whole device units, but does not
// convert it to device units.
func (d GlyphUnit) Round() GlyphUnit {
	return (d + Scale>>1) & ^(Scale - 1)
}

// GlyphGeometry contains width and positioning
// information for a single glyph.
type GlyphGeometry struct {
	Width   GlyphUnit // the logical width to use for the the character.
	xOffset GlyphUnit // horizontal offset from nominal character position.
	yOffset GlyphUnit // vertical offset from nominal character position.
}

// glyphVisAttr is used to communicate information between
// the shaping phase and the rendering phase.
// More attributes may be added in the future.
type glyphVisAttr struct {
	// set for the first logical glyph in each cluster. (Clusters
	// are stored in visual order, within the cluster, glyphs
	// are always ordered in logical order, since visual
	// order is meaningless; that is, in Arabic text, accent glyphs
	// follow the glyphs for the base character.)
	isClusterStart bool // =  1;
}

// GlyphInfo represents a single glyph together with
// positioning information and visual attributes.
type GlyphInfo struct {
	Glyph    Glyph         // the glyph itself.
	Geometry GlyphGeometry // the positional information about the glyph.
	attr     glyphVisAttr  // the visual attributes of the glyph.
}

// shapeFlags influences the shaping process.
// These can be passed to pango_shape_with_flags().
type shapeFlags uint8

const (
	shapeNONE shapeFlags = 0 // Default value.
	// Round glyph positions and widths to whole device units. This option should
	// be set if the target renderer can't do subpixel positioning of glyphs.
	shapeROUND_POSITIONS shapeFlags = 1
)

// GlyphString structure is used to store strings
// of glyphs with geometry and visual attribute information - ready for drawing
type GlyphString struct {
	// Array of glyph information for the glyph string
	Glyphs []GlyphInfo

	// logical cluster info, indexed by the rune index
	// within the text corresponding to the glyph string
	logClusters []int

	// space int
}

// setSize resize a glyph string to the given length, reusing
// the current storage if possible
func (str *GlyphString) setSize(newLen int) {
	if newLen <= cap(str.Glyphs) {
		str.Glyphs = str.Glyphs[:newLen]
		str.logClusters = str.logClusters[:newLen]
	} else { // re-allocate
		str.Glyphs = make([]GlyphInfo, newLen)
		str.logClusters = make([]int, newLen)
	}
}

func (glyphs GlyphString) reverse() {
	gs, lc := glyphs.Glyphs, glyphs.logClusters
	for i := len(gs)/2 - 1; i >= 0; i-- { // gs and lc have the same size
		opp := len(gs) - 1 - i
		gs[i], gs[opp] = gs[opp], gs[i]
		lc[i], lc[opp] = lc[opp], lc[i]
	}
}

// pango_glyph_string_get_width computes the logical width of the glyph string as can also be computed
// using pango_glyph_string_extents(). However, since this only computes the
// width, it's much faster.
// This is in fact only a convenience function that
// computes the sum of geometry.width for each glyph in `glyphs`.
func (glyphs *GlyphString) pango_glyph_string_get_width() GlyphUnit {
	var width GlyphUnit

	for _, g := range glyphs.Glyphs {
		width += g.Geometry.Width
	}

	return width
}

// simple shaping relying on font metrics
func (glyphs *GlyphString) fallbackShape(text []rune, analysis *Analysis) {
	glyphs.setSize(len(text))

	cluster := 0
	for i, wc := range text {
		if !unicode.Is(unicode.Mn, wc) {
			cluster = i
		}

		var glyph Glyph
		if pangoIsZeroWidth(wc) {
			glyph = GLYPH_EMPTY
		} else {
			glyph = AsUnknownGlyph(wc)
		}

		var logicalRect Rectangle
		analysis.Font.GlyphExtents(glyph, nil, &logicalRect)

		glyphs.Glyphs[i].Glyph = glyph

		glyphs.Glyphs[i].Geometry.xOffset = 0
		glyphs.Glyphs[i].Geometry.yOffset = 0
		glyphs.Glyphs[i].Geometry.Width = GlyphUnit(logicalRect.Width)

		glyphs.logClusters[i] = cluster
	}

	if analysis.Level&1 != 0 {
		glyphs.reverse()
	}
}

// Shape is a convenience shortcut for ShapeRange(text, 0, len(text), analysis).
func (glyphs *GlyphString) Shape(text []rune, analysis *Analysis) {
	glyphs.ShapeRange(text, 0, len(text), analysis)
}

// ShapeRange convert the characters into glyphs,
// using a segment of text and the corresponding
// `Analysis` structure returned from the itemization.
// You may also pass in only a substring of the item from the itemization.
//
// `paragraphText` is the full paragraph text, which will be used to perform
// certain cross-item shaping interactions. The actual text to shape is
// delimited by `itemOffset` and `itemLength`.
func (glyphs *GlyphString) ShapeRange(paragraphText []rune, itemOffset, itemLength int, analysis *Analysis) {
	glyphs.shapeWithFlags(paragraphText, itemOffset, itemLength, analysis, shapeNONE)
}

func hint(value GlyphUnit, scaleInv, scale Fl) GlyphUnit {
	return GlyphUnit(Fl(GlyphUnit(Fl(value)*scale)) * scaleInv).Round()
}

// shapeWithFlags is similar to shapeRange(), except it also takes
// flags that can influence the shaping process.
func (glyphs *GlyphString) shapeWithFlags(paragraphText []rune, itemOffset, itemLength int, analysis *Analysis,
	flags shapeFlags) {

	itemText := paragraphText[itemOffset : itemOffset+itemLength]
	if len(itemText) == 0 {
		fmt.Println("empty item !", itemOffset, itemLength, len(paragraphText))
	}
	if analysis.Font != nil {
		glyphs.pango_hb_shape(analysis.Font, analysis, paragraphText, itemOffset, itemLength)

		if len(glyphs.Glyphs) == 0 {
			if debugMode {
				// If a font has been correctly chosen, but no glyphs are output,
				// there's probably something wrong with the font.
				//
				// Trying to be informative, we print out the font description,
				// and the text, but to not flood the terminal with
				// zillions of the message, we set a flag to only err once per
				// font.
				log.Printf("shaping failure, expect ugly output. font='%s', text='%s' : %v",
					analysis.Font.Describe(false), string(itemText), itemText)
			}
		}
	}

	if len(glyphs.Glyphs) == 0 {
		glyphs.fallbackShape(itemText, analysis)
		if len(glyphs.Glyphs) == 0 {
			return
		}
	}

	// make sure last_cluster is invalid
	lastCluster := glyphs.logClusters[0] - 1
	for i, lo := range glyphs.logClusters {
		// Set glyphs[i].attr.is_cluster_start based on logClusters[]
		if lo != lastCluster {
			glyphs.Glyphs[i].attr.isClusterStart = true
			lastCluster = lo
		} else {
			glyphs.Glyphs[i].attr.isClusterStart = false
		}

		// Shift glyph if width is negative, and negate width.
		// This is useful for rotated font matrices and shouldn't harm in normal cases.
		if glyphs.Glyphs[i].Geometry.Width < 0 {
			glyphs.Glyphs[i].Geometry.Width = -glyphs.Glyphs[i].Geometry.Width
			glyphs.Glyphs[i].Geometry.xOffset += glyphs.Glyphs[i].Geometry.Width
		}
	}

	// Make sure glyphstring direction conforms to analysis.level
	if lc := glyphs.logClusters; (analysis.Level&1) != 0 && lc[0] < lc[len(lc)-1] {
		log.Println("pango: expected RTL run but got LTR. Fixing.")

		// *Fix* it so we don't crash later
		glyphs.reverse()
	}

	if flags&shapeROUND_POSITIONS != 0 {
		for i := range glyphs.Glyphs {
			glyphs.Glyphs[i].Geometry.Width = glyphs.Glyphs[i].Geometry.Width.Round()
			glyphs.Glyphs[i].Geometry.xOffset = glyphs.Glyphs[i].Geometry.xOffset.Round()
			glyphs.Glyphs[i].Geometry.yOffset = glyphs.Glyphs[i].Geometry.yOffset.Round()
		}
	}
}

func (glyphs *GlyphString) _pango_shape_shape(text []rune, shapeLogical Rectangle) {
	glyphs.setSize(len(text))

	for i := range text {
		glyphs.Glyphs[i].Glyph = GLYPH_EMPTY
		glyphs.Glyphs[i].Geometry.xOffset = 0
		glyphs.Glyphs[i].Geometry.yOffset = 0
		glyphs.Glyphs[i].Geometry.Width = GlyphUnit(shapeLogical.Width)
		glyphs.Glyphs[i].attr.isClusterStart = true
		glyphs.logClusters[i] = i
	}
}

func (glyphs *GlyphString) pad_glyphstring_right(state *paraBreakState, adjustment GlyphUnit) {
	glyph := len(glyphs.Glyphs) - 1

	for glyph >= 0 && glyphs.Glyphs[glyph].Geometry.Width == 0 {
		glyph--
	}

	if glyph < 0 {
		return
	}

	state.remaining_width -= adjustment
	glyphs.Glyphs[glyph].Geometry.Width += adjustment
	if glyphs.Glyphs[glyph].Geometry.Width < 0 {
		state.remaining_width += glyphs.Glyphs[glyph].Geometry.Width
		glyphs.Glyphs[glyph].Geometry.Width = 0
	}
}

func (glyphs *GlyphString) pad_glyphstring_left(state *paraBreakState, adjustment GlyphUnit) {
	glyph := 0

	for glyph < len(glyphs.Glyphs) && glyphs.Glyphs[glyph].Geometry.Width == 0 {
		glyph++
	}

	if glyph == len(glyphs.Glyphs) {
		return
	}

	state.remaining_width -= adjustment
	glyphs.Glyphs[glyph].Geometry.Width += adjustment
	glyphs.Glyphs[glyph].Geometry.xOffset += adjustment
}

// extentsRange computes the extents of a sub-portion of a glyph string,
// with indices such that start <= index < end.
// The extents are relative to the start of the glyph string range (the origin of their
// coordinate system is at the start of the range, not at the start of the entire
// glyph string).
func (glyphs *GlyphString) extentsRange(start, end int, font Font, inkRect, logicalRect *Rectangle) {
	// Note that the handling of empty rectangles for ink
	// and logical rectangles is different. A zero-height ink
	// rectangle makes no contribution to the overall ink rect,
	// while a zero-height logical rect still reserves horizontal
	// width. Also, we may return zero-width, positive height
	// logical rectangles, while we'll never do that for the
	// ink rect.
	if start > end {
		return
	}

	if inkRect == nil && logicalRect == nil {
		return
	}

	if inkRect != nil {
		inkRect.X, inkRect.Y, inkRect.Width, inkRect.Height = 0, 0, 0, 0
	}

	if logicalRect != nil {
		logicalRect.X, logicalRect.Y, logicalRect.Width, logicalRect.Height = 0, 0, 0, 0
	}

	var xPos int32
	for i := start; i < end; i++ {
		var glyphInk, glyphLogical Rectangle

		geometry := &glyphs.Glyphs[i].Geometry

		font.GlyphExtents(glyphs.Glyphs[i].Glyph, &glyphInk, &glyphLogical)

		if inkRect != nil && glyphInk.Width != 0 && glyphInk.Height != 0 {
			if inkRect.Width == 0 || inkRect.Height == 0 {
				inkRect.X = xPos + glyphInk.X + int32(geometry.xOffset)
				inkRect.Width = glyphInk.Width
				inkRect.Y = glyphInk.Y + int32(geometry.yOffset)
				inkRect.Height = glyphInk.Height
			} else {
				new_x := min32(inkRect.X, xPos+glyphInk.X+int32(geometry.xOffset))
				inkRect.Width = max32(inkRect.X+inkRect.Width,
					xPos+glyphInk.X+glyphInk.Width+int32(geometry.xOffset)) - new_x
				inkRect.X = new_x

				new_y := min32(inkRect.Y, glyphInk.Y+int32(geometry.yOffset))
				inkRect.Height = max32(inkRect.Y+inkRect.Height,
					glyphInk.Y+glyphInk.Height+int32(geometry.yOffset)) - new_y
				inkRect.Y = new_y
			}
		}

		if logicalRect != nil {
			logicalRect.Width += int32(geometry.Width)

			if i == start {
				logicalRect.Y = glyphLogical.Y
				logicalRect.Height = glyphLogical.Height
			} else {
				new_y := min32(logicalRect.Y, glyphLogical.Y)
				logicalRect.Height = max32(logicalRect.Y+logicalRect.Height,
					glyphLogical.Y+glyphLogical.Height) - new_y
				logicalRect.Y = new_y
			}
		}

		xPos += int32(geometry.Width)
	}
}

// Extents compute the logical and ink extents of a glyph string. See the documentation
// for Font.GlyphExtents() for details about the interpretation
// of the rectangles.
func (glyphs *GlyphString) Extents(font Font, inkRect, logicalRect *Rectangle) {
	glyphs.extentsRange(0, len(glyphs.Glyphs), font, inkRect, logicalRect)
}

/* The initial implementation here is script independent,
 * but it might actually need to be virtualized into the
 * rendering modules. Otherwise, we probably will end up
 * enforcing unnatural cursor behavior for some languages.
 *
 * The only distinction that Uniscript makes is whether
 * cursor positioning is allowed within clusters or not.
 */

// IndexToX converts from character position, given by `index` to x position. (X position
// is measured from the left edge of the run). Character positions
// are computed by dividing up each cluster into equal portions.
// If `trailing` is `false`, it computes the result for the beginning of the character.
func (glyphs *GlyphString) IndexToX(text []rune, analysis *Analysis, index int,
	trailing bool) GlyphUnit {
	var (
		endIndex, startIndex      = -1, -1
		width, endXpos, startXpos GlyphUnit
	)

	if len(text) == 0 || len(glyphs.Glyphs) == 0 {
		return 0
	}

	/* Calculate the starting and ending character positions
	* and x positions for the cluster
	 */
	if analysis.Level%2 != 0 /* Right to left */ {
		for _, g := range glyphs.Glyphs {
			width += g.Geometry.Width
		}

		for i := len(glyphs.Glyphs) - 1; i >= 0; i-- {
			if glyphs.logClusters[i] > index {
				endIndex = glyphs.logClusters[i]
				endXpos = width
				break
			}

			if glyphs.logClusters[i] != startIndex {
				startIndex = glyphs.logClusters[i]
				startXpos = width
			}

			width -= glyphs.Glyphs[i].Geometry.Width
		}
	} else /* Left to right */ {
		for i := 0; i < len(glyphs.Glyphs); i++ {
			if glyphs.logClusters[i] > index {
				endIndex = glyphs.logClusters[i]
				endXpos = width
				break
			}

			if glyphs.logClusters[i] != startIndex {
				startIndex = glyphs.logClusters[i]
				startXpos = width
			}

			width += glyphs.Glyphs[i].Geometry.Width
		}
	}

	if endIndex == -1 {
		endIndex = len(text)
		endXpos = width
		if analysis.Level%2 != 0 {
			endXpos = 0
		}
	}

	/* Calculate offset of character within cluster */

	clusterChars := GlyphUnit(endIndex - startIndex)
	clusterOffset := GlyphUnit(index - startIndex)

	if trailing {
		clusterOffset += 1
	}

	if clusterChars == 0 /* pedantic */ {
		return startXpos
	}

	return ((clusterChars-clusterOffset)*startXpos +
		clusterOffset*endXpos) / clusterChars
}
