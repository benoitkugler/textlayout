package pango

import (
	"fmt"
	"log"
	"sync"
	"unicode"

	"github.com/benoitkugler/textlayout/fribidi"
)

// /**
//  * SECTION:main
//  * @title:Rendering
//  * @short_description:Functions to run the rendering pipeline
//  *
//  * The Pango rendering pipeline takes a string of
//  * Unicode characters and converts it into glyphs.
//  * The functions Described in this section accomplish
//  * various steps of this process.
//  *
//  * ![](pipeline.png)
//  */

//  /**
//   * SECTION:context
//   * @title:Contexts
//   * @short_description: Global context object
//   *
//   * The #Context structure stores global information
//   * influencing Pango's operation, such as the fontmap used
//   * to look up fonts, and default values such as the default
//   * language, default gravity, or default font.
//   */

// Context stores global information
// used to control the itemization process.
type Context struct {
	Matrix       *Matrix
	set_language Language // the global language tag for the context.
	language     Language // same as set_language but with a default value, instead of empty
	//    GObject parent_instance;

	fontMap   FontMap
	font_desc FontDescription

	serial, fontmapSerial uint

	base_dir Direction
	//    PangoGravity base_gravity;
	resolved_gravity Gravity
	gravity_hint     GravityHint

	round_glyph_positions bool
}

// NewContext creates a `Context` connected to `fontmap`,
// and initialized to default values.
func NewContext(fontmap FontMap) *Context {
	var context Context

	context.base_dir = DIRECTION_WEAK_LTR

	context.serial = 1
	context.language = DefaultLanguage()
	context.round_glyph_positions = true

	context.font_desc = NewFontDescription()
	context.font_desc.SetFamily("serif")
	context.font_desc.SetStyle(STYLE_NORMAL)
	context.font_desc.SetVariant(VARIANT_NORMAL)
	context.font_desc.SetWeight(WEIGHT_NORMAL)
	context.font_desc.SetStretch(STRETCH_NORMAL)
	context.font_desc.SetSize(12 * Scale)

	context.setFontMap(fontmap)

	return &context
}

// pango_context_load_font loads the font in one of the fontmaps in the context
// that is the closest match for `desc`, or nil if no font matched.
func (context *Context) pango_context_load_font(desc *FontDescription) Font {
	if context == nil || context.fontMap == nil {
		return nil
	}
	return LoadFont(context.fontMap, context, desc)
}

// Itemize breaks a piece of text into segments with consistent
// directional level and shaping engine, applying `attrs`.
//
// Each rune of `text` will be contained in exactly one of the items in the returned list;
// the generated list of items will be in logical order (the start
// offsets of the items are ascending).
// `startIndex` is the first rune index in `text` to process, and `length`
// the number of characters to process  after `startIndex`.
//
// `cachedIter` should be an iterator over `attrs` currently positioned at a
// range before or containing `startIndex`; `cachedIter` will be advanced to
// the range covering the position just after `startIndex` + `length`.
// (i.e. if itemizing in a loop, just keep passing in the same `cachedIter`).
func (context *Context) Itemize(text []rune, startIndex int, length int,
	attrs AttrList, cachedIter *AttrIterator) *ItemList {
	if context == nil || startIndex < 0 || length < 0 {
		return nil
	}

	return context.itemizeWithBaseDir(context.base_dir,
		text, startIndex, length, attrs, cachedIter)
}

// itemizeWithBaseDir is like `pango_itemize`, but the base direction to use when
// computing bidirectional levels is specified explicitly rather than gotten from `context`.
func (context *Context) itemizeWithBaseDir(baseDir Direction, text []rune,
	startIndex, length int,
	attrs AttrList, cachedIter *AttrIterator) *ItemList {
	if context == nil || len(text) == 0 {
		return nil
	}

	state := context.newItemizeState(text, baseDir, startIndex, length,
		attrs, cachedIter, nil)
	for do := true; do; { // do ... while
		state.processRun()
		do = state.next()
	}

	state.itemize_state_finish()

	// convert to list and reverse
	var out *ItemList
	for _, item := range state.result {
		out = &ItemList{Data: item, next: out}
	}
	return out
}

// Sets the font map to be searched when fonts are looked-up in this context.
func (context *Context) setFontMap(fontMap FontMap) {
	if fontMap == context.fontMap {
		return
	}

	context.contextChanged()

	context.fontMap = fontMap
	context.fontmapSerial = fontMap.GetSerial()
}

// SetLanguage sets the global language tag for the context. The default language
// for the locale of the running process can be found using
// pango_language_get_default().
func (context *Context) SetLanguage(language Language) {
	if language != context.language {
		context.contextChanged()
	}

	context.set_language = language
	if language != "" {
		context.language = language
	} else {
		context.language = DefaultLanguage()
	}
}

