package opentype

import (
	"strings"
	"testing"
)

func cmp(a, b string) int {
	da := len(a)
	if p := strings.IndexByte(a, '-'); p != -1 {
		da = p
	}

	db := len(b)
	if p := strings.IndexByte(b, '-'); p != -1 {
		db = p
	}
	m := max(da, db)
	if len(a) > m {
		a = a[:m]
	}
	if len(b) > m {
		b = b[:m]
	}
	return strings.Compare(a, b)
}

func TestLanguageOrder(t *testing.T) {
	for i, l := range ot_languages {
		if i == 0 {
			continue
		}
		c := cmp(ot_languages[i-1].language, l.language)
		if c > 0 {
			t.Fatalf("ot_languages not sorted at index %d: %s %d %s\n",
				i, ot_languages[i-1].language, c, l.language)
		}
	}
}
