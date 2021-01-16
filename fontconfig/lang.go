package fontconfig

import (
	"fmt"
	"log"
	"strings"
)

// fontconfig/src/fclang.c Copyright © 2002 Keith Packard

type FcLangResult uint8

const (
	FcLangEqual              FcLangResult = 0
	FcLangDifferentCountry   FcLangResult = 1
	FcLangDifferentTerritory FcLangResult = 1
	FcLangDifferentLang      FcLangResult = 2
)

/* Objects MT-safe for readonly access. */

type FcLangCharSet struct {
	lang    string
	charset FcCharset
}

type FcLangCharSetRange struct {
	begin, end int
}

// FcLangSet holds the set of languages supported
// by a font.
// These are computed for a font based on orthographic information built into the
// fontconfig library. Fontconfig has orthographies for all of the ISO 639-1
// languages except for MS, NA, PA, PS, QU, RN, RW, SD, SG, SN, SU and ZA.
type FcLangSet struct {
	extra FcStrSet
	page  [langPageSize]uint32
}

func newCharSetFromLang(lang string) *FcCharset {
	country := -1

	for i, lcs := range fcLangCharSets {
		switch FcLangCompare(lang, lcs.lang) {
		case FcLangEqual:
			return &lcs.charset
		case FcLangDifferentTerritory:
			if country == -1 {
				country = i
			}
		}
	}
	if country == -1 {
		return nil
	}
	return &fcLangCharSets[country].charset
}

func FcLangNormalize(lang string) string {
	var (
		result string
		orig   = lang
	)

	lang = strings.ToLower(lang)
	switch lang {
	case "c", "c.utf-8, c.utf8", "posix":
		return "en"
	}

	/* from the comments in glibc:
	 *
	 * LOCALE can consist of up to four recognized parts for the XPG syntax:
	 *
	 *            language[_territory[.codeset]][@modifier]
	 *
	 * Beside the first all of them are allowed to be missing.  If the
	 * full specified locale is not found, the less specific one are
	 * looked for.  The various part will be stripped off according to
	 * the following order:
	 *            (1) codeset
	 *            (2) normalized codeset
	 *            (3) territory
	 *            (4) modifier
	 *
	 * So since we don't take care of the codeset part here, what patterns
	 * we need to deal with is:
	 *
	 *   1. language_territory@modifier
	 *   2. language@modifier
	 *   3. language
	 *
	 * then. and maybe no need to try language_territory here.
	 */
	var modifier, territory string

	if modifierI := strings.IndexByte(lang, '@'); modifierI != -1 {
		modifier = lang[modifierI+1:]
		lang = lang[0:modifierI]
	}
	encoding := strings.IndexByte(lang, '.')
	if encoding != -1 {
		lang = lang[0:encoding]
	}
	territoryI := strings.IndexByte(lang, '_')
	if territoryI == -1 {
		territoryI = strings.IndexByte(lang, '-')
	}
	if territoryI != -1 {
		territory = lang[territoryI+1:]
		lang = lang[0:territoryI]
	}
	llen := len(lang)
	tlen := len(territory)
	tm := territory
	if llen < 2 || llen > 3 {
		log.Printf("fontconfig: ignoring %s: not a valid language tag", lang)
		return result
	}
	if tlen != 0 && (tlen < 2 || tlen > 3) && !(territory[0] == 'z' && tlen < 5) {
		log.Printf("fontconfig: ignoring %s: not a valid region tag", lang)
		return result
	}
	if modifier != "" {
		tm += "@" + modifier
	}
	if territory != "" {
		if debugMode {
			fmt.Printf("Checking the existence of %s.orth\n", lang+"_"+tm)
		}
		if findLangIndex(lang+"_"+tm) < 0 {
		} else {
			return lang + "_" + tm
		}
	}
	if modifier != "" {
		if debugMode {
			fmt.Printf("Checking the existence of %s.orth\n", lang+"@"+modifier)
		}
		if findLangIndex(lang+"@"+modifier) < 0 {
		} else {
			return lang + "@" + modifier
		}
	}
	if debugMode {
		fmt.Printf("Checking the existence of %s.orth\n", lang)
	}
	if findLangIndex(lang) < 0 {
		// there seems no languages matched in orth. add the language as is for fallback.
		result = orig
	} else {
		result = lang
	}

	return result
}

func langEnd(c string) bool {
	return c == "" || c[0] == '-'
}

