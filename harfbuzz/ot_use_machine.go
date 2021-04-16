package harfbuzz

// Code generated with ragel -Z -o ot_use_machine.go ot_use_machine.rl ; sed -i '/^\/\/line/ d' ot_use_machine.go ; goimports -w ot_use_machine.go  DO NOT EDIT.

// ported from harfbuzz/src/hb-ot-shape-complex-use-machine.rl Copyright © 2015 Mozilla Foundation. Google, Inc. Jonathan Kew Behdad Esfahbod

const (
	useIndependentCluster = iota
	useViramaTerminatedCluster
	useSakotTerminatedCluster
	useStandardCluster
	useNumberJoinerTerminatedCluster
	useNumeralCluster
	useSymbolCluster
	useHieroglyphCluster
	useBrokenCluster
	useNonCluster
)

const (
	useSyllableMachine_ex_B     = 1
	useSyllableMachine_ex_CMAbv = 31
	useSyllableMachine_ex_CMBlw = 32
	useSyllableMachine_ex_CS    = 43
	useSyllableMachine_ex_FAbv  = 24
	useSyllableMachine_ex_FBlw  = 25
	useSyllableMachine_ex_FMAbv = 45
	useSyllableMachine_ex_FMBlw = 46
	useSyllableMachine_ex_FMPst = 47
	useSyllableMachine_ex_FPst  = 26
	useSyllableMachine_ex_G     = 49
	useSyllableMachine_ex_GB    = 5
	useSyllableMachine_ex_H     = 12
	useSyllableMachine_ex_HN    = 13
	useSyllableMachine_ex_HVM   = 44
	useSyllableMachine_ex_J     = 50
	useSyllableMachine_ex_MAbv  = 27
	useSyllableMachine_ex_MBlw  = 28
	useSyllableMachine_ex_MPre  = 30
	useSyllableMachine_ex_MPst  = 29
	useSyllableMachine_ex_N     = 4
	useSyllableMachine_ex_O     = 0
	useSyllableMachine_ex_R     = 18
	useSyllableMachine_ex_S     = 19
	useSyllableMachine_ex_SB    = 51
	useSyllableMachine_ex_SE    = 52
	useSyllableMachine_ex_SMAbv = 41
	useSyllableMachine_ex_SMBlw = 42
	useSyllableMachine_ex_SUB   = 11
	useSyllableMachine_ex_Sk    = 48
	useSyllableMachine_ex_VAbv  = 33
	useSyllableMachine_ex_VBlw  = 34
	useSyllableMachine_ex_VMAbv = 37
	useSyllableMachine_ex_VMBlw = 38
	useSyllableMachine_ex_VMPre = 23
	useSyllableMachine_ex_VMPst = 39
	useSyllableMachine_ex_VPre  = 22
	useSyllableMachine_ex_VPst  = 35
	useSyllableMachine_ex_ZWNJ  = 14
)

var _useSyllableMachine_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3,
	1, 4, 1, 5, 1, 6, 1, 7,
	1, 8, 1, 9, 1, 10, 1, 11,
	1, 12, 1, 13, 1, 14, 1, 15,
	1, 16,
}

var _useSyllableMachine_key_offsets []int16 = []int16{
	0, 1, 2, 38, 62, 86, 87, 103,
	114, 120, 125, 129, 131, 132, 142, 151,
	159, 160, 167, 182, 196, 209, 227, 244,
	263, 286, 298, 299, 300, 326, 328, 329,
	353, 369, 380, 386, 391, 395, 397, 398,
	408, 417, 425, 432, 447, 461, 474, 492,
	509, 528, 551, 563, 564, 565, 566, 595,
	619, 621, 622, 624, 626, 629,
}

