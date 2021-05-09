package graphite

const (
	//  COLL_TESTONLY = 0  // default - test other glyphs for collision with this one, but don't move this one
	COLL_FIX      uint16 = 1 << iota // fix collisions involving this glyph
	COLL_IGNORE                      // ignore this glyph altogether
	COLL_START                       // start of range of possible collisions
	COLL_END                         // end of range of possible collisions
	COLL_KERN                        // collisions with this glyph are fixed by adding kerning space after it
	COLL_ISCOL                       // this glyph has a collision
	COLL_KNOWN                       // we've figured out what's happening with this glyph
	COLL_ISSPACE                     // treat this glyph as a space with regard to kerning
	COLL_TEMPLOCK                    // Lock glyphs that have been given priority positioning
	////COLL_JUMPABLE = 128    // moving glyphs may jump this stationary glyph in any direction - DELETE
	////COLL_OVERLAP = 256    // use maxoverlap to restrict - DELETE
)

// slot attributes related to collision-fixing
type slotCollision struct {
	limit        rect
	shift        Position // adjustment within the given pass
	offset       Position // total adjustment for collisions
	exclOffset   Position
	margin       uint16
	marginWt     uint16
	flags        uint16
	exclGlyph    uint16
	seqClass     uint16
	seqProxClass uint16
	seqOrder     uint16
	seqAboveXoff int16
	seqAboveWt   uint16
	seqBelowXlim int16
	seqBelowWt   uint16
	seqValignHt  uint16
	seqValignWt  uint16
}

func (sc *slotCollision) ignore() bool {
	return (sc.flags&COLL_IGNORE) != 0 || (sc.flags&COLL_ISSPACE) != 0
}

type exclusion struct {
	x    float32 // x position
	xm   float32 // xmax position
	c    float32 // constant + sum(MiXi^2)
	sm   float32 // sum(Mi)
	smx  float32 // sum(MiXi)
	open bool
}

func newExclusionWeightedXY(xmin, xmax, f, a0, m, xi, c float32) exclusion {
	return exclusion{
		x: xmin, xm: xmax,
		sm:  m + f,
		smx: m * xi,
		c:   m*xi*xi + f*a0*a0 + c,
	}
}

func newExclusionWeightedSD(xmin, xmax, f, a0,
	m, xi, ai, c float32, nega bool) exclusion {
	xia := xi + ai
	if nega {
		xia = xi - ai
	}
	return exclusion{
		x: xmin, xm: xmax,
		sm:  0.25 * (m + 2.*f),
		smx: 0.25 * m * xia,
		c:   0.25*(m*xia*xia+2.*f*a0*a0) + c,
	}
}

// represents the possible movement of a given glyph in a given direction
// (horizontally, vertically, or diagonally).
// A vector is needed to represent disjoint ranges, eg, -300..-150, 20..200, 500..750.
// Each pair represents the min/max of a sub-range.
type zones struct {
	exclusions                           []exclusion
	margin_len, margin_weight, pos, posm float32
}

func (zo *zones) initialise(xmin, xmax, margin_len,
	margin_weight, a0 float32, isXY bool) {
	zo.margin_len = margin_len
	zo.margin_weight = margin_weight
	zo.pos = xmin
	zo.posm = xmax
	zo.exclusions = zo.exclusions[:0]
	var ex exclusion
	if isXY {
		ex = newExclusionWeightedXY(xmin, xmax, 1, a0, 0, 0, 0)
	} else {
		ex = newExclusionWeightedSD(xmin, xmax, 1, a0, 0, 0, 0, 0, false)
	}
	zo.exclusions = append(zo.exclusions, ex)
	zo.exclusions[0].open = true
}

type shiftCollider struct {
	target *Slot // the glyph to fix

	ranges       [4]zones // possible movements in 4 directions (horizontally, vertically, diagonally)
	len          [4]float32
	limit        rect
	currShift    Position
	currOffset   Position
	origin       Position // Base for all relative calculations
	margin       float32
	marginWt     float32
	seqClass     uint16
	seqProxClass uint16
	seqOrder     uint16
}

