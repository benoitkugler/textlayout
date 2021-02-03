package opentype

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz/common"
	"github.com/benoitkugler/textlayout/language"
)

// ported from harfbuzz/src/hb-ot-shape-complex-arabic.cc, hb-ot-shape-complex-arabic-fallback.hh Copyright © 2010,2012  Google, Inc. Behdad Esfahbod

//  /* buffer var allocations */
//  #define arabic_shaping_action() complex_var_u8_auxiliary() /* arabic shaping action */

const HB_BUFFER_SCRATCH_FLAG_ARABIC_HAS_STCH = common.HB_BUFFER_SCRATCH_FLAG_COMPLEX0

//  /* See:
//   * https://github.com/harfbuzz/harfbuzz/commit/6e6f82b6f3dde0fc6c3c7d991d9ec6cfff57823d#commitcomment-14248516 */
//  #define HB_ARABIC_GENERAL_CATEGORY_IS_WORD(gen_cat) \
// 	 (FLAG_UNSAFE (gen_cat) & \
// 	  (FLAG (HB_UNICODE_GENERAL_CATEGORY_UNASSIGNED) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_PRIVATE_USE) | \
// 	   /*FLAG (HB_UNICODE_GENERAL_CATEGORY_LOWERCASE_LETTER) |*/ \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_MODIFIER_LETTER) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_OTHER_LETTER) | \
// 	   /*FLAG (HB_UNICODE_GENERAL_CATEGORY_TITLECASE_LETTER) |*/ \
// 	   /*FLAG (HB_UNICODE_GENERAL_CATEGORY_UPPERCASE_LETTER) |*/ \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_SPACING_MARK) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_ENCLOSING_MARK) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_NON_SPACING_MARK) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_DECIMAL_NUMBER) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_LETTER_NUMBER) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_OTHER_NUMBER) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_CURRENCY_SYMBOL) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_MODIFIER_SYMBOL) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_MATH_SYMBOL) | \
// 	   FLAG (HB_UNICODE_GENERAL_CATEGORY_OTHER_SYMBOL)))

//  /*
//   * Joining types:
//   */

//  /*
//   * Bits used in the joining tables
//   */
//  enum hb_arabic_joining_type_t {
//    JOINING_TYPE_U		= 0,
//    JOINING_TYPE_L		= 1,
//    JOINING_TYPE_R		= 2,
//    JOINING_TYPE_D		= 3,
//    JOINING_TYPE_C		= JOINING_TYPE_D,
//    JOINING_GROUP_ALAPH		= 4,
//    JOINING_GROUP_DALATH_RISH	= 5,
//    NUM_STATE_MACHINE_COLS	= 6,

//    JOINING_TYPE_T = 7,
//    JOINING_TYPE_X = 8  /* means: use general-category to choose between U or T. */
//  };

//  #include "hb-ot-shape-complex-arabic-table.hh"

//  static unsigned int get_joining_type (hb_codepoint_t u, hb_unicode_general_category_t gen_cat)
//  {
//    unsigned int j_type = joining_type(u);
//    if (likely (j_type != JOINING_TYPE_X))
// 	 return j_type;

//    return (FLAG_UNSAFE(gen_cat) &
// 	   (FLAG(HB_UNICODE_GENERAL_CATEGORY_NON_SPACING_MARK) |
// 		FLAG(HB_UNICODE_GENERAL_CATEGORY_ENCLOSING_MARK) |
// 		FLAG(HB_UNICODE_GENERAL_CATEGORY_FORMAT))
// 	  ) ?  JOINING_TYPE_T : JOINING_TYPE_U;
//  }

func featureIsSyriac(tag hb_tag_t) bool {
	return '2' <= byte(tag) && byte(tag) <= '3'
}

var arabic_features = [...]hb_tag_t{
	newTag('i', 's', 'o', 'l'),
	newTag('f', 'i', 'n', 'a'),
	newTag('f', 'i', 'n', '2'),
	newTag('f', 'i', 'n', '3'),
	newTag('m', 'e', 'd', 'i'),
	newTag('m', 'e', 'd', '2'),
	newTag('i', 'n', 'i', 't'),
	0,
}

/* Same order as the feature array */
const (
	ISOL = iota
	FINA
	FIN2
	FIN3
	MEDI
	MED2
	INIT

	NONE

	/* We abuse the same byte for other things... */
	STCH_FIXED
	STCH_REPEATING
)

//  static const struct arabic_state_table_entry {
// 	 uint8_t prev_action;
// 	 uint8_t curr_action;
// 	 uint16_t next_state;
//  } arabic_state_table[][NUM_STATE_MACHINE_COLS] =
//  {
//    /*   jt_U,          jt_L,          jt_R,          jt_D,          jg_ALAPH,      jg_DALATH_RISH */

//    /* State 0: prev was U, not willing to join. */
//    { {NONE,NONE,0}, {NONE,ISOL,2}, {NONE,ISOL,1}, {NONE,ISOL,2}, {NONE,ISOL,1}, {NONE,ISOL,6}, },

