package truetype

import (
	"fmt"
	"os"
	"testing"
)

func TestGlyf(t *testing.T) {
	for _, filename := range []string{
		"testdata/Roboto-BoldItalic.ttf",
		"testdata/open-sans-v15-latin-regular.woff",
	} {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}
		head, err := font.HeadTable()
		if err != nil {
			t.Fatal(err)
		}

		gs, err := font.glyfTable(head)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(len(gs))
	}
}
