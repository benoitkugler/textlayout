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

// encode the actions to apply to the input string
// it is directly obtained from the font file
type pass struct {
	// assign column to a subset of the glyph indices (GID . column; column < NumColumns)
	constraint    *code // optional
	columns       []uint16
	rules         []rule
	successStates [][]uint16 // (state index - numSuccess) . rule numbers (index into `rules`)
	startStates   []uint16
	transitions   [][]uint16 // each sub array has length NumColums

	collisionThreshold float32
	isReverseDirection bool
	collisionLoops     uint8
	kerningColls       uint8

	numStates                    uint16
	maxPreContext, minPreContext uint16
	maxRuleLoop                  uint8
}

// sanitizes and interprets one pass subtable
func newPass(tablePass *silfPass, context codeContext) (out pass, err error) {
	out.isReverseDirection = (tablePass.Flags>>5)&0x1 != 0
	out.collisionLoops = tablePass.Flags & 0x7
	out.kerningColls = (tablePass.Flags >> 3) & 0x3
	out.collisionThreshold = float32(tablePass.collisionThreshold)

	out.maxPreContext, out.minPreContext = uint16(tablePass.maxRulePreContext), uint16(tablePass.minRulePreContext)
	out.startStates = tablePass.startStates
	out.numStates = tablePass.NumRows
	out.transitions = tablePass.stateTransitions
	out.maxRuleLoop = tablePass.MaxRuleLoop

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

	if len(tablePass.passConstraint) != 0 {
		context.Pt = PASS_TYPE_UNKNOWN
		constraint, err := newCode(true, tablePass.passConstraint, tablePass.rulePreContext[0], tablePass.ruleSortKeys[0], context)
		if err != nil {
			return out, fmt.Errorf("invalid silf pass constraint: %s", err)
		}
		out.constraint = &constraint
	}

	return out, nil
}

func (pass *pass) testPassConstraint(m *machine) (bool, error) {
	if pass.constraint == nil {
		return true, nil
	}

	m.map_.reset(m.map_.segment.First, 0)
	m.map_.pushSlot(m.map_.segment.First)
	map_ := m.map_.getSlice(0)
	ret, err := m.run(pass.constraint, map_)

	if debugMode > 1 {
		fmt.Println("constraint", ret != 0 && err == nil)
	}

	return ret != 0 && err == nil, err
}

