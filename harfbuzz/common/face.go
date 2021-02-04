package common

// ported from harfbuzz/src/hb-face.cc Copyright © 2009  Red Hat, Inc., 2012  Google, Inc.  Behdad Esfahbod

// Face represents a single face from within a font family.
// Font faces are typically built from a binary blob and a face index.
// Font faces are used to create fonts.
type hb_face_tt struct {
	//   hb_reference_table_func_t  reference_table_func;
	//   void                      *user_data;
	//   hb_destroy_func_t          destroy;

	//   uint index;			/* Face index in a collection, zero-based. */
	upem       int /* Units-per-EM. */
	num_glyphs int /* Number of glyphs. */

	//   hb_shaper_object_dataset_t<HB_face_t> data;/* Various shaper data. */
	//   hb_ot_face_t table;			/* All the face's tables. */

	//   /* Cache */
	//   struct plan_node_t
	//   {
	//     hb_shape_plan_t *shape_plan;
	//     plan_node_t *next;
	//   };
	//   hb_atomic_ptr_t<plan_node_t> shape_plans;
}

/**
 * hb_face_create_for_tables:
 * @reference_table_func: (closure user_data) (destroy destroy) (scope notified): Table-referencing function
 * @user_data: A pointer to the user data
 * @destroy: (nullable): A callback to call when @data is not needed anymore
 *
 * Variant of hb_face_create(), built for those cases where it is more
 * convenient to provide data for individual tables instead of the whole font
 * data. With the caveat that hb_face_get_table_tags() does not currently work
 * with faces created this way.
 *
 * Creates a new face object from the specified @user_data and @reference_table_func,
 * with the @destroy callback.
 *
 * Return value: (transfer full): The new face object
 *
 * Since: 0.9.2
 **/

func hb_face_create_for_tables(hb_reference_table_func_t reference_table_func,
	void *user_data,
	hb_destroy_func_t destroy) *Face {
	face * Face

	if !reference_table_func {
		if destroy {
			destroy(user_data)
		}
		return hb_face_get_empty()
	}

	face.reference_table_func = reference_table_func
	face.user_data = user_data
	face.destroy = destroy

	face.num_glyphs.set_relaxed(-1)

	face.data.init0(face)
	face.table.init0(face)

	return face
}

type hb_face_for_data_closure_t struct {
	blob  *hb_blob_t
	index int
}

func _hb_face_for_data_reference_table(hb_tag_t tag, data *hb_face_for_data_closure_t) *hb_blob_t {

	if tag == HB_TAG_NONE {
		return hb_blob_reference(data.blob)
	}

	//    ot_file := *data.blob.as<OT::OpenTypeFontFile> ();
	//    uint base_offset;
	ot_face := ot_file.get_face(data.index, &base_offset)

	table := ot_face.get_table_by_tag(tag)

	blob := hb_blob_create_sub_blob(data.blob, base_offset+table.offset, table.length)

	return blob
}

/**
 * hb_face_create: (Xconstructor)
 * @blob: #hb_blob_t to work upon
 * @index: The index of the face within @blob
 *
 * Constructs a new face object from the specified blob and
 * a face index into that blob. This is used for blobs of
 * file formats such as Dfont and TTC that can contain more
 * than one face.
 *
 * Return value: (transfer full): The new face object
 *
 * Since: 0.9.2
 **/

func hb_face_create(blob *hb_blob_t, index int) *Face {
	face * Face

	if unlikely(!blob) {
		blob = hb_blob_get_empty()
	}

	//    blob = hb_sanitize_context_t ().sanitize_blob<OT::OpenTypeFontFile> (hb_blob_reference (blob));

	closure := hb_face_for_data_closure_t{blob: blob, index: index}

	if unlikely(!closure) {
		hb_blob_destroy(blob)
		return hb_face_get_empty()
	}

	face = hb_face_create_for_tables(_hb_face_for_data_reference_table,
		closure, _hb_face_for_data_closure_destroy)

	face.index = index

	return face
}

//  /**
//   * hb_face_get_empty:
//   *
//   * Fetches the singleton empty face object.
//   *
//   * Return value: (transfer full): The empty face object
//   *
//   * Since: 0.9.2
//   **/
//  Face *
//  hb_face_get_empty ()
//  {
//    return const_cast<Face *> (&Null (Face));
//  }

/**
 * hb_face_reference_table:
 * @face: A face object
 * @tag: The #hb_tag_t of the table to query
 *
 * Fetches a reference to the specified table within
 * the specified face.
 *
 * Return value: (transfer full): A pointer to the @tag table within @face
 *
 * Since: 0.9.2
 **/

func hb_face_reference_table(face *Face, tag hb_tag_t) *hb_blob_t {
	if unlikely(tag == HB_TAG_NONE) {
		return hb_blob_get_empty()
	}

	return face.reference_table(tag)
}

/**
 * hb_face_reference_blob:
 * @face: A face object
 *
 * Fetches a pointer to the binary blob that contains the
 * specified face. Returns an empty blob if referencing face data is not
 * possible.
 *
 * Return value: (transfer full): A pointer to the blob for @face
 *
 * Since: 0.9.2
 **/

