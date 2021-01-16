// Package fcfonts is an implementation of
// the font tooling required by Pango, using textlayout/fontconfig
// and textlayout/fonts.
package fcfonts

import (
	"github.com/benoitkugler/textlayout/fontconfig"
	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
)

// ported from pangofc-font.c:

var (
	_ pango.Font     = (*PangoFcFont)(nil)
	_ pango.Font     = (*PangoFT2Font)(nil)
	_ pango.Coverage = coverage{}
)

type coverage struct {
	fc.Charset
}

// Convert the given `charset` into a new Coverage object.
func fromCharset(charset fc.Charset) pango.Coverage {
	return coverage{charset.Copy()}
}

// Get returns true if the rune is covered
func (c coverage) Get(index rune) bool { return c.HasChar(index) }

func (c coverage) Set(index rune, covered bool) {
	if covered {
		c.AddChar(index)
	} else {
		c.DelChar(index)
	}
}

// Copy returns a deep copy of the coverage
func (c coverage) Copy() pango.Coverage { return coverage{c.Charset.Copy()} }

// Decoder represents a decoder that an application provides
// for handling a font that is encoded in a custom way.
type Decoder interface {
	// GetCharset returns a charset given a font that
	// includes a list of supported characters in the font.
	// The implementation must be fast because the method is called
	// separately for each character to determine Unicode coverage.
	GetCharset(font *PangoFcFont) fontconfig.Charset

	// GetGlyph returns a single glyph for a given Unicode code point.
	GetGlyph(font *PangoFcFont, r rune) pango.Glyph
}

type FcFontPrivate struct {
	decoder Decoder
	key     *PangoFcFontKey
}

type PangoFT2Font struct {
	PangoFcFont

	//   FT_Face face;
	//   int load_flags;
	size int

	//   GSList *metrics_by_lang;

	//   GHashTable *glyph_info;
	//   GDestroyNotify glyph_cache_destroy;
}

func newPangoFT2Font(pattern fontconfig.Pattern, fontmap *PangoFcFontMap) *PangoFT2Font {
	var ft2font PangoFT2Font
	if ds := pattern.GetFloats(fontconfig.FC_PIXEL_SIZE); len(ds) != 0 {
		ft2font.size = int(ds[0] * float64(pango.PangoScale))
	}
	ft2font.font_pattern = pattern
	ft2font.fontmap = fontmap
	return &ft2font
}

//  #define PANGO_FT2_FONT_CLASS(klass)      (G_TYPE_CHECK_CLASS_CAST ((klass), PANGO_TYPE_FT2_FONT, PangoFT2FontClass))
//  #define PANGO_FT2_IS_FONT_CLASS(klass)   (G_TYPE_CHECK_CLASS_TYPE ((klass), PANGO_TYPE_FT2_FONT))
//  #define PANGO_FT2_FONT_GET_CLASS(obj)    (G_TYPE_INSTANCE_GET_CLASS ((obj), PANGO_TYPE_FT2_FONT, PangoFT2FontClass))

// func _pango_ft2_font_new (PangoFT2FontMap *ft2fontmap,
// 			  FcPattern       *pattern) *PangoFT2Font{
//    PangoFontMap *fontmap = PANGO_FONT_MAP (ft2fontmap);
//    PangoFT2Font *ft2font;
//    double d;

//    g_return_val_if_fail (fontmap != NULL, NULL);
//    g_return_val_if_fail (pattern != NULL, NULL);

//    ft2font = (PangoFT2Font *)g_object_new (PANGO_TYPE_FT2_FONT,
// 					   "pattern", pattern,
// 					   "fontmap", fontmap,
// 					   NULL);

//    if (FcPatternGetDouble (pattern, FC_PIXEL_SIZE, 0, &d) == FcResultMatch)
// 	 ft2font.size = d*PANGO_SCALE;

//    return ft2font;
//  }

//  func load_fallback_face (PangoFT2Font *ft2font,
// 			 const char   *original_file) {
//    PangoFcFont *fcfont = PANGO_FC_FONT (ft2font);
//    FcPattern *sans;
//    FcPattern *matched;
//    FcResult result;
//    FT_Error error;
//    FcChar8 *filename2 = NULL;
//    gchar *name;
//    int id;

//    sans = FcPatternBuild (NULL,
// 			  FC_FAMILY,     FcTypeString, "sans",
// 			  FC_PIXEL_SIZE, FcTypeDouble, (double)ft2font.size / PANGO_SCALE,
// 			  NULL);

//    _pango_ft2_font_map_default_substitute ((PangoFcFontMap *)fcfont.fontmap, sans);

//    matched = FcFontMatch (pango_fc_font_map_get_config ((PangoFcFontMap *)fcfont.fontmap), sans, &result);

//    if (FcPatternGetString (matched, FC_FILE, 0, &filename2) != FcResultMatch)
// 	 goto bail1;

//    if (FcPatternGetInteger (matched, FC_INDEX, 0, &id) != FcResultMatch)
// 	 goto bail1;

//    error = FT_New_Face (_pango_ft2_font_map_get_library (fcfont.fontmap),
// 				(char *) filename2, id, &ft2font.face);

//    if (error)
// 	 {
// 	 bail1:
// 	   name = pango_font_description_to_string (fcfont.description);
// 	   g_error ("Unable to open font file %s for font %s, exiting\n", filename2, name);
// 	 }
//    else
// 	 {
// 	   name = pango_font_description_to_string (fcfont.description);
// 	   g_warning ("Unable to open font file %s for font %s, falling back to %s\n", original_file, name, filename2);
// 	   g_free (name);
// 	 }

//    FcPatternDestroy (sans);
//    FcPatternDestroy (matched);
//  }

// func set_transform (PangoFT2Font *ft2font) {
//    PangoFcFont *fcfont = (PangoFcFont *)ft2font;
//    FcMatrix *fc_matrix;

//    if (FcPatternGetMatrix (fcfont.font_pattern, FC_MATRIX, 0, &fc_matrix) == FcResultMatch)
// 	 {
// 	   FT_Matrix ft_matrix;

// 	   ft_matrix.xx = 0x10000L * fc_matrix.xx;
// 	   ft_matrix.yy = 0x10000L * fc_matrix.yy;
// 	   ft_matrix.xy = 0x10000L * fc_matrix.xy;
// 	   ft_matrix.yx = 0x10000L * fc_matrix.yx;

// 	   FT_Set_Transform (ft2font.face, &ft_matrix, NULL);
// 	 }
//  }

