package pango

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

const lineColNumber = 88888 // FIXME: use xml position available in go 1.18 release

// CSS size levels
type sizeLevel int8

const (
	xxSmall sizeLevel = -3
	xSmall  sizeLevel = -2
	small   sizeLevel = -1
	medium  sizeLevel = 0
	large   sizeLevel = 1
	xLarge  sizeLevel = 2
	xxLarge sizeLevel = 3
)

func (scaleLevel sizeLevel) scaleFactor(base Fl) Fl {
	factor := base

	// 1.2 is the CSS scale factor between sizes

	if scaleLevel > 0 {
		for i := sizeLevel(0); i < scaleLevel; i++ {
			factor *= 1.2
		}
	} else if scaleLevel < 0 {
		for i := scaleLevel; i < 0; i++ {
			factor /= 1.2
		}
	}

	return factor
}

type markupData struct {
	attr_list   AttrList
	text        []rune
	tag_stack   []*openTag
	to_apply    AttrList
	index       int
	accelMarker rune
	accelChar   rune
}

func (md *markupData) openTag() *openTag {
	if len(md.attr_list) != 0 {
		return nil
	}

	var parent *openTag
	if len(md.tag_stack) != 0 {
		parent = md.tag_stack[0]
	}

	ot := &openTag{}
	ot.startIndex = md.index

	if parent == nil {
		ot.baseScaleFactor = 1.0
	} else {
		ot.baseScaleFactor = parent.baseScaleFactor
		ot.baseFontSize = parent.baseFontSize
		ot.hasBaseFontSize = parent.hasBaseFontSize
		ot.scaleLevel = parent.scaleLevel
	}

	// prepend
	md.tag_stack = append(md.tag_stack, nil)
	copy(md.tag_stack[1:], md.tag_stack)
	md.tag_stack[0] = ot

	return ot
}

// markup_data_close_tag
func (md *markupData) endElementHandler() {
	// pop the stack
	ot := md.tag_stack[0]
	md.tag_stack = md.tag_stack[1:]

	// Adjust end indexes, and push each attr onto the front of the
	// to_apply list. This means that outermost tags are on the front of
	// that list; if we apply the list in order, then the innermost
	// tags will "win", which is correct.
	// tmp_list := ot.attrs
	L := len(ot.attrs)
	md.to_apply = append(make(AttrList, L), md.to_apply...) // allocate the space needed
	for i, a := range ot.attrs {
		a.StartIndex = ot.startIndex
		a.EndIndex = md.index
		md.to_apply[L-1-i] = a
	}

	if ot.scaleLevelDelta != 0 {
		// We affected relative font size; create an appropriate
		// attribute and reverse our effects on the current level
		var a *Attribute
		if ot.hasBaseFontSize {
			// Create a font using the absolute point size
			// as the base size to be scaled from
			a = NewAttrSize(int(ot.scaleLevel.scaleFactor(1.0) * Fl(ot.baseFontSize)))
		} else {
			// Create a font using the current scale factor
			// as the base size to be scaled from
			a = NewAttrScale(ot.scaleLevel.scaleFactor(ot.baseScaleFactor))
		}

		a.StartIndex = ot.startIndex
		a.EndIndex = md.index

		md.to_apply.insertAt(0, a)
	}
}

func (user_data *markupData) startElementHandler(element_name string, attrs []xml.Attr) error {
	var parse_func tagParseFunc

	switch element_name {
	case "b":
		parse_func = b_parse_func
	case "big":
		parse_func = big_parse_func
	case "i":
		parse_func = i_parse_func
	case "markup":
		parse_func = markup_parse_func
	case "span":
		parse_func = span_parse_func
	case "s":
		parse_func = s_parse_func
	case "sub":
		parse_func = sub_parse_func
	case "sup":
		parse_func = sup_parse_func
	case "small":
		parse_func = small_parse_func
	case "tt":
		parse_func = tt_parse_func
	case "u":
		parse_func = u_parse_func
	default:
		return fmt.Errorf("Unknown tag '%s' on line %d char %d", element_name, lineColNumber, lineColNumber)
	}

	ot := user_data.openTag()

	// note ot may be nil if the user didn't want the attribute list
	err := parse_func(ot, attrs)
	return err
}

