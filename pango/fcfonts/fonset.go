package fcfonts

import (
	"container/list"

	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
)

var _ pango.Fontset = (*fcFontset)(nil)

type fcFontset struct {
	key *PangoFcFontsetKey

	patterns   *fcPatterns
	patterns_i int

	fonts []*Font
	// coverages []pango.Coverage

	cache_link *list.Element
}

func pango_fc_fontset_new(key PangoFcFontsetKey, patterns *fcPatterns) *fcFontset {
	var fontset fcFontset

	fontset.key = &key
	fontset.patterns = patterns

	return &fontset
}

func (fontset *fcFontset) GetLanguage() pango.Language { return fontset.key.language }

func (fontset *fcFontset) pango_fc_fontset_load_next_font() *Font {

	pattern := fontset.patterns.pattern
	fontPattern, prepare := fontset.patterns.pango_fc_patterns_get_font_pattern(fontset.patterns_i)
	fontset.patterns_i++
	if fontPattern == nil {
		return nil
	}

	if prepare {
		fontPattern = (*fc.Config)(nil).PrepareRender(pattern, fontPattern) // TODO:
	}

	font := fontset.key.fontmap.newFont(*fontset.key, fontPattern)

	return font
}

// lazy loading
func (fontset *fcFontset) getFontAt(i int) *Font {
	for i >= len(fontset.fonts) {
		font := fontset.pango_fc_fontset_load_next_font()
		fontset.fonts = append(fontset.fonts, font)
		// fontset.coverages = append(fontset.coverages, nil)
		if font == nil {
			return nil
		}
	}

	return fontset.fonts[i]
}

func (fontset *fcFontset) Foreach(fn pango.FontsetForeachFunc) {
	for i := 0; ; i++ {
		font := fontset.getFontAt(i)
		if fn(font) {
			return
		}
	}
}

// func (fontset *fcFontset) GetFont(wc rune) pango.Font {
// 	for i := 0; fontset.getFontAt(i) != nil; i++ {
// 		font := fontset.fonts[i]
// 		coverage := fontset.coverages[i]

// 		if coverage == nil {
// 			coverage = font.GetCoverage(fontset.key.language)
// 			fontset.coverages[i] = coverage
// 		}

// 		level := coverage.Get(wc)

// 		if level {
// 			return font
// 		}
// 	}

// 	return nil
// }
