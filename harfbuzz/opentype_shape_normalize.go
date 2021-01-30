package harfbuzz

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
)

// ported from harfbuzz/src/hb-ot-shape-normalize.cc Copyright © 2011,2012  Google, Inc. Behdad Esfahbod

/*
 * HIGHLEVEL DESIGN:
 *
 * This file exports one main function: otShapeNormalize().
 *
 * This function closely reflects the Unicode Normalization Algorithm,
 * yet it's different.
 *
 * Each shaper specifies whether it prefers decomposed (NFD) or composed (NFC).
 * The logic however tries to use whatever the font can support.
 *
 * In general what happens is that: each grapheme is decomposed in a chain
 * of 1:2 decompositions, marks reordered, and then recomposed if desired,
 * so far it's like Unicode Normalization.  However, the decomposition and
 * recomposition only happens if the font supports the resulting characters.
 *
 * The goals are:
 *
 *   - Try to render all canonically equivalent strings similarly.  To really
 *     achieve this we have to always do the full decomposition and then
 *     selectively recompose from there.  It's kinda too expensive though, so
 *     we skip some cases.  For example, if composed is desired, we simply
 *     don't touch 1-character clusters that are supported by the font, even
 *     though their NFC may be different.
 *
 *   - When a font has a precomposed character for a sequence but the 'ccmp'
 *     feature in the font is not adequate, use the precomposed character
 *     which typically has better mark positioning.
 *
 *   - When a font does not support a combining mark, but supports it precomposed
 *     with previous base, use that.  This needs the itemizer to have this
 *     knowledge too.  We need to provide assistance to the itemizer.
 *
 *   - When a font does not support a character but supports its canonical
 *     decomposition, well, use the decomposition.
 *
 *   - The complex shapers can customize the compose and decompose functions to
 *     offload some of their requirements to the normalizer.  For example, the
 *     Indic shaper may want to disallow recomposing of two matras.
 */

const HB_OT_SHAPE_COMPLEX_MAX_COMBINING_MARKS = 32

type hb_ot_shape_normalization_mode_t uint8

const (
	HB_OT_SHAPE_NORMALIZATION_MODE_NONE hb_ot_shape_normalization_mode_t = iota
	HB_OT_SHAPE_NORMALIZATION_MODE_DECOMPOSED
	HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS                  // never composes base-to-base
	HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS_NO_SHORT_CIRCUIT // always fully decomposes and then recompose back

	HB_OT_SHAPE_NORMALIZATION_MODE_AUTO    // see below for logic.
	HB_OT_SHAPE_NORMALIZATION_MODE_DEFAULT = HB_OT_SHAPE_NORMALIZATION_MODE_AUTO
)

type hb_ot_shape_normalize_context_t struct {
	plan   *hb_ot_shape_plan_t
	buffer *hb_buffer_t
	font   *hb_font_t
	// hb_unicode_funcs_t *unicode;
	decompose func(c *hb_ot_shape_normalize_context_t, ab rune) (a, b rune, ok bool)
	compose   func(c *hb_ot_shape_normalize_context_t, a, b rune) (ab rune, ok bool)
}

//  static bool
//  decompose_unicode (c  *hb_ot_shape_normalize_context_t,
// 			rune  ab,
// 			rune *a,
// 			rune *b)
//  {
//    return (bool) c.unicode.decompose (ab, a, b);
//  }

//  static bool
//  compose_unicode (c  *hb_ot_shape_normalize_context_t,
// 		  rune  a,
// 		  rune  b,
// 		  rune *ab)
//  {
//    return (bool) c.unicode.compose (a, b, ab);
//  }

func setGlyph(info *hb_glyph_info_t, font *hb_font_t) {
	info.glyph_index, _ = font.face.GetNominalGlyph(info.codepoint)
}

func outputChar(buffer *hb_buffer_t, unichar rune, glyph fonts.GlyphIndex) {
	buffer.cur(0).glyph_index = glyph
	buffer.output_glyph(unichar) // this is very confusing indeed.
	buffer.prev().setUnicodeProps(buffer)
}

func (buffer *hb_buffer_t) nextChar(glyph fonts.GlyphIndex) {
	buffer.cur(0).glyph_index = glyph
	buffer.next_glyph()
}

// returns 0 if didn't decompose, number of resulting characters otherwise.
func decompose(c *hb_ot_shape_normalize_context_t, shortest bool, ab rune) int {
	var a_glyph, b_glyph fonts.GlyphIndex
	buffer := c.buffer
	font := c.font
	a, b, ok := c.decompose(c, ab)
	if !ok {
		b_glyph, ok = font.face.GetNominalGlyph(b)
		if b != 0 && !ok {
			return 0
		}
	}

	a_glyph, has_a := font.face.GetNominalGlyph(a)
	if shortest && has_a {
		/// output a and b
		outputChar(buffer, a, a_glyph)
		if b != 0 {
			outputChar(buffer, b, b_glyph)
			return 2
		}
		return 1
	}

	if ret := decompose(c, shortest, a); ret != 0 {
		if b != 0 {
			outputChar(buffer, b, b_glyph)
			return ret + 1
		}
		return ret
	}

	if has_a {
		outputChar(buffer, a, a_glyph)
		if b != 0 {
			outputChar(buffer, b, b_glyph)
			return 2
		}
		return 1
	}

	return 0
}

