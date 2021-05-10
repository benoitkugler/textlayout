package type1

import (
	"errors"
	"fmt"
	"log"

	ps "github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

var _ ps.PsOperatorHandler = (*type1Metrics)(nil)

type Position struct {
	X, Y int32
}

// type1Metrics handles the Type1 charstring operators needed
// to fetch metrics.
// Most of the operators are ignored
type type1Metrics struct {
	cs ps.CharstringReader

	leftBearing, advance Position
}

func (type1Metrics) Context() ps.PsContext { return ps.Type1Charstring }

// Run only look for metrics information, that is the 'sbw' and 'hsbw' operators.
// Since they must be the first, we dont support the other operators and return an error if found.
func (data *type1Metrics) oldRun(operator ps.PsOperator, state *ps.Machine) error {
	if operator.Operator == 13 && !operator.IsEscaped { // hsbw
		if state.ArgStack.Top < 2 {
			return errors.New("invalid stack size for 'hsbw' in Type1 charstring")
		}
		data.leftBearing.X += state.ArgStack.Vals[state.ArgStack.Top-2]
		data.advance.X = state.ArgStack.Vals[state.ArgStack.Top-1]
		data.advance.Y = 0
		return ps.ErrInterrupt // stop early
	}
	if operator.Operator == 7 && operator.IsEscaped { // sbw
		if state.ArgStack.Top < 4 {
			return errors.New("invalid stack size for 'sbw' in Type1 charstring")
		}
		data.leftBearing.X += state.ArgStack.Vals[state.ArgStack.Top-4]
		data.leftBearing.Y += state.ArgStack.Vals[state.ArgStack.Top-3]
		data.advance.X = state.ArgStack.Vals[state.ArgStack.Top-2]
		data.advance.Y = state.ArgStack.Vals[state.ArgStack.Top-1]
		return ps.ErrInterrupt // stop early
	}

	return fmt.Errorf("unsupported operand %s in Type1 charstring", operator)
}

func (met *type1Metrics) Apply(op ps.PsOperator, state *ps.Machine) error {
	var err error
	if !op.IsEscaped {
		switch op.Operator {
		case 1: // hstem
			met.cs.Hstem(state)
		case 3: // vstem
			met.cs.Vstem(state)
		case 4: // vmoveto
			err = met.cs.Vmoveto(state)
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
		case 14: // endchar
			return ps.ErrInterrupt
		case 21: // rmoveto
			err = met.cs.Rmoveto(state)
		case 22: // hmoveto
			err = met.cs.Hmoveto(state)
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
		case 1: // vstem3
			met.cs.Vstem(state)
		case 2:
			met.cs.Hstem(state)

		case 7: // sbw
			if state.ArgStack.Top < 4 {
				return errors.New("invalid stack size for 'sbw' in Type1 charstring")
			}
			met.leftBearing.X += state.ArgStack.Vals[state.ArgStack.Top-4]
			met.leftBearing.Y += state.ArgStack.Vals[state.ArgStack.Top-3]
			met.advance.X = state.ArgStack.Vals[state.ArgStack.Top-2]
			met.advance.Y = state.ArgStack.Vals[state.ArgStack.Top-1]
		case 0, 6, 16, 17, 33: // dotsection, seac, callothersubr, pop, setcurrentpoint
			// FIXME: ignoring this operation
			log.Println("unhandled operation", op)
		default:
			// no other operands are allowed before the ones handled above
			err = fmt.Errorf("invalid operator %s in charstring", op)
		}
	}
	state.ArgStack.Clear()
	return err
}
