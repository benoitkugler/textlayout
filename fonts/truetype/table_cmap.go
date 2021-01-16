package truetype

import (
	"encoding/binary"
	"errors"

	"golang.org/x/text/encoding/charmap"
)

const (
	// This value is arbitrary, but defends against parsing malicious font
	// files causing excessive memory allocations. For reference, Adobe's
	// SourceHanSansSC-Regular.otf has 65535 glyphs and:
	//	- its format-4  cmap table has  1581 segments.
	//	- its format-12 cmap table has 16498 segments.
	maxCmapSegments = 30000
)

var be = binary.BigEndian

var (
	errUnsupportedCmapEncodings        = errors.New("unsupported cmap encodings")
	errInvalidCmapTable                = errors.New("invalid cmap table")
	errUnsupportedNumberOfCmapSegments = errors.New("unsupported number of cmap segments")
)

// GlyphIndex is a glyph index in a Font.
type GlyphIndex uint16

// CmapIter is an interator over a Cmap
type CmapIter interface {
	// Next returns true if the iterator still has data to yield
	Next() bool

	// Char must be called only when `Next` has returned `true`
	Char() (rune, GlyphIndex)
}

// Cmap stores a compact representation of a cmap,
// offering both on-demand rune lookup and full rune range.
type Cmap interface {
	// Iter returns a new iterator over the cmap
	// Multiple iterators may be used over the same cmap
	// The returned interface is garanted not to be nil
	Iter() CmapIter

	// Lookup avoid the construction of a map and provides
	// an alternative when only few runes need to be fetched.
	// It returns 0 if there is no glyph for r.
	// https://www.microsoft.com/typography/OTSPEC/cmap.htm says that
	// "Character codes that do not correspond to any glyph in the font should be mapped to glyph index 0.
	// The glyph at this location must be a special glyph representing a missing character, commonly known as .notdef."
	Lookup(rune) GlyphIndex
}

type cmap0 map[rune]GlyphIndex

type cmap0Iter struct {
	data cmap0
	keys []rune
	pos  int
}

func (it *cmap0Iter) Next() bool {
	return it.pos < len(it.keys)
}

func (it *cmap0Iter) Char() (rune, GlyphIndex) {
	r := it.keys[it.pos]
	it.pos++
	return r, it.data[r]
}

