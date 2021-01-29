package harfbuzz

import (
	"unicode"

	"github.com/benoitkugler/textlayout/unicodedata"
)

var uni = hb_unicode_funcs_t{}

// enum value to allow compact storage (see generalCategories)
type generalCategory uint8

const (
	unassigned generalCategory = iota
	control
	format
	privateUse
	surrogate
	lowercaseLetter
	modifierLetter
	otherLetter
	titlecaseLetter
	uppercaseLetter
	spacingMark
	enclosingMark
	nonSpacingMark
	decimalNumber
	letterNumber
	otherNumber
	connectPunctuation
	dashPunctuation
	closePunctuation
	finalPunctuation
	initialPunctuation
	otherPunctuation
	openPunctuation
	currencySymbol
	modifierSymbol
	mathSymbol
	otherSymbol
	lineSeparator
	paragraphSeparator
	spaceSeparator
)

// correspondance with *unicode.RangeTable classes
var generalCategories = [...]*unicode.RangeTable{
	unassigned:         nil,
	control:            unicode.Cc,
	format:             unicode.Cf,
	privateUse:         unicode.Co,
	surrogate:          unicode.Cs,
	lowercaseLetter:    unicode.Ll,
	modifierLetter:     unicode.Lm,
	otherLetter:        unicode.Lo,
	titlecaseLetter:    unicode.Lt,
	uppercaseLetter:    unicode.Lu,
	spacingMark:        unicode.Mc,
	enclosingMark:      unicode.Me,
	nonSpacingMark:     unicode.Mn,
	decimalNumber:      unicode.Nd,
	letterNumber:       unicode.Nl,
	otherNumber:        unicode.No,
	connectPunctuation: unicode.Pc,
	dashPunctuation:    unicode.Pd,
	closePunctuation:   unicode.Pe,
	finalPunctuation:   unicode.Pf,
	initialPunctuation: unicode.Pi,
	otherPunctuation:   unicode.Po,
	openPunctuation:    unicode.Ps,
	currencySymbol:     unicode.Sc,
	modifierSymbol:     unicode.Sk,
	mathSymbol:         unicode.Sm,
	otherSymbol:        unicode.So,
	lineSeparator:      unicode.Zl,
	paragraphSeparator: unicode.Zp,
	spaceSeparator:     unicode.Zs,
}

func (g generalCategory) isMark() bool {
	return g == spacingMark || g == enclosingMark || g == nonSpacingMark
}

// Modified combining marks
const (
	/* Hebrew
	 *
	 * We permute the "fixed-position" classes 10-26 into the order
	 * described in the SBL Hebrew manual:
	 *
	 * https://www.sbl-site.org/Fonts/SBLHebrewUserManual1.5x.pdf
	 *
	 * (as recommended by:
	 *  https://forum.fontlab.com/archive-old-microsoft-volt-group/vista-and-diacritic-ordering/msg22823/)
	 *
	 * More details here:
	 * https://bugzilla.mozilla.org/show_bug.cgi?id=662055
	 */
	mcc10 = 22 /* sheva */
	mcc11 = 15 /* hataf segol */
	mcc12 = 16 /* hataf patah */
	mcc13 = 17 /* hataf qamats */
	mcc14 = 23 /* hiriq */
	mcc15 = 18 /* tsere */
	mcc16 = 19 /* segol */
	mcc17 = 20 /* patah */
	mcc18 = 21 /* qamats */
	mcc19 = 14 /* holam */
	mcc20 = 24 /* qubuts */
	mcc21 = 12 /* dagesh */
	mcc22 = 25 /* meteg */
	mcc23 = 13 /* rafe */
	mcc24 = 10 /* shin dot */
	mcc25 = 11 /* sin dot */
	mcc26 = 26 /* point varika */

	/*
	 * Arabic
	 *
	 * Modify to move Shadda (ccc=33) before other marks.  See:
	 * https://unicode.org/faq/normalization.html#8
	 * https://unicode.org/faq/normalization.html#9
	 */
	mcc27 = 28 /* fathatan */
	mcc28 = 29 /* dammatan */
	mcc29 = 30 /* kasratan */
	mcc30 = 31 /* fatha */
	mcc31 = 32 /* damma */
	mcc32 = 33 /* kasra */
	mcc33 = 27 /* shadda */
	mcc34 = 34 /* sukun */
	mcc35 = 35 /* superscript alef */

	/* Syriac */
	mcc36 = 36 /* superscript alaph */

	/* Telugu
	 *
	 * Modify Telugu length marks (ccc=84, ccc=91).
	 * These are the only matras in the main Indic scripts range that have
	 * a non-zero ccc.  That makes them reorder with the Halant (ccc=9).
	 * Assign 4 and 5, which are otherwise unassigned.
	 */
	mcc84 = 4 /* length mark */
	mcc91 = 5 /* ai length mark */

	/* Thai
	 *
	 * Modify U+0E38 and U+0E39 (ccc=103) to be reordered before U+0E3A (ccc=9).
	 * Assign 3, which is unassigned otherwise.
	 * Uniscribe does this reordering too.
	 */
	mcc103 = 3   /* sara u / sara uu */
	mcc107 = 107 /* mai * */

	/* Lao */
	mcc118 = 118 /* sign u / sign uu */
	mcc122 = 122 /* mai * */

	/* Tibetan
	 *
	 * In case of multiple vowel-signs, use u first (but after achung)
	 * this allows Dzongkha multi-vowel shortcuts to render correctly
	 */
	mcc129 = 129 /* sign aa */
	mcc130 = 132 /* sign i */
	mcc132 = 131 /* sign u */
)

