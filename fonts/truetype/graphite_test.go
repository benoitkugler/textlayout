package truetype

import (
	"fmt"
	"os"
	"testing"
)

func TestTableSilf(t *testing.T) {
	filename := "testdata/graphite/AwamiNastaliq-Regular.ttf"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	ta, err := font.findTableBuffer(font.tables[tagSilf])
	if err != nil {
		t.Fatal(err)
	}

	silf, err := parseTableSilf(ta)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(silf)
}
