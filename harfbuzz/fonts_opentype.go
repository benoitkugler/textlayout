package harfbuzz

// ported from harfbuzz/src/hb-ot-metrics.cc, bh-ot-font.cc Copyright © 2018-2019  Ebrahim Byagowi,  Behdad Esfahbod, Roozbeh Pournader

/**
 * SECTION:hb-ot-metrics
 * @title: hb-ot-metrics
 * @short_description: OpenType Metrics
 * @include: hb-ot.h
 *
 * Functions for fetching metrics from fonts.
 **/

//  static float
//  _fix_ascender_descender (float value, hb_ot_metrics_tag_t metrics_tag)
//  {
//    if (metrics_tag == HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER ||
// 	   metrics_tag == HB_OT_METRICS_TAG_VERTICAL_ASCENDER)
// 	 return fabs ((double) value);
//    if (metrics_tag == HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER ||
// 	   metrics_tag == HB_OT_METRICS_TAG_VERTICAL_DESCENDER)
// 	 return -fabs ((double) value);
//    return value;
//  }

//  /* The common part of _get_position logic needed on hb-ot-font and here
// 	to be able to have slim builds without the not always needed parts */
//  bool
//  _hb_ot_metrics_get_position_common (Font           *font,
// 					 hb_ot_metrics_tag_t  metrics_tag,
// 					 Position       *position     /* OUT.  May be NULL. */)
//  {
//    Face *face = font->face;
//    switch ((unsigned) metrics_tag)
//    {
//  #ifndef HB_NO_VAR
//  #define GET_VAR face->table.MVAR->get_var (metrics_tag, font->coords, font->num_coords)
//  #else
//  #define GET_VAR .0f
//  #endif
//  #define GET_METRIC_X(TABLE, ATTR) \
//    (face->table.TABLE->has_data () && \
// 	 (position && (*position = font->em_scalef_x (_fix_ascender_descender ( \
// 	   face->table.TABLE->ATTR + GET_VAR, metrics_tag))), true))
//  #define GET_METRIC_Y(TABLE, ATTR) \
//    (face->table.TABLE->has_data () && \
// 	 (position && (*position = font->em_scalef_y (_fix_ascender_descender ( \
// 	   face->table.TABLE->ATTR + GET_VAR, metrics_tag))), true))
//    case HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER:
// 	 return (face->table.OS2->use_typo_metrics () && GET_METRIC_Y (OS2, sTypoAscender)) ||
// 		GET_METRIC_Y (hhea, ascender);
//    case HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER:
// 	 return (face->table.OS2->use_typo_metrics () && GET_METRIC_Y (OS2, sTypoDescender)) ||
// 		GET_METRIC_Y (hhea, descender);
//    case HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP:
// 	 return (face->table.OS2->use_typo_metrics () && GET_METRIC_Y (OS2, sTypoLineGap)) ||
// 		GET_METRIC_Y (hhea, lineGap);
//    case HB_OT_METRICS_TAG_VERTICAL_ASCENDER:  return GET_METRIC_X (vhea, ascender);
//    case HB_OT_METRICS_TAG_VERTICAL_DESCENDER: return GET_METRIC_X (vhea, descender);
//    case HB_OT_METRICS_TAG_VERTICAL_LINE_GAP:  return GET_METRIC_X (vhea, lineGap);
//  #undef GET_METRIC_Y
//  #undef GET_METRIC_X
//  #undef GET_VAR
//    default:                               assert (0); return false;
//    }
//  }

//  #ifndef HB_NO_METRICS

//  #if 0
//  static bool
//  _get_gasp (Face *face, float *result, hb_ot_metrics_tag_t metrics_tag)
//  {
//    const OT::GaspRange& range = face->table.gasp->get_gasp_range (metrics_tag - HB_TAG ('g','s','p','0'));
//    if (&range == &Null (OT::GaspRange)) return false;
//    if (result) *result = range.rangeMaxPPEM + font->face->table.MVAR->get_var (metrics_tag, font->coords, font->num_coords);
//    return true;
//  }
//  #endif

