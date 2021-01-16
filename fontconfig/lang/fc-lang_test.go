package main

import (
	"fmt"
	"testing"
)

func TestRuneBytes(t *testing.T) {
	r := rune(0xffffff)
	fmt.Println(r >= 0xffffff)
	fmt.Println(byte(r), byte(r>>8), byte(r>>16))
}
