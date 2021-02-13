package harfbuzz

import "testing"

func assert(t *testing.T, cond bool) {
	if !cond {
		t.Error("assertion error")
	}
}

func TestDirection(t *testing.T) {
	assert(t, LeftToRight.IsHorizontal() && !LeftToRight.IsVertical())
	assert(t, RightToLeft.IsHorizontal() && !RightToLeft.IsVertical())
	assert(t, !TopToBottom.IsHorizontal() && TopToBottom.IsVertical())
	assert(t, !BottomToTop.IsHorizontal() && BottomToTop.IsVertical())

	assert(t, LeftToRight.isForward())
	assert(t, TopToBottom.isForward())
	assert(t, !RightToLeft.isForward())
	assert(t, !BottomToTop.isForward())

	assert(t, !LeftToRight.IsBackward())
	assert(t, !TopToBottom.IsBackward())
	assert(t, RightToLeft.IsBackward())
	assert(t, BottomToTop.IsBackward())

	assert(t, BottomToTop.reverse() == TopToBottom)
	assert(t, TopToBottom.reverse() == BottomToTop)
	assert(t, LeftToRight.reverse() == LeftToRight)
	assert(t, LeftToRight.reverse() == LeftToRight)
}

func TestFlag(t *testing.T) {
	if (HB_GLYPH_FLAG_DEFINED & (HB_GLYPH_FLAG_DEFINED + 1)) != 0 {
		t.Error("assertion failed")
	}
}
