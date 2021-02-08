package harfbuzz

import (
	"unicode"

	"github.com/benoitkugler/textlayout/unicodedata"
)

// Uni exposes some lookup functions for Unicode properties.
var Uni = hb_unicode_funcs_t{}

// GeneralCategory is an enum value to allow compact storage (see generalCategories)
type GeneralCategory uint8

const (
	Unassigned GeneralCategory = iota
	Control
	Format
	PrivateUse
	Surrogate
	LowercaseLetter
	ModifierLetter
	OtherLetter
	TitlecaseLetter
	UppercaseLetter
	SpacingMark
	EnclosingMark
	NonSpacingMark
	DecimalNumber
	LetterNumber
	OtherNumber
	ConnectPunctuation
	DashPunctuation
	ClosePunctuation
	FinalPunctuation
	InitialPunctuation
	OtherPunctuation
	OpenPunctuation
	CurrencySymbol
	ModifierSymbol
	MathSymbol
	OtherSymbol
	LineSeparator
	ParagraphSeparator
	SpaceSeparator
)

// correspondance with *unicode.RangeTable classes
var generalCategories = [...]*unicode.RangeTable{
	Unassigned:         nil,
	Control:            unicode.Cc,
	Format:             unicode.Cf,
	PrivateUse:         unicode.Co,
	Surrogate:          unicode.Cs,
	LowercaseLetter:    unicode.Ll,
	ModifierLetter:     unicode.Lm,
	OtherLetter:        unicode.Lo,
	TitlecaseLetter:    unicode.Lt,
	UppercaseLetter:    unicode.Lu,
	SpacingMark:        unicode.Mc,
	EnclosingMark:      unicode.Me,
	NonSpacingMark:     unicode.Mn,
	DecimalNumber:      unicode.Nd,
	LetterNumber:       unicode.Nl,
	OtherNumber:        unicode.No,
	ConnectPunctuation: unicode.Pc,
	DashPunctuation:    unicode.Pd,
	ClosePunctuation:   unicode.Pe,
	FinalPunctuation:   unicode.Pf,
	InitialPunctuation: unicode.Pi,
	OtherPunctuation:   unicode.Po,
	OpenPunctuation:    unicode.Ps,
	CurrencySymbol:     unicode.Sc,
	ModifierSymbol:     unicode.Sk,
	MathSymbol:         unicode.Sm,
	OtherSymbol:        unicode.So,
	LineSeparator:      unicode.Zl,
	ParagraphSeparator: unicode.Zp,
	SpaceSeparator:     unicode.Zs,
}