//  /* Private tags for https://github.com/harfbuzz/harfbuzz/issues/1866 */
//  #define _HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER_OS2   HB_TAG ('O','a','s','c')
//  #define _HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER_HHEA  HB_TAG ('H','a','s','c')
//  #define _HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER_OS2  HB_TAG ('O','d','s','c')
//  #define _HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER_HHEA HB_TAG ('H','d','s','c')
//  #define _HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP_OS2   HB_TAG ('O','l','g','p')
//  #define _HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP_HHEA  HB_TAG ('H','l','g','p')

//  /**
//   * hb_ot_metrics_get_position:
//   * @font: an #Font object.
//   * @metrics_tag: tag of metrics value you like to fetch.
//   * @position: (out) (optional): result of metrics value from the font.
//   *
//   * Fetches metrics value corresponding to @metrics_tag from @font.
//   *
//   * Returns: Whether found the requested metrics in the font.
//   * Since: 2.6.0
//   **/
//  hb_bool_t
//  hb_ot_metrics_get_position (Font           *font,
// 				 hb_ot_metrics_tag_t  metrics_tag,
// 				 Position       *position     /* OUT.  May be NULL. */)
//  {
//    Face *face = font->face;
//    switch ((unsigned) metrics_tag)
//    {
//    case HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER:
//    case HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER:
//    case HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP:
//    case HB_OT_METRICS_TAG_VERTICAL_ASCENDER:
//    case HB_OT_METRICS_TAG_VERTICAL_DESCENDER:
//    case HB_OT_METRICS_TAG_VERTICAL_LINE_GAP:           return _hb_ot_metrics_get_position_common (font, metrics_tag, position);
//  #ifndef HB_NO_VAR
//  #define GET_VAR hb_ot_metrics_get_variation (font, metrics_tag)
//  #else
//  #define GET_VAR 0
//  #endif
//  #define GET_METRIC_X(TABLE, ATTR) \
//    (face->table.TABLE->has_data () && \
// 	 (position && (*position = font->em_scalef_x (face->table.TABLE->ATTR + GET_VAR)), true))
//  #define GET_METRIC_Y(TABLE, ATTR) \
//    (face->table.TABLE->has_data () && \
// 	 (position && (*position = font->em_scalef_y (face->table.TABLE->ATTR + GET_VAR)), true))
//    case HB_OT_METRICS_TAG_HORIZONTAL_CLIPPING_ASCENT:  return GET_METRIC_Y (OS2, usWinAscent);
//    case HB_OT_METRICS_TAG_HORIZONTAL_CLIPPING_DESCENT: return GET_METRIC_Y (OS2, usWinDescent);
//    case HB_OT_METRICS_TAG_HORIZONTAL_CARET_RISE:       return GET_METRIC_Y (hhea, caretSlopeRise);
//    case HB_OT_METRICS_TAG_HORIZONTAL_CARET_RUN:        return GET_METRIC_X (hhea, caretSlopeRun);
//    case HB_OT_METRICS_TAG_HORIZONTAL_CARET_OFFSET:     return GET_METRIC_X (hhea, caretOffset);
//    case HB_OT_METRICS_TAG_VERTICAL_CARET_RISE:         return GET_METRIC_X (vhea, caretSlopeRise);
//    case HB_OT_METRICS_TAG_VERTICAL_CARET_RUN:          return GET_METRIC_Y (vhea, caretSlopeRun);
//    case HB_OT_METRICS_TAG_VERTICAL_CARET_OFFSET:       return GET_METRIC_Y (vhea, caretOffset);
//    case HB_OT_METRICS_TAG_X_HEIGHT:                    return GET_METRIC_Y (OS2->v2 (), sxHeight);
//    case HB_OT_METRICS_TAG_CAP_HEIGHT:                  return GET_METRIC_Y (OS2->v2 (), sCapHeight);
//    case HB_OT_METRICS_TAG_SUBSCRIPT_EM_X_SIZE:         return GET_METRIC_X (OS2, ySubscriptXSize);
//    case HB_OT_METRICS_TAG_SUBSCRIPT_EM_Y_SIZE:         return GET_METRIC_Y (OS2, ySubscriptYSize);
//    case HB_OT_METRICS_TAG_SUBSCRIPT_EM_X_OFFSET:       return GET_METRIC_X (OS2, ySubscriptXOffset);
//    case HB_OT_METRICS_TAG_SUBSCRIPT_EM_Y_OFFSET:       return GET_METRIC_Y (OS2, ySubscriptYOffset);
//    case HB_OT_METRICS_TAG_SUPERSCRIPT_EM_X_SIZE:       return GET_METRIC_X (OS2, ySuperscriptXSize);
//    case HB_OT_METRICS_TAG_SUPERSCRIPT_EM_Y_SIZE:       return GET_METRIC_Y (OS2, ySuperscriptYSize);
//    case HB_OT_METRICS_TAG_SUPERSCRIPT_EM_X_OFFSET:     return GET_METRIC_X (OS2, ySuperscriptXOffset);
//    case HB_OT_METRICS_TAG_SUPERSCRIPT_EM_Y_OFFSET:     return GET_METRIC_Y (OS2, ySuperscriptYOffset);
//    case HB_OT_METRICS_TAG_STRIKEOUT_SIZE:              return GET_METRIC_Y (OS2, yStrikeoutSize);
//    case HB_OT_METRICS_TAG_STRIKEOUT_OFFSET:            return GET_METRIC_Y (OS2, yStrikeoutPosition);
//    case HB_OT_METRICS_TAG_UNDERLINE_SIZE:              return GET_METRIC_Y (post->table, underlineThickness);
//    case HB_OT_METRICS_TAG_UNDERLINE_OFFSET:            return GET_METRIC_Y (post->table, underlinePosition);

