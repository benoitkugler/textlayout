package psinterpreter

import (
	"errors"
	"fmt"
)

// PathBounds represents a control bounds for
// a glyph outline (in font units).
type PathBounds struct {
	XMin, YMin, XMax, YMax int32
}

// Update enlarges the current bounds to include the Point (x,y).
func (p *PathBounds) Update(pt Point) {
	if pt.x < p.XMin {
		p.XMin = pt.x
	}
	if pt.x > p.XMax {
		p.XMax = pt.x
	}
	if pt.y < p.YMin {
		p.YMin = pt.y
	}
	if pt.y > p.YMax {
		p.YMax = pt.y
	}
}

// Point is a 2D Point in font units.
type Point struct{ x, y int32 }

// Move translates the Point.
func (p *Point) Move(dx, dy int32) {
	p.x += dx
	p.y += dy
}

// CharstringReader provides implementation
// of the operators found in a font charstring.
type CharstringReader struct {
	Bounds PathBounds

	vstemCount   int32
	hstemCount   int32
	hintmaskSize int32

	currentPoint Point
	isPathOpen   bool

	seenHintmask bool
}

func (out *CharstringReader) Hstem(state *Machine) {
	out.hstemCount += state.ArgStack.Top / 2
}

func (out *CharstringReader) Vstem(state *Machine) {
	out.vstemCount += state.ArgStack.Top / 2
}

func (out *CharstringReader) determineHintmaskSize(state *Machine) {
	if !out.seenHintmask {
		out.vstemCount += state.ArgStack.Top / 2
		out.hintmaskSize = (out.hstemCount + out.vstemCount + 7) >> 3
		out.seenHintmask = true
	}
}

func (out *CharstringReader) Hintmask(state *Machine) {
	out.determineHintmaskSize(state)
	state.SkipBytes(out.hintmaskSize)
}

func (out *CharstringReader) line(pt Point) {
	if !out.isPathOpen {
		out.isPathOpen = true
		out.Bounds.Update(out.currentPoint)
	}
	out.currentPoint = pt
	out.Bounds.Update(pt)
}

func (out *CharstringReader) curve(pt1, pt2, pt3 Point) {
	if !out.isPathOpen {
		out.isPathOpen = true
		out.Bounds.Update(out.currentPoint)
	}
	/* include control Points */
	out.Bounds.Update(pt1)
	out.Bounds.Update(pt2)
	out.currentPoint = pt3
	out.Bounds.Update(pt3)
}

func (out *CharstringReader) doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6 Point) {
	out.curve(pt1, pt2, pt3)
	out.curve(pt4, pt5, pt6)
}

func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// ------------------------------------------------------------

// LocalSubr pops the subroutine index and call it
func LocalSubr(state *Machine) error {
	if state.ArgStack.Top < 1 {
		return errors.New("invalid callsubr operator (empty stack)")
	}
	index := state.ArgStack.Pop()
	return state.CallSubroutine(index, true)
}

// GlobalSubr pops the subroutine index and call it
func GlobalSubr(state *Machine) error {
	if state.ArgStack.Top < 1 {
		return errors.New("invalid callgsubr operator (empty stack)")
	}
	index := state.ArgStack.Pop()
	return state.CallSubroutine(index, false)
}

func (out *CharstringReader) ClosePath() { out.isPathOpen = false }

func (out *CharstringReader) Rmoveto(state *Machine) error {
	if state.ArgStack.Top < 2 {
		return errors.New("invalid rmoveto operator")
	}
	out.currentPoint.x += state.ArgStack.Pop()
	out.currentPoint.y += state.ArgStack.Pop()
	out.isPathOpen = false
	return nil
}

func (out *CharstringReader) Vmoveto(state *Machine) error {
	if state.ArgStack.Top < 1 {
		return errors.New("invalid vmoveto operator")
	}
	out.currentPoint.y += state.ArgStack.Pop()
	out.isPathOpen = false
	return nil
}

