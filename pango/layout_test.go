package pango_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/benoitkugler/textlayout/pango"
)

type layoutParams struct {
	width       int
	ellipsizeAt int
	ellipsize   pango.EllipsizeMode
	wrap        pango.WrapMode
}

func parseEllipsizeMode(v string) pango.EllipsizeMode {
	switch strings.ToLower(v) {
	case "none":
		return pango.ELLIPSIZE_NONE
	case "start":
		return pango.ELLIPSIZE_START
	case "middle":
		return pango.ELLIPSIZE_MIDDLE
	case "end":
		return pango.ELLIPSIZE_END
	default:
		return 0
	}
}

func parseWrapMode(v string) pango.WrapMode {
	switch strings.ToLower(v) {
	case "word":
		return pango.WRAP_WORD
	case "char":
		return pango.WRAP_CHAR
	case "word_char":
		return pango.WRAP_WORD_CHAR
	default:
		return 0
	}
}

func parse_params(params string) (out layoutParams) {
	if strings.TrimSpace(params) == "" {
		return out
	}
	options := strings.Split(params, ",")
	for _, option := range options {
		chunks := strings.Split(option, "=")
		name, value := chunks[0], chunks[1]
		//   str2 = g_strsplit (strings[i], "=", -1);
		switch name {
		case "width":
			out.width, _ = strconv.Atoi(value)
		case "ellipsize-at":
			out.ellipsizeAt, _ = strconv.Atoi(value)
		case "ellipsize":
			out.ellipsize = parseEllipsizeMode(value)
		case "wrap":
			out.wrap = parseWrapMode(value)
		}
	}
	return out
}

func dumpLines(layout *pango.Layout) string {
	var out string

	text := layout.Text
	iter := layout.GetIter()

	index := iter.Index
	indexEnd := 0
	i := 0
	for hasMore := true; hasMore; {
		line := iter.GetLine()
		hasMore = iter.NextLine()
		i++

		var charStr []rune
		if hasMore {
			indexEnd = iter.Index
			charStr = text[index:indexEnd]
		} else {
			charStr = text[index:]
		}

		// the reference tests have index in byte
		indexByte := len(string(text[:index]))
		out += fmt.Sprintf("i=%d, index=%d, paragraph-start=%d, dir=%s '%s'\n",
			i, indexByte, boolToInt(line.IsParagraphStart), line.ResolvedDir, string(charStr))

		index = indexEnd
	}

	return out
}

func dumpRuns(layout *pango.Layout) string {
	var out string

	text := layout.Text
	iter := layout.GetIter()

	i := 0
	for hasMore := true; hasMore; {
		run := iter.GetRun()
		index := iter.Index
		hasMore = iter.NextRun()
		i++

		// the reference tests have index in byte
		indexByte := len(string(text[:index]))

		if run != nil {
			item := run.Item
			charStr := string(text[item.Offset : item.Offset+item.Length])
			// font := item.Analysis.Font.Describe(false).String()
			out += fmt.Sprintf("i=%d, index=%d, chars=%d, level=%d, gravity=%s, flags=%d, font=%s, script=%s, language=%s, '%s'\n",
				i, indexByte, item.Length, item.Analysis.Level, item.Analysis.Gravity,
				item.Analysis.Flags,
				"OMITTED", /* for some reason, this fails on build.gnome.org, so leave it out */
				strings.ToLower(item.Analysis.Script.String()), item.Analysis.Language, charStr)
			out += printAttributes(item.Analysis.ExtraAttrs)
		} else {
			out += fmt.Sprintf("i=%d, index=%d, no run, line end\n",
				i, indexByte)
		}
	}

	return out
}

func printAttributes(attrs pango.AttrList) string {
	chunks := make([]string, len(attrs))
	for i, attr := range attrs {
		chunks[i] = attr.String() + "\n"
	}
	return strings.Join(chunks, "")
}

func testOneLayout(t *testing.T, filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	context := pango.NewContext(newChachedFontMap())

	chunks := strings.SplitN(string(content), "\n", 2)
	params, markup := chunks[0], chunks[1]

	options := parse_params(params)

	layout := pango.NewLayout(context)
	desc := pango.NewFontDescriptionFrom("Cantarell 11")
	layout.SetFontDescription(&desc)

	err = layout.SetMarkup([]byte(markup))
	if err != nil {
		t.Fatal(err)
	}

	if options.width != 0 {
		layout.SetWidth(pango.GlyphUnit(options.width) * pango.Scale)
	}
	layout.SetEllipsize(options.ellipsize)
	layout.SetWrap(options.wrap)

	out := string(layout.Text)
	out += "\n--- parameters\n\n"
	out += fmt.Sprintf("wrapped: %d\n", boolToInt(layout.IsWrapped()))
	out += fmt.Sprintf("ellipsized: %d\n", boolToInt(layout.IsEllipsized()))
	out += fmt.Sprintf("lines: %d\n", layout.GetLineCount())
	if options.width != 0 {
		out += fmt.Sprintf("width: %d\n", layout.Width)
	}
	out += "\n--- attributes\n\n"

	out += layout.Attributes.String()
	out += "\n--- lines\n\n"
	out += dumpLines(layout)
	out += "\n--- runs\n\n"
	out += dumpRuns(layout)

	return out
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestLayout(t *testing.T) {
	if err := os.Setenv("LC_ALL", "en_US.utf8"); err != nil {
		t.Fatal(err)
	}

	files, err := ioutil.ReadDir("test/layouts")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".markup") {
			continue
		}

		got := testOneLayout(t, filepath.Join("test/layouts", file.Name()))

		expFilename := strings.TrimSuffix(file.Name(), ".markup") + ".expected"
		exp, err := ioutil.ReadFile(filepath.Join("test/layouts", expFilename))
		if err != nil {
			t.Fatal(err)
		}

		if got != string(exp) {
			t.Fatalf("expected\n%s\n got\n%s", exp, got)
		}
	}
}

func TestShapeTabCrash(t *testing.T) {
	// test that we don't crash in shape_tab when the layout
	// is such that we don't have effective attributes
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)
	layout.SetText("one\ttwo")
	layout.IsEllipsized()
}

func TestItemizeEmpty(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	context.Itemize(nil, 0, 0, nil)
}
