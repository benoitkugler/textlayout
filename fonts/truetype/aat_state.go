package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

// state tables: https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6Tables.html#StateTables

// AATStateTable is an extended state table.
type AATStateTable struct {
	nClasses uint32 //  for some reasons, this may differ from Class.Extent
	class    Class
	states   []uint16
	entries  []AATStateEntry
}

// extended is true for morx, false for mort
// data must be truncated to the end of the state table data
// every `Data` field of the entries will be of length `entryDataSize`
// meaning they can later be safely interpreted
func parseStateTable(data []byte, entryDataSize int) (out AATStateTable, err error) {
	var classOffset, stateOffset, entryOffset uint32
	if len(data) < 16 {
		return out, errors.New("invalid morx state table (EOF)")
	}
	out.nClasses = binary.BigEndian.Uint32(data)
	classOffset = binary.BigEndian.Uint32(data[4:])
	stateOffset = binary.BigEndian.Uint32(data[8:])
	entryOffset = binary.BigEndian.Uint32(data[12:])

	// Ensure pre-defined classes fit.
	if out.nClasses < 4 {
		return out, fmt.Errorf("invalid number of classes in state table: %d", out.nClasses)
	}

	out.class, err = parseAATLookupTable(data, classOffset, stateOffset)
	if err != nil {
		return out, err
	}

	if stateOffset > entryOffset || len(data) < int(entryOffset) {
		return out, errors.New("invalid morx state table (EOF)")
	}
	out.states, err = parseUint16s(data[stateOffset:entryOffset], int(entryOffset-stateOffset)/2)

	out.entries = parseStateEntries(data[entryOffset:], entryDataSize)

	// sanitize states indexes
	for _, stateIndex := range out.states {
		if int(stateIndex) >= len(out.entries) {
			return out, errors.New("invalid morx state table (EOF)")
		}
	}

	return out, err
}

// GetClass return the class for the given glyph, with the correct default value.
func (t *AATStateTable) GetClass(glyph fonts.GlyphIndex) uint16 {
	if glyph == 0xFFFF { // deleted glyph
		return 2 // class deleted
	}
	c, ok := t.class.ClassID(glyph)
	if !ok {
		return 1 // class out of bounds
	}
	return c
}

// GetEntry return the entry for the given state and class,
// and handle invalid values.
func (t AATStateTable) GetEntry(state, class uint16) AATStateEntry {
	if uint32(class) >= t.nClasses {
		class = 1 // class out of bounds
	}
	// TODO: if possible, sanitize early and remove the check
	index := int(uint32(state)*t.nClasses + uint32(class))
	if index >= len(t.states) {
		return AATStateEntry{}
	}
	entry := t.states[index]
	return t.entries[entry]
}

// AATStateEntry is the shared type for entries
// in a state table. See the various AsXXX methods
// to exploit its content.
type AATStateEntry struct {
	NewState uint16
	Flags    uint16 // Table specific.
	// Remaining of the entry, context specific
	data [4]byte
}

// data is at the start of the entries array
// assume extraDataSize <= 4
func parseStateEntries(data []byte, extraDataSize int) []AATStateEntry {
	entrySize := 4 + extraDataSize
	count := len(data) / entrySize
	out := make([]AATStateEntry, count)
	for i := range out {
		out[i].NewState = binary.BigEndian.Uint16(data[i*entrySize:])
		out[i].Flags = binary.BigEndian.Uint16(data[i*entrySize+2:])
		copy(out[i].data[:], data[i*entrySize+4:(i+1)*entrySize])
	}
	return out
}

// AsMorxContextual reads the internal data for entries in morx contextual subtable.
// The returned indexes use 0xFFFF as empty value.
func (e AATStateEntry) AsMorxContextual() (markIndex, currentIndex uint16) {
	markIndex = binary.BigEndian.Uint16(e.data[:])
	currentIndex = binary.BigEndian.Uint16(e.data[2:])
	return
}

// AsMorxInsertion reads the internal data for entries in morx insertion subtable.
// The returned indexes use 0xFFFF as empty value.
func (e AATStateEntry) AsMorxInsertion() (currentIndex, markedIndex uint16) {
	currentIndex = binary.BigEndian.Uint16(e.data[:])
	markedIndex = binary.BigEndian.Uint16(e.data[2:])
	return
}

// AsMorxLigature reads the internal data for entries in morx ligature subtable.
func (e AATStateEntry) AsMorxLigature() (ligActionIndex uint16) {
	return binary.BigEndian.Uint16(e.data[:])
}

// AAT lookup implementing Class

type lookupFormat0 []uint16

func (l lookupFormat0) ClassID(gid fonts.GlyphIndex) (uint16, bool) {
	if int(gid) >= len(l) {
		return 0, false
	}
	return l[gid], true
}

func (l lookupFormat0) GlyphSize() int { return len(l) }

func (l lookupFormat0) Extent() int {
	max := uint16(0)
	for _, r := range l {
		if r >= max {
			max = r
		}
	}
	return int(max) + 1
}

// data is bounded by the next offset
func parseAATLookupFormat0(data []byte) (lookupFormat0, error) {
	return parseUint16s(data[2:], (len(data)-2)/2)
}

