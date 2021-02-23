package truetype

import (
	"encoding/binary"
	"errors"
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
