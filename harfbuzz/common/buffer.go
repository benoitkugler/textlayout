package common

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

/* ported from harfbuzz/src/hb-buffer.hh and hb-buffer.h
 * Copyright © 1998-2004  David Turner and Werner Lemberg
 * Copyright © 2004,2007,2009,2010  Red Hat, Inc.
 * Copyright © 2011,2012  Google, Inc.
 * Red Hat Author(s): Owen Taylor, Behdad Esfahbod
 * Google Author(s): Behdad Esfahbod */

//  #ifndef HB_BUFFER_MAX_LEN_FACTOR
//  #define HB_BUFFER_MAX_LEN_FACTOR 64
//  #endif
//  #ifndef HB_BUFFER_MAX_LEN_MIN
//  #define HB_BUFFER_MAX_LEN_MIN 16384
//  #endif
//  #ifndef HB_BUFFER_MAX_LEN_DEFAULT
//  #define HB_BUFFER_MAX_LEN_DEFAULT 0x3FFFFFFF /* Shaping more than a billion chars? Let us know! */
//  #endif

//  #ifndef HB_BUFFER_MAX_OPS_FACTOR
//  #define HB_BUFFER_MAX_OPS_FACTOR 1024
//  #endif
//  #ifndef HB_BUFFER_MAX_OPS_MIN
//  #define HB_BUFFER_MAX_OPS_MIN 16384
//  #endif
//  #ifndef HB_BUFFER_MAX_OPS_DEFAULT
//  #define HB_BUFFER_MAX_OPS_DEFAULT 0x1FFFFFFF /* Shaping more than a billion operations? Let us know! */
//  #endif

//  static_assert ((sizeof (GlyphInfo) == 20), "");
//  static_assert ((sizeof (GlyphInfo) == sizeof (hb_glyph_position_t)), "");

type Mask = uint32

const (
	// Indicates that if input text is broken at the
	// beginning of the cluster this glyph is part of,
	// then both sides need to be re-shaped, as the
	// result might be different.  On the flip side,
	// it means that when this flag is not present,
	// then it's safe to break the glyph-run at the
	// beginning of this cluster, and the two sides
	// represent the exact same result one would get
	// if breaking input text at the beginning of
	// this cluster and shaping the two sides
	// separately.  This can be used to optimize
	// paragraph layout, by avoiding re-shaping
	// of each line after line-breaking, or limiting
	// the reshaping to a small piece around the
	// breaking point only.
	HB_GLYPH_FLAG_UNSAFE_TO_BREAK Mask = 0x00000001

	// OR of all defined flags
	HB_GLYPH_FLAG_DEFINED Mask = HB_GLYPH_FLAG_UNSAFE_TO_BREAK
)

const (
	// The following are used internally; not derived from GDEF.
	Substituted truetype.GlyphProps = 1 << (iota + 4)
	Ligated
	Multiplied

	Preserve = Substituted | Ligated | Multiplied
)

const IS_LIG_BASE = 0x10

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

// GeneralCategory extracts the general category.
func (prop unicodeProp) GeneralCategory() GeneralCategory {
	return GeneralCategory(prop & UPROPS_MASK_GEN_CAT)
}

// GlyphInfo holds information about the
// glyphs and their relation to input text.
type GlyphInfo struct {
	// either a Unicode code point (before shaping) or a glyph index
	// (after shaping).
	Codepoint rune

	// the index of the character in the original text that corresponds
	// to this #GlyphInfo, or whatever the client passes to
	// hb_buffer_add(). More than one #GlyphInfo can have the same
	// `cluster` value, if they resulted from the same character (e.g. one
	// to many glyph substitution), and when more than one character gets
	// merged in the same glyph (e.g. many to one glyph substitution) the
	// #GlyphInfo will have the smallest cluster value of them.
	// By default some characters are merged into the same cluster
	// (e.g. combining marks have the same cluster as their bases)
	// even if they are separate glyphs, hb_buffer_set_cluster_level()
	// allow selecting more fine-grained cluster handling.
	cluster int

	Mask Mask

	Unicode     unicodeProp
	glyph_index fonts.GlyphIndex

	// TODO: first uint16 of glyph_index in the C code
	GlyphProps uint16

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
	// TODO: third byte of glyph_index in the C code
	lig_props uint8

	// GSUB/GPOS shaping boundaries
	// TODO: fourth byte of glyph_index in the C code
	// also used as auxiliary storage by complex shapers
	Aux uint8
}