// GetMetrics get overall metric information for a particular font
// description.
//
// Since the metrics may be substantially different for
// different scripts, a language tag can be provided to indicate that
// the metrics should be retrieved that correspond to the script(s)
// used by that language. Empty language means that the language tag from the context
// will be used. If no language tag is set on the context, metrics
// for the default language (as determined by pango_language_get_default())
// will be returned.
//
// The `FontDescription` is interpreted in the same way as
// by pango_itemize(), and the family name may be a comma separated
// list of figures. If characters from multiple of these families
// would be used to render the string, then the returned fonts would
// be a composite of the metrics for the fonts loaded for the
// individual families.
// `nil` means that the font description from the context will be used.
func (context *Context) GetMetrics(desc *FontDescription, lang Language) FontMetrics {
	if desc == nil {
		desc = &context.font_desc
	}

	if lang == "" {
		lang = context.language
	}

	currentFonts := context.fontMap.LoadFontset(context, desc, lang)
	metrics := getBaseMetrics(currentFonts)

	sampleStr := []rune(GetSampleString(lang))
	items := context.itemize_with_font(sampleStr, desc)

	metrics.update_metrics_from_items(lang, sampleStr, items)

	return metrics
}

// contextChanged forces a change in the context, which will cause any `Layout`
// using this context to re-layout.
//
// This function is only useful when implementing a new backend
// for Pango, something applications won't do. Backends should
// call this function if they have attached extra data to the context
// and such data is changed.
func (context *Context) contextChanged() {
	context.serial++
	if context.serial == 0 {
		context.serial++
	}
}

// GetLanguage retrieves the global language tag for the context.
func (context *Context) GetLanguage() Language { return context.set_language }

// SetFontDescription sets the default font description for the context
func (context *Context) SetFontDescription(desc FontDescription) {
	if !desc.pango_font_description_equal(context.font_desc) {
		context.contextChanged()
		context.font_desc = desc
	}
}

//  static void
//  update_resolved_gravity (context *Context)
//  {
//    if (context.base_gravity == PANGO_GRAVITY_AUTO)
// 	 context.resolved_gravity = pango_gravity_get_for_matrix (context.matrix);
//    else
// 	 context.resolved_gravity = context.base_gravity;
//  }

//  /**
//   * pango_context_set_matrix:
//   * `context`: a #Context
//   * @matrix: (allow-none): a #PangoMatrix, or %nil to unset any existing
//   * matrix. (No matrix set is the same as setting the identity matrix.)
//   *
//   * Sets the transformation matrix that will be applied when rendering
//   * with this context. Note that reported metrics are in the user space
//   * coordinates before the application of the matrix, not device-space
//   * coordinates after the application of the matrix. So, they don't scale
//   * with the matrix, though they may change slightly for different
//   * matrices, depending on how the text is fit to the pixel grid.
//   *
//   * Since: 1.6
//   **/
//  void
//  pango_context_set_matrix (Context       *context,
// 			   const PangoMatrix  *matrix)
//  {
//    g_return_if_fail (PANGO_IS_CONTEXT (context));

//    if (context.matrix || matrix)
// 	 contextChanged (context);

//    if (context.matrix)
// 	 pango_matrix_free (context.matrix);
//    if (matrix)
// 	 context.matrix = pango_matrix_copy (matrix);
//    else
// 	 context.matrix = nil;

//    update_resolved_gravity (context);
//  }

//  /**
//   * pango_context_get_matrix:
//   * `context`: a #Context
//   *
//   * Gets the transformation matrix that will be applied when
//   * rendering with this context. See pango_context_set_matrix().
//   *
//   * Return value: (nullable): the matrix, or %nil if no matrix has
//   *  been set (which is the same as the identity matrix). The returned
//   *  matrix is owned by Pango and must not be modified or freed.
//   *
//   * Since: 1.6
//   **/
//  const PangoMatrix *
//  pango_context_get_matrix (context *Context)
//  {
//    g_return_val_if_fail (PANGO_IS_CONTEXT (context), nil);

//    return context.matrix;
//  }

//  /**
//   * pango_context_get_font_map:
//   * `context`: a #Context
//   *
//   * Gets the #PangoFontMap used to look up fonts for this context.
//   *
//   * Return value: (transfer none): the font map for the #Context.
//   *               This value is owned by Pango and should not be unreferenced.
//   *
//   * Since: 1.6
//   **/
//  PangoFontMap *
//  pango_context_get_font_map (context *Context)
//  {
//    g_return_val_if_fail (PANGO_IS_CONTEXT (context), nil);

//    return context.fontMap;
//  }

//  /**
//   * pango_context_list_families:
//   * `context`: a #Context
//   * @families: (out) (array length=n_families) (transfer container): location to store a pointer to
//   *            an array of #PangoFontFamily *. This array should be freed
//   *            with g_free().
//   * @n_families: (out): location to store the number of elements in @descs
//   *
//   * List all families for a context.
//   **/
//  void
//  pango_context_list_families (Context          *context,
// 				  PangoFontFamily     ***families,
// 				  int                   *n_families)
//  {
//    g_return_if_fail (context != nil);
//    g_return_if_fail (families == nil || n_families != nil);

//    if (n_families == nil)
// 	 return;

//    if (context.fontMap == nil)
// 	 {
// 	   *n_families = 0;
// 	   if (families)
// 	 *families = nil;

// 	   return;
// 	 }
//    else
// 	 pango_font_map_list_families (context.fontMap, families, n_families);
//  }

//  /**
//   * pango_context_load_Fontset:
//   * `context`: a #Context
//   * @desc: a #PangoFontDescription describing the fonts to load
//   * @language: a #PangoLanguage the fonts will be used for
//   *
//   * Load a set of fonts in the context that can be used to render
//   * a font matching @desc.
//   *
//   * Returns: (transfer full) (nullable): the newly allocated
//   *          #PangoFontset loaded, or %nil if no font matched.
//   **/
//  PangoFontset *
//  pango_context_load_Fontset (Context               *context,
// 				 const PangoFontDescription *desc,
// 				 PangoLanguage             *language)
//  {
//    g_return_val_if_fail (context != nil, nil);

