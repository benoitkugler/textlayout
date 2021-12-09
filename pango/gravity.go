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
	{value: int(GRAVITY_SOUTH), str: "south"},
	{value: int(GRAVITY_EAST), str: "east"},
	{value: int(GRAVITY_NORTH), str: "north"},
	{value: int(GRAVITY_WEST), str: "west"},
	{value: int(GRAVITY_AUTO), str: "auto"},
	// {value: int(GRAVITY_SOUTH), str: "Not-Rotated"},
	// {value: int(GRAVITY_NORTH), str: "Upside-Down"},
	// {value: int(GRAVITY_EAST), str: "Rotated-Left"},
	// {value: int(GRAVITY_WEST), str: "Rotated-Right"},
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
	props := getScriptProperties(script)

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
	language.Kharoshthi:   {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Khar */

	/* Unicode-5.0 additions */
	language.Unknown:    {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Zzzz */
	language.Balinese:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Bali */
	language.Cuneiform:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Xsux */
	language.Phoenician: {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Phnx */
	language.Phags_Pa:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Phag */
	language.Nko:        {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Nkoo */

	/* Unicode-5.1 additions */
	language.Kayah_Li:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Kali */
	language.Lepcha:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Lepc */
	language.Rejang:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Rjng */
	language.Sundanese:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Sund */
	language.Saurashtra: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Saur */
	language.Cham:       {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Cham */
	language.Ol_Chiki:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Olck */
	language.Vai:        {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Vaii */
	language.Carian:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Cari */
	language.Lycian:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Lyci */
	language.Lydian:     {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Lydi */

	/* Unicode-5.2 additions */
	language.Avestan:                {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Avst */
	language.Bamum:                  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Bamu */
	language.Egyptian_Hieroglyphs:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Egyp */
	language.Imperial_Aramaic:       {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Armi */
	language.Inscriptional_Pahlavi:  {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Phli */
	language.Inscriptional_Parthian: {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Prti */
	language.Javanese:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Java */
	language.Kaithi:                 {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Kthi */
	language.Lisu:                   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Lisu */
	language.Meetei_Mayek:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Mtei */
	language.Old_South_Arabian:      {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Sarb */
	language.Old_Turkic:             {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Orkh */
	language.Samaritan:              {rtl, vectDirTtb, GRAVITY_SOUTH, false},  /* Samr */
	language.Tai_Tham:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Lana */
	language.Tai_Viet:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tavt */

	/* Unicode-6.0 additions */
	language.Batak:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Batk */
	language.Brahmi:  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Brah */
	language.Mandaic: {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Mand */

	/* Unicode-6.1 additions */
	language.Chakma:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Cakm */
	language.Meroitic_Cursive:     {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Merc */
	language.Meroitic_Hieroglyphs: {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Mero */
	language.Miao:                 {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Plrd */
	language.Sharada:              {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Shrd */
	language.Sora_Sompeng:         {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Sora */
	language.Takri:                {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Takr */

	/* Unicode-7.0 additions */
	language.Bassa_Vah:          {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Bass */
	language.Caucasian_Albanian: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Aghb */
	language.Duployan:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Dupl */
	language.Elbasan:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Elba */
	language.Grantha:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Gran */
	language.Khojki:             {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Khoj */
	language.Khudawadi:          {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Sind */
	language.Linear_A:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Lina */
	language.Mahajani:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Mahj */
	language.Manichaean:         {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Mani */
	language.Mende_Kikakui:      {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Mend */
	language.Modi:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Modi */
	language.Mro:                {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Mroo */
	language.Nabataean:          {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Nbat */
	language.Old_North_Arabian:  {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Narb */
	language.Old_Permic:         {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Perm */
	language.Pahawh_Hmong:       {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Hmng */
	language.Palmyrene:          {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Palm */
	language.Pau_Cin_Hau:        {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Pauc */
	language.Psalter_Pahlavi:    {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Phlp */
	language.Siddham:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Sidd */
	language.Tirhuta:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tirh */
	language.Warang_Citi:        {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Wara */

	/* Unicode-8.0 additions */
	language.Ahom:                  {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Ahom */
	language.Anatolian_Hieroglyphs: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Hluw */
	language.Hatran:                {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Hatr */
	language.Multani:               {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Mult */
	language.Old_Hungarian:         {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Hung */
	language.SignWriting:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Sgnw */

	/* Unicode-9.0 additions */
	language.Adlam:     {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Adlm */
	language.Bhaiksuki: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Bhks */
	language.Marchen:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Marc */
	language.Newa:      {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Newa */
	language.Osage:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Osge */
	language.Tangut:    {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tang */

	/* Unicode-10.0 additions */
	language.Masaram_Gondi:    {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Gonm */
	language.Nushu:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Nshu */
	language.Soyombo:          {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Soyo */
	language.Zanabazar_Square: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Zanb */

	/* Unicode-11.0 additions */
	language.Dogra:           {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Dogr */
	language.Gunjala_Gondi:   {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Gong */
	language.Hanifi_Rohingya: {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Rohg */
	language.Makasar:         {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Maka */
	language.Medefaidrin:     {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Medf */
	language.Old_Sogdian:     {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Sogo */
	language.Sogdian:         {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Sogd */

	/* Unicode-12.0 additions */
	language.Elymaic:                {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Elym */
	language.Nandinagari:            {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Nand */
	language.Nyiakeng_Puachue_Hmong: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Rohg */
	language.Wancho:                 {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Wcho */

	/* Unicode-13.0 additions */
	language.Chorasmian:          {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Chrs */
	language.Dives_Akuru:         {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Diak */
	language.Khitan_Small_Script: {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Kits */
	language.Yezidi:              {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Yezi */

	// TODO: Unicode 14
	// {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Cpmn */
	// {rtl, vectDirNone, GRAVITY_SOUTH, false}, /* Ougr */
	// {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Tnsa */
	// {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Toto */
	// {ltr, vectDirNone, GRAVITY_SOUTH, false}, /* Vith */
}

func getScriptProperties(script Script) scriptProps { return scriptProperties[script] }

// finds the gravity that best matches the rotation component
// in a `matrix`.
//
// Returns the gravity of `matrix`, which will never be
// GRAVITY_AUTO, or GRAVITY_SOUTH if `matrix` is nil.
func gravityFromMatrix(matrix *Matrix) Gravity {
	if matrix == nil {
		return GRAVITY_SOUTH
	}
	x := matrix.Xy
	y := matrix.Yy

	if fabs(x) > fabs(y) {
		if x > 0 {
			return GRAVITY_WEST
		}
		return GRAVITY_EAST
	} else {
		if y < 0 {
			return GRAVITY_NORTH
		}
		return GRAVITY_SOUTH
	}
}
