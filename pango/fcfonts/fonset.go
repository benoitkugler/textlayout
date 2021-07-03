package fcfonts

import (
	"container/list"

	"github.com/benoitkugler/textlayout/pango"
)

var _ pango.Fontset = (*Fontset)(nil)

type Fontset struct {
	key        *PangoFontsetKey
	patterns   *Patterns
	cache_link *list.Element
	fonts      []*Font // lazily filled
	patterns_i int
}

func pango_Fontset_new(key PangoFontsetKey, patterns *Patterns) *Fontset {
	var fs Fontset

	fs.key = &key
	fs.patterns = patterns

	return &fs
}

func (fs *Fontset) GetLanguage() pango.Language { return fs.key.language }

func (fs *Fontset) pango_Fontset_load_next_font() *Font {
	pattern := fs.patterns.pattern
	fontPattern, prepare := fs.patterns.getFontPattern(fs.patterns_i)
	fs.patterns_i++
	if fontPattern == nil {
		return nil
	}

	if prepare {
		fontPattern = fs.patterns.fontmap.config.PrepareRender(pattern, fontPattern)
	}

	font := fs.key.fontmap.newFont(*fs.key, fontPattern)

	return font
}

// lazy loading
func (fs *Fontset) getFontAt(i int) *Font {
	for i >= len(fs.fonts) {
		font := fs.pango_Fontset_load_next_font()
		fs.fonts = append(fs.fonts, font)
		// Fontset.coverages = append(Fontset.coverages, nil)
		if font == nil {
			return nil
		}
	}

	return fs.fonts[i]
}

func (fs *Fontset) Foreach(fn pango.FontsetForeachFunc) {
	for i := 0; ; i++ {
		font := fs.getFontAt(i)
		if fn(font) {
			return
		}
	}
}

// func (Fontset *Fontset) GetFont(wc rune) pango.Font {
// 	for i := 0; Fontset.getFontAt(i) != nil; i++ {
// 		font := Fontset.fonts[i]
// 		coverage := Fontset.coverages[i]

// 		if coverage == nil {
// 			coverage = font.GetCoverage(Fontset.key.language)
// 			Fontset.coverages[i] = coverage
// 		}

// 		level := coverage.Get(wc)

// 		if level {
// 			return font
// 		}
// 	}

// 	return nil
// }
