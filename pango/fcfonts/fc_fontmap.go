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

const FONTSET_CACHE_SIZE = 256

/* Overview:
 *
 * All programming is a practice in caching data. PangoFcFontMap is the
 * major caching container of a Pango system on a Linux desktop. Here is
 * a short overview of how it all works.
 *
 * In short, Fontconfig search patterns are constructed and a fontset loaded
 * using them. Here is how we achieve that:
 *
 * - All FcPattern's referenced by any object in the fontmap are uniquified
 *   and cached in the fontmap. This both speeds lookups based on patterns
 *   faster, and saves memory. This is handled by fontmap.priv.pattern_hash.
 *   The patterns are cached indefinitely.
 *
 * - The results of a FcFontSort() are used to populate fontsets.  However,
 *   FcFontSort() relies on the search pattern only, which includes the font
 *   size but not the full font matrix.  The fontset however depends on the
 *   matrix.  As a result, multiple fontsets may need results of the
 *   FcFontSort() on the same input pattern (think rotating text).  As such,
 *   we cache FcFontSort() results in fontmap.priv.patterns_hash which
 *   is a refcounted structure.  This level of abstraction also allows for
 *   optimizations like calling FcFontMatch() instead of FcFontSort(), and
 *   only calling FcFontSort() if any patterns other than the first match
 *   are needed.  Another possible optimization would be to call FcFontSort()
 *   without trimming, and do the trimming lazily as we go.  Only pattern sets
 *   already referenced by a fontset are cached.
 *
 * - A number of most-recently-used fontsets are cached and reused when
 *   needed.  This is achieved using fontmap.priv.fontset_hash and
 *   fontmap.priv.fontset_cache.
 *
 * - All fonts created by any of our fontsets are also cached and reused.
 *   This is what fontmap.priv.font_hash does.
 *
 * - Data that only depends on the font file and face index is cached and
 *   reused by multiple fonts.  This includes coverage and cmap cache info.
 *   This is done using fontmap.priv.font_face_data_hash.
 *
 * Upon a cache_clear() request, all caches are emptied.  All objects (fonts,
 * fontsets, faces, families) having a reference from outside will still live
 * and may reference the fontmap still, but will not be reused by the fontmap.
 *
 *
 * Todo:
 *
 * - Make PangoCoverage a GObject and subclass it as PangoFcCoverage which
 *   will directly use FcCharset. (#569622)
 *
 * - Lazy trimming of FcFontSort() results.  Requires fontconfig with
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

// type PangoCairoFcFontMap struct {
// 	parent_instance PangoFcFontMap

// 	serial uint
// 	dpi    float64
// }

//  type  PangoFcFindFuncInfo struct
//  {
//    PangoFcDecoderFindFunc findfunc;
//    gpointer               user_data;
//    GDestroyNotify         dnotify;
//    gpointer               ddata;
//  };

//  static FcPattern *uniquifyPattern ( fcfontmap *PangoFcFontMap,
// 					 pattern *FcPattern      );

//  gpointer get_gravity_class (void);

//  gpointer
//  get_gravity_class (void)
//  {
//    static GEnumClass *class = nil; /* MT-safe */

//    if (g_once_init_enter (&class))
// 	 g_once_init_leave (&class, (gpointer)g_type_class_ref (PANGO_TYPE_GRAVITY));

//    return class;
//  }

//  static guint
//  pango_fc_font_face_data_hash (faceData *key)
//  {
//    return g_str_hash (key.filename) ^ key.id;
//  }

//  static bool
//  pango_fc_font_face_data_equal (faceData *key1,
// 					faceData *key2)
//  {
//    return key1.id == key2.id &&
// 	  (key1 == key2 || 0 == strcmp (key1.filename, key2.filename));
//  }

//  static void
//  pango_fc_font_face_data_free (faceData *data)
//  {
//    FcPatternDestroy (data.pattern);

//    if (data.coverage)
// 	 pango_coverage_unref (data.coverage);

//    g_free (data.languages);

//    hb_face_destroy (data.hb_face);

//    g_slice_free (faceData, data);
//  }

//  /* Fowler / Noll / Vo (FNV) Hash (http://www.isthe.com/chongo/tech/comp/fnv/)
//   *
//   * Not necessarily better than a lot of other hashes, but should be OK, and
//   * well tested with binary data.
//   */

//  #define FNV_32_PRIME ((guint32)0x01000193)
//  #define FNV1_32_INIT ((guint32)0x811c9dc5)

//  static guint32
//  hash_bytes_fnv (unsigned char *buffer,
// 		 int            len,
// 		 guint32        hval)
//  {
//    while (len--)
// 	 {
// 	   hval *= FNV_32_PRIME;
// 	   hval ^= *buffer++;
// 	 }

//    return hval;
//  }

