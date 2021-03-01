package type1c

import (
	"errors"
	"fmt"

	ps "github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

type pathBounds struct {
	xMin, yMin, xMax, yMax int32
}

// grow the current bounds to include the point (x,y)
func (p *pathBounds) update(pt point) {
	if pt.x < p.xMin {
		p.xMin = pt.x
	}
	if pt.x > p.xMax {
		p.xMax = pt.x
	}
	if pt.y < p.yMin {
		p.yMin = pt.y
	}
	if pt.y > p.yMax {
		p.yMax = pt.y
	}
}

type point struct{ x, y int32 }

func (p *point) move(dx, dy int32) {
	p.x += dx
	p.y += dy
}

// type2CharstringHandler implements operators needed to fetch Type2 charstring metrics
type type2CharstringHandler struct {
	// found in private DICT, needed since we can't differenciate
	// no width set from 0 width
	// `width` must be initialized to default width
	nominalWidthX int32

	width int32

	currentPoint           point
	bounds                 pathBounds
	isPathOpen             bool
	vstemCount, hstemCount int32
}

func (type2CharstringHandler) Context() ps.PsContext { return ps.Type2Charstring }

func (met *type2CharstringHandler) Run(op ps.PsOperator, state *ps.Inter) error {
	var err error
	if !op.IsEscaped {
		switch op.Operator {
		case 21: // rmoveto
			err = met.rmoveto(state)
		case 22: // hmoveto
			err = met.hmoveto(state)
		case 4: // vmoveto
			err = met.vmoveto(state)
		case 1, 18: // hstem, hstemhm
			met.hstem(state)
		case 3, 23: // vstem, vstemhm
			met.vstem(state)
		case 19, 20: // hintmask, cntrmask
			// variable number of arguments, but always even
			// for xxxmask, if there are arguments on the stack, then this is an impliied stem
			if state.ArgStack.Top&1 != 0 {
				met.width = met.nominalWidthX + state.ArgStack.Vals[0]
			}
		case 14: // endchar
			if state.ArgStack.Top > 0 { // width is optional
				met.width = met.nominalWidthX + state.ArgStack.Vals[0]
			}
		case 5: // rlineto
			met.rlineto(state)
		case 6: // hlineto
			met.hlineto(state)
		case 7: // vlineto
			met.vlineto(state)
		case 8: // rrcurveto
			met.rrcurveto(state)
		case 24: // rcurveline
			err = met.rcurveline(state)
		case 25: // rlinecurve
			err = met.rlinecurve(state)
		case 26: // vvcurveto
			met.vvcurveto(state)
		case 27: // hhcurveto
			met.hhcurveto(state)
		case 30: // vhcurveto
			met.vhcurveto(state)
		case 31: // hvcurveto
			met.hvcurveto(state)
		default:
			// no other operands are allowed before the ones handled above
			err = fmt.Errorf("invalid operator %s in charstring", op)
		}
	} else {
		switch op.Operator {
		case 34: // hflex
			err = met.hflex(state)
		case 35: // flex
			err = met.flex(state)
		case 36: // hflex1
			err = met.hflex1(state)
		case 37: // flex1
			err = met.flex1(state)
		default:
			// no other operands are allowed before the ones handled above
			err = fmt.Errorf("invalid operator %s in charstring", op)
		}
	}
	state.ArgStack.PopAll()
	return err
}

// --------------------- shared helpers ---------------------

func (met *type2CharstringHandler) line(pt point) {
	if !met.isPathOpen {
		met.isPathOpen = true
		met.bounds.update(met.currentPoint)
	}
	met.currentPoint = pt
	met.bounds.update(pt)
}

func (met *type2CharstringHandler) curve(pt1, pt2, pt3 point) {
	if !met.isPathOpen {
		met.isPathOpen = true
		met.bounds.update(met.currentPoint)
	}
	/* include control points */
	met.bounds.update(pt1)
	met.bounds.update(pt2)
	met.currentPoint = pt3
	met.bounds.update(pt3)
}

func (met *type2CharstringHandler) doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6 point) {
	met.curve(pt1, pt2, pt3)
	met.curve(pt4, pt5, pt6)
}

func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// ------------------------------------------------------------

func (met *type2CharstringHandler) rmoveto(state *ps.Inter) error {
	if state.ArgStack.Top > 2 { // width is optional
		met.width = met.nominalWidthX + state.ArgStack.Vals[0]
	}
	if state.ArgStack.Top < 2 {
		return errors.New("invalid rmoveto operator")
	}
	met.currentPoint.x += state.ArgStack.Pop()
	met.currentPoint.y += state.ArgStack.Pop()
	met.isPathOpen = false
	return nil
}

func (met *type2CharstringHandler) vmoveto(state *ps.Inter) error {
	if state.ArgStack.Top > 1 { // width is optional
		met.width = met.nominalWidthX + state.ArgStack.Vals[0]
	}
	if state.ArgStack.Top < 1 {
		return errors.New("invalid vmoveto operator")
	}
	met.currentPoint.y += state.ArgStack.Pop()
	met.isPathOpen = false
	return nil
}

