package fontconfig

import (
	"fmt"
	"testing"
)

// func TestBuildCache(t *testing.T) {
// 	setupCacheFile()
// }

func TestScan(t *testing.T) {
	c := NewConfig()
	fs, err := c.ScanFontDirectories(testFontDir)
	if err != nil {
		t.Fatalf("scaning dir %s: %s", testFontDir, err)
	}
	spacings := map[int32]int{}
	nbVar, nbInstances := 0, 0
	for _, font := range fs {
		sp, _ := font.GetInt(SPACING)
		spacings[sp]++
		if isVariable, _ := font.GetBool(VARIABLE); isVariable == True {
			fmt.Println(font.FaceID())
			nbVar++
		}
		if font.FaceID().Instance != 0 {
			fmt.Println(font.FaceID())
			nbInstances++
		}
	}
	fmt.Println(spacings, nbVar, nbInstances)
}

func BenchmarkScan(b *testing.B) {
	var (
		c    Config
		seen = make(strSet)
	)
	for i := 0; i < b.N; i++ {
		_, err := c.readDir(testFontDir, seen)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFindFont(t *testing.T) {
	fs := cachedFS()
	// fmt.Println(fs.List(BuildPattern(PatternElement{Object: FAMILY, Value: String("Courier")})))
	fmt.Println(fs.List(nil, FAMILY, FILE))
}

func TestSortLinux(t *testing.T) {
	fs := cachedFS()

	c := NewConfig()
	if err := c.LoadFromDir("confs"); err != nil {
		t.Fatal(err)
	}

	query := BuildPattern(PatternElement{Object: FAMILY, Value: String("serif")})

	c.Substitute(query, nil, MatchQuery)
	query.SubstituteDefault()

	sorted, _ := fs.Sort(query, true)
	file, _ := sorted[0].GetString(FILE)
	exp := "/usr/share/fonts/truetype/dejavu/DejaVuSerif.ttf"
	if file != exp {
		t.Fatalf("expected %s, got %s", exp, file)
	}
}

func TestScanItalic(t *testing.T) {
	c := NewConfig()
	fs, err := c.ScanFontFile("test/DejaVuSerif-Italic.ttf")
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 1 {
		t.Fatalf("unexpected length: %d", len(fs))
	}
	font := fs[0]

	if slant, _ := font.GetInt(SLANT); slant != SLANT_ITALIC {
		t.Fatalf("expected italic (%d), got %d", SLANT_ITALIC, slant)
	}
}

// func TestScanCourier(t *testing.T) {
// 	c := NewConfig()
// 	fs, err := c.ScanFontFile("/System/Library/fonts/Courier.dfont")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(len(fs))
// }
