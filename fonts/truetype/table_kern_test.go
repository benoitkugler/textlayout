package truetype

import (
	"fmt"
	"os"
	"testing"
)

func TestKern(t *testing.T) {
	for _, file := range []string{
		"testdata/Castoro-Regular.ttf",
		"testdata/Castoro-Italic.ttf",
		"testdata/FreeSerif.ttf",
	} {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		font, err := Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		kern, err := font.KernTable(true)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("	kerns (prio kern):", kern.Size())

		kern, err = font.KernTable(false)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("	kerns (prio GPOS):", kern.Size())

		widths, err := font.HtmxTable()
		if err != nil {
			t.Fatal(err)
		}
		for gid := range widths {
			a, b := GlyphIndex(gid), GlyphIndex(gid+1)
			_, _ = kern.KernPair(a, b)
		}

		f.Close()
	}
}
