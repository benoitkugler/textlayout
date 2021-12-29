package pango

// LayoutIter is a structure that can be used to
// iterate over the visual extents of a `Layout`.
type LayoutIter struct {
	layout *Layout

	/* If run is NULL, it means we're on a "virtual run"
	* at the end of the line with 0 width
	 */
	runs *RunList
	run  *GlyphItem

	/* list of Extents for each line in layout coordinates */
	lineExtents []extents

	lines     []*LayoutLine
	lineIndex int // index of the current line into lines

	// Index is the current readonly rune index in the text.
	// Note that iterating forward by char moves in visual order, not logical order, so indexes may not be
	// sequential. Also, the index may be equal to the length of the text
	// in the layout, if on the `nil` run (see `GetRun()`).
	Index int

	/* glyph offset to the current cluster start */
	clusterStart int

	/* first glyph in the next cluster */
	nextClusterGlyph int

	/* number of Unicode chars in current cluster */
	clusterNumChars int

	/* visual position of current character within the cluster */
	characterPosition int

	/* X position of the current run */
	runX Unit

	/* Width of the current run */
	runWidth   Unit
	endXOffset Unit

	/* X position of the left side of the current cluster */
	clusterX Unit

	/* The width of the current cluster */
	clusterWidth Unit

	/* the real width of layout */
	layoutWidth Unit

	/* this run is left-to-right */
	ltr bool
}

// GetIter returns an iterator to iterate over the visual extents of the layout.
// The first item is already loaded, meaning that a typical use of the iterator would be:
//		iter := GetIter()
//		for do := true; do; do = iter.NextXXX() {
//			item := iter.GetXXX()
//      }
func (layout *Layout) GetIter() *LayoutIter {
	var iter LayoutIter

	iter.layout = layout

	layout.checkLines()

	if len(layout.lines) == 0 {
		return &iter
	}

	iter.lines = layout.lines
	iter.lineIndex = 0

	runStartIndex := iter.line().StartIndex
	iter.runs = iter.line().Runs

	if iter.runs != nil {
		iter.run = iter.runs.Data
		runStartIndex = iter.run.Item.Offset
	} else {
		iter.run = nil
	}

	iter.lineExtents = nil

	if layout.Width == -1 {
		var logicalRect Rectangle

		iter.lineExtents = layout.getExtentsInternal(nil, &logicalRect, true)
		iter.layoutWidth = Unit(logicalRect.Width)
	} else {
		iter.lineExtents = layout.getExtentsInternal(nil, nil, true)
		iter.layoutWidth = layout.Width
	}

	iter.updateRun(runStartIndex)

	return &iter
}

// return the cluster width and the index of the next cluster start
func (gs *GlyphString) clusterWidth(clusterStart int) (Unit, int) {
	width := gs.Glyphs[clusterStart].Geometry.Width
	for i := clusterStart + 1; i < len(gs.Glyphs); i++ {
		glyph := gs.Glyphs[i]
		if glyph.attr.isClusterStart {
			return width, i
		}

		width += glyph.Geometry.Width
	}

	return width, len(gs.Glyphs)
}

// Sets up the iter for the start of a new cluster. clusterStartIndex
// is the rune index of the cluster start relative to the run.
func (iter *LayoutIter) updateCluster(clusterStartIndex int) {
	iter.characterPosition = 0

	gs := iter.run.Glyphs
	iter.clusterWidth, iter.nextClusterGlyph = gs.clusterWidth(iter.clusterStart)

	var clusterLength int
	if iter.ltr {
		// For LTR text, finding the length of the cluster is easy
		// since logical and visual runs are in the same direction.
		if iter.nextClusterGlyph < len(gs.Glyphs) {
			clusterLength = gs.LogClusters[iter.nextClusterGlyph] - clusterStartIndex
		} else {
			clusterLength = iter.run.Item.Length - clusterStartIndex
		}
	} else {
		// For RTL text, we have to scan backwards to find the previous
		// visual cluster which is the next logical cluster.
		i := iter.clusterStart
		for i > 0 && gs.LogClusters[i-1] == clusterStartIndex {
			i--
		}

		if i == 0 {
			clusterLength = iter.run.Item.Length - clusterStartIndex
		} else {
			clusterLength = gs.LogClusters[i-1] - clusterStartIndex
		}
	}

	iter.clusterNumChars = clusterLength

	if iter.ltr {
		iter.Index = iter.run.Item.Offset + clusterStartIndex
	} else {
		iter.Index = iter.run.Item.Offset + clusterStartIndex + clusterLength - 1
	}
}

