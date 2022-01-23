package fontconfig

import (
	"strings"
)

func cmpIgnoreCase(s1, s2 string) int {
	return strings.Compare(strings.ToLower(s1), strings.ToLower(s2))
}

func cmpIgnoreBlanksAndCase(s1, s2 string) int {
	return strings.Compare(ignoreBlanksAndCase(s1), ignoreBlanksAndCase(s2))
}

// Returns the location of `substr` in  `s`, ignoring case.
// Returns -1 if `substr` is not present in `s`.
func indexIgnoreCase(s, substr string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(substr))
}

// The bulk of the time in FcFontMatch and Sort goes to
// walking long lists of family names. We speed this up with a
// hash table.
type familyEntry struct {
	strongValue float32
	weakValue   float32
}

// map with strings key, ignoring blank and case
type blankCaseMap map[string]*familyEntry

var rp = strings.NewReplacer(" ", "", "\t", "")

func ignoreBlanksAndCase(s1 string) string { return rp.Replace(strings.ToLower(s1)) }

func (h blankCaseMap) lookup(s string) (*familyEntry, bool) {
	s = ignoreBlanksAndCase(s)
	e, ok := h[s]
	return e, ok
}

func (h blankCaseMap) add(s string, v *familyEntry) {
	s = ignoreBlanksAndCase(s)
	h[s] = v
}

// IgnoreBlanksAndCase
type familyBlankMap map[string]int

func (h familyBlankMap) lookup(s String) (int, bool) {
	ss := ignoreBlanksAndCase(string(s))
	e, ok := h[ss]
	return e, ok
}

func (h familyBlankMap) add(s String, v int) {
	ss := ignoreBlanksAndCase(string(s))
	h[ss] = v
}

func (h familyBlankMap) del(s String) {
	ss := ignoreBlanksAndCase(string(s))
	delete(h, ss)
}

// IgnoreCase
type familyMap map[string]int

func (h familyMap) lookup(s String) (int, bool) {
	ss := strings.ToLower(string(s))
	e, ok := h[ss]
	return e, ok
}

func (h familyMap) add(s String, v int) {
	ss := strings.ToLower(string(s))
	h[ss] = v
}

func (h familyMap) del(s String) {
	ss := strings.ToLower(string(s))
	delete(h, ss)
}
