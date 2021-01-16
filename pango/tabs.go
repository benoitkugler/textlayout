package pango

const (
	PANGO_TAB_LEFT TabAlign = iota // the tab stop appears to the left of the text

	/* These are not supported now, but may be in the
	* future.
	*
	*  PANGO_TAB_RIGHT,
	*  PANGO_TAB_CENTER,
	*  PANGO_TAB_NUMERIC
	 */
)

// A TabAlign specifies where a tab stop appears relative to the text
type TabAlign uint8

type Tab struct {
	location int /* Offset in pixels of this tab stop
	 * from the left margin of the text.
	 */
	alignment TabAlign /* Where the tab stop appears relative
	 * to the text.
	 */
}

/**
 * TabArray:
 *
 * A #TabArray struct contains an array
 * of tab stops. Each tab stop has an alignment and a position.
 */
type TabArray struct {
	tabs                []Tab
	positions_in_pixels bool
}

//  static void
//  init_tabs (TabArray *array, gint start, gint end)
//  {
//    while (start < end)
// 	 {
// 	   array.tabs[start].location = 0;
// 	   array.tabs[start].alignment = PANGO_TAB_LEFT;
// 	   ++start;
// 	 }
//  }

/**
 * pango_tab_array_new:
 * @initial_size: Initial number of tab stops to allocate, can be 0
 * @positions_in_pixels: whether positions are in pixel units
 *
 * Creates an array of @initial_size tab stops. Tab stops are specified in
 * pixel units if @positions_in_pixels is %TRUE, otherwise in Pango
 * units. All stops are initially at position 0.
 *
 * Return value: the newly allocated #TabArray, which should
 *               be freed with pango_tab_array_free().
 **/
//  TabArray*
//  pango_tab_array_new (gint initial_size,
// 			  gboolean positions_in_pixels)
//  {
//    TabArray *array;

//    g_return_val_if_fail (initial_size >= 0, NULL);

//    /* alloc enough to treat array.tabs as an array of length
// 	* size, though it's declared as an array of length 1.
// 	* If we allowed tab array resizing we'd need to drop this
// 	* optimization.
// 	*/
//    array = g_slice_new (TabArray);
//    array.size = initial_size;
//    array.allocated = initial_size;

//    if (array.allocated > 0)
// 	 {
// 	   array.tabs = g_new (Tab, array.allocated);
// 	   init_tabs (array, 0, array.allocated);
// 	 }
//    else
// 	 array.tabs = NULL;

//    array.positions_in_pixels = positions_in_pixels;

//    return array;
//  }

/**
 * pango_tab_array_new_with_positions:
 * @size: number of tab stops in the array
 * @positions_in_pixels: whether positions are in pixel units
 * @first_alignment: alignment of first tab stop
 * @first_position: position of first tab stop
 * @...: additional alignment/position pairs
 *
 * This is a convenience function that creates a #TabArray
 * and allows you to specify the alignment and position of each
 * tab stop. You <emphasis>must</emphasis> provide an alignment
 * and position for @size tab stops.
 *
 * Return value: the newly allocated #TabArray, which should
 *               be freed with pango_tab_array_free().
 **/
//  TabArray  *
//  pango_tab_array_new_with_positions (gint           size,
// 					 gboolean       positions_in_pixels,
// 					 TabAlign  first_alignment,
// 					 gint           first_position,
// 					 ...)
//  {
//    TabArray *array;
//    va_list args;
//    int i;

//    g_return_val_if_fail (size >= 0, NULL);

//    array = pango_tab_array_new (size, positions_in_pixels);

//    if (size == 0)
// 	 return array;

//    array.tabs[0].alignment = first_alignment;
//    array.tabs[0].location = first_position;

//    if (size == 1)
// 	 return array;

//    va_start (args, first_position);

//    i = 1;
//    while (i < size)
// 	 {
// 	   TabAlign align = va_arg (args, TabAlign);
// 	   int pos = va_arg (args, int);

// 	   array.tabs[i].alignment = align;
// 	   array.tabs[i].location = pos;

// 	   ++i;
// 	 }

//    va_end (args);

//    return array;
//  }

//  G_DEFINE_BOXED_TYPE (TabArray, pango_tab_array,
// 					  pango_tab_array_copy,
// 					  pango_tab_array_free);

