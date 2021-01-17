package fontconfig

import (
	"fmt"
	"log"
	"math"
)

type FcOp uint

const (
	FcOpInteger FcOp = iota
	FcOpDouble
	FcOpString
	FcOpMatrix
	FcOpRange
	FcOpBool
	FcOpCharSet
	FcOpLangSet
	FcOpNil
	FcOpField
	FcOpConst
	FcOpAssign
	FcOpAssignReplace
	FcOpPrependFirst
	FcOpPrepend
	FcOpAppend
	FcOpAppendLast
	FcOpDelete
	FcOpDeleteAll
	FcOpQuest
	FcOpOr
	FcOpAnd
	FcOpEqual
	FcOpNotEqual
	FcOpContains
	FcOpListing
	FcOpNotContains
	FcOpLess
	FcOpLessEqual
	FcOpMore
	FcOpMoreEqual
	FcOpPlus
	FcOpMinus
	FcOpTimes
	FcOpDivide
	FcOpNot
	FcOpComma
	FcOpFloor
	FcOpCeil
	FcOpRound
	FcOpTrunc
	FcOpInvalid
)

func opWithFlags(x FcOp, f int) FcOp {
	return x | FcOp(f)<<16
}

func (x FcOp) getOp() FcOp {
	return x & 0xffff
}

func (x FcOp) getFlags() int {
	return (int(x) & 0xffff0000) >> 16
}

func (x FcOp) String() string {
	flagsString := ""
	if x.getFlags()&FcOpFlagIgnoreBlanks != 0 {
		flagsString = " (ignore blanks)"
	}
	switch x.getOp() {
	case FcOpInteger:
		return "Integer"
	case FcOpDouble:
		return "Double"
	case FcOpString:
		return "String"
	case FcOpMatrix:
		return "Matrix"
	case FcOpRange:
		return "Range"
	case FcOpBool:
		return "Bool"
	case FcOpCharSet:
		return "CharSet"
	case FcOpLangSet:
		return "LangSet"
	case FcOpField:
		return "Field"
	case FcOpConst:
		return "Const"
	case FcOpAssign:
		return "Assign"
	case FcOpAssignReplace:
		return "AssignReplace"
	case FcOpPrepend:
		return "Prepend"
	case FcOpPrependFirst:
		return "PrependFirst"
	case FcOpAppend:
		return "Append"
	case FcOpAppendLast:
		return "AppendLast"
	case FcOpDelete:
		return "Delete"
	case FcOpDeleteAll:
		return "DeleteAll"
	case FcOpQuest:
		return "Quest"
	case FcOpOr:
		return "Or"
	case FcOpAnd:
		return "And"
	case FcOpEqual:
		return "Equal" + flagsString
	case FcOpNotEqual:
		return "NotEqual" + flagsString
	case FcOpLess:
		return "Less"
	case FcOpLessEqual:
		return "LessEqual"
	case FcOpMore:
		return "More"
	case FcOpMoreEqual:
		return "MoreEqual"
	case FcOpContains:
		return "Contains"
	case FcOpNotContains:
		return "NotContains"
	case FcOpPlus:
		return "Plus"
	case FcOpMinus:
		return "Minus"
	case FcOpTimes:
		return "Times"
	case FcOpDivide:
		return "Divide"
	case FcOpNot:
		return "Not"
	case FcOpNil:
		return "Nil"
	case FcOpComma:
		return "Comma"
	case FcOpFloor:
		return "Floor"
	case FcOpCeil:
		return "Ceil"
	case FcOpRound:
		return "Round"
	case FcOpTrunc:
		return "Trunc"
	case FcOpListing:
		return "Listing" + flagsString
	default:
		return "Invalid"
	}
}

const FcOpFlagIgnoreBlanks = 1

var fcCompareOps = map[string]FcOp{
	"eq":           FcOpEqual,
	"not_eq":       FcOpNotEqual,
	"less":         FcOpLess,
	"less_eq":      FcOpLessEqual,
	"more":         FcOpMore,
	"more_eq":      FcOpMoreEqual,
	"contains":     FcOpContains,
	"not_contains": FcOpNotContains,
}
var fcModeOps = map[string]FcOp{
	"assign":         FcOpAssign,
	"assign_replace": FcOpAssignReplace,
	"prepend":        FcOpPrepend,
	"prepend_first":  FcOpPrependFirst,
	"append":         FcOpAppend,
	"append_last":    FcOpAppendLast,
	"delete":         FcOpDelete,
	"delete_all":     FcOpDeleteAll,
}

