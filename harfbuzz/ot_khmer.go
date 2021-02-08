package harfbuzz

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from harfbuzz/src/hb-ot-shape-complex-khmer.cc Copyright © 2011,2012  Google, Inc. Behdad Esfahbod

var _ hb_ot_complex_shaper_t = (*complexShaperKhmer)(nil)

//  Khmer shaper
type complexShaperKhmer struct {
	plan khmerShapePlan
}

var khmerFeatures = [...]hb_ot_map_feature_t{
	/*
	* Basic features.
	* These features are applied in order, one at a time, after reordering.
	 */
	{newTag('p', 'r', 'e', 'f'), F_MANUAL_JOINERS},
	{newTag('b', 'l', 'w', 'f'), F_MANUAL_JOINERS},
	{newTag('a', 'b', 'v', 'f'), F_MANUAL_JOINERS},
	{newTag('p', 's', 't', 'f'), F_MANUAL_JOINERS},
	{newTag('c', 'f', 'a', 'r'), F_MANUAL_JOINERS},
	/*
	* Other features.
	* These features are applied all at once after clearing syllables.
	 */
	{newTag('p', 'r', 'e', 's'), F_GLOBAL_MANUAL_JOINERS},
	{newTag('a', 'b', 'v', 's'), F_GLOBAL_MANUAL_JOINERS},
	{newTag('b', 'l', 'w', 's'), F_GLOBAL_MANUAL_JOINERS},
	{newTag('p', 's', 't', 's'), F_GLOBAL_MANUAL_JOINERS},
}

// Must be in the same order as the khmerFeatures array.
const (
	KHMER_PREF = iota
	KHMER_BLWF
	KHMER_ABVF
	KHMER_PSTF
	KHMER_CFAR

	_KHMER_PRES
	_KHMER_ABVS
	_KHMER_BLWS
	_KHMER_PSTS

	KHMER_NUM_FEATURES
	KHMER_BASIC_FEATURES = _KHMER_PRES /* Don't forget to update this! */
)

func (cs *complexShaperKhmer) collectFeatures(plan *hb_ot_shape_planner_t) {
	map_ := &plan.map_

	/* Do this before any lookups have been applied. */
	map_.add_gsub_pause(setupSyllablesKhmer)
	map_.add_gsub_pause(cs.reorderKhmer)

	/* Testing suggests that Uniscribe does NOT pause between basic
	* features.  Test with KhmerUI.ttf and the following three
	* sequences:
	*
	*   U+1789,U+17BC
	*   U+1789,U+17D2,U+1789
	*   U+1789,U+17D2,U+1789,U+17BC
	*
	* https://github.com/harfbuzz/harfbuzz/issues/974
	 */
	map_.enable_feature(newTag('l', 'o', 'c', 'l'))
	map_.enable_feature(newTag('c', 'c', 'm', 'p'))

	i := 0
	for ; i < KHMER_BASIC_FEATURES; i++ {
		map_.add_feature_ext(khmerFeatures[i].tag, khmerFeatures[i].flags, 1)
	}

	map_.add_gsub_pause(_hb_clear_syllables)

	for ; i < KHMER_NUM_FEATURES; i++ {
		map_.add_feature_ext(khmerFeatures[i].tag, khmerFeatures[i].flags, 1)
	}
}

func (complexShaperKhmer) overrideFeatures(plan *hb_ot_shape_planner_t) {
	map_ := &plan.map_

	/* Khmer spec has 'clig' as part of required shaping features:
	* "Apply feature 'clig' to form ligatures that are desired for
	* typographical correctness.", hence in overrides... */
	map_.enable_feature(newTag('c', 'l', 'i', 'g'))

	/* Uniscribe does not apply 'kern' in Khmer. */
	if uniscribe_bug_compatible {
		map_.disable_feature(newTag('k', 'e', 'r', 'n'))
	}

	map_.disable_feature(newTag('l', 'i', 'g', 'a'))
}

type khmerShapePlan struct {
	virama_glyph fonts.GlyphIndex
	mask_array   [KHMER_NUM_FEATURES]Mask
}

//    bool get_virama_glyph (font * Font, rune *pglyph) const
//    {
// 	 rune glyph = virama_glyph;
// 	 if (unlikely (virama_glyph == (rune) -1))
// 	 {
// 	   if (!font.get_nominal_glyph (0x17D2u, &glyph))
// 	 glyph = 0;
// 	   /* Technically speaking, the spec says we should apply 'locl' to virama too.
// 		* Maybe one day... */

