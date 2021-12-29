package pango

import (
	"fmt"
)

// AttrKind is an enum for the supported attributes
// (see the constants).
type AttrKind uint8

const (
	ATTR_INVALID              AttrKind = iota // does not happen
	ATTR_LANGUAGE                             // language (AttrString)
	ATTR_FAMILY                               // font family name list (AttrString)
	ATTR_STYLE                                // font slant style (AttrInt)
	ATTR_WEIGHT                               // font weight (AttrInt)
	ATTR_VARIANT                              // font variant (normal or small caps) (AttrInt)
	ATTR_STRETCH                              // font stretch (AttrInt)
	ATTR_SIZE                                 // font size in points scaled by `Scale` (AttrInt)
	ATTR_FONT_DESC                            // font description (AttrFontDesc)
	ATTR_FOREGROUND                           // foreground color (AttrColor)
	ATTR_BACKGROUND                           // background color (AttrColor)
	ATTR_UNDERLINE                            // whether the text has an underline (AttrInt)
	ATTR_STRIKETHROUGH                        // whether the text is struck-through (AttrInt)
	ATTR_RISE                                 // baseline displacement (AttrInt)
	ATTR_SHAPE                                // shape (AttrShape)
	ATTR_SCALE                                // font size scale factor (AttrFloat)
	ATTR_FALLBACK                             // whether fallback is enabled (AttrInt)
	ATTR_LETTER_SPACING                       // letter spacing (AttrInt)
	ATTR_UNDERLINE_COLOR                      // underline color (AttrColor)
	ATTR_STRIKETHROUGH_COLOR                  // strikethrough color (AttrColor)
	ATTR_ABSOLUTE_SIZE                        // font size in pixels scaled by `Scale` (AttrInt)
	ATTR_GRAVITY                              // base text gravity (AttrInt)
	ATTR_GRAVITY_HINT                         // gravity hint (AttrInt)
	ATTR_FONT_FEATURES                        // OpenType font features (AttrString).
	ATTR_FOREGROUND_ALPHA                     // foreground alpha (AttrInt).
	ATTR_BACKGROUND_ALPHA                     // background alpha (AttrInt).
	ATTR_ALLOW_BREAKS                         // whether breaks are allowed (AttrInt).
	ATTR_SHOW                                 // how to render invisible characters (AttrInt).
	ATTR_INSERT_HYPHENS                       // whether to insert hyphens at intra-word line breaks (AttrInt).
	ATTR_OVERLINE                             // whether the text has an overline (AttrInt).
	ATTR_OVERLINE_COLOR                       // overline color (AttrColor).
	ATTR_LINE_HEIGHT                          // line height factor (AttrFloat)
	ATTR_ABSOLUTE_LINE_HEIGHT                 // line height (AttrInt)
	ATTR_TEXT_TRANSFORM                       // (AttrInt)
	ATTR_WORD                                 // override segmentation to classify the range of the attribute as a single word (AttrInt)
	ATTR_SENTENCE                             // override segmentation to classify the range of the attribute as a single sentence (AttrInt)
	ATTR_BASELINE_SHIFT                       // baseline displacement (AttrInt)
	ATTR_FONT_SCALE                           // font-relative size change (AttrInt)
)

var typeNames = [...]string{
	ATTR_INVALID:              "",
	ATTR_LANGUAGE:             "language",
	ATTR_FAMILY:               "family",
	ATTR_STYLE:                "style",
	ATTR_WEIGHT:               "weight",
	ATTR_VARIANT:              "variant",
	ATTR_STRETCH:              "stretch",
	ATTR_SIZE:                 "size",
	ATTR_FONT_DESC:            "font-desc",
	ATTR_FOREGROUND:           "foreground",
	ATTR_BACKGROUND:           "background",
	ATTR_UNDERLINE:            "underline",
	ATTR_STRIKETHROUGH:        "strikethrough",
	ATTR_RISE:                 "rise",
	ATTR_SHAPE:                "shape",
	ATTR_SCALE:                "scale",
	ATTR_FALLBACK:             "fallback",
	ATTR_LETTER_SPACING:       "letter-spacing",
	ATTR_UNDERLINE_COLOR:      "underline-color",
	ATTR_STRIKETHROUGH_COLOR:  "strikethrough-color",
	ATTR_ABSOLUTE_SIZE:        "absolute-size",
	ATTR_GRAVITY:              "gravity",
	ATTR_GRAVITY_HINT:         "gravity-hint",
	ATTR_FONT_FEATURES:        "font-features",
	ATTR_FOREGROUND_ALPHA:     "foreground-alpha",
	ATTR_BACKGROUND_ALPHA:     "background-alpha",
	ATTR_ALLOW_BREAKS:         "allow-breaks",
	ATTR_SHOW:                 "show",
	ATTR_INSERT_HYPHENS:       "insert-hyphens",
	ATTR_OVERLINE:             "overline",
	ATTR_OVERLINE_COLOR:       "overline-color",
	ATTR_LINE_HEIGHT:          "line-height",
	ATTR_ABSOLUTE_LINE_HEIGHT: "absolute-line-height",
	ATTR_TEXT_TRANSFORM:       "text-transform",
	ATTR_WORD:                 "word",
	ATTR_SENTENCE:             "sentence",
	ATTR_BASELINE_SHIFT:       "baseline-shift",
	ATTR_FONT_SCALE:           "font-scale",
}

