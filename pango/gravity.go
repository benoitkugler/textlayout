package pango

import (
	"strings"

	"github.com/benoitkugler/textlayout/language"
)

// Gravity represents the orientation of glyphs in a segment
// of text.  This is useful when rendering vertical text layouts. In
// those situations, the layout is rotated using a non-identity Matrix,
// and then glyph orientation is controlled using Gravity.
//
// See also: GravityHint
type Gravity uint8

const (
	GRAVITY_SOUTH Gravity = iota // Glyphs stand upright (default)
	GRAVITY_EAST                 // Glyphs are rotated 90 degrees clockwise
	GRAVITY_NORTH                // Glyphs are upside-down
	GRAVITY_WEST                 // Glyphs are rotated 90 degrees counter-clockwise
	GRAVITY_AUTO                 // Gravity is resolved from the context matrix
)

// GravityMap exposes the string representation for the Graviy values.
var GravityMap = enumMap{
	{value: int(GRAVITY_SOUTH), str: "South"},
	{value: int(GRAVITY_SOUTH), str: "Not-Rotated"},
	{value: int(GRAVITY_NORTH), str: "North"},
	{value: int(GRAVITY_NORTH), str: "Upside-Down"},
	{value: int(GRAVITY_EAST), str: "East"},
	{value: int(GRAVITY_EAST), str: "Rotated-Left"},
	{value: int(GRAVITY_WEST), str: "West"},
	{value: int(GRAVITY_WEST), str: "Rotated-Right"},
}

// IsVertical returns whether `g` represents vertical writing directions.
func (g Gravity) IsVertical() bool {
	return g == GRAVITY_EAST || g == GRAVITY_WEST
}

// IsImproper returns whether `g` represents a gravity that results in reversal of text direction.
func (g Gravity) IsImproper() bool {
	return g == GRAVITY_WEST || g == GRAVITY_NORTH
}

func (g Gravity) String() string {
	return strings.ToLower(GravityMap.ToString("gravity", int(g)))
}

// GravityHint defines how horizontal scripts should behave in a
// vertical context.  That is, English excerpt in a vertical paragraph for
// example.
type GravityHint uint8

const (
	GRAVITY_HINT_NATURAL GravityHint = iota // scripts will take their natural gravity based on the base gravity and the script
	GRAVITY_HINT_STRONG                     // always use the base gravity set, regardless of the script
	// For scripts not in their natural direction (eg.
	// Latin in East gravity), choose per-script gravity such that every script
	// respects the line progression.  This means, Latin and Arabic will take
	// opposite gravities and both flow top-to-bottom for example.
	GRAVITY_HINT_LINE
)

var gravityhint_map = enumMap{
	{value: int(GRAVITY_HINT_NATURAL), str: "natural"},
	{value: int(GRAVITY_HINT_STRONG), str: "strong"},
	{value: int(GRAVITY_HINT_LINE), str: "line"},
}

/**
 * pango_gravity_get_for_script_and_width:
 * @script: Script to query
 * @wide: %true for wide characters as returned by g_unichar_iswide()
 * @base_gravity: base gravity of the paragraph
 * @hint: orientation hint
 *
 * Based on the script, East Asian width, base gravity, and hint,
 * returns actual gravity to use in laying out a single character
 * or Item.
 *
 * pango_gravity_get_for_script_and_width is similar to pango_gravity_get_for_script() except
 * that this function makes a distinction between narrow/half-width and
 * wide/full-width characters also.  Wide/full-width characters always
 * stand <emphasis>upright</emphasis>, that is, they always take the base gravity,
 * whereas narrow/full-width characters are always rotated in vertical
 * context.
 *
 * If @base_gravity is %GRAVITY_AUTO, it is first replaced with the
 * preferred gravity of @script.
 *
 * Return value: resolved gravity suitable to use for a run of text
 * with @script and @wide.
 */
func pango_gravity_get_for_script_and_width(script Script, wide bool,
	base_gravity Gravity, hint GravityHint) Gravity {
	props := get_script_properties(script)

	if base_gravity == GRAVITY_AUTO {
		base_gravity = props.preferredGravity
	}

	vertical := base_gravity.IsVertical()

	// Everything is designed such that a system with no vertical support
	// renders everything correctly horizontally.  So, if not in a vertical
	// gravity, base and resolved gravities are always the same.
	//
	// Wide characters are always upright.
	if !vertical || wide {
		return base_gravity
	}

	// If here, we have a narrow character in a vertical gravity setting.
	// Resolve depending on the hint.
	switch hint {
	case GRAVITY_HINT_STRONG:
		return base_gravity
	case GRAVITY_HINT_LINE:
		if (base_gravity == GRAVITY_EAST) != (props.horizDir == DIRECTION_RTL) {
			return GRAVITY_SOUTH
		}
		return GRAVITY_NORTH
	default:
		if props.vertDir == vectDirNone {
			return GRAVITY_SOUTH
		}
		if (base_gravity == GRAVITY_EAST) != (props.vertDir == vectDirBtt) {
			return GRAVITY_SOUTH
		}
		return GRAVITY_NORTH
	}
}

