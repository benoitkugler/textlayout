package opentype

import (
	"github.com/benoitkugler/textlayout/harfbuzz/common"
	cm "github.com/benoitkugler/textlayout/harfbuzz/common"
	"github.com/benoitkugler/textlayout/language"
)

// ported from harfbuzz/src/hb-ot-shape-complex-thai.cc Copyright © 2010,2012  Google, Inc.  Behdad Esfahbod

/* Thai / Lao shaper */

var _ hb_ot_complex_shaper_t = complexShaperThai{}

type complexShaperThai struct{}

/* PUA shaping */

// thai_consonant_type_t
const (
	NC = iota
	AC
	RC
	DC
	NOT_CONSONANT
	numConsonantTypes = NOT_CONSONANT
)

func getConsonantType(u rune) uint8 {
	switch u {
	case 0x0E1B, 0x0E1D, 0x0E1F /* , 0x0E2C*/ :
		return AC
	case 0x0E0D, 0x0E10:
		return RC
	case 0x0E0E, 0x0E0F:
		return DC
	}
	if 0x0E01 <= u && u <= 0x0E2E {
		return NC
	}
	return NOT_CONSONANT
}

// thai_mark_type_t
const (
	AV = iota
	BV
	T
	NOT_MARK
	NUM_MARK_TYPES = NOT_MARK
)

func getMarkType(u rune) uint8 {
	if u == 0x0E31 || (0x0E34 <= u && u <= 0x0E37) ||
		u == 0x0E47 || (0x0E4D <= u && u <= 0x0E4E) {
		return AV
	}
	if 0x0E38 <= u && u <= 0x0E3A {
		return BV
	}
	if 0x0E48 <= u && u <= 0x0E4C {
		return T
	}
	return NOT_MARK
}

// thai_action_t
const (
	NOP = iota
	SD  /* Shift combining-mark down */
	SL  /* Shift combining-mark left */
	SDL /* Shift combining-mark down-left */
	RD  /* Remove descender from base */
)

type thaiPuaMapping struct {
	u, win_pua, mac_pua rune
}

var (
	sdMappings = [...]thaiPuaMapping{
		{0x0E48, 0xF70A, 0xF88B}, /* MAI EK */
		{0x0E49, 0xF70B, 0xF88E}, /* MAI THO */
		{0x0E4A, 0xF70C, 0xF891}, /* MAI TRI */
		{0x0E4B, 0xF70D, 0xF894}, /* MAI CHATTAWA */
		{0x0E4C, 0xF70E, 0xF897}, /* THANTHAKHAT */
		{0x0E38, 0xF718, 0xF89B}, /* SARA U */
		{0x0E39, 0xF719, 0xF89C}, /* SARA UU */
		{0x0E3A, 0xF71A, 0xF89D}, /* PHINTHU */
		{0x0000, 0x0000, 0x0000},
	}
	sdlMappings = [...]thaiPuaMapping{
		{0x0E48, 0xF705, 0xF88C}, /* MAI EK */
		{0x0E49, 0xF706, 0xF88F}, /* MAI THO */
		{0x0E4A, 0xF707, 0xF892}, /* MAI TRI */
		{0x0E4B, 0xF708, 0xF895}, /* MAI CHATTAWA */
		{0x0E4C, 0xF709, 0xF898}, /* THANTHAKHAT */
		{0x0000, 0x0000, 0x0000},
	}
	slMappings = [...]thaiPuaMapping{
		{0x0E48, 0xF713, 0xF88A}, /* MAI EK */
		{0x0E49, 0xF714, 0xF88D}, /* MAI THO */
		{0x0E4A, 0xF715, 0xF890}, /* MAI TRI */
		{0x0E4B, 0xF716, 0xF893}, /* MAI CHATTAWA */
		{0x0E4C, 0xF717, 0xF896}, /* THANTHAKHAT */
		{0x0E31, 0xF710, 0xF884}, /* MAI HAN-AKAT */
		{0x0E34, 0xF701, 0xF885}, /* SARA I */
		{0x0E35, 0xF702, 0xF886}, /* SARA II */
		{0x0E36, 0xF703, 0xF887}, /* SARA UE */
		{0x0E37, 0xF704, 0xF888}, /* SARA UEE */
		{0x0E47, 0xF712, 0xF889}, /* MAITAIKHU */
		{0x0E4D, 0xF711, 0xF899}, /* NIKHAHIT */
		{0x0000, 0x0000, 0x0000},
	}
	rdMappings = [...]thaiPuaMapping{
		{0x0E0D, 0xF70F, 0xF89A}, /* YO YING */
		{0x0E10, 0xF700, 0xF89E}, /* THO THAN */
		{0x0000, 0x0000, 0x0000},
	}
)