//    return pango_font_map_load_Fontset (context.fontMap, context, desc, language);
//  }

//  /**
//   * pango_context_get_font_description:
//   * `context`: a #Context
//   *
//   * Retrieve the default font description for the context.
//   *
//   * Return value: (transfer none): a pointer to the context's default font
//   *               description. This value must not be modified or freed.
//   **/
//  PangoFontDescription *
//  pango_context_get_font_description (context *Context)
//  {
//    g_return_val_if_fail (context != nil, nil);

//    return context.font_desc;
//  }

//  /**
//   * pango_context_set_base_dir:
//   * `context`: a #Context
//   * @direction: the new base direction
//   *
//   * Sets the base direction for the context.
//   *
//   * The base direction is used in applying the Unicode bidirectional
//   * algorithm; if the @direction is %PANGO_DIRECTION_LTR or
//   * %PANGO_DIRECTION_RTL, then the value will be used as the paragraph
//   * direction in the Unicode bidirectional algorithm.  A value of
//   * %PANGO_DIRECTION_WEAK_LTR or %PANGO_DIRECTION_WEAK_RTL is used only
//   * for paragraphs that do not contain any strong characters themselves.
//   **/
//  void
//  pango_context_set_base_dir (Context  *context,
// 				 PangoDirection direction)
//  {
//    g_return_if_fail (context != nil);

//    if (direction != context.base_dir)
// 	 contextChanged (context);

//    context.base_dir = direction;
//  }

//  /**
//   * pango_context_get_base_dir:
//   * `context`: a #Context
//   *
//   * Retrieves the base direction for the context. See
//   * pango_context_set_base_dir().
//   *
//   * Return value: the base direction for the context.
//   **/
//  PangoDirection
//  pango_context_get_base_dir (context *Context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_DIRECTION_LTR);

//    return context.base_dir;
//  }

//  /**
//   * pango_context_set_base_gravity:
//   * `context`: a #Context
//   * @gravity: the new base gravity
//   *
//   * Sets the base gravity for the context.
//   *
//   * The base gravity is used in laying vertical text out.
//   *
//   * Since: 1.16
//   **/
//  void
//  pango_context_set_base_gravity (Context  *context,
// 				 PangoGravity gravity)
//  {
//    g_return_if_fail (context != nil);

//    if (gravity != context.base_gravity)
// 	 contextChanged (context);

//    context.base_gravity = gravity;

//    update_resolved_gravity (context);
//  }

//  /**
//   * pango_context_get_base_gravity:
//   * `context`: a #Context
//   *
//   * Retrieves the base gravity for the context. See
//   * pango_context_set_base_gravity().
//   *
//   * Return value: the base gravity for the context.
//   *
//   * Since: 1.16
//   **/
//  PangoGravity
//  pango_context_get_base_gravity (context *Context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_GRAVITY_SOUTH);

//    return context.base_gravity;
//  }

//  /**
//   * pango_context_get_gravity:
//   * `context`: a #Context
//   *
//   * Retrieves the gravity for the context. This is similar to
//   * pango_context_get_base_gravity(), except for when the base gravity
//   * is %PANGO_GRAVITY_AUTO for which pango_gravity_get_for_matrix() is used
//   * to return the gravity from the current context matrix.
//   *
//   * Return value: the resolved gravity for the context.
//   *
//   * Since: 1.16
//   **/
//  PangoGravity
//  pango_context_get_gravity (context *Context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_GRAVITY_SOUTH);

//    return context.resolved_gravity;
//  }

//  /**
//   * pango_context_set_gravity_hint:
//   * `context`: a #Context
//   * @hint: the new gravity hint
//   *
//   * Sets the gravity hint for the context.
//   *
//   * The gravity hint is used in laying vertical text out, and is only relevant
//   * if gravity of the context as returned by pango_context_get_gravity()
//   * is set %PANGO_GRAVITY_EAST or %PANGO_GRAVITY_WEST.
//   *
//   * Since: 1.16
//   **/
//  void
//  pango_context_set_gravity_hint (Context    *context,
// 				 PangoGravityHint hint)
//  {
//    g_return_if_fail (context != nil);

//    if (hint != context.gravity_hint)
// 	 contextChanged (context);

//    context.gravity_hint = hint;
//  }

//  /**
//   * pango_context_get_gravity_hint:
//   * `context`: a #Context
//   *
//   * Retrieves the gravity hint for the context. See
//   * pango_context_set_gravity_hint() for details.
//   *
//   * Return value: the gravity hint for the context.
//   *
//   * Since: 1.16
//   **/
//  PangoGravityHint
//  pango_context_get_gravity_hint (context *Context)
//  {
//    g_return_val_if_fail (context != nil, PANGO_GRAVITY_HINT_NATURAL);

//    return context.gravity_hint;
//  }

//  /**********************************************************************/

func (iterator *AttrIterator) advance_attr_iterator_to(start_index int) bool {
	start_range, end_range := iterator.StartIndex, iterator.EndIndex

	for start_index >= end_range {
		if !iterator.pango_attr_iterator_next() {
			return false
		}
		start_range, end_range = iterator.StartIndex, iterator.EndIndex
	}

	if start_range > start_index {
		log.Println("In pango_itemize(), the cached iterator passed in " +
			"had already moved beyond the start_index")
	}

	return true
}

