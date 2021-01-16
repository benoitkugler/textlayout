package pango

import (
	"strconv"
	"strings"
)

// pango_color_parse_with_alpha returns a color from a string specification. The
// string can either one of a large set of standard names (http://dev.w3.org/csswg/css-color/#named-colors),
// or it can be a hexadecimal value in the
// form '#rgb' '#rrggbb' '#rrrgggbbb' or '#rrrrggggbbbb' where
// 'r', 'g' and 'b' are hex digits of the red, green, and blue
// components of the color, respectively.
//
// Additionally, if `withAlpha` is `true`, parse strings of the form
// '#rgba', '#rrggbbaa', '#rrrrggggbbbbaaaa', and returns the value specified
// by the hex digits for 'a'. If no alpha component is found
// in `spec`, `alpha` is set to 0xffff (for a solid color).
func pango_color_parse_with_alpha(spec string, withAlpha bool) (AttrColor, uint16, bool) {
	if len(spec) == 0 {
		return AttrColor{}, 0, false
	}
	alpha := uint16(0xffff)

	var color AttrColor

	if spec[0] == '#' {
		has_alpha := false

		len := len(spec) - 1
		switch len {
		case 3, 6, 9, 12:
			len /= 3
			has_alpha = false
		case 4, 8, 16:
			if !withAlpha {
				return AttrColor{}, 0, false
			}
			len /= 4
			has_alpha = true
		default:
			return AttrColor{}, 0, false
		}
		spec = spec[1:]

		r, err := strconv.ParseInt(spec[0:len], 16, 32)
		if err != nil {
			return AttrColor{}, 0, false
		}
		g, err := strconv.ParseInt(spec[len:2*len], 16, 32)
		if err != nil {
			return AttrColor{}, 0, false
		}
		b, err := strconv.ParseInt(spec[2*len:3*len], 16, 32)
		if err != nil {
			return AttrColor{}, 0, false
		}

		if has_alpha {
			a, err := strconv.ParseInt(spec[3*len:4*len], 16, 32)
			if err != nil {
				return AttrColor{}, 0, false
			}
			bits := len * 4
			a <<= 16 - bits
			for bits < 16 {
				a |= (a >> bits)
				bits *= 2
			}
			alpha = uint16(a)
		}

		bits := len * 4
		r <<= 16 - bits
		g <<= 16 - bits
		b <<= 16 - bits
		for bits < 16 {
			r |= (r >> bits)
			g |= (g >> bits)
			b |= (b >> bits)
			bits *= 2
		}
		color.Red = uint16(r)
		color.Green = uint16(g)
		color.Blue = uint16(b)
	} else {
		var ok bool
		color, ok = colorEntries[strings.ToLower(spec)]
		if !ok {
			return AttrColor{}, 0, false
		}
	}
	return color, alpha, true
}

// pango_color_parse returns a color from a string specification,
// rejecting color with alpha component.
// See `pango_color_parse_with_alpha` for details.
func pango_color_parse(spec string) (AttrColor, bool) {
	c, _, b := pango_color_parse_with_alpha(spec, false)
	return c, b
}
