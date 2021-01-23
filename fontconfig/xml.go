package fontconfig

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

// Walks the configuration in 'file' and, if `load` is true, constructs the internal representation
// in 'config'.
func (config *Config) parseConfig(name string) error {
	dest, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}
	filename, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	if isDir(filename) {
		return config.LoadFromDir(filename)
	}

	fi, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("fontconfig: can't open such file %s: %s", filename, err)
	}
	defer fi.Close()

	err = config.parseAndLoadFromMemory(filename, fi)
	return err
}

// compact form of the tag
type elemTag uint8

const (
	FcElementNone elemTag = iota
	FcElementFontconfig
	FcElementMatch
	FcElementAlias
	FcElementDescription

	FcElementPrefer
	FcElementAccept
	FcElementDefault
	FcElementFamily

	FcElementSelectfont
	FcElementAcceptfont
	FcElementRejectfont
	FcElementGlob
	FcElementPattern
	FcElementPatelt

	FcElementTest
	FcElementEdit
	FcElementInt
	FcElementDouble
	FcElementString
	FcElementMatrix
	FcElementRange
	FcElementBool
	FcElementCharSet
	FcElementLangSet
	FcElementName
	FcElementConst
	FcElementOr
	FcElementAnd
	FcElementEq
	FcElementNotEq
	FcElementLess
	FcElementLessEq
	FcElementMore
	FcElementMoreEq
	FcElementContains
	FcElementNotContains
	FcElementPlus
	FcElementMinus
	FcElementTimes
	FcElementDivide
	FcElementNot
	FcElementIf
	FcElementFloor
	FcElementCeil
	FcElementRound
	FcElementTrunc
	FcElementUnknown
)

var fcElementMap = [...]string{
	FcElementFontconfig:  "fontconfig",
	FcElementMatch:       "match",
	FcElementAlias:       "alias",
	FcElementDescription: "description",

	FcElementPrefer:  "prefer",
	FcElementAccept:  "accept",
	FcElementDefault: "default",
	FcElementFamily:  "family",

	FcElementSelectfont: "selectfont",
	FcElementAcceptfont: "acceptfont",
	FcElementRejectfont: "rejectfont",
	FcElementGlob:       "glob",
	FcElementPattern:    "pattern",
	FcElementPatelt:     "patelt",

	FcElementTest:        "test",
	FcElementEdit:        "edit",
	FcElementInt:         "int",
	FcElementDouble:      "double",
	FcElementString:      "string",
	FcElementMatrix:      "matrix",
	FcElementRange:       "range",
	FcElementBool:        "bool",
	FcElementCharSet:     "charset",
	FcElementLangSet:     "langset",
	FcElementName:        "name",
	FcElementConst:       "const",
	FcElementOr:          "or",
	FcElementAnd:         "and",
	FcElementEq:          "eq",
	FcElementNotEq:       "not_eq",
	FcElementLess:        "less",
	FcElementLessEq:      "less_eq",
	FcElementMore:        "more",
	FcElementMoreEq:      "more_eq",
	FcElementContains:    "contains",
	FcElementNotContains: "not_contains",
	FcElementPlus:        "plus",
	FcElementMinus:       "minus",
	FcElementTimes:       "times",
	FcElementDivide:      "divide",
	FcElementNot:         "not",
	FcElementIf:          "if",
	FcElementFloor:       "floor",
	FcElementCeil:        "ceil",
	FcElementRound:       "round",
	FcElementTrunc:       "trunc",
}

var fcElementIgnoreName = [...]string{
	"its:",
}

func elemFromName(name string) elemTag {
	for i, elem := range fcElementMap {
		if name == elem {
			return elemTag(i)
		}
	}
	for _, ignoreName := range fcElementIgnoreName {
		if strings.HasPrefix(name, ignoreName) {
			return FcElementNone
		}
	}
	return FcElementUnknown
}

func (e elemTag) String() string {
	if int(e) >= len(fcElementMap) {
		return fmt.Sprintf("invalid element %d", e)
	}
	return fcElementMap[e]
}

// pStack is one XML containing tag
type pStack struct {
	element elemTag
	attr    []xml.Attr
	str     *bytes.Buffer // inner text content
	values  []vstack
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

	vstackInteger
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
	tag vstackTag
	u   exprNode
}

type configParser struct {
	name string

	pstack []pStack // the top of the stack is at the end of the slice
	// vstack []vstack // idem

	config  *Config
	ruleset *ruleSet
}

