package harfbuzz

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// Support functions for OpenType shaping related queries.
// ported from src/hb-ot-shape.cc Copyright © 2009,2010  Red Hat, Inc. 2010,2011,2012  Google, Inc. Behdad Esfahbod

/*
 * GSUB/GPOS feature query and enumeration interface
 */

const (
	// Special value for script index indicating unsupported script.
	HB_OT_LAYOUT_NO_SCRIPT_INDEX = 0xFFFF
	// Special value for feature index indicating unsupported feature.
	HB_OT_LAYOUT_NO_FEATURE_INDEX = 0xFFFF
	// Special value for language index indicating default or unsupported language.
	HB_OT_LAYOUT_DEFAULT_LANGUAGE_INDEX = 0xFFFF
	// Special value for variations index indicating unsupported variation.
	HB_OT_LAYOUT_NO_VARIATIONS_INDEX = 0xFFFFFFFF
)

type hb_ot_shape_planner_t struct {
	face                             hb_face_t
	props                            hb_segment_properties_t
	map_                             hb_ot_map_builder_t
	aat_map                          hb_aat_map_builder_t
	apply_morx                       bool
	script_zero_marks                bool
	script_fallback_mark_positioning bool
	shaper                           hb_ot_complex_shaper_t
}

func new_hb_ot_shape_planner_t(face hb_face_t, props hb_segment_properties_t) *hb_ot_shape_planner_t {
	var out hb_ot_shape_planner_t
	out.map_ = new_hb_ot_map_builder_t(face, props)
	out.aat_map = hb_aat_map_builder_t{face: face}

	/* https://github.com/harfbuzz/harfbuzz/issues/2124 */
	_, gsub := face.get_gsubgpos_table()
	out.apply_morx = hb_aat_layout_has_substitution(face) && (props.direction.isHorizontal() || gsub == nil)

	out.shaper = hb_ot_shape_complex_categorize(out)

	zwm, fb := out.shaper.marksBehavior()
	out.script_zero_marks = zwm != HB_OT_SHAPE_ZERO_WIDTH_MARKS_NONE
	out.script_fallback_mark_positioning = fb

	/* https://github.com/harfbuzz/harfbuzz/issues/1528 */
	if out.apply_morx && out.shaper != &_hb_ot_complex_shaper_default {
		out.shaper = &_hb_ot_complex_shaper_dumber
	}
	return &out
}

func (planner *hb_ot_shape_planner_t) compile(plan *hb_ot_shape_plan_t, key hb_ot_shape_plan_key_t) {
	plan.props = planner.props
	plan.shaper = planner.shaper
	planner.map_.compile(&plan.map_, key)
	if planner.apply_morx {
		planner.aat_map.compile(&plan.aat_map)
	}

	plan.frac_mask = plan.map_.get_1_mask(newTag('f', 'r', 'a', 'c'))
	plan.numr_mask = plan.map_.get_1_mask(newTag('n', 'u', 'm', 'r'))
	plan.dnom_mask = plan.map_.get_1_mask(newTag('d', 'n', 'o', 'm'))
	plan.has_frac = plan.frac_mask != 0 || (plan.numr_mask != 0 && plan.dnom_mask != 0)

	plan.rtlm_mask = plan.map_.get_1_mask(newTag('r', 't', 'l', 'm'))
	plan.has_vert = plan.map_.get_1_mask(newTag('v', 'e', 'r', 't')) != 0

	kern_tag := newTag('v', 'k', 'r', 'n')
	if planner.props.direction.isHorizontal() {
		kern_tag = newTag('k', 'e', 'r', 'n')
	}

	plan.kern_mask, _ = plan.map_.get_mask(kern_tag)
	plan.requested_kerning = plan.kern_mask != 0
	plan.trak_mask, _ = plan.map_.get_mask(newTag('t', 'r', 'a', 'k'))
	plan.requested_tracking = plan.trak_mask != 0

	has_gpos_kern := plan.map_.get_feature_index(1, kern_tag) != HB_OT_LAYOUT_NO_FEATURE_INDEX
	disable_gpos := plan.shaper.gpos_tag && plan.shaper.gpos_tag != plan.map_.chosen_script[1]

	// Decide who provides glyph classes. GDEF or Unicode.
	if planner.face.getGDEF().Class == nil {
		plan.fallback_glyph_classes = true
	}

	// Decide who does substitutions. GSUB, morx, or fallback.
	plan.apply_morx = planner.apply_morx

	//  Decide who does positioning. GPOS, kerx, kern, or fallback.
	_, hasAatPositioning := planner.face.getKerx()
	if hasAatPositioning {
		plan.apply_kerx = true
	} else if _, gpos := planner.face.get_gsubgpos_table(); !planner.apply_morx && !disable_gpos && gpos != nil {
		plan.apply_gpos = true
	}

	if !plan.apply_kerx && (!has_gpos_kern || !plan.apply_gpos) {
		// apparently Apple applies kerx if GPOS kern was not applied.
		if hasAatPositioning {
			plan.apply_kerx = true
		} else if kerns := planner.face.getKerns(); kerns != nil {
			plan.apply_kern = true
		}
	}

	plan.zero_marks = planner.script_zero_marks && !plan.apply_kerx &&
		(!plan.apply_kern || !planner.face.hasMachineKerning())
	plan.has_gpos_mark = plan.map_.get_1_mask(newTag('m', 'a', 'r', 'k')) != 0

	plan.adjust_mark_positioning_when_zeroing = !plan.apply_gpos && !plan.apply_kerx &&
		(!plan.apply_kern || !planner.face.hasCrossKerning())

	plan.fallback_mark_positioning = plan.adjust_mark_positioning_when_zeroing && planner.script_fallback_mark_positioning

	// currently we always apply trak.
	plan.apply_trak = plan.requested_tracking && planner.face.hasTrackTable()
}

