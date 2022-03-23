package fontconfig

import (
	"fmt"
	"log"
	"math"
)

type opKind uint64 // a flag might be added, see `withFlags` and `getOp`

const (
	opInt opKind = iota
	opDouble
	opString
	opMatrix
	opRange
	opBool
	opCharSet
	opLangSet
	opNil
	opField
	opConst
	opAssign
	opAssignReplace
	opPrependFirst
	opPrepend
	opAppend
	opAppendLast
	opDelete
	opDeleteAll
	opQuest
	opOr
	opAnd
	opEqual
	opNotEqual
	opContains
	opListing
	opNotContains
	opLess
	opLessEqual
	opMore
	opMoreEqual
	opPlus
	opMinus
	opTimes
	opDivide
	opNot
	opComma
	opFloor
	opCeil
	opRound
	opTrunc
	opInvalid
)

func opWithFlags(x opKind, f int64) opKind {
	return x | opKind(f)<<16
}

func (x opKind) getOp() opKind {
	return x & 0xffff
}

func (x opKind) getFlags() int64 {
	return (int64(x) & 0xffff0000) >> 16
}

func (x opKind) String() string {
	flagsString := ""
	if x.getFlags()&opFlagIgnoreBlanks != 0 {
		flagsString = " (ignore blanks)"
	}
	switch x.getOp() {
	case opInt:
		return "Int"
	case opDouble:
		return "Double"
	case opString:
		return "String"
	case opMatrix:
		return "Matrix"
	case opRange:
		return "Range"
	case opBool:
		return "Bool"
	case opCharSet:
		return "CharSet"
	case opLangSet:
		return "LangSet"
	case opField:
		return "Field"
	case opConst:
		return "Const"
	case opAssign:
		return "Assign"
	case opAssignReplace:
		return "AssignReplace"
	case opPrepend:
		return "Prepend"
	case opPrependFirst:
		return "PrependFirst"
	case opAppend:
		return "Append"
	case opAppendLast:
		return "AppendLast"
	case opDelete:
		return "Delete"
	case opDeleteAll:
		return "DeleteAll"
	case opQuest:
		return "Quest"
	case opOr:
		return "Or"
	case opAnd:
		return "And"
	case opEqual:
		return "Equal" + flagsString
	case opNotEqual:
		return "NotEqual" + flagsString
	case opLess:
		return "Less"
	case opLessEqual:
		return "LessEqual"
	case opMore:
		return "More"
	case opMoreEqual:
		return "MoreEqual"
	case opContains:
		return "Contains"
	case opNotContains:
		return "NotContains"
	case opPlus:
		return "Plus"
	case opMinus:
		return "Minus"
	case opTimes:
		return "Times"
	case opDivide:
		return "Divide"
	case opNot:
		return "Not"
	case opNil:
		return "Nil"
	case opComma:
		return "Comma"
	case opFloor:
		return "Floor"
	case opCeil:
		return "Ceil"
	case opRound:
		return "Round"
	case opTrunc:
		return "Trunc"
	case opListing:
		return "Listing" + flagsString
	default:
		return "Invalid"
	}
}

const opFlagIgnoreBlanks = 1

var compareOps = map[string]opKind{
	"eq":           opEqual,
	"not_eq":       opNotEqual,
	"less":         opLess,
	"less_eq":      opLessEqual,
	"more":         opMore,
	"more_eq":      opMoreEqual,
	"contains":     opContains,
	"not_contains": opNotContains,
}

var modeOps = map[string]opKind{
	"assign":         opAssign,
	"assign_replace": opAssignReplace,
	"prepend":        opPrepend,
	"prepend_first":  opPrependFirst,
	"append":         opAppend,
	"append_last":    opAppendLast,
	"delete":         opDelete,
	"delete_all":     opDeleteAll,
}

type exprMatrix struct {
	xx, xy, yx, yy *expression
}

type exprName struct {
	object Object
	kind   matchKind
}

type exprTree struct {
	left, right *expression
}

type exprNode interface {
	// returns a deep copy of the expression
	copyExpr() exprNode
	// return the Go source code for this expression node
	asGoSource() string
}