func (md *markupData) textHandler(text []rune) {
	if md.accelMarker == 0 { // just append all the text
		md.text = append(md.text, text...)
		md.index += len(text)
		return
	}

	// Parse the accelerator

	var (
		ulineIndex = -1
		// ranges are index into text, or -1
		rangeEnd   = -1
		rangeStart = 0
	)

	for i, c := range text {
		if rangeEnd != -1 {
			if c == md.accelMarker {
				// escaped accel marker; move range_end past the accel marker that came before,
				// append the whole thing
				rangeEnd += 1
				md.text = append(md.text, text[rangeStart:rangeEnd]...)
				md.index += rangeEnd - rangeStart

				// set next range_start, skipping accel marker
				rangeStart = i + 1
			} else {
				// Don't append the accel marker (leave range_end
				// alone); set the accel char to c; record location for
				// underline attribute
				if md.accelChar == 0 {
					md.accelChar = c
				}

				md.text = append(md.text, text[rangeStart:rangeEnd]...)
				md.index += rangeEnd - rangeStart

				// The underline should go underneath the char
				// we're setting as the next range_start
				// Add the underline indicating the accelerator
				attr := NewAttrUnderline(UNDERLINE_LOW)
				ulineIndex = md.index
				attr.StartIndex = ulineIndex
				attr.EndIndex = ulineIndex + 1

				md.attr_list.Change(attr)

				/* set next range_start to include this char */
				rangeStart = i
			}

			/* reset range_end */
			rangeEnd = -1
		} else if c == md.accelMarker {
			rangeEnd = i
		}
	}

	md.text = append(md.text, text[rangeStart:]...)
	md.index += len(text) - rangeStart
}

func (n *markupData) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// start by handling the new element
	err := n.startElementHandler(start.Name.Local, start.Attr)
	if err != nil {
		return err
	}
	// then process the inner content: text or kid element
	for {
		next, err := d.Token()
		if err != nil {
			return err
		}

		// Token is one of StartElement, EndElement, CharData, Comment, ProcInst, or Directive
		switch next := next.(type) {
		case xml.CharData:
			// handle text and keep going
			n.textHandler([]rune(string(next)))
		case xml.EndElement:
			// closing current element: return after processing
			n.endElementHandler()
			return nil
		case xml.StartElement:
			// new kid: recurse and keep going for other kids or text
			err := n.UnmarshalXML(d, next)
			if err != nil {
				return err
			}
		default:
			// ignored, just keep going
		}
	}
}

type openTag struct {
	attrs      AttrList
	startIndex int
	/* Our impact on scale_level, so we know whether we
	* need to create an attribute ourselves on close
	 */
	scaleLevelDelta int
	baseFontSize    int32
	/* Base scale factor currently in effect
	* or size that this tag
	* forces, or parent's scale factor or size.
	 */
	baseScaleFactor Fl
	/* Current total scale level; reset whenever
	* an absolute size is set.
	* Each "larger" ups it 1, each "smaller" decrements it 1
	 */
	scaleLevel      sizeLevel
	hasBaseFontSize bool // = 1;
}

func (ot *openTag) addAttribute(attr *Attribute) {
	if ot == nil {
		return
	}
	ot.attrs.insertAt(0, attr)
}

func (ot *openTag) setAbsoluteFontScale(scale Fl) {
	ot.baseScaleFactor = scale
	ot.hasBaseFontSize = false
	ot.scaleLevel = 0
	ot.scaleLevelDelta = 0
}

func (ot *openTag) setAbsoluteFontSize(fontSize int32) {
	ot.baseFontSize = fontSize
	ot.hasBaseFontSize = true
	ot.scaleLevel = 0
	ot.scaleLevelDelta = 0
}

func newParser(accelMarker rune) *markupData {
	md := &markupData{}
	md.accelMarker = accelMarker
	return md
}

// ParsedMarkup exposes the result
// of parsing a simple markup language for text with attributes.
//
// Frequently, you want to display some text to the user with attributes
// applied to part of the text (for example, you might want bold or
// italicized words). With the base Pango interfaces, you could create a
// `AttrList` and apply it to the text; the problem is that you'd
// need to apply attributes to some numeric range of characters, for
// example "characters 12-17." This is broken from an internationalization
// standpoint; once the text is translated, the word you wanted to
// italicize could be in a different position.
//
// The solution is to include the text attributes in the string to be
// translated. Pango provides this feature with a small markup language.
// You can parse a marked-up string into the string text plus a
// `AttrList` using ParseMarkup().
//
// A simple example of a marked-up string might be:
//
// <span foreground="blue" size="x-large">Blue text</span> is <i>cool</i>!
//
// See ParseMarkup for complete syntax.
type ParsedMarkup struct {
	Attr      AttrList // Attributes extracted from the markup
	Text      []rune   // Text with tags stripped
	AccelChar rune     // Accelerator char
}

