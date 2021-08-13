package pango_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/benoitkugler/textlayout/pango"
)

func escapeTextInput(text []rune) string {
	out := ""
	for _, r := range text {
		if r == 0x0A || r == 0x2028 || !unicode.IsPrint(r) {
			out += fmt.Sprintf("[0x%04x]", r)
		} else {
			out += string(r)
		}
	}
	return out
}

func stringByteIndex(attr pango.Attribute, text []rune) string {
	// to obtain the same result as the C implementation
	// we convert to int32
	byteStart := len(string(text[:attr.StartIndex]))
	byteEnd := len(string(text[:attr.EndIndex]))
	return fmt.Sprintf("[%d,%d]%s=%s", byteStart, byteEnd, attr.Kind, attr.Data)
}

func testOneItemize(t *testing.T, filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	context := pango.NewContext(newChachedFontMap())

	// Skip initial comments
	lines := bytes.Split(content, []byte("\n"))
	for len(lines) > 0 && bytes.HasPrefix(lines[0], []byte{'#'}) {
		lines = lines[1:]
	}
	test := bytes.Join(lines, nil)

	parsed, err := pango.ParseMarkup(test, 0)
	if err != nil {
		t.Fatal(err)
	}

	itemizeAttrs := parsed.Attr.Filter(affectsItemization)

	L := len(parsed.Text)
	if parsed.Text[L-1] == '\n' {
		L--
	}
	items := context.Itemize(parsed.Text, 0, L, itemizeAttrs)

	items.ApplyAttributes(parsed.Attr)

	s1 := "Items:  "
	s2 := "Font:   "
	s3 := "Script: "
	s4 := "Lang:   "
	s5 := "Bidi:   "
	s6 := "Attrs:  "

	for l := items; l != nil; l = l.Next {
		item := l.Data
		desc := item.Analysis.Font.Describe(false)
		var sep string
		if l != items {
			sep = "|"
		}
		s1 += sep
		s1 += escapeTextInput(parsed.Text[item.Offset : item.Offset+item.Length])

		s2 += fmt.Sprintf("%s%s", sep, desc)
		s3 += fmt.Sprintf("%s%s", sep, strings.ToLower(item.Analysis.Script.String()))
		s4 += fmt.Sprintf("%s%s", sep, item.Analysis.Language)
		s5 += fmt.Sprintf("%s%d", sep, item.Analysis.Level)
		s6 += sep

		for i, a := range item.Analysis.ExtraAttrs {
			if i != 0 {
				s6 += ","
			}
			s6 += stringByteIndex(*a, parsed.Text)
		}

		M := maxs(len(s1), len(s2), len(s3), len(s4), len(s5), len(s6))
		s1 += strings.Repeat(" ", M-len(s1))
		s2 += strings.Repeat(" ", M-len(s2))
		s3 += strings.Repeat(" ", M-len(s3))
		s4 += strings.Repeat(" ", M-len(s4))
		s5 += strings.Repeat(" ", M-len(s5))
		s6 += strings.Repeat(" ", M-len(s6))
	}

	out := string(test) + "\n\n" +
		s1 + "\n" +
		s2 + "\n" +
		s3 + "\n" +
		s4 + "\n" +
		s5 + "\n" +
		s6 + "\n"
	return out
}

func maxs(vs ...int) int {
	m := vs[0]
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

func affectsItemization(attr *pango.Attribute) bool {
	switch attr.Kind {
	/* These affect font selection */
	case pango.ATTR_LANGUAGE, pango.ATTR_FAMILY, pango.ATTR_STYLE, pango.ATTR_WEIGHT, pango.ATTR_VARIANT, pango.ATTR_STRETCH, pango.ATTR_SIZE, pango.ATTR_FONT_DESC,
		pango.ATTR_SCALE, pango.ATTR_FALLBACK, pango.ATTR_ABSOLUTE_SIZE, pango.ATTR_GRAVITY, pango.ATTR_GRAVITY_HINT,
		/* These are part of ItemProperties, so need to break runs */
		pango.ATTR_SHAPE, pango.ATTR_RISE, pango.ATTR_UNDERLINE, pango.ATTR_STRIKETHROUGH, pango.ATTR_LETTER_SPACING:
		return true
	default:
		return false
	}
}

func TestItemize(t *testing.T) {
	if err := os.Setenv("LC_ALL", "en_US.utf8"); err != nil {
		t.Fatal(err)
	}

	files, err := ioutil.ReadDir("test/itemize")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".items") {
			continue
		}

		got := testOneItemize(t, filepath.Join("test/itemize", file.Name()))

		expFilename := strings.TrimSuffix(file.Name(), ".items") + ".expected"
		exp, err := ioutil.ReadFile(filepath.Join("test/itemize", expFilename))
		if err != nil {
			t.Fatal(err)
		}

		if got != string(exp) {
			t.Fatalf("expected\n%s\n got\n%s", exp, got)
		}
	}
}
