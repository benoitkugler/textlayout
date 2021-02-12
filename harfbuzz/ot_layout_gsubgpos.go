package harfbuzz

import (
	"math"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-layout-gsubgpos.hh Copyright © 2007,2008,2009,2010  Red Hat, Inc. 2010,2012  Google, Inc.  Behdad Esfahbod

type TLookup interface {
	// accumulate the subtables coverage into the diggest
	collectCoverage(*SetDigest)
	// walk the subtables to add them to the context
	dispatch(*hb_get_subtables_context_t)
}

/*
 * GSUB/GPOS Common
 */

const IgnoreFlags = uint16(truetype.IgnoreBaseGlyphs | truetype.IgnoreLigatures | truetype.IgnoreMarks)

type hb_ot_layout_lookup_accelerator_t struct {
	digest    SetDigest // union of all the subtables coverage
	subtables hb_get_subtables_context_t
}

func (ac *hb_ot_layout_lookup_accelerator_t) init(lookup TLookup) {
	ac.digest = SetDigest{}
	lookup.collectCoverage(&ac.digest)
	ac.subtables = nil
	lookup.dispatch(&ac.subtables)
}

// apply the subtables and stops at the first success.
func (ac *hb_ot_layout_lookup_accelerator_t) apply(c *hb_ot_apply_context_t) bool {
	for _, table := range ac.subtables {
		if table.apply(c) {
			return true
		}
	}
	return false
}

type hb_apply_func_t interface {
	apply(c *hb_ot_apply_context_t) bool
	// get_coverage() truetype.Coverage
}

// represents one layout subtable, with its own coverage
type hb_applicable_t struct {
	digest SetDigest
	obj    hb_apply_func_t
}

func new_hb_applicable_t(table truetype.LookupGSUBSubtable) hb_applicable_t {
	ap := hb_applicable_t{obj: gsubSubtable(table)}
	ap.digest.collectCoverage(table.Coverage)
	return ap
}

func (ap hb_applicable_t) apply(c *hb_ot_apply_context_t) bool {
	return ap.digest.mayHave(c.buffer.cur(0).Glyph) && ap.obj.apply(c)
}

type hb_get_subtables_context_t []hb_applicable_t

type otProxy struct {
	tableIndex   int // GPOS/GSUB
	inplace      bool
	accels       []hb_ot_layout_lookup_accelerator_t
	recurse_func recurse_func_t
}

//  {
//    template <typename Type>
//    static inline bool apply_to (const void *obj, OT::c *hb_ot_apply_context_t)
//    {
// 	 const Type *typed_obj = (const Type *) obj;
// 	 return typed_obj.apply (c);
//    }

//    typedef hb_vector_t<hb_applicable_t> array_t;

//    /* Dispatch interface. */
//    template <typename T>
//    return_t dispatch (const T &obj)
//    {
// 	 hb_applicable_t *entry = array.push();
// 	 entry.init (obj, apply_to<T>);
// 	 return hb_empty_t ();
//    }
//    static return_t default_return_value () { return hb_empty_t (); }
//  };

//  struct hb_intersects_context_t :
// 		hb_dispatch_context_t<hb_intersects_context_t, bool>
//  {
//    template <typename T>
//    return_t dispatch (const T &obj) { return obj.intersects (this.glyphs); }
//    static return_t default_return_value () { return false; }
//    bool stop_sublookup_iteration (return_t r) const { return r; }

//    const hb_set_t *glyphs;

//    hb_intersects_context_t (const hb_set_t *glyphs_) :
// 				  glyphs (glyphs_) {}
//  };

//  struct hb_closure_context_t :
// 		hb_dispatch_context_t<hb_closure_context_t>
//  {
//    typedef return_t (*recurse_func_t) (hb_closure_context_t *c, uint lookupIndex);
//    template <typename T>
//    return_t dispatch (const T &obj) { obj.closure (this); return hb_empty_t (); }
//    static return_t default_return_value () { return hb_empty_t (); }
//    void recurse (uint lookupIndex)
//    {
// 	 if (unlikely (nesting_level_left == 0 || !recurse_func))
// 	   return;

// 	 nesting_level_left--;
// 	 recurse_func (this, lookupIndex);
// 	 nesting_level_left++;
//    }

//    bool lookup_limit_exceeded ()
//    { return lookup_count > HB_MAX_LOOKUP_INDICES; }

//    bool should_visit_lookup (uint lookupIndex)
//    {
// 	 if (lookup_count++ > HB_MAX_LOOKUP_INDICES)
// 	   return false;

// 	 if (is_lookup_done (lookupIndex))
// 	   return false;

// 	 done_lookups.set (lookupIndex, glyphs.get_population ());
// 	 return true;
//    }

//    bool is_lookup_done (uint lookupIndex)
//    {
// 	 if (done_lookups.in_error ())
// 	   return true;

// 	 /* Have we visited this lookup with the current set of glyphs? */
// 	 return done_lookups.get (lookupIndex) == glyphs.get_population ();
//    }

//    Face *face;
//    hb_set_t *glyphs;
//    hb_set_t output[1];
//    recurse_func_t recurse_func;
//    uint nesting_level_left;

//    hb_closure_context_t (Face *face_,
// 			 hb_set_t *glyphs_,
// 			 hb_map_t *done_lookups_,
// 			 uint nesting_level_left_ = HB_MAX_NESTING_LEVEL) :
// 			   face (face_),
// 			   glyphs (glyphs_),
// 			   recurse_func (nil),
// 			   nesting_level_left (nesting_level_left_),
// 			   done_lookups (done_lookups_),
// 			   lookup_count (0)
//    {}

//    ~hb_closure_context_t () { flush (); }

//    void set_recurse_func (recurse_func_t func) { recurse_func = func; }

//    void flush ()
//    {
// 	 hb_set_del_range (output, face.get_num_glyphs (), hb_set_get_max (output));	/* Remove invalid glyphs. */
// 	 hb_set_union (glyphs, output);
// 	 hb_set_clear (output);
//    }

//    private:
//    hb_map_t *done_lookups;
//    uint lookup_count;
//  };

//  struct hb_closure_lookups_context_t :
// 		hb_dispatch_context_t<hb_closure_lookups_context_t>
//  {
//    typedef return_t (*recurse_func_t) (hb_closure_lookups_context_t *c, unsigned lookupIndex);
//    template <typename T>
//    return_t dispatch (const T &obj) { obj.closure_lookups (this); return hb_empty_t (); }
//    static return_t default_return_value () { return hb_empty_t (); }
//    void recurse (unsigned lookupIndex)
//    {
// 	 if (unlikely (nesting_level_left == 0 || !recurse_func))
// 	   return;

// 	 /* Return if new lookup was recursed to before. */
// 	 if (is_lookup_visited (lookupIndex))
// 	   return;

// 	 set_lookup_visited (lookupIndex);
// 	 nesting_level_left--;
// 	 recurse_func (this, lookupIndex);
// 	 nesting_level_left++;
//    }

//    void set_lookup_visited (unsigned lookupIndex)
//    { visited_lookups.add (lookupIndex); }

//    void set_lookup_inactive (unsigned lookupIndex)
//    { inactive_lookups.add (lookupIndex); }

//    bool lookup_limit_exceeded ()
//    { return lookup_count > HB_MAX_LOOKUP_INDICES; }

//    bool is_lookup_visited (unsigned lookupIndex)
//    {
// 	 if (lookup_count++ > HB_MAX_LOOKUP_INDICES)
// 	   return true;

// 	 if (visited_lookups.in_error ())
// 	   return true;

// 	 return visited_lookups.has (lookupIndex);
//    }

//    Face *face;
//    const hb_set_t *glyphs;
//    recurse_func_t recurse_func;
//    uint nesting_level_left;

//    hb_closure_lookups_context_t (Face *face_,
// 				 const hb_set_t *glyphs_,
// 				 hb_set_t *visited_lookups_,
// 				 hb_set_t *inactive_lookups_,
// 				 unsigned nesting_level_left_ = HB_MAX_NESTING_LEVEL) :
// 				 face (face_),
// 				 glyphs (glyphs_),
// 				 recurse_func (nil),
// 				 nesting_level_left (nesting_level_left_),
// 				 visited_lookups (visited_lookups_),
// 				 inactive_lookups (inactive_lookups_),
// 				 lookup_count (0) {}

//    void set_recurse_func (recurse_func_t func) { recurse_func = func; }

//    private:
//    hb_set_t *visited_lookups;
//    hb_set_t *inactive_lookups;
//    uint lookup_count;
//  };

//  struct hb_would_apply_context_t :
// 		hb_dispatch_context_t<hb_would_apply_context_t, bool>
//  {
//    template <typename T>
//    return_t dispatch (const T &obj) { return obj.would_apply (this); }
//    static return_t default_return_value () { return false; }
//    bool stop_sublookup_iteration (return_t r) const { return r; }

//    Face *face;
//    const hb_codepoint_t *glyphs;
//    uint len;
//    bool zero_context;

//    hb_would_apply_context_t (Face *face_,
// 				 const hb_codepoint_t *glyphs_,
// 				 uint len_,
// 				 bool zero_context_) :
// 				   face (face_),
// 				   glyphs (glyphs_),
// 				   len (len_),
// 				   zero_context (zero_context_) {}
//  };

//  struct hb_collect_glyphs_context_t :
// 		hb_dispatch_context_t<hb_collect_glyphs_context_t>
//  {
//    typedef return_t (*recurse_func_t) (hb_collect_glyphs_context_t *c, uint lookupIndex);
//    template <typename T>
//    return_t dispatch (const T &obj) { obj.collect_glyphs (this); return hb_empty_t (); }
//    static return_t default_return_value () { return hb_empty_t (); }
//    void recurse (uint lookupIndex)
//    {
// 	 if (unlikely (nesting_level_left == 0 || !recurse_func))
// 	   return;

// 	 /* Note that GPOS sets recurse_func to nil already, so it doesn't get
// 	  * past the previous check.  For GSUB, we only want to collect the output
// 	  * glyphs in the recursion.  If output is not requested, we can go home now.
// 	  *
// 	  * Note further, that the above is not exactly correct.  A recursed lookup
// 	  * is allowed to match input that is not matched in the context, but that's
// 	  * not how most fonts are built.  It's possible to relax that and recurse
// 	  * with all sets here if it proves to be an issue.
// 	  */

// 	 if (output == hb_set_get_empty ())
// 	   return;

// 	 /* Return if new lookup was recursed to before. */
// 	 if (recursed_lookups.has (lookupIndex))
// 	   return;

// 	 hb_set_t *old_before = before;
// 	 hb_set_t *old_input  = input;
// 	 hb_set_t *old_after  = after;
// 	 before = input = after = hb_set_get_empty ();

// 	 nesting_level_left--;
// 	 recurse_func (this, lookupIndex);
// 	 nesting_level_left++;

// 	 before = old_before;
// 	 input  = old_input;
// 	 after  = old_after;

// 	 recursed_lookups.add (lookupIndex);
//    }

//    Face *face;
//    hb_set_t *before;
//    hb_set_t *input;
//    hb_set_t *after;
//    hb_set_t *output;
//    recurse_func_t recurse_func;
//    hb_set_t *recursed_lookups;
//    uint nesting_level_left;

//    hb_collect_glyphs_context_t (Face *face_,
// 					hb_set_t  *glyphs_before, /* OUT.  May be NULL */
// 					hb_set_t  *glyphs_input,  /* OUT.  May be NULL */
// 					hb_set_t  *glyphs_after,  /* OUT.  May be NULL */
// 					hb_set_t  *glyphs_output, /* OUT.  May be NULL */
// 					uint nesting_level_left_ = HB_MAX_NESTING_LEVEL) :
// 				   face (face_),
// 				   before (glyphs_before ? glyphs_before : hb_set_get_empty ()),
// 				   input  (glyphs_input  ? glyphs_input  : hb_set_get_empty ()),
// 				   after  (glyphs_after  ? glyphs_after  : hb_set_get_empty ()),
// 				   output (glyphs_output ? glyphs_output : hb_set_get_empty ()),
// 				   recurse_func (nil),
// 				   recursed_lookups (hb_set_create ()),
// 				   nesting_level_left (nesting_level_left_) {}
//    ~hb_collect_glyphs_context_t () { hb_set_destroy (recursed_lookups); }

//    void set_recurse_func (recurse_func_t func) { recurse_func = func; }
//  };

//  template <typename set_t>
//  struct hb_collect_coverage_context_t :
// 		hb_dispatch_context_t<hb_collect_coverage_context_t<set_t>, const Coverage &>
//  {
//    typedef const Coverage &return_t; // Stoopid that we have to dupe this here.
//    template <typename T>
//    return_t dispatch (const T &obj) { return obj.get_coverage (); }
//    static return_t default_return_value () { return Null (Coverage); }
//    bool stop_sublookup_iteration (return_t r) const
//    {
// 	 r.collectCoverage (set);
// 	 return false;
//    }

//    hb_collect_coverage_context_t (set_t *set_) :
// 					set (set_) {}

//    set_t *set;
//  };

// `value` interpretation is dictated by the context
type match_func_t = func(gid fonts.GlyphIndex, value uint16) bool

// interprets `value` as a Glyph
func matchGlyph(gid fonts.GlyphIndex, value uint16) bool { return gid == fonts.GlyphIndex(value) }

// TODO: maybe inline manually
// interprets `value` as a Class
func matchClass(class truetype.Class) match_func_t {
	return func(gid fonts.GlyphIndex, value uint16) bool {
		return class.ClassID(gid) == value
	}
}

// interprets `value` as an index in coverage array
func matchCoverage(covs []truetype.Coverage) match_func_t {
	return func(gid fonts.GlyphIndex, value uint16) bool {
		_, covered := covs[value].Index(gid)
		return covered
	}
}

const (
	No = iota
	Yes
	Maybe
)

type hb_ot_apply_context_matcher_t struct {
	lookupProps uint32
	ignore_zwnj bool
	ignore_zwj  bool
	mask        Mask
	syllable    uint8
	matchFunc   match_func_t
}

func (m hb_ot_apply_context_matcher_t) may_match(info *GlyphInfo, glyph_data uint16) uint8 {
	if info.mask&m.mask == 0 || (m.syllable != 0 && m.syllable != info.syllable) {
		return No
	}

	if m.matchFunc != nil {
		if m.matchFunc(info.Glyph, glyph_data) {
			return Yes
		}
		return No
	}

	return Maybe
}

func (m hb_ot_apply_context_matcher_t) may_skip(c *hb_ot_apply_context_t, info *GlyphInfo) uint8 {
	if !c.check_glyph_property(info, m.lookupProps) {
		return Yes
	}

	if info.isDefaultIgnorableAndNotHidden() && (m.ignore_zwnj || !info.isZwnj()) &&
		(m.ignore_zwj || !info.isZwj()) {
		return Maybe
	}

	return No
}

type skippingIterator struct {
	c                *hb_ot_apply_context_t
	idx              int
	matcher          hb_ot_apply_context_matcher_t
	match_glyph_data []uint16

	num_items, end int
}

func (it *skippingIterator) init(c *hb_ot_apply_context_t, context_match bool) {
	it.c = c
	it.match_glyph_data = nil
	it.matcher.matchFunc = nil
	it.matcher.lookupProps = c.lookupProps
	/* Ignore ZWNJ if we are matching GPOS, or matching GSUB context and asked to. */
	it.matcher.ignore_zwnj = c.table_index == 1 || (context_match && c.auto_zwnj)
	/* Ignore ZWJ if we are matching context, or asked to. */
	it.matcher.ignore_zwj = context_match || c.auto_zwj
	if context_match {
		it.matcher.mask = math.MaxUint32
	} else {
		it.matcher.mask = c.lookup_mask
	}
}

// 	 void set_lookup_props (uint lookupProps)
// 	 {
// 	   matcher.set_lookup_props (lookupProps);
// 	 }
// 	 void set_match_func (matcher_t::match_func_t match_func_,
// 			  const void *match_data_,
// 			  const HBUINT16 glyph_data[])
// 	 {
// 	   matcher.set_match_func (match_func_, match_data_);
// 	   match_glyph_data = glyph_data;
// 	 }

func (it *skippingIterator) reset(startIndex, numItems int) {
	it.idx = startIndex
	it.num_items = numItems
	it.end = len(it.c.buffer.Info)
	if startIndex == it.c.buffer.idx {
		it.matcher.syllable = it.c.buffer.cur(0).syllable
	} else {
		it.matcher.syllable = 0
	}
}

// 	 void reject ()
// 	 {
// 	   num_items++;
// 	   if (match_glyph_data) match_glyph_data--;
// 	 }

func (it *skippingIterator) may_skip(info *GlyphInfo) uint8 { return it.matcher.may_skip(it.c, info) }

func (it *skippingIterator) next() bool {
	//    assert (num_items > 0);
	for it.idx+it.num_items < it.end {
		it.idx++
		info := &it.c.buffer.Info[it.idx]

		skip := it.matcher.may_skip(it.c, info)
		if skip == Yes {
			continue
		}

		match := it.matcher.may_match(info, it.match_glyph_data[0])
		if match == Yes || (match == Maybe && skip == No) {
			it.num_items--
			if len(it.match_glyph_data) != 0 {
				it.match_glyph_data = it.match_glyph_data[1:]
			}
			return true
		}

		if skip == No {
			return false
		}
	}
	return false
}

func (it *skippingIterator) prev() bool {
	//    assert (num_items > 0);
	for it.idx > it.num_items-1 {
		it.idx--
		info := &it.c.buffer.outInfo[it.idx]

		skip := it.matcher.may_skip(it.c, info)
		if skip == Yes {
			continue
		}

		match := it.matcher.may_match(info, it.match_glyph_data[0])
		if match == Yes || (match == Maybe && skip == No) {
			it.num_items--
			if len(it.match_glyph_data) != 0 {
				it.match_glyph_data = it.match_glyph_data[1:]
			}
			return true
		}

		if skip == No {
			return false
		}
	}
	return false
}

type recurse_func_t = func(c *hb_ot_apply_context_t, lookupIndex uint16) bool

type hb_ot_apply_context_t struct {
	iter_input, iter_context skippingIterator

	font         *Font
	face         Face
	buffer       *Buffer
	recurse_func recurse_func_t
	gdef         truetype.TableGDEF
	//    const VariationStore &var_store;

	direction          Direction
	lookup_mask        Mask
	table_index        int /* GSUB/GPOS */
	lookupIndex        uint16
	lookupProps        uint32
	nesting_level_left int

	has_glyph_classes bool
	auto_zwnj         bool
	auto_zwj          bool
	random            bool

	randomState uint32

	// see get1N()
	indices []uint16
}

func new_hb_ot_apply_context_t(table_index int, font *Font, buffer *Buffer) hb_ot_apply_context_t {
	var out hb_ot_apply_context_t
	// iter_input (), iter_context (),
	out.font = font
	out.face = font.Face
	out.buffer = buffer
	// TODO:
	//    gdef ( *face.table.GDEF.table)
	//    var_store (gdef.get_var_store ()),
	out.direction = buffer.Props.Direction
	out.lookup_mask = 1
	out.table_index = table_index
	out.lookupIndex = math.MaxUint16 // TODO: check
	out.nesting_level_left = HB_MAX_NESTING_LEVEL
	out.has_glyph_classes = gdef.has_glyph_classes()
	out.auto_zwnj = true
	out.auto_zwj = true
	out.randomState = 1

	out.init_iters()
	return out
}

func (c *hb_ot_apply_context_t) init_iters() {
	c.iter_input.init(c, false)
	c.iter_context.init(c, true)
}

func (c *hb_ot_apply_context_t) set_lookup_mask(mask Mask) {
	c.lookup_mask = mask
	c.init_iters()
}

func (c *hb_ot_apply_context_t) set_auto_zwj(autoZwj bool) {
	c.auto_zwj = autoZwj
	c.init_iters()
}

func (c *hb_ot_apply_context_t) set_auto_zwnj(autoZwnj bool) {
	c.auto_zwnj = autoZwnj
	c.init_iters()
}

func (c *hb_ot_apply_context_t) set_lookup_props(lookupProps uint32) {
	c.lookupProps = lookupProps
	c.init_iters()
}

func (c *hb_ot_apply_context_t) hb_ot_layout_substitute_lookup(lookup lookupGSUB, accel *hb_ot_layout_lookup_accelerator_t) {
	c.apply_string(lookup, accel)
}

func (c *hb_ot_apply_context_t) check_glyph_property(info *GlyphInfo, matchProps uint32) bool {
	glyphProps := info.glyphProps

	/* Not covered, if, for example, glyph class is ligature and
	 * matchProps includes LookupFlags::IgnoreLigatures */
	if (glyphProps & uint16(matchProps) & IgnoreFlags) != 0 {
		return false
	}

	if glyphProps&truetype.Mark != 0 {
		return c.matchPropertiesMark(info.Glyph, glyphProps, matchProps)
	}

	return true
}

func (c *hb_ot_apply_context_t) matchPropertiesMark(glyph fonts.GlyphIndex, glyphProps uint16, matchProps uint32) bool {
	/* If using mark filtering sets, the high short of
	 * matchProps has the set index. */
	if truetype.LookupFlag(matchProps)&truetype.UseMarkFilteringSet != 0 {
		return gdef.mark_set_covers(matchProps>>16, glyph)
	}

	/* The second byte of matchProps has the meaning
	 * "ignore marks of attachment type different than
	 * the attachment type specified." */
	if truetype.LookupFlag(matchProps)&truetype.MarkAttachmentType != 0 {
		return uint16(matchProps)&truetype.MarkAttachmentType == (glyphProps & truetype.MarkAttachmentType)
	}

	return true
}

func (c *hb_ot_apply_context_t) setGlyphProps(glyphIndex fonts.GlyphIndex) {
	c.setGlyphPropsExt(glyphIndex, 0, false, false)
}

func (c *hb_ot_apply_context_t) setGlyphPropsExt(glyphIndex fonts.GlyphIndex, class_guess uint16, ligature, component bool) {
	add_in := c.buffer.cur(0).glyphProps & Preserve
	add_in |= Substituted
	if ligature {
		add_in |= Ligated
		/* In the only place that the MULTIPLIED bit is used, Uniscribe
		* seems to only care about the "last" transformation between
		* Ligature and Multiple substitutions.  Ie. if you ligate, expand,
		* and ligate again, it forgives the multiplication and acts as
		* if only ligation happened.  As such, clear MULTIPLIED bit.
		 */
		add_in &= ^Multiplied
	}
	if component {
		add_in |= Multiplied
	}
	if c.has_glyph_classes {
		c.buffer.cur(0).glyphProps = add_in | c.gdef.GetGlyphProps(glyphIndex)
	} else if class_guess != 0 {
		c.buffer.cur(0).glyphProps = add_in | class_guess
	}
}

func (c *hb_ot_apply_context_t) replaceGlyph(glyphIndex fonts.GlyphIndex) {
	c.setGlyphProps(glyphIndex)
	c.buffer.replaceGlyphIndex(glyphIndex)
}

func (c *hb_ot_apply_context_t) randomNumber() uint32 {
	/* http://www.cplusplus.com/reference/random/minstd_rand/ */
	c.randomState = c.randomState * 48271 % 2147483647
	return c.randomState
}

//  `input` starts with second glyph (`inputCount` = len(input)+1)
func (c *hb_ot_apply_context_t) contextApplyLookup(input []uint16, lookupRecord []truetype.SequenceLookup, lookupContext match_func_t) bool {
	matchLength := 0
	var matchPositions [HB_MAX_CONTEXT_LENGTH]int
	hasMatch, matchLength, _ := c.matchInput(input, lookupContext, matchPositions)
	if !hasMatch {
		return false
	}

	c.buffer.unsafeToBreak(c.buffer.idx, c.buffer.idx+matchLength)
	c.applyLookup(len(input)+1, matchPositions, lookupRecord, matchLength)
	return true
}

//  `input` starts with second glyph (`inputCount` = len(input)+1)
// lookupsContexts : backtrack, input, lookahead
func (c *hb_ot_apply_context_t) chainContextApplyLookup(backtrack, input, lookahead []uint16,
	lookupRecord []truetype.SequenceLookup, lookupContexts [3]match_func_t) bool {
	var matchPositions [HB_MAX_CONTEXT_LENGTH]int

	hasMatch, matchLength, _ := c.matchInput(input, lookupContexts[1], matchPositions)
	if !hasMatch {
		return false
	}

	hasMatch, startIndex := c.matchBacktrack(backtrack, lookupContexts[0])
	if !hasMatch {
		return false
	}

	hasMatch, endIndex := c.matchLookahead(lookahead, lookupContexts[2], matchLength)
	if !hasMatch {
		return false
	}

	c.buffer.unsafeToBreakFromOutbuffer(startIndex, endIndex)
	c.applyLookup(len(input)+1, matchPositions, lookupRecord, matchLength)
	return true
}

// 		hb_dispatch_context_t<hb_ot_apply_context_t, bool, HB_DEBUG_APPLY>
//  {
//

//    const char *get_name () { return "APPLY"; }
//    typedef return_t (*recurse_func_t) (c *hb_ot_apply_context_t, uint lookupIndex);
//    template <typename T>
//    return_t dispatch (const T &obj) { return obj.apply (this); }
//    static return_t default_return_value () { return false; }
//    bool stop_sublookup_iteration (return_t r) const { return r; }

//    void set_random (bool random_) { random = random_; }
//    void set_recurse_func (recurse_func_t func) { recurse_func = func; }
//    void set_lookup_index (uint lookup_index_) { lookupIndex = lookup_index_; }

//    void ReplaceGlyph_inplace (hb_codepoint_t glyphIndex) const
//    {
// 	 setGlyphProps (glyphIndex);
// 	 buffer.cur(0).Codepoint = glyphIndex;
//    }
//    void ReplaceGlyph_with_ligature (hb_codepoint_t glyphIndex,
// 					 uint class_guess) const
//    {
// 	 setGlyphProps (glyphIndex, class_guess, true);
// 	 buffer.replaceGlyph (glyphIndex);
//    }
//    void OutputGlyph_for_component (hb_codepoint_t glyphIndex,
// 					uint class_guess) const
//    {
// 	 setGlyphProps (glyphIndex, class_guess, false, true);
// 	 buffer.outputGlyph (glyphIndex);
//    }
//  };

//  typedef bool (*intersects_func_t) (const hb_set_t *glyphs, const HBUINT16 &value, const void *data);
//  typedef void (*collect_glyphs_func_t) (hb_set_t *glyphs, const HBUINT16 &value, const void *data);
//  typedef bool (*match_func_t) (hb_codepoint_t glyph_id, const HBUINT16 &value, const void *data);

//  struct ContextClosureFuncs
//  {
//    intersects_func_t intersects;
//  };
//  struct ContextCollectGlyphsFuncs
//  {
//    collect_glyphs_func_t collect;
//  };
//  struct ContextApplyFuncs
//  {
//    match_func_t match;
//  };

//  static inline bool intersects_glyph (const hb_set_t *glyphs, const HBUINT16 &value, const void *data HB_UNUSED)
//  {
//    return glyphs.has (value);
//  }
//  static inline bool intersects_class (const hb_set_t *glyphs, const HBUINT16 &value, const void *data)
//  {
//    const ClassDef &class_def = *reinterpret_cast<const ClassDef *>(data);
//    return class_def.intersects_class (glyphs, value);
//  }
//  static inline bool intersects_coverage (const hb_set_t *glyphs, const HBUINT16 &value, const void *data)
//  {
//    const OffsetTo<Coverage> &coverage = (const OffsetTo<Coverage>&)value;
//    return (data+coverage).intersects (glyphs);
//  }

//  static inline bool array_is_subset_of (const hb_set_t *glyphs,
// 						uint count,
// 						const HBUINT16 values[],
// 						intersects_func_t intersects_func,
// 						const void *intersects_data)
//  {
//    for (const HBUINT16 &_ : + hb_iter (values, count))
// 	 if (!intersects_func (glyphs, _, intersects_data)) return false;
//    return true;
//  }

//  static inline void collect_glyph (hb_set_t *glyphs, const HBUINT16 &value, const void *data HB_UNUSED)
//  {
//    glyphs.add (value);
//  }
//  static inline void collect_class (hb_set_t *glyphs, const HBUINT16 &value, const void *data)
//  {
//    const ClassDef &class_def = *reinterpret_cast<const ClassDef *>(data);
//    class_def.collect_class (glyphs, value);
//  }
//  static inline void collectCoverage (hb_set_t *glyphs, const HBUINT16 &value, const void *data)
//  {
//    const OffsetTo<Coverage> &coverage = (const OffsetTo<Coverage>&)value;
//    (data+coverage).collectCoverage (glyphs);
//  }
//  static inline void collect_array (hb_collect_glyphs_context_t *c HB_UNUSED,
// 				   hb_set_t *glyphs,
// 				   uint count,
// 				   const HBUINT16 values[],
// 				   collect_glyphs_func_t collect_func,
// 				   const void *collect_data)
//  {
//    return
//    + hb_iter (values, count)
//    | hb_apply ([&] (const HBUINT16 &_) { collect_func (glyphs, _, collect_data); })
//    ;
//  }

//  static inline bool match_glyph (hb_codepoint_t glyph_id, const HBUINT16 &value, const void *data HB_UNUSED)
//  {
//    return glyph_id == value;
//  }
//  static inline bool match_class (hb_codepoint_t glyph_id, const HBUINT16 &value, const void *data)
//  {
//    const ClassDef &class_def = *reinterpret_cast<const ClassDef *>(data);
//    return class_def.get_class (glyph_id) == value;
//  }
//  static inline bool match_coverage (hb_codepoint_t glyph_id, const HBUINT16 &value, const void *data)
//  {
//    const OffsetTo<Coverage> &coverage = (const OffsetTo<Coverage>&)value;
//    return (data+coverage).get_coverage (glyph_id) != NOT_COVERED;
//  }

//  static inline bool would_match_input (hb_would_apply_context_t *c,
// 					   uint count, /* Including the first glyph (not matched) */
// 					   const HBUINT16 input[], /* Array of input values--start with second glyph */
// 					   match_func_t matchFunc,
// 					   const void *match_data)
//  {
//    if (count != c.len)
// 	 return false;

//    for (uint i = 1; i < count; i++)
// 	 if (likely (!matchFunc (c.glyphs[i], input[i - 1], match_data)))
// 	   return false;

//    return true;
//  }

// `input` starts with second glyph (`inputCount` = len(input)+1)
func (c *hb_ot_apply_context_t) matchInput(input []uint16, matchFunc match_func_t,
	matchPositions [HB_MAX_CONTEXT_LENGTH]int) (bool, int, uint8) {
	count := len(input) + 1
	if count > HB_MAX_CONTEXT_LENGTH {
		return false, 0, 0
	}

	buffer := c.buffer

	skippyIter := &c.iter_input
	skippyIter.reset(buffer.idx, count-1)
	skippyIter.matcher.matchFunc = matchFunc
	skippyIter.match_glyph_data = input

	/*
	* This is perhaps the trickiest part of OpenType...  Remarks:
	*
	* - If all components of the ligature were marks, we call this a mark ligature.
	*
	* - If there is no GDEF, and the ligature is NOT a mark ligature, we categorize
	*   it as a ligature glyph.
	*
	* - Ligatures cannot be formed across glyphs attached to different components
	*   of previous ligatures.  Eg. the sequence is LAM,SHADDA,LAM,FATHA,HEH, and
	*   LAM,LAM,HEH form a ligature, leaving SHADDA,FATHA next to eachother.
	*   However, it would be wrong to ligate that SHADDA,FATHA sequence.
	*   There are a couple of exceptions to this:
	*
	*   o If a ligature tries ligating with marks that belong to it itself, go ahead,
	*     assuming that the font designer knows what they are doing (otherwise it can
	*     break Indic stuff when a matra wants to ligate with a conjunct,
	*
	*   o If two marks want to ligate and they belong to different components of the
	*     same ligature glyph, and said ligature glyph is to be ignored according to
	*     mark-filtering rules, then allow.
	*     https://github.com/harfbuzz/harfbuzz/issues/545
	 */

	totalComponentCount := buffer.cur(0).getLigNumComps()

	firstLigId := buffer.cur(0).getLigId()
	firstLigComp := buffer.cur(0).getLigComp()

	const (
		ligbaseNotChecked = iota
		ligbaseMayNotSkip
		ligbaseMaySkip
	)
	ligbase := ligbaseNotChecked
	matchPositions[0] = buffer.idx
	for i := 1; i < count; i++ {
		if !skippyIter.next() {
			return false, 0, 0
		}

		matchPositions[i] = skippyIter.idx

		thisLigId := buffer.Info[skippyIter.idx].getLigId()
		thisLigComp := buffer.Info[skippyIter.idx].getLigComp()

		if firstLigId != 0 && firstLigComp != 0 {
			/* If first component was attached to a previous ligature component,
			* all subsequent components should be attached to the same ligature
			* component, otherwise we shouldn't ligate them... */
			if firstLigId != thisLigId || firstLigComp != thisLigComp {
				/* ...unless, we are attached to a base ligature and that base
				 * ligature is ignorable. */
				if ligbase == ligbaseNotChecked {
					found := false
					out := buffer.outInfo
					j := len(out)
					for j != 0 && out[j-1].getLigId() == firstLigId {
						if out[j-1].getLigComp() == 0 {
							j--
							found = true
							break
						}
						j--
					}

					if found && skippyIter.may_skip(&out[j]) == Yes {
						ligbase = ligbaseMaySkip
					} else {
						ligbase = ligbaseMayNotSkip
					}
				}

				if ligbase == ligbaseMayNotSkip {
					return false, 0, 0
				}
			}
		} else {
			/* If first component was NOT attached to a previous ligature component,
			* all subsequent components should also NOT be attached to any ligature
			* component, unless they are attached to the first component itself! */
			if thisLigId != 0 && thisLigComp != 0 && (thisLigId != firstLigId) {
				return false, 0, 0
			}
		}

		totalComponentCount += buffer.Info[skippyIter.idx].getLigNumComps()
	}

	endOffset := skippyIter.idx - buffer.idx + 1

	return true, endOffset, totalComponentCount
}

// `count` and `matchPositions` include the first glyph
func (c *hb_ot_apply_context_t) ligateInput(count int, matchPositions [HB_MAX_CONTEXT_LENGTH]int,
	matchLength int, ligGlyph fonts.GlyphIndex, totalComponentCount uint8) {
	buffer := c.buffer

	buffer.mergeClusters(buffer.idx, buffer.idx+matchLength)

	/* - If a base and one or more marks ligate, consider that as a base, NOT
	*   ligature, such that all following marks can still attach to it.
	*   https://github.com/harfbuzz/harfbuzz/issues/1109
	*
	* - If all components of the ligature were marks, we call this a mark ligature.
	*   If it *is* a mark ligature, we don't allocate a new ligature id, and leave
	*   the ligature to keep its old ligature id.  This will allow it to attach to
	*   a base ligature in GPOS.  Eg. if the sequence is: LAM,LAM,SHADDA,FATHA,HEH,
	*   and LAM,LAM,HEH for a ligature, they will leave SHADDA and FATHA with a
	*   ligature id and component value of 2.  Then if SHADDA,FATHA form a ligature
	*   later, we don't want them to lose their ligature id/component, otherwise
	*   GPOS will fail to correctly position the mark ligature on top of the
	*   LAM,LAM,HEH ligature.  See:
	*     https://bugzilla.gnome.org/show_bug.cgi?id=676343
	*
	* - If a ligature is formed of components that some of which are also ligatures
	*   themselves, and those ligature components had marks attached to *their*
	*   components, we have to attach the marks to the new ligature component
	*   positions!  Now *that*'s tricky!  And these marks may be following the
	*   last component of the whole sequence, so we should loop forward looking
	*   for them and update them.
	*
	*   Eg. the sequence is LAM,LAM,SHADDA,FATHA,HEH, and the font first forms a
	*   'calt' ligature of LAM,HEH, leaving the SHADDA and FATHA with a ligature
	*   id and component == 1.  Now, during 'liga', the LAM and the LAM-HEH ligature
	*   form a LAM-LAM-HEH ligature.  We need to reassign the SHADDA and FATHA to
	*   the new ligature with a component value of 2.
	*
	*   This in fact happened to a font...  See:
	*   https://bugzilla.gnome.org/show_bug.cgi?id=437633
	 */

	isBaseLigature := buffer.Info[matchPositions[0]].isBaseGlyph()
	isMarkLigature := buffer.Info[matchPositions[0]].isMark()
	for i := 1; i < count; i++ {
		if !buffer.Info[matchPositions[i]].isMark() {
			isBaseLigature = false
			isMarkLigature = false
			break
		}
	}
	isLigature := !isBaseLigature && !isMarkLigature

	klass, ligId := uint16(0), uint8(0)
	if isLigature {
		klass = truetype.Ligature
		ligId = buffer.allocateLigId()
	}
	lastLigId := buffer.cur(0).getLigId()
	lastNumComponents := buffer.cur(0).getLigNumComps()
	componentsSoFar := lastNumComponents

	if isLigature {
		buffer.cur(0).setLigPropsForLigature(ligId, totalComponentCount)
		if buffer.cur(0).unicode.generalCategory() == NonSpacingMark {
			buffer.cur(0).setGeneralCategory(OtherLetter)
		}
	}

	// ReplaceGlyph_with_ligature
	c.setGlyphPropsExt(ligGlyph, klass, true, false)
	buffer.replaceGlyphIndex(ligGlyph)

	for i := 1; i < count; i++ {
		for buffer.idx < matchPositions[i] {
			if isLigature {
				thisComp := buffer.cur(0).getLigComp()
				if thisComp == 0 {
					thisComp = lastNumComponents
				}
				newLigComp := componentsSoFar - lastNumComponents +
					min8(thisComp, lastNumComponents)
				buffer.cur(0).setLigPropsForMark(ligId, newLigComp)
			}
			buffer.nextGlyph()
		}

		lastLigId = buffer.cur(0).getLigId()
		lastNumComponents = buffer.cur(0).getLigNumComps()
		componentsSoFar += lastNumComponents

		/* Skip the base glyph */
		buffer.skipGlyph()
	}

	if !isMarkLigature && lastLigId != 0 {
		/* Re-adjust components for any marks following. */
		for i := buffer.idx; i < len(buffer.Info); i++ {
			if lastLigId != buffer.Info[i].getLigId() {
				break
			}

			thisComp := buffer.Info[i].getLigComp()
			if thisComp == 0 {
				break
			}

			newLigComp := componentsSoFar - lastNumComponents +
				min8(thisComp, lastNumComponents)
			buffer.Info[i].setLigPropsForMark(ligId, newLigComp)
		}
	}
}

func (c *hb_ot_apply_context_t) recurse(subLookupIndex uint16) bool {
	if c.nesting_level_left == 0 || c.recurse_func == nil || c.buffer.max_ops <= 0 {
		if c.buffer.max_ops <= 0 {
			c.buffer.max_ops--
			return false
		}
		c.buffer.max_ops--
	}

	c.nesting_level_left--
	ret := c.recurse_func(c, subLookupIndex)
	c.nesting_level_left++
	return ret
}

// `count` and `matchPositions` include the first glyph
// `lookupRecord` is in design order
func (c *hb_ot_apply_context_t) applyLookup(count int, matchPositions [HB_MAX_CONTEXT_LENGTH]int,
	lookupRecord []truetype.SequenceLookup, matchLength int) {
	buffer := c.buffer
	var end int

	/* All positions are distance from beginning of *output* buffer.
	* Adjust. */
	{
		bl := buffer.backtrackLen()
		end = bl + matchLength

		delta := bl - buffer.idx
		/* Convert positions to new indexing. */
		for j := 0; j < count; j++ {
			matchPositions[j] += delta
		}
	}

	for _, lk := range lookupRecord {
		idx := int(lk.InputIndex)
		if idx >= count {
			continue
		}

		/* Don't recurse to ourself at same position.
		 * Note that this test is too naive, it doesn't catch longer loops. */
		if idx == 0 && lk.LookupIndex == c.lookupIndex {
			continue
		}

		buffer.moveTo(matchPositions[idx])

		if buffer.max_ops <= 0 {
			break
		}

		origLen := buffer.backtrackLen() + buffer.lookaheadLen()
		if !c.recurse(lk.LookupIndex) {
			continue
		}

		newLen := buffer.backtrackLen() + buffer.lookaheadLen()
		delta := newLen - origLen

		if delta == 0 {
			continue
		}

		/* Recursed lookup changed buffer len. Adjust.
		 *
		 * TODO:
		 *
		 * Right now, if buffer length increased by n, we assume n new glyphs
		 * were added right after the current position, and if buffer length
		 * was decreased by n, we assume n match positions after the current
		 * one where removed.  The former (buffer length increased) case is
		 * fine, but the decrease case can be improved in at least two ways,
		 * both of which are significant:
		 *
		 *   - If recursed-to lookup is MultipleSubst and buffer length
		 *     decreased, then it's current match position that was deleted,
		 *     NOT the one after it.
		 *
		 *   - If buffer length was decreased by n, it does not necessarily
		 *     mean that n match positions where removed, as there might
		 *     have been marks and default-ignorables in the sequence.  We
		 *     should instead drop match positions between current-position
		 *     and current-position + n instead.
		 *
		 * It should be possible to construct tests for both of these cases.
		 */

		end += delta
		if end <= int(matchPositions[idx]) {
			/* End might end up being smaller than matchPositions[idx] if the recursed
			* lookup ended up removing many items, more than we have had matched.
			* Just never rewind end back and get out of here.
			* https://bugs.chromium.org/p/chromium/issues/detail?id=659496 */
			end = matchPositions[idx]
			/* There can't be any further changes. */
			break
		}

		next := idx + 1 /* next now is the position after the recursed lookup. */

		if delta > 0 {
			if delta+count > HB_MAX_CONTEXT_LENGTH {
				break
			}
		} else {
			/* NOTE: delta is negative. */
			delta = max(delta, int(next)-int(count))
			next -= delta
		}

		/* Shift! */
		copy(matchPositions[next+delta:], matchPositions[next:count])
		next += delta
		count += delta

		/* Fill in new entries. */
		for j := idx + 1; j < next; j++ {
			matchPositions[j] = matchPositions[j-1] + 1
		}

		/* And fixup the rest. */
		for ; next < count; next++ {
			matchPositions[next] += delta
		}
	}

	buffer.moveTo(end)
}

func (c *hb_ot_apply_context_t) matchBacktrack(backtrack []uint16, matchFunc match_func_t) (bool, int) {
	skippyIter := &c.iter_context
	skippyIter.reset(c.buffer.backtrackLen(), len(backtrack))
	skippyIter.matcher.matchFunc = matchFunc
	skippyIter.match_glyph_data = backtrack

	for i := 0; i < len(backtrack); i++ {
		if !skippyIter.prev() {
			return false, 0
		}
	}

	return true, skippyIter.idx
}

func (c *hb_ot_apply_context_t) matchLookahead(lookahead []uint16, matchFunc match_func_t, offset int) (bool, int) {
	skippyIter := &c.iter_context
	skippyIter.reset(c.buffer.idx+offset-1, len(lookahead))
	skippyIter.matcher.matchFunc = matchFunc
	skippyIter.match_glyph_data = lookahead

	for i := 0; i < len(lookahead); i++ {
		if !skippyIter.next() {
			return false, 0
		}
	}

	return true, skippyIter.idx + 1
}

//  struct LookupRecord
//  {
//    LookupRecord* copy (hb_serialize_context_t *c,
// 			   const hb_map_t         *lookup_map) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 auto *out = c.embed (*this);
// 	 if (unlikely (!out)) return_trace (nil);

// 	 out.lookupListIndex = hb_map_get (lookup_map, lookupListIndex);
// 	 return_trace (out);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return_trace (c.check_struct (this));
//    }

//    HBUINT16	sequenceIndex;		/* Index into current glyph
// 					  * sequence--first glyph = 0 */
//    HBUINT16	lookupListIndex;	/* Lookup to apply to that
// 					  * position--zero--based */
//    public:
//    DEFINE_SIZE_STATIC (4);
//  };

//  template <typename context_t>
//  static inline void recurse_lookups (context_t *c,
// 					 uint lookupCount,
// 					 const LookupRecord lookupRecord[] /* Array of LookupRecords--in design order */)
//  {
//    for (uint i = 0; i < lookupCount; i++)
// 	 c.recurse (lookupRecord[i].lookupListIndex);
//  }

//  /* Contextual lookups */

//  struct ContextClosureLookupContext
//  {
//    ContextClosureFuncs funcs;
//    const void *intersects_data;
//  };

//  struct ContextCollectGlyphsLookupContext
//  {
//    ContextCollectGlyphsFuncs funcs;
//    const void *collect_data;
//  };

//  struct ContextApplyLookupContext
//  {
//    ContextApplyFuncs funcs;
//    const void *match_data;
//  };

//  static inline bool context_intersects (const hb_set_t *glyphs,
// 						uint inputCount, /* Including the first glyph (not matched) */
// 						const HBUINT16 input[], /* Array of input values--start with second glyph */
// 						ContextClosureLookupContext &lookup_context)
//  {
//    return array_is_subset_of (glyphs,
// 				  inputCount ? inputCount - 1 : 0, input,
// 				  lookup_context.funcs.intersects, lookup_context.intersects_data);
//  }

//  static inline void context_closure_lookup (hb_closure_context_t *c,
// 						uint inputCount, /* Including the first glyph (not matched) */
// 						const HBUINT16 input[], /* Array of input values--start with second glyph */
// 						uint lookupCount,
// 						const LookupRecord lookupRecord[],
// 						ContextClosureLookupContext &lookup_context)
//  {
//    if (context_intersects (c.glyphs,
// 			   inputCount, input,
// 			   lookup_context))
// 	 recurse_lookups (c,
// 			  lookupCount, lookupRecord);
//  }

//  static inline void context_collect_glyphs_lookup (hb_collect_glyphs_context_t *c,
// 						   uint inputCount, /* Including the first glyph (not matched) */
// 						   const HBUINT16 input[], /* Array of input values--start with second glyph */
// 						   uint lookupCount,
// 						   const LookupRecord lookupRecord[],
// 						   ContextCollectGlyphsLookupContext &lookup_context)
//  {
//    collect_array (c, c.input,
// 		  inputCount ? inputCount - 1 : 0, input,
// 		  lookup_context.funcs.collect, lookup_context.collect_data);
//    recurse_lookups (c,
// 			lookupCount, lookupRecord);
//  }

//  static inline bool context_would_apply_lookup (hb_would_apply_context_t *c,
// 							uint inputCount, /* Including the first glyph (not matched) */
// 							const HBUINT16 input[], /* Array of input values--start with second glyph */
// 							uint lookupCount HB_UNUSED,
// 							const LookupRecord lookupRecord[] HB_UNUSED,
// 							ContextApplyLookupContext &lookup_context)
//  {
//    return would_match_input (c,
// 				 inputCount, input,
// 				 lookup_context.funcs.match, lookup_context.match_data);
//  }

//  struct Rule
//  {
//    bool intersects (const hb_set_t *glyphs, ContextClosureLookupContext &lookup_context) const
//    {
// 	 return context_intersects (glyphs,
// 					inputCount, inputZ.arrayZ,
// 					lookup_context);
//    }

//    void closure (hb_closure_context_t *c, ContextClosureLookupContext &lookup_context) const
//    {
// 	 if (unlikely (c.lookup_limit_exceeded ())) return;

// 	 const UnsizedArrayOf<LookupRecord> &lookupRecord = StructAfter<UnsizedArrayOf<LookupRecord>>
// 								(inputZ.as_array ((inputCount ? inputCount - 1 : 0)));
// 	 context_closure_lookup (c,
// 				 inputCount, inputZ.arrayZ,
// 				 lookupCount, lookupRecord.arrayZ,
// 				 lookup_context);
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const
//    {
// 	 if (unlikely (c.lookup_limit_exceeded ())) return;

// 	 const UnsizedArrayOf<LookupRecord> &lookupRecord = StructAfter<UnsizedArrayOf<LookupRecord>>
// 								(inputZ.as_array (inputCount ? inputCount - 1 : 0));
// 	 recurse_lookups (c, lookupCount, lookupRecord.arrayZ);
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c,
// 				ContextCollectGlyphsLookupContext &lookup_context) const
//    {
// 	 const UnsizedArrayOf<LookupRecord> &lookupRecord = StructAfter<UnsizedArrayOf<LookupRecord>>
// 								(inputZ.as_array (inputCount ? inputCount - 1 : 0));
// 	 context_collect_glyphs_lookup (c,
// 					inputCount, inputZ.arrayZ,
// 					lookupCount, lookupRecord.arrayZ,
// 					lookup_context);
//    }

//    bool would_apply (hb_would_apply_context_t *c,
// 			 ContextApplyLookupContext &lookup_context) const
//    {
// 	 const UnsizedArrayOf<LookupRecord> &lookupRecord = StructAfter<UnsizedArrayOf<LookupRecord>>
// 								(inputZ.as_array (inputCount ? inputCount - 1 : 0));
// 	 return context_would_apply_lookup (c,
// 						inputCount, inputZ.arrayZ,
// 						lookupCount, lookupRecord.arrayZ,
// 						lookup_context);
//    }

//    bool apply (c *hb_ot_apply_context_t,
// 		   ContextApplyLookupContext &lookup_context) const
//    {
// 	 TRACE_APPLY (this);
// 	 const UnsizedArrayOf<LookupRecord> &lookupRecord = StructAfter<UnsizedArrayOf<LookupRecord>>
// 								(inputZ.as_array (inputCount ? inputCount - 1 : 0));
// 	 return_trace (contextApplyLookup (c, inputCount, inputZ.arrayZ, lookupCount, lookupRecord.arrayZ, lookup_context));
//    }

//    bool serialize (hb_serialize_context_t *c,
// 		   const hb_map_t *input_mapping, /* old.new glyphid or class mapping */
// 		   const hb_map_t *lookup_map) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 auto *out = c.start_embed (this);
// 	 if (unlikely (!c.extend_min (out))) return_trace (false);

// 	 out.inputCount = inputCount;
// 	 out.lookupCount = lookupCount;

// 	 const hb_array_t<const HBUINT16> input = inputZ.as_array (inputCount - 1);
// 	 for (const auto org : input)
// 	 {
// 	   HBUINT16 d;
// 	   d = input_mapping.get (org);
// 	   c.copy (d);
// 	 }

// 	 const UnsizedArrayOf<LookupRecord> &lookupRecord = StructAfter<UnsizedArrayOf<LookupRecord>>
// 								(inputZ.as_array ((inputCount ? inputCount - 1 : 0)));
// 	 for (unsigned i = 0; i < (unsigned) lookupCount; i++)
// 	   c.copy (lookupRecord[i], lookup_map);

// 	 return_trace (true);
//    }

//    bool subset (hb_subset_context_t *c,
// 			const hb_map_t *lookup_map,
// 			const hb_map_t *klass_map = nil) const
//    {
// 	 TRACE_SUBSET (this);

// 	 const hb_array_t<const HBUINT16> input = inputZ.as_array ((inputCount ? inputCount - 1 : 0));
// 	 if (!input.length) return_trace (false);

// 	 const hb_map_t *mapping = klass_map == nil ? c.plan.glyph_map : klass_map;
// 	 if (!hb_all (input, mapping)) return_trace (false);
// 	 return_trace (serialize (c.serializer, mapping, lookup_map));
//    }

//    public:
//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return_trace (inputCount.sanitize (c) &&
// 		   lookupCount.sanitize (c) &&
// 		   c.check_range (inputZ.arrayZ,
// 				   inputZ.item_size * (inputCount ? inputCount - 1 : 0) +
// 				   LookupRecord::static_size * lookupCount));
//    }

//    protected:
//    HBUINT16	inputCount;		/* Total number of glyphs in input
// 					  * glyph sequence--includes the first
// 					  * glyph */
//    HBUINT16	lookupCount;		/* Number of LookupRecords */
//    UnsizedArrayOf<HBUINT16>
// 		 inputZ;			/* Array of match inputs--start with
// 					  * second glyph */
//  /*UnsizedArrayOf<LookupRecord>
// 		 lookupRecordX;*/	/* Array of LookupRecords--in
// 					  * design order */
//    public:
//    DEFINE_SIZE_ARRAY (4, inputZ);
//  };

//  struct RuleSet
//  {
//    bool intersects (const hb_set_t *glyphs,
// 			ContextClosureLookupContext &lookup_context) const
//    {
// 	 return
// 	 + hb_iter (rule)
// 	 | hb_map (hb_add (this))
// 	 | hb_map ([&] (const Rule &_) { return _.intersects (glyphs, lookup_context); })
// 	 | hb_any
// 	 ;
//    }

//    void closure (hb_closure_context_t *c,
// 		 ContextClosureLookupContext &lookup_context) const
//    {
// 	 if (unlikely (c.lookup_limit_exceeded ())) return;

// 	 return
// 	 + hb_iter (rule)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const Rule &_) { _.closure (c, lookup_context); })
// 	 ;
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const
//    {
// 	 if (unlikely (c.lookup_limit_exceeded ())) return;

// 	 return
// 	 + hb_iter (rule)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const Rule &_) { _.closure_lookups (c); })
// 	 ;
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c,
// 				ContextCollectGlyphsLookupContext &lookup_context) const
//    {
// 	 return
// 	 + hb_iter (rule)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const Rule &_) { _.collect_glyphs (c, lookup_context); })
// 	 ;
//    }

//    bool would_apply (hb_would_apply_context_t *c,
// 			 ContextApplyLookupContext &lookup_context) const
//    {
// 	 return
// 	 + hb_iter (rule)
// 	 | hb_map (hb_add (this))
// 	 | hb_map ([&] (const Rule &_) { return _.would_apply (c, lookup_context); })
// 	 | hb_any
// 	 ;
//    }

//    bool apply (c *hb_ot_apply_context_t,
// 		   ContextApplyLookupContext &lookup_context) const
//    {
// 	 TRACE_APPLY (this);
// 	 return_trace (
// 	 + hb_iter (rule)
// 	 | hb_map (hb_add (this))
// 	 | hb_map ([&] (const Rule &_) { return _.apply (c, lookup_context); })
// 	 | hb_any
// 	 )
// 	 ;
//    }

//    bool subset (hb_subset_context_t *c,
// 			const hb_map_t *lookup_map,
// 			const hb_map_t *klass_map = nil) const
//    {
// 	 TRACE_SUBSET (this);

// 	 auto snap = c.serializer.snapshot ();
// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return_trace (false);

// 	 for (const OffsetTo<Rule>& _ : rule)
// 	 {
// 	   if (!_) continue;
// 	   auto *o = out.rule.serialize_append (c.serializer);
// 	   if (unlikely (!o)) continue;

// 	   auto o_snap = c.serializer.snapshot ();
// 	   if (!o.serialize_subset (c, _, this, lookup_map, klass_map))
// 	   {
// 	 out.rule.pop ();
// 	 c.serializer.revert (o_snap);
// 	   }
// 	 }

// 	 bool ret = bool (out.rule);
// 	 if (!ret) c.serializer.revert (snap);

// 	 return_trace (ret);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return_trace (rule.sanitize (c, this));
//    }

//    protected:
//    OffsetArrayOf<Rule>
// 		 rule;			/* Array of Rule tables
// 					  * ordered by preference */
//    public:
//    DEFINE_SIZE_ARRAY (2, rule);
//  };

//  struct ContextFormat1
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 struct ContextClosureLookupContext lookup_context = {
// 	   {intersects_glyph},
// 	   nil
// 	 };

// 	 return
// 	 + hb_zip (this+coverage, ruleSet)
// 	 | hb_filter (*glyphs, hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_map (hb_add (this))
// 	 | hb_map ([&] (const RuleSet &_) { return _.intersects (glyphs, lookup_context); })
// 	 | hb_any
// 	 ;
//    }

//    void closure (hb_closure_context_t *c) const
//    {
// 	 struct ContextClosureLookupContext lookup_context = {
// 	   {intersects_glyph},
// 	   nil
// 	 };

// 	 + hb_zip (this+coverage, ruleSet)
// 	 | hb_filter (*c.glyphs, hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const RuleSet &_) { _.closure (c, lookup_context); })
// 	 ;
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const
//    {
// 	 + hb_iter (ruleSet)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const RuleSet &_) { _.closure_lookups (c); })
// 	 ;
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const {}

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 (this+coverage).collectCoverage (c.input);

// 	 struct ContextCollectGlyphsLookupContext lookup_context = {
// 	   {collect_glyph},
// 	   nil
// 	 };

// 	 + hb_iter (ruleSet)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const RuleSet &_) { _.collect_glyphs (c, lookup_context); })
// 	 ;
//    }

//    bool would_apply (hb_would_apply_context_t *c) const
//    {
// 	 const RuleSet &rule_set = this+ruleSet[(this+coverage).get_coverage (c.glyphs[0])];
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_glyph},
// 	   nil
// 	 };
// 	 return rule_set.would_apply (c, lookup_context);
//    }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (c *hb_ot_apply_context_t) const
//    {
// 	 TRACE_APPLY (this);
// 	 uint index = (this+coverage).get_coverage (c.buffer.cur(0).Codepoint);
// 	 if (likely (index == NOT_COVERED))
// 	   return_trace (false);

// 	 const RuleSet &rule_set = this+ruleSet[index];
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_glyph},
// 	   nil
// 	 };
// 	 return_trace (rule_set.apply (c, lookup_context));
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return_trace (false);
// 	 out.format = format;

