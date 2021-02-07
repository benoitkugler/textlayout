package opentype

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts"
	cm "github.com/benoitkugler/textlayout/harfbuzz/common"
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

const shapeComplexMaxCombiningMarks = 32

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
	buffer *cm.Buffer
	font   *cm.Font
	// hb_unicode_funcs_t *unicode;
	decompose func(c *hb_ot_shape_normalize_context_t, ab rune) (a, b rune, ok bool)
	compose   func(c *hb_ot_shape_normalize_context_t, a, b rune) (ab rune, ok bool)
}

func setGlyph(info *cm.GlyphInfo, font *cm.Font) {
	info.GlyphIndex, _ = font.Face.GetNominalGlyph(info.Codepoint)
}

func outputChar(buffer *cm.Buffer, unichar rune, glyph fonts.GlyphIndex) {
	buffer.Cur(0).GlyphIndex = glyph
	buffer.OutputGlyph(unichar) // this is very confusing indeed.
	buffer.Prev().SetUnicodeProps(buffer)
}

func nextChar(buffer *cm.Buffer, glyph fonts.GlyphIndex) {
	buffer.Cur(0).GlyphIndex = glyph
	buffer.NextGlyph()
}

// returns 0 if didn't decompose, number of resulting characters otherwise.
func decompose(c *hb_ot_shape_normalize_context_t, shortest bool, ab rune) int {
	var a_glyph, b_glyph fonts.GlyphIndex
	buffer := c.buffer
	font := c.font
	a, b, ok := c.decompose(c, ab)
	if !ok {
		b_glyph, ok = font.Face.GetNominalGlyph(b)
		if b != 0 && !ok {
			return 0
		}
	}

	a_glyph, has_a := font.Face.GetNominalGlyph(a)
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
	u := buffer.Cur(0).Codepoint
	glyph, ok := c.font.Face.GetNominalGlyph(u)

	if shortest && ok {
		nextChar(buffer, glyph)
		return
	}

	if decompose(c, shortest, u) != 0 {
		buffer.SkipGlyph()
		return
	}

	if !shortest && ok {
		nextChar(buffer, glyph)
		return
	}

	if buffer.Cur(0).IsUnicodeSpace() {
		//  rune space_glyph;
		spaceType := cm.Uni.SpaceFallbackType(u)
		if space_glyph, ok := c.font.Face.GetNominalGlyph(0x0020); spaceType != cm.NOT_SPACE && ok {
			buffer.Cur(0).SetUnicodeSpaceFallbackType(spaceType)
			nextChar(buffer, space_glyph)
			buffer.ScratchFlags |= cm.HB_BUFFER_SCRATCH_FLAG_HAS_SPACE_FALLBACK
			return
		}
	}

	if u == 0x2011 {
		/* U+2011 is the only sensible character that is a no-break version of another character
		 * and not a space. The space ones are handled already.  Handle this lone one. */
		if other_glyph, ok := c.font.Face.GetNominalGlyph(0x2010); ok {
			nextChar(buffer, other_glyph)
			return
		}
	}

	nextChar(buffer, glyph)
}

func (c *hb_ot_shape_normalize_context_t) handleVariationSelectorCluster(end int) {
	buffer := c.buffer
	font := c.font
	for buffer.Idx < end-1 {
		if cm.Uni.IsVariationSelector(buffer.Cur(+1).Codepoint) {
			var ok bool
			buffer.Cur(0).GlyphIndex, ok = font.Face.GetVariationGlyph(buffer.Cur(0).Codepoint, buffer.Cur(+1).Codepoint)
			if ok {
				r := buffer.Cur(0).Codepoint
				buffer.ReplaceGlyphs(2, []rune{r})
			} else {
				// Just pass on the two characters separately, let GSUB do its magic.
				setGlyph(buffer.Cur(0), font)
				buffer.NextGlyph()
				setGlyph(buffer.Cur(0), font)
				buffer.NextGlyph()
			}
			// skip any further variation selectors.
			for buffer.Idx < end && cm.Uni.IsVariationSelector(buffer.Cur(0).Codepoint) {
				setGlyph(buffer.Cur(0), font)
				buffer.NextGlyph()
			}
		} else {
			setGlyph(buffer.Cur(0), font)
			buffer.NextGlyph()
		}
	}
	if buffer.Idx < end {
		setGlyph(buffer.Cur(0), font)
		buffer.NextGlyph()
	}
}

func (c *hb_ot_shape_normalize_context_t) decomposeMultiCharCluster(end int, shortCircuit bool) {
	buffer := c.buffer
	for i := buffer.Idx; i < end; i++ {
		if cm.Uni.IsVariationSelector(buffer.Info[i].Codepoint) {
			c.handleVariationSelectorCluster(end)
			return
		}
	}
	for buffer.Idx < end {
		c.decomposeCurrentCharacter(shortCircuit)
	}
}

func compareCombiningClass(pa, pb *cm.GlyphInfo) int {
	a := pa.GetModifiedCombiningClass()
	b := pb.GetModifiedCombiningClass()
	if a < b {
		return -1
	} else if a == b {
		return 0
	}
	return 1
}

