package truetype

import (
	"encoding/binary"
	"errors"
)

var (
	errInvalidHtmxTable = errors.New("invalid htmx table")
	errInvalidMaxpTable = errors.New("invalid maxp table")
)

func parseMaxpTable(input []byte) (numGlyphs uint16, err error) {
	if len(input) < 6 {
		return 0, errInvalidMaxpTable
	}
	out := binary.BigEndian.Uint16(input[4:6])
	return out, nil
}

type tableHVmtx []Metric // with length numGlyphs

type Metric struct {
	Advance, SideBearing int16
}

// pad the width if numberOfHMetrics < numGlyphs
func parseHVmtxTable(input []byte, numberOfHMetrics, numGlyphs uint16) (tableHVmtx, error) {
	if numberOfHMetrics == 0 {
		return nil, errors.New("number of glyph metrics is 0")
	}

	if len(input) < 4*int(numberOfHMetrics) {
		return nil, errInvalidHtmxTable
	}
	widths := make(tableHVmtx, numberOfHMetrics)
	for i := range widths {
		// we ignore the Glyph left side bearing
		widths[i].Advance = int16(binary.BigEndian.Uint16(input[2*i:]))
		widths[i].SideBearing = int16(binary.BigEndian.Uint16(input[2*i+2:]))
	}
	if numberOfHMetrics < numGlyphs { // pad with the last value
		widths = append(widths, make(tableHVmtx, numGlyphs-numberOfHMetrics)...)
		lastWidth := widths[numberOfHMetrics-1]
		for i := numberOfHMetrics; i < numGlyphs; i++ {
			widths[i] = lastWidth
		}
	}
	return widths, nil
}