type hb_ot_shape_plan_t struct {
	props   hb_segment_properties_t
	shaper  hb_ot_complex_shaper_t
	map_    hb_ot_map_t
	aat_map hb_aat_map_t

	data interface{} // TODO: precise if possible

	frac_mask, numr_mask, dnom_mask hb_mask_t
	rtlm_mask                       hb_mask_t
	kern_mask                       hb_mask_t
	trak_mask                       hb_mask_t

	requested_kerning  bool
	requested_tracking bool
	has_frac           bool

	has_vert                             bool
	has_gpos_mark                        bool
	zero_marks                           bool
	fallback_glyph_classes               bool
	fallback_mark_positioning            bool
	adjust_mark_positioning_when_zeroing bool

	apply_gpos bool
	apply_kern bool
	apply_kerx bool
	apply_morx bool
	apply_trak bool
}

func (sp *hb_ot_shape_plan_t) init0(face hb_face_t, key *hb_shape_plan_key_t) {
	planner := new_hb_ot_shape_planner_t(face, key.props)

	planner.hb_ot_shape_collect_features(key.userFeatures)

	planner.compile(sp, key.ot)

	sp.data = sp.shaper.data_create(sp)
}

func (sp *hb_ot_shape_plan_t) substitute(font *hb_font_t, buffer *hb_buffer_t) {
	if sp.apply_morx {
		hb_aat_layout_substitute(sp, font, buffer)
	} else {
		sp.map_.substitute(sp, font, buffer)
	}
}

//  void
//  hb_ot_shape_plan_t::position (font *hb_font_t,
// 				   buffer *hb_buffer_t) const
//  {
//    if (this.apply_gpos)
// 	 map.position (this, font, buffer);
//  #ifndef HB_NO_AAT_SHAPE
//    else if (this.apply_kerx)
// 	 hb_aat_layout_position (this, font, buffer);
//  #endif
//  #ifndef HB_NO_OT_KERN
//    else if (this.apply_kern)
// 	 hb_ot_layout_kern (this, font, buffer);
//  #endif
//    else
// 	 _hb_ot_shape_fallback_kern (this, font, buffer);

//  #ifndef HB_NO_AAT_SHAPE
//    if (this.apply_trak)
// 	 hb_aat_layout_track (this, font, buffer);
//  #endif
//  }

var (
	common_features = [...]hb_ot_map_feature_t{
		{newTag('a', 'b', 'v', 'm'), F_GLOBAL},
		{newTag('b', 'l', 'w', 'm'), F_GLOBAL},
		{newTag('c', 'c', 'm', 'p'), F_GLOBAL},
		{newTag('l', 'o', 'c', 'l'), F_GLOBAL},
		{newTag('m', 'a', 'r', 'k'), F_GLOBAL_MANUAL_JOINERS},
		{newTag('m', 'k', 'm', 'k'), F_GLOBAL_MANUAL_JOINERS},
		{newTag('r', 'l', 'i', 'g'), F_GLOBAL},
	}

	horizontal_features = [...]hb_ot_map_feature_t{
		{newTag('c', 'a', 'l', 't'), F_GLOBAL},
		{newTag('c', 'l', 'i', 'g'), F_GLOBAL},
		{newTag('c', 'u', 'r', 's'), F_GLOBAL},
		{newTag('d', 'i', 's', 't'), F_GLOBAL},
		{newTag('k', 'e', 'r', 'n'), F_GLOBAL_HAS_FALLBACK},
		{newTag('l', 'i', 'g', 'a'), F_GLOBAL},
		{newTag('r', 'c', 'l', 't'), F_GLOBAL},
	}
)

