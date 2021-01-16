package pango

import (
	"log"
	"unicode"
)

// Glyph represents a single glyph in the output form of a str.
type Glyph uint32

// pangoScale represents the scale between dimensions used
// for Pango distances and device units. (The definition of device
// units is dependent on the output device; it will typically be pixels
// for a screen, and points for a printer.) pangoScale is currently
// 1024, but this may be changed in the future.
//
// When setting font sizes, device units are always considered to be
// points (as in "12 point font"), rather than pixels.
const pangoScale = 1024

const PANGO_UNKNOWN_GLYPH_WIDTH = 10
const PANGO_UNKNOWN_GLYPH_HEIGHT = 14

// GlyphUnit is used to store dimensions within
// Pango. Dimensions are stored in 1/pangoScale of a device unit.
// (A device unit might be a pixel for screen display, or
// a point on a printer.) pangoScale is currently 1024, and
// may change in the future (unlikely though), but you should not
// depend on its exact value. .
type GlyphUnit int32

// Pixels converts from glyph units into device units with correct rounding.
func (g GlyphUnit) Pixels() int {
	return (int(g) + 512) >> 10
}

// PANGO_UNITS_ROUND rounds a dimension to whole device units, but does not
// convert it to device units.
func (d GlyphUnit) PANGO_UNITS_ROUND() GlyphUnit {
	return (d + pangoScale>>1) & ^(pangoScale - 1)
}

// GlyphGeometry contains width and positioning
// information for a single glyph.
type GlyphGeometry struct {
	width    GlyphUnit // the logical width to use for the the character.
	x_offset GlyphUnit // horizontal offset from nominal character position.
	y_offset GlyphUnit // vertical offset from nominal character position.
}

// GlyphVisAttr is used to communicate information between
// the shaping phase and the rendering phase.
// More attributes may be added in the future.
type GlyphVisAttr struct {
	// set for the first logical glyph in each cluster. (Clusters
	// are stored in visual order, within the cluster, glyphs
	// are always ordered in logical order, since visual
	// order is meaningless; that is, in Arabic text, accent glyphs
	// follow the glyphs for the base character.)
	is_cluster_start bool // =  1;
}

// GlyphInfo represents a single glyph together with
// positioning information and visual attributes.
type GlyphInfo struct {
	glyph    Glyph         // the glyph itself.
	geometry GlyphGeometry // the positional information about the glyph.
	attr     GlyphVisAttr  // the visual attributes of the glyph.
}

// ShapeFlags influences the shaping process.
// These can be passed to pango_shape_with_flags().
type ShapeFlags uint8

const (
	PANGO_SHAPE_NONE ShapeFlags = 0 // Default value.
	// Round glyph positions and widths to whole device units. This option should
	// be set if the target renderer can't do subpixel positioning of glyphs.
	PANGO_SHAPE_ROUND_POSITIONS ShapeFlags = 1
)

// GlyphString structure is used to store strings
// of glyphs with geometry and visual attribute information - ready for drawing
type GlyphString struct {
	// array of glyph information for the glyph string
	// with size num_glyphs
	glyphs []GlyphInfo

	// logical cluster info, indexed by the rune index
	// within the text corresponding to the glyph string
	log_clusters []int

	// space int
}

//  pango_glyph_string_set_size resize a glyph string to the given length.
func (str *GlyphString) pango_glyph_string_set_size(newLen int) {
	if newLen < 0 {
		return
	}
	// the C implementation does a much more careful re-allocation...
	str.glyphs = make([]GlyphInfo, newLen)
	str.log_clusters = make([]int, newLen)
}

func (glyphs GlyphString) reverse() {
	gs, lc := glyphs.glyphs, glyphs.log_clusters
	for i := len(gs)/2 - 1; i >= 0; i-- { // gs and lc have the same size
		opp := len(gs) - 1 - i
		gs[i], gs[opp] = gs[opp], gs[i]
		lc[i], lc[opp] = lc[opp], lc[i]
	}
}

// pango_glyph_string_get_width computes the logical width of the glyph string as can also be computed
// using pango_glyph_string_extents().  However, since this only computes the
// width, it's much faster.
// This is in fact only a convenience function that
// computes the sum of geometry.width for each glyph in `glyphs`.
func (glyphs *GlyphString) pango_glyph_string_get_width() GlyphUnit {
	var width GlyphUnit

	for _, g := range glyphs.glyphs {
		width += g.geometry.width
	}

	return width
}

