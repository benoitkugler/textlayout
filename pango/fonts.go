package pango

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"
)

// contains acceptable strings value
type enumMap []struct {
	value int
	str   string
}

func (e enumMap) FromString(str string) (int, bool) {
	for _, v := range e {
		if field_matches(v.str, str) {
			return v.value, true
		}
	}
	return 0, false
}

// if v is not found, it is printed as "what=v"
func (e enumMap) ToString(what string, v int) string {
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

	v, found := map_.FromString(str)

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
	STYLE_NORMAL  Style = iota //  the font is upright.
	STYLE_OBLIQUE              //  the font is slanted, but in a roman style.
	STYLE_ITALIC               //  the font is slanted in an italic style.
)

var style_map = enumMap{
	{value: int(STYLE_NORMAL), str: ""},
	{value: int(STYLE_NORMAL), str: "Roman"},
	{value: int(STYLE_OBLIQUE), str: "Oblique"},
	{value: int(STYLE_ITALIC), str: "Italic"},
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
	STRETCH_ULTRA_CONDENSED Stretch = iota // ultra condensed width
	STRETCH_EXTRA_CONDENSED                // extra condensed width
	STRETCH_CONDENSED                      // condensed width
	STRETCH_SEMI_CONDENSED                 // semi condensed width
	STRETCH_NORMAL                         // the normal width
	STRETCH_SEMI_EXPANDED                  // semi expanded width
	STRETCH_EXPANDED                       // expanded width
	STRETCH_EXTRA_EXPANDED                 // extra expanded width
	STRETCH_ULTRA_EXPANDED                 // ultra expanded width
)

var stretch_map = enumMap{
	{value: int(STRETCH_ULTRA_CONDENSED), str: "Ultra-Condensed"},
	{value: int(STRETCH_EXTRA_CONDENSED), str: "Extra-Condensed"},
	{value: int(STRETCH_CONDENSED), str: "Condensed"},
	{value: int(STRETCH_SEMI_CONDENSED), str: "Semi-Condensed"},
	{value: int(STRETCH_NORMAL), str: ""},
	{value: int(STRETCH_SEMI_EXPANDED), str: "Semi-Expanded"},
	{value: int(STRETCH_EXPANDED), str: "Expanded"},
	{value: int(STRETCH_EXTRA_EXPANDED), str: "Extra-Expanded"},
	{value: int(STRETCH_ULTRA_EXPANDED), str: "Ultra-Expanded"},
}

// FontMask bits correspond to fields in a `FontDescription` that have been set.
type FontMask int16

const (
	F_FAMILY     FontMask = 1 << iota // the font family is specified.
	F_STYLE                           // the font style is specified.
	F_VARIANT                         // the font variant is specified.
	F_WEIGHT                          // the font weight is specified.
	F_STRETCH                         // the font stretch is specified.
	F_SIZE                            // the font size is specified.
	F_GRAVITY                         // the font gravity is specified (Since: 1.16.)
	F_VARIATIONS                      // OpenType font variations are specified (Since: 1.42)
)

/* CSS scale factors (1.2 factor between each size) */
const (
	PangoScale_XX_SMALL = 0.5787037037037 //  The scale factor for three shrinking steps (1 / (1.2 * 1.2 * 1.2)).
	PangoScale_X_SMALL  = 0.6944444444444 //  The scale factor for two shrinking steps (1 / (1.2 * 1.2)).
	PangoScale_SMALL    = 0.8333333333333 //  The scale factor for one shrinking step (1 / 1.2).
	PangoScale_MEDIUM   = 1.0             //  The scale factor for normal size (1.0).
	PangoScale_LARGE    = 1.2             //  The scale factor for one magnification step (1.2).
	PangoScale_X_LARGE  = 1.44            //  The scale factor for two magnification steps (1.2 * 1.2).
	PangoScale_XX_LARGE = 1.728           //  The scale factor for three magnification steps (1.2 * 1.2 * 1.2).
)

var pfd_defaults = FontDescription{
	FamilyName: "",

	Style:      STYLE_NORMAL,
	Variant:    PANGO_VARIANT_NORMAL,
	Weight:     PANGO_WEIGHT_NORMAL,
	Stretch:    STRETCH_NORMAL,
	Gravity:    PANGO_GRAVITY_SOUTH,
	Variations: "",

	mask:           0,
	SizeIsAbsolute: false,

	Size: 0,
}