func (pa *pass) findAndDoRule(slot *Slot, m *machine, fsm *finiteStateMachine) (*Slot, error) {
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

func (pass *pass) runFSM(fsm *finiteStateMachine, slot *Slot) bool {
	fsm.reset(slot, pass.maxPreContext)
	if fsm.slots.preContext < uint16(pass.minPreContext) {
		return false
	}

	state := pass.startStates[pass.maxPreContext-fsm.slots.preContext]
	var freeSlots uint8 = MAX_SLOTS
	successStart := pass.numStates - uint16(len(pass.successStates)) // order checked in silfPassHeader.sanitize
	for do := true; do; do = state != 0 && slot != nil {
		fsm.slots.pushSlot(slot)
		if int(slot.GlyphID) >= len(pass.columns) || pass.columns[slot.GlyphID] == 0xffff ||
			decrease(&freeSlots) == 0 || int(state) >= len(pass.transitions) {
			return freeSlots != 0
		}

		transitions := pass.transitions[state]
		state = transitions[pass.columns[slot.GlyphID]]
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
			slot = smap.segment.First
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
	coll *shiftCollider, isRev, isRTL bool, moved, hasCol *bool) (fixed bool) {
	var nbor *Slot // neighboring slot
	cFix := seg.getCollisionInfo(slotFix)
	if !coll.initSlot(seg, slotFix, cFix.limit, float32(cFix.margin), float32(cFix.marginWt),
		cFix.shift, cFix.offset, isRTL) {
		return false
	}
	collides := false
	// When we're processing forward, ignore kernable glyphs that preceed the target glyph.
	// When processing backward, don't ignore these until we pass slotFix.
	ignoreForKern := !isRev
	base := slotFix.findRoot()

	// Look for collisions with the neighboring glyphs.
	for nbor = start; nbor != nil; {
		cNbor := seg.getCollisionInfo(nbor)
		sameCluster := nbor.isChildOf(base)
		if nbor != slotFix && // don't process if this is the slot of interest
			!(cNbor.ignore()) && // don't process if ignoring
			(nbor == base || sameCluster || // process if in the same cluster as slotFix
				!inKernCluster(seg, nbor)) && // or this cluster is not to be kerned || (isRTL ^ ignoreForKern))       // or it comes before(ltr) or after(isRTL)
			(!isRev || // if processing forwards then good to merge otherwise only:
				!(cNbor.flags&COLL_FIX != 0) || // merge in immovable stuff
				((cNbor.flags&COLL_KERN != 0) && !sameCluster) || // ignore other kernable clusters
				(cNbor.flags&COLL_ISCOL != 0)) && // test against other collided glyphs
			!coll.mergeSlot(seg, nbor, cNbor, cNbor.shift, !ignoreForKern, sameCluster, false, &collides) {
			return false
		} else if nbor == slotFix {
			// Switching sides of this glyph - if we were ignoring kernable stuff before, don't anymore.
			ignoreForKern = !ignoreForKern
		}

		collConst := COLL_END
		if isRev {
			collConst = COLL_START
		}
		if nbor != start && (cNbor.flags&collConst != 0) {
			break
		}

		if isRev {
			nbor = nbor.prev
		} else {
			nbor = nbor.Next
		}
	}
	isCol := false
	if collides || cFix.shift.X != 0. || cFix.shift.Y != 0. {
		var shift Position
		shift, isCol = coll.resolve(seg)
		// isCol has been set to true if a collision remains.
		if abs(shift.X) < 1e38 && abs(shift.Y) < 1e38 {
			if sqr(shift.X-cFix.shift.X)+sqr(shift.Y-cFix.shift.Y) >= sqr(pass.collisionThreshold) {
				*moved = true
			}
			cFix.shift = shift
			if slotFix.child != nil {
				var bbox rect
				here := slotFix.Position.add(shift)
				clusterMin := here.X
				slotFix.child.finalise(seg, nil, here, &bbox, 0, &clusterMin, isRTL, false, 0)
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
	*hasCol = *hasCol || isCol
	return true
}

func (pass *pass) collisionShift(seg *Segment, isRTL bool) bool {
	var shiftcoll shiftCollider
	// bool isfirst = true;
	hasCollisions := false
	start := seg.First // turn on collision fixing for the first slot
	var end *Slot
	moved := false

	// #if !defined GRAPHITE2_NTRACING
	//     if (dbgout)
	//         *dbgout << "collisions" << json::array
	//             << json::flat << json::object << "num-loops" << m_numCollRuns << json::close;
	// #endif

	for start != nil {
		// #if !defined GRAPHITE2_NTRACING
		//         if (dbgout)  *dbgout << json::object << "phase" << "1" << "moves" << json::array;
		// #endif
		hasCollisions = false
		end = nil
		// phase 1 : position shiftable glyphs, ignoring kernable glyphs
		for s := start; s != nil; s = s.Next {
			c := seg.getCollisionInfo(s)
			if start != nil && (c.flags&(COLL_FIX|COLL_KERN)) == COLL_FIX && !pass.resolveCollisions(seg, s, start, &shiftcoll, false, isRTL, &moved, &hasCollisions) {
				return false
			}
			if s != start && (c.flags&COLL_END) != 0 {
				end = s.Next
				break
			}
		}

		// #if !defined GRAPHITE2_NTRACING
		//         if (dbgout)
		//             *dbgout << json::close << json::close; // phase-1
		// #endif

		// phase 2 : loop until happy.
		for i := 0; i < int(pass.collisionLoops)-1; i++ {
			if hasCollisions || moved {

				// #if !defined GRAPHITE2_NTRACING
				//                 if (dbgout)
				//                     *dbgout << json::object << "phase" << "2a" << "loop" << i << "moves" << json::array;
				// #endif
				// phase 2a : if any shiftable glyphs are in collision, iterate backwards,
				// fixing them and ignoring other non-collided glyphs. Note that this handles ONLY
				// glyphs that are actually in collision from phases 1 or 2b, and working backwards
				// has the intended effect of breaking logjams.
				if hasCollisions {
					hasCollisions = false
					// #if 0
					// moved = true;
					// for (Slot *s = start; s != end; s = s.Next)
					// {
					//     SlotCollision * c = seg.collisionInfo(s);
					//     c.setShift(Position(0, 0));
					// }
					// #endif
					lend := seg.last
					if end != nil {
						lend = end.prev
					}
					lstart := start.prev
					for s := lend; s != lstart; s = s.prev {
						c := seg.getCollisionInfo(s)
						if start != nil && (c.flags&(COLL_FIX|COLL_KERN|COLL_ISCOL)) == (COLL_FIX|COLL_ISCOL) { // ONLY if this glyph is still colliding
							if !pass.resolveCollisions(seg, s, lend, &shiftcoll, true, isRTL, &moved, &hasCollisions) {
								return false
							}
							c.flags = c.flags | COLL_TEMPLOCK
						}
					}
				}

				// #if !defined GRAPHITE2_NTRACING
				//                 if (dbgout)
				//                     *dbgout << json::close << json::close // phase 2a
				//                         << json::object << "phase" << "2b" << "loop" << i << "moves" << json::array;
				// #endif

				// phase 2b : redo basic diacritic positioning pass for ALL glyphs. Each successive loop adjusts
				// glyphs from their current adjusted position, which has the effect of gradually minimizing the
				// resulting adjustment; ie, the final result will be gradually closer to the original location.
				// Also it allows more flexibility in the final adjustment, since it is moving along the
				// possible 8 vectors from successively different starting locations.
				if moved {
					moved = false
					for s := start; s != end; s = s.Next {
						c := seg.getCollisionInfo(s)
						if start != nil && (c.flags&(COLL_FIX|COLL_TEMPLOCK|COLL_KERN)) == COLL_FIX &&
							!pass.resolveCollisions(seg, s, start, &shiftcoll, false, isRTL, &moved, &hasCollisions) {
							return false
						} else if c.flags&COLL_TEMPLOCK != 0 {
							c.flags = c.flags & ^COLL_TEMPLOCK
						}
					}
				}
				//      if (!hasCollisions) // no, don't leave yet because phase 2b will continue to improve things
				//          break;
				// #if !defined GRAPHITE2_NTRACING
				//                 if (dbgout)
				//                     *dbgout << json::close << json::close; // phase 2
				// #endif
			}
		}
		if end == nil {
			break
		}
		start = nil
		for s := end.prev; s != nil; s = s.Next {
			if seg.getCollisionInfo(s).flags&COLL_START != 0 {
				start = s
				break
			}
		}
	}
	return true
}

func (pass *pass) collisionKern(seg *Segment, isRTL bool) bool {
	start := seg.First
	var (
		ymin float32 = 1e38
		ymax float32 = -1e38
	)

	// phase 3 : handle kerning of clusters
	// #if !defined GRAPHITE2_NTRACING
	//     if (dbgout)
	//         *dbgout << json::object << "phase" << "3" << "moves" << json::array;
	// #endif

	for s := seg.First; s != nil; s = s.Next {
		if int(s.GlyphID) >= len(seg.face.glyphs) {
			return false
		}
		c := seg.getCollisionInfo(s)
		bbox := seg.face.getGlyph(s.GlyphID).bbox
		y := s.Position.Y + c.shift.Y
		if c.flags&COLL_ISSPACE == 0 {
			ymax = max(y+bbox.tr.Y, ymax)
			ymin = min(y+bbox.bl.Y, ymin)
		}
		if start != nil && (c.flags&(COLL_KERN|COLL_FIX)) == (COLL_KERN|COLL_FIX) {
			pass.resolveKern(seg, s, start, isRTL, &ymin, &ymax)
		}
		if c.flags&COLL_END != 0 {
			start = nil
		}
		if c.flags&COLL_START != 0 {
			start = s
		}
	}

	// #if !defined GRAPHITE2_NTRACING
	//     if (dbgout)
	//         *dbgout << json::close << json::close; // phase 3
	// #endif
	return true
}

const (
	KernNone = iota
	KernCrossSpace
	KernInWord
	// Kernreserved
)

func (pass *pass) resolveKern(seg *Segment, slotFix, start *Slot, isRTL bool, ymin, ymax *float32) float32 {
	var currSpace float32
	collides := false
	space_count := 0
	base := slotFix.findRoot()
	cFix := seg.getCollisionInfo(base)
	// const GlyphCache &gc = seg.getFace().glyphs();
	bbb := seg.face.getGlyph(slotFix.GlyphID).bbox
	by := slotFix.Position.Y + cFix.shift.Y

	if base != slotFix {
		cFix.flags = cFix.flags | COLL_KERN | COLL_FIX
		return 0
	}
	seenEnd := (cFix.flags & COLL_END) != 0
	isInit := false
	coll := newKernCollider()

	*ymax = max(by+bbb.tr.Y, *ymax)
	*ymin = min(by+bbb.bl.Y, *ymin)
	for nbor := slotFix.Next; nbor != nil; nbor = nbor.Next {
		if nbor.isChildOf(base) {
			continue
		}
		if int(nbor.GlyphID) >= len(seg.face.glyphs) {
			return 0.
		}
		bb := seg.face.getGlyph(nbor.GlyphID).bbox
		cNbor := seg.getCollisionInfo(nbor)
		if (bb.bl.Y == 0. && bb.tr.Y == 0.) || (cNbor.flags&COLL_ISSPACE) != 0 {
			if pass.kerningColls == KernInWord {
				break
			}
			// Add space for a space glyph.
			currSpace += nbor.Advance.X
			space_count++
		} else {
			space_count = 0
			if nbor != slotFix && !cNbor.ignore() {
				seenEnd = true
				if !isInit {
					if !coll.initSlot(seg, slotFix, cFix.limit, float32(cFix.margin),
						cFix.shift, cFix.offset, isRTL, *ymin, *ymax) {
						return 0.
					}
					isInit = true
				}
				maybeCollide := coll.mergeSlot(seg, nbor, cNbor.shift, currSpace, isRTL)
				collides = collides || maybeCollide
			}
		}
		if cNbor.flags&COLL_END != 0 {
			if seenEnd && space_count < 2 {
				break
			} else {
				seenEnd = true
			}
		}
	}
	if collides {
		mv := coll.resolve(isRTL)
		coll.shift(mv, isRTL)
		delta := slotFix.Advance.add(mv).sub(cFix.shift)
		slotFix.Advance = delta
		cFix.shift = mv
		return mv.X
	}
	return 0.
}

func (pass *pass) collisionFinish(seg *Segment) {
	for s := seg.First; s != nil; s = s.Next {
		c := seg.getCollisionInfo(s)
		if c.shift.X != 0 || c.shift.Y != 0 {
			newOffset := c.shift
			var nullPosition Position
			c.offset = newOffset.add(c.offset)
			c.shift = nullPosition
		}
	}
	//    seg.positionSlots();

	// #if !defined GRAPHITE2_NTRACING
	//         if (dbgout)
	//             *dbgout << json::close;
	// #endif
}

func (pass *pass) runGraphite(m *machine, fsm *finiteStateMachine, reverse bool) (bool, error) {
	s := m.map_.segment.First
	if s == nil {
		return true, nil
	}
	if ok, err := pass.testPassConstraint(m); !ok {
		return true, err
	}
	if reverse {
		m.map_.segment.reverseSlots()
		s = m.map_.segment.First
	}
	if len(pass.rules) != 0 {
		currHigh := s.Next

		if debugMode > 1 {
			fmt.Println("rules keys:", fsm.rules)
		}

		m.map_.highwater = currHigh
		lc := pass.maxRuleLoop

		for do := true; do; do = s != nil {
			if _, err := pass.findAndDoRule(s, m, fsm); err != nil {
				return false, err
			}
			if s != nil && (s == m.map_.highwater || m.map_.highpassed || decrease(&lc) == 0) {
				if lc == 0 {
					s = m.map_.highwater
				}
				lc = pass.maxRuleLoop
				if s != nil {
					m.map_.highwater = s.Next
				}
			}
		}
	}

	collisions := pass.collisionLoops != 0 || pass.kerningColls != 0

	if !collisions || !m.map_.segment.hasCollisionInfo() {
		return true, nil
	}

	if pass.collisionLoops != 0 {
		if (m.map_.segment.flags & SEG_INITCOLLISIONS) == 0 {
			m.map_.segment.positionSlots(nil, nil, nil, m.map_.isRTL, true)
			//            m.map_.segment.flags(m.map_.segment.flags | Segment::SEG_INITCOLLISIONS);
		}
		if !pass.collisionShift(m.map_.segment, m.map_.isRTL) {
			return false, nil
		}
	}
	if (pass.kerningColls != 0) && !pass.collisionKern(m.map_.segment, m.map_.isRTL) {
		return false, nil
	}
	if collisions {
		pass.collisionFinish(m.map_.segment)
	}
	return true, nil
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
	attrCollision      byte  // Glyph attribute number for collision.flags attribute (several more collision attrs come after it...)

	indexBidiPass byte // (0xFF) means no bidi pass
	indexPosPass  byte // index of the first positionning pass
	hasCollision  bool
	isRTL         bool
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
	out.attrCollision = silf.AttrCollisions

	out.indexBidiPass = silf.IBidi
	out.indexPosPass = silf.IPos
	out.hasCollision = silf.Flags&0x20 != 0
	out.isRTL = silf.Direction&1 != 0
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

func (s *passes) runGraphite(seg *Segment, firstPass, lastPass uint8, doBidi bool) bool {
	maxSize := len(seg.charinfo) * MAX_SEG_GROWTH_FACTOR

	map_ := newSlotMap(seg, s.isRTL, maxSize)
	fsm := &finiteStateMachine{slots: map_}
	m := newMachine(map_)

	lbidi := s.indexBidiPass

	if lastPass == 0 {
		if firstPass == lastPass && lbidi == 0xFF {
			return true
		}
		lastPass = uint8(len(s.passes))
	}
	if (firstPass < lbidi || (doBidi && firstPass == lbidi)) && (lastPass >= lbidi || (doBidi && lastPass+1 == lbidi)) {
		lastPass++
	} else {
		lbidi = 0xFF
	}

	for i := firstPass; i < lastPass; i++ {
		// bidi and mirroring
		if i == lbidi {

			if debugMode >= 1 {
				fmt.Printf("pass %d, segment direction %v", i, seg.currdir())
			}

			if seg.currdir() != s.isRTL {
				seg.reverseSlots()
			}
			if mirror := s.attrMirroring; mirror != 0 && (seg.dir&3) == 3 {
				seg.doMirror(mirror)
			}
			i--
			lbidi = lastPass
			lastPass--
			continue
		}

		// #if !defined GRAPHITE2_NTRACING
		//         if (dbgout)
		//         {
		//             *dbgout << json::item << json::object
		// //						<< "pindex" << i   // for debugging
		//                         << "id"     << i+1
		//                         << "slotsdir" << (seg.currdir() ? "rtl" : "ltr")
		//                         << "passdir" << ((s.Direction & 1) ^ s.passes[i].isReverseDir() ? "rtl" : "ltr")
		//                         << "slots"  << json::array;
		//             seg.positionSlots(0, 0, 0, seg.currdir());
		//             for(Slot * s = seg.first(); s; s = s.next())
		//                 *dbgout     << dslot(seg, s);
		//             *dbgout         << json::close;
		//         }
		// #endif

		// test whether to reorder, prepare for positioning
		reverse := (lbidi == 0xFF) && (seg.currdir() != (s.isRTL != s.passes[i].isReverseDirection))
		var err error
		if i >= 32 || (seg.passBits&(1<<i)) == 0 || s.passes[i].collisionLoops != 0 {
			var ok bool
			ok, err = s.passes[i].runGraphite(m, fsm, reverse)
			if !ok {
				return false
			}
		}
		// only subsitution passes can change segment length, cached subsegments are short for their text
		if err != nil || (len(seg.charinfo) != 0 && len(seg.charinfo) > maxSize) {
			return false
		}
	}
	return true
}