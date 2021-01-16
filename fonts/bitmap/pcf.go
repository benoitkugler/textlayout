package bitmap

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
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

const header = "\x01fcp"

type Font struct {
	properties          propertiesTable
	bitmap              bitmapTable
	metrics, inkMetrics metricsTable
	encoding            encodingTable
	accelerator         *acceleratorTable // BDF accelerator if present, normal if not
	scalableWidths      scalableWidthsTable
	names               namesTable
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
	kind, format, size, offset uint32
}

func (p *parser) tocEntry() (out tocEntry, err error) {
	if len(p.data) < p.pos+16 {
		return out, errors.New("corrupted toc entry")
	}
	out.kind = binary.LittleEndian.Uint32(p.data[p.pos:])
	out.format = binary.LittleEndian.Uint32(p.data[p.pos+4:])
	out.size = binary.LittleEndian.Uint32(p.data[p.pos+8:])
	out.offset = binary.LittleEndian.Uint32(p.data[p.pos+12:])
	p.pos += 16
	return out, nil
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
		return nil, errors.New("invalid properties table")
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
		return nil, errors.New("invalid properties table")
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
	for i := range offsets {
		offsets[i] = order.Uint32(p.data[p.pos+i*4:])
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
		return bitmapTable{}, fmt.Errorf("invalid bitmap table")
	}
	data := p.data[p.pos : p.pos+bitmapLength]
	p.pos += bitmapLength

	return bitmapTable{data: data, offsets: offsets}, nil
}

//   class BitmapTable
//     getter bitmaps : Array(Bytes)
//     getter padding_bytes : Int32
//     getter data_bytes : Int32

//     # TODO: Raise if format != DEFAULT
//     def initialize(io)
//       format = io.read_bytes(Int32, IO::ByteFormat::LittleEndian)

//       # 0 => byte (8bit), 1 => short (16bit), 2 => int (32bit)
//       # TODO: What is this needed for?
//       glyph_pad = format & 3
//       @padding_bytes = glyph_pad == 0 ? 1 : glyph_pad * 2

//       byte_mask = (format & 4) != 0 # set => most significant byte first
//       bit_mask = (format & 8) != 0  # set => most significant bit first

//       # 0 => byte (8bit), 1 => short (16bit), 2 => int (32bit)
//       scan_unit = (format >> 4) & 3
//       @data_bytes = scan_unit == 0 ? 1 : scan_unit * 2

//       puts "Unsupported bit_mask: #{bit_mask}" unless bit_mask
//       byte_format = byte_mask ? IO::ByteFormat::BigEndian : IO::ByteFormat::BigEndian

//       # :compressed_metrics is equiv. to :accel_w_inkbounds
//       main_format = [:default, :inkbounds, :compressed_metrics][format >> 8]

//       glyph_count = io.read_bytes(Int32, byte_format)
//       offsets = [] of Int32

//       glyph_count.times do
//         offsets << io.read_bytes(Int32, byte_format)
//       end

//       bitmap_sizes = [] of Int32
//       4.times do
//         bitmap_sizes << io.read_bytes(Int32, byte_format)
//       end

//       @bitmaps = [] of Bytes

//       slice = Bytes.new(bitmap_sizes[glyph_pad])
//       read = io.read_fully(slice)

//       raise "Failed to read bitmap data" if bitmap_sizes[glyph_pad] != read

//       offsets.each do |off|
//         @bitmaps << (slice + off)
//       end

//       # bitmap_data = io.pos
//       # offsets.each do |off|
//       #   size = bitmap_sizes[glyph_pad] / glyph_count
//       #   slice = Bytes.new(size)

//       #   io.seek(bitmap_data + off) do
//       #     n_read = io.read(slice)
//       #     raise "Failed to read bitmap data" if n_read != size
//       #     @bitmaps << slice
//       #   end
//       # end
//     end
//   end

