package fontconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from fontconfig/src/fccfg.c Copyright © 2000 Keith Packard

// Config holds a complete configuration of the library.
//
// This object is used to transform the patterns used in queries and returned
// as results. It also provides a way to exclude particular files/directories/patterns
// when scanning the available fonts.
//
// A configuration is constructed from XML data files, with the `LoadFromMemory`
// and `LoadFromDir` methods.
// The 'standard' default configuration is provided in the 'confs/' directory and
// as the global variable `Standard`.
//
// See also `ScanAndCache` for an all-in-one wrapper.
type Config struct {
	// Substitution instructions for patterns and fonts;
	subst []ruleSet

	// the name is user defined, and the object assigned by the library
	customObjects map[string]Object

	// List of patterns used to control font file selection
	acceptGlobs    strSet
	rejectGlobs    strSet
	acceptPatterns Fontset
	rejectPatterns Fontset

	// maximum difference of all custom objects used
	// used to allocate appropriate intermediate storage
	// for performing a whole set of substitutions
	maxObjects int
}

// NewConfig returns a new empty, initialized configuration
func NewConfig() *Config {
	var config Config

	config.acceptGlobs = make(strSet)
	config.rejectGlobs = make(strSet)
	config.customObjects = make(map[string]Object)

	return &config
}

// Copy returns a deep copy of the configuration.
func (c *Config) Copy() *Config {
	if c == nil {
		return c
	}

	out := *c
	out.subst = make([]ruleSet, len(c.subst))
	for i, v := range c.subst {
		out.subst[i] = v.copy()
	}

	out.customObjects = make(map[string]Object, len(c.customObjects))
	for k, v := range c.customObjects {
		out.customObjects[k] = v
	}

	out.acceptGlobs = make(strSet, len(c.acceptGlobs))
	for k, v := range c.acceptGlobs {
		out.acceptGlobs[k] = v
	}
	out.rejectGlobs = make(strSet, len(c.rejectGlobs))
	for k, v := range c.rejectGlobs {
		out.rejectGlobs[k] = v
	}

	out.acceptPatterns = make(Fontset, len(c.acceptPatterns))
	for i, v := range c.acceptPatterns {
		out.acceptPatterns[i] = v.Duplicate()
	}
	out.rejectPatterns = make(Fontset, len(c.rejectPatterns))
	for i, v := range c.rejectPatterns {
		out.rejectPatterns[i] = v.Duplicate()
	}

	return &out
}

// Walks the configuration in `r` and constructs the internal representation
// in `config`.
// The new rules are added to the configuration, meaning that several file
// can be merged by repeated calls.
func (config *Config) LoadFromMemory(r io.Reader) error {
	return config.parseAndLoadFromMemory("memory", r)
}

// LoadFromDir scans this directory, loading all files of the form [0-9]*.conf,
// and recurse through the subdirectories.
// It may be used with the included folder 'confs/' to build a 'standard' configuration.
// See `LoadFromMemory` if you want control over individual files.
func (config *Config) LoadFromDir(dir string) error {
	if debugMode {
		fmt.Printf("\tScanning config dir %s\n", dir)
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("fontconfig: cannot read config %s : %s", path, err)
		}
		if info.IsDir() { // keep going
			return nil
		}
		// add all files of the form [0-9]*.conf
		name := info.Name()
		if isConf := name != "" && '0' <= name[0] && name[0] <= '9' && strings.HasSuffix(name, ".conf"); !isConf {
			return nil // ignore the file
		}

		if info.Mode()&os.ModeSymlink != 0 {
			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
		}

		fi, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("fontconfig: can't open such file %s: %s", path, err)
		}
		defer fi.Close()

		err = config.parseAndLoadFromMemory(path, fi)
		return err
	}

	err := filepath.Walk(dir, walkFn)
	return err
}

// ScanFontDirectories recursively scans the given directories, opening the
// valid font files and building the associated font patterns.
// Symbolic links for files are resolved, but not for directories.
// The rules with kind `MatchScan` in `config` are applied to the results.
// The <selectfont> rules defined in the configuration are applied to filter
// the returned set.
//
// An error is returned if the directory traversal fails, not for invalid font files,
// which are simply ignored.
func (config *Config) ScanFontDirectories(dirs ...string) (Fontset, error) {
	seen := make(strSet) // keep track of visited dirs to avoid double includes
	var out Fontset
	for _, dir := range dirs {
		fonts, err := config.readDir(dir, seen)
		if err != nil {
			return nil, err
		}
		out = append(out, fonts...)
	}
	return out, nil
}

// ScanFontFile scans one font file (see ScanFontDirectories for more details).
// Here, an error is returned for an invalid font file.
// Note that only the pattern-based font selector specified in the config (if any),
// are applied.
func (config *Config) ScanFontFile(path string) (Fontset, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return config.ScanFontRessource(file, path)
}

// ScanFontRessource is the same as `ScanFontFile`, for general content.
// `contentID` is included in the returned patterns as the file name.
func (config *Config) ScanFontRessource(content fonts.Resource, contentID string) (Fontset, error) {
	fonts := scanOneFontFile(content, contentID, config)
	if len(fonts) == 0 {
		return nil, fmt.Errorf("invalid (or empty) font file %s", contentID)
	}
	// pattern selector
	var out Fontset
	for _, f := range fonts {
		if config.acceptFont(f) {
			out = append(out, f)
		}
	}
	return fonts, nil
}

// Recursively scan a directory and build the font patterns,
// which are fetch from the font files and eddited according to `config`.
// `seen` is updated with the visited dirs, and already seen ones are ignored.
// files and patterns are filtered according to accept/reject criteria
// An error is returned if the walk fails, not for invalid font files
func (config *Config) readDir(dir string, seen strSet) (Fontset, error) {
	if debugMode {
		fmt.Println("adding fonts from", dir)
	}

	var out Fontset
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("invalid font location: %s", err)
		}
		if info.IsDir() { // keep going
			if seen[path] {
				return filepath.SkipDir
			}
			seen[path] = true
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

		// path selector
		if !config.acceptFilename(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		fonts := scanOneFontFile(file, path, config)
		file.Close()

		if debugMode {
			if len(fonts) == 0 {
				fmt.Println("invalid font file", path)
			}
		}

		// pattern selector
		for _, f := range fonts {
			if config.acceptFont(f) {
				out = append(out, f)
			}
		}
		return nil
	}

	err := filepath.Walk(dir, walkFn)

	return out, err
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

// uses filename-based font source selectors to accept/reject a file
func (config *Config) acceptFilename(filename string) bool {
	globsMatch := func(globs strSet, name string) bool {
		for glob := range globs {
			if ok, _ := filepath.Match(glob, name); ok {
				return true
			}
		}
		return false
	}

	if globsMatch(config.acceptGlobs, filename) {
		return true
	}
	if globsMatch(config.rejectGlobs, filename) {
		return false
	}
	return true
}

// uses font-pattern based font source selectors to accept/reject a file
func (config *Config) acceptFont(font Pattern) bool {
	patternsMatch := func(patterns Fontset, font Pattern) bool {
		for _, f := range patterns {
			if patternMatchAny(f, font) {
				return true
			}
		}
		return false
	}

	if patternsMatch(config.acceptPatterns, font) {
		return true
	}
	if patternsMatch(config.rejectPatterns, font) {
		return false
	}
	return true
}
