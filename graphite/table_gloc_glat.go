package graphite

import (
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts/binaryreader"
)

func parseTableGloc(data []byte, numGlyphs int) ([]uint32, uint16, error) {
	r := binaryreader.NewReader(data)
	if len(data) < 8 {
		return nil, 0, errors.New("invalid Gloc table (EOF)")
	}
	_, _ = r.Uint32()
	flags, _ := r.Uint16()
	numAttributes, _ := r.Uint16()
	isLong := flags&1 != 0
	var locations []uint32
	if isLong {
		var err error
		if err != nil {
			return nil, 0, fmt.Errorf("invalid Gloc table: %s", err)
		}
		locations, err = r.Uint32s(numGlyphs + 1)
	} else {
		tmp, err := r.Uint16s(numGlyphs + 1)
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

func parseTableGlat(data []byte, locations []uint32) (TableFeat, error) {
	r := binaryreader.NewReader(data)
	_, err := r.Uint32()
	if err != nil {
		return nil, fmt.Errorf("invalid table Glat: %s", err)
	}
}
