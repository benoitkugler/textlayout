package type1

import (
	"fmt"
	"os"
	"testing"

	tokenizer "github.com/benoitkugler/pstokenizer"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestOpen(t *testing.T) {
	for _, file := range []string{
		"test/c0419bt_.pfb",
		"test/CalligrapherRegular.pfb",
		"test/Z003-MediumItalic.t1",
	} {
		b, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(b)
		if err != nil {
			t.Fatal(err)
		}
		if len(font.charstrings) == 0 {
			t.Fatal("font", file, "with no charstrings")
		}

		if font.Encoding == nil {
			t.Fatal("expected encoding")
		}

		if font.GlyphName(10) == "" {
			t.Fatal("expected glyph names")
		}

		font.LoadSummary()

		b.Close()
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, file := range []string{
			"test/CalligrapherRegular.pfb",
			"test/Z003-MediumItalic.t1",
		} {
			fi, err := os.Open(file)
			if err != nil {
				b.Fatal(err)
			}
			_, err = Parse(fi)
			if err != nil {
				b.Fatal(err)
			}
			fi.Close()
		}
	}
}

func TestTokenize(t *testing.T) {
	file := "test/CalligrapherRegular.pfb"
	b, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	s1, s2, err := openPfb(b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(s1), len(s2))

	tks, err := tokenizer.Tokenize(s1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(tks))

	// the tokenizer can't handle binary segment
	s2 = decryptSegment(s2)
	tks, err = tokenizer.Tokenize(s2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(tks))
}

func TestMetrics(t *testing.T) {
	for _, file := range []string{
		"test/c0419bt_.pfb",
		"test/CalligrapherRegular.pfb",
		"test/Z003-MediumItalic.t1",
	} {
		b, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(b)
		if err != nil {
			t.Fatal(err)
		}

		if font.Upem() != 1000 {
			t.Fatalf("expected upem 1000, got %d", font.Upem())
		}

		if p, ok := font.LineMetric(fonts.UnderlinePosition, nil); !ok || p > 0 {
			t.Fatalf("unexpected underline position %f", p)
		}

		ext, ok := font.FontHExtents(nil)
		if !ok {
			t.Fatalf("missing font horizontal extents")
		}
		if ext.Ascender < 0 || ext.Descender > 0 {
			t.Fatalf("unexpected sign for ascender and descender: %v", ext)
		}

		gid, ok := font.NominalGlyph(0x20)
		if !ok {
			t.Fatalf("missing space")
		}

		if adv := font.HorizontalAdvance(gid, nil); adv == 0 {
			t.Fatal("missing horizontal advance")
		}

		b.Close()
	}
}

func TestCharstrings(t *testing.T) {
	for _, file := range []string{
		"test/c0419bt_.pfb",
		"test/CalligrapherRegular.pfb",
		"test/Z003-MediumItalic.t1",
	} {
		b, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(b)
		if err != nil {
			t.Fatal(err)
		}

		for gid := range font.charstrings {
			if gid != 854 {
				continue
			}
			_, _, err := font.parseGlyphMetrics(fonts.GID(gid), false)
			if err != nil {
				t.Fatal(err)
			}
		}

		b.Close()
	}
}
