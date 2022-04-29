package graphite

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"reflect"
	"strings"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/graphite"
)

func TestCodeString(t *testing.T) {
	o := ocPUSH_BYTE
	if o.String() != "PUSH_BYTE" {
		t.Fatal()
	}
}

func TestSimpleInstructions(t *testing.T) {
	o := func(v opcode) byte { return byte(v) }
	valid := []byte{
		o(ocPUSH_BYTE), 43,
		o(ocPUSH_BYTE), 42,
		o(ocPUSH_BYTE), 11, o(ocPUSH_BYTE), 13, o(ocADD),
		o(ocPUSH_BYTE), 4, o(ocSUB),
		o(ocCOND),
		// o(PUSH_LONG), 0x80, 0x50, 0x00, 0x00,
		// o(PUSH_LONG), 0xff, 0xff, 0xff, 0xff,
		// o(DIV),
		o(ocPOP_RET),
	}
	simpleUnderflow := []byte{
		o(ocPUSH_BYTE), 43,
		o(ocPUSH_BYTE), 42,
		o(ocPUSH_BYTE), 11, o(ocPUSH_BYTE), 13, o(ocADD),
		o(ocPUSH_BYTE), 4, o(ocSUB),
		o(ocCOND),
		o(ocPUSH_LONG), 0x80, 0x00, 0x00, 0x00,
		o(ocPUSH_LONG), 0xff, 0xff, 0xff, 0xff,
		o(ocDIV),
		o(ocCOND), // Uncomment to cause an underflow
		o(ocPOP_RET),
	}

	invalidDiv := []byte{
		o(ocPUSH_BYTE), 43,
		o(ocPUSH_BYTE), 42,
		o(ocPUSH_LONG), 1, 2, 3, 4,
		o(ocPUSH_BYTE), 11, o(ocPUSH_BYTE), 13, o(ocADD),
		o(ocPUSH_BYTE), 4, o(ocSUB),
		o(ocCOND),
		o(ocPUSH_LONG), 0x80, 0x00, 0x00, 0x00,
		o(ocPUSH_LONG), 0xff, 0xff, 0xff, 0xff,
		o(ocDIV),
		o(ocPOP_RET),
	}

	context := codeContext{NumAttributes: 8, NumFeatures: 1}
	prog, err := newCode(false, valid, 0, 0, context, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = newCode(false, simpleUnderflow, 0, 0, context, false)
	if err != underfullStack {
		t.Fatalf("expected underfull error, got %s", err)
	}

	progOverflow, err := newCode(false, invalidDiv, 0, 0, context, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	sm := newSlotMap(new(Segment), false, 0)
	m := newMachine(&sm)
	m.map_.pushSlot(new(Slot))
	out, _, err := m.run(&prog, 1)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if out != 42 {
		t.Fatalf("expected 42, got %d", out)
	}

	if _, _, err := m.run(&progOverflow, 1); err != machineDiedEarly {
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
func readExpectedOpCodes(filename string) [][]expectedRule {
	data, err := testdata.Files.ReadFile(filename)
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

			for opc, data := range opcodeTable {
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

func instrsToOpcodes(l []instr) (out []opcode) {
	for _, v := range l {
		out = append(out, v.code)
	}
	return out
}

func TestOpCodesValues(t *testing.T) {
	testFontOpCodes(t, "Annapurnarc2")
	testFontOpCodes(t, "MagyarLinLibertineG")
}

func testFontOpCodes(t *testing.T, fontName string) {
	b, err := testdata.Files.ReadFile(fontName + "_bytecodes.json")
	if err != nil {
		t.Fatal(err)
	}
	var passes []extractedPass
	if err = json.Unmarshal(b, &passes); err != nil {
		t.Fatal(err)
	}

	expected := readExpectedOpCodes(fontName + ".ttx")
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
				loadedCode, err := newCode(false, code, inputRule.PreContext, inputRule.SortKey, context, true)
				if err != nil {
					t.Fatalf("unexpected error on code %v", code)
				}

				expectedOpCodes := rule.action
				if got := instrsToOpcodes(loadedCode.instrs); !reflect.DeepEqual(got, expectedOpCodes) {
					t.Fatalf("expected %v, got %v", expectedOpCodes, got)
				}
			}

			if len(rule.constraint) != 0 { // test constraint rule
				code := inputRule.Constraint
				loadedCode, err := newCode(true, code, inputRule.PreContext, inputRule.SortKey, context, true)
				if err != nil {
					t.Fatalf("unexpected error on code %v", code)
				}

				expectedOpCodes := rule.constraint
				if got := instrsToOpcodes(loadedCode.instrs); !reflect.DeepEqual(got, expectedOpCodes) {
					t.Fatalf("pass %d, rule %d : expected %v, got %v", i, j, expectedOpCodes, got)
				}
			}

		}
	}
}

func TestOpCodesAnalysis(t *testing.T) {
	ft := loadGraphite(t, "MagyarLinLibertineG.ttf")
	got := instrsToOpcodes(ft.silf[0].passes[0].ruleTable[194].action.instrs)
	// extracted from running graphite test magyar1
	expected := []opcode{
		ocTEMP_COPY, ocPUT_GLYPH, ocNEXT, ocTEMP_COPY, ocPUT_GLYPH, ocNEXT, ocINSERT, ocPUT_COPY,
		ocNEXT, ocINSERT, ocPUT_COPY, ocNEXT, ocINSERT, ocPUT_GLYPH, ocNEXT, ocINSERT, ocPUT_GLYPH, ocNEXT,
		ocPUSH_BYTE, ocPOP_RET,
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected\n%v\n got \n%v", expected, got)
	}
}