/***************************************************************************
 * We cache the results of character,Fontset => font in a hash table
 ***************************************************************************/

// we could maybe use a sync.Map ?
type FontCache struct {
	store map[rune]Font
	lock  sync.RWMutex
}

// NewFontCache initialize a new font cache.
func NewFontCache() *FontCache {
	return &FontCache{store: make(map[rune]Font)}
}

func (cache *FontCache) font_cache_get(wc rune) (Font, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	f, b := cache.store[wc]
	return f, b
}

func (cache *FontCache) font_cache_insert(wc rune, font Font) {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	cache.store[wc] = font
}

//  /**********************************************************************/

type ChangedFlags uint8

const (
	EMBEDDING_CHANGED ChangedFlags = 1 << iota
	SCRIPT_CHANGED
	LANG_CHANGED
	FONT_CHANGED
	DERIVED_LANG_CHANGED
	WIDTH_CHANGED
	EMOJI_CHANGED
)

type WidthIter struct {
	text       []rune
	start, end int // index into text
	upright    bool
}

func (iter *WidthIter) reset(text []rune) {
	iter.text = text
	iter.start, iter.end = 0, 0
	iter.next()
}

/* https://www.unicode.org/Public/11.0.0/ucd/VerticalOrientation.txt
* VO=U or Tu table generated by tools/gen-vertical-orientation-U-table.py.
*
* FIXME: In the future, If GLib supports VerticalOrientation, please use it.
 */
var upright = [...][2]rune{
	{0x00A7, 0x00A7},
	{0x00A9, 0x00A9},
	{0x00AE, 0x00AE},
	{0x00B1, 0x00B1},
	{0x00BC, 0x00BE},
	{0x00D7, 0x00D7},
	{0x00F7, 0x00F7},
	{0x02EA, 0x02EB},
	{0x1100, 0x11FF},
	{0x1401, 0x167F},
	{0x18B0, 0x18FF},
	{0x2016, 0x2016},
	{0x2020, 0x2021},
	{0x2030, 0x2031},
	{0x203B, 0x203C},
	{0x2042, 0x2042},
	{0x2047, 0x2049},
	{0x2051, 0x2051},
	{0x2065, 0x2065},
	{0x20DD, 0x20E0},
	{0x20E2, 0x20E4},
	{0x2100, 0x2101},
	{0x2103, 0x2109},
	{0x210F, 0x210F},
	{0x2113, 0x2114},
	{0x2116, 0x2117},
	{0x211E, 0x2123},
	{0x2125, 0x2125},
	{0x2127, 0x2127},
	{0x2129, 0x2129},
	{0x212E, 0x212E},
	{0x2135, 0x213F},
	{0x2145, 0x214A},
	{0x214C, 0x214D},
	{0x214F, 0x2189},
	{0x218C, 0x218F},
	{0x221E, 0x221E},
	{0x2234, 0x2235},
	{0x2300, 0x2307},
	{0x230C, 0x231F},
	{0x2324, 0x2328},
	{0x232B, 0x232B},
	{0x237D, 0x239A},
	{0x23BE, 0x23CD},
	{0x23CF, 0x23CF},
	{0x23D1, 0x23DB},
	{0x23E2, 0x2422},
	{0x2424, 0x24FF},
	{0x25A0, 0x2619},
	{0x2620, 0x2767},
	{0x2776, 0x2793},
	{0x2B12, 0x2B2F},
	{0x2B50, 0x2B59},
	{0x2BB8, 0x2BD1},
	{0x2BD3, 0x2BEB},
	{0x2BF0, 0x2BFF},
	{0x2E80, 0x3007},
	{0x3012, 0x3013},
	{0x3020, 0x302F},
	{0x3031, 0x309F},
	{0x30A1, 0x30FB},
	{0x30FD, 0xA4CF},
	{0xA960, 0xA97F},
	{0xAC00, 0xD7FF},
	{0xE000, 0xFAFF},
	{0xFE10, 0xFE1F},
	{0xFE30, 0xFE48},
	{0xFE50, 0xFE57},
	{0xFE5F, 0xFE62},
	{0xFE67, 0xFE6F},
	{0xFF01, 0xFF07},
	{0xFF0A, 0xFF0C},
	{0xFF0E, 0xFF19},
	{0xFF1F, 0xFF3A},
	{0xFF3C, 0xFF3C},
	{0xFF3E, 0xFF3E},
	{0xFF40, 0xFF5A},
	{0xFFE0, 0xFFE2},
	{0xFFE4, 0xFFE7},
	{0xFFF0, 0xFFF8},
	{0xFFFC, 0xFFFD},
	{0x10980, 0x1099F},
	{0x11580, 0x115FF},
	{0x11A00, 0x11AAF},
	{0x13000, 0x1342F},
	{0x14400, 0x1467F},
	{0x16FE0, 0x18AFF},
	{0x1B000, 0x1B12F},
	{0x1B170, 0x1B2FF},
	{0x1D000, 0x1D1FF},
	{0x1D2E0, 0x1D37F},
	{0x1D800, 0x1DAAF},
	{0x1F000, 0x1F7FF},
	{0x1F900, 0x1FA6F},
	{0x20000, 0x2FFFD},
	{0x30000, 0x3FFFD},
	{0xF0000, 0xFFFFD},
	{0x100000, 0x10FFFD},
}

