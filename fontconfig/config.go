package fontconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ported from fontconfig/src/fccfg.c Copyright © 2000 Keith Packard

const (
	FcQualAny uint8 = iota
	FcQualAll
	FcQualFirst
	FcQualNotFirst
)

type ruleTest struct {
	kind   matchKind
	qual   uint8
	object Object
	op     FcOp
	expr   *FcExpr
}

// String returns a human friendly representation of a Test
func (test ruleTest) String() string {
	out := "test: "
	switch test.kind {
	case FcMatchPattern:
		out += "pattern "
	case FcMatchFont:
		out += "font "
	case FcMatchScan:
		out += "scan "
	case matchKindEnd: // shouldn't be reached
		return out
	}
	switch test.qual {
	case FcQualAny:
		out += "any "
	case FcQualAll:
		out += "all "
	case FcQualFirst:
		out += "first "
	case FcQualNotFirst:
		out += "not_first "
	}
	out += fmt.Sprintf("%s %s %s", test.object, test.op, test.expr)
	return out
}

// Check the tests to see if they all match the pattern
// exit = 0 to keep, = 1 to break the inner loop, = 2 to return
func (r ruleTest) match(selectedKind matchKind, p, pPat Pattern, data familyTable,
	valuePos []int, elt []ValueList, tst []*ruleTest) (table *familyTable, exit int) {
	m := p
	table = &data
	if selectedKind == FcMatchFont && r.kind == FcMatchPattern {
		m = pPat
		table = nil
	}
	e := m[r.object]
	object := r.object
	// different 'kind' won't be the target of edit
	if elt[object] == nil && selectedKind == r.kind {
		elt[object] = e
		tst[object] = &r
	}
	// If there's no such field in the font, then FcQualAll matches for FcQualAny does not
	if e == nil {
		if r.qual == FcQualAll {
			valuePos[object] = -1
			return table, 0
		} else {
			return table, 1
		}
	}
	// Check to see if there is a match, mark the location to apply match-relative edits
	vlIndex := matchValueList(m, pPat, selectedKind, r, e, table)
	// different 'kind' won't be the target of edit
	if valuePos[object] == -1 && selectedKind == r.kind && vlIndex != -1 {
		valuePos[object] = vlIndex
	}
	if vlIndex == -1 || (r.qual == FcQualFirst && vlIndex != 0) ||
		(r.qual == FcQualNotFirst && vlIndex == 0) {
		return table, 2
	}
	return table, 0
}

type ruleEdit struct {
	object  Object
	op      FcOp
	expr    *FcExpr
	binding FcValueBinding
}

func (edit ruleEdit) String() string {
	return fmt.Sprintf("edit: %s %s %s", edit.object, edit.op, edit.expr)
}

func (r ruleEdit) edit(selectedKind matchKind, p, pPat Pattern, table *familyTable,
	valuePos []int, elt []ValueList, tst []*ruleTest) {
	object := r.object

	// Evaluate the list of expressions
	l := r.expr.FcConfigValues(p, pPat, selectedKind, r.binding)
	if tst[object] != nil && (tst[object].kind == FcMatchFont || selectedKind == FcMatchPattern) {
		elt[object] = p[tst[object].object]
	}

	switch r.op.getOp() {
	case FcOpAssign:
		// If there was a test, then replace the matched value with the newList list of values
		if valuePos[object] != -1 {
			thisValue := valuePos[object]

			// Append the newList list of values after the current value
			elt[object].insert(thisValue, true, l, r.object, table)

			//  Delete the marked value
			if thisValue != -1 {
				elt[object].del(thisValue, object, table)
			}

			// Adjust a pointer into the value list to ensure future edits occur at the same place
			break
		}
		fallthrough
	case FcOpAssignReplace:
		// Delete all of the values and insert the newList set
		p.FcConfigPatternDel(r.object, table)
		p.FcConfigPatternAdd(r.object, l, true, table)
		// Adjust a pointer into the value list as they no longer point to anything valid
		valuePos[object] = -1
	case FcOpPrepend:
		if valuePos[object] != -1 {
			elt[object].insert(valuePos[object], false, l, r.object, table)
			break
		}
		fallthrough
	case FcOpPrependFirst:
		p.FcConfigPatternAdd(r.object, l, false, table)
	case FcOpAppend:
		if valuePos[object] != -1 {
			elt[object].insert(valuePos[object], true, l, r.object, table)
			break
		}
		fallthrough
	case FcOpAppendLast:
		p.FcConfigPatternAdd(r.object, l, true, table)
	case FcOpDelete:
		if valuePos[object] != -1 {
			elt[object].del(valuePos[object], object, table)
			break
		}
		fallthrough
	case FcOpDeleteAll:
		p.FcConfigPatternDel(r.object, table)
	}
	// Now go through the pattern and eliminate any properties without data
	p.canon(r.object)
}

type rule interface {
	fmt.Stringer
	isRule()
}

func (ruleTest) isRule() {}
func (ruleEdit) isRule() {}

func revertRules(arr []rule) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
}

type ruleSet struct {
	name        string
	description string
	domain      string
	enabled     bool
	subst       [matchKindEnd][][]rule
}

func newRuleSet(name string) *ruleSet {
	var ret ruleSet
	ret.name = name
	return &ret
}

func (rs *ruleSet) String() string {
	out := "RuleSet " + rs.name + "\n"
	for i, v := range rs.subst {
		if len(v) == 0 {
			continue
		}
		out += fmt.Sprintf("\tkind %s:\n", matchKind(i))
		for _, l := range v {
			for _, r := range l {
				out += fmt.Sprintf("\t\t%s\n", r)
			}
			fmt.Println()
		}
	}
	return out
}

func (rs *ruleSet) add(rules []rule, kind matchKind) int {
	rs.subst[kind] = append(rs.subst[kind], rules)

	var n Object
	for _, r := range rules {
		switch r := r.(type) {
		case ruleTest:
			if r.kind == FcMatchDefault {
				r.kind = kind
			}
			if n < r.object {
				n = r.object
			}
		case ruleEdit:
			if n < r.object {
				n = r.object
			}
		}
	}

	if debugMode {
		fmt.Printf("Add Rule(kind:%d, name: %s) ", kind, rs.name)
		fmt.Println(rules)
	}

	ret := n - FirstCustomObject
	if ret < 0 {
		return 0
	}
	return int(ret)
}

