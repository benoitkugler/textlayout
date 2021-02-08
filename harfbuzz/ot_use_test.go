package harfbuzz

import "testing"

func TestUSE(t *testing.T) {
	if !(JOINING_FORM_INIT < 4 && JOINING_FORM_ISOL < 4 && JOINING_FORM_MEDI < 4 && JOINING_FORM_FINA < 4) {
		t.Error()
	}
}
