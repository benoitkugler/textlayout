package pango

import (
	"fmt"
)

const (
	ATTR_INVALID             AttrType = iota // does not happen
	ATTR_LANGUAGE                            // language (AttrLanguage)
	ATTR_FAMILY                              // font family name list (AttrString)
	ATTR_STYLE                               // font slant style (AttrInt)
	ATTR_WEIGHT                              // font weight (AttrInt)
	ATTR_VARIANT                             // font variant (normal or small caps) (AttrInt)
	ATTR_STRETCH                             // font stretch (AttrInt)
	ATTR_SIZE                                // font size in points scaled by %PANGO_SCALE (AttrInt)
	ATTR_FONT_DESC                           // font description (AttrFontDesc)
	ATTR_FOREGROUND                          // foreground color (AttrColor)
	ATTR_BACKGROUND                          // background color (AttrColor)
	ATTR_UNDERLINE                           // whether the text has an underline (AttrInt)
	ATTR_STRIKETHROUGH                       // whether the text is struck-through (AttrInt)
	ATTR_RISE                                // baseline displacement (AttrInt)
	ATTR_SHAPE                               // shape (AttrShape)
	ATTR_SCALE                               // font size scale factor (AttrFloat)
	ATTR_FALLBACK                            // whether fallback is enabled (AttrInt)
	ATTR_LETTER_SPACING                      // letter spacing (AttrInt)
	ATTR_UNDERLINE_COLOR                     // underline color (AttrColor)
	ATTR_STRIKETHROUGH_COLOR                 // strikethrough color (AttrColor)
	ATTR_ABSOLUTE_SIZE                       // font size in pixels scaled by %PANGO_SCALE (AttrInt)
	ATTR_GRAVITY                             // base text gravity (AttrInt)
	ATTR_GRAVITY_HINT                        // gravity hint (AttrInt)
	ATTR_FONT_FEATURES                       // OpenType font features (AttrString). Since 1.38
	ATTR_FOREGROUND_ALPHA                    // foreground alpha (AttrInt). Since 1.38
	ATTR_BACKGROUND_ALPHA                    // background alpha (AttrInt). Since 1.38
	ATTR_ALLOW_BREAKS                        // whether breaks are allowed (AttrInt). Since 1.44
	ATTR_SHOW                                // how to render invisible characters (AttrInt). Since 1.44
	ATTR_INSERT_HYPHENS                      // whether to insert hyphens at intra-word line breaks (AttrInt). Since 1.44
	ATTR_OVERLINE                            // whether the text has an overline (AttrInt). Since 1.46
	ATTR_OVERLINE_COLOR                      // overline color (AttrColor). Since 1.46
)

type AttrType uint8

var typeNames = [...]string{
	ATTR_INVALID:             "",
	ATTR_LANGUAGE:            "language",
	ATTR_FAMILY:              "family",
	ATTR_STYLE:               "style",
	ATTR_WEIGHT:              "weight",
	ATTR_VARIANT:             "variant",
	ATTR_STRETCH:             "stretch",
	ATTR_SIZE:                "size",
	ATTR_FONT_DESC:           "font-desc",
	ATTR_FOREGROUND:          "foreground",
	ATTR_BACKGROUND:          "background",
	ATTR_UNDERLINE:           "underline",
	ATTR_STRIKETHROUGH:       "strikethrough",
	ATTR_RISE:                "rise",
	ATTR_SHAPE:               "shape",
	ATTR_SCALE:               "scale",
	ATTR_FALLBACK:            "fallback",
	ATTR_LETTER_SPACING:      "letter-spacing",
	ATTR_UNDERLINE_COLOR:     "underline-color",
	ATTR_STRIKETHROUGH_COLOR: "strikethrough-color",
	ATTR_ABSOLUTE_SIZE:       "absolute-size",
	ATTR_GRAVITY:             "gravity",
	ATTR_GRAVITY_HINT:        "gravity-hint",
	ATTR_FONT_FEATURES:       "font-features",
	ATTR_FOREGROUND_ALPHA:    "foreground-alpha",
	ATTR_BACKGROUND_ALPHA:    "background-alpha",
	ATTR_ALLOW_BREAKS:        "allow-breaks",
	ATTR_SHOW:                "show",
	ATTR_INSERT_HYPHENS:      "insert-hyphens",
	ATTR_OVERLINE:            "overline",
	ATTR_OVERLINE_COLOR:      "overline-color",
}

func (t AttrType) String() string {
	if int(t) >= len(typeNames) {
		return "<invalid>"
	}
	return typeNames[t]
}