type FcExprMatrix struct {
	xx, xy, yx, yy *FcExpr
}

type FcExprName struct {
	object Object
	kind   FcMatchKind
}

type exprTree struct {
	left, right *FcExpr
}

type exprNode interface {
	isExpr()
}

type FcExpr struct {
	op FcOp
	u  exprNode
}

func (FcExprMatrix) isExpr() {}
func (FcExprName) isExpr()   {}
func (exprTree) isExpr()     {}
func (FcTest) isExpr()       {}
func (FcEdit) isExpr()       {}
func (*FcExpr) isExpr()      {}
func (Pattern) isExpr()      {}

// union {
// int		ival;
// double		dval;
// const FcChar8	*sval;
// FcExprMatrix	*mexpr;
// FcBool		bval;
// FcCharset	*cval;
// FcLangSet	*lval;
// FcRange		*rval;

// FcExprName	name;
// const FcChar8	*constant;
// struct {
//     struct _FcExpr *left, *right;
// } tree;
// } u;

func newExprOp(left, right *FcExpr, op FcOp) *FcExpr {
	return &FcExpr{op: op, u: exprTree{left: left, right: right}}
}

func (expr *FcExpr) String() string {
	if expr == nil {
		return "nil"
	}

	switch expr.op.getOp() {
	case FcOpInteger, FcOpDouble, FcOpString, FcOpRange, FcOpBool, FcOpConst:
		return fmt.Sprintf("%s", expr.u)
	case FcOpMatrix:
		m := expr.u.(FcExprMatrix)
		return fmt.Sprintf("[%s %s; %s %s]", m.xx, m.xy, m.yx, m.yy)
	case FcOpCharSet:
		return "charset"
	case FcOpLangSet:
		return fmt.Sprintf("langset: %s", expr.u.(FcLangSet))
	case FcOpNil:
		return ("nil")
	case FcOpField:
		name := expr.u.(FcExprName)
		return fmt.Sprintf("%s (%s)", name.object, name.kind)
	case FcOpQuest:
		tree := expr.u.(exprTree)
		treeRight := tree.right.u.(exprTree)
		return fmt.Sprintf("%s quest %s colon %s", tree.left, treeRight.left, treeRight.right)
	case FcOpAssign, FcOpAssignReplace, FcOpPrependFirst, FcOpPrepend, FcOpAppend, FcOpAppendLast, FcOpOr,
		FcOpAnd, FcOpEqual, FcOpNotEqual, FcOpLess, FcOpLessEqual, FcOpMore, FcOpMoreEqual, FcOpContains, FcOpListing,
		FcOpNotContains, FcOpPlus, FcOpMinus, FcOpTimes, FcOpDivide, FcOpComma:
		tree := expr.u.(exprTree)
		return fmt.Sprintf("%s %s %s", tree.left, expr.op, tree.right)
	case FcOpNot:
		return fmt.Sprintf("Not %s", expr.u.(exprTree).left)
	case FcOpFloor:
		return fmt.Sprintf("Floor %s", expr.u.(exprTree).left)
	case FcOpCeil:
		return fmt.Sprintf("Ceil %s", expr.u.(exprTree).left)
	case FcOpRound:
		return fmt.Sprintf("Round %s", expr.u.(exprTree).left)
	case FcOpTrunc:
		return fmt.Sprintf("Trunc %s", expr.u.(exprTree).left)
	default:
		return "<invalid expr>"
	}
}

