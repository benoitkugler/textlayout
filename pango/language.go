package pango

import (
	"os"
	"strings"
	"sync"

	"github.com/benoitkugler/textlayout/language"
)

type Language = language.Language

//  /* We embed a private struct right *before* a where a PangoLanguage *
//   * points to.
//   */

//  typedef struct {
//    gconstpointer lang_info;
//    gconstpointer script_for_lang;

//    int magic; /* Used for verification */
//  } PangoLanguagePrivate;

// //  #define PANGO_LANGUAGE_PRIVATE_MAGIC 0x0BE4DAD0

//  static void
//  pango_language_private_init (PangoLanguagePrivate *priv)
//  {
//    priv.magic = PANGO_LANGUAGE_PRIVATE_MAGIC;

//    priv.lang_info = (gconstpointer) -1;
//    priv.script_for_lang = (gconstpointer) -1;
//  }

//  static PangoLanguagePrivate *
//  pango_language_get_private (PangoLanguage *language)
//  {
//    PangoLanguagePrivate *priv;

//    if (!language)
// 	 return nil;

//    priv = (PangoLanguagePrivate *)(void *)((char *)language - sizeof (PangoLanguagePrivate));

//    if (G_UNLIKELY (priv.magic != PANGO_LANGUAGE_PRIVATE_MAGIC))
// 	 {
// 	   g_critical ("Invalid PangoLanguage.  Did you pass in a straight string instead of calling pango_language_from_string()?");
// 	   return nil;
// 	 }

//    return priv;
//  }

const languageSeparators = ";:, \t"

// var canonMap = [256]byte{
// 	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// 	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '-', 0, 0,
// 	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 0, 0, 0, 0, 0, 0,
// 	'-', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
// 	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 0, 0, 0, 0, '-',
// 	0, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
// 	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 0, 0, 0, 0, 0,
// }

// // canonicalize the language input
// func canonicalize(v []byte) []byte {
// 	out := make([]byte, 0, len(v))
// 	for _, b := range v {
// 		can := canonMap[b]
// 		if can != 0 {
// 			out = append(out, can)
// 		}
// 	}
// 	return out
// }

// func lang_equal (v1,  v2 []byte) bool {
//  i := 0
//  _ = v2[len(v1)-1] // hint for BCE
//    for ; i < len(v1); i ++ {
// 	   p1, p2 := v1[i], v2[i]
// 		if !(canonMap[p1] !=0  && canonMap[p1] == canonMap[p2]) {
// 			break
// 		}
// 	 }

//    return (canonMap[v1[i]] == canonMap[v2[i]]);
//  }

// func lang_hash(key []byte) uint32 {
// 	var h uint32
// 	for _, p := range key {
// 		if canonMap[p] == 0 {
// 			break
// 		}
// 		h = (h << 5) - h + uint32(canonMap[p])
// 	}

// 	return h
// }

var (
	defaultLanguage     Language
	defaultLanguageOnce sync.Once
)

// DefaultLanguage calls language.DefaultLanguage.
func DefaultLanguage() Language {
	return language.DefaultLanguage()
	// defaultLanguageOnce.Do(func() {
	// 	defaultLanguage = language.DefaultLanguage()
	// })

	// return defaultLanguage
}

// var (
// 	languageHashTable     = map[uint32]Language{}
// 	languageHashTableLock = sync.Mutex{}
// )

// pango_language_from_string takes a RFC-3066 format language tag as a string and convert it to a
// Language pointer that can be efficiently copied (copy the
// pointer) and compared with other language tags (compare the
// pointer.)
//
// This function first canonicalizes the string by converting it to
// lowercase, mapping '_' to '-', and stripping all characters other
// than letters and '-'.
//
// Use DefaultLanguage() if you want to get the Language for
// the current locale of the process.
func pango_language_from_string(l string) Language {
	// Note: for now, we use a very simple implementation:
	// we just store the canonical form
	// The C implementation has a more refined algorithm,
	// using a map from language to pointer,
	// which replace comparison and copies of string to comparison and copies of pointers

	return language.NewLanguage(l)

	// hash := lang_hash(can)

	// result, ok := languageHashTable[hash]
	// if !ok { // insert a new language
	// 	languageHashTableLock.Lock()
	// 	defer languageHashTableLock.Unlock()
	// 	languageHashTable[hash] = Language(hash)
	// }
}

