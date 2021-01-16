package pango

import "unicode"

/**
 * SECTION:scripts
 * @short_description:Identifying writing systems and languages
 * @title:Scripts and Languages
 *
 * The functions in this section are used to identify the writing
 * system, or <firstterm>script</firstterm> of individual characters
 * and of ranges within a larger text string.
 */

// Script enumeration identifies different writing
// systems. The values correspond to the names as defined in the
// Unicode standard and used as keys in the unicode.Scripts map
// defined in the standard library.
type Script string

const (
	SCRIPT_UNKNOWN                Script = ""
	SCRIPT_ADLAM                  Script = "Adlam"
	SCRIPT_AHOM                   Script = "Ahom"
	SCRIPT_ANATOLIAN_HIEROGLYPHS  Script = "Anatolian_Hieroglyphs"
	SCRIPT_ARABIC                 Script = "Arabic"
	SCRIPT_ARMENIAN               Script = "Armenian"
	SCRIPT_AVESTAN                Script = "Avestan"
	SCRIPT_BALINESE               Script = "Balinese"
	SCRIPT_BAMUM                  Script = "Bamum"
	SCRIPT_BASSA_VAH              Script = "Bassa_Vah"
	SCRIPT_BATAK                  Script = "Batak"
	SCRIPT_BENGALI                Script = "Bengali"
	SCRIPT_BHAIKSUKI              Script = "Bhaiksuki"
	SCRIPT_BOPOMOFO               Script = "Bopomofo"
	SCRIPT_BRAHMI                 Script = "Brahmi"
	SCRIPT_BRAILLE                Script = "Braille"
	SCRIPT_BUGINESE               Script = "Buginese"
	SCRIPT_BUHID                  Script = "Buhid"
	SCRIPT_CANADIAN_ABORIGINAL    Script = "Canadian_Aboriginal"
	SCRIPT_CARIAN                 Script = "Carian"
	SCRIPT_CAUCASIAN_ALBANIAN     Script = "Caucasian_Albanian"
	SCRIPT_CHAKMA                 Script = "Chakma"
	SCRIPT_CHAM                   Script = "Cham"
	SCRIPT_CHEROKEE               Script = "Cherokee"
	SCRIPT_COMMON                 Script = "Common"
	SCRIPT_COPTIC                 Script = "Coptic"
	SCRIPT_CUNEIFORM              Script = "Cuneiform"
	SCRIPT_CYPRIOT                Script = "Cypriot"
	SCRIPT_CYRILLIC               Script = "Cyrillic"
	SCRIPT_DESERET                Script = "Deseret"
	SCRIPT_DEVANAGARI             Script = "Devanagari"
	SCRIPT_DOGRA                  Script = "Dogra"
	SCRIPT_DUPLOYAN               Script = "Duployan"
	SCRIPT_EGYPTIAN_HIEROGLYPHS   Script = "Egyptian_Hieroglyphs"
	SCRIPT_ELBASAN                Script = "Elbasan"
	SCRIPT_ELYMAIC                Script = "Elymaic"
	SCRIPT_ETHIOPIC               Script = "Ethiopic"
	SCRIPT_GEORGIAN               Script = "Georgian"
	SCRIPT_GLAGOLITIC             Script = "Glagolitic"
	SCRIPT_GOTHIC                 Script = "Gothic"
	SCRIPT_GRANTHA                Script = "Grantha"
	SCRIPT_GREEK                  Script = "Greek"
	SCRIPT_GUJARATI               Script = "Gujarati"
	SCRIPT_GUNJALA_GONDI          Script = "Gunjala_Gondi"
	SCRIPT_GURMUKHI               Script = "Gurmukhi"
	SCRIPT_HAN                    Script = "Han"
	SCRIPT_HANGUL                 Script = "Hangul"
	SCRIPT_HANIFI_ROHINGYA        Script = "Hanifi_Rohingya"
	SCRIPT_HANUNOO                Script = "Hanunoo"
	SCRIPT_HATRAN                 Script = "Hatran"
	SCRIPT_HEBREW                 Script = "Hebrew"
	SCRIPT_HIRAGANA               Script = "Hiragana"
	SCRIPT_IMPERIAL_ARAMAIC       Script = "Imperial_Aramaic"
	SCRIPT_INHERITED              Script = "Inherited"
	SCRIPT_INSCRIPTIONAL_PAHLAVI  Script = "Inscriptional_Pahlavi"
	SCRIPT_INSCRIPTIONAL_PARTHIAN Script = "Inscriptional_Parthian"
	SCRIPT_JAVANESE               Script = "Javanese"
	SCRIPT_KAITHI                 Script = "Kaithi"
	SCRIPT_KANNADA                Script = "Kannada"
	SCRIPT_KATAKANA               Script = "Katakana"
	SCRIPT_KAYAH_LI               Script = "Kayah_Li"
	SCRIPT_KHAROSHTHI             Script = "Kharoshthi"
	SCRIPT_KHMER                  Script = "Khmer"
	SCRIPT_KHOJKI                 Script = "Khojki"
	SCRIPT_KHUDAWADI              Script = "Khudawadi"
	SCRIPT_LAO                    Script = "Lao"
	SCRIPT_LATIN                  Script = "Latin"
	SCRIPT_LEPCHA                 Script = "Lepcha"
	SCRIPT_LIMBU                  Script = "Limbu"
	SCRIPT_LINEAR_A               Script = "Linear_A"
	SCRIPT_LINEAR_B               Script = "Linear_B"
	SCRIPT_LISU                   Script = "Lisu"
	SCRIPT_LYCIAN                 Script = "Lycian"
	SCRIPT_LYDIAN                 Script = "Lydian"
	SCRIPT_MAHAJANI               Script = "Mahajani"
	SCRIPT_MAKASAR                Script = "Makasar"
	SCRIPT_MALAYALAM              Script = "Malayalam"
	SCRIPT_MANDAIC                Script = "Mandaic"
	SCRIPT_MANICHAEAN             Script = "Manichaean"
	SCRIPT_MARCHEN                Script = "Marchen"
	SCRIPT_MASARAM_GONDI          Script = "Masaram_Gondi"
	SCRIPT_MEDEFAIDRIN            Script = "Medefaidrin"
	SCRIPT_MEETEI_MAYEK           Script = "Meetei_Mayek"
	SCRIPT_MENDE_KIKAKUI          Script = "Mende_Kikakui"
	SCRIPT_MEROITIC_CURSIVE       Script = "Meroitic_Cursive"
	SCRIPT_MEROITIC_HIEROGLYPHS   Script = "Meroitic_Hieroglyphs"
	SCRIPT_MIAO                   Script = "Miao"
	SCRIPT_MODI                   Script = "Modi"
	SCRIPT_MONGOLIAN              Script = "Mongolian"
	SCRIPT_MRO                    Script = "Mro"
	SCRIPT_MULTANI                Script = "Multani"
	SCRIPT_MYANMAR                Script = "Myanmar"
	SCRIPT_NABATAEAN              Script = "Nabataean"
	SCRIPT_NANDINAGARI            Script = "Nandinagari"
	SCRIPT_NEW_TAI_LUE            Script = "New_Tai_Lue"
	SCRIPT_NEWA                   Script = "Newa"
	SCRIPT_NKO                    Script = "Nko"
	SCRIPT_NUSHU                  Script = "Nushu"
	SCRIPT_NYIAKENG_PUACHUE_HMONG Script = "Nyiakeng_Puachue_Hmong"
	SCRIPT_OGHAM                  Script = "Ogham"
	SCRIPT_OL_CHIKI               Script = "Ol_Chiki"
	SCRIPT_OLD_HUNGARIAN          Script = "Old_Hungarian"
	SCRIPT_OLD_ITALIC             Script = "Old_Italic"
	SCRIPT_OLD_NORTH_ARABIAN      Script = "Old_North_Arabian"
	SCRIPT_OLD_PERMIC             Script = "Old_Permic"
	SCRIPT_OLD_PERSIAN            Script = "Old_Persian"
	SCRIPT_OLD_SOGDIAN            Script = "Old_Sogdian"
	SCRIPT_OLD_SOUTH_ARABIAN      Script = "Old_South_Arabian"
	SCRIPT_OLD_TURKIC             Script = "Old_Turkic"
	SCRIPT_ORIYA                  Script = "Oriya"
	SCRIPT_OSAGE                  Script = "Osage"
	SCRIPT_OSMANYA                Script = "Osmanya"
	SCRIPT_PAHAWH_HMONG           Script = "Pahawh_Hmong"
	SCRIPT_PALMYRENE              Script = "Palmyrene"
	SCRIPT_PAU_CIN_HAU            Script = "Pau_Cin_Hau"
	SCRIPT_PHAGS_PA               Script = "Phags_Pa"
	SCRIPT_PHOENICIAN             Script = "Phoenician"
	SCRIPT_PSALTER_PAHLAVI        Script = "Psalter_Pahlavi"
	SCRIPT_REJANG                 Script = "Rejang"
	SCRIPT_RUNIC                  Script = "Runic"
	SCRIPT_SAMARITAN              Script = "Samaritan"
	SCRIPT_SAURASHTRA             Script = "Saurashtra"
	SCRIPT_SHARADA                Script = "Sharada"
	SCRIPT_SHAVIAN                Script = "Shavian"
	SCRIPT_SIDDHAM                Script = "Siddham"
	SCRIPT_SIGNWRITING            Script = "SignWriting"
	SCRIPT_SINHALA                Script = "Sinhala"
	SCRIPT_SOGDIAN                Script = "Sogdian"
	SCRIPT_SORA_SOMPENG           Script = "Sora_Sompeng"
	SCRIPT_SOYOMBO                Script = "Soyombo"
	SCRIPT_SUNDANESE              Script = "Sundanese"
	SCRIPT_SYLOTI_NAGRI           Script = "Syloti_Nagri"
	SCRIPT_SYRIAC                 Script = "Syriac"
	SCRIPT_TAGALOG                Script = "Tagalog"
	SCRIPT_TAGBANWA               Script = "Tagbanwa"
	SCRIPT_TAI_LE                 Script = "Tai_Le"
	SCRIPT_TAI_THAM               Script = "Tai_Tham"
	SCRIPT_TAI_VIET               Script = "Tai_Viet"
	SCRIPT_TAKRI                  Script = "Takri"
	SCRIPT_TAMIL                  Script = "Tamil"
	SCRIPT_TANGUT                 Script = "Tangut"
	SCRIPT_TELUGU                 Script = "Telugu"
	SCRIPT_THAANA                 Script = "Thaana"
	SCRIPT_THAI                   Script = "Thai"
	SCRIPT_TIBETAN                Script = "Tibetan"
	SCRIPT_TIFINAGH               Script = "Tifinagh"
	SCRIPT_TIRHUTA                Script = "Tirhuta"
	SCRIPT_UGARITIC               Script = "Ugaritic"
	SCRIPT_VAI                    Script = "Vai"
	SCRIPT_WANCHO                 Script = "Wancho"
	SCRIPT_WARANG_CITI            Script = "Warang_Citi"
	SCRIPT_YI                     Script = "Yi"
	SCRIPT_ZANABAZAR_SQUARE       Script = "Zanabazar_Square"
)

