package pango

import (
	"fmt"
	"strings"
	"testing"
)

// TestAttributesBasic
// TestAttributesEqual
// TestList
// TestListChange
// TODO: TestListSplice
// TODO: TestListSplice2
// TestListFilter
// TestIter
// TestIterGet
// TestIterGetFont
// TestIterGetAttrs
// TODO: TestListUpdate
// TODO: TestListUpdate2
// TODO: TestListEqual
// TestInsert
// TestMerge
// TestMerge2
// TestIterEpsilonZero

func testCopy(t *testing.T, attr *Attribute) {
	a := attr.copy()
	assertTrue(t, attr.equals(*a), "cloned values")
}

func TestAttributesBasic(t *testing.T) {
	testCopy(t, NewAttrLanguage(pango_language_from_string("ja-JP")))
	testCopy(t, NewAttrFamily("Times"))
	testCopy(t, NewAttrForeground(AttrColor{100, 200, 300}))
	testCopy(t, NewAttrBackground(AttrColor{100, 200, 300}))
	testCopy(t, NewAttrSize(1024))
	testCopy(t, NewAttrAbsoluteSize(1024))
	testCopy(t, NewAttrStyle(STYLE_ITALIC))
	testCopy(t, NewAttrWeight(WEIGHT_ULTRALIGHT))
	testCopy(t, NewAttrVariant(VARIANT_SMALL_CAPS))
	testCopy(t, NewAttrStretch(STRETCH_SEMI_EXPANDED))
	testCopy(t, NewAttrFontDescription(NewFontDescriptionFrom("Computer Modern 12")))
	testCopy(t, NewAttrUnderline(UNDERLINE_LOW))
	testCopy(t, NewAttrUnderline(UNDERLINE_ERROR_LINE))
	testCopy(t, NewAttrUnderlineColor(AttrColor{100, 200, 300}))
	testCopy(t, NewAttrOverline(OVERLINE_SINGLE))
	testCopy(t, NewAttrOverlineColor(AttrColor{100, 200, 300}))
	testCopy(t, NewAttrStrikethrough(true))
	testCopy(t, NewAttrStrikethroughColor(AttrColor{100, 200, 300}))
	testCopy(t, NewAttrRise(256))
	testCopy(t, NewAttrScale(2.56))
	testCopy(t, NewAttrFallback(false))
	testCopy(t, NewAttrLetterSpacing(1024))

	rect := Rectangle{X: 0, Y: 0, Width: 10, Height: 10}
	testCopy(t, NewAttrShape(rect, rect))
	testCopy(t, NewAttrGravity(GRAVITY_SOUTH))
	testCopy(t, NewAttrGravityHint(GRAVITY_HINT_STRONG))
	testCopy(t, NewAttrAllowBreaks(false))
	testCopy(t, NewAttrShow(SHOW_SPACES))
	testCopy(t, NewAttrInsertHyphens(false))
}

func TestAttributesEqual(t *testing.T) {
	/* check that pango_attribute_equal compares values, but not ranges */
	attr1 := NewAttrSize(10)
	attr2 := NewAttrSize(20)
	attr3 := NewAttrSize(20)
	attr3.StartIndex = 1
	attr3.EndIndex = 2

	assertFalse(t, attr1.equals(*attr2), "attribute equality")
	assertTrue(t, attr2.equals(*attr3), "attribute equality")
}

func PrintAttribute(attr *Attribute, text []rune) string {
	return fmt.Sprintf("[%d,%d]%s=%s",
		AsByteIndex(text, attr.StartIndex), AsByteIndex(text, attr.EndIndex), attr.Kind, attr.Data)
}

func print_attributes(attrs AttrList) string {
	chunks := make([]string, len(attrs))
	for i, attr := range attrs {
		chunks[i] = attr.String() + "\n"
	}
	return strings.Join(chunks, "")
}

func assert_attributes(t *testing.T, attrs AttrList, expected string) {
	s := print_attributes(attrs)
	if s != expected {
		t.Errorf("-----\nattribute list mismatch\nexpected:\n%s-----\nreceived:\n%s-----\n",
			expected, s)
	}
}

func assert_attr_iterator(t *testing.T, iter *attrIterator, expected string) {
	attrs := iter.getAttributes()
	assert_attributes(t, attrs, expected)
}

