package fcfonts

import (
	"container/list"
	"fmt"
	"os"

	"github.com/benoitkugler/textlayout/fontconfig"
	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/benoitkugler/textlayout/pango"
)

var _ pango.FontMap = (*FontMap)(nil)

type fontMapPrivate struct {
	FontsetTable FontsetHash
	fontsetCache *list.List // *PangoFontset /* Recently used Fontsets */

	font_hash fontHash

	patternsHash PatternHash

	font_face_data_hash map[faceDataKey]*faceData // font file name/id -> font data

	// families []*PangoFcFamily // List of all families available, nil means uninitialized

	// dpi float64

	/* Decoders */
	// GSList *findfuncs

	config *fc.Config

	// Database stores all the potential fonts, coming from
	// a fontconfig scan (or a cache)
	// This value is initialised at the start and should not be mutated.
	Database fc.Fontset

	closed bool // = 1;
}

// FontMap implements pango.FontMap using 'fontconfig' and 'fonts'.
type FontMap struct {
	context_key_get        func(*pango.Context) int
	fontset_key_substitute func(*PangoFontsetKey, fc.Pattern)
	// default_substitute     func(fc.Pattern)

	fontMapPrivate

	// Function to call on prepared patterns to do final config tweaking.
	// substitute_func    PangoFcSubstituteFunc
	// substitute_data    gpointer
	// substitute_destroy GDestroyNotify

	// TODO: check the design of C "class"

	// fields of the PangoFT2FontMap of the C code

	//  library FT_Library

	dpi_x, dpi_y float32
	serial       uint
}

type faceDataKey = fonts.FaceID

// NewFontMap creates a new font map, used
// to cache information about available fonts, and holds
// certain global parameters such as the resolution and
// the default substitute function.
// The Config object will be used to query information from the database.
func NewFontMap(c *fontconfig.Config, database fontconfig.Fontset) *FontMap {
	var priv fontMapPrivate

	priv.font_hash = make(fontHash)
	priv.FontsetTable = make(FontsetHash)
	priv.patternsHash = make(PatternHash)
	priv.font_face_data_hash = make(map[faceDataKey]*faceData)
	priv.config = c
	priv.Database = database
	priv.fontsetCache = list.New()
	// priv.dpi = -1

	return &FontMap{fontMapPrivate: priv}
}

func (fontmap *FontMap) getFontFaceData(fontPattern fc.Pattern) (faceDataKey, *faceData) {
	key := fontPattern.FaceID()

	data := fontmap.font_face_data_hash[key]
	if data != nil {
		return key, data
	}

	// data = &faceData{pattern: fontPattern}
	data = &faceData{}
	data.format = fontPattern.Format()
	// other fields are loaded lazilly

	fontmap.font_face_data_hash[key] = data

	return key, data
}

// retrieves the `HB_face_t` for the given `font`
func (fontmap *FontMap) getHBFace(font *fcFont) (harfbuzz.Face, error) {
	key, data := fontmap.getFontFaceData(font.fontPattern)

	if data.hb_face == nil {
		f, err := os.Open(key.File)
		if err != nil {
			return nil, fmt.Errorf("font file not found: %s", err)
		}
		defer f.Close()

		loader := data.format.Loader()
		if loader == nil { // should not happen for pattern scanned from disk
			return nil, fmt.Errorf("unsupported file format %s", data.format)
		}

		fonts, err := loader.Load(f)
		if err != nil {
			return nil, fmt.Errorf("corrupted font file (with type %s): %s", data.format, key.File)
		}
		if int(key.Index) >= len(fonts) {
			return nil, fmt.Errorf("out of range font index: %d", key.Index)
		}

		data.hb_face = fonts[key.Index]
	}

	return data.hb_face, nil
}

func (fontmap *FontMap) GetSerial() uint { return fontmap.serial }

func (fontmap *FontMap) pango_font_map_get_patterns(key *PangoFontsetKey) *Patterns {
	pattern := key.makePattern()
	key.defaultSubstitute(fontmap, pattern)
	return fontmap.pango_patterns_new(pattern)
}

func (fontmap *FontMap) cacheFontset(fs *Fontset) {
	cache := fontmap.fontsetCache

	if fs.cache_link != nil {
		if fs.cache_link == cache.Front() {
			return
		}
		// Already in cache, move to head
		// if fs.cache_link == cache.Back() {
		// 	cache.tail = fs.cache_link.prev
		// }
		cache.Remove(fs.cache_link)
	} else {
		// Add to cache initially
		if cache.Len() == fontsetCacheSize {
			tmp_Fontset := cache.Remove(cache.Front()).(*Fontset)
			tmp_Fontset.cache_link = nil
			fontmap.FontsetTable.remove(*tmp_Fontset.key)
		}

		fs.cache_link = &list.Element{Value: fs}
	}

	cache.PushFront(fs.cache_link.Value)
}

func (fontmap *FontMap) LoadFontset(context *pango.Context, desc *pango.FontDescription, language pango.Language) pango.Fontset {
	key := fontmap.newFontsetKey(context, desc, language)

	fontset := fontmap.FontsetTable.lookup(key)
	if fontset == nil {
		patterns := fontmap.pango_font_map_get_patterns(&key)
		fontset = pango_Fontset_new(key, patterns)
		fontmap.FontsetTable.insert(*fontset.key, fontset)
	}

	fontmap.cacheFontset(fontset)
	return fontset
}

type faceData struct {
	hb_face harfbuzz.Face
	format  fc.FontFormat
	// pattern   fc.Pattern // pattern that owns filename
	coverage  pango.Coverage
	languages []pango.Language // TODO: check usage

}
