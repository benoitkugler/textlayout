package harfbuzz

import "math/bits"

// ported from harfbuzz/src/hb-ot-map.cc, hb-ot-map.hh Copyright © 2009,2010  Red Hat, Inc. 2010,2011,2013  Google, Inc. Behdad Esfahbod

const (
	// Maximum number of OpenType tags that can correspond to a given `hb_script_t`.
	HB_OT_MAX_TAGS_PER_SCRIPT = 3
	// Maximum number of OpenType tags that can correspond to a given `hb_language_t`.
	HB_OT_MAX_TAGS_PER_LANGUAGE = 3
)

var table_tags = [2]hb_tag_t{HB_OT_TAG_GSUB, HB_OT_TAG_GPOS}

type hb_ot_map_feature_flags_t uint8

const (
	F_GLOBAL                hb_ot_map_feature_flags_t = 1 << iota /* Feature applies to all characters; results in no mask allocated for it. */
	F_HAS_FALLBACK                                                /* Has fallback implementation, so include mask bit even if feature not found. */
	F_MANUAL_ZWNJ                                                 /* Don't skip over ZWNJ when matching **context**. */
	F_MANUAL_ZWJ                                                  /* Don't skip over ZWJ when matching **input**. */
	F_GLOBAL_SEARCH                                               /* If feature not found in LangSys, look for it in global feature list and pick one. */
	F_RANDOM                                                      /* Randomly select a glyph from an AlternateSubstFormat1 subtable. */
	F_NONE                  hb_ot_map_feature_flags_t = 0
	F_MANUAL_JOINERS                                  = F_MANUAL_ZWNJ | F_MANUAL_ZWJ
	F_GLOBAL_MANUAL_JOINERS                           = F_GLOBAL | F_MANUAL_JOINERS
	F_GLOBAL_HAS_FALLBACK                             = F_GLOBAL | F_HAS_FALLBACK
)

const (
	HB_OT_MAP_MAX_BITS  = 8
	HB_OT_MAP_MAX_VALUE = (1 << HB_OT_MAP_MAX_BITS) - 1
)

type hb_ot_map_feature_t struct {
	tag   hb_tag_t
	flags hb_ot_map_feature_flags_t
}

type feature_info_t struct {
	tag           hb_tag_t
	seq           int /* sequence#, used for stable sorting only */
	max_value     uint
	flags         hb_ot_map_feature_flags_t
	default_value uint    /* for non-global features, what should the unset glyphs take */
	stage         [2]uint /* GSUB/GPOS */

	// HB_INTERNAL static int cmp (const void *pa, const void *pb)
	// {
	//   const feature_info_t *a = (const feature_info_t *) pa;
	//   const feature_info_t *b = (const feature_info_t *) pb;
	//   return (a->tag != b->tag) ?  (a->tag < b->tag ? -1 : 1) :
	//      (a->seq < b->seq ? -1 : a->seq > b->seq ? 1 : 0);
	// }
}

type stage_info_t struct {
	index      int
	pause_func pause_func_t
}

type hb_ot_map_builder_t struct {

	//   public:

	face  hb_face_t
	props hb_segment_properties_t

	chosen_script                [2]hb_tag_t
	found_script                 [2]bool
	script_index, language_index [2]uint

	//   private:

	current_stage [2]uint /* GSUB/GPOS */
	feature_infos []feature_info_t
	stages        [2][]stage_info_t /* GSUB/GPOS */
}

//  void hb_ot_map_t::collect_lookups (uint table_index, hb_set_t *lookups_out) const
//  {
//    for (uint i = 0; i < lookups[table_index].length; i++)
// 	 lookups_out.add (lookups[table_index][i].index);
//  }

func new_hb_ot_map_builder_t(face hb_face_t, props hb_segment_properties_t) *hb_ot_map_builder_t {
	var out hb_ot_map_builder_t

	out.face = face
	out.props = props

	/* Fetch script/language indices for GSUB/GPOS.  We need these later to skip
	* features not available in either table and not waste precious bits for them. */

	script_count := HB_OT_MAX_TAGS_PER_SCRIPT
	language_count := HB_OT_MAX_TAGS_PER_LANGUAGE
	var script_tags [HB_OT_MAX_TAGS_PER_SCRIPT]hb_tag_t
	var language_tags [HB_OT_MAX_TAGS_PER_LANGUAGE]hb_tag_t

	hb_ot_tags_from_script_and_language(props.script, props.language, &script_count, script_tags, &language_count, language_tags)

	for table_index, table_tag := range table_tags {
		out.found_script[table_index] = hb_ot_layout_table_select_script(face, table_tag, script_count, script_tags, &script_index[table_index], &chosen_script[table_index])
		hb_ot_layout_script_select_language(face, table_tag, script_index[table_index], language_count, language_tags, &language_index[table_index])
	}

	return &out
}

