package truetype

import (
	"fmt"
	"os"
	"testing"
)

func TestKerx(t *testing.T) {
	filename := "testdata/Bangla Sangam MN.ttc"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	fonts, err := Loader.Load(file)
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	for _, font := range fonts {
		font := font.(*Font)

		out, err := font.KerxTable()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(out)

	}
}
