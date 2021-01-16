package pango

import (
	"testing"
)

func TestScriptIter(t *testing.T) {
	var test_data = [...]struct {
		text []rune
		code Script
	}{
		{[]rune("\u0020\u0946\u0939\u093F\u0928\u094D\u0926\u0940\u0020"), SCRIPT_DEVANAGARI},
		{[]rune("\u0627\u0644\u0639\u0631\u0628\u064A\u0629\u0020"), SCRIPT_ARABIC},
		{[]rune("\u0420\u0443\u0441\u0441\u043A\u0438\u0439\u0020"), SCRIPT_CYRILLIC},
		{[]rune("English ("), SCRIPT_LATIN},
		{[]rune("\u0E44\u0E17\u0E22"), SCRIPT_THAI},
		{[]rune(") "), SCRIPT_LATIN},
		{[]rune("\u6F22\u5B75"), SCRIPT_HAN},
		{[]rune("\u3068\u3072\u3089\u304C\u306A\u3068"), SCRIPT_HIRAGANA},
		{[]rune("\u30AB\u30BF\u30AB\u30CA"), SCRIPT_KATAKANA},
		{[]rune("\U00010400\U00010401\U00010402\U00010403"), SCRIPT_DESERET},
	}

	var all []rune
	for _, td := range test_data {
		all = append(all, td.text...)
	}

	iter := pango_script_iter_new(all)

	pos := 0
	for i, td := range test_data {
		next_pos := pos + len(td.text)

		start, end, script := iter.script_start, iter.script_end, iter.script_code

		assertTrue(t, start == pos, "start position")
		assertTrue(t, end == next_pos, "end position")
		assertTrue(t, script == td.code, "script code")

		result := iter.pango_script_iter_next()
		assertTrue(t, result == (i != len(test_data)-1), "has next script")

		pos = next_pos
	}
}

func TestEmptyScript(t *testing.T) { // Test an empty string.
	iter := pango_script_iter_new(nil)
	start, end, script := iter.script_start, iter.script_end, iter.script_code

	assertTrue(t, start == 0, "start is at begining")
	assertTrue(t, end == 0, "end is at begining")
	assertTrue(t, script == SCRIPT_COMMON, "script is common")
	assertFalse(t, iter.pango_script_iter_next(), "has no more script")
}