/**
 * pango_tab_array_copy:
 * @src: #TabArray to copy
 *
 * Copies a #TabArray
 *
 * Return value: the newly allocated #TabArray, which should
 *               be freed with pango_tab_array_free().
 **/
func (src *TabArray) pango_tab_array_copy() *TabArray {
	if src == nil {
		return nil
	}
	copy := src
	copy.tabs = append([]Tab(nil), src.tabs...)

	return copy
}

/**
 * pango_tab_array_resize:
 * @tabArray: a #TabArray
 * @new_size: new size of the array
 *
 * Resizes a tab array. You must subsequently initialize any tabs that
 * were added as a result of growing the array.
 *
 **/
//  void
//  pango_tab_array_resize (TabArray *tabArray,
// 			 gint           new_size)
//  {
//    if (new_size > tabArray.allocated)
// 	 {
// 	   gint current_end = tabArray.allocated;

// 	   /* Ratchet allocated size up above the index. */
// 	   if (tabArray.allocated == 0)
// 	 tabArray.allocated = 2;

// 	   while (new_size > tabArray.allocated)
// 	 tabArray.allocated = tabArray.allocated * 2;

// 	   tabArray.tabs = g_renew (Tab, tabArray.tabs,
// 				  tabArray.allocated);

// 	   init_tabs (tabArray, current_end, tabArray.allocated);
// 	 }

//    tabArray.size = new_size;
//  }

/**
 * pango_tab_array_set_tab:
 * @tabArray: a #TabArray
 * @tabIndex: the index of a tab stop
 * @alignment: tab alignment
 * @location: tab location in Pango units
 *
 * Sets the alignment and location of a tab stop.
 * @alignment must always be #PANGO_TAB_LEFT in the current
 * implementation.
 *
 **/
//  void
//  pango_tab_array_set_tab  (TabArray *tabArray,
// 			   gint           tabIndex,
// 			   TabAlign  alignment,
// 			   gint           location)
//  {
//    g_return_if_fail (tabArray != NULL);
//    g_return_if_fail (tabIndex >= 0);
//    g_return_if_fail (alignment == PANGO_TAB_LEFT);
//    g_return_if_fail (location >= 0);

//    if (tabIndex >= tabArray.size)
// 	 pango_tab_array_resize (tabArray, tabIndex + 1);

//    tabArray.tabs[tabIndex].alignment = alignment;
//    tabArray.tabs[tabIndex].location = location;
//  }

//  pango_tab_array_get_tab gets the alignment and position of a tab stop.
func (tabArray *TabArray) pango_tab_array_get_tab(tabIndex int) (TabAlign, int) {
	return tabArray.tabs[tabIndex].alignment, tabArray.tabs[tabIndex].location
}

/**
 * pango_tab_array_get_tabs:
 * @tabArray: a #TabArray
 * @alignments: (out) (allow-none): location to store an array of tab
 *   stop alignments, or %NULL
 * @locations: (out) (allow-none) (array): location to store an array
 *   of tab positions, or %NULL
 *
 * If non-%NULL, @alignments and @locations are filled with allocated
 * arrays of length pango_tab_array_get_size(). You must free the
 * returned array.
 *
 **/
//  void
//  pango_tab_array_get_tabs (TabArray *tabArray,
// 			   TabAlign **alignments,
// 			   gint          **locations)
//  {
//    gint i;

//    g_return_if_fail (tabArray != NULL);

//    if (alignments)
// 	 *alignments = g_new (TabAlign, tabArray.size);

//    if (locations)
// 	 *locations = g_new (gint, tabArray.size);

//    i = 0;
//    while (i < tabArray.size)
// 	 {
// 	   if (alignments)
// 	 (*alignments)[i] = tabArray.tabs[i].alignment;
// 	   if (locations)
// 	 (*locations)[i] = tabArray.tabs[i].location;

// 	   ++i;
// 	 }
//  }

/**
 * pango_tab_array_get_positions_in_pixels:
 * @tabArray: a #TabArray
 *
 * Returns %TRUE if the tab positions are in pixels, %FALSE if they are
 * in Pango units.
 *
 * Return value: whether positions are in pixels.
 **/
//  gboolean
//  pango_tab_array_get_positions_in_pixels (TabArray *tabArray)
//  {
//    g_return_val_if_fail (tabArray != NULL, FALSE);

//    return tabArray.positions_in_pixels;
//  }
