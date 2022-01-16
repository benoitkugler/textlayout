package type1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/glyphsnames"
	ps "github.com/benoitkugler/textlayout/fonts/psinterpreter"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

var _ fonts.Face = (*Font)(nil)

type loader struct{}

// Load implements fonts.FontLoader. When the error is `nil`,
// one (and only one) font is returned.
func Load(file fonts.Resource) (fonts.Faces, error) {
	f, err := Parse(file)
	if err != nil {
		return nil, err
	}
	return fonts.Faces{f}, nil
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

// LoadMetrics returns the font itself.
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

// loadGlyph returns the advance of the glyph with index `index`
// The return value is expressed in font units.
// An error is returned for invalid index values and for invalid
// charstring glyph data.
// inSeac is used to check for recursion in seac glyphs
func (f *Font) loadGlyph(index fonts.GID, inSeac bool) ([]fonts.Segment, ps.PathBounds, int32, error) {
	if int(index) >= len(f.charstrings) {
		return nil, ps.PathBounds{}, 0, errors.New("invalid glyph index")
	}

	var (
		psi    ps.Machine
		parser type1CharstringParser
	)
	err := psi.Run(f.charstrings[index].data, f.subrs, nil, &parser)
	if err != nil {
		return nil, ps.PathBounds{}, 0, err
	}
	// handle the special case of seac glyph
	if parser.seac != nil {
		if inSeac {
			return nil, ps.PathBounds{}, 0, errors.New("invalid nested seac operator")
		}
		var (
			bounds   ps.PathBounds
			segments []fonts.Segment
		)
		segments, bounds, err = f.seacMetrics(*parser.seac)
		if err != nil {
			return nil, ps.PathBounds{}, 0, err
		}
		return segments, bounds, parser.advance.X, err
	}
	return parser.cs.Segments, parser.cs.Bounds, parser.advance.X, err
}

func (f *Font) seacMetrics(seac seac) ([]fonts.Segment, ps.PathBounds, error) {
	aGlyph, err := f.glyphIndexFromStandardCode(seac.aCode)
	if err != nil {
		return nil, ps.PathBounds{}, err
	}
	bGlyph, err := f.glyphIndexFromStandardCode(seac.bCode)
	if err != nil {
		return nil, ps.PathBounds{}, err
	}
	segmentsBase, boundsBase, _, err := f.loadGlyph(bGlyph, true)
	if err != nil {
		return nil, ps.PathBounds{}, err
	}

	segmentsAccent, boundsAccent, _, err := f.loadGlyph(aGlyph, true)
	if err != nil {
		return nil, ps.PathBounds{}, err
	}

	// translate the accent
	// See the erratum https://adobe-type-tools.github.io/font-tech-notes/pdfs/5015.Type1_Supp.pdf
	offsetOriginX := boundsBase.Min.X - boundsAccent.Min.X + seac.accentOrigin.X
	offsetOriginY := seac.accentOrigin.Y
	boundsAccent.Min.Move(offsetOriginX, offsetOriginY)
	boundsAccent.Max.Move(offsetOriginX, offsetOriginY)
	offsetOriginXF, offsetOriginYF := float32(offsetOriginX), float32(offsetOriginY)
	for i := range segmentsAccent {
		argsSlice := segmentsAccent[i].ArgsSlice()
		for j := range argsSlice {
			argsSlice[j].Move(offsetOriginXF, offsetOriginYF)
		}
	}

	// union with the base
	boundsBase.Enlarge(boundsAccent.Min)
	boundsBase.Enlarge(boundsAccent.Max)
	segmentsBase = append(segmentsBase, segmentsAccent...)

	return segmentsBase, boundsBase, nil
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

func (Font) LoadBitmaps() []fonts.BitmapSize { return nil }