type expression struct {
	u  exprNode
	op opKind
}

func (exp exprMatrix) copyExpr() exprNode {
	return exprMatrix{xx: exp.xx.copyT(), xy: exp.xy.copyT(), yx: exp.yx.copyT(), yy: exp.yy.copyT()}
}

func (exp exprName) copyExpr() exprNode { return exp }

func (exp exprTree) copyExpr() exprNode {
	return exprTree{left: exp.left.copyT(), right: exp.right.copyT()}
}

func (exp ruleTest) copyExpr() exprNode { return exp.copyT() }

func (exp ruleEdit) copyExpr() exprNode { return exp.copyT() }

func (exp *expression) copyExpr() exprNode { return exp.copyT() }

func (exp *expression) copyT() *expression {
	if exp == nil {
		return nil
	}
	out := *exp
	out.u = exp.u.copyExpr()
	return &out
}

func (exp Pattern) copyExpr() exprNode { return exp.Duplicate() }

func newExprOp(left, right *expression, op opKind) *expression {
	return &expression{op: op, u: exprTree{left: left, right: right}}
}

func (expr expression) String() string {
	switch expr.op.getOp() {
	case opInt, opDouble, opString, opRange, opBool, opConst:
		return fmt.Sprintf("%v", expr.u)
	case opMatrix:
		m := expr.u.(exprMatrix)
		return fmt.Sprintf("[%s %s; %s %s]", m.xx, m.xy, m.yx, m.yy)
	case opCharSet:
		return "charset"
	case opLangSet:
		return fmt.Sprintf("langset: %s", expr.u.(Langset))
	case opNil:
		return "nil"
	case opField:
		name := expr.u.(exprName)
		return fmt.Sprintf("%s (%s)", name.object, name.kind)
	case opQuest:
		tree := expr.u.(exprTree)
		treeRight := tree.right.u.(exprTree)
		return fmt.Sprintf("%s quest %s colon %s", tree.left, treeRight.left, treeRight.right)
	case opAssign, opAssignReplace, opPrependFirst, opPrepend, opAppend, opAppendLast, opOr,
		opAnd, opEqual, opNotEqual, opLess, opLessEqual, opMore, opMoreEqual, opContains, opListing,
		opNotContains, opPlus, opMinus, opTimes, opDivide, opComma:
		tree := expr.u.(exprTree)
		return fmt.Sprintf("%s %s %s", tree.left, expr.op, tree.right)
	case opNot:
		return fmt.Sprintf("Not %s", expr.u.(exprTree).left)
	case opFloor:
		return fmt.Sprintf("Floor %s", expr.u.(exprTree).left)
	case opCeil:
		return fmt.Sprintf("Ceil %s", expr.u.(exprTree).left)
	case opRound:
		return fmt.Sprintf("Round %s", expr.u.(exprTree).left)
	case opTrunc:
		return fmt.Sprintf("Trunc %s", expr.u.(exprTree).left)
	default:
		return "<invalid expr>"
	}
}