var _hb_modified_combining_class = [256]uint8{
	0, /* HB_UNICODE_COMBINING_CLASS_NOT_REORDERED */
	1, /* HB_UNICODE_COMBINING_CLASS_OVERLAY */
	2, 3, 4, 5, 6,
	7, /* HB_UNICODE_COMBINING_CLASS_NUKTA */
	8, /* HB_UNICODE_COMBINING_CLASS_KANA_VOICING */
	9, /* HB_UNICODE_COMBINING_CLASS_VIRAMA */

	/* Hebrew */
	mcc10,
	mcc11,
	mcc12,
	mcc13,
	mcc14,
	mcc15,
	mcc16,
	mcc17,
	mcc18,
	mcc19,
	mcc20,
	mcc21,
	mcc22,
	mcc23,
	mcc24,
	mcc25,
	mcc26,

	/* Arabic */
	mcc27,
	mcc28,
	mcc29,
	mcc30,
	mcc31,
	mcc32,
	mcc33,
	mcc34,
	mcc35,

	/* Syriac */
	mcc36,

	37, 38, 39,
	40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
	60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
	80, 81, 82, 83,

	/* Telugu */
	mcc84,
	85, 86, 87, 88, 89, 90,
	mcc91,
	92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102,

	/* Thai */
	mcc103,
	104, 105, 106,
	mcc107,
	108, 109, 110, 111, 112, 113, 114, 115, 116, 117,

	/* Lao */
	mcc118,
	119, 120, 121,
	mcc122,
	123, 124, 125, 126, 127, 128,

	/* Tibetan */
	mcc129,
	mcc130,
	131,
	mcc132,
	133, 134, 135, 136, 137, 138, 139,

	140, 141, 142, 143, 144, 145, 146, 147, 148, 149,
	150, 151, 152, 153, 154, 155, 156, 157, 158, 159,
	160, 161, 162, 163, 164, 165, 166, 167, 168, 169,
	170, 171, 172, 173, 174, 175, 176, 177, 178, 179,
	180, 181, 182, 183, 184, 185, 186, 187, 188, 189,
	190, 191, 192, 193, 194, 195, 196, 197, 198, 199,

	200, /* HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW_LEFT */
	201,
	202, /* HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW */
	203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213,
	214, /* HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE */
	215,
	216, /* HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE_RIGHT */
	217,
	218, /* HB_UNICODE_COMBINING_CLASS_BELOW_LEFT */
	219,
	220, /* HB_UNICODE_COMBINING_CLASS_BELOW */
	221,
	222, /* HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT */
	223,
	224, /* HB_UNICODE_COMBINING_CLASS_LEFT */
	225,
	226, /* HB_UNICODE_COMBINING_CLASS_RIGHT */
	227,
	228, /* HB_UNICODE_COMBINING_CLASS_ABOVE_LEFT */
	229,
	230, /* HB_UNICODE_COMBINING_CLASS_ABOVE */
	231,
	232, /* HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT */
	233, /* HB_UNICODE_COMBINING_CLASS_DOUBLE_BELOW */
	234, /* HB_UNICODE_COMBINING_CLASS_DOUBLE_ABOVE */
	235, 236, 237, 238, 239,
	240, /* HB_UNICODE_COMBINING_CLASS_IOTA_SUBSCRIPT */
	241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254,
	255, /* HB_UNICODE_COMBINING_CLASS_INVALID */
}