//    /* Private tags */
//    case _HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER_OS2:    return GET_METRIC_Y (OS2, sTypoAscender);
//    case _HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER_HHEA:   return GET_METRIC_Y (hhea, ascender);
//    case _HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER_OS2:   return GET_METRIC_Y (OS2, sTypoDescender);
//    case _HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER_HHEA:  return GET_METRIC_Y (hhea, descender);
//    case _HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP_OS2:    return GET_METRIC_Y (OS2, sTypoLineGap);
//    case _HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP_HHEA:   return GET_METRIC_Y (hhea, lineGap);
//  #undef GET_METRIC_Y
//  #undef GET_METRIC_X
//  #undef GET_VAR
//    default:                                        return false;
//    }
//  }

//  #ifndef HB_NO_VAR
//  /**
//   * hb_ot_metrics_get_variation:
//   * @font: an #Font object.
//   * @metrics_tag: tag of metrics value you like to fetch.
//   *
//   * Fetches metrics value corresponding to @metrics_tag from @font with the
//   * current font variation settings applied.
//   *
//   * Returns: The requested metric value.
//   *
//   * Since: 2.6.0
//   **/
//  float
//  hb_ot_metrics_get_variation (Font *font, hb_ot_metrics_tag_t metrics_tag)
//  {
//    return font->face->table.MVAR->get_var (metrics_tag, font->coords, font->num_coords);
//  }

//  /**
//   * hb_ot_metrics_get_x_variation:
//   * @font: an #Font object.
//   * @metrics_tag: tag of metrics value you like to fetch.
//   *
//   * Fetches horizontal metrics value corresponding to @metrics_tag from @font
//   * with the current font variation settings applied.
//   *
//   * Returns: The requested metric value.
//   *
//   * Since: 2.6.0
//   **/
//  Position
//  hb_ot_metrics_get_x_variation (Font *font, hb_ot_metrics_tag_t metrics_tag)
//  {
//    return font->em_scalef_x (hb_ot_metrics_get_variation (font, metrics_tag));
//  }

//  /**
//   * hb_ot_metrics_get_y_variation:
//   * @font: an #Font object.
//   * @metrics_tag: tag of metrics value you like to fetch.
//   *
//   * Fetches vertical metrics value corresponding to @metrics_tag from @font with
//   * the current font variation settings applied.
//   *
//   * Returns: The requested metric value.
//   *
//   * Since: 2.6.0
//   **/
//  Position
//  hb_ot_metrics_get_y_variation (Font *font, hb_ot_metrics_tag_t metrics_tag)
//  {
//    return font->em_scalef_y (hb_ot_metrics_get_variation (font, metrics_tag));
//  }

