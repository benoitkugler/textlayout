package type1c

// code is adapted from golang.org/x/image/font/sfnt

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	ps "github.com/benoitkugler/textlayout/fonts/psinterpreter"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

var (
	errInvalidCFFTable                = errors.New("invalid CFF font file")
	errUnsupportedCFFVersion          = errors.New("unsupported CFF version")
	errUnsupportedRealNumberEncoding  = errors.New("unsupported real number encoding")
	errUnsupportedCFFFDSelectTable    = errors.New("unsupported FD Select version")
	errUnsupportedNumberOfSubroutines = errors.New("unsupported number of subroutines")

	be = binary.BigEndian
)

const (
	// Adobe's SourceHanSansSC-Regular.otf has up to 30000 subroutines.
	maxNumSubroutines = 40000

	// since SID = 0 means .notdef, we use a reserved value
	// to mean unset
	unsetSID = uint16(0xFFFF)
)

type userStrings [][]byte

// return either the predefined string or the user defined one
func (u userStrings) getString(sid uint16) (string, error) {
	if sid == unsetSID {
		return "", nil
	}
	if sid < 391 {
		return stdStrings[sid], nil
	}
	sid -= 391
	if int(sid) >= len(u) {
		return "", errInvalidCFFTable
	}
	return string(u[sid]), nil
}

// Compact Font Format (CFF) fonts are written in PostScript, a stack-based
// programming language.
//
// A fundamental concept is a DICT, or a key-value map, expressed in reverse
// Polish notation. For example, this sequence of operations:
//	- push the number 379
//	- version operator
//	- push the number 392
//	- Notice operator
//	- etc
//	- push the number 100
//	- push the number 0
//	- push the number 500
//	- push the number 800
//	- FontBBox operator
//	- etc
// defines a DICT that maps "version" to the String ID (SID) 379, "Notice" to
// the SID 392, "FontBBox" to the four numbers [100, 0, 500, 800], etc.
//
// The first 391 String IDs (starting at 0) are predefined as per the CFF spec
// Appendix A, in 5176.CFF.pdf referenced below. For example, 379 means
// "001.000". String ID 392 is not predefined, and is mapped by a separate
// structure, the "String INDEX", inside the CFF data. (String ID 391 is also
// not predefined. Specifically for ../testdata/CFFTest.otf, 391 means
// "uni4E2D", as this font contains a glyph for U+4E2D).
//
// The actual glyph vectors are similarly encoded (in PostScript), in a format
// called Type 2 Charstrings. The wire encoding is similar to but not exactly
// the same as CFF's. For example, the byte 0x05 means FontBBox for CFF DICTs,
// but means rlineto (relative line-to) for Type 2 Charstrings. See
// 5176.CFF.pdf Appendix H and 5177.Type2.pdf Appendix A in the PDF files
// referenced below.
//
// The relevant specifications are:
// 	- http://wwwimages.adobe.com/content/dam/Adobe/en/devnet/font/pdfs/5176.CFF.pdf
// 	- http://wwwimages.adobe.com/content/dam/Adobe/en/devnet/font/pdfs/5177.Type2.pdf
type cffParser struct {
	src    []byte // whole input
	offset int    // current position
}