func (t AttrKind) String() string {
	if int(t) >= len(typeNames) {
		return "<invalid>"
	}
	return typeNames[t]
}

// ShowFlags affects how Pango treats characters that are normally
// not visible in the output.
type ShowFlags uint8

const (
	SHOW_SPACES      ShowFlags = 1 << iota //  Render spaces, tabs and newlines visibly
	SHOW_LINE_BREAKS                       //  Render line breaks visibly
	SHOW_IGNORABLES                        //  Render default-ignorable Unicode characters visibly
	SHOW_NONE        ShowFlags = 0         //  No special treatment for invisible characters
)

var showflags_map = enumMap{
	{value: int(SHOW_NONE), str: "none"},
	{value: int(SHOW_SPACES), str: "spaces"},
	{value: int(SHOW_LINE_BREAKS), str: "line-breaks"},
	{value: int(SHOW_IGNORABLES), str: "ignorables"},
}

// Underline enumeration is used to specify
// whether text should be underlined, and if so, the type
// of underlining.
type Underline uint8

const (
	UNDERLINE_NONE   Underline = iota // no underline should be drawn
	UNDERLINE_SINGLE                  // a single underline should be drawn
	UNDERLINE_DOUBLE                  // a double underline should be drawn
	// a single underline should be drawn at a
	// position beneath the ink extents of the text being
	// underlined. This should be used only for underlining
	// single characters, such as for keyboard accelerators.
	// UNDERLINE_SINGLE should be used for extended
	// portions of text.
	UNDERLINE_LOW
	// a wavy underline should be drawn below.
	// This underline is typically used to indicate an error such
	// as a possible mispelling; in some cases a contrasting color
	// may automatically be used.
	UNDERLINE_ERROR
	UNDERLINE_SINGLE_LINE // like UNDERLINE_SINGLE, but drawn continuously across multiple runs.
	UNDERLINE_DOUBLE_LINE // like UNDERLINE_DOUBLE, but drawn continuously across multiple runs.
	UNDERLINE_ERROR_LINE  // like UNDERLINE_ERROR, but drawn continuously across multiple runs.
)

var underline_map = enumMap{
	{value: int(UNDERLINE_NONE), str: "none"},
	{value: int(UNDERLINE_SINGLE), str: "single"},
	{value: int(UNDERLINE_DOUBLE), str: "double"},
	{value: int(UNDERLINE_LOW), str: "low"},
	{value: int(UNDERLINE_ERROR), str: "error"},
	{value: int(UNDERLINE_SINGLE_LINE), str: "single-line"},
	{value: int(UNDERLINE_DOUBLE_LINE), str: "double-line"},
	{value: int(UNDERLINE_ERROR_LINE), str: "error-line"},
}

// Overline is used to specify
// whether text should be overlined, and if so, the type
// of line.
type Overline uint8

const (
	OVERLINE_NONE   Overline = iota // no overline should be drawn
	OVERLINE_SINGLE                 // Draw a single line above the ink extents of the text being underlined.
)

var overline_map = enumMap{
	{value: int(OVERLINE_NONE), str: "none"},
	{value: int(OVERLINE_SINGLE), str: "single"},
}

// TextTransform affects how Pango treats characters during shaping.
type TextTransform uint8

const (
	// Leave text unchanged
	TEXT_TRANSFORM_NONE TextTransform = iota
	// Display letters and numbers as lowercase
	TEXT_TRANSFORM_LOWERCASE
	// Display letters and numbers as uppercase
	TEXT_TRANSFORM_UPPERCASE
	// Display the first character of a word
	TEXT_TRANSFORM_CAPITALIZE
)

