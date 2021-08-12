package pango

// ported from break-arabic.c:

const (
	alefWithMaddaAbove = 0x0622
	yehWithHamzaAbove  = 0x0626
	alef               = 0x0627
	waw                = 0x0648
	yeh                = 0x064A
	maddahAbove        = 0x0653
	hamzaAbove         = 0x0654
	hamzaBelow         = 0x0655
)

/*
 * Arabic characters with canonical decompositions that are not just
 * ligatures.  The characters U+06C0, U+06C2, and U+06D3 are intentionally
 * excluded as they are marked as "not an independent letter" in Unicode
 * Character Database's NamesList.txt
 */
func isComposite(c rune) bool { return alefWithMaddaAbove <= c && c <= yehWithHamzaAbove }

/* If a character is the second part of a composite Arabic character with an Alef */
func isCompositeWithAleft(c rune) bool { return maddahAbove <= c && c <= hamzaBelow }

func breakArabic(text []rune, attrs []CharAttr) {
	// See http://bugzilla.gnome.org/show_bug.cgi?id=350132

	var prevWc rune
	for i, thisWc := range text {
		/*
		* Unset backspace_deletes_character for various combinations.
		*
		* A few more combinations may need to be handled here, but are not
		* handled yet, as expectations of users is not known or may differ
		* among different languages or users:
		* some letters combined with U+0658 ARABIC MARK NOON GHUNNA;
		* combinations considered one letter in Azerbaijani (waw+SUKUN and
		* FARSI_YEH+hamzaAbove); combinations of yeh and ALEF_MAKSURA with
		* hamzaBelow (Qur'anic); TATWEEL+hamzaAbove (Qur'anic).
		*
		* FIXME: Ordering these in some other way may lower the time spent here, or not.
		 */
		if isComposite(thisWc) || (prevWc == alef && isCompositeWithAleft(thisWc)) ||
			(thisWc == hamzaAbove && (prevWc == waw || prevWc == yeh)) {
			attrs[i+1].setBackspaceDeletesCharacter(false)
		}

		prevWc = thisWc
	}
}
