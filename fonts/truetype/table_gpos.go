package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"github.com/benoitkugler/textlayout/fonts"
)

var errInvalidGPOSKern = errors.New("invalid GPOS kerning subtable")

// TableGPOS provides precise control over glyph placement
// for sophisticated text layout and rendering in each script
// and language system that a font supports.
type TableGPOS struct {
	TableLayout
	Lookups []LookupGPOS
}

func parseTableGPOS(data []byte) (*TableGPOS, error) {
	tableLayout, lookups, err := parseTableLayout(data)
	if err != nil {
		return nil, err
	}
	out := &TableGPOS{
		TableLayout: tableLayout,
		Lookups:     make([]LookupGPOS, len(lookups)),
	}
	for i, l := range lookups {
		out.Lookups[i], err = l.parseGPOS(uint16(len(lookups)))
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// sum up the kerning information from the lookups.
// Note that this is an over simplification, since we fetch kerning for all language/scripts
func (t *TableGPOS) horizontalKerning() (Kerns, error) {
	var kerns kernUnions
	for _, lookup := range t.Lookups {
		if lookup.Type != GPOSPair {
			continue
		}
		for _, subtable := range lookup.Subtables {
			switch data := subtable.Data.(type) {
			case GPOSPair1:
				// we only support kerning with X_ADVANCE for first glyph
				if data.FormatFirst&XAdvance == 0 || data.FormatSecond != 0 {
					continue
				}
				out := pairPosKern{cov: subtable.Coverage, list: make([][]pairKern, len(data.Values))}
				for i, v := range data.Values {
					vi := make([]pairKern, len(v))
					for j, k := range v {
						vi[j].right = k.SecondGlyph
						vi[j].kern = k.First.XAdvance
					}
					out.list[i] = vi
				}
				kerns = append(kerns, out)

			case GPOSPair2:
				// we only support kerning with X_ADVANCE for first glyph
				if data.FormatFirst&XAdvance == 0 || data.FormatSecond != 0 {
					continue
				}

				out := classKerns{
					coverage: subtable.Coverage,
					class1:   data.First, class2: data.Second,
					kerns: make([][]int16, len(data.Values)),
				}
				for i, vs := range data.Values {
					vi := make([]int16, len(vs))
					for j, v := range vs {
						vi[j] = v[0].XAdvance
					}
					out.kerns[i] = vi
				}
				kerns = append(kerns, out)
			}
		}
	}

	if len(kerns) == 0 {
		// no kerning information
		return nil, errors.New("missing GPOS kerning information")
	}

	return kerns, nil
}

// GPOSType identifies the kind of lookup format, for GPOS tables.
type GPOSType uint16

const (
	GPOSSingle         GPOSType = 1 + iota // Adjust position of a single glyph
	GPOSPair                               // Adjust position of a pair of glyphs
	GPOSCursive                            // Attach cursive glyphs
	GPOSMarkToBase                         // Attach a combining mark to a base glyph
	GPOSMarkToLigature                     // Attach a combining mark to a ligature
	GPOSMarkToMark                         // Attach a combining mark to another mark
	GPOSContext                            // Position one or more glyphs in context
	GPOSChained                            // Position one or more glyphs in chained context
	gposExtension                          // Extension mechanism for other positionings
)

// GPOSSubtable is one of the subtables of a
// GPOS lookup.
type GPOSSubtable struct {
	// For GPOSChained - Format 3, its the coverage of the first input.
	Coverage Coverage
	Data     interface{ Type() GPOSType }
}

// LookupGPOS is a lookup for GPOS tables.
type LookupGPOS struct {
	Type GPOSType
	LookupOptions
	// After successful parsing, it is a non empty array
	// with all subtables of the same `GPOSType`.
	Subtables []GPOSSubtable
}

// interpret the lookup as a GPOS lookup
// lookupLength is used to sanitize nested lookups
func (header lookup) parseGPOS(lookupListLength uint16) (out LookupGPOS, err error) {
	out.Type = GPOSType(header.kind)
	out.LookupOptions = header.LookupOptions

	out.Subtables = make([]GPOSSubtable, len(header.subtableOffsets))
	for i, offset := range header.subtableOffsets {
		out.Subtables[i], err = parseGPOSSubtable(header.data, int(offset), out.Type, lookupListLength)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

func parseGPOSSubtable(data []byte, offset int, kind GPOSType, lookupListLength uint16) (out GPOSSubtable, err error) {
	// read the format and coverage
	if offset+4 >= len(data) {
		return out, fmt.Errorf("invalid lookup subtable offset %d", offset)
	}
	format := binary.BigEndian.Uint16(data[offset:])

	// almost all table have a coverage offset, right after the format; special case the others
	// see below for the coverage
	if kind == gposExtension || (kind == GPOSChained || kind == GPOSContext) && format == 3 {
		out.Coverage = CoverageList{}
	} else {
		covOffset := binary.BigEndian.Uint16(data[offset+2:]) // relative to the subtable
		out.Coverage, err = parseCoverage(data[offset:], covOffset)
		if err != nil {
			return out, fmt.Errorf("invalid GPOS table (format %d-%d): %s", kind, format, err)
		}
	}

	// read the actual lookup
	switch kind {
	case GPOSSingle:
		out.Data, err = parseGPOSSingle(format, data[offset:], out.Coverage)
	case GPOSPair:
		out.Data, err = parseGPOSPair(format, data[offset:], out.Coverage)
	case GPOSCursive:
		out.Data, err = parseGPOSCursive(data[offset:], out.Coverage)
	case GPOSMarkToBase:
		out.Data, err = parseGPOSMarkToBase(data[offset:], out.Coverage)
	case GPOSMarkToLigature:
		out.Data, err = parseGPOSMarkToLigature(data[offset:], out.Coverage)
	case GPOSMarkToMark:
		out.Data, err = parseGPOSMarkToMark(data[offset:], out.Coverage)
	case GPOSContext:
		out.Data, err = parseGPOSContext(format, data[offset:], lookupListLength, &out.Coverage)
	case GPOSChained:
		out.Data, err = parseGPOSChained(format, data[offset:], lookupListLength, &out.Coverage)
	case gposExtension:
		out, err = parseGPOSExtension(data[offset:], lookupListLength)
	default:
		return out, fmt.Errorf("unsupported gsub lookup type %d", kind)
	}
	return out, err
}

type GPOSSingle1 struct {
	Format GPOSValueFormat
	Value  GPOSValueRecord
}

type GPOSSingle2 struct {
	Format GPOSValueFormat
	Values []GPOSValueRecord
}

func (GPOSSingle1) Type() GPOSType { return GPOSSingle }
func (GPOSSingle2) Type() GPOSType { return GPOSSingle }

func parseGPOSSingle(format uint16, data []byte, cov Coverage) (interface{ Type() GPOSType }, error) {
	switch format {
	case 1:
		return parseGPOSSingleFormat1(data)
	case 2:
		return parseGPOSSingleFormat2(data, cov)
	default:
		return nil, fmt.Errorf("unsupported single positionning format: %d", format)
	}
}

func parseGPOSSingleFormat1(data []byte) (GPOSSingle1, error) {
	if len(data) < 6 {
		return GPOSSingle1{}, errors.New("invalid single positionning subtable format 1 (EOF)")
	}
	valueFormat := GPOSValueFormat(binary.BigEndian.Uint16(data[4:]))
	v, _, err := parseGPOSValueRecord(valueFormat, data, 6)
	if err != nil {
		return GPOSSingle1{}, fmt.Errorf("invalid single positionning subtable format 1: %s", err)
	}
	return GPOSSingle1{Format: valueFormat, Value: v}, nil
}

// cov is used to sanitize
func parseGPOSSingleFormat2(data []byte, cov Coverage) (out GPOSSingle2, err error) {
	if len(data) < 8 {
		return out, errors.New("invalid single positionning subtable format 2 (EOF)")
	}
	out.Format = GPOSValueFormat(binary.BigEndian.Uint16(data[4:]))
	count := binary.BigEndian.Uint16(data[6:])

	if cov.Size() != int(count) {
		return out, errors.New("invalid single positionning subtable format 2 (EOF)")
	}

	offset := 8
	out.Values = make([]GPOSValueRecord, count)
	for i := range out.Values {
		out.Values[i], offset, err = parseGPOSValueRecord(out.Format, data, offset)
		if err != nil {
			return out, fmt.Errorf("invalid single positionning subtable format 2: %s", err)
		}
	}
	return out, nil
}

type GPOSPairValueRecord struct {
	SecondGlyph   fonts.GlyphIndex // Glyph ID of second glyph in the pair
	First, Second GPOSValueRecord  // Positioning data for both glyphs
}

type GPOSPair1 struct {
	FormatFirst, FormatSecond GPOSValueFormat
	Values                    [][]GPOSPairValueRecord // one set for each glyph in the coverage
}

type GPOSPair2 struct {
	First, Second             Class
	FormatFirst, FormatSecond GPOSValueFormat
	// Positionning for first and second glyphs, with size First.Extent() x Second.Extent()
	Values [][][2]GPOSValueRecord
}

func (GPOSPair1) Type() GPOSType { return GPOSPair }
func (GPOSPair2) Type() GPOSType { return GPOSPair }

func parseGPOSPair(format uint16, data []byte, cov Coverage) (interface{ Type() GPOSType }, error) {
	switch format {
	case 1:
		return parseGPOSPairFormat1(data, cov)
	case 2:
		return parseGPOSPairFormat2(data, cov)
	default:
		return nil, fmt.Errorf("unsupported pair positionning format: %d", format)
	}
}

func parseGPOSPairFormat1(buf []byte, coverage Coverage) (out GPOSPair1, err error) {
	const headerSize = 10 // including posFormat and coverageOffset
	if len(buf) < headerSize {
		return out, errors.New("invalid pair positionning subtable format 1 (EOF)")
	}
	out.FormatFirst = GPOSValueFormat(binary.BigEndian.Uint16(buf[4:]))
	out.FormatSecond = GPOSValueFormat(binary.BigEndian.Uint16(buf[6:]))
	pairSetCount := int(binary.BigEndian.Uint16(buf[8:]))

	if coverage.Size() != pairSetCount {
		return out, errors.New("invalid pair positionning subtable format 1")
	}

	offsets, err := parseUint16s(buf[10:], pairSetCount)
	if err != nil {
		return out, fmt.Errorf("invalid pair positionning subtable format 1: %s", err)
	}
	out.Values = make([][]GPOSPairValueRecord, len(offsets))
	for i, offset := range offsets {
		out.Values[i], err = parsePositionPairValueRecordSet(buf, offset, out.FormatFirst, out.FormatSecond)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

// cov is used to sanitize
func parseGPOSPairFormat2(buf []byte, cov Coverage) (out GPOSPair2, err error) {
	const headerSize = 16 // including posFormat and coverageOffset
	if len(buf) < headerSize {
		return out, errors.New("invalid pair positionning subtable format 2 (EOF)")
	}

	out.FormatFirst = GPOSValueFormat(binary.BigEndian.Uint16(buf[4:]))
	out.FormatSecond = GPOSValueFormat(binary.BigEndian.Uint16(buf[6:]))

	cdef1Offset := be.Uint16(buf[8:])
	cdef2Offset := be.Uint16(buf[10:])
	class1Count := int(be.Uint16(buf[12:]))
	class2Count := int(be.Uint16(buf[14:]))

	out.First, err = parseClass(buf, cdef1Offset)
	if err != nil {
		return out, err
	}
	out.Second, err = parseClass(buf, cdef2Offset)
	if err != nil {
		return out, err
	}

	if out.First.Extent() != class1Count {
		return out, errors.New("invalid pair positionning subtable format 2")
	}
	if out.Second.Extent() != class2Count {
		return out, errors.New("invalid pair positionning subtable format 2")
	}

	out.Values = make([][][2]GPOSValueRecord, class1Count)
	offset := headerSize
	for i := range out.Values {
		vi := make([][2]GPOSValueRecord, class2Count)
		for j := range vi {
			vi[j][0], offset, err = parseGPOSValueRecord(out.FormatFirst, buf, offset)
			if err != nil {
				return out, fmt.Errorf("invalid pair positionning subtable format 2: %s", err)
			}
			vi[j][1], offset, err = parseGPOSValueRecord(out.FormatSecond, buf, offset)
			if err != nil {
				return out, fmt.Errorf("invalid pair positionning subtable format 2: %s", err)
			}
		}
		out.Values[i] = vi
	}

	return out, nil
}

func parsePositionPairValueRecordSet(data []byte, offset uint16, fmt1, fmt2 GPOSValueFormat) ([]GPOSPairValueRecord, error) {
	if len(data) < 2+int(offset) {
		return nil, errors.New("invalid pair set table (EOF)")
	}
	data = data[offset:]
	count := binary.BigEndian.Uint16(data)
	out := make([]GPOSPairValueRecord, count)
	offsetR := 2
	var err error
	for i := range out {
		if len(data) < 2+offsetR {
			return nil, errors.New("invalid pair set table (EOF)")
		}
		out[i].SecondGlyph = fonts.GlyphIndex(binary.BigEndian.Uint16(data[offsetR:]))
		out[i].First, offsetR, err = parseGPOSValueRecord(fmt1, data, offsetR+2)
		if err != nil {
			return nil, fmt.Errorf("invalid pair set table: %s", err)
		}
		out[i].Second, offsetR, err = parseGPOSValueRecord(fmt2, data, offsetR)
		if err != nil {
			return nil, fmt.Errorf("invalid pair set table: %s", err)
		}
	}
	return out, nil
}

type GPOSCursive1 [][2]GPOSAnchor // entry, exit (may be null)

func (GPOSCursive1) Type() GPOSType { return GPOSCursive }

func parseGPOSCursive(data []byte, cov Coverage) (GPOSCursive1, error) {
	if len(data) < 6 {
		return nil, errors.New("invalid cursive positionning subtable (EOF)")
	}
	count := binary.BigEndian.Uint16(data[4:])
	if len(data) < 6+4-int(count) {
		return nil, errors.New("invalid cursive positionning subtable (EOF)")
	}
	out := make(GPOSCursive1, count)
	var err error
	for i := range out {
		entryOffset := binary.BigEndian.Uint16(data[6+4*i:])  // may be null
		exitOffset := binary.BigEndian.Uint16(data[6+4*i+2:]) // may be null
		if entryOffset != 0 {
			out[i][0], err = parseGPOSAnchor(data, entryOffset)
			if err != nil {
				return nil, err
			}
		}
		if exitOffset != 0 {
			out[i][1], err = parseGPOSAnchor(data, entryOffset)
			if err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

type GPOSMarkToBase1 struct {
	BaseCoverage Coverage
	Marks        []GPOSMark
	Bases        [][]GPOSAnchor // one set for each index in `BaseCoverage`, each with same length
}

func (GPOSMarkToBase1) Type() GPOSType { return GPOSMarkToBase }

func parseGPOSMarkToBase(data []byte, markCov Coverage) (out GPOSMarkToBase1, err error) {
	if len(data) < 12 {
		return out, errors.New("invalid mark-to-base positionning subtable (EOF)")
	}
	baseCovOffset := binary.BigEndian.Uint16(data[4:])
	markClassCount := int(binary.BigEndian.Uint16(data[6:]))
	markArrayOffset := binary.BigEndian.Uint16(data[8:])
	baseArrayOffset := int(binary.BigEndian.Uint16(data[10:]))

	out.BaseCoverage, err = parseCoverage(data, baseCovOffset)
	if err != nil {
		return out, fmt.Errorf("invalid mark-to-base positionning subtable: %s", err)
	}

	out.Marks, err = parseGPOSMarkArray(data, markArrayOffset)
	if err != nil {
		return out, fmt.Errorf("invalid mark-to-base positionning subtable: %s", err)
	}

	if markCov.Size() != len(out.Marks) {
		return out, errors.New("invalid mark-to-base positionning subtable")
	}

	if len(data) < baseArrayOffset+2 {
		return out, errors.New("invalid mark-to-base positionning subtable (EOF)")
	}
	data = data[baseArrayOffset:]
	baseCount := int(binary.BigEndian.Uint16(data))
	if len(data) < 2*baseCount*markClassCount {
		return out, errors.New("invalid mark-to-base positionning subtable (EOF)")
	}
	if out.BaseCoverage.Size() != baseCount {
		return out, errors.New("invalid mark-to-base positionning subtable (EOF)")
	}
	out.Bases = make([][]GPOSAnchor, baseCount)
	for i := range out.Bases {
		vi := make([]GPOSAnchor, markClassCount)
		for j := range vi {
			anchorOffset := binary.BigEndian.Uint16(data[2+(i*markClassCount+j)*2:])
			if anchorOffset == 0 {
				continue
			}
			vi[j], err = parseGPOSAnchor(data, anchorOffset)
			if err != nil {
				return out, err
			}
		}
		out.Bases[i] = vi
	}

	return out, nil
}

type GPOSMarkToLigature1 struct {
	LigatureCoverage Coverage
	Marks            []GPOSMark
	Ligatures        [][][]GPOSAnchor // one set for each index in `LigatureCoverage`
}

func (GPOSMarkToLigature1) Type() GPOSType { return GPOSMarkToLigature }

func parseGPOSMarkToLigature(data []byte, markCov Coverage) (out GPOSMarkToLigature1, err error) {
	if len(data) < 12 {
		return out, errors.New("invalid mark-to-ligature positionning subtable (EOF)")
	}
	ligCovOffset := binary.BigEndian.Uint16(data[4:])
	markClassCount := int(binary.BigEndian.Uint16(data[6:]))
	markArrayOffset := binary.BigEndian.Uint16(data[8:])
	ligArrayOffset := int(binary.BigEndian.Uint16(data[10:]))

	out.LigatureCoverage, err = parseCoverage(data, ligCovOffset)
	if err != nil {
		return out, fmt.Errorf("invalid mark-to-ligature positionning subtable: %s", err)
	}

	out.Marks, err = parseGPOSMarkArray(data, markArrayOffset)
	if err != nil {
		return out, fmt.Errorf("invalid mark-to-ligature positionning subtable: %s", err)
	}

	if markCov.Size() != len(out.Marks) {
		return out, errors.New("invalid mark-to-ligature positionning subtable")
	}

	if len(data) < ligArrayOffset+2 {
		return out, errors.New("invalid mark-to-ligature positionning subtable (EOF)")
	}
	data = data[ligArrayOffset:]
	ligatureCount := int(binary.BigEndian.Uint16(data))
	if len(data) < 2*ligatureCount {
		return out, errors.New("invalid mark-to-ligature positionning subtable (EOF)")
	}
	if out.LigatureCoverage.Size() != ligatureCount {
		return out, errors.New("invalid mark-to-ligature positionning subtable (EOF)")
	}
	out.Ligatures = make([][][]GPOSAnchor, ligatureCount)
	for i := range out.Ligatures {
		ligatureAttachOffset := binary.BigEndian.Uint16(data[2+i*2:])
		if len(data) < int(ligatureAttachOffset)+2 {
			return out, errors.New("invalid mark-to-ligature positionning subtable (EOF)")
		}
		ligatureAttachData := data[ligatureAttachOffset:]
		componentCount := binary.BigEndian.Uint16(ligatureAttachData)
		if len(ligatureAttachData) < 2+int(componentCount)*2*markClassCount {
			return out, errors.New("invalid mark-to-ligature positionning subtable (EOF)")
		}
		vi := make([][]GPOSAnchor, componentCount)
		for j := range vi {
			vij := make([]GPOSAnchor, markClassCount)
			for k := range vij {
				anchorOffset := binary.BigEndian.Uint16(ligatureAttachData[2+(j*markClassCount+k)*2:])
				if anchorOffset == 0 {
					continue
				}
				vij[k], err = parseGPOSAnchor(ligatureAttachData, anchorOffset)
				if err != nil {
					return out, err
				}
			}
			vi[j] = vij
		}
		out.Ligatures[i] = vi
	}

	return out, nil
}

type GPOSMarkToMark1 struct {
	Mark2Coverage Coverage
	Marks1        []GPOSMark
	Marks2        [][]GPOSAnchor // one set for each index in `Mark2Coverage`, each with same length
}

func (GPOSMarkToMark1) Type() GPOSType { return GPOSMarkToMark }

func parseGPOSMarkToMark(data []byte, mark1Cov Coverage) (GPOSMarkToMark1, error) {
	// same structure as mark-to-base
	out, err := parseGPOSMarkToBase(data, mark1Cov)
	return GPOSMarkToMark1{Mark2Coverage: out.BaseCoverage, Marks1: out.Marks, Marks2: out.Bases}, err
}

type (
	GPOSContext1 [][]SequenceRule
	GPOSContext2 SequenceContext2
	GPOSContext3 SequenceContext3
)

func (GPOSContext1) Type() GPOSType { return GPOSContext }
func (GPOSContext2) Type() GPOSType { return GPOSContext }
func (GPOSContext3) Type() GPOSType { return GPOSContext }

// lookupLength is used to sanitize lookup indexes.
// cov is used for ContextFormat3
func parseGPOSContext(format uint16, data []byte, lookupLength uint16, cov *Coverage) (interface{ Type() GPOSType }, error) {
	switch format {
	case 1:
		out, err := parseSequenceContext1(data, lookupLength)
		return GPOSContext1(out), err
	case 2:
		out, err := parseSequenceContext2(data, lookupLength)
		return GPOSContext2(out), err
	case 3:
		out, err := parseSequenceContext3(data, lookupLength)
		if len(out.Coverages) != 0 {
			*cov = out.Coverages[0]
		}
		return GPOSContext3(out), err
	default:
		return nil, fmt.Errorf("unsupported sequence context format %d", format)
	}
}

type (
	GPOSChainedContext1 [][]ChainedSequenceRule
	GPOSChainedContext2 ChainedSequenceContext2
	GPOSChainedContext3 ChainedSequenceContext3
)

func (GPOSChainedContext1) Type() GPOSType { return GPOSChained }
func (GPOSChainedContext2) Type() GPOSType { return GPOSChained }
func (GPOSChainedContext3) Type() GPOSType { return GPOSChained }

// lookupLength is used to sanitize lookup indexes.
// cov is used for ContextFormat3
func parseGPOSChained(format uint16, data []byte, lookupLength uint16, cov *Coverage) (interface{ Type() GPOSType }, error) {
	switch format {
	case 1:
		out, err := parseChainedSequenceContext1(data, lookupLength)
		return GPOSChainedContext1(out), err
	case 2:
		out, err := parseChainedSequenceContext2(data, lookupLength)
		return GPOSChainedContext2(out), err
	case 3:
		out, err := parseChainedSequenceContext3(data, lookupLength)
		if len(out.Input) != 0 {
			*cov = out.Input[0]
		}
		return GPOSChainedContext3(out), err
	default:
		return nil, fmt.Errorf("unsupported sequence context format %d", format)
	}
}

// returns the extension subtable instead
func parseGPOSExtension(data []byte, lookupListLength uint16) (GPOSSubtable, error) {
	if len(data) < 8 {
		return GPOSSubtable{}, errors.New("invalid extension positionning table")
	}
	extensionType := GPOSType(binary.BigEndian.Uint16(data[2:]))
	offset := binary.BigEndian.Uint32(data[4:])

	if extensionType == gposExtension {
		return GPOSSubtable{}, errors.New("invalid extension positionning table")
	}

	return parseGPOSSubtable(data, int(offset), extensionType, lookupListLength)
}

//
// ---------------- Simplified API for horizontal kerning ----------------
//

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

type classKerns struct {
	coverage       Coverage
	class1, class2 Class
	kerns          [][]int16 // size numClass1 * numClass2
}

func (c classKerns) KernPair(left, right fonts.GlyphIndex) (int16, bool) {
	// check coverage to avoid selection of default class 0
	_, found := c.coverage.Index(left)
	if !found {
		return 0, false
	}
	idxa, _ := c.class1.ClassID(left)
	idxb, _ := c.class2.ClassID(right)
	return c.kerns[idxa][idxb], true
}

func (c classKerns) Size() int { return c.class1.GlyphSize() * c.class2.GlyphSize() }

//
// ---------------------------- shared format ----------------------------
//

// GPOSValueFormat is a mask indicating which field
// are set in a GPOSValueRecord.
// It is often shared between many records.
type GPOSValueFormat uint16

// number of fields present
func (f GPOSValueFormat) size() int { return bits.OnesCount16(uint16(f)) }

const (
	XPlacement GPOSValueFormat = 1 << iota /* Includes horizontal adjustment for placement */
	YPlacement                             /* Includes vertical adjustment for placement */
	XAdvance                               /* Includes horizontal adjustment for advance */
	YAdvance                               /* Includes vertical adjustment for advance */
	XPlaDevice                             /* Includes horizontal Device table for placement */
	YPlaDevice                             /* Includes vertical Device table for placement */
	XAdvDevice                             /* Includes horizontal Device table for advance */
	YAdvDevice                             /* Includes vertical Device table for advance */
	// ignored                                /* Was used in TrueType Open for MM fonts */
	// reserved                               /* For future use */

	//  Mask for having any Device table
	Devices = XPlaDevice | YPlaDevice | XAdvDevice | YAdvDevice
)

type GPOSValueRecord struct {
	// format     gposValueFormat
	XPlacement int16      // Horizontal adjustment for placement--in design units
	YPlacement int16      // Vertical adjustment for placement--in design units
	XAdvance   int16      // Horizontal adjustment for advance--in design units (only used for horizontal writing)
	YAdvance   int16      // Vertical adjustment for advance--in design units (only used for vertical writing)
	XPlaDevice GPOSDevice // Offset to Device table for horizontal placement (may be nil)
	YPlaDevice GPOSDevice // Offset to Device table for vertical placement (may be nil)
	XAdvDevice GPOSDevice // Offset to Device table for horizontal advance (may be nil)
	YAdvDevice GPOSDevice // Offset to Device table for vertical advance (may be nil)
}

// data starts at the immediate parent table. return the shifted offset
func parseGPOSValueRecord(format GPOSValueFormat, data []byte, offset int) (out GPOSValueRecord, _ int, err error) {
	if len(data) < offset {
		return out, 0, errors.New("invalid value record (EOF)")
	}

	size := format.size() // number of fields present
	if size == 0 {        // return early
		return out, offset, nil
	}
	// start by parsing the list of values
	values, err := parseUint16s(data[offset:], size)
	if err != nil {
		return out, 0, fmt.Errorf("invalid value record: %s", err)
	}
	// follow the order
	if format&XPlacement != 0 {
		out.XPlacement = int16(values[0])
		values = values[1:]
	}
	if format&YPlacement != 0 {
		out.YPlacement = int16(values[0])
		values = values[1:]
	}
	if format&XAdvance != 0 {
		out.XAdvance = int16(values[0])
		values = values[1:]
	}
	if format&YAdvance != 0 {
		out.YAdvance = int16(values[0])
		values = values[1:]
	}
	if format&XPlaDevice != 0 {
		if devOffset := values[0]; devOffset != 0 {
			out.XPlaDevice, err = parseGPOSDevice(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		values = values[1:]
	}
	if format&YPlaDevice != 0 {
		if devOffset := values[0]; devOffset != 0 {
			out.YPlaDevice, err = parseGPOSDevice(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		values = values[1:]
	}
	if format&XAdvDevice != 0 {
		if devOffset := values[0]; devOffset != 0 {
			out.XAdvDevice, err = parseGPOSDevice(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		values = values[1:]
	}
	if format&YAdvDevice != 0 {
		if devOffset := values[0]; devOffset != 0 {
			out.YAdvDevice, err = parseGPOSDevice(data, devOffset)
			if err != nil {
				return out, 0, err
			}
		}
		values = values[1:]
	}
	return out, offset + 2*size, err
}

type GPOSAnchor interface {
	isAnchor()
}

func (GPOSAnchorFormat1) isAnchor() {}
func (GPOSAnchorFormat2) isAnchor() {}
func (GPOSAnchorFormat3) isAnchor() {}

func parseGPOSAnchor(data []byte, offset uint16) (GPOSAnchor, error) {
	if len(data) < 2+int(offset) {
		return nil, errors.New("invalid anchor table (EOF)")
	}
	switch format := binary.BigEndian.Uint16(data[offset:]); format {
	case 1:
		return parseGPOSAnchorFormat1(data)
	case 2:
		return parseGPOSAnchorFormat2(data)
	case 3:
		return parseGPOSAnchorFormat3(data)
	default:
		return nil, fmt.Errorf("unsupported anchor subtable format: %d", format)
	}
}

type GPOSAnchorFormat1 struct {
	X, Y int16 // in design units
}

// data starts at format
func parseGPOSAnchorFormat1(data []byte) (out GPOSAnchorFormat1, err error) {
	if len(data) < 6 {
		return out, errors.New("invalid anchor table format 1 (EOF)")
	}
	out.X = int16(binary.BigEndian.Uint16(data[2:]))
	out.Y = int16(binary.BigEndian.Uint16(data[4:]))
	return out, err
}

type GPOSAnchorFormat2 struct {
	GPOSAnchorFormat1
	AnchorPoint fonts.GlyphIndex
}

// data starts at format
func parseGPOSAnchorFormat2(data []byte) (out GPOSAnchorFormat2, err error) {
	if len(data) < 8 {
		return out, errors.New("invalid anchor table format 2 (EOF)")
	}
	out.X = int16(binary.BigEndian.Uint16(data[2:]))
	out.Y = int16(binary.BigEndian.Uint16(data[4:]))
	out.AnchorPoint = fonts.GlyphIndex(binary.BigEndian.Uint16(data[6:]))
	return out, err
}

type GPOSAnchorFormat3 struct {
	GPOSAnchorFormat1
	xDeviceOffset, yDeviceOffset uint16
}

// data starts at format
func parseGPOSAnchorFormat3(data []byte) (out GPOSAnchorFormat3, err error) {
	if len(data) < 10 {
		return out, errors.New("invalid anchor table format 3 (EOF)")
	}
	out.X = int16(binary.BigEndian.Uint16(data[2:]))
	out.Y = int16(binary.BigEndian.Uint16(data[4:]))
	out.xDeviceOffset = binary.BigEndian.Uint16(data[6:])
	out.yDeviceOffset = binary.BigEndian.Uint16(data[8:])
	return out, err
}

type GPOSMark struct {
	ClassValue uint16
	Anchor     GPOSAnchor
}

func parseGPOSMarkArray(data []byte, offset uint16) ([]GPOSMark, error) {
	if len(data) < 2+int(offset) {
		return nil, errors.New("invalid positionning mark array (EOF)")
	}
	data = data[offset:]
	count := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+4*count {
		return nil, errors.New("invalid positionning mark array (EOF)")
	}
	out := make([]GPOSMark, count)
	var err error
	for i := range out {
		out[i].ClassValue = binary.BigEndian.Uint16(data[2+4*i:])
		anchorOffset := binary.BigEndian.Uint16(data[2+4*i+2:])
		out[i].Anchor, err = parseGPOSAnchor(data, anchorOffset)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// GPOSDevice is either an GPOSDeviceHinting for standard fonts,
// or a GPOSDeviceVariation for variable fonts.
type GPOSDevice interface {
	isDevice()
}

func (GPOSDeviceHinting) isDevice()   {}
func (GPOSDeviceVariation) isDevice() {}

type GPOSDeviceHinting struct {
	StartSize, EndSize uint16 // correction range, in ppem
	Values             []int8 // with length endSize - startSize + 1
}

type GPOSDeviceVariation struct {
	DeltaSetOuter, DeltaSetInner uint16 // index into the item variation store
}

func parseGPOSDevice(data []byte, offset uint16) (GPOSDevice, error) {
	if len(data) < int(offset)+6 {
		return nil, errors.New("invalid positionning device subtable (EOF)")
	}
	first := binary.BigEndian.Uint16(data[offset:])
	second := binary.BigEndian.Uint16(data[offset+2:])
	format := binary.BigEndian.Uint16(data[offset+4:])

	switch format {
	case 1, 2, 3:
		var out GPOSDeviceHinting

		out.StartSize, out.EndSize = first, second
		if out.EndSize < out.StartSize {
			return nil, errors.New("invalid positionning device subtable")
		}

		nbPerUint16 := 16 / (1 << format) // 8, 4 or 2
		outLength := int(out.EndSize - out.StartSize + 1)
		count := outLength / nbPerUint16
		uint16s, err := parseUint16s(data[offset+6:], count)
		if err != nil {
			return nil, err
		}
		out.Values = make([]int8, count*nbPerUint16) // handle rounding error by reslicing after
		switch format {
		case 1:
			for i, u := range uint16s {
				uint16As2Bits(out.Values[i*8:], u)
			}
		case 2:
			for i, u := range uint16s {
				uint16As4Bits(out.Values[i*4:], u)
			}
		case 3:
			for i, u := range uint16s {
				uint16As8Bits(out.Values[i*2:], u)
			}
		}
		out.Values = out.Values[:outLength]
		return out, nil
	case 0x8000:
		return GPOSDeviceVariation{DeltaSetOuter: first, DeltaSetInner: second}, nil
	default:
		return nil, fmt.Errorf("unsupported positionning device subtable: %d", format)
	}
}

// write 8 elements
func uint16As2Bits(dst []int8, u uint16) {
	const mask = 0xFE // 11111110
	dst[0] = int8((0-uint8(u>>15&1))&mask | uint8(u>>14&1))
	dst[1] = int8((0-uint8(u>>13&1))&mask | uint8(u>>12&1))
	dst[2] = int8((0-uint8(u>>11&1))&mask | uint8(u>>10&1))
	dst[3] = int8((0-uint8(u>>9&1))&mask | uint8(u>>8&1))
	dst[4] = int8((0-uint8(u>>7&1))&mask | uint8(u>>6&1))
	dst[5] = int8((0-uint8(u>>5&1))&mask | uint8(u>>4&1))
	dst[6] = int8((0-uint8(u>>3&1))&mask | uint8(u>>2&1))
	dst[7] = int8((0-uint8(u>>1&1))&mask | uint8(u>>0&1))
}

// write 4 elements
func uint16As4Bits(dst []int8, u uint16) {
	const mask = 0xF8 // 11111000

	dst[0] = int8((0-uint8(u>>15&1))&mask | uint8(u>>12&0x07))
	dst[1] = int8((0-uint8(u>>11&1))&mask | uint8(u>>8&0x07))
	dst[2] = int8((0-uint8(u>>7&1))&mask | uint8(u>>4&0x07))
	dst[3] = int8((0-uint8(u>>3&1))&mask | uint8(u>>0&0x07))
}

// write 2 elements
func uint16As8Bits(dst []int8, u uint16) {
	dst[0] = int8(u >> 8)
	dst[1] = int8(u)
}