func compute_derived_language(lang Language, script Script) Language {
	var derivedLang Language

	/* Make sure the language tag is consistent with the derived
	 * script. There is no point in marking up a section of
	 * Arabic text with the "en" language tag.
	 */
	if lang != "" && pango_language_includes_script(lang, script) {
		derivedLang = lang
	} else {
		derivedLang = pango_script_get_sample_language(script)
		/* If we don't find a sample language for the script, we
		 * use a language tag that shouldn't actually be used
		 * anywhere. This keeps fontconfig (for the PangoFc*
		 * backend) from using the language tag to affect the
		 * sort order. I don't have a reference for 'xx' being
		 * safe here, though Keith Packard claims it is.
		 */
		if derivedLang == "" {
			derivedLang = "xx"
		}
	}

	return derivedLang
}

/**
 * `language`: (nullable): a language tag (see pango_language_from_string()),
 *            `nil` is allowed and matches nothing but '*'
 * @rangeList: a list of language ranges, separated by ';', ':',
 *   ',', or space characters.
 *   Each element must either be '*', or a RFC 3066 language range
 *   canonicalized as by pango_language_from_string()
 *
 * pangoLanguageMatches checks if a language tag matches one of the elements in a list of
 * language ranges. A language tag is considered to match a range
 * in the list if the range is '*', the range is exactly the tag,
 * or the range is a prefix of the tag, and the character after it in the tag is '-'.
 **/
func pangoLanguageMatches(lang_ Language, rangeList string) bool {
	langRs := strings.FieldsFunc(rangeList, func(r rune) bool {
		switch r {
		case ';', ':', ',', ' ', '\t':
			return true
		default:
			return false
		}
	})

	lang := string(lang_)
	for _, langR := range langRs {
		end := len(langR)
		if end >= len(lang) { // truncate end if needed
			end = len(lang) - 1
		}

		if langR[0] == '*' || (lang != "" && strings.HasPrefix(lang, langR) &&
			(len(lang) == len(langR) || lang[end] == '-')) {
			return true
		}
	}

	return false
}

func langCompareFirstComponent(ra, rb Language) int {
	a, b := string(ra), string(rb)
	p := strings.Index(a, "-")
	da := len(a)
	if p != -1 {
		da = p
	}
	p = strings.Index(b, "-")
	db := len(b)
	if p != -1 {
		db = p
	}
	total := max(da, db)
	if total <= len(a) {
		a = a[:total]
	}
	if total <= len(b) {
		b = b[:total]
	}
	return strings.Compare(a, b)
}

type recordScript struct {
	lang    Language
	scripts [3]Script // 3 is the maximum number of possible script for a language
}

type recordSample struct {
	lang   Language
	sample string
}

type languageRecord interface {
	language() Language
}

func (r recordScript) language() Language { return r.lang }
func (r recordSample) language() Language { return r.lang }

// return the index into `base`
func binarySearch(key Language, base []languageRecord) (int, bool) {
	min, max := 0, len(base)-1
	for min <= max {
		mid := (min + max) / 2
		p := base[mid]
		c := langCompareFirstComponent(key, p.language())
		if c < 0 {
			max = mid - 1
		} else if c > 0 {
			min = mid + 1
		} else {
			return mid, true
		}
	}
	return 0, false
}

// Finds the best record for `language` in an array of record,
// which must be sorted on language code.
func findBestLangMatch(language Language, records []languageRecord) languageRecord {
	r, ok := binarySearch(language, records)
	if !ok {
		return nil
	}

	/* find the best match among all those that have the same first-component */

	/* go to the final one matching in the first component */
	for r+1 < len(records) && langCompareFirstComponent(language, records[r+1].language()) == 0 {
		r += 1
	}

	/* go back, find which one matches completely */
	for 0 <= r && langCompareFirstComponent(language, records[r].language()) == 0 {
		if pangoLanguageMatches(language, string(records[r].language())) {
			return records[r]
		}
		r -= 1
	}

	return nil
}

func findBestLangMatchCached(language Language, records []languageRecord) languageRecord {
	// TODO: add caching
	return findBestLangMatch(language, records)
}

// SampleString get a string that is representative of the characters needed to
// render a particular language.
//
// The sample text may be a pangram, but is not necessarily.  It is chosen to
// be demonstrative of normal text in the language, as well as exposing font
// feature requirements unique to the language.  It is suitable for use
// as sample text in a font selection dialog.
//
// If `language` is empty, the default language as found by
// DefaultLanguage() is used.
//
// If Pango does not have a sample string for `language`, the classic
// "The quick brown fox..." is returned.
func SampleString(lang Language) string {
	if lang == "" {
		lang = DefaultLanguage()
	}

	lang_info := findBestLangMatchCached(lang, langTexts)

	if lang_info != nil {
		return lang_info.(recordSample).sample
	}

	return "The quick brown fox jumps over the lazy dog."
}

