package pango

import (
	"testing"
)

func TestParse(t *testing.T) {
	desc := pango_font_description_from_string("Cantarell 14")

	assertEquals(t, desc.FamilyName, "Cantarell")
	assertFalse(t, desc.SizeIsAbsolute, "font size is not absolute")
	assertEquals(t, desc.Size, 14*PangoScale)
	assertEquals(t, desc.Style, STYLE_NORMAL)
	assertEquals(t, desc.Variant, PANGO_VARIANT_NORMAL)
	assertEquals(t, desc.Weight, PANGO_WEIGHT_NORMAL)
	assertEquals(t, desc.Stretch, STRETCH_NORMAL)
	assertEquals(t, desc.Gravity, PANGO_GRAVITY_SOUTH)
	assertEquals(t, desc.mask, F_FAMILY|F_STYLE|F_VARIANT|F_WEIGHT|F_STRETCH|F_SIZE)

	desc = pango_font_description_from_string("Sans Bold Italic Condensed 22.5px")

	assertEquals(t, desc.FamilyName, "Sans")
	assertTrue(t, desc.SizeIsAbsolute, "font size is absolute")
	assertEquals(t, desc.Size, 225*PangoScale/10)
	assertEquals(t, desc.Style, STYLE_ITALIC)
	assertEquals(t, desc.Variant, PANGO_VARIANT_NORMAL)
	assertEquals(t, desc.Weight, PANGO_WEIGHT_BOLD)
	assertEquals(t, desc.Stretch, STRETCH_CONDENSED)
	assertEquals(t, desc.Gravity, PANGO_GRAVITY_SOUTH)
	assertEquals(t, desc.mask, F_FAMILY|F_STYLE|F_VARIANT|F_WEIGHT|F_STRETCH|F_SIZE)
}

func TestRoundtrip(t *testing.T) {
	desc := pango_font_description_from_string("Cantarell 14")
	str := desc.String()
	assertEquals(t, str, "Cantarell 14")

	desc = pango_font_description_from_string("Sans Bold Italic Condensed 22.5px")
	str = desc.String()
	assertEquals(t, str, "Sans Bold Italic Condensed 22.5px")
}

func TestVariation(t *testing.T) {
	desc1 := pango_font_description_from_string("Cantarell 14")
	assertTrue(t, desc1.mask&F_VARIATIONS == 0, "no variations")
	assertTrue(t, desc1.Variations == "", "variations is empty")

	str := desc1.String()
	assertEquals(t, str, "Cantarell 14")

	desc2 := pango_font_description_from_string("Cantarell 14 @wght=100,wdth=235")
	assertTrue(t, desc2.mask&F_VARIATIONS != 0, "has variations")
	assertEquals(t, desc2.Variations, "wght=100,wdth=235")

	str = desc2.String()
	assertEquals(t, str, "Cantarell 14 @wght=100,wdth=235")

	assertFalse(t, desc1.pango_font_description_equal(desc2), "different font descriptions")

	desc1.Setvariations("wght=100,wdth=235")
	assertTrue(t, desc1.mask&F_VARIATIONS != 0, "has variations")
	assertEquals(t, desc1.Variations, "wght=100,wdth=235")

	assertTrue(t, desc1.pango_font_description_equal(desc2), "same fonts")
}

func TestMetrics(t *testing.T) {
	//    PangoFontDescription *desc;
	//    PangoFontMetrics *metrics;
	//    char *str;

	//    if (strcmp (G_OBJECT_TYPE_NAME (pango_context_get_font_map (context)), "PangoCairoWin32FontMap") == 0)
	// 	 desc = pango_font_description_from_string ("Verdana 11");
	//    else
	// 	 desc = pango_font_description_from_string ("Cantarell 11");

	//    str = pango_font_description_to_string (desc);

	//    metrics = pango_context_get_metrics (context, desc, pango_language_get_default ());

	//    g_test_message ("%s metrics", str);
	//    g_test_message ("\tascent: %d", pango_font_metrics_get_ascent (metrics));
	//    g_test_message ("\tdescent: %d", pango_font_metrics_get_descent (metrics));
	//    g_test_message ("\theight: %d", pango_font_metrics_get_height (metrics));
	//    g_test_message ("\tchar width: %d",
	// 				   pango_font_metrics_get_approximate_char_width (metrics));
	//    g_test_message ("\tdigit width: %d",
	// 				   pango_font_metrics_get_approximate_digit_width (metrics));
	//    g_test_message ("\tunderline position: %d",
	// 				   pango_font_metrics_get_underline_position (metrics));
	//    g_test_message ("\tunderline thickness: %d",
	// 				   pango_font_metrics_get_underline_thickness (metrics));
	//    g_test_message ("\tstrikethrough position: %d",
	// 				   pango_font_metrics_get_strikethrough_position (metrics));
	//    g_test_message ("\tstrikethrough thickness: %d",
	// 				   pango_font_metrics_get_strikethrough_thickness (metrics));

	//    pango_font_metrics_unref (metrics);
	//    g_free (str);
	//    pango_font_description_free (desc);
}

