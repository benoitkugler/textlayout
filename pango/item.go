package pango

import (
	"container/list"
	"log"
	"unicode"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/fribidi"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/benoitkugler/textlayout/language"
)

const (
	// Whether the segment should be shifted to center around the baseline.
	// Used in vertical writing directions mostly.
	AFCenterdBaseline uint8 = 1 << iota
	// Used to mark runs that hold ellipsized text in an ellipsized layout
	AFIsEllipsis
	// Add an hyphen at the end of the run during shaping
	AFNeedHyphen
)

// ItemList is a single linked list of Item elements.
type ItemList struct {
	Data *Item
	Next *ItemList
}

// insert data as second element. panic if l is nil
func (l *ItemList) insertSecond(data *Item) {
	next := l.Next
	l.Next = &ItemList{Data: data, Next: next}
}

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

// copy copy an existing `Item`.
func (item *Item) copy() *Item {
	if item == nil {
		return nil
	}
	result := *item // shallow copy
	result.Analysis.ExtraAttrs = item.Analysis.ExtraAttrs.copy()
	return &result
}

// split modifies `orig` to cover only the text after `splitIndex`,
// which is relative to the start of the item,
// and returns a new item that covers the text before `splitIndex` that
// used to be in `orig`. You can think of `splitIndex` as the length of
// the returned item.
// split returns `nil` if `splitIndex` is 0 or
// greater than or equal to the length of `orig` (that is, there must
// be at least one byte assigned to each item, you can't create a
// zero-length item).
//
// A new item representing text before `splitIndex` is returned.
func (orig *Item) split(splitIndex int) *Item {
	if splitIndex <= 0 || splitIndex >= orig.Length {
		return nil
	}

	newItem := orig.copy()
	newItem.Length = splitIndex

	orig.Offset += splitIndex
	orig.Length -= splitIndex

	return newItem
}

// unSplit undoes the effect of a Item.split() call with
// the same arguments: `splitIndex` is the value passed to split().
func (orig *Item) unSplit(splitIndex int) {
	orig.Offset -= splitIndex
	orig.Length += splitIndex
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
			attrs.Change(attr.copy())
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
	return item.getProperties().letterSpacing
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
	showingSpace  bool
	Rise          GlyphUnit
	letterSpacing GlyphUnit

	lineHeight         Fl
	absoluteLineHeight GlyphUnit
}

func (item *Item) getProperties() itemProperties {
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
		case ATTR_LETTER_SPACING:
			properties.letterSpacing = GlyphUnit(attr.Data.(AttrInt))
		case ATTR_SHAPE:
			s := attr.Data.(AttrShape)
			properties.shape = &s
		case ATTR_LINE_HEIGHT:
			properties.lineHeight = float32(attr.Data.(AttrFloat))
		case ATTR_ABSOLUTE_LINE_HEIGHT:
			properties.absoluteLineHeight = GlyphUnit(attr.Data.(AttrInt))
		case ATTR_SHOW:
			properties.showingSpace = ShowFlags(attr.Data.(AttrInt))&SHOW_SPACES != 0
		}
	}
	return properties
}

type scaleItem struct {
	attr  *Attribute
	scale float32
}

func (context *Context) collectFontScale(stack *list.List, item, prev *Item) (float32, bool) {
	retval := false

	for _, attr := range item.Analysis.ExtraAttrs {
		if attr.Kind == ATTR_FONT_SCALE {
			if attr.StartIndex == item.Offset {
				entry := &scaleItem{attr: attr}
				stack.PushFront(entry)

				var (
					hbFont *harfbuzz.Font
					yScale int32
				)
				if prev != nil {
					hbFont = prev.Analysis.Font.GetHarfbuzzFont()
					yScale = hbFont.YScale
				}

				switch FontScale(attr.Data.(AttrInt)) {
				case FONT_SCALE_NONE:
					// do nothing
				case FONT_SCALE_SUPERSCRIPT:
					entry.scale = 1 / 1.2
					if hbFont != nil {
						if y_size, ok := hbFont.Face().LineMetric(fonts.SuperscriptEmYSize); ok {
							entry.scale = y_size / float32(yScale)
						}
					}
				case FONT_SCALE_SUBSCRIPT:
					entry.scale = 1 / 1.2
					if hbFont != nil {
						if y_size, ok := hbFont.Face().LineMetric(fonts.SubscriptEmYSize); ok {
							entry.scale = y_size / float32(yScale)
						}
					}

				case FONT_SCALE_SMALL_CAPS:
					entry.scale = 0.8
					if hbFont != nil {
						capHeight, ok1 := hbFont.Face().LineMetric(fonts.CapHeight)
						xHeight, ok2 := hbFont.Face().LineMetric(fonts.XHeight)
						if ok1 && ok2 {
							entry.scale = xHeight / float32(capHeight)
						}
					}
				}
			}
		}
	}

	var scale float32 = 1.0

	for l := stack.Front(); l != nil; l = l.Next() {
		entry := l.Value.(*scaleItem)
		scale *= entry.scale
		retval = true
	}

	for l := stack.Front(); l != nil; {
		entry := l.Value.(*scaleItem)
		next := l.Next()

		if entry.attr.EndIndex == item.Offset+item.Length {
			stack.Remove(l)
		}

		l = next
	}

	return scale, retval
}

