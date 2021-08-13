package pango

import (
	"unicode"

	ucd "github.com/benoitkugler/textlayout/unicodedata"
)

/*
 * Implementation of pango_emoji_iter is based on Chromium's Ragel-based
 * parser (https://chromium-review.googlesource.com/c/chromium/src/+/1264577)
 * and the resulting pango/emoji_presentation_scanner.c.
 */

const (
	kCombiningEnclosingCircleBackslashCharacter rune = 0x20E0
	kCombiningEnclosingKeycapCharacter          rune = 0x20E3
	kVariationSelector15Character               rune = 0xFE0E
	kVariationSelector16Character               rune = 0xFE0F
	kZeroWidthJoinerCharacter                   rune = 0x200D
)

type emojiScannerCategory uint8

const (
	emoji emojiScannerCategory = iota
	emojiTextPresentation
	emojiEmojiPresentation
	emojiModifierBase
	emojiModifier
	emojiVsBase
	regionalIndicator
	keycapBase
	combiningEnclosingKeycap
	combiningEnclosingCircleBackslash
	zwj
	vs15
	vs16
	tagBase
	tagSequence
	tagTerm
	kMaxEmojiScannerCategory
)

func emojiSegmentationCategory(r rune) emojiScannerCategory {
	/* Specific ones first. */
	if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || r == ' ' {
		return kMaxEmojiScannerCategory
	}

	if '0' <= r && r <= '9' {
		return keycapBase
	}

	switch r {
	case kCombiningEnclosingKeycapCharacter:
		return combiningEnclosingKeycap
	case kCombiningEnclosingCircleBackslashCharacter:
		return combiningEnclosingCircleBackslash
	case kZeroWidthJoinerCharacter:
		return zwj
	case kVariationSelector15Character:
		return vs15
	case kVariationSelector16Character:
		return vs16
	case 0x1F3F4:
		return tagBase
	case 0xE007F:
		return tagTerm
	}

	if (r >= 0xE0030 && r <= 0xE0039) || (r >= 0xE0061 && r <= 0xE007A) {
		return tagSequence
	}

	if unicode.Is(ucd.Emoji_Modifier_Base, r) {
		return emojiModifierBase
	}
	if unicode.Is(ucd.Emoji_Modifier, r) {
		return emojiModifier
	}
	if r >= 0x1F1E6 && r <= 0x1F1FF { // Regional_Indicator
		return regionalIndicator
	}
	if (r >= '0' && r <= '9') || r == '#' || r == '*' { // Emoji_Keycap_Base
		return keycapBase
	}
	if unicode.Is(ucd.Emoji_Presentation, r) {
		return emojiEmojiPresentation
	}

	if unicode.Is(ucd.Emoji, r) {
		return emojiTextPresentation
	}

	/* Ragel state machine will interpret unknown category as "any". */
	return kMaxEmojiScannerCategory
}

type emojiIter struct {
	text []rune

	types  []emojiScannerCategory
	cursor int // index into types

	textEnd    int // end of the run (index into text)
	start, end int // current position (index into text)

	isEmoji bool
}

func (iter *emojiIter) reset(text []rune, textStart, length int) {
	iter.types = make([]emojiScannerCategory, length)
	for i, p := range text[textStart : textStart+length] {
		iter.types[i] = emojiSegmentationCategory(p)
	}
	iter.start, iter.end = textStart, textStart
	iter.textEnd = textStart + length
	iter.cursor = 0
	iter.text = text
	iter.isEmoji = false
	iter.next()
}

func (iter *emojiIter) next() bool {
	if iter.end > iter.textEnd {
		return false
	}

	iter.start = iter.end

	var isEmoji bool
	cursor := iter.cursor
	oldCursor := iter.cursor
	cursor += scanEmojiPresentation(iter.types[cursor:], &isEmoji)

	for do := true; do; do = iter.isEmoji == isEmoji {
		iter.cursor = cursor
		iter.isEmoji = isEmoji

		if cursor == len(iter.types) {
			break
		}
		cursor += scanEmojiPresentation(iter.types[cursor:], &isEmoji)
	}

	iter.end = iter.start + iter.cursor - oldCursor

	return true
}
