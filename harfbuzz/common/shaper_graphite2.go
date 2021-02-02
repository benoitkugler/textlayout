package common

// ported from harfbuzz/src/hb-graphite2.cc
// Copyright © 2011  Martin Hosken
// Copyright © 2011  SIL International
// Copyright © 2011,2012  Google, Inc.  Behdad Esfahbod

/**
 * SECTION:hb-graphite2
 * @title: hb-graphite2
 * @short_description: Graphite2 integration
 * @include: hb-graphite2.h
 *
 * Functions for using HarfBuzz with fonts that include Graphite features.
 *
 * For Graphite features to work, you must be sure that HarfBuzz was compiled
 * with the `graphite2` shaping engine enabled. Currently, the default is to
 * not enable `graphite2` shaping.
 **/

type grface struct{}

/*
 * shaper face data
 */

type hb_graphite2_tablelist_t []struct {
	// blob *hb_blob_t
	tag uint
}

type hb_graphite2_face_data_t struct {
	face hb_face_t
	// grface gr_face
	tlist hb_graphite2_tablelist_t
}

// func hb_graphite2_get_table (face_data *hb_graphite2_face_data_t, tag uint) {
//    tlist := face_data.tlist;

//     var blob *hb_blob_t

//    for _, p :=range tlist {
// 	 if (p.tag == tag) {
// 	   blob = p.blob;
// 	   break;
// 	 }
// 	}

//    if blob == nil {
// 	 blob = face_data.face.reference_table (tag);

// 	 hb_graphite2_tablelist_t *p = (hb_graphite2_tablelist_t *) calloc (1, sizeof (hb_graphite2_tablelist_t));
// 	 if (unlikely (!p)) {
// 	   hb_blob_destroy (blob);
// 	   return nil;
// 	 }
// 	 p.blob = blob;
// 	 p.tag = tag;

//  retry:
// 	 hb_graphite2_tablelist_t *tlist = face_data.tlist;
// 	 p.next = tlist;

// 	 if (unlikely (!face_data.tlist.cmpexch (tlist, p)))
// 	   goto retry;
//    }

//    uint tlen;
//    const char *d = hb_blob_get_data (blob, &tlen);
//    *len = tlen;
//    return d;
//  }

// func  _hb_graphite2_shaper_face_data_create (hb_face_t *face) *hb_graphite2_face_data_t  {
//    hb_blob_t *silf_blob = face.reference_table (HB_GRAPHITE2_TAG_SILF);
//    /* Umm, we just reference the table to check whether it exists.
// 	* Maybe add better API for this? */
//    if (!hb_blob_get_length (silf_blob)){
// 	 hb_blob_destroy (silf_blob);
// 	 return nil;
//    }
//    hb_blob_destroy (silf_blob);

//    hb_graphite2_face_data_t *data = (hb_graphite2_face_data_t *) calloc (1, sizeof (hb_graphite2_face_data_t));
//    if (unlikely (!data))
// 	 return nil;

//    data.face = face;
//    data.grface = gr_make_face (data, &hb_graphite2_get_table, gr_face_preloadAll);

//    if (unlikely (!data.grface)) {
// 	 free (data);
// 	 return nil;
//    }

//    return data;
//  }

/**
 * hb_graphite2_face_get_gr_face:
 * @face: @hb_face_t to query
 *
 * Fetches the Graphite2 gr_face corresponding to the specified
 * #hb_face_t face object.
 *
 * Return value: the gr_face found
 *
 * Since: 0.9.10
 */

// func hb_graphite2_face_get_gr_face (hb_face_t *face) *gr_face{
//    const hb_graphite2_face_data_t *data = face.data.graphite2;
//    return data ? data.grface : nil;
//  }

/*
 * shaper font data
 */

type hb_graphite2_font_data_t struct{}

/*
 * shaper
 */

type hb_graphite2_cluster_t struct {
	base_char  uint
	num_chars  uint
	base_glyph uint
	num_glyphs uint
	cluster    uint
	advance    uint
}

