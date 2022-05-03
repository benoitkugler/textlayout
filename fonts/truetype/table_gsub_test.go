package truetype

import (
	"bytes"
	"reflect"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/truetype"
)

func TestGSUB(t *testing.T) {
	filenames := []string{
		"Raleway-v4020-Regular.otf",
		"Estedad-VF.ttf",
		"Mada-VF.ttf",
	}

	filenames = append(filenames, dirFiles(t, "layout_fonts/gsub")...)

	for _, filename := range filenames {
		file, err := testdata.Files.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := NewFontParser(bytes.NewReader(file))
		if err != nil {
			t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
		}

		sub, err := font.GSUBTable()
		if err != nil {
			t.Fatal(filename, err)
		}
		for _, l := range sub.Lookups {
			for _, s := range l.Subtables {
				_ = s.Coverage.Size()
			}
		}
	}
}

func TestGSUBIndic(t *testing.T) {
	filename := "ToyIndicGSUB.ttf"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := NewFontParser(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	sub, err := font.GSUBTable()
	if err != nil {
		t.Fatal(filename, err)
	}

	expected := TableGSUB{
		TableLayout: TableLayout{
			Scripts: []Script{
				{
					Tag: MustNewTag("beng"),
					DefaultLanguage: &LangSys{
						RequiredFeatureIndex: 0xFFFF,
						Features:             []uint16{0, 2},
					},
				},
				{
					Tag: MustNewTag("bng2"),
					DefaultLanguage: &LangSys{
						RequiredFeatureIndex: 0xFFFF,
						Features:             []uint16{1, 2},
					},
				},
			},
			Features: []FeatureRecord{
				{
					Tag: MustNewTag("init"),
					Feature: Feature{
						LookupIndices: []uint16{0},
					},
				},
				{
					Tag: MustNewTag("init"),
					Feature: Feature{
						LookupIndices: []uint16{1},
					},
				},
				{
					Tag: MustNewTag("blws"),
					Feature: Feature{
						LookupIndices: []uint16{2},
					},
				},
			},
		},
		Lookups: []LookupGSUB{
			{
				Type: 1,
				LookupOptions: LookupOptions{
					Flag: 0,
				},
				Subtables: []GSUBSubtable{
					{
						Coverage: CoverageList{6, 7},
						Data:     GSUBSingle1(3),
					},
				},
			},
			{
				Type: 1,
				LookupOptions: LookupOptions{
					Flag: 0,
				},
				Subtables: []GSUBSubtable{
					{
						Coverage: CoverageList{6, 7},
						Data:     GSUBSingle1(3),
					},
				},
			},
			{
				Type: 6,
				LookupOptions: LookupOptions{
					Flag: 0,
				},
				Subtables: []GSUBSubtable{
					{
						Coverage: CoverageList{5},
						Data: GSUBChainedContext2{
							BacktrackClass: classFormat1{
								startGlyph: 2,
								classIDs:   []uint32{1},
							},
							InputClass: classFormat1{
								startGlyph: 5,
								classIDs:   []uint32{1},
							},
							LookaheadClass: classFormat2{},
							SequenceSets: [][]ChainedSequenceRule{
								nil,
								{
									ChainedSequenceRule{
										Backtrack: []uint16{1},
										Lookahead: []uint16{},
										SequenceRule: SequenceRule{
											Input: []uint16{},
											Lookups: []SequenceLookup{
												{InputIndex: 0, LookupIndex: 3},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Type: 1,
				LookupOptions: LookupOptions{
					Flag: 0,
				},
				Subtables: []GSUBSubtable{
					{
						Coverage: CoverageList{5},
						Data:     GSUBSingle1(6),
					},
				},
			},
		},
	}
	if exp, got := expected.Scripts, sub.Scripts; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	if exp, got := expected.Features, sub.Features; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected %v, got %v", exp, got)
	}
	if exp, got := expected.Lookups, sub.Lookups; !reflect.DeepEqual(exp, got) {
		t.Fatalf("expected \n%v\n, got \n%v\n", exp, got)
	}
}

func TestGSUBLigature(t *testing.T) {
	filename := "ToyGSUBLigature.ttf"
	file, err := testdata.Files.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := NewFontParser(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	sub, err := font.GSUBTable()
	if err != nil {
		t.Fatal(filename, err)
	}
	lookup := sub.Lookups[0].Subtables[0]

	coverage := CoverageList{3, 4, 7, 8, 9}
	expected := GSUBLigature1{
		{ // glyph="3"
			{Components: []uint16{5}, Glyph: 6},
		},
		{ // glyph="4"
			{Components: []uint16{4}, Glyph: 31},
			{Components: []uint16{7}, Glyph: 32},
			{Components: []uint16{11}, Glyph: 34},
		},
		{ // glyph="7"
			{Components: []uint16{4}, Glyph: 37},
			{Components: []uint16{7, 7}, Glyph: 40},
			{Components: []uint16{7}, Glyph: 8},
			{Components: []uint16{8}, Glyph: 40},
			{Components: []uint16{11}, Glyph: 38},
		},
		{ // glyph="8"
			{Components: []uint16{7}, Glyph: 40},
			{Components: []uint16{11}, Glyph: 42},
		},
		{ // glyph="9"
			{Components: []uint16{4}, Glyph: 44},
			{Components: []uint16{7}, Glyph: 45},
			{Components: []uint16{9}, Glyph: 10},
			{Components: []uint16{11}, Glyph: 46},
		},
	}

	if !reflect.DeepEqual(lookup.Coverage, coverage) {
		t.Fatalf("invalid coverage: expected %v, got %v", coverage, lookup.Coverage)
	}
	if !reflect.DeepEqual(lookup.Data, expected) {
		t.Fatalf("invalid lookup: expected %v, got %v", expected, lookup.Data)
	}
}
