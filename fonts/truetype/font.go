// Package truetype provides support for OpenType and TrueType font formats, used in PDF.
//
// It is largely influenced by github.com/ConradIrwin/font and golang.org/x/image/font/sfnt,
// and FreeType2.
package truetype

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	type1c "github.com/benoitkugler/textlayout/fonts/type1C"
)

var Loader fonts.FontLoader = loader{}

var _ fonts.Face = (*Font)(nil)

type loader struct{}

type fixed struct {
	Major int16
	Minor uint16
}

type longdatetime struct {
	SecondsSince1904 uint64
}

var (
	// errMissingHead is returned by ParseOTF when the font has no head section.
	errMissingHead = errors.New("missing head table in font")

	// errInvalidChecksum is returned by ParseOTF if the font's checksum is wrong
	errInvalidChecksum = errors.New("invalid checksum")

	// errUnsupportedFormat is returned from Parse if parsing failed
	errUnsupportedFormat = errors.New("unsupported font format")

	// errMissingTable is returned from *Table if the table does not exist in the font.
	errMissingTable = errors.New("missing table")

	errUnsupportedTableOffsetLength = errors.New("unsupported table offset or length")
	errInvalidDfont                 = errors.New("invalid dfont")
)

// Font represents a SFNT font, which is the underlying representation found
// in .otf and .ttf files.
// SFNT is a container format, which contains a number of tables identified by
// Tags. Depending on the type of glyphs embedded in the file which tables will
// exist. In particular, there's a big different between TrueType glyphs (usually .ttf)
// and CFF/PostScript Type 2 glyphs (usually .otf)
type Font struct {
	file   fonts.Resource        // source, needed to parse each table
	tables map[Tag]*tableSection // header only, contents is processed on demand

	// Optionnal, only present in variable fonts
	Fvar *TableFvar

	// Cmap is not empty after successful parsing
	Cmap TableCmap

	Names TableName

	Head TableHead

	// Type represents the kind of glyphs in this font.
	// It is one of TypeTrueType, TypeTrueTypeApple, TypePostScript1, TypeOpenType
	Type Tag

	// NumGlyphs exposes the number of glyph indexes present in the font.
	NumGlyphs uint16

	// True for fonts which include a 'hbed' table instead
	// of a 'head' table. Apple uses it as a flag that a font doesn't have
	// any glyph outlines but only embedded bitmaps
	isBinary bool
}

// tableSection represents a table within the font file.
type tableSection struct {
	offset  uint32 // Offset into the file this table starts.
	length  uint32 // Length of this table within the file.
	zLength uint32 // Uncompressed length of this table.
}

// loads the table corresponding to the 'head' tag.
// if a 'bhed' Apple table is present, it replaces the 'head' one
func (font *Font) loadHeadTable() error {
	s, hasbhed := font.tables[tagBhed]
	if !hasbhed {
		var hasHead bool
		s, hasHead = font.tables[tagHead]
		if !hasHead {
			return errors.New("missing required head (or bhed) table")
		}
	}
	font.isBinary = hasbhed

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return err
	}

	font.Head, err = parseTableHead(buf)
	return err
}

// loads the table corresponding to the 'name' tag.
// error only if the table is present and invalid
func (font *Font) tryAndLoadNameTable() error {
	s, found := font.tables[tagName]
	if !found {
		return nil
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return err
	}

	font.Names, err = parseTableName(buf)
	return err
}

// GlyfTable parse the 'glyf' table.
// Note that glyphs may be defined in various format (like CFF or bitmaps), and stored
// in other tables.
func (font *Font) GlyfTable() (TableGlyf, error) {
	buf, err := font.GetRawTable(tagLoca)
	if err != nil {
		return nil, err
	}

	loca, err := parseTableLoca(buf, int(font.NumGlyphs), font.Head.indexToLocFormat == 1)
	if err != nil {
		return nil, err
	}

	buf, err = font.GetRawTable(tagGlyf)
	if err != nil {
		return nil, err
	}

	return parseTableGlyf(buf, loca)
}

