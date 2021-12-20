package type1

import (
	"errors"
	"fmt"

	ps "github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

var _ ps.PsOperatorHandler = (*type1CharstringParser)(nil)

// metrics for a seac glyph need to read other glyphs
type seac struct {
	aCode, bCode          int32
	accentOrigin          ps.Point
	accentLeftSideBearing int32
}

// type1CharstringParser handles the Type1 charstring operators needed
// to fetch metrics.
type type1CharstringParser struct {
	seac *seac // filled for seac operators

	flexPoints []ps.Point // filled with OtherSub(2) opcodes

	cs ps.CharstringReader

	inFlex bool // alter the behavior of moveto opcodes

	leftBearing, advance ps.Point
}

func (type1CharstringParser) Context() ps.PsContext { return ps.Type1Charstring }

func (met *type1CharstringParser) Apply(op ps.PsOperator, state *ps.Machine) error {
	var err error
	if !op.IsEscaped {
		switch op.Operator {
		case 1: // hstem
			met.cs.Hstem(state)
		case 3: // vstem
			met.cs.Vstem(state)
		case 4: // vmoveto
			if met.inFlex {
				if state.ArgStack.Top < 1 {
					return errors.New("invalid vmoveto operator")
				}
				y := state.ArgStack.Pop()
				met.flexPoints = append(met.flexPoints, ps.Point{X: 0, Y: y})
			} else {
				err = met.cs.Vmoveto(state)
			}
		case 5: // rlineto
			met.cs.Rlineto(state)
		case 6: // hlineto
			met.cs.Hlineto(state)
		case 7: // vlineto
			met.cs.Vlineto(state)
		case 8: // rrcurveto
			met.cs.Rrcurveto(state)
		case 9: // closepath
			met.cs.ClosePath()
		case 10: // callsubr
			return ps.LocalSubr(state) // do not clear the arg stack
		case 11: // return
			return state.Return() // do not clear the arg stack
		case 13: // hsbw
			if state.ArgStack.Top < 2 {
				return errors.New("invalid stack size for 'hsbw' in Type1 charstring")
			}
			met.leftBearing.X += state.ArgStack.Vals[state.ArgStack.Top-2]
			met.advance.X = state.ArgStack.Vals[state.ArgStack.Top-1]
			met.advance.Y = 0
			// "This command also sets the current point to (sbx, 0),
			// but does not place the point in the character path."
			met.cs.CurrentPoint = ps.Point{X: met.leftBearing.X, Y: 0}
		case 14: // endchar
			return ps.ErrInterrupt
		case 21: // rmoveto
			if met.inFlex {
				if state.ArgStack.Top < 2 {
					return errors.New("invalid rmoveto operator")
				}
				y := state.ArgStack.Pop()
				x := state.ArgStack.Pop()
				met.flexPoints = append(met.flexPoints, ps.Point{X: x, Y: y})
			} else {
				err = met.cs.Rmoveto(state)
			}
		case 22: // hmoveto
			if met.inFlex {
				if state.ArgStack.Top < 1 {
					return errors.New("invalid hmoveto operator")
				}
				x := state.ArgStack.Pop()
				met.flexPoints = append(met.flexPoints, ps.Point{X: x, Y: 0})
			} else {
				err = met.cs.Hmoveto(state)
			}
		case 30: // vhcurveto
			met.cs.Vhcurveto(state)
		case 31: // hvcurveto
			met.cs.Hvcurveto(state)

		default:
			// no other operands are allowed before the ones handled above
			err = fmt.Errorf("invalid operator %s in charstring", op)
		}
	} else {
		switch op.Operator {
		case 0: // dotsection
			// just clear the stack
		case 1: // vstem3
			met.cs.Vstem(state)
		case 2:
			met.cs.Hstem(state)
		case 6: // seac
			if state.ArgStack.Top < 5 {
				return errors.New("invalid stack size for 'seac' in Type1 charstring")
			}
			met.seac = &seac{
				aCode: state.ArgStack.Vals[state.ArgStack.Top-1],
				bCode: state.ArgStack.Vals[state.ArgStack.Top-2],
				accentOrigin: ps.Point{
					Y: state.ArgStack.Vals[state.ArgStack.Top-3],
					X: state.ArgStack.Vals[state.ArgStack.Top-4],
				},
				accentLeftSideBearing: state.ArgStack.Vals[state.ArgStack.Top-5],
			}
		case 7: // sbw
			if state.ArgStack.Top < 4 {
				return errors.New("invalid stack size for 'sbw' in Type1 charstring")
			}
			met.leftBearing.X += state.ArgStack.Vals[state.ArgStack.Top-4]
			met.leftBearing.Y += state.ArgStack.Vals[state.ArgStack.Top-3]
			met.advance.X = state.ArgStack.Vals[state.ArgStack.Top-2]
			met.advance.Y = state.ArgStack.Vals[state.ArgStack.Top-1]
		case 16: // callothersubr
			return met.otherSub(state) // do not clear the stack
		case 17: // pop: actually it pushes back to the stack
			if int(state.ArgStack.Top) >= len(state.ArgStack.Vals) {
				return errors.New("stack overflow in Type1 charstring")
			}
			state.ArgStack.Top++
			return nil // do not clear the stack
		case 33: // setcurrentpoint
			if state.ArgStack.Top < 2 {
				return errors.New("invalid setcurrentpoint operator (empty stack)")
			}
			met.cs.CurrentPoint.Y = state.ArgStack.Pop()
			met.cs.CurrentPoint.X = state.ArgStack.Pop()
		default:
			// no other operands are allowed before the ones handled above
			err = fmt.Errorf("invalid operator %s in charstring", op)
		}
	}
	state.ArgStack.Clear()
	return err
}

func (met *type1CharstringParser) otherSub(state *ps.Machine) error {
	if state.ArgStack.Top < 2 {
		return errors.New("invalid stack size for 'callothersubr' in Type1 charstring")
	}
	index := state.ArgStack.Pop() // index
	nbArgs := state.ArgStack.Pop()
	state.ArgStack.PopN(nbArgs)

	// we only support the Flex feature
	switch index {
	case 0: // end flex
		met.inFlex = false

		if nbArgs != 3 {
			return fmt.Errorf("invalid number of arguments for StartFlex other sub: %d", nbArgs)
		}
		if len(met.flexPoints) < 7 {
			return fmt.Errorf("invalid number of flex points for EndFlex other sub: %d", len(met.flexPoints))
		}

		// reference point is relative to start point
		reference := &met.flexPoints[0]
		reference.Move(met.cs.CurrentPoint.X, met.cs.CurrentPoint.Y)

		// first point is relative to reference point
		first := &met.flexPoints[1]
		first.Move(reference.X, reference.Y)

		// make the first point relative to the start point
		first.Move(-met.cs.CurrentPoint.X, -met.cs.CurrentPoint.Y)

		met.cs.RelativeCurveTo(met.flexPoints[1], met.flexPoints[2], met.flexPoints[3])
		met.cs.RelativeCurveTo(met.flexPoints[4], met.flexPoints[5], met.flexPoints[6])

		// reset the flex points
		met.flexPoints = met.flexPoints[:0]

		// update the stack with the return values (after popping the args)
		state.ArgStack.Vals[state.ArgStack.Top] = met.cs.CurrentPoint.X
		state.ArgStack.Vals[state.ArgStack.Top+1] = met.cs.CurrentPoint.Y
	case 1: // start flex
		if nbArgs != 0 {
			return fmt.Errorf("invalid number of arguments for StartFlex other sub: %d", nbArgs)
		}
		met.inFlex = true
		met.flexPoints = met.flexPoints[:0]
	case 2: // add flex vector
		if nbArgs != 0 {
			return fmt.Errorf("invalid number of arguments for StartFlex other sub: %d", nbArgs)
		}
		// implemented in the moveto op codes
	default:
		// not handled
	}
	return nil
}