func (mb *hb_ot_map_builder_t) add_feature(tag hb_tag_t, flags hb_ot_map_feature_flags_t, value uint) {
	var info feature_info_t
	info.tag = tag
	info.seq = len(mb.feature_infos) + 1
	info.max_value = value
	info.flags = flags
	if (flags & F_GLOBAL) != 0 {
		info.default_value = value
	}
	info.stage = mb.current_stage

	mb.feature_infos = append(mb.feature_infos, info)
}

//  void
//  hb_ot_map_builder_t::add_lookups (hb_ot_map_t  &m,
// 				   uint  table_index,
// 				   uint  feature_index,
// 				   uint  variations_index,
// 				   hb_mask_t     mask,
// 				   bool          auto_zwnj,
// 				   bool          auto_zwj,
// 				   bool          random)
//  {
//    uint lookup_indices[32];
//    uint offset, len;
//    uint table_lookup_count;

//    table_lookup_count = hb_ot_layout_table_get_lookup_count (face, table_tags[table_index]);

//    offset = 0;
//    do {
// 	 len = ARRAY_LENGTH (lookup_indices);
// 	 hb_ot_layout_feature_with_variations_get_lookups (face,
// 							   table_tags[table_index],
// 							   feature_index,
// 							   variations_index,
// 							   offset, &len,
// 							   lookup_indices);

// 	 for (uint i = 0; i < len; i++)
// 	 {
// 	   if (lookup_indices[i] >= table_lookup_count)
// 	 continue;
// 	   hb_ot_map_t::lookup_map_t *lookup = m.lookups[table_index].push ();
// 	   lookup.mask = mask;
// 	   lookup.index = lookup_indices[i];
// 	   lookup.auto_zwnj = auto_zwnj;
// 	   lookup.auto_zwj = auto_zwj;
// 	   lookup.random = random;
// 	 }

// 	 offset += len;
//    } while (len == ARRAY_LENGTH (lookup_indices));
//  }

type pause_func_t = func(plan *hb_ot_shape_plan_t, font *hb_font_t, buffer *hb_buffer_t)

func (m *hb_ot_map_builder_t) add_pause(table_index int, pause_func pause_func_t) {
	s := m.stages[table_index].push()
	s.index = m.current_stage[table_index]
	s.pause_func = pause_func

	m.current_stage[table_index]++
}

