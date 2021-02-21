package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

type TableGlyf []GlyphData // length numGlyphs

const maxCompositeNesting = 20 // protect against malicious fonts

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

type contourPoint struct {
	x, y       float32
	isEndPoint bool
	isExplicit bool // this point is referenced, i.e., explicit deltas specified */
}

func (c *contourPoint) translate(x, y float32) {
	c.x += x
	c.y += y
}

func (c *contourPoint) transform(matrix [4]float32) {
	px := c.x*matrix[0] + c.y*matrix[2]
	c.y = c.x*matrix[1] + c.y*matrix[3]
	c.x = px
}

type GlyphData struct {
	data interface{ isGlyphData() }

	Xmin, Ymin, Xmax, Ymax int16
}

func (simpleGlyphData) isGlyphData()    {}
func (compositeGlyphData) isGlyphData() {}

// including phantom points
func (g GlyphData) pointNumbersCount() int {
	switch g := g.data.(type) {
	case simpleGlyphData:
		if L := len(g.endPtsOfContours); L >= 1 {
			return int(g.endPtsOfContours[L-1]) + 1 + 4
		}
	case compositeGlyphData:
		/* pseudo component points for each component in composite glyph */
		return len(g.glyphs) + 4
	}
	return 4
}

func parseGlyphData(data []byte, offset uint32) (out GlyphData, err error) {
	if len(data) < int(offset)+10 {
		return out, errors.New("invalid 'glyf' table (EOF)")
	}
	data = data[offset:]
	numberOfContours := int(int16(binary.BigEndian.Uint16(data))) // careful with the conversion to signed integer
	out.Xmin = int16(binary.BigEndian.Uint16(data[2:]))
	out.Ymin = int16(binary.BigEndian.Uint16(data[4:]))
	out.Xmax = int16(binary.BigEndian.Uint16(data[6:]))
	out.Ymax = int16(binary.BigEndian.Uint16(data[8:]))
	if numberOfContours >= 0 { // simple glyph
		out.data, err = parseSimpleGlyphData(data[10:], numberOfContours)
	} else { // composite glyph
		out.data, err = parseCompositeGlyphData(data[10:])
	}
	return out, err
}

type simpleGlyphData struct {
	endPtsOfContours []uint16
	instructions     []byte
}

func (sg simpleGlyphData) getContourPoints(phantomOnly bool) []contourPoint {
	numPoints := sg.endPtsOfContours[len(sg.endPtsOfContours)-1] + 1

	points := make([]contourPoint, numPoints)
	if phantomOnly {
		return points
	}
	return points // TODO: complete for phantomOnly = false
	// for _, end := range sg.endPtsOfContours {
	// 	points[sg.endPtsOfContours[i]].isEndPoint = true
	// }

	// /* Read flags */
	// for (unsigned int i = 0; i < numPoints; i++){
	// if (unlikely (!bytes.check_range (p))) return false;
	// uint8_t flag = *p++;
	// points[i].flag = flag;
	// if (flag & FLAG_REPEAT)
	// {
	// if (unlikely (!bytes.check_range (p))) return false;
	// unsigned int repeat_count = *p++;
	// while ((repeat_count-- > 0) && (++i < numPoints))
	// points[i].flag = flag;
	// }
	// }

	// /* Read x & y coordinates */
	// return read_points (p, points, bytes, [] (contour_point_t &p, float v) { p.x = v; },
	//  FLAG_X_SHORT, FLAG_X_SAME)
	// && read_points (p, points, bytes, [] (contour_point_t &p, float v) { p.y = v; },
	//  FLAG_Y_SHORT, FLAG_Y_SAME);
	// }
}

// data starts after the glyph header
func parseSimpleGlyphData(data []byte, numberOfContours int) (out simpleGlyphData, err error) {
	out.endPtsOfContours, err = parseUint16s(data, numberOfContours)
	if err != nil {
		return out, fmt.Errorf("invalid simple glyph data: %s", err)
	}

	out.instructions, _, err = parseGlyphInstruction(data[2*numberOfContours:])
	if err != nil {
		return out, fmt.Errorf("invalid simple glyph data: %s", err)
	}

	return out, err
}

type compositeGlyphData struct {
	glyphs       []compositeGlyphPart
	instructions []byte
}

type compositeGlyphPart struct {
	flags      uint16
	glyphIndex fonts.GlyphIndex
	arg1, arg2 uint16     // before interpretation
	scale      [4]float32 // x, 01, 10, y
}

