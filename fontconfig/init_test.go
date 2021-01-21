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
	fmt.Println(cf.fontDirs)

	ti := time.Now()
	err = cf.FcConfigBuildFonts()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(time.Since(ti))
}

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := readDir("/usr/share/fonts")
		if err != nil {
			b.Fatal(err)
		}
	}
}
