package fontconfig

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Walks the configuration in `buffer` and constructs the internal representation
// in `config`. Any includes files referenced from within `memory` will be loaded
// and parsed.
func (config *Config) ParseAndLoadFromMemory(buffer []byte) error {
	return config.parseAndLoadFromMemory("memory", buffer, true)
}

func (config *Config) parseAndLoadFromMemory(filename string, buffer []byte, load bool) error {
	if debugMode {
		fmt.Printf("\tProcessing config file from %s, load: %v\n", filename, load)
	}

	parser := newConfigParser(filename, config, load)

	err := xml.Unmarshal(buffer, parser)
	if err != nil {
		return fmt.Errorf("cannot process config file from %s: %s", filename, err)
	}

	if load {
		config.subst = append(config.subst, parser.ruleset)
	}
	// config.rulesetList = append(config.rulesetList, parser.ruleset)

	return nil
}

func (config *Config) parseAndLoadDir(dir string, load bool) error {
	d, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("fontconfig: cannot open config dir %s : %s", dir, err)
	}

	if debugMode {
		fmt.Printf("\tScanning config dir %s\n", dir)
	}

	if load {
		err = config.addConfigDir(dir)
		if err != nil {
			return err
		}
	}

	var files []string
	const tail = ".conf"
	for _, e := range d {
		/// Add all files of the form [0-9]*.conf
		if name := e.Name(); name != "" && '0' <= name[0] && name[0] <= '9' && strings.HasSuffix(name, tail) {
			file := dir + "/" + name
			files = append(files, file)
		}
	}

	sort.Strings(files)

	for _, file := range files {
		err := config.parseConfig(file, load)
		if err != nil {
			return err
		}
	}

	return nil
}

// Walks the configuration in 'file' and, if `load` is true, constructs the internal representation
// in 'config'.  Any include files referenced from within 'file' will be loaded
// and parsed.
func (config *Config) parseConfig(name string, load bool) error {
	// TODO: support windows
	// #ifdef _WIN32
	//     if (!pGetSystemWindowsDirectory)
	//     {
	//         HMODULE hk32 = GetModuleHandleA("kernel32.dll");
	//         if (!(pGetSystemWindowsDirectory = (pfnGetSystemWindowsDirectory) GetProcAddress(hk32, "GetSystemWindowsDirectoryA")))
	//             pGetSystemWindowsDirectory = (pfnGetSystemWindowsDirectory) GetWindowsDirectory;
	//     }
	//     if (!pSHGetFolderPathA)
	//     {
	//         HMODULE hSh = LoadLibraryA("shfolder.dll");
	//         /* the check is done later, because there is no provided fallback */
	//         if (hSh)
	//             pSHGetFolderPathA = (pfnSHGetFolderPathA) GetProcAddress(hSh, "SHGetFolderPathA");
	//     }
	// #endif

	filename := config.GetFilename(name)
	if filename == "" {
		return fmt.Errorf("fontconfig: no such file: %s", name)
	}

	realfilename := realFilename(filename)

	if config.availConfigFiles[realfilename] {
		return nil
	}

	if load {
		config.configFiles[filename] = true
	}
	config.availConfigFiles[realfilename] = true

	if isDir(realfilename) {
		return config.parseAndLoadDir(realfilename, load)
	}

	content, err := ioutil.ReadFile(realfilename)
	if err != nil {
		return fmt.Errorf("fontconfig: can't open such file %s: %s", realfilename, err)
	}

	err = config.parseAndLoadFromMemory(filename, content, load)
	return err
}

// compact form of the tag
type elemTag uint8

const (
	FcElementNone elemTag = iota
	FcElementFontconfig
	FcElementDir
	FcElementCacheDir
	FcElementCache
	FcElementInclude
	FcElementConfig
	FcElementMatch
	FcElementAlias
	FcElementDescription
	FcElementRemapDir
	FcElementResetDirs

	FcElementRescan

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
	FcElementDir:         "dir",
	FcElementCacheDir:    "cachedir",
	FcElementCache:       "cache",
	FcElementInclude:     "include",
	FcElementConfig:      "config",
	FcElementMatch:       "match",
	FcElementAlias:       "alias",
	FcElementDescription: "description",
	FcElementRemapDir:    "remap-dir",
	FcElementResetDirs:   "reset-dirs",

	FcElementRescan: "rescan",

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
	name     string
	scanOnly bool

	pstack []pStack // the top of the stack is at the end of the slice
	// vstack []vstack // idem

	config  *Config
	ruleset *ruleSet
}

