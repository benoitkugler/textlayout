package fcfonts

import (
	"log"

	"github.com/benoitkugler/textlayout/pango"
)

var _ pango.Fontset = (*Fontset)(nil)

// Fontset implements the pango.Fontset interface.
type Fontset struct {
	key                 *fontsetKey
	patterns            *fcPatterns
	fonts               []*Font // lazily filled
	currentPatternIndex int
}

func newFontset(key fontsetKey, patterns *fcPatterns) *Fontset {
	var fs Fontset

	fs.key = &key
	fs.patterns = patterns

	return &fs
}

func (fs *Fontset) GetLanguage() pango.Language { return fs.key.language }

// may return nil
func (fs *Fontset) loadNextFont() *Font {
	pattern := fs.patterns.pattern
	fontPattern, prepare := fs.patterns.getFontPattern(fs.currentPatternIndex)
	fs.currentPatternIndex++
	if fontPattern == nil {
		return nil
	}

	if prepare {
		fontPattern = fs.patterns.fontmap.Config.PrepareRender(pattern, fontPattern)
	}

	font, err := fs.key.fontmap.newFont(*fs.key, fontPattern)
	if err != nil {
		log.Println(err)
	}

	return font
}

// lazy loading
func (fs *Fontset) getFontAt(i int) *Font {
	for i >= len(fs.fonts) {
		font := fs.loadNextFont()
		fs.fonts = append(fs.fonts, font)
		if font == nil {
			return nil
		}
	}

	return fs.fonts[i]
}

func (fs *Fontset) Foreach(fn pango.FontsetForeachFunc) {
	for i := 0; ; i++ {
		font := fs.getFontAt(i)
		if font == nil {
			continue
		}
		if fn(font) {
			return
		}
	}
}