func (e *FcExpr) FcConfigEvaluate(p, p_pat Pattern, kind FcMatchKind) FcValue {
	var v FcValue
	op := e.op.getOp()
	switch op {
	case FcOpInteger, FcOpDouble, FcOpString, FcOpCharSet, FcOpLangSet, FcOpRange, FcOpBool:
		v = e.u.(FcValue)
	case FcOpMatrix:
		mexpr := e.u.(FcExprMatrix)
		v = Matrix{} // promotion hint
		xx, xxIsFloat := FcConfigPromote(mexpr.xx.FcConfigEvaluate(p, p_pat, kind), v).(Float)
		xy, xyIsFloat := FcConfigPromote(mexpr.xy.FcConfigEvaluate(p, p_pat, kind), v).(Float)
		yx, yxIsFloat := FcConfigPromote(mexpr.yx.FcConfigEvaluate(p, p_pat, kind), v).(Float)
		yy, yyIsFloat := FcConfigPromote(mexpr.yy.FcConfigEvaluate(p, p_pat, kind), v).(Float)

		if xxIsFloat && xyIsFloat && yxIsFloat && yyIsFloat {
			v = Matrix{Xx: float64(xx), Xy: float64(xy), Yx: float64(yx), Yy: float64(yy)}
		} else {
			v = nil
		}
	case FcOpField:
		name := e.u.(FcExprName)
		var res FcResult
		if kind == FcMatchFont && name.kind == FcMatchPattern {
			v, res = p_pat.FcPatternObjectGet(name.object, 0)
			if res != FcResultMatch {
				v = nil
			}
		} else if kind == FcMatchPattern && name.kind == FcMatchFont {
			log.Println("fFontconfig: <name> tag has target=\"font\" in a <match target=\"pattern\">.")
			v = nil
		} else {
			v, res = p_pat.FcPatternObjectGet(name.object, 0)
			if res != FcResultMatch {
				v = nil
			}
		}
	case FcOpConst:
		if ct, ok := nameConstant(e.u.(String)); ok {
			v = Int(ct)
		} else {
			v = nil
		}
	case FcOpQuest:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		if vb, isBool := vl.(FcBool); isBool {
			right := tree.right.u.(exprTree)
			if vb != 0 {
				v = right.left.FcConfigEvaluate(p, p_pat, kind)
			} else {
				v = right.right.FcConfigEvaluate(p, p_pat, kind)
			}
		} else {
			v = nil
		}
	case FcOpEqual, FcOpNotEqual, FcOpLess, FcOpLessEqual, FcOpMore, FcOpMoreEqual, FcOpContains, FcOpNotContains, FcOpListing:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		vr := tree.right.FcConfigEvaluate(p, p_pat, kind)
		cp := compareValue(vl, e.op, vr)
		v = FcFalse
		if cp {
			v = FcTrue
		}
	case FcOpOr, FcOpAnd, FcOpPlus, FcOpMinus, FcOpTimes, FcOpDivide:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		vr := tree.right.FcConfigEvaluate(p, p_pat, kind)
		vle := FcConfigPromote(vl, vr)
		vre := FcConfigPromote(vr, vle)
		v = nil
		switch vle := vle.(type) {
		case Float:
			vre, sameType := vre.(Float)
			if !sameType {
				break
			}
			switch op {
			case FcOpPlus:
				v = vle + vre
			case FcOpMinus:
				v = vle - vre
			case FcOpTimes:
				v = vle * vre
			case FcOpDivide:
				v = vle / vre
			}
			if vf, ok := v.(Float); ok && vf == Float(int(vf)) {
				v = Int(vf)
			}
		case FcBool:
			vre, sameType := vre.(FcBool)
			if !sameType {
				break
			}
			switch op {
			case FcOpOr:
				v = vle | vre
			case FcOpAnd:
				v = vle & vre
			}
		case String:
			vre, sameType := vre.(String)
			if !sameType {
				break
			}
			switch op {
			case FcOpPlus:
				v = vle + vre
			}
		case Matrix:
			vre, sameType := vre.(Matrix)
			if !sameType {
				break
			}
			switch op {
			case FcOpTimes:
				v = vle.Multiply(vre)
			}
		case Charset:
			vre, sameType := vre.(Charset)
			if !sameType {
				break
			}
			switch op {
			case FcOpPlus:
				v = charsetUnion(vle, vre)
			case FcOpMinus:
				v = charsetSubtract(vle, vre)
			}
		case FcLangSet:
			vre, sameType := vre.(FcLangSet)
			if !sameType {
				break
			}
			switch op {
			case FcOpPlus:
				v = langSetUnion(vle, vre)
			case FcOpMinus:
				v = langSetSubtract(vle, vre)
			}
		}
	case FcOpNot:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		if b, ok := vl.(FcBool); ok {
			v = 1 - b&1
		}
	case FcOpFloor, FcOpCeil, FcOpRound, FcOpTrunc:
		tree := e.u.(exprTree)
		vl := tree.left.FcConfigEvaluate(p, p_pat, kind)
		switch vl := vl.(type) {
		case Int:
			v = vl
		case Float:
			switch op {
			case FcOpFloor:
				v = Int(math.Floor(float64(vl)))
			case FcOpCeil:
				v = Int(math.Ceil(float64(vl)))
			case FcOpRound:
				v = Int(math.Round(float64(vl)))
			case FcOpTrunc:
				v = Int(math.Trunc(float64(vl)))
			}
		}
	}
	return v
}

