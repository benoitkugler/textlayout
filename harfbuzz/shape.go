package harfbuzz

import (
	"fmt"
	"sync"
)

// ported from harfbuzz/src/hb-shape.cc, harfbuzz/src/hb-shape-plan.cc
// Copyright © 2009  Red Hat, Inc.
// Copyright © 2012  Google, Inc.
// Red Hat Author(s): Behdad Esfahbod
// Google Author(s): Behdad Esfahbod

/**
 * SECTION:hb-shape
 * @title: hb-shape
 * @short_description: Conversion of text strings into positioned glyphs
 * @include: hb.h
 *
 * Shaping is the central operation of HarfBuzz. Shaping operates on buffers,
 * which are sequences of Unicode characters that use the same font and have
 * the same text direction, script, and language. After shaping the buffer
 * contains the output glyphs and their positions.
 **/

//  static const char *nil_shaper_list[] = {nullptr};

// func create() []string {
// 	var out []string
// 	shapers := _hb_shapers_get()
// 	for _, s := range shapers {
// 		out = append(out, s.name)
// 	}
// 	return out
// }

// Shapes `buffer` using `font`, turning its Unicode characters content to
// positioned glyphs. If `features` is not empty, it will be used to control the
// features applied during shaping. If two features have the same tag but
// overlapping ranges the value of the feature with the higher index takes
// precedence.
// The two slices returned have the same length.
// Note that the returned slices point to memory owned by the buffer,
// and are thus only valid until the buffer content is modified.
// Return value: false if all shapers failed, true otherwise
func (buffer *Buffer) Shape(font *Font, features []Feature) ([]GlyphInfo, []GlyphPosition, bool) {
	shape_plan := hb_shape_plan_create_cached2(font.Face, &buffer.Props,
		features, font.coords, nil)
	res := shape_plan.hb_shape_plan_execute(font, buffer, features)
	return buffer.info, buffer.pos, res
}

// shaper shapes a string of runes.
// Depending on the font used, different shapers will be choosen.
type shaper interface {
	shape(*ShapePlan, *Font, *Buffer, []Feature) bool
}

// use interface since equality check is needed
type hb_shape_func_t = func(shape_plan *ShapePlan,
	font *Font, buffer *Buffer, features []Feature) bool

type hb_ot_shape_plan_key_t = [2]int

func hb_ot_shape_plan_key_t_init(face Face, coords []float32) hb_ot_shape_plan_key_t {
	gsub := face.hb_ot_layout_table_find_feature_variations(HB_OT_TAG_GSUB, coords)
	gpos := face.hb_ot_layout_table_find_feature_variations(HB_OT_TAG_GPOS, coords)
	return [2]int{gsub, gpos}
}

type hb_shape_plan_key_t struct {
	props SegmentProperties

	userFeatures []Feature // len num_user_features

	ot hb_ot_shape_plan_key_t

	shaper_func hb_shape_func_t
	shaper_name string
}

/*
 * hb_shape_plan_key_t
 */

func (plan *hb_shape_plan_key_t) init(copy bool,
	face Face, props *SegmentProperties,
	userFeatures []Feature, coords []float32, shaperList []string) {
	// TODO: for now, shaperList is ignored

	plan.props = *props
	if !copy {
		plan.userFeatures = userFeatures
	} else {
		plan.userFeatures = append([]Feature(nil), userFeatures...)
		/* Make start/end uniform to easier catch bugs. */
		for i := range plan.userFeatures {
			if plan.userFeatures[i].Start != FeatureGlobalStart {
				plan.userFeatures[i].Start = 1
			}
			if plan.userFeatures[i].End != FeatureGlobalEnd {
				plan.userFeatures[i].End = 2
			}
		}
	}
	plan.ot = hb_ot_shape_plan_key_t_init(face, coords)

	// Choose shaper.

	if face.data.graphite2 {
		this.shaper_func = _hb_graphite2_shape
		this.shaper_name = "graphite2"
	} else if face.data.ot {
		this.shaper_func = _hb_ot_shape
		this.shaper_name = "ot"
	} else if face.data.fallback {
		this.shaper_func = _hb_fallback_shape
		this.shaper_name = "fallback"
	}
}

