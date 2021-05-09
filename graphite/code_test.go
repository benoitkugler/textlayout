package graphite

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"testing"
)

func TestSimpleInstructions(t *testing.T) {
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

	context := codeContext{NumAttributes: 8, NumFeatures: 1}
	prog, err := newCode(false, valid, 0, 0, context)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = newCode(false, simpleUnderflow, 0, 0, context)
	if err != underfullStack {
		t.Fatalf("expected underfull error, got %s", err)
	}

	progOverflow, err := newCode(false, invalidDiv, 0, 0, context)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	sm := newSlotMap(new(Segment), false, 0)
	m := newMachine(sm)
	m.map_.pushSlot(new(Slot))
	out, err := m.run(&prog, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if out != 42 {
		t.Fatalf("expected 42, got %d", out)
	}

	if _, err := m.run(&progOverflow, nil); err != machine_died_early {
		t.Fatalf("expected machine_died_early error, got %s", err)
	}
}

type expectedRule struct {
	action     []opcode
	constraint []opcode
	preContext uint8
	sortKey    uint16
}

// load a .ttx file, produced by fonttools
func readExpectedOpCodes() [][]expectedRule {
	data, err := ioutil.ReadFile("testdata/Annapurnarc2.ttx")
	if err != nil {
		log.Fatal(err)
	}

	type xmlDoc struct {
		Passes []struct {
			Rules []struct {
				Action     string `xml:"action"`
				Constraint string `xml:"constraint"`
				PreContext uint8  `xml:"precontext,attr"`
				SortKey    uint16 `xml:"sortkey,attr"`
			} `xml:"rules>rule"`
			Index int `xml:"_index,attr"`
		} `xml:"Silf>silf>passes>pass"`
	}
	var doc xmlDoc
	err = xml.Unmarshal(data, &doc)
	if err != nil {
		log.Fatal(err)
	}

	parseOpCodes := func(s string) []opcode {
		chunks := strings.Split(s, "\n")
		var out []opcode
	mainLoop:
		for _, c := range chunks {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			// remove the arguments
			if i := strings.IndexByte(c, '('); i != -1 {
				c = c[:i]
			}

			for opc, data := range opcode_table {
				if data.name == c {
					out = append(out, opcode(opc))
					continue mainLoop
				}
			}
			log.Fatalf("unknown opcode: %s", c)
		}
		return out
	}

	var out [][]expectedRule
	for _, pass := range doc.Passes {
		var rules []expectedRule
		for _, rule := range pass.Rules {
			r := expectedRule{
				preContext: rule.PreContext,
				sortKey:    rule.SortKey,
				action:     parseOpCodes(rule.Action),
				constraint: parseOpCodes(rule.Constraint),
			}
			rules = append(rules, r)
		}
		out = append(out, rules)
	}

	return out
}

func TestOpCodesValues(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/Annapurnac2_bytecodes.json")
	if err != nil {
		t.Fatal(err)
	}
	var passes []extractedPass
	if err = json.Unmarshal(b, &passes); err != nil {
		t.Fatal(err)
	}

	expected := readExpectedOpCodes()
	if len(passes) != len(expected) {
		t.Fatal("invalid length")
	}

	for i, exp := range expected {
		input := passes[i]
		if len(input.Rules) != len(exp) {
			t.Fatal("invalid length")
		}

		for j, rule := range exp {
			inputRule := input.Rules[j]
			context := input.Context

			if rule.preContext != inputRule.PreContext {
				t.Fatalf("expected %d, got %d", rule.preContext, inputRule.PreContext)
			}
			if rule.sortKey != inputRule.SortKey {
				t.Fatalf("expected %d, got %d", rule.sortKey, inputRule.SortKey)
			}

			if len(rule.action) != 0 { // test action rule
				code := inputRule.Action
				loadedCode, err := newCode(false, code, inputRule.PreContext, inputRule.SortKey, context)
				if err != nil {
					t.Fatalf("unexpected error on code %v", code)
				}
				expectedOpCodes := rule.action
				if !reflect.DeepEqual(loadedCode.opCodes, expectedOpCodes) {
					t.Fatalf("expected %v, got %v", expectedOpCodes, loadedCode.opCodes)
				}
			}

			if len(rule.constraint) != 0 { // test constraint rule
				code := inputRule.Constraint
				loadedCode, err := newCode(true, code, inputRule.PreContext, inputRule.SortKey, context)
				if err != nil {
					t.Fatalf("unexpected error on code %v", code)
				}
				expectedOpCodes := rule.constraint
				if !reflect.DeepEqual(loadedCode.opCodes, expectedOpCodes) {
					t.Fatalf("pass %d, rule %d : expected %v, got %v", i, j, expectedOpCodes, loadedCode.opCodes)
				}
			}

		}
	}
}
