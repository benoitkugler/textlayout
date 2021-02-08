package harfbuzz

// ported from harfbuzz/src/hb-fallback-shape.cc Copyright © 2011  Google, Inc. Behdad Esfahbod

var _ shaper = shaperFallback{}

// shaperFallback implements a naive shaper, which does the minimum,
// without requiring advanced Opentype font features.
type shaperFallback struct{}

func (shaperFallback) shape(_ *ShapePlan, font *Font, buffer *Buffer, _ []Feature) bool {
	space, hasSpace := font.Face.GetNominalGlyph(' ')

	buffer.ClearPositions()

	direction := buffer.Props.Direction
	info := buffer.info
	pos := buffer.pos
	for i := range info {
		if hasSpace && Uni.IsDefaultIgnorable(info[i].codepoint) {
			info[i].codepoint = space
			pos[i].XAdvance = 0
			pos[i].YAdvance = 0
			continue
		}
		info[i].codepoint, _ = font.Face.GetNominalGlyph(info[i].codepoint)
		pos[i].XAdvance, pos[i].YAdvance = font.GetGlyphAdvanceForDirection(info[i].codepoint, direction)
		pos[i].XOffset, pos[i].YOffset = font.SubtractGlyphOriginForDirection(info[i].codepoint, direction,
			pos[i].XOffset, pos[i].YOffset)
	}

	if direction.IsBackward() {
		buffer.Reverse()
	}

	buffer.SafeToBreakAll()

	return true
}