func (planner *hb_ot_shape_planner_t) hb_ot_shape_collect_features(userFeatures []hb_feature_t) {
	map_ := &planner.map_

	map_.enable_feature(newTag('r', 'v', 'r', 'n'))
	map_.add_gsub_pause(nil)

	switch planner.props.direction {
	case HB_DIRECTION_LTR:
		map_.enable_feature(newTag('l', 't', 'r', 'a'))
		map_.enable_feature(newTag('l', 't', 'r', 'm'))
	case HB_DIRECTION_RTL:
		map_.enable_feature(newTag('r', 't', 'l', 'a'))
		map_.add_feature(newTag('r', 't', 'l', 'm'))
	}

	/* Automatic fractions. */
	map_.add_feature(newTag('f', 'r', 'a', 'c'))
	map_.add_feature(newTag('n', 'u', 'm', 'r'))
	map_.add_feature(newTag('d', 'n', 'o', 'm'))

	/* Random! */
	map_.enable_feature_ext(newTag('r', 'a', 'n', 'd'), F_RANDOM, HB_OT_MAP_MAX_VALUE)

	/* Tracking.  We enable dummy feature here just to allow disabling
	* AAT 'trak' table using features.
	* https://github.com/harfbuzz/harfbuzz/issues/1303 */
	map_.enable_feature_ext(newTag('t', 'r', 'a', 'k'), F_HAS_FALLBACK, 1)

	map_.enable_feature(newTag('H', 'A', 'R', 'F'))

	planner.shaper.collect_features(planner)

	map_.enable_feature(newTag('B', 'U', 'Z', 'Z'))

	for _, feat := range common_features {
		map_.add_feature_ext(feat.tag, feat.flags, 1)
	}

	if planner.props.direction.isHorizontal() {
		for _, feat := range horizontal_features {
			map_.add_feature_ext(feat.tag, feat.flags, 1)
		}
	} else {
		/* We really want to find a 'vert' feature if there's any in the font, no
		 * matter which script/langsys it is listed (or not) under.
		 * See various bugs referenced from:
		 * https://github.com/harfbuzz/harfbuzz/issues/63 */
		map_.enable_feature_ext(newTag('v', 'e', 'r', 't'), F_GLOBAL_SEARCH, 1)
	}

	for _, f := range userFeatures {
		ftag := F_NONE
		if f.start == HB_FEATURE_GLOBAL_START && f.end == HB_FEATURE_GLOBAL_END {
			ftag = F_GLOBAL
		}
		map_.add_feature_ext(f.tag, ftag, f.value)
	}

	if planner.apply_morx {
		aat_map := &planner.aat_map
		for _, f := range userFeatures {
			aat_map.add_feature(f.tag, f.value)
		}
	}

	planner.shaper.override_features(planner)
}

//  /*
//   * shaper face data
//   */

//  struct hb_ot_face_data_t {};

//  hb_ot_face_data_t *
//  _hb_ot_shaper_face_data_create (hb_face_t *face)
//  {
//    return (hb_ot_face_data_t *) HB_SHAPER_DATA_SUCCEEDED;
//  }

//  void
//  _hb_ot_shaper_face_data_destroy (hb_ot_face_data_t *data)
//  {
//  }

//  /*
//   * shaper font data
//   */

//  struct hb_ot_font_data_t {};

//  hb_ot_font_data_t *
//  _hb_ot_shaper_font_data_create (hb_font_t *font HB_UNUSED)
//  {
//    return (hb_ot_font_data_t *) HB_SHAPER_DATA_SUCCEEDED;
//  }

//  void
//  _hb_ot_shaper_font_data_destroy (hb_ot_font_data_t *data HB_UNUSED)
//  {
//  }

/*
 * shaper
 */

type otContext struct {
	plan         *hb_ot_shape_plan_t
	font         *hb_font_t
	face         hb_face_t
	buffer       *hb_buffer_t
	userFeatures []hb_feature_t

	// transient stuff
	target_direction hb_direction_t
}

/* Main shaper */

/* Prepare */

/* Implement enough of Unicode Graphemes here that shaping
 * in reverse-direction wouldn't break graphemes.  Namely,
 * we mark all marks and ZWJ and ZWJ,Extended_Pictographic
 * sequences as continuations.  The foreach_grapheme()
 * macro uses this bit.
 *
 * https://www.unicode.org/reports/tr29/#Regex_Definitions
 */
