// Generator of the mapping from OpenType tags to BCP 47 tags && vice
// versa.
//
// It creates an array, matching the tags from the OpenType
// languages system tag list to the language subtags of the BCP 47 language
// subtag registry, with some manual adjustments. The mappings are
// supplemented with macrolanguages' sublanguages && retired codes'
// replacements, according to BCP 47 && some manual additions where BCP 47
// omits a retired code entirely.
//
// Also generated is a function, `hb_ot_ambiguous_tag_to_language`,
// intended for use by `hb_ot_tag_to_language`. It maps OpenType tags
// back to BCP 47 tags. Ambiguous OpenType tags (those that correspond to
// multiple BCP 47 tags) are listed here, except when the alphabetically
// first BCP 47 tag happens to be the chosen disambiguated tag. In that
// case, the fallback behavior will choose the right tag anyway.
//
// Input files:
// * https://docs.microsoft.com/en-us/typography/opentype/spec/languagetags
// * https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/text/unicode/norm"
)

var (
	bcp47 = newBCP47Parser()
	ot    = newOpenTypeRegistryParser()
)

// def write (s):
// 	sys.stdout.flush ()
// 	sys.stdout.buffer.write (s.encode ('utf-8'))

// def html_unescape (parser, entity):
// 	return unescape (entity)

func expect(condition bool, message string) {
	if !condition {
		log.Fatal("assertion error ", message)
	}
}

const DEFAULT_LANGUAGE_SYSTEM = ""

// from https://www-01.sil.org/iso639-3/iso-639-3.tab
var ISO_639_3_TO_1 = map[string]string{
	"aar": "aa",
	"abk": "ab",
	"afr": "af",
	"aka": "ak",
	"amh": "am",
	"ara": "ar",
	"arg": "an",
	"asm": "as",
	"ava": "av",
	"ave": "ae",
	"aym": "ay",
	"aze": "az",
	"bak": "ba",
	"bam": "bm",
	"bel": "be",
	"ben": "bn",
	"bis": "bi",
	"bod": "bo",
	"bos": "bs",
	"bre": "br",
	"bul": "bg",
	"cat": "ca",
	"ces": "cs",
	"cha": "ch",
	"che": "ce",
	"chu": "cu",
	"chv": "cv",
	"cor": "kw",
	"cos": "co",
	"cre": "cr",
	"cym": "cy",
	"dan": "da",
	"deu": "de",
	"div": "dv",
	"dzo": "dz",
	"ell": "el",
	"eng": "en",
	"epo": "eo",
	"est": "et",
	"eus": "eu",
	"ewe": "ee",
	"fao": "fo",
	"fas": "fa",
	"fij": "fj",
	"fin": "fi",
	"fra": "fr",
	"fry": "fy",
	"ful": "ff",
	"gla": "gd",
	"gle": "ga",
	"glg": "gl",
	"glv": "gv",
	"grn": "gn",
	"guj": "gu",
	"hat": "ht",
	"hau": "ha",
	"hbs": "sh",
	"heb": "he",
	"her": "hz",
	"hin": "hi",
	"hmo": "ho",
	"hrv": "hr",
	"hun": "hu",
	"hye": "hy",
	"ibo": "ig",
	"ido": "io",
	"iii": "ii",
	"iku": "iu",
	"ile": "ie",
	"ina": "ia",
	"ind": "id",
	"ipk": "ik",
	"isl": "is",
	"ita": "it",
	"jav": "jv",
	"jpn": "ja",
	"kal": "kl",
	"kan": "kn",
	"kas": "ks",
	"kat": "ka",
	"kau": "kr",
	"kaz": "kk",
	"khm": "km",
	"kik": "ki",
	"kin": "rw",
	"kir": "ky",
	"kom": "kv",
	"kon": "kg",
	"kor": "ko",
	"kua": "kj",
	"kur": "ku",
	"lao": "lo",
	"lat": "la",
	"lav": "lv",
	"lim": "li",
	"lin": "ln",
	"lit": "lt",
	"ltz": "lb",
	"lub": "lu",
	"lug": "lg",
	"mah": "mh",
	"mal": "ml",
	"mar": "mr",
	"mkd": "mk",
	"mlg": "mg",
	"mlt": "mt",
	"mol": "mo",
	"mon": "mn",
	"mri": "mi",
	"msa": "ms",
	"mya": "my",
	"nau": "na",
	"nav": "nv",
	"nbl": "nr",
	"nde": "nd",
	"ndo": "ng",
	"nep": "ne",
	"nld": "nl",
	"nno": "nn",
	"nob": "nb",
	"nor": "no",
	"nya": "ny",
	"oci": "oc",
	"oji": "oj",
	"ori": "||",
	"orm": "om",
	"oss": "os",
	"pan": "pa",
	"pli": "pi",
	"pol": "pl",
	"por": "pt",
	"pus": "ps",
	"que": "qu",
	"roh": "rm",
	"ron": "ro",
	"run": "rn",
	"rus": "ru",
	"sag": "sg",
	"san": "sa",
	"sin": "si",
	"slk": "sk",
	"slv": "sl",
	"sme": "se",
	"smo": "sm",
	"sna": "sn",
	"snd": "sd",
	"som": "so",
	"sot": "st",
	"spa": "es",
	"sqi": "sq",
	"srd": "sc",
	"srp": "sr",
	"ssw": "ss",
	"sun": "su",
	"swa": "sw",
	"swe": "sv",
	"tah": "ty",
	"tam": "ta",
	"tat": "tt",
	"tel": "te",
	"tgk": "tg",
	"tgl": "tl",
	"tha": "th",
	"tir": "ti",
	"ton": "to",
	"tsn": "tn",
	"tso": "ts",
	"tuk": "tk",
	"tur": "tr",
	"twi": "tw",
	"uig": "ug",
	"ukr": "uk",
	"urd": "ur",
	"uzb": "uz",
	"ven": "ve",
	"vie": "vi",
	"vol": "vo",
	"wln": "wa",
	"wol": "wo",
	"xho": "xh",
	"yid": "yi",
	"yor": "yo",
	"zha": "za",
	"zho": "zh",
	"zul": "zu",
}

