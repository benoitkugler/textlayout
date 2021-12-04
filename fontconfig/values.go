package fontconfig

import (
	"fmt"
	"io"
	"log"
	"strings"
)

var Identity = Matrix{1, 0, 0, 1}

// Object encode the properties of a font.
// Standard properties are built in the package,
// but custom ones may also be integrated inside patterns
// and configuration files.
type Object uint16

const (
	invalid         Object = iota
	FAMILY                 // with type String
	FAMILYLANG             // with type String
	STYLE                  // with type String
	STYLELANG              // with type String
	FULLNAME               // with type String
	FULLNAMELANG           // with type String
	SLANT                  // with type Int
	WEIGHT                 // with type Range
	WIDTH                  // with type Range
	SIZE                   // with type Range
	ASPECT                 // with type Double
	PIXEL_SIZE             // with type Double
	SPACING                // with type Int
	FOUNDRY                // with type String
	ANTIALIAS              // with type Bool
	HINT_STYLE             // with type Int
	HINTING                // with type Bool
	VERTICAL_LAYOUT        // with type Bool
	AUTOHINT               // with type Bool
	GLOBAL_ADVANCE         // with type Bool
	FILE                   // with type String
	INDEX                  // with type Int
	RASTERIZER             // with type String
	OUTLINE                // with type Bool
	SCALABLE               // with type Bool
	DPI                    // with type Double
	RGBA                   // with type Int
	SCALE                  // with type Double
	MINSPACE               // with type Bool
	CHARWIDTH              // with type Int
	CHAR_HEIGHT            // with type Int
	MATRIX                 // with type Matrix
	CHARSET                // with type CharSet
	LANG                   // with type LangSet
	FONTVERSION            // with type Int
	CAPABILITY             // with type String
	FONTFORMAT             // with type String
	EMBOLDEN               // with type Bool
	EMBEDDED_BITMAP        // with type Bool
	DECORATIVE             // with type Bool
	LCD_FILTER             // with type Int
	NAMELANG               // with type String
	FONT_FEATURES          // with type String
	PRGNAME                // with type String
	HASH                   // with type String
	POSTSCRIPT_NAME        // with type String
	COLOR                  // with type Bool
	SYMBOL                 // with type Bool
	FONT_VARIATIONS        // with type String
	VARIABLE               // with type Bool
	FONT_HAS_HINT          // with type Bool
	ORDER                  // with type Int
	// Custom objects should be defined starting from this value
	FirstCustomObject
)

// Bool is a tri-state boolean (see the associated constants)
type Bool uint8

const (
	False    Bool = iota // common `false`
	True                 // common `true`
	DontCare             // unspecified
)

func (b Bool) String() string {
	switch b {
	case False:
		return "false"
	case True:
		return "true"
	case DontCare:
		return "dont-care"
	default:
		return fmt.Sprintf("<Bool %d>", b)
	}
}

type Range struct {
	Begin, End float32
}

func rangePromote(v Float) Range {
	return Range{Begin: float32(v), End: float32(v)}
}

// returns true if a is inside b
func (a Range) isInRange(b Range) bool {
	return a.Begin >= b.Begin && a.End <= b.End
}

func rangeCompare(op opKind, a, b Range) bool {
	switch op {
	case opEqual:
		return a.Begin == b.Begin && a.End == b.End
	case opContains, opListing:
		return a.isInRange(b)
	case opNotEqual:
		return a.Begin != b.Begin || a.End != b.End
	case opNotContains:
		return !a.isInRange(b)
	case opLess:
		return a.End < b.Begin
	case opLessEqual:
		return a.End <= b.Begin
	case opMore:
		return a.Begin > b.End
	case opMoreEqual:
		return a.Begin >= b.End
	}
	return false
}

type Matrix struct {
	Xx, Xy, Yx, Yy float32
}

// return a * b
func (a Matrix) Multiply(b Matrix) Matrix {
	var r Matrix
	r.Xx = a.Xx*b.Xx + a.Xy*b.Yx
	r.Xy = a.Xx*b.Xy + a.Xy*b.Yy
	r.Yx = a.Yx*b.Xx + a.Yy*b.Yx
	r.Yy = a.Yx*b.Xy + a.Yy*b.Yy
	return r
}

