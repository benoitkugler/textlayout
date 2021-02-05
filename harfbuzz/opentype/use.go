package opentype

// ported from harfbuzz/src/hb-ot-shape-complex-use.cc Copyright © 2015  Mozilla Foundation. Google, Inc. Jonathan Kew, Behdad Esfahbod

/*
 * Universal Shaping Engine.
 * https://docs.microsoft.com/en-us/typography/script-development/use
 */

/*
 * Basic features.
 * These features are applied all at once, before reordering.
 */
var useBasicFeatures = [...]hb_tag_t{
	newTag('r', 'k', 'r', 'f'),
	newTag('a', 'b', 'v', 'f'),
	newTag('b', 'l', 'w', 'f'),
	newTag('h', 'a', 'l', 'f'),
	newTag('p', 's', 't', 'f'),
	newTag('v', 'a', 't', 'u'),
	newTag('c', 'j', 'c', 't'),
}

var useTopographicalFeatures = [...]hb_tag_t{
	newTag('i', 's', 'o', 'l'),
	newTag('i', 'n', 'i', 't'),
	newTag('m', 'e', 'd', 'i'),
	newTag('f', 'i', 'n', 'a'),
}

/* Same order as useTopographicalFeatures. */
const (
	JOINING_FORM_ISOL = iota
	JOINING_FORM_INIT
	JOINING_FORM_MEDI
	JOINING_FORM_FINA
	_JOINING_FORM_NONE
)

/*
 * Other features.
 * These features are applied all at once, after reordering and
 * clearing syllables.
 */
var useOtherFeatures = [...]hb_tag_t{
	newTag('a', 'b', 'v', 's'),
	newTag('b', 'l', 'w', 's'),
	newTag('h', 'a', 'l', 'n'),
	newTag('p', 'r', 'e', 's'),
	newTag('p', 's', 't', 's'),
}

type complexShaperUSE struct{}

func (complexShaperUSE) collect_features(plan *hb_ot_shape_planner_t) {
	map_ := &plan.map_

	/* Do this before any lookups have been applied. */
	map_.add_gsub_pause(setupSyllablesUse)

	/* "Default glyph pre-processing group" */
	map_.enable_feature(newTag('l', 'o', 'c', 'l'))
	map_.enable_feature(newTag('c', 'c', 'm', 'p'))
	map_.enable_feature(newTag('n', 'u', 'k', 't'))
	map_.enable_feature(newTag('a', 'k', 'h', 'n'), F_MANUAL_ZWJ)

	/* "Reordering group" */
	map_.add_gsub_pause(_hb_clear_substitution_flags)
	map_.add_feature(newTag('r', 'p', 'h', 'f'), F_MANUAL_ZWJ)
	map_.add_gsub_pause(record_rphf_use)
	map_.add_gsub_pause(_hb_clear_substitution_flags)
	map_.enable_feature(newTag('p', 'r', 'e', 'f'), F_MANUAL_ZWJ)
	map_.add_gsub_pause(record_pref_use)

	/* "Orthographic unit shaping group" */
	for _, basicFeat := range useBasicFeatures {
		map_.enable_feature(basicFeat, F_MANUAL_ZWJ)
	}

	map_.add_gsub_pause(reorder_use)
	map_.add_gsub_pause(_hb_clear_syllables)

	/* "Topographical features" */
	for _, topoFeat := range useTopographicalFeatures {
		map_.add_feature(topoFeat)
	}
	map_.add_gsub_pause(nullptr)

	/* "Standard typographic presentation" */
	for _, otherFeat := range useOtherFeatures {
		map_.enable_feature(otherFeat, F_MANUAL_ZWJ)
	}
}

//  struct use_shape_plan_t
//  {
//    hb_mask_t rphf_mask;

//    arabic_shape_plan_t *arabic_plan;
//  };

//  static void *
//  data_create_use (plan *hb_ot_shape_plan_t)
//  {
//    use_shape_plan_t *use_plan = (use_shape_plan_t *) calloc (1, sizeof (use_shape_plan_t));
//    if (unlikely (!use_plan))
// 	 return nullptr;

