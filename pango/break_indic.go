package pango

import (
	"github.com/benoitkugler/textlayout/language"
)

// ported from break-indic.c:

const (
	devRra   = 0x0931 /* 0930 + 093c */
	devQa    = 0x0958 /* 0915 + 093c */
	devYa    = 0x095F /* 092f + 003c */
	devKhha  = 0x0959
	devGhha  = 0x095A
	devZa    = 0x095B
	devDddha = 0x095C
	devRha   = 0x095D
	devFa    = 0x095E
	devYya   = 0x095F

	/* Bengali */
	/* for split matras in all brahmi based script */
	bengaliSignO  = 0x09CB /* 09c7 + 09be */
	bengaliSignAu = 0x09CC /* 09c7 + 09d7 */
	bengaliRra    = 0x09DC
	bengaliRha    = 0x09DD
	bengaliYya    = 0x09DF

	/* Gurumukhi */
	gurumukhiLla  = 0x0A33
	gurumukhiSha  = 0x0A36
	gurumukhiKhha = 0x0A59
	gurumukhiGhha = 0x0A5A
	gurumukhiZa   = 0x0A5B
	gurumukhiRra  = 0x0A5C
	gurumukhiFa   = 0x0A5E

	/* Oriya */
	oriyaAi = 0x0B48
	oriyaO  = 0x0B4B
	oriyaAu = 0x0B4C

	/* Telugu */
	teluguEe = 0x0C47
	teluguAi = 0x0C48

	/* Tamil */
	tamilO  = 0x0BCA
	tamilOo = 0x0BCB
	tamilAu = 0x0BCC

	/* Kannada */
	kndaEe = 0x0CC7
	kndaAi = 0x0CC8
	kndaO  = 0x0CCA
	kndaOo = 0x0CCB

	/* Malayalam */
	mlymO  = 0x0D4A
	mlymOo = 0x0D4B
	mlymAu = 0x0D4C
)

func isCompositeWithBrahmiNukta(c rune) bool {
	return (c >= bengaliRra && c <= bengaliYya) ||
		(c >= devQa && c <= devYa) || (c == devRra) || (c >= devKhha && c <= devYya) ||
		(c >= kndaEe && c <= kndaAi) || (c >= kndaO && c <= kndaOo) ||
		(c == tamilO) || (c == tamilOo) || (c == tamilAu) ||
		(c == teluguEe) || (c == teluguAi) ||
		(c == oriyaAi) || (c == oriyaO) || (c == oriyaAu) ||
		(c >= gurumukhiKhha && c <= gurumukhiRra) || (c == gurumukhiFa) || (c == gurumukhiLla) || (c == gurumukhiSha)
}

func isSplitMatraBrahmi(c rune) bool {
	return (c == bengaliSignO) || (c == bengaliSignAu) ||
		(c >= mlymO && c <= mlymAu)
}

func notCursorPosition(attr *CharAttr) {
	if !attr.IsMandatoryBreak() {
		attr.setCursorPosition(false)
		attr.setCharBreak(false)
		attr.setLineBreak(false)
		attr.setMandatoryBreak(false)
	}
}

func breakIndic(text []rune, analysis *Analysis, attrs []CharAttr) {
	var (
		prevWc     rune
		isConjunct bool
	)
	for i, thisWc := range text {
		if isCompositeWithBrahmiNukta(thisWc) || isSplitMatraBrahmi(thisWc) {
			attrs[i+1].setBackspaceDeletesCharacter(false)
		}

		var nextWc, nextNextWc rune
		if i+1 < len(text) {
			nextWc = text[i+1]
		}
		if i+2 < len(text) {
			nextNextWc = text[i+2]
		}

		switch analysis.Script {
		case language.Sinhala:
			/*
			* TODO: The cursor position should be based on the state table.
			*       This is the wrong place to be doing this.
			 */

			/*
			* The cursor should treat as a single glyph:
			* SINHALA CONS + 0x0DCA + 0x200D + SINHALA CONS
			* SINHALA CONS + 0x200D + 0x0DCA + SINHALA CONS
			 */
			if (thisWc == 0x0DCA && nextWc == 0x200D) || (thisWc == 0x200D && nextWc == 0x0DCA) {
				notCursorPosition(&attrs[i])
				notCursorPosition(&attrs[i+1])
				isConjunct = true
			} else if isConjunct && (prevWc == 0x200D || prevWc == 0x0DCA) &&
				thisWc >= 0x0D9A && thisWc <= 0x0DC6 {
				notCursorPosition(&attrs[i])
				isConjunct = false
			} else if !isConjunct && prevWc == 0x0DCA && thisWc != 0x200D {
				/*
				* Consonant clusters do NOT result in implicit conjuncts
				* in SINHALA orthography.
				 */
				attrs[i].setCursorPosition(true)
			}

		default:
			if prevWc != 0 && (thisWc == 0x200D || thisWc == 0x200C) {
				notCursorPosition(&attrs[i])
				if nextWc != 0 {
					notCursorPosition(&attrs[i+1])
					if (nextNextWc != 0) &&
						(nextWc == 0x09CD || /* Bengali */
							nextWc == 0x0ACD || /* Gujarati */
							nextWc == 0x094D || /* Hindi */
							nextWc == 0x0CCD || /* Kannada */
							nextWc == 0x0D4D || /* Malayalam */
							nextWc == 0x0B4D || /* Oriya */
							nextWc == 0x0A4D || /* Punjabi */
							nextWc == 0x0BCD || /* Tamil */
							nextWc == 0x0C4D) /* Telugu */ {
						notCursorPosition(&attrs[i+2])
					}
				}
			}
		}

		prevWc = thisWc
	}
}
