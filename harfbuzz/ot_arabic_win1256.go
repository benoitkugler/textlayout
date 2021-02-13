package harfbuzz

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-shape-complex-arabic-win1256.hh Copyright © 2014  Google, Inc. Behdad Esfahbod

type manifest struct {
	tag    hb_tag_t
	lookup *lookupGSUB
}

var arabicWin1256GsubLookups = [...]manifest{
	{newTag('r', 'l', 'i', 'g'), &rligLookup},
	{newTag('i', 'n', 'i', 't'), &initLookup},
	{newTag('m', 'e', 'd', 'i'), &mediLookup},
	{newTag('f', 'i', 'n', 'a'), &finaLookup},
	{newTag('r', 'l', 'i', 'g'), &rligMarksLookup},
}

// Lookups
var (
	initLookup = lookupGSUB{
		Flag: truetype.IgnoreMarks,
		Subtables: []truetype.GSUBSubtable{
			initmediSubLookup,
			initSubLookup,
		},
	}
	mediLookup = lookupGSUB{
		Flag: truetype.IgnoreMarks,
		Subtables: []truetype.GSUBSubtable{
			initmediSubLookup,
			mediSubLookup,
			medifinaLamAlefSubLookup,
		},
	}
	finaLookup = lookupGSUB{
		Flag: truetype.IgnoreMarks,
		Subtables: []truetype.GSUBSubtable{
			finaSubLookup,
			/* We don't need this one currently as the sequence inherits masks
			 * from the first item. Just in case we change that in the future
			 * to be smart about Arabic masks when ligating... */
			medifinaLamAlefSubLookup,
		},
	}
	rligLookup = lookupGSUB{
		Flag:      truetype.IgnoreMarks,
		Subtables: []truetype.GSUBSubtable{lamAlefLigaturesSubLookup},
	}
	rligMarksLookup = lookupGSUB{
		Subtables: []truetype.GSUBSubtable{shaddaLigaturesSubLookup},
	}
)

// init/medi/fina forms
var (
	initmediSubLookup = truetype.GSUBSubtable{
		Coverage: truetype.CoverageList{198, 200, 201, 202, 203, 204, 205, 206, 211, 212, 213, 214, 223, 225, 227, 228, 236, 237},
		Data:     truetype.GSUBSingle2{162, 4, 5, 5, 6, 7, 9, 11, 13, 14, 15, 26, 140, 141, 142, 143, 154, 154},
	}
	initSubLookup = truetype.GSUBSubtable{
		Coverage: truetype.CoverageList{218, 219, 221, 222, 229},
		Data:     truetype.GSUBSingle2{27, 30, 128, 131, 144},
	}
	mediSubLookup = truetype.GSUBSubtable{
		Coverage: truetype.CoverageList{218, 219, 221, 222, 229},
		Data:     truetype.GSUBSingle2{28, 31, 129, 138, 149},
	}
	finaSubLookup = truetype.GSUBSubtable{
		Coverage: truetype.CoverageList{194, 195, 197, 198, 199, 201, 204, 205, 206, 218, 219, 229, 236, 237},
		Data:     truetype.GSUBSingle2{2, 1, 3, 181, 0, 159, 8, 10, 12, 29, 127, 152, 160, 156},
	}
	medifinaLamAlefSubLookup = truetype.GSUBSubtable{
		Coverage: truetype.CoverageList{165, 178, 180, 252},
		Data:     truetype.GSUBSingle2{170, 179, 185, 255},
	}
)

type ligs = []truetype.LigatureGlyph

var (
	// Lam+Alef ligatures
	lamAlefLigaturesSubLookup = truetype.GSUBSubtable{
		Coverage: truetype.CoverageList{225},
		Data:     truetype.GSUBLigature1{shaddaLigatureSet},
	}
	lamLigatureSet = ligs{
		truetype.LigatureGlyph{
			Glyph:      199,
			Components: []fonts.GlyphIndex{165},
		},
		truetype.LigatureGlyph{
			Glyph:      195,
			Components: []fonts.GlyphIndex{178},
		},
		truetype.LigatureGlyph{
			Glyph:      194,
			Components: []fonts.GlyphIndex{180},
		},
		truetype.LigatureGlyph{
			Glyph:      197,
			Components: []fonts.GlyphIndex{252},
		},
	}

	// Shadda ligatures
	shaddaLigaturesSubLookup = truetype.GSUBSubtable{
		Coverage: truetype.CoverageList{248},
		Data:     truetype.GSUBLigature1{shaddaLigatureSet},
	}
	shaddaLigatureSet = ligs{
		truetype.LigatureGlyph{
			Glyph:      243,
			Components: []fonts.GlyphIndex{172},
		},
		truetype.LigatureGlyph{
			Glyph:      245,
			Components: []fonts.GlyphIndex{173},
		},
		truetype.LigatureGlyph{
			Glyph:      246,
			Components: []fonts.GlyphIndex{175},
		},
	}
)