// lookupFormat2 is the same as classFormat2, but with start and end are reversed in the binary
func parseAATLookupFormat2(data []byte) (classFormat2, error) {
	const headerSize = 12 // including classFormat
	if len(data) < headerSize {
		return nil, errors.New("invalid AAT lookup format 2 (EOF)")
	}

	unitSize := binary.BigEndian.Uint16(data[2:])
	num := int(binary.BigEndian.Uint16(data[4:]))
	// 3 other field ignored
	if unitSize != 6 {
		return nil, fmt.Errorf("unexpected AAT lookup segment size: %d", unitSize)
	}

	if len(data) < headerSize+num*6 {
		return nil, errors.New("invalid AAT lookup format 2 (EOF)")
	}

	out := make(classFormat2, num)
	for i := range out {
		out[i].end = fonts.GlyphIndex(binary.BigEndian.Uint16(data[headerSize+i*6:]))
		out[i].start = fonts.GlyphIndex(binary.BigEndian.Uint16(data[headerSize+i*6+2:]))
		out[i].targetClassID = binary.BigEndian.Uint16(data[headerSize+i*6+4:])
	}
	return out, nil
}

// lookupFormat4 is the same as  classFormat2, but with start and end are reversed in the binary and offset are used
func parseAATLookupFormat4(data []byte) (classFormat2, error) {
	const headerSize = 12 // including classFormat
	if len(data) < headerSize {
		return nil, errors.New("invalid AAT lookup format 4 (EOF)")
	}

	unitSize := binary.BigEndian.Uint16(data[2:])
	num := int(binary.BigEndian.Uint16(data[4:]))
	// 3 other field ignored
	if unitSize != 6 {
		return nil, fmt.Errorf("unexpected AAT lookup segment size: %d", unitSize)
	}

	if len(data) < headerSize+num*6 {
		return nil, errors.New("invalid AAT lookup format 4 (EOF)")
	}

	out := make(classFormat2, num)
	for i := range out {
		out[i].end = fonts.GlyphIndex(binary.BigEndian.Uint16(data[headerSize+i*6:]))
		out[i].start = fonts.GlyphIndex(binary.BigEndian.Uint16(data[headerSize+i*6+2:]))
		offset := int(binary.BigEndian.Uint16(data[headerSize+i*6+4:]))
		if len(data) < offset+2 {
			return nil, errors.New("invalid AAT lookup format 4 (EOF)")
		}
		out[i].targetClassID = binary.BigEndian.Uint16(data[offset:])
	}
	return out, nil
}

// sorted pairs of GlyphIndex, value
type lookupFormat6 []struct {
	gid   fonts.GlyphIndex
	value uint16
}

func (l lookupFormat6) ClassID(gid fonts.GlyphIndex) (uint16, bool) {
	// binary search
	for i, j := 0, len(l); i < j; {
		h := i + (j-i)/2
		entry := l[h]
		if gid < entry.gid {
			j = h
		} else if entry.gid < gid {
			i = h + 1
		} else {
			return entry.value, true
		}
	}
	return 0, false
}

func (l lookupFormat6) GlyphSize() int { return len(l) }

func (l lookupFormat6) Extent() int {
	max := uint16(0)
	for _, r := range l {
		if r.value >= max {
			max = r.value
		}
	}
	return int(max) + 1
}

func parseAATLookupFormat6(data []byte) (lookupFormat6, error) {
	const headerSize = 12 // including classFormat
	if len(data) < headerSize {
		return nil, errors.New("invalid AAT lookup format 6 (EOF)")
	}

	unitSize := binary.BigEndian.Uint16(data[2:])
	num := int(binary.BigEndian.Uint16(data[4:]))
	// 3 other field ignored
	if unitSize != 4 {
		return nil, fmt.Errorf("unexpected AAT lookup segment size: %d", unitSize)
	}

	if len(data) < headerSize+num*4 {
		return nil, errors.New("invalid AAT lookup format 6 (EOF)")
	}

	out := make(lookupFormat6, num)
	for i := range out {
		out[i].gid = fonts.GlyphIndex(binary.BigEndian.Uint16(data[headerSize+i*4:]))
		out[i].value = binary.BigEndian.Uint16(data[headerSize+i*4+2:])
	}
	return out, nil
}

// lookupFormat8 is the same as ClassFormat1
func parseAATLookupFormat8(data []byte) (classFormat1, error) {
	return parseClassFormat1(data)
}

// in this context (value of two bytes) lookupFormat10 is the same as ClassFormat1
func parseAATLookupFormat10(data []byte) (classFormat1, error) {
	return parseClassFormat1(data)
}

// nextOffset is used for unbounded lookups
func parseAATLookupTable(data []byte, offset, nextOffset uint32) (Class, error) {
	if len(data) < int(offset)+2 || len(data) < int(nextOffset) {
		return nil, errors.New("invalid AAT lookup table (EOF)")
	}
	data = data[offset:nextOffset]

	switch format := binary.BigEndian.Uint16(data); format {
	case 0:
		return parseAATLookupFormat0(data)
	case 2:
		return parseAATLookupFormat2(data)
	case 4:
		return parseAATLookupFormat4(data)
	case 6:
		return parseAATLookupFormat6(data)
	case 8:
		return parseAATLookupFormat8(data)
	case 10:
		return parseAATLookupFormat10(data)
	default:
		return nil, fmt.Errorf("invalid AAT lookup table kind : %d", format)
	}
}
