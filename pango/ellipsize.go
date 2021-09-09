package pango

import (
	"math"

	"github.com/benoitkugler/textlayout/fribidi"
)

/* Overall, the way we ellipsize is we grow a "gap" out from an original
 * gap center position until:
 *
 *  line_width - gap_width + ellipsize_width <= goalWidth
 *
 * Line:  [-------------------------------------------]
 * Runs:  [------)[---------------)[------------------]
 * Gap center:                 *
 * Gap:             [----------------------]
 *
 * The gap center may be at the start or end in which case the gap grows
 * in only one direction.
 *
 * Note the line and last run are logically closed at the end; this allows
 * us to use a gap position at x=line_width and still have it be part of
 * of a run.
 *
 * We grow the gap out one "span" at a time, where a span is simply a
 * consecutive run of clusters that we can't interrupt with an ellipsis.
 *
 * When choosing whether to grow the gap at the start or the end, we
 * calculate the next span to remove in both directions and see which
 * causes the smaller increase in:
 *
 *  MAX (gap_end - gap_center, gap_start - gap_center)
 *
 * All computations are done using logical order; the ellipsization
 * process occurs before the runs are ordered into visual order.
 */

// EllipsizeMode describes what sort of (if any)
// ellipsization should be applied to a line of text. In
// the ellipsization process characters are removed from the
// text in order to make it fit to a given width and replaced
// with an ellipsis.
type EllipsizeMode uint8

const (
	ELLIPSIZE_NONE   EllipsizeMode = iota // No ellipsization
	ELLIPSIZE_START                       // Omit characters at the start of the text
	ELLIPSIZE_MIDDLE                      // Omit characters in the middle of the text
	ELLIPSIZE_END                         // Omit characters at the end of the text
)

// keeps information about a single run
type runInfo struct {
	run         *GlyphItem
	startOffset int       // Character offset of run start
	width       GlyphUnit // Width of run in Pango units
}

// iterator to a position within the ellipsized line
type lineIter struct {
	runIter  GlyphItemIter
	runIndex int
}

// state of ellipsization process
type ellipsizeState struct {
	ellipsisRun *GlyphItem // Run created to hold ellipsis
	layout      *Layout    // Layout being ellipsized
	attrs       AttrList   // Attributes used for itemization/shaping

	runInfo []runInfo // Array of information about each run, of size `n_runs`

	lineStartAttr *attrIterator // Cached AttrIterator for the start of the run

	gapStartAttr *attrIterator // Attribute iterator pointing to a range containing the first character in gap

	gapStartIter lineIter // Iteratator pointig to the first cluster in gap
	gapEndIter   lineIter // Iterator pointing to last cluster in gap

	gapStartX GlyphUnit // x position of start of gap, in Pango units
	gapEndX   GlyphUnit // x position of end of gap, in Pango units

	ellipsisWidth GlyphUnit // Width of ellipsis, in Pango units
	totalWidth    GlyphUnit // Original width of line in Pango units
	gapCenter     GlyphUnit // Goal for center of gap

	shapeFlags shapeFlags

	// Whether the first character in the ellipsized
	// is wide; this triggers us to try to use a
	// mid-line ellipsis instead of a baseline
	ellipsisIsCJK bool
}

// Compute global information needed for the itemization process

func (line *LayoutLine) newState(attrs AttrList, sf shapeFlags) ellipsizeState {
	var state ellipsizeState

	state.layout = line.layout

	state.attrs = attrs
	state.shapeFlags = sf

	state.runInfo = make([]runInfo, line.Runs.length())

	startOffset := line.StartIndex
	for l, i := line.Runs, 0; l != nil; l, i = l.Next, i+1 {
		run := l.Data
		width := run.Glyphs.getWidth()
		state.runInfo[i].run = run
		state.runInfo[i].width = width
		state.runInfo[i].startOffset = startOffset
		state.totalWidth += width
		startOffset += run.Item.Length
	}

	return state
}

// computes the width of a single cluster
func (iter lineIter) getClusterWidth() GlyphUnit {
	runIter := iter.runIter
	glyphs := runIter.glyphItem.Glyphs

	var width GlyphUnit
	if runIter.startGlyph < runIter.endGlyph { // LTR
		for i := runIter.startGlyph; i < runIter.endGlyph; i++ {
			width += glyphs.Glyphs[i].Geometry.Width
		}
	} else { // RTL
		for i := runIter.startGlyph; i > runIter.endGlyph; i-- {
			width += glyphs.Glyphs[i].Geometry.Width
		}
	}

	return width
}

