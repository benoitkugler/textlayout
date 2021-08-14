package pango_test

import (
	"testing"

	"github.com/benoitkugler/textlayout/pango"
)

/* Note: The test expects that any newline sequence is of length 1
 * use \n (not \r\n) in the test texts.
 * I think the iterator itself should support \r\n without trouble,
 * but there are comments in layout-iter.c suggesting otherwise.
 */
var iterTestTexts = []string{
	/* English with embedded RTL runs (from ancient-hebrew.org) */
	"The Hebrew word \xd7\x90\xd7\x93\xd7\x9d\xd7\x94 (adamah) is the feminine form of \xd7\x90\xd7\x93\xd7\x9d meaning \"ground\"\n",
	/* Arabic, with vowel marks (from Sura Al Fatiha) */
	"\xd8\xa8\xd9\x90\xd8\xb3\xd9\x92\xd9\x85\xd9\x90 \xd8\xa7\xd9\x84\xd9\x84\xd9\x91\xd9\x87\xd9\x90 \xd8\xa7\xd9\x84\xd8\xb1\xd9\x91\xd9\x8e\xd8\xad\xd9\x92\xd9\x85\xd9\x80\xd9\x8e\xd9\x86\xd9\x90 \xd8\xa7\xd9\x84\xd8\xb1\xd9\x91\xd9\x8e\xd8\xad\xd9\x90\xd9\x8a\xd9\x85\xd9\x90\n\xd8\xa7\xd9\x84\xd9\x92\xd8\xad\xd9\x8e\xd9\x85\xd9\x92\xd8\xaf\xd9\x8f \xd9\x84\xd9\x84\xd9\x91\xd9\x87\xd9\x90 \xd8\xb1\xd9\x8e\xd8\xa8\xd9\x91\xd9\x90 \xd8\xa7\xd9\x84\xd9\x92\xd8\xb9\xd9\x8e\xd8\xa7\xd9\x84\xd9\x8e\xd9\x85\xd9\x90\xd9\x8a\xd9\x86\xd9\x8e\n",
	/* Arabic, with embedded LTR runs (from a Linux guide) */
	"\xd8\xa7\xd9\x84\xd9\x85\xd8\xaa\xd8\xba\xd9\x8a\xd8\xb1 LC_ALL \xd9\x8a\xd8\xba\xd9\x8a\xd9\x8a\xd8\xb1 \xd9\x83\xd9\x84 \xd8\xa7\xd9\x84\xd9\x85\xd8\xaa\xd8\xba\xd9\x8a\xd8\xb1\xd8\xa7\xd8\xaa \xd8\xa7\xd9\x84\xd8\xaa\xd9\x8a \xd8\xaa\xd8\xa8\xd8\xaf\xd8\xa3 \xd8\xa8\xd8\xa7\xd9\x84\xd8\xb1\xd9\x85\xd8\xb2 LC.",
	/* Hebrew, with vowel marks (from Genesis) */
	"\xd7\x91\xd6\xbc\xd6\xb0\xd7\xa8\xd6\xb5\xd7\x90\xd7\xa9\xd7\x81\xd6\xb4\xd7\x99\xd7\xaa, \xd7\x91\xd6\xbc\xd6\xb8\xd7\xa8\xd6\xb8\xd7\x90 \xd7\x90\xd6\xb1\xd7\x9c\xd6\xb9\xd7\x94\xd6\xb4\xd7\x99\xd7\x9d, \xd7\x90\xd6\xb5\xd7\xaa \xd7\x94\xd6\xb7\xd7\xa9\xd6\xbc\xd7\x81\xd6\xb8\xd7\x9e\xd6\xb7\xd7\x99\xd6\xb4\xd7\x9d, \xd7\x95\xd6\xb0\xd7\x90\xd6\xb5\xd7\xaa \xd7\x94\xd6\xb8\xd7\x90\xd6\xb8\xd7\xa8\xd6\xb6\xd7\xa5",
	/* Hebrew, with embedded LTR runs (from a Linux guide) */
	"\xd7\x94\xd7\xa7\xd7\x9c\xd7\x93\xd7\x94 \xd7\xa2\xd7\x9c \xd7\xa9\xd7\xa0\xd7\x99 \xd7\x94 SHIFT\xd7\x99\xd7\x9d (\xd7\x99\xd7\x9e\xd7\x99\xd7\x9f \xd7\x95\xd7\xa9\xd7\x9e\xd7\x90\xd7\x9c \xd7\x91\xd7\x99\xd7\x97\xd7\x93) \xd7\x90\xd7\x9e\xd7\x95\xd7\xa8\xd7\x99\xd7\x9d \xd7\x9c\xd7\x94\xd7\x93\xd7\x9c\xd7\x99\xd7\xa7 \xd7\x90\xd7\xaa \xd7\xa0\xd7\x95\xd7\xa8\xd7\xaa \xd7\x94 Scroll Lock , \xd7\x95\xd7\x9c\xd7\x94\xd7\xa2\xd7\x91\xd7\x99\xd7\xa8 \xd7\x90\xd7\x95\xd7\xaa\xd7\xa0\xd7\x95 \xd7\x9c\xd7\x9e\xd7\xa6\xd7\x91 \xd7\x9b\xd7\xaa\xd7\x99\xd7\x91\xd7\x94 \xd7\x91\xd7\xa2\xd7\x91\xd7\xa8\xd7\x99\xd7\xaa.",
	/* Different line terminators */
	"AAAA\nBBBB\nCCCC\n",
	"DDDD\rEEEE\rFFFF\r",
	"GGGG\r\nHHHH\r\nIIII\r\n",
	"asdf",
}

