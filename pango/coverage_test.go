package pango

import (
	"testing"
)

func TestCoverageBasic(t *testing.T) {

	coverage := pango_coverage_new()

	for i := rune(0); i < 100; i++ {
		assertEquals(t, coverage.get(i), PANGO_COVERAGE_NONE)
	}

	for i := rune(0); i < 100; i++ {
		coverage.set(i, PANGO_COVERAGE_EXACT)
	}

	for i := rune(0); i < 100; i++ {
		assertEquals(t, coverage.get(i), PANGO_COVERAGE_EXACT)
	}

	for i := rune(0); i < 100; i++ {
		coverage.set(i, PANGO_COVERAGE_NONE)
	}

	for i := rune(0); i < 100; i++ {
		assertEquals(t, coverage.get(i), PANGO_COVERAGE_NONE)
	}

}

func TestCoverageCopy(t *testing.T) {

	coverage := pango_coverage_new()

	for i := rune(0); i < 100; i++ {
		coverage.set(i, PANGO_COVERAGE_EXACT)
	}
	coverage2 := coverage.copy()

	for i := rune(0); i < 50; i++ {
		coverage.set(i, PANGO_COVERAGE_NONE)
	}
	for i := rune(0); i < 100; i++ {
		assertEquals(t, coverage2.get(i), PANGO_COVERAGE_EXACT)
	}
}
