package pango

import "github.com/benoitkugler/textlayout/language"

// Gravity represents the orientation of glyphs in a segment
// of text.  This is useful when rendering vertical text layouts.  In
// those situations, the layout is rotated using a non-identity PangoMatrix,
// and then glyph orientation is controlled using Gravity.
//
// Not every value in this enumeration makes sense for every usage of
// Gravity; for example, `PANGO_GRAVITY_AUTO` only can be passed to
// pango_context_set_base_gravity() and can only be returned by
// pango_context_get_base_gravity().
//
// See also: GravityHint
type Gravity uint8

const (
	PANGO_GRAVITY_SOUTH Gravity = iota // Glyphs stand upright (default)
	PANGO_GRAVITY_EAST                 // Glyphs are rotated 90 degrees clockwise
	PANGO_GRAVITY_NORTH                // Glyphs are upside-down
	PANGO_GRAVITY_WEST                 // Glyphs are rotated 90 degrees counter-clockwise
	PANGO_GRAVITY_AUTO                 // Gravity is resolved from the context matrix
)

var GravityMap = enumMap{
	{value: int(PANGO_GRAVITY_SOUTH), str: "Not-Rotated"},
	{value: int(PANGO_GRAVITY_SOUTH), str: "South"},
	{value: int(PANGO_GRAVITY_NORTH), str: "Upside-Down"},
	{value: int(PANGO_GRAVITY_NORTH), str: "North"},
	{value: int(PANGO_GRAVITY_EAST), str: "Rotated-Left"},
	{value: int(PANGO_GRAVITY_EAST), str: "East"},
	{value: int(PANGO_GRAVITY_WEST), str: "Rotated-Right"},
	{value: int(PANGO_GRAVITY_WEST), str: "West"},
}

// whether `g` represents vertical writing directions.
func (g Gravity) IsVertical() bool {
	return g == PANGO_GRAVITY_EAST || g == PANGO_GRAVITY_WEST
}

// IsImproper returns whether a `Gravity` represents a gravity that results in reversal of text direction.
func (gravity Gravity) IsImproper() bool {
	return gravity == PANGO_GRAVITY_WEST || gravity == PANGO_GRAVITY_NORTH
}

// GravityHint defines how horizontal scripts should behave in a
// vertical context.  That is, English excerpt in a vertical paragraph for
// example.
type GravityHint uint8

const (
	PANGO_GRAVITY_HINT_NATURAL GravityHint = iota // scripts will take their natural gravity based on the base gravity and the script
	PANGO_GRAVITY_HINT_STRONG                     // always use the base gravity set, regardless of the script
	// For scripts not in their natural direction (eg.
	// Latin in East gravity), choose per-script gravity such that every script
	// respects the line progression.  This means, Latin and Arabic will take
	// opposite gravities and both flow top-to-bottom for example.
	PANGO_GRAVITY_HINT_LINE
)