func (e *expression) evaluate(result, query Pattern, kind matchKind) Value {
	var v Value
	op := e.op.getOp()
	switch op {
	case opInt, opDouble, opString, opCharSet, opLangSet, opRange, opBool:
		v = e.u.(Value)
	case opMatrix:
		mexpr := e.u.(exprMatrix)
		v = Matrix{} // promotion hint
		xx, xxIsFloat := promote(mexpr.xx.evaluate(result, query, kind), v).(Float)
		xy, xyIsFloat := promote(mexpr.xy.evaluate(result, query, kind), v).(Float)
		yx, yxIsFloat := promote(mexpr.yx.evaluate(result, query, kind), v).(Float)
		yy, yyIsFloat := promote(mexpr.yy.evaluate(result, query, kind), v).(Float)

		if xxIsFloat && xyIsFloat && yxIsFloat && yyIsFloat {
			v = Matrix{Xx: float32(xx), Xy: float32(xy), Yx: float32(yx), Yy: float32(yy)}
		} else {
			v = nil
		}
	case opField:
		name := e.u.(exprName)
		var res Result
		if kind == MatchResult && name.kind == MatchQuery {
			v, res = query.GetAt(name.object, 0)
			if res != ResultMatch {
				v = nil
			}
		} else if kind == MatchQuery && name.kind == MatchResult {
			log.Println("fontconfig: <name> tag has target=\"font\" in a <match target=\"pattern\">.")
			v = nil
		} else {
			v, res = result.GetAt(name.object, 0)
			if res != ResultMatch {
				v = nil
			}
		}
	case opConst:
		if ct, ok := nameConstant(e.u.(String)); ok {
			v = Int(ct)
		} else {
			v = nil
		}
	case opQuest:
		tree := e.u.(exprTree)
		vl := tree.left.evaluate(result, query, kind)
		if vb, isBool := vl.(Bool); isBool {
			right := tree.right.u.(exprTree)
			if vb != 0 {
				v = right.left.evaluate(result, query, kind)
			} else {
				v = right.right.evaluate(result, query, kind)
			}
		} else {
			v = nil
		}
	case opEqual, opNotEqual, opLess, opLessEqual, opMore, opMoreEqual, opContains, opNotContains, opListing:
		tree := e.u.(exprTree)
		vl := tree.left.evaluate(result, query, kind)
		vr := tree.right.evaluate(result, query, kind)
		cp := compareValue(vl, e.op, vr)
		v = False
		if cp {
			v = True
		}
	case opOr, opAnd, opPlus, opMinus, opTimes, opDivide:
		tree := e.u.(exprTree)
		vl := tree.left.evaluate(result, query, kind)
		vr := tree.right.evaluate(result, query, kind)
		vle := promote(vl, vr)
		vre := promote(vr, vle)
		v = nil
		switch vle := vle.(type) {
		case Float:
			vre, sameType := vre.(Float)
			if !sameType {
				break
			}
			switch op {
			case opPlus:
				v = vle + vre
			case opMinus:
				v = vle - vre
			case opTimes:
				v = vle * vre
			case opDivide:
				v = vle / vre
			}
			if vf, ok := v.(Float); ok && vf == Float(int(vf)) {
				v = Int(vf)
			}
		case Bool:
			vre, sameType := vre.(Bool)
			if !sameType {
				break
			}
			switch op {
			case opOr:
				v = vle | vre
			case opAnd:
				v = vle & vre
			}
		case String:
			vre, sameType := vre.(String)
			if !sameType {
				break
			}
			switch op {
			case opPlus:
				v = vle + vre
			}
		case Matrix:
			vre, sameType := vre.(Matrix)
			if !sameType {
				break
			}
			switch op {
			case opTimes:
				v = vle.Multiply(vre)
			}
		case Charset:
			vre, sameType := vre.(Charset)
			if !sameType {
				break
			}
			switch op {
			case opPlus:
				v = charsetUnion(vle, vre)
			case opMinus:
				v = charsetSubtract(vle, vre)
			}
		case Langset:
			vre, sameType := vre.(Langset)
			if !sameType {
				break
			}
			switch op {
			case opPlus:
				v = langSetUnion(vle, vre)
			case opMinus:
				v = langSetSubtract(vle, vre)
			}
		}
	case opNot:
		tree := e.u.(exprTree)
		vl := tree.left.evaluate(result, query, kind)
		if b, ok := vl.(Bool); ok {
			v = 1 - b&1
		}
	case opFloor, opCeil, opRound, opTrunc:
		tree := e.u.(exprTree)
		vl := tree.left.evaluate(result, query, kind)
		switch vl := vl.(type) {
		case Int:
			v = vl
		case Float:
			switch op {
			case opFloor:
				v = Int(math.Floor(float64(vl)))
			case opCeil:
				v = Int(math.Ceil(float64(vl)))
			case opRound:
				v = Int(math.Round(float64(vl)))
			case opTrunc:
				v = Int(math.Trunc(float64(vl)))
			}
		}
	}
	return v
}

