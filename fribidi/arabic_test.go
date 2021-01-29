package fribidi

import "testing"

func TestArabicShape(t *testing.T) {
	r := getArabicShapePres(1604, 0)
	if r != 65245 {
		t.Errorf("expected 65245, got %d", r)
	}
}
