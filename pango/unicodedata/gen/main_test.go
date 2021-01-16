package main

import (
	"fmt"
	"testing"
	"unicode"
)

func TestU(t *testing.T) {
	for _, r := range []rune{'\t', '\n', '\r', '\f'} {
		fmt.Println(unicode.Is(unicode.Cc, r))
	}
}