//  /**
//   * pango_ft2_font_get_face: (skip)
//   * @font: a #PangoFont
//   *
//   * Returns the native FreeType2 `FT_Face` structure used for this #PangoFont.
//   * This may be useful if you want to use FreeType2 functions directly.
//   *
//   * Use pango_fc_font_lock_face() instead; when you are done with a
//   * face from pango_fc_font_lock_face() you must call
//   * pango_fc_font_unlock_face().
//   *
//   * Return value: (nullable): a pointer to a `FT_Face` structure, with the
//   *   size set correctly, or %NULL if @font is %NULL.
//   **/
// func  pango_ft2_font_get_face (PangoFont *font)  FT_Face  {
//    PangoFT2Font *ft2font = (PangoFT2Font *)font;
//    PangoFcFont *fcfont = (PangoFcFont *)font;
//    FT_Error error;
//    FcPattern *pattern;
//    FcChar8 *filename;
//    FcBool antialias, hinting, autohint;
//    int hintstyle;
//    int id;

//    if (G_UNLIKELY (!font))
// 	 return NULL;

//    pattern = fcfont.font_pattern;

//    if (!ft2font.face)
// 	 {
// 	   ft2font.load_flags = 0;

// 	   /* disable antialiasing if requested */
// 	   if (FcPatternGetBool (pattern,
// 				 FC_ANTIALIAS, 0, &antialias) != FcResultMatch)
// 	 antialias = FcTrue;

// 	   if (antialias)
// 	 ft2font.load_flags |= FT_LOAD_NO_BITMAP;
// 	   else
// 	 ft2font.load_flags |= FT_LOAD_TARGET_MONO;

// 	   /* disable hinting if requested */
// 	   if (FcPatternGetBool (pattern,
// 				 FC_HINTING, 0, &hinting) != FcResultMatch)
// 	 hinting = FcTrue;

//  #ifdef FC_HINT_STYLE
// 	   if (FcPatternGetInteger (pattern, FC_HINT_STYLE, 0, &hintstyle) != FcResultMatch)
// 	 hintstyle = FC_HINT_FULL;

// 	   if (!hinting || hintstyle == FC_HINT_NONE)
// 		   ft2font.load_flags |= FT_LOAD_NO_HINTING;

// 	   switch (hintstyle) {
// 	   case FC_HINT_SLIGHT:
// 	   case FC_HINT_MEDIUM:
// 	 ft2font.load_flags |= FT_LOAD_TARGET_LIGHT;
// 	 break;
// 	   default:
// 	 ft2font.load_flags |= FT_LOAD_TARGET_NORMAL;
// 	 break;
// 	   }
//  #else
// 	   if (!hinting)
// 		   ft2font.load_flags |= FT_LOAD_NO_HINTING;
//  #endif

// 	   /* force autohinting if requested */
// 	   if (FcPatternGetBool (pattern,
// 				 FC_AUTOHINT, 0, &autohint) != FcResultMatch)
// 	 autohint = FcFalse;

// 	   if (autohint)
// 	 ft2font.load_flags |= FT_LOAD_FORCE_AUTOHINT;

// 	   if (FcPatternGetString (pattern, FC_FILE, 0, &filename) != FcResultMatch)
// 		   goto bail0;

// 	   if (FcPatternGetInteger (pattern, FC_INDEX, 0, &id) != FcResultMatch)
// 		   goto bail0;

// 	   error = FT_New_Face (_pango_ft2_font_map_get_library (fcfont.fontmap),
// 				(char *) filename, id, &ft2font.face);
// 	   if (error != FT_Err_Ok)
// 	 {
// 	 bail0:
// 	   load_fallback_face (ft2font, (char *) filename);
// 	 }

// 	   g_assert (ft2font.face);

// 	   set_transform (ft2font);

// 	   error = FT_Set_Char_Size (ft2font.face,
// 				 PANGO_PIXELS_26_6 (ft2font.size),
// 				 PANGO_PIXELS_26_6 (ft2font.size),
// 				 0, 0);
// 	   if (error)
// 	 g_warning ("Error in FT_Set_Char_Size: %d", error);
// 	 }

//    return ft2font.face;
//  }

func pango_ft2_font_init(PangoFT2Font *ft2font) {
	ft2font.face = NULL

	ft2font.size = 0

	ft2font.glyph_info = g_hash_table_new(NULL, NULL)
}

//  static void
//  pango_ft2_font_class_init (PangoFT2FontClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoFontClass *font_class = PANGO_FONT_CLASS (class);
//    PangoFcFontClass *fc_font_class = PANGO_FC_FONT_CLASS (class);

//    object_class.finalize = pango_ft2_font_finalize;

//    font_class.get_glyph_extents = pango_ft2_font_get_glyph_extents;

//    fc_font_class.lock_face = pango_ft2_font_real_lock_face;
//    fc_font_class.unlock_face = pango_ft2_font_real_unlock_face;
//  }

func (ft2font *PangoFT2Font) pango_ft2_font_get_glyph_info(PangoGlyph glyph,
	gboolean create) *PangoFT2GlyphInfo {
	//    PangoFT2Font *ft2font = (PangoFT2Font *)font;
	//    PangoFcFont *fcfont = (PangoFcFont *)font;
	PangoFT2GlyphInfo * info

	info = g_hash_table_lookup(ft2font.glyph_info, GUINT_TO_POINTER(glyph))

	if (info == NULL) && create {
		info = g_slice_new0(PangoFT2GlyphInfo)

		pango_fc_font_get_raw_extents(fcfont,
			glyph,
			&info.ink_rect,
			&info.logical_rect)

		g_hash_table_insert(ft2font.glyph_info, GUINT_TO_POINTER(glyph), info)
	}

	return info
}

func (font *PangoFT2Font) GetGlyphExtents(glyph pango.Glyph, inkRect, logicalRect *pango.Rectangle) {
	empty := false

	if glyph == pango.PANGO_GLYPH_EMPTY {
		glyph = font.getGlyph(' ')
		empty = true
	}

	if glyph&pango.PANGO_GLYPH_UNKNOWN_FLAG != 0 {
		PangoFontMetrics * metrics = pango_font_get_metrics(font, NULL)

		if metrics {
			if ink_rect {
				ink_rect.x = PANGO_SCALE
				ink_rect.width = metrics.approximate_char_width - 2*PANGO_SCALE
				ink_rect.y = -(metrics.ascent - PANGO_SCALE)
				ink_rect.height = metrics.ascent + metrics.descent - 2*PANGO_SCALE
			}
			if logical_rect {
				logical_rect.x = 0
				logical_rect.width = metrics.approximate_char_width
				logical_rect.y = -metrics.ascent
				logical_rect.height = metrics.ascent + metrics.descent
			}

			pango_font_metrics_unref(metrics)
		} else {
			if ink_rect {
				ink_rect.x, ink_rect.y, ink_rect.height, ink_rect.width = 0, 0, 0, 0
			}
			if logical_rect {
				logical_rect.x, logical_rect.y, logical_rect.height, logical_rect.width = 0, 0, 0, 0
			}
		}
		return
	}

	info = pango_ft2_font_get_glyph_info(font, glyph, true)

	if ink_rect {
		*ink_rect = info.ink_rect
	}
	if logical_rect {
		*logical_rect = info.logical_rect
	}

	if empty {
		if ink_rect {
			ink_rect.x, ink_rect.y, ink_rect.height, ink_rect.width = 0, 0, 0, 0
		}
		if logical_rect {
			logical_rect.x, logical_rect.width = 0, 0
		}
		return
	}
}

