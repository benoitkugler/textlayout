package truetype

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

// using the iterator
func compileCmap(c Cmap) map[rune]GlyphIndex {
	out := make(map[rune]GlyphIndex)
	iter := c.Iter()
	for iter.Next() {
		r, g := iter.Char()
		out[r] = g
	}
	return out
}

func compileNativeCmap(c Cmap) map[rune]GlyphIndex {
	switch c := c.(type) {
	case cmap0:
		return c
	case cmap4:
		return c.Compile()
	case cmap6:
		return c.Compile()
	case cmap12:
		return c.Compile()
	default:
		panic("should not happen")
	}
}

func (s cmap12) Compile() map[rune]GlyphIndex {
	chars := map[rune]GlyphIndex{}
	for _, cm := range s {
		for c := cm.start; c <= cm.end; c++ {
			chars[rune(c)] = GlyphIndex(c - cm.start + cm.delta)
		}
	}
	return chars
}

func (s cmap6) Compile() map[rune]GlyphIndex {
	chars := make(map[rune]GlyphIndex, len(s.entries))
	for i, entry := range s.entries {
		chars[rune(i)+s.firstCode] = GlyphIndex(entry)
	}
	return chars
}

func (s cmap4) Compile() map[rune]GlyphIndex {
	out := make(map[rune]GlyphIndex, len(s))
	for _, entry := range s {
		if entry.indexes == nil {
			for c := entry.start; c <= entry.end; c++ {
				out[rune(c)] = GlyphIndex(c + entry.delta)
				if c == 65535 { // avoid overflow
					break
				}
			}
		} else {
			for i, gi := range entry.indexes {
				out[rune(i)+rune(entry.start)] = gi
			}
		}
	}
	return out
}

func testCm(t *testing.T, cmap Cmap) {
	all := compileCmap(cmap)
	fmt.Println("	cmap:", len(all))
	for r, c := range all {
		c2 := cmap.Lookup(r)
		if c2 != c {
			t.Errorf("inconsistent lookup for rune %d : got %d and %d", r, c, c2)
		}
	}

	all2 := compileNativeCmap(cmap)
	if !reflect.DeepEqual(all, all2) {
		t.Error("inconsistant compile functions")
	}
}

func TestCmap(t *testing.T) {
	for _, file := range []string{
		"testdata/Roboto-BoldItalic.ttf",
		"testdata/Raleway-v4020-Regular.otf",
		"testdata/Castoro-Regular.ttf",
		"testdata/Castoro-Italic.ttf",
		"testdata/FreeSerif.ttf",
		"testdata/AnjaliOldLipi-Regular.ttf",
	} {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(f)
		if err != nil {
			t.Fatal(err)
		}
		cmap, err := font.CmapTable()
		if err != nil {
			t.Fatal(err)
		}

		testCm(t, cmap)

		f.Close()
	}
}

func TestCmap6(t *testing.T) {
	cm := cmap6{
		firstCode: 45684,
		entries:   []uint16{1, 2, 5, 9, 489, 8231, 84},
	}
	testCm(t, cm)
}
