package fontconfig

import (
	"fmt"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	cf, err := initLoadOwnConfig()
	if err != nil {
		t.Fatal(err)
	}
	// fmt.Println(cf.fontDirs)

	ti := time.Now()
	fs, err := cf.BuildFonts()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(time.Since(ti))
	fmt.Println("patterns collected: ", len(fs))
}

func BenchmarkLoad(b *testing.B) {
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