var _useSyllableMachine_trans_keys []byte = []byte{
	1, 1, 0, 1, 4, 5, 11, 12,
	13, 18, 19, 23, 24, 25, 26, 27,
	28, 30, 31, 32, 33, 34, 35, 37,
	38, 39, 41, 42, 43, 44, 45, 46,
	47, 48, 49, 51, 22, 29, 11, 12,
	23, 24, 25, 26, 27, 28, 30, 31,
	32, 33, 34, 35, 37, 38, 39, 44,
	45, 46, 47, 48, 22, 29, 11, 12,
	23, 24, 25, 26, 27, 28, 30, 33,
	34, 35, 37, 38, 39, 44, 45, 46,
	47, 48, 22, 29, 31, 32, 1, 22,
	23, 24, 25, 26, 33, 34, 35, 37,
	38, 39, 44, 45, 46, 47, 48, 23,
	24, 25, 26, 37, 38, 39, 45, 46,
	47, 48, 24, 25, 26, 45, 46, 47,
	25, 26, 45, 46, 47, 26, 45, 46,
	47, 45, 46, 46, 24, 25, 26, 37,
	38, 39, 45, 46, 47, 48, 24, 25,
	26, 38, 39, 45, 46, 47, 48, 24,
	25, 26, 39, 45, 46, 47, 48, 1,
	24, 25, 26, 45, 46, 47, 48, 23,
	24, 25, 26, 33, 34, 35, 37, 38,
	39, 44, 45, 46, 47, 48, 23, 24,
	25, 26, 34, 35, 37, 38, 39, 44,
	45, 46, 47, 48, 23, 24, 25, 26,
	35, 37, 38, 39, 44, 45, 46, 47,
	48, 22, 23, 24, 25, 26, 28, 29,
	33, 34, 35, 37, 38, 39, 44, 45,
	46, 47, 48, 22, 23, 24, 25, 26,
	29, 33, 34, 35, 37, 38, 39, 44,
	45, 46, 47, 48, 23, 24, 25, 26,
	27, 28, 33, 34, 35, 37, 38, 39,
	44, 45, 46, 47, 48, 22, 29, 11,
	12, 23, 24, 25, 26, 27, 28, 30,
	32, 33, 34, 35, 37, 38, 39, 44,
	45, 46, 47, 48, 22, 29, 1, 23,
	24, 25, 26, 37, 38, 39, 45, 46,
	47, 48, 13, 4, 11, 12, 23, 24,
	25, 26, 27, 28, 30, 31, 32, 33,
	34, 35, 37, 38, 39, 41, 42, 44,
	45, 46, 47, 48, 22, 29, 41, 42,
	42, 11, 12, 23, 24, 25, 26, 27,
	28, 30, 33, 34, 35, 37, 38, 39,
	44, 45, 46, 47, 48, 22, 29, 31,
	32, 22, 23, 24, 25, 26, 33, 34,
	35, 37, 38, 39, 44, 45, 46, 47,
	48, 23, 24, 25, 26, 37, 38, 39,
	45, 46, 47, 48, 24, 25, 26, 45,
	46, 47, 25, 26, 45, 46, 47, 26,
	45, 46, 47, 45, 46, 46, 24, 25,
	26, 37, 38, 39, 45, 46, 47, 48,
	24, 25, 26, 38, 39, 45, 46, 47,
	48, 24, 25, 26, 39, 45, 46, 47,
	48, 24, 25, 26, 45, 46, 47, 48,
	23, 24, 25, 26, 33, 34, 35, 37,
	38, 39, 44, 45, 46, 47, 48, 23,
	24, 25, 26, 34, 35, 37, 38, 39,
	44, 45, 46, 47, 48, 23, 24, 25,
	26, 35, 37, 38, 39, 44, 45, 46,
	47, 48, 22, 23, 24, 25, 26, 28,
	29, 33, 34, 35, 37, 38, 39, 44,
	45, 46, 47, 48, 22, 23, 24, 25,
	26, 29, 33, 34, 35, 37, 38, 39,
	44, 45, 46, 47, 48, 23, 24, 25,
	26, 27, 28, 33, 34, 35, 37, 38,
	39, 44, 45, 46, 47, 48, 22, 29,
	11, 12, 23, 24, 25, 26, 27, 28,
	30, 32, 33, 34, 35, 37, 38, 39,
	44, 45, 46, 47, 48, 22, 29, 1,
	23, 24, 25, 26, 37, 38, 39, 45,
	46, 47, 48, 1, 4, 13, 1, 5,
	11, 12, 13, 23, 24, 25, 26, 27,
	28, 30, 31, 32, 33, 34, 35, 37,
	38, 39, 41, 42, 44, 45, 46, 47,
	48, 22, 29, 11, 12, 23, 24, 25,
	26, 27, 28, 30, 31, 32, 33, 34,
	35, 37, 38, 39, 44, 45, 46, 47,
	48, 22, 29, 41, 42, 42, 1, 5,
	50, 52, 49, 50, 52, 49, 51,
}

