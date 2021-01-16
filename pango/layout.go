package pango

import (
	"log"
	"unicode/utf8"
)

/**
 * While complete access to the layout capabilities of Pango is provided
 * using the detailed interfaces for itemization and shaping, using
 * that functionality directly involves writing a fairly large amount
 * of code. The objects and functions in this section provide a
 * high-level driver for formatting entire paragraphs of text
 * at once. This includes paragraph-level functionality such as
 * line-breaking, justification, alignment and ellipsization.
 */

// Only one character has type G_UNICODE_LINE_SEPARATOR in Unicode 5.0: update this if that changes.
const LINE_SEPARATOR = 0x2028

// WrapMode describes how to wrap the lines of a `Layout` to the desired width.
type WrapMode uint8

const (
	PANGO_WRAP_WORD      WrapMode = iota // wrap lines at word boundaries.
	PANGO_WRAP_CHAR                      // wrap lines at character boundaries.
	PANGO_WRAP_WORD_CHAR                 // wrap lines at word boundaries, but fall back to character boundaries if there is not enough space for a full word.
)

// Layout represents an entire paragraph
// of text. It is initialized with a #PangoContext, UTF-8 string
// and set of attributes for that string. Once that is done, the
// set of formatted lines can be extracted from the object,
// the layout can be rendered, and conversion between logical
// character positions within the layout's text, and the physical
// position of the resulting glyphs can be made.
//
// There are also a number of parameters to adjust the formatting
// of a #Layout, which are illustrated in <xref linkend="parameters"/>.
// It is possible, as well, to ignore the 2-D setup, and simply
// treat the results of a #Layout as a list of lines.
//
// <figure id="parameters">
// <title>Adjustable parameters (on the left) and font metrics (on the right) for a Layout</title>
// <graphic fileref="layout.png" format="PNG"></graphic>
// </figure>
type Layout struct {
	//   GObject parent_instance;

	/* If you add fields to Layout be sure to update _copy()
	 * unless you add a value between copy_begin and copy_end.
	 */

	/* Referenced items */
	context   *Context
	attrs     AttrList
	font_desc *FontDescription
	tabs      *TabArray

	/* Dupped */
	text []rune
	// n_chars = len(text) : number of characters in layout */
	// length int /* length of text in bytes */

	serial         uint
	context_serial uint

	width        GlyphUnit /* wrap/ellipsize width, in device units, or -1 if not set */
	height       int       /* ellipsize width, in device units if positive, number of lines if negative */
	indent       GlyphUnit /* amount by which first line should be shorter */
	spacing      int       /* spacing between lines */
	line_spacing float32   /* factor to apply to line height */

	justify              bool          // = 1;
	alignment            Alignment     // = 2;
	single_paragraph     bool          // = 1;
	auto_dir             bool          // = 1;
	wrap                 WrapMode      // = 2;
	is_wrapped           bool          // = 1;		/* Whether the layout has any wrapped lines */
	ellipsize            EllipsizeMode // = 2;		/* PangoEllipsizeMode */
	is_ellipsized        bool          // = 1;	/* Whether the layout has any ellipsized lines */
	unknown_glyphs_count int           /* number of unknown glyphs */

	//   /* some caching */
	logical_rect_cached bool // = true
	ink_rect_cached     bool // = true
	//   Rectangle logical_rect;
	//   Rectangle ink_rect;
	tabWidth GlyphUnit /* Cached width of a tab. -1 == not yet calculated */

	// logical attributes for layout's text, allocated
	// once in check_lines; has length len(text)+1
	log_attrs []CharAttr

	lines []*LayoutLine
	//   guint line_count;		/* Number of lines in @lines. 0 if lines is %nil */
}

// SetText sets the text of the layout, validating `text` and rendering invalid UTF-8
// with a placeholder glyph.
//
// Note that if you have used pango_layout_set_markup() or
// pango_layout_set_markup_with_accel() on @layout before, you may
// want to call pango_layout_set_attributes() to clear the attributes
// set on the layout from the markup as this function does not clear
// attributes.
func (layout *Layout) SetText(text string) {
	layout.text = layout.text[:0]
	b := []byte(text)
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		b = b[size:]
		layout.text = append(layout.text, r)
	}
	layout.layout_changed()
}

func (layout *Layout) check_context_changed() {
	old_serial := layout.context_serial

	layout.context_serial = layout.context.pango_context_get_serial()

	if old_serial != layout.context_serial {
		layout.pango_layout_context_changed()
	}
}

// Forces recomputation of any state in the `Layout` that
// might depend on the layout's context. This function should
// be called if you make changes to the context subsequent
// to creating the layout.
func (layout *Layout) pango_layout_context_changed() {
	if layout == nil {
		return
	}

	layout.layout_changed()
	layout.tabWidth = -1
}

func (layout *Layout) layout_changed() {
	layout.serial++
	if layout.serial == 0 {
		layout.serial++
	}
	layout.pango_layout_clear_lines()
}

func (layout *Layout) pango_layout_clear_lines() {
	// TODO: we could keep the underlying arrays
	layout.lines = nil
	layout.log_attrs = nil
	layout.unknown_glyphs_count = -1
	layout.logical_rect_cached = false
	layout.ink_rect_cached = false
	layout.is_ellipsized = false
	layout.is_wrapped = false
}

// Sets the text attributes for a layout object.
func (layout *Layout) pango_layout_set_attributes(attrs AttrList) {
	if layout == nil {
		return
	}

	if layout.attrs.pango_attr_list_equal(attrs) {
		return
	}

	// We always clear lines such that this function can be called
	// whenever attrs changes.
	layout.attrs = attrs
	layout.layout_changed()
	layout.tabWidth = -1
}

// Sets the tabs to use for `layout`, overriding the default tabs
// (by default, tabs are every 8 spaces). If tabs is nil, the default
// tabs are reinstated.
func (layout *Layout) pango_layout_set_tabs(tabs *TabArray) {
	if layout == nil {
		return
	}

	if tabs != layout.tabs {
		layout.tabs = tabs.pango_tab_array_copy()
		layout.layout_changed()
	}
}

/**
 * pango_layout_get_line_readonly:
 * @layout: a #Layout
 * @line: the index of a line, which must be between 0 and
 *        <literal>pango_layout_get_line_count(layout) - 1</literal>, inclusive.
 *
 * Retrieves a particular line from a #Layout.
 *
 * This is a faster alternative to pango_layout_get_line(),
 * but the user is not expected
 * to modify the contents of the line (glyphs, glyph widths, etc.).
 *
 * Return value: (transfer none) (nullable): the requested
 *               #LayoutLine, or %nil if the index is out of
 *               range. This layout line can be ref'ed and retained,
 *               but will become invalid if changes are made to the
 *               #Layout.  No changes should be made to the line.
 **/
// func (layout *Layout) pango_layout_get_line_readonly(line int) *LayoutLine {
// 	GSList * list_item

// 	if line < 0 {
// 		return nil
// 	}

// 	layout.pango_layout_check_lines()

// 	list_item = g_slist_nth(layout.lines, line)

// 	if list_item {
// 		LayoutLine * line = list_item.data
// 		return line
// 	}

// 	return nil
// }

func affects_break_or_shape(attr *Attribute) bool {
	switch attr.Type {
	/* Affects breaks */
	case ATTR_ALLOW_BREAKS:
		return true
	/* Affects shaping */
	case ATTR_INSERT_HYPHENS, ATTR_FONT_FEATURES, ATTR_SHOW:
		return true
	default:
		return false
	}
}

func affects_itemization(attr *Attribute) bool {
	switch attr.Type {
	/* These affect font selection */
	case ATTR_LANGUAGE, ATTR_FAMILY, ATTR_STYLE, ATTR_WEIGHT, ATTR_VARIANT, ATTR_STRETCH, ATTR_SIZE, ATTR_FONT_DESC, ATTR_SCALE, ATTR_FALLBACK, ATTR_ABSOLUTE_SIZE, ATTR_GRAVITY, ATTR_GRAVITY_HINT:
		return true
	/* These need to be constant across runs */
	case ATTR_LETTER_SPACING, ATTR_SHAPE, ATTR_RISE:
		return true
	default:
		return false
	}
}

// func (layout *Layout) pango_layout_check_lines() {
// 	//    const char *start;
// 	//    bool done = false;
// 	//    int start_offset;
// 	//    PangoAttrList *itemize_attrs;
// 	//    PangoAttrList *shape_attrs;
// 	//    PangoAttrIterator iter;
// 	//    PangoDirection prev_base_dir = PANGO_DIRECTION_NEUTRAL, base_dir = PANGO_DIRECTION_NEUTRAL;
// 	//    ParaBreakState state;

// 	layout.check_context_changed()

// 	if len(layout.lines) != 0 {
// 		return
// 	}

// 	// assert(!layout.log_attrs)

// 	/* For simplicity, we make sure at this point that layout.text
// 	* is non-nil even if it is zero length
// 	 */
// 	if len(layout.text) == 0 {
// 		layout.SetText("")
// 	}

// 	attrs := layout.pango_layout_get_effective_attributes()
// 	var shape_attrs, itemize_attrs AttrList
// 	if len(attrs) != 0 {
// 		shape_attrs = attrs.pango_attr_list_filter(affects_break_or_shape)
// 		itemize_attrs = attrs.pango_attr_list_filter(affects_itemization)
// 		// shape_attrs = attr_list_filter(attrs, affects_break_or_shape, nil)
// 		// itemize_attrs = attr_list_filter(attrs, affects_itemization, nil)

// 		if len(itemize_attrs) != 0 {
// 			_attr_list_get_iterator(itemize_attrs, &iter)
// 		}
// 	}

// 	layout.log_attrs = make([]CharAttr, len(layout.text)+1)

// 	start_offset := 0
// 	start := layout.text

// 	/* Find the first strong direction of the text */
// 	if layout.auto_dir {
// 		prev_base_dir = pango_find_base_dir(layout.text, layout.length)
// 		if prev_base_dir == PANGO_DIRECTION_NEUTRAL {
// 			prev_base_dir = layout.context.base_dir
// 		}
// 	} else {
// 		base_dir = layout.context.base_dir
// 	}

// 	/* these are only used if layout.height >= 0 */
// 	state.remaining_height = layout.height
// 	state.line_height = -1
// 	if layout.height >= 0 {
// 		var logical Rectangle
// 		pango_layout_get_empty_extents_at_index(layout, 0, &logical)
// 		state.line_height = logical.height
// 	}

// 	done := false
// 	for !done {
// 		//    int delim_len;
// 		//    const char *end;
// 		//    int delimiter_index, next_para_index;

// 		if layout.single_paragraph {
// 			delimiter_index = layout.length
// 			next_para_index = layout.length
// 		} else {
// 			pango_find_paragraph_boundary(start,
// 				(layout.text+layout.length)-start,
// 				&delimiter_index,
// 				&next_para_index)
// 		}

// 		assert(next_para_index >= delimiter_index)

// 		if layout.auto_dir {
// 			base_dir = pango_find_base_dir(start, delimiter_index)

// 			/* Propagate the base direction for neutral paragraphs */
// 			if base_dir == PANGO_DIRECTION_NEUTRAL {
// 				base_dir = prev_base_dir
// 			} else {
// 				prev_base_dir = base_dir
// 			}
// 		}

// 		end = start + delimiter_index

// 		delim_len = next_para_index - delimiter_index

// 		if end == (layout.text + layout.length) {
// 			done = true
// 		}

// 		assert(end <= (layout.text + layout.length))
// 		assert(start <= (layout.text + layout.length))
// 		assert(delim_len < 4) /* PS is 3 bytes */
// 		assert(delim_len >= 0)

// 		state.attrs = itemize_attrs
// 		var iterPointer interface{}
// 		if itemize_attrs {
// 			iterPointer = &iter
// 		}
// 		state.items = pango_itemize_with_base_dir(layout.context,
// 			base_dir,
// 			layout.text,
// 			start-layout.text,
// 			end-start,
// 			itemize_attrs, iterPointer)

// 		apply_attributes_to_items(state.items, shape_attrs)

// 		get_items_log_attrs(layout.text,
// 			start-layout.text,
// 			delimiter_index+delim_len,
// 			state.items,
// 			layout.log_attrs+start_offset,
// 			len(layout.text)+1-start_offset)

// 		state.base_dir = base_dir
// 		state.line_of_par = 1
// 		state.start_offset = start_offset
// 		state.line_start_offset = start_offset
// 		state.line_start_index = start - layout.text

// 		state.glyphs = nil
// 		state.log_widths = nil
// 		state.need_hyphen = nil

// 		/* for deterministic bug hunting's sake set everything! */
// 		state.line_width = -1
// 		state.remaining_width = -1
// 		state.log_widths_offset = 0

// 		state.hyphen_width = -1

// 		if state.items {
// 			for state.items {
// 				process_line(layout, &state)
// 			}
// 		} else {
// 			LayoutLine * empty_line

// 			empty_line = pango_layout_line_new(layout)
// 			empty_line.start_index = state.line_start_index
// 			empty_line.is_paragraph_start = true
// 			line_set_resolved_dir(empty_line, base_dir)

// 			addLine(empty_line, &state)
// 		}

// 		if layout.height >= 0 && state.remaining_height < state.line_height {
// 			done = true
// 		}

// 		if !done {
// 			start_offset += pango_utf8_strlen(start, (end-start)+delim_len)
// 		}

// 		start = end + delim_len
// 	}

// 	apply_attributes_to_runs(layout, attrs)
// 	layout.lines = g_slist_reverse(layout.lines)
// }

func (layout *Layout) pango_layout_get_effective_attributes() AttrList {
	var attrs AttrList

	if len(layout.attrs) != 0 {
		attrs = layout.attrs.pango_attr_list_copy()
	}

	if layout.font_desc != nil {
		attr := pango_attr_font_desc_new(*layout.font_desc)
		attrs.pango_attr_list_insert_before(attr)
	}

	if layout.single_paragraph {
		attr := pango_attr_show_new(PANGO_SHOW_LINE_BREAKS)
		attrs.pango_attr_list_insert_before(attr)
	}

	return attrs
}

func (layout *Layout) pango_layout_get_empty_extents_at_index(index int, logical_rect *Rectangle) {
	if logical_rect == nil {
		return
	}

	font_desc := layout.context.font_desc // copy

	if layout.font_desc != nil {
		font_desc.pango_font_description_merge(layout.font_desc, true)
	}

	// Find the font description for this line
	if len(layout.attrs) != 0 {
		iter := layout.attrs.pango_attr_list_get_iterator()
		hasNext := true // do ... while
		for hasNext {
			start, end := iter.StartIndex, iter.EndIndex

			if start <= index && index < end {
				iter.pango_attr_iterator_get_font(&font_desc, nil, nil)
				break
			}

			hasNext = iter.pango_attr_iterator_next()
		}
	}

	font := layout.context.pango_context_load_font(&font_desc)
	if font != nil {
		metrics := pango_font_get_metrics(font, layout.context.set_language)
		// if metrics {
		logical_rect.y = -metrics.ascent
		logical_rect.height = -logical_rect.y + metrics.descent
		// } else {
		// 	logical_rect.y = 0
		// 	logical_rect.height = 0
		// }
	} else {
		logical_rect.y = 0
		logical_rect.height = 0
	}

	logical_rect.x = 0
	logical_rect.width = 0
}

// /**
//  * LayoutIter:
//  *
//  * A #LayoutIter structure can be used to
//  * iterate over the visual extents of a #Layout.
//  *
//  * The #LayoutIter structure is opaque, and
//  * has no user-visible fields.
//  */

//  #include "config.h"
//  #include "pango-glyph.h"		/* For pango_shape() */
//  #include "pango-break.h"
//  #include "pango-item.h"
//  #include "pango-engine.h"
//  #include "pango-impl-utils.h"
//  #include "pango-glyph-item.h"
//  #include <string.h>

//  #include "pango-layout-private.h"
//  #include "pango-attributes-private.h"

//  typedef struct _ItemProperties ItemProperties;
//  typedef struct _ParaBreakState ParaBreakState;

//  typedef struct _LayoutLinePrivate LayoutLinePrivate;

//  struct _LayoutClass
//  {
//    GObjectClass parent_class;

//  };

//  #define LINE_IS_VALID(line) ((line) && (line).layout != nil)

//  #ifdef G_DISABLE_CHECKS
//  #define ITER_IS_INVALID(iter) false
//  #else
//  #define ITER_IS_INVALID(iter) G_UNLIKELY (check_invalid ((iter), G_STRLOC))
//  static bool
//  check_invalid (LayoutIter *iter,
// 			const char      *loc)
//  {
//    if (iter.line.layout == nil)
// 	 {
// 	   g_warning ("%s: Layout changed since LayoutIter was created, iterator invalid", loc);
// 	   return true;
// 	 }
//    else
// 	 {
// 	   return false;
// 	 }
//  }
//  #endif

//  static void check_context_changed  (Layout *layout);
//  static void layout_changed  (Layout *layout);

//  static void pango_layout_clear_lines (Layout *layout);
//  static void pango_layout_check_lines (Layout *layout);

//  static PangoAttrList *pango_layout_get_effective_attributes (Layout *layout);

//  static LayoutLine * pango_layout_line_new         (Layout     *layout);
//  static void              pango_layout_line_postprocess (line *LayoutLine,
// 							 state *ParaBreakState ,
// 							 bool         wrapped);

//  static int *pango_layout_line_get_log2vis_map (LayoutLine  *line,
// 							bool          strong);
//  static int *pango_layout_line_get_vis2log_map (LayoutLine  *line,
// 							bool          strong);
//  static void pango_layout_line_leaked (line *LayoutLine);

//  /* doesn't leak line */
//  static LayoutLine* _pango_layout_iter_get_line (LayoutIter *iter);

//  static void pango_layout_get_item_properties (PangoItem      *item,
// 						   ItemProperties *properties);

//  static void pango_layout_get_empty_extents_at_index (layout *Layout    ,
// 							  int             index,
// 							  Rectangle *logical_rect);

//  static void pango_layout_finalize    (GObject          *object);

//  G_DEFINE_TYPE (Layout, pango_layout, G_TYPE_OBJECT)

//  static void
//  pango_layout_init (Layout *layout)
//  {
//    layout.serial = 1;
//    layout.attrs = nil;
//    layout.font_desc = nil;
//    layout.text = nil;
//    layout.length = 0;
//    layout.width = -1;
//    layout.height = -1;
//    layout.indent = 0;
//    layout.spacing = 0;
//    layout.line_spacing = 0.0;

//    layout.alignment = PANGO_ALIGN_LEFT;
//    layout.justify = false;
//    layout.auto_dir = true;

//    layout.log_attrs = nil;
//    layout.lines = nil;
//    layout.line_count = 0;

//    layout.tabWidth = -1;
//    layout.unknown_glyphs_count = -1;

//    layout.wrap = PANGO_WRAP_WORD;
//    layout.is_wrapped = false;
//    layout.ellipsize = PANGO_ELLIPSIZE_NONE;
//    layout.is_ellipsized = false;
//  }

//  static void
//  pango_layout_class_init (LayoutClass *klass)
//  {
//    GObjectClass *object_class = G_OBJECT_CLASS (klass);

//    object_class.finalize = pango_layout_finalize;
//  }

//  static void
//  pango_layout_finalize (GObject *object)
//  {
//    Layout *layout;

//    layout = PANGO_LAYOUT (object);

//    pango_layout_clear_lines (layout);

//    if (layout.context)
// 	 g_object_unref (layout.context);

//    if (layout.attrs)
// 	 attr_list_unref (layout.attrs);

//    g_free (layout.text);

//    if (layout.font_desc)
// 	 pango_font_description_free (layout.font_desc);

//    if (layout.tabs)
// 	 pango_tab_array_free (layout.tabs);

//    G_OBJECT_CLASS (pango_layout_parent_class).finalize (object);
//  }

//  /**
//   * pango_layout_new:
//   * @context: a #PangoContext
//   *
//   * Create a new #Layout object with attributes initialized to
//   * default values for a particular #PangoContext.
//   *
//   * Return value: the newly allocated #Layout, with a reference
//   *               count of one, which should be freed with
//   *               g_object_unref().
//   **/
//  Layout *
//  pango_layout_new (PangoContext *context)
//  {
//    Layout *layout;

//    g_return_val_if_fail (context != nil, nil);

//    layout = g_object_new (PANGO_TYPE_LAYOUT, nil);

//    layout.context = context;
//    layout.context_serial = pango_context_get_serial (context);
//    g_object_ref (context);

//    return layout;
//  }

//  /**
//   * pango_layout_copy:
//   * @src: a #Layout
//   *
//   * Does a deep copy-by-value of the @src layout. The attribute list,
//   * tab array, and text from the original layout are all copied by
//   * value.
//   *
//   * Return value: (transfer full): the newly allocated #Layout,
//   *               with a reference count of one, which should be freed
//   *               with g_object_unref().
//   **/
//  Layout*
//  pango_layout_copy (Layout *src)
//  {
//    Layout *layout;

//    g_return_val_if_fail (PANGO_IS_LAYOUT (src), nil);

//    /* Copy referenced members */

//    layout = pango_layout_new (src.context);
//    if (src.attrs)
// 	 layout.attrs = attr_list_copy (src.attrs);
//    if (src.font_desc)
// 	 layout.font_desc = pango_font_description_copy (src.font_desc);
//    if (src.tabs)
// 	 layout.tabs = pango_tab_array_copy (src.tabs);

//    /* Dupped */
//    layout.text = g_strdup (src.text);

//    /* Value fields */
//    memcpy (&layout.copy_begin, &src.copy_begin,
// 	   G_STRUCT_OFFSET (Layout, copy_end) - G_STRUCT_OFFSET (Layout, copy_begin));

//    return layout;
//  }

//  /**
//   * pango_layout_get_context:
//   * @layout: a #Layout
//   *
//   * Retrieves the #PangoContext used for this layout.
//   *
//   * Return value: (transfer none): the #PangoContext for the layout.
//   * This does not have an additional refcount added, so if you want to
//   * keep a copy of this around, you must reference it yourself.
//   **/
//  PangoContext *
//  pango_layout_get_context (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, nil);

//    return layout.context;
//  }

//  /**
//   * pango_layout_set_width:
//   * @layout: a #Layout.
//   * @width: the desired width in Pango units, or -1 to indicate that no
//   *         wrapping or ellipsization should be performed.
//   *
//   * Sets the width to which the lines of the #Layout should wrap or
//   * ellipsized.  The default value is -1: no width set.
//   **/
//  void
//  pango_layout_set_width (Layout *layout,
// 			 int          width)
//  {
//    g_return_if_fail (layout != nil);