// a BCP 47 language tag
type LanguageTag struct {
	subtags []string // The list of subtags in this tag.
	// Whether this tag is grandfathered.
	// If ``true``, the entire lowercased tag is the ``language``
	// and the other subtag fields are empty.
	grandfathered bool
	language      string // The language subtag.
	script        string // The script subtag.
	region        string // The region subtag.
	variant       string // The variant subtag.
}

func findFirst(fn func(string) bool, l []string) string {
	for _, s := range l {
		if fn(s) {
			return s
		}
	}
	return ""
}

// tag is a BCP 47 language tag.
func newLanguageTag(tag string) LanguageTag {
	var out LanguageTag
	tag = strings.ToLower(tag)
	out.subtags = strings.Split(tag, "-")
	_, out.grandfathered = bcp47.grandfathered[tag]
	if out.grandfathered {
		out.language = tag
	} else {
		out.language = out.subtags[0]
		out.script = findFirst(func(s string) bool { return len(s) == 4 && s[0] > '9' }, out.subtags)
		out.region = findFirst(func(s string) bool { return len(s) == 2 && s[0] > '9' || len(s) == 3 && s[0] <= '9' }, out.subtags[1:])
		out.variant = findFirst(func(s string) bool { return len(s) > 4 || len(s) == 4 && s[0] <= '9' }, out.subtags)
	}
	return out
}

// Return whether this tag is too complex to represent as a
// ``LangTag`` in the generated code.
//
// Complex tags need to be handled in
// ``hb_ot_tags_from_complex_language``.
func (l LanguageTag) isComplex() bool {
	return !(len(l.subtags) == 1 || l.grandfathered &&
		len(l.subtags[1]) != 3 && setEqual(ot.from_bcp_47[l.subtags[0]], ot.from_bcp_47[l.language]))
}

// Return the group into which this tag should be categorized in
// ``hb_ot_tags_from_complex_language``.
//
// The group is the first letter of the tag, or ``'und'`` if this tag
// should not be matched in a ``switch`` statement in the generated
// code.
func (l LanguageTag) getGroup() string {
	if l.language == "und" || len(bcp47.prefixes[l.variant]) == 1 {
		return "und"
	}
	return l.language[0:1]
}

// a parser for the OpenType language system tag registry
type OpenTypeRegistryParser struct {
	header string            // The "last updated" line of the registry.
	names  map[string]string // A map of language system tags to the names they are given in the registry.
	// A map of language system tags to
	// numbers. If a single BCP 47 tag corresponds to multiple
	// OpenType tags, the tags are ordered in increasing order by
	// rank. The rank is based on the number of BCP 47 tags
	// associated with a tag, though it may be manually modified.
	ranks     map[string]int
	to_bcp_47 map[string]map[string]bool // A map of OpenType language system tags to sets of BCP 47 tags.
	// ``to_bcp_47`` inverted. Its values start as unsorted sets;
	// ``sortLanguages`` converts them to sorted lists.
	from_bcp_47 map[string]map[string]bool
}

func newOpenTypeRegistryParser() OpenTypeRegistryParser {
	var out OpenTypeRegistryParser
	out.names = make(map[string]string)
	out.ranks = make(map[string]int)
	out.to_bcp_47 = make(map[string]map[string]bool)
	out.from_bcp_47 = make(map[string]map[string]bool)
	return out
}

func (pr *OpenTypeRegistryParser) walkTree(n *html.Node) {
	switch n.DataAtom {
	case atom.Meta:
		for _, value := range n.Attr {
			if value.Key == "name" && value.Val == "updated_at" {
				var buf bytes.Buffer
				_ = html.Render(&buf, n)
				pr.header = buf.String()
				break
			}
		}
	// case atom.Td:
	// pr._td = true
	// pr._current_tr = append(pr._current_tr, "")
	case atom.Tr:
		// pr._current_tr = nil
		pr.handleTr(n)
		return // handleTr already do the recursion
	}

	// recursion
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		pr.walkTree(ch)
	}

}

func tdContent(n *html.Node) (string, bool) {
	if n.Type == html.TextNode {
		return n.Data, true
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if ct, ok := tdContent(ch); ok {
			return ct, true
		}
	}
	return "", false
}

// n is a <tr> element
func (pr *OpenTypeRegistryParser) handleTr(n *html.Node) {
	var currentTr []string
	for td := n.FirstChild; td != nil; td = td.NextSibling {
		if td.DataAtom == atom.Th {
			return // avoid header
		} else if td.DataAtom == atom.Td {
			ct, ok := tdContent(td)
			if !ok {
				return
			}
			currentTr = append(currentTr, ct)
		} else {
			continue
		}
	}

	expect(2 <= len(currentTr) && len(currentTr) <= 3, "invalid <tr> length")

	name := strings.TrimSpace(currentTr[0])
	tag := strings.Trim(currentTr[1], "\t\n\v\f\r '")
	rank := 0
	if len(tag) > 4 {
		expect(strings.HasSuffix(tag, " (deprecated)"), fmt.Sprintf("ill-formed OpenType tag: %s", tag))
		name += " (deprecated)"
		tag = strings.Split(tag, " ")[0]
		rank = 1
	}
	pr.names[tag] = strings.TrimSuffix(name, " languages")
	if len(currentTr) == 2 || currentTr[2] == "" {
		return
	}

	isoCodes := strings.TrimSpace(currentTr[2])
	s := pr.to_bcp_47[tag]
	if s == nil {
		s = make(map[string]bool)
	}
	for _, code := range strings.Split(strings.ReplaceAll(isoCodes, " ", ""), ",") {
		if c, ok := ISO_639_3_TO_1[code]; ok {
			code = c
		}
		s[code] = true
	}
	pr.to_bcp_47[tag] = s
	rank += 2 * len(pr.to_bcp_47[tag])
	pr.ranks[tag] = rank
}

// 	def handle_charref (self, name):
// 		self.handle_data (html_unescape (self, '&#%s;' % name))

// 	def handle_entityref (self, name):
// 		self.handle_data (html_unescape (self, '&%s;' % name))

