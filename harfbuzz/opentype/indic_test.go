package opentype

import "testing"

func TestIndic(t *testing.T) {
	if INDIC_SYLLABIC_CATEGORY_AVAGRAHA != OT_Symbol {
		t.Error("assertion error")
	}
}