// ParseMarkup parses marked-up text to create
// a plain-text string and an attribute list.
//
// If `accelMarker` is nonzero, the given character will mark the
// character following it as an accelerator. For example, `accelMarker`
// might be an ampersand or underscore. All characters marked
// as an accelerator will receive a `PANGO_UNDERLINE_LOW` attribute,
// and the first character so marked will be returned in `accelChar`.
// Two `accelMarker` characters following each other produce a single
// literal `accelMarker` character.
//
// A marked-up document follows an XML format : the root tag is `<markup>`, but
// ParseMarkup() allows you to omit this tag, so you will most
// likely never need to use it. The most general markup tag is `<span>`,
// then there are some convenience tags.
//
// ## Span attributes
//
// `<span>` has the following attributes:
//
// * `font_desc`:
//   A font description string, such as "Sans Italic 12".
//   See NewFontDescriptionFromString() for a description of the
//   format of the string representation . Note that any other span
//   attributes will override this description. So if you have "Sans Italic"
//   and also a `style="normal"` attribute, you will get Sans normal,
//   not italic.
//
// * `font_family`:
//   A font family name
//
// * `font_size`, `size`:
//   Font size in 1024ths of a point, or one of the absolute
//   sizes `xx-small`, `x-small`, `small`, `medium`, `large`,
//   `x-large`, `xx-large`, or one of the relative sizes `smaller`
//   or `larger`. If you want to specify a absolute size, it's usually
//   easier to take advantage of the ability to specify a partial
//   font description using `font`; you can use `font='12.5'`
//   rather than `size='12800'`.
//
// * `font_style`:
//   One of `normal`, `oblique`, `italic`
//
// * `font_weight`:
//   One of `ultralight`, `light`, `normal`, `bold`,
//   `ultrabold`, `heavy`, or a numeric weight
//
// * `font_variant`:
//   One of `normal` or `smallcaps`
//
// * `font_stretch`, `stretch`:
//   One of `ultracondensed`, `extracondensed`, `condensed`,
//   `semicondensed`, `normal`, `semiexpanded`, `expanded`,
//   `extraexpanded`, `ultraexpanded`
//
// * `font_features`:
//   A comma-separated list of OpenType font feature
//   settings, in the same syntax as accepted by CSS. E.g:
//   `font_features='dlig=1, -kern, afrc on'`
//
// * `foreground`, `fgcolor`:
//   An RGB color specification such as `#00FF00` or a color
//   name such as `red`. Since 1.38, an RGBA color specification such
//   as `#00FF007F` will be interpreted as specifying both a foreground
//   color and foreground alpha.
//
// * `background`, `bgcolor`:
//   An RGB color specification such as `#00FF00` or a color
//   name such as `red`.
//   Since 1.38, an RGBA color specification such as `#00FF007F` will
//   be interpreted as specifying both a background color and
//   background alpha.
//
// * `alpha`, `fgalpha`:
//   An alpha value for the foreground color, either a plain
//   integer between 1 and 65536 or a percentage value like `50%`.
//
// * `background_alpha`, `bgalpha`:
//   An alpha value for the background color, either a plain
//   integer between 1 and 65536 or a percentage value like `50%`.
//
// * `underline`:
//   One of `none`, `single`, `double`, `low`, `error`,
//   `single-line`, `double-line` or `error-line`.
//
// * `underline_color`:
//   The color of underlines; an RGB color
//   specification such as `#00FF00` or a color name such as `red`
//
// * `overline`:
//   One of `none` or `single`
//
// * `overline_color`:
//   The color of overlines; an RGB color
//   specification such as `#00FF00` or a color name such as `red`
//
// * `rise`:
//   Vertical displacement, in Pango units. Can be negative for
//   subscript, positive for superscript.
//
// * `strikethrough`
//   `true` or `false` whether to strike through the text
//
// * `strikethrough_color`:
//   The color of strikethrough lines; an RGB
//   color specification such as `#00FF00` or a color name such as `red`
//
// * `fallback`:
//   `true` or `false` whether to enable fallback. If
//   disabled, then characters will only be used from the closest
//   matching font on the system. No fallback will be done to other
//   fonts on the system that might contain the characters in the text.
//   Fallback is enabled by default. Most applications should not
//   disable fallback.
//
// * `allow_breaks`:
//   `true` or `false` whether to allow line breaks or not. If
//   not allowed, the range will be kept in a single run as far
//   as possible. Breaks are allowed by default.
//
// * `insert_hyphens`:`
//   `true` or `false` whether to insert hyphens when breaking
//   lines in the middle of a word. Hyphens are inserted by default.
//
// * `show`:
//   A value determining how invisible characters are treated.
//   Possible values are `spaces`, `line-breaks`, `ignorables`
//   or combinations, such as `spaces|line-breaks`.
//
// * `lang`:
//   A language code, indicating the text language
//
// * `letter_spacing`:
//   Inter-letter spacing in 1024ths of a point.
//
// * `gravity`:
//   One of `south`, `east`, `north`, `west`, `auto`.
//
// * `gravity_hint`:
//   One of `natural`, `strong`, `line`.
//
// ## Convenience tags
//
// The following convenience tags are provided:
//
// * `<b>`:
//   Bold
//
// * `<big>`:
//   Makes font relatively larger, equivalent to `<span size="larger">`
//
// * `<i>`:
//   Italic
//
// * `<s>`:
//   Strikethrough
//
// * `<sub>`:
//   Subscript
//
// * `<sup>`:
//   Superscript
//
// * `<small>`:
//   Makes font relatively smaller, equivalent to `<span size="smaller">`
//
// * `<tt>`:
//   Monospace
//
// * `<u>`:
//   Underline
func ParseMarkup(markupText []byte, accelMarker rune) (ParsedMarkup, error) {
	nested := append(append([]byte("<markup>"), markupText...), "</markup>"...)

	context := newParser(accelMarker)
	err := xml.Unmarshal(nested, context)
	if err != nil {
		return ParsedMarkup{}, err
	}

	return markupParserFinish(context), nil
}

