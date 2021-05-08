package pango

import "github.com/benoitkugler/textlayout/pango/unicodedata"

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
	EMOJI emojiScannerCategory = iota
	EMOJI_TEXT_PRESENTATION
	EMOJI_EMOJI_PRESENTATION
	EMOJI_MODIFIER_BASE
	EMOJI_MODIFIER
	EMOJI_VS_BASE
	REGIONAL_INDICATOR
	KEYCAP_BASE
	COMBINING_ENCLOSING_KEYCAP
	COMBINING_ENCLOSING_CIRCLE_BACKSLASH
	ZWJ
	VS15
	VS16
	TAG_BASE
	TAG_SEQUENCE
	TAG_TERM
	kMaxEmojiScannerCategory
)

func _pango_EmojiSegmentationCategory(r rune) emojiScannerCategory {
	/* Specific ones first. */
	switch r {
	case kCombiningEnclosingKeycapCharacter:
		return COMBINING_ENCLOSING_KEYCAP
	case kCombiningEnclosingCircleBackslashCharacter:
		return COMBINING_ENCLOSING_CIRCLE_BACKSLASH
	case kZeroWidthJoinerCharacter:
		return ZWJ
	case kVariationSelector15Character:
		return VS15
	case kVariationSelector16Character:
		return VS16
	case 0x1F3F4:
		return TAG_BASE
	}

	if (r >= 0xE0030 && r <= 0xE0039) ||
		(r >= 0xE0061 && r <= 0xE007A) {
		return TAG_SEQUENCE
	}
	if r == 0xE007F {
		return TAG_TERM
	}
	if unicodedata.IsEmojiModifierBase(r) {
		return EMOJI_MODIFIER_BASE
	}
	if unicodedata.IsEmojiModifier(r) {
		return EMOJI_MODIFIER
	}
	if r >= 0x1F1E6 && r <= 0x1F1FF { // Regional_Indicator
		return REGIONAL_INDICATOR
	}
	if (r >= '0' && r <= '9') || r == '#' || r == '*' { // Emoji_Keycap_Base
		return KEYCAP_BASE
	}
	if unicodedata.IsEmojiPresentation(r) {
		return EMOJI_EMOJI_PRESENTATION
	}
	if unicodedata.IsEmoji(r) && !unicodedata.IsEmojiPresentation(r) {
		return EMOJI_TEXT_PRESENTATION
	}
	if unicodedata.IsEmoji(r) {
		return EMOJI
	}

	/* Ragel state machine will interpret unknown category as "any". */
	return kMaxEmojiScannerCategory
}

type EmojiIter struct {
	text       []rune
	types      []emojiScannerCategory
	cursor     int // index into types
	start, end int // index into text

	isEmoji bool
}

func (iter *EmojiIter) _pango_emoji_iter_init(text []rune) {
	iter.types = make([]emojiScannerCategory, len(text))
	for i, p := range text {
		iter.types[i] = _pango_EmojiSegmentationCategory(p)
	}
	iter._pango_emoji_iter_next()
}

func (iter *EmojiIter) _pango_emoji_iter_next() bool {
	if int(iter.end) > len(iter.text) {
		return false
	}

	iter.start = iter.end

	var is_emoji bool
	cursor := iter.cursor
	old_cursor := iter.cursor
	cursor = scan_emoji_presentation(iter.types[cursor:], &is_emoji)

	do := true // do ... for
	for do {
		iter.cursor = cursor
		iter.isEmoji = is_emoji

		if int(cursor) == len(iter.text) {
			break
		}

		cursor = scan_emoji_presentation(iter.types[cursor:], &is_emoji)
		do = iter.isEmoji == is_emoji
	}

	iter.end = iter.start + iter.cursor - old_cursor

	return true
}

var (
	_emoji_presentation_actions = [...]byte{
		0, 1, 0, 1, 1, 1, 5, 1,
		6, 1, 7, 1, 8, 1, 9, 1,
		10, 1, 11, 2, 2, 3, 2, 2,
		4,
	}

	_emoji_presentation_key_offsets = [...]byte{
		0, 5, 7, 14, 18, 20, 21, 24,
		29, 30, 34, 36,
	}

	_emoji_presentation_trans_keys = [...]emojiScannerCategory{
		3, 7, 13, 0, 2, 14, 15, 2,
		3, 6, 7, 13, 0, 1, 9, 10,
		11, 12, 10, 12, 10, 4, 10, 12,
		4, 9, 10, 11, 12, 6, 9, 10,
		11, 12, 8, 10, 9, 10, 11, 12,
		14, 0,
	}

	_emoji_presentation_single_lengths = [...]byte{
		3, 2, 5, 4, 2, 1, 3, 5,
		1, 4, 2, 5,
	}

	_emoji_presentation_range_lengths = [...]byte{
		1, 0, 1, 0, 0, 0, 0, 0,
		0, 0, 0, 0,
	}

	_emoji_presentation_index_offsets = [...]byte{
		0, 5, 8, 15, 20, 23, 25, 29,
		35, 37, 42, 45,
	}

	_emoji_presentation_indicies = [...]byte{
		2, 1, 1, 1, 0, 4, 5, 3,
		7, 8, 10, 11, 12, 6, 9, 5,
		13, 14, 15, 0, 13, 15, 16, 13,
		16, 15, 13, 15, 16, 15, 5, 13,
		14, 15, 16, 5, 17, 5, 13, 14,
		18, 17, 5, 13, 16, 5, 13, 14,
		15, 4, 16, 0,
	}

	_emoji_presentation_trans_targs = [...]emojiScannerCategory{
		2, 4, 6, 2, 1, 2, 3, 3,
		7, 2, 8, 9, 11, 0, 2, 5,
		2, 2, 10,
	}

	_emoji_presentation_trans_actions = [...]byte{
		17, 19, 19, 15, 0, 7, 22, 19,
		19, 9, 0, 22, 19, 0, 5, 19,
		11, 13, 19,
	}

	_emoji_presentation_to_state_actions = [...]byte{
		0, 0, 1, 0, 0, 0, 0, 0,
		0, 0, 0, 0,
	}

	_emoji_presentation_from_state_actions = [...]byte{
		0, 0, 3, 0, 0, 0, 0, 0,
		0, 0, 0, 0,
	}

	_emoji_presentation_eof_trans = [...]byte{
		1, 4, 0, 1, 17, 17, 17, 17,
		18, 18, 17, 17,
	}
)

