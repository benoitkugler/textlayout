package fontconfig

import (
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"

	"github.com/benoitkugler/textlayout/fonts"
)

// An Pattern holds a set of names with associated value lists; each name refers to a
// property of a font, also called `Object`. Patterns are used as inputs to the matching code as
// well as holding information about specific fonts. Each property can hold
// one or more values; conventionally all of the same type, although the
// interface doesn't demand that.
type Pattern map[Object]*valueList

// NewPattern returns an empty, initalized pattern
func NewPattern() Pattern { return make(map[Object]*valueList) }

// Format returns the font format of this pattern.
func (p Pattern) Format() FontFormat {
	f, _ := p.GetString(FONTFORMAT)
	return FontFormat(f)
}

// FaceID reads the FILE and INDEX records.
// Note that it usually only makes sense to call this function
// for results (for which the fields are garanteed to be valid).
func (p Pattern) FaceID() fonts.FaceID {
	var out fonts.FaceID
	out.File, _ = p.GetString(FILE)
	index, _ := p.GetInt(INDEX)
	out.Index = uint16(uint32(index))
	out.Instance = uint16(uint32(index) >> 16)
	return out
}

// Duplicate returns a new pattern that matches
// `p`. Each pattern may be modified without affecting the other one.
func (p Pattern) Duplicate() Pattern {
	out := make(Pattern, len(p))
	for o, l := range p {
		out[o] = l.duplicate()
	}
	return out
}

// Add adds the given value for the given object, with a strong binding.
// `appendMode` controls the location of insertion in the current list.
// A copy of `value` is made, so that the library
// retains no reference to any user-supplied data.
func (p Pattern) Add(object Object, value Value, appendMode bool) {
	p.addWithBinding(object, value, vbStrong, appendMode)
}

func (p Pattern) addWithBinding(object Object, value Value, binding valueBinding, appendMode bool) {
	newV := valueElt{Value: value, Binding: binding}
	p.addList(object, valueList{newV}, appendMode)
}

// AddBool append `value` to `object` field
func (p Pattern) AddBool(object Object, value bool) {
	var fBool Bool
	if value {
		fBool = True
	}
	p.Add(object, fBool, true)
}

// AddInt append `value` to `object` field
func (p Pattern) AddInt(object Object, value int32) {
	p.Add(object, Int(value), true)
}

// AddFloat append `value` to `object` field
func (p Pattern) AddFloat(object Object, value float32) {
	p.Add(object, Float(value), true)
}

// AddString append `value` to `object` field
func (p Pattern) AddString(object Object, value string) {
	p.Add(object, String(value), true)
}

// Add adds the given list of values for the given object.
// `appendMode` controls the location of insertion in the current list.
func (p Pattern) addList(object Object, list valueList, appendMode bool) {
	// Make sure the stored type is valid for built-in objects
	for _, value := range list {
		if !object.hasValidType(value.Value) {
			log.Printf("fontconfig: pattern object %s does not accept value %v", object, value.Value)
			return
		}
	}

	e := p[object]
	if e == nil {
		e = &valueList{}
		p[object] = e
	}
	if appendMode {
		*e = append(*e, *list.duplicate()...)
	} else {
		*e = e.prepend(*list.duplicate()...)
	}
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
	var (
		hash []byte
		buf  [6]byte
	)
	for _, object := range p.sortedKeys() {
		v := p[object]
		if v == nil || len(*v) == 0 {
			continue
		}
		// to make the hash unique, we add the length of each object value
		tmp := v.Hash()
		binary.BigEndian.PutUint16(buf[:], uint16(object))
		binary.BigEndian.PutUint32(buf[2:], uint32(len(tmp)))
		hash = append(append(hash, buf[:]...), tmp...)
	}
	return string(hash)
}

// String returns a human friendly representation,
// mainly used for debugging.
func (p Pattern) String() string {
	if p == nil {
		return "<nil>"
	}
	s := fmt.Sprintf("Pattern(len %d):\n", len(p))

	for obj, vs := range p {
		s += fmt.Sprintf("\t%s: %v\n", obj, vs)
	}

	return s
}

// resolve the nil values
func (p Pattern) getVals(obj Object) valueList {
	out := p[obj]
	if out == nil {
		return nil
	}
	return *out
}

// GetAt returns the value in position `id` for `object`, without type conversion.
func (p Pattern) GetAt(object Object, id int) (Value, Result) {
	e := p[object]
	if e == nil {
		return nil, ResultNoMatch
	}
	if id >= len(*e) {
		return nil, ResultNoId
	}
	return (*e)[id].Value, ResultMatch
}

// GetBool return the potential Bool at `object`, index 0, if any.
func (p Pattern) GetBool(object Object) (Bool, bool) {
	v, r := p.GetAt(object, 0)
	if r != ResultMatch {
		return 0, false
	}
	out, ok := v.(Bool)
	return out, ok
}