//  /**
//   * pango_ft2_font_get_kerning:
//   * @font: a #PangoFont
//   * @left: the left #PangoGlyph
//   * @right: the right #PangoGlyph
//   *
//   * Retrieves kerning information for a combination of two glyphs.
//   *
//   * Use pango_fc_font_kern_glyphs() instead.
//   *
//   * Return value: The amount of kerning (in Pango units) to apply for
//   * the given combination of glyphs.
//   **/
//  int
//  pango_ft2_font_get_kerning (PangoFont *font,
// 				 PangoGlyph left,
// 				 PangoGlyph right)
//  {
//    PangoFcFont *fc_font = PANGO_FC_FONT (font);

//    FT_Face face;
//    FT_Error error;
//    FT_Vector kerning;

//    face = pango_fc_font_lock_face (fc_font);
//    if (!face)
// 	 return 0;

//    if (!FT_HAS_KERNING (face))
// 	 {
// 	   pango_fc_font_unlock_face (fc_font);
// 	   return 0;
// 	 }

//    error = FT_Get_Kerning (face, left, right, ft_kerning_default, &kerning);
//    if (error != FT_Err_Ok)
// 	 {
// 	   pango_fc_font_unlock_face (fc_font);
// 	   return 0;
// 	 }

//    pango_fc_font_unlock_face (fc_font);
//    return PANGO_UNITS_26_6 (kerning.x);
//  }

//  static FT_Face
//  pango_ft2_font_real_lock_face (font *PangoFcFont)
//  {
//    return pango_ft2_font_get_face ((PangoFont *)font);
//  }

//  static void
//  pango_ft2_font_real_unlock_face (font *PangoFcFont G_GNUC_UNUSED)
//  {
//  }

//  /* Utility functions */

//  /**
//   * pango_ft2_get_unknown_glyph:
//   * @font: a #PangoFont
//   *
//   * Return the index of a glyph suitable for drawing unknown characters with
//   * @font, or %PANGO_GLYPH_EMPTY if no suitable glyph found.
//   *
//   * If you want to draw an unknown-box for a character that is not covered
//   * by the font,
//   * use AsUnknownGlyph() instead.
//   *
//   * Return value: a glyph index into @font, or %PANGO_GLYPH_EMPTY
//   **/
//  PangoGlyph
//  pango_ft2_get_unknown_glyph (PangoFont *font)
//  {
//    FT_Face face = pango_ft2_font_get_face (font);
//    if (face && FT_IS_SFNT (face))
// 	 /* TrueType fonts have an 'unknown glyph' box on glyph index 0 */
// 	 return 0;
//    else
// 	 return PANGO_GLYPH_EMPTY;
//  }

//  void *
//  _pango_ft2_font_get_cache_glyph_data (PangoFont *font,
// 					  int        glyph_index)
//  {
//    PangoFT2GlyphInfo *info;

//    if (!PANGO_FT2_IS_FONT (font))
// 	 return NULL;

//    info = pango_ft2_font_get_glyph_info (font, glyph_index, false);

//    if (info == NULL)
// 	 return NULL;

//    return info.cached_glyph;
//  }

//  void
//  _pango_ft2_font_set_cache_glyph_data (PangoFont     *font,
// 					  int            glyph_index,
// 					  void          *cached_glyph)
//  {
//    PangoFT2GlyphInfo *info;

//    if (!PANGO_FT2_IS_FONT (font))
// 	 return;

//    info = pango_ft2_font_get_glyph_info (font, glyph_index, true);

//    info.cached_glyph = cached_glyph;

//    /* TODO: Implement limiting of the number of cached glyphs */
//  }

//  void
//  _pango_ft2_font_set_glyph_cache_destroy (PangoFont      *font,
// 					  GDestroyNotify  destroy_notify)
//  {
//    if (!PANGO_FT2_IS_FONT (font))
// 	 return;

//    PANGO_FT2_FONT (font).glyph_cache_destroy = destroy_notify;
//  }

// PangoFcFont is a base class for font implementations
// using the Fontconfig and FreeType libraries and is used in
// conjunction with `PangoFcFontMap`. When deriving from this
// class, you need to implement all of its virtual functions
// other than shutdown() along with the GetGlyphExtents()
// virtual function from `PangoFont`.
type PangoFcFont struct {
	parent_instance pango.Font

	font_pattern fc.Pattern      // fully resolved pattern
	fontmap      *PangoFcFontMap // associated map
	priv         FcFontPrivate   // used internally
	matrix       pango.Matrix    // used internally
	description  pango.FontDescription

	metrics_by_lang []interface{}

	is_hinted      bool //  = 1;
	is_transformed bool //  = 1;
}

func (font *PangoFcFont) Describe(absolute bool) pango.FontDescription {
	if !absolute {
		return font.description
	}

	desc := font.description

	size, ok := font.font_pattern.GetFloat(fc.FC_PIXEL_SIZE)
	if ok {
		desc.SetAbsoluteSize(int(size * float64(pango.PangoScale)))
	}

	return desc
}

func (font *PangoFcFont) GetCoverage(_ pango.Language) pango.Coverage {
	if font.priv.decoder != nil {
		charset := font.priv.decoder.GetCharset(font)
		return fromCharset(charset)
	}

	if font.fontmap == nil {
		return coverage{}
	}

	data := font.fontmap.getFontFaceData(font.font_pattern)
	if data == nil {
		return coverage{}
	}

	if data.coverage == nil {
		// Pull the coverage out of the pattern, this doesn't require loading the font
		charset, _ := font.font_pattern.GetCharset(fc.FC_CHARSET)
		data.coverage = fromCharset(charset) // stores it into the map
	}

	return data.coverage
}

func (font *PangoFcFont) GetFontMap() pango.FontMap { return font.fontmap }

