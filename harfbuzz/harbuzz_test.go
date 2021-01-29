package harfbuzz

import "testing"

func TestDirection(t *testing.T) {
	if !HB_DIRECTION_LTR.isHorizontal() || HB_DIRECTION_LTR.isVertical() {
		t.Error("wrong direction")
	}
	if !HB_DIRECTION_RTL.isHorizontal() || HB_DIRECTION_RTL.isVertical() {
		t.Error("wrong direction")
	}
	if HB_DIRECTION_TTB.isHorizontal() || !HB_DIRECTION_TTB.isVertical() {
		t.Error("wrong direction")
	}
	if HB_DIRECTION_BTT.isHorizontal() || !HB_DIRECTION_BTT.isVertical() {
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
