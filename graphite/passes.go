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
	// assign column to a subset of the glyph indices (GID . column; column < NumColumns)
	columns       []uint16
	rules         []rule
	successStates [][]uint16 // (state index - numSuccess) . rule numbers (index into `rules`)
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

func (pa *pass) findNDoRule(slot *Slot, m *machine, fsm *FiniteStateMachine) (*Slot, error) {
	if pa.runFSM(fsm, slot) {
		// Search for the first rule which passes the constraint
		var (
			i int
			r uint16
		)
		for i, r = range fsm.rules {
			ok, err := pa.testConstraint(&fsm.ruleTable[r], m)
			if err != nil {
				return slot, fmt.Errorf("finding rule: %s", err)
			}
			if ok {
				break
			}
		}

		if i < len(fsm.rules) {
			r := fsm.rules[i]
			rule := &fsm.ruleTable[r]
			var (
				adv int32
				err error
			)
			adv, slot, err = pa.doAction(&rule.action, m)
			if err != nil {
				return slot, fmt.Errorf("finding rule: %s", err)
			}
			if rule.action.delete {
				slot = fsm.slots.collectGarbage(slot)
			}
			slot = pa.adjustSlot(adv, slot, &fsm.slots)
			return slot, nil
		}
	}

	slot = slot.Next
	return slot, nil
}

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

func (pass *pass) testConstraint(r *rule, m *machine) (bool, error) {
	currContext := m.map_.preContext
	rulePreContext := uint16(r.preContext)
	if currContext < rulePreContext || int(r.sortKey+currContext-rulePreContext) > m.map_.size {
		return false, nil
	}

	map_ := m.map_.slots[1+currContext-rulePreContext:] // TODO: check slotMap slice access
	if map_[r.sortKey-1] == nil {
		return false, nil
	}

	if len(r.constraint.instrs) == 0 {
		return true, nil
	}
	// assert(r.constraint.constraint())
	for n := r.sortKey; n != 0 && len(map_) != 0; n, map_ = n-1, map_[1:] {
		if map_[0] == nil {
			continue
		}
		ret, err := m.run(&r.constraint, map_)
		if err != nil {
			return false, err
		}
		if ret == 0 {
			return false, nil
		}
	}

	return true, nil
}

func (pass *pass) doAction(code *code, m *machine) (int32, *Slot, error) {
	// assert(codeptr);
	if len(code.instrs) == 0 {
		return 0, nil, nil
	}
	smap := &m.map_
	map_ := smap.getSlice(int(smap.preContext))
	smap.highpassed = false

	ret, err := m.run(code, map_)
	if err != nil {
		smap.highwater = nil
		return 0, nil, err
	}

	return ret, map_[0], nil
}

func (pass *pass) adjustSlot(delta int32, slot *Slot, smap *slotMap) *Slot {
	if slot == nil {
		if smap.highpassed || slot == smap.highwater {
			slot = smap.segment.last
			delta++
			if smap.highwater == nil || smap.highwater == slot {
				smap.highpassed = false
			}
		} else {
			slot = smap.segment.first
			delta--
		}
	}
	if delta < 0 {
		for delta += 1; delta <= 0 && slot != nil; delta++ {
			slot = slot.prev
			if smap.highpassed && smap.highwater == slot {
				smap.highpassed = false
			}
		}
	} else if delta > 0 {
		for delta--; delta >= 0 && slot != nil; delta-- {
			if slot == smap.highwater && slot != nil {
				smap.highpassed = true
			}
			slot = slot.Next
		}
	}

	return slot
}

// Can slot s be kerned, or is it attached to something that can be kerned?
func inKernCluster(seg *Segment, s *Slot) bool {
	c := seg.getCollisionInfo(s)
	if c.flags&COLL_KERN != 0 /** && c.flags & COLL_FIX **/ {
		return true
	}
	for s.parent != nil {
		s = s.parent
		c = seg.getCollisionInfo(s)
		if c.flags&COLL_KERN != 0 /** && c.flags & COLL_FIX **/ {
			return true
		}
	}
	return false
}