//    /* State 1: prev was R or ISOL/ALAPH, not willing to join. */
//    { {NONE,NONE,0}, {NONE,ISOL,2}, {NONE,ISOL,1}, {NONE,ISOL,2}, {NONE,FIN2,5}, {NONE,ISOL,6}, },

//    /* State 2: prev was D/L in ISOL form, willing to join. */
//    { {NONE,NONE,0}, {NONE,ISOL,2}, {INIT,FINA,1}, {INIT,FINA,3}, {INIT,FINA,4}, {INIT,FINA,6}, },

//    /* State 3: prev was D in FINA form, willing to join. */
//    { {NONE,NONE,0}, {NONE,ISOL,2}, {MEDI,FINA,1}, {MEDI,FINA,3}, {MEDI,FINA,4}, {MEDI,FINA,6}, },

//    /* State 4: prev was FINA ALAPH, not willing to join. */
//    { {NONE,NONE,0}, {NONE,ISOL,2}, {MED2,ISOL,1}, {MED2,ISOL,2}, {MED2,FIN2,5}, {MED2,ISOL,6}, },

//    /* State 5: prev was FIN2/FIN3 ALAPH, not willing to join. */
//    { {NONE,NONE,0}, {NONE,ISOL,2}, {ISOL,ISOL,1}, {ISOL,ISOL,2}, {ISOL,FIN2,5}, {ISOL,ISOL,6}, },

//    /* State 6: prev was DALATH/RISH, not willing to join. */
//    { {NONE,NONE,0}, {NONE,ISOL,2}, {NONE,ISOL,1}, {NONE,ISOL,2}, {NONE,FIN3,5}, {NONE,ISOL,6}, }
//  };

type complexShaperArabic struct{}

func (cs *complexShaperArabic) collect_features_arabic(plan *hb_ot_shape_planner_t) {
	map_ := &plan.map_

	/* We apply features according to the Arabic spec, with pauses
	* in between most.
	*
	* The pause between init/medi/... and rlig is required.  See eg:
	* https://bugzilla.mozilla.org/show_bug.cgi?id=644184
	*
	* The pauses between init/medi/... themselves are not necessarily
	* needed as only one of those features is applied to any character.
	* The only difference it makes is when fonts have contextual
	* substitutions.  We now follow the order of the spec, which makes
	* for better experience if that's what Uniscribe is doing.
	*
	* At least for Arabic, looks like Uniscribe has a pause between
	* rlig and calt.  Otherwise the IranNastaliq's ALLAH ligature won't
	* work.  However, testing shows that rlig and calt are applied
	* together for Mongolian in Uniscribe.  As such, we only add a
	* pause for Arabic, not other scripts.
	*
	* A pause after calt is required to make KFGQPC Uthmanic Script HAFS
	* work correctly.  See https://github.com/harfbuzz/harfbuzz/issues/505
	 */

	map_.enable_feature(newTag('s', 't', 'c', 'h'))
	map_.add_gsub_pause(recordStch)

	map_.enable_feature(newTag('c', 'c', 'm', 'p'))
	map_.enable_feature(newTag('l', 'o', 'c', 'l'))

	map_.add_gsub_pause(nil)

	for _, arabFeat := range arabic_features {
		has_fallback := plan.props.script == language.Arabic && !featureIsSyriac(arabFeat)
		fl := F_NONE
		if has_fallback {
			fl = F_HAS_FALLBACK
		}
		map_.add_feature_ext(arabFeat, fl, 1)
		map_.add_gsub_pause(nil)
	}

	/* Unicode says a ZWNJ means "don't ligate". In Arabic script
	* however, it says a ZWJ should also mean "don't ligate". So we run
	* the main ligating features as MANUAL_ZWJ. */

	map_.enable_feature_ext(newTag('r', 'l', 'i', 'g'), F_MANUAL_ZWJ|F_HAS_FALLBACK, 1)

	if plan.props.script == language.Arabic {
		map_.add_gsub_pause(arabicFallbackShape)
	}
	/* No pause after rclt.  See 98460779bae19e4d64d29461ff154b3527bf8420. */
	map_.enable_feature_ext(newTag('r', 'c', 'l', 't'), F_MANUAL_ZWJ, 1)
	map_.enable_feature_ext(newTag('c', 'a', 'l', 't'), F_MANUAL_ZWJ, 1)
	map_.add_gsub_pause(nil)

	/* The spec includes 'cswh'.  Earlier versions of Windows
	* used to enable this by default, but testing suggests
	* that Windows 8 and later do not enable it by default,
	* and spec now says 'Off by default'.
	* We disabled this in ae23c24c32.
	* Note that IranNastaliq uses this feature extensively
	* to fixup broken glyph sequences.  Oh well...
	* Test case: U+0643,U+0640,U+0631. */
	//map_.enable_feature (newTag('c','s','w','h'));
	map_.enable_feature(newTag('m', 's', 'e', 't'))
}

type arabic_shape_plan_t struct {
	/* The "+ 1" in the next array is to accommodate for the "NONE" command,
	* which is not an OpenType feature, but this simplifies the code by not
	* having to do a "if (... < NONE) ..." and just rely on the fact that
	* mask_array[NONE] == 0. */
	mask_array [len(arabic_features) + 1]common.Mask

	fallback_plan *arabic_fallback_plan_t

	do_fallback bool
	has_stch    bool
}

