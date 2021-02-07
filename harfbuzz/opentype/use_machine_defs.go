package opentype

import cm "github.com/benoitkugler/textlayout/harfbuzz/common"

// logic needed by the USE rl parser

func notStandardDefaultIgnorable(i cm.GlyphInfo) bool {
	return !(i.ComplexCategory == useSyllableMachine_ex_O && i.IsDefaultIgnorable())
}

type pairUSE struct {
	i int // index in the original info slice
	v cm.GlyphInfo
}

type machineIndexUSE struct {
	j int // index in the filtered slice
	p pairUSE
}

func preprocessInfoUSE(info []cm.GlyphInfo) []machineIndexUSE {
	filterMark := func(p pairUSE) bool {
		if p.v.ComplexCategory == useSyllableMachine_ex_ZWNJ {
			for i := p.i + 1; i < len(info); i++ {
				if notStandardDefaultIgnorable(info[i]) {
					return !info[i].IsUnicodeMark()
				}
			}
		}
		return true
	}
	var tmp []pairUSE
	for i, v := range info {
		if notStandardDefaultIgnorable(v) {
			p := pairUSE{i, v}
			if filterMark(p) {
				tmp = append(tmp, p)
			}
		}
	}
	data := make([]machineIndexUSE, len(tmp))
	for j, p := range tmp {
		data[j] = machineIndexUSE{j: j, p: p}
	}
	return data
}

func foundSyllableUSE(syllableType uint8, data []machineIndexUSE, ts, te int, info []cm.GlyphInfo, syllableSerial *uint8) {
	for i := data[ts].p.i; i < data[te].p.i; i++ {
		info[i].Syllable = (*syllableSerial << 4) | syllableType
	}
	*syllableSerial++
	if *syllableSerial == 16 {
		*syllableSerial = 1
	}
}
