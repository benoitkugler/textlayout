package fontconfig

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// test only: print debug information to stdout
	debugMode = true

	homeEnabled = true
	// FONTCONFIG_FILE is used to override the default configuration file.
	FONTCONFIG_FILE = "fonts.conf"
)

// FcConfigHome returns the current user's home directory, if it is available, and if using it
// is enabled, and "" otherwise.
func FcConfigHome() string {
	if homeEnabled {
		home := os.Getenv("HOME")

		if home == "" && runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		}

		return home
	}
	return ""
}

// FontSet contains a list of Patterns, containing the
// results of listing fonts.
type FontSet []Pattern

// toAbsPath constructs an absolute pathname from
// `s`. It converts any leading '~' characters in
// to the value of the HOME environment variable, and any relative paths are
// converted to absolute paths using the current working directory. Sequences
// of '/' characters are converted to a single '/', and names containing the
// current directory '.' or parent directory '..' are correctly reconstructed.
// Returns "" if '~' is the leading character and HOME is unset or disabled
func toAbsPath(s string) (string, error) {
	if usesHome(s) {
		home := FcConfigHome()
		if home == "" {
			return "", errors.New("home is disabled")
		}
		s = filepath.Join(home, s[1:])
	}
	return filepath.Abs(s)
}

type strSet map[string]bool

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

// clear all the entries
func (set strSet) reset() {
	for k := range set {
		delete(set, k)
	}
}

type FcStrList struct {
	set strSet
	n   int
}

type FcSetName uint8

const (
	FcSetSystem FcSetName = iota
	FcSetApplication
)

type FcResult uint8

const (
	FcResultMatch FcResult = iota
	FcResultNoMatch
	FcResultTypeMismatch
	FcResultNoId
	FcResultOutOfMemory
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

	FC_PROPORTIONAL = 0
	FC_DUAL         = 90
	FC_MONO         = 100
	FC_CHARCELL     = 110

	/* sub-pixel order */
	FC_RGBA_UNKNOWN = 0
	FC_RGBA_RGB     = 1
	FC_RGBA_BGR     = 2
	FC_RGBA_VRGB    = 3
	FC_RGBA_VBGR    = 4
	FC_RGBA_NONE    = 5

	/* hinting style */
	FC_HINT_NONE   = 0
	FC_HINT_SLIGHT = 1
	FC_HINT_MEDIUM = 2
	FC_HINT_FULL   = 3

	/* LCD filter */
	FC_LCD_NONE    = 0
	FC_LCD_DEFAULT = 1
	FC_LCD_LIGHT   = 2
	FC_LCD_LEGACY  = 3
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
	ot, fc float64
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

func lerp(x, x1, x2, y1, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	// assert(dx > 0 && dy >= 0 && x1 <= x && x <= x2)
	return y1 + (x-x1)*dy/dx
}

func FcWeightFromOpenTypeDouble(otWeight float64) float64 {
	if otWeight < 0 {
		return -1
	}

	otWeight = math.Min(otWeight, weightMap[len(weightMap)-1].ot)

	var i int
	for i = 1; otWeight > weightMap[i].ot; i++ {
	}

	if otWeight == weightMap[i].ot {
		return weightMap[i].fc
	}

	// interpolate between two items
	return lerp(otWeight, weightMap[i-1].ot, weightMap[i].ot, weightMap[i-1].fc, weightMap[i].fc)
}

func FcWeightToOpenTypeDouble(fcWeight float64) float64 {
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

// always returns at least one language
func FcGetDefaultLangs() map[string]bool {
	// TODO: the C implementation caches the result

	result := make(map[string]bool)

	langs := os.Getenv("FC_LANG")
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
		ok := addLangs(result, langs)
		if !ok {
			result["en"] = true
		}
	} else {
		result["en"] = true
	}

	return result
}

func getProgramName() string {
	e, _ := os.Executable()
	return e
}
