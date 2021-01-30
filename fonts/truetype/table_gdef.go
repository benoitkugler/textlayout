package truetype

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

type TableGDEF struct {
	// Identify the class of the glyph:
	//	1:	Base glyph, single character, spacing glyph
	//	2:	Ligature glyph (multiple character, spacing glyph)
	//	3:	Mark glyph (non-spacing combining glyph)
	//	4:	Component glyph (part of single character, spacing glyph)
	Class class
	// Class to which a mark glyph may belong
	MarkAttach class
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
			out.Class, err = fetchClassLookup(buf, header.GlyphClassDefOffset)
			if err != nil {
				return out, err
			}
		}
	default:
		return out, fmt.Errorf("unsupported GDEF table version")
	}
	return out, nil
}

// GlyphProps is a 16-bit integer where the lower 8-bit have bits representing
// glyph class, and high 8-bit the mark attachment type (if any).
// Not to be confused with lookup_props which is very similar.
type GlyphProps = uint16

const (
	// The following three match LookupFlags::Ignore* numbers.
	BaseGlyph GlyphProps = 1 << (iota + 1)
	Ligature
	Mark
)

// GetGlyphProps return a summary of the glyph properties.
func (t TableGDEF) GetGlyphProps(glyph fonts.GlyphIndex) GlyphProps {
	klass := t.Class.ClassID(glyph)
	switch klass {
	case 1:
		return BaseGlyph
	case 2:
		return Ligature
	case 3:
		klass = t.MarkAttach.ClassID(glyph)
		return Mark | GlyphProps(klass)<<8
	default:
		return 0
	}
}