func (buffer *hb_buffer_t) setUnicodeProps() {
	info := buffer.info
	for i := 0; i < len(info); i++ {
		info[i].setUnicodeProps(buffer)

		/* Marks are already set as continuation by the above line.
		 * Handle Emoji_Modifier and ZWJ-continuation. */
		if info[i].unicode.generalCategory() == modifierSymbol && (0x1F3FB <= info[i].codepoint && info[i].codepoint <= 0x1F3FF) {
			info[i].setContinuation()
		} else if info[i].isZwj() {
			info[i].setContinuation()
			if i+1 < len(buffer.info) && uni.isExtendedPictographic(info[i+1].codepoint) {
				i++
				info[i].setUnicodeProps(buffer)
				info[i].setContinuation()
			}
		} else if 0xE0020 <= info[i].codepoint && info[i].codepoint <= 0xE007F {
			/* Or part of the Other_Grapheme_Extend that is not marks.
			 * As of Unicode 11 that is just:
			 *
			 * 200C          ; Other_Grapheme_Extend # Cf       ZERO WIDTH NON-JOINER
			 * FF9E..FF9F    ; Other_Grapheme_Extend # Lm   [2] HALFWIDTH KATAKANA VOICED SOUND MARK..HALFWIDTH KATAKANA SEMI-VOICED SOUND MARK
			 * E0020..E007F  ; Other_Grapheme_Extend # Cf  [96] TAG SPACE..CANCEL TAG
			 *
			 * ZWNJ is special, we don't want to merge it as there's no need, and keeping
			 * it separate results in more granular clusters.  Ignore Katakana for now.
			 * Tags are used for Emoji sub-region flag sequences:
			 * https://github.com/harfbuzz/harfbuzz/issues/1556
			 */
			info[i].setContinuation()
		}
	}
}

func (buffer *hb_buffer_t) insertDottedCircle(font *hb_font_t) {
	if buffer.flags&HB_BUFFER_FLAG_DO_NOT_INSERT_DOTTED_CIRCLE != 0 {
		return
	}

	if buffer.flags&HB_BUFFER_FLAG_BOT == 0 || len(buffer.context[0]) != 0 ||
		!buffer.info[0].isUnicodeMark() {
		return
	}

	if !font.has_glyph(0x25CC) {
		return
	}

	dottedcircle := hb_glyph_info_t{codepoint: 0x25CC}
	dottedcircle.setUnicodeProps(buffer)

	buffer.clear_output()

	buffer.idx = 0
	dottedcircle.cluster = buffer.cur(0).cluster
	dottedcircle.mask = buffer.cur(0).mask
	buffer.out_info = append(buffer.out_info, dottedcircle)
	for buffer.idx < len(buffer.info) {
		buffer.next_glyph()
	}
	buffer.swap_buffers()
}

func (buffer *hb_buffer_t) formClusters() {
	if buffer.scratch_flags&HB_BUFFER_SCRATCH_FLAG_HAS_NON_ASCII == 0 {
		return
	}

	iter, count := buffer.graphemesIterator()
	if buffer.cluster_level == HB_BUFFER_CLUSTER_LEVEL_MONOTONE_GRAPHEMES {
		for start, end := iter.next(); start < count; start, end = iter.next() {
			buffer.merge_clusters(start, end)
		}
	} else {
		for start, end := iter.next(); start < count; start, end = iter.next() {
			buffer.unsafe_to_break(start, end)
		}
	}
}

func (buffer *hb_buffer_t) ensureNativeDirection() {
	direction := buffer.props.direction
	horiz_dir := hb_script_get_horizontal_direction(buffer.props.script)

	/* TODO vertical:
	* The only BTT vertical script is Ogham, but it's not clear to me whether OpenType
	* Ogham fonts are supposed to be implemented BTT or not.  Need to research that
	* first. */
	if (direction.isHorizontal() && direction != horiz_dir && horiz_dir != HB_DIRECTION_INVALID) ||
		(direction.isVertical() && direction != HB_DIRECTION_TTB) {

		iter, count := buffer.graphemesIterator()
		if buffer.cluster_level == HB_BUFFER_CLUSTER_LEVEL_MONOTONE_CHARACTERS {
			for start, end := iter.next(); start < count; start, end = iter.next() {
				buffer.merge_clusters(start, end)
				buffer.reverse_range(start, end)
			}
		} else {
			for start, end := iter.next(); start < count; start, end = iter.next() {
				// form_clusters() merged clusters already, we don't merge.
				buffer.reverse_range(start, end)
			}
		}
		buffer.reverse()

		buffer.props.direction = buffer.props.direction.reverse()
	}
}

/*
 * Substitute
 */

