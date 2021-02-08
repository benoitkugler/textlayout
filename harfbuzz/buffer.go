package harfbuzz

import (
	"math"

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
//  static_assert ((sizeof (GlyphInfo) == sizeof (GlyphPosition)), "");

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
	// The buffer contains input characters (before shaping).
	HB_BUFFER_CONTENT_TYPE_UNICODE
	// The buffer contains output glyphs (after shaping).
	HB_BUFFER_CONTENT_TYPE_GLYPHS
)

// maximum length of additional context added outside
// input text
const CONTEXT_LENGTH = 5

// Buffer is the main structure holding the input text segment and its properties before shaping,
// and output glyphs and their information after shaping.
type Buffer struct {
	// Information about how the text in the buffer should be treated
	Flags        Flags
	ClusterLevel ClusterLevel

	// Default to U+FFFD
	Replacement rune

	// Glyph that replaces invisible characters in
	// the shaping result. If set to zero (default), the glyph for the
	// U+0020 SPACE character is used. Otherwise, this value is used
	// verbatim.
	Invisible fonts.GlyphIndex

	// Props is required to correctly interpret the input runes.
	Props SegmentProperties

	scratchFlags hb_buffer_scratch_flags_t /* Have space-fallback, etc. */
	max_len      uint                      /* Maximum allowed len. */
	max_ops      int                       /* Maximum allowed operations. */

	/* Buffer contents */
	content_type hb_buffer_content_type_t

	// successful bool; /* Allocations successful */
	have_output    bool /* Whether we have an output buffer going on */
	have_positions bool /* Whether we have positions */

	idx int // Cursor into `info` and `pos` arrays

	// Info is used as internal storage during the shaping,
	// and also exposes the result: the glyph to display
	// and its original Cluster value.
	info []GlyphInfo
	// Pos gives the position of the glyphs resulting from the shapping
	pos     []GlyphPosition
	outInfo []GlyphInfo // with length out_len (if have_output)

	serial uint

	/* Text before / after the main buffer contents.
	* Always in Unicode, and ordered outward !
	* Index 0 is for "pre-Context", 1 for "post-Context". */
	context [2][]rune
}

// AddRune appends a character with the Unicode value of `codepoint` to `b`, and
// gives it the initial cluster value of `cluster`. Clusters can be any thing
// the client wants, they are usually used to refer to the index of the
// character in the input text stream and are output in the
// `GlyphInfo.cluster` field.
// This also clears the posterior context.
func (b *Buffer) AddRune(codepoint rune, cluster int) {
	b.append(codepoint, cluster)
	b.clear_context(1)
}

func (b *Buffer) append(codepoint rune, cluster int) {
	b.info = append(b.info, GlyphInfo{codepoint: codepoint, Cluster: cluster})
	b.pos = append(b.pos, GlyphPosition{})
}