//    if (width < 0)
// 	 width = -1;

//    if (width != layout.width)
// 	 {
// 	   layout.width = width;

// 	   /* Increasing the width can only decrease the line count */
// 	   if (layout.line_count == 1 && width > layout.width)
// 		 return;

// 	   layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_width:
//   * @layout: a #Layout
//   *
//   * Gets the width to which the lines of the #Layout should wrap.
//   *
//   * Return value: the width in Pango units, or -1 if no width set.
//   **/
//  int
//  pango_layout_get_width (layout *Layout    )
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.width;
//  }

//  /**
//   * pango_layout_set_height:
//   * @layout: a #Layout.
//   * @height: the desired height of the layout in Pango units if positive,
//   *          or desired number of lines if negative.
//   *
//   * Sets the height to which the #Layout should be ellipsized at.  There
//   * are two different behaviors, based on whether @height is positive or
//   * negative.
//   *
//   * If @height is positive, it will be the maximum height of the layout.  Only
//   * lines would be shown that would fit, and if there is any text omitted,
//   * an ellipsis added.  At least one line is included in each paragraph regardless
//   * of how small the height value is.  A value of zero will render exactly one
//   * line for the entire layout.
//   *
//   * If @height is negative, it will be the (negative of) maximum number of lines per
//   * paragraph.  That is, the total number of lines shown may well be more than
//   * this value if the layout contains multiple paragraphs of text.
//   * The default value of -1 means that first line of each paragraph is ellipsized.
//   * This behvaior may be changed in the future to act per layout instead of per
//   * paragraph.  File a bug against pango at <ulink
//   * url="http://bugzilla.gnome.org/">http://bugzilla.gnome.org/</ulink> if your
//   * code relies on this behavior.
//   *
//   * Height setting only has effect if a positive width is set on
//   * @layout and ellipsization mode of @layout is not %PANGO_ELLIPSIZE_NONE.
//   * The behavior is undefined if a height other than -1 is set and
//   * ellipsization mode is set to %PANGO_ELLIPSIZE_NONE, and may change in the
//   * future.
//   *
//   * Since: 1.20
//   **/
//  void
//  pango_layout_set_height (Layout *layout,
// 			  int          height)
//  {
//    g_return_if_fail (layout != nil);

//    if (height != layout.height)
// 	 {
// 	   layout.height = height;

// 	   /* Do not invalidate if the number of lines requested is
// 		* larger than the total number of lines in layout.
// 		* Bug 549003
// 		*/
// 	   if (layout.ellipsize != PANGO_ELLIPSIZE_NONE &&
// 	   !(layout.lines && layout.is_ellipsized == false &&
// 		 height < 0 && layout.line_count <= (guint) -height))
// 	 layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_height:
//   * @layout: a #Layout
//   *
//   * Gets the height of layout used for ellipsization.  See
//   * pango_layout_set_height() for details.
//   *
//   * Return value: the height, in Pango units if positive, or
//   * number of lines if negative.
//   *
//   * Since: 1.20
//   **/
//  int
//  pango_layout_get_height (layout *Layout    )
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.height;
//  }

//  /**
//   * pango_layout_set_wrap:
//   * @layout: a #Layout
//   * @wrap: the wrap mode
//   *
//   * Sets the wrap mode; the wrap mode only has effect if a width
//   * is set on the layout with pango_layout_set_width().
//   * To turn off wrapping, set the width to -1.
//   **/
//  void
//  pango_layout_set_wrap (Layout  *layout,
// 				PangoWrapMode wrap)
//  {
//    g_return_if_fail (PANGO_IS_LAYOUT (layout));

//    if (layout.wrap != wrap)
// 	 {
// 	   layout.wrap = wrap;

// 	   if (layout.width != -1)
// 	 layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_wrap:
//   * @layout: a #Layout
//   *
//   * Gets the wrap mode for the layout.
//   *
//   * Use pango_layout_is_wrapped() to query whether any paragraphs
//   * were actually wrapped.
//   *
//   * Return value: active wrap mode.
//   **/
//  PangoWrapMode
//  pango_layout_get_wrap (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), 0);

//    return layout.wrap;
//  }

//  /**
//   * pango_layout_is_wrapped:
//   * @layout: a #Layout
//   *
//   * Queries whether the layout had to wrap any paragraphs.
//   *
//   * This returns `true` if a positive width is set on @layout,
//   * ellipsization mode of @layout is set to %PANGO_ELLIPSIZE_NONE,
//   * and there are paragraphs exceeding the layout width that have
//   * to be wrapped.
//   *
//   * Return value: `true` if any paragraphs had to be wrapped, %false
//   * otherwise.
//   *
//   * Since: 1.16
//   */
//  bool
//  pango_layout_is_wrapped (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, false);

//    pango_layout_check_lines (layout);

//    return layout.is_wrapped;
//  }

//  /**
//   * pango_layout_set_indent:
//   * @layout: a #Layout.
//   * @indent: the amount by which to indent.
//   *
//   * Sets the width in Pango units to indent each paragraph. A negative value
//   * of @indent will produce a hanging indentation. That is, the first line will
//   * have the full width, and subsequent lines will be indented by the
//   * absolute value of @indent.
//   *
//   * The indent setting is ignored if layout alignment is set to
//   * %PANGO_ALIGN_CENTER.
//   **/
//  void
//  pango_layout_set_indent (Layout *layout,
// 			  int          indent)
//  {
//    g_return_if_fail (layout != nil);

//    if (indent != layout.indent)
// 	 {
// 	   layout.indent = indent;
// 	   layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_indent:
//   * @layout: a #Layout
//   *
//   * Gets the paragraph indent width in Pango units. A negative value
//   * indicates a hanging indentation.
//   *
//   * Return value: the indent in Pango units.
//   **/
//  int
//  pango_layout_get_indent (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.indent;
//  }

//  /**
//   * pango_layout_set_spacing:
//   * @layout: a #Layout.
//   * @spacing: the amount of spacing
//   *
//   * Sets the amount of spacing in Pango unit between
//   * the lines of the layout. When placing lines with
//   * spacing, Pango arranges things so that
//   *
//   * line2.top = line1.bottom + spacing
//   *
//   * Note: Since 1.44, Pango defaults to using the
//   * line height (as determined by the font) for placing
//   * lines. The @spacing set with this function is only
//   * taken into account when the line-height factor is
//   * set to zero with pango_layout_set_line_spacing().
//   **/
//  void
//  pango_layout_set_spacing (Layout *layout,
// 			   int          spacing)
//  {
//    g_return_if_fail (layout != nil);

//    if (spacing != layout.spacing)
// 	 {
// 	   layout.spacing = spacing;
// 	   layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_spacing:
//   * @layout: a #Layout
//   *
//   * Gets the amount of spacing between the lines of the layout.
//   *
//   * Return value: the spacing in Pango units.
//   **/
//  int
//  pango_layout_get_spacing (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, 0);
//    return layout.spacing;
//  }

//  /**
//   * pango_layout_set_line_spacing:
//   * @layout: a #Layout
//   * @factor: the new line spacing factor
//   *
//   * Sets a factor for line spacing.
//   * Typical values are: 0, 1, 1.5, 2.
//   * The default values is 0.
//   *
//   * If @factor is non-zero, lines are placed
//   * so that
//   *
//   * baseline2 = baseline1 + factor * height2
//   *
//   * where height2 is the line height of the
//   * second line (as determined by the font(s)).
//   * In this case, the spacing set with
//   * pango_layout_set_spacing() is ignored.
//   *
//   * If @factor is zero, spacing is applied as
//   * before.
//   *
//   * Since: 1.44
//   */
//  void
//  pango_layout_set_line_spacing (Layout *layout,
// 					float        factor)
//  {
//    g_return_if_fail (layout != nil);

//    if (layout.line_spacing != factor)
// 	 {
// 	   layout.line_spacing = factor;
// 	   layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_line_spacing:
//   * @layout: a #Layout
//   *
//   * Gets the value that has been
//   * set with pango_layout_set_line_spacing().
//   *
//   * Since: 1.44
//   */
//  float
//  pango_layout_get_line_spacing (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, 1.0);
//    return layout.line_spacing;
//  }

//  /**
//   * pango_layout_set_font_description:
//   * @layout: a #Layout
//   * @desc: (allow-none): the new #PangoFontDescription, or %nil to unset the
//   *        current font description
//   *
//   * Sets the default font description for the layout. If no font
//   * description is set on the layout, the font description from
//   * the layout's context is used.
//   **/
//  void
//  pango_layout_set_font_description (Layout                 *layout,
// 					 const PangoFontDescription *desc)
//  {
//    g_return_if_fail (layout != nil);

//    if (desc != layout.font_desc &&
// 	   (!desc || !layout.font_desc || !pango_font_description_equal(desc, layout.font_desc)))
// 	 {
// 	   if (layout.font_desc)
// 	 pango_font_description_free (layout.font_desc);

// 	   layout.font_desc = desc ? pango_font_description_copy (desc) : nil;

// 	   layout_changed (layout);
// 	   layout.tabWidth = -1;
// 	 }
//  }

//  /**
//   * pango_layout_get_font_description:
//   * @layout: a #Layout
//   *
//   * Gets the font description for the layout, if any.
//   *
//   * Return value: (nullable): a pointer to the layout's font
//   *  description, or %nil if the font description from the layout's
//   *  context is inherited. This value is owned by the layout and must
//   *  not be modified or freed.
//   *
//   * Since: 1.8
//   **/
//  const PangoFontDescription *
//  pango_layout_get_font_description (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), nil);

//    return layout.font_desc;
//  }

//  /**
//   * pango_layout_set_justify:
//   * @layout: a #Layout
//   * @justify: whether the lines in the layout should be justified.
//   *
//   * Sets whether each complete line should be stretched to
//   * fill the entire width of the layout. This stretching is typically
//   * done by adding whitespace, but for some scripts (such as Arabic),
//   * the justification may be done in more complex ways, like extending
//   * the characters.
//   *
//   * Note that this setting is not implemented and so is ignored in Pango
//   * older than 1.18.
//   **/
//  void
//  pango_layout_set_justify (Layout *layout,
// 			   bool     justify)
//  {
//    g_return_if_fail (layout != nil);

//    if (justify != layout.justify)
// 	 {
// 	   layout.justify = justify;

// 	   if (layout.is_ellipsized || layout.is_wrapped)
// 	 layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_justify:
//   * @layout: a #Layout
//   *
//   * Gets whether each complete line should be stretched to fill the entire
//   * width of the layout.
//   *
//   * Return value: the justify.
//   **/
//  bool
//  pango_layout_get_justify (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, false);
//    return layout.justify;
//  }

//  /**
//   * pango_layout_set_auto_dir:
//   * @layout: a #Layout
//   * @auto_dir: if `true`, compute the bidirectional base direction
//   *   from the layout's contents.
//   *
//   * Sets whether to calculate the bidirectional base direction
//   * for the layout according to the contents of the layout;
//   * when this flag is on (the default), then paragraphs in
// 	@layout that begin with strong right-to-left characters
//   * (Arabic and Hebrew principally), will have right-to-left
//   * layout, paragraphs with letters from other scripts will
//   * have left-to-right layout. Paragraphs with only neutral
//   * characters get their direction from the surrounding paragraphs.
//   *
//   * When %false, the choice between left-to-right and
//   * right-to-left layout is done according to the base direction
//   * of the layout's #PangoContext. (See pango_context_set_base_dir()).
//   *
//   * When the auto-computed direction of a paragraph differs from the
//   * base direction of the context, the interpretation of
//   * %PANGO_ALIGN_LEFT and %PANGO_ALIGN_RIGHT are swapped.
//   *
//   * Since: 1.4
//   **/
//  void
//  pango_layout_set_auto_dir (Layout *layout,
// 				bool     auto_dir)
//  {
//    g_return_if_fail (PANGO_IS_LAYOUT (layout));

//    auto_dir = auto_dir != false;

//    if (auto_dir != layout.auto_dir)
// 	 {
// 	   layout.auto_dir = auto_dir;
// 	   layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_auto_dir:
//   * @layout: a #Layout
//   *
//   * Gets whether to calculate the bidirectional base direction
//   * for the layout according to the contents of the layout.
//   * See pango_layout_set_auto_dir().
//   *
//   * Return value: `true` if the bidirectional base direction
//   *   is computed from the layout's contents, %false otherwise.
//   *
//   * Since: 1.4
//   **/
//  bool
//  pango_layout_get_auto_dir (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), false);

//    return layout.auto_dir;
//  }

//  /**
//   * pango_layout_set_alignment:
//   * @layout: a #Layout
//   * @alignment: the alignment
//   *
//   * Sets the alignment for the layout: how partial lines are
//   * positioned within the horizontal space available.
//   **/
//  void
//  pango_layout_set_alignment (layout *Layout,
// 				 PangoAlignment alignment)
//  {
//    g_return_if_fail (layout != nil);

//    if (alignment != layout.alignment)
// 	 {
// 	   layout.alignment = alignment;
// 	   layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_alignment:
//   * @layout: a #Layout
//   *
//   * Gets the alignment for the layout: how partial lines are
//   * positioned within the horizontal space available.
//   *
//   * Return value: the alignment.
//   **/
//  PangoAlignment
//  pango_layout_get_alignment (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, PANGO_ALIGN_LEFT);
//    return layout.alignment;
//  }

//  /**
//   * pango_layout_get_tabs:
//   * @layout: a #Layout
//   *
//   * Gets the current #TabArray used by this layout. If no
//   * #TabArray has been set, then the default tabs are in use
//   * and %nil is returned. Default tabs are every 8 spaces.
//   * The return value should be freed with pango_tab_array_free().
//   *
//   * Return value: (nullable): a copy of the tabs for this layout, or
//   * %nil.
//   **/
//  TabArray*
//  pango_layout_get_tabs (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), nil);

//    if (layout.tabs)
// 	 return pango_tab_array_copy (layout.tabs);
//    else
// 	 return nil;
//  }

//  /**
//   * pango_layout_set_single_paragraph_mode:
//   * @layout: a #Layout
//   * @setting: new setting
//   *
//   * If @setting is `true`, do not treat newlines and similar characters
//   * as paragraph separators; instead, keep all text in a single paragraph,
//   * and display a glyph for paragraph separator characters. Used when
//   * you want to allow editing of newlines on a single text line.
//   **/
//  void
//  pango_layout_set_single_paragraph_mode (Layout *layout,
// 					 bool     setting)
//  {
//    g_return_if_fail (PANGO_IS_LAYOUT (layout));

//    setting = setting != false;

//    if (layout.single_paragraph != setting)
// 	 {
// 	   layout.single_paragraph = setting;
// 	   layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_single_paragraph_mode:
//   * @layout: a #Layout
//   *
//   * Obtains the value set by pango_layout_set_single_paragraph_mode().
//   *
//   * Return value: `true` if the layout does not break paragraphs at
//   * paragraph separator characters, %false otherwise.
//   **/
//  bool
//  pango_layout_get_single_paragraph_mode (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), false);

//    return layout.single_paragraph;
//  }

//  /**
//   * pango_layout_set_ellipsize:
//   * @layout: a #Layout
//   * @ellipsize: the new ellipsization mode for @layout
//   *
//   * Sets the type of ellipsization being performed for @layout.
//   * Depending on the ellipsization mode @ellipsize text is
//   * removed from the start, middle, or end of text so they
//   * fit within the width and height of layout set with
//   * pango_layout_set_width() and pango_layout_set_height().
//   *
//   * If the layout contains characters such as newlines that
//   * force it to be layed out in multiple paragraphs, then whether
//   * each paragraph is ellipsized separately or the entire layout
//   * is ellipsized as a whole depends on the set height of the layout.
//   * See pango_layout_set_height() for details.
//   *
//   * Since: 1.6
//   **/
//  void
//  pango_layout_set_ellipsize (Layout        *layout,
// 				 PangoEllipsizeMode  ellipsize)
//  {
//    g_return_if_fail (PANGO_IS_LAYOUT (layout));

//    if (ellipsize != layout.ellipsize)
// 	 {
// 	   layout.ellipsize = ellipsize;

// 	   if (layout.is_ellipsized || layout.is_wrapped)
// 	 layout_changed (layout);
// 	 }
//  }

//  /**
//   * pango_layout_get_ellipsize:
//   * @layout: a #Layout
//   *
//   * Gets the type of ellipsization being performed for @layout.
//   * See pango_layout_set_ellipsize()
//   *
//   * Return value: the current ellipsization mode for @layout.
//   *
//   * Use pango_layout_is_ellipsized() to query whether any paragraphs
//   * were actually ellipsized.
//   *
//   * Since: 1.6
//   **/
//  PangoEllipsizeMode
//  pango_layout_get_ellipsize (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), PANGO_ELLIPSIZE_NONE);

//    return layout.ellipsize;
//  }

//  /**
//   * pango_layout_is_ellipsized:
//   * @layout: a #Layout
//   *
//   * Queries whether the layout had to ellipsize any paragraphs.
//   *
//   * This returns `true` if the ellipsization mode for @layout
//   * is not %PANGO_ELLIPSIZE_NONE, a positive width is set on @layout,
//   * and there are paragraphs exceeding that width that have to be
//   * ellipsized.
//   *
//   * Return value: `true` if any paragraphs had to be ellipsized, %false
//   * otherwise.
//   *
//   * Since: 1.16
//   */
//  bool
//  pango_layout_is_ellipsized (Layout *layout)
//  {
//    g_return_val_if_fail (layout != nil, false);

//    pango_layout_check_lines (layout);

//    return layout.is_ellipsized;
//  }

//  /**
//   * pango_layout_set_text:
//   * @layout: a #Layout
//   * @text: the text
//   * @length: maximum length of @text, in bytes. -1 indicates that
//   *          the string is nul-terminated and the length should be
//   *          calculated.  The text will also be truncated on
//   *          encountering a nul-termination even when @length is
//   *          positive.
//   *
//   * Sets the text of the layout.
//   *
//   * This function validates @text and renders invalid UTF-8
//   * with a placeholder glyph.
//   *
//   * Note that if you have used pango_layout_set_markup() or
//   * pango_layout_set_markup_with_accel() on @layout before, you may
//   * want to call pango_layout_set_attributes() to clear the attributes
//   * set on the layout from the markup as this function does not clear
//   * attributes.
//   **/
//  void
//  pango_layout_set_text (Layout *layout,
// 				const char  *text,
// 				int          length)
//  {
//    char *old_text, *start, *end;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (length == 0 || text != nil);

//    old_text = layout.text;

//    if (length < 0)
// 	 {
// 	   layout.length = strlen (text);
// 	   layout.text = g_strndup (text, layout.length);
// 	 }
//    else if (length > 0)
// 	 {
// 	   /* This is not exactly what we want.  We don't need the padding...
// 		*/
// 	   layout.length = length;
// 	   layout.text = g_strndup (text, length);
// 	 }
//    else
// 	 {
// 	   layout.length = 0;
// 	   layout.text = g_malloc0 (1);
// 	 }

//    /* validate it, and replace invalid bytes with -1 */
//    start = layout.text;
//    for (;;) {
// 	 bool valid;

// 	 valid = g_utf8_validate (start, -1, (const char **)&end);

// 	 if (!*end)
// 	   break;

// 	 /* Replace invalid bytes with -1.  The -1 will be converted to
// 	  * ((gunichar) -1) by glib, and that in turn yields a glyph value of
// 	  * ((PangoGlyph) -1) by PANGO_GET_UNKNOWN_GLYPH(-1),
// 	  * and that's PANGO_GLYPH_INVALID_INPUT.
// 	  */
// 	 if (!valid)
// 	   *end++ = -1;

// 	 start = end;
//    }

//    if (start != layout.text)
// 	 /* TODO: Write out the beginning excerpt of text? */
// 	 g_warning ("Invalid UTF-8 string passed to pango_layout_set_text()");

//    len(layout.text) = pango_utf8_strlen (layout.text, -1);
//    layout.length = strlen (layout.text);

//    layout_changed (layout);

//    g_free (old_text);
//  }

//  /**
//   * pango_layout_get_text:
//   * @layout: a #Layout
//   *
//   * Gets the text in the layout. The returned text should not
//   * be freed or modified.
//   *
//   * Return value: the text in the @layout.
//   **/
//  const char*
//  pango_layout_get_text (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), nil);

//    /* We don't ever want to return nil as the text.
// 	*/
//    if (G_UNLIKELY (!layout.text))
// 	 return "";

//    return layout.text;
//  }

//  /**
//   * pango_layout_get_character_count:
//   * @layout: a #Layout
//   *
//   * Returns the number of Unicode characters in the
//   * the text of @layout.
//   *
//   * Return value: the number of Unicode characters
//   *     in the text of @layout
//   *
//   * Since: 1.30
//   */
//  gint
//  pango_layout_get_character_count (Layout *layout)
//  {
//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), 0);

//    return len(layout.text);
//  }

//  /**
//   * pango_layout_set_markup:
//   * @layout: a #Layout
//   * @markup: marked-up text
//   * @length: length of marked-up text in bytes, or -1 if @markup is
//   *          null-terminated
//   *
//   * Same as pango_layout_set_markup_with_accel(), but
//   * the markup text isn't scanned for accelerators.
//   *
//   **/
//  void
//  pango_layout_set_markup (Layout *layout,
// 			  const char  *markup,
// 			  int          length)
//  {
//    pango_layout_set_markup_with_accel (layout, markup, length, 0, nil);
//  }

//  /**
//   * pango_layout_set_markup_with_accel:
//   * @layout: a #Layout
//   * @markup: marked-up text
//   * (see <link linkend="PangoMarkupFormat">markup format</link>)
//   * @length: length of marked-up text in bytes, or -1 if @markup is
//   *          null-terminated
//   * @accel_marker: marker for accelerators in the text
//   * @accel_char: (out caller-allocates) (allow-none): return location
//   *                    for first located accelerator, or %nil
//   *
//   * Sets the layout text and attribute list from marked-up text (see
//   * <link linkend="PangoMarkupFormat">markup format</link>). Replaces
//   * the current text and attribute list.
//   *
//   * If @accel_marker is nonzero, the given character will mark the
//   * character following it as an accelerator. For example, @accel_marker
//   * might be an ampersand or underscore. All characters marked
//   * as an accelerator will receive a %PANGO_UNDERLINE_LOW attribute,
//   * and the first character so marked will be returned in @accel_char.
//   * Two @accel_marker characters following each other produce a single
//   * literal @accel_marker character.
//   **/
//  void
//  pango_layout_set_markup_with_accel (layout *Layout    ,
// 					 const char     *markup,
// 					 int             length,
// 					 gunichar        accel_marker,
// 					 gunichar       *accel_char)
//  {
//    PangoAttrList *list = nil;
//    char *text = nil;
//    GError *error;