func TestList(t *testing.T) {
	var list AttrList

	list.Insert(NewAttrSize(10))
	list.Insert(NewAttrSize(20))
	list.Insert(NewAttrSize(30))

	assert_attributes(t, list, "[0,2147483647]size=10\n"+
		"[0,2147483647]size=20\n"+
		"[0,2147483647]size=30\n")

	list = nil

	/* test that insertion respects StartIndex */
	list.Insert(NewAttrSize(10))
	attr := NewAttrSize(20)
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.Insert(attr)
	list.Insert(NewAttrSize(30))
	attr = NewAttrSize(40)
	attr.StartIndex = 10
	attr.EndIndex = 40
	list.pango_attr_list_insert_before(attr)

	assert_attributes(t, list, "[0,2147483647]size=10\n"+
		"[0,2147483647]size=30\n"+
		"[10,40]size=40\n"+
		"[10,20]size=20\n")
}

func TestListChange(t *testing.T) {
	var list AttrList

	attr := NewAttrSize(10)
	attr.StartIndex = 0
	attr.EndIndex = 10
	list.Insert(attr)
	attr = NewAttrSize(20)
	attr.StartIndex = 20
	attr.EndIndex = 30
	list.Insert(attr)
	attr = NewAttrWeight(WEIGHT_BOLD)
	attr.StartIndex = 0
	attr.EndIndex = 30
	list.Insert(attr)

	assert_attributes(t, list, "[0,10]size=10\n"+
		"[0,30]weight=700\n"+
		"[20,30]size=20\n")

	/* simple insertion with pango_attr_list_change */
	attr = NewAttrVariant(VARIANT_SMALL_CAPS)
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.Change(attr)

	assert_attributes(t, list, "[0,10]size=10\n"+
		"[0,30]weight=700\n"+
		"[10,20]variant=1\n"+
		"[20,30]size=20\n")

	/* insertion with splitting */
	attr = NewAttrWeight(WEIGHT_LIGHT)
	attr.StartIndex = 15
	attr.EndIndex = 20
	list.Change(attr)

	assert_attributes(t, list, "[0,10]size=10\n"+
		"[0,15]weight=700\n"+
		"[10,20]variant=1\n"+
		"[15,20]weight=300\n"+
		"[20,30]size=20\n"+
		"[20,30]weight=700\n")

	/* insertion with joining */
	attr = NewAttrSize(20)
	attr.StartIndex = 5
	attr.EndIndex = 20
	list.Change(attr)

	assert_attributes(t, list, "[0,5]size=10\n"+
		"[0,15]weight=700\n"+
		"[5,30]size=20\n"+
		"[10,20]variant=1\n"+
		"[15,20]weight=300\n"+
		"[20,30]weight=700\n")
}

// func TestListSplice (t *testing.T,void) {
//    PangoAttrList *base;
//    PangoAttrList *list;
//    PangoAttrList *other;
//    PangoAttribute *attr;

//    base = pango_attr_list_new ();
//    attr = pango_attr_size_new (10);
//    attr.StartIndex = 0;
//    attr.EndIndex = -1;
//    insert (base, attr);
//    attr = pango_attr_weight_new (PANGO_WEIGHT_BOLD);
//    attr.StartIndex = 10;
//    attr.EndIndex = 15;
//    insert (base, attr);
//    attr = pango_attr_variant_new (PANGO_VARIANT_SMALL_CAPS);
//    attr.StartIndex = 20;
//    attr.EndIndex = 30;
//    insert (base, attr);

//    assert_attributes (t,base, "[0,2147483647]size=10\n"
// 						   "[10,15]weight=700\n"
// 						   "[20,30]variant=1\n");

//    /* splice in an empty list */
//    list = copy (base);
//    other = pango_attr_list_new ();
//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[0,2147483647]size=10\n"
// 						   "[10,20]weight=700\n"
// 						   "[25,35]variant=1\n");

//    pango_attr_list_unref (list);
//    pango_attr_list_unref (other);

//    /* splice in some attributes */
//    list = copy (base);
//    other = pango_attr_list_new ();
//    attr = pango_attr_size_new (20);
//    attr.StartIndex = 0;
//    attr.EndIndex = 3;
//    insert (other, attr);
//    attr = pango_attr_stretch_new (STRETCH_CONDENSED);
//    attr.StartIndex = 2;
//    attr.EndIndex = 4;
//    insert (other, attr);

