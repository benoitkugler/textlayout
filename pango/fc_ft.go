package pango

import "github.com/benoitkugler/textlayout/fontconfig"

type PangoFT2Font struct {
	PangoFcFont

	//   FT_Face face;
	//   int load_flags;
	size int

	//   GSList *metrics_by_lang;

	//   GHashTable *glyph_info;
	//   GDestroyNotify glyph_cache_destroy;
}

func newPangoFT2Font(pattern fontconfig.FcPattern, fontmap *PangoFcFontMap) *PangoFT2Font {
	var ft2font PangoFT2Font
	if ds := pattern.GetFloats(fontconfig.FC_PIXEL_SIZE); len(ds) != 0 {
		ft2font.size = int(ds[0] * float64(pangoScale))
	}
	ft2font.font_pattern = pattern
	ft2font.fontmap = fontmap
	return &ft2font
}
