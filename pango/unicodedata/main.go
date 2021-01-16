package unicodedata

import (
	"unicode"
)

// An enum that works as the states of the Hangul syllables system.
type JamoType int8

const (
	JAMO_LV  JamoType = iota /* break HANGUL_LV_SYLLABLE */
	JAMO_LVT                 /* break HANGUL_LVT_SYLLABLE */
	JAMO_L                   /* break HANGUL_L_JAMO */
	JAMO_V                   /* break HANGUL_V_JAMO */
	JAMO_T                   /* break HANGUL_T_JAMO */
	NO_JAMO                  /* Other */
)

// There are Hangul syllables encoded as characters, that act like a
// sequence of Jamos. For each character we define a JamoType
// that the character starts with, and one that it ends with.  This
// decomposes JAMO_LV and JAMO_LVT to simple other JAMOs.  So for
// example, a character with LineBreak type
// BreakH2 has start=JAMO_L and end=JAMO_V.
type charJamoProps struct {
	Start, End JamoType
}

/* Map from JamoType to CharJamoProps that hold only simple
 * JamoTypes (no LV or LVT) or none.
 */
var HangulJamoProps = [6]charJamoProps{
	{Start: JAMO_L, End: JAMO_V},   /* JAMO_LV */
	{Start: JAMO_L, End: JAMO_T},   /* JAMO_LVT */
	{Start: JAMO_L, End: JAMO_L},   /* JAMO_L */
	{Start: JAMO_V, End: JAMO_V},   /* JAMO_V */
	{Start: JAMO_T, End: JAMO_T},   /* JAMO_T */
	{Start: NO_JAMO, End: NO_JAMO}, /* NO_JAMO */
}

func IsEmoji(r rune) bool {
	return unicode.Is(_Emoji, r)
}
func IsEmojiPresentation(r rune) bool {
	return unicode.Is(_Emoji_Presentation, r)
}
func IsEmojiModifier(r rune) bool {
	return unicode.Is(_Emoji_Modifier, r)
}
func IsEmojiModifierBase(r rune) bool {
	return unicode.Is(_Emoji_Modifier_Base, r)
}
func IsEmojiExtendedPictographic(r rune) bool {
	return unicode.Is(_Extended_Pictographic, r)
}

func IsEmojiBaseCharacter(r rune) bool {
	return unicode.Is(_Emoji, r)
}

func IsVirama(r rune) bool {
	return unicode.Is(_Virama, r)
}

func IsVowelDependent(r rune) bool {
	return unicode.Is(_Vowel_Dependent, r)
}

func BreakClass(r rune) (string, *unicode.RangeTable) {
	for name, class := range Breaks {
		if unicode.Is(class, r) {
			return name, class
		}
	}
	return "", BreakXX
}

// Jamo returns the Jamo Type of `btype` or NO_JAMO
func Jamo(btype *unicode.RangeTable) JamoType {
	switch btype {
	case BreakH2:
		return JAMO_LV
	case BreakH3:
		return JAMO_LVT
	case BreakJL:
		return JAMO_L
	case BreakJV:
		return JAMO_V
	case BreakJT:
		return JAMO_T
	default:
		return NO_JAMO
	}
}