type hb_unicode_funcs_t struct {
	//   hb_object_header_t header;

	//   hb_unicode_funcs_t *parent;

	// #define HB_UNICODE_FUNC_IMPLEMENT(return_type, name) \
	//   return_type name (hb_codepoint_t unicode) { return func.name (this, unicode, user_data.name); }
	// HB_UNICODE_FUNCS_IMPLEMENT_CALLBACKS_SIMPLE
	// #undef HB_UNICODE_FUNC_IMPLEMENT

	//   hb_bool_t compose (hb_codepoint_t a, hb_codepoint_t b,
	// 		     hb_codepoint_t *ab)
	//   {
	//     *ab = 0;
	//     if (unlikely (!a || !b)) return false;
	//     return func.compose (this, a, b, ab, user_data.compose);
	//   }

	//   hb_bool_t decompose (hb_codepoint_t ab,
	// 		       hb_codepoint_t *a, hb_codepoint_t *b)
	//   {
	//     *a = ab; *b = 0;
	//     return func.decompose (this, ab, a, b, user_data.decompose);
	//   }

	//   unsigned int decompose_compatibility (hb_codepoint_t  u,
	// 					hb_codepoint_t *decomposed)
	//   {
	// #ifdef HB_DISABLE_DEPRECATED
	//     unsigned int ret  = 0;
	// #else
	//     unsigned int ret = func.decompose_compatibility (this, u, decomposed, user_data.decompose_compatibility);
	// #endif
	//     if (ret == 1 && u == decomposed[0]) {
	//       decomposed[0] = 0;
	//       return 0;
	//     }
	//     decomposed[ret] = 0;
	//     return ret;
	//   }

	//   static hb_bool_t
	//   is_variation_selector (hb_codepoint_t unicode)
	//   {
	//     /* U+180B..180D MONGOLIAN FREE VARIATION SELECTORs are handled in the
	//      * Arabic shaper.  No need to match them here. */
	//     return unlikely (hb_in_ranges<hb_codepoint_t> (unicode,
	// 						   0xFE00u, 0xFE0Fu, /* VARIATION SELECTOR-1..16 */
	// 						   0xE0100u, 0xE01EFu));  /* VARIATION SELECTOR-17..256 */
	//   }

	//   /* Space estimates based on:
	//    * https://unicode.org/charts/PDF/U2000.pdf
	//    * https://docs.microsoft.com/en-us/typography/develop/character-design-standards/whitespace
	//    */
	//   enum space_t {
	//     NOT_SPACE = 0,
	//     SPACE_EM   = 1,
	//     SPACE_EM_2 = 2,
	//     SPACE_EM_3 = 3,
	//     SPACE_EM_4 = 4,
	//     SPACE_EM_5 = 5,
	//     SPACE_EM_6 = 6,
	//     SPACE_EM_16 = 16,
	//     SPACE_4_EM_18,	/* 4/18th of an EM! */
	//     SPACE,
	//     SPACE_FIGURE,
	//     SPACE_PUNCTUATION,
	//     SPACE_NARROW,
	//   };
	//   static space_t
	//   space_fallback_type (hb_codepoint_t u)
	//   {
	//     switch (u)
	//     {
	//       /* All GC=Zs chars that can use a fallback. */
	//       default:	    return NOT_SPACE;	/* U+1680 OGHAM SPACE MARK */
	//       case 0x0020u: return SPACE;	/* U+0020 SPACE */
	//       case 0x00A0u: return SPACE;	/* U+00A0 NO-BREAK SPACE */
	//       case 0x2000u: return SPACE_EM_2;	/* U+2000 EN QUAD */
	//       case 0x2001u: return SPACE_EM;	/* U+2001 EM QUAD */
	//       case 0x2002u: return SPACE_EM_2;	/* U+2002 EN SPACE */
	//       case 0x2003u: return SPACE_EM;	/* U+2003 EM SPACE */
	//       case 0x2004u: return SPACE_EM_3;	/* U+2004 THREE-PER-EM SPACE */
	//       case 0x2005u: return SPACE_EM_4;	/* U+2005 FOUR-PER-EM SPACE */
	//       case 0x2006u: return SPACE_EM_6;	/* U+2006 SIX-PER-EM SPACE */
	//       case 0x2007u: return SPACE_FIGURE;	/* U+2007 FIGURE SPACE */
	//       case 0x2008u: return SPACE_PUNCTUATION;	/* U+2008 PUNCTUATION SPACE */
	//       case 0x2009u: return SPACE_EM_5;		/* U+2009 THIN SPACE */
	//       case 0x200Au: return SPACE_EM_16;		/* U+200A HAIR SPACE */
	//       case 0x202Fu: return SPACE_NARROW;	/* U+202F NARROW NO-BREAK SPACE */
	//       case 0x205Fu: return SPACE_4_EM_18;	/* U+205F MEDIUM MATHEMATICAL SPACE */
	//       case 0x3000u: return SPACE_EM;		/* U+3000 IDEOGRAPHIC SPACE */
	//     }
	//   }

	//   struct {
	// #define HB_UNICODE_FUNC_IMPLEMENT(name) hb_unicode_##name##_func_t name;
	//     HB_UNICODE_FUNCS_IMPLEMENT_CALLBACKS
	// #undef HB_UNICODE_FUNC_IMPLEMENT
	//   } func;

	//   struct {
	// #define HB_UNICODE_FUNC_IMPLEMENT(name) void *name;
	//     HB_UNICODE_FUNCS_IMPLEMENT_CALLBACKS
	// #undef HB_UNICODE_FUNC_IMPLEMENT
	//   } user_data;

	//   struct {
	// #define HB_UNICODE_FUNC_IMPLEMENT(name) hb_destroy_func_t name;
	//     HB_UNICODE_FUNCS_IMPLEMENT_CALLBACKS
	// #undef HB_UNICODE_FUNC_IMPLEMENT
	//   } destroy;
}

