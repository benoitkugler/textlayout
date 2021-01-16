package pango

import (
	"math"

	"github.com/benoitkugler/go-weasyprint/fribidi"
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
	PANGO_ELLIPSIZE_NONE   EllipsizeMode = iota // No ellipsization
	PANGO_ELLIPSIZE_START                       // Omit characters at the start of the text
	PANGO_ELLIPSIZE_MIDDLE                      // Omit characters in the middle of the text
	PANGO_ELLIPSIZE_END                         // Omit characters at the end of the text
)

// keeps information about a single run
type RunInfo struct {
	run          *GlyphItem
	start_offset int       // Character offset of run start
	width        GlyphUnit // Width of run in Pango units
}

// iterator to a position within the ellipsized line
type LineIter struct {
	run_iter  GlyphItemIter
	run_index int
}

// state of ellipsization process
type EllipsizeState struct {
	layout *Layout  // Layout being ellipsized
	attrs  AttrList // Attributes used for itemization/shaping

	run_info []RunInfo // Array of information about each run, of size `n_runs`
	// n_runs   int

	total_width GlyphUnit // Original width of line in Pango units
	gap_center  GlyphUnit // Goal for center of gap

	ellipsis_run   *GlyphItem // Run created to hold ellipsis
	ellipsis_width GlyphUnit  // Width of ellipsis, in Pango units

	// Whether the first character in the ellipsized
	// is wide; this triggers us to try to use a
	// mid-line ellipsis instead of a baseline
	ellipsis_is_cjk bool

	line_start_attr *AttrIterator // Cached AttrIterator for the start of the run

	gap_start_iter LineIter      // Iteratator pointig to the first cluster in gap
	gap_start_x    GlyphUnit     // x position of start of gap, in Pango units
	gap_start_attr *AttrIterator // Attribute iterator pointing to a range containing the first character in gap

	gap_end_iter LineIter  // Iterator pointing to last cluster in gap
	gap_end_x    GlyphUnit // x position of end of gap, in Pango units

	shape_flags ShapeFlags
}

// Compute global information needed for the itemization process

func (line *LayoutLine) newState(attrs AttrList, shape_flags ShapeFlags) EllipsizeState {
	var state EllipsizeState

	state.layout = line.layout

	state.attrs = attrs
	state.shape_flags = shape_flags

	state.run_info = make([]RunInfo, line.runs.length())

	start_offset := line.start_index
	for l, i := line.runs, 0; l != nil; l, i = l.next, i+1 {
		run := l.data
		width := run.glyphs.pango_glyph_string_get_width()
		state.run_info[i].run = run
		state.run_info[i].width = width
		state.run_info[i].start_offset = start_offset
		state.total_width += width
		start_offset += run.item.num_chars
	}

	return state
}

//  // Cleanup memory allocation

// func free_state (state *EllipsizeState)
//  {
//    pango_attr_list_unref (state.attrs);
//    if (state.line_start_attr)
// 	 pango_attr_iterator_destroy (state.line_start_attr);
//    if (state.gap_start_attr)
// 	 pango_attr_iterator_destroy (state.gap_start_attr);
//    g_free (state.run_info);
//  }

// computes the width of a single cluster
func (iter LineIter) getClusterWidth() GlyphUnit {
	run_iter := iter.run_iter
	glyphs := run_iter.glyphItem.glyphs

	var width GlyphUnit
	if run_iter.start_glyph < run_iter.end_glyph { // LTR
		for i := run_iter.start_glyph; i < run_iter.end_glyph; i++ {
			width += glyphs.glyphs[i].geometry.width
		}
	} else { // RTL
		for i := run_iter.start_glyph; i > run_iter.end_glyph; i-- {
			width += glyphs.glyphs[i].geometry.width
		}
	}

	return width
}

// move forward one cluster. Returns `false` if we were already at the end
func (state *EllipsizeState) lineIterNextCluster(iter *LineIter) bool {
	if !iter.run_iter.pango_glyph_item_iter_next_cluster() {
		if iter.run_index == len(state.run_info)-1 {
			return false
		} else {
			iter.run_index++
			iter.run_iter.pango_glyph_item_iter_init_start(state.run_info[iter.run_index].run, state.layout.text)
		}
	}
	return true
}