func vertCharFor(u rune) rune {
	switch u >> 8 {
	case 0x20:
		switch u {
		case 0x2013:
			return 0xfe32 // EN DASH
		case 0x2014:
			return 0xfe31 // EM DASH
		case 0x2025:
			return 0xfe30 // TWO DOT LEADER
		case 0x2026:
			return 0xfe19 // HORIZONTAL ELLIPSIS
		}
	case 0x30:
		switch u {
		case 0x3001:
			return 0xfe11 // IDEOGRAPHIC COMMA
		case 0x3002:
			return 0xfe12 // IDEOGRAPHIC FULL STOP
		case 0x3008:
			return 0xfe3f // LEFT ANGLE BRACKET
		case 0x3009:
			return 0xfe40 // RIGHT ANGLE BRACKET
		case 0x300a:
			return 0xfe3d // LEFT DOUBLE ANGLE BRACKET
		case 0x300b:
			return 0xfe3e // RIGHT DOUBLE ANGLE BRACKET
		case 0x300c:
			return 0xfe41 // LEFT CORNER BRACKET
		case 0x300d:
			return 0xfe42 // RIGHT CORNER BRACKET
		case 0x300e:
			return 0xfe43 // LEFT WHITE CORNER BRACKET
		case 0x300f:
			return 0xfe44 // RIGHT WHITE CORNER BRACKET
		case 0x3010:
			return 0xfe3b // LEFT BLACK LENTICULAR BRACKET
		case 0x3011:
			return 0xfe3c // RIGHT BLACK LENTICULAR BRACKET
		case 0x3014:
			return 0xfe39 // LEFT TORTOISE SHELL BRACKET
		case 0x3015:
			return 0xfe3a // RIGHT TORTOISE SHELL BRACKET
		case 0x3016:
			return 0xfe17 // LEFT WHITE LENTICULAR BRACKET
		case 0x3017:
			return 0xfe18 // RIGHT WHITE LENTICULAR BRACKET
		}
	case 0xfe:
		switch u {
		case 0xfe4f:
			return 0xfe34 // WAVY LOW LINE
		}
	case 0xff:
		switch u {
		case 0xff01:
			return 0xfe15 // FULLWIDTH EXCLAMATION MARK
		case 0xff08:
			return 0xfe35 // FULLWIDTH LEFT PARENTHESIS
		case 0xff09:
			return 0xfe36 // FULLWIDTH RIGHT PARENTHESIS
		case 0xff0c:
			return 0xfe10 // FULLWIDTH COMMA
		case 0xff1a:
			return 0xfe13 // FULLWIDTH COLON
		case 0xff1b:
			return 0xfe14 // FULLWIDTH SEMICOLON
		case 0xff1f:
			return 0xfe16 // FULLWIDTH QUESTION MARK
		case 0xff3b:
			return 0xfe47 // FULLWIDTH LEFT SQUARE BRACKET
		case 0xff3d:
			return 0xfe48 // FULLWIDTH RIGHT SQUARE BRACKET
		case 0xff3f:
			return 0xfe33 // FULLWIDTH LOW LINE
		case 0xff5b:
			return 0xfe37 // FULLWIDTH LEFT CURLY BRACKET
		case 0xff5d:
			return 0xfe38 // FULLWIDTH RIGHT CURLY BRACKET
		}
	}

	return u
}

func (c *otContext) otRotateChars() {
	info := c.buffer.info

	if c.target_direction.isBackward() {
		rtlmMask := c.plan.rtlm_mask

		for i := range info {
			codepoint := uni.mirroring(info[i].codepoint)
			if codepoint != info[i].codepoint && c.font.has_glyph(codepoint) {
				info[i].codepoint = codepoint
			} else {
				info[i].mask |= rtlmMask
			}
		}
	}

	if c.target_direction.isVertical() && !c.plan.has_vert {
		for i := range info {
			codepoint := vertCharFor(info[i].codepoint)
			if codepoint != info[i].codepoint && c.font.has_glyph(codepoint) {
				info[i].codepoint = codepoint
			}
		}
	}
}

func (c *otContext) setupMasksFraction() {
	if c.buffer.scratch_flags&HB_BUFFER_SCRATCH_FLAG_HAS_NON_ASCII == 0 || !c.plan.has_frac {
		return
	}

	buffer := c.buffer

	var pre_mask, post_mask hb_mask_t
	if buffer.props.direction.isBackward() {
		pre_mask = c.plan.frac_mask | c.plan.dnom_mask
		post_mask = c.plan.numr_mask | c.plan.frac_mask
	} else {
		pre_mask = c.plan.numr_mask | c.plan.frac_mask
		post_mask = c.plan.frac_mask | c.plan.dnom_mask
	}

	count := len(buffer.info)
	info := buffer.info
	for i := 0; i < count; i++ {
		if info[i].codepoint == 0x2044 /* FRACTION SLASH */ {
			start, end := i, i+1
			for start != 0 && info[start-1].unicode.generalCategory() == decimalNumber {
				start--
			}
			for end < count && info[end].unicode.generalCategory() == decimalNumber {
				end++
			}

			buffer.unsafe_to_break(start, end)

			for j := start; j < i; j++ {
				info[j].mask |= pre_mask
			}
			info[i].mask |= c.plan.frac_mask
			for j := i + 1; j < end; j++ {
				info[j].mask |= post_mask
			}

			i = end - 1
		}
	}
}

func (c *otContext) initializeMasks() {
	global_mask := c.plan.map_.global_mask
	c.buffer.reset_masks(global_mask)
}

