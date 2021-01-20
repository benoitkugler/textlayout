package fontconfig

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"testing"
)

func randString() String {
	out := make([]byte, 100)
	rand.Read(out)
	return String(out)
}

func TestSerialize(t *testing.T) {
	var out FcFontSet
	for i := range [100]int{} {

		patt := NewPattern()
		patt.Add(FC_FAMILY, randString(), true)
		patt.Add(FC_FAMILY, randString(), true)
		patt.Add(FC_SIZE, Int(10*i), true)
		patt.Add(FC_FAMILY, String("Foo"+string(rune(i))), true)
		patt.Add(FC_FAMILY, String("Bar"+string(rune(i))), true)
		patt.Add(FC_SIZE, Int(10*i), true)
		patt.Add(FC_FAMILY, randString(), true)
		patt.Add(FC_WEIGHT, Int(i*WEIGHT_MEDIUM), true)
		patt.Add(FC_FAMILY, randString(), true)
		patt.Add(FC_WEIGHT, Int(i*WEIGHT_MEDIUM), true)
		patt.Add(FC_WEIGHT, Int(i*WEIGHT_MEDIUM), true)
		patt.Add(FC_WIDTH, Int(i*WIDTH_NORMAL), true)
		patt.Add(FC_WIDTH, Int(i*WIDTH_NORMAL), true)
		r := Range{Begin: WEIGHT_MEDIUM, End: WEIGHT_BOLD}
		patt.Add(FC_WEIGHT, r, true)
		r = Range{Begin: 0.45 * float64(i), End: 48.88}
		patt.Add(FC_WEIGHT, r, true)
		patt.Add(FC_MATRIX, Matrix{1, 2.45, 3, 4.}, true)
		patt.Add(FC_ANTIALIAS, FcFalse, true)
		patt.Add(FC_AUTOHINT, FcTrue, true)
		patt.Add(FC_SCALABLE, FcDontCare, true)
		patt.Add(FC_PIXEL_SIZE, Float(45.78), true)
		patt.Add(FC_FOUNDRY, String("5456s4d"), true)
		patt.Add(FC_ORDER, Int(7845*i), true)
		ls := langsetFrom([]string{"fr", "zh-hk", "zh-mo", "zh-sg", "custom"})
		patt.Add(FC_LANG, ls, true)
		patt.Add(FC_CHARSET, fcLangCharSets[i].charset, true)
		patt.Add(FC_CHARSET, fcLangCharSets[2*i].charset, true)

		out = append(out, patt)
	}

	var (
		by, by2 bytes.Buffer
		back    FcFontSet
	)
	err := gob.NewEncoder(&by).Encode(out)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("plain:", by.Len())

	wr := gzip.NewWriter(&by2)
	_, err = io.Copy(wr, bytes.NewReader(by.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if err := wr.Close(); err != nil {
		t.Fatal(err)
	}

	fmt.Println("compressed gzip:", by2.Len())

	err = gob.NewDecoder(&by).Decode(&back)
	if err != nil {
		t.Fatal(err)
	}

	for i := range back {
		if out[i].Hash() != back[i].Hash() {
			t.Fatal("hash not preserved")
		}
	}
}
