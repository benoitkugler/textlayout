package graphite

import (
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

	valid := []byte{
		PUSH_BYTE, 43,
		PUSH_BYTE, 42,
		PUSH_BYTE, 11, PUSH_BYTE, 13, ADD,
		PUSH_BYTE, 4, SUB,
		COND,
		// PUSH_LONG, 0x80, 0x50, 0x00, 0x00,
		// PUSH_LONG, 0xff, 0xff, 0xff, 0xff,
		// DIV,
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

	invalidDiv := []byte{
		PUSH_BYTE, 43,
		PUSH_BYTE, 42,
		PUSH_LONG, 1, 2, 3, 4,
		PUSH_BYTE, 11, PUSH_BYTE, 13, ADD,
		PUSH_BYTE, 4, SUB,
		COND,
		PUSH_LONG, 0x80, 0x00, 0x00, 0x00,
		PUSH_LONG, 0xff, 0xff, 0xff, 0xff,
		DIV,
		POP_RET,
	}

	prog, err := newCode(false, valid, 0, 0, &silfSubtable{}, &font, PASS_TYPE_UNKNOWN)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = newCode(false, simpleUnderflow, 0, 0, &silfSubtable{}, &font, PASS_TYPE_UNKNOWN)
	if err != underfullStack {
		t.Fatalf("expected underfull error, got %s", err)
	}

	progOverflow, err := newCode(false, invalidDiv, 0, 0, &silfSubtable{}, &font, PASS_TYPE_UNKNOWN)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	sm := newSlotMap(new(Segment), false, 0)
	m := newMachine(sm)
	m.map_.pushSlot(new(Slot))
	out, err := m.run(prog)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if out != 42 {
		t.Fatalf("expected 42, got %d", out)
	}

	if _, err := m.run(progOverflow); err != machine_died_early {
		t.Fatalf("expected machine_died_early error, got %s", err)
	}
}
