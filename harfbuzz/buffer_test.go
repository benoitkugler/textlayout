package harfbuzz

import "testing"

// ported from harfbuzz/test/api/test-buffer.c Copyright © 2011  Google, Inc. Behdad Esfahbod

var utf32 = [7]rune{'a', 'b', 0x20000, 'd', 'e', 'f', 'g'}

const (
	bufferEmpty = iota
	bufferOneByOne
	bufferUtf32
	bufferNumTypes
)

func newTestBuffer(kind int) *Buffer {
	b := NewBuffer()

	switch kind {
	case bufferEmpty:

	case bufferOneByOne:
		for i := 1; i < len(utf32)-1; i++ {
			b.AddRune(utf32[i], i)
		}

	case bufferUtf32:
		b.AddRunes(utf32[:], 1, len(utf32)-2)

	}
	return b
}

func testBufferProperties(b *Buffer, t *testing.T) {
	/* test default properties */

	assert(t, b.Props.Direction == 0)
	assert(t, b.Props.Script == 0)
	assert(t, b.Props.Language == "")

	b.Props.Language = NewLanguage("fa")
	assert(t, b.Props.Language == NewLanguage("Fa"))

	/* test clear_contents clears all these properties: */

	//    hb_buffer_clear_contents (b);

	//    assert (t,hb_buffer_get_unicode_funcs (b) == ufuncs);
	//    assert (t,hb_buffer_get_direction (b) == HB_DIRECTION_INVALID);
	//    assert (t,hb_buffer_get_script (b) == HB_SCRIPT_INVALID);
	//    assert (t,hb_buffer_get_language (b) == NULL);

	/* but not these: */

	//    assert (t,hb_buffer_get_flags (b) != HB_BUFFER_FLAGS_DEFAULT);
	//    assert (t,hb_buffer_get_replacement_codepoint (b) != HB_BUFFER_REPLACEMENT_CODEPOINT_DEFAULT);

	/* test reset clears all properties */

	//    hb_buffer_set_direction (b, HB_DIRECTION_RTL);
	//    assert (t,hb_buffer_get_direction (b) == HB_DIRECTION_RTL);

	//    hb_buffer_set_script (b, HB_SCRIPT_ARABIC);
	//    assert (t,hb_buffer_get_script (b) == HB_SCRIPT_ARABIC);

	//    hb_buffer_set_language (b, hb_language_from_string ("fa", -1));
	//    assert (t,hb_buffer_get_language (b) == hb_language_from_string ("Fa", -1));

	//    hb_buffer_set_flags (b, HB_BUFFER_FLAG_BOT);
	//    assert (t,hb_buffer_get_flags (b) == HB_BUFFER_FLAG_BOT);

	//    hb_buffer_set_replacement_codepoint (b, (unsigned int) -1);
	//    assert (t,hb_buffer_get_replacement_codepoint (b) == (unsigned int) -1);

	//    hb_buffer_reset (b);

	//    assert (t,hb_buffer_get_unicode_funcs (b) == hb_unicode_funcs_get_default ());
	//    assert (t,hb_buffer_get_direction (b) == HB_DIRECTION_INVALID);
	//    assert (t,hb_buffer_get_script (b) == HB_SCRIPT_INVALID);
	//    assert (t,hb_buffer_get_language (b) == NULL);
	//    assert (t,hb_buffer_get_flags (b) == HB_BUFFER_FLAGS_DEFAULT);
	//    assert (t,hb_buffer_get_replacement_codepoint (b) == HB_BUFFER_REPLACEMENT_CODEPOINT_DEFAULT);
}

func testBufferContents(b *Buffer, kind int, t *testing.T) {
	if kind == bufferEmpty {
		assertEqualInt(t, len(b.Info), 0)
		return
	}

	glyphs := b.Info
	L := len(glyphs)
	assertEqualInt(t, L, 5)
	assertEqualInt(t, len(b.Pos), 5)

	for _, g := range glyphs {
		assertEqualInt(t, int(g.mask), 0)
		assertEqualInt(t, int(g.glyphProps), 0)
		assertEqualInt(t, int(g.ligProps), 0)
		assertEqualInt(t, int(g.syllable), 0)
		assertEqualInt(t, int(g.unicode), 0)
		assertEqualInt(t, int(g.complexAux), 0)
		assertEqualInt(t, int(g.complexCategory), 0)
	}

	for i, g := range glyphs {
		cluster := 1 + i
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
		assertEqualInt(t, g.Cluster, cluster)
	}

	/* reverse, test, and reverse back */

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[L-i]))
	}

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
	}

	/* reverse_clusters works same as reverse for now since each codepoint is
	* in its own cluster */

	// 	hb_buffer_reverse_clusters (b);
	//    for (i = 0; i < len; i++)
	// 	 g_assert_cmphex (glyphs[i].codepoint, ==, int(utf32[len-i]));

	//    hb_buffer_reverse_clusters (b);
	//    for (i = 0; i < len; i++)
	// 	 g_assert_cmphex (glyphs[i].codepoint, ==, int(utf32[1+i]));

	/* now form a cluster and test again */
	glyphs[2].Cluster = glyphs[1].Cluster

	/* reverse, test, and reverse back */

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[L-i]))
	}

	b.Reverse()
	for i, g := range glyphs {
		assertEqualInt(t, int(g.codepoint), int(utf32[1+i]))
	}

	/* reverse_clusters twice still should return the original string,
	* but when applied once, the 1-2 cluster should be retained. */

	//    hb_buffer_reverse_clusters (b);
	//    for i, g := range glyphs {
	// 	 unsigned int j = len-1-i;
	// 	 if (j == 1)
	// 	   j = 2;
	// 	 else if (j == 2)
	// 	   j = 1;
	// 	 g_assert_cmphex (glyphs[i].codepoint, ==, int(utf32[1+j]));
	//    }

	//    hb_buffer_reverse_clusters (b);
	//    for i, g := range glyphs
	// 	 g_assert_cmphex (glyphs[i].codepoint, ==, int(utf32[1+i]));

	/* test reset clears content */

	//    hb_buffer_reset (b);
	//    assertEqualInt (t, hb_buffer_get_length (b), ==, 0);
}

func testBufferPositions(b *Buffer, t *testing.T) {
	/* Without shaping, positions should all be zero */
	assertEqualInt(t, len(b.Info), len(b.Pos))
	for _, pos := range b.Pos {
		assertEqualInt(t, 0, int(pos.XAdvance))
		assertEqualInt(t, 0, int(pos.YAdvance))
		assertEqualInt(t, 0, int(pos.XOffset))
		assertEqualInt(t, 0, int(pos.YOffset))
		assertEqualInt(t, 0, int(pos.attachChain))
		assertEqualInt(t, 0, int(pos.attachType))
	}

	//    /* test reset clears content */
	//    hb_buffer_reset (b);
	//    assertEqualInt (t, hb_buffer_get_length (b), ==, 0);
}

func TestBuffer(t *testing.T) {
	for i := 0; i < bufferNumTypes; i++ {
		buffer := newTestBuffer(i)

		testBufferProperties(buffer, t)
		testBufferContents(buffer, i, t)
		testBufferPositions(buffer, t)
	}
}
