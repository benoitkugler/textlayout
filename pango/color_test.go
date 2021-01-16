package pango

import (
	"testing"
)

type colorSpec struct {
	spec         string
	valid        bool
	colorOrAlpha int
	red          uint16
	green        uint16
	blue         uint16
	alpha        uint16
}

const (
	COLOR = 1
	ALPHA = 2
	BOTH  = 3
)

func testOneColor(t *testing.T, spec colorSpec) {
	//   PangoColor color;
	//   bool accepted;
	//   uint16 alpha;

	if spec.colorOrAlpha&COLOR != 0 {
		color, accepted := pango_color_parse(spec.spec)
		if !spec.valid {
			assertFalse(t, accepted, "color is invalid")
		} else {
			assertTrue(t, accepted, "color is valid")
			assertEquals(t, color.Red, spec.red)
			assertEquals(t, color.Green, spec.green)
			assertEquals(t, color.Blue, spec.blue)
		}
	}

	if spec.colorOrAlpha&ALPHA != 0 {
		color, alpha, accepted := pango_color_parse_with_alpha(spec.spec, true)
		if !spec.valid {
			assertFalse(t, accepted, "color is invalid")
		} else {
			assertTrue(t, accepted, "color is valid")
			assertEquals(t, color.Red, spec.red)
			assertEquals(t, color.Green, spec.green)
			assertEquals(t, color.Blue, spec.blue)
			assertEquals(t, alpha, spec.alpha)
		}
	}
}

var specs = [...]colorSpec{
	{"#abc", true, BOTH, 0xaaaa, 0xbbbb, 0xcccc, 0xffff},
	{"#aabbcc", true, BOTH, 0xaaaa, 0xbbbb, 0xcccc, 0xffff},
	{"#aaabbbccc", true, BOTH, 0xaaaa, 0xbbbb, 0xcccc, 0xffff},
	{"#100100100", true, BOTH, 0x1001, 0x1001, 0x1001, 0xffff},
	{"#aaaabbbbcccc", true, COLOR, 0xaaaa, 0xbbbb, 0xcccc, 0xffff},
	{"#fff", true, BOTH, 0xffff, 0xffff, 0xffff, 0xffff},
	{"#ffffff", true, BOTH, 0xffff, 0xffff, 0xffff, 0xffff},
	{"#fffffffff", true, BOTH, 0xffff, 0xffff, 0xffff, 0xffff},
	{"#ffffffffffff", true, COLOR, 0xffff, 0xffff, 0xffff, 0xffff},
	{"#000", true, BOTH, 0x0000, 0x0000, 0x0000, 0xffff},
	{"#000000", true, BOTH, 0x0000, 0x0000, 0x0000, 0xffff},
	{"#000000000", true, BOTH, 0x0000, 0x0000, 0x0000, 0xffff},
	{"#000000000000", true, COLOR, 0x0000, 0x0000, 0x0000, 0xffff},
	{"#AAAABBBBCCCC", true, COLOR, 0xaaaa, 0xbbbb, 0xcccc, 0xffff},
	{"#aa bb cc ", false, BOTH, 0, 0, 0, 0},
	{"#aa bb ccc", false, BOTH, 0, 0, 0, 0},
	{"#ab", false, BOTH, 0, 0, 0, 0},
	{"#aabb", false, COLOR, 0, 0, 0, 0},
	{"#aaabb", false, BOTH, 0, 0, 0, 0},
	{"aaabb", false, BOTH, 0, 0, 0, 0},
	{"", false, BOTH, 0, 0, 0, 0},
	{"#", false, BOTH, 0, 0, 0, 0},
	{"##fff", false, BOTH, 0, 0, 0, 0},
	{"#0000ff+", false, BOTH, 0, 0, 0, 0},
	{"#0000f+", false, BOTH, 0, 0, 0, 0},
	{"#0x00x10x2", false, BOTH, 0, 0, 0, 0},
	{"#abcd", true, ALPHA, 0xaaaa, 0xbbbb, 0xcccc, 0xdddd},
	{"#aabbccdd", true, ALPHA, 0xaaaa, 0xbbbb, 0xcccc, 0xdddd},
	{"#aaaabbbbccccdddd", true, ALPHA, 0xaaaa, 0xbbbb, 0xcccc, 0xdddd},
	{"grey100", true, COLOR, 0xffff, 0xffff, 0xffff, 0xffff},
	{"darkGrey", true, COLOR, 169 * 65535 / 255, 169 * 65535 / 255, 169 * 65535 / 255, 0xffff},
	{"invalidName", false, COLOR, 0, 0, 0, 0},
}

func TestColor(t *testing.T) {
	for _, spec := range specs {
		testOneColor(t, spec)
	}
}
