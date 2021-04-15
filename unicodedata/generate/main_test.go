package main

import (
	"io/ioutil"
	"testing"
)

func TestVowel(t *testing.T) {
	b, err := ioutil.ReadFile("UnicodeData.txt")
	check(err)
	err = parseUnicodeDatabase(b)
	check(err)

	b, err = ioutil.ReadFile("Scripts.txt")
	check(err)
	scripts, err := parseAnnexTables(b)
	check(err)

	b, err = ioutil.ReadFile("ms-use/IndicShapingInvalidCluster.txt")
	check(err)
	vowelsConstraints := parseUSEInvalidCluster(b)

	// generate
	constraints, _ := aggregateVowelData(scripts, vowelsConstraints)

	if len(constraints["Devanagari"].dict[0x0905].dict) != 12 {
		t.Errorf("expected 12 constraints for rune 0x0905")
	}
}
