package type1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
)

const spaces = " \t\n\r\f"

var defautFontValues = AFMFont{
	PSInfo: fonts.PSInfo{
		UnderlinePosition:  -100,
		UnderlineThickness: 50,
	},
	encodingScheme: "FontSpecific",
	XHeight:        480,
	StdVw:          80,
}

// ParseAFMFile read a .afm file and return the associated font.
func ParseAFMFile(source io.Reader) (AFMFont, error) {
	f := defautFontValues
	// deep copy to avoid state sharing
	f.CharMetrics = map[string]CharMetric{}
	f.KernPairs = map[string][]KernPair{}

	err := f.parse(source)

	f.encodingScheme = strings.TrimSpace(f.encodingScheme)

	return f, err
}

// safely try to read one token; returns an error
// if it's not found
func readToken(tokens []string, index int) (string, error) {
	if index >= len(tokens) {
		return "", fmt.Errorf("invalid line %s : expected %d tokens", strings.Join(tokens, " "), index+1)
	}
	return tokens[index], nil
}

func readIntToken(tokens []string, index int) (int, error) {
	s, err := readToken(tokens, index)
	if err != nil {
		return 0, err
	}
	out, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid int in line %s (%s)", strings.Join(tokens, " "), err)
	}
	return out, nil
}

func readFloatToken(tokens []string, index int) (Fl, error) {
	s, err := readToken(tokens, index)
	if err != nil {
		return 0, err
	}
	out, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float in line %s (%s)", strings.Join(tokens, " "), err)
	}
	return Fl(out), nil
}

func (f *AFMFont) parse(source io.Reader) error {
	scanner := bufio.NewScanner(source)
	isMetrics := false
	for scanner.Scan() {
		line := scanner.Text()

		tok := strings.Fields(line)
		if len(tok) == 0 {
			continue
		}
		ident := tok[0]
		var err error
		switch ident {
		case "Version":
			f.Version, err = readToken(tok, 1)
		case "Notice":
			f.Notice, err = readToken(tok, 1)
		case "FontName":
			f.FontName, err = readToken(tok, 1)
		case "FullName":
			f.FullName, err = readToken(tok, 1)
		case "FamilyName":
			f.FamilyName, err = readToken(tok, 1)
		case "Weight":
			f.Weight, err = readToken(tok, 1)
		case "ItalicAngle":
			var ia Fl
			ia, err = readFloatToken(tok, 1)
			f.ItalicAngle = int(ia)
		case "IsFixedPitch":
			var s string
			s, err = readToken(tok, 1)
			f.IsFixedPitch = s == "true"
		case "CharacterSet":
			f.CharacterSet, err = readToken(tok, 1)
		case "FontBBox":
			f.Llx, err = readFloatToken(tok, 1)
			if err != nil {
				break
			}
			f.Lly, err = readFloatToken(tok, 2)
			if err != nil {
				break
			}
			f.Urx, err = readFloatToken(tok, 3)
			if err != nil {
				break
			}
			f.Ury, err = readFloatToken(tok, 4)
		case "UnderlinePosition":
			var p int
			p, err = readIntToken(tok, 1)
			f.UnderlinePosition = int(p)
		case "UnderlineThickness":
			var p int
			p, err = readIntToken(tok, 1)
			f.UnderlineThickness = int(p)
		case "EncodingScheme":
			f.encodingScheme, err = readToken(tok, 1)
		case "CapHeight":
			f.CapHeight, err = readFloatToken(tok, 1)
		case "XHeight":
			f.XHeight, err = readIntToken(tok, 1)
		case "Ascender":
			f.Ascender, err = readFloatToken(tok, 1)
		case "Descender":
			f.Descender, err = readFloatToken(tok, 1)
		case "StdHW":
			f.StdHw, err = readIntToken(tok, 1)
		case "StdVW":
			f.StdVw, err = readIntToken(tok, 1)
		}
		if err != nil {
			return err
		}
		if ident == "StartCharMetrics" {
			isMetrics = true
			break
		}
	}

	if !isMetrics {
		return errors.New("missing StartCharMetrics in font file")
	}

	for scanner.Scan() {
		line := scanner.Text()
		tok := strings.Fields(line)
		if len(tok) == 0 {
			continue
		}
		ident := tok[0]
		if ident == "EndCharMetrics" {
			isMetrics = false
			break
		}

		met := CharMetric{Width: 250}
		tok = strings.Split(line, ";")
		for len(tok) > 0 {
			tokc := strings.Fields(tok[0])
			tok = tok[1:] // go to next token
			if len(tokc) == 0 {
				continue
			}
			ident := tokc[0]
			var (
				err error
				c   int
			)
			switch ident {
			case "C":
				c, err = readIntToken(tokc, 1)
				if c >= 0 && c <= 255 { // encoded
					by := byte(c)
					met.code = &by
				}
			case "WX":
				met.Width, err = readIntToken(tokc, 1)
			case "N":
				met.name, err = readToken(tokc, 1)
			case "B":
				met.CharBBox[0], err = readIntToken(tokc, 1)
				if err != nil {
					break
				}
				met.CharBBox[1], err = readIntToken(tokc, 2)
				if err != nil {
					break
				}
				met.CharBBox[2], err = readIntToken(tokc, 3)
				if err != nil {
					break
				}
				met.CharBBox[3], err = readIntToken(tokc, 4)
			}
			if err != nil {
				return err
			}
		}
		f.CharMetrics[met.name] = met
		if met.code != nil {
			f.CharCodeToCharName[*met.code] = met.name
		}
	}

	if isMetrics {
		return errors.New("missing EndCharMetrics in font file")
	}
	for scanner.Scan() {
		line := scanner.Text()
		tok := strings.Fields(line)
		if len(tok) == 0 {
			continue
		}
		ident := tok[0]
		if ident == "EndFontMetrics" {
			goto end
		}
		if ident == "StartKernPairs" {
			isMetrics = true
			break
		}
	}
	if !isMetrics {
		return errors.New("missing EndFontMetrics in font file")
	}

	for scanner.Scan() {
		line := scanner.Text()
		tok := strings.Fields(line)
		if len(tok) == 0 {
			continue
		}
		ident := tok[0]
		if ident == "KPX" {
			first, err := readToken(tok, 1)
			if err != nil {
				return err
			}
			second, err := readToken(tok, 2)
			if err != nil {
				return err
			}
			width, err := readIntToken(tok, 3)
			if err != nil {
				return err
			}
			f.KernPairs[first] = append(f.KernPairs[first], KernPair{SndChar: second, KerningDistance: width})
		} else if ident == "EndKernPairs" {
			isMetrics = false
			break
		}
	}
	if isMetrics {
		return errors.New("missing EndKernPairs in font file")
	}
end:
	err := scanner.Err()
	return err
}
