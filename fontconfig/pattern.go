package fontconfig

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"
)

// An Pattern holds a set of names with associated value lists; each name refers to a
// property of a font, also called `Object`. Patterns are used as inputs to the matching code as
// well as holding information about specific fonts. Each property can hold
// one or more values; conventionally all of the same type, although the
// interface doesn't demand that.
type Pattern map[Object]valueList

// NewPattern returns an empty, initalized pattern
func NewPattern() Pattern { return make(map[Object]valueList) }

// Duplicate returns a new pattern that matches
// `p`. Each pattern may be modified without affecting the other.
func (p Pattern) Duplicate() Pattern {
	out := make(Pattern, len(p))
	for o, l := range p {
		out[o] = l.duplicate()
	}
	return out
}

// Add adds the given value for the given object, with a strong binding.
// `appendMode` controls the location of insertion in the current list.
func (p Pattern) Add(object Object, value Value, appendMode bool) {
	p.addWithBinding(object, value, FcValueBindingStrong, appendMode)
}

func (p Pattern) addWithBinding(object Object, value Value, binding FcValueBinding, appendMode bool) {
	newV := valueElt{Value: value, Binding: binding}
	p.AddList(object, valueList{newV}, appendMode)
}

func (p Pattern) AddBool(object Object, value bool) {
	var fBool Bool
	if value {
		fBool = 1
	}
	p.Add(object, fBool, true)
}

func (p Pattern) AddInteger(object Object, value int) {
	p.Add(object, Int(value), true)
}
func (p Pattern) AddFloat(object Object, value float64) {
	p.Add(object, Float(value), true)
}
func (p Pattern) AddString(object Object, value string) {
	p.Add(object, String(value), true)
}

// Add adds the given list of values for the given object.
// `appendMode` controls the location of insertion in the current list.
func (p Pattern) AddList(object Object, list valueList, appendMode bool) {
	//  FcPatternObjectAddWithBinding(p, FcObjectFromName(object), value, FcValueBindingStrong, append)
	// object := FcObject(objectS)

	// Make sure the stored type is valid for built-in objects
	for _, value := range list {
		if !object.hasValidType(value.Value) {
			log.Printf("fontconfig: pattern object %s does not accept value %v", object, value.Value)
			return
		}
	}

	e := p[object]
	if appendMode {
		e = append(e, list...)
	} else {
		e = e.prepend(list...)
	}
	p[object] = e
}

// Del remove all the values associated to `object`
func (p Pattern) Del(object Object) { delete(p, object) }

