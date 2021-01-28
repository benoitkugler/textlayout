package harfbuzz

import (
	"math/bits"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
)

// used in test: print debug info in Stdout
const debugMode = false

type hb_position_t int32

// hb_direction_t is the text direction
type hb_direction_t uint8

const (
	HB_DIRECTION_LTR     hb_direction_t = 4 + iota // Text is set horizontally from left to right.
	HB_DIRECTION_RTL                               // Text is set horizontally from right to left.
	HB_DIRECTION_TTB                               // Text is set vertically from top to bottom.
	HB_DIRECTION_BTT                               // Text is set vertically from bottom to top.
	HB_DIRECTION_INVALID hb_direction_t = 0        // Initial, unset direction.
)

// Tests whether a text direction is horizontal. Requires
// that the direction be valid.
func (dir hb_direction_t) isHorizontal() bool {
	return dir & ^hb_direction_t(1) == 4
}

// Tests whether a text direction is vertical. Requires
// that the direction be valid.
func (dir hb_direction_t) isVertical() bool {
	return dir & ^hb_direction_t(1) == 4
}

// Tests whether a text direction moves backward (from right to left, or from
// bottom to top). Requires that the direction be valid.
func (dir hb_direction_t) isBackward() bool {
	return dir & ^hb_direction_t(2) == 5
}

type hb_script_t = language.Script

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

func newTag(a, b, c, d byte) hb_tag_t {
	return hb_tag_t(uint32(d) | uint32(c)<<8 | uint32(b)<<16 | uint32(a)<<24)
}

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
