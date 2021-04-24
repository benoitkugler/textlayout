package graphite

import "github.com/benoitkugler/textlayout/fonts/truetype"

type segment struct {
	silf     *silfSubtable
	feats    [][]FeatureValued
	charinfo []charInfo // character info, one per input character
	dir      int        // text direction

	// Position        m_advance;          // whole segment advance
	// SlotRope        m_slots;            // Vector of slot buffers
	// AttributeRope   m_userAttrs;        // Vector of userAttrs buffers
	// JustifyRope     m_justifies;        // Slot justification info buffers
	// FeatureList     m_feats;            // feature settings referenced by charinfos in this segment
	// Slot          * m_freeSlots;        // linked list of free slots
	// SlotJustify   * m_freeJustifies;    // Slot justification blocks free list
	// SlotCollision * m_collisions;
	// const Face    * m_face;             // GrFace
	// const Silf    * m_silf;
	// Slot          * m_first;            // first slot in segment
	// Slot          * m_last;             // last slot in segment
	// size_t          m_bufSize,          // how big a buffer to create when need more slots
	//                 m_numGlyphs,
	//                 m_numCharinfo;      // size of the array and number of input characters
	// int             m_defaultOriginal;  // number of whitespace chars in the string
	// uint8           m_flags,            // General purpose flags
	//                 m_passBits;         // if bit set then skip pass
}

func newSegment(silf TableSilf, text []rune, script Tag, features []FeatureValued, dir int) *segment {
	var out segment

	// adapt convention
	script = spaceToZero(script)

	// allocate memory
	out.charinfo = make([]charInfo, len(text))

	// choose silf
	if len(silf) != 0 {
		out.silf = &silf[0]
	}

	out.dir = dir
	out.feats = append(out.feats, features)
	// TODO:

	return &out
}

func (s *segment) processRunes(cmap truetype.Cmap, text []rune) {
	for _, r := range text {
		gid := cmap.Lookup(r)
		if gid == 0 {
			gid = s.silf.findPdseudoGlyph(r)
		}
	}
}

func (s *segment) appendSlot(index int, cid rune, gid GID, indexFeat uint8) {
	// var sl slot

	s.charinfo[index].char = cid
	s.charinfo[index].featureId = indexFeat
	// s.charinfo[index].base = indexFeat
}

type charInfo struct {
	before int  // slot index before us, comes before
	after  int  // slot index after us, comes after
	char   rune // Unicode character from character stream
	// base        int   // index into input string corresponding to this charinfo
	featureId   uint8 // index into features list in the segment
	breakWeight int8  // breakweight coming from lb table
	flags       uint8 // 0,1 segment split.
}