//  void *
//  data_create_arabic (plan *hb_ot_shape_plan_t)
//  {
//    arabic_shape_plan_t *arabic_plan = (arabic_shape_plan_t *) calloc (1, sizeof (arabic_shape_plan_t));
//    if (unlikely (!arabic_plan))
// 	 return nil;

//    arabic_plan.do_fallback = plan.props.script == HB_SCRIPT_ARABIC;
//    arabic_plan.has_stch = !!plan.map_.get_1_mask (newTag ('s','t','c','h'));
//    for (unsigned int i = 0; i < ARABIC_NUM_FEATURES; i++) {
// 	 arabic_plan.mask_array[i] = plan.map_.get_1_mask (arabic_features[i]);
// 	 arabic_plan.do_fallback = arabic_plan.do_fallback &&
// 					(featureIsSyriac (arabic_features[i]) ||
// 				 plan.map_.needs_fallback (arabic_features[i]));
//    }

//    return arabic_plan;
//  }

//  void
//  data_destroy_arabic (void *data)
//  {
//    arabic_shape_plan_t *arabic_plan = (arabic_shape_plan_t *) data;

//    arabic_fallback_plan_destroy (arabic_plan.fallback_plan);

//    free (data);
//  }

//  static void
//  arabic_joining (buffer *common.Buffer)
//  {
//    unsigned int count = buffer.len;
//    hb_glyph_info_t *info = buffer.Info;
//    unsigned int prev = UINT_MAX, state = 0;

//    /* Check pre-context */
//    for (unsigned int i = 0; i < buffer.context_len[0]; i++)
//    {
// 	 unsigned int this_type = get_joining_type (buffer.context[0][i], buffer.unicode.general_category (buffer.context[0][i]));

// 	 if (unlikely (this_type == JOINING_TYPE_T))
// 	   continue;

// 	 const arabic_state_table_entry *entry = &arabic_state_table[state][this_type];
// 	 state = entry.next_state;
// 	 break;
//    }

//    for (unsigned int i = 0; i < count; i++)
//    {
// 	 unsigned int this_type = get_joining_type (info[i].codepoint, _hb_glyph_info_get_general_category (&info[i]));

// 	 if (unlikely (this_type == JOINING_TYPE_T)) {
// 	   info[i].arabic_shaping_action() = NONE;
// 	   continue;
// 	 }

// 	 const arabic_state_table_entry *entry = &arabic_state_table[state][this_type];

// 	 if (entry.prev_action != NONE && prev != UINT_MAX)
// 	 {
// 	   info[prev].arabic_shaping_action() = entry.prev_action;
// 	   buffer.unsafe_to_break (prev, i + 1);
// 	 }

// 	 info[i].arabic_shaping_action() = entry.curr_action;

// 	 prev = i;
// 	 state = entry.next_state;
//    }

//    for (unsigned int i = 0; i < buffer.context_len[1]; i++)
//    {
// 	 unsigned int this_type = get_joining_type (buffer.context[1][i], buffer.unicode.general_category (buffer.context[1][i]));

// 	 if (unlikely (this_type == JOINING_TYPE_T))
// 	   continue;

// 	 const arabic_state_table_entry *entry = &arabic_state_table[state][this_type];
// 	 if (entry.prev_action != NONE && prev != UINT_MAX)
// 	   info[prev].arabic_shaping_action() = entry.prev_action;
// 	 break;
//    }
//  }

//  static void
//  mongolian_variation_selectors (buffer *common.Buffer)
//  {
//    /* Copy arabic_shaping_action() from base to Mongolian variation selectors. */
//    unsigned int count = buffer.len;
//    hb_glyph_info_t *info = buffer.Info;
//    for (unsigned int i = 1; i < count; i++)
// 	 if (unlikely (hb_in_range<hb_codepoint_t> (info[i].codepoint, 0x180Bu, 0x180Du)))
// 	   info[i].arabic_shaping_action() = info[i - 1].arabic_shaping_action();
//  }

//  void
//  setup_masks_arabic_plan (const arabic_shape_plan_t *arabic_plan,
// 			  common.Buffer               *buffer,
// 			  hb_script_t                script)
//  {
//    HB_BUFFER_ALLOCATE_VAR (buffer, arabic_shaping_action);

//    arabic_joining (buffer);
//    if (script == HB_SCRIPT_MONGOLIAN)
// 	 mongolian_variation_selectors (buffer);

//    unsigned int count = buffer.len;
//    hb_glyph_info_t *info = buffer.Info;
//    for (unsigned int i = 0; i < count; i++)
// 	 info[i].mask |= arabic_plan.mask_array[info[i].arabic_shaping_action()];
//  }

//  static void
//  setup_masks_arabic (plan *hb_ot_shape_plan_t,
// 			 common.Buffer              *buffer,
// 			 common.Font                *font HB_UNUSED)
//  {
//    const arabic_shape_plan_t *arabic_plan = (const arabic_shape_plan_t *) plan.data;
//    setup_masks_arabic_plan (arabic_plan, buffer, plan.props.script);
//  }