// getGlyph gets the glyph index for a given Unicode character
// for `font`. If you only want to determine
// whether the font has the glyph, use pango_fc_font_has_char().
// It returns 0 if the Unicode character doesn't exist in the font.
func (font *PangoFcFont) getGlyph(wc rune) pango.Glyph {
	/* Replace NBSP with a normal space; it should be invariant that
	* they shape the same other than breaking properties. */
	if wc == 0xA0 {
		wc = 0x20
	}

	if font.priv.decoder != nil {
		return font.priv.decoder.GetGlyph(font, wc)
	}

	hb_font := font.GetHBFont()
	glyph := pango.AsUnknownGlyph(wc)

	glyph, _ = pango.HbFontGetNominalGlyph(hb_font, wc)

	return glyph

}

//  #define PANGO_FC_TYPE_FAMILY            (pango_fc_family_get_type ())
//  #define PANGO_FC_FAMILY(object)         (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FAMILY, PangoFcFamily))
//  #define PANGO_FC_IS_FAMILY(object)      (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FAMILY))

//  #define PANGO_FC_TYPE_FACE              (pango_fc_face_get_type ())
//  #define PANGO_FC_FACE(object)           (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FACE, PangoFcFace))
//  #define PANGO_FC_IS_FACE(object)        (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FACE))

//  #define PANGO_FC_TYPE_FONTSET           (pango_fc_fontset_get_type ())
//  #define PANGO_FC_FONTSET(object)        (G_TYPE_CHECK_INSTANCE_CAST ((object), PANGO_FC_TYPE_FONTSET, PangoFcFontset))
//  #define PANGO_FC_IS_FONTSET(object)     (G_TYPE_CHECK_INSTANCE_TYPE ((object), PANGO_FC_TYPE_FONTSET))

//  enum {
//    PROP_0,
//    PROP_PATTERN,
//    PROP_FONTMAP
//  };

//  typedef struct _PangoFcFontPrivate PangoFcFontPrivate;

//  struct _PangoFcFontPrivate
//  {
//    PangoFcDecoder *decoder;
//    PangoFcFontKey *key;
//  };

//  #define PANGO_FC_FONT_LOCK_FACE(font)	(PANGO_FC_FONT_GET_CLASS (font).lock_face (font))
//  #define PANGO_FC_FONT_UNLOCK_FACE(font)	(PANGO_FC_FONT_GET_CLASS (font).unlock_face (font))

//  G_DEFINE_ABSTRACT_TYPE_WITH_CODE (PangoFcFont, pango_fc_font, PANGO_TYPE_FONT,
// 								   G_ADD_PRIVATE (PangoFcFont))

//  static void
//  pango_fc_font_class_init (PangoFcFontClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoFontClass *font_class = PANGO_FONT_CLASS (class);

//    class.has_char = pango_fc_font_real_has_char;
//    class.get_glyph = pango_fc_font_real_get_glyph;
//    class.get_unknown_glyph = NULL;

//    object_class.finalize = pango_fc_font_finalize;
//    object_class.set_property = pango_fc_font_set_property;
//    object_class.get_property = pango_fc_font_get_property;
//    font_class.describe = pango_fc_font_describe;
//    font_class.describe_absolute = pango_fc_font_describe_absolute;
//    font_class.GetCoverage = GetCoverage;
//    font_class.GetMetrics = pango_fc_font_get_metrics;
//    font_class.GetFontMap = pango_fc_font_get_font_map;
//    font_class.GetFeatures = pango_fc_font_get_features;
//    font_class.CreateHBFont = pango_fc_font_create_hb_font;
//    font_class.GetFeatures = pango_fc_font_get_features;

//    g_object_class_install_property (object_class, PROP_PATTERN,
// 					g_param_spec_pointer ("pattern",
// 							  "Pattern",
// 							  "The fontconfig pattern for this font",
// 							  G_PARAM_READWRITE | G_PARAM_CONSTRUCT_ONLY |
// 							  G_PARAM_STATIC_STRINGS));
//    g_object_class_install_property (object_class, PROP_FONTMAP,
// 					g_param_spec_object ("fontmap",
// 							 "Font Map",
// 							 "The PangoFc font map this font is associated with (Since: 1.26)",
// 							 PANGO_TYPE_FC_FONT_MAP,
// 							 G_PARAM_READWRITE |
// 							 G_PARAM_STATIC_STRINGS));
//  }

//  static void
//  pango_fc_font_init (PangoFcFont *fcfont)
//  {
//    fcfont.priv = pango_fc_font_get_instance_private (fcfont);
//  }

//  static void
//  free_metrics_info (PangoFcMetricsInfo *info)
//  {
//    pango_font_metrics_unref (info.metrics);
//    g_slice_free (PangoFcMetricsInfo, info);
//  }

//  static void
//  pango_fc_font_finalize (GObject *object)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (object);
//    PangoFcFontPrivate *priv = fcfont.priv;
//    PangoFcFontMap *fontmap;

//    g_slist_foreach (fcfont.metrics_by_lang, (GFunc)free_metrics_info, NULL);
//    g_slist_free (fcfont.metrics_by_lang);

//    fontmap = g_weak_ref_get ((GWeakRef *) &fcfont.fontmap);
//    if (fontmap)
// 	 {
// 	   _pango_fc_font_map_remove (PANGO_FC_FONT_MAP (fcfont.fontmap), fcfont);
// 	   g_weak_ref_clear ((GWeakRef *) &fcfont.fontmap);
// 	   g_object_unref (fontmap);
// 	 }

//    FcPatternDestroy (fcfont.font_pattern);
//    pango_font_description_free (fcfont.description);

//    if (priv.decoder)
// 	 _pango_fc_font_set_decoder (fcfont, NULL);

//    G_OBJECT_CLASS (pango_fc_font_parent_class).finalize (object);
//  }

//  static gboolean
//  pattern_is_hinted (FcPattern *pattern)
//  {
//    FcBool hinting;

//    if (FcPatternGetBool (pattern,
// 			 FC_HINTING, 0, &hinting) != FcResultMatch)
// 	 hinting = FcTrue;

//    return hinting;
//  }

//  static gboolean
//  pattern_is_transformed (FcPattern *pattern)
//  {
//    FcMatrix *fc_matrix;

//    if (FcPatternGetMatrix (pattern, FC_MATRIX, 0, &fc_matrix) == FcResultMatch)
// 	 {
// 	   return fc_matrix.xx != 1 || fc_matrix.xy != 0 ||
// 			  fc_matrix.yx != 0 || fc_matrix.yy != 1;
// 	 }
//    else
// 	 return false;
//  }

//  static void
//  pango_fc_font_set_property (GObject       *object,
// 				 guint          prop_id,
// 				 const GValue  *value,
// 				 GParamSpec    *pspec)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (object);

//    switch (prop_id)
// 	 {
// 	 case PROP_PATTERN:
// 	   {
// 	 FcPattern *pattern = g_value_get_pointer (value);

