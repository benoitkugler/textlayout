package graphite

type passtype uint8

const (
	PASS_TYPE_UNKNOWN passtype = iota
	PASS_TYPE_LINEBREAK
	PASS_TYPE_SUBSTITUTE
	PASS_TYPE_POSITIONING
	PASS_TYPE_JUSTIFICATION
)

type silfPass struct {
	Ranges           []passRange
	ruleMap          [][]uint16 // with length NumSuccess
	startStates      []int16
	ruleSortKeys     []uint16   // with length numRules
	rulePreContext   []uint8    // with length numRules
	stateTransitions [][]uint16 // with length NumTransitional * NumColumns
	passConstraints  []byte
	ruleConstraints  [][]byte // with length numRules
	actions          [][]byte // with length numRules
	silfPassHeader
	collisionThreshold uint8
}

func (pass *silfPass) isReverseDir() bool {
	return (pass.Flags>>5)&0x1 != 0
}

func (pass *silfPass) collisionLoops() uint8 {
	return pass.Flags & 0x7
}

func (pass *silfPass) kernColls() uint8 {
	return (pass.Flags >> 3) & 0x3
}

// func (pass *silfPass) runGraphite(m *machine, fsm *FiniteStateMachine, reverse bool) bool {
// 	s := m.slotMap().segment.first()
// 	if !s || !pass.testPassConstraint(m) {
// 		return true
// 	}
// 	if reverse {
// 		m.slotMap().segment.reverseSlots()
// 		s = m.slotMap().segment.first()
// 	}
// 	if pass.NumRules != 0 {
// 		currHigh := s.next()

// 		if debugMode > 1 {
// 			fmt.Println("rules keys:", s.ruleSortKeys)
// 		}

// 		m.slotMap().highwater(currHigh)
// 		lc := pass.MaxRuleLoop
// 		decreaseLc := func() byte {
// 			lc -= 1
// 			return lc
// 		}
// 		for do := true; do; do = s != nil {
// 			pass.findNDoRule(s, m, fsm)
// 			if m.status() != Machine__finished {
// 				return false
// 			}
// 			if s && (s == m.slotMap().highwater() || m.slotMap().highpassed() || decreaseLc() == 0) {
// 				if lc == 0 {
// 					s = m.slotMap().highwater()
// 				}
// 				lc = pass.MaxRuleLoop
// 				if s {
// 					m.slotMap().highwater(s.next())
// 				}
// 			}
// 		}
// 	}

// 	collisions := pass.collisionLoops() != 0 || pass.kernColls() != 0

// 	if !collisions || !m.slotMap().segment.hasCollisionInfo() {
// 		return true
// 	}

// 	if pass.collisionLoops() != 0 {
// 		if !(m.slotMap().segment.flags() & Segment__SEG_INITCOLLISIONS) {
// 			m.slotMap().segment.positionSlots(0, 0, 0, m.slotMap().dir(), true)
// 			//            m.slotMap().segment.flags(m.slotMap().segment.flags() | Segment::SEG_INITCOLLISIONS);
// 		}
// 		if !pass.collisionShift(&m.slotMap().segment, m.slotMap().dir(), fsm.dbgout) {
// 			return false
// 		}
// 	}
// 	if (pass.kernColls()) && !pass.collisionKern(&m.slotMap().segment, m.slotMap().dir(), fsm.dbgout) {
// 		return false
// 	}
// 	if collisions && !pass.collisionFinish(&m.slotMap().segment, fsm.dbgout) {
// 		return false
// 	}
// 	return true
// }

// func (pass *silfPass) loadPassConstraint()

// func (pass *silfPass) testPassConstraint(m *machine) bool {
// 	if !m_cPConstraint {
// 		return true
// 	}

// 	assert(m_cPConstraint.constraint())

// 	m.slotMap().reset(*m.slotMap().segment.first(), 0)
// 	m.slotMap().pushSlot(m.slotMap().segment.first())
// 	map_ := m.slotMap().begin()
// 	ret := m_cPConstraint.run(m, map_)

// 	if debugMode > 1 {
// 		fmt.Println("constraint", ret != 0 && m.status() == Machine__finished)
// 	}

// 	return ret != 0 && m.status() == Machine__finished
// }
