package pango

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"
)

const (
	// The `PANGO_GLYPH_EMPTY` macro represents a `Glyph` value that has a
	// special meaning, which is a zero-width empty glyph. This is useful for
	// example in shaper modules, to use as the glyph for various zero-width
	// Unicode characters (those passing pango_is_zero_width()).
	PANGO_GLYPH_EMPTY Glyph = 0x0FFFFFFF
	// The `PANGO_GLYPH_INVALID_INPUT` macro represents a `Glyph` value that has a
	// special meaning of invalid input. `Layout` produces one such glyph
	// per invalid input UTF-8 byte and such a glyph is rendered as a crossed
	// box.
	//
	// Note that this value is defined such that it has the `PANGO_GLYPH_UNKNOWN_FLAG` on.
	PANGO_GLYPH_INVALID_INPUT Glyph = 0xFFFFFFFF

	// The `PANGO_GLYPH_UNKNOWN_FLAG` macro is a flag value that can be added to
	// a rune value of a valid Unicode character, to produce a `Glyph`
	// value, representing an unknown-character glyph for the respective rune.
	PANGO_GLYPH_UNKNOWN_FLAG = 0x10000000
)

// PANGO_GET_UNKNOWN_GLYPH returns a `Glyph` value that means no glyph was found for `wc`.
//
// The way this unknown glyphs are rendered is backend specific.  For example,
// a box with the hexadecimal Unicode code-point of the character written in it
// is what is done in the most common backends.
func PANGO_GET_UNKNOWN_GLYPH(wc rune) Glyph {
	return Glyph(wc | PANGO_GLYPH_UNKNOWN_FLAG)
}

// contains acceptable strings value
type enumMap []struct {
	value int
	str   string
}

func (e enumMap) fromString(str string) (int, bool) {
	for _, v := range e {
		if field_matches(v.str, str) {
			return v.value, true
		}
	}
	return 0, false
}

// if v is not found, it is printed as "what=v"
func (e enumMap) toString(what string, v int) string {
	for _, entry := range e {
		if entry.value == v {
			return entry.str
		}
	}
	return fmt.Sprintf("%s=%d", what, v)
}

func (map_ enumMap) parse_field(str string) (int, bool) {
	if str == "" {
		return 0, false
	}

	if field_matches("Normal", str) {
		str = ""
	}

	v, found := map_.fromString(str)

	return v, found
}

func (map_ enumMap) possibleValues() string {
	var values []string
	for _, v := range map_ {
		if v.str != "" {
			values = append(values, v.str)
		}
	}
	return strings.Join(values, "/")
}

// Style specifies the various slant styles possible for a font.
type Style uint8

const (
	PANGO_STYLE_NORMAL  Style = iota //  the font is upright.
	PANGO_STYLE_OBLIQUE              //  the font is slanted, but in a roman style.
	PANGO_STYLE_ITALIC               //  the font is slanted in an italic style.
)

var style_map = enumMap{
	{value: int(PANGO_STYLE_NORMAL), str: ""},
	{value: int(PANGO_STYLE_NORMAL), str: "Roman"},
	{value: int(PANGO_STYLE_OBLIQUE), str: "Oblique"},
	{value: int(PANGO_STYLE_ITALIC), str: "Italic"},
}

// Variant specifies capitalization variant of the font.
type Variant uint8

const (
	PANGO_VARIANT_NORMAL     Variant = iota // A normal font.
	PANGO_VARIANT_SMALL_CAPS                // A font with the lower case characters replaced by smaller variants of the capital characters.
)

var variant_map = enumMap{
	{value: int(PANGO_VARIANT_NORMAL), str: ""},
	{value: int(PANGO_VARIANT_SMALL_CAPS), str: "Small-Caps"},
}

//  Weight specifies the weight (boldness) of a font. This is a numerical
//  value ranging from 100 to 1000, but there are some predefined values:
type Weight int

