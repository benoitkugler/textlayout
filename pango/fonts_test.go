package pango

import (
	"testing"
)

func TestParse(t *testing.T) {
	desc := NewFontDescriptionFrom("Cantarell 14")

	assertEquals(t, desc.FamilyName, "Cantarell")
	assertFalse(t, desc.SizeIsAbsolute, "font size is not absolute")
	assertEquals(t, desc.Size, int32(14*Scale))
	assertEquals(t, desc.Style, STYLE_NORMAL)
	assertEquals(t, desc.Variant, VARIANT_NORMAL)
	assertEquals(t, desc.Weight, WEIGHT_NORMAL)
	assertEquals(t, desc.Stretch, STRETCH_NORMAL)
	assertEquals(t, desc.Gravity, GRAVITY_SOUTH)
	assertEquals(t, desc.mask, FmFamily|FmStyle|FmVariant|FmWeight|FmStretch|FmSize)

	desc = NewFontDescriptionFrom("Sans Bold Italic Condensed 22.5px")

	assertEquals(t, desc.FamilyName, "Sans")
	assertTrue(t, desc.SizeIsAbsolute, "font size is absolute")
	assertEquals(t, desc.Size, int32(225*Scale/10))
	assertEquals(t, desc.Style, STYLE_ITALIC)
	assertEquals(t, desc.Variant, VARIANT_NORMAL)
	assertEquals(t, desc.Weight, WEIGHT_BOLD)
	assertEquals(t, desc.Stretch, STRETCH_CONDENSED)
	assertEquals(t, desc.Gravity, GRAVITY_SOUTH)
	assertEquals(t, desc.mask, FmFamily|FmStyle|FmVariant|FmWeight|FmStretch|FmSize)
}

func TestRoundtrip(t *testing.T) {
	desc := NewFontDescriptionFrom("Cantarell 14")
	str := desc.String()
	assertEquals(t, str, "Cantarell 14")

	desc = NewFontDescriptionFrom("Sans Bold Italic Condensed 22.5px")
	str = desc.String()
	assertEquals(t, str, "Sans Bold Italic Condensed 22.5px")
}

func TestVariation(t *testing.T) {
	desc1 := NewFontDescriptionFrom("Cantarell 14")
	assertTrue(t, desc1.mask&FmVariations == 0, "no variations")
	assertTrue(t, desc1.Variations == "", "variations is empty")

	str := desc1.String()
	assertEquals(t, str, "Cantarell 14")

	desc2 := NewFontDescriptionFrom("Cantarell 14 @wght=100,wdth=235")
	assertTrue(t, desc2.mask&FmVariations != 0, "has variations")
	assertEquals(t, desc2.Variations, "wght=100,wdth=235")

	str = desc2.String()
	assertEquals(t, str, "Cantarell 14 @wght=100,wdth=235")

	assertFalse(t, desc1.pango_font_description_equal(desc2), "different font descriptions")

	desc1.SetVariations("wght=100,wdth=235")
	assertTrue(t, desc1.mask&FmVariations != 0, "has variations")
	assertEquals(t, desc1.Variations, "wght=100,wdth=235")

	assertTrue(t, desc1.pango_font_description_equal(desc2), "same fonts")
}
