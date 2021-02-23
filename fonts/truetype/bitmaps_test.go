package truetype

import (
	"fmt"
	"os"
	"testing"
)

func TestSbix(t *testing.T) {
	for _, filename := range []string{
		"testdata/ToyFeat.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		gs, err := font.sbixTable()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of strkes:", len(gs.strikes))

		file.Close()
	}
}

func TestCblc(t *testing.T) {
	for _, filename := range []string{
		"testdata/ToyCBLC1.ttf",
		"testdata/ToyCBLC2.ttf",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		gs, err := font.cblcTable()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Number of strikes:", len(gs))

		file.Close()
	}
}