func (context *Context) applyScaleToItem(item *Item, scale float32) {
	desc := item.Analysis.Font.Describe(false)
	size := scale * float32(desc.Size)

	if desc.SizeIsAbsolute {
		desc.SetAbsoluteSize(int32(size))
	} else {
		desc.SetSize(int32(size))
	}

	item.Analysis.Font = LoadFont(context.fontMap, context, &desc)
}

func (context *Context) applyFontScale(items *ItemList) {
	var (
		prev  *Item
		stack list.List
	)

	for l := items; l != nil; l = l.Next {
		item := l.Data
		if scale, ok := context.collectFontScale(&stack, item, prev); ok {
			context.applyScaleToItem(item, scale)
		}

		prev = item
	}

	if stack.Len() != 0 && debugMode {
		log.Println("Leftover font scales")
	}
}

// Handling Casing variants

func (item *Item) allFeaturesSupported(features []truetype.Tag) bool {
	font := item.Analysis.Font.GetHarfbuzzFont()
	tables := font.GetOTLayoutTables()
	if tables == nil {
		return false
	}

	script := item.Analysis.Script
	language := item.Analysis.Language

	scriptTags, languageTags := harfbuzz.NewOTTagsFromScriptAndLanguage(script, language)

	scriptIndex, _, _ := harfbuzz.SelectScript(&tables.GSUB.TableLayout, scriptTags)
	languageIndex, _ := harfbuzz.SelectLanguage(&tables.GSUB.TableLayout, scriptIndex, languageTags)

	for _, feature := range features {
		if harfbuzz.FindFeatureForLang(&tables.GSUB.TableLayout, scriptIndex, languageIndex, feature) == harfbuzz.NoFeatureIndex {
			return false
		}
	}

	return true
}

func (item *Item) variantSupported(variant Variant) bool {
	var features []truetype.Tag
	switch variant {
	case VARIANT_NORMAL, VARIANT_TITLE_CAPS:
		return true
	case VARIANT_SMALL_CAPS:
		features = []truetype.Tag{harfbuzz.NewOTTag('s', 'm', 'c', 'p')}
	case VARIANT_ALL_SMALL_CAPS:
		features = []truetype.Tag{harfbuzz.NewOTTag('s', 'm', 'c', 'p'), harfbuzz.NewOTTag('c', '2', 's', 'c')}
	case VARIANT_PETITE_CAPS:
		features = []truetype.Tag{harfbuzz.NewOTTag('p', 'c', 'a', 'p')}
	case VARIANT_ALL_PETITE_CAPS:
		features = []truetype.Tag{harfbuzz.NewOTTag('p', 'c', 'a', 'p'), harfbuzz.NewOTTag('c', '2', 'p', 'c')}
	case VARIANT_UNICASE:
		features = []truetype.Tag{harfbuzz.NewOTTag('u', 'n', 'i', 'c')}
	}

	return item.allFeaturesSupported(features)
}

func (item *Item) getFontVariant() Variant {
	return item.Analysis.Font.Describe(false).Variant
}

