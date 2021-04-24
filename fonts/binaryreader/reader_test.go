package binaryreader

import (
	"math/rand"
	"testing"
)

func TestCrash(t *testing.T) {
	var input [1000]byte
	rand.Read(input[:])
	r := NewReader(input[:])
	r.Byte()
	r.Uint16()
	r.Uint32()
	r.Uint16s(10)
	r.Uint32s(20)
	r.ReadStruct(make([]int, 5))
	r.Uint16s(1000)
	r.Uint32s(1000)
}