func (plan hb_shape_plan_key_t) user_features_match(other hb_shape_plan_key_t) bool {
	if len(plan.userFeatures) != len(other.userFeatures) {
		return false
	}
	for i, feat := range plan.userFeatures {
		if feat.Tag != other.userFeatures[i].Tag || feat.Value != other.userFeatures[i].Value ||
			(feat.Start == FeatureGlobalStart && feat.End == FeatureGlobalEnd) !=
				(other.userFeatures[i].Start == FeatureGlobalStart && other.userFeatures[i].End == FeatureGlobalEnd) {
			return false
		}
	}
	return true
}

func (plan hb_shape_plan_key_t) equal(other hb_shape_plan_key_t) bool {
	return plan.props == other.props &&
		plan.user_features_match(other) &&
		plan.ot == other.ot && plan.shaper_func == other.shaper_func
}

// Shape plans are an internal mechanism. Each plan contains state
// describing how HarfBuzz will shape a particular text segment, based on
// the combination of segment properties and the capabilities in the
// font face in use.
//
// Shape plans are not used for shaping directly, but can be queried to
// access certain information about how shaping will perform, given a set
// of specific input parameters (script, language, direction, features,
// etc.).
//
// Most client programs will not need to deal with shape plans directly.
type ShapePlan struct {
	face_unsafe Face
	key         hb_shape_plan_key_t
	ot          hb_ot_shape_plan_t
}

/**
 * hb_shape_plan_create: (Xconstructor)
 * @face: #Face to use
 * @props: The #SegmentProperties of the segment
 * @userFeatures: (array length=num_user_features): The list of user-selected features
 * @num_user_features: The number of user-selected features
 * @shaperList: (array zero-terminated=1): List of shapers to try
 *
 * Constructs a shaping plan for a combination of @face, @userFeatures, @props,
 * and @shaperList.
 *
 * Return value: (transfer full): The shaping plan
 *
 * Since: 0.9.7
 **/
func hb_shape_plan_create(face Face, props *SegmentProperties,
	userFeatures []Feature, shaperList []string) *ShapePlan {
	return hb_shape_plan_create2(face, props, userFeatures, nil, shaperList)
}

/**
 * hb_shape_plan_create2: (Xconstructor)
 * @face: #Face to use
 * @props: The #SegmentProperties of the segment
 * @userFeatures: (array length=num_user_features): The list of user-selected features
 * @num_user_features: The number of user-selected features
 * @coords: (array length=num_coords): The list of variation-space coordinates
 * @num_coords: The number of variation-space coordinates
 * @shaperList: (array zero-terminated=1): List of shapers to try
 *
 * The variable-font version of #hb_shape_plan_create.
 * Constructs a shaping plan for a combination of @face, @userFeatures, @props,
 * and @shaperList, plus the variation-space coordinates @coords.
 *
 * Return value: (transfer full): The shaping plan
 *
 * Since: 1.4.0
 **/

func hb_shape_plan_create2(face Face, props *SegmentProperties,
	userFeatures []Feature, coords []float32, shaperList []string) *ShapePlan {
	if DebugMode {
		fmt.Printf("shape plan: face:%p num_features:%d num_coords=%d shaperList:%s", face, len(userFeatures), len(coords), shaperList)
	}

	var shape_plan ShapePlan

	shape_plan.face_unsafe = face

	shape_plan.key.init(true, face, props, userFeatures, coords, shaperList)
	shape_plan.ot.init0(face, &shape_plan.key)

	return &shape_plan
}

/**
 * hb_shape_plan_get_empty:
 *
 * Fetches the singleton empty shaping plan.
 *
 * Return value: (transfer full): The empty shaping plan
 *
 * Since: 0.9.7
 **/
//  ShapePlan *
//  hb_shape_plan_get_empty ()
//  {
//    return const_cast<ShapePlan *> (&Null (ShapePlan));
//  }