type Config struct {
	// File names loaded from the configuration -- saved here as the
	// cache file must be consulted before the directories are scanned,
	// and those directives may occur in any order
	configDirs    strSet // directories to scan for fonts
	configMapDirs strSet // mapped names to generate cache entries
	// List of directories containing fonts,
	// built by recursively scanning the set
	// of configured directories
	fontDirs  strSet
	cacheDirs strSet // List of directories containing cache files.
	// Names of all of the configuration files used to create this configuration
	configFiles strSet /* config files loaded */

	// Substitution instructions for patterns and fonts;
	// maxObjects is used to allocate appropriate intermediate storage
	// for performing a whole set of substitutions
	//
	// 0.. substitutions for patterns
	// 1.. substitutions for fonts
	// 2.. substitutions for scanned fonts
	subst      []*ruleSet
	maxObjects int /* maximum number of tests in all substs */

	// List of patterns used to control font file selection
	acceptGlobs    strSet
	rejectGlobs    strSet
	acceptPatterns FontSet
	rejectPatterns FontSet

	// The set of fonts loaded from the listed directories; the
	// order within the set does not determine the font selection,
	// except in the case of identical matches in which case earlier fonts
	// match preferrentially
	fonts [FcSetApplication + 1]FontSet

	// Fontconfig can periodically rescan the system configuration
	// and font directories.  This rescanning occurs when font
	// listing requests are made, but no more often than rescanInterval
	// seconds apart.
	rescanInterval int // interval between scans, in seconds
	// time_t rescanTime     /* last time information was scanned */

	// FcExprPage *expr_pool /* pool of FcExpr's */

	sysRoot          string /* override the system root directory */
	availConfigFiles strSet /* config files available */
	// rulesetList      []*ruleSet /* List of rulesets being installed */
}

// NewFcConfig returns a new empty configuration
func NewFcConfig() *Config {
	var config Config

	config.configDirs = make(strSet)
	config.configMapDirs = make(strSet)
	config.configFiles = make(strSet)
	config.fontDirs = make(strSet)
	config.acceptGlobs = make(strSet)
	config.rejectGlobs = make(strSet)
	config.cacheDirs = make(strSet)

	config.rescanInterval = 30

	// TODO: maybe this is not enough
	config.sysRoot, _ = filepath.Abs(os.Getenv("FONTCONFIG_SYSROOT"))

	config.availConfigFiles = make(strSet)

	return &config
}

// FcConfigSubstituteWithPat performs the sequence of pattern modification operations.
// If `kind` is FcMatchPattern, then those tagged as pattern operations are applied, else
// if `kind` is FcMatchFont, those tagged as font operations are applied and
// `pPat` is used for <test> elements with target=pattern. Returns `false`
// if the substitution cannot be performed.
// If `config` is nil, the current configuration is used.
func (config *Config) FcConfigSubstituteWithPat(p, pPat Pattern, kind matchKind) {

	config = fallbackConfig(config)

	var v FcValue

	if kind == FcMatchPattern {
		strs := FcGetDefaultLangs()
		var lsund LangSet
		lsund.add("und")

		for lang := range strs {
			e := p[FC_LANG]

			for _, ll := range e {
				vvL := ll.Value

				if vv, ok := vvL.(LangSet); ok {
					var ls LangSet
					ls.add(lang)

					b := vv.FcLangSetContains(ls)
					if b {
						goto bail_lang
					}
					if vv.FcLangSetContains(lsund) {
						goto bail_lang
					}
				} else {
					vv, _ := vvL.(String)
					if FcStrCmpIgnoreCase(string(vv), lang) == 0 {
						goto bail_lang
					}
					if FcStrCmpIgnoreCase(string(vv), "und") == 0 {
						goto bail_lang
					}
				}
			}
			v = String(lang)
			p.addWithBinding(FC_LANG, v, FcValueBindingWeak, true)
		}
	bail_lang:
		var res FcResult
		v, res = p.FcPatternObjectGet(FC_PRGNAME, 0)
		if res == FcResultNoMatch {
			prgname := FcGetPrgname()
			if prgname != "" {
				p.Add(FC_PRGNAME, String(prgname), true)
			}
		}
	}

	nobjs := int(FirstCustomObject) - 1 + config.maxObjects + 2
	valuePos := make([]int, nobjs)
	elt := make([]ValueList, nobjs)
	tst := make([]*ruleTest, nobjs)

	if debugMode {
		fmt.Println("FcConfigSubstitute with pattern:")
		fmt.Println(p.String())
	}

	data := newFamilyTable(p)

	var (
		table = &data
	)
	for _, rs := range config.subst {
		if debugMode {
			fmt.Printf("\nRule Set: %s\n", rs.name)
		}
	subsLoop:
		for _, rules := range rs.subst[kind] {
			for i := range valuePos {
				elt[i] = nil
				valuePos[i] = -1
				tst[i] = nil
			}
			for _, r := range rules {
				switch r := r.(type) {
				case ruleTest:
					if debugMode {
						fmt.Println("FcConfigSubstitute test", r)
					}
					var matchLevel int
					table, matchLevel = r.match(kind, p, pPat, data, valuePos, elt, tst)
					if matchLevel == 1 {
						if debugMode {
							fmt.Println("No match")
						}
						continue subsLoop
					} else if matchLevel == 2 {
						if debugMode {
							fmt.Println("No match")
						}
						return
					}
				case ruleEdit:
					if debugMode {
						fmt.Println("FcConfigSubstitute edit", r)
					}
					r.edit(kind, p, pPat, table, valuePos, elt, tst)

					if debugMode {
						fmt.Println("FcConfigSubstitute edit", p.String())
					}
				}
			}
		}
	}
	if debugMode {
		fmt.Println("FcConfigSubstitute done", p.String())
	}

	return
}

// TODO: remove this
func (config *Config) FcConfigGetFonts(set FcSetName) FontSet {
	if config == nil {
		config = FcConfigGetCurrent()
	}
	if config == nil {
		return nil
	}
	return config.fonts[set]
}

func (config *Config) addCacheDir(d string) error {
	return addFilename(config.cacheDirs, d)
}

func (config *Config) addConfigDir(d string) error {
	return addFilename(config.configDirs, d)
}

// TODO:
func (config *Config) addDirList(set FcSetName, dirSet strSet) error {
	// FcStrList	    *dirlist;
	// FcChar8	    *dir;
	// FcCache	    *cache;

	// TODO: take care of the side effect of ading sub directories
	for dir := range dirSet {
		if debugMode {
			fmt.Printf("adding fonts from %s\n", dir)
		}
		err := readDir(dir, false, config)
		if err != nil {
			return err
		}
		// if cache == "" {
		// 	continue
		// }
		// FcConfigAddCache(config, cache, set, dirSet, dir)
		// FcDirCacheUnload(cache)
	}
	return nil
}

// reject several extensions which are not supported font files
func validFontFile(name string) bool {
	// ignore hidden file
	if name == "" || name[0] == '.' {
		return false
	}
	if strings.HasSuffix(name, ".enc.gz") || // encodings
		strings.HasSuffix(name, ".afm") || // metrics (ascii)
		strings.HasSuffix(name, ".pfm") || // metrics (binary)
		strings.HasSuffix(name, ".dir") || // summary
		strings.HasSuffix(name, ".scale") ||
		strings.HasSuffix(name, ".alias") {
		return false
	}
	return true
}

