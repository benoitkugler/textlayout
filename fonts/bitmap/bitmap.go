// Pacakge bitmap provides support for bitmap fonts
// found in .pcf files.
package bitmap

import (
	"errors"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
)

var Loader fonts.FontLoader = loader{}

var _ fonts.Face = (*Font)(nil)

// Property is either an `Atom` or an `Int`
type Property interface {
	isProperty()
}

func (Atom) isProperty() {}
func (Int) isProperty()  {}

type Atom string

type Int int32

type loader struct{}

// Load implements fonts.FontLoader. When the error is `nil`,
// one (and only one) font is returned.
func (loader) Load(file fonts.Resource) (fonts.Faces, error) {
	f, err := Parse(file)
	if err != nil {
		return nil, err
	}
	return fonts.Faces{f}, nil
}

// read the charset properties and build the cmap
// only unicode charmap is supported
func (f *Font) Cmap() (fonts.Cmap, fonts.CmapEncoding) {
	var encKind fonts.CmapEncoding

	// inspired by freetype
	reg, hasReg := f.GetBDFProperty("CHARSET_REGISTRY").(Atom)
	enc, hasEnc := f.GetBDFProperty("CHARSET_ENCODING").(Atom)

	if hasReg && hasEnc {
		/* Uh, oh, compare first letters manually to avoid dependency
		   on locales. */
		reg := strings.ToLower(string(reg))
		if strings.HasPrefix(reg, "iso") {
			if reg == "iso10646" || reg == "iso8859" && enc == "1" {
				encKind = fonts.EncUnicode
			} else if reg == "iso646.1991" && enc == "IRV" {
				/* another name for ASCII */
				encKind = fonts.EncUnicode
			}
		}
	}

	return &f.cmap, encKind
}

type encodingTable struct {
	values           []fonts.GID
	minChar, maxChar byte
	minByte, maxByte byte
	defaultChar      fonts.GID
}

type encodingIterator struct {
	origin *encodingTable
	L      int // precomputed
	pos    int // in values array
}

func (iter *encodingIterator) Next() bool {
	// go to the next glyph
	for iter.pos < len(iter.origin.values) {
		if iter.origin.values[iter.pos] != 0xFFFF {
			iter.pos++
			return true // we have a glyph
		}
		iter.pos++
	}
	return false // no more glyph
}

func (iter *encodingIterator) Char() (rune, fonts.GID) {
	// iter.pos is one ahead
	index := iter.pos - 1
	j := index % iter.L // index = i * L + j
	i := byte((index - j) / iter.L)
	r := rune(iter.origin.minByte+i)<<8 | rune(iter.origin.minChar) + rune(j)
	return r, iter.origin.values[index]
}

func (enc *encodingTable) Iter() fonts.CmapIter {
	return &encodingIterator{origin: enc, L: int(enc.maxChar-enc.minChar) + 1}
}

func (enc encodingTable) Lookup(ch rune) fonts.GID {
	if ch > 0xFFFF {
		return enc.defaultChar
	}
	enc1 := byte(ch >> 8)
	enc2 := byte(ch)
	if enc1 < enc.minByte || enc1 > enc.maxByte || enc2 < enc.minChar || enc2 > enc.maxChar {
		return enc.defaultChar
	}
	L := int(enc.maxChar-enc.minChar) + 1
	v := enc.values[int(enc1-enc.minByte)*L+int(enc2-enc.minChar)]
	if v == 0xFFFF {
		return enc.defaultChar
	}
	return v
}

// GetBDFProperty return a property from a bitmap font,
// or nil if it is not found.
func (f *Font) GetBDFProperty(s string) Property { return f.properties[s] }

func (f *Font) getStyle() (isItalic, isBold bool, familyName, styleName string) {
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

func (f *Font) LoadSummary() (fonts.FontSummary, error) {
	isItalic, isBold, familyName, styleName := f.getStyle()
	return fonts.FontSummary{
		IsItalic:          isItalic,
		IsBold:            isBold,
		Familly:           familyName,
		Style:             styleName,
		HasScalableGlyphs: false,
		HasBitmapGlyphs:   true,
		HasColorGlyphs:    false,
	}, nil
}

func (f *Font) PoscriptName() string { return "" }

func (f *Font) PostscriptInfo() (fonts.PSInfo, bool) { return fonts.PSInfo{}, false }

// type bitmap struct {
// 	rows, width uint
// 	pitch       int
// }

func (f *Font) GetAdvance(index fonts.GID) (int32, error) {
	if int(index) >= len(f.metrics) {
		return 0, errors.New("invalid glyph index")
	}
	return int32(f.metrics[index].characterWidth) * 64, nil
}

// TODO:
func (f *Font) LoadMetrics() fonts.FaceMetrics { return nil }

func mulDiv(a, b, c uint16) uint16 {
	return uint16(uint32(a) * uint32(b) / uint32(c))
}

func (f *Font) computeBitmapSize() fonts.BitmapSize {
	// adapted from freetype
	var size fonts.BitmapSize
	if h := abs(f.accelerator.fontAscent + f.accelerator.fontDescent); h <= 0xFFFF {
		size.Height = uint16(h)
	} else {
		size.Height = 0xFFFF // clamping
	}
	if w, ok := f.GetBDFProperty("AVERAGE_WIDTH").(Int); ok {
		if abs(int32(w)) > 0xFFFF*10-5 {
			size.Width = 0xFFFF // clamping
		} else {
			size.Width = uint16(abs((int32(w) + 5) / 10))
		}
	} else {
		size.Width = mulDiv(size.Height, 2, 3) // heuristic
	}

	var pointSize uint16
	if ps, ok := f.GetBDFProperty("POINT_SIZE").(Int); ok {
		/* convert from 722.7 decipoints to 72 points per inch */
		if v := abs(int32(ps)); v <= 0xFFFF*72270/7200 {
			pointSize = uint16(v * 7200 / 72270)
		} else {
			pointSize = 0xFFFF
		}
	}

	if ppem, ok := f.GetBDFProperty("PIXEL_SIZE").(Int); ok {
		if v := abs(int32(ppem)); v <= 0xFFFF {
			size.YPpem = uint16(v)
		} else {
			size.YPpem = 0xFFFF
		}
	}

	var resolutionX, resolutionY uint16
	if res, ok := f.GetBDFProperty("RESOLUTION_X").(Int); ok {
		if v := abs(int32(res)); v <= 0xFFFF {
			resolutionX = uint16(v)
		} else {
			resolutionX = 0xFFFF
		}
	}
	if res, ok := f.GetBDFProperty("RESOLUTION_Y").(Int); ok {
		if v := abs(int32(res)); v <= 0xFFFF {
			resolutionY = uint16(v)
		} else {
			resolutionY = 0xFFFF
		}
	}

	if size.YPpem == 0 {
		size.YPpem = pointSize
		if resolutionY != 0 {
			size.YPpem = mulDiv(size.YPpem, resolutionY, 72)
		}
	}

	if resolutionX != 0 && resolutionY != 0 {
		size.XPpem = mulDiv(size.YPpem, resolutionX, resolutionY)
	} else {
		size.XPpem = size.YPpem
	}

	return size
}

func abs(i int32) int32 {
	if i < 0 {
		return -i
	}
	return i
}

// LoadBitmaps always returns a one element slice.
func (f *Font) LoadBitmaps() []fonts.BitmapSize { return []fonts.BitmapSize{f.computeBitmapSize()} }