// move forward one cluster. Returns `false` if we were already at the end
func (state *ellipsizeState) lineIterNextCluster(iter *lineIter) bool {
	if !iter.runIter.NextCluster() {
		if iter.runIndex == len(state.runInfo)-1 {
			return false
		} else {
			iter.runIndex++
			iter.runIter.InitStart(state.runInfo[iter.runIndex].run, state.layout.Text)
		}
	}
	return true
}

// move backward one cluster. Returns `false` if we were already at the end
func (state *ellipsizeState) lineIterPrevCluster(iter *lineIter) bool {
	if !iter.runIter.PrevCluster() {
		if iter.runIndex == 0 {
			return false
		} else {
			iter.runIndex--
			iter.runIter.InitEnd(state.runInfo[iter.runIndex].run, state.layout.Text)
		}
	}
	return true
}

// An ellipsization boundary is defined by two things
//
// - Starts a cluster - forced by structure of code
// - Starts a grapheme - checked here
//
// In the future we'd also like to add a check for cursive connectivity here.
// This should be an addition to GlyphVisAttr
//

// checks if there is a ellipsization boundary before the cluster `iter` points to
func (state ellipsizeState) startsAtEllipsizationBoundary(iter lineIter) bool {
	runInfo := state.runInfo[iter.runIndex]

	if iter.runIter.StartChar == 0 && iter.runIndex == 0 {
		return true
	}

	return state.layout.logAttrs[runInfo.startOffset+iter.runIter.StartChar].IsCursorPosition()
}

// checks if there is a ellipsization boundary after the cluster `iter` points to
func (state ellipsizeState) endsAtEllipsizationBoundary(iter lineIter) bool {
	runInfo := state.runInfo[iter.runIndex]

	if iter.runIter.EndChar == runInfo.run.Item.Length && iter.runIndex == len(state.runInfo)-1 {
		return true
	}

	return state.layout.logAttrs[runInfo.startOffset+iter.runIter.EndChar+1].IsCursorPosition()
}

// helper function to re-itemize a string of text
func (state *ellipsizeState) itemizeText(text []rune, attrs AttrList) *Item {
	items := state.layout.context.Itemize(text, 0, len(text), attrs)

	if debugMode {
		assert(items != nil && items.Next == nil, "itemizeText")
	}
	return items.Data
}

// shapes the ellipsis using the font and is_cjk information computed by
// updateEllipsisShape() from the first character in the gap.
func (state *ellipsizeState) shapeEllipsis() {
	var attrs AttrList
	// Create/reset state.ellipsis_run
	if state.ellipsisRun == nil {
		state.ellipsisRun = new(GlyphItem)
		state.ellipsisRun.Glyphs = new(GlyphString)
	}

	if state.ellipsisRun.Item != nil {
		state.ellipsisRun.Item = nil
	}

	// Create an attribute list
	runAttrs := state.gapStartAttr.getAttributes()
	for _, attr := range runAttrs {
		attr.StartIndex = 0
		attr.EndIndex = MaxInt
		attrs.insert(attr)
	}

	fallback := NewAttrFallback(false)
	attrs.insert(fallback)

	// First try using a specific ellipsis character in the best matching font
	var ellipsisText []rune
	if state.ellipsisIsCJK {
		ellipsisText = []rune{'\u22EF'} // U+22EF: MIDLINE HORIZONTAL ELLIPSIS, used for CJK
	} else {
		ellipsisText = []rune{'\u2026'} // U+2026: HORIZONTAL ELLIPSIS
	}

	item := state.itemizeText(ellipsisText, attrs)

	// If that fails we use "..." in the first matching font
	if item.Analysis.Font == nil || !fontHasChar(item.Analysis.Font, ellipsisText[0]) {
		// Modify the fallback iter for it is inside the AttrList; Don't try this at home
		fallback.Data = AttrInt(1)
		ellipsisText = []rune("...")
		item = state.itemizeText(ellipsisText, attrs)
	}

	state.ellipsisRun.Item = item

	// Now shape
	glyphs := state.ellipsisRun.Glyphs
	glyphs.shapeWithFlags(ellipsisText, 0, len(ellipsisText), &item.Analysis, state.shapeFlags)

	state.ellipsisWidth = 0
	for _, g := range glyphs.Glyphs {
		state.ellipsisWidth += g.Geometry.Width
	}
}

// helper function to advance a AttrIterator to a particular rune index.
func (iter *attrIterator) advanceTo(newIndex int) {
	for do := true; do; do = iter.next() {
		if iter.EndIndex > newIndex {
			break
		}
	}
}

