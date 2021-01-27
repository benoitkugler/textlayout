package harfbuzz

import (
	"math/bits"
	"sort"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-map.cc, hb-ot-map.hh Copyright © 2009,2010  Red Hat, Inc. 2010,2011,2013  Google, Inc. Behdad Esfahbod

const (
	// Maximum number of OpenType tags that can correspond to a given `hb_script_t`.
	HB_OT_MAX_TAGS_PER_SCRIPT = 3
	// Maximum number of OpenType tags that can correspond to a given `hb_language_t`.
	HB_OT_MAX_TAGS_PER_LANGUAGE = 3
)

var (
	HB_OT_TAG_GSUB = truetype.TagGsub
	HB_OT_TAG_GPOS = truetype.TagGpos
	table_tags     = [2]hb_tag_t{HB_OT_TAG_GSUB, HB_OT_TAG_GPOS}
)

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
	tag hb_tag_t
	// seq           int /* sequence#, used for stable sorting only */
	max_value     uint32
	flags         hb_ot_map_feature_flags_t
	default_value uint32 /* for non-global features, what should the unset glyphs take */
	stage         [2]int /* GSUB/GPOS */
}

type stage_info_t struct {
	index      int
	pause_func pause_func_t
}

type hb_ot_map_builder_t struct {

	//   public:

	face  hb_face_t
	props hb_segment_properties_t

	chosen_script                [2]hb_tag_t // GSUB/GPOS
	found_script                 [2]bool     // GSUB/GPOS
	script_index, language_index [2]int      // GSUB/GPOS

	//   private:

	current_stage [2]int /* GSUB/GPOS */
	feature_infos []feature_info_t
	stages        [2][]stage_info_t /* GSUB/GPOS */
}

//  void hb_ot_map_t::collect_lookups (uint tableIndex, hb_set_t *lookups_out) const
//  {
//    for (uint i = 0; i < lookups[tableIndex].length; i++)
// 	 lookups_out.add (lookups[tableIndex][i].index);
//  }

func new_hb_ot_map_builder_t(face hb_face_t, props hb_segment_properties_t) *hb_ot_map_builder_t {
	var out hb_ot_map_builder_t

	out.face = face
	out.props = props

	/* Fetch script/language indices for GSUB/GPOS.  We need these later to skip
	* features not available in either table and not waste precious bits for them. */

	script_tags, language_tags := hb_ot_tags_from_script_and_language(props.script, props.language)

	gsub, gpos := face.get_gsubgpos_table() // TODO: check if its nil

	out.script_index[0], out.chosen_script[0], out.found_script[0] = selectScript(gsub, script_tags)
	out.language_index[0], _ = selectLanguage(gsub, out.script_index[0], language_tags)

	out.script_index[1], out.chosen_script[1], out.found_script[1] = selectScript(gpos, script_tags)
	out.language_index[1], _ = selectLanguage(gpos, out.script_index[1], language_tags)

	return &out
}

func (mb *hb_ot_map_builder_t) add_feature(tag hb_tag_t, flags hb_ot_map_feature_flags_t, value uint32) {
	var info feature_info_t
	info.tag = tag
	info.max_value = value
	info.flags = flags
	if (flags & F_GLOBAL) != 0 {
		info.default_value = value
	}
	info.stage = mb.current_stage

	mb.feature_infos = append(mb.feature_infos, info)
}

func (mb *hb_ot_map_builder_t) add_lookups(m *hb_ot_map_t, table *truetype.TableLayout, tableIndex, featureIndex, variationsIndex int,
	mask hb_mask_t, auto_zwnj, auto_zwj, random bool) {

	var (
		tableLookupCount = len(table.Lookups)
		lookupIndices    [32]int
		offset, L        = 0, 0
	)
	for do := true; do; do = L == len(lookupIndices) {
		L = len(lookupIndices)
		hb_ot_layout_feature_with_variations_get_lookups(face,
			table_tags[tableIndex],
			featureIndex, variationsIndex,
			offset, &L, lookupIndices)

		for _, lookupInd := range lookupIndices {
			if lookupInd >= tableLookupCount {
				continue
			}
			lookup := lookup_map_t{
				mask:      mask,
				index:     lookupInd,
				auto_zwnj: auto_zwnj,
				auto_zwj:  auto_zwj,
				random:    random,
			}
			m.lookups[tableIndex] = append(m.lookups[tableIndex], lookup)
		}

		offset += L
	}
}