//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[0,11]size=10\n"
// 						   "[10,20]weight=700\n"
// 						   "[11,14]size=20\n"
// 						   "[13,15]stretch=2\n"
// 						   "[14,2147483647]size=10\n"
// 						   "[25,35]variant=1\n");

//    pango_attr_list_unref (list);
//    pango_attr_list_unref (other);

//    pango_attr_list_unref (base);
//  }

//  /* Test that empty lists work in pango_attr_list_splice */
// func TestListSplice2 (t *testing.T,void) {
//    PangoAttrList *list;
//    PangoAttrList *other;
//    PangoAttribute *attr;

//    var list AttrList
//    other = pango_attr_list_new ();

//    pango_attr_list_splice (list, other, 11, 5);

//    g_assert_null (pango_attr_list_get_attributes (list));

//    attr = pango_attr_size_new (10);
//    attr.StartIndex = 0;
//    attr.EndIndex = -1;
//    insert (other, attr);

//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[11,2147483647]size=10\n");

//    pango_attr_list_unref (other);
//    other = pango_attr_list_new ();

//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[11,2147483647]size=10\n");

//    pango_attr_list_unref (other);
//    pango_attr_list_unref (list);
//  }

//  static gboolean
//  just_weight (PangoAttribute *attribute, gpointer user_data)
//  {
//    if (attribute.klass.type == ATTR_WEIGHT)
// 	 return true;
//    else
// 	 return false;
//  }

func TestListFilter(t *testing.T) {
	var list AttrList
	list.Insert(NewAttrSize(10))
	attr := NewAttrStretch(STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.Insert(attr)
	attr = NewAttrWeight(WEIGHT_BOLD)
	attr.StartIndex = 20
	list.Insert(attr)

	assert_attributes(t, list, "[0,2147483647]size=10\n"+
		"[10,20]stretch=2\n"+
		"[20,2147483647]weight=700\n")

	out := list.Filter(func(attr *Attribute) bool { return false })
	if len(out) != 0 {
		t.Errorf("expected empty list, got %v", out)
	}

	out = list.Filter(func(attr *Attribute) bool { return attr.Kind == ATTR_WEIGHT })
	if len(out) == 0 {
		t.Error("expected list, got 0 elements")
	}

	assert_attributes(t, list, "[0,2147483647]size=10\n"+
		"[10,20]stretch=2\n")
	assert_attributes(t, out, "[20,2147483647]weight=700\n")
}

func TestIter(t *testing.T) {
	var list AttrList
	iter := list.getIterator()

	assertFalse(t, iter.next(), "empty iterator")
	if L := iter.getAttributes(); len(L) != 0 {
		t.Errorf("expected empty list, got %v", L)
	}

	list = nil
	list.Insert(NewAttrSize(10))
	attr := NewAttrStretch(STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.Insert(attr)
	attr = NewAttrWeight(WEIGHT_BOLD)
	attr.StartIndex = 20
	list.Insert(attr)

	iter = list.getIterator()
	copy := iter.copy()
	assertEquals(t, int(iter.StartIndex), 0)
	assertEquals(t, int(iter.EndIndex), 10)
	assertTrue(t, iter.next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), 10)
	assertEquals(t, int(iter.EndIndex), 20)
	assertTrue(t, iter.next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), 20)
	assertEquals(t, int(iter.EndIndex), 30)
	assertTrue(t, iter.next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), 30)
	assertEquals(t, int(iter.EndIndex), MaxInt)
	assertTrue(t, iter.next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), MaxInt)
	assertEquals(t, int(iter.EndIndex), MaxInt)
	assertTrue(t, !iter.next(), "iterator has no more element")

	assertEquals(t, copy.StartIndex, 0)
	assertEquals(t, copy.EndIndex, 10)
}

