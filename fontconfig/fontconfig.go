// Package fontconfig provides a way to list the fonts of a system
// and to query the best match with user defined criteria.
//
// See the `Config` type for an entry point.
//
// This package is a port of the C library.
package fontconfig

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// test only: print debug information to stdout
const debugMode = false

// DefaultFontDirs return the OS-dependent usual directories for
// fonts, or an error if no one exists.
func DefaultFontDirs() ([]string, error) {
	var dirs []string
	switch runtime.GOOS {
	case "windows":
		sysRoot := os.Getenv("SYSTEMROOT")
		if sysRoot == "" {
			sysRoot = os.Getenv("SYSTEMDRIVE")
		}
		if sysRoot == "" { // try with the common C:
			sysRoot = "C:"
		}
		dir := filepath.Join(filepath.VolumeName(sysRoot), `\Windows`, "Fonts")
		dirs = []string{dir}
	case "darwin":
		dirs = []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
			"/Network/Library/Fonts",
			"/System/Library/Assets/com_apple_MobileAsset_Font3",
			"/System/Library/Assets/com_apple_MobileAsset_Font4",
			"/System/Library/Assets/com_apple_MobileAsset_Font5",
		}
	case "linux":
		dirs = []string{
			"/usr/share/fonts",
			"/usr/share/texmf/fonts/opentype/public",
		}
	case "android":
		dirs = []string{
			"/system/fonts",
			"/system/font",
			"/data/fonts",
		}
	case "ios":
		dirs = []string{
			"/System/Library/Fonts",
			"/System/Library/Fonts/Cache",
		}
	default:
		return nil, fmt.Errorf("unsupported plaform %s", runtime.GOOS)
	}

	var validDirs []string
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			log.Println("invalid font dir", dir, err)
			continue
		}
		if !info.IsDir() {
			log.Println("font dir is not a directory", dir)
			continue
		}
		validDirs = append(validDirs, dir)
	}
	if len(validDirs) == 0 {
		return nil, errors.New("no font directory found")
	}

	return validDirs, nil
}

// Fontset contains a list of Patterns, containing the
// results of listing fonts.
// The order within the set does not determine the font selection,
// except in the case of identical matches in which case earlier fonts
// match preferrentially.
type Fontset []Pattern

type strSet = map[string]bool

// Returns whether `a` contains precisely the same
// strings as `b`. Ordering of strings within the two
// sets is not considered.
func strSetEquals(a, b strSet) bool {
	if len(a) != len(b) {
		return false
	}
	for sa := range a {
		if !b[sa] {
			return false
		}
	}
	return true
}

// Result is returned when accessing elements of a pattern.
type Result uint8

const (
	ResultMatch Result = iota
	ResultNoMatch
	ResultTypeMismatch
	ResultNoId
)

const (
	SLANT_ROMAN   = 0
	SLANT_ITALIC  = 100
	SLANT_OBLIQUE = 110

	WIDTH_ULTRACONDENSED = 50
	WIDTH_EXTRACONDENSED = 63
	WIDTH_CONDENSED      = 75
	WIDTH_SEMICONDENSED  = 87
	WIDTH_NORMAL         = 100
	WIDTH_SEMIEXPANDED   = 113
	WIDTH_EXPANDED       = 125
	WIDTH_EXTRAEXPANDED  = 150
	WIDTH_ULTRAEXPANDED  = 200

	PROPORTIONAL = 0
	DUAL         = 90
	MONO         = 100
	CHARCELL     = 110

	/* sub-pixel order */
	RGBA_UNKNOWN = 0
	RGBA_RGB     = 1
	RGBA_BGR     = 2
	RGBA_VRGB    = 3
	RGBA_VBGR    = 4
	RGBA_NONE    = 5

	/* hinting style */
	HINT_NONE   = 0
	HINT_SLIGHT = 1
	HINT_MEDIUM = 2
	HINT_FULL   = 3

	/* LCD filter */
	LCD_NONE    = 0
	LCD_DEFAULT = 1
	LCD_LIGHT   = 2
	LCD_LEGACY  = 3
)

