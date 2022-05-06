package type1

import (
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/type1"
)

func TestParse(t *testing.T) {
	f, err := testdata.Files.Open("Times-Bold.afm")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = ParseAFMFile(f)
	if err != nil {
		t.Fatal(err)
	}
}

func TestKernings(t *testing.T) {
	f, err := testdata.Files.Open("Times-Bold.afm")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = ParseAFMFile(f)
	if err != nil {
		t.Fatal(err)
	}
}
