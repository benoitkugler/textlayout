package truetype

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func loadFont(t *testing.T, filename string) *Font {
	t.Helper()

	f, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}
	return font
}

func TestSmokeTest(t *testing.T) {
	for _, filename := range []string{
		"Roboto-BoldItalic.ttf",
		"Raleway-v4020-Regular.otf",
		"open-sans-v15-latin-regular.woff",
		"OldaniaADFStd-Bold.otf", // duplicate tables
	} {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
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

		fs, err := Load(bytes.NewReader(file))
		if err != nil {
			t.Fatal(err)
		}
		if len(fs) != 1 {
			t.Error("expected one font")
		}
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
	f, err := testdata.Files.ReadFile("LateefGR-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	font, err := NewFontParser(bytes.NewReader(f))
	if err != nil {
		t.Fatalf("Parse err = %q, want nil", err)
	}
	fmt.Println(font.tables)
}

func TestCollection(t *testing.T) {
	for _, filename := range []string{
		"NotoSansCJK-Bold.ttc",
		"NotoSerifCJK-Regular.ttc",
		"Courier.dfont",
		"Geneva.dfont",
		"DFONT.dfont",
	} {
		f, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		fonts, err := Load(bytes.NewReader(f))
		if err != nil {
			t.Fatal(filename, err)
		}
		for _, font := range fonts {
			_, err := font.LoadSummary()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestCFF(t *testing.T) {
	files := []string{
		"AccanthisADFStdNo2-Regular.otf",
		"STIX-BoldItalic.otf",
		"NotoSansCJK-Bold.ttc",
	}
	for _, filename := range files {
		f, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		fonts, err := NewFontParsers(bytes.NewReader(f))
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
	}
}

func TestMetrics(t *testing.T) {
	font := loadFont(t, "DejaVuSerif.ttf")

	fmt.Println(font.GlyphName(74))
	fmt.Println(font.HorizontalAdvance(74))
	ext, _ := font.GlyphExtents(74, 0, 0)
	fmt.Println(ext.Width, ext.XBearing)
}

func TestScanDescription(t *testing.T) {
	for _, filename := range []string{
		"Roboto-BoldItalic.ttf",
		"Raleway-v4020-Regular.otf",
		"open-sans-v15-latin-regular.woff",
		"OldaniaADFStd-Bold.otf", // duplicate tables
		"NotoSansCJK-Bold.ttc",
		"NotoSerifCJK-Regular.ttc",
		"Courier.dfont",
		"Geneva.dfont",
		"DFONT.dfont",
	} {
		f, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}

		fds, err := ScanFont(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}

		for _, fd := range fds {
			fd.Aspect()
			_, err := fd.LoadCmap()
			if err != nil {
				t.Fatal(err)
			}
			fd.Family()
			fd.AdditionalStyle()
		}
	}
}
