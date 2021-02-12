package truetype

import (
	"fmt"
	"os"
	"testing"
)

func TestParseGdef(t *testing.T) {
	filename := "testdata/Commissioner-VF.ttf"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	gdef, err := font.GDefTable()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gdef.Class.GlyphSize())
}
