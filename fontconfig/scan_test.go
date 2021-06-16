package fontconfig

import (
	"fmt"
	"testing"
)

func TestScan(t *testing.T) {
	c := NewConfig()
	fs, err := c.ScanFontDirectories(testFontDir)
	if err != nil {
		t.Fatalf("scaning dir %s: %s", testFontDir, err)
	}
	spacings := map[int32]int{}
	for _, font := range fs {
		sp, _ := font.GetInt(SPACING)
		spacings[sp]++
	}
	fmt.Println(spacings)
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
