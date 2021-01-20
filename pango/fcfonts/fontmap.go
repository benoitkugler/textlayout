package fcfonts

import (
	"container/list"

	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
)

var (
	_ pango.FontMap = (*FontMap)(nil)
)

type fontMapPrivate struct {
	fontsetTable  fontsetHash
	fontset_cache *list.List // *PangoFcFontset /* Recently used fontsets */

	font_hash fontHash

	patterns_hash fcPatternHash

	font_face_data_hash map[faceDataKey]*faceData // font file name/id -> font data

	// families []*PangoFcFamily // List of all families available, nil means uninitialized

	// dpi float64

	/* Decoders */
	// GSList *findfuncs

	Closed bool // = 1;

	config *fc.Config
}

// FontMap implements pango.FontMap using 'fontconfig' and 'fonts'.
type FontMap struct {
	fontMapPrivate

	// Function to call on prepared patterns to do final config tweaking.
	// substitute_func    PangoFcSubstituteFunc
	// substitute_data    gpointer
	// substitute_destroy GDestroyNotify

	// TODO: check the design of C "class"
	context_key_get        func(*pango.Context) int
	fontset_key_substitute func(*PangoFcFontsetKey, fc.Pattern)
	default_substitute     func(fc.Pattern)

	// fields of the PangoFT2FontMap of the C code

	//  library FT_Library

	serial       uint
	dpi_x, dpi_y float64
}

// NewFontMap creates a new font map, used
// to cache information about available fonts, and holds
// certain global parameters such as the resolution and
// the default substitute function.
func NewFontMap() *FontMap {
	var priv fontMapPrivate

	priv.font_hash = make(fontHash)
	priv.fontsetTable = make(fontsetHash)
	priv.patterns_hash = make(fcPatternHash)
	priv.font_face_data_hash = make(map[faceDataKey]*faceData)
	// priv.dpi = -1

	return &FontMap{fontMapPrivate: priv}
}

func (fontmap *FontMap) getFontFaceData(fontPattern fc.Pattern) (faceDataKey, *faceData) {
	var (
		key faceDataKey
		ok  bool
	)

	key.filename, ok = fontPattern.GetString(fc.FC_FILE)
	if !ok {
		return key, nil
	}

	key.id, ok = fontPattern.GetInt(fc.FC_INDEX)
	if !ok {
		return key, nil
	}

	data := fontmap.font_face_data_hash[key]
	if data != nil {
		return key, data
	}

	// data = &faceData{pattern: fontPattern}
	data = &faceData{}
	fontmap.font_face_data_hash[key] = data

	return key, data
}

// retrieves the `HB_face_t` for the given `font`
func (fontmap *FontMap) getHBFace(font *fcFont) *pango.HB_face_t {

	key, data := fontmap.getFontFaceData(font.fontPattern)

	if data.hb_face == nil {
		blob := pango.HB_blob_create_from_file(key.filename)
		data.hb_face = pango.HB_face_create(blob, key.id)
	}

	return data.hb_face
}

func (fontmap *FontMap) GetSerial() uint { return fontmap.serial }

func (fontmap *FontMap) pango_fc_font_map_get_patterns(key *PangoFcFontsetKey) *fcPatterns {
	pattern := key.pango_fc_fontset_key_make_pattern()
	key.pango_fc_default_substitute(fontmap, pattern)

	return fontmap.pango_fc_patterns_new(pattern)
}

func (fontmap *FontMap) cacheFontset(fontset *fcFontset) {
	cache := fontmap.fontset_cache

	if fontset.cache_link != nil {
		if fontset.cache_link == cache.Front() {
			return
		}
		// Already in cache, move to head
		// if fontset.cache_link == cache.Back() {
		// 	cache.tail = fontset.cache_link.prev
		// }
		cache.Remove(fontset.cache_link)
	} else {
		// Add to cache initially
		if cache.Len() == FONTSET_CACHE_SIZE {
			tmp_fontset := cache.Remove(cache.Front()).(*fcFontset)
			tmp_fontset.cache_link = nil
			fontmap.fontsetTable.remove(*tmp_fontset.key)
		}

		fontset.cache_link = &list.Element{Value: fontset}
	}

	cache.PushFront(fontset.cache_link.Value)
}

func (fontmap *FontMap) LoadFontset(context *pango.Context, desc *pango.FontDescription, language pango.Language) pango.Fontset {

	key := fontmap.newFontsetKey(context, desc, language)

	fontset := fontmap.fontsetTable.lookup(key)
	if fontset == nil {
		patterns := fontmap.pango_fc_font_map_get_patterns(&key)

		if patterns == nil {
			return nil
		}

		fontset = pango_fc_fontset_new(key, patterns)
		fontmap.fontsetTable.insert(*fontset.key, fontset)
	}

	fontmap.cacheFontset(fontset)

	return fontset
}

type faceDataKey struct {
	filename string
	id       int // needed to handle TTC files with multiple faces
}

type faceData struct {
	// pattern   fc.Pattern // pattern that owns filename
	coverage  pango.Coverage
	languages []pango.Language

	hb_face *pango.HB_face_t
}
