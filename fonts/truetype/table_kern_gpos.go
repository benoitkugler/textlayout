package truetype

import (
	"errors"
	"fmt"
	"sort"
)

var (
	errInvalidGPOSKern           = errors.New("invalid GPOS kerning subtable")
	errUnsupportedClassDefFormat = errors.New("unsupported class definition format")
)

func (t TableLayout) parseKern() (Kerns, error) {
	var kerns kernUnions

	for _, lookup := range t.Lookups {
		if lookup.Type == 2 {
			for _, subtableOffset := range lookup.subtableOffsets {
				b := lookup.data
				if len(b) < 4+int(subtableOffset) {
					return nil, errInvalidGPOSKern
				}
				b = b[subtableOffset:]
				format, coverageOffset := be.Uint16(b), be.Uint16(b[2:])

				coverage, err := fetchCoverage(b, int(coverageOffset))
				if err != nil {
					return nil, err
				}

				switch format {
				case 1: // Adjustments for Glyph Pairs
					kern, err := parsePairPosFormat1(b, coverage)
					if err != nil {
						return nil, err
					}
					kerns = append(kerns, kern)
				case 2: // Class Pair Adjustment
					kern, err := parsePairPosFormat2(b, coverage)
					if err != nil {
						return nil, err
					}
					kerns = append(kerns, kern)
				}
			}
		}
	}

	if len(kerns) == 0 {
		// no kerning information
		return nil, errors.New("missing GPOS kerning information")
	}

	return kerns, nil
}

type coverage interface {
	// returns the index into a PairPos table for the provided glyph.
	// Returns false if the glyph is not covered by this lookup.
	// Note: this method is injective: two distincts, covered glyphs are mapped
	// to distincts tables
	tableIndex(GlyphIndex) (int, bool)
}

// if l[i] = gi then gi has coverage index of i
func fetchCoverage(buf []byte, offset int) (coverage, error) {
	if len(buf) < offset+2 { // format and count
		return nil, errInvalidGPOSKern
	}
	buf = buf[offset:]
	switch format := be.Uint16(buf); format {
	case 1:
		// Coverage Format 1: coverageFormat, glyphCount, []glyphArray
		return fetchCoverageList(buf[2:])
	case 2:
		// Coverage Format 2: coverageFormat, rangeCount, []rangeRecords{startGlyphID, endGlyphID, startCoverageIndex}
		return fetchCoverageRange(buf[2:])
	default:
		return nil, fmt.Errorf("unsupported GPOS coverage format %d", format)
	}
}

// sorted in ascending order
type coverageList []GlyphIndex

func (cl coverageList) tableIndex(gi GlyphIndex) (int, bool) {
	num := len(cl)
	idx := sort.Search(num, func(i int) bool { return gi <= cl[i] })
	if idx < num && cl[idx] == gi {
		return idx, true
	}
	return 0, false
}

// func (cl coverageList) maxIndex() int { return len(cl) - 1 }

func fetchCoverageList(buf []byte) (coverageList, error) {
	const headerSize, entrySize = 2, 2
	if len(buf) < headerSize {
		return nil, errInvalidGPOSKern
	}

	num := int(be.Uint16(buf))
	if len(buf) < headerSize+num*entrySize {
		return nil, errInvalidGPOSKern
	}

	out := make(coverageList, num)
	for i := range out {
		out[i] = GlyphIndex(be.Uint16(buf[headerSize+2*i:]))
	}
	return out, nil
}

type coverageRange struct {
	start, end    GlyphIndex
	startCoverage int
}

// coverageRanges is an array of startGlyphID, endGlyphID and startCoverageIndex
// Ranges are non-overlapping.
// The following GlyphIDs/index pairs are stored as follows:
//	 pairs: 130=0, 131=1, 132=2, 133=3, 134=4, 135=5, 137=6
//   ranges: 130, 135, 0    137, 137, 6
// startCoverageIndex is used to calculate the index without counting
// the length of the preceeding ranges
type coverageRanges []coverageRange

func (cr coverageRanges) tableIndex(gi GlyphIndex) (int, bool) {
	num := len(cr)
	if num == 0 {
		return 0, false
	}

	idx := sort.Search(num, func(i int) bool { return gi <= cr[i].start })
	// idx either points to a matching start, or to the next range (or idx==num)
	// e.g. with the range example from above: 130 points to 130-135 range, 133 points to 137-137 range

	// check if gi is the start of a range, but only if sort.Search returned a valid result
	if idx < num {
		if rang := cr[idx]; gi == rang.start {
			return int(rang.startCoverage), true
		}
	}
	// check if gi is in previous range
	if idx > 0 {
		idx--
		if rang := cr[idx]; gi >= rang.start && gi <= rang.end {
			return rang.startCoverage + int(gi-rang.start), true
		}
	}

	return 0, false
}