func TestIterGet(t *testing.T) {
	var list AttrList
	list.Insert(NewAttrSize(10))
	attr := NewAttrStretch(STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.Insert(attr)
	attr = NewAttrWeight(WEIGHT_BOLD)
	attr.StartIndex = 20
	list.Insert(attr)

	iter := list.getIterator()
	iter.next()
	attr = iter.getByKind(ATTR_SIZE)
	if attr == nil {
		t.Error("expected attribute")
	}
	assertEquals(t, attr.StartIndex, 0)
	assertEquals(t, attr.EndIndex, MaxInt)
	attr = iter.getByKind(ATTR_STRETCH)
	if attr == nil {
		t.Error("expected attribute")
	}
	assertEquals(t, attr.StartIndex, 10)
	assertEquals(t, attr.EndIndex, 30)
	attr = iter.getByKind(ATTR_WEIGHT)
	if attr != nil {
		t.Errorf("expected no attribute, got %v", attr)
	}
	attr = iter.getByKind(ATTR_GRAVITY)
	if attr != nil {
		t.Errorf("expected no attribute, got %v", attr)
	}
}

func TestIterGetFont(t *testing.T) {
	var list AttrList
	list.Insert(NewAttrSize(10 * Scale))
	list.Insert(NewAttrFamily("Times"))
	attr := NewAttrStretch(STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.Insert(attr)
	attr = NewAttrLanguage(pango_language_from_string("ja-JP"))
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.Insert(attr)
	attr = NewAttrRise(100)
	attr.StartIndex = 20
	list.Insert(attr)
	attr = NewAttrFallback(false)
	attr.StartIndex = 20
	list.Insert(attr)

	var (
		lang  Language
		attrs AttrList
	)
	iter := list.getIterator()
	desc := NewFontDescription()
	iter.getFont(&desc, &lang, &attrs)
	desc2 := NewFontDescriptionFrom("Times 10")
	assertTrue(t, desc.pango_font_description_equal(desc2), "same fonts")
	if lang != "" {
		t.Errorf("expected no language got %s", lang)
	}
	if len(attrs) != 0 {
		t.Errorf("expected no attributes, got %v", attrs)
	}

	iter.next()
	desc = NewFontDescription()
	iter.getFont(&desc, &lang, &attrs)
	desc2 = NewFontDescriptionFrom("Times Condensed 10")
	assertTrue(t, desc.pango_font_description_equal(desc2), "same fonts")
	if lang == "" {
		t.Error("expected lang")
	}
	assertEquals(t, string(lang), "ja-jp")
	if len(attrs) != 0 {
		t.Errorf("expected no attributes, got %v", attrs)
	}

	iter.next()
	desc = NewFontDescription()
	iter.getFont(&desc, &lang, &attrs)
	desc2 = NewFontDescriptionFrom("Times Condensed 10")
	assertTrue(t, desc.pango_font_description_equal(desc2), "same fonts")
	if lang != "" {
		t.Errorf("expected no language got %s", lang)
	}
	assert_attributes(t, attrs, "[20,2147483647]rise=100\n"+
		"[20,2147483647]fallback=0\n")
}

func TestIterGetAttrs(t *testing.T) {
	var list AttrList
	list.Insert(NewAttrSize(10 * Scale))
	list.Insert(NewAttrFamily("Times"))
	attr := NewAttrStretch(STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.Insert(attr)
	attr = NewAttrLanguage(pango_language_from_string("ja-JP"))
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.Insert(attr)
	attr = NewAttrRise(100)
	attr.StartIndex = 20
	list.Insert(attr)
	attr = NewAttrFallback(false)
	attr.StartIndex = 20
	list.Insert(attr)

	iter := list.getIterator()
	assert_attr_iterator(t, iter, "[0,2147483647]size=10240\n"+
		"[0,2147483647]family=Times\n")

	iter.next()
	assert_attr_iterator(t, iter, "[0,2147483647]size=10240\n"+
		"[0,2147483647]family=Times\n"+
		"[10,30]stretch=2\n"+
		"[10,20]language=ja-jp\n")

	iter.next()
	assert_attr_iterator(t, iter, "[0,2147483647]size=10240\n"+
		"[0,2147483647]family=Times\n"+
		"[10,30]stretch=2\n"+
		"[20,2147483647]rise=100\n"+
		"[20,2147483647]fallback=0\n")

	iter.next()
	assert_attr_iterator(t, iter, "[0,2147483647]size=10240\n"+
		"[0,2147483647]family=Times\n"+
		"[20,2147483647]rise=100\n"+
		"[20,2147483647]fallback=0\n")

	iter.next()
	if l := iter.getAttributes(); len(l) != 0 {
		t.Errorf("expected no attributes, got %v", l)
	}
}

// TODO: enable when list_update is added
// func TestListUpdate(t *testing.T) {
// 	var list AttrList
// 	attr := pango_attr_size_new(10 * Scale)
// 	attr.StartIndex = 10
// 	attr.EndIndex = 11
// 	list.insert(attr)
// 	attr = pango_attr_rise_new(100)
// 	attr.StartIndex = 0
// 	attr.EndIndex = 200
// 	list.insert(attr)
// 	attr = pango_attr_family_new("Times")
// 	attr.StartIndex = 5
// 	attr.EndIndex = 15
// 	list.insert(attr)
// 	attr = pango_attr_fallback_new(false)
// 	attr.StartIndex = 11
// 	attr.EndIndex = 100
// 	list.insert(attr)
// 	attr = pango_attr_stretch_new(STRETCH_CONDENSED)
// 	attr.StartIndex = 30
// 	attr.EndIndex = 60
// 	list.insert(attr)

// 	assert_attributes(t, list, "[0,200]rise=100\n"+
// 		"[5,15]family=Times\n"+
// 		"[10,11]size=10240\n"+
// 		"[11,100]fallback=0\n"+
// 		"[30,60]stretch=2\n")

// 	list.pango_attr_list_update(8, 10, 20)

// 	assert_attributes(t, list, "[0,210]rise=100\n"+
// 		"[5,8]family=Times\n"+
// 		"[28,110]fallback=0\n"+
// 		"[40,70]stretch=2\n")

// }

//  /* Test that empty lists work in pango_attr_list_update */
// func TestListUpdate2 (t *testing.T,void) {
//    PangoAttrList *list;

//    var list AttrList
//    pango_attr_list_update (list, 8, 10, 20);

//    g_assert_null (pango_attr_list_get_attributes (list));

//    pango_attr_list_unref (list);
//  }

// func TestListEqual (t *testing.T,void) {
//    PangoAttrList *list1, *list2;
//    PangoAttribute *attr;

//    list1 = pango_attr_list_new ();
//    list2 = pango_attr_list_new ();

//    assertTrue (t,pango_attr_list_equal (NULL, NULL));
//    assertFalse (t,pango_attr_list_equal (list1, NULL));
//    assertFalse (t,pango_attr_list_equal (NULL, list1));
//    assertTrue (t,pango_attr_list_equal (list1, list1));
//    assertTrue (t,pango_attr_list_equal (list1, list2));

//    attr = pango_attr_size_new (10 * Scale);
//    attr.StartIndex = 0;
//    attr.EndIndex = 7;
//    insert (list1, deepCopy (attr));
//    insert (list2, deepCopy (attr));
//    pango_attribute_destroy (attr);

//    assertTrue (t,pango_attr_list_equal (list1, list2));

//    attr = pango_attr_stretch_new (STRETCH_CONDENSED);
//    attr.StartIndex = 0;
//    attr.EndIndex = 1;
//    insert (list1, deepCopy (attr));
//    assertTrue (t,!pango_attr_list_equal (list1, list2));

//    insert (list2, deepCopy (attr));
//    assertTrue (t,pango_attr_list_equal (list1, list2));
//    pango_attribute_destroy (attr);

//    attr = pango_attr_size_new (30 * Scale);
//    /* Same range as the first attribute */
//    attr.StartIndex = 0;
//    attr.EndIndex = 7;
//    insert (list2, deepCopy (attr));
//    assertTrue (t,!pango_attr_list_equal (list1, list2));
//    insert (list1, deepCopy (attr));
//    assertTrue (t,pango_attr_list_equal (list1, list2));
//    pango_attribute_destroy (attr);

//    pango_attr_list_unref (list1);
//    pango_attr_list_unref (list2);

//    /* Same range but different order */
//    {
// 	 PangoAttrList *list1, *list2;
// 	 PangoAttribute *attr1, *attr2;

// 	 list1 = pango_attr_list_new ();
// 	 list2 = pango_attr_list_new ();

// 	 attr1 = pango_attr_size_new (10 * Scale);
// 	 attr2 = pango_attr_stretch_new (STRETCH_CONDENSED);

// 	 insert (list1, deepCopy (attr1));
// 	 insert (list1, deepCopy (attr2));

// 	 insert (list2, deepCopy (attr2));
// 	 insert (list2, deepCopy (attr1));

// 	 pango_attribute_destroy (attr1);
// 	 pango_attribute_destroy (attr2);

// 	 assertTrue (t,pango_attr_list_equal (list1, list2));
// 	 assertTrue (t,pango_attr_list_equal (list2, list1));

// 	 pango_attr_list_unref (list1);
// 	 pango_attr_list_unref (list2);
//    }
//  }

func TestInsert(t *testing.T) {
	var list AttrList
	attr := NewAttrSize(10 * Scale)
	attr.StartIndex = 10
	attr.EndIndex = 11
	list.Insert(attr)
	attr = NewAttrRise(100)
	attr.StartIndex = 0
	attr.EndIndex = 200
	list.Insert(attr)
	attr = NewAttrFamily("Times")
	attr.StartIndex = 5
	attr.EndIndex = 15
	list.Insert(attr)
	attr = NewAttrFallback(false)
	attr.StartIndex = 11
	attr.EndIndex = 100
	list.Insert(attr)
	attr = NewAttrStretch(STRETCH_CONDENSED)
	attr.StartIndex = 30
	attr.EndIndex = 60
	list.Insert(attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,15]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[30,60]stretch=2\n")

	attr = NewAttrFamily("Times")
	attr.StartIndex = 10
	attr.EndIndex = 25
	list.Change(attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,25]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[30,60]stretch=2\n")

	attr = NewAttrFamily("Futura")
	attr.StartIndex = 11
	attr.EndIndex = 25
	list.Insert(attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,25]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[11,25]family=Futura\n"+
		"[30,60]stretch=2\n")
}

/* test something that gtk does */
func TestMerge(t *testing.T) {
	var list AttrList
	attr := NewAttrSize(10 * Scale)
	attr.StartIndex = 10
	attr.EndIndex = 11
	list.Insert(attr)
	attr = NewAttrRise(100)
	attr.StartIndex = 0
	attr.EndIndex = 200
	list.Insert(attr)
	attr = NewAttrFamily("Times")
	attr.StartIndex = 5
	attr.EndIndex = 15
	list.Insert(attr)
	attr = NewAttrFallback(false)
	attr.StartIndex = 11
	attr.EndIndex = 100
	list.Insert(attr)
	attr = NewAttrStretch(STRETCH_CONDENSED)
	attr.StartIndex = 30
	attr.EndIndex = 60
	list.Insert(attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,15]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[30,60]stretch=2\n")

	var list2 AttrList
	attr = NewAttrSize(10 * Scale)
	attr.StartIndex = 11
	attr.EndIndex = 13
	list2.Insert(attr)
	attr = NewAttrSize(11 * Scale)
	attr.StartIndex = 13
	attr.EndIndex = 15
	list2.Insert(attr)
	attr = NewAttrSize(12 * Scale)
	attr.StartIndex = 40
	attr.EndIndex = 50
	list2.Insert(attr)

	assert_attributes(t, list2, "[11,13]size=10240\n"+
		"[13,15]size=11264\n"+
		"[40,50]size=12288\n")

	list2.Filter(func(attr *Attribute) bool {
		list.Change(attr.copy())
		return false
	})

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,15]family=Times\n"+
		"[10,13]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[13,15]size=11264\n"+
		"[30,60]stretch=2\n"+
		"[40,50]size=12288\n")
}

