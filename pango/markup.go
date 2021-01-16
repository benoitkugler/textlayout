package pango

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

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

func (scaleLevel sizeLevel) scaleFactor(base float64) float64 {
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
	index       int
	to_apply    AttrList
	accelMarker rune
	accelChar   rune
}

func (md *markupData) markup_data_open_tag() *openTag {
	if len(md.attr_list) != 0 {
		return nil
	}

	var parent *openTag
	if len(md.tag_stack) != 0 {
		parent = md.tag_stack[0]
	}

	ot := &openTag{}
	ot.start_index = md.index

	if parent == nil {
		ot.base_scale_factor = 1.0
	} else {
		ot.base_scale_factor = parent.base_scale_factor
		ot.base_font_size = parent.base_font_size
		ot.has_base_font_size = parent.has_base_font_size
		ot.scale_level = parent.scale_level
	}

	// prepend
	md.tag_stack = append(md.tag_stack, nil)
	copy(md.tag_stack[1:], md.tag_stack)
	md.tag_stack[0] = ot

	return ot
}

// markup_data_close_tag
func (md *markupData) end_element_handler() {
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
		a.StartIndex = ot.start_index
		a.EndIndex = md.index
		md.to_apply[L-1-i] = a
	}

	if ot.scale_level_delta != 0 {
		// We affected relative font size; create an appropriate
		// attribute and reverse our effects on the current level
		var a *Attribute
		if ot.has_base_font_size {
			// Create a font using the absolute point size
			// as the base size to be scaled from
			a = pango_attr_size_new(int(ot.scale_level.scaleFactor(1.0) * float64(ot.base_font_size)))
		} else {
			// Create a font using the current scale factor
			// as the base size to be scaled from
			a = pango_attr_scale_new(ot.scale_level.scaleFactor(ot.base_scale_factor))
		}

		a.StartIndex = ot.start_index
		a.EndIndex = md.index

		md.to_apply.insert(0, a)
	}
}

func (user_data *markupData) start_element_handler(element_name string, attrs []xml.Attr) error {
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
		return fmt.Errorf("unknown tag '%s'", element_name)
	}

	ot := user_data.markup_data_open_tag()

	// note ot may be nil if the user didn't want the attribute list
	err := parse_func(ot, attrs)
	return err
}

func (md *markupData) text_handler(text []rune) {
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
				ulineIndex = md.index

				/* set next range_start to include this char */
				rangeStart = i
			}

			/* reset range_end */
			rangeEnd = -1
		} else if c == md.accelMarker {
			rangeEnd = i
		}
	}

	if rangeEnd != -1 {
		rangeEnd = len(text)
	}
	md.text = append(md.text, text[rangeStart:rangeEnd]...)
	md.index += rangeEnd - rangeStart

	if len(md.attr_list) != 0 && ulineIndex >= 0 {
		//  Add the underline indicating the accelerator
		attr := pango_attr_underline_new(PANGO_UNDERLINE_LOW)
		attr.StartIndex = ulineIndex
		attr.EndIndex = ulineIndex + 1
		md.attr_list.pango_attr_list_change(attr)
	}
}

func (n *markupData) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// start by handling the new element
	err := n.start_element_handler(start.Name.Local, start.Attr)
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
			n.text_handler([]rune(string(next)))
		case xml.EndElement:
			// closing current element: return after processing
			n.end_element_handler()
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
	attrs       AttrList
	start_index int
	/* Current total scale level; reset whenever
	* an absolute size is set.
	* Each "larger" ups it 1, each "smaller" decrements it 1
	 */
	scale_level sizeLevel
	/* Our impact on scale_level, so we know whether we
	* need to create an attribute ourselves on close
	 */
	scale_level_delta int
	/* Base scale factor currently in effect
	* or size that this tag
	* forces, or parent's scale factor or size.
	 */
	base_scale_factor  float64
	base_font_size     int
	has_base_font_size bool // = 1;
}

func (ot *openTag) add_attribute(attr *Attribute) {
	if ot == nil {
		return
	}
	ot.attrs.insert(0, attr)
}

