package pango

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

var gravity_map = enumMap{
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
func (g Gravity) isVertical() bool {
	return g == PANGO_GRAVITY_EAST || g == PANGO_GRAVITY_WEST
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

	vertical := base_gravity.isVertical()

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
	SCRIPT_COMMON:              {},                                             /* Zyyy */
	SCRIPT_INHERITED:           {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Qaai */
	SCRIPT_ARABIC:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Arab */
	SCRIPT_ARMENIAN:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Armn */
	SCRIPT_BENGALI:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Beng */
	SCRIPT_BOPOMOFO:            {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Bopo */
	SCRIPT_CHEROKEE:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cher */
	SCRIPT_COPTIC:              {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Qaac */
	SCRIPT_CYRILLIC:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cyrl (Cyrs) */
	SCRIPT_DESERET:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Dsrt */
	SCRIPT_DEVANAGARI:          {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Deva */
	SCRIPT_ETHIOPIC:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Ethi */
	SCRIPT_GEORGIAN:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Geor (Geon, Geoa) */
	SCRIPT_GOTHIC:              {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Goth */
	SCRIPT_GREEK:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Grek */
	SCRIPT_GUJARATI:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Gujr */
	SCRIPT_GURMUKHI:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Guru */
	SCRIPT_HAN:                 {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Hani */
	SCRIPT_HANGUL:              {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Hang */
	SCRIPT_HEBREW:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Hebr */
	SCRIPT_HIRAGANA:            {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Hira */
	SCRIPT_KANNADA:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Knda */
	SCRIPT_KATAKANA:            {LTR, vectDirTtb, PANGO_GRAVITY_EAST, true},    /* Kana */
	SCRIPT_KHMER:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Khmr */
	SCRIPT_LAO:                 {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Laoo */
	SCRIPT_LATIN:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Latn (Latf, Latg) */
	SCRIPT_MALAYALAM:           {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Mlym */
	SCRIPT_MONGOLIAN:           {WEAK, vectDirTtb, PANGO_GRAVITY_WEST, false},  /* Mong */
	SCRIPT_MYANMAR:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Mymr */
	SCRIPT_OGHAM:               {LTR, vectDirBtt, PANGO_GRAVITY_WEST, false},   /* Ogam */
	SCRIPT_OLD_ITALIC:          {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Ital */
	SCRIPT_ORIYA:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Orya */
	SCRIPT_RUNIC:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Runr */
	SCRIPT_SINHALA:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Sinh */
	SCRIPT_SYRIAC:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Syrc (Syrj, Syrn, Syre) */
	SCRIPT_TAMIL:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Taml */
	SCRIPT_TELUGU:              {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Telu */
	SCRIPT_THAANA:              {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Thaa */
	SCRIPT_THAI:                {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Thai */
	SCRIPT_TIBETAN:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tibt */
	SCRIPT_CANADIAN_ABORIGINAL: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cans */
	SCRIPT_YI:                  {LTR, vectDirTtb, PANGO_GRAVITY_SOUTH, true},   /* Yiii */
	SCRIPT_TAGALOG:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tglg */
	SCRIPT_HANUNOO:             {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Hano */
	SCRIPT_BUHID:               {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Buhd */
	SCRIPT_TAGBANWA:            {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tagb */

	/* Unicode-4.0 additions */
	SCRIPT_BRAILLE:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Brai */
	SCRIPT_CYPRIOT:  {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Cprt */
	SCRIPT_LIMBU:    {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Limb */
	SCRIPT_OSMANYA:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Osma */
	SCRIPT_SHAVIAN:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Shaw */
	SCRIPT_LINEAR_B: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Linb */
	SCRIPT_TAI_LE:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tale */
	SCRIPT_UGARITIC: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Ugar */

	/* Unicode-4.1 additions */
	SCRIPT_NEW_TAI_LUE:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Talu */
	SCRIPT_BUGINESE:     {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Bugi */
	SCRIPT_GLAGOLITIC:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Glag */
	SCRIPT_TIFINAGH:     {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Tfng */
	SCRIPT_SYLOTI_NAGRI: {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Sylo */
	SCRIPT_OLD_PERSIAN:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Xpeo */
	SCRIPT_KHAROSHTHI:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Khar */

	/* Unicode-5.0 additions */
	SCRIPT_UNKNOWN:    {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Zzzz */
	SCRIPT_BALINESE:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Bali */
	SCRIPT_CUNEIFORM:  {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Xsux */
	SCRIPT_PHOENICIAN: {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Phnx */
	SCRIPT_PHAGS_PA:   {LTR, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Phag */
	SCRIPT_NKO:        {RTL, vectDirNone, PANGO_GRAVITY_SOUTH, false}, /* Nkoo */
}

// TODO: cleanup
func get_script_properties(script Script) ScriptProperties { return script_properties[script] }
