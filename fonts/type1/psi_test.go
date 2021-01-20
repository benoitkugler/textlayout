package type1

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPsi(t *testing.T) {
	file := "test/CalligrapherRegular.pfb"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	font, err := ParsePFBFile(b)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(font.GetAdvance(51))
}
