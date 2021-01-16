package pango

import "github.com/benoitkugler/go-weasyprint/fribidi"

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
// for example, the return value of pango_unichar_direction()
// and pango_find_baseDir() cannot be `PANGO_DIRECTION_WEAK_LTR`
// or `PANGO_DIRECTION_WEAK_RTL`, since every character is either
// neutral or has a strong direction; on the other hand
// `PANGO_DIRECTION_NEUTRAL` doesn't make sense to pass
// to pango_itemize_with_baseDir().
//
// See `Gravity` for how vertical text is handled in Pango.
//
// If you are interested in text direction, you should
// really use fribidi directly. Direction is only
// retained because it is used in some public apis.
type Direction uint8

const (
	PANGO_DIRECTION_LTR      Direction = iota // A strong left-to-right direction
	PANGO_DIRECTION_RTL                       // A strong right-to-left direction
	_                                         // Deprecated value; treated the same as `PANGO_DIRECTION_RTL`.
	_                                         // Deprecated value; treated the same as `PANGO_DIRECTION_LTR`
	PANGO_DIRECTION_WEAK_LTR                  // A weak left-to-right direction
	PANGO_DIRECTION_WEAK_RTL                  // A weak right-to-left direction
	PANGO_DIRECTION_NEUTRAL                   // No direction specified
)

/**
 * pango_find_baseDir:
 * @text:   the text to process. Must be valid UTF-8
 * @length: length of @text in bytes (may be -1 if @text is nul-terminated)
 *
 * Searches a string the first character that has a strong
 * direction, according to the Unicode bidirectional algorithm.
 *
 * Return value: The direction corresponding to the first strong character.
 * If no such character is found, then `PANGO_DIRECTION_NEUTRAL` is returned.
 *
 * Since: 1.4
 */
func pango_find_base_dir(text []rune) Direction {
	dir := PANGO_DIRECTION_NEUTRAL
	for _, wc := range text {
		dir = pango_unichar_direction(wc)
		if dir != PANGO_DIRECTION_NEUTRAL {
			break
		}
	}

	return dir
}

// pango_unichar_direction determines the inherent direction of a character; either
// `PANGO_DIRECTION_LTR`, `PANGO_DIRECTION_RTL`, or
// `PANGO_DIRECTION_NEUTRAL`.
//
// This function is useful to categorize characters into left-to-right
// letters, right-to-left letters, and everything else.
func pango_unichar_direction(ch rune) Direction {
	fType := fribidi.GetBidiType(ch)
	if !fType.IsStrong() {
		return PANGO_DIRECTION_NEUTRAL
	} else if fType.IsRtl() {
		return PANGO_DIRECTION_RTL
	} else {
		return PANGO_DIRECTION_LTR
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
	case PANGO_DIRECTION_LTR:
		fribidiBaseDir = fribidi.LTR
	case PANGO_DIRECTION_RTL:
		fribidiBaseDir = fribidi.RTL
	case PANGO_DIRECTION_WEAK_RTL:
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

	pbaseDir = PANGO_DIRECTION_RTL
	if fribidiBaseDir == fribidi.LTR {
		pbaseDir = PANGO_DIRECTION_LTR
	}

	return pbaseDir, embeddingLevelsList
}

//   /**
//    * pango_unichar_direction:
//    * @ch: a Unicode character
//    *
//    * Determines the inherent direction of a character; either
//    * %PANGO_DIRECTION_LTR, %PANGO_DIRECTION_RTL, or
//    * %PANGO_DIRECTION_NEUTRAL.
//    *
//    * This function is useful to categorize characters into left-to-right
//    * letters, right-to-left letters, and everything else.  If full
//    * Unicode bidirectional type of a character is needed,
//    * pango_bidi_type_for_unichar() can be used instead.
//    *
//    * Return value: the direction of the character.
//    */
//   PangoDirection
//   pango_unichar_direction (rune ch)
//   {
// 	FriBidiCharType fribidi_ch_type;

// 	G_STATIC_ASSERT (sizeof (FriBidiChar) == sizeof (rune));

// 	fribidi_ch_type = fribidi.GetBidiType (ch);

// 	if (!FRIBIDI_IS_STRONG (fribidi_ch_type))
// 	  return PANGO_DIRECTION_NEUTRAL;
// 	else if (FRIBIDI_IS_RTL (fribidi_ch_type))
// 	  return PANGO_DIRECTION_RTL;
// 	else
// 	  return PANGO_DIRECTION_LTR;
//   }

//   /**
//    * pango_get_mirror_char:
//    * @ch: a Unicode character
//    * @mirrored_ch: location to store the mirrored character
//    *
//    * If @ch has the Unicode mirrored property and there is another Unicode
//    * character that typically has a glyph that is the mirror image of @ch's
//    * glyph, puts that character in the address pointed to by @mirrored_ch.
//    *
//    * Use g_unichar_get_mirror_char() instead; the docs for that function
//    * provide full details.
//    *
//    * Return value: %TRUE if @ch has a mirrored character and @mirrored_ch is
//    * filled in, %FALSE otherwise
//    **/
//   gboolean
//   pango_get_mirror_char (rune        ch,
// 				 rune       *mirrored_ch)
//   {
// 	return g_unichar_get_mirror_char (ch, mirrored_ch);
//   }
