package pango

// CoverageLevel indicates how well a font can represent a particular Unicode
// character point for a particular script.
type CoverageLevel uint8

const (
	PANGO_COVERAGE_NONE CoverageLevel = iota // The character is not representable with the font.

	// The character is represented in a way that may be
	// comprehensible but is not the correct graphical form.
	// For instance, a Hangul character represented as a
	// a sequence of Jamos, or a Latin transliteration of a Cyrillic word.
	_ // PANGO_COVERAGE_FALLBACK

	// The character is represented as basically the correct
	// graphical form, but with a stylistic variant inappropriate for
	// the current script.
	_ // PANGO_COVERAGE_APPROXIMATE

	PANGO_COVERAGE_EXACT // The character is represented as the correct graphical form.
)

// Coverage represents a map from Unicode characters to CoverageLevel.
type Coverage struct {
	// TODO: a map is not very efficient,
	// we could use unicode.RangeTable
	storage map[rune]struct{}
}

func pango_coverage_new() Coverage {
	return Coverage{storage: make(map[rune]struct{})}
}

func (c Coverage) get(index rune) CoverageLevel {
	if _, has := c.storage[index]; has {
		return PANGO_COVERAGE_EXACT
	}
	return PANGO_COVERAGE_NONE
}

func (c Coverage) set(index rune, level CoverageLevel) {
	if level != PANGO_COVERAGE_NONE {
		c.storage[index] = struct{}{}
	} else {
		delete(c.storage, index)
	}
}

func (c Coverage) copy() Coverage {
	out := Coverage{storage: make(map[rune]struct{}, len(c.storage))}
	for r := range c.storage {
		out.storage[r] = struct{}{}
	}
	return out
}
