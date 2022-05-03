package truetype

import (
	"fmt"
	"testing"
)

// func TestParseVorg(t *testing.T) {
// file := "SourceSansVariable-Roman.modcomp.ttf"
// f, err := testdata.Files.ReadFile(file)
// if err != nil {
// 	t.Fatal(err)
// }

// font, err := Parse(bytes.NewReader(f))
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
