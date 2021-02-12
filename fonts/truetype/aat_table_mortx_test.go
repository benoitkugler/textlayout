package truetype

import (
	"os"
	"testing"
)

func TestMort(t *testing.T) {
	for _, filename := range dirFiles(t, "testdata/layout_fonts/morx") {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		_, err = font.MorxTable()
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Println(out)
	}
}
