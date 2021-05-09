package graphite

import "fmt"

const MAX_SEG_GROWTH_FACTOR = 64

type charInfo struct {
	before int // slot index before us, comes before
	after  int // slot index after us, comes after
	// featureIndex int  // index into features list in the segment −> Always 0
	char rune // Unicode character from character stream
	// base        int   // index into input string corresponding to this charinfo
	breakWeight int16 // breakweight coming from lb table
	flags       uint8 // 0,1 segment split.
}

func (ch *charInfo) addFlags(val uint8) { ch.flags |= val }

// Segment represents a line of text.
type Segment struct {
	face        *graphiteFace
	silf        *passes // selected subtable
	feats       FeaturesValue
	first, last *Slot      // first and last slot in segment
	charinfo    []charInfo // character info, one per input character

	// SlotRope        m_slots;            // Vector of slot buffers
	// AttributeRope   m_userAttrs;        // Vector of userAttrs buffers
	// JustifyRope     m_justifies;        // Slot justification info buffers
	// FeatureList     m_feats;            // feature settings referenced by charinfos in this segment
	freeSlots  *Slot // linked list of free slots
	collisions []slotCollision
	advance    Position // whole segment advance

	dir int // text direction
	// SlotJustify   * m_freeJustifies;    // Slot justification blocks free list
	// const Face    * m_face;             // GrFace
	// const Silf    * m_silf;
	// size_t          m_bufSize,          // how big a buffer to create when need more slots
	numGlyphs int
	//                 m_numCharinfo;      // size of the array and number of input characters
	defaultOriginal int // number of whitespace chars in the string
	// uint8           m_flags,            // General purpose flags

	passBits uint32 // if bit set then skip pass
}

func (face *graphiteFace) newSegment(font *graphiteFont, text []rune, script Tag, features FeaturesValue, dir int) *Segment {
	var seg Segment

	// adapt convention
	script = spaceToZero(script)

	// allocate memory
	seg.charinfo = make([]charInfo, len(text))
	seg.numGlyphs = len(text)

	// choose silf
	if len(face.silf) != 0 {
		seg.silf = &face.silf[0]
	} else {
		seg.silf = &passes{}
	}

	seg.dir = dir
	seg.feats = features

	seg.processRunes(text)

	seg.finalise(font, true)
	return &seg
}

func (seg *Segment) currdir() bool { return ((seg.dir>>6)^seg.dir)&1 != 0 }

func (seg *Segment) mergePassBits(val uint32) { seg.passBits &= val }

func (seg *Segment) processRunes(text []rune) {
	for slotID, r := range text {
		gid := seg.face.cmap.Lookup(r)
		if gid == 0 {
			gid = seg.silf.findPdseudoGlyph(r)
		}
		seg.appendSlot(slotID, r, gid)
	}
}

func (seg *Segment) newSlot() *Slot {
	sl := new(Slot)
	sl.userAttrs = make([]int16, seg.silf.userAttibutes)
	return sl
}

func (seg *Segment) newJustify() *slotJustify {
	return new(slotJustify)
}

func (seg *Segment) appendSlot(index int, cid rune, gid GID) {
	sl := seg.newSlot()

	info := &seg.charinfo[index]
	info.char = cid
	// info.featureIndex = featureID
	// info.base = indexFeat
	glyph := seg.face.getGlyph(gid)
	if glyph != nil {
		info.breakWeight = glyph.attrs.get(uint16(seg.silf.attrBreakWeight))
	}

	sl.setGlyph(seg, gid)
	sl.original, sl.Before, sl.After = index, index, index
	if seg.last != nil {
		seg.last.Next = sl
	}
	sl.prev = seg.last
	seg.last = sl
	if seg.first == nil {
		seg.first = sl
	}

	if aPassBits := uint16(seg.silf.attrSkipPasses); glyph != nil && aPassBits != 0 {
		m := uint32(glyph.attrs.get(aPassBits))
		if len(seg.silf.passes) > 16 {
			m |= uint32(glyph.attrs.get(aPassBits+1)) << 16
		}
		seg.mergePassBits(m)
	}
}

func (seg *Segment) freeSlot(aSlot *Slot) {
	if aSlot == nil {
		return
	}
	if seg.last == aSlot {
		seg.last = aSlot.prev
	}
	if seg.first == aSlot {
		seg.first = aSlot.Next
	}
	if aSlot.parent != nil {
		aSlot.parent.removeChild(aSlot)
	}
	for aSlot.child != nil {
		if aSlot.child.parent == aSlot {
			aSlot.child.parent = nil
			aSlot.removeChild(aSlot.child)
		} else {
			aSlot.child = nil
		}
	}

	if debugMode >= 2 {
		fmt.Println("freeing slot")
	}

	// update next pointer
	aSlot.Next = seg.freeSlots
	seg.freeSlots = aSlot
}