func (hb_unicode_funcs_t) modified_combining_class(u rune) uint8 {
	/* This hack belongs to the USE shaper (for Tai Tham):
	 * Reorder SAKOT to ensure it comes after any tone marks. */
	if u == 0x1A60 {
		return 254
	}

	/* This hack belongs to the Tibetan shaper:
	 * Reorder PADMA to ensure it comes after any vowel marks. */
	if u == 0x0FC6 {
		return 254
	}
	/* Reorder TSA -PHRU to reorder before U+0F74 */
	if u == 0x0F39 {
		return 127
	}

	return _hb_modified_combining_class[unicodedata.LookupCombiningClass(u)]
}

// Default_Ignorable codepoints:
//
// Note: While U+115F, U+1160, U+3164 and U+FFA0 are Default_Ignorable,
// we do NOT want to hide them, as the way Uniscribe has implemented them
// is with regular spacing glyphs, and that's the way fonts are made to work.
// As such, we make exceptions for those four.
// Also ignoring U+1BCA0..1BCA3. https://github.com/harfbuzz/harfbuzz/issues/503
func (hb_unicode_funcs_t) is_default_ignorable(ch rune) bool {
	is := unicode.Is(unicode.Other_Default_Ignorable_Code_Point, ch)
	if !is {
		return false
	}
	// special cases
	if ch == '\u115F' || ch == '\u1160' || ch == '\u3164' || ch == '\uFFA0' ||
		('\U0001BCA0' <= ch && ch <= '\U0001BCA3') {
		return false
	}
	return true
}

// retrieves the General Category property for
// a specified Unicode code point, expressed as enumeration value.
func (hb_unicode_funcs_t) general_category(ch rune) generalCategory {
	for i := 1; i < len(generalCategories); i++ {
		if unicode.Is(generalCategories[i], ch) {
			return generalCategory(i)
		}
	}
	return unassigned
}

func (hb_unicode_funcs_t) isExtendedPictographic(ch rune) bool {
	return unicode.Is(unicodedata.Extended_Pictographic, ch)
}

// returns the Mirroring Glyph code point (for bi-directional
// replacement) of a code point, or itself
func (hb_unicode_funcs_t) mirroring(ch rune) rune {
	out, _ := unicodedata.LookupMirrorChar(ch)
	return out
}