// 	 g_return_if_fail (pattern != NULL);
// 	 g_return_if_fail (fcfont.font_pattern == NULL);

// 	 FcPatternReference (pattern);
// 	 fcfont.font_pattern = pattern;
// 	 fcfont.description = pango_fc_font_description_from_pattern (pattern, true);
// 	 fcfont.is_hinted = pattern_is_hinted (pattern);
// 	 fcfont.is_transformed = pattern_is_transformed (pattern);
// 	   }
// 	   goto set_decoder;

// 	 case PROP_FONTMAP:
// 	   {
// 	 PangoFcFontMap *fcfontmap = PANGO_FC_FONT_MAP (g_value_get_object (value));

// 	 g_return_if_fail (fcfont.fontmap == NULL);
// 	 g_weak_ref_set ((GWeakRef *) &fcfont.fontmap, fcfontmap);
// 	   }
// 	   goto set_decoder;

// 	 default:
// 	   G_OBJECT_WARN_INVALID_PROPERTY_ID (object, prop_id, pspec);
// 	   return;
// 	 }

//  set_decoder:
//    /* set decoder if both pattern and fontmap are set now */
//    if (fcfont.font_pattern && fcfont.fontmap)
// 	 _pango_fc_font_set_decoder (fcfont,
// 				 pango_fc_font_map_find_decoder  ((PangoFcFontMap *) fcfont.fontmap,
// 								  fcfont.font_pattern));
//  }

//  static void
//  pango_fc_font_get_property (GObject       *object,
// 				 guint          prop_id,
// 				 GValue        *value,
// 				 GParamSpec    *pspec)
//  {
//    switch (prop_id)
// 	 {
// 	 case PROP_PATTERN:
// 	   {
// 	 PangoFcFont *fcfont = PANGO_FC_FONT (object);
// 	 g_value_set_pointer (value, fcfont.font_pattern);
// 	   }
// 	   break;
// 	 case PROP_FONTMAP:
// 	   {
// 	 PangoFcFont *fcfont = PANGO_FC_FONT (object);
// 	 PangoFontMap *fontmap = g_weak_ref_get ((GWeakRef *) &fcfont.fontmap);
// 	 g_value_take_object (value, fontmap);
// 	   }
// 	   break;
// 	 default:
// 	   G_OBJECT_WARN_INVALID_PROPERTY_ID (object, prop_id, pspec);
// 	   break;
// 	 }
//  }

//  /* For Xft, it would be slightly more efficient to simply to
//   * call Xft, and also more robust against changes in Xft.
//   * But for now, we simply use the same code for all backends.
//   *
//   * The code in this function is partly based on code from Xft,
//   * Copyright 2000 Keith Packard
//   */
//  static void
//  get_face_metrics (PangoFcFont      *fcfont,
// 		   PangoFontMetrics *metrics)
//  {
//    hb_font_t *hb_font = GetHBFont (PANGO_FONT (fcfont));
//    hb_font_extents_t extents;

//    FcMatrix *fc_matrix;
//    gboolean have_transform = false;

//    hb_font_get_extents_for_direction (hb_font, HB_DIRECTION_LTR, &extents);

//    if  (FcPatternGetMatrix (fcfont.font_pattern,
// 				FC_MATRIX, 0, &fc_matrix) == FcResultMatch)
// 	 {
// 	   have_transform = (fc_matrix.xx != 1 || fc_matrix.xy != 0 ||
// 			 fc_matrix.yx != 0 || fc_matrix.yy != 1);
// 	 }

//    if (have_transform)
// 	 {
// 	   metrics.descent =  - extents.descender * fc_matrix.yy;
// 	   metrics.ascent = extents.ascender * fc_matrix.yy;
// 	   metrics.height = (extents.ascender - extents.descender + extents.line_gap) * fc_matrix.yy;
// 	 }
//    else
// 	 {
// 	   metrics.descent = - extents.descender;
// 	   metrics.ascent = extents.ascender;
// 	   metrics.height = extents.ascender - extents.descender + extents.line_gap;
// 	 }

//    metrics.underline_thickness = PANGO_SCALE;
//    metrics.underline_position = - PANGO_SCALE;
//    metrics.strikethrough_thickness = PANGO_SCALE;
//    metrics.strikethrough_position = metrics.ascent / 2;

//    /* FIXME: use the right hb version */
//  #if HB_VERSION_ATLEAST(2,5,4)
//    hb_position_t position;

//    if (hb_ot_metrics_get_position (hb_font, HB_OT_METRICS_TAG_UNDERLINE_SIZE, &position))
// 	 metrics.underline_thickness = position;

//    if (hb_ot_metrics_get_position (hb_font, HB_OT_METRICS_TAG_UNDERLINE_OFFSET, &position))
// 	 metrics.underline_position = position;

//    if (hb_ot_metrics_get_position (hb_font, HB_OT_METRICS_TAG_STRIKEOUT_SIZE, &position))
// 	 metrics.strikethrough_thickness = position;

//    if (hb_ot_metrics_get_position (hb_font, HB_OT_METRICS_TAG_STRIKEOUT_OFFSET, &position))
// 	 metrics.strikethrough_position = position;
//  #endif
//  }

//  PangoFontMetrics *
//  pango_fc_font_create_base_metrics_for_context (PangoFcFont   *fcfont,
// 							PangoContext  *context)
//  {
//    PangoFontMetrics *metrics;
//    metrics = pango_font_metrics_new ();

//    get_face_metrics (fcfont, metrics);

//    return metrics;
//  }

//  static int
//  max_glyph_width (PangoLayout *layout)
//  {
//    int max_width = 0;
//    GSList *l, *r;

//    for (l = pango_layout_get_lines_readonly (layout); l; l = l.next)
// 	 {
// 	   PangoLayoutLine *line = l.data;

// 	   for (r = line.runs; r; r = r.next)
// 	 {
// 	   PangoGlyphString *glyphs = ((PangoGlyphItem *)r.data).glyphs;
// 	   int i;

// 	   for (i = 0; i < glyphs.num_glyphs; i++)
// 		 if (glyphs.glyphs[i].geometry.width > max_width)
// 		   max_width = glyphs.glyphs[i].geometry.width;
// 	 }
// 	 }

//    return max_width;
//  }

//  static PangoFontMetrics *
//  pango_fc_font_get_metrics (PangoFont     *font,
// 				PangoLanguage *language)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (font);
//    PangoFcMetricsInfo *info = NULL; /* Quiet gcc */
//    GSList *tmp_list;
//    static int in_get_metrics;

//    const char *sample_str = pango_language_get_sample_string (language);

//    tmp_list = fcfont.metrics_by_lang;
//    while (tmp_list)
// 	 {
// 	   info = tmp_list.data;

