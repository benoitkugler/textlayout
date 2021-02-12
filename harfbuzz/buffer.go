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

const isLigBase = 0x10

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

// maximum length of additional context added outside
// input text
const CONTEXT_LENGTH = 5

/* Here is how the buffer works internally:
 *
 * There are two info pointers: info and b.outInfo.  They always have
 * the same allocated size, but different lengths.
 *
 * As an optimization, both info and b.outInfo may point to the
 * same piece of memory, which is owned by info.  This remains the
 * case as long as out_len doesn't exceed idx at any time.
 * In that case, swap_buffers() is no-op and the glyph operations operate
 * mostly in-place.
 *
 * As soon as b.outInfo gets longer than info, b.outInfo is moved over
 * to an alternate buffer (which we reuse the pos buffer for!), and its
 * current contents (out_len entries) are copied to the new place.
 * This should all remain transparent to the user.  swap_buffers() then
 * switches info and b.outInfo.
 */

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

	// successful bool; /* Allocations successful */
	haveOutput     bool /* Whether we have an output buffer going on */
	have_positions bool /* Whether we have positions */

	idx int // Cursor into `info` and `pos` arrays

	// Info is used as internal storage during the shaping,
	// and also exposes the result: the glyph to display
	// and its original Cluster value.
	Info    []GlyphInfo
	outInfo []GlyphInfo // with length out_len (if haveOutput)
	// Pos gives the position of the glyphs resulting from the shapping
	// It has the same length has `Info`.
	Pos []GlyphPosition

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
	b.Info = append(b.Info, GlyphInfo{codepoint: codepoint, Cluster: cluster})
	b.Pos = append(b.Pos, GlyphPosition{})
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
	if len(b.Info) == 0 && itemOffset > 0 {
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
}

// cur returns the glyph at the cursor, optionaly shifted by `i`.
// Its simply a syntactic sugar for `&b.Info[b.idx+i] `
func (b *Buffer) cur(i int) *GlyphInfo { return &b.Info[b.idx+i] }

func (b *Buffer) cur_pos(i int) *GlyphPosition { return &b.Pos[b.idx+i] }

// returns the last glyph of `outInfo`
func (b Buffer) prev() *GlyphInfo {
	return &b.outInfo[len(b.outInfo)-1]
}

// func (b Buffer) has_separate_output() bool { return info != b.outInfo }

func (b *Buffer) backtrackLen() int {
	if b.haveOutput {
		return len(b.outInfo)
	}
	return b.idx
}

func (b *Buffer) lookaheadLen() int { return len(b.Info) - b.idx }

func (b *Buffer) next_serial() uint {
	out := b.serial
	b.serial++
	return out
}

// Copies glyph at `idx` to `outInfo` before replacing its codepoint by `u`
// Advances `idx`
func (b *Buffer) replaceGlyph(u rune) {
	b.outInfo = append(b.outInfo, b.Info[b.idx])
	b.outInfo[len(b.outInfo)-1].codepoint = u
	b.idx++
}

// Copies glyph at `idx` to `outInfo` before replacing its codepoint by `u`
// Advances `idx`
func (b *Buffer) replaceGlyphIndex(g fonts.GlyphIndex) {
	b.outInfo = append(b.outInfo, b.Info[b.idx])
	b.outInfo[len(b.outInfo)-1].Glyph = g
	b.idx++
}

// Merges clusters in [idx:idx+numIn], then dupplicate `Info[idx]` len(glyphData) times to `outInfo`
// before replacing their codepoint by `glyphData`
// Advances `idx` by `numIn`
// Assume that idx + numIn <= len(Info)
func (b *Buffer) replaceGlyphs(numIn int, glyphData []rune) {
	b.mergeClusters(b.idx, b.idx+numIn)

	origInfo := b.Info[b.idx]
	L := len(b.outInfo)
	b.outInfo = append(b.outInfo, make([]GlyphInfo, len(glyphData))...)
	for i, d := range glyphData {
		b.outInfo[L+i] = origInfo
		b.outInfo[L+i].codepoint = d
	}

	b.idx += numIn
}