func isUpright(ch rune) bool {
	const max = len(upright)
	st := 0
	ed := max

	for st <= ed {
		mid := (st + ed) / 2
		if upright[mid][0] <= ch && ch <= upright[mid][1] {
			return true
		} else {
			if upright[mid][0] <= ch {
				st = mid + 1
			} else {
				ed = mid - 1
			}
		}
	}

	return false
}

func (iter *WidthIter) next() {
	metJoiner := false
	iter.start = iter.end

	if iter.end < len(iter.text) {
		ch := iter.text[iter.end]
		iter.end++
		iter.upright = isUpright(ch)
	}

	for iter.end < len(iter.text) {
		ch := iter.text[iter.end]
		iter.end++

		/* for zero width joiner */
		if ch == 0x200D {
			iter.end++
			metJoiner = true
			continue
		}

		/* ignore the upright check if met joiner */
		if metJoiner {
			iter.end++
			metJoiner = false
			continue
		}

		/* for variation selector, tag and emoji modifier. */
		if ch == 0xFE0E || ch == 0xFE0F || (ch >= 0xE0020 && ch <= 0xE007F) || (ch >= 0x1F3FB && ch <= 0x1F3FF) {
			iter.end++
			continue
		}

		if isUpright(ch) != iter.upright {
			break
		}
		iter.end++
	}
}

type ItemizeState struct {
	context *Context
	text    []rune
	end     int // index into text

	run_start, run_end int // index in text

	result []*Item
	item   *Item

	embedding_levels     []fribidi.Level
	embedding_end_offset int
	embedding_end        int
	embedding            fribidi.Level

	gravity           Gravity
	gravity_hint      GravityHint
	resolved_gravity  Gravity
	font_desc_gravity Gravity
	centered_baseline bool

	attr_iter *AttrIterator
	// free_attr_iter bool
	attr_end         int
	font_desc        *FontDescription
	emoji_font_desc  *FontDescription
	lang             Language
	extra_attrs      AttrList
	copy_extra_attrs bool

	changed ChangedFlags

	script_iter ScriptIter
	script_end  int    // copied from `script_iter`
	script      Script // copied from `script_iter`

	width_iter WidthIter
	emoji_iter EmojiIter

	derived_lang Language

	current_fonts   Fontset
	cache           *FontCache
	base_font       Font
	enable_fallback bool
}

func (state *ItemizeState) update_embedding_end() {
	state.embedding = state.embedding_levels[state.embedding_end_offset]
	for state.embedding_end < state.end &&
		state.embedding_levels[state.embedding_end_offset] == state.embedding {
		state.embedding_end_offset++
		state.embedding_end++
	}

	state.changed |= EMBEDDING_CHANGED
}

func (attr_list AttrList) find_attribute(type_ AttrType) *Attribute {
	for _, attr := range attr_list {
		if attr.Type == type_ {
			return attr
		}
	}
	return nil
}

func (state *ItemizeState) update_attr_iterator() {
	//    PangoLanguage *old_lang;
	//    PangoAttribute *attr;
	//    int end_index;
	end_index := state.attr_iter.EndIndex // pango_attr_iterator_range (state.attr_iter, nil, &end_index);
	if end_index < state.end {
		state.attr_end = end_index
	} else {
		state.attr_end = state.end
	}

	if state.emoji_font_desc != nil {
		state.emoji_font_desc = nil
	}

	old_lang := state.lang

	cp := state.context.font_desc // copy
	state.font_desc = &cp
	state.attr_iter.pango_attr_iterator_get_font(state.font_desc, &state.lang, &state.extra_attrs)
	if state.font_desc.mask&F_GRAVITY != 0 {
		state.font_desc_gravity = state.font_desc.Gravity
	} else {
		state.font_desc_gravity = GRAVITY_AUTO
	}

	state.copy_extra_attrs = false

	if state.lang == "" {
		state.lang = state.context.language
	}

	attr := state.extra_attrs.find_attribute(ATTR_FALLBACK)
	state.enable_fallback = (attr == nil || attr.Data.(AttrInt) != 0)

	attr = state.extra_attrs.find_attribute(ATTR_GRAVITY)
	state.gravity = GRAVITY_AUTO
	if attr != nil {
		state.gravity = Gravity(attr.Data.(AttrInt))
	}

	attr = state.extra_attrs.find_attribute(ATTR_GRAVITY_HINT)
	state.gravity_hint = state.context.gravity_hint
	if attr != nil {
		state.gravity_hint = GravityHint(attr.Data.(AttrInt))
	}

	state.changed |= FONT_CHANGED
	if state.lang != old_lang {
		state.changed |= LANG_CHANGED
	}
}

func (state *ItemizeState) updateEnd() {
	state.run_end = state.embedding_end
	if i := state.attr_end; i < state.run_end {
		state.run_end = i
	}
	if i := state.script_end; i < state.run_end {
		state.run_end = i
	}
	if i := state.width_iter.end; i < state.run_end {
		state.run_end = i
	}
	if i := state.emoji_iter.end; i < state.run_end {
		state.run_end = i
	}
}

