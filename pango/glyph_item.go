package pango

// GlyphItem is a pair of a Item and the glyphs
// resulting from shaping the text corresponding to an item.
// As an example of the usage of GlyphItem, the results
// of shaping text with Layout is a list of LayoutLine,
// each of which contains a list of GlyphItem.
type GlyphItem struct {
	item   *Item
	glyphs *GlyphString
}

func (g GlyphItem) LTR() bool {
	return g.item.analysis.level%2 == 0
}

// pango_glyph_item_split modifies `orig` to cover only the text after `splitIndex`, and
// returns a new item that covers the text before `splitIndex` that
// used to be in `orig`. You can think of `splitIndex` as the length of
// the returned item. `splitIndex` may not be 0, and it may not be
// greater than or equal to the length of `orig` (that is, there must
// be at least one rune assigned to each item, you can't create a
// zero-length item).
//
// This function is similar in function to pango_item_split() (and uses
// it internally.)
func (orig *GlyphItem) pango_glyph_item_split(text []rune, splitIndex int) *GlyphItem {
	if orig.item.num_chars <= 0 || splitIndex <= 0 || splitIndex >= orig.item.num_chars {
		return nil
	}

	var i, numGlyphs int
	if orig.LTR() {
		for i = 0; i < len(orig.glyphs.log_clusters); i++ {
			if orig.glyphs.log_clusters[i] >= splitIndex {
				break
			}
		}

		if i == len(orig.glyphs.glyphs) {
			/* No splitting necessary */
			return nil
		}

		splitIndex = orig.glyphs.log_clusters[i]
		numGlyphs = i
	} else {
		for i = len(orig.glyphs.glyphs) - 1; i >= 0; i-- {
			if orig.glyphs.log_clusters[i] >= splitIndex {
				break
			}
		}

		if i < 0 {
			/* No splitting necessary */
			return nil
		}

		splitIndex = orig.glyphs.log_clusters[i]
		numGlyphs = len(orig.glyphs.glyphs) - 1 - i
	}

	var new GlyphItem
	new.item = orig.item.pango_item_split(splitIndex)
	new.glyphs = &GlyphString{}
	new.glyphs.pango_glyph_string_set_size(numGlyphs)

	numRemaining := len(orig.glyphs.glyphs) - numGlyphs
	if orig.LTR() {
		copy(new.glyphs.glyphs, orig.glyphs.glyphs[:numGlyphs])
		copy(new.glyphs.log_clusters, orig.glyphs.log_clusters[:numGlyphs])

		copy(orig.glyphs.glyphs, orig.glyphs.glyphs[numGlyphs:])
		for i = numGlyphs; i < len(orig.glyphs.glyphs); i++ {
			orig.glyphs.log_clusters[i-numGlyphs] = orig.glyphs.log_clusters[i] - splitIndex
		}
	} else {
		copy(new.glyphs.glyphs, orig.glyphs.glyphs[numRemaining:])
		copy(new.glyphs.log_clusters, orig.glyphs.log_clusters[numRemaining:])

		for i, l := range orig.glyphs.log_clusters[:numRemaining] {
			orig.glyphs.log_clusters[i] = l - splitIndex
		}
	}

	orig.glyphs.pango_glyph_string_set_size(len(orig.glyphs.glyphs) - numGlyphs)

	return &new
}

// @text: text that @glyphItem corresponds to
//   (glyphItem.item.offset is an offset from the
//    start of @text)
// @log_attrs: (array): logical attributes for the item
//   (the first logical attribute refers to the position
//   before the first character in the item)
// pango_glyph_item_letter_space adds spacing between the graphemes of `glyphItem` to
// give the effect of typographic letter spacing.
// `letter_spacing` is specified in Pango units and may be negative, though too large
//   negative values will give ugly result
func (glyphItem *GlyphItem) pango_glyph_item_letter_space(text []rune, logAttrs []CharAttr, letterSpacing GlyphUnit) {
	spaceLeft := letterSpacing / 2

	// hinting
	if (letterSpacing & (pangoScale - 1)) == 0 {
		spaceLeft = spaceLeft.PANGO_UNITS_ROUND()
	}

	spaceRight := letterSpacing - spaceLeft
	var (
		iter   GlyphItemIter
		glyphs = glyphItem.glyphs.glyphs
	)
	haveCluster := iter.pango_glyph_item_iter_init_start(glyphItem, text)
	for ; haveCluster; haveCluster = iter.pango_glyph_item_iter_next_cluster() {
		if !logAttrs[iter.start_char].IsCursorPosition() {
			continue
		}

		if iter.start_glyph < iter.end_glyph { // LTR
			if iter.start_char > 0 {
				glyphs[iter.start_glyph].geometry.width += spaceLeft
				glyphs[iter.start_glyph].geometry.x_offset += spaceLeft
			}
			if iter.end_char < glyphItem.item.num_chars {
				glyphs[iter.end_glyph-1].geometry.width += spaceRight
			}
		} else { // RTL
			if iter.start_char > 0 {
				glyphs[iter.start_glyph].geometry.width += spaceRight
			}
			if iter.end_char < glyphItem.item.num_chars {
				glyphs[iter.end_glyph+1].geometry.x_offset += spaceLeft
				glyphs[iter.end_glyph+1].geometry.width += spaceLeft
			}
		}
	}
}