func (glyphs *GlyphString) fallback_shape(text []rune, analysis *Analysis) {

	glyphs.pango_glyph_string_set_size(len(text))

	cluster := 0
	for i, wc := range text {
		if !unicode.Is(unicode.Mn, wc) {
			cluster = i
		}

		var glyph Glyph
		if pango_is_zero_width(wc) {
			glyph = PANGO_GLYPH_EMPTY
		} else {
			glyph = PANGO_GET_UNKNOWN_GLYPH(wc)
		}

		var logicalRect Rectangle
		analysis.font.get_glyph_extents(glyph, nil, &logicalRect)

		glyphs.glyphs[i].glyph = glyph

		glyphs.glyphs[i].geometry.x_offset = 0
		glyphs.glyphs[i].geometry.y_offset = 0
		glyphs.glyphs[i].geometry.width = GlyphUnit(logicalRect.width)

		glyphs.log_clusters[i] = cluster
	}

	if analysis.level&1 != 0 {
		glyphs.reverse()
	}
}

// pango_shape_full convertc the characters into glyphs,
// using a segment of text and the corresponding
// `Analysis` structure returned from pango_itemize().
// You may also pass in only a substring of the item from pango_itemize().
//
// This is similar to pango_shape(), except it also can optionally take
// the full paragraph text as input, which will then be used to perform
// certain cross-item shaping interactions.  If you have access to the broader
// text of which `itemText` is part of, provide the broader text as
// `paragraphText`.  If `paragraphText` is empty, item text is used instead.
//
// Note that the extra attributes in the @analyis that is returned from
// pango_itemize() have indices that are relative to the entire paragraph,
// so you do not pass the full paragraph text as @paragraphText, you need
// to subtract the item offset from their indices before calling pango_shape_full().
func (glyphs *GlyphString) pango_shape_full(itemText, paragraphText []rune, analysis *Analysis) {
	glyphs.pango_shape_with_flags(itemText, paragraphText, analysis, PANGO_SHAPE_NONE)
}

// pango_shape_with_flags is similar to pango_shape_full(), except it also takes
// flags that can influence the shaping process.
func (glyphs *GlyphString) pango_shape_with_flags(itemText, paragraphText []rune, analysis *Analysis,
	flags ShapeFlags) {

	if len(paragraphText) == 0 {
		paragraphText = itemText
	}

	if analysis.font != nil {
		pango_hb_shape(analysis.font, itemText, analysis, glyphs, paragraphText)

		if len(glyphs.glyphs) == 0 {
			// If a font has been correctly chosen, but no glyphs are output,
			// there's probably something wrong with the font.
			//
			// Trying to be informative, we print out the font description,
			// and the text, but to not flood the terminal with
			// zillions of the message, we set a flag to only err once per
			// font.

			if !fontShapeFailWarnings[analysis.font] {
				fontName := analysis.font.describe().String()

				log.Printf("shaping failure, expect ugly output. font='%s', text='%s'", fontName, string(itemText))

				fontShapeFailWarningsLock.Lock()
				fontShapeFailWarnings[analysis.font] = true
				fontShapeFailWarningsLock.Unlock()
			}
		}
	}

	if len(glyphs.glyphs) == 0 {
		glyphs.fallback_shape(itemText, analysis)
		if len(glyphs.glyphs) == 0 {
			return
		}
	}

	// make sure last_cluster is invalid
	last_cluster := glyphs.log_clusters[0] - 1
	for i, lo := range glyphs.log_clusters {
		// Set glyphs[i].attr.is_cluster_start based on log_clusters[]
		if lo != last_cluster {
			glyphs.glyphs[i].attr.is_cluster_start = true
			last_cluster = lo
		} else {
			glyphs.glyphs[i].attr.is_cluster_start = false
		}

		// Shift glyph if width is negative, and negate width.
		// This is useful for rotated font matrices and shouldn't harm in normal cases.
		if glyphs.glyphs[i].geometry.width < 0 {
			glyphs.glyphs[i].geometry.width = -glyphs.glyphs[i].geometry.width
			glyphs.glyphs[i].geometry.x_offset += glyphs.glyphs[i].geometry.width
		}
	}

	// Make sure glyphstring direction conforms to analysis.level
	if lc := glyphs.log_clusters; (analysis.level&1) != 0 && lc[0] < lc[len(lc)-1] {
		log.Println("pango: expected RTL run but got LTR. Fixing.")

		// *Fix* it so we don't crash later
		glyphs.reverse()
	}

	if flags&PANGO_SHAPE_ROUND_POSITIONS != 0 {
		for i := range glyphs.glyphs {
			glyphs.glyphs[i].geometry.width = glyphs.glyphs[i].geometry.width.PANGO_UNITS_ROUND()
			glyphs.glyphs[i].geometry.x_offset = glyphs.glyphs[i].geometry.x_offset.PANGO_UNITS_ROUND()
			glyphs.glyphs[i].geometry.y_offset = glyphs.glyphs[i].geometry.y_offset.PANGO_UNITS_ROUND()
		}
	}
}