/* char iteration test:
 *  - Total num of iterations match number of chars
 *  - GlyphString's index_to_x positions match those returned by the Iter
 */
func testIterChar(t *testing.T, layout *pango.Layout) {
	text := layout.Text
	numChars := len(text)

	iter := layout.GetIter()
	iterNextOk := true

	for i := 0; i < numChars; i++ {
		if !iterNextOk {
			t.Fatalf("expected iterNextOk")
		}

		index := iter.Index
		// ptr := text[index:]
		// char_str = g_strndup(ptr, g_utf8_next_char(ptr)-ptr)
		// verbose("i=%d (visual), index = %d '%s':\n", i, index, char_str)

		extents := iter.GetCharExtents()
		// verbose("  char extents: x=%d,y=%d w=%d,h=%d\n", extents.X, extents.y, extents.width, extents.height)

		run := iter.GetRun()

		if run != nil {
			var runExtents pango.Rectangle

			/* Get needed data for the GlyphString */
			iter.GetRunExtents(nil, &runExtents)
			offset := run.Item.Offset
			// desc := run.Item.Analysis.Font.Describe(false)
			// str := desc.String()
			// rtl := run.Item.Analysis.Level % 2
			// verbose("  (current run: font=%s,offset=%d,x=%d,len=%d,rtl=%d)\n", str, offset, runExtents.X, run.item.length, rtl)

			/* Calculate expected x result using index_to_x */
			leadingX := run.Glyphs.IndexToX(text[offset:offset+run.Item.Length], &run.Item.Analysis,
				index-offset, false)
			trailingX := run.Glyphs.IndexToX(text[offset:offset+run.Item.Length], &run.Item.Analysis,
				index-offset, true)

			x0 := runExtents.X + min(leadingX, trailingX)
			x1 := runExtents.X + max(leadingX, trailingX)

			// verbose("  (index_to_x ind=%d: expected x=%d, width=%d)\n",
			// index-offset, x0, x1-x0)

			if !(extents.X == x0) {
				t.Fatalf("expected extents.X == x0, got %d %d", extents.X, x0)
			}
			if !(extents.Width == x1-x0) {
				t.Fatalf("expected extents.Width == x1-x0, got %d %d", extents.Width, x1-x0)
			}
		} else {
			/* We're on a line terminator */
		}

		iterNextOk = iter.NextChar()
		// verbose("more to go? %d\n", iterNextOk)
	}

	/* There should be one character position iterator for each character in the
	* input string */
	if iterNextOk {
		t.Fatalf("for text %s expected !iterNextOk", string(text))
	}
}

