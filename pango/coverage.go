package pango

// Coverage represents a set of Unicode characters.
// Conceptually, it is a map[rune]bool, but it can be implemented
// much more efficiently.
type Coverage interface {
	// Get returns true if the rune is covered
	Get(index rune) bool
	Set(index rune, covered bool)
	// // Copy returns a deep copy of the coverage
	// Copy() Coverage
}

// func pango_coverage_new() Coverage {
// 	return Coverage{storage: make(map[rune]struct{})}
// }

// func (c Coverage) get(index rune) CoverageLevel {
// 	if _, has := c.storage[index]; has {
// 		return PANGO_COVERAGE_EXACT
// 	}
// 	return PANGO_COVERAGE_NONE
// }

// func (c Coverage) set(index rune, level CoverageLevel) {
// 	if level != PANGO_COVERAGE_NONE {
// 		c.storage[index] = struct{}{}
// 	} else {
// 		delete(c.storage, index)
// 	}
// }

// func (c Coverage) copy() Coverage {
// 	out := Coverage{storage: make(map[rune]struct{}, len(c.storage))}
// 	for r := range c.storage {
// 		out.storage[r] = struct{}{}
// 	}
// 	return out
// }