func thaiPuaShape(u rune, action uint8, font *common.Font) rune {
	var puaMappings []thaiPuaMapping
	switch action {
	case NOP:
		return u
	case SD:
		puaMappings = sdMappings[:]
	case SDL:
		puaMappings = sdlMappings[:]
	case SL:
		puaMappings = slMappings[:]
	case RD:
		puaMappings = rdMappings[:]
	}
	for _, pua := range puaMappings {
		if pua.u == u {
			_, ok := font.Face.GetNominalGlyph(pua.win_pua)
			if ok {
				return pua.win_pua
			}
			_, ok = font.Face.GetNominalGlyph(pua.mac_pua)
			if ok {
				return pua.mac_pua
			}
			break
		}
	}
	return u
}

const (
	/* Cluster above looks like: */
	T0 = iota /*  ⣤                      */
	T1        /*     ⣼                   */
	T2        /*        ⣾                */
	T3        /*           ⣿             */
	NUM_ABOVE_STATES
)

var thaiAboveStartState = [numConsonantTypes + 1] /* For NOT_CONSONANT */ uint8{
	T0, /* NC */
	T1, /* AC */
	T0, /* RC */
	T0, /* DC */
	T3, /* NOT_CONSONANT */
}

var thai_above_state_machine = [NUM_ABOVE_STATES][NUM_MARK_TYPES]struct {
	action     uint8
	next_state uint8
}{ /*AV*/ /*BV*/ /*T*/
	/*T0*/ {{NOP, T3}, {NOP, T0}, {SD, T3}},
	/*T1*/ {{SL, T2}, {NOP, T1}, {SDL, T2}},
	/*T2*/ {{NOP, T3}, {NOP, T2}, {SL, T3}},
	/*T3*/ {{NOP, T3}, {NOP, T3}, {NOP, T3}},
}

// thai_below_state_t
const (
	B0 = iota /* No descender */
	B1        /* Removable descender */
	B2        /* Strict descender */
	NUM_BELOW_STATES
)

var thaiBelowStartState = [numConsonantTypes + 1] /* For NOT_CONSONANT */ uint8{
	B0, /* NC */
	B0, /* AC */
	B1, /* RC */
	B2, /* DC */
	B2, /* NOT_CONSONANT */
}

var thai_below_state_machine = [NUM_BELOW_STATES][NUM_MARK_TYPES]struct {
	action     uint8
	next_state uint8
}{ /*AV*/ /*BV*/ /*T*/
	/*B0*/ {{NOP, B0}, {NOP, B2}, {NOP, B0}},
	/*B1*/ {{NOP, B1}, {RD, B2}, {NOP, B1}},
	/*B2*/ {{NOP, B2}, {SD, B2}, {NOP, B2}},
}

func doThaiPuaShaping(buffer *cm.Buffer, font *cm.Font) {
	above_state := thaiAboveStartState[NOT_CONSONANT]
	below_state := thaiBelowStartState[NOT_CONSONANT]
	base := 0

	info := buffer.Info
	//    unsigned int count = buffer.len;
	for i := range info {
		mt := getMarkType(info[i].Codepoint)

		if mt == NOT_MARK {
			ct := getConsonantType(info[i].Codepoint)
			above_state = thaiAboveStartState[ct]
			below_state = thaiBelowStartState[ct]
			base = i
			continue
		}

		above_edge := &thai_above_state_machine[above_state][mt]
		below_edge := &thai_below_state_machine[below_state][mt]
		above_state = above_edge.next_state
		below_state = below_edge.next_state

		// at least one of the above/below actions is NOP.
		action := below_edge.action
		if above_edge.action != NOP {
			action = above_edge.action
		}

		buffer.UnsafeToBreak(base, i)
		if action == RD {
			info[base].Codepoint = thaiPuaShape(info[base].Codepoint, action, font)
		} else {
			info[i].Codepoint = thaiPuaShape(info[i].Codepoint, action, font)
		}
	}
}

/* We only get one script at a time, so a script-agnostic implementation
* is adequate here. */
func IS_SARA_AM(x rune) bool            { return x & ^0x0080 == 0x0E33 }
func NIKHAHIT_FROM_SARA_AM(x rune) rune { return x - 0x0E33 + 0x0E4D }
func SARA_AA_FROM_SARA_AM(x rune) rune  { return x - 1 }
func IS_TONE_MARK(x rune) bool {
	u := x & ^0x0080
	return 0x0E34 <= u && u <= 0x0E37 ||
		0x0E47 <= u && u <= 0x0E4E ||
		0x0E31 <= u && u <= 0x0E31
}

/* This function implements the shaping logic documented here:
 *
 *   https://linux.thai.net/~thep/th-otf/shaping.html
 *
 * The first shaping rule listed there is needed even if the font has Thai
 * OpenType tables.  The rest do fallback positioning based on PUA codepoints.
 * We implement that only if there exist no Thai GSUB in the font.
 */
