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
func pango_extents_to_pixels(inclusive, nearest *pango.Rectangle) {
	if inclusive != nil {
		orig_x := inclusive.X
		orig_y := inclusive.Y

		inclusive.X = pango.GlyphUnit(inclusive.X).PixelsFloor()
		inclusive.Y = pango.GlyphUnit(inclusive.Y).PixelsFloor()

		inclusive.Width = pango.GlyphUnit(orig_x+inclusive.Width).PixelsCeil() - inclusive.X
		inclusive.Height = pango.GlyphUnit(orig_y+inclusive.Height).PixelsCeil() - inclusive.Y
	}

	if nearest != nil {
		orig_x := nearest.X
		orig_y := nearest.Y

		nearest.X = pango.GlyphUnit(nearest.X).Pixels()
		nearest.Y = pango.GlyphUnit(nearest.Y).Pixels()

		nearest.Width = pango.GlyphUnit(orig_x+nearest.Width).Pixels() - nearest.X
		nearest.Height = pango.GlyphUnit(orig_y+nearest.Height).Pixels() - nearest.Y
	}
}

/**
 * pango_layout_get_pixel_size:
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
func pango_layout_get_pixel_size(layout *pango.Layout) (width, height int32) {
	var logical_rect pango.Rectangle

	layout.GetExtents(nil, &logical_rect)
	pango_extents_to_pixels(&logical_rect, nil)

	return logical_rect.Width, logical_rect.Height
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
	_, height1 := pango_layout_get_pixel_size(layout)

	layout.SetWidth(100 * pango.Scale)
	layout.SetEllipsize(pango.PANGO_ELLIPSIZE_END)

	if L := layout.GetLineCount(); L != 1 {
		t.Fatalf("expected 1 line, got %d", L)
	}
	if L := layout.IsEllipsized(); !L {
		t.Fatal("expected ellipsized")
	}
	_, height2 := pango_layout_get_pixel_size(layout)

	if height1 != height2 {
		t.Fatalf("unexpected heights: %d != %d", height1, height2)
	}
}

// Test that ellipsization without attributes does not crash
func TestEllipsizeCrash(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	// context.SetFontDescription(pango.NewFontDescriptionFrom("DejaVu Serif 12"))

	layout := pango.NewLayout(context)

	layout.SetText("some text that should be ellipsized")
	if L := layout.GetLineCount(); L != 1 {
		t.Fatalf("expected 1 line, got %d", L)
	}
	layout.SetWidth(100 * pango.Scale)
	layout.SetEllipsize(pango.PANGO_ELLIPSIZE_END)

	if L := layout.GetLineCount(); L != 1 {
		t.Fatalf("expected 1 line, got %d", L)
	}
	if L := layout.IsEllipsized(); !L {
		t.Fatal("expected ellipsized")
	}
}

//  /* Check that the width of a fully ellipsized paragraph
//   * is the same as that of an explicit ellipsis.
//   */
// func TestEllipsizeFully(t* testing.T)  {
//    PangoLayout *layout;
//    PangoRectangle ink, logical;
//    PangoRectangle ink2, logical2;

//    layout = pango_layout_new (context);

//    pango_layout_set_text (layout, "…", -1);
//    pango_layout_get_extents (layout, &ink, &logical);

//    pango_layout_set_text (layout, "ellipsized", -1);

//    pango_layout_set_width (layout, 10 * PANGO_SCALE);
//    pango_layout_set_ellipsize (layout, PANGO_ELLIPSIZE_END);

//    pango_layout_get_extents (layout, &ink2, &logical2);

//    g_assert_cmpint (ink.width, ==, ink2.width);
//    g_assert_cmpint (logical.width, ==, logical2.width);

//    g_object_unref (layout);
//  }

//  int
//  main (int argc, char *argv[])
//  {
//    PangoFontMap *fontmap;

//    fontmap = pango_cairo_font_map_get_default ();
//    context = pango_font_map_create_context (fontmap);

//    g_test_init (&argc, &argv, NULL);

//    g_test_add_func ("/layout/ellipsize/height", test_ellipsize_height);
//    g_test_add_func ("/layout/ellipsize/crash", test_ellipsize_crash);
//    g_test_add_func ("/layout/ellipsize/fully", test_ellipsize_fully);

//    return g_test_run ();
//  }