func (fcfontmap *FontMap) getScaledSize(context *pango.Context, desc *pango.FontDescription) int {
	size := float64(desc.Size)

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
	matrix      pango.Matrix
	context_key int
	variations  string
}

func (fontsetKey *PangoFcFontsetKey) newFontKey(pattern fc.Pattern) PangoFcFontKey {
	var key PangoFcFontKey
	key.pattern = pattern
	key.matrix = fontsetKey.matrix
	key.variations = fontsetKey.variations
	key.context_key = fontsetKey.context_key
	return key
}

func (key *PangoFcFontKey) pango_fc_font_key_get_gravity() pango.Gravity {
	gravity := pango.PANGO_GRAVITY_SOUTH

	pattern := key.pattern

	if s, ok := pattern.GetString(fcGravity); ok {
		value, _ := pango.GravityMap.FromString(s)
		gravity = pango.Gravity(value)
	}

	return gravity
}

func (key *PangoFcFontKey) get_font_size() float64 {
	if size, ok := key.pattern.GetFloat(fc.FC_PIXEL_SIZE); ok {
		return size
	}

	/* Just in case FC_PIXEL_SIZE got unset between pango_fc_make_pattern()
	* and here. That would be very weird. */
	dpi, ok := key.pattern.GetFloat(fc.FC_DPI)
	if !ok {
		dpi = 72
	}

	if size, ok := key.pattern.GetFloat(fc.FC_SIZE); ok {
		return size * dpi / 72.
	}

	// Whatever
	return 18.
}

type PangoFcFontsetKey struct {
	fontmap     *FontMap
	language    pango.Language
	desc        pango.FontDescription
	matrix      pango.Matrix
	pixelsize   int
	resolution  float64
	context_key int
	variations  string
}

func (fcfontmap *FontMap) newFontsetKey(context *pango.Context, desc *pango.FontDescription, language pango.Language) PangoFcFontsetKey {
	if language == "" && context != nil {
		language = context.GetLanguage()
	}

	var key PangoFcFontsetKey
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

	if context != nil {
		key.context_key = fcfontmap.context_key_get(context)
	}
	return key
}

func (key *PangoFcFontsetKey) pango_fc_fontset_key_make_pattern() fc.Pattern {
	slant := pango_fc_convert_slant_to_fc(key.desc.Style)
	weight := fc.FcWeightFromOpenTypeDouble(float64(key.desc.Weight))
	width := pango_fc_convert_width_to_fc(key.desc.Stretch)

	gravity := key.desc.Gravity
	vertical := fc.FcFalse
	if gravity.IsVertical() {
		vertical = fc.FcTrue
	}

	/* The reason for passing in FC_SIZE as well as FC_PIXEL_SIZE is
	* to work around a bug in libgnomeprint where it doesn't look
	* for FC_PIXEL_SIZE. See http://bugzilla.gnome.org/show_bug.cgi?id=169020
	*
	* Putting FC_SIZE in here slightly reduces the efficiency
	* of caching of patterns and fonts when working with multiple different
	* dpi values. */
	pattern := fc.BuildPattern([]fc.PatternElement{
		// {Object: PANGO_FC_VERSION, Value: pango_version()},       // FcTypeInteger
		{Object: fc.FC_WEIGHT, Value: fc.Float(weight)},                                                // FcTypeDouble
		{Object: fc.FC_SLANT, Value: fc.Int(slant)},                                                    // FcTypeInteger
		{Object: fc.FC_WIDTH, Value: fc.Int(width)},                                                    // FcTypeInteger
		{Object: fc.FC_VERTICAL_LAYOUT, Value: vertical},                                               // FcTypeBool
		{Object: fc.FC_VARIABLE, Value: fc.FcDontCare},                                                 //  FcTypeBool
		{Object: fc.FC_DPI, Value: fc.Float(key.resolution)},                                           // FcTypeDouble
		{Object: fc.FC_SIZE, Value: fc.Float(float64(key.pixelsize) * (72. / 1024. / key.resolution))}, // FcTypeDouble
		{Object: fc.FC_PIXEL_SIZE, Value: fc.Float(key.pixelsize) / 1024.},                             // FcTypeDouble
	}...)

	if key.variations != "" {
		pattern.Add(fc.FC_FONT_VARIATIONS, fc.String(key.variations), true)
	}

	if key.desc.FamilyName != "" {
		families := strings.Split(key.desc.FamilyName, ",")
		for _, fam := range families {
			pattern.Add(fc.FC_FAMILY, fc.String(fam), true)
		}
	}

	if key.language != "" {
		pattern.Add(fc.FC_LANG, fc.String(key.language), true)
	}

	if gravity != pango.PANGO_GRAVITY_SOUTH {
		pattern.Add(fcGravity, fc.String(pango.GravityMap.ToString("gravity", int(gravity))), true)
	}

	return pattern
}

