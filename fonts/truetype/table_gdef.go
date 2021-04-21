package truetype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type TableGDEF struct {
	// Identify the class of the glyph:
	//	1:	Base glyph, single character, spacing glyph
	//	2:	Ligature glyph (multiple character, spacing glyph)
	//	3:	Mark glyph (non-spacing combining glyph)
	//	4:	Component glyph (part of single character, spacing glyph)
	// May be nil
	Class Class
	// Class to which a mark glyph may belong (may be nil)
	MarkAttach     Class
	MarkGlyphSet   []Coverage     // used in GSUB and GPOS lookups to filter which marks in a string are considered or ignored
	VariationStore VariationStore // for variable fonts, may be empty
}

func parseTableGdef(buf []byte, axisCount int) (out TableGDEF, err error) {
	r := bytes.NewReader(buf)
	var header struct {
		MajorVersion             uint16 // 	Major version of the GDEF table, = 1
		MinorVersion             uint16 // 	Minor version of the GDEF table
		GlyphClassDefOffset      uint16 // 	Offset to class definition table for glyph type, from beginning of GDEF header (may be 0)
		AttachListOffset         uint16 // 	Offset to attachment point list table, from beginning of GDEF header (may be 0)
		LigCaretListOffset       uint16 // 	Offset to ligature caret list table, from beginning of GDEF header (may be 0)
		MarkAttachClassDefOffset uint16 // 	Offset to class definition table for mark attachment type, from beginning of GDEF header (may be 0)
	}
	if err = binary.Read(r, binary.BigEndian, &header); err != nil {
		return out, err
	}

	if off := header.MarkAttachClassDefOffset; off != 0 {
		out.MarkAttach, err = parseClass(buf, off)
		if err != nil {
			return out, err
		}
	}

	switch header.MinorVersion {
	case 0, 2, 3:
		if off := header.GlyphClassDefOffset; off != 0 {
			out.Class, err = parseClass(buf, off)
			if err != nil {
				return out, err
			}
		}
		if header.MinorVersion >= 2 {
			var markGlyphSetsDefOffset uint16
			if err = binary.Read(r, binary.BigEndian, &markGlyphSetsDefOffset); err != nil {
				return out, err
			}
			if markGlyphSetsDefOffset != 0 {
				out.MarkGlyphSet, err = parseMarkGlyphSet(buf, markGlyphSetsDefOffset)
				if err != nil {
					return out, err
				}
			}
		}
		if header.MinorVersion == 3 {
			var itemVarStoreOffset uint32
			if err = binary.Read(r, binary.BigEndian, &itemVarStoreOffset); err != nil {
				return out, err
			}
			if itemVarStoreOffset != 0 {
				out.VariationStore, err = parseVariationStore(buf, itemVarStoreOffset, axisCount)
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
func (t TableGDEF) GetGlyphProps(glyph GID) GlyphProps {
	klass, _ := t.Class.ClassID(glyph)
	switch klass {
	case 1:
		return BaseGlyph
	case 2:
		return Ligature
	case 3:
		var klass uint32 // it is actually a byte
		if t.MarkAttach != nil {
			klass, _ = t.MarkAttach.ClassID(glyph)
		}
		return Mark | GlyphProps(klass)<<8
	default:
		return 0
	}
}

func parseMarkGlyphSet(data []byte, offset uint16) ([]Coverage, error) {
	if len(data) < 4+int(offset) {
		return nil, errors.New("invalid mark glyph set (EOF)")
	}
	data = data[offset:]
	// format :
	count := binary.BigEndian.Uint16(data[2:])
	if len(data) < 4*int(count) {
		return nil, errors.New("invalid mark glyph set (EOF)")
	}
	out := make([]Coverage, count)
	var err error
	for i := range out {
		covOffset := binary.BigEndian.Uint32(data[4+4*i:])
		out[i], err = parseCoverage(data, covOffset)
		if err != nil {
			return nil, fmt.Errorf("invalid mark glyph set: %s", err)
		}
	}
	return out, nil
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
