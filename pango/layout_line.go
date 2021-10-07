package pango

import (
	"container/list"
	"log"
	"math"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fribidi"
)

type extents struct {
	// Vertical position of the line's baseline in layout coords
	baseline GlyphUnit

	// Line extents in layout coords
	inkRect, logicalRect Rectangle
}

// extents cache status:
// leaked means that the user has access to this line structure or a
// run included in this line, and so can change the glyphs/glyph-widths.
// If this is true, extents caching will be disabled.
const (
	notCached uint8 = iota
	cached
	leaked
)

// GetLine retrieves a particular line from a `Layout`,
// or `nil` if the index is out of range.
// This layout line will become invalid if changes are made to the
// `Layout`.
//
// The returned line should not be modified.
func (layout *Layout) GetLine(line int) *LayoutLine {
	if line < 0 {
		return nil
	}

	layout.checkLines()

	if line >= len(layout.lines) {
		return nil
	}

	return layout.lines[line]
}

// LayoutLine represents one of the lines resulting
// from laying out a paragraph via `Layout`. `LayoutLine`
// structures are only valid until the text, attributes, or settings of the
// parent `Layout` are modified.
type LayoutLine struct {
	layout           *Layout   // the layout this line belongs to, might be nil
	Runs             *RunList  // list of runs in the line, from left to right
	StartIndex       int       // start of line as rune index into layout.Text
	Length           int       // length of line in runes
	IsParagraphStart bool      // true if this is the first line of the paragraph
	ResolvedDir      Direction // Resolved Direction of line

	cache_status uint8
	inkRect      Rectangle
	logicalRect  Rectangle
	height       GlyphUnit
}

func (layout *Layout) newLine() *LayoutLine {
	var private LayoutLine
	private.layout = layout
	return &private
}

func (line *LayoutLine) leaked() {
	line.cache_status = leaked

	if line.layout != nil {
		line.layout.logicalRectCached = false
		line.layout.inkRectCached = false
	}
}

// RunList is a linked list of `GlypmItem`
type RunList struct {
	Data *GlyphItem
	Next *RunList
}

func (l *RunList) reverse() *RunList {
	var out *RunList

	for ; l != nil; l = l.Next {
		out = &RunList{Data: l.Data, Next: out}
	}

	return out
}

func (l *RunList) length() int {
	out := 0
	for ; l != nil; l = l.Next {
		out++
	}
	return out
}

func (l *RunList) concat(snd *RunList) *RunList {
	if l == nil {
		return snd
	}
	head := l
	for ; l.Next != nil; l = l.Next {
	}
	l.Next = snd
	return head
}

func (items *RunList) reorderRunsRecurse(nItems int) *RunList {
	//    GSList *tmpList, *levelStartNode;
	//    int i, levelStartI;
	//    GSList *result = nil;

	if nItems == 0 {
		return nil
	}

	tmpList := items
	var minLevel fribidi.Level = math.MaxInt8
	for i := 0; i < nItems; i++ {
		run := tmpList.Data
		minLevel = minL(minLevel, run.Item.Analysis.Level)

		tmpList = tmpList.Next
	}

	var (
		result *RunList
		i      int
	)
	levelStartI := 0
	levelStartNode := items
	tmpList = items
	for i = 0; i < nItems; i++ {
		run := tmpList.Data

		if run.Item.Analysis.Level == minLevel {
			if minLevel%2 != 0 {
				if i > levelStartI {
					result = levelStartNode.reorderRunsRecurse(i - levelStartI).concat(result)
				}
				result = &RunList{Data: run, Next: result}
			} else {
				if i > levelStartI {
					result = result.concat(levelStartNode.reorderRunsRecurse(i - levelStartI))
				}
				result = result.concat(&RunList{Data: run})
			}

			levelStartI = i + 1
			levelStartNode = tmpList.Next
		}

		tmpList = tmpList.Next
	}

	if minLevel%2 != 0 {
		if i > levelStartI {
			result = levelStartNode.reorderRunsRecurse(i - levelStartI).concat(result)
		}
	} else {
		if i > levelStartI {
			result = result.concat(levelStartNode.reorderRunsRecurse(i - levelStartI))
		}
	}

	return result
}

// The resolved direction for the line is always one
// of LTR/RTL; not a week or neutral directions
func (line *LayoutLine) setResolvedDir(direction Direction) {
	switch direction {
	case DIRECTION_RTL, DIRECTION_WEAK_RTL:
		line.ResolvedDir = DIRECTION_RTL
	default:
		line.ResolvedDir = DIRECTION_LTR
	}

	// The direction vs. gravity dance:
	//	- If gravity is SOUTH, leave direction untouched.
	//	- If gravity is NORTH, switch direction.
	//	- If gravity is EAST, set to LTR, as
	//	  it's a clockwise-rotated layout, so the rotated
	//	  top is unrotated left.
	//	- If gravity is WEST, set to RTL, as
	//	  it's a counter-clockwise-rotated layout, so the rotated
	//	  top is unrotated right.
	//
	// A similar dance is performed in pango-context.c:
	// itemize_state_add_character().  Keep in synch.
	switch line.layout.context.resolvedGravity {
	case GRAVITY_NORTH:
		line.ResolvedDir = DIRECTION_LTR + DIRECTION_RTL - line.ResolvedDir
	case GRAVITY_EAST:
		// This is in fact why deprecated TTB_RTL is LTR
		line.ResolvedDir = DIRECTION_LTR
	case GRAVITY_WEST:
		// This is in fact why deprecated TTB_LTR is RTL
		line.ResolvedDir = DIRECTION_RTL
	}
}

