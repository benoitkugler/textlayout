package fontconfig

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
)

func randString() String {
	const chars = "azertyuiopqsmldfljgnxp:78945123&éçà)£µ%µ§"
	const L = 30
	out := make([]byte, L)
	for i := range out {
		out[i] = chars[rand.Intn(L)]
	}
	return String(out)
}

func randPatterns(N int) Fontset {
	var out Fontset
	for i := 0; i < N; i++ {

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
		r = Range{Begin: 0.45 * float32(i), End: 48.88}
		patt.Add(WEIGHT, r, true)
		patt.Add(MATRIX, Matrix{1, 2.45, 3, 4.}, true)
		patt.Add(ANTIALIAS, False, true)
		patt.Add(AUTOHINT, True, true)
		patt.Add(SCALABLE, DontCare, true)
		patt.Add(PIXEL_SIZE, Float(45.78), true)
		patt.Add(FOUNDRY, String("5456s4d"), true)
		patt.Add(ORDER, Int(7845*i), true)
		ls := langsetFrom([]string{"fr", "zh-hk", "zh-mo", "zh-sg"})
		patt.Add(LANG, ls, true)
		patt.Add(CHARSET, fcLangCharSets[i%len(fcLangCharSets)].charset, true)
		patt.Add(CHARSET, fcLangCharSets[2*i%len(fcLangCharSets)].charset, true)
		out = append(out, patt)
	}
	return out
}

func TestSerializeBin(t *testing.T) {
	fs := randPatterns(100)

	var buf bytes.Buffer
	err := fs.Serialize(&buf)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("cache file:", buf.Len()/1000, "KB")

	back, err := LoadFontset(&buf)
	if err != nil {
		t.Fatal(err)
	}

	for i := range back {
		if fs[i].Hash() != back[i].Hash() {
			t.Fatal("hash not preserved")
		}
	}
}

func TestSerializeDir(t *testing.T) {
	fs := cachedFS()
	var buf bytes.Buffer
	err := fs.Serialize(&buf)
	if err != nil {
		t.Fatal(err)
	}

	back, err := LoadFontset(&buf)
	if err != nil {
		t.Fatal(err)
	}

	for i := range back {
		if fs[i].Hash() != back[i].Hash() {
			t.Fatal("hash not preserved")
		}
	}
}

func TestCache(t *testing.T) {
	// on a real dataset
	fs := cachedFS()

	var b1 bytes.Buffer
	err := fs.SerializeGOB(&b1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GOB cache file:", b1.Len()/1000, "KB")

	ti := time.Now()
	_, err = LoadFontsetGOB(&b1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GOB cache loaded in:", time.Since(ti))

	var b2 bytes.Buffer
	err = fs.SerializeJSON(&b2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("JSON cache file:", b2.Len()/1000, "KB")

	ti = time.Now()
	_, err = LoadFontsetJSON(&b2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("JSON cache loaded in:", time.Since(ti))

	var b3 bytes.Buffer
	err = fs.Serialize(&b3)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Binary cache file:", b3.Len()/1000, "KB")

	ti = time.Now()
	_, err = LoadFontset(&b3)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Binary cache loaded in:", time.Since(ti))
}

func BenchmarkLoadCache(b *testing.B) {
	f, err := os.Open("test/cache.fc")
	if err != nil {
		b.Fatal("opening cache file for tests", err)
	}
	defer f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(io.SeekStart, 0)
		LoadFontset(f)
	}
}
