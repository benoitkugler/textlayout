package fontconfig

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ported from fontconfig/src/fcmatch.c Copyright © 2000 Keith Packard

type FcMatchKind int8

const (
	FcMatchDefault FcMatchKind = iota - 1
	FcMatchPattern
	FcMatchFont
	FcMatchScan
	FcMatchKindEnd
	FcMatchKindBegin = FcMatchPattern
)

func (k FcMatchKind) String() string {
	switch k {
	case FcMatchPattern:
		return "pattern"
	case FcMatchFont:
		return "font"
	default:
		return fmt.Sprintf("match kind <%d>", k)
	}
}

func FcCompareNumber(value1, value2 FcValue) (FcValue, float64) {
	var v1, v2 float64
	switch value := value1.(type) {
	case Int:
		v1 = float64(value)
	case Float:
		v1 = float64(value)
	default:
		return nil, -1.0
	}
	switch value := value2.(type) {
	case Int:
		v2 = float64(value)
	case Float:
		v2 = float64(value)
	default:
		return nil, -1.0
	}

	v := v2 - v1
	if v < 0 {
		v = -v
	}
	return value2, v
}

func FcCompareString(v1, v2 FcValue) (FcValue, float64) {
	bestValue := v2
	if strings.EqualFold(string(v1.(String)), string(v2.(String))) {
		return bestValue, 0
	}
	return bestValue, 1
}

// returns 0 for empty strings
func FcToLower(s string) byte {
	if s == "" {
		return 0
	}
	if 0101 <= s[0] && s[0] <= 0132 {
		return s[0] - 0101 + 0141
	}
	return s[0]
}