func (line *LayoutLine) shape_run(state *paraBreakState, item *Item) *GlyphString {
	layout := line.layout
	glyphs := &GlyphString{}

	if layout.Text[item.Offset] == '\t' {
		line.shape_tab(item, glyphs)
	} else {
		shapeFlag := shapeNONE

		if layout.context.RoundGlyphPositions {
			shapeFlag |= shapeROUND_POSITIONS
		}
		if state.properties.shape != nil {
			glyphs._pango_shape_shape(layout.Text[item.Offset:item.Offset+item.Length], state.properties.shape.logical)
		} else {
			item.Shape(layout.Text, layout.logAttrs[state.startOffset:], glyphs, shapeFlag)
		}

		if state.properties.letterSpacing != 0 {
			glyphItem := GlyphItem{Item: item, Glyphs: glyphs}

			glyphItem.letterSpace(layout.Text,
				layout.logAttrs[state.startOffset:],
				state.properties.letterSpacing)

			spaceLeft, spaceRight := distributeLetterSpacing(state.properties.letterSpacing)

			glyphs.Glyphs[0].Geometry.Width += spaceLeft
			glyphs.Glyphs[0].Geometry.XOffset += spaceLeft
			glyphs.Glyphs[len(glyphs.Glyphs)-1].Geometry.Width += spaceRight
		}
	}

	return glyphs
}

func distributeLetterSpacing(letterSpacing GlyphUnit) (spaceLeft, spaceRight GlyphUnit) {
	spaceLeft = letterSpacing / 2
	// hinting
	if (letterSpacing & (Scale - 1)) == 0 {
		spaceLeft = spaceLeft.Round()
	}
	spaceRight = letterSpacing - spaceLeft
	return
}

func (line *LayoutLine) shape_tab(item *Item, glyphs *GlyphString) {
	current_width := line.lineWidth()

	glyphs.setSize(1)

	if item.Analysis.showing_space() {
		glyphs.Glyphs[0].Glyph = AsUnknownGlyph('\t')
	} else {
		glyphs.Glyphs[0].Glyph = GLYPH_EMPTY
	}
	glyphs.Glyphs[0].Geometry.XOffset = 0
	glyphs.Glyphs[0].Geometry.YOffset = 0
	glyphs.Glyphs[0].attr.isClusterStart = true

	glyphs.LogClusters[0] = 0

	line.layout.ensure_tab_width()
	space_width := line.layout.tabWidth / 8

	for i := 0; ; i++ {
		tab_pos, is_default := line.layout.get_tab_pos(i)
		// Make sure there is at least a space-width of space between
		// tab-aligned text and the text before it.  However, only do
		// this if no tab array is set on the layout, ie. using default
		// tab positions. If user has set tab positions, respect it to
		// the pixel.
		var sw GlyphUnit = 1
		if is_default {
			sw = space_width
		}
		if GlyphUnit(tab_pos) >= current_width+sw {
			glyphs.Glyphs[0].Geometry.Width = GlyphUnit(tab_pos) - current_width
			break
		}
	}
}

func (line *LayoutLine) getWidth() GlyphUnit {
	var width GlyphUnit
	for l := line.Runs; l != nil; l = l.Next {
		width += l.Data.Glyphs.getWidth()
	}
	return width
}

func (line *LayoutLine) lineWidth() GlyphUnit {
	var width GlyphUnit

	// Compute the width of the line currently - inefficient, but easier
	// than keeping the current width of the line up to date everywhere
	for l := line.Runs; l != nil; l = l.Next {
		for _, info := range l.Data.Glyphs.Glyphs {
			width += info.Geometry.Width
		}
	}

	return width
}

func (line *LayoutLine) insertRun(state *paraBreakState, runItem *Item, lastRun bool) {
	run := GlyphItem{Item: runItem}

	if lastRun && state.log_widths_offset == 0 && runItem.Analysis.Flags&AFNeedHyphen == 0 {
		run.Glyphs = state.glyphs
	} else {
		run.Glyphs = line.shape_run(state, runItem)
	}

	if lastRun {
		state.glyphs = nil
	}

	line.Runs = &RunList{Data: &run, Next: line.Runs} // prepend
	line.Length += runItem.Length
}

func (line *LayoutLine) uninsert_run() *Item {
	runItem := line.Runs.Data.Item

	line.Runs = line.Runs.Next
	line.Length -= runItem.Length

	return runItem
}

