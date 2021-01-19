package type1

import (
	"fmt"

	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
)

var _ psinterpreter.PsOperatorHandler = (*type1Operators)(nil)

type Position struct {
	X, Y int32
}

// Type1 charstrings handler
type type1Operators struct {
	leftBearing, advance Position
}

func (type1Operators) Context() psinterpreter.PsContext { return psinterpreter.Type1Charstring }

func (data *type1Operators) Run(operator psinterpreter.PsOperator, state *psinterpreter.Inter) (int32, error) {
	ops := psContextType1Charstring[0]
	if operator.IsEscaped {
		ops = psContextType1Charstring[1]
	}
	if int(operator.Operator) >= len(ops) {
		return 0, fmt.Errorf("invalid operator %s in Type1 charstring", operator)
	}
	op := ops[operator.Operator]
	if op.run == nil {
		return 0, fmt.Errorf("invalid operator %s in Type1 charstring", operator)
	}
	err := op.run(data, state)
	return op.numPop, err
}

type psOperator struct {
	// numPop is the number of stack values to pop. -1 means "array" and -2
	// means "delta" as per 5176.CFF.pdf Table 6 "Operand Types".
	numPop int32

	// run is the function that implements the operator. Nil means that we
	// ignore the operator, other than popping its arguments off the stack.
	run func(*type1Operators, *psinterpreter.Inter) error
}

func noOpOperator(_ *type1Operators, state *psinterpreter.Inter) error {
	fmt.Println(state.ArgStack.Top)
	return nil
}

var psContextType1Charstring = [2][]psOperator{
	{
		1:  {-1, noOpOperator}, // "hstem"
		3:  {-1, noOpOperator}, // "vstem"
		4:  {-1, noOpOperator}, // "vmoveto"
		5:  {-1, noOpOperator}, // "rlineto"
		6:  {-1, noOpOperator}, // "hlineto"
		7:  {-1, noOpOperator}, // "vlineto"
		8:  {-1, noOpOperator}, // "rrcurveto"
		9:  {-1, noOpOperator}, // "closepath"
		10: {1, noOpOperator},  // "callsubr"
		11: {0, noOpOperator},  // "return"
		12: {},                 // escape
		13: {-1, func(to *type1Operators, i *psinterpreter.Inter) error {
			to.leftBearing.X += i.ArgStack.Vals[i.ArgStack.Top-2]
			return nil
		}}, // "hsbw"
		14: {-1, noOpOperator}, // "endchar"
		21: {-1, noOpOperator}, // "rmoveto"
		22: {-1, noOpOperator}, // "hmoveto"
		30: {-1, noOpOperator}, // "vhcurveto"
		31: {-1, noOpOperator}, // "hvcurveto"
	},
	{
		0:  {0, noOpOperator},  // "dotsection"
		1:  {-1, noOpOperator}, // "vstem3"
		2:  {-1, noOpOperator}, // "hstem3"
		6:  {-1, noOpOperator}, // "seac"
		7:  {-1, noOpOperator}, // "sbw"
		12: {0, noOpOperator},  // "div", manual stack managment
		16: {-1, noOpOperator}, // "callothersubr"
		17: {0, noOpOperator},  // "pop", manual stack managment
		33: {-1, noOpOperator}, // "setcurrentpoint"
	},
}
