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
		out.Lookups[i], err = l.parseGSUB()
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
	SubExtension                     // Extension Substitution (format 7.1)	Extension mechanism for other substitutions (i.e. this excludes the Extension type substitution itself)
	SubReverse                       // Reverse chaining context single (format 8.1)
)

// LookupGSUBSubtable is one of the subtables of a
// GSUB lookup.
type LookupGSUBSubtable struct {
	Coverage Coverage // empty for SubChaining - Format 3
	Data     interface{ Type() GSUBType }
}

// LookupGSUB is a lookup for GSUB tables.
// All the `Data` subtable fields have the same GSUBType (that is, the same concrete type).
type LookupGSUB struct {
	Flag             LookupFlag // Lookup qualifiers.
	Subtables        []LookupGSUBSubtable
	MarkFilteringSet uint16 // Meaningfull only if UseMarkFilteringSet is set.
}

// interpret the lookup as a GSUB lookup
func (header lookup) parseGSUB() (out LookupGSUB, err error) {
	out.Flag = header.flag
	out.Subtables, err = parseGSUBSubtables(GSUBType(header.kind), header.data, header.subtableOffsets)
	out.MarkFilteringSet = header.markFilteringSet
	return out, err
}

func parseGSUBSubtables(kind GSUBType, data []byte, offsets []uint16) (out []LookupGSUBSubtable, err error) {
	out = make([]LookupGSUBSubtable, len(offsets))
	for i, offset := range offsets {
		// read the format and coverage
		if int(offset)+4 >= len(data) {
			return nil, fmt.Errorf("invalid lookup subtable offset %d", offset)
		}
		format := binary.BigEndian.Uint16(data[offset:])

		// almost all table have a coverage offset (right after the format)
		if kind == SubChaining && format == 3 {
			out[i].Coverage = CoverageList{}
		} else {
			covOffset := binary.BigEndian.Uint16(data[offset+2:]) // relative to the subtable
			out[i].Coverage, err = parseCoverage(data[offset:], covOffset)
			if err != nil {
				return nil, err
			}
		}

		// read the actual lookup
		switch kind {
		case SubSingle:
			out[i].Data, err = parseSingleSub(format, data[offset:])
		case SubMultiple:
			out[i].Data, err = parseMultipleSub(format, data[offset:], out[i].Coverage)
		case SubAlternate:
			out[i].Data, err = parseAlternateSub(format, data[offset:], out[i].Coverage)
		case SubLigature:
			out[i].Data, err = parseLigatureSub(format, data[offset:], out[i].Coverage)
		case SubChaining:
			// TODO:
		default:
			fmt.Println("unsupported gsub lookup", kind)
		}
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (SingleSubstitution1) Type() GSUBType { return SubSingle }
func (SingleSubstitution2) Type() GSUBType { return SubSingle }

// data starts at the subtable (but format has already been read)
func parseSingleSub(format uint16, data []byte) (out interface{ Type() GSUBType }, err error) {
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
type SingleSubstitution1 int16

// data is at the begining of the subtable
func parseSingleSub1(data []byte) (SingleSubstitution1, error) {
	if len(data) < 6 {
		return 0, errors.New("invalid single subsitution table (format 1)")
	}
	// format and coverage already read
	delta := SingleSubstitution1(binary.BigEndian.Uint16(data[4:]))
	return delta, nil
}

// Single Substitution Format 2, expressed as substitutes
type SingleSubstitution2 []fonts.GlyphIndex

// data is at the begining of the subtable
func parseSingleSub2(data []byte) (SingleSubstitution2, error) {
	if len(data) < 6 {
		return nil, errors.New("invalid single subsitution table (format 2)")
	}
	// format and coverage already read
	glyphCount := binary.BigEndian.Uint16(data[4:])
	if len(data) < 6+int(glyphCount)*2 {
		return nil, errors.New("invalid single subsitution table (format 2)")
	}
	out := make(SingleSubstitution2, glyphCount)
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
	Glyph fonts.GlyphIndex // output ligature glyph
	// glyphs composing the ligature, starting after the
	// implicit first glyph, given in the coverage of the
	// SubstitutionLigature table
	Components []fonts.GlyphIndex
}

// data is at the begining of the ligature set table
func parseLigatureSet(data []byte) ([]LigatureGlyph, error) {
	if len(data) < 2 {
		return nil, errors.New("invalid ligature set table")
	}
	count := binary.BigEndian.Uint16(data)
	out := make([]LigatureGlyph, count)
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
		if int(ligOffset)+4+2*int(ligCount-1) > len(data) {
			return nil, errors.New("invalid ligature set table")
		}
		out[i].Components = make([]fonts.GlyphIndex, ligCount-1)
		for j := range out[i].Components {
			out[i].Components[j] = fonts.GlyphIndex(binary.BigEndian.Uint16(data[int(ligOffset)+4+2*j:]))
		}
	}
	return out, nil
}
