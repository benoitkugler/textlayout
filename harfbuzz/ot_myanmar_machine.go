package harfbuzz

// Code generated with ragel -Z -o ot_myanmar_machine.go ot_myanmar_machine.rl ; sed -i '/^\/\/line/ d' ot_myanmar_machine.go ; goimports -w ot_myanmar_machine.go  DO NOT EDIT.

// ported from harfbuzz/src/hb-ot-shape-complex-myanmar-machine.rl Copyright © 2015 Mozilla Foundation. Google, Inc. Behdad Esfahbod

// myanmar_syllable_type_t
const (
	myanmarConsonantSyllable = iota
	myanmarPunctuationCluster
	myanmarBrokenCluster
	myanmarNonMyanmarCluster
)

const myanmarSyllableMachine_ex_A = 10
const myanmarSyllableMachine_ex_As = 18
const myanmarSyllableMachine_ex_C = 1
const myanmarSyllableMachine_ex_CS = 19
const myanmarSyllableMachine_ex_D = 32
const myanmarSyllableMachine_ex_D0 = 20
const myanmarSyllableMachine_ex_DB = 3
const myanmarSyllableMachine_ex_GB = 11
const myanmarSyllableMachine_ex_H = 4
const myanmarSyllableMachine_ex_IV = 2
const myanmarSyllableMachine_ex_MH = 21
const myanmarSyllableMachine_ex_MR = 22
const myanmarSyllableMachine_ex_MW = 23
const myanmarSyllableMachine_ex_MY = 24
const myanmarSyllableMachine_ex_P = 31
const myanmarSyllableMachine_ex_PT = 25
const myanmarSyllableMachine_ex_Ra = 16
const myanmarSyllableMachine_ex_V = 8
const myanmarSyllableMachine_ex_VAbv = 26
const myanmarSyllableMachine_ex_VBlw = 27
const myanmarSyllableMachine_ex_VPre = 28
const myanmarSyllableMachine_ex_VPst = 29
const myanmarSyllableMachine_ex_VS = 30
const myanmarSyllableMachine_ex_ZWJ = 6
const myanmarSyllableMachine_ex_ZWNJ = 5

var _myanmarSyllableMachine_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3,
	1, 4, 1, 5, 1, 6, 1, 7,
	1, 8, 1, 9,
}

var _myanmarSyllableMachine_key_offsets []int16 = []int16{
	0, 24, 41, 47, 50, 55, 62, 67,
	71, 81, 88, 97, 105, 108, 123, 134,
	144, 153, 161, 172, 184, 196, 210, 223,
	239, 245, 248, 253, 260, 265, 269, 279,
	286, 295, 303, 306, 323, 338, 349, 359,
	368, 376, 387, 399, 411, 425, 438, 454,
	471, 487, 509, 514,
}

