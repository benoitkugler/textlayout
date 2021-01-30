package truetype

import (
	"sort"

	"github.com/benoitkugler/textlayout/fonts"
)

func fetchClassLookup(buf []byte, offset uint16) (class, error) {
	if len(buf) < int(offset)+2 {
		return nil, errInvalidGPOSKern
	}
	buf = buf[offset:]
	switch be.Uint16(buf) {
	case 1:
		return fetchClassLookupFormat1(buf)
	case 2:
		return fetchClassLookupFormat2(buf)
	default:
		return nil, errUnsupportedClassDefFormat
	}
}

type class interface {
	// ClassIDreturns the class ID for the provided glyph. Returns 0
	// (default class) for glyphs not covered by this lookup.
	ClassID(fonts.GlyphIndex) uint16
	size() int // return the number of glyh
}

type classFormat1 struct {
	startGlyph     fonts.GlyphIndex
	targetClassIDs []uint16 // array of target class IDs. gi is the index into that array (minus startGI).
}

func (c classFormat1) ClassID(gi fonts.GlyphIndex) uint16 {
	if gi < c.startGlyph || gi >= c.startGlyph+fonts.GlyphIndex(len(c.targetClassIDs)) {
		return 0
	}
	return c.targetClassIDs[gi-c.startGlyph]
}

func (c classFormat1) size() int { return len(c.targetClassIDs) }

// ClassDefFormat 1: classFormat, startGlyphID, glyphCount, []classValueArray
func fetchClassLookupFormat1(buf []byte) (classFormat1, error) {
	const headerSize = 6 // including classFormat
	if len(buf) < headerSize {
		return classFormat1{}, errInvalidGPOSKern
	}

	startGI := fonts.GlyphIndex(be.Uint16(buf[2:]))
	num := int(be.Uint16(buf[4:]))
	if len(buf) < headerSize+num*2 {
		return classFormat1{}, errInvalidGPOSKern
	}

	classIDs := make([]uint16, num)
	for i := range classIDs {
		classIDs[i] = be.Uint16(buf[6+i*2:])
	}
	return classFormat1{startGlyph: startGI, targetClassIDs: classIDs}, nil
}

type classRangeRecord struct {
	start, end    fonts.GlyphIndex
	targetClassID uint16
}

type class2 []classRangeRecord

// 'adapted' from golang/x/image/font/sfnt
func (c class2) ClassID(gi fonts.GlyphIndex) uint16 {
	num := len(c)
	if num == 0 {
		return 0 // default to class 0
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
			return class.targetClassID
		}
	}
	// check if gi is in previous range
	if idx > 0 {
		idx--
		if class := c[idx]; gi >= class.start && gi <= class.end {
			return class.targetClassID
		}
	}
	// default to class 0
	return 0
}

func (c class2) size() int {
	out := 0
	for _, class := range c {
		out += int(class.end - class.start + 1)
	}
	return out
}

// ClassDefFormat 2: classFormat, classRangeCount, []classRangeRecords
func fetchClassLookupFormat2(buf []byte) (class2, error) {
	const headerSize = 4 // including classFormat
	if len(buf) < headerSize {
		return nil, errInvalidGPOSKern
	}

	num := int(be.Uint16(buf[2:]))
	if len(buf) < headerSize+num*6 {
		return nil, errInvalidGPOSKern
	}

	out := make(class2, num)
	for i := range out {
		out[i].start = fonts.GlyphIndex(be.Uint16(buf[headerSize+i*6:]))
		out[i].end = fonts.GlyphIndex(be.Uint16(buf[headerSize+i*6+2:]))
		out[i].targetClassID = be.Uint16(buf[headerSize+i*6+4:])
	}
	return out, nil
}
