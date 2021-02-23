package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

// ---------------------------------------- sbix ----------------------------------------

type tableSbix struct {
	strikes      []bitmapStrike
	drawOutlines bool
}

func parseTableSbix(data []byte, numGlyphs int) (out tableSbix, err error) {
	if len(data) < 8 {
		return out, errors.New("invalid 'sbix' table (EOF)")
	}
	flag := binary.BigEndian.Uint16(data[2:])
	numStrikes := int(binary.BigEndian.Uint32(data[4:]))

	out.drawOutlines = flag&0x02 != 0

	if len(data) < 8+8*numStrikes {
		return out, errors.New("invalid 'sbix' table (EOF)")
	}
	out.strikes = make([]bitmapStrike, numStrikes)
	for i := range out.strikes {
		offset := binary.BigEndian.Uint32(data[8+4*i:])
		out.strikes[i], err = parseBitmapStrike(data, offset, numGlyphs)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

type bitmapStrike struct {
	// length numGlyph; items may be empty (see isNil)
	glyphs    []bitmapGlyphData
	ppem, ppi uint16
}

func parseBitmapStrike(data []byte, offset uint32, numGlyphs int) (out bitmapStrike, err error) {
	if len(data) < int(offset)+4+4*(numGlyphs+1) {
		return out, errors.New("invalud sbix bitmap strike (EOF)")
	}
	data = data[offset:]
	out.ppem = binary.BigEndian.Uint16(data)
	out.ppi = binary.BigEndian.Uint16(data[2:])

	offsets, _ := parseTableLoca(data[4:], numGlyphs, true)
	out.glyphs = make([]bitmapGlyphData, numGlyphs)
	for i := range out.glyphs {
		if offsets[i] == offsets[i+1] { // no data
			continue
		}

		out.glyphs[i], err = parseBitmapGlyphData(data, offsets[i], offsets[i+1])
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

type bitmapGlyphData struct {
	data                         []byte
	originOffsetX, originOffsetY int16 // in font units
	graphicType                  Tag
}

func (b bitmapGlyphData) isNil() bool { return b.graphicType == 0 }

func parseBitmapGlyphData(data []byte, offsetStart, offsetNext uint32) (out bitmapGlyphData, err error) {
	if len(data) < int(offsetStart)+8 || offsetStart+8 > offsetNext {
		return out, errors.New("invalid 'sbix' bitmap glyph data (EOF)")
	}
	data = data[offsetStart:]
	out.originOffsetX = int16(binary.BigEndian.Uint16(data))
	out.originOffsetY = int16(binary.BigEndian.Uint16(data[2:]))
	out.graphicType = Tag(binary.BigEndian.Uint32(data[4:]))
	out.data = data[8 : offsetNext-offsetStart]

	if out.graphicType == 0 {
		return out, errors.New("invalid 'sbix' zero bitmap type")
	}
	return out, nil
}

// ------------------------------------ cblt/cblc/eblt/eblc ------------------------------------

type tableCblc []bitmapSize

func parseTableCblc(data []byte) (tableCblc, error) {
	if len(data) < 8 {
		return nil, errors.New("invalid bitmap location table (EOF)")
	}
	numSizes := int(binary.BigEndian.Uint32(data[4:]))
	if len(data) < 8+numSizes*bitmapSizeLength {
		return nil, errors.New("invalid bitmap location table (EOF)")
	}
	out := make(tableCblc, numSizes) // guarded by the check above
	var err error
	for i := range out {
		out[i], err = parseBitmapSize(data, 8+i*bitmapSizeLength)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

type bitmapSize struct {
	subTables            []indexSubTable
	hori, vert           sbitLineMetrics
	startGlyph, endGlyph fonts.GlyphIndex
	ppemX, ppemY         uint8
	bitDepth             uint8
	flags                uint8
}

const (
	sbitLineMetricsLength = 12
	bitmapSizeLength      = 24 + 2*sbitLineMetricsLength
)

// length as been checked
func parseBitmapSize(data []byte, offset int) (out bitmapSize, err error) {
	strikeData := data[offset:]
	subtableArrayOffset := int(binary.BigEndian.Uint32(strikeData))
	// tablesSize := binary.BigEndian.Uint32(strikeData[4:])
	numberSubtables := int(binary.BigEndian.Uint32(strikeData[8:]))
	// color ref
	out.hori = parseSbitLineMetrics(strikeData[16:])
	out.vert = parseSbitLineMetrics(strikeData[16+sbitLineMetricsLength:])
	strikeData = strikeData[16+2*sbitLineMetricsLength:]
	out.startGlyph = fonts.GlyphIndex(binary.BigEndian.Uint16(strikeData))
	out.endGlyph = fonts.GlyphIndex(binary.BigEndian.Uint16(strikeData[2:]))
	out.ppemX = strikeData[4]
	out.ppemY = strikeData[5]
	out.bitDepth = strikeData[6]
	out.flags = strikeData[7]

	if len(data) < subtableArrayOffset+numberSubtables*8 {
		return out, errors.New("invalid bitmap strike subtable (EOF)")
	}

	out.subTables = make([]indexSubTable, numberSubtables)
	for i := range out.subTables {
		out.subTables[i].firstGlyph = fonts.GlyphIndex(binary.BigEndian.Uint16(data[subtableArrayOffset+8*i:]))
		out.subTables[i].lastGlyph = fonts.GlyphIndex(binary.BigEndian.Uint16(data[subtableArrayOffset+8*i+2:]))
		additionalOffset := int(binary.BigEndian.Uint32(data[subtableArrayOffset+8*i+4:]))

		err = out.subTables[i].parseIndexSubTableData(data, subtableArrayOffset+additionalOffset)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

type sbitLineMetrics struct {
	ascender, descender                        int8
	widthMax                                   uint8
	caretSlopeNumerator, caretSlopeDenominator int8
	caretOffset                                int8
	minOriginSB                                int8
	minAdvanceSB                               int8
	maxBeforeBL                                int8
	minAfterBL                                 int8
}

// data must have suffisant length
func parseSbitLineMetrics(data []byte) (out sbitLineMetrics) {
	out.ascender = int8(data[0])
	out.descender = int8(data[1])
	out.widthMax = data[2]
	out.caretSlopeNumerator = int8(data[3])
	out.caretSlopeDenominator = int8(data[4])
	out.caretOffset = int8(data[5])
	out.minOriginSB = int8(data[6])
	out.minAdvanceSB = int8(data[7])
	out.maxBeforeBL = int8(data[8])
	out.minAfterBL = int8(data[9])
	return out
}

type indexSubTable struct {
	index                 interface{ isIndexSubTable() }
	firstGlyph, lastGlyph fonts.GlyphIndex // inclusive
	imageFormat           uint16
	imageDataOffset       uint32
}

func (sb *indexSubTable) parseIndexSubTableData(data []byte, offset int) error {
	if len(data) < offset+8 {
		return errors.New("invalid bitmap index subtable (EOF)")
	}
	data = data[offset:]
	format := binary.BigEndian.Uint16(data)
	sb.imageFormat = binary.BigEndian.Uint16(data[2:])
	sb.imageDataOffset = binary.BigEndian.Uint32(data[4:])
	numGlyphs := int(sb.lastGlyph-sb.firstGlyph) + 1
	var err error
	switch format {
	case 1:
		sb.index, err = parseIndexSubTable1(data[8:], numGlyphs)
	case 2:
		sb.index, err = parseIndexSubTable2(data[8:])
	case 3:
		sb.index, err = parseIndexSubTable3(data[8:], numGlyphs)
	case 4:
		sb.index, err = parseIndexSubTable4(data[8:])
	case 5:
		sb.index, err = parseIndexSubTable5(data[8:])
	default:
		return fmt.Errorf("unsupported bitmap index subtable format: %d", format)
	}

	return err
}

func (indexSubTable1) isIndexSubTable() {}
func (indexSubTable2) isIndexSubTable() {}
func (indexSubTable3) isIndexSubTable() {}
func (indexSubTable4) isIndexSubTable() {}
func (indexSubTable5) isIndexSubTable() {}

type indexSubTable1 []uint32

// data after the header
func parseIndexSubTable1(data []byte, numGlyphs int) (indexSubTable1, error) {
	if len(data) < (numGlyphs+1)*4 {
		return nil, errors.New("invalid bitmap index subtable format 1 (EOF)")
	}
	out := parseUint32s(data, numGlyphs+1)
	return out, nil
}

type indexSubTable2 struct {
	imageSize uint32
	metrics   glyphBigMetrics
}

func parseIndexSubTable2(data []byte) (out indexSubTable2, err error) {
	if len(data) < 4+glyphBigMetricsSize {
		return out, errors.New("invalid bitmap index subtable format 2 (EOF)")
	}
	out.imageSize = binary.BigEndian.Uint32(data)
	out.metrics = parseGlyphBigMetrics(data[4:])
	return out, nil
}

type indexSubTable3 []uint16

func parseIndexSubTable3(data []byte, numGlyphs int) (indexSubTable3, error) {
	return parseUint16s(data, numGlyphs+1)
}

type indexSubTable4 []struct {
	glyph  fonts.GlyphIndex
	offset uint16
}

func parseIndexSubTable4(data []byte) (out indexSubTable4, err error) {
	if len(data) < 4 {
		return out, errors.New("invalid bitmap index subtable format 4 (EOF)")
	}
	numGlyphs := int(binary.BigEndian.Uint32(data))
	if len(data) < 4+(numGlyphs+1)*4 {
		return out, errors.New("invalid bitmap index subtable format 4 (EOF)")
	}
	out = make(indexSubTable4, numGlyphs+1)
	for i := range out {
		out[i].glyph = fonts.GlyphIndex(binary.BigEndian.Uint16(data[4+4*i:]))
		out[i].offset = binary.BigEndian.Uint16(data[4+4*i+2:])
	}
	return out, nil
}

type indexSubTable5 struct {
	glyphs    []fonts.GlyphIndex // sorted
	imageSize uint32
	metrics   glyphBigMetrics
}

func parseIndexSubTable5(data []byte) (out indexSubTable5, err error) {
	if len(data) < 8+glyphBigMetricsSize {
		return out, errors.New("invalid bitmap index subtable format 5 (EOF)")
	}
	out.imageSize = binary.BigEndian.Uint32(data)
	out.metrics = parseGlyphBigMetrics(data[4:])
	numGlyphs := int(binary.BigEndian.Uint32(data[4+glyphBigMetricsSize:]))
	data = data[8+glyphBigMetricsSize:]
	if len(data) < 2*numGlyphs {
		return out, errors.New("invalid bitmap index subtable format 5 (EOF)")
	}
	out.glyphs = make([]fonts.GlyphIndex, numGlyphs)
	for i := range out.glyphs {
		out.glyphs[i] = fonts.GlyphIndex(binary.BigEndian.Uint16(data[2*i:]))
	}
	return out, nil
}

type glyphBigMetrics struct {
	height       uint8 // Number of rows of data.
	width        uint8 // Number of columns of data.
	horiBearingX int8  // Distance in pixels from the horizontal origin to the left edge of the bitmap.
	horiBearingY int8  // Distance in pixels from the horizontal origin to the top edge of the bitmap.
	horiAdvance  uint8 // Horizontal advance width in pixels.
	vertBearingX int8  // Distance in pixels from the vertical origin to the left edge of the bitmap.
	vertBearingY int8  // Distance in pixels from the vertical origin to the top edge of the bitmap.
	vertAdvance  uint8 // Vertical advance width in pixels.
}

const glyphBigMetricsSize = 8

// data must have a sufficient length
func parseGlyphBigMetrics(data []byte) (out glyphBigMetrics) {
	out.height = data[0]
	out.width = data[1]
	out.horiBearingX = int8(data[2])
	out.horiBearingY = int8(data[3])
	out.horiAdvance = data[4]
	out.vertBearingX = int8(data[5])
	out.vertBearingY = int8(data[6])
	out.vertAdvance = data[7]
	return out
}

// type glyphBigMetrics struct {}
