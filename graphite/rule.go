package graphite

const MAX_SLOTS = 64

type slotMap struct {
	segment   *segment
	highwater *slot

	slots      [MAX_SLOTS + 1]*slot
	size       int
	maxSize    int
	highpassed bool
	//   unsigned short m_precontext;
	dir uint8
}

func (sm *slotMap) decMax() int {
	sm.maxSize--
	return sm.maxSize
}

func (sm *slotMap) get(index int) *slot {
	return sm.slots[index+1]
}

func (sm *slotMap) begin() *slot {
	// allow map to go 1 before slot_map when inserting
	// at start of segment.
	return sm.slots[1]
}

func (sm *slotMap) endMinus1() *slot {
	return sm.slots[sm.size]
}