func (line *LayoutLine) postprocess(state *paraBreakState, wrapped bool) {
	ellipsized := false

	if debugMode {
		showDebug("postprocessing", line, state)
	}

	// Truncate the logical-final whitespace in the line if we broke the line at it
	if wrapped {
		// The runs are in reverse order at this point, since we prepended them to the list.
		// So, the first run is the last logical run.
		line.zero_line_final_space(state, line.Runs.Data)
	}

	// Reverse the runs
	line.Runs = line.Runs.reverse()

	line.applyBaselineShift(state)

	// Ellipsize the line if necessary
	if state.lineWidth >= 0 && line.layout.shouldEllipsizeCurrentLine(state) {
		shapeFlags := shapeNONE

		if line.layout.context.RoundGlyphPositions {
			shapeFlags |= shapeROUND_POSITIONS
		}

		ellipsized = line.ellipsize(state.attrs, shapeFlags, state.lineWidth)
	}

	if debugMode {
		showDebug("after removing final space", line, state)
	}

	// Now convert logical to visual order
	line.pangoLayoutLineReorder()

	if debugMode {
		showDebug("after reordering", line, state)
	}

	// Fixup letter spacing between runs
	line.adjust_line_letter_spacing(state)

	if debugMode {
		showDebug("after letter spacing", line, state)
	}

	// Distribute extra space between words if justifying and line was wrapped
	if line.layout.Justify && (wrapped || ellipsized || line.layout.JustifyLastLine) {
		/* if we ellipsized, we don't have remaining_width set */
		if state.remaining_width < 0 {
			state.remaining_width = state.lineWidth - line.getWidth()
		}

		line.justifyWords(state)
	}

	if debugMode {
		showDebug("after justification", line, state)
	}

	line.layout.isWrapped = line.layout.isWrapped || wrapped
	line.layout.isEllipsized = line.layout.isEllipsized || ellipsized
}

func (line *LayoutLine) zero_line_final_space(state *paraBreakState, run *GlyphItem) {
	layout := line.layout

	lineChars := 0
	for l := line.Runs; l != nil; l = l.Next {
		if r := l.Data; r != nil {
			lineChars += r.Item.Length
		}
	}

	item := run.Item
	if layout.logAttrs[state.lineStartIndex+lineChars].IsBreakInsertsHyphen() &&
		(item.Analysis.Flags&AFNeedHyphen == 0) {

		// The last run fit onto the line without breaking it, but it still needs a hyphen
		width := run.Glyphs.getWidth()

		/* Ugly, shape_run uses state.startOffset, so temporarily rewind things
		 * to the state before the run was inserted. Otherwise, we end up passing
		 * the wrong log attrs to the shaping machinery.
		 */
		startOffset := state.startOffset
		state.startOffset = state.lineStartIndex + lineChars - item.Length

		item.Analysis.Flags |= AFNeedHyphen
		run.Glyphs = line.shape_run(state, item)

		state.startOffset = startOffset

		state.remaining_width += run.Glyphs.getWidth() - width
	}

	glyphs := run.Glyphs
	glyph := 0
	if run.LTR() {
		glyph = len(glyphs.Glyphs) - 1
	}

	if glyphs.Glyphs[glyph].Glyph == AsUnknownGlyph(0x2028) {
		return // this LS is visible
	}

	// if the final char of line forms a cluster, and it's
	// a whitespace char, zero its glyph's width as it's been wrapped
	if len(glyphs.Glyphs) < 1 || state.startOffset == 0 ||
		!layout.logAttrs[state.startOffset-1].IsWhite() {
		return
	}

	offset := 1
	if run.LTR() {
		offset = -1
	}
	if len(glyphs.Glyphs) >= 2 && glyphs.LogClusters[glyph] == glyphs.LogClusters[glyph+offset] {
		return
	}

	state.remaining_width += glyphs.Glyphs[glyph].Geometry.Width
	glyphs.Glyphs[glyph].Geometry.Width = 0
	glyphs.Glyphs[glyph].Glyph = GLYPH_EMPTY
}

func (line *LayoutLine) pangoLayoutLineReorder() {
	logicalRuns := line.Runs
	//    GSList *tmpList;
	//    bool all_even, all_odd;
	var (
		levelOr  fribidi.Level
		levelAnd fribidi.Level = 1
		length   int
	)

	/* Check if all items are in the same direction, in that case, the
	* line does not need modification and we can avoid the expensive
	* reorder runs recurse procedure.
	 */
	for tmpList := logicalRuns; tmpList != nil; tmpList = tmpList.Next {
		run := tmpList.Data

		levelOr |= run.Item.Analysis.Level
		levelAnd &= run.Item.Analysis.Level

		length++
	}

	/* If none of the levels had the LSB set, all numbers were even. */
	allEven := (levelOr & 0x1) == 0

	/* If all of the levels had the LSB set, all numbers were odd. */
	allOdd := (levelAnd & 0x1) == 1

	if !allEven && !allOdd {
		line.Runs = logicalRuns.reorderRunsRecurse(length)
	} else if allOdd {
		line.Runs = logicalRuns.reverse()
	}
}

