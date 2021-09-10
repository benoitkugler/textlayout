package pango

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

var (
	lineNoRe = regexp.MustCompile(`line [\d]+`)
	charNoRe = regexp.MustCompile(`char [\d]+`)
)

func replaceLineColNumbers(s string) string {
	s = lineNoRe.ReplaceAllString(s, fmt.Sprintf("line %d", lineColNumber))
	s = charNoRe.ReplaceAllString(s, fmt.Sprintf("char %d", lineColNumber))
	return s
}

func TestBasicParse(t *testing.T) {
	a := "<b>bold <big>big</big> <i>italic</i></b> <s>strikethrough<sub>sub</sub> <small>small</small><sup>sup</sup></s> <tt>tt <u>underline</u></tt>"
	var out markupData
	err := xml.Unmarshal([]byte("<markup>"+a+"</markup>"), &out)
	if err != nil {
		t.Fatal(err)
	}
}

// the C references tests have index in byte index
func asByteIndex(text []rune, runeIndex int) int {
	if runeIndex == MaxInt {
		return MaxInt
	}
	return len(string(text[:runeIndex]))
}

func testParseMarkup(t *testing.T, filename string) {
	contents, err := ioutil.ReadFile(filename + ".markup")
	if err != nil {
		t.Fatal(err)
	}
	var lang Language
	ret, err := ParseMarkup(contents, '_')
	if err == nil {
		out := string(ret.Text)
		out += "\n\n---\n\n"
		out += PrintAttributes(ret.Attr, ret.Text)
		out += "\n\n---\n\n"
		desc := NewFontDescription()
		iter := ret.Attr.getIterator()

		for do := true; do; do = iter.next() {
			iter.getFont(&desc, &lang, nil)
			// the C null Language is written (null)
			if lang == "" {
				lang = "(null)"
			}
			str := desc.String()
			out += fmt.Sprintf("[%d:%d] %s %s\n",
				asByteIndex(ret.Text, iter.StartIndex), asByteIndex(ret.Text, iter.EndIndex), lang, str)

		}
		if ret.AccelChar != 0 {
			out += "\n\n---\n\n"
			out += string(ret.AccelChar)
		}

		if err = diffWithFile(out, filename+".expected"); err != nil {
			t.Fatalf("file %s: %s", filename, err)
		}
	} else {
		out := fmt.Sprintf("ERROR: %s", err.Error())

		// in case of error, the line numbers are not yet correct
		// so we ignore the values in test
		b, err := ioutil.ReadFile(filename + ".expected")
		if err != nil {
			t.Fatalf("file %s: %s", filename, err)
		}

		expected := replaceLineColNumbers(string(b))
		if out != expected {
			t.Fatalf("file %s: expected\n%s\n, got\n%s", filename, expected, out)
		}
	}
}

func TestParseMarkup(t *testing.T) {
	err := os.Setenv("LC_ALL", "C")
	if err != nil {
		t.Fatal(err)
	}

	files, err := ioutil.ReadDir("test/markups")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".markup") {
			continue
		}
		testParseMarkup(t, "test/markups/"+strings.TrimSuffix(file.Name(), ".markup"))
	}
}