// GetString return the potential string at `object`, index 0, if any.
func (p Pattern) GetString(object Object) (string, bool) {
	v, r := p.getAtString(object, 0)
	return v, r == ResultMatch
}

// GetStrings returns the values with type String at `object`
func (p Pattern) GetStrings(object Object) []string {
	var out []string
	for _, v := range p.getVals(object) {
		m, ok := v.Value.(String)
		if ok {
			out = append(out, string(m))
		}
	}
	return out
}

func (p Pattern) getAtString(object Object, id int) (string, Result) {
	v, r := p.GetAt(object, id)
	if r != ResultMatch {
		return "", r
	}
	out, ok := v.(String)
	if !ok {
		return "", ResultTypeMismatch
	}
	return string(out), ResultMatch
}

// GetCharset return the potential Charset at `object`, index 0, if any.
func (p Pattern) GetCharset(object Object) (Charset, bool) {
	v, r := p.GetAt(object, 0)
	if r != ResultMatch {
		return Charset{}, false
	}
	out, ok := v.(Charset)
	return out, ok
}

// GetFloat return the potential first float at `object`, if any.
// Note that integers are also accepted.
// See GetAt for more control on type conversion
func (p Pattern) GetFloat(object Object) (float32, bool) {
	v, r := p.GetAt(object, 0)
	if r != ResultMatch {
		return 0, false
	}
	if out, ok := v.(Float); ok {
		return float32(out), true
	} else if out, ok := v.(Int); ok {
		return float32(out), true
	}
	return 0, false
}

// GetFloats returns the values with type Float at `object`
func (p Pattern) GetFloats(object Object) []float32 {
	var out []float32

	for _, v := range p.getVals(object) {
		m, ok := v.Value.(Float)
		if ok {
			out = append(out, float32(m))
		}
	}
	return out
}

// GetInt return the potential first int at `object`, if any.
func (p Pattern) GetInt(object Object) (int32, bool) {
	v, r := p.GetAt(object, 0)
	if r != ResultMatch {
		return 0, false
	}
	out, ok := v.(Int)
	return int32(out), ok
}

// GetInts returns the values with type Int at `object`
func (p Pattern) GetInts(object Object) []int32 {
	var out []int32
	for _, v := range p.getVals(object) {
		m, ok := v.Value.(Int)
		if ok {
			out = append(out, int32(m))
		}
	}
	return out
}

// GetMatrix return the potential Matrix at `object`, index 0, if any.
func (p Pattern) GetMatrix(object Object) (Matrix, bool) {
	v, r := p.GetAt(object, 0)
	if r != ResultMatch {
		return Matrix{}, false
	}
	out, ok := v.(Matrix)
	return out, ok
}

// GetMatrices returns the values with type FcMatrix at `object`
func (p Pattern) GetMatrices(object Object) []Matrix {
	var out []Matrix
	for _, v := range p.getVals(object) {
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
		for _, v := range *list {
			p.addWithBinding(object, v.Value, v.Binding, true)
		}
	}
}