func arabicFallbackShape(plan *hb_ot_shape_plan_t, font *common.Font, buffer *common.Buffer) {
	arabic_plan := plan.data.(*arabic_shape_plan_t)

	if !arabic_plan.do_fallback {
		return
	}

	fallback_plan := arabic_plan.fallback_plan
	if fallback_plan == nil {
		/* This sucks.  We need a font to build the fallback plan... */
		fallback_plan = newArabicFallbackPlan(plan, font)
	}

	fallback_plan.shape(font, buffer)
}

//  /*
//   * Stretch feature: "stch".
//   * See example here:
//   * https://docs.microsoft.com/en-us/typography/script-development/syriac
//   * We implement this in a generic way, such that the Arabic subtending
//   * marks can use it as well.
//   */

func recordStch(plan *hb_ot_shape_plan_t, _ *common.Font, buffer *common.Buffer) {
	arabic_plan := plan.data.(*arabic_shape_plan_t)
	if !arabic_plan.has_stch {
		return
	}

	/* 'stch' feature was just applied.  Look for anything that multiplied,
	* and record it for stch treatment later.  Note that rtlm, frac, etc
	* are applied before stch, but we assume that they didn't result in
	* anything multiplying into 5 pieces, so it's safe-ish... */

	//    unsigned int count = buffer.len;
	info := buffer.Info
	for i := range info {
		if info[i].Multiplied() {
			comp := info[i].GetLigComp()
			if comp%2 != 0 {
				info[i].Aux = STCH_REPEATING
			} else {
				info[i].Aux = STCH_FIXED
			}
			buffer.Flags |= HB_BUFFER_SCRATCH_FLAG_ARABIC_HAS_STCH
		}
	}
}

//  static void
//  apply_stch (plan *hb_ot_shape_plan_t HB_UNUSED,
// 		 common.Buffer              *buffer,
// 		 common.Font                *font)
//  {
//    if (likely (!(buffer.scratch_flags & HB_BUFFER_SCRATCH_FLAG_ARABIC_HAS_STCH)))
// 	 return;

//    /* The Arabic shaper currently always processes in RTL mode, so we should
// 	* stretch / position the stretched pieces to the left / preceding glyphs. */

//    /* We do a two pass implementation:
// 	* First pass calculates the exact number of extra glyphs we need,
// 	* We then enlarge buffer to have that much room,
// 	* Second pass applies the stretch, copying things to the end of buffer.
// 	*/

//    int sign = font.x_scale < 0 ? -1 : +1;
//    unsigned int extra_glyphs_needed = 0; // Set during MEASURE, used during CUT
//    enum { MEASURE, CUT } /* step_t */;

//    for (unsigned int step = MEASURE; step <= CUT; step = step + 1)
//    {
// 	 unsigned int count = buffer.len;
// 	 hb_glyph_info_t *info = buffer.Info;
// 	 hb_glyph_position_t *pos = buffer.pos;
// 	 unsigned int new_len = count + extra_glyphs_needed; // write head during CUT
// 	 unsigned int j = new_len;
// 	 for (unsigned int i = count; i; i--)
// 	 {
// 	   if (!hb_in_range<uint8_t> (info[i - 1].arabic_shaping_action(), STCH_FIXED, STCH_REPEATING))
// 	   {
// 	 if (step == CUT)
// 	 {
// 	   --j;
// 	   info[j] = info[i - 1];
// 	   pos[j] = pos[i - 1];
// 	 }
// 	 continue;
// 	   }

// 	   /* Yay, justification! */

// 	   hb_position_t w_total = 0; // Total to be filled
// 	   hb_position_t w_fixed = 0; // Sum of fixed tiles
// 	   hb_position_t w_repeating = 0; // Sum of repeating tiles
// 	   int n_fixed = 0;
// 	   int n_repeating = 0;

// 	   unsigned int end = i;
// 	   while (i &&
// 		  hb_in_range<uint8_t> (info[i - 1].arabic_shaping_action(), STCH_FIXED, STCH_REPEATING))
// 	   {
// 	 i--;
// 	 hb_position_t width = font.get_glyph_h_advance (info[i].codepoint);
// 	 if (info[i].arabic_shaping_action() == STCH_FIXED)
// 	 {
// 	   w_fixed += width;
// 	   n_fixed++;
// 	 }
// 	 else
// 	 {
// 	   w_repeating += width;
// 	   n_repeating++;
// 	 }
// 	   }
// 	   unsigned int start = i;
// 	   unsigned int context = i;
// 	   while (context &&
// 		  !hb_in_range<uint8_t> (info[context - 1].arabic_shaping_action(), STCH_FIXED, STCH_REPEATING) &&
// 		  (_hb_glyph_info_is_default_ignorable (&info[context - 1]) ||
// 		   HB_ARABIC_GENERAL_CATEGORY_IS_WORD (_hb_glyph_info_get_general_category (&info[context - 1]))))
// 	   {
// 	 context--;
// 	 w_total += pos[context].x_advance;
// 	   }
// 	   i++; // Don't touch i again.