const (
	PANGO_WEIGHT_THIN       Weight = 100  // the thin weight (= 100; Since: 1.24)
	PANGO_WEIGHT_ULTRALIGHT Weight = 200  // the ultralight weight (= 200)
	PANGO_WEIGHT_LIGHT      Weight = 300  // the light weight (= 300)
	PANGO_WEIGHT_SEMILIGHT  Weight = 350  // the semilight weight (= 350; Since: 1.36.7)
	PANGO_WEIGHT_BOOK       Weight = 380  // the book weight (= 380; Since: 1.24)
	PANGO_WEIGHT_NORMAL     Weight = 400  // the default weight (= 400)
	PANGO_WEIGHT_MEDIUM     Weight = 500  // the normal weight (= 500; Since: 1.24)
	PANGO_WEIGHT_SEMIBOLD   Weight = 600  // the semibold weight (= 600)
	PANGO_WEIGHT_BOLD       Weight = 700  // the bold weight (= 700)
	PANGO_WEIGHT_ULTRABOLD  Weight = 800  // the ultrabold weight (= 800)
	PANGO_WEIGHT_HEAVY      Weight = 900  // the heavy weight (= 900)
	PANGO_WEIGHT_ULTRAHEAVY Weight = 1000 // the ultraheavy weight (= 1000; Since: 1.24)
)

var weight_map = enumMap{
	{value: int(PANGO_WEIGHT_THIN), str: "Thin"},
	{value: int(PANGO_WEIGHT_ULTRALIGHT), str: "Ultra-Light"},
	{value: int(PANGO_WEIGHT_ULTRALIGHT), str: "Extra-Light"},
	{value: int(PANGO_WEIGHT_LIGHT), str: "Light"},
	{value: int(PANGO_WEIGHT_SEMILIGHT), str: "Semi-Light"},
	{value: int(PANGO_WEIGHT_SEMILIGHT), str: "Demi-Light"},
	{value: int(PANGO_WEIGHT_BOOK), str: "Book"},
	{value: int(PANGO_WEIGHT_NORMAL), str: ""},
	{value: int(PANGO_WEIGHT_NORMAL), str: "Regular"},
	{value: int(PANGO_WEIGHT_MEDIUM), str: "Medium"},
	{value: int(PANGO_WEIGHT_SEMIBOLD), str: "Semi-Bold"},
	{value: int(PANGO_WEIGHT_SEMIBOLD), str: "Demi-Bold"},
	{value: int(PANGO_WEIGHT_BOLD), str: "Bold"},
	{value: int(PANGO_WEIGHT_ULTRABOLD), str: "Ultra-Bold"},
	{value: int(PANGO_WEIGHT_ULTRABOLD), str: "Extra-Bold"},
	{value: int(PANGO_WEIGHT_HEAVY), str: "Heavy"},
	{value: int(PANGO_WEIGHT_HEAVY), str: "Black"},
	{value: int(PANGO_WEIGHT_ULTRAHEAVY), str: "Ultra-Heavy"},
	{value: int(PANGO_WEIGHT_ULTRAHEAVY), str: "Extra-Heavy"},
	{value: int(PANGO_WEIGHT_ULTRAHEAVY), str: "Ultra-Black"},
	{value: int(PANGO_WEIGHT_ULTRAHEAVY), str: "Extra-Black"},
}

//  Stretch specifies the width of the font relative to other designs within a family.
type Stretch uint8

const (
	PANGO_STRETCH_ULTRA_CONDENSED Stretch = iota // ultra condensed width
	PANGO_STRETCH_EXTRA_CONDENSED                // extra condensed width
	PANGO_STRETCH_CONDENSED                      // condensed width
	PANGO_STRETCH_SEMI_CONDENSED                 // semi condensed width
	PANGO_STRETCH_NORMAL                         // the normal width
	PANGO_STRETCH_SEMI_EXPANDED                  // semi expanded width
	PANGO_STRETCH_EXPANDED                       // expanded width
	PANGO_STRETCH_EXTRA_EXPANDED                 // extra expanded width
	PANGO_STRETCH_ULTRA_EXPANDED                 // ultra expanded width
)

var stretch_map = enumMap{
	{value: int(PANGO_STRETCH_ULTRA_CONDENSED), str: "Ultra-Condensed"},
	{value: int(PANGO_STRETCH_EXTRA_CONDENSED), str: "Extra-Condensed"},
	{value: int(PANGO_STRETCH_CONDENSED), str: "Condensed"},
	{value: int(PANGO_STRETCH_SEMI_CONDENSED), str: "Semi-Condensed"},
	{value: int(PANGO_STRETCH_NORMAL), str: ""},
	{value: int(PANGO_STRETCH_SEMI_EXPANDED), str: "Semi-Expanded"},
	{value: int(PANGO_STRETCH_EXPANDED), str: "Expanded"},
	{value: int(PANGO_STRETCH_EXTRA_EXPANDED), str: "Extra-Expanded"},
	{value: int(PANGO_STRETCH_ULTRA_EXPANDED), str: "Ultra-Expanded"},
}

