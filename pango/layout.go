package pango

import (
	"fmt"
	"log"
	"unicode/utf8"
)

/**
 * While complete access to the layout capabilities of Pango is provided
 * using the detailed interfaces for itemization and shaping, using
 * that functionality directly involves writing a fairly large amount
 * of code. The objects and functions in this section provide a
 * high-level driver for formatting entire paragraphs of text
 * at once. This includes paragraph-level functionality such as
 * line-breaking, justification, alignment and ellipsization.
 */

// Only one character has type G_UNICODE_LINE_SEPARATOR in Unicode 5.0: update this if that changes.
const lineSeparator = 0x2028

// WrapMode Describes how to wrap the lines of a `Layout` to the desired width.
type WrapMode uint8

const (
	WRAP_WORD      WrapMode = iota // wrap lines at word boundaries.
	WRAP_CHAR                      // wrap lines at character boundaries.
	WRAP_WORD_CHAR                 // wrap lines at word boundaries, but fall back to character boundaries if there is not enough space for a full word.
)

// Layout represents an entire paragraph
// of text. It is initialized with a `Context`, UTF-8 string
// and set of attributes for that string. Once that is done, the
// set of formatted lines can be extracted from the object,
// the layout can be rendered, and conversion between logical
// character positions within the layout's text, and the physical
// position of the resulting glyphs can be made.
//
// There are also a number of parameters to adjust the formatting
// of a `Layout`.
// It is possible, as well, to ignore the 2-D setup, and simply
// treat the results of a `Layout` as a list of lines.
type Layout struct {
	// Attributes is the attribute list for the layout.
	// This is a readonly property, see `SetAttributes` to modify it.
	Attributes AttrList

	/* Referenced items */
	context *Context

	fontDesc *FontDescription
	tabs     *tabArray

	// logical attributes for layout's text, allocated
	// once in check_lines; has length len(text)+1
	logAttrs []CharAttr

	lines []*LayoutLine // with length lines_count

	// Readonly text of the layout, see `SetText` to modify it.
	Text []rune

	serial        uint
	contextSerial uint

	// Width (in Pango units) to which the lines of the Layout should wrap or ellipsize.
	// or -1 if not set.
	// This is a readonly property, see `SetWidth` to modify it.
	Width GlyphUnit

	// Ellipsize height, in device units if positive, number of lines if negative.
	// This is a readonly property, see `SetHeight` to modify it.
	Height GlyphUnit

	// Amount by which first line should be shorter.
	// This is a readonly property, see `SetIndent` to modify it.
	Indent GlyphUnit

	// Spacing between lines.
	// This is a readonly property, see `SetSpacing` to modify it.
	Spacing GlyphUnit

	// Factor to apply to line height.
	// This is a readonly property, see `SetLineSpacing` to modify it.
	LineSpacing float32

	// Whether each complete line should be stretched to fill the entire
	// width of the layout.
	// This is a readonly property, see `SetJustify` to modify it.
	Justify bool

	// Whether the last line should be stretched to fill the
	// entire width of the layout.
	// This is a readonly property, see `SetJustifyLastLine` to modify it.
	JustifyLastLine bool

	alignment       Alignment
	singleParagraph bool
	autoDir         bool
	wrap            WrapMode
	isWrapped       bool          // Whether the layout has any wrapped lines
	ellipsize       EllipsizeMode // PangoEllipsizeMode
	isEllipsized    bool          // Whether the layout has any ellipsized lines
	// unknownGlyphsCount int           // number of unknown glyphs

	// some caching

	logicalRectCached    bool // = true
	inkRectCached        bool // = true
	inkRect, logicalRect Rectangle
	tabWidth             GlyphUnit /* Cached width of a tab. -1 == not yet calculated */

}

// NewLayout creates a new `Layout` object with attributes initialized to
// default values for a particular `Context`.
func NewLayout(context *Context) *Layout {
	var layout Layout
	layout.context = context
	layout.contextSerial = context.pango_context_get_serial()

	layout.serial = 1
	layout.Width = -1
	layout.Height = -1

	layout.autoDir = true

	layout.tabWidth = -1
	// layout.unknownGlyphsCount = -1

	return &layout
}

// SetFontDescription sets the default font description for the layout.
// If `nil`, the font description from the layout's context is used.
func (layout *Layout) SetFontDescription(desc *FontDescription) {
	if desc != nil && layout.fontDesc != nil && desc.pango_font_description_equal(*layout.fontDesc) {
		return
	}

	layout.fontDesc = nil
	if desc != nil {
		cp := *desc
		layout.fontDesc = &cp
	}

	layout.layoutChanged()
	layout.tabWidth = -1
}

// SetText sets the text of the layout, validating `text` and rendering invalid UTF-8
// with a placeholder glyph.
//
// Note that if you have used SetMarkup() on `layout` before, you may
// want to call SetAttributes() to clear the attributes
// set on the layout from the markup as this function does not clear
// attributes.
func (layout *Layout) SetText(text string) {
	layout.Text = layout.Text[:0]
	b := []byte(text)
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		b = b[size:]
		layout.Text = append(layout.Text, r)
	}

	layout.layoutChanged()
}

// SetRunes is the same as `SetText` but skips
// UTF-8 validation. `text` is copied and not retained
// by the layout.
func (layout *Layout) SetRunes(text []rune) {
	L := len(text)
	if cap(layout.Text) < L {
		layout.Text = make([]rune, L)
	}
	layout.Text = layout.Text[:L]
	copy(layout.Text, text)

	layout.layoutChanged()
}

// SetMarkup sets the layout text and attribute list from marked-up text (see
// `ParseMarkup()` for the markup format). It replaces the current text and attribute list.
// It returns an error if the markup is not correctly formatted.
func (layout *Layout) SetMarkup(markup []byte) error {
	_, err := layout.setMarkupWithAccel(markup, 0)
	return err
}

// Sets the layout text and attribute list from marked-up text (see
// the markup format). Replaces
// the current text and attribute list.
//
// If `accelMarker` is nonzero, the given character will mark the
// character following it as an accelerator. For example, `accelMarker`
// might be an ampersand or underscore. All characters marked
// as an accelerator will receive a PANGO_UNDERLINE_LOW attribute,
// and the first character so marked will be returned.
// Two `accelMarker` characters following each other produce a single
// literal `accelMarker` character.
func (layout *Layout) setMarkupWithAccel(markup []byte, accelMarker rune) (rune, error) {
	parsed, err := ParseMarkup(markup, accelMarker)
	if err != nil {
		return 0, err
	}

	layout.SetRunes(parsed.Text)
	layout.SetAttributes(parsed.Attr)

	return parsed.AccelChar, nil
}

// SetSpacing sets the amount of spacing in Pango unit between
// the lines of the layout. When placing lines with
// spacing, Pango arranges things so that
//
// line2.top = line1.bottom + spacing
//
// Note: Pango defaults to using the line height (as determined by the font) for placing
// lines. The `spacing` set with this function is only
// taken into account when the line-height factor is
// set to zero with SetLineSpacing().
func (layout *Layout) SetSpacing(spacing GlyphUnit) {
	if spacing != layout.Spacing {
		layout.Spacing = spacing
		layout.layoutChanged()
	}
}

// SetLineSpacing sets a factor for line spacing.
// Typical values are: 0, 1, 1.5, 2.
// The default values is 0.
//
// If `factor` is non-zero, lines are placed
// so that
//
// baseline2 = baseline1 + factor * height2
//
// where height2 is the line height of the
// second line (as determined by the font(s)).
// In this case, the spacing set with
// SetSpacing() is ignored.
//
// If `factor` is zero, spacing is applied as
// before.
func (layout *Layout) SetLineSpacing(factor float32) {
	if factor != layout.LineSpacing {
		layout.LineSpacing = factor
		layout.layoutChanged()
	}
}

// SetWidth sets the width (in Pango units) to which the lines of the layout should wrap or
// ellipsize.
// Pass -1 to indicate that no wrapping or ellipsization should be performed.
// The default value is -1: no width set.
func (layout *Layout) SetWidth(width GlyphUnit) {
	if width < 0 {
		width = -1
	}

	if width != layout.Width {
		layout.Width = width

		layout.layoutChanged()
	}
}

// SetHeight sets the height to which the `Layout` should be ellipsized at.
//
// There are two different behaviors, based on whether `height` is positive
// or negative.
//
// If `height` is positive, it will be the maximum height of the layout. Only
// lines would be shown that would fit, and if there is any text omitted,
// an ellipsis added. At least one line is included in each paragraph regardless
// of how small the height value is. A value of zero will render exactly one
// line for the entire layout.
//
// If `height` is negative, it will be the (negative of the) maximum number of lines
// per paragraph. That is, the total number of lines shown may well be more than
// this value if the layout contains multiple paragraphs of text.
// The default value of -1 means that the first line of each paragraph is ellipsized.
// This behavior may be changed in the future to act per layout instead of per
// paragraph.
//
// Height setting only has effect if a positive width is set on
// `layout` and ellipsization mode of `layout` is not `ELLIPSIZE_NONE`.
// The behavior is undefined if a height other than -1 is set and
// ellipsization mode is set to `ELLIPSIZE_NONE`, and may change in the
// future.
func (layout *Layout) SetHeight(height GlyphUnit) {
	if height != layout.Height {
		layout.Height = height

		/* Do not invalidate if the number of lines requested is
		* larger than the total number of lines in layout.
		* Bug 549003
		 */
		if layout.ellipsize != ELLIPSIZE_NONE &&
			!(layout.isEllipsized == false && height < 0 && len(layout.lines) <= -int(height)) {
			layout.layoutChanged()
		}
	}
}

// SetIndent sets the width in Pango units to indent each paragraph. A negative value
// of `indent` will produce a hanging indentation. That is, the first line will
// have the full width, and subsequent lines will be indented by the
// absolute value of `indent`.
//
// The indent setting is ignored if layout alignment is set to
// `ALIGN_CENTER`.
func (layout *Layout) SetIndent(indent GlyphUnit) {
	if indent != layout.Indent {
		layout.Indent = indent
		layout.layoutChanged()
	}
}

// SetAlignment sets the alignment for the layout: how partial lines are
// positioned within the horizontal space available.
func (layout *Layout) SetAlignment(alignment Alignment) {
	if alignment != layout.alignment {
		layout.alignment = alignment
		layout.layoutChanged()
	}
}

// SetEllipsize sets the type of ellipsization being performed for layout.
// Depending on the ellipsization mode `ellipsize`, text is
// removed from the start, middle, or end of text so they
// fit within the width and height of layout set with
// SetWidth() and SetHeight().
//
// If the layout contains characters such as newlines that
// force it to be layed out in multiple paragraphs, then whether
// each paragraph is ellipsized separately or the entire layout
// is ellipsized as a whole depends on the set height of the layout.
// See SetHeight() for details.
func (layout *Layout) SetEllipsize(ellipsize EllipsizeMode) {
	if ellipsize != layout.ellipsize {
		layout.ellipsize = ellipsize

		if layout.isEllipsized || layout.isWrapped {
			layout.layoutChanged()
		}
	}
}

// SetWrap sets the wrap mode; the wrap mode only has effect if a width
// is set on the layout with SetWidth().
// To turn off wrapping, set the width to -1.
func (layout *Layout) SetWrap(wrap WrapMode) {
	if layout.wrap != wrap {
		layout.wrap = wrap

		if layout.Width != -1 {
			layout.layoutChanged()
		}
	}
}

// SetJustify sets whether each complete line should be stretched to
// fill the entire width of the layout. This stretching is typically
// done by adding whitespace, but for some scripts (such as Arabic),
// the justification may be done in more complex ways, like extending
// the characters.
// Note that tabs and justification conflict with each other:
// Justification will move content away from its tab-aligned
// positions.
func (layout *Layout) SetJustify(justify bool) {
	if justify != layout.Justify {
		layout.Justify = justify

		if layout.isEllipsized || layout.isWrapped || layout.JustifyLastLine {
			layout.layoutChanged()
		}
	}
}

// SetJustifyLastLine sets whether the last line should be stretched to fill the
// entire width of the layout.
//
// This only has an effect if SetJustify has
// been called as well.
func (layout *Layout) SetJustifyLastLine(justify bool) {
	if justify != layout.JustifyLastLine {
		layout.JustifyLastLine = justify

		if layout.Justify {
			layout.layoutChanged()
		}
	}
}

// SetAutoDir sets whether to calculate the bidirectional base direction
// for the layout according to the contents of the layout;
// when this flag is on (the default), then paragraphs in
// `layout` that begin with strong right-to-left characters
// (Arabic and Hebrew principally), will have right-to-left
// layout, paragraphs with letters from other scripts will
// have left-to-right layout. Paragraphs with only neutral
// characters get their direction from the surrounding paragraphs.
//
// When `false`, the choice between left-to-right and
// right-to-left layout is done according to the base direction
// of the layout's Context. (See pango_context_set_base_dir()).
//
// When the auto-computed direction of a paragraph differs from the
// base direction of the context, the interpretation of
// `ALIGN_LEFT` and `ALIGN_RIGHT` are swapped.
func (layout *Layout) SetAutoDir(autoDir bool) {
	if autoDir != layout.autoDir {
		layout.autoDir = autoDir
		layout.layoutChanged()
	}
}

// SetSingleParagraphMode set the paragraph behavior.
//
// If `setting` is `true`, do not treat newlines and similar characters
// as paragraph separators; instead, keep all text in a single paragraph,
// and display a glyph for paragraph separator characters. Used when
// you want to allow editing of newlines on a single text line.
func (layout *Layout) SetSingleParagraphMode(setting bool) {
	if layout.singleParagraph != setting {
		layout.singleParagraph = setting
		layout.layoutChanged()
	}
}

// GetLineCount retrieves the count of lines for the layout,
// triggering a layout if needed.
func (layout *Layout) GetLineCount() int {
	if layout == nil {
		return 0
	}

	layout.checkLines()
	return len(layout.lines)
}

// GetCharacterAttributes retrieves an array of logical attributes for each character in
// the `layout`, triggering a layout if needed.
//
// The returned slice length will be one more than the total number
// of characters in the layout, since there need to be attributes
// corresponding to both the position before the first character
// and the position after the last character.
//
// The returned slice is retained by the layout and must to be modified.
func (layout *Layout) GetCharacterAttributes() []CharAttr {
	layout.checkLines()
	return layout.logAttrs
}

// IsEllipsized queries whether the layout had to ellipsize any paragraphs,
// triggering a layout if needed.
//
// This returns `true` if the ellipsization mode for `layout`
// is not ELLIPSIZE_NONE, a positive width is set on `layout`,
// and there are paragraphs exceeding that width that have to be
// ellipsized.
func (layout *Layout) IsEllipsized() bool {
	layout.checkLines()
	return layout.isEllipsized
}

// IsWrapped queries whether the layout had to wrap any paragraphs,
// triggering a layout if needed.
//
// This returns `true` if a positive width is set on `layout`,
// ellipsization mode of `layout` is set to ELLIPSIZE_NONE,
// and there are paragraphs exceeding the layout width that have
// to be wrapped.
func (layout *Layout) IsWrapped() bool {
	layout.checkLines()
	return layout.isWrapped
}

func (layout *Layout) indexToLine(index int) (lineNr int, line, lineBefore, lineAfter *LayoutLine) {
	var i int
	for i = 0; i < len(layout.lines); {
		currentLine := layout.lines[i]

		if currentLine.StartIndex > index {
			break /* index was in paragraph delimiters */
		}

		i++

		if currentLine.StartIndex+currentLine.Length > index {
			break
		}
	}

	lineNr = i - 1
	if i-2 >= 0 {
		lineBefore = layout.lines[i-2]
	}
	if i < len(layout.lines) {
		lineAfter = layout.lines[i]
	}
	if lineNr >= 0 {
		line = layout.lines[lineNr]
	}

	return
}

