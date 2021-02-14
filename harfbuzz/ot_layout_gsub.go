package harfbuzz

import (
	"math/bits"

	"github.com/benoitkugler/textlayout/fonts"
	tt "github.com/benoitkugler/textlayout/fonts/truetype"
)

const HB_MAX_CONTEXT_LENGTH = 64

var _ TLookup = lookupGSUB{}

// implements TLookup
type lookupGSUB tt.LookupGSUB

func (l lookupGSUB) collectCoverage(dst *SetDigest) {
	for _, table := range l.Subtables {
		dst.collectCoverage(table.Coverage)
	}
}

func (l lookupGSUB) dispatchSubtables(ctx *hb_get_subtables_context_t) {
	for _, table := range l.Subtables {
		*ctx = append(*ctx, new_hb_applicable_t(table))
	}
}

func (l lookupGSUB) dispatchApply(ctx *hb_ot_apply_context_t) bool {
	for _, table := range l.Subtables {
		if gsubSubtable(table).apply(ctx) {
			return true
		}
	}
	return false
}

func (l lookupGSUB) isReverse() bool { return l.Type == tt.GSUBReverse }

func apply_recurse_GSUB(c *hb_ot_apply_context_t, lookupIndex uint16) bool {
	gsub, _ := c.face.get_gsubgpos_table()
	l := lookupGSUB(gsub.Lookups[lookupIndex])

	savedLookupProps := c.lookupProps
	savedLookupIndex := c.lookupIndex

	c.lookupIndex = lookupIndex
	c.set_lookup_props(l.Props())

	ret := l.dispatchApply(c)

	c.lookupIndex = savedLookupIndex
	c.set_lookup_props(savedLookupProps)
	return ret
}

//  implements `hb_apply_func_t`
type gsubSubtable tt.GSUBSubtable

// return `true` is the subsitution found a match and was applied
func (table gsubSubtable) apply(c *hb_ot_apply_context_t) bool {
	glyph := c.buffer.cur(0)
	glyphId := glyph.Glyph
	index, ok := table.Coverage.Index(glyphId)
	if !ok {
		return false
	}

	switch data := table.Data.(type) {
	case tt.GSUBSingle1:
		/* According to the Adobe Annotated OpenType Suite, result is always
		* limited to 16bit. */
		glyphId = fonts.GlyphIndex(int(glyphId) + int(data))
		c.replaceGlyph(glyphId)
	case tt.GSUBSingle2:
		if index >= len(data) { // index is not sanitized in tt.Parse
			return false
		}
		c.replaceGlyph(data[index])

	case tt.GSUBMultiple1:
		c.applySubsSequence(data[index])

	case tt.GSUBAlternate1:
		alternates := data[index]
		return c.applySubsAlternate(alternates)

	case tt.GSUBLigature1:
		ligatureSet := data[index]
		return c.applySubsLigature(ligatureSet)

	case tt.GSUBContext1:
		if index >= len(data) { // index is not sanitized in tt.Parse
			return false
		}
		ruleSet := data[index]
		return c.applyRuleSet(ruleSet, matchGlyph)
	case tt.GSUBContext2:
		class, _ := data.Class.ClassID(glyphId)
		ruleSet := data.SequenceSets[class]
		return c.applyRuleSet(ruleSet, matchClass(data.Class))
	case tt.GSUBContext3:
		covIndices := c.get1N(1, len(data.Coverages))
		return c.contextApplyLookup(covIndices, data.SequenceLookups, matchCoverage(data.Coverages))

	case tt.GSUBChainedContext1:
		if index >= len(data) { // index is not sanitized in tt.Parse
			return false
		}
		ruleSet := data[index]
		return c.applyChainRuleSet(ruleSet, [3]match_func_t{matchGlyph, matchGlyph, matchGlyph})
	case tt.GSUBChainedContext2:
		class, _ := data.InputClass.ClassID(glyphId)
		ruleSet := data.SequenceSets[class]
		return c.applyChainRuleSet(ruleSet, [3]match_func_t{
			matchClass(data.BacktrackClass), matchClass(data.InputClass), matchClass(data.LookaheadClass),
		})
	case tt.GSUBChainedContext3:
		lB, lI, lL := len(data.Backtrack), len(data.Input), len(data.Lookahead)
		return c.chainContextApplyLookup(c.get1N(0, lB), c.get1N(1, lI), c.get1N(0, lL),
			data.SequenceLookups, [3]match_func_t{
				matchCoverage(data.Backtrack), matchCoverage(data.Input), matchCoverage(data.Lookahead),
			})

	case tt.GSUBReverseChainedContext1:
		if c.nesting_level_left != HB_MAX_NESTING_LEVEL {
			return false // no chaining to this type
		}
		lB, lL := len(data.Backtrack), len(data.Lookahead)
		hasMatch, startIndex := c.matchBacktrack(c.get1N(0, lB), matchCoverage(data.Backtrack))
		if !hasMatch {
			return false
		}

		hasMatch, endIndex := c.matchLookahead(c.get1N(0, lL), matchCoverage(data.Lookahead), 1)
		if !hasMatch {
			return false
		}

		c.buffer.unsafeToBreakFromOutbuffer(startIndex, endIndex)
		c.setGlyphProps(data.Substitutes[index])
		c.buffer.cur(0).Glyph = data.Substitutes[index]
		/* Note: We DON'T decrease buffer.idx.  The main loop does it
		 * for us.  This is useful for preventing surprises if someone
		 * calls us through a Context lookup. */

	}

	return true
}