func (c *hb_ot_shape_normalize_context_t) decomposeCurrentCharacter(shortest bool) {
	buffer := c.buffer
	u := buffer.cur(0).codepoint
	glyph, ok := c.font.face.GetNominalGlyph(u)

	if shortest && ok {
		buffer.nextChar(glyph)
		return
	}

	if decompose(c, shortest, u) != 0 {
		buffer.skip_glyph()
		return
	}

	if !shortest && ok {
		buffer.nextChar(glyph)
		return
	}

	if buffer.cur(0).isUnicodeSpace() {
		//  rune space_glyph;
		spaceType := uni.space_fallback_type(u)
		if space_glyph, ok := c.font.face.GetNominalGlyph(0x0020); spaceType != NOT_SPACE && ok {
			buffer.cur(0).setUnicodeSpaceFallbackType(spaceType)
			buffer.nextChar(space_glyph)
			buffer.scratch_flags |= HB_BUFFER_SCRATCH_FLAG_HAS_SPACE_FALLBACK
			return
		}
	}

	if u == 0x2011 {
		/* U+2011 is the only sensible character that is a no-break version of another character
		 * and not a space. The space ones are handled already.  Handle this lone one. */
		if other_glyph, ok := c.font.face.GetNominalGlyph(0x2010); ok {
			buffer.nextChar(other_glyph)
			return
		}
	}

	buffer.nextChar(glyph)
}

func (c *hb_ot_shape_normalize_context_t) handleVariationSelectorCluster(end int) {
	buffer := c.buffer
	font := c.font
	for buffer.idx < end-1 {
		if uni.is_variation_selector(buffer.cur(+1).codepoint) {
			var ok bool
			buffer.cur(0).glyph_index, ok = font.face.GetVariationGlyph(buffer.cur(0).codepoint, buffer.cur(+1).codepoint)
			if ok {
				r := buffer.cur(0).codepoint
				buffer.replace_glyphs(2, []rune{r})
			} else {
				// Just pass on the two characters separately, let GSUB do its magic.
				setGlyph(buffer.cur(0), font)
				buffer.next_glyph()
				setGlyph(buffer.cur(0), font)
				buffer.next_glyph()
			}
			// skip any further variation selectors.
			for buffer.idx < end && uni.is_variation_selector(buffer.cur(0).codepoint) {
				setGlyph(buffer.cur(0), font)
				buffer.next_glyph()
			}
		} else {
			setGlyph(buffer.cur(0), font)
			buffer.next_glyph()
		}
	}
	if buffer.idx < end {
		setGlyph(buffer.cur(0), font)
		buffer.next_glyph()
	}
}

func (c *hb_ot_shape_normalize_context_t) decomposeMultiCharCluster(end int, shortCircuit bool) {
	buffer := c.buffer
	for i := buffer.idx; i < end; i++ {
		if uni.is_variation_selector(buffer.info[i].codepoint) {
			c.handleVariationSelectorCluster(end)
			return
		}
	}
	for buffer.idx < end {
		c.decomposeCurrentCharacter(shortCircuit)
	}
}

func compareCombiningClass(pa, pb *hb_glyph_info_t) int {
	a := pa.getModifiedCombiningClass()
	b := pb.getModifiedCombiningClass()
	if a < b {
		return -1
	} else if a == b {
		return 0
	}
	return 1
}