//    g_return_if_fail (PANGO_IS_LAYOUT (layout));
//    g_return_if_fail (markup != nil);

//    error = nil;
//    if (!pango_parse_markup (markup, length,
// 				accel_marker,
// 				&list, &text,
// 				accel_char,
// 				&error))
// 	 {
// 	   g_warning ("pango_layout_set_markup_with_accel: %s", error.message);
// 	   g_error_free (error);
// 	   return;
// 	 }

//    pango_layout_set_text (layout, text, -1);
//    pango_layout_set_attributes (layout, list);
//    attr_list_unref (list);
//    g_free (text);
//  }

//  /**
//   * pango_layout_get_unknown_glyphs_count:
//   * @layout: a #Layout
//   *
//   * Counts the number unknown glyphs in @layout.  That is, zero if
//   * glyphs for all characters in the layout text were found, or more
//   * than zero otherwise.
//   *
//   * This function can be used to determine if there are any fonts
//   * available to render all characters in a certain string, or when
//   * used in combination with %ATTR_FALLBACK, to check if a
//   * certain font supports all the characters in the string.
//   *
//   * Return value: The number of unknown glyphs in @layout.
//   *
//   * Since: 1.16
//   */
//  int
//  pango_layout_get_unknown_glyphs_count (Layout *layout)
//  {
// 	 line *LayoutLine;
// 	 GlyphItem *run;
// 	 GSList *lines_list;
// 	 GSList *runs_list;
// 	 int i, count = 0;

// 	 g_return_val_if_fail (PANGO_IS_LAYOUT (layout), 0);

// 	 pango_layout_check_lines (layout);

// 	 if (layout.unknown_glyphs_count >= 0)
// 	   return layout.unknown_glyphs_count;

// 	 lines_list = layout.lines;
// 	 for (lines_list)
// 	   {
// 	 line = lines_list.data;
// 	 runs_list = line.runs;

// 	 for (runs_list)
// 	   {
// 		 run = runs_list.data;

// 		 for (i = 0; i < run.glyphs.num_glyphs; i++)
// 		   {
// 		 if (run.glyphs.glyphs[i].glyph & PANGO_GLYPH_UNKNOWN_FLAG)
// 			 count++;
// 		   }

// 		 runs_list = runs_list.next;
// 	   }
// 	 lines_list = lines_list.next;
// 	   }

// 	 layout.unknown_glyphs_count = count;
// 	 return count;
//  }

//  /**
//   * pango_layout_get_serial:
//   * @layout: a #Layout
//   *
//   * Returns the current serial number of @layout.  The serial number is
//   * initialized to an small number  larger than zero when a new layout
//   * is created and is increased whenever the layout is changed using any
//   * of the setter functions, or the #PangoContext it uses has changed.
//   * The serial may wrap, but will never have the value 0. Since it
//   * can wrap, never compare it with "less than", always use "not equals".
//   *
//   * This can be used to automatically detect changes to a #Layout, and
//   * is useful for example to decide whether a layout needs redrawing.
//   * To force the serial to be increased, use pango_layout_context_changed().
//   *
//   * Return value: The current serial number of @layout.
//   *
//   * Since: 1.32.4
//   **/
//  guint
//  pango_layout_get_serial (Layout *layout)
//  {
//    check_context_changed (layout);
//    return layout.serial;
//  }

//  /**
//   * pango_layout_get_log_attrs:
//   * @layout: a #Layout
//   * @attrs: (out)(array length=n_attrs)(transfer container):
//   *         location to store a pointer to an array of logical attributes
//   *         This value must be freed with g_free().
//   * @n_attrs: (out): location to store the number of the attributes in the
//   *           array. (The stored value will be one more than the total number
//   *           of characters in the layout, since there need to be attributes
//   *           corresponding to both the position before the first character
//   *           and the position after the last character.)
//   *
//   * Retrieves an array of logical attributes for each character in
//   * the @layout.
//   **/
//  void
//  pango_layout_get_log_attrs (layout *Layout    ,
// 				 PangoLogAttr  **attrs,
// 				 gint           *n_attrs)
//  {
//    g_return_if_fail (layout != nil);

//    pango_layout_check_lines (layout);

//    if (attrs)
// 	 {
// 	   *attrs = g_new (PangoLogAttr, len(layout.text) + 1);
// 	   memcpy (*attrs, layout.log_attrs, sizeof(PangoLogAttr) * (len(layout.text) + 1));
// 	 }

//    if (n_attrs)
// 	 *n_attrs = len(layout.text) + 1;
//  }

//  /**
//   * pango_layout_get_log_attrs_readonly:
//   * @layout: a #Layout
//   * @n_attrs: (out): location to store the number of the attributes in
//   *   the array
//   *
//   * Retrieves an array of logical attributes for each character in
//   * the @layout.
//   *
//   * This is a faster alternative to pango_layout_get_log_attrs().
//   * The returned array is part of @layout and must not be modified.
//   * Modifying the layout will invalidate the returned array.
//   *
//   * The number of attributes returned in @n_attrs will be one more
//   * than the total number of characters in the layout, since there
//   * need to be attributes corresponding to both the position before
//   * the first character and the position after the last character.
//   *
//   * Returns: (array length=n_attrs): an array of logical attributes
//   *
//   * Since: 1.30
//   */
//  const PangoLogAttr *
//  pango_layout_get_log_attrs_readonly (Layout *layout,
// 									  gint        *n_attrs)
//  {
//    if (n_attrs)
// 	 *n_attrs = 0;
//    g_return_val_if_fail (layout != nil, nil);

//    pango_layout_check_lines (layout);

//    if (n_attrs)
// 	 *n_attrs = len(layout.text) + 1;

//    return layout.log_attrs;
//  }

//  /**
//   * pango_layout_get_line_count:
//   * @layout: #Layout
//   *
//   * Retrieves the count of lines for the @layout.
//   *
//   * Return value: the line count.
//   **/
//  int
//  pango_layout_get_line_count (layout *Layout)
//  {
//    g_return_val_if_fail (layout != nil, 0);

//    pango_layout_check_lines (layout);
//    return layout.line_count;
//  }

//  /**
//   * pango_layout_get_lines:
//   * @layout: a #Layout
//   *
//   * Returns the lines of the @layout as a list.
//   *
//   * Use the faster pango_layout_get_lines_readonly() if you do not plan
//   * to modify the contents of the lines (glyphs, glyph widths, etc.).
//   *
//   * Return value: (element-type Pango.LayoutLine) (transfer none): a #GSList containing
//   * the lines in the layout. This points to internal data of the #Layout
//   * and must be used with care. It will become invalid on any change to the layout's
//   * text or properties.
//   **/
//  GSList *
//  pango_layout_get_lines (Layout *layout)
//  {
//    pango_layout_check_lines (layout);

//    if (layout.lines)
// 	 {
// 	   GSList *tmp_list = layout.lines;
// 	   for (tmp_list)
// 	 {
// 	   line *LayoutLine = tmp_list.data;
// 	   tmp_list = tmp_list.next;

// 	   pango_layout_line_leaked (line);
// 	 }
// 	 }

//    return layout.lines;
//  }

//  /**
//   * pango_layout_get_lines_readonly:
//   * @layout: a #Layout
//   *
//   * Returns the lines of the @layout as a list.
//   *
//   * This is a faster alternative to pango_layout_get_lines(),
//   * but the user is not expected
//   * to modify the contents of the lines (glyphs, glyph widths, etc.).
//   *
//   * Return value: (element-type Pango.LayoutLine) (transfer none): a #GSList containing
//   * the lines in the layout. This points to internal data of the #Layout and
//   * must be used with care. It will become invalid on any change to the layout's
//   * text or properties.  No changes should be made to the lines.
//   *
//   * Since: 1.16
//   **/
//  GSList *
//  pango_layout_get_lines_readonly (Layout *layout)
//  {
//    pango_layout_check_lines (layout);

//    return layout.lines;
//  }

//  /**
//   * pango_layout_get_line:
//   * @layout: a #Layout
//   * @line: the index of a line, which must be between 0 and
//   *        <literal>pango_layout_get_line_count(layout) - 1</literal>, inclusive.
//   *
//   * Retrieves a particular line from a #Layout.
//   *
//   * Use the faster pango_layout_get_line_readonly() if you do not plan
//   * to modify the contents of the line (glyphs, glyph widths, etc.).
//   *
//   * Return value: (transfer none) (nullable): the requested
//   *               #LayoutLine, or %nil if the index is out of
//   *               range. This layout line can be ref'ed and retained,
//   *               but will become invalid if changes are made to the
//   *               #Layout.
//   **/
//  LayoutLine *
//  pango_layout_get_line (Layout *layout,
// 				int          line)
//  {
//    GSList *list_item;
//    g_return_val_if_fail (layout != nil, nil);

//    if (line < 0)
// 	 return nil;

//    pango_layout_check_lines (layout);

//    list_item = g_slist_nth (layout.lines, line);

//    if (list_item)
// 	 {
// 	   line *LayoutLine = list_item.data;

// 	   pango_layout_line_leaked (line);
// 	   return line;
// 	 }

//    return nil;
//  }

//  /**
//   * pango_layout_line_index_to_x:
//   * @line:     a #LayoutLine
//   * @index_:   byte offset of a grapheme within the layout
//   * @trailing: an integer indicating the edge of the grapheme to retrieve
//   *            the position of. If > 0, the trailing edge of the grapheme,
//   *            if 0, the leading of the grapheme.
//   * @x_pos: (out): location to store the x_offset (in Pango unit)
//   *
//   * Converts an index within a line to a X position.
//   *
//   **/
//  void
//  pango_layout_line_index_to_x (LayoutLine  *line,
// 				   int               index,
// 				   int               trailing,
// 				   int              *x_pos)
//  {
//    Layout *layout = line.layout;
//    GSList *run_list = line.runs;
//    int width = 0;

//    for (run_list)
// 	 {
// 	   GlyphItem *run = run_list.data;

// 	   if (run.item.offset <= index && run.item.offset + run.item.length > index)
// 	 {
// 	   int offset = g_utf8_pointer_to_offset (layout.text, layout.text + index);
// 	   if (trailing)
// 		 {
// 		   for (index < line.start_index + line.length &&
// 			  offset + 1 < len(layout.text) &&
// 			  !layout.log_attrs[offset + 1].is_cursor_position)
// 		 {
// 		   offset++;
// 		   index = g_utf8_next_char (layout.text + index) - layout.text;
// 		 }
// 		 }
// 	   else
// 		 {
// 		   for (index > line.start_index &&
// 			  !layout.log_attrs[offset].is_cursor_position)
// 		 {
// 		   offset--;
// 		   index = g_utf8_prev_char (layout.text + index) - layout.text;
// 		 }

// 		 }

// 	   pango_glyph_string_index_to_x (run.glyphs,
// 					  layout.text + run.item.offset,
// 					  run.item.length,
// 					  &run.item.analysis,
// 					  index - run.item.offset, trailing, x_pos);
// 	   if (x_pos)
// 		 *x_pos += width;

// 	   return;
// 	 }

// 	   width += pango_glyph_string_get_width (run.glyphs);

// 	   run_list = run_list.next;
// 	 }

//    if (x_pos)
// 	 *x_pos = width;
//  }

//  static LayoutLine *
//  pango_layout_index_to_line (Layout      *layout,
// 				 int               index,
// 				 int              *line_nr,
// 				 LayoutLine **line_before,
// 				 LayoutLine **line_after)
//  {
//    GSList *tmp_list;
//    GSList *line_list;
//    line *LayoutLine = nil;
//    LayoutLine *prev_line = nil;
//    int i = -1;

//    line_list = tmp_list = layout.lines;
//    for (tmp_list)
// 	 {
// 	   LayoutLine *tmp_line = tmp_list.data;

// 	   if (tmp_line.start_index > index)
// 	 break; /* index was in paragraph delimiters */

// 	   prev_line = line;
// 	   line = tmp_line;
// 	   line_list = tmp_list;
// 	   i++;

// 	   if (line.start_index + line.length > index)
// 	 break;

// 	   tmp_list = tmp_list.next;
// 	 }

//    if (line_nr)
// 	 *line_nr = i;

//    if (line_before)
// 	 *line_before = prev_line;

//    if (line_after)
// 	 *line_after = (line_list && line_list.next) ? line_list.next.data : nil;

//    return line;
//  }

//  static LayoutLine *
//  pango_layout_index_to_line_and_extents (Layout     *layout,
// 					 int              index,
// 					 Rectangle  *line_rect)
//  {
//    LayoutIter iter;
//    line *LayoutLine = nil;

//    _pango_layout_get_iter (layout, &iter);

//    if (!ITER_IS_INVALID (&iter))
// 	 for (true)
// 	   {
// 	 LayoutLine *tmp_line = _pango_layout_iter_get_line (&iter);

// 	 if (tmp_line.start_index > index)
// 		 break; /* index was in paragraph delimiters */

// 	 line = tmp_line;

// 	 pango_layout_iter_get_line_extents (&iter, nil, line_rect);

// 	 if (line.start_index + line.length > index)
// 	   break;

// 	 if (!pango_layout_iter_next_line (&iter))
// 	   break; /* Use end of last line */
// 	   }

//    _pango_layout_iter_destroy (&iter);

//    return line;
//  }

//  /**
//   * pango_layout_index_to_line_x:
//   * @layout:    a #Layout
//   * @index_:    the byte index of a grapheme within the layout.
//   * @trailing:  an integer indicating the edge of the grapheme to retrieve the
//   *             position of. If > 0, the trailing edge of the grapheme, if 0,
//   *             the leading of the grapheme.
//   * @line: (out) (allow-none): location to store resulting line index. (which will
//   *               between 0 and pango_layout_get_line_count(layout) - 1), or %nil
//   * @x_pos: (out) (allow-none): location to store resulting position within line
//   *              (%pangoScale units per device unit), or %nil
//   *
//   * Converts from byte @index_ within the @layout to line and X position.
//   * (X position is measured from the left edge of the line)
//   */
//  void
//  pango_layout_index_to_line_x (Layout *layout,
// 				   int          index,
// 				   bool     trailing,
// 				   int         *line,
// 				   int         *x_pos)
//  {
//    int line_num;
//    LayoutLine *layout_line = nil;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (index >= 0);
//    g_return_if_fail (index <= layout.length);

//    pango_layout_check_lines (layout);

//    layout_line = pango_layout_index_to_line (layout, index,
// 						 &line_num, nil, nil);

//    if (layout_line)
// 	 {
// 	   /* use end of line if index was in the paragraph delimiters */
// 	   if (index > layout_line.start_index + layout_line.length)
// 	 index = layout_line.start_index + layout_line.length;

// 	   if (line)
// 	 *line = line_num;

// 	   pango_layout_line_index_to_x (layout_line, index, trailing, x_pos);
// 	 }
//    else
// 	 {
// 	   if (line)
// 	 *line = -1;
// 	   if (x_pos)
// 	 *x_pos = -1;
// 	 }
//  }

//  /**
//   * pango_layout_move_cursor_visually:
//   * @layout:       a #Layout.
//   * @strong:       whether the moving cursor is the strong cursor or the
//   *                weak cursor. The strong cursor is the cursor corresponding
//   *                to text insertion in the base direction for the layout.
//   * @old_index:    the byte index of the grapheme for the old index
//   * @old_trailing: if 0, the cursor was at the leading edge of the
//   *                grapheme indicated by @old_index, if > 0, the cursor
//   *                was at the trailing edge.
//   * @direction:    direction to move cursor. A negative
//   *                value indicates motion to the left.
//   * @new_index: (out): location to store the new cursor byte index. A value of -1
//   *                indicates that the cursor has been moved off the beginning
//   *                of the layout. A value of %G_MAXINT indicates that
//   *                the cursor has been moved off the end of the layout.
//   * @new_trailing: (out): number of characters to move forward from the
//   *                location returned for @new_index to get the position
//   *                where the cursor should be displayed. This allows
//   *                distinguishing the position at the beginning of one
//   *                line from the position at the end of the preceding
//   *                line. @new_index is always on the line where the
//   *                cursor should be displayed.
//   *
//   * Computes a new cursor position from an old position and
//   * a count of positions to move visually. If @direction is positive,
//   * then the new strong cursor position will be one position
//   * to the right of the old cursor position. If @direction is negative,
//   * then the new strong cursor position will be one position
//   * to the left of the old cursor position.
//   *
//   * In the presence of bidirectional text, the correspondence
//   * between logical and visual order will depend on the direction
//   * of the current run, and there may be jumps when the cursor
//   * is moved off of the end of a run.
//   *
//   * Motion here is in cursor positions, not in characters, so a
//   * single call to pango_layout_move_cursor_visually() may move the
//   * cursor over multiple characters when multiple characters combine
//   * to form a single grapheme.
//   **/
//  void
//  pango_layout_move_cursor_visually (Layout *layout,
// 					bool     strong,
// 					int          old_index,
// 					int          old_trailing,
// 					int          direction,
// 					int         *new_index,
// 					int         *new_trailing)
//  {
//    line *LayoutLine = nil;
//    LayoutLine *prev_line;
//    LayoutLine *next_line;

//    int *log2vis_map;
//    int *vis2log_map;
//    int n_vis;
//    int vis_pos, vis_pos_old, log_pos;
//    int start_offset;
//    bool off_start = false;
//    bool off_end = false;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (old_index >= 0 && old_index <= layout.length);
//    g_return_if_fail (old_index < layout.length || old_trailing == 0);
//    g_return_if_fail (new_index != nil);
//    g_return_if_fail (new_trailing != nil);

//    direction = (direction >= 0 ? 1 : -1);

//    pango_layout_check_lines (layout);

//    /* Find the line the old cursor is on */
//    line = pango_layout_index_to_line (layout, old_index,
// 					  nil, &prev_line, &next_line);

//    start_offset = g_utf8_pointer_to_offset (layout.text, layout.text + line.start_index);

//    for (old_trailing--)
// 	 old_index = g_utf8_next_char (layout.text + old_index) - layout.text;

//    log2vis_map = pango_layout_line_get_log2vis_map (line, strong);
//    n_vis = pango_utf8_strlen (layout.text + line.start_index, line.length);

//    /* Clamp old_index to fit on the line */
//    if (old_index > (line.start_index + line.length))
// 	 old_index = line.start_index + line.length;

//    vis_pos = log2vis_map[old_index - line.start_index];

//    g_free (log2vis_map);

//    /* Handling movement between lines */
//    if (vis_pos == 0 && direction < 0)
// 	 {
// 	   if (line.resolved_dir == PANGO_DIRECTION_LTR)
// 	 off_start = true;
// 	   else
// 	 off_end = true;
// 	 }
//    else if (vis_pos == n_vis && direction > 0)
// 	 {
// 	   if (line.resolved_dir == PANGO_DIRECTION_LTR)
// 	 off_end = true;
// 	   else
// 	 off_start = true;
// 	 }

//    if (off_start || off_end)
// 	 {
// 	   /* If we move over a paragraph boundary, count that as
// 		* an extra position in the motion
// 		*/
// 	   bool paragraph_boundary;

// 	   if (off_start)
// 	 {
// 	   if (!prev_line)
// 		 {
// 		   *new_index = -1;
// 		   *new_trailing = 0;
// 		   return;
// 		 }
// 	   line = prev_line;
// 	   paragraph_boundary = (line.start_index + line.length != old_index);
// 	 }
// 	   else
// 	 {
// 	   if (!next_line)
// 		 {
// 		   *new_index = G_MAXINT;
// 		   *new_trailing = 0;
// 		   return;
// 		 }
// 	   line = next_line;
// 	   paragraph_boundary = (line.start_index != old_index);
// 	 }

// 	   n_vis = pango_utf8_strlen (layout.text + line.start_index, line.length);
// 	   start_offset = g_utf8_pointer_to_offset (layout.text, layout.text + line.start_index);

// 	   if (vis_pos == 0 && direction < 0)
// 	 {
// 	   vis_pos = n_vis;
// 	   if (paragraph_boundary)
// 		 vis_pos++;
// 	 }
// 	   else /* (vis_pos == n_vis && direction > 0) */
// 	 {
// 	   vis_pos = 0;
// 	   if (paragraph_boundary)
// 		 vis_pos--;
// 	 }
// 	 }

//    vis2log_map = pango_layout_line_get_vis2log_map (line, strong);

//    vis_pos_old = vis_pos + direction;
//    log_pos = g_utf8_pointer_to_offset (layout.text + line.start_index,
// 					   layout.text + line.start_index + vis2log_map[vis_pos_old]);
//    do
// 	 {
// 	   vis_pos += direction;
// 	   log_pos += g_utf8_pointer_to_offset (layout.text + line.start_index + vis2log_map[vis_pos_old],
// 						layout.text + line.start_index + vis2log_map[vis_pos]);
// 	   vis_pos_old = vis_pos;
// 	 }
//    for (vis_pos > 0 && vis_pos < n_vis &&
// 	  !layout.log_attrs[start_offset + log_pos].is_cursor_position);

//    *new_index = line.start_index + vis2log_map[vis_pos];
//    g_free (vis2log_map);

//    *new_trailing = 0;

//    if (*new_index == line.start_index + line.length && line.length > 0)
// 	 {
// 	   do
// 	 {
// 	   log_pos--;
// 	   *new_index = g_utf8_prev_char (layout.text + *new_index) - layout.text;
// 	   (*new_trailing)++;
// 	 }
// 	   for (log_pos > 0 && !layout.log_attrs[start_offset + log_pos].is_cursor_position);
// 	 }
//  }