// indexToLineX converts from run `index` within the `layout` to line and X position.
// (X position is measured from the left edge of the line)
// `trailing` indicates the edge of the grapheme to retrieve the
// position of: if true, the trailing edge of the grapheme, else,
// the leading of the grapheme.
func (layout *Layout) indexToLineX(index int, trailing bool) (line int, xPos GlyphUnit) {
	if index < 0 || index > len(layout.Text) {
		return 0, 0
	}

	layout.checkLines()

	lineNum, layoutLine, _, _ := layout.indexToLine(index)
	if layoutLine != nil {
		/* use end of line if index was in the paragraph delimiters */
		if index > layoutLine.StartIndex+layoutLine.Length {
			index = layoutLine.StartIndex + layoutLine.Length
		}

		xPos = layoutLine.IndexToX(index, trailing)
		return lineNum, xPos
	}

	return -1, -1
}

// GetDirectionAt gets the text direction at the given character
// position in `layout` text.
func (layout *Layout) GetDirectionAt(index int) Direction {
	line, _, _ := layout.indexToLineAndExtents(index)

	if line != nil {
		return line.getCharDirection(index)
	}

	return DIRECTION_LTR
}

func (layout *Layout) indexToLineAndExtents(index int) (line *LayoutLine, lineRect, runRect Rectangle) {
	iter := layout.GetIter()

	if len(iter.lines) == 0 {
		return
	}

	for {
		tmpLine := iter.line()

		if tmpLine.StartIndex > index {
			break /* index was in paragraph delimiters */
		}

		line = tmpLine

		iter.GetLineExtents(nil, &lineRect)

		if iter.lineIndex+1 >= len(iter.lines) || (iter.lines[iter.lineIndex+1]).StartIndex > index {
			runRect = lineRect

			for {
				run := iter.run

				if run == nil {
					break
				}

				iter.GetRunExtents(nil, &runRect)

				if run.Item.Offset <= index && index < run.Item.Offset+run.Item.Length {
					break
				}

				if !iter.NextRun() {
					break
				}
			}

			break
		}

		if !iter.NextLine() {
			break /* Use end of last line */
		}
	}

	return line, lineRect, runRect
}

func (layout *Layout) check_context_changed() {
	old_serial := layout.contextSerial

	layout.contextSerial = layout.context.pango_context_get_serial()

	if old_serial != layout.contextSerial {
		layout.pango_layout_context_changed()
	}
}

// Forces recomputation of any state in the `Layout` that
// might depend on the layout's context. This function should
// be called if you make changes to the context subsequent
// to creating the layout.
func (layout *Layout) pango_layout_context_changed() {
	if layout == nil {
		return
	}

	layout.layoutChanged()
	layout.tabWidth = -1
}

func (layout *Layout) layoutChanged() {
	layout.serial++
	if layout.serial == 0 {
		layout.serial++
	}
	layout.clearLines()
}

func (layout *Layout) clearLines() {
	// TODO: we could keep the underlying arrays
	layout.lines = nil
	layout.logAttrs = nil
	// layout.unknownGlyphsCount = -1
	layout.logicalRectCached = false
	layout.inkRectCached = false
	layout.isEllipsized = false
	layout.isWrapped = false
}

// Sets the text attributes for a layout object.
func (layout *Layout) SetAttributes(attrs AttrList) {
	if layout == nil {
		return
	}

	if layout.Attributes.pango_attr_list_equal(attrs) {
		return
	}

	// We always clear lines such that this function can be called
	// whenever attrs changes.
	layout.Attributes = attrs
	layout.layoutChanged()
	layout.tabWidth = -1
}

// Sets the tabs to use for `layout`, overriding the default tabs
// (by default, tabs are every 8 spaces). If tabs is nil, the default
// tabs are reinstated.
func (layout *Layout) pango_layout_set_tabs(tabs *tabArray) {
	if layout == nil {
		return
	}

	if tabs != layout.tabs {
		layout.tabs = tabs.pango_tab_array_copy()
		layout.layoutChanged()
	}
}

func affects_break_or_shape(attr *Attribute) bool {
	switch attr.Kind {
	/* Affects breaks */
	case ATTR_ALLOW_BREAKS, ATTR_WORD, ATTR_SENTENCE:
		return true
	/* Affects shaping */
	case ATTR_INSERT_HYPHENS, ATTR_FONT_FEATURES, ATTR_SHOW:
		return true
	default:
		return false
	}
}

func affects_itemization(attr *Attribute) bool {
	switch attr.Kind {
	/* These affect font selection */
	case ATTR_LANGUAGE, ATTR_FAMILY, ATTR_STYLE, ATTR_WEIGHT, ATTR_VARIANT, ATTR_STRETCH,
		ATTR_SIZE, ATTR_FONT_DESC, ATTR_SCALE, ATTR_FALLBACK, ATTR_ABSOLUTE_SIZE, ATTR_GRAVITY,
		ATTR_GRAVITY_HINT:
		return true
	/* These need to be constant across runs */
	case ATTR_LETTER_SPACING, ATTR_SHAPE, ATTR_RISE, ATTR_LINE_HEIGHT, ATTR_ABSOLUTE_LINE_HEIGHT, ATTR_TEXT_TRANSFORM:
		return true
	default:
		return false
	}
}

// if withLine is true, returns a list of line extents in layout coordinates, else returns nil
func (layout *Layout) getExtentsInternal(inkRect, logicalRect *Rectangle, withLine bool) (lines []extents) {
	layout.checkLines()
	if inkRect != nil && layout.inkRectCached {
		*inkRect = layout.inkRect
		inkRect = nil // mark as filled
	}
	if logicalRect != nil && layout.logicalRectCached {
		*logicalRect = layout.logicalRect
		logicalRect = nil // mark as filled
	}
	if inkRect == nil && logicalRect == nil && !withLine {
		return nil
	}

	/* When we are not wrapping, we need the overall width of the layout to
	* figure out the x_offsets of each line. However, we only need the
	* x_offsets if we are computing the inkRect or individual line extents.*/
	width := layout.Width

	var needWidth bool
	if layout.autoDir {
		/* If one of the lines of the layout is not left aligned, then we need
		* the width of the layout to calculate line x-offsets; this requires
		* looping through the lines for layout.autoDir. */
		for _, line := range layout.lines {
			if line.getAlignment() != ALIGN_LEFT {
				needWidth = true
				break
			}
		}
	} else if layout.alignment != ALIGN_LEFT {
		needWidth = true
	}

	if width == -1 && needWidth && (inkRect != nil || withLine) {
		var overallLogical Rectangle
		layout.getExtentsInternal(nil, &overallLogical, false)
		width = GlyphUnit(overallLogical.Width)
	}

	if logicalRect != nil {
		*logicalRect = Rectangle{}
	}

	if withLine { // avoid allocation when not asked
		lines = make([]extents, len(layout.lines))
	}

	var baseline, yOffset GlyphUnit
	for lineIndex, line := range layout.lines {
		// Line extents in layout coords (origin at 0,0 of the layout)
		var (
			lineInkLayout, lineLogicalLayout Rectangle
			newPos                           GlyphUnit
		)
		// This block gets the line extents in layout coords
		{
			var ptr *Rectangle
			if inkRect != nil {
				ptr = &lineInkLayout
			}
			line.getLineExtentsLayoutCoords(layout, width, yOffset,
				&baseline, ptr, &lineLogicalLayout)

			if withLine {
				ext := &lines[lineIndex]
				ext.baseline = baseline
				ext.inkRect = lineInkLayout
				ext.logicalRect = lineLogicalLayout
			}
		}

		if inkRect != nil {
			/* Compute the union of the current inkRect with
			* lineInkLayout  */

			if lineIndex == 0 {
				*inkRect = lineInkLayout
			} else {
				newPos = minG(inkRect.X, lineInkLayout.X)
				inkRect.Width = maxG(inkRect.X+inkRect.Width, lineInkLayout.X+lineInkLayout.Width) - newPos
				inkRect.X = newPos

				newPos = minG(inkRect.Y, lineInkLayout.Y)
				inkRect.Height = maxG(inkRect.Y+inkRect.Height, lineInkLayout.Y+lineInkLayout.Height) - newPos
				inkRect.Y = newPos
			}
		}

		if logicalRect != nil {
			if layout.Width == -1 {
				/* When no width is set on layout, we can just compute the max of the
				* line lengths to get the horizontal extents ... logicalRect.x = 0. */
				logicalRect.Width = maxG(logicalRect.Width, lineLogicalLayout.Width)
			} else {
				/* When a width is set, we have to compute the union of the horizontal
				* extents of all the lines. */
				if lineIndex == 0 {
					logicalRect.X = lineLogicalLayout.X
					logicalRect.Width = lineLogicalLayout.Width
				} else {
					newPos = minG(logicalRect.X, lineLogicalLayout.X)
					logicalRect.Width = maxG(logicalRect.X+logicalRect.Width,
						lineLogicalLayout.X+lineLogicalLayout.Width) - newPos
					logicalRect.X = newPos

				}
			}

			logicalRect.Height = lineLogicalLayout.Y + lineLogicalLayout.Height - logicalRect.Y
		}

		yOffset = lineLogicalLayout.Y + lineLogicalLayout.Height + layout.Spacing
	}

	if inkRect != nil {
		layout.inkRect = *inkRect
		layout.inkRectCached = true
	}
	if logicalRect != nil {
		layout.logicalRect = *logicalRect
		layout.logicalRectCached = true
	}

	return lines
}

// GetExtents computes the ink and logical extents of `layout`. Logical extents
// are usually what you want for positioning things. Note that both extents
// may have non-zero x and y. You may want to use those to offset where you
// render the layout. Not doing that is a very typical bug that shows up as
// right-to-left layouts not being correctly positioned in a layout with
// a set width.
//
// The extents are given in layout coordinates and in Pango units; layout
// coordinates begin at the top left corner of the layout.
// Pass `nil` if you dont need one of the extents.
func (layout *Layout) GetExtents(inkRect, logicalRect *Rectangle) {
	layout.getExtentsInternal(inkRect, logicalRect, false)
}

func (layout *Layout) pango_layout_get_effective_attributes() AttrList {
	var attrs AttrList

	if len(layout.Attributes) != 0 {
		attrs = layout.Attributes.pango_attr_list_copy()
	}

	if layout.fontDesc != nil {
		attr := NewAttrFontDescription(*layout.fontDesc)
		attrs.pango_attr_list_insert_before(attr)
	}

	if layout.singleParagraph {
		attr := NewAttrShow(SHOW_LINE_BREAKS)
		attrs.pango_attr_list_insert_before(attr)
	}

	return attrs
}

func (layout *Layout) getEmptyExtentsAndHeightAt(index int, logicalRect *Rectangle) (height GlyphUnit) {
	if logicalRect == nil {
		return
	}

	fontDesc := layout.context.fontDesc // copy

	if layout.fontDesc != nil {
		fontDesc.pango_font_description_merge(layout.fontDesc, true)
	}

	// Find the font description for this line
	if len(layout.Attributes) != 0 {
		iter := layout.Attributes.getIterator()
		for hasNext := true; hasNext; hasNext = iter.next() {
			if iter.StartIndex <= index && index < iter.EndIndex {
				iter.getFont(&fontDesc, nil, nil)
				break
			}
		}
	}

	font := layout.context.loadFont(&fontDesc)
	if font != nil {
		metrics := FontGetMetrics(font, layout.context.setLanguage)
		// if metrics {
		logicalRect.Y = -metrics.Ascent
		logicalRect.Height = -logicalRect.Y + metrics.Descent
		height = metrics.Height
		// } else {
		// 	logicalRect.y = 0
		// 	logicalRect.height = 0
		// }
	} else {
		logicalRect.Y = 0
		logicalRect.Height = 0
	}

	logicalRect.X = 0
	logicalRect.Width = 0

	return height
}

// GetLinesReadonly is a faster alternative to pango_layout_get_lines(),
// but the user must not modify the contents of the lines (glyphs, glyph widths, etc.).
//
// The returned lines will become invalid on any change to the layout's
// text or properties.
func (layout *Layout) GetLinesReadonly() []*LayoutLine {
	layout.checkLines()

	return layout.lines
}

/**
 * pango_layout_get_line_readonly:
 * @layout: a #Layout
 * @line: the index of a line, which must be between 0 and
 *        <literal>pango_layout_get_line_count(layout) - 1</literal>, inclusive.
 *
 * Retrieves a particular line from a #Layout.
 *
 * This is a faster alternative to GetLine(),
 * but the user is not expected
 * to modify the contents of the line (glyphs, glyph widths, etc.).
 *
 * Return value: (transfer none) (nullable): the requested
 *               #LayoutLine, or %nil if the index is out of
 *               range. This layout line can be ref'ed and retained,
 *               but will become invalid if changes are made to the
 *               #Layout.  No changes should be made to the line.
 **/
// func (layout *Layout) pango_layout_get_line_readonly(line int) *LayoutLine {
// 	GSList * list_item

// 	if line < 0 {
// 		return nil
// 	}

// 	layout.checkLines()

// 	list_item = g_slist_nth(layout.lines, line)

// 	if list_item {
// 		LayoutLine * line = list_item.data
// 		return line
// 	}

// 	return nil
// }

// /**
//  * LayoutIter:
//  *
//  * A #LayoutIter structure can be used to
//  * iterate over the visual extents of a #Layout.
//  *
//  * The #LayoutIter structure is opaque, and
//  * has no user-visible fields.
//  */

//  #include "config.h"
//  #include "pango-glyph.h"		/* For pango_shape() */
//  #include "pango-break.h"
//  #include "pango-item.h"
//  #include "pango-engine.h"
//  #include "pango-impl-utils.h"
//  #include "pango-glyph-item.h"
//  #include <string.h>

//  #include "pango-layout-private.h"
//  #include "pango-attributes-private.h"

//  typedef struct _ItemProperties ItemProperties;
//  typedef struct _ParaBreakState ParaBreakState;

//  typedef struct _LayoutLinePrivate LayoutLinePrivate;

//  struct _LayoutClass
//  {
//    GObjectClass parent_class;

//  };

//  #define LINE_IS_VALID(line) ((line) && (line).layout != nil)

//  #ifdef G_DISABLE_CHECKS
//  #define ITER_IS_INVALID(iter) false
//  #else
//  #define ITER_IS_INVALID(iter) G_UNLIKELY (check_invalid ((iter), G_STRLOC))
//  static bool
//  check_invalid (LayoutIter *iter,
// 			const char      *loc)
//  {
//    if (iter.line.layout == nil)
// 	 {
// 	   g_warning ("%s: Layout changed since LayoutIter was created, iterator invalid", loc);
// 	   return true;
// 	 }
//    else
// 	 {
// 	   return false;
// 	 }
//  }
//  #endif

//  static void check_context_changed  (layout *Layout);
//  static void layoutChanged  (layout *Layout);

//  static void pango_layout_clear_lines (layout *Layout);
//  static void checkLines (layout *Layout);

//  static PangoAttrList *pango_layout_get_effective_attributes (layout *Layout);

//  static LayoutLine * pango_layout_line_new         (Layout     *layout);
//  static void              pango_layout_line_postprocess (line *LayoutLine,
// 							 state *ParaBreakState ,
// 							 bool         wrapped);

//  static int *pango_layout_line_get_log2vis_map (LayoutLine  *line,
// 							bool          strong);
//  static int *pango_layout_line_get_vis2log_map (LayoutLine  *line,
// 							bool          strong);
//  static void leaked (line *LayoutLine);

//  /* doesn't leak line */
//  static LayoutLine* _pango_layout_iter_get_line (LayoutIter *iter);

//  static void pango_layout_get_item_properties (PangoItem      *item,
// 						   ItemProperties *properties);

//  static void pango_layout_get_empty_extents_at_index (layout *Layout    ,
// 							  int             index,
// 							  Rectangle *logicalRect);

//  static void pango_layout_finalize    (GObject          *object);

//  G_DEFINE_TYPE (Layout, pango_layout, G_TYPE_OBJECT)

//  static void
//  pango_layout_class_init (LayoutClass *klass)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (klass);

//    object_class.finalize = pango_layout_finalize;
//  }

//  static void
//  pango_layout_finalize (GObject *object)
//  {
//    layout *Layout;

//    layout = PANGO_LAYOUT (object);

//    pango_layout_clear_lines (layout);

//    if (layout.context)
// 	 g_object_unref (layout.context);