func (met *type2CharstringHandler) hmoveto(state *ps.Inter) error {
	if state.ArgStack.Top > 1 { // width is optional
		met.width = met.nominalWidthX + state.ArgStack.Vals[0]
	}
	if state.ArgStack.Top < 1 {
		return errors.New("invalid hmoveto operator")
	}
	met.currentPoint.x += state.ArgStack.Pop()
	met.isPathOpen = false
	return nil
}

func (met *type2CharstringHandler) rlineto(state *ps.Inter) {
	for i := int32(0); i+2 <= state.ArgStack.Top; i += 2 {
		newPoint := met.currentPoint
		newPoint.move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		met.line(newPoint)
	}
	state.ArgStack.PopAll()
}

func (met *type2CharstringHandler) hlineto(state *ps.Inter) {
	var i int32
	for ; i+2 <= state.ArgStack.Top; i += 2 {
		newPoint := met.currentPoint
		newPoint.x += state.ArgStack.Vals[i]
		met.line(newPoint)
		newPoint.y += state.ArgStack.Vals[i+1]
		met.line(newPoint)
	}
	if i < state.ArgStack.Top {
		newPoint := met.currentPoint
		newPoint.x += state.ArgStack.Vals[i]
		met.line(newPoint)
	}
}

func (met *type2CharstringHandler) vlineto(state *ps.Inter) {
	var i int32
	for ; i+2 <= state.ArgStack.Top; i += 2 {
		newPoint := met.currentPoint
		newPoint.y += state.ArgStack.Vals[i]
		met.line(newPoint)
		newPoint.x += state.ArgStack.Vals[i+1]
		met.line(newPoint)
	}
	if i < state.ArgStack.Top {
		newPoint := met.currentPoint
		newPoint.y += state.ArgStack.Vals[i]
		met.line(newPoint)
	}
}

func (met *type2CharstringHandler) rrcurveto(state *ps.Inter) {
	for i := int32(0); i+6 <= state.ArgStack.Top; i += 6 {
		pt1 := met.currentPoint
		pt1.move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		pt2 := pt1
		pt2.move(state.ArgStack.Vals[i+2], state.ArgStack.Vals[i+3])
		pt3 := pt2
		pt3.move(state.ArgStack.Vals[i+4], state.ArgStack.Vals[i+5])
		met.curve(pt1, pt2, pt3)
	}
}

func (met *type2CharstringHandler) hhcurveto(state *ps.Inter) {
	var (
		i   int32
		pt1 = met.currentPoint
	)
	if (state.ArgStack.Top & 1) != 0 {
		pt1.y += (state.ArgStack.Vals[i])
		i++
	}
	for ; i+4 <= state.ArgStack.Top; i += 4 {
		pt1.x += state.ArgStack.Vals[i]
		pt2 := pt1
		pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.x += state.ArgStack.Vals[i+3]
		met.curve(pt1, pt2, pt3)
		pt1 = met.currentPoint
	}
}

func (met *type2CharstringHandler) vhcurveto(state *ps.Inter) {
	//    pt1,: pt2, pt3;
	var i int32
	if (state.ArgStack.Top % 8) >= 4 {
		pt1 := met.currentPoint
		pt1.y += state.ArgStack.Vals[i]
		pt2 := pt1
		pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.x += state.ArgStack.Vals[i+3]
		i += 4

		for ; i+8 <= state.ArgStack.Top; i += 8 {
			met.curve(pt1, pt2, pt3)
			pt1 = met.currentPoint
			pt1.x += (state.ArgStack.Vals[i])
			pt2 = pt1
			pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 = pt2
			pt3.y += (state.ArgStack.Vals[i+3])
			met.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.y += (state.ArgStack.Vals[i+4])
			pt2 = pt1
			pt2.move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.x += (state.ArgStack.Vals[i+7])
		}
		if i < state.ArgStack.Top {
			pt3.y += (state.ArgStack.Vals[i])
		}
		met.curve(pt1, pt2, pt3)
	} else {
		for ; i+8 <= state.ArgStack.Top; i += 8 {
			pt1 := met.currentPoint
			pt1.y += (state.ArgStack.Vals[i])
			pt2 := pt1
			pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 := pt2
			pt3.x += (state.ArgStack.Vals[i+3])
			met.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.x += (state.ArgStack.Vals[i+4])
			pt2 = pt1
			pt2.move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.y += (state.ArgStack.Vals[i+7])
			if (state.ArgStack.Top-i < 16) && ((state.ArgStack.Top & 1) != 0) {
				pt3.x += (state.ArgStack.Vals[i+8])
			}
			met.curve(pt1, pt2, pt3)
		}
	}
}

