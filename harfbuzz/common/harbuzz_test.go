package common

import "testing"

func TestDirection(t *testing.T) {
	if !HB_DIRECTION_LTR.IsHorizontal() || HB_DIRECTION_LTR.IsVertical() {
		t.Error("wrong direction")
	}
	if !HB_DIRECTION_RTL.IsHorizontal() || HB_DIRECTION_RTL.IsVertical() {
		t.Error("wrong direction")
	}
	if HB_DIRECTION_TTB.IsHorizontal() || !HB_DIRECTION_TTB.IsVertical() {
		t.Error("wrong direction")
	}
	if HB_DIRECTION_BTT.IsHorizontal() || !HB_DIRECTION_BTT.IsVertical() {
		t.Error("wrong direction")
	}

	if HB_DIRECTION_BTT.reverse() != HB_DIRECTION_TTB {
		t.Error("wrong reverse")
	}
	if HB_DIRECTION_TTB.reverse() != HB_DIRECTION_BTT {
		t.Error("wrong reverse")
	}
	if HB_DIRECTION_LTR.reverse() != HB_DIRECTION_LTR {
		t.Error("wrong reverse")
	}
	if HB_DIRECTION_LTR.reverse() != HB_DIRECTION_LTR {
		t.Error("wrong reverse")
	}
}

func TestFlag(t *testing.T) {
	if (HB_GLYPH_FLAG_DEFINED & (HB_GLYPH_FLAG_DEFINED + 1)) != 0 {
		t.Error("assertion failed")
	}
}