// Font is used to represent a font in a rendering-system-independent matter.
// The concretes types implementing this interface MUST be valid map keys.
type Font interface {
	// Describe returns a description of the font.
	// The font size set in points, unless `absolute` is true,
	// meaning the font size is in device units.
	Describe(absolute bool) FontDescription

	// GetCoverage computes the coverage map for a given font and language tag.
	GetCoverage(language Language) Coverage

	// GetGlyphExtents gets the logical and ink extents of a glyph within a font. The
	// coordinate system for each rectangle has its origin at the
	// base line and horizontal origin of the character with increasing
	// coordinates extending to the right and down. The units
	// of the rectangles are in 1/PANGO_SCALE of a device unit.
	GetGlyphExtents(glyph Glyph, inkRect, logicalRect *Rectangle)

	// GetMetrics gets overall metric information for a font. Since the metrics may be
	// substantially different for different scripts, a language tag can
	// be provided to indicate that the metrics should
	// correspond to the script(s) used by that language.
	GetMetrics(language Language) FontMetrics

	// Gets the font map for which the font was created.
	GetFontMap() FontMap

	// GetFeatures obtains the OpenType features that are provided by the font.
	// These are passed to the rendering system, together with features
	// that have been explicitly set via attributes.
	//
	// Note that this does not include OpenType features which the
	// rendering system enables by default.
	// GetFeatures() []hb_feature_t

	// GetHBFont returns a hb_font_t object backing this font.
	// Implementations should create the font once and cache it.
	GetHBFont() *Hb_font_t
}

// pango_font_has_char returns whether the font provides a glyph for this character.
// `font` must not be nil
func pango_font_has_char(font Font, wc rune) bool {
	coverage := font.GetCoverage(pango_language_get_default())
	return coverage.Get(wc)
}

// FontDescription represents the description
// of an ideal font. These structures are used both to list
// what fonts are available on the system and also for specifying
// the characteristics of a font to load.
//
// This struct track the modifications to its field via a bit mask. Thus,
// the SetXXX methods should be used to mutate it.
//
// This struct does not hold any pointer types: it can be copied by value.
type FontDescription struct {
	FamilyName string

	Style   Style
	Variant Variant
	Weight  Weight
	Stretch Stretch
	Gravity Gravity

	Variations string

	mask           FontMask
	SizeIsAbsolute bool // = : 1;

	Size int
}

// NewFontDescription creates a new font description structure with all fields unset,
// but with default values.
func NewFontDescription() FontDescription {
	return pfd_defaults // copy
}

// pango_font_description_from_string creates a new font description from a string representation in the
// form
//
// "[FAMILY-LIST] [STYLE-OPTIONS] [SIZE] [VARIATIONS]",
//
// where FAMILY-LIST is a comma-separated list of families optionally
// terminated by a comma, STYLE_OPTIONS is a whitespace-separated list
// of words where each word Describes one of style, variant, weight,
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
	desc := NewFontDescription()

	desc.mask = F_STYLE |
		F_WEIGHT |
		F_VARIANT |
		F_STRETCH

	fields := strings.Fields(str)
	if len(fields) == 0 {
		return desc
	}

	// Look for variations at the end of the string */
	if word := fields[len(fields)-1]; word[0] == '@' {
		/* XXX: actually validate here */
		desc.Variations = word[1:]
		desc.mask |= F_VARIATIONS
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
			desc.Size = int(size*PangoScale + 0.5)
			desc.SizeIsAbsolute = size_is_absolute
			desc.mask |= F_SIZE
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
		desc.FamilyName = strings.Join(families, ",")
		desc.mask |= F_FAMILY
	}

	return desc
}