// FontMask bits correspond to fields in a `FontDescription` that have been set.
type FontMask int16

const (
	PANGO_FONT_MASK_FAMILY     FontMask = 1 << iota // the font family is specified.
	PANGO_FONT_MASK_STYLE                           // the font style is specified.
	PANGO_FONT_MASK_VARIANT                         // the font variant is specified.
	PANGO_FONT_MASK_WEIGHT                          // the font weight is specified.
	PANGO_FONT_MASK_STRETCH                         // the font stretch is specified.
	PANGO_FONT_MASK_SIZE                            // the font size is specified.
	PANGO_FONT_MASK_GRAVITY                         // the font gravity is specified (Since: 1.16.)
	PANGO_FONT_MASK_VARIATIONS                      // OpenType font variations are specified (Since: 1.42)
)

/* CSS scale factors (1.2 factor between each size) */
const (
	pangoScale_XX_SMALL = 0.5787037037037 //  The scale factor for three shrinking steps (1 / (1.2 * 1.2 * 1.2)).
	pangoScale_X_SMALL  = 0.6944444444444 //  The scale factor for two shrinking steps (1 / (1.2 * 1.2)).
	pangoScale_SMALL    = 0.8333333333333 //  The scale factor for one shrinking step (1 / 1.2).
	pangoScale_MEDIUM   = 1.0             //  The scale factor for normal size (1.0).
	pangoScale_LARGE    = 1.2             //  The scale factor for one magnification step (1.2).
	pangoScale_X_LARGE  = 1.44            //  The scale factor for two magnification steps (1.2 * 1.2).
	pangoScale_XX_LARGE = 1.728           //  The scale factor for three magnification steps (1.2 * 1.2 * 1.2).
)

var pfd_defaults = FontDescription{
	family_name: "",

	style:      PANGO_STYLE_NORMAL,
	variant:    PANGO_VARIANT_NORMAL,
	weight:     PANGO_WEIGHT_NORMAL,
	stretch:    PANGO_STRETCH_NORMAL,
	gravity:    PANGO_GRAVITY_SOUTH,
	variations: "",

	mask:             0,
	size_is_absolute: false,

	size: 0,
}

// Font is used to represent a font in a rendering-system-independent matter.
// The concretes types implementing this interface shouls be pointers, since
// they will be used as map keys: they MUST at least be comparable types.
type Font interface {
	describe() FontDescription
	get_coverage(language Language) Coverage
	get_glyph_extents(glyph Glyph, inkRect, logicalRect *Rectangle)
	get_metrics(language Language) FontMetrics
	get_font_map() FontMap
	describe_absolute() FontDescription
	// get_features       (,  hb_feature_t   *features,				 guint           len,				 guint          *num_features) void
	// create_hb_font     () hb_font_t *
}

// pango_font_has_char returns whether the font provides a glyph for this character.
// `font` must not be nil
func pango_font_has_char(font Font, wc rune) bool {
	coverage := font.get_coverage(pango_language_get_default())
	result := coverage.get(wc)
	return result != PANGO_COVERAGE_NONE
}

// FontDescription represents the description
// of an ideal font. These structures are used both to list
// what fonts are available on the system and also for specifying
// the characteristics of a font to load.
// This struct does not hold any pointer types: it can be copied by value.
type FontDescription struct {
	family_name string

	style   Style
	variant Variant
	weight  Weight
	stretch Stretch
	gravity Gravity

	variations string

	mask             FontMask
	size_is_absolute bool // = : 1;

	size int
}

// pango_font_description_new creates a new font description structure with all fields unset.
func pango_font_description_new() FontDescription {
	return pfd_defaults // copy
}