func (info *GlyphInfo) setUnicodeProps(buffer *cm.Buffer) {
	u := info.Codepoint
	gen_cat := uni.GeneralCategory(u)
	props := unicodeProp(gen_cat)

	if u >= 0x80 {
		buffer.scratch_flags |= HB_BUFFER_SCRATCH_FLAG_HAS_NON_ASCII

		if uni.is_default_ignorable(u) {
			buffer.scratch_flags |= HB_BUFFER_SCRATCH_FLAG_HAS_DEFAULT_IGNORABLES
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
				buffer.scratch_flags |= HB_BUFFER_SCRATCH_FLAG_HAS_CGJ
				props |= UPROPS_MASK_HIDDEN
			}
		}

		if gen_cat.isMark() {
			props |= UPROPS_MASK_CONTINUATION
			props |= unicodeProp(uni.modified_combining_class(u)) << 8
		}
	}

	info.Unicode = props
}

func (info *GlyphInfo) set_cluster(cluster int, mask Mask) {
	if info.cluster != cluster {
		if mask&HB_GLYPH_FLAG_UNSAFE_TO_BREAK != 0 {
			info.Mask |= HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		} else {
			info.Mask &= ^HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		}
	}
	info.cluster = cluster
}

func (info *GlyphInfo) setContinuation() {
	info.Unicode |= UPROPS_MASK_CONTINUATION
}
func (info *GlyphInfo) isContinuation() bool {
	return info.Unicode&UPROPS_MASK_CONTINUATION != 0
}

func (info *GlyphInfo) isUnicodeSpace() bool {
	return info.Unicode.GeneralCategory() == spaceSeparator
}

func (info *GlyphInfo) isUnicodeFormat() bool {
	return info.Unicode.GeneralCategory() == format
}

func (info *GlyphInfo) isZwnj() bool {
	return info.isUnicodeFormat() && (info.Unicode&UPROPS_MASK_Cf_ZWNJ) != 0
}

func (info *GlyphInfo) isZwj() bool {
	return info.isUnicodeFormat() && (info.Unicode&UPROPS_MASK_Cf_ZWJ) != 0
}

func (info *GlyphInfo) isJoiner() bool {
	return info.isUnicodeFormat() && (info.Unicode&(UPROPS_MASK_Cf_ZWNJ|UPROPS_MASK_Cf_ZWJ)) != 0
}

func (info *GlyphInfo) isUnicodeMark() bool {
	return (info.Unicode & UPROPS_MASK_GEN_CAT).GeneralCategory().isMark()
}

func (info *GlyphInfo) setUnicodeSpaceFallbackType(s uint8) {
	if !info.isUnicodeSpace() {
		return
	}
	info.Unicode = unicodeProp(s)<<8 | info.Unicode&0xFF
}

func (info *GlyphInfo) GetModifiedCombiningClass() uint8 {
	if info.isUnicodeMark() {
		return uint8(info.Unicode >> 8)
	}
	return 0
}

func (info *GlyphInfo) unhide() {
	info.Unicode &= ^UPROPS_MASK_HIDDEN
}

func (info *GlyphInfo) SetModifiedCombiningClass(modifiedClass uint8) {
	if !info.isUnicodeMark() {
		return
	}
	info.Unicode = (unicodeProp(modifiedClass) << 8) | (info.Unicode & 0xFF)
}

func (info *GlyphInfo) isLigated() bool {
	return info.GlyphProps&Ligated != 0
}

func (info *GlyphInfo) getLigId() uint8 {
	return info.lig_props >> 5
}

func (info *GlyphInfo) isLigatedInternal() bool {
	return info.lig_props&IS_LIG_BASE != 0
}