var _useSyllableMachine_single_lengths []byte = []byte{
	1, 1, 34, 22, 20, 1, 16, 11,
	6, 5, 4, 2, 1, 10, 9, 8,
	1, 7, 15, 14, 13, 18, 17, 17,
	21, 12, 1, 1, 24, 2, 1, 20,
	16, 11, 6, 5, 4, 2, 1, 10,
	9, 8, 7, 15, 14, 13, 18, 17,
	17, 21, 12, 1, 1, 1, 27, 22,
	2, 1, 2, 2, 3, 2,
}

var _useSyllableMachine_range_lengths []byte = []byte{
	0, 0, 1, 1, 2, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 1,
	1, 0, 0, 0, 1, 0, 0, 2,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 1, 1,
	0, 0, 0, 0, 0, 0,
}

var _useSyllableMachine_index_offsets []int16 = []int16{
	0, 2, 4, 40, 64, 87, 89, 106,
	118, 125, 131, 136, 139, 141, 152, 162,
	171, 173, 181, 197, 212, 226, 245, 263,
	282, 305, 318, 320, 322, 348, 351, 353,
	376, 393, 405, 412, 418, 423, 426, 428,
	439, 449, 458, 466, 482, 497, 511, 530,
	548, 567, 590, 603, 605, 607, 609, 638,
	662, 665, 667, 670, 673, 677,
}

var _useSyllableMachine_indicies []byte = []byte{
	1, 0, 2, 0, 3, 4, 6, 7,
	1, 8, 9, 10, 11, 13, 14, 15,
	16, 17, 18, 19, 20, 21, 22, 23,
	24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 8, 35, 36, 12, 5,
	38, 39, 41, 42, 43, 44, 45, 46,
	47, 4, 48, 49, 50, 51, 52, 53,
	54, 55, 56, 57, 58, 39, 40, 37,
	38, 39, 41, 42, 43, 44, 45, 46,
	47, 49, 50, 51, 52, 53, 54, 55,
	56, 57, 58, 39, 40, 48, 37, 38,
	59, 40, 41, 42, 43, 44, 49, 50,
	51, 52, 53, 54, 41, 56, 57, 58,
	60, 37, 41, 42, 43, 44, 52, 53,
	54, 56, 57, 58, 60, 37, 42, 43,
	44, 56, 57, 58, 37, 43, 44, 56,
	57, 58, 37, 44, 56, 57, 58, 37,
	56, 57, 37, 57, 37, 42, 43, 44,
	52, 53, 54, 56, 57, 58, 60, 37,
	42, 43, 44, 53, 54, 56, 57, 58,
	60, 37, 42, 43, 44, 54, 56, 57,
	58, 60, 37, 62, 61, 42, 43, 44,
	56, 57, 58, 60, 37, 41, 42, 43,
	44, 49, 50, 51, 52, 53, 54, 41,
	56, 57, 58, 60, 37, 41, 42, 43,
	44, 50, 51, 52, 53, 54, 41, 56,
	57, 58, 60, 37, 41, 42, 43, 44,
	51, 52, 53, 54, 41, 56, 57, 58,
	60, 37, 40, 41, 42, 43, 44, 46,
	40, 49, 50, 51, 52, 53, 54, 41,
	56, 57, 58, 60, 37, 40, 41, 42,
	43, 44, 40, 49, 50, 51, 52, 53,
	54, 41, 56, 57, 58, 60, 37, 41,
	42, 43, 44, 45, 46, 49, 50, 51,
	52, 53, 54, 41, 56, 57, 58, 60,
	40, 37, 38, 39, 41, 42, 43, 44,
	45, 46, 47, 48, 49, 50, 51, 52,
	53, 54, 55, 56, 57, 58, 39, 40,
	37, 38, 41, 42, 43, 44, 52, 53,
	54, 56, 57, 58, 60, 59, 64, 63,
	6, 65, 38, 39, 41, 42, 43, 44,
	45, 46, 47, 4, 48, 49, 50, 51,
	52, 53, 54, 11, 66, 55, 56, 57,
	58, 39, 40, 37, 11, 66, 67, 66,
	67, 1, 69, 13, 14, 15, 16, 17,
	18, 19, 22, 23, 24, 25, 26, 27,
	31, 32, 33, 34, 69, 12, 21, 68,
	12, 13, 14, 15, 16, 22, 23, 24,
	25, 26, 27, 13, 32, 33, 34, 70,
	68, 13, 14, 15, 16, 25, 26, 27,
	32, 33, 34, 70, 68, 14, 15, 16,
	32, 33, 34, 68, 15, 16, 32, 33,
	34, 68, 16, 32, 33, 34, 68, 32,
	33, 68, 33, 68, 14, 15, 16, 25,
	26, 27, 32, 33, 34, 70, 68, 14,
	15, 16, 26, 27, 32, 33, 34, 70,
	68, 14, 15, 16, 27, 32, 33, 34,
	70, 68, 14, 15, 16, 32, 33, 34,
	70, 68, 13, 14, 15, 16, 22, 23,
	24, 25, 26, 27, 13, 32, 33, 34,
	70, 68, 13, 14, 15, 16, 23, 24,
	25, 26, 27, 13, 32, 33, 34, 70,
	68, 13, 14, 15, 16, 24, 25, 26,
	27, 13, 32, 33, 34, 70, 68, 12,
	13, 14, 15, 16, 18, 12, 22, 23,
	24, 25, 26, 27, 13, 32, 33, 34,
	70, 68, 12, 13, 14, 15, 16, 12,
	22, 23, 24, 25, 26, 27, 13, 32,
	33, 34, 70, 68, 13, 14, 15, 16,
	17, 18, 22, 23, 24, 25, 26, 27,
	13, 32, 33, 34, 70, 12, 68, 1,
	69, 13, 14, 15, 16, 17, 18, 19,
	21, 22, 23, 24, 25, 26, 27, 31,
	32, 33, 34, 69, 12, 68, 1, 13,
	14, 15, 16, 25, 26, 27, 32, 33,
	34, 70, 68, 1, 71, 72, 68, 9,
	68, 4, 4, 1, 69, 9, 13, 14,
	15, 16, 17, 18, 19, 20, 21, 22,
	23, 24, 25, 26, 27, 28, 29, 31,
	32, 33, 34, 69, 12, 68, 1, 69,
	13, 14, 15, 16, 17, 18, 19, 20,
	21, 22, 23, 24, 25, 26, 27, 31,
	32, 33, 34, 69, 12, 68, 28, 29,
	68, 29, 68, 4, 4, 71, 74, 35,
	73, 35, 74, 74, 73, 35, 36, 73,
}