func (ot *openTag) open_tag_set_absolute_font_scale(scale float64) {
	ot.base_scale_factor = scale
	ot.has_base_font_size = false
	ot.scale_level = 0
	ot.scale_level_delta = 0
}

func (ot *openTag) open_tag_set_absolute_font_size(fontSize int) {
	ot.base_font_size = fontSize
	ot.has_base_font_size = true
	ot.scale_level = 0
	ot.scale_level_delta = 0
}

func newParser(accelMarker rune) *markupData {
	md := &markupData{}
	md.accelMarker = accelMarker
	return md
}

// Simple markup language for text with attributes
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
// `AttrList` using pango_parse_markup()
//
// A simple example of a marked-up string might be:
//
// <span foreground="blue" size="x-large">Blue text</span> is <i>cool</i>!
//
//
// Pango uses #GMarkup to parse this language, which means that XML
// features such as numeric character entities such as `&#169;` for
// © can be used too.
//
// The root tag of a marked-up document is `<markup>`, but
// pango_parse_markup() allows you to omit this tag, so you will most
// likely never need to use it. The most general markup tag is `<span>`,
// then there are some convenience tags.
//
// ## Span attributes
//
// `<span>` has the following attributes:
//
// * `font_desc`:
//   A font description string, such as "Sans Italic 12".
//   See pango_font_description_from_string() for a description of the
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
type ParsedMarkup struct {
	Attr      AttrList
	Text      []rune // Text with tags stripped
	AccelChar rune   // Accelerator char
}

// pango_parse_markup parses marked-up text to create
// a plain-text string and an attribute list.
//
// If `accelMarker` is nonzero, the given character will mark the
// character following it as an accelerator. For example, `accelMarker`
// might be an ampersand or underscore. All characters marked
// as an accelerator will receive a `PANGO_UNDERLINE_LOW` attribute,
// and the first character so marked will be returned in `accelChar`.
// Two `accelMarker` characters following each other produce a single
// literal `accelMarker` character.
func pango_parse_markup(markup_text []byte, accelMarker rune) (ParsedMarkup, error) {
	markup_text = bytes.TrimLeft(markup_text, " \t\n\r")

	nested := append(append([]byte("<markup>"), markup_text...), "</markup>"...)

	context := newParser(accelMarker)
	err := xml.Unmarshal(nested, context)
	if err != nil {
		return ParsedMarkup{}, err
	}

	return pango_markup_parser_finish(context), nil
}

/**
 * pango_markup_parser_finish:
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
func pango_markup_parser_finish(md *markupData) ParsedMarkup {
	// The apply list has the most-recently-closed tags first;
	// we want to apply the least-recently-closed tag last.
	for _, attr := range md.to_apply {
		// Innermost tags before outermost
		md.attr_list.pango_attr_list_insert(attr)
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

func checkNoAttrs(elem string, names []xml.Attr) error {
	if len(names) != 0 {
		return fmt.Errorf("tag '%s' does not support attributes", elem)
	}
	return nil
}

func b_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("b", names); err != nil {
		return err
	}
	tag.add_attribute(pango_attr_weight_new(PANGO_WEIGHT_BOLD))
	return nil
}

func big_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("big", names); err != nil {
		return err
	}
	/* Grow text one level */
	if tag != nil {
		tag.scale_level_delta += 1
		tag.scale_level += 1
	}

	return nil
}

func parseAbsoluteSize(tag *openTag, size string) bool {
	level := medium
	switch size {
	case "xx-small":
		level = xxSmall
	case "x-small":
		level = xSmall
	case "small":
		level = small
	case "medium":
		level = medium
	case "large":
		level = large
	case "x-large":
		level = xLarge
	case "xx-large":
		level = xxLarge
	default:
		return false
	}

	// This is "absolute" in that it's relative to the base font,
	// but not to sizes created by any other tags
	factor := level.scaleFactor(1.0)
	tag.add_attribute(pango_attr_scale_new(factor))
	if tag != nil {
		tag.open_tag_set_absolute_font_scale(factor)
	}
	return true
}