// parse the OpenType language system tag registry.
func (pr *OpenTypeRegistryParser) parse(filename string) {
	data, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	root, err := html.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	pr.walkTree(root)

	expect(pr.header != "", "no header")

	for tag, iso_codes := range pr.to_bcp_47 {
		for iso_code := range iso_codes {
			s := pr.from_bcp_47[iso_code]
			if s == nil {
				s = make(map[string]bool)
			}
			s[tag] = true
			pr.from_bcp_47[iso_code] = s
		}
	}
}

// add a language as if it were in the registry.
// If `bcp47Tag` is more than just
// a language subtag, and if the language subtag is a
// macrolanguage, then new languages are added corresponding
// to the macrolanguages' individual languages with the
// remainder of the tag appended.
func (pr *OpenTypeRegistryParser) addLanguage(bcp47Tag, otTag string) {
	to, from := pr.to_bcp_47[otTag], pr.from_bcp_47[bcp47Tag]
	if to == nil {
		to = map[string]bool{}
	}
	if from == nil {
		from = make(map[string]bool)
	}
	to[bcp47Tag] = true
	from[otTag] = true
	pr.to_bcp_47[otTag] = to
	pr.from_bcp_47[bcp47Tag] = from

	if _, in := bcp47.grandfathered[strings.ToLower(bcp47Tag)]; !in {
		splitted := strings.SplitN(bcp47Tag, "-", 2)
		if len(splitted) != 2 {
			return
		}
		macrolanguage, suffix := splitted[0], splitted[1]
		if v, ok := bcp47.macrolanguages[macrolanguage]; ok {
			s := make(map[string]bool)
			for language := range v {
				if _, ok := bcp47.grandfathered[strings.ToLower(language)]; !ok {
					s[fmt.Sprintf("%s-%s", language, suffix)] = true
				}
			}
			bcp47.macrolanguages[fmt.Sprintf("%s-%s", macrolanguage, suffix)] = s
		}
	}
}

func _remove_language(tag_1 string, dict_1, dict_2 map[string]map[string]bool) {
	popped := dict_1[tag_1]
	delete(dict_1, tag_1)
	for tag_2 := range popped {
		delete(dict_2[tag_2], tag_1)
		if len(dict_2[tag_2]) == 0 {
			delete(dict_2, tag_2)
		}
	}
}

// Remove an OpenType tag from the registry.
func (pr OpenTypeRegistryParser) remove_language_ot(ot_tag string) {
	_remove_language(ot_tag, pr.to_bcp_47, pr.from_bcp_47)
}

// Remove a BCP 47 tag from the registry.
func (pr OpenTypeRegistryParser) remove_language_bcp_47(bcp47Tag string) {
	_remove_language(bcp47Tag, pr.from_bcp_47, pr.to_bcp_47)
}

// Copy mappings from macrolanguages to individual languages.
//
// If a BCP 47 tag for an individual mapping has no OpenType
// mapping but its macrolanguage does, the mapping is copied to
// the individual language. For example, als (Tosk Albanian) has no
// explicit mapping, so it inherits from sq (Albanian) the mapping
// to SQI.
//
// If a BCP 47 tag for a macrolanguage has no OpenType mapping but
// all of its individual languages do && they all map to the same
// tags, the mapping is copied to the macrolanguage.
func (pr *OpenTypeRegistryParser) inheritFromMacrolanguages() {
	originalOtFromBcp_47 := pr.from_bcp_47
	for macrolanguage, languages := range bcp47.macrolanguages {
		otMacrolanguages := make(map[string]bool)
		for k := range originalOtFromBcp_47[macrolanguage] {
			otMacrolanguages[k] = true
		}
		if len(otMacrolanguages) != 0 {
			for ot_macrolanguage := range otMacrolanguages {
				for language := range languages {
					pr.addLanguage(language, ot_macrolanguage)
					pr.ranks[ot_macrolanguage] += 1
				}
			}
		} else {
			for language := range languages {
				if _, in := originalOtFromBcp_47[language]; in {
					if len(otMacrolanguages) != 0 {
						ml := originalOtFromBcp_47[language]
						if len(ml) != 0 {
							otMacrolanguages = setIntersection(otMacrolanguages, ml)
						}
					} else {
						otMacrolanguages = setUnion(otMacrolanguages, originalOtFromBcp_47[language])
					}
				} else {
					otMacrolanguages = map[string]bool{}
				}
				if len(otMacrolanguages) == 0 {
					break
				}
			}
			for otMacrolanguage := range otMacrolanguages {
				pr.addLanguage(macrolanguage, otMacrolanguage)
			}
		}
	}
}

// sort the values of ``from_bcp_47`` in ascending rank order,
// and also return the sorted keys
func (pr OpenTypeRegistryParser) sortLanguages() (map[string][]string, []string) {
	out := make(map[string][]string)
	var keys []string
	sortKey := func(t, language string) int {
		return pr.ranks[t] + rankDelta(language, t)
	}
	for language, tags := range pr.from_bcp_47 {
		keys = append(keys, language)
		var ls []string
		for s := range tags {
			ls = append(ls, s)
		}
		sort.Strings(ls)
		sort.SliceStable(ls, func(i, j int) bool {
			return sortKey(ls[i], language) < sortKey(ls[j], language)
		})
		out[language] = ls
	}
	sort.Strings(keys)
	return out, keys
}

// a parser for the BCP 47 subtag registry.
type BCP47Parser struct {
	header string // The "File-Date" line of the registry.
	// A map of subtags to the names they are given in the registry. Each value is a
	// ``'\\n'``-separated list of names.
	names map[string]string
	// A map of language subtags to strings suffixed to language names,
	// including suffixes to explain language scopes.
	scopes map[string]string
	// A map of language subtags to the sets of language subtags which
	// inherit from them. See  ``OpenTypeRegistryParser.inheritFromMacrolanguages``.
	macrolanguages map[string]map[string]bool
	prefixes       map[string]map[string]bool //  A map of variant subtags to their prefixes.
	grandfathered  map[string]bool            // The set of grandfathered tags, normalized to lowercase.
}

