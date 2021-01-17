package pango

import "github.com/benoitkugler/textlayout/fonts/truetype"

// TODO:
func pango_hb_shape(font Font, item_text []rune, analysis *Analysis, glyphs *GlyphString, paragraph_text []rune) {
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

func HB_ot_var_named_instance_get_design_coords(*HB_face_t, int, *int, []float64) {}

func HB_font_set_var_coords_design(*Hb_font_t, []float64) {}

type HB_face_t struct {
	//   hb_object_header_t header;

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
func HbFontGetNominalGlyph(font *Hb_font_t, u rune) (Glyph, bool) {
	// return font.get_nominal_glyph(unicode, glyph)
	return 0, false
}

// TODO:
func hb_font_get_glyph_h_advance(font *Hb_font_t, glyph Glyph) int32 {
	//    return font->get_glyph_h_advance (glyph);
	return 0
}

type hb_position_t int32

// hb_direction_t is the text direction
type hb_direction_t uint8

const (
	HB_DIRECTION_LTR     hb_direction_t = 4 + iota // Text is set horizontally from left to right.
	HB_DIRECTION_RTL                               // Text is set horizontally from right to left.
	HB_DIRECTION_TTB                               // Text is set vertically from top to bottom.
	HB_DIRECTION_BTT                               // Text is set vertically from bottom to top.
	HB_DIRECTION_INVALID hb_direction_t = 0        // Initial, unset direction.
)

/* Note that typically ascender is positive and descender negative in coordinate systems that grow up. */
// TODO: use plain ints if possible
type hb_font_extents_t struct {
	Ascender  hb_position_t // typographic ascender.
	Descender hb_position_t // typographic descender.
	LineGap   hb_position_t // suggested line spacing gap.
}

// TODO:
func HBFontGetExtentsForDirection(font *Hb_font_t, direction hb_direction_t) hb_font_extents_t {
	return hb_font_extents_t{}
}

// TODO:
func HbOtMetricsGetPosition(font *Hb_font_t, tag truetype.Tag) (hb_position_t, bool) {
	return 0, false
}

// Note that height is negative in coordinate systems that grow up.
type HB_glyph_extents_t struct {
	XBearing hb_position_t // left side of glyph from origin
	YBearing hb_position_t // top side of glyph from origin
	Width    hb_position_t // distance from left to right side
	Height   hb_position_t // distance from top to bottom side
}

func HB_font_get_glyph_extents(font *Hb_font_t, glyph Glyph) HB_glyph_extents_t {
	return HB_glyph_extents_t{}
}

func HB_font_get_glyph_advance_for_direction(*Hb_font_t, Glyph, hb_direction_t) (x, y hb_position_t) {
	return 0, 0
}
