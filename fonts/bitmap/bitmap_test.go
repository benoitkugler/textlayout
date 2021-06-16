package bitmap

import (
	"fmt"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

var files = []string{
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
}

func TestCmap(t *testing.T) {
	for _, file := range files {
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
			g2, _ := font.cmap.Lookup(r)
			if g2 != g1 {
				t.Fatalf("inconsitent cmap iterator: 0x%04x : %d != %d", r, g1, g2)
			}
		}

		font.cmap.Lookup(0xFFFF)
	}
}

func TestSize(t *testing.T) {
	for i, file := range files {
		fi, err := os.Open(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(fi)
		if err != nil {
			t.Fatal(file, err)
		}
		fi.Close()

		if got := font.computeBitmapSize(); got != expectedSizes[i] {
			t.Fatalf("font %s: expected size %v, got %v", file, expectedSizes[i], got)
		}
	}
}

func TestAdvances(t *testing.T) {
	for i, file := range files {
		fi, err := os.Open(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(fi)
		if err != nil {
			t.Fatal(file, err)
		}
		fi.Close()

		expHAdvs := aAdvances[i].hAdvances
		for _, r := range aAdvances[i].runes {
			gid, _ := font.NominalGlyph(r)

			hAdv := font.HorizontalAdvance(fonts.GID(gid), nil)
			vAdv := font.VerticalAdvance(fonts.GID(gid), nil)
			expHAdv := expHAdvs[0]
			if len(expHAdvs) != 1 {
				expHAdv = expHAdvs[gid]
			}

			expVAdv := aAdvances[i].vAdvance
			if hAdv != expHAdv {
				t.Fatalf("horizontal advance font %s, glyph %d, expected %g, got %g", file, gid, expHAdv, hAdv)
			}
			if vAdv != expVAdv {
				t.Fatalf("vertical advance font %s, glyph %d, expected %g, got %g", file, gid, expVAdv, vAdv)
			}
		}
	}
}