//  static bool
//  pango_fc_fontset_key_equal (const key *PangoFcFontsetKey_a,
// 				 const key *PangoFcFontsetKey_b)
//  {
//    if (key_a.language == key_b.language &&
// 	   key_a.pixelsize == key_b.pixelsize &&
// 	   key_a.resolution == key_b.resolution &&
// 	   ((key_a.variations == nil && key_b.variations == nil) ||
// 		(key_a.variations && key_b.variations && (strcmp (key_a.variations, key_b.variations) == 0))) &&
// 	   pango_font_description_equal (key_a.desc, key_b.desc) &&
// 	   0 == memcmp (&key_a.matrix, &key_b.matrix, 4 * sizeof (float64)))
// 	 {
// 	   if (key_a.context_key)
// 	 return PANGO_FC_FONT_MAP_GET_CLASS (key_a.fontmap).context_key_equal (key_a.fontmap,
// 										 key_a.context_key,
// 										 key_b.context_key);
// 	   else
// 		 return key_a.context_key == key_b.context_key;
// 	 }
//    else
// 	 return false;
//  }

//  static void
//  pango_fc_fontset_key_free (key *PangoFcFontsetKey)
//  {
//    pango_font_description_free (key.desc);
//    g_free (key.variations);

//    if (key.context_key)
// 	 PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_free (key.fontmap,
// 								   key.context_key);

//    g_slice_free (PangoFcFontsetKey, key);
//  }

//  /**
//   * pango_fc_fontset_key_get_language:
//   * @key: the fontset key
//   *
//   * Gets the language member of @key.
//   *
//   * Returns: the language
//   *
//   * Since: 1.24
//   **/
//  PangoLanguage *
//  pango_fc_fontset_key_get_language (const key *PangoFcFontsetKey)
//  {
//    return key.language;
//  }

//  /**
//   * pango_fc_fontset_key_get_description:
//   * @key: the fontset key
//   *
//   * Gets the font description of @key.
//   *
//   * Returns: the font description, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const FontDescription *
//  pango_fc_fontset_key_get_description (const key *PangoFcFontsetKey)
//  {
//    return key.desc;
//  }

//  /**
//   * pango_fc_fontset_key_get_matrix:
//   * @key: the fontset key
//   *
//   * Gets the matrix member of @key.
//   *
//   * Returns: the matrix, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const Matrix *
//  pango_fc_fontset_key_get_matrix      (const key *PangoFcFontsetKey)
//  {
//    return &key.matrix;
//  }

//  /**
//   * pango_fc_fontset_key_get_absolute_size:
//   * @key: the fontset key
//   *
//   * Gets the absolute font size of @key in Pango units.  This is adjusted
//   * for both resolution and transformation matrix.
//   *
//   * Returns: the pixel size of @key.
//   *
//   * Since: 1.24
//   **/
//  float64
//  pango_fc_fontset_key_get_absolute_size   (const key *PangoFcFontsetKey)
//  {
//    return key.pixelsize;
//  }

//  /**
//   * pango_fc_fontset_key_get_resolution:
//   * @key: the fontset key
//   *
//   * Gets the resolution of @key
//   *
//   * Returns: the resolution of @key
//   *
//   * Since: 1.24
//   **/
//  float64
//  pango_fc_fontset_key_get_resolution  (const key *PangoFcFontsetKey)
//  {
//    return key.resolution;
//  }

//  /**
//   * pango_fc_fontset_key_get_context_key:
//   * @key: the font key
//   *
//   * Gets the context key member of @key.
//   *
//   * Returns: the context key, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  gpointer
//  pango_fc_fontset_key_get_context_key (const key *PangoFcFontsetKey)
//  {
//    return key.context_key;
//  }

//  /*
//   * PangoFcFontKey
//   */

//  static guint
//  pango_fc_font_key_hash (const key *PangoFcFontKey)
//  {
// 	 guint32 hash = FNV1_32_INIT;

// 	 /* We do a bytewise hash on the doubles */
// 	 hash = hash_bytes_fnv ((unsigned char *)(&key.matrix), sizeof (float64) * 4, hash);

// 	 if (key.variations)
// 	   hash ^= g_str_hash (key.variations);

// 	 if (key.context_key)
// 	   hash ^= PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_hash (key.fontmap,
// 										 key.context_key);

// 	 return (hash ^ GPOINTER_TO_UINT (key.pattern));
//  }

//  static void
//  pango_fc_font_key_free (key *PangoFcFontKey)
//  {
//    if (key.pattern)
// 	 FcPatternDestroy (key.pattern);

//    if (key.context_key)
// 	 PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_free (key.fontmap,
// 								   key.context_key);

//    g_free (key.variations);

