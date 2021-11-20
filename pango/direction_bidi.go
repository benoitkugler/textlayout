package pango

import "github.com/benoitkugler/textlayout/fribidi"

/**
 * SECTION:bidi
 * @short_description:Types and functions for bidirectional text
 * @title:Bidirectional Text
 * @see_also:
 * pango_context_get_baseDir(),
 * pango_context_set_baseDir(),
 * pango_itemize_with_baseDir()
 *
 * Pango supports bidirectional text (like Arabic and Hebrew) automatically.
 * Some applications however, need some help to correctly handle bidirectional text.
 *
 * The Direction type can be used with pango_context_set_baseDir() to
 * instruct Pango about direction of text, though in most cases Pango detects
 * that correctly and automatically.  The rest of the facilities in this section
 * are used internally by Pango already, and are provided to help applications
 * that need more direct control over bidirectional setting of text.
 */

// The Direction type represents a direction in the
// Unicode bidirectional algorithm.
//
// Not every value in this
// enumeration makes sense for every usage of Direction;
// for example, the return value of characterDirection()
// and findBaseDirection() cannot be `DIRECTION_WEAK_LTR`
// or `DIRECTION_WEAK_RTL`, since every character is either
// neutral or has a strong direction; on the other hand
// `DIRECTION_NEUTRAL` doesn't make sense to pass
// to pango_itemize_with_baseDir().
//
// See `Gravity` for how vertical text is handled in Pango.
//
// If you are interested in text direction, you should
// really use fribidi directly. Direction is only
// retained because it is used in some public apis.
type Direction uint8

const (
	DIRECTION_LTR      Direction = iota // A strong left-to-right direction
	DIRECTION_RTL                       // A strong right-to-left direction
	_                                   // Deprecated value; treated the same as `DIRECTION_RTL`.
	_                                   // Deprecated value; treated the same as `DIRECTION_LTR`
	DIRECTION_WEAK_LTR                  // A weak left-to-right direction
	DIRECTION_WEAK_RTL                  // A weak right-to-left direction
	DIRECTION_NEUTRAL                   // No direction specified
)

// String returns a nickname for the direction
func (d Direction) String() string {
	switch d {
	case DIRECTION_LTR:
		return "ltr"
	case DIRECTION_RTL:
		return "rtl"
	case DIRECTION_WEAK_LTR:
		return "weak_ltr"
	case DIRECTION_WEAK_RTL:
		return "weak_rtl"
	case DIRECTION_NEUTRAL:
		return "neutral"
	default:
		return "<invalid direction>"
	}
}

/**
 * findBaseDirection:
 * @text:   the text to process. Must be valid UTF-8
 * @length: length of @text in bytes (may be -1 if @text is nul-terminated)
 *
 * Searches a string the first character that has a strong
 * direction, according to the Unicode bidirectional algorithm.
 *
 * Return value: The direction corresponding to the first strong character.
 * If no such character is found, then `DIRECTION_NEUTRAL` is returned.
 *
 * Since: 1.4
 */
func findBaseDirection(text []rune) Direction {
	dir := DIRECTION_NEUTRAL
	for _, wc := range text {
		dir = characterDirection(wc)
		if dir != DIRECTION_NEUTRAL {
			break
		}
	}

	return dir
}

// characterDirection determines the inherent direction of a character; either
// `DIRECTION_LTR`, `DIRECTION_RTL`, or
// `DIRECTION_NEUTRAL`.
//
// This function is useful to categorize characters into left-to-right
// letters, right-to-left letters, and everything else.
func characterDirection(ch rune) Direction {
	fType := fribidi.GetBidiType(ch)
	if !fType.IsStrong() {
		return DIRECTION_NEUTRAL
	} else if fType.IsRtl() {
		return DIRECTION_RTL
	} else {
		return DIRECTION_LTR
	}
}

