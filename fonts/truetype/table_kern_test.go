package truetype

import (
	"bytes"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestKern(t *testing.T) {
	for _, file := range []string{
		"Castoro-Regular.ttf",
		"Castoro-Italic.ttf",
		"FreeSerif.ttf",
	} {
		f, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		font, err := NewFontParser(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		var kern SimpleKerns
		if _, has := font.tables[tagKern]; has {
			_, err = font.KernTable(ng)
			if err != nil {
				t.Fatal(err)
			}
			// fmt.Println("	kerns (kern):", kern.Size())
		}

		if _, has := font.tables[TagGpos]; has {
			var gpos TableGPOS
			gpos, err = font.GPOSTable()
			if err != nil {
				t.Fatal(err)
			}
			kern, err = gpos.horizontalKerning()
			if err != nil {
				t.Fatal(err)
			}
			// fmt.Println("	kerns (GPOS):", kern.Size())
		}

		widths, err := font.HtmxTable(ng)
		if err != nil {
			t.Fatal(err)
		}
		for gid := range widths {
			a, b := GID(gid), GID(gid+1)
			_ = kern.KernPair(a, b)
		}
	}
}

func TestKernAAT(t *testing.T) {
	f, err := testdata.Files.ReadFile("ToyKern1.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := NewFontParser(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}

	ng, err := font.NumGlyphs()
	if err != nil {
		t.Fatal(err)
	}

	kerns, err := font.KernTable(ng)
	if err != nil {
		t.Fatal(err)
	}

	for _, k := range kerns {
		if simple, ok := k.Data.(SimpleKerns); ok {
			for i := GID(0); i < GID(ng); i++ {
				for j := GID(0); j < GID(ng); j++ {
					simple.KernPair(i, j)
				}
			}
		}
	}

	expectedSubtable1 := map[[2]GID]int16{
		{67, 68}: 0,
		{68, 69}: 0,
		{69, 70}: -30,
		{70, 71}: 0,
		{71, 72}: 0,
		{72, 73}: -20,
		{73, 74}: 0,
		{74, 75}: 0,
		{75, 76}: 0,
		{76, 77}: 0,
		{77, 78}: 0,
		{78, 79}: 0,
		{79, 80}: 0,
		{80, 81}: 0,
		{81, 82}: 0,
		{36, 57}: 0,
	}
	for k, exp := range expectedSubtable1 {
		got := kerns[1].Data.(Kern2).KernPair(k[0], k[1])
		if exp != got {
			t.Fatalf("invalid kern subtable : for (%d, %d) expected %d, got %d", k[0], k[1], exp, got)
		}
	}
	expectedSubtable2 := map[[2]GID]int16{
		{36, 57}: -80,
	}
	for k, exp := range expectedSubtable2 {
		got := kerns[2].Data.(Kern2).KernPair(k[0], k[1])
		if exp != got {
			t.Fatalf("invalid kern subtable : for (%d, %d) expected %d, got %d", k[0], k[1], exp, got)
		}
	}
}
