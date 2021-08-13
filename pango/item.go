package pango

import (
	"unicode"

	"github.com/benoitkugler/textlayout/fribidi"
	"github.com/benoitkugler/textlayout/language"
)

const (
	// Whether the segment should be shifted to center around the baseline.
	// Used in vertical writing directions mostly.
	AFCenterdBaseline = 1 << iota
	// Used to mark runs that hold ellipsized text in an ellipsized layout
	AFIsEllipsis
	// Add an hyphen at the end of the run during shaping
	AFNeedHyphen
)

// Analysis stores information about the properties of a segment of text.
type Analysis struct {
	Font Font // the font for this segment.

	Language   Language // the detected language for this segment.
	ExtraAttrs AttrList // extra attributes for this segment.
	Script     Script   // the detected script for this segment

	Level   fribidi.Level //  the bidirectional level for this segment.
	Gravity Gravity       //  the glyph orientation for this segment.
	Flags   uint8         //  boolean flags for this segment.
}

func (analysis *Analysis) showing_space() bool {
	for _, attr := range analysis.ExtraAttrs {
		if attr.Kind == ATTR_SHOW && (ShowFlags(attr.Data.(AttrInt))&SHOW_SPACES != 0) {
			return true
		}
	}
	return false
}

// Item stores information about a segment of text.
type Item struct {
	Analysis Analysis // Analysis results for the item.
	Offset   int      // Rune offset of the start of this item in text.
	Length   int      // Number of runes in the item.
}

// pango_item_copy copy an existing `Item`.
func (item *Item) pango_item_copy() *Item {
	if item == nil {
		return nil
	}
	result := *item // shallow copy
	result.Analysis.ExtraAttrs = item.Analysis.ExtraAttrs.pango_attr_list_copy()
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
	if splitIndex <= 0 || splitIndex >= orig.Length {
		return nil
	}

	new_item := orig.pango_item_copy()
	new_item.Length = splitIndex

	orig.Offset += splitIndex
	orig.Length -= splitIndex

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
func (item *Item) pango_item_apply_attrs(iter *attrIterator) {
	compare_attr := func(a1, a2 *Attribute) bool {
		return a1.equals(*a2) &&
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

	for do := true; do; do = iter.next() {
		start, end := iter.StartIndex, iter.EndIndex

		if start >= item.Offset+item.Length {
			break
		}

		if end >= item.Offset {
			list := iter.getAttributes()
			for _, data := range list {
				if !isInList(data) {
					attrs.insertAt(0, data.copy())
				}
			}
		}

		if end >= item.Offset+item.Length {
			break
		}
	}

	item.Analysis.ExtraAttrs = append(item.Analysis.ExtraAttrs, attrs...)
}

// returns a slice with length item.num_chars
func (item *Item) get_need_hyphen(text []rune) []bool {
	var attrs AttrList
	for _, attr := range item.Analysis.ExtraAttrs {
		if attr.Kind == ATTR_INSERT_HYPHENS {
			attrs.pango_attr_list_change(attr.copy())
		}
	}
	iter := attrs.getIterator()

	prevSpace, prevHyphen := true, true

	needHyphen := make([]bool, item.Length)
	for i, wc := range text[item.Offset : item.Offset+item.Length] {
		var start, end int
		insertHyphens := true

		pos := item.Offset + i
		for do := true; do; do = iter.next() {
			start, end = iter.StartIndex, iter.EndIndex
			if end > pos {
				break
			}
		}

		if start <= pos && pos < end {
			attr := iter.getByKind(ATTR_INSERT_HYPHENS)
			if attr != nil {
				insertHyphens = attr.Data.(AttrInt) == 1
			}

			/* Some scripts don't use hyphen.*/
			switch item.Analysis.Script {
			case language.Common, language.Han, language.Hangul, language.Hiragana, language.Katakana:
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

		if prevSpace || space {
			needHyphen[i] = false
		} else if prevHyphen || hyphen {
			needHyphen[i] = false
		} else {
			needHyphen[i] = insertHyphens
		}

		prevSpace = space
		prevHyphen = hyphen
	}

	needHyphen[item.Length-1] = false

	return needHyphen
}

func (item *Item) find_hyphen_width() GlyphUnit {
	if item.Analysis.Font == nil {
		return 0
	}

	// This is not technically correct, since
	// a) we may end up inserting a different hyphen
	// b) we should reshape the entire run
	// But it is close enough in practice
	hbFont := item.Analysis.Font.GetHarfbuzzFont()
	glyph, ok := hbFont.Face().NominalGlyph(0x2010)
	if !ok {
		glyph, ok = hbFont.Face().NominalGlyph('-')
	}
	if ok {
		return GlyphUnit(hbFont.GlyphHAdvance(glyph))
	}

	return 0
}

func (item *Item) get_item_letter_spacing() GlyphUnit {
	return item.pango_layout_get_item_properties().letterSpacing
}

// Note that rise, letter_spacing, shape are constant across items,
// since we pass them into itemization.
//
// uline and strikethrough can vary across an item, so we collect
// all the values that we find.
type itemProperties struct {
	shape *AttrShape // non nil <=> shape_set  =  true

	ulineSingle   bool // = 1;
	ulineDouble   bool // = 1;
	ulineLow      bool // = 1;
	ulineError    bool // = 1;
	strikethrough bool // = 1;
	olineSingle   bool // = 1;
	Rise          GlyphUnit
	letterSpacing GlyphUnit

	lineHeight         Fl
	absoluteLineHeight GlyphUnit
}

func (item *Item) pango_layout_get_item_properties() itemProperties {
	var properties itemProperties
	for _, attr := range item.Analysis.ExtraAttrs {
		switch attr.Kind {
		case ATTR_UNDERLINE:
			switch Underline(attr.Data.(AttrInt)) {
			case UNDERLINE_SINGLE, UNDERLINE_SINGLE_LINE:
				properties.ulineSingle = true
			case UNDERLINE_DOUBLE, UNDERLINE_DOUBLE_LINE:
				properties.ulineDouble = true
			case UNDERLINE_LOW:
				properties.ulineLow = true
			case UNDERLINE_ERROR, UNDERLINE_ERROR_LINE:
				properties.ulineError = true
			}
		case ATTR_OVERLINE:
			switch Overline(attr.Data.(AttrInt)) {
			case OVERLINE_SINGLE:
				properties.olineSingle = true
			}
		case ATTR_STRIKETHROUGH:
			properties.strikethrough = attr.Data.(AttrInt) == 1
		case ATTR_RISE:
			properties.Rise = GlyphUnit(attr.Data.(AttrInt))
		case ATTR_LETTER_SPACING:
			properties.letterSpacing = GlyphUnit(attr.Data.(AttrInt))
		case ATTR_SHAPE:
			s := attr.Data.(AttrShape)
			properties.shape = &s
		case ATTR_LINE_HEIGHT:
			properties.lineHeight = float32(attr.Data.(AttrFloat))
		case ATTR_ABSOLUTE_LINE_HEIGHT:
			properties.absoluteLineHeight = GlyphUnit(attr.Data.(AttrInt))
		}
	}
	return properties
}