// ShowFlags affects how Pango treats characters that are normally
// not visible in the output.
type ShowFlags uint8

const (
	PANGO_SHOW_NONE        ShowFlags = 0         //  No special treatment for invisible characters
	PANGO_SHOW_SPACES      ShowFlags = 1 << iota //  Render spaces, tabs and newlines visibly
	PANGO_SHOW_LINE_BREAKS                       //  Render line breaks visibly
	PANGO_SHOW_IGNORABLES                        //  Render default-ignorable Unicode characters visibly
)

var showflags_map = enumMap{
	{value: int(PANGO_SHOW_NONE), str: ""},
	{value: int(PANGO_SHOW_SPACES), str: "spaces"},
	{value: int(PANGO_SHOW_LINE_BREAKS), str: "line-breaks"},
	{value: int(PANGO_SHOW_IGNORABLES), str: "ignorables"},
}

// Underline enumeration is used to specify
// whether text should be underlined, and if so, the type
// of underlining.
type Underline uint8

const (
	PANGO_UNDERLINE_NONE   Underline = iota // no underline should be drawn
	PANGO_UNDERLINE_SINGLE                  // a single underline should be drawn
	PANGO_UNDERLINE_DOUBLE                  // a double underline should be drawn
	// a single underline should be drawn at a
	// position beneath the ink extents of the text being
	// underlined. This should be used only for underlining
	// single characters, such as for keyboard accelerators.
	// PANGO_UNDERLINE_SINGLE should be used for extended
	// portions of text.
	PANGO_UNDERLINE_LOW
	// a wavy underline should be drawn below.
	// This underline is typically used to indicate an error such
	// as a possible mispelling; in some cases a contrasting color
	// may automatically be used.
	PANGO_UNDERLINE_ERROR
	PANGO_UNDERLINE_SINGLE_LINE // like PANGO_UNDERLINE_SINGLE, but drawn continuously across multiple runs.
	PANGO_UNDERLINE_DOUBLE_LINE // like PANGO_UNDERLINE_DOUBLE, but drawn continuously across multiple runs.
	PANGO_UNDERLINE_ERROR_LINE  // like PANGO_UNDERLINE_ERROR, but drawn continuously across multiple runs.
)

var underline_map = enumMap{
	{value: int(PANGO_UNDERLINE_NONE), str: "none"},
	{value: int(PANGO_UNDERLINE_SINGLE), str: "single"},
	{value: int(PANGO_UNDERLINE_DOUBLE), str: "double"},
	{value: int(PANGO_UNDERLINE_LOW), str: "low"},
	{value: int(PANGO_UNDERLINE_ERROR), str: "error"},
	{value: int(PANGO_UNDERLINE_SINGLE_LINE), str: "single-line"},
	{value: int(PANGO_UNDERLINE_DOUBLE_LINE), str: "double-line"},
	{value: int(PANGO_UNDERLINE_ERROR_LINE), str: "error-line"},
}

// Overline is used to specify
// whether text should be overlined, and if so, the type
// of line.
type Overline uint8

const (
	PANGO_OVERLINE_NONE   Overline = iota // no overline should be drawn
	PANGO_OVERLINE_SINGLE                 // Draw a single line above the ink extents of the text being underlined.
)

var overline_map = enumMap{
	{value: int(PANGO_OVERLINE_NONE), str: "none"},
	{value: int(PANGO_OVERLINE_SINGLE), str: "single"},
}

// AttrData stores the type specific value of
// an attribute.
type AttrData interface {
	fmt.Stringer
	copy() AttrData
	equals(other AttrData) bool
}

type AttrInt int

func (a AttrInt) copy() AttrData             { return a }
func (a AttrInt) String() string             { return fmt.Sprintf("%d", a) }
func (a AttrInt) equals(other AttrData) bool { return a == other }

type AttrFloat float64

func (a AttrFloat) copy() AttrData             { return a }
func (a AttrFloat) String() string             { return fmt.Sprintf("%f", a) }
func (a AttrFloat) equals(other AttrData) bool { return a == other }

type AttrString string

func (a AttrString) copy() AttrData             { return a }
func (a AttrString) String() string             { return string(a) }
func (a AttrString) equals(other AttrData) bool { return a == other }

func (a Language) copy() AttrData             { return a }
func (a Language) String() string             { return string(a) }
func (a Language) equals(other AttrData) bool { return a == other }

func (a FontDescription) copy() AttrData { return a }
func (a FontDescription) equals(other AttrData) bool {
	otherDesc, ok := other.(FontDescription)
	if !ok {
		return false
	}
	return a.mask == otherDesc.mask && a.pango_font_description_equal(otherDesc)
}

