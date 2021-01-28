package harfbuzz

import "unicode"

var uni = hb_unicode_funcs_t{}

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

	//   unsigned int
	//   modified_combining_class (hb_codepoint_t u)
	//   {
	//     /* XXX This hack belongs to the USE shaper (for Tai Tham):
	//      * Reorder SAKOT to ensure it comes after any tone marks. */
	//     if (unlikely (u == 0x1A60u)) return 254;

	//     /* XXX This hack belongs to the Tibetan shaper:
	//      * Reorder PADMA to ensure it comes after any vowel marks. */
	//     if (unlikely (u == 0x0FC6u)) return 254;
	//     /* Reorder TSA -PHRU to reorder before U+0F74 */
	//     if (unlikely (u == 0x0F39u)) return 127;

	//     return _hb_modified_combining_class[combining_class (u)];
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
