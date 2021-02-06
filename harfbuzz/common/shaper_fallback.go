package common

// ported from harfbuzz/src/hb-fallback-shape.cc Copyright © 2011  Google, Inc. Behdad Esfahbod

// implements a default shapper for fonts without advanced features

type hb_fallback_face_data_t struct{}

type hb_fallback_font_data_t struct{}

func _hb_fallback_shape(_ *hb_shape_plan_t, font *Font, buffer *Buffer, _ []hb_feature_t) bool {
	space, hasSpace := font.Face.GetNominalGlyph(' ')

	buffer.clear_positions()

	direction := buffer.Props.Direction
	info := buffer.Info
	pos := buffer.Pos
	for i := range info {
		if hasSpace && Uni.is_default_ignorable(info[i].Codepoint) {
			info[i].Codepoint = space
			pos[i].XAdvance = 0
			pos[i].y_advance = 0
			continue
		}
		info[i].Codepoint, _ = font.Face.GetNominalGlyph(info[i].Codepoint)
		pos[i].XAdvance, pos[i].y_advance = font.get_glyph_advance_for_direction(info[i].Codepoint, direction)
		pos[i].XOffset, pos[i].y_offset = font.subtract_glyph_origin_for_direction(info[i].Codepoint, direction,
			pos[i].XOffset, pos[i].y_offset)
	}

	if direction.isBackward() {
		buffer.reverse()
	}

	buffer.safe_to_break_all()

	return true
}