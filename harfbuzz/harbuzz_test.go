package harfbuzz

import "testing"

func TestDirection(t *testing.T) {
	if !LeftToRight.IsHorizontal() || LeftToRight.IsVertical() {
		t.Error("wrong direction")
	}
	if !RightToLeft.IsHorizontal() || RightToLeft.IsVertical() {
		t.Error("wrong direction")
	}
	if TopToBottom.IsHorizontal() || !TopToBottom.IsVertical() {
		t.Error("wrong direction")
	}
	if BottomToTop.IsHorizontal() || !BottomToTop.IsVertical() {
		t.Error("wrong direction")
	}

	if BottomToTop.reverse() != TopToBottom {
		t.Error("wrong reverse")
	}
	if TopToBottom.reverse() != BottomToTop {
		t.Error("wrong reverse")
	}
	if LeftToRight.reverse() != LeftToRight {
		t.Error("wrong reverse")
	}
	if LeftToRight.reverse() != LeftToRight {
		t.Error("wrong reverse")
	}
}

func TestFlag(t *testing.T) {
	if (HB_GLYPH_FLAG_DEFINED & (HB_GLYPH_FLAG_DEFINED + 1)) != 0 {
		t.Error("assertion failed")
	}
}