//    if (layout.attrs)
// 	 attr_list_unref (layout.attrs);

//    g_free (layout.text);

//    if (layout.font_desc)
// 	 pango_font_description_free (layout.font_desc);

//    if (layout.tabs)
// 	 pango_tab_array_free (layout.tabs);

//    G_OBJECT_CLASS (pango_layout_parent_class).finalize (object);
//  }

//  /**
//   * pango_layout_copy:
//   * @src: a #Layout
//   *
//   * Does a deep copy-by-value of the @src layout. The attribute list,
//   * tab array, and text from the original layout are all copied by
//   * value.
//   *
//   * Return value: (transfer full): the newly allocated #Layout,
//   *               with a reference count of one, which should be freed
//   *               with g_object_unref().
//   **/
//  Layout*
//  pango_layout_copy (Layout *src)
//  {
//    layout *Layout;

//    g_return_val_if_fail (PANGO_IS_LAYOUT (src), nil);

//    /* Copy referenced members */

//    layout = NewLayout (src.context);
//    if (src.attrs)
// 	 layout.attrs = attr_list_copy (src.attrs);
//    if (src.font_desc)
// 	 layout.font_desc = pango_font_description_copy (src.font_desc);
//    if (src.tabs)
// 	 layout.tabs = pango_tab_array_copy (src.tabs);

//    /* Dupped */
//    layout.text = g_strdup (src.text);

//    /* Value fields */
//    memcpy (&layout.copy_begin, &src.copy_begin,
// 	   G_STRUCT_OFFSET (Layout, copy_end) - G_STRUCT_OFFSET (Layout, copy_begin));

//    return layout;
//  }

//  /**
//   * pango_layout_get_context:
//   * @layout: a #Layout
//   *
//   * Retrieves the #PangoContext used for this layout.
//   *
//   * Return value: (transfer none): the #PangoContext for the layout.
//   * This does not have an additional refcount added, so if you want to
//   * keep a copy of this around, you must reference it yourself.
//   **/
//  PangoContext *
//  pango_layout_get_context (layout *Layout)
//  {
//    g_return_val_if_fail (layout != nil, nil);

//    return layout.context;
//  }

//  /**
//   * pango_layout_get_width:
//   * @layout: a #Layout
//   *
//   * Gets the width to which the lines of the #Layout should wrap.
//   *
//   * Return value: the width in Pango units, or -1 if no width set.
//   **/
//  int
//  pango_layout_get_width (layout *Layout    )
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.width;
//  }

//  /**
//   * SetHeight:
//   * @layout: a #Layout.
//   * @height: the desired height of the layout in Pango units if positive,
//   *          or desired number of lines if negative.
//   *
//   * Sets the height to which the #Layout should be ellipsized at.  There
//   * are two different behaviors, based on whether @height is positive or
//   * negative.
//   *
//   * If @height is positive, it will be the maximum height of the layout.  Only
//   * lines would be shown that would fit, and if there is any text omitted,
//   * an ellipsis added.  At least one line is included in each paragraph regardless
//   * of how small the height value is.  A value of zero will render exactly one
//   * line for the entire layout.
//   *
//   * If @height is negative, it will be the (negative of) maximum number of lines per
//   * paragraph.  That is, the total number of lines shown may well be more than
//   * this value if the layout contains multiple paragraphs of text.
//   * The default value of -1 means that first line of each paragraph is ellipsized.
//   * This behvaior may be changed in the future to act per layout instead of per
//   * paragraph.  File a bug against pango at <ulink
//   * url="http://bugzilla.gnome.org/">http://bugzilla.gnome.org/</ulink> if your
//   * code relies on this behavior.
//   *
//   * Height setting only has effect if a positive width is set on
//   * @layout and ellipsization mode of @layout is not %ELLIPSIZE_NONE.
//   * The behavior is undefined if a height other than -1 is set and
//   * ellipsization mode is set to %ELLIPSIZE_NONE, and may change in the
//   * future.
//   *
//   * Since: 1.20
//   **/
//  void
//  SetHeight (layout *Layout,
// 			  int          height)
//  {
//    g_return_if_fail (layout != nil);

//    if (height != layout.height)
// 	 {
// 	   layout.height = height;

// 	   /* Do not invalidate if the number of lines requested is
// 		* larger than the total number of lines in layout.
// 		* Bug 549003
// 		*/
// 	   if (layout.ellipsize != ELLIPSIZE_NONE &&
// 	   !(layout.lines && layout.is_ellipsized == false &&
// 		 height < 0 && layout.line_count <= (guint) -height))
// 	 layoutChanged (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_height:
//   * @layout: a #Layout
//   *
//   * Gets the height of layout used for ellipsization.  See
//   * SetHeight() for details.
//   *
//   * Return value: the height, in Pango units if positive, or
//   * number of lines if negative.
//   *
//   * Since: 1.20
//   **/
//  int
//  pango_layout_get_height (layout *Layout    )
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.height;
//  }

//  /**
//   * pango_layout_get_wrap:
//   * @layout: a #Layout
//   *
//   * Gets the wrap mode for the layout.
//   *
//   * Use IsWrapped() to query whether any paragraphs
//   * were actually wrapped.
//   *
//   * Return value: active wrap mode.
//   **/
//  PangoWrapMode
//  pango_layout_get_wrap (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), 0);

//    return layout.wrap;
//  }

//  /**
//   * pango_layout_get_indent:
//   * @layout: a #Layout
//   *
//   * Gets the paragraph indent width in Pango units. A negative value
//   * indicates a hanging indentation.
//   *
//   * Return value: the indent in Pango units.
//   **/
//  int
//  pango_layout_get_indent (layout *Layout)
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.indent;
//  }

//  /**
//   * pango_layout_get_spacing:
//   * @layout: a #Layout
//   *
//   * Gets the amount of spacing between the lines of the layout.
//   *
//   * Return value: the spacing in Pango units.
//   **/
//  int
//  pango_layout_get_spacing (layout *Layout)
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.spacing;
//  }

//  /**
//   * pango_layout_get_line_spacing:
//   * @layout: a #Layout
//   *
//   * Gets the value that has been
//   * set with SetLineSpacing().
//   *
//   * Since: 1.44
//   */
//  float
//  pango_layout_get_line_spacing (layout *Layout)
//  {
//    g_return_val_if_fail (layout != nil, 1.0);
//    return layout.line_spacing;
//  }

//  /**
//   * pango_layout_get_font_description:
//   * @layout: a #Layout
//   *
//   * Gets the font description for the layout, if any.
//   *
//   * Return value: (nullable): a pointer to the layout's font
//   *  description, or %nil if the font description from the layout's
//   *  context is inherited. This value is owned by the layout and must
//   *  not be modified or freed.
//   *
//   * Since: 1.8
//   **/
//  const PangoFontDescription *
//  pango_layout_get_font_description (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), nil);

//    return layout.font_desc;
//  }

//  /**
//   * pango_layout_get_auto_dir:
//   * @layout: a #Layout
//   *
//   * Gets whether to calculate the bidirectional base direction
//   * for the layout according to the contents of the layout.
//   * See SetAutoDir().
//   *
//   * Return value: `true` if the bidirectional base direction
//   *   is computed from the layout's contents, %false otherwise.
//   *
//   * Since: 1.4
//   **/
//  bool
//  pango_layout_get_auto_dir (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), false);

//    return layout.autoDir;
//  }

//  /**
//   * pango_layout_getAlignment:
//   * @layout: a #Layout
//   *
//   * Gets the alignment for the layout: how partial lines are
//   * positioned within the horizontal space available.
//   *
//   * Return value: the alignment.
//   **/
//  PangoAlignment
//  pango_layout_getAlignment (layout *Layout)
//  {
//    g_return_val_if_fail (layout != nil, PANGO_ALIGN_LEFT);
//    return layout.alignment;
//  }

//  /**
//   * pango_layout_get_tabs:
//   * @layout: a #Layout
//   *
//   * Gets the current #TabArray used by this layout. If no
//   * #TabArray has been set, then the default tabs are in use
//   * and %nil is returned. Default tabs are every 8 spaces.
//   * The return value should be freed with pango_tab_array_free().
//   *
//   * Return value: (nullable): a copy of the tabs for this layout, or
//   * %nil.
//   **/
//  TabArray*
//  pango_layout_get_tabs (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), nil);

//    if (layout.tabs)
// 	 return pango_tab_array_copy (layout.tabs);
//    else
// 	 return nil;
//  }

//  /**
//   * pango_layout_get_single_paragraph_mode:
//   * @layout: a #Layout
//   *
//   * Obtains the value set by SetSingleParagraphMode().
//   *
//   * Return value: `true` if the layout does not break paragraphs at
//   * paragraph separator characters, %false otherwise.
//   **/
//  bool
//  pango_layout_get_single_paragraph_mode (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), false);

//    return layout.single_paragraph;
//  }

//  /**
//   * pango_layout_get_ellipsize:
//   * @layout: a #Layout
//   *
//   * Gets the type of ellipsization being performed for @layout.
//   * See pango_layout_set_ellipsize()
//   *
//   * Return value: the current ellipsization mode for @layout.
//   *
//   * Use pango_layout_is_ellipsized() to query whether any paragraphs
//   * were actually ellipsized.
//   *
//   * Since: 1.6
//   **/
//  PangoEllipsizeMode
//  pango_layout_get_ellipsize (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), ELLIPSIZE_NONE);

//    return layout.ellipsize;
//  }

//  /**
//   * pango_layout_set_text:
//   * @layout: a #Layout
//   * @text: the text
//   * @length: maximum length of @text, in bytes. -1 indicates that
//   *          the string is nul-terminated and the length should be
//   *          calculated.  The text will also be truncated on
//   *          encountering a nul-termination even when @length is
//   *          positive.
//   *
//   * Sets the text of the layout.
//   *
//   * This function validates @text and renders invalid UTF-8
//   * with a placeholder glyph.
//   *
//   * Note that if you have used SetMarkup() or
//   * setMarkupWithAccel() on @layout before, you may
//   * want to call setAttributes() to clear the attributes
//   * set on the layout from the markup as this function does not clear
//   * attributes.
//   **/
//  void
//  pango_layout_set_text (layout *Layout,
// 				const char  *text,
// 				int          length)
//  {
//    char *old_text, *start, *end;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (length == 0 || text != nil);

//    old_text = layout.text;

//    if (length < 0)
// 	 {
// 	   layout.length = strlen (text);
// 	   layout.text = g_strndup (text, layout.length);
// 	 }
//    else if (length > 0)
// 	 {
// 	   /* This is not exactly what we want.  We don't need the padding...
// 		*/
// 	   layout.length = length;
// 	   layout.text = g_strndup (text, length);
// 	 }
//    else
// 	 {
// 	   layout.length = 0;
// 	   layout.text = g_malloc0 (1);
// 	 }

//    /* validate it, and replace invalid bytes with -1 */
//    start = layout.text;
//    for (;;) {
// 	 bool valid;

// 	 valid = g_utf8_validate (start, -1, (const char **)&end);

// 	 if (!*end)
// 	   break;

// 	 /* Replace invalid bytes with -1.  The -1 will be converted to
// 	  * ((gunichar) -1) by glib, and that in turn yields a glyph value of
// 	  * ((PangoGlyph) -1) by AsUnknownGlyph(-1),
// 	  * and that's GLYPH_INVALID_INPUT.
// 	  */
// 	 if (!valid)
// 	   *end++ = -1;

// 	 start = end;
//    }

//    if (start != layout.text)
// 	 /* TODO: Write out the beginning excerpt of text? */
// 	 g_warning ("Invalid UTF-8 string passed to pango_layout_set_text()");

//    len(layout.text) = pango_utf8_strlen (layout.text, -1);
//    layout.length = strlen (layout.text);

//    layoutChanged (layout);

//    g_free (old_text);
//  }

//  /**
//   * pango_layout_get_text:
//   * @layout: a #Layout
//   *
//   * Gets the text in the layout. The returned text should not
//   * be freed or modified.
//   *
//   * Return value: the text in the @layout.
//   **/
//  const char*
//  pango_layout_get_text (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), nil);

//    /* We don't ever want to return nil as the text.
// 	*/
//    if (G_UNLIKELY (!layout.text))
// 	 return "";

//    return layout.text;
//  }

//  /**
//   * pango_layout_get_character_count:
//   * @layout: a #Layout
//   *
//   * Returns the number of Unicode characters in the
//   * the text of @layout.
//   *
//   * Return value: the number of Unicode characters
//   *     in the text of @layout
//   *
//   * Since: 1.30
//   */
//  gint
//  pango_layout_get_character_count (layout *Layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), 0);

//    return len(layout.text);
//  }

//  /**
//   * pango_layout_get_unknown_glyphs_count:
//   * @layout: a #Layout
//   *
//   * Counts the number unknown glyphs in @layout.  That is, zero if
//   * glyphs for all characters in the layout text were found, or more
//   * than zero otherwise.
//   *
//   * This function can be used to determine if there are any fonts
//   * available to render all characters in a certain string, or when
//   * used in combination with %ATTR_FALLBACK, to check if a
//   * certain font supports all the characters in the string.
//   *
//   * Return value: The number of unknown glyphs in @layout.
//   *
//   * Since: 1.16
//   */
//  int
//  pango_layout_get_unknown_glyphs_count (layout *Layout)
//  {
// 	 line *LayoutLine;
// 	 GlyphItem *run;
// 	 GSList *lines_list;
// 	 GSList *runs_list;
// 	 int i, count = 0;

// 	 g_return_val_if_fail (PANGO_IS_LAYOUT (layout), 0);

// 	 checkLines (layout);

// 	 if (layout.unknown_glyphs_count >= 0)
// 	   return layout.unknown_glyphs_count;

// 	 lines_list = layout.lines;
// 	 for (lines_list)
// 	   {
// 	 line = lines_list.data;
// 	 runs_list = line.runs;

// 	 for (runs_list)
// 	   {
// 		 run = runs_list.data;

// 		 for (i = 0; i < run.glyphs.num_glyphs; i++)
// 		   {
// 		 if (run.glyphs.glyphs[i].glyph & GLYPH_UNKNOWN_FLAG)
// 			 count++;
// 		   }

// 		 runs_list = runs_list.next;
// 	   }
// 	 lines_list = lines_list.next;
// 	   }

// 	 layout.unknown_glyphs_count = count;
// 	 return count;
//  }

//  /**
//   * pango_layout_get_serial:
//   * @layout: a #Layout
//   *
//   * Returns the current serial number of @layout.  The serial number is
//   * initialized to an small number  larger than zero when a new layout
//   * is created and is increased whenever the layout is changed using any
//   * of the setter functions, or the #PangoContext it uses has changed.
//   * The serial may wrap, but will never have the value 0. Since it
//   * can wrap, never compare it with "less than", always use "not equals".
//   *
//   * This can be used to automatically detect changes to a #Layout, and
//   * is useful for example to decide whether a layout needs redrawing.
//   * To force the serial to be increased, use pango_layout_context_changed().
//   *
//   * Return value: The current serial number of @layout.
//   *
//   * Since: 1.32.4
//   **/
//  guint
//  pango_layout_get_serial (layout *Layout)
//  {
//    check_context_changed (layout);
//    return layout.serial;
//  }

//  /**
//   * pango_layout_get_log_attrs_readonly:
//   * @layout: a #Layout
//   * @n_attrs: (out): location to store the number of the attributes in
//   *   the array
//   *
//   * Retrieves an array of logical attributes for each character in
//   * the @layout.
//   *
//   * This is a faster alternative to GetCharacterAttributes().
//   * The returned array is part of @layout and must not be modified.
//   * Modifying the layout will invalidate the returned array.
//   *
//   * The number of attributes returned in @n_attrs will be one more
//   * than the total number of characters in the layout, since there
//   * need to be attributes corresponding to both the position before
//   * the first character and the position after the last character.
//   *
//   * Returns: (array length=n_attrs): an array of logical attributes
//   *
//   * Since: 1.30
//   */
//  const PangoLogAttr *
//  pango_layout_get_log_attrs_readonly (layout *Layout,
// 									  gint        *n_attrs)
//  {
//    if (n_attrs)
// 	 *n_attrs = 0;
//    g_return_val_if_fail (layout != nil, nil);