// AttrColor is used to represent a color in an uncalibrated RGB color-space.
type AttrColor struct {
	Red, Green, Blue uint16
}

func (a AttrColor) copy() AttrData             { return a }
func (a AttrColor) equals(other AttrData) bool { return a == other }

// String returns a textual specification of the color in the hexadecimal form
// '#rrrrggggbbbb', where r,g and b are hex digits representing
// the red, green, and blue components respectively.
func (a AttrColor) String() string { return fmt.Sprintf("#%04x%04x%04x", a.Red, a.Green, a.Blue) }

// AttrShape is used to represent attributes which impose shape restrictions.
type AttrShape struct {
	ink, logical Rectangle
}

func (a AttrShape) copy() AttrData             { return a }
func (a AttrShape) equals(other AttrData) bool { return a == other }
func (a AttrShape) String() string             { return "shape" }

func (shape AttrShape) _pango_shape_get_extents(n_chars int, inkRect, logicalRect *Rectangle) {
	if n_chars > 0 {
		if inkRect != nil {
			inkRect.x = min(shape.ink.x, shape.ink.x+shape.logical.width*(n_chars-1))
			inkRect.width = max(shape.ink.width, shape.ink.width+shape.logical.width*(n_chars-1))
			inkRect.y = shape.ink.y
			inkRect.height = shape.ink.height
		}
		if logicalRect != nil {
			logicalRect.x = min(shape.logical.x, shape.logical.x+shape.logical.width*(n_chars-1))
			logicalRect.width = max(shape.logical.width, shape.logical.width+shape.logical.width*(n_chars-1))
			logicalRect.y = shape.logical.y
			logicalRect.height = shape.logical.height
		}
	} else {
		if inkRect != nil {
			inkRect.x, inkRect.y, inkRect.width, inkRect.height = 0, 0, 0, 0
		}

		if logicalRect != nil {
			logicalRect.x, logicalRect.y, logicalRect.width, logicalRect.height = 0, 0, 0, 0
		}
	}
}

// Attribute are used as the input to the itemization process and also when
// creating a `Layout`. The common portion of the attribute holds
// the range to which the value applies.
// By default an attribute will have an all-inclusive range of [0,maxInt].
type Attribute struct {
	Type AttrType
	Data AttrData
	// Indexes into the underlying rune slice (note that
	// we diverge here from the C library, which works on byte slices).
	// The character at `EndIndex` is not included in the range.
	StartIndex, EndIndex int
}

// pango_attribute_init initializes StartIndex to 0 and EndIndex to maxInt
// such that the attribute applies to the entire text by default.
func (attr *Attribute) pango_attribute_init() {
	attr.StartIndex = 0
	attr.EndIndex = maxInt
}

// Compare two attributes for equality. This compares only the
// actual value of the two attributes and not the ranges that the
// attributes apply to.
func (attr1 Attribute) pango_attribute_equal(attr2 Attribute) bool {
	return attr1.Type == attr2.Type && attr1.Data.equals(attr2.Data)
}

// Make a deep copy of an attribute.
func (a *Attribute) pango_attribute_copy() *Attribute {
	if a == nil {
		return a
	}
	out := *a
	out.Data = a.Data.copy()
	return &out
}

func (attr Attribute) String() string {
	// to obtain the same result as the C implementation
	// we convert to int32
	return fmt.Sprintf("[%d,%d]%s=%s", int32(attr.StartIndex), int32(attr.EndIndex), attr.Type, attr.Data)
}

// Create a new font description attribute. This attribute
// allows setting family, style, weight, variant, stretch,
// and size simultaneously.
func pango_attr_font_desc_new(desc FontDescription) *Attribute {
	//    PangoAttrFontDesc *result = g_slice_new (PangoAttrFontDesc);
	//    pango_attribute_init (&result.attr, &klass);
	//    result.desc = pango_font_description_copy (desc);

	out := Attribute{Type: ATTR_FONT_DESC, Data: desc}
	out.pango_attribute_init()
	return &out
}

// Create a new attribute that influences how invisible
// characters are rendered.
func pango_attr_show_new(flags ShowFlags) *Attribute {
	out := Attribute{Type: ATTR_SHOW, Data: AttrInt(flags)}
	out.pango_attribute_init()
	return &out
}

