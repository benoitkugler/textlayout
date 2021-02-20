package truetype

import (
	"encoding/binary"
	"io"
)

// Tag represents an open-type name.
// These are technically uint32's, but are usually
// displayed in ASCII as they are all acronyms.
// See https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6.html#Overview
type Tag uint32

// MustNewTag gives you the Tag corresponding to the acronym.
// This function will panic if the string passed in is not 4 bytes long.
func MustNewTag(str string) Tag {
	bytes := []byte(str)

	if len(bytes) != 4 {
		panic("invalid tag: must be exactly 4 bytes")
	}

	return newTag(bytes)
}

func newTag(bytes []byte) Tag {
	return Tag(binary.BigEndian.Uint32(bytes))
}

// String returns the ASCII representation of the tag.
func (tag Tag) String() string {
	return string([]byte{
		byte(tag >> 24 & 0xFF),
		byte(tag >> 16 & 0xFF),
		byte(tag >> 8 & 0xFF),
		byte(tag & 0xFF),
	})
}

// parseUint16s interpret data as a (big endian) uint16 slice.
// It returns an error if data is not long enough for the given `length`.
func parseUint16s(data []byte, count int) ([]uint16, error) {
	if len(data) < 2*count {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]uint16, count)
	for i := range out {
		out[i] = binary.BigEndian.Uint16(data[2*i:])
	}
	return out, nil
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