func min(a, b pango.GlyphUnit) pango.GlyphUnit {
	if a < b {
		return a
	}
	return b
}

func max(a, b pango.GlyphUnit) pango.GlyphUnit {
	if a > b {
		return a
	}
	return b
}

func testIterCluster(t *testing.T, layout *pango.Layout) {
	var (
		expectedNextX pango.GlyphUnit
		lastLine      *pango.LayoutLine
	)
	iter := layout.GetIter()
	iterNextOk := true

	for iterNextOk {
		line := iter.GetLine()

		/* Every cluster is part of a run */
		if iter.GetRun() == nil {
			t.Fatal("expected non nil run")
		}

		var extents pango.Rectangle
		iter.GetClusterExtents(nil, &extents)

		iterNextOk = iter.NextCluster()

		// index := iter.Index
		//    verbose ("index = %d:\n", index);
		//    verbose ("  cluster extents: x=%d,y=%d w=%d,h=%d\n",
		// 		extents.X, extents.y,
		// 		extents.Width, extents.height);
		//    verbose ("more to go? %d\n", iterNextOk);

		/* All the clusters on a line should be next to each other and occupy
		* the entire line. They advance linearly from left to right */
		if !(extents.Width >= 0) {
			t.Fatalf("expected extents.Width >= 0, got %d", extents.Width)
		}

		if lastLine == line {
			if !(extents.X == expectedNextX) {
				t.Fatalf("expected extents.X == expectedNextX, got %d %d", extents.X, expectedNextX)
			}
		}

		expectedNextX = extents.X + extents.Width

		lastLine = line
	}

	if iterNextOk {
		t.Fatal("expected !iterNextOk")
	}
}

func TestLayoutIter(t *testing.T) {
	const LAYOUT_WIDTH = 80 * pango.Scale

	context := pango.NewContext(newChachedFontMap())

	desc := pango.NewFontDescriptionFrom("Cantarell 11")
	context.SetFontDescription(desc)
	layout := pango.NewLayout(context)

	layout.SetWidth(LAYOUT_WIDTH)

	for _, text := range iterTestTexts {
		// verbose ("--------- checking next text ----------\n");
		// verbose (" <%s>\n", *ptext);
		// verbose ( "len=%ld, bytes=%ld\n", (long)g_utf8_strlen (*ptext, -1), (long)strlen (*ptext));

		layout.SetText(text)
		testIterChar(t, layout)
		testIterCluster(t, layout)
	}
}

func TestGlyphItemIter(t *testing.T) {
	context := pango.NewContext(newChachedFontMap())

	desc := pango.NewFontDescriptionFrom("Cantarell 11")
	context.SetFontDescription(desc)
	layout := pango.NewLayout(context)

	/* This shouldn't form any ligatures. */
	layout.SetText("test تست")
	text := layout.Text

	line := layout.GetLine(0)
	for l := line.Runs; l != nil; l = l.Next {
		run := l.Data
		for direction := 0; direction < 2; direction++ {
			var iter pango.GlyphItemIter
			have_cluster := iter.InitEnd(run, text)
			if direction != 0 {
				have_cluster = iter.InitStart(run, text)
			}
			for have_cluster {
				if direction != 0 {
					have_cluster = iter.NextCluster()
				} else {
					have_cluster = iter.PrevCluster()
				}
				// verbose("start index %d end index %d\n", iter.StartIndex, iter.EndIndex)
				if !(iter.StartIndex < iter.EndIndex) {
					t.Fatalf("expected iter.StartIndex < iter.EndIndex, got %d %d", iter.StartIndex, iter.EndIndex)
				}
				if !(iter.StartIndex+2 >= iter.EndIndex) {
					t.Fatalf("expected iter.StartIndex+2 >= iter.EndIndex, got %d %d", iter.StartIndex, iter.EndIndex)
				}
				if !(iter.StartChar+1 == iter.EndChar) {
					t.Fatalf("expected iter.StartChar+1 == iter.EndChar, got %d %d", iter.StartChar, iter.EndChar)
				}
			}
		}
	}
}
