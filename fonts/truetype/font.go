// Package truetype provides support for OpenType and TrueType font formats, used in PDF.
//
// It is largely influenced by github.com/ConradIrwin/font and golang.org/x/image/font/sfnt,
// and FreeType2.
package truetype

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
)

var Loader fonts.FontLoader = loader{}

var _ fonts.Font = (*Font)(nil)

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
	file   fonts.Ressource       // source, needed to parse each table
	tables map[Tag]*tableSection // header only, contents is processed on demand

	// Optionnal, only present in variable fonts
	Fvar *TableFvar

	// Cmap is not empty after successful parsing
	Cmap TableCmap

	Name TableName

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
func (font *Font) loadNameTable() error {
	s, found := font.tables[tagName]
	if !found {
		return errors.New("missing required 'name' table")
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return err
	}

	font.Name, err = parseTableName(buf)
	return err
}

func (font *Font) glyfTable() (TableGlyf, error) {
	s, found := font.tables[tagLoca]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}

	loca, err := parseTableLoca(buf, int(font.NumGlyphs), font.Head.IndexToLocFormat == 1)
	if err != nil {
		return nil, err
	}

	s, found = font.tables[tagGlyf]
	if !found {
		return nil, errMissingTable
	}

	buf, err = font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}

	return parseTableGlyf(buf, loca)
}

func (font *Font) HheaTable() (*TableHVhea, error) {
	s, found := font.tables[tagHhea]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}

	return parseTableHVhea(buf)
}

func (font *Font) VheaTable() (*TableHVhea, error) {
	s, found := font.tables[tagVhea]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}

	return parseTableHVhea(buf)
}

func (font *Font) OS2Table() (*TableOS2, error) {
	s, found := font.tables[tagOS2]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}

	return parseTableOS2(buf)
}