func otShapeNormalize(plan *hb_ot_shape_plan_t, buffer *cm.Buffer, font *cm.Font) {
	if len(buffer.Info) == 0 {
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
	buffer.ClearOutput()
	count := len(buffer.Info)
	buffer.Idx = 0
	var end int
	for do := true; do; do = buffer.Idx < end {
		for end = buffer.Idx + 1; end < count; end++ {
			if buffer.Info[end].IsUnicodeMark() {
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
			for i = buffer.Idx; i < end; i++ {
				buffer.Info[i].GlyphIndex, ok = font.Face.GetNominalGlyph(buffer.Info[i].Codepoint)
				if !ok {
					break
				}
			}
			buffer.NextGlyphs(i - buffer.Idx)
		}
		c.decomposeCurrentCharacter(mightShortCircuit)

		if buffer.Idx == count {
			break
		}

		allSimple = false

		// find all the marks now.
		for end = buffer.Idx + 1; end < count; end++ {
			if !buffer.Info[end].IsUnicodeMark() {
				break
			}
		}

		// idx to end is one non-simple cluster.
		c.decomposeMultiCharCluster(end, alwaysShortCircuit)
	}
	buffer.SwapBuffers()

	/* Second round, reorder (inplace) */

	if !allSimple {
		if cm.DebugMode {
			fmt.Println("start reorder")
		}
		count = len(buffer.Info)
		for i := 0; i < count; i++ {
			if buffer.Info[i].GetModifiedCombiningClass() == 0 {
				continue
			}

			var end int
			for end = i + 1; end < count; end++ {
				if buffer.Info[end].GetModifiedCombiningClass() == 0 {
					break
				}
			}

			// we are going to do a O(n^2).  Only do this if the sequence is short.
			if end-i > shapeComplexMaxCombiningMarks {
				i = end
				continue
			}

			buffer.Sort(i, end, compareCombiningClass)

			plan.shaper.reorderMarks(plan, buffer, i, end)

			i = end
		}
		if cm.DebugMode {
			fmt.Println("end reorder")
		}
	}

	if buffer.ScratchFlags&cm.HB_BUFFER_SCRATCH_FLAG_HAS_CGJ != 0 {
		/* For all CGJ, check if it prevented any reordering at all.
		 * If it did NOT, then make it skippable.
		 * https://github.com/harfbuzz/harfbuzz/issues/554 */
		for i := 1; i+1 < len(buffer.Info); i++ {
			if buffer.Info[i].Codepoint == 0x034F /*CGJ*/ &&
				(buffer.Info[i+1].GetModifiedCombiningClass() == 0 || buffer.Info[i-1].GetModifiedCombiningClass() <= buffer.Info[i+1].GetModifiedCombiningClass()) {
				buffer.Info[i].Unhide()
			}
		}
	}

	/* Third round, recompose */

	if !allSimple &&
		(mode == HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS ||
			mode == HB_OT_SHAPE_NORMALIZATION_MODE_COMPOSED_DIACRITICS_NO_SHORT_CIRCUIT) {
		/* As noted in the comment earlier, we don't try to combine
		 * ccc=0 chars with their previous Starter. */

		buffer.ClearOutput()
		count = len(buffer.Info)
		starter := 0
		buffer.NextGlyph()
		for buffer.Idx < count {
			//    rune composed, glyph;
			/* We don't try to compose a non-mark character with it's preceding starter.
			* This is both an optimization to avoid trying to compose every two neighboring
			* glyphs in most scripts AND a desired feature for Hangul.  Apparently Hangul
			* fonts are not designed to mix-and-match pre-composed syllables and Jamo. */
			if buffer.Cur(0).IsUnicodeMark() {
				/* If there's anything between the starter and this char, they should have CCC
				* smaller than this character's. */
				if starter == len(buffer.OutInfo)-1 ||
					buffer.Prev().GetModifiedCombiningClass() < buffer.Cur(0).GetModifiedCombiningClass() {
					/* And compose. */
					composed, ok := c.compose(&c, buffer.OutInfo[starter].Codepoint, buffer.Cur(0).Codepoint)
					if ok {
						/* And the font has glyph for the composite. */
						glyph, ok := font.Face.GetNominalGlyph(composed)
						if ok {
							/* Composes. */
							buffer.NextGlyph() /* Copy to out-buffer. */
							buffer.MergeOutClusters(starter, len(buffer.OutInfo))
							buffer.OutInfo = buffer.OutInfo[:len(buffer.OutInfo)-1] // remove the second composable.
							/* Modify starter and carry on. */
							buffer.OutInfo[starter].Codepoint = composed
							buffer.OutInfo[starter].GlyphIndex = glyph
							buffer.OutInfo[starter].SetUnicodeProps(buffer)
						}
					}
					continue
				}
			}

			/* Blocked, or doesn't compose. */
			buffer.NextGlyph()

			if buffer.Prev().GetModifiedCombiningClass() == 0 {
				starter = len(buffer.OutInfo) - 1
			}
		}
		buffer.SwapBuffers()
	}
}
