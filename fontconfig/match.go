package fontconfig

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ported from fontconfig/src/fcmatch.c Copyright © 2000 Keith Packard

type matchKind int8

const (
	matchDefault matchKind = iota - 1
	// Rules in the config to apply to the query pattern (FcMatchPattern)
	MatchQuery
	// Rules in the config to apply to the fonts
	// returned as the result of a query (FcMatchFont)
	MatchResult
	// Rules in the config to apply to the fonts obtained
	// during the scan (FcMatchScan)
	MatchScan
	matchKindEnd
)

func (k matchKind) String() string {
	switch k {
	case MatchQuery, matchDefault:
		return "query"
	case MatchResult:
		return "result"
	case MatchScan:
		return "scan"
	default:
		return fmt.Sprintf("<match kind %d>", k)
	}
}

func compareNumber(value1, value2 Value) (Value, float32) {
	var v1, v2 float32
	switch value := value1.(type) {
	case Int:
		v1 = float32(value)
	case Float:
		v1 = float32(value)
	default:
		return nil, -1.0
	}
	switch value := value2.(type) {
	case Int:
		v2 = float32(value)
	case Float:
		v2 = float32(value)
	default:
		return nil, -1.0
	}

	v := v2 - v1
	if v < 0 {
		v = -v
	}
	return value2, v
}

func compareString(v1, v2 Value) (Value, float32) {
	bestValue := v2
	if strings.EqualFold(string(v1.(String)), string(v2.(String))) {
		return bestValue, 0
	}
	return bestValue, 1
}

// returns 0 for empty strings
func toLower(s string) byte {
	if s == "" {
		return 0
	}
	if 0101 <= s[0] && s[0] <= 0132 {
		return s[0] - 0101 + 0141
	}
	return s[0]
}

func compareFamily(v1, v2 Value) (Value, float32) {
	// rely on the guarantee in PatternObjectAddWithBinding that
	// families are always FcTypeString.
	v1String := string(v1.(String))
	v2String := string(v2.(String))

	bestValue := v2

	if toLower(v1String) != toLower(v2String) &&
		v1String[0] != ' ' && v2String[0] != ' ' {
		return bestValue, 1.0
	}

	if ignoreBlanksAndCase(v1String) == ignoreBlanksAndCase(v2String) {
		return bestValue, 0
	}
	return bestValue, 1
}

var delimReplacer = strings.NewReplacer(" ", "", "-", "")

func matchIgnoreCaseAndDelims(s1, s2 string) int {
	s1, s2 = delimReplacer.Replace(s1), delimReplacer.Replace(s2)
	s1, s2 = strings.ToLower(s1), strings.ToLower(s2)
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}
	i := 0
	for ; i < l; i++ {
		if s1[i] != s2[i] {
			break
		}
	}
	return i
}

func comparePostScript(v1, v2 Value) (Value, float32) {
	v1_string := string(v1.(String))
	v2_string := string(v2.(String))

	bestValue := v2

	if toLower(v1_string) != toLower(v2_string) &&
		v1_string[0] != ' ' && v2_string[0] != ' ' {
		return bestValue, 1.0
	}

	n := matchIgnoreCaseAndDelims(v1_string, v2_string)
	length := len(v1_string)

	return bestValue, float32(length-n) / float32(length)
}

func compareLang(val1, val2 Value) (Value, float32) {
	var result langResult
	switch v1 := val1.(type) {
	case Langset:
		switch v2 := val2.(type) {
		case Langset:
			result = langSetCompare(v1, v2)
		case String:
			result = v1.hasLang(string(v2))
		default:
			return nil, -1.0
		}
	case String:
		switch v2 := val2.(type) {
		case Langset:
			result = v2.hasLang(string(v1))
		case String:
			result = langCompare(string(v1), string(v2))
		default:
			return nil, -1.0
		}
		break
	default:
		return nil, -1.0
	}
	bestValue := val2
	switch result {
	case langEqual:
		return bestValue, 0
	case langDifferentCountry:
		return bestValue, 1
	default:
		return bestValue, 2
	}
}