// pango_glyph_item_get_logical_widths determine the screen width corresponding to each character. When
// multiple characters compose a single cluster, the width of the entire
// cluster is divided equally among the characters.
// It returns an array whose length is the number of characters in glyphItem (equal to
// glyphItem.item.num_chars)
func (glyphItem *GlyphItem) pango_glyph_item_get_logical_widths(text []rune) []GlyphUnit {
	logicalWidths := make([]GlyphUnit, glyphItem.item.num_chars)

	dir := -1
	if glyphItem.LTR() {
		dir = +1
	}

	var iter GlyphItemIter
	hasCluster := iter.pango_glyph_item_iter_init_start(glyphItem, text)
	for ; hasCluster; hasCluster = iter.pango_glyph_item_iter_next_cluster() {
		var clusterWidth GlyphUnit
		for glyphIndex := iter.start_glyph; glyphIndex != iter.end_glyph; glyphIndex += dir {
			clusterWidth += glyphItem.glyphs.glyphs[glyphIndex].geometry.width
		}

		numChars := GlyphUnit(iter.end_char - iter.start_char)
		if numChars != 0 { // pedantic
			charWidth := clusterWidth / numChars

			for charIndex := iter.start_char; charIndex < iter.end_char; charIndex++ {
				logicalWidths[charIndex] = charWidth
			}

			// add any residues to the first char
			logicalWidths[iter.start_char] += clusterWidth - (charWidth * numChars)
		}
	}

	return logicalWidths
}

func (run *GlyphItem) pango_layout_run_get_extents_and_height(runInk, runLogical *Rectangle, height *int) {
	var (
		logical Rectangle
		metrics *FontMetrics
	)

	if runInk == nil && runLogical == nil {
		return
	}

	properties := run.item.pango_layout_get_item_properties()

	has_underline := properties.uline_single || properties.uline_double ||
		properties.uline_low || properties.uline_error
	has_overline := properties.oline_single

	if runLogical == nil && (run.item.analysis.flags&PANGO_ANALYSIS_FLAG_CENTERED_BASELINE) != 0 {
		runLogical = &logical
	}

	if runLogical == nil && (has_underline || has_overline || properties.strikethrough) {
		runLogical = &logical
	}

	if properties.shape != nil {
		properties.shape._pango_shape_get_extents(run.item.num_chars, runInk, runLogical)
	} else {
		run.glyphs.pango_glyph_string_extents(run.item.analysis.font, runInk, runLogical)
	}

	if runInk != nil && (has_underline || has_overline || properties.strikethrough) {
		if metrics == nil {
			me := run.item.analysis.font.get_metrics(run.item.analysis.language)
			metrics = &me
		}

		underlineThickness := metrics.underline_thickness
		underlinePosition := metrics.underline_position
		strikethroughThickness := metrics.strikethrough_thickness
		strikethroughPosition := metrics.strikethrough_position

		// the underline/strikethrough takes x,width of logical_rect. reflect
		// that into ink_rect.
		newPos := min(runInk.x, runLogical.x)
		runInk.width = max(runInk.x+runInk.width, runLogical.x+runLogical.width) - newPos
		runInk.x = newPos

		// We should better handle the case of height==0 in the following cases.
		// If runInk.height == 0, we should adjust runInk.y appropriately.

		if properties.strikethrough {
			if runInk.height == 0 {
				runInk.height = strikethroughThickness
				runInk.y = -strikethroughPosition
			}
		}

		if properties.oline_single {
			runInk.y -= underlineThickness
			runInk.height += underlineThickness
		}

		if properties.uline_low {
			runInk.height += 2 * underlineThickness
		}
		if properties.uline_single {
			runInk.height = max(runInk.height, underlineThickness-underlinePosition-runInk.y)
		}
		if properties.uline_double {
			runInk.height = max(runInk.height, 3*underlineThickness-underlinePosition-runInk.y)
		}
		if properties.uline_error {
			runInk.height = max(runInk.height, 3*underlineThickness-underlinePosition-runInk.y)
		}
	}

	if height != nil {
		if metrics == nil {
			me := run.item.analysis.font.get_metrics(run.item.analysis.language)
			metrics = &me
		}
		*height = metrics.height
	}

	if run.item.analysis.flags&PANGO_ANALYSIS_FLAG_CENTERED_BASELINE != 0 {
		is_hinted := (runLogical.y & runLogical.height & (pangoScale - 1)) == 0
		adjustment := GlyphUnit(runLogical.y + runLogical.height/2)

		if is_hinted {
			adjustment = adjustment.PANGO_UNITS_ROUND()
		}

		properties.rise += adjustment
	}

	if properties.rise != 0 {
		if runInk != nil {
			runInk.y -= int(properties.rise)
		}

		if runLogical != nil {
			runLogical.y -= int(properties.rise)
		}
	}
}

