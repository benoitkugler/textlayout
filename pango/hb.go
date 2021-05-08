package pango

// TODO:
func pango_hb_shape(font Font, item_text []rune, analysis *Analysis, glyphs *GlyphString, paragraph_text []rune) {
	// PangoHbShapeContext context = { 0, };
	// BufferFlags hb_buffer_flags;
	// hb_font_t *hb_font;
	// hb_buffer_t *hb_buffer;
	// Direction hb_direction;
	// gboolean free_buffer;
	// GlyphInfo *hb_glyph;
	// hb_glyph_position_t *hb_position;
	// int last_cluster;
	// guint i, num_glyphs;
	// unsigned int item_offset = item_text - paragraph_text;
	// hb_feature_t features[32];
	// unsigned int num_features = 0;
	// PangoGlyphInfo *infos;

	// context.show_flags = find_show_flags(analysis)
	// hb_font = pango_font_get_hb_font_for_context(font, &context)
	// hb_buffer := acquire_buffer(&free_buffer) // TODO: cache

	// hb_direction := HB_DIRECTION_LTR
	// if PANGO_GRAVITY_IS_VERTICAL(analysis.gravity) {
	// 	hb_direction = HB_DIRECTION_TTB
	// }
	// if analysis.level % 2 {
	// 	hb_direction = HB_DIRECTION_REVERSE(hb_direction)
	// }
	// if PANGO_GRAVITY_IS_IMPROPER(analysis.gravity) {
	// 	hb_direction = HB_DIRECTION_REVERSE(hb_direction)
	// }

	// hb_buffer_flags = HB_BUFFER_FLAG_BOT | HB_BUFFER_FLAG_EOT

	// if context.show_flags & PANGO_SHOW_IGNORABLES {
	// 	hb_buffer_flags |= HB_BUFFER_FLAG_PRESERVE_DEFAULT_IGNORABLES
	// }

	// /* setup buffer */

	// hb_buffer_set_direction(hb_buffer, hb_direction)
	// hb_buffer_set_script(hb_buffer, g_unicode_script_to_iso15924(analysis.script))
	// hb_buffer_set_language(hb_buffer, hb_language_from_string(pango_language_to_string(analysis.language), -1))
	// hb_buffer_setCluster_level(hb_buffer, HB_BUFFER_CLUSTER_LEVEL_MONOTONE_CHARACTERS)
	// hb_buffer_set_flags(hb_buffer, hb_buffer_flags)
	// hb_buffer_set_invisible_glyph(hb_buffer, PANGO_GLYPH_EMPTY)

	// // use AddRunes
	// hb_buffer_add_utf8(hb_buffer, paragraph_text, paragraph_length, item_offset, item_length)
	// if analysis.flags & PANGO_ANALYSIS_FLAG_NEED_HYPHEN {
	// 	/* Insert either a Unicode or ASCII hyphen. We may
	// 	 * want to look for script-specific hyphens here.  */
	// 	p := paragraph_text + item_offset + item_length
	// 	// int last_char_len = p - g_utf8_prev_char (p);
	// 	// hb_codepoint_t glyph;

	// 	if hb_font_get_nominal_glyph(hb_font, 0x2010, &glyph) {
	// 		AddRune(hb_buffer, 0x2010, item_offset+item_length-last_char_len)
	// 	} else if hb_font_get_nominal_glyph(hb_font, '-', &glyph) {
	// 		AddRune(hb_buffer, '-', item_offset+item_length-last_char_len)
	// 	}
	// }

	// pango_font_get_features(font, features, G_N_ELEMENTS(features), &num_features)
	// apply_extra_attributes(analysis.extra_attrs, features, G_N_ELEMENTS(features), &num_features)

	// Shape(hb_font, hb_buffer, features, num_features)

	// if PANGO_GRAVITY_IS_IMPROPER(analysis.gravity) {
	// 	hb_buffer_reverse(hb_buffer)
	// }

	// /* buffer output */
	// num_glyphs = hb_buffer_get_length(hb_buffer)
	// hb_glyph = hb_buffer_get_glyph_infos(hb_buffer, NULL)
	// pango_glyph_string_set_size(glyphs, num_glyphs)
	// infos = glyphs.glyphs
	// last_cluster = -1
	// for i = 0; i < num_glyphs; i++ {
	// 	infos[i].glyph = hb_glyph.Codepoint
	// 	glyphs.log_clusters[i] = hb_glyph.cluster - item_offset
	// 	infos[i].attr.is_cluster_start = glyphs.log_clusters[i] != last_cluster
	// 	hb_glyph++
	// 	last_cluster = glyphs.log_clusters[i]
	// }

	// hb_position = hb_buffer_get_glyph_positions(hb_buffer, NULL)
	// if PANGO_GRAVITY_IS_VERTICAL(analysis.gravity) {
	// 	for i = 0; i < num_glyphs; i++ {
	// 		/* 90 degrees rotation counter-clockwise. */
	// 		infos[i].geometry.width = hb_position.y_advance
	// 		infos[i].geometry.XOffset = hb_position.y_offset
	// 		infos[i].geometry.y_offset = -hb_position.XOffset
	// 		hb_position++
	// 	}
	// } else /* horizontal */ {
	// 	for i = 0; i < num_glyphs; i++ {
	// 		infos[i].geometry.width = hb_position.XAdvance
	// 		infos[i].geometry.XOffset = hb_position.XOffset
	// 		infos[i].geometry.y_offset = -hb_position.y_offset
	// 		hb_position++
	// 	}
	// }
}
