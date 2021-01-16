package truetype

import (
	"errors"
)

var (
	errInvalidKernTable     = errors.New("invalid kern table")
	errUnsupportedKernTable = errors.New("unsupported kern table")
)

// Kerns store a compact form of the (horizontal) kerning
// values.
type Kerns interface {
	// KernPair return the kern value for the given pair, if any.
	// The value is expressed in glyph units and
	// is negative when glyphs should be closer.
	KernPair(left, right GlyphIndex) (int16, bool)
	// Size returns the number of kerning pairs
	Size() int
}

// key is left << 16 + right
type simpleKerns map[uint32]int16

func (s simpleKerns) KernPair(left, right GlyphIndex) (int16, bool) {
	out, has := s[uint32(left)<<16|uint32(right)]
	return out, has
}

func (s simpleKerns) Size() int { return len(s) }

// assume non overlapping kerns, otherwise the return value is undefined
type kernUnions []Kerns

func (ks kernUnions) KernPair(left, right GlyphIndex) (int16, bool) {
	for _, k := range ks {
		out, has := k.KernPair(left, right)
		if has {
			return out, true
		}
	}
	return 0, false
}

func (ks kernUnions) Size() int {
	out := 0
	for _, k := range ks {
		out += k.Size()
	}
	return out
}

func parseKernTable(input []byte) (simpleKerns, error) {
	const headerSize = 4
	if len(input) < headerSize {
		return nil, errInvalidKernTable
	}
	version := be.Uint16(input[:2])
	numTables := be.Uint16(input[2:4])

	// like golang/x, we dont support apple version
	if version != 0 {
		return nil, errUnsupportedKernTable
	}

	input = input[4:]
	out := simpleKerns{}
	for i := uint16(0); i < numTables; i++ {
		nbRead, err := parseKernSubtable(input, out)
		if err != nil {
			return nil, err
		}
		input = input[nbRead:]
	}
	return out, nil
}

// returns the number of bytes read
func parseKernSubtable(input []byte, out simpleKerns) (int, error) {
	const subtableHeaderSize = 6
	if len(input) < subtableHeaderSize {
		return 0, errInvalidKernTable
	}
	// skip version and length
	format, coverage := input[4], input[5]
	if coverage != 0x01 {
		// We only support horizontal kerning.
		return 0, errUnsupportedKernTable
	}
	if format != 0 {
		// following other implementation
		return 0, errUnsupportedKernTable
	}

	read, err := parseKernFormat0(input[6:], out)
	return subtableHeaderSize + read, err
}

func parseKernFormat0(input []byte, out simpleKerns) (int, error) {
	const headerSize, entrySize = 8, 6
	if len(input) < headerSize {
		return 0, errInvalidKernTable
	}
	numPairs := be.Uint16(input)

	// skip searchRange , entrySelector , rangeShift

	subtableProperSize := headerSize + entrySize*int(numPairs)
	if len(input) < subtableProperSize {
		return 0, errInvalidKernTable
	}

	// we opt for a brute force approach:
	// we could instead store a slice of {left, right, value} to reduce
	// memory usage
	for i := 0; i < int(numPairs); i++ {
		left := GlyphIndex(be.Uint16(input[entrySize*i:]))
		right := GlyphIndex(be.Uint16(input[entrySize*i+2:]))
		out[uint32(left)<<16|uint32(right)] = int16(be.Uint16(input[entrySize*i+4:]))
	}
	return subtableProperSize, nil
}
