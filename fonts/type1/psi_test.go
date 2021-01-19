package type1

import (
	"io/ioutil"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/psinterpreter"
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
	for _, chars := range font.charstrings {
		var (
			psi     psinterpreter.Inter
			handler type1Operators
		)
		err := psi.Run(chars, &handler)
		if err != nil {
			t.Fatal(err)
		}
	}
}