func newBCP47Parser() BCP47Parser {
	var out BCP47Parser
	out.names = make(map[string]string)
	out.scopes = make(map[string]string)
	out.macrolanguages = make(map[string]map[string]bool)
	out.prefixes = make(map[string]map[string]bool)
	out.grandfathered = make(map[string]bool)
	return out
}

// Parse the BCP 47 subtag registry.
func (pr *BCP47Parser) parse(filename string) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	var subtag_type, subtag string
	deprecated := false
	has_preferred_value := false
	line_buffer := ""

	re := regexp.MustCompile(` (\(family\)|\((individual |macro)language\)|languages)$`)

	for _, lineB := range bytes.Split(b, []byte{'\n'}) {
		line := strings.TrimRightFunc(string(lineB), unicode.IsSpace)
		if strings.HasPrefix(line, " ") {
			line_buffer += line[1:]
			continue
		}
		line, line_buffer = line_buffer, line
		if strings.HasPrefix(line, "Type: ") {
			subtag_type = strings.Split(line, " ")[1]
			deprecated = false
			has_preferred_value = false
		} else if strings.HasPrefix(line, "Subtag: ") || strings.HasPrefix(line, "Tag: ") {
			subtag = strings.Split(line, " ")[1]
			if subtag_type == "grandfathered" {
				pr.grandfathered[strings.ToLower(subtag)] = true
			}
		} else if strings.HasPrefix(line, "Description: ") {
			description := strings.ReplaceAll(strings.SplitN(line, " ", 2)[1], " (individual language)", "")
			description = re.ReplaceAllString(description, "")
			if _, in := pr.names[subtag]; in {
				pr.names[subtag] += "\n" + description
			} else {
				pr.names[subtag] = description
			}
		} else if subtag_type == "language" || subtag_type == "grandfathered" {
			if strings.HasPrefix(line, "Scope: ") {
				scope := strings.Split(line, " ")[1]
				if scope == "macrolanguage" {
					scope = " [macrolanguage]"
				} else if scope == "collection" {
					scope = " [family]"
				} else {
					continue
				}
				pr.scopes[subtag] = scope
			} else if strings.HasPrefix(line, "Deprecated: ") {
				pr.scopes[subtag] = " (retired code)" + pr.scopes[subtag]
				deprecated = true
			} else if deprecated && strings.HasPrefix(line, "Comments: see ") {
				// If a subtag is split into multiple replacement subtags,
				// it essentially represents a macrolanguage.
				for _, language := range strings.Split(strings.ReplaceAll(line, ",", ""), " ")[2:] {
					pr._add_macrolanguage(subtag, language)
				}
			} else if strings.HasPrefix(line, "Preferred-Value: ") {
				// If a subtag is deprecated in favor of a single replacement subtag,
				// it is either a dialect || synonym of the preferred subtag. Either
				// way, it is close enough to the truth to consider the replacement
				// the macrolanguage of the deprecated language.
				has_preferred_value = true
				macrolanguage := strings.Split(line, " ")[1]
				pr._add_macrolanguage(macrolanguage, subtag)
			} else if !has_preferred_value && strings.HasPrefix(line, "Macrolanguage: ") {
				pr._add_macrolanguage(strings.Split(line, " ")[1], subtag)
			} else if subtag_type == "variant" {
			}
			if strings.HasPrefix(line, "Deprecated: ") {
				pr.scopes[subtag] = " (retired code)" + pr.scopes[subtag]
			} else if strings.HasPrefix(line, "Prefix: ") {
				pref := pr.prefixes[subtag]
				if pref == nil {
					pref = make(map[string]bool)
				}
				pref[strings.Split(line, " ")[1]] = true
				pr.prefixes[subtag] = pref
			}
		} else if strings.HasPrefix(line, "File-Date: ") {
			pr.header = line
		}
	}

	expect(pr.header != "", "no header")
}

func (pr *BCP47Parser) _add_macrolanguage(macrolanguage, language string) {
	if _, in := ot.from_bcp_47[language]; !in {
		for l := range pr.macrolanguages[language] {
			pr._add_macrolanguage(macrolanguage, l)
		}
	}
	if _, in := ot.from_bcp_47[macrolanguage]; !in {
		for _, ls := range pr.macrolanguages {
			if ls[macrolanguage] {
				ls[language] = true
				return
			}
		}
	}
	ml := pr.macrolanguages[macrolanguage]
	if ml == nil {
		ml = make(map[string]bool)
	}
	ml[language] = true
	pr.macrolanguages[macrolanguage] = ml
}

// make every language have at most one macrolanguage.
func (pr *BCP47Parser) removeExtraMacrolanguages() {
	inverted := make(map[string][]string)
	for macrolanguage, languages := range pr.macrolanguages {
		for language := range languages {
			inverted[language] = append(inverted[language], macrolanguage)
		}
	}
	for _, macrolanguages := range inverted {
		if len(macrolanguages) > 1 {
			sort.Slice(macrolanguages, func(i, j int) bool {
				return len(pr.macrolanguages[macrolanguages[i]]) < len(pr.macrolanguages[macrolanguages[j]])
			})
			biggestMacrolanguage := macrolanguages[len(macrolanguages)-1]
			macrolanguages = macrolanguages[:len(macrolanguages)-1]
			for _, macrolanguage := range macrolanguages {
				pr._add_macrolanguage(biggestMacrolanguage, macrolanguage)
			}
		}
	}
}

// def _get_name_piece (self, subtag):
// 	"""Return the first name of a subtag plus its scope suffix.

// 	Args:
// 		subtag (str): A BCP 47 subtag.

// 	Returns:
// 		The name form of ``subtag``.
// 	"""
// 	return self.names[subtag].split ('\n')[0] + self.scopes.get (subtag, '')

// def get_name (self, lt):
// 	"""Return the names of the subtags in a language tag.

// 	Args:
// 		lt (LanguageTag): A BCP 47 language tag.

// 	Returns:
// 		The name form of ``lt``.
// 	"""
// 	name = self._get_name_piece (lt.language)
// 	if lt.script:
// 		name += '; ' + self._get_name_piece (lt.script.title ())
// 	if lt.region:
// 		name += '; ' + self._get_name_piece (lt.region.upper ())
// 	if lt.variant:
// 		name += '; ' + self._get_name_piece (lt.variant)
// 	return name

