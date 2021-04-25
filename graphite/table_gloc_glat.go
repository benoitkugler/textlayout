package graphite

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"github.com/benoitkugler/textlayout/fonts/binaryreader"
)

type attributSetEntry struct {
	attributes []int16
	firstKey   uint16
}

// sorted by `firstKey`; firstKey + len(attributes) <= nextFirstKey
type attributeSet []attributSetEntry

func (as attributeSet) get(key uint16) (int16, bool) {
	// binary search
	for i, j := 0, len(as); i < j; {
		h := i + (j-i)/2
		entry := as[h]
		if key < entry.firstKey {
			j = h
		} else if entry.firstKey+uint16(len(entry.attributes))-1 < key {
			i = h + 1
		} else {
			return entry.attributes[key-entry.firstKey], true
		}
	}
	return 0, false
}

type tableGlat []attributeSet // with len >= numGlyphs

func parseTableGloc(data []byte, numGlyphs int) ([]uint32, uint16, error) {
	r := binaryreader.NewReader(data)
	if len(data) < 8 {
		return nil, 0, errors.New("invalid Gloc table (EOF)")
	}
	_, _ = r.Uint32()
	flags, _ := r.Uint16()
	numAttributes, _ := r.Uint16()
	isLong := flags&1 != 0

	// the number of locations may be greater than numGlyphs,
	// since there may be pseudo-glyphs
	// compute if from the end of the table:
	byteLength := len(data) - (int(numAttributes) * int(flags&2)) - 8
	numLocations := byteLength / 2
	if isLong {
		numLocations = byteLength / 4
	}

	if numLocations < numGlyphs+1 {
		return nil, 0, fmt.Errorf("invalid Gloc table: %d locations for %d glyphs ", numLocations, numGlyphs)
	}

	var locations []uint32
	if isLong {
		var err error
		if err != nil {
			return nil, 0, fmt.Errorf("invalid Gloc table: %s", err)
		}
		locations, err = r.Uint32s(numLocations)
	} else {
		tmp, err := r.Uint16s(numLocations)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid Gloc table: %s", err)
		}
		locations = make([]uint32, len(tmp))
		for i, o := range tmp {
			locations[i] = uint32(o)
		}
	}

	return locations, numAttributes, nil
}

// locations has length numGlyphs + 1
func parseTableGlat(data []byte, locations []uint32) (tableGlat, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid table Glat: (EOF)")
	}
	version := uint16(binary.BigEndian.Uint32(data) >> 16) // major
	out := make(tableGlat, len(locations)-1)
	var err error
	for i := range out {
		start, end := locations[i], locations[i+1]
		if start >= end {
			continue
		}
		if len(data) < int(end) {
			return nil, fmt.Errorf("invalid offset for table Glat: %d < %d", len(data), end)
		}
		glyphData := data[start:end]
		out[i], err = parseOneGlyphAttr(glyphData, version)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func parseOneGlyphAttr(data []byte, version uint16) (attributeSet, error) {
	if version >= 3 { // skip the octabox metrics
		if len(data) < 2 {
			return nil, errors.New("invalid Glat entry (EOF)")
		}
		bitmap := binary.BigEndian.Uint16(data)
		metricsLength := 6 + 8*bits.OnesCount16(bitmap)
		if len(data) < metricsLength {
			return nil, errors.New("invalid Glat entry (EOF)")
		}
		data = data[metricsLength:]
	}
	r := binaryreader.NewReader(data)
	var (
		out        attributeSet
		lastEndKey = -1
	)
	if version < 2 { // one byte
		for len(r.Data()) >= 2 {
			attNum, _ := r.Byte()
			num, _ := r.Byte()
			attributes, err := r.Int16s(int(num))
			if err != nil {
				return nil, fmt.Errorf("invalid Glat entry attributes: %s", err)
			}

			if int(attNum) < lastEndKey {
				return nil, fmt.Errorf("invalid Glat entry attribute key: %d", attNum)
			}
			lastEndKey = int(attNum) + len(attributes)

			out = append(out, attributSetEntry{
				firstKey:   uint16(attNum),
				attributes: attributes,
			})
		}
	} else { // same with two bytes header fields
		for len(r.Data()) >= 4 {
			attNum, _ := r.Uint16()
			num, _ := r.Uint16()
			attributes, err := r.Int16s(int(num))
			if err != nil {
				return nil, fmt.Errorf("invalid Glat entry attributes: %s", err)
			}

			if int(attNum) < lastEndKey {
				return nil, fmt.Errorf("invalid Glat entry attribute key: %d", attNum)
			}
			lastEndKey = int(attNum) + len(attributes)

			out = append(out, attributSetEntry{
				firstKey:   attNum,
				attributes: attributes,
			})
		}
	}

	return out, nil
}