// When doing shaping, we add the letter spacing value for a
// run after every grapheme in the run. This produces ugly
// asymmetrical results, so what this routine does is redistributes
// that space to the beginning and the end of the run.
//
// We also trim the letter spacing from runs adjacent to
// tabs and from the outside runs of the lines so that things
// line up properly. The line breaking and tab positioning
// were computed without this trimming so they are no longer
// exactly correct, but this won't be very noticeable in most
// cases.
func (line *LayoutLine) adjust_line_letter_spacing(state *paraBreakState) {
	layout := line.layout

	// If we have tab stops and the resolved direction of the
	// line is RTL, then we need to walk through the line
	// in reverse direction to figure out the corrections for
	// tab stops.
	reversed := false
	if line.ResolvedDir == DIRECTION_RTL {
		for l := line.Runs; l != nil; l = l.Next {
			if layout.is_tab_run(l.Data) {
				line.Runs = line.Runs.reverse()
				reversed = true
				break
			}
		}
	}

	// Walk over the runs in the line, redistributing letter
	// spacing from the end of the run to the start of the
	// run and trimming letter spacing from the ends of the
	// runs adjacent to the ends of the line or tab stops.
	//
	// We accumulate a correction factor from this trimming
	// which we add onto the next tab stop space to keep the
	// things properly aligned.
	var lastRun *GlyphItem
	var tabAdjustment GlyphUnit
	for l := line.Runs; l != nil; l = l.Next {
		run := l.Data
		var nextRun *GlyphItem
		if l.Next != nil {
			nextRun = l.Next.Data
		}

		if layout.is_tab_run(run) {
			run.Glyphs.pad_glyphstring_right(state, tabAdjustment)
			tabAdjustment = 0
		} else {
			visualNextRun, visualLastRun := nextRun, lastRun
			if reversed {
				visualNextRun, visualLastRun = lastRun, nextRun
			}
			runSpacing := run.Item.get_item_letter_spacing()

			spaceLeft, spaceRight := distributeLetterSpacing(runSpacing)

			if run.Glyphs.Glyphs[0].Geometry.Width == 0 {
				/* we've zeroed this space glyph at the end of line, now remove
				* the letter spacing added to its adjacent glyph */
				run.Glyphs.pad_glyphstring_left(state, -spaceLeft)
			} else if visualLastRun == nil || layout.is_tab_run(visualLastRun) {
				run.Glyphs.pad_glyphstring_left(state, -spaceLeft)
				tabAdjustment += spaceLeft
			}

			if run.Glyphs.Glyphs[len(run.Glyphs.Glyphs)-1].Geometry.Width == 0 {
				/* we've zeroed this space glyph at the end of line, now remove
				* the letter spacing added to its adjacent glyph */
				run.Glyphs.pad_glyphstring_right(state, -spaceRight)
			} else if visualNextRun == nil || layout.is_tab_run(visualNextRun) {
				run.Glyphs.pad_glyphstring_right(state, -spaceRight)
				tabAdjustment += spaceRight
			}
		}

		lastRun = run
	}

	if reversed {
		line.Runs = line.Runs.reverse()
	}
}

const (
	measure = iota
	adjust
)

func (line *LayoutLine) justifyWords(state *paraBreakState) {
	text := line.layout.Text
	logAttrs := line.layout.logAttrs

	var addedSoFar, spacesSoFar, total_space_width GlyphUnit
	//    GSList *run_iter;

	totalRemainingWidth := state.remaining_width
	if totalRemainingWidth <= 0 {
		return
	}

	// hint to full pixel if total remaining width was so
	isHinted := (totalRemainingWidth & (Scale - 1)) == 0

	for mode := measure; mode <= adjust; mode++ {
		addedSoFar = 0
		spacesSoFar = 0

		for runIter := line.Runs; runIter != nil; runIter = runIter.Next {
			run := runIter.Data
			glyphs := run.Glyphs

			// We need character offset of the start of the run.  We don't have this.
			// Compute by counting from the beginning of the line.  The naming is
			// confusing.  Note that:
			//
			// run.Item.Offset        is byte offset of start of run in layout.Text.
			// state.lineStartIndex  is byte offset of start of line in layout.Text.
			// state.line_startOffset is character offset of start of line in layout.Text.
			if debugMode {
				assert(run.Item.Offset >= state.lineStartIndex, "justifyWords")
			}
			offset := state.lineStartIndex + run.Item.Offset - state.lineStartIndex
			var clusterIter GlyphItemIter
			haveCluster := clusterIter.InitStart(run, text)
			for ; haveCluster; haveCluster = clusterIter.NextCluster() {

				if !logAttrs[offset+clusterIter.StartChar].IsExpandableSpace() {
					continue
				}

				dir := -1
				if clusterIter.startGlyph < clusterIter.endGlyph {
					dir = 1
				}
				for i := clusterIter.startGlyph; i != clusterIter.endGlyph; i += dir {
					glyph_width := glyphs.Glyphs[i].Geometry.Width

					if glyph_width == 0 {
						continue
					}

					spacesSoFar += glyph_width

					if mode == adjust {
						adjustment := GlyphUnit(uint64(spacesSoFar)*uint64(totalRemainingWidth)/uint64(total_space_width)) - addedSoFar
						if isHinted {
							adjustment = adjustment.Round()
						}

						glyphs.Glyphs[i].Geometry.Width += adjustment
						addedSoFar += adjustment
					}
				}
			}
		}

		if mode == measure {
			total_space_width = spacesSoFar

			if total_space_width == 0 {
				line.justify_clusters(state)
				return
			}
		}
	}

	state.remaining_width -= addedSoFar
}

