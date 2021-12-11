package pango_test

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/benoitkugler/textlayout/pango"
)

// Computes the logical and ink extents of @layoutLine in device units.
// This function just calls GetExtents() followed by
// two extentsToPixels() calls, rounding @inkRect and @logicalRect
// such that the rounded rectangles fully contain the unrounded one (that is,
// passes them as first argument to extentsToPixels()).
func layoutLineGetPixelExtents(layoutLine *pango.LayoutLine, inkRect, logicalRect *pango.Rectangle) {
	layoutLine.GetExtents(inkRect, logicalRect)
	extentsToPixels(inkRect, nil)
	extentsToPixels(logicalRect, nil)
}

type layoutParams struct {
	width           int
	height          int
	indent          int
	spacing         int
	lineSpacing     float32
	ellipsizeAt     int
	ellipsize       pango.EllipsizeMode
	wrap            pango.WrapMode
	alignment       pango.Alignment
	justify         bool
	autoDir         bool
	singleParagraph bool
	gravity         pango.Gravity
	//  *tabs PangoTabArray
}

func newParams() layoutParams {
	return layoutParams{
		width:           -1,
		height:          -1,
		indent:          0,
		spacing:         0,
		lineSpacing:     0.0,
		ellipsize:       pango.ELLIPSIZE_NONE,
		wrap:            pango.WRAP_WORD,
		alignment:       pango.ALIGN_LEFT,
		justify:         false,
		autoDir:         true,
		singleParagraph: false,
		gravity:         pango.GRAVITY_AUTO,
		// tabs:             nil,
	}
}

func parseEllipsizeMode(v string) pango.EllipsizeMode {
	switch strings.ToLower(v) {
	case "none":
		return pango.ELLIPSIZE_NONE
	case "start":
		return pango.ELLIPSIZE_START
	case "middle":
		return pango.ELLIPSIZE_MIDDLE
	case "end":
		return pango.ELLIPSIZE_END
	default:
		return 0
	}
}

func parseWrapMode(v string) pango.WrapMode {
	switch strings.ToLower(v) {
	case "word":
		return pango.WRAP_WORD
	case "char":
		return pango.WRAP_CHAR
	case "word_char":
		return pango.WRAP_WORD_CHAR
	default:
		return 0
	}
}

func parseAlignment(v string) pango.Alignment {
	switch strings.ToLower(v) {
	case "left":
		return pango.ALIGN_LEFT
	case "center":
		return pango.ALIGN_CENTER
	case "right":
		return pango.ALIGN_RIGHT
	default:
		return 0
	}
}

func parseParams(params string) layoutParams {
	out := newParams()

	if strings.TrimSpace(params) == "" {
		return out
	}
	options := strings.Split(params, ",")
	for _, option := range options {
		chunks := strings.Split(option, "=")
		name, value := chunks[0], chunks[1]
		switch name {
		case "width":
			out.width, _ = strconv.Atoi(value)
		case "height":
			out.height, _ = strconv.Atoi(value)
		case "indent":
			out.indent, _ = strconv.Atoi(value)
		case "spacing":
			out.spacing, _ = strconv.Atoi(value)
		case "line_spacing":
			f, _ := strconv.ParseFloat(value, 32)
			out.lineSpacing = float32(f)
		case "ellipsize-at":
			out.ellipsizeAt, _ = strconv.Atoi(value)
		case "ellipsize":
			out.ellipsize = parseEllipsizeMode(value)
		case "wrap":
			out.wrap = parseWrapMode(value)
		case "alignment":
			out.alignment = parseAlignment(value)
		case "gravity":
			g, _ := pango.GravityMap.FromString(value)
			out.gravity = pango.Gravity(g)
		case "justify":
			out.justify = value == "true"
		case "auto_dir":
			out.autoDir = value == "true"
		case "single_paragraph":
			out.singleParagraph = value == "true"
		}
	}
	return out
}

func dumpLines(layout *pango.Layout) string {
	var out string

	text := layout.Text
	iter := layout.GetIter()

	index := iter.Index
	indexEnd := 0
	i := 0
	for hasMore := true; hasMore; {
		line := iter.GetLine()
		hasMore = iter.NextLine()
		i++

		var charStr []rune
		if hasMore {
			indexEnd = iter.Index
			charStr = text[index:indexEnd]
		} else {
			charStr = text[index:]
		}

		indexByte := pango.AsByteIndex(text, index)
		out += fmt.Sprintf("i=%d, index=%d, paragraph-start=%d, dir=%s '%s'\n",
			i, indexByte, boolToInt(line.IsParagraphStart), line.ResolvedDir, string(charStr))

		index = indexEnd
	}

	return out
}

