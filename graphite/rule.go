package graphite

const MAX_SLOTS = 64

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

func (sm *slotMap) begin() *Slot {
	// allow map to go 1 before slot_map when inserting
	// at start of segment.
	return sm.slots[1]
}

func (sm *slotMap) endMinus1() *Slot {
	return sm.slots[sm.size]
}