func (desc *FontDescription) find_field_any(str string) bool {
	if field_matches("Normal", str) {
		return true
	}
	// try each of the possible field
	if v, ok := weight_map.FromString(str); ok {
		desc.Setweight(Weight(v))
		return true
	}
	if v, ok := style_map.FromString(str); ok {
		desc.Setstyle(Style(v))
		return true
	}
	if v, ok := stretch_map.FromString(str); ok {
		desc.Setstretch(Stretch(v))
		return true
	}
	if v, ok := variant_map.FromString(str); ok {
		desc.Setvariant(Variant(v))
		return true
	}
	if v, ok := GravityMap.FromString(str); ok {
		desc.Setgravity(Gravity(v))
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
	if desc.FamilyName != "" && (desc.mask&F_FAMILY != 0) {
		fam := desc.FamilyName

		/* We need to add a trailing comma if the family name ends
		* in a keyword like "Bold", or if the family name ends in
		* a number and no keywords will be added.
		 */
		// TODO:
		// strings.Split(desc.family_name, ",")
		//    p = getword (desc.family_name, desc.family_name + strlen(desc.family_name), &wordlen, ",");
		if desc.Weight == PANGO_WEIGHT_NORMAL &&
			desc.Style == STYLE_NORMAL &&
			desc.Stretch == STRETCH_NORMAL &&
			desc.Variant == PANGO_VARIANT_NORMAL &&
			(desc.mask&(F_GRAVITY|F_SIZE) == 0) {
			fam += ","
		}
		chunks = append(chunks, fam)
	}

	if s := weight_map.ToString("weight", int(desc.Weight)); s != "" {
		chunks = append(chunks, s)
	}
	if s := style_map.ToString("style", int(desc.Style)); s != "" {
		chunks = append(chunks, s)
	}
	if s := stretch_map.ToString("stretch", int(desc.Stretch)); s != "" {
		chunks = append(chunks, s)
	}
	if s := variant_map.ToString("variant", int(desc.Variant)); s != "" {
		chunks = append(chunks, s)
	}
	if desc.mask&F_GRAVITY != 0 {
		if s := GravityMap.ToString("gravity", int(desc.Gravity)); s != "" {
			chunks = append(chunks)
		}
	}

	if len(chunks) == 0 {
		chunks = append(chunks, "Normal")
	}

	if desc.mask&F_SIZE != 0 {
		size := fmt.Sprintf("%g", float64(desc.Size)/PangoScale)

		if desc.SizeIsAbsolute {
			size += "px"
		}
		chunks = append(chunks, size)
	}

	if desc.Variations != "" && desc.mask&F_VARIATIONS != 0 {
		v := "@" + desc.Variations
		chunks = append(chunks, v)
	}

	return strings.Join(chunks, " ")
}

// pango_font_description_equal compares two font descriptions for equality.
// Two font descriptions are considered equal if the fonts they Describe are provably identical.
// This means that their masks do not have to match, as long as other fields
// are all the same.
// Note that two font descriptions may result in identical fonts
// being loaded, but still compare `false`.
func (desc1 FontDescription) pango_font_description_equal(desc2 FontDescription) bool {
	return desc1.Style == desc2.Style &&
		desc1.Variant == desc2.Variant &&
		desc1.Weight == desc2.Weight &&
		desc1.Stretch == desc2.Stretch &&
		desc1.Size == desc2.Size &&
		desc1.SizeIsAbsolute == desc2.SizeIsAbsolute &&
		desc1.Gravity == desc2.Gravity &&
		(desc1.FamilyName == desc2.FamilyName || strings.EqualFold(desc1.FamilyName, desc2.FamilyName)) &&
		desc1.Variations == desc2.Variations
}

// Sets the style field of a FontDescription. The
// Style enumeration Describes whether the font is slanted and
// the manner in which it is slanted; it can be either
// `STYLE_NORMAL`, `STYLE_ITALIC`, or `STYLE_OBLIQUE`.
// Most fonts will either have a italic style or an oblique
// style, but not both, and font matching in Pango will
// match italic specifications with oblique fonts and vice-versa
// if an exact match is not found.
func (desc *FontDescription) Setstyle(style Style) {
	if desc == nil {
		return
	}

	desc.Style = style
	desc.mask |= F_STYLE
}

// Sets the size field of a font description in fractional points.
// `size` is the size of the font in points, scaled by PangoScale.
// That is, a `size` value of 10 * PangoScale is a 10 point font. The conversion
// factor between points and device units depends on system configuration
// and the output device. For screen display, a logical DPI of 96 is
// common, in which case a 10 point font corresponds to a 10 * (96 / 72) = 13.3
// pixel font.
//
// This is mutually exclusive with SetAbsoluteSize(),
// to use if you need a particular size in device units
func (desc *FontDescription) SetSize(size int) {
	if desc == nil || size < 0 {
		return
	}

	desc.Size = size
	desc.SizeIsAbsolute = false
	desc.mask |= F_SIZE
}

// Sets the size field of a font description, in device units.
// `size` is the new size, in Pango units. There are `PangoScale` Pango units in one
// device unit. For an output backend where a device unit is a pixel, a `size`
// value of 10 * PangoScale gives a 10 pixel font.
//
// This is mutually exclusive with SetSize() which sets the font size
// in points.
func (desc *FontDescription) SetAbsoluteSize(size int) {
	if desc == nil || size < 0 {
		return
	}

	desc.Size = size
	desc.SizeIsAbsolute = true
	desc.mask |= F_SIZE
}

// Sets the stretch field of a font description. The stretch field
// specifies how narrow or wide the font should be.
func (desc *FontDescription) Setstretch(stretch Stretch) {
	if desc == nil {
		return
	}
	desc.Stretch = stretch
	desc.mask |= F_STRETCH
}

// Sets the weight field of a font description. The weight field
// specifies how bold or light the font should be. In addition
// to the values of the Weight enumeration, other intermediate
// numeric values are possible.
func (desc *FontDescription) Setweight(weight Weight) {
	if desc == nil {
		return
	}

	desc.Weight = weight
	desc.mask |= F_WEIGHT
}

// Sets the variant field of a font description. The variant
// can either be `VARIANT_NORMAL` or `VARIANT_SMALL_CAPS`.
func (desc *FontDescription) Setvariant(variant Variant) {
	if desc == nil {
		return
	}

	desc.Variant = variant
	desc.mask |= F_VARIANT
}

// Setfamily sets the family name field of a font description. The family
// name represents a family of related font styles, and will
// resolve to a particular `FontFamily`. In some uses of
// `FontDescription`, it is also possible to use a comma
// separated list of family names for this field.
func (desc *FontDescription) Setfamily(family string) {
	if desc == nil || desc.FamilyName == family {
		return
	}

	if family != "" {
		desc.FamilyName = family
		desc.mask |= F_FAMILY
	} else {
		desc.FamilyName = pfd_defaults.FamilyName
		desc.mask &= ^F_FAMILY
	}
}

// Setgravity sets the gravity field of a font description. The gravity field
// specifies how the glyphs should be rotated.  If @gravity is
// %PANGO_GRAVITY_AUTO, this actually unsets the gravity mask on
// the font description.
func (desc *FontDescription) Setgravity(gravity Gravity) {
	if desc == nil {
		return
	}

	if gravity == PANGO_GRAVITY_AUTO {
		desc.UnsetFields(F_GRAVITY)
		return
	}

	desc.Gravity = gravity
	desc.mask |= F_GRAVITY
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
func (desc *FontDescription) Setvariations(variations string) {
	if desc == nil {
		return
	}
	if variations != "" {
		desc.Variations = variations
		desc.mask |= F_VARIATIONS
	} else {
		desc.Variations = pfd_defaults.Variations
		desc.mask &= ^F_VARIATIONS
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
	if newMask&F_FAMILY != 0 {
		desc.Setfamily(descToMerge.FamilyName)
	}
	if newMask&F_STYLE != 0 {
		desc.Style = descToMerge.Style
	}
	if newMask&F_VARIANT != 0 {
		desc.Variant = descToMerge.Variant
	}
	if newMask&F_WEIGHT != 0 {
		desc.Weight = descToMerge.Weight
	}
	if newMask&F_STRETCH != 0 {
		desc.Stretch = descToMerge.Stretch
	}
	if newMask&F_SIZE != 0 {
		desc.Size = descToMerge.Size
		desc.SizeIsAbsolute = descToMerge.SizeIsAbsolute
	}
	if newMask&F_GRAVITY != 0 {
		desc.Gravity = descToMerge.Gravity
	}
	if newMask&F_VARIATIONS != 0 {
		desc.Setvariations(descToMerge.Variations)
	}
	desc.mask |= newMask
}

// Unsets some of the fields in a `FontDescription`.  The unset
// fields will get back to their default values.
func (desc *FontDescription) UnsetFields(toUnset FontMask) {
	if desc == nil {
		return
	}

	unsetDesc := pfd_defaults
	unsetDesc.mask = toUnset

	desc.pango_font_description_merge(&unsetDesc, true)

	desc.mask &= ^toUnset
}

// AsHash returns a FontDescription suitable
// to be used as map key. In particular, the family_name is lowered, and `mask`
// is ignored.
func (desc FontDescription) AsHash() FontDescription {
	desc.FamilyName = strings.ToLower(desc.FamilyName)
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
	Ascent int

	// Distance from the baseline to the logical bottom of a line of text.
	// (The logical bottom may be above or below the bottom of the
	// actual drawn ink. It is necessary to lay out the text to figure
	// where the ink will be.)
	Descent int

	// Distance between successive baselines in wrapped text.
	Height int

	// Representative value useful for example for
	// determining the initial size for a window. Actual characters in
	// text will be wider and narrower than this.
	ApproximateCharWidth int

	// Same as `approximate_char_width` but for digits.
	// This value is generally somewhat more accurate than `approximate_char_width` for digits.
	ApproximateDigitWidth int

	// Distance above the baseline of the top of the underline.
	// Since most fonts have underline positions beneath the baseline, this value is typically negative.
	UnderlinePosition int

	// Suggested thickness to draw for the underline.
	UnderlineThickness int

	// Distance above the baseline of the top of the strikethrough.
	StrikethroughPosition int
	// Suggested thickness to draw for the strikethrough.
	StrikethroughThickness int
}

// FontGetMetrics gets overall metric information for a font.
// Since the metrics may be substantially different for different scripts, a language tag can
// be provided to indicate that the metrics should be retrieved that
// correspond to the script(s) used by that language.
//
// If `font` is `nil`, this function gracefully returns some sane values.
func FontGetMetrics(font Font, language Language) FontMetrics {
	if font != nil {
		return font.GetMetrics(language)
	}
	var metrics FontMetrics

	metrics.Ascent = PangoScale * unknownGlyphHeight
	metrics.Descent = 0
	metrics.Height = 0
	metrics.ApproximateCharWidth = PangoScale * unknownGlyphWidth
	metrics.ApproximateDigitWidth = PangoScale * unknownGlyphWidth
	metrics.UnderlinePosition = -PangoScale
	metrics.UnderlineThickness = PangoScale
	metrics.StrikethroughPosition = PangoScale * unknownGlyphHeight / 2
	metrics.StrikethroughThickness = PangoScale

	return metrics
}

func (metrics *FontMetrics) update_metrics_from_items(language Language, text []rune, items []*Item) {
	// This should typically be called with a sample text string.
	if len(text) == 0 {
		return
	}

	fontsSeen := map[Font]bool{}
	metrics.ApproximateCharWidth = 0
	var glyphs GlyphString

	for _, item := range items {
		font := item.analysis.font

		if seen := fontsSeen[font]; font != nil && !seen {
			rawMetrics := FontGetMetrics(font, language)
			fontsSeen[font] = true

			// metrics will already be initialized from the first font in the Fontset
			metrics.Ascent = max(metrics.Ascent, rawMetrics.Ascent)
			metrics.Descent = max(metrics.Descent, rawMetrics.Descent)
			metrics.Height = max(metrics.Height, rawMetrics.Height)
		}

		glyphs.pango_shape_full(text[item.offset:item.offset+item.num_chars], text, &item.analysis)
		metrics.ApproximateCharWidth += int(glyphs.pango_glyph_string_get_width())
	}

	textWidth := pangoStrWidth(text)
	metrics.ApproximateCharWidth /= textWidth
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

// Note: the C implementation also uses Family and Faces,
// but it is not required by pango itself, so we simplify
// and do not use the notion

// // FontFamily is used to represent a family of related
// // font faces. The faces in a family share a common design, but differ in
// // slant, weight, width and other aspects.
// type FontFamily interface {
// 	// ListFaces lists the different font faces that make up the family. The faces
// 	// in a family share a common design, but differ in slant, weight,
// 	// width and other aspects.
// 	ListFaces() []FontFace

// 	// GetName gets the name of the family. The name is unique among all
// 	// fonts for the font backend and can be used in a FontDescription
// 	// to specify that a face from this family is desired.
// 	GetName() string

// 	// IsMonospace returns `true` if the family is monospace.
// 	// A monospace font is a font designed for text display where the
// 	// characters form a regular grid. For Western languages this would
// 	// mean that the advance width of all characters are the same, but
// 	// this categorization also includes Asian fonts which include
// 	// double-width characters: characters that occupy two grid cells.
// 	IsMonospace() bool

// 	// IsVariable returns `true` if the font has axes that can be modified to
// 	// produce different faces.
// 	IsVariable() bool

// 	// GetFace gets the FontFace of the family with the given name.
// 	// If `name` is empty, the family's default face (fontconfig calls it "Regular")
// 	// will be returned.
// 	// `nil` is returned if no face with the given name exists.
// 	GetFace(name string) FontFace
// }

// // FontFace is used to represent a group of fonts with
// // the same family, slant, weight, width, but varying sizes.
// type FontFace interface {
// 	// GetFaceName gets a name representing the style of this face among the
// 	// different faces in the FontFamily for the face. This
// 	// name is unique among all faces in the family and is suitable
// 	// for displaying to users.
// 	GetFaceName() string

// 	// Describe returns the family, style, variant, weight and stretch of
// 	// a FontFace. The size field of the resulting font description
// 	// will be unset.
// 	Describe() FontDescription

// 	// ListSizes lists the available sizes for a font. This is only applicable to bitmap
// 	// fonts. For scalable fonts, returns an empty array. The sizes returned
// 	// must be expressed in Pango units and sorted in ascending order.
// 	ListSizes() []int

// 	// IsSynthesized returns whether a FontFace is synthesized by the underlying
// 	// font rendering engine from another face, perhaps by shearing, emboldening,
// 	// or lightening it.
// 	IsSynthesized() bool

// 	// Gets the FontFamily that face belongs to.
// 	GetFamily() FontFamily
// }
