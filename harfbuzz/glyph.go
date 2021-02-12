package harfbuzz

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// GlyphPosition holds the positions of the
// glyph in both horizontal and vertical directions.
// All positions are relative to the current point.
type GlyphPosition struct {
	// How much the line advances after drawing this glyph when setting
	// text in horizontal direction.
	XAdvance Position
	// How much the glyph moves on the X-axis before drawing it, this
	// should not affect how much the line advances.
	XOffset Position

	// How much the line advances after drawing this glyph when setting
	// text in vertical direction.
	YAdvance Position
	// How much the glyph moves on the Y-axis before drawing it, this
	// should not affect how much the line advances.
	YOffset Position
}

// unicodeProp is a two-byte number. The low byte includes:
// - General_Category: 5 bits
// - A bit each for:
//   -> Is it Default_Ignorable(); we have a modified Default_Ignorable().
//   -> Whether it's one of the three Mongolian Free Variation Selectors,
//     CGJ, or other characters that are hidden but should not be ignored
//     like most other Default_Ignorable()s do during matching.
//   -> Whether it's a grapheme continuation.
//
// The high-byte has different meanings, switched by the General_Category:
// - For Mn,Mc,Me: the modified Combining_Class.
// - For Cf: whether it's ZWJ, ZWNJ, or something else.
// - For Ws: index of which space character this is, if space fallback
//   is needed, ie. we don't set this by default, only if asked to.
type unicodeProp uint16

const (
	UPROPS_MASK_GEN_CAT   unicodeProp = 1<<5 - 1 // 11111
	UPROPS_MASK_IGNORABLE unicodeProp = 1 << (5 + iota)
	UPROPS_MASK_HIDDEN                // MONGOLIAN FREE VARIATION SELECTOR 1..3, or TAG characters
	UPROPS_MASK_CONTINUATION

	// if GEN_CAT=FORMAT, top byte masks
	UPROPS_MASK_Cf_ZWJ
	UPROPS_MASK_Cf_ZWNJ
)

// generalCategory extracts the general category.
func (prop unicodeProp) generalCategory() generalCategory {
	return generalCategory(prop & UPROPS_MASK_GEN_CAT)
}

// GlyphInfo holds information about the
// glyphs and their relation to input text.
// They are internally created from user input,
// and the shapping sets the `Glyph` field.
type GlyphInfo struct {
	// Cluster is the index of the character in the original text that corresponds
	// to this `GlyphInfo`, or whatever the client passes to
	// `Buffer.Add()`.
	// More than one glyph can have the same
	// `Cluster` value, if they resulted from the same character (e.g. one
	// to many glyph substitution), and when more than one character gets
	// merged in the same glyph (e.g. many to one glyph substitution) the
	// glyph will have the smallest Cluster value of them.
	// By default some characters are merged into the same Cluster
	// (e.g. combining marks have the same Cluster as their bases)
	// even if they are separate glyphs. See Buffer.ClusterLevel
	// for more fine-grained Cluster handling.
	Cluster int

	// Glyph is the result of the selection of concrete glyph
	// after shaping, and refers to the font used.
	Glyph fonts.GlyphIndex

	// input value of the shapping
	codepoint rune

	mask Mask

	// in C code: var1

	// GDEF glyph properties
	glyphProps uint16

	// GSUB/GPOS ligature tracking
	// When a ligature is formed:
	//
	//   - The ligature glyph and any marks in between all the same newly allocated
	//     lig_id,
	//   - The ligature glyph will get lig_num_comps set to the number of components
	//   - The marks get lig_comp > 0, reflecting which component of the ligature
	//     they were applied to.
	//   - This is used in GPOS to attach marks to the right component of a ligature
	//     in MarkLigPos,
	//   - Note that when marks are ligated together, much of the above is skipped
	//     and the current lig_id reused.
	//
	// When a multiple-substitution is done:
	//
	//   - All resulting glyphs will have lig_id = 0,
	//   - The resulting glyphs will have lig_comp = 0, 1, 2, ... respectively.
	//   - This is used in GPOS to attach marks to the first component of a
	//     multiple substitution in MarkBasePos.
	//
	// The numbers are also used in GPOS to do mark-to-mark positioning only
	// to marks that belong to the same component of the same ligature.
	ligProps uint8
	// GSUB/GPOS shaping boundaries
	syllable uint8

	// in C code: var2

	unicode unicodeProp

	complexCategory, complexAux uint8 // storage interpreted by complex shapers
}