// 	   DEBUG_MSG (ARABIC, nil, "%s stretch at (%d,%d,%d)",
// 		  step == MEASURE ? "measuring" : "cutting", context, start, end);
// 	   DEBUG_MSG (ARABIC, nil, "rest of word:    count=%d width %d", start - context, w_total);
// 	   DEBUG_MSG (ARABIC, nil, "fixed tiles:     count=%d width=%d", n_fixed, w_fixed);
// 	   DEBUG_MSG (ARABIC, nil, "repeating tiles: count=%d width=%d", n_repeating, w_repeating);

// 	   /* Number of additional times to repeat each repeating tile. */
// 	   int n_copies = 0;

// 	   hb_position_t w_remaining = w_total - w_fixed;
// 	   if (sign * w_remaining > sign * w_repeating && sign * w_repeating > 0)
// 	 n_copies = (sign * w_remaining) / (sign * w_repeating) - 1;

// 	   /* See if we can improve the fit by adding an extra repeat and squeezing them together a bit. */
// 	   hb_position_t extra_repeat_overlap = 0;
// 	   hb_position_t shortfall = sign * w_remaining - sign * w_repeating * (n_copies + 1);
// 	   if (shortfall > 0 && n_repeating > 0)
// 	   {
// 	 ++n_copies;
// 	 hb_position_t excess = (n_copies + 1) * sign * w_repeating - sign * w_remaining;
// 	 if (excess > 0)
// 	   extra_repeat_overlap = excess / (n_copies * n_repeating);
// 	   }

// 	   if (step == MEASURE)
// 	   {
// 	 extra_glyphs_needed += n_copies * n_repeating;
// 	 DEBUG_MSG (ARABIC, nil, "will add extra %d copies of repeating tiles", n_copies);
// 	   }
// 	   else
// 	   {
// 	 buffer.unsafe_to_break (context, end);
// 	 hb_position_t x_offset = 0;
// 	 for (unsigned int k = end; k > start; k--)
// 	 {
// 	   hb_position_t width = font.get_glyph_h_advance (info[k - 1].codepoint);

// 	   unsigned int repeat = 1;
// 	   if (info[k - 1].arabic_shaping_action() == STCH_REPEATING)
// 		 repeat += n_copies;

// 	   DEBUG_MSG (ARABIC, nil, "appending %d copies of glyph %d; j=%d",
// 			  repeat, info[k - 1].codepoint, j);
// 	   for (unsigned int n = 0; n < repeat; n++)
// 	   {
// 		 x_offset -= width;
// 		 if (n > 0)
// 		   x_offset += extra_repeat_overlap;
// 		 pos[k - 1].x_offset = x_offset;
// 		 /* Append copy. */
// 		 --j;
// 		 info[j] = info[k - 1];
// 		 pos[j] = pos[k - 1];
// 	   }
// 	 }
// 	   }
// 	 }

// 	 if (step == MEASURE)
// 	 {
// 	   if (unlikely (!buffer.ensure (count + extra_glyphs_needed)))
// 	 break;
// 	 }
// 	 else
// 	 {
// 	   assert (j == 0);
// 	   buffer.len = new_len;
// 	 }
//    }
//  }

//  static void
//  postprocess_glyphs_arabic (plan *hb_ot_shape_plan_t,
// 				common.Buffer              *buffer,
// 				common.Font                *font)
//  {
//    apply_stch (plan, buffer, font);

//    HB_BUFFER_DEALLOCATE_VAR (buffer, arabic_shaping_action);
//  }

//  /* https://www.unicode.org/reports/tr53/ */

//  static hb_codepoint_t
//  modifier_combining_marks[] =
//  {
//    0x0654u, /* ARABIC HAMZA ABOVE */
//    0x0655u, /* ARABIC HAMZA BELOW */
//    0x0658u, /* ARABIC MARK NOON GHUNNA */
//    0x06DCu, /* ARABIC SMALL HIGH SEEN */
//    0x06E3u, /* ARABIC SMALL LOW SEEN */
//    0x06E7u, /* ARABIC SMALL HIGH YEH */
//    0x06E8u, /* ARABIC SMALL HIGH NOON */
//    0x08D3u, /* ARABIC SMALL LOW WAW */
//    0x08F3u, /* ARABIC SMALL HIGH WAW */
//  };

//  static inline bool
//  info_is_mcm (const hb_glyph_info_t &info)
//  {
//    hb_codepoint_t u = info.codepoint;
//    for (unsigned int i = 0; i < ARRAY_LENGTH (modifier_combining_marks); i++)
// 	 if (u == modifier_combining_marks[i])
// 	   return true;
//    return false;
//  }

//  static void
//  reorder_marks_arabic (plan *hb_ot_shape_plan_t HB_UNUSED,
// 			   common.Buffer              *buffer,
// 			   unsigned int              start,
// 			   unsigned int              end)
//  {
//    hb_glyph_info_t *info = buffer.Info;

//    DEBUG_MSG (ARABIC, buffer, "Reordering marks from %d to %d", start, end);

