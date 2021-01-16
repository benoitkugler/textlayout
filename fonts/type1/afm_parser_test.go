package type1

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("test/Times-Bold.afm")
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
	f, err := os.Open("test/Times-Bold.afm")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = ParseAFMFile(f)
	if err != nil {
		t.Fatal(err)
	}

}