// 	 const hb_map_t *lookup_map = c.table_tag == HB_OT_TAG_GSUB ? c.plan.gsub_lookups : c.plan.gpos_lookups;
// 	 hb_sorted_vector_t<hb_codepoint_t> new_coverage;
// 	 + hb_zip (this+coverage, ruleSet)
// 	 | hb_filter (glyphset, hb_first)
// 	 | hb_filter (subset_offset_array (c, out.ruleSet, this, lookup_map), hb_second)
// 	 | hb_map (hb_first)
// 	 | hb_map (glyph_map)
// 	 | hb_sink (new_coverage)
// 	 ;

// 	 out.coverage.serialize (c.serializer, out)
// 		  .serialize (c.serializer, new_coverage.iter ());
// 	 return_trace (bool (new_coverage));
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return_trace (coverage.sanitize (c, this) && ruleSet.sanitize (c, this));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    OffsetTo<Coverage>
// 		 coverage;		/* Offset to Coverage table--from
// 					  * beginning of table */
//    OffsetArrayOf<RuleSet>
// 		 ruleSet;		/* Array of RuleSet tables
// 					  * ordered by Coverage Index */
//    public:
//    DEFINE_SIZE_ARRAY (6, ruleSet);
//  };

//  struct ContextFormat2
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 if (!(this+coverage).intersects (glyphs))
// 	   return false;

