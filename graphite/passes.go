package graphite

import (
	"errors"
	"fmt"
)

type passtype uint8

const (
	PASS_TYPE_UNKNOWN passtype = iota
	PASS_TYPE_LINEBREAK
	PASS_TYPE_SUBSTITUTE
	PASS_TYPE_POSITIONING
	PASS_TYPE_JUSTIFICATION
)

// encode the actions to apply to the input string
// it is directly obtained from the font file
type pass struct {
	// assign column to a subset of the glyph indices (GID -> column; column < NumColumns)
	columns       []uint16
	rules         []rule
	successStates [][]uint16 // (state index - numSuccess) -> rule numbers (index into `rules`)
	startStates   []uint16
	transitions   [][]uint16 // each sub array has length NumColums

	isReverseDirection bool
	collisionLoops     uint8
	kerningColls       uint8

	numStates                    uint16
	maxPreContext, minPreContext uint16
}

// sanitizes and interprets one pass subtable
func newPass(tablePass *silfPass, context codeContext) (out pass, err error) {
	out.isReverseDirection = (tablePass.Flags>>5)&0x1 != 0
	out.collisionLoops = tablePass.Flags & 0x7
	out.kerningColls = (tablePass.Flags >> 3) & 0x3

	out.maxPreContext, out.minPreContext = uint16(tablePass.maxRulePreContext), uint16(tablePass.minRulePreContext)
	out.startStates = tablePass.startStates
	out.numStates = tablePass.NumRows
	out.transitions = tablePass.stateTransitions

	if err = tablePass.sanitize(); err != nil {
		return out, fmt.Errorf("invalid silf pass subtable: %s", err)
	}

	out.columns, err = tablePass.computeColumns()
	if err != nil {
		return out, fmt.Errorf("invalid silf pass columns: %s", err)
	}

	out.rules, err = tablePass.computeRules(context)
	if err != nil {
		return out, fmt.Errorf("invalid silf pass rules: %s", err)
	}

	return out, nil
}

// compute the columns from the ranges
func (pass *silfPass) computeColumns() ([]uint16, error) {
	if len(pass.ranges) == 0 {
		return nil, nil
	}
	numGlyphs := pass.ranges[len(pass.ranges)-1].LastId + 1
	cols := make([]uint16, numGlyphs)
	for i := range cols {
		cols[i] = 0xFFFF
	}
	for _, range_ := range pass.ranges {
		ci := range_.FirstId
		ciEnd := range_.LastId + 1
		col := range_.ColId

		if ci >= ciEnd || ciEnd > numGlyphs || col >= pass.NumColumns {
			return nil, fmt.Errorf("invalid pass range: %v", range_)
		}

		// A glyph must only belong to one column at a time
		for ci != ciEnd && cols[ci] == 0xffff {
			cols[ci] = col
			ci++
		}

		if ci != ciEnd {
			// we exit early, meaning a column was already attributed to a glyph
			return nil, errors.New("invalid pass range")
		}
	}
	return cols, nil
}

// load the code for the rules
func (pass *silfPass) computeRules(context codeContext) ([]rule, error) {
	var err error
	out := make([]rule, pass.NumRules)
	for i := range pass.ruleSortKeys {
		r := rule{
			sortKey:    pass.ruleSortKeys[i],
			preContext: pass.rulePreContext[i],
		}

		if r.preContext > pass.maxRulePreContext || r.preContext < pass.minRulePreContext {
			return nil, fmt.Errorf("invalid rule preContext %d for [%d ... %d]", r.preContext, pass.minRulePreContext, pass.maxRulePreContext)
		}

		r.action, err = newCode(false, pass.actions[i], r.preContext, r.sortKey, context)
		if err != nil {
			return nil, fmt.Errorf("invalid rule action code: %s", err)
		}
		r.constraint, err = newCode(true, pass.ruleConstraints[i], r.preContext, r.sortKey, context)
		if err != nil {
			return nil, fmt.Errorf("invalid rule constraint code: %s", err)
		}

		out[i] = r
	}
	return out, nil
}

// performs the equivalent of --a in C
func decrease(a *uint8) uint8 {
	*a -= 1
	return *a
}

