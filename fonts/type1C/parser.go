package type1C

// code is adapted from golang.org/x/image/font/sfnt

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
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
)

type userStrings [][]byte

// return either the predefined string or the user defined one
func (u userStrings) getString(sid uint16) (string, error) {
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

// read 4 bytes to check its a supported CFF file
func checkHeader(r io.Reader) error {
	var buf [4]byte
	r.Read(buf[:])
	if buf[0] != 1 || buf[1] != 0 || buf[2] != 4 {
		return errUnsupportedCFFVersion
	}
	return nil
}

func (p *cffParser) parse() (*CFF, error) {
	// header was checked prior to this call

	var (
		out CFF
		err error
	)

	// Parse the Name INDEX.
	out.fontNames, err = p.parseNames()
	if err != nil {
		return nil, err
	}
	// TODO: make this limitation optional
	// 9.9 - Embedded Font Programs
	// Although CFF enables multiple font or CIDFont programs to be bundled together in a
	// single file, an embedded CFF font file in PDF shall consist of exactly one font or CIDFont (
	if len(out.fontNames) != 1 {
		return nil, errInvalidCFFTable
	}

	topDict, err := p.parseTopDict()
	if err != nil {
		return nil, err
	}

	// parse the String INDEX.
	strs, err := p.parseUserStrings()
	if err != nil {
		return nil, err
	}

	out.PSInfo, err = topDict.toInfo(strs)
	if err != nil {
		return nil, err
	}
	out.cidFontName, err = strs.getString(topDict.cidFontName)
	if err != nil {
		return nil, err
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
		if err := p.parseIndexLocations(gsubrs, offSize); err != nil {
			return nil, err
		}
	}

	// Parse the CharStrings INDEX, whose location was found in the Top DICT.
	if err := p.seek(topDict.charStringsOffset); err != nil {
		return nil, err
	}
	out.charstrings, err = p.parseIndex()
	if err != nil {
		return nil, err
	}
	numGlyphs := uint16(len(out.charstrings))

	charset, err := p.parseCharset(topDict, int(numGlyphs))
	if err != nil {
		return nil, err
	}

	out.Encoding, err = p.parseEncoding(topDict, numGlyphs, charset, strs)
	if err != nil {
		return nil, err
	}

	if !topDict.isCIDFont {
		// Parse the Private DICT, whose location was found in the Top DICT.
		_, err := p.parsePrivateDICT(topDict.privateDictOffset, topDict.privateDictLength)
		if err != nil {
			return nil, err
		}
	} else {
		// Parse the Font Dict Select data, whose location was found in the Top
		// DICT.
		out.fdSelect, err = p.parseFDSelect(topDict.fdSelect, numGlyphs)
		if err != nil {
			return nil, err
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
	return &out, nil
}

func (p *cffParser) parseTopDict() (*topDictData, error) {
	// Parse the Top DICT INDEX.
	count, offSize, err := p.parseIndexHeader()
	if err != nil {
		return nil, err
	}
	// 5176.CFF.pdf section 8 "Top DICT INDEX" says that the count here
	// should match the count of the Name INDEX, which is 1.
	if count != 1 {
		return nil, errInvalidCFFTable
	}
	var locBuf [2]uint32
	if err := p.parseIndexLocations(locBuf[:], offSize); err != nil {
		return nil, err
	}
	buf, err := p.read(int(locBuf[1] - locBuf[0]))
	if err != nil {
		return nil, err
	}

	var (
		psi     ps.Inter
		topDict topDictData
	)
	// set default value before parsing
	topDict.underlinePosition = -100
	topDict.underlineThickness = 50

	if err = psi.Run(buf, nil, &topDict); err != nil {
		return nil, err
	}
	return &topDict, nil
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
func (p *cffParser) parseCharset(topDict *topDictData, numGlyphs int) ([]uint16, error) {
	// Predefined charset may have offset of 0 to 2 // Table 22
	var charset []uint16
	switch charsetOffset := topDict.charsetOffset; charsetOffset {
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
			buf, err := p.read(2 * (numGlyphs - 1)) // ".notdef" is omited, and has an implicit SID of 0
			if err != nil {
				return nil, err
			}
			for i := 1; i < numGlyphs; i++ {
				charset[i] = be.Uint16(buf[2*i-2:])
			}
		case 1:
			for i := 1; i < numGlyphs; {
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
			for i := 1; i < numGlyphs; {
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
func (p *cffParser) parseEncoding(topDict *topDictData, numGlyphs uint16, charset []uint16, strs userStrings) (*simpleencodings.Encoding, error) {
	// Predefined encoding may have offset of 0 to 1 // Table 16
	switch encodingOffset := topDict.encodingOffset; encodingOffset {
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
				name, err := strs.getString(charset[i])
				if err != nil {
					return nil, err
				}
				encoding[c] = name
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
		return 0, errors.New("invalig glyph index")
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

// Parse the Local Subrs [Subroutines] INDEX, whose location was found in
// the Private DICT.
func (p *cffParser) parsePrivateDICT(offset, length int32) (subrs []uint32, err error) {
	if length == 0 {
		return
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
	if err := p.seek(offset + priv.subrsOffset); err != nil {
		return nil, errInvalidCFFTable
	}
	count, offSize, err := p.parseIndexHeader()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	if count > maxNumSubroutines {
		return nil, errUnsupportedNumberOfSubroutines
	}
	subrs = make([]uint32, count+1)
	if err = p.parseIndexLocations(subrs, offSize); err != nil {
		return nil, err
	}
	return subrs, err
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

// resolve the strings: for SID = 0, we rather use
// empty strings than .notdef
func (t topDictData) toInfo(strs userStrings) (out fonts.PSInfo, err error) {
	getS := func(sid uint16) (string, error) {
		if sid == 0 {
			return "", nil
		}
		return strs.getString(sid)
	}
	out.Version, err = getS(t.version)
	if err != nil {
		return out, err
	}
	out.Notice, err = getS(t.notice)
	if err != nil {
		return out, err
	}
	out.FullName, err = getS(t.fullName)
	if err != nil {
		return out, err
	}
	out.FamilyName, err = getS(t.familyName)
	if err != nil {
		return out, err
	}
	out.Weight, err = getS(t.weight)
	if err != nil {
		return out, err
	}
	out.IsFixedPitch = t.isFixedPitch
	out.ItalicAngle = int(t.italicAngle)
	out.UnderlinePosition = int(t.underlinePosition)
	out.UnderlineThickness = int(t.underlineThickness)
	return out, nil
}

func (topDict *topDictData) Context() ps.PsContext { return ps.TopDict }

func (topDict *topDictData) Run(op ps.PsOperator, state *ps.Inter) (int32, error) {
	ops := topDictOperators[0]
	if op.IsEscaped {
		ops = topDictOperators[1]
	}
	if int(op.Operator) >= len(ops) {
		return 0, fmt.Errorf("invalid operator %s in Top Dict", op)
	}
	opFunc := ops[op.Operator]
	if opFunc.run == nil {
		return 0, fmt.Errorf("invalid operator %s in Top Dict", op)
	}
	if state.ArgStack.Top < opFunc.numPop {
		return 0, fmt.Errorf("invalid number of arguments for operator %s in Top Dict", op)
	}
	err := opFunc.run(topDict, state)
	return opFunc.numPop, err
}

// The Top DICT operators are defined by 5176.CFF.pdf Table 9 "Top DICT
// Operator Entries" and Table 10 "CIDFont Operator Extensions".
type topDictOperator struct {
	// numPop is the number of stack values to pop. -1 means "array" and -2
	// means "delta" as per 5176.CFF.pdf Table 6 "Operand Types".
	numPop int32

	// run is the function that implements the operator. Nil means that we
	// ignore the operator, other than popping its arguments off the stack.
	run func(*topDictData, *ps.Inter) error
}

func topDictNoOp(*topDictData, *ps.Inter) error { return nil }

var topDictOperators = [2][]topDictOperator{
	// 1-byte operators.
	{
		0: {+1 /*version*/, func(t *topDictData, s *ps.Inter) error {
			t.version = s.ArgStack.Uint16()
			return nil
		}},
		1: {+1 /*Notice*/, func(t *topDictData, s *ps.Inter) error {
			t.notice = s.ArgStack.Uint16()
			return nil
		}},
		2: {+1 /*FullName*/, func(t *topDictData, s *ps.Inter) error {
			t.fullName = s.ArgStack.Uint16()
			return nil
		}},
		3: {+1 /*FamilyName*/, func(t *topDictData, s *ps.Inter) error {
			t.familyName = s.ArgStack.Uint16()
			return nil
		}},
		4: {+1 /*Weight*/, func(t *topDictData, s *ps.Inter) error {
			t.weight = s.ArgStack.Uint16()
			return nil
		}},
		5:  {-1 /*FontBBox*/, topDictNoOp},
		13: {+1 /*UniqueID*/, topDictNoOp},
		14: {-1 /*XUID*/, topDictNoOp},
		15: {+1 /*charset*/, func(t *topDictData, s *ps.Inter) error {
			t.charsetOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}},
		16: {+1 /*Encoding*/, func(t *topDictData, s *ps.Inter) error {
			t.encodingOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}},
		17: {+1 /*CharStrings*/, func(t *topDictData, s *ps.Inter) error {
			t.charStringsOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}},
		18: {+2 /*Private*/, func(t *topDictData, s *ps.Inter) error {
			t.privateDictLength = s.ArgStack.Vals[s.ArgStack.Top-2]
			t.privateDictOffset = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}},
	},
	// 2-byte operators. The first byte is the escape byte.
	{
		0: {+1 /*Copyright*/, topDictNoOp},
		1: {+1 /*isFixedPitch*/, func(t *topDictData, s *ps.Inter) error {
			t.isFixedPitch = s.ArgStack.Vals[s.ArgStack.Top-1] == 1
			return nil
		}},
		2: {+1 /*ItalicAngle*/, func(t *topDictData, s *ps.Inter) error {
			t.italicAngle = s.ArgStack.Float()
			return nil
		}},
		3: {+1 /*UnderlinePosition*/, func(t *topDictData, s *ps.Inter) error {
			t.underlinePosition = s.ArgStack.Float()
			return nil
		}},
		4: {+1 /*UnderlineThickness*/, func(t *topDictData, s *ps.Inter) error {
			t.underlineThickness = s.ArgStack.Float()
			return nil
		}},
		5: {+1 /*PaintType*/, topDictNoOp},
		6: {+1 /*CharstringType*/, func(tdd *topDictData, i *ps.Inter) error {
			if version := i.ArgStack.Vals[i.ArgStack.Top-1]; version != 2 {
				return fmt.Errorf("charstring type %d not supported", version)
			}
			return nil
		}},
		7:  {-1 /*FontMatrix*/, topDictNoOp},
		8:  {+1 /*StrokeWidth*/, topDictNoOp},
		20: {+1 /*SyntheticBase*/, topDictNoOp},
		21: {+1 /*PostScript*/, topDictNoOp},
		22: {+1 /*BaseFontName*/, topDictNoOp},
		23: {-2 /*BaseFontBlend*/, topDictNoOp},
		30: {+3 /*ROS*/, func(t *topDictData, s *ps.Inter) error {
			t.isCIDFont = true
			return nil
		}},
		31: {+1 /*CIDFontVersion*/, topDictNoOp},
		32: {+1 /*CIDFontRevision*/, topDictNoOp},
		33: {+1 /*CIDFontType*/, topDictNoOp},
		34: {+1 /*CIDCount*/, topDictNoOp},
		35: {+1 /*UIDBase*/, topDictNoOp},
		36: {+1 /*FDArray*/, func(t *topDictData, s *ps.Inter) error {
			t.fdArray = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}},
		37: {+1 /*FDSelect*/, func(t *topDictData, s *ps.Inter) error {
			t.fdSelect = s.ArgStack.Vals[s.ArgStack.Top-1]
			return nil
		}},
		38: {+1 /*FontName*/, func(t *topDictData, s *ps.Inter) error {
			t.cidFontName = s.ArgStack.Uint16()
			return nil
		}},
	},
}

// privateDict contains fields specific to the Private DICT context.
type privateDict struct {
	subrsOffset                  int32
	defaultWidthX, nominalWidthX int32
}

func (privateDict) Context() psinterpreter.PsContext { return ps.PrivateDict }

// The Private DICT operators are defined by 5176.CFF.pdf Table 23 "Private
// DICT Operators".
func (priv *privateDict) Run(op ps.PsOperator, state *ps.Inter) (int32, error) {
	if !op.IsEscaped { // 1-byte operators.
		switch op.Operator {
		case 6, 7, 8, 9: // "BlueValues" "OtherBlues" "FamilyBlues" "FamilyOtherBlues"
			return -2, nil
		case 10, 11: // "StdHW" "StdVW"
			return 1, nil
		case 20: // "defaultWidthX"
			if state.ArgStack.Top < 1 {
				return 0, errors.New("invalid stack size for 'defaultWidthX' in private Dict charstring")
			}
			priv.defaultWidthX = state.ArgStack.Vals[state.ArgStack.Top-1]
			return 1, nil
		case 21: // "nominalWidthX"
			if state.ArgStack.Top < 1 {
				return 0, errors.New("invalid stack size for 'nominalWidthX' in private Dict charstring")
			}
			priv.nominalWidthX = state.ArgStack.Vals[state.ArgStack.Top-1]
			return 1, nil
		case 19: // "Subrs" pop 1
			if state.ArgStack.Top < 1 {
				return 0, errors.New("invalid stack size for 'subrs' in private Dict charstring")
			}
			priv.subrsOffset = state.ArgStack.Vals[state.ArgStack.Top-1]
			return 1, nil
		}
	} else { // 2-byte operators. The first byte is the escape byte.
		switch op.Operator {
		case 9, 10, 11, 14, 17, 18, 19: // "BlueScale" "BlueShift" "BlueFuzz" "ForceBold" "LanguageGroup" "ExpansionFactor" "initialRandomSeed"
			return 1, nil
		case 12, 13: //  "StemSnapH"  "StemSnapV"
			return -2, nil
		}
	}
	return 0, errors.New("invalid operand in private Dict charstring")
}

// type2Metrics implements operators needed to fetch Type2 charstring metrics
type type2Metrics struct {
	// found in private DICT, needed since we can't differenciate
	// no width set from 0 width
	// `width` must be initialized to default width
	nominalWidthX int32

	width int32
}

func (type2Metrics) Context() psinterpreter.PsContext { return ps.Type2Charstring }

// we only read the first operators
func (met *type2Metrics) Run(op ps.PsOperator, state *ps.Inter) (int32, error) {
	if !op.IsEscaped {
		switch op.Operator {
		case 21: // rmoveto
			if state.ArgStack.Top > 2 { // width is optional
				met.width = met.nominalWidthX + state.ArgStack.Vals[0]
			}
			return 0, psinterpreter.ErrInterrupt
		case 4, 22: // vmoveto, hmoveto
			if state.ArgStack.Top > 1 { // width is optional
				met.width = met.nominalWidthX + state.ArgStack.Vals[0]
			}
			return 0, psinterpreter.ErrInterrupt
		case 1, 3, 18, 23, 19, 20: // hstem, vstem, hstemhm, vstemhm, hintmask, cntrmask
			// variable number of arguments, but always even
			// for xxxmask, if there are arguments on the stack, then this is an impliied stem
			if state.ArgStack.Top&1 != 0 {
				met.width = met.nominalWidthX + state.ArgStack.Vals[0]
			}
			return 0, psinterpreter.ErrInterrupt
		case 14: // endchar
			if state.ArgStack.Top > 0 { // width is optional
				met.width = met.nominalWidthX + state.ArgStack.Vals[0]
			}
			return 0, psinterpreter.ErrInterrupt
		}
	}
	// no other operands are allowed before the ones handled above
	return 0, errors.New("invalid operand in private Dict charstring")
}

// // psType2CharstringsData contains fields specific to the Type 2 Charstrings
// // context.
// type psType2CharstringsData struct {
// 	f          *Font
// 	b          *Buffer
// 	x          int32
// 	y          int32
// 	firstX     int32
// 	firstY     int32
// 	hintBits   int32
// 	seenWidth  bool
// 	ended      bool
// 	glyphIndex GlyphIndex
// 	// fdSelectIndexPlusOne is the result of the Font Dict Select lookup, plus
// 	// one. That plus one lets us use the zero value to denote either unused
// 	// (for CFF fonts with a single Font Dict) or lazily evaluated.
// 	fdSelectIndexPlusOne int32
// }

// func (d *psType2CharstringsData) initialize(f *Font, b *Buffer, glyphIndex GlyphIndex) {
// 	*d = psType2CharstringsData{
// 		f:          f,
// 		b:          b,
// 		glyphIndex: glyphIndex,
// 	}
// }

// func (d *psType2CharstringsData) closePath() {
// 	if d.x != d.firstX || d.y != d.firstY {
// 		d.b.segments = append(d.b.segments, Segment{
// 			Op: SegmentOpLineTo,
// 			Args: [3]fixed.Point26_6{{
// 				X: fixed.Int26_6(d.firstX),
// 				Y: fixed.Int26_6(d.firstY),
// 			}},
// 		})
// 	}
// }

// func (d *psType2CharstringsData) moveTo(dx, dy int32) {
// 	d.closePath()
// 	d.x += dx
// 	d.y += dy
// 	d.b.segments = append(d.b.segments, Segment{
// 		Op: SegmentOpMoveTo,
// 		Args: [3]fixed.Point26_6{{
// 			X: fixed.Int26_6(d.x),
// 			Y: fixed.Int26_6(d.y),
// 		}},
// 	})
// 	d.firstX = d.x
// 	d.firstY = d.y
// }

// func (d *psType2CharstringsData) lineTo(dx, dy int32) {
// 	d.x += dx
// 	d.y += dy
// 	d.b.segments = append(d.b.segments, Segment{
// 		Op: SegmentOpLineTo,
// 		Args: [3]fixed.Point26_6{{
// 			X: fixed.Int26_6(d.x),
// 			Y: fixed.Int26_6(d.y),
// 		}},
// 	})
// }

// func (d *psType2CharstringsData) cubeTo(dxa, dya, dxb, dyb, dxc, dyc int32) {
// 	d.x += dxa
// 	d.y += dya
// 	xa := fixed.Int26_6(d.x)
// 	ya := fixed.Int26_6(d.y)
// 	d.x += dxb
// 	d.y += dyb
// 	xb := fixed.Int26_6(d.x)
// 	yb := fixed.Int26_6(d.y)
// 	d.x += dxc
// 	d.y += dyc
// 	xc := fixed.Int26_6(d.x)
// 	yc := fixed.Int26_6(d.y)
// 	d.b.segments = append(d.b.segments, Segment{
// 		Op:   SegmentOpCubeTo,
// 		Args: [3]fixed.Point26_6{{X: xa, Y: ya}, {X: xb, Y: yb}, {X: xc, Y: yc}},
// 	})
// }

type psInterpreter struct{}

type psOperator struct {
	// numPop is the number of stack values to pop. -1 means "array" and -2
	// means "delta" as per 5176.CFF.pdf Table 6 "Operand Types".
	numPop int32
	// name is the operator name. An empty name (i.e. the zero value for the
	// struct overall) means an unrecognized 1-byte operator.
	name string
	// run is the function that implements the operator. Nil means that we
	// ignore the operator, other than popping its arguments off the stack.
	run func(*psInterpreter) error
}

// psOperators holds the 1-byte and 2-byte operators for PostScript interpreter
// contexts.
var psOperators = [...][2][]psOperator{
	// // The Type 2 Charstring operators are defined by 5177.Type2.pdf Appendix A
	// // "Type 2 Charstring Command Codes".
	// psContextType2Charstring: {{
	// 	// 1-byte operators.
	// 	0:  {}, // Reserved.
	// 	2:  {}, // Reserved.
	// 	1:  {-1, "hstem", t2CStem},
	// 	3:  {-1, "vstem", t2CStem},
	// 	18: {-1, "hstemhm", t2CStem},
	// 	23: {-1, "vstemhm", t2CStem},
	// 	5:  {-1, "rlineto", t2CRlineto},
	// 	6:  {-1, "hlineto", t2CHlineto},
	// 	7:  {-1, "vlineto", t2CVlineto},
	// 	8:  {-1, "rrcurveto", t2CRrcurveto},
	// 	9:  {}, // Reserved.
	// 	10: {+1, "callsubr", t2CCallsubr},
	// 	11: {+0, "return", t2CReturn},
	// 	12: {}, // escape.
	// 	13: {}, // Reserved.
	// 	14: {-1, "endchar", t2CEndchar},
	// 	15: {}, // Reserved.
	// 	16: {}, // Reserved.
	// 	17: {}, // Reserved.
	// 	19: {-1, "hintmask", t2CMask},
	// 	20: {-1, "cntrmask", t2CMask},
	// 	4:  {-1, "vmoveto", t2CVmoveto},
	// 	21: {-1, "rmoveto", t2CRmoveto},
	// 	22: {-1, "hmoveto", t2CHmoveto},
	// 	24: {-1, "rcurveline", t2CRcurveline},
	// 	25: {-1, "rlinecurve", t2CRlinecurve},
	// 	26: {-1, "vvcurveto", t2CVvcurveto},
	// 	27: {-1, "hhcurveto", t2CHhcurveto},
	// 	28: {}, // shortint.
	// 	29: {+1, "callgsubr", t2CCallgsubr},
	// 	30: {-1, "vhcurveto", t2CVhcurveto},
	// 	31: {-1, "hvcurveto", t2CHvcurveto},
	// }, {
	// 	// 2-byte operators. The first byte is the escape byte.
	// 	34: {+7, "hflex", t2CHflex},
	// 	36: {+9, "hflex1", t2CHflex1},
	// 	// TODO: more operators.
	// }},
}

// // t2CReadWidth reads the optional width adjustment. If present, it is on the
// // bottom of the arg stack. nArgs is the expected number of arguments on the
// // stack. A negative nArgs means a multiple of 2.
// //
// // 5177.Type2.pdf page 16 Note 4 says: "The first stack-clearing operator,
// // which must be one of hstem, hstemhm, vstem, vstemhm, cntrmask, hintmask,
// // hmoveto, vmoveto, rmoveto, or endchar, takes an additional argument — the
// // width... which may be expressed as zero or one numeric argument."
// func t2CReadWidth(p *psInterpreter, nArgs int32) {
// 	if p.type2Charstrings.seenWidth {
// 		return
// 	}
// 	p.type2Charstrings.seenWidth = true
// 	if nArgs >= 0 {
// 		if p.argStack.top != nArgs+1 {
// 			return
// 		}
// 	} else if p.argStack.top&1 == 0 {
// 		return
// 	}
// 	// When parsing a standalone CFF, we'd save the value of p.argStack.a[0]
// 	// here as it defines the glyph's width (horizontal advance). Specifically,
// 	// if present, it is a delta to the font-global nominalWidthX value found
// 	// in the Private DICT. If absent, the glyph's width is the defaultWidthX
// 	// value in that dict. See 5176.CFF.pdf section 15 "Private DICT Data".
// 	//
// 	// For a CFF embedded in an SFNT font (i.e. an OpenType font), glyph widths
// 	// are already stored in the hmtx table, separate to the CFF table, and it
// 	// is simpler to parse that table for all OpenType fonts (PostScript and
// 	// TrueType). We therefore ignore the width value here, and just remove it
// 	// from the bottom of the argStack.
// 	copy(p.argStack.a[:p.argStack.top-1], p.argStack.a[1:p.argStack.top])
// 	p.argStack.top--
// }

// func t2CStem(p *psInterpreter) error {
// 	t2CReadWidth(p, -1)
// 	if p.argStack.top%2 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	// We update the number of hintBits need to parse hintmask and cntrmask
// 	// instructions, but this Type 2 Charstring implementation otherwise
// 	// ignores the stem hints.
// 	p.type2Charstrings.hintBits += p.argStack.top / 2
// 	if p.type2Charstrings.hintBits > maxHintBits {
// 		return errUnsupportedNumberOfHints
// 	}
// 	return nil
// }

// func t2CMask(p *psInterpreter) error {
// 	// 5176.CFF.pdf section 4.3 "Hint Operators" says that "If hstem and vstem
// 	// hints are both declared at the beginning of a charstring, and this
// 	// sequence is followed directly by the hintmask or cntrmask operators, the
// 	// vstem hint operator need not be included."
// 	//
// 	// What we implement here is more permissive (but the same as what the
// 	// FreeType implementation does, and simpler than tracking the previous
// 	// operator and other hinting state): if a hintmask is given any arguments
// 	// (i.e. the argStack is non-empty), we run an implicit vstem operator.
// 	//
// 	// Note that the vstem operator consumes from p.argStack, but the hintmask
// 	// or cntrmask operators consume from p.instructions.
// 	if p.argStack.top != 0 {
// 		if err := t2CStem(p); err != nil {
// 			return err
// 		}
// 	} else if !p.type2Charstrings.seenWidth {
// 		p.type2Charstrings.seenWidth = true
// 	}

// 	hintBytes := (p.type2Charstrings.hintBits + 7) / 8
// 	if len(p.instructions) < int(hintBytes) {
// 		return errInvalidCFFTable
// 	}
// 	p.instructions = p.instructions[hintBytes:]
// 	return nil
// }

// func t2CHmoveto(p *psInterpreter) error {
// 	t2CReadWidth(p, 1)
// 	if p.argStack.top != 1 {
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.moveTo(p.argStack.a[0], 0)
// 	return nil
// }

// func t2CVmoveto(p *psInterpreter) error {
// 	t2CReadWidth(p, 1)
// 	if p.argStack.top != 1 {
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.moveTo(0, p.argStack.a[0])
// 	return nil
// }

// func t2CRmoveto(p *psInterpreter) error {
// 	t2CReadWidth(p, 2)
// 	if p.argStack.top != 2 {
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.moveTo(p.argStack.a[0], p.argStack.a[1])
// 	return nil
// }

// func t2CHlineto(p *psInterpreter) error { return t2CLineto(p, false) }
// func t2CVlineto(p *psInterpreter) error { return t2CLineto(p, true) }

// func t2CLineto(p *psInterpreter, vertical bool) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 1 {
// 		return errInvalidCFFTable
// 	}
// 	for i := int32(0); i < p.argStack.top; i, vertical = i+1, !vertical {
// 		dx, dy := p.argStack.a[i], int32(0)
// 		if vertical {
// 			dx, dy = dy, dx
// 		}
// 		p.type2Charstrings.lineTo(dx, dy)
// 	}
// 	return nil
// }

// func t2CRlineto(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 2 || p.argStack.top%2 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	for i := int32(0); i < p.argStack.top; i += 2 {
// 		p.type2Charstrings.lineTo(p.argStack.a[i], p.argStack.a[i+1])
// 	}
// 	return nil
// }

// // As per 5177.Type2.pdf section 4.1 "Path Construction Operators",
// //
// // rcurveline is:
// //	- {dxa dya dxb dyb dxc dyc}+ dxd dyd
// //
// // rlinecurve is:
// //	- {dxa dya}+ dxb dyb dxc dyc dxd dyd

// func t2CRcurveline(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 8 || p.argStack.top%6 != 2 {
// 		return errInvalidCFFTable
// 	}
// 	i := int32(0)
// 	for iMax := p.argStack.top - 2; i < iMax; i += 6 {
// 		p.type2Charstrings.cubeTo(
// 			p.argStack.a[i+0],
// 			p.argStack.a[i+1],
// 			p.argStack.a[i+2],
// 			p.argStack.a[i+3],
// 			p.argStack.a[i+4],
// 			p.argStack.a[i+5],
// 		)
// 	}
// 	p.type2Charstrings.lineTo(p.argStack.a[i], p.argStack.a[i+1])
// 	return nil
// }

// func t2CRlinecurve(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 8 || p.argStack.top%2 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	i := int32(0)
// 	for iMax := p.argStack.top - 6; i < iMax; i += 2 {
// 		p.type2Charstrings.lineTo(p.argStack.a[i], p.argStack.a[i+1])
// 	}
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[i+0],
// 		p.argStack.a[i+1],
// 		p.argStack.a[i+2],
// 		p.argStack.a[i+3],
// 		p.argStack.a[i+4],
// 		p.argStack.a[i+5],
// 	)
// 	return nil
// }

// // As per 5177.Type2.pdf section 4.1 "Path Construction Operators",
// //
// // hhcurveto is:
// //	- dy1 {dxa dxb dyb dxc}+
// //
// // vvcurveto is:
// //	- dx1 {dya dxb dyb dyc}+
// //
// // hvcurveto is one of:
// //	- dx1 dx2 dy2 dy3 {dya dxb dyb dxc dxd dxe dye dyf}* dxf?
// //	- {dxa dxb dyb dyc dyd dxe dye dxf}+ dyf?
// //
// // vhcurveto is one of:
// //	- dy1 dx2 dy2 dx3 {dxa dxb dyb dyc dyd dxe dye dxf}* dyf?
// //	- {dya dxb dyb dxc dxd dxe dye dyf}+ dxf?

// func t2CHhcurveto(p *psInterpreter) error { return t2CCurveto(p, false, false) }
// func t2CVvcurveto(p *psInterpreter) error { return t2CCurveto(p, false, true) }
// func t2CHvcurveto(p *psInterpreter) error { return t2CCurveto(p, true, false) }
// func t2CVhcurveto(p *psInterpreter) error { return t2CCurveto(p, true, true) }

// // t2CCurveto implements the hh / vv / hv / vh xxcurveto operators. N relative
// // cubic curve requires 6*N control points, but only 4*N+0 or 4*N+1 are used
// // here: all (or all but one) of the piecewise cubic curve's tangents are
// // implicitly horizontal or vertical.
// //
// // swap is whether that implicit horizontal / vertical constraint swaps as you
// // move along the piecewise cubic curve. If swap is false, the constraints are
// // either all horizontal or all vertical. If swap is true, it alternates.
// //
// // vertical is whether the first implicit constraint is vertical.
// func t2CCurveto(p *psInterpreter, swap, vertical bool) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 4 {
// 		return errInvalidCFFTable
// 	}

// 	i := int32(0)
// 	switch p.argStack.top & 3 {
// 	case 0:
// 		// No-op.
// 	case 1:
// 		if swap {
// 			break
// 		}
// 		i = 1
// 		if vertical {
// 			p.type2Charstrings.x += p.argStack.a[0]
// 		} else {
// 			p.type2Charstrings.y += p.argStack.a[0]
// 		}
// 	default:
// 		return errInvalidCFFTable
// 	}

// 	for i != p.argStack.top {
// 		i = t2CCurveto4(p, swap, vertical, i)
// 		if i < 0 {
// 			return errInvalidCFFTable
// 		}
// 		if swap {
// 			vertical = !vertical
// 		}
// 	}
// 	return nil
// }

// func t2CCurveto4(p *psInterpreter, swap bool, vertical bool, i int32) (j int32) {
// 	if i+4 > p.argStack.top {
// 		return -1
// 	}
// 	dxa := p.argStack.a[i+0]
// 	dya := int32(0)
// 	dxb := p.argStack.a[i+1]
// 	dyb := p.argStack.a[i+2]
// 	dxc := p.argStack.a[i+3]
// 	dyc := int32(0)
// 	i += 4

// 	if vertical {
// 		dxa, dya = dya, dxa
// 	}

// 	if swap {
// 		if i+1 == p.argStack.top {
// 			dyc = p.argStack.a[i]
// 			i++
// 		}
// 	}

// 	if swap != vertical {
// 		dxc, dyc = dyc, dxc
// 	}

// 	p.type2Charstrings.cubeTo(dxa, dya, dxb, dyb, dxc, dyc)
// 	return i
// }

// func t2CRrcurveto(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 6 || p.argStack.top%6 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	for i := int32(0); i != p.argStack.top; i += 6 {
// 		p.type2Charstrings.cubeTo(
// 			p.argStack.a[i+0],
// 			p.argStack.a[i+1],
// 			p.argStack.a[i+2],
// 			p.argStack.a[i+3],
// 			p.argStack.a[i+4],
// 			p.argStack.a[i+5],
// 		)
// 	}
// 	return nil
// }

// // For the flex operators, we ignore the flex depth and always produce cubic
// // segments, not linear segments. It's not obvious why the Type 2 Charstring
// // format cares about switching behavior based on a metric in pixels, not in
// // ideal font units. The Go vector rasterizer has no problems with almost
// // linear cubic segments.

// func t2CHflex(p *psInterpreter) error {
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[0], 0,
// 		p.argStack.a[1], +p.argStack.a[2],
// 		p.argStack.a[3], 0,
// 	)
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[4], 0,
// 		p.argStack.a[5], -p.argStack.a[2],
// 		p.argStack.a[6], 0,
// 	)
// 	return nil
// }

// func t2CHflex1(p *psInterpreter) error {
// 	dy1 := p.argStack.a[1]
// 	dy2 := p.argStack.a[3]
// 	dy5 := p.argStack.a[7]
// 	dy6 := -dy1 - dy2 - dy5
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[0], dy1,
// 		p.argStack.a[2], dy2,
// 		p.argStack.a[4], 0,
// 	)
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[5], 0,
// 		p.argStack.a[6], dy5,
// 		p.argStack.a[8], dy6,
// 	)
// 	return nil
// }

// // subrBias returns the subroutine index bias as per 5177.Type2.pdf section 4.7
// // "Subroutine Operators".
// func subrBias(numSubroutines int) int32 {
// 	if numSubroutines < 1240 {
// 		return 107
// 	}
// 	if numSubroutines < 33900 {
// 		return 1131
// 	}
// 	return 32768
// }

// func t2CCallgsubr(p *psInterpreter) error {
// 	return t2CCall(p, p.type2Charstrings.f.cached.glyphData.gsubrs)
// }

// func t2CCallsubr(p *psInterpreter) error {
// 	t := &p.type2Charstrings
// 	d := &t.f.cached.glyphData
// 	subrs := d.singleSubrs
// 	if d.multiSubrs != nil {
// 		if t.fdSelectIndexPlusOne == 0 {
// 			index, err := d.fdSelect.lookup(t.f, t.b, t.glyphIndex)
// 			if err != nil {
// 				return err
// 			}
// 			if index < 0 || len(d.multiSubrs) <= index {
// 				return errInvalidCFFTable
// 			}
// 			t.fdSelectIndexPlusOne = int32(index + 1)
// 		}
// 		subrs = d.multiSubrs[t.fdSelectIndexPlusOne-1]
// 	}
// 	return t2CCall(p, subrs)
// }

// func t2CCall(p *psInterpreter, subrs []uint32) error {
// 	if p.callStack.top == psCallStackSize || len(subrs) == 0 {
// 		return errInvalidCFFTable
// 	}
// 	length := uint32(len(p.instructions))
// 	p.callStack.a[p.callStack.top] = psCallStackEntry{
// 		offset: p.instrOffset + p.instrLength - length,
// 		length: length,
// 	}
// 	p.callStack.top++

// 	subrIndex := p.argStack.a[p.argStack.top-1] + subrBias(len(subrs)-1)
// 	if subrIndex < 0 || int32(len(subrs)-1) <= subrIndex {
// 		return errInvalidCFFTable
// 	}
// 	i := subrs[subrIndex+0]
// 	j := subrs[subrIndex+1]
// 	if j < i {
// 		return errInvalidCFFTable
// 	}
// 	if j-i > maxGlyphDataLength {
// 		return errUnsupportedGlyphDataLength
// 	}
// 	buf, err := p.type2Charstrings.b.view(&p.type2Charstrings.f.src, int(i), int(j-i))
// 	if err != nil {
// 		return err
// 	}

// 	p.instructions = buf
// 	p.instrOffset = i
// 	p.instrLength = j - i
// 	return nil
// }

// func t2CReturn(p *psInterpreter) error {
// 	if p.callStack.top <= 0 {
// 		return errInvalidCFFTable
// 	}
// 	p.callStack.top--
// 	o := p.callStack.a[p.callStack.top].offset
// 	n := p.callStack.a[p.callStack.top].length
// 	buf, err := p.type2Charstrings.b.view(&p.type2Charstrings.f.src, int(o), int(n))
// 	if err != nil {
// 		return err
// 	}

// 	p.instructions = buf
// 	p.instrOffset = o
// 	p.instrLength = n
// 	return nil
// }

// func t2CEndchar(p *psInterpreter) error {
// 	t2CReadWidth(p, 0)
// 	if p.argStack.top != 0 || p.hasMoreInstructions() {
// 		if p.argStack.top == 4 {
// 			// TODO: process the implicit "seac" command as per 5177.Type2.pdf
// 			// Appendix C "Compatibility and Deprecated Operators".
// 			return errUnsupportedType2Charstring
// 		}
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.closePath()
// 	p.type2Charstrings.ended = true
// 	return nil
// }