const (
	WEIGHT_THIN       = 0
	WEIGHT_EXTRALIGHT = 40
	WEIGHT_ULTRALIGHT = WEIGHT_EXTRALIGHT
	WEIGHT_LIGHT      = 50
	WEIGHT_DEMILIGHT  = 55
	WEIGHT_SEMILIGHT  = WEIGHT_DEMILIGHT
	WEIGHT_BOOK       = 75
	WEIGHT_REGULAR    = 80
	WEIGHT_NORMAL     = WEIGHT_REGULAR
	WEIGHT_MEDIUM     = 100
	WEIGHT_DEMIBOLD   = 180
	WEIGHT_SEMIBOLD   = WEIGHT_DEMIBOLD
	WEIGHT_BOLD       = 200
	WEIGHT_EXTRABOLD  = 205
	WEIGHT_ULTRABOLD  = WEIGHT_EXTRABOLD
	WEIGHT_BLACK      = 210
	WEIGHT_HEAVY      = WEIGHT_BLACK
	WEIGHT_EXTRABLACK = 215
	WEIGHT_ULTRABLACK = WEIGHT_EXTRABLACK
)

var weightMap = [...]struct {
	ot, fc float32
}{
	{0, WEIGHT_THIN},
	{100, WEIGHT_THIN},
	{200, WEIGHT_EXTRALIGHT},
	{300, WEIGHT_LIGHT},
	{350, WEIGHT_DEMILIGHT},
	{380, WEIGHT_BOOK},
	{400, WEIGHT_REGULAR},
	{500, WEIGHT_MEDIUM},
	{600, WEIGHT_DEMIBOLD},
	{700, WEIGHT_BOLD},
	{800, WEIGHT_EXTRABOLD},
	{900, WEIGHT_BLACK},
	{1000, WEIGHT_EXTRABLACK},
}

func lerp(x, x1, x2, y1, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1
	// assert(dx > 0 && dy >= 0 && x1 <= x && x <= x2)
	return y1 + (x-x1)*dy/dx
}

// WeightFromOT returns a float value
// to use with `WEIGHT`, from a float in the 1..1000 range, resembling
// the numbers from OpenType specification's OS/2 usWeight numbers, which
// are also similar to CSS font-weight numbers.
// If input is negative, zero, or greater than 1000, returns -1.
// This function linearly interpolates between various WEIGHT constants.
// As such, the returned value does not necessarily match any of the predefined constants.
func WeightFromOT(otWeight float32) float32 {
	if otWeight < 0 {
		return -1
	}

	otWeight = minF(otWeight, weightMap[len(weightMap)-1].ot)

	var i int
	for i = 1; otWeight > weightMap[i].ot; i++ {
	}

	if otWeight == weightMap[i].ot {
		return weightMap[i].fc
	}

	// interpolate between two items
	return lerp(otWeight, weightMap[i-1].ot, weightMap[i].ot, weightMap[i-1].fc, weightMap[i].fc)
}

// WeightToOT is the inverse of `WeightFromOT`.
// If the input is less than `WEIGHT_THIN` or greater than `WEIGHT_EXTRABLACK`, it returns -1.
// Otherwise returns a number in the range 1 to 1000.
func WeightToOT(fcWeight float32) float32 {
	if fcWeight < 0 || fcWeight > WEIGHT_EXTRABLACK {
		return -1
	}

	var i int
	for i = 1; fcWeight > weightMap[i].fc; i++ {
	}

	if fcWeight == weightMap[i].fc {
		return weightMap[i].ot
	}

	// interpolate between two items.
	return lerp(fcWeight, weightMap[i-1].fc, weightMap[i].fc, weightMap[i-1].ot, weightMap[i].ot)
}

var (
	defaultLangs     strSet
	defaultLangsLock sync.Mutex
)

// Returns a string set of the default languages according to the environment variables on the system.
// This function looks for them in order of FC_LANG, LC_ALL, LC_CTYPE and LANG then.
// If there are no valid values in those environment variables, "en" will be set as fallback.
// Thus, it always returns at least one language.
func getDefaultLangs() strSet {
	defaultLangsLock.Lock()
	defer defaultLangsLock.Unlock()

	if defaultLangs != nil {
		return defaultLangs
	}

	defaultLangs = make(strSet)

	langs := os.Getenv("LANG")
	if langs == "" {
		langs = os.Getenv("LC_ALL")
	}
	if langs == "" {
		langs = os.Getenv("LC_CTYPE")
	}
	if langs == "" {
		langs = os.Getenv("LANG")
	}
	if langs != "" {
		ok := addLangs(defaultLangs, langs)
		if !ok {
			defaultLangs["en"] = true
		}
	} else {
		defaultLangs["en"] = true
	}

	return defaultLangs
}

func getProgramName() string {
	e, _ := os.Executable()
	return e
}
