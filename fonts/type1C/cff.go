// Parser for the CFF font format
// defined at https://www.adobe.com/content/dam/acom/en/devnet/font/pdfs/5176.CFF.pdf
package type1C

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/simpleencodings"
)

var _ fonts.Font = (*CFF)(nil)

// CFF represents a parsed CFF font.
type CFF struct {
	fonts.PSInfo
	Encoding    *simpleencodings.Encoding
	fontNames   [][]byte // name index, one name per font
	cidFontName string
	fdSelect    fdSelect
	charstrings [][]byte // indexed by glyph ID
}

// Parse parse a .cff font file.
// Although the format natively support multiple fonts,
// this package only support single font files.
func Parse(file fonts.Ressource) (*CFF, error) {
	if err := checkHeader(file); err != nil {
		return nil, err
	}
	file.Seek(0, io.SeekStart)
	// if this is really needed, we can modify the parser to directly use `file`
	// without reading all in memory
	input, err := ioutil.ReadAll(file)
	if err != nil {
		return &CFF{}, err
	}
	p := cffParser{src: input}
	p.skip(4)
	return p.parse()
}

func (f *CFF) PostscriptInfo() (fonts.PSInfo, bool) { return f.PSInfo, true }

func (f *CFF) PoscriptName() string { return f.PSInfo.FontName }

// Strip all subset prefixes of the form `ABCDEF+'.  Usually, there
// is only one, but font names like `APCOOG+JFABTD+FuturaBQ-Bold'
// have been seen in the wild.
func removeSubsetPrefix(name []byte) []byte {
	for keep := true; keep; {
		if len(name) >= 7 && name[6] == '+' {
			for idx := 0; idx < 6; idx++ {
				/* ASCII uppercase letters */
				if !('A' <= name[idx] && name[idx] <= 'Z') {
					keep = false
				}
			}
			if keep {
				name = name[7:]
			}
		} else {
			keep = false
		}
	}
	return name
}

// remove the style part from the family name (if present).
func removeStyle(familyName, styleName string) string {
	if lF, lS := len(familyName), len(styleName); lF > lS {
		idx := 1
		for ; idx <= len(styleName); idx++ {
			if familyName[lF-idx] != styleName[lS-idx] {
				break
			}
		}

		if idx > lS {
			// familyName ends with styleName; remove it
			idx = lF - lS - 1

			// also remove special characters
			// between real family name and style
			for idx > 0 &&
				(familyName[idx] == '-' || familyName[idx] == ' ' ||
					familyName[idx] == '_' || familyName[idx] == '+') {
				idx--
			}

			if idx > 0 {
				familyName = familyName[:idx+1]
			}
		}
	}
	return familyName
}

func (f *CFF) Style() (isItalic, isBold bool, styleName string) {
	// adapted from freetype

	// retrieve font family & style name
	familyName := f.PSInfo.FamilyName
	if familyName == "" {
		// TODO: adapt for multiple fonts
		familyName = string(removeSubsetPrefix(f.fontNames[0]))
	}
	if familyName != "" {
		full := f.PSInfo.FullName
		// char * fullp = full
		// char * family = cffface.familyName

		// We try to extract the style name from the full name.
		// We need to ignore spaces and dashes during the search.
		for i, j := 0, 0; i < len(full); {
			// skip common characters at the start of both strings
			if full[i] == familyName[j] {
				i++
				j++
				continue
			}

			// ignore spaces and dashes in full name during comparison
			if full[i] == ' ' || full[i] == '-' {
				i++
				continue
			}

			// ignore spaces and dashes in family name during comparison
			if familyName[j] == ' ' || familyName[j] == '-' {
				j++
				continue
			}

			if j == len(familyName) && i < len(full) {
				/* The full name begins with the same characters as the  */
				/* family name, with spaces and dashes removed.  In this */
				/* case, the remaining string in `full' will be used as */
				/* the style name.                                       */
				styleName = full[i:]

				/* remove the style part from the family name (if present) */
				familyName = removeStyle(familyName, styleName)
			}
			break
		}
	} else {
		// do we have a `/FontName' for a CID-keyed font?
		if f.cidFontName != "" {
			familyName = f.cidFontName
		}
	}

	if styleName == "" {
		// assume "Regular" style if we don't know better
		styleName = "Regular"
	}

	isItalic = f.PSInfo.ItalicAngle != 0
	isBold = f.PSInfo.Weight == "Bold" || f.PSInfo.Weight == "Black"

	// double check
	if !isBold {
		isBold = strings.HasPrefix(styleName, "Bold") || strings.HasPrefix(styleName, "Black")
	}
	return
}
