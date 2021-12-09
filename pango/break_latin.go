package pango

import "unicode"

// ported from break-latin.c

func breakLatin(text []rune, analysis *Analysis, attrs []CharAttr) {
	if analysis != nil && !analysis.Language.IsDerivedFrom("ca") {
		return
	}

	var prevWc rune
	for i, wc := range text {
		/* Catalan middle dot does not break words */
		if wc == 0x00b7 && i+1 < len(text) {
			nextWc := text[i+1]
			if unicode.ToLower(nextWc) == 'l' && unicode.ToLower(prevWc) == 'l' {
				attrs[i].setWordEnd(false)
				attrs[i+1].setWordStart(false)
			}
		}
		prevWc = wc
	}
}
