package glyphsnames

import (
	"sort"
	"testing"
)

func TestTables(t *testing.T) {
	if !sort.SliceIsSorted(glyphToRuneTable[:], func(i, j int) bool { return glyphToRuneTable[i].name < glyphToRuneTable[j].name }) {
		t.Fatal("glyphToRuneTable is not sorted")
	}

	if glyphToRune("251d") != 0x251d {
		t.Fatal()
	}
	if glyphToRune("udieresisbelow") != 0x1e73 {
		t.Fatal()
	}
	if glyphToRune("udieresisacutexxx") != 0 {
		t.Fatal()
	}

	if !sort.SliceIsSorted(glyphToAliasTable[:], func(i, j int) bool { return glyphToAliasTable[i].name < glyphToAliasTable[j].name }) {
		t.Fatal("glyphToAliasTable is not sorted")
	}

	if glyphToAlias("zeronojoin") != "afii61664" {
		t.Fatal()
	}
}
