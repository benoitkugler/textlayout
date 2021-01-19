package type1

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/benoitkugler/textlayout/fonts"
)

var _ fonts.Font = (*PFBFont)(nil)

const (
	// the pdf header length.
	// (start-marker (1 byte), ascii-/binary-marker (1 byte), size (4 byte))
	// 3*6 == 18
	pfbHeaderLength = 18

	// the start marker.
	startMarker = 0x80

	// the ascii marker.
	asciiMarker = 0x01

	// the binary marker.
	binaryMarker = 0x02
)

// The record types in the pfb-file.
var pfbRecords = [...]int{asciiMarker, binaryMarker, asciiMarker}

// PFBFont exposes the content of a .pfb file.
// The main field, regarding PDF processing, is the Encoding
// entry, which defines the "builtin encoding" of the font.
type PFBFont struct {
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
	charstrings map[string][]byte
}

func (f *PFBFont) PostscriptInfo() (fonts.PSInfo, bool) { return f.PSInfo, true }

func (f *PFBFont) PoscriptName() string { return f.PSInfo.FontName }

func (f *PFBFont) Style() (isItalic, isBold bool, styleName string) {
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

// --------------------- parser ---------------------

type stream struct {
	*bytes.Reader
}

func (s stream) read() int {
	c, err := s.Reader.ReadByte()
	if err != nil {
		return -1
	}
	return int(c)
}

// fetchs the segments of a .pfb font file.
func openPfb(pfb []byte) (segment1, segment2 []byte, err error) {
	in := stream{bytes.NewReader(pfb)}
	pfbdata := make([]byte, len(pfb)-pfbHeaderLength)
	var lengths [len(pfbRecords)]int
	pointer := 0
	for records := 0; records < len(pfbRecords); records++ {
		if in.read() != startMarker {
			return nil, nil, errors.New("Start marker missing")
		}

		if in.read() != pfbRecords[records] {
			return nil, nil, errors.New("Incorrect record type")
		}

		size := in.read()
		size += in.read() << 8
		size += in.read() << 16
		size += in.read() << 24
		lengths[records] = size
		if pointer >= len(pfbdata) {
			return nil, nil, errors.New("attempted to read past EOF")
		}
		inL := io.LimitedReader{R: in, N: int64(size)}
		got, err := inL.Read(pfbdata[pointer:])
		if err != nil {
			return nil, nil, err
		}
		pointer += got
	}

	return pfbdata[0:lengths[0]], pfbdata[lengths[0] : lengths[0]+lengths[1]], nil
}

// ParsePFBFile is a convenience wrapper, reading and
// parsing a .pfb font file.
func ParsePFBFile(pfb []byte) (PFBFont, error) {
	seg1, seg2, err := openPfb(pfb)
	if err != nil {
		return PFBFont{}, fmt.Errorf("invalid .pfb font file: %s", err)
	}
	font, err := Parse(seg1, seg2)
	if err != nil {
		return PFBFont{}, fmt.Errorf("invalid .pfb font file: %s", err)
	}
	return font, nil
}
