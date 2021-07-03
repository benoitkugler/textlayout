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
	fmt.Println("Loaded fonts from cache:", len(fm.database))
}

func TestMetrics(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	// desc := pango.NewFontDescriptionFrom("Courier 11")
	desc := pango.NewFontDescriptionFrom("Cantarell 11")

	fmt.Println("metrics for: ", desc.String())

	metrics := context.GetMetrics(&desc, pango.DefaultLanguage())

	fmt.Println("\tascent", metrics.Ascent)
	fmt.Println("\tdescent", metrics.Descent)
	fmt.Println("\theight", metrics.Height)
	fmt.Println("\tchar width", metrics.ApproximateCharWidth)
	fmt.Println("\tdigit width", metrics.ApproximateDigitWidth)
	fmt.Println("\tunderline position", metrics.UnderlinePosition)
	fmt.Println("\tunderline thickness", metrics.UnderlineThickness)
	fmt.Println("\tstrikethrough position", metrics.StrikethroughPosition)
	fmt.Println("\tstrikethrough thickness", metrics.StrikethroughThickness)
}

func TestExtents(t *testing.T) {
	str := []rune("Composer")

	context := pango.NewContext(newChachedFontMap())

	context.SetFontDescription(pango.NewFontDescriptionFrom("Cantarell 11"))

	items := context.Itemize(str, 0, len(str), nil, nil)
	glyphs := pango.GlyphString{}
	item := items.Data
	glyphs.Shape(str, &item.Analysis)
	var ink, log pango.Rectangle
	glyphs.Extents(item.Analysis.Font, &ink, &log)

	if ink.Width < 0 {
		t.Fatalf("expected ink.Width >= 0, got %d", ink.Width)
	}
	if ink.Height < 0 {
		t.Fatalf("expected ink.Height >= 0, got %d", ink.Height)
	}
	if log.Width < 0 {
		t.Fatalf("expected log.Width >= 0, got %d", log.Width)
	}
	if log.Height < 0 {
		t.Fatalf("expected log.Height >= 0, got %d", log.Height)
	}
}

func TestEnumerate(t *testing.T) {
	fontmap := newChachedFontMap()
	context := pango.NewContext(fontmap)

	// TODO: fix if we support families
	//    pango_font_map_list_families (fontmap, &families, &n_families);
	//    g_assert_cmpint (n_families, >, 0);

	//    for (i = 0; i < n_families; i++)
	// 	 {
	// 	   family = pango_font_map_GetFamily (fontmap, pango_font_family_GetName (families[i]));
	// 	   g_assert_true (family == families[i]);
	// 	 }

	//    pango_font_family_ListFaces (families[0], &faces, &n_faces);
	//    g_assert_cmpint (n_faces, >, 0);
	//    for (i = 0; i < n_faces; i++)
	// 	 {
	// 	   face = pango_font_family_GetFace (families[0], pango_font_face_GetFaceName (faces[i]));
	// 	   g_assert_true (face == faces[i]);
	// 	 }

	var desc pango.FontDescription
	desc.SetFamily("Courier")
	desc.SetSize(10 * pango.Scale)

	font := pango.LoadFont(fontmap, context, &desc)
	if font == nil {
		t.Fatalf("no font found for %s", desc)
	}
}