/*
 * From language to script
 */

// pango_language_get_scripts determines the scripts used to to write `language`.
// If nothing is known about the language tag `language`,
// or if `language` is empty, then `nil` is returned.
// The list of scripts returned starts with the script that the
// language uses most and continues to the one it uses least.
//
// Most languages use only one script for writing, but there are
// some that use two (Latin and Cyrillic for example), and a few
// use three (Japanese for example). Applications should not make
// any assumptions on the maximum number of scripts returned
// though, except that it is positive if the return value is not
// `nil`, and it is a small number.
//
// The pango_language_includes_script() function uses this function
// internally.
func pango_language_get_scripts(lang Language) []Script {
	scriptRec := findBestLangMatchCached(lang, pango_script_for_lang)

	if scriptRec == nil {
		return nil
	}
	script_for_lang := scriptRec.(recordScript)
	if script_for_lang.scripts[0] == 0 {
		return nil
	}
	for j, s := range script_for_lang.scripts {
		if s == 0 {
			return script_for_lang.scripts[:j]
		}
	}

	return nil
}

// pango_language_includes_script determines if `script` is one of the scripts used to
// write `language`. The returned value is conservative;
// if nothing is known about the language tag `language`,
// `true` will be returned, since, as far as Pango knows,
// `script` might be used to write `language`.
//
// This routine is used in Pango's itemization process when
// determining if a supplied language tag is relevant to
// a particular section of text. It probably is not useful for
// applications in most circumstances.
//
// This function uses pango_language_get_scripts() internally.
//
// Return value: `true` if `script` is one of the scripts used
// to write `language` or if nothing is known about `language`
// (including the case that `language` is `nil`),
// `false` otherwise.
func pango_language_includes_script(lang Language, script Script) bool {
	if !script.IsRealScript() {
		return true
	}

	scripts := pango_language_get_scripts(lang)
	if len(scripts) == 0 {
		return true
	}

	for _, s := range scripts {
		if s == script {
			return true
		}
	}

	return false
}

/*
 * From script to language
 */

func parse_default_languages() []Language {
	p, ok := os.LookupEnv("PANGO_LANGUAGE")

	if !ok {
		p, ok = os.LookupEnv("LANGUAGE")
	}

	if !ok {
		return nil
	}

	var langs []Language

	done := false
	for !done {
		end := strings.IndexAny(p, languageSeparators)
		if end == -1 {
			end = len(p)
			done = true
		} else {
			p = p[:end]
		}

		/* skip empty languages, and skip the language 'C' */
		if end != 0 && !(end == 1 && p[0] == 'C') {
			l := pango_language_from_string(p)
			langs = append(langs, l)
		}

		if !done {
			p = p[end:]
		}
	}

	return langs
}

var (
	languagesLock sync.Mutex
	languages     []Language          /* MT-safe */
	hash          map[Script]Language /* MT-safe */
)

func pango_script_get_default_language(script Script) Language {
	languagesLock.Lock()
	defer languagesLock.Unlock()

	if hash == nil { // initialize
		languages = parse_default_languages()
		hash = make(map[Script]Language)
	}

	if len(languages) == 0 {
		return ""
	}

	result, ok := hash[script]
	if ok {
		return result
	}

	for _, p := range languages {
		if pango_language_includes_script(p, script) {
			result = p
			break
		}
	}

	hash[script] = result

	return result
}

//  /**
//   * pango_language_get_preferred:
//   *
//   * Returns the list of languages that the user prefers, as specified
//   * by the PANGO_LANGUAGE or LANGUAGE environment variables, in order
//   * of preference. Note that this list does not necessarily include
//   * the language returned by DefaultLanguage().
//   *
//   * When choosing language-specific resources, such as the sample
//   * text returned by GetSampleString(), you should
//   * first try the default language, followed by the languages returned
//   * by this function.
//   *
//   * Returns: (transfer none) (nullable): a `nil`-terminated array of
//   *    PangoLanguage*
//   *
//   * Since: 1.48
//   */
//  PangoLanguage **
//  pango_language_get_preferred (void)
//  {
//    /* We call this just for its side-effect of initializing languages */
//    _pango_script_get_default_language (PANGO_SCRIPT_COMMON);