func (info *GlyphInfo) GetLigComp() uint8 {
	if info.isLigatedInternal() {
		return 0
	}
	return info.lig_props & 0x0F
}

func (info *GlyphInfo) getLigNumComps() uint8 {
	if (info.GlyphProps&truetype.Ligature) != 0 && info.isLigatedInternal() {
		return info.lig_props & 0x0F
	}
	return 1
}

func (info *GlyphInfo) IsDefaultIgnorable() bool {
	return (info.Unicode&UPROPS_MASK_IGNORABLE) != 0 && !info.isLigated()
}

func (info *GlyphInfo) getUnicodeSpaceFallbackType() uint8 {
	if info.isUnicodeSpace() {
		return uint8(info.Unicode >> 8)
	}
	return NOT_SPACE
}

func (info *GlyphInfo) isMark() bool {
	return info.GlyphProps&truetype.Mark != 0
}

func (info *GlyphInfo) Multiplied() bool {
	return info.GlyphProps&Multiplied != 0
}

// The hb_glyph_position_t is the structure that holds the positions of the
// glyph in both horizontal and vertical directions.
// All positions are relative to the current point.
type hb_glyph_position_t struct {
	// how much the line advances after drawing this glyph when setting
	// text in horizontal direction.
	XAdvance Position
	// how much the line advances after drawing this glyph when setting
	// text in vertical direction.
	y_advance Position
	// how much the glyph moves on the X-axis before drawing it, this
	// should not affect how much the line advances.
	XOffset Position
	// how much the glyph moves on the Y-axis before drawing it, this
	// should not affect how much the line advances.
	y_offset Position
}

type hb_buffer_flags_t uint16

const (
	// flag indicating that special handling of the beginning
	// of text paragraph can be applied to this buffer. Should usually
	// be set, unless you are passing to the buffer only part
	// of the text without the full context.
	HB_BUFFER_FLAG_BOT hb_buffer_flags_t = 1 << iota /* Beginning-of-text */
	// flag indicating that special handling of the end of text
	// paragraph can be applied to this buffer, similar to
	// @HB_BUFFER_FLAG_BOT.
	HB_BUFFER_FLAG_EOT
	// flag indication that character with Default_Ignorable
	// Unicode property should use the corresponding glyph
	// from the font, instead of hiding them (done by
	// replacing them with the space glyph and zeroing the
	// advance width.)  This flag takes precedence over
	// @HB_BUFFER_FLAG_REMOVE_DEFAULT_IGNORABLES.
	HB_BUFFER_FLAG_PRESERVE_DEFAULT_IGNORABLES
	// flag indication that character with Default_Ignorable
	// Unicode property should be removed from glyph string
	// instead of hiding them (done by replacing them with the
	// space glyph and zeroing the advance width.)
	// @HB_BUFFER_FLAG_PRESERVE_DEFAULT_IGNORABLES takes
	// precedence over this flag.
	HB_BUFFER_FLAG_REMOVE_DEFAULT_IGNORABLES
	// flag indicating that a dotted circle should
	// not be inserted in the rendering of incorrect
	// character sequences (such at <0905 093E>).
	HB_BUFFER_FLAG_DO_NOT_INSERT_DOTTED_CIRCLE

	HB_BUFFER_FLAG_DEFAULT hb_buffer_flags_t = 0
)

type hb_buffer_cluster_level_t uint8

const (
	// Return cluster values grouped by graphemes into monotone order.
	HB_BUFFER_CLUSTER_LEVEL_MONOTONE_GRAPHEMES hb_buffer_cluster_level_t = iota
	//  Return cluster values grouped into monotone order.
	HB_BUFFER_CLUSTER_LEVEL_MONOTONE_CHARACTERS
	// Don't group cluster values.
	HB_BUFFER_CLUSTER_LEVEL_CHARACTERS
	// Default cluster level
	HB_BUFFER_CLUSTER_LEVEL_DEFAULT = HB_BUFFER_CLUSTER_LEVEL_MONOTONE_GRAPHEMES
)

type hb_buffer_scratch_flags_t uint32

