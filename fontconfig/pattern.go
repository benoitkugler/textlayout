package fontconfig

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"
)

// An Pattern holds a set of names with associated value lists; each name refers to a
// property of a font. FcPatterns are used as inputs to the matching code as
// well as holding information about specific fonts. Each property can hold
// one or more values; conventionally all of the same type, although the
// interface doesn't demand that.
type Pattern map[FcObject]FcValueList

// NewFcPattern returns an empty, initalized pattern
func NewFcPattern() Pattern {
	return make(map[FcObject]FcValueList)
}

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
func (p Pattern) Add(object FcObject, value FcValue, appendMode bool) {
	p.addWithBinding(object, value, FcValueBindingStrong, appendMode)
}

func (p Pattern) addWithBinding(object FcObject, value FcValue, binding FcValueBinding, appendMode bool) {
	newV := valueElt{value: value, binding: binding}
	p.AddList(object, FcValueList{newV}, appendMode)
}

func (p Pattern) FcPatternObjectAddBool(object FcObject, value bool) {
	var fBool FcBool
	if value {
		fBool = 1
	}
	p.Add(object, fBool, true)
}

func (p Pattern) FcPatternObjectAddInteger(object FcObject, value int) {
	p.Add(object, Int(value), true)
}
func (p Pattern) FcPatternObjectAddDouble(object FcObject, value float64) {
	p.Add(object, Float(value), true)
}
func (p Pattern) FcPatternObjectAddString(object FcObject, value string) {
	p.Add(object, String(value), true)
}

// Add adds the given list of values for the given object.
// `appendMode` controls the location of insertion in the current list.
func (p Pattern) AddList(object FcObject, list FcValueList, appendMode bool) {
	//  FcPatternObjectAddWithBinding(p, FcObjectFromName(object), value, FcValueBindingStrong, append)
	// object := FcObject(objectS)

	// Make sure the stored type is valid for built-in objects
	for _, value := range list {
		if !object.hasValidType(value.value) {
			log.Printf("fontconfig: pattern object %s does not accept value %v", object, value.value)
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
func (p Pattern) Del(object FcObject) { delete(p, object) }

func (p Pattern) sortedKeys() []FcObject {
	keys := make([]FcObject, 0, len(p))
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

func (p Pattern) FcPatternObjectGet(object FcObject, id int) (FcValue, FcResult) {
	e := p[object]
	if e == nil {
		return nil, FcResultNoMatch
	}
	if id >= len(e) {
		return nil, FcResultNoId
	}
	return e[id].value, FcResultMatch
}

func (p Pattern) FcPatternObjectGetBool(object FcObject, id int) (FcBool, FcResult) {
	v, r := p.FcPatternObjectGet(object, id)
	if r != FcResultMatch {
		return 0, r
	}
	out, ok := v.(FcBool)
	if !ok {
		return 0, FcResultTypeMismatch
	}
	return out, FcResultMatch
}

// GetString return the potential string at `object`, index 0, if any.
func (p Pattern) GetString(object FcObject) (string, bool) {
	v, r := p.FcPatternObjectGet(object, 0)
	if r != FcResultMatch {
		return "", false
	}
	out, ok := v.(String)
	return string(out), ok
}

func (p Pattern) FcPatternObjectGetString(object FcObject, id int) (string, FcResult) {
	v, r := p.FcPatternObjectGet(object, id)
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
func (p Pattern) GetCharset(object FcObject) (Charset, bool) {
	v, r := p.FcPatternObjectGet(object, 0)
	if r != FcResultMatch {
		return Charset{}, false
	}
	out, ok := v.(Charset)
	return out, ok
}

// GetFloat return the potential first float at `object`, if any.
func (p Pattern) GetFloat(object FcObject) (float64, bool) {
	v, r := p.FcPatternObjectGet(object, 0)
	if r != FcResultMatch {
		return 0, false
	}
	out, ok := v.(Float)
	return float64(out), ok
}

// GetFloats returns the values with type Float at `object`
func (p Pattern) GetFloats(object FcObject) []float64 {
	var out []float64
	for _, v := range p[object] {
		m, ok := v.value.(Float)
		if ok {
			out = append(out, float64(m))
		}
	}
	return out
}

// GetInt return the potential first int at `object`, if any.
func (p Pattern) GetInt(object FcObject) (int, bool) {
	v, r := p.FcPatternObjectGet(object, 0)
	if r != FcResultMatch {
		return 0, false
	}
	out, ok := v.(Int)
	return int(out), ok
}

// GetInts returns the values with type Int at `object`
func (p Pattern) GetInts(object FcObject) []int {
	var out []int
	for _, v := range p[object] {
		m, ok := v.value.(Int)
		if ok {
			out = append(out, int(m))
		}
	}
	return out
}

// GetMatrices returns the values with type FcMatrix at `object`
func (p Pattern) GetMatrices(object FcObject) []FcMatrix {
	var out []FcMatrix
	for _, v := range p[object] {
		m, ok := v.value.(FcMatrix)
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
			p.addWithBinding(object, v.value, v.binding, true)
		}
	}
}

func (pat Pattern) addFullname() bool {
	b, res := pat.FcPatternObjectGetBool(FC_VARIABLE, 0)
	if res == FcResultMatch && b != FcFalse {
		return true
	}

	var (
		lang, style string
		n           int
	)
	lang, res = pat.FcPatternObjectGetString(FC_FAMILYLANG, n)
	for ; res == FcResultMatch; lang, res = pat.FcPatternObjectGetString(FC_FAMILYLANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	family, res := pat.FcPatternObjectGetString(FC_FAMILY, n)
	if res != FcResultMatch {
		return false
	}
	family = strings.TrimRightFunc(family, unicode.IsSpace)
	lang = ""
	lang, res = pat.FcPatternObjectGetString(FC_STYLELANG, n)
	for ; res == FcResultMatch; lang, res = pat.FcPatternObjectGetString(FC_STYLELANG, n) {
		if lang == "en" {
			break
		}
		n++
		lang = ""
	}
	if lang == "" {
		n = 0
	}
	style, res = pat.FcPatternObjectGetString(FC_STYLE, n)
	if res != FcResultMatch {
		return false
	}

	style = strings.TrimLeftFunc(style, unicode.IsSpace)
	sbuf := []byte(family)
	if FcStrCmpIgnoreBlanksAndCase(style, "Regular") != 0 {
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
	Object FcObject
	Value  FcValue
}

// TODO: check the pointer types in values
func FcPatternBuild(elements ...PatternElement) Pattern {
	p := make(Pattern, len(elements))
	for _, el := range elements {
		p.Add(el.Object, el.Value, true)
	}
	return p
}

func (p Pattern) FcConfigPatternAdd(object FcObject, list FcValueList, append bool, table *FamilyTable) {
	e := p[object]
	e.insert(-1, append, list, object, table)
	p[object] = e
}

// Delete all values associated with a field
func (p Pattern) FcConfigPatternDel(object FcObject, table *FamilyTable) {
	e := p[object]

	if object == FC_FAMILY && table != nil {
		for _, v := range e {
			table.del(v.value.(String))
		}
	}

	delete(p, object)
}

// remove the empty lists
func (p Pattern) canon(object FcObject) {
	e := p[object]
	if len(e) == 0 {
		delete(p, object)
	}
}