var _useSyllableMachine_trans_targs []byte = []byte{
	2, 31, 42, 2, 3, 2, 26, 28,
	51, 52, 54, 29, 32, 33, 34, 35,
	36, 46, 47, 48, 55, 49, 43, 44,
	45, 39, 40, 41, 56, 57, 58, 50,
	37, 38, 2, 59, 61, 2, 4, 5,
	6, 7, 8, 9, 10, 21, 22, 23,
	24, 18, 19, 20, 13, 14, 15, 25,
	11, 12, 2, 2, 16, 2, 17, 2,
	27, 2, 30, 2, 2, 0, 1, 2,
	53, 2, 60,
}

var _useSyllableMachine_trans_actions []byte = []byte{
	33, 5, 5, 7, 0, 13, 0, 0,
	0, 0, 5, 0, 5, 5, 0, 0,
	0, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 0, 0, 0, 5,
	0, 0, 11, 0, 0, 19, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 9, 15, 0, 17, 0, 23,
	0, 21, 0, 25, 29, 0, 0, 31,
	0, 27, 0,
}

var _useSyllableMachine_to_state_actions []byte = []byte{
	0, 0, 1, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0,
}

var _useSyllableMachine_from_state_actions []byte = []byte{
	0, 0, 3, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0,
}

var _useSyllableMachine_eof_trans []int16 = []int16{
	1, 1, 0, 38, 38, 60, 38, 38,
	38, 38, 38, 38, 38, 38, 38, 38,
	62, 38, 38, 38, 38, 38, 38, 38,
	38, 60, 64, 66, 38, 68, 68, 69,
	69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 69, 69, 69, 69, 69,
	69, 69, 69, 72, 69, 69, 69, 69,
	69, 69, 72, 74, 74, 74,
}