var textTransformMap = enumMap{
	{value: int(TEXT_TRANSFORM_NONE), str: "none"},
	{value: int(TEXT_TRANSFORM_LOWERCASE), str: "lowercase"},
	{value: int(TEXT_TRANSFORM_UPPERCASE), str: "uppercase"},
	{value: int(TEXT_TRANSFORM_CAPITALIZE), str: "capitalize"},
}

// BaselineShift affects baseline shifts between runs.
type BaselineShift uint8

const (
	// Leave the baseline unchanged
	BASELINE_SHIFT_NONE BaselineShift = iota
	// Shift the baseline to the superscript position, relative to the previous run
	BASELINE_SHIFT_SUPERSCRIPT
	// Shift the baseline to the subscript position, relative to the previous run
	BASELINE_SHIFT_SUBSCRIPT
)

var baselineShitMap = enumMap{
	{value: int(BASELINE_SHIFT_NONE), str: "none"},
	{value: int(BASELINE_SHIFT_SUPERSCRIPT), str: "superscript"},
	{value: int(BASELINE_SHIFT_SUBSCRIPT), str: "subscript"},
}

type FontScale uint8

const (
	FONT_SCALE_NONE        FontScale = iota // Leave the font size unchanged
	FONT_SCALE_SUPERSCRIPT                  // Change the font to a size suitable for superscripts
	FONT_SCALE_SUBSCRIPT                    // Change the font to a size suitable for subscripts
	FONT_SCALE_SMALL_CAPS                   // Change the font to a size suitable for Small Caps.
)

var fontScaleMap = enumMap{
	{value: int(FONT_SCALE_NONE), str: "none"},
	{value: int(FONT_SCALE_SUPERSCRIPT), str: "superscript"},
	{value: int(FONT_SCALE_SUBSCRIPT), str: "subscript"},
}

// AttrData stores the type specific value of
// an attribute.
type AttrData interface {
	fmt.Stringer
	copy() AttrData
	equals(other AttrData) bool
}

// AttrInt is an int
type AttrInt int

func (a AttrInt) copy() AttrData             { return a }
func (a AttrInt) String() string             { return fmt.Sprintf("%d", a) }
func (a AttrInt) equals(other AttrData) bool { return a == other }

// AttrFloat is a float
type AttrFloat Fl

func (a AttrFloat) copy() AttrData             { return a }
func (a AttrFloat) String() string             { return fmt.Sprintf("%f", a) }
func (a AttrFloat) equals(other AttrData) bool { return a == other }

// AttrString is a string
type AttrString string

func (a AttrString) copy() AttrData             { return a }
func (a AttrString) String() string             { return string(a) }
func (a AttrString) equals(other AttrData) bool { return a == other }

// func (a Language) copy() AttrData             { return a }
// func (a Language) String() string             { return string(a) }
// func (a Language) equals(other AttrData) bool { return a == other }

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

func (shape AttrShape) getExtents(nChars int32, inkRect, logicalRect *Rectangle) {
	if nChars > 0 {
		N := Unit(nChars - 1)
		if inkRect != nil {
			inkRect.X = minG(shape.ink.X, shape.ink.X+shape.logical.Width*N)
			inkRect.Width = maxG(shape.ink.Width, shape.ink.Width+shape.logical.Width*N)
			inkRect.Y = shape.ink.Y
			inkRect.Height = shape.ink.Height
		}
		if logicalRect != nil {
			logicalRect.X = minG(shape.logical.X, shape.logical.X+shape.logical.Width*N)
			logicalRect.Width = maxG(shape.logical.Width, shape.logical.Width+shape.logical.Width*N)
			logicalRect.Y = shape.logical.Y
			logicalRect.Height = shape.logical.Height
		}
	} else {
		if inkRect != nil {
			inkRect.X, inkRect.Y, inkRect.Width, inkRect.Height = 0, 0, 0, 0
		}

		if logicalRect != nil {
			logicalRect.X, logicalRect.Y, logicalRect.Width, logicalRect.Height = 0, 0, 0, 0
		}
	}
}

// Attribute are used as the input to the itemization process and also when
// creating a `Layout`. The common portion of the attribute holds
// the range to which the value applies.
// By default an attribute will have an all-inclusive range of [0,maxInt].
type Attribute struct {
	Data AttrData // the kind specific value
	Kind AttrKind
	// Indexes into the underlying rune slice (note that
	// we diverge here from the C library, which works on byte slices).
	// The character at `EndIndex` is not included in the range.
	StartIndex, EndIndex int
}