func (p Pattern) sortedKeys() []Object {
	keys := make([]Object, 0, len(p))
	for r := range p {
		keys = append(keys, r)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// Hash returns a value, usable as map key, and
// defining the pattern in terms of equality:
// two patterns with the same hash are considered equal.
func (p Pattern) Hash() string {
	var hash []byte
	for _, object := range p.sortedKeys() {
		v := p[object]
		hash = append(append(hash, byte(object), ':'), v.Hash()...)
	}
	return string(hash)
}

// String returns a human friendly representation,
// mainly used for debugging.
func (p Pattern) String() string {
	s := fmt.Sprintf("%d elements pattern:\n", len(p))

	for obj, vs := range p {
		s += fmt.Sprintf("\t%s: %v\n", obj, vs)
	}
	return s
}

// GetAt returns the value in position `id` for `object`, without type conversion.
func (p Pattern) GetAt(object Object, id int) (Value, FcResult) {
	e := p[object]
	if e == nil {
		return nil, FcResultNoMatch
	}
	if id >= len(e) {
		return nil, FcResultNoId
	}
	return e[id].Value, FcResultMatch
}

// GetBool return the potential Bool at `object`, index 0, if any.
func (p Pattern) GetBool(object Object) (Bool, bool) {
	v, r := p.GetAt(object, 0)
	if r != FcResultMatch {
		return 0, false
	}
	out, ok := v.(Bool)
	return out, ok
}

// GetString return the potential string at `object`, index 0, if any.
func (p Pattern) GetString(object Object) (string, bool) {
	v, r := p.GetAt(object, 0)
	if r != FcResultMatch {
		return "", false
	}
	out, ok := v.(String)
	return string(out), ok
}

func (p Pattern) GetAtString(object Object, id int) (string, FcResult) {
	v, r := p.GetAt(object, id)
	if r != FcResultMatch {
		return "", r
	}
	out, ok := v.(String)
	if !ok {
		return "", FcResultTypeMismatch
	}
	return string(out), FcResultMatch
}

// GetCharset return the potential Charset at `object`, index 0, if any.
func (p Pattern) GetCharset(object Object) (Charset, bool) {
	v, r := p.GetAt(object, 0)
	if r != FcResultMatch {
		return Charset{}, false
	}
	out, ok := v.(Charset)
	return out, ok
}

// GetFloat return the potential first float at `object`, if any.
func (p Pattern) GetFloat(object Object) (float64, bool) {
	v, r := p.GetAt(object, 0)
	if r != FcResultMatch {
		return 0, false
	}
	out, ok := v.(Float)
	return float64(out), ok
}

// GetFloats returns the values with type Float at `object`
func (p Pattern) GetFloats(object Object) []float64 {
	var out []float64
	for _, v := range p[object] {
		m, ok := v.Value.(Float)
		if ok {
			out = append(out, float64(m))
		}
	}
	return out
}

// GetInt return the potential first int at `object`, if any.
func (p Pattern) GetInt(object Object) (int, bool) {
	v, r := p.GetAt(object, 0)
	if r != FcResultMatch {
		return 0, false
	}
	out, ok := v.(Int)
	return int(out), ok
}

// GetInts returns the values with type Int at `object`
func (p Pattern) GetInts(object Object) []int {
	var out []int
	for _, v := range p[object] {
		m, ok := v.Value.(Int)
		if ok {
			out = append(out, int(m))
		}
	}
	return out
}

// GetMatrix return the potential Matrix at `object`, index 0, if any.
func (p Pattern) GetMatrix(object Object) (Matrix, bool) {
	v, r := p.GetAt(object, 0)
	if r != FcResultMatch {
		return Matrix{}, false
	}
	out, ok := v.(Matrix)
	return out, ok
}

// GetMatrices returns the values with type FcMatrix at `object`
func (p Pattern) GetMatrices(object Object) []Matrix {
	var out []Matrix
	for _, v := range p[object] {
		m, ok := v.Value.(Matrix)
		if ok {
			out = append(out, m)
		}
	}
	return out
}

// Add all of the elements in 's' to 'p'
func (p Pattern) append(s Pattern) {
	for object, list := range s {
		for _, v := range list {
			p.addWithBinding(object, v.Value, v.Binding, true)
		}
	}
}

func (pat Pattern) addFullname() bool {
	if b, _ := pat.GetBool(FC_VARIABLE); b != FcFalse {
		return true
	}

	var (
		style string
		n     int
	)
	lang, res := pat.GetAtString(FC_FAMILYLANG, n)
	for ; res == FcResultMatch; lang, res = pat.GetAtString(FC_FAMILYLANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	family, res := pat.GetAtString(FC_FAMILY, n)
	if res != FcResultMatch {
		return false
	}
	family = strings.TrimRightFunc(family, unicode.IsSpace)
	lang = ""
	lang, res = pat.GetAtString(FC_STYLELANG, n)
	for ; res == FcResultMatch; lang, res = pat.GetAtString(FC_STYLELANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	style, res = pat.GetAtString(FC_STYLE, n)
	if res != FcResultMatch {
		return false
	}

	style = strings.TrimLeftFunc(style, unicode.IsSpace)
	sbuf := []byte(family)
	if cmpIgnoreBlanksAndCase(style, "Regular") != 0 {
		sbuf = append(sbuf, ' ')
		sbuf = append(sbuf, style...)
	}
	pat.Del(FC_FULLNAME)
	pat.Add(FC_FULLNAME, String(sbuf), true)
	pat.Del(FC_FULLNAMELANG)
	pat.Add(FC_FULLNAMELANG, String("en"), true)

	return true
}

type PatternElement struct {
	Object Object
	Value  Value
}

// TODO: check the pointer types in values
func BuildPattern(elements ...PatternElement) Pattern {
	p := make(Pattern, len(elements))
	for _, el := range elements {
		p.Add(el.Object, el.Value, true)
	}
	return p
}

func (p Pattern) addWithTable(object Object, list valueList, append bool, table *familyTable) {
	e := p[object]
	e.insert(-1, append, list, object, table)
	p[object] = e
}

// Delete all values associated with a field
func (p Pattern) FcConfigPatternDel(object Object, table *familyTable) {
	e := p[object]

	if object == FC_FAMILY && table != nil {
		for _, v := range e {
			table.del(v.Value.(String))
		}
	}

	delete(p, object)
}

// remove the empty lists
func (p Pattern) canon(object Object) {
	e := p[object]
	if len(e) == 0 {
		delete(p, object)
	}
}

var boolDefaults = [...]struct {
	field Object
	value bool
}{
	{FC_HINTING, true},          /* !FT_LOAD_NO_HINTING */
	{FC_VERTICAL_LAYOUT, false}, /* FC_LOAD_VERTICAL_LAYOUT */
	{FC_AUTOHINT, false},        /* FC_LOAD_FORCE_AUTOHINT */
	{FC_GLOBAL_ADVANCE, true},   /* !FC_LOAD_IGNORE_GLOBAL_ADVANCE_WIDTH */
	{FC_EMBEDDED_BITMAP, true},  /* !FC_LOAD_NO_BITMAP */
	{FC_DECORATIVE, false},
	{FC_SYMBOL, false},
	{FC_VARIABLE, false},
}

// SubstituteDefault performs default substitutions in a pattern,
// supplying default values for underspecified font patterns:
// 	- unspecified style or weight are set to Medium
// 	- unspecified style or slant are set to Roman
// 	- unspecified pixel size are given one computed from any
// 		specified point size (default 12), dpi (default 75) and scale (default 1).
func (pattern Pattern) SubstituteDefault() {
	if pattern[FC_WEIGHT] == nil {
		pattern.AddInteger(FC_WEIGHT, WEIGHT_NORMAL)
	}

	if pattern[FC_SLANT] == nil {
		pattern.AddInteger(FC_SLANT, SLANT_ROMAN)
	}

	if pattern[FC_WIDTH] == nil {
		pattern.AddInteger(FC_WIDTH, WIDTH_NORMAL)
	}

	for _, boolDef := range boolDefaults {
		if pattern[boolDef.field] == nil {
			pattern.AddBool(boolDef.field, boolDef.value)
		}
	}

	size := 12.0
	sizeObj, _ := pattern.GetAt(FC_SIZE, 0)
	switch sizeObj := sizeObj.(type) {
	case Float:
		size = float64(sizeObj)
	case Range:
		size = (sizeObj.Begin + sizeObj.End) * .5
	}

	scale, ok := pattern.GetFloat(FC_SCALE)
	if !ok {
		scale = 1.0
	}

	dpi, ok := pattern.GetFloat(FC_DPI)
	if !ok {
		dpi = 75.0
	}

	if pixelSize := pattern[FC_PIXEL_SIZE]; len(pixelSize) == 0 {
		pattern.Del(FC_SCALE)
		pattern.AddFloat(FC_SCALE, scale)
		pixelsize := float64(size) * scale
		pattern.Del(FC_DPI)
		pattern.AddFloat(FC_DPI, dpi)
		pixelsize *= dpi / 72.0
		pattern.AddFloat(FC_PIXEL_SIZE, pixelsize)
	} else {
		sizeF, _ := pixelSize[0].Value.(Float)
		size = float64(sizeF) / dpi * 72.0 / scale
	}
	pattern.Del(FC_SIZE)
	pattern.AddFloat(FC_SIZE, size)

	if pattern[FC_FONTVERSION] == nil {
		pattern.AddInteger(FC_FONTVERSION, 0x7fffffff)
	}

	if pattern[FC_HINT_STYLE] == nil {
		pattern.AddInteger(FC_HINT_STYLE, FC_HINT_FULL)
	}

	if pattern[FC_NAMELANG] == nil {
		pattern.AddString(FC_NAMELANG, getDefaultLang())
	}

	/* shouldn't be failed. */
	namelang, _ := pattern.GetAt(FC_NAMELANG, 0)

	/* Add a fallback to ensure the english name when the requested language
	 * isn't available. this would helps for the fonts that have non-English
	 * name at the beginning.
	 */
	/* Set "en-us" instead of "en" to avoid giving higher score to "en".
	 * This is a hack for the case that the orth is not like ll-cc, because,
	 * if no namelang isn't explicitly set, it will has something like ll-cc
	 * according to current locale. which may causes FcLangDifferentTerritory
	 * at FcLangCompare(). thus, the English name is selected so that
	 * exact matched "en" has higher score than ll-cc.
	 */
	lang := String("en-us")
	if pattern[FC_FAMILYLANG] == nil {
		pattern.Add(FC_FAMILYLANG, namelang, true)
		pattern.addWithBinding(FC_FAMILYLANG, lang, FcValueBindingWeak, true)
	}
	if pattern[FC_STYLELANG] == nil {
		pattern.Add(FC_STYLELANG, namelang, true)
		pattern.addWithBinding(FC_STYLELANG, lang, FcValueBindingWeak, true)
	}
	if pattern[FC_FULLNAMELANG] == nil {
		pattern.Add(FC_FULLNAMELANG, namelang, true)
		pattern.addWithBinding(FC_FULLNAMELANG, lang, FcValueBindingWeak, true)
	}

	if pattern[FC_PRGNAME] == nil {
		if prgname := getProgramName(); prgname != "" {
			pattern.AddString(FC_PRGNAME, prgname)
		}
	}

	if pattern[FC_ORDER] == nil {
		pattern.AddInteger(FC_ORDER, 0)
	}
}
