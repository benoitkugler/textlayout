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
		fontFile  string
		script    language.Script
		direction Direction
	}{
		{
			"fa-thelittleprince.txt - Amiri",
			"testdata/perf_reference/texts/fa-thelittleprince.txt",
			"testdata/perf_reference/fonts/Amiri-Regular.ttf",
			language.Arabic,
			RightToLeft,
		},
		{
			"fa-thelittleprince.txt - NotoNastaliqUrdu",
			"testdata/perf_reference/texts/fa-thelittleprince.txt",
			"testdata/perf_reference/fonts/NotoNastaliqUrdu-Regular.ttf",
			language.Arabic,
			RightToLeft,
		},

		{
			"fa-monologue.txt - Amiri",
			"testdata/perf_reference/texts/fa-monologue.txt",
			"testdata/perf_reference/fonts/Amiri-Regular.ttf",
			language.Arabic,
			RightToLeft,
		},
		{
			"fa-monologue.txt - NotoNastaliqUrdu",
			"testdata/perf_reference/texts/fa-monologue.txt",
			"testdata/perf_reference/fonts/NotoNastaliqUrdu-Regular.ttf",
			language.Arabic,
			RightToLeft,
		},

		{
			"en-thelittleprince.txt - Roboto",
			"testdata/perf_reference/texts/en-thelittleprince.txt",
			"testdata/perf_reference/fonts/Roboto-Regular.ttf",
			language.Latin,
			LeftToRight,
		},

		{
			"en-words.txt - Roboto",
			"testdata/perf_reference/texts/en-words.txt",
			"testdata/perf_reference/fonts/Roboto-Regular.ttf",
			language.Latin,
			LeftToRight,
		},
	}

	for _, run := range runs {
		b.Run(run.name, func(b *testing.B) {
			shapeOne(b, run.textFile, run.fontFile, run.direction, run.script)
		})
	}
}

func shapeOne(b *testing.B, textFile, fontFile string, direction Direction, script language.Script) {
	f, err := os.Open(fontFile)
	check(err)
	defer f.Close()

	fonts, err := tt.Loader.Load(f)
	check(err)

	font := NewFont(fonts[0])

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