func (line *LayoutLine) justify_clusters(state *paraBreakState) {
	text := line.layout.Text
	logAttrs := line.layout.logAttrs

	var addedSoFar, gapsSoFar, totalGaps GlyphUnit
	//    bool isHinted;
	//    GSList *run_iter;

	totalRemainingWidth := state.remaining_width
	if totalRemainingWidth <= 0 {
		return
	}

	/* hint to full pixel if total remaining width was so */
	isHinted := (totalRemainingWidth & (Scale - 1)) == 0

	for mode := measure; mode <= adjust; mode++ {
		var (
			residual        GlyphUnit
			leftedge        = true
			rightmostGlyphs *GlyphString
			rightmostSpace  GlyphUnit
		)
		addedSoFar = 0
		gapsSoFar = 0

		for run_iter := line.Runs; run_iter != nil; run_iter = run_iter.Next {
			run := run_iter.Data
			glyphs := run.Glyphs
			//    PangoGlyphItemIter clusterIter;
			//    bool haveCluster;
			//    int dir;
			//    int offset;

			dir := -1
			if run.LTR() {
				dir = +1
			}

			// We need character offset of the start of the run.  We don't have this.
			// Compute by counting from the beginning of the line.  The naming is
			// confusing.  Note that:
			//
			// run.Item.Offset        is rune offset of start of run in layout.Text.
			// state.lineStartIndex  is rune offset of start of line in layout.Text.
			// state.line_startOffset is character offset of start of line in layout.Text.
			if debugMode {
				assert(run.Item.Offset >= state.lineStartIndex, "justifyClusters")
			}

			offset := state.lineStartIndex + run.Item.Offset - state.lineStartIndex

			var (
				clusterIter GlyphItemIter
				haveCluster bool
			)
			if dir > 0 {
				haveCluster = clusterIter.InitStart(run, text)
			} else {
				haveCluster = clusterIter.InitEnd(run, text)
			}
			for haveCluster {
				/* don't expand in the middle of graphemes */
				if !logAttrs[offset+clusterIter.StartChar].IsCursorPosition() {
					continue
				}

				var width GlyphUnit
				for i := clusterIter.startGlyph; i != clusterIter.endGlyph; i += dir {
					width += glyphs.Glyphs[i].Geometry.Width
				}

				/* also don't expand zero-width clusters. */
				if width == 0 {
					continue
				}

				gapsSoFar++

				if mode == adjust {

					adjustment := totalRemainingWidth/totalGaps + residual
					if isHinted {
						old_adjustment := adjustment
						adjustment = adjustment.Round()
						residual = old_adjustment - adjustment
					}
					/* distribute to before/after */
					spaceLeft, spaceRight := distributeLetterSpacing(adjustment)

					var leftmost, rightmost int
					if clusterIter.startGlyph < clusterIter.endGlyph {
						/* LTR */
						leftmost = clusterIter.startGlyph
						rightmost = clusterIter.endGlyph - 1
					} else {
						/* RTL */
						leftmost = clusterIter.endGlyph + 1
						rightmost = clusterIter.startGlyph
					}
					/* Don't add to left-side of left-most glyph of left-most non-zero run. */
					if leftedge {
						leftedge = false
					} else {
						glyphs.Glyphs[leftmost].Geometry.Width += spaceLeft
						glyphs.Glyphs[leftmost].Geometry.XOffset += spaceLeft
						addedSoFar += spaceLeft
					}
					/* Don't add to right-side of right-most glyph of right-most non-zero run. */
					{
						/* Save so we can undo later. */
						rightmostGlyphs = glyphs
						rightmostSpace = spaceRight

						glyphs.Glyphs[rightmost].Geometry.Width += spaceRight
						addedSoFar += spaceRight
					}
				}

				if dir > 0 {
					haveCluster = clusterIter.NextCluster()
				} else {
					haveCluster = clusterIter.PrevCluster()
				}
			}
		}

		if mode == measure {
			totalGaps = gapsSoFar - 1

			if totalGaps == 0 {
				/* a single cluster, can't really justify it */
				return
			}
		} else /* mode == ADJUST */ {
			if rightmostGlyphs != nil {
				rightmostGlyphs.Glyphs[len(rightmostGlyphs.Glyphs)-1].Geometry.Width -= rightmostSpace
				addedSoFar -= rightmostSpace
			}
		}
	}

	state.remaining_width -= addedSoFar
}

func (line *LayoutLine) addLine(state *paraBreakState) {
	layout := line.layout

	layout.lines = append(layout.lines, line)

	if layout.Height >= 0 {
		var logicalRect Rectangle
		line.GetExtents(nil, &logicalRect)
		state.remainingHeight -= logicalRect.Height
		state.remainingHeight -= layout.Spacing
		state.lineHeight = logicalRect.Height
	}
}

// GetExtents computes the logical and ink extents of a layout line. See
// Font.getGlyphExtents() for details about the interpretation
// of the rectangles.
func (line *LayoutLine) GetExtents(inkRect, logicalRect *Rectangle) {
	line.getExtentsAndHeight(inkRect, logicalRect, nil)
}