func dumpRuns(layout *pango.Layout) string {
	var out string

	text := layout.Text
	iter := layout.GetIter()

	i := 0
	for hasMore := true; hasMore; hasMore = iter.NextRun() {
		run := iter.GetRun()
		index := iter.Index
		i++

		// the reference tests have index in byte
		indexByte := len(string(text[:index]))

		if run != nil {
			item := run.Item
			charStr := string(text[item.Offset : item.Offset+item.Length])
			// font := item.Analysis.Font.Describe(false).String()
			out += fmt.Sprintf("i=%d, index=%d, chars=%d, level=%d, gravity=%s, flags=%d, font=%s, script=%s, language=%s, '%s'\n",
				i, indexByte, item.Length, item.Analysis.Level, item.Analysis.Gravity,
				item.Analysis.Flags,
				"OMITTED", /* for some reason, this fails on build.gnome.org, so leave it out */
				strings.ToLower(item.Analysis.Script.String()), item.Analysis.Language, charStr)
			out += printAttributes(item.Analysis.ExtraAttrs, text)
		} else {
			out += fmt.Sprintf("i=%d, index=%d, no run, line end\n",
				i, indexByte)
		}
	}

	return out
}

func dumpDirections(layout *pango.Layout) string {
	out := ""
	for i := range layout.Text {
		out += fmt.Sprintf("%d ", layout.GetDirectionAt(i))
	}
	out += "\n"
	return out
}

func dumpCursorPositions(layout *pango.Layout) string {
	index := 0
	trailing := 0
	out := ""
	for index < math.MaxInt32 {
		out += fmt.Sprintf("%d(%d) ", pango.AsByteIndex(layout.Text, index), trailing)

		index += trailing

		index, trailing = layout.MoveCursorVisually(true, index, 0, 1)
	}

	out += "\n"
	return out
}

func printAttributes(attrs pango.AttrList, text []rune) string {
	chunks := make([]string, len(attrs))
	for i, attr := range attrs {
		chunks[i] = pango.PrintAttribute(attr, text) + "\n"
	}
	return strings.Join(chunks, "")
}

func assertRectangleSizeContained(t *testing.T, r1, r2 pango.Rectangle, msg string) {
	assertTrue(t, r1.Width <= r2.Width && r1.Height <= r2.Height,
		fmt.Sprintf("%s expected contained size, got r1: %v, r2: %v", msg, r1, r2))
}

func assertRectangleContained(t *testing.T, r1, r2 *pango.Rectangle) {
	assertTrue(t, r1.X >= r2.X &&
		r1.Y >= r2.Y &&
		r1.X+r1.Width <= r2.X+r2.Width &&
		r1.Y+r1.Height <= r2.Y+r2.Height, "expected contained rectangle")
}

func testInternalLayout(t *testing.T, layout *pango.Layout) {
	var inkRect, logicalRect pango.Rectangle
	layout.GetExtents(&inkRect, &logicalRect)
	extentsToPixels(&inkRect, nil)

	lines := layout.GetLinesReadonly()
	for _, line := range lines {
		var lineInk, lineLogical, lineInk1, lineLogical1 pango.Rectangle

		line.GetExtents(&lineInk, &lineLogical)
		lineX := lineLogical.X
		lineWidth := lineLogical.Width
		extentsToPixels(&lineInk, nil)
		extentsToPixels(&lineLogical, nil)
		layoutLineGetPixelExtents(line, &lineInk1, &lineLogical1)

		/* Not in layout coordinates, so just compare sizes */
		assertRectangleSizeContained(t, lineInk, inkRect, "line ink")
		assertRectangleSizeContained(t, lineLogical, logicalRect, "line logical")
		assertRectangleSizeContained(t, lineInk1, inkRect, "line ink 1")
		assertRectangleSizeContained(t, lineLogical1, logicalRect, "line logical 1")

		if layout.IsEllipsized() {
			continue
		}

		iter := layout.GetIter()
		for iter.GetLine() != line {
			iter.NextLine()
		}
		for done := false; !done && iter.GetLine() == line; {
			index := iter.Index
			run := iter.GetRun()

			if !iter.NextCluster() {
				done = true
			}

			x := line.IndexToX(index, false)
			assertTrue(t, x >= 0 && x <= lineWidth, fmt.Sprintf("unexpected x pos: %d", x))

			if run == nil {
				break
			}

			prevIndex := run.Item.Offset
			nextIndex := run.Item.Offset + run.Item.Length

			ranges := line.GetXRanges(prevIndex, nextIndex)

			/* The index is within the run, so the x should be in one of the ranges */
			var foundRange bool
			if NR := len(ranges); NR > 0 {
				for k := 0; k < NR; k++ {
					if x+lineX >= ranges[2*k] && x+lineX <= ranges[2*k+1] {
						foundRange = true
						break
					}
				}
			}

			assertTrue(t, foundRange, "expected to find range")
		}
	}

	layout.GetExtents(&inkRect, &logicalRect)

	iter := layout.GetIter()
	assertTrue(t, iter.Index == 0, fmt.Sprintf("unexpected iter index %d", iter.Index))
	for do := true; do; do = iter.NextLine() {
		line := iter.GetLine()
		var lineInk, lineLogical pango.Rectangle
		iter.GetLineExtents(&lineInk, &lineLogical)
		baseline := iter.GetBaseline()

		assertRectangleContained(t, &lineInk, &inkRect)
		assertRectangleContained(t, &lineLogical, &logicalRect)

		assertTrue(t, lineLogical.Y <= baseline, "unexpected baseline")
		assertTrue(t, baseline <= lineLogical.Y+lineLogical.Height, "unexpected baseline")

		if iter.IsAtLastLine() {
			assertTrue(t, line.StartIndex+line.Length <= len(layout.Text), "unexpected last line")
		}

		run := iter.GetRun()
		if run != nil {
			text := layout.Text
			widths := run.GetLogicalWidths(text)
			assertTrue(t, len(widths) == run.Item.Length, "unexpected length")
		}
	}
}