// reverse the slots but keep diacritics in their same position after their bases
func (seg *Segment) reverseSlots() {
	seg.dir = seg.dir ^ 64 // invert the reverse flag
	if seg.first == seg.last {
		return
	} // skip 0 or 1 glyph runs

	var (
		curr                  = seg.first
		t, tlast, tfirst, out *Slot
	)

	for curr != nil && seg.getSlotBidiClass(curr) == 16 {
		curr = curr.Next
	}
	if curr == nil {
		return
	}
	tfirst = curr.prev
	tlast = curr

	for curr != nil {
		if seg.getSlotBidiClass(curr) == 16 {
			d := curr.Next
			for d != nil && seg.getSlotBidiClass(d) == 16 {
				d = d.Next
			}
			if d != nil {
				d = d.prev
			} else {
				d = seg.last
			}
			p := out.Next // one after the diacritics. out can't be null
			if p != nil {
				p.prev = d
			} else {
				tlast = d
			}
			t = d.Next
			d.Next = p
			curr.prev = out
			out.Next = curr
		} else { // will always fire first time round the loop
			if out != nil {
				out.prev = curr
			}
			t = curr.Next
			curr.Next = out
			out = curr
		}
		curr = t
	}
	out.prev = tfirst
	if tfirst != nil {
		tfirst.Next = out
	} else {
		seg.first = out
	}
	seg.last = tlast
}

// TODO: check if font is not always passed as nil
func (seg *Segment) positionSlots(font *graphiteFont, iStart, iEnd *Slot, isRtl, isFinal bool) Position {
	var (
		currpos    Position
		clusterMin float32
		bbox       rect
		reorder    = (seg.currdir() != isRtl)
	)

	if reorder {
		seg.reverseSlots()
		iStart, iEnd = iEnd, iStart
	}
	if iStart == nil {
		iStart = seg.first
	}
	if iEnd == nil {
		iEnd = seg.last
	}

	if iStart == nil || iEnd == nil { // only true for empty segments
		return currpos
	}

	if isRtl {
		for s, end := iEnd, iStart.prev; s != nil && s != end; s = s.prev {
			if s.isBase() {
				clusterMin = currpos.x
				currpos = s.finalise(seg, font, currpos, &bbox, 0, &clusterMin, isRtl, isFinal, 0)
			}
		}
	} else {
		for s, end := iStart, iEnd.Next; s != nil && s != end; s = s.Next {
			if s.isBase() {
				clusterMin = currpos.x
				currpos = s.finalise(seg, font, currpos, &bbox, 0, &clusterMin, isRtl, isFinal, 0)
			}
		}
	}
	if reorder {
		seg.reverseSlots()
	}
	return currpos
}

func (seg *Segment) doMirror(aMirror byte) {
	for s := seg.first; s != nil; s = s.Next {
		g := GID(seg.face.getGlyphAttr(s.glyphID, uint16(aMirror)))
		if g != 0 && (seg.dir&4 == 0 || seg.face.getGlyphAttr(s.glyphID, uint16(aMirror)+1) == 0) {
			s.setGlyph(seg, g)
		}
	}
}

func (seg *Segment) getSlotBidiClass(s *Slot) int8 {
	if res := s.bidiCls; res != -1 {
		return res
	}
	res := int8(seg.face.getGlyphAttr(s.glyphID, uint16(seg.silf.attrDirectionality)))
	s.bidiCls = res
	return res
}

// check the bounds and return nil if needed
func (seg *Segment) getCharInfo(index int) *charInfo {
	if index < len(seg.charinfo) {
		return &seg.charinfo[index]
	}
	return nil
}

// check the bounds and return nil if needed
func (seg *Segment) getCollisionInfo(s *Slot) *slotCollision {
	if s.index < len(seg.collisions) {
		return &seg.collisions[s.index]
	}
	return nil
}

func (seg *Segment) getFeature(findex uint8) int32 {
	if feat := seg.feats.findFeature(Tag(findex)); feat != nil {
		return int32(feat.Value)
	}
	return 0
}

func (seg *Segment) setFeature(findex uint8, val int16) {
	if feat := seg.feats.findFeature(Tag(findex)); feat != nil {
		feat.Value = val
	}
}

func (seg *Segment) getGlyphMetric(iSlot *Slot, metric, attrLevel uint8, rtl bool) int32 {
	if attrLevel > 0 {
		is := iSlot.findRoot()
		return is.clusterMetric(seg, metric, attrLevel, rtl)
	}
	return seg.face.getGlyphMetric(iSlot.glyphID, metric)
}

func (seg *Segment) finalise(font *graphiteFont, reverse bool) {
	if seg.first == nil || seg.last == nil {
		return
	}

	seg.advance = seg.positionSlots(font, seg.first, seg.last, seg.silf.isRTL, true)
	// associateChars(0, seg.numCharinfo);
	if reverse && seg.currdir() != (seg.dir&1 != 0) {
		seg.reverseSlots()
	}
	seg.linkClusters(seg.first, seg.last)
}

func (seg *Segment) linkClusters(s, end *Slot) {
	end = end.Next

	for ; s != end && !s.isBase(); s = s.Next {
	}
	ls := s

	if seg.dir&1 != 0 {
		for ; s != end; s = s.Next {
			if !s.isBase() {
				continue
			}

			s.sibling = ls
			ls = s
		}
	} else {
		for ; s != end; s = s.Next {
			if !s.isBase() {
				continue
			}

			ls.sibling = s
			ls = s
		}
	}
}
