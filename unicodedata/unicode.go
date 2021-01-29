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