// recursively scan a directory and build the font patterns
func readDir(dir string, force bool, config *Config) error {
	var timings loadTiming

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("error %s, skipping directory", err)
			return nil
		}
		if info.IsDir() { // keep going
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
		}
		if !validFontFile(info.Name()) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = readFontFile(file, &timings)
		if err != nil {
			log.Printf("invalid font file %s: %s", path, err)
		}
		file.Close()
		// TODO: bridge with scanFontConfig()
		return nil
	}
	err := filepath.Walk(dir, walkFn)
	fmt.Println(timings)
	return err
}

//  Scan the specified directory and construct a cache of its contents
func FcDirCacheScan(dir string, config *Config) FcCache {
	//  FcStrSet		*dirs;
	//  FontSet		*set;
	//  FcCache		*cache = NULL;
	//  struct stat		dirStat;
	sysroot := config.getSysRoot()
	//  #ifndef _WIN32
	// 	 int			fd = -1;
	//  #endif

	d := dir
	if sysroot != "" {
		d = filepath.Join(sysroot, dir)
	}

	if debugMode {
		fmt.Printf("cache scan dir %s\n", d)
	}

	_, err := os.Stat(d)
	if err != nil {
		return ""
	}

	var (
		set  FontSet
		dirs = make(strSet)
	)

	//  #ifndef _WIN32
	// 	 fd = FcDirCacheLock (dir, config);
	//  #endif
	// Scan the dir

	// Do not pass sysroot here. FcDirScanConfig() do take care of it
	if !FcDirScanConfig(&set, dirs, dir, config) {
		return ""
	}

	// Build the cache object
	cache := FcDirCacheBuild(set, dir, dirs)

	/// Write out the cache file, ignoring any troubles
	FcDirCacheWrite(cache, config)

	//  #ifndef _WIN32
	// 	 FcDirCacheUnlock (fd);
	//  #endif

	return cache
}

// GetFilename returns the filename associated to an external entity name.
// This provides applications a way to convert various configuration file
// references into filename form.
//
// An empty `name` indicates that the default configuration file should
// be used; which file this references can be overridden with the
// FONTCONFIG_FILE environment variable.
// Next, if the name starts with `~`, it refers to a file in the current users home directory.
// Otherwise if the name doesn't start with '/', it refers to a file in the default configuration
// directory; the built-in default directory can be overridden with the
// FONTCONFIG_PATH environment variable.
//
// The result of this function is affected by the FONTCONFIG_SYSROOT environment variable or equivalent functionality.
func (config *Config) GetFilename(url string) string {
	config = fallbackConfig(config)

	sysroot := config.getSysRoot()

	if url == "" {
		url = os.Getenv("FONTCONFIG_FILE")
		if url == "" {
			url = FONTCONFIG_FILE
		}
	}
	if filepath.IsAbs(url) {
		if sysroot != "" {
			// Workaround to avoid adding sysroot repeatedly
			if url == sysroot {
				sysroot = ""
			}
		}
		return fileExists(sysroot, url)
	}

	file := ""
	if usesHome(url) {
		if dir := FcConfigHome(); dir != "" {
			s := dir
			if sysroot != "" {
				s = filepath.Join(sysroot, dir)
			}
			file = fileExists(s, url[1:])
		}
	} else {
		paths := getPaths()
		for _, p := range paths {
			s := p
			if sysroot != "" {
				s = filepath.Join(sysroot, p)
			}
			file = fileExists(s, url)
			if file != "" {
				break
			}
		}
	}

	return file
}

func realFilename(resolvedName string) string {
	dest, err := os.Readlink(resolvedName)
	if err != nil {
		return resolvedName
	}

	out, err := filepath.Abs(dest)
	if err != nil {
		out = dest
	}

	return out
}

func (config *Config) getSysRoot() string {
	if config == nil {
		config = FcConfigGetCurrent()
		if config == nil {
			return ""
		}
	}
	return config.sysRoot
}

// Set 'sysroot' as the system root directory. All file paths used or created with
// this 'config' (including file properties in patterns) will be considered or
// made relative to this 'sysroot'. This allows a host to generate caches for
// targets at build time. This also allows a cache to be re-targeted to a
// different base directory if 'getSysRoot' is used to resolve file paths.
func (config *Config) setSysRoot(sysroot string) {
	s := sysroot
	// if sysroot != "" {
	// TODO:
	// s = getRealPath(sysroot)
	// }

	config.sysRoot = s
}

func (config *Config) globAdd(glob string, accept bool) {
	set := config.rejectGlobs
	if accept {
		set = config.acceptGlobs
	}
	set[glob] = true
}

func (config *Config) patternsAdd(pattern Pattern, accept bool) {
	set := &config.rejectPatterns
	if accept {
		set = &config.acceptPatterns
	}
	*set = append(*set, pattern)
}

func (config *Config) FcConfigSubstitute(p Pattern, kind matchKind) {
	config.FcConfigSubstituteWithPat(p, nil, kind)
}

// FcConfigBuildFonts scans the current list of directories in the configuration
// and build the set of available fonts. Note that
// any changes to the configuration after this call have indeterminate effects.
// TODO: addDirList is not working yet
func (config *Config) FcConfigBuildFonts() error {
	config = fallbackConfig(config)

	config.fonts[FcSetSystem] = nil
	err := config.addDirList(FcSetSystem, config.fontDirs)

	if debugMode {
		fmt.Println(config.fonts[FcSetSystem])
	}
	return err
}

/* Objects MT-safe for readonly access. */

// #if defined (_WIN32) && !defined (R_OK)
// #define R_OK 4
// #endif

// #if defined(_WIN32) && !defined(S_ISFIFO)
// #define S_ISFIFO(m) 0
// #endif

// static FcConfig    *_fcConfig; /* MT-safe */
// static FcMutex	   *_lock;

// static void
// lock_config (void)
// {
//     FcMutex *lock;
// retry:
//     lock = fc_atomic_ptr_get (&_lock);
//     if (!lock)
//     {
// 	lock = (FcMutex *) malloc (sizeof (FcMutex));
// 	FcMutexInit (lock);
// 	if (!fc_atomic_ptr_cmpexch (&_lock, nil, lock))
// 	{
// 	    FcMutexFinish (lock);
// 	    goto retry;
// 	}
// 	FcMutexLock (lock);
// 	/* Initialize random state */
// 	FcRandom ();
// 	return;
//     }
//     FcMutexLock (lock);
// }

// static void
// unlock_config (void)
// {
//     FcMutex *lock;
//     lock = fc_atomic_ptr_get (&_lock);
//     FcMutexUnlock (lock);
// }

