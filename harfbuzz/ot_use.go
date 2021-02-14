package harfbuzz

import (
	"fmt"

	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

// ported from harfbuzz/src/hb-ot-shape-complex-use.cc Copyright © 2015  Mozilla Foundation. Google, Inc. Jonathan Kew, Behdad Esfahbod

/*
 * Universal Shaping Engine.
 * https://docs.microsoft.com/en-us/typography/script-development/use
 */

var _ hb_ot_complex_shaper_t = (*complexShaperUSE)(nil)

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

type useShapePlan struct {
	rphf_mask Mask

	arabic_plan *arabic_shape_plan_t
}

type complexShaperUSE struct {
	complexShaperNil

	plan useShapePlan
}

func (cs *complexShaperUSE) collectFeatures(plan *hb_ot_shape_planner_t) {
	map_ := &plan.map_

	/* Do this before any lookups have been applied. */
	map_.add_gsub_pause(cs.setupSyllablesUse)

	/* "Default glyph pre-processing group" */
	map_.enable_feature(newTag('l', 'o', 'c', 'l'))
	map_.enable_feature(newTag('c', 'c', 'm', 'p'))
	map_.enable_feature(newTag('n', 'u', 'k', 't'))
	map_.enable_feature_ext(newTag('a', 'k', 'h', 'n'), F_MANUAL_ZWJ, 1)

	/* "Reordering group" */
	map_.add_gsub_pause(clearSubstitutionFlags)
	map_.add_feature_ext(newTag('r', 'p', 'h', 'f'), F_MANUAL_ZWJ, 1)
	map_.add_gsub_pause(cs.recordRphfUse)
	map_.add_gsub_pause(clearSubstitutionFlags)
	map_.enable_feature_ext(newTag('p', 'r', 'e', 'f'), F_MANUAL_ZWJ, 1)
	map_.add_gsub_pause(recordPrefUse)

	/* "Orthographic unit shaping group" */
	for _, basicFeat := range useBasicFeatures {
		map_.enable_feature_ext(basicFeat, F_MANUAL_ZWJ, 1)
	}

	map_.add_gsub_pause(reorderUse)
	map_.add_gsub_pause(_hb_clear_syllables)

	/* "Topographical features" */
	for _, topoFeat := range useTopographicalFeatures {
		map_.add_feature(topoFeat)
	}
	map_.add_gsub_pause(nil)

	/* "Standard typographic presentation" */
	for _, otherFeat := range useOtherFeatures {
		map_.enable_feature_ext(otherFeat, F_MANUAL_ZWJ, 1)
	}
}

func (cs *complexShaperUSE) dataCreate(plan *hb_ot_shape_plan_t) {
	var usePlan useShapePlan

	usePlan.rphf_mask = plan.map_.get_1_mask(newTag('r', 'p', 'h', 'f'))

	if ucd.HasArabicJoining(plan.props.Script) {
		pl := newArabicPlan(plan)
		usePlan.arabic_plan = &pl
	}

	cs.plan = usePlan
}

func (cs *complexShaperUSE) setupMasks(plan *hb_ot_shape_plan_t, buffer *Buffer, _ *Font) {
	use_plan := cs.plan
	/* Do this before allocating complexCategory. */
	if use_plan.arabic_plan != nil {
		use_plan.arabic_plan.setupMasks(buffer, plan.props.Script)
	}

	/* We cannot setup masks here.  We save information about characters
	* and setup masks later on in a pause-callback. */

	info := buffer.Info
	for i := range info {
		info[i].complexCategory = getUSECategory(info[i].codepoint)
	}
}

func (cs *complexShaperUSE) setupRphfMask(buffer *Buffer) {
	use_plan := cs.plan

	mask := use_plan.rphf_mask
	if mask == 0 {
		return
	}

	info := buffer.Info
	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		limit := 1
		if info[start].complexCategory != useSyllableMachine_ex_R {
			limit = min(3, end-start)
		}
		for i := start; i < start+limit; i++ {
			info[i].mask |= mask
		}
	}
}

func (cs *complexShaperUSE) setupTopographicalMasks(plan *hb_ot_shape_plan_t, buffer *Buffer) {
	if cs.plan.arabic_plan != nil {
		return
	}
	var (
		masks    [4]Mask
		allMasks uint32
	)
	for i := range masks {
		masks[i] = plan.map_.get_1_mask(useTopographicalFeatures[i])
		if masks[i] == plan.map_.global_mask {
			masks[i] = 0
		}
		allMasks |= masks[i]
	}
	if allMasks == 0 {
		return
	}
	otherMasks := ^allMasks

	lastStart := 0
	lastForm := _JOINING_FORM_NONE
	info := buffer.Info
	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		syllableType := info[start].syllable & 0x0F
		switch syllableType {
		case useIndependentCluster, useSymbolCluster, useHieroglyphCluster, useNonCluster:
			// these don't join.  Nothing to do.
			lastForm = _JOINING_FORM_NONE

		case useViramaTerminatedCluster, useSakotTerminatedCluster, useStandardCluster, useNumberJoinerTerminatedCluster, useNumeralCluster, useBrokenCluster:
			join := lastForm == JOINING_FORM_FINA || lastForm == JOINING_FORM_ISOL
			if join {
				// fixup previous syllable's form.
				if lastForm == JOINING_FORM_FINA {
					lastForm = JOINING_FORM_MEDI
				} else {
					lastForm = JOINING_FORM_INIT
				}
				for i := lastStart; i < start; i++ {
					info[i].mask = (info[i].mask & otherMasks) | masks[lastForm]
				}
			}

			// form for this syllable.
			lastForm = JOINING_FORM_ISOL
			if join {
				lastForm = JOINING_FORM_FINA
			}
			for i := start; i < end; i++ {
				info[i].mask = (info[i].mask & otherMasks) | masks[lastForm]
			}
		}

		lastStart = start
	}
}

