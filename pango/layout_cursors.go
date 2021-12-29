package pango

import (
	"sort"

	"github.com/benoitkugler/textlayout/fribidi"
)

type CursorPos struct {
	pos int
	x   Unit
}

//   static int
//   compare_cursor (gconstpointer v1,
// 				  gconstpointer v2)
//   {
// 	const CursorPos *c1 = v1;
// 	const CursorPos *c2 = v2;

// 	return c1.x - c2.x;
//   }

// GetCursorPos determines, given an index within a layout, the positions that of the
// strong and weak cursors if the insertion point is at that index.
//
// The position of each cursor is stored as a zero-width rectangle
// with the height of the run extents.
//
// The strong cursor location is the location where characters of the
// directionality equal to the base direction of the layout are inserted.
// The weak cursor location is the location where characters of the
// directionality opposite to the base direction of the layout are inserted.
//
// The strong cursor has a little arrow pointing to the right, the weak
// cursor to the left. Typing a 'c' in this situation will insert the
// character after the 'b', and typing another Hebrew character, like 'ג',
// will insert it at the end.
func (layout *Layout) GetCursorPos(index int, strongPos, weakPos *Rectangle) {
	if index < 0 || index > len(layout.Text) {
		return
	}

	line, lineRect, runRect := layout.indexToLineAndExtents(index)

	if debugMode {
		assert(index >= line.StartIndex, "getCursorPos")
	}

	var (
		dir1, dir2     Direction
		level1, level2 fribidi.Level
		x1Trailing, x2 Unit
	)

	/* Examine the trailing edge of the character before the cursor */
	if index == line.StartIndex {
		dir1 = line.ResolvedDir
		level1 = 1
		if dir1 == DIRECTION_LTR {
			level1 = 0
		}
		if line.ResolvedDir == DIRECTION_LTR {
			x1Trailing = 0
		} else {
			x1Trailing = Unit(lineRect.Width)
		}
	} else {
		prevIndex := index - 1
		level1 = line.getCharLevel(prevIndex)
		dir1 = DIRECTION_LTR
		if level1%2 != 0 {
			dir1 = DIRECTION_RTL
		}
		x1Trailing = line.IndexToX(prevIndex, true)
	}

	/* Examine the leading edge of the character after the cursor */
	if index >= line.StartIndex+line.Length {
		dir2 = line.ResolvedDir
		level2 = 1
		if dir2 == DIRECTION_LTR {
			level2 = 0
		}
		if line.ResolvedDir == DIRECTION_LTR {
			x2 = Unit(lineRect.Width)
		} else {
			x2 = 0
		}
	} else {
		x2 = line.IndexToX(index, false)
		level2 = line.getCharLevel(index)

		dir2 = DIRECTION_LTR
		if level2%2 != 0 {
			dir2 = DIRECTION_RTL
		}
	}

	if strongPos != nil {
		strongPos.X = lineRect.X

		if dir1 == line.ResolvedDir && (dir2 != dir1 || level1 < level2) {
			strongPos.X += x1Trailing
		} else {
			strongPos.X += x2
		}

		strongPos.Y = runRect.Y
		strongPos.Width = 0
		strongPos.Height = runRect.Height
	}

	if weakPos != nil {
		weakPos.X = lineRect.X

		if dir1 == line.ResolvedDir && (dir2 != dir1 || level1 < level2) {
			weakPos.X += x2
		} else {
			weakPos.X += x1Trailing
		}

		weakPos.Y = runRect.Y
		weakPos.Width = 0
		weakPos.Height = runRect.Height
	}
}

func (line *LayoutLine) getCursors(strong bool) []CursorPos {
	layout := line.layout

	lineNo, _ := layout.indexToLineX(line.StartIndex+line.Length, false)

	end := line.StartIndex + line.Length
	line2 := layout.GetLine(lineNo)
	if line2 == line {
		end++
	}

	var pos Rectangle
	arg1, arg2 := (*Rectangle)(nil), &pos
	if strong {
		arg1, arg2 = &pos, (*Rectangle)(nil)
	}
	var cursors []CursorPos
	for j := line.StartIndex; j < end; j++ {
		if layout.logAttrs[j].IsCursorPosition() {
			layout.GetCursorPos(j, arg1, arg2)

			cursors = append(cursors, CursorPos{
				pos: j,
				x:   pos.X,
			})
		}
	}

	sort.Slice(cursors, func(i, j int) bool { return cursors[i].x < cursors[j].x })

	return cursors
}

