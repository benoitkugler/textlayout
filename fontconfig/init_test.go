package fontconfig

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	cf, err := initLoadOwnConfig()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cf.fontDirs)

	err = cf.FcConfigBuildFonts()
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkLoad(b *testing.B) {
	cf, err := initLoadOwnConfig()
	if err != nil {
		b.Fatal(err)
	}
	fmt.Println(cf.fontDirs)

	for i := 0; i < b.N; i++ {
		err = cf.FcConfigBuildFonts()
		if err != nil {
			b.Fatal(err)
		}
	}
}
