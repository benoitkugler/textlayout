// Package fonts provides supports for parsing
// several font formats (postscript, bitmap and truetype)
// and provides a common API.
// It does not support CIDType1 fonts.
package fonts

// PSInfo exposes global properties of a postscript font.
type PSInfo struct {
	FontName    string // Postscript font name.
	FullName    string // full name of the font.
	FamilyName  string // family name of the font.
	Version     string // font program version identifier (optional)
	Notice      string // font name trademark or copyright notice (optional)
	Weight      string // Weight of the font: normal, bold, etc.
	ItalicAngle int    // italic angle of the font, usually 0. or negative.

	IsFixedPitch bool // true if all the characters have the same width.

	UnderlinePosition  int
	UnderlineThickness int
}

// Font provides a unified access to various font formats.
type Font interface {
	PostscriptInfo() (PSInfo, bool)
	// PoscriptName returns the PoscriptName of the font,
	// or an empty string.
	PoscriptName() string
	// Style return the basic information about the
	// style of the font
	Style() (isItalic, isBold bool, style string)
}

// GlyphIndex is used to identify glyphs in a font.
// It is internal to the font and should be confused with
// Unicode code points.
type GlyphIndex uint16

// Ressource is a combination of io.Reader, io.Seeker and io.ReaderAt.
// This interface is satisfied by most things that you'd want
// to parse, for example *os.File, io.SectionReader or *bytes.Buffer.
type Ressource interface {
	Read([]byte) (int, error)
	ReadAt([]byte, int64) (int, error)
	Seek(int64, int) (int64, error)
}