func (out *CharstringReader) Hmoveto(state *Machine) error {
	if state.ArgStack.Top < 1 {
		return errors.New("invalid hmoveto operator")
	}
	out.currentPoint.x += state.ArgStack.Pop()
	out.isPathOpen = false
	return nil
}

func (out *CharstringReader) Rlineto(state *Machine) {
	for i := int32(0); i+2 <= state.ArgStack.Top; i += 2 {
		newPoint := out.currentPoint
		newPoint.Move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		out.line(newPoint)
	}
	state.ArgStack.Clear()
}

func (out *CharstringReader) Hlineto(state *Machine) {
	var i int32
	for ; i+2 <= state.ArgStack.Top; i += 2 {
		newPoint := out.currentPoint
		newPoint.x += state.ArgStack.Vals[i]
		out.line(newPoint)
		newPoint.y += state.ArgStack.Vals[i+1]
		out.line(newPoint)
	}
	if i < state.ArgStack.Top {
		newPoint := out.currentPoint
		newPoint.x += state.ArgStack.Vals[i]
		out.line(newPoint)
	}
}

func (out *CharstringReader) Vlineto(state *Machine) {
	var i int32
	for ; i+2 <= state.ArgStack.Top; i += 2 {
		newPoint := out.currentPoint
		newPoint.y += state.ArgStack.Vals[i]
		out.line(newPoint)
		newPoint.x += state.ArgStack.Vals[i+1]
		out.line(newPoint)
	}
	if i < state.ArgStack.Top {
		newPoint := out.currentPoint
		newPoint.y += state.ArgStack.Vals[i]
		out.line(newPoint)
	}
}

func (out *CharstringReader) Rrcurveto(state *Machine) {
	for i := int32(0); i+6 <= state.ArgStack.Top; i += 6 {
		pt1 := out.currentPoint
		pt1.Move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		pt2 := pt1
		pt2.Move(state.ArgStack.Vals[i+2], state.ArgStack.Vals[i+3])
		pt3 := pt2
		pt3.Move(state.ArgStack.Vals[i+4], state.ArgStack.Vals[i+5])
		out.curve(pt1, pt2, pt3)
	}
}

func (out *CharstringReader) Hhcurveto(state *Machine) {
	var (
		i   int32
		pt1 = out.currentPoint
	)
	if (state.ArgStack.Top & 1) != 0 {
		pt1.y += (state.ArgStack.Vals[i])
		i++
	}
	for ; i+4 <= state.ArgStack.Top; i += 4 {
		pt1.x += state.ArgStack.Vals[i]
		pt2 := pt1
		pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.x += state.ArgStack.Vals[i+3]
		out.curve(pt1, pt2, pt3)
		pt1 = out.currentPoint
	}
}

func (out *CharstringReader) Vhcurveto(state *Machine) {
	var i int32
	if (state.ArgStack.Top % 8) >= 4 {
		pt1 := out.currentPoint
		pt1.y += state.ArgStack.Vals[i]
		pt2 := pt1
		pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.x += state.ArgStack.Vals[i+3]
		i += 4

		for ; i+8 <= state.ArgStack.Top; i += 8 {
			out.curve(pt1, pt2, pt3)
			pt1 = out.currentPoint
			pt1.x += (state.ArgStack.Vals[i])
			pt2 = pt1
			pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 = pt2
			pt3.y += (state.ArgStack.Vals[i+3])
			out.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.y += (state.ArgStack.Vals[i+4])
			pt2 = pt1
			pt2.Move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.x += (state.ArgStack.Vals[i+7])
		}
		if i < state.ArgStack.Top {
			pt3.y += (state.ArgStack.Vals[i])
		}
		out.curve(pt1, pt2, pt3)
	} else {
		for ; i+8 <= state.ArgStack.Top; i += 8 {
			pt1 := out.currentPoint
			pt1.y += (state.ArgStack.Vals[i])
			pt2 := pt1
			pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 := pt2
			pt3.x += (state.ArgStack.Vals[i+3])
			out.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.x += (state.ArgStack.Vals[i+4])
			pt2 = pt1
			pt2.Move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.y += (state.ArgStack.Vals[i+7])
			if (state.ArgStack.Top-i < 16) && ((state.ArgStack.Top & 1) != 0) {
				pt3.x += (state.ArgStack.Vals[i+8])
			}
			out.curve(pt1, pt2, pt3)
		}
	}
}

