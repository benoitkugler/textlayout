package fcfonts

import (
	"fmt"
	"os"
	"strings"

	"github.com/benoitkugler/textlayout/fontconfig"
	fc "github.com/benoitkugler/textlayout/fontconfig"
	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/benoitkugler/textlayout/pango"
)

// pangofc-fontmap.c: Base fontmap type for fontconfig-based backends

/*
 * PangoFcFontMap is a base class for font map implementations using the
 * Fontconfig and FreeType libraries. It is used in the
 * <link linkend="pango-Xft-Fonts-and-Rendering">Xft</link> and
 * <link linkend="pango-FreeType-Fonts-and-Rendering">FreeType</link>
 * backends shipped with Pango, but can also be used when creating
 * new backends. Any backend deriving from this base class will
 * take advantage of the wide range of shapers implemented using
 * FreeType that come with Pango.
 */

const fontsetCacheSize = 256

/* Overview:
 *
 * All programming is a practice in caching data. PangoFcFontMap is the
 * major caching container of a Pango system on a Linux desktop. Here is
 * a short overview of how it all works.
 *
 * In short, Fontconfig search patterns are constructed and a Fontset loaded
 * using them. Here is how we achieve that:
 *
 * - All Pattern's referenced by any object in the fontmap are uniquified
 *   and cached in the fontmap. This both speeds lookups based on patterns
 *   faster, and saves memory. This is handled by fontmap.priv.pattern_hash.
 *   The patterns are cached indefinitely.
 *
 * - The results of a Sort() are used to populate Fontsets.  However,
 *   Sort() relies on the search pattern only, which includes the font
 *   size but not the full font matrix.  The Fontset however depends on the
 *   matrix.  As a result, multiple Fontsets may need results of the
 *   Sort() on the same input pattern (think rotating text).  As such,
 *   we cache Sort() results in fontmap.priv.patterns_hash which
 *   is a refcounted structure.  This level of abstraction also allows for
 *   optimizations like calling FcFontMatch() instead of Sort(), and
 *   only calling Sort() if any patterns other than the first match
 *   are needed.  Another possible optimization would be to call Sort()
 *   without trimming, and do the trimming lazily as we go.  Only pattern sets
 *   already referenced by a Fontset are cached.
 *
 * - A number of most-recently-used Fontsets are cached and reused when
 *   needed.  This is achieved using fontmap.priv.Fontset_hash and
 *   fontmap.priv.Fontset_cache.
 *
 * - All fonts created by any of our Fontsets are also cached and reused.
 *   This is what fontmap.priv.font_hash does.
 *
 * - Data that only depends on the font file and face index is cached and
 *   reused by multiple fonts.  This includes coverage and cmap cache info.
 *   This is done using fontmap.priv.font_face_data_hash.
 *
 * Upon a cache_clear() request, all caches are emptied.  All objects (fonts,
 * Fontsets, faces, families) having a reference from outside will still live
 * and may reference the fontmap still, but will not be reused by the fontmap.
 *
 *
 */

const (
	// String representing a fontconfig property name that Pango sets on any
	// fontconfig pattern it passes to fontconfig if a `Gravity` other
	// than PANGO_GRAVITY_SOUTH is desired.
	//
	// The property will have a `Gravity` value as a string, like "east".
	// This can be used to write fontconfig configuration rules to choose
	// different fonts for horizontal and vertical writing directions.
	fcGravity fc.Object = fc.FirstCustomObject + iota

	// String representing a fontconfig property name that Pango reads from font
	// patterns to populate list of OpenType font variations to be used for a font.
	//
	// The property will have a string elements, each of which a comma-separated
	// list of OpenType axis setting of the form AXIS=VALUE.
	fcFontVariations
)

type faceData struct {
	hbFace harfbuzz.Face
	format fc.FontFormat
}

var _ pango.FontMap = (*FontMap)(nil)

// FontMap implements pango.FontMap using 'fontconfig' and 'fonts'.
type FontMap struct {
	fontsetTable fontsetCache

	fontHash fontHash

	patternsHash patternHash

	fontKeyHash map[faceDataKey]*faceData // font file name/id -> font data

	config *fc.Config

	// Database stores all the potential fonts, coming from
	// a fontconfig scan (or a cache).
	// This value is initialised at the start and should not be mutated.
	Database fc.Fontset

	dpiX, dpiY float32
	serial     uint
}