// 	   /* Our get_nominal_glyph() function needs a font, so we can't get the virama glyph
// 		* during shape planning...  Instead, overwrite it here.  It's safe.  Don't worry! */
// 	   virama_glyph = glyph;
// 	 }

// 	 *pglyph = glyph;
// 	 return glyph != 0;
//    }

func (cs *complexShaperKhmer) dataCreate(plan *hb_ot_shape_plan_t) {
	var khmerPlan khmerShapePlan

	khmerPlan.virama_glyph = ^fonts.GlyphIndex(0)

	for i := range khmerPlan.mask_array {
		if khmerFeatures[i].flags&F_GLOBAL == 0 {
			khmerPlan.mask_array[i] = plan.map_.get_1_mask(khmerFeatures[i].tag)
		}
	}

	cs.plan = khmerPlan
}

func (cs *complexShaperKhmer) setupMasks(_ *hb_ot_shape_plan_t, buffer *Buffer, _ *Font) {
	/* We cannot setup masks here.  We save information about characters
	* and setup masks later on in a pause-callback. */

	info := buffer.info
	for i := range info {
		setKhmerProperties(&info[i])
	}
}

/* Note: This enum is duplicated in the -machine.rl source file.
 * Not sure how to avoid duplication. */
const (
	OT_Robatic = 20
	OT_Xgroup  = 21
	OT_Ygroup  = 22
)

func setKhmerProperties(info *GlyphInfo) {
	u := info.codepoint
	type_ := indicGetCategories(u)
	cat := uint8(type_ & 0xFF)
	pos := uint8(type_ >> 8)

	/*
	 * Re-assign category
	 *
	 * These categories are experimentally extracted from what Uniscribe allows.
	 */
	switch u {
	case 0x179A:
		cat = OT_Ra

	case 0x17CC, 0x17C9, 0x17CA:
		cat = OT_Robatic

	case 0x17C6, 0x17CB, 0x17CD, 0x17CE, 0x17CF, 0x17D0, 0x17D1:
		cat = OT_Xgroup

	case 0x17C7, 0x17C8, 0x17DD, 0x17D3: /* Just guessing. Uniscribe doesn't categorize it. */
		cat = OT_Ygroup
	}

	/*
	 * Re-assign position.
	 */
	if cat == OT_M {
		switch pos {
		case POS_PRE_C:
			cat = OT_VPre
		case POS_BELOW_C:
			cat = OT_VBlw
		case POS_ABOVE_C:
			cat = OT_VAbv
		case POS_POST_C:
			cat = OT_VPst
		}
	}

	info.complexCategory = cat
}

func setupSyllablesKhmer(_ *hb_ot_shape_plan_t, font *Font, buffer *Buffer) {
	findSyllablesKhmer(buffer)
	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		buffer.UnsafeToBreak(start, end)
	}
}

func foundSyllableKhmer(syllableType uint8, ts, te int, info []GlyphInfo, syllableSerial *uint8) {
	for i := ts; i < te; i++ {
		info[i].syllable = (*syllableSerial << 4) | syllableType
	}
	*syllableSerial++
	if *syllableSerial == 16 {
		*syllableSerial = 1
	}
}

/* Rules from:
 * https://docs.microsoft.com/en-us/typography/script-development/devanagari */
