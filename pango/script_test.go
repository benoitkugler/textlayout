package pango

import (
	"reflect"
	"testing"

	"github.com/benoitkugler/textlayout/language"
)

func TestScriptIter(t *testing.T) {
	testData := [...]struct {
		text []rune
		code Script
	}{
		{[]rune("\u0020\u0946\u0939\u093F\u0928\u094D\u0926\u0940\u0020"), language.Devanagari},
		{[]rune("\u0627\u0644\u0639\u0631\u0628\u064A\u0629\u0020"), language.Arabic},
		{[]rune("\u0420\u0443\u0441\u0441\u043A\u0438\u0439\u0020"), language.Cyrillic},
		{[]rune("English ("), language.Latin},
		{[]rune("\u0E44\u0E17\u0E22"), language.Thai},
		{[]rune(") "), language.Latin},
		{[]rune("\u6F22\u5B75"), language.Han},
		{[]rune("\u3068\u3072\u3089\u304C\u306A\u3068"), language.Hiragana},
		{[]rune("\u30AB\u30BF\u30AB\u30CA"), language.Katakana},
		{[]rune("\U00010400\U00010401\U00010402\U00010403"), language.Deseret},
	}

	var all []rune
	for _, td := range testData {
		all = append(all, td.text...)
	}

	iter := newScriptIter(all)

	pos := 0
	for i, td := range testData {
		next_pos := pos + len(td.text)

		start, end, script := iter.scriptStart, iter.scriptEnd, iter.scriptCode

		assertTrue(t, start == pos, "start position")
		assertTrue(t, end == next_pos, "end position")
		assertTrue(t, script == td.code, "script code")

		result := iter.next()
		assertTrue(t, result == (i != len(testData)-1), "has next script")

		pos = next_pos
	}
}

func TestEmptyScript(t *testing.T) { // Test an empty string.
	iter := newScriptIter(nil)
	start, end, script := iter.scriptStart, iter.scriptEnd, iter.scriptCode

	assertTrue(t, start == 0, "start is at begining")
	assertTrue(t, end == 0, "end is at begining")
	assertTrue(t, script == language.Common, "script is common")
	assertFalse(t, iter.next(), "has no more script")
}

func TestScripts(t *testing.T) {
	type res struct {
		start, end int
		script     Script
	}
	text := []rune("Hello שלום Γειά σας")
	iter := newScriptIter(text)
	exps := []res{
		{0, 6, language.Latin},
		{6, 11, language.Hebrew},
		{11, 19, language.Greek},
	}
	var gots []res
	for do := true; do; do = iter.next() {
		gots = append(gots, res{iter.scriptStart, iter.scriptEnd, iter.scriptCode})
	}

	if !reflect.DeepEqual(exps, gots) {
		t.Fatalf("expected %v, got %v", exps, gots)
	}
}
