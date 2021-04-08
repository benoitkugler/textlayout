package truetype

import (
	"os"
	"reflect"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestMort(t *testing.T) {
	for _, filename := range dirFiles(t, "testdata/layout_fonts/morx") {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		_, err = font.MorxTable()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseMorx(t *testing.T) {
	// imported from fonttools

	// Taken from “Example 2: A ligature table” in
	// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6morx.html
	// as retrieved on 2017-09-11.
	//
	// Compared to the example table in Apple’s specification, we’ve
	// made the following changes:
	//
	// * at offsets 0..35, we’ve prepended 36 bytes of boilerplate
	//   to make the data a structurally valid ‘morx’ table;
	//
	// * at offsets 88..91 (offsets 52..55 in Apple’s document), we’ve
	//   changed the range of the third segment from 23..24 to 26..28.
	//   The hexdump values in Apple’s specification are completely wrong;
	//   the values from the comments would work, but they can be encoded
	//   more compactly than in the specification example. For round-trip
	//   testing, we omit the ‘f’ glyph, which makes AAT lookup format 2
	//   the most compact encoding;
	//
	// * at offsets 92..93 (offsets 56..57 in Apple’s document), we’ve
	//   changed the glyph class of the third segment from 5 to 6, which
	//   matches the values from the comments to the spec (but not the
	//   Apple’s hexdump).
	morxLigatureData := deHexStr(
		"0002 0000 " + //  0: Version=2, Reserved=0
			"0000 0001 " + //  4: MorphChainCount=1
			"0000 0001 " + //  8: DefaultFlags=1
			"0000 00DA " + // 12: StructLength=218 (+8=226)
			"0000 0000 " + // 16: MorphFeatureCount=0
			"0000 0001 " + // 20: MorphSubtableCount=1
			"0000 00CA " + // 24: Subtable[0].StructLength=202 (+24=226)
			"80 " + // 28: Subtable[0].CoverageFlags=0x80
			"00 00 " + // 29: Subtable[0].Reserved=0
			"02 " + // 31: Subtable[0].MorphType=2/LigatureMorph
			"0000 0001 " + // 32: Subtable[0].SubFeatureFlags=0x1

			// State table header.
			"0000 0007 " + // 36: STXHeader.ClassCount=7
			"0000 001C " + // 40: STXHeader.ClassTableOffset=28 (+36=64)
			"0000 0040 " + // 44: STXHeader.StateArrayOffset=64 (+36=100)
			"0000 0078 " + // 48: STXHeader.EntryTableOffset=120 (+36=156)
			"0000 0090 " + // 52: STXHeader.LigActionsOffset=144 (+36=180)
			"0000 009C " + // 56: STXHeader.LigComponentsOffset=156 (+36=192)
			"0000 00AE " + // 60: STXHeader.LigListOffset=174 (+36=210)

			// Glyph class table.
			"0002 0006 " + // 64: ClassTable.LookupFormat=2, .UnitSize=6
			"0003 000C " + // 68:   .NUnits=3, .SearchRange=12
			"0001 0006 " + // 72:   .EntrySelector=1, .RangeShift=6
			"0016 0014 0004 " + // 76: GlyphID 20..22 [a..c] -> GlyphClass 4
			"0018 0017 0005 " + // 82: GlyphID 23..24 [d..e] -> GlyphClass 5
			"001C 001A 0006 " + // 88: GlyphID 26..28 [g..i] -> GlyphClass 6
			"FFFF FFFF 0000 " + // 94: <end of lookup>

			// State array.
			"0000 0000 0000 0000 0001 0000 0000 " + // 100: State[0][0..6]
			"0000 0000 0000 0000 0001 0000 0000 " + // 114: State[1][0..6]
			"0000 0000 0000 0000 0001 0002 0000 " + // 128: State[2][0..6]
			"0000 0000 0000 0000 0001 0002 0003 " + // 142: State[3][0..6]

			// Entry table.
			"0000 0000 " + // 156: Entries[0].NewState=0, .Flags=0
			"0000 " + // 160: Entries[0].ActionIndex=<n/a> because no 0x2000 flag
			"0002 8000 " + // 162: Entries[1].NewState=2, .Flags=0x8000 (SetComponent)
			"0000 " + // 166: Entries[1].ActionIndex=<n/a> because no 0x2000 flag
			"0003 8000 " + // 168: Entries[2].NewState=3, .Flags=0x8000 (SetComponent)
			"0000 " + // 172: Entries[2].ActionIndex=<n/a> because no 0x2000 flag
			"0000 A000 " + // 174: Entries[3].NewState=0, .Flags=0xA000 (SetComponent,Act)
			"0000 " + // 178: Entries[3].ActionIndex=0 (start at Action[0])

			// Ligature actions table.
			"3FFF FFE7 " + // 180: Action[0].Flags=0, .GlyphIndexDelta=-25
			"3FFF FFED " + // 184: Action[1].Flags=0, .GlyphIndexDelta=-19
			"BFFF FFF2 " + // 188: Action[2].Flags=<end of list>, .GlyphIndexDelta=-14

			// Ligature component table.
			"0000 0001 " + // 192: LigComponent[0]=0, LigComponent[1]=1
			"0002 0003 " + // 196: LigComponent[2]=2, LigComponent[3]=3
			"0000 0004 " + // 200: LigComponent[4]=0, LigComponent[5]=4
			"0000 0008 " + // 204: LigComponent[6]=0, LigComponent[7]=8
			"0010      " + // 208: LigComponent[8]=16

			// Ligature list.
			"03E8 03E9 " + // 210: LigList[0]=1000, LigList[1]=1001
			"03EA 03EB " + // 214: LigList[2]=1002, LigList[3]=1003
			"03EC 03ED " + // 218: LigList[4]=1004, LigList[3]=1005
			"03EE 03EF ") // 222: LigList[5]=1006, LigList[6]=1007

	if len(morxLigatureData) != 226 {
		t.Error()
	}

	out, err := parseTableMorx(morxLigatureData, 1515)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 1 {
		t.Fatalf("expected one chain, got %d", len(out))
	}
	chain := out[0]

	const vertical, logical uint8 = 0x80, 0x10
	expMachine := AATStateTable{
		nClasses: 7,
		class: classFormat2{
			{start: 20, end: 22, targetClassID: 4},
			{start: 23, end: 24, targetClassID: 5},
			{start: 26, end: 28, targetClassID: 6},
		},
		states: [][]uint16{
			{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000, 0x0000}, // State[0][0..6]
			{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000, 0x0000}, // State[1][0..6]
			{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0002, 0x0000}, // State[2][0..6]
			{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0002, 0x0003}, // State[3][0..6]

		},
		entries: []AATStateEntry{
			{NewState: 0, Flags: 0},
			{NewState: 0x0002, Flags: 0x8000},
			{NewState: 0x0003, Flags: 0x8000},
			{NewState: 0, Flags: 0xA000},
		},
	}
	expData := MorxLigatureSubtable{
		LigatureAction: []uint32{
			0x3FFFFFE7,
			0x3FFFFFED,
			0xBFFFFFF2,
		},
		Machine: expMachine,
		Ligatures: []fonts.GlyphIndex{
			1000, 1001, 1002, 1003, 1004, 1005, 1006, 1007,
		},
		Component: []uint16{0, 1, 2, 3, 0, 4, 0, 8, 16},
	}
	expected := MorxChain{
		DefaultFlags: 1,
		Subtables: []MortxSubtable{
			{
				Coverage: vertical,
				Flags:    1,
				Data:     expData,
			},
		},
	}

	if exp, got := expected.DefaultFlags, chain.DefaultFlags; exp != got {
		t.Fatalf("expected %d, got %d", exp, got)
	}
	if exp, got := len(expected.Subtables), len(chain.Subtables); exp != got {
		t.Fatalf("expected %d, got %d", exp, got)
	}
	expTable, gotTable := expected.Subtables[0], chain.Subtables[0]
	if exp, got := expTable.Coverage, gotTable.Coverage; exp != got {
		t.Fatalf("expected %d, got %d", exp, got)
	}
	if exp, got := expTable.Flags, gotTable.Flags; exp != got {
		t.Fatalf("expected %d, got %d", exp, got)
	}
	gotData, ok := gotTable.Data.(MorxLigatureSubtable)
	if !ok {
		t.Fatalf("expected MorxLigatureSubtable, got %T", gotTable.Data)
	}
	if exp, got := expData.Ligatures, gotData.Ligatures; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	if exp, got := expData.Component, gotData.Component; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	if exp, got := expData.LigatureAction, gotData.LigatureAction; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	gotMachine := gotData.Machine
	if exp, got := expMachine.nClasses, gotMachine.nClasses; exp != got {
		t.Fatalf("expected %d, got %d", exp, got)
	}
	if exp, got := expMachine.class, gotMachine.class; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	if exp, got := expMachine.states, gotMachine.states; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	if exp, got := expMachine.entries, gotMachine.entries; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
}
