package pango

import (
	"testing"

	"github.com/benoitkugler/textlayout/language"
)

func TestSampleWidth(t *testing.T) {
	// assert that the sample returned is always not empty
	assert := func(l Language) {
		s := SampleString(l)
		if len([]rune(s)) == 0 {
			t.Errorf("empty sample for language %s", l)
		}
	}
	for _, rec := range langTexts {
		l := rec.language()
		assert(l)
	}

	// test also the default language
	assert(Language(""))
}

func TestMatchOwn(t *testing.T) {
	for _, rec := range langTexts {
		l := rec.language()
		if !pangoLanguageMatches(l, string(l)) {
			t.Errorf("language %s should match itself", l)
		}
	}
}

func TestLanguageEmojiCrash(t *testing.T) {
	lang := language.NewLanguage("und-zsye")
	pango_language_get_scripts(lang)
}
