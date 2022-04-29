package truetype

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
	"github.com/benoitkugler/textlayout/fonts"
)

func TestVariations(t *testing.T) {
	expected := []VarAxis{
		{Tag: MustNewTag("wght"), Minimum: 100, Maximum: 900, Default: 100},
		{Tag: MustNewTag("slnt"), Minimum: -12, Maximum: 0, Default: 0},
		{Tag: MustNewTag("FLAR"), Minimum: 0, Maximum: 100, Default: 0},
		{Tag: MustNewTag("VOLM"), Minimum: 0, Maximum: 100, Default: 0},
	}
	f, err := testdata.Files.ReadFile("Commissioner-VF.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := NewFontParser(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}

	names, err := font.tryAndLoadNameTable()
	if err != nil {
		t.Fatal(err)
	}

	vari, err := font.tryAndLoadFvarTable(names)
	if err != nil {
		t.Fatal(err)
	}

	if len(vari.Axis) != len(expected) {
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
}

func TestGvar(t *testing.T) {
	for _, file := range []string{
		"ToyVar1.ttf",
		"SelawikVar.ttf",
		"Commissioner-VF.ttf",
		"SourceSansVariable-Roman-nohvar-41,C1.ttf",
	} {
		f, err := testdata.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		font, err := NewFontParser(bytes.NewReader(f))
		if err != nil {
			t.Fatal(err)
		}

		ng, err := font.NumGlyphs()
		if err != nil {
			t.Fatal(err)
		}

		head, err := font.loadHeadTable()
		if err != nil {
			t.Fatal(err)
		}

		glyphs, err := font.GlyfTable(ng, head.indexToLocFormat)
		if err != nil {
			t.Fatal(err)
		}

		names, err := font.tryAndLoadNameTable()
		if err != nil {
			t.Fatal(err)
		}

		fvar, err := font.tryAndLoadFvarTable(names)
		if err != nil {
			t.Fatal(err)
		}

		ta, err := font.gvarTable(glyphs, fvar)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(len(ta.variations))
	}
}

func TestHvar(t *testing.T) {
	f, err := testdata.Files.ReadFile("Commissioner-VF.ttf")
	if err != nil {
		t.Fatal(err)
	}

	font, err := NewFontParser(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}

	names, err := font.tryAndLoadNameTable()
	if err != nil {
		t.Fatal(err)
	}

	fvar, err := font.tryAndLoadFvarTable(names)
	if err != nil {
		t.Fatal(err)
	}

	ta, err := font.hvarTable(fvar)
	if err != nil {
		t.Fatal(err)
	}

	ng, err := font.NumGlyphs()
	if err != nil {
		t.Fatal(err)
	}

	coords := []float32{-0.4, 0, 0.8, 1}
	for gid := GID(0); gid < fonts.GID(ng); gid++ {
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
	font := loadFont(t, "TestCFF2VF.otf")

	/* Design coords as input */
	designCoords := []float32{206.}
	coords := font.NormalizeVariations(designCoords)
	if exp := float32(-16116.88); exp != coords[0]*(1<<14) {
		t.Fatalf("expected %f, got %f", exp, coords[0]*(1<<14))
	}

	for weight := float32(200); weight < 901; weight++ {
		font.NormalizeVariations([]float32{weight})
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

	data := deHexStr("DE AD C0 00 20 00 DE AD")
	if got := parseTupleRecord(data[2:], 2); !reflect.DeepEqual(got, []float32{-1, 0.5}) {
		t.Errorf("expected %v, got %v", []float32{-1, 0.5}, got)
	}

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
	glyphData, err := parseOneGlyphVariationData(skiaGvarIData, 0, false, 2, 18)
	if err != nil {
		t.Fatal(err)
	}
	if len(glyphData) != 8 {
		t.Errorf("expected 8 tuples, got %d", len(glyphData))
	}
	deltas := []int16{
		257, -127, -128, -130, -130, -130, -130, -127, 257, 259, 260, 260, 260, 258, 0, 130, 0, 0,
		0, 0, 58, 90, 62, 67, 32, 0, 0, 14, 64, 21, 69, 124, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(glyphData[0].deltas, deltas) {
		t.Errorf("expected %v\n, got \n%v", deltas, glyphData[0].deltas)
	}
}

func TestGlyphExtentsVar(t *testing.T) {
	font := loadFont(t, "SourceSansVariable-Roman.anchor.ttf")

	coords := font.NormalizeVariations([]float32{500})
	font.SetVarCoordinates(coords)

	ext2, _ := font.GlyphExtents(2, 0, 0)

	fmt.Println("Extents from points with var", ext2)
}

func TestNormalize(t *testing.T) {
	axis := []VarAxis{
		{Tag: 0x77676874, Minimum: 38, Default: 88, Maximum: 250},
		{Tag: 0x77647468, Minimum: 60, Default: 402, Maximum: 402},
		{Tag: 0x6f70737a, Minimum: 10, Default: 14, Maximum: 72},
		// {Tag: 0x584f5051, Minimum: 5, Default: 88, Maximum: 500},
		// {Tag: 0x58545241, Minimum: 42, Default: 402, Maximum: 402},
		// {Tag: 0x594f5051, Minimum: 4, Default: 50, Maximum: 85},
		// {Tag: 0x59544c43, Minimum: 445, Default: 500, Maximum: 600},
		// {Tag: 0x59545345, Minimum: 0, Default: 18, Maximum: 48},
		// {Tag: 0x47524144, Minimum: 88, Default: 88, Maximum: 150},
		// {Tag: 0x58544348, Minimum: 800, Default: 1000, Maximum: 1200},
		// {Tag: 0x59544348, Minimum: 800, Default: 1000, Maximum: 1200},
		// {Tag: 0x59544153, Minimum: 650, Default: 750, Maximum: 850},
		// {Tag: 0x59544445, Minimum: 150, Default: 250, Maximum: 350},
		// {Tag: 0x59545543, Minimum: 650, Default: 750, Maximum: 950},
		// {Tag: 0x59545241, Minimum: 800, Default: 1000, Maximum: 1200},
	}
	vars := []Variation{
		{Tag: MustNewTag("wdth"), Value: 60},
	}

	tf := TableFvar{Axis: axis}

	coords := tf.GetDesignCoordsDefault(vars)
	if exp := []float32{88, 60, 14}; !reflect.DeepEqual(coords, exp) {
		t.Fatalf("expected %v, got %v", exp, coords)
	}
}