func (g GeneralCategory) IsMark() bool {
	return g == SpacingMark || g == EnclosingMark || g == NonSpacingMark
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
	Mcc10 uint8 = 22 /* sheva */
	Mcc11 uint8 = 15 /* hataf segol */
	Mcc12 uint8 = 16 /* hataf patah */
	Mcc13 uint8 = 17 /* hataf qamats */
	Mcc14 uint8 = 23 /* hiriq */
	Mcc15 uint8 = 18 /* tsere */
	Mcc16 uint8 = 19 /* segol */
	Mcc17 uint8 = 20 /* patah */
	Mcc18 uint8 = 21 /* qamats */
	Mcc19 uint8 = 14 /* holam */
	Mcc20 uint8 = 24 /* qubuts */
	Mcc21 uint8 = 12 /* dagesh */
	Mcc22 uint8 = 25 /* meteg */
	Mcc23 uint8 = 13 /* rafe */
	Mcc24 uint8 = 10 /* shin dot */
	Mcc25 uint8 = 11 /* sin dot */
	Mcc26 uint8 = 26 /* point varika */

	/*
	 * Arabic
	 *
	 * Modify to move Shadda (ccc=33) before other marks.  See:
	 * https://unicode.org/faq/normalization.html#8
	 * https://unicode.org/faq/normalization.html#9
	 */
	Mcc27 uint8 = 28 /* fathatan */
	Mcc28 uint8 = 29 /* dammatan */
	Mcc29 uint8 = 30 /* kasratan */
	Mcc30 uint8 = 31 /* fatha */
	Mcc31 uint8 = 32 /* damma */
	Mcc32 uint8 = 33 /* kasra */
	Mcc33 uint8 = 27 /* shadda */
	Mcc34 uint8 = 34 /* sukun */
	Mcc35 uint8 = 35 /* superscript alef */

	/* Syriac */
	Mcc36 uint8 = 36 /* superscript alaph */

	/* Telugu
	 *
	 * Modify Telugu length marks (ccc=84, ccc=91).
	 * These are the only matras in the main Indic scripts range that have
	 * a non-zero ccc.  That makes them reorder with the Halant (ccc=9).
	 * Assign 4 and 5, which are otherwise unassigned.
	 */
	Mcc84 uint8 = 4 /* length mark */
	Mcc91 uint8 = 5 /* ai length mark */

	/* Thai
	 *
	 * Modify U+0E38 and U+0E39 (ccc=103) to be reordered before U+0E3A (ccc=9).
	 * Assign 3, which is unassigned otherwise.
	 * Uniscribe does this reordering too.
	 */
	Mcc103 uint8 = 3   /* sara u / sara uu */
	Mcc107 uint8 = 107 /* mai * */

	/* Lao */
	Mcc118 uint8 = 118 /* sign u / sign uu */
	Mcc122 uint8 = 122 /* mai * */

	/* Tibetan
	 *
	 * In case of multiple vowel-signs, use u first (but after achung)
	 * this allows Dzongkha multi-vowel shortcuts to render correctly
	 */
	Mcc129 = 129 /* sign aa */
	Mcc130 = 132 /* sign i */
	Mcc132 = 131 /* sign u */
)