const (
	HB_BUFFER_SCRATCH_FLAG_DEFAULT       hb_buffer_scratch_flags_t = 0x00000000
	HB_BUFFER_SCRATCH_FLAG_HAS_NON_ASCII hb_buffer_scratch_flags_t = 1 << iota
	HB_BUFFER_SCRATCH_FLAG_HAS_DEFAULT_IGNORABLES
	HB_BUFFER_SCRATCH_FLAG_HAS_SPACE_FALLBACK
	HB_BUFFER_SCRATCH_FLAG_HAS_GPOS_ATTACHMENT
	HB_BUFFER_SCRATCH_FLAG_HAS_UNSAFE_TO_BREAK
	HB_BUFFER_SCRATCH_FLAG_HAS_CGJ

	/* Reserved for complex shapers' internal use. */
	HB_BUFFER_SCRATCH_FLAG_COMPLEX0 hb_buffer_scratch_flags_t = 0x01000000
	HB_BUFFER_SCRATCH_FLAG_COMPLEX1 hb_buffer_scratch_flags_t = 0x02000000
	HB_BUFFER_SCRATCH_FLAG_COMPLEX2 hb_buffer_scratch_flags_t = 0x04000000
	HB_BUFFER_SCRATCH_FLAG_COMPLEX3 hb_buffer_scratch_flags_t = 0x08000000
)

type hb_buffer_content_type_t uint8

const (
	// Initial value for new buffer.
	HB_BUFFER_CONTENT_TYPE_INVALID hb_buffer_content_type_t = iota
	//The buffer contains input characters (before shaping).
	HB_BUFFER_CONTENT_TYPE_UNICODE
	//The buffer contains output glyphs (after shaping).
	HB_BUFFER_CONTENT_TYPE_GLYPHS
)

// The structure that holds various text properties of an #Buffer. Can be
// set and retrieved using hb_buffer_set_segment_properties() and
// hb_buffer_get_segment_properties(), respectively.
type hb_segment_properties_t struct {
	// the #Direction of the buffer, see hb_buffer_set_direction().
	Direction Direction
	// the #hb_script_t of the buffer, see hb_buffer_set_script().
	script hb_script_t
	//  the #hb_language_t of the buffer, see hb_buffer_set_language().
	language hb_language_t
}

// maximum length of additional context added outside
// input text
const CONTEXT_LENGTH = 5

//  Buffer is the main structure holding the input text and its properties before shaping,
// and output glyphs and their information after shaping.
type Buffer struct {
	/* Information about how the text in the buffer should be treated */
	//    hb_unicode_funcs_t *unicode; /* Unicode functions */
	flags         hb_buffer_flags_t /* BOT / EOT / etc. */
	cluster_level hb_buffer_cluster_level_t

	replacement rune /* U+FFFD or something else. */

	// rune that replaces invisible characters in
	// the shaping result.  If set to zero (default), the glyph for the
	// U+0020 SPACE character is used. Otherwise, this value is used
	// verbatim.
	invisible fonts.GlyphIndex
	Flags     hb_buffer_scratch_flags_t /* Have space-fallback, etc. */
	max_len   uint                      /* Maximum allowed len. */
	max_ops   int                       /* Maximum allowed operations. */

	/* Buffer contents */
	content_type hb_buffer_content_type_t
	Props        hb_segment_properties_t /* Script, language, direction */

	// successful bool; /* Allocations successful */
	have_output    bool /* Whether we have an output buffer going on */
	have_positions bool /* Whether we have positions */

	Idx int // Cursor into `info` and `pos` arrays

	Info     []GlyphInfo           // with length len, cap allocated
	Pos      []hb_glyph_position_t // with length len, cap allocated
	out_info []GlyphInfo           // with length out_len (if have_output)

	serial uint

	/* Text before / after the main buffer contents.
	* Always in Unicode, and ordered outward !
	* Index 0 is for "pre-Context", 1 for "post-Context". */
	Context [2][]rune
}

// Cur returns the glyph at the cursor, optionaly shifted by `i`.
// Its simply a syntactic sugar for `&b.Info[b.Idx+i] `
func (b *Buffer) Cur(i int) *GlyphInfo { return &b.Info[b.Idx+i] }

