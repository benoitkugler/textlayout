package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

// TableGSUB is the Glyph Substitution (GSUB) table.
// It provides data for substition of glyphs for appropriate rendering of scripts,
// such as cursively-connecting forms in Arabic script,
// or for advanced typographic effects, such as ligatures.
type TableGSUB struct {
	TableLayout
	Lookups []LookupGSUB
}

func parseTableGSUB(data []byte) (*TableGSUB, error) {
	tableLayout, lookups, err := parseTableLayout(data)
	if err != nil {
		return nil, err
	}
	out := TableGSUB{
		TableLayout: tableLayout,
		Lookups:     make([]LookupGSUB, len(lookups)),
	}
	for i, l := range lookups {
		out.Lookups[i], err = l.parseGSUB(uint16(len(lookups)))
		if err != nil {
			return nil, err
		}
	}
	return &out, nil
}

// GSUBType identifies the kind of lookup format, for GSUB tables.
type GSUBType uint16

const (
	SubSingle    GSUBType = 1 + iota // Single (format 1.1 1.2)	Replace one glyph with one glyph
	SubMultiple                      // Multiple (format 2.1)	Replace one glyph with more than one glyph
	SubAlternate                     // Alternate (format 3.1)	Replace one glyph with one of many glyphs
	SubLigature                      // Ligature (format 4.1)	Replace multiple glyphs with one glyph
	SubContext                       // Context (format 5.1 5.2 5.3)	Replace one or more glyphs in context
	SubChaining                      // Chaining Context (format 6.1 6.2 6.3)	Replace one or more glyphs in chained context
	// Extension Substitution (format 7.1) Extension mechanism for other substitutions
	// The table pointed at by this lookup is returned instead.
	subExtension
	SubReverse // Reverse chaining context single (format 8.1)
)

// LookupGSUBSubtable is one of the subtables of a
// GSUB lookup.
type LookupGSUBSubtable struct {
	// For SubChaining - Format 3, its the coverage of the first input.
	Coverage Coverage
	Data     interface{ Type() GSUBType }
}

type data = interface{ Type() GSUBType }

// LookupGSUB is a lookup for GSUB tables.
// All the `Data` subtable fields have the same GSUBType.
type LookupGSUB struct {
	Flag             LookupFlag // Lookup qualifiers.
	Subtables        []LookupGSUBSubtable
	MarkFilteringSet uint16 // Meaningfull only if UseMarkFilteringSet is set.
}