func spanParseInt(attrName, attrVal string) (int, error) {
	out, err := strconv.Atoi(attrVal)
	if err != nil {
		return 0, fmt.Errorf("value of '%s' attribute on <span> tag should be an integer, not '%s': %s", attrName, attrVal, err)
	}
	return out, nil
}

func spanParseBoolean(attrName, attrVal string) (bool, error) {
	switch attrVal {
	case "true", "yes", "t", "y":
		return true, nil
	case "false", "no", "f", "n":
		return false, nil
	default:
		return false, fmt.Errorf("value of '%s' attribute on <span> tag should have one of "+
			"'true/yes/t/y' or 'false/no/f/n': '%s' is not valid", attrName, attrVal)
	}
}

func spanParseColor(attrName, attrVal string, withAlpha bool) (AttrColor, uint16, error) {
	out, alpha, ok := pango_color_parse_with_alpha(attrVal, withAlpha)
	if !ok {
		return out, alpha, fmt.Errorf("value of '%s' attribute on <span> tag should be a color specification, not '%s'",
			attrName, attrVal)
	}

	return out, alpha, nil
}

func spanParseAlpha(attrName, attrVal string) (uint16, error) {
	hasPercent := false
	if strings.HasSuffix(attrVal, "%") {
		attrVal = attrVal[:len(attrVal)-1]
		hasPercent = true
	}
	intVal, err := strconv.Atoi(attrVal)
	if err != nil {
		return 0, fmt.Errorf("value of '%s' attribute on <span> tag should be an integer, not '%s': %s",
			attrName, attrVal, err)
	}

	if !hasPercent {
		return uint16(intVal), nil
	}

	if intVal > 0 && intVal <= 100 {
		return uint16(intVal * 0xffff / 100), nil
	}
	return 0, fmt.Errorf("value of '%s' attribute on <span> tag should be between 0 and 65536 or a percentage, not '%s'",
		attrName, attrVal)
}

func span_parse_enum(attrName, attrVal string, enum enumMap) (int, error) {
	out, ok := enum.fromString(attrVal)

	if !ok {
		return 0, fmt.Errorf("'%s' is not a valid value for the '%s' "+
			"attribute on <span> tag; valid values are %s",
			attrVal, attrName, enum.possibleValues())
	}

	return out, nil
}

func spanParseShowflags(attrName, attrVal string) (ShowFlags, error) {
	flags := strings.Split(attrVal, "|")
	var out ShowFlags
	for _, flag := range flags {
		v, ok := showflags_map.fromString(flag)
		if !ok {
			return 0, fmt.Errorf("'%s' is not a valid value for the '%s' "+
				"attribute on <span> tag; valid values are %s or combinations with |",
				attrVal, attrName, showflags_map.possibleValues())
		}
		out |= ShowFlags(v)
	}
	return out, nil
}

