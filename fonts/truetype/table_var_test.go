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

func deHexStr(s string) []byte {
	s = strings.Join(strings.Split(s, " "), "")
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
