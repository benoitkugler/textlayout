package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
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

func TestIndicCombineCategories(t *testing.T) {
	if got := indicCombineCategories("Pure_Killer", "Top"); got != 1543 {
		t.Fatalf("expected %d, got %d", 1543, got)
	}
}

func TestIndic(t *testing.T) {
	b, err := ioutil.ReadFile("UnicodeData.txt")
	check(err)
	err = parseUnicodeDatabase(b)
	check(err)

	b, err = ioutil.ReadFile("Blocks.txt")
	check(err)
	blocks, err := parseAnnexTables(b)
	check(err)
	b, err = ioutil.ReadFile("IndicSyllabicCategory.txt")
	check(err)
	indicS, err := parseAnnexTables(b)
	check(err)
	b, err = ioutil.ReadFile("IndicPositionalCategory.txt")
	check(err)
	indicP, err := parseAnnexTables(b)
	check(err)

	startsExp := []rune{0x0028, 0x00B0, 0x0900, 0x1000, 0x1780, 0x1CD0, 0x2008, 0x2070, 0xA8E0, 0xA9E0, 0xAA60}
	endsExp := []rune{0x003F + 1, 0x00D7 + 1, 0x0DF7 + 1, 0x109F + 1, 0x17EF + 1, 0x1CFF + 1, 0x2017 + 1, 0x2087 + 1, 0xA8FF + 1, 0xA9FF + 1, 0xAA7F + 1}
	starts, ends := generateIndicTable(indicS, indicP, blocks, io.Discard)

	if !reflect.DeepEqual(starts, startsExp) {
		t.Fatalf("wrong starts; expected %v, got %v", startsExp, starts)
	}
	if !reflect.DeepEqual(ends, endsExp) {
		t.Fatalf("wrong ends; expected %v, got %v", endsExp, ends)
	}
}

func TestScripts(t *testing.T) {
	b, err := ioutil.ReadFile("Scripts.txt")
	check(err)
	scriptsRanges, err := parseAnnexTablesAsRanges(b)
	check(err)

	b, err = ioutil.ReadFile("Scripts-iso15924.txt")
	check(err)
	scriptNames, err := parseScriptNames(b)
	check(err)

	fmt.Println(len(compactScriptLookupTable(scriptsRanges, scriptNames)))
}
