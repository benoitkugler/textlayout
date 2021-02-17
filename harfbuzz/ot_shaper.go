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
	face                             Face
	props                            SegmentProperties
	map_                             hb_ot_map_builder_t
	aat_map                          hb_aat_map_builder_t
	apply_morx                       bool
	script_zero_marks                bool
	script_fallback_mark_positioning bool
	shaper                           hb_ot_complex_shaper_t
}

func new_hb_ot_shape_planner_t(face Face, props SegmentProperties) *hb_ot_shape_planner_t {
	var out hb_ot_shape_planner_t
	out.map_ = new_hb_ot_map_builder_t(face, props)
	out.aat_map = hb_aat_map_builder_t{face: face}

	/* https://github.com/harfbuzz/harfbuzz/issues/2124 */
	out.apply_morx = aatLayoutHasSubstitution(face) && (props.Direction.IsHorizontal() || !otLayoutHasSubstitution(face))

	out.shaper = out.shapeComplexCategorize()

	zwm, fb := out.shaper.marksBehavior()
	out.script_zero_marks = zwm != HB_OT_SHAPE_ZERO_WIDTH_MARKS_NONE
	out.script_fallback_mark_positioning = fb

	/* https://github.com/harfbuzz/harfbuzz/issues/1528 */
	if _, isDefault := out.shaper.(complexShaperDefault); out.apply_morx && !isDefault {
		out.shaper = complexShaperDefault{dumb: true}
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
	if planner.props.Direction.IsHorizontal() {
		kern_tag = newTag('k', 'e', 'r', 'n')
	}

	plan.kern_mask, _ = plan.map_.get_mask(kern_tag)
	plan.requested_kerning = plan.kern_mask != 0
	plan.trak_mask, _ = plan.map_.get_mask(newTag('t', 'r', 'a', 'k'))
	plan.requested_tracking = plan.trak_mask != 0

	has_gpos_kern := plan.map_.get_feature_index(1, kern_tag) != HB_OT_LAYOUT_NO_FEATURE_INDEX
	disable_gpos := plan.shaper.gposTag() != 0 && plan.shaper.gposTag() != plan.map_.chosen_script[1]

	// Decide who provides glyph classes. GDEF or Unicode.
	if planner.face.getGDEF().Class == nil {
		plan.fallback_glyph_classes = true
	}

	// Decide who does substitutions. GSUB, morx, or fallback.
	plan.apply_morx = planner.apply_morx

	//  Decide who does positioning. GPOS, kerx, kern, or fallback.
	hasKerx := aatLayoutHasPositioning(planner.face)
	if hasKerx {
		plan.apply_kerx = true
	} else if _, gpos := planner.face.get_gsubgpos_table(); !planner.apply_morx && !disable_gpos && gpos != nil {
		plan.apply_gpos = true
	}

	if !plan.apply_kerx && (!has_gpos_kern || !plan.apply_gpos) {
		// apparently Apple applies kerx if GPOS kern was not applied.
		if hasKerx {
			plan.apply_kerx = true
		} else if hasKerning(planner.face) {
			plan.apply_kern = true
		}
	}

	plan.zero_marks = planner.script_zero_marks && !plan.apply_kerx &&
		(!plan.apply_kern || !hasMachineKerning(planner.face))
	plan.has_gpos_mark = plan.map_.get_1_mask(newTag('m', 'a', 'r', 'k')) != 0

	plan.adjust_mark_positioning_when_zeroing = !plan.apply_gpos && !plan.apply_kerx &&
		(!plan.apply_kern || !hasCrossKerning(planner.face))

	plan.fallback_mark_positioning = plan.adjust_mark_positioning_when_zeroing && planner.script_fallback_mark_positioning

	// currently we always apply trak.
	plan.apply_trak = plan.requested_tracking && aatLayoutHasTracking(planner.face)
}

type hb_ot_shape_plan_t struct {
	props   SegmentProperties
	shaper  hb_ot_complex_shaper_t
	map_    hb_ot_map_t
	aat_map hb_aat_map_t

	frac_mask, numr_mask, dnom_mask Mask
	rtlm_mask                       Mask
	kern_mask                       Mask
	trak_mask                       Mask

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

func (sp *hb_ot_shape_plan_t) init0(face Face, key *hb_shape_plan_key_t) {
	planner := new_hb_ot_shape_planner_t(face, key.props)

	planner.hb_ot_shape_collect_features(key.userFeatures)

	planner.compile(sp, key.ot)

	sp.shaper.dataCreate(sp)
}

func (sp *hb_ot_shape_plan_t) substitute(font *Font, buffer *Buffer) {
	if sp.apply_morx {
		sp.aatLayoutSubstitute(font, buffer)
	} else {
		sp.map_.substitute(sp, font, buffer)
	}
}

func (sp *hb_ot_shape_plan_t) position(font *Font, buffer *Buffer) {
	if sp.apply_gpos {
		sp.map_.position(sp, font, buffer)
	} else if sp.apply_kerx {
		sp.aatLayoutPosition(font, buffer)
	} else if sp.apply_kern {
		sp.otLayoutKern(font, buffer)
	} else {
		// deprecated
	}

	if sp.apply_trak {
		sp.aatLayoutTrack(font, buffer)
	}
}

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

func (planner *hb_ot_shape_planner_t) hb_ot_shape_collect_features(userFeatures []Feature) {
	map_ := &planner.map_

	map_.enable_feature(newTag('r', 'v', 'r', 'n'))
	map_.add_gsub_pause(nil)

	switch planner.props.Direction {
	case LeftToRight:
		map_.enable_feature(newTag('l', 't', 'r', 'a'))
		map_.enable_feature(newTag('l', 't', 'r', 'm'))
	case RightToLeft:
		map_.enable_feature(newTag('r', 't', 'l', 'a'))
		map_.add_feature(newTag('r', 't', 'l', 'm'))
	}

	/* Automatic fractions. */
	map_.add_feature(newTag('f', 'r', 'a', 'c'))
	map_.add_feature(newTag('n', 'u', 'm', 'r'))
	map_.add_feature(newTag('d', 'n', 'o', 'm'))

	/* Random! */
	map_.enable_feature_ext(newTag('r', 'a', 'n', 'd'), F_RANDOM, otMapMaxValue)

	/* Tracking.  We enable dummy feature here just to allow disabling
	* AAT 'trak' table using features.
	* https://github.com/harfbuzz/harfbuzz/issues/1303 */
	map_.enable_feature_ext(newTag('t', 'r', 'a', 'k'), F_HAS_FALLBACK, 1)

	map_.enable_feature(newTag('H', 'A', 'R', 'F'))

	planner.shaper.collectFeatures(planner)

	map_.enable_feature(newTag('B', 'U', 'Z', 'Z'))

	for _, feat := range common_features {
		map_.add_feature_ext(feat.tag, feat.flags, 1)
	}

	if planner.props.Direction.IsHorizontal() {
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
		if f.Start == FeatureGlobalStart && f.End == FeatureGlobalEnd {
			ftag = F_GLOBAL
		}
		map_.add_feature_ext(f.Tag, ftag, f.Value)
	}

	if planner.apply_morx {
		aat_map := &planner.aat_map
		for _, f := range userFeatures {
			aat_map.add_feature(f.Tag, f.Value)
		}
	}

	planner.shaper.overrideFeatures(planner)
}

/*
 * shaper
 */

type otContext struct {
	plan         *hb_ot_shape_plan_t
	font         *Font
	face         Face
	buffer       *Buffer
	userFeatures []Feature

	// transient stuff
	target_direction Direction
}

/* Main shaper */

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
	info := c.buffer.Info

	if c.target_direction.IsBackward() {
		rtlmMask := c.plan.rtlm_mask

		for i := range info {
			codepoint := Uni.Mirroring(info[i].codepoint)
			if codepoint != info[i].codepoint && c.font.HasGlyph(codepoint) {
				info[i].codepoint = codepoint
			} else {
				info[i].mask |= rtlmMask
			}
		}
	}

	if c.target_direction.IsVertical() && !c.plan.has_vert {
		for i := range info {
			codepoint := vertCharFor(info[i].codepoint)
			if codepoint != info[i].codepoint && c.font.HasGlyph(codepoint) {
				info[i].codepoint = codepoint
			}
		}
	}
}

func (c *otContext) setupMasksFraction() {
	if c.buffer.scratchFlags&HB_BUFFER_SCRATCH_FLAG_HAS_NON_ASCII == 0 || !c.plan.has_frac {
		return
	}

	buffer := c.buffer

	var pre_mask, post_mask Mask
	if buffer.Props.Direction.isForward() {
		pre_mask = c.plan.numr_mask | c.plan.frac_mask
		post_mask = c.plan.frac_mask | c.plan.dnom_mask
	} else {
		pre_mask = c.plan.frac_mask | c.plan.dnom_mask
		post_mask = c.plan.numr_mask | c.plan.frac_mask
	}

	count := len(buffer.Info)
	info := buffer.Info
	for i := 0; i < count; i++ {
		if info[i].codepoint == 0x2044 /* FRACTION SLASH */ {
			start, end := i, i+1
			for start != 0 && info[start-1].unicode.generalCategory() == DecimalNumber {
				start--
			}
			for end < count && info[end].unicode.generalCategory() == DecimalNumber {
				end++
			}

			buffer.unsafeToBreak(start, end)

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
	c.buffer.resetMasks(c.plan.map_.global_mask)
}

func (c *otContext) setupMasks() {
	map_ := &c.plan.map_
	buffer := c.buffer

	c.setupMasksFraction()

	c.plan.shaper.setupMasks(c.plan, buffer, c.font)

	for _, feature := range c.userFeatures {
		if !(feature.Start == FeatureGlobalStart && feature.End == FeatureGlobalEnd) {
			mask, shift := map_.get_mask(feature.Tag)
			buffer.setMasks(feature.Value<<shift, mask, feature.Start, feature.End)
		}
	}
}

func zeroWidthDefaultIgnorables(buffer *Buffer) {
	if buffer.scratchFlags&HB_BUFFER_SCRATCH_FLAG_HAS_DEFAULT_IGNORABLES == 0 ||
		buffer.Flags&PreserveDefaultIgnorables != 0 ||
		buffer.Flags&RemoveDefaultIgnorables != 0 {
		return
	}

	pos := buffer.Pos
	for i, info := range buffer.Info {
		if info.isDefaultIgnorable() {
			pos[i].XAdvance, pos[i].YAdvance, pos[i].XOffset, pos[i].YOffset = 0, 0, 0, 0
		}
	}
}

func hideDefaultIgnorables(buffer *Buffer, font *Font) {
	if buffer.scratchFlags&HB_BUFFER_SCRATCH_FLAG_HAS_DEFAULT_IGNORABLES == 0 ||
		buffer.Flags&PreserveDefaultIgnorables != 0 {
		return
	}

	info := buffer.Info

	var (
		invisible = buffer.Invisible
		ok        bool
	)
	if invisible == 0 {
		invisible, ok = font.Face.GetNominalGlyph(' ')
	}
	if buffer.Flags&RemoveDefaultIgnorables == 0 && ok {
		// replace default-ignorables with a zero-advance invisible glyph.
		for i := range info {
			if info[i].isDefaultIgnorable() {
				info[i].Glyph = invisible
			}
		}
	} else {
		otLayoutDeleteGlyphsInplace(buffer, (*GlyphInfo).isDefaultIgnorable)
	}
}

// use unicodeProp to assign a class
func synthesizeGlyphClasses(buffer *Buffer) {
	info := buffer.Info
	for i := range info {
		/* Never mark default-ignorables as marks.
		 * They won't get in the way of lookups anyway,
		 * but having them as mark will cause them to be skipped
		 * over if the lookup-flag says so, but at least for the
		 * Mongolian variation selectors, looks like Uniscribe
		 * marks them as non-mark.  Some Mongolian fonts without
		 * GDEF rely on this.  Another notable character that
		 * this applies to is COMBINING GRAPHEME JOINER. */
		class := truetype.Mark
		if info[i].unicode.generalCategory() != NonSpacingMark || info[i].isDefaultIgnorable() {
			class = truetype.BaseGlyph
		}

		info[i].glyphProps = class
	}
}

func (c *otContext) substituteBeforePosition() {
	buffer := c.buffer
	// normalize and sets Glyph

	c.otRotateChars()

	otShapeNormalize(c.plan, buffer, c.font)

	c.setupMasks()

	// this is unfortunate to go here, but necessary...
	if c.plan.fallback_mark_positioning {
		fallbackMarkPositionRecategorizeMarks(buffer)
	}

	// Glyph fields are now set up ...
	// ... apply complex substitution from font

	layoutSubstituteStart(c.font, buffer)

	if c.plan.fallback_glyph_classes {
		synthesizeGlyphClasses(c.buffer)
	}

	c.plan.substitute(c.font, buffer)
}

func (c *otContext) substituteAfterPosition() {
	hideDefaultIgnorables(c.buffer, c.font)
	if c.plan.apply_morx {
		aatLayoutRemoveDeletedGlyphsInplace(c.buffer)
	}

	if debugMode {
		fmt.Println("start postprocess-glyphs")
	}
	c.plan.shaper.postprocessGlyphs(c.plan, c.buffer, c.font)
	if debugMode {
		fmt.Println("end postprocess-glyphs")
	}
}

/*
 * Position
 */

func zeroMarkWidthsByGdef(buffer *Buffer, adjustOffsets bool) {
	for i, inf := range buffer.Info {
		if inf.isMark() {
			pos := &buffer.Pos[i]
			if adjustOffsets { // adjustMarkOffsets
				pos.XOffset -= pos.XAdvance
				pos.YOffset -= pos.YAdvance
			}
			// zeroMarkWidth
			pos.XAdvance = 0
			pos.YAdvance = 0
		}
	}
}

// override Pos array with default values
func (c *otContext) positionDefault() {
	direction := c.buffer.Props.Direction
	info := c.buffer.Info
	pos := c.buffer.Pos
	if direction.IsHorizontal() {
		for i, inf := range info {
			pos[i].XAdvance, pos[i].YAdvance = c.font.GetGlyphHAdvance(inf.Glyph), 0
			pos[i].XOffset, pos[i].YOffset = c.font.subtract_glyph_h_origin(inf.Glyph, 0, 0)
		}
	} else {
		for i, inf := range info {
			pos[i].XAdvance, pos[i].YAdvance = 0, c.font.GetGlyphVAdvance(inf.Glyph)
			pos[i].XOffset, pos[i].YOffset = c.font.subtract_glyph_v_origin(inf.Glyph, 0, 0)
		}
	}
	if c.buffer.scratchFlags&HB_BUFFER_SCRATCH_FLAG_HAS_SPACE_FALLBACK != 0 {
		fallbackSpaces(c.font, c.buffer)
	}
}

func (c *otContext) positionComplex() {
	info := c.buffer.Info
	pos := c.buffer.Pos

	/* If the font has no GPOS and direction is forward, then when
	* zeroing mark widths, we shift the mark with it, such that the
	* mark is positioned hanging over the previous glyph.  When
	* direction is backward we don't shift and it will end up
	* hanging over the next glyph after the final reordering.
	*
	* Note: If fallback positioning happens, we don't care about
	* this as it will be overriden. */
	adjustOffsetsWhenZeroing := c.plan.adjust_mark_positioning_when_zeroing && c.buffer.Props.Direction.isForward()

	// we change glyph origin to what GPOS expects (horizontal), apply GPOS, change it back.

	for i, inf := range info {
		pos[i].XOffset, pos[i].YOffset = c.font.add_glyph_h_origin(inf.Glyph, pos[i].XOffset, pos[i].YOffset)
	}

	otLayoutPositionStart(c.font, c.buffer)

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

	// finish off. Has to follow a certain order.
	zeroWidthDefaultIgnorables(c.buffer)
	if c.plan.apply_morx {
		aatLayoutZeroWidthDeletedGlyphs(c.buffer)
	}
	otLayoutPositionFinishOffsets(c.font, c.buffer)

	for i, inf := range info {
		pos[i].XOffset, pos[i].YOffset = c.font.subtract_glyph_h_origin(inf.Glyph, pos[i].XOffset, pos[i].YOffset)
	}

	if c.plan.fallback_mark_positioning {
		fallbackMarkPosition(c.plan, c.font, c.buffer, adjustOffsetsWhenZeroing)
	}
}

func (c *otContext) position() {
	c.buffer.clearPositions()

	c.positionDefault()

	c.positionComplex()

	if c.buffer.Props.Direction.IsBackward() {
		c.buffer.Reverse()
	}
}

/* Propagate cluster-level glyph flags to be the same on all cluster glyphs.
 * Simplifies using them. */
func propagateFlags(buffer *Buffer) {
	if buffer.scratchFlags&HB_BUFFER_SCRATCH_FLAG_HAS_UNSAFE_TO_BREAK == 0 {
		return
	}

	info := buffer.Info

	iter, count := buffer.ClusterIterator()
	for start, end := iter.Next(); start < count; start, end = iter.Next() {
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

// shaperOpentype is the main shaper of this library.
// It handles complex language and Opentype layout features found in fonts.
type shaperOpentype struct{}

var _ shaper = shaperOpentype{}

// pull it all together!
func (shaperOpentype) shape(shape_plan *ShapePlan, font *Font, buffer *Buffer, features []Feature) {
	c := otContext{plan: &shape_plan.ot, font: font, face: font.Face, buffer: buffer, userFeatures: features}
	c.buffer.scratchFlags = HB_BUFFER_SCRATCH_FLAG_DEFAULT

	const maxOpsFactor = 1024
	c.buffer.max_ops = len(c.buffer.Info) * maxOpsFactor

	// save the original direction, we use it later.
	c.target_direction = c.buffer.Props.Direction

	c.buffer.clearOutput()

	c.initializeMasks()
	c.buffer.setUnicodeProps()
	c.buffer.insertDottedCircle(c.font)

	c.buffer.formClusters()

	c.buffer.ensureNativeDirection()

	if debugMode {
		fmt.Println("start preprocess-text")
	}
	c.plan.shaper.preprocessText(c.plan, c.buffer, c.font)
	if debugMode {
		fmt.Println("end preprocess-text")
	}

	c.substituteBeforePosition()
	c.position()
	c.substituteAfterPosition()

	propagateFlags(c.buffer)

	c.buffer.Props.Direction = c.target_direction

	c.buffer.max_ops = maxOpsDefault
}

//  /**
//   * hb_ot_shape_plan_collect_lookups:
//   * @shape_plan: #ShapePlan to query
//   * @table_tag: GSUB or GPOS
//   * @lookup_indexes: (out): The #hb_set_t set of lookups returned
//   *
//   * Computes the complete set of GSUB or GPOS lookups that are applicable
//   * under a given @shape_plan.
//   *
//   * Since: 0.9.7
//   **/
//  void
//  hb_ot_shape_plan_collect_lookups (ShapePlan *shape_plan,
// 				   hb_tag_t         table_tag,
// 				   hb_set_t        *lookup_indexes /* OUT */)
//  {
//    shape_plan.ot.collect_lookups (table_tag, lookup_indexes);
//  }

//  /* TODO Move this to hb-ot-shape-normalize, make it do decompose, and make it public. */
//  static void
//  add_char (Font          *font,
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
// 	 rune m = unicode.Mirroring (u);
// 	 if (m != u && font.get_nominal_glyph (m, &glyph))
// 	   glyphs.add (glyph);
//    }
//  }

//  /**
//   * hb_ot_shape_glyphs_closure:
//   * @font: #Font to work upon
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
//  hb_ot_shape_glyphs_closure (Font          *font,
// 				 Buffer        *buffer,
// 				 const  Feature *features,
// 				 unsigned int        num_features,
// 				 hb_set_t           *glyphs)
//  {
//    const char *shapers[] = {"ot", nil};
//    ShapePlan *shape_plan = hb_shape_plan_create_cached (font.Face, &buffer.Props,
// 								  features, num_features, shapers);

//    bool mirror = GetHorizontalDirection (buffer.Props.script) == RightToLeft;

//    unsigned int count = buffer.len;
//    GlyphInfo *info = buffer.Info;
//    for (unsigned int i = 0; i < count; i++)
// 	 add_char (font, buffer.unicode, mirror, info[i].Codepoint, glyphs);

//    hb_set_t *lookups = hb_set_create ();
//    hb_ot_shape_plan_collect_lookups (shape_plan, HB_OT_TAG_GSUB, lookups);
//    hb_ot_layout_lookups_substitute_closure (font.Face, lookups, glyphs);

//    hb_set_destroy (lookups);

//    hb_shape_plan_destroy (shape_plan);
//  }

//  #endif
