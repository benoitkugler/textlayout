package type1

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

// var Loader fonts.FontLoader = loader{}

// var _ fonts.Font = (*Font)(nil)

type loader struct{}

// Load implements fonts.FontLoader. When the error is `nil`,
// one (and only one) font is returned.
// func (loader) Load(file fonts.Resource) (fonts.Fonts, error) {
// 	f, err := Parse(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return fonts.Fonts{f}, nil
// }

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
	cmap     map[rune]fonts.GID // see synthetizeCmap

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

func (f *Font) Style() (isItalic, isBold bool, familyName, styleName string) {
	// ported from freetype/src/type1/t1objs.c

	/* get style name -- be careful, some broken fonts only */
	/* have a `/FontName' dictionary entry!                 */
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

func (Font) GlyphKind() (scalable, bitmap, color bool) {
	return true, false, false
}

// font metrics

// var _ fonts.FontMetrics = (*Font)(nil)

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

// Type1 fonts have no natural notion of Unicode code points
// We use a glyph names table to identify the most commonly used runes
func (f *Font) synthetizeCmap() {
	m := f.Encoding.NameToRune()
	f.cmap = make(map[rune]fonts.GID, len(m))
	for gid, charstring := range f.charstrings {
		glyphName := charstring.name
		r := m[glyphName]
		f.cmap[r] = fonts.GID(gid)
	}
}

func (f *Font) NominalGlyph(ch rune) (fonts.GID, bool) {
	out, ok := f.cmap[ch]
	return out, ok
}

// VariationGlyph is not supported by Type1 fonts.
func (f *Font) VariationGlyph(ch, varSelector rune) (fonts.GID, bool) { return 0, false }

// getHAdvance returns the advance of the glyph with index `index`
// The return value is expressed in font units.
// An error is returned for invalid index values and for invalid
// charstring glyph data.
func (f *Font) getHAdvance(index fonts.GID) (int32, error) {
	if int(index) >= len(f.charstrings) {
		return 0, errors.New("invalid glyph index")
	}
	var (
		psi     psinterpreter.Machine
		handler type1Metrics
	)
	err := psi.Run(f.charstrings[index].data, f.subrs, nil, &handler)
	return handler.advance.X, err
}

// HorizontalAdvance returns the advance of the glyph with index `index`
// The return value is expressed in font units.
// 0 is returned for invalid index values and for invalid
// charstring glyph data.
func (f *Font) HorizontalAdvance(gid fonts.GID, _ []float32) float32 {
	adv, err := f.getHAdvance(gid)
	if err != nil {
		return 0
	}
	return float32(adv)
}

func (f *Font) VerticalAdvance(gid fonts.GID, _ []float32) float32 { return 0 }

// GlyphHOrigin always return 0,0
func (f *Font) GlyphHOrigin(fonts.GID, []float32) (x, y fonts.Position, found bool) {
	return 0, 0, true
}