const (
	useSyllableMachine_start       int = 2
	useSyllableMachine_first_final int = 2
	useSyllableMachine_error       int = -1
)

const useSyllableMachine_en_main int = 2

func findSyllablesUse(buffer *Buffer) {
	info := buffer.Info
	data := preprocessInfoUSE(info)
	p, pe := 0, len(data)
	eof := pe
	var cs, act, ts, te int

	{
		cs = useSyllableMachine_start
		ts = 0
		te = 0
		act = 0
	}

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
		_acts = int(_useSyllableMachine_from_state_actions[cs])
		_nacts = uint(_useSyllableMachine_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _useSyllableMachine_actions[_acts-1] {
			case 1:
				ts = p
			}
		}

		_keys = int(_useSyllableMachine_key_offsets[cs])
		_trans = int(_useSyllableMachine_index_offsets[cs])

		_klen = int(_useSyllableMachine_single_lengths[cs])
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
				case ((data[p]).p.v.complexCategory) < _useSyllableMachine_trans_keys[_mid]:
					_upper = _mid - 1
				case ((data[p]).p.v.complexCategory) > _useSyllableMachine_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_useSyllableMachine_range_lengths[cs])
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
				case ((data[p]).p.v.complexCategory) < _useSyllableMachine_trans_keys[_mid]:
					_upper = _mid - 2
				case ((data[p]).p.v.complexCategory) > _useSyllableMachine_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_useSyllableMachine_indicies[_trans])
	_eof_trans:
		cs = int(_useSyllableMachine_trans_targs[_trans])

		if _useSyllableMachine_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_useSyllableMachine_trans_actions[_trans])
		_nacts = uint(_useSyllableMachine_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _useSyllableMachine_actions[_acts-1] {
			case 2:
				te = p + 1

			case 3:
				te = p + 1
				{
					foundSyllableUSE(useIndependentCluster, data, ts, te, info, &syllableSerial)
				}
			case 4:
				te = p + 1
				{
					foundSyllableUSE(useStandardCluster, data, ts, te, info, &syllableSerial)
				}
			case 5:
				te = p + 1
				{
					foundSyllableUSE(useBrokenCluster, data, ts, te, info, &syllableSerial)
				}
			case 6:
				te = p + 1
				{
					foundSyllableUSE(useNonCluster, data, ts, te, info, &syllableSerial)
				}
			case 7:
				te = p
				p--
				{
					foundSyllableUSE(useViramaTerminatedCluster, data, ts, te, info, &syllableSerial)
				}
			case 8:
				te = p
				p--
				{
					foundSyllableUSE(useSakotTerminatedCluster, data, ts, te, info, &syllableSerial)
				}
			case 9:
				te = p
				p--
				{
					foundSyllableUSE(useStandardCluster, data, ts, te, info, &syllableSerial)
				}
			case 10:
				te = p
				p--
				{
					foundSyllableUSE(useNumberJoinerTerminatedCluster, data, ts, te, info, &syllableSerial)
				}
			case 11:
				te = p
				p--
				{
					foundSyllableUSE(useNumeralCluster, data, ts, te, info, &syllableSerial)
				}
			case 12:
				te = p
				p--
				{
					foundSyllableUSE(useSymbolCluster, data, ts, te, info, &syllableSerial)
				}
			case 13:
				te = p
				p--
				{
					foundSyllableUSE(useHieroglyphCluster, data, ts, te, info, &syllableSerial)
				}
			case 14:
				te = p
				p--
				{
					foundSyllableUSE(useBrokenCluster, data, ts, te, info, &syllableSerial)
				}
			case 15:
				te = p
				p--
				{
					foundSyllableUSE(useNonCluster, data, ts, te, info, &syllableSerial)
				}
			case 16:
				p = (te) - 1
				{
					foundSyllableUSE(useBrokenCluster, data, ts, te, info, &syllableSerial)
				}
			}
		}

	_again:
		_acts = int(_useSyllableMachine_to_state_actions[cs])
		_nacts = uint(_useSyllableMachine_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _useSyllableMachine_actions[_acts-1] {
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
			if _useSyllableMachine_eof_trans[cs] > 0 {
				_trans = int(_useSyllableMachine_eof_trans[cs] - 1)
				goto _eof_trans
			}
		}

	}

	_ = act // needed by Ragel, but unused
}