func (s cmap0) Iter() CmapIter {
	keys := make([]rune, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return &cmap0Iter{data: s, keys: keys}
}

func (s cmap0) Lookup(r rune) GlyphIndex {
	return s[r] // will be 0 if r is not in s
}

type cmap4 []cmapEntry16

type cmap4Iter struct {
	data cmap4
	pos1 int // into data
	pos2 int // either into data[pos1].indexes or an offset between start and end
}

func (it *cmap4Iter) Next() bool {
	return it.pos1 < len(it.data)
}

func (it *cmap4Iter) Char() (r rune, gy GlyphIndex) {
	entry := it.data[it.pos1]
	if entry.indexes == nil {
		r = rune(it.pos2 + int(entry.start))
		gy = GlyphIndex(uint16(it.pos2) + entry.start + entry.delta)
		if uint16(it.pos2) == entry.end-entry.start {
			// we have read the last glyph in this part
			it.pos2 = 0
			it.pos1++
		} else {
			it.pos2++
		}
	} else { // pos2 is the array index
		r = rune(it.pos2) + rune(entry.start)
		gy = entry.indexes[it.pos2]
		if it.pos2 == len(entry.indexes)-1 {
			// we have read the last glyph in this part
			it.pos2 = 0
			it.pos1++
		} else {
			it.pos2++
		}
	}

	return r, gy
}

func (s cmap4) Iter() CmapIter {
	return &cmap4Iter{data: s}
}

func (s cmap4) Lookup(r rune) GlyphIndex {
	if uint32(r) > 0xffff {
		return 0
	}
	// binary search
	c := uint16(r)
	for i, j := 0, len(s); i < j; {
		h := i + (j-i)/2
		entry := s[h]
		if c < entry.start {
			j = h
		} else if entry.end < c {
			i = h + 1
		} else if entry.indexes == nil {
			return GlyphIndex(c + entry.delta)
		} else {
			return entry.indexes[c-entry.start]
		}
	}
	return 0
}

type cmap6 struct {
	firstCode rune
	entries   []uint16
}

type cmap6Iter struct {
	data cmap6
	pos  int // index into data.entries
}

func (it *cmap6Iter) Next() bool {
	return it.pos < len(it.data.entries)
}

func (it *cmap6Iter) Char() (rune, GlyphIndex) {
	entry := it.data.entries[it.pos]
	r := rune(it.pos) + it.data.firstCode
	gy := GlyphIndex(entry)
	it.pos++
	return r, gy
}

func (s cmap6) Iter() CmapIter {
	return &cmap6Iter{data: s}
}

func (s cmap6) Lookup(r rune) GlyphIndex {
	if r < s.firstCode {
		return 0
	}
	c := int(r - s.firstCode)
	if c >= len(s.entries) {
		return 0
	}
	return GlyphIndex(s.entries[c])
}

type cmap12 []cmapEntry32

type cmap12Iter struct {
	data cmap12
	pos1 int // into data
	pos2 int // offset from start
}

func (it *cmap12Iter) Next() bool {
	return it.pos1 < len(it.data)
}

func (it *cmap12Iter) Char() (r rune, gy GlyphIndex) {
	entry := it.data[it.pos1]
	r = rune(it.pos2 + int(entry.start))
	gy = GlyphIndex(it.pos2 + int(entry.delta))
	if uint32(it.pos2) == entry.end-entry.start {
		// we have read the last glyph in this part
		it.pos2 = 0
		it.pos1++
	} else {
		it.pos2++
	}

	return r, gy
}

func (s cmap12) Iter() CmapIter {
	return &cmap12Iter{data: s}
}

func (s cmap12) Lookup(r rune) GlyphIndex {
	c := uint32(r)
	// binary search
	for i, j := 0, len(s); i < j; {
		h := i + (j-i)/2
		entry := s[h]
		if c < entry.start {
			j = h
		} else if entry.end < c {
			i = h + 1
		} else {
			return GlyphIndex(c - entry.start + entry.delta)
		}
	}
	return 0
}

// https://www.microsoft.com/typography/OTSPEC/cmap.htm
// direct adaption from golang.org/x/image/font/sfnt
func parseTableCmap(input []byte) (Cmap, error) {
	const headerSize, entrySize = 4, 8
	if len(input) < headerSize {
		return nil, errInvalidCmapTable
	}
	u := be.Uint16(input[2:4])

	numSubtables := int(u)
	if len(input) < headerSize+entrySize*numSubtables {
		return nil, errInvalidCmapTable
	}

	var (
		bestWidth  uint8
		bestOffset uint32
		bestLength uint32
		bestFormat uint16
	)

	// Scan all of the subtables, picking the widest supported one. See the
	// platformEncodingWidth comment for more discussion of width.
	for i := 0; i < numSubtables; i++ {
		bufSubtable := input[headerSize+entrySize*i : headerSize+entrySize*(i+1)]
		pid := be.Uint16(bufSubtable)
		psid := be.Uint16(bufSubtable[2:])
		width := platformEncodingWidth(pid, psid)
		if width <= bestWidth {
			continue
		}
		offset := be.Uint32(bufSubtable[4:])

		if offset > uint32(len(input)-4) {
			return nil, errInvalidCmapTable
		}
		bufFormat := input[offset : offset+4]
		format := be.Uint16(bufFormat)
		if !supportedCmapFormat(format, pid, psid) {
			continue
		}
		length := uint32(be.Uint16(bufFormat[2:]))

		bestWidth = width
		bestOffset = offset
		bestLength = length
		bestFormat = format
	}

	if bestWidth == 0 {
		return nil, errUnsupportedCmapEncodings
	}

	m, err := parseCmapIndex(input, bestOffset, bestLength, bestFormat)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Platform IDs and Platform Specific IDs as per
// https://www.microsoft.com/typography/otspec/name.htm
const (
	pidUnicode   = 0
	pidMacintosh = 1
	pidWindows   = 3

	psidUnicode2BMPOnly        = 3
	psidUnicode2FullRepertoire = 4
	// Note that FontForge may generate a bogus Platform Specific ID (value 10)
	// for the Unicode Platform ID (value 0). See
	// https://github.com/fontforge/fontforge/issues/2728

	psidMacintoshRoman = 0

	psidWindowsSymbol = 0
	psidWindowsUCS2   = 1
	psidWindowsUCS4   = 10
)

// The various cmap formats are described at
// https://www.microsoft.com/typography/otspec/cmap.htm

func supportedCmapFormat(format, pid, psid uint16) bool {
	switch format {
	case 0:
		return pid == pidMacintosh && psid == psidMacintoshRoman
	case 4:
		return true
	case 6:
		return true
	case 12:
		return true
	}
	return false
}

// platformEncodingWidth returns the number of bytes per character assumed by
// the given Platform ID and Platform Specific ID.
//
// Very old fonts, from before Unicode was widely adopted, assume only 1 byte
// per character: a character map.
//
// Old fonts, from when Unicode meant the Basic Multilingual Plane (BMP),
// assume that 2 bytes per character is sufficient.
//
// Recent fonts naturally support the full range of Unicode code points, which
// can take up to 4 bytes per character. Such fonts might still choose one of
// the legacy encodings if e.g. their repertoire is limited to the BMP, for
// greater compatibility with older software, or because the resultant file
// size can be smaller.
func platformEncodingWidth(pid, psid uint16) uint8 {
	switch pid {
	case pidUnicode:
		switch psid {
		case psidUnicode2BMPOnly:
			return 2
		case psidUnicode2FullRepertoire:
			return 4
		}

	case pidMacintosh:
		switch psid {
		case psidMacintoshRoman:
			return 1
		}

	case pidWindows:
		switch psid {
		case psidWindowsSymbol:
			return 2
		case psidWindowsUCS2:
			return 2
		case psidWindowsUCS4:
			return 4
		}
	}
	return 0
}

func parseCmapIndex(input []byte, offset, length uint32, format uint16) (Cmap, error) {
	switch format {
	case 0:
		return parseCmapFormat0(input, offset, length)
	case 4:
		return parseCmapFormat4(input, offset, length)
	case 6:
		return parseCmapFormat6(input, offset, length)
	case 12:
		return parseCmapFormat12(input, offset)
	}
	panic("unreachable")
}

func parseCmapFormat0(input []byte, offset, length uint32) (Cmap, error) {
	if length != 6+256 || offset+length > uint32(len(input)) {
		return nil, errInvalidCmapTable
	}
	var table [256]byte
	glyphsBuf := input[offset : offset+length]
	copy(table[:], glyphsBuf[6:])

	chars := cmap0{}
	for x, index := range table {
		r := charmap.Macintosh.DecodeByte(byte(x))
		// The source rune r is not representable in the Macintosh-Roman encoding.
		if r != 0 {
			chars[r] = GlyphIndex(index)
		}
	}
	return chars, nil
}

func parseCmapFormat4(input []byte, offset, length uint32) (Cmap, error) {
	const headerSize = 14
	if offset+headerSize > uint32(len(input)) {
		return nil, errInvalidCmapTable
	}
	segBuff := input[offset : offset+headerSize]
	offset += headerSize

	segCount := be.Uint16(segBuff[6:])
	if segCount&1 != 0 {
		return nil, errInvalidCmapTable
	}
	segCount /= 2
	if segCount > maxCmapSegments {
		return nil, errUnsupportedNumberOfCmapSegments
	}

	eLength := 8*uint32(segCount) + 2
	if offset+eLength > uint32(len(input)) {
		return nil, errInvalidCmapTable
	}
	glypBuf := input[offset : offset+eLength]
	offset += eLength

	indexesBase := offset
	indexesLength := uint32(len(input)) - offset

	entries := make(cmap4, segCount)
	L := int(segCount)
	for i := 0; i < L; i++ {
		cm := cmapEntry16{
			end:   be.Uint16(glypBuf[0*L+0+2*i:]),
			start: be.Uint16(glypBuf[2*L+2+2*i:]),
			delta: be.Uint16(glypBuf[4*L+2+2*i:]),
		}
		offset := be.Uint16(glypBuf[6*L+2+2*i:])
		if offset != 0 {
			// we resolve the indexes
			cm.indexes = make([]GlyphIndex, cm.end-cm.start+1)
			for j := range cm.indexes {
				glyphOffset := uint32(offset) + 2*uint32(i-int(segCount)+j)
				if glyphOffset > indexesLength || glyphOffset+2 > indexesLength {
					return nil, errInvalidCmapTable
				}
				x := input[indexesBase+glyphOffset : indexesBase+glyphOffset+2]
				cm.indexes[j] = GlyphIndex(be.Uint16(x))
			}
		}
		entries[i] = cm
	}
	return entries, nil
}

func parseCmapFormat6(input []byte, offset, length uint32) (Cmap, error) {
	const headerSize = 10
	if offset+headerSize > uint32(len(input)) {
		return nil, errInvalidCmapTable
	}
	bufHeader := input[offset : offset+headerSize]
	offset += headerSize

	firstCode := be.Uint16(bufHeader[6:])
	entryCount := be.Uint16(bufHeader[8:])

	eLength := 2 * uint32(entryCount)
	if offset+eLength > uint32(len(input)) {
		return nil, errInvalidCmapTable
	}

	bufGlyph := input[offset : offset+eLength]
	offset += eLength

	entries := make([]uint16, entryCount)
	for i := range entries {
		entries[i] = be.Uint16(bufGlyph[2*i:])
	}
	return cmap6{firstCode: rune(firstCode), entries: entries}, nil
}

func parseCmapFormat12(input []byte, offset uint32) (Cmap, error) {
	const headerSize = 16
	if offset+headerSize > uint32(len(input)) {
		return nil, errInvalidCmapTable
	}
	bufHeader := input[offset : offset+headerSize]
	length := be.Uint32(bufHeader[4:])
	if uint32(len(input)) < offset || length > uint32(len(input))-offset {
		return nil, errInvalidCmapTable
	}
	offset += headerSize

	numGroups := be.Uint32(bufHeader[12:])
	if numGroups > maxCmapSegments {
		return nil, errUnsupportedNumberOfCmapSegments
	}

	eLength := 12 * numGroups
	if headerSize+eLength != length {
		return nil, errInvalidCmapTable
	}
	bufGlyphs := input[offset : offset+eLength]
	offset += eLength

	entries := make(cmap12, numGroups)
	for i := range entries {
		entries[i] = cmapEntry32{
			start: be.Uint32(bufGlyphs[0+12*i:]),
			end:   be.Uint32(bufGlyphs[4+12*i:]),
			delta: be.Uint32(bufGlyphs[8+12*i:]),
		}
	}
	return entries, nil
}

// if indexes is nil, delta is used
type cmapEntry16 struct {
	end, start uint16
	delta      uint16
	// we prefere not to keep a link to a buffer (via an offset)
	// and eagerly resolve it
	indexes []GlyphIndex // length end - start + 1
}

type cmapEntry32 struct {
	start, end, delta uint32
}