func (glyphs *GlyphString) _pango_shape_shape(text []rune, shapeLogical Rectangle) {
	glyphs.pango_glyph_string_set_size(len(text))

	for i := range text {
		glyphs.glyphs[i].glyph = PANGO_GLYPH_EMPTY
		glyphs.glyphs[i].geometry.x_offset = 0
		glyphs.glyphs[i].geometry.y_offset = 0
		glyphs.glyphs[i].geometry.width = GlyphUnit(shapeLogical.width)
		glyphs.glyphs[i].attr.is_cluster_start = true
		glyphs.log_clusters[i] = i
	}
}

func (glyphs *GlyphString) pad_glyphstring_right(state *ParaBreakState, adjustment GlyphUnit) {
	glyph := len(glyphs.glyphs) - 1

	for glyph >= 0 && glyphs.glyphs[glyph].geometry.width == 0 {
		glyph--
	}

	if glyph < 0 {
		return
	}

	state.remaining_width -= adjustment
	glyphs.glyphs[glyph].geometry.width += adjustment
	if glyphs.glyphs[glyph].geometry.width < 0 {
		state.remaining_width += glyphs.glyphs[glyph].geometry.width
		glyphs.glyphs[glyph].geometry.width = 0
	}
}

func (glyphs *GlyphString) pad_glyphstring_left(state *ParaBreakState, adjustment GlyphUnit) {
	glyph := 0

	for glyph < len(glyphs.glyphs) && glyphs.glyphs[glyph].geometry.width == 0 {
		glyph++
	}

	if glyph == len(glyphs.glyphs) {
		return
	}

	state.remaining_width -= adjustment
	glyphs.glyphs[glyph].geometry.width += adjustment
	glyphs.glyphs[glyph].geometry.x_offset += adjustment
}

// pango_glyph_string_extents_range computes the extents of a sub-portion of a glyph string,
// with indices such that start <= index < end.
// The extents are relative to the start of the glyph string range (the origin of their
// coordinate system is at the start of the range, not at the start of the entire
// glyph string).
func (glyphs *GlyphString) pango_glyph_string_extents_range(start, end int, font Font, inkRect, logicalRect *Rectangle) {
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
		inkRect.x, inkRect.y, inkRect.width, inkRect.height = 0, 0, 0, 0
	}

	if logicalRect != nil {
		logicalRect.x, logicalRect.y, logicalRect.width, logicalRect.height = 0, 0, 0, 0
	}

	var xPos int
	for i := start; i < end; i++ {
		var glyphInk, glyphLogical Rectangle

		geometry := &glyphs.glyphs[i].geometry

		font.get_glyph_extents(glyphs.glyphs[i].glyph, &glyphInk, &glyphLogical)

		if inkRect != nil && glyphInk.width != 0 && glyphInk.height != 0 {
			if inkRect.width == 0 || inkRect.height == 0 {
				inkRect.x = xPos + glyphInk.x + int(geometry.x_offset)
				inkRect.width = glyphInk.width
				inkRect.y = glyphInk.y + int(geometry.y_offset)
				inkRect.height = glyphInk.height
			} else {
				new_x := min(inkRect.x, xPos+glyphInk.x+int(geometry.x_offset))
				inkRect.width = max(inkRect.x+inkRect.width,
					xPos+glyphInk.x+glyphInk.width+int(geometry.x_offset)) - new_x
				inkRect.x = new_x

				new_y := min(inkRect.y, glyphInk.y+int(geometry.y_offset))
				inkRect.height = max(inkRect.y+inkRect.height,
					glyphInk.y+glyphInk.height+int(geometry.y_offset)) - new_y
				inkRect.y = new_y
			}
		}

		if logicalRect != nil {
			logicalRect.width += int(geometry.width)

			if i == start {
				logicalRect.y = glyphLogical.y
				logicalRect.height = glyphLogical.height
			} else {
				new_y := min(logicalRect.y, glyphLogical.y)
				logicalRect.height = max(logicalRect.y+logicalRect.height,
					glyphLogical.y+glyphLogical.height) - new_y
				logicalRect.y = new_y
			}
		}

		xPos += int(geometry.width)
	}
}

// pango_glyph_string_extents compute the logical and ink extents of a glyph string. See the documentation
// for pango_font_get_glyph_extents() for details about the interpretation
// of the rectangles.
func (glyphs *GlyphString) pango_glyph_string_extents(font Font, inkRect, logicalRect *Rectangle) {
	glyphs.pango_glyph_string_extents_range(0, len(glyphs.glyphs), font, inkRect, logicalRect)
}