// we use int16 even for compressed for simplicity
type metric struct {
	leftSidedBearing    int16
	rightSidedBearing   int16
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
		out.leftSidedBearing = int16(pr.data[pr.pos] - 0x80)
		out.rightSidedBearing = int16(pr.data[pr.pos+1] - 0x80)
		out.characterWidth = int16(pr.data[pr.pos+2] - 0x80)
		out.characterAscent = int16(pr.data[pr.pos+3] - 0x80)
		out.characterDescent = int16(pr.data[pr.pos+4] - 0x80)
		pr.pos += metricCompressedSize
	} else {
		if len(pr.data) < pr.pos+metricUncompressedSize {
			return out, fmt.Errorf("invalid uncompressed metric data")
		}
		out.leftSidedBearing = int16(order.Uint16(pr.data[pr.pos:]))
		out.rightSidedBearing = int16(order.Uint16(pr.data[pr.pos+2:]))
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

type encodingTable map[uint16]uint16

func (pr *parser) encodingTable() (encodingTable, error) {
	format, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	if format&formatMask != defaultFormat {
		return nil, fmt.Errorf("invalid encoding table format: %d", format)
	}

	order := getOrder(format)

	if len(pr.data) < pr.pos+10 {
		return nil, fmt.Errorf("invalid encoding table")
	}

	minChar := order.Uint16(pr.data[pr.pos:])
	maxChar := order.Uint16(pr.data[pr.pos+2:])
	minByte := order.Uint16(pr.data[pr.pos+4:])
	maxByte := order.Uint16(pr.data[pr.pos+6:])
	defaultChar := order.Uint16(pr.data[pr.pos+8:])
	pr.pos += 10

	count := int(maxByte-minByte+1) * int(maxChar-minChar+1)
	if len(pr.data) < pr.pos+2*count {
		return nil, fmt.Errorf("invalid encoding table")

	}
	out := make(encodingTable, count)

	for ma := minByte; ma <= maxByte; ma++ {
		for mi := minChar; mi <= maxChar; mi++ {
			value := order.Uint16(pr.data[pr.pos:])
			pr.pos += 2

			full := ma<<8 | mi
			if value != 0xffff {
				out[full] = value
			} else {
				out[full] = defaultChar
			}
		}
	}

	return out, nil
}

type acceleratorTable struct {
	// true if for all i:
	// max(metrics[i].rightSideBearing - metrics[i].characterWidth)  <= minbounds.leftSideBearing
	noOverlap       bool
	constantMetrics bool // Means the perchar field of the XFontStruct can be nil

	// constantMetrics true and forall characters:
	//      the left side bearing==0
	//      the right side bearing== the character's width
	//      the character's ascent==the font's ascent
	//      the character's descent==the font's descent
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
func (f Font) validate() error {
	nbGlyphs := len(f.bitmap.offsets)
	if L := len(f.scalableWidths); f.scalableWidths != nil && L != nbGlyphs {
		return fmt.Errorf("invalid number of widths: expected %d, got %d", nbGlyphs, L)
	}
	if L := len(f.names); f.names != nil && L != nbGlyphs {
		return fmt.Errorf("invalid number of names: expected %d, got %d", nbGlyphs, L)
	}
	if L := len(f.metrics); f.metrics != nil && L != nbGlyphs {
		return fmt.Errorf("invalid number of metrics: expected %d, got %d", nbGlyphs, L)
	}
	if f.accelerator == nil {
		return fmt.Errorf("missing accelerator table")
	}
	return nil
}

func parse(data []byte) (*Font, error) {
	if len(data) < 4 || string(data[0:4]) != header {
		return nil, errors.New("not a PCF file")
	}

	pr := parser{data: data, pos: 4}
	tableCount, err := pr.u32(binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	tocEntries := make([]tocEntry, tableCount)
	for i := range tocEntries {
		tocEntries[i], err = pr.tocEntry()
		if err != nil {
			return nil, err
		}
	}

	var (
		out      Font
		bdfAccel *acceleratorTable
	)
	for _, tc := range tocEntries {
		pr.pos = int(tc.offset) // seek
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
			out.encoding, err = pr.encodingTable()
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

	err = out.validate()
	return &out, err
}

//       bitmap_table = nil
//       metrics_table = nil
//       encoding_table = nil

//       @encoding = {} of Int16 => Int16

//       tocEntries.each do |entry|
//         io.seek(entry.offset)

//         case entry.type
//           # when TableType::properties
//           #   @properties_table = PropertiesTable.new(io)
//           when TableType::bitmaps
//             bitmap_table = BitmapTable.new(io)
//           when TableType::metrics
//             metrics_table = MetricsTable.new(io)
//           when TableType::bdfEncodings
//             encoding_table = EncodingTable.new(io)
//         end
//       end

//       raise "Could not find a bitmap table" if bitmap_table.nil?
//       raise "Could not find a metrics table" if metrics_table.nil?

//       bitmaps = bitmap_table.bitmaps
//       metrics = metrics_table.metrics

//       if bitmaps.size != metrics.size
//         raise "Bitmap and metrics tables are not of the same size"
//       end

//       unless encoding_table.nil?
//         @encoding = encoding_table.encoding
//       end

//       @characters = [] of Character

//       @max_ascent = 0
//       @max_descent = 0

//       bitmaps.each_with_index do |bitmap, i|
//         metric = metrics[i]

//         if metric.characterAscent > @max_ascent
//           @max_ascent += metric.characterAscent
//         end

//         if metric.characterDescent > @max_descent
//           @max_descent += metric.characterDescent
//         end

//         char = Character.new(
//           bitmap,
//           metric.characterWidth,
//           metric.characterAscent,
//           metric.characterDescent,
//           metric.leftSidedBearing,
//           metric.rightSidedBearing,
//           bitmap_table.data_bytes,
//           bitmap_table.padding_bytes,
//         )

//         @characters << char
//       end
//     end

//     def lookup(str : String)
//       str.chars.map { |c| lookup(c) }
//     end

//     def lookup(char : Char)
//       lookup(char.ord)
//     end

//     def lookup(char)
//       @characters[@encoding[char]]
//     end
//   end

//   class Character
//     getter width : Int16
//     getter ascent : Int16
//     getter descent : Int16

//     getter leftSidedBearing : Int16
//     getter rightSidedBearing : Int16

//     @padding_bytes : Int32
//     @data_bytes : Int32
//     @bytes : Bytes

//     @bytes_per_row : Int32

//     def initialize(@bytes, @width, @ascent, @descent, @leftSidedBearing, @rightSidedBearing, @data_bytes, @padding_bytes)
//       @bytes_per_row = [(@width / 8).to_i32, 1].max

//       # Pad as needed
//       if (@bytes_per_row % @padding_bytes) != 0
//         @bytes_per_row += @padding_bytes - (@bytes_per_row % @padding_bytes)
//       end

//       # TODO: Is this last row relevant?
//       @bytes_per_row = [@bytes_per_row, @data_bytes].max

//       # needed = @bytes_per_row * (@ascent + @descent)
//       # got = @bytes.size
//     end

//     def get(x, y)
//       unless 0 <= x < @width
//         raise "Invalid x value: #{x}, must be in range (0..#{@width})"
//       end

//       unless 0 <= y < (@ascent + @descent)
//         raise "Invalid y value: #{y}, must be in range (0..#{@ascent + @descent})"
//       end

//       index = x // 8 + @bytes_per_row * y
//       shift = 7 - (x % 8)

//       if index < @bytes.size
//         @bytes[index] & (1 << (7 - (x % 8))) != 0
//       else
//         true
//       end
//     end
//   end
