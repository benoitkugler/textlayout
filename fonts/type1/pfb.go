package type1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

var Loader fonts.FontLoader = loader{}

var _ fonts.Font = (*Font)(nil)

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
	fonts.PSInfo

	Encoding Encoding

	PaintType   int
	FontType    int
	UniqueID    int
	StrokeWidth Fl
	FontID      string
	FontMatrix  []Fl
	FontBBox    []Fl

	subrs       [][]byte
	charstrings []charstring // slice indexed by glyph index
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

// GetAdvance returns the advance of the glyph with index `index`
// The return value is expressed in font units.
// An error is returned for invalid index values and for invalid
// charstring glyph data.
func (f *Font) GetAdvance(index fonts.GlyphIndex) (int32, error) {
	if int(index) >= len(f.charstrings) {
		return 0, errors.New("invalid glyph index")
	}
	var (
		psi     psinterpreter.Machine
		handler type1Metrics
	)
	err := psi.Run(f.charstrings[index].data, nil, nil, &handler)
	return handler.advance.X, err
}
