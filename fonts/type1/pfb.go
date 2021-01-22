package type1

import (
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

var Loader fonts.FontLoader = loader{}

var _ fonts.Font = (*Font)(nil)

type loader struct{}

// Load implements fonts.FontLoader. When the error is `nil`,
// one (and only one) font is returned.
func (loader) Load(file fonts.Ressource) (fonts.Fonts, error) {
	f, err := Parse(file)
	if err != nil {
		return nil, err
	}
	return fonts.Fonts{f}, nil
}

// Parse parses an Adobe Type 1 (.pfb) font file.
// See `ParseAFMFile` to read the associated Adobe font metric file.
func Parse(pfb fonts.Ressource) (*Font, error) {
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

func (f *Font) Style() (isItalic, isBold bool, styleName string) {
	/* The following code to extract the family and the style is very   */
	/* simplistic and might get some things wrong.  For a full-featured */
	/* algorithm you might have a look at the whitepaper given at       */
	/*                                                                  */
	/*   https://blogs.msdn.com/text/archive/2007/04/23/wpf-font-selection-model.aspx */

	/* get style name -- be careful, some broken fonts only */
	/* have a `/FontName' dictionary entry!                 */
	familyName := f.PSInfo.FamilyName

	if familyName != "" {
		full := f.PSInfo.FullName
		// char*  family = root.familyName;

		theSame := true

		for i, j := 0, 0; i < len(full); {
			if full[i] == familyName[j] {
				i++
				j++
			} else {
				if full[i] == ' ' || full[i] == '-' {
					i++
				} else if familyName[j] == ' ' || familyName[j] == '-' {
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
	} else {
		// do we have a `/FontName'?
		if f.PSInfo.FontName != "" {
			familyName = f.PSInfo.FontName
		}
	}

	if styleName == "" {
		if f.PSInfo.Weight != "" {
			styleName = f.PSInfo.Weight
		} else {
			// assume `Regular' style because we don't know better
			styleName = "Regular"
		}
	}

	isItalic = f.PSInfo.ItalicAngle != 0
	isBold = f.PSInfo.Weight == "Bold" || f.PSInfo.Weight == "Black"
	return
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
		psi     psinterpreter.Inter
		handler type1Metrics
	)
	err := psi.Run(f.charstrings[index].data, nil, &handler)
	return handler.advance.X, err
}
