package fontconfig

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var errOldSyntax = errors.New("element no longer supported")

func (config *Config) parseAndLoadFromMemory(filename string, content io.Reader) error {
	if debugMode {
		fmt.Printf("Processing config file from %s", filename)
	}

	parser := newConfigParser(filename, config)

	err := xml.NewDecoder(content).Decode(parser)
	if err != nil {
		return fmt.Errorf("cannot process config file from %s: %s", filename, err)
	}

	config.subst = append(config.subst, parser.ruleset)

	return nil
}

// compact form of a tag in xml file
type elemTag uint8

const (
	elementNone elemTag = iota
	elementFontconfig
	elementMatch
	elementAlias
	elementDescription

	elementPrefer
	elementAccept
	elementDefault
	elementFamily

	elementSelectfont
	elementAcceptfont
	elementRejectfont
	elementGlob
	elementPattern
	elementPatelt

	elementTest
	elementEdit
	elementInt
	elementDouble
	elementString
	elementMatrix
	elementRange
	elementBool
	elementCharSet
	elementLangSet
	elementName
	elementConst
	elementOr
	elementAnd
	elementEq
	elementNotEq
	elementLess
	elementLessEq
	elementMore
	elementMoreEq
	elementContains
	elementNotContains
	elementPlus
	elementMinus
	elementTimes
	elementDivide
	elementNot
	elementIf
	elementFloor
	elementCeil
	elementRound
	elementTrunc
)

var elementMap = [...]string{
	elementFontconfig:  "fontconfig",
	elementMatch:       "match",
	elementAlias:       "alias",
	elementDescription: "description",

	elementPrefer:  "prefer",
	elementAccept:  "accept",
	elementDefault: "default",
	elementFamily:  "family",

	elementSelectfont: "selectfont",
	elementAcceptfont: "acceptfont",
	elementRejectfont: "rejectfont",
	elementGlob:       "glob",
	elementPattern:    "pattern",
	elementPatelt:     "patelt",

	elementTest:        "test",
	elementEdit:        "edit",
	elementInt:         "int",
	elementDouble:      "double",
	elementString:      "string",
	elementMatrix:      "matrix",
	elementRange:       "range",
	elementBool:        "bool",
	elementCharSet:     "charset",
	elementLangSet:     "langset",
	elementName:        "name",
	elementConst:       "const",
	elementOr:          "or",
	elementAnd:         "and",
	elementEq:          "eq",
	elementNotEq:       "not_eq",
	elementLess:        "less",
	elementLessEq:      "less_eq",
	elementMore:        "more",
	elementMoreEq:      "more_eq",
	elementContains:    "contains",
	elementNotContains: "not_contains",
	elementPlus:        "plus",
	elementMinus:       "minus",
	elementTimes:       "times",
	elementDivide:      "divide",
	elementNot:         "not",
	elementIf:          "if",
	elementFloor:       "floor",
	elementCeil:        "ceil",
	elementRound:       "round",
	elementTrunc:       "trunc",
}

var elementIgnoreName = [...]string{
	"its",
}

func elemFromName(name xml.Name) (elemTag, error) {
	for i, elem := range elementMap {
		if name.Local == elem {
			return elemTag(i), nil
		}
	}

	for _, ignoreName := range elementIgnoreName {
		if strings.HasSuffix(name.Space, ignoreName) {
			return elementNone, nil
		}
	}
	return 0, fmt.Errorf("unknown element %v", name)
}

func (e elemTag) String() string {
	if int(e) >= len(elementMap) {
		return fmt.Sprintf("invalid element %d", e)
	}
	return elementMap[e]
}

// pStack is one XML containing tag
type pStack struct {
	str     *bytes.Buffer // inner text content
	attr    []xml.Attr
	values  []vstack // the top of the stack is at the end of the slice
	element elemTag
}

// kind of the value: sometimes the type is not enough
// to distinguish
type vstackTag uint8

const (
	vstackNone vstackTag = iota

	vstackString
	vstackFamily
	vstackConstant
	vstackGlob
	vstackName
	vstackPattern

	vstackPrefer
	vstackAccept
	vstackDefault

	vstackInt
	vstackDouble
	vstackMatrix
	vstackRange
	vstackBool
	vstackCharSet
	vstackLangSet

	vstackTest
	vstackExpr
	vstackEdit
)

