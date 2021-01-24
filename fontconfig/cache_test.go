package fontconfig

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func randString() String {
	out := make([]byte, 100)
	rand.Read(out)
	return String(out)
}

func TestSerialize(t *testing.T) {
	var out Fontset
	for i := range [100]int{} {

		patt := NewPattern()
		patt.Add(FAMILY, randString(), true)
		patt.Add(FAMILY, randString(), true)
		patt.Add(SIZE, Int(10*i), true)
		patt.Add(FAMILY, String("Foo"+string(rune(i))), true)
		patt.Add(FAMILY, String("Bar"+string(rune(i))), true)
		patt.Add(SIZE, Int(10*i), true)
		patt.Add(FAMILY, randString(), true)
		patt.Add(WEIGHT, Int(i*WEIGHT_MEDIUM), true)
		patt.Add(FAMILY, randString(), true)
		patt.Add(WEIGHT, Int(i*WEIGHT_MEDIUM), true)
		patt.Add(WEIGHT, Int(i*WEIGHT_MEDIUM), true)
		patt.Add(WIDTH, Int(i*WIDTH_NORMAL), true)
		patt.Add(WIDTH, Int(i*WIDTH_NORMAL), true)
		r := Range{Begin: WEIGHT_MEDIUM, End: WEIGHT_BOLD}
		patt.Add(WEIGHT, r, true)
		r = Range{Begin: 0.45 * float64(i), End: 48.88}
		patt.Add(WEIGHT, r, true)
		patt.Add(MATRIX, Matrix{1, 2.45, 3, 4.}, true)
		patt.Add(ANTIALIAS, False, true)
		patt.Add(AUTOHINT, True, true)
		patt.Add(SCALABLE, DontCare, true)
		patt.Add(PIXEL_SIZE, Float(45.78), true)
		patt.Add(FOUNDRY, String("5456s4d"), true)
		patt.Add(ORDER, Int(7845*i), true)
		ls := langsetFrom([]string{"fr", "zh-hk", "zh-mo", "zh-sg", "custom"})
		patt.Add(LANG, ls, true)
		patt.Add(CHARSET, fcLangCharSets[i].charset, true)
		patt.Add(CHARSET, fcLangCharSets[2*i].charset, true)

		out = append(out, patt)
	}

	var by bytes.Buffer
	err := out.Serialize(&by)
	if err != nil {
		t.Fatal(err)
	}

	back, err := LoadFontset(&by)
	if err != nil {
		t.Fatal(err)
	}

	for i := range back {
		if out[i].Hash() != back[i].Hash() {
			t.Fatal("hash not preserved")
		}
	}
}

func TestCache(t *testing.T) {
	// on a real dataset
	var c Config
	fs, err := c.ScanFontDirectories(testFontDir)
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	err = fs.Serialize(&b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("cache file:", b.Len()/1000, "KB")

	ti := time.Now()
	fs2, err := LoadFontset(&b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("cache loaded in:", time.Since(ti))

	if len(fs) != len(fs2) {
		t.Fatalf("expected same lengths, got %d and %d", len(fs), len(fs2))
	}
	for i, f := range fs {
		if f.Hash() != fs[i].Hash() {
			t.Fatalf("different fonts: %s and %s", f, fs[i])
		}
	}
}