func (b *Buffer) cur_pos(i int) *hb_glyph_position_t { return &b.Pos[b.Idx+i] }

// check the access
func (b Buffer) prev() *GlyphInfo {
	if L := len(b.out_info); L != 0 {
		return &b.out_info[L-1]
	}
	return &b.out_info[0]
}

// func (b Buffer) has_separate_output() bool { return info != out_info }

func (b *Buffer) backtrack_len() int {
	if b.have_output {
		return len(b.out_info)
	}
	return b.Idx
}

func (b *Buffer) lookahead_len() int { return len(b.Info) - b.Idx }

func (b *Buffer) next_serial() uint {
	out := b.serial
	b.serial++
	return out
}

// func (b *Buffer) replace_glyph(glyph_index rune) {
// 	if unlikely(out_info != info || out_len != idx) {
// 		if unlikely(!make_room_for(1, 1)) {
// 			return
// 		}
// 		out_info[out_len] = info[idx]
// 	}
// 	out_info[out_len].Codepoint = glyph_index

// 	idx++
// 	out_len++
// }

// makes a copy of the glyph at idx to output and replace glyph_index
func (b *Buffer) output_glyph(r rune) *GlyphInfo {
	//  if (unlikely (!make_room_for (0, 1))) return Crap (GlyphInfo);

	if b.Idx == len(b.Info) && len(b.out_info) == 0 {
		return nil
	}

	if b.Idx < len(b.Info) {
		b.out_info = append(b.out_info, b.Info[b.Idx])
	} else {
		b.out_info = append(b.out_info, b.out_info[len(b.out_info)-1])
	}
	b.out_info[len(b.out_info)].Codepoint = r

	return &b.out_info[len(b.out_info)-1]
}

// func (b *Buffer) output_info(glyph_info *GlyphInfo) {
// 	if unlikely(!make_room_for(0, 1)) {
// 		return
// 	}

// 	out_info[out_len] = glyph_info

// 	out_len++
// }

// /* Copies glyph at idx to output but doesn't advance idx */
// func (b *Buffer) copy_glyph() {
// 	if unlikely(!make_room_for(0, 1)) {
// 		return
// 	}

// 	out_info[out_len] = info[idx]

// 	out_len++
// }

// Copies glyph at idx to output and advance idx.
// If there's no output, just advance idx.
func (b *Buffer) NextGlyph() {
	if b.have_output {
		//TODO: check
		// if b.out_info != info || out_len != idx {
		// if unlikely(!make_room_for(1, 1)) {
		// return
		// }
		b.out_info = append(b.out_info, b.Info[b.Idx])
		// }
		// out_len++
	}

	b.Idx++
}

/* Copies n glyphs at idx to output and advance idx.
* If there's no output, just advance idx. */
func (b *Buffer) next_glyphs(n int) { // TODO:
	// if have_output {
	// 	if out_info != info || out_len != idx {
	// 		if unlikely(!make_room_for(n, n)) {
	// 			return
	// 		}
	// 		memmove(out_info+out_len, info+idx, n*sizeof(out_info[0]))
	// 	}
	// 	out_len += n
	// }

	// idx += n
}

// advances idx without copying to output
func (b *Buffer) skip_glyph() { b.Idx++ }

func (b *Buffer) reset_masks(mask Mask) {
	for j := range b.Info {
		b.Info[j].Mask = mask
	}
}

func (b *Buffer) set_masks(value, mask Mask, clusterStart, clusterEnd int) {
	notMask := ^mask
	value &= mask

	if mask == 0 {
		return
	}

	for i, info := range b.Info {
		if clusterStart <= info.cluster && info.cluster < clusterEnd {
			b.Info[i].Mask = (info.Mask & notMask) | value
		}
	}
}

func (b *Buffer) add_masks(mask Mask) {
	for j := range b.Info {
		b.Info[j].Mask |= mask
	}
}

