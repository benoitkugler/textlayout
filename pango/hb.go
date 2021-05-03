package pango

import "github.com/benoitkugler/textlayout/fonts/truetype"

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

type Hb_font_t struct {
	//   hb_object_header_t header;

	parent *Hb_font_t
	// face *HB_face_t

	x_scale int32
	y_scale int32
	x_mult  int64
	y_mult  int64

	x_ppem uint
	y_ppem uint

	ptem float32

	/* Font variation coordinates. */
	//   unsigned int num_coords;
	coords        []int
	design_coords []float32
}

func HB_font_create(*HB_face_t) *Hb_font_t {
	return nil
}

func HB_font_set_scale(font *Hb_font_t, x, y float64) {
}

func HB_ot_var_get_axis_infos(*HB_face_t) []truetype.VarAxis {
	return nil
}

func HB_ot_var_named_instance_get_design_coords(*HB_face_t, int, *int, []float32) {}

func HB_font_set_var_coords_design(*Hb_font_t, []float32) {}

type HB_face_t struct { //   hb_object_header_t header;
	//   hb_reference_table_func_t  reference_table_func;
	//   void                      *user_data;
	//   hb_destroy_func_t          destroy;
	//   unsigned int index;			/* Face index in a collection, zero-based. */
	//   mutable hb_atomic_int_t upem;		/* Units-per-EM. */
	//   mutable hb_atomic_int_t num_glyphs;	/* Number of glyphs. */
	//   hb_shaper_object_dataset_t<HB_face_t> data;/* Various shaper data. */
	//   hb_ot_face_t table;			/* All the face's tables. */
	//   /* Cache */
	//   struct plan_node_t
	//   {
	//     hb_shape_plan_t *shape_plan;
	//     plan_node_t *next;
	//   };
	//   hb_atomic_ptr_t<plan_node_t> shape_plans;
}

type hb_blob_t = []byte

func HB_blob_create_from_file(file string) *hb_blob_t {
	return nil
}

func HB_face_create(blob *hb_blob_t, index int) *HB_face_t {
	return nil
}

//
// TODO:
// func GetHBFont(font Font) *hb_font_t {
// 	PangoFontPrivate * priv = pango_font_get_instance_private(font)

// 	g_return_val_if_fail(PANGO_IS_FONT(font), NULL)

// 	if priv.hb_font {
// 		return priv.hb_font
// 	}

// 	priv.hb_font = font.CreateHBFont()

// 	hb_font_make_immutable(priv.hb_font)

// 	return priv.hb_font
// }

// TODO:
func HbFontNominalGlyph(font *Hb_font_t, u rune) (Glyph, bool) {
	// return font.get_nominal_glyph(unicode, glyph)
	return 0, false
}

// TODO:
func hb_font_get_glyph_h_advance(font *Hb_font_t, glyph Glyph) int32 {
	//    return font.GetGlyphHAdvance (glyph);
	return 0
}

type Position = int

// type Direction = int

/* Note that typically ascender is positive and descender negative in coordinate systems that grow up. */
// TODO: use plain ints if possible
type hb_font_extents_t struct {
	Ascender  Position // typographic ascender.
	Descender Position // typographic descender.
	LineGap   Position // suggested line spacing gap.
}

// TODO:
func HBFontGetExtentsForDirection(font *Hb_font_t, direction Direction) hb_font_extents_t {
	return hb_font_extents_t{}
}

// TODO:
func HbOtMetricsGetPosition(font *Hb_font_t, tag truetype.Tag) (Position, bool) {
	return 0, false
}

//  * Glyph extent values, measured in font units.
// Note that height is negative in coordinate systems that grow up.
type HB_glyph_extents_t struct {
	XBearing Position // left side of glyph from origin
	YBearing Position // top side of glyph from origin
	Width    Position // distance from left to right side
	Height   Position // distance from top to bottom side
}

func HB_font_get_glyph_extents(font *Hb_font_t, glyph Glyph) HB_glyph_extents_t {
	return HB_glyph_extents_t{}
}

func HB_font_get_glyph_advance_for_direction(*Hb_font_t, Glyph, Direction) (x, y Position) {
	return 0, 0
}