//    checkLines (layout);

//    if (n_attrs)
// 	 *n_attrs = len(layout.text) + 1;

//    return layout.log_attrs;
//  }

//  /**
//   * pango_layout_get_lines:
//   * @layout: a #Layout
//   *
//   * Returns the lines of the @layout as a list.
//   *
//   * Use the faster GetLinesReadonly() if you do not plan
//   * to modify the contents of the lines (glyphs, glyph widths, etc.).
//   *
//   * Return value: (element-type Pango.LayoutLine) (transfer none): a #GSList containing
//   * the lines in the layout. This points to internal data of the #Layout
//   * and must be used with care. It will become invalid on any change to the layout's
//   * text or properties.
//   **/
//  GSList *
//  pango_layout_get_lines (layout *Layout)
//  {
//    checkLines (layout);

//    if (layout.lines)
// 	 {
// 	   GSList *tmp_list = layout.lines;
// 	   for (tmp_list)
// 	 {
// 	   line *LayoutLine = tmp_list.data;
// 	   tmp_list = tmp_list.next;

// 	   leaked (line);
// 	 }
// 	 }

//    return layout.lines;
//  }

//  /**
//   * MoveCursorVisually:
//   * @layout:       a #Layout.
//   * @strong:       whether the moving cursor is the strong cursor or the
//   *                weak cursor. The strong cursor is the cursor corresponding
//   *                to text insertion in the base direction for the layout.
//   * @old_index:    the byte index of the grapheme for the old index
//   * @old_trailing: if 0, the cursor was at the leading edge of the
//   *                grapheme indicated by @old_index, if > 0, the cursor
//   *                was at the trailing edge.
//   * @direction:    direction to move cursor. A negative
//   *                value indicates motion to the left.
//   * @new_index: (out): location to store the new cursor byte index. A value of -1
//   *                indicates that the cursor has been moved off the beginning
//   *                of the layout. A value of %G_MAXINT indicates that
//   *                the cursor has been moved off the end of the layout.
//   * @new_trailing: (out): number of characters to move forward from the
//   *                location returned for @new_index to get the position
//   *                where the cursor should be displayed. This allows
//   *                distinguishing the position at the beginning of one
//   *                line from the position at the end of the preceding
//   *                line. @new_index is always on the line where the
//   *                cursor should be displayed.
//   *
//   * Computes a new cursor position from an old position and
//   * a count of positions to move visually. If @direction is positive,
//   * then the new strong cursor position will be one position
//   * to the right of the old cursor position. If @direction is negative,
//   * then the new strong cursor position will be one position
//   * to the left of the old cursor position.
//   *
//   * In the presence of bidirectional text, the correspondence
//   * between logical and visual order will depend on the direction
//   * of the current run, and there may be jumps when the cursor
//   * is moved off of the end of a run.
//   *
//   * Motion here is in cursor positions, not in characters, so a
//   * single call to MoveCursorVisually() may move the
//   * cursor over multiple characters when multiple characters combine
//   * to form a single grapheme.
//   **/
//  void
//  MoveCursorVisually (layout *Layout,
// 					bool     strong,
// 					int          old_index,
// 					int          old_trailing,
// 					int          direction,
// 					int         *new_index,
// 					int         *new_trailing)
//  {
//    line *LayoutLine = nil;
//    LayoutLine *prevLine;
//    LayoutLine *next_line;

//    int *log2vis_map;
//    int *vis2log_map;
//    int n_vis;
//    int vis_pos, vis_pos_old, log_pos;
//    int start_offset;
//    bool off_start = false;
//    bool off_end = false;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (old_index >= 0 && old_index <= layout.length);
//    g_return_if_fail (old_index < layout.length || old_trailing == 0);
//    g_return_if_fail (new_index != nil);
//    g_return_if_fail (new_trailing != nil);

//    direction = (direction >= 0 ? 1 : -1);

//    checkLines (layout);

//    /* Find the line the old cursor is on */
//    line = indexToLine (layout, old_index,
// 					  nil, &prevLine, &next_line);

//    start_offset = g_utf8_pointer_to_offset (layout.text, layout.text + line.startIndex);

//    for (old_trailing--)
// 	 old_index = g_utf8_next_char (layout.text + old_index) - layout.text;

//    log2vis_map = pango_layout_line_get_log2vis_map (line, strong);
//    n_vis = pango_utf8_strlen (layout.text + line.startIndex, line.length);

//    /* Clamp old_index to fit on the line */
//    if (old_index > (line.startIndex + line.length))
// 	 old_index = line.startIndex + line.length;

//    vis_pos = log2vis_map[old_index - line.startIndex];

//    g_free (log2vis_map);

//    /* Handling movement between lines */
//    if (vis_pos == 0 && direction < 0)
// 	 {
// 	   if (line.resolved_dir == PANGO_DIRECTION_LTR)
// 	 off_start = true;
// 	   else
// 	 off_end = true;
// 	 }
//    else if (vis_pos == n_vis && direction > 0)
// 	 {
// 	   if (line.resolved_dir == PANGO_DIRECTION_LTR)
// 	 off_end = true;
// 	   else
// 	 off_start = true;
// 	 }

//    if (off_start || off_end)
// 	 {
// 	   /* If we move over a paragraph boundary, count that as
// 		* an extra position in the motion
// 		*/
// 	   bool paragraph_boundary;

// 	   if (off_start)
// 	 {
// 	   if (!prevLine)
// 		 {
// 		   *new_index = -1;
// 		   *new_trailing = 0;
// 		   return;
// 		 }
// 	   line = prevLine;
// 	   paragraph_boundary = (line.startIndex + line.length != old_index);
// 	 }
// 	   else
// 	 {
// 	   if (!next_line)
// 		 {
// 		   *new_index = G_MAXINT;
// 		   *new_trailing = 0;
// 		   return;
// 		 }
// 	   line = next_line;
// 	   paragraph_boundary = (line.startIndex != old_index);
// 	 }

// 	   n_vis = pango_utf8_strlen (layout.text + line.startIndex, line.length);
// 	   start_offset = g_utf8_pointer_to_offset (layout.text, layout.text + line.startIndex);

// 	   if (vis_pos == 0 && direction < 0)
// 	 {
// 	   vis_pos = n_vis;
// 	   if (paragraph_boundary)
// 		 vis_pos++;
// 	 }
// 	   else /* (vis_pos == n_vis && direction > 0) */
// 	 {
// 	   vis_pos = 0;
// 	   if (paragraph_boundary)
// 		 vis_pos--;
// 	 }
// 	 }

//    vis2log_map = pango_layout_line_get_vis2log_map (line, strong);

//    vis_pos_old = vis_pos + direction;
//    log_pos = g_utf8_pointer_to_offset (layout.text + line.startIndex,
// 					   layout.text + line.startIndex + vis2log_map[vis_pos_old]);
//    do
// 	 {
// 	   vis_pos += direction;
// 	   log_pos += g_utf8_pointer_to_offset (layout.text + line.startIndex + vis2log_map[vis_pos_old],
// 						layout.text + line.startIndex + vis2log_map[vis_pos]);
// 	   vis_pos_old = vis_pos;
// 	 }
//    for (vis_pos > 0 && vis_pos < n_vis &&
// 	  !layout.log_attrs[start_offset + log_pos].is_cursor_position);

//    *new_index = line.startIndex + vis2log_map[vis_pos];
//    g_free (vis2log_map);

//    *new_trailing = 0;

//    if (*new_index == line.startIndex + line.length && line.length > 0)
// 	 {
// 	   do
// 	 {
// 	   log_pos--;
// 	   *new_index = g_utf8_prev_char (layout.text + *new_index) - layout.text;
// 	   (*new_trailing)++;
// 	 }
// 	   for (log_pos > 0 && !layout.log_attrs[start_offset + log_pos].is_cursor_position);
// 	 }
//  }

//  /**
//   * pango_layout_xy_to_index:
//   * @layout:    a #Layout
//   * @x:         the X offset (in Pango units)
//   *             from the left edge of the layout.
//   * @y:         the Y offset (in Pango units)
//   *             from the top edge of the layout
//   * @index_: (out):   location to store calculated byte index
//   * @trailing: (out): location to store a integer indicating where
//   *             in the grapheme the user clicked. It will either
//   *             be zero, or the number of characters in the
//   *             grapheme. 0 represents the leading edge of the grapheme.
//   *
//   * Converts from X and Y position within a layout to the byte
//   * index to the character at that logical position. If the
//   * Y position is not inside the layout, the closest position is chosen
//   * (the position will be clamped inside the layout). If the
//   * X position is not within the layout, then the start or the
//   * end of the line is chosen as Described for pango_layout_line_x_to_index().
//   * If either the X or Y positions were not inside the layout, then the
//   * function returns %false; on an exact hit, it returns `true`.
//   *
//   * Return value: `true` if the coordinates were inside text, %false otherwise.
//   **/
//  bool
//  pango_layout_xy_to_index (layout *Layout,
// 			   int          x,
// 			   int          y,
// 			   int         *index,
// 			   gint        *trailing)
//  {
//    LayoutIter iter;
//    LayoutLine *prevLine = nil;
//    LayoutLine *found = nil;
//    int found_line_x = 0;
//    int prev_last = 0;
//    int prev_line_x = 0;
//    bool retval = false;
//    bool outside = false;

//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), false);

//    _pango_layout_get_iter (layout, &iter);

//    do
// 	 {
// 	   Rectangle line_logical;
// 	   int first_y, last_y;

// 	   assert (!ITER_IS_INVALID (&iter));

// 	   getLineExtents (&iter, nil, &line_logical);
// 	   pango_layout_iter_get_line_yrange (&iter, &first_y, &last_y);

// 	   if (y < first_y)
// 	 {
// 	   if (prevLine && y < (prev_last + (first_y - prev_last) / 2))
// 		 {
// 		   found = prevLine;
// 		   found_line_x = prev_line_x;
// 		 }
// 	   else
// 		 {
// 		   if (prevLine == nil)
// 		 outside = true; /* off the top */

// 		   found = _pango_layout_iter_get_line (&iter);
// 		   found_line_x = x - line_logical.x;
// 		 }
// 	 }
// 	   else if (y >= first_y &&
// 			y < last_y)
// 	 {
// 	   found = _pango_layout_iter_get_line (&iter);
// 	   found_line_x = x - line_logical.x;
// 	 }

// 	   prevLine = _pango_layout_iter_get_line (&iter);
// 	   prev_last = last_y;
// 	   prev_line_x = x - line_logical.x;

// 	   if (found != nil)
// 	 break;
// 	 }
//    for (NextLine (&iter));

//    _pango_layout_iter_destroy (&iter);

//    if (found == nil)
// 	 {
// 	   /* Off the bottom of the layout */
// 	   outside = true;

// 	   found = prevLine;
// 	   found_line_x = prev_line_x;
// 	 }

//    retval = pango_layout_line_x_to_index (found,
// 					  found_line_x,
// 					  index, trailing);

//    if (outside)
// 	 retval = false;

//    return retval;
//  }

// /**
//  * pango_layout_index_to_pos:
//  * @layout: a `PangoLayout`
//  * @index_: byte index within @layout
//  * @pos: (out): rectangle in which to store the position of the grapheme
//  *
//  * Converts from an index within a `PangoLayout` to the onscreen position
//  * corresponding to the grapheme at that index.
//  *
//  * The return value is represented as rectangle. Note that `pos->x` is
//  * always the leading edge of the grapheme and `pos->x + pos->width` the
//  * trailing edge of the grapheme. If the directionality of the grapheme
//  * is right-to-left, then `pos->width` will be negative.
//  */
//  void
//  pango_layout_index_to_pos (PangoLayout    *layout,
// 							int             index,
// 							PangoRectangle *pos)
//  {
//    PangoRectangle line_logical_rect;
//    PangoRectangle run_logical_rect;
//    PangoLayoutIter iter;
//    PangoLayoutLine *layout_line = NULL;
//    int x_pos;

//    g_return_if_fail (layout != NULL);
//    g_return_if_fail (index >= 0);
//    g_return_if_fail (pos != NULL);

//    _pango_layout_get_iter (layout, &iter);

//    if (!ITER_IS_INVALID (&iter))
// 	 {
// 	   while (TRUE)
// 		 {
// 		   PangoLayoutLine *tmp_line = _pango_layout_iter_get_line (&iter);

// 		   if (tmp_line->start_index > index)
// 			 {
// 			   /* index is in the paragraph delimiters, move to
// 				* end of previous line
// 				*
// 				* This shouldn’t occur in the first loop iteration as the first
// 				* line’s start_index should always be 0.
// 				*/
// 			   g_assert (layout_line != NULL);
// 			   index = layout_line->start_index + layout_line->length;
// 			   break;
// 			 }

// 		   pango_layout_iter_get_line_extents (&iter, NULL, &line_logical_rect);

// 		   layout_line = tmp_line;

// 		   if (layout_line->start_index + layout_line->length >= index)
// 			 {
// 			   do
// 				 {
// 				   PangoLayoutRun *run = _pango_layout_iter_get_run (&iter);

// 				   pango_layout_iter_get_run_extents (&iter, NULL, &run_logical_rect);

// 				   if (!run)
// 					 break;

// 				   if (run->item->offset <= index && index < run->item->offset + run->item->length)
// 					 break;
// 				  }
// 				while (pango_layout_iter_next_run (&iter));

// 			   if (layout_line->start_index + layout_line->length > index)
// 				 break;
// 			 }

// 		   if (!pango_layout_iter_next_line (&iter))
// 			 {
// 			   index = layout_line->start_index + layout_line->length;
// 			   break;
// 			 }
// 		 }

// 	   pos->y = run_logical_rect.y;
// 	   pos->height = run_logical_rect.height;

// 	   pango_layout_line_index_to_x (layout_line, index, 0, &x_pos);
// 	   pos->x = line_logical_rect.x + x_pos;

// 	   if (index < layout_line->start_index + layout_line->length)
// 		 {
// 		   pango_layout_line_index_to_x (layout_line, index, 1, &x_pos);
// 		   pos->width = (line_logical_rect.x + x_pos) - pos->x;
// 		 }
// 	   else
// 		 pos->width = 0;
// 	 }

//    _pango_layout_iter_destroy (&iter);
//  }

//  static void
//  pango_layout_line_get_range (line *LayoutLine,
// 				  char           **start,
// 				  char           **end)
//  {
//    char *p;

//    p = line.layout.text + line.startIndex;

//    if (start)
// 	 *start = p;
//    if (end)
// 	 *end = p + line.length;
//  }

//  static int *
//  pango_layout_line_get_vis2log_map (line *LayoutLine,
// 					bool         strong)
//  {
//    layout *Layout = line.layout;
//    PangoDirection prev_dir;
//    PangoDirection cursor_dir;
//    GSList *tmp_list;
//    gchar *start, *end;
//    int *result;
//    int pos;
//    int n_chars;

//    pango_layout_line_get_range (line, &start, &end);
//    n_chars = pango_utf8_strlen (start, end - start);

//    result = g_new (int, n_chars + 1);

//    if (strong)
// 	 cursor_dir = line.resolved_dir;
//    else
// 	 cursor_dir = (line.resolved_dir == PANGO_DIRECTION_LTR) ? PANGO_DIRECTION_RTL : PANGO_DIRECTION_LTR;

//    /* Handle the first visual position
// 	*/
//    if (line.resolved_dir == cursor_dir)
// 	 result[0] = line.resolved_dir == PANGO_DIRECTION_LTR ? 0 : end - start;

//    prev_dir = line.resolved_dir;
//    pos = 0;
//    tmp_list = line.runs;
//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;
// 	   int run_n_chars = run.item.num_chars;
// 	   PangoDirection run_dir = (run.item.analysis.level % 2) ? PANGO_DIRECTION_RTL : PANGO_DIRECTION_LTR;
// 	   char *p = layout.text + run.item.offset;
// 	   int i;

// 	   /* pos is the visual position at the start of the run */
// 	   /* p is the logical byte index at the start of the run */