//    use_plan.rphf_mask = plan.map.get_1_mask (newTag('r','p','h','f'));

//    if (has_arabic_joining (plan.props.script))
//    {
// 	 use_plan.arabic_plan = (arabic_shape_plan_t *) data_create_arabic (plan);
// 	 if (unlikely (!use_plan.arabic_plan))
// 	 {
// 	   free (use_plan);
// 	   return nullptr;
// 	 }
//    }

//    return use_plan;
//  }

//  static void
//  setup_masks_use (plan *hb_ot_shape_plan_t,
// 		  hb_buffer_t              *buffer,
// 		  hb_font_t                *font HB_UNUSED)
//  {
//    const use_shape_plan_t *use_plan = (const use_shape_plan_t *) plan.data;

//    /* Do this before allocating use_category(). */
//    if (use_plan.arabic_plan)
//    {
// 	 setup_masks_arabic_plan (use_plan.arabic_plan, buffer, plan.props.script);
//    }

//    HB_BUFFER_ALLOCATE_VAR (buffer, use_category);

//    /* We cannot setup masks here.  We save information about characters
// 	* and setup masks later on in a pause-callback. */

//    unsigned int count = buffer.len;
//    hb_glyph_info_t *info = buffer.info;
//    for (unsigned int i = 0; i < count; i++)
// 	 info[i].use_category() = hb_use_get_category (info[i].codepoint);
//  }

//  static void
//  setup_rphf_mask (plan *hb_ot_shape_plan_t,
// 		  buffer *hb_buffer_t)
//  {
//    const use_shape_plan_t *use_plan = (const use_shape_plan_t *) plan.data;

//    hb_mask_t mask = use_plan.rphf_mask;
//    if (!mask) return;

//    hb_glyph_info_t *info = buffer.info;

//    foreach_syllable (buffer, start, end)
//    {
// 	 unsigned int limit = info[start].use_category() == USE(R) ? 1 : hb_min (3u, end - start);
// 	 for (unsigned int i = start; i < start + limit; i++)
// 	   info[i].mask |= mask;
//    }
//  }

//  static void
//  setup_topographical_masks (plan *hb_ot_shape_plan_t,
// 				buffer *hb_buffer_t)
//  {
//    const use_shape_plan_t *use_plan = (const use_shape_plan_t *) plan.data;
//    if (use_plan.arabic_plan)
// 	 return;

//    static_assert ((JOINING_FORM_INIT < 4 && JOINING_FORM_ISOL < 4 && JOINING_FORM_MEDI < 4 && JOINING_FORM_FINA < 4), "");
//    hb_mask_t masks[4], all_masks = 0;
//    for (unsigned int i = 0; i < 4; i++)
//    {
// 	 masks[i] = plan.map.get_1_mask (useTopographicalFeatures[i]);
// 	 if (masks[i] == plan.map.get_global_mask ())
// 	   masks[i] = 0;
// 	 all_masks |= masks[i];
//    }
//    if (!all_masks)
// 	 return;
//    hb_mask_t other_masks = ~all_masks;

//    unsigned int last_start = 0;
//    joining_form_t last_form = _JOINING_FORM_NONE;
//    hb_glyph_info_t *info = buffer.info;
//    foreach_syllable (buffer, start, end)
//    {
// 	 use_syllable_type_t syllable_type = (use_syllable_type_t) (info[start].syllable() & 0x0F);
// 	 switch (syllable_type)
// 	 {
// 	   case use_independent_cluster:
// 	   case use_symbol_cluster:
// 	   case use_hieroglyph_cluster:
// 	   case use_non_cluster:
// 	 /* These don't join.  Nothing to do. */
// 	 last_form = _JOINING_FORM_NONE;
// 	 break;

// 	   case use_virama_terminated_cluster:
// 	   case use_sakot_terminated_cluster:
// 	   case use_standard_cluster:
// 	   case use_number_joiner_terminated_cluster:
// 	   case use_numeral_cluster:
// 	   case use_broken_cluster:

// 	 bool join = last_form == JOINING_FORM_FINA || last_form == JOINING_FORM_ISOL;

