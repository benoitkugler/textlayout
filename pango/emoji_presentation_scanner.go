package pango

// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated with ragel -Z -o emoji_presentation_scanner.go emoji_presentation_scanner.rl ; sed -i '/^\/\/line/ d' emoji_presentation_scanner.go ; goimports -w emoji_presentation_scanner.go  DO NOT EDIT.

var _emoji_presentation_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 5, 1, 6,
	1, 7, 1, 8, 1, 9, 1, 10,
	1, 11, 2, 2, 3, 2, 2, 4,
}

var _emoji_presentation_key_offsets []byte = []byte{
	0, 5, 7, 14, 18, 20, 21, 24,
	29, 30, 34, 36,
}

var _emoji_presentation_trans_keys []byte = []byte{
	3, 7, 13, 0, 2, 14, 15, 2,
	3, 6, 7, 13, 0, 1, 9, 10,
	11, 12, 10, 12, 10, 4, 10, 12,
	4, 9, 10, 11, 12, 6, 9, 10,
	11, 12, 8, 10, 9, 10, 11, 12,
	14,
}

var _emoji_presentation_single_lengths []byte = []byte{
	3, 2, 5, 4, 2, 1, 3, 5,
	1, 4, 2, 5,
}

var _emoji_presentation_range_lengths []byte = []byte{
	1, 0, 1, 0, 0, 0, 0, 0,
	0, 0, 0, 0,
}

var _emoji_presentation_index_offsets []byte = []byte{
	0, 5, 8, 15, 20, 23, 25, 29,
	35, 37, 42, 45,
}

var _emoji_presentation_indicies []byte = []byte{
	2, 1, 1, 1, 0, 4, 5, 3,
	7, 8, 10, 11, 12, 6, 9, 5,
	13, 14, 15, 0, 13, 15, 16, 13,
	16, 15, 13, 15, 16, 15, 5, 13,
	14, 15, 16, 5, 17, 5, 13, 14,
	18, 17, 5, 13, 16, 5, 13, 14,
	15, 4, 16,
}

var _emoji_presentation_trans_targs []byte = []byte{
	2, 4, 6, 2, 1, 2, 3, 3,
	7, 2, 8, 9, 11, 0, 2, 5,
	2, 2, 10,
}

var _emoji_presentation_trans_actions []byte = []byte{
	17, 19, 19, 15, 0, 7, 22, 19,
	19, 9, 0, 22, 19, 0, 5, 19,
	11, 13, 19,
}

var _emoji_presentation_to_state_actions []byte = []byte{
	0, 0, 1, 0, 0, 0, 0, 0,
	0, 0, 0, 0,
}

var _emoji_presentation_from_state_actions []byte = []byte{
	0, 0, 3, 0, 0, 0, 0, 0,
	0, 0, 0, 0,
}

var _emoji_presentation_eof_trans []byte = []byte{
	1, 4, 0, 1, 17, 17, 17, 17,
	18, 18, 17, 17,
}

const emoji_presentation_start int = 2

func scanEmojiPresentation(data []emojiScannerCategory, isEmoji *bool) int {
	var p, ts, te, act, cs int

	pe := len(data)
	eof := pe

	{
		cs = emoji_presentation_start
		ts = 0
		te = 0
		act = 0
	}

	{
		var _klen int
		var _trans int
		var _acts int
		var _nacts uint
		var _keys int
		if p == pe {
			goto _test_eof
		}
	_resume:
		_acts = int(_emoji_presentation_from_state_actions[cs])
		_nacts = uint(_emoji_presentation_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _emoji_presentation_actions[_acts-1] {
			case 1:
				ts = p

			}
		}

		_keys = int(_emoji_presentation_key_offsets[cs])
		_trans = int(_emoji_presentation_index_offsets[cs])

		_klen = int(_emoji_presentation_single_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + _klen - 1)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + ((_upper - _lower) >> 1)
				switch {
				case (uint8(data[p])) < _emoji_presentation_trans_keys[_mid]:
					_upper = _mid - 1
				case (uint8(data[p])) > _emoji_presentation_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_emoji_presentation_range_lengths[cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + (_klen << 1) - 2)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + (((_upper - _lower) >> 1) & ^1)
				switch {
				case (uint8(data[p])) < _emoji_presentation_trans_keys[_mid]:
					_upper = _mid - 2
				case (uint8(data[p])) > _emoji_presentation_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_emoji_presentation_indicies[_trans])
	_eof_trans:
		cs = int(_emoji_presentation_trans_targs[_trans])

		if _emoji_presentation_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_emoji_presentation_trans_actions[_trans])
		_nacts = uint(_emoji_presentation_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _emoji_presentation_actions[_acts-1] {
			case 2:
				te = p + 1

			case 3:
				act = 2
			case 4:
				act = 3
			case 5:
				te = p + 1
				{
					*isEmoji = false
					return te
				}
			case 6:
				te = p + 1
				{
					*isEmoji = true
					return te
				}
			case 7:
				te = p + 1
				{
					*isEmoji = false
					return te
				}
			case 8:
				te = p
				p--
				{
					*isEmoji = true
					return te
				}
			case 9:
				te = p
				p--
				{
					*isEmoji = false
					return te
				}
			case 10:
				p = (te) - 1
				{
					*isEmoji = true
					return te
				}
			case 11:
				switch act {
				case 2:
					{
						p = (te) - 1
						*isEmoji = true
						return te
					}
				case 3:
					{
						p = (te) - 1
						*isEmoji = false
						return te
					}
				}

			}
		}

	_again:
		_acts = int(_emoji_presentation_to_state_actions[cs])
		_nacts = uint(_emoji_presentation_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _emoji_presentation_actions[_acts-1] {
			case 0:
				ts = 0

			}
		}

		p++
		if p != pe {
			goto _resume
		}
	_test_eof:
		{
		}
		if p == eof {
			if _emoji_presentation_eof_trans[cs] > 0 {
				_trans = int(_emoji_presentation_eof_trans[cs] - 1)
				goto _eof_trans
			}
		}

	}

	_, _ = act, ts // needed by Ragel, but unused

	/* Should not be reached. */
	*isEmoji = false
	return pe
}
