package pango

import (
	"unicode"

	"github.com/benoitkugler/textlayout/fribidi"
	"github.com/benoitkugler/textlayout/language"
	"github.com/benoitkugler/textlayout/unicodedata"
	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

const paragraphSeparator rune = 0x2029

const (
	LineBreak CharAttr = 1 << iota
	MandatoryBreak
	CharBreak
	White
	CursorPosition
	WordStart
	WordEnd
	SentenceBoundary
	SentenceStart
	SentenceEnd
	BackspaceDeletesCharacter
	ExpandableSpace
	WordBoundary
	BreakInsertsHyphen
	BreakRemovesPreceding
)

// CharAttr is a flag storing information
// about the attributes of a single character.
// See the various IsXXX methods for more informations.
type CharAttr uint16

// IsLineBreak return true if one can break line in front of character
func (c CharAttr) IsLineBreak() bool {
	return c&LineBreak != 0
}

// IsMandatoryBreak return true if one must break line in front of character
func (c CharAttr) IsMandatoryBreak() bool {
	return c&MandatoryBreak != 0
}

// IsCharBreak returns true if one can break here when doing character wrapping.
func (c CharAttr) IsCharBreak() bool {
	return c&CharBreak != 0
}

// IsWhite checks if it is whitespace character.
func (c CharAttr) IsWhite() bool {
	return c&White != 0
}

// IsCursorPosition returns true if the cursor can appear in front of character.
// i.e. this is a grapheme boundary, or the first character
// in the text.
// This flag implements Unicode's
// http://www.unicode.org/reports/tr29/ Grapheme
// Cluster Boundaries semantics.
func (c CharAttr) IsCursorPosition() bool {
	return c&CursorPosition != 0
}

// IsWordStart checks if it is the first character in a word.
func (c CharAttr) IsWordStart() bool {
	return c&WordStart != 0
}

// IsWordEnd checks if it is the first non-word char after a word
// 	Note that in degenerate cases, you could have both IsWordStart
//  and IsWordEnd set for some character.
func (c CharAttr) IsWordEnd() bool {
	return c&WordEnd != 0
}

// is a sentence boundary.
// There are two ways to divide sentences. The first assigns all
// inter-sentence whitespace/control/format chars to some sentence,
// so all chars are in some sentence; IsSentenceBoundary denotes
// the boundaries there. The second way doesn't assign
// between-sentence spaces, etc. to any sentence, so
// IsSentenceStart/IsSentenceEnd mark the boundaries of those sentences.
func (c CharAttr) IsSentenceBoundary() bool {
	return c&SentenceBoundary != 0
}

// IsSentenceStart checks if it is the first character in a sentence.
func (c CharAttr) IsSentenceStart() bool {
	return c&SentenceStart != 0
}

// IsSentenceEnd checks if it is the first char after a sentence.
// Note that in degenerate cases, you could have both IsSentenceStart
// and IsSentenceEnd set for some character (e.g. no space after a
// period, so the next sentence starts right away).
func (c CharAttr) IsSentenceEnd() bool {
	return c&SentenceEnd != 0
}

// IsBackspaceDeletesCharacter returns true if backspace deletes one character
// rather than the entire grapheme cluster. This
// field is only meaningful on grapheme
// boundaries (where IsCursorPosition is
// set).  In some languages, the full grapheme
// (e.g.  letter + diacritics) is considered a
// unit, while in others, each decomposed
// character in the grapheme is a unit. In the
// default implementation of pangoDefaultBreak, this
// bit is set on all grapheme boundaries except
// those following Latin, Cyrillic or Greek base characters.
func (c CharAttr) IsBackspaceDeletesCharacter() bool {
	return c&BackspaceDeletesCharacter != 0
}

// IsExpandableSpace checks if it is a whitespace character that can possibly be
// expanded for justification purposes.
func (c CharAttr) IsExpandableSpace() bool {
	return c&ExpandableSpace != 0
}

// IsWordBoundary checks if it is a word boundary, as defined by UAX#29.
// More specifically, means that this is not a position in the middle
// of a word.  For example, both sides of a punctuation mark are
// considered word boundaries.  This flag is particularly useful when
// selecting text word-by-word.
// This flag implements Unicode's
// http://www.unicode.org/reports/tr29/ Word
// Boundaries semantics.
func (c CharAttr) IsWordBoundary() bool {
	return c&WordBoundary != 0
}

// IsBreakInsertsHyphen determines if, when breaking lines before this char, insert a hyphen.
func (c CharAttr) IsBreakInsertsHyphen() bool {
	return c&BreakInsertsHyphen != 0
}

// IsBreakRemovesPreceding determines if, when breaking lines before this char, remove the
// preceding char.
func (c CharAttr) IsBreakRemovesPreceding() bool {
	return c&BreakRemovesPreceding != 0
}

func (c *CharAttr) setLineBreak(is bool) {
	if is {
		*c = *c | LineBreak
	} else {
		*c = *c &^ LineBreak
	}
}

func (c *CharAttr) setMandatoryBreak(is bool) {
	if is {
		*c = *c | MandatoryBreak
	} else {
		*c = *c &^ MandatoryBreak
	}
}

func (c *CharAttr) setCharBreak(is bool) {
	if is {
		*c = *c | CharBreak
	} else {
		*c = *c &^ CharBreak
	}
}

func (c *CharAttr) setWhite(is bool) {
	if is {
		*c = *c | White
	} else {
		*c = *c &^ White
	}
}

func (c *CharAttr) setCursorPosition(is bool) {
	if is {
		*c = *c | CursorPosition
	} else {
		*c = *c &^ CursorPosition
	}
}

func (c *CharAttr) setWordStart(is bool) {
	if is {
		*c = *c | WordStart
	} else {
		*c = *c &^ WordStart
	}
}

func (c *CharAttr) setWordEnd(is bool) {
	if is {
		*c = *c | WordEnd
	} else {
		*c = *c &^ WordEnd
	}
}

func (c *CharAttr) setSentenceBoundary(is bool) {
	if is {
		*c = *c | SentenceBoundary
	} else {
		*c = *c &^ SentenceBoundary
	}
}

func (c *CharAttr) setSentenceStart(is bool) {
	if is {
		*c = *c | SentenceStart
	} else {
		*c = *c &^ SentenceStart
	}
}

func (c *CharAttr) setSentenceEnd(is bool) {
	if is {
		*c = *c | SentenceEnd
	} else {
		*c = *c &^ SentenceEnd
	}
}

func (c *CharAttr) setBackspaceDeletesCharacter(is bool) {
	if is {
		*c = *c | BackspaceDeletesCharacter
	} else {
		*c = *c &^ BackspaceDeletesCharacter
	}
}

func (c *CharAttr) setExpandableSpace(is bool) {
	if is {
		*c = *c | ExpandableSpace
	} else {
		*c = *c &^ ExpandableSpace
	}
}

func (c *CharAttr) setWordBoundary(is bool) {
	if is {
		*c = *c | WordBoundary
	} else {
		*c = *c &^ WordBoundary
	}
}

func (c *CharAttr) setBreakInsertsHyphen(is bool) {
	if is {
		*c = *c | BreakInsertsHyphen
	} else {
		*c = *c &^ BreakInsertsHyphen
	}
}

func (c *CharAttr) setBreakRemovesPreceding(is bool) {
	if is {
		*c = *c | BreakRemovesPreceding
	} else {
		*c = *c &^ BreakRemovesPreceding
	}
}

// See Grapheme_Cluster_Break Property Values table of UAX#29
type graphemeBreakType uint8

const (
	gb_Other graphemeBreakType = iota
	gb_ControlCRLF
	gb_Extend
	gb_ZWJ
	gb_Prepend
	gb_SpacingMark
	gb_InHangulSyllable /* Handles all of L, V, T, LV, LVT rules */
	/* Use state machine to handle emoji sequence */
	/* Rule GB12 and GB13 */
	gb_RI_Odd  /* Meets odd number of RI */
	gb_RI_Even /* Meets even number of RI */
)

/* See Word_Break Property Values table of UAX#29 */
type wordBreakType uint8

const (
	wb_Other wordBreakType = iota
	wb_NewlineCRLF
	wb_ExtendFormat
	wb_Katakana
	wb_Hebrew_Letter
	wb_ALetter
	wb_MidNumLet
	wb_MidLetter
	wb_MidNum
	wb_Numeric
	wb_ExtendNumLet
	wb_RI_Odd
	wb_RI_Even
	wb_WSegSpace
)

/* See Sentence_Break Property Values table of UAX#29 */
type sentenceBreakType uint8

const (
	sb_Other sentenceBreakType = iota
	sb_ExtendFormat
	sb_ParaSep
	sb_Sp
	sb_Lower
	sb_Upper
	sb_OLetter
	sb_Numeric
	sb_ATerm
	sb_SContinue
	sb_STerm
	sb_Close
	/* Rules SB8 and SB8a */
	sb_ATerm_Close_Sp
	sb_STerm_Close_Sp
)

/* Rule LB25 with Example 7 of Customization */
type lineBreakType uint8

const (
	lb_Other lineBreakType = iota
	lb_Numeric
	lb_Numeric_Close
	lb_RI_Odd
	lb_RI_Even
)

// Previously "123foo" was two words. But in UAX 29 of Unicode,
// we now don't break words between consecutive letters and numbers
type wordType uint8

const (
	wordNone wordType = iota
	wordLetters
	wordNumbers
)

type breakOpportunity uint8

const (
	break_ALREADY_HANDLED breakOpportunity = iota /* didn't use the table */
	break_PROHIBITED                              /* no break, even if spaces intervene */
	break_IF_SPACES                               /* "indirect break" (only if there are spaces) */
	break_ALLOWED                                 /* "direct break" (can always break here) */
	// TR 14 has two more break-opportunity classes,
	// "indirect break opportunity for combining marks following a space"
	// and "prohibited break for combining marks"
	// but we handle that inline in the code.
)

// ComputeCharacterAttributes computes a `CharAttr` for each character in `text`.
//
// The returned array has one `CharAttr` for each position in `text`: if
// `text` contains N characters, it has N+1 positions, including the
// last position at the end of the text.
//
// `text` should be an entire paragraph; logical attributes can't be computed without context
// (for example you need to see spaces on either side of a word to know
// the word is a word).
//
// `level` is the bidirectionel embedding level; pass -1 if unknown
func ComputeCharacterAttributes(text []rune, level fribidi.Level) []CharAttr {
	analysis := Analysis{Level: level}

	logAttrs := make([]CharAttr, len(text)+1)

	pangoDefaultBreak(text, logAttrs)

	var charOffset int
	iter := newScriptIter(text)
	for do := true; do; do = iter.next() {
		runStart, runEnd, script := iter.scriptStart, iter.scriptEnd, iter.scriptCode
		analysis.Script = script
		charsInRange := runEnd - runStart
		pangoTailorBreak(text[runStart:runEnd], &analysis, -1, logAttrs[charOffset:charOffset+charsInRange+1])
		charOffset += charsInRange
	}

	return logAttrs
}

// Apply language-specific tailoring to the breaks in
// `logAttrs`, which are assumed to have been produced
// by pangoDefaultBreak().
//
// If `offset` is not -1, it is used to apply attributes
// from `analysis` that are relevant to line breaking.
func pangoTailorBreak(text []rune, analysis *Analysis, offset int, logAttrs []CharAttr) {
	start := &logAttrs[0]
	attrBefore := *start

	if tailorBreak(text, analysis, offset, logAttrs) {
		// if tailored, we enforce some of the attrs from before
		// tailoring at the boundary

		start.setBackspaceDeletesCharacter(attrBefore.IsBackspaceDeletesCharacter())

		start.setLineBreak(start.IsLineBreak() || attrBefore.IsLineBreak())
		start.setMandatoryBreak(start.IsMandatoryBreak() || attrBefore.IsMandatoryBreak())
		start.setCursorPosition(start.IsCursorPosition() || attrBefore.IsCursorPosition())
	}
}

func tailorBreak(text []rune, analysis *Analysis, itemOffset int, attrs []CharAttr) bool {
	res := breakByScript(text, analysis, attrs)

	if itemOffset >= 0 && analysis.ExtraAttrs != nil {
		hasBreakAttrs := breakAttrs(text, analysis.ExtraAttrs, itemOffset, attrs)
		res = res || hasBreakAttrs
	}

	return res
}

func breakByScript(text []rune, analysis *Analysis, attrs []CharAttr) bool {
	switch analysis.Script {
	case language.Arabic:
		breakArabic(text, attrs)
	case language.Devanagari, language.Bengali, language.Gurmukhi, language.Gujarati,
		language.Oriya, language.Tamil, language.Telugu, language.Kannada,
		language.Malayalam, language.Sinhala:
		breakIndic(text, analysis, attrs)
	case language.Thai:
		breakThai(text, attrs)
	default:
		return false
	}

	return true
}

func removeBreaksFromRange(text []rune, logAttrs []CharAttr, startPos, endPos int) {
	// assume our range doesn't start after a hyphen or in a zws sequence
	afterZws := false
	afterHyphen := false
	for pos := startPos + 1; pos < endPos; pos++ {
		/* Mandatory breaks aren't tailorable */
		if !logAttrs[pos].IsMandatoryBreak() {
			logAttrs[pos].setLineBreak(false)
		}

		ch := text[pos]
		bt := unicodedata.LookupBreakClass(ch)

		/* Hyphens and visible word dividers */
		if afterHyphen {
			logAttrs[pos].setLineBreak(true)
		}

		afterHyphen = ch == 0x00ad || /* Soft Hyphen */
			ch == 0x05A0 || ch == 0x2010 || /* Breaking Hyphens */
			ch == 0x2012 || ch == 0x2013 ||
			ch == 0x05BE || ch == 0x0F0B || /* Visible word dividers */
			ch == 0x1361 || ch == 0x17D8 ||
			ch == 0x17DA || ch == 0x2027 ||
			ch == 0x007C

		/* ZWS sequence */
		if afterZws && bt != unicodedata.BreakSP {
			logAttrs[pos].setLineBreak(true)
		}

		afterZws = bt == unicodedata.BreakZW || (bt == unicodedata.BreakSP && afterZws)
	}
}

func posFromOffset(start, end, offset int, text []rune, logAttrs []CharAttr) (int, int) {
	var startPos, endPos int
	if start >= offset {
		startPos = start - offset
	}
	if end >= offset+len(text) {
		endPos = len(logAttrs)
	} else {
		endPos = end - offset
	}
	return startPos, endPos
}

func handleAllowBreaks(text []rune, list AttrList, offset int, logAttrs []CharAttr) bool {
	tailored := false

	iter := list.getIterator()
	for do := true; do; do = iter.next() {
		attr := iter.getByKind(ATTR_ALLOW_BREAKS)

		if attr == nil || attr.Data.(AttrInt) != 0 {
			continue
		}
		start, end := iter.StartIndex, iter.EndIndex
		startPos, endPos := posFromOffset(start, end, offset, text, logAttrs)

		for pos := startPos + 1; pos < endPos; pos++ {
			logAttrs[pos].setCharBreak(false)
		}

		removeBreaksFromRange(text, logAttrs, startPos, endPos)

		tailored = true
	}

	return tailored
}

func handleWords(text []rune, attrs AttrList, offset int, logAttrs []CharAttr) bool {
	tailored := false

	iter := attrs.getIterator()
	for do := true; do; do = iter.next() {
		attr := iter.getByKind(ATTR_WORD)
		if attr == nil {
			continue
		}

		start, end := attr.StartIndex, attr.EndIndex
		startPos, endPos := posFromOffset(start, end, offset, text, logAttrs)

		for pos := startPos + 1; pos < endPos; pos++ {
			logAttrs[pos].setWordStart(false)
			logAttrs[pos].setWordEnd(false)
			logAttrs[pos].setWordBoundary(false)
		}

		removeBreaksFromRange(text, logAttrs, startPos, endPos)

		if start >= offset {
			inWord := false
			for pos := startPos; pos >= 0; pos-- {
				if logAttrs[pos].IsWordEnd() {
					inWord = pos == startPos
					break
				}
				if pos < startPos && logAttrs[pos].IsWordStart() {
					inWord = true
					break
				}
			}
			logAttrs[startPos].setWordStart(true)
			logAttrs[startPos].setWordEnd(inWord)
			logAttrs[startPos].setWordBoundary(true)

			/* Allow line breaks before words */
			if startPos > 0 {
				logAttrs[startPos].setLineBreak(true)
			}

			tailored = true
		}

		if end < offset+len(text) {
			inWord := false
			for pos := endPos; pos < len(logAttrs); pos++ {
				if logAttrs[pos].IsWordStart() {
					inWord = pos == endPos
					break
				}
				if pos > endPos && logAttrs[pos].IsWordEnd() {
					inWord = true
					break
				}
			}
			logAttrs[endPos].setWordStart(inWord)
			logAttrs[endPos].setWordEnd(true)
			logAttrs[endPos].setWordBoundary(true)

			/* Allow line breaks before words */
			if inWord {
				logAttrs[endPos].setLineBreak(true)
			}

			tailored = true
		}
	}

	return tailored
}

func handleSentences(text []rune, attrs AttrList, offset int, logAttrs []CharAttr) bool {
	tailored := false

	iter := attrs.getIterator()
	for do := true; do; do = iter.next() {
		attr := iter.getByKind(ATTR_SENTENCE)
		if attr == nil {
			continue
		}

		start, end := attr.StartIndex, attr.EndIndex
		startPos, endPos := posFromOffset(start, end, offset, text, logAttrs)

		for pos := startPos + 1; pos < endPos; pos++ {
			logAttrs[pos].setSentenceStart(false)
			logAttrs[pos].setSentenceEnd(false)
			logAttrs[pos].setSentenceBoundary(false)
		}

		if start >= offset {
			inSentence := false
			for pos := startPos - 1; pos >= 0; pos-- {
				if logAttrs[pos].IsSentenceEnd() {
					break
				}
				if logAttrs[pos].IsSentenceStart() {
					inSentence = true
					break
				}
			}
			logAttrs[startPos].setSentenceStart(true)
			logAttrs[startPos].setSentenceEnd(inSentence)
			logAttrs[startPos].setSentenceBoundary(true)

			tailored = true
		}
		if end < offset+len(text) {
			inSentence := false
			for pos := endPos + 1; endPos < len(logAttrs); pos++ {
				if logAttrs[pos].IsSentenceStart() {
					break
				}
				if logAttrs[pos].IsSentenceEnd() {
					inSentence = true
					break
				}
			}
			logAttrs[endPos].setSentenceStart(inSentence)
			logAttrs[endPos].setSentenceEnd(true)
			logAttrs[endPos].setSentenceBoundary(true)

			tailored = true
		}
	}

	return tailored
}

func handleHyphens(text []rune, attrs AttrList, offset int, logAttrs []CharAttr) bool {
	tailored := false

	iter := attrs.getIterator()
	for do := true; do; do = iter.next() {
		attr := iter.getByKind(ATTR_INSERT_HYPHENS)
		if attr == nil || attr.Data.(AttrInt) != 0 {
			continue
		}

		startPos, endPos := posFromOffset(attr.StartIndex, attr.EndIndex, offset, text, logAttrs)

		for pos := startPos + 1; pos < endPos; pos++ {
			if !logAttrs[pos].IsBreakRemovesPreceding() {
				logAttrs[pos].setBreakInsertsHyphen(false)

				tailored = true
			}
		}
	}

	return tailored
}

func breakAttrs(text []rune, attributes AttrList, offset int, logAttrs []CharAttr) bool {
	var allowBreaks, words, sentences, hyphens AttrList

	for _, attr := range attributes {
		switch attr.Kind {
		case ATTR_ALLOW_BREAKS:
			allowBreaks.insert(attr.copy())
		case ATTR_WORD:
			words.insert(attr.copy())
		case ATTR_SENTENCE:
			sentences.insert(attr.copy())
		case ATTR_INSERT_HYPHENS:
			hyphens.insert(attr.copy())
		}
	}

	t1 := handleAllowBreaks(text, allowBreaks, offset, logAttrs)
	t2 := handleWords(text, words, offset, logAttrs)
	t3 := handleSentences(text, sentences, offset, logAttrs)
	t4 := handleHyphens(text, hyphens, offset, logAttrs)

	return t1 || t2 || t3 || t4
}

func backspaceDeleteCharacter(wc rune) bool {
	return !((wc >= 0x0020 && wc <= 0x02AF) || (wc >= 0x1E00 && wc <= 0x1EFF)) &&
		!(wc >= 0x0400 && wc <= 0x052F) &&
		!((wc >= 0x0370 && wc <= 0x3FF) || (wc >= 0x1F00 && wc <= 0x1FFF)) &&
		!(wc >= 0x3040 && wc <= 0x30FF) &&
		!(wc >= 0xAC00 && wc <= 0xD7A3) &&
		!unicode.Is(ucd.Emoji, wc)
}

func isOtherTerm(sbType sentenceBreakType) bool {
	/* not in (OLetter | Upper | Lower | ParaSep | SATerm) */
	return !(sbType == sb_OLetter ||
		sbType == sb_Upper || sbType == sb_Lower ||
		sbType == sb_ParaSep ||
		sbType == sb_ATerm || sbType == sb_STerm ||
		sbType == sb_ATerm_Close_Sp ||
		sbType == sb_STerm_Close_Sp)
}

func labelAlphabetic(breakType *unicode.RangeTable, script Script, wbType *wordBreakType) {
	if breakType != ucd.BreakSA && script != language.Hiragana {
		*wbType = wb_ALetter /* ALetter */
	}
}

/* Types of Japanese characters */
func japanese(wc rune) bool { return wc >= 0x2F00 && wc <= 0x30FF }
func kanji(wc rune) bool    { return wc >= 0x2F00 && wc <= 0x2FDF }
func hiragana(wc rune) bool { return wc >= 0x3040 && wc <= 0x309F }
func katakana(wc rune) bool { return wc >= 0x30A0 && wc <= 0x30FF }

func resolveSentenceBreakType(wc rune, sbType sentenceBreakType, type_, breakType *unicode.RangeTable) sentenceBreakType {
	if sbType == sb_Other {
		switch type_ {
		case unicode.Cc:
			if wc == '\r' || wc == '\n' {
				sbType = sb_ParaSep
			} else if wc == 0x0009 || wc == 0x000B || wc == 0x000C {
				sbType = sb_Sp
			} else if wc == 0x0085 {
				sbType = sb_ParaSep
			}
		case unicode.Zs:
			if wc == 0x0020 || wc == 0x00A0 || wc == 0x1680 ||
				(wc >= 0x2000 && wc <= 0x200A) ||
				wc == 0x202F || wc == 0x205F || wc == 0x3000 {
				sbType = sb_Sp
			}

		case unicode.Zl, unicode.Zp:
			sbType = sb_ParaSep
		case unicode.Cf, unicode.Mc, unicode.Me, unicode.Mn:
			sbType = sb_ExtendFormat /* Extend, Format */
		case unicode.Lm:
			if wc >= 0xFF9E && wc <= 0xFF9F {
				sbType = sb_ExtendFormat /* Other_Grapheme_Extend */
			}
		case unicode.Lt:
			sbType = sb_Upper
		case unicode.Pd:
			if wc == 0x002D ||
				(wc >= 0x2013 && wc <= 0x2014) ||
				(wc >= 0xFE31 && wc <= 0xFE32) ||
				wc == 0xFE58 ||
				wc == 0xFE63 ||
				wc == 0xFF0D {
				sbType = sb_SContinue
			}
		case unicode.Po:
			if wc == 0x05F3 {
				sbType = sb_OLetter
			} else if wc == 0x002E || wc == 0x2024 ||
				wc == 0xFE52 || wc == 0xFF0E {
				sbType = sb_ATerm
			}
			if wc == 0x002C ||
				wc == 0x003A ||
				wc == 0x055D ||
				(wc >= 0x060C && wc <= 0x060D) ||
				wc == 0x07F8 ||
				wc == 0x1802 ||
				wc == 0x1808 ||
				wc == 0x3001 ||
				(wc >= 0xFE10 && wc <= 0xFE11) ||
				wc == 0xFE13 ||
				(wc >= 0xFE50 && wc <= 0xFE51) ||
				wc == 0xFE55 ||
				wc == 0xFF0C ||
				wc == 0xFF1A ||
				wc == 0xFF64 {
				sbType = sb_SContinue
			}
			if unicode.Is(ucd.STerm, wc) {
				sbType = sb_STerm
			}
		}
	}

	if sbType == sb_Other {
		switch type_ {
		case unicode.Ll:
			sbType = sb_Lower
		case unicode.Lu:
			sbType = sb_Upper
		case unicode.Lt, unicode.Lm, unicode.Lo:
			sbType = sb_OLetter
		}

		if type_ == unicode.Pe || type_ == unicode.Ps || breakType == ucd.BreakQU {
			sbType = sb_Close
		}
	}

	return sbType
}

// This is the default break algorithm. It applies Unicode
// rules without language-specific tailoring.
// To avoid allocations, `attrs` must be passed, and must have a length of len(text)+1.
//
// See pangoTailorBreak() for language-specific breaks.
func pangoDefaultBreak(text []rune, attrs []CharAttr) {
	before := attrs[0]

	// The rationale for all this is in section 5.15 of the Unicode 3.0 book,
	// the line breaking stuff is also in TR14 on unicode.org
	// This is a default break implementation that should work for nearly all
	// languages. Language engines can override it optionally.

	var (
		prevWc, nextWc rune

		prevJamo          = ucd.NO_JAMO
		prevSpaceOrHyphen = false

		prevBreakType     *unicode.RangeTable
		prevPrevBreakType = ucd.BreakXX

		prevGbType              = gb_Other
		metExtendedPictographic = false

		prevPrevWbType = wb_Other
		prevWbType     = wb_Other
		prevWbI        = -1

		prevPrevSbType = sb_Other
		prevSbType     = sb_Other
		prevSbI        = -1

		prevLbType = lb_Other

		currentWordType               = wordNone
		lastWordLetter, baseCharacter rune

		lastSentenceStart, lastNonSpace = -1, -1

		almostDone, done bool
		i                int
	)

	if len(text) == 0 {
		nextWc = paragraphSeparator
		almostDone = true
	} else {
		nextWc = text[0]
	}

	nextBreakType := ucd.LookupBreakClass(nextWc)
	for i = 0; !done; i++ {
		var (
			makesHangulSyllable bool
			breakOp             breakOpportunity
			rowBreakType        *unicode.RangeTable
		)
		wc := nextWc
		breakType := nextBreakType

		if almostDone {
			// If we have already reached the end of `text`, gUtf8NextChar()
			// may not increment next
			nextWc = 0
			nextBreakType = ucd.BreakXX
			done = true
		} else {

			if i+1 >= len(text) {
				// This is how we fill in the last element (end position) of the
				// attr array - assume there's a paragraph separators off the end
				// of @text.
				nextWc = paragraphSeparator
				almostDone = true
			} else {
				nextWc = text[i+1]
			}

			nextBreakType = ucd.LookupBreakClass(nextWc)
		}

		type_ := ucd.LookupType(wc)
		jamo := ucd.Jamo(breakType)

		/* Determine wheter this forms a Hangul syllable with prev. */
		if jamo == ucd.NO_JAMO {
			makesHangulSyllable = false
		} else {
			prevEnd := ucd.HangulJamoProps[prevJamo].End
			thisStart := ucd.HangulJamoProps[jamo].Start

			/* See comments before ISJAMO */
			makesHangulSyllable = (prevEnd == thisStart) || (prevEnd+1 == thisStart)
		}

		switch type_ {
		case unicode.Zs, unicode.Zl, unicode.Zp:
			attrs[i].setWhite(true)
		case unicode.Cc:
			attrs[i].setWhite(wc == '\t' || wc == '\n' || wc == '\r' || wc == '\f')
		default:
			attrs[i].setWhite(false)
		}

		// Just few spaces have variable width. So explicitly mark them.
		attrs[i].setExpandableSpace((0x0020 == wc || 0x00A0 == wc))

		isExtendedPictographic := unicode.Is(ucd.Extended_Pictographic, wc)

		// ---- UAX#29 Grapheme Boundaries ----
		var isGraphemeBoundary bool
		{
			/* Find the GraphemeBreakType of wc */
			gbType := gb_Other
			switch type_ {
			case unicode.Cf:
				if wc == 0x200C {
					gbType = gb_Extend
					break
				} else if wc == 0x200D {
					gbType = gb_ZWJ
					break
				} else if (wc >= 0x600 && wc <= 0x605) ||
					wc == 0x6DD ||
					wc == 0x70F ||
					wc == 0x8E2 ||
					wc == 0xD4E ||
					wc == 0x110BD ||
					(wc >= 0x111C2 && wc <= 0x111C3) {
					gbType = gb_Prepend
					break
				}
				/* Tag chars */
				if wc >= 0xE0020 && wc <= 0xE00FF {
					gbType = gb_Extend
					break
				}
				fallthrough
			case unicode.Cc, unicode.Zl, unicode.Zp, unicode.Cs:
				gbType = gb_ControlCRLF
			case nil:
				/* Unassigned default ignorables */
				if (wc >= 0xFFF0 && wc <= 0xFFF8) || (wc >= 0xE0000 && wc <= 0xE0FFF) {
					gbType = gb_ControlCRLF
					break
				}
				fallthrough
			case unicode.Lo:
				if makesHangulSyllable {
					gbType = gb_InHangulSyllable
				}
			case unicode.Lm:
				if wc >= 0xFF9E && wc <= 0xFF9F {
					gbType = gb_Extend /* Other_Grapheme_Extend */
				}
			case unicode.Mc:
				gbType = gb_SpacingMark /* SpacingMark */
				if wc >= 0x0900 {
					if wc == 0x09BE || wc == 0x09D7 ||
						wc == 0x0B3E || wc == 0x0B57 || wc == 0x0BBE || wc == 0x0BD7 ||
						wc == 0x0CC2 || wc == 0x0CD5 || wc == 0x0CD6 ||
						wc == 0x0D3E || wc == 0x0D57 || wc == 0x0DCF || wc == 0x0DDF ||
						wc == 0x1D165 || (wc >= 0x1D16E && wc <= 0x1D172) {
						gbType = gb_Extend /* Other_Grapheme_Extend */
					}
				}
			case unicode.Me, unicode.Mn:
				gbType = gb_Extend /* Grapheme_Extend */
			case unicode.So:
				if wc >= 0x1F1E6 && wc <= 0x1F1FF {
					if prevGbType == gb_RI_Odd {
						gbType = gb_RI_Even
					} else {
						gbType = gb_RI_Odd
					}
				}
			case unicode.Sk:
				/* Fitzpatrick modifiers */
				if wc >= 0x1F3FB && wc <= 0x1F3FF {
					gbType = gb_Extend
				}
			}

			/* Rule GB11 */
			if metExtendedPictographic {
				if gbType == gb_Extend {
					metExtendedPictographic = true
				} else if unicode.Is(ucd.Extended_Pictographic, prevWc) && gbType == gb_ZWJ {
					metExtendedPictographic = true
				} else if prevGbType == gb_Extend && gbType == gb_ZWJ {
					metExtendedPictographic = true
				} else if prevGbType == gb_ZWJ && isExtendedPictographic {
					metExtendedPictographic = true
				} else {
					metExtendedPictographic = false
				}
			}
			/* Grapheme Cluster Boundary Rules */
			isGraphemeBoundary = true /* Rule GB999 */
			/* We apply Rules GB1 && GB2 at the end of the function */
			if wc == '\n' && prevWc == '\r' {
				isGraphemeBoundary = false /* Rule GB3 */
			} else if prevGbType == gb_ControlCRLF || gbType == gb_ControlCRLF {
				isGraphemeBoundary = true /* Rules GB4 && GB5 */
			} else if gbType == gb_InHangulSyllable {
				isGraphemeBoundary = false /* Rules GB6, GB7, GB8 */
			} else if gbType == gb_Extend {
				isGraphemeBoundary = false /* Rule GB9 */
			} else if gbType == gb_ZWJ {
				isGraphemeBoundary = false /* Rule GB9 */
			} else if gbType == gb_SpacingMark {
				isGraphemeBoundary = false /* Rule GB9a */
			} else if prevGbType == gb_Prepend {
				isGraphemeBoundary = false /* Rule GB9b */
			} else if isExtendedPictographic { /* Rule GB11 */
				if prevGbType == gb_ZWJ && metExtendedPictographic {
					isGraphemeBoundary = false
				}
			} else if prevGbType == gb_RI_Odd && gbType == gb_RI_Even {
				isGraphemeBoundary = false /* Rule GB12 && GB13 */
			}

			if isExtendedPictographic {
				metExtendedPictographic = true
			}

			attrs[i].setCursorPosition(isGraphemeBoundary)
			/* If this is a grapheme boundary, we have to decide if backspace
			 * deletes a character or the whole grapheme cluster */
			if isGraphemeBoundary {
				attrs[i].setBackspaceDeletesCharacter(backspaceDeleteCharacter(baseCharacter))

				/* Dependent Vowels for Indic language */
				if unicode.Is(ucd.IndicVirama, prevWc) || unicode.Is(ucd.IndicVowel_Dependent, prevWc) {
					attrs[i].setBackspaceDeletesCharacter(true)
				}
			} else {
				attrs[i].setBackspaceDeletesCharacter(false)
			}

			prevGbType = gbType
		}

		script := language.LookupScript(wc)
		/* ---- UAX#29 Word Boundaries ---- */
		var isWordBoundary bool
		{
			if isGraphemeBoundary || (wc >= 0x1F1E6 && wc <= 0x1F1FF) { /* Rules WB3 and WB4 */
				/* Find the WordBreakType of wc */
				wbType := wb_Other

				if script == language.Katakana {
					wbType = wb_Katakana
				}

				if script == language.Hebrew && type_ == unicode.Lo {
					wbType = wb_Hebrew_Letter
				}

				if wbType == wb_Other {
					switch wc >> 8 {
					case 0x30:
						if wc == 0x3031 || wc == 0x3032 || wc == 0x3033 || wc == 0x3034 || wc == 0x3035 ||
							wc == 0x309b || wc == 0x309c || wc == 0x30a0 || wc == 0x30fc {
							wbType = wb_Katakana /* Katakana exceptions */
						}
					case 0xFF:
						if wc == 0xFF70 {
							wbType = wb_Katakana /* Katakana exceptions */
						} else if wc >= 0xFF9E && wc <= 0xFF9F {
							wbType = wb_ExtendFormat /* Other_Grapheme_Extend */
						}
					case 0x05:
						if wc == 0x05F3 {
							wbType = wb_ALetter /* ALetter exceptions */
						}
					}
				}

				if wbType == wb_Other {
					switch breakType {
					case ucd.BreakNU:
						if wc != 0x066C {
							wbType = wb_Numeric /* Numeric */
						}
					case ucd.BreakIS:
						if wc != 0x003A && wc != 0xFE13 && wc != 0x002E {
							wbType = wb_MidNum /* MidNum */
						}
					}
				}

				if wbType == wb_Other {
					switch type_ {
					case unicode.Cc:
						if wc != 0x000D && wc != 0x000A && wc != 0x000B && wc != 0x000C && wc != 0x0085 {
							break
						}
						fallthrough
					case unicode.Zl, unicode.Zp:
						wbType = wb_NewlineCRLF /* CR, LF, Newline */
					case unicode.Cf, unicode.Mc, unicode.Me, unicode.Mn:
						wbType = wb_ExtendFormat /* Extend, Format */
					case unicode.Pc:
						wbType = wb_ExtendNumLet /* ExtendNumLet */
					case unicode.Pf, unicode.Pi:
						if wc == 0x2018 || wc == 0x2019 {
							wbType = wb_MidNumLet /* MidNumLet */
						}
					case unicode.Po:
						if wc == 0x0027 || wc == 0x002e || wc == 0x2024 ||
							wc == 0xfe52 || wc == 0xff07 || wc == 0xff0e {
							wbType = wb_MidNumLet /* MidNumLet */
						} else if wc == 0x00b7 || wc == 0x05f4 || wc == 0x2027 || wc == 0x003a || wc == 0x0387 ||
							wc == 0xfe13 || wc == 0xfe55 || wc == 0xff1a {
							wbType = wb_MidLetter /* wb_MidLetter */
						} else if wc == 0x066c ||
							wc == 0xfe50 || wc == 0xfe54 || wc == 0xff0c || wc == 0xff1b {
							wbType = wb_MidNum /* MidNum */
						}
					case unicode.So:
						if wc >= 0x24B6 && wc <= 0x24E9 { /* Other_Alphabetic */
							labelAlphabetic(breakType, script, &wbType)
						}
						if wc >= 0x1F1E6 && wc <= 0x1F1FF {
							if prevWbType == wb_RI_Odd {
								wbType = wb_RI_Even
							} else {
								wbType = wb_RI_Odd
							}
						}

					case unicode.Lo, unicode.Nl:
						if wc == 0x3006 || wc == 0x3007 ||
							(wc >= 0x3021 && wc <= 0x3029) ||
							(wc >= 0x3038 && wc <= 0x303A) ||
							(wc >= 0x3400 && wc <= 0x4DB5) ||
							(wc >= 0x4E00 && wc <= 0x9FC3) ||
							(wc >= 0xF900 && wc <= 0xFA2D) ||
							(wc >= 0xFA30 && wc <= 0xFA6A) ||
							(wc >= 0xFA70 && wc <= 0xFAD9) ||
							(wc >= 0x20000 && wc <= 0x2A6D6) ||
							(wc >= 0x2F800 && wc <= 0x2FA1D) {
							break /* ALetter exceptions: Ideographic */
						}
						labelAlphabetic(breakType, script, &wbType)
					case unicode.Ll, unicode.Lm, unicode.Lt, unicode.Lu:
						labelAlphabetic(breakType, script, &wbType)
					}
				}

				if wbType == wb_Other {
					if type_ == unicode.Zs && breakType != ucd.BreakGL {
						wbType = wb_WSegSpace
					}
				}

				/* Word Cluster Boundary Rules */

				/* We apply Rules WB1 and WB2 at the end of the function */
				if prevWbType == wb_NewlineCRLF && prevWbI+1 == i {
					/* The extra check for prevWbI is to correctly handle sequences like
					 * Newline ÷ Extend × Extend
					 * since we have not skipped ExtendFormat yet.
					 */
					isWordBoundary = true /* Rule WB3a */
				} else if wbType == wb_NewlineCRLF {
					isWordBoundary = true /* Rule WB3b */
				} else if prevWc == 0x200D && isExtendedPictographic {
					isWordBoundary = false /* Rule WB3c */
				} else if prevWbType == wb_WSegSpace &&
					wbType == wb_WSegSpace && prevWbI+1 == i {
					isWordBoundary = false /* Rule WB3d */
				} else if wbType == wb_ExtendFormat {
					isWordBoundary = false /* Rules WB4? */
				} else if (prevWbType == wb_ALetter ||
					prevWbType == wb_Hebrew_Letter ||
					prevWbType == wb_Numeric) &&
					(wbType == wb_ALetter ||
						wbType == wb_Hebrew_Letter ||
						wbType == wb_Numeric) {
					isWordBoundary = false /* Rules WB5, WB8, WB9, WB10 */
				} else if prevWbType == wb_Katakana && wbType == wb_Katakana {
					isWordBoundary = false /* Rule WB13 */
				} else if (prevWbType == wb_ALetter ||
					prevWbType == wb_Hebrew_Letter ||
					prevWbType == wb_Numeric ||
					prevWbType == wb_Katakana ||
					prevWbType == wb_ExtendNumLet) &&
					wbType == wb_ExtendNumLet {
					isWordBoundary = false /* Rule WB13a */
				} else if prevWbType == wb_ExtendNumLet &&
					(wbType == wb_ALetter ||
						wbType == wb_Hebrew_Letter ||
						wbType == wb_Numeric ||
						wbType == wb_Katakana) {
					isWordBoundary = false /* Rule WB13b */
				} else if ((prevPrevWbType == wb_ALetter ||
					prevPrevWbType == wb_Hebrew_Letter) &&
					(wbType == wb_ALetter ||
						wbType == wb_Hebrew_Letter)) &&
					(prevWbType == wb_MidLetter ||
						prevWbType == wb_MidNumLet ||
						prevWc == 0x0027) {
					attrs[prevWbI].setWordBoundary(false) /* Rule WB6 */
					isWordBoundary = false                /* Rule WB7 */
				} else if prevWbType == wb_Hebrew_Letter && wc == 0x0027 {
					isWordBoundary = false /* Rule WB7a */
				} else if prevPrevWbType == wb_Hebrew_Letter && prevWc == 0x0022 &&
					wbType == wb_Hebrew_Letter {
					attrs[prevWbI].setWordBoundary(false) /* Rule WB7b */
					isWordBoundary = false                /* Rule WB7c */
				} else if (prevPrevWbType == wb_Numeric && wbType == wb_Numeric) &&
					(prevWbType == wb_MidNum || prevWbType == wb_MidNumLet ||
						prevWc == 0x0027) {
					isWordBoundary = false                /* Rule WB11 */
					attrs[prevWbI].setWordBoundary(false) /* Rule WB12 */
				} else if prevWbType == wb_RI_Odd && wbType == wb_RI_Even {
					isWordBoundary = false /* Rule WB15 and WB16 */
				} else {
					isWordBoundary = true /* Rule WB999 */
				}

				if wbType != wb_ExtendFormat {
					prevPrevWbType = prevWbType
					prevWbType = wbType
					prevWbI = i
				}
			}
			attrs[i].setWordBoundary(isWordBoundary)
		}

		/* ---- UAX#29 Sentence Boundaries ---- */
		var isSentenceBoundary bool
		{
			if isWordBoundary || wc == '\r' || wc == '\n' { /* Rules SB3 and SB5 */
				/* Find the SentenceBreakType of wc */
				sbType := sb_Other

				if breakType == ucd.BreakNU {
					sbType = sb_Numeric /* Numeric */
				}

				sbType = resolveSentenceBreakType(wc, sbType, type_, breakType)

				/* Sentence Boundary Rules */

				/* We apply Rules SB1 and SB2 at the end of the function */
				switch {
				case wc == '\n' && prevWc == '\r':
					isSentenceBoundary = false /* Rule SB3 */
				case prevSbType == sb_ParaSep && prevSbI+1 == i:
					/* The extra check for prevSbI is to correctly handle sequences like
					 * ParaSep ÷ Extend × Extend
					 * since we have not skipped ExtendFormat yet.
					 */

					isSentenceBoundary = true /* Rule SB4 */

				case sbType == sb_ExtendFormat:
					isSentenceBoundary = false /* Rule SB5? */
				case prevSbType == sb_ATerm && sbType == sb_Numeric:
					isSentenceBoundary = false /* Rule SB6 */
				case (prevPrevSbType == sb_Upper ||
					prevPrevSbType == sb_Lower) &&
					prevSbType == sb_ATerm &&
					sbType == sb_Upper:
					isSentenceBoundary = false /* Rule SB7 */
				case prevSbType == sb_ATerm && sbType == sb_Close:
					sbType = sb_ATerm
				case prevSbType == sb_STerm && sbType == sb_Close:
					sbType = sb_STerm
				case prevSbType == sb_ATerm && sbType == sb_Sp:
					sbType = sb_ATerm_Close_Sp
				case prevSbType == sb_STerm && sbType == sb_Sp:
					sbType = sb_STerm_Close_Sp
				/* Rule SB8 */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp) && sbType == sb_Lower:
					isSentenceBoundary = false
				case (prevPrevSbType == sb_ATerm ||
					prevPrevSbType == sb_ATerm_Close_Sp) &&
					isOtherTerm(prevSbType) && sbType == sb_Lower:

					attrs[prevSbI].setSentenceBoundary(false)
					attrs[prevSbI].setSentenceEnd(false)
					lastSentenceStart = -1
					for j := prevSbI - 1; j >= 0; j-- {
						attrs[j].setSentenceEnd(false)
						if attrs[j].IsSentenceBoundary() {
							lastSentenceStart = j
							break
						}
					}
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp ||
					prevSbType == sb_STerm ||
					prevSbType == sb_STerm_Close_Sp) &&
					(sbType == sb_SContinue || sbType == sb_ATerm || sbType == sb_STerm):
					isSentenceBoundary = false /* Rule SB8a */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_STerm) &&
					(sbType == sb_Close || sbType == sb_Sp ||
						sbType == sb_ParaSep):
					isSentenceBoundary = false /* Rule SB9 */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp ||
					prevSbType == sb_STerm ||
					prevSbType == sb_STerm_Close_Sp) &&
					(sbType == sb_Sp || sbType == sb_ParaSep):
					isSentenceBoundary = false /* Rule SB10 */
				case (prevSbType == sb_ATerm ||
					prevSbType == sb_ATerm_Close_Sp ||
					prevSbType == sb_STerm ||
					prevSbType == sb_STerm_Close_Sp) &&
					sbType != sb_ParaSep:
					isSentenceBoundary = true /* Rule SB11 */
				default:
					isSentenceBoundary = false /* Rule SB998 */
				}

				if sbType != sb_ExtendFormat &&
					!((prevPrevSbType == sb_ATerm || prevPrevSbType == sb_ATerm_Close_Sp) &&
						isOtherTerm(prevSbType) &&
						isOtherTerm(sbType)) {
					prevPrevSbType = prevSbType
					prevSbType = sbType
					prevSbI = i
				}
			}

			if i == 0 || done {
				isSentenceBoundary = true /* Rules SB1 and SB2 */
			}
			attrs[i].setSentenceBoundary(isSentenceBoundary)
		}
		/* ---- Line breaking ---- */
		breakOp = break_ALREADY_HANDLED

		rowBreakType = prevBreakType
		if prevBreakType == ucd.BreakSP {
			rowBreakType = prevPrevBreakType
		}

		attrs[i].setCharBreak(false)
		attrs[i].setLineBreak(false)
		attrs[i].setMandatoryBreak(false)

		/* Rule LB1:
		assign a line breaking class to each code point of the input. */
		switch breakType {
		case ucd.BreakAI, ucd.BreakSG, ucd.BreakXX:
			breakType = ucd.BreakAL
		case ucd.BreakSA:
			if type_ == unicode.Mn || type_ == unicode.Mc {
				breakType = ucd.BreakCM
			} else {
				breakType = ucd.BreakAL
			}
		case ucd.BreakCJ:
			breakType = ucd.BreakNS
		}

		/* If it's not a grapheme boundary, it's not a line break either */
		if attrs[i].IsCursorPosition() ||
			breakType == ucd.BreakEM ||
			breakType == ucd.BreakZWJ ||
			breakType == ucd.BreakCM ||
			breakType == ucd.BreakJL ||
			breakType == ucd.BreakJV ||
			breakType == ucd.BreakJT ||
			breakType == ucd.BreakH2 ||
			breakType == ucd.BreakH3 ||
			breakType == ucd.BreakRI {

			/* Find the LineBreakType of wc */
			lbType := lb_Other

			if breakType == ucd.BreakNU {
				lbType = lb_Numeric
			}
			if breakType == ucd.BreakSY ||
				breakType == ucd.BreakIS {
				if !(prevLbType == lb_Numeric) {
					lbType = lb_Other
				}
			}

			if breakType == ucd.BreakCL ||
				breakType == ucd.BreakCP {
				if prevLbType == lb_Numeric {
					lbType = lb_Numeric_Close
				} else {
					lbType = lb_Other
				}
			}

			if breakType == ucd.BreakRI {
				if prevLbType == lb_RI_Odd {
					lbType = lb_RI_Even
				} else {
					lbType = lb_RI_Odd
				}
			}

			attrs[i].setLineBreak(true /* Rule LB31 */)
			/* Unicode doesn't specify char wrap;
			   we wrap around all chars currently. */
			if attrs[i].IsCursorPosition() {
				attrs[i].setCharBreak(true)
			}
			/* Make any necessary replacements first */
			if rowBreakType == ucd.BreakXX {
				rowBreakType = ucd.BreakAL
			}
			/* add the line break rules in reverse order to override
			   the lower priority rules. */

			/* Rule LB30 */
			if (prevBreakType == ucd.BreakAL ||
				prevBreakType == ucd.BreakHL ||
				prevBreakType == ucd.BreakNU) &&
				breakType == ucd.BreakOP {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == ucd.BreakCP &&
				(breakType == ucd.BreakAL ||
					breakType == ucd.BreakHL ||
					breakType == ucd.BreakNU) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB30a */
			if prevLbType == lb_RI_Odd && lbType == lb_RI_Even {
				breakOp = break_PROHIBITED
			}
			/* Rule LB30b */
			if prevBreakType == ucd.BreakEB &&
				breakType == ucd.BreakEM {
				breakOp = break_PROHIBITED
			}
			/* Rule LB29 */
			if prevBreakType == ucd.BreakIS &&
				(breakType == ucd.BreakAL ||
					breakType == ucd.BreakHL) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB28 */
			if (prevBreakType == ucd.BreakAL ||
				prevBreakType == ucd.BreakHL) &&
				(breakType == ucd.BreakAL ||
					breakType == ucd.BreakHL) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB27 */
			if (prevBreakType == ucd.BreakJL ||
				prevBreakType == ucd.BreakJV ||
				prevBreakType == ucd.BreakJT ||
				prevBreakType == ucd.BreakH2 ||
				prevBreakType == ucd.BreakH3) &&
				(breakType == ucd.BreakIN || breakType == ucd.BreakPO) {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == ucd.BreakPR &&
				(breakType == ucd.BreakJL ||
					breakType == ucd.BreakJV ||
					breakType == ucd.BreakJT ||
					breakType == ucd.BreakH2 ||
					breakType == ucd.BreakH3) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB26 */
			if prevBreakType == ucd.BreakJL &&
				(breakType == ucd.BreakJL ||
					breakType == ucd.BreakJV ||
					breakType == ucd.BreakH2 ||
					breakType == ucd.BreakH3) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == ucd.BreakJV ||
				prevBreakType == ucd.BreakH2) &&
				(breakType == ucd.BreakJV ||
					breakType == ucd.BreakJT) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == ucd.BreakJT ||
				prevBreakType == ucd.BreakH3) &&
				breakType == ucd.BreakJT {
				breakOp = break_PROHIBITED
			}
			/* Rule LB25 with Example 7 of Customization */
			if (prevBreakType == ucd.BreakPR ||
				prevBreakType == ucd.BreakPO) &&
				breakType == ucd.BreakNU {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == ucd.BreakPR ||
				prevBreakType == ucd.BreakPO) &&
				(breakType == ucd.BreakOP ||
					breakType == ucd.BreakHY) &&
				nextBreakType == ucd.BreakNU {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == ucd.BreakOP ||
				prevBreakType == ucd.BreakHY) &&
				breakType == ucd.BreakNU {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == ucd.BreakNU &&
				(breakType == ucd.BreakNU ||
					breakType == ucd.BreakSY ||
					breakType == ucd.BreakIS) {
				breakOp = break_PROHIBITED
			}
			if prevLbType == lb_Numeric &&
				(breakType == ucd.BreakNU ||
					breakType == ucd.BreakSY ||
					breakType == ucd.BreakIS ||
					breakType == ucd.BreakCL ||
					breakType == ucd.BreakCP) {
				breakOp = break_PROHIBITED
			}
			if (prevLbType == lb_Numeric ||
				prevLbType == lb_Numeric_Close) &&
				(breakType == ucd.BreakPO ||
					breakType == ucd.BreakPR) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB24 */
			if (prevBreakType == ucd.BreakPR ||
				prevBreakType == ucd.BreakPO) &&
				(breakType == ucd.BreakAL ||
					breakType == ucd.BreakHL) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == ucd.BreakAL ||
				prevBreakType == ucd.BreakHL) &&
				(breakType == ucd.BreakPR || breakType == ucd.BreakPO) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB23 */
			if (prevBreakType == ucd.BreakAL ||
				prevBreakType == ucd.BreakHL) &&
				breakType == ucd.BreakNU {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == ucd.BreakNU &&
				(breakType == ucd.BreakAL ||
					breakType == ucd.BreakHL) {
				breakOp = break_PROHIBITED
			}
			/* Rule LB23a */
			if prevBreakType == ucd.BreakPR &&
				(breakType == ucd.BreakID ||
					breakType == ucd.BreakEB ||
					breakType == ucd.BreakEM) {
				breakOp = break_PROHIBITED
			}
			if (prevBreakType == ucd.BreakID ||
				prevBreakType == ucd.BreakEB ||
				prevBreakType == ucd.BreakEM) &&
				breakType == ucd.BreakPO {
				breakOp = break_PROHIBITED
			}

			/* Rule LB22 */
			if breakType == ucd.BreakIN {
				if prevBreakType == ucd.BreakAL ||
					prevBreakType == ucd.BreakHL {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == ucd.BreakEX {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == ucd.BreakID ||
					prevBreakType == ucd.BreakEB ||
					prevBreakType == ucd.BreakEM {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == ucd.BreakIN {
					breakOp = break_PROHIBITED
				}
				if prevBreakType == ucd.BreakNU {
					breakOp = break_PROHIBITED
				}
			}

			if breakType == ucd.BreakBA ||
				breakType == ucd.BreakHY ||
				breakType == ucd.BreakNS ||
				prevBreakType == ucd.BreakBB {
				breakOp = break_PROHIBITED /* Rule LB21 */
			}
			if prevPrevBreakType == ucd.BreakHL &&
				(prevBreakType == ucd.BreakHY ||
					prevBreakType == ucd.BreakBA) {
				breakOp = break_PROHIBITED /* Rule LB21a */
			}
			if prevBreakType == ucd.BreakSY &&
				breakType == ucd.BreakHL {
				breakOp = break_PROHIBITED /* Rule LB21b */
			}
			if prevBreakType == ucd.BreakCB ||
				breakType == ucd.BreakCB {
				breakOp = break_ALLOWED /* Rule LB20 */
			}
			if prevBreakType == ucd.BreakQU ||
				breakType == ucd.BreakQU {
				breakOp = break_PROHIBITED /* Rule LB19 */
			}

			/* handle related rules for Space as state machine here,
			   and override the pair table result. */
			if prevBreakType == ucd.BreakSP { /* Rule LB18 */
				breakOp = break_ALLOWED
			}
			if rowBreakType == ucd.BreakB2 &&
				breakType == ucd.BreakB2 {
				breakOp = break_PROHIBITED /* Rule LB17 */
			}
			if (rowBreakType == ucd.BreakCL ||
				rowBreakType == ucd.BreakCP) &&
				breakType == ucd.BreakNS {
				breakOp = break_PROHIBITED /* Rule LB16 */
			}
			if rowBreakType == ucd.BreakQU &&
				breakType == ucd.BreakOP {
				breakOp = break_PROHIBITED /* Rule LB15 */
			}
			if rowBreakType == ucd.BreakOP {
				breakOp = break_PROHIBITED /* Rule LB14 */
			}
			/* Rule LB13 with Example 7 of Customization */
			if breakType == ucd.BreakEX {
				breakOp = break_PROHIBITED
			}
			if prevBreakType != ucd.BreakNU &&
				(breakType == ucd.BreakCL ||
					breakType == ucd.BreakCP ||
					breakType == ucd.BreakIS ||
					breakType == ucd.BreakSY) {
				breakOp = break_PROHIBITED
			}
			if prevBreakType == ucd.BreakGL {
				breakOp = break_PROHIBITED /* Rule LB12 */
			}
			if breakType == ucd.BreakGL &&
				(prevBreakType != ucd.BreakSP &&
					prevBreakType != ucd.BreakBA &&
					prevBreakType != ucd.BreakHY) {
				breakOp = break_PROHIBITED /* Rule LB12a */
			}
			if prevBreakType == ucd.BreakWJ ||
				breakType == ucd.BreakWJ {
				breakOp = break_PROHIBITED /* Rule LB11 */
			}

			/* Rule LB9 */
			if breakType == ucd.BreakCM ||
				breakType == ucd.BreakZWJ {
				if !(prevBreakType == ucd.BreakBK ||
					prevBreakType == ucd.BreakCR ||
					prevBreakType == ucd.BreakLF ||
					prevBreakType == ucd.BreakNL ||
					prevBreakType == ucd.BreakSP ||
					prevBreakType == ucd.BreakZW) {
					breakOp = break_PROHIBITED
				}
			}

			if rowBreakType == ucd.BreakZW {
				breakOp = break_ALLOWED /* Rule LB8 */
			}
			if prevWc == 0x200D {
				breakOp = break_PROHIBITED /* Rule LB8a */
			}
			if breakType == ucd.BreakSP ||
				breakType == ucd.BreakZW {
				breakOp = break_PROHIBITED /* Rule LB7 */
			}
			/* Rule LB6 */
			if breakType == ucd.BreakBK ||
				breakType == ucd.BreakCR ||
				breakType == ucd.BreakLF ||
				breakType == ucd.BreakNL {
				breakOp = break_PROHIBITED
			}
			/* Rules LB4 and LB5 */
			if prevBreakType == ucd.BreakBK ||
				(prevBreakType == ucd.BreakCR && wc != '\n') ||
				prevBreakType == ucd.BreakLF ||
				prevBreakType == ucd.BreakNL {
				attrs[i].setMandatoryBreak(true)
				breakOp = break_ALLOWED
			}

			switch breakOp {
			case break_PROHIBITED:
				/* can't break here */
				attrs[i].setLineBreak(false)
			case break_IF_SPACES:
				/* break if prev char was space */
				if prevBreakType != ucd.BreakSP {
					attrs[i].setLineBreak(false)
				}
			case break_ALLOWED:
				attrs[i].setLineBreak(true)
			case break_ALREADY_HANDLED:
			}

			/* Rule LB9 */
			if !(breakType == ucd.BreakCM ||
				breakType == ucd.BreakZWJ) {
				/* Rule LB25 with Example 7 of Customization */
				if breakType == ucd.BreakNU ||
					breakType == ucd.BreakSY ||
					breakType == ucd.BreakIS {
					if prevLbType != lb_Numeric {
						prevLbType = lbType
					} /* else don't change the prevLbType */
				} else {
					prevLbType = lbType
				}
			}
			/* else don't change the prevLbType for Rule LB9 */
		}

		if breakType != ucd.BreakSP {
			/* Rule LB9 */
			if breakType == ucd.BreakCM || breakType == ucd.BreakZWJ {
				if i == 0 /* start of text */ ||
					prevBreakType == ucd.BreakBK ||
					prevBreakType == ucd.BreakCR ||
					prevBreakType == ucd.BreakLF ||
					prevBreakType == ucd.BreakNL ||
					prevBreakType == ucd.BreakSP ||
					prevBreakType == ucd.BreakZW {
					prevBreakType = ucd.BreakAL /* Rule LB10 */
				} /* else don't change the prevBreakType for Rule LB9 */
			} else {
				prevPrevBreakType = prevBreakType
				prevBreakType = breakType
			}
			prevJamo = jamo
		} else {
			if prevBreakType != ucd.BreakSP {
				prevPrevBreakType = prevBreakType
				prevBreakType = breakType
			}
			/* else don't change the prevBreakType */
		}

		/* ---- Word breaks ---- */

		/* default to not a word start/end */
		attrs[i].setWordStart(false)
		attrs[i].setWordEnd(false)

		if currentWordType != wordNone {
			/* Check for a word end */
			switch type_ {
			case unicode.Mc, unicode.Me, unicode.Mn, unicode.Cf:
			/* nothing, we just eat these up as part of the word */
			case unicode.Ll, unicode.Lm, unicode.Lo, unicode.Lt, unicode.Lu:
				if currentWordType == wordLetters {
					/* Japanese special cases for ending the word */
					if japanese(lastWordLetter) || japanese(wc) {
						if (hiragana(lastWordLetter) &&
							!hiragana(wc)) ||
							(katakana(lastWordLetter) &&
								!(katakana(wc) || hiragana(wc))) ||
							(kanji(lastWordLetter) &&
								!(hiragana(wc) || kanji(wc))) ||
							(japanese(lastWordLetter) &&
								!japanese(wc)) ||
							(!japanese(lastWordLetter) &&
								japanese(wc)) {
							attrs[i].setWordEnd(true)
						}
					}
				}
				lastWordLetter = wc
			case unicode.Nd, unicode.Nl, unicode.No:
				lastWordLetter = wc
			default:
				/* Punctuation, control/format chars, etc. all end a word. */
				attrs[i].setWordEnd(true)
				currentWordType = wordNone
			}
		} else {
			/* Check for a word start */
			switch type_ {
			case unicode.Ll, unicode.Lm, unicode.Lo, unicode.Lt, unicode.Lu:
				currentWordType = wordLetters
				lastWordLetter = wc
				attrs[i].setWordStart(true)
			case unicode.Nd, unicode.Nl, unicode.No:
				currentWordType = wordNumbers
				lastWordLetter = wc
				attrs[i].setWordStart(true)
			default:
				/* No word here */
			}
		}

		/* ---- Sentence breaks ---- */
		{

			/* default to not a sentence start/end */
			attrs[i].setSentenceStart(false)
			attrs[i].setSentenceEnd(false)

			/* maybe start sentence */
			if lastSentenceStart == -1 && !isSentenceBoundary {
				lastSentenceStart = i - 1
			}
			/* remember last non space character position */
			if i > 0 && !attrs[i-1].IsWhite() {
				lastNonSpace = i
			}
			/* meets sentence end, mark both sentence start and end */
			if lastSentenceStart != -1 && isSentenceBoundary {
				if lastNonSpace >= lastSentenceStart {
					attrs[lastSentenceStart].setSentenceStart(true)
					attrs[lastNonSpace].setSentenceEnd(true)
				}

				lastSentenceStart = -1
				lastNonSpace = -1
			}

			/* meets space character, move sentence start */
			if lastSentenceStart != -1 && lastSentenceStart == i-1 && attrs[i-1].IsWhite() {
				lastSentenceStart++
			}
		}

		/* --- Hyphens --- */
		{
			attrs[i].setBreakInsertsHyphen(false)
			attrs[i].setBreakRemovesPreceding(false)
			var insertHyphens, spaceOrHyphen bool
			switch script {
			case language.Common, language.Han, language.Hangul, language.Hiragana, language.Katakana:
				insertHyphens = false
			default:
				insertHyphens = true
			}

			switch type_ {
			case unicode.Zl, unicode.Zp, unicode.Zs:
				spaceOrHyphen = true
			case unicode.Cc:
				if wc == '\t' || wc == '\n' || wc == '\r' || wc == '\f' {
					spaceOrHyphen = true
				}
			}

			if !spaceOrHyphen {
				if wc == '-' || /* Hyphen-minus */
					wc == 0x058a || /* Armenian hyphen */
					wc == 0x1400 || /* Canadian syllabics hyphen */
					wc == 0x1806 || /* Mongolian todo hyphen */
					wc == 0x2010 || /* Hyphen */
					wc == 0x2e17 || /* Double oblique hyphen */
					wc == 0x2e40 || /* Double hyphen */
					wc == 0x30a0 || /* Katakana-Hiragana double hyphen */
					wc == 0xfe63 || /* Small hyphen-minus */
					wc == 0xff0d /* Fullwidth hyphen-minus */ {
					spaceOrHyphen = true
				}
			}

			if attrs[i].IsWordBoundary() {
				attrs[i].setBreakInsertsHyphen(false)
			} else if prevSpaceOrHyphen {
				attrs[i].setBreakInsertsHyphen(false)
			} else if spaceOrHyphen {
				attrs[i].setBreakInsertsHyphen(false)
			} else {
				attrs[i].setBreakInsertsHyphen(insertHyphens)
			}

			if prevWc == 0x007C || /* Vertical Line */
				prevWc == 0x2027 /* Hyphenation point */ {
				attrs[i].setBreakInsertsHyphen(true)
				attrs[i].setBreakRemovesPreceding(true)
			}

			prevSpaceOrHyphen = spaceOrHyphen

		}
		prevWc = wc

		/* wc might not be a valid Unicode base character, but really all we
		 * need to know is the last non-combining character */
		if type_ != unicode.Mc &&
			type_ != unicode.Me &&
			type_ != unicode.Mn {
			baseCharacter = wc
		}
	}
	i--

	attrs[0].setCursorPosition(true) /* Rule GB1 */
	attrs[i].setCursorPosition(true /* Rule GB2 */)

	attrs[0].setWordBoundary(true) /* Rule WB1 */
	attrs[i].setWordBoundary(true /* Rule WB2 */)

	attrs[0].setLineBreak(false) /* Rule LB2 */
	attrs[i].setLineBreak(true /* Rule LB3 */)
	attrs[i].setMandatoryBreak(true /* Rule LB3 */)

	attrs[0].setLineBreak(attrs[0].IsLineBreak() || before.IsLineBreak())
	attrs[0].setMandatoryBreak(attrs[0].IsMandatoryBreak() || before.IsMandatoryBreak())
	attrs[0].setCursorPosition(attrs[0].IsCursorPosition() || before.IsCursorPosition())
}

// findParagraphBoundary locates a paragraph boundary in `text`.
//
// A boundary is caused by delimiter characters, such as
// a newline, carriage return, carriage return-newline pair,
// or Unicode paragraph separator character.
//
// The index of the run of delimiters is returned in
// `delimiter`. The index of the start
// of the paragraph (index after all delimiters) is stored
// in `start`.
//
// If no delimiters are found, both `delimiter`
// and `start` are filled with the length of `text`
// (an index one off the end).
func findParagraphBoundary(text []rune) (delimiter, start int) {
	// Note: we return indexes in the rune slice, not in the utf8 byte string,
	// diverging from the C implementation

	// Only one character has type G_UNICODE_PARAGRAPH_SEPARATOR in
	// Unicode 5.0; update the following code if that changes.

	start, delimiter = -1, -1

	var prevSep rune

	for i, p := range text {
		if prevSep == '\n' || prevSep == paragraphSeparator {
			start = i
			break
		} else if prevSep == '\r' {
			// don't break between \r and \n
			if p != '\n' {
				start = i
				break
			}
		}

		if p == '\n' || p == '\r' || p == paragraphSeparator {
			if delimiter == -1 {
				delimiter = i
			}
			prevSep = p
		} else {
			prevSep = 0
		}
	}

	if delimiter == -1 {
		delimiter = len(text)
	}
	if start == -1 {
		start = len(text)
	}

	return delimiter, start
}

// pango_attr_break applies customization from attributes to the breaks in `attrs`.
//
// The line breaks are assumed to have been produced
// by `pangoDefaultBreak` and `pangoTailorBreak`
func pango_attr_break(text []rune, attributes AttrList, offset int, attrs []CharAttr) {
	start := &attrs[0]
	attrBefore := *start

	if breakAttrs(text, attributes, offset, attrs) {
		// if tailored, we enforce some of the attrs from before
		// tailoring at the boundary

		start.setBackspaceDeletesCharacter(attrBefore.IsBackspaceDeletesCharacter())

		start.setLineBreak(start.IsLineBreak() || attrBefore.IsLineBreak())
		start.setMandatoryBreak(start.IsMandatoryBreak() || attrBefore.IsMandatoryBreak())
		start.setCursorPosition(start.IsCursorPosition() || attrBefore.IsCursorPosition())
	}
}
