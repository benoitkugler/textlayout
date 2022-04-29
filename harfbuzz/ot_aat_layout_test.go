package harfbuzz

import (
	"sort"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/test/api/test-aat-layout.c Copyright © 2018  Ebrahim Byagowi

func TestAATFeaturesSorted(t *testing.T) {
	var tags []int
	for _, f := range featureMappings {
		tags = append(tags, int(f.otFeatureTag))
	}
	if !sort.IntsAreSorted(tags) {
		t.Fatalf("expected sorted tags, got %v", tags)
	}
}

var feat = openFontFile("fonts/aat-feat.ttf").LayoutTables().Feat

func aatLayoutGetFeatureTypes(feat truetype.TableFeat) []aatLayoutFeatureType {
	out := make([]aatLayoutFeatureType, len(feat))
	for i, f := range feat {
		out[i] = f.Feature
	}
	return out
}

func aatLayoutFeatureTypeGetNameID(feat truetype.TableFeat, feature uint16) int {
	if f := feat.GetFeature(feature); f != nil {
		return int(f.NameIndex)
	}
	return -1
}

func aatLayoutFeatureTypeGetSelectorInfos(feat truetype.TableFeat, feature uint16) ([]truetype.AATFeatureSelector, uint16, int) {
	if f := feat.GetFeature(feature); f != nil {
		l, s := f.GetSelectorInfos()
		return l, s, len(f.Settings)
	}
	return nil, 0xFFFF, 0
}

func TestAatGetFeatureTypes(t *testing.T) {
	features := aatLayoutGetFeatureTypes(feat)
	assertEqualInt(t, 11, len(feat))

	assertEqualInt(t, 1, int(features[0]))
	assertEqualInt(t, 3, int(features[1]))
	assertEqualInt(t, 6, int(features[2]))

	assertEqualInt(t, 258, aatLayoutFeatureTypeGetNameID(feat, features[0]))
	assertEqualInt(t, 261, aatLayoutFeatureTypeGetNameID(feat, features[1]))
	assertEqualInt(t, 265, aatLayoutFeatureTypeGetNameID(feat, features[2]))
}

func TestAatGetFeatureSelectors(t *testing.T) {
	settings, defaultIndex, total := aatLayoutFeatureTypeGetSelectorInfos(feat, aatLayoutFeatureTypeDesignComplexityType)
	assertEqualInt(t, 4, total)
	assertEqualInt(t, 0, int(defaultIndex))

	assertEqualInt(t, 0, int(settings[0].Enable))
	assertEqualInt(t, 294, int(settings[0].Name))

	assertEqualInt(t, 1, int(settings[1].Enable))
	assertEqualInt(t, 295, int(settings[1].Name))

	assertEqualInt(t, 2, int(settings[2].Enable))
	assertEqualInt(t, 296, int(settings[2].Name))

	settings, defaultIndex, total = aatLayoutFeatureTypeGetSelectorInfos(feat, aatLayoutFeatureTypeTypographicExtras)

	assertEqualInt(t, 1, total)
	assertEqualInt(t, aatLayoutNoSelectorIndex, int(defaultIndex))

	assertEqualInt(t, 8, int(settings[0].Enable))
	assertEqualInt(t, 308, int(settings[0].Name))
}

func TestAatHas(t *testing.T) {
	morx := openFontFile("fonts/aat-morx.ttf")

	assert(t, len(morx.LayoutTables().Morx) != 0)

	trak := openFontFile("fonts/aat-trak.ttf")
	assert(t, !trak.LayoutTables().Trak.IsEmpty())
}