func (out *CharstringReader) Hvcurveto(state *Machine) {
	//    pt1,: pt2, pt3;
	var i int32
	if (state.ArgStack.Top % 8) >= 4 {
		pt1 := out.currentPoint
		pt1.x += (state.ArgStack.Vals[i])
		pt2 := pt1
		pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.y += (state.ArgStack.Vals[i+3])
		i += 4

		for ; i+8 <= state.ArgStack.Top; i += 8 {
			out.curve(pt1, pt2, pt3)
			pt1 = out.currentPoint
			pt1.y += (state.ArgStack.Vals[i])
			pt2 = pt1
			pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 = pt2
			pt3.x += (state.ArgStack.Vals[i+3])
			out.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.x += state.ArgStack.Vals[i+4]
			pt2 = pt1
			pt2.Move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.y += state.ArgStack.Vals[i+7]
		}
		if i < state.ArgStack.Top {
			pt3.x += (state.ArgStack.Vals[i])
		}
		out.curve(pt1, pt2, pt3)
	} else {
		for ; i+8 <= state.ArgStack.Top; i += 8 {
			pt1 := out.currentPoint
			pt1.x += (state.ArgStack.Vals[i])
			pt2 := pt1
			pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 := pt2
			pt3.y += (state.ArgStack.Vals[i+3])
			out.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.y += (state.ArgStack.Vals[i+4])
			pt2 = pt1
			pt2.Move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.x += (state.ArgStack.Vals[i+7])
			if (state.ArgStack.Top-i < 16) && ((state.ArgStack.Top & 1) != 0) {
				pt3.y += state.ArgStack.Vals[i+8]
			}
			out.curve(pt1, pt2, pt3)
		}
	}
}

func (out *CharstringReader) Rcurveline(state *Machine) error {
	argCount := state.ArgStack.Top
	if argCount < 8 {
		return fmt.Errorf("expected at least 8 operands for <rcurveline>, got %d", argCount)
	}

	var i int32
	curveLimit := argCount - 2
	for ; i+6 <= curveLimit; i += 6 {
		pt1 := out.currentPoint
		pt1.Move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		pt2 := pt1
		pt2.Move(state.ArgStack.Vals[i+2], state.ArgStack.Vals[i+3])
		pt3 := pt2
		pt3.Move(state.ArgStack.Vals[i+4], state.ArgStack.Vals[i+5])
		out.curve(pt1, pt2, pt3)
	}

	pt1 := out.currentPoint
	pt1.Move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
	out.line(pt1)

	return nil
}

func (out *CharstringReader) Rlinecurve(state *Machine) error {
	argCount := state.ArgStack.Top
	if argCount < 8 {
		return fmt.Errorf("expected at least 8 operands for <rlinecurve>, got %d", argCount)
	}
	var i int32
	lineLimit := argCount - 6
	for ; i+2 <= lineLimit; i += 2 {
		pt1 := out.currentPoint
		pt1.Move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		out.line(pt1)
	}

	pt1 := out.currentPoint
	pt1.Move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
	pt2 := pt1
	pt2.Move(state.ArgStack.Vals[i+2], state.ArgStack.Vals[i+3])
	pt3 := pt2
	pt3.Move(state.ArgStack.Vals[i+4], state.ArgStack.Vals[i+5])
	out.curve(pt1, pt2, pt3)

	return nil
}

func (out *CharstringReader) Vvcurveto(state *Machine) {
	var i int32
	pt1 := out.currentPoint
	if (state.ArgStack.Top & 1) != 0 {
		pt1.x += state.ArgStack.Vals[i]
		i++
	}
	for ; i+4 <= state.ArgStack.Top; i += 4 {
		pt1.y += state.ArgStack.Vals[i]
		pt2 := pt1
		pt2.Move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.y += state.ArgStack.Vals[i+3]
		out.curve(pt1, pt2, pt3)
		pt1 = out.currentPoint
	}
}