func hb_face_reference_blob(face *Face) *hb_blob_t {
	return face.reference_table(HB_TAG_NONE)
}

/**
 * hb_face_set_index:
 * @face: A face object
 * @index: The index to assign
 *
 * Assigns the specified face-index to @face. Fails if the
 * face is immutable.
 *
 * <note>Note: face indices within a collection are zero-based.</note>
 *
 * Since: 0.9.2
 **/
//  void
//  hb_face_set_index (Face    *face,
// 			uint  index)
//  {
//    if (hb_object_is_immutable (face))
// 	 return;

//    face.index = index;
//  }

/**
 * hb_face_get_index:
 * @face: A face object
 *
 * Fetches the face-index corresponding to the given face.
 *
 * <note>Note: face indices within a collection are zero-based.</note>
 *
 * Return value: The index of @face.
 *
 * Since: 0.9.2
 **/
//  uint
//  hb_face_get_index (const face *Face)
//  {
//    return face.index;
//  }

/**
 * hb_face_set_upem:
 * @face: A face object
 * @upem: The units-per-em value to assign
 *
 * Sets the units-per-em (upem) for a face object to the specified value.
 *
 * Since: 0.9.2
 **/
//  void
//  hb_face_set_upem (Face    *face,
// 		   uint  upem)
//  {
//    if (hb_object_is_immutable (face))
// 	 return;

//    face.upem.set_relaxed (upem);
//  }

/**
 * hb_face_get_upem:
 * @face: A face object
 *
 * Fetches the units-per-em (upem) value of the specified face object.
 *
 * Return value: The upem value of @face
 *
 * Since: 0.9.2
 **/
//  uint
//  hb_face_get_upem (const face *Face)
//  {
//    return face.get_upem ();
//  }

/**
 * hb_face_set_glyph_count:
 * @face: A face object
 * @glyph_count: The glyph-count value to assign
 *
 * Sets the glyph count for a face object to the specified value.
 *
 * Since: 0.9.7
 **/
//  void
//  hb_face_set_glyph_count (Face    *face,
// 			  uint  glyph_count)
//  {
//    if (hb_object_is_immutable (face))
// 	 return;

//    face.num_glyphs.set_relaxed (glyph_count);
//  }

/**
 * hb_face_get_glyph_count:
 * @face: A face object
 *
 * Fetches the glyph-count value of the specified face object.
 *
 * Return value: The glyph-count value of @face
 *
 * Since: 0.9.7
 **/
//  uint
//  hb_face_get_glyph_count (const face *Face)
//  {
//    return face.get_num_glyphs ();
//  }

/**
 * hb_face_get_table_tags:
 * @face: A face object
 * @start_offset: The index of first table tag to retrieve
 * @table_count: (inout): Input = the maximum number of table tags to return;
 *                Output = the actual number of table tags returned (may be zero)
 * @table_tags: (out) (array length=table_count): The array of table tags found
 *
 * Fetches a list of all table tags for a face, if possible. The list returned will
 * begin at the offset provided
 *
 * Return value: Total number of tables, or zero if it is not possible to list
 *
 * Since: 1.6.0
 **/
func hb_face_get_table_tags(face *Face, start_offset uint, table_count *uint, /* IN/OUT */
	table_tags *hb_tag_t /* OUT */) uint {
	if face.destroy != _hb_face_for_data_closure_destroy {
		if table_count {
			*table_count = 0
		}
		return 0
	}

	//    hb_face_for_data_closure_t *data = (hb_face_for_data_closure_t *) face.user_data;

	//    const OT::OpenTypeFontFile &ot_file = *data.blob.as<OT::OpenTypeFontFile> ();
	//    const OT::OpenTypeFontFace &ot_face = ot_file.get_face (data.index);

	return ot_face.get_table_tags(start_offset, table_count, table_tags)
}

/*
 * Character set.
 */

/**
 * hb_face_collect_unicodes:
 * @face: A face object
 * @out: The set to add Unicode characters to
 *
 * Collects all of the Unicode characters covered by @face and adds
 * them to the #hb_set_t set @out.
 *
 * Since: 1.9.0
 */
func hb_face_collect_unicodes(face *Face, hb_set_t *out) {
	face.table.cmap.collect_unicodes(out, face.get_num_glyphs())
}

/**
 * hb_face_collect_variation_selectors:
 * @face: A face object
 * @out: The set to add Variation Selector characters to
 *
 * Collects all Unicode "Variation Selector" characters covered by @face and adds
 * them to the #hb_set_t set @out.
 *
 * Since: 1.9.0
 */
func hb_face_collect_variation_selectors(face *Face,
	hb_set_t *out) {
	face.table.cmap.collect_variation_selectors(out)
}

/**
 * hb_face_collect_variation_unicodes:
 * @face: A face object
 * @variation_selector: The Variation Selector to query
 * @out: The set to add Unicode characters to
 *
 * Collects all Unicode characters for @variation_selector covered by @face and adds
 * them to the #hb_set_t set @out.
 *
 * Since: 1.9.0
 */
