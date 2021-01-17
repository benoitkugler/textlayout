package fcfonts

import (
	"github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
)

type fcFontKeyHash struct {
	pattern     string
	matrix      pango.Matrix
	context_key int
	variations  string
}

type fontHash map[fcFontKeyHash]*Font // (GHashFunc)pango_fc_font_key_hash,  (GEqualFunc)pango_fc_font_key_equal

func (m fontHash) lookup(p PangoFcFontKey) *Font {
	key := fcFontKeyHash{pattern: p.pattern.Hash(), matrix: p.matrix,
		context_key: p.context_key, variations: p.variations}
	return m[key]
}

func (m fontHash) insert(p PangoFcFontKey, v *Font) {
	key := fcFontKeyHash{pattern: p.pattern.Hash(), matrix: p.matrix,
		context_key: p.context_key, variations: p.variations}
	m[key] = v
}

func (m fontHash) remove(p PangoFcFontKey) {
	key := fcFontKeyHash{pattern: p.pattern.Hash(), matrix: p.matrix,
		context_key: p.context_key, variations: p.variations}
	delete(m, key)
}

type fontsetHash map[PangoFcFontsetKey]*fcFontset

func (m fontsetHash) lookup(p PangoFcFontsetKey) *fcFontset {
	p.desc = p.desc.AsHash()
	p.fontmap = nil
	return m[p]
}

func (m fontsetHash) insert(p PangoFcFontsetKey, v *fcFontset) {
	p.desc = p.desc.AsHash()
	p.fontmap = nil
	m[p] = v
}

func (m fontsetHash) remove(p PangoFcFontsetKey) {
	p.desc = p.desc.AsHash()
	p.fontmap = nil
	delete(m, p)
}

type fcPatternHash map[string]*fcPatterns

func (m fcPatternHash) lookup(p fontconfig.Pattern) *fcPatterns { return m[p.Hash()] }

func (m fcPatternHash) insert(p fontconfig.Pattern, pts *fcPatterns) { m[p.Hash()] = pts }

// ------------------------------------------------------------------------------------
