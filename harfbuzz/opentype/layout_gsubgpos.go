package opentype

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-layout-gsubgpos.hh Copyright © 2007,2008,2009,2010  Red Hat, Inc. 2010,2012  Google, Inc.  Behdad Esfahbod

type TLookup interface {
	// accumulate the sibtables coverage into the diggest
	collect_coverage(*cm.SetDigest)
	// walk the subtables to add them to the context
	dispatch(*hb_get_subtables_context_t)
}

/*
 * GSUB/GPOS Common
 */

const IgnoreFlags = uint16(truetype.IgnoreBaseGlyphs | truetype.IgnoreLigatures | truetype.IgnoreMarks)

type hb_ot_layout_lookup_accelerator_t struct {
	digest    cm.SetDigest // union of all the subtables coverage
	subtables hb_get_subtables_context_t
}

func (ac *hb_ot_layout_lookup_accelerator_t) init(lookup TLookup) {
	ac.digest = cm.SetDigest{}
	lookup.collect_coverage(&ac.digest)
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

func collect_coverage(dst *cm.SetDigest, cov truetype.Coverage) {
	switch cov := cov.(type) {
	case truetype.CoverageList:
		dst.AddArray(cov)
	case truetype.CoverageRanges:
		for _, r := range cov {
			dst.AddRange(r.Start, r.End)
		}
	}
}

// implements TLookup
type lookupGSUB truetype.LookupGSUB

func (l lookupGSUB) collect_coverage(dst *cm.SetDigest) {
	for _, table := range l.Subtables {
		collect_coverage(dst, table.Coverage)
	}
}

func (l lookupGSUB) dispatch(ctx *hb_get_subtables_context_t) {
	for _, table := range l.Subtables {
		*ctx = append(*ctx, new_hb_applicable_t(table))
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

type hb_apply_func_t interface {
	apply(c *hb_ot_apply_context_t) bool
	get_coverage() truetype.Coverage
}

// represents one layout subtable, with its own coverage
type hb_applicable_t struct {
	digest cm.SetDigest
	obj    hb_apply_func_t
}

func new_hb_applicable_t(table truetype.LookupGSUBSubtable) hb_applicable_t {
	ap := hb_applicable_t{obj: table.Data}
	collect_coverage(&ap.digest, table.Coverage)
	return ap
}

func (ap hb_applicable_t) apply(c *hb_ot_apply_context_t) bool {
	return ap.digest.MayHave(c.buffer.Cur(0).Codepoint) && ap.obj.apply(c)

}

type hb_get_subtables_context_t []hb_applicable_t

//  {
//    template <typename Type>
//    static inline bool apply_to (const void *obj, OT::hb_ot_apply_context_t *c)
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
//    typedef return_t (*recurse_func_t) (hb_closure_context_t *c, uint lookup_index);
//    template <typename T>
//    return_t dispatch (const T &obj) { obj.closure (this); return hb_empty_t (); }
//    static return_t default_return_value () { return hb_empty_t (); }
//    void recurse (uint lookup_index)
//    {
// 	 if (unlikely (nesting_level_left == 0 || !recurse_func))
// 	   return;

// 	 nesting_level_left--;
// 	 recurse_func (this, lookup_index);
// 	 nesting_level_left++;
//    }

//    bool lookup_limit_exceeded ()
//    { return lookup_count > HB_MAX_LOOKUP_INDICES; }

//    bool should_visit_lookup (uint lookup_index)
//    {
// 	 if (lookup_count++ > HB_MAX_LOOKUP_INDICES)
// 	   return false;

// 	 if (is_lookup_done (lookup_index))
// 	   return false;

// 	 done_lookups.set (lookup_index, glyphs.get_population ());
// 	 return true;
//    }

//    bool is_lookup_done (uint lookup_index)
//    {
// 	 if (done_lookups.in_error ())
// 	   return true;

// 	 /* Have we visited this lookup with the current set of glyphs? */
// 	 return done_lookups.get (lookup_index) == glyphs.get_population ();
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
// 			   recurse_func (nullptr),
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
//    typedef return_t (*recurse_func_t) (hb_closure_lookups_context_t *c, unsigned lookup_index);
//    template <typename T>
//    return_t dispatch (const T &obj) { obj.closure_lookups (this); return hb_empty_t (); }
//    static return_t default_return_value () { return hb_empty_t (); }
//    void recurse (unsigned lookup_index)
//    {
// 	 if (unlikely (nesting_level_left == 0 || !recurse_func))
// 	   return;

// 	 /* Return if new lookup was recursed to before. */
// 	 if (is_lookup_visited (lookup_index))
// 	   return;

// 	 set_lookup_visited (lookup_index);
// 	 nesting_level_left--;
// 	 recurse_func (this, lookup_index);
// 	 nesting_level_left++;
//    }

//    void set_lookup_visited (unsigned lookup_index)
//    { visited_lookups.add (lookup_index); }

//    void set_lookup_inactive (unsigned lookup_index)
//    { inactive_lookups.add (lookup_index); }

//    bool lookup_limit_exceeded ()
//    { return lookup_count > HB_MAX_LOOKUP_INDICES; }

//    bool is_lookup_visited (unsigned lookup_index)
//    {
// 	 if (lookup_count++ > HB_MAX_LOOKUP_INDICES)
// 	   return true;

// 	 if (visited_lookups.in_error ())
// 	   return true;

// 	 return visited_lookups.has (lookup_index);
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
// 				 recurse_func (nullptr),
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
//    typedef return_t (*recurse_func_t) (hb_collect_glyphs_context_t *c, uint lookup_index);
//    template <typename T>
//    return_t dispatch (const T &obj) { obj.collect_glyphs (this); return hb_empty_t (); }
//    static return_t default_return_value () { return hb_empty_t (); }
//    void recurse (uint lookup_index)
//    {
// 	 if (unlikely (nesting_level_left == 0 || !recurse_func))
// 	   return;

// 	 /* Note that GPOS sets recurse_func to nullptr already, so it doesn't get
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
// 	 if (recursed_lookups.has (lookup_index))
// 	   return;

// 	 hb_set_t *old_before = before;
// 	 hb_set_t *old_input  = input;
// 	 hb_set_t *old_after  = after;
// 	 before = input = after = hb_set_get_empty ();

// 	 nesting_level_left--;
// 	 recurse_func (this, lookup_index);
// 	 nesting_level_left++;

// 	 before = old_before;
// 	 input  = old_input;
// 	 after  = old_after;

// 	 recursed_lookups.add (lookup_index);
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
// 				   recurse_func (nullptr),
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
// 	 r.collect_coverage (set);
// 	 return false;
//    }

//    hb_collect_coverage_context_t (set_t *set_) :
// 					set (set_) {}

//    set_t *set;
//  };

type match_func_t = func(gid fonts.GlyphIndex, value *uint16)

type hb_ot_apply_context_matcher_t struct {
	lookup_props uint
	ignore_zwnj  bool
	ignore_zwj   bool
	mask         cm.Mask
	syllable     uint8
	match_func   match_func_t
	// const void *match_data;
}

type hb_ot_apply_context_t struct {
	//    skipping_iterator_t iter_input, iter_context;

	font   *cm.Font
	face   cm.Face
	buffer *cm.Buffer
	//    recurse_func_t recurse_func;
	//    const GDEF &gdef;
	//    const VariationStore &var_store;

	direction          cm.Direction
	lookup_mask        cm.Mask
	table_index        int /* GSUB/GPOS */
	lookup_index       int
	lookup_props       uint32
	nesting_level_left int

	has_glyph_classes bool
	auto_zwnj         bool
	auto_zwj          bool
	random            bool

	random_state uint32
}

func new_hb_ot_apply_context_t(table_index int, font *cm.Font, buffer *cm.Buffer) hb_ot_apply_context_t {
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
	out.lookup_index = -1 // TODO: check
	out.nesting_level_left = HB_MAX_NESTING_LEVEL
	out.has_glyph_classes = gdef.has_glyph_classes()
	out.auto_zwnj = true
	out.auto_zwj = true
	out.random_state = 1

	out.init_iters()
	return out
}

func (c *hb_ot_apply_context_t) init_iters() {
	c.iter_input.init(c, false)
	c.iter_context.init(c, true)
}

func (c *hb_ot_apply_context_t) set_lookup_mask(mask cm.Mask) {
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
	c.lookup_props = lookupProps
	c.init_iters()
}

func (c *hb_ot_apply_context_t) hb_ot_layout_substitute_lookup(lookup lookupGSUB, accel *hb_ot_layout_lookup_accelerator_t) {
	c.apply_string(lookup, accel)
}

func (c *hb_ot_apply_context_t) check_glyph_property(info *cm.GlyphInfo, matchProps uint32) bool {
	glyph := info.Codepoint
	glyphProps := info.GlyphProps

	/* Not covered, if, for example, glyph class is ligature and
	 * matchProps includes LookupFlags::IgnoreLigatures */
	if (glyphProps & uint16(matchProps) & IgnoreFlags) != 0 {
		return false
	}

	if glyphProps&truetype.Mark != 0 {
		return c.matchPropertiesMark(glyph, glyphProps, matchProps)
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

// 		hb_dispatch_context_t<hb_ot_apply_context_t, bool, HB_DEBUG_APPLY>
//  {
//    struct matcher_t
//    {
// 	 matcher_t () :
// 		  lookup_props (0),
// 		  ignore_zwnj (false),
// 		  ignore_zwj (false),
// 		  mask (-1),
//  #define arg1(arg) (arg) /* Remove the macro to see why it's needed! */
// 		  syllable arg1(0),
//  #undef arg1
// 		  match_func (nullptr),
// 		  match_data (nullptr) {}

// 	 void set_ignore_zwnj (bool ignore_zwnj_) { ignore_zwnj = ignore_zwnj_; }
// 	 void set_ignore_zwj (bool ignore_zwj_) { ignore_zwj = ignore_zwj_; }
// 	 void set_lookup_props (uint lookup_props_) { lookup_props = lookup_props_; }
// 	 void set_mask (cm.Mask mask_) { mask = mask_; }
// 	 void set_syllable (uint8_t syllable_)  { syllable = syllable_; }
// 	 void set_match_func (match_func_t match_func_,
// 			  const void *match_data_)
// 	 { match_func = match_func_; match_data = match_data_; }

// 	 enum may_match_t {
// 	   MATCH_NO,
// 	   MATCH_YES,
// 	   MATCH_MAYBE
// 	 };

// 	 may_match_t may_match (const GlyphInfo &info,
// 				const HBUINT16        *glyph_data) const
// 	 {
// 	   if (!(info.mask & mask) ||
// 	   (syllable && syllable != info.syllable ()))
// 	 return MATCH_NO;

// 	   if (match_func)
// 	 return match_func (info.Codepoint, *glyph_data, match_data) ? MATCH_YES : MATCH_NO;

// 	   return MATCH_MAYBE;
// 	 }

// 	 enum may_skip_t {
// 	   SKIP_NO,
// 	   SKIP_YES,
// 	   SKIP_MAYBE
// 	 };

// 	 may_skip_t may_skip (const hb_ot_apply_context_t *c,
// 			  const GlyphInfo       &info) const
// 	 {
// 	   if (!c.check_glyph_property (&info, lookup_props))
// 	 return SKIP_YES;

// 	   if (unlikely (_hb_glyph_info_is_default_ignorable_and_not_hidden (&info) &&
// 			 (ignore_zwnj || !_hb_glyph_info_is_zwnj (&info)) &&
// 			 (ignore_zwj || !_hb_glyph_info_is_zwj (&info))))
// 	 return SKIP_MAYBE;

// 	   return SKIP_NO;
// 	 }

//    struct skipping_iterator_t
//    {
// 	 void init (hb_ot_apply_context_t *c_, bool context_match = false)
// 	 {
// 	   c = c_;
// 	   match_glyph_data = nullptr;
// 	   matcher.set_match_func (nullptr, nullptr);
// 	   matcher.set_lookup_props (c.lookup_props);
// 	   /* Ignore ZWNJ if we are matching GPOS, or matching GSUB context and asked to. */
// 	   matcher.set_ignore_zwnj (c.table_index == 1 || (context_match && c.auto_zwnj));
// 	   /* Ignore ZWJ if we are matching context, or asked to. */
// 	   matcher.set_ignore_zwj  (context_match || c.auto_zwj);
// 	   matcher.set_mask (context_match ? -1 : c.lookup_mask);
// 	 }
// 	 void set_lookup_props (uint lookup_props)
// 	 {
// 	   matcher.set_lookup_props (lookup_props);
// 	 }
// 	 void set_match_func (matcher_t::match_func_t match_func_,
// 			  const void *match_data_,
// 			  const HBUINT16 glyph_data[])
// 	 {
// 	   matcher.set_match_func (match_func_, match_data_);
// 	   match_glyph_data = glyph_data;
// 	 }

// 	 void reset (uint start_index_,
// 		 uint num_items_)
// 	 {
// 	   idx = start_index_;
// 	   num_items = num_items_;
// 	   end = c.buffer.len;
// 	   matcher.set_syllable (start_index_ == c.buffer.idx ? c.buffer.Cur().syllable () : 0);
// 	 }

// 	 void reject ()
// 	 {
// 	   num_items++;
// 	   if (match_glyph_data) match_glyph_data--;
// 	 }

// 	 matcher_t::may_skip_t
// 	 may_skip (const GlyphInfo &info) const
// 	 { return matcher.may_skip (c, info); }

// 	 bool next ()
// 	 {
// 	   assert (num_items > 0);
// 	   while (idx + num_items < end)
// 	   {
// 	 idx++;
// 	 const GlyphInfo &info = c.buffer.info[idx];

// 	 matcher_t::may_skip_t skip = matcher.may_skip (c, info);
// 	 if (unlikely (skip == matcher_t::SKIP_YES))
// 	   continue;

// 	 matcher_t::may_match_t match = matcher.may_match (info, match_glyph_data);
// 	 if (match == matcher_t::MATCH_YES ||
// 		 (match == matcher_t::MATCH_MAYBE &&
// 		  skip == matcher_t::SKIP_NO))
// 	 {
// 	   num_items--;
// 	   if (match_glyph_data) match_glyph_data++;
// 	   return true;
// 	 }

// 	 if (skip == matcher_t::SKIP_NO)
// 	   return false;
// 	   }
// 	   return false;
// 	 }
// 	 bool prev ()
// 	 {
// 	   assert (num_items > 0);
// 	   while (idx > num_items - 1)
// 	   {
// 	 idx--;
// 	 const GlyphInfo &info = c.buffer.out_info[idx];

// 	 matcher_t::may_skip_t skip = matcher.may_skip (c, info);
// 	 if (unlikely (skip == matcher_t::SKIP_YES))
// 	   continue;

// 	 matcher_t::may_match_t match = matcher.may_match (info, match_glyph_data);
// 	 if (match == matcher_t::MATCH_YES ||
// 		 (match == matcher_t::MATCH_MAYBE &&
// 		  skip == matcher_t::SKIP_NO))
// 	 {
// 	   num_items--;
// 	   if (match_glyph_data) match_glyph_data++;
// 	   return true;
// 	 }

// 	 if (skip == matcher_t::SKIP_NO)
// 	   return false;
// 	   }
// 	   return false;
// 	 }

// 	 uint idx;
// 	 protected:
// 	 hb_ot_apply_context_t *c;
// 	 matcher_t matcher;
// 	 const HBUINT16 *match_glyph_data;

// 	 uint num_items;
// 	 uint end;
//    };

//    const char *get_name () { return "APPLY"; }
//    typedef return_t (*recurse_func_t) (hb_ot_apply_context_t *c, uint lookup_index);
//    template <typename T>
//    return_t dispatch (const T &obj) { return obj.apply (this); }
//    static return_t default_return_value () { return false; }
//    bool stop_sublookup_iteration (return_t r) const { return r; }
//    return_t recurse (uint sub_lookup_index)
//    {
// 	 if (unlikely (nesting_level_left == 0 || !recurse_func || buffer.max_ops-- <= 0))
// 	   return default_return_value ();

// 	 nesting_level_left--;
// 	 bool ret = recurse_func (this, sub_lookup_index);
// 	 nesting_level_left++;
// 	 return ret;
//    }

//    void set_random (bool random_) { random = random_; }
//    void set_recurse_func (recurse_func_t func) { recurse_func = func; }
//    void set_lookup_index (uint lookup_index_) { lookup_index = lookup_index_; }

//    uint32_t random_number ()
//    {
// 	 /* http://www.cplusplus.com/reference/random/minstd_rand/ */
// 	 random_state = random_state * 48271 % 2147483647;
// 	 return random_state;
//    }

//    void _set_glyph_props (hb_codepoint_t glyph_index,
// 			   uint class_guess = 0,
// 			   bool ligature = false,
// 			   bool component = false) const
//    {
// 	 uint add_in = _hb_glyph_info_get_glyph_props (&buffer.Cur()) &
// 			   HB_OT_LAYOUT_GLYPH_PROPS_PRESERVE;
// 	 add_in |= HB_OT_LAYOUT_GLYPH_PROPS_SUBSTITUTED;
// 	 if (ligature)
// 	 {
// 	   add_in |= HB_OT_LAYOUT_GLYPH_PROPS_LIGATED;
// 	   /* In the only place that the MULTIPLIED bit is used, Uniscribe
// 		* seems to only care about the "last" transformation between
// 		* Ligature and Multiple substitutions.  Ie. if you ligate, expand,
// 		* and ligate again, it forgives the multiplication and acts as
// 		* if only ligation happened.  As such, clear MULTIPLIED bit.
// 		*/
// 	   add_in &= ~HB_OT_LAYOUT_GLYPH_PROPS_MULTIPLIED;
// 	 }
// 	 if (component)
// 	   add_in |= HB_OT_LAYOUT_GLYPH_PROPS_MULTIPLIED;
// 	 if (likely (has_glyph_classes))
// 	   _hb_glyph_info_set_glyph_props (&buffer.Cur(), add_in | gdef.get_glyph_props (glyph_index));
// 	 else if (class_guess)
// 	   _hb_glyph_info_set_glyph_props (&buffer.Cur(), add_in | class_guess);
//    }

//    void replace_glyph (hb_codepoint_t glyph_index) const
//    {
// 	 _set_glyph_props (glyph_index);
// 	 buffer.replace_glyph (glyph_index);
//    }
//    void replace_glyph_inplace (hb_codepoint_t glyph_index) const
//    {
// 	 _set_glyph_props (glyph_index);
// 	 buffer.Cur().Codepoint = glyph_index;
//    }
//    void replace_glyph_with_ligature (hb_codepoint_t glyph_index,
// 					 uint class_guess) const
//    {
// 	 _set_glyph_props (glyph_index, class_guess, true);
// 	 buffer.replace_glyph (glyph_index);
//    }
//    void output_glyph_for_component (hb_codepoint_t glyph_index,
// 					uint class_guess) const
//    {
// 	 _set_glyph_props (glyph_index, class_guess, false, true);
// 	 buffer.output_glyph (glyph_index);
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
//  static inline void collect_coverage (hb_set_t *glyphs, const HBUINT16 &value, const void *data)
//  {
//    const OffsetTo<Coverage> &coverage = (const OffsetTo<Coverage>&)value;
//    (data+coverage).collect_coverage (glyphs);
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
// 					   match_func_t match_func,
// 					   const void *match_data)
//  {
//    if (count != c.len)
// 	 return false;

//    for (uint i = 1; i < count; i++)
// 	 if (likely (!match_func (c.glyphs[i], input[i - 1], match_data)))
// 	   return false;

//    return true;
//  }
//  static inline bool match_input (hb_ot_apply_context_t *c,
// 				 uint count, /* Including the first glyph (not matched) */
// 				 const HBUINT16 input[], /* Array of input values--start with second glyph */
// 				 match_func_t match_func,
// 				 const void *match_data,
// 				 uint *end_offset,
// 				 uint match_positions[HB_MAX_CONTEXT_LENGTH],
// 				 uint *p_total_component_count = nullptr)
//  {
//    TRACE_APPLY (nullptr);

//    if (unlikely (count > HB_MAX_CONTEXT_LENGTH)) return_trace (false);

//    cm.Buffer *buffer = c.buffer;

//    hb_ot_apply_context_t::skipping_iterator_t &skippy_iter = c.iter_input;
//    skippy_iter.reset (buffer.idx, count - 1);
//    skippy_iter.set_match_func (match_func, match_data, input);

//    /*
// 	* This is perhaps the trickiest part of OpenType...  Remarks:
// 	*
// 	* - If all components of the ligature were marks, we call this a mark ligature.
// 	*
// 	* - If there is no GDEF, and the ligature is NOT a mark ligature, we categorize
// 	*   it as a ligature glyph.
// 	*
// 	* - Ligatures cannot be formed across glyphs attached to different components
// 	*   of previous ligatures.  Eg. the sequence is LAM,SHADDA,LAM,FATHA,HEH, and
// 	*   LAM,LAM,HEH form a ligature, leaving SHADDA,FATHA next to eachother.
// 	*   However, it would be wrong to ligate that SHADDA,FATHA sequence.
// 	*   There are a couple of exceptions to this:
// 	*
// 	*   o If a ligature tries ligating with marks that belong to it itself, go ahead,
// 	*     assuming that the font designer knows what they are doing (otherwise it can
// 	*     break Indic stuff when a matra wants to ligate with a conjunct,
// 	*
// 	*   o If two marks want to ligate and they belong to different components of the
// 	*     same ligature glyph, and said ligature glyph is to be ignored according to
// 	*     mark-filtering rules, then allow.
// 	*     https://github.com/harfbuzz/harfbuzz/issues/545
// 	*/

//    uint total_component_count = 0;
//    total_component_count += _hb_glyph_info_get_lig_num_comps (&buffer.Cur());

//    uint first_lig_id = _hb_glyph_info_get_lig_id (&buffer.Cur());
//    uint first_lig_comp = _hb_glyph_info_get_lig_comp (&buffer.Cur());

//    enum {
// 	 LIGBASE_NOT_CHECKED,
// 	 LIGBASE_MAY_NOT_SKIP,
// 	 LIGBASE_MAY_SKIP
//    } ligbase = LIGBASE_NOT_CHECKED;

//    match_positions[0] = buffer.idx;
//    for (uint i = 1; i < count; i++)
//    {
// 	 if (!skippy_iter.next ()) return_trace (false);

// 	 match_positions[i] = skippy_iter.idx;

// 	 uint this_lig_id = _hb_glyph_info_get_lig_id (&buffer.info[skippy_iter.idx]);
// 	 uint this_lig_comp = _hb_glyph_info_get_lig_comp (&buffer.info[skippy_iter.idx]);

// 	 if (first_lig_id && first_lig_comp)
// 	 {
// 	   /* If first component was attached to a previous ligature component,
// 		* all subsequent components should be attached to the same ligature
// 		* component, otherwise we shouldn't ligate them... */
// 	   if (first_lig_id != this_lig_id || first_lig_comp != this_lig_comp)
// 	   {
// 	 /* ...unless, we are attached to a base ligature and that base
// 	  * ligature is ignorable. */
// 	 if (ligbase == LIGBASE_NOT_CHECKED)
// 	 {
// 	   bool found = false;
// 	   const auto *out = buffer.out_info;
// 	   uint j = buffer.out_len;
// 	   while (j && _hb_glyph_info_get_lig_id (&out[j - 1]) == first_lig_id)
// 	   {
// 		 if (_hb_glyph_info_get_lig_comp (&out[j - 1]) == 0)
// 		 {
// 		   j--;
// 		   found = true;
// 		   break;
// 		 }
// 		 j--;
// 	   }

// 	   if (found && skippy_iter.may_skip (out[j]) == hb_ot_apply_context_t::matcher_t::SKIP_YES)
// 		 ligbase = LIGBASE_MAY_SKIP;
// 	   else
// 		 ligbase = LIGBASE_MAY_NOT_SKIP;
// 	 }

// 	 if (ligbase == LIGBASE_MAY_NOT_SKIP)
// 	   return_trace (false);
// 	   }
// 	 }
// 	 else
// 	 {
// 	   /* If first component was NOT attached to a previous ligature component,
// 		* all subsequent components should also NOT be attached to any ligature
// 		* component, unless they are attached to the first component itself! */
// 	   if (this_lig_id && this_lig_comp && (this_lig_id != first_lig_id))
// 	 return_trace (false);
// 	 }

// 	 total_component_count += _hb_glyph_info_get_lig_num_comps (&buffer.info[skippy_iter.idx]);
//    }

//    *end_offset = skippy_iter.idx - buffer.idx + 1;

//    if (p_total_component_count)
// 	 *p_total_component_count = total_component_count;

//    return_trace (true);
//  }
//  static inline bool ligate_input (hb_ot_apply_context_t *c,
// 				  uint count, /* Including the first glyph */
// 				  const uint match_positions[HB_MAX_CONTEXT_LENGTH], /* Including the first glyph */
// 				  uint match_length,
// 				  hb_codepoint_t lig_glyph,
// 				  uint total_component_count)
//  {
//    TRACE_APPLY (nullptr);

//    cm.Buffer *buffer = c.buffer;

//    buffer.MergeClusters (buffer.idx, buffer.idx + match_length);

//    /* - If a base and one or more marks ligate, consider that as a base, NOT
// 	*   ligature, such that all following marks can still attach to it.
// 	*   https://github.com/harfbuzz/harfbuzz/issues/1109
// 	*
// 	* - If all components of the ligature were marks, we call this a mark ligature.
// 	*   If it *is* a mark ligature, we don't allocate a new ligature id, and leave
// 	*   the ligature to keep its old ligature id.  This will allow it to attach to
// 	*   a base ligature in GPOS.  Eg. if the sequence is: LAM,LAM,SHADDA,FATHA,HEH,
// 	*   and LAM,LAM,HEH for a ligature, they will leave SHADDA and FATHA with a
// 	*   ligature id and component value of 2.  Then if SHADDA,FATHA form a ligature
// 	*   later, we don't want them to lose their ligature id/component, otherwise
// 	*   GPOS will fail to correctly position the mark ligature on top of the
// 	*   LAM,LAM,HEH ligature.  See:
// 	*     https://bugzilla.gnome.org/show_bug.cgi?id=676343
// 	*
// 	* - If a ligature is formed of components that some of which are also ligatures
// 	*   themselves, and those ligature components had marks attached to *their*
// 	*   components, we have to attach the marks to the new ligature component
// 	*   positions!  Now *that*'s tricky!  And these marks may be following the
// 	*   last component of the whole sequence, so we should loop forward looking
// 	*   for them and update them.
// 	*
// 	*   Eg. the sequence is LAM,LAM,SHADDA,FATHA,HEH, and the font first forms a
// 	*   'calt' ligature of LAM,HEH, leaving the SHADDA and FATHA with a ligature
// 	*   id and component == 1.  Now, during 'liga', the LAM and the LAM-HEH ligature
// 	*   form a LAM-LAM-HEH ligature.  We need to reassign the SHADDA and FATHA to
// 	*   the new ligature with a component value of 2.
// 	*
// 	*   This in fact happened to a font...  See:
// 	*   https://bugzilla.gnome.org/show_bug.cgi?id=437633
// 	*/

//    bool is_base_ligature = _hb_glyph_info_is_base_glyph (&buffer.info[match_positions[0]]);
//    bool is_mark_ligature = _hb_glyph_info_is_mark (&buffer.info[match_positions[0]]);
//    for (uint i = 1; i < count; i++)
// 	 if (!_hb_glyph_info_is_mark (&buffer.info[match_positions[i]]))
// 	 {
// 	   is_base_ligature = false;
// 	   is_mark_ligature = false;
// 	   break;
// 	 }
//    bool is_ligature = !is_base_ligature && !is_mark_ligature;

//    uint klass = is_ligature ? HB_OT_LAYOUT_GLYPH_PROPS_LIGATURE : 0;
//    uint lig_id = is_ligature ? _hb_allocate_lig_id (buffer) : 0;
//    uint last_lig_id = _hb_glyph_info_get_lig_id (&buffer.Cur());
//    uint last_num_components = _hb_glyph_info_get_lig_num_comps (&buffer.Cur());
//    uint components_so_far = last_num_components;

//    if (is_ligature)
//    {
// 	 _hb_glyph_info_set_lig_props_for_ligature (&buffer.Cur(), lig_id, total_component_count);
// 	 if (_hb_glyph_info_get_general_category (&buffer.Cur()) == HB_UNICODE_GENERAL_CATEGORY_NON_SPACING_MARK)
// 	 {
// 	   _hb_glyph_info_set_general_category (&buffer.Cur(), HB_UNICODE_GENERAL_CATEGORY_OTHER_LETTER);
// 	 }
//    }
//    c.replace_glyph_with_ligature (lig_glyph, klass);

//    for (uint i = 1; i < count; i++)
//    {
// 	 while (buffer.idx < match_positions[i] && buffer.successful)
// 	 {
// 	   if (is_ligature)
// 	   {
// 	 uint this_comp = _hb_glyph_info_get_lig_comp (&buffer.Cur());
// 	 if (this_comp == 0)
// 	   this_comp = last_num_components;
// 	 uint new_lig_comp = components_so_far - last_num_components +
// 					 hb_min (this_comp, last_num_components);
// 	   _hb_glyph_info_set_lig_props_for_mark (&buffer.Cur(), lig_id, new_lig_comp);
// 	   }
// 	   buffer.NextGlyph ();
// 	 }

// 	 last_lig_id = _hb_glyph_info_get_lig_id (&buffer.Cur());
// 	 last_num_components = _hb_glyph_info_get_lig_num_comps (&buffer.Cur());
// 	 components_so_far += last_num_components;

// 	 /* Skip the base glyph */
// 	 buffer.idx++;
//    }

//    if (!is_mark_ligature && last_lig_id)
//    {
// 	 /* Re-adjust components for any marks following. */
// 	 for (unsigned i = buffer.idx; i < buffer.len; ++i)
// 	 {
// 	   if (last_lig_id != _hb_glyph_info_get_lig_id (&buffer.info[i])) break;

// 	   unsigned this_comp = _hb_glyph_info_get_lig_comp (&buffer.info[i]);
// 	   if (!this_comp) break;

// 	   unsigned new_lig_comp = components_so_far - last_num_components +
// 				   hb_min (this_comp, last_num_components);
// 	   _hb_glyph_info_set_lig_props_for_mark (&buffer.info[i], lig_id, new_lig_comp);
// 	 }
//    }
//    return_trace (true);
//  }

//  static inline bool match_backtrack (hb_ot_apply_context_t *c,
// 					 uint count,
// 					 const HBUINT16 backtrack[],
// 					 match_func_t match_func,
// 					 const void *match_data,
// 					 uint *match_start)
//  {
//    TRACE_APPLY (nullptr);

//    hb_ot_apply_context_t::skipping_iterator_t &skippy_iter = c.iter_context;
//    skippy_iter.reset (c.buffer.backtrack_len (), count);
//    skippy_iter.set_match_func (match_func, match_data, backtrack);

//    for (uint i = 0; i < count; i++)
// 	 if (!skippy_iter.prev ())
// 	   return_trace (false);

//    *match_start = skippy_iter.idx;

//    return_trace (true);
//  }

//  static inline bool match_lookahead (hb_ot_apply_context_t *c,
// 					 uint count,
// 					 const HBUINT16 lookahead[],
// 					 match_func_t match_func,
// 					 const void *match_data,
// 					 uint offset,
// 					 uint *end_index)
//  {
//    TRACE_APPLY (nullptr);

//    hb_ot_apply_context_t::skipping_iterator_t &skippy_iter = c.iter_context;
//    skippy_iter.reset (c.buffer.idx + offset - 1, count);
//    skippy_iter.set_match_func (match_func, match_data, lookahead);

//    for (uint i = 0; i < count; i++)
// 	 if (!skippy_iter.next ())
// 	   return_trace (false);

//    *end_index = skippy_iter.idx + 1;

//    return_trace (true);
//  }

//  struct LookupRecord
//  {
//    LookupRecord* copy (hb_serialize_context_t *c,
// 			   const hb_map_t         *lookup_map) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 auto *out = c.embed (*this);
// 	 if (unlikely (!out)) return_trace (nullptr);

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

//  static inline bool apply_lookup (hb_ot_apply_context_t *c,
// 				  uint count, /* Including the first glyph */
// 				  uint match_positions[HB_MAX_CONTEXT_LENGTH], /* Including the first glyph */
// 				  uint lookupCount,
// 				  const LookupRecord lookupRecord[], /* Array of LookupRecords--in design order */
// 				  uint match_length)
//  {
//    TRACE_APPLY (nullptr);

//    cm.Buffer *buffer = c.buffer;
//    int end;

//    /* All positions are distance from beginning of *output* buffer.
// 	* Adjust. */
//    {
// 	 uint bl = buffer.backtrack_len ();
// 	 end = bl + match_length;

// 	 int delta = bl - buffer.idx;
// 	 /* Convert positions to new indexing. */
// 	 for (uint j = 0; j < count; j++)
// 	   match_positions[j] += delta;
//    }

//    for (uint i = 0; i < lookupCount && buffer.successful; i++)
//    {
// 	 uint idx = lookupRecord[i].sequenceIndex;
// 	 if (idx >= count)
// 	   continue;

// 	 /* Don't recurse to ourself at same position.
// 	  * Note that this test is too naive, it doesn't catch longer loops. */
// 	 if (idx == 0 && lookupRecord[i].lookupListIndex == c.lookup_index)
// 	   continue;

// 	 if (unlikely (!buffer.move_to (match_positions[idx])))
// 	   break;

// 	 if (unlikely (buffer.max_ops <= 0))
// 	   break;

// 	 uint orig_len = buffer.backtrack_len () + buffer.lookahead_len ();
// 	 if (!c.recurse (lookupRecord[i].lookupListIndex))
// 	   continue;

// 	 uint new_len = buffer.backtrack_len () + buffer.lookahead_len ();
// 	 int delta = new_len - orig_len;

// 	 if (!delta)
// 	   continue;

// 	 /* Recursed lookup changed buffer len.  Adjust.
// 	  *
// 	  * TODO:
// 	  *
// 	  * Right now, if buffer length increased by n, we assume n new glyphs
// 	  * were added right after the current position, and if buffer length
// 	  * was decreased by n, we assume n match positions after the current
// 	  * one where removed.  The former (buffer length increased) case is
// 	  * fine, but the decrease case can be improved in at least two ways,
// 	  * both of which are significant:
// 	  *
// 	  *   - If recursed-to lookup is MultipleSubst and buffer length
// 	  *     decreased, then it's current match position that was deleted,
// 	  *     NOT the one after it.
// 	  *
// 	  *   - If buffer length was decreased by n, it does not necessarily
// 	  *     mean that n match positions where removed, as there might
// 	  *     have been marks and default-ignorables in the sequence.  We
// 	  *     should instead drop match positions between current-position
// 	  *     and current-position + n instead.
// 	  *
// 	  * It should be possible to construct tests for both of these cases.
// 	  */

// 	 end += delta;
// 	 if (end <= int (match_positions[idx]))
// 	 {
// 	   /* End might end up being smaller than match_positions[idx] if the recursed
// 		* lookup ended up removing many items, more than we have had matched.
// 		* Just never rewind end back and get out of here.
// 		* https://bugs.chromium.org/p/chromium/issues/detail?id=659496 */
// 	   end = match_positions[idx];
// 	   /* There can't be any further changes. */
// 	   break;
// 	 }

// 	 uint next = idx + 1; /* next now is the position after the recursed lookup. */

// 	 if (delta > 0)
// 	 {
// 	   if (unlikely (delta + count > HB_MAX_CONTEXT_LENGTH))
// 	 break;
// 	 }
// 	 else
// 	 {
// 	   /* NOTE: delta is negative. */
// 	   delta = hb_max (delta, (int) next - (int) count);
// 	   next -= delta;
// 	 }

// 	 /* Shift! */
// 	 memmove (match_positions + next + delta, match_positions + next,
// 		  (count - next) * sizeof (match_positions[0]));
// 	 next += delta;
// 	 count += delta;

// 	 /* Fill in new entries. */
// 	 for (uint j = idx + 1; j < next; j++)
// 	   match_positions[j] = match_positions[j - 1] + 1;

// 	 /* And fixup the rest. */
// 	 for (; next < count; next++)
// 	   match_positions[next] += delta;
//    }

//    buffer.move_to (end);

//    return_trace (true);
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
//  static inline bool context_apply_lookup (hb_ot_apply_context_t *c,
// 					  uint inputCount, /* Including the first glyph (not matched) */
// 					  const HBUINT16 input[], /* Array of input values--start with second glyph */
// 					  uint lookupCount,
// 					  const LookupRecord lookupRecord[],
// 					  ContextApplyLookupContext &lookup_context)
//  {
//    uint match_length = 0;
//    uint match_positions[HB_MAX_CONTEXT_LENGTH];
//    return match_input (c,
// 			   inputCount, input,
// 			   lookup_context.funcs.match, lookup_context.match_data,
// 			   &match_length, match_positions)
// 	   && (c.buffer.UnsafeToBreak (c.buffer.idx, c.buffer.idx + match_length),
// 	   apply_lookup (c,
// 				inputCount, match_positions,
// 				lookupCount, lookupRecord,
// 				match_length));
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

//    bool apply (hb_ot_apply_context_t *c,
// 		   ContextApplyLookupContext &lookup_context) const
//    {
// 	 TRACE_APPLY (this);
// 	 const UnsizedArrayOf<LookupRecord> &lookupRecord = StructAfter<UnsizedArrayOf<LookupRecord>>
// 								(inputZ.as_array (inputCount ? inputCount - 1 : 0));
// 	 return_trace (context_apply_lookup (c, inputCount, inputZ.arrayZ, lookupCount, lookupRecord.arrayZ, lookup_context));
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
// 			const hb_map_t *klass_map = nullptr) const
//    {
// 	 TRACE_SUBSET (this);

// 	 const hb_array_t<const HBUINT16> input = inputZ.as_array ((inputCount ? inputCount - 1 : 0));
// 	 if (!input.length) return_trace (false);

// 	 const hb_map_t *mapping = klass_map == nullptr ? c.plan.glyph_map : klass_map;
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

//    bool apply (hb_ot_apply_context_t *c,
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
// 			const hb_map_t *klass_map = nullptr) const
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
// 	   nullptr
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
// 	   nullptr
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
// 	 (this+coverage).collect_coverage (c.input);

// 	 struct ContextCollectGlyphsLookupContext lookup_context = {
// 	   {collect_glyph},
// 	   nullptr
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
// 	   nullptr
// 	 };
// 	 return rule_set.would_apply (c, lookup_context);
//    }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 uint index = (this+coverage).get_coverage (c.buffer.Cur().Codepoint);
// 	 if (likely (index == NOT_COVERED))
// 	   return_trace (false);

// 	 const RuleSet &rule_set = this+ruleSet[index];
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_glyph},
// 	   nullptr
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
// 	 (this+coverage).collect_coverage (c.input);

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

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 uint index = (this+coverage).get_coverage (c.buffer.Cur().Codepoint);
// 	 if (likely (index == NOT_COVERED)) return_trace (false);

// 	 const ClassDef &class_def = this+classDef;
// 	 index = class_def.get_class (c.buffer.Cur().Codepoint);
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
// 	 (this+coverageZ[0]).collect_coverage (c.input);

// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 struct ContextCollectGlyphsLookupContext lookup_context = {
// 	   {collect_coverage},
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

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 uint index = (this+coverageZ[0]).get_coverage (c.buffer.Cur().Codepoint);
// 	 if (likely (index == NOT_COVERED)) return_trace (false);

// 	 const LookupRecord *lookupRecord = &StructAfter<LookupRecord> (coverageZ.as_array (glyphCount));
// 	 struct ContextApplyLookupContext lookup_context = {
// 	   {match_coverage},
// 	   this
// 	 };
// 	 return_trace (context_apply_lookup (c, glyphCount, (const HBUINT16 *) (coverageZ.arrayZ + 1), lookupCount, lookupRecord, lookup_context));
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

//  static inline bool chain_context_apply_lookup (hb_ot_apply_context_t *c,
// 							uint backtrackCount,
// 							const HBUINT16 backtrack[],
// 							uint inputCount, /* Including the first glyph (not matched) */
// 							const HBUINT16 input[], /* Array of input values--start with second glyph */
// 							uint lookaheadCount,
// 							const HBUINT16 lookahead[],
// 							uint lookupCount,
// 							const LookupRecord lookupRecord[],
// 							ChainContextApplyLookupContext &lookup_context)
//  {
//    uint start_index = 0, match_length = 0, end_index = 0;
//    uint match_positions[HB_MAX_CONTEXT_LENGTH];
//    return match_input (c,
// 			   inputCount, input,
// 			   lookup_context.funcs.match, lookup_context.match_data[1],
// 			   &match_length, match_positions)
// 	   && match_backtrack (c,
// 			   backtrackCount, backtrack,
// 			   lookup_context.funcs.match, lookup_context.match_data[0],
// 			   &start_index)
// 	   && match_lookahead (c,
// 			   lookaheadCount, lookahead,
// 			   lookup_context.funcs.match, lookup_context.match_data[2],
// 			   match_length, &end_index)
// 	   && (c.buffer.unsafe_to_break_from_outbuffer (start_index, end_index),
// 	   apply_lookup (c,
// 			 inputCount, match_positions,
// 			 lookupCount, lookupRecord,
// 			 match_length));
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