// 	 const ClassDef &class_def = this+classDef;

// 	 struct ContextClosureLookupContext lookup_context = {
// 	   {intersects_class},
// 	   &class_def
// 	 };

// 	 return
// 	 + hb_iter (ruleSet)
// 	 | hb_map (hb_add (this))
// 	 | hb_enumerate
// 	 | hb_map ([&] (const hb_pair_t<unsigned, const RuleSet &> p)
// 		   { return class_def.intersects_class (glyphs, p.first) &&
// 				p.second.intersects (glyphs, lookup_context); })
// 	 | hb_any
// 	 ;
//    }

//    void closure (hb_closure_context_t *c) const
//    {
// 	 if (!(this+coverage).intersects (c.glyphs))
// 	   return;

// 	 const ClassDef &class_def = this+classDef;

// 	 struct ContextClosureLookupContext lookup_context = {
// 	   {intersects_class},
// 	   &class_def
// 	 };

// 	 return
// 	 + hb_enumerate (ruleSet)
// 	 | hb_filter ([&] (unsigned _)
// 		  { return class_def.intersects_class (c.glyphs, _); },
// 		  hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const RuleSet &_) { _.closure (c, lookup_context); })
// 	 ;
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const
//    {
// 	 + hb_iter (ruleSet)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const RuleSet &_) { _.closure_lookups (c); })
// 	 ;
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const {}

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 (this+coverage).collectCoverage (c.input);