// static void
// free_lock (void)
// {
//     FcMutex *lock;
//     lock = fc_atomic_ptr_get (&_lock);
//     if (lock && fc_atomic_ptr_cmpexch (&_lock, lock, nil))
//     {
// 	FcMutexFinish (lock);
// 	free (lock);
//     }
// }

// static void
// FcDestroyAsRule (void *data)
// {
//     FcRuleDestroy (data);
// }

// static void
// FcDestroyAsRuleSet (void *data)
// {
//     FcRuleSetDestroy (data);
// }

// FcBool
// FcConfigInit (void)
// {
//   return ensure () ? true : false;
// }

// void
// FcConfigFini (void)
// {
//     FcConfig *cfg = fc_atomic_ptr_get (&_fcConfig);
//     if (cfg && fc_atomic_ptr_cmpexch (&_fcConfig, cfg, nil))
// 	FcConfigDestroy (cfg);
//     free_lock ();
// }

// static FcChar8 *
// FcConfigRealPath(const FcChar8 *path)
// {
//     char	resolved_name[FC_PATH_MAX+1];
//     char	*resolved_ret;

//     if (!path)
// 	return nil;

// #ifndef _WIN32
//     resolved_ret = realpath((const char *) path, resolved_name);
// #else
//     if (GetFullPathNameA ((LPCSTR) path, FC_PATH_MAX, resolved_name, nil) == 0)
//     {
//         fprintf (stderr, "Fontconfig warning: GetFullPathNameA failed.\n");
//         return nil;
//     }
//     resolved_ret = resolved_name;
// #endif
//     if (resolved_ret)
// 	path = (FcChar8 *) resolved_ret;
//     return toAbsPath(path);
// }

// static FcFileTime
// FcConfigNewestFile (FcStrSet *files)
// {
//     FcStrList	    *list = FcStrListCreate (files);
//     FcFileTime	    newest = { 0, false };
//     FcChar8	    *file;
//     struct  stat    statb;

//     if (list)
//     {
// 	for ((file = FcStrListNext (list)))
// 	    if (FcStat (file, &statb) == 0)
// 		if (!newest.set || statb.st_mtime - newest.time > 0)
// 		{
// 		    newest.set = true;
// 		    newest.time = statb.st_mtime;
// 		}
// 	FcStrListDone (list);
//     }
//     return newest;
// }

// FcBool
// FcConfigUptoDate (config *FcConfig)
// {
//     FcFileTime	config_time, config_dir_time, font_time;
//     time_t	now = time(0);
//     FcBool	ret = true;

//     config = fallbackConfig (config);
//     if (!config)
// 	return false;

//     config_time = FcConfigNewestFile (config.configFiles);
//     config_dir_time = FcConfigNewestFile (config.configDirs);
//     font_time = FcConfigNewestFile (config.fontDirs);
//     if ((config_time.set && config_time.time - config.rescanTime > 0) ||
// 	(config_dir_time.set && (config_dir_time.time - config.rescanTime) > 0) ||
// 	(font_time.set && (font_time.time - config.rescanTime) > 0))
//     {
// 	/* We need to check for potential clock problems here (OLPC ticket #6046) */
// 	if ((config_time.set && (config_time.time - now) > 0) ||
//     	(config_dir_time.set && (config_dir_time.time - now) > 0) ||
//         (font_time.set && (font_time.time - now) > 0))
// 	{
// 	    fprintf (stderr,
//                     "Fontconfig warning: Directory/file mtime in the future. New fonts may not be detected.\n");
// 	    config.rescanTime = now;
// 	    goto bail;
// 	}
// 	else
// 	{
// 	    ret = false;
// 	    goto bail;
// 	}
//     }
//     config.rescanTime = now;
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// FcExpr *
// FcConfigAllocExpr (config *FcConfig)
// {
//     if (!config.expr_pool || config.expr_pool.next == config.expr_pool.end)
//     {
// 	FcExprPage *new_page;

// 	new_page = malloc (sizeof (FcExprPage));
// 	if (!new_page)
// 	    return 0;

// 	new_page.next_page = config.expr_pool;
// 	new_page.next = new_page.exprs;
// 	config.expr_pool = new_page;
//     }

//     return config.expr_pool.next++;
// }

var (
	defaultConfig     *Config
	defaultConfigLock sync.Mutex
)

// fallback to the current global configuration if `config` is nil
func fallbackConfig(config *Config) *Config {
	if config != nil {
		return config
	}

	// TODO:
	/* lock during obtaining the value from _fcConfig and count up refcount there,
	 * there are the race between them.
	 */
	// lock_config ();
	// retry:
	// config = fc_atomic_ptr_get (&_fcConfig);
	// if (!config) 	{
	//     unlock_config ();

	//     config = initLoadConfigAndFonts ();
	//     if (!config)
	// 	goto retry;
	//     lock_config ();
	//     if (!fc_atomic_ptr_cmpexch (&_fcConfig, nil, config))
	//     {
	// 	FcConfigDestroy (config);
	// 	goto retry;
	//     }
	// }
	// FcRefInc (&config.ref);
	// unlock_config ();

	return config
}

// FcConfigGetCurrent returns the current default configuration.
func FcConfigGetCurrent() *Config { return ensure() }

func ensure() *Config {
	defaultConfigLock.Lock()
	defer defaultConfigLock.Unlock()

	if defaultConfig == nil {
		var err error
		defaultConfig, err = initLoadConfigAndFonts()
		if err != nil {
			log.Fatalf("invalid default configuration: %s", err)
		}
	}

	return defaultConfig
}

// void
// FcConfigDestroy (config *FcConfig)
// {
//     FcSetName	set;
//     FcExprPage	*page;
//     FcMatchKind	k;

//     if (FcRefDec (&config.ref) != 1)
// 	return;

//     (void) fc_atomic_ptr_cmpexch (&_fcConfig, config, nil);

//     FcStrSetDestroy (config.configDirs);
//     FcStrSetDestroy (config.configMapDirs);
//     FcStrSetDestroy (config.fontDirs);
//     FcStrSetDestroy (config.cacheDirs);
//     FcStrSetDestroy (config.configFiles);
//     FcStrSetDestroy (config.acceptGlobs);
//     FcStrSetDestroy (config.rejectGlobs);
//     FcFontSetDestroy (config.acceptPatterns);
//     FcFontSetDestroy (config.rejectPatterns);

//     for (k = FcMatchKindBegin; k < FcMatchKindEnd; k++)
// 	FcPtrListDestroy (config.subst[k]);
//     FcPtrListDestroy (config.rulesetList);
//     FcStrSetDestroy (config.availConfigFiles);
//     for (set = FcSetSystem; set <= FcSetApplication; set++)
// 	if (config.fonts[set])
// 	    FcFontSetDestroy (config.fonts[set]);

