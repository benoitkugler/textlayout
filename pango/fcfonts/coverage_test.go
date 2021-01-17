package fcfonts

import (
	"testing"
)

func assert(t *testing.T, b bool) {
	if !b {
		t.Fatal("assertion error")
	}
}

func TestCoverageBasic(t *testing.T) {
	var cov coverage

	for i := rune(0); i < 100; i++ {
		assert(t, !cov.Get(i))
	}

	for i := rune(0); i < 100; i++ {
		cov.Set(i, true)
	}

	for i := rune(0); i < 100; i++ {
		assert(t, cov.Get(i))
	}

	for i := rune(0); i < 100; i++ {
		cov.Set(i, false)
	}

	for i := rune(0); i < 100; i++ {
		assert(t, !cov.Get(i))
	}
}

func TestCoverageCopy(t *testing.T) {
	var cov coverage

	for i := rune(0); i < 100; i++ {
		cov.Set(i, true)
	}
	cov2 := cov.Copy()

	for i := rune(0); i < 50; i++ {
		cov2.Set(i, false)
	}
	for i := rune(0); i < 50; i++ {
		assert(t, !cov2.Get(i))
	}
	for i := rune(51); i < 100; i++ {
		assert(t, cov2.Get(i))
	}
	for i := rune(0); i < 100; i++ {
		assert(t, cov.Get(i))
	}
}