// Create a new font size scale attribute. The base font for the
// affected text will have its size multiplied by `scaleFactor`.
func pango_attr_scale_new(scaleFactor float64) *Attribute {
	out := Attribute{Type: ATTR_SCALE, Data: AttrFloat(scaleFactor)}
	out.pango_attribute_init()
	return &out
}

// Create a new font-size attribute in fractional points.
func pango_attr_size_new(size int) *Attribute {
	out := Attribute{Type: ATTR_SIZE, Data: AttrInt(size)}
	out.pango_attribute_init()
	return &out
}

//  Create a new font weight attribute.
func pango_attr_weight_new(weight Weight) *Attribute {
	out := Attribute{Type: ATTR_WEIGHT, Data: AttrInt(weight)}
	out.pango_attribute_init()
	return &out
}

// Create a new font variant attribute (normal or small caps)
func pango_attr_variant_new(variant Variant) *Attribute {
	out := Attribute{Type: ATTR_VARIANT, Data: AttrInt(variant)}
	out.pango_attribute_init()
	return &out
}

// Create a new gravity attribute, which should not be `PANGO_GRAVITY_AUTO`.
func pango_attr_gravity_new(gravity Gravity) *Attribute {
	out := Attribute{Type: ATTR_GRAVITY, Data: AttrInt(gravity)}
	out.pango_attribute_init()
	return &out
}

// Create a new gravity_hint attribute.
func pango_attr_gravity_hint_new(gravity_hint GravityHint) *Attribute {
	out := Attribute{Type: ATTR_GRAVITY_HINT, Data: AttrInt(gravity_hint)}
	out.pango_attribute_init()
	return &out
}

// Create a new font slant style attribute.
func pango_attr_style_new(style Style) *Attribute {
	out := Attribute{Type: ATTR_STYLE, Data: AttrInt(style)}
	out.pango_attribute_init()
	return &out
}

// Create a new font-size attribute in device units.
func pango_attr_size_new_absolute(size int) *Attribute {
	out := Attribute{Type: ATTR_ABSOLUTE_SIZE, Data: AttrInt(size)}
	out.pango_attribute_init()
	return &out
}

// Create a new letter-spacing attribute, the amount of extra space to add between graphemes
// of the text, in Pango units.
func pango_attr_letter_spacing_new(letterSpacing int) *Attribute {
	out := Attribute{Type: ATTR_LETTER_SPACING, Data: AttrInt(letterSpacing)}
	out.pango_attribute_init()
	return &out
}

// Create a new font stretch attribute
func pango_attr_stretch_new(stretch Stretch) *Attribute {
	out := Attribute{Type: ATTR_STRETCH, Data: AttrInt(stretch)}
	out.pango_attribute_init()
	return &out
}

// Create a new baseline displacement attribute: `rise` is the amount
// that the text should be displaced vertically, in Pango units.
// Positive values displace the text upwards.
func pango_attr_rise_new(rise int) *Attribute {
	out := Attribute{Type: ATTR_RISE, Data: AttrInt(rise)}
	out.pango_attribute_init()
	return &out
}

// Create a new underline-style attribute.
func pango_attr_underline_new(underline Underline) *Attribute {
	out := Attribute{Type: ATTR_UNDERLINE, Data: AttrInt(underline)}
	out.pango_attribute_init()
	return &out
}

// Create a new foreground alpha attribute.
func pango_attr_foreground_alpha_new(alpha uint16) *Attribute {
	out := Attribute{Type: ATTR_FOREGROUND_ALPHA, Data: AttrInt(alpha)}
	out.pango_attribute_init()
	return &out
}

// Create a new background alpha attribute.
func pango_attr_background_alpha_new(alpha uint16) *Attribute {
	out := Attribute{Type: ATTR_BACKGROUND_ALPHA, Data: AttrInt(alpha)}
	out.pango_attribute_init()
	return &out
}

// Create a new overline-style attribute.
func pango_attr_overline_new(overline Overline) *Attribute {
	out := Attribute{Type: ATTR_OVERLINE, Data: AttrInt(overline)}
	out.pango_attribute_init()
	return &out
}

// Create a new strikethrough-style attribute.
func pango_attr_strikethrough_new(strikethrough bool) *Attribute {
	v := AttrInt(0)
	if strikethrough {
		v = 1
	}
	out := Attribute{Type: ATTR_STRIKETHROUGH, Data: v}
	out.pango_attribute_init()
	return &out
}

// Create a new allow-breaks attribute.
//
// If breaks are disabled, the range will be kept in a
// single run, as far as possible.
func pango_attr_allow_breaks_new(allowBreaks bool) *Attribute {
	v := AttrInt(0)
	if allowBreaks {
		v = 1
	}
	out := Attribute{Type: ATTR_ALLOW_BREAKS, Data: v}
	out.pango_attribute_init()
	return &out
}