// pango_font_description_from_string creates a new font description from a string representation in the
// form
//
// "[FAMILY-LIST] [STYLE-OPTIONS] [SIZE] [VARIATIONS]",
//
// where FAMILY-LIST is a comma-separated list of families optionally
// terminated by a comma, STYLE_OPTIONS is a whitespace-separated list
// of words where each word describes one of style, variant, weight,
// stretch, or gravity, and SIZE is a decimal number (size in points)
// or optionally followed by the unit modifier "px" for absolute size.
// VARIATIONS is a comma-separated list of font variation
// specifications of the form "@axis=value" (the = sign is optional).
//
// The following words are understood as styles:
// "Normal", "Roman", "Oblique", "Italic".
//
// The following words are understood as variants:
// "Small-Caps".
//
// The following words are understood as weights:
// "Thin", "Ultra-Light", "Extra-Light", "Light", "Semi-Light",
// "Demi-Light", "Book", "Regular", "Medium", "Semi-Bold", "Demi-Bold",
// "Bold", "Ultra-Bold", "Extra-Bold", "Heavy", "Black", "Ultra-Black",
// "Extra-Black".
//
// The following words are understood as stretch values:
// "Ultra-Condensed", "Extra-Condensed", "Condensed", "Semi-Condensed",
// "Semi-Expanded", "Expanded", "Extra-Expanded", "Ultra-Expanded".
//
// The following words are understood as gravity values:
// "Not-Rotated", "South", "Upside-Down", "North", "Rotated-Left",
// "East", "Rotated-Right", "West".
//
// Any one of the options may be absent. If FAMILY-LIST is absent, then
// the family_name field of the resulting font description will be
// initialized to nil. If STYLE-OPTIONS is missing, then all style
// options will be set to the default values. If SIZE is missing, the
// size in the resulting font description will be set to 0.
//
// A default value is returned on invalid or empty input.
//
// A typical example:
//
// "Cantarell Italic Light 15 @wght=200"
func pango_font_description_from_string(str string) FontDescription {
	desc := pango_font_description_new()

	desc.mask = PANGO_FONT_MASK_STYLE |
		PANGO_FONT_MASK_WEIGHT |
		PANGO_FONT_MASK_VARIANT |
		PANGO_FONT_MASK_STRETCH

	fields := strings.Fields(str)
	if len(fields) == 0 {
		return desc
	}

	// Look for variations at the end of the string */
	if word := fields[len(fields)-1]; word[0] == '@' {
		/* XXX: actually validate here */
		desc.variations = word[1:]
		desc.mask |= PANGO_FONT_MASK_VARIATIONS
		fields = fields[:len(fields)-1]
	}

	/* Look for a size */
	if len(fields) != 0 {
		word := fields[len(fields)-1]
		var size_is_absolute bool
		if strings.HasSuffix(word, "px") {
			size_is_absolute = true
			word = strings.TrimSuffix(word, "px")
		}
		size, err := strconv.ParseFloat(word, 64)
		if err != nil {
			// just ignore invalid floats: they maybe do not refers to a size
		} else if size < 0 || size > 1000000 {
			log.Println("invalid size value:", size)
		} else { // word is a valid float
			desc.size = int(size*pangoScale + 0.5)
			desc.size_is_absolute = size_is_absolute
			desc.mask |= PANGO_FONT_MASK_SIZE
			fields = fields[:len(fields)-1]
		}

	}

	// Now parse style words
	for len(fields) != 0 {
		word := fields[len(fields)-1]
		if !desc.find_field_any(word) {
			break
		} else {
			fields = fields[:len(fields)-1]
		}
	}

	// Remainder is family list. Trim off trailing commas and leading and trailing white space
	if len(fields) != 0 {
		families := strings.Split(strings.Join(fields, " "), ",")
		/* Now sanitize it to trim space from around individual family names.
		* bug #499624 */
		for i, f := range families {
			families[i] = strings.TrimSpace(f)
		}
		desc.family_name = strings.Join(families, ",")
		desc.mask |= PANGO_FONT_MASK_FAMILY
	}

	return desc
}

func (desc *FontDescription) find_field_any(str string) bool {
	if field_matches("Normal", str) {
		return true
	}
	// try each of the possible field
	if v, ok := weight_map.fromString(str); ok {
		desc.pango_font_description_set_weight(Weight(v))
		return true
	}
	if v, ok := style_map.fromString(str); ok {
		desc.pango_font_description_set_style(Style(v))
		return true
	}
	if v, ok := stretch_map.fromString(str); ok {
		desc.pango_font_description_set_stretch(Stretch(v))
		return true
	}
	if v, ok := variant_map.fromString(str); ok {
		desc.pango_font_description_set_variant(Variant(v))
		return true
	}
	if v, ok := gravity_map.fromString(str); ok {
		desc.pango_font_description_set_gravity(Gravity(v))
		return true
	}
	return false
}

