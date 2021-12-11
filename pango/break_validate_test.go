package pango

import (
	"fmt"
	"unicode"

	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

type charForeachFunc = func(pos int,
	wc, prevWc, nextWc rune,
	type_, prevType, nextType *unicode.RangeTable,
	attr, prevAttr, nextAttr *CharAttr,
	afterZws *bool) error

func logAttrForeach(text []rune, attrs []CharAttr, fn charForeachFunc) error {
	afterZws := false

	for i, wc := range text {
		var (
			prevWc, nextWc     rune
			prevType, nextType *unicode.RangeTable
			prevAttr           *CharAttr
		)
		type_ := ucd.LookupType(wc)
		attr, nextAttr := &attrs[i], &attrs[i+1]
		if i > 0 {
			prevWc = text[i-1]
			prevType = ucd.LookupType(prevWc)
			prevAttr = &attrs[i-1]
		}
		if i < len(text)-1 {
			nextWc = text[i+1]
			nextType = ucd.LookupType(nextWc)
		}

		if err := fn(i,
			wc, prevWc, nextWc,
			type_, prevType, nextType,
			attr, prevAttr, nextAttr, &afterZws); err != nil {
			return err
		}
	}

	return nil
}

func checkLineChar(pos int,
	wc, prevWc, nextWc rune,
	type_, prevType, nextType *unicode.RangeTable,
	attr, prevAttr, nextAttr *CharAttr,
	afterZws *bool) error {

	breakType := ucd.LookupBreakClass(wc)
	next_break_type := ucd.LookupBreakClass(nextWc)

	var prevBreakType *unicode.RangeTable
	if prevWc != 0 {
		prevBreakType = ucd.LookupBreakClass(prevWc)
	}

	*afterZws = (prevBreakType == ucd.BreakZW ||
		(prevBreakType == ucd.BreakSP && *afterZws))

	if wc == '\n' && prevWc == '\r' {
		if attr.IsLineBreak() {
			return fmt.Errorf("char %#x %d: Do not break between \\r and \\n (LB5)", wc, pos)
		}
	}

	if prevWc == 0 && wc != 0 {
		if attr.IsLineBreak() {
			return fmt.Errorf("char %#x %d: Do not break before first char (LB2)", wc, pos)
		}
	}

	if nextWc == 0 {
		if !nextAttr.IsLineBreak() {
			return fmt.Errorf("char %#x %d: Always break after the last char (LB3)", wc, pos)
		}
	}

	if prevBreakType == ucd.BreakBK {
		if !attr.IsMandatoryBreak() {
			return fmt.Errorf("char %#x %d: Always break after hard line breaks (LB4)", wc, pos)
		}
	}

	if prevBreakType == ucd.BreakCR ||
		prevBreakType == ucd.BreakLF ||
		prevBreakType == ucd.BreakNL {
		if !attr.IsMandatoryBreak() {
			return fmt.Errorf("char %#x %d: Always break after CR, LF and NL (LB5)", wc, pos)
		}
	}

	if breakType == ucd.BreakBK ||
		breakType == ucd.BreakCR ||
		breakType == ucd.BreakLF ||
		breakType == ucd.BreakNL {
		if attr.IsLineBreak() {
			return fmt.Errorf("char %#x %d: Do not break before hard line beaks (LB6)", wc, pos)
		}
	}

	if breakType == ucd.BreakSP ||
		breakType == ucd.BreakZW {
		if attr.IsLineBreak() && prevAttr != nil &&
			!attr.IsMandatoryBreak() &&
			!(nextWc != 0 && next_break_type == ucd.BreakCM) {
			return fmt.Errorf("char %#x %d: Can't break before a space unless mandatory precedes or combining mark follows (LB7)", wc, pos)
		}
	}

	if breakType != ucd.BreakZW &&
		breakType != ucd.BreakSP &&
		*afterZws {
		if !attr.IsLineBreak() {
			return fmt.Errorf("char %#x %d: Break before a char following ZWS, even if spaces intervene (LB8)", wc, pos)
		}
	}

	if breakType == ucd.BreakZWJ {
		if attr.IsLineBreak() {
			return fmt.Errorf("char %#x %d: Do not break after ZWJ (LB8a)", wc, pos)
		}
	}

	if prevBreakType == ucd.BreakWJ ||
		breakType == ucd.BreakWJ {
		if attr.IsLineBreak() {
			return fmt.Errorf("char %#x %d: Do not break before or after WJ (LB11)", wc, pos)
		}
	}

	if prevBreakType == ucd.BreakGL {
		return fmt.Errorf("char %#x %d: Do not break after GL (LB12)", wc, pos)
	}

	/* internal consistency */

	if attr.IsMandatoryBreak() && !attr.IsLineBreak() {
		return fmt.Errorf("char %#x %d: Mandatory breaks must also be marked as regular breaks", wc, pos)
	}

	return nil
}

func checkLineInvariants(text []rune, attrs []CharAttr) error {
	return logAttrForeach(text, attrs, checkLineChar)
}

func checkGraphemeInvariants(text []rune, attrs []CharAttr) error {
	return nil
}

func checkWordInvariants(text []rune, attrs []CharAttr) error {
	const (
		AFTER_START = iota
		AFTER_END
	)
	state := AFTER_END
	for i, attr := range attrs {
		/* Check that word starts and ends are alternating */
		switch state {
		case AFTER_END:
			if attr.IsWordStart() {
				if attr.IsWordEnd() {
					state = AFTER_END
				} else {
					state = AFTER_START
				}
				break
			}
			if attr.IsWordEnd() {
				return fmt.Errorf("char %d: Unexpected word end", i)
			}
		case AFTER_START:
			if attr.IsWordEnd() {
				if attr.IsWordStart() {
					state = AFTER_START
				} else {
					state = AFTER_END
				}
				break
			}
			if attr.IsWordStart() {
				return fmt.Errorf("char %d: Unexpected word start", i)
			}
		}

		/* Check that words don't end in the middle of graphemes */
		if attr.IsWordBoundary() && !attr.IsCursorPosition() {
			return fmt.Errorf("char %d: Word ends inside a grapheme", i)
		}
	}

	return nil
}

func checkSentenceInvariants(text []rune, attrs []CharAttr) error {
	const (
		AFTER_START = iota
		AFTER_END
	)
	state := AFTER_END
	for i, attr := range attrs {
		/* Check that word starts and ends are alternating */
		switch state {
		case AFTER_END:
			if attr.IsSentenceStart() {
				if attr.IsSentenceEnd() {
					state = AFTER_END
				} else {
					state = AFTER_START
				}
				break
			}
			if attr.IsSentenceEnd() {
				return fmt.Errorf("char %d: Unexpected sentence end", i)
			}

		case AFTER_START:
			if attr.IsSentenceEnd() {
				if attr.IsSentenceStart() {
					state = AFTER_START
				} else {
					state = AFTER_END
				}
				break
			}
			if attr.IsSentenceStart() {
				return fmt.Errorf("char %d: Unexpected sentence start", i)
			}
		}
	}

	return nil
}

func checkSpaceInvariants(text []rune, attrs []CharAttr) error {
	for i, attr := range attrs {
		if attr.IsExpandableSpace() && !attr.IsWhite() {
			return fmt.Errorf("char %d: Expandable space must be space", i)
		}
	}

	return nil
}

// validateLogAttrs applies sanity checks to @log_attrs.
//
// This function checks some conditions that Pango
// relies on. It is not guaranteed to be an exhaustive
// validity test. Currentlty, it checks that
//
// - There's no break before the first char
// - Mandatory breaks are line breaks
// - Line breaks are char breaks
// - Lines aren't broken between \\r and \\n
// - Lines aren't broken before a space (unless the break
//   is mandatory, or the space precedes a combining mark)
// - Lines aren't broken between two open punctuation
//   or between two close punctuation characters
// - Lines aren't broken between a letter and a quotation mark
// - Word starts and ends alternate
// - Sentence starts and ends alternate
// - Expandable spaces are spaces
// - Words don't end in the middle of graphemes
// - Sentences don't end in the middle of words
func ValidateCharacterAttributes(text []rune, attrs []CharAttr) error {
	if len(attrs) != len(text)+1 {
		return fmt.Errorf("Array has wrong length")
	}

	if err := checkLineInvariants(text, attrs); err != nil {
		return err
	}

	if err := checkGraphemeInvariants(text, attrs); err != nil {
		return err
	}

	if err := checkWordInvariants(text, attrs); err != nil {
		return err
	}

	if err := checkSentenceInvariants(text, attrs); err != nil {
		return err
	}

	if err := checkSpaceInvariants(text, attrs); err != nil {
		return err
	}

	return nil
}