func (met *type2CharstringHandler) hvcurveto(state *ps.Inter) {
	//    pt1,: pt2, pt3;
	var i int32
	if (state.ArgStack.Top % 8) >= 4 {
		pt1 := met.currentPoint
		pt1.x += (state.ArgStack.Vals[i])
		pt2 := pt1
		pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.y += (state.ArgStack.Vals[i+3])
		i += 4

		for ; i+8 <= state.ArgStack.Top; i += 8 {
			met.curve(pt1, pt2, pt3)
			pt1 = met.currentPoint
			pt1.y += (state.ArgStack.Vals[i])
			pt2 = pt1
			pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 = pt2
			pt3.x += (state.ArgStack.Vals[i+3])
			met.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.x += state.ArgStack.Vals[i+4]
			pt2 = pt1
			pt2.move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.y += state.ArgStack.Vals[i+7]
		}
		if i < state.ArgStack.Top {
			pt3.x += (state.ArgStack.Vals[i])
		}
		met.curve(pt1, pt2, pt3)
	} else {
		for ; i+8 <= state.ArgStack.Top; i += 8 {
			pt1 := met.currentPoint
			pt1.x += (state.ArgStack.Vals[i])
			pt2 := pt1
			pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
			pt3 := pt2
			pt3.y += (state.ArgStack.Vals[i+3])
			met.curve(pt1, pt2, pt3)

			pt1 = pt3
			pt1.y += (state.ArgStack.Vals[i+4])
			pt2 = pt1
			pt2.move(state.ArgStack.Vals[i+5], state.ArgStack.Vals[i+6])
			pt3 = pt2
			pt3.x += (state.ArgStack.Vals[i+7])
			if (state.ArgStack.Top-i < 16) && ((state.ArgStack.Top & 1) != 0) {
				pt3.y += state.ArgStack.Vals[i+8]
			}
			met.curve(pt1, pt2, pt3)
		}
	}
}

func (met *type2CharstringHandler) rcurveline(state *ps.Inter) error {
	argCount := state.ArgStack.Top
	if argCount < 8 {
		return fmt.Errorf("expected at least 8 operands for <rcurveline>, got %d", argCount)
	}

	var i int32
	curveLimit := argCount - 2
	for ; i+6 <= curveLimit; i += 6 {
		pt1 := met.currentPoint
		pt1.move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		pt2 := pt1
		pt2.move(state.ArgStack.Vals[i+2], state.ArgStack.Vals[i+3])
		pt3 := pt2
		pt3.move(state.ArgStack.Vals[i+4], state.ArgStack.Vals[i+5])
		met.curve(pt1, pt2, pt3)
	}

	pt1 := met.currentPoint
	pt1.move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
	met.line(pt1)

	return nil
}

func (met *type2CharstringHandler) rlinecurve(state *ps.Inter) error {
	argCount := state.ArgStack.Top
	if argCount < 8 {
		return fmt.Errorf("expected at least 8 operands for <rlinecurve>, got %d", argCount)
	}
	var i int32
	lineLimit := argCount - 6
	for ; i+2 <= lineLimit; i += 2 {
		pt1 := met.currentPoint
		pt1.move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
		met.line(pt1)
	}

	pt1 := met.currentPoint
	pt1.move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
	pt2 := pt1
	pt2.move(state.ArgStack.Vals[i+2], state.ArgStack.Vals[i+3])
	pt3 := pt2
	pt3.move(state.ArgStack.Vals[i+4], state.ArgStack.Vals[i+5])
	met.curve(pt1, pt2, pt3)

	return nil
}

func (met *type2CharstringHandler) vvcurveto(state *ps.Inter) {
	var i int32
	pt1 := met.currentPoint
	if (state.ArgStack.Top & 1) != 0 {
		pt1.x += state.ArgStack.Vals[i]
		i++
	}
	for ; i+4 <= state.ArgStack.Top; i += 4 {
		pt1.y += state.ArgStack.Vals[i]
		pt2 := pt1
		pt2.move(state.ArgStack.Vals[i+1], state.ArgStack.Vals[i+2])
		pt3 := pt2
		pt3.y += state.ArgStack.Vals[i+3]
		met.curve(pt1, pt2, pt3)
		pt1 = met.currentPoint
	}
}

func (met *type2CharstringHandler) hflex(state *ps.Inter) error {
	if state.ArgStack.Top != 7 {
		return fmt.Errorf("expected 7 operands for <hflex>, got %d", state.ArgStack.Top)
	}

	pt1 := met.currentPoint
	pt1.x += state.ArgStack.Vals[0]
	pt2 := pt1
	pt2.move(state.ArgStack.Vals[1], state.ArgStack.Vals[2])
	pt3 := pt2
	pt3.x += state.ArgStack.Vals[3]
	pt4 := pt3
	pt4.x += state.ArgStack.Vals[4]
	pt5 := pt4
	pt5.x += state.ArgStack.Vals[5]
	pt5.y = pt1.y
	pt6 := pt5
	pt6.x += state.ArgStack.Vals[6]

	met.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}