// init initializes StartIndex to 0 and EndIndex to MaxInt
// such that the attribute applies to the entire text by default.
func (attr *Attribute) init() {
	attr.StartIndex = 0
	attr.EndIndex = MaxInt
}

// Compare two attributes for equality. This compares only the
// actual value of the two attributes and not the ranges that the
// attributes apply to.
func (attr1 Attribute) equals(attr2 Attribute) bool {
	return attr1.Kind == attr2.Kind && attr1.Data.equals(attr2.Data)
}

// Make a deep copy of an attribute.
func (a *Attribute) copy() *Attribute {
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
	return fmt.Sprintf("[%d,%d]%s=%s", int32(attr.StartIndex), int32(attr.EndIndex), attr.Kind, attr.Data)
}

// NewAttrFontDescription creates a new font description attribute. This attribute
// allows setting family, style, weight, variant, stretch,
// and size simultaneously.
func NewAttrFontDescription(desc FontDescription) *Attribute {
	out := Attribute{Kind: ATTR_FONT_DESC, Data: desc}
	out.init()
	return &out
}

// NewAttrShow creates a new attribute that influences how invisible
// characters are rendered.
func NewAttrShow(flags ShowFlags) *Attribute {
	out := Attribute{Kind: ATTR_SHOW, Data: AttrInt(flags)}
	out.init()
	return &out
}

// NewAttrScale creates a new font size scale attribute. The base font for the
// affected text will have its size multiplied by `scaleFactor`.
func NewAttrScale(scaleFactor Fl) *Attribute {
	out := Attribute{Kind: ATTR_SCALE, Data: AttrFloat(scaleFactor)}
	out.init()
	return &out
}

// NewAttrSize creates a new font-size attribute in fractional points.
func NewAttrSize(size int) *Attribute {
	out := Attribute{Kind: ATTR_SIZE, Data: AttrInt(size)}
	out.init()
	return &out
}

// NewAttrWeight creates a new font weight attribute.
func NewAttrWeight(weight Weight) *Attribute {
	out := Attribute{Kind: ATTR_WEIGHT, Data: AttrInt(weight)}
	out.init()
	return &out
}

// NewAttrVariant creates a new font variant attribute (normal or small caps)
func NewAttrVariant(variant Variant) *Attribute {
	out := Attribute{Kind: ATTR_VARIANT, Data: AttrInt(variant)}
	out.init()
	return &out
}

// NewAttrGravity creates a new gravity attribute, which should not be `PANGO_GRAVITY_AUTO`.
func NewAttrGravity(gravity Gravity) *Attribute {
	out := Attribute{Kind: ATTR_GRAVITY, Data: AttrInt(gravity)}
	out.init()
	return &out
}

// NewAttrGravityHint creates a new gravity_hint attribute.
func NewAttrGravityHint(gravityHint GravityHint) *Attribute {
	out := Attribute{Kind: ATTR_GRAVITY_HINT, Data: AttrInt(gravityHint)}
	out.init()
	return &out
}

// NewAttrStyle creates a new font slant style attribute.
func NewAttrStyle(style Style) *Attribute {
	out := Attribute{Kind: ATTR_STYLE, Data: AttrInt(style)}
	out.init()
	return &out
}

// NewAttrAbsoluteSize creates a new font-size attribute in device units.
func NewAttrAbsoluteSize(size int) *Attribute {
	out := Attribute{Kind: ATTR_ABSOLUTE_SIZE, Data: AttrInt(size)}
	out.init()
	return &out
}

// NewAttrLetterSpacing creates a new letter-spacing attribute, the amount of extra space to add between graphemes
// of the text, in Pango units.
func NewAttrLetterSpacing(letterSpacing int32) *Attribute {
	out := Attribute{Kind: ATTR_LETTER_SPACING, Data: AttrInt(letterSpacing)}
	out.init()
	return &out
}

// NewAttrStretch creates a new font stretch attribute.
func NewAttrStretch(stretch Stretch) *Attribute {
	out := Attribute{Kind: ATTR_STRETCH, Data: AttrInt(stretch)}
	out.init()
	return &out
}

// NewAttrRise creates a new baseline displacement attribute: `rise` is the amount
// that the text should be displaced vertically, in Pango units.
// Positive values displace the text upwards.
func NewAttrRise(rise int) *Attribute {
	out := Attribute{Kind: ATTR_RISE, Data: AttrInt(rise)}
	out.init()
	return &out
}