// pango_parse_style parses a font style. The allowed values are "normal",
// "italic" and "oblique", case variations being
// ignored.
func pango_parse_style(str string) (Style, bool) {
	i, b := style_map.parse_field(str)
	return Style(i), b
}

// pango_parse_variant parses a font variant. The allowed values are "normal"
// and "smallcaps" or "small_caps", case variations being
// ignored.
func pango_parse_variant(str string) (Variant, bool) {
	i, b := variant_map.parse_field(str)
	return Variant(i), b
}

// pango_parse_weight parses a font weight. The allowed values are "heavy",
// "ultrabold", "bold", "normal", "light", "ultraleight"
// and integers. Case variations are ignored.
func pango_parse_weight(str string) (Weight, bool) {
	i, b := weight_map.parse_field(str)
	return Weight(i), b
}

// pango_parse_stretch parses a font stretch. The allowed values are
// "ultra_condensed", "extra_condensed", "condensed",
// "semi_condensed", "normal", "semi_expanded", "expanded",
// "extra_expanded" and "ultra_expanded". Case variations are
// ignored and the '_' characters may be omitted.
func pango_parse_stretch(str string) (Stretch, bool) {
	i, b := stretch_map.parse_field(str)
	return Stretch(i), b
}

var hyphenStripper = strings.NewReplacer("-", "")

// TODO: check this is correct
func field_matches(s1, s2 string) bool {
	return hyphenStripper.Replace(strings.ToLower(s1)) == hyphenStripper.Replace(strings.ToLower(s2))
}

// Creates a string representation of a font description.
// The family list in the string description will only have a terminating comma if the
// last word of the list is a valid style option.
func (desc FontDescription) String() string {
	var chunks []string
	if desc.family_name != "" && (desc.mask&PANGO_FONT_MASK_FAMILY != 0) {
		fam := desc.family_name

		/* We need to add a trailing comma if the family name ends
		* in a keyword like "Bold", or if the family name ends in
		* a number and no keywords will be added.
		 */
		// TODO:
		// strings.Split(desc.family_name, ",")
		//    p = getword (desc.family_name, desc.family_name + strlen(desc.family_name), &wordlen, ",");
		if desc.weight == PANGO_WEIGHT_NORMAL &&
			desc.style == PANGO_STYLE_NORMAL &&
			desc.stretch == PANGO_STRETCH_NORMAL &&
			desc.variant == PANGO_VARIANT_NORMAL &&
			(desc.mask&(PANGO_FONT_MASK_GRAVITY|PANGO_FONT_MASK_SIZE) == 0) {
			fam += ","
		}
		chunks = append(chunks, fam)
	}

	if s := weight_map.toString("weight", int(desc.weight)); s != "" {
		chunks = append(chunks, s)
	}
	if s := style_map.toString("style", int(desc.style)); s != "" {
		chunks = append(chunks, s)
	}
	if s := stretch_map.toString("stretch", int(desc.stretch)); s != "" {
		chunks = append(chunks, s)
	}
	if s := variant_map.toString("variant", int(desc.variant)); s != "" {
		chunks = append(chunks, s)
	}
	if desc.mask&PANGO_FONT_MASK_GRAVITY != 0 {
		if s := gravity_map.toString("gravity", int(desc.gravity)); s != "" {
			chunks = append(chunks)
		}
	}

	if len(chunks) == 0 {
		chunks = append(chunks, "Normal")
	}

	if desc.mask&PANGO_FONT_MASK_SIZE != 0 {
		size := fmt.Sprintf("%g", float64(desc.size)/pangoScale)

		if desc.size_is_absolute {
			size += "px"
		}
		chunks = append(chunks, size)
	}

	if desc.variations != "" && desc.mask&PANGO_FONT_MASK_VARIATIONS != 0 {
		v := "@" + desc.variations
		chunks = append(chunks, v)
	}

	return strings.Join(chunks, " ")
}

