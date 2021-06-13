package fontconfig

import "fmt"

// ported from fontconfig/src/fclist.c Copyright © 2000 Keith Packard

// font must have a containing value for every value in query
func listMatchAny(query, font valueList) bool {
	for _, pat := range query {
		found := false
		for _, fnt := range font {
			// make sure the font 'contains' the pattern.
			// (OpListing is OpContains except for strings where it requires an exact match)
			if compareValue(fnt.Value, opWithFlags(opListing, opFlagIgnoreBlanks), pat.Value) {
				found = true
				if debugMode {
					fmt.Printf("\t\tmatching values: %v and %v\n", fnt.Value, pat.Value)
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// returns true if and only if all objects in "p" match "font"
func patternMatchAny(p, font Pattern) bool {
	for object, pe := range p {
		if object == NAMELANG {
			// "namelang" object is the alias object to change "familylang",
			// "stylelang" and "fullnamelang" object all together. it won't be
			// available on the font pattern. so checking its availability
			// causes no results. we should ignore it here.
			continue
		}
		fe := font.getVals(object)
		if debugMode {
			fmt.Printf("\tmatching object %s...\n", object)
		}
		match := listMatchAny(*pe, fe)
		if !match {
			return false
		}
	}
	return true
}

// restrict the hash to the objects in `objs`
func patternHash(font Pattern, objs []Object) string {
	crible := make(Pattern, len(objs))
	for _, obj := range objs {
		crible[obj] = font[obj]
	}
	return crible.Hash()
}

func (font Pattern) getDefaultObjectLangIndex(object Object, lang string) int {
	idx := -1
	defidx := -1

	e := font.getVals(object)
	for i, v := range e {
		if s, ok := v.Value.(String); ok {
			res := langCompare(string(s), lang)
			if res == langEqual {
				return i
			}
			if res == langDifferentCountry && idx < 0 {
				idx = i
			}
			if defidx < 0 {
				// workaround for fonts that has non-English value at the head of values.
				res = langCompare(string(s), "en")
				if res == langEqual {
					defidx = i
				}
			}
		}
	}

	if idx > 0 {
		return idx
	}
	if defidx > 0 {
		return defidx
	}
	return 0
}

func listAppend(table map[string]Pattern, font Pattern, os []Object, lang string) {
	familyidx := -1
	defidx := 0
	fullnameidx := -1
	styleidx := -1

	hash := patternHash(font, os)
	if _, has := table[hash]; has {
		return
	}

	pattern := NewPattern()
	for _, obj := range os {
		switch obj {
		case FAMILY, FAMILYLANG:
			if familyidx < 0 {
				familyidx = font.getDefaultObjectLangIndex(FAMILYLANG, lang)
			}
			defidx = familyidx
		case FULLNAME, FULLNAMELANG:
			if fullnameidx < 0 {
				fullnameidx = font.getDefaultObjectLangIndex(FULLNAMELANG, lang)
			}
			defidx = fullnameidx
		case STYLE, STYLELANG:
			if styleidx < 0 {
				styleidx = font.getDefaultObjectLangIndex(STYLELANG, lang)
			}
			defidx = styleidx
		default:
			defidx = 0
		}

		e := font.getVals(obj)
		for idx, v := range e {
			pattern.Add(obj, v.Value, defidx != idx)
		}
	}
	table[hash] = pattern
}

func getDefaultLang() string {
	langs := getDefaultLangs()
	for s := range langs {
		return s
	}
	return ""
}

// List selects fonts matching `p` (all if p is empty), creates patterns from those fonts containing
// only the objects in `objs` and returns the set of unique such patterns.
// If no objects are specified, default to all builtin objects.
func (set Fontset) List(p Pattern, objs ...Object) Fontset {
	table := make(map[string]Pattern)

	if len(objs) == 0 { // default to all objects
		for i := Object(1); i < FirstCustomObject; i++ {
			objs = append(objs, i)
		}
	}

	// Walk all available fonts adding those that match to the hash table
	for _, font := range set {
		if patternMatchAny(p, font) {
			lang, res := p.getAtString(NAMELANG, 0)
			if res != ResultMatch {
				lang = getDefaultLang()
			}
			listAppend(table, font, objs, lang)
		}
	}

	// Walk the hash table and build a font set
	ret := make(Fontset, 0, len(table))
	for _, font := range table {
		ret = append(ret, font)
	}
	return ret
}
