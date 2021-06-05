package type1

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/glyphsnames"
	ps "github.com/benoitkugler/textlayout/fonts/psinterpreter"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

var Loader fonts.FontLoader = loader{}

var _ fonts.Face = (*Font)(nil)

type loader struct{}

// Load implements fonts.FontLoader. When the error is `nil`,
// one (and only one) font is returned.
func (loader) Load(file fonts.Resource) (fonts.Fonts, error) {
	f, err := Parse(file)
	if err != nil {
		return nil, err
	}
	return fonts.Fonts{f}, nil
}

// Parse parses an Adobe Type 1 (.pfb) font file.
// See `ParseAFMFile` to read the associated Adobe font metric file.
func Parse(pfb fonts.Resource) (*Font, error) {
	seg1, seg2, err := openPfb(pfb)
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb font file: %s", err)
	}
	font, err := parse(seg1, seg2)
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb font file: %s", err)
	}

	// we follow freetype by placing the .notdef glyph at GID 0
	// this is not visible from the outside since the cmap will be
	// changed accordingly
	font.checkAndSwapGlyphNotdef()

	font.synthetizeCmap()

	return &font, nil
}

type charstring struct {
	name string
	data []byte
}

// Font exposes the content of a .pfb file.
// The main field, regarding PDF processing, is the Encoding
// entry, which defines the "builtin encoding" of the font.
type Font struct {
	Encoding *simpleencodings.Encoding
	cmap     fonts.CmapSimple // see synthetizeCmap

	FontID      string
	FontBBox    []Fl
	subrs       [][]byte     // local subroutines
	charstrings []charstring // slice indexed by glyph index
	FontMatrix  []Fl

	fonts.PSInfo

	StrokeWidth Fl

	PaintType int
	FontType  int
	UniqueID  int
}

func (f *Font) PostscriptInfo() (fonts.PSInfo, bool) { return f.PSInfo, true }

func (f *Font) PoscriptName() string { return f.PSInfo.FontName }

func (f *Font) getStyle() (isItalic, isBold bool, familyName, styleName string) {
	// ported from freetype/src/type1/t1objs.c

	// get style name -- be careful, some broken fonts only
	// have a `/FontName' dictionary entry!
	familyName = f.PSInfo.FamilyName
	if familyName != "" {
		full := f.PSInfo.FullName

		theSame := true

		for i, j := 0, 0; i < len(full); {
			if j < len(familyName) && full[i] == familyName[j] {
				i++
				j++
			} else {
				if full[i] == ' ' || full[i] == '-' {
					i++
				} else if j < len(familyName) && (familyName[j] == ' ' || familyName[j] == '-') {
					j++
				} else {
					theSame = false
					if j == len(familyName) {
						styleName = full[i:]
					}
					break
				}
			}
		}

		if theSame {
			styleName = "Regular"
		}
	}

	styleName = strings.TrimSpace(styleName)
	if styleName == "" {
		styleName = strings.TrimSpace(f.PSInfo.Weight)
	}
	if styleName == "" { // assume `Regular' style because we don't know better
		styleName = "Regular"
	}

	isItalic = f.PSInfo.ItalicAngle != 0
	isBold = f.PSInfo.Weight == "Bold" || f.PSInfo.Weight == "Black"
	return
}

func (f *Font) LoadMetrics() fonts.FaceMetrics { return f }

func (f *Font) LoadSummary() (fonts.FontSummary, error) {
	isItalic, isBold, familyName, styleName := f.getStyle()
	return fonts.FontSummary{
		IsItalic:          isItalic,
		IsBold:            isBold,
		Familly:           familyName,
		Style:             styleName,
		HasScalableGlyphs: true,
		HasBitmapGlyphs:   false,
		HasColorGlyphs:    false,
	}, nil
}

// font metrics

var _ fonts.FaceMetrics = (*Font)(nil)

// Upem reads the FontMatrix to extract the scaling factor (the maximum between x and y coordinates)
func (f *Font) Upem() uint16 {
	if len(f.FontMatrix) < 4 {
		return 1000 // typical value for Type1 fonts
	}
	xx, yy := math.Abs(f.FontMatrix[0]), math.Abs(f.FontMatrix[3])
	var (
		upemX uint16 = 1000
		upemY        = upemX
	)
	if xx != 0 {
		upemX = uint16(1 / xx)
	}
	if yy != 0 {
		upemY = uint16(1 / yy)
	}
	if upemX > upemY {
		return upemX
	}
	return upemY
}

func (f *Font) GlyphName(gid fonts.GID) string {
	if int(gid) >= len(f.charstrings) {
		return ""
	}
	return f.charstrings[gid].name
}

func (f *Font) LineMetric(metric fonts.LineMetric, _ []float32) (float32, bool) {
	switch metric {
	case fonts.UnderlinePosition:
		return float32(f.PSInfo.UnderlinePosition), true
	case fonts.UnderlineThickness:
		return float32(f.PSInfo.UnderlineThickness), true
	default:
		return 0, false
	}
}

func (f *Font) FontHExtents(_ []float32) (fonts.FontExtents, bool) {
	var extents fonts.FontExtents
	if len(f.FontBBox) < 4 {
		return extents, false
	}
	yMin, yMax := f.FontBBox[1], f.FontBBox[3]
	// following freetype here
	extents.Ascender = float32(yMax)
	extents.Descender = float32(yMin)

	extents.LineGap = float32(f.Upem()) * 1.2
	if extents.LineGap < extents.Ascender-extents.Descender {
		extents.LineGap = extents.Ascender - extents.Descender
	}
	return extents, true
}

// FontVExtents returns zero values.
func (f *Font) FontVExtents(_ []float32) (fonts.FontExtents, bool) {
	return fonts.FontExtents{}, false
}