var _myanmarSyllableMachine_trans_keys []byte = []byte{
	3, 4, 8, 10, 11, 16, 18, 19,
	21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 31, 32, 1, 2, 5, 6,
	3, 4, 8, 10, 18, 21, 22, 23,
	24, 25, 26, 27, 28, 29, 30, 5,
	6, 8, 18, 25, 29, 5, 6, 8,
	5, 6, 8, 25, 29, 5, 6, 3,
	8, 10, 18, 25, 5, 6, 8, 18,
	25, 5, 6, 8, 25, 5, 6, 3,
	8, 10, 18, 21, 25, 26, 29, 5,
	6, 3, 8, 10, 25, 29, 5, 6,
	3, 8, 10, 18, 25, 26, 29, 5,
	6, 3, 8, 10, 25, 26, 29, 5,
	6, 16, 1, 2, 3, 8, 10, 18,
	21, 22, 23, 24, 25, 26, 27, 28,
	29, 5, 6, 3, 8, 10, 18, 25,
	26, 27, 28, 29, 5, 6, 3, 8,
	10, 25, 26, 27, 28, 29, 5, 6,
	3, 8, 10, 25, 26, 27, 29, 5,
	6, 3, 8, 10, 25, 27, 29, 5,
	6, 3, 8, 10, 25, 26, 27, 28,
	29, 30, 5, 6, 3, 8, 10, 21,
	23, 25, 26, 27, 28, 29, 5, 6,
	3, 8, 10, 18, 21, 25, 26, 27,
	28, 29, 5, 6, 3, 8, 10, 18,
	21, 22, 23, 25, 26, 27, 28, 29,
	5, 6, 3, 8, 10, 21, 22, 23,
	25, 26, 27, 28, 29, 5, 6, 3,
	4, 8, 10, 18, 21, 22, 23, 24,
	25, 26, 27, 28, 29, 5, 6, 8,
	18, 25, 29, 5, 6, 8, 5, 6,
	8, 25, 29, 5, 6, 3, 8, 10,
	18, 25, 5, 6, 8, 18, 25, 5,
	6, 8, 25, 5, 6, 3, 8, 10,
	18, 21, 25, 26, 29, 5, 6, 3,
	8, 10, 25, 29, 5, 6, 3, 8,
	10, 18, 25, 26, 29, 5, 6, 3,
	8, 10, 25, 26, 29, 5, 6, 16,
	1, 2, 3, 4, 8, 10, 18, 21,
	22, 23, 24, 25, 26, 27, 28, 29,
	30, 5, 6, 3, 8, 10, 18, 21,
	22, 23, 24, 25, 26, 27, 28, 29,
	5, 6, 3, 8, 10, 18, 25, 26,
	27, 28, 29, 5, 6, 3, 8, 10,
	25, 26, 27, 28, 29, 5, 6, 3,
	8, 10, 25, 26, 27, 29, 5, 6,
	3, 8, 10, 25, 27, 29, 5, 6,
	3, 8, 10, 25, 26, 27, 28, 29,
	30, 5, 6, 3, 8, 10, 21, 23,
	25, 26, 27, 28, 29, 5, 6, 3,
	8, 10, 18, 21, 25, 26, 27, 28,
	29, 5, 6, 3, 8, 10, 18, 21,
	22, 23, 25, 26, 27, 28, 29, 5,
	6, 3, 8, 10, 21, 22, 23, 25,
	26, 27, 28, 29, 5, 6, 3, 4,
	8, 10, 18, 21, 22, 23, 24, 25,
	26, 27, 28, 29, 5, 6, 3, 4,
	8, 10, 18, 21, 22, 23, 24, 25,
	26, 27, 28, 29, 30, 5, 6, 3,
	4, 8, 10, 18, 21, 22, 23, 24,
	25, 26, 27, 28, 29, 5, 6, 3,
	4, 8, 10, 11, 16, 18, 21, 22,
	23, 24, 25, 26, 27, 28, 29, 30,
	32, 1, 2, 5, 6, 11, 16, 32,
	1, 2, 8,
}

var _myanmarSyllableMachine_single_lengths []byte = []byte{
	20, 15, 4, 1, 3, 5, 3, 2,
	8, 5, 7, 6, 1, 13, 9, 8,
	7, 6, 9, 10, 10, 12, 11, 14,
	4, 1, 3, 5, 3, 2, 8, 5,
	7, 6, 1, 15, 13, 9, 8, 7,
	6, 9, 10, 10, 12, 11, 14, 15,
	14, 18, 3, 1,
}

var _myanmarSyllableMachine_range_lengths []byte = []byte{
	2, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 2, 1, 0,
}

var _myanmarSyllableMachine_index_offsets []int16 = []int16{
	0, 23, 40, 46, 49, 54, 61, 66,
	70, 80, 87, 96, 104, 107, 122, 133,
	143, 152, 160, 171, 183, 195, 209, 222,
	238, 244, 247, 252, 259, 264, 268, 278,
	285, 294, 302, 305, 322, 337, 348, 358,
	367, 375, 386, 398, 410, 424, 437, 453,
	470, 486, 507, 512,
}