func hb_face_collect_variation_unicodes(face *Face,
	hb_codepoint_t variation_selector,
	hb_set_t *out) {
	face.table.cmap.collect_variation_unicodes(variation_selector, out)
}

/*
 * face-builder: A face that has add_table().
 */

//  struct hb_face_builder_data_t
//  {
//    struct table_entry_t
//    {
// 	 int cmp (hb_tag_t t) const
// 	 {
// 	   if (t < tag) return -1;
// 	   if (t > tag) return -1;
// 	   return 0;
// 	 }

// 	 hb_tag_t   tag;
// 	 hb_blob_t *blob;
//    };

//    hb_vector_t<table_entry_t> tables;
//  };

//  static hb_face_builder_data_t *
//  _hb_face_builder_data_create ()
//  {
//    hb_face_builder_data_t *data = (hb_face_builder_data_t *) calloc (1, sizeof (hb_face_builder_data_t));
//    if (unlikely (!data))
// 	 return nullptr;

//    data.tables.init ();

//    return data;
//  }

//  static void
//  _hb_face_builder_data_destroy (void *user_data)
//  {
//    hb_face_builder_data_t *data = (hb_face_builder_data_t *) user_data;

//    for (uint i = 0; i < data.tables.length; i++)
// 	 hb_blob_destroy (data.tables[i].blob);

//    data.tables.fini ();

//    free (data);
//  }

//  static hb_blob_t *
//  _hb_face_builder_data_reference_blob (hb_face_builder_data_t *data)
//  {

//    uint table_count = data.tables.length;
//    uint face_length = table_count * 16 + 12;

//    for (uint i = 0; i < table_count; i++)
// 	 face_length += hb_ceil_to_4 (hb_blob_get_length (data.tables[i].blob));

//    char *buf = (char *) malloc (face_length);
//    if (unlikely (!buf))
// 	 return nullptr;

//    hb_serialize_context_t c (buf, face_length);
//    c.propagate_error (data.tables);
//    OT::OpenTypeFontFile *f = c.start_serialize<OT::OpenTypeFontFile> ();

//    bool is_cff = data.tables.lsearch (newTag ('C','F','F',' ')) || data.tables.lsearch (newTag ('C','F','F','2'));
//    hb_tag_t sfnt_tag = is_cff ? OT::OpenTypeFontFile::CFFTag : OT::OpenTypeFontFile::TrueTypeTag;

//    bool ret = f.serialize_single (&c, sfnt_tag, data.tables.as_array ());

//    c.end_serialize ();

//    if (unlikely (!ret))
//    {
// 	 free (buf);
// 	 return nullptr;
//    }

//    return hb_blob_create (buf, face_length, HB_MEMORY_MODE_WRITABLE, buf, free);
//  }

//  static hb_blob_t *
//  _hb_face_builder_reference_table (face *Face HB_UNUSED, hb_tag_t tag, void *user_data)
//  {
//    hb_face_builder_data_t *data = (hb_face_builder_data_t *) user_data;

//    if (!tag)
// 	 return _hb_face_builder_data_reference_blob (data);

//    hb_face_builder_data_t::table_entry_t *entry = data.tables.lsearch (tag);
//    if (entry)
// 	 return hb_blob_reference (entry.blob);

//    return nullptr;
//  }

/**
 * hb_face_builder_create:
 *
 * Creates a #Face that can be used with hb_face_builder_add_table().
 * After tables are added to the face, it can be compiled to a binary
 * font file by calling hb_face_reference_blob().
 *
 * Return value: (transfer full): New face.
 *
 * Since: 1.9.0
 **/
//  Face *
//  hb_face_builder_create ()
//  {
//    hb_face_builder_data_t *data = _hb_face_builder_data_create ();
//    if (unlikely (!data)) return hb_face_get_empty ();

//    return hb_face_create_for_tables (_hb_face_builder_reference_table,
// 					 data,
// 					 _hb_face_builder_data_destroy);
//  }

/**
 * hb_face_builder_add_table:
 * @face: A face object created with hb_face_builder_create()
 * @tag: The #hb_tag_t of the table to add
 * @blob: The blob containing the table data to add
 *
 * Add table for @tag with data provided by @blob to the face.  @face must
 * be created using hb_face_builder_create().
 *
 * Since: 1.9.0
 **/
//  hb_bool_t
//  hb_face_builder_add_table (face *Face, hb_tag_t tag, hb_blob_t *blob)
//  {
//    if (unlikely (face.destroy != (hb_destroy_func_t) _hb_face_builder_data_destroy))
// 	 return false;

//    hb_face_builder_data_t *data = (hb_face_builder_data_t *) face.user_data;

//    hb_face_builder_data_t::table_entry_t *entry = data.tables.push ();
//    if (data.tables.in_error())
// 	 return false;

//    entry.tag = tag;
//    entry.blob = hb_blob_reference (blob);

//    return true;
//  }