func (line *LayoutLine) getExtentsAndHeight(inkRect, logicalRect *Rectangle, height *GlyphUnit) {
	if line == nil || line.layout == nil {
		return
	}

	caching := false

	if inkRect == nil && logicalRect == nil && height == nil {
		return
	}

	switch line.cache_status {
	case cached:
		if inkRect != nil {
			*inkRect = line.inkRect
		}
		if logicalRect != nil {
			*logicalRect = line.logicalRect
		}
		if height != nil {
			*height = line.height
		}
		return
	case notCached:
		caching = true
		if inkRect == nil {
			inkRect = &line.inkRect
		}
		if logicalRect == nil {
			logicalRect = &line.logicalRect
		}
		if height == nil {
			height = &line.height
		}
	case leaked:
	}

	if inkRect != nil {
		inkRect.X, inkRect.Y, inkRect.Width, inkRect.Height = 0, 0, 0, 0
	}

	if logicalRect != nil {
		logicalRect.X, logicalRect.Y, logicalRect.Width, logicalRect.Height = 0, 0, 0, 0
	}

	if height != nil {
		*height = 0
	}

	var xPos GlyphUnit
	tmpList := line.Runs
	for tmpList != nil {
		run := tmpList.Data
		var (
			runInk, runLogical Rectangle
			newPos, runHeight  GlyphUnit
		)
		run.getExtentsAndHeight(&runInk, nil, &runLogical, &runHeight)

		if inkRect != nil {
			if inkRect.Width == 0 || inkRect.Height == 0 {
				*inkRect = runInk
				inkRect.X += xPos
			} else if runInk.Width != 0 && runInk.Height != 0 {
				newPos = minG(inkRect.X, xPos+runInk.X)
				inkRect.Width = maxG(inkRect.X+inkRect.Width, xPos+runInk.X+runInk.Width) - newPos
				inkRect.X = newPos

				newPos = minG(inkRect.Y, runInk.Y)
				inkRect.Height = maxG(inkRect.Y+inkRect.Height, runInk.Y+runInk.Height) - newPos
				inkRect.Y = newPos
			}
		}

		if logicalRect != nil {
			newPos = minG(logicalRect.X, xPos+runLogical.X)
			logicalRect.Width = maxG(logicalRect.X+logicalRect.Width, xPos+runLogical.X+runLogical.Width) - newPos
			logicalRect.X = newPos

			newPos = minG(logicalRect.Y, runLogical.Y)
			logicalRect.Height = maxG(logicalRect.Y+logicalRect.Height, runLogical.Y+runLogical.Height) - newPos
			logicalRect.Y = newPos
		}

		if height != nil {
			*height = maxG(*height, runHeight)
		}

		xPos += runLogical.Width
		tmpList = tmpList.Next
	}

	if line.Runs == nil {
		rect := logicalRect
		if rect == nil {
			rect = &Rectangle{}
		}
		*height = line.getEmptyExtentsAndHeight(logicalRect)
	}

	if caching {
		if &line.inkRect != inkRect {
			line.inkRect = *inkRect
		}
		if &line.logicalRect != logicalRect {
			line.logicalRect = *logicalRect
		}
		if &line.height != height {
			line.height = *height
		}
		line.cache_status = cached
	}
}

func (line *LayoutLine) getEmptyExtentsAndHeight(logicalRect *Rectangle) GlyphUnit {
	return line.layout.getEmptyExtentsAndHeightAt(line.StartIndex, logicalRect)
}

func (line *LayoutLine) getAlignment() Alignment {
	layout := line.layout
	alignment := layout.alignment

	if alignment != ALIGN_CENTER && line.layout.autoDir &&
		line.ResolvedDir.directionSimple() == -layout.context.baseDir.directionSimple() {
		if alignment == ALIGN_LEFT {
			alignment = ALIGN_RIGHT
		} else if alignment == ALIGN_RIGHT {
			alignment = ALIGN_LEFT
		}
	}

	return alignment
}

func (line *LayoutLine) get_x_offset(layout *Layout, layoutWidth, lineWidth GlyphUnit) GlyphUnit {
	alignment := line.getAlignment()

	var xOffset GlyphUnit
	// Alignment
	if layoutWidth == 0 {
		xOffset = 0
	} else if alignment == ALIGN_RIGHT {
		xOffset = layoutWidth - lineWidth
	} else if alignment == ALIGN_CENTER {
		xOffset = (layoutWidth - lineWidth) / 2
		// hinting
		if (layoutWidth|lineWidth)&(Scale-1) == 0 {
			xOffset = xOffset.Round()
		}
	}

	// Indentation

	/* For center, we ignore indentation; I think I've seen word
	* processors that still do the indentation here as if it were
	* indented left/right, though we can't sensibly do that without
	* knowing whether left/right is the "normal" thing for this text */

	if alignment == ALIGN_CENTER {
		return xOffset
	}

	if line.IsParagraphStart {
		if layout.Indent > 0 {
			if alignment == ALIGN_LEFT {
				xOffset += layout.Indent
			} else {
				xOffset -= layout.Indent
			}
		}
	} else {
		if layout.Indent < 0 {
			if alignment == ALIGN_LEFT {
				xOffset -= layout.Indent
			} else {
				xOffset += layout.Indent
			}
		}
	}
	return xOffset
}