// Fix collisions for the given slot.
// Return true if everything was fixed, false if there are still collisions remaining.
// isRev means be we are processing backwards.
func (pass *pass) resolveCollisions(seg *Segment, slotFix, start *Slot,
	coll *shiftCollider, isRev bool, dir int) (fixed, moved, hasCol bool) {
	var nbor *Slot // neighboring slot
	cFix := seg.getCollisionInfo(slotFix)
	if !coll.initSlot(seg, slotFix, cFix.limit, float32(cFix.margin), float32(cFix.marginWt),
		cFix.shift, cFix.offset, dir) {
		return false, false, false
	}
	collides := false
	// When we're processing forward, ignore kernable glyphs that preceed the target glyph.
	// When processing backward, don't ignore these until we pass slotFix.
	ignoreForKern := !isRev
	rtl := dir&1 != 0
	base := slotFix.findRoot()

	// Look for collisions with the neighboring glyphs.
	for nbor = start; nbor != nil; {
		cNbor := seg.getCollisionInfo(nbor)
		sameCluster := nbor.isChildOf(base)
		if nbor != slotFix && // don't process if this is the slot of interest
			!(cNbor.ignore()) && // don't process if ignoring
			(nbor == base || sameCluster || // process if in the same cluster as slotFix
				!inKernCluster(seg, nbor)) && // or this cluster is not to be kerned || (rtl ^ ignoreForKern))       // or it comes before(ltr) or after(rtl)
			(!isRev || // if processing forwards then good to merge otherwise only:
				!(cNbor.flags&COLL_FIX != 0) || // merge in immovable stuff
				((cNbor.flags&COLL_KERN != 0) && !sameCluster) || // ignore other kernable clusters
				(cNbor.flags&COLL_ISCOL != 0)) && // test against other collided glyphs
			!coll.mergeSlot(seg, nbor, cNbor, cNbor.shift(), !ignoreForKern, sameCluster, collides, false) {
			return false, false, false
		} else if nbor == slotFix {
			// Switching sides of this glyph - if we were ignoring kernable stuff before, don't anymore.
			ignoreForKern = !ignoreForKern
		}

		coll_const := COLL_END
		if isRev {
			coll_const = COLL_START
		}
		if nbor != start && (cNbor.flags&coll_const != 0) {
			break
		}

		if isRev {
			nbor = nbor.prev
		} else {
			nbor = nbor.Next
		}
	}
	isCol := false
	if collides || cFix.shift.x != 0. || cFix.shift.y != 0. {
		shift := coll.resolve(seg, isCol, dbgout)
		// isCol has been set to true if a collision remains.
		if std__fabs(shift.x) < 1e38 && std__fabs(shift.y) < 1e38 {
			if sqr(shift.x-cFix.shift.x)+sqr(shift.y-cFix.shift.y) >= m_colThreshold*m_colThreshold {
				moved = true
			}
			cFix.shift = shift
			if slotFix.child != nil {
				var bbox rect
				here := slotFix.Position.add(shift)
				clusterMin := here.x
				slotFix.child.finalise(seg, nil, here, &bbox, 0, &clusterMin, rtl, false, 0)
			}
		}
	} else {
		// This glyph is not colliding with anything.
		// #if !defined GRAPHITE2_NTRACING
		// 	if (dbgout)
		// 	{
		// 		*dbgout << json::object
		// 						<< "missed" << objectid(dslot(seg, slotFix));
		// 		coll.outputJsonDbg(dbgout, seg, -1);
		// 		*dbgout << json::close;
		// 	}
		// #endif
	}

	// Set the is-collision flag bit.
	if isCol {
		cFix.flags = cFix.flags | COLL_ISCOL | COLL_KNOWN
	} else {
		cFix.flags = (cFix.flags & ^COLL_ISCOL) | COLL_KNOWN
	}
	hasCol = hasCol || isCol
	return true, moved, hasCol
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
