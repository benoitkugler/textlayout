package common

import "testing"

func TestDigest(t *testing.T) {
	const (
		setTypeSize = 2
		numBits     = 3 + 1 + 1
	)
	if shift0 >= setTypeSize*8 {
		t.Error()
	}
	if shift0+numBits > setTypeSize*8 {
		t.Error()
	}
	if shift1 >= setTypeSize*8 {
		t.Error()
	}
	if shift1+numBits > setTypeSize*8 {
		t.Error()
	}
	if shift2 >= setTypeSize*8 {
		t.Error()
	}
	if shift2+numBits > setTypeSize*8 {
		t.Error()
	}
}

func TestDigestHas(t *testing.T) {
	var d SetDigest
	for i := setType(10); i < 65_000; i += 7 {
		d.Add(i)
	}
	for i := setType(10); i < 65_000; i += 7 {
		if !d.MayHave(i) {
			t.Errorf("expected MayHave for %d", i)
		}
		if d.MayHave(i + 1) {
			t.Errorf("expected not have for %d", i+1)
		}
	}
}