// 	 const ClassDef &class_def = this+classDef;
// 	 struct ContextCollectGlyphsLookupContext lookup_context = {
// 	   {collect_class},
// 	   &class_def
// 	 };

// 	 + hb_iter (ruleSet)
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const RuleSet &_) { _.collect_glyphs (c, lookup_context); })
// 	 ;
//    }

//    bool would_apply (hb_would_apply_context_t *c) const
//    {
// 	 const ClassDef &class_def = this+classDef;
// 	 uint index = class_def.get_class (c.glyphs[0]);
// 	 const RuleSet &rule_set = this+ruleSet[index];
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_class},
// 	   &class_def
// 	 };
// 	 return rule_set.would_apply (c, lookup_context);
//    }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (c *hb_ot_apply_context_t) const
//    {
// 	 TRACE_APPLY (this);
// 	 uint index = (this+coverage).get_coverage (c.buffer.cur(0).Codepoint);
// 	 if (likely (index == NOT_COVERED)) return_trace (false);

// 	 const ClassDef &class_def = this+classDef;
// 	 index = class_def.get_class (c.buffer.cur(0).Codepoint);
// 	 const RuleSet &rule_set = this+ruleSet[index];
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_class},
// 	   &class_def
// 	 };
// 	 return_trace (rule_set.apply (c, lookup_context));
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return_trace (false);
// 	 out.format = format;
// 	 if (unlikely (!out.coverage.serialize_subset (c, coverage, this)))
// 	   return_trace (false);

