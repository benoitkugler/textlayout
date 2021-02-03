// Package unicodedata provides additional lookup functions for unicode
// properties, not covered by the standard package unicode.
package unicodedata

import "unicode"

// LookupCombiningClass returns the class used for the Canonical Ordering Algorithm in the Unicode Standard,
// or 255 if not found.
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
	return 255
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

const (
	hangulSBASE  = 0xAC00
	hangulLBASE  = 0x1100
	hangulVBASE  = 0x1161
	hangulTBASE  = 0x11A7
	hangulSCOUNT = 11172
	hangulLCOUNT = 19
	hangulVCOUNT = 21
	hangulTCOUNT = 28
	hangulNCOUNT = hangulVCOUNT * hangulTCOUNT
)

func decomposeHangul(ab rune) (a, b rune, ok bool) {
	si := ab - hangulSBASE

	if si >= hangulSCOUNT {
		return 0, 0, false
	}

	if si%hangulTCOUNT != 0 { /* LV,T */
		return hangulSBASE + (si/hangulTCOUNT)*hangulTCOUNT, hangulTBASE + (si % hangulTCOUNT), true
	} /* L,V */
	return hangulLBASE + (si / hangulNCOUNT), hangulVBASE + (si%hangulNCOUNT)/hangulTCOUNT, true
}

func composeHangul(a, b rune) (rune, bool) {
	if a >= hangulSBASE && a < (hangulSBASE+hangulSCOUNT) && b > hangulTBASE && b < (hangulTBASE+hangulTCOUNT) && (a-hangulSBASE)%hangulTCOUNT == 0 {
		/* LV,T */
		return a + (b - hangulTBASE), true
	} else if a >= hangulLBASE && a < (hangulLBASE+hangulLCOUNT) && b >= hangulVBASE && b < (hangulVBASE+hangulVCOUNT) {
		/* L,V */
		li := a - hangulLBASE
		vi := b - hangulVBASE
		return hangulSBASE + li*hangulNCOUNT + vi*hangulTCOUNT, true
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
	return 0, 0, false
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