// makes a copy of the glyph at idx to output and replace in output `codepoint`
// by `r`. Does NOT adavance `idx`
func (b *Buffer) outputGlyph(r rune) *GlyphInfo {
	out := b.output()
	out.codepoint = r
	return out
}

func (b *Buffer) output() *GlyphInfo {
	if b.idx == len(b.Info) && len(b.outInfo) == 0 {
		return &GlyphInfo{}
	}

	if b.idx < len(b.Info) {
		b.outInfo = append(b.outInfo, b.Info[b.idx])
	} else {
		b.outInfo = append(b.outInfo, b.outInfo[len(b.outInfo)-1])
	}
	out := &b.outInfo[len(b.outInfo)-1]

	return out
}

// same as outputGlyph
func (b *Buffer) outputGlyphIndex(g fonts.GlyphIndex) *GlyphInfo {
	out := b.output()
	out.Glyph = g
	return out
}

func (b *Buffer) OutputInfo(glyphInfo GlyphInfo) {
	b.outInfo = append(b.outInfo, glyphInfo)
}

// Copies glyph at idx to output but doesn't advance idx
func (b *Buffer) copyGlyph() {
	b.outInfo = append(b.outInfo, b.Info[b.idx])
}

// Copies glyph at `idx` to `outInfo` and advance `idx`.
// If there's no output, just advance `idx`.
func (b *Buffer) nextGlyph() {
	// TODO: remove this condition
	if b.haveOutput {
		b.outInfo = append(b.outInfo, b.Info[b.idx])
	}

	b.idx++
}

// Copies `n` glyphs from `idx` to `outInfo` and advances `idx`.
// If there's no output, just advance idx.
func (b *Buffer) nextGlyphs(n int) {
	// TODO: remove this condition
	if b.haveOutput {
		b.outInfo = append(b.outInfo, b.Info[b.idx:b.idx+n]...)
	}
	b.idx += n
}

// skipGlyph advances idx without copying to output
func (b *Buffer) skipGlyph() { b.idx++ }

func (b *Buffer) resetMasks(mask Mask) {
	for j := range b.Info {
		b.Info[j].mask = mask
	}
}

func (b *Buffer) setMasks(value, mask Mask, clusterStart, clusterEnd int) {
	notMask := ^mask
	value &= mask

	if mask == 0 {
		return
	}

	for i, info := range b.Info {
		if clusterStart <= info.Cluster && info.Cluster < clusterEnd {
			b.Info[i].mask = (info.mask & notMask) | value
		}
	}
}

func (b *Buffer) add_masks(mask Mask) {
	for j := range b.Info {
		b.Info[j].mask |= mask
	}
}

