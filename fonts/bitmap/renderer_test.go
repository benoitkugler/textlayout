package bitmap

import (
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestGlyphData(t *testing.T) {
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

		expWidths := bitmapDims[i].widths
		expHeights := bitmapDims[i].heights
		for j, r := range bitmapDims[i].runes {
			gid, _ := font.NominalGlyph(r)

			data := font.GlyphData(gid, 10, 10).(fonts.GlyphBitmap)

			if data.Width != expWidths[j] {
				t.Fatalf("bitmap width font %s, glyph %d (rune %d), expected %d, got %d", file, gid, r, expWidths[j], data.Width)
			}
			if data.Height != expHeights[j] {
				t.Fatalf("bitmap height font %s, glyph %d (rune %d), expected %d, got %d", file, gid, r, expHeights[j], data.Height)
			}
		}
	}
}