/**
 * markupParserFinish:
 * @context: A valid parse context that was returned from pango_markup_parser_new()
 * @attr_list: (out) (allow-none): address of return location for a `AttrList`, or %nil
 * @text: (out) (allow-none): address of return location for text with tags stripped, or %nil
 * `accelChar`: (out) (allow-none): address of return location for accelerator char, or %nil
 * @error: address of return location for errors, or %nil
 *
 * After feeding a pango markup parser some data with g_markup_parse_context_parse(),
 * use this function to get the list of pango attributes and text out of the
 * markup. This function will not free @context, use g_markup_parse_context_free()
 * to do so.
 */
func markupParserFinish(md *markupData) ParsedMarkup {
	// The apply list has the most-recently-closed tags first;
	// we want to apply the least-recently-closed tag last.
	for _, attr := range md.to_apply {
		// Innermost tags before outermost
		md.attr_list.Insert(attr)
	}
	md.to_apply = nil

	var out ParsedMarkup
	out.Attr = md.attr_list
	md.attr_list = nil

	out.Text = md.text
	out.AccelChar = md.accelChar

	return out
}

type tagParseFunc = func(tag *openTag, names []xml.Attr) error

// check that names is empty
func checkNoAttrs(elem string, names []xml.Attr) error {
	if len(names) != 0 {
		return fmt.Errorf("Tag '%s' does not support attribute '%s' on line %d char %d",
			elem, names[0].Name.Local, lineColNumber, lineColNumber)
	}
	return nil
}

func b_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("b", names); err != nil {
		return err
	}
	tag.addAttribute(NewAttrWeight(WEIGHT_BOLD))
	return nil
}

func big_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("big", names); err != nil {
		return err
	}
	/* Grow text one level */
	if tag != nil {
		tag.scaleLevelDelta += 1
		tag.scaleLevel += 1
	}

	return nil
}

func parsePercentage(input string) (Fl, bool) {
	if !strings.HasSuffix(input, "%") {
		return 0, false
	}
	input = strings.TrimSuffix(input, "%")
	out, err := strconv.ParseFloat(input, 32)
	return Fl(out), err == nil
}

