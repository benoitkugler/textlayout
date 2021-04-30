package graphite

const (
	DELETED uint8 = 1 << iota
	INSERTED
	COPIED
	POSITIONED
	ATTACHED
)

type slot struct {
	prev, next *slot // linked list of slots
	parent     *slot // index to parent we are attached to
	child      *slot // index to first child slot that attaches to us
	sibling    *slot // index to next child that attaches to our parent

	userAttrs []int16 // pointer to user attributes

	original    int // charinfo that originated this slot (e.g. for feature values)
	before      int // charinfo index of before association
	after       int // charinfo index of after association
	index       int // slot index given to this slot during finalising
	glyphID     GID
	realGlyphID GID
	position    position // absolute position of glyph
	shift       position // .shift slot attribute
	advance     position // .advance slot attribute
	attach      position // attachment point on us
	with        position // attachment point position on parent
	// float    m_just;        // Justification inserted space
	flags uint8 // holds bit flags
	// byte     m_attLevel;    // attachment level
	bidiCls int8 // bidirectional class
	// byte     m_bidiLevel;   // bidirectional level
	// SlotJustify *m_justs;   // pointer to justification parameters
}

func (sl *slot) isDeleted() bool {
	return sl.flags&DELETED != 0
}

func (sl *slot) markDeleted(state bool) {
	if state {
		sl.flags |= DELETED
	} else {
		sl.flags &= ^DELETED
	}
}

func (sl *slot) isCopied() bool {
	return sl.flags&COPIED != 0
}

func (sl *slot) markCopied(state bool) {
	if state {
		sl.flags |= COPIED
	} else {
		sl.flags &= ^COPIED
	}
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
