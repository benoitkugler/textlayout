package graphite

import (
	"encoding/json"
	"fmt"
)

// this file implements logging helpers, which are only used
// in debug mode

func dumpJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "\t")
	return string(b)
}

func (ci charInfo) MarshalJSON() ([]byte, error) {
	type charInfoSlotJSON struct {
		Before int `json:"before"`
		After  int `json:"after"`
	}
	type charInfoJSON struct {
		Offset  int              `json:"offset"`
		Unicode rune             `json:"unicode"`
		Break   int16            `json:"break"`
		Flags   uint8            `json:"flags"`
		Slot    charInfoSlotJSON `json:"slot"`
	}
	out := charInfoJSON{
		Offset:  ci.base,
		Unicode: ci.char,
		Break:   ci.breakWeight,
		Flags:   ci.flags,
		Slot: charInfoSlotJSON{
			Before: ci.before,
			After:  ci.after,
		},
	}
	return json.Marshal(out)
}

type slotCharInfoJSON struct {
	Original int `json:"original"`
	Before   int `json:"before"`
	After    int `json:"after"`
}

func (ci slotCharInfoJSON) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("original: %d, before: %d, after: %d", ci.Original, ci.Before, ci.After)), nil
}

func (pos Position) MarshalText() ([]byte, error) {
	v := [2]float32{pos.X, pos.Y}
	return []byte(fmt.Sprintf("%v", v)), nil
}

func (s *Slot) objectID() string { return fmt.Sprintf("%p", s) }

type slotParentJSON struct {
	Id     string   `json:"id"`
	Level  int32    `json:"level"`
	Offset Position `json:"offset"`
}

type slotJSON struct {
	Id            string           `json:"id"`
	Gid           GID              `json:"gid"`
	Charinfo      slotCharInfoJSON `json:"charinfo"`
	Origin        Position         `json:"origin"`
	Shift         Position         `json:"shift"`
	Advance       Position         `json:"advance"`
	Insert        bool             `json:"insert"`
	Break         int32            `json:"break"`
	Justification float32          `json:"justification,omitempty"`
	Bidi          uint8            `json:"bidi,omitempty"`
	Parent        *slotParentJSON  `json:"parent,omitempty"`
	User          []int16          `json:"user"`
	Children      []string         `json:"children,omitempty"`
}

// returns a JSON compatible representation of the slot
func (s *Slot) json(seg *Segment) slotJSON {
	out := slotJSON{
		Id:  s.objectID(),
		Gid: s.GlyphID,
		Charinfo: slotCharInfoJSON{
			Original: s.original,
			Before:   s.Before,
			After:    s.After,
		},
		Origin: s.Position,
		Shift: Position{
			X: float32(s.getAttr(nil, gr_slatShiftX, 0)),
			Y: float32(s.getAttr(nil, gr_slatShiftY, 0)),
		},
		Advance:       s.Advance,
		Insert:        s.CanInsertBefore(),
		Break:         s.getAttr(seg, gr_slatBreak, 0),
		Justification: s.just,
		Bidi:          s.bidiLevel,
		User:          s.userAttrs,
	}
	if !s.isBase() {
		out.Parent = &slotParentJSON{
			Id:     s.parent.objectID(),
			Level:  s.getAttr(nil, gr_slatAttLevel, 0),
			Offset: s.attach.sub(s.with),
		}
	}
	if s.child != nil {
		for c := s.child; c != nil; c = c.sibling {
			out.Children = append(out.Children, c.objectID())
		}
	}
	if cslot := seg.getCollisionInfo(s); cslot != nil {
		// 		// Note: the reason for using Positions to lump together related attributes is to make the
		// 		// JSON output slightly more compact.
		// 		j << "collision" << json::flat << json::object
		// //              << "shift" << cslot.shift() -- not used pass level, only within the collision routine itself
		// 			  << "offset" << cslot.offset()
		// 			  << "limit" << cslot.limit()
		// 			  << "flags" << cslot.flags()
		// 			  << "margin" << Position(cslot.margin(), cslot.marginWt())
		// 			  << "exclude" << cslot.exclGlyph()
		// 			  << "excludeoffset" << cslot.exclOffset();
		// 		if (cslot.seqOrder() != 0)
		// 		{
		// 			j << "seqclass" << Position(cslot.seqClass(), cslot.seqProxClass())
		// 				<< "seqorder" << cslot.seqOrder()
		// 				<< "seqabove" << Position(cslot.seqAboveXoff(), cslot.seqAboveWt())
		// 				<< "seqbelow" << Position(cslot.seqBelowXlim(), cslot.seqBelowWt())
		// 				<< "seqvalign" << Position(cslot.seqValignHt(), cslot.seqValignWt());
		// 		}
		// 		j << json::close;
		// 	}
		// 	return j << json::close;
	}
	return out
}

func (seg *Segment) slotsJSON() (out []slotJSON) {
	for s := seg.First; s != nil; s = s.Next {
		out = append(out, s.json(seg))
	}
	return out
}

func (s *passes) passJSON(seg *Segment, i uint8) string {
	type passJSON struct {
		ID       uint8      `json:"index"`
		Slotsdir string     `json:"slotsdir"`
		Passdir  string     `json:"passdir"`
		Slots    []slotJSON `json:"slots"`
	}
	sd, pd := "ltr", "ltr"
	if seg.currdir() {
		sd = "rtl"
	}
	if s.isRTL != s.passes[i].isReverseDirection {
		pd = "rtl"
	}
	debug := passJSON{
		ID:       i + 1,
		Slotsdir: sd,
		Passdir:  pd,
		Slots:    seg.slotsJSON(),
	}
	v, _ := json.MarshalIndent(debug, "", "\t")
	return string(v)
}

func input_slot(slots *slotMap, n int) *Slot {
	s := slots.get(int(slots.preContext) + n)
	if !s.isCopied() {
		return s
	}

	if s.prev != nil {
		return s.prev.Next
	}
	if s.Next != nil {
		return s.Next.prev
	}
	return slots.segment.last
}

func dumpRuleEventConsidered(fsm *finiteStateMachine, length int) {
	// *fsm.dbgout << "considered" << json::array;
	for _, ruleIndex := range fsm.rules[:length] {
		r := fsm.ruleTable[ruleIndex]
		if uint16(r.preContext) > fsm.slots.preContext {
			continue
		}
		rj := struct {
			ID     uint16
			Failed bool
			Input  struct {
				Start  string
				Length uint16
			}
		}{
			ID:     ruleIndex,
			Failed: true,
			Input: struct {
				Start  string
				Length uint16
			}{
				Start:  input_slot(&fsm.slots, -int(r.preContext)).objectID(),
				Length: r.sortKey,
			},
		}
		fmt.Println(dumpJSON(rj))
	}
}
