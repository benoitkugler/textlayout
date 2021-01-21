package type1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	tk "github.com/benoitkugler/pstokenizer"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

var _ fonts.Font = (*Font)(nil)

const (
	// start marker of a segment
	startMarker = 0x80

	// marker of the ascii segment
	asciiMarker = 0x01

	// marker of the binary segment
	binaryMarker = 0x02
)

// --------------------- parser ---------------------

func readOneRecord(pfb fonts.Ressource, expectedMarker byte, totalSize int64) ([]byte, error) {
	var buffer [6]byte

	_, err := pfb.Read(buffer[:])
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb file: missing record marker")
	}
	if buffer[0] != startMarker {
		return nil, errors.New("invalid .pfb file: start marker missing")
	}

	if buffer[1] != expectedMarker {
		return nil, errors.New("invalid .pfb file: incorrect record type")
	}

	size := int64(binary.LittleEndian.Uint32(buffer[2:]))
	if size >= totalSize {
		return nil, errors.New("corrupted .pfb file")
	}
	out := make([]byte, size)
	_, err = pfb.Read(out)
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb file: %s", err)
	}
	return out, nil
}

// fetchs the segments of a .pfb font file.
// see https://www.adobe.com/content/dam/acom/en/devnet/font/pdfs/5040.Download_Fonts.pdf
// IBM PC format
func openPfb(pfb fonts.Ressource) (segment1, segment2 []byte, err error) {
	totalSize, err := pfb.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, nil, err
	}
	_, err = pfb.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}

	// ascii record
	segment1, err = readOneRecord(pfb, asciiMarker, totalSize)
	if err != nil {
		// try with the brute force approach for file who have no tag
		segment1, segment2, err = seekMarkers(pfb)
		if err == nil {
			return segment1, segment2, nil
		}
		return nil, nil, err
	}

	// binary record
	segment2, err = readOneRecord(pfb, binaryMarker, totalSize)
	if err != nil {
		return nil, nil, err
	}
	// ignore the last segment, which is not needed

	return segment1, segment2, nil
}

const (
	headerT11 = "%!FontType"
	headerT12 = "%!PS-AdobeFont"
)

// fallback when no binary marker are present:
// we look for the currentfile exec pattern, then for the cleartomark
func seekMarkers(pfb fonts.Ressource) (segment1, segment2 []byte, err error) {
	_, err = pfb.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}

	// quickly return for invalid files
	var buffer [len(headerT12)]byte
	pfb.Read(buffer[:])
	if h := string(buffer[:]); !(strings.HasPrefix(h, headerT11) || strings.HasPrefix(h, headerT12)) {
		return nil, nil, errors.New("not a Type1 font file")
	}

	_, err = pfb.Seek(0, io.SeekStart)
	if err != nil {
		return nil, nil, err
	}
	data, err := ioutil.ReadAll(pfb)
	if err != nil {
		return nil, nil, err
	}
	const exec = "currentfile eexec"
	index := bytes.Index(data, []byte(exec))
	if index == -1 {
		return nil, nil, errors.New("not a Type1 font file")
	}
	segment1 = data[:index+len(exec)]
	segment2 = data[index+len(exec):]
	if len(segment2) != 0 && tk.IsAsciiWhitespace(segment2[0]) { // end of line
		segment2 = segment2[1:]
	}
	return segment1, segment2, nil
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