//    g_slice_free (PangoFcFontKey, key);
//  }

//  static PangoFcFontKey *
//  pango_fc_font_key_copy (const PangoFcFontKey *old)
//  {
//    key *PangoFcFontKey = g_slice_new (PangoFcFontKey);

//    key.fontmap = old.fontmap;
//    FcPatternReference (old.pattern);
//    key.pattern = old.pattern;
//    key.matrix = old.matrix;
//    key.variations = g_strdup (old.variations);
//    if (old.context_key)
// 	 key.context_key = PANGO_FC_FONT_MAP_GET_CLASS (key.fontmap).context_key_copy (key.fontmap,
// 											  old.context_key);
//    else
// 	 key.context_key = nil;

//    return key;
//  }

//  /* Public API */

//  /**
//   * pango_fc_font_key_get_pattern:
//   * @key: the font key
//   *
//   * Gets the fontconfig pattern member of @key.
//   *
//   * Returns: the pattern, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const FcPattern *
//  pango_fc_font_key_get_pattern (const key *PangoFcFontKey)
//  {
//    return key.pattern;
//  }

//  /**
//   * pango_fc_font_key_get_matrix:
//   * @key: the font key
//   *
//   * Gets the matrix member of @key.
//   *
//   * Returns: the matrix, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  const Matrix *
//  pango_fc_font_key_get_matrix (const key *PangoFcFontKey)
//  {
//    return &key.matrix;
//  }

//  /**
//   * pango_fc_font_key_get_context_key:
//   * @key: the font key
//   *
//   * Gets the context key member of @key.
//   *
//   * Returns: the context key, which is owned by @key and should not be modified.
//   *
//   * Since: 1.24
//   **/
//  gpointer
//  pango_fc_font_key_get_context_key (const key *PangoFcFontKey)
//  {
//    return key.context_key;
//  }

//  const char *
//  pango_fc_font_key_get_variations (const key *PangoFcFontKey)
//  {
//    return key.variations;
//  }

// ------------------------------- PangoFcPatterns -------------------------------

type fcPatterns struct {
	fontmap *FontMap

	pattern fc.Pattern
	match   fc.Pattern
	fontset fc.FontSet
}

func (fontmap *FontMap) pango_fc_patterns_new(pat fc.Pattern) *fcPatterns {

	if pats := fontmap.patterns_hash.lookup(pat); pats != nil {
		return pats
	}

	var pats fcPatterns

	pats.fontmap = fontmap
	pats.pattern = pat
	fontmap.patterns_hash.insert(pat, &pats)

	return &pats
}

func pango_fc_is_supported_font_format(pattern fc.Pattern) bool {
	fontformat, res := pattern.GetAtString(fc.FC_FONTFORMAT, 0)
	if res != fc.FcResultMatch {
		return false
	}

	/* harfbuzz supports only SFNT fonts. */
	/* FIXME: "CFF" is used for both CFF in OpenType and bare CFF files, but
	* HarfBuzz does not support the later and FontConfig does not seem
	* to have a way to tell them apart.
	 */
	return fontformat == "TrueType" || fontformat == "CFF"
}

func filter_fontset_by_format(fontset fc.FontSet) fc.FontSet {
	var result fc.FontSet

	for _, fontPattern := range fontset {
		if pango_fc_is_supported_font_format(fontPattern) {
			result = append(result, fontPattern)
		}
	}

	return result
}

