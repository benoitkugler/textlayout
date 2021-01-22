// Package psinterpreter implement a Postscript interpreter
// required to parse .CFF files, and Type1 and Type2 Charstrings.
// This package provides the low-level mechanisms needed to
// read such formats; the data is consumed in higher level packages,
// which implement `PsOperatorHandler`.
package psinterpreter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
)

var (
	// ErrInterrupt signals the interpretter to stop early, without erroring.
	ErrInterrupt                     = errors.New("interruption")
	errInvalidCFFTable               = errors.New("invalid ps instructions")
	errUnsupportedCFFVersion         = errors.New("unsupported CFF version")
	errUnsupportedRealNumberEncoding = errors.New("unsupported real number encoding")

	be = binary.BigEndian
)

const (
	// psArgStackSize is the argument stack size for a PostScript interpreter.
	// 5176.CFF.pdf section 4 "DICT Data" says that "An operator may be
	// preceded by up to a maximum of 48 operands". 5177.Type2.pdf Appendix B
	// "Type 2 Charstring Implementation Limits" says that "Argument stack 48".
	// T1_SPEC.pdf 6.1 Encoding as a limitation of 24.
	psArgStackSize = 48

	// Similarly, Appendix B says "Subr nesting, stack limit 10".
	psCallStackSize = 10

	maxRealNumberStrLen = 64 // Maximum length in bytes of the "-123.456E-7" representation.
)

// PsContext is the flavour of the Postcript language.
type PsContext uint32

const (
	TopDict         PsContext = iota // Top dict in CFF files
	PrivateDict                      // Private dict in CFF files
	Type2Charstring                  // Charstring in CFF files
	Type1Charstring                  // Charstring in Type1 font files
)

type ArgStack struct {
	Vals [psArgStackSize]int32
	// Effecive size currently in use. the First value to
	// pop is at index Top-1
	Top int32
}

// Uint16 returns the top level value as uint16.
func (a *ArgStack) Uint16() uint16 { return uint16(a.Vals[a.Top-1]) }

// Float return the value of a real number, stored as its binary representation.
func (a *ArgStack) Float() float32 {
	return math.Float32frombits(uint32(a.Vals[a.Top-1]))
}

// Inter is a PostScript interpreter.
// A same interpreter may be re-used using muliples `Run` calls
type Inter struct {
	ctx          PsContext
	instructions []byte
	subrs        [][]byte

	ArgStack ArgStack

	callStack struct {
		a   [psCallStackSize][]byte // parent instructions
		top int32                   // effecive size currently in use
	}

	parseNumberBuf [maxRealNumberStrLen]byte
}

func (p *Inter) hasMoreInstructions() bool {
	if len(p.instructions) != 0 {
		return true
	}
	for i := int32(0); i < p.callStack.top; i++ {
		if len(p.callStack.a[i]) != 0 {
			return true
		}
	}
	return false
}

// 5176.CFF.pdf section 4 "DICT Data" says that "Two-byte operators have an
// initial escape byte of 12".
const escapeByte = 12

// Run runs the instructions in the PostScript context asked by `handler`.
// `subrs` contains the subroutines that may be called in the instructions.
func (p *Inter) Run(instructions []byte, subrs [][]byte, handler PsOperatorHandler) error {
	p.ctx = handler.Context()
	p.instructions = instructions
	p.subrs = subrs
	p.ArgStack.Top = 0
	p.callStack.top = 0

	for len(p.instructions) > 0 {
		// Push a numeric operand on the stack, if applicable.
		if hasResult, err := p.parseNumber(); hasResult {
			if err != nil {
				return err
			}
			continue
		}

		// Otherwise, execute an operator.
		b := p.instructions[0]
		p.instructions = p.instructions[1:]

		// check for the escape byte
		escaped := b == escapeByte
		if escaped {
			if len(p.instructions) <= 0 {
				return errInvalidCFFTable
			}
			b = p.instructions[0]
			p.instructions = p.instructions[1:]
		}

		numPop, err := handler.Run(PsOperator{Operator: b, IsEscaped: escaped}, p)
		if err == ErrInterrupt { // stop cleanly
			return nil
		}
		if err != nil {
			return err
		}
		if p.ArgStack.Top < numPop {
			return errInvalidCFFTable
		}
		if numPop < 0 { // pop all
			p.ArgStack.Top = 0
		} else {
			p.ArgStack.Top -= numPop
		}

	}
	return nil
}