func (iter *LayoutIter) updateRun(runStartIndex int) {
	lineExt := &iter.lineExtents[iter.lineIndex]

	// Note that in iter_new() the iter.run_width
	// is garbage but we don't use it since we're on the first run of
	// a line.
	if iter.runs == iter.line().Runs {
		iter.runX = Unit(lineExt.logicalRect.X)
	} else {
		iter.runX += iter.endXOffset + iter.runWidth
		if iter.run != nil {
			iter.runX += iter.run.startXOffset
		}
	}

	if iter.run != nil {
		iter.runWidth = iter.run.Glyphs.getWidth()
		iter.endXOffset = iter.run.endXOffset
	} else {
		/* The empty run at the end of a line */
		iter.runWidth = 0
		iter.endXOffset = 0
	}

	if iter.run != nil {
		iter.ltr = (iter.run.Item.Analysis.Level % 2) == 0
	} else {
		iter.ltr = true
	}

	iter.clusterStart = 0
	iter.clusterX = iter.runX

	if iter.run != nil {
		iter.updateCluster(iter.run.Glyphs.LogClusters[0])
	} else {
		iter.clusterWidth = 0
		iter.characterPosition = 0
		iter.clusterNumChars = 0
		iter.Index = runStartIndex
	}
}

// does no check
func (iter *LayoutIter) line() *LayoutLine { return iter.lines[iter.lineIndex] }

// GetLine gets the current line.
//
// Use the faster pango_layout_iter_get_line_readonly() if you do not plan
// to modify the contents of the line (glyphs, glyph widths, etc.).
func (iter *LayoutIter) GetLine() *LayoutLine {
	iter.line().leaked()
	return iter.line()
}

// NextLine moves `iter` forward to the start of the next line. If `iter` is
// already on the last line, returns `false`.
func (iter *LayoutIter) NextLine() bool {
	if iter.lineIndex+1 >= len(iter.lines) {
		return false
	}

	iter.lineIndex++
	iter.runs = iter.line().Runs

	if iter.runs != nil {
		iter.run = iter.runs.Data
	} else {
		iter.run = nil
	}

	iter.updateRun(iter.line().StartIndex)

	return true
}

// GetRun gets the current run. When iterating by run, at the end of each
// line, there's a position with a `nil` run, so this function can return
// `nil`. The `nil` run at the end of each line ensures that all lines have
// at least one run, even lines consisting of only a newline.
//
// Use the faster pango_layout_iter_get_run_readonly() if you do not plan
// to modify the contents of the run (glyphs, glyph widths, etc.).
func (iter *LayoutIter) GetRun() *GlyphItem {
	iter.line().leaked()
	return iter.run
}

// NextRun moves `iter` forward to the next run in visual order. If `iter` was
// already at the end of the layout, returns `false`.
func (iter *LayoutIter) NextRun() bool {
	if iter.run == nil {
		return iter.NextLine()
	}

	nextLink := iter.runs.Next

	var nextRunStart int
	if nextLink == nil {
		// Moving on to the zero-width "virtual run" at the end of each line
		nextRunStart = iter.run.Item.Offset + iter.run.Item.Length
		iter.run = nil
		iter.runs = nil
	} else {
		iter.runs = nextLink
		iter.run = iter.runs.Data
		nextRunStart = iter.run.Item.Offset
	}

	iter.updateRun(nextRunStart)

	return true
}