func (met *type2CharstringHandler) flex(state *ps.Inter) error {
	if state.ArgStack.Top != 13 {
		return fmt.Errorf("expected 13 operands for <flex>, got %d", state.ArgStack.Top)
	}

	pt1 := met.currentPoint
	pt1.move(state.ArgStack.Vals[0], state.ArgStack.Vals[1])
	pt2 := pt1
	pt2.move(state.ArgStack.Vals[2], state.ArgStack.Vals[3])
	pt3 := pt2
	pt3.move(state.ArgStack.Vals[4], state.ArgStack.Vals[5])
	pt4 := pt3
	pt4.move(state.ArgStack.Vals[6], state.ArgStack.Vals[7])
	pt5 := pt4
	pt5.move(state.ArgStack.Vals[8], state.ArgStack.Vals[9])
	pt6 := pt5
	pt6.move(state.ArgStack.Vals[10], state.ArgStack.Vals[11])

	met.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}

func (met *type2CharstringHandler) hflex1(state *ps.Inter) error {
	if state.ArgStack.Top != 9 {
		return fmt.Errorf("expected 9 operands for <hflex1>, got %d", state.ArgStack.Top)
	}
	pt1 := met.currentPoint
	pt1.move(state.ArgStack.Vals[0], state.ArgStack.Vals[1])
	pt2 := pt1
	pt2.move(state.ArgStack.Vals[2], state.ArgStack.Vals[3])
	pt3 := pt2
	pt3.x += state.ArgStack.Vals[4]
	pt4 := pt3
	pt4.x += state.ArgStack.Vals[5]
	pt5 := pt4
	pt5.move(state.ArgStack.Vals[6], state.ArgStack.Vals[7])
	pt6 := pt5
	pt6.x += state.ArgStack.Vals[8]
	pt6.y = met.currentPoint.y

	met.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}

func (met *type2CharstringHandler) flex1(state *ps.Inter) error {
	if state.ArgStack.Top != 11 {
		return fmt.Errorf("expected 11 operands for <flex1>, got %d", state.ArgStack.Top)
	}

	var d point
	for i := 0; i < 10; i += 2 {
		d.move(state.ArgStack.Vals[i], state.ArgStack.Vals[i+1])
	}

	pt1 := met.currentPoint
	pt1.move(state.ArgStack.Vals[0], state.ArgStack.Vals[1])
	pt2 := pt1
	pt2.move(state.ArgStack.Vals[2], state.ArgStack.Vals[3])
	pt3 := pt2
	pt3.move(state.ArgStack.Vals[4], state.ArgStack.Vals[5])
	pt4 := pt3
	pt4.move(state.ArgStack.Vals[6], state.ArgStack.Vals[7])
	pt5 := pt4
	pt5.move(state.ArgStack.Vals[8], state.ArgStack.Vals[9])
	pt6 := pt5

	if abs(d.x) > abs(d.y) {
		pt6.x += state.ArgStack.Vals[10]
		pt6.y = met.currentPoint.y
	} else {
		pt6.x = met.currentPoint.x
		pt6.y += state.ArgStack.Vals[10]
	}

	met.doubleCurve(pt1, pt2, pt3, pt4, pt5, pt6)
	return nil
}

func (met *type2CharstringHandler) hstem(state *ps.Inter) {
	met.hstemCount += state.ArgStack.Top / 2
}

func (met *type2CharstringHandler) vstem(state *ps.Inter) {
	met.vstemCount += state.ArgStack.Top / 2
}

// // psType2CharstringsData contains fields specific to the Type 2 Charstrings
// // context.
// type psType2CharstringsData struct {
// 	f          *Font
// 	b          *Buffer
// 	x          int32
// 	y          int32
// 	firstX     int32
// 	firstY     int32
// 	hintBits   int32
// 	seenWidth  bool
// 	ended      bool
// 	glyphIndex GlyphIndex
// 	// fdSelectIndexPlusOne is the result of the Font Dict Select lookup, plus
// 	// one. That plus one lets us use the zero value to denote either unused
// 	// (for CFF fonts with a single Font Dict) or lazily evaluated.
// 	fdSelectIndexPlusOne int32
// }

// func (d *psType2CharstringsData) initialize(f *Font, b *Buffer, glyphIndex GlyphIndex) {
// 	*d = psType2CharstringsData{
// 		f:          f,
// 		b:          b,
// 		glyphIndex: glyphIndex,
// 	}
// }

// func (d *psType2CharstringsData) closePath() {
// 	if d.x != d.firstX || d.y != d.firstY {
// 		d.b.segments = append(d.b.segments, Segment{
// 			Op: SegmentOpLineTo,
// 			Args: [3]fixed.Point26_6{{
// 				X: fixed.Int26_6(d.firstX),
// 				Y: fixed.Int26_6(d.firstY),
// 			}},
// 		})
// 	}
// }

// func (d *psType2CharstringsData) moveTo(dx, dy int32) {
// 	d.closePath()
// 	d.x += dx
// 	d.y += dy
// 	d.b.segments = append(d.b.segments, Segment{
// 		Op: SegmentOpMoveTo,
// 		Args: [3]fixed.Point26_6{{
// 			X: fixed.Int26_6(d.x),
// 			Y: fixed.Int26_6(d.y),
// 		}},
// 	})
// 	d.firstX = d.x
// 	d.firstY = d.y
// }

