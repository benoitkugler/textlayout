package harfbuzz

import (
	"encoding/hex"
	"strings"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-tag.cc Copyright © 2009  Red Hat, Inc. 2011  Google, Inc. Behdad Esfahbod, Roozbeh Pournader

var (
	// OpenType script tag, `DFLT`, for features that are not script-specific.
	HB_OT_TAG_DEFAULT_SCRIPT = newTag('D', 'F', 'L', 'T')
	// OpenType language tag, `dflt`. Not a valid language tag, but some fonts
	// mistakenly use it.
	HB_OT_TAG_DEFAULT_LANGUAGE = newTag('d', 'f', 'l', 't')
)

//  /* hb_script_t */

//  static hb_tag_t
//  hb_ot_old_tag_from_script (hb_script_t script)
//  {
//    /* This seems to be accurate as of end of 2012. */

//    switch ((hb_tag_t) script)
//    {
// 	 case HB_SCRIPT_INVALID:		return HB_OT_TAG_DEFAULT_SCRIPT;

// 	 /* KATAKANA and HIRAGANA both map to 'kana' */
// 	 case HB_SCRIPT_HIRAGANA:		return HB_TAG('k','a','n','a');

// 	 /* Spaces at the end are preserved, unlike ISO 15924 */
// 	 case HB_SCRIPT_LAO:			return HB_TAG('l','a','o',' ');
// 	 case HB_SCRIPT_YI:			return HB_TAG('y','i',' ',' ');
// 	 /* Unicode-5.0 additions */
// 	 case HB_SCRIPT_NKO:			return HB_TAG('n','k','o',' ');
// 	 /* Unicode-5.1 additions */
// 	 case HB_SCRIPT_VAI:			return HB_TAG('v','a','i',' ');
//    }

//    /* Else, just change first char to lowercase and return */
//    return ((hb_tag_t) script) | 0x20000000u;
//  }

//  static hb_script_t
//  hb_ot_old_tag_to_script (hb_tag_t tag)
//  {
//    if (unlikely (tag == HB_OT_TAG_DEFAULT_SCRIPT))
// 	 return HB_SCRIPT_INVALID;

//    /* This side of the conversion is fully algorithmic. */

//    /* Any spaces at the end of the tag are replaced by repeating the last
// 	* letter.  Eg 'nko ' -> 'Nkoo' */
//    if (unlikely ((tag & 0x0000FF00u) == 0x00002000u))
// 	 tag |= (tag >> 8) & 0x0000FF00u; /* Copy second letter to third */
//    if (unlikely ((tag & 0x000000FFu) == 0x00000020u))
// 	 tag |= (tag >> 8) & 0x000000FFu; /* Copy third letter to fourth */

//    /* Change first char to uppercase and return */
//    return (hb_script_t) (tag & ~0x20000000u);
//  }

//  static hb_tag_t
//  hb_ot_new_tag_from_script (hb_script_t script)
//  {
//    switch ((hb_tag_t) script) {
// 	 case HB_SCRIPT_BENGALI:		return HB_TAG('b','n','g','2');
// 	 case HB_SCRIPT_DEVANAGARI:		return HB_TAG('d','e','v','2');
// 	 case HB_SCRIPT_GUJARATI:		return HB_TAG('g','j','r','2');
// 	 case HB_SCRIPT_GURMUKHI:		return HB_TAG('g','u','r','2');
// 	 case HB_SCRIPT_KANNADA:		return HB_TAG('k','n','d','2');
// 	 case HB_SCRIPT_MALAYALAM:		return HB_TAG('m','l','m','2');
// 	 case HB_SCRIPT_ORIYA:		return HB_TAG('o','r','y','2');
// 	 case HB_SCRIPT_TAMIL:		return HB_TAG('t','m','l','2');
// 	 case HB_SCRIPT_TELUGU:		return HB_TAG('t','e','l','2');
// 	 case HB_SCRIPT_MYANMAR:		return HB_TAG('m','y','m','2');
//    }

//    return HB_OT_TAG_DEFAULT_SCRIPT;
//  }

//  static hb_script_t
//  hb_ot_new_tag_to_script (hb_tag_t tag)
//  {
//    switch (tag) {
// 	 case HB_TAG('b','n','g','2'):	return HB_SCRIPT_BENGALI;
// 	 case HB_TAG('d','e','v','2'):	return HB_SCRIPT_DEVANAGARI;
// 	 case HB_TAG('g','j','r','2'):	return HB_SCRIPT_GUJARATI;
// 	 case HB_TAG('g','u','r','2'):	return HB_SCRIPT_GURMUKHI;
// 	 case HB_TAG('k','n','d','2'):	return HB_SCRIPT_KANNADA;
// 	 case HB_TAG('m','l','m','2'):	return HB_SCRIPT_MALAYALAM;
// 	 case HB_TAG('o','r','y','2'):	return HB_SCRIPT_ORIYA;
// 	 case HB_TAG('t','m','l','2'):	return HB_SCRIPT_TAMIL;
// 	 case HB_TAG('t','e','l','2'):	return HB_SCRIPT_TELUGU;
// 	 case HB_TAG('m','y','m','2'):	return HB_SCRIPT_MYANMAR;
//    }

//    return HB_SCRIPT_UNKNOWN;
//  }

//  #ifndef HB_DISABLE_DEPRECATED
//  void
//  hb_ot_tags_from_script (hb_script_t  script,
// 			 hb_tag_t    *script_tag_1,
// 			 hb_tag_t    *script_tag_2)
//  {
//    unsigned int count = 2;
//    hb_tag_t tags[2];
//    hb_ot_tags_from_script_and_language (script, HB_LANGUAGE_INVALID, &count, tags, nullptr, nullptr);
//    *script_tag_1 = count > 0 ? tags[0] : HB_OT_TAG_DEFAULT_SCRIPT;
//    *script_tag_2 = count > 1 ? tags[1] : HB_OT_TAG_DEFAULT_SCRIPT;
//  }
//  #endif

//  /*
//   * Complete list at:
//   * https://docs.microsoft.com/en-us/typography/opentype/spec/scripttags
//   *
//   * Most of the script tags are the same as the ISO 15924 tag but lowercased.
//   * So we just do that, and handle the exceptional cases in a switch.
//   */

//  static void
//  hb_ot_all_tags_from_script (hb_script_t   script,
// 				 unsigned int *count /* IN/OUT */,
// 				 hb_tag_t     *tags /* OUT */)
//  {
//    unsigned int i = 0;

//    hb_tag_t new_tag = hb_ot_new_tag_from_script (script);
//    if (unlikely (new_tag != HB_OT_TAG_DEFAULT_SCRIPT))
//    {
// 	 /* HB_SCRIPT_MYANMAR maps to 'mym2', but there is no 'mym3'. */
// 	 if (new_tag != HB_TAG('m','y','m','2'))
// 	   tags[i++] = new_tag | '3';
// 	 if (*count > i)
// 	   tags[i++] = new_tag;
//    }

//    if (*count > i)
//    {
// 	 hb_tag_t old_tag = hb_ot_old_tag_from_script (script);
// 	 if (old_tag != HB_OT_TAG_DEFAULT_SCRIPT)
// 	   tags[i++] = old_tag;
//    }

//    *count = i;
//  }

//  /**
//   * hb_ot_tag_to_script:
//   * @tag: a script tag
//   *
//   * Converts a script tag to an #hb_script_t.
//   *
//   * Return value: The #hb_script_t corresponding to @tag.
//   *
//   **/
//  hb_script_t
//  hb_ot_tag_to_script (hb_tag_t tag)
//  {
//    unsigned char digit = tag & 0x000000FFu;
//    if (unlikely (digit == '2' || digit == '3'))
// 	 return hb_ot_new_tag_to_script (tag & 0xFFFFFF32);

//    return hb_ot_old_tag_to_script (tag);
//  }

//  /* hb_language_t */

//  static bool
//  subtag_matches (const char *lang_str,
// 		 const char *limit,
// 		 const char *subtag)
//  {
//    do {
// 	 const char *s = strstr (lang_str, subtag);
// 	 if (!s || s >= limit)
// 	   return false;
// 	 if (!ISALNUM (s[strlen (subtag)]))
// 	   return true;
// 	 lang_str = s + strlen (subtag);
//    } for (true);
//  }

//  static hb_bool_t
//  lang_matches (const char *lang_str, const char *spec)
//  {
//    unsigned int len = strlen (spec);

//    return strncmp (lang_str, spec, len) == 0 &&
// 	  (lang_str[len] == '\0' || lang_str[len] == '-');
//  }

//  struct LangTag
//  {
//    char language[4];
//    hb_tag_t tag;

//    int cmp (const char *a) const
//    {
// 	 const char *b = this->language;
// 	 unsigned int da, db;
// 	 const char *p;

// 	 p = strchr (a, '-');
// 	 da = p ? (unsigned int) (p - a) : strlen (a);

// 	 p = strchr (b, '-');
// 	 db = p ? (unsigned int) (p - b) : strlen (b);

// 	 return strncmp (a, b, hb_max (da, db));
//    }
//    int cmp (const LangTag *that) const
//    { return cmp (that->language); }
//  };

//  #include "hb-ot-tag-table.hh"

//  /* The corresponding languages IDs for the following IDs are unclear,
//   * overlap, or are architecturally weird. Needs more research. */

//  /*{"??",	{HB_TAG('B','C','R',' ')}},*/	/* Bible Cree */
//  /*{"zh?",	{HB_TAG('C','H','N',' ')}},*/	/* Chinese (seen in Microsoft fonts) */
//  /*{"ar-Syrc?",	{HB_TAG('G','A','R',' ')}},*/	/* Garshuni */
//  /*{"??",	{HB_TAG('N','G','R',' ')}},*/	/* Nagari */
//  /*{"??",	{HB_TAG('Y','I','C',' ')}},*/	/* Yi Classic */
//  /*{"zh?",	{HB_TAG('Z','H','P',' ')}},*/	/* Chinese Phonetic */

//  #ifndef HB_DISABLE_DEPRECATED
//  hb_tag_t
//  hb_ot_tag_from_language (hb_language_t language)
//  {
//    unsigned int count = 1;
//    hb_tag_t tags[1];
//    hb_ot_tags_from_script_and_language (HB_SCRIPT_UNKNOWN, language, nullptr, nullptr, &count, tags);
//    return count > 0 ? tags[0] : HB_OT_TAG_DEFAULT_LANGUAGE;
//  }
//  #endif

func hb_ot_tags_from_language(lang_str string, limit int, count *int, tags []hb_tag_t) {
	//    unsigned int tag_idx;

	// check for matches of multiple subtags.
	if hb_ot_tags_from_complex_language(lang_str, limit, count, tags) {
		return
	}

	/* Find a language matching in the first component. */
	s := strings.IndexByte(lang_str, '-')
	if s != -1 && limit >= 6 {
		extlang_end := strings.IndexByte(lang_str[s+1:], '-')
		/* If there is an extended language tag, use it. */
		ref := extlang_end - s - 1
		if extlang_end == -1 {
			ref = len(lang_str[s+1:])
		}
		if ref == 3 && isAlpha(lang_str[s+1]) {
			lang_str = lang_str[s+1:]
		}
	}

	if tag_idx := bfindLanguage(lang_str); tag_idx != -1 {
		for tag_idx != 0 && ot_languages[tag_idx].language == ot_languages[tag_idx-1].language {
			tag_idx--
		}
		var i int
		for i = 0; i < *count && tag_idx+i < len(ot_languages) &&
			ot_languages[tag_idx+i].tag != 0 &&
			ot_languages[tag_idx+i].language == ot_languages[tag_idx].language; i++ {
			tags[i] = ot_languages[tag_idx+i].tag
		}
		*count = i
		return
	}

	if s == -1 {
		s = len(lang_str)
	}
	if s == 3 {
		// assume it's ISO-639-3 and upper-case and use it.
		tags[0] = newTag(lang_str[0], lang_str[1], lang_str[2], ' ') & ^0x20202000
		*count = 1
		return
	}

	*count = 0
}

func parse_private_use_subtag(private_use_subtag string, count *int,
	tags []hb_tag_t, prefix string, normalize func(byte) byte) bool {

	s := strings.Index(private_use_subtag, prefix)
	if s == -1 {
		return false
	}

	var tag [4]byte
	s += len(prefix)
	if private_use_subtag[s] == '-' {
		s += 1
		nb, _ := hex.Decode(tag[:], []byte(private_use_subtag[s:]))
		if nb != 8 {
			return false
		}
	} else {
		var i int
		for ; i < 4 && isAlnum(private_use_subtag[s+i]); i++ {
			tag[i] = normalize(private_use_subtag[s+i])
		}
		if i == 0 {
			return false
		}

		for ; i < 4; i++ {
			tag[i] = ' '
		}
	}
	tags[0] = newTag(tag[0], tag[1], tag[2], tag[3])
	if (tags[0] & 0xDFDFDFDF) == HB_OT_TAG_DEFAULT_SCRIPT {
		tags[0] ^= ^truetype.Tag(0xDFDFDFDF)
	}
	*count = 1
	return true
}

/**
 * hb_ot_tags_from_script_and_language:
 * @script: an #hb_script_t to convert.
 * @language: an #hb_language_t to convert.
 * @script_count: (inout) (optional): maximum number of script tags to retrieve (IN)
 * and actual number of script tags retrieved (OUT)
 * @script_tags: (out) (optional): array of size at least @script_count to store the
 * script tag results
 * @language_count: (inout) (optional): maximum number of language tags to retrieve
 * (IN) and actual number of language tags retrieved (OUT)
 * @language_tags: (out) (optional): array of size at least @language_count to store
 * the language tag results
 *
 * Converts an #hb_script_t and an #hb_language_t to script and language tags.
 **/
func hb_ot_tags_from_script_and_language(script hb_script_t, language hb_language_t,
	script_count, language_count int) (script_tags, language_tags []hb_tag_t) {
	needs_script := true

	if language == HB_LANGUAGE_INVALID {
		if language_count && language_tags && *language_count {
			*language_count = 0
		}
	} else {
		lang_str := hb_language_to_string(language)
		limit := -1
		private_use_subtag := ""
		if lang_str[0] == 'x' && lang_str[1] == '-' {
			private_use_subtag = lang_str
		} else {
			var s int
			for s = 1; s < len(lang_str); s++ { // s index in lang_str
				if lang_str[s-1] == '-' && lang_str[s+1] == '-' {
					if lang_str[s] == 'x' {
						private_use_subtag = lang_str[s:]
						if limit == -1 {
							limit = s - 1
						}
						break
					} else if limit == -1 {
						limit = s - 1
					}
				}
			}
			if limit == -1 {
				limit = s
			}
		}

		needs_script = !parse_private_use_subtag(private_use_subtag, &script_count, script_tags, "-hbsc", toLower)
		needs_language := !parse_private_use_subtag(private_use_subtag, &language_count, language_tags, "-hbot", toUpper)

		if needs_language && language_count != 0 {
			hb_ot_tags_from_language(lang_str, limit, language_count, language_tags)
		}
	}

	if needs_script && script_count != 0 {
		hb_ot_all_tags_from_script(script, script_count, script_tags)
	}
}

//  /**
//   * hb_ot_tag_to_language:
//   * @tag: an language tag
//   *
//   * Converts a language tag to an #hb_language_t.
//   *
//   * Return value: (transfer none) (nullable):
//   * The #hb_language_t corresponding to @tag.
//   *
//   * Since: 0.9.2
//   **/
//  hb_language_t
//  hb_ot_tag_to_language (hb_tag_t tag)
//  {
//    unsigned int i;

//    if (tag == HB_OT_TAG_DEFAULT_LANGUAGE)
// 	 return nullptr;

//    {
// 	 hb_language_t disambiguated_tag = hb_ot_ambiguous_tag_to_language (tag);
// 	 if (disambiguated_tag != HB_LANGUAGE_INVALID)
// 	   return disambiguated_tag;
//    }

//    for (i = 0; i < ARRAY_LENGTH (ot_languages); i++)
// 	 if (ot_languages[i].tag == tag)
// 	   return hb_language_from_string (ot_languages[i].language, -1);

//    /* Return a custom language in the form of "x-hbot-AABBCCDD".
// 	* If it's three letters long, also guess it's ISO 639-3 and lower-case and
// 	* prepend it (if it's not a registered tag, the private use subtags will
// 	* ensure that calling hb_ot_tag_from_language on the result will still return
// 	* the same tag as the original tag).
// 	*/
//    {
// 	 char buf[20];
// 	 char *str = buf;
// 	 if (ISALPHA (tag >> 24)
// 	 && ISALPHA ((tag >> 16) & 0xFF)
// 	 && ISALPHA ((tag >> 8) & 0xFF)
// 	 && (tag & 0xFF) == ' ')
// 	 {
// 	   buf[0] = TOLOWER (tag >> 24);
// 	   buf[1] = TOLOWER ((tag >> 16) & 0xFF);
// 	   buf[2] = TOLOWER ((tag >> 8) & 0xFF);
// 	   buf[3] = '-';
// 	   str += 4;
// 	 }
// 	 snprintf (str, 16, "x-hbot-%08x", tag);
// 	 return hb_language_from_string (&*buf, -1);
//    }
//  }

//  /**
//   * hb_ot_tags_to_script_and_language:
//   * @script_tag: a script tag
//   * @language_tag: a language tag
//   * @script: (out) (optional): the #hb_script_t corresponding to @script_tag.
//   * @language: (out) (optional): the #hb_language_t corresponding to @script_tag and
//   * @language_tag.
//   *
//   * Converts a script tag and a language tag to an #hb_script_t and an
//   * #hb_language_t.
//   *
//   * Since: 2.0.0
//   **/
//  void
//  hb_ot_tags_to_script_and_language (hb_tag_t       script_tag,
// 					hb_tag_t       language_tag,
// 					hb_script_t   *script /* OUT */,
// 					hb_language_t *language /* OUT */)
//  {
//    hb_script_t script_out = hb_ot_tag_to_script (script_tag);
//    if (script)
// 	 *script = script_out;
//    if (language)
//    {
// 	 unsigned int script_count = 1;
// 	 hb_tag_t primary_script_tag[1];
// 	 hb_ot_tags_from_script_and_language (script_out,
// 					  HB_LANGUAGE_INVALID,
// 					  &script_count,
// 					  primary_script_tag,
// 					  nullptr, nullptr);
// 	 *language = hb_ot_tag_to_language (language_tag);
// 	 if (script_count == 0 || primary_script_tag[0] != script_tag)
// 	 {
// 	   unsigned char *buf;
// 	   const char *lang_str = hb_language_to_string (*language);
// 	   size_t len = strlen (lang_str);
// 	   buf = (unsigned char *) malloc (len + 16);
// 	   if (unlikely (!buf))
// 	   {
// 	 *language = nullptr;
// 	   }
// 	   else
// 	   {
// 	 int shift;
// 	 memcpy (buf, lang_str, len);
// 	 if (lang_str[0] != 'x' || lang_str[1] != '-') {
// 	   buf[len++] = '-';
// 	   buf[len++] = 'x';
// 	 }
// 	 buf[len++] = '-';
// 	 buf[len++] = 'h';
// 	 buf[len++] = 'b';
// 	 buf[len++] = 's';
// 	 buf[len++] = 'c';
// 	 buf[len++] = '-';
// 	 for (shift = 28; shift >= 0; shift -= 4)
// 	   buf[len++] = TOHEX (script_tag >> shift);
// 	 *language = hb_language_from_string ((char *) buf, len);
// 	 free (buf);
// 	   }
// 	 }
//    }
//  }
