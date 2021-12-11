package pango_test

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/benoitkugler/textlayout/pango"
	"github.com/benoitkugler/textlayout/unicodedata"
)

func assertFalse(t *testing.T, b bool, message string) {
	t.Helper()
	if b {
		t.Fatal(message + ": expected false, got true")
	}
}

func assertTrue(t *testing.T, b bool, message string) {
	t.Helper()
	if !b {
		t.Fatal(message + ": expected true, got false")
	}
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

	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)
	layout.SetMarkup([]byte(lines))

	text := layout.Text

	attrs := layout.GetCharacterAttributes()
	if err := pango.ValidateCharacterAttributes(text, attrs); err != nil {
		return "", err
	}

	s1 := "Breaks: "
	s2 := "Whitespace: "
	s3 := "Sentences:"
	s4 := "Words:"
	s5 := "Graphemes:"
	s6 := "Hyphens:"

	st := "Text: "

	m := maxs(len(s1), len(s2), len(s3), len(s5), len(s5))

	s1 += strings.Repeat(" ", m-len(s1))
	s2 += strings.Repeat(" ", m-len(s2))
	s3 += strings.Repeat(" ", m-len(s3))
	s4 += strings.Repeat(" ", m-len(s4))
	s5 += strings.Repeat(" ", m-len(s5))
	s6 += strings.Repeat(" ", m-len(s6))
	st += strings.Repeat(" ", m-len(st))

	for i, log := range attrs {
		b := 0
		w := 0
		o := 0
		s := 0
		g := 0
		h := 0

		if log.IsMandatoryBreak() {
			s1 += "L"
			b++
		} else if log.IsLineBreak() {
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

		if log.IsSentenceBoundary() {
			s3 += "b"
			s++
		}
		if log.IsSentenceStart() {
			s3 += "s"
			s++
		}
		if log.IsSentenceEnd() {
			s3 += "e"
			s++
		}

		if log.IsWordBoundary() {
			s4 += "b"
			o++
		}
		if log.IsWordStart() {
			s4 += "s"
			o++
		}
		if log.IsWordEnd() {
			s4 += "e"
			o++
		}
		if log.IsCursorPosition() {
			s5 += "b"
			g++
		}
		if log.IsBreakRemovesPreceding() {
			s6 += "r"
			h++
		}
		if log.IsBreakInsertsHyphen() {
			s6 += "i"
			h++
		}

		m = maxs(b, w, o, s, g, h)

		st += strings.Repeat(" ", m)
		s1 += strings.Repeat(" ", m-b)
		s2 += strings.Repeat(" ", m-w)
		s3 += strings.Repeat(" ", m-s)
		s4 += strings.Repeat(" ", m-o)
		s5 += strings.Repeat(" ", m-g)
		s6 += strings.Repeat(" ", m-h)

		if i < len(text) {
			ch := text[i]
			if ch == 0x20 {
				st += "[ ]"
				s1 += "   "
				s2 += "   "
				s3 += "   "
				s4 += "   "
				s5 += "   "
				s6 += "   "
			} else if unicode.IsPrint(ch) &&
				!(unicode.Is(unicode.Zl, ch) || unicode.Is(unicode.Zp, ch)) {
				st += string(rune(0x2066))
				st += string(ch)
				st += string(rune(0x2069))
				s1 += " "
				s2 += " "
				s3 += " "
				s4 += " "
				s5 += " "
				s6 += " "
			} else {
				str := fmt.Sprintf("[%#02x]", ch)
				st += str
				s1 += strings.Repeat(" ", len(str))
				s2 += strings.Repeat(" ", len(str))
				s3 += strings.Repeat(" ", len(str))
				s4 += strings.Repeat(" ", len(str))
				s5 += strings.Repeat(" ", len(str))
				s6 += strings.Repeat(" ", len(str))
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
	st += s5
	st += "\n"
	st += s6
	st += "\n"
	return st, nil
}

func TestBreaks(t *testing.T) {
	files := [...]string{
		"test/breaks/one",
		"test/breaks/two",
		"test/breaks/three",
		// "test/breaks/four", we dont support tailored break for thai language
		"test/breaks/five",
		"test/breaks/six",
		"test/breaks/seven",
		"test/breaks/eight",
		"test/breaks/nine",
		"test/breaks/ten",
		"test/breaks/eleven",
		"test/breaks/twelve",
		"test/breaks/thirteen",
		"test/breaks/fourteen",
		"test/breaks/fifteen",
		"test/breaks/sixteen",
		"test/breaks/seventeen",
	}
	for _, file := range files {
		s, err := testFile(file + ".break")
		if err != nil {
			t.Fatalf("file %s: %s", file, err)
		}

		exp, err := ioutil.ReadFile(file + ".expected")
		if err != nil {
			t.Fatal(err)
		}
		// pango actually compare without case
		if sGot, sExp := strings.ToLower(s), strings.ToLower(string(exp)); sGot != sExp {
			t.Fatalf("break for file %s expected\n%s\n got \n%s", file, sExp, sGot)
		}
	}
}

type charForeachFunc = func(t *testing.T,
	wc, prevWc, nextWc rune,
	attr, prevAttr, nextAttr *pango.CharAttr,
)

func logAttrForeach(t *testing.T, text []rune, attrs []pango.CharAttr, fn charForeachFunc) {
	if len(text) == 0 {
		return
	}
	for i, wc := range text {
		var prevWc, nextWc rune
		var prevAttr, nextAttr *pango.CharAttr
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
	attr, prevAttr, nextAttr *pango.CharAttr,
) {
	prevBreakType := unicodedata.BreakXX
	breakType := unicodedata.LookupBreakClass(wc)
	if prevWc != 0 {
		prevBreakType = unicodedata.LookupBreakClass(prevWc)
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
		nextBreak := unicodedata.LookupBreakClass(nextWc)
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

	attrs := pango.ComputeCharacterAttributes(text, -1)

	logAttrForeach(t, text, attrs, checkLineChar) // line invariants
}

func parseLine(line string) (string, []bool, error) {
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
func isEqual(b bool, v pango.CharAttr, ref pango.CharAttr) bool {
	if b {
		return v&ref == ref
	}
	return v&ref == 0
}

func assertEqualAttrs(t *testing.T, bools []bool, attrs []pango.CharAttr, ref pango.CharAttr) {
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
	refs := [...]pango.CharAttr{
		pango.CursorPosition,
		pango.CursorPosition,
		pango.CharBreak,
		pango.WordBoundary,
		pango.SentenceBoundary,
		pango.LineBreak | pango.MandatoryBreak,
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
			s, bs, err := parseLine(line)
			if err != nil {
				t.Fatal(err)
			}
			got := pango.ComputeCharacterAttributes([]rune(s), -1)
			assertEqualAttrs(t, bs, got, refs[i])
		}
	}
}
