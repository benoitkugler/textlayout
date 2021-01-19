package type1

import (
	"fmt"
	"io/ioutil"
	"testing"

	tokenizer "github.com/benoitkugler/pstokenizer"
)

func TestOpen(t *testing.T) {
	file := "test/CalligrapherRegular.pfb"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	font, err := ParsePFBFile(b)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(font.subrs))
	fmt.Println(len(font.charstrings))
}

func TestTokenize(t *testing.T) {
	file := "test/CalligrapherRegular.pfb"
	b, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
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
	s2, err = decryptSegment(s2)
	if err != nil {
		t.Fatal(err)
	}
	tks, err = tokenizer.Tokenize(s2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(tks))
}
