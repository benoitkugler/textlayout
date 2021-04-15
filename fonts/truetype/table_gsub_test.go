package truetype

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestGSUB(t *testing.T) {
	filenames := []string{
		"testdata/Raleway-v4020-Regular.otf",
		"testdata/Estedad-VF.ttf",
		"testdata/Mada-VF.ttf",
	}

	filenames = append(filenames, dirFiles(t, "testdata/layout_fonts/gsub")...)

	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open %q: %s\n", filename, err)
		}

		font, err := Parse(file)
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
	filename := "testdata/ToyIndicGSUB.ttf"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := Parse(file)
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
								classIDs:   []uint16{1},
							},
							InputClass: classFormat1{
								startGlyph: 5,
								classIDs:   []uint16{1},
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

func TestINvalid(t *testing.T) {
	filename := "/home/benoit/go/src/github.com/benoitkugler/textlayout/harfbuzz/testdata/data/in-house/fonts/1a3d8f381387dd29be1e897e4b5100ac8b4829e1.ttf"
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %q: %s\n", filename, err)
	}

	font, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse(%q) err = %q, want nil", filename, err)
	}

	sub, err := font.GSUBTable()
	if err != nil {
		t.Fatal(filename, err)
	}
	fmt.Println(sub.Lookups[0].Subtables[0].Data.(GSUBSingle1))
	fmt.Println(sub.Lookups[0].Subtables[0].Coverage)
	// pos, err := font.GPOSTable()
	// if err != nil {
	// 	t.Fatal(filename, err)
	// }
	// fmt.Println(pos.Scripts)
}
