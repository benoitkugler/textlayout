package fontconfig

import (
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"strings"
)

// fontconfig/src/fclang.c Copyright © 2002 Keith Packard

type langResult uint8

const (
	langEqual              langResult = 0 // exact match
	langDifferentCountry   langResult = 1
	langDifferentTerritory langResult = 1 // same primary language but not exact
	langDifferentLang      langResult = 2
)

type langToCharset struct {
	lang    string
	charset Charset
}

type langCharsetRange struct {
	begin, end int
}

// Langset holds the set of languages supported by a font.
// These are computed for a font based on orthographic information built into the
// fontconfig library. Fontconfig has orthographies for all of the ISO 639-1
// languages except for MS, NA, PA, PS, QU, RN, RW, SD, SG, SN, SU and ZA.
type Langset struct {
	extra strSet
	page  [langPageSize]uint32
}

var _ hasher = Langset{}

func (l Langset) hash() []byte {
	out := make([]byte, langPageSize*4)
	for i, v := range l.page {
		binary.BigEndian.PutUint32(out[4*i:], v)
	}
	extra := make([]string, len(l.extra))
	for ex := range l.extra {
		extra = append(extra, ex)
	}
	sort.Strings(extra)
	for _, s := range extra {
		out = append(out, s...)
	}
	return out
}