//  /**
//   * pango_layout_xy_to_index:
//   * @layout:    a #Layout
//   * @x:         the X offset (in Pango units)
//   *             from the left edge of the layout.
//   * @y:         the Y offset (in Pango units)
//   *             from the top edge of the layout
//   * @index_: (out):   location to store calculated byte index
//   * @trailing: (out): location to store a integer indicating where
//   *             in the grapheme the user clicked. It will either
//   *             be zero, or the number of characters in the
//   *             grapheme. 0 represents the leading edge of the grapheme.
//   *
//   * Converts from X and Y position within a layout to the byte
//   * index to the character at that logical position. If the
//   * Y position is not inside the layout, the closest position is chosen
//   * (the position will be clamped inside the layout). If the
//   * X position is not within the layout, then the start or the
//   * end of the line is chosen as described for pango_layout_line_x_to_index().
//   * If either the X or Y positions were not inside the layout, then the
//   * function returns %false; on an exact hit, it returns `true`.
//   *
//   * Return value: `true` if the coordinates were inside text, %false otherwise.
//   **/
//  bool
//  pango_layout_xy_to_index (Layout *layout,
// 			   int          x,
// 			   int          y,
// 			   int         *index,
// 			   gint        *trailing)
//  {
//    LayoutIter iter;
//    LayoutLine *prev_line = nil;
//    LayoutLine *found = nil;
//    int found_line_x = 0;
//    int prev_last = 0;
//    int prev_line_x = 0;
//    bool retval = false;
//    bool outside = false;

//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), false);

//    _pango_layout_get_iter (layout, &iter);

//    do
// 	 {
// 	   Rectangle line_logical;
// 	   int first_y, last_y;

// 	   assert (!ITER_IS_INVALID (&iter));

// 	   pango_layout_iter_get_line_extents (&iter, nil, &line_logical);
// 	   pango_layout_iter_get_line_yrange (&iter, &first_y, &last_y);

// 	   if (y < first_y)
// 	 {
// 	   if (prev_line && y < (prev_last + (first_y - prev_last) / 2))
// 		 {
// 		   found = prev_line;
// 		   found_line_x = prev_line_x;
// 		 }
// 	   else
// 		 {
// 		   if (prev_line == nil)
// 		 outside = true; /* off the top */

// 		   found = _pango_layout_iter_get_line (&iter);
// 		   found_line_x = x - line_logical.x;
// 		 }
// 	 }
// 	   else if (y >= first_y &&
// 			y < last_y)
// 	 {
// 	   found = _pango_layout_iter_get_line (&iter);
// 	   found_line_x = x - line_logical.x;
// 	 }

// 	   prev_line = _pango_layout_iter_get_line (&iter);
// 	   prev_last = last_y;
// 	   prev_line_x = x - line_logical.x;

// 	   if (found != nil)
// 	 break;
// 	 }
//    for (pango_layout_iter_next_line (&iter));

//    _pango_layout_iter_destroy (&iter);

//    if (found == nil)
// 	 {
// 	   /* Off the bottom of the layout */
// 	   outside = true;

// 	   found = prev_line;
// 	   found_line_x = prev_line_x;
// 	 }

//    retval = pango_layout_line_x_to_index (found,
// 					  found_line_x,
// 					  index, trailing);

//    if (outside)
// 	 retval = false;

//    return retval;
//  }

//  /**
//   * pango_layout_index_to_pos:
//   * @layout: a #Layout
//   * @index_: byte index within @layout
//   * @pos: (out): rectangle in which to store the position of the grapheme
//   *
//   * Converts from an index within a #Layout to the onscreen position
//   * corresponding to the grapheme at that index, which is represented
//   * as rectangle.  Note that <literal>pos.x</literal> is always the leading
//   * edge of the grapheme and <literal>pos.x + pos.width</literal> the trailing
//   * edge of the grapheme. If the directionality of the grapheme is right-to-left,
//   * then <literal>pos.width</literal> will be negative.
//   **/
//  void
//  pango_layout_index_to_pos (layout *Layout    ,
// 				int             index,
// 				Rectangle *pos)
//  {
//    Rectangle logical_rect;
//    LayoutIter iter;
//    LayoutLine *layout_line = nil;
//    int x_pos;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (index >= 0);
//    g_return_if_fail (pos != nil);

//    _pango_layout_get_iter (layout, &iter);

//    if (!ITER_IS_INVALID (&iter))
// 	 {
// 	   for (true)
// 	 {
// 	   LayoutLine *tmp_line = _pango_layout_iter_get_line (&iter);

// 	   if (tmp_line.start_index > index)
// 		 {
// 		   /* index is in the paragraph delim&iters, move to
// 			* end of previous line
// 			*
// 			* This shouldn’t occur in the first loop &iteration as the first
// 			* line’s start_index should always be 0.
// 			*/
// 		   assert (layout_line != nil);
// 		   index = layout_line.start_index + layout_line.length;
// 		   break;
// 		 }

// 	   layout_line = tmp_line;

// 	   pango_layout_iter_get_line_extents (&iter, nil, &logical_rect);

// 	   if (layout_line.start_index + layout_line.length > index)
// 		 break;

// 	   if (!pango_layout_iter_next_line (&iter))
// 		 {
// 		   index = layout_line.start_index + layout_line.length;
// 		   break;
// 		 }
// 	 }

// 	   pos.y = logical_rect.y;
// 	   pos.height = logical_rect.height;

// 	   pango_layout_line_index_to_x (layout_line, index, 0, &x_pos);
// 	   pos.x = logical_rect.x + x_pos;

// 	   if (index < layout_line.start_index + layout_line.length)
// 	 {
// 	   pango_layout_line_index_to_x (layout_line, index, 1, &x_pos);
// 	   pos.width = (logical_rect.x + x_pos) - pos.x;
// 	 }
// 	   else
// 	 pos.width = 0;
// 	 }

//    _pango_layout_iter_destroy (&iter);
//  }

//  static void
//  pango_layout_line_get_range (line *LayoutLine,
// 				  char           **start,
// 				  char           **end)
//  {
//    char *p;

//    p = line.layout.text + line.start_index;

//    if (start)
// 	 *start = p;
//    if (end)
// 	 *end = p + line.length;
//  }

//  static int *
//  pango_layout_line_get_vis2log_map (line *LayoutLine,
// 					bool         strong)
//  {
//    Layout *layout = line.layout;
//    PangoDirection prev_dir;
//    PangoDirection cursor_dir;
//    GSList *tmp_list;
//    gchar *start, *end;
//    int *result;
//    int pos;
//    int n_chars;

//    pango_layout_line_get_range (line, &start, &end);
//    n_chars = pango_utf8_strlen (start, end - start);

//    result = g_new (int, n_chars + 1);

//    if (strong)
// 	 cursor_dir = line.resolved_dir;
//    else
// 	 cursor_dir = (line.resolved_dir == PANGO_DIRECTION_LTR) ? PANGO_DIRECTION_RTL : PANGO_DIRECTION_LTR;

//    /* Handle the first visual position
// 	*/
//    if (line.resolved_dir == cursor_dir)
// 	 result[0] = line.resolved_dir == PANGO_DIRECTION_LTR ? 0 : end - start;

//    prev_dir = line.resolved_dir;
//    pos = 0;
//    tmp_list = line.runs;
//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;
// 	   int run_n_chars = run.item.num_chars;
// 	   PangoDirection run_dir = (run.item.analysis.level % 2) ? PANGO_DIRECTION_RTL : PANGO_DIRECTION_LTR;
// 	   char *p = layout.text + run.item.offset;
// 	   int i;

// 	   /* pos is the visual position at the start of the run */
// 	   /* p is the logical byte index at the start of the run */

// 	   if (run_dir == PANGO_DIRECTION_LTR)
// 	 {
// 	   if ((cursor_dir == PANGO_DIRECTION_LTR) ||
// 		   (prev_dir == run_dir))
// 		 result[pos] = p - start;

// 	   p = g_utf8_next_char (p);

// 	   for (i = 1; i < run_n_chars; i++)
// 		 {
// 		   result[pos + i] = p - start;
// 		   p = g_utf8_next_char (p);
// 		 }

// 	   if (cursor_dir == PANGO_DIRECTION_LTR)
// 		 result[pos + run_n_chars] = p - start;
// 	 }
// 	   else
// 	 {
// 	   if (cursor_dir == PANGO_DIRECTION_RTL)
// 		 result[pos + run_n_chars] = p - start;

// 	   p = g_utf8_next_char (p);

// 	   for (i = 1; i < run_n_chars; i++)
// 		 {
// 		   result[pos + run_n_chars - i] = p - start;
// 		   p = g_utf8_next_char (p);
// 		 }

// 	   if ((cursor_dir == PANGO_DIRECTION_RTL) ||
// 		   (prev_dir == run_dir))
// 		 result[pos] = p - start;
// 	 }

// 	   pos += run_n_chars;
// 	   prev_dir = run_dir;
// 	   tmp_list = tmp_list.next;
// 	 }

//    /* And the last visual position
// 	*/
//    if ((cursor_dir == line.resolved_dir) || (prev_dir == line.resolved_dir))
// 	 result[pos] = line.resolved_dir == PANGO_DIRECTION_LTR ? end - start : 0;

//    return result;
//  }

//  static int *
//  pango_layout_line_get_log2vis_map (line *LayoutLine,
// 					bool         strong)
//  {
//    gchar *start, *end;
//    int *reverse_map;
//    int *result;
//    int i;
//    int n_chars;

//    pango_layout_line_get_range (line, &start, &end);
//    n_chars = pango_utf8_strlen (start, end - start);
//    result = g_new0 (int, end - start + 1);

//    reverse_map = pango_layout_line_get_vis2log_map (line, strong);

//    for (i=0; i <= n_chars; i++)
// 	 result[reverse_map[i]] = i;

//    g_free (reverse_map);

//    return result;
//  }

//  static PangoDirection
//  pango_layout_line_get_char_direction (LayoutLine *layout_line,
// 					   int              index)
//  {
//    GSList *run_list;

//    run_list = layout_line.runs;
//    for (run_list)
// 	 {
// 	   GlyphItem *run = run_list.data;

// 	   if (run.item.offset <= index && run.item.offset + run.item.length > index)
// 	 return run.item.analysis.level % 2 ? PANGO_DIRECTION_RTL : PANGO_DIRECTION_LTR;

// 	   run_list = run_list.next;
// 	 }

//    assert_not_reached ();

//    return PANGO_DIRECTION_LTR;
//  }

//  /**
//   * pango_layout_get_direction:
//   * @layout: a #Layout
//   * @index: the byte index of the char
//   *
//   * Gets the text direction at the given character
//   * position in @layout.
//   *
//   * Returns: the text direction at @index
//   *
//   * Since: 1.46
//   */
//  PangoDirection
//  pango_layout_get_direction (Layout *layout,
// 							 int          index)
//  {
//    line *LayoutLine;

//    line = pango_layout_index_to_line_and_extents (layout, index, nil);

//    if (line)
// 	 return pango_layout_line_get_char_direction (line, index);

//    return PANGO_DIRECTION_LTR;
//  }

//  /**
//   * pango_layout_get_cursor_pos:
//   * @layout: a #Layout
//   * @index_: the byte index of the cursor
//   * @strong_pos: (out) (allow-none): location to store the strong cursor position
//   *                     (may be %nil)
//   * @weak_pos: (out) (allow-none): location to store the weak cursor position (may be %nil)
//   *
//   * Given an index within a layout, determines the positions that of the
//   * strong and weak cursors if the insertion point is at that
//   * index. The position of each cursor is stored as a zero-width
//   * rectangle. The strong cursor location is the location where
//   * characters of the directionality equal to the base direction of the
//   * layout are inserted.  The weak cursor location is the location
//   * where characters of the directionality opposite to the base
//   * direction of the layout are inserted.
//   **/
//  void
//  pango_layout_get_cursor_pos (layout *Layout    ,
// 				  int             index,
// 				  Rectangle *strong_pos,
// 				  Rectangle *weak_pos)
//  {
//    PangoDirection dir1;
//    Rectangle line_rect;
//    LayoutLine *layout_line = nil; /* Quiet GCC */
//    int x1_trailing;
//    int x2;

//    g_return_if_fail (layout != nil);
//    g_return_if_fail (index >= 0 && index <= layout.length);

//    layout_line = pango_layout_index_to_line_and_extents (layout, index,
// 							 &line_rect);

//    assert (index >= layout_line.start_index);

//    /* Examine the trailing edge of the character before the cursor */
//    if (index == layout_line.start_index)
// 	 {
// 	   dir1 = layout_line.resolved_dir;
// 	   if (layout_line.resolved_dir == PANGO_DIRECTION_LTR)
// 	 x1_trailing = 0;
// 	   else
// 	 x1_trailing = line_rect.width;
// 	 }
//    else if (index >= layout_line.start_index + layout_line.length)
// 	 {
// 	   dir1 = layout_line.resolved_dir;
// 	   if (layout_line.resolved_dir == PANGO_DIRECTION_LTR)
// 	 x1_trailing = line_rect.width;
// 	   else
// 	 x1_trailing = 0;
// 	 }
//    else
// 	 {
// 	   gint prev_index = g_utf8_prev_char (layout.text + index) - layout.text;
// 	   dir1 = pango_layout_line_get_char_direction (layout_line, prev_index);
// 	   pango_layout_line_index_to_x (layout_line, prev_index, true, &x1_trailing);
// 	 }

//    /* Examine the leading edge of the character after the cursor */
//    if (index >= layout_line.start_index + layout_line.length)
// 	 {
// 	   if (layout_line.resolved_dir == PANGO_DIRECTION_LTR)
// 	 x2 = line_rect.width;
// 	   else
// 	 x2 = 0;
// 	 }
//    else
// 	 {
// 	   pango_layout_line_index_to_x (layout_line, index, false, &x2);
// 	 }

//    if (strong_pos)
// 	 {
// 	   strong_pos.x = line_rect.x;

// 	   if (dir1 == layout_line.resolved_dir)
// 	 strong_pos.x += x1_trailing;
// 	   else
// 	 strong_pos.x += x2;

// 	   strong_pos.y = line_rect.y;
// 	   strong_pos.width = 0;
// 	   strong_pos.height = line_rect.height;
// 	 }

//    if (weak_pos)
// 	 {
// 	   weak_pos.x = line_rect.x;

// 	   if (dir1 == layout_line.resolved_dir)
// 	 weak_pos.x += x2;
// 	   else
// 	 weak_pos.x += x1_trailing;

// 	   weak_pos.y = line_rect.y;
// 	   weak_pos.width = 0;
// 	   weak_pos.height = line_rect.height;
// 	 }
//  }

//  static inline int
//  direction_simple (PangoDirection d)
//  {
//    switch (d)
// 	 {
// 	 case PANGO_DIRECTION_LTR :
// 	 case PANGO_DIRECTION_WEAK_LTR :
// 	 case PANGO_DIRECTION_TTB_RTL :
// 	   return 1;
// 	 case PANGO_DIRECTION_RTL :
// 	 case PANGO_DIRECTION_WEAK_RTL :
// 	 case PANGO_DIRECTION_TTB_LTR :
// 	   return -1;
// 	 case PANGO_DIRECTION_NEUTRAL :
// 	   return 0;
// 	 /* no default, compiler should complain if a new values is added */
// 	 }
//    /* not reached */
//    return 0;
//  }

//  static PangoAlignment
//  get_alignment (Layout     *layout,
// 			line *LayoutLine)
//  {
//    PangoAlignment alignment = layout.alignment;

//    if (alignment != PANGO_ALIGN_CENTER && line.layout.auto_dir &&
// 	   direction_simple (line.resolved_dir) ==
// 	   -direction_simple (pango_context_get_base_dir (layout.context)))
// 	 {
// 	   if (alignment == PANGO_ALIGN_LEFT)
// 	 alignment = PANGO_ALIGN_RIGHT;
// 	   else if (alignment == PANGO_ALIGN_RIGHT)
// 	 alignment = PANGO_ALIGN_LEFT;
// 	 }

//    return alignment;
//  }

//  static void
//  get_x_offset (Layout     *layout,
// 		   line *LayoutLine,
// 		   int              layout_width,
// 		   int              line_width,
// 		   int             *x_offset)
//  {
//    PangoAlignment alignment = get_alignment (layout, line);

//    /* Alignment */
//    if (layout_width == 0)
// 	 *x_offset = 0;
//    else if (alignment == PANGO_ALIGN_RIGHT)
// 	 *x_offset = layout_width - line_width;
//    else if (alignment == PANGO_ALIGN_CENTER) {
// 	 *x_offset = (layout_width - line_width) / 2;
// 	 /* hinting */
// 	 if (((layout_width | line_width) & (pangoScale - 1)) == 0)
// 	   {
// 	 *x_offset = PANGO_UNITS_ROUND (*x_offset);
// 	   }
//    } else
// 	 *x_offset = 0;

//    /* Indentation */

//    /* For center, we ignore indentation; I think I've seen word
// 	* processors that still do the indentation here as if it were
// 	* indented left/right, though we can't sensibly do that without
// 	* knowing whether left/right is the "normal" thing for this text
// 	*/

//    if (alignment == PANGO_ALIGN_CENTER)
// 	 return;

//    if (line.is_paragraph_start)
// 	 {
// 	   if (layout.indent > 0)
// 	 {
// 	   if (alignment == PANGO_ALIGN_LEFT)
// 		 *x_offset += layout.indent;
// 	   else
// 		 *x_offset -= layout.indent;
// 	 }
// 	 }
//    else
// 	 {
// 	   if (layout.indent < 0)
// 	 {
// 	   if (alignment == PANGO_ALIGN_LEFT)
// 		 *x_offset -= layout.indent;
// 	   else
// 		 *x_offset += layout.indent;
// 	 }
// 	 }
//  }

//  static void
//  pango_layout_line_get_extents_and_height (line *LayoutLine,
// 										   Rectangle *ink,
// 										   Rectangle *logical,
// 										   int            *height);
//  static void
//  get_line_extents_layout_coords (Layout     *layout,
// 				 line *LayoutLine,
// 				 int              layout_width,
// 				 int              y_offset,
// 				 int             *baseline,
// 				 Rectangle  *line_ink_layout,
// 				 Rectangle  *line_logical_layout)
//  {
//    int x_offset;
//    /* Line extents in line coords (origin at line baseline) */
//    Rectangle line_ink;
//    Rectangle line_logical;
//    bool first_line;
//    int new_baseline;
//    int height;

//    if (layout.lines.data == line)
// 	 first_line = true;
//    else
// 	 first_line = false;

//    pango_layout_line_get_extents_and_height (line, line_ink_layout ? &line_ink : nil,
// 											 &line_logical,
// 											 &height);

//    get_x_offset (layout, line, layout_width, line_logical.width, &x_offset);

//    if (first_line || !baseline || layout.line_spacing == 0.0)
// 	 new_baseline = y_offset - line_logical.y;
//    else
// 	 new_baseline = *baseline + layout.line_spacing * height;

//    /* Convert the line's extents into layout coordinates */
//    if (line_ink_layout)
// 	 {
// 	   *line_ink_layout = line_ink;
// 	   line_ink_layout.x = line_ink.x + x_offset;
// 	   line_ink_layout.y = new_baseline + line_ink.y;
// 	 }

//    if (line_logical_layout)
// 	 {
// 	   *line_logical_layout = line_logical;
// 	   line_logical_layout.x = line_logical.x + x_offset;
// 	   line_logical_layout.y = new_baseline + line_logical.y;
// 	 }

//    if (baseline)
// 	 *baseline = new_baseline;
//  }

//  /* if non-nil line_extents returns a list of line extents
//   * in layout coordinates
//   */
//  static void
//  pango_layout_get_extents_internal (layout *Layout    ,
// 					Rectangle *ink_rect,
// 					Rectangle *logical_rect,
// 									Extents        **line_extents)
//  {
//    GSList *line_list;
//    int y_offset = 0;
//    int width;
//    bool need_width = false;
//    int line_index = 0;
//    int baseline;

//    g_return_if_fail (layout != nil);

//    pango_layout_check_lines (layout);

//    if (ink_rect && layout.ink_rect_cached)
// 	 {
// 	   *ink_rect = layout.ink_rect;
// 	   ink_rect = nil;
// 	 }
//    if (logical_rect && layout.logical_rect_cached)
// 	 {
// 	   *logical_rect = layout.logical_rect;
// 	   logical_rect = nil;
// 	 }
//    if (!ink_rect && !logical_rect && !line_extents)
// 	 return;

//    /* When we are not wrapping, we need the overall width of the layout to
// 	* figure out the x_offsets of each line. However, we only need the
// 	* x_offsets if we are computing the ink_rect or individual line extents.
// 	*/
//    width = layout.width;

//    if (layout.auto_dir)
// 	 {
// 	   /* If one of the lines of the layout is not left aligned, then we need
// 		* the width of the layout to calculate line x-offsets; this requires
// 		* looping through the lines for layout.auto_dir.
// 		*/
// 	   line_list = layout.lines;
// 	   for (line_list && !need_width)
// 	 {
// 	   line *LayoutLine = line_list.data;

// 	   if (get_alignment (layout, line) != PANGO_ALIGN_LEFT)
// 		 need_width = true;

// 	   line_list = line_list.next;
// 	 }
// 	 }
//    else if (layout.alignment != PANGO_ALIGN_LEFT)
// 	 need_width = true;

//    if (width == -1 && need_width && (ink_rect || line_extents))
// 	 {
// 	   Rectangle overall_logical;

// 	   pango_layout_get_extents_internal (layout, nil, &overall_logical, nil);
// 	   width = overall_logical.width;
// 	 }

//    if (logical_rect)
// 	 {
// 	   logical_rect.x = 0;
// 	   logical_rect.y = 0;
// 	   logical_rect.width = 0;
// 	   logical_rect.height = 0;
// 	 }

//    if (line_extents && layout.line_count > 0)
// 	 {
// 	   *line_extents = g_malloc (sizeof (Extents) * layout.line_count);
// 	 }

//    baseline = 0;
//    line_list = layout.lines;
//    for (line_list)
// 	 {
// 	   line *LayoutLine = line_list.data;
// 	   /* Line extents in layout coords (origin at 0,0 of the layout) */
// 	   Rectangle line_ink_layout;
// 	   Rectangle line_logical_layout;

// 	   int new_pos;

// 	   /* This block gets the line extents in layout coords */
// 	   {
// 	 get_line_extents_layout_coords (layout, line,
// 					 width, y_offset,
// 					 &baseline,
// 					 ink_rect ? &line_ink_layout : nil,
// 					 &line_logical_layout);

// 	 if (line_extents && layout.line_count > 0)
// 	   {
// 		 Extents *ext = &(*line_extents)[line_index];
// 		 ext.baseline = baseline;
// 		 ext.ink_rect = line_ink_layout;
// 		 ext.logical_rect = line_logical_layout;
// 	   }
// 	   }

// 	   if (ink_rect)
// 	 {
// 	   /* Compute the union of the current ink_rect with
// 		* line_ink_layout
// 		*/