// Create a new insert-hyphens attribute.
//
// Pango will insert hyphens when breaking lines in the middle
// of a word. This attribute can be used to suppress the hyphen.
func pango_attr_insert_hyphens_new(insertHyphens bool) *Attribute {
	v := AttrInt(0)
	if insertHyphens {
		v = 1
	}
	out := Attribute{Type: ATTR_INSERT_HYPHENS, Data: v}
	out.pango_attribute_init()
	return &out
}

// Create a new font fallback attribute.
//
// If fallback is disabled, characters will only be used from the
// closest matching font on the system. No fallback will be done to
// other fonts on the system that might contain the characters in the
// text.
func pango_attr_fallback_new(enableFallback bool) *Attribute {
	f := 0
	if enableFallback {
		f = 1
	}
	out := Attribute{Type: ATTR_FALLBACK, Data: AttrInt(f)}
	out.pango_attribute_init()
	return &out
}

// Create a new language tag attribute
func pango_attr_language_new(language Language) *Attribute {
	out := Attribute{Type: ATTR_LANGUAGE, Data: language}
	out.pango_attribute_init()
	return &out
}

// Create a new font family attribute: `family` is
// the family or comma separated list of families.
func pango_attr_family_new(family string) *Attribute {
	out := Attribute{Type: ATTR_FAMILY, Data: AttrString(family)}
	out.pango_attribute_init()
	return &out
}

// Create a new foreground color attribute.
func pango_attr_foreground_new(color AttrColor) *Attribute {
	out := Attribute{Type: ATTR_FOREGROUND, Data: color}
	out.pango_attribute_init()
	return &out
}

// Create a new background color attribute.
func pango_attr_background_new(color AttrColor) *Attribute {
	out := Attribute{Type: ATTR_BACKGROUND, Data: color}
	out.pango_attribute_init()
	return &out
}

// Create a new underline color attribute. This attribute
// modifies the color of underlines. If not set, underlines
// will use the foreground color.
func pango_attr_underline_color_new(color AttrColor) *Attribute {
	out := Attribute{Type: ATTR_UNDERLINE_COLOR, Data: color}
	out.pango_attribute_init()
	return &out
}

// Create a new overline color attribute. This attribute
// modifies the color of overlines. If not set, overlines
// will use the foreground color.
func pango_attr_overline_color_new(color AttrColor) *Attribute {
	out := Attribute{Type: ATTR_OVERLINE_COLOR, Data: color}
	out.pango_attribute_init()
	return &out
}

// Create a new strikethrough color attribute. This attribute
// modifies the color of strikethrough lines. If not set, strikethrough lines
// will use the foreground color.
func pango_attr_strikethrough_color_new(color AttrColor) *Attribute {
	out := Attribute{Type: ATTR_STRIKETHROUGH_COLOR, Data: color}
	out.pango_attribute_init()
	return &out
}

// Create a new shape attribute. A shape is used to impose a
// particular ink and logical rectangle on the result of shaping a
// particular glyph. This might be used, for instance, for
// embedding a picture or a widget inside a `Layout`.
func pango_attr_shape_new(ink, logical Rectangle) *Attribute {
	out := Attribute{Type: ATTR_SHAPE, Data: AttrShape{ink: ink, logical: logical}}
	out.pango_attribute_init()
	return &out
}

// Create a new font features tag attribute, from a string with OpenType font features, in CSS syntax
func pango_attr_font_features_new(features string) *Attribute {
	out := Attribute{Type: ATTR_FONT_FEATURES, Data: AttrString(features)}
	out.pango_attribute_init()
	return &out
}

type AttrList []*Attribute

// pango_attr_list_copy returns a deep copy of the list,
// calling `pango_attribute_copy` for each element.
func (list AttrList) pango_attr_list_copy() AttrList {
	out := make(AttrList, len(list))
	for i, v := range list {
		out[i] = v.pango_attribute_copy()
	}
	return out
}

func (list *AttrList) pango_attr_list_insert_internal(attr *Attribute, before bool) {
	startIndex := attr.StartIndex

	if len(*list) == 0 {
		*list = append(*list, attr)
		return
	}

	lastAttr := (*list)[len(*list)-1]

	if lastAttr.StartIndex < startIndex || (!before && lastAttr.StartIndex == startIndex) {
		*list = append(*list, attr)
	} else {
		for i, cur := range *list {
			if cur.StartIndex > startIndex || (before && cur.StartIndex == startIndex) {
				list.insert(i, attr)
				break
			}
		}
	}
}