var _myanmarSyllableMachine_indicies []byte = []byte{
	2, 3, 5, 6, 1, 7, 8, 9,
	10, 11, 12, 13, 14, 15, 16, 17,
	18, 19, 20, 1, 1, 4, 0, 22,
	23, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 24, 21,
	25, 38, 32, 36, 24, 21, 25, 24,
	21, 25, 32, 36, 24, 21, 39, 25,
	32, 40, 32, 24, 21, 25, 40, 32,
	24, 21, 25, 32, 24, 21, 22, 25,
	26, 41, 41, 32, 42, 36, 24, 21,
	22, 25, 26, 32, 36, 24, 21, 22,
	25, 26, 41, 32, 42, 36, 24, 21,
	22, 25, 26, 32, 42, 36, 24, 21,
	1, 1, 21, 22, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35, 36,
	24, 21, 22, 25, 26, 43, 32, 33,
	34, 35, 36, 24, 21, 22, 25, 26,
	32, 33, 34, 35, 36, 24, 21, 22,
	25, 26, 32, 33, 34, 36, 24, 21,
	22, 25, 26, 32, 34, 36, 24, 21,
	22, 25, 26, 32, 33, 34, 35, 36,
	43, 24, 21, 22, 25, 26, 28, 30,
	32, 33, 34, 35, 36, 24, 21, 22,
	25, 26, 43, 28, 32, 33, 34, 35,
	36, 24, 21, 22, 25, 26, 44, 28,
	29, 30, 32, 33, 34, 35, 36, 24,
	21, 22, 25, 26, 28, 29, 30, 32,
	33, 34, 35, 36, 24, 21, 22, 23,
	25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 24, 21, 5, 47,
	14, 18, 46, 45, 5, 46, 45, 5,
	14, 18, 46, 45, 48, 5, 14, 49,
	14, 46, 45, 5, 49, 14, 46, 45,
	5, 14, 46, 45, 2, 5, 6, 50,
	50, 14, 51, 18, 46, 45, 2, 5,
	6, 14, 18, 46, 45, 2, 5, 6,
	50, 14, 51, 18, 46, 45, 2, 5,
	6, 14, 51, 18, 46, 45, 52, 52,
	45, 2, 3, 5, 6, 8, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19,
	46, 45, 2, 5, 6, 8, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 46,
	45, 2, 5, 6, 53, 14, 15, 16,
	17, 18, 46, 45, 2, 5, 6, 14,
	15, 16, 17, 18, 46, 45, 2, 5,
	6, 14, 15, 16, 18, 46, 45, 2,
	5, 6, 14, 16, 18, 46, 45, 2,
	5, 6, 14, 15, 16, 17, 18, 53,
	46, 45, 2, 5, 6, 10, 12, 14,
	15, 16, 17, 18, 46, 45, 2, 5,
	6, 53, 10, 14, 15, 16, 17, 18,
	46, 45, 2, 5, 6, 54, 10, 11,
	12, 14, 15, 16, 17, 18, 46, 45,
	2, 5, 6, 10, 11, 12, 14, 15,
	16, 17, 18, 46, 45, 2, 3, 5,
	6, 8, 10, 11, 12, 13, 14, 15,
	16, 17, 18, 46, 45, 22, 23, 25,
	26, 55, 28, 29, 30, 31, 32, 33,
	34, 35, 36, 37, 24, 21, 22, 56,
	25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 24, 21, 2, 3,
	5, 6, 1, 1, 8, 10, 11, 12,
	13, 14, 15, 16, 17, 18, 19, 1,
	1, 46, 45, 1, 1, 1, 1, 57,
	58, 57,
}

var _myanmarSyllableMachine_trans_targs []byte = []byte{
	0, 1, 24, 34, 0, 25, 31, 47,
	36, 50, 37, 42, 43, 44, 27, 39,
	40, 41, 30, 46, 51, 0, 2, 12,
	0, 3, 9, 13, 14, 19, 20, 21,
	5, 16, 17, 18, 8, 23, 4, 6,
	7, 10, 11, 15, 22, 0, 0, 26,
	28, 29, 32, 33, 35, 38, 45, 48,
	49, 0, 0,
}

