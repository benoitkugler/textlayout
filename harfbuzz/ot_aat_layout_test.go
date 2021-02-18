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

var feat = hb_test_open_font_file("testdata/fonts/aat-feat.ttf").LayoutTables().Feat

func hb_aat_layout_get_feature_types(feat truetype.TableFeat) []hb_aat_layout_feature_type_t {
	out := make([]hb_aat_layout_feature_type_t, len(feat))
	for i, f := range feat {
		out[i] = f.Feature
	}
	return out
}

func hb_aat_layout_feature_type_get_name_id(feat truetype.TableFeat, feature uint16) int {
	if f := feat.GetFeature(feature); f != nil {
		return int(f.NameIndex)
	}
	return -1
}

func hb_aat_layout_feature_type_get_selector_infos(feat truetype.TableFeat, feature uint16) ([]truetype.AATFeatureSelector, uint16, int) {
	if f := feat.GetFeature(feature); f != nil {
		l, s := f.GetSelectorInfos()
		return l, s, len(f.Settings)
	}
	return nil, 0xFFFF, 0
}

func TestAatGetFeatureTypes(t *testing.T) {
	features := hb_aat_layout_get_feature_types(feat)
	assertEqualInt(t, 11, len(feat))

	assertEqualInt(t, 1, int(features[0]))
	assertEqualInt(t, 3, int(features[1]))
	assertEqualInt(t, 6, int(features[2]))

	assertEqualInt(t, 258, hb_aat_layout_feature_type_get_name_id(feat, features[0]))
	assertEqualInt(t, 261, hb_aat_layout_feature_type_get_name_id(feat, features[1]))
	assertEqualInt(t, 265, hb_aat_layout_feature_type_get_name_id(feat, features[2]))
}

func TestAatGetFeatureSelectors(t *testing.T) {
	settings, default_index, total := hb_aat_layout_feature_type_get_selector_infos(feat, HB_AAT_LAYOUT_FEATURE_TYPE_DESIGN_COMPLEXITY_TYPE)
	assertEqualInt(t, 4, total)
	assertEqualInt(t, 0, int(default_index))

	assertEqualInt(t, 0, int(settings[0].Enable))
	assertEqualInt(t, 294, int(settings[0].Name))

	assertEqualInt(t, 1, int(settings[1].Enable))
	assertEqualInt(t, 295, int(settings[1].Name))

	assertEqualInt(t, 2, int(settings[2].Enable))
	assertEqualInt(t, 296, int(settings[2].Name))

	settings, default_index, total = hb_aat_layout_feature_type_get_selector_infos(feat, HB_AAT_LAYOUT_FEATURE_TYPE_TYPOGRAPHIC_EXTRAS)

	assertEqualInt(t, 1, total)
	assertEqualInt(t, HB_AAT_LAYOUT_NO_SELECTOR_INDEX, int(default_index))

	assertEqualInt(t, 8, int(settings[0].Enable))
	assertEqualInt(t, 308, int(settings[0].Name))
}

func TestAatHas(t *testing.T) {
	morx := hb_test_open_font_file("testdata/fonts/aat-morx.ttf")

	assert(t, len(morx.LayoutTables().Morx) != 0)

	trak := hb_test_open_font_file("testdata/fonts/aat-trak.ttf")
	assert(t, !trak.LayoutTables().Trak.IsEmpty())
}