// NewAttrUnderline creates a new underline-style attribute.
func NewAttrUnderline(underline Underline) *Attribute {
	out := Attribute{Kind: ATTR_UNDERLINE, Data: AttrInt(underline)}
	out.init()
	return &out
}

// NewAttrForegroundAlpha creates a new foreground alpha attribute.
func NewAttrForegroundAlpha(alpha uint16) *Attribute {
	out := Attribute{Kind: ATTR_FOREGROUND_ALPHA, Data: AttrInt(alpha)}
	out.init()
	return &out
}

// NewAttrBackgroundAlpha creates a new background alpha attribute.
func NewAttrBackgroundAlpha(alpha uint16) *Attribute {
	out := Attribute{Kind: ATTR_BACKGROUND_ALPHA, Data: AttrInt(alpha)}
	out.init()
	return &out
}

// NewAttrOverline creates a new overline-style attribute.
func NewAttrOverline(overline Overline) *Attribute {
	out := Attribute{Kind: ATTR_OVERLINE, Data: AttrInt(overline)}
	out.init()
	return &out
}

// NewAttrStrikethrough creates a new strikethrough-style attribute.
func NewAttrStrikethrough(strikethrough bool) *Attribute {
	v := AttrInt(0)
	if strikethrough {
		v = 1
	}
	out := Attribute{Kind: ATTR_STRIKETHROUGH, Data: v}
	out.init()
	return &out
}

// NewAttrAllowBreaks creates a new allow-breaks attribute.
//
// If breaks are disabled, the range will be kept in a
// single run, as far as possible.
func NewAttrAllowBreaks(allowBreaks bool) *Attribute {
	v := AttrInt(0)
	if allowBreaks {
		v = 1
	}
	out := Attribute{Kind: ATTR_ALLOW_BREAKS, Data: v}
	out.init()
	return &out
}

// NewAttrInsertHyphens creates a new insert-hyphens attribute.
//
// Pango will insert hyphens when breaking lines in the middle
// of a word. This attribute can be used to suppress the hyphen.
func NewAttrInsertHyphens(insertHyphens bool) *Attribute {
	v := AttrInt(0)
	if insertHyphens {
		v = 1
	}
	out := Attribute{Kind: ATTR_INSERT_HYPHENS, Data: v}
	out.init()
	return &out
}

// NewAttrFallback creates a new font fallback attribute.
//
// If fallback is disabled, characters will only be used from the
// closest matching font on the system. No fallback will be done to
// other fonts on the system that might contain the characters in the
// text.
func NewAttrFallback(enableFallback bool) *Attribute {
	f := 0
	if enableFallback {
		f = 1
	}
	out := Attribute{Kind: ATTR_FALLBACK, Data: AttrInt(f)}
	out.init()
	return &out
}

// NewAttrLanguage creates a new language tag attribute
func NewAttrLanguage(language Language) *Attribute {
	out := Attribute{Kind: ATTR_LANGUAGE, Data: AttrString(language)}
	out.init()
	return &out
}

// NewAttrFamily creates a new font family attribute: `family` is
// the family or comma separated list of families.
func NewAttrFamily(family string) *Attribute {
	out := Attribute{Kind: ATTR_FAMILY, Data: AttrString(family)}
	out.init()
	return &out
}

// NewAttrForeground creates a new foreground color attribute.
func NewAttrForeground(color AttrColor) *Attribute {
	out := Attribute{Kind: ATTR_FOREGROUND, Data: color}
	out.init()
	return &out
}

// NewAttrBackground creates a new background color attribute.
func NewAttrBackground(color AttrColor) *Attribute {
	out := Attribute{Kind: ATTR_BACKGROUND, Data: color}
	out.init()
	return &out
}

// NewAttrUnderlineColor creates a new underline color attribute. This attribute
// modifies the color of underlines. If not set, underlines
// will use the foreground color.
func NewAttrUnderlineColor(color AttrColor) *Attribute {
	out := Attribute{Kind: ATTR_UNDERLINE_COLOR, Data: color}
	out.init()
	return &out
}

// NewAttrOverlineColor creates a new overline color attribute. This attribute
// modifies the color of overlines. If not set, overlines
// will use the foreground color.
func NewAttrOverlineColor(color AttrColor) *Attribute {
	out := Attribute{Kind: ATTR_OVERLINE_COLOR, Data: color}
	out.init()
	return &out
}

// NewAttrStrikethroughColor creates a new strikethrough color attribute. This attribute
// modifies the color of strikethrough lines. If not set, strikethrough lines
// will use the foreground color.
func NewAttrStrikethroughColor(color AttrColor) *Attribute {
	out := Attribute{Kind: ATTR_STRIKETHROUGH_COLOR, Data: color}
	out.init()
	return &out
}