func (shape_plan *ShapePlan) _hb_shape_plan_execute_internal(font *Font, buffer *Buffer, features []Feature) bool {
	if DebugMode {
		fmt.Printf("execute shape plan num_features=%d shaper_func=%p, shaper_name=%s",
			len(features), shape_plan.key.shaper_func, shape_plan.key.shaper_name)
	}
	//    assert (shape_plan.face_unsafe == font.face);
	//    assert (hb_segment_properties_equal (&shape_plan.key.props, &buffer.props));

	if shape_plan.key.shaper_func == _hb_graphite2_shape {
		return font.data.graphite2 && _hb_graphite2_shape(shape_plan, font, buffer, features)
	} else if shape_plan.key.shaper_func == _hb_ot_shape {
		return font.data.ot && _hb_ot_shape(shape_plan, font, buffer, features)
	} else if shape_plan.key.shaper_func == _hb_fallback_shape {
		return font.data.fallback && _hb_fallback_shape(shape_plan, font, buffer, features)
	}

	return false
}

/**
 * hb_shape_plan_execute:
 * @shape_plan: A shaping plan
 * @font: The #Font to use
 * @buffer: The #Buffer to work upon
 * @features: (array length=num_features): Features to enable
 * @num_features: The number of features to enable
 *
 * Executes the given shaping plan on the specified buffer, using
 * the given @font and @features.
 *
 * Return value: %true if success, %false otherwise.
 *
 * Since: 0.9.7
 **/
func (shape_plan *ShapePlan) hb_shape_plan_execute(
	font *Font, buffer *Buffer,
	features []Feature) bool {
	ret := shape_plan._hb_shape_plan_execute_internal(font, buffer, features)

	if ret && buffer.content_type == HB_BUFFER_CONTENT_TYPE_UNICODE {
		buffer.content_type = HB_BUFFER_CONTENT_TYPE_GLYPHS
	}

	return ret
}

/*
 * Caching
 */

/**
 * hb_shape_plan_create_cached:
 * @face: #Face to use
 * @props: The #SegmentProperties of the segment
 * @userFeatures: (array length=num_user_features): The list of user-selected features
 * @num_user_features: The number of user-selected features
 * @shaperList: (array zero-terminated=1): List of shapers to try
 *
 * Creates a cached shaping plan suitable for reuse, for a combination
 * of @face, @userFeatures, @props, and @shaperList.
 *
 * Return value: (transfer full): The shaping plan
 *
 * Since: 0.9.7
 **/
func hb_shape_plan_create_cached(face Face,
	props *SegmentProperties,
	userFeatures []Feature, shaperList []string) *ShapePlan {
	return hb_shape_plan_create_cached2(face, props, userFeatures, nil, shaperList)
}

var (
	planCache     map[Face][]*ShapePlan
	planCacheLock sync.Mutex
)

/**
 * hb_shape_plan_create_cached2:
 * @face: #Face to use
 * @props: The #SegmentProperties of the segment
 * @userFeatures: (array length=num_user_features): The list of user-selected features
 * @num_user_features: The number of user-selected features
 * @coords: (array length=num_coords): The list of variation-space coordinates
 * @num_coords: The number of variation-space coordinates
 * @shaperList: (array zero-terminated=1): List of shapers to try
 *
 * The variable-font version of #hb_shape_plan_create_cached.
 * Creates a cached shaping plan suitable for reuse, for a combination
 * of @face, @userFeatures, @props, and @shaperList, plus the
 * variation-space coordinates @coords.
 *
 * Return value: (transfer full): The shaping plan
 *
 * Since: 1.4.0
 **/
func hb_shape_plan_create_cached2(face Face,
	props *SegmentProperties,
	userFeatures []Feature, coords []float32, shaperList []string) *ShapePlan {
	if DebugMode {
		fmt.Printf("shape plan: face:%p num_features:%d shaperList:%s", face, len(userFeatures), shaperList)
	}

	var key hb_shape_plan_key_t
	key.init(false, face, props, userFeatures, coords, shaperList)

	planCacheLock.Lock()
	defer planCacheLock.Unlock()

	plans := planCache[face]

	for _, plan := range plans {
		if plan.key.equal(key) {
			if DebugMode {
				fmt.Println(plan, "fulfilled from cache")
			}
			return plan
		}
	}
	plan := hb_shape_plan_create2(face, props, userFeatures, coords, shaperList)

	plans = append(plans, plan)
	planCache[face] = plans
	if DebugMode {
		fmt.Println(plan, "inserted into cache")
	}

	return plan
}