// updates the shaping of the ellipsis if necessary when we move the
// position of the start of the gap.
//
// The shaping of the ellipsis is determined by two things:
// - The font attributes applied to the first character in the gap
// - Whether the first character in the gap is wide or not. If the
//   first character is wide, then we assume that we are ellipsizing
//   East-Asian text, so prefer a mid-line ellipsizes to a baseline
//   ellipsis, since that's typical practice for Chinese/Japanese/Korean.
func (state *ellipsizeState) updateEllipsisShape() {
	recompute := false

	// Unfortunately, we can only advance AttrIterator forward; so each
	// time we back up we need to go forward to find the new position. To make
	// this not utterly slow, we cache an iterator at the start of the line
	if state.lineStartAttr == nil {
		state.lineStartAttr = state.attrs.getIterator()
		state.lineStartAttr.advanceTo(state.runInfo[0].run.Item.Offset)
	}

	if state.gapStartAttr != nil {
		// See if the current attribute range contains the new start position
		start, _ := state.gapStartAttr.StartIndex, state.gapStartAttr.EndIndex
		if state.gapStartIter.runIter.StartIndex < start {
			state.gapStartAttr = nil
		}
	}

	// Check whether we need to recompute the ellipsis because of new font attributes
	if state.gapStartAttr == nil {
		state.gapStartAttr = state.lineStartAttr.copy()
		state.gapStartAttr.advanceTo(state.runInfo[state.gapStartIter.runIndex].run.Item.Offset)
		recompute = true
	}

	// Check whether we need to recompute the ellipsis because we switch from CJK to not or vice-versa
	startWc := state.layout.Text[state.gapStartIter.runIter.StartIndex]
	isCJK := isWide(startWc)

	if isCJK != state.ellipsisIsCJK {
		state.ellipsisIsCJK = isCJK
		recompute = true
	}

	if recompute {
		state.shapeEllipsis()
	}
}

// computes the position of the gap center and finds the smallest span containing it
func (state *ellipsizeState) findInitialSpan() {
	switch state.layout.ellipsize {
	case ELLIPSIZE_START:
		state.gapCenter = 0
	case ELLIPSIZE_MIDDLE:
		state.gapCenter = state.totalWidth / 2
	case ELLIPSIZE_END:
		state.gapCenter = state.totalWidth
	}

	// Find the run containing the gap center

	var (
		x GlyphUnit
		i int
	)
	for ; i < len(state.runInfo); i++ {
		if x+state.runInfo[i].width > state.gapCenter {
			break
		}

		x += state.runInfo[i].width
	}

	if i == len(state.runInfo) {
		// Last run is a closed interval, so back off one run
		i--
		x -= state.runInfo[i].width
	}

	// Find the cluster containing the gap center

	state.gapStartIter.runIndex = i
	runIter := &state.gapStartIter.runIter
	glyphItem := state.runInfo[i].run

	var clusterWidth GlyphUnit
	haveCluster := runIter.InitStart(glyphItem, state.layout.Text)
	for ; haveCluster; haveCluster = runIter.NextCluster() {
		clusterWidth = state.gapStartIter.getClusterWidth()

		if x+clusterWidth > state.gapCenter {
			break
		}

		x += clusterWidth
	}

	if !haveCluster {
		// Last cluster is a closed interval, so back off one cluster
		x -= clusterWidth
	}

	state.gapEndIter = state.gapStartIter

	state.gapStartX = x
	state.gapEndX = x + clusterWidth

	// Expand the gap to a full span

	for !state.startsAtEllipsizationBoundary(state.gapStartIter) {
		state.lineIterPrevCluster(&state.gapStartIter)
		state.gapStartX -= state.gapStartIter.getClusterWidth()
	}

	for !state.endsAtEllipsizationBoundary(state.gapEndIter) {
		state.lineIterNextCluster(&state.gapEndIter)
		state.gapEndX += state.gapEndIter.getClusterWidth()
	}

	state.updateEllipsisShape()
}

// Removes one run from the start or end of the gap. Returns false
// if there's nothing left to remove in either direction.
func (state *ellipsizeState) removeOneSpan() bool {
	// Find one span backwards and forward from the gap
	new_gap_start_iter := state.gapStartIter
	new_gap_start_x := state.gapStartX
	var width GlyphUnit
	for do := true; do; do = !state.startsAtEllipsizationBoundary(new_gap_start_iter) || width == 0 {
		if !state.lineIterPrevCluster(&new_gap_start_iter) {
			break
		}
		width = new_gap_start_iter.getClusterWidth()
		new_gap_start_x -= width
	}

	new_gap_end_iter := state.gapEndIter
	new_gap_end_x := state.gapEndX
	for do := true; do; do = !state.endsAtEllipsizationBoundary(new_gap_end_iter) || width == 0 {
		if !state.lineIterNextCluster(&new_gap_end_iter) {
			break
		}
		width = new_gap_end_iter.getClusterWidth()
		new_gap_end_x += width
	}

	if state.gapEndX == new_gap_end_x && state.gapStartX == new_gap_start_x {
		return false
	}

	// In the case where we could remove a span from either end of the
	// gap, we look at which causes the smaller increase in the
	// MAX (gap_end - gap_center, gap_start - gap_center)
	if state.gapEndX == new_gap_end_x ||
		(state.gapStartX != new_gap_start_x &&
			state.gapCenter-new_gap_start_x < new_gap_end_x-state.gapCenter) {
		state.gapStartIter = new_gap_start_iter
		state.gapStartX = new_gap_start_x

		state.updateEllipsisShape()
	} else {
		state.gapEndIter = new_gap_end_iter
		state.gapEndX = new_gap_end_x
	}

	return true
}

