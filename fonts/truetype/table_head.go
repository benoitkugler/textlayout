package truetype

import (
	"bytes"
	"encoding/binary"
)

// TableHead contains critical information about the rest of the font.
// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6head.html
type TableHead struct {
	VersionNumber      fixed
	FontRevision       uint32
	CheckSumAdjustment uint32
	MagicNumber        uint32
	Flags              uint16
	UnitsPerEm         uint16
	Created            longdatetime
	Updated            longdatetime
	XMin               int16
	YMin               int16
	XMax               int16
	YMax               int16
	MacStyle           uint16
	LowestRecPPEM      uint16
	FontDirection      int16
	IndexToLocFormat   int16
	GlyphDataFormat    int16
}

func parseTableHead(buf []byte) (*TableHead, error) {
	var fields TableHead
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &fields)
	return &fields, err
}

// ExpectedChecksum is the checksum that the file should have had.
func (table *TableHead) ExpectedChecksum() uint32 {
	return 0xB1B0AFBA - table.CheckSumAdjustment
}
