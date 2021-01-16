package fontconfig

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
)

const (
	debugMode   = false
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

type FcFontSet []FcPattern // with length nfont, and cap sfont

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

type FcStrSet map[string]bool

// Returns whether `a` contains precisely the same
// strings as `b`. Ordering of strings within the two
// sets is not considered.
func FcStrSetEqual(a, b FcStrSet) bool {
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

func (set FcStrSet) reset() {
	for k := range set {
		delete(set, k)
	}
}

type FcStrList struct {
	set FcStrSet
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
	FC_SLANT_ROMAN   = 0
	FC_SLANT_ITALIC  = 100
	FC_SLANT_OBLIQUE = 110

	FC_WIDTH_ULTRACONDENSED = 50
	FC_WIDTH_EXTRACONDENSED = 63
	FC_WIDTH_CONDENSED      = 75
	FC_WIDTH_SEMICONDENSED  = 87
	FC_WIDTH_NORMAL         = 100
	FC_WIDTH_SEMIEXPANDED   = 113
	FC_WIDTH_EXPANDED       = 125
	FC_WIDTH_EXTRAEXPANDED  = 150
	FC_WIDTH_ULTRAEXPANDED  = 200

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
	FC_WEIGHT_THIN       = 0
	FC_WEIGHT_EXTRALIGHT = 40
	FC_WEIGHT_ULTRALIGHT = FC_WEIGHT_EXTRALIGHT
	FC_WEIGHT_LIGHT      = 50
	FC_WEIGHT_DEMILIGHT  = 55
	FC_WEIGHT_SEMILIGHT  = FC_WEIGHT_DEMILIGHT
	FC_WEIGHT_BOOK       = 75
	FC_WEIGHT_REGULAR    = 80
	FC_WEIGHT_NORMAL     = FC_WEIGHT_REGULAR
	FC_WEIGHT_MEDIUM     = 100
	FC_WEIGHT_DEMIBOLD   = 180
	FC_WEIGHT_SEMIBOLD   = FC_WEIGHT_DEMIBOLD
	FC_WEIGHT_BOLD       = 200
	FC_WEIGHT_EXTRABOLD  = 205
	FC_WEIGHT_ULTRABOLD  = FC_WEIGHT_EXTRABOLD
	FC_WEIGHT_BLACK      = 210
	FC_WEIGHT_HEAVY      = FC_WEIGHT_BLACK
	FC_WEIGHT_EXTRABLACK = 215
	FC_WEIGHT_ULTRABLACK = FC_WEIGHT_EXTRABLACK
)

var weightMap = [...]struct {
	ot, fc float64
}{
	{0, FC_WEIGHT_THIN},
	{100, FC_WEIGHT_THIN},
	{200, FC_WEIGHT_EXTRALIGHT},
	{300, FC_WEIGHT_LIGHT},
	{350, FC_WEIGHT_DEMILIGHT},
	{380, FC_WEIGHT_BOOK},
	{400, FC_WEIGHT_REGULAR},
	{500, FC_WEIGHT_MEDIUM},
	{600, FC_WEIGHT_DEMIBOLD},
	{700, FC_WEIGHT_BOLD},
	{800, FC_WEIGHT_EXTRABOLD},
	{900, FC_WEIGHT_BLACK},
	{1000, FC_WEIGHT_EXTRABLACK},
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
	if fcWeight < 0 || fcWeight > FC_WEIGHT_EXTRABLACK {
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
		if ok := FcStrSetAddLangs(result, langs); !ok {
			result["en"] = true
		}
	} else {
		result["en"] = true
	}

	return result
}

func FcGetPrgname() string {
	e, _ := os.Executable()
	return e
}