func newConfigParser(name string, config *Config) *configParser {
	var parser configParser

	parser.name = name
	parser.config = config
	parser.ruleset = newRuleSet(name)

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
	err := parser.startElement(start.Name.Local, start.Attr)
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

func (parse *configParser) startElement(name string, attr []xml.Attr) error {
	switch name {
	case "dir", "cachedir", "cache", "include", "config", "remap-dir", "reset-dirs", "rescan":
		return errOldSyntax
	}

	element := elemFromName(name)

	if element == FcElementUnknown {
		return parse.error("unknown element %s", name)
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

	// Don't check the attributes for FcElementNone
	if last := parse.p(); last.element != FcElementNone {
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
	case FcElementMatch:
		err = parser.parseMatch()
	case FcElementAlias:
		err = parser.parseAlias()
	case FcElementDescription:
		parser.parseDescription()

	case FcElementPrefer:
		err = parser.parseFamilies(vstackPrefer)
	case FcElementAccept:
		err = parser.parseFamilies(vstackAccept)
	case FcElementDefault:
		err = parser.parseFamilies(vstackDefault)
	case FcElementFamily:
		parser.parseFamily()

	case FcElementTest:
		err = parser.parseTest()
	case FcElementEdit:
		err = parser.parseEdit()

	case FcElementInt:
		err = parser.parseInteger()
	case FcElementDouble:
		err = parser.parseFloat()
	case FcElementString:
		parser.parseString(vstackString)
	case FcElementMatrix:
		err = parser.parseMatrix()
	case FcElementRange:
		err = parser.parseRange()
	case FcElementBool:
		err = parser.parseBool()
	case FcElementCharSet:
		err = parser.parseCharSet()
	case FcElementLangSet:
		err = parser.parseLangSet()
	case FcElementSelectfont, FcElementAcceptfont, FcElementRejectfont:
		err = parser.parseAcceptRejectFont(last.element)
	case FcElementGlob:
		parser.parseString(vstackGlob)
	case FcElementPattern:
		err = parser.parsePattern()
	case FcElementPatelt:
		err = parser.parsePatelt()
	case FcElementName:
		err = parser.parseName()
	case FcElementConst:
		parser.parseString(vstackConstant)
	case FcElementOr:
		parser.parseBinary(FcOpOr)
	case FcElementAnd:
		parser.parseBinary(FcOpAnd)
	case FcElementEq:
		parser.parseBinary(FcOpEqual)
	case FcElementNotEq:
		parser.parseBinary(FcOpNotEqual)
	case FcElementLess:
		parser.parseBinary(FcOpLess)
	case FcElementLessEq:
		parser.parseBinary(FcOpLessEqual)
	case FcElementMore:
		parser.parseBinary(FcOpMore)
	case FcElementMoreEq:
		parser.parseBinary(FcOpMoreEqual)
	case FcElementContains:
		parser.parseBinary(FcOpContains)
	case FcElementNotContains:
		parser.parseBinary(FcOpNotContains)
	case FcElementPlus:
		parser.parseBinary(FcOpPlus)
	case FcElementMinus:
		parser.parseBinary(FcOpMinus)
	case FcElementTimes:
		parser.parseBinary(FcOpTimes)
	case FcElementDivide:
		parser.parseBinary(FcOpDivide)
	case FcElementIf:
		parser.parseBinary(FcOpQuest)
	case FcElementNot:
		parser.parseUnary(FcOpNot)
	case FcElementFloor:
		parser.parseUnary(FcOpFloor)
	case FcElementCeil:
		parser.parseUnary(FcOpCeil)
	case FcElementRound:
		parser.parseUnary(FcOpRound)
	case FcElementTrunc:
		parser.parseUnary(FcOpTrunc)
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

// return true if str starts by ~
func usesHome(str string) bool { return strings.HasPrefix(str, "~") }

func xdgDataHome() string {
	if !homeEnabled {
		return ""
	}
	env := os.Getenv("XDG_DATA_HOME")
	if env != "" {
		return env
	}
	home := FcConfigHome()
	return filepath.Join(home, ".local", "share")
}

func xdgCacheHome() string {
	if !homeEnabled {
		return ""
	}
	env := os.Getenv("XDG_CACHE_HOME")
	if env != "" {
		return env
	}
	home := FcConfigHome()
	return filepath.Join(home, ".cache")
}

func xdgConfigHome() string {
	if !homeEnabled {
		return ""
	}
	env := os.Getenv("XDG_CONFIG_HOME")
	if env != "" {
		return env
	}
	home := FcConfigHome()
	return filepath.Join(home, ".config")
}

func (parse *configParser) getRealPathFromPrefix(path, prefix string, element elemTag) (string, error) {
	var parent string
	switch prefix {
	case "xdg":
		parent := xdgDataHome()
		if parent == "" { // Home directory might be disabled
			return "", nil
		}
	case "default", "cwd":
		// Nothing to do
	case "relative":
		parent = filepath.Dir(parse.name)
		if parent == "." {
			return "", nil
		}

	// #ifndef _WIN32
	// /* For Win32, check this later for dealing with special cases */
	default:
		if !filepath.IsAbs(path) && path[0] != '~' {
			return "", parse.error(
				"Use of ambiguous path in <%s> element. please add prefix=\"cwd\" if current behavior is desired.",
				element)
		}
		// #else
	}

	// TODO: support windows
	//     if  path ==  "CUSTOMFONTDIR"  {
	// 	// FcChar8 *p;
	// 	// path = buffer;
	// 	if (!GetModuleFileName (nil, (LPCH) buffer, sizeof (buffer) - 20)) 	{
	// 	    parse.message ( FcSevereError, "GetModuleFileName failed");
	// 	    return ""
	// 	}
	// 	/*
	// 	 * Must use the multi-byte aware function to search
	// 	 * for backslash because East Asian double-byte code
	// 	 * pages have characters with backslash as the second
	// 	 * byte.
	// 	 */
	// 	p = _mbsrchr (path, '\\');
	// 	if (p) *p = '\0';
	// 	strcat ((char *) path, "\\fonts");
	//     }
	//     else if (strcmp ((const char *) path, "APPSHAREFONTDIR") == 0)
	//     {
	// 	FcChar8 *p;
	// 	path = buffer;
	// 	if (!GetModuleFileName (nil, (LPCH) buffer, sizeof (buffer) - 20))
	// 	{
	// 	    parse.message ( FcSevereError, "GetModuleFileName failed");
	// 	    return nil;
	// 	}
	// 	p = _mbsrchr (path, '\\');
	// 	if (p) *p = '\0';
	// 	strcat ((char *) path, "\\..\\share\\fonts");
	//     }
	//     else if (strcmp ((const char *) path, "WINDOWSFONTDIR") == 0)
	//     {
	// 	int rc;
	// 	path = buffer;
	// 	rc = pGetSystemWindowsDirectory ((LPSTR) buffer, sizeof (buffer) - 20);
	// 	if (rc == 0 || rc > sizeof (buffer) - 20)
	// 	{
	// 	    parse.message ( FcSevereError, "GetSystemWindowsDirectory failed");
	// 	    return nil;
	// 	}
	// 	if (path [strlen ((const char *) path) - 1] != '\\')
	// 	    strcat ((char *) path, "\\");
	// 	strcat ((char *) path, "fonts");
	//     }
	//     else
	//     {
	// 	if (!prefix)
	// 	{
	// 	    if (!FcStrIsAbsoluteFilename (path) && path[0] != '~')
	// 		parse.message ( FcSevereWarning, "Use of ambiguous path in <%s> element. please add prefix=\"cwd\" if current behavior is desired.", FcElementReverseMap (parse.pstack.element));
	// 	}
	//     }
	// #endif

	if parent != "" {
		return filepath.Join(parent, path), nil
	}
	return path, nil
}

func (parser *configParser) lexBool(bool_ string) (Bool, error) {
	result, err := nameBool(bool_)
	if err != nil {
		return 0, parser.error("\"%s\" is not known boolean", bool_)
	}
	return result, nil
}

func (parser *configParser) lexBinding(bindingString string) (FcValueBinding, error) {
	switch bindingString {
	case "", "weak":
		return FcValueBindingWeak, nil
	case "strong":
		return FcValueBindingStrong, nil
	case "same":
		return FcValueBindingSame, nil
	default:
		return 0, parser.error("invalid binding \"%s\"", bindingString)
	}
}

func isDir(s string) bool {
	f, err := os.Stat(s)
	if err != nil {
		return false
	}
	return f.IsDir()
}

func isFile(s string) bool {
	f, err := os.Stat(s)
	if err != nil {
		return false
	}
	return !f.IsDir()
}

func isLink(s string) bool {
	f, err := os.Stat(s)
	if err != nil {
		return false
	}
	return f.Mode() == os.ModeSymlink
}

// return true on success
func rename(old, new string) bool { return os.Rename(old, new) == nil }

// return true on success
func symlink(old, new string) bool { return os.Symlink(old, new) == nil }

var (
	userdir, userconf string
	userValuesLock    sync.Mutex
)

func getUserdir(s string) string {
	userValuesLock.Lock()
	defer userValuesLock.Unlock()
	if userdir == "" {
		userdir = s
	}
	return userdir
}

func getUserconf(s string) string {
	userValuesLock.Lock()
	defer userValuesLock.Unlock()
	if userconf == "" {
		userconf = s
	}
	return userconf
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
		family, accept, prefer, def *FcExpr
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
				family = vstack.u.(*FcExpr)
			}
		case vstackPrefer:
			prefer = vstack.u.(*FcExpr)
			vstack.tag = vstackNone
		case vstackAccept:
			accept = vstack.u.(*FcExpr)
			vstack.tag = vstackNone
		case vstackDefault:
			def = vstack.u.(*FcExpr)
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

	t, err := parser.newTest(MatchQuery, FcQualAny, FC_FAMILY,
		opWithFlags(FcOpEqual, FcOpFlagIgnoreBlanks), family)
	if err != nil {
		return err
	}
	rule.tests = append(rule.tests, t)

	if prefer != nil {
		edit, err := parser.newEdit(FC_FAMILY, FcOpPrepend, prefer, binding)
		if err != nil {
			return err
		}
		rule.edits = append(rule.edits, edit)
	}
	if accept != nil {
		edit, err := parser.newEdit(FC_FAMILY, FcOpAppend, accept, binding)
		if err != nil {
			return err
		}
		rule.edits = append(rule.edits, edit)
	}
	if def != nil {
		edit, err := parser.newEdit(FC_FAMILY, FcOpAppendLast, def, binding)
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
	object Object, compare FcOp, expr *FcExpr) (ruleTest, error) {
	test := ruleTest{kind: kind, qual: qual, op: FcOp(compare), expr: expr}
	o := objects[object.String()]
	test.object = o.object
	var err error
	if o.parser != nil {
		err = parser.typecheckExpr(expr, o.parser)
	}
	return test, err
}

func (parser *configParser) newEdit(object Object, op FcOp, expr *FcExpr, binding FcValueBinding) (ruleEdit, error) {
	e := ruleEdit{object: object, op: op, expr: expr, binding: binding}
	var err error
	if o := objects[object.String()]; o.parser != nil {
		err = parser.typecheckExpr(expr, o.parser)
	}
	return e, err
}

func (parser *configParser) popExpr() *FcExpr {
	var expr *FcExpr
	vstack := parser.v()
	if vstack == nil {
		return nil
	}
	switch vstack.tag {
	case vstackString, vstackFamily:
		expr = &FcExpr{op: FcOpString, u: vstack.u}
	case vstackName:
		expr = &FcExpr{op: FcOpField, u: vstack.u}
	case vstackConstant:
		expr = &FcExpr{op: FcOpConst, u: vstack.u}
	case vstackPrefer, vstackAccept, vstackDefault:
		expr = vstack.u.(*FcExpr)
		vstack.tag = vstackNone
	case vstackInteger:
		expr = &FcExpr{op: FcOpInteger, u: vstack.u}
	case vstackDouble:
		expr = &FcExpr{op: FcOpDouble, u: vstack.u}
	case vstackMatrix:
		expr = &FcExpr{op: FcOpMatrix, u: vstack.u}
	case vstackRange:
		expr = &FcExpr{op: FcOpRange, u: vstack.u}
	case vstackBool:
		expr = &FcExpr{op: FcOpBool, u: vstack.u}
	case vstackCharSet:
		expr = &FcExpr{op: FcOpCharSet, u: vstack.u}
	case vstackLangSet:
		expr = &FcExpr{op: FcOpLangSet, u: vstack.u}
	case vstackTest, vstackExpr:
		expr = vstack.u.(*FcExpr)
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
func (parser *configParser) popBinary(op FcOp) *FcExpr {
	var expr *FcExpr

	for left := parser.popExpr(); left != nil; left = parser.popExpr() {
		if expr != nil {
			expr = newExprOp(left, expr, op)
		} else {
			expr = left
		}
	}
	return expr
}

func (parser *configParser) pushExpr(tag vstackTag, expr *FcExpr) {
	vstack := parser.createVAndPush()
	vstack.u = expr
	vstack.tag = tag
}

func (parser *configParser) parseBinary(op FcOp) {
	expr := parser.popBinary(op)
	if expr != nil {
		parser.pushExpr(vstackExpr, expr)
	}
}

// This builds a a unary operator, it consumes only a single operand
func (parser *configParser) parseUnary(op FcOp) {
	operand := parser.popExpr()
	if operand != nil {
		expr := newExprOp(operand, nil, op)
		parser.pushExpr(vstackExpr, expr)
	}
}

func (parser *configParser) parseInteger() error {
	last := parser.p()
	if last == nil {
		return nil
	}
	s := last.str.String()
	last.str.Reset()

	d, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("\"%s\": not a valid integer", s)
	}

	vstack := parser.createVAndPush()
	vstack.u = Int(d)
	vstack.tag = vstackInteger
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
		kind = MatchDefault
	default:
		return parser.error("invalid name target \"%s\"", kindString)
	}

	if last == nil {
		return nil
	}
	s := last.str.String()
	last.str.Reset()
	object := getRegisterObjectType(s)

	vstack := parser.createVAndPush()
	vstack.u = FcExprName{object: object.object, kind: kind}
	vstack.tag = vstackName
	return nil
}

func (parser *configParser) parseMatrix() error {
	var m FcExprMatrix

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
		d     [2]float64
		dflag = false
	)
	values := parser.p().values
	if len(values) != 2 {
		return fmt.Errorf("wrong numbers %d of elements in range", len(values))
	}
	for i, vstack := range values {
		switch vstack.tag {
		case vstackInteger:
			if dflag {
				d[i] = float64(vstack.u.(Int))
			} else {
				n[i] = int(vstack.u.(Int))
			}
		case vstackDouble:
			if i == 0 && !dflag {
				d[1] = float64(n[1])
			}
			d[i] = float64(vstack.u.(Float))
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
		r = Range{Begin: float64(n[0]), End: float64(n[1])}
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
		case vstackInteger:
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
	var expr *FcExpr

	val := parser.p().values
	for i := range val {
		vstack := val[len(val)-1-i]
		if vstack.tag != vstackFamily {
			return parser.error("non-family")
		}
		left := vstack.u.(*FcExpr)
		vstack.tag = vstackNone
		if expr != nil {
			expr = newExprOp(left, expr, FcOpComma)
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

	expr := &FcExpr{op: FcOpString, u: String(s)}
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
		compare FcOp
		flags   int
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
		kind = MatchDefault
	default:
		return parser.error("invalid test target \"%s\"", kindString)
	}

	switch qualString := last.getAttr("qual"); qualString {
	case "", "any":
		qual = FcQualAny
	case "all":
		qual = FcQualAll
	case "first":
		qual = FcQualFirst
	case "not_first":
		qual = FcQualNotFirst
	default:
		return parser.error("invalid test qual \"%s\"", qualString)
	}
	name := last.getAttr("name")
	if name == "" {
		return parser.error("missing test name")
	} else {
		ot := getRegisterObjectType(name)
		object = ot.object
	}
	compareString := last.getAttr("compare")
	if compareString == "" {
		compare = FcOpEqual
	} else {
		var ok bool
		compare, ok = fcCompareOps[compareString]
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
			flags |= FcOpFlagIgnoreBlanks
		}
	}
	expr := parser.popBinary(FcOpComma)
	if expr == nil {
		return parser.error("missing test expression")
	}
	if expr.op == FcOpComma {
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
		mode   FcOp
		last   = parser.p()
		object Object
	)

	name := last.getAttr("name")
	if name == "" {
		return parser.error("missing edit name")
	} else {
		ot := getRegisterObjectType(name)
		object = ot.object
	}
	modeString := last.getAttr("mode")
	if modeString == "" {
		mode = FcOpAssign
	} else {
		var ok bool
		mode, ok = fcModeOps[modeString]
		if !ok {
			return parser.error("invalid edit mode \"%s\"", modeString)
		}
	}
	binding, err := parser.lexBinding(last.getAttr("binding"))
	if err != nil {
		return err
	}

	expr := parser.popBinary(FcOpComma)
	if (mode == FcOpDelete || mode == FcOpDeleteAll) && expr != nil {
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
			parser.config.globAdd(string(vstack.u.(String)), element == FcElementAcceptfont)
		case vstackPattern:
			parser.config.patternsAdd(vstack.u.(Pattern), element == FcElementAcceptfont)
		default:
			return parser.error("bad font selector")
		}
	}
	parser.p().values = nil
	return nil
}

func (parser *configParser) parsePattern() error {
	pattern := NewPattern()

	//  TODO: fix this if the order matter
	for _, vstack := range parser.p().values {
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
	ot := getRegisterObjectType(name)
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
	case vstackString, vstackInteger, vstackDouble, vstackBool,
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