func (b *Buffer) mergeClusters(start, end int) {
	if end-start < 2 {
		return
	}

	if b.ClusterLevel == Characters {
		b.unsafeToBreak(start, end)
		return
	}

	cluster := b.Info[start].Cluster

	for i := start + 1; i < end; i++ {
		cluster = min(cluster, b.Info[i].Cluster)
	}

	/* Extend end */
	for end < len(b.Info) && b.Info[end-1].Cluster == b.Info[end].Cluster {
		end++
	}

	/* Extend start */
	for b.idx < start && b.Info[start-1].Cluster == b.Info[start].Cluster {
		start--
	}

	/* If we hit the start of buffer, continue in out-buffer. */
	// TODO: check the usage of out info
	if b.idx == start {
		startC := b.Info[start].Cluster
		for i := len(b.outInfo); i != 0 && b.outInfo[i-1].Cluster == startC; i-- {
			b.outInfo[i-1].setCluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.Info[i].setCluster(cluster, 0)
	}
}

// merge clusters for deleting current glyph, and skip it.
func (b *Buffer) deleteGlyph() {
	/* The logic here is duplicated in hb_ot_hide_default_ignorables(). */

	cluster := b.Info[b.idx].Cluster
	if b.idx+1 < len(b.Info) && cluster == b.Info[b.idx+1].Cluster {
		/* Cluster survives; do nothing. */
		goto done
	}

	if len(b.outInfo) != 0 {
		/* Merge cluster backward. */
		if cluster < b.outInfo[len(b.outInfo)-1].Cluster {
			mask := b.Info[b.idx].mask
			oldCluster := b.outInfo[len(b.outInfo)-1].Cluster
			for i := len(b.outInfo); i != 0 && b.outInfo[i-1].Cluster == oldCluster; i-- {
				b.outInfo[i-1].setCluster(cluster, mask)
			}
		}
		goto done
	}

	if b.idx+1 < len(b.Info) {
		/* Merge cluster forward. */
		b.mergeClusters(b.idx, b.idx+2)
		goto done
	}

done:
	b.skipGlyph()
}

// unsafeToBreak adds the flag `HB_GLYPH_FLAG_UNSAFE_TO_BREAK`
// when needed, between `start` and `end`.
func (b *Buffer) unsafeToBreak(start, end int) {
	if end-start < 2 {
		return
	}
	b.unsafeToBreakImpl(start, end)
}

func (b *Buffer) unsafeToBreakImpl(start, end int) {
	cluster := findMinCluster(b.Info, start, end, maxInt)
	b.unsafeToBreakSetMask(b.Info, start, end, cluster)
}

// return the smallest cluster between `cluster` and  infos[start:end]
func findMinCluster(infos []GlyphInfo, start, end, cluster int) int {
	for i := start; i < end; i++ {
		cluster = min(cluster, infos[i].Cluster)
	}
	return cluster
}

func (b *Buffer) unsafeToBreakSetMask(infos []GlyphInfo,
	start, end, cluster int) {
	for i := start; i < end; i++ {
		if cluster != infos[i].Cluster {
			b.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_UNSAFE_TO_BREAK
			infos[i].mask |= HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		}
	}
}

func (b *Buffer) unsafeToBreakFromOutbuffer(start, end int) {
	if !b.haveOutput {
		b.unsafeToBreakImpl(start, end)
		return
	}

	//   assert (start <= out_len);
	//   assert (idx <= end);

	cluster := math.MaxInt32
	cluster = findMinCluster(b.outInfo, start, len(b.outInfo), cluster)
	cluster = findMinCluster(b.Info, b.idx, end, cluster)
	b.unsafeToBreakSetMask(b.outInfo, start, len(b.outInfo), cluster)
	b.unsafeToBreakSetMask(b.Info, b.idx, end, cluster)
}

// reset `b.outInfo`, and adjust `pos` to have
// same length as `Info` (without zeroing its values)
func (b *Buffer) clearPositions() {
	b.haveOutput = false
	b.have_positions = true

	b.outInfo = b.outInfo[:0]

	L := len(b.Info)
	if cap(b.Pos) >= L {
		b.Pos = b.Pos[:L]
	} else {
		b.Pos = make([]GlyphPosition, L)
	}
}

// truncate `outInfo` and set `haveOutput` to true
func (b *Buffer) clearOutput() {
	b.haveOutput = true
	b.have_positions = false // TODO: remove ?

	b.outInfo = b.outInfo[:0]
}

// isAlias reports whether x and y share the same base array.
// Note: isAlias assumes that the capacity of underlying arrays
// is never changed; i.e. that there are
// no 3-operand slice expressions in this code (or worse,
// reflect-based operations to the same effect).
func isAlias(x, y []GlyphInfo) bool {
	return cap(x) != 0 && cap(y) != 0 && &x[0:cap(x)][cap(x)-1] == &y[0:cap(y)][cap(y)-1]
}

// ensure grow the slices to `size`, re-allocating and copying if needed.
func (b *Buffer) ensure(size int) {
	sameOutput := isAlias(b.Info, b.outInfo)
	if L := len(b.Info); L < size {
		b.Info = append(b.Info, make([]GlyphInfo, size-L)...)
		b.Pos = append(b.Pos, make([]GlyphPosition, size-L)...)

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

func (b *Buffer) unsafeToBreakAll() { b.unsafeToBreakImpl(0, len(b.Info)) }

// safeToBreakAll remove the flag `HB_GLYPH_FLAG_UNSAFE_TO_BREAK`
// to all glyphs.
func (b *Buffer) safeToBreakAll() {
	info := b.Info
	for i := range info {
		info[i].mask &= ^HB_GLYPH_FLAG_UNSAFE_TO_BREAK
	}
}

// reverses the subslice [start:end] of the buffer contents
func (b *Buffer) reverseRange(start, end int) {
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

// Reverse reverses buffer contents, that is the `Info` and `Pos` slices.
func (b *Buffer) Reverse() { b.reverseRange(0, len(b.Info)) }

// swap back the temporary outInfo buffer to `Info`
// and resets the cursor `idx`
// Assume that haveOutput is true, and toogle it.
func (b *Buffer) swapBuffers() {
	b.haveOutput = false
	b.Info, b.outInfo = b.outInfo, b.Info
	b.idx = 0
}

// returns an unique id
func (b *Buffer) allocateLigId() uint8 {
	ligId := uint8(b.serial & 0x07)
	b.serial++
	if ligId == 0 { // in case of overflow
		ligId = b.allocateLigId()
	}
	return ligId
}

func (b *Buffer) shiftForward(count int) {
	//   assert (have_output);
	L := len(b.Info)
	b.Info = append(b.Info, make([]GlyphInfo, count)...)
	copy(b.Info[b.idx+count:], b.Info[b.idx:L])
	b.idx += count
}

func (b *Buffer) moveTo(i int) {
	if !b.haveOutput {
		// assert(i <= len)
		b.idx = i
		return
	}

	// assert(i <= out_len+(len-idx))
	outL := len(b.outInfo)
	if outL < i {
		count := i - outL
		b.outInfo = append(b.outInfo, b.Info[b.idx:count+b.idx]...)
		b.idx += count
	} else if outL > i {
		/* Tricky part: rewinding... */
		count := outL - i

		if b.idx < count {
			b.shiftForward(count + 0)
		}

		// assert(idx >= count)

		b.idx -= count
		copy(b.Info[b.idx:], b.outInfo[outL-count:outL])
		b.outInfo = b.outInfo[:outL-count]
	}
}

// iterator over the grapheme of a buffer
type graphemesIterator struct {
	buffer *Buffer
	start  int
}

// at the end of the buffer, start >= len(info)
func (g *graphemesIterator) Next() (start, end int) {
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
type ClusterIterator struct {
	buffer *Buffer
	start  int
}

func (c *ClusterIterator) Next() (start, end int) {
	info := c.buffer.Info
	count := len(c.buffer.Info)
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
	return &ClusterIterator{buffer: buffer}, len(buffer.Info)
}

// iterator over syllables of a buffer
type syllableIterator struct {
	buffer *Buffer
	start  int
}

func (c *syllableIterator) Next() (start, end int) {
	info := c.buffer.Info
	count := len(c.buffer.Info)
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
	return &syllableIterator{buffer: buffer}, len(buffer.Info)
}

// only modifies Info, thus assume Pos is not used yet
func (b *Buffer) sort(start, end int, compar func(a, b *GlyphInfo) int) {
	for i := start + 1; i < end; i++ {
		j := i
		for j > start && compar(&b.Info[j-1], &b.Info[i]) > 0 {
			j--
		}
		if i == j {
			continue
		}
		// move item i to occupy place for item j, shift what's in between.
		b.mergeClusters(j, i+1)

		t := b.Info[i]
		copy(b.Info[j+1:], b.Info[j:i])
		b.Info[j] = t
	}
}

func (b *Buffer) mergeOutClusters(start, end int) {
	if b.ClusterLevel == Characters {
		return
	}

	if end-start < 2 {
		return
	}

	cluster := b.outInfo[start].Cluster

	for i := start + 1; i < end; i++ {
		cluster = min(cluster, b.outInfo[i].Cluster)
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
		endC := b.outInfo[end-1].Cluster
		for i := b.idx; i < len(b.Info) && b.Info[i].Cluster == endC; i++ {
			b.Info[i].setCluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.outInfo[i].setCluster(cluster, 0)
	}
}