func (cs *complexShaperUSE) setupSyllablesUse(plan *hb_ot_shape_plan_t, _ *Font, buffer *Buffer) {
	findSyllablesUse(buffer)
	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		buffer.unsafeToBreak(start, end)
	}
	cs.setupRphfMask(buffer)
	cs.setupTopographicalMasks(plan, buffer)
}

func (cs *complexShaperUSE) recordRphfUse(plan *hb_ot_shape_plan_t, _ *Font, buffer *Buffer) {
	use_plan := cs.plan

	mask := use_plan.rphf_mask
	if mask == 0 {
		return
	}
	info := buffer.Info

	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		// mark a substituted repha as USE(R).
		for i := start; i < end && (info[i].mask&mask) != 0; i++ {
			if glyphInfoSubstituted(&info[i]) {
				info[i].complexCategory = useSyllableMachine_ex_R
				break
			}
		}
	}
}

func recordPrefUse(_ *hb_ot_shape_plan_t, _ *Font, buffer *Buffer) {
	info := buffer.Info

	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		// mark a substituted pref as VPre, as they behave the same way.
		for i := start; i < end; i++ {
			if glyphInfoSubstituted(&info[i]) {
				info[i].complexCategory = useSyllableMachine_ex_VPre
				break
			}
		}
	}
}

func isHalantUse(info *GlyphInfo) bool {
	return (info.complexCategory == useSyllableMachine_ex_H || info.complexCategory == useSyllableMachine_ex_HVM) &&
		!info.Ligated()
}

func reorderSyllableUse(buffer *Buffer, start, end int) {
	syllableType := (buffer.Info[start].syllable & 0x0F)
	/* Only a few syllable types need reordering. */
	const mask = 1<<useViramaTerminatedCluster |
		1<<useSakotTerminatedCluster |
		1<<useStandardCluster |
		1<<useBrokenCluster
	if 1<<syllableType&mask == 0 {
		return
	}

	info := buffer.Info

	const postBaseFlags64 = (1<<useSyllableMachine_ex_FAbv |
		1<<useSyllableMachine_ex_FBlw |
		1<<useSyllableMachine_ex_FPst |
		1<<useSyllableMachine_ex_MAbv |
		1<<useSyllableMachine_ex_MBlw |
		1<<useSyllableMachine_ex_MPst |
		1<<useSyllableMachine_ex_MPre |
		1<<useSyllableMachine_ex_VAbv |
		1<<useSyllableMachine_ex_VBlw |
		1<<useSyllableMachine_ex_VPst |
		1<<useSyllableMachine_ex_VPre |
		1<<useSyllableMachine_ex_VMAbv |
		1<<useSyllableMachine_ex_VMBlw |
		1<<useSyllableMachine_ex_VMPst |
		1<<useSyllableMachine_ex_VMPre)

	/* Move things forward. */
	if info[start].complexCategory == useSyllableMachine_ex_R && end-start > 1 {
		/* Got a repha.  Reorder it towards the end, but before the first post-base
		 * glyph. */
		for i := start + 1; i < end; i++ {
			isPostBaseGlyph := (1<<(info[i].complexCategory)&postBaseFlags64) != 0 ||
				isHalantUse(&info[i])
			if isPostBaseGlyph || i == end-1 {
				/* If we hit a post-base glyph, move before it; otherwise move to the
				 * end. Shift things in between backward. */

				if isPostBaseGlyph {
					i--
				}

				buffer.mergeClusters(start, i+1)
				t := info[start]
				copy(info[start:i], info[start+1:])
				info[i] = t

				break
			}
		}
	}

	/* Move things back. */
	j := start
	for i := start; i < end; i++ {
		flag := 1 << (info[i].complexCategory)
		if isHalantUse(&info[i]) {
			/* If we hit a halant, move after it; otherwise move to the beginning, and
			* shift things in between forward. */
			j = i + 1
		} else if flag&(1<<useSyllableMachine_ex_VPre|1<<useSyllableMachine_ex_VMPre) != 0 &&
			/* Only move the first component of a MultipleSubst. */
			0 == info[i].getLigComp() && j < i {
			buffer.mergeClusters(j, i+1)
			t := info[i]
			copy(info[j+1:], info[j:i])
			info[j] = t
		}
	}
}

func reorderUse(plan *hb_ot_shape_plan_t, font *Font, buffer *Buffer) {
	if debugMode {
		fmt.Println("USE - start reordering USE")
	}
	hb_syllabic_insert_dotted_circles(font, buffer, useBrokenCluster,
		useSyllableMachine_ex_B, useSyllableMachine_ex_R)

	iter, count := buffer.SyllableIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
		reorderSyllableUse(buffer, start, end)
	}
	if debugMode {
		fmt.Println("USE - end reordering USE")
	}
}

func (cs *complexShaperUSE) preprocessText(_ *hb_ot_shape_plan_t, buffer *Buffer, _ *Font) {
	preprocessTextVowelConstraints(buffer)
}

func (cs *complexShaperUSE) compose(_ *hb_ot_shape_normalize_context_t, a, b rune) (rune, bool) {
	// avoid recomposing split matras.
	if Uni.generalCategory(a).isMark() {
		return 0, false
	}

	return Uni.Compose(a, b)
}

func (complexShaperUSE) marksBehavior() (hb_ot_shape_zero_width_marks_type_t, bool) {
	return HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_EARLY, false
}

func (complexShaperUSE) normalizationPreference() hb_ot_shape_normalization_mode_t {
	return HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS_NO_SHORT_CIRCUIT
}