func parseAbsoluteSize(tag *openTag, size string) bool {
	var factor Fl
	switch size {
	// This is "absolute" in that it's relative to the base font,
	// but not to sizes created by any other tags
	case "xx-small":
		factor = xxSmall.scaleFactor(1.)
	case "x-small":
		factor = xSmall.scaleFactor(1.)
	case "small":
		factor = small.scaleFactor(1.)
	case "medium":
		factor = medium.scaleFactor(1.)
	case "large":
		factor = large.scaleFactor(1.)
	case "x-large":
		factor = xLarge.scaleFactor(1.)
	case "xx-large":
		factor = xxLarge.scaleFactor(1.)
	default:
		if val, ok := parsePercentage(size); ok {
			factor = val / 100
		} else {
			return false
		}
	}

	tag.addAttribute(NewAttrScale(factor))
	if tag != nil {
		tag.setAbsoluteFontScale(factor)
	}
	return true
}

func spanParseInt(attrName, attrVal string) (int32, error) {
	out, err := strconv.ParseInt(attrVal, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Value of '%s' attribute on <span> tag on line %d could not be parsed; should be an integer, not '%s'",
			attrName, lineColNumber, attrVal)
	}
	return int32(out), nil
}

func spanParseFloat(attrName, attrVal string) (Fl, error) {
	out, err := strconv.ParseFloat(attrVal, 32)
	if err != nil {
		return 0, fmt.Errorf("Value of '%s' attribute on <span> tag should be a float, not '%s': %s", attrName, attrVal, err)
	}
	return Fl(out), nil
}

func spanParseBoolean(attrName, attrVal string) (bool, error) {
	switch attrVal {
	case "true", "yes", "t", "y":
		return true, nil
	case "false", "no", "f", "n":
		return false, nil
	default:
		return false, fmt.Errorf("Value of '%s' attribute on <span> tag line %d should have one of "+
			"'true/yes/t/y' or 'false/no/f/n': '%s' is not valid", attrName, lineColNumber, attrVal)
	}
}

func spanParseColor(attrName, attrVal string, withAlpha bool) (AttrColor, uint16, error) {
	out, alpha, ok := pango_color_parse_with_alpha(attrVal, withAlpha)
	if !ok {
		return out, alpha, fmt.Errorf("Value of '%s' attribute on <span> tag on line %d could not be parsed; should be a color specification, not '%s'",
			attrName, lineColNumber, attrVal)
	}

	return out, alpha, nil
}

func spanParseAlpha(attrName, attrVal string) (uint16, error) {
	hasPercent := false
	if strings.HasSuffix(attrVal, "%") {
		attrVal = attrVal[:len(attrVal)-1]
		hasPercent = true
	}
	intVal, err := strconv.ParseInt(attrVal, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Value of '%s' attribute on <span> tag on line %d could not be parsed; should be an integer, not '%s'",
			attrName, lineColNumber, attrVal)
	}

	if !hasPercent {
		return uint16(intVal), nil
	}

	if intVal > 0 && intVal <= 100 {
		return uint16(intVal * 0xffff / 100), nil
	}
	return 0, fmt.Errorf("Value of '%s' attribute on <span> tag on line %d could not be parsed; should be between 0 and 65536 or a percentage, not '%s'",
		attrName, lineColNumber, attrVal)
}

func spanParseEnum(attrName, attrVal string, enum enumMap) (int, error) {
	out, ok := enum.FromString(attrVal)

	if !ok {
		return 0, fmt.Errorf("'%s' is not a valid value for the '%s' "+
			"attribute on <span> tag, line %d; valid values are %s",
			attrVal, attrName, lineColNumber, enum.possibleValues())
	}

	return out, nil
}

func spanParseShowflags(attrName, attrVal string) (ShowFlags, error) {
	flags := strings.Split(attrVal, "|")
	var out ShowFlags
	for _, flag := range flags {
		flag = strings.TrimSpace(flag)
		v, ok := showflags_map.FromString(flag)
		if !ok {
			return 0, fmt.Errorf("'%s' is not a valid value for the '%s' "+
				"attribute on <span> tag, line %d; valid values are %s or combinations with |",
				attrVal, attrName, lineColNumber, showflags_map.possibleValues())
		}
		out |= ShowFlags(v)
	}
	return out, nil
}

func checkAttribute(value *string, newAttrName, newAttrValue string) error {
	if *value != "" {
		return fmt.Errorf("Attribute '%s' occurs twice on <span> tag on line %d char %d, may only occur once",
			newAttrName, lineColNumber, lineColNumber)
	}
	*value = newAttrValue
	return nil
}

