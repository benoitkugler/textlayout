package graphite

type slot struct { // unsigned short m_glyphid;        // glyph id
	// uint16 m_realglyphid;
	// uint32 m_original;      // charinfo that originated this slot (e.g. for feature values)
	// uint32 m_before;        // charinfo index of before association
	// uint32 m_after;         // charinfo index of after association
	// uint32 m_index;         // slot index given to this slot during finalising
	// Slot *m_parent;         // index to parent we are attached to
	// Slot *m_child;          // index to first child slot that attaches to us
	// Slot *m_sibling;        // index to next child that attaches to our parent
	// Position m_position;    // absolute position of glyph
	// Position m_shift;       // .shift slot attribute
	// Position m_advance;     // .advance slot attribute
	// Position m_attach;      // attachment point on us
	// Position m_with;        // attachment point position on parent
	// float    m_just;        // Justification inserted space
	// uint8    m_flags;       // holds bit flags
	// byte     m_attLevel;    // attachment level
	// int8     m_bidiCls;     // bidirectional class
	// byte     m_bidiLevel;   // bidirectional level
	// int16   *m_userAttr;    // pointer to user attributes
	// SlotJustify *m_justs;   // pointer to justification parameters
}

func (sl *slot) setGlyph(seg *segment, glyphid GID) {
	// m_glyphid = glyphid;
	// m_bidiCls = -1;
	// if (!theGlyph)
	// {
	//     theGlyph = seg->getFace()->glyphs().glyphSafe(glyphid);
	//     if (!theGlyph)
	//     {
	//         m_realglyphid = 0;
	//         m_advance = Position(0.,0.);
	//         return;
	//     }
	// }
	// m_realglyphid = theGlyph->attrs()[seg->silf()->aPseudo()];
	// if (m_realglyphid > seg->getFace()->glyphs().numGlyphs())
	//     m_realglyphid = 0;
	// const GlyphFace *aGlyph = theGlyph;
	// if (m_realglyphid)
	// {
	//     aGlyph = seg->getFace()->glyphs().glyphSafe(m_realglyphid);
	//     if (!aGlyph) aGlyph = theGlyph;
	// }
	// m_advance = Position(aGlyph->theAdvance().x, 0.);
	// if (seg->silf()->aPassBits())
	// {
	//     seg->mergePassBits(uint8(theGlyph->attrs()[seg->silf()->aPassBits()]));
	//     if (seg->silf()->numPasses() > 16)
	//         seg->mergePassBits(theGlyph->attrs()[seg->silf()->aPassBits()+1] << 16);
	// }
}
