package fontconfig

import (
	"testing"
)

// ported from fontconfig/test/test-bz89617.c: 2000 Keith Packard 2015 Akira TAGOH

func comp(l1, l2 string) langResult {
	var ls1, ls2 Langset

	ls1.add(l1)
	ls2.add(l2)

	return langSetCompare(ls1, ls2)
}

func TestCompareLang(t *testing.T) {
	/* 1 */
	if comp("ku-am", "ku-iq") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-am", "ku-iq")
	}

	/* 2 */
	if comp("ku-am", "ku-ir") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-am", "ku-ir")
	}

	/* 3 */
	if comp("ku-am", "ku-tr") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-am", "ku-tr")
	}

	/* 4 */
	if comp("ku-iq", "ku-ir") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-iq", "ku-ir")
	}

	/* 5 */
	if comp("ku-iq", "ku-tr") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-iq", "ku-tr")
	}

	/* 6 */
	if comp("ku-ir", "ku-tr") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-ir", "ku-tr")
	}

	/* 7 */
	if comp("ps-af", "ps-pk") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ps-af", "ps-pk")
	}

	/* 8 */
	if comp("ti-er", "ti-et") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ti-er", "ti-et")
	}

	/* 9 */
	if comp("zh-cn", "zh-hk") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-hk")
	}

	/* 10 */
	if comp("zh-cn", "zh-mo") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-mo")
	}

	/* 11 */
	if comp("zh-cn", "zh-sg") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-sg")
	}

	/* 12 */
	if comp("zh-cn", "zh-tw") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-tw")
	}

	/* 13 */
	if comp("zh-hk", "zh-mo") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-hk", "zh-mo")
	}

	/* 14 */
	if comp("zh-hk", "zh-sg") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-hk", "zh-sg")
	}

	/* 15 */
	if comp("zh-hk", "zh-tw") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-hk", "zh-tw")
	}

	/* 16 */
	if comp("zh-mo", "zh-sg") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-mo", "zh-sg")
	}

	/* 17 */
	if comp("zh-mo", "zh-tw") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-mo", "zh-tw")
	}

	/* 18 */
	if comp("zh-sg", "zh-tw") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-sg", "zh-tw")
	}

	/* 19 */
	if comp("mn-mn", "mn-cn") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "mn-mn", "mn-cn")
	}

	/* 20 */
	if comp("pap-an", "pap-aw") != langDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "pap-an", "pap-aw")
	}
}

func langsetFrom(langs []string) Langset {
	var ls Langset
	for _, lang := range langs {
		ls.add(lang)
	}
	return ls
}

func TestBasicLangsetOps(t *testing.T) {
	langs := []string{
		"ku-am",
		"ku-am",
		"ku-iq",
		"ku-ir",
		"ps-af",
		"ti-er",
		"zh-cn",
		"zh-cn",
		"zh-hk",
		"zh-hk",
		"zh-mo",
		"zh-sg",
		"mn-mn",
		"pap-an",
	}
	ls := langsetFrom(langs)
	for _, lang := range langs {
		if !ls.containsLang(lang) {
			t.Errorf("missing language %s", lang)
		}
	}
}
