package truetype

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
)

func loadFont(t *testing.T, filename string) *Font {
	t.Helper()

	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return font
}

func TestSmokeTest(t *testing.T) {
	for _, filename := range []string{
		"testdata/Roboto-BoldItalic.ttf",
		"testdata/Raleway-v4020-Regular.otf",
		"testdata/open-sans-v15-latin-regular.woff",
		"testdata/OldaniaADFStd-Bold.otf", // duplicate tables
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}
		for tag := range font.tables {
			_ = tag.String()
		}

		_, err = font.OS2Table()
		if err != nil {
			t.Fatal(err)
		}
		_, err = font.GPOSTable()
		if err != nil {
			t.Fatal(err)
		}
		_, err = font.GSUBTable()
		if err != nil {
			t.Fatal(err)
		}

		_, err = font.HheaTable()
		if err != nil {
			t.Fatal(err)
		}

		font.loadTables()

		fs, err := Loader.Load(file)
		if err != nil {
			t.Fatal(err)
		}
		if len(fs) != 1 {
			t.Error("expected one font")
		}
		file.Close()
	}
}

func TestParseCrashers(t *testing.T) {
	font, err := Parse(bytes.NewReader([]byte{}))
	if font != nil || err == nil {
		t.Fail()
	}

	for range [50]int{} {
		input := make([]byte, 20000)
		rand.Read(input)
		font, err = Parse(bytes.NewReader(input))
		if font != nil || err == nil {
			t.Error("expected error on random input")
		}
	}
}

func TestTables(t *testing.T) {
	f, err := os.Open("testdata/LateefGR-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	font, err := NewFontParser(f)
	if err != nil {
		t.Fatalf("Parse err = %q, want nil", err)
	}
	fmt.Println(font.tables)
}

func TestCollection(t *testing.T) {
	for _, filename := range []string{
		"testdata/NotoSansCJK-Bold.ttc",
		"testdata/NotoSerifCJK-Regular.ttc",
		"testdata/Courier.dfont",
		"testdata/Geneva.dfont",
		"testdata/DFONT.dfont",
	} {
		f, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}
		fonts, err := Loader.Load(f)
		if err != nil {
			t.Fatal(filename, err)
		}
		for _, font := range fonts {
			_, err := font.LoadSummary()
			if err != nil {
				t.Fatal(err)
			}
		}
		f.Close()
	}
}

func TestCFF(t *testing.T) {
	files := []string{
		"testdata/AccanthisADFStdNo2-Regular.otf",
		"testdata/STIX-BoldItalic.otf",
		"testdata/NotoSansCJK-Bold.ttc",
	}
	for _, filename := range files {
		f, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}
		fonts, err := NewFontParsers(f)
		if err != nil {
			t.Fatal(filename, err)
		}

		for _, font := range fonts {
			ng, err := font.NumGlyphs()
			if err != nil {
				t.Fatal(err)
			}

			_, err = font.cffTable(ng)
			if err != nil {
				t.Fatal(filename, err)
			}
		}
		f.Close()
	}
}

func TestMetrics(t *testing.T) {
	font := loadFont(t, "testdata/DejaVuSerif.ttf")

	fmt.Println(font.GlyphName(74))
	fmt.Println(font.HorizontalAdvance(74))
	ext, _ := font.GlyphExtents(74, 0, 0)
	fmt.Println(ext.Width, ext.XBearing)
}

func TestScanDescription(t *testing.T) {
	for _, filename := range []string{
		"testdata/Roboto-BoldItalic.ttf",
		"testdata/Raleway-v4020-Regular.otf",
		"testdata/open-sans-v15-latin-regular.woff",
		"testdata/OldaniaADFStd-Bold.otf", // duplicate tables
		"testdata/NotoSansCJK-Bold.ttc",
		"testdata/NotoSerifCJK-Regular.ttc",
		"testdata/Courier.dfont",
		"testdata/Geneva.dfont",
		"testdata/DFONT.dfont",
	} {
		f, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ScanFont(f)
		if err != nil {
			t.Fatal(err)
		}

		f.Close()
	}
}
