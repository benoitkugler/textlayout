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