// Tack `attrs` onto the attributes of glyphItem
func (glyphItem *GlyphItem) append_attrs(attrs AttrList) {
	glyphItem.item.analysis.extra_attrs = append(glyphItem.item.analysis.extra_attrs, attrs...)
}

type ApplyAttrsState struct {
	iter         GlyphItemIter
	segmentAttrs AttrList
}

// split the glyph item at the start of the current cluster
func (state *ApplyAttrsState) splitBeforeClusterStart() *GlyphItem {
	splitLen := state.iter.start_index - state.iter.glyphItem.item.offset
	splitItem := state.iter.glyphItem.pango_glyph_item_split(state.iter.text, splitLen)
	splitItem.append_attrs(state.segmentAttrs)

	// adjust iteration to account for the split
	if state.iter.glyphItem.LTR() {
		state.iter.start_glyph -= len(splitItem.glyphs.glyphs)
		state.iter.end_glyph -= len(splitItem.glyphs.glyphs)
	}

	state.iter.start_char -= splitItem.item.num_chars
	state.iter.end_char -= splitItem.item.num_chars

	return splitItem
}

// pango_glyph_item_apply_attrs splits a shaped item into multiple items based
// on an attribute list. The idea is that if you have attributes
// that don't affect shaping, such as color or underline, to avoid
// affecting shaping, you filter them out (pango_attr_list_filter()),
// apply the shaping process and then reapply them to the result using
// this function.
//
// All attributes that start or end inside a cluster are applied
// to that cluster; for instance, if half of a cluster is underlined
// and the other-half strikethrough, then the cluster will end
// up with both underline and strikethrough attributes. In these
// cases, it may happen that item.extra_attrs for some of the
// result items can have multiple attributes of the same type.
func (glyphItem *GlyphItem) pango_glyph_item_apply_attrs(text []rune, list AttrList) *runList {
	//    PangoAttrIterator iter;
	//    GSList *result = null;
	//    ApplyAttrsState state;
	//    gboolean startNewSegment = false;
	//    gboolean haveCluster;
	//    gboolean isEllipsis;

	// This routine works by iterating through the item cluster by
	// cluster; we accumulate the attributes that we need to
	// add to the next output item, and decide when to split
	// off an output item based on two criteria:
	//
	// A) If start_index < attribute_start < end_index
	//    (attribute starts within cluster) then we need
	//    to split between the last cluster and this cluster.
	// B) If start_index < attribute_end <= end_index,
	//    (attribute ends within cluster) then we need to
	//    split between this cluster and the next one.

	var (
		rangeStart, rangeEnd         int
		haveCluster, startNewSegment bool
		result                       *runList
	)
	// Advance the attr iterator to the start of the item

	iter := list.pango_attr_list_get_iterator()
	for do := true; do; do = iter.pango_attr_iterator_next() {
		rangeStart, rangeEnd = iter.StartIndex, iter.EndIndex
		if rangeEnd > glyphItem.item.offset {
			break
		}
	}

	var state ApplyAttrsState
	state.segmentAttrs = iter.pango_attr_iterator_get_attrs()

	isEllipsis := (glyphItem.item.analysis.flags & PANGO_ANALYSIS_FLAG_IS_ELLIPSIS) != 0

	// Short circuit the case when we don't actually need to split the item
	if isEllipsis || (rangeStart <= glyphItem.item.offset &&
		rangeEnd >= glyphItem.item.offset+glyphItem.item.num_chars) {
		goto out
	}

	haveCluster = state.iter.pango_glyph_item_iter_init_start(glyphItem, text)
	for ; haveCluster; haveCluster = state.iter.pango_glyph_item_iter_next_cluster() {
		haveNext := false

		/* [rangeStart,rangeEnd] is the first range that intersects
		* the current cluster.
		 */

		/* Split item into two, if this cluster isn't a continuation
		* of the last cluster
		 */
		if startNewSegment {
			result = &runList{next: result, data: state.splitBeforeClusterStart()}
			state.segmentAttrs = iter.pango_attr_iterator_get_attrs()
		}

		startNewSegment = false

		// Loop over all ranges that intersect this cluster; exiting
		// leaving [rangeStart,rangeEnd] being the first range that
		// intersects the next cluster.
		for do := true; do; do = haveNext {
			if rangeEnd > state.iter.end_index {
				/* Range intersects next cluster */
				break
			}

			// Since ranges end in this cluster, the next cluster goes into a
			// separate segment
			startNewSegment = true

			haveNext = iter.pango_attr_iterator_next()
			rangeStart, rangeEnd = iter.StartIndex, iter.EndIndex

			if rangeStart >= state.iter.end_index {
				// New range doesn't intersect this cluster */
				// No gap between ranges, so previous range must of ended
				// at cluster boundary.
				if debugMode {
					assert(rangeStart == state.iter.end_index && startNewSegment)
				}
				break
			}

			/* If any ranges start *inside* this cluster, then we need
			* to split the previous cluster into a separate segment
			 */
			if rangeStart > state.iter.start_index &&
				state.iter.start_index != glyphItem.item.offset {
				newAttrs := state.segmentAttrs.pango_attr_list_copy()
				result = &runList{next: result, data: state.splitBeforeClusterStart()}
				state.segmentAttrs = newAttrs
			}

			state.segmentAttrs = append(state.segmentAttrs, iter.pango_attr_iterator_get_attrs()...)
		}
	}

out:
	// what's left in glyphItem is the remaining portion
	glyphItem.append_attrs(state.segmentAttrs)
	result = &runList{next: result, data: glyphItem}
	if glyphItem.LTR() {
		result = result.reverse()
	}

	return result
}

