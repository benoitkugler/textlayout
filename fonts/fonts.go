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

// FontSummary stores basic informations about the
// style of the font.
type FontSummary struct {
	Familly          string
	Style            string
	IsItalic, IsBold bool

	HasScalableGlyphs, HasBitmapGlyphs, HasColorGlyphs bool
}

// FaceLayout gives acces to layout related informations.
type FaceLayout interface {
	// LoadMetrics fetches all the informations related to the font metrics.
	// Conceptually, this method just returns it receiver, but this enables lazy loading.
	LoadMetrics() FaceMetrics
}

// Face provides a unified access to various font formats.
// It describes the content of one font from a font file.
type Face interface {
	FaceLayout

	PostscriptInfo() (PSInfo, bool)

	// PoscriptName returns the PoscriptName of the font,
	// or an empty string.
	PoscriptName() string

	// LoadSummary fetchs global details about the font.
	// Conceptually, this method just returns it receiver, but this enables lazy loading.
	LoadSummary() (FontSummary, error)
}

// Fonts is the parsed content of a font ressource.
// Note that variable fonts are not repeated in this slice,
// since instances are accessed on each font.
type Fonts []Face

// FontLoader implements the general parsing
// of a font file. Some font format support to store several
// fonts inside one file. For the other formats, the returned slice will
// have length 1.
type FontLoader interface {
	Load(file Resource) (Fonts, error)
}

// GID is used to identify glyphs in a font.
// It is mostly internal to the font and should not be confused with
// Unicode code points.
type GID uint16

// CmapIter is an interator over a Cmap.
type CmapIter interface {
	// Next returns true if the iterator still has data to yield
	Next() bool

	// Char must be called only when `Next` has returned `true`
	Char() (rune, GID)
}

// Cmap stores a compact representation of a cmap,
// offering both on-demand rune lookup and full rune range.
// It is conceptually equivalent to a map[rune]GID, but is often
// implemented for efficiently.
type Cmap interface {
	// Iter returns a new iterator over the cmap
	// Multiple iterators may be used over the same cmap
	// The returned interface is garanted not to be nil
	Iter() CmapIter

	// Lookup avoid the construction of a map and provides
	// an alternative when only few runes need to be fetched.
	// It returns a default value when no glyph is provided.
	Lookup(rune) GID
}

var (
	_ Cmap     = CmapSimple(nil)
	_ CmapIter = (*cmap0Iter)(nil)
)

// CmapSimple is a map based Cmap implementation.
type CmapSimple map[rune]GID

type cmap0Iter struct {
	data CmapSimple
	keys []rune
	pos  int
}

func (it *cmap0Iter) Next() bool {
	return it.pos < len(it.keys)
}

func (it *cmap0Iter) Char() (rune, GID) {
	r := it.keys[it.pos]
	it.pos++
	return r, it.data[r]
}

func (s CmapSimple) Iter() CmapIter {
	keys := make([]rune, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return &cmap0Iter{data: s, keys: keys}
}

func (s CmapSimple) Lookup(r rune) GID {
	return s[r] // will be 0 if r is not in s
}

// Position is expressed in font units.
type Position = int32

// FontExtents exposes font-wide extent values, measured in font units.
// Note that typically ascender is positive and descender negative in coordinate systems that grow up.
type FontExtents struct {
	Ascender  float32 // Typographic ascender.
	Descender float32 // Typographic descender.
	LineGap   float32 // Suggested line spacing gap.
}

type LineMetric uint8

const (
	// Distance above the baseline of the top of the underline.
	// Since most fonts have underline positions beneath the baseline, this value is typically negative.
	UnderlinePosition LineMetric = iota

	// Suggested thickness to draw for the underline.
	UnderlineThickness

	// Distance above the baseline of the top of the strikethrough.
	StrikethroughPosition

	// Suggested thickness to draw for the strikethrough.
	StrikethroughThickness
)

// GlyphExtents exposes extent values, measured in font units.
// Note that height is negative in coordinate systems that grow up.
type GlyphExtents struct {
	XBearing float32 // Left side of glyph from origin
	YBearing float32 // Top side of glyph from origin
	Width    float32 // Distance from left to right side
	Height   float32 // Distance from top to bottom side
}

// FaceMetrics exposes details of the font content.
// It is distinct from the `Face`interface to allow lazy loading.
type FaceMetrics interface {
	// Upem returns the units per em of the font file.
	// If not found, should return 1000 as fallback value.
	// This value is only relevant for scalable fonts.
	Upem() uint16

	// GlyphName returns the name of the given glyph, or an empty
	// string if the glyph is invalid or has no name.
	GlyphName(gid GID) string

	// LineMetric returns the metric identified by `metric` (in fonts units), or false.
	//  `varCoords` (in normalized coordinates) is only useful for variable fonts.
	LineMetric(metric LineMetric, varCoords []float32) (float32, bool)

	// FontHExtents returns the extents of the font for horizontal text, or false
	// it not available, in font units.
	// `varCoords` (in normalized coordinates) is only useful for variable fonts.
	FontHExtents(varCoords []float32) (FontExtents, bool)

	// FontVExtents is the same as `FontHExtents`, but for vertical text.
	FontVExtents(varCoords []float32) (FontExtents, bool)

	// NominalGlyph returns the glyph used to represent the given rune,
	// or false if not found.
	NominalGlyph(ch rune) (GID, bool)

	// VariationGlyph retrieves the glyph ID for a specified Unicode code point
	// followed by a specified Variation Selector code point, or false if not found
	VariationGlyph(ch, varSelector rune) (GID, bool)

	// HorizontalAdvance returns the horizontal advance in font units.
	// When no data is available but the glyph index is valid, this method
	// should return a default value (the upem number for example).
	// If the glyph is invalid it should return 0.
	// `coords` is used by variable fonts, and is specified in normalized coordinates.
	HorizontalAdvance(gid GID, coords []float32) float32

	// VerticalAdvance is the same as `HorizontalAdvance`, but for vertical advance.
	VerticalAdvance(gid GID, coords []float32) float32

	// GlyphHOrigin fetches the (X,Y) coordinates of the origin (in font units) for a glyph ID,
	// for horizontal text segments.
	// Returns `false` if not available.
	GlyphHOrigin(GID, []float32) (x, y Position, found bool)

	// GlyphVOrigin is the same as `GlyphHOrigin`, but for vertical text segments.
	GlyphVOrigin(GID, []float32) (x, y Position, found bool)

	// GlyphExtents retrieve the extents for a specified glyph, of false, if not available.
	// `coords` is used by variable fonts, and is specified in normalized coordinates.
	// For bitmap glyphs, the closest resolution to `xPpem` and `yPpem` is selected.
	GlyphExtents(glyph GID, coords []float32, xPpem, yPpem uint16) (GlyphExtents, bool)

	// NormalizeVariations should normalize the given design-space coordinates. The minimum and maximum
	// values for the axis are mapped to the interval [-1,1], with the default
	// axis value mapped to 0.
	// This should be a no-op for non-variable fonts.
	NormalizeVariations(coords []float32) []float32
}