// Looks up the script for a particular character (as defined by
// Unicode Standard Annex #24).
func pango_script_for_unichar(r rune) Script {
	for name, table := range unicode.Scripts {
		if unicode.Is(table, r) {
			return Script(name)
		}
	}
	return SCRIPT_UNKNOWN
}

const PAREN_STACK_DEPTH = 128

type ParenStackEntry struct {
	pair_index  int
	script_code Script
}

// ScriptIter is used to break a string of
// Unicode text into runs by Unicode script.
//
// The range to which the iterator currently points,
// is the set of indexes `i` where `script_start` <= i < `script_end`.
// That is, it doesn't include the character stored at `script_end`.
type ScriptIter struct {
	text []rune

	script_start int // index into text
	script_end   int // index into text
	script_code  Script

	paren_stack [PAREN_STACK_DEPTH]ParenStackEntry
	paren_sp    int
}

// pango_script_iter_new creates a new script iterator,initialized to point at the first range in the text.
// If the string is empty, it will point at an empty range.
func pango_script_iter_new(text []rune) *ScriptIter {
	var out ScriptIter
	out._pango_script_iter_init(text)
	return &out
}

//  static ScriptIter *pango_script_iter_copy (ScriptIter *iter);

//  G_DEFINE_BOXED_TYPE (ScriptIter,
// 					  pango_script_iter,
// 					  pango_script_iter_copy,
// 					  pango_script_iter_free)

