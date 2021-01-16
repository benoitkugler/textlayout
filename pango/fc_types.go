package pango

import (
	"github.com/benoitkugler/textlayout/fontconfig"
)

// we emulate the behavior of the C implementation with custom maps

// type fontHash map[uint32]*PangoFcFont // (GHashFunc)pango_fc_font_key_hash,  (GEqualFunc)pango_fc_font_key_equal

type fontsetHash map[PangoFcFontsetKey]*PangoFcFontset // (GHashFunc)pango_fc_fontset_key_hash, (GEqualFunc)pango_fc_fontset_key_equal

func (m fontsetHash) lookup(p PangoFcFontsetKey) *PangoFcFontset {
	p.desc = p.desc.pango_font_description_hash()
	return m[p]
}

func (m fontsetHash) insert(p PangoFcFontsetKey, v *PangoFcFontset) {
	p.desc = p.desc.pango_font_description_hash()
	m[p] = v
}

func (m fontsetHash) remove(p PangoFcFontsetKey) {
	p.desc = p.desc.pango_font_description_hash()
	delete(m, p)
}

type fcPatternHash map[string]*PangoFcPatterns // (GHashFunc) FcPatternHash,(GEqualFunc) FcPatternEqual

func (m fcPatternHash) lookup(p fontconfig.FcPattern) *PangoFcPatterns { return m[p.Hash()] }

func (m fcPatternHash) insert(p fontconfig.FcPattern, pts *PangoFcPatterns) { m[p.Hash()] = pts }

// ------------------------------------------------------------------------------------
