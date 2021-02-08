package harfbuzz

// ported from harfbuzz/src/hb-fallback-shape.cc Copyright © 2011  Google, Inc. Behdad Esfahbod

var _ Shaper = ShaperFallback{}

// ShaperFallback implements a naive shaper, which does the minimum,
// without requiring advanced Opentype font features.
type ShaperFallback struct{}

// Shape implements harfbuzz.Shaper
func (ShaperFallback) Shape(_ *ShapePlan, font *Font, buffer *Buffer, _ []Feature) bool {
	space, hasSpace := font.Face.GetNominalGlyph(' ')

	buffer.ClearPositions()

	direction := buffer.Props.Direction
	info := buffer.Info
	pos := buffer.Pos
	for i := range info {
		if hasSpace && Uni.IsDefaultIgnorable(info[i].Codepoint) {
			info[i].Codepoint = space
			pos[i].XAdvance = 0
			pos[i].YAdvance = 0
			continue
		}
		info[i].Codepoint, _ = font.Face.GetNominalGlyph(info[i].Codepoint)
		pos[i].XAdvance, pos[i].YAdvance = font.GetGlyphAdvanceForDirection(info[i].Codepoint, direction)
		pos[i].XOffset, pos[i].YOffset = font.SubtractGlyphOriginForDirection(info[i].Codepoint, direction,
			pos[i].XOffset, pos[i].YOffset)
	}

	if direction.IsBackward() {
		buffer.Reverse()
	}

	buffer.SafeToBreakAll()

	return true
}
