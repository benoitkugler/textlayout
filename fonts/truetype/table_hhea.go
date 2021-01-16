package truetype

import (
	"bytes"
	"encoding/binary"
)

type TableHhea struct {
	Version             fixed
	Ascent              int16
	Descent             int16
	LineGap             int16
	AdvanceWidthMax     uint16
	MinLeftSideBearing  int16
	MinRightSideBearing int16
	XMaxExtent          int16
	CaretSlopeRise      int16
	CaretSlopeRun       int16
	CaretOffset         int16
	Reserved1           int16
	Reserved2           int16
	Reserved3           int16
	Reserved4           int16
	MetricDataformat    int16
	NumOfLongHorMetrics int16
}

func parseTableHhea(buf []byte) (*TableHhea, error) {
	var fields TableHhea
	err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &fields)
	return &fields, err
}
