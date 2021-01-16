package pango

import (
	"unicode"

	"github.com/benoitkugler/go-weasyprint/fribidi"
)

const (
	// Whether the segment should be shifted to center around the baseline.
	// Used in vertical writing directions mostly.
	PANGO_ANALYSIS_FLAG_CENTERED_BASELINE = 1 << iota
	// Used to mark runs that hold ellipsized text in an ellipsized layout
	PANGO_ANALYSIS_FLAG_IS_ELLIPSIS
	// Add an hyphen at the end of the run during shaping
	PANGO_ANALYSIS_FLAG_NEED_HYPHEN
)

//  /**
//   * PANGO_ANALYSIS_FLAG_IS_ELLIPSIS:
//   *
//   * This flag is .
//   *
//   * Since: 1.36.7
//   */

//  /**
//   * PANGO_ANALYSIS_FLAG_NEED_HYPHEN:
//   *
//   * This flag tells Pango to
//   *
//   * Since: 1.44
//   */

// Analysis stores information about the properties of a segment of text.
type Analysis struct {
	// shape_engine *PangoEngineShape
	// lang_engine  *PangoEngineLang
	font Font // the font for this segment.

	level   fribidi.Level //  the bidirectional level for this segment.
	gravity Gravity       //  the glyph orientation for this segment (A #PangoGravity).
	flags   uint8         //  boolean flags for this segment (Since: 1.16).

	script   Script   // the detected script for this segment (A #PangoScript) (Since: 1.18).
	language Language // the detected language for this segment.

	extra_attrs AttrList // extra attributes for this segment.
}

func (analysis *Analysis) showing_space() bool {
	for _, attr := range analysis.extra_attrs {
		if attr.Type == ATTR_SHOW && (ShowFlags(attr.Data.(AttrInt))&PANGO_SHOW_SPACES != 0) {
			return true
		}
	}
	return false
}

// Item stores information about a segment of text.
type Item struct {
	offset    int      // rune offset of the start of this item in text.
	num_chars int      // number of Unicode characters in the item.
	analysis  Analysis // analysis results for the item.
}

// pango_item_copy copy an existing `Item`.
func (item *Item) pango_item_copy() *Item {
	if item == nil {
		return nil
	}
	result := *item // shallow copy
	result.analysis.extra_attrs = item.analysis.extra_attrs.pango_attr_list_copy()
	return &result
}

// pango_item_split modifies `orig` to cover only the text after `splitIndex`,
// which is relative to the start of the item,
// and returns a new item that covers the text before `splitIndex` that
// used to be in `orig`. You can think of `splitIndex` as the length of
// the returned item.
// pango_item_split returns `nil` if `splitIndex` is 0 or
// greater than or equal to the length of `orig` (that is, there must
// be at least one byte assigned to each item, you can't create a
// zero-length item).
//
// A new item representing text before `splitIndex` is returned.
func (orig *Item) pango_item_split(splitIndex int) *Item {
	if splitIndex <= 0 || splitIndex >= orig.num_chars {
		return nil
	}

	new_item := orig.pango_item_copy()
	new_item.num_chars = splitIndex

	orig.offset += splitIndex
	orig.num_chars -= splitIndex

	return new_item
}

// pango_item_apply_attrs add attributes to an `Item`. The idea is that you have
// attributes that don't affect itemization, such as font features,
// so you filter them out using pango_attr_list_filter(), itemize
// your text, then reapply the attributes to the resulting items
// using this function.
//
// `iter` should be positioned before the range of the item,
// and will be advanced past it. This function is meant to be called
// in a loop over the items resulting from itemization, while passing
// `iter` to each call.
func (item *Item) pango_item_apply_attrs(iter *AttrIterator) {

	compare_attr := func(a1, a2 *Attribute) bool {
		return a1.pango_attribute_equal(*a2) &&
			a1.StartIndex == a2.StartIndex &&
			a1.EndIndex == a2.EndIndex
	}

	var attrs AttrList

	isInList := func(data *Attribute) bool {
		for _, a := range attrs {
			if compare_attr(a, data) {
				return true
			}
		}
		return false
	}

	for do := true; do; do = iter.pango_attr_iterator_next() {
		start, end := iter.StartIndex, iter.EndIndex

		if start >= item.offset+item.num_chars {
			break
		}

		if end >= item.offset {
			list := iter.pango_attr_iterator_get_attrs()
			for _, data := range list {
				if !isInList(data) {
					attrs.insert(0, data.pango_attribute_copy())
				}
			}
		}

		if end >= item.offset+item.num_chars {
			break
		}
	}

	item.analysis.extra_attrs = append(item.analysis.extra_attrs, attrs...)
}

