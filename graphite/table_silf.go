package graphite

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type graphiteSilfSubtableHeaderV3 struct {
	RuleVersion   uint32 // Version of stack-machine language used in rules
	PassOffset    uint16 // offset of oPasses[0] relative to start of sub-table
	PseudosOffset uint16 // offset of pMaps[0] relative to start of sub-table
}

type graphiteSilfSubtablePart1 struct {
	MaxGlyphID   uint16 // Maximum valid glyph ID (including line-break & pseudo-glyphs)
	ExtraAscent  int16  // Em-units to be added to the font’s ascent
	ExtraDescent int16  // Em-units to be added to the font’s descent
	NumPasses    byte   // Number of rendering description passes
	ISubst       byte   // Index of first substitution pass
	IPos         byte   // Index of first Positioning pass
	IJust        byte   // Index of first Justification pass
	IBidi        byte   // Index of first pass after the bidi pass(must be <= iPos); 0xFF implies no bidi pass
	// Bit 0: True (1) if there is any start-, end-, or cross-line contextualization
	// Bit 1: True (1) if cross-line contextualization can be ignored for optimization
	// Bits 2-4: space contextual flags
	// Bit 5: automatic collision fixing
	Flags          byte
	MaxPreContext  byte // Max range for preceding cross-line-boundary contextualization
	MaxPostContext byte // Max range for following cross-line-boundary contextualization

	AttrPseudo         byte // Glyph attribute number that is used for actual glyph ID for a pseudo glyph
	AttrBreakWeight    byte // Glyph attribute number of breakweight attribute
	AttrDirectionality byte // Glyph attribute number for directionality attribute
	AttrMirroring      byte // Glyph attribute number for mirror.glyph (mirror.isEncoded comes directly after)
	AttrSkipPasses     byte // Glyph attribute of bitmap indicating key glyphs for pass optimization
	NumJLevels         byte // Number of justification levels; 0 if no justification
}

type graphiteSilfSubtablePart2 struct {
	NumLigComp     uint16 // Number of initial glyph attributes that represent ligature components
	NumUserDefn    byte   // Number of user-defined slot attributes
	MaxCompPerLig  byte   // Maximum number of components per ligature
	Direction      byte   // Supported direction(s)
	AttrCollisions byte   // Glyph attribute number for collision.flags attribute (several more collision attrs come after it...)

	_ [3]byte // reserved
}

