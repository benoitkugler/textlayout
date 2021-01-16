package fontconfig

import (
	"strings"
	"unicode"
)

func FcStrCmpIgnoreCase(s1, s2 string) int {
	return strings.Compare(strings.ToLower(s1), strings.ToLower(s2))
}

func FcStrCmpIgnoreBlanksAndCase(s1, s2 string) int {
	return strings.Compare(ignoreBlanksAndCase(s1), ignoreBlanksAndCase(s2))
}

// Returns the location of `substr` in  `s`, ignoring case.
// Returns -1 if `substr` is not present in `s`.
func FcStrStrIgnoreCase(s, substr string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(substr))
}

// The bulk of the time in FcFontMatch and FcFontSort goes to
// walking long lists of family names. We speed this up with a
// hash table.
type familyEntry struct {
	strongValue float64
	weakValue   float64
}

// map with strings key, ignoring blank and case
type blankCaseMap map[string]*familyEntry

func ignoreBlanksAndCase(s string) string {
	s = strings.ToLower(s)
	return strings.TrimFunc(s, unicode.IsSpace)
}

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