func checkAttribute(value *string, newAttrName, newAttrValue string) error {
	if *value != "" {
		return fmt.Errorf("attribute '%s' occurs twice on <span> tag, may only occur once", newAttrName)
	}
	*value = newAttrValue
	return nil
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
			return fmt.Errorf("attribute '%s' is not allowed on the <span> tag", newAttrName)
		}
	}

	// Parse desc first, then modify it with other font-related attributes.
	if desc != "" {
		parsed := pango_font_description_from_string(desc)
		tag.add_attribute(pango_attr_font_desc_new(parsed))
		if tag != nil {
			tag.open_tag_set_absolute_font_size(parsed.size)
		}
	}

	if family != "" {
		tag.add_attribute(pango_attr_family_new(family))
	}

	if size != "" {
		if size == "smaller" {
			if tag != nil {
				tag.scale_level_delta -= 1
				tag.scale_level -= 1
			}
		} else if size == "larger" {
			if tag != nil {
				tag.scale_level_delta += 1
				tag.scale_level += 1
			}
		} else if parseAbsoluteSize(tag, size) {
			/* nothing */
		} else {
			n, err := strconv.Atoi(size)
			if err != nil {
				return fmt.Errorf("value of 'size' attribute on <span> tag should be an integer or a string such as 'small', not '%s'",
					size)
			}

			tag.add_attribute(pango_attr_size_new(n))
			if tag != nil {
				tag.open_tag_set_absolute_font_size(n)
			}
		}
	}

	if style != "" {
		if pangoStyle, ok := pango_parse_style(style); ok {
			tag.add_attribute(pango_attr_style_new(pangoStyle))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'style' attribute on <span>; "+
				"valid values are 'normal', 'oblique', 'italic'", style)
		}
	}

	if weight != "" {
		if pangoWeight, ok := pango_parse_weight(weight); ok {
			tag.add_attribute(pango_attr_weight_new(pangoWeight))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'weight' "+
				"attribute on <span> tag; valid values are for example 'light', 'ultrabold' or a number",
				weight)
		}
	}

	if variant != "" {
		if pangoVariant, ok := pango_parse_variant(variant); ok {
			tag.add_attribute(pango_attr_variant_new(pangoVariant))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'variant' "+
				"attribute on <span> tag; valid values are "+
				"'normal', 'smallcaps'", variant)
		}
	}

	if stretch != "" {
		if pangoStretch, ok := pango_parse_stretch(stretch); ok {
			tag.add_attribute(pango_attr_stretch_new(pangoStretch))
		} else {
			return fmt.Errorf("'%s' is not a valid value for the 'stretch' "+
				"attribute on <span> tag; valid values are for example 'condensed', "+
				"'ultraexpanded', 'normal'", stretch)
		}
	}

	if foreground != "" {
		color, alpha, err := spanParseColor("foreground", foreground, true)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_foreground_new(color))
		if alpha != 0xffff {
			tag.add_attribute(pango_attr_foreground_alpha_new(alpha))
		}
	}

	if background != "" {
		color, alpha, err := spanParseColor("background", background, true)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_background_new(color))
		if alpha != 0xffff {
			tag.add_attribute(pango_attr_background_alpha_new(alpha))
		}
	}

	if alpha != "" {
		val, err := spanParseAlpha("alpha", alpha)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_foreground_alpha_new(val))
	}

	if backgroundAlpha != "" {
		val, err := spanParseAlpha("background_alpha", backgroundAlpha)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_background_alpha_new(val))
	}

	if underline != "" {
		ul, err := span_parse_enum("underline", underline, underline_map)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_underline_new(Underline(ul)))
	}

	if underlineColor != "" {
		color, _, err := spanParseColor("underline_color", underlineColor, false)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_underline_color_new(color))
	}

	if overline != "" {
		ol, err := span_parse_enum("overline", overline, overline_map)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_overline_new(Overline(ol)))
	}

	if overlineColor != "" {
		color, _, err := spanParseColor("overline_color", overlineColor, false)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_overline_color_new(color))
	}

	if gravity != "" {
		gr, err := span_parse_enum("gravity", gravity, gravity_map)
		if err != nil {
			return err
		}
		if Gravity(gr) == PANGO_GRAVITY_AUTO {
			return fmt.Errorf("'%s' is not a valid value for the 'gravity' "+
				"attribute on <span> tag; valid values are for example 'south', 'east', 'north', 'west'",
				gravity)
		}
		tag.add_attribute(pango_attr_gravity_new(Gravity(gr)))
	}

	if gravityHint != "" {
		hint, err := span_parse_enum("gravity_hint", gravityHint, gravityhint_map)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_gravity_hint_new(GravityHint(hint)))
	}

	if strikethrough != "" {
		b, err := spanParseBoolean("strikethrough", strikethrough)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_strikethrough_new(b))
	}

	if strikethrough_color != "" {
		color, _, err := spanParseColor("strikethrough_color", strikethrough_color, false)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_strikethrough_color_new(color))
	}

	if fallback != "" {
		b, err := spanParseBoolean("fallback", fallback)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_fallback_new(b))
	}

	if show != "" {
		flags, err := spanParseShowflags("show", show)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_show_new(flags))
	}

	if rise != "" {
		n, err := spanParseInt("rise", rise)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_rise_new(n))
	}

	if letterSpacing != "" {
		n, err := spanParseInt("letter_spacing", letterSpacing)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_letter_spacing_new(n))
	}

	if lang != "" {
		tag.add_attribute(pango_attr_language_new(pango_language_from_string(lang)))
	}

	if fontFeatures != "" {
		tag.add_attribute(pango_attr_font_features_new(fontFeatures))
	}

	if allow_breaks != "" {
		b, err := spanParseBoolean("allow_breaks", allow_breaks)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_allow_breaks_new(b))
	}

	if insertHyphens != "" {
		b, err := spanParseBoolean("insert_hyphens", insertHyphens)
		if err != nil {
			return err
		}
		tag.add_attribute(pango_attr_insert_hyphens_new(b))
	}

	return nil
}