// reproduce what the links example in gtk4-demo does
// with the colored Google link
func TestMerge2(t *testing.T) {
	var list AttrList
	attr := NewAttrUnderline(UNDERLINE_SINGLE)
	attr.StartIndex = 0
	attr.EndIndex = 10
	list.Insert(attr)
	attr = NewAttrForeground(AttrColor{0, 0, 0xffff})
	attr.StartIndex = 0
	attr.EndIndex = 10
	list.Insert(attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,10]foreground=#00000000ffff\n")

	attr = NewAttrForeground(AttrColor{0xffff, 0, 0})
	attr.StartIndex = 2
	attr.EndIndex = 3

	list.Change(attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,2]foreground=#00000000ffff\n"+
		"[2,3]foreground=#ffff00000000\n"+
		"[3,10]foreground=#00000000ffff\n")

	attr = NewAttrForeground(AttrColor{0, 0xffff, 0})
	attr.StartIndex = 3
	attr.EndIndex = 4

	list.Change(attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,2]foreground=#00000000ffff\n"+
		"[2,3]foreground=#ffff00000000\n"+
		"[3,4]foreground=#0000ffff0000\n"+
		"[4,10]foreground=#00000000ffff\n")

	attr = NewAttrForeground(AttrColor{0, 0, 0xffff})
	attr.StartIndex = 4
	attr.EndIndex = 5

	list.Change(attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,2]foreground=#00000000ffff\n"+
		"[2,3]foreground=#ffff00000000\n"+
		"[3,4]foreground=#0000ffff0000\n"+
		"[4,10]foreground=#00000000ffff\n")
}

