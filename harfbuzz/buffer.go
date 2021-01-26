package harfbuzz

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

//  static_assert ((sizeof (hb_glyph_info_t) == 20), "");
//  static_assert ((sizeof (hb_glyph_info_t) == sizeof (hb_glyph_position_t)), "");

type hb_mask_t uint32

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
	HB_GLYPH_FLAG_UNSAFE_TO_BREAK hb_mask_t = 0x00000001

	// OR of all defined flags
	HB_GLYPH_FLAG_DEFINED hb_mask_t = HB_GLYPH_FLAG_UNSAFE_TO_BREAK
)

// The #hb_glyph_info_t is the structure that holds information about the
// glyphs and their relation to input text.
type hb_glyph_info_t struct {
	// either a Unicode code point (before shaping) or a glyph index
	// (after shaping).
	codepoint rune

	// the index of the character in the original text that corresponds
	// to this #hb_glyph_info_t, or whatever the client passes to
	// hb_buffer_add(). More than one #hb_glyph_info_t can have the same
	// `cluster` value, if they resulted from the same character (e.g. one
	// to many glyph substitution), and when more than one character gets
	// merged in the same glyph (e.g. many to one glyph substitution) the
	// #hb_glyph_info_t will have the smallest cluster value of them.
	// By default some characters are merged into the same cluster
	// (e.g. combining marks have the same cluster as their bases)
	// even if they are separate glyphs, hb_buffer_set_cluster_level()
	// allow selecting more fine-grained cluster handling.
	cluster int

	mask hb_mask_t
}

func (inf *hb_glyph_info_t) set_cluster(cluster int, mask hb_mask_t) {
	if inf.cluster != cluster {
		if mask&HB_GLYPH_FLAG_UNSAFE_TO_BREAK != 0 {
			inf.mask |= HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		} else {
			inf.mask &= ^HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		}
	}
	inf.cluster = cluster
}

