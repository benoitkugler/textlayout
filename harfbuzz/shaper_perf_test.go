package harfbuzz

import (
	"io/ioutil"
	"os"
	"testing"

	tt "github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/language"
)

// ported from harfbuzz/perf

func BenchmarkShaping(b *testing.B) {
	runs := []struct {
		name      string
		textFile  string
		direction Direction
		script    language.Script
		fontFile  string
	}{
		{
			"fa-thelittleprince.txt - Amiri",
			"perf/texts/fa-thelittleprince.txt",
			RightToLeft,
			language.Arabic,
			"perf/fonts/Amiri-Regular.ttf",
		},
		{
			"fa-thelittleprince.txt - NotoNastaliqUrdu",
			"perf/texts/fa-thelittleprince.txt",
			RightToLeft,
			language.Arabic,
			"perf/fonts/NotoNastaliqUrdu-Regular.ttf",
		},

		{
			"fa-monologue.txt - Amiri",
			"perf/texts/fa-monologue.txt",
			RightToLeft,
			language.Arabic,
			"perf/fonts/Amiri-Regular.ttf",
		},
		{
			"fa-monologue.txt - NotoNastaliqUrdu",
			"perf/texts/fa-monologue.txt",
			RightToLeft,
			language.Arabic,
			"perf/fonts/NotoNastaliqUrdu-Regular.ttf",
		},

		{
			"en-thelittleprince.txt - Roboto",
			"perf/texts/en-thelittleprince.txt",
			LeftToRight,
			language.Latin,
			"perf/fonts/Roboto-Regular.ttf",
		},

		{
			"en-words.txt - Roboto",
			"perf/texts/en-words.txt",
			LeftToRight,
			language.Latin,
			"perf/fonts/Roboto-Regular.ttf",
		},
	}
}

func shapeOne(b *testing.B, textFile, fontFile string, direction Direction, script language.Script) {
	f, err := os.Open(fontFile)
	check(err)
	defer f.Close()

	fonts, err := tt.Loader.Load(f)
	check(err)

	font := NewFont(fonts[0].LoadMetrics())

	textB, err := ioutil.ReadFile(textFile)
	check(err)
	text := []rune(string(textB))

	buf := NewBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.AddRunes(text, 0, -1)
		buf.Props.Direction = direction
		buf.Props.Script = script
		buf.Shape(font, nil)
		buf.Clear()
	}
}