func (l *AttrList) insert(i int, attr *Attribute) {
	*l = append(*l, nil)
	copy((*l)[i+1:], (*l)[i:])
	(*l)[i] = attr
}

func (l *AttrList) remove(i int) {
	copy((*l)[i:], (*l)[i+1:])
	(*l)[len((*l))-1] = nil
	(*l) = (*l)[:len((*l))-1]
}

// Insert the given attribute into the list. It will
// be inserted after all other attributes with a matching
// `StartIndex`.
func (list *AttrList) pango_attr_list_insert(attr *Attribute) {
	if list == nil {
		return
	}
	list.pango_attr_list_insert_internal(attr, false)
}

// Insert the given attribute into the `AttrList`. It will
// be inserted before all other attributes with a matching
// `StartIndex`.
func (list *AttrList) pango_attr_list_insert_before(attr *Attribute) {
	if list == nil {
		return
	}
	list.pango_attr_list_insert_internal(attr, true)
}

// Insert the given attribute into the `AttrList`. It will
// replace any attributes of the same type on that segment
// and be merged with any adjoining attributes that are identical.
//
// This function is slower than pango_attr_list_insert() for
// creating an attribute list in order (potentially much slower
// for large lists). However, pango_attr_list_insert() is not
// suitable for continually changing a set of attributes
// since it never removes or combines existing attributes.
func (list *AttrList) pango_attr_list_change(attr *Attribute) {
	if list == nil {
		return
	}

	startIndex := attr.StartIndex
	endIndex := attr.EndIndex

	if startIndex == endIndex {
		/* empty, nothing to do */
		return
	}

	if len(*list) == 0 {
		list.pango_attr_list_insert(attr)
		return
	}

	var i, p int
	inserted := false
	for i, p = 0, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]

		if tmp_attr.StartIndex > startIndex {
			list.insert(i, attr)
			inserted = true
			break
		}

		if tmp_attr.Type != attr.Type {
			continue
		}

		if tmp_attr.EndIndex < startIndex {
			continue /* This attr does not overlap with the new one */
		}

		// tmp_attr.StartIndex <= startIndex
		// tmp_attr.EndIndex >= startIndex

		if tmp_attr.pango_attribute_equal(*attr) { // We can merge the new attribute with this attribute
			if tmp_attr.EndIndex >= endIndex {
				// We are totally overlapping the previous attribute.
				// No action is needed.
				return
			}
			tmp_attr.EndIndex = endIndex
			attr = tmp_attr
			inserted = true
			break
		} else { // Split, truncate, or remove the old attribute
			if tmp_attr.EndIndex > endIndex {
				end_attr := tmp_attr.pango_attribute_copy()
				end_attr.StartIndex = endIndex
				list.pango_attr_list_insert(end_attr)
			}

			if tmp_attr.StartIndex == startIndex {
				list.remove(i)
				break
			} else {
				tmp_attr.EndIndex = startIndex
			}
		}
	}

	if !inserted { // we didn't insert attr yet
		list.pango_attr_list_insert(attr)
		return
	}

	/* We now have the range inserted into the list one way or the
	* other. Fix up the remainder */
	/* Attention: No i = 0 here. */
	for i, p = i+1, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]

		if tmp_attr.StartIndex > endIndex {
			break
		}

		if tmp_attr.Type != attr.Type {
			continue
		}

		if tmp_attr.EndIndex <= attr.EndIndex || tmp_attr.pango_attribute_equal(*attr) {
			/* We can merge the new attribute with this attribute. */
			attr.EndIndex = max(endIndex, tmp_attr.EndIndex)
			list.remove(i)
			i--
			p--
			continue
		} else {
			/* Trim the start of this attribute that it begins at the end
			* of the new attribute. This may involve moving
			* it in the list to maintain the required non-decreasing
			* order of start indices
			 */
			tmp_attr.StartIndex = attr.EndIndex
			// TODO: Is the following loop missing something ?
			// for k, m := i+1, len(*list); k < m; k++ {
			// 	tmp_attr2 := (*list)[k]
			// 	if tmp_attr2.StartIndex >= tmp_attr.StartIndex {
			// 		break
			// 	}
			// }
		}
	}
}