func (iter *ScriptIter) _pango_script_iter_init(text []rune) {
	iter.text = text

	iter.script_code = SCRIPT_COMMON

	iter.paren_sp = -1

	iter.pango_script_iter_next()
}

//  static ScriptIter *
//  pango_script_iter_copy (ScriptIter *iter)
//  {
//    return g_slice_dup (ScriptIter, iter);
//  }

//  void
//  _pango_script_iter_fini (ScriptIter *iter)
//  {
//  }

/**
 * pango_script_iter_get_range:
 * @iter: a #ScriptIter
 * @start: (out) (allow-none): location to store start position of the range, or %NULL
 * @end: (out) (allow-none): location to store end position of the range, or %NULL
 * @script: (out) (allow-none): location to store script for range, or %NULL
 *
 * Gets information about the range to which @iter currently points.
 * The range is the set of locations p where *start <= p < *end.
 * (That is, it doesn't include the character stored at *end)
 *
 * Note that while the type of the @script argument is declared
 * as Script, as of Pango 1.18, this function simply returns
 * GUnicodeScript values. Callers must be prepared to handle unknown
 * values.
 *
 * Since: 1.4
 **/
//  void
//  pango_script_iter_get_range (ScriptIter  *iter,
// 							  const char      **start,
// 							  const char      **end,
// 							  Script      *script)
//  {
//    if (start)
// 	 *start = iter.script_start;
//    if (end)
// 	 *end = iter.script_end;
//    if (script)
// 	 *script = iter.script_code;
//  }

