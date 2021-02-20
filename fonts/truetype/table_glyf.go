package truetype

import (
	"encoding/binary"
	"errors"
)

type TableGlyf []GlyphData // length numGlyphs

// locaOffsets has length numGlyphs + 1
func parseTableGlyf(data []byte, locaOffsets []uint32) (TableGlyf, error) {
	out := make(TableGlyf, len(locaOffsets)-1)
	var err error
	for i := range out {
		// If a glyph has no outline, then loca[n] = loca [n+1].
		if locaOffsets[i] == locaOffsets[i+1] {
			continue
		}
		out[i], err = parseGlyphData(data, locaOffsets[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

type GlyphData struct {
	numberOfContours       int16
	Xmin, Ymin, Xmax, Ymax int16

	endPtsOfContours []uint16 // for simple glyph
}

// including phantom points
func (g GlyphData) pointNumbersCount() int {
	if L := len(g.endPtsOfContours); L >= 1 {
		return int(g.endPtsOfContours[L-1]) + 1 + 4
	}
	return 4
	// TODO: composite glyph
}

func parseGlyphData(data []byte, offset uint32) (out GlyphData, err error) {
	if len(data) < int(offset)+10 {
		return out, errors.New("invalid 'glyf' table (EOF)")
	}
	data = data[offset:]
	out.numberOfContours = int16(binary.BigEndian.Uint16(data))
	out.Xmin = int16(binary.BigEndian.Uint16(data[2:]))
	out.Ymin = int16(binary.BigEndian.Uint16(data[4:]))
	out.Xmax = int16(binary.BigEndian.Uint16(data[6:]))
	out.Ymax = int16(binary.BigEndian.Uint16(data[8:]))

	if out.numberOfContours >= 0 {
		out.endPtsOfContours, err = parseSimpleGlyphData(data[10:], int(out.numberOfContours))
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

// data starts after the glyph header
func parseSimpleGlyphData(data []byte, numberOfContours int) ([]uint16, error) {
	if len(data) < 2*numberOfContours+2 {
		return nil, errors.New("invalid simple glyph data (EOF)")
	}
	endsPts, _ := parseUint16s(data, numberOfContours)

	// instructionLength := binary.BigEndian.Uint16(data[2*numberOfContours:])
	// if len(data) < 2*numberOfContours+2+int(instructionLength) {
	// 	return nil, errors.New("invalid simple glyph data (EOF)")
	// }

	// if numberOfContours == 0 {
	// 	return nil, nil // TODO:
	// }

	// numPoints := endsPts[numberOfContours-1] + 1
	return endsPts, nil
}

// shared with gvar
func parseTableLoca(data []byte, numGlyphs int, isLong bool) ([]uint32, error) {
	var size int
	if isLong {
		size = (numGlyphs + 1) * 4
	} else {
		size = (numGlyphs + 1) * 2
	}
	if len(data) < size {
		return nil, errors.New("invalid location table (EOF)")
	}
	out := make([]uint32, numGlyphs+1)
	if isLong {
		for i := range out {
			out[i] = binary.BigEndian.Uint32(data[4*i:])
		}
	} else {
		for i := range out {
			out[i] = 2 * uint32(binary.BigEndian.Uint16(data[2*i:])) // The actual local offset divided by 2 is stored.
		}
	}
	return out, nil
}