func (out *CharstringReader) Hflex(state *Machine) error {
	if state.ArgStack.Top != 7 {
		return fmt.Errorf("expected 7 operands for <hflex>, got %d", state.ArgStack.Top)
	}

	pt1 := out.currentPoint
	pt1.x += state.ArgStack.Vals[0]
	pt2 := pt1
	pt2.Move(state.ArgStack.Vals[1], state.ArgStack.Vals[2])
	pt3 := pt2
	pt3.x += state.ArgStack.Vals[3]
	pt4 := pt3
	pt4.x += state.ArgStack.Vals[4]
	pt5 := pt4
	pt5.x += state.ArgStack.Vals[5]
	pt5.y = pt1.y
	pt6 := pt5
	pt6.x += state.ArgStack.Vals[6]

	out.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}

func (out *CharstringReader) Flex(state *Machine) error {
	if state.ArgStack.Top != 13 {
		return fmt.Errorf("expected 13 operands for <flex>, got %d", state.ArgStack.Top)
	}

	pt1 := out.currentPoint
	pt1.Move(state.ArgStack.Vals[0], state.ArgStack.Vals[1])
	pt2 := pt1
	pt2.Move(state.ArgStack.Vals[2], state.ArgStack.Vals[3])
	pt3 := pt2
	pt3.Move(state.ArgStack.Vals[4], state.ArgStack.Vals[5])
	pt4 := pt3
	pt4.Move(state.ArgStack.Vals[6], state.ArgStack.Vals[7])
	pt5 := pt4
	pt5.Move(state.ArgStack.Vals[8], state.ArgStack.Vals[9])
	pt6 := pt5
	pt6.Move(state.ArgStack.Vals[10], state.ArgStack.Vals[11])

	out.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}

func (out *CharstringReader) Hflex1(state *Machine) error {
	if state.ArgStack.Top != 9 {
		return fmt.Errorf("expected 9 operands for <hflex1>, got %d", state.ArgStack.Top)
	}
	pt1 := out.currentPoint
	pt1.Move(state.ArgStack.Vals[0], state.ArgStack.Vals[1])
	pt2 := pt1
	pt2.Move(state.ArgStack.Vals[2], state.ArgStack.Vals[3])
	pt3 := pt2
	pt3.x += state.ArgStack.Vals[4]
	pt4 := pt3
	pt4.x += state.ArgStack.Vals[5]
	pt5 := pt4
	pt5.Move(state.ArgStack.Vals[6], state.ArgStack.Vals[7])
	pt6 := pt5
	pt6.x += state.ArgStack.Vals[8]
	pt6.y = out.currentPoint.y

	out.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}

func (out *CharstringReader) Flex1(state *Machine) error {
	if state.ArgStack.Top != 11 {
		return fmt.Errorf("expected 11 operands for <flex1>, got %d", state.ArgStack.Top)
	}

	var d Point
	for i := 0; i < 10; i += 2 {
		d.Move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
	}

	pt1 := out.currentPoint
	pt1.Move(state.ArgStack.Vals[0], state.ArgStack.Vals[1])
	pt2 := pt1
	pt2.Move(state.ArgStack.Vals[2], state.ArgStack.Vals[3])
	pt3 := pt2
	pt3.Move(state.ArgStack.Vals[4], state.ArgStack.Vals[5])
	pt4 := pt3
	pt4.Move(state.ArgStack.Vals[6], state.ArgStack.Vals[7])
	pt5 := pt4
	pt5.Move(state.ArgStack.Vals[8], state.ArgStack.Vals[9])
	pt6 := pt5

	if abs(d.x) > abs(d.y) {
		pt6.x += state.ArgStack.Vals[10]
		pt6.y = out.currentPoint.y
	} else {
		pt6.x = out.currentPoint.x
		pt6.y += state.ArgStack.Vals[10]
	}

	out.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}
