package harfbuzz

import (
	"github.com/benoitkugler/textlayout/language"
)

type hb_ot_shape_zero_width_marks_type_t uint8

const (
	HB_OT_SHAPE_ZERO_WIDTH_MARKS_NONE hb_ot_shape_zero_width_marks_type_t = iota
	HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_EARLY
	HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_LATE
)

// implements the specialisation for a script
type hb_ot_complex_shaper_t interface {
	marksBehavior() (zero_width_marks hb_ot_shape_zero_width_marks_type_t, fallback_position bool)
	normalizationPreference() hb_ot_shape_normalization_mode_t
	// If not 0, then must match found GPOS script tag for
	// GPOS to be applied. Otherwise, fallback positioning will be used.
	gposTag() hb_tag_t

	// collectFeatures is alled during shape_plan().
	// Shapers should use plan.map to add their features and callbacks.
	collectFeatures(plan *hb_ot_shape_planner_t)

	// overrideFeatures is called during shape_plan().
	// Shapers should use plan.map to override features and add callbacks after
	// common features are added.
	overrideFeatures(plan *hb_ot_shape_planner_t)

	// dataCreate is called at the end of shape_plan().
	dataCreate(plan *hb_ot_shape_plan_t)

	// called during shape(), shapers can use to modify text before shaping starts.
	preprocessText(plan *hb_ot_shape_plan_t, buffer *Buffer, font *Font)

	// called during shape()'s normalization: may use decompose_unicode as fallback
	decompose(c *hb_ot_shape_normalize_context_t, ab rune) (a, b rune, ok bool)

	// called during shape()'s normalization: may use compose_unicode as fallback
	compose(c *hb_ot_shape_normalize_context_t, a, b rune) (ab rune, ok bool)

	// called during shape(), shapers should use map to get feature masks and set on buffer.
	// Shapers may NOT modify characters.
	setupMasks(plan *hb_ot_shape_plan_t, buffer *Buffer, font *Font)

	// called during shape(), shapers can use to modify ordering of combining marks.
	reorderMarks(plan *hb_ot_shape_plan_t, buffer *Buffer, start, end int)

	// called during shape(), shapers can use to modify glyphs after shaping ends.
	postprocessGlyphs(plan *hb_ot_shape_plan_t, buffer *Buffer, font *Font)
}

/*
 * For lack of a better place, put Zawgyi script hack here.
 * https://github.com/harfbuzz/harfbuzz/issues/1162
 */
var scriptMyanmar_Zawgyi = language.Script(newTag('Q', 'a', 'a', 'g'))

func hb_ot_shape_complex_categorize(planner *hb_ot_shape_planner_t) hb_ot_complex_shaper_t {
	switch planner.props.script {
	case language.Arabic, language.Syriac:
		/* For Arabic script, use the Arabic shaper even if no OT script tag was found.
		 * This is because we do fallback shaping for Arabic script (and not others).
		 * But note that Arabic shaping is applicable only to horizontal layout; for
		 * vertical text, just use the generic shaper instead. */
		if (planner.map_.chosen_script[0] != HB_OT_TAG_DEFAULT_SCRIPT ||
			planner.props.script == language.Arabic) &&
			planner.props.direction.IsHorizontal() {
			return &complexShaperArabic{}
		}
		return complexShaperDefault{}
	case language.Thai, language.Lao:
		return complexShaperThai{}
	case language.Hangul:
		return &complexShaperHangul{}
	case language.Hebrew:
		return complexShaperHebrew{}
	case language.Bengali, language.Devanagari, language.Gujarati, language.Gurmukhi, language.Kannada,
		language.Malayalam, language.Oriya, language.Tamil, language.Telugu, language.Sinhala:
		/* If the designer designed the font for the 'DFLT' script,
		 * (or we ended up arbitrarily pick 'latn'), use the default shaper.
		 * Otherwise, use the specific shaper.
		 *
		 * If it's indy3 tag, send to USE. */
		if planner.map_.chosen_script[0] == newTag('D', 'F', 'L', 'T') ||
			planner.map_.chosen_script[0] == newTag('l', 'a', 't', 'n') {
			return complexShaperDefault{}
		} else if (planner.map_.chosen_script[0] & 0x000000FF) == '3' {
			return &complexShaperUSE{}
		}
		return &complexShaperIndic{}
	case language.Khmer:
		return &complexShaperKhmer{}
	case language.Myanmar:
		/* If the designer designed the font for the 'DFLT' script,
		 * (or we ended up arbitrarily pick 'latn'), use the default shaper.
		 * Otherwise, use the specific shaper.
		 *
		 * If designer designed for 'mymr' tag, also send to default
		 * shaper.  That's tag used from before Myanmar shaping spec
		 * was developed.  The shaping spec uses 'mym2' tag. */
		if planner.map_.chosen_script[0] == newTag('D', 'F', 'L', 'T') ||
			planner.map_.chosen_script[0] == newTag('l', 'a', 't', 'n') ||
			planner.map_.chosen_script[0] == newTag('m', 'y', 'm', 'r') {
			return complexShaperDefault{}
		}
		return complexShaperMyanmar{}

	case scriptMyanmar_Zawgyi:
		/* Ugly Zawgyi encoding.
		 * Disable all auto processing.
		 * https://github.com/harfbuzz/harfbuzz/issues/1162 */
		return complexShaperDefault{dumb: true, disableNorm: true}
	case language.Tibetan, language.Mongolian, language.Buhid, language.Hanunoo, language.Tagalog,
		language.Tagbanwa, language.Limbu, language.Tai_Le, language.Buginese, language.Kharoshthi,
		language.Syloti_Nagri, language.Tifinagh, language.Balinese, language.Nko, language.Phags_Pa,
		language.Cham, language.Kayah_Li, language.Lepcha, language.Rejang, language.Saurashtra,
		language.Sundanese, language.Egyptian_Hieroglyphs, language.Javanese, language.Kaithi,
		language.Meetei_Mayek, language.Tai_Tham, language.Tai_Viet, language.Batak,
		language.Brahmi, language.Mandaic, language.Chakma, language.Miao, language.Sharada,
		language.Takri, language.Duployan, language.Grantha, language.Khojki, language.Khudawadi,
		language.Mahajani, language.Manichaean, language.Modi, language.Pahawh_Hmong,
		language.Psalter_Pahlavi, language.Siddham, language.Tirhuta, language.Ahom, language.Multani,
		language.Adlam, language.Bhaiksuki, language.Marchen, language.Newa, language.Masaram_Gondi,
		language.Soyombo, language.Zanabazar_Square, language.Dogra, language.Gunjala_Gondi,
		language.Hanifi_Rohingya, language.Makasar, language.Medefaidrin, language.Old_Sogdian,
		language.Sogdian, language.Elymaic, language.Nandinagari, language.Nyiakeng_Puachue_Hmong,
		language.Wancho, language.Chorasmian, language.Dives_Akuru:

		/* If the designer designed the font for the 'DFLT' script,
		 * (or we ended up arbitrarily pick 'latn'), use the default shaper.
		 * Otherwise, use the specific shaper.
		 * Note that for some simple scripts, there may not be *any*
		 * GSUB/GPOS needed, so there may be no scripts found! */
		if planner.map_.chosen_script[0] == newTag('D', 'F', 'L', 'T') ||
			planner.map_.chosen_script[0] == newTag('l', 'a', 't', 'n') {
			return complexShaperDefault{}
		}
		return &complexShaperUSE{}
	default:
		return complexShaperDefault{}
	}
}

