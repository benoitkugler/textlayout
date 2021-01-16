package truetype

import (
	"bytes"
	"encoding/binary"
	"io"
)

type TableOS2 struct {
	Version             uint16
	XAvgCharWidth       uint16
	USWeightClass       uint16
	USWidthClass        uint16
	FSType              uint16
	YSubscriptXSize     int16
	YSubscriptYSize     int16
	YSubscriptXOffset   int16
	YSubscriptYOffset   int16
	YSuperscriptXSize   int16
	YSuperscriptYSize   int16
	YSuperscriptXOffset int16
	YSuperscriptYOffset int16
	YStrikeoutSize      int16
	YStrikeoutPosition  int16
	SFamilyClass        int16
	Panose              [10]byte
	UlCharRange         [4]uint32
	AchVendID           TableTag
	FsSelection         uint16
	FsFirstCharIndex    uint16
	FsLastCharIndex     uint16
	STypoAscender       int16
	STypoDescender      int16
	STypoLineGap        int16
	UsWinAscent         uint16
	UsWinDescent        uint16
	UlCodePageRange1    uint32
	UlCodePageRange2    uint32
	SxHeigh             int16
	SCapHeight          int16
	UsDefaultChar       uint16
	UsBreakChar         uint16
	UsMaxContext        uint16
	UsLowerPointSize    uint16
	UsUpperPointSize    uint16
}

func parseTableOS2(buf []byte) (*TableOS2, error) {
	var table TableOS2
	if err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &table); err != nil {
		// Different versions of the table are different lengths, as such
		// we may not already read every field.
		if err != io.ErrUnexpectedEOF {
			return nil, err
		}

		// TODO Check the len(buf) is expected for this version
	}

	return &table, nil
}