// func (d *psType2CharstringsData) lineTo(dx, dy int32) {
// 	d.x += dx
// 	d.y += dy
// 	d.b.segments = append(d.b.segments, Segment{
// 		Op: SegmentOpLineTo,
// 		Args: [3]fixed.Point26_6{{
// 			X: fixed.Int26_6(d.x),
// 			Y: fixed.Int26_6(d.y),
// 		}},
// 	})
// }

// func (d *psType2CharstringsData) cubeTo(dxa, dya, dxb, dyb, dxc, dyc int32) {
// 	d.x += dxa
// 	d.y += dya
// 	xa := fixed.Int26_6(d.x)
// 	ya := fixed.Int26_6(d.y)
// 	d.x += dxb
// 	d.y += dyb
// 	xb := fixed.Int26_6(d.x)
// 	yb := fixed.Int26_6(d.y)
// 	d.x += dxc
// 	d.y += dyc
// 	xc := fixed.Int26_6(d.x)
// 	yc := fixed.Int26_6(d.y)
// 	d.b.segments = append(d.b.segments, Segment{
// 		Op:   SegmentOpCubeTo,
// 		Args: [3]fixed.Point26_6{{X: xa, Y: ya}, {X: xb, Y: yb}, {X: xc, Y: yc}},
// 	})
// }

type psInterpreter struct{}

type psOperator struct {
	// run is the function that implements the operator. Nil means that we
	// ignore the operator, other than popping its arguments off the stack.
	run func(*psInterpreter) error
	// name is the operator name. An empty name (i.e. the zero value for the
	// struct overall) means an unrecognized 1-byte operator.
	name string
	// numPop is the number of stack values to pop. -1 means "array" and -2
	// means "delta" as per 5176.CFF.pdf Table 6 "Operand Types".
	numPop int32
}

// psOperators holds the 1-byte and 2-byte operators for PostScript interpreter
// contexts.
var psOperators = [...][2][]psOperator{
	// // The Type 2 Charstring operators are defined by 5177.Type2.pdf Appendix A
	// // "Type 2 Charstring Command Codes".
	// psContextType2Charstring: {{
	// 	// 1-byte operators.
	// 	0:  {}, // Reserved.
	// 	2:  {}, // Reserved.
	// 	1:  {-1, "hstem", t2CStem},
	// 	3:  {-1, "vstem", t2CStem},
	// 	18: {-1, "hstemhm", t2CStem},
	// 	23: {-1, "vstemhm", t2CStem},
	// 	5:  {-1, "rlineto", t2CRlineto},
	// 	6:  {-1, "hlineto", t2CHlineto},
	// 	7:  {-1, "vlineto", t2CVlineto},
	// 	8:  {-1, "rrcurveto", t2CRrcurveto},
	// 	9:  {}, // Reserved.
	// 	10: {+1, "callsubr", t2CCallsubr},
	// 	11: {+0, "return", t2CReturn},
	// 	12: {}, // escape.
	// 	13: {}, // Reserved.
	// 	14: {-1, "endchar", t2CEndchar},
	// 	15: {}, // Reserved.
	// 	16: {}, // Reserved.
	// 	17: {}, // Reserved.
	// 	19: {-1, "hintmask", t2CMask},
	// 	20: {-1, "cntrmask", t2CMask},
	// 	4:  {-1, "vmoveto", t2CVmoveto},
	// 	21: {-1, "rmoveto", t2CRmoveto},
	// 	22: {-1, "hmoveto", t2CHmoveto},
	// 	24: {-1, "rcurveline", t2CRcurveline},
	// 	25: {-1, "rlinecurve", t2CRlinecurve},
	// 	26: {-1, "vvcurveto", t2CVvcurveto},
	// 	27: {-1, "hhcurveto", t2CHhcurveto},
	// 	28: {}, // shortint.
	// 	29: {+1, "callgsubr", t2CCallgsubr},
	// 	30: {-1, "vhcurveto", t2CVhcurveto},
	// 	31: {-1, "hvcurveto", t2CHvcurveto},
	// }, {
	// 	// 2-byte operators. The first byte is the escape byte.
	// 	34: {+7, "hflex", t2CHflex},
	// 	36: {+9, "hflex1", t2CHflex1},
	// 	// TODO: more operators.
	// }},
}