func (d Direction) directionSimple() int {
	switch d {
	case DIRECTION_LTR, DIRECTION_WEAK_LTR:
		return 1
	case DIRECTION_RTL, DIRECTION_WEAK_RTL:
		return -1
	default:
		return 0
	}
}

// pango_log2vis_get_embedding_levels returns the bidirectional embedding levels of the input paragraph
// as defined by the Unicode Bidirectional Algorithm available at:
//
//   http://www.unicode.org/reports/tr9/
//
// If the input base direction is a weak direction, the direction of the
// characters in the text will determine the final resolved direction.
// The embedding levels slice as one item per Unicode character.
func pango_log2vis_get_embedding_levels(text []rune, pbaseDir Direction) (Direction, []fribidi.Level) {
	var (
		oredTypes      fribidi.CharType
		andedStrongs   fribidi.CharType = fribidi.RLE
		fribidiBaseDir fribidi.ParType
	)

	// G_STATIC_ASSERT (sizeof (FriBidiLevel) == sizeof (guint8));
	// G_STATIC_ASSERT (sizeof (FriBidiChar) == sizeof (rune));

	switch pbaseDir {
	case DIRECTION_LTR:
		fribidiBaseDir = fribidi.LTR
	case DIRECTION_RTL:
		fribidiBaseDir = fribidi.RTL
	case DIRECTION_WEAK_RTL:
		fribidiBaseDir = fribidi.WRTL
	default:
		fribidiBaseDir = fribidi.WLTR
	}

	nChars := len(text)

	bidiTypes := make([]fribidi.CharType, nChars)
	bracketTypes := make([]fribidi.BracketType, nChars)

	for i, ch := range text {
		charType := fribidi.GetBidiType(ch)

		bidiTypes[i] = charType
		oredTypes |= charType
		if charType.IsStrong() {
			andedStrongs &= charType
		}
		if bidiTypes[i] == fribidi.ON {
			bracketTypes[i] = fribidi.GetBracket(ch)
		} else {
			bracketTypes[i] = fribidi.NoBracket
		}
	}

	var embeddingLevelsList []fribidi.Level

	// Short-circuit FriBidi call for unidirectional text.
	// For details see:
	// https://bugzilla.gnome.org/show_bug.cgi?id=590183

	// The case that all resolved levels will be ltr.
	// No isolates, all strongs be LTR, there should be no Arabic numbers
	// (or letters for that matter), and one of the following:
	// - baseDir doesn't have an RTL taste.
	// - there are letters, and baseDir is weak.
	if !oredTypes.IsIsolate() && !oredTypes.IsRtl() && !oredTypes.IsArabic() &&
		(!fribidiBaseDir.IsRtl() || (fribidiBaseDir.IsWeak() && oredTypes.IsLetter())) {
		// all LTR
		fribidiBaseDir = fribidi.LTR
		embeddingLevelsList = make([]fribidi.Level, nChars) // all zero
	} else if !oredTypes.IsIsolate() && !oredTypes.IsNumber() && andedStrongs.IsRtl() &&
		(fribidiBaseDir.IsRtl() || (fribidiBaseDir.IsWeak() && oredTypes.IsLetter())) {
		// The case that all resolved levels will be RTL is much more complex.
		// No isolates, no numbers, all strongs are RTL, and one of the following:
		// - baseDir has an RTL taste (may be weak).
		// - there are letters, and baseDir is weak.

		// all RTL
		fribidiBaseDir = fribidi.RTL
		embeddingLevelsList = make([]fribidi.Level, nChars) // all one
		for i := range embeddingLevelsList {
			embeddingLevelsList[i] = 1
		}
	} else {
		// full algorithm
		embeddingLevelsList, _ = fribidi.GetParEmbeddingLevels(bidiTypes, bracketTypes, &fribidiBaseDir)
	}

	pbaseDir = DIRECTION_RTL
	if fribidiBaseDir == fribidi.LTR {
		pbaseDir = DIRECTION_LTR
	}

	return pbaseDir, embeddingLevelsList
}