func compareBool(val1, val2 Value) (Value, float32) {
	v1, ok1 := val1.(Bool)
	v2, ok2 := val2.(Bool)
	if !ok1 || !ok2 {
		return nil, -1.0
	}

	var bestValue Bool
	if v2 != DontCare {
		bestValue = v2
	} else {
		bestValue = v1
	}

	if v1 == v2 {
		return bestValue, 0
	}
	return bestValue, 1
}

func compareCharSet(v1, v2 Value) (Value, float32) {
	bestValue := v2
	return bestValue, float32(charsetSubtractCount(v1.(Charset), v2.(Charset)))
}

func maxF(v1, v2 float32) float32 {
	if v1 > v2 {
		return v1
	}
	return v2
}

func minF(v1, v2 float32) float32 {
	if v1 < v2 {
		return v1
	}
	return v2
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func compareRange(v1, v2 Value) (Value, float32) {
	var b1, e1, b2, e2, d float32

	switch value1 := v1.(type) {
	case Int:
		e1 = float32(value1)
		b1 = e1
	case Float:
		e1 = float32(value1)
		b1 = e1
	case Range:
		b1 = value1.Begin
		e1 = value1.End
	default:
		return nil, -1
	}
	switch value2 := v2.(type) {
	case Int:
		e2 = float32(value2)
		b2 = e2
	case Float:
		e2 = float32(value2)
		b2 = e2
	case Range:
		b2 = value2.Begin
		e2 = value2.End
	default:
		return nil, -1
	}

	if e1 < b2 {
		d = b2
	} else if e2 < b1 {
		d = e2
	} else {
		d = (maxF(b1, b2) + minF(e1, e2)) * .5
	}

	bestValue := Float(d)

	/// if the ranges overlap, it's a match, otherwise return closest distance.
	if e1 < b2 || e2 < b1 {
		return bestValue, minF(absF(b2-e1), absF(b1-e2))
	}
	return bestValue, 0.0
}

func compareSize(v1, v2 Value) (Value, float32) {
	var b1, e1, b2, e2 float32

	switch value1 := v1.(type) {
	case Int:
		e1 = float32(value1)
		b1 = e1
	case Float:
		e1 = float32(value1)
		b1 = e1
	case Range:
		b1 = value1.Begin
		e1 = value1.End
	default:
		return nil, -1
	}
	switch value2 := v2.(type) {
	case Int:
		e2 = float32(value2)
		b2 = e2
	case Float:
		e2 = float32(value2)
		b2 = e2
	case Range:
		b2 = value2.Begin
		e2 = value2.End
	default:
		return nil, -1
	}

	bestValue := Float((b1 + e1) * .5)

	// if the ranges overlap, it's a match, otherwise return closest distance.
	if e1 < b2 || e2 < b1 {
		return bestValue, minF(absF(b2-e1), absF(b1-e2))
	}
	if b2 != e2 && b1 == e2 { /* Semi-closed interval. */
		return bestValue, 1e-15
	}
	return bestValue, 0.0
}

func strGlobMatch(glob, st string) bool {
	var str int // index in st
	for i, c := range []byte(glob) {
		switch c {
		case '*':
			// short circuit common case
			if i == len(glob)-1 {
				return true
			}
			// short circuit another common case
			if i < len(glob)-1 && glob[i+1] == '*' {

				l1 := len(st) - str
				l2 := len(glob)
				if l1 < l2 {
					return false
				}
				str += (l1 - l2)
			}
			for str < len(st) {
				if strGlobMatch(glob, st[str:]) {
					return true
				}
				str++
			}
			return false
		case '?':
			if str == len(st) {
				return false
			}
			str++
		default:
			if st[str] != c {
				return false
			}
			str++
		}
	}
	return str == len(st)
}

func compareFilename(v1, v2 Value) (Value, float32) {
	s1, s2 := string(v1.(String)), string(v2.(String))
	bestValue := String(s2)
	if s1 == s2 {
		return bestValue, 0.0
	}
	if cmpIgnoreCase(s1, s2) == 0 {
		return bestValue, 1.0
	}
	if strGlobMatch(s1, s2) {
		return bestValue, 2.0
	}
	return bestValue, 3.0
}

// Canonical match priority order
type matcherPriority int8

const (
	priFILE matcherPriority = iota
	priFONTFORMAT
	priVARIABLE
	priSCALABLE
	priCOLOR
	priFOUNDRY
	priCHARSET
	priFAMILY_STRONG
	priPOSTSCRIPT_NAME_STRONG
	priLANG
	priFAMILY_WEAK
	priPOSTSCRIPT_NAME_WEAK
	priSYMBOL
	priSPACING
	priSIZE
	priPIXEL_SIZE
	priSTYLE
	priSLANT
	priWEIGHT
	priWIDTH
	priFONT_HAS_HINT
	priDECORATIVE
	priANTIALIAS
	priRASTERIZER
	priOUTLINE
	priORDER
	priFONTVERSION
	priorityEnd

	priFILE_WEAK            = priFILE
	priFILE_STRONG          = priFILE
	priFONTFORMAT_WEAK      = priFONTFORMAT
	priFONTFORMAT_STRONG    = priFONTFORMAT
	priVARIABLE_WEAK        = priVARIABLE
	priVARIABLE_STRONG      = priVARIABLE
	priSCALABLE_WEAK        = priSCALABLE
	priSCALABLE_STRONG      = priSCALABLE
	priCOLOR_WEAK           = priCOLOR
	priCOLOR_STRONG         = priCOLOR
	priFOUNDRY_WEAK         = priFOUNDRY
	priFOUNDRY_STRONG       = priFOUNDRY
	priCHARSET_WEAK         = priCHARSET
	priCHARSET_STRONG       = priCHARSET
	priLANG_WEAK            = priLANG
	priLANG_STRONG          = priLANG
	priSYMBOL_WEAK          = priSYMBOL
	priSYMBOL_STRONG        = priSYMBOL
	priSPACING_WEAK         = priSPACING
	priSPACING_STRONG       = priSPACING
	priSIZE_WEAK            = priSIZE
	priSIZE_STRONG          = priSIZE
	priPIXEL_SIZE_WEAK      = priPIXEL_SIZE
	priPIXEL_SIZE_STRONG    = priPIXEL_SIZE
	priSTYLE_WEAK           = priSTYLE
	priSTYLE_STRONG         = priSTYLE
	priSLANT_WEAK           = priSLANT
	priSLANT_STRONG         = priSLANT
	priWEIGHT_WEAK          = priWEIGHT
	priWEIGHT_STRONG        = priWEIGHT
	priWIDTH_WEAK           = priWIDTH
	priWIDTH_STRONG         = priWIDTH
	priFONT_HAS_HINT_WEAK   = priFONT_HAS_HINT
	priFONT_HAS_HINT_STRONG = priFONT_HAS_HINT
	priDECORATIVE_WEAK      = priDECORATIVE
	priDECORATIVE_STRONG    = priDECORATIVE
	priANTIALIAS_WEAK       = priANTIALIAS
	priANTIALIAS_STRONG     = priANTIALIAS
	priRASTERIZER_WEAK      = priRASTERIZER
	priRASTERIZER_STRONG    = priRASTERIZER
	priOUTLINE_WEAK         = priOUTLINE
	priOUTLINE_STRONG       = priOUTLINE
	priORDER_WEAK           = priORDER
	priORDER_STRONG         = priORDER
	priFONTVERSION_WEAK     = priFONTVERSION
	priFONTVERSION_STRONG   = priFONTVERSION
)

type matcher struct {
	compare      func(v1, v2 Value) (Value, float32)
	object       Object
	strong, weak matcherPriority
}

// Order is significant, it defines the precedence of
// each value, earlier values are more significant than
// later values
var fcMatchers = [...]matcher{
	{nil, invalid, -1, -1},
	{compareFamily, FAMILY, priFAMILY_STRONG, priFAMILY_WEAK},
	{nil, FAMILYLANG, -1, -1},
	{compareString, STYLE, priSTYLE_STRONG, priSTYLE_WEAK},
	{nil, STYLELANG, -1, -1},
	{nil, FULLNAME, -1, -1},
	{nil, FULLNAMELANG, -1, -1},
	{compareNumber, SLANT, priSLANT_STRONG, priSLANT_WEAK},
	{compareRange, WEIGHT, priWEIGHT_STRONG, priWEIGHT_WEAK},
	{compareRange, WIDTH, priWIDTH_STRONG, priWIDTH_WEAK},
	{compareSize, SIZE, priSIZE_STRONG, priSIZE_WEAK},
	{nil, ASPECT, -1, -1},
	{compareNumber, PIXEL_SIZE, priPIXEL_SIZE_STRONG, priPIXEL_SIZE_WEAK},
	{compareNumber, SPACING, priSPACING_STRONG, priSPACING_WEAK},
	{compareString, FOUNDRY, priFOUNDRY_STRONG, priFOUNDRY_WEAK},
	{compareBool, ANTIALIAS, priANTIALIAS_STRONG, priANTIALIAS_WEAK},
	{nil, HINT_STYLE, -1, -1},
	{nil, HINTING, -1, -1},
	{nil, VERTICAL_LAYOUT, -1, -1},
	{nil, AUTOHINT, -1, -1},
	{nil, GLOBAL_ADVANCE, -1, -1},
	{compareFilename, FILE, priFILE_STRONG, priFILE_WEAK},
	{nil, INDEX, -1, -1},
	{compareString, RASTERIZER, priRASTERIZER_STRONG, priRASTERIZER_WEAK},
	{compareBool, OUTLINE, priOUTLINE_STRONG, priOUTLINE_WEAK},
	{compareBool, SCALABLE, priSCALABLE_STRONG, priSCALABLE_WEAK},
	{nil, DPI, -1, -1},
	{nil, RGBA, -1, -1},
	{nil, SCALE, -1, -1},
	{nil, MINSPACE, -1, -1},
	{nil, CHARWIDTH, -1, -1},
	{nil, CHAR_HEIGHT, -1, -1},
	{nil, MATRIX, -1, -1},
	{compareCharSet, CHARSET, priCHARSET_STRONG, priCHARSET_WEAK},
	{compareLang, LANG, priLANG_STRONG, priLANG_WEAK},
	{compareNumber, FONTVERSION, priFONTVERSION_STRONG, priFONTVERSION_WEAK},
	{nil, CAPABILITY, -1, -1},
	{compareString, FONTFORMAT, priFONTFORMAT_STRONG, priFONTFORMAT_WEAK},
	{nil, EMBOLDEN, -1, -1},
	{nil, EMBEDDED_BITMAP, -1, -1},
	{compareBool, DECORATIVE, priDECORATIVE_STRONG, priDECORATIVE_WEAK},
	{nil, LCD_FILTER, -1, -1},
	{nil, NAMELANG, -1, -1},
	{nil, FONT_FEATURES, -1, -1},
	{nil, PRGNAME, -1, -1},
	{nil, HASH, -1, -1},
	{comparePostScript, POSTSCRIPT_NAME, priPOSTSCRIPT_NAME_STRONG, priPOSTSCRIPT_NAME_WEAK},
	{compareBool, COLOR, priCOLOR_STRONG, priCOLOR_WEAK},
	{compareBool, SYMBOL, priSYMBOL_STRONG, priSYMBOL_WEAK},
	{nil, FONT_VARIATIONS, -1, -1},
	{compareBool, VARIABLE, priVARIABLE_STRONG, priVARIABLE_WEAK},
	{compareBool, FONT_HAS_HINT, priFONT_HAS_HINT_STRONG, priFONT_HAS_HINT_WEAK},
	{compareNumber, ORDER, priORDER_STRONG, priORDER_WEAK},
}

func (object Object) toMatcher(includeLang bool) *matcher {
	if includeLang {
		switch object {
		case FAMILYLANG, STYLELANG, FULLNAMELANG:
			object = LANG
		}
	}
	if int(object) >= len(fcMatchers) ||
		fcMatchers[object].compare == nil ||
		fcMatchers[object].strong == -1 ||
		fcMatchers[object].weak == -1 {
		return nil
	}

	return &fcMatchers[object]
}

func fdFromPatternList(object Object, match *matcher,
	pattern, target valueList, value []float32) (Value, Result, int, bool) {
	if match == nil {
		return target[0].Value, 0, 0, true
	}
	var (
		result    Result
		bestValue Value
		pos       int
	)
	weak := match.weak
	strong := match.strong

	best := float32(math.MaxFloat32)
	bestStrong := float32(math.MaxFloat32)
	bestWeak := float32(math.MaxFloat32)
	for j, v1 := range pattern {
		for k, v2 := range target {
			matchValue, v := match.compare(v1.Value, v2.Value)
			if v < 0 {
				result = ResultTypeMismatch
				return nil, result, 0, false
			}
			v = v*1000 + float32(j)
			if v < best {
				bestValue = matchValue
				best = v
				pos = k
			}
			if weak == strong {
				// found the best possible match
				if best < 1000 {
					goto done
				}
			} else if v1.Binding == vbStrong {
				if v < bestStrong {
					bestStrong = v
				}
			} else {
				if v < bestWeak {
					bestWeak = v
				}
			}
		}
	}
done:

	if debugMode {
		fmt.Printf("\tbest score for object <%s>: %g\n", object, best)
		fmt.Println("\tpattern values :", pattern, "target:", target)
	}

	if value != nil {
		if weak == strong {
			value[strong] += best
		} else {
			value[weak] += bestWeak
			value[strong] += bestStrong
		}
	}
	return bestValue, result, pos, true
}

type compareData = blankCaseMap

func (pat Pattern) newCompareData() compareData {
	table := make(blankCaseMap)

	elt := pat.getVals(FAMILY)
	for i, l := range elt {
		key := string(l.hash()) // l must have type string, but we are cautious
		e, ok := table.lookup(key)
		if !ok {
			e = new(familyEntry)
			e.strongValue = math.MaxFloat32
			e.weakValue = math.MaxFloat32
			table.add(key, e)
		}
		if l.Binding == vbWeak {
			if i := float32(i); i < e.weakValue {
				e.weakValue = i
			}
		} else {
			if i := float32(i); i < e.strongValue {
				e.strongValue = i
			}
		}
	}

	return table
}

func (table blankCaseMap) compareFamilies(v2orig valueList, value []float32) {
	strongValue := float32(math.MaxFloat32)
	weakValue := float32(math.MaxFloat32)

	for _, v2 := range v2orig {
		key, _ := v2.Value.(String) // should be string, but we are cautious
		e, ok := table.lookup(string(key))
		if ok {
			if e.strongValue < strongValue {
				strongValue = e.strongValue
			}
			if e.weakValue < weakValue {
				weakValue = e.weakValue
			}
		}
	}

	value[priFAMILY_STRONG] = strongValue
	value[priFAMILY_WEAK] = weakValue
}

// compare returns a value indicating the distance between the two lists of values
func (data compareData) compare(pat, font Pattern, value []float32) (bool, Result) {
	for i := range value {
		value[i] = 0.0
	}

	var result Result
	for obj1, eltI1 := range pat {
		eltI2, ok := font[obj1]
		if !ok {
			continue
		}

		if obj1 == FAMILY && data != nil {
			data.compareFamilies(*eltI2, value)
		} else {
			match := obj1.toMatcher(false)
			_, result, _, ok = fdFromPatternList(obj1, match, *eltI1, *eltI2, value)
			if !ok {
				return false, result
			}
		}
	}
	return true, result
}

// PrepareRender creates a new pattern consisting of elements of `font` not appearing
// in `pat`, elements of `pat` not appearing in `font` and the best matching
// value from `pat` for elements appearing in both.  The result is passed to
// `config.Substitute` with `kind = MatchResult` and then returned.
func (config *Config) PrepareRender(pat, font Pattern) Pattern {
	var (
		variations strings.Builder
		v          Value
	)

	variable, _ := font.GetBool(VARIABLE)

	new := NewPattern()

	for _, obj := range font.sortedKeys() {
		fe := font.getVals(obj)
		if obj == FAMILYLANG || obj == STYLELANG || obj == FULLNAMELANG {
			// ignore those objects. we need to deal with them another way
			continue
		}
		if obj == FAMILY || obj == STYLE || obj == FULLNAME {
			// using the fact that FAMILY + 1 == FAMILYLANG,
			// STYLE + 1 == STYLELANG,  FULLNAME + 1 == FULLNAMELANG
			lObject := obj + 1
			fel, pel := font.getVals(lObject), pat[lObject]

			if fel != nil && pel != nil {
				// The font has name languages, and pattern asks for specific language(s).
				// Match on language and and prefer that result.
				// Note:  Currently the code only give priority to first matching language.
				var (
					n  int
					ok bool
				)
				match := lObject.toMatcher(true)
				_, _, n, ok = fdFromPatternList(lObject, match, *pel, fel, nil)
				if !ok {
					return nil
				}

				var ln, ll valueList
				for j := 0; j < len(fe) || j < len(fel); j++ {
					if j == n {
						if j < len(fe) {
							ln = ln.prepend(valueElt{Value: fe[j].Value, Binding: vbStrong})
						}
						if j < len(fel) {
							ll = ll.prepend(valueElt{Value: fel[j].Value, Binding: vbStrong})
						}
					} else {
						if j < len(fe) {
							ln = append(ln, valueElt{Value: fe[j].Value, Binding: vbStrong})
						}
						if j < len(fel) {
							ll = append(ll, valueElt{Value: fel[j].Value, Binding: vbStrong})
						}
					}
				}
				new.addList(obj, ln, false)
				new.addList(lObject, ll, false)

				continue
			} else if fel != nil {
				//  Pattern doesn't ask for specific language.  Copy all for name and lang
				new.addList(obj, *fe.duplicate(), false)
				new.addList(lObject, *fel.duplicate(), false)

				continue
			}
		}

		pe := pat.getVals(obj)
		if pe != nil {
			match := obj.toMatcher(false)
			var ok bool
			v, _, _, ok = fdFromPatternList(obj, match, pe, fe, nil)
			if !ok {
				return nil
			}
			new.Add(obj, v, false)

			// Set font-variations settings for standard axes in variable fonts.
			if _, isRange := fe[0].Value.(Range); variable != 0 && isRange &&
				(obj == WEIGHT || obj == WIDTH || obj == SIZE) {
				tag := "    "
				num := float32(v.(Float)) //  v.type == FcTypeDouble
				if variations.Len() != 0 {
					variations.WriteByte(',')
				}
				switch obj {
				case WEIGHT:
					tag = "wght"
					num = WeightToOT(num)
				case WIDTH:
					tag = "wdth"
				case SIZE:
					tag = "opsz"
				}
				fmt.Fprintf(&variations, "%4s=%g", tag, num)
			}
		} else {
			new.addList(obj, *fe.duplicate(), true)
		}
	}
	for _, obj := range pat.sortedKeys() {
		pe := pat.getVals(obj)
		fe := font[obj]
		if fe == nil &&
			obj != FAMILYLANG && obj != STYLELANG && obj != FULLNAMELANG {
			new.addList(obj, *pe.duplicate(), false)
		}
	}

	if variable != 0 && variations.Len() != 0 {
		if vars, res := new.getAtString(FONT_VARIATIONS, 0); res == ResultMatch {
			variations.WriteByte(',')
			variations.WriteString(vars)
			new.Del(FONT_VARIATIONS)
		}

		new.Add(FONT_VARIATIONS, String(variations.String()), true)
	}

	config.Substitute(new, pat, MatchResult)
	return new
}

func (set Fontset) matchInternal(p Pattern) Pattern {
	var (
		score, bestscore [priorityEnd]float32
		best             Pattern
	)

	if debugMode {
		fmt.Println()
		fmt.Println("Starting match to")
		fmt.Println(p.String())
	}

	data := p.newCompareData()

	for f, pat := range set {
		if debugMode {
			fmt.Printf("Font %d: %s", f, pat)
		}
		ok, _ := data.compare(p, pat, score[:])
		if !ok {
			return nil
		}
		if debugMode {
			fmt.Println("Score     ", score)
			fmt.Println()
		}
		for i, bs := range bestscore {
			if best != nil && bs < score[i] {
				break
			}
			if best == nil || score[i] < bs {
				for j, s := range score {
					bestscore[j] = s
				}
				best = pat
				break
			}
		}
	}

	if debugMode {
		fmt.Println("Best score", bestscore)
		fmt.Println()
	}

	return best
}

// Sort returns the list of fonts from `set` sorted by closeness to `pattern`.
// If `trim` is true, elements in the list which don't include Unicode coverage not provided by
// earlier elements in the list are elided.
// The union of Unicode coverage of all of the fonts is returned.
// This function should be called only after `Config.Substitute` and `Pattern.SubstituteDefault` have
// been called for `p`; otherwise the results will not be correct.
// The returned `Fontset` references Pattern structures which may be shared
// by the return value from multiple `Sort` calls, applications cannot
// modify these patterns. Instead, they should be passed, along with
// `p`, to `Config.PrepareRender()` which combines them into a complete pattern.
func (set Fontset) Sort(p Pattern, trim bool) (Fontset, Charset) {
	if debugMode {
		fmt.Println("Sort input :", p.String())
	}

	nPatternLang := 0
	for res := ResultMatch; res == ResultMatch; nPatternLang++ {
		_, res = p.GetAt(LANG, nPatternLang)
	}

	nodes := make([]*sortNode, len(set))
	patternLangSat := make([]bool, nPatternLang)

	data := p.newCompareData()

	for i, font := range set {
		if debugMode {
			f, _ := font.GetString(FILE)
			fmt.Printf("Font %d: %s\n", i, f)
		}
		newPtr := new(sortNode)
		newPtr.pattern = font
		ok, _ := data.compare(p, newPtr.pattern, newPtr.score[:])
		if !ok {
			return nil, Charset{}
		}
		if debugMode {
			fmt.Println("Score", newPtr.score)
		}
		nodes[i] = newPtr
	}

	sort.Slice(nodes, func(i, j int) bool { return sortCompare(nodes[i], nodes[j]) })

	for _, node := range nodes {
		satisfies := false
		//  if this node matches any language, check which ones and satisfy those entries
		if node.score[priLANG] < 2000 {
			for i, pls := range patternLangSat {
				patternLang, res1 := p.GetAt(LANG, i)
				nodeLang, res2 := node.pattern.GetAt(LANG, 0)
				if !pls && res1 == ResultMatch && res2 == ResultMatch {
					_, compare := compareLang(patternLang, nodeLang)
					if compare >= 0 && compare < 2 {
						patternLangSat[i] = true
						satisfies = true
						break
					}
				}
			}
		}
		if !satisfies {
			node.score[priLANG] = 10000
		}
	}

	// re-sort once the language issues have been settled
	sort.Slice(nodes, func(i, j int) bool { return sortCompare(nodes[i], nodes[j]) })

	return sortWalk(nodes, trim)
}

// Match finds the font in `set` most closely matching
// `pattern` and returns the result of
// `config.PrepareRender` for that font and the provided
// pattern. This function should be called only after
// `config.Substitute` and `p.SubstituteDefault` have been called for
// `p`; otherwise the results will not be correct.
func (set Fontset) Match(p Pattern, config *Config) Pattern {
	best := set.matchInternal(p)
	if best == nil {
		return nil
	}
	return config.PrepareRender(p, best)
}

type sortNode struct {
	pattern Pattern
	score   [priorityEnd]float32
}

func sortCompare(a, b *sortNode) bool {
	var ad, bd float32
	for i := range a.score {
		ad, bd = a.score[i], b.score[i]
		if ad != bd {
			break
		}
	}

	return ad < bd
}

func sortWalk(n []*sortNode, trim bool) (Fontset, Charset) {
	var (
		fs  Fontset
		csp Charset
	)
	for i, node := range n {
		addsChars := false

		// Only fetch node charset if we'd need it
		if trim {
			ncs, ok := node.pattern.GetCharset(CHARSET)
			if !ok {
				continue
			}
			addsChars = csp.merge(ncs)
		}

		// If this font isn't a subset of the previous fonts, add it to the list
		if i == 0 || !trim || addsChars {
			if debugMode {
				fmt.Println("Add ", node.pattern.String())
			}
			fs = append(fs, node.pattern)
		}
	}

	return fs, csp
}