func (state *ItemizeState) updateForNewRun() {
	// This block should be moved to update_attr_iterator, but I'm too lazy to do it right now
	if state.changed&(FONT_CHANGED|SCRIPT_CHANGED|WIDTH_CHANGED) != 0 {
		/* Font-desc gravity overrides everything */
		if state.font_desc_gravity != GRAVITY_AUTO {
			state.resolved_gravity = state.font_desc_gravity
		} else {
			gravity := state.gravity
			gravity_hint := state.gravity_hint

			if gravity == GRAVITY_AUTO {
				gravity = state.context.resolved_gravity
			}

			state.resolved_gravity = pango_gravity_get_for_script_and_width(state.script,
				state.width_iter.upright, gravity, gravity_hint)
		}

		if state.font_desc_gravity != state.resolved_gravity {
			state.font_desc.SetGravity(state.resolved_gravity)
			state.changed |= FONT_CHANGED
		}
	}

	if state.changed&(SCRIPT_CHANGED|LANG_CHANGED) != 0 {
		old_derived_lang := state.derived_lang
		state.derived_lang = compute_derived_language(state.lang, state.script)
		if old_derived_lang != state.derived_lang {
			state.changed |= DERIVED_LANG_CHANGED
		}
	}

	if state.changed&(EMOJI_CHANGED) != 0 {
		state.changed |= FONT_CHANGED
	}

	if state.changed&(FONT_CHANGED|DERIVED_LANG_CHANGED) != 0 && state.current_fonts != nil {
		state.current_fonts = nil
		state.cache = nil
	}

	if state.current_fonts == nil {
		is_emoji := state.emoji_iter.isEmoji
		if is_emoji && state.emoji_font_desc == nil {
			cp := *state.font_desc // copy
			state.emoji_font_desc = &cp
			state.emoji_font_desc.SetFamily("emoji")
		}
		fontDescArg := state.font_desc
		if is_emoji {
			fontDescArg = state.emoji_font_desc
		}
		state.current_fonts = state.context.fontMap.LoadFontset(
			state.context, fontDescArg, state.derived_lang)
		state.cache = getFontCache(state.current_fonts)
	}

	if (state.changed&FONT_CHANGED) != 0 && state.base_font != nil {
		state.base_font = nil
	}
}

func (state *ItemizeState) processRun() {
	lastWasForcedBreak := false

	state.updateForNewRun()

	if debugMode { //  We should never get an empty run
		assert(state.run_end > state.run_start, fmt.Sprintf("processRun: %d <= %d", state.run_end, state.run_start))
	}

	for pos, wc := range state.text[state.run_start:state.run_end] {
		isForcedBreak := (wc == '\t' || wc == LINE_SEPARATOR)
		var font Font

		// We don't want space characters to affect font selection; in general,
		// it's always wrong to select a font just to render a space.
		// We assume that all fonts have the ASCII space, and for other space
		// characters if they don't, HarfBuzz will compatibility-decompose them
		// to ASCII space...
		// See bugs #355987 and #701652.
		//
		// We don't want to change fonts just for variation selectors.
		// See bug #781123.
		//
		// Finally, don't change fonts for line or paragraph separators.
		isIn := unicode.In(wc, unicode.Cc, unicode.Cf, unicode.Cs, unicode.Zl, unicode.Zp)
		if isIn || (unicode.Is(unicode.Zs, wc) && wc != '\u1680' /* OGHAM SPACE MARK */) ||
			(wc >= '\ufe00' && wc <= '\ufe0f') || (wc >= '\U000e0100' && wc <= '\U000e01ef') {
			font = nil
		} else {
			font, _ = state.get_font(wc)
		}

		state.addCharacter(font, isForcedBreak || lastWasForcedBreak, pos)

		lastWasForcedBreak = isForcedBreak
	}

	/* Finish the final item from the current segment */
	state.item.length = state.run_end - state.item.offset
	if state.item.Analysis.Font == nil {
		font, ok := state.get_font(' ')
		if !ok {
			// only warn once per fontmap/script pair
			if shouldWarn(state.context.fontMap, state.script) {
				log.Printf("failed to choose a font for script %s: expect ugly output", state.script)
			}
		}
		state.fillFont(font)
	}
	state.item = nil
}

type getFontInfo struct {
	font Font
	lang Language
	wc   rune
}

func (info *getFontInfo) get_font_foreach(Fontset Fontset, font Font) bool {
	if font == nil {
		return false
	}

	if pango_font_has_char(font, info.wc) {
		info.font = font
		return true
	}

	if Fontset == nil {
		info.font = font
		return true
	}

	return false
}

func (state *ItemizeState) get_font(wc rune) (Font, bool) {
	// We'd need a separate cache when fallback is disabled, but since lookup
	// with fallback disabled is faster anyways, we just skip caching.
	if state.enable_fallback {
		if font, ok := state.cache.font_cache_get(wc); ok {
			return font, true
		}
	}

	info := getFontInfo{lang: state.derived_lang, wc: wc}

	if state.enable_fallback {
		state.current_fonts.Foreach(func(font Font) bool {
			return info.get_font_foreach(state.current_fonts, font)
		})
	} else {
		info.get_font_foreach(nil, state.get_base_font())
	}

	/* skip caching if fallback disabled (see above) */
	if state.enable_fallback {
		state.cache.font_cache_insert(wc, info.font)
	}
	return info.font, true
}

//  }