func (pats *fcPatterns) pango_fc_patterns_get_font_pattern(i int) (fc.Pattern, bool) {
	if i == 0 {
		if pats.match == nil && pats.fontset == nil {
			pats.match = pats.fontmap.fontset.Match(pats.pattern, pats.fontmap.config)
		}

		if pats.match != nil && pango_fc_is_supported_font_format(pats.match) {
			return pats.match, false
		}
	}

	if pats.fontset == nil {
		var (
			filtered [2]fc.FontSet
			n        int
		)

		for i := range filtered {
			fonts := pats.fontmap.config.FcConfigGetFonts(fc.FcSetName(i))
			if fonts != nil {
				filtered[n] = filter_fontset_by_format(fonts)
				n++
			}
		}

		pats.fontset, _, _ = fc.Sort(filtered[:], pats.pattern, true)

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
//  pango_fc_font_map_get_n_items (GListModel *list)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (list);

//    ensureFamilies (fcfontmap);

//    return fcfontmap.priv.n_families;
//  }

//  static gpointer
//  pango_fc_font_map_get_item (GListModel *list,
// 							 guint       position)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (list);

//    ensureFamilies (fcfontmap);

//    if (position < fcfontmap.priv.n_families)
// 	 return g_object_ref (fcfontmap.priv.families[position]);

//    return nil;
//  }

//  static void
//  pango_fc_font_map_list_model_init (GListModelInterface *iface)
//  {
//    iface.get_item_type = pango_fc_font_map_get_item_type;
//    iface.get_n_items = pango_fc_font_map_get_n_items;
//    iface.get_item = pango_fc_font_map_get_item;
//  }

//  /**
//   * pango_fc_font_map_add_decoder_find_func:
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
//  pango_fc_font_map_add_decoder_find_func (PangoFcFontMap        *fcfontmap,
// 					  PangoFcDecoderFindFunc findfunc,
// 					  gpointer               user_data,
// 					  GDestroyNotify         dnotify)
//  {
//    fontMapPrivate *priv;
//    PangoFcFindFuncInfo *info;

//    g_return_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap));

//    priv = fcfontmap.priv;

//    info = g_slice_new (PangoFcFindFuncInfo);

//    info.findfunc = findfunc;
//    info.user_data = user_data;
//    info.dnotify = dnotify;

//    priv.findfuncs = g_slist_append (priv.findfuncs, info);
//  }

//  /**
//   * pango_fc_font_map_find_decoder:
//   * @fcfontmap: The #PangoFcFontMap to use.
//   * @pattern: The #FcPattern to find the decoder for.
//   *
//   * Finds the decoder to use for @pattern.  Decoders can be added to
//   * a font map using pango_fc_font_map_add_decoder_find_func().
//   *
//   * Returns: (transfer full) (nullable): a newly created #PangoFcDecoder
//   *   object or %nil if no decoder is set for @pattern.
//   *
//   * Since: 1.26
//   **/
//  PangoFcDecoder *
//  pango_fc_font_map_find_decoder  ( fcfontmap *PangoFcFontMap,
// 				  pattern *FcPattern      )
//  {
//    GSList *l;

//    g_return_val_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap), nil);
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
//  pango_fc_font_map_finalize (GObject *object)
//  {
//     fcfontmap *PangoFcFontMap = PANGO_FC_FONT_MAP (object);

//    pango_fc_font_map_shutdown (fcfontmap);

//    if (fcfontmap.substitute_destroy)
// 	 fcfontmap.substitute_destroy (fcfontmap.substitute_data);

//    G_OBJECT_CLASS (pango_fc_font_map_parent_class).finalize (object);
//  }

//  /* Remove mapping from fcfont.key to fcfont */
//  /* Closely related to shutdown_font() */
//  void
//  _pango_fc_font_map_remove ( fcfontmap *PangoFcFontMap,
// 				PangoFcFont    *fcfont)
//  {
//    fontMapPrivate *priv = fcfontmap.priv;
//    key *PangoFcFontKey;

//    key = _pango_fc_font_get_font_key (fcfont);
//    if (key)
// 	 {
// 	   /* Only remove from fontmap hash if we are in it.  This is not necessarily
// 		* the case after a cache_clear() call. */
// 	   if (priv.font_hash &&
// 	   fcfont == g_hash_table_lookup (priv.font_hash, key))
// 		 {
// 	   g_hash_table_remove (priv.font_hash, key);
// 	 }
// 	   _pango_fc_font_set_font_key (fcfont, nil);
// 	   pango_fc_font_key_free (key);
// 	 }
//  }

func pango_fc_convert_slant_to_fc(pangoStyle pango.Style) int {
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

func pango_fc_convert_width_to_fc(pangoStretch pango.Stretch) int {
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

func (fontmap *FontMap) newFont(fontsetKey PangoFcFontsetKey, match fc.Pattern) *Font {
	if fontmap.Closed {
		return nil
	}

	key := fontsetKey.newFontKey(match)

	if fcfont := fontmap.font_hash.lookup(key); fcfont != nil {
		return fcfont
	}

	// TODO: check
	// class = PANGO_FC_FONT_MAP_GET_CLASS(fontmap)

	// if class.create_font {
	// 	fcfont = class.create_font(fontmap, &key)
	// } else {
	pangoMatrix := fontsetKey.matrix
	//    FcMatrix fcMatrix, *fcMatrixVal;
	//    int i;

	// Fontconfig has the Y axis pointing up, Pango, down.
	fcMatrix := fc.Matrix{Xx: pangoMatrix.Xx, Xy: -pangoMatrix.Xy, Yx: -pangoMatrix.Yx, Yy: pangoMatrix.Yy}

	pattern := match.Duplicate()

	for _, fcMatrixVal := range pattern.GetMatrices(fc.FC_MATRIX) {
		fcMatrix = fcMatrix.Multiply(fcMatrixVal)
	}

	pattern.Del(fc.FC_MATRIX)
	pattern.Add(fc.FC_MATRIX, fcMatrix, true)

	// TODO: check new_font interface
	fcfont := newFont(pattern, fontmap)

	fcfont.matrix = key.matrix

	// cache it on fontmap
	fontmap.font_hash.insert(key, fcfont)

	return fcfont
}

func (fontsetkey *PangoFcFontsetKey) pango_fc_default_substitute(fontmap *FontMap, pattern fc.Pattern) {
	if fontmap.fontset_key_substitute != nil {
		fontmap.fontset_key_substitute(fontsetkey, pattern)
	} else if fontmap.default_substitute != nil {
		fontmap.default_substitute(pattern)
	}
}

//  void
//  pango_fc_font_map_set_default_substitute (PangoFcFontMap        *fontmap,
// 					   PangoFcSubstituteFunc func,
// 					   gpointer              data,
// 					   GDestroyNotify        notify)
//  {
//    if (fontmap.substitute_destroy)
// 	 fontmap.substitute_destroy (fontmap.substitute_data);

//    fontmap.substitute_func = func;
//    fontmap.substitute_data = data;
//    fontmap.substitute_destroy = notify;

//    pango_fc_font_map_substitute_changed (fontmap);
//  }

//  void
//  pango_fc_font_map_substitute_changed (fontmap *PangoFcFontMap) {
//    pango_fc_font_map_cache_clear(fontmap);
//    pango_font_map_changed(PANGO_FONT_MAP (fontmap));
//  }

func (fontmap *FontMap) getResolution(*pango.Context) float64 { return fontmap.dpi_y }

//  /**
//   * pango_fc_font_map_cache_clear:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Clear all cached information and fontsets for this font map;
//   * this should be called whenever there is a change in the
//   * output of the default_substitute() virtual function of the
//   * font map, or if fontconfig has been reinitialized to new
//   * configuration.
//   *
//   * Since: 1.4
//   **/
//  void
//  pango_fc_font_map_cache_clear ( fcfontmap *PangoFcFontMap)
//  {
//    guint removed, added;

//    if (G_UNLIKELY (fcfontmap.priv.closed))
// 	 return;

//    removed = fcfontmap.priv.n_families;

//    pango_fc_font_map_fini (fcfontmap);
//    pango_fc_font_map_init (fcfontmap);

//    ensureFamilies (fcfontmap);

//    added = fcfontmap.priv.n_families;

//    g_list_model_items_changed (G_LIST_MODEL (fcfontmap), 0, removed, added);

//    pango_font_map_changed (PANGO_FONT_MAP (fcfontmap));
//  }

//  static void
//  pango_fc_font_map_changed (PangoFontMap *fontmap)
//  {
//    /* we emit GListModel::changed in pango_fc_font_map_cache_clear() */
//  }

//  /**
//   * pango_fc_font_map_config_changed:
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Informs font map that the fontconfig configuration (ie, FcConfig object)
//   * used by this font map has changed.  This currently calls
//   * pango_fc_font_map_cache_clear() which ensures that list of fonts, etc
//   * will be regenerated using the updated configuration.
//   *
//   * Since: 1.38
//   **/
//  void
//  pango_fc_font_map_config_changed ( fcfontmap *PangoFcFontMap)
//  {
//    pango_fc_font_map_cache_clear (fcfontmap);
//  }

//  /**
//   * pango_fc_font_map_set_config: (skip)
//   * @fcfontmap: a #PangoFcFontMap
//   * @fcconfig: (nullable): a `FcConfig`, or %nil
//   *
//   * Set the FcConfig for this font map to use.  The default value
//   * is %nil, which causes Fontconfig to use its global "current config".
//   * You can create a new FcConfig object and use this API to attach it
//   * to a font map.
//   *
//   * This is particularly useful for example, if you want to use application
//   * fonts with Pango.  For that, you would create a fresh FcConfig, add your
//   * app fonts to it, and attach it to a new Pango font map.
//   *
//   * If @fcconfig is different from the previous config attached to the font map,
//   * pango_fc_font_map_config_changed() is called.
//   *
//   * This function acquires a reference to the FcConfig object; the caller
//   * does NOT need to retain a reference.
//   *
//   * Since: 1.38
//   **/
//  void
//  pango_fc_font_map_set_config ( fcfontmap *PangoFcFontMap,
// 				   FcConfig       *fcconfig)
//  {
//    FcConfig *oldconfig;

//    g_return_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap));

//    oldconfig = fcfontmap.priv.config;

//    if (fcconfig)
// 	 FcConfigReference (fcconfig);

//    fcfontmap.priv.config = fcconfig;

//    if (oldconfig != fcconfig)
// 	 pango_fc_font_map_config_changed (fcfontmap);

//    if (oldconfig)
// 	 FcConfigDestroy (oldconfig);
//  }

//  /**
//   * pango_fc_font_map_get_config: (skip)
//   * @fcfontmap: a #PangoFcFontMap
//   *
//   * Fetches the `FcConfig` attached to a font map.
//   *
//   * See also: pango_fc_font_map_set_config()
//   *
//   * Returns: (nullable): the `FcConfig` object attached to @fcfontmap, which
//   *          might be %nil.
//   *
//   * Since: 1.38
//   **/
//  FcConfig *
//  pango_fc_font_map_get_config ( fcfontmap *PangoFcFontMap)
//  {
//    g_return_val_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap), nil);

//    return fcfontmap.priv.config;
//  }

//  typedef struct {
//    PangoCoverage parent_instance;

//    FcCharSet *charset;
//  } PangoFcCoverage;

//  typedef struct {
//    PangoCoverageClass parent_class;
//  } PangoFcCoverageClass;

//  GType pango_fc_coverage_get_type (void) G_GNUC_CONST;

//  G_DEFINE_TYPE (PangoFcCoverage, pango_fc_coverage, PANGO_TYPE_COVERAGE)

//  static void
//  pango_fc_coverage_init (PangoFcCoverage *coverage)
//  {
//  }

//  static PangoCoverageLevel
//  pango_fc_coverage_real_get (PangoCoverage *coverage,
// 							 int            index)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)coverage;

//    return FcCharSetHasChar (fc_coverage.charset, index)
// 		  ? PANGO_COVERAGE_EXACT
// 		  : PANGO_COVERAGE_NONE;
//  }

//  static void
//  pango_fc_coverage_real_set (PangoCoverage *coverage,
// 							 int            index,
// 							 PangoCoverageLevel level)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)coverage;

//    if (level == PANGO_COVERAGE_NONE)
// 	 FcCharSetDelChar (fc_coverage.charset, index);
//    else
// 	 FcCharSetAddChar (fc_coverage.charset, index);
//  }

//  static PangoCoverage *
//  pango_fc_coverage_real_copy (PangoCoverage *coverage)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)coverage;
//    PangoFcCoverage *copy;

//    copy = g_object_new (pango_fc_coverage_get_type (), nil);
//    copy.charset = FcCharSetCopy (fc_coverage.charset);

//    return (PangoCoverage *)copy;
//  }

//  static void
//  pango_fc_coverage_finalize (GObject *object)
//  {
//    PangoFcCoverage *fc_coverage = (PangoFcCoverage*)object;

//    FcCharSetDestroy (fc_coverage.charset);

//    G_OBJECT_CLASS (pango_fc_coverage_parent_class).finalize (object);
//  }

//  static void
//  pango_fc_coverage_class_init (PangoFcCoverageClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);
//    PangoCoverageClass *coverage_class = PANGO_COVERAGE_CLASS (class);

//    object_class.finalize = pango_fc_coverage_finalize;

//    coverage_class.get = pango_fc_coverage_real_get;
//    coverage_class.set = pango_fc_coverage_real_set;
//    coverage_class.copy = pango_fc_coverage_real_copy;
//  }

//  static PangoLanguage **
//  _pango_fc_font_map_fc_to_languages (FcLangSet *langset)
//  {
//    FcStrSet *strset;
//    FcStrList *list;
//    FcChar8 *s;
//    GArray *langs;

//    langs = g_array_new (true, false, sizeof (PangoLanguage *));

//    strset = FcLangSetGetLangs (langset);
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
//  _pango_fc_font_map_get_languages ( fcfontmap *PangoFcFontMap,
// 								   PangoFcFont    *fcfont)
//  {
//    faceData *data;
//    FcLangSet *langset;

//    data = getFontFaceData (fcfontmap, fcfont.font_pattern);
//    if (G_UNLIKELY (!data))
// 	 return nil;

//    if (G_UNLIKELY (data.languages == nil))
// 	 {
// 	   /*
// 		* Pull the languages out of the pattern, this
// 		* doesn't require loading the font
// 		*/
// 	   if (FcPatternGetLangSet (fcfont.font_pattern, FC_LANG, 0, &langset) != FcResultMatch)
// 		 return nil;

// 	   data.languages = _pango_fc_font_map_fc_to_languages (langset);
// 	 }

//    return data.languages;
//  }
//  /**
//   * pango_fc_font_map_create_context:
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
//  pango_fc_font_map_create_context ( fcfontmap *PangoFcFontMap)
//  {
//    g_return_val_if_fail (PANGO_IS_FC_FONT_MAP (fcfontmap), nil);

//    return NewContext (PANGO_FONT_MAP (fcfontmap));
//  }

//  static void
//  shutdown_font (gpointer        key,
// 			PangoFcFont    *fcfont,
// 			 fcfontmap *PangoFcFontMap)
//  {
//    _pango_fc_font_shutdown (fcfont);

//    _pango_fc_font_set_font_key (fcfont, nil);
//    pango_fc_font_key_free (key);
//  }

//  /**
//   * pango_fc_font_map_shutdown:
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
//  pango_fc_font_map_shutdown ( fcfontmap *PangoFcFontMap)
//  {
//    fontMapPrivate *priv = fcfontmap.priv;
//    int i;

//    if (priv.closed)
// 	 return;

//    g_hash_table_foreach (priv.font_hash, (GHFunc) shutdown_font, fcfontmap);
//    for (i = 0; i < priv.n_families; i++)
// 	 priv.families[i].fontmap = nil;

//    pango_fc_font_map_fini (fcfontmap);

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
//  pango_fc_convert_weight_to_pango (float64 fc_weight)
//  {
//  #ifdef HAVE_FCWEIGHTFROMOPENTYPEDOUBLE
//    return FcWeightToOpenTypeDouble (fc_weight);
//  #else
//    return FcWeightToOpenType (fc_weight);
//  #endif
//  }

//  static PangoStyle
//  pango_fc_convert_slant_to_pango (int fc_style)
//  {
//    switch (fc_style)
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
//  pango_fc_convert_width_to_pango (int fc_stretch)
//  {
//    switch (fc_stretch)
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

//  G_DEFINE_TYPE (PangoFcFace, pango_fc_face, PANGO_TYPE_FONT_FACE)

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
//  pango_fc_face_finalize (GObject *object)
//  {
//    PangoFcFace *fcface = PANGO_FC_FACE (object);

//    g_free (fcface.style);
//    FcPatternDestroy (fcface.pattern);

//    G_OBJECT_CLASS (pango_fc_face_parent_class).finalize (object);
//  }

//  static void
//  pango_fc_face_init (PangoFcFace *self)
//  {
//  }

//  static void
//  pango_fc_face_class_init (PangoFcFaceClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);

//    object_class.finalize = pango_fc_face_finalize;

//    class.Describe = pango_fc_face_Describe;
//    class.GetFaceName = pango_fc_face_GetFaceName;
//    class.ListSizes = pango_fc_face_ListSizes;
//    class.IsSynthesized = pango_fc_face_IsSynthesized;
//    class.GetFamily = pango_fc_face_GetFamily;
//  }

//  /*
//   * PangoFcFamily
//   */

//  typedef PangoFontFamilyClass PangoFcFamilyClass;

//  static GType
//  pango_fc_family_get_item_type (GListModel *list)
//  {
//    return PANGO_TYPE_FONT_FACE;
//  }

//  static void ensure_faces (PangoFcFamily *family);

//  static guint
//  pango_fc_family_get_n_items (GListModel *list)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (list);

//    ensure_faces (fcfamily);

//    return (guint)fcfamily.n_faces;
//  }

//  static gpointer
//  pango_fc_family_get_item (GListModel *list,
// 						   guint       position)
//  {
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (list);

//    ensure_faces (fcfamily);

//    if (position < fcfamily.n_faces)
// 	 return g_object_ref (fcfamily.faces[position]);

//    return nil;
//  }

//  static void
//  pango_fc_family_list_model_init (GListModelInterface *iface)
//  {
//    iface.get_item_type = pango_fc_family_get_item_type;
//    iface.get_n_items = pango_fc_family_get_n_items;
//    iface.get_item = pango_fc_family_get_item;
//  }

//  G_DEFINE_TYPE_WITH_CODE (PangoFcFamily, pango_fc_family, PANGO_TYPE_FONT_FAMILY,
// 						  G_IMPLEMENT_INTERFACE (G_TYPE_LIST_MODEL, pango_fc_family_list_model_init))

//  static void
//  pango_fc_family_finalize (GObject *object)
//  {
//    int i;
//    PangoFcFamily *fcfamily = PANGO_FC_FAMILY (object);

//    g_free (fcfamily.family_name);

//    for (i = 0; i < fcfamily.n_faces; i++)
// 	 {
// 	   fcfamily.faces[i].family = nil;
// 	   g_object_unref (fcfamily.faces[i]);
// 	 }
//    FcFontSetDestroy (fcfamily.patterns);
//    g_free (fcfamily.faces);

//    G_OBJECT_CLASS (pango_fc_family_parent_class).finalize (object);
//  }

//  static void
//  pango_fc_family_class_init (PangoFcFamilyClass *class)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (class);

//    object_class.finalize = pango_fc_family_finalize;

//    class.ListFaces = pango_fc_family_ListFaces;
//    class.GetFace = pango_fc_family_GetFace;
//    class.GetName = pango_fc_family_GetName;
//    class.IsMonospace = pango_fc_family_IsMonospace;
//    class.IsVariable = pango_fc_family_IsVariable;
//  }

//  static void
//  pango_fc_family_init (PangoFcFamily *fcfamily)
//  {
//    fcfamily.n_faces = -1;
//  }
