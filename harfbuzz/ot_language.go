package harfbuzz

import (
	"strings"

	"github.com/benoitkugler/textlayout/harfbuzz/internal"
	"github.com/benoitkugler/textlayout/language"
)

type LangTag struct {
	language string
	tag      hb_tag_t
}

func bfindLanguage(lang string) int {
	low, high := 0, len(ot_languages)
	for low <= high {
		mid := (low + high) / 2
		p := ot_languages[mid]

		if lang < p.language {
			high = mid - 1
		} else if lang > p.language {
			low = mid + 1
		} else {
			return mid
		}
	}
	return -1
}

func subtagMatches(lang_str string, limit int, subtag string) bool {
	for {
		s := strings.Index(lang_str, subtag)
		if s == -1 || s >= limit {
			return false
		}
		if !internal.IsAlnum(lang_str[s+len(subtag)]) {
			return true
		}
		lang_str = lang_str[s+len(subtag):]
	}
}

func langMatches(lang_str, spec string) bool {
	l := len(spec)
	return strings.HasPrefix(lang_str, spec) && (len(lang_str) == l || lang_str[l] == '-')
}

// Converts `str` representing a BCP 47 language tag to the corresponding Language.
func hb_language_from_string(str string) internal.Language {
	return internal.Language(language.Canonicalize([]byte(str)))
}

// func hb_language_to_string(l Language) string { return string(l) }