func testOneLayout(t *testing.T, filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	context := pango.NewContext(newChachedFontMap())

	chunks := strings.SplitN(string(content), "\n", 2)
	params, markup := chunks[0], chunks[1]
	if strings.HasPrefix(params, "#") {
		params = ""
	}

	options := parseParams(params)

	layout := pango.NewLayout(context)
	desc := pango.NewFontDescriptionFrom("Cantarell 11")
	layout.SetFontDescription(&desc)

	err = layout.SetMarkup([]byte(markup))
	if err != nil {
		t.Fatal(err)
	}

	context.SetBaseGravity(options.gravity)
	if options.width > 0 {
		layout.SetWidth(pango.GlyphUnit(options.width) * pango.Scale)
	} else {
		layout.SetWidth(-1)
	}
	if options.height > 0 {
		layout.SetHeight(pango.GlyphUnit(options.height) * pango.Scale)
	} else {
		layout.SetHeight(pango.GlyphUnit(options.height))
	}
	layout.SetIndent(pango.GlyphUnit(options.indent) * pango.Scale)
	layout.SetEllipsize(options.ellipsize)
	layout.SetWrap(options.wrap)
	layout.SetSpacing(pango.GlyphUnit(options.spacing) * pango.Scale)
	layout.SetLineSpacing(options.lineSpacing)
	layout.SetAlignment(options.alignment)
	layout.SetJustify(options.justify)
	layout.SetAutoDir(options.autoDir)
	layout.SetSingleParagraphMode(options.singleParagraph)

	// SetTabs(layout, options.tabs)

	// testInternalLayout(t, layout)

	return generateLayoutDump(layout, options)
}

func generateLayoutDump(layout *pango.Layout, params layoutParams) string {
	out := string(layout.Text)
	out += "\n--- parameters\n\n"
	out += fmt.Sprintf("wrapped: %d\n", boolToInt(layout.IsWrapped()))
	out += fmt.Sprintf("ellipsized: %d\n", boolToInt(layout.IsEllipsized()))
	out += fmt.Sprintf("lines: %d\n", layout.GetLineCount())
	if params.width > 0 {
		out += fmt.Sprintf("width: %d\n", layout.Width)
	}
	if params.height > 0 {
		out += fmt.Sprintf("height: %d\n", layout.Height)
	}
	if params.indent != 0 {
		out += fmt.Sprintf("indent: %d\n", layout.Indent)
	}

	out += "\n--- attributes\n\n"
	out += pango.PrintAttributes(layout.Attributes, layout.Text)

	out += "\n--- directions\n\n"
	out += dumpDirections(layout)

	out += "\n--- cursor positions\n\n"
	out += dumpCursorPositions(layout)

	out += "\n--- lines\n\n"
	out += dumpLines(layout)
	out += "\n--- runs\n\n"
	out += dumpRuns(layout)

	return out
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestLayout(t *testing.T) {
	if err := os.Setenv("LC_ALL", "en_US.utf8"); err != nil {
		t.Fatal(err)
	}
	files, err := ioutil.ReadDir("test/layouts")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		// test failing in pango C reference
		switch file.Name() {
		case "valid-19.markup", "valid-21.markup", "valid-22.markup":
			continue
		}

		if !strings.HasSuffix(file.Name(), ".markup") {
			continue
		}

		got := testOneLayout(t, filepath.Join("test/layouts", file.Name()))

		expFilename := strings.TrimSuffix(file.Name(), ".markup") + ".expected"
		exp, err := ioutil.ReadFile(filepath.Join("test/layouts", expFilename))
		if err != nil {
			t.Fatal(err)
		}

		if got != string(exp) {
			t.Fatalf("file %s: expected\n%s\n got\n%s", expFilename, exp, got)
		}
	}
}

