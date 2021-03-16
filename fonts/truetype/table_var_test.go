package truetype

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/benoitkugler/textlayout/fonts"
)

func TestVariations(t *testing.T) {
	expected := []VarAxis{
		{Tag: MustNewTag("wght"), Minimum: 100, Maximum: 900, Default: 100},
		{Tag: MustNewTag("slnt"), Minimum: -12, Maximum: 0, Default: 0},
		{Tag: MustNewTag("FLAR"), Minimum: 0, Maximum: 100, Default: 0},
		{Tag: MustNewTag("VOLM"), Minimum: 0, Maximum: 100, Default: 0},
	}
	f, err := os.Open("testdata/Commissioner-VF.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	vari := font.Fvar
	if vari == nil || len(vari.Axis) != len(expected) {
		t.Errorf("invalid number of axis")
	}
	for i, axe := range vari.Axis {
		// ignore flags and strid
		axe.flags = expected[i].flags
		axe.strid = expected[i].strid
		if axe != expected[i] {
			t.Errorf("expected %v, got %v", expected[i], axe)
		}
	}

	_, err = font.avarTable()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGvar(t *testing.T) {
	for _, file := range []string{
		"testdata/ToyVar1.ttf",
		"testdata/SelawikVar.ttf",
		"testdata/Commissioner-VF.ttf",
		"testdata/SourceSansVariable-Roman-nohvar-41,C1.ttf",
	} {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		font, err := Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		glyphs, err := font.glyfTable()
		if err != nil {
			t.Fatal(err)
		}

		ta, err := font.gvarTable(glyphs)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(len(ta.variations))

		f.Close()
	}
}

func TestHvar(t *testing.T) {
	f, err := os.Open("testdata/Commissioner-VF.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	ta, err := font.hvarTable()
	if err != nil {
		t.Fatal(err)
	}

	coords := []float32{-0.4, 0, 0.8, 1}
	for gid := GID(0); gid < fonts.GlyphIndex(font.NumGlyphs); gid++ {
		ta.getAdvanceVar(gid, coords)
	}
}

func TestPackedPointCount(t *testing.T) {
	inputs := [][]byte{
		{0, 0},
		{0x32, 0},
		{0x81, 0x22},
	}
	expecteds := []uint16{0, 50, 290}
	for i, input := range inputs {
		if got, _, _ := getPackedPointCount(input); got != expecteds[i] {
			t.Fatalf("expected %d, got %d", expecteds[i], got)
		}
	}
}

func TestPackedDeltas(t *testing.T) {
	in := []byte{0x03, 0x0A, 0x97, 0x00, 0xC6, 0x87, 0x41, 0x10, 0x22, 0xFB, 0x34}
	out, err := unpackDeltas(in, 14)
	if err != nil {
		t.Fatal(err)
	}
	if exp := []int16{10, -105, 0, -58, 0, 0, 0, 0, 0, 0, 0, 0, 4130, -1228}; !reflect.DeepEqual(out, exp) {
		t.Fatalf("expected %v, got %v", exp, out)
	}
}

// ported from harfbuzz/test/api/test-var-coords.c Copyright © 2019 Ebrahim Byagowi

func TestGetVarCoords(t *testing.T) {
	f, err := os.Open("testdata/TestCFF2VF.otf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	font, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	met := font.LoadMetrics()
	/* Design coords as input */
	designCoords := []float32{206.}
	coords := met.NormalizeVariations(designCoords)
	if exp := float32(-16116.88); exp != coords[0]*(1<<14) {
		t.Fatalf("expected %f, got %f", exp, coords[0]*(1<<14))
	}

	for weight := float32(200); weight < 901; weight++ {
		met.NormalizeVariations([]float32{weight})
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func deHexStr(s string) []byte {
	s = strings.Join(strings.Split(s, " "), "")
	if len(s)%2 != 0 {
		s += "0"
	}
	out, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func TestParseGvar(t *testing.T) {
	// imported from fonttools

	gvarData := deHexStr("0001 0000 " + //   0: majorVersion=1 minorVersion=0
		"0002 0000 " + //   4: axisCount=2 sharedTupleCount=0
		"0000001C " + //   8: offsetToSharedTuples=28
		"0003 0000 " + //  12: glyphCount=3 flags=0
		"0000001C " + //  16: offsetToGlyphVariationData=28
		"0000 0000 000C 002F " + //  20: offsets=[0,0,12,47], times 2: [0,0,24,94],
		//                 //           +offsetToGlyphVariationData: [28,28,52,122]
		//
		// 28: Glyph variation data for glyph //0, ".notdef"
		// ------------------------------------------------
		// (no variation data for this glyph)
		//
		// 28: Glyph variation data for glyph //1, "space"
		// ----------------------------------------------
		"8001 000C " + //  28: tupleVariationCount=1|TUPLES_SHARE_POINT_NUMBERS, offsetToData=12(+28=40)
		"000A " + //  32: tvHeader[0].variationDataSize=10
		"8000 " + //  34: tvHeader[0].tupleIndex=EMBEDDED_PEAK
		"0000 2CCD " + //  36: tvHeader[0].peakTuple={wght:0.0, wdth:0.7}
		"00 " + //  40: all points
		"03 01 02 03 04 " + //  41: deltaX=[1, 2, 3, 4]
		"03 0b 16 21 2C " + //  46: deltaY=[11, 22, 33, 44]
		"00 " + //  51: padding
		//
		// 52: Glyph variation data for glyph //2, "I"
		// ------------------------------------------
		"8002 001c " + //  52: tupleVariationCount=2|TUPLES_SHARE_POINT_NUMBERS, offsetToData=28(+52=80)
		"0012 " + //  56: tvHeader[0].variationDataSize=18
		"C000 " + //  58: tvHeader[0].tupleIndex=EMBEDDED_PEAK|INTERMEDIATE_REGION
		"2000 0000 " + //  60: tvHeader[0].peakTuple={wght:0.5, wdth:0.0}
		"0000 0000 " + //  64: tvHeader[0].intermediateStart={wght:0.0, wdth:0.0}
		"4000 0000 " + //  68: tvHeader[0].intermediateEnd={wght:1.0, wdth:0.0}
		"0016 " + //  72: tvHeader[1].variationDataSize=22
		"A000 " + //  74: tvHeader[1].tupleIndex=EMBEDDED_PEAK|PRIVATE_POINTS
		"C000 3333 " + //  76: tvHeader[1].peakTuple={wght:-1.0, wdth:0.8}
		"00 " + //  80: all points
		"07 03 01 04 01 " + //  81: deltaX.len=7, deltaX=[3, 1, 4, 1,
		"05 09 02 06 " + //  86:                       5, 9, 2, 6]
		"07 03 01 04 01 " + //  90: deltaY.len=7, deltaY=[3, 1, 4, 1,
		"05 09 02 06 " + //  95:                       5, 9, 2, 6]
		"06 " + //  99: 6 points
		"05 00 01 03 01 " + // 100: runLen=5(+1=6); delta-encoded run=[0, 1, 4, 5,
		"01 01 " + // 105:                                    6, 7]
		"05 f8 07 fc 03 fe 01 " + // 107: deltaX.len=5, deltaX=[-8,7,-4,3,-2,1]
		"05 a8 4d 2c 21 ea 0b " + // 114: deltaY.len=5, deltaY=[-88,77,44,33,-22,11]
		"00") // 121: padding

	if len(gvarData) != 122 {
		t.Fatal("invalid length for input binary data")
	}

	gvarDataEmptyVariations := deHexStr("0001 0000 " + //  0: majorVersion=1 minorVersion=0
		"0002 0000 " + //  4: axisCount=2 sharedTupleCount=0
		"0000001c " + //  8: offsetToSharedTuples=28
		"0003 0000 " + // 12: glyphCount=3 flags=0
		"0000001c " + // 16: offsetToGlyphVariationData=28
		"0000 0000 0000 0000") // 20: offsets=[0, 0, 0, 0]
	if len(gvarDataEmptyVariations) != 28 {
		t.Fatal("invalid length for input binary data")
	}

	gvarExpected := tableGvar{
		sharedTuples: [][]float32{},
		variations: []glyphVariationData{
			0: {},
			1: []tupleVariation{{
				tupleVariationHeader: tupleVariationHeader{
					peakTuple:         []float32{0, 0.7000122},
					variationDataSize: 0x000A,
					tupleIndex:        0x8000,
				},
				deltas: []int16{1, 2, 3, 4, 11, 22, 33, 44},
			}},
			2: []tupleVariation{
				{
					tupleVariationHeader: tupleVariationHeader{
						variationDataSize:      0x0012,
						tupleIndex:             0xC000,
						peakTuple:              []float32{0.5, 0},
						intermediateStartTuple: []float32{0, 0},
						intermediateEndTuple:   []float32{1, 0},
					},
					deltas: []int16{
						3, 1, 4, 1, 5, 9, 2, 6,
						3, 1, 4, 1, 5, 9, 2, 6,
					},
				},
				{
					tupleVariationHeader: tupleVariationHeader{
						variationDataSize: 0x0016,
						tupleIndex:        0xA000,
						peakTuple:         []float32{-1, 0.7999878},
					},
					pointNumbers: []uint16{0, 1, 4, 5, 6, 7},
					deltas: []int16{
						-8, 7, -4, 3, -2, 1,
						-88, 77, 44, 33, -22, 11,
					},
				},
			},
		},
	}

	gvarEmptyVariationsExpected := tableGvar{
		variations: make([]glyphVariationData, 3),
	}

	glyphs := TableGlyf{
		0: GlyphData{}, 1: GlyphData{},
		2: GlyphData{data: simpleGlyphData{points: make([]glyphContourPoint, 4)}},
	}

	out, err := parseTableGvar(gvarData, 2, glyphs)
	if err != nil {
		t.Fatalf("parsing gvar table: %s", err)
	}
	if fmt.Sprintf("%v", out) != fmt.Sprintf("%v", gvarExpected) {
		t.Fatalf("expected \n%v\n, got \n%v", gvarExpected, out)
	}

	out, err = parseTableGvar(gvarDataEmptyVariations, 2, glyphs)
	if err != nil {
		t.Fatalf("parsing gvar table: %s", err)
	}
	if fmt.Sprintf("%v", out) != fmt.Sprintf("%v", gvarEmptyVariationsExpected) {
		t.Fatalf("expected \n%v\n, got \n%v", gvarEmptyVariationsExpected, out)
	}
}

func TestTupleVariations(t *testing.T) {
	// imported from fonttools

	// def hexencode(s):
	// 	h = hexStr(s).upper()
	// 	return ' '.join([h[i:i+2] for i in range(0, len(h), 2)])

	// AXES = {
	// 	"wdth": (0.25, 0.375, 0.5),
	// 	"wght": (0.0, 1.0, 1.0),
	// 	"opsz": (-0.75, -0.75, 0.0)
	// }

	// Tuple Variation Store of uppercase I in the Skia font, as printed in Apple's
	// TrueType spec. The actual Skia font uses a different table for uppercase I
	// than what is printed in Apple's spec, but we still want to make sure that
	// we can parse the data as it appears in the specification.
	// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6gvar.html
	skiaGvarIData := deHexStr(
		"00 08 00 24 00 33 20 00 00 15 20 01 00 1B 20 02 " +
			"00 24 20 03 00 15 20 04 00 26 20 07 00 0D 20 06 " +
			"00 1A 20 05 00 40 01 01 01 81 80 43 FF 7E FF 7E " +
			"FF 7E FF 7E 00 81 45 01 01 01 03 01 04 01 04 01 " +
			"04 01 02 80 40 00 82 81 81 04 3A 5A 3E 43 20 81 " +
			"04 0E 40 15 45 7C 83 00 0D 9E F3 F2 F0 F0 F0 F0 " +
			"F3 9E A0 A1 A1 A1 9F 80 00 91 81 91 00 0D 0A 0A " +
			"09 0A 0A 0A 0A 0A 0A 0A 0A 0A 0A 0B 80 00 15 81 " +
			"81 00 C4 89 00 C4 83 00 0D 80 99 98 96 96 96 96 " +
			"99 80 82 83 83 83 81 80 40 FF 18 81 81 04 E6 F9 " +
			"10 21 02 81 04 E8 E5 EB 4D DA 83 00 0D CE D3 D4 " +
			"D3 D3 D3 D5 D2 CE CC CD CD CD CD 80 00 A1 81 91 " +
			"00 0D 07 03 04 02 02 02 03 03 07 07 08 08 08 07 " +
			"80 00 09 81 81 00 28 40 00 A4 02 24 24 66 81 04 " +
			"08 FA FA FA 28 83 00 82 02 FF FF FF 83 02 01 01 " +
			"01 84 91 00 80 06 07 08 08 08 08 0A 07 80 03 FE " +
			"FF FF FF 81 00 08 81 82 02 EE EE EE 8B 6D 00")

	fmt.Println(len(skiaGvarIData))
	// def test_compile_sharedPeaks_nonIntermediate_sharedPoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 0.5), "wdth": (0.0, 0.8, 0.8)},
	// 		[(7,4), (8,5), (9,6)])
	// 	axisTags = ["wght", "wdth"]
	// 	sharedPeakIndices = { var.compileCoord(axisTags): 0x77 }
	// 	tup, deltas, _ = var.compile(axisTags, sharedPeakIndices,
	// 	                          sharedPoints={0,1,2})
	// 	# len(deltas)=8; flags=None; tupleIndex=0x77
	// 	# embeddedPeaks=[]; intermediateCoord=[]
	// 	self.assertEqual("00 08 00 77", hexencode(tup))
	// 	self.assertEqual("02 07 08 09 "     # deltaX: [7, 8, 9]
	// 					 "02 04 05 06",     # deltaY: [4, 5, 6]
	// 					 hexencode(deltas))

	// def test_compile_sharedPeaks_intermediate_sharedPoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.3, 0.5, 0.7), "wdth": (0.1, 0.8, 0.9)},
	// 		[(7,4), (8,5), (9,6)])
	// 	axisTags = ["wght", "wdth"]
	// 	sharedPeakIndices = { var.compileCoord(axisTags): 0x77 }
	// 	tup, deltas, _ = var.compile(axisTags, sharedPeakIndices,
	// 	                          sharedPoints={0,1,2})
	// 	# len(deltas)=8; flags=INTERMEDIATE_REGION; tupleIndex=0x77
	// 	# embeddedPeak=[]; intermediateCoord=[(0.3, 0.1), (0.7, 0.9)]
	// 	self.assertEqual("00 08 40 77 13 33 06 66 2C CD 39 9A", hexencode(tup))
	// 	self.assertEqual("02 07 08 09 "     # deltaX: [7, 8, 9]
	// 					 "02 04 05 06",     # deltaY: [4, 5, 6]
	// 					 hexencode(deltas))

	// def test_compile_sharedPeaks_nonIntermediate_privatePoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 0.5), "wdth": (0.0, 0.8, 0.8)},
	// 		[(7,4), (8,5), (9,6)])
	// 	axisTags = ["wght", "wdth"]
	// 	sharedPeakIndices = { var.compileCoord(axisTags): 0x77 }
	// 	tup, deltas, _ = var.compile(axisTags, sharedPeakIndices,
	// 	                          sharedPoints=None)
	// 	# len(deltas)=9; flags=PRIVATE_POINT_NUMBERS; tupleIndex=0x77
	// 	# embeddedPeak=[]; intermediateCoord=[]
	// 	self.assertEqual("00 09 20 77", hexencode(tup))
	// 	self.assertEqual("00 "              # all points in glyph
	// 					 "02 07 08 09 "     # deltaX: [7, 8, 9]
	// 					 "02 04 05 06",     # deltaY: [4, 5, 6]
	// 					 hexencode(deltas))

	// def test_compile_sharedPeaks_intermediate_privatePoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 1.0), "wdth": (0.0, 0.8, 1.0)},
	// 		[(7,4), (8,5), (9,6)])
	// 	axisTags = ["wght", "wdth"]
	// 	sharedPeakIndices = { var.compileCoord(axisTags): 0x77 }
	// 	tuple, deltas, _ = var.compile(axisTags,
	// 	                            sharedPeakIndices, sharedPoints=None)
	// 	# len(deltas)=9; flags=PRIVATE_POINT_NUMBERS; tupleIndex=0x77
	// 	# embeddedPeak=[]; intermediateCoord=[(0.0, 0.0), (1.0, 1.0)]
	// 	self.assertEqual("00 09 60 77 00 00 00 00 40 00 40 00",
	// 	                 hexencode(tuple))
	// 	self.assertEqual("00 "              # all points in glyph
	// 					 "02 07 08 09 "     # deltaX: [7, 8, 9]
	// 					 "02 04 05 06",     # deltaY: [4, 5, 6]
	// 					 hexencode(deltas))

	// def test_compile_embeddedPeak_nonIntermediate_sharedPoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 0.5), "wdth": (0.0, 0.8, 0.8)},
	// 		[(7,4), (8,5), (9,6)])
	// 	tup, deltas, _ = var.compile(axisTags=["wght", "wdth"],
	// 	                          sharedCoordIndices={}, sharedPoints={0, 1, 2})
	// 	# len(deltas)=8; flags=EMBEDDED_PEAK_TUPLE
	// 	# embeddedPeak=[(0.5, 0.8)]; intermediateCoord=[]
	// 	self.assertEqual("00 08 80 00 20 00 33 33", hexencode(tup))
	// 	self.assertEqual("02 07 08 09 "     # deltaX: [7, 8, 9]
	// 					 "02 04 05 06",     # deltaY: [4, 5, 6]
	// 					 hexencode(deltas))

	// def test_compile_embeddedPeak_nonIntermediate_sharedConstants(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 0.5), "wdth": (0.0, 0.8, 0.8)},
	// 		[3, 1, 4])
	// 	tup, deltas, _ = var.compile(axisTags=["wght", "wdth"],
	// 	                          sharedCoordIndices={}, sharedPoints={0, 1, 2})
	// 	# len(deltas)=4; flags=EMBEDDED_PEAK_TUPLE
	// 	# embeddedPeak=[(0.5, 0.8)]; intermediateCoord=[]
	// 	self.assertEqual("00 04 80 00 20 00 33 33", hexencode(tup))
	// 	self.assertEqual("02 03 01 04",     # delta: [3, 1, 4]
	// 					 hexencode(deltas))

	// def test_compile_embeddedPeak_intermediate_sharedPoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 1.0), "wdth": (0.0, 0.8, 0.8)},
	// 		[(7,4), (8,5), (9,6)])
	// 	tup, deltas, _ = var.compile(axisTags=["wght", "wdth"],
	// 	                          sharedCoordIndices={},
	// 	                          sharedPoints={0, 1, 2})
	// 	# len(deltas)=8; flags=EMBEDDED_PEAK_TUPLE
	// 	# embeddedPeak=[(0.5, 0.8)]; intermediateCoord=[(0.0, 0.0), (1.0, 0.8)]
	// 	self.assertEqual("00 08 C0 00 20 00 33 33 00 00 00 00 40 00 33 33",
	// 	                hexencode(tup))
	// 	self.assertEqual("02 07 08 09 "  # deltaX: [7, 8, 9]
	// 					 "02 04 05 06",  # deltaY: [4, 5, 6]
	// 					 hexencode(deltas))

	// def test_compile_embeddedPeak_nonIntermediate_privatePoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 0.5), "wdth": (0.0, 0.8, 0.8)},
	// 		[(7,4), (8,5), (9,6)])
	// 	tup, deltas, _ = var.compile(
	// 		axisTags=["wght", "wdth"], sharedCoordIndices={}, sharedPoints=None)
	// 	# len(deltas)=9; flags=PRIVATE_POINT_NUMBERS|EMBEDDED_PEAK_TUPLE
	// 	# embeddedPeak=[(0.5, 0.8)]; intermediateCoord=[]
	// 	self.assertEqual("00 09 A0 00 20 00 33 33", hexencode(tup))
	// 	self.assertEqual("00 "           # all points in glyph
	// 	                 "02 07 08 09 "  # deltaX: [7, 8, 9]
	// 	                 "02 04 05 06",  # deltaY: [4, 5, 6]
	// 	                 hexencode(deltas))

	// def test_compile_embeddedPeak_nonIntermediate_privateConstants(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.0, 0.5, 0.5), "wdth": (0.0, 0.8, 0.8)},
	// 		[7, 8, 9])
	// 	tup, deltas, _ = var.compile(
	// 		axisTags=["wght", "wdth"], sharedCoordIndices={}, sharedPoints=None)
	// 	# len(deltas)=5; flags=PRIVATE_POINT_NUMBERS|EMBEDDED_PEAK_TUPLE
	// 	# embeddedPeak=[(0.5, 0.8)]; intermediateCoord=[]
	// 	self.assertEqual("00 05 A0 00 20 00 33 33", hexencode(tup))
	// 	self.assertEqual("00 "           # all points in glyph
	// 	                 "02 07 08 09",  # delta: [7, 8, 9]
	// 	                 hexencode(deltas))

	// def test_compile_embeddedPeak_intermediate_privatePoints(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.4, 0.5, 0.6), "wdth": (0.7, 0.8, 0.9)},
	// 		[(7,4), (8,5), (9,6)])
	// 	tup, deltas, _ = var.compile(
	// 		axisTags = ["wght", "wdth"],
	// 		sharedCoordIndices={}, sharedPoints=None)
	// 	# len(deltas)=9;
	// 	# flags=PRIVATE_POINT_NUMBERS|INTERMEDIATE_REGION|EMBEDDED_PEAK_TUPLE
	// 	# embeddedPeak=(0.5, 0.8); intermediateCoord=[(0.4, 0.7), (0.6, 0.9)]
	// 	self.assertEqual("00 09 E0 00 20 00 33 33 19 9A 2C CD 26 66 39 9A",
	// 	                 hexencode(tup))
	// 	self.assertEqual("00 "              # all points in glyph
	// 	                 "02 07 08 09 "     # deltaX: [7, 8, 9]
	// 	                 "02 04 05 06",     # deltaY: [4, 5, 6]
	// 	                 hexencode(deltas))

	// def test_compile_embeddedPeak_intermediate_privateConstants(self):
	// 	var = TupleVariation(
	// 		{"wght": (0.4, 0.5, 0.6), "wdth": (0.7, 0.8, 0.9)},
	// 		[7, 8, 9])
	// 	tup, deltas, _ = var.compile(
	// 		axisTags = ["wght", "wdth"],
	// 		sharedCoordIndices={}, sharedPoints=None)
	// 	# len(deltas)=5;
	// 	# flags=PRIVATE_POINT_NUMBERS|INTERMEDIATE_REGION|EMBEDDED_PEAK_TUPLE
	// 	# embeddedPeak=(0.5, 0.8); intermediateCoord=[(0.4, 0.7), (0.6, 0.9)]
	// 	self.assertEqual("00 05 E0 00 20 00 33 33 19 9A 2C CD 26 66 39 9A",
	// 	                 hexencode(tup))
	// 	self.assertEqual("00 "             # all points in glyph
	// 	                 "02 07 08 09",    # delta: [7, 8, 9]
	// 	                 hexencode(deltas))

	// def test_compileCoord(self):
	// 	var = TupleVariation({"wght": (-1.0, -1.0, -1.0), "wdth": (0.4, 0.5, 0.6)}, [None] * 4)
	// 	self.assertEqual("C0 00 20 00", hexencode(var.compileCoord(["wght", "wdth"])))
	// 	self.assertEqual("20 00 C0 00", hexencode(var.compileCoord(["wdth", "wght"])))
	// 	self.assertEqual("C0 00", hexencode(var.compileCoord(["wght"])))

	// def test_compileIntermediateCoord(self):
	// 	var = TupleVariation({"wght": (-1.0, -1.0, 0.0), "wdth": (0.4, 0.5, 0.6)}, [None] * 4)
	// 	self.assertEqual("C0 00 19 9A 00 00 26 66", hexencode(var.compileIntermediateCoord(["wght", "wdth"])))
	// 	self.assertEqual("19 9A C0 00 26 66 00 00", hexencode(var.compileIntermediateCoord(["wdth", "wght"])))
	// 	self.assertEqual(None, var.compileIntermediateCoord(["wght"]))
	// 	self.assertEqual("19 9A 26 66", hexencode(var.compileIntermediateCoord(["wdth"])))

	data := deHexStr("DE AD C0 00 20 00 DE AD")
	if got := parseTupleRecord(data[2:], 2); !reflect.DeepEqual(got, []float32{-1, 0.5}) {
		t.Errorf("expected %v, got %v", []float32{-1, 0.5}, got)
	}

	// def test_compilePoints(self):
	// 	compilePoints = lambda p: TupleVariation.compilePoints(set(p), numPointsInGlyph=999)
	// 	self.assertEqual("00", hexencode(compilePoints(range(999))))  # all points in glyph
	// 	self.assertEqual("01 00 07", hexencode(compilePoints([7])))
	// 	self.assertEqual("01 80 FF FF", hexencode(compilePoints([65535])))
	// 	self.assertEqual("02 01 09 06", hexencode(compilePoints([9, 15])))
	// 	self.assertEqual("06 05 07 01 F7 02 01 F2", hexencode(compilePoints([7, 8, 255, 257, 258, 500])))
	// 	self.assertEqual("03 01 07 01 80 01 EC", hexencode(compilePoints([7, 8, 500])))
	// 	self.assertEqual("04 01 07 01 81 BE E7 0C 0F", hexencode(compilePoints([7, 8, 0xBEEF, 0xCAFE])))
	// 	self.maxDiff = None
	// 	self.assertEqual("81 2C" +  # 300 points (0x12c) in total
	// 			 " 7F 00" + (127 * " 01") +  # first run, contains 128 points: [0 .. 127]
	// 			 " 7F" + (128 * " 01") +  # second run, contains 128 points: [128 .. 255]
	// 			 " 2B" + (44 * " 01"),  # third run, contains 44 points: [256 .. 299]
	// 			 hexencode(compilePoints(range(300))))
	// 	self.assertEqual("81 8F" +  # 399 points (0x18f) in total
	// 			 " 7F 00" + (127 * " 01") +  # first run, contains 128 points: [0 .. 127]
	// 			 " 7F" + (128 * " 01") +  # second run, contains 128 points: [128 .. 255]
	// 			 " 7F" + (128 * " 01") +  # third run, contains 128 points: [256 .. 383]
	// 			 " 0E" + (15 * " 01"),  # fourth run, contains 15 points: [384 .. 398]
	// 			 hexencode(compilePoints(range(399))))

	// numPointsInGlyph = 65536
	// allPoints = list(range(numPointsInGlyph))
	decompilePoints := func(points []uint16, lengthRead int, dataS string) {
		data := deHexStr(dataS)
		numbers, dataNew, err := parsePointNumbers(data)
		if err != nil {
			t.Fatalf("can't parse point numbers input %v (%s): %s", data, dataS, err)
		}
		if lengthRead == -1 {
			lengthRead = len(data)
		}
		if len(dataNew)+lengthRead != len(data) {
			t.Errorf("invalid length read: expected %d, got %d", lengthRead, len(data)-len(dataNew))
		}
		if !reflect.DeepEqual(numbers, points) {
			t.Errorf("for %v, expected point numbers %v, got %v", data, points, numbers)
		}
	}

	// all points in glyph
	decompilePoints(nil, 1, "00")
	// all points in glyph (in overly verbose encoding, not explicitly prohibited by spec)
	decompilePoints(nil, 2, "80 00")
	// 2 points; first run: [9, 9+6]
	decompilePoints([]uint16{9, 15}, 4, "02 01 09 06")
	// 2 points; first run: [0xBEEF, 0xCAFE]. (0x0C0F = 0xCAFE - 0xBEEF)
	decompilePoints([]uint16{0xBEEF, 0xCAFE}, 6, "02 81 BE EF 0C 0F")
	// 1 point; first run: [7]
	decompilePoints([]uint16{7}, 3, "01 00 07")
	// 1 point; first run: [7] in overly verbose encoding
	decompilePoints([]uint16{7}, 4, "01 80 00 07")
	// 1 point; first run: [65535]; requires words to be treated as unsigned numbers
	decompilePoints([]uint16{65535}, 4, "01 80 FF FF")
	// 4 points; first run: [7, 8]; second run: [255, 257]. 257 is stored in delta-encoded bytes (0xFF + 2).
	decompilePoints([]uint16{7, 8, 263, 265}, 7, "04 01 07 01 01 FF 02")
	// combination of all encodings, followed by 4 bytes of unused data
	decompilePoints([]uint16{7, 8, 0xBEEF, 0xCAFE}, 9, "04 01 07 01 81 BE E7 0C 0F DE AD DE AD")

	rangeN := func(N int) []uint16 {
		out := make([]uint16, N)
		for i := range out {
			out[i] = uint16(i)
		}
		return out
	}
	decompilePoints(rangeN(300), -1,
		"81 2C"+ // 300 points (0x12c) in total
			" 7F 00"+strings.Repeat(" 01", 127)+ // first run, contains 128 points: [0 .. 127]
			" 7F"+strings.Repeat(" 01", 128)+ // second run, contains 128 points: [128 .. 255]
			" AB"+strings.Repeat(" 00 01", 44), // third run, contains 44 points: [256 .. 299]
	)
	decompilePoints(rangeN(399), -1,
		"81 8F"+ // 399 points (0x18f) in total
			" 7F 00"+strings.Repeat(" 01", 127)+ // first run, contains 128 points: [0 .. 127]
			" 7F"+strings.Repeat(" 01", 128)+ // second run, contains 128 points: [128 .. 255]
			" FF"+strings.Repeat(" 00 01", 128)+ // third run, contains 128 points: [256 .. 383]
			" 8E"+strings.Repeat(" 00 01", 15), // fourth run, contains 15 points: [384 .. 398]
	)

	// def test_decompilePoints_shouldAcceptBadPointNumbers(self):
	// 	decompilePoints = TupleVariation.decompilePoints_
	// 	# 2 points; first run: [3, 9].
	// 	numPointsInGlyph = 8
	// 	with CapturingLogHandler(log, "WARNING") as captor:
	// 		decompilePoints(numPointsInGlyph,
	// 		                deHexStr("02 01 03 06"), 0, "cvar")
	// 	self.assertIn("point 9 out of range in 'cvar' table",
	// 	              [r.msg for r in captor.records])

	// def test_compileDeltas_points(self):
	// 	var = TupleVariation({}, [(0,0), (1, 0), (2, 0), None, (4, 0), (5, 0)])
	// 	points = {1, 2, 3, 4}
	// 	# deltaX for points: [1, 2, 4]; deltaY for points: [0, 0, 0]
	// 	self.assertEqual("02 01 02 04 82", hexencode(var.compileDeltas(points)))

	// def test_compileDeltas_constants(self):
	// 	var = TupleVariation({}, [0, 1, 2, None, 4, 5])
	// 	cvts = {1, 2, 3, 4}
	// 	# delta for cvts: [1, 2, 4]
	// 	self.assertEqual("02 01 02 04", hexencode(var.compileDeltas(cvts)))

	// def test_compileDeltaValues(self):
	// 	compileDeltaValues = lambda values: hexencode(TupleVariation.compileDeltaValues_(values))
	// 	# zeroes
	// 	self.assertEqual("80", compileDeltaValues([0]))
	// 	self.assertEqual("BF", compileDeltaValues([0] * 64))
	// 	self.assertEqual("BF 80", compileDeltaValues([0] * 65))
	// 	self.assertEqual("BF A3", compileDeltaValues([0] * 100))
	// 	self.assertEqual("BF BF BF BF", compileDeltaValues([0] * 256))
	// 	# bytes
	// 	self.assertEqual("00 01", compileDeltaValues([1]))
	// 	self.assertEqual("06 01 02 03 7F 80 FF FE", compileDeltaValues([1, 2, 3, 127, -128, -1, -2]))
	// 	self.assertEqual("3F" + (64 * " 7F"), compileDeltaValues([127] * 64))
	// 	self.assertEqual("3F" + (64 * " 7F") + " 00 7F", compileDeltaValues([127] * 65))
	// 	# words
	// 	self.assertEqual("40 66 66", compileDeltaValues([0x6666]))
	// 	self.assertEqual("43 66 66 7F FF FF FF 80 00", compileDeltaValues([0x6666, 32767, -1, -32768]))
	// 	self.assertEqual("7F" + (64 * " 11 22"), compileDeltaValues([0x1122] * 64))
	// 	self.assertEqual("7F" + (64 * " 11 22") + " 40 11 22", compileDeltaValues([0x1122] * 65))
	// 	# bytes, zeroes, bytes: a single zero is more compact when encoded as part of the bytes run
	// 	self.assertEqual("04 7F 7F 00 7F 7F", compileDeltaValues([127, 127, 0, 127, 127]))
	// 	self.assertEqual("01 7F 7F 81 01 7F 7F", compileDeltaValues([127, 127, 0, 0, 127, 127]))
	// 	self.assertEqual("01 7F 7F 82 01 7F 7F", compileDeltaValues([127, 127, 0, 0, 0, 127, 127]))
	// 	self.assertEqual("01 7F 7F 83 01 7F 7F", compileDeltaValues([127, 127, 0, 0, 0, 0, 127, 127]))
	// 	# bytes, zeroes
	// 	self.assertEqual("01 01 00", compileDeltaValues([1, 0]))
	// 	self.assertEqual("00 01 81", compileDeltaValues([1, 0, 0]))
	// 	# words, bytes, words: a single byte is more compact when encoded as part of the words run
	// 	self.assertEqual("42 66 66 00 02 77 77", compileDeltaValues([0x6666, 2, 0x7777]))
	// 	self.assertEqual("40 66 66 01 02 02 40 77 77", compileDeltaValues([0x6666, 2, 2, 0x7777]))
	// 	# words, zeroes, words
	// 	self.assertEqual("40 66 66 80 40 77 77", compileDeltaValues([0x6666, 0, 0x7777]))
	// 	self.assertEqual("40 66 66 81 40 77 77", compileDeltaValues([0x6666, 0, 0, 0x7777]))
	// 	self.assertEqual("40 66 66 82 40 77 77", compileDeltaValues([0x6666, 0, 0, 0, 0x7777]))
	// 	# words, zeroes, bytes
	// 	self.assertEqual("40 66 66 80 02 01 02 03", compileDeltaValues([0x6666, 0, 1, 2, 3]))
	// 	self.assertEqual("40 66 66 81 02 01 02 03", compileDeltaValues([0x6666, 0, 0, 1, 2, 3]))
	// 	self.assertEqual("40 66 66 82 02 01 02 03", compileDeltaValues([0x6666, 0, 0, 0, 1, 2, 3]))
	// 	# words, zeroes
	// 	self.assertEqual("40 66 66 80", compileDeltaValues([0x6666, 0]))
	// 	self.assertEqual("40 66 66 81", compileDeltaValues([0x6666, 0, 0]))
	// 	# bytes or words from floats
	// 	self.assertEqual("00 01", compileDeltaValues([1.1]))
	// 	self.assertEqual("00 02", compileDeltaValues([1.9]))
	// 	self.assertEqual("40 66 66", compileDeltaValues([0x6666 + 0.1]))
	// 	self.assertEqual("40 66 66", compileDeltaValues([0x6665 + 0.9]))

	decompileDeltas := func(deltas []int16, numPoints int, dataS string) {
		data := deHexStr(dataS)
		out, err := unpackDeltas(data, numPoints)
		if err != nil {
			t.Fatalf("can't parse deltas from %s: %s", dataS, err)
		}
		if !reflect.DeepEqual(deltas, out) {
			t.Errorf("for %v, expected deltas %v, got %v", data, deltas, out)
		}
	}
	// 83 = zero values (0x80), count = 4 (1 + 0x83 & 0x3F)
	decompileDeltas([]int16{0, 0, 0, 0}, 4, "83")
	// 41 01 02 FF FF = signed 16-bit values (0x40), count = 2 (1 + 0x41 & 0x3F)
	decompileDeltas([]int16{258, -1}, 2, "41 01 02 FF FF")
	// 01 81 07 = signed 8-bit values, count = 2 (1 + 0x01 & 0x3F)
	decompileDeltas([]int16{-127, 7}, 2, "01 81 07")
	// combination of all three encodings, followed by 4 bytes of unused data
	decompileDeltas([]int16{0, 0, 0, 0, 258, -127, -128}, 7, "83 40 01 02 01 81 80 DE AD BE EF")

	// def test_compileSharedTuples(self):
	// 	# Below, the peak coordinate {"wght": 1.0, "wdth": 0.7} appears
	// 	# three times; {"wght": 1.0, "wdth": 0.8} appears twice.
	// 	# Because the start and end of variation ranges is not encoded
	// 	# into the shared pool, they should get ignored.
	// 	deltas = [None] * 4
	// 	variations = [
	// 		TupleVariation({
	// 			"wght": (1.0, 1.0, 1.0),
	// 			"wdth": (0.5, 0.7, 1.0)
	// 		}, deltas),
	// 		TupleVariation({
	// 			"wght": (1.0, 1.0, 1.0),
	// 			"wdth": (0.2, 0.7, 1.0)
	// 		}, deltas),
	// 		TupleVariation({
	// 			"wght": (1.0, 1.0, 1.0),
	// 			"wdth": (0.2, 0.8, 1.0)
	// 		}, deltas),
	// 		TupleVariation({
	// 			"wght": (1.0, 1.0, 1.0),
	// 			"wdth": (0.3, 0.7, 1.0)
	// 		}, deltas),
	// 		TupleVariation({
	// 			"wght": (1.0, 1.0, 1.0),
	// 			"wdth": (0.3, 0.8, 1.0)
	// 		}, deltas),
	// 		TupleVariation({
	// 			"wght": (1.0, 1.0, 1.0),
	// 			"wdth": (0.3, 0.9, 1.0)
	//         }, deltas)
	// 	]
	// 	result = compileSharedTuples(["wght", "wdth"], variations)
	// 	self.assertEqual([hexencode(c) for c in result],
	// 	                 ["40 00 2C CD", "40 00 33 33"])

	// Shared tuples in the 'gvar' table of the Skia font, as printed
	// in Apple's TrueType specification.
	// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6gvar.html
	skiaGvarSharedTuplesData := deHexStr(
		"40 00 00 00 C0 00 00 00 00 00 40 00 00 00 C0 00 " +
			"C0 00 C0 00 40 00 C0 00 40 00 40 00 C0 00 40 00")

	skiaGvarSharedTuples := [][]float32{
		{1.0, 0.0},
		{-1.0, 0.0},
		{0.0, 1.0},
		{0.0, -1.0},
		{-1.0, -1.0},
		{1.0, -1.0},
		{1.0, 1.0},
		{-1.0, 1.0},
	}
	sharedTuples, err := parseSharedTuples(skiaGvarSharedTuplesData, 0, 2, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sharedTuples, skiaGvarSharedTuples) {
		t.Errorf("expected %v, got %v", skiaGvarSharedTuples, sharedTuples)
	}

	// def test_compileTupleVariationStore_roundTrip_gvar(self):
	// 	deltas = [(1,1), (2,2), (3,3), (4,4)]
	// 	variations = [
	// 		TupleVariation({"wght": (0.5, 1.0, 1.0), "wdth": (1.0, 1.0, 1.0)},
	// 		               deltas),
	// 		TupleVariation({"wght": (1.0, 1.0, 1.0), "wdth": (1.0, 1.0, 1.0)},
	// 		               deltas)
	// 	]
	// 	tupleVariationCount, tuples, data = compileTupleVariationStore(
	// 		variations, pointCount=4, axisTags=["wght", "wdth"],
	// 		sharedTupleIndices={})
	// 	self.assertEqual(
	// 		decompileTupleVariationStore("gvar", ["wght", "wdth"],
	// 		                             tupleVariationCount, pointCount=4,
	// 		                             sharedTuples={}, data=(tuples + data),
	// 		                             pos=0, dataPos=len(tuples)),
	//         variations)

	// def test_decompileTupleVariationStore_Skia_I(self):
	// 	tvar = decompileTupleVariationStore(
	// 		tableTag="gvar", axisTags=["wght", "wdth"],
	// 		tupleVariationCount=8, pointCount=18,
	// 		sharedTuples=skiaGvarSharedTuples,
	// 		data=skiaGvarIData, pos=4, dataPos=36)
	// 	self.assertEqual(len(tvar), 8)
	// 	self.assertEqual(tvar[0].axes, {"wght": (0.0, 1.0, 1.0)})
	// 	self.assertEqual(
	// 		" ".join(["%d,%d" % c for c in tvar[0].coordinates]),
	// 		"257,0 -127,0 -128,58 -130,90 -130,62 -130,67 -130,32 -127,0 "
	// 		"257,0 259,14 260,64 260,21 260,69 258,124 0,0 130,0 0,0 0,0")

	// def test_decompileTupleVariationStore_empty(self):
	// 	self.assertEqual(
	// 		decompileTupleVariationStore(tableTag="gvar", axisTags=[],
	// 		                             tupleVariationCount=0, pointCount=5,
	// 		                             sharedTuples=[],
	// 		                             data=b"", pos=4, dataPos=4),
	// 		[])

	// def test_getTupleSize(self):
	// 	getTupleSize = TupleVariation.getTupleSize_
	// 	numAxes = 3
	// 	self.assertEqual(4 + numAxes * 2, getTupleSize(0x8042, numAxes))
	// 	self.assertEqual(4 + numAxes * 4, getTupleSize(0x4077, numAxes))
	// 	self.assertEqual(4, getTupleSize(0x2077, numAxes))
	// 	self.assertEqual(4, getTupleSize(11, numAxes))

	// def test_inferRegion(self):
	// 	start, end = inferRegion_({"wght": -0.3, "wdth": 0.7})
	// 	self.assertEqual(start, {"wght": -0.3, "wdth": 0.0})
	// 	self.assertEqual(end, {"wght": 0.0, "wdth": 0.7})

	// @staticmethod
	// def xml_lines(writer):
	// 	content = writer.file.getvalue().decode("utf-8")
	// 	return [line.strip() for line in content.splitlines()][1:]

	// def test_getCoordWidth(self):
	// 	empty = TupleVariation({}, [])
	// 	self.assertEqual(empty.getCoordWidth(), 0)

	// 	empty = TupleVariation({}, [None])
	// 	self.assertEqual(empty.getCoordWidth(), 0)

	// 	gvarTuple = TupleVariation({}, [None, (0, 0)])
	// 	self.assertEqual(gvarTuple.getCoordWidth(), 2)

	// 	cvarTuple = TupleVariation({}, [None, 0])
	// 	self.assertEqual(cvarTuple.getCoordWidth(), 1)

	// 	cvarTuple.coordinates[1] *= 1.0
	// 	self.assertEqual(cvarTuple.getCoordWidth(), 1)

	// 	with self.assertRaises(TypeError):
	// 		TupleVariation({}, [None, "a"]).getCoordWidth()

	// def test_scaleDeltas_cvar(self):
	// 	var = TupleVariation({}, [100, None])

	// 	var.scaleDeltas(1.0)
	// 	self.assertEqual(var.coordinates, [100, None])

	// 	var.scaleDeltas(0.333)
	// 	self.assertAlmostEqual(var.coordinates[0], 33.3)
	// 	self.assertIsNone(var.coordinates[1])

	// 	var.scaleDeltas(0.0)
	// 	self.assertEqual(var.coordinates, [0, None])

	// def test_scaleDeltas_gvar(self):
	// 	var = TupleVariation({}, [(100, 200), None])

	// 	var.scaleDeltas(1.0)
	// 	self.assertEqual(var.coordinates, [(100, 200), None])

	// 	var.scaleDeltas(0.333)
	// 	self.assertAlmostEqual(var.coordinates[0][0], 33.3)
	// 	self.assertAlmostEqual(var.coordinates[0][1], 66.6)
	// 	self.assertIsNone(var.coordinates[1])

	// 	var.scaleDeltas(0.0)
	// 	self.assertEqual(var.coordinates, [(0, 0), None])

	// def test_roundDeltas_cvar(self):
	// 	var = TupleVariation({}, [55.5, None, 99.9])
	// 	var.roundDeltas()
	// 	self.assertEqual(var.coordinates, [56, None, 100])

	// def test_roundDeltas_gvar(self):
	// 	var = TupleVariation({}, [(55.5, 100.0), None, (99.9, 100.0)])
	// 	var.roundDeltas()
	// 	self.assertEqual(var.coordinates, [(56, 100), None, (100, 100)])

	// def test_calcInferredDeltas(self):
	// 	var = TupleVariation({}, [(0, 0), None, None, None])
	// 	coords = [(1, 1), (1, 1), (1, 1), (1, 1)]

	// 	var.calcInferredDeltas(coords, [])

	// 	self.assertEqual(
	// 		var.coordinates,
	// 		[(0, 0), (0, 0), (0, 0), (0, 0)]
	// 	)

	// def test_calcInferredDeltas_invalid(self):
	// 	# cvar tuples can't have inferred deltas
	// 	with self.assertRaises(TypeError):
	// 		TupleVariation({}, [0]).calcInferredDeltas([], [])

	// 	# origCoords must have same length as self.coordinates
	// 	with self.assertRaises(ValueError):
	// 		TupleVariation({}, [(0, 0), None]).calcInferredDeltas([], [])

	// 	# at least 4 phantom points required
	// 	with self.assertRaises(AssertionError):
	// 		TupleVariation({}, [(0, 0), None]).calcInferredDeltas([(0, 0), (0, 0)], [])

	// 	with self.assertRaises(AssertionError):
	// 		TupleVariation({}, [(0, 0)] + [None]*5).calcInferredDeltas(
	// 			[(0, 0)]*6,
	// 			[1, 0]  # endPts not in increasing order
	// 		)

	// def test_optimize(self):
	// 	var = TupleVariation({"wght": (0.0, 1.0, 1.0)}, [(0, 0)]*5)

	// 	var.optimize([(0, 0)]*5, [0])

	// 	self.assertEqual(var.coordinates, [None, None, None, None, None])

	// def test_optimize_isComposite(self):
	// 	# when a composite glyph's deltas are all (0, 0), we still want
	// 	# to write out an entry in gvar, else macOS doesn't apply any
	// 	# variations to the composite glyph (even if its individual components
	// 	# do vary).
	// 	# https://github.com/fonttools/fonttools/issues/1381
	// 	var = TupleVariation({"wght": (0.0, 1.0, 1.0)}, [(0, 0)]*5)
	// 	var.optimize([(0, 0)]*5, [0], isComposite=True)
	// 	self.assertEqual(var.coordinates, [(0, 0)]*5)

	// 	# it takes more than 128 (0, 0) deltas before the optimized tuple with
	// 	# (None) inferred deltas (except for the first) becomes smaller than
	// 	# the un-optimized one that has all deltas explicitly set to (0, 0).
	// 	var = TupleVariation({"wght": (0.0, 1.0, 1.0)}, [(0, 0)]*129)
	// 	var.optimize([(0, 0)]*129, list(range(129-4)), isComposite=True)
	// 	self.assertEqual(var.coordinates, [(0, 0)] + [None]*128)

	// def test_sum_deltas_gvar(self):
	// 	var1 = TupleVariation(
	// 		{},
	// 		[
	// 			(-20, 0), (-20, 0), (20, 0), (20, 0),
	// 			(0, 0), (0, 0), (0, 0), (0, 0),
	// 		]
	// 	)
	// 	var2 = TupleVariation(
	// 		{},
	// 		[
	// 			(-10, 0), (-10, 0), (10, 0), (10, 0),
	// 			(0, 0), (20, 0), (0, 0), (0, 0),
	// 		]
	// 	)

	// 	var1 += var2

	// 	self.assertEqual(
	// 		var1.coordinates,
	// 		[
	// 			(-30, 0), (-30, 0), (30, 0), (30, 0),
	// 			(0, 0), (20, 0), (0, 0), (0, 0),
	// 		]
	// 	)

	// def test_sum_deltas_gvar_invalid_length(self):
	// 	var1 = TupleVariation({}, [(1, 2)])
	// 	var2 = TupleVariation({}, [(1, 2), (3, 4)])

	// 	with self.assertRaisesRegex(ValueError, "deltas with different lengths"):
	// 		var1 += var2

	// def test_sum_deltas_gvar_with_inferred_points(self):
	// 	var1 = TupleVariation({}, [(1, 2), None])
	// 	var2 = TupleVariation({}, [(2, 3), None])

	// 	with self.assertRaisesRegex(ValueError, "deltas with inferred points"):
	// 		var1 += var2

	// def test_sum_deltas_cvar(self):
	// 	axes = {"wght": (0.0, 1.0, 1.0)}
	// 	var1 = TupleVariation(axes, [0, 1, None, None])
	// 	var2 = TupleVariation(axes, [None, 2, None, 3])
	// 	var3 = TupleVariation(axes, [None, None, None, 4])

	// 	var1 += var2
	// 	var1 += var3

	// 	self.assertEqual(var1.coordinates, [0, 3, None, 7])
}