func (line *LayoutLine) getLineExtentsLayoutCoords(layout *Layout,
	layoutWidth GlyphUnit, yOffset GlyphUnit, baseline *GlyphUnit,
	lineInkLayout, lineLogicalLayout *Rectangle) {
	var (
		// Line extents in line coords (origin at line baseline)
		lineInk, lineLogical Rectangle
		height, newBaseline  GlyphUnit
	)

	firstLine := false
	if len(layout.lines) != 0 && layout.lines[0] == line {
		firstLine = true
	}

	line.getExtentsAndHeight(&lineInk, &lineLogical, &height)

	xOffset := line.get_x_offset(layout, layoutWidth, GlyphUnit(lineLogical.Width))

	if firstLine || baseline == nil || layout.LineSpacing == 0.0 {
		newBaseline = yOffset - lineLogical.Y
	} else {
		newBaseline = *baseline + GlyphUnit(layout.LineSpacing*float32(height))
	}

	// Convert the line's extents into layout coordinates
	if lineInkLayout != nil {
		*lineInkLayout = lineInk
		lineInkLayout.X = lineInk.X + xOffset
		lineInkLayout.Y = newBaseline + lineInk.Y
	}

	if lineLogicalLayout != nil {
		*lineLogicalLayout = lineLogical
		lineLogicalLayout.X = lineLogical.X + xOffset
		lineLogicalLayout.Y = newBaseline + lineLogical.Y
	}

	if baseline != nil {
		*baseline = newBaseline
	}
}

func (line *LayoutLine) getCharDirection(index int) Direction {
	for runList := line.Runs; runList != nil; runList = runList.Next {
		run := runList.Data

		if run.Item.Offset <= index && run.Item.Offset+run.Item.Length > index {
			if run.LTR() {
				return DIRECTION_LTR
			}
			return DIRECTION_RTL
		}

	}

	return DIRECTION_LTR
}

func (line *LayoutLine) getCharLevel(index int) fribidi.Level {
	for runList := line.Runs; runList != nil; runList = runList.Next {
		run := runList.Data

		if run.Item.Offset <= index && run.Item.Offset+run.Item.Length > index {
			return run.Item.Analysis.Level
		}
	}

	return 0
}

// IndexToX converts an index within a line to a X position
// `trailing` indicates the edge of the grapheme to retrieve
// the position of : if true, the trailing edge of the grapheme,
// else the leading of the grapheme.
func (line *LayoutLine) IndexToX(index int, trailing bool) GlyphUnit {
	layout := line.layout
	var width GlyphUnit

	for runList := line.Runs; runList != nil; runList = runList.Next {
		run := runList.Data

		if run.Item.Offset <= index && run.Item.Offset+run.Item.Length > index {
			if trailing {
				for index < line.StartIndex+line.Length && index+1 < len(layout.Text) &&
					!layout.logAttrs[index+1].IsCursorPosition() {
					index++
				}
			} else {
				for index > line.StartIndex && !layout.logAttrs[index].IsCursorPosition() {
					index--
				}
			}

			attrOffset := run.Item.Offset
			xPos := run.Glyphs.indexToXFull(layout.Text[attrOffset:attrOffset+run.Item.Length],
				&run.Item.Analysis, index-attrOffset, trailing, layout.logAttrs[attrOffset:])
			xPos += width

			return xPos
		}

		width += run.Glyphs.getWidth()
	}

	return width
}

// GetXRanges gets a list of visual ranges corresponding to a given logical range.
// `startIndex` is the start rune index of the logical range. If this value
//   is less than the start index for the line, then the first range
//   will extend all the way to the leading edge of the layout. Otherwise,
//   it will start at the leading edge of the first character.
// `endIndex` is the ending rune index of the logical range. If this value is
//   greater than the end index for the line, then the last range will
//   extend all the way to the trailing edge of the layout. Otherwise,
//   it will end at the trailing edge of the last character.
//
// 	The returned slice will be of length
//   `2*nRanges`, with each range starting at `ranges[2*n]` and of
//   width `ranges[2*n + 1] - ranges[2*n]`. The coordinates are relative to the layout and are in
//   Pango units.
//
// This list is not necessarily minimal - there may be consecutive
// ranges which are adjacent. The ranges will be sorted from left to
// right. The ranges are with respect to the left edge of the entire
// layout, not with respect to the line.
func (line *LayoutLine) GetXRanges(startIndex, endIndex int) []GlyphUnit {
	if line.layout == nil || startIndex > endIndex {
		return nil
	}
	alignment := line.getAlignment()

	width := line.layout.Width
	if width == -1 && alignment != ALIGN_LEFT {
		var logicalRect Rectangle
		line.layout.GetExtents(nil, &logicalRect)
		width = logicalRect.Width
	}

	var logicalRect Rectangle
	line.GetExtents(nil, &logicalRect)
	lineWidth := logicalRect.Width

	xOffset := line.get_x_offset(line.layout, width, lineWidth)

	lineStartIndex := line.StartIndex

	/* Allocate the maximum possible size */
	ranges := make([]GlyphUnit, 0, 2*(2+line.Runs.length()))
	if xOffset > 0 &&
		((line.ResolvedDir == DIRECTION_LTR && startIndex < lineStartIndex) ||
			(line.ResolvedDir == DIRECTION_RTL && endIndex > lineStartIndex+line.Length)) {
		ranges = append(ranges, 0, xOffset)
	}

	var accumulatedWidth GlyphUnit
	for tmpList := line.Runs; tmpList != nil; tmpList = tmpList.Next {
		run := tmpList.Data
		if startIndex < run.Item.Offset+run.Item.Length &&
			endIndex > run.Item.Offset {
			runStartIndex := max(startIndex, run.Item.Offset)
			runEndIndex := min(endIndex, run.Item.Offset+run.Item.Length)

			if debugMode {
				assert(runEndIndex > 0, "GetXRanges")
			}

			// back the endIndex off one since we want to find the trailing edge of the preceding character
			runEndIndex--

			runStartX := run.Glyphs.IndexToX(line.layout.Text[run.Item.Offset:run.Item.Offset+run.Item.Length],
				&run.Item.Analysis,
				runStartIndex-run.Item.Offset, false)
			runEndX := run.Glyphs.IndexToX(line.layout.Text[run.Item.Offset:run.Item.Offset+run.Item.Length],
				&run.Item.Analysis,
				runEndIndex-run.Item.Offset, true)

			ranges = append(ranges,
				xOffset+accumulatedWidth+minG(runStartX, runEndX),
				xOffset+accumulatedWidth+maxG(runStartX, runEndX),
			)
		}

		if tmpList.Next != nil {
			accumulatedWidth += run.Glyphs.getWidth()
		}
	}

	if xOffset+lineWidth < line.layout.Width &&
		((line.ResolvedDir == DIRECTION_LTR && endIndex > lineStartIndex+line.Length) ||
			(line.ResolvedDir == DIRECTION_RTL && startIndex < lineStartIndex)) {
		ranges = append(ranges, xOffset+lineWidth, line.layout.Width)
	}

	return ranges
}

