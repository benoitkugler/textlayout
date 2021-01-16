package fcfonts

import (
	"container/list"

	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/pango"
)

// PangoFcFontMap is a base class for font map implementations
// using the Fontconfig and FreeType libraries. To create a new
// backend using Fontconfig and FreeType, you derive from this class
// and implement a new_font() virtual function that creates an
// instance deriving from PangoFcFont.
type PangoFcFontMap struct {
	// parent_instance FontMap

	fontMapPrivate

	// Function to call on prepared patterns to do final config tweaking.
	// substitute_func    PangoFcSubstituteFunc
	// substitute_data    gpointer
	// substitute_destroy GDestroyNotify

	// TODO: check the design of C "class"
	context_key_get        func(*pango.Context) int
	fontset_key_substitute func(*PangoFcFontsetKey, *fc.Pattern)
	default_substitute     func(*fc.Pattern)
}

func (fcfontmap *PangoFcFontMap) getFontFaceData(font_pattern fc.Pattern) *faceData {
	var (
		key faceDataKey
		ok  bool
	)

	key.filename, ok = font_pattern.GetString(fc.FC_FILE)
	if !ok {
		return nil
	}

	key.id, ok = font_pattern.GetInt(fc.FC_INDEX)
	if !ok {
		return nil
	}

	data := fcfontmap.font_face_data_hash[key]
	if data != nil {
		return data
	}

	data = &faceData{pattern: font_pattern}
	fcfontmap.font_face_data_hash[key] = data

	return data
}

type fontMapPrivate struct {
	fontset_hash  fontsetHash
	fontset_cache *list.List // *PangoFcFontset /* Recently used fontsets */

	font_hash map[PangoFcFontKey]*PangoFcFont

	patterns_hash fcPatternHash

	font_face_data_hash map[faceDataKey]*faceData // font file name/id -> font data

	// List of all families availible
	families [][]PangoFcFamily
	//    int n_families;		/* -1 == uninitialized */

	dpi float64

	/* Decoders */
	// GSList *findfuncs

	Closed bool // = 1;

	config *fc.FcConfig
}

type faceDataKey struct {
	filename string
	id       int // needed to handle TTC files with multiple faces
}

type faceData struct {
	pattern   fc.Pattern /* Referenced pattern that owns filename */
	coverage  pango.Coverage
	languages []pango.Language

	// hb_face_t *hb_face
}