// parse value
type vstack struct {
	u   exprNode
	tag vstackTag
}

type configParser struct {
	name string

	config  *Config
	ruleset ruleSet

	pstack []pStack // the top of the stack is at the end of the slice
}

func newConfigParser(name string, config *Config) *configParser {
	var parser configParser

	parser.name = name
	parser.config = config
	parser.ruleset.name = name

	return &parser
}

func (parse *configParser) error(format string, args ...interface{}) error {
	var s string
	s = fmt.Sprintf(`fontconfig %s: "%s": `, s, parse.name)
	s += fmt.Sprintf(format+"\n", args...)
	return errors.New(s)
}

func (parser *configParser) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// start by handling the new element
	err := parser.startElement(start.Name, start.Attr)
	if err != nil {
		return err
	}

	// then process the inner content: text or kid element
	for {
		next, err := d.Token()
		if err != nil {
			return err
		}
		// Token is one of StartElement, EndElement, CharData, Comment, ProcInst, or Directive
		switch next := next.(type) {
		case xml.CharData:
			// handle text and keep going
			parser.text(next)
		case xml.EndElement:
			// closing current element: return after processing
			err := parser.endElement()
			return err
		case xml.StartElement:
			// new kid: recurse and keep going for other kids or text
			err := parser.UnmarshalXML(d, next)
			if err != nil {
				return err
			}
		default:
			// ignored, just keep going
		}
	}
}

// return value may be nil if the stack is empty
func (parse *configParser) p() *pStack {
	if len(parse.pstack) == 0 {
		return nil
	}
	return &parse.pstack[len(parse.pstack)-1]
}

// return value may be nil if the stack is empty
func (parse *configParser) v() *vstack {
	if last := parse.p(); last != nil && len(last.values) != 0 {
		return &last.values[len(last.values)-1]
	}
	return nil
}

func (parser *configParser) text(s []byte) {
	p := parser.p()
	if p == nil {
		return
	}
	p.str.Write(s)
}

// add a value to the previous p element, or discard it
func (parser *configParser) createVAndPush() *vstack {
	if len(parser.pstack) >= 2 {
		ps := &parser.pstack[len(parser.pstack)-2]
		ps.values = append(ps.values, vstack{})
		return &ps.values[len(ps.values)-1]
	}
	return nil
}

func (parse *configParser) startElement(name xml.Name, attr []xml.Attr) error {
	switch name.Local {
	case "dir", "cachedir", "cache", "include", "config", "remap-dir", "reset-dirs", "rescan":
		return errOldSyntax
	}

	element, err := elemFromName(name)
	if err != nil {
		return parse.error("start element: %s", err)
	}

	parse.pstackPush(element, attr)
	return nil
}

// push at the end of the slice
func (parse *configParser) pstackPush(element elemTag, attr []xml.Attr) {
	new := pStack{
		element: element,
		attr:    attr,
		str:     new(bytes.Buffer),
	}
	parse.pstack = append(parse.pstack, new)
}

func (parse *configParser) pstackPop() error {
	// the encoding/xml package makes sur tag are matching
	// so parse.pstack has at least one element

	// Don't check the attributes for elementNone
	if last := parse.p(); last.element != elementNone {
		// error on unused attrs.
		for _, attr := range last.attr {
			if attr.Name.Local != "" {
				return parse.error("invalid attribute %s", attr.Name.Local)
			}
		}
	}

	parse.pstack = parse.pstack[:len(parse.pstack)-1]
	return nil
}

// pop from the last vstack
func (parse *configParser) vstackPop() {
	last := parse.p()
	if last == nil || len(last.values) == 0 {
		return
	}
	last.values = last.values[:len(last.values)-1]
}