// Checks whether `list` and `otherList` contain the same attributes and
// whether those attributes apply to the same ranges.
// Beware that this will return wrong values if any list contains duplicates.
func (list AttrList) pango_attr_list_equal(otherList AttrList) bool {
	if len(list) != len(otherList) {
		return false
	}

	var skipBitmask uint64
	for _, attr := range list {
		attrEqual := false
		for otherAttrIndex, otherAttr := range otherList {
			var otherAttrBitmask uint64
			if otherAttrIndex < 64 {
				otherAttrBitmask = 1 << otherAttrIndex
			}

			if (skipBitmask & otherAttrBitmask) != 0 {
				continue
			}

			if attr.StartIndex == otherAttr.StartIndex &&
				attr.EndIndex == otherAttr.EndIndex && attr.pango_attribute_equal(*otherAttr) {
				skipBitmask |= otherAttrBitmask
				attrEqual = true
				break
			}
		}

		if !attrEqual {
			return false
		}
	}

	return true
}

// Return value: `true` if the attribute should be selected for
// filtering, `false` otherwise.
type pangoAttrFilterFunc = func(attr *Attribute) bool

// Given a AttrList and callback function, removes any elements
// of `list` for which `fn` returns `true` and inserts them into
// a new list (possibly empty if no attributes of the given types were found)
func (list *AttrList) pango_attr_list_filter(fn pangoAttrFilterFunc) AttrList {
	if list == nil {
		return nil
	}
	var out AttrList
	for i, p := 0, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]
		if fn(tmp_attr) {
			list.remove(i)
			i-- /* Need to look at this index again */
			p--
			out = append(out, tmp_attr)
		}
	}

	return out
}

// AttrIterator is used to represent an
// iterator through an `AttrList`. A new iterator is created
// with pango_attr_list_get_iterator(). Once the iterator
// is created, it can be advanced through the style changes
// in the text using pango_attr_iterator_next(). At each
// style change, the range of the current style segment and the
// attributes currently in effect can be queried.
type AttrIterator struct {
	attrs *AttrList

	attribute_stack AttrList

	attr_index           int
	StartIndex, EndIndex int // index into the underlying text
}

// pango_attr_list_get_iterator creates a iterator initialized to the beginning of the list.
// `list` must not be modified until this iterator is freed.
func (list *AttrList) pango_attr_list_get_iterator() *AttrIterator {
	if list == nil {
		return nil
	}
	iterator := AttrIterator{attrs: list}

	if !iterator.pango_attr_iterator_next() {
		iterator.EndIndex = maxInt
	}

	return &iterator
}

// pango_attr_iterator_copy returns a copy of `iterator`.
func (iterator *AttrIterator) pango_attr_iterator_copy() *AttrIterator {
	if iterator == nil {
		return nil
	}
	copy := *iterator
	// dont deep copy the attributes themselves
	copy.attribute_stack = append(AttrList{}, iterator.attribute_stack...)
	return &copy
}

// pango_attr_iterator_next advances the iterator until the next change of style, and
// returns `false` if the iterator is at the end of the list, otherwise `true`
func (iterator *AttrIterator) pango_attr_iterator_next() bool {
	if iterator == nil {
		return false
	}

	if iterator.attr_index >= len(*iterator.attrs) && len(iterator.attribute_stack) == 0 {
		return false
	}
	iterator.StartIndex = iterator.EndIndex
	iterator.EndIndex = maxInt

	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		if attr.EndIndex == iterator.StartIndex {
			iterator.attribute_stack.remove(i)
		} else {
			iterator.EndIndex = min(iterator.EndIndex, attr.EndIndex)
		}
	}

	for {
		if iterator.attr_index >= len(*iterator.attrs) {
			break
		}
		attr := (*iterator.attrs)[iterator.attr_index]

		if attr.StartIndex != iterator.StartIndex {
			break
		}

		if attr.EndIndex > iterator.StartIndex {
			iterator.attribute_stack = append(iterator.attribute_stack, attr)
			iterator.EndIndex = min(iterator.EndIndex, attr.EndIndex)
		}

		iterator.attr_index++ /* NEXT! */
	}

	if iterator.attr_index < len(*iterator.attrs) {
		attr := (*iterator.attrs)[iterator.attr_index]
		iterator.EndIndex = min(iterator.EndIndex, attr.StartIndex)
	}

	return true
}

// pango_attr_iterator_get_attrs gets a list of all attributes at the current position of the
// iterator.
func (iterator AttrIterator) pango_attr_iterator_get_attrs() AttrList {
	var attrs AttrList

	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		found := false
		for _, old_attr := range attrs {
			if attr.Type == old_attr.Type {
				found = true
				break
			}
		}
		if !found {
			attrs = append(AttrList{attr}, attrs...)
		}
	}

	return attrs
}

