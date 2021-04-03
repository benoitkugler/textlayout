package truetype

import (
	"fmt"
	"testing"
)

// func TestParseVorg(t *testing.T) {
// file := "testdata/SourceSansVariable-Roman.modcomp.ttf"
// f, err := os.Open(file)
// if err != nil {
// 	t.Fatal(err)
// }
// defer f.Close()

// font, err := Parse(f)
// if err != nil {
// 	t.Fatal(err)
// }

// v, err := font.vorgTable()
// if err != nil {
// 	t.Fatal(err)
// }
// fmt.Println(v)
// }

func TestScan(t *testing.T) {
	var v1, v2 float32
	fmt.Println(fmt.Sscanf("4. aaa", "%f %f", &v1, &v2))
}