// GlyphItemIter is an iterator over the clusters in a
// `GlyphItem`. The forward direction of the
// iterator is the logical direction of text. That is, with increasing
// `start_index` and `start_char` values. If `glyphItem` is right-to-left
// (that is, if `glyphItem.item.analysis.level` is odd),
// then `start_glyph` decreases as the iterator moves forward. Moreover,
// in right-to-left cases, `start_glyph` is greater than `end_glyph`.
//
// An iterator should be initialized using either of
// `pango_glyph_item_iter_init_start()` and
// `pango_glyph_item_iter_init_end()`, for forward and backward iteration
// respectively, and walked over using any desired mixture of
// `pango_glyph_item_iter_next_cluster()` and
// `pango_glyph_item_iter_prev_cluster()`.
//
// Note that `text` is the start of the text to layout, which is then
// indexed by `glyphItem.item.offset` to get to the
// text of `glyphItem`. The `start_index` and `end_index` values can directly
// index into `text`. The `start_glyph`, `end_glyph`, `start_char`, and `end_char`
// values however are zero-based for the `glyphItem`. For each cluster, the
// item pointed at by the start variables is included in the cluster while
// the one pointed at by end variables is not.
type GlyphItemIter struct {
	glyphItem *GlyphItem
	text      []rune

	start_glyph, end_glyph int // index into text[glyphItem.item.offset:]
	start_char, end_char   int // index into text[glyphItem.item.offset:]

	start_index, end_index int // index into text
}

