package pango

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestBasicParse(t *testing.T) {
	a := "<b>bold <big>big</big> <i>italic</i></b> <s>strikethrough<sub>sub</sub> <small>small</small><sup>sup</sup></s> <tt>tt <u>underline</u></tt>"
	var out markupData
	err := xml.Unmarshal([]byte("<markup>"+a+"</markup>"), &out)
	if err != nil {
		t.Fatal(err)
	}
}

func testParseMarkup(t *testing.T, filename string) {
	contents, err := ioutil.ReadFile(filename + ".markup")
	if err != nil {
		t.Fatal(err)
	}
	var (
		out  string
		lang Language
	)
	ret, err := pango_parse_markup(contents, 0)
	if err == nil {
		out += string(ret.Text)
		out += "\n\n---\n\n"
		out += print_attr_list(ret.Attr)
		out += "\n\n---\n\n"
		desc := pango_font_description_new()
		iter := ret.Attr.pango_attr_list_get_iterator()
		do := true
		for do {
			iter.pango_attr_iterator_get_font(&desc, &lang, nil)
			// the C null Language is written (null)
			if lang == "" {
				lang = "(null)"
			}
			str := desc.String()
			out += fmt.Sprintf("[%d:%d] %s %s\n", iter.StartIndex, iter.EndIndex, lang, str)
			do = iter.pango_attr_iterator_next()
		}
	} else {
		out += fmt.Sprintf("ERROR: %s", err.Error())
	}

	if err := diffWithFile(out, filename+".expected"); err != nil {
		t.Fatalf("file %s: %s", filename, err)
	}
}

func TestParseMarkup(t *testing.T) {
	err := os.Setenv("LC_ALL", "C")
	if err != nil {
		t.Fatal(err)
	}

	// setlocale(LC_ALL, "")

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