func newCharSetFromLang(lang string) *Charset {
	country := -1

	for i, lcs := range fcLangCharSets {
		switch langCompare(lang, lcs.lang) {
		case langEqual:
			return &lcs.charset
		case langDifferentTerritory:
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

// normalizeLang returns a string to make `lang` suitable on to be used in fontconfig.
func normalizeLang(lang string) string {
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
		if findLangIndex(lang+"_"+tm) < 0 {
		} else {
			return lang + "_" + tm
		}
	}
	if modifier != "" {
		if findLangIndex(lang+"@"+modifier) < 0 {
		} else {
			return lang + "@" + modifier
		}
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

func langCompare(s1, s2 string) langResult {
	result := langDifferentLang

	isUnd := toLower(s1) == 'u' && toLower(s1[1:]) == 'n' &&
		toLower(s1[2:]) == 'd' && langEnd(s1[3:])

	for i := 0; ; i++ {
		c1 := toLower(s1[i:])
		c2 := toLower(s2[i:])
		if c1 != c2 {
			if !isUnd && langEnd(s1[i:]) && langEnd(s2[i:]) {
				return langDifferentTerritory
			}
			return result
		} else if c1 == 0 {
			if isUnd {
				return result
			}
			return langEqual
		} else if c1 == '-' {
			if !isUnd {
				result = langDifferentTerritory
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
		c1 := toLower(super)
		c2 := toLower(sub)
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
	firstChar := toLower(lang)
	var secondChar byte
	if firstChar != 0 {
		secondChar = toLower(lang[1:])
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
			cmp = cmpIgnoreCase(fcLangCharSets[mid].lang, lang)
		} else {
			/* fast path for resolving 2-letter languages (by far the most common) after
			* finding the first char (probably already true because of the hash table) */
			cmp = int(fcLangCharSets[mid].lang[1]) - int(secondChar)
			if cmp == 0 && (len(fcLangCharSets[mid].lang) > 2 || len(lang) > 2) {
				cmp = cmpIgnoreCase(fcLangCharSets[mid].lang[2:], lang[2:])
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
func (ls *Langset) add(lang string) {
	id := findLangIndex(lang)
	if id >= 0 {
		ls.bitSet(id)
		return
	}
	if ls.extra == nil {
		ls.extra = make(strSet)
	}
	ls.extra[lang] = true
}

func (ls Langset) String() string {
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

func (ls *Langset) bitSet(id int) {
	by := fcLangCharSetIndices[id]
	bucket := by >> 5
	ls.page[bucket] |= 1 << (by & 0x1f)
}

func (ls Langset) bitGet(id int) bool {
	by := fcLangCharSetIndices[id]
	bucket := by >> 5
	return (ls.page[bucket]>>(by&0x1f))&1 != 0
}

func (ls *Langset) bitReset(id int) {
	by := fcLangCharSetIndices[id]
	bucket := by >> 5
	ls.page[bucket] &= ^(1 << (by & 0x1f))
}

func langsetEqual(lsa, lsb Langset) bool {
	if lsa.page != lsb.page {
		return false
	}
	return strSetEquals(lsa.extra, lsb.extra)
}

func (ls Langset) containsLang(lang string) bool {
	id := findLangIndex(lang)
	if id < 0 {
		id = -id - 1
	} else if ls.bitGet(id) {
		return true
	}
	// search up and down among equal languages for a match
	for i := id - 1; i >= 0; i-- {
		if langCompare(fcLangCharSets[i].lang, lang) == langDifferentLang {
			break
		}
		if ls.bitGet(i) && langContains(fcLangCharSets[i].lang, lang) {
			return true
		}
	}
	for i := id; i < len(fcLangCharSets); i++ {
		if langCompare(fcLangCharSets[i].lang, lang) == langDifferentLang {
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
func (lsa Langset) includes(lsb Langset) bool {
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

// NewLangset parse a set of language of the form <lang1>|<lang2>|...
// Each language should be of the form Ll-Tt where Ll is a
// two or three letter language from ISO 639 and Tt is a territory from ISO 3166.
func NewLangset(str string) Langset {
	var ls Langset
	for _, lang := range strings.Split(str, "|") {
		ls.add(lang)
	}
	return ls
}

// Copy returns a deep copy of `ls`.
func (ls Langset) Copy() Langset {
	var new Langset
	new.page = ls.page
	new.extra = make(strSet, len(ls.extra))
	for e := range ls.extra {
		new.extra[e] = true
	}
	return new
}

// return true it at least one language has been added
func addLangs(strs strSet, languages string) bool {
	var ret bool
	ls := strings.Split(languages, ":")
	for _, lang := range ls {
		if lang == "" { // ignore an empty item
			continue
		}
		normalizedLang := normalizeLang(lang)
		if normalizedLang != "" {
			strs[normalizedLang] = true
			ret = true
		}
	}

	return ret
}

// Keep Han languages separated by eliminating languages that the codePageRange bits says aren't supported
var codePageRange = [...]struct {
	lang string
	bit  byte
}{
	{"ja", 17},
	{"zh-cn", 18},
	{"ko", 19},
	{"zh-tw", 20},
}

func isExclusiveLang(lang string) bool {
	for _, cp := range codePageRange {
		if langCompare(lang, cp.lang) == langEqual {
			return true
		}
	}
	return false
}

func buildLangSet(charset Charset, exclusiveLang string) Langset {
	var exclusiveCharset *Charset
	if exclusiveLang != "" {
		exclusiveCharset = newCharSetFromLang(exclusiveLang)
	}

	var ls Langset

mainLoop:
	for i, langCharset := range fcLangCharSets {
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
		if missing == 0 {
			ls.bitSet(i)
		}
	}

	return ls
}

func (ls *Langset) del(lang string) {
	id := findLangIndex(lang)
	if id >= 0 {
		ls.bitReset(id)
	} else {
		delete(ls.extra, lang)
	}
}

func (ls *Langset) hasLang(lang string) langResult {
	id := findLangIndex(lang)
	if id < 0 {
		id = -id - 1
	} else if ls.bitGet(id) {
		return langEqual
	}
	best := langDifferentLang
	for i := id - 1; i >= 0; i-- {
		r := langCompare(lang, fcLangCharSets[i].lang)
		if r == langDifferentLang {
			break
		}
		if ls.bitGet(i) && r < best {
			best = r
		}
	}
	for i := id; i < len(fcLangCharSets); i++ {
		r := langCompare(lang, fcLangCharSets[i].lang)
		if r == langDifferentLang {
			break
		}
		if ls.bitGet(i) && r < best {
			best = r
		}
	}
	for extra := range ls.extra {
		if best <= langEqual {
			break
		}
		if r := langCompare(lang, extra); r < best {
			best = r
		}
	}
	return best
}

func (ls *Langset) compareStrSet(set strSet) langResult {
	best := langDifferentLang
	for extra := range set {
		if best <= langEqual {
			break
		}
		if r := ls.hasLang(extra); r < best {
			best = r
		}
	}
	return best
}

func langSetCompare(lsa, lsb Langset) langResult {
	var aInCountrySet, bInCountrySet uint32

	for i := range lsa.page {
		if lsa.page[i]&lsb.page[i] != 0 {
			return langEqual
		}
	}
	best := langDifferentLang
	for _, langCountry := range fcLangCountrySets {
		aInCountrySet = 0
		bInCountrySet = 0

		for i := range lsa.page {
			aInCountrySet |= lsa.page[i] & langCountry[i]
			bInCountrySet |= lsb.page[i] & langCountry[i]

			if aInCountrySet != 0 && bInCountrySet != 0 {
				best = langDifferentTerritory
				break
			}
		}
	}
	if lsa.extra != nil {
		if r := lsb.compareStrSet(lsa.extra); r < best {
			best = r
		}
	}
	if best > langEqual && lsb.extra != nil {
		if r := lsa.compareStrSet(lsb.extra); r < best {
			best = r
		}
	}
	return best
}

func langSetOperate(a, b Langset, fn func(ls *Langset, s string)) Langset {
	langset := a.Copy()
	set := b.getLangs()
	for str := range set {
		fn(&langset, str)
	}
	return langset
}

func langSetUnion(a, b Langset) Langset {
	return langSetOperate(a, b, (*Langset).add)
}

func langSetSubtract(a, b Langset) Langset {
	return langSetOperate(a, b, (*Langset).del)
}

func langSetPromote(lang String) Langset {
	var ls Langset
	if lang != "" {
		id := findLangIndex(string(lang))
		if id >= 0 {
			ls.bitSet(id)
		} else {
			ls.extra = strSet{string(lang): true}
		}
	}
	return ls
}

// Returns a string set of all languages in `ls`.
func (ls Langset) getLangs() strSet {
	langs := make(strSet)

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

//  Bool
//  FcNameUnparseLangSet (FcStrBuf *buf, const langSet *ls)
//  {
// 	 int		i, bit, count;
// 	 FcChar32	bits;
// 	 Bool	first = true;

// 	 count = MIN (NUM_LANG_SET_MAP, NUM_LANG_SET_MAP);
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

//  Bool
//  langSetSerializeAlloc (FcSerialize *serialize, const langSet *l)
//  {
// 	 if (!FcSerializeAlloc (serialize, l, sizeof (langSet)))
// 	 return false;
// 	 return true;
//  }

//  langSet *
//  langSetSerialize(FcSerialize *serialize, const langSet *l)
//  {
// 	 langSet	*l_serialize = FcSerializePtr (serialize, l);

// 	 if (!l_serialize)
// 	 return NULL;
// 	 memset (l_serialize.map_, '\0', sizeof (l_serialize.map_));
// 	 memcpy (l_serialize.map_, l.map_, MIN (sizeof (l_serialize.map_),NUM_LANG_SET_MAP * sizeof (l.map_[0])));
// 	 l_serialiNUM_LANG_SET_MAP = NUM_LANG_SET_MAP;
// 	 l_serialize.extra = NULL; /* We don't serialize ls.extra */
// 	 return l_serialize;
//  }
