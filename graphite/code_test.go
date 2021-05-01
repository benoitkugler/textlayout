package graphite

import (
	"fmt"
	"os"
	"testing"
)

func TestSimpleInstructions(t *testing.T) {
	f, err := os.Open("testdata/Dummy.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := ParseFont(f)
	if err != nil {
		t.Fatal(err)
	}

	simpleValid := []byte{
		PUSH_BYTE, 43,
		PUSH_BYTE, 42,
		PUSH_BYTE, 11, PUSH_BYTE, 13, ADD,
		PUSH_BYTE, 4, SUB,
		COND,
		PUSH_LONG, 0x80, 0x00, 0x00, 0x00,
		PUSH_LONG, 0xff, 0xff, 0xff, 0xff,
		DIV,
		POP_RET,
	}
	simpleUnderflow := []byte{
		PUSH_BYTE, 43,
		PUSH_BYTE, 42,
		PUSH_BYTE, 11, PUSH_BYTE, 13, ADD,
		PUSH_BYTE, 4, SUB,
		COND,
		PUSH_LONG, 0x80, 0x00, 0x00, 0x00,
		PUSH_LONG, 0xff, 0xff, 0xff, 0xff,
		DIV,
		COND, // Uncomment to cause an underflow
		POP_RET,
	}

	// simpleOveflow := []byte{
	// 	PUSH_BYTE, 43,
	// 	PUSH_BYTE, 42,
	// 	PUSH_LONG, 1, 2, 3, 4,
	// 	PUSH_BYTE, 11, PUSH_BYTE, 13, ADD,
	// 	PUSH_BYTE, 4, SUB,
	// 	COND,
	// 	PUSH_LONG, 0x80, 0x00, 0x00, 0x00,
	// 	PUSH_LONG, 0xff, 0xff, 0xff, 0xff,
	// 	DIV,
	// 	POP_RET,
	// }

	prog, err := newCode(false, simpleValid, 0, 0, &silfSubtable{}, &font, PASS_TYPE_UNKNOWN)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	fmt.Println(prog)

	_, err = newCode(false, simpleUnderflow, 0, 0, &silfSubtable{}, &font, PASS_TYPE_UNKNOWN)
	if err != underfullStack {
		t.Fatalf("expected underfull error, got %s", err)
	}
}
