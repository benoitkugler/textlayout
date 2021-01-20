package type1

import (
	"errors"
	"fmt"

	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

var _ psinterpreter.PsOperatorHandler = (*type1Metrics)(nil)

type Position struct {
	X, Y int32
}

// type1Metrics handles the Type1 charstring operators needed
// to fetch metrics.
// Most of the operators are ignored
type type1Metrics struct {
	leftBearing, advance Position
}

func (type1Metrics) Context() psinterpreter.PsContext { return psinterpreter.Type1Charstring }

// Run only look for metrics information, that is the 'sbw' and 'hsbw' operators.
// Since they must be the first, we dont support the other operators and return an error if found.
func (data *type1Metrics) Run(operator psinterpreter.PsOperator, state *psinterpreter.Inter) (int32, error) {
	if operator.Operator == 13 && !operator.IsEscaped { // hsbw
		if state.ArgStack.Top < 2 {
			return 0, errors.New("invalid stack size for 'hsbw' in Type1 charstring")
		}
		data.leftBearing.X += state.ArgStack.Vals[state.ArgStack.Top-2]
		data.advance.X = state.ArgStack.Vals[state.ArgStack.Top-1]
		data.advance.Y = 0
		return 0, psinterpreter.ErrInterrupt // stop early
	}
	if operator.Operator == 7 && operator.IsEscaped { // sbw
		if state.ArgStack.Top < 4 {
			return 0, errors.New("invalid stack size for 'sbw' in Type1 charstring")
		}
		data.leftBearing.X += state.ArgStack.Vals[state.ArgStack.Top-4]
		data.leftBearing.Y += state.ArgStack.Vals[state.ArgStack.Top-3]
		data.advance.X = state.ArgStack.Vals[state.ArgStack.Top-2]
		data.advance.Y = state.ArgStack.Vals[state.ArgStack.Top-1]
		return 0, psinterpreter.ErrInterrupt // stop early
	}

	return 0, fmt.Errorf("unsupported operand %s in Type1 charstring", operator)
}