// fix the potentail misplaced .notdef glyphs
func (f *Font) checkAndSwapGlyphNotdef() {
	if len(f.charstrings) == 0 || f.charstrings[0].name == Notdef {
		return
	}

	for i, v := range f.charstrings {
		if v.name == Notdef {
			f.charstrings[0], f.charstrings[i] = f.charstrings[i], f.charstrings[0]
			break
		}
	}
}

// Type1 fonts have no natural notion of Unicode code points
// We use a glyph names table to identify the most commonly used runes
func (f *Font) synthetizeCmap() {
	f.cmap = make(map[rune]fonts.GID)
	for gid, charstring := range f.charstrings {
		glyphName := charstring.name
		r, _ := glyphsnames.GlyphToRune(glyphName)
		f.cmap[r] = fonts.GID(gid)
	}
}

func (f *Font) NominalGlyph(ch rune) (fonts.GID, bool) {
	out, ok := f.cmap[ch]
	return out, ok
}

// VariationGlyph is not supported by Type1 fonts.
func (f *Font) VariationGlyph(ch, varSelector rune) (fonts.GID, bool) { return 0, false }

// parseGlyphMetrics returns the advance of the glyph with index `index`
// The return value is expressed in font units.
// An error is returned for invalid index values and for invalid
// charstring glyph data.
// inSeac is used to check for recursion in seac glyphs
// initialPoint is used when parsing a seac accent, it should be (0,0) otherwise
func (f *Font) parseGlyphMetrics(index fonts.GID, inSeac bool) (ps.PathBounds, int32, error) {
	if int(index) >= len(f.charstrings) {
		return ps.PathBounds{}, 0, errors.New("invalid glyph index")
	}

	var (
		psi    ps.Machine
		parser type1CharstringParser
	)
	err := psi.Run(f.charstrings[index].data, f.subrs, nil, &parser)
	if err != nil {
		return ps.PathBounds{}, 0, err
	}
	// handle the special case of seac glyph
	if parser.seac != nil {
		if inSeac {
			return ps.PathBounds{}, 0, errors.New("invalid nested seac operator")
		}
		var bounds ps.PathBounds
		bounds, err = f.seacMetrics(*parser.seac)
		if err != nil {
			return ps.PathBounds{}, 0, err
		}
		return bounds, parser.advance.X, err
	}
	return parser.cs.Bounds, parser.advance.X, err
}

func (f *Font) seacMetrics(seac seac) (ps.PathBounds, error) {
	aGlyph, err := f.glyphIndexFromStandardCode(seac.aCode)
	if err != nil {
		return ps.PathBounds{}, err
	}
	bGlyph, err := f.glyphIndexFromStandardCode(seac.bCode)
	if err != nil {
		return ps.PathBounds{}, err
	}
	boundsBase, _, err := f.parseGlyphMetrics(bGlyph, true)
	if err != nil {
		return ps.PathBounds{}, err
	}

	boundsAccent, _, err := f.parseGlyphMetrics(aGlyph, true)
	if err != nil {
		return ps.PathBounds{}, err
	}

	// translate the accent
	// See the erratum https://adobe-type-tools.github.io/font-tech-notes/pdfs/5015.Type1_Supp.pdf
	offsetOriginX := boundsBase.Min.X - boundsAccent.Min.X + seac.accentOrigin.X
	offsetOriginY := seac.accentOrigin.Y
	boundsAccent.Min.Move(offsetOriginX, offsetOriginY)
	boundsAccent.Max.Move(offsetOriginX, offsetOriginY)

	// union with the base
	boundsBase.Enlarge(boundsAccent.Min)
	boundsBase.Enlarge(boundsAccent.Max)

	return boundsBase, nil
}

func (f *Font) glyphIndexFromStandardCode(code int32) (fonts.GID, error) {
	if code < 0 || int(code) > len(simpleencodings.AdobeStandard) {
		return 0, fmt.Errorf("invalid char code in seac: %d", code)
	}
	glyphName := simpleencodings.AdobeStandard[code]
	for gid, charstring := range f.charstrings {
		if charstring.name == glyphName {
			return fonts.GID(gid), nil
		}
	}
	return 0, fmt.Errorf("unknown glyph name in seac: %s", glyphName)
}

// HorizontalAdvance returns the advance of the glyph with index `index`
// The return value is expressed in font units.
// 0 is returned for invalid index values and for invalid
// charstring glyph data.
func (f *Font) HorizontalAdvance(gid fonts.GID, _ []float32) float32 {
	_, adv, err := f.parseGlyphMetrics(gid, false)
	if err != nil {
		return 0
	}
	return float32(adv)
}

func (f *Font) VerticalAdvance(gid fonts.GID, _ []float32) float32 { return 0 }

// GlyphHOrigin always return 0,0,true
func (Font) GlyphHOrigin(fonts.GID, []float32) (x, y fonts.Position, found bool) {
	return 0, 0, true
}

// GlyphVOrigin always return 0,0,false
func (Font) GlyphVOrigin(fonts.GID, []float32) (x, y fonts.Position, found bool) {
	return 0, 0, false
}

func (f *Font) GlyphExtents(glyph fonts.GID, _ []float32, _, _ uint16) (fonts.GlyphExtents, bool) {
	bbox, _, err := f.parseGlyphMetrics(glyph, false)
	if err != nil {
		return fonts.GlyphExtents{}, false
	}
	return fonts.GlyphExtents{
		XBearing: float32(bbox.Min.X),
		YBearing: float32(bbox.Max.Y),
		Width:    float32(bbox.Max.X - bbox.Min.X),
		Height:   float32(bbox.Min.Y - bbox.Max.Y),
	}, true
}

func (Font) NormalizeVariations(coords []float32) []float32 { return coords }
