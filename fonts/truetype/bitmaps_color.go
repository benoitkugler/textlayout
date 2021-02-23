package truetype

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/benoitkugler/textlayout/fonts"
)

// group the location (cblc) and the data (cbdt)
type bitmapTable []bitmapSize

// return nil if the table is empty
func (t bitmapTable) chooseStrike(xPpem, yPpem uint16) *bitmapSize {
	if len(t) == 0 {
		return nil
	}
	request := maxu16(xPpem, yPpem)
	if request == 0 {
		request = math.MaxUint16 // choose largest strike
	}
	var (
		bestIndex = 0
		bestPpem  = maxu16(t[0].ppemX, t[0].ppemY)
	)
	for i, s := range t {
		ppem := maxu16(s.ppemX, s.ppemY)
		if request <= ppem && ppem < bestPpem || request > bestPpem && ppem > bestPpem {
			bestIndex = i
			bestPpem = ppem
		}
	}
	return &t[bestIndex]
}

func parseTableBitmap(locationTable, rawDataTable []byte) (bitmapTable, error) {
	if len(locationTable) < 8 {
		return nil, errors.New("invalid bitmap location table (EOF)")
	}
	numSizes := int(binary.BigEndian.Uint32(locationTable[4:]))
	if len(locationTable) < 8+numSizes*bitmapSizeLength {
		return nil, errors.New("invalid bitmap location table (EOF)")
	}
	out := make(bitmapTable, numSizes) // guarded by the check above
	var err error
	for i := range out {
		out[i], err = parseBitmapSize(locationTable, 8+i*bitmapSizeLength, rawDataTable)
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
	ppemX, ppemY         uint16
	bitDepth             uint8
	flags                uint8
}

const (
	sbitLineMetricsLength = 12
	bitmapSizeLength      = 24 + 2*sbitLineMetricsLength
)

// length as been checked
func parseBitmapSize(data []byte, offset int, rawImageData []byte) (out bitmapSize, err error) {
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
	out.ppemX = uint16(strikeData[4])
	out.ppemY = uint16(strikeData[5])
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

		err = out.subTables[i].parseIndexSubTableData(data, subtableArrayOffset+additionalOffset, rawImageData)
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
	// imageDataOffset       uint32
}

func (sb *indexSubTable) parseIndexSubTableData(data []byte, offset int, rawData []byte) error {
	if len(data) < offset+8 {
		return errors.New("invalid bitmap index subtable (EOF)")
	}
	data = data[offset:]
	format := binary.BigEndian.Uint16(data)
	sb.imageFormat = binary.BigEndian.Uint16(data[2:])
	imageDataOffset := int(binary.BigEndian.Uint32(data[4:]))
	numGlyphs := int(sb.lastGlyph-sb.firstGlyph) + 1

	if len(rawData) < imageDataOffset {
		return errors.New("invalid bitmap data table (EOF)")
	}
	imageData := rawData[imageDataOffset:]

	var err error
	switch format {
	case 1:
		sb.index, err = parseIndexSubTable1(sb.imageFormat, imageData, data[8:], numGlyphs)
	case 2:
		sb.index, err = parseIndexSubTable2(sb.imageFormat, imageData, data[8:], numGlyphs)
	case 3:
		sb.index, err = parseIndexSubTable3(sb.imageFormat, imageData, data[8:], numGlyphs)
	case 4:
		sb.index, err = parseIndexSubTable4(sb.imageFormat, imageData, data[8:])
	case 5:
		sb.index, err = parseIndexSubTable5(sb.imageFormat, imageData, data[8:])
	default:
		return fmt.Errorf("unsupported bitmap index subtable format: %d", format)
	}

	return err
}

func (indexSubTable1And3) isIndexSubTable() {}
func (indexSubTable2) isIndexSubTable()     {}
func (indexSubTable4) isIndexSubTable()     {}
func (indexSubTable5) isIndexSubTable()     {}

type indexSubTable1And3 []bitmapData

// data starts after the header, imageData at the image
func parseIndexSubTable1(imageFormat uint16, imageData, data []byte, numGlyphs int) (indexSubTable1And3, error) {
	if len(data) < (numGlyphs+1)*4 {
		return nil, errors.New("invalid bitmap index subtable format 1 (EOF)")
	}
	offsets := parseUint32s(data, numGlyphs+1)
	out := make([]bitmapData, numGlyphs)
	var err error
	for i := range out {
		if offsets[i] == offsets[i+1] {
			continue
		}
		out[i], err = parseBitmapData(imageData, offsets[i], offsets[i+1], imageFormat)
		if err != nil {
			return nil, fmt.Errorf("invalid bitmap index format 1: %s", err)
		}
	}
	return out, nil
}

type indexSubTable2 struct {
	glyphs  []bitmapData
	metrics bigGlyphMetrics
}

func parseIndexSubTable2(imageFormat uint16, imageData, data []byte, numGlyphs int) (out indexSubTable2, err error) {
	if len(data) < 4+bigGlyphMetricsSize {
		return out, errors.New("invalid bitmap index subtable format 2 (EOF)")
	}
	imageSize := binary.BigEndian.Uint32(data)
	out.metrics = parseBigGlyphMetrics(data[4:])
	out.glyphs = make([]bitmapData, numGlyphs)
	for i := range out.glyphs {
		out.glyphs[i], err = parseBitmapData(imageData, imageSize*uint32(i), (imageSize+1)*uint32(i), imageFormat)
		if err != nil {
			return out, fmt.Errorf("invalid bitmap index format 2: %s", err)
		}
	}
	return out, nil
}

func parseIndexSubTable3(imageFormat uint16, imageData, data []byte, numGlyphs int) (indexSubTable1And3, error) {
	offsets, err := parseUint16s(data, numGlyphs+1)
	if err != nil {
		return nil, err
	}
	out := make([]bitmapData, numGlyphs)
	for i := range out {
		if offsets[i] == offsets[i+1] {
			continue
		}
		out[i], err = parseBitmapData(imageData, uint32(offsets[i]), uint32(offsets[i+1]), imageFormat)
		if err != nil {
			return nil, fmt.Errorf("invalid bitmap index format 3: %s", err)
		}
	}
	return out, err
}

type indexedBitmapGlyph struct {
	data  bitmapData
	glyph fonts.GlyphIndex
}

type indexSubTable4 []indexedBitmapGlyph

func parseIndexSubTable4(imageFormat uint16, imageData, data []byte) (out indexSubTable4, err error) {
	if len(data) < 4 {
		return out, errors.New("invalid bitmap index subtable format 4 (EOF)")
	}
	numGlyphs := int(binary.BigEndian.Uint32(data))
	if len(data) < 4+(numGlyphs+1)*4 {
		return out, errors.New("invalid bitmap index subtable format 4 (EOF)")
	}
	out = make(indexSubTable4, numGlyphs)
	var currentOffset, nextOffset uint32
	nextOffset = uint32(binary.BigEndian.Uint16(data[4+2:]))
	for i := range out {
		out[i].glyph = fonts.GlyphIndex(binary.BigEndian.Uint16(data[4+4*i:]))
		currentOffset = nextOffset
		nextOffset = uint32(binary.BigEndian.Uint16(data[4+4*(i+1)+2:]))
		out[i].data, err = parseBitmapData(imageData, currentOffset, nextOffset, imageFormat)
		if err != nil {
			return nil, fmt.Errorf("invalid bitmap index format 4: %s", err)
		}
	}
	return out, nil
}

type indexSubTable5 struct {
	glyphs  []indexedBitmapGlyph // sorted by glyph index
	metrics bigGlyphMetrics
}

func parseIndexSubTable5(imageFormat uint16, imageData, data []byte) (out indexSubTable5, err error) {
	if len(data) < 8+bigGlyphMetricsSize {
		return out, errors.New("invalid bitmap index subtable format 5 (EOF)")
	}
	imageSize := binary.BigEndian.Uint32(data)
	out.metrics = parseBigGlyphMetrics(data[4:])
	numGlyphs := int(binary.BigEndian.Uint32(data[4+bigGlyphMetricsSize:]))
	data = data[8+bigGlyphMetricsSize:]
	if len(data) < 2*numGlyphs {
		return out, errors.New("invalid bitmap index subtable format 5 (EOF)")
	}
	out.glyphs = make([]indexedBitmapGlyph, numGlyphs)
	for i := range out.glyphs {
		out.glyphs[i].glyph = fonts.GlyphIndex(binary.BigEndian.Uint16(data[2*i:]))
		out.glyphs[i].data, err = parseBitmapData(imageData, imageSize*uint32(i), (imageSize+1)*uint32(i), imageFormat)
		if err != nil {
			return out, fmt.Errorf("invalid bitmap index format 5: %s", err)
		}
	}
	return out, nil
}

type smallGlyphMetrics struct {
	height       uint8 // Number of rows of data.
	width        uint8 // Number of columns of data.
	horiBearingX int8  // Distance in pixels from the horizontal origin to the left edge of the bitmap.
	horiBearingY int8  // Distance in pixels from the horizontal origin to the top edge of the bitmap.
	horiAdvance  uint8 // Horizontal advance width in pixels.
}

type bigGlyphMetrics struct {
	smallGlyphMetrics

	vertBearingX int8  // Distance in pixels from the vertical origin to the left edge of the bitmap.
	vertBearingY int8  // Distance in pixels from the vertical origin to the top edge of the bitmap.
	vertAdvance  uint8 // Vertical advance width in pixels.
}

const (
	smallGlyphMetricsSize = 5
	bigGlyphMetricsSize   = smallGlyphMetricsSize + 3
)

// data must have a sufficient length
func parseSmallGlyphMetrics(data []byte) (out smallGlyphMetrics) {
	out.height = data[0]
	out.width = data[1]
	out.horiBearingX = int8(data[2])
	out.horiBearingY = int8(data[3])
	out.horiAdvance = data[4]
	return out
}

// data must have a sufficient length
func parseBigGlyphMetrics(data []byte) (out bigGlyphMetrics) {
	out.smallGlyphMetrics = parseSmallGlyphMetrics(data)
	out.vertBearingX = int8(data[5])
	out.vertBearingY = int8(data[6])
	out.vertAdvance = data[7]
	return out
}

// --------------------- actual bitmap data ---------------------

type bitmapData interface{ isBitmapData() }

func (bitmapDataFormat2) isBitmapData()  {}
func (bitmapDataFormat5) isBitmapData()  {}
func (bitmapDataFormat17) isBitmapData() {}
func (bitmapDataFormat18) isBitmapData() {}
func (bitmapDataFormat19) isBitmapData() {}

func parseBitmapData(imageData []byte, start, end uint32, format uint16) (bitmapData, error) {
	if len(imageData) < int(end) || start > end {
		return nil, errors.New("invalid bitmap data table (EOF)")
	}
	imageData = imageData[start:end]
	switch format {
	case 1, 4, 6, 7, 8, 9:
		return nil, fmt.Errorf("valid but currently not implemented bitmap image format: %d", format)
	case 2:
		return parseBitmapDataFormat2(imageData)
	case 5:
		return parseBitmapDataFormat5(imageData)
	case 17:
		return parseBitmapDataFormat17(imageData)
	case 18:
		return parseBitmapDataFormat18(imageData)
	case 19:
		return parseBitmapDataFormat19(imageData)
	default:
		return nil, fmt.Errorf("unsupported bitmap image format: %d", format)
	}
}

// small metrics, bit-aligned data
type bitmapDataFormat2 struct {
	data         []byte            // Bit-aligned bitmap data
	glyphMetrics smallGlyphMetrics // Metrics information for the glyph
}

// data start at the image data
func parseBitmapDataFormat2(data []byte) (out bitmapDataFormat2, err error) {
	if len(data) < smallGlyphMetricsSize {
		return out, errors.New("invalid bitmap data format 2 (EOF)")
	}
	out.glyphMetrics = parseSmallGlyphMetrics(data)
	out.data = data[smallGlyphMetricsSize:]
	return out, nil
}

// Format 5: metrics in CBLC table, bit-aligned image data only
type bitmapDataFormat5 []byte // Bit-aligned bitmap data

// data start at the image data
func parseBitmapDataFormat5(data []byte) (out bitmapDataFormat5, err error) {
	return data, nil
}

// small metrics, PNG image data
type bitmapDataFormat17 struct {
	data         []byte            // Raw PNG data
	glyphMetrics smallGlyphMetrics // Metrics information for the glyph
}

// data start at the image data
func parseBitmapDataFormat17(data []byte) (out bitmapDataFormat17, err error) {
	if len(data) < smallGlyphMetricsSize+4 {
		return out, errors.New("invalid bitmap data format 17 (EOF)")
	}
	out.glyphMetrics = parseSmallGlyphMetrics(data)
	length := int(binary.BigEndian.Uint32(data[smallGlyphMetricsSize:]))
	if len(data) < smallGlyphMetricsSize+4+length {
		return out, errors.New("invalid bitmap data format 17 (EOF)")
	}
	out.data = data[smallGlyphMetricsSize+4 : smallGlyphMetricsSize+4+length]
	return out, nil
}

// big metrics, PNG image data
type bitmapDataFormat18 struct {
	data         []byte          //	Raw PNG data
	glyphMetrics bigGlyphMetrics //	Metrics information for the glyph
}

// data start at the image data
func parseBitmapDataFormat18(data []byte) (out bitmapDataFormat18, err error) {
	if len(data) < bigGlyphMetricsSize+4 {
		return out, errors.New("invalid bitmap data format 18 (EOF)")
	}
	out.glyphMetrics = parseBigGlyphMetrics(data)
	length := int(binary.BigEndian.Uint32(data[bigGlyphMetricsSize:]))
	if len(data) < bigGlyphMetricsSize+4+length {
		return out, errors.New("invalid bitmap data format 18 (EOF)")
	}
	out.data = data[bigGlyphMetricsSize+4 : bigGlyphMetricsSize+4+length]
	return out, nil
}

// Format 19: metrics in CBLC table, PNG image data
type bitmapDataFormat19 []byte // Raw PNG data

// data start at the image data
func parseBitmapDataFormat19(data []byte) (out bitmapDataFormat19, err error) {
	if len(data) < 4 {
		return out, errors.New("invalid bitmap data format 19 (EOF)")
	}
	length := int(binary.BigEndian.Uint32(data))
	if len(data) < 4+length {
		return out, errors.New("invalid bitmap data format 19 (EOF)")
	}
	return data[4 : 4+length], nil
}
