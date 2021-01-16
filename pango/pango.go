package pango

import (
	"fmt"
	"log"
	"math"

	"github.com/benoitkugler/go-weasyprint/fribidi"
	"golang.org/x/text/width"
)

// enables additional checks, to use only during developpement or testing
const debugMode = false

// assert is only used in debug mode
func assert(b bool) {
	if !b {
		log.Fatal("assertion error")
	}
}

func showDebug(where string, line *LayoutLine, state *ParaBreakState) {
	line_width := line.pango_layout_line_get_width()

	fmt.Printf("rem %d + line %d = %d		%s",
		state.remaining_width,
		line_width,
		state.remaining_width+line_width,
		where)
}

// Alignment describes how to align the lines of a `Layout` within the
// available space. If the `Layout` is set to justify
// using pango_layout_set_justify(), this only has effect for partial lines.
type Alignment uint8

const (
	PANGO_ALIGN_LEFT   Alignment = iota // Put all available space on the right
	PANGO_ALIGN_CENTER                  // Center the line within the available space
	PANGO_ALIGN_RIGHT                   // Put all available space on the left
)

// Rectangle represents a rectangle. It is frequently
// used to represent the logical or ink extents of a single glyph or section
// of text. (See, for instance, pango_font_get_glyph_extents())
type Rectangle struct {
	x      int // X coordinate of the left side of the rectangle.
	y      int // Y coordinate of the the top side of the rectangle.
	width  int // width of the rectangle.
	height int // height of the rectangle.
}

const maxInt = int(^uint32(0) >> 1)

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func minL(a, b fribidi.Level) fribidi.Level {
	if a > b {
		return b
	}
	return a
}

func maxG(a, b GlyphUnit) GlyphUnit {
	if a < b {
		return b
	}
	return a
}

// pango_is_zero_width checks `ch` to see if it is a character that should not be
// normally rendered on the screen.  This includes all Unicode characters
// with "ZERO WIDTH" in their name, as well as bidi formatting characters, and
// a few other ones.
func pango_is_zero_width(ch rune) bool {
	//  00AD  SOFT HYPHEN
	//  034F  COMBINING GRAPHEME JOINER
	//
	//  200B  ZERO WIDTH SPACE
	//  200C  ZERO WIDTH NON-JOINER
	//  200D  ZERO WIDTH JOINER
	//  200E  LEFT-TO-RIGHT MARK
	//  200F  RIGHT-TO-LEFT MARK
	//
	//  2028  LINE SEPARATOR
	//
	//  202A  LEFT-TO-RIGHT EMBEDDING
	//  202B  RIGHT-TO-LEFT EMBEDDING
	//  202C  POP DIRECTIONAL FORMATTING
	//  202D  LEFT-TO-RIGHT OVERRIDE
	//  202E  RIGHT-TO-LEFT OVERRIDE
	//
	//  2060  WORD JOINER
	//  2061  FUNCTION APPLICATION
	//  2062  INVISIBLE TIMES
	//  2063  INVISIBLE SEPARATOR
	//
	//  FEFF  ZERO WIDTH NO-BREAK SPACE
	return (ch & ^0x007F == 0x2000 &&
		((ch >= 0x200B && ch <= 0x200F) || (ch >= 0x202A && ch <= 0x202E) ||
			(ch >= 0x2060 && ch <= 0x2063) || (ch == 0x2028))) ||
		(ch == 0x00AD || ch == 0x034F || ch == 0xFEFF)
}

// return true for east asian wide characters
func isWide(r rune) bool {
	switch width.LookupRune(r).Kind() {
	case width.EastAsianFullwidth, width.EastAsianWide:
		return true
	default:
		return false
	}
}

// Matrix is a transformation between user-space
// coordinates and device coordinates. The transformation
// is given by
//
// x_device = x_user * matrix.xx + y_user * matrix.xy + matrix.x0;
// y_device = x_user * matrix.yx + y_user * matrix.yy + matrix.y0;
type Matrix struct {
	xx, xy, yx, yy, x0, y0 float64
}

var PANGO_MATRIX_INIT = Matrix{1, 0, 0, 1, 0, 0}

/**
 * pango_matrix_get_font_scale_factor:
 * @matrix: (allow-none): a #PangoMatrix, may be %NULL
 *
 * Returns the scale factor of a matrix on the height of the font.
 * That is, the scale factor in the direction perpendicular to the
 * vector that the X coordinate is mapped to.  If the scale in the X
 * coordinate is needed as well, use pango_matrix_get_font_scale_factors().
 *
 * Return value: the scale factor of @matrix on the height of the font,
 * or 1.0 if @matrix is %NULL.
 *
 * Since: 1.12
 **/
func (matrix Matrix) pango_matrix_get_font_scale_factor() float64 {
	_, yscale := matrix.pango_matrix_get_font_scale_factors()
	return yscale
}

/**
 * pango_matrix_get_font_scale_factors:
 * @matrix: (nullable): a #PangoMatrix, or %NULL
 * @xscale: (out) (allow-none): output scale factor in the x direction, or %NULL
 * @yscale: (out) (allow-none): output scale factor perpendicular to the x direction, or %NULL
 *
 * Calculates the scale factor of a matrix on the width and height of the font.
 * That is, @xscale is the scale factor in the direction of the X coordinate,
 * and @yscale is the scale factor in the direction perpendicular to the
 * vector that the X coordinate is mapped to.
 *
 * Note that output numbers will always be non-negative.
 **/
func (matrix Matrix) pango_matrix_get_font_scale_factors() (xscale, yscale float64) {
	// Based on cairo-matrix.c:_cairo_matrix_compute_scale_factors()
	// Copyright 2005, Keith Packard

	x := matrix.xx
	y := matrix.yx
	xscale = math.Sqrt(x*x + y*y)

	if xscale != 0 {
		det := matrix.xx*matrix.yy - matrix.yx*matrix.xy

		/*
		* ignore mirroring
		 */
		if det < 0 {
			det = -det
		}

		yscale = det / xscale
	} else {
		yscale = 0.
	}
	return
}
