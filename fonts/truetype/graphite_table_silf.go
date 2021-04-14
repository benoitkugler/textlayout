package truetype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type graphiteSilfSubtableHeader struct {
	RuleVersion   uint32 // Version of stack-machine language used in rules 3.0 – added
	PassOffset    uint16 // offset of oPasses[0] relative to start of sub-table 3.0 – added
	PseudosOffset uint16 // offset of pMaps[0] relative to start of sub-table 3.0 – added
	MaxGlyphID    uint16 // Maximum valid glyph ID (including line-break & pseudo-glyphs)
	ExtraAscent   int16  // Em-units to be added to the font’s ascent
	ExtraDescent  int16  // Em-units to be added to the font’s descent
	NumPasses     byte   // Number of rendering description passes
	ISubst        byte   // Index of first substitution pass
	IPos          byte   // Index of first Positioning pass
	IJust         byte   // Index of first Justification pass
	IBidi         byte   // Index of first pass after the bidi pass(must be <= iPos); 0xFF implies no bidi pass
	Flags         byte   // Bit 0: True (1) if there is any start-, end-, or cross-line contextualization
	// // Bit 1: True (1) if cross-line contextualization can be
	// // ignored for optimization
	// // Bits 2-4: space contextual flags
	// // Bit 5: automatic collision fixing
	MaxPreContext  byte // Max range for preceding cross-line-boundary contextualization
	MaxPostContext byte // Max range for following cross-line-boundary contextualization
	// // 4.0 – added bit 1
	// // 5.0 – added bit 5 Glyph attribute number that is used for actual glyph
	// // ID for a pseudo glyph
	AttrPsuedo         byte //
	AttrBreakWeight    byte // Glyph attribute number of breakweight attribute
	AttrDirectionality byte // Glyph attribute number for directionality attribute
	AttrMirroring      byte // Glyph attribute number for mirror.glyph
	// // (mirror.isEncoded comes directly after) 2.0 – added;
	// // 4.0 – used
	AttrSkipPasses byte   //  Glyph attribute of bitmap indicating key glyphs for pass optimization 2.0 – added;  4.0 – used
	NumJLevels     byte   //  Number of justification levels; 0 if no justification 2.0 – added Justification -Level jLevels[ ] Justification information for each level. 2.0 – added
	NumLigComp     uint16 //  Number of initial glyph attributes that represent ligature components
	NumUserDefn    byte   //  Number of user-defined slot attributes
	MaxCompPerLig  byte   //  Maximum number of components per ligature
	Direction      byte   //  Supported direction(s)
	AttrCollisions byte   //  Glyph attribute number for collision.flags attribute (several more collision attrs come after it...)

	// 3 bytes reserved

	// // 4.1 – used
	// // 2.0 – added
	// byte numCritFeatures Number of critical features 2.0 – added
	// uint16 critFeatures[ ] Array of critical features 2.0 – added
	// byte reserved
	// // 2.0 – added
	// byte numScriptTag Number of scripts this subtable supports
	// uint32 scriptTag[ ] Array of numScriptTag script tags
	// uint16 lbGID Glyph ID for line-break psuedo-glyph
	// uint32 oPasses[ ] Offets to passes relative to the start of this subtable; numPasses + 1 of these
	// uint16 numPseudo Number of Unicode -> pseudo-glyph mappings
	// uint16 searchPseudo (max power of 2 <= numPseudo) * sizeof(PseudoMap)
	// uint16 pseudoSelector log 2 (max power of 2<= numPseudo)
	// uint16 pseudoShift numPseudo - searchPseudo
	// PseudoMap pMaps[ ] Mappings between Unicode and pseudo-glyphs in
	// order of Unicode
	// ClassMap classes Classes object storing replacement classes used in
	// actions
	// SIL_Pass passes[ ] Array of passes
}

func parseTableSilf(data []byte) (interface{}, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid table Silf (EOF)")
	}
	version := binary.BigEndian.Uint32(data)
	if version < 0x00020000 {
		return nil, fmt.Errorf("invalid table Silf version: %x", version)
	}
	endVersion := 4
	if version >= 0x00030000 {
		endVersion += 4
	}
	if len(data) < endVersion+4 {
		return nil, errors.New("invalid table Silf (EOF)")
	}
	fmt.Printf("%x ", version)
	numSub := int(binary.BigEndian.Uint16(data[endVersion:]))
	fmt.Println(endVersion, numSub)
	if len(data) < endVersion+4+numSub*4 {
		return nil, errors.New("invalid table Silf (EOF)")
	}
	offsets := parseUint32s(data[endVersion+4:], numSub)
	// fmt.Println(offsets)
	for i, offset := range offsets {
		if len(data) < int(offset) {
			return nil, fmt.Errorf("invalid Silf subtable offset: %d", offset)
		}
		r := bytes.NewReader(data[offset:])
		var header graphiteSilfSubtableHeader
		err := binary.Read(r, binary.BigEndian, &header)
		if err != nil {
			return nil, fmt.Errorf("invalid Silf subtable: %s", err)
		}

		fmt.Println(i, header)
	}

	return nil, nil
}