func (info *GlyphInfo) setUnicodeProps(buffer *Buffer) {
	u := info.codepoint
	gen_cat := Uni.generalCategory(u)
	props := unicodeProp(gen_cat)

	if u >= 0x80 {
		buffer.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_NON_ASCII

		if Uni.isDefaultIgnorable(u) {
			buffer.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_DEFAULT_IGNORABLES
			props |= UPROPS_MASK_IGNORABLE
			if u == 0x200C {
				props |= UPROPS_MASK_Cf_ZWNJ
			} else if u == 0x200D {
				props |= UPROPS_MASK_Cf_ZWJ
			} else if 0x180B <= u && u <= 0x180D {
				/* Mongolian Free Variation Selectors need to be remembered
				 * because although we need to hide them like default-ignorables,
				 * they need to non-ignorable during shaping.  This is similar to
				 * what we do for joiners in Indic-like shapers, but since the
				 * FVSes are GC=Mn, we have use a separate bit to remember them.
				 * Fixes:
				 * https://github.com/harfbuzz/harfbuzz/issues/234 */
				props |= UPROPS_MASK_HIDDEN
			} else if 0xE0020 <= u && u <= 0xE007F {
				/* TAG characters need similar treatment. Fixes:
				 * https://github.com/harfbuzz/harfbuzz/issues/463 */
				props |= UPROPS_MASK_HIDDEN
			} else if u == 0x034F {
				/* COMBINING GRAPHEME JOINER should not be skipped; at least some times.
				 * https://github.com/harfbuzz/harfbuzz/issues/554 */
				buffer.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_CGJ
				props |= UPROPS_MASK_HIDDEN
			}
		}

		if gen_cat.isMark() {
			props |= UPROPS_MASK_CONTINUATION
			props |= unicodeProp(Uni.modified_combining_class(u)) << 8
		}
	}

	info.unicode = props
}

func (info *GlyphInfo) setGeneralCategory(genCat generalCategory) {
	/* Clears top-byte. */
	info.unicode = unicodeProp(genCat) | (info.unicode & (0xFF & ^UPROPS_MASK_GEN_CAT))
}

func (info *GlyphInfo) setCluster(cluster int, mask Mask) {
	if info.Cluster != cluster {
		if mask&HB_GLYPH_FLAG_UNSAFE_TO_BREAK != 0 {
			info.mask |= HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		} else {
			info.mask &= ^HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		}
	}
	info.Cluster = cluster
}

func (info *GlyphInfo) setContinuation() {
	info.unicode |= UPROPS_MASK_CONTINUATION
}

func (info *GlyphInfo) isContinuation() bool {
	return info.unicode&UPROPS_MASK_CONTINUATION != 0
}

func (info *GlyphInfo) resetContinutation() { info.unicode &= ^UPROPS_MASK_CONTINUATION }

func (info *GlyphInfo) IsUnicodeSpace() bool {
	return info.unicode.generalCategory() == SpaceSeparator
}

func (info *GlyphInfo) isUnicodeFormat() bool {
	return info.unicode.generalCategory() == Format
}

func (info *GlyphInfo) isZwnj() bool {
	return info.isUnicodeFormat() && (info.unicode&UPROPS_MASK_Cf_ZWNJ) != 0
}