func (mb *hb_ot_map_builder_t) compile(m *hb_ot_map_t, key *hb_ot_shape_plan_key_t) {
	global_bit_mask := HB_GLYPH_FLAG_DEFINED + 1
	global_bit_shift := bits.OnesCount32(HB_GLYPH_FLAG_DEFINED)

	m.global_mask = global_bit_mask

	var (
		required_feature_index [2]uint
		required_feature_tag   [2]hb_tag_t
		/* We default to applying required feature in stage 0.  If the required
		* feature has a tag that is known to the shaper, we apply required feature
		* in the stage for that tag.
		 */
		required_feature_stage [2]uint
	)

	for table_index = 0; table_index < 2; table_index++ {
		m.chosen_script[table_index] = chosen_script[table_index]
		m.found_script[table_index] = found_script[table_index]

		hb_ot_layout_language_get_required_feature(face,
			table_tags[table_index],
			script_index[table_index],
			language_index[table_index],
			&required_feature_index[table_index],
			&required_feature_tag[table_index])
	}

	/* Sort features and merge duplicates */
	if feature_infos.length {
		feature_infos.qsort()
		//  uint j = 0;
		for i = 1; i < feature_infos.length; i++ {
			if feature_infos[i].tag != feature_infos[j].tag {
				j++
				feature_infos[j] = feature_infos[i]
			} else {
				if feature_infos[i].flags & F_GLOBAL {
					feature_infos[j].flags |= F_GLOBAL
					feature_infos[j].max_value = feature_infos[i].max_value
					feature_infos[j].default_value = feature_infos[i].default_value
				} else {
					if feature_infos[j].flags & F_GLOBAL {
						feature_infos[j].flags ^= F_GLOBAL
					}
					feature_infos[j].max_value = hb_max(feature_infos[j].max_value, feature_infos[i].max_value)
					/* Inherit default_value from j */
				}
				feature_infos[j].flags |= (feature_infos[i].flags & F_HAS_FALLBACK)
				feature_infos[j].stage[0] = hb_min(feature_infos[j].stage[0], feature_infos[i].stage[0])
				feature_infos[j].stage[1] = hb_min(feature_infos[j].stage[1], feature_infos[i].stage[1])
			}
		}
		feature_infos.shrink(j + 1)
	}

	/* Allocate bits now */
	next_bit := global_bit_shift + 1

	for i = 0; i < feature_infos.length; i++ {
		const feature_info_t *info = &feature_infos[i]

		bits_needed := 0

		if (info.flags & F_GLOBAL) && info.max_value == 1 {
			/* Uses the global bit */
			bits_needed = 0
		} else {
			/* Limit bits per feature. */
			bits_needed = hb_min(HB_OT_MAP_MAX_BITS, hb_bit_storage(info.max_value))
		}

		if !info.max_value || next_bit+bits_needed > 8*sizeof(hb_mask_t) {
			continue /* Feature disabled, or not enough bits. */
		}

		found := false
		var feature_index [2]uint
		for table_index = 0; table_index < 2; table_index++ {
			if required_feature_tag[table_index] == info.tag {
				required_feature_stage[table_index] = info.stage[table_index]
			}

			found = found || hb_ot_layout_language_find_feature(face,
				table_tags[table_index],
				script_index[table_index],
				language_index[table_index],
				info.tag,
				&feature_index[table_index])
		}
		if !found && (info.flags & F_GLOBAL_SEARCH) {
			for table_index = 0; table_index < 2; table_index++ {
				found = found || hb_ot_layout_table_find_feature(face,
					table_tags[table_index],
					info.tag,
					&feature_index[table_index])
			}
		}
		if !found && !(info.flags & F_HAS_FALLBACK) {
			continue
		}

		map_ := m.features.push()

		map_.tag = info.tag
		map_.index[0] = feature_index[0]
		map_.index[1] = feature_index[1]
		map_.stage[0] = info.stage[0]
		map_.stage[1] = info.stage[1]
		map_.auto_zwnj = !(info.flags & F_MANUAL_ZWNJ)
		map_.auto_zwj = !(info.flags & F_MANUAL_ZWJ)
		map_.random = !!(info.flags & F_RANDOM)
		if (info.flags & F_GLOBAL) && info.max_value == 1 {
			/* Uses the global bit */
			map_.shift = global_bit_shift
			map_.mask = global_bit_mask
		} else {
			map_.shift = next_bit
			map_.mask = (1 << (next_bit + bits_needed)) - (1 << next_bit)
			next_bit += bits_needed
			m.global_mask |= (info.default_value << map_.shift) & map_.mask
		}
		map_._1_mask = (1 << map_.shift) & map_.mask
		map_.needs_fallback = !found

	}
	feature_infos.shrink(0) /* Done with these */

	add_gsub_pause(nullptr)
	add_gpos_pause(nullptr)

	for table_index := 0; table_index < 2; table_index++ {
		/* Collect lookup indices for features */

		stage_index := 0
		last_num_lookups := 0
		for stage = 0; stage < current_stage[table_index]; stage++ {
			if required_feature_index[table_index] != HB_OT_LAYOUT_NO_FEATURE_INDEX &&
				required_feature_stage[table_index] == stage {
				add_lookups(m, table_index,
					required_feature_index[table_index],
					key.variations_index[table_index],
					global_bit_mask)
			}

			for i = 0; i < m.features.length; i++ {
				if m.features[i].stage[table_index] == stage {
					add_lookups(m, table_index,
						m.features[i].index[table_index],
						key.variations_index[table_index],
						m.features[i].mask,
						m.features[i].auto_zwnj,
						m.features[i].auto_zwj,
						m.features[i].random)
				}
			}

			/* Sort lookups and merge duplicates */
			if last_num_lookups < m.lookups[table_index].length {
				m.lookups[table_index].qsort(last_num_lookups, m.lookups[table_index].length)

				j := last_num_lookups
				for i := j + 1; i < m.lookups[table_index].length; i++ {
					if m.lookups[table_index][i].index != m.lookups[table_index][j].index {
						j++
						m.lookups[table_index][j] = m.lookups[table_index][i]
					} else {
						m.lookups[table_index][j].mask |= m.lookups[table_index][i].mask
						m.lookups[table_index][j].auto_zwnj &= m.lookups[table_index][i].auto_zwnj
						m.lookups[table_index][j].auto_zwj &= m.lookups[table_index][i].auto_zwj
					}
				}
				m.lookups[table_index].shrink(j + 1)
			}

			last_num_lookups = m.lookups[table_index].length

			if stage_index < stages[table_index].length && stages[table_index][stage_index].index == stage {
				stage_map := m.stages[table_index].push()
				stage_map.last_lookup = last_num_lookups
				stage_map.pause_func = stages[table_index][stage_index].pause_func

				stage_index++
			}
		}
	}
}