// pango_attr_iterator_get finds the current attribute of a particular type at the iterator
// location. When multiple attributes of the same type overlap,
// the attribute whose range starts closest to the current location is used.
// It returns `nil` if no attribute of that type applies to the current location.
func (iterator AttrIterator) pango_attr_iterator_get(type_ AttrType) *Attribute {
	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		if attr.Type == type_ {
			return attr
		}
	}
	return nil
}

// pango_attr_iterator_get_font gets the font and other attributes at the current iterator position.
// `desc` is a FontDescription to fill in with the current values.
// If non-nil, `language` is a location to store language tag for item, or zero if none is found.
// If non-nil, `extra_attrs` is a location in which to store a list of non-font
// attributes at the the current position; only the highest priority
// value of each attribute will be added to this list.
func (iterator AttrIterator) pango_attr_iterator_get_font(desc *FontDescription, language *Language, extra_attrs *AttrList) {
	if desc == nil {
		return
	}
	//    int i;

	if language != nil {
		*language = ""
	}

	if extra_attrs != nil {
		*extra_attrs = nil
	}

	if len(iterator.attribute_stack) == 0 {
		return
	}

	var (
		mask                    FontMask
		haveScale, haveLanguage bool
		scale                   AttrFloat
	)
	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]

		switch attr.Type {
		case ATTR_FONT_DESC:
			attrDesc := attr.Data.(FontDescription)
			new_mask := attrDesc.mask & ^mask
			mask |= new_mask
			desc.pango_font_description_unset_fields(new_mask)
			desc.pango_font_description_merge(&attrDesc, false)
		case ATTR_FAMILY:
			if mask&PANGO_FONT_MASK_FAMILY == 0 {
				mask |= PANGO_FONT_MASK_FAMILY
				desc.pango_font_description_set_family(string(attr.Data.(AttrString)))
			}
		case ATTR_STYLE:
			if mask&PANGO_FONT_MASK_STYLE == 0 {
				mask |= PANGO_FONT_MASK_STYLE
				desc.pango_font_description_set_style(Style(attr.Data.(AttrInt)))
			}
		case ATTR_VARIANT:
			if mask&PANGO_FONT_MASK_VARIANT == 0 {
				mask |= PANGO_FONT_MASK_VARIANT
				desc.pango_font_description_set_variant(Variant(attr.Data.(AttrInt)))
			}
		case ATTR_WEIGHT:
			if mask&PANGO_FONT_MASK_WEIGHT == 0 {
				mask |= PANGO_FONT_MASK_WEIGHT
				desc.pango_font_description_set_weight(Weight(attr.Data.(AttrInt)))
			}
		case ATTR_STRETCH:
			if mask&PANGO_FONT_MASK_STRETCH == 0 {
				mask |= PANGO_FONT_MASK_STRETCH
				desc.pango_font_description_set_stretch(Stretch(attr.Data.(AttrInt)))
			}
		case ATTR_SIZE:
			if mask&PANGO_FONT_MASK_SIZE == 0 {
				mask |= PANGO_FONT_MASK_SIZE
				desc.pango_font_description_set_size(int(attr.Data.(AttrInt)))
			}
		case ATTR_ABSOLUTE_SIZE:
			if mask&PANGO_FONT_MASK_SIZE == 0 {
				mask |= PANGO_FONT_MASK_SIZE
				desc.pango_font_description_set_absolute_size(int(attr.Data.(AttrInt)))
			}
		case ATTR_SCALE:
			if !haveScale {
				haveScale = true
				scale = attr.Data.(AttrFloat)
			}
		case ATTR_LANGUAGE:
			if language != nil {
				if !haveLanguage {
					haveLanguage = true
					*language = attr.Data.(Language)
				}
			}
		default:
			if extra_attrs != nil {
				found := false

				/* Hack: special-case FONT_FEATURES.  We don't want them to
				* override each other, so we never merge them.  This should
				* be fixed when we implement attr-merging. */
				if attr.Type != ATTR_FONT_FEATURES {
					for _, old_attr := range *extra_attrs {
						if attr.Type == old_attr.Type {
							found = true
							break
						}
					}
				}

				if !found {
					*extra_attrs = append(AttrList{attr.pango_attribute_copy()}, *extra_attrs...)
				}
			}
		}
	}

	if haveScale {
		if desc.size_is_absolute {
			desc.pango_font_description_set_absolute_size(int(scale * AttrFloat(desc.size)))
		} else {
			desc.pango_font_description_set_size(int(scale * AttrFloat(desc.size)))
		}
	}
}