// 	   if (info.sample_str == sample_str)    /* We _don't_ need strcmp */
// 	 break;

// 	   tmp_list = tmp_list.next;
// 	 }

//    if (!tmp_list)
// 	 {
// 	   PangoFontMap *fontmap;
// 	   PangoContext *context;

// 	   fontmap = g_weak_ref_get ((GWeakRef *) &fcfont.fontmap);
// 	   if (!fontmap)
// 	 return pango_font_metrics_new ();

// 	   info = g_slice_new0 (PangoFcMetricsInfo);

// 	   /* Note: we need to add info to the list before calling
// 		* into PangoLayout below, to prevent recursion
// 		*/
// 	   fcfont.metrics_by_lang = g_slist_prepend (fcfont.metrics_by_lang,
// 						  info);

// 	   info.sample_str = sample_str;

// 	   context = pango_font_map_create_context (fontmap);
// 	   pango_context_set_language (context, language);

// 	   info.metrics = pango_fc_font_create_base_metrics_for_context (fcfont, context);

// 	   if (!in_get_metrics)
// 		 {
// 		   /* Compute derived metrics */
// 		   PangoLayout *layout;
// 		   PangoRectangle extents;
// 		   const char *sample_str = pango_language_get_sample_string (language);
// 		   PangoFontDescription *desc = pango_font_describe_with_absolute_size (font);
// 		   gulong sample_str_width;

// 		   in_get_metrics = 1;

// 		   layout = pango_layout_new (context);
// 		   pango_layout_set_font_description (layout, desc);
// 		   pango_font_description_free (desc);

// 		   pango_layout_set_text (layout, sample_str, -1);
// 		   pango_layout_get_extents (layout, NULL, &extents);

// 		   sample_str_width = pango_utf8_strwidth (sample_str);
// 		   g_assert (sample_str_width > 0);
// 		   info.metrics.approximate_char_width = extents.width / sample_str_width;

// 		   pango_layout_set_text (layout, "0123456789", -1);
// 		   info.metrics.approximate_digit_width = max_glyph_width (layout);

// 		   g_object_unref (layout);

// 		   in_get_metrics = 0;
// 		 }

// 	   g_object_unref (context);
// 	   g_object_unref (fontmap);
// 	 }

//    return pango_font_metrics_ref (info.metrics);
//  }

//  static PangoFontMap *
//  pango_fc_font_get_font_map (font *PangoFcFont)
//  {
//    PangoFcFont *fcfont = PANGO_FC_FONT (font);

//    /* MT-unsafe.  Oh well...  The API is unsafe. */
//    return fcfont.fontmap;
//  }

//  static gboolean
//  pango_fc_font_real_has_char (font *PangoFcFont,
// 				  gunichar     wc)
//  {
//    FcCharSet *charset;

//    if (FcPatternGetCharSet (font.font_pattern,
// 				FC_CHARSET, 0, &charset) != FcResultMatch)
// 	 return false;

//    return FcCharSetHasChar (charset, wc);
//  }

//  /**
//   * pango_fc_font_lock_face: (skip)
//   * @font: a #PangoFcFont.
//   *
//   * Gets the FreeType `FT_Face` associated with a font,
//   * This face will be kept around until you call
//   * pango_fc_font_unlock_face().
//   *
//   * Return value: the FreeType `FT_Face` associated with @font.
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use GetHBFont() instead
//   **/
//  FT_Face
//  pango_fc_font_lock_face (font *PangoFcFont)
//  {
//    g_return_val_if_fail (PANGO_IS_FC_FONT (font), NULL);

//    return PANGO_FC_FONT_LOCK_FACE (font);
//  }

//  /**
//   * pango_fc_font_unlock_face:
//   * @font: a #PangoFcFont.
//   *
//   * Releases a font previously obtained with
//   * pango_fc_font_lock_face().
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use GetHBFont() instead
//   **/
//  void
//  pango_fc_font_unlock_face (font *PangoFcFont)
//  {
//    g_return_if_fail (PANGO_IS_FC_FONT (font));

//    PANGO_FC_FONT_UNLOCK_FACE (font);
//  }

//  /**
//   * pango_fc_font_has_char:
//   * @font: a #PangoFcFont
//   * @wc: Unicode codepoint to look up
//   *
//   * Determines whether @font has a glyph for the codepoint @wc.
//   *
//   * Return value: %true if @font has the requested codepoint.
//   *
//   * Since: 1.4
//   * Deprecated: 1.44: Use pango_font_has_char()
//   **/
//  gboolean
//  pango_fc_font_has_char (font *PangoFcFont,
// 			 gunichar     wc)
//  {
//    PangoFcFontPrivate *priv = font.priv;
//    FcCharSet *charset;

//    g_return_val_if_fail (PANGO_IS_FC_FONT (font), false);

//    if (priv.decoder)
// 	 {
// 	   charset = pango_fc_decoder_get_charset (priv.decoder, font);
// 	   return FcCharSetHasChar (charset, wc);
// 	 }

//    return PANGO_FC_FONT_GET_CLASS (font).has_char (font, wc);
//  }

//  /**
//   * pango_fc_font_get_unknown_glyph:
//   * @font: a #PangoFcFont
//   * @wc: the Unicode character for which a glyph is needed.
//   *
//   * Returns the index of a glyph suitable for drawing @wc as an
//   * unknown character.
//   *
//   * Use AsUnknownGlyph() instead.
//   *
//   * Return value: a glyph index into @font.
//   *
//   * Since: 1.4
//   **/
//  PangoGlyph
//  pango_fc_font_get_unknown_glyph (font *PangoFcFont,
// 				  gunichar     wc)
//  {
//    if (font && PANGO_FC_FONT_GET_CLASS (font).get_unknown_glyph)
// 	 return PANGO_FC_FONT_GET_CLASS (font).get_unknown_glyph (font, wc);

//    return AsUnknownGlyph (wc);
//  }

//  void
//  _pango_fc_font_shutdown (font *PangoFcFont)
//  {
//    g_return_if_fail (PANGO_IS_FC_FONT (font));

//    if (PANGO_FC_FONT_GET_CLASS (font).shutdown)
// 	 PANGO_FC_FONT_GET_CLASS (font).shutdown (font);
//  }

//  /**
//   * pango_fc_font_kern_glyphs:
//   * @font: a #PangoFcFont
//   * @glyphs: a #PangoGlyphString
//   *
//   * This function used to adjust each adjacent pair of glyphs
//   * in @glyphs according to kerning information in @font.
//   *
//   * Since 1.44, it does nothing.
//   *
//   *
//   * Since: 1.4
//   * Deprecated: 1.32
//   **/
//  void
//  pango_fc_font_kern_glyphs (PangoFcFont      *font,
// 				PangoGlyphString *glyphs)
//  {
//  }