func i_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("i", names); err != nil {
		return err
	}
	tag.add_attribute(pango_attr_style_new(PANGO_STYLE_ITALIC))
	return nil
}

func markup_parse_func(tag *openTag, names []xml.Attr) error {
	/* We don't do anything with this tag at the moment. */
	return nil
}

func s_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("s", names); err != nil {
		return err
	}
	tag.add_attribute(pango_attr_strikethrough_new(true))
	return nil
}

const SUPERSUB_RISE = 5000

func sub_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("sub", names); err != nil {
		return err
	}
	/* Shrink font, and set a negative rise */
	if tag != nil {
		tag.scale_level_delta -= 1
		tag.scale_level -= 1
	}
	tag.add_attribute(pango_attr_rise_new(-SUPERSUB_RISE))
	return nil
}

func sup_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("sup", names); err != nil {
		return err
	}
	/* Shrink font, and set a positive rise */
	if tag != nil {
		tag.scale_level_delta -= 1
		tag.scale_level -= 1
	}
	tag.add_attribute(pango_attr_rise_new(SUPERSUB_RISE))
	return nil
}

func small_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("small", names); err != nil {
		return err
	}
	// Shrink text one level
	if tag != nil {
		tag.scale_level_delta -= 1
		tag.scale_level -= 1
	}
	return nil
}

func tt_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("tt", names); err != nil {
		return err
	}
	tag.add_attribute(pango_attr_family_new("Monospace"))
	return nil
}

func u_parse_func(tag *openTag, names []xml.Attr) error {
	if err := checkNoAttrs("u", names); err != nil {
		return err
	}
	tag.add_attribute(pango_attr_underline_new(PANGO_UNDERLINE_SINGLE))
	return nil
}

//  /**
//   * pango_markup_parser_new:
//   * `accelMarker`: character that precedes an accelerator, or 0 for none
//   *
//   * Parses marked-up text (see
//   * <link linkend="PangoMarkupFormat">markup format</link>) to create
//   * a plain-text string and an attribute list.
//   *
//   * If `accelMarker` is nonzero, the given character will mark the
//   * character following it as an accelerator. For example, `accelMarker`
//   * might be an ampersand or underscore. All characters marked
//   * as an accelerator will receive a %PANGO_UNDERLINE_LOW attribute,
//   * and the first character so marked will be returned in `accelChar`,
//   * when calling finish(). Two `accelMarker` characters following each
//   * other produce a single literal `accelMarker` character.
//   *
//   * To feed markup to the parser, use g_markup_parse_context_parse()
//   * on the returned #GMarkupParseContext. When done with feeding markup
//   * to the parser, use pango_markup_parser_finish() to get the data out
//   * of it, and then use g_markup_parse_context_free() to free it.
//   *
//   * This function is designed for applications that read pango markup
//   * from streams. To simply parse a string containing pango markup,
//   * the simpler pango_parse_markup() API is recommended instead.
//   *
//   * Return value: (transfer none): a #GMarkupParseContext that should be
//   * destroyed with g_markup_parse_context_free().
//   *
//   * Since: 1.31.0
//   **/
//  GMarkupParseContext *
//  pango_markup_parser_new (gunichar               accelMarker)
//  {
//    GError *error = nil;
//    GMarkupParseContext *context;
//    context = newParser (accelMarker, &error, true);

//    if (context == nil)
// 	 g_critical ("Had error when making markup parser: %s\n", error.message);

//    return context;
//  }
