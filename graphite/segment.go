package graphite

const MAX_SEG_GROWTH_FACTOR = 64

type charInfo struct {
	before       int  // slot index before us, comes before
	after        int  // slot index after us, comes after
	featureIndex int  // index into features list in the segment
	char         rune // Unicode character from character stream
	// base        int   // index into input string corresponding to this charinfo
	breakWeight int16 // breakweight coming from lb table
	flags       uint8 // 0,1 segment split.
}

type segment struct {
	face        *graphiteFace
	silf        *silfSubtable // selected subtable
	feats       [][]FeatureValued
	first, last *slot      // first and last slot in segment
	charinfo    []charInfo // character info, one per input character
	dir         int        // text direction

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
	// size_t          m_bufSize,          // how big a buffer to create when need more slots
	//                 m_numGlyphs,
	//                 m_numCharinfo;      // size of the array and number of input characters
	// int             m_defaultOriginal;  // number of whitespace chars in the string
	// uint8           m_flags,            // General purpose flags

	passBits uint32 // if bit set then skip pass
}

func (face *graphiteFace) newSegment(text []rune, script Tag, features []FeatureValued, dir int) *segment {
	var seg segment

	// adapt convention
	script = spaceToZero(script)

	// allocate memory
	seg.charinfo = make([]charInfo, len(text))

	// choose silf
	if len(face.silf) != 0 {
		seg.silf = &face.silf[0]
	} else {
		seg.silf = &silfSubtable{}
	}

	seg.dir = dir
	seg.feats = append(seg.feats, features)

	seg.processRunes(text, len(seg.feats)-1)
	return &seg
}

func (seg *segment) currdir() bool { return ((seg.dir>>6)^seg.dir)&1 != 0 }

func (seg *segment) mergePassBits(val uint32) { seg.passBits &= val }

func (seg *segment) processRunes(text []rune, featureID int) {
	for slotID, r := range text {
		gid := seg.face.cmap.Lookup(r)
		if gid == 0 {
			gid = seg.silf.findPdseudoGlyph(r)
		}
		seg.appendSlot(slotID, r, gid, featureID)
	}
}

func (seg *segment) appendSlot(index int, cid rune, gid GID, featureID int) {
	info := &seg.charinfo[index]
	info.char = cid
	info.featureIndex = featureID
	// info.base = indexFeat
	glyph := seg.face.getGlyph(gid)
	if glyph != nil {
		info.breakWeight = glyph.attrs.get(uint16(seg.silf.AttrBreakWeight))
	}

	sl := new(slot)
	sl.setGlyph(seg, gid)
	sl.original, sl.before, sl.after = index, index, index
	if seg.last != nil {
		seg.last.next = sl
	}
	sl.prev = seg.last
	seg.last = sl
	if seg.first == nil {
		seg.first = sl
	}

	if aPassBits := uint16(seg.silf.AttrSkipPasses); glyph != nil && aPassBits != 0 {
		m := uint32(glyph.attrs.get(aPassBits))
		if seg.silf.NumPasses > 16 {
			m |= uint32(glyph.attrs.get(aPassBits+1)) << 16
		}
		seg.mergePassBits(m)
	}
}

// reverse the slots but keep diacritics in their same position after their bases
func (seg *segment) reverseSlots() {
	seg.dir = seg.dir ^ 64 // invert the reverse flag
	if seg.first == seg.last {
		return
	} // skip 0 or 1 glyph runs

	var (
		curr                  = seg.first
		t, tlast, tfirst, out *slot
	)

	for curr != nil && seg.getSlotBidiClass(curr) == 16 {
		curr = curr.next
	}
	if curr == nil {
		return
	}
	tfirst = curr.prev
	tlast = curr

	for curr != nil {
		if seg.getSlotBidiClass(curr) == 16 {
			d := curr.next
			for d != nil && seg.getSlotBidiClass(d) == 16 {
				d = d.next
			}
			if d != nil {
				d = d.prev
			} else {
				d = seg.last
			}
			p := out.next // one after the diacritics. out can't be null
			if p != nil {
				p.prev = d
			} else {
				tlast = d
			}
			t = d.next
			d.next = p
			curr.prev = out
			out.next = curr
		} else { // will always fire first time round the loop
			if out != nil {
				out.prev = curr
			}
			t = curr.next
			curr.next = out
			out = curr
		}
		curr = t
	}
	out.prev = tfirst
	if tfirst != nil {
		tfirst.next = out
	} else {
		seg.first = out
	}
	seg.last = tlast
}

func (seg *segment) doMirror(aMirror byte) {
	for s := seg.first; s != nil; s = s.next {
		g := GID(seg.face.getGlyphAttr(s.glyphID, uint16(aMirror)))
		if g != 0 && (seg.dir&4 == 0 || seg.face.getGlyphAttr(s.glyphID, uint16(aMirror)+1) == 0) {
			s.setGlyph(seg, g)
		}
	}
}

func (seg *segment) getSlotBidiClass(s *slot) int8 {
	if res := s.bidiCls; res != -1 {
		return res
	}
	res := int8(seg.face.getGlyphAttr(s.glyphID, uint16(seg.silf.AttrDirectionality)))
	s.bidiCls = res
	return res
}
