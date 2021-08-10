package fontconfig

import (
	"fmt"
	"testing"
)

func test(query string, expect Pattern) error {
	var c Config
	pat, err := c.parseName([]byte(query))
	if err != nil {
		return err
	}
	if p, e := pat.Hash(), expect.Hash(); p != e {
		return fmt.Errorf("expected %s, got %s", expect, pat)
	}
	return nil
}

func TestParseName(t *testing.T) {
	var expect Pattern

	expect = NewPattern()
	expect.Add(FAMILY, String("sans-serif"), true)
	if err := test("sans\\-serif", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(FAMILY, String("Foo"), true)
	expect.Add(SIZE, Int(10), true)
	if err := test("Foo-10", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(FAMILY, String("Foo"), true)
	expect.Add(FAMILY, String("Bar"), true)
	expect.Add(SIZE, Int(10), true)
	if err := test("Foo,Bar-10", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(FAMILY, String("Foo"), true)
	expect.Add(WEIGHT, Int(WEIGHT_MEDIUM), true)
	if err := test("Foo:weight=medium", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(FAMILY, String("Foo"), true)
	expect.Add(WEIGHT, Int(WEIGHT_MEDIUM), true)
	if err := test("Foo:weight_medium", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(WEIGHT, Int(WEIGHT_MEDIUM), true)
	if err := test(":medium", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(WIDTH, Int(WIDTH_NORMAL), true)
	if err := test(":normal", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(WIDTH, Int(WIDTH_NORMAL), true)
	if err := test(":normal", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	r := Range{Begin: WEIGHT_MEDIUM, End: WEIGHT_BOLD}
	expect.Add(WEIGHT, r, true)
	if err := test(":weight=[medium bold]", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	r = Range{Begin: 0.45, End: 48.88}
	expect.Add(WEIGHT, r, true)
	if err := test(":weight=[0.45 48.88]", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(MATRIX, Matrix{1, 2.45, 3, 4.}, true)
	if err := test(":matrix=1 2.45 3. 4", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(ANTIALIAS, False, true)
	expect.Add(AUTOHINT, True, true)
	expect.Add(SCALABLE, DontCare, true)
	if err := test(":antialias=n:autohint=true:scalable=2", expect); err != nil {
		t.Fatal(err)
	}

	expect = NewPattern()
	expect.Add(PIXEL_SIZE, Float(45.78), true)
	expect.Add(FOUNDRY, String("5456s4d"), true)
	expect.Add(ORDER, Int(7845), true)
	if err := test(":pixelsize=45.78:foundry=5456s4d:order=7845", expect); err != nil {
		t.Fatal(err)
	}
}