//    unsigned int i = start;
//    for (unsigned int cc = 220; cc <= 230; cc += 10)
//    {
// 	 DEBUG_MSG (ARABIC, buffer, "Looking for %d's starting at %d", cc, i);
// 	 while (i < end && info_cc(info[i]) < cc)
// 	   i++;
// 	 DEBUG_MSG (ARABIC, buffer, "Looking for %d's stopped at %d", cc, i);

// 	 if (i == end)
// 	   break;

// 	 if (info_cc(info[i]) > cc)
// 	   continue;

// 	 unsigned int j = i;
// 	 while (j < end && info_cc(info[j]) == cc && info_is_mcm (info[j]))
// 	   j++;

// 	 if (i == j)
// 	   continue;

// 	 DEBUG_MSG (ARABIC, buffer, "Found %d's from %d to %d", cc, i, j);

// 	 /* Shift it! */
// 	 DEBUG_MSG (ARABIC, buffer, "Shifting %d's: %d %d", cc, i, j);
// 	 hb_glyph_info_t temp[HB_OT_SHAPE_COMPLEX_MAX_COMBINING_MARKS];
// 	 assert (j - i <= ARRAY_LENGTH (temp));
// 	 buffer.merge_clusters (start, j);
// 	 memmove (temp, &info[i], (j - i) * sizeof (hb_glyph_info_t));
// 	 memmove (&info[start + j - i], &info[start], (i - start) * sizeof (hb_glyph_info_t));
// 	 memmove (&info[start], temp, (j - i) * sizeof (hb_glyph_info_t));

// 	 /* Renumber CC such that the reordered sequence is still sorted.
// 	  * 22 and 26 are chosen because they are smaller than all Arabic categories,
// 	  * and are folded back to 220/230 respectively during fallback mark positioning.
// 	  *
// 	  * We do this because the CGJ-handling logic in the normalizer relies on
// 	  * mark sequences having an increasing order even after this reordering.
// 	  * https://github.com/harfbuzz/harfbuzz/issues/554
// 	  * This, however, does break some obscure sequences, where the normalizer
// 	  * might compose a sequence that it should not.  For example, in the seequence
// 	  * ALEF, HAMZAH, MADDAH, we should NOT try to compose ALEF+MADDAH, but with this
// 	  * renumbering, we will.
// 	  */
// 	 unsigned int new_start = start + j - i;
// 	 unsigned int new_cc = cc == 220 ? HB_MODIFIED_COMBINING_CLASS_CCC22 : HB_MODIFIED_COMBINING_CLASS_CCC26;
// 	 while (start < new_start)
// 	 {
// 	   _hb_glyph_info_set_modified_combining_class (&info[start], new_cc);
// 	   start++;
// 	 }

// 	 i = j;
//    }
//  }

//  const hb_ot_complex_shaper_t hbOtComplexShaperArabic =
//  {
//    collect_features_arabic,
//    nil, /* override_features */
//    data_create_arabic,
//    data_destroy_arabic,
//    nil, /* preprocess_text */
//    postprocess_glyphs_arabic,
//    HB_OT_SHAPE_NORMALIZATION_MODE_DEFAULT,
//    nil, /* decompose */
//    nil, /* compose */
//    setup_masks_arabic,
//    newTag_NONE, /* gpos_tag */
//    reorder_marks_arabic,
//    HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_LATE,
//    true, /* fallback_position */
//  };

/* Features ordered the same as the entries in shaping_table rows,
 * followed by rlig.  Don't change. */
var arabicFallbackFeatures = [...]hb_tag_t{
	newTag('i', 'n', 'i', 't'),
	newTag('m', 'e', 'd', 'i'),
	newTag('f', 'i', 'n', 'a'),
	newTag('i', 's', 'o', 'l'),
	newTag('r', 'l', 'i', 'g'),
}

func arabicFallbackSynthesizeLookupSingle(font *common.Font, featureIndex uint16) *truetype.LookupGSUB {
	var (
		glyphs, substitutes [SHAPING_TABLE_LAST - SHAPING_TABLE_FIRST + 1]fonts.GlyphIndex
		numGlyphs           = 0
	)

	// populate arrays
	for u := SHAPING_TABLE_FIRST; u < SHAPING_TABLE_LAST+1; u++ {
		s := shaping_table[u-SHAPING_TABLE_FIRST][featureIndex]
		var u_glyph, s_glyph fonts.GlyphIndex

		if s == 0 || !hb_font_get_glyph(font, u, 0, &u_glyph) || !hb_font_get_glyph(font, s, 0, &s_glyph) ||
			u_glyph == s_glyph || u_glyph > 0xFFFF || s_glyph > 0xFFFF {
			continue
		}

		glyphs[numGlyphs] = u_glyph
		substitutes[numGlyphs] = s_glyph

		numGlyphs++
	}

	if numGlyphs == 0 {
		return nil
	}

	/* Bubble-sort or something equally good!
	* May not be good-enough for presidential candidate interviews, but good-enough for us... */
	//    hb_stable_sort (&glyphs[0], numGlyphs,
	// 		   (int(*)(const OT::HBUINT16*, const OT::HBUINT16 *)) OT::HBGlyphID::cmp,
	// 		   &substitutes[0]);

	// each glyph takes four bytes max, and there's some overhead.
	var buf [(SHAPING_TABLE_LAST-SHAPING_TABLE_FIRST+1)*4 + 128]byte
	c := hb_serialize_context_t(buf, sizeof(buf))
	lookup := c.start_serialize() // <OT::SubstLookup>
	ret := lookup.serialize_single(&c, IgnoreMarks,
		hb_sorted_array(glyphs, numGlyphs),
		hb_array(substitutes, numGlyphs))
	c.end_serialize()

	return c.copy()
}