// 	   if (run_dir == PANGO_DIRECTION_LTR)
// 	 {
// 	   if ((cursor_dir == PANGO_DIRECTION_LTR) ||
// 		   (prev_dir == run_dir))
// 		 result[pos] = p - start;

// 	   p = g_utf8_next_char (p);

// 	   for (i = 1; i < run_n_chars; i++)
// 		 {
// 		   result[pos + i] = p - start;
// 		   p = g_utf8_next_char (p);
// 		 }

// 	   if (cursor_dir == PANGO_DIRECTION_LTR)
// 		 result[pos + run_n_chars] = p - start;
// 	 }
// 	   else
// 	 {
// 	   if (cursor_dir == PANGO_DIRECTION_RTL)
// 		 result[pos + run_n_chars] = p - start;

// 	   p = g_utf8_next_char (p);

// 	   for (i = 1; i < run_n_chars; i++)
// 		 {
// 		   result[pos + run_n_chars - i] = p - start;
// 		   p = g_utf8_next_char (p);
// 		 }

// 	   if ((cursor_dir == PANGO_DIRECTION_RTL) ||
// 		   (prev_dir == run_dir))
// 		 result[pos] = p - start;
// 	 }

// 	   pos += run_n_chars;
// 	   prev_dir = run_dir;
// 	   tmp_list = tmp_list.next;
// 	 }

//    /* And the last visual position
// 	*/
//    if ((cursor_dir == line.resolved_dir) || (prev_dir == line.resolved_dir))
// 	 result[pos] = line.resolved_dir == PANGO_DIRECTION_LTR ? end - start : 0;

//    return result;
//  }

//  static int *
//  pango_layout_line_get_log2vis_map (line *LayoutLine,
// 					bool         strong)
//  {
//    gchar *start, *end;
//    int *reverse_map;
//    int *result;
//    int i;
//    int n_chars;

//    pango_layout_line_get_range (line, &start, &end);
//    n_chars = pango_utf8_strlen (start, end - start);
//    result = g_new0 (int, end - start + 1);

//    reverse_map = pango_layout_line_get_vis2log_map (line, strong);

//    for (i=0; i <= n_chars; i++)
// 	 result[reverse_map[i]] = i;

//    g_free (reverse_map);

//    return result;
//  }

//  /**
//   * getCursorPos:
//   * @layout: a #Layout
//   * @index_: the byte index of the cursor
//   * @strong_pos: (out) (allow-none): location to store the strong cursor position
//   *                     (may be %nil)
//   * @weak_pos: (out) (allow-none): location to store the weak cursor position (may be %nil)
//   *
//   * Given an index within a layout, determines the positions that of the
//   * strong and weak cursors if the insertion point is at that
//   * index. The position of each cursor is stored as a zero-width
//   * rectangle. The strong cursor location is the location where
//   * characters of the directionality equal to the base direction of the
//   * layout are inserted.  The weak cursor location is the location
//   * where characters of the directionality opposite to the base
//   * direction of the layout are inserted.
//   **/
//  void
//  getCursorPos (layout *Layout    ,
// 				  int             index,
// 				  Rectangle *strong_pos,
// 				  Rectangle *weak_pos)
//  {
//    PangoDirection dir1;
//    Rectangle lineRect;
//    LayoutLine *layoutLine = nil; /* Quiet GCC */
//    int x1_trailing;
//    int x2;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (index >= 0 && index <= layout.length);

//    layoutLine = indexToLineAndExtents (layout, index,
// 							 &lineRect);

//    assert (index >= layoutLine.startIndex);

//    /* Examine the trailing edge of the character before the cursor */
//    if (index == layoutLine.startIndex)
// 	 {
// 	   dir1 = layoutLine.resolved_dir;
// 	   if (layoutLine.resolved_dir == PANGO_DIRECTION_LTR)
// 	 x1_trailing = 0;
// 	   else
// 	 x1_trailing = lineRect.width;
// 	 }
//    else if (index >= layoutLine.startIndex + layoutLine.length)
// 	 {
// 	   dir1 = layoutLine.resolved_dir;
// 	   if (layoutLine.resolved_dir == PANGO_DIRECTION_LTR)
// 	 x1_trailing = lineRect.width;
// 	   else
// 	 x1_trailing = 0;
// 	 }
//    else
// 	 {
// 	   gint prev_index = g_utf8_prev_char (layout.text + index) - layout.text;
// 	   dir1 = getCharDirection (layoutLine, prev_index);
// 	   indexToX (layoutLine, prev_index, true, &x1_trailing);
// 	 }

//    /* Examine the leading edge of the character after the cursor */
//    if (index >= layoutLine.startIndex + layoutLine.length)
// 	 {
// 	   if (layoutLine.resolved_dir == PANGO_DIRECTION_LTR)
// 	 x2 = lineRect.width;
// 	   else
// 	 x2 = 0;
// 	 }
//    else
// 	 {
// 	   indexToX (layoutLine, index, false, &x2);
// 	 }

//    if (strong_pos)
// 	 {
// 	   strong_pos.x = lineRect.x;

// 	   if (dir1 == layoutLine.resolved_dir)
// 	 strong_pos.x += x1_trailing;
// 	   else
// 	 strong_pos.x += x2;

// 	   strong_pos.y = lineRect.y;
// 	   strong_pos.width = 0;
// 	   strong_pos.height = lineRect.height;
// 	 }

//    if (weak_pos)
// 	 {
// 	   weak_pos.x = lineRect.x;

// 	   if (dir1 == layoutLine.resolved_dir)
// 	 weak_pos.x += x2;
// 	   else
// 	 weak_pos.x += x1_trailing;

// 	   weak_pos.y = lineRect.y;
// 	   weak_pos.width = 0;
// 	   weak_pos.height = lineRect.height;
// 	 }
//  }

//  /**
//   * pango_layout_get_pixel_extents:
//   * @layout:   a #Layout
//   * @inkRect: (out) (allow-none): rectangle used to store the extents of the
//   *                   layout as drawn or %nil to indicate that the result is
//   *                   not needed.
//   * @logicalRect: (out) (allow-none): rectangle used to store the logical
//   *                       extents of the layout or %nil to indicate that the
//   *                       result is not needed.
//   *
//   * Computes the logical and ink extents of @layout in device units.
//   * This function just calls GetExtents() followed by
//   * two pango_extents_to_pixels() calls, rounding @inkRect and @logicalRect
//   * such that the rounded rectangles fully contain the unrounded one (that is,
//   * passes them as first argument to pango_extents_to_pixels()).
//   **/
//  void
//  pango_layout_get_pixel_extents (layout *Layout,
// 				 Rectangle *inkRect,
// 				 Rectangle *logicalRect)
//  {
//    g_return_if_fail (PANGO_IS_LAYOUT (layout));

//    GetExtents (layout, inkRect, logicalRect);
//    pango_extents_to_pixels (inkRect, nil);
//    pango_extents_to_pixels (logicalRect, nil);
//  }

//  /**
//   * pango_layout_get_size:
//   * @layout: a #Layout
//   * @width: (out) (allow-none): location to store the logical width, or %nil
//   * @height: (out) (allow-none): location to store the logical height, or %nil
//   *
//   * Determines the logical width and height of a #Layout
//   * in Pango units (device units scaled by %Scale). This
//   * is simply a convenience function around GetExtents().
//   **/
//  void
//  pango_layout_get_size (layout *Layout,
// 				int         *width,
// 				int         *height)
//  {
//    Rectangle logicalRect;

//    GetExtents (layout, nil, &logicalRect);

//    if (width)
// 	 *width = logicalRect.width;
//    if (height)
// 	 *height = logicalRect.height;
//  }

//  /**
//   * pango_layout_get_pixel_size:
//   * @layout: a #Layout
//   * @width: (out) (allow-none): location to store the logical width, or %nil
//   * @height: (out) (allow-none): location to store the logical height, or %nil
//   *
//   * Determines the logical width and height of a #Layout
//   * in device units. (pango_layout_get_size() returns the width
//   * and height scaled by %Scale.) This
//   * is simply a convenience function around
//   * pango_layout_get_pixel_extents().
//   **/
//  void
//  pango_layout_get_pixel_size (layout *Layout,
// 				  int         *width,
// 				  int         *height)
//  {
//    Rectangle logicalRect;

//    pango_layout_get_extents_internal (layout, nil, &logicalRect, nil);
//    pango_extents_to_pixels (&logicalRect, nil);

//    if (width)
// 	 *width = logicalRect.width;
//    if (height)
// 	 *height = logicalRect.height;
//  }

//  /**
//   * pango_layout_get_baseline:
//   * @layout: a #Layout
//   *
//   * Gets the Y position of baseline of the first line in @layout.
//   *
//   * Return value: baseline of first line, from top of @layout.
//   *
//   * Since: 1.22
//   **/
//  int
//  pango_layout_get_baseline (layout *Layout    )
//  {
//    int baseline;
//    extents *extents = nil;

//    /* XXX this is kinda inefficient */
//    pango_layout_get_extents_internal (layout, nil, nil, &extents);
//    baseline = extents ? extents[0].baseline : 0;

//    g_free (extents);

//    return baseline;
//  }

//

//
//  /*****************
//   * Line Breaking *
//   *****************/

//  static void shape_tab (LayoutLine  *line,
// 						PangoItem        *item,
// 				PangoGlyphString *glyphs);

//  static void
//  free_run (GlyphItem *run, gpointer data)
//  {
//    bool free_item = data != nil;
//    if (free_item)
// 	 pango_item_free (run.item);

//    pango_glyph_string_free (run.glyphs);
//    g_slice_free (GlyphItem, run);
//  }

// setup the cached value `tabWidth` if not already defined
func (layout *Layout) ensure_tab_width() {
	if layout.tabWidth != -1 {
		return
	}
	// Find out how wide 8 spaces are in the context's default
	// font. Utter performance killer. :-(
	glyphs := &GlyphString{}
	//    PangoItem *item;
	//    GList *items;
	//    PangoAttribute *attr;
	//    PangoAttrList *layout_attrs;
	var (
		language Language
		tmpAttrs AttrList
	)
	font_desc := layout.context.fontDesc // copy

	shape_flags := shapeNONE
	if layout.context.roundGlyphPositions {
		shape_flags |= shapeROUND_POSITIONS
	}

	layout_attrs := layout.pango_layout_get_effective_attributes()
	if layout_attrs != nil {
		iter := layout_attrs.getIterator()
		iter.getFont(&font_desc, &language, nil)
	}

	attr := NewAttrFontDescription(font_desc)
	tmpAttrs.pango_attr_list_insert_before(attr)

	if language != "" {
		attr = NewAttrLanguage(language)
		tmpAttrs.pango_attr_list_insert_before(attr)
	}

	items := layout.context.Itemize([]rune{' '}, 0, 1, tmpAttrs)

	item := items.Data
	spaces := []rune("        ")
	glyphs.shapeWithFlags(spaces, 0, len(spaces), &item.Analysis, shape_flags)

	layout.tabWidth = glyphs.getWidth()

	// We need to make sure the tabWidth is > 0 so finding tab positions
	// terminates. This check should be necessary only under extreme
	// problems with the font.
	if layout.tabWidth <= 0 {
		layout.tabWidth = 50 * Scale /* pretty much arbitrary */
	}
}

//  For now we only need the tab position, we assume
//  all tabs are left-aligned.
func (layout *Layout) get_tab_pos(index int) (int, bool) {
	var (
		nTabs     int
		inPixels  bool
		isDefault = true
	)

	if layout.tabs != nil {
		nTabs = len(layout.tabs.tabs)
		inPixels = layout.tabs.positions_in_pixels
		isDefault = false
	}

	if index < nTabs {
		_, pos := layout.tabs.pango_tab_array_get_tab(index)
		if inPixels {
			return pos * Scale, isDefault
		}
		return pos, isDefault
	}

	if nTabs > 0 {
		// Extrapolate tab position, repeating the last tab gap to infinity.

		_, lastPos := layout.tabs.pango_tab_array_get_tab(nTabs - 1)

		var nextToLastPos int
		if nTabs > 1 {
			_, nextToLastPos = layout.tabs.pango_tab_array_get_tab(nTabs - 2)
		}

		if inPixels {
			nextToLastPos *= Scale
			lastPos *= Scale
		}

		var tabWidth int
		if lastPos > nextToLastPos {
			tabWidth = lastPos - nextToLastPos
		} else {
			tabWidth = int(layout.tabWidth)
		}

		return lastPos + tabWidth*(index-nTabs+1), isDefault
	}
	// No tab array set, so use default tab width
	return int(layout.tabWidth) * index, isDefault
}

func (layout *Layout) canBreakAt(offset int, alwaysWrapChar bool) bool {
	// We probably should have a mode where we treat all white-space as
	// of fungible width - appropriate for typography but not for
	// editing.
	wrap := layout.wrap

	if wrap == WRAP_WORD_CHAR {
		if alwaysWrapChar {
			wrap = WRAP_CHAR
		} else {
			wrap = WRAP_WORD
		}
	}

	if offset == len(layout.Text) {
		return true
	} else if wrap == WRAP_WORD {
		return layout.logAttrs[offset].IsLineBreak()
	} else if wrap == WRAP_CHAR {
		return layout.logAttrs[offset].IsCharBreak()
	} else {
		if debugMode {
			log.Println("canBreakAt : broken Layout")
		}
		return true
	}
}

func (layout *Layout) canBreakIn(start_offset, num_chars int, allowBreakAtStart bool) bool {
	i := 1
	if allowBreakAtStart {
		i = 0
	}

	for ; i < num_chars; i++ {
		if layout.canBreakAt(start_offset+i, false) {
			return true
		}
	}

	return false
}

//  static inline void
//  distributeLetterSpacing (int  letter_spacing,
// 				int *space_left,
// 				int *space_right)
//  {
//    *space_left = letter_spacing / 2;
//    /* hinting */
//    if ((letter_spacing & (Scale - 1)) == 0)
// 	 {
// 	   *space_left = PANGO_UNITS_ROUND (*space_left);
// 	 }
//    *space_right = letter_spacing - *space_left;
//  }

// ItemList is a single linked list of Item elements.
type ItemList struct {
	Data *Item
	Next *ItemList
}

type paraBreakState struct {
	/* maintained per layout */
	line_height      GlyphUnit /* Estimate of height of current line; < 0 is no estimate */
	remaining_height GlyphUnit /* Remaining height of the layout;  only defined if layout.height >= 0 */

	/* maintained per paragraph */
	attrs       AttrList  /* Attributes being used for itemization */
	items       *ItemList /* This paragraph turned into items */
	base_dir    Direction /* Current resolved base direction */
	line_of_par int       /* Line of the paragraph, starting at 1 for first line */

	glyphs            *GlyphString   /* Glyphs for the first item in state.items */
	startOffset       int            /* Character offset of first item in state.items in layout.text */
	properties        itemProperties /* Properties for the first item in state.items */
	log_widths        []GlyphUnit    /* Logical widths for first item in state.items.. */
	log_widths_offset int            /* Offset into log_widths to the point corresponding
	 * to the remaining portion of the first item */

	need_hyphen []bool /* Insert a hyphen if breaking here ? */
	// TODO: cleanup since line_start_index = line_start_offset
	lineStartIndex  int /* Start index of line in layout.text */
	lineStartOffset int /* Character offset of line in layout.text */

	/* maintained per line */
	lineWidth       GlyphUnit /* Goal width of line currently processing; < 0 is infinite */
	remaining_width GlyphUnit /* Amount of space remaining on line; < 0 is infinite */

	hyphen_width GlyphUnit /* How much space a hyphen will take */
}

//  static bool
//  should_ellipsize_current_line (layout *Layout    ,
// 					ParaBreakState *state);

func (layout *Layout) break_needs_hyphen(state *paraBreakState, pos int) bool {
	c := layout.logAttrs[state.startOffset+pos]
	return c.IsBreakInsertsHyphen() || c.IsBreakRemovesPreceding()
}

func (state *paraBreakState) ensure_hyphen_width() {
	if state.hyphen_width < 0 {
		item := state.items.Data
		state.hyphen_width = item.find_hyphen_width()
	}
}

