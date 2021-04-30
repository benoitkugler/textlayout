package graphite

const MAX_SLOTS = 64

type slotMap struct {
	segment   *segment
	highwater *slot

	slots      [MAX_SLOTS + 1]*slot
	size       int
	highpassed bool
	//   unsigned short m_precontext;
	maxSize int
	//   uint8          m_dir;
}

func (sm *slotMap) decMax() int {
	sm.maxSize--
	return sm.maxSize
}