// NewAttrShape creates a new shape attribute. A shape is used to impose a
// particular ink and logical rectangle on the result of shaping a
// particular glyph. This might be used, for instance, for
// embedding a picture or a widget inside a `Layout`.
func NewAttrShape(ink, logical Rectangle) *Attribute {
	out := Attribute{Kind: ATTR_SHAPE, Data: AttrShape{ink: ink, logical: logical}}
	out.init()
	return &out
}

// NewAttrFontFeatures creates a new font features tag attribute, from a string with OpenType font features, in CSS syntax
func NewAttrFontFeatures(features string) *Attribute {
	out := Attribute{Kind: ATTR_FONT_FEATURES, Data: AttrString(features)}
	out.init()
	return &out
}

// NewAttrLineHeight modifies the height of logical line extents by a factor.
//
// This affects the values returned by
// LayoutLine.getExtents(),
// LayoutLine.getPixelExtents() and
// LayoutIter.getLineExtents() methods.
func NewAttrLineHeight(factor float32) *Attribute {
	out := Attribute{Kind: ATTR_LINE_HEIGHT, Data: AttrFloat(factor)}
	out.init()
	return &out
}

// NewAttrAbsoluteLineHeight overrides the height of logical line extents to be `height`,
// in `Scale`-ths of a point
//
// This affects the values returned by
// LayoutLine.getExtents(),
// LayoutLine.getPixelExtents() and
// LayoutIter.getLineExtents() methods.
func NewAttrAbsoluteLineHeight(height int) *Attribute {
	out := Attribute{Kind: ATTR_ABSOLUTE_LINE_HEIGHT, Data: AttrInt(height)}
	out.init()
	return &out
}

// NewAttrWord marks the range of the attribute as a single word.
//
// Note that this may require adjustments to word and
// sentence classification around the range.
func NewAttrWord() *Attribute {
	out := Attribute{Kind: ATTR_WORD, Data: AttrInt(1)}
	out.init()
	return &out
}

// NewAttrSentence marks the range of the attribute as a single sentence.
//
// Note that this may require adjustments to word and
// sentence classification around the range.
func NewAttrSentence() *Attribute {
	out := Attribute{Kind: ATTR_SENTENCE, Data: AttrInt(1)}
	out.init()
	return &out
}

// NewAttrTextTransform marks the range of the attribute as a single sentence.
//
// Note that this may require adjustments to word and
// sentence classification around the range.
func NewAttrTextTransform(textTransform TextTransform) *Attribute {
	out := Attribute{Kind: ATTR_TEXT_TRANSFORM, Data: AttrInt(textTransform)}
	out.init()
	return &out
}

// NewAttrBaselineShift creates a new baseline displacement attribute.
//
// The effect of this attribute is to shift the baseline of a run,
// relative to the run of preceding run.
// `rise` is either a `BaselineShift` enumeration value or an absolute value (> 1024)
// in Pango units, relative to the baseline of the previous run.
// Positive values displace the text upwards.
func NewAttrBaselineShift(rise int) *Attribute {
	out := Attribute{Kind: ATTR_BASELINE_SHIFT, Data: AttrInt(rise)}
	out.init()
	return &out
}

// NewAttrFontScale creates a new font scale attribute.
//
// The effect of this attribute is to change the font size of a run,
// relative to the size of preceding run. `scale` indicates font size change relative
// to the size of the previous run.
func NewAttrFontScale(scale FontScale) *Attribute {
	out := Attribute{Kind: ATTR_FONT_SCALE, Data: AttrInt(scale)}
	out.init()
	return &out
}

type AttrList []*Attribute

// copy returns a deep copy of the list,
// calling `deepCopy` for each element.
func (list AttrList) copy() AttrList {
	out := make(AttrList, len(list))
	for i, v := range list {
		out[i] = v.copy()
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
				list.insertAt(i, attr)
				break
			}
		}
	}
}

