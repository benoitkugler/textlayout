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
 * Shaping is the central operation of HarfBuzz. Shaping operates on buffers,
 * which are sequences of Unicode characters that use the same font and have
 * the same text direction, script, and language. After shaping the buffer
 * contains the output glyphs and their positions.
 **/

// Shapes `buffer` using `font`, turning its Unicode characters content to
// positioned glyphs. If `features` is not empty, it will be used to control the
// features applied during shaping. If two features have the same tag but
// overlapping ranges the value of the feature with the higher index takes
// precedence.
func (buffer *Buffer) Shape(font *Font, features []Feature) {
	shape_plan := hb_shape_plan_create_cached2(font, buffer.Props, features, font.coords, nil)
	shape_plan.hb_shape_plan_execute(font, buffer, features)
}

// shaper shapes a string of runes.
// Depending on the font used, different shapers will be choosen.
type shaper interface {
	shape(*Font, *Buffer, []Feature)
}

// use interface since equality check is needed
type hb_shape_func_t = func(shape_plan *ShapePlan,
	font *Font, buffer *Buffer, features []Feature) bool

type hb_shape_plan_key_t struct {
	props SegmentProperties

	userFeatures []Feature // len num_user_features

	// ot hb_ot_shape_plan_key_t

	shaper shaper
}

/*
 * hb_shape_plan_key_t
 */

func (plan *hb_shape_plan_key_t) init(copy bool,
	font *Font, props SegmentProperties,
	userFeatures []Feature, coords []float32, shaperList []string) {
	// TODO: for now, shaperList is ignored

	plan.props = props
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

	// Choose shaper.

	if _, ok := font.face.(FaceGraphite); ok {
		plan.shaper = shaperGraphite{} // TODO:
	} else if font.otTables != nil {
		plan.shaper = newShaperOpentype(font.otTables, props, userFeatures, coords)
	} else {
		plan.shaper = shaperFallback{}
	}
}

func (plan hb_shape_plan_key_t) userFeaturesMatch(other hb_shape_plan_key_t) bool {
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
		plan.userFeaturesMatch(other) && plan.shaper == other.shaper // TODO: check equality condition
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
	// ot          hb_ot_shape_plan_t
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
func hb_shape_plan_create(font *Font, props SegmentProperties,
	userFeatures []Feature, shaperList []string) *ShapePlan {
	return hb_shape_plan_create2(font, props, userFeatures, nil, shaperList)
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

func hb_shape_plan_create2(font *Font, props SegmentProperties,
	userFeatures []Feature, coords []float32, shaperList []string) *ShapePlan {
	if debugMode {
		fmt.Printf("shape plan: face:%p num_features:%d num_coords=%d shaperList:%s", &font.face, len(userFeatures), len(coords), shaperList)
	}

	var shape_plan ShapePlan

	shape_plan.face_unsafe = font.face

	shape_plan.key.init(true, font, props, userFeatures, coords, shaperList)

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

// Executes the given shaping plan on the specified `buffer`, using
// the given `font` and `features`.
func (shape_plan *ShapePlan) hb_shape_plan_execute(font *Font, buffer *Buffer, features []Feature) {
	if debugMode {
		fmt.Printf("execute shape plan num_features=%d shaper_type=%T", len(features), shape_plan.key.shaper)
		//    assert (shape_plan.face_unsafe == font.face);
		//    assert (hb_segment_properties_equal (&shape_plan.key.props, &buffer.props));
	}

	shape_plan.key.shaper.shape(font, buffer, features)
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
func hb_shape_plan_create_cached(font *Font, props SegmentProperties,
	userFeatures []Feature, shaperList []string) *ShapePlan {
	return hb_shape_plan_create_cached2(font, props, userFeatures, nil, shaperList)
}

var (
	planCache     = map[Face][]*ShapePlan{}
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
func hb_shape_plan_create_cached2(font *Font,
	props SegmentProperties,
	userFeatures []Feature, coords []float32, shaperList []string) *ShapePlan {
	if debugMode {
		fmt.Printf("shape plan: face:%p num_features:%d shaperList:%s", &font.face, len(userFeatures), shaperList)
	}

	var key hb_shape_plan_key_t
	key.init(false, font, props, userFeatures, coords, shaperList)

	planCacheLock.Lock()
	defer planCacheLock.Unlock()

	plans := planCache[font.face]

	for _, plan := range plans {
		if plan.key.equal(key) {
			if debugMode {
				fmt.Println(plan, "fulfilled from cache")
			}
			return plan
		}
	}
	plan := hb_shape_plan_create2(font, props, userFeatures, coords, shaperList)

	plans = append(plans, plan)
	planCache[font.face] = plans
	if debugMode {
		fmt.Println(plan, "inserted into cache")
	}

	return plan
}