func (parser *configParser) typecheckValue(value, type_ typeMeta) error {
	if (value == typeInt{}) {
		value = typeFloat{}
	}
	if (type_ == typeInt{}) {
		type_ = typeFloat{}
	}
	if value != type_ {
		if (value == typeLangSet{} && type_ == typeString{}) ||
			(value == typeString{} && type_ == typeLangSet{}) ||
			(value == typeFloat{} && type_ == typeRange{}) {
			return nil
		}
		if type_ == nil {
			return nil
		}
		/* It's perfectly fine to use user-define elements in expressions,
		 * so don't warn in that case. */
		if value == nil {
			return nil
		}
		return fmt.Errorf("saw %T, expected %T", value, type_)
	}
	return nil
}

func (parser *configParser) typecheckExpr(expr *expression, type_ typeMeta) (err error) {
	// If parsing the expression failed, some nodes may be nil
	if expr == nil {
		return nil
	}

	defer func() {
		if err != nil {
			err = parser.error("expression %s: %s", expr, err)
		}
	}()

	switch expr.op.getOp() {
	case opInt, opDouble:
		err = parser.typecheckValue(typeFloat{}, type_)
	case opString:
		err = parser.typecheckValue(typeString{}, type_)
	case opMatrix:
		err = parser.typecheckValue(typeMatrix{}, type_)
	case opBool:
		err = parser.typecheckValue(typeBool{}, type_)
	case opCharSet:
		err = parser.typecheckValue(typeCharSet{}, type_)
	case opLangSet:
		err = parser.typecheckValue(typeLangSet{}, type_)
	case opRange:
		err = parser.typecheckValue(typeRange{}, type_)
	case opField:
		name := expr.u.(exprName)
		o, ok := objects[name.object.String()]
		if ok {
			err = parser.typecheckValue(o.typeInfo, type_)
		}
	case opConst:
		c := nameGetConstant(string(expr.u.(String)))
		if c != nil {
			o, ok := objects[c.object.String()]
			if ok {
				err = parser.typecheckValue(o.typeInfo, type_)
			}
		} else {
			err = parser.error("invalid constant used : %s", expr.u.(String))
		}
	case opQuest:
		tree := expr.u.(exprTree)
		if err = parser.typecheckExpr(tree.left, typeBool{}); err != nil {
			return err
		}
		rightTree := tree.right.u.(exprTree)
		if err = parser.typecheckExpr(rightTree.left, type_); err != nil {
			return err
		}
		if err = parser.typecheckExpr(rightTree.right, type_); err != nil {
			return err
		}
	case opEqual, opNotEqual, opLess, opLessEqual, opMore, opMoreEqual, opContains, opNotContains, opListing:
		err = parser.typecheckValue(typeBool{}, type_)
	case opComma, opOr, opAnd, opPlus, opMinus, opTimes, opDivide:
		tree := expr.u.(exprTree)
		if err = parser.typecheckExpr(tree.left, type_); err != nil {
			return err
		}
		err = parser.typecheckExpr(tree.right, type_)
	case opNot:
		tree := expr.u.(exprTree)
		if err = parser.typecheckValue(typeBool{}, type_); err != nil {
			return err
		}
		err = parser.typecheckExpr(tree.left, typeBool{})
	case opFloor, opCeil, opRound, opTrunc:
		tree := expr.u.(exprTree)
		if err = parser.typecheckValue(typeFloat{}, type_); err != nil {
			return err
		}
		err = parser.typecheckExpr(tree.left, typeFloat{})
	}
	return err
}

func promote(v, u Value) Value {
	// the C implemention use a pre-allocated buffer to avoid allocations
	// we choose to simplify and not use buffer
	switch val := v.(type) {
	case Int:
		v = promoteFloat(Float(val), u)
	case Float:
		v = promoteFloat(val, u)
	case nil:
		switch u.(type) {
		case Matrix:
			v = Identity
		case Langset:
			v = langSetPromote("")
		case Charset:
			v = Charset{}
		}
	case String:
		if _, ok := u.(Langset); ok {
			v = langSetPromote(val)
		}
	}
	return v
}

func promoteFloat(val Float, u Value) Value {
	if _, ok := u.(Range); ok {
		return rangePromote(val)
	}
	return val
}

