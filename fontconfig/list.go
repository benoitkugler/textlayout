package fontconfig

// ported from fontconfig/src/fclist.c Copyright © 2000 Keith Packard

// fntOrig must have a containing value for every value in patOrig
func listMatchAny(patOrig, fntOrig ValueList) bool {
	for _, pat := range patOrig {
		found := false
		for _, fnt := range fntOrig {
			// make sure the font 'contains' the pattern.
			// (OpListing is OpContains except for strings where it requires an exact match)
			if compareValue(fnt.Value, opWithFlags(FcOpListing, FcOpFlagIgnoreBlanks), pat.Value) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// returns true iff all objects in "p" match "font"
func patternMatchAny(p, font Pattern) bool {
	for object, pe := range p {
		if object == FC_NAMELANG {
			// "namelang" object is the alias object to change "familylang",
			// "stylelang" and "fullnamelang" object all together. it won't be
			// available on the font pattern. so checking its availability
			// causes no results. we should ignore it here.
			continue
		}
		fe := font[object]
		if !listMatchAny(pe, fe) {
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

	e := font[object]
	for i, v := range e {
		if s, ok := v.Value.(String); ok {
			res := FcLangCompare(string(s), lang)
			if res == FcLangEqual {
				return i
			}
			if res == FcLangDifferentCountry && idx < 0 {
				idx = i
			}
			if defidx < 0 {
				// workaround for fonts that has non-English value at the head of values.
				res = FcLangCompare(string(s), "en")
				if res == FcLangEqual {
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
	tablePat := table[hash]

	for _, obj := range os {
		switch obj {
		case FC_FAMILY, FC_FAMILYLANG:
			if familyidx < 0 {
				familyidx = font.getDefaultObjectLangIndex(FC_FAMILYLANG, lang)
			}
			defidx = familyidx
		case FC_FULLNAME, FC_FULLNAMELANG:
			if fullnameidx < 0 {
				fullnameidx = font.getDefaultObjectLangIndex(FC_FULLNAMELANG, lang)
			}
			defidx = fullnameidx
		case FC_STYLE, FC_STYLELANG:
			if styleidx < 0 {
				styleidx = font.getDefaultObjectLangIndex(FC_STYLELANG, lang)
			}
			defidx = styleidx
		default:
			defidx = 0
		}

		e := font[obj]
		for idx, v := range e {
			tablePat.Add(obj, v.Value, defidx != idx)
		}
	}
}

func getDefaultLang() string {
	langs := FcGetDefaultLangs()
	for s := range langs {
		return s
	}
	return ""
}

func fontSetList(config *Config, sets []FontSet, p Pattern, os []Object) FontSet {
	table := make(map[string]Pattern)

	// Walk all available fonts adding those that match to the hash table
	for _, s := range sets {
		for _, font := range s {
			if patternMatchAny(p, font) {
				lang, res := p.FcPatternObjectGetString(FC_NAMELANG, 0)
				if res != FcResultMatch {
					lang = getDefaultLang()
				}
				listAppend(table, font, os, lang)
			}
		}
	}

	// Walk the hash table and build a font set
	ret := make(FontSet, 0, len(table))
	for _, font := range table {
		ret = append(ret, font)
	}
	return ret
}

// List selects fonts matching `p` (all if it is nil), creates patterns from those fonts containing
// only the objects in `objs` and returns the set of unique such patterns.
// TODO: check the call with nil config
func List(config *Config, p Pattern, objs ...Object) FontSet {
	var sets []FontSet
	if config.fonts[FcSetSystem] != nil {
		sets = append(sets, config.fonts[FcSetSystem])
	}
	if config.fonts[FcSetApplication] != nil {
		sets = append(sets, config.fonts[FcSetApplication])
	}
	return fontSetList(config, sets, p, objs)
}
