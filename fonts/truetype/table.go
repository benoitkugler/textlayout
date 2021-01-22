package truetype

import (
	"compress/zlib"
	"encoding/binary"
	"io"
)

var (
	// tagHead represents the 'head' table, which contains the font header
	tagHead = MustNewTag("head")
	// tagMaxp represents the 'maxp' table, which contains the maximum profile
	tagMaxp = MustNewTag("maxp")
	// tagHmtx represents the 'hmtx' table, which contains the horizontal metrics
	tagHmtx = MustNewTag("hmtx")
	// tagHhea represents the 'hhea' table, which contains the horizonal header
	tagHhea = MustNewTag("hhea")
	// tagOS2 represents the 'OS/2' table, which contains windows-specific metadata
	tagOS2 = MustNewTag("OS/2")
	// tagName represents the 'name' table, which contains font name information
	tagName = MustNewTag("name")
	// tagGpos represents the 'GPOS' table, which contains Glyph Positioning features
	tagGpos = MustNewTag("GPOS")
	// tagGsub represents the 'GSUB' table, which contains Glyph Substitution features
	tagGsub = MustNewTag("GSUB")

	tagCmap = MustNewTag("cmap")
	tagKern = MustNewTag("kern")
	tagPost = MustNewTag("post")
	TagSilf = MustNewTag("Silf")
	TagPrep = MustNewTag("prep")
	tagGlyf = MustNewTag("glyf")
	tagCFF  = MustNewTag("CFF ")
	tagCFF2 = MustNewTag("CFF2")
	tagSbix = MustNewTag("sbix")
	tagBhed = MustNewTag("bhed")
	tagCBLC = MustNewTag("CBLC")
	tagCBDT = MustNewTag("CBDT")
	tagEBLC = MustNewTag("EBLC")
	tagBloc = MustNewTag("bloc")
	tagCOLR = MustNewTag("COLR")
	tagFvar = MustNewTag("fvar")
	tagAvar = MustNewTag("avar")

	// TypeTrueType is the first four bytes of an OpenType file containing a TrueType font
	TypeTrueType = Tag(0x00010000)
	// TypeAppleTrueType is the first four bytes of an OpenType file containing a TrueType font
	// (specifically one designed for Apple products, it's recommended to use TypeTrueType instead)
	TypeAppleTrueType = MustNewTag("true")
	// TypePostScript1 is the first four bytes of an OpenType file containing a PostScript Type 1 font
	TypePostScript1 = MustNewTag("typ1")
	// TypeOpenType is the first four bytes of an OpenType file containing a PostScript Type 2 font
	// as specified by OpenType
	TypeOpenType = MustNewTag("OTTO")

	// SignatureWOFF is the magic number at the start of a WOFF file.
	SignatureWOFF = MustNewTag("wOFF")

	ttcTag = MustNewTag("ttcf")

	// // SignatureWOFF2 is the magic number at the start of a WOFF2 file.
	// SignatureWOFF2 = MustNewTag("wOF2")
)

// dfontResourceDataOffset is the assumed value of a dfont file's resource data
// offset.
//
// https://github.com/kreativekorp/ksfl/wiki/Macintosh-Resource-File-Format
// says that "A Mac OS resource file... [starts with an] offset from start of
// file to start of resource data section... [usually] 0x0100". In theory,
// 0x00000100 isn't always a magic number for identifying dfont files. In
// practice, it seems to work.
const dfontResourceDataOffset = 0x00000100

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
