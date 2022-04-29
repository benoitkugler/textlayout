package truetype

import (
	"bytes"
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestMort(t *testing.T) {
	for _, filename := range dirFiles(t, "layout_fonts/morx") {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		_, err = font.MorxTable(ng)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseMorxLigature(t *testing.T) {
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
		Ligatures: []fonts.GID{
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

func TestMorxInsertion(t *testing.T) {
	// imported from fonttools

	// Taken from the `morx` table of the second font in DevanagariSangamMN.ttc
	// on macOS X 10.12.6; manually pruned to just contain the insertion lookup.
	morxInsertionData := deHexStr(
		"0002 0000 " + //  0: Version=2, Reserved=0
			"0000 0001 " + //  4: MorphChainCount=1
			"0000 0001 " + //  8: DefaultFlags=1
			"0000 00A4 " + // 12: StructLength=164 (+8=172)
			"0000 0000 " + // 16: MorphFeatureCount=0
			"0000 0001 " + // 20: MorphSubtableCount=1
			"0000 0094 " + // 24: Subtable[0].StructLength=148 (+24=172)
			"00 " + // 28: Subtable[0].CoverageFlags=0x00
			"00 00 " + // 29: Subtable[0].Reserved=0
			"05 " + // 31: Subtable[0].MorphType=5/InsertionMorph
			"0000 0001 " + // 32: Subtable[0].SubFeatureFlags=0x1
			"0000 0006 " + // 36: STXHeader.ClassCount=6
			"0000 0014 " + // 40: STXHeader.ClassTableOffset=20 (+36=56)
			"0000 004A " + // 44: STXHeader.StateArrayOffset=74 (+36=110)
			"0000 006E " + // 48: STXHeader.EntryTableOffset=110 (+36=146)
			"0000 0086 " + // 52: STXHeader.InsertionActionOffset=134 (+36=170)
			// Glyph class table.
			"0002 0006 " + //  56: ClassTable.LookupFormat=2, .UnitSize=6
			"0006 0018 " + //  60:   .NUnits=6, .SearchRange=24
			"0002 000C " + //  64:   .EntrySelector=2, .RangeShift=12
			"00AC 00AC 0005 " + //  68: GlyphID 172..172 -> GlyphClass 5
			"01EB 01E6 0005 " + //  74: GlyphID 486..491 -> GlyphClass 5
			"01F0 01F0 0004 " + //  80: GlyphID 496..496 -> GlyphClass 4
			"01F8 01F6 0004 " + //  88: GlyphID 502..504 -> GlyphClass 4
			"01FC 01FA 0004 " + //  92: GlyphID 506..508 -> GlyphClass 4
			"0250 0250 0005 " + //  98: GlyphID 592..592 -> GlyphClass 5
			"FFFF FFFF 0000 " + // 104: <end of lookup>
			// State array.
			"0000 0000 0000 0000 0001 0000 " + // 110: State[0][0..5]
			"0000 0000 0000 0000 0001 0000 " + // 122: State[1][0..5]
			"0000 0000 0001 0000 0001 0002 " + // 134: State[2][0..5]
			// Entry table.
			"0000 0000 " + // 146: Entries[0].NewState=0, .Flags=0
			"FFFF " + // 150: Entries[0].CurrentInsertIndex=<None>
			"FFFF " + // 152: Entries[0].MarkedInsertIndex=<None>
			"0002 0000 " + // 154: Entries[1].NewState=0, .Flags=0
			"FFFF " + // 158: Entries[1].CurrentInsertIndex=<None>
			"FFFF " + // 160: Entries[1].MarkedInsertIndex=<None>
			"0000 " + // 162: Entries[2].NewState=0
			"2820 " + // 164:   .Flags=CurrentIsKashidaLike,CurrentInsertBefore
			//        .CurrentInsertCount=1, .MarkedInsertCount=0
			"0000 " + // 166: Entries[1].CurrentInsertIndex=0
			"FFFF " + // 168: Entries[1].MarkedInsertIndex=<None>
			// Insertion action table.
			"022F") // 170: InsertionActionTable[0]=GlyphID 559

	if len(morxInsertionData) != 172 {
		t.Error()
	}

	out, err := parseTableMorx(morxInsertionData, 910)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 1 {
		t.Fatalf("expected one chain, got %d", len(out))
	}
	chain := out[0]

	const vertical, logical uint8 = 0, 0
	expMachine := AATStateTable{
		nClasses: 6,
		class: classFormat2{
			{start: 172, end: 172, targetClassID: 5},
			{start: 486, end: 491, targetClassID: 5},
			{start: 496, end: 496, targetClassID: 4},
			{start: 502, end: 504, targetClassID: 4},
			{start: 506, end: 508, targetClassID: 4},
			{start: 592, end: 592, targetClassID: 5},
		},
		states: [][]uint16{
			{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000}, // 110: State[0][0..5]
			{0x0000, 0x0000, 0x0000, 0x0000, 0x0001, 0x0000}, // 122: State[1][0..5]
			{0x0000, 0x0000, 0x0001, 0x0000, 0x0001, 0x0002}, // 134: State[2][0..5]
		},
		entries: []AATStateEntry{
			{NewState: 0, Flags: 0, data: [4]byte{0xff, 0xff, 0xff, 0xff}},
			{NewState: 0x0002, Flags: 0, data: [4]byte{0xff, 0xff, 0xff, 0xff}},
			{NewState: 0, Flags: 0x2820, data: [4]byte{0, 0, 0xff, 0xff}},
		},
	}
	expData := MorxInsertionSubtable{
		Insertions: []fonts.GID{0x022f},
		Machine:    expMachine,
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
	gotData, ok := gotTable.Data.(MorxInsertionSubtable)
	if !ok {
		t.Fatalf("expected MorxInsertionSubtable, got %T", gotTable.Data)
	}
	if exp, got := expData.Insertions, gotData.Insertions; !reflect.DeepEqual(exp, got) {
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

func TestATTClassFormat4(t *testing.T) {
	// adapted from fontttools
	classData := deHexStr(
		"0004 0006 0003 000C 0001 0006 " +
			"0002 0001 001E " + // glyph 1..2: mapping at offset 0x1E
			"0005 0004 001E " + // glyph 4..5: mapping at offset 0x1E
			"FFFF FFFF FFFF " + // end of search table
			"0007 0008")
	class, err := parseAATLookupTable(classData, 0, 4, false)
	if err != nil {
		t.Fatal(err)
	}
	gids := []GID{1, 2, 3, 4, 5}
	classes := []uint32{7, 8, 0, 7, 8}
	for i, gid := range gids {
		if c, _ := class.ClassID(gid); c != classes[i] {
			t.Errorf("class format 4: expected %d for glyph %d, got %d", classes[i], gid, c)
		}
	}
	if nb := class.GlyphSize(); nb != 4 {
		t.Errorf("class format 4: invalid glyph size %d", nb)
	}
	if nb := class.Extent(); nb != 9 {
		t.Errorf("class format 4: invalid glyph size %d", nb)
	}

	// extracted from macos Tamil MN font
	classData = []byte{0, 4, 0, 6, 0, 5, 0, 24, 0, 2, 0, 6, 0, 151, 0, 129, 0, 42, 0, 156, 0, 153, 0, 88, 0, 163, 0, 163, 0, 96, 1, 48, 1, 48, 0, 98, 255, 255, 255, 255, 0, 100, 0, 4, 0, 10, 0, 11, 0, 12, 0, 13, 0, 14, 0, 15, 0, 16, 0, 17, 0, 18, 0, 19, 0, 20, 0, 21, 0, 22, 0, 23, 0, 24, 0, 25, 0, 26, 0, 27, 0, 28, 0, 29, 0, 30, 0, 31, 0, 5, 0, 6, 0, 7, 0, 8, 0, 9, 0, 32}
	class, err = parseAATLookupTable(classData, 0, 0xFFFF, false)
	if err != nil {
		t.Fatal(err)
	}
	gids = []GID{132, 129, 144, 145, 146, 140, 137, 130, 135, 138, 133, 139, 142, 143, 136, 134, 147, 141, 151, 132, 150, 148, 149, 304, 153, 154, 163, 155, 156}
	classes = []uint32{
		12, 4, 24, 25, 26, 20, 17, 10, 15, 18, 13, 19, 22, 23, 16, 14, 27, 21, 31, 12, 30, 28, 29, 32, 5, 6, 9, 7, 8,
	}
	for i, gid := range gids {
		if c, _ := class.ClassID(gid); c != classes[i] {
			t.Errorf("class format 4: expected %d for glyph %d, got %d", classes[i], gid, c)
		}
	}
	if nb := class.Extent(); nb != 33 {
		t.Errorf("class format 4: invalid glyph size %d", nb)
	}
}
