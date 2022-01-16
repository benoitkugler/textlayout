package type1

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
)

var _ fonts.FontDescriptor = (*Font)(nil)

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

	return []fonts.FontDescriptor{&font}, nil
}

func (f *Font) Family() string {
	return f.PSInfo.FamilyName
}

func (f *Font) AdditionalStyle() string {
	// ported from freetype/src/type1/t1objs.c
	var styleName string

	familyName := f.PSInfo.FamilyName
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
	return styleName
}

func (f *Font) Aspect() (fonts.Style, fonts.Weight, fonts.Stretch) {
	style := fonts.StyleNormal
	if isItalic := f.PSInfo.ItalicAngle != 0; isItalic {
		style = fonts.StyleItalic
	}

	var weight fonts.Weight
	switch f.PSInfo.Weight {
	case "Bold":
		weight = fonts.WeightBold
	case "Black":
		weight = fonts.WeightBlack
	}

	return style, weight, 0
}

func (f *Font) LoadCmap() (fonts.Cmap, error) { return f.cmap, nil }
