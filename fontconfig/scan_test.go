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