//  static OT::SubstLookup *
//  arabicFallbackSynthesizeLookupLigature (plan *hb_ot_shape_plan_t HB_UNUSED,
// 						 font *common.Font)
//  {
//    OT::HBGlyphID first_glyphs[ARRAY_LENGTH_CONST (ligature_table)];
//    unsigned int first_glyphs_indirection[ARRAY_LENGTH_CONST (ligature_table)];
//    unsigned int ligature_per_first_glyph_count_list[ARRAY_LENGTH_CONST (first_glyphs)];
//    unsigned int num_first_glyphs = 0;

//    /* We know that all our ligatures are 2-component */
//    OT::HBGlyphID ligature_list[ARRAY_LENGTH_CONST (first_glyphs) * ARRAY_LENGTH_CONST(ligature_table[0].ligatures)];
//    unsigned int component_count_list[ARRAY_LENGTH_CONST (ligature_list)];
//    OT::HBGlyphID component_list[ARRAY_LENGTH_CONST (ligature_list) * 1/* One extra component per ligature */];
//    unsigned int num_ligatures = 0;

//    /* Populate arrays */

//    /* Sort out the first-glyphs */
//    for (unsigned int first_glyph_idx = 0; first_glyph_idx < ARRAY_LENGTH (first_glyphs); first_glyph_idx++)
//    {
// 	 hb_codepoint_t first_u = ligature_table[first_glyph_idx].first;
// 	 hb_codepoint_t first_glyph;
// 	 if (!hb_font_get_glyph (font, first_u, 0, &first_glyph))
// 	   continue;
// 	 first_glyphs[num_first_glyphs] = first_glyph;
// 	 ligature_per_first_glyph_count_list[num_first_glyphs] = 0;
// 	 first_glyphs_indirection[num_first_glyphs] = first_glyph_idx;
// 	 num_first_glyphs++;
//    }
//    hb_stable_sort (&first_glyphs[0], num_first_glyphs,
// 		   (int(*)(const OT::HBUINT16*, const OT::HBUINT16 *)) OT::HBGlyphID::cmp,
// 		   &first_glyphs_indirection[0]);

//    /* Now that the first-glyphs are sorted, walk again, populate ligatures. */
//    for (unsigned int i = 0; i < num_first_glyphs; i++)
//    {
// 	 unsigned int first_glyph_idx = first_glyphs_indirection[i];

// 	 for (unsigned int second_glyph_idx = 0; second_glyph_idx < ARRAY_LENGTH (ligature_table[0].ligatures); second_glyph_idx++)
// 	 {
// 	   hb_codepoint_t second_u   = ligature_table[first_glyph_idx].ligatures[second_glyph_idx].second;
// 	   hb_codepoint_t ligature_u = ligature_table[first_glyph_idx].ligatures[second_glyph_idx].ligature;
// 	   hb_codepoint_t second_glyph, ligature_glyph;
// 	   if (!second_u ||
// 	   !hb_font_get_glyph (font, second_u,   0, &second_glyph) ||
// 	   !hb_font_get_glyph (font, ligature_u, 0, &ligature_glyph))
// 	 continue;

// 	   ligature_per_first_glyph_count_list[i]++;

// 	   ligature_list[num_ligatures] = ligature_glyph;
// 	   component_count_list[num_ligatures] = 2;
// 	   component_list[num_ligatures] = second_glyph;
// 	   num_ligatures++;
// 	 }
//    }

//    if (!num_ligatures)
// 	 return nil;

//    /* 16 bytes per ligature ought to be enough... */
//    char buf[ARRAY_LENGTH_CONST (ligature_list) * 16 + 128];
//    hb_serialize_context_t c (buf, sizeof (buf));
//    OT::SubstLookup *lookup = c.start_serialize<OT::SubstLookup> ();
//    bool ret = lookup.serialize_ligature (&c,
// 					  OT::LookupFlag::IgnoreMarks,
// 					  hb_sorted_array (first_glyphs, num_first_glyphs),
// 					  hb_array (ligature_per_first_glyph_count_list, num_first_glyphs),
// 					  hb_array (ligature_list, num_ligatures),
// 					  hb_array (component_count_list, num_ligatures),
// 					  hb_array (component_list, num_ligatures));
//    c.end_serialize ();
//    /* TODO sanitize the results? */

//    return ret && !c.in_error () ? c.copy<OT::SubstLookup> () : nil;
//  }

