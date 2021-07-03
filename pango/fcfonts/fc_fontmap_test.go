package fcfonts

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
)

const fontConfigCacheFile = "testdata/cache.fc"

func TestCreateCache(t *testing.T) {
	fmt.Println("Scanning fonts with empty config...")
	out, err := scanAndCache(fontConfigCacheFile)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Cache file written (%d fonts)\n", len(out))
}

func newChachedFontMap() *FontMap {
	var c fontconfig.Config
	f, err := os.Open(fontConfigCacheFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fs, err := fontconfig.LoadFontset(f)
	if err != nil {
		log.Fatal(err)
	}

	return NewFontMap(&c, fs)
}

func TestInitFontMap(t *testing.T) {
	fm := newChachedFontMap()
	fmt.Println("Loaded fonts from cache", len(fm.database))
}

func TestMetrics(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	desc := pango.NewFontDescriptionFrom("Courier 11")

	fmt.Println(newChachedFontMap().LoadFontset(context, &desc, pango.DefaultLanguage()).(*Fontset).fonts)

	// fmt.Println("metrics for: ", desc.String())

	// metrics := context.GetMetrics(&desc, pango.DefaultLanguage())

	// fmt.Println("\tascent", metrics.Ascent)
	// fmt.Println("\tdescent", metrics.Descent)
	// fmt.Println("\theight", metrics.Height)
	// fmt.Println("\tchar width", metrics.ApproximateCharWidth)
	// fmt.Println("\tdigit width", metrics.ApproximateDigitWidth)
	// fmt.Println("\tunderline position", metrics.UnderlinePosition)
	// fmt.Println("\tunderline thickness", metrics.UnderlineThickness)
	// fmt.Println("\tstrikethrough position", metrics.StrikethroughPosition)
	// fmt.Println("\tstrikethrough thickness", metrics.StrikethroughThickness)
}
