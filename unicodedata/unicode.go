// Package unicodedata provides additional lookup functions for unicode
// properties, not covered by the standard package unicode.
package unicodedata

import (
	"unicode"
)

// LookupCombiningClass returns the class used for the Canonical Ordering Algorithm in the Unicode Standard,
// defaulting to 0.
//
// From http://www.unicode.org/reports/tr44/#Canonical_Combining_Class:
// "This property could be considered either an enumerated property or a numeric property:
// the principal use of the property is in terms of the numeric values.
// For the property value names associated with different numeric values,
// see DerivedCombiningClass.txt and Canonical Combining Class Values."
func LookupCombiningClass(ch rune) uint8 {
	for i, t := range combiningClasses {
		if t == nil {
			continue
		}
		if unicode.Is(t, ch) {
			return uint8(i)
		}
	}
	return 0
}

// LookupMirrorChar finds the mirrored equivalent of a character as defined in
// the file BidiMirroring.txt of the Unicode Character Database available at
// http://www.unicode.org/Public/UNIDATA/BidiMirroring.txt.
//
// If the input character is declared as a mirroring character in the
// Unicode standard and has a mirrored equivalent, it is returned with `true`.
// Otherwise the input character itself returned with `false`.
func LookupMirrorChar(ch rune) (rune, bool) {
	m, ok := mirroring[ch]
	if !ok {
		m = ch
	}
	return m, ok
}

// Algorithmic hangul syllable [de]composition
const (
	HangulSBase  = 0xAC00
	HangulLBase  = 0x1100
	HangulVBase  = 0x1161
	HangulTBase  = 0x11A7
	HangulSCount = 11172
	HangulLCount = 19
	HangulVCount = 21
	HangulTCount = 28
	HangulNCount = HangulVCount * HangulTCount
)

func decomposeHangul(ab rune) (a, b rune, ok bool) {
	si := ab - HangulSBase

	if si < 0 || si >= HangulSCount {
		return 0, 0, false
	}

	if si%HangulTCount != 0 { /* LV,T */
		return HangulSBase + (si/HangulTCount)*HangulTCount, HangulTBase + (si % HangulTCount), true
	} /* L,V */
	return HangulLBase + (si / HangulNCount), HangulVBase + (si%HangulNCount)/HangulTCount, true
}

func composeHangul(a, b rune) (rune, bool) {
	if a >= HangulSBase && a < (HangulSBase+HangulSCount) && b > HangulTBase && b < (HangulTBase+HangulTCount) && (a-HangulSBase)%HangulTCount == 0 {
		/* LV,T */
		return a + (b - HangulTBase), true
	} else if a >= HangulLBase && a < (HangulLBase+HangulLCount) && b >= HangulVBase && b < (HangulVBase+HangulVCount) {
		/* L,V */
		li := a - HangulLBase
		vi := b - HangulVBase
		return HangulSBase + li*HangulNCount + vi*HangulTCount, true
	}
	return 0, false
}

// Decompose decompose an input Unicode code point,
// returning the two decomposed code points, if successful.
// It returns `false` otherwise.
func Decompose(ab rune) (a, b rune, ok bool) {
	if a, b, ok = decomposeHangul(ab); ok {
		return a, b, true
	}
	if m1, ok := decompose1[ab]; ok {
		return m1, 0, true
	}
	if m2, ok := decompose2[ab]; ok {
		return m2[0], m2[1], true
	}
	return ab, 0, false
}

// Compose composes a sequence of two input Unicode code
// points by canonical equivalence, returning the composed code, if successful.
// It returns `false` otherwise
func Compose(a, b rune) (rune, bool) {
	if ab, ok := composeHangul(a, b); ok {
		return ab, true
	}
	u := compose[[2]rune{a, b}]
	return u, u != 0
}

// ArabicJoining is a property used to shape Arabic runes.
// See the table ArabicJoinings.
type ArabicJoining byte

const (
	U          ArabicJoining = 'U' // Un-joining, e.g. Full Stop
	R          ArabicJoining = 'R' // Right-joining, e.g. Arabic Letter Dal
	Alaph      ArabicJoining = 'a' // Alaph group (included in kind R)
	DalathRish ArabicJoining = 'd' // Dalat Rish group (included in kind R)
	D          ArabicJoining = 'D' // Dual-joining, e.g. Arabic Letter Ain
	C          ArabicJoining = 'C' // Join-Causing, e.g. Tatweel, ZWJ
	L          ArabicJoining = 'L' // Left-joining, i.e. fictional
	T          ArabicJoining = 'T' // Transparent, e.g. Arabic Fatha
	G          ArabicJoining = 'G' // Ignored, e.g. LRE, RLE, ZWNBSP
)