//    return languages;
//  }

// pango_script_get_sample_language finds a language tag that is reasonably
// representative of that script. This will usually be the
// most widely spoken or used language written in that script:
// for instance, the sample language for `SCRIPT_CYRILLIC`
// is 'ru' (Russian), the sample language for `SCRIPT_ARABIC` is 'ar'.
//
// For some scripts, no sample language will be returned because there
// is no language that is sufficiently representative. The best
// example of this is `SCRIPT_HAN`, where various different
// variants of written Chinese, Japanese, and Korean all use
// significantly different sets of Han characters and forms
// of shared characters. No sample language can be provided
// for many historical scripts as well.
//
// As of 1.18, this function checks the environment variables
// PANGO_LANGUAGE and LANGUAGE (checked in that order) first.
// If one of them is set, it is parsed as a list of language tags
// separated by colons or other separators.  This function
// will return the first language in the parsed list that Pango
// believes may use `script` for writing.  This last predicate
// is tested using pango_language_includes_script().  This can
// be used to control Pango's font selection for non-primary
// languages.  For example, a PANGO_LANGUAGE enviroment variable
// set to "en:fa" makes Pango choose fonts suitable for Persian (fa)
// instead of Arabic (ar) when a segment of Arabic text is found
// in an otherwise non-Arabic text.  The same trick can be used to
// choose a default language for `SCRIPT_HAN` when setting
// context language is not feasible.
func pango_script_get_sample_language(script Script) Language {
	/* Note that in the following, we want
	* pango_language_includes_script() for the sample language
	* to include the script, so alternate orthographies
	* (Shavian for English, Osmanya for Somali, etc), typically
	* have no sample language
	 */

	result := pango_script_get_default_language(script)
	if result != "" {
		return result
	}

	return sampleLanguages[script]
}