// move backward one cluster. Returns `false` if we were already at the end
func (state *EllipsizeState) lineIterPrevCluster(iter *LineIter) bool {
	if !iter.run_iter.pango_glyph_item_iter_prev_cluster() {
		if iter.run_index == 0 {
			return false
		} else {
			iter.run_index--
			iter.run_iter.pango_glyph_item_iter_init_end(state.run_info[iter.run_index].run, state.layout.text)
		}
	}
	return true
}

//  //
//   * An ellipsization boundary is defined by two things
//   *
//   * - Starts a cluster - forced by structure of code
//   * - Starts a grapheme - checked here
//   *
//   * In the future we'd also like to add a check for cursive connectivity here.
//   * This should be an addition to #PangoGlyphVisAttr
//   *

// checks if there is a ellipsization boundary before the cluster `iter` points to
func (state EllipsizeState) startsAtEllipsizationBoundary(iter LineIter) bool {
	runInfo := state.run_info[iter.run_index]

	if iter.run_iter.start_char == 0 && iter.run_index == 0 {
		return true
	}

	return state.layout.log_attrs[runInfo.start_offset+iter.run_iter.start_char].IsCursorPosition()
}

// checks if there is a ellipsization boundary after the cluster `iter` points to
func (state EllipsizeState) endsAtEllipsizationBoundary(iter LineIter) bool {
	run_info := state.run_info[iter.run_index]

	if iter.run_iter.end_char == run_info.run.item.num_chars && iter.run_index == len(state.run_info)-1 {
		return true
	}

	return state.layout.log_attrs[run_info.start_offset+iter.run_iter.end_char+1].IsCursorPosition()
}

// helper function to re-itemize a string of text
func (state *EllipsizeState) itemizeText(text []rune, attrs AttrList) *Item {
	items := state.layout.context.pango_itemize(text, 0, len(text), attrs, nil)

	if debugMode {
		assert(items != nil && items.next == nil)
	}
	return items.data
}

// shapes the ellipsis using the font and is_cjk information computed by
// updateEllipsisShape() from the first character in the gap.
func (state *EllipsizeState) shapeEllipsis() {
	var attrs AttrList
	// Create/reset state.ellipsis_run
	if state.ellipsis_run == nil {
		state.ellipsis_run = new(GlyphItem)
	}

	if state.ellipsis_run.item != nil {
		state.ellipsis_run.item = nil
	}

	// Create an attribute list
	run_attrs := state.gap_start_attr.pango_attr_iterator_get_attrs()
	for _, attr := range run_attrs {
		attr.StartIndex = 0
		attr.EndIndex = maxInt
		attrs.pango_attr_list_insert(attr)
	}

	fallback := pango_attr_fallback_new(false)
	attrs.pango_attr_list_insert(fallback)

	// First try using a specific ellipsis character in the best matching font
	var ellipsis_text []rune
	if state.ellipsis_is_cjk {
		ellipsis_text = []rune{'\u22EF'} // U+22EF: MIDLINE HORIZONTAL ELLIPSIS, used for CJK
	} else {
		ellipsis_text = []rune{'\u2026'} // U+2026: HORIZONTAL ELLIPSIS
	}

	item := state.itemizeText(ellipsis_text, attrs)

	// If that fails we use "..." in the first matching font
	if item.analysis.font == nil || !pango_font_has_char(item.analysis.font, ellipsis_text[0]) {
		// Modify the fallback iter for it is inside the AttrList; Don't try this at home
		fallback.Data = AttrInt(1)
		ellipsis_text = []rune("...")
		item = state.itemizeText(ellipsis_text, attrs)
	}

	state.ellipsis_run.item = item

	// Now shape
	glyphs := state.ellipsis_run.glyphs
	glyphs.pango_shape_with_flags(ellipsis_text, ellipsis_text, &item.analysis, state.shape_flags)

	state.ellipsis_width = 0
	for _, g := range glyphs.glyphs {
		state.ellipsis_width += g.geometry.width
	}
}

