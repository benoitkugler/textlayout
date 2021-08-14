package pango_test

import (
	"testing"

	"github.com/benoitkugler/textlayout/pango"
)

// ported from test-ellipsize.c: Test Pango harfbuzz apis

// @inclusive: (allow-none): rectangle to round to pixels inclusively, or %NULL.
// @nearest: (allow-none): rectangle to round to nearest pixels, or %NULL.
//
// Converts extents from Pango units to device units, dividing by the
// %PANGO_SCALE factor and performing rounding.
//
// The @inclusive rectangle is converted by flooring the x/y coordinates and extending
// width/height, such that the final rectangle completely includes the original
// rectangle.
//
// The @nearest rectangle is converted by rounding the coordinates
// of the rectangle to the nearest device unit (pixel).
//
// The rule to which argument to use is: if you want the resulting device-space
// rectangle to completely contain the original rectangle, pass it in as @inclusive.
// If you want two touching-but-not-overlapping rectangles stay
// touching-but-not-overlapping after rounding to device units, pass them in
// as @nearest.
func extentsToPixels(inclusive, nearest *pango.Rectangle) {
	if inclusive != nil {
		origX := inclusive.X
		origY := inclusive.Y

		inclusive.X = pango.GlyphUnit(inclusive.X.PixelsFloor())
		inclusive.Y = pango.GlyphUnit(inclusive.Y.PixelsFloor())

		inclusive.Width = pango.GlyphUnit((origX + inclusive.Width).PixelsCeil()) - inclusive.X
		inclusive.Height = pango.GlyphUnit((origY + inclusive.Height).PixelsCeil()) - inclusive.Y
	}

	if nearest != nil {
		origX := nearest.X
		origY := nearest.Y

		nearest.X = pango.GlyphUnit(nearest.X.Pixels())
		nearest.Y = pango.GlyphUnit(nearest.Y.Pixels())

		nearest.Width = pango.GlyphUnit((origX + nearest.Width).Pixels()) - nearest.X
		nearest.Height = pango.GlyphUnit((origY + nearest.Height).Pixels()) - nearest.Y
	}
}

/**
 * getPixelSize:
 * @layout: a #PangoLayout
 * @width: (out) (allow-none): location to store the logical width, or %NULL
 * @height: (out) (allow-none): location to store the logical height, or %NULL
 *
 * Determines the logical width and height of a #PangoLayout
 * in device units. (pango_layout_get_size() returns the width
 * and height scaled by %PANGO_SCALE.) This
 * is simply a convenience function around
 * pango_layout_get_pixel_extents().
 **/
func getPixelSize(layout *pango.Layout) (width, height int32) {
	var logicalRect pango.Rectangle

	layout.GetExtents(nil, &logicalRect)
	extentsToPixels(&logicalRect, nil)

	return int32(logicalRect.Width), int32(logicalRect.Height)
}

// Test that ellipsization does not change the height of a layout.
// See https://gitlab.gnome.org/GNOME/pango/issues/397
func TestEllipsizeHeight(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())

	layout := pango.NewLayout(context)

	desc := pango.NewFontDescriptionFrom("Fixed 7")
	layout.SetFontDescription(&desc)

	layout.SetText("some text that should be ellipsized")
	if L := layout.GetLineCount(); L != 1 {
		t.Fatalf("expected 1 line, got %d", L)
	}
	_, height1 := getPixelSize(layout)

	layout.SetWidth(100 * pango.Scale)
	layout.SetEllipsize(pango.ELLIPSIZE_END)

	if L := layout.GetLineCount(); L != 1 {
		t.Fatalf("expected 1 line, got %d", L)
	}
	if L := layout.IsEllipsized(); !L {
		t.Fatal("expected ellipsized")
	}
	_, height2 := getPixelSize(layout)

	if height1 != height2 {
		t.Fatalf("unexpected heights: %d != %d", height1, height2)
	}
}

// Test that ellipsization without attributes does not crash
func TestEllipsizeCrash(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())

	layout := pango.NewLayout(context)

	layout.SetText("some text that should be ellipsized")
	if L := layout.GetLineCount(); L != 1 {
		t.Fatalf("expected 1 line, got %d", L)
	}
	layout.SetWidth(100 * pango.Scale)
	layout.SetEllipsize(pango.ELLIPSIZE_END)

	if L := layout.GetLineCount(); L != 1 {
		t.Fatalf("expected 1 line, got %d", L)
	}
	if L := layout.IsEllipsized(); !L {
		t.Fatal("expected ellipsized")
	}
}

/* Check that the width of a fully ellipsized paragraph
 * is the same as that of an explicit ellipsis.
 */
func TestEllipsizeFully(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())

	layout := pango.NewLayout(context)
	layout.SetText("…")

	var ink, logical pango.Rectangle
	layout.GetExtents(&ink, &logical)

	layout.SetText("ellipsized")
	layout.SetWidth(10 * pango.Scale)
	layout.SetEllipsize(pango.ELLIPSIZE_END)

	var ink2, logical2 pango.Rectangle
	layout.GetExtents(&ink2, &logical2)

	if ink.Width != ink2.Width {
		t.Fatalf("expected same widths, got %d and %d", ink.Height, ink2.Width)
	}
	if logical.Width != logical2.Width {
		t.Fatalf("expected same widths, got %d and %d", logical.Height, logical2.Width)
	}
}
