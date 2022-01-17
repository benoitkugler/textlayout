package type1

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

var _ fonts.FontDescriptor = (*fontDescriptor)(nil)

// only parses the ASCII segment, defering
// the charstring extraction to LoadCmap
type fontDescriptor struct {
	info     fonts.PSInfo
	encoding *simpleencodings.Encoding
}

// ScanFont lazily parse `file` to extract the information about the font.
// If no error occurs, the returned slice has always length 1.
func ScanFont(file fonts.Resource) ([]fonts.FontDescriptor, error) {
	seg1, _, err := openPfb(file)
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb font file: %s", err)
	}
	font, err := parse(seg1, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid .pfb font file: %s", err)
	}

	fd := fontDescriptor{
		info:     font.PSInfo,
		encoding: font.Encoding,
	}

	return []fonts.FontDescriptor{&fd}, nil
}

func (fd *fontDescriptor) Family() string {
	return fd.info.FamilyName
}

func (fd *fontDescriptor) AdditionalStyle() string {
	// ported from freetype/src/type1/t1objs.c
	var styleName string

	familyName := fd.info.FamilyName
	if familyName != "" {
		full := fd.info.FullName

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
		styleName = strings.TrimSpace(fd.info.Weight)
	}
	if styleName == "" { // assume `Regular' style because we don't know better
		styleName = "Regular"
	}
	return styleName
}

func (fd *fontDescriptor) Aspect() (fonts.Style, fonts.Weight, fonts.Stretch) {
	style := fonts.StyleNormal
	if isItalic := fd.info.ItalicAngle != 0; isItalic {
		style = fonts.StyleItalic
	}

	var weight fonts.Weight
	switch fd.info.Weight {
	case "Bold":
		weight = fonts.WeightBold
	case "Black":
		weight = fonts.WeightBlack
	}

	return style, weight, 0
}

// LoadCmap returns a cmap whose GID are invalid, but
// whose runes are correct, which is good enough to build coverage information.
func (fd *fontDescriptor) LoadCmap() (fonts.Cmap, error) {
	out := fonts.CmapSimple{}
	for r := range fd.encoding.RuneToByte() {
		out[r] = 1
	}
	return out, nil
}