// AddRunes appends characters from `text` array to `b`. `itemOffset` is the
// position of the first character from `text` that will be appended, and
// `itemLength` is the number of character to add.
// When shaping part of a larger text (e.g. a run of text from a paragraph),
// instead of passing just the substring
// corresponding to the run, it is preferable to pass the whole
// paragraph and specify the run start and length as `itemOffset` and
// `itemLength`, respectively, to give HarfBuzz the full context to be able,
// for example, to do cross-run Arabic shaping or properly handle combining
// marks at start of run.
func (b *Buffer) AddRunes(text []rune, itemOffset, itemLength int) {
	/* If buffer is empty and pre-context provided, install it.
	* This check is written this way, to make sure people can
	* provide pre-context in one add_utf() call, then provide
	* text in a follow-up call.  See:
	*
	* https://bugzilla.mozilla.org/show_bug.cgi?id=801410#c13 */
	if len(b.info) == 0 && itemOffset > 0 {
		// add pre-context
		b.clear_context(0)
		prev := itemOffset - 1
		for prev >= 0 && len(b.context[0]) < CONTEXT_LENGTH {
			b.context[0] = append(b.context[0], text[prev])
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
	b.context[1] = text[itemOffset+itemLength : s]

	b.content_type = HB_BUFFER_CONTENT_TYPE_UNICODE
}

// Cur returns the glyph at the cursor, optionaly shifted by `i`.
// Its simply a syntactic sugar for `&b.Info[b.idx+i] `
func (b *Buffer) Cur(i int) *GlyphInfo { return &b.info[b.idx+i] }

func (b *Buffer) cur_pos(i int) *GlyphPosition { return &b.pos[b.idx+i] }

// check the access
func (b Buffer) Prev() *GlyphInfo {
	if L := len(b.outInfo); L != 0 {
		return &b.outInfo[L-1]
	}
	return &b.outInfo[0]
}

// func (b Buffer) has_separate_output() bool { return info != out_info }

func (b *Buffer) backtrack_len() int {
	if b.have_output {
		return len(b.outInfo)
	}
	return b.idx
}

func (b *Buffer) lookahead_len() int { return len(b.info) - b.idx }

func (b *Buffer) next_serial() uint {
	out := b.serial
	b.serial++
	return out
}

// TODO:
func (b *Buffer) ReplaceGlyph(glyph_index rune) {
	// if unlikely(out_info != info || out_len != idx) {
	// 	if unlikely(!make_room_for(1, 1)) {
	// 		return
	// 	}
	// 	out_info[out_len] = info[idx]
	// }
	// out_info[out_len].Codepoint = glyph_index

	// idx++
	// out_len++
}

// makes a copy of the glyph at idx to output and replace glyph_index
func (b *Buffer) OutputGlyph(r rune) *GlyphInfo {
	//  if (unlikely (!make_room_for (0, 1))) return Crap (GlyphInfo);

	if b.idx == len(b.info) && len(b.outInfo) == 0 {
		return nil
	}

	if b.idx < len(b.info) {
		b.outInfo = append(b.outInfo, b.info[b.idx])
	} else {
		b.outInfo = append(b.outInfo, b.outInfo[len(b.outInfo)-1])
	}
	b.outInfo[len(b.outInfo)].codepoint = r

	return &b.outInfo[len(b.outInfo)-1]
}

func (b *Buffer) OutputInfo(glyphInfo GlyphInfo) {
	b.outInfo = append(b.outInfo, glyphInfo)
}

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
		// TODO: check
		// if b.out_info != info || out_len != idx {
		// if unlikely(!make_room_for(1, 1)) {
		// return
		// }
		b.outInfo = append(b.outInfo, b.info[b.idx])
		// }
		// out_len++
	}

	b.idx++
}