// Fixes up the properties of the ellipsis run once we've determined the final extents of the gap
func (state *ellipsizeState) fixupEllipsisRun(extraWidth GlyphUnit) {
	glyphs := state.ellipsisRun.Glyphs
	item := state.ellipsisRun.Item

	// Make the entire glyphstring into a single logical cluster
	for i := range glyphs.Glyphs {
		glyphs.logClusters[i] = 0
		glyphs.Glyphs[i].attr.isClusterStart = false
	}

	glyphs.Glyphs[0].attr.isClusterStart = true
	glyphs.Glyphs[len(glyphs.Glyphs)-1].Geometry.Width += extraWidth

	// Fix up the item to point to the entire elided text
	item.Offset = state.gapStartIter.runIter.StartIndex
	item.Length = state.gapEndIter.runIter.EndIndex - item.Offset

	// The level for the item is the minimum level of the elided text
	var level fribidi.Level = math.MaxInt8
	for _, rf := range state.runInfo[state.gapStartIter.runIndex : state.gapEndIter.runIndex+1] {
		level = minL(level, rf.run.Item.Analysis.Level)
	}

	item.Analysis.Level = level

	item.Analysis.Flags |= AFIsEllipsis
}

// Computes the new list of runs for the line
func (state *ellipsizeState) getRunList() *RunList {
	var partialStartRun, partialEndRun *GlyphItem
	// We first cut out the pieces of the starting and ending runs we want to
	// preserve; we do the end first in case the end and the start are
	// the same. Doing the start first would disturb the indices for the end.
	runInfo := &state.runInfo[state.gapEndIter.runIndex]
	runIter := &state.gapEndIter.runIter
	if runIter.EndChar != runInfo.run.Item.Length {
		partialEndRun = runInfo.run
		runInfo.run = runInfo.run.split(state.layout.Text, runIter.EndIndex-runInfo.run.Item.Offset)
	}

	runInfo = &state.runInfo[state.gapStartIter.runIndex]
	runIter = &state.gapStartIter.runIter
	if runIter.StartChar != 0 {
		partialStartRun = runInfo.run.split(state.layout.Text, runIter.StartIndex-runInfo.run.Item.Offset)
	}

	// Now assemble the new list of runs
	var result *RunList
	for _, rf := range state.runInfo[0:state.gapStartIter.runIndex] {
		result = &RunList{Data: rf.run, Next: result}
	}

	if partialStartRun != nil {
		result = &RunList{Data: partialStartRun, Next: result}
	}

	result = &RunList{Data: state.ellipsisRun, Next: result}

	if partialEndRun != nil {
		result = &RunList{Data: partialEndRun, Next: result}
	}

	for _, rf := range state.runInfo[state.gapEndIter.runIndex+1:] {
		result = &RunList{Data: rf.run, Next: result}
	}

	return result.reverse()
}

// computes the width of the line as currently ellipsized
func (state *ellipsizeState) currentWidth() GlyphUnit {
	return state.totalWidth - (state.gapEndX - state.gapStartX) + state.ellipsisWidth
}

// ellipsize ellipsizes a `LayoutLine`, with the runs still in logical order,
// and according to the layout's policy to fit within the set width of the layout.
// It returns whether the line had to be ellipsized
func (line *LayoutLine) ellipsize(attrs AttrList, shapeFlag shapeFlags, goalWidth GlyphUnit) bool {
	if line.layout.ellipsize == ELLIPSIZE_NONE || goalWidth < 0 {
		return false
	}

	state := line.newState(attrs, shapeFlag)
	if state.totalWidth <= goalWidth {
		return false
	}

	state.findInitialSpan()

	for state.currentWidth() > goalWidth {
		if !state.removeOneSpan() {
			break
		}
	}

	state.fixupEllipsisRun(maxG(goalWidth-state.currentWidth(), 0))

	line.Runs = state.getRunList()
	return true
}
