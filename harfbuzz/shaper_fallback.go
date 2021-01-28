package harfbuzz

// ported from harfbuzz/src/hb-fallback-shape.cc Copyright © 2011  Google, Inc. Behdad Esfahbod

// implements a default shapper for fonts without advanced features

type hb_fallback_face_data_t struct{}

type hb_fallback_font_data_t struct{}

func _hb_fallback_shape(_ *hb_shape_plan_t, font *hb_font_t, buffer *hb_buffer_t, _ []hb_feature_t) bool {
	/* TODO
	*
	* - Apply fallback kern.
	* - Handle Variation Selectors?
	* - Apply normalization?
	*
	* This will make the fallback shaper into a dumb "TrueType"
	* shaper which many people unfortunately still request.
	 */

	space, hasSpace := font.face.GetNominalGlyph(' ')

	buffer.clear_positions()

	direction := buffer.props.direction
	info := buffer.info
	pos := buffer.pos
	for i := range info {
		if hasSpace && uni.is_default_ignorable(info[i].codepoint) {
			info[i].codepoint = space
			pos[i].x_advance = 0
			pos[i].y_advance = 0
			continue
		}
		info[i].codepoint, _ = font.face.GetNominalGlyph(info[i].codepoint)
		font.get_glyph_advance_for_direction(info[i].codepoint, direction,
			&pos[i].x_advance, &pos[i].y_advance)
		font.subtract_glyph_origin_for_direction(info[i].codepoint, direction,
			&pos[i].x_offset, &pos[i].y_offset)
	}

	if direction.isBackward() {
		buffer.reverse()
	}

	buffer.safe_to_break_all()

	return true
}