// Split listItem into upper- and lowercase runs, and
// add font scale and text transform attributes to make
// them be appear according to variant. The logAttrs are
// needed for taking text transforms into account when
// determining the case of characters int he run.
func splitItemForVariant(text []rune, logAttrs []CharAttr, variant Variant, listItem *ItemList) {
	item := listItem.Data
	transform := TEXT_TRANSFORM_NONE
	lowercaseScale := FONT_SCALE_NONE
	uppercaseScale := FONT_SCALE_NONE

	switch variant {
	case VARIANT_ALL_SMALL_CAPS, VARIANT_ALL_PETITE_CAPS:
		uppercaseScale = FONT_SCALE_SMALL_CAPS
		fallthrough
	case VARIANT_SMALL_CAPS, VARIANT_PETITE_CAPS:
		transform = TEXT_TRANSFORM_UPPERCASE
		lowercaseScale = FONT_SCALE_SMALL_CAPS
	case VARIANT_UNICASE:
		uppercaseScale = FONT_SCALE_SMALL_CAPS
	case VARIANT_NORMAL, VARIANT_TITLE_CAPS:
	}

	itemTransform := item.Analysis.findTextTransform()

	p, end := item.Offset, item.Offset+item.Length
	for p < end {
		p0 := p
		isWordStart := logAttrs != nil && logAttrs[p].IsWordStart()
		for ; p < end; p++ {
			if wc := text[p]; !(itemTransform == TEXT_TRANSFORM_LOWERCASE || considerAsSpace(wc) ||
				(unicode.IsLower(wc) &&
					!(itemTransform == TEXT_TRANSFORM_UPPERCASE ||
						(itemTransform == TEXT_TRANSFORM_CAPITALIZE && isWordStart)))) {
				break
			}
			isWordStart = logAttrs != nil && logAttrs[p].IsWordStart()
		}

		if p0 < p {
			newItem := item

			/* p0 .. p is a lowercase segment */
			if p < end {
				newItem = item.split(p - p0)
				listItem.Data = newItem
				listItem.insertSecond(item)
				listItem = listItem.Next
			}

			if transform != TEXT_TRANSFORM_NONE {
				attr := NewAttrTextTransform(transform)
				attr.StartIndex = newItem.Offset
				attr.EndIndex = newItem.Offset + newItem.Length
				newItem.Analysis.ExtraAttrs = append(newItem.Analysis.ExtraAttrs, attr)
			}

			if lowercaseScale != FONT_SCALE_NONE {
				attr := NewAttrFontScale(lowercaseScale)
				attr.StartIndex = newItem.Offset
				attr.EndIndex = newItem.Offset + newItem.Length
				newItem.Analysis.ExtraAttrs = append(newItem.Analysis.ExtraAttrs, attr)
			}
		}

		p0 = p
		isWordStart = logAttrs != nil && logAttrs[p].IsWordStart()
		for ; p < end; p++ {
			if wc := text[p]; !(itemTransform == TEXT_TRANSFORM_UPPERCASE ||
				considerAsSpace(wc) ||
				!(itemTransform == TEXT_TRANSFORM_LOWERCASE || unicode.IsLower(wc)) ||
				(itemTransform == TEXT_TRANSFORM_CAPITALIZE && isWordStart)) {
				break
			}
			isWordStart = logAttrs != nil && logAttrs[p].IsWordStart()
		}

		if p0 < p {
			newItem := item

			/* p0 .. p is a uppercase segment */
			if p < end {
				newItem = item.split(p - p0)
				listItem.Data = newItem
				listItem.insertSecond(item)
				listItem = listItem.Next
			}

			if uppercaseScale != FONT_SCALE_NONE {
				attr := NewAttrFontScale(uppercaseScale)
				attr.StartIndex = newItem.Offset
				attr.EndIndex = newItem.Offset + newItem.Length
				newItem.Analysis.ExtraAttrs = append(newItem.Analysis.ExtraAttrs, attr)
			}
		}
	}
}

func handleVariantsForItem(text []rune, logAttrs []CharAttr, l *ItemList) {
	item := l.Data

	variant := item.getFontVariant()
	if !item.variantSupported(variant) {
		splitItemForVariant(text, logAttrs, variant, l)
	}
}

func handleVariants(text []rune, logAttrs []CharAttr, items *ItemList) {
	var next *ItemList

	for l := items; l != nil; l = next {
		next = l.Next
		handleVariantsForItem(text, logAttrs, l)
	}
}

// reverse `arr` and convert it to a linked list
func reverseItemsToList(arr []*Item) *ItemList {
	var out *ItemList
	for _, v := range arr {
		out = &ItemList{Data: v, Next: out}
	}
	return out
}

func (context *Context) postProcessItems(text []rune, logAttrs []CharAttr, items *ItemList) *ItemList {
	handleVariants(text, logAttrs, items)

	/* apply font-scale */
	context.applyFontScale(items)

	return items
}
