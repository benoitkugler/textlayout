package harfbuzz

import (
	"testing"

	tt "github.com/benoitkugler/textlayout/fonts/truetype"
)

func TestOTFeature(t *testing.T) {
	face := openFontFile("fonts/cv01.otf").LayoutTables()

	cv01 := tt.NewTag('c', 'v', '0', '1')

	featureIndex := FindFeatureForLang(&face.GSUB.TableLayout, 0, DefaultLanguageIndex, cv01)
	if featureIndex == NoFeatureIndex {
		t.Fatal("failed to find feature index")
	}

	// if (!hb_ot_layout_feature_get_name_ids (face, HB_OT_TAG_GSUB, featureIndex,
	// 					&label_id, &tooltip_id, &sample_id,
	// 					&num_named_parameters, &first_param_id))
	//   g_error ("Failed to get name ids");

	// assertEqualInt (t, label_id, 256);
	// assertEqualInt (t, tooltip_id, 257);
	// assertEqualInt (t, sample_id, 258);
	// assertEqualInt (t, num_named_parameters, 2);
	// assertEqualInt (t, first_param_id, 259);

	// all_chars = hb_ot_layout_feature_get_characters (face, HB_OT_TAG_GSUB, featureIndex,
	// 						 0, &char_count, characters);

	// assertEqualInt (t, all_chars, 2);
	// assertEqualInt (t, char_count, 2);
	// assertEqualInt (t, characters, ==, 10);
	// assertEqualInt (t, characters, ==, 24030);

	// char_count = 100;
	// characters[1] = 1234;
	// all_chars = hb_ot_layout_feature_get_characters (face, HB_OT_TAG_GSUB, featureIndex,
	// 						 1, &char_count, characters);
	// assertEqualInt (t, all_chars, 2);
	// assertEqualInt (t, char_count, 1);
	// assertEqualInt (t, characters[0],  24030);
	// assertEqualInt (t, characters[1],  1234);

	// unsigned int num_entries;
	// const hb_ot_name_entry_t *entries;
	// hb_ot_name_id_t name_id;
	// hb_language_t lang;
	// char text[10];
	// unsigned int text_size = 10;
	// entries = hb_ot_name_list_names (face, &num_entries);
	// g_assert_cmpuint (12, ==, num_entries);
	// name_id = entries[3].name_id;
	// g_assert_cmpuint (3, ==, name_id);
	// lang = entries[3].language;
	// g_assert_cmpstr (languageToString (lang), ==, "en");
	// g_assert_cmpuint (27, ==, hb_ot_name_get_utf8 (face, name_id, lang, &text_size, text));
	// g_assert_cmpuint (9, ==, text_size);
	// g_assert_cmpstr (text, ==, "FontForge");
}
