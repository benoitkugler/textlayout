package harfbuzz

import (
	"math"
	"math/bits"

	tt "github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
)

// debugMode is only used in test: when `true`, it prints debug info in Stdout.
const debugMode = false

// Direction is the text direction.
// The zero value is the initial, unset, invalid direction.
type Direction uint8

const (
	LeftToRight Direction = 4 + iota // Text is set horizontally from left to right.
	RightToLeft                      // Text is set horizontally from right to left.
	TopToBottom                      // Text is set vertically from top to bottom.
	BottomToTop                      // Text is set vertically from bottom to top.
)

// Tests whether a text direction is horizontal. Requires
// that the direction be valid.
func (dir Direction) isHorizontal() bool { return dir & ^Direction(1) == 4 }

// Tests whether a text direction is vertical. Requires
// that the direction be valid.
func (dir Direction) isVertical() bool { return dir & ^Direction(1) == 6 }

// Tests whether a text direction moves backward (from right to left, or from
// bottom to top). Requires that the direction be valid.
func (dir Direction) IsBackward() bool { return dir & ^Direction(2) == 5 }

// Tests whether a text direction moves forward (from left to right, or from
// top to bottom). Requires that the direction be valid.
func (dir Direction) isForward() bool { return dir & ^Direction(2) == 4 }

// Reverses a text direction. Requires that the direction
// be valid.
func (dir Direction) reverse() Direction {
	return dir ^ 1
}

// SegmentProperties holds various text properties of a `Buffer`.
type SegmentProperties struct {
	// Direction is the text flow direction of the buffer. No shaping can happen without
	// setting direction, and it controls the visual direction for the
	// output glyphs; for RTL direction the glyphs will be reversed. Many layout
	// features depend on the proper setting of the direction, for example,
	// reversing RTL text before shaping, then shaping with LTR direction is not
	// the same as keeping the text in logical order and shaping with RTL
	// direction.
	Direction Direction

	// Script is crucial for choosing the proper shaping behaviour for scripts that
	// require it (e.g. Arabic) and the OpenType features defined in the font
	// to be applied.
	//
	// See the package language for predefined values.
	Script language.Script

	// Languages are crucial for selecting which OpenType feature to apply to the
	// buffer which can result in applying language-specific behaviour. Languages
	// are orthogonal to the scripts, and though they are related, they are
	// different concepts and should not be confused with each other.
	Language Language
}

// Flags controls some fine tunning of the shaping
// (see the constants).
type Flags uint16

const (
	// Flag indicating that special handling of the beginning
	// of text paragraph can be applied to this buffer. Should usually
	// be set, unless you are passing to the buffer only part
	// of the text without the full context.
	Bot Flags = 1 << iota
	// Flag indicating that special handling of the end of text
	// paragraph can be applied to this buffer, similar to
	// `Bot`.
	Eot
	// Flag indication that character with Default_Ignorable
	// Unicode property should use the corresponding glyph
	// from the font, instead of hiding them (done by
	// replacing them with the space glyph and zeroing the
	// advance width.)  This flag takes precedence over
	// `RemoveDefaultIgnorables`.
	PreserveDefaultIgnorables
	// Flag indication that character with Default_Ignorable
	// Unicode property should be removed from glyph string
	// instead of hiding them (done by replacing them with the
	// space glyph and zeroing the advance width.)
	// `PreserveDefaultIgnorables` takes
	// precedence over this flag.
	RemoveDefaultIgnorables
	// Flag indicating that a dotted circle should
	// not be inserted in the rendering of incorrect
	// character sequences (such at <0905 093E>).
	DoNotinsertDottedCircle
)

// ClusterLevel allows selecting more fine-grained Cluster handling.
// It defaults to `MonotoneGraphemes`.
type ClusterLevel uint8

const (
	//  Return cluster values grouped into monotone order.
	MonotoneCharacters ClusterLevel = iota
	// Return cluster values grouped by graphemes into monotone order.
	MonotoneGraphemes
	// Don't group cluster values.
	Characters
)

// Feature holds information about requested
// feature application. The feature will be applied with the given value to all
// glyphs which are in clusters between `start` (inclusive) and `end` (exclusive).
// Setting start to `FeatureGlobalStart` and end to `FeatureGlobalEnd`
// specifies that the feature always applies to the entire buffer.
type Feature struct {
	Tag tt.Tag
	// Value of the feature: 0 disables the feature, non-zero (usually
	// 1) enables the feature. For features implemented as lookup type 3 (like
	// 'salt') `Value` is a one-based index into the alternates.
	Value uint32
	// The cluster to Start applying this feature setting (inclusive)
	Start int
	// The cluster to End applying this feature setting (exclusive)
	End int
}

const (
	// Special setting for `Feature.Start` to apply the feature from the start
	// of the buffer.
	FeatureGlobalStart = 0
	// Special setting for `Feature.End` to apply the feature from to the end
	// of the buffer.
	FeatureGlobalEnd = int(^uint(0) >> 1)
)

type Position = int32

type hb_script_t = language.Script

// Fetches the `Direction` of a script when it is
// set horizontally. All right-to-left scripts will return
// `RightToLeft`. All left-to-right scripts will return
// `LeftToRight`.  Scripts that can be written either
// horizontally or vertically will return `Invalid`.
// Unknown scripts will return `LeftToRight`.
func getHorizontalDirection(script hb_script_t) Direction {
	/* https://docs.google.com/spreadsheets/d/1Y90M0Ie3MUJ6UVCRDOypOtijlMDLNNyyLk36T6iMu0o */
	switch script {
	case language.Arabic, language.Hebrew, language.Syriac, language.Thaana,
		language.Cypriot, language.Kharoshthi, language.Phoenician, language.Nko, language.Lydian,
		language.Avestan, language.Imperial_Aramaic, language.Inscriptional_Pahlavi, language.Inscriptional_Parthian, language.Old_South_Arabian, language.Old_Turkic,
		language.Samaritan, language.Mandaic, language.Meroitic_Cursive, language.Meroitic_Hieroglyphs, language.Manichaean, language.Mende_Kikakui,
		language.Nabataean, language.Old_North_Arabian, language.Palmyrene, language.Psalter_Pahlavi, language.Hatran, language.Adlam, language.Hanifi_Rohingya,
		language.Old_Sogdian, language.Sogdian, language.Elymaic, language.Chorasmian, language.Yezidi:

		return RightToLeft

	/* https://github.com/harfbuzz/harfbuzz/issues/1000 */
	case language.Old_Hungarian, language.Old_Italic, language.Runic:
		return 0
	}

	return LeftToRight
}

// store the canonicalized BCP 47 tag
type Language string

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func min8(a, b uint8) uint8 {
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

func Max32(a, b uint32) uint32 {
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

// bitStorage returns the number of bits needed to store the number.
func bitStorage(v uint32) int { return 32 - bits.LeadingZeros32(v) }

func roundf(f float32) Position {
	return Position(math.Round(float64(f)))
}