// 	 if (join)
// 	 {
// 	   /* Fixup previous syllable's form. */
// 	   last_form = last_form == JOINING_FORM_FINA ? JOINING_FORM_MEDI : JOINING_FORM_INIT;
// 	   for (unsigned int i = last_start; i < start; i++)
// 		 info[i].mask = (info[i].mask & other_masks) | masks[last_form];
// 	 }

// 	 /* Form for this syllable. */
// 	 last_form = join ? JOINING_FORM_FINA : JOINING_FORM_ISOL;
// 	 for (unsigned int i = start; i < end; i++)
// 	   info[i].mask = (info[i].mask & other_masks) | masks[last_form];

// 	 break;
// 	 }

// 	 last_start = start;
//    }
//  }

func setupSyllablesUse(plan *hb_ot_shape_plan_t, _ *hb_font_t, buffer *hb_buffer_t) {
	find_syllables_use(buffer)
	foreach_syllable(buffer, start, end)
	buffer.unsafe_to_break(start, end)
	setup_rphf_mask(plan, buffer)
	setup_topographical_masks(plan, buffer)
}

//  static void
//  record_rphf_use (plan *hb_ot_shape_plan_t,
// 		  hb_font_t *font HB_UNUSED,
// 		  buffer *hb_buffer_t)
//  {
//    const use_shape_plan_t *use_plan = (const use_shape_plan_t *) plan.data;

//    hb_mask_t mask = use_plan.rphf_mask;
//    if (!mask) return;
//    hb_glyph_info_t *info = buffer.info;

//    foreach_syllable (buffer, start, end)
//    {
// 	 /* Mark a substituted repha as USE(R). */
// 	 for (unsigned int i = start; i < end && (info[i].mask & mask); i++)
// 	   if (_hb_glyph_info_substituted (&info[i]))
// 	   {
// 	 info[i].use_category() = USE(R);
// 	 break;
// 	   }
//    }
//  }

//  static void
//  record_pref_use (plan *hb_ot_shape_plan_t HB_UNUSED,
// 		  hb_font_t *font HB_UNUSED,
// 		  buffer *hb_buffer_t)
//  {
//    hb_glyph_info_t *info = buffer.info;

//    foreach_syllable (buffer, start, end)
//    {
// 	 /* Mark a substituted pref as VPre, as they behave the same way. */
// 	 for (unsigned int i = start; i < end; i++)
// 	   if (_hb_glyph_info_substituted (&info[i]))
// 	   {
// 	 info[i].use_category() = USE(VPre);
// 	 break;
// 	   }
//    }
//  }

//  static inline bool
//  is_halant_use (const hb_glyph_info_t &info)
//  {
//    return (info.use_category() == USE(H) || info.use_category() == USE(HVM)) &&
// 	  !_hb_glyph_info_ligated (&info);
//  }

//  static void
//  reorder_syllable_use (buffer *hb_buffer_t, unsigned int start, unsigned int end)
//  {
//    use_syllable_type_t syllable_type = (use_syllable_type_t) (buffer.info[start].syllable() & 0x0F);
//    /* Only a few syllable types need reordering. */
//    if (unlikely (!(FLAG_UNSAFE (syllable_type) &
// 		   (FLAG (use_virama_terminated_cluster) |
// 			FLAG (use_sakot_terminated_cluster) |
// 			FLAG (use_standard_cluster) |
// 			FLAG (use_broken_cluster) |
// 			0))))
// 	 return;

//    hb_glyph_info_t *info = buffer.info;

//  #define POST_BASE_FLAGS64 (FLAG64 (USE(FAbv)) | \
// 				FLAG64 (USE(FBlw)) | \
// 				FLAG64 (USE(FPst)) | \
// 				FLAG64 (USE(MAbv)) | \
// 				FLAG64 (USE(MBlw)) | \
// 				FLAG64 (USE(MPst)) | \
// 				FLAG64 (USE(MPre)) | \
// 				FLAG64 (USE(VAbv)) | \
// 				FLAG64 (USE(VBlw)) | \
// 				FLAG64 (USE(VPst)) | \
// 				FLAG64 (USE(VPre)) | \
// 				FLAG64 (USE(VMAbv)) | \
// 				FLAG64 (USE(VMBlw)) | \
// 				FLAG64 (USE(VMPst)) | \
// 				FLAG64 (USE(VMPre)))