// zero byte struct providing no-ops, used to reduced boilerplate
type complexShaperNil struct{}

func (complexShaperNil) gposTag() hb_tag_t { return 0 }

func (complexShaperNil) collectFeatures(plan *hb_ot_shape_planner_t)  {}
func (complexShaperNil) overrideFeatures(plan *hb_ot_shape_planner_t) {}
func (complexShaperNil) dataCreate(plan *hb_ot_shape_plan_t)          {}
func (complexShaperNil) decompose(_ *hb_ot_shape_normalize_context_t, ab rune) (a, b rune, ok bool) {
	return Uni.Decompose(ab)
}

func (complexShaperNil) compose(_ *hb_ot_shape_normalize_context_t, a, b rune) (ab rune, ok bool) {
	return Uni.Compose(a, b)
}
func (complexShaperNil) preprocessText(*hb_ot_shape_plan_t, *Buffer, *Font) {}
func (complexShaperNil) postprocessGlyphs(*hb_ot_shape_plan_t, *Buffer, *Font) {
}
func (complexShaperNil) setupMasks(*hb_ot_shape_plan_t, *Buffer, *Font)      {}
func (complexShaperNil) reorderMarks(*hb_ot_shape_plan_t, *Buffer, int, int) {}

type complexShaperDefault struct {
	complexShaperNil

	/* if true, no mark advance zeroing / fallback positioning.
	 * Dumbest shaper ever, basically. */
	dumb        bool
	disableNorm bool
}

func (cs complexShaperDefault) marksBehavior() (hb_ot_shape_zero_width_marks_type_t, bool) {
	if cs.dumb {
		return HB_OT_SHAPE_ZERO_WIDTH_MARKS_NONE, false
	}
	return HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_LATE, true
}

func (cs complexShaperDefault) normalizationPreference() hb_ot_shape_normalization_mode_t {
	if cs.disableNorm {
		return HB_OT_SHAPE_NORMALIZATION_MODE_NONE
	}
	return HB_OT_SHAPE_NORMALIZATION_MODE_DEFAULT
}

func hb_syllabic_insert_dotted_circles(font *Font, buffer *Buffer, brokenSyllableType,
	dottedcircleCategory uint8, rephaCategory int) {
	if (buffer.Flags & DoNotinsertDottedCircle) != 0 {
		return
	}

	hasBrokenSyllables := false
	info := buffer.Info
	for _, inf := range info {
		if (inf.syllable & 0x0F) == brokenSyllableType {
			hasBrokenSyllables = true
			break
		}
	}
	if !hasBrokenSyllables {
		return
	}

	dottedcircleGlyph, ok := font.Face.GetNominalGlyph(0x25CC)
	if !ok {
		return
	}

	dottedcircle := GlyphInfo{
		codepoint:       dottedcircleGlyph,
		complexCategory: dottedcircleCategory,
	}

	buffer.clearOutput()

	buffer.idx = 0
	var last_syllable uint8
	for buffer.idx < len(buffer.Info) {
		syllable := buffer.cur(0).syllable
		if last_syllable != syllable && (syllable&0x0F) == brokenSyllableType {
			last_syllable = syllable

			ginfo := dottedcircle
			ginfo.Cluster = buffer.cur(0).Cluster
			ginfo.mask = buffer.cur(0).mask
			ginfo.syllable = buffer.cur(0).syllable

			/* Insert dottedcircle after possible Repha. */
			if rephaCategory != -1 {
				for buffer.idx < len(buffer.Info) &&
					last_syllable == buffer.cur(0).syllable &&
					buffer.cur(0).complexCategory == uint8(rephaCategory) {
					buffer.nextGlyph()
				}
			}
			buffer.OutputInfo(ginfo)
		} else {
			buffer.nextGlyph()
		}
	}
	buffer.swapBuffers()
}