/**
 * SECTION:hb-ot-font
 * @title: hb-ot-font
 * @short_description: OpenType font implementation
 * @include: hb-ot.h
 *
 * Functions for using OpenType fonts with Shape().  Note that fonts returned
 * by hb_font_create() default to using these functions, so most clients would
 * never need to call these functions directly.
 **/

// func () hb_ot_get_nominal_glyph (unicode rune) (fonts.GlyphID, bool) {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;
//    return ot_face->cmap->get_nominal_glyph (unicode, glyph);
//  }

//  static unsigned int
//  hb_ot_get_nominal_glyphs (Font *font HB_UNUSED,
// 			   void *font_data,
// 			   unsigned int count,
// 			   const hb_codepoint_t *first_unicode,
// 			   unsigned int unicode_stride,
// 			   hb_codepoint_t *first_glyph,
// 			   unsigned int glyph_stride,
// 			   void *user_data HB_UNUSED)
//  {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;
//    return ot_face->cmap->get_nominal_glyphs (count,
// 						 first_unicode, unicode_stride,
// 						 first_glyph, glyph_stride);
//  }

//  static hb_bool_t
//  hb_ot_get_variation_glyph (Font *font HB_UNUSED,
// 				void *font_data,
// 				hb_codepoint_t unicode,
// 				hb_codepoint_t variation_selector,
// 				hb_codepoint_t *glyph,
// 				void *user_data HB_UNUSED)
//  {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;
//    return ot_face->cmap->get_variation_glyph (unicode, variation_selector, glyph);
//  }

 static void
 hb_ot_get_glyph_h_advances (Font* font, void* font_data,
				 unsigned count,
				 const hb_codepoint_t *first_glyph,
				 unsigned glyph_stride,
				 Position *first_advance,
				 unsigned advance_stride,
				 void *user_data HB_UNUSED)
 {
   const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;
   const OT::hmtx_accelerator_t &hmtx = *ot_face->hmtx;

   for (unsigned int i = 0; i < count; i++)
   {
	 *first_advance = font->em_scale_x (hmtx.get_advance (*first_glyph, font));
	 first_glyph = &StructAtOffsetUnaligned<hb_codepoint_t> (first_glyph, glyph_stride);
	 first_advance = &StructAtOffsetUnaligned<Position> (first_advance, advance_stride);
   }
 }

//  static void
//  hb_ot_get_glyph_v_advances (Font* font, void* font_data,
// 				 unsigned count,
// 				 const hb_codepoint_t *first_glyph,
// 				 unsigned glyph_stride,
// 				 Position *first_advance,
// 				 unsigned advance_stride,
// 				 void *user_data HB_UNUSED)
//  {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;
//    const OT::vmtx_accelerator_t &vmtx = *ot_face->vmtx;

//    for (unsigned int i = 0; i < count; i++)
//    {
// 	 *first_advance = font->em_scale_y (-(int) vmtx.get_advance (*first_glyph, font));
// 	 first_glyph = &StructAtOffsetUnaligned<hb_codepoint_t> (first_glyph, glyph_stride);
// 	 first_advance = &StructAtOffsetUnaligned<Position> (first_advance, advance_stride);
//    }
//  }

//  static hb_bool_t
//  hb_ot_get_glyph_v_origin (Font *font,
// 			   void *font_data,
// 			   hb_codepoint_t glyph,
// 			   Position *x,
// 			   Position *y,
// 			   void *user_data HB_UNUSED)
//  {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;

//    *x = font->GetGlyphHAdvance (glyph) / 2;

//  #ifndef HB_NO_OT_FONT_CFF
//    const OT::VORG &VORG = *ot_face->VORG;
//    if (VORG.has_data ())
//    {
// 	 *y = font->em_scale_y (VORG.get_y_origin (glyph));
// 	 return true;
//    }
//  #endif