const (
	emoji_presentation_start                 emojiScannerCategory = 2
	emoji_presentation_en_text_and_emoji_run                      = 2
)

// pe is the end of p
// cursor is the index into p
func scan_emoji_presentation(p []emojiScannerCategory, is_emoji *bool) (cursor int) {
	// eof := pe

	//   unsigned act;
	//   int cs;

	// line 107 "emoji_presentation_scanner.c"
	cs := emoji_presentation_start
	// te = 0
	act := uint(0)

	// line 115 "emoji_presentation_scanner.c"
	var (
		te                    int // index into p
		_acts                 []byte
		_keys                 []emojiScannerCategory
		_nacts, _trans, _klen byte
	)

	if len(p) <= 1 {
		goto _test_eof
	}
_resume:
	_acts = _emoji_presentation_actions[_emoji_presentation_from_state_actions[cs]:]
	_nacts = _acts[0]
	_acts = _acts[1:]
	for _nacts > 0 { // TODO: optimise ?
		_nacts--
		switch _acts[0] {
		case 1:
			// line 1 "NONE"
			// line 134 "emoji_presentation_scanner.c"
		}
		_acts = _acts[1:]
	}

	_keys = _emoji_presentation_trans_keys[_emoji_presentation_key_offsets[cs]:]
	_trans = _emoji_presentation_index_offsets[cs]

	_klen = _emoji_presentation_single_lengths[cs]
	if _klen > 0 {
		var _lower, _mid byte
		_upper := _klen - 1
		for {
			if _upper < _lower {
				break
			}

			_mid = _lower + ((_upper - _lower) >> 1)
			if p[0] < _keys[_mid] {
				_upper = _mid - 1
			} else if p[0] > _keys[_mid] {
				_lower = _mid + 1
			} else {
				_trans += _mid
				goto _match
			}
		}
		_keys = _keys[_klen:]
		_trans += _klen
	}

	_klen = _emoji_presentation_range_lengths[cs]
	if _klen > 0 {
		var _lower, _mid byte
		_upper := (_klen << 1) - 2
		for {
			if _upper < _lower {
				break
			}

			_mid = _lower + (((_upper - _lower) >> 1) & ^byte(1))
			if p[0] < _keys[_mid] {
				_upper = _mid - 2
			} else if p[0] > _keys[_mid+1] {
				_lower = _mid + 2
			} else {
				_trans += _mid >> 1
				goto _match
			}
		}
		_trans += _klen
	}

_match:
	_trans = _emoji_presentation_indicies[_trans]
_eof_trans:
	cs = _emoji_presentation_trans_targs[_trans]

	if _emoji_presentation_trans_actions[_trans] == 0 {
		goto _again
	}

	_acts = _emoji_presentation_actions[_emoji_presentation_trans_actions[_trans]:]
	_nacts = _acts[0]
	_acts = _acts[1:]
	for _nacts > 0 {
		_nacts--
		switch _acts[0] {
		case 2:
			// line 1 "NONE"
			te = 1
		case 3:
			// line 74 "emoji_presentation_scanner.rl"
			act = 2
		case 4:
			// line 75 "emoji_presentation_scanner.rl"
			act = 3
		case 5:
			// line 73 "emoji_presentation_scanner.rl"
			te = 1
			*is_emoji = false
			return te
		case 6:
			// line 74 "emoji_presentation_scanner.rl"
			te = 1
			*is_emoji = true
			return te
		case 7:
			// line 75 "emoji_presentation_scanner.rl"
			te = 1
			*is_emoji = false
			return te
		case 8:
			// line 74 "emoji_presentation_scanner.rl"
			te = 0
			*is_emoji = true
			return te
		case 9:
			// line 75 "emoji_presentation_scanner.rl"
			te = 0
			*is_emoji = false
			return te
		case 10:
			// line 74 "emoji_presentation_scanner.rl"
			*is_emoji = true
			return te
		case 11:
			// line 1 "NONE"
			switch act {
			case 2:
				*is_emoji = true
				return te
			case 3:
				*is_emoji = false
				return te
			}
			// line 248 "emoji_presentation_scanner.c"
		}
		_acts = _acts[1:]
	}

_again:
	_acts = _emoji_presentation_actions[_emoji_presentation_to_state_actions[cs]:]
	_nacts = _acts[0]
	_acts = _acts[1:]
	for _nacts > 0 {
		_nacts--
		switch _acts[0] {
		case 0:
			// line 1 "NONE"
			// line 261 "emoji_presentation_scanner.c"
		}
		_acts = _acts[1:]
	}
	p = p[1:]
	if len(p) >= 2 {
		goto _resume
	}
_test_eof:

	if len(p) == 1 {
		if _emoji_presentation_eof_trans[cs] > 0 {
			_trans = _emoji_presentation_eof_trans[cs] - 1
			goto _eof_trans
		}
	}

	// line 94 "emoji_presentation_scanner.rl"

	/* Should not be reached. */
	*is_emoji = false
	return len(p) - 1
}
