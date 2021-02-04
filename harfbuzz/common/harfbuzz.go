package common

import (
	"math/bits"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
)

// DebugMode is only used in test: when `true`, it prints debug info in Stdout.
const DebugMode = false

type Position = int32

// Direction is the text direction
type Direction uint8

const (
	HB_DIRECTION_LTR     Direction = 4 + iota // Text is set horizontally from left to right.
	HB_DIRECTION_RTL                          // Text is set horizontally from right to left.
	HB_DIRECTION_TTB                          // Text is set vertically from top to bottom.
	HB_DIRECTION_BTT                          // Text is set vertically from bottom to top.
	HB_DIRECTION_INVALID Direction = 0        // Initial, unset direction.
)

// Tests whether a text direction is horizontal. Requires
// that the direction be valid.
func (dir Direction) isHorizontal() bool {
	return dir & ^Direction(1) == 4
}

// Tests whether a text direction is vertical. Requires
// that the direction be valid.
func (dir Direction) isVertical() bool {
	return dir & ^Direction(1) == 4
}

// Tests whether a text direction moves backward (from right to left, or from
// bottom to top). Requires that the direction be valid.
func (dir Direction) isBackward() bool {
	return dir & ^Direction(2) == 5
}

// Reverses a text direction. Requires that the direction
// be valid.
func (dir Direction) reverse() Direction {
	return dir ^ 1
}

type hb_script_t = language.Script

// Fetches the `Direction` of a script when it is
// set horizontally. All right-to-left scripts will return
// `HB_DIRECTION_RTL`. All left-to-right scripts will return
// `HB_DIRECTION_LTR`.  Scripts that can be written either
// horizontally or vertically will return `HB_DIRECTION_INVALID`.
// Unknown scripts will return `HB_DIRECTION_LTR`.
func hb_script_get_horizontal_direction(script hb_script_t) Direction {
	/* https://docs.google.com/spreadsheets/d/1Y90M0Ie3MUJ6UVCRDOypOtijlMDLNNyyLk36T6iMu0o */
	switch script {
	case language.Arabic, language.Hebrew, language.Syriac, language.Thaana,
		language.Cypriot, language.Kharoshthi, language.Phoenician, language.Nko, language.Lydian,
		language.Avestan, language.Imperial_Aramaic, language.Inscriptional_Pahlavi, language.Inscriptional_Parthian, language.Old_South_Arabian, language.Old_Turkic,
		language.Samaritan, language.Mandaic, language.Meroitic_Cursive, language.Meroitic_Hieroglyphs, language.Manichaean, language.Mende_Kikakui,
		language.Nabataean, language.Old_North_Arabian, language.Palmyrene, language.Psalter_Pahlavi, language.Hatran, language.Adlam, language.Hanifi_Rohingya,
		language.Old_Sogdian, language.Sogdian, language.Elymaic, language.Chorasmian, language.Yezidi:

		return HB_DIRECTION_RTL

	/* https://github.com/harfbuzz/harfbuzz/issues/1000 */
	case language.Old_Hungarian, language.Old_Italic, language.Runic:

		return HB_DIRECTION_INVALID
	}

	return HB_DIRECTION_LTR
}

// store the canonicalized BCP 47 tag
type hb_language_t string

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func max32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func isAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isAlnum(c byte) bool { return isAlpha(c) || (c >= '0' && c <= '9') }
func toUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}

const maxInt = int(^uint32(0) >> 1)

type glyphIndex uint16

type hb_tag_t = truetype.Tag

// hb_feature_t holds information about requested
// feature application. The feature will be applied with the given value to all
// glyphs which are in clusters between `start` (inclusive) and `end` (exclusive).
// Setting start to `HB_FEATURE_GLOBAL_START` and end to `HB_FEATURE_GLOBAL_END`
// specifies that the feature always applies to the entire buffer.
type hb_feature_t struct {
	tag hb_tag_t
	// value of the feature: 0 disables the feature, non-zero (usually
	// 1) enables the feature. For features implemented as lookup type 3 (like
	// 'salt') `value` is a one-based index into the alternates.
	value uint32
	// the cluster to start applying this feature setting (inclusive)
	start int
	// the cluster to end applying this feature setting (exclusive)
	end int
}

const (
	// Special setting for `hb_feature_t.start` to apply the feature from the start
	// of the buffer.
	HB_FEATURE_GLOBAL_START = 0
	// Special setting for `hb_feature_t.end` to apply the feature from to the end
	// of the buffer.
	HB_FEATURE_GLOBAL_END = int(^uint(0) >> 1)
)

// returns the number of bits needed to store number
func hb_bit_storage(v uint32) int { return 32 - bits.LeadingZeros32(v) }