func parseLength(attrVal string) (int, bool) {
	if v, err := strconv.Atoi(attrVal); err == nil {
		return v, true
	}

	if !strings.HasSuffix(attrVal, "pt") {
		return 0, false
	}
	val, err := strconv.ParseFloat(strings.TrimSuffix(attrVal, "pt"), 32)

	return int(val * Scale), err == nil
}

func span_parse_func(tag *openTag, attrs []xml.Attr) error {
	var (
		family              string
		size                string
		style               string
		weight              string
		variant             string
		stretch             string
		desc                string
		foreground          string
		background          string
		underline           string
		underlineColor      string
		overline            string
		overlineColor       string
		strikethrough       string
		strikethrough_color string
		rise                string
		baselineShift       string
		letterSpacing       string
		lang                string
		fallback            string
		gravity             string
		gravityHint         string
		fontFeatures        string
		alpha               string
		backgroundAlpha     string
		allow_breaks        string
		insertHyphens       string
		show                string
		lineHeight          string
		textTransform       string
		segment             string
		fontScale           string
	)

	for _, attr := range attrs {
		newAttrName := strings.ReplaceAll(attr.Name.Local, "-", "_")
		switch newAttrName {
		case "allow_breaks":
			err := checkAttribute(&allow_breaks, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "alpha":
			err := checkAttribute(&alpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "background":
			err := checkAttribute(&background, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "bgcolor":
			err := checkAttribute(&background, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "background_alpha":
			err := checkAttribute(&backgroundAlpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "bgalpha":
			err := checkAttribute(&backgroundAlpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "baseline_shift":
			err := checkAttribute(&baselineShift, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "color":
			err := checkAttribute(&foreground, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "fallback":
			err := checkAttribute(&fallback, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font":
			err := checkAttribute(&desc, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_desc":
			err := checkAttribute(&desc, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "face":
			err := checkAttribute(&family, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_family":
			err := checkAttribute(&family, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_size":
			err := checkAttribute(&size, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_stretch":
			err := checkAttribute(&stretch, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_style":
			err := checkAttribute(&style, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_variant":
			err := checkAttribute(&variant, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_weight":
			err := checkAttribute(&weight, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "foreground":
			err := checkAttribute(&foreground, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "fgcolor":
			err := checkAttribute(&foreground, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "fgalpha":
			err := checkAttribute(&alpha, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_features":
			err := checkAttribute(&fontFeatures, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "font_scale":
			err := checkAttribute(&fontScale, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "show":
			err := checkAttribute(&show, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "size":
			err := checkAttribute(&size, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "stretch":
			err := checkAttribute(&stretch, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "strikethrough":
			err := checkAttribute(&strikethrough, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "strikethrough_color":
			err := checkAttribute(&strikethrough_color, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "style":
			err := checkAttribute(&style, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "segment":
			err := checkAttribute(&segment, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "text_transform":
			err := checkAttribute(&textTransform, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "gravity":
			err := checkAttribute(&gravity, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "gravity_hint":
			err := checkAttribute(&gravityHint, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "insert_hyphens":
			err := checkAttribute(&insertHyphens, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "lang":
			err := checkAttribute(&lang, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "letter_spacing":
			err := checkAttribute(&letterSpacing, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "line_height":
			err := checkAttribute(&lineHeight, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "overline":
			err := checkAttribute(&overline, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "overline_color":
			err := checkAttribute(&overlineColor, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "underline":
			err := checkAttribute(&underline, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "underline_color":
			err := checkAttribute(&underlineColor, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "rise":
			err := checkAttribute(&rise, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "variant":
			err := checkAttribute(&variant, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		case "weight":
			err := checkAttribute(&weight, newAttrName, attr.Value)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Attribute '%s' is not allowed on the <span> tag on line %d char %d", newAttrName, lineColNumber, lineColNumber)
		}
	}

	// Parse desc first, then modify it with other font-related attributes.
	if desc != "" {
		parsed := NewFontDescriptionFrom(desc)
		tag.addAttribute(NewAttrFontDescription(parsed))
		if tag != nil {
			tag.setAbsoluteFontSize(parsed.Size)
		}
	}

	if family != "" {
		tag.addAttribute(NewAttrFamily(family))
	}

	if size != "" {
		if n, ok := parseLength(size); ok {
			tag.addAttribute(NewAttrSize(n))
			if tag != nil {
				tag.setAbsoluteFontSize(int32(n))
			}
		} else if size == "smaller" {
			if tag != nil {
				tag.scaleLevelDelta -= 1
				tag.scaleLevel -= 1
			}
		} else if size == "larger" {
			if tag != nil {
				tag.scaleLevelDelta += 1
				tag.scaleLevel += 1
			}
		} else if parseAbsoluteSize(tag, size) {
			/* nothing */
		} else {
			return fmt.Errorf("Value of 'size' attribute on <span> tag on line %d could not be parsed; should be an integer, or a string such as 'small', not '%s'",
				lineColNumber, size)
		}
	}

	if style != "" {
		if pangoStyle, ok := pango_parse_style(style); ok {
			tag.addAttribute(NewAttrStyle(pangoStyle))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'style' attribute on <span> tag, line %d; "+
				"valid values are 'normal', 'oblique', 'italic'",
				style, lineColNumber)
		}
	}

	if weight != "" {
		if pangoWeight, ok := pango_parse_weight(weight); ok {
			tag.addAttribute(NewAttrWeight(pangoWeight))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'weight' "+
				"attribute on <span> tag, line %d; valid values are for example 'light', 'ultrabold' or a number",
				weight, lineColNumber)
		}
	}

	if variant != "" {
		if pangoVariant, ok := pango_parse_variant(variant); ok {
			tag.addAttribute(NewAttrVariant(pangoVariant))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'variant' "+
				"attribute on <span> tag, line %d; valid values are "+
				"'normal', 'smallcaps'", variant, lineColNumber)
		}
	}

	if stretch != "" {
		if pangoStretch, ok := pango_parse_stretch(stretch); ok {
			tag.addAttribute(NewAttrStretch(pangoStretch))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'stretch' "+
				"attribute on <span> tag, line %d; valid values are for example 'condensed', "+
				"'ultraexpanded', 'normal'", stretch, lineColNumber)
		}
	}

	if foreground != "" {
		color, alpha, err := spanParseColor("foreground", foreground, true)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrForeground(color))
		if alpha != 0xffff {
			tag.addAttribute(NewAttrForegroundAlpha(alpha))
		}
	}

	if background != "" {
		color, alpha, err := spanParseColor("background", background, true)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrBackground(color))
		if alpha != 0xffff {
			tag.addAttribute(NewAttrBackgroundAlpha(alpha))
		}
	}

	if alpha != "" {
		val, err := spanParseAlpha("alpha", alpha)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrForegroundAlpha(val))
	}

	if backgroundAlpha != "" {
		val, err := spanParseAlpha("background_alpha", backgroundAlpha)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrBackgroundAlpha(val))
	}

	if underline != "" {
		ul, err := spanParseEnum("underline", underline, underline_map)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrUnderline(Underline(ul)))
	}

	if underlineColor != "" {
		color, _, err := spanParseColor("underline_color", underlineColor, false)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrUnderlineColor(color))
	}

	if overline != "" {
		ol, err := spanParseEnum("overline", overline, overline_map)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrOverline(Overline(ol)))
	}

	if overlineColor != "" {
		color, _, err := spanParseColor("overline_color", overlineColor, false)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrOverlineColor(color))
	}

	if gravity != "" {
		gr, err := spanParseEnum("gravity", gravity, GravityMap)
		if err != nil {
			return err
		}
		if Gravity(gr) == GRAVITY_AUTO {
			return fmt.Errorf("'%s' is not a valid value for the 'gravity' "+
				"attribute on <span> tag, line %d; valid values are for example 'south', 'east', 'north', 'west'",
				gravity, lineColNumber)
		}
		tag.addAttribute(NewAttrGravity(Gravity(gr)))
	}

	if gravityHint != "" {
		hint, err := spanParseEnum("gravity_hint", gravityHint, gravityhint_map)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrGravityHint(GravityHint(hint)))
	}

	if strikethrough != "" {
		b, err := spanParseBoolean("strikethrough", strikethrough)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrStrikethrough(b))
	}

	if strikethrough_color != "" {
		color, _, err := spanParseColor("strikethrough_color", strikethrough_color, false)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrStrikethroughColor(color))
	}

	if fallback != "" {
		b, err := spanParseBoolean("fallback", fallback)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrFallback(b))
	}

	if show != "" {
		flags, err := spanParseShowflags("show", show)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrShow(flags))
	}

	if textTransform != "" {
		v, err := spanParseEnum("text_transform", textTransform, textTransformMap)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrTextTransform(TextTransform(v)))
	}

	if rise != "" {
		if n, ok := parseLength(rise); ok {
			tag.addAttribute(NewAttrRise(n))
		} else {
			return fmt.Errorf("Value of 'rise' attribute on <span> tag on line %d "+
				"could not be parsed; should be an integer, or a "+
				"string such as '5.5pt', not '%s'", lineColNumber, rise)
		}
	}

	if baselineShift != "" {
		if shift, err := spanParseEnum("baseline_shift", baselineShift, baselineShitMap); err == nil {
			tag.addAttribute(NewAttrBaselineShift(shift))
		} else if shift, ok := parseLength(baselineShift); ok && (shift > 1024 || shift < -1024) {
			tag.addAttribute(NewAttrBaselineShift(shift))
		} else {
			return fmt.Errorf("Value of 'baseline_shift' attribute on <span> tag on line %d "+
				"could not be parsed; should be 'superscript' or 'subscript' or an integer, or a "+
				"string such as '5.5pt', not '%s'", lineColNumber, rise)
		}
	}

	if fontScale != "" {
		fs, err := spanParseEnum("font_scale", baselineShift, fontScaleMap)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrFontScale(FontScale(fs)))
	}

	if letterSpacing != "" {
		n, err := spanParseInt("letter_spacing", letterSpacing)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrLetterSpacing(n))
	}

	if lineHeight != "" {
		f, err := spanParseFloat("line_height", lineHeight)
		if err != nil {
			return err
		}
		if f > 1024.0 && !strings.ContainsRune(lineHeight, '.') {
			tag.addAttribute(NewAttrAbsoluteLineHeight(int(f)))
		} else {
			tag.addAttribute(NewAttrLineHeight(f))
		}
	}

	if lang != "" {
		tag.addAttribute(NewAttrLanguage(pango_language_from_string(lang)))
	}

	if fontFeatures != "" {
		tag.addAttribute(NewAttrFontFeatures(fontFeatures))
	}

	if allow_breaks != "" {
		b, err := spanParseBoolean("allow_breaks", allow_breaks)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrAllowBreaks(b))
	}

	if insertHyphens != "" {
		b, err := spanParseBoolean("insert_hyphens", insertHyphens)
		if err != nil {
			return err
		}
		tag.addAttribute(NewAttrInsertHyphens(b))
	}

	if segment != "" {
		switch segment {
		case "word":
			tag.addAttribute(NewAttrWord())
		case "sentence":
			tag.addAttribute(NewAttrSentence())
		default:
			return fmt.Errorf("Value of 'segment' attribute on <span> tag on line %d "+
				"could not be parsed; should be one of 'word' or 'sentence', not '%s'", lineColNumber, segment)
		}
	}
	return nil
}

func i_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("i", names); err != nil {
		return err
	}
	tag.addAttribute(NewAttrStyle(STYLE_ITALIC))
	return nil
}

func markup_parse_func(tag *openTag, names []xml.Attr) error {
	/* We don't do anything with this tag at the moment. */
	return checkNoAttrs("markup", names)
}

func s_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("s", names); err != nil {
		return err
	}
	tag.addAttribute(NewAttrStrikethrough(true))
	return nil
}

func sub_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("sub", names); err != nil {
		return err
	}
	tag.addAttribute(NewAttrFontScale(FONT_SCALE_SUBSCRIPT))
	tag.addAttribute(NewAttrBaselineShift(int(BASELINE_SHIFT_SUBSCRIPT)))
	return nil
}

func sup_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("sup", names); err != nil {
		return err
	}
	tag.addAttribute(NewAttrFontScale(FONT_SCALE_SUPERSCRIPT))
	tag.addAttribute(NewAttrBaselineShift(int(BASELINE_SHIFT_SUPERSCRIPT)))
	return nil
}

func small_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("small", names); err != nil {
		return err
	}
	// Shrink text one level
	if tag != nil {
		tag.scaleLevelDelta -= 1
		tag.scaleLevel -= 1
	}
	return nil
}

func tt_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("tt", names); err != nil {
		return err
	}
	tag.addAttribute(NewAttrFamily("Monospace"))
	return nil
}

func u_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("u", names); err != nil {
		return err
	}
	tag.addAttribute(NewAttrUnderline(UNDERLINE_SINGLE))
	return nil
}