func (context *Context) newItemizeState(text []rune, baseDir Direction,
	startIndex, length int,
	attrs AttrList, cachedIter *AttrIterator, desc *FontDescription) *ItemizeState {
	var state ItemizeState
	state.context = context
	state.text = text
	state.end = startIndex + length

	state.run_start = startIndex
	state.changed = EMBEDDING_CHANGED | SCRIPT_CHANGED | LANG_CHANGED |
		FONT_CHANGED | WIDTH_CHANGED | EMOJI_CHANGED

	// First, apply the bidirectional algorithm to break the text into directional runs.
	baseDir, state.embedding_levels = pango_log2vis_get_embedding_levels(text[startIndex:startIndex+length], baseDir)

	state.embedding_end_offset = 0
	state.embedding_end = startIndex
	state.update_embedding_end()

	// Initialize the attribute iterator
	if cachedIter != nil {
		state.attr_iter = cachedIter
	} else if len(attrs) != 0 {
		state.attr_iter = attrs.pango_attr_list_get_iterator()
	}

	if state.attr_iter != nil {
		state.attr_iter.advance_attr_iterator_to(startIndex)
		state.update_attr_iterator()
	} else {
		if desc == nil {
			cp := state.context.font_desc
			state.font_desc = &cp
		} else {
			state.font_desc = desc
		}
		state.lang = state.context.language
		state.extra_attrs = nil
		state.copy_extra_attrs = false

		state.attr_end = state.end
		state.enable_fallback = true
	}

	// Initialize the script iterator
	state.script_iter._pango_script_iter_init(text[startIndex:])
	state.script_end, state.script = state.script_iter.script_end, state.script_iter.script_code

	state.width_iter.reset(text[startIndex:])
	state.emoji_iter.reset(text[startIndex:])

	if state.emoji_iter.isEmoji {
		state.width_iter.end = max(state.width_iter.end, state.emoji_iter.end)
	}

	state.updateEnd()

	if state.font_desc.mask&F_GRAVITY != 0 {
		state.font_desc_gravity = state.font_desc.Gravity
	} else {
		state.font_desc_gravity = GRAVITY_AUTO
	}

	state.gravity = GRAVITY_AUTO
	state.centered_baseline = state.context.resolved_gravity.IsVertical()
	state.gravity_hint = state.context.gravity_hint
	state.resolved_gravity = GRAVITY_AUTO

	return &state
}

func (state *ItemizeState) next() bool {
	if state.run_end == state.end {
		return false
	}

	state.changed = 0

	state.run_start = state.run_end

	if state.run_end == state.embedding_end {
		state.update_embedding_end()
	}

	if state.run_end == state.attr_end {
		state.attr_iter.pango_attr_iterator_next()
		state.update_attr_iterator()
	}

	if state.run_end == state.script_end {
		state.script_iter.pango_script_iter_next()
		state.script_end, state.script = state.script_iter.script_end, state.script_iter.script_code
		state.changed |= SCRIPT_CHANGED
	}
	if state.run_end == state.emoji_iter.end {
		state.emoji_iter.next()
		state.changed |= EMOJI_CHANGED

		if state.emoji_iter.isEmoji {
			state.width_iter.end = max(state.width_iter.end, state.emoji_iter.end)
		}
	}
	if state.run_end == state.width_iter.end {
		state.width_iter.next()
		state.changed |= WIDTH_CHANGED
	}

	state.updateEnd()

	return true
}

func (state *ItemizeState) fillFont(font Font) {
	for _, item := range state.result {
		if item.Analysis.Font != nil {
			break
		}
		if font != nil {
			item.Analysis.Font = font
		}
	}
}

// pos is the index into text
func (state *ItemizeState) addCharacter(font Font, force_break bool, pos int) {
	if item := state.item; item != nil {
		if item.Analysis.Font == nil && font != nil {
			state.fillFont(font)
		} else if item.Analysis.Font != nil && font == nil {
			font = item.Analysis.Font
		}

		if !force_break && item.Analysis.Font == font {
			item.length++
			return
		}

		item.length = pos - item.offset
	}

	state.item = &Item{}
	state.item.offset = pos
	state.item.length = 1

	state.item.Analysis.Font = font

	state.item.Analysis.level = state.embedding
	state.item.Analysis.gravity = state.resolved_gravity

	/* The level vs. gravity dance:
	*	- If gravity is SOUTH, leave level untouched.
	*	- If gravity is NORTH, step level one up, to
	*	  not get mirrored upside-down text.
	*	- If gravity is EAST, step up to an even level, as
	*	  it's a clockwise-rotated layout, so the rotated
	*	  top is unrotated left.
	*	- If gravity is WEST, step up to an odd level, as
	*	  it's a counter-clockwise-rotated layout, so the rotated
	*	  top is unrotated right.
	*
	* A similar dance is performed in pango-layout.c:
	* line_set_resolved_dir().  Keep in synch.
	 */
	switch state.item.Analysis.gravity {
	case GRAVITY_NORTH:
		state.item.Analysis.level++
	case GRAVITY_EAST:
		state.item.Analysis.level += 1
		state.item.Analysis.level &= ^1
	case GRAVITY_WEST:
		state.item.Analysis.level |= 1
	}

	if state.centered_baseline {
		state.item.Analysis.flags = PANGO_ANALYSIS_FLAG_CENTERED_BASELINE
	} else {
		state.item.Analysis.flags = 0
	}

	state.item.Analysis.script = state.script
	state.item.Analysis.language = state.derived_lang

	if state.copy_extra_attrs {
		state.item.Analysis.extra_attrs = state.extra_attrs.pango_attr_list_copy()
	} else {
		state.item.Analysis.extra_attrs = state.extra_attrs
		state.copy_extra_attrs = true
	}

	// prepend
	state.result = append(state.result, nil)
	copy(state.result, state.result[1:])
	state.result[0] = state.item
}