// func (pass *silfPass) runGraphite(m *machine, fsm *FiniteStateMachine, reverse bool) bool {
// 	s := m.map_.segment.first
// 	if s == nil || !pass.testPassConstraint(m) {
// 		return true
// 	}
// 	if reverse {
// 		m.map_.segment.reverseSlots()
// 		s = m.map_.segment.first
// 	}
// 	if pass.NumRules != 0 {
// 		currHigh := s.Next

// 		if debugMode > 1 {
// 			fmt.Println("rules keys:", fsm.rules)
// 		}

// 		m.map_.highwater = currHigh
// 		lc := pass.MaxRuleLoop

// 		for do := true; do; do = s != nil {
// 			pass.findNDoRule(s, m, fsm)
// 			if m.status() != Machine__finished {
// 				return false
// 			}
// 			if s && (s == m.map_.highwater() || m.map_.highpassed() || decrease(&lc) == 0) {
// 				if lc == 0 {
// 					s = m.map_.highwater()
// 				}
// 				lc = pass.MaxRuleLoop
// 				if s {
// 					m.map_.highwater(s.next())
// 				}
// 			}
// 		}
// 	}

// 	collisions := pass.collisionLoops() != 0 || pass.kernColls() != 0

// 	if !collisions || !m.map_.segment.hasCollisionInfo() {
// 		return true
// 	}

// 	if pass.collisionLoops() != 0 {
// 		if !(m.map_.segment.flags() & Segment__SEG_INITCOLLISIONS) {
// 			m.map_.segment.positionSlots(0, 0, 0, m.map_.dir(), true)
// 			//            m.map_.segment.flags(m.map_.segment.flags() | Segment::SEG_INITCOLLISIONS);
// 		}
// 		if !pass.collisionShift(&m.map_.segment, m.map_.dir(), fsm.dbgout) {
// 			return false
// 		}
// 	}
// 	if (pass.kernColls()) && !pass.collisionKern(&m.map_.segment, m.map_.dir(), fsm.dbgout) {
// 		return false
// 	}
// 	if collisions && !pass.collisionFinish(&m.map_.segment, fsm.dbgout) {
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

// 	m.slotMap().reset(*m.slotMap().segment.first, 0)
// 	m.slotMap().pushSlot(m.slotMap().segment.first)
// 	map_ := m.slotMap().begin()
// 	ret := m_cPConstraint.run(m, map_)

// 	if debugMode > 1 {
// 		fmt.Println("constraint", ret != 0 && m.status() == Machine__finished)
// 	}

// 	return ret != 0 && m.status() == Machine__finished
// }

// func (pass *silfPass) findNDoRule(slot *Slot, m *machine, fsm *FiniteStateMachine) {
// assert(slot);
// TODO:
//     if (runFSM(fsm, slot)) {
//         // Search for the first rule which passes the constraint
//         const RuleEntry *        r = fsm.rules.begin(),
//                         * const re = fsm.rules.end();
//         while (r != re && !testConstraint(*r.rule, m))
//         {
//             ++r;
//             if (m.status() != Machine::finished)
//                 return;
//         }

// #if !defined GRAPHITE2_NTRACING
//         if (fsm.dbgout)
//         {
//             if (fsm.rules.size() != 0)
//             {
//                 *fsm.dbgout << json::item << json::object;
//                 dumpRuleEventConsidered(fsm, *r);
//                 if (r != re)
//                 {
//                     const int adv = doAction(r.rule.action, slot, m);
//                     dumpRuleEventOutput(fsm, *r.rule, slot);
//                     if (r.rule.action.deletes()) fsm.slots.collectGarbage(slot);
//                     adjustSlot(adv, slot, fsm.slots);
//                     *fsm.dbgout << "cursor" << objectid(dslot(&fsm.slots.segment, slot))
//                             << json::close; // Close RuelEvent object