func compareValue(leftO Value, op opKind, rightO Value) bool {
	flags := op.getFlags()
	op = op.getOp()
	retNoMatchingType := false
	if op == opNotEqual || op == opNotContains {
		retNoMatchingType = true
	}
	ret := false

	// to avoid checking for type equality we begin by promoting
	// and we will check later in the type switch
	leftO = promote(leftO, rightO)
	rightO = promote(rightO, leftO)

	switch l := leftO.(type) {
	case Int:
		r, sameType := rightO.(Int)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opEqual, opContains, opListing:
			ret = l == r
		case opNotEqual, opNotContains:
			ret = l != r
		case opLess:
			ret = l < r
		case opLessEqual:
			ret = l <= r
		case opMore:
			ret = l > r
		case opMoreEqual:
			ret = l >= r
		}
	case Float:
		r, sameType := rightO.(Float)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opEqual, opContains, opListing:
			ret = l == r
		case opNotEqual, opNotContains:
			ret = l != r
		case opLess:
			ret = l < r
		case opLessEqual:
			ret = l <= r
		case opMore:
			ret = l > r
		case opMoreEqual:
			ret = l >= r
		}
	case Bool:
		r, sameType := rightO.(Bool)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opEqual:
			ret = l == r
		case opContains, opListing:
			ret = l == r || l >= DontCare
		case opNotEqual:
			ret = l != r
		case opNotContains:
			ret = !(l == r || l >= DontCare)
		case opLess:
			ret = l != r && r >= DontCare
		case opLessEqual:
			ret = l == r || r >= DontCare
		case opMore:
			ret = l != r && l >= DontCare
		case opMoreEqual:
			ret = l == r || l >= DontCare
		}
	case String:
		r, sameType := rightO.(String)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opEqual, opListing:
			if flags&opFlagIgnoreBlanks != 0 {
				ret = cmpIgnoreBlanksAndCase(string(l), string(r)) == 0
			} else {
				ret = cmpIgnoreCase(string(l), string(r)) == 0
			}
		case opContains:
			ret = indexIgnoreCase(string(l), string(r)) != -1
		case opNotEqual:
			if flags&opFlagIgnoreBlanks != 0 {
				ret = cmpIgnoreBlanksAndCase(string(l), string(r)) != 0
			} else {
				ret = cmpIgnoreCase(string(l), string(r)) != 0
			}
		case opNotContains:
			ret = indexIgnoreCase(string(l), string(r)) == -1
		}
	case Matrix:
		r, sameType := rightO.(Matrix)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opEqual, opContains, opListing:
			ret = l == r
		case opNotEqual, opNotContains:
			ret = !(l == r)
		}
	case Charset:
		r, sameType := rightO.(Charset)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opContains, opListing:
			// left contains right if right is a subset of left
			ret = r.isSubset(l)
		case opNotContains:
			// left contains right if right is a subset of left
			ret = !r.isSubset(l)
		case opEqual:
			ret = charsetEqual(l, r)
		case opNotEqual:
			ret = !charsetEqual(l, r)
		}
	case Langset:
		r, sameType := rightO.(Langset)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opContains, opListing:
			ret = l.includes(r)
		case opNotContains:
			ret = !l.includes(r)
		case opEqual:
			ret = langsetEqual(l, r)
		case opNotEqual:
			ret = !langsetEqual(l, r)
		}
	case nil:
		sameType := rightO == nil
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case opEqual, opContains, opListing:
			ret = true
		}
	case Range:
		r, sameType := rightO.(Range)
		if !sameType {
			return retNoMatchingType
		}
		ret = rangeCompare(op, l, r)
	}
	return ret
}

func (e *expression) toValues(p, p_pat Pattern, kind matchKind, binding valueBinding) valueList {
	if e == nil {
		return nil
	}

	var l valueList
	if e.op.getOp() == opComma {
		tree := e.u.(exprTree)
		v := tree.left.evaluate(p, p_pat, kind)
		next := tree.right.toValues(p, p_pat, kind, binding)
		l = append(valueList{valueElt{Value: v, Binding: binding}}, next...)
	} else {
		v := e.evaluate(p, p_pat, kind)
		l = valueList{valueElt{Value: v, Binding: binding}}
	}

	if l[0].Value == nil {
		l = l[1:]
	}

	return l
}
