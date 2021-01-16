// Pacakge bitmap provides support for bitmap fonts
// found in .pcf files.
package bitmap

import (
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"golang.org/x/image/math/fixed"
)

var _ fonts.Font = (*Font)(nil)

// Property is either an `Atom` or an `Int`
type Property interface {
	isProperty()
}

func (Atom) isProperty() {}
func (Int) isProperty()  {}

type Atom string

type Int int32

type Size struct {
	Height, Width int16

	// size fixed.Point26_6

	XPpem, YPpem fixed.Int26_6
}

// GetBDFProperty return a property from a bitmap font,
// or nil if it is not found.
func (f *Font) GetBDFProperty(s string) Property { return f.properties[s] }

func (f *Font) Style() (isItalic, isBold bool, style string) {
	// ported from freetype
	// need to convert spaces to dashes for add_style_name and setwidth_name

	var strs []string

	if prop, _ := f.GetBDFProperty("ADD_STYLE_NAME").(Atom); prop != "" &&
		!(prop[0] == 'N' || prop[0] == 'n') {
		strs = append(strs, strings.ReplaceAll(string(prop), " ", "-"))
	}

	if prop, _ := f.GetBDFProperty("WEIGHT_NAME").(Atom); prop != "" &&
		(prop[0] == 'B' || prop[0] == 'b') {
		isBold = true
		strs = append(strs, "Bold")
	}

	if prop, _ := f.GetBDFProperty("SLANT").(Atom); prop != "" &&
		(prop[0] == 'O' || prop[0] == 'o' || prop[0] == 'I' || prop[0] == 'i') {
		isItalic = true
		if prop[0] == 'O' || prop[0] == 'o' {
			strs = append(strs, "Oblique")
		} else {
			strs = append(strs, "Italic")
		}
	}

	if prop, _ := f.GetBDFProperty("SETWIDTH_NAME").(Atom); prop != "" &&
		!(prop[0] == 'N' || prop[0] == 'n') {
		strs = append(strs, strings.ReplaceAll(string(prop), " ", "-"))
	}

	// separate elements with a space
	style = strings.Join(strs, " ")

	return isItalic, isBold, style
}

func (f *Font) PoscriptName() string { return "" }

func (f *Font) PostscriptInfo() (fonts.PSInfo, bool) { return fonts.PSInfo{}, false }