func TestShapeTabCrash(t *testing.T) {
	// test that we don't crash in shape_tab when the layout
	// is such that we don't have effective attributes
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)
	layout.SetText("one\ttwo")
	layout.IsEllipsized()
}

func TestItemizeEmpty(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	context.Itemize(nil, 0, 0, nil)
}

func TestIterExtents(t *testing.T) {
	tests := []string{
		"Some long text that has multiple lines that are wrapped by Pango.",
	}

	context := pango.NewContext(newChachedFontMap())

	for _, test := range tests {
		layout := pango.NewLayout(context)
		layout.SetText(test)
		layout.SetWidth(60 * pango.Scale)

		var layoutExtents, lineExtents, runExtents,
			clusterExtents, charExtents, pos pango.Rectangle

		layout.GetExtents(nil, &layoutExtents)

		iter := layout.GetIter()
		for do := true; do; do = iter.NextChar() {
			iter.GetLineExtents(nil, &lineExtents)
			iter.GetRunExtents(nil, &runExtents)
			iter.GetClusterExtents(nil, &clusterExtents)
			charExtents = iter.GetCharExtents()

			layout.IndexToPos(iter.Index, &pos)
			if pos.Width < 0 {
				pos.X += pos.Width
				pos.Width = -pos.Width
			}

			assertRectangleContained(t, &lineExtents, &layoutExtents)
			assertRectangleContained(t, &runExtents, &lineExtents)
			assertRectangleContained(t, &clusterExtents, &runExtents)
			assertRectangleContained(t, &charExtents, &clusterExtents)

			assertRectangleContained(t, &pos, &lineExtents)
		}
	}
}

func TestEmptyLineHeight(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	description := pango.NewFontDescription()

	var ext1, ext2, ext3 pango.Rectangle
	for size := int32(10); size <= 20; size++ {
		description.SetSize(size)

		layout := pango.NewLayout(context)
		layout.SetFontDescription(&description)
		layout.GetExtents(nil, &ext1)
		layout.SetText("a")
		layout.GetExtents(nil, &ext2)
		assertTrue(t, ext1.Height == ext2.Height, "")

		layout.SetText("Pg")
		layout.GetExtents(nil, &ext3)
		assertTrue(t, ext2.Height == ext3.Height, "")
	}
}

func TestWrapChar(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)
	layout.SetText("Rows can have suffix widgets")
	layout.SetWrap(pango.WRAP_WORD_CHAR)

	layout.SetWidth(0)
	w0, h0 := layout.GetSize()

	layout.SetWidth(w0)
	w, h := layout.GetSize()

	assertTrue(t, w0 == w, "")
	assertTrue(t, h0 >= h, "")
}

func TestSmallCapsCrash(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)
	desc := pango.NewFontDescriptionFrom("Cantarell Small-Caps 11")
	layout.SetFontDescription(&desc)

	layout.SetText("Pere Ràfols Soler\nEqualiser, LV2\nAudio: 1, 1\nMidi: 0, 0\nControls: 53, 2\nCV: 0, 0")

	w, h := layout.GetSize()
	assertTrue(t, w > h, "")
}

// FIXME:
func TestHeightAndBaseline(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())
	layout := pango.NewLayout(context)
	desc := pango.NewFontDescriptionFrom("Helvetica,Apple Color Emoji 36px")
	layout.SetText("Go 1.17 Release Notes")
	layout.SetFontDescription(&desc)

	line := layout.GetLine(0)
	var rect pango.Rectangle
	line.GetExtents(nil, &rect)
	baseline := layout.GetBaseline()

	var expHeight, expBaseline pango.GlyphUnit = 45056, 34816

	if len([]rune("Go 1.17 Release Notes")) != len([]byte("Go 1.17 Release Notes")) {
		t.Fatal("should be ASCII content")
	}

	if rect.Height != expHeight {
		t.Fatalf("height: expected %d, got %d", expHeight, rect.Height)
	}
	if baseline != expBaseline {
		t.Fatalf("height: expected %d, got %d", expBaseline, baseline)
	}
}
