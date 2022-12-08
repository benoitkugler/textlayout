package bitmap

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/benoitkugler/textlayout/fonts"
)

// parser for .pcf bitmap fonts

// inspired by from https://github.com/stumpycr/pcf-parser
// and https://fontforge.org/docs/techref/pcf-format.html

const (
	properties = 1 << iota
	accelerators
	metrics
	bitmaps
	inkMetrics
	bdfEncodings
	sWidths
	glyphNames
	bdfAccelerators
)

const (
	defaultFormat      = 0x00000000
	inkbounds          = 0x00000200
	accelWInkbounds    = 0x00000100
	COMPRESSED_METRICS = 0x00000100

	// modifiers
	glyphPadMask = 3 << 0 /* See the bitmap table for explanation */
	byteMask     = 1 << 2 /* If set then Most Sig Byte First */
	bitMask      = 1 << 3 /* If set then Most Sig Bit First */
	scanUnitMask = 3 << 4 /* See the bitmap table for explanation */

	formatMask = ^uint32(0xFF) // keep the higher bits
)

// implementations limits to protect
// against malicious files
// see freetype/pcf for rationale
const (
	nbTablesMax     = 20
	nbPropertiesMax = 512
	nbMetricsMax    = 65536 // same for bitmaps and widths
)

const pcfHeader = "\x01fcp"

type Font struct {
	accelerator    *acceleratorTable // BDF accelerator if present, normal if not
	properties     propertiesTable
	bitmap         bitmapTable
	metrics        metricsTable // with length numGlyphs
	inkMetrics     metricsTable
	scalableWidths scalableWidthsTable
	names          namesTable
	cmap           encodingTable
}