type faceDataKey = fonts.FaceID

// NewFontMap creates a new font map, used
// to cache information about available fonts, and holds
// certain global parameters such as the resolution and
// the default substitute function.
// The `config` object will be used to query information from the `database`.
func NewFontMap(config *fontconfig.Config, database fontconfig.Fontset) *FontMap {
	var fm FontMap

	fm.fontHash = make(fontHash)
	fm.fontsetTable = make(fontsetCache)
	fm.patternsHash = make(patternHash)
	fm.fontKeyHash = make(map[faceDataKey]*faceData)
	fm.config = config
	fm.Database = database
	// priv.dpi = -1
	fm.serial = 1
	fm.dpiX = 96
	fm.dpiY = 96
	return &fm
}

func (fontmap *FontMap) getFontFaceData(fontPattern fc.Pattern) (faceDataKey, *faceData) {
	key := fontPattern.FaceID()

	data := fontmap.fontKeyHash[key]
	if data != nil {
		return key, data
	}

	data = &faceData{}
	data.format = fontPattern.Format()
	// other fields are loaded lazilly

	fontmap.fontKeyHash[key] = data

	return key, data
}

// retrieves the `HB_face_t` for the given `font`
func (fontmap *FontMap) getHBFace(font *fcFont) (harfbuzz.Face, error) {
	key, data := fontmap.getFontFaceData(font.Pattern)

	if data.hbFace == nil {
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

		data.hbFace = fonts[key.Index]
	}

	return data.hbFace, nil
}

func (fontmap *FontMap) GetSerial() uint { return fontmap.serial }

func (fontmap *FontMap) getPatterns(key *fontsetKey) *fcPatterns {
	pattern := key.makePattern()
	key.defaultSubstitute(fontmap, pattern)
	return fontmap.newFcPatterns(pattern)
}

func (fontmap *FontMap) LoadFontset(context *pango.Context, desc *pango.FontDescription, language pango.Language) pango.Fontset {
	key := fontmap.newFontsetKey(context, desc, language)

	fontset := fontmap.fontsetTable.lookup(key)
	if fontset == nil {
		patterns := fontmap.getPatterns(&key)
		fontset = newFontset(key, patterns)
		fontmap.fontsetTable.insert(*fontset.key, fontset)
	}

	return fontset
}

func (fcfontmap *FontMap) getScaledSize(context *pango.Context, desc *pango.FontDescription) int {
	size := float32(desc.Size)
	if !desc.SizeIsAbsolute {
		dpi := fcfontmap.getResolution(context)

		size = size * dpi / 72.
	}
	_, scale := context.Matrix.GetFontScaleFactors()
	return int(.5 + scale*size)
}

type fcFontKey struct {
	pattern    fc.Pattern
	variations string
	matrix     pango.Matrix
	contextKey int
}

func (fsKey *fontsetKey) newFontKey(pattern fc.Pattern) fcFontKey {
	var key fcFontKey
	key.pattern = pattern
	key.matrix = fsKey.matrix
	key.variations = fsKey.variations
	return key
}

func (key *fcFontKey) pango_font_key_get_gravity() pango.Gravity {
	gravity := pango.GRAVITY_SOUTH

	pattern := key.pattern

	if s, ok := pattern.GetString(fcGravity); ok {
		value, _ := pango.GravityMap.FromString(s)
		gravity = pango.Gravity(value)
	}

	return gravity
}

func (key *fcFontKey) getFontSize() float32 {
	if size, ok := key.pattern.GetFloat(fc.PIXEL_SIZE); ok {
		return size
	}

	/* Just in case PIXEL_SIZE got unset between pango_make_pattern()
	* and here. That would be very weird. */
	dpi, ok := key.pattern.GetFloat(fc.DPI)
	if !ok {
		dpi = 72
	}

	if size, ok := key.pattern.GetFloat(fc.SIZE); ok {
		return size * dpi / 72.
	}

	// Whatever
	return 18.
}