// hasher may be implemented by complex value types,
// for which a custom hash is needed, beyong their string representation.
// The hash must entirely define the object: same hash means same values.
// See `Pattern.Hash` for more details.
type hasher interface {
	hash() []byte
}

// Value is a sum type for the values
// of the properties of a pattern
type Value interface {
	// Copy returns a deep copy of the value.
	copy() Value
	exprNode                      // usable as expression node
	serializeBin(io.Writer) error // exportable in custom binary format
}

func (v Int) copy() Value     { return v }
func (v Float) copy() Value   { return v }
func (v String) copy() Value  { return v }
func (v Bool) copy() Value    { return v }
func (v Charset) copy() Value { return v.Copy() }
func (v Langset) copy() Value { return v.Copy() }
func (v Matrix) copy() Value  { return v }
func (v Range) copy() Value   { return v }

func (v Int) copyExpr() exprNode     { return v.copy() }
func (v Float) copyExpr() exprNode   { return v.copy() }
func (v String) copyExpr() exprNode  { return v.copy() }
func (v Bool) copyExpr() exprNode    { return v.copy() }
func (v Charset) copyExpr() exprNode { return v.copy() }
func (v Langset) copyExpr() exprNode { return v.copy() }
func (v Matrix) copyExpr() exprNode  { return v.copy() }
func (v Range) copyExpr() exprNode   { return v.copy() }

type Int int32

type Float float32

type String string

// validate the basic data types
func (object Object) hasValidType(val Value) bool {
	_, isInt := val.(Int)
	_, isFloat := val.(Float)
	switch object {
	case FAMILY, FAMILYLANG, STYLE, STYLELANG, FULLNAME, FULLNAMELANG, FOUNDRY,
		RASTERIZER, CAPABILITY, NAMELANG, FONT_FEATURES, PRGNAME, HASH, POSTSCRIPT_NAME,
		FONTFORMAT, FILE, FONT_VARIATIONS: // string
		_, isString := val.(String)
		return isString
	case ORDER, SLANT, SPACING, HINT_STYLE, RGBA, INDEX,
		CHARWIDTH, LCD_FILTER, FONTVERSION, CHAR_HEIGHT: // Int
		return isInt
	case WEIGHT, WIDTH, SIZE: // range
		_, isRange := val.(Range)
		return isInt || isFloat || isRange
	case ASPECT, PIXEL_SIZE, SCALE, DPI: // float
		return isInt || isFloat
	case ANTIALIAS, HINTING, VERTICAL_LAYOUT, AUTOHINT, GLOBAL_ADVANCE, OUTLINE, SCALABLE,
		MINSPACE, EMBOLDEN, COLOR, SYMBOL, VARIABLE, FONT_HAS_HINT, EMBEDDED_BITMAP, DECORATIVE: // bool
		_, isBool := val.(Bool)
		return isBool
	case MATRIX: // Matrix
		_, isMatrix := val.(Matrix)
		return isMatrix
	case CHARSET: // CharSet
		_, isCharSet := val.(Charset)
		return isCharSet
	case LANG: // LangSet
		_, isLangSet := val.(Langset)
		_, isString := val.(String)
		return isLangSet || isString
	default:
		// no validation
		return true
	}
}

// Compares two values. Ints and Doubles are compared as numbers; otherwise
// the two values have to be the same type to be considered equal. Strings are
// compared ignoring case.
func valueEqual(va, vb Value) bool {
	if v, ok := va.(Int); ok {
		va = Float(v)
	}
	if v, ok := vb.(Int); ok {
		vb = Float(v)
	}

	switch va := va.(type) {
	case nil:
		return vb == nil
	case Float:
		if vb, ok := vb.(Float); ok {
			return va == vb
		}
	case String:
		if vb, ok := vb.(String); ok {
			return cmpIgnoreCase(string(va), string(vb)) == 0
		}
	case Bool:
		if vb, ok := vb.(Bool); ok {
			return va == vb
		}
	case Matrix:
		if vb, ok := vb.(Matrix); ok {
			return va == vb
		}
	case Charset:
		if vb, ok := vb.(Charset); ok {
			return charsetEqual(va, vb)
		}
	case Langset:
		if vb, ok := vb.(Langset); ok {
			return langsetEqual(va, vb)
		}
	case Range:
		if vb, ok := vb.(Range); ok {
			return va.isInRange(vb)
		}
	}
	return false
}