func (c *otContext) setupMasks() {
	map_ := &c.plan.map_
	buffer := c.buffer

	c.setupMasksFraction()

	c.plan.shaper.setup_masks(c.plan, buffer, c.font)

	for _, feature := range c.userFeatures {
		if !(feature.start == HB_FEATURE_GLOBAL_START && feature.end == HB_FEATURE_GLOBAL_END) {
			mask, shift := map_.get_mask(feature.tag)
			buffer.set_masks(feature.value<<shift, mask, feature.start, feature.end)
		}
	}
}

func zeroWidthDefaultIgnorables(buffer *hb_buffer_t) {
	if buffer.scratch_flags&HB_BUFFER_SCRATCH_FLAG_HAS_DEFAULT_IGNORABLES == 0 ||
		buffer.flags&HB_BUFFER_FLAG_PRESERVE_DEFAULT_IGNORABLES != 0 ||
		buffer.flags&HB_BUFFER_FLAG_REMOVE_DEFAULT_IGNORABLES != 0 {
		return
	}

	pos := buffer.pos
	for i, info := range buffer.info {
		if info.isDefaultIgnorable() {
			pos[i].x_advance, pos[i].y_advance, pos[i].x_offset, pos[i].y_offset = 0, 0, 0, 0
		}
	}
}

func hideDefaultIgnorables(buffer *hb_buffer_t, font *hb_font_t) {
	if buffer.scratch_flags&HB_BUFFER_SCRATCH_FLAG_HAS_DEFAULT_IGNORABLES == 0 ||
		buffer.flags&HB_BUFFER_FLAG_PRESERVE_DEFAULT_IGNORABLES != 0 {
		return
	}

	info := buffer.info

	var (
		invisible = buffer.invisible
		ok        bool
	)
	if invisible == 0 {
		invisible, ok = font.face.GetNominalGlyph(' ')
	}
	if buffer.flags&HB_BUFFER_FLAG_REMOVE_DEFAULT_IGNORABLES == 0 && ok {
		// replace default-ignorables with a zero-advance invisible glyph.
		for i := range info {
			if info[i].isDefaultIgnorable() {
				info[i].codepoint = invisible
			}
		}
	} else {
		hb_ot_layout_delete_glyphs_inplace(buffer, (*hb_glyph_info_t).isDefaultIgnorable)
	}
}

func mapGlyphsFast(buffer *hb_buffer_t) {
	// normalization process sets up glyph_index(), we just copy it.
	info := buffer.info
	for i := range info {
		info[i].codepoint = info[i].glyph_index
	}
	buffer.content_type = HB_BUFFER_CONTENT_TYPE_GLYPHS
}

func hb_synthesize_glyph_classes(buffer *hb_buffer_t) {
	info := buffer.info
	for i := range info {
		/* Never mark default-ignorables as marks.
		 * They won't get in the way of lookups anyway,
		 * but having them as mark will cause them to be skipped
		 * over if the lookup-flag says so, but at least for the
		 * Mongolian variation selectors, looks like Uniscribe
		 * marks them as non-mark.  Some Mongolian fonts without
		 * GDEF rely on this.  Another notable character that
		 * this applies to is COMBINING GRAPHEME JOINER. */
		klass := truetype.Mark
		if info[i].unicode.generalCategory() != nonSpacingMark || info[i].isDefaultIgnorable() {
			klass = truetype.BaseGlyph
		}

		info[i].glyph_props = klass
	}
}

func (c *otContext) otSubstituteDefault() {
	buffer := c.buffer

	c.otRotateChars()

	otShapeNormalize(c.plan, buffer, c.font)

	c.setupMasks()

	// this is unfortunate to go here, but necessary...
	if c.plan.fallback_mark_positioning {
		fallbackMarkPositionRecategorizeMarks(buffer)
	}

	mapGlyphsFast(buffer)
}

func (c *otContext) substituteComplex() {
	buffer := c.buffer

	hb_ot_layout_substitute_start(c.font, buffer)

	if c.plan.fallback_glyph_classes {
		hb_synthesize_glyph_classes(c.buffer)
	}

	c.plan.substitute(c.font, buffer)
}

func (c *otContext) substitutePre() {
	c.otSubstituteDefault()
	c.substituteComplex()
}

func (c *otContext) substitutePost() {
	hideDefaultIgnorables(c.buffer, c.font)
	if c.plan.apply_morx {
		hb_aat_layout_remove_deleted_glyphs(c.buffer)
	}

	if debugMode {
		fmt.Println("start postprocess-glyphs")
	}
	c.plan.shaper.postprocess_glyphs(c.plan, c.buffer, c.font)
	if debugMode {
		fmt.Println("end postprocess-glyphs")
	}
}

/*
 * Position
 */