func (b *Buffer) MergeClusters(start, end int) {
	if end-start < 2 {
		return
	}

	if b.cluster_level == HB_BUFFER_CLUSTER_LEVEL_CHARACTERS {
		b.UnsafeToBreak(start, end)
		return
	}

	cluster := b.Info[start].cluster

	for i := start + 1; i < end; i++ {
		cluster = min(cluster, b.Info[i].cluster)
	}

	/* Extend end */
	for end < len(b.Info) && b.Info[end-1].cluster == b.Info[end].cluster {
		end++
	}

	/* Extend start */
	for b.Idx < start && b.Info[start-1].cluster == b.Info[start].cluster {
		start--
	}

	/* If we hit the start of buffer, continue in out-buffer. */
	if b.Idx == start {
		for i := len(b.out_info); i != 0 && b.out_info[i-1].cluster == b.Info[start].cluster; i-- {
			b.out_info[i-1].set_cluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.Info[i].set_cluster(cluster, 0)
	}
}

//    /* Merge clusters for deleting current glyph, and skip it. */
//    HB_INTERNAL void delete_glyph ();

func (b *Buffer) UnsafeToBreak(start, end int) {
	if end-start < 2 {
		return
	}
	b.unsafe_to_break_impl(start, end)
}

func (b *Buffer) unsafe_to_break_impl(start, end int) {
	cluster := _unsafe_to_break_find_min_cluster(b.Info, start, end, maxInt)
	b._unsafe_to_break_set_mask(b.Info, start, end, cluster)
}

func _unsafe_to_break_find_min_cluster(infos []GlyphInfo,
	start, end, cluster int) int {
	for i := start; i < end; i++ {
		cluster = min(cluster, infos[i].cluster)
	}
	return cluster
}

func (b *Buffer) _unsafe_to_break_set_mask(infos []GlyphInfo,
	start, end, cluster int) {
	for i := start; i < end; i++ {
		if cluster != infos[i].cluster {
			b.Flags |= HB_BUFFER_SCRATCH_FLAG_HAS_UNSAFE_TO_BREAK
			infos[i].Mask |= HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		}
	}
}

// zeros the `pos` array and truncate `out_info`
func (b *Buffer) clear_positions() {
	b.have_output = false
	b.have_positions = true

	b.out_info = b.Info[:0]
	for i := range b.Pos {
		b.Pos[i] = hb_glyph_position_t{}
	}
}

func (b *Buffer) clear_output() {
	b.have_output = true
	b.have_positions = false

	b.out_info = b.Info[:0]
}

// Ensure grow the slices to `size`, re-allocating and copying if needed.
func (b *Buffer) Ensure(size int) {
	if L := len(b.Info); L <= size {
		b.Info = append(b.Info, make([]GlyphInfo, size-L)...)
		b.Pos = append(b.Pos, make([]hb_glyph_position_t, size-L)...)
	}
}

//    { return likely (!size || size < allocated) ? true : enlarge (size); }

//    bool ensure_inplace (uint size)
//    { return likely (!size || size < allocated); }

//    void assert_glyphs ()
//    {
// 	 assert ((content_type == HB_BUFFER_CONTENT_TYPE_GLYPHS) ||
// 		 (!len && (content_type == HB_BUFFER_CONTENT_TYPE_INVALID)));
//    }
//    void assert_unicode ()
//    {
// 	 assert ((content_type == HB_BUFFER_CONTENT_TYPE_UNICODE) ||
// 		 (!len && (content_type == HB_BUFFER_CONTENT_TYPE_INVALID)));
//    }
//    bool ensure_glyphs ()
//    {
// 	 if (unlikely (content_type != HB_BUFFER_CONTENT_TYPE_GLYPHS))
// 	 {
// 	   if (content_type != HB_BUFFER_CONTENT_TYPE_INVALID)
// 	 return false;
// 	   assert (len == 0);
// 	   content_type = HB_BUFFER_CONTENT_TYPE_GLYPHS;
// 	 }
// 	 return true;
//    }
//    bool ensure_unicode ()
//    {
// 	 if (unlikely (content_type != HB_BUFFER_CONTENT_TYPE_UNICODE))
// 	 {
// 	   if (content_type != HB_BUFFER_CONTENT_TYPE_INVALID)
// 	 return false;
// 	   assert (len == 0);
// 	   content_type = HB_BUFFER_CONTENT_TYPE_UNICODE;
// 	 }
// 	 return true;
//    }

//    typedef long scratch_buffer_t;

func (b *Buffer) clear_context(side uint) { b.Context[side] = b.Context[side][:0] }

//    void unsafe_to_break_all () { unsafe_to_break_impl (0, len); }

func (b *Buffer) safe_to_break_all() {
	for i := range b.Info {
		b.Info[i].Mask &= ^HB_GLYPH_FLAG_UNSAFE_TO_BREAK
	}
}

/* Loop over clusters. Duplicated in foreach_syllable(). */
//  #define foreach_cluster(buffer, start, end) \
//    for (uint \
// 		_count = buffer.len, \
// 		start = 0, end = _count ? _next_cluster (buffer, 0) : 0; \
// 		start < _count; \
// 		start = end, end = _next_cluster (buffer, start))

//  static inline uint
//  _next_cluster (Buffer *buffer, uint start)
//  {
//    GlyphInfo *info = buffer.Info;
//    uint count = buffer.len;

//    uint cluster = info[start].cluster;
//    for (++start < count && cluster == info[start].cluster)
// 	 ;

//    return start;
//  }

//  #define HB_BUFFER_XALLOCATE_VAR(b, func, var) \
//    b.func (offsetof (GlyphInfo, var) - offsetof(GlyphInfo, var1), \
// 		sizeof (b.Info[0].var))
//  #define HB_BUFFER_ALLOCATE_VAR(b, var)		HB_BUFFER_XALLOCATE_VAR (b, allocate_var,   var ())
//  #define HB_BUFFER_DEALLOCATE_VAR(b, var)	HB_BUFFER_XALLOCATE_VAR (b, deallocate_var, var ())
//  #define HB_BUFFER_ASSERT_VAR(b, var)		HB_BUFFER_XALLOCATE_VAR (b, assert_var,     var ())

// Appends a character with the Unicode value of `codepoint` to `b`, and
// gives it the initial cluster value of `cluster`. Clusters can be any thing
// the client wants, they are usually used to refer to the index of the
// character in the input text stream and are output in the
// `GlyphInfo.cluster` field.
func (b *Buffer) hb_buffer_add(codepoint rune, cluster int) {
	b.append(codepoint, cluster)
	b.clear_context(1)
}

func (b *Buffer) append(codepoint rune, cluster int) {
	b.Info = append(b.Info, GlyphInfo{Codepoint: codepoint, cluster: cluster})
	b.Pos = append(b.Pos, hb_glyph_position_t{})
}

// Appends characters from `text` array to `b`. `itemOffset` is the
// position of the first character from `text` that will be appended, and
// `itemLength` is the number of character to add. When shaping part of a larger text
// (e.g. a run of text from a paragraph), instead of passing just the substring
// corresponding to the run, it is preferable to pass the whole
// paragraph and specify the run start and length as `itemOffset` and
// `itemLength`, respectively, to give HarfBuzz the full context to be able,
// for example, to do cross-run Arabic shaping or properly handle combining
// marks at stat of run.
func (b *Buffer) hb_buffer_add_codepoints(text []rune, itemOffset, itemLength int) {

	/* If buffer is empty and pre-context provided, install it.
	* This check is written this way, to make sure people can
	* provide pre-context in one add_utf() call, then provide
	* text in a follow-up call.  See:
	*
	* https://bugzilla.mozilla.org/show_bug.cgi?id=801410#c13 */
	if len(b.Info) == 0 && itemOffset > 0 {
		// add pre-context
		b.clear_context(0)
		prev := itemOffset - 1
		for prev >= 0 && len(b.Context[0]) < CONTEXT_LENGTH {
			b.Context[0] = append(b.Context[0], text[prev])
		}
	}

	for i, u := range text[itemOffset : itemOffset+itemLength] {
		b.append(u, itemOffset+i)
	}

	// add post-context
	s := itemOffset + itemLength + CONTEXT_LENGTH
	if s > len(text) {
		s = len(text)
	}
	b.Context[1] = text[itemOffset+itemLength : s]

	b.content_type = HB_BUFFER_CONTENT_TYPE_UNICODE
}

// reverses a subslice of the buffer contents
func (b *Buffer) reverse_range(start, end int) {
	if end-start < 2 {
		return
	}
	info := b.Info[start:end]
	pos := b.Pos[start:end]
	L := len(info)
	_ = pos[L] // BCE
	for i := L/2 - 1; i >= 0; i-- {
		opp := L - 1 - i
		info[i], info[opp] = info[opp], info[i]
		pos[i], pos[opp] = pos[opp], pos[i] // same length
	}
}

// reverses buffer contents.
func (b *Buffer) reverse() { b.reverse_range(0, len(b.Info)) }

// TODO:
func (b *Buffer) swap_buffers() {}

// iterator over the grapheme of a buffer
type graphemesIterator struct {
	buffer *Buffer
	start  int
}

// at the end of the buffer, start >= len(info)
func (g *graphemesIterator) next() (start, end int) {
	info := g.buffer.Info
	count := len(info)
	start = g.start
	for end = g.start + 1; end < count && info[end].isContinuation(); end++ {
	}
	g.start = end
	return start, end
}

func (buffer *Buffer) graphemesIterator() (*graphemesIterator, int) {
	return &graphemesIterator{buffer: buffer}, len(buffer.Info)
}

// iterator over clusters of a buffer
type clusterIterator struct {
	buffer *Buffer
	start  int
}

func (c *clusterIterator) next() (start, end int) {
	info := c.buffer.Info
	count := len(c.buffer.Info)
	start = c.start
	if count == 0 {
		return
	}
	cluster := info[start].cluster
	for end = start + 1; end < count && cluster == info[end].cluster; end++ {
	}
	c.start = end
	return start, end
}

func (buffer *Buffer) clusterIterator() (*clusterIterator, int) {
	return &clusterIterator{buffer: buffer}, len(buffer.Info)
}

func (b *Buffer) replace_glyphs(num_in int, glyph_data []rune) {
	//   if (unlikely (!make_room_for (num_in, num_out))) return;

	//   assert (idx + num_in <= len);

	b.MergeClusters(b.Idx, b.Idx+num_in)

	orig_info := info[idx]
	pinfo := &b.out_info[out_len]
	for _, d := range glyph_data {
		*pinfo = orig_info
		pinfo.Codepoint = d
		pinfo++
	}

	b.Idx += num_in
	out_len += len(glyph_data)
}

func (b *Buffer) sort(start, end int, compar func(a, b *GlyphInfo) int) {
	//   assert (!have_positions);
	for i := start + 1; i < end; i++ {
		j := i
		for j > start && compar(&b.Info[j-1], &b.Info[i]) > 0 {
			j--
		}
		if i == j {
			continue
		}
		// move item i to occupy place for item j, shift what's in between.
		b.MergeClusters(j, i+1)

		t := b.Info[i]
		copy(b.Info[j+1:], b.Info[j:i])
		b.Info[j] = t
	}
}

func (b *Buffer) merge_out_clusters(start, end int) {
	if b.cluster_level == HB_BUFFER_CLUSTER_LEVEL_CHARACTERS {
		return
	}

	if end-start < 2 {
		return
	}

	cluster := b.out_info[start].cluster

	for i := start + 1; i < end; i++ {
		cluster = min(cluster, b.out_info[i].cluster)
	}

	/* Extend start */
	for start != 0 && b.out_info[start-1].cluster == b.out_info[start].cluster {
		start--
	}

	/* Extend end */
	for end < len(b.out_info) && b.out_info[end-1].cluster == b.out_info[end].cluster {
		end++
	}

	/* If we hit the end of out-buffer, continue in buffer. */
	if end == len(b.out_info) {
		for i := b.Idx; i < len(b.Info) && b.Info[i].cluster == b.out_info[end-1].cluster; i++ {
			b.Info[i].set_cluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.out_info[i].set_cluster(cluster, 0)
	}
}