// pango_glyph_item_iter_next_cluster advances the iterator to the next cluster in the glyph item.
func (iter *GlyphItemIter) pango_glyph_item_iter_next_cluster() bool {
	var (
		glyph_index = iter.end_glyph
		glyphs      = iter.glyphItem.glyphs
		cluster     int
		item        = iter.glyphItem.item
	)

	if iter.glyphItem.LTR() {
		if glyph_index == len(glyphs.glyphs) {
			return false
		}
	} else {
		if glyph_index < 0 {
			return false
		}
	}

	iter.start_glyph = iter.end_glyph
	iter.start_index = iter.end_index
	iter.start_char = iter.end_char

	if iter.glyphItem.LTR() {
		cluster = glyphs.log_clusters[glyph_index]
		for {
			glyph_index++

			if glyph_index == len(glyphs.glyphs) {
				iter.end_index = item.offset + item.num_chars
				iter.end_char = item.num_chars
				break
			}

			if glyphs.log_clusters[glyph_index] > cluster {
				iter.end_index = item.offset + glyphs.log_clusters[glyph_index]
				iter.end_char += iter.end_index - iter.start_index
				break
			}
		}
	} else { /* RTL */
		cluster = glyphs.log_clusters[glyph_index]
		for {
			glyph_index--

			if glyph_index < 0 {
				iter.end_index = item.offset + item.num_chars
				iter.end_char = item.num_chars
				break
			}

			if glyphs.log_clusters[glyph_index] > cluster {
				iter.end_index = item.offset + glyphs.log_clusters[glyph_index]
				iter.end_char += iter.end_index - iter.start_index
				break
			}
		}
	}

	iter.end_glyph = glyph_index

	if debugMode {
		assert(iter.start_char <= iter.end_char)
		assert(iter.end_char <= item.num_chars)
	}

	return true
}

// pango_glyph_item_iter_prev_cluster moves the iterator to the preceding cluster in the glyph item.
func (iter *GlyphItemIter) pango_glyph_item_iter_prev_cluster() bool {
	var (
		glyph_index = iter.start_glyph
		glyphs      = iter.glyphItem.glyphs
		cluster     int
		item        = iter.glyphItem.item
	)

	if iter.glyphItem.LTR() {
		if glyph_index == 0 {
			return false
		}
	} else {
		if glyph_index == len(glyphs.glyphs)-1 {
			return false
		}

	}

	iter.end_glyph = iter.start_glyph
	iter.end_index = iter.start_index
	iter.end_char = iter.start_char

	if iter.glyphItem.LTR() {
		cluster = glyphs.log_clusters[glyph_index-1]
		for {
			if glyph_index == 0 {
				iter.start_index = item.offset
				iter.start_char = 0
				break
			}

			glyph_index--

			if glyphs.log_clusters[glyph_index] < cluster {
				glyph_index++
				iter.start_index = item.offset + glyphs.log_clusters[glyph_index]
				iter.start_char -= iter.end_index - iter.start_index
				break
			}
		}
	} else { /* RTL */
		cluster = glyphs.log_clusters[glyph_index+1]
		for {
			if glyph_index == len(glyphs.glyphs)-1 {
				iter.start_index = item.offset
				iter.start_char = 0
				break
			}

			glyph_index++

			if glyphs.log_clusters[glyph_index] < cluster {
				glyph_index--
				iter.start_index = item.offset + glyphs.log_clusters[glyph_index]
				iter.start_char -= iter.end_index - iter.start_index
				break
			}
		}
	}

	iter.start_glyph = glyph_index

	if debugMode {
		assert(iter.start_char <= iter.end_char)
		assert(0 <= iter.start_char)
	}

	return true
}

// pango_glyph_item_iter_init_start initializes a #GlyphItemIter structure to point to the
// first cluster in a glyph item.
func (iter *GlyphItemIter) pango_glyph_item_iter_init_start(glyphItem *GlyphItem, text []rune) bool {
	iter.glyphItem = glyphItem
	iter.text = text

	if glyphItem.LTR() {
		iter.end_glyph = 0
	} else {
		iter.end_glyph = len(glyphItem.glyphs.glyphs) - 1
	}

	iter.end_index = glyphItem.item.offset
	iter.end_char = 0

	iter.start_glyph = iter.end_glyph
	iter.start_index = iter.end_index
	iter.start_char = iter.end_char

	// advance onto the first cluster of the glyph item
	return iter.pango_glyph_item_iter_next_cluster()
}

// pango_glyph_item_iter_init_end initializes a `GlyphItemIter` structure to point to the
// last cluster in a glyph item.
func (iter *GlyphItemIter) pango_glyph_item_iter_init_end(glyphItem *GlyphItem, text []rune) bool {
	iter.glyphItem = glyphItem
	iter.text = text

	if glyphItem.LTR() {
		iter.start_glyph = len(glyphItem.glyphs.glyphs)
	} else {
		iter.start_glyph = -1
	}

	iter.start_index = glyphItem.item.offset + glyphItem.item.num_chars
	iter.start_char = glyphItem.item.num_chars

	iter.end_glyph = iter.start_glyph
	iter.end_index = iter.start_index
	iter.end_char = iter.start_char

	/* Advance onto the first cluster of the glyph item */
	return iter.pango_glyph_item_iter_prev_cluster()
}