func (pat Pattern) addFullname() bool {
	if b, _ := pat.GetBool(VARIABLE); b != False {
		return true
	}

	var (
		style string
		n     int
	)
	lang, res := pat.getAtString(FAMILYLANG, n)
	for ; res == ResultMatch; lang, res = pat.getAtString(FAMILYLANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	family, res := pat.getAtString(FAMILY, n)
	if res != ResultMatch {
		return false
	}
	family = strings.TrimRightFunc(family, unicode.IsSpace)
	lang = ""
	lang, res = pat.getAtString(STYLELANG, n)
	for ; res == ResultMatch; lang, res = pat.getAtString(STYLELANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	style, res = pat.getAtString(STYLE, n)
	if res != ResultMatch {
		return false
	}

	style = strings.TrimLeftFunc(style, unicode.IsSpace)
	sbuf := []byte(family)
	if cmpIgnoreBlanksAndCase(style, "Regular") != 0 {
		sbuf = append(sbuf, ' ')
		sbuf = append(sbuf, style...)
	}
	pat.Del(FULLNAME)
	pat.Add(FULLNAME, String(sbuf), true)
	pat.Del(FULLNAMELANG)
	pat.Add(FULLNAMELANG, String("en"), true)

	return true
}

// PatternElement associate one value to a categorie (or Object)
type PatternElement struct {
	Value  Value
	Object Object
}

func BuildPattern(elements ...PatternElement) Pattern {
	p := make(Pattern, len(elements))
	for _, el := range elements {
		p.Add(el.Object, el.Value, true)
	}
	return p
}

func (p Pattern) addWithTable(object Object, list valueList, append bool, table *familyTable) {
	e := p[object]
	if e == nil {
		e = &valueList{}
		p[object] = e
	}
	e.insert(-1, append, list, object, table)
}

// Delete all values associated with a field
func (p Pattern) delWithTable(object Object, table *familyTable) {
	e := p.getVals(object)

	if object == FAMILY && table != nil {
		for _, v := range e {
			table.del(v.Value.(String))
		}
	}

	delete(p, object)
}

// remove the nil and empty lists
func (p Pattern) canon(object Object) {
	e := p[object]
	if e == nil || len(*e) == 0 {
		delete(p, object)
	}
}

var boolDefaults = [...]struct {
	field Object
	value bool
}{
	{HINTING, true},          /* !FT_LOAD_NO_HINTING */
	{VERTICAL_LAYOUT, false}, /* LOAD_VERTICAL_LAYOUT */
	{AUTOHINT, false},        /* LOAD_FORCE_AUTOHINT */
	{GLOBAL_ADVANCE, true},   /* !LOAD_IGNORE_GLOBAL_ADVANCE_WIDTH */
	{EMBEDDED_BITMAP, true},  /* !LOAD_NO_BITMAP */
	{DECORATIVE, false},
	{SYMBOL, false},
	{VARIABLE, false},
}

// SubstituteDefault performs default substitutions in a pattern,
// supplying default values for underspecified font patterns:
// 	- unspecified style or weight are set to Medium
// 	- unspecified style or slant are set to Roman
// 	- unspecified pixel size are given one computed from any
// 		specified point size (default 12), dpi (default 75) and scale (default 1).
func (pattern Pattern) SubstituteDefault() {
	if pattern[WEIGHT] == nil {
		pattern.AddInt(WEIGHT, WEIGHT_NORMAL)
	}

	if pattern[SLANT] == nil {
		pattern.AddInt(SLANT, SLANT_ROMAN)
	}

	if pattern[WIDTH] == nil {
		pattern.AddInt(WIDTH, WIDTH_NORMAL)
	}

	for _, boolDef := range boolDefaults {
		if pattern[boolDef.field] == nil {
			pattern.AddBool(boolDef.field, boolDef.value)
		}
	}

	size := float32(12.0)
	sizeObj, _ := pattern.GetAt(SIZE, 0)
	switch sizeObj := sizeObj.(type) {
	case Int:
		size = float32(sizeObj)
	case Float:
		size = float32(sizeObj)
	case Range:
		size = (sizeObj.Begin + sizeObj.End) * .5
	}

	scale, ok := pattern.GetFloat(SCALE)
	if !ok {
		scale = 1.0
	}

	dpi, ok := pattern.GetFloat(DPI)
	if !ok {
		dpi = 75.0
	}

	if pixelSize := pattern.getVals(PIXEL_SIZE); len(pixelSize) == 0 {
		pattern.Del(SCALE)
		pattern.AddFloat(SCALE, scale)
		pixelsize := float32(size) * scale
		pattern.Del(DPI)
		pattern.AddFloat(DPI, dpi)
		pixelsize *= dpi / 72.0
		pattern.AddFloat(PIXEL_SIZE, pixelsize)
	} else {
		sizeF, _ := pixelSize[0].Value.(Float)
		size = float32(sizeF) / dpi * 72.0 / scale
	}
	pattern.Del(SIZE)
	pattern.AddFloat(SIZE, size)

	if pattern[FONTVERSION] == nil {
		pattern.AddInt(FONTVERSION, 0x7fffffff)
	}

	if pattern[HINT_STYLE] == nil {
		pattern.AddInt(HINT_STYLE, HINT_FULL)
	}

	if pattern[NAMELANG] == nil {
		pattern.AddString(NAMELANG, getDefaultLang())
	}

	/* shouldn't be failed. */
	namelang, _ := pattern.GetAt(NAMELANG, 0)

	/* Add a fallback to ensure the english name when the requested language
	 * isn't available. this would helps for the fonts that have non-English
	 * name at the beginning.
	 */
	/* Set "en-us" instead of "en" to avoid giving higher score to "en".
	 * This is a hack for the case that the orth is not like ll-cc, because,
	 * if no namelang isn't explicitly set, it will has something like ll-cc
	 * according to current locale. which may causes langDifferentTerritory
	 * at langCompare(). thus, the English name is selected so that
	 * exact matched "en" has higher score than ll-cc.
	 */
	lang := String("en-us")
	if pattern[FAMILYLANG] == nil {
		pattern.Add(FAMILYLANG, namelang, true)
		pattern.addWithBinding(FAMILYLANG, lang, vbWeak, true)
	}
	if pattern[STYLELANG] == nil {
		pattern.Add(STYLELANG, namelang, true)
		pattern.addWithBinding(STYLELANG, lang, vbWeak, true)
	}
	if pattern[FULLNAMELANG] == nil {
		pattern.Add(FULLNAMELANG, namelang, true)
		pattern.addWithBinding(FULLNAMELANG, lang, vbWeak, true)
	}

	if pattern[PRGNAME] == nil {
		if prgname := getProgramName(); prgname != "" {
			pattern.AddString(PRGNAME, prgname)
		}
	}

	if pattern[ORDER] == nil {
		pattern.AddInt(ORDER, 0)
	}
}
