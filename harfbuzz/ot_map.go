package harfbuzz

import (
	"fmt"
	"math"
	"math/bits"
	"sort"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-map.cc, hb-ot-map.hh Copyright © 2009,2010  Red Hat, Inc. 2010,2011,2013  Google, Inc. Behdad Esfahbod

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
	HB_OT_MAP_MAX_BITS = 8
	otMapMaxValue      = (1 << HB_OT_MAP_MAX_BITS) - 1
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

	face  Face
	props SegmentProperties

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

func new_hb_ot_map_builder_t(face Face, props SegmentProperties) hb_ot_map_builder_t {
	var out hb_ot_map_builder_t

	out.face = face
	out.props = props

	/* Fetch script/language indices for GSUB/GPOS.  We need these later to skip
	* features not available in either table and not waste precious bits for them. */

	script_tags, language_tags := hb_ot_tags_from_script_and_language(props.Script, props.Language)

	gsub, gpos := face.get_gsubgpos_table() // TODO: check if its nil

	out.script_index[0], out.chosen_script[0], out.found_script[0] = selectScript(gsub, script_tags)
	out.language_index[0], _ = selectLanguage(gsub, out.script_index[0], language_tags)

	out.script_index[1], out.chosen_script[1], out.found_script[1] = selectScript(gpos, script_tags)
	out.language_index[1], _ = selectLanguage(gpos, out.script_index[1], language_tags)

	return out
}

func (mb *hb_ot_map_builder_t) add_feature_ext(tag hb_tag_t, flags hb_ot_map_feature_flags_t, value uint32) {
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

type pause_func_t = func(plan *hb_ot_shape_plan_t, font *Font, buffer *Buffer)

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

func (mb *hb_ot_map_builder_t) enable_feature_ext(tag hb_tag_t, flags hb_ot_map_feature_flags_t, value uint32) {
	mb.add_feature_ext(tag, F_GLOBAL|flags, value)
}
func (mb *hb_ot_map_builder_t) enable_feature(tag hb_tag_t)  { mb.enable_feature_ext(tag, F_NONE, 1) }
func (mb *hb_ot_map_builder_t) add_feature(tag hb_tag_t)     { mb.add_feature_ext(tag, F_NONE, 1) }
func (mb *hb_ot_map_builder_t) disable_feature(tag hb_tag_t) { mb.add_feature_ext(tag, F_GLOBAL, 0) }

func (mb *hb_ot_map_builder_t) compile(m *hb_ot_map_t, key hb_ot_shape_plan_key_t) {
	globalBitMask := HB_GLYPH_FLAG_DEFINED + 1
	globalBitShift := bits.OnesCount32(uint32(HB_GLYPH_FLAG_DEFINED))

	m.global_mask = globalBitMask

	var (
		requiredFeatureIndex [2]uint16 // HB_OT_LAYOUT_NO_FEATURE_INDEX for empty
		requiredFeatureTag   [2]hb_tag_t
		/* We default to applying required feature in stage 0. If the required
		* feature has a tag that is known to the shaper, we apply the required feature
		* in the stage for that tag. */
		requiredFeatureStage [2]int
	)

	gsub, gpos := mb.face.get_gsubgpos_table()
	tables := [2]*truetype.TableLayout{gsub, gpos}

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
				mb.feature_infos[j].max_value = Max32(mb.feature_infos[j].max_value, feat.max_value)
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
			bitsNeeded = min(HB_OT_MAP_MAX_BITS, bitStorage(info.max_value))
		}

		if info.max_value == 0 || nextBit+bitsNeeded > 32 {
			continue // feature disabled, or not enough bits.
		}

		var (
			found        = false
			featureIndex [2]uint16
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

	for tableIndex, table := range tables {
		// collect lookup indices for features

		stage_index := 0
		lastNumLookups := 0
		for stage := 0; stage < mb.current_stage[tableIndex]; stage++ {
			if requiredFeatureIndex[tableIndex] != HB_OT_LAYOUT_NO_FEATURE_INDEX &&
				requiredFeatureStage[tableIndex] == stage {
				m.add_lookups(table, tableIndex, requiredFeatureIndex[tableIndex],
					key[tableIndex], globalBitMask, true, true, false)
			}

			for _, feat := range m.features {
				if feat.stage[tableIndex] == stage {
					m.add_lookups(table, tableIndex,
						feat.index[tableIndex],
						key[tableIndex],
						feat.mask,
						feat.auto_zwnj,
						feat.auto_zwj,
						feat.random)
				}
			}

			// sort lookups and merge duplicates

			if ls := m.lookups[tableIndex]; lastNumLookups < len(ls) {
				view := ls[lastNumLookups:]
				sort.Slice(view, func(i, j int) bool { return view[i].index < view[j].index })

				j := lastNumLookups
				for i := j + 1; i < len(ls); i++ {
					if ls[i].index != ls[j].index {
						j++
						ls[j] = ls[i]
					} else {
						ls[j].mask |= ls[i].mask
						ls[j].auto_zwnj = ls[j].auto_zwnj && ls[i].auto_zwnj
						ls[j].auto_zwj = ls[j].auto_zwj && ls[i].auto_zwj
					}
				}
				m.lookups[tableIndex] = m.lookups[tableIndex][:j+1]
			}

			lastNumLookups = len(m.lookups[tableIndex])

			if stage_index < len(mb.stages[tableIndex]) && mb.stages[tableIndex][stage_index].index == stage {
				stage_map := stage_map_t{
					last_lookup: lastNumLookups,
					pause_func:  mb.stages[tableIndex][stage_index].pause_func,
				}
				m.stages[tableIndex] = append(m.stages[tableIndex], stage_map)
				stage_index++
			}
		}
	}
}

type feature_map_t struct {
	tag            hb_tag_t  /* should be first for our bsearch to work */
	index          [2]uint16 /* GSUB/GPOS */
	stage          [2]int    /* GSUB/GPOS */
	shift          int
	mask           Mask
	_1_mask        Mask /* mask for value=1, for quick access */
	needs_fallback bool // = 1;
	auto_zwnj      bool // = 1;
	auto_zwj       bool // = 1;
	random         bool // = 1;

	// int cmp (const hb_tag_t tag_) const
	// { return tag_ < tag ? -1 : tag_ > tag ? 1 : 0; }
}

func bsearchFeature(features []feature_map_t, tag hb_tag_t) *feature_map_t {
	low, high := 0, len(features)
	for low < high {
		mid := low + (high-low)/2 // avoid overflow when computing mid
		p := features[mid].tag
		if tag < p {
			high = mid
		} else if tag > p {
			low = mid + 1
		} else {
			return &features[mid]
		}
	}
	return nil
}

type lookup_map_t struct {
	index     uint16
	auto_zwnj bool // = 1;
	auto_zwj  bool // = 1;
	random    bool // = 1;
	mask      Mask

	// HB_INTERNAL static int cmp (const void *pa, const void *pb)
	// {
	//   const lookup_map_t *a = (const lookup_map_t *) pa;
	//   const lookup_map_t *b = (const lookup_map_t *) pb;
	//   return a.index < b.index ? -1 : a.index > b.index ? 1 : 0;
	// }
}

type stage_map_t struct {
	last_lookup int /* Cumulative */
	pause_func  pause_func_t
}

type hb_ot_map_t struct {
	chosen_script [2]hb_tag_t
	found_script  [2]bool

	global_mask Mask

	features []feature_map_t   // sorted
	lookups  [2][]lookup_map_t /* GSUB/GPOS */
	stages   [2][]stage_map_t  /* GSUB/GPOS */
}

//   friend struct hb_ot_map_builder_t;

func (m *hb_ot_map_t) needs_fallback(featureTag hb_tag_t) bool {
	if ma := bsearchFeature(m.features, featureTag); ma != nil {
		return ma.needs_fallback
	}
	return false
}

func (m *hb_ot_map_t) get_mask(featureTag hb_tag_t) (Mask, int) {
	if ma := bsearchFeature(m.features, featureTag); ma != nil {
		return ma.mask, ma.shift
	}
	return 0, 0
}

func (m *hb_ot_map_t) get_1_mask(featureTag hb_tag_t) Mask {
	if ma := bsearchFeature(m.features, featureTag); ma != nil {
		return ma._1_mask
	}
	return 0
}

func (m *hb_ot_map_t) get_feature_index(tableIndex int, featureTag hb_tag_t) uint16 {
	if ma := bsearchFeature(m.features, featureTag); ma != nil {
		return ma.index[tableIndex]
	}
	return HB_OT_LAYOUT_NO_FEATURE_INDEX
}

func (m *hb_ot_map_t) get_feature_stage(tableIndex int, featureTag hb_tag_t) int {
	if ma := bsearchFeature(m.features, featureTag); ma != nil {
		return ma.stage[tableIndex]
	}
	return math.MaxInt32
}

func (m *hb_ot_map_t) get_stage_lookups(tableIndex, stage int) []lookup_map_t {
	if stage > len(m.stages[tableIndex]) {
		return nil
	}
	start, end := 0, len(m.lookups[tableIndex])
	if stage != 0 {
		start = m.stages[tableIndex][stage-1].last_lookup
	}
	if stage < len(m.stages[tableIndex]) {
		end = m.stages[tableIndex][stage].last_lookup
	}
	return m.lookups[tableIndex][start:end]
}

func (m *hb_ot_map_t) add_lookups(table *truetype.TableLayout, tableIndex int, featureIndex uint16, variationsIndex int,
	mask Mask, autoZwnj, autoZwj, random bool) {
	lookupIndices := getFeatureLookupsWithVar(table, featureIndex, variationsIndex)
	for _, lookupInd := range lookupIndices {
		lookup := lookup_map_t{
			mask:      mask,
			index:     lookupInd,
			auto_zwnj: autoZwnj,
			auto_zwj:  autoZwj,
			random:    random,
		}
		m.lookups[tableIndex] = append(m.lookups[tableIndex], lookup)
	}
}

// apply the GSUB table
func (m *hb_ot_map_t) substitute(plan *hb_ot_shape_plan_t, font *Font, buffer *Buffer) {
	if debugMode {
		fmt.Println("SUBSTITUTE - start table GSUB")
	}
	m.apply(0, plan, font, buffer)
	if debugMode {
		fmt.Println("SUBSTITUTE - end table GSUB")
	}
}

func (m *hb_ot_map_t) apply(proxy otProxy, plan *hb_ot_shape_plan_t, font *Font, buffer *Buffer) {
	tableIndex := proxy.tableIndex
	i := 0
	c := new_hb_ot_apply_context_t(tableIndex, font, buffer)
	c.recurse_func = proxy.recurse_func

	for _, stage := range m.stages[tableIndex] {
		for ; i < stage.last_lookup; i++ {
			lookupIndex := m.lookups[tableIndex][i].index

			if debugMode {
				fmt.Printf("APPLY - start lookup %d", lookupIndex)
			}

			c.lookupIndex = lookupIndex
			c.set_lookup_mask(m.lookups[tableIndex][i].mask)
			c.set_auto_zwj(m.lookups[tableIndex][i].auto_zwj)
			c.set_auto_zwnj(m.lookups[tableIndex][i].auto_zwnj)
			if m.lookups[tableIndex][i].random {
				c.random = true
				buffer.unsafeToBreakAll()
			}
			c.apply_string(proxy.table.get_lookup(lookupIndex), &proxy.accels[lookupIndex])

			if debugMode {
				fmt.Printf("APPLY - end lookup %d", lookupIndex)
			}

		}

		if stage.pause_func != nil {
			buffer.clearOutput()
			stage.pause_func(plan, font, buffer)
		}
	}
}