// GPOSTable returns the Glyph Positioning table identified with the 'GPOS' tag.
func (font *Font) GPOSTable() (TableGPOS, error) {
	s, found := font.tables[TagGpos]
	if !found {
		return TableGPOS{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return TableGPOS{}, err
	}

	return parseTableGPOS(buf)
}

// GSUBTable returns the Glyph Substitution table identified with the 'GSUB' tag.
func (font *Font) GSUBTable() (TableGSUB, error) {
	s, found := font.tables[TagGsub]
	if !found {
		return TableGSUB{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return TableGSUB{}, err
	}

	return parseTableGSUB(buf)
}

// GDEFTable returns the Glyph Definition table identified with the 'GDEF' tag.
func (font *Font) GDEFTable() (TableGDEF, error) {
	s, found := font.tables[TagGdef]
	if !found {
		return TableGDEF{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
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
	s, found := font.tables[tagPost]
	if !found {
		return PostTable{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return PostTable{}, err
	}

	return parseTablePost(buf, font.NumGlyphs)
}

// loadNumGlyphs parses the 'maxp' table to find the number of glyphs in the font.
func (font *Font) loadNumGlyphs() error {
	maxpSection, found := font.tables[tagMaxp]
	if !found {
		return errMissingTable
	}

	buf, err := font.findTableBuffer(maxpSection)
	if err != nil {
		return err
	}

	font.NumGlyphs, err = parseMaxpTable(buf)
	return err
}

// HtmxTable returns the glyphs widths (array of size numGlyphs),
// expressed in fonts units.
func (font *Font) HtmxTable() (tableHVmtx, error) {
	hhea, err := font.HheaTable()
	if err != nil {
		return nil, err
	}

	htmxSection, found := font.tables[tagHmtx]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(htmxSection)
	if err != nil {
		return nil, err
	}

	return parseHVmtxTable(buf, uint16(hhea.numOfLongMetrics), font.NumGlyphs)
}

// LayoutTables exposes advanced layout tables.
// All the fields are optionnals.
type LayoutTables struct {
	GDEF TableGDEF // An empty table has a nil Class
	Trak TableTrak
	Ankr TableAnkr
	Feat TableFeat
	Morx TableMorx
	Kern TableKernx
	Kerx TableKernx
	GSUB TableGSUB
	GPOS TableGPOS
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
	section, found := font.tables[tagKern]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(section)
	if err != nil {
		return nil, err
	}

	return parseKernTable(buf, int(font.NumGlyphs))
}

// MorxTable parse the AAT 'morx' table.
func (font *Font) MorxTable() (TableMorx, error) {
	s, found := font.tables[tagMorx]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}

	return parseTableMorx(buf, int(font.NumGlyphs))
}

// KerxTable parse the AAT 'kerx' table.
func (font *Font) KerxTable() (TableKernx, error) {
	s, found := font.tables[tagKerx]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}

	return parseTableKerx(buf, int(font.NumGlyphs))
}

// AnkrTable parse the AAT 'ankr' table.
func (font *Font) AnkrTable() (TableAnkr, error) {
	s, found := font.tables[tagAnkr]
	if !found {
		return TableAnkr{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return TableAnkr{}, err
	}

	return parseTableAnkr(buf, int(font.NumGlyphs))
}

// TrakTable parse the AAT 'trak' table.
func (font *Font) TrakTable() (TableTrak, error) {
	section, found := font.tables[tagTrak]
	if !found {
		return TableTrak{}, errMissingTable
	}

	buf, err := font.findTableBuffer(section)
	if err != nil {
		return TableTrak{}, err
	}

	return parseTrakTable(buf)
}

// FeatTable parse the AAT 'feat' table.
func (font *Font) FeatTable() (TableFeat, error) {
	section, found := font.tables[tagFeat]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(section)
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

	font.Fvar, err = parseTableFvar(buf, font.Name)
	return err
}

func (font *Font) avarTable() (tableAvar, error) {
	s, found := font.tables[tagAvar]
	if !found {
		return nil, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return nil, err
	}
	return parseTableAvar(buf, len(font.Fvar.Axis))
}

func (font *Font) gvarTable(glyphs TableGlyf) (tableGvar, error) {
	s, found := font.tables[tagGvar]
	if !found {
		return tableGvar{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return tableGvar{}, err
	}
	return parseTableGvar(buf, len(font.Fvar.Axis), glyphs)
}

func (font *Font) hvarTable() (tableHVvar, error) {
	s, found := font.tables[tagHvar]
	if !found {
		return tableHVvar{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return tableHVvar{}, err
	}
	return parseTableHVvar(buf, len(font.Fvar.Axis))
}

func (font *Font) vvarTable() (tableHVvar, error) {
	s, found := font.tables[tagVvar]
	if !found {
		return tableHVvar{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return tableHVvar{}, err
	}
	return parseTableHVvar(buf, len(font.Fvar.Axis))
}

func (font *Font) mvarTable() (TableMvar, error) {
	s, found := font.tables[tagMvar]
	if !found {
		return TableMvar{}, errMissingTable
	}

	buf, err := font.findTableBuffer(s)
	if err != nil {
		return TableMvar{}, err
	}
	return parseTableMvar(buf, len(font.Fvar.Axis))
}

// Parse parses an OpenType or TrueType file and returns a Font.
// It only loads the minimal required tables: 'head', 'maxp', 'name' and 'cmap' tables.
// It also look for an 'fvar' table and parses it if found.
// The underlying file is still needed to parse the remaining tables, and must not be closed.
// See Loader for support for collections.
func Parse(file fonts.Ressource) (*Font, error) {
	return parseOneFont(file, 0, false)
}

// Load implements fonts.FontLoader. For collection font files (.ttc, .otc),
// multiple fonts may be returned.
func (loader) Load(file fonts.Ressource) (fonts.Fonts, error) {
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
	case SignatureWOFF:
		f, err = parseWOFF(file, 0, false)
	case TypeTrueType, TypeOpenType, TypePostScript1, TypeAppleTrueType:
		f, err = parseOTF(file, 0, false)
	case ttcTag:
		offsets, err = parseTTCHeader(file)
	case dfontResourceDataOffset:
		offsets, err = parseDfont(file)
		relativeOffset = true
	default:
		return nil, errUnsupportedFormat
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
func parseOneFont(file fonts.Ressource, offset uint32, relativeOffset bool) (f *Font, err error) {
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
	err = f.loadNameTable()
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

// HasTable returns `true` is the font has the given table.
func (f *Font) HasTable(tag Tag) bool {
	_, has := f.tables[tag]
	return has
}

func (f *Font) PostscriptInfo() (fonts.PSInfo, bool) {
	return fonts.PSInfo{}, false
}

// PoscriptName returns the optional PoscriptName of the font
func (f *Font) PoscriptName() string {
	// adapted from freetype

	// TODO: support multiple masters

	// scan the name table to see whether we have a Postscript name here,
	// either in Macintosh or Windows platform encodings
	windows, mac := f.Name.getEntry(NamePostscript)

	// prefer Windows entries over Apple
	if windows != nil {
		return windows.String()
	}
	if mac != nil {
		return mac.String()
	}
	return ""
}

// TODO: polish and cache on the font
type fontDetails struct {
	head       *TableHead
	os2        *TableOS2
	hasOutline bool
	hasColor   bool
}

// load various tables to compute meta data
func (f *Font) analyze() (fontDetails, error) {
	var out fontDetails
	if f.HasTable(tagCBLC) || f.HasTable(tagSbix) || f.HasTable(tagCOLR) {
		out.hasColor = true
	}
	out.head = &f.Head

	// do we have outlines in there ?
	out.hasOutline = f.HasTable(tagGlyf) || f.HasTable(tagCFF) || f.HasTable(tagCFF2)

	isAppleSbix := f.HasTable(tagSbix)

	// Apple 'sbix' color bitmaps are rendered scaled and then the 'glyf'
	// outline rendered on top.  We don't support that yet, so just ignore
	// the 'glyf' outline and advertise it as a bitmap-only font.
	if isAppleSbix {
		out.hasOutline = false
	}

	isAppleSbit := f.isBinary

	hasCblc := f.HasTable(tagCBLC)
	hasCbdt := f.HasTable(tagCBDT)

	// Ignore outlines for CBLC/CBDT fonts.
	if hasCblc || hasCbdt {
		out.hasOutline = false
	}

	// OpenType 1.8.2 introduced limits to this value;
	// however, they make sense for older SFNT fonts also
	if out.head.UnitsPerEm < 16 || out.head.UnitsPerEm > 16384 {
		return out, fmt.Errorf("invalid UnitsPerEm value %d", out.head.UnitsPerEm)
	}

	// TODO: check if this is needed
	// /* the following tables are often not present in embedded TrueType */
	// /* fonts within PDF documents, so don't check for them.            */
	// LOAD_(maxp)
	// LOAD_(cmap)

	// /* the following tables are optional in PCL fonts -- */
	// /* don't check for errors                            */
	// LOAD_(name)
	// LOAD_(post)

	// do not load the metrics headers and tables if this is an Apple
	// sbit font file
	if isAppleSbit {
		return out, nil
	}

	// load the `hhea' and `hmtx' tables
	_, err := f.HheaTable()
	if err == nil {
		_, err = f.HtmxTable()
		if err != nil {
			return out, err
		}
	} else {
		// No `hhea' table necessary for SFNT Mac fonts.
		if f.Type == TypeAppleTrueType {
			out.hasOutline = false
		} else {
			return out, errors.New("horizontal header is missing")
		}
	}

	// TODO:
	// try to load the `vhea' and `vmtx' tables
	// LOADM_(hhea, 1)
	// if !error {
	// 	LOADM_(hmtx, 1)
	// 	if !error {
	// 		face.vertical_info = 1
	// 	}
	// }
	// if error && FT_ERR_NEQ(error, Table_Missing) {
	// 	goto Exit
	// }

	out.os2, _ = f.OS2Table() // we treat the table as missing if there are any errors
	return out, nil
}

// TODO: handle the error in a first processing step (distinct from Parse though)
func (f *Font) Style() (isItalic, isBold bool, familyName, styleName string) {
	details, _ := f.analyze()

	// Bit 8 of the `fsSelection' field in the `OS/2' table denotes
	// a WWS-only font face.  `WWS' stands for `weight', width', and
	// `slope', a term used by Microsoft's Windows Presentation
	// Foundation (WPF).  This flag has been introduced in version
	// 1.5 of the OpenType specification (May 2008).

	if details.os2 != nil && details.os2.FsSelection&256 != 0 {
		familyName = f.Name.getName(NamePreferredFamily)
		if familyName == "" {
			familyName = f.Name.getName(NameFontFamily)
		}

		styleName = f.Name.getName(NamePreferredSubfamily)
		if styleName == "" {
			styleName = f.Name.getName(NameFontSubfamily)
		}
	} else {
		familyName = f.Name.getName(NameWWSFamily)
		if familyName == "" {
			familyName = f.Name.getName(NamePreferredFamily)
		}
		if familyName == "" {
			familyName = f.Name.getName(NameFontFamily)
		}

		styleName = f.Name.getName(NameWWSSubfamily)
		if styleName == "" {
			styleName = f.Name.getName(NamePreferredSubfamily)
		}
		if styleName == "" {
			styleName = f.Name.getName(NameFontSubfamily)
		}
	}

	styleName = strings.TrimSpace(styleName)
	if styleName == "" { // assume `Regular' style because we don't know better
		styleName = "Regular"
	}

	// Compute style flags.
	if details.hasOutline && details.os2 != nil {
		// We have an OS/2 table; use the `fsSelection' field.  Bit 9
		// indicates an oblique font face.  This flag has been
		// introduced in version 1.5 of the OpenType specification.
		isItalic = details.os2.FsSelection&(1<<9) != 0 || details.os2.FsSelection&1 != 0
		isBold = details.os2.FsSelection&(1<<5) != 0
	} else if details.head != nil { // TODO: remove when error is handled
		// this is an old Mac font, use the header field
		isBold = details.head.MacStyle&1 != 0
		isItalic = details.head.MacStyle&2 != 0
	}

	return
}

func (f *Font) GlyphKind() (scalable, bitmap, color bool) {
	// TODO: support for bitmap
	details, _ := f.analyze()
	return details.hasOutline, false, details.hasColor
}