func parseTableSilf(data []byte) ([]silfSubtable, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid table Silf (EOF)")
	}
	version := uint16(binary.BigEndian.Uint32(data) >> 16)
	if version < 2 {
		return nil, fmt.Errorf("invalid table Silf version: %x", version)
	}
	endVersion := 4
	if version >= 3 {
		endVersion += 4
	}
	if len(data) < endVersion+4 {
		return nil, errors.New("invalid table Silf (EOF)")
	}

	numSub := int(binary.BigEndian.Uint16(data[endVersion:]))
	if len(data) < endVersion+4+numSub*4 {
		return nil, errors.New("invalid table Silf (EOF)")
	}
	offsets := make([]uint32, numSub)
	_ = binary.Read(bytes.NewReader(data[endVersion+4:]), binary.BigEndian, offsets)

	out := make([]silfSubtable, numSub)
	var err error
	for i, offset := range offsets {
		out[i], err = parseSubtableSilf(data, offset, version)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

type graphitePseudoMap struct {
	Unicode rune
	NPseudo uint16
}

type silfSubtable struct {
	justificationLevels []JustificationLevel
	scriptTags          []uint32
	classMap            classMap
	passes              []silfPass
}

type binSearchHeader struct {
	NumRecord uint16
	_         [3]uint16 // ignored
}

func parseSubtableSilf(data []byte, offset uint32, version uint16) (out silfSubtable, err error) {
	if len(data) < int(offset) {
		return out, fmt.Errorf("invalid Silf subtable offset: %d", offset)
	}
	r := bytes.NewReader(data[offset:])
	var (
		headerv3    graphiteSilfSubtableHeaderV3
		headerPart1 graphiteSilfSubtablePart1
		headerPart2 graphiteSilfSubtablePart2
	)
	if version >= 3 {
		err = binary.Read(r, binary.BigEndian, &headerv3)
		if err != nil {
			return out, fmt.Errorf("invalid Silf subtable: %s", err)
		}
	}
	err = binary.Read(r, binary.BigEndian, &headerPart1)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	out.justificationLevels = make([]JustificationLevel, headerPart1.NumJLevels) // allocation guarded by the uint8 constraint
	err = binary.Read(r, binary.BigEndian, out.justificationLevels)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	err = binary.Read(r, binary.BigEndian, &headerPart2)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	numCritFeatures, err := r.ReadByte() // Number of critical features
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	critFeatures := make([]uint16, numCritFeatures)
	err = binary.Read(r, binary.BigEndian, critFeatures)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	_, _ = r.ReadByte() // byte reserved

	numScriptTag, err := r.ReadByte() // Number of scripts this subtable supports
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	out.scriptTags = make([]uint32, numScriptTag) // Array of numScriptTag script tags
	err = binary.Read(r, binary.BigEndian, out.scriptTags)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	var lbGID uint16 // Glyph ID for line-break psuedo-glyph
	err = binary.Read(r, binary.BigEndian, &lbGID)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	oPasses := make([]uint32, headerPart1.NumPasses+1) // Offets to passes relative to the start of this subtable
	err = binary.Read(r, binary.BigEndian, oPasses)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	var mapsHeader binSearchHeader
	err = binary.Read(r, binary.BigEndian, &mapsHeader)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	pMaps := make([]graphitePseudoMap, mapsHeader.NumRecord) // Mappings between Unicode and pseudo-glyphs in order of Unicode
	err = binary.Read(r, binary.BigEndian, pMaps)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	out.classMap, err = parseGraphiteClassMap(data[len(data)-r.Len():], version)
	if err != nil {
		return out, err
	}

	out.passes = make([]silfPass, headerPart1.NumPasses)
	for i := range out.passes {
		start := oPasses[i]
		out.passes[i], err = parseSilfPass(data, start)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

type JustificationLevel struct {
	attrStretch byte    //  Glyph attribute number for justify.X.stretch
	attrShrink  byte    //  Glyph attribute number for justify.X.shrink
	attrStep    byte    //  Glyph attribute number for justify.X.step
	attrWeight  byte    //  Glyph attribute number for justify.X.weight
	runto       byte    //  Which level starts the next stage
	_           [3]byte // reserved
}

type classMap struct {
	// numClass  uint16
	// numLinear uint16
	// oClass    []uint32      // Array of numClass + 1 offsets to class arrays from the beginning of the class map
	glyphs  []GID                 // Glyphs for linear classes
	lookups []graphiteLookupClass // An array of numClass – numLinear lookups
}

// data starts at the class map
func parseGraphiteClassMap(data []byte, version uint16) (out classMap, err error) {
	if len(data) < 4 {
		return out, errors.New("invalid Silf Class Map (EOF)")
	}
	numClass := binary.BigEndian.Uint16(data)      // Number of replacement classes
	numLinear := binary.BigEndian.Uint16(data[2:]) // Number of linearly stored replacement classes

	offsets := make([]uint32, numClass+1)

	if version >= 4 {
		if len(data) < 4+4*int(numClass+1) {
			return out, errors.New("invalid Silf Class Map (EOF)")
		}
		for i := range offsets {
			offsets[i] = binary.BigEndian.Uint32(data[4+4*i:])
		}
	} else {
		if len(data) < 4+2*int(numClass+1) {
			return out, errors.New("invalid Silf Class Map (EOF)")
		}
		for i := range offsets {
			offsets[i] = uint32(binary.BigEndian.Uint16(data[4+2*i:]))
		}
	}
	if err != nil {
		return out, fmt.Errorf("invalid Silf Class Map: %s", err)
	}

	if numClass < numLinear {
		return out, fmt.Errorf("invalid Silf Class Map (%d < %d)", numClass, numLinear)
	}

	out.glyphs = make([]GID, numLinear)
	for i := range out.glyphs {
		start := int(offsets[i])
		if len(data) < start+2 {
			return out, fmt.Errorf("invalid Silf Class Map offset (%d)", start)
		}
		out.glyphs[i] = binary.BigEndian.Uint16(data)
	}

	out.lookups = make([]graphiteLookupClass, numClass-numLinear)
	r := bytes.NewReader(data)
	for i := range out.lookups {
		offset := int64(offsets[int(numLinear)+i])
		_, _ = r.Seek(offset, io.SeekStart) // delay error checking
		out.lookups[i], err = parseGraphiteLookupClass(r)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

type graphiteLookupPair struct {
	Glyph GID
	Index uint16
}

type graphiteLookupClass []graphiteLookupPair

func parseGraphiteLookupClass(r *bytes.Reader) (graphiteLookupClass, error) {
	var numsIDS uint16
	err := binary.Read(r, binary.BigEndian, &numsIDS)
	if err != nil {
		return nil, fmt.Errorf("invalid Silf Lookup Class: %s", err)
	}
	_, _ = r.Seek(6, io.SeekCurrent)
	out := make(graphiteLookupClass, numsIDS)
	err = binary.Read(r, binary.BigEndian, out)
	if err != nil {
		return nil, fmt.Errorf("invalid Silf Lookup Class: %s", err)
	}
	return out, err
}

type silfPassHeader struct {
	// Bits 0-2: collision fixing max loop
	// Bits 3-4: auto kerning
	// Bit 5: flip direction 5.0 – added
	Flags           byte
	MaxRuleLoop     byte   // MaxRuleLoop for this pass
	MaxRuleContext  byte   // Number of slots of input needed to run this pass
	MaxBackup       byte   // Number of slots by which the following pass needs to trail this pass (ie, the maximum this pass is allowed to back up)
	NumRules        uint16 // Number of action code blocks
	FsmOffset       uint16 // offset to numRows relative to the beginning of the SIL_Pass block 2.0 – inserted ; 3.0 – use for fsmOffset
	PcCode          uint32 // Offset to start of pass constraint code from start of subtable (*passConstraints[0]*) 2.0 - added
	RcCode          uint32 // Offset to start of rule constraint code from start of subtable (*ruleConstraints[0]*)
	ACode           uint32 // Offset to start of action code relative to start of subtable (*actions[0]*)
	ODebug          uint32 // Offset to debug arrays (*dActions[0]*); equals 0 if debug stripped
	NumRows         uint16 // Number of FSM states
	NumTransitional uint16 // Number of transitional states in the FSM (length of *states* matrix)
	NumSuccess      uint16 // Number of success states in the FSM (size of *oRuleMap* array)
	NumColumns      uint16 // Number of FSM columns; 0 means no FSM
}

type silfPass struct {
	Ranges      []passRange
	ruleMap     [][]uint16 // with length NumSuccess
	startStates []int16
	silfPassHeader
}

type passRange struct {
	FirstId GID    // First Glyph id in the range
	LastId  GID    // Last Glyph id in the range
	ColId   uint16 // Column index for this range
}

func parseSilfPass(data []byte, offset uint32) (out silfPass, err error) {
	if len(data) < int(offset)+32 {
		return out, errors.New("invalid Silf Pass offset (EOF)")
	}
	r := bytes.NewReader(data[offset:])
	_ = binary.Read(r, binary.BigEndian, &out.silfPassHeader) // length was checked

	var rangeHeader binSearchHeader
	err = binary.Read(r, binary.BigEndian, &rangeHeader)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	out.Ranges = make([]passRange, rangeHeader.NumRecord)
	err = binary.Read(r, binary.BigEndian, out.Ranges)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	oRuleMap := make([]uint16, out.NumSuccess+1)
	err = binary.Read(r, binary.BigEndian, oRuleMap)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	ruleMapSlice := make([]uint16, oRuleMap[len(oRuleMap)-1])
	err = binary.Read(r, binary.BigEndian, ruleMapSlice)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	out.ruleMap = make([][]uint16, out.NumSuccess)
	for i := range out.ruleMap {
		start, end := oRuleMap[i], oRuleMap[i+1]
		if start > end || int(end) > len(ruleMapSlice) {
			continue
		}
		out.ruleMap[i] = ruleMapSlice[start:end]
	}

	minRulePreContext, err := r.ReadByte() // Minimum number of items in any rule’s context before the first modified rule item
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	maxRulePreContext, err := r.ReadByte() // Maximum number of items in any rule’s context before the first modified rule item
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}
	if maxRulePreContext < minRulePreContext {
		return out, fmt.Errorf("invalid Silf subtable: (%d < %d)", maxRulePreContext, minRulePreContext)
	}
	out.startStates = make([]int16, maxRulePreContext-minRulePreContext+1)
	err = binary.Read(r, binary.BigEndian, out.startStates)
	if err != nil {
		return out, fmt.Errorf("invalid Silf subtable: %s", err)
	}

	return out, nil
}