//  /**
//   * _pango_fc_font_get_decoder:
//   * @font: a #PangoFcFont
//   *
//   * This will return any custom decoder set on this font.
//   *
//   * Return value: The custom decoder
//   *
//   * Since: 1.6
//   **/

//  PangoFcDecoder *
//  _pango_fc_font_get_decoder (font *PangoFcFont)
//  {
//    PangoFcFontPrivate *priv = font.priv;

//    return priv.decoder;
//  }

//  /**
//   * _pango_fc_font_set_decoder:
//   * @font: a #PangoFcFont
//   * @decoder: a #PangoFcDecoder to set for this font
//   *
//   * This sets a custom decoder for this font.  Any previous decoder
//   * will be released before this one is set.
//   *
//   * Since: 1.6
//   **/

//  void
//  _pango_fc_font_set_decoder (PangoFcFont    *font,
// 				 PangoFcDecoder *decoder)
//  {
//    PangoFcFontPrivate *priv = font.priv;

//    if (priv.decoder)
// 	 g_object_unref (priv.decoder);

//    priv.decoder = decoder;

//    if (priv.decoder)
// 	 g_object_ref (priv.decoder);
//  }

//  PangoFcFontKey *
//  _pango_fc_font_get_font_key (PangoFcFont *fcfont)
//  {
//    PangoFcFontPrivate *priv = fcfont.priv;

//    return priv.key;
//  }

//  void
//  _pango_fc_font_set_font_key (PangoFcFont    *fcfont,
// 				  PangoFcFontKey *key)
//  {
//    PangoFcFontPrivate *priv = fcfont.priv;

//    priv.key = key;
//  }

//  /**
//   * pango_fc_font_get_raw_extents:
//   * @fcfont: a #PangoFcFont
//   * @glyph: the glyph index to load
//   * @ink_rect: (out) (optional): location to store ink extents of the
//   *   glyph, or %NULL
//   * @logical_rect: (out) (optional): location to store logical extents
//   *   of the glyph or %NULL
//   *
//   * Gets the extents of a single glyph from a font. The extents are in
//   * user space; that is, they are not transformed by any matrix in effect
//   * for the font.
//   *
//   * Long term, this functionality probably belongs in the default
//   * implementation of the GetGlyphExtents() virtual function.
//   * The other possibility would be to to make it public in something
//   * like it's current form, and also expose glyph information
//   * caching functionality similar to pango_ft2_font_set_glyph_info().
//   *
//   * Since: 1.6
//   **/
//  void
//  pango_fc_font_get_raw_extents (PangoFcFont    *fcfont,
// 					PangoGlyph      glyph,
// 					PangoRectangle *ink_rect,
// 					PangoRectangle *logical_rect)
//  {
//    g_return_if_fail (PANGO_IS_FC_FONT (fcfont));

//    if (glyph == PANGO_GLYPH_EMPTY)
// 	 {
// 	   if (ink_rect)
// 	 {
// 	   ink_rect.x = 0;
// 	   ink_rect.width = 0;
// 	   ink_rect.y = 0;
// 	   ink_rect.height = 0;
// 	 }

// 	   if (logical_rect)
// 	 {
// 	   logical_rect.x = 0;
// 	   logical_rect.width = 0;
// 	   logical_rect.y = 0;
// 	   logical_rect.height = 0;
// 	 }
// 	 }
//    else
// 	 {
// 	   hb_font_t *hb_font = GetHBFont (PANGO_FONT (fcfont));
// 	   hb_glyph_extents_t extents;
// 	   hb_font_extents_t font_extents;

// 	   hb_font_get_glyph_extents (hb_font, glyph, &extents);
// 	   hb_font_get_extents_for_direction (hb_font, HB_DIRECTION_LTR, &font_extents);

// 	   if (ink_rect)
// 	 {
// 	   ink_rect.x = extents.x_bearing;
// 	   ink_rect.width = extents.width;
// 	   ink_rect.y = -extents.y_bearing;
// 	   ink_rect.height = -extents.height;
// 	 }

// 	   if (logical_rect)
// 	 {
// 		   hb_position_t x, y;

// 		   hb_font_get_glyph_advance_for_direction (hb_font,
// 													glyph,
// 													HB_DIRECTION_LTR,
// 													&x, &y);

// 	   logical_rect.x = 0;
// 	   logical_rect.width = x;
// 	   logical_rect.y = - font_extents.ascender;
// 	   logical_rect.height = font_extents.ascender - font_extents.descender;
// 	 }
// 	 }
//  }

//  static void
//  pango_fc_font_get_features (PangoFont    *font,
// 							 hb_feature_t *features,
// 							 guint         len,
// 							 guint        *num_features)
//  {
//    /* Setup features from fontconfig pattern. */
//    PangoFcFont *fc_font = PANGO_FC_FONT (font);
//    if (fc_font.font_pattern)
// 	 {
// 	   char *s;
// 	   while (*num_features < len &&
// 			  FcResultMatch == FcPatternGetString (fc_font.font_pattern,
// 												   PANGO_FC_FONT_FEATURES,
// 												   *num_features,
// 												   (FcChar8 **) &s))
// 		 {
// 		   gboolean ret = hb_feature_from_string (s, -1, &features[*num_features]);
// 		   features[*num_features].start = 0;
// 		   features[*num_features].end   = (unsigned int) -1;
// 		   if (ret)
// 			 (*num_features)++;
// 		 }
// 	 }
//  }

//  extern gpointer get_gravity_class (void);

//  static PangoGravity
//  pango_fc_font_key_get_gravity (PangoFcFontKey *key)
//  {
//    const FcPattern *pattern;
//    PangoGravity gravity = PANGO_GRAVITY_SOUTH;
//    FcChar8 *s;

//    pattern = pango_fc_font_key_get_pattern (key);
//    if (FcPatternGetString (pattern, fcGravity, 0, (FcChar8 **)&s) == FcResultMatch)
// 	 {
// 	   GEnumValue *value = g_enum_get_value_by_nick (get_gravity_class (), (char *)s);
// 	   gravity = value.value;
// 	 }

//    return gravity;
//  }

//  static double
//  get_font_size (PangoFcFontKey *key)
//  {
//    const FcPattern *pattern;
//    double size;
//    double dpi;

//    pattern = pango_fc_font_key_get_pattern (key);
//    if (FcPatternGetDouble (pattern, FC_PIXEL_SIZE, 0, &size) == FcResultMatch)
// 	 return size;

//    /* Just in case FC_PIXEL_SIZE got unset between pango_fc_make_pattern()
// 	* and here.  That would be very weird.
// 	*/