// 	   if (line_list == layout.lines)
// 		 {
// 		   *ink_rect = line_ink_layout;
// 		 }
// 	   else
// 		 {
// 		   new_pos = MIN (ink_rect.x, line_ink_layout.x);
// 		   ink_rect.width =
// 		 MAX (ink_rect.x + ink_rect.width,
// 			  line_ink_layout.x + line_ink_layout.width) - new_pos;
// 		   ink_rect.x = new_pos;

// 		   new_pos = MIN (ink_rect.y, line_ink_layout.y);
// 		   ink_rect.height =
// 		 MAX (ink_rect.y + ink_rect.height,
// 			  line_ink_layout.y + line_ink_layout.height) - new_pos;
// 		   ink_rect.y = new_pos;
// 		 }
// 	 }

// 	   if (logical_rect)
// 	 {
// 	   if (layout.width == -1)
// 		 {
// 		   /* When no width is set on layout, we can just compute the max of the
// 			* line lengths to get the horizontal extents ... logical_rect.x = 0.
// 			*/
// 		   logical_rect.width = MAX (logical_rect.width, line_logical_layout.width);
// 		 }
// 	   else
// 		 {
// 		   /* When a width is set, we have to compute the union of the horizontal
// 			* extents of all the lines.
// 			*/
// 		   if (line_list == layout.lines)
// 		 {
// 		   logical_rect.x = line_logical_layout.x;
// 		   logical_rect.width = line_logical_layout.width;
// 		 }
// 		   else
// 		 {
// 		   new_pos = MIN (logical_rect.x, line_logical_layout.x);
// 		   logical_rect.width =
// 			 MAX (logical_rect.x + logical_rect.width,
// 			  line_logical_layout.x + line_logical_layout.width) - new_pos;
// 		   logical_rect.x = new_pos;

// 		 }
// 		 }

// 	   logical_rect.height = line_logical_layout.y + line_logical_layout.height - logical_rect.y;
// 	 }

// 	   y_offset = line_logical_layout.y + line_logical_layout.height + layout.spacing;
// 	   line_list = line_list.next;
// 	   line_index ++;
// 	 }

//    if (ink_rect)
// 	 {
// 	   layout.ink_rect = *ink_rect;
// 	   layout.ink_rect_cached = true;
// 	 }
//    if (logical_rect)
// 	 {
// 	   layout.logical_rect = *logical_rect;
// 	   layout.logical_rect_cached = true;
// 	 }
//  }

//  /**
//   * pango_layout_get_extents:
//   * @layout:   a #Layout
//   * @ink_rect: (out) (allow-none): rectangle used to store the extents of the
//   *                   layout as drawn or %nil to indicate that the result is
//   *                   not needed.
//   * @logical_rect: (out) (allow-none):rectangle used to store the logical
//   *                      extents of the layout or %nil to indicate that the
//   *                      result is not needed.
//   *
//   * Computes the logical and ink extents of @layout. Logical extents
//   * are usually what you want for positioning things.  Note that both extents
//   * may have non-zero x and y.  You may want to use those to offset where you
//   * render the layout.  Not doing that is a very typical bug that shows up as
//   * right-to-left layouts not being correctly positioned in a layout with
//   * a set width.
//   *
//   * The extents are given in layout coordinates and in Pango units; layout
//   * coordinates begin at the top left corner of the layout.
//   */
//  void
//  pango_layout_get_extents (layout *Layout    ,
// 			   Rectangle *ink_rect,
// 			   Rectangle *logical_rect)
//  {
//    g_return_if_fail (layout != nil);

//    pango_layout_get_extents_internal (layout, ink_rect, logical_rect, nil);
//  }

//  /**
//   * pango_layout_get_pixel_extents:
//   * @layout:   a #Layout
//   * @ink_rect: (out) (allow-none): rectangle used to store the extents of the
//   *                   layout as drawn or %nil to indicate that the result is
//   *                   not needed.
//   * @logical_rect: (out) (allow-none): rectangle used to store the logical
//   *                       extents of the layout or %nil to indicate that the
//   *                       result is not needed.
//   *
//   * Computes the logical and ink extents of @layout in device units.
//   * This function just calls pango_layout_get_extents() followed by
//   * two pango_extents_to_pixels() calls, rounding @ink_rect and @logical_rect
//   * such that the rounded rectangles fully contain the unrounded one (that is,
//   * passes them as first argument to pango_extents_to_pixels()).
//   **/
//  void
//  pango_layout_get_pixel_extents (Layout *layout,
// 				 Rectangle *ink_rect,
// 				 Rectangle *logical_rect)
//  {
//    g_return_if_fail (PANGO_IS_LAYOUT (layout));

//    pango_layout_get_extents (layout, ink_rect, logical_rect);
//    pango_extents_to_pixels (ink_rect, nil);
//    pango_extents_to_pixels (logical_rect, nil);
//  }

//  /**
//   * pango_layout_get_size:
//   * @layout: a #Layout
//   * @width: (out) (allow-none): location to store the logical width, or %nil
//   * @height: (out) (allow-none): location to store the logical height, or %nil
//   *
//   * Determines the logical width and height of a #Layout
//   * in Pango units (device units scaled by %pangoScale). This
//   * is simply a convenience function around pango_layout_get_extents().
//   **/
//  void
//  pango_layout_get_size (Layout *layout,
// 				int         *width,
// 				int         *height)
//  {
//    Rectangle logical_rect;

//    pango_layout_get_extents (layout, nil, &logical_rect);

//    if (width)
// 	 *width = logical_rect.width;
//    if (height)
// 	 *height = logical_rect.height;
//  }

//  /**
//   * pango_layout_get_pixel_size:
//   * @layout: a #Layout
//   * @width: (out) (allow-none): location to store the logical width, or %nil
//   * @height: (out) (allow-none): location to store the logical height, or %nil
//   *
//   * Determines the logical width and height of a #Layout
//   * in device units. (pango_layout_get_size() returns the width
//   * and height scaled by %pangoScale.) This
//   * is simply a convenience function around
//   * pango_layout_get_pixel_extents().
//   **/
//  void
//  pango_layout_get_pixel_size (Layout *layout,
// 				  int         *width,
// 				  int         *height)
//  {
//    Rectangle logical_rect;

//    pango_layout_get_extents_internal (layout, nil, &logical_rect, nil);
//    pango_extents_to_pixels (&logical_rect, nil);

//    if (width)
// 	 *width = logical_rect.width;
//    if (height)
// 	 *height = logical_rect.height;
//  }

//  /**
//   * pango_layout_get_baseline:
//   * @layout: a #Layout
//   *
//   * Gets the Y position of baseline of the first line in @layout.
//   *
//   * Return value: baseline of first line, from top of @layout.
//   *
//   * Since: 1.22
//   **/
//  int
//  pango_layout_get_baseline (layout *Layout    )
//  {
//    int baseline;
//    Extents *extents = nil;

//    /* XXX this is kinda inefficient */
//    pango_layout_get_extents_internal (layout, nil, nil, &extents);
//    baseline = extents ? extents[0].baseline : 0;

//    g_free (extents);

//    return baseline;
//  }

//

//  static void
//  pango_layout_line_leaked (line *LayoutLine)
//  {
//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;

//    private.cache_status = LEAKED;

//    if (line.layout)
// 	 {
// 	   line.layout.logical_rect_cached = false;
// 	   line.layout.ink_rect_cached = false;
// 	 }
//  }

//
//  /*****************
//   * Line Breaking *
//   *****************/

//  static void shape_tab (LayoutLine  *line,
// 						PangoItem        *item,
// 				PangoGlyphString *glyphs);

//  static void
//  free_run (GlyphItem *run, gpointer data)
//  {
//    bool free_item = data != nil;
//    if (free_item)
// 	 pango_item_free (run.item);

//    pango_glyph_string_free (run.glyphs);
//    g_slice_free (GlyphItem, run);
//  }

// setup the cached value `tabWidth` if not already defined
func (layout *Layout) ensure_tab_width() {
	if layout.tabWidth != -1 {
		return
	}
	// Find out how wide 8 spaces are in the context's default
	// font. Utter performance killer. :-(
	glyphs := &GlyphString{}
	//    PangoItem *item;
	//    GList *items;
	//    PangoAttribute *attr;
	//    PangoAttrList *layout_attrs;
	var (
		language  Language
		tmp_attrs AttrList
	)
	font_desc := layout.context.font_desc // copy

	shape_flags := PANGO_SHAPE_NONE
	if layout.context.round_glyph_positions {
		shape_flags |= PANGO_SHAPE_ROUND_POSITIONS
	}

	layout_attrs := layout.pango_layout_get_effective_attributes()
	if layout_attrs != nil {
		iter := layout_attrs.pango_attr_list_get_iterator()
		iter.pango_attr_iterator_get_font(&font_desc, &language, nil)
	}

	attr := pango_attr_font_desc_new(font_desc)
	tmp_attrs.pango_attr_list_insert_before(attr)

	if language != "" {
		attr = pango_attr_language_new(language)
		tmp_attrs.pango_attr_list_insert_before(attr)
	}

	items := layout.context.pango_itemize([]rune{' '}, 0, 1, tmp_attrs, nil)

	item := items.data
	glyphs.pango_shape_with_flags([]rune("        "), []rune("        "), &item.analysis, shape_flags)

	layout.tabWidth = glyphs.pango_glyph_string_get_width()

	// We need to make sure the tabWidth is > 0 so finding tab positions
	// terminates. This check should be necessary only under extreme
	// problems with the font.
	if layout.tabWidth <= 0 {
		layout.tabWidth = 50 * pangoScale /* pretty much arbitrary */
	}
}

//  For now we only need the tab position, we assume
//  all tabs are left-aligned.
func (layout *Layout) get_tab_pos(index int) (int, bool) {
	var (
		nTabs     int
		inPixels  bool
		isDefault = true
	)

	if layout.tabs != nil {
		nTabs = len(layout.tabs.tabs)
		inPixels = layout.tabs.positions_in_pixels
		isDefault = false
	}

	if index < nTabs {
		_, pos := layout.tabs.pango_tab_array_get_tab(index)
		if inPixels {
			return pos * pangoScale, isDefault
		}
		return pos, isDefault
	}

	if nTabs > 0 {
		// Extrapolate tab position, repeating the last tab gap to infinity.

		_, lastPos := layout.tabs.pango_tab_array_get_tab(nTabs - 1)

		var nextToLastPos int
		if nTabs > 1 {
			_, nextToLastPos = layout.tabs.pango_tab_array_get_tab(nTabs - 2)
		}

		if inPixels {
			nextToLastPos *= pangoScale
			lastPos *= pangoScale
		}

		var tabWidth int
		if lastPos > nextToLastPos {
			tabWidth = lastPos - nextToLastPos
		} else {
			tabWidth = int(layout.tabWidth)
		}

		return lastPos + tabWidth*(index-nTabs+1), isDefault
	}
	// No tab array set, so use default tab width
	return int(layout.tabWidth) * index, isDefault
}

func (layout *Layout) canBreakAt(offset int, alwaysWrapChar bool) bool {
	// We probably should have a mode where we treat all white-space as
	// of fungible width - appropriate for typography but not for
	// editing.
	wrap := layout.wrap

	if wrap == PANGO_WRAP_WORD_CHAR {
		if alwaysWrapChar {
			wrap = PANGO_WRAP_CHAR
		} else {
			wrap = PANGO_WRAP_WORD
		}
	}

	if offset == len(layout.text) {
		return true
	} else if wrap == PANGO_WRAP_WORD {
		return layout.log_attrs[offset].IsLineBreak()
	} else if wrap == PANGO_WRAP_CHAR {
		return layout.log_attrs[offset].IsCharBreak()
	} else {
		if debugMode {
			log.Println("canBreakAt : broken Layout")
		}
		return true
	}
}

func (layout *Layout) canBreakIn(start_offset, num_chars int, allow_break_at_start bool) bool {
	i := 1
	if allow_break_at_start {
		i = 0
	}

	for ; i < num_chars; i++ {
		if layout.canBreakAt(start_offset+i, false) {
			return true
		}
	}

	return false
}

//  static inline void
//  distributeLetterSpacing (int  letter_spacing,
// 				int *space_left,
// 				int *space_right)
//  {
//    *space_left = letter_spacing / 2;
//    /* hinting */
//    if ((letter_spacing & (pangoScale - 1)) == 0)
// 	 {
// 	   *space_left = PANGO_UNITS_ROUND (*space_left);
// 	 }
//    *space_right = letter_spacing - *space_left;
//  }

type itemList struct {
	data *Item
	next *itemList
}

type ParaBreakState struct {
	/* maintained per layout */
	line_height      int /* Estimate of height of current line; < 0 is no estimate */
	remaining_height int /* Remaining height of the layout;  only defined if layout.height >= 0 */

	/* maintained per paragraph */
	attrs       AttrList  /* Attributes being used for itemization */
	items       *itemList /* This paragraph turned into items */
	base_dir    Direction /* Current resolved base direction */
	line_of_par int       /* Line of the paragraph, starting at 1 for first line */

	glyphs            *GlyphString   /* Glyphs for the first item in state.items */
	start_offset      int            /* Character offset of first item in state.items in layout.text */
	properties        ItemProperties /* Properties for the first item in state.items */
	log_widths        []GlyphUnit    /* Logical widths for first item in state.items.. */
	log_widths_offset int            /* Offset into log_widths to the point corresponding
	 * to the remaining portion of the first item */

	need_hyphen []bool /* Insert a hyphen if breaking here ? */
	// TODO: cleanup since line_start_index = line_start_offset
	line_start_index  int /* Start index of line in layout.text */
	line_start_offset int /* Character offset of line in layout.text */

	/* maintained per line */
	line_width      GlyphUnit /* Goal width of line currently processing; < 0 is infinite */
	remaining_width GlyphUnit /* Amount of space remaining on line; < 0 is infinite */

	hyphen_width GlyphUnit /* How much space a hyphen will take */
}

//  static bool
//  should_ellipsize_current_line (layout *Layout    ,
// 					ParaBreakState *state);

func (layout *Layout) break_needs_hyphen(state *ParaBreakState, pos int) bool {
	if state.log_widths_offset+pos == 0 {
		return false
	}

	if state.need_hyphen[state.log_widths_offset+pos-1] {
		return true
	}

	return false
}

func (layout *Layout) find_break_extra_width(state *ParaBreakState, pos int) GlyphUnit {
	// Check whether to insert a hyphen
	if layout.break_needs_hyphen(state, pos) {
		if state.hyphen_width < 0 {
			item := state.items.data
			state.hyphen_width = item.find_hyphen_width()
		}

		return state.hyphen_width
	}
	return 0
}

//  #if 0
//  # define DEBUG debug
//  void
//  debug (const char *where, line *LayoutLine, ParaBreakState *state)
//  {
//    int line_width = pango_layout_line_get_width (line);

//    g_debug ("rem %d + line %d = %d		%s",
// 		state.remaining_width,
// 		line_width,
// 		state.remaining_width + line_width,
// 		where);
//  }
//  #else
//  # define DEBUG(where, line, state) do { } for (0)
//  #endif

type BreakResult uint8

const (
	BREAK_NONE_FIT       BreakResult = iota // Couldn't fit anything.
	BREAK_SOME_FIT                          // The item was broken in the middle.
	BREAK_ALL_FIT                           // Everything fit.
	BREAK_EMPTY_FIT                         // Nothing fit, but that was ok, as we can break at the first char.
	BREAK_LINE_SEPARATOR                    // Item begins with a line separator.
)

// process_item tries to insert as much as possible of the first item of
// `state.items` onto `line`.
//
// If `forceFit` is `true`, then `BREAK_NONE_FIT` will never
// be returned, a run will be added even if inserting the minimum amount
// will cause the line to overflow. This is used at the start of a line
// and until we've found at least some place to break.
//
// If `noBreakAtEnd` is `true`, then `BREAK_ALL_FIT` will never be
// returned even if everything fits; the run will be broken earlier,
// or `BREAK_NONE_FIT` returned. This is used when the end of the
// run is not a break position.
func (layout *Layout) process_item(line *LayoutLine, state *ParaBreakState,
	forceFit bool, noBreakAtEnd bool) BreakResult {
	//    length int;
	//    i int;
	item := state.items.data
	shape_set := false
	processing_new_item := false

	if state.glyphs == nil {
		state.properties = item.pango_layout_get_item_properties()
		state.glyphs = line.shape_run(state, item)

		state.log_widths = nil
		state.need_hyphen = nil
		state.log_widths_offset = 0

		processing_new_item = true
	}

	if !layout.single_paragraph && layout.text[item.offset] == LINE_SEPARATOR &&
		!layout.should_ellipsize_current_line(state) {
		line.insert_run(state, item, true)
		state.log_widths_offset += item.num_chars

		return BREAK_LINE_SEPARATOR
	}

	if state.remaining_width < 0 && !noBreakAtEnd /* Wrapping off */ {
		line.insert_run(state, item, true)

		return BREAK_ALL_FIT
	}

	var width GlyphUnit
	if processing_new_item {
		width = state.glyphs.pango_glyph_string_get_width()
	} else {
		for _, w := range state.log_widths[state.log_widths_offset : state.log_widths_offset+item.num_chars] {
			width += w
		}
	}

	if (width <= state.remaining_width || (item.num_chars == 1 && line.runs == nil)) && !noBreakAtEnd {
		state.remaining_width -= width
		state.remaining_width = maxG(state.remaining_width, 0)
		line.insert_run(state, item, true)

		return BREAK_ALL_FIT
	} else {
		//    int num_chars = item.num_chars;
		//    int break_num_chars = num_chars;
		//    int orig_width = width;

		if processing_new_item {
			glyph_item := GlyphItem{item: item, glyphs: state.glyphs}
			state.log_widths = glyph_item.pango_glyph_item_get_logical_widths(layout.text)
			state.need_hyphen = item.get_need_hyphen(layout.text)
		}

	retry_break:

		// See how much of the item we can stuff in the line.
		width = 0
		var (
			break_width                    = width
			orig_width                     = width
			retrying_with_char_breaks      = false
			break_extra_width, extra_width GlyphUnit
			num_chars                      int
			break_num_chars                = item.num_chars
		)
		for num_chars = 0; num_chars < item.num_chars; num_chars++ {
			if width+extra_width > state.remaining_width && break_num_chars < item.num_chars {
				break
			}

			// If there are no previous runs we have to take care to grab at least one char.
			if layout.canBreakAt(state.start_offset+num_chars, retrying_with_char_breaks) &&
				(num_chars > 0 || line.runs != nil) {
				break_num_chars = num_chars
				break_width = width
				break_extra_width = extra_width

				extra_width = layout.find_break_extra_width(state, num_chars)
			} else {
				extra_width = 0
			}

			width += state.log_widths[state.log_widths_offset+num_chars]
		}

		// If there's a space at the end of the line, include that also.
		// The logic here should match zero_line_final_space().
		// XXX Currently it doesn't quite match the logic there.  We don't check
		// the cluster here.  But should be fine in practice.
		if break_num_chars > 0 && break_num_chars < item.num_chars &&
			layout.log_attrs[state.start_offset+break_num_chars-1].IsWhite() {
			break_width -= state.log_widths[state.log_widths_offset+break_num_chars-1]
		}

		if layout.wrap == PANGO_WRAP_WORD_CHAR && forceFit && break_width+break_extra_width > state.remaining_width && !retrying_with_char_breaks {
			retrying_with_char_breaks = true
			num_chars = item.num_chars
			width = orig_width
			break_num_chars = num_chars
			break_width = width
			goto retry_break
		}

		if forceFit || break_width+break_extra_width <= state.remaining_width /* Successfully broke the item */ {
			if state.remaining_width >= 0 {
				state.remaining_width -= break_width
				state.remaining_width = maxG(state.remaining_width, 0)
			}

			if break_num_chars == item.num_chars {
				if layout.break_needs_hyphen(state, break_num_chars) {
					item.analysis.flags |= PANGO_ANALYSIS_FLAG_NEED_HYPHEN
				}
				line.insert_run(state, item, true)

				return BREAK_ALL_FIT
			} else if break_num_chars == 0 {
				return BREAK_EMPTY_FIT
			} else {
				new_item := item.pango_item_split(break_num_chars)

				if layout.break_needs_hyphen(state, break_num_chars) {
					new_item.analysis.flags |= PANGO_ANALYSIS_FLAG_NEED_HYPHEN
				}
				/* Add the width back, to the line, reshape, subtract the new width */
				state.remaining_width += break_width
				line.insert_run(state, new_item, false)
				break_width = line.runs.data.glyphs.pango_glyph_string_get_width()
				state.remaining_width -= break_width

				state.log_widths_offset += break_num_chars

				// shaped items should never be broken
				if debugMode {
					assert(!shape_set)
				}

				return BREAK_SOME_FIT
			}
		} else {
			state.glyphs = nil
			state.log_widths = nil
			state.need_hyphen = nil
			return BREAK_NONE_FIT
		}
	}
}

func (layout *Layout) should_ellipsize_current_line(state *ParaBreakState) bool {
	if layout.ellipsize == PANGO_ELLIPSIZE_NONE || layout.width < 0 {
		return false
	}

	if layout.height >= 0 {
		/* state.remaining_height is height of layout left */

		/* if we can't stuff two more lines at the current guess of line height,
		* the line we are going to produce is going to be the last line */
		return state.line_height*2 > state.remaining_height
	} else {
		/* -layout.height is number of lines per paragraph to show */
		return state.line_of_par == -layout.height
	}
}

