package harfbuzz

import (
	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

// ported from harfbuzz/src/hb-ot-shape-complex-hangul.cc Copyright © 2013  Google, Inc. Behdad Esfahbod

var _ hb_ot_complex_shaper_t = (*complexShaperHangul)(nil)

// Hangul shaper
type complexShaperHangul struct {
	complexShaperNil

	plan hangulShapePlan
}

/* Same order as the feature array below */
const (
	_JMO = iota

	LJMO
	VJMO
	TJMO

	FIRST_HANGUL_FEATURE = LJMO
	HANGUL_FEATURE_COUNT = TJMO + 1
)

var hangul_features = [HANGUL_FEATURE_COUNT]hb_tag_t{
	0,
	newTag('l', 'j', 'm', 'o'),
	newTag('v', 'j', 'm', 'o'),
	newTag('t', 'j', 'm', 'o'),
}

func (complexShaperHangul) collectFeatures(plan *hb_ot_shape_planner_t) {
	map_ := &plan.map_

	for i := FIRST_HANGUL_FEATURE; i < HANGUL_FEATURE_COUNT; i++ {
		map_.add_feature(hangul_features[i])
	}
}

func (complexShaperHangul) overrideFeatures(plan *hb_ot_shape_planner_t) {
	/* Uniscribe does not apply 'calt' for Hangul, and certain fonts
	* (Noto Sans CJK, Source Sans Han, etc) apply all of jamo lookups
	* in calt, which is not desirable. */
	plan.map_.disable_feature(newTag('c', 'a', 'l', 't'))
}

type hangulShapePlan struct {
	mask_array [HANGUL_FEATURE_COUNT]Mask
}

func (cs *complexShaperHangul) dataCreate(plan *hb_ot_shape_plan_t) {
	var hangulPlan hangulShapePlan

	for i := range hangulPlan.mask_array {
		hangulPlan.mask_array[i] = plan.map_.get_1_mask(hangul_features[i])
	}

	cs.plan = hangulPlan
}

func isCombiningT(u rune) bool {
	return ucd.HangulTBase+1 <= u && u <= ucd.HangulTBase+ucd.HangulTCount-1
}

func isL(u rune) bool {
	return 0x1100 <= u && u <= 0x115F || 0xA960 <= u && u <= 0xA97C
}

func isV(u rune) bool {
	return 0x1160 <= u && u <= 0x11A7 || 0xD7B0 <= u && u <= 0xD7C6
}

func isT(u rune) bool {
	return 0x11A8 <= u && u <= 0x11FF || 0xD7CB <= u && u <= 0xD7FB
}

//  /* buffer var allocations */
//  #define complexAux complex_var_u8_auxiliary() /* hangul jamo shaping feature */

func isZeroWidthChar(font *Font, unicode rune) bool {
	glyph, ok := font.Face.GetNominalGlyph(unicode)
	return ok && font.GetGlyphHAdvance(glyph) == 0
}

func (cs *complexShaperHangul) preprocessText(_ *hb_ot_shape_plan_t, buffer *Buffer, font *Font) {
	/* Hangul syllables come in two shapes: LV, and LVT.  Of those:
	*
	*   - LV can be precomposed, or decomposed.  Lets call those
	*     <LV> and <L,V>,
	*   - LVT can be fully precomposed, partically precomposed, or
	*     fully decomposed.  Ie. <LVT>, <LV,T>, or <L,V,T>.
	*
	* The composition / decomposition is mechanical.  However, not
	* all <L,V> sequences compose, and not all <LV,T> sequences
	* compose.
	*
	* Here are the specifics:
	*
	*   - <L>: U+1100..115F, U+A960..A97F
	*   - <V>: U+1160..11A7, U+D7B0..D7C7
	*   - <T>: U+11A8..11FF, U+D7CB..D7FB
	*
	*   - Only the <L,V> sequences for some of the U+11xx ranges combine.
	*   - Only <LV,T> sequences for some of the Ts in U+11xx range combine.
	*
	* Here is what we want to accomplish in this shaper:
	*
	*   - If the whole syllable can be precomposed, do that,
	*   - Otherwise, fully decompose and apply ljmo/vjmo/tjmo features.
	*   - If a valid syllable is followed by a Hangul tone mark, reorder the tone
	*     mark to precede the whole syllable - unless it is a zero-width glyph, in
	*     which case we leave it untouched, assuming it's designed to overstrike.
	*
	* That is, of the different possible syllables:
	*
	*   <L>
	*   <L,V>
	*   <L,V,T>
	*   <LV>
	*   <LVT>
	*   <LV, T>
	*
	* - <L> needs no work.
	*
	* - <LV> and <LVT> can stay the way they are if the font supports them, otherwise we
	*   should fully decompose them if font supports.
	*
	* - <L,V> and <L,V,T> we should compose if the whole thing can be composed.
	*
	* - <LV,T> we should compose if the whole thing can be composed, otherwise we should
	*   decompose.
	 */

	buffer.clearOutput()
	// Extent of most recently seen syllable; valid only if start < end
	var start, end int
	count := len(buffer.Info)

	for buffer.idx = 0; buffer.idx < count; {
		u := buffer.cur(0).codepoint

		if 0x302E <= u && u <= 0x302F { // isHangulTone
			/*
			* We could cache the width of the tone marks and the existence of dotted-circle,
			* but the use of the Hangul tone mark characters seems to be rare enough that
			* I didn't bother for now.
			 */
			if start < end && end == len(buffer.outInfo) {
				/* Tone mark follows a valid syllable; move it in front, unless it's zero width. */
				buffer.unsafeToBreakFromOutbuffer(start, buffer.idx)
				buffer.nextGlyph()
				if !isZeroWidthChar(font, u) {
					buffer.mergeOutClusters(start, end+1)
					info := buffer.outInfo
					tone := info[end]
					copy(info[start+1:], info[start:end])
					info[start] = tone
				}
			} else {
				/* No valid syllable as base for tone mark; try to insert dotted circle. */
				if buffer.Flags&DoNotinsertDottedCircle == 0 && font.HasGlyph(0x25CC) {
					var chars [2]rune
					if !isZeroWidthChar(font, u) {
						chars[0] = u
						chars[1] = 0x25CC
					} else {
						chars[0] = 0x25CC
						chars[1] = u
					}
					buffer.replaceGlyphs(1, chars[:])
				} else {
					/* No dotted circle available in the font; just leave tone mark untouched. */
					buffer.nextGlyph()
				}
			}
			start = len(buffer.outInfo)
			end = len(buffer.outInfo)
			continue
		}

		start = len(buffer.outInfo) /* Remember current position as a potential syllable start;
		 * will only be used if we set end to a later position.
		 */

		if isL(u) && buffer.idx+1 < count {
			l := u
			v := buffer.cur(+1).codepoint
			if isV(v) {
				/* Have <L,V> or <L,V,T>. */
				var t, tindex rune
				if buffer.idx+2 < count {
					t = buffer.cur(+2).codepoint
					if isT(t) {
						tindex = t - ucd.HangulTBase /* Only used if isCombiningT (t); otherwise invalid. */
					} else {
						t = 0 /* The next character was not a trailing jamo. */
					}
				}
				offset := 2
				if t != 0 {
					offset = 3
				}
				buffer.unsafeToBreak(buffer.idx, buffer.idx+offset)

				/* We've got a syllable <L,V,T?>; see if it can potentially be composed. */
				if (ucd.HangulLBase <= l && l <= ucd.HangulLBase+ucd.HangulLCount-1) && (ucd.HangulVBase <= v && v <= ucd.HangulVBase+ucd.HangulVCount-1) && (t == 0 || isCombiningT(t)) {
					/* Try to compose; if this succeeds, end is set to start+1. */
					s := ucd.HangulSBase + (l-ucd.HangulLBase)*ucd.HangulNCount + (v-ucd.HangulVBase)*ucd.HangulTCount + tindex
					if font.HasGlyph(s) {
						buffer.replaceGlyphs(offset, []rune{s})
						end = start + 1
						continue
					}
				}

				/* We didn't compose, either because it's an Old Hangul syllable without a
				 * precomposed character in Unicode, or because the font didn't support the
				 * necessary precomposed glyph.
				 * Set jamo features on the individual glyphs, and advance past them.
				 */
				buffer.cur(0).complexAux = LJMO
				buffer.nextGlyph()
				buffer.cur(0).complexAux = VJMO
				buffer.nextGlyph()
				if t != 0 {
					buffer.cur(0).complexAux = TJMO
					buffer.nextGlyph()
					end = start + 3
				} else {
					end = start + 2
				}
				if buffer.ClusterLevel == MonotoneGraphemes {
					buffer.mergeOutClusters(start, end)
				}
				continue
			}
		} else if ucd.HangulSBase <= u && u <= ucd.HangulSBase+ucd.HangulSCount-1 { // is combined S
			/* Have <LV>, <LVT>, or <LV,T> */
			s := u
			HasGlyph := font.HasGlyph(s)
			lindex := (s - ucd.HangulSBase) / ucd.HangulNCount
			nindex := (s - ucd.HangulSBase) % ucd.HangulNCount
			vindex := nindex / ucd.HangulTCount
			tindex := nindex % ucd.HangulTCount

			if tindex == 0 && buffer.idx+1 < count && isCombiningT(buffer.cur(+1).codepoint) {
				/* <LV,T>, try to combine. */
				newTindex := buffer.cur(+1).codepoint - ucd.HangulTBase
				newS := s + newTindex
				if font.HasGlyph(newS) {
					buffer.replaceGlyphs(2, []rune{newS})
					end = start + 1
					continue
				} else {
					buffer.unsafeToBreak(buffer.idx, buffer.idx+2) /* Mark unsafe between LV and T. */
				}
			}

			/* Otherwise, decompose if font doesn't support <LV> or <LVT>,
			* or if having non-combining <LV,T>.  Note that we already handled
			* combining <LV,T> above. */
			if !HasGlyph || (tindex == 0 && buffer.idx+1 < count && isT(buffer.cur(+1).codepoint)) {
				decomposed := [3]rune{
					ucd.HangulLBase + lindex,
					ucd.HangulVBase + vindex,
					ucd.HangulTBase + tindex,
				}
				if font.HasGlyph(decomposed[0]) && font.HasGlyph(decomposed[1]) &&
					(tindex == 0 || font.HasGlyph(decomposed[2])) {
					sLen := 2
					if tindex != 0 {
						sLen = 3
					}
					buffer.replaceGlyphs(1, decomposed[:sLen])

					/* If we decomposed an LV because of a non-combining T following,
					* we want to include this T in the syllable.
					 */
					if HasGlyph && tindex == 0 {
						buffer.nextGlyph()
						sLen++
					}

					/* We decomposed S: apply jamo features to the individual glyphs
					* that are now in buffer.OutInfo.
					 */
					info := buffer.outInfo
					end = start + sLen

					i := start
					info[i].complexAux = LJMO
					i++
					info[i].complexAux = VJMO
					i++
					if i < end {
						info[i].complexAux = TJMO
						i++
					}

					if buffer.ClusterLevel == MonotoneGraphemes {
						buffer.mergeOutClusters(start, end)
					}
					continue
				} else if tindex == 0 && buffer.idx+1 < count && isT(buffer.cur(+1).codepoint) {
					buffer.unsafeToBreak(buffer.idx, buffer.idx+2) /* Mark unsafe between LV and T. */
				}
			}

			if HasGlyph {
				/* We didn't decompose the S, so just advance past it. */
				end = start + 1
				buffer.nextGlyph()
				continue
			}
		}

		/* Didn't find a recognizable syllable, so we leave end <= start;
		 * this will prevent tone-mark reordering happening.
		 */
		buffer.nextGlyph()
	}
	buffer.swapBuffers()
}

func (cs *complexShaperHangul) setupMasks(_ *hb_ot_shape_plan_t, buffer *Buffer, _ *Font) {
	hangulPlan := cs.plan

	info := buffer.Info
	for i := range info {
		info[i].mask |= hangulPlan.mask_array[info[i].complexAux]
	}
}

func (complexShaperHangul) marksBehavior() (hb_ot_shape_zero_width_marks_type_t, bool) {
	return HB_OT_SHAPE_ZERO_WIDTH_MARKS_NONE, false
}

func (complexShaperHangul) normalizationPreference() hb_ot_shape_normalization_mode_t {
	return HB_OT_SHAPE_NORMALIZATION_MODE_NONE
}

func (complexShaperHangul) gposTag() hb_tag_t { return 0 }