func (parser *configParser) endElement() error {
	last := parser.p()
	if last == nil { // nothing to do
		return nil
	}
	var err error
	switch last.element {
	case elementMatch:
		err = parser.parseMatch()
	case elementAlias:
		err = parser.parseAlias()
	case elementDescription:
		parser.parseDescription()

	case elementPrefer:
		err = parser.parseFamilies(vstackPrefer)
	case elementAccept:
		err = parser.parseFamilies(vstackAccept)
	case elementDefault:
		err = parser.parseFamilies(vstackDefault)
	case elementFamily:
		parser.parseFamily()

	case elementTest:
		err = parser.parseTest()
	case elementEdit:
		err = parser.parseEdit()

	case elementInt:
		err = parser.parseInt()
	case elementDouble:
		err = parser.parseFloat()
	case elementString:
		parser.parseString(vstackString)
	case elementMatrix:
		err = parser.parseMatrix()
	case elementRange:
		err = parser.parseRange()
	case elementBool:
		err = parser.parseBool()
	case elementCharSet:
		err = parser.parseCharSet()
	case elementLangSet:
		err = parser.parseLangSet()
	case elementSelectfont, elementAcceptfont, elementRejectfont:
		err = parser.parseAcceptRejectFont(last.element)
	case elementGlob:
		parser.parseString(vstackGlob)
	case elementPattern:
		err = parser.parsePattern()
	case elementPatelt:
		err = parser.parsePatelt()
	case elementName:
		err = parser.parseName()
	case elementConst:
		parser.parseString(vstackConstant)
	case elementOr:
		parser.parseBinary(opOr)
	case elementAnd:
		parser.parseBinary(opAnd)
	case elementEq:
		parser.parseBinary(opEqual)
	case elementNotEq:
		parser.parseBinary(opNotEqual)
	case elementLess:
		parser.parseBinary(opLess)
	case elementLessEq:
		parser.parseBinary(opLessEqual)
	case elementMore:
		parser.parseBinary(opMore)
	case elementMoreEq:
		parser.parseBinary(opMoreEqual)
	case elementContains:
		parser.parseBinary(opContains)
	case elementNotContains:
		parser.parseBinary(opNotContains)
	case elementPlus:
		parser.parseBinary(opPlus)
	case elementMinus:
		parser.parseBinary(opMinus)
	case elementTimes:
		parser.parseBinary(opTimes)
	case elementDivide:
		parser.parseBinary(opDivide)
	case elementIf:
		parser.parseBinary(opQuest)
	case elementNot:
		parser.parseUnary(opNot)
	case elementFloor:
		parser.parseUnary(opFloor)
	case elementCeil:
		parser.parseUnary(opCeil)
	case elementRound:
		parser.parseUnary(opRound)
	case elementTrunc:
		parser.parseUnary(opTrunc)
	}
	if err != nil {
		return err
	}

	return parser.pstackPop()
}

func (last *pStack) getAttr(attr string) string {
	if last == nil {
		return ""
	}

	attrs := last.attr

	for i, attrXml := range attrs {
		if attr == attrXml.Name.Local {
			attrs[i].Name.Local = "" // Mark as used.
			return attrXml.Value
		}
	}
	return ""
}

func (parser *configParser) lexBool(bool_ string) (Bool, error) {
	result, err := nameBool(bool_)
	if err != nil {
		return 0, parser.error("\"%s\" is not known boolean", bool_)
	}
	return result, nil
}

func (parser *configParser) lexBinding(bindingString string) (valueBinding, error) {
	switch bindingString {
	case "", "weak":
		return vbWeak, nil
	case "strong":
		return vbStrong, nil
	case "same":
		return vbSame, nil
	default:
		return 0, parser.error("invalid binding \"%s\"", bindingString)
	}
}

func (parse *configParser) parseMatch() error {
	var kind matchKind
	kindName := parse.p().getAttr("target")
	switch kindName {
	case "pattern":
		kind = MatchQuery
	case "font":
		kind = MatchResult
	case "scan":
		kind = MatchScan
	case "":
		kind = MatchQuery
	default:
		return parse.error("invalid match target \"%s\"", kindName)
	}

	var rule directive
	for _, vstack := range parse.p().values {
		switch vstack.tag {
		case vstackTest:
			r := vstack.u
			rule.tests = append(rule.tests, r.(ruleTest))
			vstack.tag = vstackNone
		case vstackEdit:
			edit := vstack.u.(ruleEdit)
			if kind == MatchScan && edit.object >= FirstCustomObject {
				return fmt.Errorf("<match target=\"scan\"> cannot edit user-defined object \"%s\"", edit.object)
			}
			rule.edits = append(rule.edits, edit)
			vstack.tag = vstackNone
		default:
			return parse.error("invalid match element")
		}
	}
	parse.p().values = nil

	if len(rule.edits)+len(rule.tests) == 0 {
		return parse.error("No <test> nor <edit> elements in <match>")
	}
	maxObj := parse.ruleset.add(rule, kind)
	if parse.config.maxObjects < maxObj {
		parse.config.maxObjects = maxObj
	}
	return nil
}

