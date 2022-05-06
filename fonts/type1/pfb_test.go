package type1

import (
	"bytes"
	"fmt"
	"testing"

	tokenizer "github.com/benoitkugler/pstokenizer"
	testdata "github.com/benoitkugler/textlayout-testdata/type1"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestOpen(t *testing.T) {
	for _, filename := range []string{
		"c0419bt_.pfb",
		"CalligrapherRegular.pfb",
		"Z003-MediumItalic.t1",
	} {
		b, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		font, err := Parse(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if len(font.charstrings) == 0 {
			t.Fatal("font", filename, "with no charstrings")
		}

		if font.Encoding == nil {
			t.Fatal("expected encoding")
		}

		if font.GlyphName(10) == "" {
			t.Fatal("expected glyph names")
		}

		font.LoadSummary()
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, filename := range []string{
			"CalligrapherRegular.pfb",
			"Z003-MediumItalic.t1",
		} {
			by, err := testdata.Files.ReadFile(filename)
			if err != nil {
				b.Fatal(err)
			}

			_, err = Parse(bytes.NewReader(by))
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func TestTokenize(t *testing.T) {
	filename := "CalligrapherRegular.pfb"
	b, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	s1, s2, err := openPfb(bytes.NewReader(b))
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
	for _, filename := range []string{
		"c0419bt_.pfb",
		"CalligrapherRegular.pfb",
		"Z003-MediumItalic.t1",
	} {
		b, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}

		if font.Upem() != 1000 {
			t.Fatalf("expected upem 1000, got %d", font.Upem())
		}

		if p, ok := font.LineMetric(fonts.UnderlinePosition); !ok || p > 0 {
			t.Fatalf("unexpected underline position %f", p)
		}

		ext, ok := font.FontHExtents()
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

		if adv := font.HorizontalAdvance(gid); adv == 0 {
			t.Fatal("missing horizontal advance")
		}
	}
}

func TestCharstrings(t *testing.T) {
	for _, filename := range []string{
		"c0419bt_.pfb",
		"CalligrapherRegular.pfb",
		"Z003-MediumItalic.t1",
	} {
		b, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}

		for gid := range font.charstrings {
			if gid != 854 {
				continue
			}
			_, _, _, err := font.loadGlyph(fonts.GID(gid), false)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestScanDescription(t *testing.T) {
	for _, filename := range []string{
		"c0419bt_.pfb",
		"CalligrapherRegular.pfb",
		"Z003-MediumItalic.t1",
	} {
		b, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		l, err := ScanFont(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}

		if len(l) != 1 {
			t.Fatalf("unexected length %d", len(l))
		}

		fd := l[0]
		fmt.Println(fd.Family())
		fmt.Println(fd.Aspect())
		fmt.Println(fd.AdditionalStyle())

		cmap, err := fd.LoadCmap()
		if err != nil {
			t.Fatal(err)
		}
		if len(cmap.(fonts.CmapSimple)) == 0 {
			t.Fatal("invalid cmap")
		}
	}
}
