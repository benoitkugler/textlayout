package graphite

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/pierrec/lz4"
)

const uncompressedLimit = 10000000

// Some fonts contains tables compressed with the lz4 format
// this function expect the table to begin with a version (uint32)
// and check if it then has a valid encoding scheme
func decompressTable(data []byte) ([]byte, uint16, error) {
	if len(data) < 4 {
		return nil, 0, errors.New("invalid table (EOF)")
	}
	version := uint16(binary.BigEndian.Uint32(data) >> 16) // major

	if version >= 3 && len(data) >= 8 {
		compression := binary.BigEndian.Uint32(data[4:8])
		scheme := compression >> 27
		switch scheme {
		case 0: // no compression, just go on

		case 1: // lz4 compression
			size := compression & 0x07ffffff
			if size > uncompressedLimit {
				return nil, 0, fmt.Errorf("unsupported size for uncompressed Glat table: %d", size)
			}
			uncompressed := make([]byte, size)
			_, err := lz4.UncompressBlock(data[8:], uncompressed)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid lz4 compressed table: %s", err)
			}
			return uncompressed, version, nil
		default:
			return nil, 0, fmt.Errorf("unsupported compression scheme: %d", scheme)
		}
	}
	return data, version, nil
}
