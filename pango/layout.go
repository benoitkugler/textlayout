package pango

import (
	"container/list"
	"fmt"
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
	tabs     *TabArray

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
	Width Unit

	// Ellipsize height, in device units if positive, number of lines if negative.
	// This is a readonly property, see `SetHeight` to modify it.
	Height Unit

	// Amount by which first line should be shorter.
	// This is a readonly property, see `SetIndent` to modify it.
	Indent Unit

	// Spacing between lines.
	// This is a readonly property, see `SetSpacing` to modify it.
	Spacing Unit

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
	tabWidth             Unit /* Cached width of a tab. -1 == not yet calculated */
	decimal              rune
}

// NewLayout creates a new `Layout` object with attributes initialized to
// default values for a particular `Context`.
func NewLayout(context *Context) *Layout {
	var layout Layout
	layout.context = context
	layout.contextSerial = context.getSerial()

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
// as an accelerator will receive a UNDERLINE_LOW attribute,
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
func (layout *Layout) SetSpacing(spacing Unit) {
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
func (layout *Layout) SetWidth(width Unit) {
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
func (layout *Layout) SetHeight(height Unit) {
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
func (layout *Layout) SetIndent(indent Unit) {
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

// SetTabs sets the tabs to use for `layout`, overriding the default tabs
// (by default, tabs are every 8 spaces). If `tabs` is nil, the default
// tabs are reinstated.
// Note that tabs and justification conflict with each other:
// justification will move content away from its tab-aligned
// positions. The same is true for alignments other than `ALIGN_LEFT`.
func (layout *Layout) SetTabs(tabs *TabArray) {
	if tabs != layout.tabs {
		layout.tabs = tabs.Copy() // handle nil value
		layout.tabs.sort()        // handle nil value
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
func (layout *Layout) indexToLineX(index int, trailing bool) (line int, xPos Unit) {
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

			for {
				run := iter.run

				iter.GetRunExtents(nil, &runRect)

				if run == nil {
					break
				}

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

func (layout *Layout) checkContextChanged() {
	oldSerial := layout.contextSerial

	layout.contextSerial = layout.context.getSerial()

	if oldSerial != layout.contextSerial {
		layout.contextChanged()
	}
}

// Forces recomputation of any state in the `Layout` that
// might depend on the layout's context. This function should
// be called if you make changes to the context subsequent
// to creating the layout.
func (layout *Layout) contextChanged() {
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
	layout.lines = layout.lines[:0]
	layout.logAttrs = layout.logAttrs[:0]
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

func affectsBreakOrShape(attr *Attribute) bool {
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

func affectsItemization(attr *Attribute) bool {
	switch attr.Kind {
	/* These affect font selection */
	case ATTR_LANGUAGE, ATTR_FAMILY, ATTR_STYLE, ATTR_WEIGHT, ATTR_VARIANT, ATTR_STRETCH,
		ATTR_SIZE, ATTR_FONT_DESC, ATTR_SCALE, ATTR_FALLBACK, ATTR_ABSOLUTE_SIZE, ATTR_GRAVITY,
		ATTR_GRAVITY_HINT, ATTR_FONT_SCALE:
		return true
	/* These need to be constant across runs */
	case ATTR_LETTER_SPACING, ATTR_SHAPE, ATTR_RISE, ATTR_LINE_HEIGHT, ATTR_ABSOLUTE_LINE_HEIGHT,
		ATTR_TEXT_TRANSFORM, ATTR_BASELINE_SHIFT:
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
		width = Unit(overallLogical.Width)
	}

	if logicalRect != nil {
		*logicalRect = Rectangle{}
	}

	if withLine { // avoid allocation when not asked
		lines = make([]extents, len(layout.lines))
	}

	var baseline, yOffset Unit
	for lineIndex, line := range layout.lines {
		// Line extents in layout coords (origin at 0,0 of the layout)
		var (
			lineInkLayout, lineLogicalLayout Rectangle
			newPos                           Unit
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

// GetSize determines the logical width and height of a `Layout` in Pango
// units.
//
// This is simply a convenience wrapper around `Layout.GetExtents`.
func (layout *Layout) GetSize() (Unit, Unit) {
	var logicalRect Rectangle
	layout.GetExtents(nil, &logicalRect)
	return logicalRect.Width, logicalRect.Height
}

func (layout *Layout) getEffectiveAttributes() AttrList {
	var attrs AttrList

	if len(layout.Attributes) != 0 {
		attrs = layout.Attributes.copy()
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

func (layout *Layout) getEmptyExtentsAndHeightAt(index int, logicalRect *Rectangle, applyLineHeight bool) (height Unit) {
	if logicalRect == nil {
		return
	}

	fontDesc := layout.context.fontDesc // copy

	if layout.fontDesc != nil {
		fontDesc.pango_font_description_merge(layout.fontDesc, true)
	}

	// Find the font description for this line
	var (
		lineHeightFactor   Fl
		absoluteLineHeight Unit
	)
	if len(layout.Attributes) != 0 {
		iter := layout.Attributes.getIterator()
		for hasNext := true; hasNext; hasNext = iter.next() {
			if iter.StartIndex <= index && index < iter.EndIndex {
				iter.getFont(&fontDesc, nil, nil)

				attr := iter.getByKind(ATTR_LINE_HEIGHT)
				if attr != nil {
					lineHeightFactor = Fl(attr.Data.(AttrFloat))
				}

				attr = iter.getByKind(ATTR_ABSOLUTE_LINE_HEIGHT)
				if attr != nil {
					absoluteLineHeight = Unit(attr.Data.(AttrInt))
				}

				break
			}
		}
	}

	font := layout.context.loadFont(&fontDesc)
	if font != nil {
		metrics := FontGetMetrics(font, layout.context.setLanguage)

		logicalRect.Y = -metrics.Ascent
		logicalRect.Height = -logicalRect.Y + metrics.Descent
		height = metrics.Height

		if applyLineHeight && (absoluteLineHeight != 0 || lineHeightFactor != 0.0) {
			lineHeight := maxG(absoluteLineHeight, Unit(lineHeightFactor*Fl(logicalRect.Height)))
			leading := lineHeight - logicalRect.Height
			logicalRect.Y -= leading / 2
			logicalRect.Height += leading
		}

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

// IndexToPos converts from an index within a `Layout` to the onscreen position
// corresponding to the grapheme at that index.
//
// The return value is represented as rectangle. Note that `pos.X` is
// always the leading edge of the grapheme and `pos.X + pos.Width` the
// trailing edge of the grapheme. If the directionality of the grapheme
// is right-to-left, then `pos.Width` will be negative.
func (layout *Layout) IndexToPos(index int, pos *Rectangle) {
	if pos == nil || index < 0 {
		return
	}

	iter := layout.GetIter()
	var (
		line                            *LayoutLine
		lineLogicalRect, runLogicalRect Rectangle
	)
	for {
		tmpLine := iter.GetLine()

		if tmpLine.StartIndex > index {
			/* index is in the paragraph delimiters, move to
			* end of previous line
			*
			* This shouldn’t occur in the first loop iteration as the first
			* line’s StartIndex should always be 0.
			 */
			if debugMode {
				assert(line != nil, "indexToPos")
			}
			index = line.StartIndex + line.Length
			break
		}

		iter.GetLineExtents(nil, &lineLogicalRect)

		line = tmpLine

		if line.StartIndex+line.Length >= index {
			for do := true; do; do = iter.NextRun() {
				run := iter.GetRun()
				iter.GetRunExtents(nil, &runLogicalRect)

				if run == nil {
					break
				}

				if run.Item.Offset <= index && index < run.Item.Offset+run.Item.Length {
					break
				}
			}

			if line.StartIndex+line.Length > index {
				break
			}
		}

		if !iter.NextLine() {
			index = line.StartIndex + line.Length
			break
		}
	}

	pos.Y = runLogicalRect.Y
	pos.Height = runLogicalRect.Height

	xPos := line.IndexToX(index, false)
	pos.X = lineLogicalRect.X + xPos

	if index < line.StartIndex+line.Length {
		xPos = line.IndexToX(index, true)
		pos.Width = (lineLogicalRect.X + xPos) - pos.X
	} else {
		pos.Width = 0
	}
}

// GetBaseline gets the Y position of baseline of the first line in `layout`, from top of the layout.
func (layout *Layout) GetBaseline() Unit {
	lines := layout.getExtentsInternal(nil, nil, true)
	if len(lines) >= 1 {
		return lines[0].baseline
	}
	return 0
}

// setup the cached value `tabWidth` if not already defined
func (layout *Layout) ensureTabWidth() {
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
	if layout.context.RoundGlyphPositions {
		shape_flags |= shapeROUND_POSITIONS
	}

	layout_attrs := layout.getEffectiveAttributes()
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

// resolve the tab at index `index`
func (line *LayoutLine) get_tab_pos(index int) (tab Tab, isDefault bool) {
	var (
		nTabs    int
		inPixels bool
		layout   = line.layout
		offset   Unit
	)

	if layout.alignment != ALIGN_CENTER {
		if line.IsParagraphStart && layout.Indent >= 0 {
			offset = layout.Indent
		} else if !line.IsParagraphStart && layout.Indent < 0 {
			offset = -layout.Indent
		}
	}

	isDefault = true
	if layout.tabs != nil {
		nTabs = len(layout.tabs.Tabs)
		inPixels = layout.tabs.PositionsInPixels
		isDefault = false
	}

	if index < nTabs {
		tab = layout.tabs.Tabs[index]
		if inPixels {
			tab.Location *= Scale
		}
	} else if nTabs > 0 {
		// Extrapolate tab position, repeating the last tab gap to infinity.

		tab = layout.tabs.Tabs[nTabs-1]
		lastPos := tab.Location

		var nextToLastPos Unit
		if nTabs > 1 {
			nextToLastPos = layout.tabs.Tabs[nTabs-2].Location
		}

		if inPixels {
			nextToLastPos *= Scale
			lastPos *= Scale
		}

		var tabWidth Unit
		if lastPos > nextToLastPos {
			tabWidth = lastPos - nextToLastPos
		} else {
			tabWidth = layout.tabWidth
		}

		tab.Location = lastPos + tabWidth*Unit(index-nTabs+1)
	} else {
		// No tab array set, so use default tab width
		tab = Tab{
			Location:  layout.tabWidth * Unit(index),
			Alignment: TAB_LEFT,
		}
	}

	tab.Location -= offset

	return tab, isDefault
}

func (layout *Layout) ensureDecimal() {
	if layout.decimal == 0 {
		layout.decimal = '.'
	}
}

func (layout *Layout) canBreakAt(offset int, wrap WrapMode) bool {
	if offset == len(layout.Text) {
		return true
	} else if wrap == WRAP_CHAR {
		return layout.logAttrs[offset].IsCharBreak()
	} else {
		return layout.logAttrs[offset].IsLineBreak()
	}
}

func (layout *Layout) canBreakIn(startOffset, Length int, allowBreakAtStart bool) bool {
	i := 1
	if allowBreakAtStart {
		i = 0
	}

	for ; i < Length; i++ {
		if layout.canBreakAt(startOffset+i, layout.wrap) {
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
// 	   *space_left = UNITS_ROUND (*space_left);
// 	 }
//    *space_right = letter_spacing - *space_left;
//  }

type paraBreakState struct {
	/* maintained per layout */
	lineHeight      Unit /* Estimate of height of current line; < 0 is no estimate */
	remainingHeight Unit /* Remaining height of the layout;  only defined if layout.height >= 0 */

	/* maintained per paragraph */
	attrs       AttrList  /* Attributes being used for itemization */
	items       *ItemList /* This paragraph turned into items */
	base_dir    Direction /* Current resolved base direction */
	line_of_par int       /* Line of the paragraph, starting at 1 for first line */

	glyphs          *GlyphString   /* Glyphs for the first item in state.items */
	startOffset     int            /* Character offset of first item in state.items in layout.text */
	properties      itemProperties /* Properties for the first item in state.items */
	logWidths       []Unit         /* Logical widths for first item in state.items.. */
	logWidthsOffset int            /* Offset into logWidths to the point corresponding
	 * to the remaining portion of the first item */

	lineStartIndex int /* Start index of line in layout.text */

	/* maintained per line */
	lineWidth      Unit /* Goal width of line currently processing; < 0 is infinite */
	remainingWidth Unit /* Amount of space remaining on line; < 0 is infinite */

	hyphen_width Unit /* How much space a hyphen will take */

	baselineShifts list.List

	lastTab lastTabState
}

func (item *Item) get_decimal_prefix_width(glyphs *GlyphString, text []rune, decimal rune) (width Unit, found bool) {
	glyphItem := GlyphItem{Item: item, Glyphs: glyphs}

	logWidths := make([]Unit, item.Length)
	glyphItem.getLogicalWidths(text, logWidths)

	for i, c := range text[item.Offset : item.Offset+item.Length] {
		if c == decimal {
			width += logWidths[i] / 2
			found = true
			break
		}

		width += logWidths[i]
	}

	return width, found
}

func (layout *Layout) breakNeedsHyphen(state *paraBreakState, pos int) bool {
	c := layout.logAttrs[state.startOffset+pos]
	return c.IsBreakInsertsHyphen() || c.IsBreakRemovesPreceding()
}

// resizeLogicalWidths resize logWidths to the given length, reusing
// the current storage if possible
func (state *paraBreakState) resizeLogicalWidths(newLen int) {
	if newLen <= cap(state.logWidths) {
		state.logWidths = state.logWidths[:newLen]
	} else { // re-allocate
		state.logWidths = make([]Unit, newLen)
	}
}

func (state *paraBreakState) ensure_hyphen_width() {
	if state.hyphen_width < 0 {
		item := state.items.Data
		state.hyphen_width = item.find_hyphen_width()
	}
}

func (layout *Layout) find_break_extraWidth(state *paraBreakState, pos int) Unit {
	// Check whether to insert a hyphen
	// or whether we are breaking after one of those
	// characters that turn into a hyphen,
	// or after a space.
	if attr := layout.logAttrs[state.startOffset+pos]; attr.IsBreakInsertsHyphen() {
		state.ensure_hyphen_width()

		if attr.IsBreakRemovesPreceding() && pos > 0 {
			return state.hyphen_width - state.logWidths[state.logWidthsOffset+pos-1]
		} else {
			return state.hyphen_width
		}
	} else if pos > 0 &&
		layout.logAttrs[state.startOffset+pos-1].IsWhite() {
		return -state.logWidths[state.logWidthsOffset+pos-1]
	}
	return 0
}

func (layout *Layout) computeLogWidths(state *paraBreakState) {
	item := state.items.Data
	glyphItem := GlyphItem{Item: item, Glyphs: state.glyphs}

	if item.Length > cap(state.logWidths) {
		state.logWidths = make([]Unit, item.Length)
	} else {
		state.logWidths = state.logWidths[:item.Length]
	}

	glyphItem.getLogicalWidths(layout.Text, state.logWidths)
}

// If lastTab is set, we've added a tab and remainingWidth has been updated to
// account for its origin width, which is lastTab_pos - lastTab_width. shape_run
// updates the tab width, so we need to consider the delta when comparing
// against remainingWidth.
func (state *paraBreakState) tabWidthChange() Unit {
	if state.lastTab.glyphs != nil {
		return state.lastTab.glyphs.Glyphs[0].Geometry.Width - (state.lastTab.tab.Location - state.lastTab.width)
	}

	return 0
}

type breakResult uint8

const (
	brNONE_FIT       breakResult = iota // Couldn't fit anything.
	brSOME_FIT                          // The item was broken in the middle.
	brALL_FIT                           // Everything fit.
	brEMPTY_FIT                         // Nothing fit, but that was ok, as we can break at the first char.
	brLINE_SEPARATOR                    // Item begins with a line separator.
)

// tryAddItemToLine tries to insert as much as possible of the first item of
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
//
// This function is the core of our line-breaking, and it is long and involved.
// Here is an outline of the algorithm, without all the bookkeeping:
//
// if item appears to fit entirely
//   measure it
//   if it actually fits
//     return BREAK_ALL_FIT
//
// retry_break:
//   for each position p in the item
//     if adding more is 'obviously' not going to help and we have a breakpoint
//       exit the loop
//     if p is a possible break position
//       if p is 'obviously' going to fit
//         bc = p
//       else
//         measure breaking at p (taking extra break width into account
//         if we don't have a break candidate yet)
//           bc = p
//         else
//           if p is better than bc
//             bc = p
//
//   if bc does not fit and we can loosen break conditions
//     loosen break conditions and retry break
//
// return bc
func (layout *Layout) tryAddItemToLine(line *LayoutLine, state *paraBreakState,
	forceFit, noBreakAtEnd, isLastItem bool) breakResult {
	item := state.items.Data
	shapeSet := false
	processingNewItem := false

	if debugMode {
		fmt.Printf("tryAddItemToLine '%s'. Remaining width %d\n",
			string(layout.Text[item.Offset:item.Offset+item.Length]), state.remainingWidth)
	}

	/* We don't want to shape more than necessary, so we keep the results
	 * of shaping a new item in state.glyphs, state.logWidths. Once
	 * we break off initial parts of the item, we update state.logWidths_offset
	 * to take that into account. Note that the widths we calculate from the
	 * logWidths are an approximation, because a) logWidths are just
	 * evenly divided for clusters, and b) clusters may change as we
	 * break in the middle (think ff- i).
	 *
	 * We use state.logWidths_offset != 0 to detect if we are dealing
	 * with the original item, or one that has been chopped off.
	 */
	if state.glyphs == nil {
		state.properties = item.getProperties()
		state.glyphs = line.shapeRun(state, item)

		state.logWidthsOffset = 0

		processingNewItem = true
	}

	if !layout.singleParagraph && layout.Text[item.Offset] == lineSeparator &&
		!layout.shouldEllipsizeCurrentLine(state) {
		line.insertRun(state, item, nil, true)
		state.logWidthsOffset += item.Length

		return brLINE_SEPARATOR
	}

	if state.remainingWidth < 0 && !noBreakAtEnd /* Wrapping off */ {
		line.insertRun(state, item, nil, true)

		if debugMode {
			fmt.Println("no wrapping, all-fit")
		}
		return brALL_FIT
	}

	var width Unit
	if processingNewItem {
		width = state.glyphs.getWidth()
	} else {
		for _, w := range state.logWidths[state.logWidthsOffset : state.logWidthsOffset+item.Length] {
			width += w
		}
	}

	if layout.Text[item.Offset] == '\t' {
		line.insertRun(state, item, nil, true)
		state.remainingWidth -= width
		state.remainingWidth = maxG(state.remainingWidth, 0)

		if debugMode {
			fmt.Println("tab run, all-fit")
		}

		return brALL_FIT
	}

	var extraWidth Unit
	if !noBreakAtEnd &&
		layout.canBreakAt(state.startOffset+item.Length, WRAP_WORD) {
		if processingNewItem {
			layout.computeLogWidths(state)
			processingNewItem = false
		}
		extraWidth = layout.find_break_extraWidth(state, item.Length)
	}

	if (width+extraWidth <= state.remainingWidth || (item.Length == 1 && line.Runs == nil) ||
		(state.lastTab.glyphs != nil && state.lastTab.tab.Alignment != TAB_LEFT)) &&
		!noBreakAtEnd {

		if debugMode {
			fmt.Printf("%d + %d <= %d\n", width, extraWidth, state.remainingWidth)
		}
		glyphs := line.shapeRun(state, item)

		width = glyphs.getWidth() + state.tabWidthChange()

		if width+extraWidth <= state.remainingWidth || (item.Length == 1 && line.Runs == nil) {
			line.insertRun(state, item, glyphs, true)

			state.remainingWidth -= width
			state.remainingWidth = maxG(state.remainingWidth, 0)

			if debugMode {
				fmt.Printf("early accept '%.*s', all-fit, remaining %d\n",
					item.Length, string(layout.Text[item.Offset:]), state.remainingWidth)
			}
			return brALL_FIT
		}

		/* if it doesn't fit after shaping, discard and proceed to break the item */
	}

	// From here on, we look for a way to break item

	width = 0
	var (
		orig_width      = width
		origExtraWidth  = extraWidth
		breakWidth      = width
		breakExtraWidth = extraWidth
		breakNumChars   = item.Length
		wrap            = layout.wrap

		// retrying_with_char_breaks = false
		breakGlyphs *GlyphString
		numChars    int
	)

	// Add some safety margin here. If we are farther away from the end of the
	// line than this, we don't look carefully at a break possibility.
	metrics := item.Analysis.Font.GetMetrics(item.Analysis.Language)
	safe_distance := metrics.ApproximateCharWidth * 3

	if processingNewItem {
		layout.computeLogWidths(state)
		processingNewItem = false
	}

retryBreak:
	limit := item.Length + 1
	if noBreakAtEnd {
		limit--
	}
	for numChars = 0; numChars < limit; numChars++ {
		extraWidth = layout.find_break_extraWidth(state, numChars)

		// We don't want to walk the entire item if we can help it, but
		// we need to keep going at least until we've found a breakpoint
		// that 'works' (as in, it doesn't overflow the budget we have,
		// or there is no hope of finding a better one).
		//
		// We rely on the fact that MIN(width + extraWidth, width) is
		// monotonically increasing.

		if minG(width+extraWidth, width) > state.remainingWidth+safe_distance && breakNumChars < item.Length {

			if debugMode {
				fmt.Printf("at %d, MIN(%d, %d + %d) > %d + MARGIN, breaking at %d\n",
					numChars, width, extraWidth, width, state.remainingWidth, breakNumChars)
			}

			break
		}

		// If there are no previous runs we have to take care to grab at least one char.
		if layout.canBreakAt(state.startOffset+numChars, wrap) &&
			(numChars > 0 || line.Runs != nil) {

			if debugMode {
				fmt.Printf("possible breakpoint: %d, extraWidth %d\n", numChars, extraWidth)
			}

			if numChars == 0 || width+extraWidth < state.remainingWidth-safe_distance {
				if debugMode {
					fmt.Println("trivial accept")
				}
				breakNumChars = numChars
				breakWidth = width
				breakExtraWidth = extraWidth
			} else {
				newItem := item
				if numChars < item.Length {
					newItem = item.split(numChars)

					if layout.breakNeedsHyphen(state, numChars) {
						newItem.Analysis.Flags |= AFNeedHyphen
					} else {
						newItem.Analysis.Flags &= ^AFNeedHyphen
					}
				}

				glyphs := line.shapeRun(state, newItem)

				newBreakWidth := glyphs.getWidth() + state.tabWidthChange()

				if numChars > 0 &&
					(item != newItem || !isLastItem) && // We don't collapse space at the very end
					layout.logAttrs[state.startOffset+numChars-1].IsWhite() {
					extraWidth = -state.logWidths[state.logWidthsOffset+numChars-1]
				} else if item == newItem && !isLastItem && layout.breakNeedsHyphen(state, numChars) {
					extraWidth = state.hyphen_width
				} else {
					extraWidth = 0
				}

				if debugMode {
					fmt.Printf("measured breakpoint %d: %d, extra %d\n", numChars, newBreakWidth, extraWidth)
				}

				if newItem != item {
					item.unSplit(numChars)
				}

				if breakNumChars == item.Length ||
					newBreakWidth+extraWidth <= state.remainingWidth ||
					newBreakWidth+extraWidth < breakWidth+breakExtraWidth {

					if debugMode {
						fmt.Printf("accept breakpoint %d: %d + %d <= %d + %d\n",
							numChars, newBreakWidth, extraWidth, breakWidth, breakExtraWidth)
						fmt.Printf("replace bp %d by %d\n", breakNumChars, numChars)
					}
					breakNumChars = numChars
					breakWidth = newBreakWidth
					breakExtraWidth = extraWidth

					breakGlyphs = glyphs
				} else {

					if debugMode {
						fmt.Printf("ignore breakpoint %d\n", numChars)
					}
				}
			}
		}

		if debugMode {
			fmt.Printf("bp now %d\n", breakNumChars)
		}

		if numChars < item.Length {
			width += state.logWidths[state.logWidthsOffset+numChars]
		}
	}

	if wrap == WRAP_WORD_CHAR && forceFit && breakWidth+breakExtraWidth > state.remainingWidth {
		// Try again, with looser conditions
		if debugMode {
			fmt.Printf("does not fit, try again with wrap-char\n")
		}
		wrap = WRAP_CHAR
		breakNumChars = item.Length
		breakWidth = orig_width
		breakExtraWidth = origExtraWidth
		breakGlyphs = nil

		goto retryBreak
	}

	if forceFit || breakWidth+breakExtraWidth <= state.remainingWidth /* Successfully broke the item */ {
		if state.remainingWidth >= 0 {
			state.remainingWidth -= breakWidth + breakExtraWidth
			state.remainingWidth = maxG(state.remainingWidth, 0)
		}

		if breakNumChars == item.Length {
			if layout.canBreakAt(state.startOffset+breakNumChars, wrap) && layout.breakNeedsHyphen(state, breakNumChars) {
				item.Analysis.Flags |= AFNeedHyphen
			}
			line.insertRun(state, item, nil, true)

			if debugMode {
				fmt.Printf("all-fit '%.*s', remaining %d\n",
					item.Length, string(layout.Text[item.Offset:]), state.remainingWidth)
			}
			return brALL_FIT
		} else if breakNumChars == 0 {

			if debugMode {
				fmt.Printf("empty-fit, remaining %d\n", state.remainingWidth)
			}

			return brEMPTY_FIT
		} else {
			newItem := item.split(breakNumChars)

			if layout.breakNeedsHyphen(state, breakNumChars) {
				newItem.Analysis.Flags |= AFNeedHyphen
			}

			line.insertRun(state, newItem, breakGlyphs, false)

			state.logWidthsOffset += breakNumChars

			// shaped items should never be broken
			if debugMode {
				assert(!shapeSet, "processItem: break")

				fmt.Printf("some-fit '%.*s', remaining %d\n",
					item.Length, string(layout.Text[item.Offset:]), state.remainingWidth)
			}

			return brSOME_FIT
		}
	} else {
		state.glyphs = nil

		if debugMode {
			fmt.Printf("none-fit, remaining %d\n", state.remainingWidth)
		}

		return brNONE_FIT
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
		return state.lineHeight*2 > state.remainingHeight
	} else {
		/* -layout.height is number of lines per paragraph to show */
		return state.line_of_par == int(-layout.Height)
	}
}

// the hard work begins here !
func (layout *Layout) processLine(state *paraBreakState) {
	var (
		haveBreak           = false  /* If we've seen a possible break yet */
		breakRemainingWidth Unit     /* Remaining width before adding run with break */
		breakStartOffset    = 0      /* Start offset before adding run with break */
		breakLink           *RunList /* Link holding run before break */
		wrapped             = false  /* If we had to wrap the line */
	)

	line := layout.newLine()
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
		state.remainingWidth = -1
	} else {
		state.remainingWidth = state.lineWidth
	}

	state.lastTab.glyphs = nil
	state.lastTab.index = 0
	state.lastTab.tab.Alignment = TAB_LEFT

	if debugMode {
		showDebug("starting to fill line\n", line, state)
	}

	for state.items != nil {
		item := state.items.Data

		oldNumChars := item.Length
		oldRemainingWidth := state.remainingWidth
		firstItemInLine := line.Runs == nil
		lastItemInLine := state.items.Next == nil

		result := layout.tryAddItemToLine(line, state, !haveBreak, false, lastItemInLine)

		switch result {
		case brALL_FIT:
			if layout.Text[item.Offset] != '\t' &&
				layout.canBreakIn(state.startOffset, oldNumChars, !firstItemInLine) {
				haveBreak = true
				breakRemainingWidth = oldRemainingWidth
				breakStartOffset = state.startOffset
				breakLink = line.Runs.Next

				if debugMode {
					fmt.Println("all-fit, have break")
				}
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
			// Back up over unused runs to run where there is a break
			for line.Runs != nil && line.Runs != breakLink {
				run := line.Runs.Data

				// If we uninsert the current tab run,  we need to reset the tab state
				if run.Glyphs == state.lastTab.glyphs {
					state.lastTab.glyphs = nil
					state.lastTab.index = 0
					state.lastTab.tab.Alignment = TAB_LEFT
				}

				state.items = &ItemList{Data: line.uninsert_run(), Next: state.items}
			}

			state.startOffset = breakStartOffset
			state.remainingWidth = breakRemainingWidth
			lastItemInLine = state.items.Next == nil

			/* Reshape run to break */
			item = state.items.Data

			oldNumChars = item.Length
			result = layout.tryAddItemToLine(line, state, true, true, lastItemInLine)
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

	if debugMode {
		fmt.Printf("line %d done. remaining %d\n", state.line_of_par, state.remainingWidth)
	}

	line.addLine(state)
	state.line_of_par++
	state.lineStartIndex += line.Length
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

func (layout *Layout) applyAttributesToRuns(attrs AttrList) {
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

	layout.checkContextChanged()

	if len(layout.lines) != 0 {
		return
	}

	attrs := layout.getEffectiveAttributes()
	if attrs != nil {
		shapeAttrs = attrs.Filter(affectsBreakOrShape)
		itemizeAttrs = attrs.Filter(affectsItemization)
		if itemizeAttrs != nil {
			iter = *itemizeAttrs.getIterator()
		}
	}

	if len(layout.logAttrs) == 0 {
		L := len(layout.Text) + 1
		if cap(layout.logAttrs) >= L {
			layout.logAttrs = layout.logAttrs[:L]
		} else {
			layout.logAttrs = make([]CharAttr, L)
		}
		needLogAttrs = true
	} else {
		needLogAttrs = false
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
	state.remainingHeight = layout.Height
	state.lineHeight = -1
	if layout.Height >= 0 {
		var logical Rectangle
		height := layout.getEmptyExtentsAndHeightAt(0, &logical, true)
		if layout.LineSpacing == 0 {
			state.lineHeight = logical.Height
		} else {
			state.lineHeight = Unit(layout.LineSpacing * float32(height))
		}
	}

	state.logWidths = state.logWidths[:0]
	state.baselineShifts.Init()

	if debugMode {
		fmt.Printf("START layout for %s\n", string(layout.Text))
	}

	for done := false; !done; {
		var delimiterIndex, nextParaIndex int

		if layout.singleParagraph {
			delimiterIndex = len(layout.Text)
			nextParaIndex = len(layout.Text)
		} else {
			delimiterIndex, nextParaIndex = findParagraphBoundary(layout.Text[start:])
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

		state.attrs = itemizeAttrs

		if debugMode {
			fmt.Println("Itemizing...")
		}

		var cachedIter *attrIterator
		if itemizeAttrs != nil {
			cachedIter = &iter
		}
		state.items = layout.context.itemizeWithFont(
			layout.Text[:end],
			start,
			nil,
			baseDir,
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

		state.items = layout.context.postProcessItems(layout.Text, layout.logAttrs, state.items)

		state.base_dir = baseDir
		state.line_of_par = 1
		state.startOffset = start
		state.lineStartIndex = start

		state.glyphs = nil

		// for deterministic bug hunting's sake set everything!
		state.lineWidth = -1
		state.remainingWidth = -1
		state.logWidthsOffset = 0

		state.hyphen_width = -1

		if state.items != nil {
			for state.items != nil {
				if debugMode {
					fmt.Println("Processing line...")
				}
				layout.processLine(&state)
			}
		} else {
			emptyLine := layout.newLine()
			emptyLine.StartIndex = state.lineStartIndex
			emptyLine.IsParagraphStart = true
			emptyLine.setResolvedDir(baseDir)
			emptyLine.addLine(&state)
		}

		if layout.Height >= 0 && state.remainingHeight < state.lineHeight {
			done = true
		}

		start = end + delimLen
	}
	layout.applyAttributesToRuns(attrs)
}

func (layout *Layout) isTabRun(run *GlyphItem) bool {
	return layout.Text[run.Item.Offset] == '\t'
}