func (font *Font) cffTable() (*type1c.Font, error) {
	buf, err := font.GetRawTable(tagCFF)
	if err != nil {
		return nil, err
	}

	out, err := type1c.Parse(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	if N := out.NumGlyphs(); N != font.NumGlyphs {
		return nil, fmt.Errorf("invalid number of glyphs in CFF table (%d != %d)", N, font.NumGlyphs)
	}

	return out, nil
}

func (font *Font) sbixTable() (tableSbix, error) {
	buf, err := font.GetRawTable(tagSbix)
	if err != nil {
		return tableSbix{}, err
	}

	return parseTableSbix(buf, int(font.NumGlyphs))
}

// parse cblc and cbdt tables
func (font *Font) colorBitmapTable() (bitmapTable, error) {
	buf, err := font.GetRawTable(tagCBLC)
	if err != nil {
		return nil, err
	}

	rawImageData, err := font.GetRawTable(tagCBDT)
	if err != nil {
		return nil, err
	}

	return parseTableBitmap(buf, rawImageData)
}

// parse eblc and ebdt tables
func (font *Font) grayBitmapTable() (bitmapTable, error) {
	buf, err := font.GetRawTable(tagEBLC)
	if err != nil {
		return nil, err
	}

	rawImageData, err := font.GetRawTable(tagEBDT)
	if err != nil {
		return nil, err
	}

	return parseTableBitmap(buf, rawImageData)
}

func (font *Font) HheaTable() (*TableHVhea, error) {
	buf, err := font.GetRawTable(tagHhea)
	if err != nil {
		return nil, err
	}

	return parseTableHVhea(buf)
}

func (font *Font) VheaTable() (*TableHVhea, error) {
	buf, err := font.GetRawTable(tagVhea)
	if err != nil {
		return nil, err
	}

	return parseTableHVhea(buf)
}

func (font *Font) OS2Table() (*TableOS2, error) {
	buf, err := font.GetRawTable(tagOS2)
	if err != nil {
		return nil, err
	}

	return parseTableOS2(buf)
}

// GPOSTable returns the Glyph Positioning table identified with the 'GPOS' tag.
func (font *Font) GPOSTable() (TableGPOS, error) {
	buf, err := font.GetRawTable(TagGpos)
	if err != nil {
		return TableGPOS{}, err
	}

	return parseTableGPOS(buf)
}

// GSUBTable returns the Glyph Substitution table identified with the 'GSUB' tag.
func (font *Font) GSUBTable() (TableGSUB, error) {
	buf, err := font.GetRawTable(TagGsub)
	if err != nil {
		return TableGSUB{}, err
	}

	return parseTableGSUB(buf)
}

// GDEFTable returns the Glyph Definition table identified with the 'GDEF' tag.
func (font *Font) GDEFTable() (TableGDEF, error) {
	buf, err := font.GetRawTable(TagGdef)
	if err != nil {
		return TableGDEF{}, err
	}

	axisCount := 0
	if font.Fvar != nil {
		axisCount = len(font.Fvar.Axis)
	}
	return parseTableGdef(buf, axisCount)
}

func (font *Font) loadCmapTable() error {
	s, found := font.tables[tagCmap]
	if !found {
		return errors.New("missing required 'cmap' table")
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return fmt.Errorf("invalid required cmap table: %s", err)
	}

	font.Cmap, err = parseTableCmap(buf)
	return err
}

// PostTable returns the Post table names
func (font *Font) PostTable() (PostTable, error) {
	buf, err := font.GetRawTable(tagPost)
	if err != nil {
		return PostTable{}, err
	}

	return parseTablePost(buf, font.NumGlyphs)
}

// loadNumGlyphs parses the 'maxp' table to find the number of glyphs in the font.
func (font *Font) loadNumGlyphs() error {
	buf, err := font.GetRawTable(tagMaxp)
	if err != nil {
		return err
	}

	font.NumGlyphs, err = parseTableMaxp(buf)
	return err
}

// HtmxTable returns the glyphs horizontal metrics (array of size numGlyphs),
// expressed in fonts units.
func (font *Font) HtmxTable() (TableHVmtx, error) {
	hhea, err := font.HheaTable()
	if err != nil {
		return nil, err
	}

	buf, err := font.GetRawTable(tagHmtx)
	if err != nil {
		return nil, err
	}

	return parseHVmtxTable(buf, uint16(hhea.numOfLongMetrics), font.NumGlyphs)
}

// VtmxTable returns the glyphs vertical metrics (array of size numGlyphs),
// expressed in fonts units.
func (font *Font) VtmxTable() (TableHVmtx, error) {
	vhea, err := font.VheaTable()
	if err != nil {
		return nil, err
	}

	buf, err := font.GetRawTable(tagVmtx)
	if err != nil {
		return nil, err
	}

	return parseHVmtxTable(buf, uint16(vhea.numOfLongMetrics), font.NumGlyphs)
}

// LayoutTables exposes advanced layout tables.
// All the fields are optionnals.
type LayoutTables struct {
	GDEF TableGDEF // An absent table has a nil Class
	Trak TableTrak
	Ankr TableAnkr
	Feat TableFeat
	Morx TableMorx
	Kern TableKernx
	Kerx TableKernx
	GSUB TableGSUB // An absent table has a nil slice of lookups
	GPOS TableGPOS // An absent table has a nil slice of lookups
}

// LayoutTables try and parse all the advanced layout tables.
// When parsing yields an error, it is ignored and `nil` is returned.
// See the individual methods for more control over error handling.
func (font *Font) LayoutTables() LayoutTables {
	var out LayoutTables
	if tb, err := font.GDEFTable(); err == nil {
		out.GDEF = tb
	}
	if tb, err := font.GSUBTable(); err == nil {
		out.GSUB = tb
	}
	if tb, err := font.GPOSTable(); err == nil {
		out.GPOS = tb
	}

	if tb, err := font.MorxTable(); err == nil {
		out.Morx = tb
	}
	if tb, err := font.KernTable(); err == nil {
		out.Kern = tb
	}
	if tb, err := font.KerxTable(); err == nil {
		out.Kerx = tb
	}
	if tb, err := font.AnkrTable(); err == nil {
		out.Ankr = tb
	}
	if tb, err := font.TrakTable(); err == nil {
		out.Trak = tb
	}
	if tb, err := font.FeatTable(); err == nil {
		out.Feat = tb
	}

	return out
}

func (font *Font) KernTable() (TableKernx, error) {
	buf, err := font.GetRawTable(tagKern)
	if err != nil {
		return nil, err
	}

	return parseKernTable(buf, int(font.NumGlyphs))
}

// MorxTable parse the AAT 'morx' table.
func (font *Font) MorxTable() (TableMorx, error) {
	buf, err := font.GetRawTable(tagMorx)
	if err != nil {
		return nil, err
	}

	return parseTableMorx(buf, int(font.NumGlyphs))
}

// KerxTable parse the AAT 'kerx' table.
func (font *Font) KerxTable() (TableKernx, error) {
	buf, err := font.GetRawTable(tagKerx)
	if err != nil {
		return nil, err
	}

	return parseTableKerx(buf, int(font.NumGlyphs))
}

// AnkrTable parse the AAT 'ankr' table.
func (font *Font) AnkrTable() (TableAnkr, error) {
	buf, err := font.GetRawTable(tagAnkr)
	if err != nil {
		return TableAnkr{}, err
	}

	return parseTableAnkr(buf, int(font.NumGlyphs))
}

// TrakTable parse the AAT 'trak' table.
func (font *Font) TrakTable() (TableTrak, error) {
	buf, err := font.GetRawTable(tagTrak)
	if err != nil {
		return TableTrak{}, err
	}

	return parseTrakTable(buf)
}

// FeatTable parse the AAT 'feat' table.
func (font *Font) FeatTable() (TableFeat, error) {
	buf, err := font.GetRawTable(tagFeat)
	if err != nil {
		return nil, err
	}

	return parseTableFeat(buf)
}

// error only if the table is present and invalid
func (font *Font) tryAndLoadFvarTable() error {
	s, found := font.tables[tagFvar]
	if !found {
		return nil
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return err
	}

	font.Fvar, err = parseTableFvar(buf, font.Names)
	return err
}

func (font *Font) avarTable() (tableAvar, error) {
	buf, err := font.GetRawTable(tagAvar)
	if err != nil {
		return nil, err
	}

	return parseTableAvar(buf, len(font.Fvar.Axis))
}

func (font *Font) gvarTable(glyphs TableGlyf) (tableGvar, error) {
	buf, err := font.GetRawTable(tagGvar)
	if err != nil {
		return tableGvar{}, err
	}

	return parseTableGvar(buf, len(font.Fvar.Axis), glyphs)
}

func (font *Font) hvarTable() (tableHVvar, error) {
	buf, err := font.GetRawTable(tagHvar)
	if err != nil {
		return tableHVvar{}, err
	}

	return parseTableHVvar(buf, len(font.Fvar.Axis))
}

func (font *Font) vvarTable() (tableHVvar, error) {
	buf, err := font.GetRawTable(tagVvar)
	if err != nil {
		return tableHVvar{}, err
	}

	return parseTableHVvar(buf, len(font.Fvar.Axis))
}

func (font *Font) mvarTable() (TableMvar, error) {
	buf, err := font.GetRawTable(tagMvar)
	if err != nil {
		return TableMvar{}, err
	}

	return parseTableMvar(buf, len(font.Fvar.Axis))
}

func (font *Font) vorgTable() (tableVorg, error) {
	buf, err := font.GetRawTable(tagVorg)
	if err != nil {
		return tableVorg{}, err
	}

	return parseTableVorg(buf)
}

// Parse parses an OpenType or TrueType file and returns a Font.
// It only loads the minimal required tables: 'head', 'maxp', 'name' and 'cmap' tables.
// It also look for an 'fvar' table and parses it if found.
// The underlying file is still needed to parse the remaining tables, and must not be closed.
// See Loader for support for collections.
func Parse(file fonts.Resource) (*Font, error) {
	return parseOneFont(file, 0, false)
}

// Load implements fonts.FontLoader. For collection font files (.ttc, .otc),
// multiple fonts may be returned.
func (loader) Load(file fonts.Resource) (fonts.Fonts, error) {
	_, err := file.Seek(0, io.SeekStart) // file might have been used before
	if err != nil {
		return nil, err
	}

	var bytes [4]byte
	_, err = file.Read(bytes[:])
	if err != nil {
		return nil, err
	}
	magic := newTag(bytes[:])

	file.Seek(0, io.SeekStart)

	var (
		f              *Font
		offsets        []uint32
		relativeOffset bool
	)
	switch magic {
	case SignatureWOFF, TypeTrueType, TypeOpenType, TypePostScript1, TypeAppleTrueType:
		f, err = parseOneFont(file, 0, false)
	case ttcTag:
		offsets, err = parseTTCHeader(file)
	case dfontResourceDataOffset:
		offsets, err = parseDfont(file)
		relativeOffset = true
	default:
		return nil, fmt.Errorf("unsupported font format %v", bytes)
	}
	if err != nil {
		return nil, err
	}

	// only one font
	if f != nil {
		return fonts.Fonts{f}, nil
	}

	// collection
	out := make(fonts.Fonts, len(offsets))
	for i, o := range offsets {
		out[i], err = parseOneFont(file, o, relativeOffset)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// load 'maxp' as well
func parseOneFont(file fonts.Resource, offset uint32, relativeOffset bool) (f *Font, err error) {
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("invalid offset: %s", err)
	}

	var bytes [4]byte
	_, err = file.Read(bytes[:])
	if err != nil {
		return nil, err
	}
	magic := newTag(bytes[:])

	switch magic {
	case SignatureWOFF:
		f, err = parseWOFF(file, offset, relativeOffset)
	case TypeTrueType, TypeOpenType, TypePostScript1, TypeAppleTrueType:
		f, err = parseOTF(file, offset, relativeOffset)
	default:
		// no more collections allowed here
		return nil, errUnsupportedFormat
	}

	if err != nil {
		return nil, err
	}

	err = f.loadNumGlyphs()
	if err != nil {
		return nil, err
	}
	err = f.loadCmapTable()
	if err != nil {
		return nil, err
	}
	err = f.loadHeadTable()
	if err != nil {
		return nil, err
	}
	err = f.tryAndLoadNameTable()
	if err != nil {
		return nil, err
	}
	err = f.tryAndLoadFvarTable()
	return f, err
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

		buf = make([]byte, s.zLength)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
	} else {
		buf = make([]byte, s.length)
		if _, err := font.file.ReadAt(buf, int64(s.offset)); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

// HasTable returns `true` is the font has the given table.
func (font *Font) HasTable(tag Tag) bool {
	_, has := font.tables[tag]
	return has
}

// GetRawTable returns the binary content of the given table,
// or an error if not found.
// Note that many tables are already interpreted by this package,
// see the various XXXTable().
func (font *Font) GetRawTable(tag Tag) ([]byte, error) {
	s, found := font.tables[tag]
	if !found {
		return nil, errMissingTable
	}

	return font.findTableBuffer(s)
}

func (font *Font) PostscriptInfo() (fonts.PSInfo, bool) {
	return fonts.PSInfo{}, false
}

// PoscriptName returns the optional PoscriptName of the font
func (font *Font) PoscriptName() string {
	// adapted from freetype

	// TODO: support multiple masters

	// scan the name table to see whether we have a Postscript name here,
	// either in Macintosh or Windows platform encodings
	windows, mac := font.Names.getEntry(NamePostscript)

	// prefer Windows entries over Apple
	if windows != nil {
		return windows.String()
	}
	if mac != nil {
		return mac.String()
	}
	return ""
}

type fontSummary struct {
	head            *TableHead
	os2             *TableOS2
	names           TableName
	hasOutline      bool
	hasBitmap       bool
	hasColor        bool
	hasVerticalInfo bool
}

// loadSummary loads various tables to compute meta data about the font
func (font *Font) loadSummary() (fontSummary, error) {
	// adapted from freetype

	var out fontSummary
	out.names = font.Names
	if font.HasTable(tagCBLC) || font.HasTable(tagSbix) || font.HasTable(tagCOLR) {
		out.hasColor = true
	}
	out.head = &font.Head

	// do we have outlines in there ?
	out.hasOutline = font.HasTable(tagGlyf) || font.HasTable(tagCFF) || font.HasTable(tagCFF2)

	isAppleSbix := font.HasTable(tagSbix)

	// Apple 'sbix' color bitmaps are rendered scaled and then the 'glyf'
	// outline rendered on top.  We don't support that yet, so just ignore
	// the 'glyf' outline and advertise it as a bitmap-only font.
	if isAppleSbix {
		out.hasOutline = false
	}

	isAppleSbit := font.isBinary

	hasCblc := font.HasTable(tagCBLC)
	hasCbdt := font.HasTable(tagCBDT)

	// Ignore outlines for CBLC/CBDT fonts.
	if hasCblc || hasCbdt {
		out.hasOutline = false
	}

	out.hasBitmap = hasCblc && hasCbdt || font.HasTable(tagEBDT) && font.HasTable(tagEBLC) || isAppleSbit || isAppleSbix

	// OpenType 1.8.2 introduced limits to this value;
	// however, they make sense for older SFNT fonts also
	if out.head.UnitsPerEm < 16 || out.head.UnitsPerEm > 16384 {
		return out, fmt.Errorf("invalid UnitsPerEm value %d", out.head.UnitsPerEm)
	}

	// do not load the metrics headers and tables if this is an Apple
	// sbit font file
	if isAppleSbit {
		return out, nil
	}

	// load the `hhea' and `hmtx' tables
	_, err := font.HheaTable()
	if err == nil {
		_, err = font.HtmxTable()
		if err != nil {
			return out, err
		}
	} else {
		// No `hhea' table necessary for SFNT Mac fonts.
		if font.Type == TypeAppleTrueType {
			out.hasOutline = false
		} else {
			return out, errors.New("horizontal header is missing")
		}
	}

	// try to load the `vhea' and `vmtx' tables
	if font.HasTable(tagVhea) {
		_, err = font.VheaTable()
		if err != nil {
			return out, err
		}
		_, err := font.VtmxTable()
		out.hasVerticalInfo = err == nil
	}

	out.os2, _ = font.OS2Table() // we treat the table as missing if there are any errors
	return out, nil
}

func (font *Font) LoadSummary() (fonts.FontSummary, error) {
	sm, err := font.loadSummary()
	if err != nil {
		return fonts.FontSummary{}, err
	}
	isItalic, isBold, familyName, styleName := sm.getStyle()
	return fonts.FontSummary{
		IsItalic: isItalic,
		IsBold:   isBold,
		Familly:  familyName,
		Style:    styleName,
		// a font with no bitmaps and no outlines is scalable;
		// it has only empty glyphs then
		HasScalableGlyphs: !sm.hasBitmap,
		HasBitmapGlyphs:   sm.hasBitmap,
		HasColorGlyphs:    sm.hasColor,
	}, nil
}

// getStyle sum up the style of the font
func (summary fontSummary) getStyle() (isItalic, isBold bool, familyName, styleName string) {
	// Bit 8 of the `fsSelection' field in the `OS/2' table denotes
	// a WWS-only font face.  `WWS' stands for `weight', width', and
	// `slope', a term used by Microsoft's Windows Presentation
	// Foundation (WPF).  This flag has been introduced in version
	// 1.5 of the OpenType specification (May 2008).

	if summary.os2 != nil && summary.os2.FsSelection&256 != 0 {
		familyName = summary.names.getName(NamePreferredFamily)
		if familyName == "" {
			familyName = summary.names.getName(NameFontFamily)
		}

		styleName = summary.names.getName(NamePreferredSubfamily)
		if styleName == "" {
			styleName = summary.names.getName(NameFontSubfamily)
		}
	} else {
		familyName = summary.names.getName(NameWWSFamily)
		if familyName == "" {
			familyName = summary.names.getName(NamePreferredFamily)
		}
		if familyName == "" {
			familyName = summary.names.getName(NameFontFamily)
		}

		styleName = summary.names.getName(NameWWSSubfamily)
		if styleName == "" {
			styleName = summary.names.getName(NamePreferredSubfamily)
		}
		if styleName == "" {
			styleName = summary.names.getName(NameFontSubfamily)
		}
	}

	styleName = strings.TrimSpace(styleName)
	if styleName == "" { // assume `Regular' style because we don't know better
		styleName = "Regular"
	}

	// Compute style flags.
	if summary.hasOutline && summary.os2 != nil {
		// We have an OS/2 table; use the `fsSelection' field.  Bit 9
		// indicates an oblique font face.  This flag has been
		// introduced in version 1.5 of the OpenType specification.
		isItalic = summary.os2.FsSelection&(1<<9) != 0 || summary.os2.FsSelection&1 != 0
		isBold = summary.os2.FsSelection&(1<<5) != 0
	} else {
		// this is an old Mac font, use the header field
		isBold = summary.head.MacStyle&1 != 0
		isItalic = summary.head.MacStyle&2 != 0
	}

	return
}