func revertTests(arr []ruleTest) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
}

func (parser *configParser) parseAlias() error {
	var (
		family, accept, prefer, def *expression
		rule                        directive // we append, then reverse
		last                        = parser.p()
	)
	binding, err := parser.lexBinding(last.getAttr("binding"))
	if err != nil {
		return err
	}

	vals := last.values
	for i := range vals {
		vstack := vals[len(vals)-i-1]
		switch vstack.tag {
		case vstackFamily:
			if family != nil {
				return parser.error("Having multiple <family> in <alias> isn't supported and may not work as expected")
			} else {
				family = vstack.u.(*expression)
			}
		case vstackPrefer:
			prefer = vstack.u.(*expression)
			vstack.tag = vstackNone
		case vstackAccept:
			accept = vstack.u.(*expression)
			vstack.tag = vstackNone
		case vstackDefault:
			def = vstack.u.(*expression)
			vstack.tag = vstackNone
		case vstackTest:
			rule.tests = append(rule.tests, vstack.u.(ruleTest))
			vstack.tag = vstackNone
		default:
			return parser.error("bad alias")
		}
	}
	revertTests(rule.tests)
	last.values = nil

	if family == nil {
		return fmt.Errorf("missing family in alias")
	}

	if prefer == nil && accept == nil && def == nil {
		return nil
	}

	t, err := parser.newTest(MatchQuery, qualAny, FAMILY,
		opWithFlags(opEqual, opFlagIgnoreBlanks), family)
	if err != nil {
		return err
	}
	rule.tests = append(rule.tests, t)

	if prefer != nil {
		edit, err := parser.newEdit(FAMILY, opPrepend, prefer, binding)
		if err != nil {
			return err
		}
		rule.edits = append(rule.edits, edit)
	}
	if accept != nil {
		edit, err := parser.newEdit(FAMILY, opAppend, accept, binding)
		if err != nil {
			return err
		}
		rule.edits = append(rule.edits, edit)
	}
	if def != nil {
		edit, err := parser.newEdit(FAMILY, opAppendLast, def, binding)
		if err != nil {
			return err
		}
		rule.edits = append(rule.edits, edit)
	}
	maxObj := parser.ruleset.add(rule, MatchQuery)
	if parser.config.maxObjects < maxObj {
		parser.config.maxObjects = maxObj
	}
	return nil
}

func (parser *configParser) newTest(kind matchKind, qual uint8,
	object Object, compare opKind, expr *expression) (ruleTest, error) {
	test := ruleTest{kind: kind, qual: qual, op: opKind(compare), expr: expr}

	o := objects[object.String()]
	test.object = object
	var err error
	if o.typeInfo != nil {
		err = parser.typecheckExpr(expr, o.typeInfo)
	}
	if err != nil {
		return test, fmt.Errorf("; for object %s", object)
	}
	return test, nil
}

func (parser *configParser) newEdit(object Object, op opKind, expr *expression, binding valueBinding) (ruleEdit, error) {
	e := ruleEdit{object: object, op: op, expr: expr, binding: binding}
	var err error
	if o := objects[object.String()]; o.typeInfo != nil {
		err = parser.typecheckExpr(expr, o.typeInfo)
	}
	if err != nil {
		return e, fmt.Errorf("; for object %s", object)
	}
	return e, err
}

func (parser *configParser) popExpr() *expression {
	var expr *expression
	vstack := parser.v()
	if vstack == nil {
		return nil
	}
	switch vstack.tag {
	case vstackString, vstackFamily:
		expr = &expression{op: opString, u: vstack.u}
	case vstackName:
		expr = &expression{op: opField, u: vstack.u}
	case vstackConstant:
		expr = &expression{op: opConst, u: vstack.u}
	case vstackPrefer, vstackAccept, vstackDefault:
		expr = vstack.u.(*expression)
		vstack.tag = vstackNone
	case vstackInt:
		expr = &expression{op: opInt, u: vstack.u}
	case vstackDouble:
		expr = &expression{op: opDouble, u: vstack.u}
	case vstackMatrix:
		expr = &expression{op: opMatrix, u: vstack.u}
	case vstackRange:
		expr = &expression{op: opRange, u: vstack.u}
	case vstackBool:
		expr = &expression{op: opBool, u: vstack.u}
	case vstackCharSet:
		expr = &expression{op: opCharSet, u: vstack.u}
	case vstackLangSet:
		expr = &expression{op: opLangSet, u: vstack.u}
	case vstackTest, vstackExpr:
		expr = vstack.u.(*expression)
		vstack.tag = vstackNone
	}
	parser.vstackPop()
	return expr
}

