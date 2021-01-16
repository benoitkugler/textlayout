package fontconfig

import (
	"fmt"
	"testing"
)

func test(query string, expect FcPattern) error {
	pat, err := FcNameParse([]byte(query))
	if err != nil {
		return err
	}
	if p, e := pat.Hash(), expect.Hash(); p != e {
		return fmt.Errorf("expected %s, got %s", expect, pat)
	}
	return nil
}

func TestParseName(t *testing.T) {
	var expect FcPattern

	expect = NewFcPattern()
	expect.Add(FC_FAMILY, String("sans-serif"), true)
	if err := test("sans\\-serif", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_FAMILY, String("Foo"), true)
	expect.Add(FC_SIZE, Int(10), true)
	if err := test("Foo-10", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_FAMILY, String("Foo"), true)
	expect.Add(FC_FAMILY, String("Bar"), true)
	expect.Add(FC_SIZE, Int(10), true)
	if err := test("Foo,Bar-10", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_FAMILY, String("Foo"), true)
	expect.Add(FC_WEIGHT, Int(FC_WEIGHT_MEDIUM), true)
	if err := test("Foo:weight=medium", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_FAMILY, String("Foo"), true)
	expect.Add(FC_WEIGHT, Int(FC_WEIGHT_MEDIUM), true)
	if err := test("Foo:weight_medium", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_WEIGHT, Int(FC_WEIGHT_MEDIUM), true)
	if err := test(":medium", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_WIDTH, Int(FC_WIDTH_NORMAL), true)
	if err := test(":normal", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_WIDTH, Int(FC_WIDTH_NORMAL), true)
	if err := test(":normal", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	r := FcRange{Begin: FC_WEIGHT_MEDIUM, End: FC_WEIGHT_BOLD}
	expect.Add(FC_WEIGHT, r, true)
	if err := test(":weight=[medium bold]", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	r = FcRange{Begin: 0.45, End: 48.88}
	expect.Add(FC_WEIGHT, r, true)
	if err := test(":weight=[0.45 48.88]", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_MATRIX, FcMatrix{1, 2.45, 3, 4.}, true)
	if err := test(":matrix=1 2.45 3. 4", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_ANTIALIAS, FcFalse, true)
	expect.Add(FC_AUTOHINT, FcTrue, true)
	expect.Add(FC_SCALABLE, FcDontCare, true)
	if err := test(":antialias=n:autohint=true:scalable=2", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewFcPattern()
	expect.Add(FC_PIXEL_SIZE, Float(45.78), true)
	expect.Add(FC_FOUNDRY, String("5456s4d"), true)
	expect.Add(FC_ORDER, Int(7845), true)
	if err := test(":pixelsize=45.78:foundry=5456s4d:order=7845", expect); err != nil {
		t.Fatal(err)
	}
}