type pause_func_t = func(plan *hb_ot_shape_plan_t, font *hb_font_t, buffer *hb_buffer_t)

func (mb *hb_ot_map_builder_t) add_pause(tableIndex int, fn pause_func_t) {
	s := stage_info_t{
		index:      mb.current_stage[tableIndex],
		pause_func: fn,
	}
	mb.stages[tableIndex] = append(mb.stages[tableIndex], s)
	mb.current_stage[tableIndex]++
}

func (mb *hb_ot_map_builder_t) add_gsub_pause(fn pause_func_t) { mb.add_pause(0, fn) }
func (mb *hb_ot_map_builder_t) add_gpos_pause(fn pause_func_t) { mb.add_pause(1, fn) }

func (mb *hb_ot_map_builder_t) compile(m *hb_ot_map_t, key *hb_ot_shape_plan_key_t) {
	globalBitMask := HB_GLYPH_FLAG_DEFINED + 1
	globalBitShift := bits.OnesCount32(uint32(HB_GLYPH_FLAG_DEFINED))

	m.global_mask = globalBitMask

	var (
		requiredFeatureIndex [2]int // -1 for empty
		requiredFeatureTag   [2]hb_tag_t
		/* We default to applying required feature in stage 0. If the required
		* feature has a tag that is known to the shaper, we apply the required feature
		* in the stage for that tag. */
		requiredFeatureStage [2]int
	)

	gsub, gpos := mb.face.get_gsubgpos_table()
	m.chosen_script = mb.chosen_script
	m.found_script = mb.found_script
	requiredFeatureIndex[0], requiredFeatureTag[0] = getRequiredFeature(gsub, mb.script_index[0], mb.language_index[0])
	requiredFeatureIndex[1], requiredFeatureTag[1] = getRequiredFeature(gpos, mb.script_index[1], mb.language_index[1])

	// sort features and merge duplicates
	if len(mb.feature_infos) != 0 {
		sort.SliceStable(mb.feature_infos, func(i, j int) bool {
			return mb.feature_infos[i].tag < mb.feature_infos[j].tag
		})
		j := 0
		for i, feat := range mb.feature_infos {
			if i == 0 {
				continue
			}
			if feat.tag != mb.feature_infos[j].tag {
				j++
				mb.feature_infos[j] = feat
				continue
			}
			if feat.flags&F_GLOBAL != 0 {
				mb.feature_infos[j].flags |= F_GLOBAL
				mb.feature_infos[j].max_value = feat.max_value
				mb.feature_infos[j].default_value = feat.default_value
			} else {
				if mb.feature_infos[j].flags&F_GLOBAL != 0 {
					mb.feature_infos[j].flags ^= F_GLOBAL
				}
				mb.feature_infos[j].max_value = max32(mb.feature_infos[j].max_value, feat.max_value)
				// inherit default_value from j
			}
			mb.feature_infos[j].flags |= (feat.flags & F_HAS_FALLBACK)
			mb.feature_infos[j].stage[0] = min(mb.feature_infos[j].stage[0], feat.stage[0])
			mb.feature_infos[j].stage[1] = min(mb.feature_infos[j].stage[1], feat.stage[1])
		}
		mb.feature_infos = mb.feature_infos[0 : j+1]
	}

	// allocate bits now
	nextBit := globalBitShift + 1

	for _, info := range mb.feature_infos {

		bitsNeeded := 0

		if (info.flags&F_GLOBAL) != 0 && info.max_value == 1 {
			// uses the global bit
			bitsNeeded = 0
		} else {
			// limit bits per feature.
			bitsNeeded = min(HB_OT_MAP_MAX_BITS, hb_bit_storage(info.max_value))
		}

		if info.max_value == 0 || nextBit+bitsNeeded > 32 {
			continue // feature disabled, or not enough bits.
		}

		var (
			found        = false
			featureIndex [2]int
			tables       = [2]*truetype.TableLayout{gsub, gpos}
		)
		for tableIndex, table := range tables {
			if requiredFeatureTag[tableIndex] == info.tag {
				requiredFeatureStage[tableIndex] = info.stage[tableIndex]
			}
			featureIndex[tableIndex] = findFeature(table, mb.script_index[tableIndex], mb.language_index[tableIndex], info.tag)
			found = found || featureIndex[tableIndex] != HB_OT_LAYOUT_NO_FEATURE_INDEX
		}
		if !found && (info.flags&F_GLOBAL_SEARCH) != 0 {
			for tableIndex, table := range tables {
				featureIndex[tableIndex] = hb_ot_layout_table_find_feature(table, info.tag)
				found = found || featureIndex[tableIndex] != HB_OT_LAYOUT_NO_FEATURE_INDEX
			}
		}
		if !found && info.flags&F_HAS_FALLBACK == 0 {
			continue
		}

		var map_ feature_map_t
		map_.tag = info.tag
		map_.index = featureIndex
		map_.stage = info.stage
		map_.auto_zwnj = info.flags&F_MANUAL_ZWNJ == 0
		map_.auto_zwj = info.flags&F_MANUAL_ZWJ == 0
		map_.random = info.flags&F_RANDOM != 0
		if (info.flags&F_GLOBAL) != 0 && info.max_value == 1 {
			// uses the global bit
			map_.shift = globalBitShift
			map_.mask = globalBitMask
		} else {
			map_.shift = nextBit
			map_.mask = (1 << (nextBit + bitsNeeded)) - (1 << nextBit)
			nextBit += bitsNeeded
			m.global_mask |= (info.default_value << map_.shift) & map_.mask
		}
		map_._1_mask = (1 << map_.shift) & map_.mask
		map_.needs_fallback = !found
		m.features = append(m.features, map_)
	}
	mb.feature_infos = mb.feature_infos[:0] // done with these

	mb.add_gsub_pause(nil)
	mb.add_gpos_pause(nil)

	for tableIndex := 0; tableIndex < 2; tableIndex++ {
		// collect lookup indices for features

		stage_index := 0
		last_num_lookups := 0
		for stage := 0; stage < mb.current_stage[tableIndex]; stage++ {
			if requiredFeatureIndex[tableIndex] != HB_OT_LAYOUT_NO_FEATURE_INDEX &&
				requiredFeatureStage[tableIndex] == stage {
				add_lookups(m, tableIndex, requiredFeatureIndex[tableIndex],
					key.variationsIndex[tableIndex], globalBitMask)
			}

			for i = 0; i < m.features.length; i++ {
				if m.features[i].stage[tableIndex] == stage {
					add_lookups(m, tableIndex,
						m.features[i].index[tableIndex],
						key.variationsIndex[tableIndex],
						m.features[i].mask,
						m.features[i].auto_zwnj,
						m.features[i].auto_zwj,
						m.features[i].random)
				}
			}

			// sort lookups and merge duplicates
			if last_num_lookups < m.lookups[tableIndex].length {
				m.lookups[tableIndex].qsort(last_num_lookups, m.lookups[tableIndex].length)

				j := last_num_lookups
				for i := j + 1; i < m.lookups[tableIndex].length; i++ {
					if m.lookups[tableIndex][i].index != m.lookups[tableIndex][j].index {
						j++
						m.lookups[tableIndex][j] = m.lookups[tableIndex][i]
					} else {
						m.lookups[tableIndex][j].mask |= m.lookups[tableIndex][i].mask
						m.lookups[tableIndex][j].auto_zwnj &= m.lookups[tableIndex][i].auto_zwnj
						m.lookups[tableIndex][j].auto_zwj &= m.lookups[tableIndex][i].auto_zwj
					}
				}
				m.lookups[tableIndex].shrink(j + 1)
			}

			last_num_lookups = m.lookups[tableIndex].length

			if stage_index < stages[tableIndex].length && stages[tableIndex][stage_index].index == stage {
				stage_map := m.stages[tableIndex].push()
				stage_map.last_lookup = last_num_lookups
				stage_map.pause_func = stages[tableIndex][stage_index].pause_func

				stage_index++
			}
		}
	}
}

