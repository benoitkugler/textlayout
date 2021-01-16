package pango

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/benoitkugler/go-weasyprint/layout/text/unicodedata"
)

func maxs(as ...int) int {
	m := 0
	for _, a := range as {
		if a > m {
			m = a
		}
	}
	return m
}

func testFile(filename string) (string, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	lines := string(contents)

	/* Skip initial comments */
	for strings.HasPrefix(lines, "#") {
		lines = strings.SplitN(lines, "\n", 2)[1]
	}
	text := []rune(lines)

	attrs := GetLogAttrs(text, -1)
	s1 := "Breaks: "
	s2 := "Whitespace: "
	s3 := "Words:"
	s4 := "Sentences:"

	st := "Text: "

	m := maxs(len(s1), len(s2), len(s2), len(s2), len(st))

	s1 += strings.Repeat(" ", m-len(s1))
	s2 += strings.Repeat(" ", m-len(s2))
	s3 += strings.Repeat(" ", m-len(s3))
	s4 += strings.Repeat(" ", m-len(s4))
	st += strings.Repeat(" ", m-len(st))

	for i, log := range attrs {
		b := 0
		w := 0
		o := 0
		s := 0

		// FIXME: the test ref from pango doesn't have a line
		// break at the end, maybe because it uses the full layout function ?
		if log.IsMandatoryBreak() && i != len(attrs)-1 {
			s1 += "L"
			b++
		} else if log.IsLineBreak() && i != len(attrs)-1 {
			s1 += "l"
			b++
		}
		if log.IsCharBreak() {
			s1 += "c"
			b++
		}

		if log.IsExpandableSpace() {
			s2 += "x"
			w++
		} else if log.IsWhite() {
			s2 += "w"
			w++
		}

		if log.IsWordBoundary() {
			s3 += "b"
			o++
		}
		if log.IsWordStart() {
			s3 += "s"
			o++
		}
		if log.IsWordEnd() {
			s3 += "e"
			o++
		}

		if log.IsSentenceBoundary() {
			s4 += "b"
			s++
		}
		if log.IsSentenceStart() {
			s4 += "s"
			s++
		}
		if log.IsSentenceEnd() {
			s4 += "e"
			s++
		}

		m = maxs(b, w, o, s)

		st += strings.Repeat(" ", m)
		s1 += strings.Repeat(" ", m-b)
		s2 += strings.Repeat(" ", m-w)
		s3 += strings.Repeat(" ", m-o)
		s4 += strings.Repeat(" ", m-s)

		if i < len(text) {
			ch := text[i]
			if ch == 0x20 {
				st += "[ ]"
				s1 += "   "
				s2 += "   "
				s3 += "   "
				s4 += "   "
			} else if unicode.IsPrint(ch) &&
				!(unicode.Is(unicode.Zl, ch) || unicode.Is(unicode.Zp, ch)) {
				st += string(ch)
				s1 += " "
				s2 += " "
				s3 += " "
				s4 += " "
			} else {
				str := fmt.Sprintf("[%#02x]", ch)
				st += str
				s1 += strings.Repeat(" ", len(str))
				s2 += strings.Repeat(" ", len(str))
				s3 += strings.Repeat(" ", len(str))
				s4 += strings.Repeat(" ", len(str))
			}
		}
	}
	st += "\n"
	st += s1
	st += "\n"
	st += s2
	st += "\n"
	st += s3
	st += "\n"
	st += s4
	st += "\n"
	return st, nil
}

func TestBreaks(t *testing.T) {
	files := [...]string{
		"test/breaks/one",
		"test/breaks/two",
		"test/breaks/three",
		// "test/four", we dont support language specific tailoring
	}
	for _, file := range files {
		s, err := testFile(file + ".break")
		if err != nil {
			t.Fatal(err)
		}

		exp, err := ioutil.ReadFile(file + ".expected")
		if err != nil {
			t.Fatal(err)
		}
		if s != string(exp) {
			fmt.Println(s)
			fmt.Println(string(exp))
			t.Errorf("file %s", file)
		}
	}
}

type charForeachFunc = func(t *testing.T,
	wc, prevWc, nextWc rune,
	attr, prevAttr, nextAttr *CharAttr,
)

func logAttrForeach(t *testing.T, text []rune, attrs []CharAttr, fn charForeachFunc) {
	if len(text) == 0 {
		return
	}
	for i, wc := range text {
		var prevWc, nextWc rune
		var prevAttr, nextAttr *CharAttr
		if i+1 < len(text) {
			nextWc = text[i+1]
		}
		if i > 0 {
			prevWc = text[i-1]
			prevAttr = &attrs[i-1]
		}
		if i+1 < len(attrs) {
			nextAttr = &attrs[i+1]
		}
		fn(t, wc, prevWc, nextWc,
			&attrs[i], prevAttr, nextAttr)
	}
}

