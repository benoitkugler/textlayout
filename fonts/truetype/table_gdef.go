package truetype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

type TableGDEF struct {
	// Identify the class of the glyph:
	//	1:	Base glyph, single character, spacing glyph
	//	2:	Ligature glyph (multiple character, spacing glyph)
	//	3:	Mark glyph (non-spacing combining glyph)
	//	4:	Component glyph (part of single character, spacing glyph)
	Class Class
	// Class to which a mark glyph may belong
	MarkAttach Class

	VariationStore VariationStore // for variable fonts, may be empty
}

func parseTableGdef(buf []byte) (out TableGDEF, err error) {
	r := bytes.NewReader(buf)
	var header struct {
		MajorVersion             uint16 // 	Major version of the GDEF table, = 1
		MinorVersion             uint16 // 	Minor version of the GDEF table
		GlyphClassDefOffset      uint16 // 	Offset to class definition table for glyph type, from beginning of GDEF header (may be 0)
		AttachListOffset         uint16 // 	Offset to attachment point list table, from beginning of GDEF header (may be 0)
		LigCaretListOffset       uint16 // 	Offset to ligature caret list table, from beginning of GDEF header (may be 0)
		MarkAttachClassDefOffset uint16 // 	Offset to class definition table for mark attachment type, from beginning of GDEF header (may be 0)
	}
	if err := binary.Read(r, binary.BigEndian, &header); err != nil {
		return out, err
	}

	switch header.MinorVersion {
	case 0, 2, 3:
		if header.GlyphClassDefOffset != 0 {
			out.Class, err = parseClass(buf, header.GlyphClassDefOffset)
			if err != nil {
				return out, err
			}
		}
		if header.MinorVersion == 3 { // read the additional two fields
			var fields struct {
				MarkGlyphSetsDefOffset uint16 // Offset to the table of mark glyph set definitions, from beginning of GDEF header (may be NULL)
				ItemVarStoreOffset     uint32 // Offset to the Item Variation Store table, from beginning of GDEF header (may be NULL)
			}
			if err := binary.Read(r, binary.BigEndian, &fields); err != nil {
				return out, err
			}
			if fields.ItemVarStoreOffset != 0 {
				out.VariationStore, err = parseItemVariationStore(buf, fields.ItemVarStoreOffset)
				if err != nil {
					return out, err
				}
			}
		}
	default:
		return out, fmt.Errorf("unsupported GDEF table version")
	}
	return out, nil
}

// GlyphProps is a 16-bit integer where the lower 8-bit have bits representing
// glyph class, and high 8-bit the mark attachment type (if any).
type GlyphProps = uint16

const (
	BaseGlyph GlyphProps = 1 << (iota + 1)
	Ligature
	Mark
)

// GetGlyphProps return a summary of the glyph properties.
func (t TableGDEF) GetGlyphProps(glyph fonts.GlyphIndex) GlyphProps {
	klass, _ := t.Class.ClassID(glyph)
	switch klass {
	case 1:
		return BaseGlyph
	case 2:
		return Ligature
	case 3:
		klass, _ = t.MarkAttach.ClassID(glyph)
		return Mark | GlyphProps(klass)<<8
	default:
		return 0
	}
}

// VariationRegion stores start, peek, end coordinates.
type VariationRegion [3]float32

// return the factor
func (reg VariationRegion) evaluate(coord float32) float32 {
	start, peak, end := reg[0], reg[1], reg[2]
	if peak == 0 || coord == peak {
		return 1.
	}

	if coord <= start || end <= coord {
		return 0.
	}

	/* Interpolate */
	if coord < peak {
		return (coord - start) / (peak - start)
	}
	return (end - coord) / (end - peak)
}

// TODO: sanitize array length using FVar
// After successful parsing, every region indexes in `Datas` elements are valid.
type VariationStore struct {
	Regions [][]VariationRegion // for each region, for each axis
	Datas   []ItemVariationData
}