// pango_font_description_equal compares two font descriptions for equality.
// Two font descriptions are considered equal if the fonts they describe are provably identical.
// This means that their masks do not have to match, as long as other fields
// are all the same.
// Note that two font descriptions may result in identical fonts
// being loaded, but still compare `false`.
func (desc1 FontDescription) pango_font_description_equal(desc2 FontDescription) bool {
	return desc1.style == desc2.style &&
		desc1.variant == desc2.variant &&
		desc1.weight == desc2.weight &&
		desc1.stretch == desc2.stretch &&
		desc1.size == desc2.size &&
		desc1.size_is_absolute == desc2.size_is_absolute &&
		desc1.gravity == desc2.gravity &&
		(desc1.family_name == desc2.family_name || strings.EqualFold(desc1.family_name, desc2.family_name)) &&
		desc1.variations == desc2.variations
}

// Sets the style field of a FontDescription. The
// Style enumeration describes whether the font is slanted and
// the manner in which it is slanted; it can be either
// `STYLE_NORMAL`, `STYLE_ITALIC`, or `STYLE_OBLIQUE`.
// Most fonts will either have a italic style or an oblique
// style, but not both, and font matching in Pango will
// match italic specifications with oblique fonts and vice-versa
// if an exact match is not found.
func (desc *FontDescription) pango_font_description_set_style(style Style) {
	if desc == nil {
		return
	}

	desc.style = style
	desc.mask |= PANGO_FONT_MASK_STYLE
}

// Sets the size field of a font description in fractional points.
// `size` is the size of the font in points, scaled by pangoScale.
// That is, a `size` value of 10 * pangoScale is a 10 point font. The conversion
// factor between points and device units depends on system configuration
// and the output device. For screen display, a logical DPI of 96 is
// common, in which case a 10 point font corresponds to a 10 * (96 / 72) = 13.3
// pixel font.
//
// This is mutually exclusive with pango_font_description_set_absolute_size(),
// to use if you need a particular size in device units
func (desc *FontDescription) pango_font_description_set_size(size int) {
	if desc == nil || size < 0 {
		return
	}

	desc.size = size
	desc.size_is_absolute = false
	desc.mask |= PANGO_FONT_MASK_SIZE
}

// Sets the size field of a font description, in device units.
// `size` is the new size, in Pango units. There are `pangoScale` Pango units in one
// device unit. For an output backend where a device unit is a pixel, a `size`
// value of 10 * pangoScale gives a 10 pixel font.
//
// This is mutually exclusive with pango_font_description_set_size() which sets the font size
// in points.
func (desc *FontDescription) pango_font_description_set_absolute_size(size int) {
	if desc == nil || size < 0 {
		return
	}

	desc.size = size
	desc.size_is_absolute = true
	desc.mask |= PANGO_FONT_MASK_SIZE
}

// Sets the stretch field of a font description. The stretch field
// specifies how narrow or wide the font should be.
func (desc *FontDescription) pango_font_description_set_stretch(stretch Stretch) {
	if desc == nil {
		return
	}
	desc.stretch = stretch
	desc.mask |= PANGO_FONT_MASK_STRETCH
}

// Sets the weight field of a font description. The weight field
// specifies how bold or light the font should be. In addition
// to the values of the Weight enumeration, other intermediate
// numeric values are possible.
func (desc *FontDescription) pango_font_description_set_weight(weight Weight) {
	if desc == nil {
		return
	}

	desc.weight = weight
	desc.mask |= PANGO_FONT_MASK_WEIGHT
}

// Sets the variant field of a font description. The variant
// can either be `VARIANT_NORMAL` or `VARIANT_SMALL_CAPS`.
func (desc *FontDescription) pango_font_description_set_variant(variant Variant) {
	if desc == nil {
		return
	}

	desc.variant = variant
	desc.mask |= PANGO_FONT_MASK_VARIANT
}

// pango_font_description_set_family sets the family name field of a font description. The family
// name represents a family of related font styles, and will
// resolve to a particular `FontFamily`. In some uses of
// `FontDescription`, it is also possible to use a comma
// separated list of family names for this field.
func (desc *FontDescription) pango_font_description_set_family(family string) {
	if desc == nil || desc.family_name == family {
		return
	}

	if family != "" {
		desc.family_name = family
		desc.mask |= PANGO_FONT_MASK_FAMILY
	} else {
		desc.family_name = pfd_defaults.family_name
		desc.mask &= ^PANGO_FONT_MASK_FAMILY
	}
}