//     page = config.expr_pool;
//     for (page)
//     {
//       FcExprPage *next = page.next_page;
//       free (page);
//       page = next;
//     }
//     if (config.sysRoot)
// 	FcStrFree (config.sysRoot);

//     free (config);
// }

// /*
//  * Add cache to configuration, adding fonts and directories
//  */

// FcBool
// FcConfigAddCache (config *FcConfig, FcCache *cache,
// 		  FcSetName set, FcStrSet *dirSet, FcChar8 *forDir)
// {
//     FcFontSet	*fs;
//     intptr_t	*dirs;
//     int		i;
//     FcBool      relocated = false;

//     if (strcmp ((char *)FcCacheDir(cache), (char *)forDir) != 0)
//       relocated = true;

//     /*
//      * Add fonts
//      */
//     fs = FcCacheSet (cache);
//     if (fs)
//     {
// 	int	nref = 0;

// 	for (i = 0; i < fs.nfont; i++)
// 	{
// 	    FcPattern	*font = FcFontSetFont (fs, i);
// 	    FcChar8	*font_file;
// 	    FcChar8	*relocated_font_file = nil;

// 	    if (FcPatternObjectGetString (font, FC_FILE,
// 					  0, &font_file) == FcResultMatch)
// 	    {
// 		if (relocated)
// 		  {
// 		    FcChar8 *slash = FcStrLastSlash (font_file);
// 		    relocated_font_file = filepath.Join (forDir, slash + 1, nil);
// 		    font_file = relocated_font_file;
// 		  }

// 		/*
// 		 * Check to see if font is banned by filename
// 		 */
// 		if (!FcConfigAcceptFilename (config, font_file))
// 		{
// 		    free (relocated_font_file);
// 		    continue;
// 		}
// 	    }

// 	    /*
// 	     * Check to see if font is banned by pattern
// 	     */
// 	    if (!FcConfigAcceptFont (config, font))
// 	    {
// 		free (relocated_font_file);
// 		continue;
// 	    }

// 	    if (relocated_font_file)
// 	    {
// 	      font = FcPatternCacheRewriteFile (font, cache, relocated_font_file);
// 	      free (relocated_font_file);
// 	    }

// 	    if (FcFontSetAdd (config.fonts[set], font))
// 		nref++;
// 	}
// 	FcDirCacheReference (cache, nref);
//     }

//     /*
//      * Add directories
//      */
//     dirs = FcCacheDirs (cache);
//     if (dirs)
//     {
// 	for (i = 0; i < cache.dirs_count; i++)
// 	{
// 	    const FcChar8 *dir = FcCacheSubdir (cache, i);
// 	    FcChar8 *s = nil;

// 	    if (relocated)
// 	    {
// 		FcChar8 *base = FcStrBasename (dir);
// 		dir = s = filepath.Join (forDir, base, nil);
// 		FcStrFree (base);
// 	    }
// 	    if (FcConfigAcceptFilename (config, dir))
// 		FcStrSetAddFilename (dirSet, dir);
// 	    if (s)
// 		FcStrFree (s);
// 	}
//     }
//     return true;
// }

// FcBool
// FcConfigSetCurrent (config *FcConfig)
// {
//     FcConfig *cfg;

//     if (config)
//     {
// 	if (!config.fonts[FcSetSystem])
// 	    if (!FcConfigBuildFonts (config))
// 		return false;
// 	FcRefInc (&config.ref);
//     }

//     lock_config ();
// retry:
//     cfg = fc_atomic_ptr_get (&_fcConfig);

//     if (config == cfg)
//     {
// 	unlock_config ();
// 	if (config)
// 	    FcConfigDestroy (config);
// 	return true;
//     }

//     if (!fc_atomic_ptr_cmpexch (&_fcConfig, cfg, config))
// 	goto retry;
//     unlock_config ();
//     if (cfg)
// 	FcConfigDestroy (cfg);

//     return true;
// }

// FcBool
// FcConfigAddConfigDir (config *FcConfig,
// 		      const FcChar8 *d)
// {
//     return FcStrSetAddFilename (config.configDirs, d);
// }

// FcStrList *
// FcConfigGetConfigDirs (FcConfig   *config)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.configDirs);
//     FcConfigDestroy (config);

//     return ret;
// }

// Adds `s` to `set`, applying toAbsPath so that leading '~' values are replaced
// with the value of the HOME environment variable.
func addFilename(set strSet, s string) error {
	s, err := toAbsPath(s)
	if err != nil {
		return err
	}
	set[s] = true
	return nil
}

func (config *Config) addFontDir(d, m, salt string) error {
	if debugMode {
		if m != "" {
			fmt.Printf("add font dir: %s . %s %s\n", d, m, salt)
		} else if salt != "" {
			fmt.Printf("add font dir: %s %s\n", d, salt)
		} else {
			fmt.Println("add font dir:", d)
		}
	}
	return addFilenamePairWithSalt(config.fontDirs, d, m, salt)
}

func addFilenamePairWithSalt(set strSet, a, b, salt string) error {
	var err error
	a, err = toAbsPath(a)
	if err != nil {
		return err
	}
	b, err = toAbsPath(b)
	if err != nil {
		return err
	}
	fmt.Println(a, b)
	// override maps with new one if exists
	c := a + b
	for s := range set {
		if strings.HasPrefix(s, c) {
			delete(set, s)
		}
	}
	set[a] = true
	return nil
}

// FcBool
// FcConfigResetFontDirs (config *FcConfig)
// {
//     if (FcDebug() & FC_DBG_CACHE)
//     {
// 	printf ("Reset font directories!\n");
//     }
//     return FcStrSetDeleteAll (config.fontDirs);
// }

// FcStrList *
// FcConfigGetFontDirs (config *FcConfig)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.fontDirs);
//     FcConfigDestroy (config);

//     return ret;
// }

// static FcBool
// FcConfigPathStartsWith(const FcChar8	*path,
// 		       const FcChar8	*start)
// {
//     int len = strlen((char *) start);

//     if (strncmp((char *) path, (char *) start, len) != 0)
// 	return false;

//     switch (path[len]) {
//     case '\0':
//     case FC_DIR_SEPARATOR:
// 	return true;
//     default:
// 	return false;
//     }
// }