// interpret the lookup as a GSUB lookup
// lookupLength is used to sanitize nested lookups
func (header lookup) parseGSUB(lookupListLength uint16) (out LookupGSUB, err error) {
	out.Flag = header.flag
	out.MarkFilteringSet = header.markFilteringSet

	out.Subtables = make([]LookupGSUBSubtable, len(header.subtableOffsets))
	for i, offset := range header.subtableOffsets {
		out.Subtables[i], err = parseGSUBSubtable(header.data, int(offset), GSUBType(header.kind), lookupListLength)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

func parseGSUBSubtable(data []byte, offset int, kind GSUBType, lookupListLength uint16) (out LookupGSUBSubtable, err error) {
	// read the format and coverage
	if offset+4 >= len(data) {
		return out, fmt.Errorf("invalid lookup subtable offset %d", offset)
	}
	format := binary.BigEndian.Uint16(data[offset:])

	// almost all table have a coverage offset, right after the format; special case the others
	// see below for the coverage
	if kind == subExtension || (kind == SubChaining || kind == SubContext) && format == 3 {
		out.Coverage = CoverageList{}
	} else {
		covOffset := binary.BigEndian.Uint16(data[offset+2:]) // relative to the subtable
		out.Coverage, err = parseCoverage(data[offset:], covOffset)
		if err != nil {
			return out, fmt.Errorf("invalid GSUB table (format %d-%d): %s", kind, format, err)
		}
	}

	// read the actual lookup
	switch kind {
	case SubSingle:
		out.Data, err = parseSingleSub(format, data[offset:])
	case SubMultiple:
		out.Data, err = parseMultipleSub(format, data[offset:], out.Coverage)
	case SubAlternate:
		out.Data, err = parseAlternateSub(format, data[offset:], out.Coverage)
	case SubLigature:
		out.Data, err = parseLigatureSub(format, data[offset:], out.Coverage)
	case SubContext:
		out.Data, err = parseSequenceContextSub(format, data[offset:], lookupListLength, &out.Coverage)
	case SubChaining:
		out.Data, err = parseChainedSequenceContextSub(format, data[offset:], lookupListLength, &out.Coverage)
	case subExtension:
		out, err = parseExtensionSub(data[offset:], lookupListLength)
	case SubReverse:
		out.Data, err = parseReverseChainedSequenceContextSub(format, data[offset:], out.Coverage)
	default:
		return out, fmt.Errorf("unsupported gsub lookup type %d", kind)
	}
	return out, err
}

func (SubstitutionSingle1) Type() GSUBType { return SubSingle }
func (SubstitutionSingle2) Type() GSUBType { return SubSingle }

// data starts at the subtable (but format has already been read)
func parseSingleSub(format uint16, data []byte) (out data, err error) {
	switch format {
	case 1:
		return parseSingleSub1(data)
	case 2:
		return parseSingleSub2(data)
	default:
		return nil, fmt.Errorf("unsupported single substitution format: %d", format)
	}
}

// Single Substitution Format 1, expressed as a delta
// from the coverage.
type SubstitutionSingle1 int16

// data is at the begining of the subtable
func parseSingleSub1(data []byte) (SubstitutionSingle1, error) {
	if len(data) < 6 {
		return 0, errors.New("invalid single subsitution table (format 1)")
	}
	// format and coverage already read
	delta := SubstitutionSingle1(binary.BigEndian.Uint16(data[4:]))
	return delta, nil
}

// Single Substitution Format 2, expressed as substitutes
type SubstitutionSingle2 []fonts.GlyphIndex

// data is at the begining of the subtable
func parseSingleSub2(data []byte) (SubstitutionSingle2, error) {
	if len(data) < 6 {
		return nil, errors.New("invalid single subsitution table (format 2)")
	}
	// format and coverage already read
	glyphCount := binary.BigEndian.Uint16(data[4:])
	if len(data) < 6+int(glyphCount)*2 {
		return nil, errors.New("invalid single subsitution table (format 2)")
	}
	out := make(SubstitutionSingle2, glyphCount)
	for i := range out {
		out[i] = fonts.GlyphIndex(binary.BigEndian.Uint16(data[6+2*i:]))
	}
	return out, nil
}

type SubstitutionMultiple [][]fonts.GlyphIndex

func (SubstitutionMultiple) Type() GSUBType { return SubMultiple }

// data starts at the subtable (but format has already been read)
func parseMultipleSub(format uint16, data []byte, cov Coverage) (SubstitutionMultiple, error) {
	if len(data) < 6 {
		return nil, errors.New("invalid multiple subsitution table")
	}

	// format and coverage already processed
	count := binary.BigEndian.Uint16(data[4:])

	// check length conformance
	if cov.Size() != int(count) {
		return nil, errors.New("invalid multiple subsitution table")
	}

	if 6+int(count)*2 > len(data) {
		return nil, fmt.Errorf("invalid multiple subsitution table")
	}

	out := make(SubstitutionMultiple, count)
	var err error
	for i := range out {
		offset := binary.BigEndian.Uint16(data[6+2*i:])
		if int(offset) > len(data) {
			return out, errors.New("invalid multiple subsitution table")
		}
		out[i], err = parseMultipleSet(data[offset:])
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

func parseMultipleSet(data []byte) ([]fonts.GlyphIndex, error) {
	if len(data) < 2 {
		return nil, errors.New("invalid multiple subsitution table")
	}
	count := binary.BigEndian.Uint16(data)
	if 2+int(count)*2 > len(data) {
		return nil, fmt.Errorf("invalid multiple subsitution table")
	}
	out := make([]fonts.GlyphIndex, count)
	for i := range out {
		out[i] = fonts.GlyphIndex(binary.BigEndian.Uint16(data[2+2*i:]))
	}
	return out, nil
}

type SubstitutionAlternate [][]fonts.GlyphIndex

func (SubstitutionAlternate) Type() GSUBType { return SubAlternate }

// data starts at the subtable (but format has already been read)
func parseAlternateSub(format uint16, data []byte, cov Coverage) (SubstitutionAlternate, error) {
	out, err := parseMultipleSub(format, data, cov)
	if err != nil {
		return nil, errors.New("invalid alternate substitution table")
	}
	return SubstitutionAlternate(out), nil
}

// SubstitutionLigature stores one ligature set per glyph in the coverage.
type SubstitutionLigature [][]LigatureGlyph

func (SubstitutionLigature) Type() GSUBType { return SubLigature }

func parseLigatureSub(format uint16, data []byte, cov Coverage) (SubstitutionLigature, error) {
	if len(data) < 6 {
		return nil, errors.New("invalid ligature subsitution table")
	}

	// format and coverage already processed
	count := binary.BigEndian.Uint16(data[4:])

	// check length conformance
	if cov.Size() != int(count) {
		return nil, errors.New("invalid ligature subsitution table")
	}

	if 6+int(count)*2 > len(data) {
		return nil, fmt.Errorf("invalid ligature subsitution table")
	}
	out := make([][]LigatureGlyph, count)
	var err error
	for i := range out {
		ligSetOffset := binary.BigEndian.Uint16(data[6+2*i:])
		if int(ligSetOffset) > len(data) {
			return out, errors.New("invalid ligature subsitution table")
		}
		out[i], err = parseLigatureSet(data[ligSetOffset:])
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

type LigatureGlyph struct {
	Glyph fonts.GlyphIndex // Output ligature glyph
	// Glyphs composing the ligature, starting after the
	// implicit first glyph, given in the coverage of the
	// SubstitutionLigature table
	Components []uint16
}

// data is at the begining of the ligature set table
func parseLigatureSet(data []byte) ([]LigatureGlyph, error) {
	if len(data) < 2 {
		return nil, errors.New("invalid ligature set table")
	}
	count := binary.BigEndian.Uint16(data)
	out := make([]LigatureGlyph, count)
	var err error
	for i := range out {
		ligOffset := binary.BigEndian.Uint16(data[2+2*i:])
		if int(ligOffset)+4 > len(data) {
			return nil, errors.New("invalid ligature set table")
		}
		out[i].Glyph = fonts.GlyphIndex(binary.BigEndian.Uint16(data[ligOffset:]))
		ligCount := binary.BigEndian.Uint16(data[ligOffset+2:])
		if ligCount == 0 {
			return nil, errors.New("invalid ligature set table")
		}
		out[i].Components, err = parseUint16s(data[ligOffset+4:], int(ligCount)-1)
		if err != nil {
			return nil, fmt.Errorf("invalid ligature set table: %s", err)
		}
	}
	return out, nil
}

type (
	SubstitutionContext1 [][]SequenceRule
	SubstitutionContext2 SequenceContext2
	SubstitutionContext3 SequenceContext3
)

func (SubstitutionContext1) Type() GSUBType { return SubContext }
func (SubstitutionContext2) Type() GSUBType { return SubContext }
func (SubstitutionContext3) Type() GSUBType { return SubContext }

// lookupLength is used to sanitize lookup indexes.
// cov is used for ContextFormat3
func parseSequenceContextSub(format uint16, data []byte, lookupLength uint16, cov *Coverage) (data, error) {
	switch format {
	case 1:
		out, err := parseSequenceContext1(data, lookupLength)
		return SubstitutionContext1(out), err
	case 2:
		out, err := parseSequenceContext2(data, lookupLength)
		return SubstitutionContext2(out), err
	case 3:
		out, err := parseSequenceContext3(data, lookupLength)
		if len(out.Coverages) != 0 {
			*cov = out.Coverages[0]
		}
		return SubstitutionContext3(out), err
	default:
		return nil, fmt.Errorf("unsupported sequence context format %d", format)
	}
}

type (
	SubstitutionChainedContext1 [][]ChainedSequenceRule
	SubstitutionChainedContext2 ChainedSequenceContext2
	SubstitutionChainedContext3 ChainedSequenceContext3
)

func (SubstitutionChainedContext1) Type() GSUBType { return SubChaining }
func (SubstitutionChainedContext2) Type() GSUBType { return SubChaining }
func (SubstitutionChainedContext3) Type() GSUBType { return SubChaining }

// lookupLength is used to sanitize lookup indexes.
// cov is used for ContextFormat3
func parseChainedSequenceContextSub(format uint16, data []byte, lookupLength uint16, cov *Coverage) (data, error) {
	switch format {
	case 1:
		out, err := parseChainedSequenceContext1(data, lookupLength)
		return SubstitutionChainedContext1(out), err
	case 2:
		out, err := parseChainedSequenceContext2(data, lookupLength)
		return SubstitutionChainedContext2(out), err
	case 3:
		out, err := parseChainedSequenceContext3(data, lookupLength)
		if len(out.Input) != 0 {
			*cov = out.Input[0]
		}
		return SubstitutionChainedContext3(out), err
	default:
		return nil, fmt.Errorf("unsupported sequence context format %d", format)
	}
}

// returns the extension subtable instead
func parseExtensionSub(data []byte, lookupListLength uint16) (LookupGSUBSubtable, error) {
	if len(data) < 8 {
		return LookupGSUBSubtable{}, errors.New("invalid extension substitution table")
	}
	extensionType := GSUBType(binary.BigEndian.Uint16(data[2:]))
	offset := binary.BigEndian.Uint32(data[4:])

	if extensionType == subExtension {
		return LookupGSUBSubtable{}, errors.New("invalid extension substitution table")
	}

	return parseGSUBSubtable(data, int(offset), extensionType, lookupListLength)
}

type SubstitutionReverseChainedContext struct {
	Backtrack   []Coverage
	Lookahead   []Coverage
	Substitutes []fonts.GlyphIndex
}

func (SubstitutionReverseChainedContext) Type() GSUBType { return SubReverse }

func parseReverseChainedSequenceContextSub(format uint16, data []byte, cov Coverage) (out SubstitutionReverseChainedContext, err error) {
	if len(data) < 6 {
		return out, errors.New("invalid reversed chained sequence context format 3 table")
	}
	covCount := binary.BigEndian.Uint16(data[4:])
	out.Backtrack = make([]Coverage, covCount)
	for i := range out.Backtrack {
		covOffset := binary.BigEndian.Uint16(data[6+2*i:])
		out.Backtrack[i], err = parseCoverage(data, covOffset)
		if err != nil {
			return out, err
		}
	}
	endBacktrack := 6 + 2*int(covCount)

	if len(data) < endBacktrack+2 {
		return out, errors.New("invalid reversed chained sequence context format 3 table")
	}
	covCount = binary.BigEndian.Uint16(data[endBacktrack:])
	out.Lookahead = make([]Coverage, covCount)
	for i := range out.Lookahead {
		covOffset := binary.BigEndian.Uint16(data[endBacktrack+2+2*i:])
		out.Lookahead[i], err = parseCoverage(data, covOffset)
		if err != nil {
			return out, err
		}
	}
	endLookahead := endBacktrack + 2 + 2*int(covCount)

	if len(data) < endBacktrack+2 {
		return out, errors.New("invalid reversed chained sequence context format 3 table")
	}
	glyphCount := binary.BigEndian.Uint16(data[endLookahead:])

	if cov.Size() != int(glyphCount) {
		return out, errors.New("invalid reversed chained sequence context format 3 table")
	}

	if len(data) < endLookahead+2+2*int(glyphCount) {
		return out, errors.New("invalid reversed chained sequence context format 3 table")
	}
	out.Substitutes = make([]fonts.GlyphIndex, glyphCount)
	for i := range out.Substitutes {
		out.Substitutes[i] = fonts.GlyphIndex(binary.BigEndian.Uint16(data[endLookahead+2+2*i:]))
	}
	return out, err
}