//    if (FcPatternGetDouble (pattern, FC_DPI, 0, &dpi) != FcResultMatch)
// 	 dpi = 72;

//    if (FcPatternGetDouble (pattern, FC_SIZE, 0, &size) == FcResultMatch)
// 	 return size * dpi / 72.;

//    /* Whatever */
//    return 18.;
//  }

//  static void
//  parse_variations (const char            *variations,
// 				   hb_ot_var_axis_info_t *axes,
// 				   int                    n_axes,
// 				   float                 *coords)
//  {
//    const char *p;
//    const char *end;
//    hb_variation_t var;
//    int i;

//    p = variations;
//    while (p && *p)
// 	 {
// 	   end = strchr (p, ',');
// 	   if (hb_variation_from_string (p, end ? end - p: -1, &var))
// 		 {
// 		   for (i = 0; i < n_axes; i++)
// 			 {
// 			   if (axes[i].tag == var.tag)
// 				 {
// 				   coords[axes[i].axis_index] = var.value;
// 				   break;
// 				 }
// 			 }
// 		 }

// 	   p = end ? end + 1 : NULL;
// 	 }
//  }

//  static hb_font_t *
//  pango_fc_font_create_hb_font (font *PangoFcFont)
//  {
//    PangoFcFont *fc_font = PANGO_FC_FONT (font);
//    PangoFcFontKey *key;
//    hb_face_t *hb_face;
//    hb_font_t *hb_font;
//    double x_scale_inv, y_scale_inv;
//    double x_scale, y_scale;
//    double size;

//    x_scale_inv = y_scale_inv = 1.0;
//    size = 1.0;

//    key = _pango_fc_font_get_font_key (fc_font);
//    if (key)
// 	 {
// 	   const FcPattern *pattern = pango_fc_font_key_get_pattern (key);
// 	   const PangoMatrix *matrix;
// 	   PangoMatrix matrix2;
// 	   PangoGravity gravity;
// 	   FcMatrix fc_matrix, *fc_matrix_val;
// 	   double x, y;
// 	   int i;

// 	   matrix = pango_fc_font_key_get_matrix (key);
// 	   pango_matrix_get_font_scale_factors (matrix, &x_scale_inv, &y_scale_inv);

// 	   FcMatrixInit (&fc_matrix);
// 	   for (i = 0; FcPatternGetMatrix (pattern, FC_MATRIX, i, &fc_matrix_val) == FcResultMatch; i++)
// 		 FcMatrixMultiply (&fc_matrix, &fc_matrix, fc_matrix_val);

// 	   matrix2.xx = fc_matrix.xx;
// 	   matrix2.yx = fc_matrix.yx;
// 	   matrix2.xy = fc_matrix.xy;
// 	   matrix2.yy = fc_matrix.yy;
// 	   pango_matrix_get_font_scale_factors (&matrix2, &x, &y);

// 	   x_scale_inv /= x;
// 	   y_scale_inv /= y;

// 	   gravity = pango_fc_font_key_get_gravity (key);
// 	   if (PANGO_GRAVITY_IS_IMPROPER (gravity))
// 		 {
// 		   x_scale_inv = -x_scale_inv;
// 		   y_scale_inv = -y_scale_inv;
// 		 }
// 	   size = get_font_size (key);
// 	 }

//    x_scale = 1. / x_scale_inv;
//    y_scale = 1. / y_scale_inv;

//    hb_face = pango_fc_font_map_get_hb_face (PANGO_FC_FONT_MAP (fc_font.fontmap), fc_font);

//    hb_font = hb_font_create (hb_face);
//    hb_font_set_scale (hb_font,
// 					  size * PANGO_SCALE * x_scale,
// 					  size * PANGO_SCALE * y_scale);

//    if (key)
// 	 {
// 	   const FcPattern *pattern = pango_fc_font_key_get_pattern (key);
// 	   const char *variations;
// 	   int index;
// 	   unsigned int n_axes;
// 	   hb_ot_var_axis_info_t *axes;
// 	   float *coords;
// 	   int i;

// 	   n_axes = hb_ot_var_get_axis_infos (hb_face, 0, NULL, NULL);
// 	   if (n_axes == 0)
// 		 goto done;

// 	   axes = g_new0 (hb_ot_var_axis_info_t, n_axes);
// 	   coords = g_new (float, n_axes);

// 	   hb_ot_var_get_axis_infos (hb_face, 0, &n_axes, axes);
// 	   for (i = 0; i < n_axes; i++)
// 		 coords[axes[i].axis_index] = axes[i].default_value;

// 	   if (FcPatternGetInteger (pattern, FC_INDEX, 0, &index) == FcResultMatch &&
// 		   index != 0)
// 		 {
// 		   unsigned int instance = (index >> 16) - 1;
// 		   hb_ot_var_named_instance_get_design_coords (hb_face, instance, &n_axes, coords);
// 		 }

// 	   if (FcPatternGetString (pattern, fcFontVariations, 0, (FcChar8 **)&variations) == FcResultMatch)
// 		 parse_variations (variations, axes, n_axes, coords);

// 	   variations = pango_fc_font_key_get_variations (key);
// 	   if (variations)
// 		 parse_variations (variations, axes, n_axes, coords);

// 	   hb_font_set_var_coords_design (hb_font, coords, n_axes);

// 	   g_free (coords);
// 	   g_free (axes);
// 	 }

//  done:
//    return hb_font;
//  }

//  /**
//   * pango_fc_font_get_languages:
//   * @font: a #PangoFcFont
//   *
//   * Returns the languages that are supported by @font.
//   *
//   * This corresponds to the FC_LANG member of the FcPattern.
//   *
//   * The returned array is only valid as long as the font
//   * and its fontmap are valid.
//   *
//   * Returns: (transfer none) (nullable): a %NULL-terminated
//   *    array of PangoLanguage*
//   *
//   * Since: 1.48
//   */
//  PangoLanguage **
//  pango_fc_font_get_languages (font *PangoFcFont)
//  {
//    PangoFcFontMap *fontmap;
//    PangoLanguage **languages;

//    fontmap = g_weak_ref_get ((GWeakRef *) &font.fontmap);
//    if (!fontmap)
// 	 return NULL;

//    languages  = _pango_fc_font_map_get_languages (fontmap, font);
//    g_object_unref (fontmap);

//    return languages;
//  }

//  /**
//   * pango_fc_font_get_pattern: (skip)
//   * @font: a #PangoFcFont
//   *
//   * Returns the FcPattern that @font is based on.
//   *
//   * Returns: the fontconfig pattern for this font
//   *
//   * Since: 1.48
//   */
//  FcPattern *
//  pango_fc_font_get_pattern (font *PangoFcFont)
//  {
//    return font.font_pattern;
//  }