func (layout *Layout) find_break_extra_width(state *paraBreakState, pos int) GlyphUnit {
	// Check whether to insert a hyphen
	if attr := layout.logAttrs[state.startOffset+pos]; attr.IsBreakInsertsHyphen() {
		state.ensure_hyphen_width()

		if attr.IsBreakRemovesPreceding() {
			wc := layout.Text[state.startOffset+pos-1]
			return state.hyphen_width - state.items.Data.find_char_width(wc)
		} else {
			return state.hyphen_width
		}
	}
	return 0
}

//  #if 0
//  # define DEBUG debug
//  void
//  debug (const char *where, line *LayoutLine, ParaBreakState *state)
//  {
//    int line_width = pango_layout_line_get_width (line);

//    g_debug ("rem %d + line %d = %d		%s",
// 		state.remaining_width,
// 		line_width,
// 		state.remaining_width + line_width,
// 		where);
//  }
//  #else
//  # define DEBUG(where, line, state) do { } for (0)
//  #endif

type breakResult uint8

const (
	brNONE_FIT       breakResult = iota // Couldn't fit anything.
	brSOME_FIT                          // The item was broken in the middle.
	brALL_FIT                           // Everything fit.
	brEMPTY_FIT                         // Nothing fit, but that was ok, as we can break at the first char.
	brLINE_SEPARATOR                    // Item begins with a line separator.
)

// process_item tries to insert as much as possible of the first item of
// `state.items` onto `line`.
//
// If `forceFit` is `true`, then `BREAK_NONE_FIT` will never
// be returned, a run will be added even if inserting the minimum amount
// will cause the line to overflow. This is used at the start of a line
// and until we've found at least some place to break.
//
// If `noBreakAtEnd` is `true`, then `BREAK_ALL_FIT` will never be
// returned even if everything fits; the run will be broken earlier,
// or `BREAK_NONE_FIT` returned. This is used when the end of the
// run is not a break position.
func (layout *Layout) process_item(line *LayoutLine, state *paraBreakState,
	forceFit bool, noBreakAtEnd bool) breakResult {
	//    length int;
	//    i int;
	item := state.items.Data
	shape_set := false
	processing_new_item := false

	if state.glyphs == nil {
		state.properties = item.pango_layout_get_item_properties()
		state.glyphs = line.shape_run(state, item)

		state.log_widths = nil
		state.log_widths_offset = 0

		processing_new_item = true
	}

	if !layout.singleParagraph && layout.Text[item.Offset] == lineSeparator &&
		!layout.shouldEllipsizeCurrentLine(state) {
		line.insert_run(state, item, true)
		state.log_widths_offset += item.Length

		return brLINE_SEPARATOR
	}

	if state.remaining_width < 0 && !noBreakAtEnd /* Wrapping off */ {
		line.insert_run(state, item, true)

		return brALL_FIT
	}

	var width GlyphUnit
	if processing_new_item {
		width = state.glyphs.getWidth()
	} else {
		for _, w := range state.log_widths[state.log_widths_offset : state.log_widths_offset+item.Length] {
			width += w
		}
	}

	if (width <= state.remaining_width || (item.Length == 1 && line.Runs == nil)) && !noBreakAtEnd {
		state.remaining_width -= width
		state.remaining_width = maxG(state.remaining_width, 0)
		line.insert_run(state, item, true)

		return brALL_FIT
	} else {
		//    int num_chars = item.num_chars;
		//    int break_num_chars = num_chars;
		//    int orig_width = width;

		if processing_new_item {
			glyph_item := GlyphItem{Item: item, Glyphs: state.glyphs}
			state.log_widths = glyph_item.GetLogicalWidths(layout.Text)
		}

	retry_break:

		// See how much of the item we can stuff in the line.
		width = 0
		var (
			break_width                    = width
			orig_width                     = width
			retrying_with_char_breaks      = false
			break_extra_width, extra_width GlyphUnit
			num_chars                      int
			break_num_chars                = item.Length
		)
		for num_chars = 0; num_chars < item.Length; num_chars++ {
			if width+extra_width > state.remaining_width && break_num_chars < item.Length {
				break
			}

			// If there are no previous runs we have to take care to grab at least one char.
			if layout.canBreakAt(state.startOffset+num_chars, retrying_with_char_breaks) &&
				(num_chars > 0 || line.Runs != nil) {
				break_num_chars = num_chars
				break_width = width
				break_extra_width = extra_width

				extra_width = layout.find_break_extra_width(state, num_chars)
			} else {
				extra_width = 0
			}

			width += state.log_widths[state.log_widths_offset+num_chars]
		}

		// If there's a space at the end of the line, include that also.
		// The logic here should match zero_line_final_space().
		// XXX Currently it doesn't quite match the logic there.  We don't check
		// the cluster here.  But should be fine in practice.
		if break_num_chars > 0 && break_num_chars < item.Length &&
			layout.logAttrs[state.startOffset+break_num_chars-1].IsWhite() {
			break_width -= state.log_widths[state.log_widths_offset+break_num_chars-1]
		}

		if layout.wrap == WRAP_WORD_CHAR && forceFit && break_width+break_extra_width > state.remaining_width && !retrying_with_char_breaks {
			retrying_with_char_breaks = true
			num_chars = item.Length
			width = orig_width
			break_num_chars = num_chars
			break_width = width
			goto retry_break
		}

		if forceFit || break_width+break_extra_width <= state.remaining_width /* Successfully broke the item */ {
			if state.remaining_width >= 0 {
				state.remaining_width -= break_width
				state.remaining_width = maxG(state.remaining_width, 0)
			}

			if break_num_chars == item.Length {
				if layout.break_needs_hyphen(state, break_num_chars) {
					item.Analysis.Flags |= AFNeedHyphen
				}
				line.insert_run(state, item, true)

				return brALL_FIT
			} else if break_num_chars == 0 {
				return brEMPTY_FIT
			} else {
				new_item := item.pango_item_split(break_num_chars)

				if layout.break_needs_hyphen(state, break_num_chars) {
					new_item.Analysis.Flags |= AFNeedHyphen
				}
				/* Add the width back, to the line, reshape, subtract the new width */
				state.remaining_width += break_width
				line.insert_run(state, new_item, false)
				break_width = line.Runs.Data.Glyphs.getWidth()
				state.remaining_width -= break_width

				state.log_widths_offset += break_num_chars

				// shaped items should never be broken
				if debugMode {
					assert(!shape_set, "processItem: break")
				}

				return brSOME_FIT
			}
		} else {
			state.glyphs = nil
			state.log_widths = nil
			return brNONE_FIT
		}
	}
}

func (layout *Layout) shouldEllipsizeCurrentLine(state *paraBreakState) bool {
	if layout.ellipsize == ELLIPSIZE_NONE || layout.Width < 0 {
		return false
	}

	if layout.Height >= 0 {
		/* state.remaining_height is height of layout left */

		/* if we can't stuff two more lines at the current guess of line height,
		* the line we are going to produce is going to be the last line */
		return state.line_height*2 > state.remaining_height
	} else {
		/* -layout.height is number of lines per paragraph to show */
		return state.line_of_par == int(-layout.Height)
	}
}

// the hard work begins here !
func (layout *Layout) process_line(state *paraBreakState) {
	//    line *LayoutLine;
	var (
		haveBreak           = false   /* If we've seen a possible break yet */
		breakRemainingWidth GlyphUnit /* Remaining width before adding run with break */
		breakStartOffset    = 0       /* Start offset before adding run with break */
		breakLink           *RunList  /* Link holding run before break */
		wrapped             = false   /* If we had to wrap the line */
	)

	line := layout.pango_layout_line_new()
	line.StartIndex = state.lineStartIndex
	line.IsParagraphStart = state.line_of_par == 1
	line.setResolvedDir(state.base_dir)

	state.lineWidth = layout.Width
	if state.lineWidth >= 0 && layout.alignment != ALIGN_CENTER {
		if line.IsParagraphStart && layout.Indent >= 0 {
			state.lineWidth -= layout.Indent
		} else if !line.IsParagraphStart && layout.Indent < 0 {
			state.lineWidth += layout.Indent
		}
		if state.lineWidth < 0 {
			state.lineWidth = 0
		}
	}

	if layout.shouldEllipsizeCurrentLine(state) {
		state.remaining_width = -1
	} else {
		state.remaining_width = state.lineWidth
	}

	if debugMode {
		showDebug("starting to fill line\n", line, state)
	}

	for state.items != nil {
		item := state.items.Data

		oldNumChars := item.Length
		oldRemainingWidth := state.remaining_width
		firstItemInLine := line.Runs != nil

		result := layout.process_item(line, state, !haveBreak, false)

		switch result {
		case brALL_FIT:
			if layout.canBreakIn(state.startOffset, oldNumChars, firstItemInLine) {
				haveBreak = true
				breakRemainingWidth = oldRemainingWidth
				breakStartOffset = state.startOffset
				breakLink = line.Runs.Next
			}
			state.items = state.items.Next
			state.startOffset += oldNumChars
		case brEMPTY_FIT:
			wrapped = true
			goto done
		case brSOME_FIT:
			state.startOffset += oldNumChars - item.Length
			wrapped = true
			goto done
		case brNONE_FIT:
			/* Back up over unused runs to run where there is a break */
			for line.Runs != nil && line.Runs != breakLink {
				state.items = &ItemList{Data: line.uninsert_run(), Next: state.items}
			}

			state.startOffset = breakStartOffset
			state.remaining_width = breakRemainingWidth

			/* Reshape run to break */
			item = state.items.Data

			oldNumChars = item.Length
			result = layout.process_item(line, state, true, true)
			if debugMode {
				assert(result == brSOME_FIT || result == brEMPTY_FIT, "processLines: break")
			}

			state.startOffset += oldNumChars - item.Length

			wrapped = true
			goto done
		case brLINE_SEPARATOR:
			state.items = state.items.Next
			state.startOffset += oldNumChars
			// A line-separate is just a forced break. Set wrapped, so we do justification
			wrapped = true
			goto done
		}
	}

done:
	line.postprocess(state, wrapped)
	line.addLine(state)
	state.line_of_par++
	state.lineStartIndex += line.Length
	state.lineStartOffset = state.startOffset
}

// logAttrs must have at least length: length+1
func getItemsLogAttrs(text []rune, start, length int, items *ItemList, logAttrs []CharAttr, attrs AttrList) {
	pangoDefaultBreak(text[start:start+length], logAttrs)
	offset := 0
	for l := items; l != nil; l = l.Next {
		item := l.Data

		if debugMode {
			assert(item.Offset <= start+length,
				fmt.Sprintf("expected item.Offset <= start+length, got %d > %d", item.Offset, start+length))
			assert(item.Length <= (start+length)-item.Offset,
				fmt.Sprintf("expected item.Length <= (start+length)-item.Offset, got %d > %d", item.Length, (start+length)-item.Offset))
		}

		pangoTailorBreak(text[item.Offset:item.Offset+item.Length],
			&item.Analysis, -1, logAttrs[offset:offset+item.Length+1])
		offset += item.Length
	}

	if items != nil {
		pango_attr_break(text[start:], attrs, items.Data.Offset, logAttrs)
	}
}

// ApplyAttributes apply `attrs` to the list.
func (items *ItemList) ApplyAttributes(attrs AttrList) {
	if attrs == nil {
		return
	}

	iter := attrs.getIterator()

	for l := items; l != nil; l = l.Next {
		l.Data.pango_item_apply_attrs(iter)
	}
}

func (layout *Layout) apply_attributes_to_runs(attrs AttrList) {
	if attrs == nil {
		return
	}

	for _, line := range layout.lines {
		old_runs := line.Runs.reverse()
		line.Runs = nil
		for rl := old_runs; rl != nil; rl = rl.Next {
			glyph_item := rl.Data

			new_runs := glyph_item.pango_glyph_item_apply_attrs(layout.Text, attrs)

			line.Runs = new_runs.concat(line.Runs)
		}
	}
}

// performs the actual layout
func (layout *Layout) checkLines() {
	var (
		itemizeAttrs, shapeAttrs AttrList
		iter                     attrIterator
		state                    paraBreakState
		prevBaseDir              = DIRECTION_NEUTRAL
		baseDir                  = DIRECTION_NEUTRAL
		needLogAttrs             bool
	)

	layout.check_context_changed()

	if len(layout.lines) != 0 {
		return
	}

	attrs := layout.pango_layout_get_effective_attributes()
	if attrs != nil {
		shapeAttrs = attrs.Filter(affects_break_or_shape)
		itemizeAttrs = attrs.Filter(affects_itemization)
		if itemizeAttrs != nil {
			iter = *itemizeAttrs.getIterator()
		}
	}

	start := 0 // index in text

	// Find the first strong direction of the text
	if layout.autoDir {
		prevBaseDir = findBaseDirection(layout.Text)
		if prevBaseDir == DIRECTION_NEUTRAL {
			prevBaseDir = layout.context.baseDir
		}
	} else {
		baseDir = layout.context.baseDir
	}

	// these are only used if layout.height >= 0
	state.remaining_height = layout.Height
	state.line_height = -1
	if layout.Height >= 0 {
		var logical Rectangle
		height := layout.getEmptyExtentsAndHeightAt(0, &logical)
		if layout.LineSpacing == 0 {
			state.line_height = logical.Height
		} else {
			state.line_height = GlyphUnit(layout.LineSpacing * float32(height))
		}
	}

	if layout.logAttrs == nil {
		layout.logAttrs = make([]CharAttr, len(layout.Text)+1)
		needLogAttrs = true
	} else {
		needLogAttrs = false
	}

	for done := false; !done; {
		var delimiterIndex, nextParaIndex int

		if layout.singleParagraph {
			delimiterIndex = len(layout.Text)
			nextParaIndex = len(layout.Text)
		} else {
			delimiterIndex, nextParaIndex = pango_find_paragraph_boundary(layout.Text[start:])
		}

		if debugMode {
			assert(nextParaIndex >= delimiterIndex,
				fmt.Sprintf("checkLines nextParaIndex: %d %d", nextParaIndex, delimiterIndex))
		}

		if layout.autoDir {
			baseDir = findBaseDirection(layout.Text[start : start+delimiterIndex])

			/* Propagate the base direction for neutral paragraphs */
			if baseDir == DIRECTION_NEUTRAL {
				baseDir = prevBaseDir
			} else {
				prevBaseDir = baseDir
			}
		}

		end := start + delimiterIndex // index into text

		delimLen := nextParaIndex - delimiterIndex

		if end == len(layout.Text) {
			done = true
		}

		if debugMode {
			assert(end <= len(layout.Text) && start <= len(layout.Text), "checkLines")
			// PS is 3 bytes
			assert(delimLen < 4 && delimLen >= 0, fmt.Sprintf("checkLines delimLen: %d", delimLen))
		}

		var cachedIter *attrIterator
		if itemizeAttrs != nil {
			cachedIter = &iter
		}
		state.attrs = itemizeAttrs

		if debugMode {
			fmt.Println("Itemizing...")
		}
		state.items = layout.context.itemizeWithBaseDir(
			baseDir,
			layout.Text,
			start, end-start,
			itemizeAttrs, cachedIter)

		if debugMode {
			fmt.Println("Applying attributes...")
		}
		state.items.ApplyAttributes(shapeAttrs)

		if needLogAttrs {
			if debugMode {
				fmt.Println("Computing logical attributes...")
			}

			getItemsLogAttrs(layout.Text, start, delimiterIndex+delimLen,
				state.items, layout.logAttrs[start:], shapeAttrs)
		}

		state.base_dir = baseDir
		state.line_of_par = 1
		state.startOffset = start
		state.lineStartOffset = start
		state.lineStartIndex = start

		state.glyphs = nil
		state.log_widths = nil

		/* for deterministic bug hunting's sake set everything! */
		state.lineWidth = -1
		state.remaining_width = -1
		state.log_widths_offset = 0

		state.hyphen_width = -1

		if state.items != nil {
			for state.items != nil {
				if debugMode {
					fmt.Println("Processing line...")
				}
				layout.process_line(&state)
			}
		} else {
			empty_line := layout.pango_layout_line_new()
			empty_line.StartIndex = state.lineStartIndex
			empty_line.IsParagraphStart = true
			empty_line.setResolvedDir(baseDir)
			empty_line.addLine(&state)
		}

		if layout.Height >= 0 && state.remaining_height < state.line_height {
			done = true
		}

		start = end + delimLen
	}

	layout.apply_attributes_to_runs(attrs)
	reverseLines(layout.lines)
}

