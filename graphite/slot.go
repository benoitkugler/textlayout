package graphite

type slot struct {
	prev, next  *slot // linked list of slots
	original    int   // charinfo that originated this slot (e.g. for feature values)
	before      int   // charinfo index of before association
	after       int   // charinfo index of after association
	index       int   // slot index given to this slot during finalising
	glyphID     GID
	realGlyphID GID
	// Slot *m_parent;         // index to parent we are attached to
	// Slot *m_child;          // index to first child slot that attaches to us
	// Slot *m_sibling;        // index to next child that attaches to our parent
	position position // absolute position of glyph
	shift    position // .shift slot attribute
	advance  position // .advance slot attribute
	attach   position // attachment point on us
	with     position // attachment point position on parent
	// float    m_just;        // Justification inserted space
	// uint8    m_flags;       // holds bit flags
	// byte     m_attLevel;    // attachment level
	bidiCls int8 // bidirectional class
	// byte     m_bidiLevel;   // bidirectional level
	// int16   *m_userAttr;    // pointer to user attributes
	// SlotJustify *m_justs;   // pointer to justification parameters
}

func (sl *slot) setGlyph(seg *segment, glyphID GID) {
	sl.glyphID = glyphID
	sl.bidiCls = -1
	theGlyph := seg.face.getGlyph(glyphID)
	if theGlyph == nil {
		sl.realGlyphID = 0
		sl.advance = position{}
		return

	}
	sl.realGlyphID = GID(theGlyph.attrs.get(uint16(seg.silf.AttrPseudo)))
	if sl.realGlyphID > seg.face.numGlyphs {
		sl.realGlyphID = 0
	}
	aGlyph := theGlyph
	if sl.realGlyphID != 0 {
		aGlyph = seg.face.getGlyph(sl.realGlyphID)
		if aGlyph == nil {
			aGlyph = theGlyph
		}
	}
	sl.advance = position{x: float32(aGlyph.advance), y: 0.}
	if seg.silf.AttrSkipPasses != 0 {
		seg.mergePassBits(uint32(theGlyph.attrs.get(uint16(seg.silf.AttrSkipPasses))))
		if seg.silf.NumPasses > 16 {
			seg.mergePassBits(uint32(theGlyph.attrs.get(uint16(seg.silf.AttrSkipPasses)+1)) << 16)
		}
	}
}
