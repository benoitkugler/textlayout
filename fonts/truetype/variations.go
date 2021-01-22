package truetype

import (
	"errors"
	"strconv"
)

// VariableFont is implemented by formats with variable font
// support.
// TODO: polish
type VariableFont interface {
	Variations() TableFvar
}

type VarAxis struct {
	Tag     Tag     // Tag identifying the design variation for the axis.
	Minimum float64 // mininum value on the variation axis that the font covers
	Default float64 // default position on the axis
	Maximum float64 // maximum value on the variation axis that the font covers
	flags   uint16  // 	Axis qualifiers — see details below.
	strid   NameID  // name entry in the font's ‘name’ table
}

type Variation struct {
	Tag   Tag // variation-axis identifier tag
	Value float64
}

// NewVariation parse the string representation of a variation
// of the form tag=value
func NewVariation(s string) (Variation, error) {
	pr := parser{data: []byte(s)}
	return pr.parseOneVariation()
}

type parser struct {
	data []byte
	pos  int
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\f' || c == '\n' || c == '\r' || c == '\t' || c == '\v'
}

func (p *parser) skipSpaces() {
	for p.pos < len(p.data) && isSpace(p.data[p.pos]) {
		p.pos++
	}
}

// return true if `c` was found
func (p *parser) parseChar(c byte) bool {
	p.skipSpaces()

	if p.pos == len(p.data) || p.data[p.pos] != c {
		return false
	}
	p.pos++
	return true
}

// func (p *parser) parse_uint() (uint, bool) {
// 	start := p.pos
// 	// go to the next space
// 	for p.pos < len(p.data) && !isSpace(p.data[p.pos]) {
// 		p.pos++
// 	}
// 	out, err := strconv.Atoi(string(p.data[start:p.pos]))
// 	// Intentionally use Atoi inside instead, such that -1 turns into "big number"...
// 	return uint(out), err == nil
// }

// static bool
// parse_uint32 (const char **pp, const char *end, uint32_t *pv)
// {
//   /* Intentionally use hb_parse_int inside instead of hb_parse_uint,
//    * such that -1 turns into "big number"... */
//   int v;
//   if (unlikely (!hb_parse_int (pp, end, &v))) return false;

//   *pv = v;
//   return true;
// }

// static bool
// parse_bool (const char **pp, const char *end, uint32_t *pv)
// {
//   skipSpaces (pp, end);

//   const char *p = *pp;
//   while (*pp < end && ISALPHA(**pp))
//     (*pp)++;

//   /* CSS allows on/off as aliases 1/0. */
//   if (*pp - p == 2
//       && TOLOWER (p[0]) == 'o'
//       && TOLOWER (p[1]) == 'n')
//     *pv = 1;
//   else if (*pp - p == 3
// 	   && TOLOWER (p[0]) == 'o'
// 	   && TOLOWER (p[1]) == 'f'
// 	   && TOLOWER (p[2]) == 'f')
//     *pv = 0;
//   else
//     return false;

//   return true;
// }

func isAlnum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func (p *parser) parseTag() (Tag, error) {
	p.skipSpaces()

	var quote byte

	if p.pos < len(p.data) && (p.data[p.pos] == '\'' || p.data[p.pos] == '"') {
		quote = p.data[p.pos]
		p.pos++
	}

	start := p.pos
	for p.pos < len(p.data) && (isAlnum(p.data[p.pos]) || p.data[p.pos] == '_') {
		p.pos++
	}

	if p.pos == start || p.pos > start+4 {
		return 0, errors.New("invalid tag length")
	}

	// padd with space if necessary, since MustNewTag requires 4 bytes
	tagBytes := [4]byte{' ', ' ', ' ', ' '}
	copy(tagBytes[:], p.data[start:p.pos])
	tag := MustNewTag(string(tagBytes[:]))

	if quote != 0 {
		/* CSS expects exactly four bytes.  And we only allow quotations for
		 * CSS compatibility.  So, enforce the length. */
		if p.pos != start+4 {
			return 0, errors.New("tag must have 4 bytes")
		}
		if p.pos == len(p.data) || p.data[p.pos] != quote {
			return 0, errors.New("tag is missing end quote")
		}
		p.pos++
	}

	return tag, nil
}

func (p *parser) parseVariationValue() (float64, error) {
	p.parseChar('=') // Optional.
	start := p.pos
	// go to the next space
	for p.pos < len(p.data) && !isSpace(p.data[p.pos]) {
		p.pos++
	}
	v, err := strconv.ParseFloat(string(p.data[start:p.pos]), 64)
	return v, err
}

func (p *parser) parseOneVariation() (vari Variation, err error) {
	vari.Tag, err = p.parseTag()
	if err != nil {
		return
	}
	vari.Value, err = p.parseVariationValue()
	if err != nil {
		return
	}
	p.skipSpaces()
	return
}

type VarInstance struct {
	Coords    []float64 // length: number of axis
	Subfamily NameID

	PSStringID NameID
}

type TableFvar struct {
	Axis      []VarAxis
	Instances []VarInstance // contains the default instance
}

// IsDefaultInstance returns `true` is `instance` has the same
// coordinates as the default instance.
func (t TableFvar) IsDefaultInstance(it VarInstance) bool {
	for i, c := range it.Coords {
		if c != t.Axis[i].Default {
			return false
		}
	}
	return true
}

// add the default instance if it not already explicitely present
func (t *TableFvar) checkDefaultInstance(names TableName) {
	for _, instance := range t.Instances {
		if t.IsDefaultInstance(instance) {
			return
		}
	}

	// add the default instance
	// choose the subfamily entry
	subFamily := NamePreferredSubfamily
	if v1, v2 := names.getEntry(subFamily); v1 == nil && v2 == nil {
		subFamily = NameFontSubfamily
	}
	defaultInstance := VarInstance{
		Coords:     make([]float64, len(t.Axis)),
		Subfamily:  subFamily,
		PSStringID: NamePostscript,
	}
	for i, axe := range t.Axis {
		defaultInstance.Coords[i] = axe.Default
	}
	t.Instances = append(t.Instances, defaultInstance)
}

// convert from design coordinates to normalized coordinates,
// applying an optional 'avar' table
func (mmvar *TableFvar) ft_var_to_normalized(coords, normalized []float64, avar *tableAvar) {
	// ported from freetype2

	// Axis normalization is a two-stage process.  First we normalize
	// based on the [min,def,max] values for the axis to be [-1,0,1].
	// Then, if there's an `avar' table, we renormalize this range.

	for i, a := range mmvar.Axis {
		coord := coords[i]

		if coord > a.Maximum || coord < a.Minimum { // out of range: clamping
			if coord > a.Maximum {
				coord = a.Maximum
			} else {
				coord = a.Minimum
			}
		}

		if coord < a.Default {
			normalized[i] = -(coord - a.Default) / (a.Minimum - a.Default)
		} else if coord > a.Default {
			normalized[i] = (coord - a.Default) / (a.Maximum - a.Default)
		} else {
			normalized[i] = 0
		}
	}

	if avar != nil { // now applying 'avar'
		for i, av := range avar.axisSegmentMaps {
			for j := 1; j < len(av); j++ {
				previous, pair := av[j-1], av[j]
				if normalized[i] < pair.from {
					normalized[i] =
						previous.to + (normalized[i]-previous.from)*
							(pair.to-previous.to)/(pair.from-previous.from)
					break
				}
			}
		}
	}
}