type fontsetKey struct {
	fontmap    *FontMap
	language   pango.Language
	variations string
	desc       pango.FontDescription
	matrix     pango.Matrix
	pixelsize  int
	resolution float32
}

func (fcfontmap *FontMap) newFontsetKey(context *pango.Context, desc *pango.FontDescription, language pango.Language) fontsetKey {
	if language == "" && context != nil {
		language = context.GetLanguage()
	}

	var key fontsetKey
	key.fontmap = fcfontmap

	if context != nil && context.Matrix != nil {
		key.matrix = *context.Matrix
	} else {
		key.matrix = pango.Identity
	}
	key.matrix.X0, key.matrix.Y0 = 0, 0

	key.pixelsize = fcfontmap.getScaledSize(context, desc)
	key.resolution = fcfontmap.getResolution(context)
	key.language = language
	key.variations = desc.Variations
	key.desc = *desc
	key.desc.UnsetFields(pango.FmSize | pango.FmVariations)

	return key
}

// makePattern translates the pango font description into
// a fontconfig query pattern (without performing any substitutions)
func (key *fontsetKey) makePattern() fc.Pattern {
	slant := slantToFC(key.desc.Style)
	weight := fc.WeightFromOT(float32(key.desc.Weight))
	width := widthToFC(key.desc.Stretch)

	gravity := key.desc.Gravity
	vertical := fc.False
	if gravity.IsVertical() {
		vertical = fc.True
	}

	// The reason for passing in SIZE as well as PIXEL_SIZE is
	// to work around a bug in libgnomeprint where it doesn't look
	// for PIXEL_SIZE. See http://bugzilla.gnome.org/show_bug.cgi?id=169020
	//
	// Putting SIZE in here slightly reduces the efficiency
	// of caching of patterns and fonts when working with multiple different
	// dpi values.
	//
	// Do not pass FC_VERTICAL_LAYOUT true as HarfBuzz shaping assumes false.
	pattern := fc.BuildPattern([]fc.PatternElement{
		// {Object: PANGO_VERSION, Value: pango_version()},       // FcTypeInteger
		{Object: fc.WEIGHT, Value: fc.Float(weight)},                                                // FcTypeDouble
		{Object: fc.SLANT, Value: fc.Int(slant)},                                                    // FcTypeInteger
		{Object: fc.WIDTH, Value: fc.Int(width)},                                                    // FcTypeInteger
		{Object: fc.VERTICAL_LAYOUT, Value: vertical},                                               // FcTypeBool
		{Object: fc.VARIABLE, Value: fc.DontCare},                                                   //  FcTypeBool
		{Object: fc.DPI, Value: fc.Float(key.resolution)},                                           // FcTypeDouble
		{Object: fc.SIZE, Value: fc.Float(float32(key.pixelsize) * (72. / 1024. / key.resolution))}, // FcTypeDouble
		{Object: fc.PIXEL_SIZE, Value: fc.Float(key.pixelsize) / 1024.},                             // FcTypeDouble
	}...)

	if key.variations != "" {
		pattern.Add(fc.FONT_VARIATIONS, fc.String(key.variations), true)
	}

	if key.desc.FamilyName != "" {
		families := strings.Split(key.desc.FamilyName, ",")
		for _, fam := range families {
			pattern.Add(fc.FAMILY, fc.String(fam), true)
		}
	}

	if key.language != "" {
		pattern.Add(fc.LANG, fc.String(key.language), true)
	}

	if gravity != pango.GRAVITY_SOUTH {
		pattern.Add(fcGravity, fc.String(pango.GravityMap.ToString("gravity", int(gravity))), true)
	}

	return pattern
}

// ------------------------------- PangoPatterns -------------------------------

type fcPatterns struct {
	fontmap *FontMap

	pattern fc.Pattern
	match   fc.Pattern
	fontset fc.Fontset // the result of fontconfig query
}

func (fontmap *FontMap) newFcPatterns(pat fc.Pattern) *fcPatterns {
	if pats := fontmap.patternsHash.lookup(pat); pats != nil {
		return pats
	}

	var pats fcPatterns

	pats.fontmap = fontmap
	pats.pattern = pat
	fontmap.patternsHash.insert(pat, &pats)

	return &pats
}

