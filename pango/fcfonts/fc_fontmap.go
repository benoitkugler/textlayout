package fcfonts

import (
	"strings"

	fc "github.com/benoitkugler/textlayout/fontconfig"
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
 * Todo:
 *
 * - Make PangoCoverage a GObject and subclass it as PangoFcCoverage which
 *   will directly use FcCharset. (#569622)
 *
 * - Lazy trimming of Sort() results.  Requires fontconfig with
 *   FcCharSetMerge().
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

func (fcfontmap *FontMap) getScaledSize(context *pango.Context, desc *pango.FontDescription) int {
	size := float32(desc.Size)

	if !desc.SizeIsAbsolute {
		dpi := fcfontmap.getResolution(context)

		size = size * dpi / 72.
	}

	_, scale := context.Matrix.GetFontScaleFactors()
	return int(.5 + scale*size)
}

type PangoFcFontKey struct {
	// fontmap     *PangoFcFontMap // TODO: check if this is correct
	pattern     fc.Pattern
	variations  string
	matrix      pango.Matrix
	context_key int
}

func (FontsetKey *PangoFontsetKey) newFontKey(pattern fc.Pattern) PangoFcFontKey {
	var key PangoFcFontKey
	key.pattern = pattern
	key.matrix = FontsetKey.matrix
	key.variations = FontsetKey.variations
	key.context_key = FontsetKey.context_key
	return key
}

func (key *PangoFcFontKey) pango_font_key_get_gravity() pango.Gravity {
	gravity := pango.PANGO_GRAVITY_SOUTH

	pattern := key.pattern

	if s, ok := pattern.GetString(fcGravity); ok {
		value, _ := pango.GravityMap.FromString(s)
		gravity = pango.Gravity(value)
	}

	return gravity
}

func (key *PangoFcFontKey) get_font_size() float32 {
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

type PangoFontsetKey struct {
	fontmap     *FontMap
	language    pango.Language
	variations  string
	desc        pango.FontDescription
	matrix      pango.Matrix
	pixelsize   int
	resolution  float32
	context_key int
}

func (fcfontmap *FontMap) newFontsetKey(context *pango.Context, desc *pango.FontDescription, language pango.Language) PangoFontsetKey {
	if language == "" && context != nil {
		language = context.GetLanguage()
	}

	var key PangoFontsetKey
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
	key.desc.UnsetFields(pango.F_SIZE | pango.F_VARIATIONS)

	if context != nil && fcfontmap.context_key_get != nil {
		key.context_key = fcfontmap.context_key_get(context)
	}
	return key
}

// makePattern translates the pango font description into
// a fontconfig query pattern (without performing any substitutions)
func (key *PangoFontsetKey) makePattern() fc.Pattern {
	slant := pango_convert_slant_to_fc(key.desc.Style)
	weight := fc.WeightFromOT(float32(key.desc.Weight))
	width := pango_convert_width_to_fc(key.desc.Stretch)

	gravity := key.desc.Gravity
	vertical := fc.False
	if gravity.IsVertical() {
		vertical = fc.True
	}

	/* The reason for passing in SIZE as well as PIXEL_SIZE is
	* to work around a bug in libgnomeprint where it doesn't look
	* for PIXEL_SIZE. See http://bugzilla.gnome.org/show_bug.cgi?id=169020
	*
	* Putting SIZE in here slightly reduces the efficiency
	* of caching of patterns and fonts when working with multiple different
	* dpi values. */
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

	if gravity != pango.PANGO_GRAVITY_SOUTH {
		pattern.Add(fcGravity, fc.String(pango.GravityMap.ToString("gravity", int(gravity))), true)
	}

	return pattern
}

// ------------------------------- PangoPatterns -------------------------------

type Patterns struct {
	fontmap *FontMap

	pattern fc.Pattern
	match   fc.Pattern
	fontset fc.Fontset // the result of fontconfig query
}

func (fontmap *FontMap) pango_patterns_new(pat fc.Pattern) *Patterns {
	if pats := fontmap.patternsHash.lookup(pat); pats != nil {
		return pats
	}

	var pats Patterns

	pats.fontmap = fontmap
	pats.pattern = pat
	fontmap.patternsHash.insert(pat, &pats)

	return &pats
}

func filterByFormat(fs fc.Fontset) fc.Fontset {
	// we actually supports more formats than Harfbuzz

	// var result fc.Fontset

	// for _, fontPattern := range fs {
	// 	if pango_is_supported_font_format(fontPattern) {
	// 		result = append(result, fontPattern)
	// 	}
	// }

	return fs
}

func (pats *Patterns) getFontPattern(i int) (fc.Pattern, bool) {
	if i == 0 {
		if pats.match == nil && pats.fontset == nil {
			pats.match = pats.fontmap.database.Match(pats.pattern, pats.fontmap.config)
		}

		if pats.match != nil {
			return pats.match, false
		}
	}

	if pats.fontset == nil {
		var filtered fc.Fontset

		fonts := pats.fontmap.database
		filtered = filterByFormat(fonts)

		pats.fontset, _ = filtered.Sort(pats.pattern, true)

		if pats.match != nil {
			pats.match = nil
		}
	}

	if i < len(pats.fontset) {
		return pats.fontset[i], true
	}
	return nil, true
}

//  static guint
//  pango_font_map_get_n_items (GListModel *list)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FONT_MAP (list);

//    ensureFamilies (fcfontmap);

//    return fcfontmap.priv.n_families;
//  }

//  static gpointer
//  pango_font_map_get_item (GListModel *list,
// 							 guint       position)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FONT_MAP (list);

//    ensureFamilies (fcfontmap);

//    if (position < fcfontmap.priv.n_families)
// 	 return g_object_ref (fcfontmap.priv.families[position]);

//    return nil;
//  }

//  static void
//  pango_font_map_list_model_init (GListModelInterface *iface)
//  {
//    iface.get_item_type = pango_font_map_get_item_type;
//    iface.get_n_items = pango_font_map_get_n_items;
//    iface.get_item = pango_font_map_get_item;
//  }

//  /**
//   * pango_font_map_add_decoder_find_func:
//   * @fcfontmap: The #PangoFcFontMap to add this method to.
//   * @findfunc: The #PangoFcDecoderFindFunc callback function
//   * @user_data: User data.
//   * @dnotify: A #GDestroyNotify callback that will be called when the
//   *  fontmap is finalized and the decoder is released.
//   *
//   * This function saves a callback method in the #PangoFcFontMap that
//   * will be called whenever new fonts are created.  If the
//   * function returns a #PangoFcDecoder, that decoder will be used to
//   * determine both coverage via a #FcCharSet and a one-to-one mapping of
//   * characters to glyphs.  This will allow applications to have
//   * application-specific encodings for various fonts.
//   *
//   * Since: 1.6
//   **/
//  void
//  pango_font_map_add_decoder_find_func (PangoFcFontMap        *fcfontmap,
// 					  PangoFcDecoderFindFunc findfunc,
// 					  gpointer               user_data,
// 					  GDestroyNotify         dnotify)
//  {
//    fontMapPrivate *priv;
//    PangoFcFindFuncInfo *info;

//    g_return_if_fail (PANGO_IS_FONT_MAP (fcfontmap));

//    priv = fcfontmap.priv;

//    info = g_slice_new (PangoFcFindFuncInfo);

//    info.findfunc = findfunc;
//    info.user_data = user_data;
//    info.dnotify = dnotify;

//    priv.findfuncs = g_slist_append (priv.findfuncs, info);
//  }

//  /**
//   * pango_font_map_find_decoder:
//   * @fcfontmap: The #PangoFcFontMap to use.
//   * @pattern: The #Pattern to find the decoder for.
//   *
//   * Finds the decoder to use for @pattern.  Decoders can be added to
//   * a font map using pango_font_map_add_decoder_find_func().
//   *
//   * Returns: (transfer full) (nullable): a newly created #PangoFcDecoder
//   *   object or %nil if no decoder is set for @pattern.
//   *
//   * Since: 1.26
//   **/
//  PangoFcDecoder *
//  pango_font_map_find_decoder  ( fcfontmap *PangoFcFontMap,
// 				  pattern *Pattern      )
//  {
//    GSList *l;

//    g_return_val_if_fail (PANGO_IS_FONT_MAP (fcfontmap), nil);
//    g_return_val_if_fail (pattern != nil, nil);

//    for (l = fcfontmap.priv.findfuncs; l && l.data; l = l.next)
// 	 {
// 	   PangoFcFindFuncInfo *info = l.data;
// 	   PangoFcDecoder *decoder;

// 	   decoder = info.findfunc (pattern, info.user_data);
// 	   if (decoder)
// 	 return decoder;
// 	 }

//    return nil;
//  }

//  static void
//  pango_font_map_finalize (GObject *object)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FONT_MAP (object);

//    pango_font_map_shutdown (fcfontmap);

//    if (fcfontmap.substitute_destroy)
// 	 fcfontmap.substitute_destroy (fcfontmap.substitute_data);

//    G_OBJECT_CLASS (pango_font_map_parent_class).finalize (object);
//  }

//  /* Remove mapping from fcfont.key to fcfont */
//  /* Closely related to shutdown_font() */
//  void
//  _pango_font_map_remove ( fcfontmap *PangoFcFontMap,
// 				PangoFcFont    *fcfont)
//  {
//    fontMapPrivate *priv = fcfontmap.priv;
//    key *PangoFcFontKey;

//    key = _pango_font_get_font_key (fcfont);
//    if (key)
// 	 {
// 	   /* Only remove from fontmap hash if we are in it.  This is not necessarily
// 		* the case after a cache_clear() call. */
// 	   if (priv.font_hash &&
// 	   fcfont == g_hash_table_lookup (priv.font_hash, key))
// 		 {
// 	   g_hash_table_remove (priv.font_hash, key);
// 	 }
// 	   _pango_font_set_font_key (fcfont, nil);
// 	   pango_font_key_free (key);
// 	 }
//  }

func pango_convert_slant_to_fc(pangoStyle pango.Style) int {
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

func pango_convert_width_to_fc(pangoStretch pango.Stretch) int {
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

func (fontmap *FontMap) newFont(FontsetKey PangoFontsetKey, match fc.Pattern) *Font {
	if fontmap.closed {
		return nil
	}

	key := FontsetKey.newFontKey(match)

	if fcfont := fontmap.font_hash.lookup(key); fcfont != nil {
		return fcfont
	}

	// TODO: check
	// class = PANGO_FONT_MAP_GET_CLASS(fontmap)

	// if class.create_font {
	// 	fcfont = class.create_font(fontmap, &key)
	// } else {
	pangoMatrix := FontsetKey.matrix
	//    FcMatrix fcMatrix, *fcMatrixVal;
	//    int i;

	// Fontconfig has the Y axis pointing up, Pango, down.
	fcMatrix := fc.Matrix{Xx: pangoMatrix.Xx, Xy: -pangoMatrix.Xy, Yx: -pangoMatrix.Yx, Yy: pangoMatrix.Yy}

	pattern := match.Duplicate()

	for _, fcMatrixVal := range pattern.GetMatrices(fc.MATRIX) {
		fcMatrix = fcMatrix.Multiply(fcMatrixVal)
	}

	pattern.Del(fc.MATRIX)
	pattern.Add(fc.MATRIX, fcMatrix, true)

	// TODO: check new_font interface
	fcfont := newFont(pattern, fontmap)

	fcfont.matrix = key.matrix

	// cache it on fontmap
	fontmap.font_hash.insert(key, fcfont)

	return fcfont
}

func (key *PangoFontsetKey) defaultSubstitute(fontmap *FontMap, pattern fc.Pattern) {
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

//  void
//  pango_font_map_set_default_substitute (PangoFcFontMap        *fontmap,
// 					   PangoFcSubstituteFunc func,
// 					   gpointer              data,
// 					   GDestroyNotify        notify)
//  {
//    if (fontmap.substitute_destroy)
// 	 fontmap.substitute_destroy (fontmap.substitute_data);

//    fontmap.substitute_func = func;
//    fontmap.substitute_data = data;
//    fontmap.substitute_destroy = notify;

//    pango_font_map_substitute_changed (fontmap);
//  }

//  void
//  pango_font_map_substitute_changed (fontmap *PangoFcFontMap) {
//    pango_font_map_cache_clear(fontmap);
//    pango_font_map_changed(PANGO_FONT_MAP (fontmap));
//  }

func (fontmap *FontMap) getResolution(*pango.Context) float32 { return fontmap.dpi_y }

//  /**
//   * pango_font_map_cache_clear:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Clear all cached information and Fontsets for this font map;
//   * this should be called whenever there is a change in the
//   * output of the default_substitute() virtual function of the
//   * font map, or if fontconfig has been reinitialized to new
//   * configuration.
//   *
//   * Since: 1.4
//   **/
//  void
//  pango_font_map_cache_clear ( fcfontmap *PangoFcFontMap)
//  {
//    guint removed, added;

//    if (G_UNLIKELY (fcfontmap.priv.closed))
// 	 return;

//    removed = fcfontmap.priv.n_families;

//    pango_font_map_fini (fcfontmap);
//    pango_font_map_init (fcfontmap);

//    ensureFamilies (fcfontmap);

//    added = fcfontmap.priv.n_families;

//    g_list_model_items_changed (G_LIST_MODEL (fcfontmap), 0, removed, added);

//    pango_font_map_changed (PANGO_FONT_MAP (fcfontmap));
//  }

//  static void
//  pango_font_map_changed (PangoFontMap *fontmap)
//  {
//    /* we emit GListModel::changed in pango_font_map_cache_clear() */
//  }

//  /**
//   * pango_font_map_config_changed:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Informs font map that the fontconfig configuration (ie, Config object)
//   * used by this font map has changed.  This currently calls
//   * pango_font_map_cache_clear() which ensures that list of fonts, etc
//   * will be regenerated using the updated configuration.
//   *
//   * Since: 1.38
//   **/
//  void
//  pango_font_map_config_changed ( fcfontmap *PangoFcFontMap)
//  {
//    pango_font_map_cache_clear (fcfontmap);
//  }

//  /**
//   * pango_font_map_set_config: (skip)
//   * @fcfontmap: a #PangoFcFontMap
//   * @Config: (nullable): a `Config`, or %nil
//   *
//   * Set the Config for this font map to use.  The default value
//   * is %nil, which causes Fontconfig to use its global "current config".
//   * You can create a new Config object and use this API to attach it
//   * to a font map.
//   *
//   * This is particularly useful for example, if you want to use application
//   * fonts with Pango.  For that, you would create a fresh Config, add your
//   * app fonts to it, and attach it to a new Pango font map.
//   *
//   * If @Config is different from the previous config attached to the font map,
//   * pango_font_map_config_changed() is called.
//   *
//   * This function acquires a reference to the Config object; the caller
//   * does NOT need to retain a reference.
//   *
//   * Since: 1.38
//   **/
//  void
//  pango_font_map_set_config ( fcfontmap *PangoFcFontMap,
// 				   Config       *Config)
//  {
//    Config *oldconfig;

//    g_return_if_fail (PANGO_IS_FONT_MAP (fcfontmap));

//    oldconfig = fcfontmap.priv.config;

//    if (Config)
// 	 ConfigReference (Config);

//    fcfontmap.priv.config = Config;

//    if (oldconfig != Config)
// 	 pango_font_map_config_changed (fcfontmap);

//    if (oldconfig)
// 	 ConfigDestroy (oldconfig);
//  }

//  /**
//   * pango_font_map_get_config: (skip)
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Fetches the `Config` attached to a font map.
//   *
//   * See also: pango_font_map_set_config()
//   *
//   * Returns: (nullable): the `Config` object attached to @fcfontmap, which
//   *          might be %nil.
//   *
//   * Since: 1.38
//   **/
//  Config *
//  pango_font_map_get_config ( fcfontmap *PangoFcFontMap)
//  {
//    g_return_val_if_fail (PANGO_IS_FONT_MAP (fcfontmap), nil);

//    return fcfontmap.priv.config;
//  }

//  typedef struct {
//    PangoCoverage parent_instance;

//    FcCharSet *charset;
//  } PangoFcCoverage;

//  typedef struct {
//    PangoCoverageClass parent_class;
//  } PangoFcCoverageClass;

//  GType pango_coverage_get_type (void) G_GNUC_CONST;

//  G_DEFINE_TYPE (PangoFcCoverage, pango_coverage, PANGO_TYPE_COVERAGE)

//  static void
//  pango_coverage_init (PangoFcCoverage *coverage)
//  {
//  }

//  static PangoCoverageLevel
//  pango_coverage_real_get (PangoCoverage *coverage,
// 							 int            index)
//  {
//    PangoFcCoverage *coverage = (PangoFcCoverage*)coverage;

//    return FcCharSetHasChar (coverage.charset, index)
// 		  ? PANGO_COVERAGE_EXACT
// 		  : PANGO_COVERAGE_NONE;
//  }

//  static void
//  pango_coverage_real_set (PangoCoverage *coverage,
// 							 int            index,
// 							 PangoCoverageLevel level)
//  {
//    PangoFcCoverage *coverage = (PangoFcCoverage*)coverage;

//    if (level == PANGO_COVERAGE_NONE)
// 	 FcCharSetDelChar (coverage.charset, index);
//    else
// 	 FcCharSetAddChar (coverage.charset, index);
//  }

//  static PangoCoverage *
//  pango_coverage_real_copy (PangoCoverage *coverage)
//  {
//    PangoFcCoverage *coverage = (PangoFcCoverage*)coverage;
//    PangoFcCoverage *copy;

//    copy = g_object_new (pango_coverage_get_type (), nil);
//    copy.charset = FcCharSetCopy (coverage.charset);

//    return (PangoCoverage *)copy;
//  }

//  static void
//  pango_coverage_finalize (GObject *object)
//  {
//    PangoFcCoverage *coverage = (PangoFcCoverage*)object;

//    FcCharSetDestroy (coverage.charset);

//    G_OBJECT_CLASS (pango_coverage_parent_class).finalize (object);
//  }

//  static void
//  pango_coverage_class_init (PangoFcCoverageClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoCoverageClass *coverage_class = PANGO_COVERAGE_CLASS (class);

//    object_class.finalize = pango_coverage_finalize;

//    coverage_class.get = pango_coverage_real_get;
//    coverage_class.set = pango_coverage_real_set;
//    coverage_class.copy = pango_coverage_real_copy;
//  }

//  static PangoLanguage **
//  _pango_font_map_to_languages (langSet *langset)
//  {
//    FcStrSet *strset;
//    FcStrList *list;
//    FcChar8 *s;
//    GArray *langs;

//    langs = g_array_new (true, false, sizeof (PangoLanguage *));

//    strset = langSetGetLangs (langset);
//    list = FcStrListCreate (strset);

//    FcStrListFirst (list);
//    while ((s = FcStrListNext (list)))
// 	 {
// 	   PangoLanguage *l = pango_language_from_string ((const char *)s);
// 	   g_array_append_val (langs, l);
// 	 }

//    FcStrListDone (list);
//    FcStrSetDestroy (strset);

//    return (PangoLanguage **) g_array_free (langs, false);
//  }

//  PangoLanguage **
//  _pango_font_map_get_languages ( fcfontmap *PangoFcFontMap,
// 								   PangoFcFont    *fcfont)
//  {
//    faceData *data;
//    langSet *langset;

//    data = getFontFaceData (fcfontmap, fcfont.font_pattern);
//    if (G_UNLIKELY (!data))
// 	 return nil;

//    if (G_UNLIKELY (data.languages == nil))
// 	 {
// 	   /*
// 		* Pull the languages out of the pattern, this
// 		* doesn't require loading the font
// 		*/
// 	   if (PatternGetLangSet (fcfont.font_pattern, LANG, 0, &langset) != ResultMatch)
// 		 return nil;

// 	   data.languages = _pango_font_map_to_languages (langset);
// 	 }

//    return data.languages;
//  }
//  /**
//   * pango_font_map_create_context:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Creates a new context for this fontmap. This function is intended
//   * only for backend implementations deriving from #PangoFcFontMap;
//   * it is possible that a backend will store additional information
//   * needed for correct operation on the #Context after calling
//   * this function.
//   *
//   * Return value: (transfer full): a new #Context
//   *
//   * Since: 1.4
//   *
//   * Deprecated: 1.22: Use NewContext() instead.
//   **/
//  Context *
//  pango_font_map_create_context ( fcfontmap *PangoFcFontMap)
//  {
//    g_return_val_if_fail (PANGO_IS_FONT_MAP (fcfontmap), nil);

//    return NewContext (PANGO_FONT_MAP (fcfontmap));
//  }

//  static void
//  shutdown_font (gpointer        key,
// 			PangoFcFont    *fcfont,
// 			 fcfontmap *PangoFcFontMap)
//  {
//    _pango_font_shutdown (fcfont);

//    _pango_font_set_font_key (fcfont, nil);
//    pango_font_key_free (key);
//  }

//  /**
//   * pango_font_map_shutdown:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Clears all cached information for the fontmap and marks
//   * all fonts open for the fontmap as dead. (See the shutdown()
//   * virtual function of #PangoFcFont.) This function might be used
//   * by a backend when the underlying windowing system for the font
//   * map exits. This function is only intended to be called
//   * only for backend implementations deriving from #PangoFcFontMap.
//   *
//   * Since: 1.4
//   **/
//  void
//  pango_font_map_shutdown ( fcfontmap *PangoFcFontMap)
//  {
//    fontMapPrivate *priv = fcfontmap.priv;
//    int i;

//    if (priv.closed)
// 	 return;

//    g_hash_table_foreach (priv.font_hash, (GHFunc) shutdown_font, fcfontmap);
//    for (i = 0; i < priv.n_families; i++)
// 	 priv.families[i].fontmap = nil;

//    pango_font_map_fini (fcfontmap);

//    while (priv.findfuncs)
// 	 {
// 	   PangoFcFindFuncInfo *info;
// 	   info = priv.findfuncs.data;
// 	   if (info.dnotify)
// 	 info.dnotify (info.user_data);

// 	   g_slice_free (PangoFcFindFuncInfo, info);
// 	   priv.findfuncs = g_slist_delete_link (priv.findfuncs, priv.findfuncs);
// 	 }

//    priv.closed = true;
//  }

//  static PangoWeight
//  pango_convert_weight_to_pango (float64 weight)
//  {
//  #ifdef HAVE_FCWEIGHTFROMOPENTYPEDOUBLE
//    return FcWeightToOpenTypeDouble (weight);
//  #else
//    return FcWeightToOpenType (weight);
//  #endif
//  }

//  static PangoStyle
//  pango_convert_slant_to_pango (int style)
//  {
//    switch (style)
// 	 {
// 	 case pango.SLANT_ROMAN:
// 	   return STYLE_NORMAL;
// 	 case pango.SLANT_ITALIC:
// 	   return STYLE_ITALIC;
// 	 case pango.SLANT_OBLIQUE:
// 	   return STYLE_OBLIQUE;
// 	 default:
// 	   return STYLE_NORMAL;
// 	 }
//  }

//  static PangoStretch
//  pango_convert_width_to_pango (int stretch)
//  {
//    switch (stretch)
// 	 {
// 	 case WIDTH_NORMAL:
// 	   return STRETCH_NORMAL;
// 	 case WIDTH_ULTRACONDENSED:
// 	   return STRETCH_ULTRA_CONDENSED;
// 	 case WIDTH_EXTRACONDENSED:
// 	   return STRETCH_EXTRA_CONDENSED;
// 	 case WIDTH_CONDENSED:
// 	   return STRETCH_CONDENSED;
// 	 case WIDTH_SEMICONDENSED:
// 	   return STRETCH_SEMI_CONDENSED;
// 	 case WIDTH_SEMIEXPANDED:
// 	   return STRETCH_SEMI_EXPANDED;
// 	 case WIDTH_EXPANDED:
// 	   return STRETCH_EXPANDED;
// 	 case WIDTH_EXTRAEXPANDED:
// 	   return STRETCH_EXTRA_EXPANDED;
// 	 case WIDTH_ULTRAEXPANDED:
// 	   return STRETCH_ULTRA_EXPANDED;
// 	 default:
// 	   return STRETCH_NORMAL;
// 	 }
//  }

//  /*
//   * PangoFcFace
//   */

//  typedef PangoFontFaceClass PangoFcFaceClass;

//  G_DEFINE_TYPE (PangoFcFace, pango_face, PANGO_TYPE_FONT_FACE)

//  static int
//  compare_ints (gconstpointer ap,
// 		   gconstpointer bp)
//  {
//    int a = *(int *)ap;
//    int b = *(int *)bp;

//    if (a == b)
// 	 return 0;
//    else if (a > b)
// 	 return 1;
//    else
// 	 return -1;
//  }

//  static void
//  pango_face_finalize (GObject *object)
//  {
//    PangoFcFace *fcface = PANGO_FACE (object);

//    g_free (fcface.style);
//    PatternDestroy (fcface.pattern);

//    G_OBJECT_CLASS (pango_face_parent_class).finalize (object);
//  }

//  static void
//  pango_face_init (PangoFcFace *self)
//  {
//  }

//  static void
//  pango_face_class_init (PangoFcFaceClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);

//    object_class.finalize = pango_face_finalize;

//    class.Describe = pango_face_Describe;
//    class.GetFaceName = pango_face_GetFaceName;
//    class.ListSizes = pango_face_ListSizes;
//    class.IsSynthesized = pango_face_IsSynthesized;
//    class.GetFamily = pango_face_GetFamily;
//  }

//  /*
//   * PangoFcFamily
//   */

//  typedef PangoFontFamilyClass PangoFcFamilyClass;

//  static GType
//  pango_family_get_item_type (GListModel *list)
//  {
//    return PANGO_TYPE_FONT_FACE;
//  }

//  static void ensure_faces (PangoFcFamily *family);

//  static guint
//  pango_family_get_n_items (GListModel *list)
//  {
//    PangoFcFamily *fcfamily = PANGO_FAMILY (list);

//    ensure_faces (fcfamily);

//    return (guint)fcfamily.n_faces;
//  }

//  static gpointer
//  pango_family_get_item (GListModel *list,
// 						   guint       position)
//  {
//    PangoFcFamily *fcfamily = PANGO_FAMILY (list);

//    ensure_faces (fcfamily);

//    if (position < fcfamily.n_faces)
// 	 return g_object_ref (fcfamily.faces[position]);

//    return nil;
//  }

//  static void
//  pango_family_list_model_init (GListModelInterface *iface)
//  {
//    iface.get_item_type = pango_family_get_item_type;
//    iface.get_n_items = pango_family_get_n_items;
//    iface.get_item = pango_family_get_item;
//  }

//  G_DEFINE_TYPE_WITH_CODE (PangoFcFamily, pango_family, PANGO_TYPE_FONT_FAMILY,
// 						  G_IMPLEMENT_INTERFACE (G_TYPE_LIST_MODEL, pango_family_list_model_init))

//  static void
//  pango_family_finalize (GObject *object)
//  {
//    int i;
//    PangoFcFamily *fcfamily = PANGO_FAMILY (object);

//    g_free (fcfamily.family_name);

//    for (i = 0; i < fcfamily.n_faces; i++)
// 	 {
// 	   fcfamily.faces[i].family = nil;
// 	   g_object_unref (fcfamily.faces[i]);
// 	 }
//    FontsetDestroy (fcfamily.patterns);
//    g_free (fcfamily.faces);

//    G_OBJECT_CLASS (pango_family_parent_class).finalize (object);
//  }

//  static void
//  pango_family_class_init (PangoFcFamilyClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);

//    object_class.finalize = pango_family_finalize;

//    class.ListFaces = pango_family_ListFaces;
//    class.GetFace = pango_family_GetFace;
//    class.GetName = pango_family_GetName;
//    class.IsMonospace = pango_family_IsMonospace;
//    class.IsVariable = pango_family_IsVariable;
//  }

//  static void
//  pango_family_init (PangoFcFamily *fcfamily)
//  {
//    fcfamily.n_faces = -1;
//  }