func (complexShaperThai) preprocessText(plan *hb_ot_shape_plan_t, buffer *cm.Buffer, font *cm.Font) {

	/* The following is NOT specified in the MS OT Thai spec, however, it seems
	* to be what Uniscribe and other engines implement.  According to Eric Muller:
	*
	* When you have a SARA AM, decompose it in NIKHAHIT + SARA AA, *and* move the
	* NIKHAHIT backwards over any tone mark (0E48-0E4B).
	*
	* <0E14, 0E4B, 0E33> . <0E14, 0E4D, 0E4B, 0E32>
	*
	* This reordering is legit only when the NIKHAHIT comes from a SARA AM, not
	* when it's there to start with. The string <0E14, 0E4B, 0E4D> is probably
	* not what a user wanted, but the rendering is nevertheless nikhahit above
	* chattawa.
	*
	* Same for Lao.
	*
	* Note:
	*
	* Uniscribe also does some below-marks reordering.  Namely, it positions U+0E3A
	* after U+0E38 and U+0E39.  We do that by modifying the ccc for U+0E3A.
	* See unicode.modified_combining_class ().  Lao does NOT have a U+0E3A
	* equivalent.
	 */

	/*
	* Here are the characters of significance:
	*
	*			Thai	Lao
	* SARA AM:		U+0E33	U+0EB3
	* SARA AA:		U+0E32	U+0EB2
	* Nikhahit:		U+0E4D	U+0ECD
	*
	* Testing shows that Uniscribe reorder the following marks:
	* Thai:	<0E31,0E34..0E37,0E47..0E4E>
	* Lao:	<0EB1,0EB4..0EB7,0EC7..0ECE>
	*
	* Note how the Lao versions are the same as Thai + 0x80.
	 */

	buffer.ClearOutput()
	count := len(buffer.Info)
	for buffer.Idx = 0; buffer.Idx < count; {
		u := buffer.Cur(0).Codepoint
		if !IS_SARA_AM(u) {
			buffer.NextGlyph()
			continue
		}

		/* Is SARA AM. Decompose and reorder. */
		nikhahit := buffer.OutputGlyph(NIKHAHIT_FROM_SARA_AM(u))
		nikhahit.SetContinuation()
		buffer.ReplaceGlyph(SARA_AA_FROM_SARA_AM(u))

		/* Make Nikhahit be recognized as a ccc=0 mark when zeroing widths. */
		end := len(buffer.OutInfo)
		buffer.OutInfo[end-2].SetGeneralCategory(common.NonSpacingMark)

		/* Ok, let's see... */
		start := end - 2
		for start > 0 && IS_TONE_MARK(buffer.OutInfo[start-1].Codepoint) {
			start--
		}

		if start+2 < end {
			/* Move Nikhahit (end-2) to the beginning */
			buffer.MergeOutClusters(start, end)
			t := buffer.OutInfo[end-2]
			copy(buffer.OutInfo[start+1:], buffer.OutInfo[start:end-2])
			buffer.OutInfo[start] = t
		} else {
			/* Since we decomposed, and NIKHAHIT is combining, merge clusters with the
			* previous cluster. */
			if start != 0 && buffer.ClusterLevel == cm.HB_BUFFER_CLUSTER_LEVEL_MONOTONE_GRAPHEMES {
				buffer.MergeOutClusters(start-1, end)
			}
		}
	}
	buffer.SwapBuffers()

	/* If font has Thai GSUB, we are done. */
	if plan.props.script == language.Thai && !plan.map_.found_script[0] {
		doThaiPuaShaping(buffer, font)
	}
}

func (complexShaperThai) marksBehavior() (hb_ot_shape_zero_width_marks_type_t, bool) {
	return HB_OT_SHAPE_ZERO_WIDTH_MARKS_BY_GDEF_LATE, false
}

func (complexShaperThai) normalizationPreference() hb_ot_shape_normalization_mode_t {
	return HB_OT_SHAPE_NORMALIZATION_MODE_DEFAULT
}

func (complexShaperThai) compose(_ *hb_ot_shape_normalize_context_t, a, b rune) (rune, bool) {
	return cm.Uni.Compose(a, b)
}
func (complexShaperThai) decompose(c *hb_ot_shape_normalize_context_t, ab rune) (a, b rune, ok bool) {
	return cm.Uni.Decompose(ab)
}
func (complexShaperThai) gposTag() hb_tag_t { return 0 }
func (complexShaperThai) collectFeatures(plan *hb_ot_shape_planner_t)
func (complexShaperThai) overrideFeatures(plan *hb_ot_shape_planner_t)
func (complexShaperThai) dataCreate(plan *hb_ot_shape_plan_t)
func (complexShaperThai) setupMasks(plan *hb_ot_shape_plan_t, buffer *cm.Buffer, font *cm.Font)
func (complexShaperThai) reorderMarks(plan *hb_ot_shape_plan_t, buffer *cm.Buffer, start, end int)
func (complexShaperThai) postprocessGlyphs(plan *hb_ot_shape_plan_t, buffer *cm.Buffer, font *cm.Font)