var gravityhint_map = enumMap{
	{value: int(PANGO_GRAVITY_HINT_NATURAL), str: "natural"},
	{value: int(PANGO_GRAVITY_HINT_STRONG), str: "strong"},
	{value: int(PANGO_GRAVITY_HINT_LINE), str: "line"},
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
 * If @base_gravity is %PANGO_GRAVITY_AUTO, it is first replaced with the
 * preferred gravity of @script.
 *
 * Return value: resolved gravity suitable to use for a run of text
 * with @script and @wide.
 */
func pango_gravity_get_for_script_and_width(script Script, wide bool,
	base_gravity Gravity, hint GravityHint) Gravity {
	props := get_script_properties(script)

	if base_gravity == PANGO_GRAVITY_AUTO {
		base_gravity = props.preferred_gravity
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
	case PANGO_GRAVITY_HINT_STRONG:
		return base_gravity
	case PANGO_GRAVITY_HINT_LINE:
		if (base_gravity == PANGO_GRAVITY_EAST) != (props.horiz_dir == PANGO_DIRECTION_RTL) {
			return PANGO_GRAVITY_SOUTH
		}
		return PANGO_GRAVITY_NORTH
	default:
		if props.vert_dir == vectDirNone {
			return PANGO_GRAVITY_SOUTH
		}
		if (base_gravity == PANGO_GRAVITY_EAST) != (props.vert_dir == vectDirBtt) {
			return PANGO_GRAVITY_SOUTH
		}
		return PANGO_GRAVITY_NORTH
	}
}

const (
	vectDirNone = iota
	vectDirTtb
	vectDirBtt
)

type ScriptProperties struct {
	horiz_dir Direction /* Orientation in horizontal context */

	vert_dir uint8 /* Orientation in vertical context */

	preferred_gravity Gravity /* Preferred context gravity */

	// Whether script is mostly wide.
	// Wide characters are upright (ie. not rotated) in foreign context
	wide bool
}

const (
	LTR  = PANGO_DIRECTION_LTR
	RTL  = PANGO_DIRECTION_RTL
	WEAK = PANGO_DIRECTION_WEAK_LTR
)

var script_properties = map[Script]ScriptProperties{ /* ISO 15924 code */
	language.Common:              {},                                             /* Zyyy */
	language.Inherited:           {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Qaai */
	language.Arabic:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Arab */
	language.Armenian:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Armn */
	language.Bengali:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Beng */
	language.Bopomofo:            {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Bopo */
	language.Cherokee:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cher */
	language.Coptic:              {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Qaac */
	language.Cyrillic:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cyrl (Cyrs) */
	language.Deseret:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Dsrt */
	language.Devanagari:          {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Deva */
	language.Ethiopic:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Ethi */
	language.Georgian:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Geor (Geon, Geoa) */
	language.Gothic:              {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Goth */
	language.Greek:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Grek */
	language.Gujarati:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Gujr */
	language.Gurmukhi:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Guru */
	language.Han:                 {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Hani */
	language.Hangul:              {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Hang */
	language.Hebrew:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Hebr */
	language.Hiragana:            {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Hira */
	language.Kannada:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Knda */
	language.Katakana:            {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Kana */
	language.Khmer:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Khmr */
	language.Lao:                 {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Laoo */
	language.Latin:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Latn (Latf, Latg) */
	language.Malayalam:           {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Mlym */
	language.Mongolian:           {WEAK, vectDirTtb, PANGO_GRAVITY_WEST, false},  /* Mong */
	language.Myanmar:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Mymr */
	language.Ogham:               {LTR, vectDirBtt, PANGO_GRAVITY_WEST, false},   /* Ogam */
	language.Old_Italic:          {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Ital */
	language.Oriya:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Orya */
	language.Runic:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Runr */
	language.Sinhala:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Sinh */
	language.Syriac:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Syrc (Syrj, Syrn, Syre) */
	language.Tamil:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Taml */
	language.Telugu:              {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Telu */
	language.Thaana:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Thaa */
	language.Thai:                {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Thai */
	language.Tibetan:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tibt */
	language.Canadian_Aboriginal: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cans */
	language.Yi:                  {LTR, vectDirTtb, PANGO_GRAVITY_SOUTH, true},   /* Yiii */
	language.Tagalog:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tglg */
	language.Hanunoo:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Hano */
	language.Buhid:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Buhd */
	language.Tagbanwa:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tagb */

	/* Unicode-4.0 additions */
	language.Braille:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Brai */
	language.Cypriot:  {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cprt */
	language.Limbu:    {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Limb */
	language.Osmanya:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Osma */
	language.Shavian:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Shaw */
	language.Linear_B: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Linb */
	language.Tai_Le:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tale */
	language.Ugaritic: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Ugar */

	/* Unicode-4.1 additions */
	language.New_Tai_Lue:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Talu */
	language.Buginese:     {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Bugi */
	language.Glagolitic:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Glag */
	language.Tifinagh:     {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tfng */
	language.Syloti_Nagri: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Sylo */
	language.Old_Persian:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Xpeo */
	language.Kharoshthi:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Khar */

	/* Unicode-5.0 additions */
	language.Unknown:    {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Zzzz */
	language.Balinese:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Bali */
	language.Cuneiform:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Xsux */
	language.Phoenician: {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Phnx */
	language.Phags_Pa:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Phag */
	language.Nko:        {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Nkoo */
}

// TODO: cleanup
func get_script_properties(script Script) ScriptProperties { return script_properties[script] }