//    GlyphExtents extents = {0};
//    if (ot_face->glyf->get_extents (font, glyph, &extents))
//    {
// 	 const OT::vmtx_accelerator_t &vmtx = *ot_face->vmtx;
// 	 Position tsb = vmtx.get_side_bearing (font, glyph);
// 	 *y = extents.y_bearing + font->em_scale_y (tsb);
// 	 return true;
//    }

//    hb_font_extents_t font_extents;
//    font->get_h_extents_with_fallback (&font_extents);
//    *y = font_extents.ascender;

//    return true;
//  }

//  static hb_bool_t
//  hb_ot_get_glyph_extents (Font *font,
// 			  void *font_data,
// 			  hb_codepoint_t glyph,
// 			  GlyphExtents *extents,
// 			  void *user_data HB_UNUSED)
//  {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;

//  #if !defined(HB_NO_OT_FONT_BITMAP) && !defined(HB_NO_COLOR)
//    if (ot_face->sbix->get_extents (font, glyph, extents)) return true;
//  #endif
//    if (ot_face->glyf->get_extents (font, glyph, extents)) return true;
//  #ifndef HB_NO_OT_FONT_CFF
//    if (ot_face->cff1->get_extents (font, glyph, extents)) return true;
//    if (ot_face->cff2->get_extents (font, glyph, extents)) return true;
//  #endif
//  #if !defined(HB_NO_OT_FONT_BITMAP) && !defined(HB_NO_COLOR)
//    if (ot_face->CBDT->get_extents (font, glyph, extents)) return true;
//  #endif

//    // TODO Hook up side-bearings variations.
//    return false;
//  }

//  #ifndef HB_NO_OT_FONT_GLYPH_NAMES
//  static hb_bool_t
//  hb_ot_get_glyph_name (Font *font HB_UNUSED,
// 			   void *font_data,
// 			   hb_codepoint_t glyph,
// 			   char *name, unsigned int size,
// 			   void *user_data HB_UNUSED)
//  {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;
//    if (ot_face->post->get_glyph_name (glyph, name, size)) return true;
//  #ifndef HB_NO_OT_FONT_CFF
//    if (ot_face->cff1->get_glyph_name (glyph, name, size)) return true;
//  #endif
//    return false;
//  }
//  static hb_bool_t
//  hb_ot_get_glyph_from_name (Font *font HB_UNUSED,
// 				void *font_data,
// 				const char *name, int len,
// 				hb_codepoint_t *glyph,
// 				void *user_data HB_UNUSED)
//  {
//    const hb_ot_face_t *ot_face = (const hb_ot_face_t *) font_data;
//    if (ot_face->post->get_glyph_from_name (name, len, glyph)) return true;
//  #ifndef HB_NO_OT_FONT_CFF
// 	 if (ot_face->cff1->get_glyph_from_name (name, len, glyph)) return true;
//  #endif
//    return false;
//  }
//  #endif

//  static hb_bool_t
//  hb_ot_get_font_h_extents (Font *font,
// 			   void *font_data HB_UNUSED,
// 			   hb_font_extents_t *metrics,
// 			   void *user_data HB_UNUSED)
//  {
//    return _hb_ot_metrics_get_position_common (font, HB_OT_METRICS_TAG_HORIZONTAL_ASCENDER, &metrics->ascender) &&
// 	  _hb_ot_metrics_get_position_common (font, HB_OT_METRICS_TAG_HORIZONTAL_DESCENDER, &metrics->descender) &&
// 	  _hb_ot_metrics_get_position_common (font, HB_OT_METRICS_TAG_HORIZONTAL_LINE_GAP, &metrics->line_gap);
//  }

//  static hb_bool_t
//  hb_ot_get_font_v_extents (Font *font,
// 			   void *font_data HB_UNUSED,
// 			   hb_font_extents_t *metrics,
// 			   void *user_data HB_UNUSED)
//  {
//    return _hb_ot_metrics_get_position_common (font, HB_OT_METRICS_TAG_VERTICAL_ASCENDER, &metrics->ascender) &&
// 	  _hb_ot_metrics_get_position_common (font, HB_OT_METRICS_TAG_VERTICAL_DESCENDER, &metrics->descender) &&
// 	  _hb_ot_metrics_get_position_common (font, HB_OT_METRICS_TAG_VERTICAL_LINE_GAP, &metrics->line_gap);
//  }

