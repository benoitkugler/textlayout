// copied from https://git.maze.io/go/unipdf/src/branch/master/internal/textencoding
package glyphsnames

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// GlyphToRune returns the rune corresponding to glyph `glyph` if there is one.
func GlyphToRune(glyph string) (rune, bool) {
	// We treat glyph "eight.lf" the same as glyph "eight".
	if strings.Contains(glyph, ".") {
		groups := rePrefix.FindStringSubmatch(glyph)
		if groups != nil {
			glyph = groups[1]
		}
	}
	// First lookup the glyph in all the tables.
	if alias := glyphToAlias(glyph); alias != "" {
		glyph = alias
	}
	if r := glyphToRune(glyph); r != 0 {
		return r, true
	}
	if r := ligatureToCodePoint(glyph); r != 0 {
		return r, true
	}

	// Next try all the glyph naming conventions.
	if groups := reUniEncoding.FindStringSubmatch(glyph); groups != nil {
		n, err := strconv.ParseInt(groups[1], 16, 32)
		if err == nil {
			return rune(n), true
		}
	}

	if groups := reEncoding.FindStringSubmatch(glyph); groups != nil {
		n, err := strconv.Atoi(groups[1])
		if err == nil {
			return rune(n), true
		}
	}

	return 0, false
}

var (
	reEncoding    = regexp.MustCompile(`^[A-Za-z](\d{1,5})$`) // C211
	reUniEncoding = regexp.MustCompile(`^uni([\dA-F]{4})$`)   // uniFB03
	rePrefix      = regexp.MustCompile(`^(\w+)\.\w+$`)        // eight.pnum => eight
)

// ligatureToCodePoint maps ligatures without corresponding unicode code points. We use the Unicode private
// use area (https://en.wikipedia.org/wiki/Private_Use_Areas) to store them.
// These runes are mapped to strings in RuneToString which uses the reverse mappings in
// ligatureToString.
func ligatureToCodePoint(ligature string) rune {
	switch ligature {
	case "f_t":
		return 0xe000
	case "f_j":
		return 0xe001
	case "f_b":
		return 0xe002
	case "f_h":
		return 0xe003
	case "f_k":
		return 0xe004
	case "t_t":
		return 0xe005
	case "t_f":
		return 0xe006
	case "f_f_j":
		return 0xe007
	case "f_f_b":
		return 0xe008
	case "f_f_h":
		return 0xe009
	case "f_f_k":
		return 0xe00a
	case "T_h":
		return 0xe00b
	}

	return 0
}

func glyphToAlias(glyph string) string {
	index := sort.Search(len(glyphToAliasTable), func(i int) bool { return glyphToAliasTable[i].name >= glyph })
	if index < len(glyphToAliasTable) && glyphToAliasTable[index].name == glyph {
		return glyphToAliasTable[index].alias
	} else {
		return ""
	}
}

func glyphToRune(glyph string) rune {
	index := sort.Search(len(glyphToRuneTable), func(i int) bool { return glyphToRuneTable[i].name >= glyph })
	if index < len(glyphToRuneTable) && glyphToRuneTable[index].name == glyph {
		return glyphToRuneTable[index].r
	} else {
		return 0
	}
}