func setEqual(s1, s2 map[string]bool) bool {
	if len(s1) != len(s2) {
		return false
	}
	for a := range s1 {
		if !s2[a] {
			return false
		}
	}
	return true
}

func set(as ...string) map[string]bool {
	out := make(map[string]bool)
	for _, a := range as {
		out[a] = true
	}
	return out
}

func setIntersection(s1, s2 map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for v := range s1 {
		if s2[v] {
			out[v] = true
		}
	}
	return out
}
func setUnion(s1, s2 map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for v := range s1 {
		out[v] = true
	}
	for v := range s2 {
		out[v] = true
	}
	return out
}

// download and save locally
func fectchData() {
	languagetags := "https://docs.microsoft.com/en-us/typography/opentype/spec/languagetags"
	resp, err := http.Get(languagetags)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	tags, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("languagetags.html", tags, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	languageSubtagRegistry := "https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry"
	resp, err = http.Get(languageSubtagRegistry)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	subtags, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("language-subtag-registry.txt", subtags, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func parse() {
	ot.parse("languagetags.html")

	bcp47.parse(string("language-subtag-registry.txt"))

	ot.addLanguage("ary", "MOR")

	ot.addLanguage("ath", "ATH")

	ot.addLanguage("bai", "BML")

	ot.ranks["BAL"] = ot.ranks["KAR"] + 1

	ot.addLanguage("ber", "BBR")

	ot.remove_language_ot("PGR")
	ot.addLanguage("el-polyton", "PGR")

	bcp47.macrolanguages["et"] = set("ekk")

	bcp47.names["flm"] = "Falam Chin"
	bcp47.scopes["flm"] = " (retired code)"
	bcp47.macrolanguages["flm"] = set("cfm")

	ot.ranks["FNE"] = ot.ranks["TNE"] + 1

	ot.addLanguage("und-fonipa", "IPPH")

	ot.addLanguage("und-fonnapa", "APPH")

	ot.remove_language_ot("IRT")
	ot.addLanguage("ga-Latg", "IRT")

	ot.addLanguage("hy-arevmda", "HYE")

	ot.remove_language_ot("KGE")
	ot.addLanguage("und-Geok", "KGE")

	bcp47.macrolanguages["id"] = set("in")

	bcp47.macrolanguages["ijo"] = set("ijc")

	ot.addLanguage("kht", "KHN")
	ot.names["KHN"] = ot.names["KHT"] + " (Microsoft fonts)"
	ot.ranks["KHN"] = ot.ranks["KHT"] + 1

	ot.ranks["LCR"] = ot.ranks["MCR"] + 1

	ot.names["MAL"] = "Malayalam Traditional"
	ot.ranks["MLR"] += 1

	bcp47.names["mhv"] = "Arakanese"
	bcp47.scopes["mhv"] = " (retired code)"

	ot.addLanguage("no", "NOR")

	ot.addLanguage("oc-provenc", "PRO")

	ot.addLanguage("qu", "QUZ")
	ot.addLanguage("qub", "QWH")
	ot.addLanguage("qud", "QVI")
	ot.addLanguage("qug", "QVI")
	ot.addLanguage("qul", "QUH")
	ot.addLanguage("qup", "QVI")
	ot.addLanguage("qur", "QWH")
	ot.addLanguage("qus", "QUH")
	ot.addLanguage("quw", "QVI")
	ot.addLanguage("qux", "QWH")
	ot.addLanguage("qva", "QWH")
	ot.addLanguage("qvh", "QWH")
	ot.addLanguage("qvj", "QVI")
	ot.addLanguage("qvl", "QWH")
	ot.addLanguage("qvm", "QWH")
	ot.addLanguage("qvn", "QWH")
	ot.addLanguage("qvo", "QVI")
	ot.addLanguage("qvp", "QWH")
	ot.addLanguage("qvw", "QWH")
	ot.addLanguage("qvz", "QVI")
	ot.addLanguage("qwa", "QWH")
	ot.addLanguage("qws", "QWH")
	ot.addLanguage("qxa", "QWH")
	ot.addLanguage("qxc", "QWH")
	ot.addLanguage("qxh", "QWH")
	ot.addLanguage("qxl", "QVI")
	ot.addLanguage("qxn", "QWH")
	ot.addLanguage("qxo", "QWH")
	ot.addLanguage("qxr", "QVI")
	ot.addLanguage("qxt", "QWH")
	ot.addLanguage("qxw", "QWH")

	delete(bcp47.macrolanguages["ro"], "mo")
	s := bcp47.macrolanguages["ro-MD"]
	if s == nil {
		s = make(map[string]bool)
	}
	s["mo"] = true
	bcp47.macrolanguages["ro-MD"] = s

	ot.remove_language_ot("SYRE")
	ot.remove_language_ot("SYRJ")
	ot.remove_language_ot("SYRN")
	ot.addLanguage("und-Syre", "SYRE")
	ot.addLanguage("und-Syrj", "SYRJ")
	ot.addLanguage("und-Syrn", "SYRN")

	bcp47.names["xst"] = "Silt'e"
	bcp47.scopes["xst"] = " (retired code)"
	bcp47.macrolanguages["xst"] = set("stv", "wle")

	ot.addLanguage("xwo", "TOD")

	ot.remove_language_ot("ZHH")
	ot.remove_language_ot("ZHP")
	ot.remove_language_ot("ZHT")
	ot.remove_language_ot("ZHTM")
	delete(bcp47.macrolanguages["zh"], "lzh")
	delete(bcp47.macrolanguages["zh"], "yue")
	ot.addLanguage("zh-Hant-MO", "ZHH")
	ot.addLanguage("zh-Hant-MO", "ZHTM")
	ot.addLanguage("zh-Hant-HK", "ZHH")
	ot.addLanguage("zh-Hans", "ZHS")
	ot.addLanguage("zh-Hant", "ZHT")
	ot.addLanguage("zh-HK", "ZHH")
	ot.addLanguage("zh-MO", "ZHH")
	ot.addLanguage("zh-MO", "ZHTM")
	ot.addLanguage("zh-TW", "ZHT")
	ot.addLanguage("lzh", "ZHT")
	ot.addLanguage("lzh-Hans", "ZHS")
	ot.addLanguage("yue", "ZHH")
	ot.addLanguage("yue-Hans", "ZHS")

	bcp47.macrolanguages["zom"] = set("yos")

}

// Return a delta to apply to a BCP 47 tag's rank.
//
// Most OpenType tags have a constant rank, but a few have ranks that
// depend on the BCP 47 tag.
func rankDelta(bcp47, ot string) int {
	if bcp47 == "ak" && ot == "AKA" {
		return -1
	}
	if bcp47 == "tw" && ot == "TWI" {
		return -1
	}
	return 0
}

var disambiguation = map[string]string{
	"ALT":  "alt",
	"ARK":  "rki",
	"ATH":  "ath",
	"BHI":  "bhb",
	"BLN":  "bjt",
	"BTI":  "beb",
	"CCHN": "cco",
	"CMR":  "swb",
	"CPP":  "crp",
	"CRR":  "crx",
	"DUJ":  "dwu",
	"ECR":  "crj",
	"HAL":  "cfm",
	"HND":  "hnd",
	"HYE":  "hyw",
	"KIS":  "kqs",
	"KUI":  "uki",
	"LRC":  "bqi",
	"NDB":  "nd",
	"NIS":  "njz",
	"PLG":  "pce",
	"PRO":  "pro",
	"QIN":  "bgr",
	"QUH":  "quh",
	"QVI":  "qvi",
	"QWH":  "qwh",
	"SIG":  "stv",
	"SRB":  "sr",
	"ZHH":  "zh-HK",
	"ZHS":  "zh-Hans",
	"ZHT":  "zh-Hant",
	"ZHTM": "zh-MO",
}

func max(vs map[string]int) int {
	m := -math.MinInt32
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}

// convert a tag to ``newTag`` form.
func hbTag(tag string) string {
	if tag == DEFAULT_LANGUAGE_SYSTEM {
		return "0\t"
	}
	tag += "    " // pad with spaces
	return fmt.Sprintf("newTag('%s','%s','%s','%s')", string(tag[0]), string(tag[1]), string(tag[2]), string(tag[3]))
}

// return a set of variant language names from a name, joined on '\\n'.
func getVariantSet(name string) map[string]bool {
	variants := strings.FieldsFunc(name, func(r rune) bool {
		switch r {
		case '\n', '(', ')', ',':
			return true
		}
		return false
	})

	out := make(map[string]bool)
	for _, n := range variants {
		if n == "" {
			continue
		}
		n = strings.ReplaceAll(n, string('\u2019'), "'")
		var ascii []byte
		for _, b := range norm.NFD.String(n) {
			if b <= 127 {
				ascii = append(ascii, byte(b))
			}
		}
		out[strings.TrimSpace(string(ascii))] = true
	}
	return out
}

// return the names in common between two language names, joined on '\\n'.
func languageNameIntersection(a, b string) map[string]bool {
	return setIntersection(getVariantSet(a), getVariantSet(b))
}

func getMatchingLanguageName(intersection map[string]bool, candidates []string) string {
	for _, c := range candidates {
		if len(setIntersection(intersection, getVariantSet(c))) != 0 { // not disjoint
			return c
		}
	}
	return ""
}

func sameTag(bcp47Tag string, otTags []string) bool {
	return len(bcp47Tag) == 3 && len(otTags) == 1 && bcp47Tag == strings.ToLower(otTags[0])
}

func printSubtagMatches(w io.Writer, subtag, new_line string) {
	if subtag != "" {
		if new_line != "" {
			fmt.Fprintln(w)
			fmt.Fprint(w, "\t&& ")
		}
		fmt.Fprintf(w, "subtag_matches (lang_str, limit, %q)", subtag)
	}
}

func main() {
	// uncomment once to download and save locally
	// fectchData()

	parse()

	ot.inheritFromMacrolanguages()
	bcp47.removeExtraMacrolanguages()
	ot.inheritFromMacrolanguages()
	ot.names[DEFAULT_LANGUAGE_SYSTEM] = "*/"
	ot.ranks[DEFAULT_LANGUAGE_SYSTEM] = max(ot.ranks) + 1

	re := regexp.MustCompile("[A-Z]{3}$")
	for tag := range ot.names {
		if !re.MatchString(tag) {
			continue
		}
		possible_bcp_47_tag := strings.ToLower(tag)
		if _, in := bcp47.names[possible_bcp_47_tag]; in && len(ot.from_bcp_47[possible_bcp_47_tag]) == 0 {
			ot.addLanguage(possible_bcp_47_tag, DEFAULT_LANGUAGE_SYSTEM)
			bcp47.macrolanguages[possible_bcp_47_tag] = set()
		}
	}

	langs, keys := ot.sortLanguages()

	w := os.Stdout

	fmt.Fprintln(w, "package harfbuzz")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "// Code generated by langs/gen.go. DO NOT EDIT")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "var ot_languages =[...]LangTag{")

	for _, language := range keys {
		tags := langs[language]
		if language == "" || strings.IndexByte(language, '-') != -1 {
			continue
		}
		commentedOut := sameTag(language, tags)
		for _, tag := range tags {
			if commentedOut {
				fmt.Fprintf(w, "%s{%q,\t%s},", "/*", language, hbTag(tag))
			} else {
				fmt.Fprintf(w, "%s{%q,\t%s},", "  ", language, hbTag(tag))
			}
			if commentedOut {
				fmt.Fprint(w, "*/")
			}
			fmt.Fprint(w, "\t/* ")
			bcp_47Name := bcp47.names[language]
			bcp_47NameCandidates := strings.Split(bcp_47Name, "\n")
			ot_name := ot.names[tag]
			scope := bcp47.scopes[language]
			if tag == DEFAULT_LANGUAGE_SYSTEM {
				fmt.Fprintf(w, "%s%s != %s", bcp_47NameCandidates[0], scope, ot.names[strings.ToUpper(language)])
			} else {
				intersection := languageNameIntersection(bcp_47Name, ot_name)
				if len(intersection) == 0 {
					fmt.Fprintf(w, "%s%s -> %s", bcp_47NameCandidates[0], scope, ot_name)
				} else {
					name := getMatchingLanguageName(intersection, bcp_47NameCandidates)
					bcp47.names[language] = name
					if len(name) > len(ot_name) {
						fmt.Fprintf(w, "%s%s", name, scope)
					} else {
						fmt.Fprintf(w, "%s%s", ot_name, scope)
					}
				}
			}
			fmt.Fprintln(w, " */")
		}
	}

	fmt.Fprintln(w, "}")
	fmt.Fprintln(w)
}

func printComplexFunc(w io.Writer) {
	fmt.Fprintln(w, "// hb_ot_tags_from_complex_language:")
	fmt.Fprintln(w, "// @lang_str: a BCP 47 language tag to convert.")
	fmt.Fprintln(w, "// @limit: a pointer to the end of the substring of @lang_str to consider for")
	fmt.Fprintln(w, "// conversion.")
	fmt.Fprintln(w, "// @count: maximum number of language tags to retrieve (IN) && actual number of")
	fmt.Fprintln(w, "// language tags retrieved (OUT). If no tags are retrieved, it is not modified.")
	fmt.Fprintln(w, "// @tags: array of size at least @language_count to store the language tag")
	fmt.Fprintln(w, "// results")
	fmt.Fprintln(w, "//")
	fmt.Fprintln(w, "// Converts a multi-subtag BCP 47 language tag to language tags.")
	fmt.Fprintln(w, "//")
	fmt.Fprintln(w, "// Return value: Whether any language systems were retrieved.")
	fmt.Fprintln(w, "func hb_ot_tags_from_complex_language bool(lang_str string, limit int, count *int) []hb_tag_t {")

	// complex_tags = collections.defaultdict (list)
	// for initial, group in itertools.groupby ((lt_tags for lt_tags in [
	// 			(LanguageTag (language), tags)
	// 			for language, tags in sorted (ot.from_bcp_47.items (),
	// 				key=lambda i: (-len (i[0]), i[0]))
	// 		] if lt_tags[0].isComplex ()),
	// 		key=lambda lt_tags: lt_tags[0].getGroup ()):
	// 	complex_tags[initial] += group

	// for initial, items in sorted (complex_tags.items ()):
	// 	if initial != 'und':
	// 		continue
	// 	for lt, tags in items:
	// 		if lt.variant in bcp47.prefixes:
	// 			expect (next (iter (bcp47.prefixes[lt.variant])) == lt.language,
	// 					"%" i" not"a valid prefix of %s' % (lt.language, lt.variant))
	// 		fmt.Fprintf (w,'  if (', end='')
	// 		printSubtagMatches (lt.script, false)
	// 		printSubtagMatches (lt.region, false)
	// 		printSubtagMatches (lt.variant, false)
	// 		fmt.Fprintf (w,')')
	// 		fmt.Fprintf (w,'  {')
	// 		write ('    /* %s */' % bcp47.get_name (lt))
	// 		fmt.Fprintf (w,)
	// 		if len (tags) == 1:
	// 			write ('    tags[0] = %s;  /* %s */' % (hbTag (tags[0]), ot.names[tags[0]]))
	// 			fmt.Fprintf (w,)
	// 			fmt.Fprintf (w,'    *count = 1;')
	// 		else:
	// 			fmt.Fprintf (w,'    hb_tag_t possible_tags[] = {')
	// 			for tag in tags:
	// 				write ('      %s,  /* %s */' % (hbTag (tag), ot.names[tag]))
	// 				fmt.Fprintf (w,)
	// 			fmt.Fprintf (w,'    };')
	// 			fmt.Fprintf (w,'    for (i = 0; i < %s && i < *count; i++)' % len (tags))
	// 			fmt.Fprintf (w,'      tags[i] = possible_tags[i];')
	// 			fmt.Fprintf (w,'    *count = i;')
	// 		fmt.Fprintf (w,'    return true;')
	// 		fmt.Fprintf (w,'  }')

	// fmt.Fprintf (w,'  switch (lang_str[0])')
	// fmt.Fprintf (w,'  {')
	// for initial, items in sorted (complex_tags.items ()):
	// 	if initial == 'und':
	// 		continue
	// 	fmt.Fprintf (w,"  case '%s':" % initial)
	// 	for lt, tags in items:
	// 		fmt.Fprintf (w,'    if (', end='')
	// 		script = lt.script
	// 		region = lt.region
	// 		if lt.grandfathered:
	// 			fmt.Fprintf (w,'0 == strcmp (&lang_str[1], "%s")' % lt.language[1:], end='')
	// 		else:
	// 			string_literal = lt.language[1:] + '-'
	// 			if script:
	// 				string_literal += script
	// 				script = None
	// 				if region:
	// 					string_literal += '-' + region
	// 					region = None
	// 			if string_literal[-1] == '-':
	// 				fmt.Fprintf (w,'0 == strncmp (&lang_str[1], "%s", %i)' % (string_literal, len (string_literal)), end='')
	// 			else:
	// 				fmt.Fprintf (w,'lang_matches (&lang_str[1], "%s")' % string_literal, end='')
	// 		printSubtagMatches (script, true)
	// 		printSubtagMatches (region, true)
	// 		printSubtagMatches (lt.variant, true)
	// 		fmt.Fprintf (w,')')
	// 		fmt.Fprintf (w,'    {')
	// 		write ('      /* %s */' % bcp47.get_name (lt))
	// 		fmt.Fprintf (w,)
	// 		if len (tags) == 1:
	// 			write ('      tags[0] = %s;  /* %s */' % (hbTag (tags[0]), ot.names[tags[0]]))
	// 			fmt.Fprintf (w,)
	// 			fmt.Fprintf (w,'      *count = 1;')
	// 		else:
	// 			fmt.Fprintf (w,'      unsigned int i;')
	// 			fmt.Fprintf (w,'      hb_tag_t possible_tags[] = {')
	// 			for tag in tags:
	// 				write ('\t%s,  /* %s */' % (hbTag (tag), ot.names[tag]))
	// 				fmt.Fprintf (w,)
	// 			fmt.Fprintf (w,'      };')
	// 			fmt.Fprintf (w,'      for (i = 0; i < %s && i < *count; i++)' % len (tags))
	// 			fmt.Fprintf (w,'\ttags[i] = possible_tags[i];')
	// 			fmt.Fprintf (w,'      *count = i;')
	// 		fmt.Fprintf (w,'      return true;')
	// 		fmt.Fprintf (w,'    }')
	// 	fmt.Fprintf (w,'    break;')

	// fmt.Fprintf (w,'  }')
	// fmt.Fprintf (w,'  return false;')
	// fmt.Fprintf (w,'}')
	// fmt.Fprintf (w,)
	// fmt.Fprintf (w,'/**')
	// fmt.Fprintf (w,' * hb_ot_ambiguous_tag_to_language')
	// fmt.Fprintf (w,' * @tag: A language tag.')
	// fmt.Fprintf (w,' *')
	// fmt.Fprintf (w,' * Converts @tag to a BCP 47 language tag if it is ambiguous (it corresponds to')
	// fmt.Fprintf (w,' * many language tags) && the best tag is not the alphabetically first, || if')
	// fmt.Fprintf (w,' * the best tag consists of multiple subtags, || if the best tag does not appear')
	// fmt.Fprintf (w,' * in #ot_languages.')
	// fmt.Fprintf (w,' *')
	// fmt.Fprintf (w,' * Return value: The #hb_language_t corresponding to the BCP 47 language tag,')
	// fmt.Fprintf (w,' * || #HB_LANGUAGE_INVALID if @tag is not ambiguous.')
	// fmt.Fprintf (w,' **/')
	// fmt.Fprintf (w,'static hb_language_t')
	// fmt.Fprintf (w,'hb_ot_ambiguous_tag_to_language (hb_tag_t tag)')
	// fmt.Fprintf (w,'{')
	// fmt.Fprintf (w,'  switch (tag)')
	// fmt.Fprintf (w,'  {')

	// verify_disambiguation_dict ()
	// for ot_tag, bcp47Tag in sorted (disambiguation.items ()):
	// 	write ('  case %s:  /* %s */' % (hbTag (ot_tag), ot.names[ot_tag]))
	// 	fmt.Fprintf (w,)
	// 	write ('    return hb_language_from_string (\"%s\", -1);  /* %s */' % (bcp47Tag, bcp47.get_name (LanguageTag (bcp47Tag))))
	// 	fmt.Fprintf (w,)

	// fmt.Fprintf (w,'  default:')
	// fmt.Fprintf (w,'    return HB_LANGUAGE_INVALID;')
	// fmt.Fprintf (w,'  }')
	// fmt.Fprintf (w,'}')

	// fmt.Fprintf (w,)
	// fmt.Fprintf (w,'#endif /* HB_OT_TAG_TABLE_HH */')
	// fmt.Fprintf (w,)
	// fmt.Fprintf (w,'/* == End of generated table == */')

}

// def verify_disambiguation_dict ():
// 	"""Verify && normalize ``disambiguation``.

// 	``disambiguation`` is a map of ambiguous OpenType language system
// 	tags to the particular BCP 47 tags they correspond to. This function
// 	checks that all its keys really are ambiguous && that each key's
// 	value is valid for that key. It checks that no ambiguous tag is
// 	missing, except when it can figure out which BCP 47 tag is the best
// 	by itself.

// 	It modifies ``disambiguation`` to remove keys whose values are the
// 	same as those that the fallback would return anyway, && to add
// 	ambiguous keys whose disambiguations it determined automatically.

// 	Raises:
// 		AssertionError: Verification failed.
// 	"""
// 	global bcp47
// 	global disambiguation
// 	global ot
// 	for ot_tag, bcp_47_tags in ot.to_bcp_47.items ():
// 		if ot_tag == DEFAULT_LANGUAGE_SYSTEM:
// 			primary_tags = []
// 		else:
// 			primary_tags = list (t for t in bcp_47_tags if t not in bcp47.grandfathered && ot.from_bcp_47.get (t)[0] == ot_tag)
// 		if len (primary_tags) == 1:
// 			expect (ot_tag not in disambiguation, 'unnecessary disambiguation for OT tag: %s' % ot_tag)
// 			if '-' in primary_tags[0]:
// 				disambiguation[ot_tag] = primary_tags[0]
// 			else:
// 				first_tag = sorted (t for t in bcp_47_tags if t not in bcp47.grandfathered && ot_tag in ot.from_bcp_47.get (t))[0]
// 				if primary_tags[0] != first_tag:
// 					disambiguation[ot_tag] = primary_tags[0]
// 		else if len (primary_tags) == 0:
// 			expect (ot_tag not in disambiguation, 'There is no possible valid disambiguation for %s' % ot_tag)
// 		else:
// 			macrolanguages = list (t for t in primary_tags if bcp47.scopes.get (t) == ' [macrolanguage]')
// 			if len (macrolanguages) != 1:
// 				macrolanguages = list (t for t in primary_tags if bcp47.scopes.get (t) == ' [family]')
// 			if len (macrolanguages) != 1:
// 				macrolanguages = list (t for t in primary_tags if 'retired code' not in bcp47.scopes.get (t, ''))
// 			if len (macrolanguages) != 1:
// 				expect (ot_tag in disambiguation, 'ambiguous OT tag: %s %s' % (ot_tag, str (macrolanguages)))
// 				expect (disambiguation[ot_tag] in bcp_47_tags,
// 						"%" i" not"a valid disambiguation for %s' % (disambiguation[ot_tag], ot_tag))
// 			else if ot_tag not in disambiguation:
// 				disambiguation[ot_tag] = macrolanguages[0]
// 			different_bcp_47_tags = sorted (t for t in bcp_47_tags if not sameTag (t, ot.from_bcp_47.get (t)))
// 			if different_bcp_47_tags && disambiguation[ot_tag] == different_bcp_47_tags[0] && '-' not in disambiguation[ot_tag]:
// 				del disambiguation[ot_tag]
// 	for ot_tag in disambiguation.keys ():
// 		expect (ot_tag in ot.to_bcp_47, 'unknown OT tag: %s' % ot_tag)