func arabicFallbackSynthesizeLookup(plan *hb_ot_shape_plan_t,
	font *common.Font, featureIndex uint16) *truetype.LookupGSUB {
	if featureIndex < 4 {
		return arabicFallbackSynthesizeLookupSingle(plan, font, featureIndex)
	}
	return arabicFallbackSynthesizeLookupLigature(plan, font)
}

const ARABIC_FALLBACK_MAX_LOOKUPS = 5

type arabic_fallback_plan_t struct {
	num_lookups  int
	free_lookups bool

	mask_array   [ARABIC_FALLBACK_MAX_LOOKUPS]common.Mask
	lookup_array [ARABIC_FALLBACK_MAX_LOOKUPS]*truetype.LookupGSUB
	//    OT::hb_ot_layout_lookup_accelerator_t accel_array[ARABIC_FALLBACK_MAX_LOOKUPS];
}

//  #if defined(_WIN32) && !defined(HB_NO_WIN1256)
//  #define HB_WITH_WIN1256
//  #endif

//  #ifdef HB_WITH_WIN1256
//  #include "hb-ot-shape-complex-arabic-win1256.hh"
//  #endif

//  struct ManifestLookup
//  {
//    public:
//    OT::Tag tag;
//    OT::OffsetTo<OT::SubstLookup> lookupOffset;
//    public:
//    DEFINE_SIZE_STATIC (6);
//  };
//  typedef OT::ArrayOf<ManifestLookup> Manifest;

func (fbPlan *arabic_fallback_plan_t) initWin1256(plan *hb_ot_shape_plan_t, font *common.Font) bool {
	// does this font look like it's Windows-1256-encoded?
	g1, _ := font.Face.GetNominalGlyph(0x0627) /* ALEF */
	g2, _ := font.Face.GetNominalGlyph(0x0644) /* LAM */
	g3, _ := font.Face.GetNominalGlyph(0x0649) /* ALEF MAKSURA */
	g4, _ := font.Face.GetNominalGlyph(0x064A) /* YEH */
	g5, _ := font.Face.GetNominalGlyph(0x0652) /* SUKUN */
	if !(g1 == 199 && g2 == 225 && g3 == 236 && g4 == 237 && g5 == 250) {
		return false
	}

	var j int
	for _, man := range arabicWin1256GsubLookups {
		fbPlan.mask_array[j] = plan.map_.get_1_mask(man.tag)
		if fbPlan.mask_array[j] != 0 {
			fbPlan.lookup_array[j] = man.lookup
			if fbPlan.lookup_array[j] != nil {
				fbPlan.accel_array[j].init(*fbPlan.lookup_array[j])
				j++
			}
		}
	}

	fbPlan.num_lookups = j
	fbPlan.free_lookups = false

	return j > 0
}

func (fbPlan *arabic_fallback_plan_t) initUnicode(plan *hb_ot_shape_plan_t, font *common.Font) bool {
	var j int
	for i, feat := range arabicFallbackFeatures {
		fallback_plan.mask_array[j] = plan.map_.get_1_mask(feat)
		if fbPlan.mask_array[j] != 0 {
			fbPlan.lookup_array[j] = arabicFallbackSynthesizeLookup(plan, font, i)
			if fbPlan.lookup_array[j] != nil {
				fbPlan.accel_array[j].init(*fbPlan.lookup_array[j])
				j++
			}
		}
	}

	fbPlan.num_lookups = j
	fbPlan.free_lookups = true

	return j > 0
}

func newArabicFallbackPlan(plan *hb_ot_shape_plan_t, font *common.Font) *arabic_fallback_plan_t {
	var fallback_plan arabic_fallback_plan_t

	/* Try synthesizing GSUB table using Unicode Arabic Presentation Forms,
	* in case the font has cmap entries for the presentation-forms characters. */
	if fallback_plan.initUnicode(plan, font) {
		return &fallback_plan
	}

	/* See if this looks like a Windows-1256-encoded font.  If it does, use a
	* hand-coded GSUB table. */
	if fallback_plan.initWin1256(font) {
		return &fallback_plan
	}

	//    assert (fallback_plan.num_lookups == 0);
	return &arabic_fallback_plan_t{}
}

//  static void
//  arabic_fallback_plan_destroy (plan *arabic_fallback_plan_t)
//  {
//    if (!fallback_plan || fallback_plan.num_lookups == 0)
// 	 return;

//    for (unsigned int i = 0; i < fallback_plan.num_lookups; i++)
// 	 if (fallback_plan.lookup_array[i])
// 	 {
// 	   fallback_plan.accel_array[i].fini ();
// 	   if (fallback_plan.free_lookups)
// 	 free (fallback_plan.lookup_array[i]);
// 	 }

//    free (fallback_plan);
//  }

func (fbPlan *arabic_fallback_plan_t) shape(font *common.Font, buffer *common.Buffer) {
	var c hb_ot_apply_context_t = hb_ot_apply_context_t{0, font, buffer}
	for i := 0; i < fbPlan.num_lookups; i++ {
		if fbPlan.lookup_array[i] {
			c.set_lookup_mask(fbPlan.mask_array[i])
			hb_ot_layout_substitute_lookup(&c, *fbPlan.lookup_array[i], fbPlan.accel_array[i])
		}
	}
}