var paired_chars = [...]rune{
	0x0028, 0x0029, /* ascii paired punctuation */
	0x003c, 0x003e,
	0x005b, 0x005d,
	0x007b, 0x007d,
	0x00ab, 0x00bb, /* guillemets */
	0x0f3a, 0x0f3b, /* tibetan */
	0x0f3c, 0x0f3d,
	0x169b, 0x169c, /* ogham */
	0x2018, 0x2019, /* general punctuation */
	0x201c, 0x201d,
	0x2039, 0x203a,
	0x2045, 0x2046,
	0x207d, 0x207e,
	0x208d, 0x208e,
	0x27e6, 0x27e7, /* math */
	0x27e8, 0x27e9,
	0x27ea, 0x27eb,
	0x27ec, 0x27ed,
	0x27ee, 0x27ef,
	0x2983, 0x2984,
	0x2985, 0x2986,
	0x2987, 0x2988,
	0x2989, 0x298a,
	0x298b, 0x298c,
	0x298d, 0x298e,
	0x298f, 0x2990,
	0x2991, 0x2992,
	0x2993, 0x2994,
	0x2995, 0x2996,
	0x2997, 0x2998,
	0x29fc, 0x29fd,
	0x2e02, 0x2e03,
	0x2e04, 0x2e05,
	0x2e09, 0x2e0a,
	0x2e0c, 0x2e0d,
	0x2e1c, 0x2e1d,
	0x2e20, 0x2e21,
	0x2e22, 0x2e23,
	0x2e24, 0x2e25,
	0x2e26, 0x2e27,
	0x2e28, 0x2e29,
	0x3008, 0x3009, /* chinese paired punctuation */
	0x300a, 0x300b,
	0x300c, 0x300d,
	0x300e, 0x300f,
	0x3010, 0x3011,
	0x3014, 0x3015,
	0x3016, 0x3017,
	0x3018, 0x3019,
	0x301a, 0x301b,
	0xfe59, 0xfe5a,
	0xfe5b, 0xfe5c,
	0xfe5d, 0xfe5e,
	0xff08, 0xff09,
	0xff3b, 0xff3d,
	0xff5b, 0xff5d,
	0xff5f, 0xff60,
	0xff62, 0xff63,
}

