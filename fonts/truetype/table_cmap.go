package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	"golang.org/x/text/encoding/charmap"
)

// TableCmap defines the mapping of character codes to the glyph index values used in the font.
// It may contain more than one subtable, in order to support more than one character encoding scheme.
type TableCmap struct {
	Cmaps            []CmapSubtable
	unicodeVariation unicodeVariations
}

// FindSubtable returns the cmap for the given platform and encoding, or nil if not found.
func (t *TableCmap) FindSubtable(p PlatformID, e PlatformEncodingID) Cmap {
	key := uint32(p)>>16 | uint32(e)
	// binary search
	for i, j := 0, len(t.Cmaps); i < j; {
		h := i + (j-i)/2
		entryKey := t.Cmaps[h].key()
		if key < entryKey {
			j = h
		} else if entryKey < key {
			i = h + 1
		} else {
			return t.Cmaps[h].Cmap
		}
	}
	return nil
}

// BestEncoding returns the widest encoding supported. For valid fonts,
// the returned cmap won't be nil.
// It also reports if the returned cmap is a Symbol subtable.
func (t TableCmap) BestEncoding() (enc Cmap, isSymbolic bool) {
	// direct adaption from harfbuzz/src/hb-ot-cmap-table.hh

	// Prefer symbol if available.
	if subtable := t.FindSubtable(PlatformMicrosoft, PEMicrosoftSymbolCs); subtable != nil {
		return subtable, true
	}

	/* 32-bit subtables. */
	if cmap := t.FindSubtable(PlatformMicrosoft, PEMicrosoftUcs4); cmap != nil {
		return cmap, false
	}
	if cmap := t.FindSubtable(PlatformUnicode, PEUnicodeFull13); cmap != nil {
		return cmap, false
	}
	if cmap := t.FindSubtable(PlatformUnicode, PEUnicodeFull); cmap != nil {
		return cmap, false
	}

	/* 16-bit subtables. */
	if cmap := t.FindSubtable(PlatformMicrosoft, PEMicrosoftUnicodeCs); cmap != nil {
		return cmap, false
	}
	if cmap := t.FindSubtable(PlatformUnicode, PEUnicodeBMP); cmap != nil {
		return cmap, false
	}
	if cmap := t.FindSubtable(PlatformUnicode, 2); cmap != nil { // deprecated
		return cmap, false
	}
	if cmap := t.FindSubtable(PlatformUnicode, 1); cmap != nil { // deprecated
		return cmap, false
	}
	if cmap := t.FindSubtable(PlatformUnicode, 0); cmap != nil { // deprecated
		return cmap, false
	}

	if len(t.Cmaps) != 0 {
		return t.Cmaps[0].Cmap, false
	}
	return nil, false
}

type unicodeVariations []variationSelector

func (t unicodeVariations) getGlyphVariant(r, selector rune) (fonts.GlyphIndex, uint8) {
	// binary search
	for i, j := 0, len(t); i < j; {
		h := i + (j-i)/2
		entryKey := t[h].varSelector
		if selector < entryKey {
			j = h
		} else if entryKey < selector {
			i = h + 1
		} else {
			return t[h].getGlyph(r)
		}
	}
	return 0, variantNotFound
}

// CmapIter is an interator over a Cmap
type CmapIter interface {
	// Next returns true if the iterator still has data to yield
	Next() bool

	// Char must be called only when `Next` has returned `true`
	Char() (rune, fonts.GlyphIndex)
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
	Lookup(rune) fonts.GlyphIndex
}

type cmap0 map[rune]fonts.GlyphIndex

type cmap0Iter struct {
	data cmap0
	keys []rune
	pos  int
}

func (it *cmap0Iter) Next() bool {
	return it.pos < len(it.keys)
}