func TestExtents(t *testing.T) {
	//    char *str = "Composer";
	//    GList *items;
	//    PangoItem *item;
	//    PangoGlyphString *glyphs;
	//    PangoRectangle ink, log;
	//    PangoContext *context;

	//    context = NewContext (pango_cairo_font_map_get_default ());
	//    pango_context_set_font_description (context, pango_font_description_from_string ("Cantarell 11"));

	//    items = pango_itemize (context, str, 0, strlen (str), nil, nil);
	//    glyphs = pango_glyph_string_new ();
	//    item = items->data;
	//    pango_shape (str, strlen (str), &item->analysis, glyphs);
	//    pango_glyph_string_extents (glyphs, item->analysis.font, &ink, &log);

	//    g_assert_cmpint (ink.width, >=, 0);
	//    g_assert_cmpint (ink.height, >=, 0);
	//    g_assert_cmpint (log.width, >=, 0);
	//    g_assert_cmpint (log.height, >=, 0);

	//    pango_glyph_string_free (glyphs);
	//    g_list_free_full (items, (GDestroyNotify)pango_item_free);
	//    g_object_unref (context);
}

func TestEnumerate(t *testing.T) {
	//    PangoFontMap *fontmap;
	//    PangoContext *context;
	//    PangoFontFamily **families;
	//    PangoFontFamily *family;
	//    int n_families;
	//    int i;
	//    PangoFontFace **faces;
	//    PangoFontFace *face;
	//    int n_faces;
	//    PangoFontDescription *desc;
	//    PangoFont *font;
	//    gboolean found_face;

	//    fontmap = pango_cairo_font_map_get_default ();
	//    context = NewContext (fontmap);

	//    pango_font_map_list_families (fontmap, &families, &n_families);
	//    g_assert_cmpint (n_families, >, 0);

	//    for (i = 0; i < n_families; i++)
	// 	 {
	// 	   family = pango_font_map_GetFamily (fontmap, pango_font_family_GetName (families[i]));
	// 	   g_assert_true (family == families[i]);
	// 	 }

	//    pango_font_family_ListFaces (families[0], &faces, &n_faces);
	//    g_assert_cmpint (n_faces, >, 0);
	//    for (i = 0; i < n_faces; i++)
	// 	 {
	// 	   face = pango_font_family_GetFace (families[0], pango_font_face_GetFaceName (faces[i]));
	// 	   g_assert_true (face == faces[i]);
	// 	 }

	//    desc = NewFontDescription ();
	//    Setfamily (desc, pango_font_family_GetName (families[0]));
	//    SetSize (desc, 10*PangoScale);

	//    font = pango_font_map_load_font (fontmap, context, desc);
	//    face = pango_font_GetFace (font);
	//    found_face = FALSE;
	//    for (i = 0; i < n_faces; i++)
	// 	 {
	// 	   if (face == faces[i])
	// 		 {
	// 		   found_face = TRUE;
	// 		   break;
	// 		 }
	// 	 }
	//    g_assert_true (found_face);

	//    g_object_unref (font);
	//    pango_font_description_free (desc);
	//    g_free (faces);
	//    g_free (families);
	//    g_object_unref (context);
	//    g_object_unref (fontmap);
	//  }

	//  int
	//  main (int argc, char *argv[])
	//  {
	//    g_setenv ("LC_ALL", "C", TRUE);
	//    setlocale (LC_ALL, "");

	//    g_test_init (&argc, &argv, nil);

	//    context = NewContext (pango_cairo_font_map_get_default ());

	//    g_test_add_func ("/pango/font/metrics", test_metrics);
	//    g_test_add_func ("/pango/fontdescription/parse", test_parse);
	//    g_test_add_func ("/pango/fontdescription/roundtrip", test_roundtrip);
	//    g_test_add_func ("/pango/fontdescription/variation", test_variation);
	//    g_test_add_func ("/pango/font/extents", test_extents);
	//    g_test_add_func ("/pango/font/enumerate", test_enumerate);

	//    return g_test_run ();
}
