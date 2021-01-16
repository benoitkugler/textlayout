package pango

import (
	"os"
	"strings"
	"sync"
)

// Language is used to represent a language.
//
// The actual value is the RFC-3066 format of the language
type Language string

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

var canon_map = [256]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '-', 0, 0,
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 0, 0, 0, 0, 0, 0,
	'-', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 0, 0, 0, 0, '-',
	0, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 0, 0, 0, 0, 0,
}

// canonicalize the language input
func canonicalize(v []byte) []byte {
	out := make([]byte, 0, len(v))
	for _, b := range v {
		can := canon_map[b]
		if can == 0 {
			break
		}
		out = append(out, can)
	}
	return out
}

// func lang_equal (v1,  v2 []byte) bool {
//  i := 0
//  _ = v2[len(v1)-1] // hint for BCE
//    for ; i < len(v1); i ++ {
// 	   p1, p2 := v1[i], v2[i]
// 		if !(canon_map[p1] !=0  && canon_map[p1] == canon_map[p2]) {
// 			break
// 		}
// 	 }

//    return (canon_map[v1[i]] == canon_map[v2[i]]);
//  }

func lang_hash(key []byte) uint32 {
	var h uint32
	for _, p := range key {
		if canon_map[p] == 0 {
			break
		}
		h = (h << 5) - h + uint32(canon_map[p])
	}

	return h
}

// Return the Unix-style locale string for the language currently in
// effect.
// TODO: For now, we only check the  environment variables LC_ALL, LC_CTYPE or
// LANG (in that order).
func _pango_get_lc_ctype() string {
	p, ok := os.LookupEnv("LC_ALL")
	if ok {
		return p
	}

	p, ok = os.LookupEnv("LC_CTYPE")
	if ok {
		return p
	}

	p, ok = os.LookupEnv("LANG")
	if ok {
		return p
	}
	return "C"
}

var (
	defaultLanguage     Language
	defaultLanguageOnce sync.Once
)

