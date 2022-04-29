package graphite

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/graphite"
)

type slotInfo struct {
	gid     GID
	origin  Position
	unicode rune
}

type fontContext struct {
	face *GraphiteFace
	// font     *FontOptions
	features FeaturesValue
	rtl      int8
}

func newFontContext(t *testing.T, fontfile string, isRTL bool) fontContext {
	var out fontContext
	out.face = loadGraphite(t, fontfile)
	out.features = out.face.FeaturesForLang(0)
	out.rtl = int8(boolToInt(isRTL))
	return out
}

func (in fontContext) shape(text []rune) []output {
	seg := in.face.Shape(nil, text, 0, in.features, in.rtl)
	var out []output
	for s := seg.First; s != nil; s = s.Next {
		sl := slotInfo{gid: s.GID(), origin: s.Position, unicode: seg.getCharInfo(s.original).char}
		out = append(out, in.formatSlotInfo(sl))
	}
	out = append(out, output{label: "_adv_", x: roundFloat(seg.Advance.X), y: roundFloat(seg.Advance.Y)})
	return out
}

func (in fontContext) formatSlotInfo(s slotInfo) (out output) {
	if s.gid == 0 {
		out.label = fmt.Sprintf("0:%04X", s.unicode)
	} else {
		out.label = in.face.GlyphName(s.gid)
	}
	out.x = roundFloat(s.origin.X)
	out.y = roundFloat(s.origin.Y)
	return out
}

func roundFloat(f float32) float32 {
	const rounding = 0.1
	return float32(int(f/rounding)) * rounding
}

type output struct {
	label string
	x, y  float32
}

func (o *output) UnmarshalJSON(v []byte) error {
	tmp := [3]interface{}{&o.label, &o.x, &o.y}
	return json.Unmarshal(v, &tmp)
}

func makelabel(line, word int) string {
	return fmt.Sprintf("%d.%d", line, word)
}

func equalsEpsilon(v1, v2, epsilon float32) bool {
	return math.Abs(float64(v1-v2)) <= float64(epsilon)
}

func equalsOutput(o1, o2 output, epsilon float32) bool {
	return o1.label == o2.label &&
		equalsEpsilon(o1.x, o2.x, epsilon) && equalsEpsilon(o1.y, o2.y, epsilon)
}

type textRenderTest struct {
	name     string
	fontFile string
	textFile string
	isRTL    bool
	epsilon  float32
}

func (tt textRenderTest) process(t *testing.T) {
	textFile := filepath.Join("texts", "inputs", tt.textFile)
	b, err := testdata.Files.ReadFile(textFile)
	if err != nil {
		t.Fatal(err)
	}

	expectedJSONFile := filepath.Join("texts", "expecteds", tt.name+".json")
	expB, err := testdata.Files.ReadFile(expectedJSONFile)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string][][]output{}
	if err := json.Unmarshal(expB, &expected); err != nil {
		t.Fatal(err)
	}

	font := newFontContext(t, tt.fontFile, tt.isRTL)

	for lineCount, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// words := strings.Split(line, " ")
		words := []string{line}
		for wordCount, word := range words {
			label := makelabel(lineCount+1, wordCount+1)
			if len(expected[label]) == 0 {
				t.Fatal("missing entry for label", label, "in", tt.name)
			}
			exp := expected[label][0]
			got := font.shape([]rune(word))
			if len(exp) != len(got) {
				t.Fatalf("test %s (line %d): expected \n%v\n, got \n%v\n", tt.name, lineCount+1, exp, got)
			}
			for i := range exp {
				exp, got := exp[i], got[i]
				if !equalsOutput(exp, got, tt.epsilon) {
					fmt.Println(runesToText([]rune(word)))
					t.Fatalf("test %s (line %d): expected %v, got %v", tt.name, lineCount+1, exp, got)
				}
			}
		}
	}
}

var textRenderTests = []textRenderTest{
	{name: "padaukcmp1", fontFile: "Padauk.ttf", textFile: "my_HeadwordSyllables.txt"},
	{name: "chariscmp1", fontFile: "charis_r_gr.ttf", textFile: "udhr_eng.txt"},
	{name: "chariscmp2", fontFile: "charis_r_gr.ttf", textFile: "udhr_yor.txt"},
	{name: "annacmp1", fontFile: "Annapurnarc2.ttf", textFile: "udhr_nep.txt"},
	{name: "schercmp1", fontFile: "Scheherazadegr.ttf", textFile: "udhr_arb.txt", isRTL: true},
	{name: "awamicmp1", fontFile: "AwamiNastaliq-Regular.ttf", textFile: "awami_tests.txt", isRTL: true, epsilon: 1},
	{name: "awamicmp2", fontFile: "Awami_compressed_test.ttf", textFile: "awami_tests.txt", isRTL: true, epsilon: 1},
}

func TestFontTextRender(t *testing.T) {
	for _, te := range textRenderTests {
		te.process(t)
	}
}

func TestAwamiNoScale(t *testing.T) {
	input := shapingInput{name: "awami_no_scale", fontfile: "AwamiNastaliq-Regular.ttf", text: []rune{0x0759, 0x062a, 0x0759, 0x062c, 0x0634, 0x0759}, features: "", rtl: true}

	expected, err := testdata.Files.ReadFile("shape_refs/" + input.name + ".log")
	if err != nil {
		t.Fatal(err)
	}

	err = input.testWithScale(t, expected, false)
	if err != nil {
		t.Fatal(err)
	}
}