func (iter *LayoutIter) lineIsTerminated() bool {
	/* There is a real terminator at the end of each paragraph other
	* than the last.
	 */
	if iter.lineIndex < len(iter.lines)-1 {
		nextLine := iter.lines[iter.lineIndex+1]
		if nextLine.IsParagraphStart {
			return true
		}
	}

	return false
}

/* Moves to the next non-empty line. If @includeTerminators
 * is set, a line with just an explicit paragraph separator
 * is considered non-empty.
 */
func (iter *LayoutIter) nextNonemptyLine(includeTerminators bool) bool {
	for {
		result := iter.NextLine()
		if !result {
			return result
		}

		if iter.line().Runs != nil {
			return result
		}

		if includeTerminators && iter.lineIsTerminated() {
			return result
		}
	}
}

/* Moves to the next non-empty run. If @includeTerminators
 * is set, the trailing run at the end of a line with an explicit
 * paragraph separator is considered non-empty.
 */
func (iter *LayoutIter) nextNonemptyRun(includeTerminators bool) bool {
	for {
		result := iter.NextRun()
		if !result {
			return result
		}

		if iter.run != nil {
			return result
		}

		if includeTerminators && (iter.lineIsTerminated()) {
			return result
		}
	}
}

// NextCluster moves `iter` forward to the next cluster in visual order. If `iter`
// was already at the end of the layout, returns `false`.
func (iter *LayoutIter) NextCluster() bool {
	return iter.nextCluster(false)
}

// Like NextCluster(), but if `includeTerminators`
// is set, includes the fake runs/clusters for empty lines.
// (But not positions introduced by line wrapping).
func (iter *LayoutIter) nextCluster(includeTerminators bool) bool {
	if iter.run == nil {
		return iter.nextNonemptyLine(includeTerminators)
	}

	gs := iter.run.Glyphs

	nextStart := iter.nextClusterGlyph
	if nextStart == len(gs.Glyphs) {
		return iter.nextNonemptyRun(includeTerminators)
	}

	iter.clusterStart = nextStart
	iter.clusterX += iter.clusterWidth
	iter.updateCluster(gs.LogClusters[iter.clusterStart])

	return true
}

// NextChar moves `iter` forward to the next character in visual order. If `iter` was already at
// the end of the layout, returns `false`.
func (iter *LayoutIter) NextChar() bool {
	hasRN := func(s []rune) bool { return len(s) >= 2 && (s[0] == '\r' && s[1] == '\n') }

	if iter.run == nil {
		/* We need to fake an iterator position in the middle of a \r\n line terminator */
		if iter.lineIsTerminated() && hasRN(iter.layout.Text[iter.line().StartIndex+iter.line().Length:]) &&
			iter.characterPosition == 0 {
			iter.characterPosition++
			return true
		}

		return iter.nextNonemptyLine(true)
	}

	iter.characterPosition++
	if iter.characterPosition >= iter.clusterNumChars {
		return iter.nextCluster(true)
	}

	if iter.ltr {
		iter.Index += 1
	} else {
		iter.Index -= 1
	}

	return true
}

// GetCharExtents gets the extents of the current character, in layout coordinates
// (origin is the top left of the entire layout). Only logical extents
// can sensibly be obtained for characters; ink extents make sense only
// down to the level of clusters.
func (iter *LayoutIter) GetCharExtents() Rectangle {
	var clusterRect Rectangle

	iter.GetClusterExtents(nil, &clusterRect)

	if iter.run == nil {
		/* When on the nil run, cluster, char, and run all have the
		* same extents
		 */
		return clusterRect
	}

	var x0, x1 Unit
	if iter.clusterNumChars != 0 {
		x0 = (Unit(iter.characterPosition) * clusterRect.Width) / Unit(iter.clusterNumChars)
		x1 = ((Unit(iter.characterPosition) + 1) * clusterRect.Width) / Unit(iter.clusterNumChars)
	}

	return Rectangle{
		Width:  x1 - x0,
		Height: clusterRect.Height,
		Y:      clusterRect.Y,
		X:      clusterRect.X + x0,
	}
}