// returns a slice with length item.num_chars
func (item *Item) get_need_hyphen(text []rune) []bool {
	var (
		prevSpace, prevHyphen bool
		attrs                 AttrList
	)
	for _, attr := range item.analysis.extra_attrs {
		if attr.Type == ATTR_INSERT_HYPHENS {
			attrs.pango_attr_list_change(attr.pango_attribute_copy())
		}
	}
	iter := attrs.pango_attr_list_get_iterator()

	needHyphen := make([]bool, item.num_chars)
	for i, wc := range text[item.offset : item.offset+item.num_chars] {
		var start, end int
		insertHyphens := true

		pos := item.offset + i
		for do := true; do; do = iter.pango_attr_iterator_next() {
			start, end = iter.StartIndex, iter.EndIndex
			if end > pos {
				break
			}
		}

		if start <= pos && pos < end {
			attr := iter.pango_attr_iterator_get(ATTR_INSERT_HYPHENS)
			if attr != nil {
				insertHyphens = attr.Data.(AttrInt) == 1
			}

			/* Some scripts don't use hyphen.*/
			switch item.analysis.script {
			case SCRIPT_COMMON, SCRIPT_HAN, SCRIPT_HANGUL, SCRIPT_HIRAGANA, SCRIPT_KATAKANA:
				insertHyphens = false
			}
		}

		space := unicode.In(wc, unicode.Zl, unicode.Zp, unicode.Zs) ||
			(wc == '\t' || wc == '\n' || wc == '\r' || wc == '\f')

		hyphen := wc == '-' || /* Hyphen-minus */
			wc == 0x058a || /* Armenian hyphen */
			wc == 0x1400 || /* Canadian syllabics hyphen */
			wc == 0x1806 || /* Mongolian todo hyphen */
			wc == 0x2010 || /* Hyphen */
			wc == 0x2027 || /* Hyphenation point */
			wc == 0x2e17 || /* Double oblique hyphen */
			wc == 0x2e40 || /* Double hyphen */
			wc == 0x30a0 || /* Katakana-Hiragana double hyphen */
			wc == 0xfe63 || /* Small hyphen-minus */
			wc == 0xff0d /* Fullwidth hyphen-minus */

		if i == 0 {
			needHyphen[i] = false
		} else if prevSpace || space {
			needHyphen[i] = false
		} else if prevHyphen || hyphen {
			needHyphen[i] = false
		} else {
			needHyphen[i] = insertHyphens
		}

		prevSpace = space
		prevHyphen = hyphen
	}

	return needHyphen
}

func (item *Item) find_hyphen_width() GlyphUnit {
	//    hb_font_t *hb_font;
	//    hb_codepoint_t glyph;

	if item.analysis.font == nil {
		return 0
	}

	// This is not technically correct, since
	// a) we may end up inserting a different hyphen
	// b) we should reshape the entire run
	// But it is close enough in practice
	hb_font := pango_font_get_hb_font(item.analysis.font)
	glyph, ok := hb_font_get_nominal_glyph(hb_font, 0x2010)
	if !ok {
		glyph, ok = hb_font_get_nominal_glyph(hb_font, '-')
	}
	if ok {
		return GlyphUnit(hb_font_get_glyph_h_advance(hb_font, glyph))
	}

	return 0
}

func (item *Item) get_item_letter_spacing() GlyphUnit {
	return item.pango_layout_get_item_properties().letter_spacing
}

// Note that rise, letter_spacing, shape are constant across items,
// since we pass them into itemization.
//
// uline and strikethrough can vary across an item, so we collect
// all the values that we find.
type ItemProperties struct {
	uline_single   bool // = 1;
	uline_double   bool // = 1;
	uline_low      bool // = 1;
	uline_error    bool // = 1;
	strikethrough  bool // = 1;
	oline_single   bool // = 1;
	rise           GlyphUnit
	letter_spacing GlyphUnit

	shape *AttrShape // non nil <=> shape_set  =  true
}

func (item *Item) pango_layout_get_item_properties() ItemProperties {
	var properties ItemProperties
	for _, attr := range item.analysis.extra_attrs {
		switch attr.Type {
		case ATTR_UNDERLINE:
			switch Underline(attr.Data.(AttrInt)) {
			case PANGO_UNDERLINE_SINGLE, PANGO_UNDERLINE_SINGLE_LINE:
				properties.uline_single = true
			case PANGO_UNDERLINE_DOUBLE, PANGO_UNDERLINE_DOUBLE_LINE:
				properties.uline_double = true
			case PANGO_UNDERLINE_LOW:
				properties.uline_low = true
			case PANGO_UNDERLINE_ERROR, PANGO_UNDERLINE_ERROR_LINE:
				properties.uline_error = true
			}
		case ATTR_OVERLINE:
			switch Overline(attr.Data.(AttrInt)) {
			case PANGO_OVERLINE_SINGLE:
				properties.oline_single = true
			}
		case ATTR_STRIKETHROUGH:
			properties.strikethrough = attr.Data.(AttrInt) == 1
		case ATTR_RISE:
			properties.rise = GlyphUnit(attr.Data.(AttrInt))
		case ATTR_LETTER_SPACING:
			properties.letter_spacing = GlyphUnit(attr.Data.(AttrInt))
		case ATTR_SHAPE:
			s := attr.Data.(AttrShape)
			properties.shape = &s
		}
	}
	return properties
}