func checkLineChar(t *testing.T,
	wc, prevWc, nextWc rune,
	attr, prevAttr, nextAttr *CharAttr,
) {

	prevBreakType := unicodedata.BreakXX
	_, breakType := unicodedata.BreakClass(wc)
	if prevWc != 0 {
		_, prevBreakType = unicodedata.BreakClass(prevWc)
	}

	if wc == '\n' {
		if prevWc == '\r' {
			assertFalse(t, attr.IsLineBreak(), "Do not line break between \\r and \\n")
		}

		if nextAttr != nil {
			assertTrue(t, nextAttr.IsLineBreak(), "Line break after \\n")
		}
	}

	if attr.IsLineBreak() {
		assertFalse(t, prevWc == 0, "first char in string should not be marked as a line break")
	}

	if breakType == unicodedata.BreakSP {
		_, nextBreak := unicodedata.BreakClass(nextWc)
		assertFalse(t, attr.IsLineBreak() && prevAttr != nil &&
			!attr.IsMandatoryBreak() &&
			!(nextWc != 0 && nextBreak == unicodedata.BreakCM),
			fmt.Sprintf("can't break lines before a space unless a mandatory break char precedes it or a combining mark follows; prev char was: %0#6x", prevWc),
		)
	}

	if attr.IsMandatoryBreak() {
		assertTrue(t, attr.IsLineBreak(), "mandatory breaks must also be marked as regular breaks")
	}

	assertFalse(t,
		breakType == unicodedata.BreakOP &&
			prevBreakType == unicodedata.BreakOP &&
			attr.IsLineBreak() &&
			!attr.IsMandatoryBreak(),
		"can't break between two open punctuation chars")

	assertFalse(t,
		breakType == unicodedata.BreakCL &&
			prevBreakType == unicodedata.BreakCL &&
			attr.IsLineBreak() &&
			!attr.IsMandatoryBreak(),
		"can't break between two close punctuation chars")

	assertFalse(t,
		breakType == unicodedata.BreakQU &&
			prevBreakType == unicodedata.BreakAL &&
			attr.IsLineBreak() &&
			!attr.IsMandatoryBreak(),
		"can't break letter-quotemark sequence")
}

func TestBoundaries(t *testing.T) {
	b, err := ioutil.ReadFile("test/breaks/boundaries.utf8")
	if err != nil {
		t.Fatal(err)
	}

	text := []rune(string(b))

	attrs := GetLogAttrs(text, -1)

	logAttrForeach(t, text, attrs, checkLineChar) // line invariants
}

func parse_line(line string) (string, []bool, error) {
	var attrReturn []bool
	var gs string

	for _, field := range strings.Fields(line) {
		switch field {
		case string(rune(0x00f7)): /* DIVISION SIGN: boundary here */
			attrReturn = append(attrReturn, true)
		case string(rune(0x00d7)): /* MULTIPLICATION SIGN: no boundary here */
			attrReturn = append(attrReturn, false)
		case "#":
			break
		default:
			character, err := strconv.ParseUint(field, 16, 32)
			if err != nil {
				return "", nil, err
			}
			if character > 0x10ffff {
				return "", nil, fmt.Errorf("unexpected character")
			}
			gs += string(rune(character))
		}
	}
	return gs, attrReturn, nil
}

// check that as the flags in ref set according to b
func isEqual(b bool, v CharAttr, ref CharAttr) bool {
	if b {
		return v&ref == ref
	}
	return v&ref == 0
}

func assertEqualAttrs(t *testing.T, bools []bool, attrs []CharAttr, ref CharAttr) {
	if len(bools) != len(attrs) {
		t.Fatalf("exepected length %d, got %d", len(bools), len(attrs))
	}
	for i, a := range attrs {
		if !isEqual(bools[i], a, ref) {
			t.Errorf("wrong value at index %d: %d", i, a)
		}
	}
}

func TestUCD(t *testing.T) {
	files := [...]string{
		"test/breaks/GraphemeBreakTest.txt",
		"test/breaks/EmojiBreakTest.txt",
		"test/breaks/CharBreakTest.txt",
		"test/breaks/WordBreakTest.txt",
		"test/breaks/SentenceBreakTest.txt",
		"test/breaks/LineBreakTest.txt",
	}
	refs := [...]CharAttr{
		CursorPosition,
		CursorPosition,
		CharBreak,
		WordBoundary,
		SentenceBoundary,
		LineBreak | MandatoryBreak,
	}
	for i, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			if len(line) == 0 || (line[0] != 0xf7 && line[0] != 0xd7) {
				continue
			}
			s, bs, err := parse_line(line)
			if err != nil {
				t.Fatal(err)
			}
			got := GetLogAttrs([]rune(s), -1)
			assertEqualAttrs(t, bs, got, refs[i])
		}
	}
}