//  #if HB_USE_ATEXIT
//  static void free_static_ot_funcs ();
//  #endif

//  static struct hb_ot_font_funcs_lazy_loader_t : hb_font_funcs_lazy_loader_t<hb_ot_font_funcs_lazy_loader_t>
//  {
//    static hb_font_funcs_t *create ()
//    {
// 	 hb_font_funcs_t *funcs = hb_font_funcs_create ();

// 	 hb_font_funcs_set_font_h_extents_func (funcs, hb_ot_get_font_h_extents, nullptr, nullptr);
// 	 hb_font_funcs_set_font_v_extents_func (funcs, hb_ot_get_font_v_extents, nullptr, nullptr);
// 	 hb_font_funcs_set_nominal_glyph_func (funcs, hb_ot_get_nominal_glyph, nullptr, nullptr);
// 	 hb_font_funcs_set_nominal_glyphs_func (funcs, hb_ot_get_nominal_glyphs, nullptr, nullptr);
// 	 hb_font_funcs_set_variation_glyph_func (funcs, hb_ot_get_variation_glyph, nullptr, nullptr);
// 	 hb_font_funcs_set_glyph_h_advances_func (funcs, hb_ot_get_glyph_h_advances, nullptr, nullptr);
// 	 hb_font_funcs_set_glyph_v_advances_func (funcs, hb_ot_get_glyph_v_advances, nullptr, nullptr);
// 	 //hb_font_funcs_set_glyph_h_origin_func (funcs, hb_ot_get_glyph_h_origin, nullptr, nullptr);
// 	 hb_font_funcs_set_glyph_v_origin_func (funcs, hb_ot_get_glyph_v_origin, nullptr, nullptr);
// 	 hb_font_funcs_set_glyph_extents_func (funcs, hb_ot_get_glyph_extents, nullptr, nullptr);
// 	 //hb_font_funcs_set_glyph_contour_point_func (funcs, hb_ot_get_glyph_contour_point, nullptr, nullptr);
//  #ifndef HB_NO_OT_FONT_GLYPH_NAMES
// 	 hb_font_funcs_set_glyph_name_func (funcs, hb_ot_get_glyph_name, nullptr, nullptr);
// 	 hb_font_funcs_set_glyph_from_name_func (funcs, hb_ot_get_glyph_from_name, nullptr, nullptr);
//  #endif

// 	 hb_font_funcs_make_immutable (funcs);

//  #if HB_USE_ATEXIT
// 	 atexit (free_static_ot_funcs);
//  #endif

// 	 return funcs;
//    }
//  } static_ot_funcs;

//  #if HB_USE_ATEXIT
//  static
//  void free_static_ot_funcs ()
//  {
//    static_ot_funcs.free_instance ();
//  }
//  #endif

//  static hb_font_funcs_t *
//  _hb_ot_get_font_funcs ()
//  {
//    return static_ot_funcs.get_unconst ();
//  }

//  /**
//   * hb_ot_font_set_funcs:
//   * @font: #Font to work upon
//   *
//   * Sets the font functions to use when working with @font.
//   *
//   * Since: 0.9.28
//   **/
//  void
//  hb_ot_font_set_funcs (Font *font)
//  {
//    hb_font_set_funcs (font,
// 			  _hb_ot_get_font_funcs (),
// 			  &font->face->table,
// 			  nullptr);
//  }

//  int
//  _glyf_get_side_bearing_var (Font *font, hb_codepoint_t glyph, bool is_vertical)
//  {
//    return font->face->table.glyf->get_side_bearing_var (font, glyph, is_vertical);
//  }

//  unsigned
//  _glyf_get_advance_var (Font *font, hb_codepoint_t glyph, bool is_vertical)
//  {
//    return font->face->table.glyf->get_advance_var (font, glyph, is_vertical);
//  }