func (it *cmap0Iter) Char() (rune, fonts.GlyphIndex) {
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

func (s cmap0) Lookup(r rune) fonts.GlyphIndex {
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

func (it *cmap4Iter) Char() (r rune, gy fonts.GlyphIndex) {
	entry := it.data[it.pos1]
	if entry.indexes == nil {
		r = rune(it.pos2 + int(entry.start))
		gy = fonts.GlyphIndex(uint16(it.pos2) + entry.start + entry.delta)
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

func (s cmap4) Lookup(r rune) fonts.GlyphIndex {
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
			return fonts.GlyphIndex(c + entry.delta)
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

func (it *cmap6Iter) Char() (rune, fonts.GlyphIndex) {
	entry := it.data.entries[it.pos]
	r := rune(it.pos) + it.data.firstCode
	gy := fonts.GlyphIndex(entry)
	it.pos++
	return r, gy
}

func (s cmap6) Iter() CmapIter {
	return &cmap6Iter{data: s}
}

func (s cmap6) Lookup(r rune) fonts.GlyphIndex {
	if r < s.firstCode {
		return 0
	}
	c := int(r - s.firstCode)
	if c >= len(s.entries) {
		return 0
	}
	return fonts.GlyphIndex(s.entries[c])
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

func (it *cmap12Iter) Char() (r rune, gy fonts.GlyphIndex) {
	entry := it.data[it.pos1]
	r = rune(it.pos2 + int(entry.start))
	gy = fonts.GlyphIndex(it.pos2 + int(entry.delta))
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

func (s cmap12) Lookup(r rune) fonts.GlyphIndex {
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
			return fonts.GlyphIndex(c - entry.start + entry.delta)
		}
	}
	return 0
}

type CmapSubtable struct {
	Platform PlatformID
	Encoding PlatformEncodingID
	Cmap     Cmap
}

func (c CmapSubtable) key() uint32 { return uint32(c.Platform)>>16 | uint32(c.Encoding) }

// https://www.microsoft.com/typography/OTSPEC/cmap.htm
// direct adaption from golang.org/x/image/font/sfnt
func parseTableCmap(input []byte) (out TableCmap, err error) {
	const headerSize, entrySize = 4, 8
	if len(input) < headerSize {
		return out, errors.New("invalid 'cmap' table (EOF)")
	}
	// version is skipped
	numSubtables := int(binary.BigEndian.Uint16(input[2:]))
	if numSubtables == 0 {
		return out, errors.New("empty 'cmap' table")
	}
	if len(input) < headerSize+entrySize*numSubtables {
		return out, errors.New("invalid 'cmap' table (EOF)")
	}

	for i := 0; i < numSubtables; i++ {
		bufSubtable := input[headerSize+entrySize*i:]

		var cmap CmapSubtable
		cmap.Platform = PlatformID(binary.BigEndian.Uint16(bufSubtable))
		cmap.Encoding = PlatformEncodingID(binary.BigEndian.Uint16(bufSubtable[2:]))

		offset := binary.BigEndian.Uint32(bufSubtable[4:])
		if len(input) < int(offset)+2 { // format
			return out, errors.New("invalid cmap subtable (EOF)")
		}
		format := binary.BigEndian.Uint16(input[offset:])

		if format == 14 { // special case for variation selector
			if cmap.Platform != PlatformUnicode && cmap.Platform != 5 {
				return out, errors.New("invalid cmap subtable (EOF)")
			}
			out.unicodeVariation, err = parseCmapFormat14(input, offset)
			if err != nil {
				return out, err
			}
		} else {
			cmap.Cmap, err = parseCmapSubtable(format, input, uint32(offset))
			if err != nil {
				return out, err
			}
			out.Cmaps = append(out.Cmaps, cmap)
		}
	}

	return out, nil
}

// format 14 has already been handled
func parseCmapSubtable(format uint16, input []byte, offset uint32) (Cmap, error) {
	switch format {
	case 0:
		return parseCmapFormat0(input, offset)
	case 4:
		return parseCmapFormat4(input, offset)
	case 6:
		return parseCmapFormat6(input, offset)
	case 12:
		return parseCmapFormat12(input, offset)
	default:
		return nil, fmt.Errorf("unsupported cmap subtable format: %d", format)
	}
}

func parseCmapFormat0(input []byte, offset uint32) (cmap0, error) {
	if len(input) < int(offset)+6+256 {
		return nil, errors.New("invalid cmap subtable format 0 (EOF)")
	}

	chars := cmap0{}
	for x, index := range input[offset+6 : offset+6+256] {
		r := charmap.Macintosh.DecodeByte(byte(x))
		// The source rune r is not representable in the Macintosh-Roman encoding.
		if r != 0 {
			chars[r] = fonts.GlyphIndex(index)
		}
	}
	return chars, nil
}

func parseCmapFormat4(input []byte, offset uint32) (cmap4, error) {
	const headerSize = 14
	if len(input) < int(offset)+headerSize {
		return nil, errors.New("invalid cmap subtable format 4 (EOF)")
	}
	input = input[offset:]
	// segBuff := input[offset : offset+headerSize]
	// offset += headerSize

	segCount := int(binary.BigEndian.Uint16(input[6:]))
	if segCount&1 != 0 {
		return nil, errors.New("invalid cmap subtable format 4 (EOF)")
	}
	segCount /= 2

	input = input[headerSize:]
	eLength := 8*segCount + 2 // 2 is for the reservedPad field
	if len(input) < eLength {
		return nil, errors.New("invalid cmap subtable format 4 (EOF)")
	}
	glyphIdArray := input[eLength:]

	entries := make(cmap4, segCount)
	for i := range entries {
		cm := cmapEntry16{
			end:   binary.BigEndian.Uint16(input[2*i:]),
			start: binary.BigEndian.Uint16(input[2+2*(segCount+i):]),
			delta: binary.BigEndian.Uint16(input[2+2*(2*segCount+i):]),
		}
		offset := binary.BigEndian.Uint16(input[2+2*(3*segCount+i):])
		if offset != 0 {
			// we resolve the indexes
			cm.indexes = make([]fonts.GlyphIndex, cm.end-cm.start+1)
			start := 2*(i-segCount) + int(offset)
			if len(glyphIdArray) < start+2*len(cm.indexes) {
				return nil, errors.New("invalid cmap subtable format 4 (EOF)")
			}
			for j := range cm.indexes {
				cm.indexes[j] = fonts.GlyphIndex(binary.BigEndian.Uint16(glyphIdArray[start+2*j:]))
			}
		}
		entries[i] = cm
	}
	return entries, nil
}

func parseCmapFormat6(input []byte, offset uint32) (out cmap6, err error) {
	const headerSize = 10
	if len(input) < int(offset)+headerSize {
		return out, errors.New("invalid cmap subtable format 6 (EOF)")
	}
	input = input[offset:]

	out.firstCode = rune(binary.BigEndian.Uint16(input[6:]))
	entryCount := int(binary.BigEndian.Uint16(input[8:]))

	out.entries, err = parseUint16s(input[headerSize:], entryCount)
	return out, err
}

func parseCmapFormat12(input []byte, offset uint32) (Cmap, error) {
	const headerSize = 16
	if len(input) < int(offset)+headerSize {
		return nil, errors.New("invalid cmap subtable format 12 (EOF)")
	}
	input = input[offset:]
	// length := binary.BigEndian.Uint32(bufHeader[4:])
	numGroups := int(binary.BigEndian.Uint32(input[12:]))

	if len(input) < headerSize+12*numGroups {
		return nil, errors.New("invalid cmap subtable format 12 (EOF)")
	}

	entries := make(cmap12, numGroups)
	for i := range entries {
		entries[i] = cmapEntry32{
			start: binary.BigEndian.Uint32(input[headerSize+0+12*i:]),
			end:   binary.BigEndian.Uint32(input[headerSize+4+12*i:]),
			delta: binary.BigEndian.Uint32(input[headerSize+8+12*i:]),
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
	indexes []fonts.GlyphIndex // length end - start + 1
}

type cmapEntry32 struct {
	start, end, delta uint32
}

func parseCmapFormat14(data []byte, offset uint32) (unicodeVariations, error) {
	if len(data) < int(offset)+10 {
		return nil, errors.New("invalid cmap subtable format 14 (EOF)")
	}
	data = data[offset:]
	count := binary.BigEndian.Uint32(data[6:])

	if len(data) < 10+int(count)*11 {
		return nil, errors.New("invalid cmap subtable format 14 (EOF)")
	}
	out := make(unicodeVariations, count)
	var err error
	for i := range out {
		out[i].varSelector = parseUint24(data[10+11*i:])

		offsetDefault := binary.BigEndian.Uint32(data[10+11*i+3:])
		if offsetDefault != 0 {
			out[i].defaultUVS, err = parseUnicodeRanges(data, offsetDefault)
			if err != nil {
				return nil, err
			}
		}

		offsetNonDefault := binary.BigEndian.Uint32(data[10+11*i+7:])
		if offsetNonDefault != 0 {
			out[i].nonDefaultUVS, err = parseUVSMappings(data, offsetNonDefault)
			if err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

type variationSelector struct {
	varSelector   rune
	defaultUVS    []unicodeRange
	nonDefaultUVS []uvsMapping
}

const (
	variantNotFound = iota
	variantUseDefault
	variantFound
)

func (vs variationSelector) getGlyph(r rune) (fonts.GlyphIndex, uint8) {
	// binary search
	for i, j := 0, len(vs.defaultUVS); i < j; {
		h := i + (j-i)/2
		entry := vs.defaultUVS[h]
		if r < entry.start {
			j = h
		} else if entry.start+rune(entry.additionalCount) < r {
			i = h + 1
		} else {
			return 0, variantUseDefault
		}
	}

	for i, j := 0, len(vs.nonDefaultUVS); i < j; {
		h := i + (j-i)/2
		entry := vs.nonDefaultUVS[h].unicode
		if r < entry {
			j = h
		} else if entry < r {
			i = h + 1
		} else {
			return vs.nonDefaultUVS[h].glyphID, variantFound
		}
	}

	return 0, variantNotFound
}

type unicodeRange struct {
	start           rune
	additionalCount uint8 // 0 for a singleton range
}

func parseUnicodeRanges(data []byte, offset uint32) ([]unicodeRange, error) {
	if len(data) < int(offset)+4 {
		return nil, errors.New("invalid unicode ranges (EOF)")
	}
	count := binary.BigEndian.Uint32(data[offset:])
	if len(data) < int(offset)+4+4*int(count) {
		return nil, errors.New("invalid unicode ranges (EOF)")
	}
	data = data[4:]
	out := make([]unicodeRange, count)
	for i := range out {
		out[i].start = parseUint24(data[4*i:])
		out[i].additionalCount = data[4*i+3]
	}
	return out, nil
}

type uvsMapping struct {
	unicode rune
	glyphID fonts.GlyphIndex
}

func parseUVSMappings(data []byte, offset uint32) ([]uvsMapping, error) {
	if len(data) < int(offset)+4 {
		return nil, errors.New("invalid UVS mappings (EOF)")
	}
	count := binary.BigEndian.Uint32(data[offset:])
	if len(data) < int(offset)+4+5*int(count) {
		return nil, errors.New("invalid UVS mappings (EOF)")
	}
	data = data[4:]
	out := make([]uvsMapping, count)
	for i := range out {
		out[i].unicode = parseUint24(data[5*i:])
		out[i].glyphID = fonts.GlyphIndex(binary.BigEndian.Uint16(data[5*i+3:]))
	}
	return out, nil
}

// same as binary.BigEndian.Uint32, but for 24 bit uint
func parseUint24(b []byte) rune {
	_ = b[3] // BCE
	return rune(b[0])<<16 | rune(b[1])<<8 | rune(b[2])
}