// This builds a tree of binary operations. Note
// that every operator is defined so that if only
// a single operand is contained, the value of the
// whole expression is the value of the operand.
//
// This code reduces in that case to returning that
// operand.
func (parser *configParser) popBinary(op opKind) *expression {
	var expr *expression

	for left := parser.popExpr(); left != nil; left = parser.popExpr() {
		if expr != nil {
			expr = newExprOp(left, expr, op)
		} else {
			expr = left
		}
	}
	return expr
}

func (parser *configParser) pushExpr(tag vstackTag, expr *expression) {
	vstack := parser.createVAndPush()
	vstack.u = expr
	vstack.tag = tag
}

func (parser *configParser) parseBinary(op opKind) {
	expr := parser.popBinary(op)
	if expr != nil {
		parser.pushExpr(vstackExpr, expr)
	}
}

// This builds a a unary operator, it consumes only a single operand
func (parser *configParser) parseUnary(op opKind) {
	operand := parser.popExpr()
	if operand != nil {
		expr := newExprOp(operand, nil, op)
		parser.pushExpr(vstackExpr, expr)
	}
}

func (parser *configParser) parseInt() error {
	last := parser.p()
	if last == nil {
		return nil
	}
	s := last.str.String()
	last.str.Reset()

	d, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("\"%s\": not a valid Int", s)
	}

	vstack := parser.createVAndPush()
	vstack.u = Int(d)
	vstack.tag = vstackInt
	return nil
}

func (parser *configParser) parseFloat() error {
	last := parser.p()
	if last == nil {
		return nil
	}
	s := last.str.String()
	last.str.Reset()

	d, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("\"%s\": not a valid float", s)
	}

	vstack := parser.createVAndPush()
	vstack.u = Float(d)
	vstack.tag = vstackDouble
	return nil
}

func (parser *configParser) parseString(tag vstackTag) {
	last := parser.p()
	if last == nil {
		return
	}
	s := last.str.String()
	last.str.Reset()

	vstack := parser.createVAndPush()
	vstack.u = String(s)
	vstack.tag = tag
}

func (parser *configParser) parseBool() (err error) {
	last := parser.p()
	if last == nil {
		return nil
	}
	s := last.str.String()
	last.str.Reset()

	vstack := parser.createVAndPush()
	vstack.u, err = parser.lexBool(s)
	vstack.tag = vstackBool
	return err
}

func (parser *configParser) parseName() error {
	var kind matchKind
	last := parser.p()

	switch kindString := last.getAttr("target"); kindString {
	case "pattern":
		kind = MatchQuery
	case "font":
		kind = MatchResult
	case "", "default":
		kind = matchDefault
	default:
		return parser.error("invalid name target \"%s\"", kindString)
	}

	if last == nil {
		return nil
	}
	s := last.str.String()
	last.str.Reset()
	object := parser.config.getRegisterObjectType(s)

	vstack := parser.createVAndPush()
	vstack.u = exprName{object: object.object, kind: kind}
	vstack.tag = vstackName
	return nil
}

func (parser *configParser) parseMatrix() error {
	var m exprMatrix

	m.yy = parser.popExpr()
	m.yx = parser.popExpr()
	m.xy = parser.popExpr()
	m.xx = parser.popExpr()

	if m.yy == nil || m.yx == nil || m.xy == nil || m.xx == nil {
		return parser.error("Missing values in matrix element")
	}
	if parser.popExpr() != nil {
		return errors.New("wrong number of matrix elements")
	}

	vstack := parser.createVAndPush()
	vstack.u = m
	vstack.tag = vstackMatrix
	return nil
}