func (parser *configParser) typecheckValue(value, type_ typeMeta) error {
	if (value == typeInteger{}) {
		value = typeFloat{}
	}
	if (type_ == typeInteger{}) {
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
		return parser.error("saw %T, expected %T", value, type_)
	}
	return nil
}

func (parser *configParser) typecheckExpr(expr *FcExpr, type_ typeMeta) (err error) {
	// If parsing the expression failed, some nodes may be nil
	if expr == nil {
		return nil
	}

	switch expr.op.getOp() {
	case FcOpInteger, FcOpDouble:
		err = parser.typecheckValue(typeFloat{}, type_)
	case FcOpString:
		err = parser.typecheckValue(typeString{}, type_)
	case FcOpMatrix:
		err = parser.typecheckValue(typeMatrix{}, type_)
	case FcOpBool:
		err = parser.typecheckValue(typeBool{}, type_)
	case FcOpCharSet:
		err = parser.typecheckValue(typeCharSet{}, type_)
	case FcOpLangSet:
		err = parser.typecheckValue(typeLangSet{}, type_)
	case FcOpRange:
		err = parser.typecheckValue(typeRange{}, type_)
	case FcOpField:
		name := expr.u.(FcExprName)
		o, ok := objects[name.object.String()]
		if ok {
			err = parser.typecheckValue(o.parser, type_)
		}
	case FcOpConst:
		c := nameGetConstant(string(expr.u.(String)))
		if c != nil {
			o, ok := objects[c.object.String()]
			if ok {
				err = parser.typecheckValue(o.parser, type_)
			}
		} else {
			err = parser.error("invalid constant used : %s", expr.u.(String))
		}
	case FcOpQuest:
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
	case FcOpEqual, FcOpNotEqual, FcOpLess, FcOpLessEqual, FcOpMore, FcOpMoreEqual, FcOpContains, FcOpNotContains, FcOpListing:
		err = parser.typecheckValue(typeBool{}, type_)
	case FcOpComma, FcOpOr, FcOpAnd, FcOpPlus, FcOpMinus, FcOpTimes, FcOpDivide:
		tree := expr.u.(exprTree)
		if err = parser.typecheckExpr(tree.left, type_); err != nil {
			return err
		}
		err = parser.typecheckExpr(tree.right, type_)
	case FcOpNot:
		tree := expr.u.(exprTree)
		if err = parser.typecheckValue(typeBool{}, type_); err != nil {
			return err
		}
		err = parser.typecheckExpr(tree.left, typeBool{})
	case FcOpFloor, FcOpCeil, FcOpRound, FcOpTrunc:
		tree := expr.u.(exprTree)
		if err = parser.typecheckValue(typeFloat{}, type_); err != nil {
			return err
		}
		err = parser.typecheckExpr(tree.left, typeFloat{})
	}
	return err
}

// the C implemention use a pre-allocated buffer to avoid allocations
// we choose to simplify and not use buffer
func FcConfigPromote(v, u FcValue) FcValue {
	switch val := v.(type) {
	case Int:
		v = promoteFloat64(Float(val), u)
	case Float:
		v = promoteFloat64(val, u)
	case nil:
		switch u.(type) {
		case Matrix:
			v = Identity
		case FcLangSet:
			v = langSetPromote("")
		case Charset:
			v = Charset{}
		}
	case String:
		if _, ok := u.(FcLangSet); ok {
			v = langSetPromote(val)
		}
	}
	return v
}

func promoteFloat64(val Float, u FcValue) FcValue {
	if _, ok := u.(FcRange); ok {
		return FcRangePromote(val)
	}
	return val
}

