package truetype

import (
	"compress/zlib"
	"encoding/binary"
	"io"
)

var (
	// tagHead represents the 'head' table, which contains the font header
	tagHead = mustNamedTag("head")
	// tagMaxp represents the 'maxp' table, which contains the maximum profile
	tagMaxp = mustNamedTag("maxp")
	// tagHmtx represents the 'hmtx' table, which contains the horizontal metrics
	tagHmtx = mustNamedTag("hmtx")
	// tagHhea represents the 'hhea' table, which contains the horizonal header
	tagHhea = mustNamedTag("hhea")
	// tagOS2 represents the 'OS/2' table, which contains windows-specific metadata
	tagOS2 = mustNamedTag("OS/2")
	// tagName represents the 'name' table, which contains font name information
	tagName = mustNamedTag("name")
	// tagGpos represents the 'GPOS' table, which contains Glyph Positioning features
	tagGpos = mustNamedTag("GPOS")
	// tagGsub represents the 'GSUB' table, which contains Glyph Substitution features
	tagGsub = mustNamedTag("GSUB")

	tagCmap = mustNamedTag("cmap")
	tagKern = mustNamedTag("kern")
	tagPost = mustNamedTag("post")
	TagSilf = mustNamedTag("Silf")
	TagPrep = mustNamedTag("prep")
	tagGlyf = mustNamedTag("glyf")
	tagCFF  = mustNamedTag("CFF ")
	tagCFF2 = mustNamedTag("CFF2")
	tagSbix = mustNamedTag("sbix")
	tagBhed = mustNamedTag("bhed")
	tagCBLC = mustNamedTag("CBLC")
	tagCBDT = mustNamedTag("CBDT")

	// TypeTrueType is the first four bytes of an OpenType file containing a TrueType font
	TypeTrueType = TableTag(0x00010000)
	// TypeAppleTrueType is the first four bytes of an OpenType file containing a TrueType font
	// (specifically one designed for Apple products, it's recommended to use TypeTrueType instead)
	TypeAppleTrueType = mustNamedTag("true")
	// TypePostScript1 is the first four bytes of an OpenType file containing a PostScript Type 1 font
	TypePostScript1 = mustNamedTag("typ1")
	// TypeOpenType is the first four bytes of an OpenType file containing a PostScript Type 2 font
	// as specified by OpenType
	TypeOpenType = mustNamedTag("OTTO")
)

// TableTag represents an open-type table name.
// These are technically uint32's, but are usually
// displayed in ASCII as they are all acronyms.
// see https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6.html#Overview
type TableTag uint32

// mustNamedTag gives you the Tag corresponding to the acronym.
// This function will panic if the string passed in is not 4 bytes long.
func mustNamedTag(str string) TableTag {
	bytes := []byte(str)

	if len(bytes) != 4 {
		panic("invalid tag: must be exactly 4 bytes")
	}

	return TableTag(uint32(bytes[0])<<24 |
		uint32(bytes[1])<<16 |
		uint32(bytes[2])<<8 |
		uint32(bytes[3]))

}

func newTag(bytes []byte) TableTag {
	return TableTag(binary.BigEndian.Uint32(bytes))
}

func readTag(r io.Reader) (TableTag, error) {
	bytes := make([]byte, 4)
	_, err := io.ReadFull(r, bytes)
	return newTag(bytes), err
}

// String returns the ASCII representation of the tag.
func (tag TableTag) String() string {
	return string([]byte{
		byte(tag >> 24 & 0xFF),
		byte(tag >> 16 & 0xFF),
		byte(tag >> 8 & 0xFF),
		byte(tag & 0xFF),
	})
}

func (font *Font) findTableBuffer(s *tableSection) ([]byte, error) {
	var buf []byte

	if s.length != 0 && s.length < s.zLength {
		zbuf := io.NewSectionReader(font.file, int64(s.offset), int64(s.length))
		r, err := zlib.NewReader(zbuf)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		buf = make([]byte, s.zLength, s.zLength)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
	} else {
		buf = make([]byte, s.length, s.length)
		if _, err := font.file.ReadAt(buf, int64(s.offset)); err != nil {
			return nil, err
		}
	}
	return buf, nil
}