func (c *compositeGlyphPart) hasUseMyMetrics() bool {
	const useMyMetrics = 0x0200
	return c.flags&useMyMetrics != 0
}

// return true if arg1 and arg2 indicated an anchor point,
// not offsets
func (c *compositeGlyphPart) isAnchored() bool {
	const argsAreXyValues = 0x0002
	return c.flags&argsAreXyValues == 0
}

func (c *compositeGlyphPart) isScaledOffsets() bool {
	const (
		scaledComponentOffset   = 0x0800
		unscaledComponentOffset = 0x1000
	)
	return c.flags&(scaledComponentOffset|unscaledComponentOffset) == scaledComponentOffset
}

func (c *compositeGlyphPart) transformPoints(points []contourPoint) {
	if c.isAnchored() {
		return
	}
	transX, transY := float32(int16(c.arg1)), float32(int16(c.arg2))
	scale := c.scale
	if c.isScaledOffsets() {
		for i := range points {
			points[i].translate(transX, transY)
			points[i].transform(scale)
		}
	} else {
		for i := range points {
			points[i].transform(scale)
			points[i].translate(transX, transY)
		}
	}
}

// data starts after the glyph header
func parseCompositeGlyphData(data []byte) (out compositeGlyphData, err error) {
	const (
		arg1And2AreWords = 1 << iota
		_
		_
		weHaveAScale
		_
		moreComponents
		weHaveAnXAndYScale
		weHaveATwoByTwo
		weHaveInstructions
	)
	var flags uint16
	for do := true; do; do = flags&moreComponents != 0 {
		var part compositeGlyphPart

		if len(data) < 4 {
			return out, errors.New("invalid composite glyph data (EOF)")
		}
		flags = binary.BigEndian.Uint16(data)
		part.glyphIndex = fonts.GlyphIndex(binary.BigEndian.Uint16(data[2:]))

		if flags&arg1And2AreWords != 0 { // 16 bits
			if len(data) < 4+4 {
				return out, errors.New("invalid composite glyph data (EOF)")
			}
			part.arg1 = binary.BigEndian.Uint16(data[4:])
			part.arg2 = binary.BigEndian.Uint16(data[6:])
			data = data[8:]
		} else {
			if len(data) < 4+2 {
				return out, errors.New("invalid composite glyph data (EOF)")
			}
			part.arg1 = uint16(data[4])
			part.arg2 = uint16(data[5])
			data = data[6:]
		}

		part.scale[0], part.scale[3] = 1, 1
		if flags&weHaveAScale != 0 {
			if len(data) < 2 {
				return out, errors.New("invalid composite glyph data (EOF)")
			}
			part.scale[0] = fixed214ToFloat(binary.BigEndian.Uint16(data))
			part.scale[3] = part.scale[0]
			data = data[2:]
		} else if flags&weHaveAnXAndYScale != 0 {
			if len(data) < 4 {
				return out, errors.New("invalid composite glyph data (EOF)")
			}
			part.scale[0] = fixed214ToFloat(binary.BigEndian.Uint16(data))
			part.scale[3] = fixed214ToFloat(binary.BigEndian.Uint16(data[2:]))
			data = data[4:]
		} else if flags&weHaveATwoByTwo != 0 {
			if len(data) < 8 {
				return out, errors.New("invalid composite glyph data (EOF)")
			}
			part.scale[0] = fixed214ToFloat(binary.BigEndian.Uint16(data))
			part.scale[1] = fixed214ToFloat(binary.BigEndian.Uint16(data[2:]))
			part.scale[2] = fixed214ToFloat(binary.BigEndian.Uint16(data[4:]))
			part.scale[3] = fixed214ToFloat(binary.BigEndian.Uint16(data[6:]))
			data = data[8:]
		}

		out.glyphs = append(out.glyphs, part)
	}
	if flags&weHaveInstructions != 0 {
		out.instructions, _, err = parseGlyphInstruction(data)
		if err != nil {
			return out, fmt.Errorf("invalid composite glyph data: %s", err)
		}
	}
	return out, nil
}

func parseGlyphInstruction(data []byte) ([]byte, []byte, error) {
	if len(data) < 2 {
		return nil, nil, errors.New("invalid glyph instructions (EOF)")
	}
	instructionLength := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+instructionLength {
		return nil, nil, errors.New("invalid glyph instructions (EOF)")
	}
	return data[2 : 2+instructionLength], data[2+instructionLength:], nil
}