func newConfigParser(name string, config *Config, enabled bool) *configParser {
	var parser configParser

	parser.name = name
	parser.config = config
	parser.ruleset = newRuleSet(name)
	parser.scanOnly = !enabled
	parser.ruleset.enabled = enabled

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
	case FcElementDir:
		err = parser.parseDir()
	case FcElementCacheDir:
		err = parser.parseCacheDir()
	case FcElementCache:
		last.str.Reset() // discard this data; no longer used
	case FcElementInclude:
		err = parser.parseInclude()
	case FcElementMatch:
		err = parser.parseMatch()
	case FcElementAlias:
		err = parser.parseAlias()
	case FcElementDescription:
		parser.parseDescription()
	case FcElementRemapDir:
		err = parser.parseRemapDir()
	case FcElementResetDirs:
		parser.parseResetDirs()

	case FcElementRescan:
		err = parser.parseRescan()

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

func (parse *configParser) parseDir() error {
	var s string
	last := parse.p()
	if last != nil {
		s = last.str.String()
		last.str.Reset()
	}
	if len(s) == 0 {
		return parse.error("empty font directory name ignored")
	}
	attr := last.getAttr("prefix")
	salt := last.getAttr("salt")
	prefix, err := parse.getRealPathFromPrefix(s, attr, last.element)
	if err != nil {
		return err
	}
	if prefix == "" {
		// nop
	} else if !parse.scanOnly && (!usesHome(prefix) || FcConfigHome() != "") {
		if err := parse.config.addFontDir(prefix, "", salt); err != nil {
			return fmt.Errorf("fontconfig: cannot add directory %s: %s", prefix, err)
		}
	}
	return nil
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

func (parse *configParser) parseCacheDir() error {
	var prefix, data string
	last := parse.p()
	if last != nil {
		data = last.str.String()
		last.str.Reset()
	}
	if data == "" {
		return parse.error("empty cache directory name ignored")
	}

	if attr := last.getAttr("prefix"); attr != "" && attr == "xdg" {
		prefix = xdgCacheHome()
		// home directory might be disabled.: simply ignore this element.
		if prefix == "" {
			return nil
		}
	}
	if prefix != "" {
		data = filepath.Join(prefix, data)
	}
	// TODO: support windows
	// #ifdef _WIN32
	//     else if (data[0] == '/' && fontconfig_instprefix[0] != '\0')   {
	// 	// size_t plen = strlen ((const char *)fontconfig_instprefix);
	// 	// size_t dlen = strlen ((const char *)data);

	// 	prefix = malloc (plen + 1 + dlen + 1);
	// 	// strcpy ((char *) prefix, (char *) fontconfig_instprefix);
	// 	prefix[plen] = FC_DIR_SEPARATOR;
	// 	memcpy (&prefix[plen + 1], data, dlen);
	// 	prefix[plen + 1 + dlen] = 0;
	// 	FcStrFree (data);
	// 	data = prefix;
	//     }  else if data == "WINDOWSTEMPDIR_FONTCONFIG_CACHE" {
	// 	int rc;

	// 	FcStrFree (data);
	// 	data = malloc (1000);
	// 	rc = GetTempPath (800, (LPSTR) data);
	// 	if (rc == 0 || rc > 800) {
	// 	    parse.message ( FcSevereError, "GetTempPath failed");
	// 	    goto bail;
	// 	}
	// 	if (data [strlen ((const char *) data) - 1] != '\\'){
	// 	    strcat ((char *) data, "\\");}
	// 	strcat ((char *) data, "fontconfig\\cache");
	//     }   else if  data ==  "LOCAL_APPDATA_FONTCONFIG_CACHE"  {
	// 	char szFPath[MAX_PATH + 1];
	// 	size_t len;

	// 	if (!(pSHGetFolderPathA && SUCCEEDED(pSHGetFolderPathA(nil, /* CSIDL_LOCAL_APPDATA */ 28, nil, 0, szFPath)))){
	// 	    return errors.New("SHGetFolderPathA failed");
	// 	}
	// 	strncat(szFPath, "\\fontconfig\\cache", MAX_PATH - 1 - strlen(szFPath));
	// 	len = strlen(szFPath) + 1;
	// 	FcStrFree (data);
	// 	data = malloc(len);
	// 	strncpy((char *) data, szFPath, len);
	//     }
	// #endif
	if len(data) == 0 {
		return parse.error("empty cache directory name ignored")
	}
	if !parse.scanOnly && (!usesHome(data) || FcConfigHome() != "") {
		err := parse.config.addCacheDir(data)
		if err != nil {
			return fmt.Errorf("fontconfig: cannot add cache directory %s: %s", data, err)
		}
	}
	return nil
}

func (parser *configParser) lexBool(bool_ string) (FcBool, error) {
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

func (parse *configParser) parseInclude() error {
	var (
		ignoreMissing, deprecated bool
		prefix, userdir, userconf string
		last                      = parse.p()
	)
	if last == nil {
		return nil
	}

	s := last.str.String()
	last.str.Reset()
	attr := last.getAttr("ignore_missing")
	if attr != "" {
		b, err := parse.lexBool(attr)
		if err != nil {
			return err
		}
		ignoreMissing = b == FcTrue
	}
	attr = last.getAttr("deprecated")
	if attr != "" {
		b, err := parse.lexBool(attr)
		if err != nil {
			return err
		}
		deprecated = b == FcTrue
	}
	attr = last.getAttr("prefix")
	if attr == "xdg" {
		prefix = xdgConfigHome()
		// home directory might be disabled: simply ignore this element.
		if prefix == "" {
			return nil
		}
	}
	if prefix != "" {
		s = filepath.Join(prefix, s)
		if isDir(s) {
			userdir = getUserdir(s)
		} else if isFile(s) {
			userconf = getUserconf(s)
		} else {
			/* No config dir nor file on the XDG directory spec compliant place
			 * so need to guess what it is supposed to be.
			 */
			if strings.Index(s, "conf.d") != -1 {
				userdir = getUserdir(s)
			} else {
				userconf = getUserconf(s)
			}
		}
	}

	// flush the ruleset into the queue
	ruleset := parse.ruleset
	parse.ruleset = newRuleSet(ruleset.name)
	parse.ruleset.enabled = ruleset.enabled
	parse.ruleset.domain, parse.ruleset.description = ruleset.domain, ruleset.description
	parse.config.subst = append(parse.config.subst, ruleset)

	err := parse.config.parseConfig(s, !parse.scanOnly)
	if err != nil && !ignoreMissing {
		return err
	}

	if runtime.GOOS != "windows" {
		var warnConf, warnConfd bool
		filename := parse.config.GetFilename(s)
		os.Stat(filename)
		if deprecated == true && filename != "" && userdir != "" && !isLink(filename) {
			if isDir(filename) {
				parent := filepath.Dir(userdir)
				if !isDir(parent) {
					_ = os.Mkdir(parent, 0755)
				}
				if isDir(userdir) || !rename(filename, userdir) || !symlink(userdir, filename) {
					if !warnConfd {
						return parse.error("reading configurations from %s is deprecated. please move it to %s manually", s, userdir)
					}
				}
			} else {
				parent := filepath.Dir(userconf)
				if !isDir(parent) {
					_ = os.Mkdir(parent, 0755)
				}
				if isFile(userconf) || !rename(filename, userconf) || !symlink(userconf, filename) {
					if !warnConf {
						return parse.error("reading configurations from %s is deprecated. please move it to %s manually", s, userconf)
					}
				}
			}
		}

	}

	return nil
}

func (parse *configParser) parseMatch() error {
	var kind matchKind
	kindName := parse.p().getAttr("target")
	switch kindName {
	case "pattern":
		kind = FcMatchPattern
	case "font":
		kind = FcMatchFont
	case "scan":
		kind = FcMatchScan
	case "":
		kind = FcMatchPattern
	default:
		return parse.error("invalid match target \"%s\"", kindName)
	}

	var rules []rule
	for _, vstack := range parse.p().values {
		switch vstack.tag {
		case vstackTest:
			r := vstack.u
			rules = append(rules, r.(ruleTest))
			vstack.tag = vstackNone
		case vstackEdit:
			edit := vstack.u.(ruleEdit)
			if kind == FcMatchScan && edit.object >= FirstCustomObject {
				return fmt.Errorf("<match target=\"scan\"> cannot edit user-defined object \"%s\"", edit.object)
			}
			rules = append(rules, edit)
			vstack.tag = vstackNone
		default:
			return parse.error("invalid match element")
		}
	}
	parse.p().values = nil

	if len(rules) == 0 {
		return parse.error("No <test> nor <edit> elements in <match>")
	}
	n := parse.ruleset.add(rules, kind)
	if parse.config.maxObjects < n {
		parse.config.maxObjects = n
	}
	return nil
}

func (parser *configParser) parseAlias() error {
	var (
		family, accept, prefer, def *FcExpr
		rules                       []rule // we append, then reverse
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
			rules = append(rules, vstack.u.(ruleTest))
			vstack.tag = vstackNone
		default:
			return parser.error("bad alias")
		}
	}
	revertRules(rules)
	last.values = nil

	if family == nil {
		return fmt.Errorf("missing family in alias")
	}

	if prefer == nil && accept == nil && def == nil {
		return nil
	}

	t, err := parser.newTest(FcMatchPattern, FcQualAny, FC_FAMILY,
		opWithFlags(FcOpEqual, FcOpFlagIgnoreBlanks), family)
	if err != nil {
		return err
	}
	rules = append(rules, t)

	if prefer != nil {
		edit, err := parser.newEdit(FC_FAMILY, FcOpPrepend, prefer, binding)
		if err != nil {
			return err
		}
		rules = append(rules, edit)
	}
	if accept != nil {
		edit, err := parser.newEdit(FC_FAMILY, FcOpAppend, accept, binding)
		if err != nil {
			return err
		}
		rules = append(rules, edit)
	}
	if def != nil {
		edit, err := parser.newEdit(FC_FAMILY, FcOpAppendLast, def, binding)
		if err != nil {
			return err
		}
		rules = append(rules, edit)
	}
	n := parser.ruleset.add(rules, FcMatchPattern)
	if parser.config.maxObjects < n {
		parser.config.maxObjects = n
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
		kind = FcMatchPattern
	case "font":
		kind = FcMatchFont
	case "", "default":
		kind = FcMatchDefault
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
		langset LangSet
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

func (parser *configParser) parseRemapDir() error {
	last := parser.p()
	var data string
	if last != nil {
		data = last.str.String()
		last.str.Reset()
	}

	if data == "" {
		return parser.error("empty font directory name for remap ignored")
	}
	path := last.getAttr("as-path")
	if path == "" {
		return parser.error("Missing as-path in remap-dir")
	}
	attr := last.getAttr("prefix")
	salt := last.getAttr("salt")
	prefix, err := parser.getRealPathFromPrefix(data, attr, last.element)
	if err != nil {
		return err
	}
	fmt.Println(attr, prefix)
	if prefix == "" {
		/* nop */
	} else if !parser.scanOnly && (!usesHome(prefix) || FcConfigHome() != "") {
		if err := parser.config.addFontDir(prefix, path, salt); err != nil {
			return fmt.Errorf("fontconfig: cannot create remap data for %s as %s: %s", prefix, path, err)
		}
	}
	return nil
}

func (parser *configParser) parseResetDirs() {
	if !parser.scanOnly {
		parser.config.fontDirs.reset()
	}
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
		kind = FcMatchPattern
	case "font":
		kind = FcMatchFont
	case "scan":
		kind = FcMatchScan
	case "", "default":
		kind = FcMatchDefault
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

func (parser *configParser) parseRescan() error {
	for _, v := range parser.p().values {
		if v.tag != vstackInteger {
			return parser.error("non-integer rescan")
		} else {
			parser.config.rescanInterval = int(v.u.(Int))
		}
	}
	return nil
}

func (parser *configParser) parseAcceptRejectFont(element elemTag) error {
	for _, vstack := range parser.p().values {
		switch vstack.tag {
		case vstackGlob:
			if !parser.scanOnly {
				parser.config.globAdd(string(vstack.u.(String)), element == FcElementAcceptfont)
			}
		case vstackPattern:
			if !parser.scanOnly {
				parser.config.patternsAdd(vstack.u.(Pattern), element == FcElementAcceptfont)
			}
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

func (parser *configParser) popValue() (FcValue, error) {
	vstack := parser.v()
	if vstack == nil {
		return nil, nil
	}
	var value FcValue

	switch vstack.tag {
	case vstackString, vstackInteger, vstackDouble, vstackBool,
		vstackCharSet, vstackLangSet, vstackRange:
		value = vstack.u.(FcValue)
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