func (p *cffParser) parse() ([]CFF, error) {
	// header was checked prior to this call

	// Parse the Name INDEX.
	fontNames, err := p.parseNames()
	if err != nil {
		return nil, err
	}

	topDicts, err := p.parseTopDicts()
	if err != nil {
		return nil, err
	}
	// 5176.CFF.pdf section 8 "Top DICT INDEX" says that the count here
	// should match the count of the Name INDEX
	if len(topDicts) != len(fontNames) {
		return nil, errInvalidCFFTable
	}

	// parse the String INDEX.
	strs, err := p.parseUserStrings()
	if err != nil {
		return nil, err
	}

	out := make([]CFF, len(topDicts))

	// use the strings to fetch the PSInfo
	for i, topDict := range topDicts {
		out[i].fontName = fontNames[i]
		out[i].PSInfo, err = topDict.toInfo(strs)
		if err != nil {
			return nil, err
		}
		out[i].cidFontName, err = strs.getString(topDict.cidFontName)
		if err != nil {
			return nil, err
		}
	}

	// Parse the Global Subrs [Subroutines] INDEX.
	globSubrsCount, offSize, err := p.parseIndexHeader()
	if err != nil {
		return nil, err
	}
	if globSubrsCount != 0 {
		if globSubrsCount > maxNumSubroutines {
			return nil, errUnsupportedNumberOfSubroutines
		}
		gsubrs := make([]uint32, globSubrsCount+1)
		if err = p.parseIndexLocations(gsubrs, offSize); err != nil {
			return nil, err
		}
	}

	for i, topDict := range topDicts {
		// Parse the CharStrings INDEX, whose location was found in the Top DICT.
		if err = p.seek(topDict.charStringsOffset); err != nil {
			return nil, err
		}
		out[i].charstrings, err = p.parseIndex()
		if err != nil {
			return nil, err
		}
		numGlyphs := uint16(len(out[i].charstrings))

		charset, err := p.parseCharset(topDict.charsetOffset, numGlyphs)
		if err != nil {
			return nil, err
		}

		out[i].Encoding, err = p.parseEncoding(topDict.encodingOffset, numGlyphs, charset, strs)
		if err != nil {
			return nil, err
		}

		if !topDict.isCIDFont {
			// Parse the Private DICT, whose location was found in the Top DICT.
			_, err = p.parsePrivateDICT(topDict.privateDictOffset, topDict.privateDictLength)
			if err != nil {
				return nil, err
			}
		} else {
			// Parse the Font Dict Select data, whose location was found in the Top
			// DICT.
			out[i].fdSelect, err = p.parseFDSelect(topDict.fdSelect, numGlyphs)
			if err != nil {
				return nil, err
			}
		}
	}

	// 	// Parse the Font Dicts. Each one contains its own Private DICT.
	// 	if !p.seek(p.psi.topDict.fdArray) {
	// 		return nil, errInvalidCFFTable
	// 	}

	// 	count, offSize, ok := p.parseIndexHeader()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if count > maxNumFontDicts {
	// 		return nil, errUnsupportedNumberOfFontDicts
	// 	}

	// 	fdLocations := make([]uint32, count+1)
	// 	if !p.parseIndexLocations(fdLocations, count, offSize) {
	// 		return nil, err
	// 	}

	// 	privateDicts := make([]struct {
	// 		offset, length int32
	// 	}, count)

	// 	for i := range privateDicts {
	// 		length := fdLocations[i+1] - fdLocations[i]
	// 		if !p.read(int(length)) {
	// 			return nil, errInvalidCFFTable
	// 		}
	// 		p.psi.topDict.initialize()
	// 		if err = p.psi.run(psContextTopDict, buf, 0, 0); err != nil {
	// 			return nil, err
	// 		}
	// 		privateDicts[i].offset = p.psi.topDict.privateDictOffset
	// 		privateDicts[i].length = p.psi.topDict.privateDictLength
	// 	}

	// 	ret.multiSubrs = make([][]uint32, count)
	// 	for i, pd := range privateDicts {
	// 		ret.multiSubrs[i], err = p.parsePrivateDICT(pd.offset, pd.length)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// }
	return out, nil
}