var sampleLanguages = map[Script]Language{
	language.Common:    "",   /* PANGO_language.Common */
	language.Inherited: "",   /* PANGO_language.Inherited */
	language.Arabic:    "ar", /* PANGO_language.Arabic */
	language.Armenian:  "hy", /* PANGO_language.Armenian */
	language.Bengali:   "bn", /* PANGO_language.Bengali */
	/* Used primarily in Taiwan, but not part of the standard
	 * zh-tw orthography  */
	language.Bopomofo: "",    /* PANGO_language.Bopomofo */
	language.Cherokee: "chr", /* PANGO_language.Cherokee */
	language.Coptic:   "cop", /* PANGO_language.Coptic */
	language.Cyrillic: "ru",  /* PANGO_language.Cyrillic */
	/* Deseret was used to write English */
	language.Deseret:    "",   /* PANGO_language.Deseret */
	language.Devanagari: "hi", /* PANGO_language.Devanagari */
	language.Ethiopic:   "am", /* PANGO_language.Ethiopic */
	language.Georgian:   "ka", /* PANGO_language.Georgian */
	language.Gothic:     "",   /* PANGO_language.Gothic */
	language.Greek:      "el", /* PANGO_language.Greek */
	language.Gujarati:   "gu", /* PANGO_language.Gujarati */
	language.Gurmukhi:   "pa", /* PANGO_language.Gurmukhi */
	language.Han:        "",   /* PANGO_language.Han */
	language.Hangul:     "ko", /* PANGO_language.Hangul */
	language.Hebrew:     "he", /* PANGO_language.Hebrew */
	language.Hiragana:   "ja", /* PANGO_language.Hiragana */
	language.Kannada:    "kn", /* PANGO_language.Kannada */
	language.Katakana:   "ja", /* PANGO_language.Katakana */
	language.Khmer:      "km", /* PANGO_language.Khmer */
	language.Lao:        "lo", /* PANGO_language.Lao */
	language.Latin:      "en", /* PANGO_language.Latin */
	language.Malayalam:  "ml", /* PANGO_language.Malayalam */
	language.Mongolian:  "mn", /* PANGO_language.Mongolian */
	language.Myanmar:    "my", /* PANGO_language.Myanmar */
	/* Ogham was used to write old Irish */
	language.Ogham:               "",    /* PANGO_language.Ogham */
	language.Old_Italic:          "",    /* PANGO_language.Old_italic */
	language.Oriya:               "or",  /* PANGO_language.Oriya */
	language.Runic:               "",    /* PANGO_language.Runic */
	language.Sinhala:             "si",  /* PANGO_language.Sinhala */
	language.Syriac:              "syr", /* PANGO_language.Syriac */
	language.Tamil:               "ta",  /* PANGO_language.Tamil */
	language.Telugu:              "te",  /* PANGO_language.Telugu */
	language.Thaana:              "dv",  /* PANGO_language.Thaana */
	language.Thai:                "th",  /* PANGO_language.Thai */
	language.Tibetan:             "bo",  /* PANGO_language.Tibetan */
	language.Canadian_Aboriginal: "iu",  /* PANGO_language.Canadian_aboriginal */
	language.Yi:                  "",    /* PANGO_language.Yi */
	language.Tagalog:             "tl",  /* PANGO_language.Tagalog */
	/* Phillipino languages/scripts */
	language.Hanunoo:  "hnn", /* PANGO_language.Hanunoo */
	language.Buhid:    "bku", /* PANGO_language.Buhid */
	language.Tagbanwa: "tbw", /* PANGO_language.Tagbanwa */

	language.Braille: "", /* PANGO_language.Braille */
	language.Cypriot: "", /* PANGO_language.Cypriot */
	language.Limbu:   "", /* PANGO_language.Limbu */
	/* Used for Somali (so) in the past */
	language.Osmanya: "", /* PANGO_language.Osmanya */
	/* The Shavian alphabet was designed for English */
	language.Shavian:  "",    /* PANGO_language.Shavian */
	language.Linear_B: "",    /* PANGO_language.Linear_b */
	language.Tai_Le:   "",    /* PANGO_language.Tai_le */
	language.Ugaritic: "uga", /* PANGO_language.Ugaritic */

	language.New_Tai_Lue: "",    /* PANGO_language.New_tai_lue */
	language.Buginese:    "bug", /* PANGO_language.Buginese */
	/* The original script for Old Church Slavonic (chu), later
	 * written with Cyrillic */
	language.Glagolitic: "", /* PANGO_language.Glagolitic */
	/* Used for for Berber (ber), but Arabic script is more common */
	language.Tifinagh:     "",    /* PANGO_language.Tifinagh */
	language.Syloti_Nagri: "syl", /* PANGO_language.Syloti_nagri */
	language.Old_Persian:  "peo", /* PANGO_language.Old_persian */
	language.Kharoshthi:   "",    /* PANGO_language.Kharoshthi */

	language.Unknown:    "",    /* PANGO_language.Unknown */
	language.Balinese:   "",    /* PANGO_language.Balinese */
	language.Cuneiform:  "",    /* PANGO_language.Cuneiform */
	language.Phoenician: "",    /* PANGO_language.Phoenician */
	language.Phags_Pa:   "",    /* PANGO_language.Phags_pa */
	language.Nko:        "nqo", /* PANGO_language.Nko */

	/* Unicode-5.1 additions */
	language.Kayah_Li:   "", /* PANGO_language.Kayah_li */
	language.Lepcha:     "", /* PANGO_language.Lepcha */
	language.Rejang:     "", /* PANGO_language.Rejang */
	language.Sundanese:  "", /* PANGO_language.Sundanese */
	language.Saurashtra: "", /* PANGO_language.Saurashtra */
	language.Cham:       "", /* PANGO_language.Cham */
	language.Ol_Chiki:   "", /* PANGO_language.Ol_chiki */
	language.Vai:        "", /* PANGO_language.Vai */
	language.Carian:     "", /* PANGO_language.Carian */
	language.Lycian:     "", /* PANGO_language.Lycian */
	language.Lydian:     "", /* PANGO_language.Lydian */

	/* Unicode-6.0 additions */
	language.Batak:   "", /* PANGO_language.Batak */
	language.Brahmi:  "", /* PANGO_language.Brahmi */
	language.Mandaic: "", /* PANGO_language.Mandaic */

	/* Unicode-6.1 additions */
	language.Chakma:               "", /* PANGO_language.Chakma */
	language.Meroitic_Cursive:     "", /* PANGO_language.Meroitic_cursive */
	language.Meroitic_Hieroglyphs: "", /* PANGO_language.Meroitic_hieroglyphs */
	language.Miao:                 "", /* PANGO_language.Miao */
	language.Sharada:              "", /* PANGO_language.Sharada */
	language.Sora_Sompeng:         "", /* PANGO_language.Sora_sompeng */
	language.Takri:                "", /* PANGO_SCRIPT_TAKRI */
}
