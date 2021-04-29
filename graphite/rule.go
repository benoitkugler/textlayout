package graphite

const MAX_SLOTS = 64

type slotMap struct {
	segment   *segment
	highwater *slot

	slots      [MAX_SLOTS + 1]*slot
	size       int
	highpassed bool
	//   unsigned short m_precontext;
	//   int            m_maxSize;
	//   uint8          m_dir;
}