// MoveCursorVisually computes a new cursor position from an old position (`oldIndex`) and a direction.
//
// `strong` indicates whether the moving cursor is the strong cursor or the
//   weak cursor. The strong cursor is the cursor corresponding
//   to text insertion in the base direction for the layout.
// if `oldTrailing` is 0, the cursor was at the leading edge of the
//   grapheme indicated by `oldIndex`, if > 0, the cursor
//   was at the trailing edge.
// If `direction` is positive, then the new position will cause the strong
// or weak cursor to be displayed one position to right of where it was
// with the old cursor position. If `direction` is negative, it will be
// moved to the left.
//
// A value of -1 for `newIndex` indicates that the cursor has been moved off the
// beginning of the layout. A value of `MaxInt` indicates that
// the cursor has been moved off the end of the layout.
// `newTrailing` is the number of characters to move forward from
// the location returned for `newIndex` to get the position where
// the cursor should be displayed. This allows distinguishing the
// position at the beginning of one line from the position at the
// end of the preceding line. `newIndex` is always on the line where
// the cursor should be displayed.
//
// In the presence of bidirectional text, the correspondence between
// logical and visual order will depend on the direction of the current
// run, and there may be jumps when the cursor is moved off of the end
// of a run.
//
// Motion here is in cursor positions, not in characters, so a single
// call to this function may move the cursor over multiple characters
// when multiple characters combine to form a single grapheme.
func (layout *Layout) MoveCursorVisually(strong bool,
	oldIndex, oldTrailing, direction int) (newIndex, newTrailing int) {

	if oldIndex < 0 || oldIndex > len(layout.Text) {
		return
	}
	if oldIndex == len(layout.Text) && oldTrailing != 0 {
		return
	}

	if direction >= 0 {
		direction = 1
	} else {
		direction = -1
	}

	layout.checkLines()

	/* Find the line the old cursor is on */
	_, line, prevLine, nextLine := layout.indexToLine(oldIndex)

	oldIndex += oldTrailing

	/* Clamp old_index to fit on the line */
	if oldIndex > (line.StartIndex + line.Length) {
		oldIndex = line.StartIndex + line.Length
	}

	cursors := line.getCursors(strong)

	var pos Rectangle
	arg1, arg2 := (*Rectangle)(nil), &pos
	if strong {
		arg1, arg2 = &pos, (*Rectangle)(nil)
	}
	layout.GetCursorPos(oldIndex, arg1, arg2)

	visPos := -1
	for j, cursor := range cursors {
		if cursor.x == pos.X {
			visPos = j

			/* If moving left, we pick the leftmost match, otherwise
			 * the rightmost one. Without this, we can get stuck
			 */
			if direction < 0 {
				break
			}
		}
	}

	if visPos == -1 &&
		oldIndex == line.StartIndex+line.Length {
		if line.ResolvedDir == DIRECTION_LTR {
			visPos = len(cursors)
		} else {
			visPos = 0
		}
	}

	/* Handling movement between lines */
	var offStart, offEnd bool
	if line.ResolvedDir == DIRECTION_LTR {
		if oldIndex == line.StartIndex && direction < 0 {
			offStart = true
		}
		if oldIndex == line.StartIndex+line.Length && direction > 0 {
			offEnd = true
		}
	} else {
		if oldIndex == line.StartIndex+line.Length && direction < 0 {
			offStart = true
		}
		if oldIndex == line.StartIndex && direction > 0 {
			offEnd = true
		}
	}

	if offStart || offEnd {
		// If we move over a paragraph boundary, count that as
		// an extra position in the motion
		var paragraphBoundary bool

		if offStart {
			if prevLine == nil {
				return -1, 0
			}
			line = prevLine
			paragraphBoundary = (line.StartIndex+line.Length != oldIndex)
		} else {
			if nextLine == nil {
				return MaxInt, 0
			}
			line = nextLine
			paragraphBoundary = (line.StartIndex != oldIndex)
		}

		cursors = line.getCursors(strong)

		nVis := len(cursors)

		if offStart && direction < 0 {
			visPos = nVis
			if paragraphBoundary {
				visPos++
			}
		} else if offEnd && direction > 0 {
			visPos = 0
			if paragraphBoundary {
				visPos--
			}
		}
	}

	if direction < 0 {
		visPos--
	} else {
		visPos++
	}

	if 0 <= visPos && visPos < len(cursors) {
		newIndex = cursors[visPos].pos
	} else if visPos >= len(cursors)-1 {
		newIndex = line.StartIndex + line.Length
	}

	newTrailing = 0

	if newIndex == line.StartIndex+line.Length && line.Length > 0 {
		logPos := line.StartIndex + line.Length
		for do := true; do; do = logPos > line.StartIndex && !layout.logAttrs[logPos].IsCursorPosition() {
			logPos--
			newIndex -= 1
			newTrailing++
		}
	}

	return newIndex, newTrailing
}