func otShapeNormalize(plan *hb_ot_shape_plan_t, buffer *hb_buffer_t, font *hb_font_t) {
	if len(buffer.info) == 0 {
		return
	}

	mode := plan.shaper.normalizationPreference()
	if mode == HB_OT_SHAPE_NORMALIZATION_MODE_AUTO {
		if plan.has_gpos_mark {
			// https://github.com/harfbuzz/harfbuzz/issues/653#issuecomment-423905920
			mode = HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS
		} else {
			mode = HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS
		}
	}

	c := hb_ot_shape_normalize_context_t{
		plan,
		buffer,
		font,
		plan.shaper.decompose,
		plan.shaper.compose,
	}

	alwaysShortCircuit := mode == HB_OT_SHAPE_NORMALIZATION_MODE_NONE
	mightShortCircuit := alwaysShortCircuit ||
		(mode != HB_OT_SHAPE_NORMALIZATION_MODE_DECOMPOSED &&
			mode != HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS_NO_SHORT_CIRCUIT)
		//    unsigned int count;

	/* We do a fairly straightforward yet custom normalization process in three
	* separate rounds: decompose, reorder, recompose (if desired).  Currently
	* this makes two buffer swaps.  We can make it faster by moving the last
	* two rounds into the inner loop for the first round, but it's more readable
	* this way. */

	/* First round, decompose */

	allSimple := true
	buffer.clear_output()
	count := len(buffer.info)
	buffer.idx = 0
	var end int
	for do := true; do; do = buffer.idx < end {
		for end = buffer.idx + 1; end < count; end++ {
			if buffer.info[end].isUnicodeMark() {
				break
			}
		}

		if end < count {
			end-- // leave one base for the marks to cluster with.
		}
		// from idx to end are simple clusters.
		if mightShortCircuit {
			var (
				i  int
				ok bool
			)
			for i = buffer.idx; i < end; i++ {
				buffer.info[i].glyph_index, ok = font.face.GetNominalGlyph(buffer.info[i].codepoint)
				if !ok {
					break
				}
			}
			buffer.next_glyphs(i - buffer.idx)
		}
		c.decomposeCurrentCharacter(mightShortCircuit)

		if buffer.idx == count {
			break
		}

		allSimple = false

		// find all the marks now.
		for end = buffer.idx + 1; end < count; end++ {
			if !buffer.info[end].isUnicodeMark() {
				break
			}
		}

		// idx to end is one non-simple cluster.
		c.decomposeMultiCharCluster(end, alwaysShortCircuit)
	}
	buffer.swap_buffers()

	/* Second round, reorder (inplace) */

	if !allSimple {
		if debugMode {
			fmt.Println("start reorder")
		}
		count = len(buffer.info)
		for i := 0; i < count; i++ {
			if buffer.info[i].getModifiedCombiningClass() == 0 {
				continue
			}

			var end int
			for end = i + 1; end < count; end++ {
				if buffer.info[end].getModifiedCombiningClass() == 0 {
					break
				}
			}

			// we are going to do a O(n^2).  Only do this if the sequence is short.
			if end-i > HB_OT_SHAPE_COMPLEX_MAX_COMBINING_MARKS {
				i = end
				continue
			}

			buffer.sort(i, end, compareCombiningClass)

			plan.shaper.reorder_marks(plan, buffer, i, end)

			i = end
		}
		if debugMode {
			fmt.Println("end reorder")
		}
	}

	if buffer.scratch_flags&HB_BUFFER_SCRATCH_FLAG_HAS_CGJ != 0 {
		/* For all CGJ, check if it prevented any reordering at all.
		 * If it did NOT, then make it skippable.
		 * https://github.com/harfbuzz/harfbuzz/issues/554 */
		for i := 1; i+1 < len(buffer.info); i++ {
			if buffer.info[i].codepoint == 0x034F /*CGJ*/ &&
				(buffer.info[i+1].getModifiedCombiningClass() == 0 || buffer.info[i-1].getModifiedCombiningClass() <= buffer.info[i+1].getModifiedCombiningClass()) {
				buffer.info[i].unhide()
			}
		}
	}

	/* Third round, recompose */

	if !allSimple &&
		(mode == HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS ||
			mode == HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS_NO_SHORT_CIRCUIT) {
		/* As noted in the comment earlier, we don't try to combine
		 * ccc=0 chars with their previous Starter. */

		buffer.clear_output()
		count = len(buffer.info)
		starter := 0
		buffer.next_glyph()
		for buffer.idx < count {
			//    rune composed, glyph;
			/* We don't try to compose a non-mark character with it's preceding starter.
			* This is both an optimization to avoid trying to compose every two neighboring
			* glyphs in most scripts AND a desired feature for Hangul.  Apparently Hangul
			* fonts are not designed to mix-and-match pre-composed syllables and Jamo. */
			if buffer.cur(0).isUnicodeMark() {
				/* If there's anything between the starter and this char, they should have CCC
				* smaller than this character's. */
				if starter == len(buffer.out_info)-1 ||
					buffer.prev().getModifiedCombiningClass() < buffer.cur(0).getModifiedCombiningClass() {
					/* And compose. */
					composed, ok := c.compose(&c, buffer.out_info[starter].codepoint, buffer.cur(0).codepoint)
					if ok {
						/* And the font has glyph for the composite. */
						glyph, ok := font.face.GetNominalGlyph(composed)
						if ok {
							/* Composes. */
							buffer.next_glyph() /* Copy to out-buffer. */
							buffer.merge_out_clusters(starter, len(buffer.out_info))
							buffer.out_info = buffer.out_info[:len(buffer.out_info)-1] // remove the second composable.
							/* Modify starter and carry on. */
							buffer.out_info[starter].codepoint = composed
							buffer.out_info[starter].glyph_index = glyph
							buffer.out_info[starter].setUnicodeProps(buffer)
						}
					}
					continue
				}
			}

			/* Blocked, or doesn't compose. */
			buffer.next_glyph()

			if buffer.prev().getModifiedCombiningClass() == 0 {
				starter = len(buffer.out_info) - 1
			}
		}
		buffer.swap_buffers()
	}
}