// the hard work begins here !
func (layout *Layout) process_line(state *ParaBreakState) {
	//    line *LayoutLine;
	var (
		haveBreak           = false   /* If we've seen a possible break yet */
		breakRemainingWidth GlyphUnit /* Remaining width before adding run with break */
		breakStartOffset    = 0       /* Start offset before adding run with break */
		breakLink           *runList  /* Link holding run before break */
		wrapped             = false   /* If we had to wrap the line */
	)

	line := layout.pango_layout_line_new()
	line.start_index = state.line_start_index
	line.is_paragraph_start = state.line_of_par == 1
	line.line_set_resolved_dir(state.base_dir)

	state.line_width = layout.width
	if state.line_width >= 0 && layout.alignment != PANGO_ALIGN_CENTER {
		if line.is_paragraph_start && layout.indent >= 0 {
			state.line_width -= layout.indent
		} else if !line.is_paragraph_start && layout.indent < 0 {
			state.line_width += layout.indent
		}
		if state.line_width < 0 {
			state.line_width = 0
		}
	}

	if layout.should_ellipsize_current_line(state) {
		state.remaining_width = -1
	} else {
		state.remaining_width = state.line_width
	}

	if debugMode {
		showDebug("starting to fill line", &line.LayoutLine, state)
	}

	for state.items != nil {
		item := state.items.data

		oldNumChars := item.num_chars
		oldRemainingWidth := state.remaining_width
		firstItemInLine := line.runs != nil

		result := layout.process_item(&line.LayoutLine, state, !haveBreak, false)

		switch result {
		case BREAK_ALL_FIT:
			if layout.canBreakIn(state.start_offset, oldNumChars, firstItemInLine) {
				haveBreak = true
				breakRemainingWidth = oldRemainingWidth
				breakStartOffset = state.start_offset
				breakLink = line.runs.next
			}
			state.items = state.items.next
			state.start_offset += oldNumChars
		case BREAK_EMPTY_FIT:
			wrapped = true
			goto done
		case BREAK_SOME_FIT:
			state.start_offset += oldNumChars - item.num_chars
			wrapped = true
			goto done
		case BREAK_NONE_FIT:
			/* Back up over unused runs to run where there is a break */
			for line.runs != nil && line.runs != breakLink {
				state.items = &itemList{data: line.uninsert_run(), next: state.items}
			}

			state.start_offset = breakStartOffset
			state.remaining_width = breakRemainingWidth

			/* Reshape run to break */
			item = state.items.data

			oldNumChars = item.num_chars
			result = layout.process_item(&line.LayoutLine, state, true, true)
			if debugMode {
				assert(result == BREAK_SOME_FIT || result == BREAK_EMPTY_FIT)
			}

			state.start_offset += oldNumChars - item.num_chars

			wrapped = true
			goto done
		case BREAK_LINE_SEPARATOR:
			state.items = state.items.next
			state.start_offset += oldNumChars
			// A line-separate is just a forced break. Set wrapped, so we do justification
			wrapped = true
			goto done
		}
	}

done:
	line.pango_layout_line_postprocess(state, wrapped)
	line.addLine(state)
	state.line_of_par++
	state.line_start_index += line.length
	state.line_start_offset = state.start_offset
}

// logAttrs must have length: length+1
// TODO: remove the `length` argument to avoid mistakes
func get_items_log_attrs(text []rune, start, length int, items *itemList, logAttrs []CharAttr) {

	pangoDefaultBreak(text[start:start+length], logAttrs)

	offset := 0
	for l := items; l != nil; l = l.next {
		item := l.data
		// item.offset <= start+length
		// item.length <= (start+length)-item.offset
		// TODO: check this if tailor break is implemented
		pango_tailor_break(text[item.offset:item.offset+item.num_chars],
			&item.analysis, item.offset, logAttrs[offset:offset+item.num_chars+1])
		offset += item.num_chars
	}
}

func apply_attributes_to_items(items *itemList, attrs AttrList) {
	if attrs == nil {
		return
	}

	iter := attrs.pango_attr_list_get_iterator()

	for l := items; l != nil; l = l.next {
		l.data.pango_item_apply_attrs(iter)
	}
}

func (layout *Layout) apply_attributes_to_runs(attrs AttrList) {
	if attrs == nil {
		return
	}

	for _, line := range layout.lines {
		old_runs := line.runs.reverse()
		line.runs = nil
		for rl := old_runs; rl != nil; rl = rl.next {
			glyph_item := rl.data

			new_runs := glyph_item.pango_glyph_item_apply_attrs(layout.text, attrs)

			line.runs = new_runs.concat(line.runs)
		}
	}
}

func (layout *Layout) pango_layout_check_lines() {
	var (
		itemizeAttrs, shapeAttrs AttrList
		iter                     AttrIterator
		state                    ParaBreakState
		prevBaseDir              = PANGO_DIRECTION_NEUTRAL
		baseDir                  = PANGO_DIRECTION_NEUTRAL
	)

	layout.check_context_changed()

	if len(layout.lines) != 0 {
		return
	}

	attrs := layout.pango_layout_get_effective_attributes()
	if attrs != nil {
		shapeAttrs = attrs.pango_attr_list_filter(affects_break_or_shape)
		itemizeAttrs = attrs.pango_attr_list_filter(affects_itemization)
		if itemizeAttrs != nil {
			iter = *itemizeAttrs.pango_attr_list_get_iterator()
		}
	}

	startOffset := 0
	start := 0 // index in text

	// Find the first strong direction of the text
	if layout.auto_dir {
		prevBaseDir = pango_find_base_dir(layout.text)
		if prevBaseDir == PANGO_DIRECTION_NEUTRAL {
			prevBaseDir = layout.context.base_dir
		}
	} else {
		baseDir = layout.context.base_dir
	}

	// these are only used if layout.height >= 0
	state.remaining_height = layout.height
	state.line_height = -1
	if layout.height >= 0 {
		var logical Rectangle
		layout.pango_layout_get_empty_extents_at_index(0, &logical)
		state.line_height = logical.height
	}

	layout.log_attrs = make([]CharAttr, len(layout.text)+1)
	for done := false; !done; {
		var (
			delimiter_index, next_para_index int
		)

		if layout.single_paragraph {
			delimiter_index = len(layout.text)
			next_para_index = len(layout.text)
		} else {
			delimiter_index, next_para_index = pango_find_paragraph_boundary(layout.text[start:])
		}

		if layout.auto_dir {
			baseDir = pango_find_base_dir(layout.text[start : start+delimiter_index])

			/* Propagate the base direction for neutral paragraphs */
			if baseDir == PANGO_DIRECTION_NEUTRAL {
				baseDir = prevBaseDir
			} else {
				prevBaseDir = baseDir
			}
		}

		end := start + delimiter_index // index into text

		delim_len := next_para_index - delimiter_index

		if end == len(layout.text) {
			done = true
		}

		if debugMode {
			assert(end <= len(layout.text))
			assert(start <= len(layout.text))
			assert(delim_len < 4) // PS is 3 bytes
			assert(delim_len >= 0)

		}

		var cachedIter *AttrIterator
		if itemizeAttrs != nil {
			cachedIter = &iter
		}
		state.attrs = itemizeAttrs
		state.items = layout.context.pango_itemize_with_base_dir(
			baseDir,
			layout.text,
			start, end-start,
			itemizeAttrs, cachedIter)

		apply_attributes_to_items(state.items, shapeAttrs)

		get_items_log_attrs(layout.text, start, delimiter_index+delim_len,
			state.items, layout.log_attrs[startOffset:])

		state.base_dir = baseDir
		state.line_of_par = 1
		state.start_offset = startOffset
		state.line_start_offset = startOffset
		state.line_start_index = start

		state.glyphs = nil
		state.log_widths = nil
		state.need_hyphen = nil

		/* for deterministic bug hunting's sake set everything! */
		state.line_width = -1
		state.remaining_width = -1
		state.log_widths_offset = 0

		state.hyphen_width = -1

		if state.items != nil {
			for state.items != nil {
				layout.process_line(&state)
			}
		} else {
			empty_line := layout.pango_layout_line_new()
			empty_line.start_index = state.line_start_index
			empty_line.is_paragraph_start = true
			empty_line.line_set_resolved_dir(baseDir)
			empty_line.addLine(&state)
		}

		if layout.height >= 0 && state.remaining_height < state.line_height {
			done = true
		}

		if !done {
			startOffset += (end - start) + delim_len
		}

		start = end + delim_len
	}

	layout.apply_attributes_to_runs(attrs)
	reverseLines(layout.lines)
}

func reverseLines(arr []*LayoutLine) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
}

func reverseItems(arr []*Item) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
}

//  /**
//   * pango_layout_line_ref:
//   * @line: (nullable): a #LayoutLine, may be %nil
//   *
//   * Increase the reference count of a #LayoutLine by one.
//   *
//   * Return value: the line passed in.
//   *
//   * Since: 1.10
//   **/
//  LayoutLine *
//  pango_layout_line_ref (line *LayoutLine)
//  {
//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;

//    if (line == nil)
// 	 return nil;

//    g_atomic_int_inc ((int *) &private.ref_count);

//    return line;
//  }

//  /**
//   * pango_layout_line_unref:
//   * @line: a #LayoutLine
//   *
//   * Decrease the reference count of a #LayoutLine by one.
//   * If the result is zero, the line and all associated memory
//   * will be freed.
//   **/
//  void
//  pango_layout_line_unref (line *LayoutLine)
//  {
//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;

//    if (line == nil)
// 	 return;

//    g_return_if_fail (private.ref_count > 0);

//    if (g_atomic_int_dec_and_test ((int *) &private.ref_count))
// 	 {
// 	   g_slist_foreach (line.runs, (GFunc)free_run, GINT_TO_POINTER (1));
// 	   g_slist_free (line.runs);
// 	   g_slice_free (LayoutLinePrivate, private);
// 	 }
//  }

//  G_DEFINE_BOXED_TYPE (LayoutLine, pango_layout_line,
// 					  pango_layout_line_ref,
// 					  pango_layout_line_unref);

//  /**
//   * pango_layout_line_x_to_index:
//   * @line:      a #LayoutLine
//   * @x_pos:     the X offset (in Pango units)
//   *             from the left edge of the line.
//   * @index_: (out):   location to store calculated byte index for
//   *                   the grapheme in which the user clicked.
//   * @trailing: (out): location to store an integer indicating where
//   *                   in the grapheme the user clicked. It will either
//   *                   be zero, or the number of characters in the
//   *                   grapheme. 0 represents the leading edge of the grapheme.
//   *
//   * Converts from x offset to the byte index of the corresponding
//   * character within the text of the layout. If @x_pos is outside the line,
//   * @index_ and @trailing will point to the very first or very last position
//   * in the line. This determination is based on the resolved direction
//   * of the paragraph; for example, if the resolved direction is
//   * right-to-left, then an X position to the right of the line (after it)
//   * results in 0 being stored in @index_ and @trailing. An X position to the
//   * left of the line results in @index_ pointing to the (logical) last
//   * grapheme in the line and @trailing being set to the number of characters
//   * in that grapheme. The reverse is true for a left-to-right line.
//   *
//   * Return value: %false if @x_pos was outside the line, `true` if inside
//   **/
//  bool
//  pango_layout_line_x_to_index (line *LayoutLine,
// 				   int              x_pos,
// 				   int             *index,
// 				   int             *trailing)
//  {
//    GSList *tmp_list;
//    gint start_pos = 0;
//    gint first_index = 0; /* line.start_index */
//    gint first_offset;
//    gint last_index;      /* start of last grapheme in line */
//    gint last_offset;
//    gint end_index;       /* end iterator for line */
//    gint end_offset;      /* end iterator for line */
//    Layout *layout;
//    gint last_trailing;
//    bool suppress_last_trailing;

//    g_return_val_if_fail (LINE_IS_VALID (line), false);

//    layout = line.layout;

//    /* Find the last index in the line
// 	*/
//    first_index = line.start_index;

//    if (line.length == 0)
// 	 {
// 	   if (index)
// 	 *index = first_index;
// 	   if (trailing)
// 	 *trailing = 0;

// 	   return false;
// 	 }

//    assert (line.length > 0);

//    first_offset = g_utf8_pointer_to_offset (layout.text, layout.text + line.start_index);

//    end_index = first_index + line.length;
//    end_offset = first_offset + g_utf8_pointer_to_offset (layout.text + first_index, layout.text + end_index);

//    last_index = end_index;
//    last_offset = end_offset;
//    last_trailing = 0;
//    do
// 	 {
// 	   last_index = g_utf8_prev_char (layout.text + last_index) - layout.text;
// 	   last_offset--;
// 	   last_trailing++;
// 	 }
//    for (last_offset > first_offset && !layout.log_attrs[last_offset].is_cursor_position);

//    /* This is a HACK. If a program only keeps track of cursor (etc)
// 	* indices and not the trailing flag, then the trailing index of the
// 	* last character on a wrapped line is identical to the leading
// 	* index of the next line. So, we fake it and set the trailing flag
// 	* to zero.
// 	*
// 	* That is, if the text is "now is the time", and is broken between
// 	* 'now' and 'is'
// 	*
// 	* Then when the cursor is actually at:
// 	*
// 	* n|o|w| |i|s|
// 	*              ^
// 	* we lie and say it is at:
// 	*
// 	* n|o|w| |i|s|
// 	*            ^
// 	*
// 	* So the cursor won't appear on the next line before 'the'.
// 	*
// 	* Actually, any program keeping cursor
// 	* positions with wrapped lines should distinguish leading and
// 	* trailing cursors.
// 	*/
//    tmp_list = layout.lines;
//    for (tmp_list.data != line)
// 	 tmp_list = tmp_list.next;

//    if (tmp_list.next &&
// 	   line.start_index + line.length == ((LayoutLine *)tmp_list.next.data).start_index)
// 	 suppress_last_trailing = true;
//    else
// 	 suppress_last_trailing = false;

//    if (x_pos < 0)
// 	 {
// 	   /* pick the leftmost char */
// 	   if (index)
// 	 *index = (line.resolved_dir == PANGO_DIRECTION_LTR) ? first_index : last_index;
// 	   /* and its leftmost edge */
// 	   if (trailing)
// 	 *trailing = (line.resolved_dir == PANGO_DIRECTION_LTR || suppress_last_trailing) ? 0 : last_trailing;

// 	   return false;
// 	 }

//    tmp_list = line.runs;
//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;
// 	   int logical_width;

// 	   logical_width = pango_glyph_string_get_width (run.glyphs);

// 	   if (x_pos >= start_pos && x_pos < start_pos + logical_width)
// 	 {
// 	   int offset;
// 	   bool char_trailing;
// 	   int grapheme_start_index;
// 	   int grapheme_start_offset;
// 	   int grapheme_end_offset;
// 	   int pos;
// 	   int char_index;

// 	   pango_glyph_string_x_to_index (run.glyphs,
// 					  layout.text + run.item.offset, run.item.length,
// 					  &run.item.analysis,
// 					  x_pos - start_pos,
// 					  &pos, &char_trailing);

// 	   char_index = run.item.offset + pos;

// 	   /* Convert from characters to graphemes */

// 	   offset = g_utf8_pointer_to_offset (layout.text, layout.text + char_index);

// 	   grapheme_start_offset = offset;
// 	   grapheme_start_index = char_index;
// 	   for (grapheme_start_offset > first_offset &&
// 		  !layout.log_attrs[grapheme_start_offset].is_cursor_position)
// 		 {
// 		   grapheme_start_index = g_utf8_prev_char (layout.text + grapheme_start_index) - layout.text;
// 		   grapheme_start_offset--;
// 		 }

// 	   grapheme_end_offset = offset;
// 	   do
// 		 {
// 		   grapheme_end_offset++;
// 		 }
// 	   for (grapheme_end_offset < end_offset &&
// 		  !layout.log_attrs[grapheme_end_offset].is_cursor_position);

// 	   if (index)
// 		 *index = grapheme_start_index;

// 	   if (trailing)
// 		 {
// 		   if ((grapheme_end_offset == end_offset && suppress_last_trailing) ||
// 		   offset + char_trailing <= (grapheme_start_offset + grapheme_end_offset) / 2)
// 		 *trailing = 0;
// 		   else
// 		 *trailing = grapheme_end_offset - grapheme_start_offset;
// 		 }

// 	   return true;
// 	 }

// 	   start_pos += logical_width;
// 	   tmp_list = tmp_list.next;
// 	 }

//    /* pick the rightmost char */
//    if (index)
// 	 *index = (line.resolved_dir == PANGO_DIRECTION_LTR) ? last_index : first_index;

//    /* and its rightmost edge */
//    if (trailing)
// 	 *trailing = (line.resolved_dir == PANGO_DIRECTION_LTR && !suppress_last_trailing) ? last_trailing : 0;

//    return false;
//  }

//  static int
//  pango_layout_line_get_width (line *LayoutLine)
//  {
//    int width = 0;
//    GSList *tmp_list = line.runs;

//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;

// 	   width += pango_glyph_string_get_width (run.glyphs);

// 	   tmp_list = tmp_list.next;
// 	 }

//    return width;
//  }

//  /**
//   * pango_layout_line_get_x_ranges:
//   * @line:        a #LayoutLine
//   * @start_index: Start byte index of the logical range. If this value
//   *               is less than the start index for the line, then
//   *               the first range will extend all the way to the leading
//   *               edge of the layout. Otherwise it will start at the
//   *               leading edge of the first character.
//   * @end_index:   Ending byte index of the logical range. If this value
//   *               is greater than the end index for the line, then
//   *               the last range will extend all the way to the trailing
//   *               edge of the layout. Otherwise, it will end at the
//   *               trailing edge of the last character.
//   * @ranges: (out) (array length=n_ranges) (transfer full):
//   *               location to store a pointer to an array of ranges.
//   *               The array will be of length <literal>2*n_ranges</literal>,
//   *               with each range starting at <literal>(*ranges)[2*n]</literal>
//   *               and of width <literal>(*ranges)[2*n + 1] - (*ranges)[2*n]</literal>.
//   *               This array must be freed with g_free(). The coordinates are relative
//   *               to the layout and are in Pango units.
//   * @n_ranges: The number of ranges stored in @ranges.
//   *
//   * Gets a list of visual ranges corresponding to a given logical range.
//   * This list is not necessarily minimal - there may be consecutive
//   * ranges which are adjacent. The ranges will be sorted from left to
//   * right. The ranges are with respect to the left edge of the entire
//   * layout, not with respect to the line.
//   **/
//  void
//  pango_layout_line_get_x_ranges (LayoutLine  *line,
// 				 int               start_index,
// 				 int               end_index,
// 				 int             **ranges,
// 				 int              *n_ranges)
//  {
//    gint line_start_index = 0;
//    GSList *tmp_list;
//    int range_count = 0;
//    int accumulated_width = 0;
//    int x_offset;
//    int width, line_width;
//    PangoAlignment alignment;

//    g_return_if_fail (line != nil);
//    g_return_if_fail (line.layout != nil);
//    g_return_if_fail (start_index <= end_index);

//    alignment = get_alignment (line.layout, line);

//    width = line.layout.width;
//    if (width == -1 && alignment != PANGO_ALIGN_LEFT)
// 	 {
// 	   Rectangle logical_rect;
// 	   pango_layout_get_extents (line.layout, nil, &logical_rect);
// 	   width = logical_rect.width;
// 	 }

//    /* FIXME: The computations here could be optimized, by moving the
// 	* computations of the x_offset after we go through and figure
// 	* out where each range is.
// 	*/

//    {
// 	 Rectangle logical_rect;
// 	 pango_layout_line_get_extents (line, nil, &logical_rect);
// 	 line_width = logical_rect.width;
//    }

//    get_x_offset (line.layout, line, width, line_width, &x_offset);

//    line_start_index = line.start_index;

//    /* Allocate the maximum possible size */
//    if (ranges)
// 	 *ranges = g_new (int, 2 * (2 + g_slist_length (line.runs)));

//    if (x_offset > 0 &&
// 	   ((line.resolved_dir == PANGO_DIRECTION_LTR && start_index < line_start_index) ||
// 		(line.resolved_dir == PANGO_DIRECTION_RTL && end_index > line_start_index + line.length)))
// 	 {
// 	   if (ranges)
// 	 {
// 	   (*ranges)[2*range_count] = 0;
// 	   (*ranges)[2*range_count + 1] = x_offset;
// 	 }

// 	   range_count ++;
// 	 }

//    tmp_list = line.runs;
//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = (GlyphItem *)tmp_list.data;

// 	   if ((start_index < run.item.offset + run.item.length &&
// 		end_index > run.item.offset))
// 	 {
// 	   if (ranges)
// 		 {
// 		   int run_start_index = MAX (start_index, run.item.offset);
// 		   int run_end_index = MIN (end_index, run.item.offset + run.item.length);
// 		   int run_start_x, run_end_x;

// 		   assert (run_end_index > 0);

// 		   /* Back the end_index off one since we want to find the trailing edge of the preceding character */

// 		   run_end_index = g_utf8_prev_char (line.layout.text + run_end_index) - line.layout.text;

// 		   pango_glyph_string_index_to_x (run.glyphs,
// 						  line.layout.text + run.item.offset,
// 						  run.item.length,
// 						  &run.item.analysis,
// 						  run_start_index - run.item.offset, false,
// 						  &run_start_x);
// 		   pango_glyph_string_index_to_x (run.glyphs,
// 						  line.layout.text + run.item.offset,
// 						  run.item.length,
// 						  &run.item.analysis,
// 						  run_end_index - run.item.offset, true,
// 						  &run_end_x);

// 		   (*ranges)[2*range_count] = x_offset + accumulated_width + MIN (run_start_x, run_end_x);
// 		   (*ranges)[2*range_count + 1] = x_offset + accumulated_width + MAX (run_start_x, run_end_x);
// 		 }

// 	   range_count++;
// 	 }

// 	   if (tmp_list.next)
// 	 accumulated_width += pango_glyph_string_get_width (run.glyphs);

// 	   tmp_list = tmp_list.next;
// 	 }

//    if (x_offset + line_width < line.layout.width &&
// 	   ((line.resolved_dir == PANGO_DIRECTION_LTR && end_index > line_start_index + line.length) ||
// 		(line.resolved_dir == PANGO_DIRECTION_RTL && start_index < line_start_index)))
// 	 {
// 	   if (ranges)
// 	 {
// 	   (*ranges)[2*range_count] = x_offset + line_width;
// 	   (*ranges)[2*range_count + 1] = line.layout.width;
// 	 }

// 	   range_count ++;
// 	 }

//    if (n_ranges)
// 	 *n_ranges = range_count;
//  }

//  static void
//  pango_layout_line_get_extents_and_height (line *LayoutLine,
// 										   Rectangle  *ink_rect,
// 										   Rectangle  *logical_rect,
// 										   int             *height)
//  {
//    LayoutLinePrivate *private = (LayoutLinePrivate *)line;
//    GSList *tmp_list;
//    int x_pos = 0;
//    bool caching = false;

//    g_return_if_fail (LINE_IS_VALID (line));

//    if (G_UNLIKELY (!ink_rect && !logical_rect))
// 	 return;

//    switch (private.cache_status)
// 	 {
// 	 case CACHED:
// 	   {
// 	 if (ink_rect)
// 	   *ink_rect = private.ink_rect;
// 	 if (logical_rect)
// 	   *logical_rect = private.logical_rect;
// 		 if (height)
// 		   *height = private.height;
// 	 return;
// 	   }
// 	 case NOT_CACHED:
// 	   {
// 	 caching = true;
// 	 if (!ink_rect)
// 	   ink_rect = &private.ink_rect;
// 	 if (!logical_rect)
// 	   logical_rect = &private.logical_rect;
// 		 if (!height)
// 		   height = &private.height;
// 	 break;
// 	   }
// 	 case LEAKED:
// 	   {
// 	 break;
// 	   }
// 	 }