// FcStrList *
// FcConfigGetCacheDirs (config *FcConfig)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.cacheDirs);
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigAddConfigFile (config *FcConfig,
// 		       const FcChar8   *f)
// {
//     FcBool	ret;
//     FcChar8	*file = GetFilename (config, f);

//     if (!file)
// 	return false;

//     ret = FcStrSetAdd (config.configFiles, file);
//     FcStrFree (file);
//     return ret;
// }

// FcStrList *
// FcConfigGetConfigFiles (config *FcConfig)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.configFiles);
//     FcConfigDestroy (config);

//     return ret;
// }

// FcChar8 *
// FcConfigGetCache (FcConfig  *config FC_UNUSED)
// {
//     return nil;
// }

// void
// FcConfigSetFonts (config *FcConfig,
// 		  FcFontSet	*fonts,
// 		  FcSetName	set)
// {
//     if (config.fonts[set])
// 	FcFontSetDestroy (config.fonts[set]);
//     config.fonts[set] = fonts;
// }

// FcBlanks *
// FcBlanksCreate (void)
// {
//     /* Deprecated. */
//     return nil;
// }

// void
// FcBlanksDestroy (FcBlanks *b FC_UNUSED)
// {
//     /* Deprecated. */
// }

// FcBool
// FcBlanksAdd (FcBlanks *b FC_UNUSED, FcChar32 ucs4 FC_UNUSED)
// {
//     /* Deprecated. */
//     return false;
// }

// FcBool
// FcBlanksIsMember (FcBlanks *b FC_UNUSED, FcChar32 ucs4 FC_UNUSED)
// {
//     /* Deprecated. */
//     return false;
// }

// FcBlanks *
// FcConfigGetBlanks (config *FcConfig FC_UNUSED)
// {
//     /* Deprecated. */
//     return nil;
// }

// FcBool
// FcConfigAddBlank (config *FcConfig FC_UNUSED,
// 		  FcChar32    	blank FC_UNUSED)
// {
//     /* Deprecated. */
//     return false;
// }

// int
// FcConfigGetRescanInterval (config *FcConfig)
// {
//     int ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return 0;
//     ret = config.rescanInterval;
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigSetRescanInterval (config *FcConfig, int rescanInterval)
// {
//     config = fallbackConfig (config);
//     if (!config)
// 	return false;
//     config.rescanInterval = rescanInterval;
//     FcConfigDestroy (config);

//     return true;
// }

// /*
//  * A couple of typos escaped into the library
//  */
// int
// FcConfigGetRescanInverval (config *FcConfig)
// {
//     return FcConfigGetRescanInterval (config);
// }

// FcBool
// FcConfigSetRescanInverval (config *FcConfig, int rescanInterval)
// {
//     return FcConfigSetRescanInterval (config, rescanInterval);
// }

// FcBool
// FcConfigAddRule (config *FcConfig,
// 		 FcRule		*rule,
// 		 FcMatchKind	kind)
// {
//     /* deprecated */
//     return false;
// }

/* The bulk of the time in FcConfigSubstitute is spent walking
 * lists of family names. We speed this up with a hash table.
 * Since we need to take the ignore-blanks option into account,
 * we use two separate hash tables.
 */
// typedef struct
// {
//   int count;
// } FamilyTableEntry;

type familyTable struct {
	familyBlankHash familyBlankMap
	familyHash      familyMap
}

func newFamilyTable(p Pattern) familyTable {
	table := familyTable{
		familyBlankHash: make(familyBlankMap),
		familyHash:      make(familyMap),
	}

	e := p[FC_FAMILY]
	table.add(e)
	return table
}

func (table familyTable) lookup(op FcOp, s String) bool {
	flags := op.getFlags()
	var has bool

	if (flags & FcOpFlagIgnoreBlanks) != 0 {
		_, has = table.familyBlankHash.lookup(s)
	} else {
		_, has = table.familyHash.lookup(s)
	}

	return has
}

func (table familyTable) add(values ValueList) {
	for _, ll := range values {
		s := ll.Value.(String)

		count, _ := table.familyHash.lookup(s)
		count++
		table.familyHash.add(s, count)

		count, _ = table.familyBlankHash.lookup(s)
		count++
		table.familyBlankHash.add(s, count)
	}
}

func (table familyTable) del(s String) {
	count, ok := table.familyHash.lookup(s)
	if ok {
		count--
		if count == 0 {
			table.familyHash.del(s)
		} else {
			table.familyHash.add(s, count)
		}
	}

	count, ok = table.familyBlankHash.lookup(s)
	if ok {
		count--
		if count == 0 {
			table.familyBlankHash.del(s)
		} else {
			table.familyBlankHash.add(s, count)
		}
	}
}

// static FcBool
// copy_string (const void *src, void **dest)
// {
//   *dest = strdup ((char *)src);
//   return true;
// }

// return the index into values, or -1
func matchValueList(p, pPat Pattern, kind matchKind,
	t ruleTest, values ValueList, table *familyTable) int {

	var (
		value FcValue
		e     = t.expr
		ret   = -1
	)

	for e != nil {
		// Compute the value of the match expression
		if e.op.getOp() == FcOpComma {
			tree := e.u.(exprTree)
			value = tree.left.FcConfigEvaluate(p, pPat, kind)
			e = tree.right
		} else {
			value = e.FcConfigEvaluate(p, pPat, kind)
			e = nil
		}

		if t.object == FC_FAMILY && table != nil {
			op := t.op.getOp()
			if op == FcOpEqual || op == FcOpListing {
				if !table.lookup(t.op, value.(String)) {
					ret = -1
					continue
				}
			}
			if op == FcOpNotEqual && t.qual == FcQualAll {
				ret = -1
				if !table.lookup(t.op, value.(String)) {
					ret = 0
				}
				continue
			}
		}

		for i, v := range values {
			// Compare the pattern value to the match expression value
			if compareValue(v.Value, t.op, value) {
				if ret == -1 {
					ret = i
				}
				if t.qual != FcQualAll {
					break
				}
			} else {
				if t.qual == FcQualAll {
					ret = -1
					break
				}
			}
		}
	}
	return ret
}

// #if defined (_WIN32)

// static FcChar8 fontconfig_path[1000] = ""; /* MT-dontcare */
// FcChar8 fontconfig_instprefix[1000] = ""; /* MT-dontcare */

// #  if (defined (PIC) || defined (DLL_EXPORT))

// BOOL WINAPI
// DllMain (HINSTANCE hinstDLL,
// 	 DWORD     fdwReason,
// 	 LPVOID    lpvReserved);

// BOOL WINAPI
// DllMain (HINSTANCE hinstDLL,
// 	 DWORD     fdwReason,
// 	 LPVOID    lpvReserved)
// {
//   FcChar8 *p;

//   switch (fdwReason) {
//   case DLL_PROCESS_ATTACH:
//       if (!GetModuleFileName ((HMODULE) hinstDLL, (LPCH) fontconfig_path,
// 			      sizeof (fontconfig_path)))
// 	  break;

//       /* If the fontconfig DLL is in a "bin" or "lib" subfolder,
//        * assume it's a Unix-style installation tree, and use
//        * "etc/fonts" in there as FONTCONFIG_PATH. Otherwise use the
//        * folder where the DLL is as FONTCONFIG_PATH.
//        */
//       p = (FcChar8 *) strrchr ((const char *) fontconfig_path, '\\');
//       if (p)
//       {
// 	  *p = '\0';
// 	  p = (FcChar8 *) strrchr ((const char *) fontconfig_path, '\\');
// 	  if (p && (FcStrCmpIgnoreCase (p + 1, (const FcChar8 *) "bin") == 0 ||
// 		    FcStrCmpIgnoreCase (p + 1, (const FcChar8 *) "lib") == 0))
// 	      *p = '\0';
// 	  strcat ((char *) fontconfig_instprefix, (char *) fontconfig_path);
// 	  strcat ((char *) fontconfig_path, "\\etc\\fonts");
//       }
//       else
//           fontconfig_path[0] = '\0';

//       break;
//   }

//   return TRUE;
// }

// #  endif /* !PIC */

// #undef FONTCONFIG_PATH
// #define FONTCONFIG_PATH fontconfig_path

// #endif /* !_WIN32 */

// #ifndef FONTCONFIG_FILE
// #define FONTCONFIG_FILE	"fonts.conf"
// #endif

// join the path, check it, and returns it if valid
func fileExists(dir, file string) string {
	path := filepath.Join(dir, file)
	// we do a basic error checking, not taking into account the
	// various failure possible; but it should be enough ?
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

func getPaths() []string {
	env := os.Getenv("FONTCONFIG_PATH")
	paths := filepath.SplitList(env)

	// is used to override the default configuration directory.
	fontConfig := "/usr/local/etc/fonts"
	if runtime.GOOS == "windows" {
		fcPath, err := os.Executable()
		if err != nil {
			return paths
		}
		fontConfig = filepath.Join(fcPath, "fonts")
	}

	paths = append(paths, fontConfig)
	return paths
}

// static void
// FcConfigFreePath (FcChar8 **path)
// {
//     FcChar8    **p;

//     for (p = path; *p; p++)
// 	free (*p);
//     free (path);
// }

// static FcBool	homeEnabled = true; /* MT-goodenough */

// FcChar8 *
// FcConfigHome (void)
// {
//     if (homeEnabled)
//     {
//         char *home = getenv ("HOME");

// #ifdef _WIN32
// 	if (home == nil)
// 	    home = getenv ("USERPROFILE");
// #endif

// 	return (FcChar8 *) home;
//     }
//     return 0;
// }

// FcChar8 *
// FcConfigXdgCacheHome (void)
// {
//     const char *env = getenv ("XDG_CACHE_HOME");
//     FcChar8 *ret = nil;

//     if (!homeEnabled)
// 	return nil;
//     if (env && env[0])
// 	ret = FcStrCopy ((const FcChar8 *)env);
//     else
//     {
// 	const FcChar8 *home = FcConfigHome ();
// 	size_t len = home ? strlen ((const char *)home) : 0;

// 	ret = malloc (len + 7 + 1);
// 	if (ret)
// 	{
// 	    if (home)
// 		memcpy (ret, home, len);
// 	    memcpy (&ret[len], FC_DIR_SEPARATOR_S ".cache", 7);
// 	    ret[len + 7] = 0;
// 	}
//     }

//     return ret;
// }

// FcChar8 *
// FcConfigXdgDataHome (void)
// {
//     const char *env = getenv ("XDG_DATA_HOME");
//     FcChar8 *ret = nil;

//     if (!homeEnabled)
// 	return nil;
//     if (env)
// 	ret = FcStrCopy ((const FcChar8 *)env);
//     else
//     {
// 	const FcChar8 *home = FcConfigHome ();
// 	size_t len = home ? strlen ((const char *)home) : 0;

// 	ret = malloc (len + 13 + 1);
// 	if (ret)
// 	{
// 	    if (home)
// 		memcpy (ret, home, len);
// 	    memcpy (&ret[len], FC_DIR_SEPARATOR_S ".local" FC_DIR_SEPARATOR_S "share", 13);
// 	    ret[len + 13] = 0;
// 	}
//     }

//     return ret;
// }

// FcBool
// FcConfigEnableHome (FcBool enable)
// {
//     FcBool  prev = homeEnabled;
//     homeEnabled = enable;
//     return prev;
// }

// FcChar8 *
// FcConfigFilename (const FcChar8 *url)
// {
//     return GetFilename (nil, url);
// }

// FcChar8 *
// FcConfigRealFilename (FcConfig		*config,
// 		      const FcChar8	*url)
// {
//     FcChar8 *n = GetFilename (config, url);

//     if (n)
//     {
// 	FcChar8 buf[FC_PATH_MAX];
// 	ssize_t len;
// 	struct stat sb;

// 	if ((len = FcReadLink (n, buf, sizeof (buf) - 1)) != -1)
// 	{
// 	    buf[len] = 0;

// 	    /* We try to pick up a config from FONTCONFIG_FILE
// 	     * when url is null. don't try to address the real filename
// 	     * if it is a named pipe.
// 	     */
// 	    if (!url && FcStat (n, &sb) == 0 && S_ISFIFO (sb.st_mode))
// 		return n;
// 	    else if (!FcStrIsAbsoluteFilename (buf))
// 	    {
// 		FcChar8 *dirname = FcStrDirname (n);
// 		FcStrFree (n);
// 		if (!dirname)
// 		    return nil;

// 		FcChar8 *path = filepath.Join (dirname, buf, nil);
// 		FcStrFree (dirname);
// 		if (!path)
// 		    return nil;

// 		n = FcStrCanonFilename (path);
// 		FcStrFree (path);
// 	    }
// 	    else
// 	    {
// 		FcStrFree (n);
// 		n = FcStrdup (buf);
// 	    }
// 	}
//     }

//     return n;
// }

// /*
//  * Manage the application-specific fonts
//  */

// FcBool
// FcConfigAppFontAddFile (config *FcConfig,
// 			const FcChar8  *file)
// {
//     FcFontSet	*set;
//     FcStrSet	*subdirs;
//     FcStrList	*sublist;
//     FcChar8	*subdir;
//     FcBool	ret = true;

//     config = fallbackConfig (config);
//     if (!config)
// 	return false;

//     subdirs = FcStrSetCreateEx (FCSS_GROW_BY_64);
//     if (!subdirs)
//     {
// 	ret = false;
// 	goto bail;
//     }

//     set = FcConfigGetFonts (config, FcSetApplication);
//     if (!set)
//     {
// 	set = FcFontSetCreate ();
// 	if (!set)
// 	{
// 	    FcStrSetDestroy (subdirs);
// 	    ret = false;
// 	    goto bail;
// 	}
// 	FcConfigSetFonts (config, set, FcSetApplication);
//     }

//     if (!FcFileScanConfig (set, subdirs, file, config))
//     {
// 	FcStrSetDestroy (subdirs);
// 	ret = false;
// 	goto bail;
//     }
//     if ((sublist = FcStrListCreate (subdirs)))
//     {
// 	for ((subdir = FcStrListNext (sublist)))
// 	{
// 	    FcConfigAppFontAddDir (config, subdir);
// 	}
// 	FcStrListDone (sublist);
//     }
//     FcStrSetDestroy (subdirs);
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigAppFontAddDir (config *FcConfig,
// 		       const FcChar8   *dir)
// {
//     FcFontSet	*set;
//     FcStrSet	*dirs;
//     FcBool	ret = true;

//     config = fallbackConfig (config);
//     if (!config)
// 	return false;

//     dirs = FcStrSetCreateEx (FCSS_GROW_BY_64);
//     if (!dirs)
//     {
// 	ret = false;
// 	goto bail;
//     }

//     set = FcConfigGetFonts (config, FcSetApplication);
//     if (!set)
//     {
// 	set = FcFontSetCreate ();
// 	if (!set)
// 	{
// 	    FcStrSetDestroy (dirs);
// 	    ret = false;
// 	    goto bail;
// 	}
// 	FcConfigSetFonts (config, set, FcSetApplication);
//     }

//     FcStrSetAddFilename (dirs, dir);

//     if (!addDirList (config, FcSetApplication, dirs))
//     {
// 	FcStrSetDestroy (dirs);
// 	ret = false;
// 	goto bail;
//     }
//     FcStrSetDestroy (dirs);
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// void
// FcConfigAppFontClear (config *FcConfig)
// {
//     config = fallbackConfig (config);
//     if (!config)
// 	return;

//     FcConfigSetFonts (config, 0, FcSetApplication);

//     FcConfigDestroy (config);
// }

// /*
//  * Manage filename-based font source selectors
//  */

// static FcBool
// FcConfigGlobsMatch (const FcStrSet	*globs,
// 		    const FcChar8	*string)
// {
//     int	i;

//     for (i = 0; i < globs.num; i++)
// 	if (FcStrGlobMatch (globs.strs[i], string))
// 	    return true;
//     return false;
// }

// FcBool
// FcConfigAcceptFilename (config *FcConfig,
// 			const FcChar8	*filename)
// {
//     if (FcConfigGlobsMatch (config.acceptGlobs, filename))
// 	return true;
//     if (FcConfigGlobsMatch (config.rejectGlobs, filename))
// 	return false;
//     return true;
// }

// /*
//  * Manage font-pattern based font source selectors
//  */

// static FcBool
// FcConfigPatternsMatch (const FcFontSet	*patterns,
// 		       const FcPattern	*font)
// {
//     int i;

//     for (i = 0; i < patterns.nfont; i++)
// 	if (FcListPatternMatchAny (patterns.fonts[i], font))
// 	    return true;
//     return false;
// }

// FcBool
// FcConfigAcceptFont (config *FcConfig,
// 		    const FcPattern *font)
// {
//     if (FcConfigPatternsMatch (config.acceptPatterns, font))
// 	return true;
//     if (FcConfigPatternsMatch (config.rejectPatterns, font))
// 	return false;
//     return true;
// }

// void
// FcRuleSetDestroy (FcRuleSet *rs)
// {
//     FcMatchKind k;

//     if (!rs)
// 	return;
//     if (FcRefDec (&rs.ref) != 1)
// 	return;

//     if (rs.name)
// 	FcStrFree (rs.name);
//     if (rs.description)
// 	FcStrFree (rs.description);
//     if (rs.domain)
// 	FcStrFree (rs.domain);
//     for (k = FcMatchKindBegin; k < FcMatchKindEnd; k++)
// 	FcPtrListDestroy (rs.subst[k]);

//     free (rs);
// }

// void
// FcRuleSetReference (FcRuleSet *rs)
// {
//     if (!FcRefIsConst (&rs.ref))
// 	FcRefInc (&rs.ref);
// }

// void
// FcRuleSetEnable (FcRuleSet	*rs,
// 		 FcBool		flag)
// {
//     if (rs)
//     {
// 	rs.enabled = flag;
// 	/* XXX: we may want to provide a feature
// 	 * to enable/disable rulesets through API
// 	 * in the future?
// 	 */
//     }
// }

// void
// FcRuleSetAddDescription (FcRuleSet	*rs,
// 			 const FcChar8	*domain,
// 			 const FcChar8	*description)
// {
//     if (rs.domain)
// 	FcStrFree (rs.domain);
//     if (rs.description)
// 	FcStrFree (rs.description);

//     rs.domain = domain ? FcStrdup (domain) : nil;
//     rs.description = description ? FcStrdup (description) : nil;
// }

// void
// FcConfigFileInfoIterInit (FcConfig		*config,
// 			  FcConfigFileInfoIter	*iter)
// {
//     FcConfig *c;
//     FcPtrListIter *i = (FcPtrListIter *)iter;

//     if (!config)
// 	c = FcConfigGetCurrent ();
//     else
// 	c = config;
//     FcPtrListIterInit (c.rulesetList, i);
// }

// FcBool
// FcConfigFileInfoIterNext (FcConfig		*config,
// 			  FcConfigFileInfoIter	*iter)
// {
//     FcConfig *c;
//     FcPtrListIter *i = (FcPtrListIter *)iter;

//     if (!config)
// 	c = FcConfigGetCurrent ();
//     else
// 	c = config;
//     if (FcPtrListIterIsValid (c.rulesetList, i))
//     {
// 	FcPtrListIterNext (c.rulesetList, i);
//     }
//     else
// 	return false;

//     return true;
// }

// FcBool
// FcConfigFileInfoIterGet (FcConfig		*config,
// 			 FcConfigFileInfoIter	*iter,
// 			 FcChar8		**name,
// 			 FcChar8		**description,
// 			 FcBool			*enabled)
// {
//     FcConfig *c;
//     FcRuleSet *r;
//     FcPtrListIter *i = (FcPtrListIter *)iter;

//     if (!config)
// 	c = FcConfigGetCurrent ();
//     else
// 	c = config;
//     if (!FcPtrListIterIsValid (c.rulesetList, i))
// 	return false;
//     r = FcPtrListIterGetValue (c.rulesetList, i);
//     if (name)
// 	*name = FcStrdup (r.name && r.name[0] ? r.name : (const FcChar8 *) "fonts.conf");
//     if (description)
// 	*description = FcStrdup (!r.description ? _("No description") :
// 				 dgettext (r.domain ? (const char *) r.domain : GETTEXT_PACKAGE "-conf",
// 					   (const char *) r.description));
//     if (enabled)
// 	*enabled = r.enabled;

//     return true;
// }
