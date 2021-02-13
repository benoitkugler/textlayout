package truetype

import (
	"os"
	"testing"
)

func TestBits(t *testing.T) {
	var buf8 [8]int8
	uint16As2Bits(buf8[:], 0x123F)
	if exp := [8]int8{0, 1, 0, -2, 0, -1, -1, -1}; buf8 != exp {
		t.Fatalf("expected %v, got %v", exp, buf8)
	}

	var buf4 [4]int8
	uint16As4Bits(buf4[:], 0x123F)
	if exp := [4]int8{1, 2, 3, -1}; buf4 != exp {
		t.Fatalf("expected %v, got %v", exp, buf4)
	}

	var buf2 [2]int8
	uint16As8Bits(buf2[:], 0x123F)
	if exp := [2]int8{18, 63}; buf2 != exp {
		t.Fatalf("expected %v, got %v", exp, buf2)
	}
}

func TestGPOS(t *testing.T) {
	filenames := []string{
		"testdata/Raleway-v4020-Regular.otf",
		"testdata/Estedad-VF.ttf",
		"testdata/Mada-VF.ttf",
	}

	filenames = append(filenames, dirFiles(t, "testdata/layout_fonts/gpos")...)

	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		_, err = font.GposTable()
		if err != nil {
			t.Fatal(filename, err)
		}
		// for _, l := range sub.Lookups {
		// 	for _, s := range l.Subtables {
		// 		if s.Data == nil {
		// 			continue
		// 		}
		// 	}
		// }
		// fmt.Println(len(sub.Lookups), "lookups")
	}
}