func zeroMarkWidthsByGdef(buffer *hb_buffer_t, adjustOffsets bool) {
	for i, inf := range buffer.info {
		if inf.isMark() {
			pos := &buffer.pos[i]
			if adjustOffsets { // adjustMarkOffsets
				pos.x_offset -= pos.x_advance
				pos.y_offset -= pos.y_advance
			}
			// zeroMarkWidth
			pos.x_advance = 0
			pos.y_advance = 0
		}
	}
}

func (c *otContext) positionDefault() {
	direction := c.buffer.props.direction
	info := c.buffer.info
	pos := c.buffer.pos
	if direction.isHorizontal() {
		for i, inf := range info {
			pos[i].x_advance = c.font.get_glyph_h_advance(inf.codepoint)
			pos[i].x_offset, pos[i].y_offset = c.font.subtract_glyph_h_origin(inf.codepoint, pos[i].x_offset, pos[i].y_offset)
		}
	} else {
		for i, inf := range info {
			pos[i].y_advance = c.font.get_glyph_v_advance(inf.codepoint)
			pos[i].x_offset, pos[i].y_offset = c.font.subtract_glyph_v_origin(inf.codepoint, pos[i].x_offset, pos[i].y_offset)
		}
	}
	if c.buffer.scratch_flags&HB_BUFFER_SCRATCH_FLAG_HAS_SPACE_FALLBACK != 0 {
		fallbackSpaces(c.font, c.buffer)
	}
}

func (c *otContext) positionComplex() {
	info := c.buffer.info
	pos := c.buffer.pos

	/* If the font has no GPOS and direction is forward, then when
	* zeroing mark widths, we shift the mark with it, such that the
	* mark is positioned hanging over the previous glyph.  When
	* direction is backward we don't shift and it will end up
	* hanging over the next glyph after the final reordering.
	*
	* Note: If fallback positioning happens, we don't care about
	* this as it will be overriden. */
	adjustOffsetsWhenZeroing := c.plan.adjust_mark_positioning_when_zeroing && !c.buffer.props.direction.isBackward()

	// we change glyph origin to what GPOS expects (horizontal), apply GPOS, change it back.

	for i, inf := range info {
		pos[i].x_offset, pos[i].y_offset = c.font.add_glyph_h_origin(inf.codepoint, pos[i].x_offset, pos[i].y_offset)
	}

	hb_ot_layout_position_start(c.font, c.buffer)

	if c.plan.zero_marks {
		if zwm, _ := c.plan.shaper.marksBehavior(); zwm == HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_EARLY {
			zeroMarkWidthsByGdef(c.buffer, adjustOffsetsWhenZeroing)
		}
	}
	c.plan.position(c.font, c.buffer)

	if c.plan.zero_marks {
		if zwm, _ := c.plan.shaper.marksBehavior(); zwm == HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_LATE {
			zeroMarkWidthsByGdef(c.buffer, adjustOffsetsWhenZeroing)
		}
	}

	// finish off.  Has to follow a certain order.
	hb_ot_layout_position_finish_advances(c.font, c.buffer)
	zeroWidthDefaultIgnorables(c.buffer)
	if c.plan.apply_morx {
		hb_aat_layout_zero_width_deleted_glyphs(c.buffer)
	}
	hb_ot_layout_position_finish_offsets(c.font, c.buffer)

	for i, inf := range info {
		pos[i].x_offset, pos[i].y_offset = c.font.subtract_glyph_h_origin(inf.codepoint, pos[i].x_offset, pos[i].y_offset)
	}

	if c.plan.fallback_mark_positioning {
		fallbackMarkPosition(c.plan, c.font, c.buffer, adjustOffsetsWhenZeroing)
	}
}

func (c *otContext) position() {
	c.buffer.clear_positions()

	c.positionDefault()

	c.positionComplex()

	if c.buffer.props.direction.isBackward() {
		c.buffer.reverse()
	}
}

/* Propagate cluster-level glyph flags to be the same on all cluster glyphs.
 * Simplifies using them. */
func propagateFlags(buffer *hb_buffer_t) {

	if buffer.scratch_flags&HB_BUFFER_SCRATCH_FLAG_HAS_UNSAFE_TO_BREAK == 0 {
		return
	}

	info := buffer.info

	iter, count := buffer.clusterIterator()
	for start, end := iter.next(); start < count; start, end = iter.next() {
		var mask uint32
		for i := start; i < end; i++ {
			if info[i].mask&HB_GLYPH_FLAG_UNSAFE_TO_BREAK != 0 {
				mask = HB_GLYPH_FLAG_UNSAFE_TO_BREAK
				break
			}
		}
		if mask != 0 {
			for i := start; i < end; i++ {
				info[i].mask |= mask
			}
		}
	}
}