func compareValue(left_o FcValue, op FcOp, right_o FcValue) bool {
	flags := op.getFlags()
	op = op.getOp()
	retNoMatchingType := false
	if op == FcOpNotEqual || op == FcOpNotContains {
		retNoMatchingType = true
	}
	ret := false

	// to avoid checking for type equality we begin by promoting
	// and we will check later in the type switch
	left_o = FcConfigPromote(left_o, right_o)
	left_o = FcConfigPromote(right_o, left_o)

	switch l := left_o.(type) {
	case Int:
		r, sameType := right_o.(Int)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = l != r
		case FcOpLess:
			ret = l < r
		case FcOpLessEqual:
			ret = l <= r
		case FcOpMore:
			ret = l > r
		case FcOpMoreEqual:
			ret = l >= r
		}
	case Float:
		r, sameType := right_o.(Float)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = l != r
		case FcOpLess:
			ret = l < r
		case FcOpLessEqual:
			ret = l <= r
		case FcOpMore:
			ret = l > r
		case FcOpMoreEqual:
			ret = l >= r
		}
	case FcBool:
		r, sameType := right_o.(FcBool)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual:
			ret = l == r
		case FcOpContains, FcOpListing:
			ret = l == r || l >= FcDontCare
		case FcOpNotEqual:
			ret = l != r
		case FcOpNotContains:
			ret = !(l == r || l >= FcDontCare)
		case FcOpLess:
			ret = l != r && r >= FcDontCare
		case FcOpLessEqual:
			ret = l == r || r >= FcDontCare
		case FcOpMore:
			ret = l != r && l >= FcDontCare
		case FcOpMoreEqual:
			ret = l == r || l >= FcDontCare
		}
	case String:
		r, sameType := right_o.(String)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpListing:
			if flags&FcOpFlagIgnoreBlanks != 0 {
				ret = FcStrCmpIgnoreBlanksAndCase(string(l), string(r)) == 0
			} else {
				ret = FcStrCmpIgnoreCase(string(l), string(r)) == 0
			}
		case FcOpContains:
			ret = FcStrStrIgnoreCase(string(l), string(r)) != -1
		case FcOpNotEqual:
			if flags&FcOpFlagIgnoreBlanks != 0 {
				ret = FcStrCmpIgnoreBlanksAndCase(string(l), string(r)) != 0
			} else {
				ret = FcStrCmpIgnoreCase(string(l), string(r)) != 0
			}
		case FcOpNotContains:
			ret = FcStrStrIgnoreCase(string(l), string(r)) == -1
		}
	case Matrix:
		r, sameType := right_o.(Matrix)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = !(l == r)
		}
	case Charset:
		r, sameType := right_o.(Charset)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpContains, FcOpListing:
			// left contains right if right is a subset of left
			ret = r.isSubset(l)
		case FcOpNotContains:
			// left contains right if right is a subset of left
			ret = !r.isSubset(l)
		case FcOpEqual:
			ret = FcCharsetEqual(l, r)
		case FcOpNotEqual:
			ret = !FcCharsetEqual(l, r)
		}
	case FcLangSet:
		r, sameType := right_o.(FcLangSet)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpContains, FcOpListing:
			ret = l.FcLangSetContains(r)
		case FcOpNotContains:
			ret = !l.FcLangSetContains(r)
		case FcOpEqual:
			ret = FcLangSetEqual(l, r)
		case FcOpNotEqual:
			ret = !FcLangSetEqual(l, r)
		}
	case nil:
		sameType := right_o == nil
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = true
		}
	case *FtFace:
		r, sameType := right_o.(*FtFace)
		if !sameType {
			return retNoMatchingType
		}
		switch op {
		case FcOpEqual, FcOpContains, FcOpListing:
			ret = l == r
		case FcOpNotEqual, FcOpNotContains:
			ret = l != r
		}
	case FcRange:
		r, sameType := right_o.(FcRange)
		if !sameType {
			return retNoMatchingType
		}
		ret = FcRangeCompare(op, l, r)
	}
	return ret
}

func (e *FcExpr) FcConfigValues(p, p_pat Pattern, kind FcMatchKind, binding FcValueBinding) ValueList {
	if e == nil {
		return nil
	}

	var l ValueList
	if e.op.getOp() == FcOpComma {
		tree := e.u.(exprTree)
		v := tree.left.FcConfigEvaluate(p, p_pat, kind)
		next := tree.right.FcConfigValues(p, p_pat, kind, binding)
		l = append(ValueList{valueElt{value: v, binding: binding}}, next...)
	} else {
		v := e.FcConfigEvaluate(p, p_pat, kind)
		l = ValueList{valueElt{value: v, binding: binding}}
	}

	if l[0].value == nil {
		l = l[1:]
	}

	return l
}
