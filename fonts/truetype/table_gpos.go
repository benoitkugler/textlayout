package truetype

import (
	"errors"
	"fmt"
	"sort"

	"github.com/benoitkugler/textlayout/fonts"
)

var (
	errInvalidGPOSKern           = errors.New("invalid GPOS kerning subtable")
	errUnsupportedClassDefFormat = errors.New("unsupported class definition format")
)

type TableGPOS struct {
	TableLayout
	lookups []lookup
}

func parseTableGPOS(data []byte) (*TableGPOS, error) {
	tableLayout, lookups, err := parseTableLayout(data)
	if err != nil {
		return nil, err
	}
	return &TableGPOS{
		TableLayout: tableLayout,
		lookups:     lookups,
	}, nil
}

// TODO:
func (lookup lookup) parseGPOS() (kernUnions, error) {
	switch lookup.kind {
	case 2:
		var kerns kernUnions
		for _, subtableOffset := range lookup.subtableOffsets {
			b := lookup.data
			if len(b) < 4+int(subtableOffset) {
				return nil, errInvalidGPOSKern
			}
			b = b[subtableOffset:]
			format, coverageOffset := be.Uint16(b), be.Uint16(b[2:])

			coverage, err := parseCoverage(b, coverageOffset)
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
		return kerns, nil
	}
	return nil, nil
}

func (t *TableGPOS) parseKern() (Kerns, error) {
	var kerns kernUnions
	for _, lookup := range t.lookups {
		ks, err := lookup.parseGPOS()
		if err != nil {
			return nil, err
		}
		kerns = append(kerns, ks...)
	}

	if len(kerns) == 0 {
		// no kerning information
		return nil, errors.New("missing GPOS kerning information")
	}

	return kerns, nil
}

// Coverage specifies all the glyphs affected by a substitution or
// positioning operation described in a subtable.
// Conceptually is it a []GlyphIndex, but it may be implemented for efficiently.
// See the concrete types `CoverageList` and `CoverageRanges`.
type Coverage interface {
	// Index returns the index of the provided glyph, or
	// `false` if the glyph is not covered by this lookup.
	// Note: this method is injective: two distincts, covered glyphs are mapped
	// to distincts tables.
	Index(fonts.GlyphIndex) (int, bool)

	// Size return the number of glyphs covered
	Size() int
}

// if l[i] = gi then gi has coverage index of i
func parseCoverage(buf []byte, offset uint16) (Coverage, error) {
	if len(buf) < int(offset)+2 { // format and count
		return nil, errors.New("invalid coverage table")
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
		return nil, fmt.Errorf("unsupported coverage format %d", format)
	}
}

// CoverageList is a coverage with format 1.
// The glyphs are sorted in ascending order.
type CoverageList []fonts.GlyphIndex

func (cl CoverageList) Index(gi fonts.GlyphIndex) (int, bool) {
	num := len(cl)
	idx := sort.Search(num, func(i int) bool { return gi <= cl[i] })
	if idx < num && cl[idx] == gi {
		return idx, true
	}
	return 0, false
}

func (cl CoverageList) Size() int { return len(cl) }

// func (cl coverageList) maxIndex() int { return len(cl) - 1 }

func fetchCoverageList(buf []byte) (CoverageList, error) {
	const headerSize, entrySize = 2, 2
	if len(buf) < headerSize {
		return nil, errInvalidGPOSKern
	}

	num := int(be.Uint16(buf))
	if len(buf) < headerSize+num*entrySize {
		return nil, errInvalidGPOSKern
	}

	out := make(CoverageList, num)
	for i := range out {
		out[i] = fonts.GlyphIndex(be.Uint16(buf[headerSize+2*i:]))
	}
	return out, nil
}

// CoverageRange store a range of indexes, starting from StartCoverage.
// For example, for the glyphs 12,13,14,15, and the indexes 7,8,9,10,
// the CoverageRange would be {12, 15, 7}.
type CoverageRange struct {
	Start, End    fonts.GlyphIndex
	StartCoverage int
}

// CoverageRanges is a coverage with format 2.
// Ranges are non-overlapping.
// The following GlyphIDs/index pairs are stored as follows:
//	 glyphs: 130, 131, 132, 133, 134, 135, 137
//	 indexes: 0, 1, 2, 3, 4, 5, 6
//   ranges: {130, 135, 0}    {137, 137, 6}
// StartCoverage is used to calculate the index without counting
// the length of the preceeding ranges
type CoverageRanges []CoverageRange

func (cr CoverageRanges) Index(gi fonts.GlyphIndex) (int, bool) {
	num := len(cr)
	if num == 0 {
		return 0, false
	}

	idx := sort.Search(num, func(i int) bool { return gi <= cr[i].Start })
	// idx either points to a matching start, or to the next range (or idx==num)
	// e.g. with the range example from above: 130 points to 130-135 range, 133 points to 137-137 range

	// check if gi is the start of a range, but only if sort.Search returned a valid result
	if idx < num {
		if rang := cr[idx]; gi == rang.Start {
			return int(rang.StartCoverage), true
		}
	}
	// check if gi is in previous range
	if idx > 0 {
		idx--
		if rang := cr[idx]; gi >= rang.Start && gi <= rang.End {
			return rang.StartCoverage + int(gi-rang.Start), true
		}
	}

	return 0, false
}

func (cr CoverageRanges) Size() int {
	size := 0
	for _, r := range cr {
		size += int(r.End - r.Start + 1)
	}
	return size
}

// func (cr coverageRanges) maxIndex() int {
// 	lastRange := cr[len(cr)-1]
// 	return lastRange.startCoverage + int(lastRange.end-lastRange.start)
// }

func fetchCoverageRange(buf []byte) (CoverageRanges, error) {
	const headerSize, entrySize = 2, 6
	if len(buf) < headerSize {
		return nil, errInvalidGPOSKern
	}

	num := int(be.Uint16(buf))
	if len(buf) < headerSize+num*entrySize {
		return nil, errInvalidGPOSKern
	}

	out := make(CoverageRanges, num)
	for i := range out {
		out[i].Start = fonts.GlyphIndex(be.Uint16(buf[headerSize+i*entrySize:]))
		out[i].End = fonts.GlyphIndex(be.Uint16(buf[headerSize+i*entrySize+2:]))
		out[i].StartCoverage = int(be.Uint16(buf[headerSize+i*entrySize+4:]))
	}
	return out, nil
}

// offset int
func parsePairPosFormat1(buf []byte, coverage Coverage) (pairPosKern, error) {
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
	right fonts.GlyphIndex
	kern  int16
}

// slice indexed by tableIndex
type pairPosKern struct {
	cov  Coverage
	list [][]pairKern
}

func (pp pairPosKern) KernPair(a, b fonts.GlyphIndex) (int16, bool) {
	idx, found := pp.cov.Index(a)
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

func fetchPairPosGlyph(coverage Coverage, num int, glyphs []byte) (pairPosKern, error) {
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
				right: fonts.GlyphIndex(be.Uint16(glyphs[offset+2+i*4:])),
				kern:  int16(be.Uint16(glyphs[offset+2+i*4+2:])),
			}
		}
		lists[idx] = list
	}
	return pairPosKern{cov: coverage, list: lists}, nil
}