// pull it all together!
func _hb_ot_shape(shape_plan *hb_shape_plan_t, font *hb_font_t, buffer *hb_buffer_t, features []hb_feature_t) bool {
	c := otContext{plan: &shape_plan.ot, font: font, face: font.face, buffer: buffer, userFeatures: features}
	c.buffer.scratch_flags = HB_BUFFER_SCRATCH_FLAG_DEFAULT
	// TODO:
	// if !hb_unsigned_mul_overflows(c.buffer.len, HB_BUFFER_MAX_LEN_FACTOR) {
	// 	c.buffer.max_len = max(c.buffer.len*HB_BUFFER_MAX_LEN_FACTOR, HB_BUFFER_MAX_LEN_MIN)
	// }
	// if !hb_unsigned_mul_overflows(c.buffer.len, HB_BUFFER_MAX_OPS_FACTOR) {
	// 	c.buffer.max_ops = max(c.buffer.len*HB_BUFFER_MAX_OPS_FACTOR, HB_BUFFER_MAX_OPS_MIN)
	// }

	// save the original direction, we use it later.
	c.target_direction = c.buffer.props.direction

	c.buffer.clear_output()

	c.initializeMasks()
	c.buffer.setUnicodeProps()
	c.buffer.insertDottedCircle(c.font)

	c.buffer.formClusters()

	c.buffer.ensureNativeDirection()

	if debugMode {
		fmt.Println("start preprocess-text")
	}
	c.plan.shaper.preprocess_text(c.plan, c.buffer, c.font)
	if debugMode {
		fmt.Println("end preprocess-text")
	}

	c.substitutePre()
	c.position()
	c.substitutePost()

	propagateFlags(c.buffer)

	c.buffer.props.direction = c.target_direction

	// c.buffer.max_len = HB_BUFFER_MAX_LEN_DEFAULT
	// c.buffer.max_ops = HB_BUFFER_MAX_OPS_DEFAULT
	return true
}

//  /**
//   * hb_ot_shape_plan_collect_lookups:
//   * @shape_plan: #hb_shape_plan_t to query
//   * @table_tag: GSUB or GPOS
//   * @lookup_indexes: (out): The #hb_set_t set of lookups returned
//   *
//   * Computes the complete set of GSUB or GPOS lookups that are applicable
//   * under a given @shape_plan.
//   *
//   * Since: 0.9.7
//   **/
//  void
//  hb_ot_shape_plan_collect_lookups (hb_shape_plan_t *shape_plan,
// 				   hb_tag_t         table_tag,
// 				   hb_set_t        *lookup_indexes /* OUT */)
//  {
//    shape_plan.ot.collect_lookups (table_tag, lookup_indexes);
//  }

//  /* TODO Move this to hb-ot-shape-normalize, make it do decompose, and make it public. */
//  static void
//  add_char (hb_font_t          *font,
// 	   hb_unicode_funcs_t *unicode,
// 	   hb_bool_t           mirror,
// 	   rune      u,
// 	   hb_set_t           *glyphs)
//  {
//    rune glyph;
//    if (font.get_nominal_glyph (u, &glyph))
// 	 glyphs.add (glyph);
//    if (mirror)
//    {
// 	 rune m = unicode.mirroring (u);
// 	 if (m != u && font.get_nominal_glyph (m, &glyph))
// 	   glyphs.add (glyph);
//    }
//  }

//  /**
//   * hb_ot_shape_glyphs_closure:
//   * @font: #hb_font_t to work upon
//   * @buffer: The input buffer to compute from
//   * @features: (array length=num_features): The features enabled on the buffer
//   * @num_features: The number of features enabled on the buffer
//   * @glyphs: (out): The #hb_set_t set of glyphs comprising the transitive closure of the query
//   *
//   * Computes the transitive closure of glyphs needed for a specified
//   * input buffer under the given font and feature list. The closure is
//   * computed as a set, not as a list.
//   *
//   * Since: 0.9.2
//   **/
//  void
//  hb_ot_shape_glyphs_closure (hb_font_t          *font,
// 				 hb_buffer_t        *buffer,
// 				 const hb_feature_t *features,
// 				 unsigned int        num_features,
// 				 hb_set_t           *glyphs)
//  {
//    const char *shapers[] = {"ot", nil};
//    hb_shape_plan_t *shape_plan = hb_shape_plan_create_cached (font.face, &buffer.props,
// 								  features, num_features, shapers);

//    bool mirror = hb_script_get_horizontal_direction (buffer.props.script) == HB_DIRECTION_RTL;

//    unsigned int count = buffer.len;
//    hb_glyph_info_t *info = buffer.info;
//    for (unsigned int i = 0; i < count; i++)
// 	 add_char (font, buffer.unicode, mirror, info[i].codepoint, glyphs);

//    hb_set_t *lookups = hb_set_create ();
//    hb_ot_shape_plan_collect_lookups (shape_plan, HB_OT_TAG_GSUB, lookups);
//    hb_ot_layout_lookups_substitute_closure (font.face, lookups, glyphs);

//    hb_set_destroy (lookups);

//    hb_shape_plan_destroy (shape_plan);
//  }

//  #endif