// The #hb_glyph_position_t is the structure that holds the positions of the
// glyph in both horizontal and vertical directions.
// All positions are relative to the current point.
type hb_glyph_position_t struct {
	// how much the line advances after drawing this glyph when setting
	// text in horizontal direction.
	x_advance hb_position_t
	// how much the line advances after drawing this glyph when setting
	// text in vertical direction.
	y_advance hb_position_t
	// how much the glyph moves on the X-axis before drawing it, this
	// should not affect how much the line advances.
	x_offset hb_position_t
	// how much the glyph moves on the Y-axis before drawing it, this
	// should not affect how much the line advances.
	y_offset hb_position_t
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

// The structure that holds various text properties of an #hb_buffer_t. Can be
// set and retrieved using hb_buffer_set_segment_properties() and
// hb_buffer_get_segment_properties(), respectively.
type hb_segment_properties_t struct {
	// the #hb_direction_t of the buffer, see hb_buffer_set_direction().
	direction hb_direction_t
	// the #hb_script_t of the buffer, see hb_buffer_set_script().
	script hb_script_t
	//  the #hb_language_t of the buffer, see hb_buffer_set_language().
	language hb_language_t
}

// maximum length of additional context added outside
// input text
const CONTEXT_LENGTH = 5

//  hb_buffer_t is the main structure holding the input text and its properties before shaping,
// and output glyphs and their information after shaping.
type hb_buffer_t struct {
	/* Information about how the text in the buffer should be treated */
	//    hb_unicode_funcs_t *unicode; /* Unicode functions */
	flags         hb_buffer_flags_t /* BOT / EOT / etc. */
	cluster_level hb_buffer_cluster_level_t

	replacement rune /* U+FFFD or something else. */

	// rune that replaces invisible characters in
	// the shaping result.  If set to zero (default), the glyph for the
	// U+0020 SPACE character is used. Otherwise, this value is used
	// verbatim.
	invisible     rune
	scratch_flags hb_buffer_scratch_flags_t /* Have space-fallback, etc. */
	max_len       uint                      /* Maximum allowed len. */
	max_ops       int                       /* Maximum allowed operations. */

	/* Buffer contents */
	content_type hb_buffer_content_type_t
	props        hb_segment_properties_t /* Script, language, direction */

	// successful bool; /* Allocations successful */
	have_output    bool /* Whether we have an output buffer going on */
	have_positions bool /* Whether we have positions */

	idx int // Cursor into `info` and `pos` arrays

	info     []hb_glyph_info_t     // with length len, cap allocated
	pos      []hb_glyph_position_t // with length len, cap allocated
	out_info []hb_glyph_info_t     // with length out_len (if have_output)

	serial uint

	/* Text before / after the main buffer contents.
	* Always in Unicode, and ordered outward !
	* Index 0 is for "pre-context", 1 for "post-context". */
	context [2][]rune
}

func (b *hb_buffer_t) cur(i int) *hb_glyph_info_t         { return &b.info[b.idx+i] }
func (b *hb_buffer_t) cur_pos(i int) *hb_glyph_position_t { return &b.pos[b.idx+i] }

// check the access
func (b hb_buffer_t) prev() *hb_glyph_info_t {
	if L := len(b.out_info); L != 0 {
		return &b.out_info[L-1]
	}
	return &b.out_info[0]
}

// func (b hb_buffer_t) has_separate_output() bool { return info != out_info }

func (b *hb_buffer_t) backtrack_len() int {
	if b.have_output {
		return len(b.out_info)
	}
	return b.idx
}

func (b *hb_buffer_t) lookahead_len() int { return len(b.info) - b.idx }

func (b *hb_buffer_t) next_serial() uint {
	out := b.serial
	b.serial++
	return out
}

// func (b *hb_buffer_t) replace_glyph(glyph_index rune) {
// 	if unlikely(out_info != info || out_len != idx) {
// 		if unlikely(!make_room_for(1, 1)) {
// 			return
// 		}
// 		out_info[out_len] = info[idx]
// 	}
// 	out_info[out_len].codepoint = glyph_index

// 	idx++
// 	out_len++
// }

// /* Makes a copy of the glyph at idx to output and replace glyph_index */
// func (b *hb_buffer_t) output_glyph(glyph_index rune) *hb_glyph_info_t {
// 	//  if (unlikely (!make_room_for (0, 1))) return Crap (hb_glyph_info_t);

// 	if unlikely(idx == len && !out_len) {
// 		return Crap(hb_glyph_info_t)
// 	}

// 	if idx < len {
// 		out_info[out_len] = info[idx]
// 	} else {
// 		out_info[out_len] = out_info[out_len-1]
// 	}
// 	out_info[out_len].codepoint = glyph_index

// 	out_len++

// 	return out_info[out_len-1]
// }

// func (b *hb_buffer_t) output_info(glyph_info *hb_glyph_info_t) {
// 	if unlikely(!make_room_for(0, 1)) {
// 		return
// 	}

// 	out_info[out_len] = glyph_info

// 	out_len++
// }

// /* Copies glyph at idx to output but doesn't advance idx */
// func (b *hb_buffer_t) copy_glyph() {
// 	if unlikely(!make_room_for(0, 1)) {
// 		return
// 	}

// 	out_info[out_len] = info[idx]

// 	out_len++
// }

// /* Copies glyph at idx to output and advance idx.
// * If there's no output, just advance idx. */
// func (b *hb_buffer_t) next_glyph() {
// 	if have_output {
// 		if out_info != info || out_len != idx {
// 			if unlikely(!make_room_for(1, 1)) {
// 				return
// 			}
// 			out_info[out_len] = info[idx]
// 		}
// 		out_len++
// 	}

// 	idx++
// }

// /* Copies n glyphs at idx to output and advance idx.
// * If there's no output, just advance idx. */
// func (b *hb_buffer_t) next_glyphs(n uint) {
// 	if have_output {
// 		if out_info != info || out_len != idx {
// 			if unlikely(!make_room_for(n, n)) {
// 				return
// 			}
// 			memmove(out_info+out_len, info+idx, n*sizeof(out_info[0]))
// 		}
// 		out_len += n
// 	}

// 	idx += n
// }

// advances idx without copying to output
func (b *hb_buffer_t) skip_glyph() { b.idx++ }

func (b *hb_buffer_t) reset_masks(mask hb_mask_t) {
	for j := range b.info {
		b.info[j].mask = mask
	}
}
func (b *hb_buffer_t) add_masks(mask hb_mask_t) {
	for j := range b.info {
		b.info[j].mask |= mask
	}
}

func (b *hb_buffer_t) merge_clusters(start, end int) {
	if end-start < 2 {
		return
	}

	if b.cluster_level == HB_BUFFER_CLUSTER_LEVEL_CHARACTERS {
		b.unsafe_to_break(start, end)
		return
	}

	cluster := b.info[start].cluster

	for i := start + 1; i < end; i++ {
		cluster = min(cluster, b.info[i].cluster)
	}

	/* Extend end */
	for end < len(b.info) && b.info[end-1].cluster == b.info[end].cluster {
		end++
	}

	/* Extend start */
	for b.idx < start && b.info[start-1].cluster == b.info[start].cluster {
		start--
	}

	/* If we hit the start of buffer, continue in out-buffer. */
	if b.idx == start {
		for i := len(b.out_info); i != 0 && b.out_info[i-1].cluster == b.info[start].cluster; i-- {
			b.out_info[i-1].set_cluster(cluster, 0)
		}
	}

	for i := start; i < end; i++ {
		b.info[i].set_cluster(cluster, 0)
	}
}

//    /* Merge clusters for deleting current glyph, and skip it. */
//    HB_INTERNAL void delete_glyph ();

func (b *hb_buffer_t) unsafe_to_break(start, end int) {
	if end-start < 2 {
		return
	}
	b.unsafe_to_break_impl(start, end)
}

func (b *hb_buffer_t) unsafe_to_break_impl(start, end int) {
	cluster := _unsafe_to_break_find_min_cluster(b.info, start, end, maxInt)
	b._unsafe_to_break_set_mask(b.info, start, end, cluster)
}

func _unsafe_to_break_find_min_cluster(infos []hb_glyph_info_t,
	start, end, cluster int) int {
	for i := start; i < end; i++ {
		cluster = min(cluster, infos[i].cluster)
	}
	return cluster
}

func (b *hb_buffer_t) _unsafe_to_break_set_mask(infos []hb_glyph_info_t,
	start, end, cluster int) {
	for i := start; i < end; i++ {
		if cluster != infos[i].cluster {
			b.scratch_flags |= HB_BUFFER_SCRATCH_FLAG_HAS_UNSAFE_TO_BREAK
			infos[i].mask |= HB_GLYPH_FLAG_UNSAFE_TO_BREAK
		}
	}
}

//    bool ensure (uint size)
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

func (b *hb_buffer_t) clear_context(side uint) { b.context[side] = b.context[side][:0] }

//    HB_INTERNAL void sort (uint start, uint end, int(*compar)(const hb_glyph_info_t *, const hb_glyph_info_t *));

//    void unsafe_to_break_all () { unsafe_to_break_impl (0, len); }
//    void safe_to_break_all ()
//    {
// 	 for (uint i = 0; i < len; i++)
// 	   info[i].mask &= ~HB_GLYPH_FLAG_UNSAFE_TO_BREAK;
//    }
//  };
//  DECLARE_NULL_INSTANCE (hb_buffer_t);

/* Loop over clusters. Duplicated in foreach_syllable(). */
//  #define foreach_cluster(buffer, start, end) \
//    for (uint \
// 		_count = buffer.len, \
// 		start = 0, end = _count ? _next_cluster (buffer, 0) : 0; \
// 		start < _count; \
// 		start = end, end = _next_cluster (buffer, start))

//  static inline uint
//  _next_cluster (hb_buffer_t *buffer, uint start)
//  {
//    hb_glyph_info_t *info = buffer.info;
//    uint count = buffer.len;

//    uint cluster = info[start].cluster;
//    for (++start < count && cluster == info[start].cluster)
// 	 ;

//    return start;
//  }

//  #define HB_BUFFER_XALLOCATE_VAR(b, func, var) \
//    b.func (offsetof (hb_glyph_info_t, var) - offsetof(hb_glyph_info_t, var1), \
// 		sizeof (b.info[0].var))
//  #define HB_BUFFER_ALLOCATE_VAR(b, var)		HB_BUFFER_XALLOCATE_VAR (b, allocate_var,   var ())
//  #define HB_BUFFER_DEALLOCATE_VAR(b, var)	HB_BUFFER_XALLOCATE_VAR (b, deallocate_var, var ())
//  #define HB_BUFFER_ASSERT_VAR(b, var)		HB_BUFFER_XALLOCATE_VAR (b, assert_var,     var ())

// Appends a character with the Unicode value of `codepoint` to `b`, and
// gives it the initial cluster value of `cluster`. Clusters can be any thing
// the client wants, they are usually used to refer to the index of the
// character in the input text stream and are output in the
// `hb_glyph_info_t.cluster` field.
func (b *hb_buffer_t) hb_buffer_add(codepoint rune, cluster int) {
	b.append(codepoint, cluster)
	b.clear_context(1)
}

func (b *hb_buffer_t) append(codepoint rune, cluster int) {
	b.info = append(b.info, hb_glyph_info_t{codepoint: codepoint, cluster: cluster})
	b.pos = append(b.pos, hb_glyph_position_t{})
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
func (b *hb_buffer_t) hb_buffer_add_codepoints(text []rune, itemOffset, itemLength int) {

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
