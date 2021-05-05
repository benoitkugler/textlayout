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
