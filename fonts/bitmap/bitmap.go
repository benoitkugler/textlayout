// Pacakge bitmap provides support for bitmap fonts
// found in .pcf files.
package bitmap

import (
	"errors"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"golang.org/x/image/math/fixed"
)

var Loader fonts.FontLoader = loader{}

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

// GetBDFProperty return a property from a bitmap font,
// or nil if it is not found.
func (f *Font) GetBDFProperty(s string) Property { return f.properties[s] }

func (f *Font) Style() (isItalic, isBold bool, familyName, styleName string) {
	// ported from freetype/src/pcf/pcfread.c
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
	styleName = strings.Join(strs, " ")
	if styleName == "" { // assume `Regular' style because we don't know better
		styleName = "Regular"
	}

	if prop, ok := f.GetBDFProperty("FAMILY_NAME").(Atom); ok {
		// Prepend the foundry name plus a space to the family name.
		// There are many fonts just called `Fixed' which look
		// completely different, and which have nothing to do with each
		// other.  When selecting `Fixed' in KDE or Gnome one gets
		// results that appear rather random, the styleName changes often if
		// one changes the size and one cannot select some fonts at all.
		//
		// We also check whether we have `wide' characters; all put
		// together, we get family names like `Sony Fixed' or `Misc
		// Fixed Wide'.

		// int  l    = ft_strlen( prop.value.atom ) + 1;

		foundryProp, _ := f.GetBDFProperty("FOUNDRY").(Atom)

		familyName = string(prop)
		if foundryProp != "" {
			familyName = string(foundryProp + " " + prop)
		}

		pointSizeProp, hasPointSize := f.GetBDFProperty("POINT_SIZE").(Int)
		averageWidthProp, hasAverageWidth := f.GetBDFProperty("AVERAGE_WIDTH").(Int)
		if hasPointSize && hasAverageWidth {
			if averageWidthProp >= pointSizeProp {
				// This font is at least square shaped or even wider
				familyName += " Wide"
			}
		}
	}

	return
}

func (Font) GlyphKind() (scalable, bitmap, color bool) {
	return false, true, false
}

func (f *Font) PoscriptName() string { return "" }

func (f *Font) PostscriptInfo() (fonts.PSInfo, bool) { return fonts.PSInfo{}, false }

// type bitmap struct {
// 	rows, width uint
// 	pitch       int
// }

func (f *Font) GetAdvance(index fonts.GlyphIndex) (int32, error) {
	if int(index) >= len(f.metrics) {
		return 0, errors.New("invalid glyph index")
	}
	return int32(f.metrics[index].characterWidth) * 64, nil
}