func (state *ItemizeState) get_base_font() Font {
	if state.base_font == nil {
		state.base_font = LoadFont(state.context.fontMap, state.context, state.font_desc)
	}
	return state.base_font
}

func (state *ItemizeState) itemize_state_finish() {} // only memory cleanup

func (context *Context) itemize_with_font(text []rune, desc *FontDescription) []*Item {
	if len(text) == 0 {
		return nil
	}

	state := context.newItemizeState(text, context.base_dir, 0, len(text), nil, nil, desc)

	for do := true; do; do = state.next() {
		state.processRun()
	}

	state.itemize_state_finish()
	reverseItems(state.result)
	return state.result
}

func getBaseMetrics(fs Fontset) FontMetrics {
	var metrics FontMetrics

	language := fs.GetLanguage()

	// Initialize the metrics from the first font in the Fontset
	getFirstMetricsForeach := func(font Font) bool {
		metrics = FontGetMetrics(font, language)
		return true // Stops iteration
	}
	fs.Foreach(getFirstMetricsForeach)

	return metrics
}

//  static void
//  update_metrics_from_items (PangoFontMetrics *metrics,
// 				PangoLanguage    *language,
// 				const char       *text,
// 				unsigned int      text_len,
// 				GList            *items)

//  {
//    GHashTable *fonts_seen = g_hash_table_new (nil, nil);
//    PangoGlyphString *glyphs = pango_glyph_string_new ();
//    GList *l;
//    glong text_width;

//    /* This should typically be called with a sample text string. */
//    g_return_if_fail (text_len > 0);

//    metrics.approximate_char_width = 0;

//    for (l = items; l; l = l.next)
// 	 {
// 	   PangoItem *item = l.data;
// 	   PangoFont *font = item.analysis.font;

// 	   if (font != nil && g_hash_table_lookup (fonts_seen, font) == nil)
// 	 {
// 	   PangoFontMetrics *raw_metrics = FontGetMetrics (font, language);
// 	   g_hash_table_insert (fonts_seen, font, font);

// 	   /* metrics will already be initialized from the first font in the Fontset */
// 	   metrics.ascent = MAX (metrics.ascent, raw_metrics.ascent);
// 	   metrics.descent = MAX (metrics.descent, raw_metrics.descent);
// 	   metrics.height = MAX (metrics.height, raw_metrics.height);
// 	   pango_font_metrics_unref (raw_metrics);
// 	 }

// 	   pango_shape_full (text + item.offset, item.length,
// 			 text, text_len,
// 			 &item.analysis, glyphs);
// 	   metrics.approximate_char_width += pango_glyph_string_get_width (glyphs);
// 	 }

//    pango_glyph_string_free (glyphs);
//    g_hash_table_destroy (fonts_seen);

//    text_width = pango_utf8_strwidth (text);
//    g_assert (text_width > 0);
//    metrics.approximate_char_width /= text_width;
//  }

//  static void
//  check_fontmap_changed (context *Context)
//  {
//    guint old_serial = context.fontmapSerial;

//    if (!context.fontMap)
// 	 return;

//    context.fontmapSerial = pango_font_map_get_serial (context.fontMap);

//    if (old_serial != context.fontmapSerial)
// 	 contextChanged (context);
//  }

// Returns the current serial number of `context`.  The serial number is
// initialized to an small number larger than zero when a new context
// is created and is increased whenever the context is changed using any
// of the setter functions, or the #PangoFontMap it uses to find fonts has
// changed. The serial may wrap, but will never have the value 0. Since it
// can wrap, never compare it with "less than", always use "not equals".
//
// This can be used to automatically detect changes to a #Context, and
// is only useful when implementing objects that need update when their
// #Context changes, like Layout.
func (context *Context) pango_context_get_serial() uint {
	context.check_fontmap_changed()
	return context.serial
}

func (context *Context) check_fontmap_changed() {} // TODO:

//  /**
//  // pango_context_set_round_glyph_positions:
//   * `context`: a #Context
//   * @round_positions: whether to round glyph positions
//   *
//   * Sets whether font rendering with this context should
//   * round glyph positions and widths to integral positions,
//   * in device units.
//   *
//   * This is useful when the renderer can't handle subpixel
//   * positioning of glyphs.
//   *
//   * The default value is to round glyph positions, to remain
//   * compatible with previous Pango behavior.
//   *
//   * Since: 1.44
//   */
//  void
//  pango_context_set_round_glyph_positions (context *Context,
// 										  bool      round_positions)
//  {
//    if (context.round_glyph_positions != round_positions)
// 	 {
// 	   context.round_glyph_positions = round_positions;
// 	   contextChanged (context);
// 	 }
//  }

//  /**
//   * pango_context_get_round_glyph_positions:
//   * `context`: a #Context
//   *
//   * Returns whether font rendering with this context should
//   * round glyph positions and widths.
//   *
//   * Since: 1.44
//   */
//  bool
//  pango_context_get_round_glyph_positions (context *Context)
//  {
//    return context.round_glyph_positions;
//  }
