package pango

import (
	"math"

	"github.com/benoitkugler/go-weasyprint/fribidi"
)

/* Extents cache status:
*
* LEAKED means that the user has access to this line structure or a
* run included in this line, and so can change the glyphs/glyph-widths.
* If this is true, extents caching will be disabled.
 */
const (
	NOT_CACHED uint8 = iota
	CACHED
	LEAKED
)

type LayoutLinePrivate struct {
	LayoutLine

	cache_status uint8
	inkRect      Rectangle
	logicalRect  Rectangle
	height       int
}

func (layout *Layout) pango_layout_line_new() *LayoutLinePrivate {
	var private LayoutLinePrivate
	private.LayoutLine.layout = layout
	return &private
}

type runList struct {
	data *GlyphItem
	next *runList
}

func (l *runList) reverse() *runList {
	var out *runList

	for ; l != nil; l = l.next {
		out = &runList{data: l.data, next: out}
	}

	return out
}

func (l *runList) length() int {
	out := 0
	for ; l != nil; l = l.next {
		out++
	}
	return out
}

func (l *runList) concat(snd *runList) *runList {
	if l == nil {
		return snd
	}
	head := l
	for ; l.next != nil; l = l.next {
	}
	l.next = snd
	return head
}

// NB: This implement the exact same algorithm as pango_reorder_items().
func (items *runList) reorderRunsRecurse(nItems int) *runList {
	//    GSList *tmpList, *levelStartNode;
	//    int i, levelStartI;
	//    GSList *result = nil;

	if nItems == 0 {
		return nil
	}

	tmpList := items
	var minLevel fribidi.Level = math.MaxInt8
	for i := 0; i < nItems; i++ {
		run := tmpList.data
		minLevel = minL(minLevel, run.item.analysis.level)

		tmpList = tmpList.next
	}

	var (
		result *runList
		i      int
	)
	levelStartI := 0
	levelStartNode := items
	tmpList = items
	for i = 0; i < nItems; i++ {
		run := tmpList.data

		if run.item.analysis.level == minLevel {
			if minLevel%2 != 0 {
				if i > levelStartI {
					result = levelStartNode.reorderRunsRecurse(i - levelStartI).concat(result)
				}
				result = &runList{data: run, next: result}
			} else {
				if i > levelStartI {
					result = result.concat(levelStartNode.reorderRunsRecurse(i - levelStartI))
				}
				result = result.concat(&runList{data: run})
			}

			levelStartI = i + 1
			levelStartNode = tmpList.next
		}

		tmpList = tmpList.next
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

// LayoutLine represents one of the lines resulting
// from laying out a paragraph via `Layout`. `LayoutLine`
// structures are obtained by calling pango_layout_get_line() and
// are only valid until the text, attributes, or settings of the
// parent `Layout` are modified.
type LayoutLine struct {
	layout             *Layout   // the layout this line belongs to, might be nil
	start_index        int       // start of line as rune index into layout.text
	length             int       // length of line in runes
	runs               *runList  // list of runs in the line, from left to right
	is_paragraph_start bool      // = 1;  // true if this is the first line of the paragraph
	resolved_dir       Direction // = 3;  // Resolved PangoDirection of line
}

// The resolved direction for the line is always one
// of LTR/RTL; not a week or neutral directions
func (line *LayoutLine) line_set_resolved_dir(direction Direction) {
	switch direction {
	case PANGO_DIRECTION_RTL, PANGO_DIRECTION_WEAK_RTL:
		line.resolved_dir = PANGO_DIRECTION_RTL
	default:
		line.resolved_dir = PANGO_DIRECTION_LTR
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
	switch line.layout.context.resolved_gravity {
	case PANGO_GRAVITY_NORTH:
		line.resolved_dir = PANGO_DIRECTION_LTR + PANGO_DIRECTION_RTL - line.resolved_dir
	case PANGO_GRAVITY_EAST:
		// This is in fact why deprecated TTB_RTL is LTR
		line.resolved_dir = PANGO_DIRECTION_LTR
	case PANGO_GRAVITY_WEST:
		// This is in fact why deprecated TTB_LTR is RTL
		line.resolved_dir = PANGO_DIRECTION_RTL
	}
}

func (line *LayoutLine) shape_run(state *ParaBreakState, item *Item) *GlyphString {
	layout := line.layout
	glyphs := &GlyphString{}

	if layout.text[item.offset] == '\t' {
		line.shape_tab(item, glyphs)
	} else {
		shapeFlags := PANGO_SHAPE_NONE

		if layout.context.round_glyph_positions {
			shapeFlags |= PANGO_SHAPE_ROUND_POSITIONS
		}
		if state.properties.shape != nil {
			glyphs._pango_shape_shape(layout.text[item.offset:item.offset+item.num_chars], state.properties.shape.logical)
		} else {
			glyphs.pango_shape_with_flags(layout.text[item.offset:item.offset+item.num_chars],
				layout.text, &item.analysis, shapeFlags)
		}

		if state.properties.letter_spacing != 0 {
			glyphItem := GlyphItem{item: item, glyphs: glyphs}

			glyphItem.pango_glyph_item_letter_space(layout.text,
				layout.log_attrs[state.start_offset:],
				state.properties.letter_spacing)

			spaceLeft, spaceRight := distributeLetterSpacing(state.properties.letter_spacing)

			glyphs.glyphs[0].geometry.width += spaceLeft
			glyphs.glyphs[0].geometry.x_offset += spaceLeft
			glyphs.glyphs[len(glyphs.glyphs)-1].geometry.width += spaceRight
		}
	}

	return glyphs
}

func distributeLetterSpacing(letterSpacing GlyphUnit) (spaceLeft, spaceRight GlyphUnit) {
	spaceLeft = letterSpacing / 2
	// hinting
	if (letterSpacing & (pangoScale - 1)) == 0 {
		spaceLeft = spaceLeft.PANGO_UNITS_ROUND()
	}
	spaceRight = letterSpacing - spaceLeft
	return
}

func (line *LayoutLine) shape_tab(item *Item, glyphs *GlyphString) {
	current_width := line.line_width()

	glyphs.pango_glyph_string_set_size(1)

	if item.analysis.showing_space() {
		glyphs.glyphs[0].glyph = PANGO_GET_UNKNOWN_GLYPH('\t')
	} else {
		glyphs.glyphs[0].glyph = PANGO_GLYPH_EMPTY
	}
	glyphs.glyphs[0].geometry.x_offset = 0
	glyphs.glyphs[0].geometry.y_offset = 0
	glyphs.glyphs[0].attr.is_cluster_start = true

	glyphs.log_clusters[0] = 0

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
			glyphs.glyphs[0].geometry.width = GlyphUnit(tab_pos) - current_width
			break
		}
	}
}

func (line *LayoutLine) pango_layout_line_get_width() GlyphUnit {
	var width GlyphUnit
	for l := line.runs; l != nil; l = l.next {
		width += l.data.glyphs.pango_glyph_string_get_width()
	}
	return width
}

func (line *LayoutLine) line_width() GlyphUnit {
	var width GlyphUnit

	// Compute the width of the line currently - inefficient, but easier
	// than keeping the current width of the line up to date everywhere
	for l := line.runs; l != nil; l = l.next {
		for _, info := range l.data.glyphs.glyphs {
			width += info.geometry.width
		}
	}

	return width
}

func (line *LayoutLine) insert_run(state *ParaBreakState, runItem *Item, lastRun bool) {

	run := GlyphItem{item: runItem}

	if lastRun && state.log_widths_offset == 0 {
		run.glyphs = state.glyphs
	} else {
		run.glyphs = line.shape_run(state, runItem)
	}

	if lastRun {
		state.glyphs = nil
		state.log_widths = nil
		state.need_hyphen = nil
	}

	line.runs = &runList{data: &run, next: line.runs} // prepend
	line.length += runItem.num_chars
}

func (line *LayoutLine) uninsert_run() *Item {
	runItem := line.runs.data.item

	line.runs = line.runs.next
	line.length -= runItem.num_chars

	return runItem
}

func (line *LayoutLine) pango_layout_line_postprocess(state *ParaBreakState, wrapped bool) {
	ellipsized := false

	if debugMode {
		showDebug("postprocessing", line, state)
	}

	// Truncate the logical-final whitespace in the line if we broke the line at it
	if wrapped {
		// The runs are in reverse order at this point, since we prepended them to the list.
		// So, the first run is the last logical run.
		line.zero_line_final_space(state, line.runs.data)
	}

	// Reverse the runs
	line.runs = line.runs.reverse()

	// Ellipsize the line if necessary
	if state.line_width >= 0 && line.layout.should_ellipsize_current_line(state) {
		shape_flags := PANGO_SHAPE_NONE

		if line.layout.context.round_glyph_positions {
			shape_flags |= PANGO_SHAPE_ROUND_POSITIONS
		}

		ellipsized = line._pango_layout_line_ellipsize(state.attrs, shape_flags, state.line_width)
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
	if line.layout.justify && (wrapped || ellipsized) {
		/* if we ellipsized, we don't have remaining_width set */
		if state.remaining_width < 0 {
			state.remaining_width = state.line_width - line.pango_layout_line_get_width()
		}

		line.justifyWords(state)
	}

	if debugMode {
		showDebug("after justification", line, state)
	}

	line.layout.is_wrapped = line.layout.is_wrapped || wrapped
	line.layout.is_ellipsized = line.layout.is_ellipsized || ellipsized
}

func (line *LayoutLine) zero_line_final_space(state *ParaBreakState, run *GlyphItem) {
	layout := line.layout
	glyphs := run.glyphs

	glyph := 0
	if run.LTR() {
		glyph = len(glyphs.glyphs) - 1
	}

	if glyphs.glyphs[glyph].glyph == PANGO_GET_UNKNOWN_GLYPH(0x2028) {
		return // this LS is visible
	}

	// if the final char of line forms a cluster, and it's
	// a whitespace char, zero its glyph's width as it's been wrapped
	if len(glyphs.glyphs) < 1 || state.start_offset == 0 ||
		!layout.log_attrs[state.start_offset-1].IsWhite() {
		return
	}

	offset := 1
	if run.LTR() {
		offset = -1
	}
	if len(glyphs.glyphs) >= 2 && glyphs.log_clusters[glyph] == glyphs.log_clusters[glyph+offset] {
		return
	}

	state.remaining_width += glyphs.glyphs[glyph].geometry.width
	glyphs.glyphs[glyph].geometry.width = 0
	glyphs.glyphs[glyph].glyph = PANGO_GLYPH_EMPTY
}

func (line *LayoutLine) pangoLayoutLineReorder() {
	logicalRuns := line.runs
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
	for tmpList := logicalRuns; tmpList != nil; tmpList = tmpList.next {
		run := tmpList.data

		levelOr |= run.item.analysis.level
		levelAnd &= run.item.analysis.level

		length++
	}

	/* If none of the levels had the LSB set, all numbers were even. */
	allEven := (levelOr & 0x1) == 0

	/* If all of the levels had the LSB set, all numbers were odd. */
	allOdd := (levelAnd & 0x1) == 1

	if !allEven && !allOdd {
		line.runs = logicalRuns.reorderRunsRecurse(length)
	} else if allOdd {
		line.runs = logicalRuns.reverse()
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
func (line *LayoutLine) adjust_line_letter_spacing(state *ParaBreakState) {
	layout := line.layout

	// If we have tab stops and the resolved direction of the
	// line is RTL, then we need to walk through the line
	// in reverse direction to figure out the corrections for
	// tab stops.
	reversed := false
	if line.resolved_dir == PANGO_DIRECTION_RTL {
		for l := line.runs; l != nil; l = l.next {
			if layout.is_tab_run(l.data) {
				line.runs = line.runs.reverse()
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
	for l := line.runs; l != nil; l = l.next {
		run := l.data
		var nextRun *GlyphItem
		if l.next != nil {
			nextRun = l.next.data
		}

		if layout.is_tab_run(run) {
			run.glyphs.pad_glyphstring_right(state, tabAdjustment)
			tabAdjustment = 0
		} else {
			visualNextRun, visualLastRun := nextRun, lastRun
			if reversed {
				visualNextRun, visualLastRun = lastRun, nextRun
			}
			runSpacing := run.item.get_item_letter_spacing()

			spaceLeft, spaceRight := distributeLetterSpacing(runSpacing)

			if run.glyphs.glyphs[0].geometry.width == 0 {
				/* we've zeroed this space glyph at the end of line, now remove
				* the letter spacing added to its adjacent glyph */
				run.glyphs.pad_glyphstring_left(state, -spaceLeft)
			} else if visualLastRun == nil || layout.is_tab_run(visualLastRun) {
				run.glyphs.pad_glyphstring_left(state, -spaceLeft)
				tabAdjustment += spaceLeft
			}

			if run.glyphs.glyphs[len(run.glyphs.glyphs)-1].geometry.width == 0 {
				/* we've zeroed this space glyph at the end of line, now remove
				* the letter spacing added to its adjacent glyph */
				run.glyphs.pad_glyphstring_right(state, -spaceRight)
			} else if visualNextRun == nil || layout.is_tab_run(visualNextRun) {
				run.glyphs.pad_glyphstring_right(state, -spaceRight)
				tabAdjustment += spaceRight
			}
		}

		lastRun = run
	}

	if reversed {
		line.runs = line.runs.reverse()
	}
}

const (
	MEASURE = iota
	ADJUST
)

func (line *LayoutLine) justifyWords(state *ParaBreakState) {
	text := line.layout.text
	logAttrs := line.layout.log_attrs

	var addedSoFar, spacesSoFar, total_space_width GlyphUnit
	//    GSList *run_iter;

	totalRemainingWidth := state.remaining_width
	if totalRemainingWidth <= 0 {
		return
	}

	// hint to full pixel if total remaining width was so
	isHinted := (totalRemainingWidth & (pangoScale - 1)) == 0

	for mode := MEASURE; mode <= ADJUST; mode++ {
		addedSoFar = 0
		spacesSoFar = 0

		for runIter := line.runs; runIter != nil; runIter = runIter.next {
			run := runIter.data
			glyphs := run.glyphs

			// We need character offset of the start of the run.  We don't have this.
			// Compute by counting from the beginning of the line.  The naming is
			// confusing.  Note that:
			//
			// run.item.offset        is byte offset of start of run in layout.text.
			// state.line_start_index  is byte offset of start of line in layout.text.
			// state.line_start_offset is character offset of start of line in layout.text.
			if debugMode {
				assert(run.item.offset >= state.line_start_index)
			}
			offset := state.line_start_offset + run.item.offset - state.line_start_index
			var clusterIter GlyphItemIter
			haveCluster := clusterIter.pango_glyph_item_iter_init_start(run, text)
			for ; haveCluster; haveCluster = clusterIter.pango_glyph_item_iter_next_cluster() {

				if !logAttrs[offset+clusterIter.start_char].IsExpandableSpace() {
					continue
				}

				dir := -1
				if clusterIter.start_glyph < clusterIter.end_glyph {
					dir = 1
				}
				for i := clusterIter.start_glyph; i != clusterIter.end_glyph; i += dir {
					glyph_width := glyphs.glyphs[i].geometry.width

					if glyph_width == 0 {
						continue
					}

					spacesSoFar += glyph_width

					if mode == ADJUST {
						adjustment := GlyphUnit(uint64(spacesSoFar)*uint64(totalRemainingWidth)/uint64(total_space_width)) - addedSoFar
						if isHinted {
							adjustment = adjustment.PANGO_UNITS_ROUND()
						}

						glyphs.glyphs[i].geometry.width += adjustment
						addedSoFar += adjustment
					}
				}
			}
		}

		if mode == MEASURE {
			total_space_width = spacesSoFar

			if total_space_width == 0 {
				line.justify_clusters(state)
				return
			}
		}
	}

	state.remaining_width -= addedSoFar
}

func (line *LayoutLine) justify_clusters(state *ParaBreakState) {
	text := line.layout.text
	log_attrs := line.layout.log_attrs

	var addedSoFar, gapsSoFar, totalGaps GlyphUnit
	//    bool isHinted;
	//    GSList *run_iter;

	totalRemainingWidth := state.remaining_width
	if totalRemainingWidth <= 0 {
		return
	}

	/* hint to full pixel if total remaining width was so */
	isHinted := (totalRemainingWidth & (pangoScale - 1)) == 0

	for mode := MEASURE; mode <= ADJUST; mode++ {
		var (
			residual        GlyphUnit
			leftedge        = true
			rightmostGlyphs *GlyphString
			rightmostSpace  GlyphUnit
		)
		addedSoFar = 0
		gapsSoFar = 0

		for run_iter := line.runs; run_iter != nil; run_iter = run_iter.next {
			run := run_iter.data
			glyphs := run.glyphs
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
			// run.item.offset        is rune offset of start of run in layout.text.
			// state.line_start_index  is rune offset of start of line in layout.text.
			// state.line_start_offset is character offset of start of line in layout.text.
			if debugMode {
				assert(run.item.offset >= state.line_start_index)
			}

			offset := state.line_start_offset + run.item.offset - state.line_start_index

			var (
				clusterIter GlyphItemIter
				haveCluster bool
			)
			if dir > 0 {
				haveCluster = clusterIter.pango_glyph_item_iter_init_start(run, text)
			} else {
				haveCluster = clusterIter.pango_glyph_item_iter_init_end(run, text)
			}
			for haveCluster {
				/* don't expand in the middle of graphemes */
				if !log_attrs[offset+clusterIter.start_char].IsCursorPosition() {
					continue
				}

				var width GlyphUnit
				for i := clusterIter.start_glyph; i != clusterIter.end_glyph; i += dir {
					width += glyphs.glyphs[i].geometry.width
				}

				/* also don't expand zero-width clusters. */
				if width == 0 {
					continue
				}

				gapsSoFar++

				if mode == ADJUST {

					adjustment := totalRemainingWidth/totalGaps + residual
					if isHinted {
						old_adjustment := adjustment
						adjustment = adjustment.PANGO_UNITS_ROUND()
						residual = old_adjustment - adjustment
					}
					/* distribute to before/after */
					spaceLeft, spaceRight := distributeLetterSpacing(adjustment)

					var leftmost, rightmost int
					if clusterIter.start_glyph < clusterIter.end_glyph {
						/* LTR */
						leftmost = clusterIter.start_glyph
						rightmost = clusterIter.end_glyph - 1
					} else {
						/* RTL */
						leftmost = clusterIter.end_glyph + 1
						rightmost = clusterIter.start_glyph
					}
					/* Don't add to left-side of left-most glyph of left-most non-zero run. */
					if leftedge {
						leftedge = false
					} else {
						glyphs.glyphs[leftmost].geometry.width += spaceLeft
						glyphs.glyphs[leftmost].geometry.x_offset += spaceLeft
						addedSoFar += spaceLeft
					}
					/* Don't add to right-side of right-most glyph of right-most non-zero run. */
					{
						/* Save so we can undo later. */
						rightmostGlyphs = glyphs
						rightmostSpace = spaceRight

						glyphs.glyphs[rightmost].geometry.width += spaceRight
						addedSoFar += spaceRight
					}
				}

				if dir > 0 {
					haveCluster = clusterIter.pango_glyph_item_iter_next_cluster()
				} else {
					haveCluster = clusterIter.pango_glyph_item_iter_prev_cluster()
				}
			}
		}

		if mode == MEASURE {
			totalGaps = gapsSoFar - 1

			if totalGaps == 0 {
				/* a single cluster, can't really justify it */
				return
			}
		} else /* mode == ADJUST */ {
			if rightmostGlyphs != nil {
				rightmostGlyphs.glyphs[len(rightmostGlyphs.glyphs)-1].geometry.width -= rightmostSpace
				addedSoFar -= rightmostSpace
			}
		}
	}

	state.remaining_width -= addedSoFar
}

func (line *LayoutLinePrivate) addLine(state *ParaBreakState) {
	layout := line.layout

	// TODO: append if possible. we prepend, then reverse the list later
	layout.lines = append(layout.lines, nil)
	copy(layout.lines[1:], layout.lines)
	layout.lines[0] = &line.LayoutLine
	// layout.line_count++

	if layout.height >= 0 {
		var logicalRect Rectangle
		line.pango_layout_line_get_extents(nil, &logicalRect)
		state.remaining_height -= logicalRect.height
		state.remaining_height -= layout.spacing
		state.line_height = logicalRect.height
	}
}

// pango_layout_line_get_extents computes the logical and ink extents of a layout line. See
// pango_font_get_glyph_extents() for details about the interpretation
// of the rectangles.
func (line *LayoutLinePrivate) pango_layout_line_get_extents(inkRect, logicalRect *Rectangle) {
	line.pango_layout_line_get_extents_and_height(inkRect, logicalRect, nil)
}

func (private *LayoutLinePrivate) pango_layout_line_get_extents_and_height(inkRect, logicalRect *Rectangle, height *int) {
	if private == nil || private.layout == nil {
		return
	}

	//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;
	//    GSList *tmpList;
	//    int xPos = 0;
	caching := false

	if inkRect == nil && logicalRect == nil {
		return
	}

	switch private.cache_status {
	case CACHED:
		if inkRect != nil {
			*inkRect = private.inkRect
		}
		if logicalRect != nil {
			*logicalRect = private.logicalRect
		}
		if height != nil {
			*height = private.height
		}
		return
	case NOT_CACHED:
		caching = true
		if inkRect == nil {
			inkRect = &private.inkRect
		}
		if logicalRect == nil {
			logicalRect = &private.logicalRect
		}
		if height == nil {
			height = &private.height
		}
	case LEAKED:
	}

	if inkRect != nil {
		inkRect.x, inkRect.y, inkRect.width, inkRect.height = 0, 0, 0, 0
	}

	if logicalRect != nil {
		logicalRect.x, logicalRect.y, logicalRect.width, logicalRect.height = 0, 0, 0, 0
	}

	if height != nil {
		*height = 0
	}

	xPos := 0
	tmpList := private.runs
	for tmpList != nil {
		run := tmpList.data
		var (
			runInk, runLogical Rectangle
			newPos, runHeight  int
		)
		run.pango_layout_run_get_extents_and_height(&runInk, &runLogical, &runHeight)

		if inkRect != nil {
			if inkRect.width == 0 || inkRect.height == 0 {
				*inkRect = runInk
				inkRect.x += xPos
			} else if runInk.width != 0 && runInk.height != 0 {
				newPos = min(inkRect.x, xPos+runInk.x)
				inkRect.width = max(inkRect.x+inkRect.width, xPos+runInk.x+runInk.width) - newPos
				inkRect.x = newPos

				newPos = min(inkRect.y, runInk.y)
				inkRect.height = max(inkRect.y+inkRect.height, runInk.y+runInk.height) - newPos
				inkRect.y = newPos
			}
		}

		if logicalRect != nil {
			newPos = min(logicalRect.x, xPos+runLogical.x)
			logicalRect.width = max(logicalRect.x+logicalRect.width, xPos+runLogical.x+runLogical.width) - newPos
			logicalRect.x = newPos

			newPos = min(logicalRect.y, runLogical.y)
			logicalRect.height = max(logicalRect.y+logicalRect.height, runLogical.y+runLogical.height) - newPos
			logicalRect.y = newPos
		}

		if height != nil {
			*height = max(*height, runHeight)
		}

		xPos += runLogical.width
		tmpList = tmpList.next
	}

	if logicalRect != nil && private.runs == nil {
		private.pango_layout_line_get_empty_extents(logicalRect)
	}

	if caching {
		if &private.inkRect != inkRect {
			private.inkRect = *inkRect
		}
		if &private.logicalRect != logicalRect {
			private.logicalRect = *logicalRect
		}
		if &private.height != height {
			private.height = *height
		}
		private.cache_status = CACHED
	}
}

func (line *LayoutLine) pango_layout_line_get_empty_extents(logicalRect *Rectangle) {
	line.layout.pango_layout_get_empty_extents_at_index(line.start_index, logicalRect)
}
