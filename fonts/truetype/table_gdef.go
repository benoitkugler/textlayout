package truetype

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type TableGDEF struct {
	Class class
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
