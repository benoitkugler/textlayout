package pango

// GlyphItem is a pair of a Item and the glyphs
// resulting from shaping the text corresponding to an item.
// As an example of the usage of GlyphItem, the results
// of shaping text with Layout is a list of LayoutLine,
// each of which contains a list of GlyphItem.
type GlyphItem struct {
	Item   *Item
	Glyphs *GlyphString
}

// LTR returns true if the input text level was Left-To-Right.
func (g GlyphItem) LTR() bool {
	return g.Item.Analysis.Level%2 == 0
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
	if orig.Item.Length <= 0 || splitIndex <= 0 || splitIndex >= orig.Item.Length {
		return nil
	}

	var i, numGlyphs int
	if orig.LTR() {
		for i = 0; i < len(orig.Glyphs.logClusters); i++ {
			if orig.Glyphs.logClusters[i] >= splitIndex {
				break
			}
		}

		if i == len(orig.Glyphs.Glyphs) {
			/* No splitting necessary */
			return nil
		}

		splitIndex = orig.Glyphs.logClusters[i]
		numGlyphs = i
	} else {
		for i = len(orig.Glyphs.Glyphs) - 1; i >= 0; i-- {
			if orig.Glyphs.logClusters[i] >= splitIndex {
				break
			}
		}

		if i < 0 {
			/* No splitting necessary */
			return nil
		}

		splitIndex = orig.Glyphs.logClusters[i]
		numGlyphs = len(orig.Glyphs.Glyphs) - 1 - i
	}

	var new GlyphItem
	new.Item = orig.Item.pango_item_split(splitIndex)
	new.Glyphs = &GlyphString{}
	new.Glyphs.setSize(numGlyphs)

	numRemaining := len(orig.Glyphs.Glyphs) - numGlyphs
	if orig.LTR() {
		copy(new.Glyphs.Glyphs, orig.Glyphs.Glyphs[:numGlyphs])
		copy(new.Glyphs.logClusters, orig.Glyphs.logClusters[:numGlyphs])

		copy(orig.Glyphs.Glyphs, orig.Glyphs.Glyphs[numGlyphs:])
		for i = numGlyphs; i < len(orig.Glyphs.Glyphs); i++ {
			orig.Glyphs.logClusters[i-numGlyphs] = orig.Glyphs.logClusters[i] - splitIndex
		}
	} else {
		copy(new.Glyphs.Glyphs, orig.Glyphs.Glyphs[numRemaining:])
		copy(new.Glyphs.logClusters, orig.Glyphs.logClusters[numRemaining:])

		for i, l := range orig.Glyphs.logClusters[:numRemaining] {
			orig.Glyphs.logClusters[i] = l - splitIndex
		}
	}

	orig.Glyphs.setSize(len(orig.Glyphs.Glyphs) - numGlyphs)

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
	if (letterSpacing & (Scale - 1)) == 0 {
		spaceLeft = spaceLeft.Round()
	}

	spaceRight := letterSpacing - spaceLeft
	var (
		iter   GlyphItemIter
		glyphs = glyphItem.Glyphs.Glyphs
	)
	haveCluster := iter.InitStart(glyphItem, text)
	for ; haveCluster; haveCluster = iter.NextCluster() {
		if !logAttrs[iter.StartChar].IsCursorPosition() {
			continue
		}

		if iter.startGlyph < iter.endGlyph { // LTR
			if iter.StartChar > 0 {
				glyphs[iter.startGlyph].Geometry.Width += spaceLeft
				glyphs[iter.startGlyph].Geometry.xOffset += spaceLeft
			}
			if iter.EndChar < glyphItem.Item.Length {
				glyphs[iter.endGlyph-1].Geometry.Width += spaceRight
			}
		} else { // RTL
			if iter.StartChar > 0 {
				glyphs[iter.startGlyph].Geometry.Width += spaceRight
			}
			if iter.EndChar < glyphItem.Item.Length {
				glyphs[iter.endGlyph+1].Geometry.xOffset += spaceLeft
				glyphs[iter.endGlyph+1].Geometry.Width += spaceLeft
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
	logicalWidths := make([]GlyphUnit, glyphItem.Item.Length)

	dir := -1
	if glyphItem.LTR() {
		dir = +1
	}

	var iter GlyphItemIter
	hasCluster := iter.InitStart(glyphItem, text)
	for ; hasCluster; hasCluster = iter.NextCluster() {
		var clusterWidth GlyphUnit
		for glyphIndex := iter.startGlyph; glyphIndex != iter.endGlyph; glyphIndex += dir {
			clusterWidth += glyphItem.Glyphs.Glyphs[glyphIndex].Geometry.Width
		}

		numChars := GlyphUnit(iter.EndChar - iter.StartChar)
		if numChars != 0 { // pedantic
			charWidth := clusterWidth / numChars

			for charIndex := iter.StartChar; charIndex < iter.EndChar; charIndex++ {
				logicalWidths[charIndex] = charWidth
			}

			// add any residues to the first char
			logicalWidths[iter.StartChar] += clusterWidth - (charWidth * numChars)
		}
	}

	return logicalWidths
}

func (run *GlyphItem) getExtentsAndHeight(runInk, runLogical *Rectangle, height *int32) {
	var (
		logical Rectangle
		metrics *FontMetrics
	)

	if runInk == nil && runLogical == nil {
		return
	}

	properties := run.Item.pango_layout_get_item_properties()

	has_underline := properties.ulineSingle || properties.ulineDouble ||
		properties.ulineLow || properties.ulineError
	has_overline := properties.olineSingle

	if runLogical == nil && (run.Item.Analysis.Flags&AFCenterdBaseline) != 0 {
		runLogical = &logical
	}

	if runLogical == nil && (has_underline || has_overline || properties.strikethrough) {
		runLogical = &logical
	}

	if properties.shape != nil {
		properties.shape.getExtents(int32(run.Item.Length), runInk, runLogical)
	} else {
		run.Glyphs.Extents(run.Item.Analysis.Font, runInk, runLogical)
	}

	if runInk != nil && (has_underline || has_overline || properties.strikethrough) {
		if metrics == nil {
			me := run.Item.Analysis.Font.GetMetrics(run.Item.Analysis.Language)
			metrics = &me
		}

		underlineThickness := metrics.UnderlineThickness
		underlinePosition := metrics.UnderlinePosition
		strikethroughThickness := metrics.StrikethroughThickness
		strikethroughPosition := metrics.StrikethroughPosition

		// the underline/strikethrough takes x,width of logical_rect. reflect
		// that into ink_rect.
		newPos := min32(runInk.X, runLogical.X)
		runInk.Width = max32(runInk.X+runInk.Width, runLogical.X+runLogical.Width) - newPos
		runInk.X = newPos

		// We should better handle the case of height==0 in the following cases.
		// If runInk.height == 0, we should adjust runInk.y appropriately.

		if properties.strikethrough {
			if runInk.Height == 0 {
				runInk.Height = strikethroughThickness
				runInk.Y = -strikethroughPosition
			}
		}

		if properties.olineSingle {
			runInk.Y -= underlineThickness
			runInk.Height += underlineThickness
		}

		if properties.ulineLow {
			runInk.Height += 2 * underlineThickness
		}
		if properties.ulineSingle {
			runInk.Height = max32(runInk.Height, underlineThickness-underlinePosition-runInk.Y)
		}
		if properties.ulineDouble {
			runInk.Height = max32(runInk.Height, 3*underlineThickness-underlinePosition-runInk.Y)
		}
		if properties.ulineError {
			runInk.Height = max32(runInk.Height, 3*underlineThickness-underlinePosition-runInk.Y)
		}
	}

	if height != nil {
		if metrics == nil {
			me := run.Item.Analysis.Font.GetMetrics(run.Item.Analysis.Language)
			metrics = &me
		}
		*height = metrics.Height
	}

	if run.Item.Analysis.Flags&AFCenterdBaseline != 0 {
		is_hinted := (runLogical.Y & runLogical.Height & (Scale - 1)) == 0
		adjustment := GlyphUnit(runLogical.Y + runLogical.Height/2)

		if is_hinted {
			adjustment = adjustment.Round()
		}

		properties.Rise += adjustment
	}

	if properties.Rise != 0 {
		if runInk != nil {
			runInk.Y -= int32(properties.Rise)
		}

		if runLogical != nil {
			runLogical.Y -= int32(properties.Rise)
		}
	}
}

// Tack `attrs` onto the attributes of glyphItem
func (glyphItem *GlyphItem) append_attrs(attrs AttrList) {
	glyphItem.Item.Analysis.ExtraAttrs = append(glyphItem.Item.Analysis.ExtraAttrs, attrs...)
}

type applyAttrsState struct {
	segmentAttrs AttrList
	iter         GlyphItemIter
}

// split the glyph item at the start of the current cluster
func (state *applyAttrsState) splitBeforeClusterStart() *GlyphItem {
	splitLen := state.iter.StartIndex - state.iter.glyphItem.Item.Offset
	splitItem := state.iter.glyphItem.pango_glyph_item_split(state.iter.text, splitLen)
	splitItem.append_attrs(state.segmentAttrs)

	// adjust iteration to account for the split
	if state.iter.glyphItem.LTR() {
		state.iter.startGlyph -= len(splitItem.Glyphs.Glyphs)
		state.iter.endGlyph -= len(splitItem.Glyphs.Glyphs)
	}

	state.iter.StartChar -= splitItem.Item.Length
	state.iter.EndChar -= splitItem.Item.Length

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
func (glyphItem *GlyphItem) pango_glyph_item_apply_attrs(text []rune, list AttrList) *RunList {
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
		result                       *RunList
	)
	// Advance the attr iterator to the start of the item

	iter := list.getIterator()
	for do := true; do; do = iter.next() {
		rangeStart, rangeEnd = iter.StartIndex, iter.EndIndex
		if rangeEnd > glyphItem.Item.Offset {
			break
		}
	}

	var state applyAttrsState
	state.segmentAttrs = iter.getAttributes()

	isEllipsis := (glyphItem.Item.Analysis.Flags & AFIsEllipsis) != 0

	// Short circuit the case when we don't actually need to split the item
	if isEllipsis || (rangeStart <= glyphItem.Item.Offset &&
		rangeEnd >= glyphItem.Item.Offset+glyphItem.Item.Length) {
		goto out
	}

	haveCluster = state.iter.InitStart(glyphItem, text)
	for ; haveCluster; haveCluster = state.iter.NextCluster() {
		haveNext := false

		/* [rangeStart,rangeEnd] is the first range that intersects
		* the current cluster.
		 */

		/* Split item into two, if this cluster isn't a continuation
		* of the last cluster
		 */
		if startNewSegment {
			result = &RunList{Next: result, Data: state.splitBeforeClusterStart()}
			state.segmentAttrs = iter.getAttributes()
		}

		startNewSegment = false

		// Loop over all ranges that intersect this cluster; exiting
		// leaving [rangeStart,rangeEnd] being the first range that
		// intersects the next cluster.
		for do := true; do; do = haveNext {
			if rangeEnd > state.iter.EndIndex {
				/* Range intersects next cluster */
				break
			}

			// Since ranges end in this cluster, the next cluster goes into a
			// separate segment
			startNewSegment = true

			haveNext = iter.next()
			rangeStart, rangeEnd = iter.StartIndex, iter.EndIndex

			if rangeStart >= state.iter.EndIndex {
				// New range doesn't intersect this cluster */
				// No gap between ranges, so previous range must of ended
				// at cluster boundary.
				if debugMode {
					assert(rangeStart == state.iter.EndIndex && startNewSegment, "applyItemAttrs")
				}
				break
			}

			/* If any ranges start *inside* this cluster, then we need
			* to split the previous cluster into a separate segment
			 */
			if rangeStart > state.iter.StartIndex &&
				state.iter.StartIndex != glyphItem.Item.Offset {
				newAttrs := state.segmentAttrs.pango_attr_list_copy()
				result = &RunList{Next: result, Data: state.splitBeforeClusterStart()}
				state.segmentAttrs = newAttrs
			}

			state.segmentAttrs = append(state.segmentAttrs, iter.getAttributes()...)
		}
	}

out:
	// what's left in glyphItem is the remaining portion
	glyphItem.append_attrs(state.segmentAttrs)
	result = &RunList{Next: result, Data: glyphItem}
	if glyphItem.LTR() {
		result = result.reverse()
	}

	return result
}

// GlyphItemIter is an iterator over the clusters in a
// `GlyphItem`. The forward direction of the
// iterator is the logical direction of text. That is, with increasing
// `StartIndex` and `StartChar` values. If the `glyphItem` is right-to-left
// (that is, if `glyphItem.Item.Analysis.Level` is odd),
// then `startGlyph` decreases as the iterator moves forward. Moreover,
// in right-to-left cases, `startGlyph` is greater than `endGlyph`.
//
// An iterator should be initialized using either of
// `InitStart()` and `InitEnd()`, for forward and backward iteration
// respectively, and walked over using any desired mixture of
// `NextCluster()` and `PrevCluster()`.
//
// Note that `text` is the start of the text to layout, which is then
// indexed by `glyphItem.Item.Offset` to get to the
// text of `glyphItem`. The `StartIndex` and `EndIndex` values can directly
// index into `text`. The `startGlyph`, `endGlyph`, `StartChar`, and `EndChar`
// values however are zero-based for the `glyphItem`. For each cluster, the
// item pointed at by the start variables is included in the cluster while
// the one pointed at by end variables is not.
type GlyphItemIter struct {
	glyphItem *GlyphItem
	text      []rune

	startGlyph, endGlyph int // index into text[glyphItem.Item.Offset:]

	// Index into text[glyphItem.Item.Offset:]
	StartChar, EndChar int

	StartIndex, EndIndex int // Index into text
}

// NextCluster advances the iterator to the next cluster in the glyph item.
func (iter *GlyphItemIter) NextCluster() bool {
	var (
		glyphIndex = iter.endGlyph
		glyphs     = iter.glyphItem.Glyphs
		cluster    int
		item       = iter.glyphItem.Item
	)

	if iter.glyphItem.LTR() {
		if glyphIndex == len(glyphs.Glyphs) {
			return false
		}
	} else {
		if glyphIndex < 0 {
			return false
		}
	}

	iter.startGlyph = iter.endGlyph
	iter.StartIndex = iter.EndIndex
	iter.StartChar = iter.EndChar

	if iter.glyphItem.LTR() {
		cluster = glyphs.logClusters[glyphIndex]
		for {
			glyphIndex++

			if glyphIndex == len(glyphs.Glyphs) {
				iter.EndIndex = item.Offset + item.Length
				iter.EndChar = item.Length
				break
			}

			if glyphs.logClusters[glyphIndex] > cluster {
				iter.EndIndex = item.Offset + glyphs.logClusters[glyphIndex]
				iter.EndChar += iter.EndIndex - iter.StartIndex
				break
			}
		}
	} else { /* RTL */
		cluster = glyphs.logClusters[glyphIndex]
		for {
			glyphIndex--

			if glyphIndex < 0 {
				iter.EndIndex = item.Offset + item.Length
				iter.EndChar = item.Length
				break
			}

			if glyphs.logClusters[glyphIndex] > cluster {
				iter.EndIndex = item.Offset + glyphs.logClusters[glyphIndex]
				iter.EndChar += iter.EndIndex - iter.StartIndex
				break
			}
		}
	}

	iter.endGlyph = glyphIndex

	if debugMode {
		assert(iter.StartChar <= iter.EndChar && iter.EndChar <= item.Length, "nextCluster")
	}

	return true
}

// PrevCluster moves the iterator to the preceding cluster in the glyph item.
func (iter *GlyphItemIter) PrevCluster() bool {
	var (
		glyphIndex = iter.startGlyph
		glyphs     = iter.glyphItem.Glyphs
		cluster    int
		item       = iter.glyphItem.Item
	)

	if iter.glyphItem.LTR() {
		if glyphIndex == 0 {
			return false
		}
	} else {
		if glyphIndex == len(glyphs.Glyphs)-1 {
			return false
		}
	}

	iter.endGlyph = iter.startGlyph
	iter.EndIndex = iter.StartIndex
	iter.EndChar = iter.StartChar

	if iter.glyphItem.LTR() {
		cluster = glyphs.logClusters[glyphIndex-1]
		for {
			if glyphIndex == 0 {
				iter.StartIndex = item.Offset
				iter.StartChar = 0
				break
			}

			glyphIndex--

			if glyphs.logClusters[glyphIndex] < cluster {
				glyphIndex++
				iter.StartIndex = item.Offset + glyphs.logClusters[glyphIndex]
				iter.StartChar -= iter.EndIndex - iter.StartIndex
				break
			}
		}
	} else { /* RTL */
		cluster = glyphs.logClusters[glyphIndex+1]
		for {
			if glyphIndex == len(glyphs.Glyphs)-1 {
				iter.StartIndex = item.Offset
				iter.StartChar = 0
				break
			}

			glyphIndex++

			if glyphs.logClusters[glyphIndex] < cluster {
				glyphIndex--
				iter.StartIndex = item.Offset + glyphs.logClusters[glyphIndex]
				iter.StartChar -= iter.EndIndex - iter.StartIndex
				break
			}
		}
	}

	iter.startGlyph = glyphIndex

	if debugMode {
		assert(iter.StartChar <= iter.EndChar && 0 <= iter.StartChar, "prevCluster")
	}

	return true
}

// InitStart initializes a `GlyphItemIter` structure to point to the
// first cluster in a glyph item, and returns `true` if it finds one.
func (iter *GlyphItemIter) InitStart(glyphItem *GlyphItem, text []rune) bool {
	iter.glyphItem = glyphItem
	iter.text = text

	if glyphItem.LTR() {
		iter.endGlyph = 0
	} else {
		iter.endGlyph = len(glyphItem.Glyphs.Glyphs) - 1
	}

	iter.EndIndex = glyphItem.Item.Offset
	iter.EndChar = 0

	iter.startGlyph = iter.endGlyph
	iter.StartIndex = iter.EndIndex
	iter.StartChar = iter.EndChar

	// advance onto the first cluster of the glyph item
	return iter.NextCluster()
}

// InitEnd initializes a `GlyphItemIter` structure to point to the
// last cluster in a glyph item, and returns `true` if it finds one.
func (iter *GlyphItemIter) InitEnd(glyphItem *GlyphItem, text []rune) bool {
	iter.glyphItem = glyphItem
	iter.text = text

	if glyphItem.LTR() {
		iter.startGlyph = len(glyphItem.Glyphs.Glyphs)
	} else {
		iter.startGlyph = -1
	}

	iter.StartIndex = glyphItem.Item.Offset + glyphItem.Item.Length
	iter.StartChar = glyphItem.Item.Length

	iter.endGlyph = iter.startGlyph
	iter.EndIndex = iter.StartIndex
	iter.EndChar = iter.StartChar

	/* Advance onto the first cluster of the glyph item */
	return iter.PrevCluster()
}