func FcCompareFamily(v1, v2 FcValue) (FcValue, float64) {
	// rely on the guarantee in FcPatternObjectAddWithBinding that
	// families are always FcTypeString.
	v1_string := string(v1.(String))
	v2_string := string(v2.(String))

	bestValue := v2

	if FcToLower(v1_string) != FcToLower(v2_string) &&
		v1_string[0] != ' ' && v2_string[0] != ' ' {
		return bestValue, 1.0
	}

	if ignoreBlanksAndCase(v1_string) == ignoreBlanksAndCase(v2_string) {
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

func FcComparePostScript(v1, v2 FcValue) (FcValue, float64) {
	v1_string := string(v1.(String))
	v2_string := string(v2.(String))

	bestValue := v2

	if FcToLower(v1_string) != FcToLower(v2_string) &&
		v1_string[0] != ' ' && v2_string[0] != ' ' {
		return bestValue, 1.0
	}

	n := matchIgnoreCaseAndDelims(v1_string, v2_string)
	length := len(v1_string)

	return bestValue, float64(length-n) / float64(length)
}

func FcCompareLang(val1, val2 FcValue) (FcValue, float64) {
	var result FcLangResult
	switch v1 := val1.(type) {
	case FcLangSet:
		switch v2 := val2.(type) {
		case FcLangSet:
			result = FcLangSetCompare(v1, v2)
		case String:
			result = v1.hasLang(string(v2))
		default:
			return nil, -1.0
		}
	case String:
		switch v2 := val2.(type) {
		case *FcLangSet:
			result = v2.hasLang(string(v1))
		case String:
			result = FcLangCompare(string(v1), string(v2))
		default:
			return nil, -1.0
		}
		break
	default:
		return nil, -1.0
	}
	bestValue := val2
	switch result {
	case FcLangEqual:
		return bestValue, 0
	case FcLangDifferentCountry:
		return bestValue, 1
	default:
		return bestValue, 2
	}
}

func FcCompareBool(val1, val2 FcValue) (FcValue, float64) {
	v1, ok1 := val1.(FcBool)
	v2, ok2 := val2.(FcBool)
	if !ok1 || !ok2 {
		return nil, -1.0
	}

	var bestValue FcBool
	if v2 != FcDontCare {
		bestValue = v2
	} else {
		bestValue = v1
	}

	if v1 == v2 {
		return bestValue, 0
	}
	return bestValue, 1
}

func FcCompareCharSet(v1, v2 FcValue) (FcValue, float64) {
	bestValue := v2
	return bestValue, float64(charsetSubtractCount(v1.(FcCharset), v2.(FcCharset)))
}

func FcCompareRange(v1, v2 FcValue) (FcValue, float64) {
	var b1, e1, b2, e2, d float64

	switch value1 := v1.(type) {
	case Int:
		e1 = float64(value1)
		b1 = e1
	case Float:
		e1 = float64(value1)
		b1 = e1
	case FcRange:
		b1 = value1.Begin
		e1 = value1.End
	default:
		return nil, -1
	}
	switch value2 := v2.(type) {
	case Int:
		e2 = float64(value2)
		b2 = e2
	case Float:
		e2 = float64(value2)
		b2 = e2
	case FcRange:
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
		d = (math.Max(b1, b2) + math.Min(e1, e2)) * .5
	}

	bestValue := Float(d)

	/// if the ranges overlap, it's a match, otherwise return closest distance.
	if e1 < b2 || e2 < b1 {
		return bestValue, math.Min(math.Abs(b2-e1), math.Abs(b1-e2))
	}
	return bestValue, 0.0
}

func FcCompareSize(v1, v2 FcValue) (FcValue, float64) {
	var b1, e1, b2, e2 float64

	switch value1 := v1.(type) {
	case Int:
		e1 = float64(value1)
		b1 = e1
	case Float:
		e1 = float64(value1)
		b1 = e1
	case FcRange:
		b1 = value1.Begin
		e1 = value1.End
	default:
		return nil, -1
	}
	switch value2 := v2.(type) {
	case Int:
		e2 = float64(value2)
		b2 = e2
	case Float:
		e2 = float64(value2)
		b2 = e2
	case FcRange:
		b2 = value2.Begin
		e2 = value2.End
	default:
		return nil, -1
	}

	bestValue := Float((b1 + e1) * .5)

	// if the ranges overlap, it's a match, otherwise return closest distance.
	if e1 < b2 || e2 < b1 {
		return bestValue, math.Min(math.Abs(b2-e1), math.Abs(b1-e2))
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

func FcCompareFilename(v1, v2 FcValue) (FcValue, float64) {
	s1, s2 := string(v1.(String)), string(v2.(String))
	bestValue := String(s2)
	if s1 == s2 {
		return bestValue, 0.0
	}
	if strings.EqualFold(s1, s2) {
		return bestValue, 1.0
	}
	if strGlobMatch(s1, s2) {
		return bestValue, 2.0
	}
	return bestValue, 3.0
}

// Canonical match priority order
type FcMatcherPriority int8

const (
	PRI_FILE FcMatcherPriority = iota
	PRI_FONTFORMAT
	PRI_VARIABLE
	PRI_SCALABLE
	PRI_COLOR
	PRI_FOUNDRY
	PRI_CHARSET
	PRI_FAMILY_STRONG
	PRI_POSTSCRIPT_NAME_STRONG
	PRI_LANG
	PRI_FAMILY_WEAK
	PRI_POSTSCRIPT_NAME_WEAK
	PRI_SYMBOL
	PRI_SPACING
	PRI_SIZE
	PRI_PIXEL_SIZE
	PRI_STYLE
	PRI_SLANT
	PRI_WEIGHT
	PRI_WIDTH
	PRI_FONT_HAS_HINT
	PRI_DECORATIVE
	PRI_ANTIALIAS
	PRI_RASTERIZER
	PRI_OUTLINE
	PRI_ORDER
	PRI_FONTVERSION
	PRI_END

	PRI_FILE_WEAK            = PRI_FILE
	PRI_FILE_STRONG          = PRI_FILE
	PRI_FONTFORMAT_WEAK      = PRI_FONTFORMAT
	PRI_FONTFORMAT_STRONG    = PRI_FONTFORMAT
	PRI_VARIABLE_WEAK        = PRI_VARIABLE
	PRI_VARIABLE_STRONG      = PRI_VARIABLE
	PRI_SCALABLE_WEAK        = PRI_SCALABLE
	PRI_SCALABLE_STRONG      = PRI_SCALABLE
	PRI_COLOR_WEAK           = PRI_COLOR
	PRI_COLOR_STRONG         = PRI_COLOR
	PRI_FOUNDRY_WEAK         = PRI_FOUNDRY
	PRI_FOUNDRY_STRONG       = PRI_FOUNDRY
	PRI_CHARSET_WEAK         = PRI_CHARSET
	PRI_CHARSET_STRONG       = PRI_CHARSET
	PRI_LANG_WEAK            = PRI_LANG
	PRI_LANG_STRONG          = PRI_LANG
	PRI_SYMBOL_WEAK          = PRI_SYMBOL
	PRI_SYMBOL_STRONG        = PRI_SYMBOL
	PRI_SPACING_WEAK         = PRI_SPACING
	PRI_SPACING_STRONG       = PRI_SPACING
	PRI_SIZE_WEAK            = PRI_SIZE
	PRI_SIZE_STRONG          = PRI_SIZE
	PRI_PIXEL_SIZE_WEAK      = PRI_PIXEL_SIZE
	PRI_PIXEL_SIZE_STRONG    = PRI_PIXEL_SIZE
	PRI_STYLE_WEAK           = PRI_STYLE
	PRI_STYLE_STRONG         = PRI_STYLE
	PRI_SLANT_WEAK           = PRI_SLANT
	PRI_SLANT_STRONG         = PRI_SLANT
	PRI_WEIGHT_WEAK          = PRI_WEIGHT
	PRI_WEIGHT_STRONG        = PRI_WEIGHT
	PRI_WIDTH_WEAK           = PRI_WIDTH
	PRI_WIDTH_STRONG         = PRI_WIDTH
	PRI_FONT_HAS_HINT_WEAK   = PRI_FONT_HAS_HINT
	PRI_FONT_HAS_HINT_STRONG = PRI_FONT_HAS_HINT
	PRI_DECORATIVE_WEAK      = PRI_DECORATIVE
	PRI_DECORATIVE_STRONG    = PRI_DECORATIVE
	PRI_ANTIALIAS_WEAK       = PRI_ANTIALIAS
	PRI_ANTIALIAS_STRONG     = PRI_ANTIALIAS
	PRI_RASTERIZER_WEAK      = PRI_RASTERIZER
	PRI_RASTERIZER_STRONG    = PRI_RASTERIZER
	PRI_OUTLINE_WEAK         = PRI_OUTLINE
	PRI_OUTLINE_STRONG       = PRI_OUTLINE
	PRI_ORDER_WEAK           = PRI_ORDER
	PRI_ORDER_STRONG         = PRI_ORDER
	PRI_FONTVERSION_WEAK     = PRI_FONTVERSION
	PRI_FONTVERSION_STRONG   = PRI_FONTVERSION
)

type FcMatcher struct {
	object       FcObject
	compare      func(v1, v2 FcValue) (FcValue, float64)
	strong, weak FcMatcherPriority
}

// Order is significant, it defines the precedence of
// each value, earlier values are more significant than
// later values
var fcMatchers = [...]FcMatcher{
	{FC_INVALID, nil, -1, -1},
	{FC_FAMILY, FcCompareFamily, PRI_FAMILY_STRONG, PRI_FAMILY_WEAK},
	{FC_FAMILYLANG, nil, -1, -1},
	{FC_STYLE, FcCompareString, PRI_STYLE_STRONG, PRI_STYLE_WEAK},
	{FC_STYLELANG, nil, -1, -1},
	{FC_FULLNAME, nil, -1, -1},
	{FC_FULLNAMELANG, nil, -1, -1},
	{FC_SLANT, FcCompareNumber, PRI_SLANT_STRONG, PRI_SLANT_WEAK},
	{FC_WEIGHT, FcCompareRange, PRI_WEIGHT_STRONG, PRI_WEIGHT_WEAK},
	{FC_WIDTH, FcCompareRange, PRI_WIDTH_STRONG, PRI_WIDTH_WEAK},
	{FC_SIZE, FcCompareSize, PRI_SIZE_STRONG, PRI_SIZE_WEAK},
	{FC_ASPECT, nil, -1, -1},
	{FC_PIXEL_SIZE, FcCompareNumber, PRI_PIXEL_SIZE_STRONG, PRI_PIXEL_SIZE_WEAK},
	{FC_SPACING, FcCompareNumber, PRI_SPACING_STRONG, PRI_SPACING_WEAK},
	{FC_FOUNDRY, FcCompareString, PRI_FOUNDRY_STRONG, PRI_FOUNDRY_WEAK},
	{FC_ANTIALIAS, FcCompareBool, PRI_ANTIALIAS_STRONG, PRI_ANTIALIAS_WEAK},
	{FC_HINT_STYLE, nil, -1, -1},
	{FC_HINTING, nil, -1, -1},
	{FC_VERTICAL_LAYOUT, nil, -1, -1},
	{FC_AUTOHINT, nil, -1, -1},
	{FC_GLOBAL_ADVANCE, nil, -1, -1},
	{FC_FILE, FcCompareFilename, PRI_FILE_STRONG, PRI_FILE_WEAK},
	{FC_INDEX, nil, -1, -1},
	{FC_RASTERIZER, FcCompareString, PRI_RASTERIZER_STRONG, PRI_RASTERIZER_WEAK},
	{FC_OUTLINE, FcCompareBool, PRI_OUTLINE_STRONG, PRI_OUTLINE_WEAK},
	{FC_SCALABLE, FcCompareBool, PRI_SCALABLE_STRONG, PRI_SCALABLE_WEAK},
	{FC_DPI, nil, -1, -1},
	{FC_RGBA, nil, -1, -1},
	{FC_SCALE, nil, -1, -1},
	{FC_MINSPACE, nil, -1, -1},
	{FC_CHARWIDTH, nil, -1, -1},
	{FC_CHAR_HEIGHT, nil, -1, -1},
	{FC_MATRIX, nil, -1, -1},
	{FC_CHARSET, FcCompareCharSet, PRI_CHARSET_STRONG, PRI_CHARSET_WEAK},
	{FC_LANG, FcCompareLang, PRI_LANG_STRONG, PRI_LANG_WEAK},
	{FC_FONTVERSION, FcCompareNumber, PRI_FONTVERSION_STRONG, PRI_FONTVERSION_WEAK},
	{FC_CAPABILITY, nil, -1, -1},
	{FC_FONTFORMAT, FcCompareString, PRI_FONTFORMAT_STRONG, PRI_FONTFORMAT_WEAK},
	{FC_EMBOLDEN, nil, -1, -1},
	{FC_EMBEDDED_BITMAP, nil, -1, -1},
	{FC_DECORATIVE, FcCompareBool, PRI_DECORATIVE_STRONG, PRI_DECORATIVE_WEAK},
	{FC_LCD_FILTER, nil, -1, -1},
	{FC_NAMELANG, nil, -1, -1},
	{FC_FONT_FEATURES, nil, -1, -1},
	{FC_PRGNAME, nil, -1, -1},
	{FC_HASH, nil, -1, -1},
	{FC_POSTSCRIPT_NAME, FcComparePostScript, PRI_POSTSCRIPT_NAME_STRONG, PRI_POSTSCRIPT_NAME_WEAK},
	{FC_COLOR, FcCompareBool, PRI_COLOR_STRONG, PRI_COLOR_WEAK},
	{FC_SYMBOL, FcCompareBool, PRI_SYMBOL_STRONG, PRI_SYMBOL_WEAK},
	{FC_FONT_VARIATIONS, nil, -1, -1},
	{FC_VARIABLE, FcCompareBool, PRI_VARIABLE_STRONG, PRI_VARIABLE_WEAK},
	{FC_FONT_HAS_HINT, FcCompareBool, PRI_FONT_HAS_HINT_STRONG, PRI_FONT_HAS_HINT_WEAK},
	{FC_ORDER, FcCompareNumber, PRI_ORDER_STRONG, PRI_ORDER_WEAK},
}

func (object FcObject) toMatcher(includeLang bool) *FcMatcher {
	if includeLang {
		switch object {
		case FC_FAMILYLANG, FC_STYLELANG, FC_FULLNAMELANG:
			object = FC_LANG
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

func compareValueList(object FcObject, match *FcMatcher,
	v1orig FcValueList, /* pattern */
	v2orig FcValueList, /* target */
	value []float64) (FcValue, FcResult, int, bool) {

	if match == nil {
		return v2orig[0].value, 0, 0, true
	}
	var (
		result    FcResult
		bestValue FcValue
		pos       int
	)
	weak := match.weak
	strong := match.strong

	best := 1e99
	bestStrong := 1e99
	bestWeak := 1e99
	for j, v1 := range v1orig {
		for k, v2 := range v2orig {
			matchValue, v := match.compare(v1.value, v2.value)
			if v < 0 {
				result = FcResultTypeMismatch
				return nil, result, 0, false
			}
			v = v*1000 + float64(j)
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
			} else if v1.binding == FcValueBindingStrong {
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
		fmt.Printf(" %s: %g ", object, best)
		fmt.Println(v1orig)
		fmt.Print(", ")
		fmt.Println(v2orig)
		fmt.Println()
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

type FcCompareData = blankCaseMap

func (pat FcPattern) newCompareData() FcCompareData {
	table := make(blankCaseMap)

	elt := pat[FC_FAMILY]
	for i, l := range elt {
		key := string(l.hash()) // l must have type string, but we are cautious
		e, ok := table.lookup(key)
		if !ok {
			e = new(familyEntry)
			e.strongValue = 1e99
			e.weakValue = 1e99
			table.add(key, e)
		}
		if l.binding == FcValueBindingWeak {
			if i := float64(i); i < e.weakValue {
				e.weakValue = i
			}
		} else {
			if i := float64(i); i < e.strongValue {
				e.strongValue = i
			}
		}
	}

	return table
}

func (table blankCaseMap) FcCompareFamilies(v2orig FcValueList, value []float64) {
	strong_value := 1e99
	weak_value := 1e99

	for _, v2 := range v2orig {
		key := string(v2.hash()) // should be string, but we are cautious
		e, ok := table.lookup(key)
		if ok {
			if e.strongValue < strong_value {
				strong_value = e.strongValue
			}
			if e.weakValue < weak_value {
				weak_value = e.weakValue
			}
		}
	}

	value[PRI_FAMILY_STRONG] = strong_value
	value[PRI_FAMILY_WEAK] = weak_value
}

// compare returns a value indicating the distance between the two lists of values
func (data FcCompareData) compare(pat, fnt FcPattern, value []float64) (bool, FcResult) {
	for i := range value {
		value[i] = 0.0
	}

	var result FcResult
	for i1, elt_i1 := range pat {
		elt_i2, ok := fnt[i1]
		if !ok {
			continue
		}

		if i1 == FC_FAMILY && data != nil {
			data.FcCompareFamilies(elt_i2, value)
		} else {
			match := i1.toMatcher(false)
			_, result, _, ok = compareValueList(i1, match, elt_i1, elt_i2, value)
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
// FcConfigSubstituteWithPat with `kind` FcMatchFont and then returned. As in `FcConfigSubstituteWithPat`,
// a nil config may be used, defaulting to the current configuration.
func (pat FcPattern) PrepareRender(font FcPattern, config *FcConfig) FcPattern {
	//  FcPattern	    *new;
	//  int		    i;
	//  FcPatternElt    *fe, *pe;
	//  FcBool	    variable = false;
	var (
		variations strings.Builder
		v          FcValue
	)

	variable, _ := font.FcPatternObjectGetBool(FC_VARIABLE, 0)

	new := NewFcPattern()

	for _, obj := range font.sortedKeys() {
		fe := font[obj]
		if obj == FC_FAMILYLANG || obj == FC_STYLELANG || obj == FC_FULLNAMELANG {
			// ignore those objects. we need to deal with them another way
			continue
		}
		if obj == FC_FAMILY || obj == FC_STYLE || obj == FC_FULLNAME {
			// using the fact that FC_FAMILY + 1 == FC_FAMILYLANG,
			// FC_STYLE + 1 == FC_STYLELANG,  FC_FULLNAME + 1 == FC_FULLNAMELANG
			lObject := obj + 1
			fel, pel := font[lObject], pat[lObject]

			if fel != nil && pel != nil {
				// The font has name languages, and pattern asks for specific language(s).
				// Match on language and and prefer that result.
				// Note:  Currently the code only give priority to first matching language.
				var (
					n  int
					ok bool
				)
				match := lObject.toMatcher(true)
				_, _, n, ok = compareValueList(lObject, match, pel, fel, nil)
				if !ok {
					return nil
				}

				var ln, ll FcValueList
				//  j = 0, l1 = FcPatternEltValues (fe), l2 = FcPatternEltValues (fel);
				// 	  l1 != nil || l2 != nil;
				// 	  j++, l1 = l1 ? FcValueListNext (l1) : nil, l2 = l2 ? FcValueListNext (l2) : nil)
				for j := 0; j < len(fe) || j < len(fel); j++ {
					if j == n {
						if j < len(fe) {
							ln = ln.prepend(valueElt{value: fe[j].value, binding: FcValueBindingStrong})
						}
						if j < len(fel) {
							ll = ll.prepend(valueElt{value: fel[j].value, binding: FcValueBindingStrong})
						}
					} else {
						if j < len(fe) {
							ln = append(ln, valueElt{value: fe[j].value, binding: FcValueBindingStrong})
						}
						if j < len(fel) {
							ll = append(ll, valueElt{value: fel[j].value, binding: FcValueBindingStrong})
						}
					}
				}
				new.AddList(obj, ln, false)
				new.AddList(lObject, ll, false)

				continue
			} else if fel != nil {
				//  Pattern doesn't ask for specific language.  Copy all for name and lang
				new.AddList(obj, fe.duplicate(), false)
				new.AddList(lObject, fel.duplicate(), false)

				continue
			}
		}

		pe := pat[obj]
		if pe != nil {
			match := obj.toMatcher(false)
			var ok bool
			v, _, _, ok = compareValueList(obj, match, pe, fe, nil)
			if !ok {
				return nil
			}
			new.Add(obj, v, false)

			// Set font-variations settings for standard axes in variable fonts.
			if _, isRange := fe[0].value.(FcRange); variable != 0 && isRange &&
				(obj == FC_WEIGHT || obj == FC_WIDTH || obj == FC_SIZE) {
				//  double num;
				//  FcChar8 temp[128];
				tag := "    "
				num := float64(v.(Float)) //  v.type == FcTypeDouble
				if variations.Len() != 0 {
					variations.WriteByte(',')
				}
				switch obj {
				case FC_WEIGHT:
					tag = "wght"
					num = FcWeightToOpenTypeDouble(num)
				case FC_WIDTH:
					tag = "wdth"
				case FC_SIZE:
					tag = "opsz"
				}
				fmt.Fprintf(&variations, "%4s=%g", tag, num)
			}
		} else {
			new.AddList(obj, fe.duplicate(), true)
		}
	}
	for _, obj := range pat.sortedKeys() {
		pe := pat[obj]
		fe := font[obj]
		if fe == nil &&
			obj != FC_FAMILYLANG && obj != FC_STYLELANG && obj != FC_FULLNAMELANG {
			new.AddList(obj, pe.duplicate(), false)
		}
	}

	if variable != 0 && variations.Len() != 0 {
		if vars, res := new.FcPatternObjectGetString(FC_FONT_VARIATIONS, 0); res == FcResultMatch {
			variations.WriteByte(',')
			variations.WriteString(vars)
			new.Del(FC_FONT_VARIATIONS)
		}

		new.Add(FC_FONT_VARIATIONS, String(variations.String()), true)
	}

	config.FcConfigSubstituteWithPat(new, pat, FcMatchFont)
	return new
}

func (p FcPattern) fontSetMatchInternal(sets []FcFontSet) (FcPattern, FcResult) {
	var (
		score, bestscore [PRI_END]float64
		best             FcPattern
		result           FcResult
	)

	if debugMode {
		fmt.Println("Match ")
		fmt.Println(p.String())
	}

	data := p.newCompareData()

	for _, s := range sets {
		if s == nil {
			continue
		}
		for f, pat := range s {
			if debugMode {
				fmt.Printf("Font %d %s", f, pat)
			}
			var ok bool
			ok, result = data.compare(p, pat, score[:])
			if !ok {
				return nil, result
			}
			if debugMode {
				fmt.Printf("Score %v\n", score)
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
	}

	if debugMode {
		fmt.Printf("Best score %v\n", bestscore)
	}

	if best != nil {
		result = FcResultMatch
	}

	return best, result
}

// Sort returns the list of fonts from `sets` sorted by closeness to `pattern`.
// If `trim` is true, elements in the list which don't include Unicode coverage not provided by
// earlier elements in the list are elided. The union of Unicode coverage of
// all of the fonts is returned in `csp`, if `csp` is not nil.  This function
// should be called only after FcConfigSubstitute and FcDefaultSubstitute have
// been called for `p`;
// otherwise the results will not be correct.
// The returned FcFontSet references FcPattern structures which may be shared
// by the return value from multiple FcFontSort calls, applications cannot
// modify these patterns. Instead, they should be passed, along with
// `pattern` to PrepareRender() which combines them into a complete pattern.
func Sort(sets []FcFontSet, p FcPattern, trim bool) (FcFontSet, FcCharset, FcResult) {
	//  assert (p != nil);

	// There are some implementation that relying on the result of
	// "result" to check if the return value of Sort
	// is valid or not.
	// So we should initialize it to the conservative way since
	// this function doesn't return nil anymore.
	result := FcResultNoMatch

	if debugMode {
		fmt.Println("Sort ", p.String())
	}
	nnodes := 0
	for _, s := range sets {
		nnodes += len(s)
	}
	if nnodes == 0 {
		return nil, FcCharset{}, result
	}

	var (
		patternLang  FcValue
		nPatternLang = 0
	)
	for res := FcResultMatch; res == FcResultMatch; nPatternLang++ {
		patternLang, res = p.FcPatternObjectGet(FC_LANG, nPatternLang)
	}

	nodes := make([]*FcSortNode, nnodes)
	patternLangSat := make([]bool, nPatternLang)

	data := p.newCompareData()

	index := 0
	for _, s := range sets {
		for f, font := range s {
			if debugMode {
				fmt.Printf("Font %d: %s\n", f, font)
			}
			newPtr := new(FcSortNode)
			newPtr.pattern = font
			var ok bool
			ok, result = data.compare(p, newPtr.pattern, newPtr.score[:])
			if !ok {
				return nil, FcCharset{}, result
			}
			if debugMode {
				fmt.Println("Score", newPtr.score)
			}
			nodes[index] = newPtr
			index++
		}
	}

	sort.Slice(nodes, func(i, j int) bool { return sortCompare(nodes[i], nodes[j]) })

	for _, node := range nodes {
		satisfies := false
		//  if this node matches any language, check which ones and satisfy those entries
		if node.score[PRI_LANG] < 2000 {
			for i, pls := range patternLangSat {
				var res1 FcResult
				patternLang, res1 = p.FcPatternObjectGet(FC_LANG, i)
				nodeLang, res2 := node.pattern.FcPatternObjectGet(FC_LANG, 0)
				if !pls && res1 == FcResultMatch && res2 == FcResultMatch {
					_, compare := FcCompareLang(patternLang, nodeLang)
					if compare >= 0 && compare < 2 {
						patternLangSat[i] = true
						satisfies = true
						break
					}
				}
			}
		}
		if !satisfies {
			node.score[PRI_LANG] = 10000
		}
	}

	// re-sort once the language issues have been settled
	sort.Slice(nodes, func(i, j int) bool { return sortCompare(nodes[i], nodes[j]) })

	var ret FcFontSet

	csp := FcSortWalk(nodes, &ret, trim)

	if len(ret) != 0 {
		result = FcResultMatch
	}

	return ret, csp, result
}

//  FcPattern *
//  FcFontSetMatch (FcConfig    *config,
// 		 FcFontSet   **sets,
// 		 int	    nsets,
// 		 FcPattern   *p,
// 		 FcResult    *result)
//  {
// 	 FcPattern	    *best, *ret = nil;

// 	 assert (sets != nil);
// 	 assert (p != nil);
// 	 assert (result != nil);

// 	 *result = FcResultNoMatch;

// 	 config = FcConfigReference (config);
// 	 if (!config)
// 		 return nil;
// 	 best = fontSetMatchInternal (sets, nsets, p, result);
// 	 if (best)
// 	 ret = PrepareRender (config, p, best);

// 	 FcConfigDestroy (config);

// 	 return ret;
//  }

// FcFontMatch finds the font in `sets` most closely matching
// `pattern` and returns the result of
// `PrepareRender` for that font and the provided
// pattern. This function should be called only after
// `FcConfigSubstitute` and  `FcDefaultSubstitute` have been called for
// `p`; otherwise the results will not be correct.
// If `config` is nil, the current configuration is used.
func (config *FcConfig) FcFontMatch(p FcPattern) (FcPattern, FcResult) {
	config = fallbackConfig(config)
	if config == nil {
		return nil, FcResultNoMatch
	}

	var sets []FcFontSet
	if config.fonts[FcSetSystem] != nil {
		sets = append(sets, config.fonts[FcSetSystem])
	}
	if config.fonts[FcSetApplication] != nil {
		sets = append(sets, config.fonts[FcSetApplication])
	}

	var ret FcPattern
	best, result := p.fontSetMatchInternal(sets)
	if best != nil {
		ret = p.PrepareRender(best, config)
	}

	return ret, result
}

type FcSortNode struct {
	pattern FcPattern
	score   [PRI_END]float64
}

func sortCompare(a, b *FcSortNode) bool {
	ad, bd := 0., 0.

	for i := range a.score {
		ad, bd = a.score[i], b.score[i]
		if ad != bd {
			break
		}
	}

	return ad < bd
}

func FcSortWalk(n []*FcSortNode, fs *FcFontSet, trim bool) FcCharset {
	var csp FcCharset

	for i, node := range n {
		//  FcSortNode	*node = *n++;
		addsChars := false

		// Only fetch node charset if we'd need it
		if trim {
			ncs, res := node.pattern.FcPatternObjectGetCharSet(FC_CHARSET, 0)
			if res != FcResultMatch {
				continue
			}
			addsChars = csp.merge(ncs)
		}

		// If this font isn't a subset of the previous fonts, add it to the list
		if i == 0 || !trim || addsChars {
			if debugMode {
				fmt.Println("Add ", node.pattern.String())
			}
			*fs = append(*fs, node.pattern)
		}
	}

	return csp
}

//  void
//  SortDestroy (FcFontSet *fs)
//  {
// 	 FcFontSetDestroy (fs);
//  }

//  FcFontSet *
//  FcFontSort (config *FcConfig,
// 		 p FcPattern,
// 		 FcBool	trim,
// 		 FcCharset	**csp,
// 		 result *FcResult)
//  {
// 	 FcFontSet	*sets[2], *ret;
// 	 int		nsets;

// 	 assert (p != nil);
// 	 assert (result != nil);

// 	 *result = FcResultNoMatch;

// 	 config = FcConfigReference (config);
// 	 if (!config)
// 	 return nil;
// 	 nsets = 0;
// 	 if (config.fonts[FcSetSystem])
// 	 sets[nsets++] = config.fonts[FcSetSystem];
// 	 if (config.fonts[FcSetApplication])
// 	 sets[nsets++] = config.fonts[FcSetApplication];
// 	 ret = Sort (config, sets, nsets, p, trim, csp, result);
// 	 FcConfigDestroy (config);

// 	 return ret;
//  }
