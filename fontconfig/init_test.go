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
	fmt.Println(cf.configDirs)
	fmt.Println(cf.configFiles)
	fmt.Println(cf.cacheDirs)
	fmt.Println(cf.fontDirs)

}
