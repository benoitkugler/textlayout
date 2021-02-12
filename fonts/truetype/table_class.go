package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	"github.com/benoitkugler/textlayout/fonts"
)

// Class group glyph indices.
// Conceptually it is a map[fonts.GlyphIndex]uint16, but may
// be implemented more efficiently.
type Class interface {
	// ClassID returns the class ID for the provided glyph. Returns false
	// for glyphs not covered by this class.
	ClassID(fonts.GlyphIndex) (uint16, bool)

	// GlyphSize returns the number of glyphs covered.
	GlyphSize() int

	// Extent returns the maximum class ID + 1. This is the length
	// required for an array to be indexed by the class values.
	Extent() int
}

// parseClass parse `buf`, starting at `offset`.
func parseClass(buf []byte, offset uint16) (Class, error) {
	if len(buf) < int(offset)+2 {
		return nil, errors.New("invalid class table (EOF)")
	}
	buf = buf[offset:]
	switch format := binary.BigEndian.Uint16(buf); format {
	case 1:
		return parseClassFormat1(buf[2:])
	case 2:
		return parseClassLookupFormat2(buf)
	default:
		return nil, fmt.Errorf("unsupported class definition format %d", format)
	}
}

type classFormat1 struct {
	startGlyph fonts.GlyphIndex
	classIDs   []uint16 // array of target class IDs. gi is the index into that array (minus StartGlyph).
}

func (c classFormat1) ClassID(gi fonts.GlyphIndex) (uint16, bool) {
	if gi < c.startGlyph || gi >= c.startGlyph+fonts.GlyphIndex(len(c.classIDs)) {
		return 0, false
	}
	return c.classIDs[gi-c.startGlyph], true
}

func (c classFormat1) GlyphSize() int { return len(c.classIDs) }

func (c classFormat1) Extent() int {
	max := uint16(0)
	for _, cid := range c.classIDs {
		if cid >= max {
			max = cid
		}
	}
	return int(max) + 1
}

// parseClassFormat1 parses a class table, with format 1.
// For compatibility reasons, it expects `buf` to start at the first glyph,
// not at the class format.
func parseClassFormat1(buf []byte) (classFormat1, error) {
	// ClassDefFormat 1: classFormat, startGlyphID, glyphCount, []classValueArray
	const headerSize = 4 // excluding classFormat
	if len(buf) < headerSize {
		return classFormat1{}, errors.New("invalid class format 1 (EOF)")
	}

	startGI := fonts.GlyphIndex(binary.BigEndian.Uint16(buf))
	num := int(binary.BigEndian.Uint16(buf[2:]))
	classIDs, err := parseUint16s(buf[4:], num)
	if err != nil {
		return classFormat1{}, fmt.Errorf("invalid class format 1 %s", err)
	}
	return classFormat1{startGlyph: startGI, classIDs: classIDs}, nil
}

type classRangeRecord struct {
	start, end    fonts.GlyphIndex
	targetClassID uint16
}

type classFormat2 []classRangeRecord

// 'adapted' from golang/x/image/font/sfnt
func (c classFormat2) ClassID(gi fonts.GlyphIndex) (uint16, bool) {
	num := len(c)
	if num == 0 {
		return 0, false
	}

	// classRange is an array of startGlyphID, endGlyphID and target class ID.
	// Ranges are non-overlapping.
	// E.g. 130, 135, 1   137, 137, 5   etc

	idx := sort.Search(num, func(i int) bool { return gi <= c[i].start })
	// idx either points to a matching start, or to the next range (or idx==num)
	// e.g. with the range example from above: 130 points to 130-135 range, 133 points to 137-137 range

	// check if gi is the start of a range, but only if sort.Search returned a valid result
	if idx < num {
		if class := c[idx]; gi == c[idx].start {
			return class.targetClassID, true
		}
	}
	// check if gi is in previous range
	if idx > 0 {
		idx--
		if class := c[idx]; gi >= class.start && gi <= class.end {
			return class.targetClassID, true
		}
	}

	return 0, false
}

func (c classFormat2) GlyphSize() int {
	out := 0
	for _, class := range c {
		out += int(class.end - class.start + 1)
	}
	return out
}

func (c classFormat2) Extent() int {
	max := uint16(0)
	for _, r := range c {
		if r.targetClassID >= max {
			max = r.targetClassID
		}
	}
	return int(max) + 1
}

// ClassDefFormat 2: classFormat, classRangeCount, []classRangeRecords
func parseClassLookupFormat2(buf []byte) (classFormat2, error) {
	const headerSize = 4 // including classFormat
	if len(buf) < headerSize {
		return nil, errors.New("invalid class format 2 (EOF)")
	}

	num := int(binary.BigEndian.Uint16(buf[2:]))
	if len(buf) < headerSize+num*6 {
		return nil, errors.New("invalid class format 2 (EOF)")
	}

	out := make(classFormat2, num)
	for i := range out {
		out[i].start = fonts.GlyphIndex(binary.BigEndian.Uint16(buf[headerSize+i*6:]))
		out[i].end = fonts.GlyphIndex(binary.BigEndian.Uint16(buf[headerSize+i*6+2:]))
		out[i].targetClassID = binary.BigEndian.Uint16(buf[headerSize+i*6+4:])
	}
	return out, nil
}