//    /* Move things forward. */
//    if (info[start].use_category() == USE(R) && end - start > 1)
//    {
// 	 /* Got a repha.  Reorder it towards the end, but before the first post-base
// 	  * glyph. */
// 	 for (unsigned int i = start + 1; i < end; i++)
// 	 {
// 	   bool is_post_base_glyph = (FLAG64_UNSAFE (info[i].use_category()) & POST_BASE_FLAGS64) ||
// 				 is_halant_use (info[i]);
// 	   if (is_post_base_glyph || i == end - 1)
// 	   {
// 	 /* If we hit a post-base glyph, move before it; otherwise move to the
// 	  * end. Shift things in between backward. */

// 	 if (is_post_base_glyph)
// 	   i--;

// 	 buffer.merge_clusters (start, i + 1);
// 	 hb_glyph_info_t t = info[start];
// 	 memmove (&info[start], &info[start + 1], (i - start) * sizeof (info[0]));
// 	 info[i] = t;

// 	 break;
// 	   }
// 	 }
//    }

//    /* Move things back. */
//    unsigned int j = start;
//    for (unsigned int i = start; i < end; i++)
//    {
// 	 uint32_t flag = FLAG_UNSAFE (info[i].use_category());
// 	 if (is_halant_use (info[i]))
// 	 {
// 	   /* If we hit a halant, move after it; otherwise move to the beginning, and
// 		* shift things in between forward. */
// 	   j = i + 1;
// 	 }
// 	 else if (((flag) & (FLAG (USE(VPre)) | FLAG (USE(VMPre)))) &&
// 		  /* Only move the first component of a MultipleSubst. */
// 		  0 == _hb_glyph_info_get_lig_comp (&info[i]) &&
// 		  j < i)
// 	 {
// 	   buffer.merge_clusters (j, i + 1);
// 	   hb_glyph_info_t t = info[i];
// 	   memmove (&info[j + 1], &info[j], (i - j) * sizeof (info[0]));
// 	   info[j] = t;
// 	 }
//    }
//  }

//  static void
//  reorder_use (plan *hb_ot_shape_plan_t,
// 		  hb_font_t *font,
// 		  buffer *hb_buffer_t)
//  {
//    if (buffer.message (font, "start reordering USE"))
//    {
// 	 hb_syllabic_insert_dotted_circles (font, buffer,
// 						use_broken_cluster,
// 						USE(B),
// 						USE(R));

// 	 foreach_syllable (buffer, start, end)
// 	   reorder_syllable_use (buffer, start, end);

// 	 (void) buffer.message (font, "end reordering USE");
//    }

//    HB_BUFFER_DEALLOCATE_VAR (buffer, use_category);
//  }

//  static void
//  preprocess_text_use (plan *hb_ot_shape_plan_t,
// 			  hb_buffer_t              *buffer,
// 			  hb_font_t                *font)
//  {
//    _hb_preprocess_text_vowel_constraints (plan, buffer, font);
//  }

//  static bool
//  compose_use (const hb_ot_shape_normalize_context_t *c,
// 		  hb_codepoint_t  a,
// 		  hb_codepoint_t  b,
// 		  hb_codepoint_t *ab)
//  {
//    /* Avoid recomposing split matras. */
//    if (HB_UNICODE_GENERAL_CATEGORY_IS_MARK (c.unicode.general_category (a)))
// 	 return false;

//    return (bool)c.unicode.compose (a, b, ab);
//  }

//  const hb_ot_complex_shaper_t _hb_ot_complex_shaper_use =
//  {
//    collect_features_use,
//    nullptr, /* override_features */
//    data_create_use,
//    data_destroy_use,
//    preprocess_text_use,
//    nullptr, /* postprocess_glyphs */
//    HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS_NO_SHORT_CIRCUIT,
//    nullptr, /* decompose */
//    compose_use,
//    setup_masks_use,
//    newTag_NONE, /* gpos_tag */
//    nullptr, /* reorder_marks */
//    HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_EARLY,
//    false, /* fallback_position */
//  };
