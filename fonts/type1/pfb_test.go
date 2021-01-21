package type1

import (
	"fmt"
	"os"
	"testing"

	tokenizer "github.com/benoitkugler/pstokenizer"
)

func TestOpen(t *testing.T) {
	for _, file := range []string{
		"test/CalligrapherRegular.pfb",
		"test/Z003-MediumItalic.t1",
	} {
		b, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		font, err := Parse(b)
		if err != nil {
			t.Fatal(err)
		}
		if len(font.charstrings) == 0 {
			t.Fatal("font", file, "with no charstrings")
		}
		b.Close()
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, file := range []string{
			"test/CalligrapherRegular.pfb",
			"test/Z003-MediumItalic.t1",
		} {
			fi, err := os.Open(file)
			if err != nil {
				b.Fatal(err)
			}
			_, err = Parse(fi)
			if err != nil {
				b.Fatal(err)
			}
			fi.Close()
		}
	}
}

func TestTokenize(t *testing.T) {
	file := "test/CalligrapherRegular.pfb"
	b, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	s1, s2, err := openPfb(b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(s1), len(s2))

	tks, err := tokenizer.Tokenize(s1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(tks))

	// the tokenizer can't handle binary segment
	s2 = decryptSegment(s2)
	tks, err = tokenizer.Tokenize(s2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(tks))
}