func (p *cffParser) parseTopDicts() ([]topDictData, error) {
	// Parse the Top DICT INDEX.
	count, offSize, err := p.parseIndexHeader()
	if err != nil {
		return nil, err
	}
	topDictLocations := make([]uint32, count+1)
	if err := p.parseIndexLocations(topDictLocations, offSize); err != nil {
		return nil, err
	}
	out := make([]topDictData, count) // guarded by uint16 max size
	var psi ps.Inter
	for i := range out {
		length := topDictLocations[i+1] - topDictLocations[i]
		buf, err := p.read(int(length))
		if err != nil {
			return nil, err
		}
		topDict := &out[i]

		// set default value before parsing
		topDict.underlinePosition = -100
		topDict.underlineThickness = 50
		topDict.version = unsetSID
		topDict.notice = unsetSID
		topDict.fullName = unsetSID
		topDict.familyName = unsetSID
		topDict.weight = unsetSID
		topDict.cidFontName = unsetSID

		if err = psi.Run(buf, nil, topDict); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// parse the general form of an index
func (p *cffParser) parseIndex() ([][]byte, error) {
	count, offSize, err := p.parseIndexHeader()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	out := make([][]byte, count)

	stringsLocations := make([]uint32, count+1)
	if err := p.parseIndexLocations(stringsLocations, offSize); err != nil {
		return nil, err
	}

	for i := range out {
		length := stringsLocations[i+1] - stringsLocations[i]
		buf, err := p.read(int(length))
		if err != nil {
			return nil, err
		}
		out[i] = buf
	}
	return out, nil
}

// parse the Name INDEX
func (p *cffParser) parseNames() ([][]byte, error) {
	return p.parseIndex()
}

// parse the String INDEX
func (p *cffParser) parseUserStrings() (userStrings, error) {
	index, err := p.parseIndex()
	return userStrings(index), err
}

// Parse the charset data, whose location was found in the Top DICT.
func (p *cffParser) parseCharset(charsetOffset int32, numGlyphs uint16) ([]uint16, error) {
	// Predefined charset may have offset of 0 to 2 // Table 22
	var charset []uint16
	switch charsetOffset {
	case 0: // ISOAdobe
		charset = charsetISOAdobe[:]
	case 1: // Expert
		charset = charsetExpert[:]
	case 2: // ExpertSubset
		charset = charsetExpertSubset[:]
	default: // custom
		if err := p.seek(charsetOffset); err != nil {
			return nil, err
		}
		buf, err := p.read(1)
		if err != nil {
			return nil, err
		}
		charset = make([]uint16, numGlyphs)
		switch buf[0] { // format
		case 0:
			buf, err := p.read(2 * (int(numGlyphs) - 1)) // ".notdef" is omited, and has an implicit SID of 0
			if err != nil {
				return nil, err
			}
			for i := uint16(1); i < numGlyphs; i++ {
				charset[i] = be.Uint16(buf[2*i-2:])
			}
		case 1:
			for i := uint16(1); i < numGlyphs; {
				buf, err := p.read(3)
				if err != nil {
					return nil, err
				}
				first, nLeft := be.Uint16(buf), uint16(buf[2])
				for j := uint16(0); j <= nLeft && i < numGlyphs; j++ {
					charset[i] = first + j
					i++
				}
			}
		case 2:
			for i := uint16(1); i < numGlyphs; {
				buf, err := p.read(4)
				if err != nil {
					return nil, err
				}
				first, nLeft := be.Uint16(buf), be.Uint16(buf[2:])
				for j := uint16(0); j <= nLeft && i < numGlyphs; j++ {
					charset[i] = first + j
					i++
				}
			}
		default:
			return nil, errInvalidCFFTable
		}
	}
	return charset, nil
}

// Parse the encoding data, whose location was found in the Top DICT.
func (p *cffParser) parseEncoding(encodingOffset int32, numGlyphs uint16, charset []uint16, strs userStrings) (*simpleencodings.Encoding, error) {
	// Predefined encoding may have offset of 0 to 1 // Table 16
	switch encodingOffset {
	case 0: // Standard
		return &simpleencodings.Standard, nil
	case 1: // Expert
		return &expertEncoding, nil
	default: // custom
		var encoding simpleencodings.Encoding
		if err := p.seek(encodingOffset); err != nil {
			return nil, err
		}
		buf, err := p.read(2)
		if err != nil {
			return nil, err
		}
		format, size, charL := buf[0], int32(buf[1]), int32(len(charset))
		// high order bit may be set for supplemental encoding data
		switch format & 0xf { // 0111_1111
		case 0:
			if size > int32(numGlyphs) { // truncate
				size = int32(numGlyphs)
			}
			buf, err = p.read(int(size))
			if err != nil {
				return nil, err
			}
			for i := int32(1); i < size && i < charL; i++ {
				c := buf[i-1]
				encoding[c], err = strs.getString(charset[i])
				if err != nil {
					return nil, err
				}
			}
		case 1:
			buf, err = p.read(2 * int(size))
			if err != nil {
				return nil, err
			}
			nCodes := int32(1)
			for i := int32(0); i < size; i++ {
				c, nLeft := buf[2*i], buf[2*i+1]
				for j := byte(0); j < nLeft && nCodes < int32(numGlyphs) && nCodes < charL; j++ {
					encoding[c], err = strs.getString(charset[nCodes])
					if err != nil {
						return nil, err
					}
					nCodes++
					c++
				}
			}
		default:
			return nil, errInvalidCFFTable
		}

		if format&0x80 != 0 { // 1000_000
			nSupsBuf, err := p.read(1)
			if err != nil {
				return nil, err
			}
			buf, err := p.read(3 * int(nSupsBuf[0]))
			if err != nil {
				return nil, err
			}
			for i := byte(0); i < nSupsBuf[0]; i++ {
				code, sid := buf[3*i], be.Uint16(buf[3*i+1:])
				encoding[code], err = strs.getString(sid)
				if err != nil {
					return nil, err
				}
			}
		}
		return &encoding, nil
	}
}

// fdSelect holds a CFF font's Font Dict Select data.
type fdSelect interface {
	fontDictIndex(glyph fonts.GlyphIndex) (byte, error)
}

type fdSelect0 []byte

func (fds fdSelect0) fontDictIndex(glyph fonts.GlyphIndex) (byte, error) {
	if int(glyph) >= len(fds) {
		return 0, errors.New("invalid glyph index")
	}
	return fds[glyph], nil
}

type range3 struct {
	first fonts.GlyphIndex
	fd    byte
}

type fdSelect3 struct {
	ranges   []range3
	sentinel fonts.GlyphIndex // = numGlyphs
}

func (fds fdSelect3) fontDictIndex(x fonts.GlyphIndex) (byte, error) {
	lo, hi := 0, len(fds.ranges)
	for lo < hi {
		i := (lo + hi) / 2
		r := fds.ranges[i]
		xlo := r.first
		if x < xlo {
			hi = i
			continue
		}
		xhi := fds.sentinel
		if i < len(fds.ranges)-1 {
			xhi = fds.ranges[i+1].first
		}
		if xhi <= x {
			lo = i + 1
			continue
		}
		return r.fd, nil
	}
	return 0, errors.New("invalid glyph index")
}

// parseFDSelect parses the Font Dict Select data as per 5176.CFF.pdf section
// 19 "FDSelect".
func (p *cffParser) parseFDSelect(offset int32, numGlyphs uint16) (fdSelect, error) {
	if err := p.seek(offset); err != nil {
		return nil, err
	}
	buf, err := p.read(1)
	if err != nil {
		return nil, err
	}
	switch buf[0] { // format
	case 0:
		if len(p.src) < p.offset+int(numGlyphs) {
			return nil, errInvalidCFFTable
		}
		return fdSelect0(p.src[p.offset : p.offset+int(numGlyphs)]), nil
	case 3:
		buf, err = p.read(2)
		if err != nil {
			return nil, err
		}
		numRanges := be.Uint16(buf)
		if len(p.src) < p.offset+3*int(numRanges)+2 {
			return nil, errInvalidCFFTable
		}
		out := fdSelect3{
			sentinel: fonts.GlyphIndex(numGlyphs),
			ranges:   make([]range3, numRanges),
		}
		for i := range out.ranges {
			// 	buf holds the range [xlo, xhi).
			out.ranges[i].first = fonts.GlyphIndex(be.Uint16(p.src[p.offset+3*i:]))
			out.ranges[i].fd = p.src[p.offset+3*i+2]
		}
		return out, nil
	}
	return nil, errUnsupportedCFFFDSelectTable
}

// Parse Private DICT and the Local Subrs [Subroutines] INDEX
func (p *cffParser) parsePrivateDICT(offset, length int32) ([][]byte, error) {
	if length == 0 {
		return nil, nil
	}
	if err := p.seek(offset); err != nil {
		return nil, err
	}
	buf, err := p.read(int(length))
	if err != nil {
		return nil, err
	}
	var (
		psi  ps.Inter
		priv privateDict
	)
	if err = psi.Run(buf, nil, &priv); err != nil {
		return nil, err
	}

	if priv.subrsOffset == 0 {
		return nil, nil
	}

	// "The local subrs offset is relative to the beginning of the Private DICT data"
	if err = p.seek(offset + priv.subrsOffset); err != nil {
		return nil, errInvalidCFFTable
	}
	subrs, err := p.parseIndex()
	if err != nil {
		return nil, err
	}
	if len(subrs) > maxNumSubroutines {
		return nil, errUnsupportedNumberOfSubroutines
	}
	return subrs, nil
}

// read returns the n bytes from p.offset and advances p.offset by n.
func (p *cffParser) read(n int) ([]byte, error) {
	if n < 0 || len(p.src) < p.offset+n {
		return nil, errInvalidCFFTable
	}
	out := p.src[p.offset : p.offset+n]
	p.offset += n
	return out, nil
}

// skip advances p.offset by n.
func (p *cffParser) skip(n int) error {
	if len(p.src) < p.offset+n {
		return errInvalidCFFTable
	}
	p.offset += n
	return nil
}

func (p *cffParser) seek(offset int32) error {
	if offset < 0 || len(p.src) < int(offset) {
		return errInvalidCFFTable
	}
	p.offset = int(offset)
	return nil
}

func bigEndian(b []byte) uint32 {
	switch len(b) {
	case 1:
		return uint32(b[0])
	case 2:
		return uint32(b[0])<<8 | uint32(b[1])
	case 3:
		return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
	case 4:
		return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	}
	panic("unreachable")
}

func (p *cffParser) parseIndexHeader() (count uint16, offSize int32, err error) {
	buf, err := p.read(2)
	if err != nil {
		return 0, 0, err
	}
	count = be.Uint16(buf)
	// 5176.CFF.pdf section 5 "INDEX Data" says that "An empty INDEX is
	// represented by a count field with a 0 value and no additional fields.
	// Thus, the total size of an empty INDEX is 2 bytes".
	if count == 0 {
		return count, 0, nil
	}
	buf, err = p.read(1)
	if err != nil {
		return 0, 0, err
	}
	offSize = int32(buf[0])
	if offSize < 1 || 4 < offSize {
		return 0, 0, errInvalidCFFTable
	}
	return count, offSize, nil
}

func (p *cffParser) parseIndexLocations(dst []uint32, offSize int32) error {
	if len(dst) == 0 {
		return nil
	}
	buf, err := p.read(len(dst) * int(offSize))
	if err != nil {
		return err
	}

	prev := uint32(0)
	for i := range dst {
		loc := bigEndian(buf[:offSize])
		buf = buf[offSize:]

		// Locations are off by 1 byte. 5176.CFF.pdf section 5 "INDEX Data"
		// says that "Offsets in the offset array are relative to the byte that
		// precedes the object data... This ensures that every object has a
		// corresponding offset which is always nonzero".
		if loc == 0 {
			return errInvalidCFFTable
		}
		loc--

		// In the same paragraph, "Therefore the first element of the offset
		// array is always 1" before correcting for the off-by-1.
		if i == 0 {
			if loc != 0 {
				return errInvalidCFFTable
			}
		} else if loc <= prev { // Check that locations are increasing.
			return errInvalidCFFTable
		}

		// Check that locations are in bounds.
		if uint32(len(p.src)-p.offset) < loc {
			return errInvalidCFFTable
		}

		dst[i] = uint32(p.offset) + loc
		prev = loc
	}
	return nil
}

// topDictData contains fields specific to the Top DICT context.
type topDictData struct {
	// SIDs, to be decoded using the string index
	version, notice, fullName, familyName, weight      uint16
	isFixedPitch                                       bool
	italicAngle, underlinePosition, underlineThickness float32
	charsetOffset                                      int32
	encodingOffset                                     int32
	charStringsOffset                                  int32
	fdArray                                            int32
	fdSelect                                           int32
	isCIDFont                                          bool
	cidFontName                                        uint16
	privateDictOffset                                  int32
	privateDictLength                                  int32
}

// resolve the strings
func (topDict topDictData) toInfo(strs userStrings) (out fonts.PSInfo, err error) {
	out.Version, err = strs.getString(topDict.version)
	if err != nil {
		return out, err
	}
	out.Notice, err = strs.getString(topDict.notice)
	if err != nil {
		return out, err
	}
	out.FullName, err = strs.getString(topDict.fullName)
	if err != nil {
		return out, err
	}
	out.FamilyName, err = strs.getString(topDict.familyName)
	if err != nil {
		return out, err
	}
	out.Weight, err = strs.getString(topDict.weight)
	if err != nil {
		return out, err
	}
	out.IsFixedPitch = topDict.isFixedPitch
	out.ItalicAngle = int(topDict.italicAngle)
	out.UnderlinePosition = int(topDict.underlinePosition)
	out.UnderlineThickness = int(topDict.underlineThickness)
	return out, nil
}

func (topDict *topDictData) Context() ps.PsContext { return ps.TopDict }

func (topDict *topDictData) Run(op ps.PsOperator, state *ps.Inter) error {
	ops := topDictOperators[0]
	if op.IsEscaped {
		ops = topDictOperators[1]
	}
	if int(op.Operator) >= len(ops) {
		return fmt.Errorf("invalid operator %s in Top Dict", op)
	}
	opFunc := ops[op.Operator]
	if opFunc.run == nil {
		return fmt.Errorf("invalid operator %s in Top Dict", op)
	}
	if state.ArgStack.Top < opFunc.numPop {
		return fmt.Errorf("invalid number of arguments for operator %s in Top Dict", op)
	}
	err := opFunc.run(topDict, state)
	if err != nil {
		return err
	}
	err = state.ArgStack.PopN(opFunc.numPop)
	return err
}

// The Top DICT operators are defined by 5176.CFF.pdf Table 9 "Top DICT
// Operator Entries" and Table 10 "CIDFont Operator Extensions".
type topDictOperator struct {
	// run is the function that implements the operator. Nil means that we
	// ignore the operator, other than popping its arguments off the stack.
	run func(*topDictData, *ps.Inter) error

	// numPop is the number of stack values to pop. -1 means "array" and -2
	// means "delta" as per 5176.CFF.pdf Table 6 "Operand Types".
	numPop int32
}

func topDictNoOp(*topDictData, *ps.Inter) error { return nil }

var topDictOperators = [2][]topDictOperator{
	// 1-byte operators.
	{
		0: {func(t *topDictData, s *ps.Inter) error {
			t.version = s.ArgStack.Uint16()
			return nil
		}, +1 /*version*/},
		1: {func(t *topDictData, s *ps.Inter) error {
			t.notice = s.ArgStack.Uint16()
			return nil
		}, +1 /*Notice*/},
		2: {func(t *topDictData, s *ps.Inter) error {
			t.fullName = s.ArgStack.Uint16()
			return nil
		}, +1 /*FullName*/},
		3: {func(t *topDictData, s *ps.Inter) error {
			t.familyName = s.ArgStack.Uint16()
			return nil
		}, +1 /*FamilyName*/},
		4: {func(t *topDictData, s *ps.Inter) error {
			t.weight = s.ArgStack.Uint16()
			return nil
		}, +1 /*Weight*/},
		5:  {topDictNoOp, -1 /*FontBBox*/},
		13: {topDictNoOp, +1 /*UniqueID*/},
		14: {topDictNoOp, -1 /*XUID*/},
		15: {func(t *topDictData, s *ps.Inter) error {
			t.charsetOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}, +1 /*charset*/},
		16: {func(t *topDictData, s *ps.Inter) error {
			t.encodingOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}, +1 /*Encoding*/},
		17: {func(t *topDictData, s *ps.Inter) error {
			t.charStringsOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}, +1 /*CharStrings*/},
		18: {func(t *topDictData, s *ps.Inter) error {
			t.privateDictLength = s.ArgStack.Vals[s.ArgStack.Top-2]
			t.privateDictOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}, +2 /*Private*/},
	},
	// 2-byte operators. The first byte is the escape byte.
	{
		0: {topDictNoOp, +1 /*Copyright*/},
		1: {func(t *topDictData, s *ps.Inter) error {
			t.isFixedPitch = s.ArgStack.Vals[s.ArgStack.Top-1] == 1
			return nil
		}, +1 /*isFixedPitch*/},
		2: {func(t *topDictData, s *ps.Inter) error {
			t.italicAngle = s.ArgStack.Float()
			return nil
		}, +1 /*ItalicAngle*/},
		3: {func(t *topDictData, s *ps.Inter) error {
			t.underlinePosition = s.ArgStack.Float()
			return nil
		}, +1 /*UnderlinePosition*/},
		4: {func(t *topDictData, s *ps.Inter) error {
			t.underlineThickness = s.ArgStack.Float()
			return nil
		}, +1 /*UnderlineThickness*/},
		5: {topDictNoOp, +1 /*PaintType*/},
		6: {func(_ *topDictData, i *ps.Inter) error {
			if version := i.ArgStack.Vals[i.ArgStack.Top-1]; version != 2 {
				return fmt.Errorf("charstring type %d not supported", version)
			}
			return nil
		}, +1 /*CharstringType*/},
		7:  {topDictNoOp, -1 /*FontMatrix*/},
		8:  {topDictNoOp, +1 /*StrokeWidth*/},
		20: {topDictNoOp, +1 /*SyntheticBase*/},
		21: {topDictNoOp, +1 /*PostScript*/},
		22: {topDictNoOp, +1 /*BaseFontName*/},
		23: {topDictNoOp, -2 /*BaseFontBlend*/},
		30: {func(t *topDictData, _ *ps.Inter) error {
			t.isCIDFont = true
			return nil
		}, +3 /*ROS*/},
		31: {topDictNoOp, +1 /*CIDFontVersion*/},
		32: {topDictNoOp, +1 /*CIDFontRevision*/},
		33: {topDictNoOp, +1 /*CIDFontType*/},
		34: {topDictNoOp, +1 /*CIDCount*/},
		35: {topDictNoOp, +1 /*UIDBase*/},
		36: {func(t *topDictData, s *ps.Inter) error {
			t.fdArray = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}, +1 /*FDArray*/},
		37: {func(t *topDictData, s *ps.Inter) error {
			t.fdSelect = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}, +1 /*FDSelect*/},
		38: {func(t *topDictData, s *ps.Inter) error {
			t.cidFontName = s.ArgStack.Uint16()
			return nil
		}, +1 /*FontName*/},
	},
}

// privateDict contains fields specific to the Private DICT context.
type privateDict struct {
	subrsOffset                  int32
	defaultWidthX, nominalWidthX int32
}

func (privateDict) Context() ps.PsContext { return ps.PrivateDict }

// The Private DICT operators are defined by 5176.CFF.pdf Table 23 "Private
// DICT Operators".
func (priv *privateDict) Run(op ps.PsOperator, state *ps.Inter) error {
	if !op.IsEscaped { // 1-byte operators.
		switch op.Operator {
		case 6, 7, 8, 9: // "BlueValues" "OtherBlues" "FamilyBlues" "FamilyOtherBlues"
			return state.ArgStack.PopN(-2)
		case 10, 11: // "StdHW" "StdVW"
			return state.ArgStack.PopN(1)
		case 20: // "defaultWidthX"
			if state.ArgStack.Top < 1 {
				return errors.New("invalid stack size for 'defaultWidthX' in private Dict charstring")
			}
			priv.defaultWidthX = state.ArgStack.Vals[state.ArgStack.Top-1]
			return state.ArgStack.PopN(1)
		case 21: // "nominalWidthX"
			if state.ArgStack.Top < 1 {
				return errors.New("invalid stack size for 'nominalWidthX' in private Dict charstring")
			}
			priv.nominalWidthX = state.ArgStack.Vals[state.ArgStack.Top-1]
			return state.ArgStack.PopN(1)
		case 19: // "Subrs" pop 1
			if state.ArgStack.Top < 1 {
				return errors.New("invalid stack size for 'subrs' in private Dict charstring")
			}
			priv.subrsOffset = state.ArgStack.Vals[state.ArgStack.Top-1]
			return state.ArgStack.PopN(1)
		}
	} else { // 2-byte operators. The first byte is the escape byte.
		switch op.Operator {
		case 9, 10, 11, 14, 17, 18, 19: // "BlueScale" "BlueShift" "BlueFuzz" "ForceBold" "LanguageGroup" "ExpansionFactor" "initialRandomSeed"
			return state.ArgStack.PopN(1)
		case 12, 13: //  "StemSnapH"  "StemSnapV"
			return state.ArgStack.PopN(-2)
		}
	}
	return errors.New("invalid operand in private Dict charstring")
}