//    if (ink_rect)
// 	 {
// 	   ink_rect.x = 0;
// 	   ink_rect.y = 0;
// 	   ink_rect.width = 0;
// 	   ink_rect.height = 0;
// 	 }

//    if (logical_rect)
// 	 {
// 	   logical_rect.x = 0;
// 	   logical_rect.y = 0;
// 	   logical_rect.width = 0;
// 	   logical_rect.height = 0;
// 	 }

//    if (height)
// 	 *height = 0;

//    tmp_list = line.runs;
//    for (tmp_list)
// 	 {
// 	   GlyphItem *run = tmp_list.data;
// 	   int new_pos;
// 	   Rectangle run_ink;
// 	   Rectangle run_logical;
// 	   int run_height;

// 	   pango_layout_run_get_extents_and_height (run,
// 												ink_rect ? &run_ink : nil,
// 												&run_logical,
// 												height ? &run_height : nil);

// 	   if (ink_rect)
// 	 {
// 	   if (ink_rect.width == 0 || ink_rect.height == 0)
// 		 {
// 		   *ink_rect = run_ink;
// 		   ink_rect.x += x_pos;
// 		 }
// 	   else if (run_ink.width != 0 && run_ink.height != 0)
// 		 {
// 		   new_pos = MIN (ink_rect.x, x_pos + run_ink.x);
// 		   ink_rect.width = MAX (ink_rect.x + ink_rect.width,
// 					  x_pos + run_ink.x + run_ink.width) - new_pos;
// 		   ink_rect.x = new_pos;

// 		   new_pos = MIN (ink_rect.y, run_ink.y);
// 		   ink_rect.height = MAX (ink_rect.y + ink_rect.height,
// 					   run_ink.y + run_ink.height) - new_pos;
// 		   ink_rect.y = new_pos;
// 		 }
// 	 }

// 	   if (logical_rect)
// 		 {
// 		   new_pos = MIN (logical_rect.x, x_pos + run_logical.x);
// 		   logical_rect.width = MAX (logical_rect.x + logical_rect.width,
// 					  x_pos + run_logical.x + run_logical.width) - new_pos;
// 	   logical_rect.x = new_pos;

// 	   new_pos = MIN (logical_rect.y, run_logical.y);
// 	   logical_rect.height = MAX (logical_rect.y + logical_rect.height,
// 					   run_logical.y + run_logical.height) - new_pos;
// 	   logical_rect.y = new_pos;
// 		 }

// 	   if (height)
// 		 *height = MAX (*height, run_height);

// 	   x_pos += run_logical.width;
// 	   tmp_list = tmp_list.next;
// 	 }

//    if (logical_rect && !line.runs)
// 	 pango_layout_line_get_empty_extents (line, logical_rect);

//    if (caching)
// 	 {
// 	   if (&private.ink_rect != ink_rect)
// 	 private.ink_rect = *ink_rect;
// 	   if (&private.logical_rect != logical_rect)
// 	 private.logical_rect = *logical_rect;
// 	   if (&private.height != height)
// 		 private.height = *height;
// 	   private.cache_status = CACHED;
// 	 }
//  }

//  /**
//   * pango_layout_line_get_extents:
//   * @line:     a #LayoutLine
//   * @ink_rect: (out) (allow-none): rectangle used to store the extents of
//   *            the glyph string as drawn, or %nil
//   * @logical_rect: (out) (allow-none):rectangle used to store the logical
//   *                extents of the glyph string, or %nil
//   *
//   * Computes the logical and ink extents of a layout line. See
//   * pango_font_get_glyph_extents() for details about the interpretation
//   * of the rectangles.
//   */
//  void
//  pango_layout_line_get_extents (line *LayoutLine,
// 					Rectangle  *ink_rect,
// 					Rectangle  *logical_rect)
//  {
//    pango_layout_line_get_extents_and_height (line, ink_rect, logical_rect, nil);
//  }

//  /**
//   * pango_layout_line_get_height:
//   * @line:     a #LayoutLine
//   * @height: (out) (allow-none): return location for the line height
//   *
//   * Computes the height of the line, ie the distance between
//   * this and the previous lines baseline.
//   *
//   * Since: 1.44
//   */
//  void
//  pango_layout_line_get_height (line *LayoutLine,
// 				   int             *height)
//  {
//    pango_layout_line_get_extents_and_height (line, nil, nil, height);
//  }

//  /**
//   * pango_layout_line_get_pixel_extents:
//   * @layout_line: a #LayoutLine
//   * @ink_rect: (out) (allow-none): rectangle used to store the extents of
//   *                   the glyph string as drawn, or %nil
//   * @logical_rect: (out) (allow-none): rectangle used to store the logical
//   *                       extents of the glyph string, or %nil
//   *
//   * Computes the logical and ink extents of @layout_line in device units.
//   * This function just calls pango_layout_line_get_extents() followed by
//   * two pango_extents_to_pixels() calls, rounding @ink_rect and @logical_rect
//   * such that the rounded rectangles fully contain the unrounded one (that is,
//   * passes them as first argument to pango_extents_to_pixels()).
//   **/
//  void
//  pango_layout_line_get_pixel_extents (LayoutLine *layout_line,
// 					  Rectangle  *ink_rect,
// 					  Rectangle  *logical_rect)
//  {
//    g_return_if_fail (LINE_IS_VALID (layout_line));

//    pango_layout_line_get_extents (layout_line, ink_rect, logical_rect);
//    pango_extents_to_pixels (ink_rect, nil);
//    pango_extents_to_pixels (logical_rect, nil);
//  }

//  static int
//  get_item_letter_spacing (PangoItem *item)
//  {
//    ItemProperties properties;

//    pango_layout_get_item_properties (item, &properties);

//    return properties.letter_spacing;
//  }

//  static void
//  pad_glyphstring_right (PangoGlyphString *glyphs,
// 				ParaBreakState   *state,
// 				int               adjustment)
//  {
//    int glyph = glyphs.num_glyphs - 1;

//    for (glyph >= 0 && glyphs.glyphs[glyph].geometry.width == 0)
// 	 glyph--;

//    if (glyph < 0)
// 	 return;

//    state.remaining_width -= adjustment;
//    glyphs.glyphs[glyph].geometry.width += adjustment;
//    if (glyphs.glyphs[glyph].geometry.width < 0)
// 	 {
// 	   state.remaining_width += glyphs.glyphs[glyph].geometry.width;
// 	   glyphs.glyphs[glyph].geometry.width = 0;
// 	 }
//  }

//  static void
//  pad_glyphstring_left (PangoGlyphString *glyphs,
// 			   ParaBreakState   *state,
// 			   int               adjustment)
//  {
//    int glyph = 0;

//    for (glyph < glyphs.num_glyphs && glyphs.glyphs[glyph].geometry.width == 0)
// 	 glyph++;

//    if (glyph == glyphs.num_glyphs)
// 	 return;

//    state.remaining_width -= adjustment;
//    glyphs.glyphs[glyph].geometry.width += adjustment;
//    glyphs.glyphs[glyph].geometry.x_offset += adjustment;
//  }

func (layout *Layout) is_tab_run(run *GlyphItem) bool {
	return layout.text[run.item.offset] == '\t'
}

//  /* When doing shaping, we add the letter spacing value for a
//   * run after every grapheme in the run. This produces ugly
//   * asymmetrical results, so what this routine is redistributes
//   * that space to the beginning and the end of the run.
//   *
//   * We also trim the letter spacing from runs adjacent to
//   * tabs and from the outside runs of the lines so that things
//   * line up properly. The line breaking and tab positioning
//   * were computed without this trimming so they are no longer
//   * exactly correct, but this won't be very noticeable in most
//   * cases.
//   */
//  static void
//  adjust_line_letter_spacing (line *LayoutLine,
// 				 state *ParaBreakState )
//  {
//    Layout *layout = line.layout;
//    bool reversed;
//    GlyphItem *last_run;
//    int tab_adjustment;
//    GSList *l;

//    /* If we have tab stops and the resolved direction of the
// 	* line is RTL, then we need to walk through the line
// 	* in reverse direction to figure out the corrections for
// 	* tab stops.
// 	*/
//    reversed = false;
//    if (line.resolved_dir == PANGO_DIRECTION_RTL)
// 	 {
// 	   for (l = line.runs; l; l = l.next)
// 	 if (is_tab_run (layout, l.data))
// 	   {
// 		 line.runs = g_slist_reverse (line.runs);
// 		 reversed = true;
// 		 break;
// 	   }
// 	 }

//    /* Walk over the runs in the line, redistributing letter
// 	* spacing from the end of the run to the start of the
// 	* run and trimming letter spacing from the ends of the
// 	* runs adjacent to the ends of the line or tab stops.
// 	*
// 	* We accumulate a correction factor from this trimming
// 	* which we add onto the next tab stop space to keep the
// 	* things properly aligned.
// 	*/

//    last_run = nil;
//    tab_adjustment = 0;
//    for (l = line.runs; l; l = l.next)
// 	 {
// 	   GlyphItem *run = l.data;
// 	   GlyphItem *next_run = l.next ? l.next.data : nil;

// 	   if (is_tab_run (layout, run))
// 	 {
// 	   pad_glyphstring_right (run.glyphs, state, tab_adjustment);
// 	   tab_adjustment = 0;
// 	 }
// 	   else
// 	 {
// 	   GlyphItem *visual_next_run = reversed ? last_run : next_run;
// 	   GlyphItem *visual_last_run = reversed ? next_run : last_run;
// 	   int run_spacing = get_item_letter_spacing (run.item);
// 	   int space_left, space_right;

// 	   distributeLetterSpacing (run_spacing, &space_left, &space_right);

// 	   if (run.glyphs.glyphs[0].geometry.width == 0)
// 		 {
// 		   /* we've zeroed this space glyph at the end of line, now remove
// 			* the letter spacing added to its adjacent glyph */
// 		   pad_glyphstring_left (run.glyphs, state, - space_left);
// 		 }
// 	   else if (!visual_last_run || is_tab_run (layout, visual_last_run))
// 		 {
// 		   pad_glyphstring_left (run.glyphs, state, - space_left);
// 		   tab_adjustment += space_left;
// 		 }

// 	   if (run.glyphs.glyphs[run.glyphs.num_glyphs - 1].geometry.width == 0)
// 		 {
// 		   /* we've zeroed this space glyph at the end of line, now remove
// 			* the letter spacing added to its adjacent glyph */
// 		   pad_glyphstring_right (run.glyphs, state, - space_right);
// 		 }
// 	   else if (!visual_next_run || is_tab_run (layout, visual_next_run))
// 		 {
// 		   pad_glyphstring_right (run.glyphs, state, - space_right);
// 		   tab_adjustment += space_right;
// 		 }
// 	 }

// 	   last_run = run;
// 	 }

//    if (reversed)
// 	 line.runs = g_slist_reverse (line.runs);
//  }

//  static void
//  justify_clusters (line *LayoutLine,
// 		   state *ParaBreakState )
//  {
//    const gchar *text = line.layout.text;
//    const PangoLogAttr *log_attrs = line.layout.log_attrs;

//    int total_remaining_width, total_gaps = 0;
//    int added_so_far, gaps_so_far;
//    bool is_hinted;
//    GSList *run_iter;
//    enum {
// 	 MEASURE,
// 	 ADJUST
//    } mode;

//    total_remaining_width = state.remaining_width;
//    if (total_remaining_width <= 0)
// 	 return;

//    /* hint to full pixel if total remaining width was so */
//    is_hinted = (total_remaining_width & (pangoScale - 1)) == 0;

//    for (mode = MEASURE; mode <= ADJUST; mode++)
// 	 {
// 	   bool leftedge = true;
// 	   PangoGlyphString *rightmost_glyphs = nil;
// 	   int rightmost_space = 0;
// 	   int residual = 0;

// 	   added_so_far = 0;
// 	   gaps_so_far = 0;

// 	   for (run_iter = line.runs; run_iter; run_iter = run_iter.next)
// 	 {
// 	   GlyphItem *run = run_iter.data;
// 	   PangoGlyphString *glyphs = run.glyphs;
// 	   PangoGlyphItemIter cluster_iter;
// 	   bool have_cluster;
// 	   int dir;
// 	   int offset;

// 	   dir = run.item.analysis.level % 2 == 0 ? +1 : -1;

// 	   /* We need character offset of the start of the run.  We don't have this.
// 		* Compute by counting from the beginning of the line.  The naming is
// 		* confusing.  Note that:
// 		*
// 		* run.item.offset        is byte offset of start of run in layout.text.
// 		* state.line_start_index  is byte offset of start of line in layout.text.
// 		* state.line_start_offset is character offset of start of line in layout.text.
// 		*/
// 	   assert (run.item.offset >= state.line_start_index);
// 	   offset = state.line_start_offset
// 		  + pango_utf8_strlen (text + state.line_start_index,
// 					   run.item.offset - state.line_start_index);

// 	   for (have_cluster = dir > 0 ?
// 		  pango_glyph_item_iter_init_start (&cluster_iter, run, text) :
// 		  pango_glyph_item_iter_init_end   (&cluster_iter, run, text);
// 			have_cluster;
// 			have_cluster = dir > 0 ?
// 			  pango_glyph_item_iter_next_cluster (&cluster_iter) :
// 			  pango_glyph_item_iter_prev_cluster (&cluster_iter))
// 		 {
// 		   int i;
// 		   int width = 0;

// 		   /* don't expand in the middle of graphemes */
// 		   if (!log_attrs[offset + cluster_iter.start_char].is_cursor_position)
// 		 continue;

// 		   for (i = cluster_iter.start_glyph; i != cluster_iter.end_glyph; i += dir)
// 		 width += glyphs.glyphs[i].geometry.width;

// 		   /* also don't expand zero-width clusters. */
// 		   if (width == 0)
// 		 continue;

// 		   gaps_so_far++;

// 		   if (mode == ADJUST)
// 		 {
// 		   int leftmost, rightmost;
// 		   int adjustment, space_left, space_right;

// 		   adjustment = total_remaining_width / total_gaps + residual;
// 		   if (is_hinted)
// 		   {
// 			 int old_adjustment = adjustment;
// 			 adjustment = PANGO_UNITS_ROUND (adjustment);
// 			 residual = old_adjustment - adjustment;
// 		   }
// 		   /* distribute to before/after */
// 		   distributeLetterSpacing (adjustment, &space_left, &space_right);

// 		   if (cluster_iter.start_glyph < cluster_iter.end_glyph)
// 		   {
// 			 /* LTR */
// 			 leftmost  = cluster_iter.start_glyph;
// 			 rightmost = cluster_iter.end_glyph - 1;
// 		   }
// 		   else
// 		   {
// 			 /* RTL */
// 			 leftmost  = cluster_iter.end_glyph + 1;
// 			 rightmost = cluster_iter.start_glyph;
// 		   }
// 		   /* Don't add to left-side of left-most glyph of left-most non-zero run. */
// 		   if (leftedge)
// 			 leftedge = false;
// 		   else
// 		   {
// 			 glyphs.glyphs[leftmost].geometry.width    += space_left ;
// 			 glyphs.glyphs[leftmost].geometry.x_offset += space_left ;
// 			 added_so_far += space_left;
// 		   }
// 		   /* Don't add to right-side of right-most glyph of right-most non-zero run. */
// 		   {
// 			 /* Save so we can undo later. */
// 			 rightmost_glyphs = glyphs;
// 			 rightmost_space = space_right;

// 			 glyphs.glyphs[rightmost].geometry.width  += space_right;
// 			 added_so_far += space_right;
// 		   }
// 		 }
// 		 }
// 	 }

// 	   if (mode == MEASURE)
// 	 {
// 	   total_gaps = gaps_so_far - 1;

// 	   if (total_gaps == 0)
// 		 {
// 		   /* a single cluster, can't really justify it */
// 		   return;
// 		 }
// 	 }
// 	   else /* mode == ADJUST */
// 		 {
// 	   if (rightmost_glyphs)
// 		{
// 		  rightmost_glyphs.glyphs[rightmost_glyphs.num_glyphs - 1].geometry.width -= rightmost_space;
// 		  added_so_far -= rightmost_space;
// 		}
// 	 }
// 	 }

//    state.remaining_width -= added_so_far;
//  }

//  static int
//  next_cluster_start (PangoGlyphString *gs,
// 			 int               cluster_start)
//  {
//    int i;

//    i = cluster_start + 1;
//    for (i < gs.num_glyphs)
// 	 {
// 	   if (gs.glyphs[i].attr.is_cluster_start)
// 	 return i;

// 	   i++;
// 	 }

//    return gs.num_glyphs;
//  }

//  static int
//  cluster_width (PangoGlyphString *gs,
// 			int               cluster_start)
//  {
//    int i;
//    int width;

//    width = gs.glyphs[cluster_start].geometry.width;
//    i = cluster_start + 1;
//    for (i < gs.num_glyphs)
// 	 {
// 	   if (gs.glyphs[i].attr.is_cluster_start)
// 	 break;

// 	   width += gs.glyphs[i].geometry.width;
// 	   i++;
// 	 }

//    return width;
//  }

//  static inline void
//  offset_y (LayoutIter *iter,
// 	   int             *y)
//  {
//    *y += iter.line_extents[iter.line_index].baseline;
//  }

//  /* Sets up the iter for the start of a new cluster. cluster_start_index
//   * is the byte index of the cluster start relative to the run.
//   */
//  static void
//  update_cluster (LayoutIter *iter,
// 		 int              cluster_start_index)
//  {
//    char             *cluster_text;
//    PangoGlyphString *gs;
//    int               cluster_length;

//    iter.character_position = 0;

//    gs = iter.run.glyphs;
//    iter.cluster_width = cluster_width (gs, iter.cluster_start);
//    iter.next_cluster_glyph = next_cluster_start (gs, iter.cluster_start);

//    if (iter.ltr)
// 	 {
// 	   /* For LTR text, finding the length of the cluster is easy
// 		* since logical and visual runs are in the same direction.
// 		*/
// 	   if (iter.next_cluster_glyph < gs.num_glyphs)
// 	 cluster_length = gs.log_clusters[iter.next_cluster_glyph] - cluster_start_index;
// 	   else
// 	 cluster_length = iter.run.item.length - cluster_start_index;
// 	 }
//    else
// 	 {
// 	   /* For RTL text, we have to scan backwards to find the previous
// 		* visual cluster which is the next logical cluster.
// 		*/
// 	   int i = iter.cluster_start;
// 	   for (i > 0 && gs.log_clusters[i - 1] == cluster_start_index)
// 	 i--;

// 	   if (i == 0)
// 	 cluster_length = iter.run.item.length - cluster_start_index;
// 	   else
// 	 cluster_length = gs.log_clusters[i - 1] - cluster_start_index;
// 	 }

//    cluster_text = iter.layout.text + iter.run.item.offset + cluster_start_index;
//    iter.cluster_num_chars = pango_utf8_strlen (cluster_text, cluster_length);

//    if (iter.ltr)
// 	 iter.index = cluster_text - iter.layout.text;
//    else
// 	 iter.index = g_utf8_prev_char (cluster_text + cluster_length) - iter.layout.text;
//  }

//  static void
//  update_run (LayoutIter *iter,
// 		 int              run_start_index)
//  {
//    const Extents *line_ext = &iter.line_extents[iter.line_index];

//    /* Note that in iter_new() the iter.run_width
// 	* is garbage but we don't use it since we're on the first run of
// 	* a line.
// 	*/
//    if (iter.run_list_link == iter.line.runs)
// 	 iter.run_x = line_ext.logical_rect.x;
//    else
// 	 iter.run_x += iter.run_width;

//    if (iter.run)
// 	 {
// 	   iter.run_width = pango_glyph_string_get_width (iter.run.glyphs);
// 	 }
//    else
// 	 {
// 	   /* The empty run at the end of a line */
// 	   iter.run_width = 0;
// 	 }

//    if (iter.run)
// 	 iter.ltr = (iter.run.item.analysis.level % 2) == 0;
//    else
// 	 iter.ltr = true;

//    iter.cluster_start = 0;
//    iter.cluster_x = iter.run_x;

//    if (iter.run)
// 	 {
// 	   update_cluster (iter, iter.run.glyphs.log_clusters[0]);
// 	 }
//    else
// 	 {
// 	   iter.cluster_width = 0;
// 	   iter.character_position = 0;
// 	   iter.cluster_num_chars = 0;
// 	   iter.index = run_start_index;
// 	 }
//  }

//  /**
//   * pango_layout_iter_copy:
//   * @iter: (nullable): a #LayoutIter, may be %nil
//   *
//   * Copies a #LayoutIter.
//   *
//   * Return value: (nullable): the newly allocated #LayoutIter,
//   *               which should be freed with pango_layout_iter_free(),
//   *               or %nil if @iter was %nil.
//   *
//   * Since: 1.20
//   **/
//  LayoutIter *
//  pango_layout_iter_copy (LayoutIter *iter)
//  {
//    LayoutIter *new;

//    if (iter == nil)
// 	 return nil;

//    new = g_slice_new (LayoutIter);

//    new.layout = g_object_ref (iter.layout);
//    new.line_list_link = iter.line_list_link;
//    new.line = iter.line;
//    pango_layout_line_ref (new.line);

//    new.run_list_link = iter.run_list_link;
//    new.run = iter.run;
//    new.index = iter.index;

//    new.line_extents = nil;
//    if (iter.line_extents != nil)
// 	 {
// 	   new.line_extents = g_memdup (iter.line_extents,
// 									 iter.layout.line_count * sizeof (Extents));

// 	 }
//    new.line_index = iter.line_index;

//    new.run_x = iter.run_x;
//    new.run_width = iter.run_width;
//    new.ltr = iter.ltr;

//    new.cluster_x = iter.cluster_x;
//    new.cluster_width = iter.cluster_width;

//    new.cluster_start = iter.cluster_start;
//    new.next_cluster_glyph = iter.next_cluster_glyph;

//    new.cluster_num_chars = iter.cluster_num_chars;
//    new.character_position = iter.character_position;

//    new.layout_width = iter.layout_width;

//    return new;
//  }