func getOrder(format uint32) binary.ByteOrder {
	if format&byteMask != 0 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

type parser struct {
	data []byte
	pos  int
}

func (p *parser) u32(order binary.ByteOrder) (uint32, error) {
	if len(p.data) < p.pos+4 {
		return 0, errors.New("corrupted font file")
	}
	out := order.Uint32(p.data[p.pos:])
	p.pos += 4
	return out, nil
}

func (p *parser) u16(order binary.ByteOrder) (uint16, error) {
	if len(p.data) < p.pos+2 {
		return 0, errors.New("corrupted font file")
	}
	out := order.Uint16(p.data[p.pos:])
	p.pos += 2
	return out, nil
}

type tocEntry struct {
	// offset is adjusted to start right after the header+toc
	kind, format, size, offset uint32
}

// reads from r and returs the parsed entry
func parseTocEntries(r io.Reader) ([]tocEntry, error) {
	var buf [16]byte

	_, err := io.ReadFull(r, buf[:4])
	if err != nil {
		return nil, fmt.Errorf("corrupted toc entries: %s", err)
	}
	tableCount := binary.LittleEndian.Uint32(buf[:])
	// there can be most 9 tables
	if tableCount > 9 {
		return nil, fmt.Errorf("invalid .pcf file: %d tables", tableCount)
	}

	shift := 4 + 4 + tableCount*16 // header + tableCount + toc entries

	tocEntries := make([]tocEntry, tableCount)
	for i := range tocEntries {
		_, err = io.ReadFull(r, buf[:])
		if err != nil {
			return nil, fmt.Errorf("corrupted toc entry: %s", err)
		}
		tocEntries[i].kind = binary.LittleEndian.Uint32(buf[:])
		tocEntries[i].format = binary.LittleEndian.Uint32(buf[+4:])
		tocEntries[i].size = binary.LittleEndian.Uint32(buf[+8:])
		offset := binary.LittleEndian.Uint32(buf[+12:])
		if offset < shift {
			return nil, fmt.Errorf("corrupted toc offset: %d", offset)
		}
		tocEntries[i].offset = offset - shift
	}

	return tocEntries, nil
}

const propSize = 9

type prop struct {
	nameOffset   uint32
	isStringProp bool
	value        uint32 // offset or value
}

func (pr *parser) prop(order binary.ByteOrder) (prop, error) {
	if len(pr.data) < pr.pos+propSize {
		return prop{}, errors.New("invalid property")
	}
	var out prop
	out.nameOffset = order.Uint32(pr.data[pr.pos:])
	out.isStringProp = pr.data[pr.pos+4] == 1
	out.value = order.Uint32(pr.data[pr.pos+5:])
	pr.pos += propSize
	return out, nil
}

func getCString(data []byte, start uint32) (string, error) {
	if int(start) > len(data) {
		return "", errors.New("invalid offset in property")
	}
	// the srings are null terminated
	end := bytes.IndexByte(data[start:], 0)
	if end == -1 {
		return "", errors.New("invalid property")
	}
	name := string(data[start : int(start)+end]) // left the 0 off
	return name, nil
}

type propertiesTable map[string]Property

func (pr *parser) propertiesTable() (propertiesTable, error) {
	format, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	order := getOrder(format)

	nprops, err := pr.u32(order)
	if err != nil {
		return nil, err
	}
	if nprops > nbPropertiesMax {
		return nil, fmt.Errorf("number of properties (%d) exceeds implementation limit (%d)",
			nprops, nbPropertiesMax)
	}

	if len(pr.data) < pr.pos+int(nprops)*propSize {
		return nil, fmt.Errorf("invalid properties table: %d", nprops)
	}
	props := make([]prop, nprops)
	for i := range props {
		props[i], err = pr.prop(order)
		if err != nil {
			return nil, err
		}
	}

	if padding := int(nprops & 3); padding != 0 {
		pr.pos += 4 - padding // padding
	}

	stringsLength, err := pr.u32(order)
	if err != nil {
		return nil, err
	}

	if len(pr.data) < pr.pos+int(stringsLength) {
		return nil, fmt.Errorf("invalid properties table: %d", stringsLength)
	}
	rawData := pr.data[pr.pos : pr.pos+int(stringsLength)]

	out := make(propertiesTable, len(props))

	// we resolve the name properties and their values
	for _, prop := range props {
		name, err := getCString(rawData, prop.nameOffset)
		if err != nil {
			return nil, err
		}
		if prop.isStringProp {
			val, err := getCString(rawData, prop.value)
			if err != nil {
				return nil, err
			}
			out[name] = Atom(val)
		} else {
			out[name] = Int(prop.value)
		}
	}
	return out, nil
}

type bitmapTable struct {
	offsets []uint32
	data    []byte
}

func (p *parser) bitmap() (bitmapTable, error) {
	format, err := p.u32(binary.LittleEndian)
	if err != nil {
		return bitmapTable{}, err
	}
	if format&formatMask != defaultFormat {
		return bitmapTable{}, fmt.Errorf("invalid bitmap format: %d", format)
	}

	order := getOrder(format)

	count, err := p.u32(order)
	if err != nil {
		return bitmapTable{}, err
	}
	if count > nbMetricsMax {
		return bitmapTable{}, fmt.Errorf("number of glyphs (%d) exceeds implementation limit (%d)",
			count, nbMetricsMax)
	}

	if len(p.data) < p.pos+int(count)*4 {
		return bitmapTable{}, fmt.Errorf("invalid bitmap table")
	}
	offsets := make([]uint32, count)
	var maxOffset uint32
	for i := range offsets {
		offsets[i] = order.Uint32(p.data[p.pos+i*4:])
		if offsets[i] > maxOffset {
			maxOffset = offsets[i]
		}
	}
	p.pos += int(count) * 4

	var sizes [4]uint32
	if len(p.data) < p.pos+16 {
		return bitmapTable{}, fmt.Errorf("invalid bitmap table")
	}
	sizes[0] = order.Uint32(p.data[p.pos:])
	sizes[1] = order.Uint32(p.data[p.pos+4:])
	sizes[2] = order.Uint32(p.data[p.pos+8:])
	sizes[3] = order.Uint32(p.data[p.pos+12:])
	p.pos += 16

	bitmapLength := int(sizes[format&3])
	if len(p.data) < p.pos+bitmapLength {
		return bitmapTable{}, fmt.Errorf("invalid bitmap table (bitmapLength %d)", bitmapLength)
	}
	if maxOffset >= uint32(bitmapLength) {
		return bitmapTable{}, fmt.Errorf("invalid bitmap table (maxOffset %d)", maxOffset)
	}
	data := p.data[p.pos : p.pos+bitmapLength]
	p.pos += bitmapLength

	return bitmapTable{data: data, offsets: offsets}, nil
}

// we use int16 even for compressed for simplicity
type metric struct {
	leftSideBearing     int16
	rightSideBearing    int16
	characterWidth      int16
	characterAscent     int16
	characterDescent    int16
	characterAttributes uint16
}

const (
	metricCompressedSize   = 5
	metricUncompressedSize = 12
)

func (pr *parser) metric(compressed bool, order binary.ByteOrder) (metric, error) {
	var out metric
	if compressed {
		if len(pr.data) < pr.pos+metricCompressedSize {
			return out, fmt.Errorf("invalid compressed metric data")
		}
		out.leftSideBearing = int16(pr.data[pr.pos]) - 0x80
		out.rightSideBearing = int16(pr.data[pr.pos+1]) - 0x80
		out.characterWidth = int16(pr.data[pr.pos+2]) - 0x80
		out.characterAscent = int16(pr.data[pr.pos+3]) - 0x80
		out.characterDescent = int16(pr.data[pr.pos+4]) - 0x80
		pr.pos += metricCompressedSize
	} else {
		if len(pr.data) < pr.pos+metricUncompressedSize {
			return out, fmt.Errorf("invalid uncompressed metric data")
		}
		out.leftSideBearing = int16(order.Uint16(pr.data[pr.pos:]))
		out.rightSideBearing = int16(order.Uint16(pr.data[pr.pos+2:]))
		out.characterWidth = int16(order.Uint16(pr.data[pr.pos+4:]))
		out.characterAscent = int16(order.Uint16(pr.data[pr.pos+6:]))
		out.characterDescent = int16(order.Uint16(pr.data[pr.pos+8:]))
		out.characterAttributes = order.Uint16(pr.data[pr.pos+10:])
		pr.pos += metricUncompressedSize
	}
	return out, nil
}

type metricsTable []metric

func (pr *parser) metricTable() (metricsTable, error) {
	format, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	order := getOrder(format)

	compressed := format&formatMask == COMPRESSED_METRICS&formatMask
	var count int
	if compressed {
		c, er := pr.u16(order)
		count, err = int(c), er
	} else {
		c, er := pr.u32(order)
		count, err = int(c), er
	}
	if err != nil {
		return nil, err
	}

	if count > nbMetricsMax {
		return nil, fmt.Errorf("number of glyphs (%d) exceeds implementation limit (%d)",
			count, nbMetricsMax)
	}

	out := make(metricsTable, count)
	for i := range out {
		out[i], err = pr.metric(compressed, order)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (pr *parser) encodingTable() (encodingTable, error) {
	var out encodingTable
	format, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return out, err
	}

	if format&formatMask != defaultFormat {
		return out, fmt.Errorf("invalid encoding table format: %d", format)
	}

	order := getOrder(format)

	if len(pr.data) < pr.pos+10 {
		return out, fmt.Errorf("invalid encoding table")
	}

	// the values are actually byte
	minChar := order.Uint16(pr.data[pr.pos:])
	maxChar := order.Uint16(pr.data[pr.pos+2:])
	minByte := order.Uint16(pr.data[pr.pos+4:])
	maxByte := order.Uint16(pr.data[pr.pos+6:])

	if minChar > maxChar || maxChar > 0xFF || minByte > maxByte || maxByte > 0xFF {
		return out, fmt.Errorf("invalid encoding table limits: %d %d %d %d",
			minChar, maxChar, minByte, maxByte)
	}
	out.minChar = byte(minChar)
	out.maxChar = byte(maxChar)
	out.minByte = byte(minByte)
	out.maxByte = byte(maxByte)
	out.defaultChar = gid(order.Uint16(pr.data[pr.pos+8:]))
	pr.pos += 10

	count := int(maxByte-minByte+1) * int(maxChar-minChar+1) // care with overflows
	if len(pr.data) < pr.pos+2*count {
		return out, fmt.Errorf("invalid encoding table")
	}
	out.values = make([]gid, count)
	for i := range out.values {
		out.values[i] = gid(order.Uint16(pr.data[pr.pos+2*i:]))
	}
	pr.pos += 2 * count
	return out, nil
}

type acceleratorTable struct {
	// true if for all i:
	// max(metrics[i].rightSideBearing - metrics[i].characterWidth)  <= minbounds.leftSideBearing
	noOverlap       bool
	constantMetrics bool // Means the perchar field of the XFontStruct can be nil

	// constantMetrics true and forall characters:
	//  the left side bearing==0
	//  the right side bearing== the character's width
	//  the character's ascent==the font's ascent
	//  the character's descent==the font's descent
	terminalFont  bool
	constantWidth bool // monospace font like courier
	// Means that all inked bits are within the rectangle with x between [0,charwidth]
	//  and y between [-descent,ascent]. So no ink overlaps another char when drawing
	inkInside        bool
	inkMetrics       bool // true if the ink metrics differ from the metrics somewhere
	drawDirectionRTL bool // false:left to right, true:right to left
	// padding                      byte
	fontAscent                 int32
	fontDescent                int32
	maxOverlap                 int32
	minbounds, maxbounds       metric
	inkMinbounds, inkMaxbounds metric // If format is default, same as minbounds,maxbounds
}

func (pr *parser) accelerator() (*acceleratorTable, error) {
	format, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	order := getOrder(format)

	const length = 8 + 3*4 // before metrics
	if len(pr.data) < pr.pos+length {
		return nil, errors.New("invalid accelerator table")
	}

	var out acceleratorTable

	out.noOverlap = pr.data[pr.pos] == 1
	out.constantMetrics = pr.data[pr.pos+1] == 1
	out.terminalFont = pr.data[pr.pos+2] == 1
	out.constantWidth = pr.data[pr.pos+3] == 1
	out.inkInside = pr.data[pr.pos+4] == 1
	out.inkMetrics = pr.data[pr.pos+5] == 1
	out.drawDirectionRTL = pr.data[pr.pos+6] == 1
	// padding byte
	out.fontAscent = int32(order.Uint32(pr.data[pr.pos+8:]))
	out.fontDescent = int32(order.Uint32(pr.data[pr.pos+12:]))
	out.maxOverlap = int32(order.Uint32(pr.data[pr.pos+16:]))
	pr.pos += length

	out.minbounds, err = pr.metric(false, order)
	if err != nil {
		return nil, err
	}
	out.maxbounds, err = pr.metric(false, order)
	if err != nil {
		return nil, err
	}

	if format&formatMask == accelWInkbounds {
		out.inkMinbounds, err = pr.metric(false, order)
		if err != nil {
			return nil, err
		}
		out.inkMaxbounds, err = pr.metric(false, order)
		if err != nil {
			return nil, err
		}
	} else {
		out.inkMinbounds, out.inkMaxbounds = out.minbounds, out.maxbounds
	}

	return &out, nil
}

// byte offsets to bitmap data
type scalableWidthsTable []uint32

func (pr *parser) scalableWidths() (scalableWidthsTable, error) {
	format, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	order := getOrder(format)

	count, err := pr.u32(order)
	if err != nil {
		return nil, err
	}

	if count > nbMetricsMax {
		return nil, fmt.Errorf("number of glyphs (%d) exceeds implementation limit (%d)",
			count, nbMetricsMax)
	}

	if len(pr.data) < pr.pos+int(count)*4 {
		return nil, errors.New("invalid accelerator table")
	}

	out := make(scalableWidthsTable, count)
	for i := range out {
		out[i] = order.Uint32(pr.data[pr.pos+4*i:])
	}
	pr.pos += int(count) * 4

	return out, nil
}

type namesTable []string // indexed by the glyph index

func (pr *parser) names() (namesTable, error) {
	format, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	order := getOrder(format)

	count, err := pr.u32(order)
	if err != nil {
		return nil, err
	}

	if count > nbMetricsMax {
		return nil, fmt.Errorf("number of glyphs (%d) exceeds implementation limit (%d)",
			count, nbMetricsMax)
	}

	if len(pr.data) < pr.pos+int(count)*4 {
		return nil, errors.New("invalid names table")
	}

	offsets := make([]uint32, count)
	for i := range offsets {
		offsets[i] = order.Uint32(pr.data[pr.pos+4*i:])
	}
	pr.pos += int(count) * 4

	stringSize, err := pr.u32(order)
	if err != nil {
		return nil, err
	}

	if len(pr.data) < pr.pos+int(stringSize) {
		return nil, errors.New("invalid names table")
	}
	stringData := pr.data[pr.pos : pr.pos+int(stringSize)]

	out := make(namesTable, count)
	for i, offset := range offsets {
		out[i], err = getCString(stringData, offset)
		if err != nil {
			return nil, err
		}
	}
	pr.pos += int(stringSize)

	return out, nil
}

// checking the coherence between tables
func (f *Font) validate() error {
	nbGlyphs := len(f.bitmap.offsets)
	if L := len(f.metrics); L != nbGlyphs {
		return fmt.Errorf("invalid number of metrics: expected %d, got %d", nbGlyphs, L)
	}
	if f.accelerator == nil {
		return fmt.Errorf("missing accelerator table")
	}
	if L := len(f.scalableWidths); f.scalableWidths != nil && L != nbGlyphs {
		return fmt.Errorf("invalid number of widths: expected %d, got %d", nbGlyphs, L)
	}
	if L := len(f.names); f.names != nil && L != nbGlyphs {
		return fmt.Errorf("invalid number of names: expected %d, got %d", nbGlyphs, L)
	}
	if L := len(f.inkMetrics); f.inkMetrics != nil && L != nbGlyphs {
		return fmt.Errorf("invalid number of inkMetrics: expected %d, got %d", nbGlyphs, L)
	}
	return nil
}

func newParser(file fonts.Resource) (io.Reader, []tocEntry, error) {
	_, err := file.Seek(0, io.SeekStart) // file might have been used before
	if err != nil {
		return nil, nil, err
	}

	var r io.Reader
	// pcf file are often compressed so we try gzip
	r, err = gzip.NewReader(file)
	if err != nil { // not a gzip file: read from the plain file
		// gzip has read some bytes
		_, _ = file.Seek(0, io.SeekStart)
		r = file
	}
	// check the start of the file before reading all
	var headerBuf [4]byte
	if io.ReadFull(r, headerBuf[:]); string(headerBuf[:]) != pcfHeader {
		return nil, nil, errors.New("not a PCF file")
	}

	// we have a .pcf; read table of contents
	toc, err := parseTocEntries(r)
	if err != nil {
		return nil, nil, err
	}
	return r, toc, nil
}

// Parse parse a .pcf font file.
func Parse(file fonts.Resource) (*Font, error) {
	r, tocEntries, err := newParser(file)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("can't load font file: %s", err)
	}

	// we have read the header and the toc already
	pr := parser{data: data}

	var (
		out      Font
		bdfAccel *acceleratorTable
		encoding encodingTable
	)
	for _, tc := range tocEntries {
		// seek: tc.offset has been adjusted to match the state of `r`
		pr.pos = int(tc.offset)
		switch tc.kind {
		case properties:
			out.properties, err = pr.propertiesTable()
		case bitmaps:
			out.bitmap, err = pr.bitmap()
		case metrics:
			out.metrics, err = pr.metricTable()
		case inkMetrics:
			out.inkMetrics, err = pr.metricTable()
		case bdfEncodings:
			encoding, err = pr.encodingTable()
		case accelerators:
			out.accelerator, err = pr.accelerator()
		case bdfAccelerators:
			bdfAccel, err = pr.accelerator()
		case sWidths:
			out.scalableWidths, err = pr.scalableWidths()
		case glyphNames:
			out.names, err = pr.names()

		}
		if err != nil {
			return nil, err
		}
	}

	if bdfAccel != nil {
		out.accelerator = bdfAccel
	}

	err = out.concludeParsing(encoding)

	return &out, err
}
