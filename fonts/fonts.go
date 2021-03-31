// Package fonts provides supports for parsing
// several font formats (postscript, bitmap and truetype)
// and provides a common API.
// It does not support CIDType1 fonts.
package fonts

// Resource is a combination of io.Reader, io.Seeker and io.ReaderAt.
// This interface is satisfied by most things that you'd want
// to parse, for example *os.File, io.SectionReader or *bytes.Buffer.
type Resource interface {
	Read([]byte) (int, error)
	ReadAt([]byte, int64) (int, error)
	Seek(int64, int) (int64, error)
}

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
	// style of the font.
	// `style` default to 'Regular' if not found
	Style() (isItalic, isBold bool, familly, style string)

	// GlyphKind return the different kind of glyphs present in the font.
	// Note that a font can contain both scalable glyphs (outlines) and bitmap strikes
	GlyphKind() (scalable, bitmap, color bool)

	// LoadMetrics fetches all the informations related to the font metrics.
	// Conceptually, this method just return it receiver. However, in pratice,
	// this enable lazy loading.
	LoadMetrics() FontMetrics
}

// Fonts is the parsed content of a font ressource.
// Not that variable fonts are not repeated in this slice,
// since instances are accessed on each font.
type Fonts []Font

// FontLoader implements the general parsing
// of a font file. Some font format support to store several
// fonts inside one file. For the other formats, the returned slice will
// have length 1.
type FontLoader interface {
	Load(file Resource) (Fonts, error)
}

// GlyphIndex is used to identify glyphs in a font.
// It is mostly internal to the font and should not be confused with
// Unicode code points.
type GlyphIndex uint16

// Position is expressed in font units
type Position = int32

// FontExtents exposes font-wide extent values, measured in font units.
// Note that typically ascender is positive and descender negative in coordinate systems that grow up.
type FontExtents struct {
	Ascender  float32 // Typographic ascender.
	Descender float32 // Typographic descender.
	LineGap   float32 // Suggested line spacing gap.
}

// GlyphExtents exposes extent values, measured in font units.
// Note that height is negative in coordinate systems that grow up.
type GlyphExtents struct {
	XBearing float32 // Left side of glyph from origin
	YBearing float32 // Top side of glyph from origin
	Width    float32 // Distance from left to right side
	Height   float32 // Distance from top to bottom side
}

// FontMetrics exposes details of the font content.
// It is distinct from the `Font`interface to allow lazy loading.
type FontMetrics interface {
	// Returns the units per em of the font file.
	// If not found, should return 1000 as fallback value.
	GetUpem() uint16

	// Returns the extents of the font for horizontal text, or false
	// it not available, in font units.
	// `varCoords` (in normalized coordinates) is only useful for variable fonts.
	GetFontHExtents(varCoords []float32) (FontExtents, bool)

	// Return the glyph used to represent the given rune,
	// or false if not found.
	GetNominalGlyph(ch rune) (GlyphIndex, bool)

	// Retrieves the glyph ID for a specified Unicode code point
	// followed by a specified Variation Selector code point, or false if not found
	GetVariationGlyph(ch, varSelector rune) (GlyphIndex, bool)

	// Returns the horizontal advance in font units.
	// When no data is available but the glyph index is valid, this method
	// should return a default value (the upem number for example).
	// If the glyph is invalid it should return 0.
	// `coords` is used by variable fonts, and is specified in normalized coordinates.
	GetHorizontalAdvance(gid GlyphIndex, coords []float32) float32

	// Same as `GetHorizontalAdvance`, but for vertical advance.
	GetVerticalAdvance(gid GlyphIndex, coords []float32) float32

	// Fetches the (X,Y) coordinates of the origin (in font units) for a glyph ID,
	// for horizontal text segments.
	// Returns `false` if not available.
	GetGlyphHOrigin(GlyphIndex, []float32) (x, y Position, found bool)

	// Same as `GetGlyphHOrigin`, but for vertical text segments.
	GetGlyphVOrigin(GlyphIndex, []float32) (x, y Position, found bool)

	// Retrieve the extents for a specified glyph, of false, if not available.
	// `coords` is used by variable fonts, and is specified in normalized coordinates.
	// `xPpem` and `yPpem` are only used for bitmap glyphs
	GetGlyphExtents(glyph GlyphIndex, coords []float32, xPpem, yPpem uint16) (GlyphExtents, bool)

	// NormalizeVariations should normalize the given design-space coordinates. The minimum and maximum
	// values for the axis are mapped to the interval [-1,1], with the default
	// axis value mapped to 0.
	// This should be a no-op for non-variable fonts.
	NormalizeVariations(coords []float32) []float32
}