// GetClusterExtents gets the extents of the current cluster, in layout coordinates
// (origin is the top left of the entire layout).
func (iter *LayoutIter) GetClusterExtents(inkRect, logicalRect *Rectangle) {
	if iter.run == nil {
		/* When on the nil run, cluster, char, and run all have the
		* same extents
		 */
		iter.GetRunExtents(inkRect, logicalRect)
		return
	}

	iter.run.Glyphs.extentsRange(iter.clusterStart, iter.nextClusterGlyph, iter.run.Item.Analysis.Font,
		inkRect, logicalRect)

	if inkRect != nil {
		inkRect.X += iter.clusterX + iter.run.startXOffset
		inkRect.Y -= iter.run.yOffset
		offsetY(iter, &inkRect.Y)
	}

	if logicalRect != nil {
		if debugMode {
			assert(logicalRect.Width == iter.clusterWidth, "getClusterExtents")
		}
		logicalRect.X += iter.clusterX + iter.run.startXOffset
		logicalRect.Y -= iter.run.yOffset
		offsetY(iter, &logicalRect.Y)
	}
}

func offsetY(iter *LayoutIter, y *Unit) {
	*y += Unit(iter.lineExtents[iter.lineIndex].baseline)
}

// GetRunExtents gets the extents of the current run in layout coordinates
// (origin is the top left of the entire layout).
func (iter *LayoutIter) GetRunExtents(inkRect, logicalRect *Rectangle) {
	if iter.run != nil {
		iter.run.getExtentsAndHeight(inkRect, logicalRect, nil, nil)

		if inkRect != nil {
			offsetY(iter, &inkRect.Y)
			inkRect.X += iter.runX
		}

		if logicalRect != nil {
			offsetY(iter, &logicalRect.Y)
			logicalRect.X += iter.runX
		}
	} else {
		if runs := iter.line().Runs; runs != nil {
			/* The empty run at the end of a non-empty line */
			run := runs.last()
			run.getExtentsAndHeight(inkRect, logicalRect, nil, nil)
		} else {
			var r Rectangle
			iter.layout.getEmptyExtentsAndHeightAt(0, &r, false)
			if inkRect != nil {
				*inkRect = r
			}
			if logicalRect != nil {
				*logicalRect = r
			}
		}

		if inkRect != nil {
			offsetY(iter, &inkRect.Y)
			inkRect.X = iter.runX
			inkRect.Width = 0
		}

		if logicalRect != nil {
			offsetY(iter, &logicalRect.Y)
			logicalRect.X = iter.runX
			logicalRect.Width = 0
		}
	}
}

// GetLineExtents obtains the extents of the current line. `inkRect` or `logicalRect`
// can be `nil` if you aren't interested in them. Extents are in layout
// coordinates (origin is the top-left corner of the entire
// layout). Thus the extents returned by this function will be
// the same width/height but not at the same x/y as the extents
// returned from GetExtents().
func (iter *LayoutIter) GetLineExtents(inkRect, logicalRect *Rectangle) {
	ext := &iter.lineExtents[iter.lineIndex]

	if inkRect != nil {
		iter.line().getLineExtentsLayoutCoords(iter.layout,
			iter.layoutWidth, ext.logicalRect.Y, nil, inkRect, nil)
	}

	if logicalRect != nil {
		*logicalRect = ext.logicalRect
	}
}

// GetBaseline gets the Y position of the current line's baseline, in layout
// coordinates (origin at top left of the entire layout).
func (iter *LayoutIter) GetBaseline() Unit {
	if iter.lineIndex >= len(iter.lineExtents) {
		return 0
	}

	return iter.lineExtents[iter.lineIndex].baseline
}

// IsAtLastLine determines whether `iter` is on the last line of the layout.
func (iter *LayoutIter) IsAtLastLine() bool {
	return iter.lineIndex == len(iter.layout.lines)-1
}