/* Copies n glyphs at idx to output and advance idx.
* If there's no output, just advance idx. */
func (b *Buffer) NextGlyphs(n int) { // TODO:
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

// SkipGlyph advances idx without copying to output
func (b *Buffer) SkipGlyph() { b.idx++ }

func (b *Buffer) ResetMasks(mask Mask) {
	for j := range b.info {
		b.info[j].mask = mask
	}
}

func (b *Buffer) SetMasks(value, mask Mask, clusterStart, clusterEnd int) {
	notMask := ^mask
	value &= mask

	if mask == 0 {
		return
	}

	for i, info := range b.info {
		if clusterStart <= info.Cluster && info.Cluster < clusterEnd {
			b.info[i].mask = (info.mask & notMask) | value
		}
	}
}

func (b *Buffer) add_masks(mask Mask) {
	for j := range b.info {
		b.info[j].mask |= mask
	}
}

func (b *Buffer) MergeClusters(start, end int) {
	if end-start < 2 {
		return
	}

	if b.ClusterLevel == Characters {
		b.UnsafeToBreak(start, end)
		return
	}

	cluster := b.info[start].Cluster

	for i := start + 1; i < end; i++ {
		cluster = Min(cluster, b.info[i].Cluster)
	}

	/* Extend end */
	for end < len(b.info) && b.info[end-1].Cluster == b.info[end].Cluster {
		end++
	}

	/* Extend start */
	for b.idx < start && b.info[start-1].Cluster == b.info[start].Cluster {
		start--
	}

	/* If we hit the start of buffer, continue in out-buffer. */
	if b.idx == start {
		for i := len(b.outInfo); i != 0 && b.outInfo[i-1].Cluster == b.info[start].Cluster; i-- {
			b.outInfo[i-1].set_cluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.info[i].set_cluster(cluster, 0)
	}
}

//    /* Merge clusters for deleting current glyph, and skip it. */
//    HB_INTERNAL void delete_glyph ();

// UnsafeToBreak adds the flag `HB_GLYPH_FLAG_UNSAFE_TO_BREAK`
// when needed, between `start` and `end`.
func (b *Buffer) UnsafeToBreak(start, end int) {
	if end-start < 2 {
		return
	}
	b.unsafe_to_break_impl(start, end)
}

func (b *Buffer) unsafe_to_break_impl(start, end int) {
	cluster := _unsafe_to_break_find_min_cluster(b.info, start, end, maxInt)
	b._unsafe_to_break_set_mask(b.info, start, end, cluster)
}

func _unsafe_to_break_find_min_cluster(infos []GlyphInfo,
	start, end, cluster int) int {
	for i := start; i < end; i++ {
		cluster = Min(cluster, infos[i].Cluster)
	}
	return cluster
}

func (b *Buffer) _unsafe_to_break_set_mask(infos []GlyphInfo,
	start, end, cluster int) {
	for i := start; i < end; i++ {
		if cluster != infos[i].Cluster {
			b.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_UNSAFE_TO_BREAK
			infos[i].mask |= HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		}
	}
}

func (b *Buffer) UnsafeToBreakFromOutbuffer(start, end int) {
	if !b.have_output {
		b.unsafe_to_break_impl(start, end)
		return
	}

	//   assert (start <= out_len);
	//   assert (idx <= end);

	cluster := math.MaxInt32
	cluster = _unsafe_to_break_find_min_cluster(b.outInfo, start, len(b.outInfo), cluster)
	cluster = _unsafe_to_break_find_min_cluster(b.info, b.idx, end, cluster)
	b._unsafe_to_break_set_mask(b.outInfo, start, len(b.outInfo), cluster)
	b._unsafe_to_break_set_mask(b.info, b.idx, end, cluster)
}

// zeros the `pos` array and truncate `out_info`
func (b *Buffer) ClearPositions() {
	b.have_output = false
	b.have_positions = true

	b.outInfo = b.info[:0]
	for i := range b.pos {
		b.pos[i] = GlyphPosition{}
	}
}

func (b *Buffer) ClearOutput() {
	b.have_output = true
	b.have_positions = false

	b.outInfo = b.info[:0]
}

// Ensure grow the slices to `size`, re-allocating and copying if needed.
func (b *Buffer) Ensure(size int) {
	if L := len(b.info); L <= size {
		b.info = append(b.info, make([]GlyphInfo, size-L)...)
		b.pos = append(b.pos, make([]GlyphPosition, size-L)...)
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

func (b *Buffer) clear_context(side uint) { b.context[side] = b.context[side][:0] }

//    void unsafe_to_break_all () { unsafe_to_break_impl (0, len); }

// SafeToBreakAll remove the flag `HB_GLYPH_FLAG_UNSAFE_TO_BREAK`
// to all glyphs.
func (b *Buffer) SafeToBreakAll() {
	info := b.info
	for i := range info {
		info[i].mask &= ^HB_GLYPH_FLAG_UNSAFE_TO_BREAK
	}
}

// reverses a subslice of the buffer contents
func (b *Buffer) reverse_range(start, end int) {
	if end-start < 2 {
		return
	}
	info := b.info[start:end]
	pos := b.pos[start:end]
	L := len(info)
	_ = pos[L] // BCE
	for i := L/2 - 1; i >= 0; i-- {
		opp := L - 1 - i
		info[i], info[opp] = info[opp], info[i]
		pos[i], pos[opp] = pos[opp], pos[i] // same length
	}
}

// reverses buffer contents.
func (b *Buffer) Reverse() { b.reverse_range(0, len(b.info)) }

// TODO:
func (b *Buffer) SwapBuffers() {}

// iterator over the grapheme of a buffer
type graphemesIterator struct {
	buffer *Buffer
	start  int
}

// at the end of the buffer, start >= len(info)
func (g *graphemesIterator) Next() (start, end int) {
	info := g.buffer.info
	count := len(info)
	start = g.start
	for end = g.start + 1; end < count && info[end].isContinuation(); end++ {
	}
	g.start = end
	return start, end
}

func (buffer *Buffer) GraphemesIterator() (*graphemesIterator, int) {
	return &graphemesIterator{buffer: buffer}, len(buffer.info)
}

// iterator over clusters of a buffer
type ClusterIterator struct {
	buffer *Buffer
	start  int
}

func (c *ClusterIterator) Next() (start, end int) {
	info := c.buffer.info
	count := len(c.buffer.info)
	start = c.start
	if count == 0 {
		return
	}
	cluster := info[start].Cluster
	for end = start + 1; end < count && cluster == info[end].Cluster; end++ {
	}
	c.start = end
	return start, end
}

func (buffer *Buffer) ClusterIterator() (*ClusterIterator, int) {
	return &ClusterIterator{buffer: buffer}, len(buffer.info)
}

// iterator over syllables of a buffer
type syllableIterator struct {
	buffer *Buffer
	start  int
}

func (c *syllableIterator) Next() (start, end int) {
	info := c.buffer.info
	count := len(c.buffer.info)
	start = c.start
	if count == 0 {
		return
	}
	syllable := info[start].syllable
	for end = start + 1; end < count && syllable == info[end].syllable; end++ {
	}
	c.start = end
	return start, end
}

func (buffer *Buffer) SyllableIterator() (*syllableIterator, int) {
	return &syllableIterator{buffer: buffer}, len(buffer.info)
}

func (b *Buffer) ReplaceGlyphs(num_in int, glyph_data []rune) {
	//   if (unlikely (!make_room_for (num_in, num_out))) return;

	//   assert (idx + num_in <= len);

	b.MergeClusters(b.idx, b.idx+num_in)

	orig_info := info[idx]
	pinfo := &b.outInfo[out_len]
	for _, d := range glyph_data {
		*pinfo = orig_info
		pinfo.codepoint = d
		pinfo++
	}

	b.idx += num_in
	out_len += len(glyph_data)
}

func (b *Buffer) Sort(start, end int, compar func(a, b *GlyphInfo) int) {
	//   assert (!have_positions);
	for i := start + 1; i < end; i++ {
		j := i
		for j > start && compar(&b.info[j-1], &b.info[i]) > 0 {
			j--
		}
		if i == j {
			continue
		}
		// move item i to occupy place for item j, shift what's in between.
		b.MergeClusters(j, i+1)

		t := b.info[i]
		copy(b.info[j+1:], b.info[j:i])
		b.info[j] = t
	}
}

func (b *Buffer) MergeOutClusters(start, end int) {
	if b.ClusterLevel == Characters {
		return
	}

	if end-start < 2 {
		return
	}

	cluster := b.outInfo[start].Cluster

	for i := start + 1; i < end; i++ {
		cluster = Min(cluster, b.outInfo[i].Cluster)
	}

	/* Extend start */
	for start != 0 && b.outInfo[start-1].Cluster == b.outInfo[start].Cluster {
		start--
	}

	/* Extend end */
	for end < len(b.outInfo) && b.outInfo[end-1].Cluster == b.outInfo[end].Cluster {
		end++
	}

	/* If we hit the end of out-buffer, continue in buffer. */
	if end == len(b.outInfo) {
		for i := b.idx; i < len(b.info) && b.info[i].Cluster == b.outInfo[end-1].Cluster; i++ {
			b.info[i].set_cluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.outInfo[i].set_cluster(cluster, 0)
	}
}