func (l *AttrList) insertAt(i int, attr *Attribute) {
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
func (list *AttrList) Insert(attr *Attribute) {
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

// Change inserts the given attribute into the `AttrList`. It will
// replace any attributes of the same type on that segment
// and be merged with any adjoining attributes that are identical.
//
// This function is slower than insert() for
// creating an attribute list in order (potentially much slower
// for large lists). However, insert() is not
// suitable for continually changing a set of attributes
// since it never removes or combines existing attributes.
func (list *AttrList) Change(attr *Attribute) {
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
		list.Insert(attr)
		return
	}

	var i, p int
	inserted := false
	for i, p = 0, len(*list); i < p; i++ {
		tmp_attr := (*list)[i]

		if tmp_attr.StartIndex > startIndex {
			list.insertAt(i, attr)
			inserted = true
			break
		}

		if tmp_attr.Kind != attr.Kind {
			continue
		}

		if tmp_attr.EndIndex < startIndex {
			continue /* This attr does not overlap with the new one */
		}

		// tmp_attr.StartIndex <= startIndex
		// tmp_attr.EndIndex >= startIndex

		if tmp_attr.equals(*attr) { // We can merge the new attribute with this attribute
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
				end_attr := tmp_attr.copy()
				end_attr.StartIndex = endIndex
				list.Insert(end_attr)
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
		list.Insert(attr)
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

		if tmp_attr.Kind != attr.Kind {
			continue
		}

		if tmp_attr.EndIndex <= attr.EndIndex || tmp_attr.equals(*attr) {
			/* We can merge the new attribute with this attribute. */
			attr.EndIndex = max(endIndex, tmp_attr.EndIndex)
			list.remove(i)
			i--
			p--
			continue
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
				attr.EndIndex == otherAttr.EndIndex && attr.equals(*otherAttr) {
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

// Given a AttrList and callback function, removes any elements
// of `list` for which `fn` returns `true` and inserts them into
// a new list (possibly empty if no attributes of the given types were found).
func (list *AttrList) Filter(fn func(attr *Attribute) bool) AttrList {
	if list == nil {
		return nil
	}
	var out AttrList
	for i, p := 0, len(*list); i < p; i++ {
		tmpAttr := (*list)[i]
		if fn(tmpAttr) {
			list.remove(i)
			i-- /* Need to look at this index again */
			p--
			out = append(out, tmpAttr)
		}
	}

	return out
}

// attrIterator is used to represent an
// iterator through an `AttrList`. A new iterator is created
// with getIterator(). Once the iterator
// is created, it can be advanced through the style changes
// in the text using next(). At each
// style change, the range of the current style segment and the
// attributes currently in effect can be queried.
type attrIterator struct {
	attrs *AttrList

	attribute_stack AttrList

	attrIndex            int
	StartIndex, EndIndex int // index into the underlying text
}

// getIterator creates a iterator initialized to the beginning of the list.
// `list` must not be modified until this iterator is freed.
func (list *AttrList) getIterator() *attrIterator {
	if list == nil {
		return nil
	}
	iterator := attrIterator{attrs: list}

	if !iterator.next() {
		iterator.EndIndex = MaxInt
	}

	return &iterator
}

// copy returns a copy of `iterator`.
func (iterator *attrIterator) copy() *attrIterator {
	if iterator == nil {
		return nil
	}
	copy := *iterator
	// dont deep copy the attributes themselves
	copy.attribute_stack = append(AttrList{}, iterator.attribute_stack...)
	return &copy
}

// next advances the iterator until the next change of style, and
// returns `false` if the iterator is at the end of the list, otherwise `true`
func (iterator *attrIterator) next() bool {
	if iterator == nil {
		return false
	}

	if iterator.attrIndex >= len(*iterator.attrs) && len(iterator.attribute_stack) == 0 {
		return false
	}
	iterator.StartIndex = iterator.EndIndex
	iterator.EndIndex = MaxInt

	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		if attr.EndIndex == iterator.StartIndex {
			iterator.attribute_stack.remove(i)
		} else {
			iterator.EndIndex = min(iterator.EndIndex, attr.EndIndex)
		}
	}

	for {
		if iterator.attrIndex >= len(*iterator.attrs) {
			break
		}
		attr := (*iterator.attrs)[iterator.attrIndex]

		if attr.StartIndex != iterator.StartIndex {
			break
		}

		if attr.EndIndex > iterator.StartIndex {
			iterator.attribute_stack = append(iterator.attribute_stack, attr)
			iterator.EndIndex = min(iterator.EndIndex, attr.EndIndex)
		}

		iterator.attrIndex++ /* NEXT! */
	}

	if iterator.attrIndex < len(*iterator.attrs) {
		attr := (*iterator.attrs)[iterator.attrIndex]
		iterator.EndIndex = min(iterator.EndIndex, attr.StartIndex)
	}

	return true
}

// getAttributes gets a list of all attributes at the current position of the
// iterator.
func (iterator attrIterator) getAttributes() AttrList {
	var attrs AttrList

	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		found := false

		if attr.Kind != ATTR_FONT_DESC &&
			attr.Kind != ATTR_BASELINE_SHIFT && attr.Kind != ATTR_FONT_SCALE { // keep all font attributes in the returned list
			for _, oldAttr := range attrs {
				if attr.Kind == oldAttr.Kind {
					found = true
					break
				}
			}
		}

		if !found {
			attrs = append(AttrList{attr}, attrs...)
		}
	}

	return attrs
}

// getByKind finds the current attribute of a particular type at the iterator
// location. When multiple attributes of the same type overlap,
// the attribute whose range starts closest to the current location is used.
// It returns `nil` if no attribute of that type applies to the current location.
func (iterator attrIterator) getByKind(kind AttrKind) *Attribute {
	for i := len(iterator.attribute_stack) - 1; i >= 0; i-- {
		attr := iterator.attribute_stack[i]
		if attr.Kind == kind {
			return attr
		}
	}
	return nil
}

// getFont gets the font and other attributes at the current iterator position.
// `desc` is a FontDescription to fill in with the current values.
// If non-nil, `language` is a location to store language tag for item, or zero if none is found.
// If non-nil, `extraAttrs` is a location in which to store a list of non-font
// attributes at the the current position; only the highest priority
// value of each attribute will be added to this list.
func (iterator attrIterator) getFont(desc *FontDescription, lang *Language, extraAttrs *AttrList) {
	if desc == nil {
		return
	}
	//    int i;

	if lang != nil {
		*lang = ""
	}

	if extraAttrs != nil {
		*extraAttrs = nil
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

		switch attr.Kind {
		case ATTR_FONT_DESC:
			attrDesc := attr.Data.(FontDescription)
			new_mask := attrDesc.mask & ^mask
			mask |= new_mask
			desc.UnsetFields(new_mask)
			desc.pango_font_description_merge(&attrDesc, false)
		case ATTR_FAMILY:
			if mask&FmFamily == 0 {
				mask |= FmFamily
				desc.SetFamily(string(attr.Data.(AttrString)))
			}
		case ATTR_STYLE:
			if mask&FmStyle == 0 {
				mask |= FmStyle
				desc.SetStyle(Style(attr.Data.(AttrInt)))
			}
		case ATTR_VARIANT:
			if mask&FmVariant == 0 {
				mask |= FmVariant
				desc.SetVariant(Variant(attr.Data.(AttrInt)))
			}
		case ATTR_WEIGHT:
			if mask&FmWeight == 0 {
				mask |= FmWeight
				desc.SetWeight(Weight(attr.Data.(AttrInt)))
			}
		case ATTR_STRETCH:
			if mask&FmStretch == 0 {
				mask |= FmStretch
				desc.SetStretch(Stretch(attr.Data.(AttrInt)))
			}
		case ATTR_SIZE:
			if mask&FmSize == 0 {
				mask |= FmSize
				desc.SetSize(int32(attr.Data.(AttrInt)))
			}
		case ATTR_ABSOLUTE_SIZE:
			if mask&FmSize == 0 {
				mask |= FmSize
				desc.SetAbsoluteSize(int32(attr.Data.(AttrInt)))
			}
		case ATTR_SCALE:
			if !haveScale {
				haveScale = true
				scale = attr.Data.(AttrFloat)
			}
		case ATTR_LANGUAGE:
			if lang != nil {
				if !haveLanguage {
					haveLanguage = true
					*lang = Language(attr.Data.(AttrString))
				}
			}
		default:
			if extraAttrs != nil {
				found := false

				// Hack: special-case FONT_FEATURES, BASELINE_SHIFT and FONT_SCALE.
				// We don't want these to accumulate, not override each other,
				// so we never merge them.
				// This needs to be handled more systematically.
				if attr.Kind != ATTR_FONT_FEATURES &&
					attr.Kind != ATTR_BASELINE_SHIFT && attr.Kind != ATTR_FONT_SCALE {
					for _, old_attr := range *extraAttrs {
						if attr.Kind == old_attr.Kind {
							found = true
							break
						}
					}
				}

				if !found {
					*extraAttrs = append(AttrList{attr.copy()}, *extraAttrs...)
				}
			}
		}
	}

	if haveScale {
		if desc.SizeIsAbsolute {
			desc.SetAbsoluteSize(int32(scale * AttrFloat(desc.Size)))
		} else {
			desc.SetSize(int32(scale * AttrFloat(desc.Size)))
		}
	}
}