const (
	vectDirNone = iota
	vectDirTtb
	vectDirBtt
)

type scriptProps struct {
	horizDir Direction /* Orientation in horizontal context */

	vertDir uint8 /* Orientation in vertical context */

	preferredGravity Gravity /* Preferred context gravity */

	// Whether script is mostly wide.
	// Wide characters are upright (ie. not rotated) in foreign context
	wide bool
}

const (
	ltr  = DIRECTION_LTR
	rtl  = DIRECTION_RTL
	weak = DIRECTION_WEAK_LTR
)

var scriptProperties = map[Script]scriptProps{ /* ISO 15924 code */
	language.Common:              {},                                       /* Zyyy */
	language.Inherited:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Qaai */
	language.Arabic:              {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Arab */
	language.Armenian:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Armn */
	language.Bengali:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Beng */
	language.Bopomofo:            {ltr, vectDirTtb, GRAVITY_EAST, true},    /* Bopo */
	language.Cherokee:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Cher */
	language.Coptic:              {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Qaac */
	language.Cyrillic:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Cyrl (Cyrs) */
	language.Deseret:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Dsrt */
	language.Devanagari:          {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Deva */
	language.Ethiopic:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Ethi */
	language.Georgian:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Geor (Geon, Geoa) */
	language.Gothic:              {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Goth */
	language.Greek:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Grek */
	language.Gujarati:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Gujr */
	language.Gurmukhi:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Guru */
	language.Han:                 {ltr, vectDirTtb, GRAVITY_EAST, true},    /* Hani */
	language.Hangul:              {ltr, vectDirTtb, GRAVITY_EAST, true},    /* Hang */
	language.Hebrew:              {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Hebr */
	language.Hiragana:            {ltr, vectDirTtb, GRAVITY_EAST, true},    /* Hira */
	language.Kannada:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Knda */
	language.Katakana:            {ltr, vectDirTtb, GRAVITY_EAST, true},    /* Kana */
	language.Khmer:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Khmr */
	language.Lao:                 {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Laoo */
	language.Latin:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Latn (Latf, Latg) */
	language.Malayalam:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Mlym */
	language.Mongolian:           {weak, vectDirTtb, GRAVITY_WEST, false},  /* Mong */
	language.Myanmar:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Mymr */
	language.Ogham:               {ltr, vectDirBtt, GRAVITY_WEST, false},   /* Ogam */
	language.Old_Italic:          {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Ital */
	language.Oriya:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Orya */
	language.Runic:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Runr */
	language.Sinhala:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Sinh */
	language.Syriac:              {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Syrc (Syrj, Syrn, Syre) */
	language.Tamil:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Taml */
	language.Telugu:              {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Telu */
	language.Thaana:              {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Thaa */
	language.Thai:                {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Thai */
	language.Tibetan:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tibt */
	language.Canadian_Aboriginal: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Cans */
	language.Yi:                  {ltr, vectDirTtb, GRAVITY_SOUTH, true},   /* Yiii */
	language.Tagalog:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tglg */
	language.Hanunoo:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Hano */
	language.Buhid:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Buhd */
	language.Tagbanwa:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tagb */

	/* Unicode-4.0 additions */
	language.Braille:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Brai */
	language.Cypriot:  {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Cprt */
	language.Limbu:    {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Limb */
	language.Osmanya:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Osma */
	language.Shavian:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Shaw */
	language.Linear_B: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Linb */
	language.Tai_Le:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tale */
	language.Ugaritic: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Ugar */

	/* Unicode-4.1 additions */
	language.New_Tai_Lue:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Talu */
	language.Buginese:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Bugi */
	language.Glagolitic:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Glag */
	language.Tifinagh:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tfng */
	language.Syloti_Nagri: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Sylo */
	language.Old_Persian:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Xpeo */
	language.Kharoshthi:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Khar */

	/* Unicode-5.0 additions */
	language.Unknown:    {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Zzzz */
	language.Balinese:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Bali */
	language.Cuneiform:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Xsux */
	language.Phoenician: {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Phnx */
	language.Phags_Pa:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Phag */
	language.Nko:        {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Nkoo */
}

// TODO: cleanup
func get_script_properties(script Script) scriptProps { return scriptProperties[script] }