func (khmerPlan *khmerShapePlan) reorderConsonantSyllable(buffer *Buffer, start, end int) {
	info := buffer.info

	/* Setup masks. */
	{
		/* Post-base */
		mask := khmerPlan.mask_array[KHMER_BLWF] |
			khmerPlan.mask_array[KHMER_ABVF] |
			khmerPlan.mask_array[KHMER_PSTF]
		for i := start + 1; i < end; i++ {
			info[i].mask |= mask
		}
	}

	numCoengs := 0
	for i := start + 1; i < end; i++ {
		/* """
		 * When a COENG + (Cons | IndV) combination are found (and subscript count
		 * is less than two) the character combination is handled according to the
		 * subscript type of the character following the COENG.
		 *
		 * ...
		 *
		 * Subscript Type 2 - The COENG + RO characters are reordered to immediately
		 * before the base glyph. Then the COENG + RO characters are assigned to have
		 * the 'pref' OpenType feature applied to them.
		 * """
		 */
		if info[i].complexCategory == OT_Coeng && numCoengs <= 2 && i+1 < end {
			numCoengs++

			if info[i+1].complexCategory == OT_Ra {
				for j := 0; j < 2; j++ {
					info[i+j].mask |= khmerPlan.mask_array[KHMER_PREF]
				}

				/* Move the Coeng,Ro sequence to the start. */
				buffer.MergeClusters(start, i+2)
				t0 := info[i]
				t1 := info[i+1]
				copy(info[start+2:], info[start:i])
				info[start] = t0
				info[start+1] = t1

				/* Mark the subsequent stuff with 'cfar'.  Used in Khmer.
				 * Read the feature spec.
				 * This allows distinguishing the following cases with MS Khmer fonts:
				 * U+1784,U+17D2,U+179A,U+17D2,U+1782
				 * U+1784,U+17D2,U+1782,U+17D2,U+179A
				 */
				if khmerPlan.mask_array[KHMER_CFAR] != 0 {
					for j := i + 2; j < end; j++ {
						info[j].mask |= khmerPlan.mask_array[KHMER_CFAR]
					}
				}

				numCoengs = 2 /* Done. */
			}
		} else if info[i].complexCategory == OT_VPre { /* Reorder left matra piece. */
			/* Move to the start. */
			buffer.MergeClusters(start, i+1)
			t := info[i]
			copy(info[start+1:], info[start:i])
			info[start] = t
		}
	}
}

func (cs *complexShaperKhmer) reorderSyllableKhmer(buffer *Buffer, start, end int) {
	syllableType := buffer.info[start].syllable & 0x0F
	switch syllableType {
	case khmerBrokenCluster, /* We already inserted dotted-circles, so just call the consonant_syllable. */
		khmerConsonantSyllable:
		cs.plan.reorderConsonantSyllable(buffer, start, end)
	}
}

func (cs *complexShaperKhmer) reorderKhmer(_ *hb_ot_shape_plan_t, font *Font, buffer *Buffer) {
	if debugMode {
		fmt.Println("KHMER - start reordering khmer")
	}

	hb_syllabic_insert_dotted_circles(font, buffer, khmerBrokenCluster, OT_DOTTEDCIRCLE, OT_Repha)
	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		cs.reorderSyllableKhmer(buffer, start, end)
	}

	if debugMode {
		fmt.Println("KHMER - end reordering khmer")
	}
}

func (complexShaperKhmer) decompose(c *hb_ot_shape_normalize_context_t, ab rune) (rune, rune, bool) {
	switch ab {
	/*
	 * Decompose split matras that don't have Unicode decompositions.
	 */

	/* Khmer */
	case 0x17BE:
		return 0x17C1, 0x17BE, true
	case 0x17BF:
		return 0x17C1, 0x17BF, true
	case 0x17C0:
		return 0x17C1, 0x17C0, true
	case 0x17C4:
		return 0x17C1, 0x17C4, true
	case 0x17C5:
		return 0x17C1, 0x17C5, true
	}

	return Uni.Decompose(ab)
}

func (complexShaperKhmer) compose(_ *hb_ot_shape_normalize_context_t, a, b rune) (rune, bool) {
	/* Avoid recomposing split matras. */
	if Uni.GeneralCategory(a).IsMark() {
		return 0, false
	}

	return Uni.Compose(a, b)
}

func (complexShaperKhmer) marksBehavior() (hb_ot_shape_zero_width_marks_type_t, bool) {
	return HB_OT_SHAPE_ZERO_WIDTH_MARKS_NONE, false
}

func (complexShaperKhmer) normalizationPreference() hb_ot_shape_normalization_mode_t {
	return HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS_NO_SHORT_CIRCUIT
}

func (complexShaperKhmer) gposTag() hb_tag_t                                  { return 0 }
func (complexShaperKhmer) preprocessText(*hb_ot_shape_plan_t, *Buffer, *Font) {}
func (complexShaperKhmer) postprocessGlyphs(*hb_ot_shape_plan_t, *Buffer, *Font) {
}
func (complexShaperKhmer) reorderMarks(*hb_ot_shape_plan_t, *Buffer, int, int) {}