type valueElt struct {
	Value   Value
	Binding valueBinding
}

func (v valueElt) String() string {
	return fmt.Sprintf("%v (%s)", v.Value, v.Binding)
}

func (v valueElt) hash() []byte {
	if withHash, ok := v.Value.(hasher); ok {
		return withHash.hash()
	}
	return []byte(fmt.Sprintf("%v", v.Value))
}

func (v valueElt) asGoSource() string {
	return fmt.Sprintf("valueElt{Value: %s, Binding: %d}", v.Value.asGoSource(), v.Binding)
}

type valueBinding uint8

const (
	vbWeak valueBinding = iota
	vbStrong
	vbSame
)

func (b valueBinding) String() string {
	switch b {
	case vbWeak:
		return "w"
	case vbStrong:
		return "s"
	case vbSame:
		return "="
	default:
		return fmt.Sprintf("<%d>", b)
	}
}

type valueList []valueElt

func (vs valueList) Hash() []byte {
	var hash []byte
	for _, v := range vs {
		hash = append(hash, v.hash()...)
	}
	return hash
}

func (vs valueList) asGoSource() string {
	chunks := make([]string, len(vs))
	for i, v := range vs {
		chunks[i] = v.asGoSource()
	}
	return fmt.Sprintf("valueList{%s}", strings.Join(chunks, ","))
}

func (l valueList) prepend(v ...valueElt) valueList {
	l = append(l, make(valueList, len(v))...)
	copy(l[len(v):], l)
	copy(l, v)
	return l
}

// returns a deep copy
func (l *valueList) duplicate() *valueList {
	if l == nil {
		return nil
	}
	out := make(valueList, len(*l))
	for i, v := range *l {
		if v.Value != nil {
			v.Value = v.Value.copy()
		}
		out[i] = v
	}
	return &out
}

// insert `newList` into head, begining at `position`.
// If `appendMode` is true, `newList` is inserted just after `position`
// else, `newList` is inserted just before `position`.
// If position == -1, `newList` is inserted at the end or at the begining (depending on `appendMode`)
// `table` is updated for family objects.
// `newList` elements are also typecheked: false is returned if the types are invalid
func (head *valueList) insert(position int, appendMode bool, newList valueList,
	object Object, table *familyTable) bool {

	// Make sure the stored type is valid for built-in objects
	for _, l := range newList {
		if !object.hasValidType(l.Value) {
			log.Printf("fontconfig: pattern object %s does not accept value %v", object, l.Value)
			return false
		}
	}

	if object == FAMILY && table != nil {
		table.add(newList)
	}

	sameBinding := vbWeak
	if position != -1 {
		sameBinding = (*head)[position].Binding
	}

	for i, v := range newList {
		if v.Binding == vbSame {
			newList[i].Binding = sameBinding
		}
	}

	var cutoff int
	if appendMode {
		if position != -1 {
			cutoff = position + 1
		} else {
			cutoff = len(*head)
		}
	} else {
		if position != -1 {
			cutoff = position
		} else {
			cutoff = 0
		}
	}

	tmp := append(*head, make(valueList, len(newList))...) // allocate
	copy(tmp[cutoff+len(newList):], (*head)[cutoff:])      // make room for newList
	copy(tmp[cutoff:], newList)                            // insert newList
	*head = tmp

	return true
}

// remove the item at `position`
func (head *valueList) del(position int, object Object, table *familyTable) {
	if object == FAMILY && table != nil {
		table.del((*head)[position].Value.(String))
	}

	copy((*head)[position:], (*head)[position+1:])
	(*head)[len((*head))-1] = valueElt{}
	(*head) = (*head)[:len((*head))-1]
}