func (parser *configParser) parseRange() error {
	var (
		n     [2]int
		d     [2]float32
		dflag = false
	)
	values := parser.p().values
	if len(values) != 2 {
		return fmt.Errorf("wrong numbers %d of elements in range", len(values))
	}
	for i, vstack := range values {
		switch vstack.tag {
		case vstackInt:
			if dflag {
				d[i] = float32(vstack.u.(Int))
			} else {
				n[i] = int(vstack.u.(Int))
			}
		case vstackDouble:
			if i == 0 && !dflag {
				d[1] = float32(n[1])
			}
			d[i] = float32(vstack.u.(Float))
			dflag = true
		default:
			return errors.New("invalid element in range")
		}
	}
	parser.p().values = nil

	var r Range
	if dflag {
		if d[0] > d[1] {
			return errors.New("invalid range")
		}
		r = Range{Begin: d[0], End: d[1]}
	} else {
		if n[0] > n[1] {
			return errors.New("invalid range")
		}
		r = Range{Begin: float32(n[0]), End: float32(n[1])}
	}
	vstack := parser.createVAndPush()
	vstack.u = r
	vstack.tag = vstackRange
	return nil
}

func (parser *configParser) parseCharSet() error {
	var (
		charset Charset
		n       = 0
	)

	last := parser.p()
	for _, vstack := range last.values {
		switch vstack.tag {
		case vstackInt:
			r := rune(vstack.u.(Int))
			if r > maxCharsetRune {
				return parser.error("invalid character: 0x%04x", r)
			} else {
				charset.AddChar(r)
				n++
			}
		case vstackRange:
			ra := vstack.u.(Range)
			if ra.Begin <= ra.End {
				for r := rune(ra.Begin); r <= rune(ra.End); r++ {
					if r > maxCharsetRune {
						return parser.error("invalid character: 0x%04x", r)
					} else {
						charset.AddChar(r)
						n++
					}
				}
			}
		default:
			return errors.New("invalid element in charset")
		}
	}
	last.values = nil
	if n > 0 {
		vstack := parser.createVAndPush()
		vstack.u = charset
		vstack.tag = vstackCharSet
	}
	return nil
}

func (parser *configParser) parseLangSet() error {
	var (
		langset Langset
		n       = 0
	)

	for _, vstack := range parser.p().values {
		switch vstack.tag {
		case vstackString:
			s := vstack.u.(String)
			langset.add(string(s))
			n++
		default:
			return errors.New("invalid element in langset")
		}
	}
	parser.p().values = nil
	if n > 0 {
		vstack := parser.createVAndPush()
		vstack.u = langset
		vstack.tag = vstackLangSet
	}
	return nil
}

func (parser *configParser) parseFamilies(tag vstackTag) error {
	var expr *expression

	val := parser.p().values
	for i := range val {
		vstack := val[len(val)-1-i]
		if vstack.tag != vstackFamily {
			return parser.error("non-family")
		}
		left := vstack.u.(*expression)
		vstack.tag = vstackNone
		if expr != nil {
			expr = newExprOp(left, expr, opComma)
		} else {
			expr = left
		}
	}
	parser.p().values = nil
	if expr != nil {
		parser.pushExpr(tag, expr)
	}
	return nil
}

func (parser *configParser) parseFamily() {
	last := parser.p()
	if last == nil {
		return
	}
	s := last.str.String()
	last.str.Reset()

	expr := &expression{op: opString, u: String(s)}
	parser.pushExpr(vstackFamily, expr)
}

func (parser *configParser) parseDescription() {
	last := parser.p()
	if last == nil {
		return
	}
	desc := last.str.String()
	last.str.Reset()
	domain := last.getAttr("domain")
	parser.ruleset.domain, parser.ruleset.description = domain, desc
}

