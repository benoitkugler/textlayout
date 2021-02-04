package common

import "sort"

// ported from harfbuzz/src/hb-aat-map.cc, hb-att-map.hh Copyright © 2018  Google, Inc. Behdad Esfahbod

type hb_aat_map_t struct {
	chain_flags []Mask
}

type aat_feature_info_t struct {
	type_        hb_aat_layout_feature_type_t
	setting      hb_aat_layout_feature_selector_t
	is_exclusive bool

	// 	 /* compares type & setting only, not is_exclusive flag or seq number */
	// 	 int cmp (const aat_feature_info_t& f) const
	// 	 {
	// 	   return (f.type != type) ? (f.type < type ? -1 : 1) :
	// 		  (f.setting != setting) ? (f.setting < setting ? -1 : 1) : 0;
	// 	 }
	//    };
}

const selMask = ^hb_aat_layout_feature_selector_t(1)

func cmpAATFeatureInfo(a, b aat_feature_info_t) bool {
	if a.type_ != b.type_ {
		return a.type_ < b.type_
	}
	if !a.is_exclusive && (a.setting&selMask) != (b.setting&selMask) {
		return a.setting < b.setting
	}
	return false
}

type hb_aat_map_builder_t struct {
	face     Face
	features []aat_feature_info_t // sorted after compilation
}

func (mb *hb_aat_map_builder_t) add_feature(tag hb_tag_t, value uint32) {
	feat := mb.face.getFeatTable()
	if !feat {
		return
	}

	if tag == newTag('a', 'a', 'l', 't') {
		if !face.table.feat.exposes_feature(HB_AAT_LAYOUT_FEATURE_TYPE_CHARACTER_ALTERNATIVES) {
			return
		}
		info := aat_feature_info_t{
			type_:        HB_AAT_LAYOUT_FEATURE_TYPE_CHARACTER_ALTERNATIVES,
			setting:      hb_aat_layout_feature_selector_t(value),
			is_exclusive: true,
		}
		mb.features = append(mb.features, info)
		return
	}

	mapping := hb_aat_layout_find_feature_mapping(tag)
	if mapping == nil {
		return
	}

	feature := &face.table.feat.get_feature(mapping.aatFeatureType)
	if !feature.has_data() {
		/* Special case: Chain::compile_flags will fall back to the deprecated version of
		 * small-caps if necessary, so we need to check for that possibility.
		 * https://github.com/harfbuzz/harfbuzz/issues/2307 */
		if mapping.aatFeatureType == HB_AAT_LAYOUT_FEATURE_TYPE_LOWER_CASE &&
			mapping.selectorToEnable == HB_AAT_LAYOUT_FEATURE_SELECTOR_LOWER_CASE_SMALL_CAPS {
			feature = &face.table.feat.get_feature(HB_AAT_LAYOUT_FEATURE_TYPE_LETTER_CASE)
			if !feature.has_data() {
				return
			}
		} else {
			return
		}
	}

	var info aat_feature_info_t
	info.type_ = mapping.aatFeatureType
	if value != 0 {
		info.setting = mapping.selectorToEnable
	} else {
		info.setting = mapping.selectorToDisable
	}
	info.is_exclusive = feature.is_exclusive()
	mb.features = append(mb.features, info)
}

func (mb *hb_aat_map_builder_t) compile(m *hb_aat_map_t) {
	// sort features and merge duplicates
	if len(mb.features) != 0 {
		sort.SliceStable(mb.features, func(i, j int) bool {
			return cmpAATFeatureInfo(mb.features[i], mb.features[j])
		})
		j := 0
		for i := 1; i < len(mb.features); i++ {
			/* Nonexclusive feature selectors come in even/odd pairs to turn a setting on/off
			* respectively, so we mask out the low-order bit when checking for "duplicates"
			* (selectors referring to the same feature setting) here. */
			if mb.features[i].type_ != mb.features[j].type_ ||
				(!mb.features[i].is_exclusive && ((mb.features[i].setting & selMask) != (mb.features[j].setting & selMask))) {
				j++
				mb.features[j] = mb.features[i]
			}
		}
		mb.features = mb.features[:j+1]
	}

	mb.hb_aat_layout_compile_map(m)
}