func get_pair_index(ch rune) int {
	lower := 0
	upper := len(paired_chars) - 1

	for lower <= upper {
		mid := (lower + upper) / 2

		if ch < paired_chars[mid] {
			upper = mid - 1
		} else if ch > paired_chars[mid] {
			lower = mid + 1
		} else {
			return mid
		}
	}

	return -1
}

func (s Script) isRealScript() bool {
	switch s {
	case SCRIPT_UNKNOWN, SCRIPT_COMMON, SCRIPT_INHERITED:
		return false
	default:
		return true
	}
}

func (s1 Script) isSameScript(s2 Script) bool {
	return !s1.isRealScript() || !s2.isRealScript() || s1 == s2
}

// pango_script_iter_next advances to the next range. If `iter`
// is already at the end, it is left unchanged and `false`
// is returned.
func (iter *ScriptIter) pango_script_iter_next() bool {
	if iter.script_end >= len(iter.text)-1 {
		return false
	}

	start_sp := iter.paren_sp
	iter.script_code = SCRIPT_COMMON
	iter.script_start = iter.script_end

	for ; iter.script_end < len(iter.text); iter.script_end++ {
		ch := iter.text[iter.script_end]

		var pair_index int

		sc := pango_script_for_unichar(ch)
		if sc != SCRIPT_COMMON {
			pair_index = -1
		} else {
			pair_index = get_pair_index(ch)
		}

		/*
		* Paired character handling:
		*
		* if it's an open character, push it onto the stack.
		* if it's a close character, find the matching open on the
		* stack, and use that script code. Any non-matching open
		* characters above it on the stack will be poped.
		 */
		if pair_index >= 0 {
			if pair_index&1 == 0 { // is open ?
				/*
				* If the paren stack is full, empty it. This
				* means that deeply nested paired punctuation
				* characters will be ignored, but that's an unusual
				* case, and it's better to ignore them than to
				* write off the end of the stack...
				 */
				iter.paren_sp++
				if iter.paren_sp >= PAREN_STACK_DEPTH {
					iter.paren_sp = 0
				}

				iter.paren_stack[iter.paren_sp].pair_index = pair_index
				iter.paren_stack[iter.paren_sp].script_code = iter.script_code
			} else if iter.paren_sp >= 0 {
				pi := pair_index & ^1

				for iter.paren_sp >= 0 && iter.paren_stack[iter.paren_sp].pair_index != pi {
					iter.paren_sp--
				}

				if iter.paren_sp < start_sp {
					start_sp = iter.paren_sp
				}

				if iter.paren_sp >= 0 {
					sc = iter.paren_stack[iter.paren_sp].script_code
				}
			}
		}

		if iter.script_code.isSameScript(sc) {
			if !iter.script_code.isRealScript() && sc.isRealScript() {
				iter.script_code = sc

				/*
				* now that we have a final script code, fix any open
				* characters we pushed before we knew the script code.
				 */
				for start_sp < iter.paren_sp {
					start_sp++
					iter.paren_stack[start_sp].script_code = iter.script_code
				}
			}

			/*
			* if this character is a close paired character,
			* pop it from the stack
			 */
			if pair_index >= 0 && pair_index&1 != 0 && iter.paren_sp >= 0 {
				iter.paren_sp--

				if iter.paren_sp < start_sp {
					start_sp = iter.paren_sp
				}
			}
		} else {
			/* Different script, we're done */
			break
		}
	}

	return true
}