func (parser *configParser) parseTest() error {
	var (
		kind    matchKind
		qual    uint8
		compare opKind
		flags   int64
		object  Object
		last    = parser.p()
	)

	switch kindString := last.getAttr("target"); kindString {
	case "pattern":
		kind = MatchQuery
	case "font":
		kind = MatchResult
	case "scan":
		kind = MatchScan
	case "", "default":
		kind = matchDefault
	default:
		return parser.error("invalid test target \"%s\"", kindString)
	}

	switch qualString := last.getAttr("qual"); qualString {
	case "", "any":
		qual = qualAny
	case "all":
		qual = qualAll
	case "first":
		qual = qualFirst
	case "not_first":
		qual = qualNotFirst
	default:
		return parser.error("invalid test qual \"%s\"", qualString)
	}
	name := last.getAttr("name")
	if name == "" {
		return parser.error("missing test name")
	} else {
		ot := parser.config.getRegisterObjectType(name)
		object = ot.object
	}
	compareString := last.getAttr("compare")
	if compareString == "" {
		compare = opEqual
	} else {
		var ok bool
		compare, ok = compareOps[compareString]
		if !ok {
			return parser.error("invalid test compare \"%s\"", compareString)
		}
	}

	if iblanksString := last.getAttr("ignore-blanks"); iblanksString != "" {
		f, err := nameBool(iblanksString)
		if err != nil {
			return parser.error("invalid test ignore-blanks \"%s\"", iblanksString)
		}
		if f != 0 {
			flags |= opFlagIgnoreBlanks
		}
	}
	expr := parser.popBinary(opComma)
	if expr == nil {
		return parser.error("missing test expression")
	}
	if expr.op == opComma {
		return parser.error("Having multiple values in <test> isn't supported and may not work as expected")
	}
	test, err := parser.newTest(kind, qual, object, opWithFlags(compare, flags), expr)

	vstack := parser.createVAndPush()
	vstack.u = test
	vstack.tag = vstackTest
	return err
}

func (parser *configParser) parseEdit() error {
	var (
		mode   opKind
		last   = parser.p()
		object Object
	)

	name := last.getAttr("name")
	if name == "" {
		return parser.error("missing edit name")
	} else {
		ot := parser.config.getRegisterObjectType(name)
		object = ot.object
	}
	modeString := last.getAttr("mode")
	if modeString == "" {
		mode = opAssign
	} else {
		var ok bool
		mode, ok = modeOps[modeString]
		if !ok {
			return parser.error("invalid edit mode \"%s\"", modeString)
		}
	}
	binding, err := parser.lexBinding(last.getAttr("binding"))
	if err != nil {
		return err
	}

	expr := parser.popBinary(opComma)
	if (mode == opDelete || mode == opDeleteAll) && expr != nil {
		return parser.error("Expression doesn't take any effects for delete and delete_all")
	}
	edit, err := parser.newEdit(object, mode, expr, binding)

	vstack := parser.createVAndPush()
	vstack.u = edit
	vstack.tag = vstackEdit
	return err
}

func (parser *configParser) parseAcceptRejectFont(element elemTag) error {
	for _, vstack := range parser.p().values {
		switch vstack.tag {
		case vstackGlob:
			parser.config.globAdd(string(vstack.u.(String)), element == elementAcceptfont)
		case vstackPattern:
			parser.config.patternsAdd(vstack.u.(Pattern), element == elementAcceptfont)
		default:
			return parser.error("bad font selector")
		}
	}
	parser.p().values = nil
	return nil
}

func (parser *configParser) parsePattern() error {
	pattern := NewPattern()

	vals := parser.p().values
	for i := range vals {
		vstack := vals[len(vals)-1-i]
		switch vstack.tag {
		case vstackPattern:
			pattern.append(vstack.u.(Pattern))
		default:
			return parser.error("unknown pattern element")
		}
	}
	parser.p().values = nil

	vstack := parser.createVAndPush()
	vstack.u = pattern
	vstack.tag = vstackPattern
	return nil
}

func (parser *configParser) parsePatelt() error {
	pattern := NewPattern()

	name := parser.p().getAttr("name")
	if name == "" {
		return parser.error("missing pattern element name")
	}
	ot := parser.config.getRegisterObjectType(name)
	for {
		value, err := parser.popValue()
		if err != nil {
			return err
		}
		if value == nil {
			break
		}
		pattern.Add(ot.object, value, true)
	}

	vstack := parser.createVAndPush()
	vstack.u = pattern
	vstack.tag = vstackPattern
	return nil
}

func (parser *configParser) popValue() (Value, error) {
	vstack := parser.v()
	if vstack == nil {
		return nil, nil
	}
	var value Value

	switch vstack.tag {
	case vstackString, vstackInt, vstackDouble, vstackBool,
		vstackCharSet, vstackLangSet, vstackRange:
		value = vstack.u.(Value)
	case vstackConstant:
		if i, ok := nameConstant(vstack.u.(String)); ok {
			value = Int(i)
		}
	default:
		return nil, parser.error("unknown pattern element %d", vstack.tag)
	}
	parser.vstackPop()

	return value, nil
}