// See 5176.CFF.pdf section 4 "DICT Data".
func (p *Inter) parseNumber() (hasResult bool, err error) {
	number := int32(0)
	switch b := p.instructions[0]; {
	case b == 28:
		if len(p.instructions) < 3 {
			return true, errInvalidCFFTable
		}
		number, hasResult = int32(int16(be.Uint16(p.instructions[1:]))), true
		p.instructions = p.instructions[3:]

	case b == 29 && p.ctx != Type2Charstring:
		if len(p.instructions) < 5 {
			return true, errInvalidCFFTable
		}
		number, hasResult = int32(be.Uint32(p.instructions[1:])), true
		p.instructions = p.instructions[5:]

	case b == 30 && p.ctx != Type2Charstring && p.ctx != Type1Charstring:
		// Parse a real number. This isn't listed in 5176.CFF.pdf Table 3
		// "Operand Encoding" but that table lists integer encodings. Further
		// down the page it says "A real number operand is provided in addition
		// to integer operands. This operand begins with a byte value of 30
		// followed by a variable-length sequence of bytes."

		s := p.parseNumberBuf[:0]
		p.instructions = p.instructions[1:]
	loop:
		for {
			if len(p.instructions) == 0 {
				return true, errInvalidCFFTable
			}
			b := p.instructions[0]
			p.instructions = p.instructions[1:]
			// Process b's two nibbles, high then low.
			for i := 0; i < 2; i++ {
				nib := b >> 4
				b = b << 4
				if nib == 0x0f {
					f, err := strconv.ParseFloat(string(s), 32)
					if err != nil {
						return true, errInvalidCFFTable
					}
					number, hasResult = int32(math.Float32bits(float32(f))), true
					break loop
				}
				if nib == 0x0d {
					return true, errInvalidCFFTable
				}
				if len(s)+maxNibbleDefsLength > len(p.parseNumberBuf) {
					return true, errUnsupportedRealNumberEncoding
				}
				s = append(s, nibbleDefs[nib]...)
			}
		}

	case b < 32:
		// not a number: no-op.
	case b < 247:
		p.instructions = p.instructions[1:]
		number, hasResult = int32(b)-139, true
	case b < 251:
		if len(p.instructions) < 2 {
			return true, errInvalidCFFTable
		}
		b1 := p.instructions[1]
		p.instructions = p.instructions[2:]
		number, hasResult = +int32(b-247)*256+int32(b1)+108, true
	case b < 255:
		if len(p.instructions) < 2 {
			return true, errInvalidCFFTable
		}
		b1 := p.instructions[1]
		p.instructions = p.instructions[2:]
		number, hasResult = -int32(b-251)*256-int32(b1)-108, true
	case b == 255 && (p.ctx == Type2Charstring || p.ctx == Type1Charstring):
		if len(p.instructions) < 5 {
			return true, errInvalidCFFTable
		}
		number, hasResult = int32(be.Uint32(p.instructions[1:])), true
		p.instructions = p.instructions[5:]
	}

	if hasResult {
		if p.ArgStack.Top == psArgStackSize {
			return true, errInvalidCFFTable
		}
		p.ArgStack.Vals[p.ArgStack.Top] = number
		p.ArgStack.Top++
	}
	return hasResult, nil
}

const maxNibbleDefsLength = len("E-")

// nibbleDefs encodes 5176.CFF.pdf Table 5 "Nibble Definitions".
var nibbleDefs = [16]string{
	0x00: "0",
	0x01: "1",
	0x02: "2",
	0x03: "3",
	0x04: "4",
	0x05: "5",
	0x06: "6",
	0x07: "7",
	0x08: "8",
	0x09: "9",
	0x0a: ".",
	0x0b: "E",
	0x0c: "E-",
	0x0d: "",
	0x0e: "-",
	0x0f: "",
}

// CallSubroutine calls the subroutine (identified by its index).
// No argument stack modification is performed.
func (p *Inter) CallSubroutine(index int32) error {
	if int(index) >= len(p.subrs) {
		return fmt.Errorf("invalid subroutine index %d (for length %d)", index, len(p.subrs))
	}
	if p.callStack.top == psCallStackSize {
		return errors.New("maximum call stack size reached")
	}
	// save the current instructions
	p.callStack.a[p.callStack.top] = p.instructions
	p.callStack.top++

	// activate the subroutine
	p.instructions = p.subrs[index]
	return nil
}

// PsOperator is a poscript command, which may be escaped.
type PsOperator struct {
	Operator  byte
	IsEscaped bool
}

func (p PsOperator) String() string {
	if p.IsEscaped {
		return fmt.Sprintf("2-byte operator (12 %d)", p.Operator)
	}
	return fmt.Sprintf("1-byte operator (%d)", p.Operator)
}

// PsOperatorHandler defines the behaviour of an operator.
type PsOperatorHandler interface {
	// Context defines the precise behaviour of the interpreter,
	// which has small nuances depending on the context.
	Context() PsContext

	// Run implements the operator defined by `operator` (which is the second byte if `escaped` is true).
	//
	// It must return the number of stack values to pop, after running the operator.
	// -1 means all the stack and -2 means "delta" as per 5176.CFF.pdf Table 6 "Operand Types".
	// Note that this is a convenient shorcut since `state` can also be directly mutated,
	// which is required to handle subroutines and numerics operations.
	//
	// Returning `ErrInterrupt` stop the parsing of the instructions, without reporting an error.
	Run(operator PsOperator, state *Inter) (numPop int32, err error)
}