func reverseLines(arr []*LayoutLine) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
}

func reverseItems(arr []*Item) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
}

//  /**
//   * pango_layout_line_ref:
//   * @line: (nullable): a #LayoutLine, may be %nil
//   *
//   * Increase the reference count of a #LayoutLine by one.
//   *
//   * Return value: the line passed in.
//   *
//   * Since: 1.10
//   **/
//  LayoutLine *
//  pango_layout_line_ref (line *LayoutLine)
//  {
//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;

//    if (line == nil)
// 	 return nil;

//    g_atomic_int_inc ((int *) &private.ref_count);

//    return line;
//  }

//  /**
//   * pango_layout_line_unref:
//   * @line: a #LayoutLine
//   *
//   * Decrease the reference count of a #LayoutLine by one.
//   * If the result is zero, the line and all associated memory
//   * will be freed.
//   **/
//  void
//  pango_layout_line_unref (line *LayoutLine)
//  {
//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;

//    if (line == nil)
// 	 return;

//    g_return_if_fail (private.ref_count > 0);

//    if (g_atomic_int_dec_and_test ((int *) &private.ref_count))
// 	 {
// 	   g_slist_foreach (line.runs, (GFunc)free_run, GINT_TO_POINTER (1));
// 	   g_slist_free (line.runs);
// 	   g_slice_free (LayoutLinePrivate, private);
// 	 }
//  }

//  G_DEFINE_BOXED_TYPE (LayoutLine, pango_layout_line,
// 					  pango_layout_line_ref,
// 					  pango_layout_line_unref);

//  /**
//   * pango_layout_line_x_to_index:
//   * @line:      a #LayoutLine
//   * @x_pos:     the X offset (in Pango units)
//   *             from the left edge of the line.
//   * @index_: (out):   location to store calculated byte index for
//   *                   the grapheme in which the user clicked.
//   * @trailing: (out): location to store an integer indicating where
//   *                   in the grapheme the user clicked. It will either
//   *                   be zero, or the number of characters in the
//   *                   grapheme. 0 represents the leading edge of the grapheme.
//   *
//   * Converts from x offset to the byte index of the corresponding
//   * character within the text of the layout. If @x_pos is outside the line,
//   * @index_ and @trailing will point to the very first or very last position
//   * in the line. This determination is based on the resolved direction
//   * of the paragraph; for example, if the resolved direction is
//   * right-to-left, then an X position to the right of the line (after it)
//   * results in 0 being stored in @index_ and @trailing. An X position to the
//   * left of the line results in @index_ pointing to the (logical) last
//   * grapheme in the line and @trailing being set to the number of characters
//   * in that grapheme. The reverse is true for a left-to-right line.
//   *
//   * Return value: %false if @x_pos was outside the line, `true` if inside
//   **/
//  bool
//  pango_layout_line_x_to_index (line *LayoutLine,
// 				   int              x_pos,
// 				   int             *index,
// 				   int             *trailing)
//  {
//    GSList *tmp_list;
//    gint start_pos = 0;
//    gint first_index = 0; /* line.startIndex */
//    gint first_offset;
//    gint last_index;      /* start of last grapheme in line */
//    gint last_offset;
//    gint end_index;       /* end iterator for line */
//    gint end_offset;      /* end iterator for line */
//    layout *Layout;
//    gint last_trailing;
//    bool suppress_last_trailing;

//    g_return_val_if_fail (LINE_IS_VALID (line), false);

//    layout = line.layout;

//    /* Find the last index in the line
// 	*/
//    first_index = line.startIndex;

//    if (line.length == 0)
// 	 {
// 	   if (index)
// 	 *index = first_index;
// 	   if (trailing)
// 	 *trailing = 0;

// 	   return false;
// 	 }

//    assert (line.length > 0);

//    first_offset = g_utf8_pointer_to_offset (layout.text, layout.text + line.startIndex);

//    end_index = first_index + line.length;
//    end_offset = first_offset + g_utf8_pointer_to_offset (layout.text + first_index, layout.text + end_index);

//    last_index = end_index;
//    last_offset = end_offset;
//    last_trailing = 0;
//    do
// 	 {
// 	   last_index = g_utf8_prev_char (layout.text + last_index) - layout.text;
// 	   last_offset--;
// 	   last_trailing++;
// 	 }
//    for (last_offset > first_offset && !layout.log_attrs[last_offset].is_cursor_position);

//    /* This is a HACK. If a program only keeps track of cursor (etc)
// 	* indices and not the trailing flag, then the trailing index of the
// 	* last character on a wrapped line is identical to the leading
// 	* index of the next line. So, we fake it and set the trailing flag
// 	* to zero.
// 	*
// 	* That is, if the text is "now is the time", and is broken between
// 	* 'now' and 'is'
// 	*
// 	* Then when the cursor is actually at:
// 	*
// 	* n|o|w| |i|s|
// 	*              ^
// 	* we lie and say it is at:
// 	*
// 	* n|o|w| |i|s|
// 	*            ^
// 	*
// 	* So the cursor won't appear on the next line before 'the'.
// 	*
// 	* Actually, any program keeping cursor
// 	* positions with wrapped lines should distinguish leading and
// 	* trailing cursors.
// 	*/
//    tmp_list = layout.lines;
//    for (tmp_list.data != line)
// 	 tmp_list = tmp_list.next;

//    if (tmp_list.next &&
// 	   line.startIndex + line.length == ((LayoutLine *)tmp_list.next.data).startIndex)
// 	 suppress_last_trailing = true;
//    else
// 	 suppress_last_trailing = false;

//    if (x_pos < 0)
// 	 {
// 	   /* pick the leftmost char */
// 	   if (index)
// 	 *index = (line.resolved_dir == PANGO_DIRECTION_LTR) ? first_index : last_index;
// 	   /* and its leftmost edge */
// 	   if (trailing)
// 	 *trailing = (line.resolved_dir == PANGO_DIRECTION_LTR || suppress_last_trailing) ? 0 : last_trailing;

// 	   return false;
// 	 }

//    tmp_list = line.runs;
//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;
// 	   int logical_width;

// 	   logical_width = getWidth (run.glyphs);

// 	   if (x_pos >= start_pos && x_pos < start_pos + logical_width)
// 	 {
// 	   int offset;
// 	   bool char_trailing;
// 	   int grapheme_start_index;
// 	   int grapheme_start_offset;
// 	   int grapheme_end_offset;
// 	   int pos;
// 	   int char_index;

// 	   pango_glyph_string_x_to_index (run.glyphs,
// 					  layout.text + run.item.offset, run.item.length,
// 					  &run.item.analysis,
// 					  x_pos - start_pos,
// 					  &pos, &char_trailing);

// 	   char_index = run.item.offset + pos;

// 	   /* Convert from characters to graphemes */

// 	   offset = g_utf8_pointer_to_offset (layout.text, layout.text + char_index);

// 	   grapheme_start_offset = offset;
// 	   grapheme_start_index = char_index;
// 	   for (grapheme_start_offset > first_offset &&
// 		  !layout.log_attrs[grapheme_start_offset].is_cursor_position)
// 		 {
// 		   grapheme_start_index = g_utf8_prev_char (layout.text + grapheme_start_index) - layout.text;
// 		   grapheme_start_offset--;
// 		 }

// 	   grapheme_end_offset = offset;
// 	   do
// 		 {
// 		   grapheme_end_offset++;
// 		 }
// 	   for (grapheme_end_offset < end_offset &&
// 		  !layout.log_attrs[grapheme_end_offset].is_cursor_position);

// 	   if (index)
// 		 *index = grapheme_start_index;

// 	   if (trailing)
// 		 {
// 		   if ((grapheme_end_offset == end_offset && suppress_last_trailing) ||
// 		   offset + char_trailing <= (grapheme_start_offset + grapheme_end_offset) / 2)
// 		 *trailing = 0;
// 		   else
// 		 *trailing = grapheme_end_offset - grapheme_start_offset;
// 		 }

// 	   return true;
// 	 }

// 	   start_pos += logical_width;
// 	   tmp_list = tmp_list.next;
// 	 }

//    /* pick the rightmost char */
//    if (index)
// 	 *index = (line.resolved_dir == PANGO_DIRECTION_LTR) ? last_index : first_index;

//    /* and its rightmost edge */
//    if (trailing)
// 	 *trailing = (line.resolved_dir == PANGO_DIRECTION_LTR && !suppress_last_trailing) ? last_trailing : 0;

//    return false;
//  }

//  static int
//  pango_layout_line_get_width (line *LayoutLine)
//  {
//    int width = 0;
//    GSList *tmp_list = line.runs;

//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;

// 	   width += getWidth (run.glyphs);

// 	   tmp_list = tmp_list.next;
// 	 }

//    return width;
//  }

//  static void
//  pango_layout_line_get_extents_and_height (line *LayoutLine,
// 										   Rectangle  *inkRect,
// 										   Rectangle  *logicalRect,
// 										   int             *height)
//  {
//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;
//    GSList *tmp_list;
//    int x_pos = 0;
//    bool caching = false;

//    g_return_if_fail (LINE_IS_VALID (line));

//    if (G_UNLIKELY (!inkRect && !logicalRect))
// 	 return;

//    switch (private.cache_status)
// 	 {
// 	 case CACHED:
// 	   {
// 	 if (inkRect)
// 	   *inkRect = private.ink_rect;
// 	 if (logicalRect)
// 	   *logicalRect = private.logicalRect;
// 		 if (height)
// 		   *height = private.height;
// 	 return;
// 	   }
// 	 case NOT_CACHED:
// 	   {
// 	 caching = true;
// 	 if (!inkRect)
// 	   inkRect = &private.ink_rect;
// 	 if (!logicalRect)
// 	   logicalRect = &private.logicalRect;
// 		 if (!height)
// 		   height = &private.height;
// 	 break;
// 	   }
// 	 case LEAKED:
// 	   {
// 	 break;
// 	   }
// 	 }

//    if (inkRect)
// 	 {
// 	   inkRect.x = 0;
// 	   inkRect.y = 0;
// 	   inkRect.width = 0;
// 	   inkRect.height = 0;
// 	 }

//    if (logicalRect)
// 	 {
// 	   logicalRect.x = 0;
// 	   logicalRect.y = 0;
// 	   logicalRect.width = 0;
// 	   logicalRect.height = 0;
// 	 }

//    if (height)
// 	 *height = 0;

//    tmp_list = line.runs;
//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;
// 	   int newPos;
// 	   Rectangle run_ink;
// 	   Rectangle run_logical;
// 	   int run_height;

// 	   getExtentsAndHeight (run,
// 												inkRect ? &run_ink : nil,
// 												&run_logical,
// 												height ? &run_height : nil);

// 	   if (inkRect)
// 	 {
// 	   if (inkRect.width == 0 || inkRect.height == 0)
// 		 {
// 		   *inkRect = run_ink;
// 		   inkRect.x += x_pos;
// 		 }
// 	   else if (run_ink.width != 0 && run_ink.height != 0)
// 		 {
// 		   newPos = MIN (inkRect.x, x_pos + run_ink.x);
// 		   inkRect.width = MAX (inkRect.x + inkRect.width,
// 					  x_pos + run_ink.x + run_ink.width) - newPos;
// 		   inkRect.x = newPos;

// 		   newPos = MIN (inkRect.y, run_ink.y);
// 		   inkRect.height = MAX (inkRect.y + inkRect.height,
// 					   run_ink.y + run_ink.height) - newPos;
// 		   inkRect.y = newPos;
// 		 }
// 	 }

// 	   if (logicalRect)
// 		 {
// 		   newPos = MIN (logicalRect.x, x_pos + run_logical.x);
// 		   logicalRect.width = MAX (logicalRect.x + logicalRect.width,
// 					  x_pos + run_logical.x + run_logical.width) - newPos;
// 	   logicalRect.x = newPos;

// 	   newPos = MIN (logicalRect.y, run_logical.y);
// 	   logicalRect.height = MAX (logicalRect.y + logicalRect.height,
// 					   run_logical.y + run_logical.height) - newPos;
// 	   logicalRect.y = newPos;
// 		 }

// 	   if (height)
// 		 *height = MAX (*height, run_height);

// 	   x_pos += run_logical.width;
// 	   tmp_list = tmp_list.next;
// 	 }

//    if (logicalRect && !line.runs)
// 	 getEmptyExtentsAndHeight (line, logicalRect);

//    if (caching)
// 	 {
// 	   if (&private.ink_rect != inkRect)
// 	 private.ink_rect = *inkRect;
// 	   if (&private.logicalRect != logicalRect)
// 	 private.logicalRect = *logicalRect;
// 	   if (&private.height != height)
// 		 private.height = *height;
// 	   private.cache_status = CACHED;
// 	 }
//  }

//  /**
//   * GetExtents:
//   * @line:     a #LayoutLine
//   * @inkRect: (out) (allow-none): rectangle used to store the extents of
//   *            the glyph string as drawn, or %nil
//   * @logicalRect: (out) (allow-none):rectangle used to store the logical
//   *                extents of the glyph string, or %nil
//   *
//   * Computes the logical and ink extents of a layout line. See
//   * Font.getGlyphExtents() for details about the interpretation
//   * of the rectangles.
//   */
//  void
//  GetExtents (line *LayoutLine,
// 					Rectangle  *inkRect,
// 					Rectangle  *logicalRect)
//  {
//    pango_layout_line_get_extents_and_height (line, inkRect, logicalRect, nil);
//  }

//  /**
//   * pango_layout_line_get_height:
//   * @line:     a #LayoutLine
//   * @height: (out) (allow-none): return location for the line height
//   *
//   * Computes the height of the line, ie the distance between
//   * this and the previous lines baseline.
//   *
//   * Since: 1.44
//   */
//  void
//  pango_layout_line_get_height (line *LayoutLine,
// 				   int             *height)
//  {
//    pango_layout_line_get_extents_and_height (line, nil, nil, height);
//  }

//  static int
//  get_item_letter_spacing (PangoItem *item)
//  {
//    ItemProperties properties;

//    pango_layout_get_item_properties (item, &properties);

//    return properties.letter_spacing;
//  }

//  static void
//  pad_glyphstring_right (PangoGlyphString *glyphs,
// 				ParaBreakState   *state,
// 				int               adjustment)
//  {
//    int glyph = glyphs.num_glyphs - 1;

//    for (glyph >= 0 && glyphs.glyphs[glyph].geometry.width == 0)
// 	 glyph--;

//    if (glyph < 0)
// 	 return;

//    state.remaining_width -= adjustment;
//    glyphs.glyphs[glyph].geometry.width += adjustment;
//    if (glyphs.glyphs[glyph].geometry.width < 0)
// 	 {
// 	   state.remaining_width += glyphs.glyphs[glyph].geometry.width;
// 	   glyphs.glyphs[glyph].geometry.width = 0;
// 	 }
//  }

//  static void
//  pad_glyphstring_left (PangoGlyphString *glyphs,
// 			   ParaBreakState   *state,
// 			   int               adjustment)
//  {
//    int glyph = 0;

//    for (glyph < glyphs.num_glyphs && glyphs.glyphs[glyph].geometry.width == 0)
// 	 glyph++;

//    if (glyph == glyphs.num_glyphs)
// 	 return;

//    state.remaining_width -= adjustment;
//    glyphs.glyphs[glyph].geometry.width += adjustment;
//    glyphs.glyphs[glyph].geometry.xOffset += adjustment;
//  }

func (layout *Layout) is_tab_run(run *GlyphItem) bool {
	return layout.Text[run.Item.Offset] == '\t'
}

