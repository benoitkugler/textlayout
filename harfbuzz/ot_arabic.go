package harfbuzz

import (
	"fmt"
	"sort"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

// ported from harfbuzz/src/hb-ot-shape-complex-arabic.cc, hb-ot-shape-complex-arabic-fallback.hh Copyright © 2010,2012  Google, Inc. Behdad Esfahbod

var _ hb_ot_complex_shaper_t = (*complexShaperArabic)(nil)

const flagArabicHasStch = HB_BUFFER_SCRATCH_FLAG_COMPLEX0

/* See:
 * https://github.com/harfbuzz/harfbuzz/commit/6e6f82b6f3dde0fc6c3c7d991d9ec6cfff57823d#commitcomment-14248516 */
func isWord(genCat GeneralCategory) bool {
	const mask = 1<<Unassigned |
		1<<PrivateUse |
		/*1 <<  LowercaseLetter |*/
		1<<ModifierLetter |
		1<<OtherLetter |
		/*1 <<  TitlecaseLetter |*/
		/*1 <<  UppercaseLetter |*/
		1<<SpacingMark |
		1<<EnclosingMark |
		1<<NonSpacingMark |
		1<<DecimalNumber |
		1<<LetterNumber |
		1<<OtherNumber |
		1<<CurrencySymbol |
		1<<ModifierSymbol |
		1<<MathSymbol |
		1<<OtherSymbol
	return (1<<genCat)&mask != 0
}

/*
 * Joining types:
 */

// index into arabic_state_table
const (
	joiningTypeU = iota
	joiningTypeL
	joiningTypeR
	joiningTypeD
	joiningGroupAlaph
	joiningGroupDalathRish
	numStateMachineCols
	joiningTypeT
	joiningTypeC = joiningTypeD
)

func getJoiningType(u rune, genCat GeneralCategory) uint8 {
	if jType, ok := ucd.ArabicJoinings[u]; ok {
		switch jType {
		case ucd.U:
			return joiningTypeU
		case ucd.L:
			return joiningTypeL
		case ucd.R:
			return joiningTypeR
		case ucd.D:
			return joiningTypeD
		case ucd.Alaph:
			return joiningGroupAlaph
		case ucd.DalathRish:
			return joiningGroupDalathRish
		case ucd.T:
			return joiningTypeT
		case ucd.C:
			return joiningTypeC
		}
	}

	const mask = 1<<NonSpacingMark | 1<<EnclosingMark | 1<<Format
	if 1<<genCat&mask != 0 {
		return joiningTypeT
	}
	return joiningTypeU
}

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

var arabic_state_table = [...][numStateMachineCols]struct {
	prev_action uint8
	curr_action uint8
	next_state  uint16
}{
	/*   jt_U,          jt_L,          jt_R,          jt_D,          jg_ALAPH,      jg_DALATH_RISH */

	/* State 0: prev was U, not willing to join. */
	{{NONE, NONE, 0}, {NONE, ISOL, 2}, {NONE, ISOL, 1}, {NONE, ISOL, 2}, {NONE, ISOL, 1}, {NONE, ISOL, 6}},

	/* State 1: prev was R or ISOL/ALAPH, not willing to join. */
	{{NONE, NONE, 0}, {NONE, ISOL, 2}, {NONE, ISOL, 1}, {NONE, ISOL, 2}, {NONE, FIN2, 5}, {NONE, ISOL, 6}},

	/* State 2: prev was D/L in ISOL form, willing to join. */
	{{NONE, NONE, 0}, {NONE, ISOL, 2}, {INIT, FINA, 1}, {INIT, FINA, 3}, {INIT, FINA, 4}, {INIT, FINA, 6}},

	/* State 3: prev was D in FINA form, willing to join. */
	{{NONE, NONE, 0}, {NONE, ISOL, 2}, {MEDI, FINA, 1}, {MEDI, FINA, 3}, {MEDI, FINA, 4}, {MEDI, FINA, 6}},

	/* State 4: prev was FINA ALAPH, not willing to join. */
	{{NONE, NONE, 0}, {NONE, ISOL, 2}, {MED2, ISOL, 1}, {MED2, ISOL, 2}, {MED2, FIN2, 5}, {MED2, ISOL, 6}},

	/* State 5: prev was FIN2/FIN3 ALAPH, not willing to join. */
	{{NONE, NONE, 0}, {NONE, ISOL, 2}, {ISOL, ISOL, 1}, {ISOL, ISOL, 2}, {ISOL, FIN2, 5}, {ISOL, ISOL, 6}},

	/* State 6: prev was DALATH/RISH, not willing to join. */
	{{NONE, NONE, 0}, {NONE, ISOL, 2}, {NONE, ISOL, 1}, {NONE, ISOL, 2}, {NONE, FIN3, 5}, {NONE, ISOL, 6}},
}

type complexShaperArabic struct {
	complexShaperNil

	plan arabic_shape_plan_t
}

func (complexShaperArabic) marksBehavior() (hb_ot_shape_zero_width_marks_type_t, bool) {
	return HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_LATE, true
}

func (complexShaperArabic) normalizationPreference() hb_ot_shape_normalization_mode_t {
	return HB_OT_SHAPE_NORMALIZATION_MODE_DEFAULT
}

func (cs *complexShaperArabic) collectFeatures(plan *hb_ot_shape_planner_t) {
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
		has_fallback := plan.props.Script == language.Arabic && !featureIsSyriac(arabFeat)
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

	if plan.props.Script == language.Arabic {
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
	mask_array [len(arabic_features) + 1]Mask

	fallback_plan *arabic_fallback_plan_t

	do_fallback bool
	has_stch    bool
}

func newArabicPlan(plan *hb_ot_shape_plan_t) arabic_shape_plan_t {
	var arabicPlan arabic_shape_plan_t

	arabicPlan.do_fallback = plan.props.Script == language.Arabic
	arabicPlan.has_stch = plan.map_.get_1_mask(newTag('s', 't', 'c', 'h')) != 0
	for i, arabFeat := range arabic_features {
		arabicPlan.mask_array[i] = plan.map_.get_1_mask(arabFeat)
		arabicPlan.do_fallback = arabicPlan.do_fallback &&
			(featureIsSyriac(arabFeat) || plan.map_.needs_fallback(arabFeat))
	}
	return arabicPlan
}

func (cs *complexShaperArabic) dataCreate(plan *hb_ot_shape_plan_t) {
	cs.plan = newArabicPlan(plan)
}

func arabicJoining(buffer *Buffer) {
	info := buffer.Info
	prev, state := -1, uint16(0)

	// check pre-context
	for _, u := range buffer.Context[0] {
		thisType := getJoiningType(u, Uni.GeneralCategory(u))

		if thisType == joiningTypeT {
			continue
		}

		entry := &arabic_state_table[state][thisType]
		state = entry.next_state
		break
	}

	for i := 0; i < len(info); i++ {
		thisType := getJoiningType(info[i].Codepoint, info[i].Unicode.GeneralCategory())

		if thisType == joiningTypeT {
			info[i].ComplexAux = NONE
			continue
		}

		entry := &arabic_state_table[state][thisType]

		if entry.prev_action != NONE && prev != -1 {
			info[prev].ComplexAux = entry.prev_action
			buffer.UnsafeToBreak(prev, i+1)
		}

		info[i].ComplexAux = entry.curr_action

		prev = i
		state = entry.next_state
	}

	for _, u := range buffer.Context[1] {
		thisType := getJoiningType(u, Uni.GeneralCategory(u))

		if thisType == joiningTypeT {
			continue
		}

		entry := &arabic_state_table[state][thisType]
		if entry.prev_action != NONE && prev != -1 {
			info[prev].ComplexAux = entry.prev_action
		}
		break
	}
}

func mongolianVariationSelectors(buffer *Buffer) {
	// copy ComplexAux from base to Mongolian variation selectors.
	info := buffer.Info
	for i := 1; i < len(info); i++ {
		if cp := info[i].Codepoint; 0x180B <= cp && cp <= 0x180D {
			info[i].ComplexAux = info[i-1].ComplexAux
		}
	}
}

func (arabicPlan arabic_shape_plan_t) setupMasks(buffer *Buffer, script language.Script) {
	arabicJoining(buffer)
	if script == language.Mongolian {
		mongolianVariationSelectors(buffer)
	}

	info := buffer.Info
	for i := range info {
		info[i].Mask |= arabicPlan.mask_array[info[i].ComplexAux]
	}
}

func (cs *complexShaperArabic) setupMasks(plan *hb_ot_shape_plan_t, buffer *Buffer, _ *Font) {
	cs.plan.setupMasks(buffer, plan.props.Script)
}

func arabicFallbackShape(plan *hb_ot_shape_plan_t, font *Font, buffer *Buffer) {
	arabicPlan := plan.shaper.(*complexShaperArabic).plan

	if !arabicPlan.do_fallback {
		return
	}

	fallback_plan := arabicPlan.fallback_plan
	if fallback_plan == nil {
		// this sucks. We need a font to build the fallback plan...
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

func recordStch(plan *hb_ot_shape_plan_t, _ *Font, buffer *Buffer) {
	arabic_plan := plan.shaper.(*complexShaperArabic).plan
	if !arabic_plan.has_stch {
		return
	}

	/* 'stch' feature was just applied.  Look for anything that multiplied,
	* and record it for stch treatment later.  Note that rtlm, frac, etc
	* are applied before stch, but we assume that they didn't result in
	* anything multiplying into 5 pieces, so it's safe-ish... */

	info := buffer.Info
	for i := range info {
		if info[i].Multiplied() {
			comp := info[i].GetLigComp()
			if comp%2 != 0 {
				info[i].ComplexCategory = STCH_REPEATING
			} else {
				info[i].ComplexCategory = STCH_FIXED
			}
			buffer.ScratchFlags |= flagArabicHasStch
		}
	}
}

func inRange(sa uint8) bool {
	return STCH_FIXED <= sa && sa <= STCH_REPEATING
}

func (cs *complexShaperArabic) postprocessGlyphs(plan *hb_ot_shape_plan_t, buffer *Buffer, font *Font) {
	if buffer.ScratchFlags&flagArabicHasStch == 0 {
		return
	}

	/* The Arabic shaper currently always processes in RTL mode, so we should
	* stretch / position the stretched pieces to the left / preceding glyphs. */

	/* We do a two pass implementation:
	* First pass calculates the exact number of extra glyphs we need,
	* We then enlarge buffer to have that much room,
	* Second pass applies the stretch, copying things to the end of buffer. */

	sign := Position(+1)
	if font.XScale < 0 {
		sign = -1
	}
	const (
		MEASURE = iota
		CUT
	)
	var (
		originCount       = len(buffer.Info) // before enlarging
		extraGlyphsNeeded = 0                // Set during MEASURE, used during CUT
	)
	for step := MEASURE; step <= CUT; step++ {
		info := buffer.Info
		pos := buffer.Pos
		j := len(info) // enlarged after MEASURE
		for i := originCount; i != 0; i-- {
			if sa := info[i-1].ComplexAux; !inRange(sa) {
				if step == CUT {
					j--
					info[j] = info[i-1]
					pos[j] = pos[i-1]
				}
				continue
			}

			/* Yay, justification! */
			var (
				wTotal     Position // Total to be filled
				wFixed     Position // Sum of fixed tiles
				wRepeating Position // Sum of repeating tiles
				nFixed     = 0
				nRepeating = 0
			)
			end := i
			for i != 0 && inRange(info[i-1].ComplexAux) {
				i--
				width := font.GetGlyphHAdvance(info[i].Codepoint)
				if info[i].ComplexAux == STCH_FIXED {
					wFixed += width
					nFixed++
				} else {
					wRepeating += width
					nRepeating++
				}
			}
			start := i
			context := i
			for context != 0 && !inRange(info[context-1].ComplexAux) &&
				((&info[context-1]).IsDefaultIgnorable() ||
					isWord((&info[context-1]).Unicode.GeneralCategory())) {
				context--
				wTotal += pos[context].XAdvance
			}
			i++ // Don't touch i again.

			if debugMode {
				fmt.Printf("ARABIC - step %d: stretch at (%d,%d,%d)\n", step+1, context, start, end)
				fmt.Printf("ARABIC - rest of word:    count=%d width %d\n", start-context, wTotal)
				fmt.Printf("ARABIC - fixed tiles:     count=%d width=%d\n", nFixed, wFixed)
				fmt.Printf("ARABIC - repeating tiles: count=%d width=%d\n", nRepeating, wRepeating)
			}

			// number of additional times to repeat each repeating tile.
			var nCopies int

			wRemaining := wTotal - wFixed
			if sign*wRemaining > sign*wRepeating && sign*wRepeating > 0 {
				nCopies = int((sign*wRemaining)/(sign*wRepeating) - 1)
			}

			// see if we can improve the fit by adding an extra repeat and squeezing them together a bit.
			var extraRepeatOverlap Position
			shortfall := sign*wRemaining - sign*wRepeating*(Position(nCopies)+1)
			if shortfall > 0 && nRepeating > 0 {
				nCopies++
				excess := (Position(nCopies)+1)*sign*wRepeating - sign*wRemaining
				if excess > 0 {
					extraRepeatOverlap = excess / Position(nCopies*nRepeating)
				}
			}

			if step == MEASURE {
				extraGlyphsNeeded += nCopies * nRepeating
				if debugMode {
					fmt.Printf("ARABIC - will add extra %d copies of repeating tiles\n", nCopies)
				}
			} else {
				buffer.UnsafeToBreak(context, end)
				var xOffset Position
				for k := end; k > start; k-- {
					width := font.GetGlyphHAdvance(info[k-1].Codepoint)

					repeat := 1
					if info[k-1].ComplexAux == STCH_REPEATING {
						repeat += nCopies
					}

					if debugMode {
						fmt.Printf("ARABIC - appending %d copies of glyph %d; j=%d\n", repeat, info[k-1].Codepoint, j)
					}
					for n := 0; n < repeat; n++ {
						xOffset -= width
						if n > 0 {
							xOffset += extraRepeatOverlap
						}
						pos[k-1].XOffset = xOffset
						// append copy.
						j--
						info[j] = info[k-1]
						pos[j] = pos[k-1]
					}
				}
			}
		}

		if step == MEASURE {
			buffer.Ensure(originCount + extraGlyphsNeeded)
		}
	}
}

// https://www.unicode.org/reports/tr53/
var modifierCombiningMarks = [...]rune{
	0x0654, /* ARABIC HAMZA ABOVE */
	0x0655, /* ARABIC HAMZA BELOW */
	0x0658, /* ARABIC MARK NOON GHUNNA */
	0x06DC, /* ARABIC SMALL HIGH SEEN */
	0x06E3, /* ARABIC SMALL LOW SEEN */
	0x06E7, /* ARABIC SMALL HIGH YEH */
	0x06E8, /* ARABIC SMALL HIGH NOON */
	0x08D3, /* ARABIC SMALL LOW WAW */
	0x08F3, /* ARABIC SMALL HIGH WAW */
}

func infoIsMcm(info *GlyphInfo) bool {
	u := info.Codepoint
	for i := 0; i < len(modifierCombiningMarks); i++ {
		if u == modifierCombiningMarks[i] {
			return true
		}
	}
	return false
}

func (cs *complexShaperArabic) reorderMarks(_ *hb_ot_shape_plan_t, buffer *Buffer, start, end int) {
	info := buffer.Info

	if debugMode {
		fmt.Printf("ARABIC - Reordering marks from %d to %d\n", start, end)
	}

	i := start
	for cc := uint8(220); cc <= 230; cc += 10 {
		if debugMode {
			fmt.Printf("ARABIC - Looking for %d's starting at %d\n", cc, i)
		}
		for i < end && info[i].GetModifiedCombiningClass() < cc {
			i++
		}
		if debugMode {
			fmt.Printf("ARABIC - Looking for %d's stopped at %d\n", cc, i)
		}

		if i == end {
			break
		}

		if info[i].GetModifiedCombiningClass() > cc {
			continue
		}

		j := i
		for j < end && info[j].GetModifiedCombiningClass() == cc && infoIsMcm(&info[j]) {
			j++
		}

		if i == j {
			continue
		}

		if debugMode {
			fmt.Printf("ARABIC - Found %d's from %d to %d", cc, i, j)
			// shift it!
			fmt.Printf("ARABIC - Shifting %d's: %d %d", cc, i, j)
		}

		var temp [shapeComplexMaxCombiningMarks]GlyphInfo
		//  assert (j - i <= len (temp));
		buffer.MergeClusters(start, j)
		copy(temp[:j-i], info[i:])
		copy(info[start+j-i:], info[start:i])
		copy(info[start:], temp[:j-i])

		/* Renumber CC such that the reordered sequence is still sorted.
		 * 22 and 26 are chosen because they are smaller than all Arabic categories,
		 * and are folded back to 220/230 respectively during fallback mark positioning.
		 *
		 * We do this because the CGJ-handling logic in the normalizer relies on
		 * mark sequences having an increasing order even after this reordering.
		 * https://github.com/harfbuzz/harfbuzz/issues/554
		 * This, however, does break some obscure sequences, where the normalizer
		 * might compose a sequence that it should not.  For example, in the seequence
		 * ALEF, HAMZAH, MADDAH, we should NOT try to compose ALEF+MADDAH, but with this
		 * renumbering, we will. */
		newStart := start + j - i
		newCc := Mcc26
		if cc == 220 {
			newCc = Mcc26
		}
		for start < newStart {
			info[start].SetModifiedCombiningClass(newCc)
			start++
		}

		i = j
	}
}

/* Features ordered the same as the entries in ucd.ArabicShaping rows,
 * followed by rlig.  Don't change. */
var arabicFallbackFeatures = [...]hb_tag_t{
	newTag('i', 's', 'o', 'l'),
	newTag('f', 'i', 'n', 'a'),
	newTag('i', 'n', 'i', 't'),
	newTag('m', 'e', 'd', 'i'),
	newTag('r', 'l', 'i', 'g'),
}

// used to sort both array at the same time
type jointGlyphs struct {
	glyphs, substitutes []fonts.GlyphIndex
}

func (a jointGlyphs) Len() int { return len(a.glyphs) }
func (a jointGlyphs) Swap(i, j int) {
	a.glyphs[i], a.glyphs[j] = a.glyphs[j], a.glyphs[i]
	a.substitutes[i], a.substitutes[j] = a.substitutes[j], a.substitutes[i]
}
func (a jointGlyphs) Less(i, j int) bool { return a.glyphs[i] < a.glyphs[j] }

func arabicFallbackSynthesizeLookupSingle(font *Font, featureIndex int) *lookupGSUB {
	var glyphs, substitutes []fonts.GlyphIndex

	// populate arrays
	for u := rune(ucd.FirstArabicShape); u <= ucd.LastArabicShape; u++ {
		s := rune(ucd.ArabicShaping[u-ucd.FirstArabicShape][featureIndex])
		uGlyph, hasU := font.Face.GetNominalGlyph(u)
		sGlyph, hasS := font.Face.GetNominalGlyph(s)

		if s == 0 || !hasU || !hasS || uGlyph == sGlyph || uGlyph > 0xFFFF || sGlyph > 0xFFFF {
			continue
		}

		glyphs = append(glyphs, uGlyph)
		substitutes = append(substitutes, sGlyph)
	}

	if len(glyphs) == 0 {
		return nil
	}

	sort.Stable(jointGlyphs{glyphs: glyphs, substitutes: substitutes})

	return &lookupGSUB{
		Flag: truetype.IgnoreMarks,
		Subtables: []truetype.LookupGSUBSubtable{{
			Coverage: truetype.CoverageList(glyphs),
			Data:     truetype.SingleSubstitution2(substitutes),
		}},
	}
}

// used to sort both array at the same time
type glyphsIndirections struct {
	glyphs       []fonts.GlyphIndex
	indirections []int
}

func (a glyphsIndirections) Len() int { return len(a.glyphs) }
func (a glyphsIndirections) Swap(i, j int) {
	a.glyphs[i], a.glyphs[j] = a.glyphs[j], a.glyphs[i]
	a.indirections[i], a.indirections[j] = a.indirections[j], a.indirections[i]
}
func (a glyphsIndirections) Less(i, j int) bool { return a.glyphs[i] < a.glyphs[j] }

func arabicFallbackSynthesizeLookupLigature(font *Font) *lookupGSUB {
	var (
		firstGlyphs            truetype.CoverageList
		firstGlyphsIndirection []int // original index into ArabicLigatures
	)

	/* Populate arrays */

	// sort out the first-glyphs
	for firstGlyphIdx, lig := range ucd.ArabicLigatures {
		firstGlyph, ok := font.Face.GetNominalGlyph(lig.First)
		if !ok {
			continue
		}
		firstGlyphs = append(firstGlyphs, firstGlyph)
		firstGlyphsIndirection = append(firstGlyphsIndirection, firstGlyphIdx)
	}

	if len(firstGlyphs) == 0 {
		return nil
	}

	sort.Stable(glyphsIndirections{glyphs: firstGlyphs, indirections: firstGlyphsIndirection})

	var out truetype.SubstitutionLigature

	// ow that the first-glyphs are sorted, walk again, populate ligatures.
	for _, firstGlyphIdx := range firstGlyphsIndirection {
		ligs := ucd.ArabicLigatures[firstGlyphIdx].Ligatures
		var ligatureSet []truetype.LigatureGlyph
		for _, v := range ligs {
			secondU, ligatureU := v[0], v[1]
			secondGlyph, hasSecond := font.Face.GetNominalGlyph(secondU)
			ligatureGlyph, hasLigature := font.Face.GetNominalGlyph(ligatureU)
			if secondU == 0 || !hasSecond || !hasLigature {
				continue
			}
			ligatureSet = append(ligatureSet, truetype.LigatureGlyph{
				Glyph:      ligatureGlyph,
				Components: []fonts.GlyphIndex{secondGlyph}, // ligatures are 2-component
			})
		}
		out = append(out, ligatureSet)
	}

	return &lookupGSUB{Flag: truetype.IgnoreMarks, Subtables: []truetype.LookupGSUBSubtable{
		{Coverage: firstGlyphs, Data: out},
	}}
}

func arabicFallbackSynthesizeLookup(font *Font, featureIndex int) *lookupGSUB {
	if featureIndex < 4 {
		return arabicFallbackSynthesizeLookupSingle(font, featureIndex)
	}
	return arabicFallbackSynthesizeLookupLigature(font)
}

const ARABIC_FALLBACK_MAX_LOOKUPS = 5

type arabic_fallback_plan_t struct {
	num_lookups  int
	free_lookups bool

	mask_array   [ARABIC_FALLBACK_MAX_LOOKUPS]Mask
	lookup_array [ARABIC_FALLBACK_MAX_LOOKUPS]*lookupGSUB
	accel_array  [ARABIC_FALLBACK_MAX_LOOKUPS]hb_ot_layout_lookup_accelerator_t
}

func (fbPlan *arabic_fallback_plan_t) initWin1256(plan *hb_ot_shape_plan_t, font *Font) bool {
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

func (fbPlan *arabic_fallback_plan_t) initUnicode(plan *hb_ot_shape_plan_t, font *Font) bool {
	var j int
	for i, feat := range arabicFallbackFeatures {
		fbPlan.mask_array[j] = plan.map_.get_1_mask(feat)
		if fbPlan.mask_array[j] != 0 {
			fbPlan.lookup_array[j] = arabicFallbackSynthesizeLookup(font, i)
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

func newArabicFallbackPlan(plan *hb_ot_shape_plan_t, font *Font) *arabic_fallback_plan_t {
	var fbPlan arabic_fallback_plan_t

	/* Try synthesizing GSUB table using Unicode Arabic Presentation Forms,
	* in case the font has cmap entries for the presentation-forms characters. */
	if fbPlan.initUnicode(plan, font) {
		return &fbPlan
	}

	/* See if this looks like a Windows-1256-encoded font. If it does, use a
	* hand-coded GSUB table. */
	if fbPlan.initWin1256(plan, font) {
		return &fbPlan
	}

	return &arabic_fallback_plan_t{}
}

func (fbPlan *arabic_fallback_plan_t) shape(font *Font, buffer *Buffer) {
	c := new_hb_ot_apply_context_t(0, font, buffer)
	for i := 0; i < fbPlan.num_lookups; i++ {
		if fbPlan.lookup_array[i] != nil {
			c.set_lookup_mask(fbPlan.mask_array[i])
			c.hb_ot_layout_substitute_lookup(*fbPlan.lookup_array[i], &fbPlan.accel_array[i])
		}
	}
}
