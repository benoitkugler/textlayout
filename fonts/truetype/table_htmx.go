package truetype

import "errors"

var (
	errInvalidHtmxTable = errors.New("invalid htmx table")
	errInvalidMaxpTable = errors.New("invalid maxp table")
)

func parseMaxpTable(input []byte) (numGlyphs uint16, err error) {
	if len(input) < 6 {
		return 0, errInvalidMaxpTable
	}
	out := be.Uint16(input[4:6])
	return out, nil
}

// pad the width if numberOfHMetrics < numGlyphs
func parseHtmxTable(input []byte, numberOfHMetrics, numGlyphs uint16) ([]int, error) {
	if numberOfHMetrics == 0 {
		return nil, errors.New("number of glyph metrics is 0")
	}

	if len(input) < 4*int(numberOfHMetrics) {
		return nil, errInvalidHtmxTable
	}
	widths := make([]int, numberOfHMetrics)
	for i := range widths {
		// we ignore the Glyph left side bearing
		widths[i] = int(be.Uint16(input[2*i : 2*i+2]))
	}
	if numberOfHMetrics < numGlyphs { // pad
		widths = append(widths, make([]int, numGlyphs-numberOfHMetrics)...)
		lastWidth := widths[numberOfHMetrics-1]
		for i := numberOfHMetrics; i < numGlyphs; i++ {
			widths[i] = lastWidth
		}
	}
	return widths, nil
}