type classKerns struct {
	coverage       Coverage
	class1, class2 class
	numClass2      uint16
	kerns          []int16 // size numClass1 * numClass2
}

func (c classKerns) KernPair(left, right fonts.GlyphIndex) (int16, bool) {
	// check coverage to avoid selection of default class 0
	_, found := c.coverage.Index(left)
	if !found {
		return 0, false
	}
	idxa := c.class1.ClassID(left)
	idxb := c.class2.ClassID(right)
	return c.kerns[idxb+idxa*c.numClass2], true
}

func (c classKerns) Size() int { return c.class1.size() * c.class2.size() }

func parsePairPosFormat2(buf []byte, coverage Coverage) (classKerns, error) {
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

	cdef1Offset := be.Uint16(buf[8:])
	cdef2Offset := be.Uint16(buf[10:])
	numClass1 := be.Uint16(buf[12:])
	numClass2 := be.Uint16(buf[14:])
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

func fetchPairPosClass(buf []byte, cov Coverage, num1, num2 uint16, cdef1, cdef2 class) (classKerns, error) {
	if len(buf) < int(num1)*int(num2)*2 {
		return classKerns{}, errInvalidGPOSKern
	}

	kerns := make([]int16, int(num1)*int(num2))
	for i := 0; i < int(num1); i++ {
		for j := 0; j < int(num2); j++ {
			index := j + i*int(num2)
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
