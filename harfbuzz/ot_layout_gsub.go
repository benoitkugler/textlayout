package harfbuzz

import (
	"math/bits"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

const HB_MAX_CONTEXT_LENGTH = 64

// implements TLookup
type lookupGSUB truetype.LookupGSUB

func (l lookupGSUB) collect_coverage(dst *SetDigest) {
	for _, table := range l.Subtables {
		collect_coverage(dst, table.Coverage)
	}
}

func (l lookupGSUB) dispatchSubtables(ctx *hb_get_subtables_context_t) {
	for _, table := range l.Subtables {
		// if (c.stop_sublookup_iteration (r))
		// return_trace (r);
		// }
		// return_trace (c.default_return_value ());
		*ctx = append(*ctx, new_hb_applicable_t(table))
	}
}

func (l lookupGSUB) dispatchApply(ctx *hb_ot_apply_context_t) {
	for _, table := range l.Subtables {
		table.apply(ctx)
	}
}

// returns is a 32-bit integer where the lower 16-bit is LookupFlag and
// higher 16-bit is mark-filtering-set if the lookup uses one.
// Not to be confused with glyph_props which is very similar.
func (l *lookupGSUB) get_props() uint32 {
	flag := uint32(l.Flag)
	if l.Flag&truetype.UseMarkFilteringSet != 0 {
		flag |= uint32(l.MarkFilteringSet) << 16
	}
	return flag
}

func apply_recurse_GSUB(c *hb_ot_apply_context_t, lookup_index uint16) bool {
	gsub, _ := c.face.get_gsubgpos_table()
	l := lookupGSUB(gsub.Lookups[lookup_index])
	savedLookupProps := c.lookup_props
	savedLookupIndex := c.lookup_index
	c.lookup_index = lookup_index
	c.set_lookup_props(l.get_props())
	ret := l.dispatchApply(c)
	c.lookup_index = savedLookupIndex
	c.set_lookup_props(savedLookupProps)
	return ret
}

//  implements `hb_apply_func_t`
type gsubSubtable truetype.LookupGSUBSubtable

func (table gsubSubtable) apply(c *hb_ot_apply_context_t) bool {
	glyph := c.buffer.cur(0)
	glyphId := glyph.Glyph
	index, ok := table.Coverage.Index(glyphId)
	if !ok {
		return false
	}

	switch data := table.Data.(type) {
	case truetype.SingleSubstitution1:
		/* According to the Adobe Annotated OpenType Suite, result is always
		* limited to 16bit. */
		glyphId = fonts.GlyphIndex(int(glyphId) + int(data))
		c.replaceGlyph(glyphId)
	case truetype.SingleSubstitution2:
		if index >= len(data) { // index is not sanitized in truetype.Parse
			return false
		}
		c.replaceGlyph(data[index])
	case truetype.SubstitutionMultiple:
		applySubsSequence(c, data[index])
	case truetype.SubstitutionAlternate:
		alternates := data[index]
		return applySubsAlternate(c, alternates)
	case truetype.SubstitutionLigature:
		ligatureSet := data[index]

	}

	return true
}

func applySubsSequence(c *hb_ot_apply_context_t, seq []fonts.GlyphIndex) {
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
			klass = truetype.BaseGlyph
		}
		for i, g := range seq {
			c.buffer.cur(0).setLigPropsForMark(0, i)
			c.setGlyphPropsExt(g, klass, false, true)
			c.buffer.outputGlyphIndex(g)
		}
		c.buffer.skipGlyph()
	}
}

func applySubsAlternate(c *hb_ot_apply_context_t, alternates []fonts.GlyphIndex) bool {
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

func applySubsLigature(c *hb_ot_apply_context_t, ligatureSet []truetype.LigatureGlyph) bool {
	for _, lig := range ligatureSet {
		count := len(lig.Components) + 1

		/* Special-case to make it in-place and not consider this
		 * as a "ligated" substitution. */
		if count == 1 {
			c.replaceGlyph(lig.Glyph)
			return true
		}

		var matchPositions [HB_MAX_CONTEXT_LENGTH]int

		ok, matchLength, totalComponentCount := c.matchInput(count, &component[1], match_glyph, nil, matchPositions)
		if !ok {
			continue
		}

		ligateInput(c, count, matchPositions,
			matchLength, ligGlyph, totalComponentCount)

		return true
	}
	return false
}