type baselineItem struct {
	attr             *Attribute
	xOffset, yOffset GlyphUnit
}

func (state *paraBreakState) collectBaselineShift(item, prev *Item) (startXOffset, startYOffset, endXOffset, endYOffset GlyphUnit) {
	for _, attr := range item.Analysis.ExtraAttrs {
		if attr.Kind == ATTR_RISE {
			value := GlyphUnit(attr.Data.(AttrInt))
			startYOffset += value
			endYOffset -= value
		} else if attr.Kind == ATTR_BASELINE_SHIFT {
			if attr.StartIndex == item.Offset {

				entry := baselineItem{attr: attr}
				state.baselineShifts.PushFront(entry)

				value := GlyphUnit(attr.Data.(AttrInt))

				if value > 1024 || value < -1024 {
					entry.yOffset = value
				} else {
					var superscriptXOffset, superscriptYOffset, subscriptXOffset, subscriptYOffset float32

					if prev != nil {
						face := prev.Analysis.Font.GetHarfbuzzFont().Face()
						superscriptYOffset, _ = face.LineMetric(fonts.SuperscriptEmYSize)
						superscriptXOffset, _ = face.LineMetric(fonts.SuperscriptEmXOffset)
						subscriptYOffset, _ = face.LineMetric(fonts.SubscriptEmYOffset)
						subscriptXOffset, _ = face.LineMetric(fonts.SubscriptEmXOffset)
					}

					if superscriptYOffset == 0 {
						superscriptYOffset = 5000
					}
					if subscriptYOffset == 0 {
						subscriptYOffset = 5000
					}

					switch BaselineShift(value) {
					case BASELINE_SHIFT_NONE:
						entry.xOffset = 0
						entry.yOffset = 0
					case BASELINE_SHIFT_SUPERSCRIPT:
						entry.xOffset = GlyphUnit(superscriptXOffset)
						entry.yOffset = GlyphUnit(superscriptYOffset)
					case BASELINE_SHIFT_SUBSCRIPT:
						entry.xOffset = GlyphUnit(subscriptXOffset)
						entry.yOffset = GlyphUnit(-subscriptYOffset)
					}
				}

				startXOffset += entry.xOffset
				startYOffset += entry.yOffset
			}

			if attr.EndIndex == item.Offset+item.Length {
				var t *list.Element

				for t = state.baselineShifts.Front(); t != nil; t = t.Next() {
					entry := t.Value.(baselineItem)

					if attr.StartIndex == entry.attr.StartIndex &&
						attr.EndIndex == entry.attr.EndIndex &&
						attr.Data.(AttrInt) == entry.attr.Data.(AttrInt) {
						endXOffset -= entry.xOffset
						endYOffset -= entry.yOffset
					}

					state.baselineShifts.Remove(t)
					break
				}
				if t == nil && debugMode {
					log.Println("Baseline attributes mismatch")
				}
			}
		}
	}
	return
}

func (line *LayoutLine) applyBaselineShift(state *paraBreakState) {
	var (
		yOffset GlyphUnit
		prev    *Item
	)
	for l := line.Runs; l != nil; l = l.Next {
		run := l.Data
		item := run.Item

		startXOffset, startYOffset, endXOffset, endYOffset := state.collectBaselineShift(item, prev)

		yOffset += startYOffset

		run.yOffset = yOffset
		run.startXOffset = startXOffset
		run.endXOffset = endXOffset

		yOffset += endYOffset

		prev = item
	}
}
