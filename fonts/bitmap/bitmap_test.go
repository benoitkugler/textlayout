package bitmap

import (
	"bytes"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/bitmap"
	"github.com/benoitkugler/textlayout/fonts"
)

var files = []string{
	"4x6.pcf",
	"8x16.pcf.gz",
	"charB18.pcf.gz",
	"courB18.pcf.gz",
	"hanglg16.pcf.gz", // korean encoding
	"helvB18.pcf.gz",
	"lubB18.pcf.gz",
	"ncenB18.pcf.gz",
	"orp-italic.pcf.gz",
	"timB18.pcf.gz",
	"timR24-ISO8859-1.pcf.gz",
	"timR24.pcf.gz",
}

func TestCmap(t *testing.T) {
	for _, file := range files {
		fi, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(bytes.NewReader(fi))
		if err != nil {
			t.Fatal(file, err)
		}

		_, enc := font.Cmap()
		if enc != fonts.EncUnicode && enc != fonts.EncOther {
			t.Fatal()
		}

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
		fi, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(bytes.NewReader(fi))
		if err != nil {
			t.Fatal(file, err)
		}

		if got := font.computeBitmapSize(); got != expectedSizes[i] {
			t.Fatalf("font %s: expected size %v, got %v", file, expectedSizes[i], got)
		}
	}
}

func TestAdvances(t *testing.T) {
	for i, file := range files {
		fi, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		font, err := Parse(bytes.NewReader(fi))
		if err != nil {
			t.Fatal(file, err)
		}

		expHAdvs := aAdvances[i].hAdvances
		for _, r := range aAdvances[i].runes {
			gid, _ := font.NominalGlyph(r)

			hAdv := font.HorizontalAdvance(fonts.GID(gid))
			vAdv := font.VerticalAdvance(fonts.GID(gid))
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

func TestScanDescription(t *testing.T) {
	for _, file := range files {
		if file == "hanglg16.pcf.gz" {
			continue
		}
		fi, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal("can't read test file", err)
		}

		l, err := ScanFont(bytes.NewReader(fi))
		if err != nil {
			t.Fatal(err)
		}

		if len(l) != 1 {
			t.Fatalf("unexected length %d", len(l))
		}

		l[0].Family()

		_, err = l[0].LoadCmap()
		if err != nil {
			t.Fatal(file, err)
		}

	}
}