var _myanmarSyllableMachine_trans_actions []byte = []byte{
	13, 0, 0, 0, 7, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 15, 0, 0,
	5, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 17, 11, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 19, 9,
}

var _myanmarSyllableMachine_to_state_actions []byte = []byte{
	1, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0,
}

var _myanmarSyllableMachine_from_state_actions []byte = []byte{
	3, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0,
}

var _myanmarSyllableMachine_eof_trans []int16 = []int16{
	0, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22,
	46, 46, 46, 46, 46, 46, 46, 46,
	46, 46, 46, 46, 46, 46, 46, 46,
	46, 46, 46, 46, 46, 46, 46, 22,
	22, 46, 58, 58,
}

const myanmarSyllableMachine_start int = 0
const myanmarSyllableMachine_first_final int = 0
const myanmarSyllableMachine_error int = -1

const myanmarSyllableMachine_en_main int = 0

func findSyllablesMyanmar(buffer *Buffer) {
	var p, ts, te, act, cs int
	info := buffer.Info

	{
		cs = myanmarSyllableMachine_start
		ts = 0
		te = 0
		act = 0
	}

	pe := len(info)
	eof := pe

	var syllableSerial uint8 = 1

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
		_acts = int(_myanmarSyllableMachine_from_state_actions[cs])
		_nacts = uint(_myanmarSyllableMachine_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _myanmarSyllableMachine_actions[_acts-1] {
			case 1:
				ts = p

			}
		}

		_keys = int(_myanmarSyllableMachine_key_offsets[cs])
		_trans = int(_myanmarSyllableMachine_index_offsets[cs])

		_klen = int(_myanmarSyllableMachine_single_lengths[cs])
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
				case (info[p].complexCategory) < _myanmarSyllableMachine_trans_keys[_mid]:
					_upper = _mid - 1
				case (info[p].complexCategory) > _myanmarSyllableMachine_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_myanmarSyllableMachine_range_lengths[cs])
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
				case (info[p].complexCategory) < _myanmarSyllableMachine_trans_keys[_mid]:
					_upper = _mid - 2
				case (info[p].complexCategory) > _myanmarSyllableMachine_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_myanmarSyllableMachine_indicies[_trans])
	_eof_trans:
		cs = int(_myanmarSyllableMachine_trans_targs[_trans])

		if _myanmarSyllableMachine_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_myanmarSyllableMachine_trans_actions[_trans])
		_nacts = uint(_myanmarSyllableMachine_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _myanmarSyllableMachine_actions[_acts-1] {
			case 2:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarConsonantSyllable, ts, te, info, &syllableSerial)
				}
			case 3:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarNonMyanmarCluster, ts, te, info, &syllableSerial)
				}
			case 4:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarPunctuationCluster, ts, te, info, &syllableSerial)
				}
			case 5:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarBrokenCluster, ts, te, info, &syllableSerial)
				}
			case 6:
				te = p + 1
				{
					foundSyllableMyanmar(myanmarNonMyanmarCluster, ts, te, info, &syllableSerial)
				}
			case 7:
				te = p
				p--
				{
					foundSyllableMyanmar(myanmarConsonantSyllable, ts, te, info, &syllableSerial)
				}
			case 8:
				te = p
				p--
				{
					foundSyllableMyanmar(myanmarBrokenCluster, ts, te, info, &syllableSerial)
				}
			case 9:
				te = p
				p--
				{
					foundSyllableMyanmar(myanmarNonMyanmarCluster, ts, te, info, &syllableSerial)
				}
			}
		}

	_again:
		_acts = int(_myanmarSyllableMachine_to_state_actions[cs])
		_nacts = uint(_myanmarSyllableMachine_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _myanmarSyllableMachine_actions[_acts-1] {
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
			if _myanmarSyllableMachine_eof_trans[cs] > 0 {
				_trans = int(_myanmarSyllableMachine_eof_trans[cs] - 1)
				goto _eof_trans
			}
		}

	}

	_ = act // needed by Ragel, but unused
}
