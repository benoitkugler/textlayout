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

// LookupGSUB is a lookup for GSUB tables.
type LookupGSUB struct {
	Flag uint16 // Lookup qualifiers.
	Data interface{ Type() GSUBType }
}

// interpret the lookup as a GSUB lookup
func (header lookup) parseGSUB() (out LookupGSUB, err error) {
	out.Flag = header.flag
	switch GSUBType(header.kind) {
	case SubSingle:
		out.Data, err = parseSingleSub(header.subtableOffsets, header.data)
	case SubAlternate:
		// TODO:
	case SubLigature:
		out.Data, err = parseLigatureSub(header.subtableOffsets, header.data)
	case SubChaining:
		// TODO:
	default:
		fmt.Println("unsupported gsub lookup", header)
	}
	return out, err
}

// TODO: probably use an interface instead
type SubstitutionSingle struct {
	Format1 []SingleSubstitution1
	Format2 []SingleSubstitution2
}

func (SubstitutionSingle) Type() GSUBType { return SubSingle }

func parseSingleSub(offsets []uint16, data []byte) (out SubstitutionSingle, err error) {
	for _, offset := range offsets {
		if int(offset)+2 >= len(data) {
			return out, fmt.Errorf("invalid lookup subtable offset %d", offset)
		}
		format := binary.BigEndian.Uint16(data[offset:])
		switch format {
		case 1:
			s, err := parseSingleSub1(data[offset:])
			if err != nil {
				return out, err
			}
			out.Format1 = append(out.Format1, s)
		case 2:
			s, err := parseSingleSub2(data[offset:])
			if err != nil {
				return out, err
			}
			out.Format2 = append(out.Format2, s)
		default:
			return out, fmt.Errorf("unsupported single substitution format: %d", format)
		}
	}
	return out, nil
}

// Single Substitution Format 1
type SingleSubstitution1 struct {
	Coverage Coverage
	Delta    int16
}

// data is at the begining of the subtable
func parseSingleSub1(data []byte) (out SingleSubstitution1, err error) {
	if len(data) < 6 {
		return out, errors.New("invalid single subsitution table (format 1)")
	}
	// format = ...
	covOffset := binary.BigEndian.Uint16(data[2:])
	out.Coverage, err = parseCoverage(data, covOffset)
	if err != nil {
		return out, err
	}
	out.Delta = int16(binary.BigEndian.Uint16(data[4:]))
	return out, nil
}

// Single Substitution Format 2
type SingleSubstitution2 struct {
	Coverage    Coverage
	Substitutes []fonts.GlyphIndex
}

// data is at the begining of the subtable
func parseSingleSub2(data []byte) (out SingleSubstitution2, err error) {
	if len(data) < 6 {
		return out, errors.New("invalid single subsitution table (format 2)")
	}
	// format = ...
	covOffset := binary.BigEndian.Uint16(data[2:])
	out.Coverage, err = parseCoverage(data, covOffset)
	if err != nil {
		return out, err
	}
	glyphCount := binary.BigEndian.Uint16(data[4:])
	if len(data) < 6+int(glyphCount)*2 {
		return out, errors.New("invalid single subsitution table (format 2)")
	}
	out.Substitutes = make([]fonts.GlyphIndex, glyphCount)
	for i := range out.Substitutes {
		out.Substitutes[i] = fonts.GlyphIndex(binary.BigEndian.Uint16(data[6+2*i:]))
	}
	return out, nil
}

type SubstitutionLigature struct {
	Coverage  Coverage
	Ligatures [][]LigatureGlyph
}

func (SubstitutionLigature) Type() GSUBType { return SubLigature }

func parseLigatureSub(offsets []uint16, data []byte) (out SubstitutionLigature, err error) {
	if len(offsets) != 1 {
		return out, fmt.Errorf("unsupported number of subtables for ligatures %d", len(offsets))
	}
	offset := offsets[0]
	if int(offset)+6 > len(data) {
		return out, fmt.Errorf("invalid lookup subtable offset %d", offset)
	}
	data = data[offset:] // now at beginning of the subtable

	// format = ...
	coverageOffset := binary.BigEndian.Uint16(data[2:])
	out.Coverage, err = parseCoverage(data, coverageOffset)
	if err != nil {
		return out, err
	}

	count := binary.BigEndian.Uint16(data[4:])
	out.Ligatures = make([][]LigatureGlyph, count)
	if 6+int(count)*2 > len(data) {
		return out, fmt.Errorf("invalid lookup subtable")
	}
	for i := range out.Ligatures {
		ligSetOffset := binary.BigEndian.Uint16(data[6+2*i:])
		if int(ligSetOffset) > len(data) {
			return out, errors.New("invalid lookup subtable")
		}
		out.Ligatures[i], err = parseLigatureSet(data[ligSetOffset:])
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

type LigatureGlyph struct {
	Glyph      fonts.GlyphIndex
	Components []fonts.GlyphIndex // len = componentCount - 1
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