//                     return;
//                 }
//                 else
//                 {
//                     *fsm.dbgout << json::close  // close "considered" array
//                             << "output" << json::null
//                             << "cursor" << objectid(dslot(&fsm.slots.segment, slot.next()))
//                             << json::close;
//                 }
//             }
//         }
//         else
// #endif
//         {
//             if (r != re)
//             {
//                 const int adv = doAction(r.rule.action, slot, m);
//                 if (m.status() != Machine::finished) return;
//                 if (r.rule.action.deletes()) fsm.slots.collectGarbage(slot);
//                 adjustSlot(adv, slot, fsm.slots);
//                 return;
//             }
//         }
//     }

//     slot = slot.next();
//     return;
// }

func (pass *pass) runFSM(fsm *FiniteStateMachine, slot *Slot) bool {
	fsm.reset(slot, pass.maxPreContext)
	if fsm.slots.preContext < uint16(pass.minPreContext) {
		return false
	}

	state := pass.startStates[pass.maxPreContext-fsm.slots.preContext]
	var freeSlots uint8 = MAX_SLOTS
	successStart := pass.numStates - uint16(len(pass.successStates)) // order checked in silfPassHeader.sanitize
	for do := true; do; do = state != 0 && slot != nil {
		fsm.slots.pushSlot(slot)
		if int(slot.glyphID) >= len(pass.columns) || pass.columns[slot.glyphID] == 0xffff ||
			decrease(&freeSlots) == 0 || int(state) >= len(pass.transitions) {
			return freeSlots != 0
		}

		transitions := pass.transitions[state]
		state = transitions[pass.columns[slot.glyphID]]
		if state >= successStart {
			fsm.accumulateRules(pass.successStates[state-successStart])
		}

		slot = slot.Next
	}

	fsm.slots.pushSlot(slot)
	return true
}

// higher level version of a silf subtable
type passes struct {
	passes              []pass
	pseudoMaps          []pseudoMap
	justificationLevels []justificationLevel
	classMap            classMap

	userAttibutes      uint8 // Number of user-defined slot attributes
	attrPseudo         byte  // Glyph attribute number that is used for actual glyph ID for a pseudo glyph
	attrBreakWeight    byte  // Glyph attribute number of breakweight attribute
	attrDirectionality byte  // Glyph attribute number for directionality attribute
	attrMirroring      byte  // Glyph attribute number for mirror.glyph (mirror.isEncoded comes directly after)
	attrSkipPasses     byte  // Glyph attribute of bitmap indicating key glyphs for pass optimization

	isRTL bool
}

// interprets and sanitizes the subtable
func newPasses(silf *silfSubtable, numAttributes, numFeatures uint16) (out passes, err error) {
	out.passes = make([]pass, len(silf.passes))

	context := codeContext{
		NumAttributes:     numAttributes,
		NumFeatures:       numFeatures,
		NumClasses:        silf.classMap.numClasses(),
		NumUserAttributes: silf.NumUserDefn,
	}
	for i := range silf.passes {
		pass := &silf.passes[i]

		// resolve the pass type
		context.Pt = PASS_TYPE_UNKNOWN
		switch {
		case i >= int(silf.IJust):
			context.Pt = PASS_TYPE_JUSTIFICATION
		case i >= int(silf.IPos):
			context.Pt = PASS_TYPE_POSITIONING
		case i >= int(silf.ISubst):
			context.Pt = PASS_TYPE_SUBSTITUTE
		default:
			context.Pt = PASS_TYPE_LINEBREAK
		}

		out.passes[i], err = newPass(pass, context)
		if err != nil {
			return out, fmt.Errorf("invalid silf pass %d: %s", i, err)
		}
	}

	out.pseudoMaps = silf.pseudoMap
	out.justificationLevels = silf.justificationLevels
	out.classMap = silf.classMap

	out.userAttibutes = silf.NumUserDefn
	out.attrPseudo = silf.AttrPseudo
	out.attrBreakWeight = silf.AttrBreakWeight
	out.attrDirectionality = silf.AttrDirectionality
	out.attrMirroring = silf.AttrMirroring
	out.attrSkipPasses = silf.AttrSkipPasses

	out.isRTL = silf.Direction != 0
	return out, nil
}

func (s *passes) findPdseudoGlyph(r rune) GID {
	if s == nil {
		return 0
	}
	for _, rec := range s.pseudoMaps {
		if rec.Unicode == r {
			return rec.NPseudo
		}
	}
	return 0
}
