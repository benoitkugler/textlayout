package bitmap

import (
	"fmt"
	"os"
	"testing"
)

func TestCmap(t *testing.T) {
	for _, file := range []string{
		"test/4x6.pcf",
		"test/8x16.pcf.gz",
		"test/charB18.pcf.gz",
		"test/courB18.pcf.gz",
		"test/hanglg16.pcf.gz", // korean encoding
		"test/helvB18.pcf.gz",
		"test/lubB18.pcf.gz",
		"test/ncenB18.pcf.gz",
		"test/orp-italic.pcf.gz",
		"test/timB18.pcf.gz",
		"test/timR24-ISO8859-1.pcf.gz",
		"test/timR24.pcf.gz",
	} {
		fi, err := os.Open(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(fi)
		if err != nil {
			t.Fatal(file, err)
		}
		fi.Close()

		_, enc := font.Cmap()
		fmt.Println(enc)

		iter := font.cmap.Iter()
		for iter.Next() {
			r, g1 := iter.Char()
			font.cmap.Lookup(r + 1)
			g2 := font.cmap.Lookup(r)
			if g2 != g1 {
				t.Fatalf("inconsitent cmap iterator: 0x%04x : %d != %d", r, g1, g2)
			}
		}

		font.cmap.Lookup(0xFFFF)

		font.computeBitmapSize()
	}
}