// // t2CReadWidth reads the optional width adjustment. If present, it is on the
// // bottom of the arg stack. nArgs is the expected number of arguments on the
// // stack. A negative nArgs means a multiple of 2.
// //
// // 5177.Type2.pdf page 16 Note 4 says: "The first stack-clearing operator,
// // which must be one of hstem, hstemhm, vstem, vstemhm, cntrmask, hintmask,
// // hmoveto, vmoveto, rmoveto, or endchar, takes an additional argument — the
// // width... which may be expressed as zero or one numeric argument."
// func t2CReadWidth(p *psInterpreter, nArgs int32) {
// 	if p.type2Charstrings.seenWidth {
// 		return
// 	}
// 	p.type2Charstrings.seenWidth = true
// 	if nArgs >= 0 {
// 		if p.argStack.top != nArgs+1 {
// 			return
// 		}
// 	} else if p.argStack.top&1 == 0 {
// 		return
// 	}
// 	// When parsing a standalone CFF, we'd save the value of p.argStack.a[0]
// 	// here as it defines the glyph's width (horizontal advance). Specifically,
// 	// if present, it is a delta to the font-global nominalWidthX value found
// 	// in the Private DICT. If absent, the glyph's width is the defaultWidthX
// 	// value in that dict. See 5176.CFF.pdf section 15 "Private DICT Data".
// 	//
// 	// For a CFF embedded in an SFNT font (i.e. an OpenType font), glyph widths
// 	// are already stored in the hmtx table, separate to the CFF table, and it
// 	// is simpler to parse that table for all OpenType fonts (PostScript and
// 	// TrueType). We therefore ignore the width value here, and just remove it
// 	// from the bottom of the argStack.
// 	copy(p.argStack.a[:p.argStack.top-1], p.argStack.a[1:p.argStack.top])
// 	p.argStack.top--
// }

// func t2CStem(p *psInterpreter) error {
// 	t2CReadWidth(p, -1)
// 	if p.argStack.top%2 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	// We update the number of hintBits need to parse hintmask and cntrmask
// 	// instructions, but this Type 2 Charstring implementation otherwise
// 	// ignores the stem hints.
// 	p.type2Charstrings.hintBits += p.argStack.top / 2
// 	if p.type2Charstrings.hintBits > maxHintBits {
// 		return errUnsupportedNumberOfHints
// 	}
// 	return nil
// }

// func t2CMask(p *psInterpreter) error {
// 	// 5176.CFF.pdf section 4.3 "Hint Operators" says that "If hstem and vstem
// 	// hints are both declared at the beginning of a charstring, and this
// 	// sequence is followed directly by the hintmask or cntrmask operators, the
// 	// vstem hint operator need not be included."
// 	//
// 	// What we implement here is more permissive (but the same as what the
// 	// FreeType implementation does, and simpler than tracking the previous
// 	// operator and other hinting state): if a hintmask is given any arguments
// 	// (i.e. the argStack is non-empty), we run an implicit vstem operator.
// 	//
// 	// Note that the vstem operator consumes from p.argStack, but the hintmask
// 	// or cntrmask operators consume from p.instructions.
// 	if p.argStack.top != 0 {
// 		if err := t2CStem(p); err != nil {
// 			return err
// 		}
// 	} else if !p.type2Charstrings.seenWidth {
// 		p.type2Charstrings.seenWidth = true
// 	}

// 	hintBytes := (p.type2Charstrings.hintBits + 7) / 8
// 	if len(p.instructions) < int(hintBytes) {
// 		return errInvalidCFFTable
// 	}
// 	p.instructions = p.instructions[hintBytes:]
// 	return nil
// }

// func t2CHmoveto(p *psInterpreter) error {
// 	t2CReadWidth(p, 1)
// 	if p.argStack.top != 1 {
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.moveTo(p.argStack.a[0], 0)
// 	return nil
// }

// func t2CVmoveto(p *psInterpreter) error {
// 	t2CReadWidth(p, 1)
// 	if p.argStack.top != 1 {
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.moveTo(0, p.argStack.a[0])
// 	return nil
// }

// func t2CRmoveto(p *psInterpreter) error {
// 	t2CReadWidth(p, 2)
// 	if p.argStack.top != 2 {
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.moveTo(p.argStack.a[0], p.argStack.a[1])
// 	return nil
// }

// func t2CHlineto(p *psInterpreter) error { return t2CLineto(p, false) }
// func t2CVlineto(p *psInterpreter) error { return t2CLineto(p, true) }

// func t2CLineto(p *psInterpreter, vertical bool) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 1 {
// 		return errInvalidCFFTable
// 	}
// 	for i := int32(0); i < p.argStack.top; i, vertical = i+1, !vertical {
// 		dx, dy := p.argStack.a[i], int32(0)
// 		if vertical {
// 			dx, dy = dy, dx
// 		}
// 		p.type2Charstrings.lineTo(dx, dy)
// 	}
// 	return nil
// }

// func t2CRlineto(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 2 || p.argStack.top%2 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	for i := int32(0); i < p.argStack.top; i += 2 {
// 		p.type2Charstrings.lineTo(p.argStack.a[i], p.argStack.a[i+1])
// 	}
// 	return nil
// }

// // As per 5177.Type2.pdf section 4.1 "Path Construction Operators",
// //
// // rcurveline is:
// //	- {dxa dya dxb dyb dxc dyc}+ dxd dyd
// //
// // rlinecurve is:
// //	- {dxa dya}+ dxb dyb dxc dyc dxd dyd