func (info *GlyphInfo) isZwj() bool {
	return info.isUnicodeFormat() && (info.unicode&UPROPS_MASK_Cf_ZWJ) != 0
}

func (info *GlyphInfo) isJoiner() bool {
	return info.isUnicodeFormat() && (info.unicode&(UPROPS_MASK_Cf_ZWNJ|UPROPS_MASK_Cf_ZWJ)) != 0
}

func (info *GlyphInfo) isUnicodeMark() bool {
	return (info.unicode & UPROPS_MASK_GEN_CAT).generalCategory().isMark()
}

func (info *GlyphInfo) setUnicodeSpaceFallbackType(s uint8) {
	if !info.IsUnicodeSpace() {
		return
	}
	info.unicode = unicodeProp(s)<<8 | info.unicode&0xFF
}

func (info *GlyphInfo) getModifiedCombiningClass() uint8 {
	if info.isUnicodeMark() {
		return uint8(info.unicode >> 8)
	}
	return 0
}

func (info *GlyphInfo) unhide() {
	info.unicode &= ^UPROPS_MASK_HIDDEN
}

func (info *GlyphInfo) setModifiedCombiningClass(modifiedClass uint8) {
	if !info.isUnicodeMark() {
		return
	}
	info.unicode = (unicodeProp(modifiedClass) << 8) | (info.unicode & 0xFF)
}

func (info *GlyphInfo) Ligated() bool {
	return info.glyphProps&Ligated != 0
}

func (info *GlyphInfo) getLigId() uint8 {
	return info.ligProps >> 5
}

func (info *GlyphInfo) LigatedInternal() bool {
	return info.ligProps&isLigBase != 0
}

func (info *GlyphInfo) getLigComp() uint8 {
	if info.LigatedInternal() {
		return 0
	}
	return info.ligProps & 0x0F
}

func (info *GlyphInfo) getLigNumComps() uint8 {
	if (info.glyphProps&truetype.Ligature) != 0 && info.LigatedInternal() {
		return info.ligProps & 0x0F
	}
	return 1
}

func (info *GlyphInfo) setLigPropsForMark(ligId, ligComp uint8) {
	info.ligProps = (ligId << 5) | ligComp&0x0F
}

func (info *GlyphInfo) setLigPropsForLigature(ligId, ligNumComps uint8) {
	info.ligProps = (ligId << 5) | isLigBase | ligNumComps&0x0F
}

func (info *GlyphInfo) isDefaultIgnorable() bool {
	return (info.unicode&UPROPS_MASK_IGNORABLE) != 0 && !info.Ligated()
}

func (info *GlyphInfo) isDefaultIgnorableAndNotHidden() bool {
	return (info.unicode&(UPROPS_MASK_IGNORABLE|UPROPS_MASK_HIDDEN) == UPROPS_MASK_IGNORABLE) &&
		!info.Ligated()
}

func (info *GlyphInfo) getUnicodeSpaceFallbackType() uint8 {
	if info.IsUnicodeSpace() {
		return uint8(info.unicode >> 8)
	}
	return notSpace
}

func (info *GlyphInfo) isMark() bool {
	return info.glyphProps&truetype.Mark != 0
}

func (info *GlyphInfo) isBaseGlyph() bool {
	return info.glyphProps&truetype.BaseGlyph != 0
}

func (info *GlyphInfo) Multiplied() bool {
	return info.glyphProps&Multiplied != 0
}

func (info *GlyphInfo) ClearLigatedAndMultiplied() {
	info.glyphProps &= ^(Ligated | Multiplied)
}

func (info *GlyphInfo) LigatedAndDidntMultiply() bool {
	return info.Ligated() && !info.Multiplied()
}

func (info *GlyphInfo) Substituted() bool {
	return info.glyphProps&Substituted != 0
}

func (info *GlyphInfo) isLigature() bool {
	return info.glyphProps&truetype.Ligature != 0
}
