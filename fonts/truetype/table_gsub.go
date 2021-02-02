package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

func (header *Lookup) parseGSUB() (err error) {
	var ligatures SubLigature
	switch header.Type {
	case 1:
		err = parseSingleSub(header.subtableOffsets, header.data)
	case 3:
	case 4:
		ligatures, err = parseLigatureSub(header.subtableOffsets, header.data)
		fmt.Println(ligatures)
	case 6:
	default:
		fmt.Println("unsupported gsub lookup", header)
	}
	return err
}

func parseSingleSub(offsets []uint16, data []byte) error {
	for _, offset := range offsets {
		if int(offset) >= len(data) {
			return fmt.Errorf("invalid lookup subtable offset %d", offset)
		}
	}
	// fmt.Println(len(offsets))
	return nil
}

type SubLigature struct {
	Coverage  Coverage
	Ligatures [][]LigatureGlyph
}

func parseLigatureSub(offsets []uint16, data []byte) (out SubLigature, err error) {
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
	out.Coverage, err = parseCoverage(data, int(coverageOffset))
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