func TestIterEpsilonZero(t *testing.T) {
	markup := "𝜀<span rise=\"-6000\" size=\"x-small\" font_desc=\"italic\">0</span> = 𝜔<span rise=\"8000\" size=\"smaller\">𝜔<span rise=\"14000\" size=\"smaller\">𝜔<span rise=\"20000\">.<span rise=\"23000\">.<span rise=\"26000\">.</span></span></span></span></span>"
	var s string
	ret, err := ParseMarkup([]byte(markup), 0)
	if err != nil {
		t.Fatal(err)
	}

	assertEquals(t, string(ret.Text), "𝜀0 = 𝜔𝜔𝜔...")

	attr := ret.Attr.getIterator()

	print_tags_for_attributes := func(iter *attrIterator) string {
		var out string
		attr := iter.getByKind(ATTR_RISE)
		if attr != nil {
			out += fmt.Sprintf("[%d, %d]rise=%s\n", attr.StartIndex, attr.EndIndex, attr.Data)
		}
		attr = iter.getByKind(ATTR_SIZE)
		if attr != nil {
			out += fmt.Sprintf("[%d, %d]size=%d\n", attr.StartIndex, attr.EndIndex, attr.Data)
		}
		attr = iter.getByKind(ATTR_SCALE)
		if attr != nil {
			out += fmt.Sprintf("[%d, %d]scale=%f\n", attr.StartIndex, attr.EndIndex, attr.Data)
		}
		return out
	}

	for do := true; do; do = attr.next() {
		s += fmt.Sprintf("range: [%d, %d]\n", attr.StartIndex, attr.EndIndex)
		s += print_tags_for_attributes(attr)
	}

	// the value here takes into account the bytes -> runes convention,
	// computed for example with the indexByteToIndexRune helper:
	// map[0:0 1:4 2:5 3:6 4:7 5:8 6:12 7:16 8:20 9:21 10:22 11:23]
	assertEquals(t, s, "range: [0, 1]\n"+
		"range: [1, 2]\n"+
		"[1, 2]rise=-6000\n"+
		"[1, 2]scale=0.694444\n"+
		"range: [2, 6]\n"+
		"range: [6, 7]\n"+
		"[6, 11]rise=8000\n"+
		"[6, 11]scale=0.833333\n"+
		"range: [7, 8]\n"+
		"[7, 11]rise=14000\n"+
		"[7, 11]scale=0.694444\n"+
		"range: [8, 9]\n"+
		"[8, 11]rise=20000\n"+
		"[7, 11]scale=0.694444\n"+
		"range: [9, 10]\n"+
		"[9, 11]rise=23000\n"+
		"[7, 11]scale=0.694444\n"+
		"range: [10, 11]\n"+
		"[10, 11]rise=26000\n"+
		"[7, 11]scale=0.694444\n"+
		"range: [11, 2147483647]\n")
}

func TestPrintOverflow(t *testing.T) {
	a := Attribute{Data: AttrInt(2), StartIndex: 0, EndIndex: MaxInt, Kind: ATTR_SHOW}
	if s := a.String(); s != "[0,2147483647]show=2" {
		t.Fatalf("unexpected attribute string: %s", s)
	}
}

// the C references tests have index in byte index
func AsByteIndex(text []rune, runeIndex int) int {
	if runeIndex == MaxInt {
		return MaxInt
	}
	return len(string(text[:runeIndex]))
}

// PrintAttributes returns a human friendly representation of the attributes.
func PrintAttributes(attrs AttrList, text []rune) string {
	var out string
	iter := attrs.getIterator()
	for do := true; do; {
		out += fmt.Sprintf("range %d %d\n", AsByteIndex(text, iter.StartIndex), AsByteIndex(text, iter.EndIndex))
		list := iter.getAttributes()
		for _, attr := range list {
			out += PrintAttribute(attr, text) + "\n"
		}
		do = iter.next()
	}
	return out
}