// pango_font_description_set_gravity sets the gravity field of a font description. The gravity field
// specifies how the glyphs should be rotated.  If @gravity is
// %PANGO_GRAVITY_AUTO, this actually unsets the gravity mask on
// the font description.
func (desc *FontDescription) pango_font_description_set_gravity(gravity Gravity) {
	if desc == nil {
		return
	}

	if gravity == PANGO_GRAVITY_AUTO {
		desc.pango_font_description_unset_fields(PANGO_FONT_MASK_GRAVITY)
		return
	}

	desc.gravity = gravity
	desc.mask |= PANGO_FONT_MASK_GRAVITY
}

// Sets the variations field of a font description. OpenType
// font variations allow to select a font instance by specifying
// values for a number of axes, such as width or weight.
//
// The format of the variations string is AXIS1=VALUE,AXIS2=VALUE...,
// with each AXIS a 4 character tag that identifies a font axis,
// and each VALUE a floating point number. Unknown axes are ignored,
// and values are clamped to their allowed range.
//
// Pango does not currently have a way to find supported axes of
// a font. Both harfbuzz or freetype have API for this.
func (desc *FontDescription) pango_font_description_set_variations(variations string) {
	if desc == nil {
		return
	}
	if variations != "" {
		desc.variations = variations
		desc.mask |= PANGO_FONT_MASK_VARIATIONS
	} else {
		desc.variations = pfd_defaults.variations
		desc.mask &= ^PANGO_FONT_MASK_VARIATIONS
	}
}

// Merges the fields that are set in `descToMerge` into the fields in
// `desc`.  If `replaceExisting `is `false`, only fields in `desc` that
// are not already set are affected. If `true`, then fields that are
// already set will be replaced as well.
//
// If `descToMerge` is nil, this function performs nothing.
func (desc *FontDescription) pango_font_description_merge(descToMerge *FontDescription, replaceExisting bool) {
	if desc == nil || descToMerge == nil {
		return
	}
	var newMask FontMask
	if replaceExisting {
		newMask = descToMerge.mask
	} else {
		newMask = descToMerge.mask & ^desc.mask
	}
	if newMask&PANGO_FONT_MASK_FAMILY != 0 {
		desc.pango_font_description_set_family(descToMerge.family_name)
	}
	if newMask&PANGO_FONT_MASK_STYLE != 0 {
		desc.style = descToMerge.style
	}
	if newMask&PANGO_FONT_MASK_VARIANT != 0 {
		desc.variant = descToMerge.variant
	}
	if newMask&PANGO_FONT_MASK_WEIGHT != 0 {
		desc.weight = descToMerge.weight
	}
	if newMask&PANGO_FONT_MASK_STRETCH != 0 {
		desc.stretch = descToMerge.stretch
	}
	if newMask&PANGO_FONT_MASK_SIZE != 0 {
		desc.size = descToMerge.size
		desc.size_is_absolute = descToMerge.size_is_absolute
	}
	if newMask&PANGO_FONT_MASK_GRAVITY != 0 {
		desc.gravity = descToMerge.gravity
	}
	if newMask&PANGO_FONT_MASK_VARIATIONS != 0 {
		desc.pango_font_description_set_variations(descToMerge.variations)
	}
	desc.mask |= newMask
}

// Unsets some of the fields in a `FontDescription`.  The unset
// fields will get back to their default values.
func (desc *FontDescription) pango_font_description_unset_fields(toUnset FontMask) {
	if desc == nil {
		return
	}

	unsetDesc := pfd_defaults
	unsetDesc.mask = toUnset

	desc.pango_font_description_merge(&unsetDesc, true)

	desc.mask &= ^toUnset
}

// pango_font_description_hash returns a FontDescription suitable
// to be used as map key. In particular, the family_name is lowered, and `mask`
// is ignored.
func (desc FontDescription) pango_font_description_hash() FontDescription {
	desc.family_name = strings.ToLower(desc.family_name)
	desc.mask = 0
	return desc
}