type feature_map_t struct {
	tag            hb_tag_t /* should be first for our bsearch to work */
	index          [2]int   /* GSUB/GPOS */
	stage          [2]int   /* GSUB/GPOS */
	shift          int
	mask           hb_mask_t
	_1_mask        hb_mask_t /* mask for value=1, for quick access */
	needs_fallback bool      // = 1;
	auto_zwnj      bool      // = 1;
	auto_zwj       bool      // = 1;
	random         bool      // = 1;

	// int cmp (const hb_tag_t tag_) const
	// { return tag_ < tag ? -1 : tag_ > tag ? 1 : 0; }
}
type lookup_map_t struct {
	index     uint16
	auto_zwnj bool // = 1;
	auto_zwj  bool // = 1;
	random    bool // = 1;
	mask      hb_mask_t

	// HB_INTERNAL static int cmp (const void *pa, const void *pb)
	// {
	//   const lookup_map_t *a = (const lookup_map_t *) pa;
	//   const lookup_map_t *b = (const lookup_map_t *) pb;
	//   return a.index < b.index ? -1 : a.index > b.index ? 1 : 0;
	// }
}
type stage_map_t struct {
	last_lookup uint /* Cumulative */
	pause_func  pause_func_t
}
type hb_ot_map_t struct {
	chosen_script [2]hb_tag_t
	found_script  [2]bool

	// private:

	global_mask hb_mask_t

	features []feature_map_t   // sorted
	lookups  [2][]lookup_map_t /* GSUB/GPOS */
	stages   [2]stage_map_t    /* GSUB/GPOS */
	//   friend struct hb_ot_map_builder_t;

	//   void init ()
	//   {
	//     memset (this, 0, sizeof (*this));

	//     features.init ();
	//     for (uint tableIndex = 0; tableIndex < 2; tableIndex++)
	//     {
	//       lookups[tableIndex].init ();
	//       stages[tableIndex].init ();
	//     }
	//   }

	//   hb_mask_t get_global_mask () const { return global_mask; }

	//   hb_mask_t get_mask (hb_tag_t feature_tag, uint *shift = nil) const
	//   {
	//     const feature_map_t *map = features.bsearch (feature_tag);
	//     if (shift) *shift = map ? map.shift : 0;
	//     return map ? map.mask : 0;
	//   }

	//   bool needs_fallback (hb_tag_t feature_tag) const
	//   {
	//     const feature_map_t *map = features.bsearch (feature_tag);
	//     return map ? map.needs_fallback : false;
	//   }

	//   hb_mask_t get_1_mask (hb_tag_t feature_tag) const
	//   {
	//     const feature_map_t *map = features.bsearch (feature_tag);
	//     return map ? map._1_mask : 0;
	//   }

	//   uint get_feature_index (uint tableIndex, hb_tag_t feature_tag) const
	//   {
	//     const feature_map_t *map = features.bsearch (feature_tag);
	//     return map ? map.index[tableIndex] : HB_OT_LAYOUT_NO_FEATURE_INDEX;
	//   }

	//   uint get_feature_stage (uint tableIndex, hb_tag_t feature_tag) const
	//   {
	//     const feature_map_t *map = features.bsearch (feature_tag);
	//     return map ? map.stage[tableIndex] : UINT_MAX;
	//   }

	//   void get_stage_lookups (uint tableIndex, uint stage,
	// 			  const struct lookup_map_t **plookups, uint *lookup_count) const
	//   {
	//     if (unlikely (stage > stages[tableIndex].length))
	//     {
	//       *plookups = nil;
	//       *lookup_count = 0;
	//       return;
	//     }
	//     uint start = stage ? stages[tableIndex][stage - 1].last_lookup : 0;
	//     uint end   = stage < stages[tableIndex].length ? stages[tableIndex][stage].last_lookup : lookups[tableIndex].length;
	//     *plookups = end == start ? nil : &lookups[tableIndex][start];
	//     *lookup_count = end - start;
	//   }
}