// 	 hb_map_t klass_map;
// 	 out.classDef.serialize_subset (c, classDef, this, &klass_map);

// 	 const hb_map_t *lookup_map = c.table_tag == HB_OT_TAG_GSUB ? c.plan.gsub_lookups : c.plan.gpos_lookups;
// 	 bool ret = true;
// 	 int non_zero_index = 0, index = 0;
// 	 for (const hb_pair_t<unsigned, const OffsetTo<RuleSet>&> _ : + hb_enumerate (ruleSet)
// 								  | hb_filter (klass_map, hb_first))
// 	 {
// 	   auto *o = out.ruleSet.serialize_append (c.serializer);
// 	   if (unlikely (!o))
// 	   {
// 	 ret = false;
// 	 break;
// 	   }

// 	   if (o.serialize_subset (c, _.second, this, lookup_map, &klass_map))
// 	 non_zero_index = index;

// 	   index++;
// 	 }

// 	 if (!ret) return_trace (ret);

// 	 //prune empty trailing ruleSets
// 	 --index;
// 	 while (index > non_zero_index)
// 	 {
// 	   out.ruleSet.pop ();
// 	   index--;
// 	 }

// 	 return_trace (bool (out.ruleSet));
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return_trace (coverage.sanitize (c, this) && classDef.sanitize (c, this) && ruleSet.sanitize (c, this));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 2 */
//    OffsetTo<Coverage>
// 		 coverage;		/* Offset to Coverage table--from
// 					  * beginning of table */
//    OffsetTo<ClassDef>
// 		 classDef;		/* Offset to glyph ClassDef table--from
// 					  * beginning of table */
//    OffsetArrayOf<RuleSet>
// 		 ruleSet;		/* Array of RuleSet tables
// 					  * ordered by class */
//    public:
//    DEFINE_SIZE_ARRAY (8, ruleSet);
//  };