// func t2CRcurveline(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 8 || p.argStack.top%6 != 2 {
// 		return errInvalidCFFTable
// 	}
// 	i := int32(0)
// 	for iMax := p.argStack.top - 2; i < iMax; i += 6 {
// 		p.type2Charstrings.cubeTo(
// 			p.argStack.a[i+0],
// 			p.argStack.a[i+1],
// 			p.argStack.a[i+2],
// 			p.argStack.a[i+3],
// 			p.argStack.a[i+4],
// 			p.argStack.a[i+5],
// 		)
// 	}
// 	p.type2Charstrings.lineTo(p.argStack.a[i], p.argStack.a[i+1])
// 	return nil
// }

// func t2CRlinecurve(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 8 || p.argStack.top%2 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	i := int32(0)
// 	for iMax := p.argStack.top - 6; i < iMax; i += 2 {
// 		p.type2Charstrings.lineTo(p.argStack.a[i], p.argStack.a[i+1])
// 	}
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[i+0],
// 		p.argStack.a[i+1],
// 		p.argStack.a[i+2],
// 		p.argStack.a[i+3],
// 		p.argStack.a[i+4],
// 		p.argStack.a[i+5],
// 	)
// 	return nil
// }

// // As per 5177.Type2.pdf section 4.1 "Path Construction Operators",
// //
// // hhcurveto is:
// //	- dy1 {dxa dxb dyb dxc}+
// //
// // vvcurveto is:
// //	- dx1 {dya dxb dyb dyc}+
// //
// // hvcurveto is one of:
// //	- dx1 dx2 dy2 dy3 {dya dxb dyb dxc dxd dxe dye dyf}* dxf?
// //	- {dxa dxb dyb dyc dyd dxe dye dxf}+ dyf?
// //
// // vhcurveto is one of:
// //	- dy1 dx2 dy2 dx3 {dxa dxb dyb dyc dyd dxe dye dxf}* dyf?
// //	- {dya dxb dyb dxc dxd dxe dye dyf}+ dxf?

// func t2CHhcurveto(p *psInterpreter) error { return t2CCurveto(p, false, false) }
// func t2CVvcurveto(p *psInterpreter) error { return t2CCurveto(p, false, true) }
// func t2CHvcurveto(p *psInterpreter) error { return t2CCurveto(p, true, false) }
// func t2CVhcurveto(p *psInterpreter) error { return t2CCurveto(p, true, true) }

// // t2CCurveto implements the hh / vv / hv / vh xxcurveto operators. N relative
// // cubic curve requires 6*N control points, but only 4*N+0 or 4*N+1 are used
// // here: all (or all but one) of the piecewise cubic curve's tangents are
// // implicitly horizontal or vertical.
// //
// // swap is whether that implicit horizontal / vertical constraint swaps as you
// // move along the piecewise cubic curve. If swap is false, the constraints are
// // either all horizontal or all vertical. If swap is true, it alternates.
// //
// // vertical is whether the first implicit constraint is vertical.
// func t2CCurveto(p *psInterpreter, swap, vertical bool) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 4 {
// 		return errInvalidCFFTable
// 	}

// 	i := int32(0)
// 	switch p.argStack.top & 3 {
// 	case 0:
// 		// No-op.
// 	case 1:
// 		if swap {
// 			break
// 		}
// 		i = 1
// 		if vertical {
// 			p.type2Charstrings.x += p.argStack.a[0]
// 		} else {
// 			p.type2Charstrings.y += p.argStack.a[0]
// 		}
// 	default:
// 		return errInvalidCFFTable
// 	}

// 	for i != p.argStack.top {
// 		i = t2CCurveto4(p, swap, vertical, i)
// 		if i < 0 {
// 			return errInvalidCFFTable
// 		}
// 		if swap {
// 			vertical = !vertical
// 		}
// 	}
// 	return nil
// }

// func t2CCurveto4(p *psInterpreter, swap bool, vertical bool, i int32) (j int32) {
// 	if i+4 > p.argStack.top {
// 		return -1
// 	}
// 	dxa := p.argStack.a[i+0]
// 	dya := int32(0)
// 	dxb := p.argStack.a[i+1]
// 	dyb := p.argStack.a[i+2]
// 	dxc := p.argStack.a[i+3]
// 	dyc := int32(0)
// 	i += 4

// 	if vertical {
// 		dxa, dya = dya, dxa
// 	}

// 	if swap {
// 		if i+1 == p.argStack.top {
// 			dyc = p.argStack.a[i]
// 			i++
// 		}
// 	}

// 	if swap != vertical {
// 		dxc, dyc = dyc, dxc
// 	}

// 	p.type2Charstrings.cubeTo(dxa, dya, dxb, dyb, dxc, dyc)
// 	return i
// }

// func t2CRrcurveto(p *psInterpreter) error {
// 	if !p.type2Charstrings.seenWidth || p.argStack.top < 6 || p.argStack.top%6 != 0 {
// 		return errInvalidCFFTable
// 	}
// 	for i := int32(0); i != p.argStack.top; i += 6 {
// 		p.type2Charstrings.cubeTo(
// 			p.argStack.a[i+0],
// 			p.argStack.a[i+1],
// 			p.argStack.a[i+2],
// 			p.argStack.a[i+3],
// 			p.argStack.a[i+4],
// 			p.argStack.a[i+5],
// 		)
// 	}
// 	return nil
// }

