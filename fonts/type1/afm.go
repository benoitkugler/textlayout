// Package type1 implements a parser for Adobe Type1 fonts,
// defined by .afm files (https://www.adobe.com/content/dam/acom/en/devnet/font/pdfs/5004.AFM_Spec.pdf)
// and .pdf files (https://www.adobe.com/content/dam/acom/en/devnet/font/pdfs/T1_SPEC.pdf)
package type1

import (
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
)

type Fl = float32

const Notdef = ".notdef"

type CharMetric struct {
	code     *byte // nil for not encoded glyphs
	name     string
	CharBBox [4]int

	Width int
}

// KernPair represents a kerning pair, from
// an implicit first first glyph.
type KernPair struct {
	SndChar string // glyph name
	// KerningDistance is expressed in glyph units.
	// It is most often negative,
	// that is, a negative value indicates that chars
	// should be closer.
	KerningDistance int
}

// AFMFont represents a type1 font as found in a .afm
// file.
type AFMFont struct {
	// Represents the section CharMetrics in the AFM file.
	// The key is the name of the char.
	// Even not encoded chars are present
	CharMetrics        map[string]CharMetric
	CharCodeToCharName [256]string // encoded chars

	KernPairs map[string][]KernPair
	// the character set of the font.
	CharacterSet string

	encodingScheme string

	fonts.PSInfo

	Ascender  Fl
	CapHeight Fl
	Descender Fl
	Llx       Fl // the llx of the FontBox
	Lly       Fl // the lly of the FontBox
	Urx       Fl // the urx of the FontBox
	Ury       Fl // the ury of the FontBox

	XHeight int

	StdHw int
	StdVw int
}

// CharSet returns a string listing the character names defined in the font subset.
// The names in this string shall be in PDF syntax—that is, each name preceded by a slash (/).
// The names may appear in any order. The name .notdef shall be
// omitted; it shall exist in the font subset.
func (f AFMFont) CharSet() string {
	var v strings.Builder
	for name := range f.CharMetrics {
		if name != Notdef {
			v.WriteString("/" + name)
		}
	}
	return v.String()
}
