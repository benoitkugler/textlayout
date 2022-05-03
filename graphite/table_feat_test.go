package graphite

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	testdata "github.com/benoitkugler/textlayout-testdata/graphite"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// adapted from graphite/tests/featuremap/featuremaptest.cpp

type featHeader struct {
	major   uint16
	minor   uint16
	numFeat uint16
	_       uint16
	_       uint32
}

type featDefn struct {
	featId          Tag
	numFeatSettings uint16
	_               uint16
	settingsOffset  uint32
	flags           uint16
	label           truetype.NameID
}

type featSetting struct {
	value int16
	label truetype.NameID
}

const (
	featHeaderSize  = 12
	featDefnSize    = 16
	featSettingSize = 4
)

type featTableTest struct {
	defs    []featDefn
	setting []featSetting
	header  featHeader
}

type featTableTestA struct {
	header   featHeader
	defs     [1]featDefn
	settings [2]featSetting
}

func (ct featTableTestA) asFeatTableTest() featTableTest {
	return featTableTest{
		header:  ct.header,
		defs:    ct.defs[:],
		setting: ct.settings[:],
	}
}

var testDataA = featTableTestA{
	featHeader{2, 0, 1, 0, 0},
	[1]featDefn{{0x41424344, 2, 0, featHeaderSize + featDefnSize, 0, 1}},
	[2]featSetting{{0, 10}, {1, 11}},
}

type featTableTestB struct {
	header   featHeader
	defs     [2]featDefn
	settings [4]featSetting
}

func (ct featTableTestB) asFeatTableTest() featTableTest {
	return featTableTest{
		header:  ct.header,
		defs:    ct.defs[:],
		setting: ct.settings[:],
	}
}

var testDataB = featTableTestB{
	featHeader{2, 0, 2, 0, 0},
	[2]featDefn{
		{0x41424344, 2, 0, featHeaderSize + 2*featDefnSize, 0, 1},
		{0x41424345, 2, 0, featHeaderSize + 2*featDefnSize + 2*featSettingSize, 0, 2},
	},
	[4]featSetting{{0, 10}, {1, 11}, {0, 12}, {1, 13}},
}

var testDataBunsorted = featTableTestB{
	featHeader{2, 0, 2, 0, 0},
	[2]featDefn{
		{0x41424345, 2, 0, featHeaderSize + 2*featDefnSize + 2*featSettingSize, 0, 2},
		{0x41424344, 2, 0, featHeaderSize + 2*featDefnSize, 0, 1},
	},
	[4]featSetting{{0, 10}, {1, 11}, {0, 12}, {1, 13}},
}

type featTableTestC struct {
	header   featHeader
	defs     [3]featDefn
	settings [7]featSetting
}

func (ct featTableTestC) asFeatTableTest() featTableTest {
	return featTableTest{
		header:  ct.header,
		defs:    ct.defs[:],
		setting: ct.settings[:],
	}
}

var testDataCunsorted = featTableTestC{
	featHeader{2, 0, 3, 0, 0},
	[3]featDefn{
		{0x41424343, 3, 0, featHeaderSize + 3*featDefnSize + 4*featSettingSize, 0, 1},
		{0x41424345, 2, 0, featHeaderSize + 3*featDefnSize + 2*featSettingSize, 0, 3},
		{0x41424344, 2, 0, featHeaderSize + 3*featDefnSize, 0, 2},
	},
	[7]featSetting{{0, 10}, {1, 11}, {0, 12}, {1, 13}, {0, 14}, {1, 15}, {2, 16}},
}

type featTableTestD struct {
	header   featHeader
	defs     [4]featDefn
	settings [9]featSetting
}

func (ct featTableTestD) asFeatTableTest() featTableTest {
	return featTableTest{
		header:  ct.header,
		defs:    ct.defs[:],
		setting: ct.settings[:],
	}
}

var testDataDunsorted = featTableTestD{
	featHeader{2, 0, 4, 0, 0},
	[4]featDefn{
		{400, 3, 0, featHeaderSize + 4*featDefnSize + 4*featSettingSize, 0, 1},
		{100, 2, 0, featHeaderSize + 4*featDefnSize + 2*featSettingSize, 0, 3},
		{300, 2, 0, featHeaderSize + 4*featDefnSize, 0, 2},
		{200, 2, 0, featHeaderSize + 4*featDefnSize + 7*featSettingSize, 0, 2},
	},
	[9]featSetting{{0, 10}, {1, 11}, {0, 12}, {10, 13}, {0, 14}, {1, 15}, {2, 16}, {2, 17}, {4, 18}},
}

type featTableTestE struct {
	header   featHeader
	defs     [5]featDefn
	settings [11]featSetting
}

func (ct featTableTestE) asFeatTableTest() featTableTest {
	return featTableTest{
		header:  ct.header,
		defs:    ct.defs[:],
		setting: ct.settings[:],
	}
}

var testDataE = featTableTestE{
	featHeader{2, 0, 5, 0, 0},
	[5]featDefn{
		{400, 3, 0, featHeaderSize + 5*featDefnSize + 4*featSettingSize, 0, 1},
		{100, 2, 0, featHeaderSize + 5*featDefnSize + 2*featSettingSize, 0, 3},
		{500, 2, 0, featHeaderSize + 5*featDefnSize + 9*featSettingSize, 0, 3},
		{300, 2, 0, featHeaderSize + 5*featDefnSize, 0, 2},
		{200, 2, 0, featHeaderSize + 5*featDefnSize + 7*featSettingSize, 0, 2},
	},
	[11]featSetting{{0, 10}, {1, 11}, {0, 12}, {10, 13}, {0, 14}, {1, 15}, {2, 16}, {2, 17}, {4, 18}, {1, 19}, {2, 20}},
}