// initialize the Collider to hold the basic movement limits for the
// target slot, the one we are focusing on fixing.
func (sc *shiftCollider) initSlot(seg *Segment, aSlot *Slot, limit rect, margin, marginWeight float32,
	currShift, currOffset Position, dir int) bool {
	// int i;
	// float mx, mn;
	// float a, shift;
	// const GlyphCache &gc = seg.getFace().glyphs();
	gid := aSlot.glyphID
	if !gc.check(gid) {
		return false
	}
	bb := gc.getBoundingBBox(gid)
	sb := gc.getBoundingSlantBox(gid)
	// float sx = aSlot.origin().x + currShift.x;
	// float sy = aSlot.origin().y + currShift.y;
	if currOffset.x != 0. || currOffset.y != 0. {
		sc.limit = rect{limit.bl.sub(currOffset), limit.tr.sub(currOffset)}
	} else {
		sc.limit = limit
	}

	// For a ShiftCollider, these indices indicate which vector we are moving by:
	// each sc.ranges represents absolute space with respect to the origin of the slot.
	// Thus take into account true origins but subtract the vmin for the slot
	// case 0: // x direction
	mn := sc.limit.bl.x + currOffset.x
	mx := sc.limit.tr.x + currOffset.x
	sc.len[0] = bb.xa - bb.xi
	a := currOffset.y + currShift.y
	sc.ranges[0].initialise(mn, mx, margin, marginWeight, a, true)
	// case 1: // y direction
	mn = sc.limit.bl.y + currOffset.y
	mx = sc.limit.tr.y + currOffset.y
	sc.len[1] = bb.ya - bb.yi
	a = currOffset.x + currShift.x
	sc.ranges[1].initialise(mn, mx, margin, marginWeight, a, true)
	// case 2: // sum (negatively sloped diagonal boundaries)
	// pick closest x,y limit boundaries in s direction
	shift := currOffset.x + currOffset.y + currShift.x + currShift.y
	mn = -2*min(currShift.x-sc.limit.bl.x, currShift.y-sc.limit.bl.y) + shift
	mx = 2*min(sc.limit.tr.x-currShift.x, sc.limit.tr.y-currShift.y) + shift
	sc.len[2] = sb.sa - sb.si
	a = currOffset.x - currOffset.y + currShift.x - currShift.y
	sc.ranges[2].initialise(mn, mx, margin/ISQRT2, marginWeight, a, false)
	// case 3: // diff (positively sloped diagonal boundaries)
	// pick closest x,y limit boundaries in d direction
	shift = currOffset.x - currOffset.y + currShift.x - currShift.y
	mn = -2*min(currShift.x-sc.limit.bl.x, sc.limit.tr.y-currShift.y) + shift
	mx = 2*min(sc.limit.tr.x-currShift.x, currShift.y-sc.limit.bl.y) + shift
	sc.len[3] = sb.da - sb.di
	a = currOffset.x + currOffset.y + currShift.x + currShift.y
	sc.ranges[3].initialise(mn, mx, margin/ISQRT2, marginWeight, a, false)

	sc.target = aSlot
	if (dir & 1) == 0 {
		// For LTR, switch and negate x limits.
		sc.limit.bl.x = -1 * limit.tr.x
		// sc.limit.tr.x = -1 * limit.bl.x;
	}
	sc.currOffset = currOffset
	sc.currShift = currShift
	sc.origin = aSlot.Position.sub(currOffset) // the original anchor position of the glyph

	sc.margin = margin
	sc.marginWt = marginWeight

	c := seg.getCollisionInfo(aSlot)
	sc.seqClass = c.seqClass
	sc.seqProxClass = c.seqProxClass
	sc.seqOrder = c.seqOrder
	return true
}