// pango_language_get_default returns the Language for the current locale of the process.
// Note that this can change over the life of an application.
//
// On Unix systems, the return value is derived from
// `setlocale(LC_CTYPE, nil)`, and the user can
// affect this through the environment variables LC_ALL, LC_CTYPE or
// LANG (checked in that order). The locale string typically is in
// the form lang_COUNTRY, where lang is an ISO-639 language code, and
// COUNTRY is an ISO-3166 country code. For instance, sv_FI for
// Swedish as written in Finland or pt_BR for Portuguese as written in
// Brazil.
//
// On Windows, the C library does not use any such environment
// variables, and setting them won't affect the behavior of functions
// like ctime(). The user sets the locale through the Regional Options
// in the Control Panel. The C library (in the setlocale() function)
// does not use country and language codes, but country and language
// names spelled out in English.
// However, this function does check the above environment
// variables, and does return a Unix-style locale string based on
// either said environment variables or the thread's current locale.
//
// Your application should call `setlocale(LC_ALL, "")`
// for the user settings to take effect.  Gtk+ does this in its initialization
// functions automatically (by calling gtk_set_locale()).
// See <literal>man setlocale</literal> for more details.
func pango_language_get_default() Language {
	defaultLanguageOnce.Do(func() {
		lc_ctype := _pango_get_lc_ctype()
		defaultLanguage = pango_language_from_string(lc_ctype)
	})

	return defaultLanguage
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
// Use pango_language_get_default() if you want to get the Language for
// the current locale of the process.
func pango_language_from_string(language string) Language {
	// Note: for now, we use a very simple implementation:
	// we just store the canonical form
	// The C implementation has a more refined algorithm,
	// using a map from language to pointer,
	// which replace comparison and copies of string to comparison and copies of pointers

	can := canonicalize([]byte(language))
	// hash := lang_hash(can)

	// result, ok := languageHashTable[hash]
	// if !ok { // insert a new language
	// 	languageHashTableLock.Lock()
	// 	defer languageHashTableLock.Unlock()
	// 	languageHashTable[hash] = Language(hash)
	// }

	return Language(can)
}

func (lang Language) compute_derived_language(script Script) Language {
	var derivedLang Language

	/* Make sure the language tag is consistent with the derived
	 * script. There is no point in marking up a section of
	 * Arabic text with the "en" language tag.
	 */
	if lang != "" && lang.pango_language_includes_script(script) {
		derivedLang = lang
	} else {
		derivedLang = script.pango_script_get_sample_language()
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
 * pangoLanguageMatches:
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
 * or the range is a prefix of the tag, and the character after it
 * in the tag is '-'.
 **/
// TODO: maybe simplify
func pangoLanguageMatches(language Language, rangeList string) bool {
	langStr := string(language)
	p := rangeList
	done := false

	for !done {
		end := strings.IndexAny(p, languageSeparators)
		if end == -1 {
			end = len(p)
			done = true
		}

		// truncate end if needed
		endSafe := end
		if len(langStr) < end {
			endSafe = len(langStr)
		}

		if len(p) == 0 || p[0] == '*' || (langStr != "" && langStr[:endSafe] == p[:endSafe] &&
			len(langStr) == end || langStr[end] == '-') {
			return true
		}

		if !done {
			p = p[end+1:]
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

// pango_language_get_sample_string get a string that is representative of the characters needed to
// render a particular language.
//
// The sample text may be a pangram, but is not necessarily.  It is chosen to
// be demonstrative of normal text in the language, as well as exposing font
// feature requirements unique to the language.  It is suitable for use
// as sample text in a font selection dialog.
//
// If `language` is empty, the default language as found by
// pango_language_get_default() is used.
//
// If Pango does not have a sample string for `language`, the classic
// "The quick brown fox..." is returned.
func (language Language) pango_language_get_sample_string() string {

	if language == "" {
		language = pango_language_get_default()
	}

	lang_info := findBestLangMatchCached(language, lang_texts)

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
func (language Language) pango_language_get_scripts() []Script {

	scriptRec := findBestLangMatchCached(language, pango_script_for_lang)

	if scriptRec == nil {
		return nil
	}
	script_for_lang := scriptRec.(recordScript)
	if script_for_lang.scripts[0] == "" {
		return nil
	}
	for j, s := range script_for_lang.scripts {
		if s == "" {
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
func (language Language) pango_language_includes_script(script Script) bool {
	if !script.isRealScript() {
		return true
	}

	scripts := language.pango_language_get_scripts()
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

func (script Script) pango_script_get_default_language() Language {
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
		if p.pango_language_includes_script(script) {
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
//   * the language returned by pango_language_get_default().
//   *
//   * When choosing language-specific resources, such as the sample
//   * text returned by pango_language_get_sample_string(), you should
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
func (script Script) pango_script_get_sample_language() Language {
	/* Note that in the following, we want
	* pango_language_includes_script() for the sample language
	* to include the script, so alternate orthographies
	* (Shavian for English, Osmanya for Somali, etc), typically
	* have no sample language
	 */

	result := script.pango_script_get_default_language()
	if result != "" {
		return result
	}

	return sampleLanguages[script]
}

var sampleLanguages = map[Script]Language{
	SCRIPT_COMMON:    "",   /* PANGO_SCRIPT_COMMON */
	SCRIPT_INHERITED: "",   /* PANGO_SCRIPT_INHERITED */
	SCRIPT_ARABIC:    "ar", /* PANGO_SCRIPT_ARABIC */
	SCRIPT_ARMENIAN:  "hy", /* PANGO_SCRIPT_ARMENIAN */
	SCRIPT_BENGALI:   "bn", /* PANGO_SCRIPT_BENGALI */
	/* Used primarily in Taiwan, but not part of the standard
	 * zh-tw orthography  */
	SCRIPT_BOPOMOFO: "",    /* PANGO_SCRIPT_BOPOMOFO */
	SCRIPT_CHEROKEE: "chr", /* PANGO_SCRIPT_CHEROKEE */
	SCRIPT_COPTIC:   "cop", /* PANGO_SCRIPT_COPTIC */
	SCRIPT_CYRILLIC: "ru",  /* PANGO_SCRIPT_CYRILLIC */
	/* Deseret was used to write English */
	SCRIPT_DESERET:    "",   /* PANGO_SCRIPT_DESERET */
	SCRIPT_DEVANAGARI: "hi", /* PANGO_SCRIPT_DEVANAGARI */
	SCRIPT_ETHIOPIC:   "am", /* PANGO_SCRIPT_ETHIOPIC */
	SCRIPT_GEORGIAN:   "ka", /* PANGO_SCRIPT_GEORGIAN */
	SCRIPT_GOTHIC:     "",   /* PANGO_SCRIPT_GOTHIC */
	SCRIPT_GREEK:      "el", /* PANGO_SCRIPT_GREEK */
	SCRIPT_GUJARATI:   "gu", /* PANGO_SCRIPT_GUJARATI */
	SCRIPT_GURMUKHI:   "pa", /* PANGO_SCRIPT_GURMUKHI */
	SCRIPT_HAN:        "",   /* PANGO_SCRIPT_HAN */
	SCRIPT_HANGUL:     "ko", /* PANGO_SCRIPT_HANGUL */
	SCRIPT_HEBREW:     "he", /* PANGO_SCRIPT_HEBREW */
	SCRIPT_HIRAGANA:   "ja", /* PANGO_SCRIPT_HIRAGANA */
	SCRIPT_KANNADA:    "kn", /* PANGO_SCRIPT_KANNADA */
	SCRIPT_KATAKANA:   "ja", /* PANGO_SCRIPT_KATAKANA */
	SCRIPT_KHMER:      "km", /* PANGO_SCRIPT_KHMER */
	SCRIPT_LAO:        "lo", /* PANGO_SCRIPT_LAO */
	SCRIPT_LATIN:      "en", /* PANGO_SCRIPT_LATIN */
	SCRIPT_MALAYALAM:  "ml", /* PANGO_SCRIPT_MALAYALAM */
	SCRIPT_MONGOLIAN:  "mn", /* PANGO_SCRIPT_MONGOLIAN */
	SCRIPT_MYANMAR:    "my", /* PANGO_SCRIPT_MYANMAR */
	/* Ogham was used to write old Irish */
	SCRIPT_OGHAM:               "",    /* PANGO_SCRIPT_OGHAM */
	SCRIPT_OLD_ITALIC:          "",    /* PANGO_SCRIPT_OLD_ITALIC */
	SCRIPT_ORIYA:               "or",  /* PANGO_SCRIPT_ORIYA */
	SCRIPT_RUNIC:               "",    /* PANGO_SCRIPT_RUNIC */
	SCRIPT_SINHALA:             "si",  /* PANGO_SCRIPT_SINHALA */
	SCRIPT_SYRIAC:              "syr", /* PANGO_SCRIPT_SYRIAC */
	SCRIPT_TAMIL:               "ta",  /* PANGO_SCRIPT_TAMIL */
	SCRIPT_TELUGU:              "te",  /* PANGO_SCRIPT_TELUGU */
	SCRIPT_THAANA:              "dv",  /* PANGO_SCRIPT_THAANA */
	SCRIPT_THAI:                "th",  /* PANGO_SCRIPT_THAI */
	SCRIPT_TIBETAN:             "bo",  /* PANGO_SCRIPT_TIBETAN */
	SCRIPT_CANADIAN_ABORIGINAL: "iu",  /* PANGO_SCRIPT_CANADIAN_ABORIGINAL */
	SCRIPT_YI:                  "",    /* PANGO_SCRIPT_YI */
	SCRIPT_TAGALOG:             "tl",  /* PANGO_SCRIPT_TAGALOG */
	/* Phillipino languages/scripts */
	SCRIPT_HANUNOO:  "hnn", /* PANGO_SCRIPT_HANUNOO */
	SCRIPT_BUHID:    "bku", /* PANGO_SCRIPT_BUHID */
	SCRIPT_TAGBANWA: "tbw", /* PANGO_SCRIPT_TAGBANWA */

	SCRIPT_BRAILLE: "", /* PANGO_SCRIPT_BRAILLE */
	SCRIPT_CYPRIOT: "", /* PANGO_SCRIPT_CYPRIOT */
	SCRIPT_LIMBU:   "", /* PANGO_SCRIPT_LIMBU */
	/* Used for Somali (so) in the past */
	SCRIPT_OSMANYA: "", /* PANGO_SCRIPT_OSMANYA */
	/* The Shavian alphabet was designed for English */
	SCRIPT_SHAVIAN:  "",    /* PANGO_SCRIPT_SHAVIAN */
	SCRIPT_LINEAR_B: "",    /* PANGO_SCRIPT_LINEAR_B */
	SCRIPT_TAI_LE:   "",    /* PANGO_SCRIPT_TAI_LE */
	SCRIPT_UGARITIC: "uga", /* PANGO_SCRIPT_UGARITIC */

	SCRIPT_NEW_TAI_LUE: "",    /* PANGO_SCRIPT_NEW_TAI_LUE */
	SCRIPT_BUGINESE:    "bug", /* PANGO_SCRIPT_BUGINESE */
	/* The original script for Old Church Slavonic (chu), later
	 * written with Cyrillic */
	SCRIPT_GLAGOLITIC: "", /* PANGO_SCRIPT_GLAGOLITIC */
	/* Used for for Berber (ber), but Arabic script is more common */
	SCRIPT_TIFINAGH:     "",    /* PANGO_SCRIPT_TIFINAGH */
	SCRIPT_SYLOTI_NAGRI: "syl", /* PANGO_SCRIPT_SYLOTI_NAGRI */
	SCRIPT_OLD_PERSIAN:  "peo", /* PANGO_SCRIPT_OLD_PERSIAN */
	SCRIPT_KHAROSHTHI:   "",    /* PANGO_SCRIPT_KHAROSHTHI */

	SCRIPT_UNKNOWN:    "",    /* PANGO_SCRIPT_UNKNOWN */
	SCRIPT_BALINESE:   "",    /* PANGO_SCRIPT_BALINESE */
	SCRIPT_CUNEIFORM:  "",    /* PANGO_SCRIPT_CUNEIFORM */
	SCRIPT_PHOENICIAN: "",    /* PANGO_SCRIPT_PHOENICIAN */
	SCRIPT_PHAGS_PA:   "",    /* PANGO_SCRIPT_PHAGS_PA */
	SCRIPT_NKO:        "nqo", /* PANGO_SCRIPT_NKO */

	/* Unicode-5.1 additions */
	SCRIPT_KAYAH_LI:   "", /* PANGO_SCRIPT_KAYAH_LI */
	SCRIPT_LEPCHA:     "", /* PANGO_SCRIPT_LEPCHA */
	SCRIPT_REJANG:     "", /* PANGO_SCRIPT_REJANG */
	SCRIPT_SUNDANESE:  "", /* PANGO_SCRIPT_SUNDANESE */
	SCRIPT_SAURASHTRA: "", /* PANGO_SCRIPT_SAURASHTRA */
	SCRIPT_CHAM:       "", /* PANGO_SCRIPT_CHAM */
	SCRIPT_OL_CHIKI:   "", /* PANGO_SCRIPT_OL_CHIKI */
	SCRIPT_VAI:        "", /* PANGO_SCRIPT_VAI */
	SCRIPT_CARIAN:     "", /* PANGO_SCRIPT_CARIAN */
	SCRIPT_LYCIAN:     "", /* PANGO_SCRIPT_LYCIAN */
	SCRIPT_LYDIAN:     "", /* PANGO_SCRIPT_LYDIAN */

	/* Unicode-6.0 additions */
	SCRIPT_BATAK:   "", /* PANGO_SCRIPT_BATAK */
	SCRIPT_BRAHMI:  "", /* PANGO_SCRIPT_BRAHMI */
	SCRIPT_MANDAIC: "", /* PANGO_SCRIPT_MANDAIC */

	/* Unicode-6.1 additions */
	SCRIPT_CHAKMA:               "", /* PANGO_SCRIPT_CHAKMA */
	SCRIPT_MEROITIC_CURSIVE:     "", /* PANGO_SCRIPT_MEROITIC_CURSIVE */
	SCRIPT_MEROITIC_HIEROGLYPHS: "", /* PANGO_SCRIPT_MEROITIC_HIEROGLYPHS */
	SCRIPT_MIAO:                 "", /* PANGO_SCRIPT_MIAO */
	SCRIPT_SHARADA:              "", /* PANGO_SCRIPT_SHARADA */
	SCRIPT_SORA_SOMPENG:         "", /* PANGO_SCRIPT_SORA_SOMPENG */
	SCRIPT_TAKRI:                "", /* PANGO_SCRIPT_TAKRI */
}