//  G_DEFINE_BOXED_TYPE (LayoutIter, pango_layout_iter,
// 					  pango_layout_iter_copy,
// 					  pango_layout_iter_free);

//  /**
//   * pango_layout_get_iter:
//   * @layout: a #Layout
//   *
//   * Returns an iterator to iterate over the visual extents of the layout.
//   *
//   * Return value: the new #LayoutIter that should be freed using
//   *               pango_layout_iter_free().
//   **/
//  LayoutIter*
//  pango_layout_get_iter (Layout *layout)
//  {
//    LayoutIter *iter;

//    g_return_val_if_fail (PANGO_IS_LAYOUT (layout), nil);

//    iter = g_slice_new (LayoutIter);

//    _pango_layout_get_iter (layout, iter);

//    return iter;
//  }

//  void
//  _pango_layout_get_iter (layout *Layout    ,
// 						 LayoutIter*iter)
//  {
//    int run_start_index;

//    g_return_if_fail (PANGO_IS_LAYOUT (layout));

//    iter.layout = g_object_ref (layout);

//    pango_layout_check_lines (layout);

//    iter.line_list_link = layout.lines;
//    iter.line = iter.line_list_link.data;
//    pango_layout_line_ref (iter.line);

//    run_start_index = iter.line.start_index;
//    iter.run_list_link = iter.line.runs;

//    if (iter.run_list_link)
// 	 {
// 	   iter.run = iter.run_list_link.data;
// 	   run_start_index = iter.run.item.offset;
// 	 }
//    else
// 	 iter.run = nil;

//    iter.line_extents = nil;

//    if (layout.width == -1)
// 	 {
// 	   Rectangle logical_rect;

// 	   pango_layout_get_extents_internal (layout,
// 										  nil,
// 										  &logical_rect,
// 										  &iter.line_extents);
// 	   iter.layout_width = logical_rect.width;
// 	 }
//    else
// 	 {
// 	   pango_layout_get_extents_internal (layout,
// 										  nil,
// 										  nil,
// 										  &iter.line_extents);
// 	   iter.layout_width = layout.width;
// 	 }
//    iter.line_index = 0;

//    update_run (iter, run_start_index);
//  }

//  void
//  _pango_layout_iter_destroy (LayoutIter *iter)
//  {
//    if (iter == nil)
// 	 return;

//    g_free (iter.line_extents);
//    pango_layout_line_unref (iter.line);
//    g_object_unref (iter.layout);
//  }

//  /**
//   * pango_layout_iter_free:
//   * @iter: (nullable): a #LayoutIter, may be %nil
//   *
//   * Frees an iterator that's no longer in use.
//   **/
//  void
//  pango_layout_iter_free (LayoutIter *iter)
//  {
//    if (iter == nil)
// 	 return;

//    _pango_layout_iter_destroy (iter);
//    g_slice_free (LayoutIter, iter);
//  }

//  /**
//   * pango_layout_iter_get_index:
//   * @iter: a #LayoutIter
//   *
//   * Gets the current byte index. Note that iterating forward by char
//   * moves in visual order, not logical order, so indexes may not be
//   * sequential. Also, the index may be equal to the length of the text
//   * in the layout, if on the %nil run (see pango_layout_iter_get_run()).
//   *
//   * Return value: current byte index.
//   **/
//  int
//  pango_layout_iter_get_index (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return 0;

//    return iter.index;
//  }

//  /**
//   * pango_layout_iter_get_run:
//   * @iter: a #LayoutIter
//   *
//   * Gets the current run. When iterating by run, at the end of each
//   * line, there's a position with a %nil run, so this function can return
//   * %nil. The %nil run at the end of each line ensures that all lines have
//   * at least one run, even lines consisting of only a newline.
//   *
//   * Use the faster pango_layout_iter_get_run_readonly() if you do not plan
//   * to modify the contents of the run (glyphs, glyph widths, etc.).
//   *
//   * Return value: (transfer none) (nullable): the current run.
//   **/
//  GlyphItem*
//  pango_layout_iter_get_run (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    pango_layout_line_leaked (iter.line);

//    return iter.run;
//  }

//  /**
//   * pango_layout_iter_get_run_readonly:
//   * @iter: a #LayoutIter
//   *
//   * Gets the current run. When iterating by run, at the end of each
//   * line, there's a position with a %nil run, so this function can return
//   * %nil. The %nil run at the end of each line ensures that all lines have
//   * at least one run, even lines consisting of only a newline.
//   *
//   * This is a faster alternative to pango_layout_iter_get_run(),
//   * but the user is not expected
//   * to modify the contents of the run (glyphs, glyph widths, etc.).
//   *
//   * Return value: (transfer none) (nullable): the current run, that
//   * should not be modified.
//   *
//   * Since: 1.16
//   **/
//  GlyphItem*
//  pango_layout_iter_get_run_readonly (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    pango_layout_line_leaked (iter.line);

//    return iter.run;
//  }

//  /* an inline-able version for local use */
//  static LayoutLine*
//  _pango_layout_iter_get_line (LayoutIter *iter)
//  {
//    return iter.line;
//  }

//  /**
//   * pango_layout_iter_get_line:
//   * @iter: a #LayoutIter
//   *
//   * Gets the current line.
//   *
//   * Use the faster pango_layout_iter_get_line_readonly() if you do not plan
//   * to modify the contents of the line (glyphs, glyph widths, etc.).
//   *
//   * Return value: (transfer none): the current line.
//   **/
//  LayoutLine*
//  pango_layout_iter_get_line (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    pango_layout_line_leaked (iter.line);

//    return iter.line;
//  }

//  /**
//   * pango_layout_iter_get_line_readonly:
//   * @iter: a #LayoutIter
//   *
//   * Gets the current line for read-only access.
//   *
//   * This is a faster alternative to pango_layout_iter_get_line(),
//   * but the user is not expected
//   * to modify the contents of the line (glyphs, glyph widths, etc.).
//   *
//   * Return value: (transfer none): the current line, that should not be
//   * modified.
//   *
//   * Since: 1.16
//   **/
//  LayoutLine*
//  pango_layout_iter_get_line_readonly (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    return iter.line;
//  }

//  /**
//   * pango_layout_iter_at_last_line:
//   * @iter: a #LayoutIter
//   *
//   * Determines whether @iter is on the last line of the layout.
//   *
//   * Return value: `true` if @iter is on the last line.
//   **/
//  bool
//  pango_layout_iter_at_last_line (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return false;

//    return iter.line_index == iter.layout.line_count - 1;
//  }

//  /**
//   * pango_layout_iter_get_layout:
//   * @iter: a #LayoutIter
//   *
//   * Gets the layout associated with a #LayoutIter.
//   *
//   * Return value: (transfer none): the layout associated with @iter.
//   *
//   * Since: 1.20
//   **/
//  Layout*
//  pango_layout_iter_get_layout (LayoutIter *iter)
//  {
//    /* check is redundant as it simply checks that iter.layout is not nil */
//    if (ITER_IS_INVALID (iter))
// 	 return nil;

//    return iter.layout;
//  }

//  static bool
//  line_is_terminated (LayoutIter *iter)
//  {
//    /* There is a real terminator at the end of each paragraph other
// 	* than the last.
// 	*/
//    if (iter.line_list_link.next)
// 	 {
// 	   LayoutLine *next_line = iter.line_list_link.next.data;
// 	   if (next_line.is_paragraph_start)
// 	 return true;
// 	 }

//    return false;
//  }

//  /* Moves to the next non-empty line. If @include_terminators
//   * is set, a line with just an explicit paragraph separator
//   * is considered non-empty.
//   */
//  static bool
//  next_nonempty_line (LayoutIter *iter,
// 			 bool         include_terminators)
//  {
//    bool result;

//    for (true)
// 	 {
// 	   result = pango_layout_iter_next_line (iter);
// 	   if (!result)
// 	 break;

// 	   if (iter.line.runs)
// 	 break;

// 	   if (include_terminators && line_is_terminated (iter))
// 	 break;
// 	 }

//    return result;
//  }

//  /* Moves to the next non-empty run. If @include_terminators
//   * is set, the trailing run at the end of a line with an explicit
//   * paragraph separator is considered non-empty.
//   */
//  static bool
//  next_nonempty_run (LayoutIter *iter,
// 			 bool         include_terminators)
//  {
//    bool result;

//    for (true)
// 	 {
// 	   result = pango_layout_iter_next_run (iter);
// 	   if (!result)
// 	 break;

// 	   if (iter.run)
// 	 break;

// 	   if (include_terminators && line_is_terminated (iter))
// 	 break;
// 	 }

//    return result;
//  }

//  /* Like pango_layout_next_cluster(), but if @include_terminators
//   * is set, includes the fake runs/clusters for empty lines.
//   * (But not positions introduced by line wrapping).
//   */
//  static bool
//  next_cluster_internal (LayoutIter *iter,
// 				bool         include_terminators)
//  {
//    PangoGlyphString *gs;
//    int               next_start;

//    if (ITER_IS_INVALID (iter))
// 	 return false;

//    if (iter.run == nil)
// 	 return next_nonempty_line (iter, include_terminators);

//    gs = iter.run.glyphs;

//    next_start = iter.next_cluster_glyph;
//    if (next_start == gs.num_glyphs)
// 	 {
// 	   return next_nonempty_run (iter, include_terminators);
// 	 }
//    else
// 	 {
// 	   iter.cluster_start = next_start;
// 	   iter.cluster_x += iter.cluster_width;
// 	   update_cluster(iter, gs.log_clusters[iter.cluster_start]);

// 	   return true;
// 	 }
//  }

//  /**
//   * pango_layout_iter_next_char:
//   * @iter: a #LayoutIter
//   *
//   * Moves @iter forward to the next character in visual order. If @iter was already at
//   * the end of the layout, returns %false.
//   *
//   * Return value: whether motion was possible.
//   **/
//  bool
//  pango_layout_iter_next_char (LayoutIter *iter)
//  {
//    const char *text;

//    if (ITER_IS_INVALID (iter))
// 	 return false;

//    if (iter.run == nil)
// 	 {
// 	   /* We need to fake an iterator position in the middle of a \r\n line terminator */
// 	   if (line_is_terminated (iter) &&
// 	   strncmp (iter.layout.text + iter.line.start_index + iter.line.length, "\r\n", 2) == 0 &&
// 	   iter.character_position == 0)
// 	 {
// 	   iter.character_position++;
// 	   return true;
// 	 }

// 	   return next_nonempty_line (iter, true);
// 	 }

//    iter.character_position++;
//    if (iter.character_position >= iter.cluster_num_chars)
// 	 return next_cluster_internal (iter, true);

//    text = iter.layout.text;
//    if (iter.ltr)
// 	 iter.index = g_utf8_next_char (text + iter.index) - text;
//    else
// 	 iter.index = g_utf8_prev_char (text + iter.index) - text;

//    return true;
//  }

//  /**
//   * pango_layout_iter_next_cluster:
//   * @iter: a #LayoutIter
//   *
//   * Moves @iter forward to the next cluster in visual order. If @iter
//   * was already at the end of the layout, returns %false.
//   *
//   * Return value: whether motion was possible.
//   **/
//  bool
//  pango_layout_iter_next_cluster (LayoutIter *iter)
//  {
//    return next_cluster_internal (iter, false);
//  }

//  /**
//   * pango_layout_iter_next_run:
//   * @iter: a #LayoutIter
//   *
//   * Moves @iter forward to the next run in visual order. If @iter was
//   * already at the end of the layout, returns %false.
//   *
//   * Return value: whether motion was possible.
//   **/
//  bool
//  pango_layout_iter_next_run (LayoutIter *iter)
//  {
//    int next_run_start; /* byte index */
//    GSList *next_link;

//    if (ITER_IS_INVALID (iter))
// 	 return false;

//    if (iter.run == nil)
// 	 return pango_layout_iter_next_line (iter);

//    next_link = iter.run_list_link.next;

//    if (next_link == nil)
// 	 {
// 	   /* Moving on to the zero-width "virtual run" at the end of each
// 		* line
// 		*/
// 	   next_run_start = iter.run.item.offset + iter.run.item.length;
// 	   iter.run = nil;
// 	   iter.run_list_link = nil;
// 	 }
//    else
// 	 {
// 	   iter.run_list_link = next_link;
// 	   iter.run = iter.run_list_link.data;
// 	   next_run_start = iter.run.item.offset;
// 	 }

//    update_run (iter, next_run_start);

//    return true;
//  }

//  /**
//   * pango_layout_iter_next_line:
//   * @iter: a #LayoutIter
//   *
//   * Moves @iter forward to the start of the next line. If @iter is
//   * already on the last line, returns %false.
//   *
//   * Return value: whether motion was possible.
//   **/
//  bool
//  pango_layout_iter_next_line (LayoutIter *iter)
//  {
//    GSList *next_link;

//    if (ITER_IS_INVALID (iter))
// 	 return false;

//    next_link = iter.line_list_link.next;

//    if (next_link == nil)
// 	 return false;

//    iter.line_list_link = next_link;

//    pango_layout_line_unref (iter.line);

//    iter.line = iter.line_list_link.data;

//    pango_layout_line_ref (iter.line);

//    iter.run_list_link = iter.line.runs;

//    if (iter.run_list_link)
// 	 iter.run = iter.run_list_link.data;
//    else
// 	 iter.run = nil;

//    iter.line_index ++;

//    update_run (iter, iter.line.start_index);

//    return true;
//  }

//  /**
//   * pango_layout_iter_get_char_extents:
//   * @iter: a #LayoutIter
//   * @logical_rect: (out caller-allocates): rectangle to fill with
//   *   logical extents
//   *
//   * Gets the extents of the current character, in layout coordinates
//   * (origin is the top left of the entire layout). Only logical extents
//   * can sensibly be obtained for characters; ink extents make sense only
//   * down to the level of clusters.
//   *
//   **/
//  void
//  pango_layout_iter_get_char_extents (LayoutIter *iter,
// 					 Rectangle  *logical_rect)
//  {
//    Rectangle cluster_rect;
//    int            x0, x1;

//    if (ITER_IS_INVALID (iter))
// 	 return;

//    if (logical_rect == nil)
// 	 return;

//    pango_layout_iter_get_cluster_extents (iter, nil, &cluster_rect);

//    if (iter.run == nil)
// 	 {
// 	   /* When on the nil run, cluster, char, and run all have the
// 		* same extents
// 		*/
// 	   *logical_rect = cluster_rect;
// 	   return;
// 	 }

//    if (iter.cluster_num_chars)
//    {
// 	 x0 = (iter.character_position * cluster_rect.width) / iter.cluster_num_chars;
// 	 x1 = ((iter.character_position + 1) * cluster_rect.width) / iter.cluster_num_chars;
//    }
//    else
//    {
// 	 x0 = x1 = 0;
//    }

//    logical_rect.width = x1 - x0;
//    logical_rect.height = cluster_rect.height;
//    logical_rect.y = cluster_rect.y;
//    logical_rect.x = cluster_rect.x + x0;
//  }

//  /**
//   * pango_layout_iter_get_cluster_extents:
//   * @iter: a #LayoutIter
//   * @ink_rect: (out) (allow-none): rectangle to fill with ink extents, or %nil
//   * @logical_rect: (out) (allow-none): rectangle to fill with logical extents, or %nil
//   *
//   * Gets the extents of the current cluster, in layout coordinates
//   * (origin is the top left of the entire layout).
//   *
//   **/
//  void
//  pango_layout_iter_get_cluster_extents (LayoutIter *iter,
// 						Rectangle  *ink_rect,
// 						Rectangle  *logical_rect)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return;

//    if (iter.run == nil)
// 	 {
// 	   /* When on the nil run, cluster, char, and run all have the
// 		* same extents
// 		*/
// 	   pango_layout_iter_get_run_extents (iter, ink_rect, logical_rect);
// 	   return;
// 	 }

//    pango_glyph_string_extents_range (iter.run.glyphs,
// 					 iter.cluster_start,
// 					 iter.next_cluster_glyph,
// 					 iter.run.item.analysis.font,
// 					 ink_rect,
// 					 logical_rect);

//    if (ink_rect)
// 	 {
// 	   ink_rect.x += iter.cluster_x;
// 	   offset_y (iter, &ink_rect.y);
// 	 }

//    if (logical_rect)
// 	 {
// 	   assert (logical_rect.width == iter.cluster_width);
// 	   logical_rect.x += iter.cluster_x;
// 	   offset_y (iter, &logical_rect.y);
// 	 }
//  }

//  /**
//   * pango_layout_iter_get_run_extents:
//   * @iter: a #LayoutIter
//   * @ink_rect: (out) (allow-none): rectangle to fill with ink extents, or %nil
//   * @logical_rect: (out) (allow-none): rectangle to fill with logical extents, or %nil
//   *
//   * Gets the extents of the current run in layout coordinates
//   * (origin is the top left of the entire layout).
//   *
//   **/
//  void
//  pango_layout_iter_get_run_extents (LayoutIter *iter,
// 					Rectangle  *ink_rect,
// 					Rectangle  *logical_rect)
//  {
//    if (G_UNLIKELY (!ink_rect && !logical_rect))
// 	 return;

//    if (ITER_IS_INVALID (iter))
// 	 return;

//    if (iter.run)
// 	 {
// 	   pango_layout_run_get_extents_and_height (iter.run, ink_rect, logical_rect, nil);

// 	   if (ink_rect)
// 	 {
// 	   offset_y (iter, &ink_rect.y);
// 	   ink_rect.x += iter.run_x;
// 	 }

// 	   if (logical_rect)
// 	 {
// 	   offset_y (iter, &logical_rect.y);
// 	   logical_rect.x += iter.run_x;
// 	 }
// 	 }
//    else
// 	 {
// 	   /* The empty run at the end of a line */

// 	   pango_layout_iter_get_line_extents (iter, ink_rect, logical_rect);

// 	   if (ink_rect)
// 	 {
// 	   ink_rect.x = iter.run_x;
// 	   ink_rect.width = 0;
// 	 }

// 	   if (logical_rect)
// 	 {
// 	   logical_rect.x = iter.run_x;
// 	   logical_rect.width = 0;
// 	 }
// 	 }
//  }

//  /**
//   * pango_layout_iter_get_line_extents:
//   * @iter: a #LayoutIter
//   * @ink_rect: (out) (allow-none): rectangle to fill with ink extents, or %nil
//   * @logical_rect: (out) (allow-none): rectangle to fill with logical extents, or %nil
//   *
//   * Obtains the extents of the current line. @ink_rect or @logical_rect
//   * can be %nil if you aren't interested in them. Extents are in layout
//   * coordinates (origin is the top-left corner of the entire
//   * #Layout).  Thus the extents returned by this function will be
//   * the same width/height but not at the same x/y as the extents
//   * returned from pango_layout_line_get_extents().
//   *
//   **/
//  void
//  pango_layout_iter_get_line_extents (LayoutIter *iter,
// 					 Rectangle  *ink_rect,
// 					 Rectangle  *logical_rect)
//  {
//    const Extents *ext;

//    if (ITER_IS_INVALID (iter))
// 	 return;

//    ext = &iter.line_extents[iter.line_index];

//    if (ink_rect)
// 	 {
// 	   get_line_extents_layout_coords (iter.layout, iter.line,
// 					   iter.layout_width,
// 					   ext.logical_rect.y,
// 									   nil,
// 					   ink_rect,
// 					   nil);
// 	 }

//    if (logical_rect)
// 	 *logical_rect = ext.logical_rect;
//  }

//  /**
//   * pango_layout_iter_get_line_yrange:
//   * @iter: a #LayoutIter
//   * @y0_: (out) (allow-none): start of line, or %nil
//   * @y1_: (out) (allow-none): end of line, or %nil
//   *
//   * Divides the vertical space in the #Layout being iterated over
//   * between the lines in the layout, and returns the space belonging to
//   * the current line.  A line's range includes the line's logical
//   * extents, plus half of the spacing above and below the line, if
//   * pango_layout_set_spacing() has been called to set layout spacing.
//   * The Y positions are in layout coordinates (origin at top left of the
//   * entire layout).
//   *
//   * Note: Since 1.44, Pango uses line heights for placing lines,
//   * and there may be gaps between the ranges returned by this
//   * function.
//   */
//  void
//  pango_layout_iter_get_line_yrange (LayoutIter *iter,
// 					int             *y0,
// 					int             *y1)
//  {
//    const Extents *ext;
//    int half_spacing;

//    if (ITER_IS_INVALID (iter))
// 	 return;

//    ext = &iter.line_extents[iter.line_index];

//    half_spacing = iter.layout.spacing / 2;

//    /* Note that if layout.spacing is odd, the remainder spacing goes
// 	* above the line (this is pretty arbitrary of course)
// 	*/

//    if (y0)
// 	 {
// 	   /* No spacing above the first line */

// 	   if (iter.line_index == 0)
// 	 *y0 = ext.logical_rect.y;
// 	   else
// 	 *y0 = ext.logical_rect.y - (iter.layout.spacing - half_spacing);
// 	 }

//    if (y1)
// 	 {
// 	   /* No spacing below the last line */
// 	   if (iter.line_index == iter.layout.line_count - 1)
// 	 *y1 = ext.logical_rect.y + ext.logical_rect.height;
// 	   else
// 	 *y1 = ext.logical_rect.y + ext.logical_rect.height + half_spacing;
// 	 }
//  }

//  /**
//   * pango_layout_iter_get_baseline:
//   * @iter: a #LayoutIter
//   *
//   * Gets the Y position of the current line's baseline, in layout
//   * coordinates (origin at top left of the entire layout).
//   *
//   * Return value: baseline of current line.
//   **/
//  int
//  pango_layout_iter_get_baseline (LayoutIter *iter)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return 0;

//    return iter.line_extents[iter.line_index].baseline;
//  }

//  /**
//   * pango_layout_iter_get_layout_extents:
//   * @iter: a #LayoutIter
//   * @ink_rect: (out) (allow-none): rectangle to fill with ink extents,
//   *            or %nil
//   * @logical_rect: (out) (allow-none): rectangle to fill with logical
//   *                extents, or %nil
//   *
//   * Obtains the extents of the #Layout being iterated
//   * over. @ink_rect or @logical_rect can be %nil if you
//   * aren't interested in them.
//   *
//   **/
//  void
//  pango_layout_iter_get_layout_extents  (LayoutIter *iter,
// 						Rectangle  *ink_rect,
// 						Rectangle  *logical_rect)
//  {
//    if (ITER_IS_INVALID (iter))
// 	 return;

//    pango_layout_get_extents (iter.layout, ink_rect, logical_rect);
//  }