func FcLangCompare(s1, s2 string) FcLangResult {
	result := FcLangDifferentLang

	isUnd := FcToLower(s1) == 'u' && FcToLower(s1[1:]) == 'n' &&
		FcToLower(s1[2:]) == 'd' && langEnd(s1[3:])

	for i := 0; ; i++ {
		c1 := FcToLower(s1[i:])
		c2 := FcToLower(s2[i:])
		if c1 != c2 {
			if !isUnd && langEnd(s1[i:]) && langEnd(s2[i:]) {
				result = FcLangDifferentTerritory
			}
			return result
		} else if c1 == 0 {
			if isUnd {
				return result
			}
			return FcLangEqual
		} else if c1 == '-' {
			if !isUnd {
				result = FcLangDifferentTerritory
			}
		}

		// If we parsed past "und-", then do not consider it undefined anymore,
		// as there's *something* specified.
		if isUnd && i == 3 {
			isUnd = false
		}
	}
}

// Return true when super contains sub.
//
// super contains sub if super and sub have the same
// language and either the same country or one
// is missing the country
func langContains(super, sub string) bool {
	for {
		c1 := FcToLower(super)
		c2 := FcToLower(sub)
		if c1 != c2 {
			// see if super has a country for sub is missing one
			if c1 == '-' && c2 == 0 {
				return true
			}
			// see if sub has a country for super is missing one
			if c1 == 0 && c2 == '-' {
				return true
			}
			return false
		} else if c1 == 0 {
			return true
		}
		super, sub = super[1:], sub[1:]
	}
}

//  FcStrSet *
//  FcGetLangs (void)
//  {
// 	 FcStrSet *langs;
// 	 int	i;

// 	 langs = FcStrSetCreate();
// 	 if (!langs)
// 	 return 0;

// 	 for (i = 0; i < NUM_LANG_CHAR_SET; i++)
// 	 FcStrSetAdd (langs, fcLangCharSets[i].lang);

// 	 return langs;
//  }

//  FcLangSet *
//  FcLangSetCreate (void)
//  {
// 	 FcLangSet	*ls;

// 	 ls = malloc (sizeof (FcLangSet));
// 	 if (!ls)
// 	 return 0;
// 	 memset (ls.map_, '\0', sizeof (ls.map_));
// 	 NUM_LANG_SET_MAP = NUM_LANG_SET_MAP;
// 	 ls.extra = 0;
// 	 return ls;
//  }

//  void
//  FcLangSetDestroy (FcLangSet *ls)
//  {
// 	 if (!ls)
// 	 return;

// 	 if (ls.extra)
// 	 FcStrSetDestroy (ls.extra);
// 	 free (ls);
//  }

/* When the language isn't found, the return value r is such that:
 *  1) r < 0
 *  2) -r -1 is the index of the first language in fcLangCharSets that comes
 *     after the 'lang' argument in lexicographic order.
 *
 *  The -1 is necessary to avoid problems with language id 0 (otherwise, we
 *  wouldn't be able to distinguish between “language found, id is 0” and
 *  “language not found, sorts right before the language with id 0”).
 */
