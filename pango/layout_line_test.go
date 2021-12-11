package pango_test

import (
	"testing"

	"github.com/benoitkugler/textlayout/pango"
)

func TestLineHeight(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)

	layout.SetText("one\ttwo")
	line := layout.GetLine(0)
	height := line.GetExtents(nil, nil)
	assertTrue(t, height > 0, "")
}

func TestLineHeight2(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)

	layout.SetText("one")

	line := layout.GetLine(0)
	height1 := line.GetExtents(nil, nil)

	layout.SetText("")

	line = layout.GetLine(0)
	height2 := line.GetExtents(nil, nil)

	assertTrue(t, height1 == height2, "")
}

func TestLineHeight3(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)

	layout.SetText("one")

	var attrs pango.AttrList
	attrs.Insert(pango.NewAttrLineHeight(2))
	layout.SetAttributes(attrs)

	line := layout.GetLine(0)
	height1 := line.GetExtents(nil, nil)

	layout.SetText("")

	line = layout.GetLine(0)
	height2 := line.GetExtents(nil, nil)
	assertTrue(t, height1 == height2, "")
}

func TestRunHeight(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)

	layout.SetText("one")

	var logical1, logical2 pango.Rectangle
	iter := layout.GetIter()
	iter.GetRunExtents(nil, &logical1)

	layout.SetText("")

	iter = layout.GetIter()
	iter.GetRunExtents(nil, &logical2)

	assertTrue(t, logical1.Height == logical2.Height, "")
}

func TestCursorHeight(t *testing.T) {
	var strong pango.Rectangle

	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)

	layout.SetText("one\ttwo")
	layout.GetCursorPos(0, &strong, nil)

	assertTrue(t, strong.Height > 0, "")
}

func TestCursorHeight2(t *testing.T) {
	var strong1, strong2 pango.Rectangle

	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)

	layout.SetText("one")
	layout.GetCursorPos(0, &strong1, nil)

	layout.SetText("")
	layout.GetCursorPos(0, &strong2, nil)

	assertTrue(t, strong1.Height == strong2.Height, "")
}
