package harfbuzz

import (
	"log"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

func check(err error) {
	if err != nil {
		log.Fatal("unexpected error:", err)
	}
}

func assert(t *testing.T, cond bool) {
	if !cond {
		t.Fatal("assertion error")
	}
}

func assertEqualInt(t *testing.T, expected, got int) {
	if expected != got {
		t.Fatalf("expected %d, got %d", expected, got)
	}
}

// opens truetype fonts
func openFontFile(filename string) *truetype.Font {
	f, err := os.Open(filename)
	check(err)

	font, err := truetype.Parse(f)
	check(err)

	return font
}

func TestDirection(t *testing.T) {
	assert(t, LeftToRight.isHorizontal() && !LeftToRight.isVertical())
	assert(t, RightToLeft.isHorizontal() && !RightToLeft.isVertical())
	assert(t, !TopToBottom.isHorizontal() && TopToBottom.isVertical())
	assert(t, !BottomToTop.isHorizontal() && BottomToTop.isVertical())

	assert(t, LeftToRight.isForward())
	assert(t, TopToBottom.isForward())
	assert(t, !RightToLeft.isForward())
	assert(t, !BottomToTop.isForward())

	assert(t, !LeftToRight.isBackward())
	assert(t, !TopToBottom.isBackward())
	assert(t, RightToLeft.isBackward())
	assert(t, BottomToTop.isBackward())

	assert(t, BottomToTop.reverse() == TopToBottom)
	assert(t, TopToBottom.reverse() == BottomToTop)
	assert(t, LeftToRight.reverse() == RightToLeft)
	assert(t, RightToLeft.reverse() == LeftToRight)
}

func TestFlag(t *testing.T) {
	if (glyphFlagDefined & (glyphFlagDefined + 1)) != 0 {
		t.Error("assertion failed")
	}
}

func TestTypesLanguage(t *testing.T) {
	fa := NewLanguage("fa")
	faIR := NewLanguage("fa_IR")
	faIr := NewLanguage("fa-ir")
	en := NewLanguage("en")

	assert(t, fa != "")
	assert(t, faIR != "")
	assert(t, faIR == faIr)

	assert(t, en != "")
	assert(t, en != fa)

	/* Test recall */
	assert(t, en == NewLanguage("en"))
	assert(t, en == NewLanguage("eN"))
	assert(t, en == NewLanguage("En"))

	assert(t, NewLanguage("") == "")
	assert(t, NewLanguage("e") != "")
}