func (pats *fcPatterns) getFontPattern(i int) (fc.Pattern, bool) {
	if i == 0 {
		if pats.match == nil && pats.fontset == nil {
			pats.match = pats.fontmap.Database.Match(pats.pattern, pats.fontmap.config)
		}

		if pats.match != nil {
			return pats.match, false
		}
	}

	if pats.fontset == nil {
		fonts := pats.fontmap.Database
		// we actually supports more formats than Harfbuzz, no need to filter

		pats.fontset, _ = fonts.Sort(pats.pattern, true)

		if pats.match != nil {
			pats.match = nil
		}
	}

	if i < len(pats.fontset) {
		return pats.fontset[i], true
	}
	return nil, true
}

func slantToFC(pangoStyle pango.Style) int {
	switch pangoStyle {
	case pango.STYLE_NORMAL:
		return fc.SLANT_ROMAN
	case pango.STYLE_ITALIC:
		return fc.SLANT_ITALIC
	case pango.STYLE_OBLIQUE:
		return fc.SLANT_OBLIQUE
	default:
		return fc.SLANT_ROMAN
	}
}

func widthToFC(pangoStretch pango.Stretch) int {
	switch pangoStretch {
	case pango.STRETCH_NORMAL:
		return fc.WIDTH_NORMAL
	case pango.STRETCH_ULTRA_CONDENSED:
		return fc.WIDTH_ULTRACONDENSED
	case pango.STRETCH_EXTRA_CONDENSED:
		return fc.WIDTH_EXTRACONDENSED
	case pango.STRETCH_CONDENSED:
		return fc.WIDTH_CONDENSED
	case pango.STRETCH_SEMI_CONDENSED:
		return fc.WIDTH_SEMICONDENSED
	case pango.STRETCH_SEMI_EXPANDED:
		return fc.WIDTH_SEMIEXPANDED
	case pango.STRETCH_EXPANDED:
		return fc.WIDTH_EXPANDED
	case pango.STRETCH_EXTRA_EXPANDED:
		return fc.WIDTH_EXTRAEXPANDED
	case pango.STRETCH_ULTRA_EXPANDED:
		return fc.WIDTH_ULTRAEXPANDED
	default:
		return fc.WIDTH_NORMAL
	}
}

// also load the underlying harbuzz font
func (fontmap *FontMap) newFont(fsKey fontsetKey, match fc.Pattern) (*Font, error) {
	key := fsKey.newFontKey(match)

	if fcfont := fontmap.fontHash.lookup(key); fcfont != nil {
		return fcfont, nil
	}

	pangoMatrix := fsKey.matrix
	// Fontconfig has the Y axis pointing up, Pango, down.
	fcMatrix := fc.Matrix{Xx: pangoMatrix.Xx, Xy: -pangoMatrix.Xy, Yx: -pangoMatrix.Yx, Yy: pangoMatrix.Yy}

	pattern := match.Duplicate()

	for _, fcMatrixVal := range pattern.GetMatrices(fc.MATRIX) {
		fcMatrix = fcMatrix.Multiply(fcMatrixVal)
	}

	pattern.Del(fc.MATRIX)
	pattern.Add(fc.MATRIX, fcMatrix, true)

	fcfont := newFont(pattern, fontmap)

	fcfont.matrix = key.matrix

	// cache it on fontmap
	fontmap.fontHash.insert(key, fcfont)

	err := fcfont.loadHBFont()

	return fcfont, err
}

func (key *fontsetKey) defaultSubstitute(fontmap *FontMap, pattern fc.Pattern) {
	// inlined version of pango_cairo_fc_font_map_fontset_key_substitute
	fontmap.config.Substitute(pattern, nil, fc.MatchQuery)

	// if fontmap.substitute_func {
	// 	fontmap.substitute_func(pattern, fontmap.substitute_data)
	// }
	// if key != nil  {
	// 	cairo_ft_font_options_substitute(pango_fc_fontset_key_get_context_key(fontkey),
	// 		pattern)
	// }

	pattern.SubstituteDefault()
}

func (fontmap *FontMap) getResolution(*pango.Context) float32 { return fontmap.dpiY }