// FontMetrics holds the overall metric information
// for a font (possibly restricted to a script).
// All values are expressed in Pango units.
type FontMetrics struct {
	// Distance from the baseline to the logical top of a line of text.
	// (The logical top may be above or below the top of the
	// actual drawn ink. It is necessary to lay out the text to figure
	// where the ink will be.)
	ascent int

	// Distance from the baseline to the logical bottom of a line of text.
	// (The logical bottom may be above or below the bottom of the
	// actual drawn ink. It is necessary to lay out the text to figure
	// where the ink will be.)
	descent int

	// Distance between successive baselines in wrapped text.
	height int

	// Representative value useful for example for
	// determining the initial size for a window. Actual characters in
	// text will be wider and narrower than this.
	approximate_char_width int

	// Same as `approximate_char_width` but for digits.
	// This value is generally somewhat more accurate than `approximate_char_width` for digits.
	approximate_digit_width int

	// Distance above the baseline of the top of the underline.
	// Since most fonts have underline positions beneath the baseline, this value is typically negative.
	underline_position int

	// Suggested thickness to draw for the underline.
	underline_thickness int

	// Distance above the baseline of the top of the strikethrough.
	strikethrough_position int
	// Suggested thickness to draw for the strikethrough.
	strikethrough_thickness int
}

// pango_font_get_metrics gets overall metric information for a font.
// Since the metrics may be substantially different for different scripts, a language tag can
// be provided to indicate that the metrics should be retrieved that
// correspond to the script(s) used by that language.
//
// If `font` is `nil`, this function gracefully sets some sane values in the
// output variables and returns.
func pango_font_get_metrics(font Font, language Language) FontMetrics {
	if font == nil {
		var metrics FontMetrics

		metrics.ascent = pangoScale * PANGO_UNKNOWN_GLYPH_HEIGHT
		metrics.descent = 0
		metrics.height = 0
		metrics.approximate_char_width = pangoScale * PANGO_UNKNOWN_GLYPH_WIDTH
		metrics.approximate_digit_width = pangoScale * PANGO_UNKNOWN_GLYPH_WIDTH
		metrics.underline_position = -pangoScale
		metrics.underline_thickness = pangoScale
		metrics.strikethrough_position = pangoScale * PANGO_UNKNOWN_GLYPH_HEIGHT / 2
		metrics.strikethrough_thickness = pangoScale

		return metrics
	}

	return font.get_metrics(language)
}

func (metrics *FontMetrics) update_metrics_from_items(language Language, text []rune, items []*Item) {
	// This should typically be called with a sample text string.
	if len(text) == 0 {
		return
	}

	fontsSeen := map[Font]bool{}
	metrics.approximate_char_width = 0
	var glyphs GlyphString

	for _, item := range items {
		font := item.analysis.font

		if seen := fontsSeen[font]; font != nil && !seen {
			rawMetrics := pango_font_get_metrics(font, language)
			fontsSeen[font] = true

			// metrics will already be initialized from the first font in the fontset
			metrics.ascent = max(metrics.ascent, rawMetrics.ascent)
			metrics.descent = max(metrics.descent, rawMetrics.descent)
			metrics.height = max(metrics.height, rawMetrics.height)
		}

		glyphs.pango_shape_full(text[item.offset:item.offset+item.num_chars], text, &item.analysis)
		metrics.approximate_char_width += int(glyphs.pango_glyph_string_get_width())
	}

	textWidth := pangoStrWidth(text)
	metrics.approximate_char_width /= textWidth
}

func pangoStrWidth(p []rune) int {
	var out int

	for _, c := range p {
		if isZeroWidth(c) {
			// + 0
		} else if isWide(c) {
			out += 2
		} else {
			out += 1
		}
	}

	return out
}

// isZeroWidth determines if a given character typically takes zero width when rendered.
// The return value is `true` for all non-spacing and enclosing marks
// (e.g., combining accents), format characters, zero-width
// space, but not U+00AD SOFT HYPHEN.
func isZeroWidth(c rune) bool {
	if c == 0x00AD {
		return false
	}

	if unicode.In(c, unicode.Mn, unicode.Me, unicode.Cf) {
		return true
	}

	if (c >= 0x1160 && c < 0x1200) || c == 0x200B {
		return true
	}

	return false
}

// FontFamily is used to represent a family of related
// font faces. The faces in a family share a common design, but differ in
// slant, weight, width and other aspects.
type FontFamily interface {
	list_faces() []FontFace
	get_name() string
	is_monospace() bool
	is_variable() bool
	get_face(name string) FontFace
}

// FontFace is used to represent a group of fonts with
// the same family, slant, weight, width, but varying sizes.
type FontFace interface {
	get_face_name() string
	describe() FontDescription
	list_sizes() []int
	is_synthesized() bool
	get_family() FontFamily
}
