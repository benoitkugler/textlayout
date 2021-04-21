package graphite

import (
	"fmt"
	"os"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

func TestTableSilf(t *testing.T) {
	filename := "testdata/Annapurnarc2.ttf"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	font, err := truetype.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	ta, err := font.GetRawTable(tagSilf)
	if err != nil {
		t.Fatal(err)
	}

	silf, err := parseTableSilf(ta)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(silf[0].passes[0].startStates)
}