//  struct ContextFormat3
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 if (!(this+coverageZ[0]).intersects (glyphs))
// 	   return false;

// 	 struct ContextClosureLookupContext lookup_context = {
// 	   {intersects_coverage},
// 	   this
// 	 };
// 	 return context_intersects (glyphs,
// 					glyphCount, (const HBUINT16 *) (coverageZ.arrayZ + 1),
// 					lookup_context);
//    }

//    void closure (hb_closure_context_t *c) const
//    {
// 	 if (!(this+coverageZ[0]).intersects (c.glyphs))
// 	   return;

// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 struct ContextClosureLookupContext lookup_context = {
// 	   {intersects_coverage},
// 	   this
// 	 };
// 	 context_closure_lookup (c,
// 				 glyphCount, (const HBUINT16 *) (coverageZ.arrayZ + 1),
// 				 lookupCount, lookupRecord,
// 				 lookup_context);
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const
//    {
// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 recurse_lookups (c, lookupCount, lookupRecord);
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const {}

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 (this+coverageZ[0]).collectCoverage (c.input);

// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 struct ContextCollectGlyphsLookupContext lookup_context = {
// 	   {collectCoverage},
// 	   this
// 	 };

// 	 context_collect_glyphs_lookup (c,
// 					glyphCount, (const HBUINT16 *) (coverageZ.arrayZ + 1),
// 					lookupCount, lookupRecord,
// 					lookup_context);
//    }

