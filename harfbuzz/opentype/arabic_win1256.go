package opentype

import (
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-shape-complex-arabic-win1256.hh Copyright © 2014  Google, Inc. Behdad Esfahbod

const lookupFlagIgnoreMarks = 0x08

type manifest struct {
	tag    hb_tag_t
	lookup *truetype.LookupGSUB
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
	initLookup = truetype.LookupGSUB{
		Flag: lookupFlagIgnoreMarks,
		Data: truetype.SubstitutionSingle{Format2: []truetype.SingleSubstitution2{
			initmediSubLookup,
			initSubLookup,
		}},
	}
	mediLookup = truetype.LookupGSUB{
		Flag: lookupFlagIgnoreMarks,
		Data: truetype.SubstitutionSingle{Format2: []truetype.SingleSubstitution2{
			initmediSubLookup,
			mediSubLookup,
			medifinaLamAlefSubLookup,
		}},
	}
	finaLookup = truetype.LookupGSUB{
		Flag: lookupFlagIgnoreMarks,
		Data: truetype.SubstitutionSingle{Format2: []truetype.SingleSubstitution2{
			finaSubLookup,
			/* We don't need this one currently as the sequence inherits masks
			 * from the first item. Just in case we change that in the future
			 * to be smart about Arabic masks when ligating... */
			medifinaLamAlefSubLookup,
		}},
	}
	rligLookup = truetype.LookupGSUB{
		Flag: lookupFlagIgnoreMarks,
		Data: lamAlefLigaturesSubLookup,
	}
	rligMarksLookup = truetype.LookupGSUB{
		Data: shaddaLigaturesSubLookup,
	}
)

// init/medi/fina forms
var (
	initmediSubLookup = truetype.SingleSubstitution2{
		Coverage:    truetype.CoverageList{198, 200, 201, 202, 203, 204, 205, 206, 211, 212, 213, 214, 223, 225, 227, 228, 236, 237},
		Substitutes: []fonts.GlyphIndex{162, 4, 5, 5, 6, 7, 9, 11, 13, 14, 15, 26, 140, 141, 142, 143, 154, 154},
	}
	initSubLookup = truetype.SingleSubstitution2{
		Coverage:    truetype.CoverageList{218, 219, 221, 222, 229},
		Substitutes: []fonts.GlyphIndex{27, 30, 128, 131, 144},
	}
	mediSubLookup = truetype.SingleSubstitution2{
		Coverage:    truetype.CoverageList{218, 219, 221, 222, 229},
		Substitutes: []fonts.GlyphIndex{28, 31, 129, 138, 149},
	}
	finaSubLookup = truetype.SingleSubstitution2{
		Coverage:    truetype.CoverageList{194, 195, 197, 198, 199, 201, 204, 205, 206, 218, 219, 229, 236, 237},
		Substitutes: []fonts.GlyphIndex{2, 1, 3, 181, 0, 159, 8, 10, 12, 29, 127, 152, 160, 156},
	}
	medifinaLamAlefSubLookup = truetype.SingleSubstitution2{
		Coverage:    truetype.CoverageList{165, 178, 180, 252},
		Substitutes: []fonts.GlyphIndex{170, 179, 185, 255},
	}
)

type ligs = []truetype.LigatureGlyph

var (
	// Lam+Alef ligatures
	lamAlefLigaturesSubLookup = truetype.SubstitutionLigature{
		Coverage:  truetype.CoverageList{225},
		Ligatures: []ligs{shaddaLigatureSet},
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
	shaddaLigaturesSubLookup = truetype.SubstitutionLigature{
		Coverage:  truetype.CoverageList{248},
		Ligatures: []ligs{shaddaLigatureSet},
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
