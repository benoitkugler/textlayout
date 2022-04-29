package harfbuzz

import (
	"bytes"
	"log"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/harfbuzz"
	tttestdata "github.com/benoitkugler/textlayout-testdata/truetype"
	tt "github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
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

func assertEqualInt32(t *testing.T, got, expected int32) {
	if expected != got {
		t.Fatalf("expected %d, got %d", expected, got)
	}
}

// opens truetype fonts from truetype testdata.
func openFontFileTT(filename string) *tt.Font {
	f, err := tttestdata.Files.ReadFile(filename)
	check(err)

	font, err := tt.Parse(bytes.NewReader(f))
	check(err)

	return font
}

// opens truetype fonts from harfbuzz testdata.
func openFontFile(filename string) *tt.Font {
	f, err := testdata.Files.ReadFile(filename)
	check(err)

	font, err := tt.Parse(bytes.NewReader(f))
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

	assert(t, BottomToTop.Reverse() == TopToBottom)
	assert(t, TopToBottom.Reverse() == BottomToTop)
	assert(t, LeftToRight.Reverse() == RightToLeft)
	assert(t, RightToLeft.Reverse() == LeftToRight)
}

func TestFlag(t *testing.T) {
	if (glyphFlagDefined & (glyphFlagDefined + 1)) != 0 {
		t.Error("assertion failed")
	}
}

func TestTypesLanguage(t *testing.T) {
	fa := language.NewLanguage("fa")
	faIR := language.NewLanguage("fa_IR")
	faIr := language.NewLanguage("fa-ir")
	en := language.NewLanguage("en")

	assert(t, fa != "")
	assert(t, faIR != "")
	assert(t, faIR == faIr)

	assert(t, en != "")
	assert(t, en != fa)

	/* Test recall */
	assert(t, en == language.NewLanguage("en"))
	assert(t, en == language.NewLanguage("eN"))
	assert(t, en == language.NewLanguage("En"))

	assert(t, language.NewLanguage("") == "")
	assert(t, language.NewLanguage("e") != "")
}

func TestParseVariations(t *testing.T) {
	datas := [...]struct {
		input    string
		expected tt.Variation
	}{
		{" frea=45.78", tt.Variation{Tag: tt.MustNewTag("frea"), Value: 45.78}},
		{"G45E=45", tt.Variation{Tag: tt.MustNewTag("G45E"), Value: 45}},
		{"fAAD 45.78", tt.Variation{Tag: tt.MustNewTag("fAAD"), Value: 45.78}},
		{"fr 45.78", tt.Variation{Tag: tt.MustNewTag("fr  "), Value: 45.78}},
		{"fr=45.78", tt.Variation{Tag: tt.MustNewTag("fr  "), Value: 45.78}},
		{"fr=-45.4", tt.Variation{Tag: tt.MustNewTag("fr  "), Value: -45.4}},
		{"'fr45'=-45.4", tt.Variation{Tag: tt.MustNewTag("fr45"), Value: -45.4}}, // with quotes
		{`"frZD"=-45.4`, tt.Variation{Tag: tt.MustNewTag("frZD"), Value: -45.4}}, // with quotes
	}
	for _, data := range datas {
		out, err := ParseVariation(data.input)
		if err != nil {
			t.Fatalf("error on %s: %s", data.input, err)
		}
		if out != data.expected {
			t.Fatalf("for %s, expected %v, got %v", data.input, data.expected, out)
		}
	}
}

func TestParseFeature(t *testing.T) {
	inputs := [...]string{
		"kern",
		"+kern",
		"-kern",
		"kern=0",
		"kern=1",
		"aalt=2",
		"kern[]",
		"kern[:]",
		"kern[5:]",
		"kern[:5]",
		"kern[3:5]",
		"kern[3]",
		"aalt[3:5]=2",
	}
	expecteds := [...]Feature{
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 0, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 0, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("aalt"), Value: 2, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 0, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 5, End: FeatureGlobalEnd},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 0, End: 5},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 3, End: 5},
		{Tag: tt.MustNewTag("kern"), Value: 1, Start: 3, End: 4},
		{Tag: tt.MustNewTag("aalt"), Value: 2, Start: 3, End: 5},
	}
	for i, input := range inputs {
		f, err := ParseFeature(input)
		if err != nil {
			t.Fatalf("unexpected error on input <%s> : %s", input, err)
		}
		if exp := expecteds[i]; f != exp {
			t.Fatalf("for <%s>, expected %v, got %v", input, exp, f)
		}
	}
}