func findLangIndex(lang string) int {
	firstChar := FcToLower(lang)
	var secondChar byte
	if firstChar != 0 {
		secondChar = FcToLower(lang[1:])
	}

	var low, high, mid, cmp int
	if firstChar < 'a' {
		low = 0
		high = fcLangCharSetRanges[0].begin
	} else if firstChar > 'z' {
		low = fcLangCharSetRanges[25].begin
		high = len(fcLangCharSets) - 1
	} else {
		low = fcLangCharSetRanges[firstChar-'a'].begin
		high = fcLangCharSetRanges[firstChar-'a'].end
		/* no matches */
		if low > high {
			/* one past next entry after where it would be */
			return -(low + 1)
		}
	}
	for low <= high {
		mid = (high + low) >> 1
		if fcLangCharSets[mid].lang[0] != firstChar {
			cmp = FcStrCmpIgnoreCase(fcLangCharSets[mid].lang, lang)
		} else {
			/* fast path for resolving 2-letter languages (by far the most common) after
			 * finding the first char (probably already true because of the hash table) */
			cmp = int(fcLangCharSets[mid].lang[1] - secondChar)
			if cmp == 0 && (len(fcLangCharSets[mid].lang) >= 2 || len(lang) >= 2) {
				cmp = FcStrCmpIgnoreCase(fcLangCharSets[mid].lang[2:], lang[2:])
			}
		}
		if cmp == 0 {
			return mid
		}
		if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if cmp < 0 {
		mid++
	}
	return -(mid + 1)
}

// add adds `lang` to `ls`.
// `lang` should be of the form Ll-Tt where Ll is a
// two or three letter language from ISO 639 and Tt is a territory from ISO 3166.
func (ls *FcLangSet) add(lang string) {
	id := findLangIndex(lang)
	if id >= 0 {
		ls.bitSet(id)
		return
	}
	if ls.extra == nil {
		ls.extra = make(FcStrSet)
	}
	ls.extra[lang] = true
}

func (ls FcLangSet) String() string {
	var chunks []string

	for i, bits := range ls.page {
		if bits != 0 {
			for bit := 0; bit <= 31; bit++ {
				if bits&(1<<bit) != 0 {
					id := (i << 5) | bit
					chunks = append(chunks, fcLangCharSets[fcLangCharSetIndicesInv[id]].lang)
				}
			}
		}
	}

	for extra := range ls.extra {
		chunks = append(chunks, extra)
	}
	return strings.Join(chunks, "|")
}

func (ls *FcLangSet) bitSet(id int) {
	by := fcLangCharSetIndices[id]
	bucket := by >> 5
	ls.page[bucket] |= 1 << (by & 0x1f)
}

func (ls FcLangSet) bitGet(id int) bool {
	by := fcLangCharSetIndices[id]
	bucket := by >> 5
	return (ls.page[bucket]>>(by&0x1f))&1 != 0
}

func (ls *FcLangSet) bitReset(id int) {
	by := fcLangCharSetIndices[id]
	bucket := by >> 5
	ls.page[bucket] &= ^(1 << (by & 0x1f))
}

func FcLangSetEqual(lsa, lsb FcLangSet) bool {
	if lsa.page != lsb.page {
		return false
	}
	return FcStrSetEqual(lsa.extra, lsb.extra)
}

func (ls FcLangSet) containsLang(lang string) bool {
	id := findLangIndex(lang)
	if id < 0 {
		id = -id - 1
	} else if ls.bitGet(id) {
		return true
	}
	// search up and down among equal languages for a match
	for i := id - 1; i >= 0; i-- {
		if FcLangCompare(fcLangCharSets[i].lang, lang) == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && langContains(fcLangCharSets[i].lang, lang) {
			return true
		}
	}
	for i := id; i < len(fcLangCharSets); i++ {
		if FcLangCompare(fcLangCharSets[i].lang, lang) == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && langContains(fcLangCharSets[i].lang, lang) {
			return true
		}
	}

	var extra string
	for extra = range ls.extra {
		if langContains(extra, lang) {
			break
		}
	}
	return extra != ""
}

// return true if lsa contains every language in lsb
func (lsa FcLangSet) FcLangSetContains(lsb FcLangSet) bool {
	if debugMode {
		fmt.Println("FcLangSet ", lsa)
		fmt.Println(" contains ", lsb)
		fmt.Println("")
	}
	// check bitmaps for missing language support
	for i := range lsb.page {
		missing := lsb.page[i] & ^lsa.page[i]
		if missing != 0 {
			for j := 0; j < 32; j++ {
				if missing&(1<<j) != 0 {
					tmpL := fcLangCharSets[fcLangCharSetIndicesInv[i*32+j]].lang
					if !lsa.containsLang(tmpL) {
						if debugMode {
							fmt.Printf("\tMissing bitmap %s\n", tmpL)
						}
						return false
					}
				}
			}
		}
	}
	var extra string
	for extra := range lsb.extra {
		if !lsa.containsLang(extra) {
			if debugMode {
				fmt.Printf("\tMissing string %s\n", extra)
			}
			break
		}
	}
	if extra != "" {
		return false
	}
	return true
}

func FcNameParseLangSet(str string) FcLangSet {
	var ls FcLangSet
	for _, lang := range strings.Split(str, "|") {
		ls.add(lang)
	}
	return ls
}

// copy creates a new FcLangSet object and
// populates it with the contents of `ls`.
func (ls FcLangSet) copy() FcLangSet {
	var new FcLangSet
	new.page = ls.page
	new.extra = make(FcStrSet, len(ls.extra))
	for e := range ls.extra {
		new.extra[e] = true
	}
	return new
}

func FcStrSetAddLangs(strs FcStrSet, languages string) bool {
	var ret bool
	ls := strings.Split(languages, ":")
	for _, lang := range ls {
		if lang == "" { // ignore an empty item
			continue
		}
		normalizedLang := FcLangNormalize(lang)
		if normalizedLang != "" {
			strs[normalizedLang] = true
			ret = true
		}
	}

	return ret
}

// Keep Han languages separated by eliminating languages that the codePageRange bits says aren't supported
var codePageRange = [...]struct {
	bit  byte
	lang string
}{
	{17, "ja"},
	{18, "zh-cn"},
	{19, "ko"},
	{20, "zh-tw"},
}

func isExclusiveLang(lang string) bool {
	for _, cp := range codePageRange {
		if FcLangCompare(lang, cp.lang) == FcLangEqual {
			return true
		}
	}
	return false
}

func buildLangSet(charset FcCharset, exclusiveLang string) FcLangSet {
	var exclusiveCharset *FcCharset
	if exclusiveLang != "" {
		exclusiveCharset = newCharSetFromLang(exclusiveLang)
	}

	var ls FcLangSet

	if debugMode {
		fmt.Printf("font charset %v \n", charset)
	}

mainLoop:
	for i, langCharset := range fcLangCharSets {
		if debugMode {
			fmt.Printf("%s charset %v\n", langCharset.lang, langCharset.charset)
		}

		/*
		 * Check for Han charsets to make fonts
		 * which advertise support for a single language
		 * not support other Han languages
		 */
		if exclusiveCharset != nil && isExclusiveLang(langCharset.lang) {
			if len(langCharset.charset.pageNumbers) != len(exclusiveCharset.pageNumbers) {
				continue
			}

			for j, leaf := range langCharset.charset.pages {
				if leaf != exclusiveCharset.pages[j] {
					continue mainLoop
				}
			}
		}

		missing := charsetSubtractCount(langCharset.charset, charset)
		if debugMode {
			if missing != 0 && missing < 10 {
				missed := charsetSubtract(langCharset.charset, charset)
				fmt.Printf("\n%s(%d) {", langCharset.lang, missing)
				for pos, page := range missed.pages {
					ucs4 := int(missed.pageNumbers[pos] << 8)
					for i, v := range page {
						for j := 0; j < 32; j++ {
							if v&(1<<j) != 0 {
								fmt.Printf(" %04x", ucs4+i*32+j)
							}
						}
					}
				}
				fmt.Printf(" }\n\t")
			} else {
				fmt.Printf("%s(%d) ", langCharset.lang, missing)
			}
		}
		if missing == 0 {
			ls.bitSet(i)
		}
	}

	return ls
}

func (ls *FcLangSet) del(lang string) {
	id := findLangIndex(lang)
	if id >= 0 {
		ls.bitReset(id)
	} else {
		delete(ls.extra, lang)
	}
}

func (ls *FcLangSet) hasLang(lang string) FcLangResult {
	id := findLangIndex(lang)
	if id < 0 {
		id = -id - 1
	} else if ls.bitGet(id) {
		return FcLangEqual
	}
	best := FcLangDifferentLang
	for i := id - 1; i >= 0; i-- {
		r := FcLangCompare(lang, fcLangCharSets[i].lang)
		if r == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && r < best {
			best = r
		}
	}
	for i := id; i < len(fcLangCharSets); i++ {
		r := FcLangCompare(lang, fcLangCharSets[i].lang)
		if r == FcLangDifferentLang {
			break
		}
		if ls.bitGet(i) && r < best {
			best = r
		}
	}
	for extra := range ls.extra {
		if best <= FcLangEqual {
			break
		}
		if r := FcLangCompare(lang, extra); r < best {
			best = r
		}
	}
	return best
}

func (ls *FcLangSet) compareStrSet(set FcStrSet) FcLangResult {
	best := FcLangDifferentLang
	for extra := range set {
		if best <= FcLangEqual {
			break
		}
		if r := ls.hasLang(extra); r < best {
			best = r
		}
	}
	return best
}

func FcLangSetCompare(lsa, lsb FcLangSet) FcLangResult {
	var aInCountrySet, bInCountrySet uint32

	for i := range lsa.page {
		if lsa.page[i]&lsb.page[i] != 0 {
			return FcLangEqual
		}
	}
	best := FcLangDifferentLang
	for _, langCountry := range fcLangCountrySets {
		aInCountrySet = 0
		bInCountrySet = 0

		for i := range lsa.page {
			aInCountrySet |= lsa.page[i] & langCountry[i]
			bInCountrySet |= lsb.page[i] & langCountry[i]

			if aInCountrySet != 0 && bInCountrySet != 0 {
				best = FcLangDifferentTerritory
				break
			}
		}
	}
	if lsa.extra != nil {
		if r := lsb.compareStrSet(lsa.extra); r < best {
			best = r
		}
	}
	if best > FcLangEqual && lsb.extra != nil {
		if r := lsa.compareStrSet(lsb.extra); r < best {
			best = r
		}
	}
	return best
}

func langSetOperate(a, b FcLangSet, fn func(ls *FcLangSet, s string)) FcLangSet {
	langset := a.copy()
	set := b.getLangs()
	for str := range set {
		fn(&langset, str)
	}
	return langset
}

func langSetUnion(a, b FcLangSet) FcLangSet {
	return langSetOperate(a, b, (*FcLangSet).add)
}

func langSetSubtract(a, b FcLangSet) FcLangSet {
	return langSetOperate(a, b, (*FcLangSet).del)
}

func langSetPromote(lang String) FcLangSet {
	var ls FcLangSet
	if lang != "" {
		id := findLangIndex(string(lang))
		if id >= 0 {
			ls.bitSet(id)
		} else {
			ls.extra = FcStrSet{string(lang): true}
		}
	}
	return ls
}

// Returns a string set of all languages in `ls`.
func (ls FcLangSet) getLangs() FcStrSet {
	langs := make(FcStrSet)

	for i, lg := range fcLangCharSets {
		if ls.bitGet(i) {
			langs[lg.lang] = true
		}
	}

	for extra := range ls.extra {
		langs[extra] = true
	}

	return langs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//  FcChar32
//  FcLangSetHash (const FcLangSet *ls)
//  {
// 	 FcChar32	h = 0;
// 	 int		i, count;

// 	 count = FC_MIN (NUM_LANG_SET_MAP, NUM_LANG_SET_MAP);
// 	 for (i = 0; i < count; i++)
// 	 h ^= ls.map_[i];
// 	 if (ls.extra)
// 	 h ^= ls.extra.num;
// 	 return h;
//  }

//  FcBool
//  FcNameUnparseLangSet (FcStrBuf *buf, const FcLangSet *ls)
//  {
// 	 int		i, bit, count;
// 	 FcChar32	bits;
// 	 FcBool	first = true;

// 	 count = FC_MIN (NUM_LANG_SET_MAP, NUM_LANG_SET_MAP);
// 	 for (i = 0; i < count; i++)
// 	 {
// 	 if ((bits = ls.map_[i]))
// 	 {
// 		 for (bit = 0; bit <= 31; bit++)
// 		 if (bits & (1U << bit))
// 		 {
// 			 int id = (i << 5) | bit;
// 			 if (!first)
// 			 if (!FcStrBufChar (buf, '|'))
// 				 return false;
// 			 if (!FcStrBufString (buf, fcLangCharSets[fcLangCharSetIndicesInv[id]].lang))
// 			 return false;
// 			 first = false;
// 		 }
// 	 }
// 	 }
// 	 if (ls.extra)
// 	 {
// 	 FcStrList   *list = FcStrListCreate (ls.extra);
// 	 FcChar8	    *extra;

// 	 if (!list)
// 		 return false;
// 	 for ((extra = FcStrListNext (list)))
// 	 {
// 		 if (!first)
// 		 if (!FcStrBufChar (buf, '|'))
// 				 {
// 					 FcStrListDone (list);
// 			 return false;
// 				 }
// 		 if (!FcStrBufString (buf, extra))
// 				 {
// 					 FcStrListDone (list);
// 					 return false;
// 				 }
// 		 first = false;
// 	 }
// 		 FcStrListDone (list);
// 	 }
// 	 return true;
//  }

//  FcBool
//  FcLangSetSerializeAlloc (FcSerialize *serialize, const FcLangSet *l)
//  {
// 	 if (!FcSerializeAlloc (serialize, l, sizeof (FcLangSet)))
// 	 return false;
// 	 return true;
//  }

//  FcLangSet *
//  FcLangSetSerialize(FcSerialize *serialize, const FcLangSet *l)
//  {
// 	 FcLangSet	*l_serialize = FcSerializePtr (serialize, l);

// 	 if (!l_serialize)
// 	 return NULL;
// 	 memset (l_serialize.map_, '\0', sizeof (l_serialize.map_));
// 	 memcpy (l_serialize.map_, l.map_, FC_MIN (sizeof (l_serialize.map_),NUM_LANG_SET_MAP * sizeof (l.map_[0])));
// 	 l_serialiNUM_LANG_SET_MAP = NUM_LANG_SET_MAP;
// 	 l_serialize.extra = NULL; /* We don't serialize ls.extra */
// 	 return l_serialize;
//  }
