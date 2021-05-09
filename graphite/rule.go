package graphite

const MAX_SLOTS = 64

type rule struct {
	action, constraint code   // loaded code
	preContext         uint8  // number of items in the context before the first modified item
	sortKey            uint16 // precedence of the rule
}

type slotMap struct {
	segment   *Segment
	highwater *Slot

	slots      [MAX_SLOTS + 1]*Slot
	size       int
	maxSize    int
	highpassed bool
	preContext uint16
	dir        bool
}

func newSlotMap(seg *Segment, direction bool, maxSize int) slotMap {
	return slotMap{
		segment: seg,
		dir:     direction,
		maxSize: maxSize,
	}
}

func (sm *slotMap) reset(slot *Slot, ctxt uint16) {
	sm.size = 0
	sm.preContext = ctxt
	sm.slots[0] = slot.prev
}

func (sm *slotMap) pushSlot(s *Slot) {
	sm.size++
	sm.slots[sm.size] = s
}

func (sm *slotMap) decMax() int {
	sm.maxSize--
	return sm.maxSize
}

func (sm *slotMap) get(index int) *Slot {
	return sm.slots[index+1]
}

// TODO: check the usage of get and getSlice
func (sm *slotMap) getSlice(index int) []*Slot {
	return sm.slots[index+1:]
}

func (sm *slotMap) begin() *Slot {
	// allow map to go 1 before slot_map when inserting
	// at start of segment.
	return sm.slots[1]
}

func (sm *slotMap) endMinus1() *Slot {
	return sm.slots[sm.size]
}

func (sm *slotMap) collectGarbage(aSlot *Slot) *Slot {
	for s := sm.slots[1:]; len(s) != 0; s = s[1:] {
		slot := s[0]
		if slot != nil && (slot.isDeleted() || slot.isCopied()) {
			if slot == aSlot {
				if slot.prev != nil {
					aSlot = slot.prev
				} else {
					aSlot = slot.Next
				}
			}
			sm.segment.freeSlot(slot)
		}
	}
	return aSlot
}