// return a slice containing [start, start+1, ..., end-1],
// using an internal buffer to avoid allocations
// these indices are used to refer to coverage
func (c *hb_ot_apply_context_t) get1N(start, end int) []uint16 {
	if end > cap(c.indices) {
		c.indices = make([]uint16, end)
		for i := range c.indices {
			c.indices[i] = uint16(i)
		}
	}
	return c.indices[start:end]
}

func (c *hb_ot_apply_context_t) applySubsSequence(seq []fonts.GlyphIndex) {
	/* Special-case to make it in-place and not consider this
	 * as a "multiplied" substitution. */
	switch len(seq) {
	case 1:
		c.replaceGlyph(seq[0])
	case 0:
		/* Spec disallows this, but Uniscribe allows it.
		 * https://github.com/harfbuzz/harfbuzz/issues/253 */
		c.buffer.deleteGlyph()
	default:
		var klass uint16
		if c.buffer.cur(0).isContinuation() {
			klass = tt.BaseGlyph
		}
		for i, g := range seq {
			c.buffer.cur(0).setLigPropsForMark(0, uint8(i))
			c.setGlyphPropsExt(g, klass, false, true)
			c.buffer.outputGlyphIndex(g)
		}
		c.buffer.skipGlyph()
	}
}

func (c *hb_ot_apply_context_t) applySubsAlternate(alternates []fonts.GlyphIndex) bool {
	count := uint32(len(alternates))
	if count == 0 {
		return false
	}

	glyphMask := c.buffer.cur(0).mask
	lookupMask := c.lookup_mask

	/* Note: This breaks badly if two features enabled this lookup together. */

	shift := bits.TrailingZeros32(lookupMask)
	altIndex := (lookupMask & glyphMask) >> shift

	/* If altIndex is MAX_VALUE, randomize feature if it is the rand feature. */
	if altIndex == otMapMaxValue && c.random {
		altIndex = c.randomNumber()%count + 1
	}

	if altIndex > count || altIndex == 0 {
		return false
	}

	c.replaceGlyph(alternates[altIndex-1])
	return true
}

func (c *hb_ot_apply_context_t) applySubsLigature(ligatureSet []tt.LigatureGlyph) bool {
	for _, lig := range ligatureSet {
		count := len(lig.Components) + 1

		/* Special-case to make it in-place and not consider this
		 * as a "ligated" substitution. */
		if count == 1 {
			c.replaceGlyph(lig.Glyph)
			return true
		}

		var matchPositions [HB_MAX_CONTEXT_LENGTH]int

		ok, matchLength, totalComponentCount := c.matchInput(lig.Components, matchGlyph, matchPositions)
		if !ok {
			continue
		}

		c.ligateInput(count, matchPositions, matchLength, lig.Glyph, totalComponentCount)

		return true
	}
	return false
}

func (c *hb_ot_apply_context_t) applyRuleSet(ruleSet []tt.SequenceRule, match match_func_t) bool {
	applied := false
	for _, rule := range ruleSet {
		b := c.contextApplyLookup(rule.Input, rule.Lookups, match)
		applied = applied || b
	}
	return applied
}

func (c *hb_ot_apply_context_t) applyChainRuleSet(ruleSet []tt.ChainedSequenceRule, match [3]match_func_t) bool {
	applied := false
	for _, rule := range ruleSet {
		b := c.chainContextApplyLookup(rule.Backtrack, rule.Input, rule.Lookahead, rule.Lookups, match)
		applied = applied || b
	}
	return applied
}