// helper function to advance a AttrIterator to a particular rune index.
func advanceIteratorTo(iter *AttrIterator, newIndex int) {
	for do := true; do; do = iter.pango_attr_iterator_next() {
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
func (state *EllipsizeState) updateEllipsisShape() {
	recompute := false

	// Unfortunately, we can only advance AttrIterator forward; so each
	// time we back up we need to go forward to find the new position. To make
	// this not utterly slow, we cache an iterator at the start of the line
	if state.line_start_attr == nil {
		state.line_start_attr = state.attrs.pango_attr_list_get_iterator()
		advanceIteratorTo(state.line_start_attr, state.run_info[0].run.item.offset)
	}

	if state.gap_start_attr != nil {
		// See if the current attribute range contains the new start position
		start, _ := state.gap_start_attr.StartIndex, state.gap_start_attr.EndIndex
		if state.gap_start_iter.run_iter.start_index < start {
			state.gap_start_attr = nil
		}
	}

	// Check whether we need to recompute the ellipsis because of new font attributes
	if state.gap_start_attr == nil {
		state.gap_start_attr = state.line_start_attr.pango_attr_iterator_copy()
		advanceIteratorTo(state.gap_start_attr, state.run_info[state.gap_start_iter.run_index].run.item.offset)
		recompute = true
	}

	// Check whether we need to recompute the ellipsis because we switch from CJK to not or vice-versa
	start_wc := state.layout.text[state.gap_start_iter.run_iter.start_index]
	is_cjk := isWide(start_wc)

	if is_cjk != state.ellipsis_is_cjk {
		state.ellipsis_is_cjk = is_cjk
		recompute = true
	}

	if recompute {
		state.shapeEllipsis()
	}
}

// computes the position of the gap center and finds the smallest span containing it
func (state *EllipsizeState) findInitialSpan() {
	switch state.layout.ellipsize {
	case PANGO_ELLIPSIZE_START:
		state.gap_center = 0
	case PANGO_ELLIPSIZE_MIDDLE:
		state.gap_center = state.total_width / 2
	case PANGO_ELLIPSIZE_END:
		state.gap_center = state.total_width
	}

	// Find the run containing the gap center

	var (
		x GlyphUnit
		i int
	)
	for ; i < len(state.run_info); i++ {
		if x+state.run_info[i].width > state.gap_center {
			break
		}

		x += state.run_info[i].width
	}

	if i == len(state.run_info) {
		// Last run is a closed interval, so back off one run
		i--
		x -= state.run_info[i].width
	}

	// Find the cluster containing the gap center

	state.gap_start_iter.run_index = i
	run_iter := &state.gap_start_iter.run_iter
	glyph_item := state.run_info[i].run

	var cluster_width GlyphUnit // Quiet GCC, the line must have at least one cluster
	have_cluster := run_iter.pango_glyph_item_iter_init_start(glyph_item, state.layout.text)
	for ; have_cluster; have_cluster = run_iter.pango_glyph_item_iter_next_cluster() {
		cluster_width = state.gap_start_iter.getClusterWidth()

		if x+cluster_width > state.gap_center {
			break
		}

		x += cluster_width
	}

	if !have_cluster {
		// Last cluster is a closed interval, so back off one cluster
		x -= cluster_width
	}

	state.gap_end_iter = state.gap_start_iter

	state.gap_start_x = x
	state.gap_end_x = x + cluster_width

	// Expand the gap to a full span

	for !state.startsAtEllipsizationBoundary(state.gap_start_iter) {
		state.lineIterPrevCluster(&state.gap_start_iter)
		state.gap_start_x -= state.gap_start_iter.getClusterWidth()
	}

	for !state.endsAtEllipsizationBoundary(state.gap_end_iter) {
		state.lineIterNextCluster(&state.gap_end_iter)
		state.gap_end_x += state.gap_end_iter.getClusterWidth()
	}

	state.updateEllipsisShape()
}

// Removes one run from the start or end of the gap. Returns false
// if there's nothing left to remove in either direction.
func (state *EllipsizeState) removeOneSpan() bool {
	// Find one span backwards and forward from the gap
	new_gap_start_iter := state.gap_start_iter
	new_gap_start_x := state.gap_start_x
	var width GlyphUnit
	for do := true; do; do = !state.startsAtEllipsizationBoundary(new_gap_start_iter) || width == 0 {
		if !state.lineIterPrevCluster(&new_gap_start_iter) {
			break
		}
		width = new_gap_start_iter.getClusterWidth()
		new_gap_start_x -= width
	}

	new_gap_end_iter := state.gap_end_iter
	new_gap_end_x := state.gap_end_x
	for do := true; do; do = !state.endsAtEllipsizationBoundary(new_gap_end_iter) || width == 0 {
		if !state.lineIterNextCluster(&new_gap_end_iter) {
			break
		}
		width = new_gap_end_iter.getClusterWidth()
		new_gap_end_x += width
	}

	if state.gap_end_x == new_gap_end_x && state.gap_start_x == new_gap_start_x {
		return false
	}

	// In the case where we could remove a span from either end of the
	// gap, we look at which causes the smaller increase in the
	// MAX (gap_end - gap_center, gap_start - gap_center)
	if state.gap_end_x == new_gap_end_x ||
		(state.gap_start_x != new_gap_start_x &&
			state.gap_center-new_gap_start_x < new_gap_end_x-state.gap_center) {
		state.gap_start_iter = new_gap_start_iter
		state.gap_start_x = new_gap_start_x

		state.updateEllipsisShape()
	} else {
		state.gap_end_iter = new_gap_end_iter
		state.gap_end_x = new_gap_end_x
	}

	return true
}

// Fixes up the properties of the ellipsis run once we've determined the final extents of the gap
func (state *EllipsizeState) fixupEllipsisRun() {
	glyphs := state.ellipsis_run.glyphs
	item := state.ellipsis_run.item

	// Make the entire glyphstring into a single logical cluster
	for i := range glyphs.glyphs {
		glyphs.log_clusters[i] = 0
		glyphs.glyphs[i].attr.is_cluster_start = false
	}

	glyphs.glyphs[0].attr.is_cluster_start = true

	// Fix up the item to point to the entire elided text
	item.offset = state.gap_start_iter.run_iter.start_index
	item.num_chars = state.gap_end_iter.run_iter.end_index - item.offset

	// The level for the item is the minimum level of the elided text
	var level fribidi.Level = math.MaxInt8
	for _, rf := range state.run_info[state.gap_start_iter.run_index : state.gap_end_iter.run_index+1] {
		level = minL(level, rf.run.item.analysis.level)
	}

	item.analysis.level = level

	item.analysis.flags |= PANGO_ANALYSIS_FLAG_IS_ELLIPSIS
}

// Computes the new list of runs for the line
func (state *EllipsizeState) getRunList() *runList {
	var partial_start_run, partial_end_run *GlyphItem
	// We first cut out the pieces of the starting and ending runs we want to
	// preserve; we do the end first in case the end and the start are
	// the same. Doing the start first would disturb the indices for the end.
	run_info := &state.run_info[state.gap_end_iter.run_index]
	run_iter := &state.gap_end_iter.run_iter
	if run_iter.end_char != run_info.run.item.num_chars {
		partial_end_run = run_info.run
		run_info.run = run_info.run.pango_glyph_item_split(state.layout.text, run_iter.end_index-run_info.run.item.offset)
	}

	run_info = &state.run_info[state.gap_start_iter.run_index]
	run_iter = &state.gap_start_iter.run_iter
	if run_iter.start_char != 0 {
		partial_start_run = run_info.run.pango_glyph_item_split(state.layout.text, run_iter.start_index-run_info.run.item.offset)
	}

	// Now assemble the new list of runs
	var result *runList
	for _, rf := range state.run_info[0:state.gap_start_iter.run_index] {
		result = &runList{data: rf.run, next: result}
	}

	if partial_start_run != nil {
		result = &runList{data: partial_start_run, next: result}
	}

	result = &runList{data: state.ellipsis_run, next: result}

	if partial_end_run != nil {
		result = &runList{data: partial_end_run, next: result}
	}

	for _, rf := range state.run_info[state.gap_end_iter.run_index+1:] {
		result = &runList{data: rf.run, next: result}
	}

	return result.reverse()
}

// computes the width of the line as currently ellipsized
func (state *EllipsizeState) currentWidth() GlyphUnit {
	return state.total_width - (state.gap_end_x - state.gap_start_x) + state.ellipsis_width
}

// _pango_layout_line_ellipsize ellipsizes a `LayoutLine`, with the runs still in logical order,
// and according to the layout's policy to fit within the set width of the layout.
// It returns whether the line had to be ellipsized
func (line *LayoutLine) _pango_layout_line_ellipsize(attrs AttrList, shapeFlags ShapeFlags, goalWidth GlyphUnit) bool {
	if line.layout.ellipsize == PANGO_ELLIPSIZE_NONE || goalWidth < 0 {
		return false
	}

	state := line.newState(attrs, shapeFlags)

	if state.total_width <= goalWidth {
		return false
	}

	state.findInitialSpan()

	for state.currentWidth() > goalWidth {
		if !state.removeOneSpan() {
			break
		}
	}

	state.fixupEllipsisRun()

	line.runs = state.getRunList()
	return true

}