// // For the flex operators, we ignore the flex depth and always produce cubic
// // segments, not linear segments. It's not obvious why the Type 2 Charstring
// // format cares about switching behavior based on a metric in pixels, not in
// // ideal font units. The Go vector rasterizer has no problems with almost
// // linear cubic segments.

// func t2CHflex(p *psInterpreter) error {
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[0], 0,
// 		p.argStack.a[1], +p.argStack.a[2],
// 		p.argStack.a[3], 0,
// 	)
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[4], 0,
// 		p.argStack.a[5], -p.argStack.a[2],
// 		p.argStack.a[6], 0,
// 	)
// 	return nil
// }

// func t2CHflex1(p *psInterpreter) error {
// 	dy1 := p.argStack.a[1]
// 	dy2 := p.argStack.a[3]
// 	dy5 := p.argStack.a[7]
// 	dy6 := -dy1 - dy2 - dy5
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[0], dy1,
// 		p.argStack.a[2], dy2,
// 		p.argStack.a[4], 0,
// 	)
// 	p.type2Charstrings.cubeTo(
// 		p.argStack.a[5], 0,
// 		p.argStack.a[6], dy5,
// 		p.argStack.a[8], dy6,
// 	)
// 	return nil
// }

// // subrBias returns the subroutine index bias as per 5177.Type2.pdf section 4.7
// // "Subroutine Operators".
// func subrBias(numSubroutines int) int32 {
// 	if numSubroutines < 1240 {
// 		return 107
// 	}
// 	if numSubroutines < 33900 {
// 		return 1131
// 	}
// 	return 32768
// }

// func t2CCallgsubr(p *psInterpreter) error {
// 	return t2CCall(p, p.type2Charstrings.f.cached.glyphData.gsubrs)
// }

// func t2CCallsubr(p *psInterpreter) error {
// 	t := &p.type2Charstrings
// 	d := &t.f.cached.glyphData
// 	subrs := d.singleSubrs
// 	if d.multiSubrs != nil {
// 		if t.fdSelectIndexPlusOne == 0 {
// 			index, err := d.fdSelect.lookup(t.f, t.b, t.glyphIndex)
// 			if err != nil {
// 				return err
// 			}
// 			if index < 0 || len(d.multiSubrs) <= index {
// 				return errInvalidCFFTable
// 			}
// 			t.fdSelectIndexPlusOne = int32(index + 1)
// 		}
// 		subrs = d.multiSubrs[t.fdSelectIndexPlusOne-1]
// 	}
// 	return t2CCall(p, subrs)
// }

// func t2CCall(p *psInterpreter, subrs []uint32) error {
// 	if p.callStack.top == psCallStackSize || len(subrs) == 0 {
// 		return errInvalidCFFTable
// 	}
// 	length := uint32(len(p.instructions))
// 	p.callStack.a[p.callStack.top] = psCallStackEntry{
// 		offset: p.instrOffset + p.instrLength - length,
// 		length: length,
// 	}
// 	p.callStack.top++

// 	subrIndex := p.argStack.a[p.argStack.top-1] + subrBias(len(subrs)-1)
// 	if subrIndex < 0 || int32(len(subrs)-1) <= subrIndex {
// 		return errInvalidCFFTable
// 	}
// 	i := subrs[subrIndex+0]
// 	j := subrs[subrIndex+1]
// 	if j < i {
// 		return errInvalidCFFTable
// 	}
// 	if j-i > maxGlyphDataLength {
// 		return errUnsupportedGlyphDataLength
// 	}
// 	buf, err := p.type2Charstrings.b.view(&p.type2Charstrings.f.src, int(i), int(j-i))
// 	if err != nil {
// 		return err
// 	}

// 	p.instructions = buf
// 	p.instrOffset = i
// 	p.instrLength = j - i
// 	return nil
// }

// func t2CReturn(p *psInterpreter) error {
// 	if p.callStack.top <= 0 {
// 		return errInvalidCFFTable
// 	}
// 	p.callStack.top--
// 	o := p.callStack.a[p.callStack.top].offset
// 	n := p.callStack.a[p.callStack.top].length
// 	buf, err := p.type2Charstrings.b.view(&p.type2Charstrings.f.src, int(o), int(n))
// 	if err != nil {
// 		return err
// 	}

// 	p.instructions = buf
// 	p.instrOffset = o
// 	p.instrLength = n
// 	return nil
// }

// func t2CEndchar(p *psInterpreter) error {
// 	t2CReadWidth(p, 0)
// 	if p.argStack.top != 0 || p.hasMoreInstructions() {
// 		if p.argStack.top == 4 {
// 			// TODO: process the implicit "seac" command as per 5177.Type2.pdf
// 			// Appendix C "Compatibility and Deprecated Operators".
// 			return errUnsupportedType2Charstring
// 		}
// 		return errInvalidCFFTable
// 	}
// 	p.type2Charstrings.closePath()
// 	p.type2Charstrings.ended = true
// 	return nil
// }