var _hb_modified_combining_class = [256]uint8{
	0, /* HB_UNICODE_COMBINING_CLASS_NOT_REORDERED */
	1, /* HB_UNICODE_COMBINING_CLASS_OVERLAY */
	2, 3, 4, 5, 6,
	7, /* HB_UNICODE_COMBINING_CLASS_NUKTA */
	8, /* HB_UNICODE_COMBINING_CLASS_KANA_VOICING */
	9, /* HB_UNICODE_COMBINING_CLASS_VIRAMA */

	/* Hebrew */
	Mcc10,
	Mcc11,
	Mcc12,
	Mcc13,
	Mcc14,
	Mcc15,
	Mcc16,
	Mcc17,
	Mcc18,
	Mcc19,
	Mcc20,
	Mcc21,
	Mcc22,
	Mcc23,
	Mcc24,
	Mcc25,
	Mcc26,

	/* Arabic */
	Mcc27,
	Mcc28,
	Mcc29,
	Mcc30,
	Mcc31,
	Mcc32,
	Mcc33,
	Mcc34,
	Mcc35,

	/* Syriac */
	Mcc36,

	37, 38, 39,
	40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
	60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
	80, 81, 82, 83,

	/* Telugu */
	Mcc84,
	85, 86, 87, 88, 89, 90,
	Mcc91,
	92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102,

	/* Thai */
	Mcc103,
	104, 105, 106,
	Mcc107,
	108, 109, 110, 111, 112, 113, 114, 115, 116, 117,

	/* Lao */
	Mcc118,
	119, 120, 121,
	Mcc122,
	123, 124, 125, 126, 127, 128,

	/* Tibetan */
	Mcc129,
	Mcc130,
	131,
	Mcc132,
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

type hb_unicode_funcs_t struct { //   hb_object_header_t header;
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
func (hb_unicode_funcs_t) IsDefaultIgnorable(ch rune) bool {
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
func (hb_unicode_funcs_t) GeneralCategory(ch rune) GeneralCategory {
	for i := 1; i < len(generalCategories); i++ {
		if unicode.Is(generalCategories[i], ch) {
			return GeneralCategory(i)
		}
	}
	return Unassigned
}

func (hb_unicode_funcs_t) isExtendedPictographic(ch rune) bool {
	return unicode.Is(unicodedata.Extended_Pictographic, ch)
}

// returns the Mirroring Glyph code point (for bi-directional
// replacement) of a code point, or itself
func (hb_unicode_funcs_t) Mirroring(ch rune) rune {
	out, _ := unicodedata.LookupMirrorChar(ch)
	return out
}

/* Space estimates based on:
 * https://unicode.org/charts/PDF/U2000.pdf
 * https://docs.microsoft.com/en-us/typography/develop/character-design-standards/whitespace
 */
const (
	SPACE_EM_16   = 16 + iota
	SPACE_4_EM_18 // 4/18th of an EM!
	SPACE
	SPACE_FIGURE
	SPACE_PUNCTUATION
	SPACE_NARROW
	NOT_SPACE  = 0
	SPACE_EM   = 1
	SPACE_EM_2 = 2
	SPACE_EM_3 = 3
	SPACE_EM_4 = 4
	SPACE_EM_5 = 5
	SPACE_EM_6 = 6
)

func (hb_unicode_funcs_t) SpaceFallbackType(u rune) uint8 {
	switch u {
	// all GC=Zs chars that can use a fallback.
	case 0x0020:
		return SPACE /* U+0020 SPACE */
	case 0x00A0:
		return SPACE /* U+00A0 NO-BREAK SPACE */
	case 0x2000:
		return SPACE_EM_2 /* U+2000 EN QUAD */
	case 0x2001:
		return SPACE_EM /* U+2001 EM QUAD */
	case 0x2002:
		return SPACE_EM_2 /* U+2002 EN SPACE */
	case 0x2003:
		return SPACE_EM /* U+2003 EM SPACE */
	case 0x2004:
		return SPACE_EM_3 /* U+2004 THREE-PER-EM SPACE */
	case 0x2005:
		return SPACE_EM_4 /* U+2005 FOUR-PER-EM SPACE */
	case 0x2006:
		return SPACE_EM_6 /* U+2006 SIX-PER-EM SPACE */
	case 0x2007:
		return SPACE_FIGURE /* U+2007 FIGURE SPACE */
	case 0x2008:
		return SPACE_PUNCTUATION /* U+2008 PUNCTUATION SPACE */
	case 0x2009:
		return SPACE_EM_5 /* U+2009 THIN SPACE */
	case 0x200A:
		return SPACE_EM_16 /* U+200A HAIR SPACE */
	case 0x202F:
		return SPACE_NARROW /* U+202F NARROW NO-BREAK SPACE */
	case 0x205F:
		return SPACE_4_EM_18 /* U+205F MEDIUM MATHEMATICAL SPACE */
	case 0x3000:
		return SPACE_EM /* U+3000 IDEOGRAPHIC SPACE */
	default:
		return NOT_SPACE /* U+1680 OGHAM SPACE MARK */
	}
}

func (hb_unicode_funcs_t) IsVariationSelector(r rune) bool {
	/* U+180B..180D MONGOLIAN FREE VARIATION SELECTORs are handled in the
	 * Arabic shaper.  No need to match them here. */
	/* VARIATION SELECTOR-1..16 */
	/* VARIATION SELECTOR-17..256 */
	return (0xFE00 <= r && r <= 0xFE0F) || (0xE0100 <= r && r <= 0xE01EF)
}

func (hb_unicode_funcs_t) Decompose(ab rune) (a, b rune, ok bool) { return unicodedata.Decompose(ab) }
func (hb_unicode_funcs_t) Compose(a, b rune) (rune, bool)         { return unicodedata.Compose(a, b) }

/* Prepare */

/* Implement enough of Unicode Graphemes here that shaping
 * in reverse-direction wouldn't break graphemes.  Namely,
 * we mark all marks and ZWJ and ZWJ,Extended_Pictographic
 * sequences as continuations.  The foreach_grapheme()
 * macro uses this bit.
 *
 * https://www.unicode.org/reports/tr29/#Regex_Definitions
 */
func (buffer *Buffer) SetUnicodeProps() {
	info := buffer.info
	for i := 0; i < len(info); i++ {
		info[i].SetUnicodeProps(buffer)

		/* Marks are already set as continuation by the above line.
		 * Handle Emoji_Modifier and ZWJ-continuation. */
		if info[i].unicode.GeneralCategory() == ModifierSymbol && (0x1F3FB <= info[i].codepoint && info[i].codepoint <= 0x1F3FF) {
			info[i].setContinuation()
		} else if info[i].isZwj() {
			info[i].setContinuation()
			if i+1 < len(buffer.info) && Uni.isExtendedPictographic(info[i+1].codepoint) {
				i++
				info[i].SetUnicodeProps(buffer)
				info[i].setContinuation()
			}
		} else if 0xE0020 <= info[i].codepoint && info[i].codepoint <= 0xE007F {
			/* Or part of the Other_Grapheme_Extend that is not marks.
			 * As of Unicode 11 that is just:
			 *
			 * 200C          ; Other_Grapheme_Extend # Cf       ZERO WIDTH NON-JOINER
			 * FF9E..FF9F    ; Other_Grapheme_Extend # Lm   [2] HALFWIDTH KATAKANA VOICED SOUND MARK..HALFWIDTH KATAKANA SEMI-VOICED SOUND MARK
			 * E0020..E007F  ; Other_Grapheme_Extend # Cf  [96] TAG SPACE..CANCEL TAG
			 *
			 * ZWNJ is special, we don't want to merge it as there's no need, and keeping
			 * it separate results in more granular clusters.  Ignore Katakana for now.
			 * Tags are used for Emoji sub-region flag sequences:
			 * https://github.com/harfbuzz/harfbuzz/issues/1556
			 */
			info[i].setContinuation()
		}
	}
}

func (buffer *Buffer) InsertDottedCircle(font *Font) {
	if buffer.Flags&DoNotInsertDottedCircle != 0 {
		return
	}

	if buffer.Flags&Bot == 0 || len(buffer.context[0]) != 0 ||
		!buffer.info[0].IsUnicodeMark() {
		return
	}

	if !font.HasGlyph(0x25CC) {
		return
	}

	dottedcircle := GlyphInfo{codepoint: 0x25CC}
	dottedcircle.SetUnicodeProps(buffer)

	buffer.ClearOutput()

	buffer.idx = 0
	dottedcircle.Cluster = buffer.Cur(0).Cluster
	dottedcircle.mask = buffer.Cur(0).mask
	buffer.outInfo = append(buffer.outInfo, dottedcircle)
	for buffer.idx < len(buffer.info) {
		buffer.NextGlyph()
	}
	buffer.SwapBuffers()
}

func (buffer *Buffer) FormClusters() {
	if buffer.scratchFlags&HB_BUFFER_SCRATCH_FLAG_HAS_NON_ASCII == 0 {
		return
	}

	iter, count := buffer.GraphemesIterator()
	if buffer.ClusterLevel == MonotoneGraphemes {
		for start, end := iter.Next(); start < count; start, end = iter.Next() {
			buffer.MergeClusters(start, end)
		}
	} else {
		for start, end := iter.Next(); start < count; start, end = iter.Next() {
			buffer.UnsafeToBreak(start, end)
		}
	}
}

func (buffer *Buffer) EnsureNativeDirection() {
	direction := buffer.Props.Direction
	horiz_dir := GetHorizontalDirection(buffer.Props.Script)

	/* TODO vertical:
	* The only BTT vertical script is Ogham, but it's not clear to me whether OpenType
	* Ogham fonts are supposed to be implemented BTT or not.  Need to research that
	* first. */
	if (direction.IsHorizontal() && direction != horiz_dir && horiz_dir != Invalid) ||
		(direction.IsVertical() && direction != TopToBottom) {

		iter, count := buffer.GraphemesIterator()
		if buffer.ClusterLevel == MonotoneCharacters {
			for start, end := iter.Next(); start < count; start, end = iter.Next() {
				buffer.MergeClusters(start, end)
				buffer.reverse_range(start, end)
			}
		} else {
			for start, end := iter.Next(); start < count; start, end = iter.Next() {
				// form_clusters() merged clusters already, we don't merge.
				buffer.reverse_range(start, end)
			}
		}
		buffer.Reverse()

		buffer.Props.Direction = buffer.Props.Direction.reverse()
	}
}