// func (cr coverageRanges) maxIndex() int {
// 	lastRange := cr[len(cr)-1]
// 	return lastRange.startCoverage + int(lastRange.end-lastRange.start)
// }

func fetchCoverageRange(buf []byte) (coverageRanges, error) {
	const headerSize, entrySize = 2, 6
	if len(buf) < headerSize {
		return nil, errInvalidGPOSKern
	}

	num := int(be.Uint16(buf))
	if len(buf) < headerSize+num*entrySize {
		return nil, errInvalidGPOSKern
	}

	out := make(coverageRanges, num)
	for i := range out {
		out[i].start = GlyphIndex(be.Uint16(buf[headerSize+i*entrySize:]))
		out[i].end = GlyphIndex(be.Uint16(buf[headerSize+i*entrySize+2:]))
		out[i].startCoverage = int(be.Uint16(buf[headerSize+i*entrySize+4:]))
	}
	return out, nil
}

// offset int
func parsePairPosFormat1(buf []byte, coverage coverage) (pairPosKern, error) {
	// PairPos Format 1: posFormat, coverageOffset, valueFormat1,
	// valueFormat2, pairSetCount, []pairSetOffsets
	const headerSize = 10 // including posFormat and coverageOffset
	if len(buf) < headerSize {
		return pairPosKern{}, errInvalidGPOSKern
	}
	valueFormat1, valueFormat2, nPairs := be.Uint16(buf[4:]), be.Uint16(buf[6:]), int(be.Uint16(buf[8:]))

	// check valueFormat1 and valueFormat2 flags
	if valueFormat1 != 0x04 || valueFormat2 != 0x00 {
		// we only support kerning with X_ADVANCE for first glyph
		return pairPosKern{}, nil
	}

	// PairPos table contains an array of offsets to PairSet
	// tables, which contains an array of PairValueRecords.
	// Calculate length of complete PairPos table by jumping to
	// last PairSet.
	// We need to iterate all offsets to find the last pair as
	// offsets are not sorted and can be repeated.
	if len(buf) < headerSize+nPairs*2 {
		return pairPosKern{}, errInvalidGPOSKern
	}
	var lastPairSetOffset int
	for n := 0; n < nPairs; n++ {
		pairOffset := int(be.Uint16(buf[headerSize+n*2:]))
		if pairOffset > lastPairSetOffset {
			lastPairSetOffset = pairOffset
		}
	}

	if len(buf) < lastPairSetOffset+2 {
		return pairPosKern{}, errInvalidGPOSKern
	}

	pairValueCount := int(be.Uint16(buf[lastPairSetOffset:]))
	// Each PairSet contains the secondGlyph (u16) and one or more value records (all u16).
	// We only support lookup tables with one value record (X_ADVANCE, see valueFormat1/2 above).
	lastPairSetLength := 2 + pairValueCount*4

	length := lastPairSetOffset + lastPairSetLength
	if len(buf) < length {
		return pairPosKern{}, errInvalidGPOSKern
	}
	return fetchPairPosGlyph(coverage, nPairs, buf)
}

type pairKern struct {
	right GlyphIndex
	kern  int16
}

// slice indexed by tableIndex
type pairPosKern struct {
	cov  coverage
	list [][]pairKern
}

func (pp pairPosKern) KernPair(a, b GlyphIndex) (int16, bool) {
	idx, found := pp.cov.tableIndex(a)
	if !found {
		return 0, false
	}
	if idx >= len(pp.list) { // coverage might be corrupted
		return 0, false
	}

	list := pp.list[idx]
	for _, secondGlyphIndex := range list {
		if secondGlyphIndex.right == b {
			return secondGlyphIndex.kern, true
		}
		if secondGlyphIndex.right > b { // list is sorted
			return 0, false
		}
	}
	return 0, false
}

func (pp pairPosKern) Size() int {
	out := 0
	for _, l := range pp.list {
		out += len(l)
	}
	return out
}

func fetchPairPosGlyph(coverage coverage, num int, glyphs []byte) (pairPosKern, error) {
	// glyphs length is checked before calling this function

	lists := make([][]pairKern, num)
	for idx := range lists {
		offset := int(be.Uint16(glyphs[10+idx*2:]))
		if offset+1 >= len(glyphs) {
			return pairPosKern{}, errInvalidGPOSKern
		}

		count := int(be.Uint16(glyphs[offset:]))
		if len(glyphs) < offset+2+4*count {
			return pairPosKern{}, errInvalidGPOSKern
		}

		list := make([]pairKern, count)
		for i := range list {
			list[i] = pairKern{
				right: GlyphIndex(be.Uint16(glyphs[offset+2+i*4:])),
				kern:  int16(be.Uint16(glyphs[offset+2+i*4+2:])),
			}
		}
		lists[idx] = list
	}
	return pairPosKern{cov: coverage, list: lists}, nil
}