func parseItemVariationStore(data []byte, offset uint32) (out VariationStore, err error) {
	if len(data) < int(offset)+8 {
		return out, errors.New("invalid item variation store (EOF)")
	}
	data = data[offset:]
	// format is ignored
	regionsOffset := binary.BigEndian.Uint32(data[2:])
	count := binary.BigEndian.Uint16(data[6:])

	out.Regions, err = parseItemVariationRegions(data, regionsOffset)
	if err != nil {
		return out, err
	}

	if len(data) < 8+4*int(count) {
		return out, errors.New("invalid item variation store (EOF)")
	}
	out.Datas = make([]ItemVariationData, count)
	for i := range out.Datas {
		subtableOffset := binary.BigEndian.Uint32(data[8+4*i:])
		out.Datas[i], err = parseItemVariationData(data, subtableOffset, uint16(len(out.Regions)))
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

func parseItemVariationRegions(data []byte, offset uint32) ([][]VariationRegion, error) {
	if len(data) < int(offset)+4 {
		return nil, errors.New("invalid item variation regions list (EOF)")
	}
	data = data[offset:]
	axisCount := int(binary.BigEndian.Uint16(data))
	regionCount := int(binary.BigEndian.Uint16(data[2:]))

	if len(data) < 4+6*axisCount*regionCount {
		return nil, errors.New("invalid item variation regions list (EOF)")
	}
	regions := make([][]VariationRegion, regionCount)
	for i := range regions {
		ri := make([]VariationRegion, axisCount)
		for j := range ri {
			start := fixed214ToFloat(binary.BigEndian.Uint16(data[4+(i*axisCount+j)*6:]))
			peak := fixed214ToFloat(binary.BigEndian.Uint16(data[4+(i*axisCount+j)*6+2:]))
			end := fixed214ToFloat(binary.BigEndian.Uint16(data[4+(i*axisCount+j)*6+4:]))

			if start > peak || peak > end {
				return nil, errors.New("invalid item variation regions list")
			}
			if start < 0 && end > 0 && peak != 0 {
				return nil, errors.New("invalid item variation regions list")
			}
			ri[j] = VariationRegion{start, peak, end}
		}
		regions[i] = ri
	}
	return regions, nil
}

type ItemVariationData struct {
	RegionIndexes []uint16  // Array of indices into the variation region list for the regions referenced by this item variation data table.
	Deltas        [][]int16 // Each row as the same length as `RegionIndexes`
}

func parseItemVariationData(data []byte, offset uint32, nbRegions uint16) (out ItemVariationData, err error) {
	if len(data) < int(offset)+6 {
		return out, errors.New("invalid item variation data subtable (EOF)")
	}
	data = data[offset:]
	itemCount := int(binary.BigEndian.Uint16(data))
	shortDeltaCount := int(binary.BigEndian.Uint16(data[2:]))
	regionIndexCount := int(binary.BigEndian.Uint16(data[4:]))

	out.RegionIndexes, err = parseUint16s(data[6:], regionIndexCount)
	if err != nil {
		return out, fmt.Errorf("invalid item variation data subtable: %s", err)
	}
	// sanitize the indexes
	for _, regionIndex := range out.RegionIndexes {
		if regionIndex >= nbRegions {
			return out, fmt.Errorf("invalid item variation region index: %d (for size %d)", regionIndex, nbRegions)
		}
	}

	data = data[6+2*regionIndexCount:] // length checked by the previous `parseUint16s` call
	rowLength := shortDeltaCount + regionIndexCount
	if len(data) < itemCount*rowLength {
		return out, errors.New("invalid item variation data subtable (EOF)")
	}
	if shortDeltaCount > regionIndexCount {
		return out, errors.New("invalid item variation data subtable")
	}
	out.Deltas = make([][]int16, itemCount)
	for i := range out.Deltas {
		vi := make([]int16, regionIndexCount)
		j := 0
		for ; j < shortDeltaCount; j++ {
			vi[j] = int16(binary.BigEndian.Uint16(data[2*j:]))
		}
		for ; j < regionIndexCount; j++ {
			vi[j] = int16(int8(data[shortDeltaCount+j]))
		}
		out.Deltas[i] = vi
		data = data[rowLength:]
	}
	return out, nil
}