// TODO:
func _hb_graphite2_shape(_ *hb_shape_plan_t, font *Font, buffer *Buffer, features []hb_feature_t) bool {
	// face := font.face
	// grface := face.data.graphite2.grface

	// lang := hb_language_to_string(buffer.props.language)
	// lang_len := strings.IndexByte(lang, '-')
	// tagLang := 0
	// if lang != "" {
	// 	tagLang = hb_tag_from_string(lang[:lang_len])
	// }
	// feats := gr_face_featureval_for_lang(grface, tagLang)

	// for _, feature := range features {
	// 	fref := gr_face_find_fref(grface, feature.tag)
	// 	if fref {
	// 		gr_fref_set_feature_value(fref, feature.value, feats)
	// 	}
	// }

	// //    gr_segment *seg = nil;
	// //    const gr_slot *is;
	// //    uint ci = 0, ic = 0;
	// //    uint curradvx = 0, curradvy = 0;

	// //    uint scratch_size;
	// scratch := buffer.get_scratch_buffer()
	// chars := []rune(scratch)

	// for i, info := range buffer.Info {
	// 	chars[i] = buffer.Info[i].codepoint
	// }

	// /* TODO ensure_native_direction. */

	// script_tag, _ := hb_ot_tags_from_script_and_language(buffer.props.script, HB_LANGUAGE_INVALID)
	// tagScript := HB_OT_TAG_DEFAULT_SCRIPT
	// if len(script_tag) != 0 {
	// 	tagScript = script_tag[len(script_tag)-1]
	// }
	// mask := 2 | 0
	// if buffer.props.direction == HB_DIRECTION_RTL {
	// 	mask = 2 | 1
	// }
	// seg := gr_make_seg(nil, grface, tagScript, feats, gr_utf32, chars, buffer.len, mask)

	// if seg == nil {
	// 	return false
	// }

	// glyph_count := gr_seg_n_slots(seg)
	// if glyph_count == 0 {
	// 	buffer.len = 0
	// 	return true
	// }

	// buffer.ensure(glyph_count)
	// //    scratch = buffer.get_scratch_buffer ();
	// //    for ((DIV_CEIL (sizeof (hb_graphite2_cluster_t) * buffer.len, sizeof (*scratch)) +
	// // 	   DIV_CEIL (sizeof (hb_codepoint_t) * glyph_count, sizeof (*scratch))) > scratch_size)
	// //    {
	// // 	 if (unlikely (!buffer.ensure (buffer.allocated * 2)))
	// // 	 {
	// // 	   if (feats) gr_featureval_destroy (feats);
	// // 	   gr_seg_destroy (seg);
	// // 	   return false;
	// // 	 }
	// // 	 scratch = buffer.get_scratch_buffer (&scratch_size);
	// //    }

	// //  #define ALLOCATE_ARRAY(Type, name, len) \
	// //    Type *name = (Type *) scratch; \
	// //    do { \
	// // 	 uint _consumed = DIV_CEIL ((len) * sizeof (Type), sizeof (*scratch)); \
	// // 	 assert (_consumed <= scratch_size); \
	// // 	 scratch += _consumed; \
	// // 	 scratch_size -= _consumed; \
	// //    } while (0)

	// //    ALLOCATE_ARRAY (hb_graphite2_cluster_t, clusters, buffer.len);
	// //    ALLOCATE_ARRAY (hb_codepoint_t, gids, glyph_count);

	// //  #undef ALLOCATE_ARRAY

	// memset(clusters, 0, sizeof(clusters[0])*buffer.len)

	// //    hb_codepoint_t *pg = gids;
	// clusters[0].cluster = buffer.Info[0].cluster
	// upem := hb_face_get_upem(face)
	// xscale := font.x_scale / upem
	// yscale := font.y_scale / upem
	// yscale *= yscale / xscale
	// curradv := 0
	// if buffer.props.direction.isBackward() {
	// 	curradv = gr_slot_origin_X(gr_seg_first_slot(seg)) * xscale
	// 	clusters[0].advance = gr_seg_advance_X(seg)*xscale - curradv
	// } else {
	// 	clusters[0].advance = 0
	// }
	// for is, ic := gr_seg_first_slot(seg), 0; is != nil; is, ic = gr_slot_next_in_segment(is), ic+1 {
	// 	before := gr_slot_before(is)
	// 	after := gr_slot_after(is)
	// 	*pg = gr_slot_gid(is)
	// 	pg++
	// 	for clusters[ci].base_char > before && ci {
	// 		clusters[ci-1].num_chars += clusters[ci].num_chars
	// 		clusters[ci-1].num_glyphs += clusters[ci].num_glyphs
	// 		clusters[ci-1].advance += clusters[ci].advance
	// 		ci--
	// 	}

	// 	if gr_slot_can_insert_before(is) && clusters[ci].num_chars && before >= clusters[ci].base_char+clusters[ci].num_chars {
	// 		hb_graphite2_cluster_t * c = clusters + ci + 1
	// 		c.base_char = clusters[ci].base_char + clusters[ci].num_chars
	// 		c.cluster = buffer.Info[c.base_char].cluster
	// 		c.num_chars = before - c.base_char
	// 		c.base_glyph = ic
	// 		c.num_glyphs = 0
	// 		if HB_DIRECTION_IS_BACKWARD(buffer.props.direction) {
	// 			c.advance = curradv - gr_slot_origin_X(is)*xscale
	// 			curradv -= c.advance
	// 		} else {
	// 			c.advance = 0
	// 			clusters[ci].advance += gr_slot_origin_X(is)*xscale - curradv
	// 			curradv += clusters[ci].advance
	// 		}
	// 		ci++
	// 	}
	// 	clusters[ci].num_glyphs++

	// 	if clusters[ci].base_char+clusters[ci].num_chars < after+1 {
	// 		clusters[ci].num_chars = after + 1 - clusters[ci].base_char
	// 	}
	// }

	// if HB_DIRECTION_IS_BACKWARD(buffer.props.direction) {
	// 	clusters[ci].advance += curradv
	// } else {
	// 	clusters[ci].advance += gr_seg_advance_X(seg)*xscale - curradv
	// }
	// ci++

	// for i := 0; i < ci; i++ {
	// 	for j := 0; j < clusters[i].num_glyphs; j++ {
	// 		hb_glyph_info_t * info = &buffer.Info[clusters[i].base_glyph+j]
	// 		info.codepoint = gids[clusters[i].base_glyph+j]
	// 		info.cluster = clusters[i].cluster
	// 		info.var1.i32 = clusters[i].advance // all glyphs in the cluster get the same advance
	// 	}
	// }
	// buffer.len = glyph_count

	// /* Positioning. */
	// currclus := UINT_MAX
	// const hb_glyph_info_t *info = buffer.Info
	// hb_glyph_position_t * pPos = hb_buffer_get_glyph_positions(buffer, nil)
	// if !buffer.props.direction.isBackward() {
	// 	curradvx = 0
	// 	for is = gr_seg_first_slot(seg); is != nil; pPos, info, is = pPos+1, info+1, gr_slot_next_in_segment(is) {
	// 		pPos.x_offset = gr_slot_origin_X(is)*xscale - curradvx
	// 		pPos.y_offset = gr_slot_origin_Y(is)*yscale - curradvy
	// 		if info.cluster != currclus {
	// 			pPos.x_advance = info.var1.i32
	// 			curradvx += pPos.x_advance
	// 			currclus = info.cluster
	// 		} else {
	// 			pPos.x_advance = 0.
	// 		}

	// 		pPos.y_advance = gr_slot_advance_Y(is, grface, nil) * yscale
	// 		curradvy += pPos.y_advance
	// 	}
	// } else {
	// 	curradvx = gr_seg_advance_X(seg) * xscale
	// 	for is = gr_seg_first_slot(seg); is != nil; pPos, info, is = pPos+1, info+1, gr_slot_next_in_segment(is) {
	// 		if info.cluster != currclus {
	// 			pPos.x_advance = info.var1.i32
	// 			curradvx -= pPos.x_advance
	// 			currclus = info.cluster
	// 		} else {
	// 			pPos.x_advance = 0.
	// 		}

	// 		pPos.y_advance = gr_slot_advance_Y(is, grface, nil) * yscale
	// 		curradvy -= pPos.y_advance
	// 		pPos.x_offset = gr_slot_origin_X(is)*xscale - info.var1.i32 - curradvx + pPos.x_advance
	// 		pPos.y_offset = gr_slot_origin_Y(is)*yscale - curradvy
	// 	}
	// 	hb_buffer_reverse_clusters(buffer)
	// }

	// buffer.unsafe_to_break_all()

	return true
}
