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
	for i := 0; i < b.N; i++ {
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
	for i := 0; i < b.N; i++ {
		buf.AddRunes(text, 0, -1)
		buf.Props.Direction = direction
		buf.Props.Script = script
		buf.Shape(font, nil)
		buf.Clear()
	}
}