//  /* When doing shaping, we add the letter spacing value for a
//   * run after every grapheme in the run. This produces ugly
//   * asymmetrical results, so what this routine is redistributes
//   * that space to the beginning and the end of the run.
//   *
//   * We also trim the letter spacing from runs adjacent to
//   * tabs and from the outside runs of the lines so that things
//   * line up properly. The line breaking and tab positioning
//   * were computed without this trimming so they are no longer
//   * exactly correct, but this won't be very noticeable in most
//   * cases.
//   */
//  static void
//  adjust_line_letter_spacing (line *LayoutLine,
// 				 state *ParaBreakState )
//  {
//    layout *Layout = line.layout;
//    bool reversed;
//    GlyphItem *last_run;
//    int tab_adjustment;
//    GSList *l;

//    /* If we have tab stops and the resolved direction of the
// 	* line is RTL, then we need to walk through the line
// 	* in reverse direction to figure out the corrections for
// 	* tab stops.
// 	*/
//    reversed = false;
//    if (line.resolved_dir == PANGO_DIRECTION_RTL)
// 	 {
// 	   for (l = line.runs; l; l = l.next)
// 	 if (is_tab_run (layout, l.data))
// 	   {
// 		 line.runs = g_slist_reverse (line.runs);
// 		 reversed = true;
// 		 break;
// 	   }
// 	 }

//    /* Walk over the runs in the line, redistributing letter
// 	* spacing from the end of the run to the start of the
// 	* run and trimming letter spacing from the ends of the
// 	* runs adjacent to the ends of the line or tab stops.
// 	*
// 	* We accumulate a correction factor from this trimming
// 	* which we add onto the next tab stop space to keep the
// 	* things properly aligned.
// 	*/

//    last_run = nil;
//    tab_adjustment = 0;
//    for (l = line.runs; l; l = l.next)
// 	 {
// 	   GlyphItem *run = l.data;
// 	   GlyphItem *next_run = l.next ? l.next.data : nil;

// 	   if (is_tab_run (layout, run))
// 	 {
// 	   pad_glyphstring_right (run.glyphs, state, tab_adjustment);
// 	   tab_adjustment = 0;
// 	 }
// 	   else
// 	 {
// 	   GlyphItem *visual_next_run = reversed ? last_run : next_run;
// 	   GlyphItem *visual_last_run = reversed ? next_run : last_run;
// 	   int run_spacing = get_item_letter_spacing (run.item);
// 	   int space_left, space_right;

// 	   distributeLetterSpacing (run_spacing, &space_left, &space_right);

// 	   if (run.glyphs.glyphs[0].geometry.width == 0)
// 		 {
// 		   /* we've zeroed this space glyph at the end of line, now remove
// 			* the letter spacing added to its adjacent glyph */
// 		   pad_glyphstring_left (run.glyphs, state, - space_left);
// 		 }
// 	   else if (!visual_last_run || is_tab_run (layout, visual_last_run))
// 		 {
// 		   pad_glyphstring_left (run.glyphs, state, - space_left);
// 		   tab_adjustment += space_left;
// 		 }

// 	   if (run.glyphs.glyphs[run.glyphs.num_glyphs - 1].geometry.width == 0)
// 		 {
// 		   /* we've zeroed this space glyph at the end of line, now remove
// 			* the letter spacing added to its adjacent glyph */
// 		   pad_glyphstring_right (run.glyphs, state, - space_right);
// 		 }
// 	   else if (!visual_next_run || is_tab_run (layout, visual_next_run))
// 		 {
// 		   pad_glyphstring_right (run.glyphs, state, - space_right);
// 		   tab_adjustment += space_right;
// 		 }
// 	 }

// 	   last_run = run;
// 	 }

//    if (reversed)
// 	 line.runs = g_slist_reverse (line.runs);
//  }

//  static void
//  justify_clusters (line *LayoutLine,
// 		   state *ParaBreakState )
//  {
//    const gchar *text = line.layout.text;
//    const PangoLogAttr *log_attrs = line.layout.log_attrs;

//    int total_remaining_width, total_gaps = 0;
//    int added_so_far, gaps_so_far;
//    bool is_hinted;
//    GSList *run_iter;
//    enum {
// 	 MEASURE,
// 	 ADJUST
//    } mode;

//    total_remaining_width = state.remaining_width;
//    if (total_remaining_width <= 0)
// 	 return;

//    /* hint to full pixel if total remaining width was so */
//    is_hinted = (total_remaining_width & (Scale - 1)) == 0;

//    for (mode = MEASURE; mode <= ADJUST; mode++)
// 	 {
// 	   bool leftedge = true;
// 	   PangoGlyphString *rightmost_glyphs = nil;
// 	   int rightmost_space = 0;
// 	   int residual = 0;

// 	   added_so_far = 0;
// 	   gaps_so_far = 0;

// 	   for (run_iter = line.runs; run_iter; run_iter = run_iter.next)
// 	 {
// 	   GlyphItem *run = run_iter.data;
// 	   PangoGlyphString *glyphs = run.glyphs;
// 	   PangoGlyphItemIter cluster_iter;
// 	   bool have_cluster;
// 	   int dir;
// 	   int offset;

// 	   dir = run.item.analysis.level % 2 == 0 ? +1 : -1;

// 	   /* We need character offset of the start of the run.  We don't have this.
// 		* Compute by counting from the beginning of the line.  The naming is
// 		* confusing.  Note that:
// 		*
// 		* run.item.offset        is byte offset of start of run in layout.text.
// 		* state.line_start_index  is byte offset of start of line in layout.text.
// 		* state.line_start_offset is character offset of start of line in layout.text.
// 		*/
// 	   assert (run.item.offset >= state.line_start_index);
// 	   offset = state.line_start_offset
// 		  + pango_utf8_strlen (text + state.line_start_index,
// 					   run.item.offset - state.line_start_index);

// 	   for (have_cluster = dir > 0 ?
// 		  InitStart (&cluster_iter, run, text) :
// 		  InitEnd   (&cluster_iter, run, text);
// 			have_cluster;
// 			have_cluster = dir > 0 ?
// 			  NextCluster (&cluster_iter) :
// 			  PrevCluster (&cluster_iter))
// 		 {
// 		   int i;
// 		   int width = 0;

// 		   /* don't expand in the middle of graphemes */
// 		   if (!log_attrs[offset + cluster_iter.start_char].is_cursor_position)
// 		 continue;

// 		   for (i = cluster_iter.start_glyph; i != cluster_iter.end_glyph; i += dir)
// 		 width += glyphs.glyphs[i].geometry.width;

// 		   /* also don't expand zero-width clusters. */
// 		   if (width == 0)
// 		 continue;

// 		   gaps_so_far++;

// 		   if (mode == ADJUST)
// 		 {
// 		   int leftmost, rightmost;
// 		   int adjustment, space_left, space_right;

// 		   adjustment = total_remaining_width / total_gaps + residual;
// 		   if (is_hinted)
// 		   {
// 			 int old_adjustment = adjustment;
// 			 adjustment = PANGO_UNITS_ROUND (adjustment);
// 			 residual = old_adjustment - adjustment;
// 		   }
// 		   /* distribute to before/after */
// 		   distributeLetterSpacing (adjustment, &space_left, &space_right);

// 		   if (cluster_iter.start_glyph < cluster_iter.end_glyph)
// 		   {
// 			 /* LTR */
// 			 leftmost  = cluster_iter.start_glyph;
// 			 rightmost = cluster_iter.end_glyph - 1;
// 		   }
// 		   else
// 		   {
// 			 /* RTL */
// 			 leftmost  = cluster_iter.end_glyph + 1;
// 			 rightmost = cluster_iter.start_glyph;
// 		   }
// 		   /* Don't add to left-side of left-most glyph of left-most non-zero run. */
// 		   if (leftedge)
// 			 leftedge = false;
// 		   else
// 		   {
// 			 glyphs.glyphs[leftmost].geometry.width    += space_left ;
// 			 glyphs.glyphs[leftmost].geometry.xOffset += space_left ;
// 			 added_so_far += space_left;
// 		   }
// 		   /* Don't add to right-side of right-most glyph of right-most non-zero run. */
// 		   {
// 			 /* Save so we can undo later. */
// 			 rightmost_glyphs = glyphs;
// 			 rightmost_space = space_right;

// 			 glyphs.glyphs[rightmost].geometry.width  += space_right;
// 			 added_so_far += space_right;
// 		   }
// 		 }
// 		 }
// 	 }

// 	   if (mode == MEASURE)
// 	 {
// 	   total_gaps = gaps_so_far - 1;

// 	   if (total_gaps == 0)
// 		 {
// 		   /* a single cluster, can't really justify it */
// 		   return;
// 		 }
// 	 }
// 	   else /* mode == ADJUST */
// 		 {
// 	   if (rightmost_glyphs)
// 		{
// 		  rightmost_glyphs.glyphs[rightmost_glyphs.num_glyphs - 1].geometry.width -= rightmost_space;
// 		  added_so_far -= rightmost_space;
// 		}
// 	 }
// 	 }

//    state.remaining_width -= added_so_far;
//  }

//  /**
//   * pango_layout_iter_copy:
//   * @iter: (nullable): a #LayoutIter, may be %nil
//   *
//   * Copies a #LayoutIter.
//   *
//   * Return value: (nullable): the newly allocated #LayoutIter,
//   *               which should be freed with pango_layout_iter_free(),
//   *               or %nil if @iter was %nil.
//   *
//   * Since: 1.20
//   **/
//  LayoutIter *
//  pango_layout_iter_copy (LayoutIter *iter)
//  {
//    LayoutIter *new;

//    if (iter == nil)
// 	 return nil;

//    new = g_slice_new (LayoutIter);

//    new.layout = g_object_ref (iter.layout);
//    new.lines = iter.lines;
//    new.line = iter.line;
//    pango_layout_line_ref (new.line);

//    new.run_list_link = iter.run_list_link;
//    new.run = iter.run;
//    new.index = iter.index;

//    new.line_extents = nil;
//    if (iter.line_extents != nil)
// 	 {
// 	   new.line_extents = g_memdup (iter.line_extents,
// 									 iter.layout.line_count * sizeof (extents));

// 	 }
//    new.lineIndex = iter.lineIndex;

//    new.run_x = iter.run_x;
//    new.run_width = iter.run_width;
//    new.ltr = iter.ltr;

//    new.cluster_x = iter.cluster_x;
//    new.cluster_width = iter.cluster_width;

//    new.cluster_start = iter.cluster_start;
//    new.next_cluster_glyph = iter.next_cluster_glyph;

//    new.cluster_num_chars = iter.cluster_num_chars;
//    new.character_position = iter.character_position;

//    new.layout_width = iter.layout_width;

//    return new;
//  }

//  G_DEFINE_BOXED_TYPE (LayoutIter, pango_layout_iter,
// 					  pango_layout_iter_copy,
// 					  pango_layout_iter_free);

//  void
//  _pango_layout_iter_destroy (LayoutIter *iter)
//  {
//    if (iter == nil)
// 	 return;

//    g_free (iter.line_extents);
//    pango_layout_line_unref (iter.line);
//    g_object_unref (iter.layout);
//  }

//  /**
//   * pango_layout_iter_free:
//   * @iter: (nullable): a #LayoutIter, may be %nil
//   *
//   * Frees an iterator that's no longer in use.
//   **/
//  void
//  pango_layout_iter_free (LayoutIter *iter)
//  {
//    if (iter == nil)
// 	 return;

//    _pango_layout_iter_destroy (iter);
//    g_slice_free (LayoutIter, iter);
//  }

//  /**
//   * pango_layout_iter_get_run_readonly:
//   * @iter: a #LayoutIter
//   *
//   * Gets the current run. When iterating by run, at the end of each
//   * line, there's a position with a %nil run, so this function can return
//   * %nil. The %nil run at the end of each line ensures that all lines have
//   * at least one run, even lines consisting of only a newline.
//   *
//   * This is a faster alternative to GetRun(),
//   * but the user is not expected
//   * to modify the contents of the run (glyphs, glyph widths, etc.).
//   *
//   * Return value: (transfer none) (nullable): the current run, that
//   * should not be modified.
//   *
//   * Since: 1.16
//   **/
//  GlyphItem*
//  pango_layout_iter_get_run_readonly (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    leaked (iter.line);

//    return iter.run;
//  }

//  /* an inline-able version for local use */
//  static LayoutLine*
//  _pango_layout_iter_get_line (LayoutIter *iter)
//  {
//    return iter.line;
//  }

//  /**
//   * pango_layout_iter_get_line_readonly:
//   * @iter: a #LayoutIter
//   *
//   * Gets the current line for read-only access.
//   *
//   * This is a faster alternative to GetLine(),
//   * but the user is not expected
//   * to modify the contents of the line (glyphs, glyph widths, etc.).
//   *
//   * Return value: (transfer none): the current line, that should not be
//   * modified.
//   *
//   * Since: 1.16
//   **/
//  LayoutLine*
//  pango_layout_iter_get_line_readonly (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    return iter.line;
//  }

//  /**
//   * IsAtLastLine:
//   * @iter: a #LayoutIter
//   *
//   * Determines whether @iter is on the last line of the layout.
//   *
//   * Return value: `true` if @iter is on the last line.
//   **/
//  bool
//  IsAtLastLine (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return false;

//    return iter.lineIndex == iter.layout.line_count - 1;
//  }

//  /**
//   * pango_layout_iter_get_layout:
//   * @iter: a #LayoutIter
//   *
//   * Gets the layout associated with a #LayoutIter.
//   *
//   * Return value: (transfer none): the layout associated with @iter.
//   *
//   * Since: 1.20
//   **/
//  Layout*
//  pango_layout_iter_get_layout (LayoutIter *iter)
//  {
//    /* check is redundant as it simply checks that iter.layout is not nil */
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    return iter.layout;
//  }

//  /**
//   * pango_layout_iter_get_line_yrange:
//   * @iter: a #LayoutIter
//   * @y0_: (out) (allow-none): start of line, or %nil
//   * @y1_: (out) (allow-none): end of line, or %nil
//   *
//   * Divides the vertical space in the #Layout being iterated over
//   * between the lines in the layout, and returns the space belonging to
//   * the current line.  A line's range includes the line's logical
//   * extents, plus half of the spacing above and below the line, if
//   * SetSpacing() has been called to set layout spacing.
//   * The Y positions are in layout coordinates (origin at top left of the
//   * entire layout).
//   *
//   * Note: Since 1.44, Pango uses line heights for placing lines,
//   * and there may be gaps between the ranges returned by this
//   * function.
//   */
//  void
//  pango_layout_iter_get_line_yrange (LayoutIter *iter,
// 					int             *y0,
// 					int             *y1)
//  {
//    const extents *ext;
//    int half_spacing;

//    if (ITER_IS_INVALID (iter))
// 	 return;

//    ext = &iter.line_extents[iter.lineIndex];

//    half_spacing = iter.layout.spacing / 2;

//    /* Note that if layout.spacing is odd, the remainder spacing goes
// 	* above the line (this is pretty arbitrary of course)
// 	*/

//    if (y0)
// 	 {
// 	   /* No spacing above the first line */

// 	   if (iter.lineIndex == 0)
// 	 *y0 = ext.logicalRect.y;
// 	   else
// 	 *y0 = ext.logicalRect.y - (iter.layout.spacing - half_spacing);
// 	 }

//    if (y1)
// 	 {
// 	   /* No spacing below the last line */
// 	   if (iter.lineIndex == iter.layout.line_count - 1)
// 	 *y1 = ext.logicalRect.y + ext.logicalRect.height;
// 	   else
// 	 *y1 = ext.logicalRect.y + ext.logicalRect.height + half_spacing;
// 	 }
//  }

//  /**
//   * pango_layout_iter_get_layout_extents:
//   * @iter: a #LayoutIter
//   * @inkRect: (out) (allow-none): rectangle to fill with ink extents,
//   *            or %nil
//   * @logicalRect: (out) (allow-none): rectangle to fill with logical
//   *                extents, or %nil
//   *
//   * Obtains the extents of the #Layout being iterated
//   * over. @inkRect or @logicalRect can be %nil if you
//   * aren't interested in them.
//   *
//   **/
//  void
//  pango_layout_iter_get_layout_extents  (LayoutIter *iter,
// 						Rectangle  *inkRect,
// 						Rectangle  *logicalRect)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return;

//    GetExtents (iter.layout, inkRect, logicalRect);
//  }