var testBadOffset = featTableTestE{
	featHeader{2, 0, 5, 0, 0},
	[5]featDefn{
		{400, 3, 0, featHeaderSize + 5*featDefnSize + 4*featSettingSize, 0, 1},
		{100, 2, 0, featHeaderSize + 5*featDefnSize + 2*featSettingSize, 0, 3},
		{500, 2, 0, featHeaderSize + 5*featDefnSize + 9*featSettingSize, 0, 3},
		{300, 2, 0, featHeaderSize + 5*featDefnSize, 0, 2},
		{200, 2, 0, featHeaderSize + 5*featDefnSize + 10*featSettingSize, 0, 2},
	},
	[11]featSetting{{0, 10}, {1, 11}, {0, 12}, {10, 13}, {0, 14}, {1, 15}, {2, 16}, {2, 17}, {4, 18}, {1, 19}, {2, 20}},
}

func asBinary(data interface{}) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, data)
	return buf.Bytes()
}

func testOneParseFeat(t *testing.T, data interface{ asFeatTableTest() featTableTest }) {
	inputData := asBinary(data)

	out, err := parseTableFeat(inputData)
	if err != nil {
		t.Fatalf("unexpected error on table feat: %s", err)
	}

	expected := data.asFeatTableTest()
	if len(out) != int(expected.header.numFeat) {
		t.Fatalf("expected %d features, got %d", expected.header.numFeat, len(out))
	}

	for _, def := range expected.defs {
		feat, ok := out.findFeature(def.featId)
		if !ok {
			t.Fatalf("feature %d not found", def.featId)
		}
		if feat.label != def.label {
			t.Fatalf("expected %d, got %d", def.label, feat.label)
		}
		if len(feat.settings) != int(def.numFeatSettings) {
			t.Fatalf("expected %d feature settings, got %d", int(def.numFeatSettings), len(feat.settings))
		}
		settingIndex := int((def.settingsOffset - featHeaderSize - featDefnSize*uint32(expected.header.numFeat)) / featSettingSize)
		for j, gotFeat := range feat.settings {
			if exp := expected.setting[settingIndex+j].label; gotFeat.Label != exp {
				t.Fatalf("expected %d, got %d", exp, gotFeat.Label)
			}
		}
	}
}

func TestParseTableFeat(t *testing.T) {
	testOneParseFeat(t, testDataA)
	testOneParseFeat(t, testDataB)
	testOneParseFeat(t, testDataBunsorted)
	testOneParseFeat(t, testDataCunsorted)
	testOneParseFeat(t, testDataDunsorted)
	testOneParseFeat(t, testDataE)

	badInput := asBinary(testBadOffset)
	if _, err := parseTableFeat(badInput); err == nil {
		t.Fatalf("expected error on bad input")
	}
}

func dumpFeatures(ft *GraphiteFace) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d features\n", len(ft.feat))
	for _, feat := range ft.feat {
		label := ft.names.SelectEntry(feat.label)
		if label != nil {
			if (byte(feat.id>>24) >= 0x20 && byte(feat.id>>24) < 0x7F) &&
				(byte(feat.id>>16) >= 0x20 && byte(feat.id>>16) < 0x7F) &&
				(byte(feat.id>>8) >= 0x20 && byte(feat.id>>8) < 0x7F) &&
				(byte(feat.id) >= 0x20 && byte(feat.id) < 0x7F) {
				fmt.Fprintf(&buf, "%d %c%c%c%c %s\n", feat.id, byte(feat.id>>24), byte(feat.id>>16), byte(feat.id>>8), byte(feat.id), label)
			} else {
				fmt.Fprintf(&buf, "%d %s\n", feat.id, label)
			}
		} else {
			fmt.Fprintf(&buf, "%d\n", feat.id)
		}

		for _, setting := range feat.settings {
			labelName := ""
			if label := ft.names.SelectEntry(setting.Label); label != nil {
				labelName = label.String()
			}
			fmt.Fprintf(&buf, "\t%d\t%s\n", setting.Value, labelName)
		}
	}

	fmt.Fprintf(&buf, "Feature Languages:")
	for _, lang := range ft.sill {
		fmt.Fprintf(&buf, "\t")
		for j := 4; j != 0; j-- {
			c := byte(lang.langcode >> (j*8 - 8))
			if (c >= 0x20) && (c < 0x80) {
				fmt.Fprintf(&buf, "%c", c)
			}
		}
	}
	buf.WriteString("\n")
	return buf.Bytes()
}

func TestFindFeatures(t *testing.T) {
	// compare with graphite log output
	fonts := []string{
		"charis.ttf",
		"Padauk.ttf",
		"Scheherazadegr.ttf",
	}
	for _, filename := range fonts {
		expected, err := testdata.Files.ReadFile(strings.TrimSuffix(filename, ".ttf") + "_feat.log")
		if err != nil {
			t.Fatal(err)
		}

		ft := loadGraphite(t, filename)
		got := dumpFeatures(ft)

		if !bytes.Equal(expected, got) {
			t.Fatalf("expected \n%s\n got \n%s", expected, got)
		}
	}
}

func TestGetFeature(t *testing.T) {
	ft := loadGraphite(t, "Padauk.ttf")
	feats := ft.FeaturesForLang(0)
	if len(feats) != 11 {
		t.Fatal("wrong number of default features")
	}
	dotc := truetype.MustNewTag("dotc")
	if ft.feat[6].id != dotc {
		t.Fatal("expected dotc")
	}

	if feats.FindFeature(dotc) == nil {
		t.Fatal("feature not found")
	}
}