//    bool would_apply (hb_would_apply_context_t *c) const
//    {
// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_coverage},
// 	   this
// 	 };
// 	 return context_would_apply_lookup (c,
// 						glyphCount, (const HBUINT16 *) (coverageZ.arrayZ + 1),
// 						lookupCount, lookupRecord,
// 						lookup_context);
//    }

//    const Coverage &get_coverage () const { return this+coverageZ[0]; }

//    bool apply (c *hb_ot_apply_context_t) const
//    {
// 	 TRACE_APPLY (this);
// 	 uint index = (this+coverageZ[0]).get_coverage (c.buffer.cur(0).Codepoint);
// 	 if (likely (index == NOT_COVERED)) return_trace (false);

// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_coverage},
// 	   this
// 	 };
// 	 return_trace (contextApplyLookup (c, glyphCount, (const HBUINT16 *) (coverageZ.arrayZ + 1), lookupCount, lookupRecord, lookup_context));
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 auto *out = c.serializer.start_embed (this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return_trace (false);

// 	 out.format = format;
// 	 out.glyphCount = glyphCount;
// 	 out.lookupCount = lookupCount;

// 	 auto coverages = coverageZ.as_array (glyphCount);

// 	 for (const OffsetTo<Coverage>& offset : coverages)
// 	 {
// 	   auto *o = c.serializer.allocate_size<OffsetTo<Coverage>> (OffsetTo<Coverage>::static_size);
// 	   if (unlikely (!o)) return_trace (false);
// 	   if (!o.serialize_subset (c, offset, this)) return_trace (false);
// 	 }

// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 const hb_map_t *lookup_map = c.table_tag == HB_OT_TAG_GSUB ? c.plan.gsub_lookups : c.plan.gpos_lookups;
// 	 for (unsigned i = 0; i < (unsigned) lookupCount; i++)
// 	   c.serializer.copy (lookupRecord[i], lookup_map);

// 	 return_trace (true);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 if (!c.check_struct (this)) return_trace (false);
// 	 uint count = glyphCount;
// 	 if (!count) return_trace (false); /* We want to access coverageZ[0] freely. */
// 	 if (!c.check_array (coverageZ.arrayZ, count)) return_trace (false);
// 	 for (uint i = 0; i < count; i++)
// 	   if (!coverageZ[i].sanitize (c, this)) return_trace (false);
// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 return_trace (c.check_array (lookupRecord, lookupCount));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 3 */
//    HBUINT16	glyphCount;		/* Number of glyphs in the input glyph
// 					  * sequence */
//    HBUINT16	lookupCount;		/* Number of LookupRecords */
//    UnsizedArrayOf<OffsetTo<Coverage>>
// 		 coverageZ;		/* Array of offsets to Coverage
// 					  * table in glyph sequence order */
//  /*UnsizedArrayOf<LookupRecord>
// 		 lookupRecordX;*/	/* Array of LookupRecords--in
// 					  * design order */
//    public:
//    DEFINE_SIZE_ARRAY (6, coverageZ);
//  };

//  struct Context
//  {
//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return_trace (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return_trace (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 case 2: return_trace (c.dispatch (u.format2, hb_forward<Ts> (ds)...));
// 	 case 3: return_trace (c.dispatch (u.format3, hb_forward<Ts> (ds)...));
// 	 default:return_trace (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    ContextFormat1	format1;
//    ContextFormat2	format2;
//    ContextFormat3	format3;
//    } u;
//  };

//  /* Chaining Contextual lookups */

//  struct ChainContextClosureLookupContext
//  {
//    ContextClosureFuncs funcs;
//    const void *intersects_data[3];
//  };

//  struct ChainContextCollectGlyphsLookupContext
//  {
//    ContextCollectGlyphsFuncs funcs;
//    const void *collect_data[3];
//  };

//  struct ChainContextApplyLookupContext
//  {
//    ContextApplyFuncs funcs;
//    const void *match_data[3];
//  };

//  static inline bool chain_context_intersects (const hb_set_t *glyphs,
// 						  uint backtrackCount,
// 						  const HBUINT16 backtrack[],
// 						  uint inputCount, /* Including the first glyph (not matched) */
// 						  const HBUINT16 input[], /* Array of input values--start with second glyph */
// 						  uint lookaheadCount,
// 						  const HBUINT16 lookahead[],
// 						  ChainContextClosureLookupContext &lookup_context)
//  {
//    return array_is_subset_of (glyphs,
// 				  backtrackCount, backtrack,
// 				  lookup_context.funcs.intersects, lookup_context.intersects_data[0])
// 	   && array_is_subset_of (glyphs,
// 				  inputCount ? inputCount - 1 : 0, input,
// 				  lookup_context.funcs.intersects, lookup_context.intersects_data[1])
// 	   && array_is_subset_of (glyphs,
// 				  lookaheadCount, lookahead,
// 				  lookup_context.funcs.intersects, lookup_context.intersects_data[2]);
//  }

//  static inline void chain_context_closure_lookup (hb_closure_context_t *c,
// 						  uint backtrackCount,
// 						  const HBUINT16 backtrack[],
// 						  uint inputCount, /* Including the first glyph (not matched) */
// 						  const HBUINT16 input[], /* Array of input values--start with second glyph */
// 						  uint lookaheadCount,
// 						  const HBUINT16 lookahead[],
// 						  uint lookupCount,
// 						  const LookupRecord lookupRecord[],
// 						  ChainContextClosureLookupContext &lookup_context)
//  {
//    if (chain_context_intersects (c.glyphs,
// 				 backtrackCount, backtrack,
// 				 inputCount, input,
// 				 lookaheadCount, lookahead,
// 				 lookup_context))
// 	 recurse_lookups (c,
// 			  lookupCount, lookupRecord);
//  }

//  static inline void chain_context_collect_glyphs_lookup (hb_collect_glyphs_context_t *c,
// 							 uint backtrackCount,
// 							 const HBUINT16 backtrack[],
// 							 uint inputCount, /* Including the first glyph (not matched) */
// 							 const HBUINT16 input[], /* Array of input values--start with second glyph */
// 							 uint lookaheadCount,
// 							 const HBUINT16 lookahead[],
// 							 uint lookupCount,
// 							 const LookupRecord lookupRecord[],
// 							 ChainContextCollectGlyphsLookupContext &lookup_context)
//  {
//    collect_array (c, c.before,
// 		  backtrackCount, backtrack,
// 		  lookup_context.funcs.collect, lookup_context.collect_data[0]);
//    collect_array (c, c.input,
// 		  inputCount ? inputCount - 1 : 0, input,
// 		  lookup_context.funcs.collect, lookup_context.collect_data[1]);
//    collect_array (c, c.after,
// 		  lookaheadCount, lookahead,
// 		  lookup_context.funcs.collect, lookup_context.collect_data[2]);
//    recurse_lookups (c,
// 			lookupCount, lookupRecord);
//  }

//  static inline bool chain_context_would_apply_lookup (hb_would_apply_context_t *c,
// 							  uint backtrackCount,
// 							  const HBUINT16 backtrack[] HB_UNUSED,
// 							  uint inputCount, /* Including the first glyph (not matched) */
// 							  const HBUINT16 input[], /* Array of input values--start with second glyph */
// 							  uint lookaheadCount,
// 							  const HBUINT16 lookahead[] HB_UNUSED,
// 							  uint lookupCount HB_UNUSED,
// 							  const LookupRecord lookupRecord[] HB_UNUSED,
// 							  ChainContextApplyLookupContext &lookup_context)
//  {
//    return (c.zero_context ? !backtrackCount && !lookaheadCount : true)
// 	   && would_match_input (c,
// 				 inputCount, input,
// 				 lookup_context.funcs.match, lookup_context.match_data[1]);
//  }

//  struct ChainContext
//  {
//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return_trace (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return_trace (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 case 2: return_trace (c.dispatch (u.format2, hb_forward<Ts> (ds)...));
// 	 case 3: return_trace (c.dispatch (u.format3, hb_forward<Ts> (ds)...));
// 	 default:return_trace (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;	/* Format identifier */
//    ChainContextFormat1	format1;
//    ChainContextFormat2	format2;
//    ChainContextFormat3	format3;
//    } u;
//  };

//  template <typename T>
//  struct ExtensionFormat1
//  {
//    uint get_type () const { return extensionLookupType; }

//    template <typename X>
//    const X& get_subtable () const
//    { return this + reinterpret_cast<const LOffsetTo<typename T::SubTable> &> (extensionOffset); }

//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, format);
// 	 if (unlikely (!c.may_dispatch (this, this))) return_trace (c.no_dispatch_return_value ());
// 	 return_trace (get_subtable<typename T::SubTable> ().dispatch (c, get_type (), hb_forward<Ts> (ds)...));
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    { dispatch (c); }

//    /* This is called from may_dispatch() above with hb_sanitize_context_t. */
//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return_trace (c.check_struct (this) &&
// 		   extensionLookupType != T::SubTable::Extension);
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier. Set to 1. */
//    HBUINT16	extensionLookupType;	/* Lookup type of subtable referenced
// 					  * by ExtensionOffset (i.e. the
// 					  * extension subtable). */
//    Offset32	extensionOffset;	/* Offset to the extension subtable,
// 					  * of lookup type subtable. */
//    public:
//    DEFINE_SIZE_STATIC (8);
//  };

//  template <typename T>
//  struct Extension
//  {
//    uint get_type () const
//    {
// 	 switch (u.format) {
// 	 case 1: return u.format1.get_type ();
// 	 default:return 0;
// 	 }
//    }
//    template <typename X>
//    const X& get_subtable () const
//    {
// 	 switch (u.format) {
// 	 case 1: return u.format1.template get_subtable<typename T::SubTable> ();
// 	 default:return Null (typename T::SubTable);
// 	 }
//    }

//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return_trace (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return_trace (u.format1.dispatch (c, hb_forward<Ts> (ds)...));
// 	 default:return_trace (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    ExtensionFormat1<T>	format1;
//    } u;
//  };
