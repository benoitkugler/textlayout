package fontconfig

import "testing"

// ported from fontconfig/test/test-bz89617.c: 2000 Keith Packard 2015 Akira TAGOH

func comp(l1, l2 string) FcLangResult {
	var ls1, ls2 FcLangSet

	ls1.add(l1)
	ls2.add(l2)

	return FcLangSetCompare(ls1, ls2)
}

func TestCompareLang(t *testing.T) {
	/* 1 */
	if comp("ku-am", "ku-iq") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-am", "ku-iq")
	}

	/* 2 */
	if comp("ku-am", "ku-ir") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-am", "ku-ir")
	}

	/* 3 */
	if comp("ku-am", "ku-tr") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-am", "ku-tr")
	}

	/* 4 */
	if comp("ku-iq", "ku-ir") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-iq", "ku-ir")
	}

	/* 5 */
	if comp("ku-iq", "ku-tr") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-iq", "ku-tr")
	}

	/* 6 */
	if comp("ku-ir", "ku-tr") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ku-ir", "ku-tr")
	}

	/* 7 */
	if comp("ps-af", "ps-pk") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ps-af", "ps-pk")
	}

	/* 8 */
	if comp("ti-er", "ti-et") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "ti-er", "ti-et")
	}

	/* 9 */
	if comp("zh-cn", "zh-hk") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-hk")
	}

	/* 10 */
	if comp("zh-cn", "zh-mo") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-mo")
	}

	/* 11 */
	if comp("zh-cn", "zh-sg") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-sg")
	}

	/* 12 */
	if comp("zh-cn", "zh-tw") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-cn", "zh-tw")
	}

	/* 13 */
	if comp("zh-hk", "zh-mo") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-hk", "zh-mo")
	}

	/* 14 */
	if comp("zh-hk", "zh-sg") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-hk", "zh-sg")
	}

	/* 15 */
	if comp("zh-hk", "zh-tw") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-hk", "zh-tw")
	}

	/* 16 */
	if comp("zh-mo", "zh-sg") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-mo", "zh-sg")
	}

	/* 17 */
	if comp("zh-mo", "zh-tw") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-mo", "zh-tw")
	}

	/* 18 */
	if comp("zh-sg", "zh-tw") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "zh-sg", "zh-tw")
	}

	/* 19 */
	if comp("mn-mn", "mn-cn") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "mn-mn", "mn-cn")
	}

	/* 20 */
	if comp("pap-an", "pap-aw") != FcLangDifferentTerritory {
		t.Errorf("wrong comparison for %s and %s", "pap-an", "pap-aw")
	}

}