type classKerns struct {
	coverage       coverage
	class1, class2 class
	numClass2      int
	kerns          []int16 // size numClass1 * numClass2
}

func (c classKerns) KernPair(left, right GlyphIndex) (int16, bool) {
	// check coverage to avoid selection of default class 0
	_, found := c.coverage.tableIndex(left)
	if !found {
		return 0, false
	}
	idxa := c.class1.glyphClassID(left)
	idxb := c.class2.glyphClassID(right)
	return c.kerns[idxb+idxa*c.numClass2], true
}

func (c classKerns) Size() int { return c.class1.size() * c.class2.size() }

func parsePairPosFormat2(buf []byte, coverage coverage) (classKerns, error) {
	// PairPos Format 2:
	// posFormat, coverageOffset, valueFormat1, valueFormat2,
	// classDef1Offset, classDef2Offset, class1Count, class2Count,
	// []class1Records
	const headerSize = 16 // including posFormat and coverageOffset
	if len(buf) < headerSize {
		return classKerns{}, errInvalidGPOSKern
	}

	valueFormat1, valueFormat2 := be.Uint16(buf[4:]), be.Uint16(buf[6:])
	// check valueFormat1 and valueFormat2 flags
	if valueFormat1 != 0x04 || valueFormat2 != 0x00 {
		// we only support kerning with X_ADVANCE for first glyph
		return classKerns{}, nil
	}

	cdef1Offset := int(be.Uint16(buf[8:]))
	cdef2Offset := int(be.Uint16(buf[10:]))
	numClass1 := int(be.Uint16(buf[12:]))
	numClass2 := int(be.Uint16(buf[14:]))
	// var cdef1, cdef2 classLookupFunc
	cdef1, err := fetchClassLookup(buf, cdef1Offset)
	if err != nil {
		return classKerns{}, err
	}
	cdef2, err := fetchClassLookup(buf, cdef2Offset)
	if err != nil {
		return classKerns{}, err
	}

	return fetchPairPosClass(
		buf[headerSize:],
		coverage,
		numClass1,
		numClass2,
		cdef1,
		cdef2,
	)
}

func fetchClassLookup(buf []byte, offset int) (class, error) {
	if len(buf) < offset+2 {
		return nil, errInvalidGPOSKern
	}
	buf = buf[offset:]
	switch be.Uint16(buf) {
	case 1:
		return fetchClassLookupFormat1(buf)
	case 2:
		// ClassDefFormat 2: classFormat, classRangeCount, []classRangeRecords
		return fetchClassLookupFormat2(buf)
	default:
		return nil, errUnsupportedClassDefFormat
	}
}

type class interface {
	// glyphClassIDreturns the class ID for the provided glyph. Returns 0
	// (default class) for glyphs not covered by this lookup.
	glyphClassID(GlyphIndex) int
	size() int // return the number of glyh
}

type classFormat1 struct {
	startGlyph     GlyphIndex
	targetClassIDs []int // array of target class IDs. gi is the index into that array (minus startGI).
}

func (c classFormat1) glyphClassID(gi GlyphIndex) int {
	if gi < c.startGlyph || gi >= c.startGlyph+GlyphIndex(len(c.targetClassIDs)) {
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

	startGI := GlyphIndex(be.Uint16(buf[2:]))
	num := int(be.Uint16(buf[4:]))
	if len(buf) < headerSize+num*2 {
		return classFormat1{}, errInvalidGPOSKern
	}

	classIDs := make([]int, num)
	for i := range classIDs {
		classIDs[i] = int(be.Uint16(buf[6+i*2:]))
	}
	return classFormat1{startGlyph: startGI, targetClassIDs: classIDs}, nil
}

type classRangeRecord struct {
	start, end    GlyphIndex
	targetClassID int
}

type class2 []classRangeRecord

// 'adapted' from golang/x/image/font/sfnt
func (c class2) glyphClassID(gi GlyphIndex) int {
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
		out[i].start = GlyphIndex(be.Uint16(buf[headerSize+i*6:]))
		out[i].end = GlyphIndex(be.Uint16(buf[headerSize+i*6+2:]))
		out[i].targetClassID = int(be.Uint16(buf[headerSize+i*6+4:]))
	}
	return out, nil
}

func fetchPairPosClass(buf []byte, cov coverage, num1, num2 int, cdef1, cdef2 class) (classKerns, error) {
	if len(buf) < num1*num2*2 {
		return classKerns{}, errInvalidGPOSKern
	}

	kerns := make([]int16, num1*num2)
	for i := 0; i < num1; i++ {
		for j := 0; j < num2; j++ {
			index := j + i*num2
			kerns[index] = int16(be.Uint16(buf[index*2:]))
		}
	}

	return classKerns{
		coverage:  cov,
		class1:    cdef1,
		class2:    cdef2,
		kerns:     kerns,
		numClass2: num2,
	}, nil
}
