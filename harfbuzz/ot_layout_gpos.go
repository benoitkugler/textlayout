package harfbuzz

import (
	"github.com/benoitkugler/textlayout/fonts"
	tt "github.com/benoitkugler/textlayout/fonts/truetype"
)

// ported from harfbuzz/src/hb-ot-layout-gpos-table.hh Copyright © 2007,2008,2009,2010  Red Hat, Inc.; 2010,2012,2013  Google, Inc.  Behdad Esfahbod

//  struct MarkArray;
//  static void Markclass_closure_and_remap_indexes (const Coverage  &mark_coverage,
// 						  const MarkArray &mark_array,
// 						  const hb_set_t  &glyphset,
// 						  hb_map_t*        klass_mapping /* INOUT */);

// attach_type_t
const (
	attachTypeNone = 0x00

	/* Each attachment should be either a mark or a cursive; can't be both. */
	attachTypeMark    = 0x01
	attachTypeCursive = 0x02
)

//  /* Shared Tables: ValueRecord, Anchor Table, and MarkArray */

//  typedef HBUINT16 Value;

//  typedef UnsizedArrayOf<Value> ValueRecord;

//  struct ValueFormat : HBUINT16
//  {
//    enum Flags {
// 	 xPlacement	= 0x0001u,	/* Includes horizontal adjustment for placement */
// 	 yPlacement	= 0x0002u,	/* Includes vertical adjustment for placement */
// 	 xAdvance	= 0x0004u,	/* Includes horizontal adjustment for advance */
// 	 yAdvance	= 0x0008u,	/* Includes vertical adjustment for advance */
// 	 xPlaDevice	= 0x0010u,	/* Includes horizontal Device table for placement */
// 	 yPlaDevice	= 0x0020u,	/* Includes vertical Device table for placement */
// 	 xAdvDevice	= 0x0040u,	/* Includes horizontal Device table for advance */
// 	 yAdvDevice	= 0x0080u,	/* Includes vertical Device table for advance */
// 	 ignored	= 0x0F00u,	/* Was used in TrueType Open for MM fonts */
// 	 reserved	= 0xF000u,	/* For future use */

// 	 devices	= 0x00F0u	/* Mask for having any Device table */
//    };

//  /* All fields are options.  Only those available advance the value pointer. */
//  #if 0
//    HBINT16		xPlacement;	/* Horizontal adjustment for
// 					  * placement--in design units */
//    HBINT16		yPlacement;	/* Vertical adjustment for
// 					  * placement--in design units */
//    HBINT16		xAdvance;	/* Horizontal adjustment for
// 					  * advance--in design units (only used
// 					  * for horizontal writing) */
//    HBINT16		yAdvance;	/* Vertical adjustment for advance--in
// 					  * design units (only used for vertical
// 					  * writing) */
//    OffsetTo<Device>	xPlaDevice;	/* Offset to Device table for
// 					  * horizontal placement--measured from
// 					  * beginning of PosTable (may be NULL) */
//    OffsetTo<Device>	yPlaDevice;	/* Offset to Device table for vertical
// 					  * placement--measured from beginning
// 					  * of PosTable (may be NULL) */
//    OffsetTo<Device>	xAdvDevice;	/* Offset to Device table for
// 					  * horizontal advance--measured from
// 					  * beginning of PosTable (may be NULL) */
//    OffsetTo<Device>	yAdvDevice;	/* Offset to Device table for vertical
// 					  * advance--measured from beginning of
// 					  * PosTable (may be NULL) */
//  #endif

//    unsigned int get_len () const  { return hb_popcount ((unsigned int) *this); }
//    unsigned int get_size () const { return get_len () * Value::static_size; }

//    bool apply_value (hb_ot_apply_context_t *c,
// 			 const void            *base,
// 			 const Value           *values,
// 			 hb_glyphPosition_t   &glyphPos) const
//    {
// 	 bool ret = false;
// 	 unsigned int format = *this;
// 	 if (!format) return ret;

// 	 hb_font_t *font = c.font;
// 	 bool horizontal = (c.IsHorizontal()direction);

// 	 if (format & xPlacement) glyphPos.XOffset  += font.em_scale_x (get_short (values++, &ret));
// 	 if (format & yPlacement) glyphPos.YOffset  += font.em_scale_y (get_short (values++, &ret));
// 	 if (format & xAdvance) {
// 	   if (likely (horizontal)) glyphPos.XAdvance += font.em_scale_x (get_short (values, &ret));
// 	   values++;
// 	 }
// 	 /* YAdvance values grow downward but font-space grows upward, hence negation */
// 	 if (format & yAdvance) {
// 	   if (unlikely (!horizontal)) glyphPos.YAdvance -= font.em_scale_y (get_short (values, &ret));
// 	   values++;
// 	 }

// 	 if (!has_device ()) return ret;

// 	 bool useXDevice = font.x_ppem || font.num_coords;
// 	 bool useYDevice = font.y_ppem || font.num_coords;

// 	 if (!useXDevice && !useYDevice) return ret;

// 	 const VariationStore &store = c.var_store;

// 	 /* pixel . fractional pixel */
// 	 if (format & xPlaDevice) {
// 	   if (useXDevice) glyphPos.XOffset  += (base + get_device (values, &ret)).get_x_delta (font, store);
// 	   values++;
// 	 }
// 	 if (format & yPlaDevice) {
// 	   if (useYDevice) glyphPos.YOffset  += (base + get_device (values, &ret)).get_y_delta (font, store);
// 	   values++;
// 	 }
// 	 if (format & xAdvDevice) {
// 	   if (horizontal && useXDevice) glyphPos.XAdvance += (base + get_device (values, &ret)).get_x_delta (font, store);
// 	   values++;
// 	 }
// 	 if (format & yAdvDevice) {
// 	   /* YAdvance values grow downward but font-space grows upward, hence negation */
// 	   if (!horizontal && useYDevice) glyphPos.YAdvance -= (base + get_device (values, &ret)).get_y_delta (font, store);
// 	   values++;
// 	 }
// 	 return ret;
//    }

//    void serialize_copy (hb_serialize_context_t *c, const void *base,
// 				const Value *values, const hb_map_t *layout_variation_idx_map) const
//    {
// 	 unsigned int format = *this;
// 	 if (!format) return;

// 	 if (format & xPlacement) c.copy (*values++);
// 	 if (format & yPlacement) c.copy (*values++);
// 	 if (format & xAdvance)   c.copy (*values++);
// 	 if (format & yAdvance)   c.copy (*values++);

// 	 if (format & xPlaDevice) copy_device (c, base, values++, layout_variation_idx_map);
// 	 if (format & yPlaDevice) copy_device (c, base, values++, layout_variation_idx_map);
// 	 if (format & xAdvDevice) copy_device (c, base, values++, layout_variation_idx_map);
// 	 if (format & yAdvDevice) copy_device (c, base, values++, layout_variation_idx_map);
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c,
// 				   const void *base,
// 				   const hb_array_t<const Value>& values) const
//    {
// 	 unsigned format = *this;
// 	 unsigned i = 0;
// 	 if (format & xPlacement) i++;
// 	 if (format & yPlacement) i++;
// 	 if (format & xAdvance) i++;
// 	 if (format & yAdvance) i++;
// 	 if (format & xPlaDevice)
// 	 {
// 	   (base + get_device (&(values[i]))).collect_variation_indices (c.layout_variation_indices);
// 	   i++;
// 	 }

// 	 if (format & ValueFormat::yPlaDevice)
// 	 {
// 	   (base + get_device (&(values[i]))).collect_variation_indices (c.layout_variation_indices);
// 	   i++;
// 	 }

// 	 if (format & ValueFormat::xAdvDevice)
// 	 {

// 	   (base + get_device (&(values[i]))).collect_variation_indices (c.layout_variation_indices);
// 	   i++;
// 	 }

// 	 if (format & ValueFormat::yAdvDevice)
// 	 {

// 	   (base + get_device (&(values[i]))).collect_variation_indices (c.layout_variation_indices);
// 	   i++;
// 	 }
//    }

//    private:
//    bool sanitize_value_devices (hb_sanitize_context_t *c, const void *base, const Value *values) const
//    {
// 	 unsigned int format = *this;

// 	 if (format & xPlacement) values++;
// 	 if (format & yPlacement) values++;
// 	 if (format & xAdvance)   values++;
// 	 if (format & yAdvance)   values++;

// 	 if ((format & xPlaDevice) && !get_device (values++).sanitize (c, base)) return false;
// 	 if ((format & yPlaDevice) && !get_device (values++).sanitize (c, base)) return false;
// 	 if ((format & xAdvDevice) && !get_device (values++).sanitize (c, base)) return false;
// 	 if ((format & yAdvDevice) && !get_device (values++).sanitize (c, base)) return false;

// 	 return true;
//    }

//    static inline OffsetTo<Device>& get_device (Value* value)
//    {
// 	 return *static_cast<OffsetTo<Device> *> (value);
//    }
//    static inline const OffsetTo<Device>& get_device (const Value* value, bool *worked=nullptr)
//    {
// 	 if (worked) *worked |= bool (*value);
// 	 return *static_cast<const OffsetTo<Device> *> (value);
//    }

//    bool copy_device (hb_serialize_context_t *c, const void *base,
// 			 const Value *src_value, const hb_map_t *layout_variation_idx_map) const
//    {
// 	 Value	*dst_value = c.copy (*src_value);

// 	 if (!dst_value) return false;
// 	 if (*dst_value == 0) return true;

// 	 *dst_value = 0;
// 	 c.push ();
// 	 if ((base + get_device (src_value)).copy (c, layout_variation_idx_map))
// 	 {
// 	   c.add_link (*dst_value, c.pop_pack ());
// 	   return true;
// 	 }
// 	 else
// 	 {
// 	   c.pop_discard ();
// 	   return false;
// 	 }
//    }

//    static inline const HBINT16& get_short (const Value* value, bool *worked=nullptr)
//    {
// 	 if (worked) *worked |= bool (*value);
// 	 return *reinterpret_cast<const HBINT16 *> (value);
//    }

//    public:

//    bool has_device () const
//    {
// 	 unsigned int format = *this;
// 	 return (format & devices) != 0;
//    }

//    bool sanitize_value (hb_sanitize_context_t *c, const void *base, const Value *values) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_range (values, get_size ()) && (!has_device () || sanitize_value_devices (c, base, values)));
//    }

//    bool sanitize_values (hb_sanitize_context_t *c, const void *base, const Value *values, unsigned int count) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 unsigned int len = get_len ();

// 	 if (!c.check_range (values, count, get_size ())) return false;

// 	 if (!has_device ()) return true;

// 	 for (unsigned int i = 0; i < count; i++) {
// 	   if (!sanitize_value_devices (c, base, values))
// 	 return false;
// 	   values += len;
// 	 }

// 	 return true;
//    }

//    /* Just sanitize referenced Device tables.  Doesn't check the values themselves. */
//    bool sanitize_values_stride_unsafe (hb_sanitize_context_t *c, const void *base, const Value *values, unsigned int count, unsigned int stride) const
//    {
// 	 TRACE_SANITIZE (this);

// 	 if (!has_device ()) return true;

// 	 for (unsigned int i = 0; i < count; i++) {
// 	   if (!sanitize_value_devices (c, base, values))
// 	 return false;
// 	   values += stride;
// 	 }

// 	 return true;
//    }
//  };

//  template<typename Iterator>
//  static void SinglePos_serialize (hb_serialize_context_t *c,
// 				  const void *src,
// 				  Iterator it,
// 				  ValueFormat valFormat,
// 				  const hb_map_t *layout_variation_idx_map);

//  struct AnchorFormat1
//  {
//    void get_anchor (hb_ot_apply_context_t *c, hb_codepoint_t glyph_id HB_UNUSED,
// 			float *x, float *y) const
//    {
// 	 hb_font_t *font = c.font;
// 	 *x = font.em_fscale_x (xCoordinate);
// 	 *y = font.em_fscale_y (yCoordinate);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this));
//    }

//    AnchorFormat1* copy (hb_serialize_context_t *c) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 return (c.embed<AnchorFormat1> (this));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    FWORD		xCoordinate;		/* Horizontal value--in design units */
//    FWORD		yCoordinate;		/* Vertical value--in design units */
//    public:
//    DEFINE_SIZE_STATIC (6);
//  };

//  struct AnchorFormat2
//  {
//    void get_anchor (hb_ot_apply_context_t *c, hb_codepoint_t glyph_id,
// 			float *x, float *y) const
//    {
// 	 hb_font_t *font = c.font;

//  #ifdef HB_NO_HINTING
// 	 *x = font.em_fscale_x (xCoordinate);
// 	 *y = font.em_fscale_y (yCoordinate);
// 	 return;
//  #endif

// 	 unsigned int x_ppem = font.x_ppem;
// 	 unsigned int y_ppem = font.y_ppem;
// 	 Position cx = 0, cy = 0;
// 	 bool ret;

// 	 ret = (x_ppem || y_ppem) &&
// 	   font.get_glyph_contour_point_for_origin (glyph_id, anchorPoint, HB_DIRECTION_LTR, &cx, &cy);
// 	 *x = ret && x_ppem ? cx : font.em_fscale_x (xCoordinate);
// 	 *y = ret && y_ppem ? cy : font.em_fscale_y (yCoordinate);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this));
//    }

//    AnchorFormat2* copy (hb_serialize_context_t *c) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 return (c.embed<AnchorFormat2> (this));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 2 */
//    FWORD		xCoordinate;		/* Horizontal value--in design units */
//    FWORD		yCoordinate;		/* Vertical value--in design units */
//    HBUINT16	anchorPoint;		/* Index to glyph contour point */
//    public:
//    DEFINE_SIZE_STATIC (8);
//  };

//  struct AnchorFormat3
//  {
//    void get_anchor (hb_ot_apply_context_t *c, hb_codepoint_t glyph_id HB_UNUSED,
// 			float *x, float *y) const
//    {
// 	 hb_font_t *font = c.font;
// 	 *x = font.em_fscale_x (xCoordinate);
// 	 *y = font.em_fscale_y (yCoordinate);

// 	 if (font.x_ppem || font.num_coords)
// 	   *x += (this+xDeviceTable).get_x_delta (font, c.var_store);
// 	 if (font.y_ppem || font.num_coords)
// 	   *y += (this+yDeviceTable).get_y_delta (font, c.var_store);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this) && xDeviceTable.sanitize (c, this) && yDeviceTable.sanitize (c, this));
//    }

//    AnchorFormat3* copy (hb_serialize_context_t *c,
// 				const hb_map_t *layout_variation_idx_map) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 if (!layout_variation_idx_map) return (nullptr);

// 	 auto *out = c.embed<AnchorFormat3> (this);
// 	 if (unlikely (!out)) return (nullptr);

// 	 out.xDeviceTable.serialize_copy (c, xDeviceTable, this, 0, hb_serialize_context_t::Head, layout_variation_idx_map);
// 	 out.yDeviceTable.serialize_copy (c, yDeviceTable, this, 0, hb_serialize_context_t::Head, layout_variation_idx_map);
// 	 return (out);
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 (this+xDeviceTable).collect_variation_indices (c.layout_variation_indices);
// 	 (this+yDeviceTable).collect_variation_indices (c.layout_variation_indices);
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 3 */
//    FWORD		xCoordinate;		/* Horizontal value--in design units */
//    FWORD		yCoordinate;		/* Vertical value--in design units */
//    OffsetTo<Device>
// 		 xDeviceTable;		/* Offset to Device table for X
// 					  * coordinate-- from beginning of
// 					  * Anchor table (may be NULL) */
//    OffsetTo<Device>
// 		 yDeviceTable;		/* Offset to Device table for Y
// 					  * coordinate-- from beginning of
// 					  * Anchor table (may be NULL) */
//    public:
//    DEFINE_SIZE_STATIC (10);
//  };

//  struct Anchor
//  {
//    void get_anchor (hb_ot_apply_context_t *c, hb_codepoint_t glyph_id,
// 			float *x, float *y) const
//    {
// 	 *x = *y = 0;
// 	 switch (u.format) {
// 	 case 1: u.format1.get_anchor (c, glyph_id, x, y); return;
// 	 case 2: u.format2.get_anchor (c, glyph_id, x, y); return;
// 	 case 3: u.format3.get_anchor (c, glyph_id, x, y); return;
// 	 default:					      return;
// 	 }
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 if (!u.format.sanitize (c)) return false;
// 	 switch (u.format) {
// 	 case 1: return (u.format1.sanitize (c));
// 	 case 2: return (u.format2.sanitize (c));
// 	 case 3: return (u.format3.sanitize (c));
// 	 default:return true;
// 	 }
//    }

//    Anchor* copy (hb_serialize_context_t *c, const hb_map_t *layout_variation_idx_map) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 switch (u.format) {
// 	 case 1: return (reinterpret_cast<Anchor *> (u.format1.copy (c)));
// 	 case 2: return (reinterpret_cast<Anchor *> (u.format2.copy (c)));
// 	 case 3: return (reinterpret_cast<Anchor *> (u.format3.copy (c, layout_variation_idx_map)));
// 	 default:return (nullptr);
// 	 }
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 switch (u.format) {
// 	 case 1: case 2:
// 	   return;
// 	 case 3:
// 	   u.format3.collect_variation_indices (c);
// 	   return;
// 	 default: return;
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    AnchorFormat1		format1;
//    AnchorFormat2		format2;
//    AnchorFormat3		format3;
//    } u;
//    public:
//    DEFINE_SIZE_UNION (2, format);
//  };

//  struct AnchorMatrix
//  {
//    const Anchor& get_anchor (unsigned int row, unsigned int col,
// 				 unsigned int cols, bool *found) const
//    {
// 	 *found = false;
// 	 if (unlikely (row >= rows || col >= cols)) return Null (Anchor);
// 	 *found = !matrixZ[row * cols + col].is_null ();
// 	 return this+matrixZ[row * cols + col];
//    }

//    template <typename Iterator,
// 		 hb_requires (hb_is_iterator (Iterator))>
//    void collect_variation_indices (hb_collect_variation_indices_context_t *c,
// 				   Iterator index_iter) const
//    {
// 	 for (unsigned i : index_iter)
// 	   (this+matrixZ[i]).collect_variation_indices (c);
//    }

//    template <typename Iterator,
// 		 hb_requires (hb_is_iterator (Iterator))>
//    bool serialize (hb_serialize_context_t *c,
// 		   unsigned                num_rows,
// 		   AnchorMatrix const     *offset_matrix,
// 		   const hb_map_t         *layout_variation_idx_map,
// 		   Iterator                index_iter)
//    {
// 	 TRACE_SERIALIZE (this);
// 	 if (!index_iter) return false;
// 	 if (unlikely (!c.extend_min ((*this))))  return false;

// 	 this.rows = num_rows;
// 	 for (const unsigned i : index_iter)
// 	 {
// 	   auto *offset = c.embed (offset_matrix.matrixZ[i]);
// 	   if (!offset) return false;
// 	   offset.serialize_copy (c, offset_matrix.matrixZ[i],
// 				   offset_matrix, c.to_bias (this),
// 				   hb_serialize_context_t::Head,
// 				   layout_variation_idx_map);
// 	 }

// 	 return true;
//    }

//    bool subset (hb_subset_context_t *c,
// 			unsigned cols,
// 			const hb_map_t *klass_mapping) const
//    {
// 	 TRACE_SUBSET (this);
// 	 auto *out = c.serializer.start_embed (*this);

// 	 auto indexes =
// 	 + hb_range (rows * cols)
// 	 | hb_filter ([=] (unsigned index) { return klass_mapping.has (index % cols); })
// 	 ;

// 	 out.serialize (c.serializer,
// 					 (unsigned) rows,
// 					 this,
// 					 c.plan.layout_variation_idx_map,
// 					 indexes);
// 	 return true;
//    }

//    bool sanitize (hb_sanitize_context_t *c, unsigned int cols) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 if (!c.check_struct (this)) return false;
// 	 if (unlikely (hb_unsigned_mul_overflows (rows, cols))) return false;
// 	 unsigned int count = rows * cols;
// 	 if (!c.check_array (matrixZ.arrayZ, count)) return false;
// 	 for (unsigned int i = 0; i < count; i++)
// 	   if (!matrixZ[i].sanitize (c, this)) return false;
// 	 return true;
//    }

//    HBUINT16	rows;			/* Number of rows */
//    UnsizedArrayOf<OffsetTo<Anchor>>
// 		 matrixZ;		/* Matrix of offsets to Anchor tables--
// 					  * from beginning of AnchorMatrix table */
//    public:
//    DEFINE_SIZE_ARRAY (2, matrixZ);
//  };

//  struct MarkRecord
//  {
//    friend struct MarkArray;

//    unsigned get_class () const { return (unsigned) klass; }
//    bool sanitize (hb_sanitize_context_t *c, const void *base) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this) && markAnchor.sanitize (c, base));
//    }

//    MarkRecord *copy (hb_serialize_context_t *c,
// 			 const void             *src_base,
// 			 unsigned                dst_bias,
// 			 const hb_map_t         *klass_mapping,
// 			 const hb_map_t         *layout_variation_idx_map) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 auto *out = c.embed (this);
// 	 if (unlikely (!out)) return (nullptr);

// 	 out.klass = klass_mapping.get (klass);
// 	 out.markAnchor.serialize_copy (c, markAnchor, src_base, dst_bias, hb_serialize_context_t::Head, layout_variation_idx_map);
// 	 return (out);
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c,
// 				   const void *src_base) const
//    {
// 	 (src_base+markAnchor).collect_variation_indices (c);
//    }

//    protected:
//    HBUINT16	klass;			/* Class defined for this mark */
//    OffsetTo<Anchor>
// 		 markAnchor;		/* Offset to Anchor table--from
// 					  * beginning of MarkArray table */
//    public:
//    DEFINE_SIZE_STATIC (4);
//  };

//  struct MarkArray : ArrayOf<MarkRecord>	/* Array of MarkRecords--in Coverage order */
//  {
//    bool apply (hb_ot_apply_context_t *c,
// 		   unsigned int mark_index, unsigned int glyph_index,
// 		   const AnchorMatrix &anchors, unsigned int class_count,
// 		   unsigned int glyphPos) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 const MarkRecord &record = ArrayOf<MarkRecord>::operator[](mark_index);
// 	 unsigned int markClass = record.klass;

// 	 const Anchor& markAnchor = this + record.markAnchor;
// 	 bool found;
// 	 const Anchor& glyphAnchor = anchors.get_anchor (glyph_index, markClass, class_count, &found);
// 	 /* If this subtable doesn't have an anchor for this base and this class,
// 	  * return false such that the subsequent subtables have a chance at it. */
// 	 if (unlikely (!found)) return false;

// 	 float markX, markY, baseX, baseY;

// 	 buffer.unsafe_to_break (glyphPos, buffer.idx);
// 	 markAnchor.get_anchor (c, buffer.cur().codepoint, &markX, &markY);
// 	 glyphAnchor.get_anchor (c, buffer.Info[glyphPos].codepoint, &baseX, &baseY);

// 	 hb_glyphPosition_t &o = buffer.curPos();
// 	 o.XOffset = roundf (baseX - markX);
// 	 o.YOffset = roundf (baseY - markY);
// 	 o.attach_type = attachTypeMark;
// 	 o.attach_chain = (int) glyphPos - (int) buffer.idx;
// 	 buffer.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_GPOS_ATTACHMENT;

// 	 buffer.idx++;
// 	 return true;
//    }

//    template<typename Iterator,
// 		hb_requires (hb_is_source_of (Iterator, MarkRecord))>
//    bool serialize (hb_serialize_context_t *c,
// 		   const hb_map_t         *klass_mapping,
// 		   const hb_map_t         *layout_variation_idx_map,
// 		   const void             *base,
// 		   Iterator                it)
//    {
// 	 TRACE_SERIALIZE (this);
// 	 if (unlikely (!c.extend_min (*this))) return false;
// 	 if (unlikely (!c.check_assign (len, it.len ()))) return false;
// 	 c.copy_all (it, base, c.to_bias (this), klass_mapping, layout_variation_idx_map);
// 	 return true;
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (ArrayOf<MarkRecord>::sanitize (c, this));
//    }
//  };

//  /* Lookups */

//  struct SinglePosFormat1
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    { return (this+coverage).intersects (glyphs); }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}
//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 if (!valueFormat.has_device ()) return;

// 	 auto it =
// 	 + hb_iter (this+coverage)
// 	 | hb_filter (c.glyph_set)
// 	 ;

// 	 if (!it) return;
// 	 valueFormat.collect_variation_indices (c, this, values.as_array (valueFormat.get_len ()));
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    { if (unlikely (!(this+coverage).collect_coverage (c.input))) return; }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int index = (this+coverage).get_coverage  (buffer.cur().codepoint);
// 	 if (likely (index == NOT_COVERED)) return false;

// 	 valueFormat.apply_value (c, this, values, buffer.curPos());

// 	 buffer.idx++;
// 	 return true;
//    }

//    template<typename Iterator,
// 		hb_requires (hb_is_iterator (Iterator))>
//    void serialize (hb_serialize_context_t *c,
// 		   const void *src,
// 		   Iterator it,
// 		   ValueFormat valFormat,
// 		   const hb_map_t *layout_variation_idx_map)
//    {
// 	 auto out = c.extend_min (*this);
// 	 if (unlikely (!out)) return;
// 	 if (unlikely (!c.check_assign (valueFormat, valFormat))) return;

// 	 + it
// 	 | hb_map (hb_second)
// 	 | hb_apply ([&] (hb_array_t<const Value> _)
// 		 { valFormat.serialize_copy (c, src, &_, layout_variation_idx_map); })
// 	 ;

// 	 auto glyphs =
// 	 + it
// 	 | hb_map_retains_sorting (hb_first)
// 	 ;

// 	 coverage.serialize (c, this).serialize (c, glyphs);
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto it =
// 	 + hb_iter (this+coverage)
// 	 | hb_filter (glyphset)
// 	 | hb_map_retains_sorting (glyph_map)
// 	 | hb_zip (hb_repeat (values.as_array (valueFormat.get_len ())))
// 	 ;

// 	 bool ret = bool (it);
// 	 SinglePos_serialize (c.serializer, this, it, valueFormat, c.plan.layout_variation_idx_map);
// 	 return (ret);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this) &&
// 		   coverage.sanitize (c, this) &&
// 		   valueFormat.sanitize_value (c, this, values));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    OffsetTo<Coverage>
// 		 coverage;		/* Offset to Coverage table--from
// 					  * beginning of subtable */
//    ValueFormat	valueFormat;		/* Defines the types of data in the
// 					  * ValueRecord */
//    ValueRecord	values;			/* Defines positioning
// 					  * value(s)--applied to all glyphs in
// 					  * the Coverage table */
//    public:
//    DEFINE_SIZE_ARRAY (6, values);
//  };

//  struct SinglePosFormat2
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    { return (this+coverage).intersects (glyphs); }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}
//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 if (!valueFormat.has_device ()) return;

// 	 auto it =
// 	 + hb_zip (this+coverage, hb_range ((unsigned) valueCount))
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 ;

// 	 if (!it) return;

// 	 unsigned sub_length = valueFormat.get_len ();
// 	 const hb_array_t<const Value> values_array = values.as_array (valueCount * sub_length);

// 	 for (unsigned i : + it
// 			   | hb_map (hb_second))
// 	   valueFormat.collect_variation_indices (c, this, values_array.sub_array (i * sub_length, sub_length));

//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    { if (unlikely (!(this+coverage).collect_coverage (c.input))) return; }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int index = (this+coverage).get_coverage  (buffer.cur().codepoint);
// 	 if (likely (index == NOT_COVERED)) return false;

// 	 if (likely (index >= valueCount)) return false;

// 	 valueFormat.apply_value (c, this,
// 				  &values[index * valueFormat.get_len ()],
// 				  buffer.curPos());

// 	 buffer.idx++;
// 	 return true;
//    }

//    template<typename Iterator,
// 		hb_requires (hb_is_iterator (Iterator))>
//    void serialize (hb_serialize_context_t *c,
// 		   const void *src,
// 		   Iterator it,
// 		   ValueFormat valFormat,
// 		   const hb_map_t *layout_variation_idx_map)
//    {
// 	 auto out = c.extend_min (*this);
// 	 if (unlikely (!out)) return;
// 	 if (unlikely (!c.check_assign (valueFormat, valFormat))) return;
// 	 if (unlikely (!c.check_assign (valueCount, it.len ()))) return;

// 	 + it
// 	 | hb_map (hb_second)
// 	 | hb_apply ([&] (hb_array_t<const Value> _)
// 		 { valFormat.serialize_copy (c, src, &_, layout_variation_idx_map); })
// 	 ;

// 	 auto glyphs =
// 	 + it
// 	 | hb_map_retains_sorting (hb_first)
// 	 ;

// 	 coverage.serialize (c, this).serialize (c, glyphs);
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 unsigned sub_length = valueFormat.get_len ();
// 	 auto values_array = values.as_array (valueCount * sub_length);

// 	 auto it =
// 	 + hb_zip (this+coverage, hb_range ((unsigned) valueCount))
// 	 | hb_filter (glyphset, hb_first)
// 	 | hb_map_retains_sorting ([&] (const hb_pair_t<hb_codepoint_t, unsigned>& _)
// 				   {
// 				 return hb_pair (glyph_map[_.first],
// 						 values_array.sub_array (_.second * sub_length,
// 									 sub_length));
// 				   })
// 	 ;

// 	 bool ret = bool (it);
// 	 SinglePos_serialize (c.serializer, this, it, valueFormat, c.plan.layout_variation_idx_map);
// 	 return (ret);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this) &&
// 		   coverage.sanitize (c, this) &&
// 		   valueFormat.sanitize_values (c, this, values, valueCount));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 2 */
//    OffsetTo<Coverage>
// 		 coverage;		/* Offset to Coverage table--from
// 					  * beginning of subtable */
//    ValueFormat	valueFormat;		/* Defines the types of data in the
// 					  * ValueRecord */
//    HBUINT16	valueCount;		/* Number of ValueRecords */
//    ValueRecord	values;			/* Array of ValueRecords--positioning
// 					  * values applied to glyphs */
//    public:
//    DEFINE_SIZE_ARRAY (8, values);
//  };

//  struct SinglePos
//  {
//    template<typename Iterator,
// 		hb_requires (hb_is_iterator (Iterator))>
//    unsigned get_format (Iterator glyph_val_iter_pairs)
//    {
// 	 hb_array_t<const Value> first_val_iter = hb_second (*glyph_val_iter_pairs);

// 	 for (const auto iter : glyph_val_iter_pairs)
// 	   for (const auto _ : hb_zip (iter.second, first_val_iter))
// 	 if (_.first != _.second)
// 	   return 2;

// 	 return 1;
//    }

//    template<typename Iterator,
// 		hb_requires (hb_is_iterator (Iterator))>
//    void serialize (hb_serialize_context_t *c,
// 		   const void *src,
// 		   Iterator glyph_val_iter_pairs,
// 		   ValueFormat valFormat,
// 		   const hb_map_t *layout_variation_idx_map)
//    {
// 	 if (unlikely (!c.extend_min (u.format))) return;
// 	 unsigned format = 2;

// 	 if (glyph_val_iter_pairs) format = get_format (glyph_val_iter_pairs);

// 	 u.format = format;
// 	 switch (u.format) {
// 	 case 1: u.format1.serialize (c, src, glyph_val_iter_pairs, valFormat, layout_variation_idx_map);
// 		 return;
// 	 case 2: u.format2.serialize (c, src, glyph_val_iter_pairs, valFormat, layout_variation_idx_map);
// 		 return;
// 	 default:return;
// 	 }
//    }

//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 case 2: return (c.dispatch (u.format2, hb_forward<Ts> (ds)...));
// 	 default:return (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    SinglePosFormat1	format1;
//    SinglePosFormat2	format2;
//    } u;
//  };

//  template<typename Iterator>
//  static void
//  SinglePos_serialize (hb_serialize_context_t *c,
// 			  const void *src,
// 			  Iterator it,
// 			  ValueFormat valFormat,
// 			  const hb_map_t *layout_variation_idx_map)
//  { c.start_embed<SinglePos> ().serialize (c, src, it, valFormat, layout_variation_idx_map); }

//  struct PairValueRecord
//  {
//    friend struct PairSet;

//    int cmp (hb_codepoint_t k) const
//    { return secondGlyph.cmp (k); }

//    struct serialize_closure_t
//    {
// 	 const void 		*base;
// 	 const ValueFormat	*valueFormats;
// 	 unsigned		len1; /* valueFormats[0].get_len() */
// 	 const hb_map_t 	*glyph_map;
// 	 const hb_map_t      *layout_variation_idx_map;
//    };

//    bool serialize (hb_serialize_context_t *c,
// 		   serialize_closure_t *closure) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 auto *out = c.start_embed (*this);
// 	 if (unlikely (!c.extend_min (out))) return false;

// 	 out.secondGlyph = (*closure.glyph_map)[secondGlyph];

// 	 closure.valueFormats[0].serialize_copy (c, closure.base, &values[0], closure.layout_variation_idx_map);
// 	 closure.valueFormats[1].serialize_copy (c, closure.base, &values[closure.len1], closure.layout_variation_idx_map);

// 	 return true;
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c,
// 				   const ValueFormat *valueFormats,
// 				   const void *base) const
//    {
// 	 unsigned record1_len = valueFormats[0].get_len ();
// 	 unsigned record2_len = valueFormats[1].get_len ();
// 	 const hb_array_t<const Value> values_array = values.as_array (record1_len + record2_len);

// 	 if (valueFormats[0].has_device ())
// 	   valueFormats[0].collect_variation_indices (c, base, values_array.sub_array (0, record1_len));

// 	 if (valueFormats[1].has_device ())
// 	   valueFormats[1].collect_variation_indices (c, base, values_array.sub_array (record1_len, record2_len));
//    }

//    protected:
//    HBGlyphID	secondGlyph;		/* GlyphID of second glyph in the
// 					  * pair--first glyph is listed in the
// 					  * Coverage table */
//    ValueRecord	values;			/* Positioning data for the first glyph
// 					  * followed by for second glyph */
//    public:
//    DEFINE_SIZE_ARRAY (2, values);
//  };

//  struct PairSet
//  {
//    friend struct PairPosFormat1;

//    bool intersects (const hb_set_t *glyphs,
// 			const ValueFormat *valueFormats) const
//    {
// 	 unsigned int len1 = valueFormats[0].get_len ();
// 	 unsigned int len2 = valueFormats[1].get_len ();
// 	 unsigned int record_size = HBUINT16::static_size * (1 + len1 + len2);

// 	 const PairValueRecord *record = &firstPairValueRecord;
// 	 unsigned int count = len;
// 	 for (unsigned int i = 0; i < count; i++)
// 	 {
// 	   if (glyphs.has (record.secondGlyph))
// 	 return true;
// 	   record = &StructAtOffset<const PairValueRecord> (record, record_size);
// 	 }
// 	 return false;
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c,
// 				const ValueFormat *valueFormats) const
//    {
// 	 unsigned int len1 = valueFormats[0].get_len ();
// 	 unsigned int len2 = valueFormats[1].get_len ();
// 	 unsigned int record_size = HBUINT16::static_size * (1 + len1 + len2);

// 	 const PairValueRecord *record = &firstPairValueRecord;
// 	 c.input.add_array (&record.secondGlyph, len, record_size);
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c,
// 				   const ValueFormat *valueFormats) const
//    {
// 	 unsigned len1 = valueFormats[0].get_len ();
// 	 unsigned len2 = valueFormats[1].get_len ();
// 	 unsigned record_size = HBUINT16::static_size * (1 + len1 + len2);

// 	 const PairValueRecord *record = &firstPairValueRecord;
// 	 unsigned count = len;
// 	 for (unsigned i = 0; i < count; i++)
// 	 {
// 	   if (c.glyph_set.has (record.secondGlyph))
// 	   { record.collect_variation_indices (c, valueFormats, this); }

// 	   record = &StructAtOffset<const PairValueRecord> (record, record_size);
// 	 }
//    }

//    bool apply (hb_ot_apply_context_t *c,
// 		   const ValueFormat *valueFormats,
// 		   unsigned int pos) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int len1 = valueFormats[0].get_len ();
// 	 unsigned int len2 = valueFormats[1].get_len ();
// 	 unsigned int record_size = HBUINT16::static_size * (1 + len1 + len2);

// 	 const PairValueRecord *record = hb_bsearch (buffer.Info[pos].codepoint,
// 						 &firstPairValueRecord,
// 						 len,
// 						 record_size);
// 	 if (record)
// 	 {
// 	   /* Note the intentional use of "|" instead of short-circuit "||". */
// 	   if (valueFormats[0].apply_value (c, this, &record.values[0], buffer.curPos()) |
// 	   valueFormats[1].apply_value (c, this, &record.values[len1], buffer.Pos[pos]))
// 	 buffer.unsafe_to_break (buffer.idx, pos + 1);
// 	   if (len2)
// 	 pos++;
// 	   buffer.idx = pos;
// 	   return true;
// 	 }
// 	 return false;
//    }

//    bool subset (hb_subset_context_t *c,
// 			const ValueFormat valueFormats[2]) const
//    {
// 	 TRACE_SUBSET (this);
// 	 auto snap = c.serializer.snapshot ();

// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return false;
// 	 out.len = 0;

// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 unsigned len1 = valueFormats[0].get_len ();
// 	 unsigned len2 = valueFormats[1].get_len ();
// 	 unsigned record_size = HBUINT16::static_size + Value::static_size * (len1 + len2);

// 	 PairValueRecord::serialize_closure_t closure =
// 	 {
// 	   this,
// 	   valueFormats,
// 	   len1,
// 	   &glyph_map,
// 	   c.plan.layout_variation_idx_map
// 	 };

// 	 const PairValueRecord *record = &firstPairValueRecord;
// 	 unsigned count = len, num = 0;
// 	 for (unsigned i = 0; i < count; i++)
// 	 {
// 	   if (glyphset.has (record.secondGlyph)
// 	  && record.serialize (c.serializer, &closure)) num++;
// 	   record = &StructAtOffset<const PairValueRecord> (record, record_size);
// 	 }

// 	 out.len = num;
// 	 if (!num) c.serializer.revert (snap);
// 	 return (num);
//    }

//    struct sanitize_closure_t
//    {
// 	 const ValueFormat *valueFormats;
// 	 unsigned int len1; /* valueFormats[0].get_len() */
// 	 unsigned int stride; /* 1 + len1 + len2 */
//    };

//    bool sanitize (hb_sanitize_context_t *c, const sanitize_closure_t *closure) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 if (!(c.check_struct (this)
// 		&& c.check_range (&firstPairValueRecord,
// 			   len,
// 			   HBUINT16::static_size,
// 			   closure.stride))) return false;

// 	 unsigned int count = len;
// 	 const PairValueRecord *record = &firstPairValueRecord;
// 	 return (closure.valueFormats[0].sanitize_values_stride_unsafe (c, this, &record.values[0], count, closure.stride) &&
// 		   closure.valueFormats[1].sanitize_values_stride_unsafe (c, this, &record.values[closure.len1], count, closure.stride));
//    }

//    protected:
//    HBUINT16		len;	/* Number of PairValueRecords */
//    PairValueRecord	firstPairValueRecord;
// 				 /* Array of PairValueRecords--ordered
// 				  * by GlyphID of the second glyph */
//    public:
//    DEFINE_SIZE_MIN (2);
//  };

//  struct PairPosFormat1
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 return
// 	 + hb_zip (this+coverage, pairSet)
// 	 | hb_filter (*glyphs, hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_map ([glyphs, this] (const OffsetTo<PairSet> &_)
// 		   { return (this+_).intersects (glyphs, valueFormat); })
// 	 | hb_any
// 	 ;
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}
//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 if ((!valueFormat[0].has_device ()) && (!valueFormat[1].has_device ())) return;

// 	 auto it =
// 	 + hb_zip (this+coverage, pairSet)
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 ;

// 	 if (!it) return;
// 	 + it
// 	 | hb_map (hb_add (this))
// 	 | hb_apply ([&] (const PairSet& _) { _.collect_variation_indices (c, valueFormat); })
// 	 ;
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 if (unlikely (!(this+coverage).collect_coverage (c.input))) return;
// 	 unsigned int count = pairSet.len;
// 	 for (unsigned int i = 0; i < count; i++)
// 	   (this+pairSet[i]).collect_glyphs (c, valueFormat);
//    }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int index = (this+coverage).get_coverage  (buffer.cur().codepoint);
// 	 if (likely (index == NOT_COVERED)) return false;

// 	 hb_ot_apply_context_t::skipping_iterator_t &skippyIter = c.iter_input;
// 	 skippyIter.reset (buffer.idx, 1);
// 	 if (!skippyIter.next ()) return false;

// 	 return ((this+pairSet[index]).apply (c, valueFormat, skippyIter.idx));
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);

// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return false;
// 	 out.format = format;
// 	 out.valueFormat[0] = valueFormat[0];
// 	 out.valueFormat[1] = valueFormat[1];

// 	 hb_sorted_vector_t<hb_codepoint_t> new_coverage;

// 	 + hb_zip (this+coverage, pairSet)
// 	 | hb_filter (glyphset, hb_first)
// 	 | hb_filter ([this, c, out] (const OffsetTo<PairSet>& _)
// 		  {
// 			auto *o = out.pairSet.serialize_append (c.serializer);
// 			if (unlikely (!o)) return false;
// 			auto snap = c.serializer.snapshot ();
// 			bool ret = o.serialize_subset (c, _, this, valueFormat);
// 			if (!ret)
// 			{
// 			  out.pairSet.pop ();
// 			  c.serializer.revert (snap);
// 			}
// 			return ret;
// 		  },
// 		  hb_second)
// 	 | hb_map (hb_first)
// 	 | hb_map (glyph_map)
// 	 | hb_sink (new_coverage)
// 	 ;

// 	 out.coverage.serialize (c.serializer, out)
// 		  .serialize (c.serializer, new_coverage.iter ());

// 	 return (bool (new_coverage));
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);

// 	 if (!c.check_struct (this)) return false;

// 	 unsigned int len1 = valueFormat[0].get_len ();
// 	 unsigned int len2 = valueFormat[1].get_len ();
// 	 PairSet::sanitize_closure_t closure =
// 	 {
// 	   valueFormat,
// 	   len1,
// 	   1 + len1 + len2
// 	 };

// 	 return (coverage.sanitize (c, this) && pairSet.sanitize (c, this, &closure));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    OffsetTo<Coverage>
// 		 coverage;		/* Offset to Coverage table--from
// 					  * beginning of subtable */
//    ValueFormat	valueFormat[2];		/* [0] Defines the types of data in
// 					  * ValueRecord1--for the first glyph
// 					  * in the pair--may be zero (0) */
// 					 /* [1] Defines the types of data in
// 					  * ValueRecord2--for the second glyph
// 					  * in the pair--may be zero (0) */
//    OffsetArrayOf<PairSet>
// 		 pairSet;		/* Array of PairSet tables
// 					  * ordered by Coverage Index */
//    public:
//    DEFINE_SIZE_ARRAY (10, pairSet);
//  };

//  struct PairPosFormat2
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 return (this+coverage).intersects (glyphs) &&
// 		(this+classDef2).intersects (glyphs);
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}
//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 if ((!valueFormat1.has_device ()) && (!valueFormat2.has_device ())) return;

// 	 hb_set_t class1_set, class2_set;
// 	 for (const unsigned cp : c.glyph_set.iter ())
// 	 {
// 	   unsigned klass1 = (this+classDef1).get (cp);
// 	   unsigned klass2 = (this+classDef2).get (cp);
// 	   class1_set.add (klass1);
// 	   class2_set.add (klass2);
// 	 }

// 	 if (class1_set.is_empty () || class2_set.is_empty ()) return;

// 	 unsigned len1 = valueFormat1.get_len ();
// 	 unsigned len2 = valueFormat2.get_len ();
// 	 const hb_array_t<const Value> values_array = values.as_array ((unsigned)class1Count * (unsigned) class2Count * (len1 + len2));
// 	 for (const unsigned class1_idx : class1_set.iter ())
// 	 {
// 	   for (const unsigned class2_idx : class2_set.iter ())
// 	   {
// 	 unsigned start_offset = (class1_idx * (unsigned) class2Count + class2_idx) * (len1 + len2);
// 	 if (valueFormat1.has_device ())
// 	   valueFormat1.collect_variation_indices (c, this, values_array.sub_array (start_offset, len1));

// 	 if (valueFormat2.has_device ())
// 	   valueFormat2.collect_variation_indices (c, this, values_array.sub_array (start_offset+len1, len2));
// 	   }
// 	 }
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 if (unlikely (!(this+coverage).collect_coverage (c.input))) return;
// 	 if (unlikely (!(this+classDef2).collect_coverage (c.input))) return;
//    }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int index = (this+coverage).get_coverage  (buffer.cur().codepoint);
// 	 if (likely (index == NOT_COVERED)) return false;

// 	 hb_ot_apply_context_t::skipping_iterator_t &skippyIter = c.iter_input;
// 	 skippyIter.reset (buffer.idx, 1);
// 	 if (!skippyIter.next ()) return false;

// 	 unsigned int len1 = valueFormat1.get_len ();
// 	 unsigned int len2 = valueFormat2.get_len ();
// 	 unsigned int record_len = len1 + len2;

// 	 unsigned int klass1 = (this+classDef1).get_class (buffer.cur().codepoint);
// 	 unsigned int klass2 = (this+classDef2).get_class (buffer.Info[skippyIter.idx].codepoint);
// 	 if (unlikely (klass1 >= class1Count || klass2 >= class2Count)) return false;

// 	 const Value *v = &values[record_len * (klass1 * class2Count + klass2)];
// 	 /* Note the intentional use of "|" instead of short-circuit "||". */
// 	 if (valueFormat1.apply_value (c, this, v, buffer.curPos()) |
// 	 valueFormat2.apply_value (c, this, v + len1, buffer.Pos[skippyIter.idx]))
// 	   buffer.unsafe_to_break (buffer.idx, skippyIter.idx + 1);

// 	 buffer.idx = skippyIter.idx;
// 	 if (len2)
// 	   buffer.idx++;

// 	 return true;
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return false;
// 	 out.format = format;
// 	 out.valueFormat1 = valueFormat1;
// 	 out.valueFormat2 = valueFormat2;

// 	 hb_map_t klass1_map;
// 	 out.classDef1.serialize_subset (c, classDef1, this, &klass1_map);
// 	 out.class1Count = klass1_map.get_population ();

// 	 hb_map_t klass2_map;
// 	 out.classDef2.serialize_subset (c, classDef2, this, &klass2_map);
// 	 out.class2Count = klass2_map.get_population ();

// 	 unsigned len1 = valueFormat1.get_len ();
// 	 unsigned len2 = valueFormat2.get_len ();

// 	 + hb_range ((unsigned) class1Count)
// 	 | hb_filter (klass1_map)
// 	 | hb_apply ([&] (const unsigned class1_idx)
// 		 {
// 		   + hb_range ((unsigned) class2Count)
// 		   | hb_filter (klass2_map)
// 		   | hb_apply ([&] (const unsigned class2_idx)
// 				   {
// 				 unsigned idx = (class1_idx * (unsigned) class2Count + class2_idx) * (len1 + len2);
// 				 valueFormat1.serialize_copy (c.serializer, this, &values[idx], c.plan.layout_variation_idx_map);
// 				 valueFormat2.serialize_copy (c.serializer, this, &values[idx + len1], c.plan.layout_variation_idx_map);
// 				   })
// 		   ;
// 		 })
// 	 ;

// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto it =
// 	 + hb_iter (this+coverage)
// 	 | hb_filter (glyphset)
// 	 | hb_map_retains_sorting (glyph_map)
// 	 ;

// 	 out.coverage.serialize (c.serializer, out).serialize (c.serializer, it);
// 	 return (out.class1Count && out.class2Count && bool (it));
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 if (!(c.check_struct (this)
// 		&& coverage.sanitize (c, this)
// 		&& classDef1.sanitize (c, this)
// 		&& classDef2.sanitize (c, this))) return false;

// 	 unsigned int len1 = valueFormat1.get_len ();
// 	 unsigned int len2 = valueFormat2.get_len ();
// 	 unsigned int stride = len1 + len2;
// 	 unsigned int record_size = valueFormat1.get_size () + valueFormat2.get_size ();
// 	 unsigned int count = (unsigned int) class1Count * (unsigned int) class2Count;
// 	 return (c.check_range ((const void *) values,
// 				   count,
// 				   record_size) &&
// 		   valueFormat1.sanitize_values_stride_unsafe (c, this, &values[0], count, stride) &&
// 		   valueFormat2.sanitize_values_stride_unsafe (c, this, &values[len1], count, stride));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 2 */
//    OffsetTo<Coverage>
// 		 coverage;		/* Offset to Coverage table--from
// 					  * beginning of subtable */
//    ValueFormat	valueFormat1;		/* ValueRecord definition--for the
// 					  * first glyph of the pair--may be zero
// 					  * (0) */
//    ValueFormat	valueFormat2;		/* ValueRecord definition--for the
// 					  * second glyph of the pair--may be
// 					  * zero (0) */
//    OffsetTo<ClassDef>
// 		 classDef1;		/* Offset to ClassDef table--from
// 					  * beginning of PairPos subtable--for
// 					  * the first glyph of the pair */
//    OffsetTo<ClassDef>
// 		 classDef2;		/* Offset to ClassDef table--from
// 					  * beginning of PairPos subtable--for
// 					  * the second glyph of the pair */
//    HBUINT16	class1Count;		/* Number of classes in ClassDef1
// 					  * table--includes Class0 */
//    HBUINT16	class2Count;		/* Number of classes in ClassDef2
// 					  * table--includes Class0 */
//    ValueRecord	values;			/* Matrix of value pairs:
// 					  * class1-major, class2-minor,
// 					  * Each entry has value1 and value2 */
//    public:
//    DEFINE_SIZE_ARRAY (16, values);
//  };

//  struct PairPos
//  {
//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 case 2: return (c.dispatch (u.format2, hb_forward<Ts> (ds)...));
// 	 default:return (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    PairPosFormat1	format1;
//    PairPosFormat2	format2;
//    } u;
//  };

//  struct EntryExitRecord
//  {
//    friend struct CursivePosFormat1;

//    bool sanitize (hb_sanitize_context_t *c, const void *base) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (entryAnchor.sanitize (c, base) && exitAnchor.sanitize (c, base));
//    }

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c,
// 				   const void *src_base) const
//    {
// 	 (src_base+entryAnchor).collect_variation_indices (c);
// 	 (src_base+exitAnchor).collect_variation_indices (c);
//    }

//    EntryExitRecord* copy (hb_serialize_context_t *c,
// 			  const void *src_base,
// 			  const void *dst_base,
// 			  const hb_map_t *layout_variation_idx_map) const
//    {
// 	 TRACE_SERIALIZE (this);
// 	 auto *out = c.embed (this);
// 	 if (unlikely (!out)) return (nullptr);

// 	 out.entryAnchor.serialize_copy (c, entryAnchor, src_base, c.to_bias (dst_base), hb_serialize_context_t::Head, layout_variation_idx_map);
// 	 out.exitAnchor.serialize_copy (c, exitAnchor, src_base, c.to_bias (dst_base), hb_serialize_context_t::Head, layout_variation_idx_map);
// 	 return (out);
//    }

//    protected:
//    OffsetTo<Anchor>
// 		 entryAnchor;		/* Offset to EntryAnchor table--from
// 					  * beginning of CursivePos
// 					  * subtable--may be NULL */
//    OffsetTo<Anchor>
// 		 exitAnchor;		/* Offset to ExitAnchor table--from
// 					  * beginning of CursivePos
// 					  * subtable--may be NULL */
//    public:
//    DEFINE_SIZE_STATIC (4);
//  };

//  struct CursivePosFormat1
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    { return (this+coverage).intersects (glyphs); }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 + hb_zip (this+coverage, entryExitRecord)
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_apply ([&] (const EntryExitRecord& record) { record.collect_variation_indices (c, this); })
// 	 ;
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    { if (unlikely (!(this+coverage).collect_coverage (c.input))) return; }

//    const Coverage &get_coverage () const { return this+coverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;

// 	 const EntryExitRecord &thisRecord = entryExitRecord[(this+coverage).get_coverage  (buffer.cur().codepoint)];
// 	 if (!thisRecord.entryAnchor) return false;

// 	 hb_ot_apply_context_t::skipping_iterator_t &skippyIter = c.iter_input;
// 	 skippyIter.reset (buffer.idx, 1);
// 	 if (!skippyIter.prev ()) return false;

// 	 const EntryExitRecord &prevRecord = entryExitRecord[(this+coverage).get_coverage  (buffer.Info[skippyIter.idx].codepoint)];
// 	 if (!prevRecord.exitAnchor) return false;

// 	 unsigned int i = skippyIter.idx;
// 	 unsigned int j = buffer.idx;

// 	 buffer.unsafe_to_break (i, j);
// 	 float entryX, entryY, exitX, exitY;
// 	 (this+prevRecord.exitAnchor).get_anchor (c, buffer.Info[i].codepoint, &exitX, &exitY);
// 	 (this+thisRecord.entryAnchor).get_anchor (c, buffer.Info[j].codepoint, &entryX, &entryY);

// 	 hb_glyphPosition_t *pos = buffer.Pos;

// 	 Position d;
// 	 /* Main-direction adjustment */
// 	 switch (c.direction) {
// 	   case HB_DIRECTION_LTR:
// 	 pos[i].XAdvance  = roundf (exitX) + pos[i].XOffset;

// 	 d = roundf (entryX) + pos[j].XOffset;
// 	 pos[j].XAdvance -= d;
// 	 pos[j].XOffset  -= d;
// 	 break;
// 	   case HB_DIRECTION_RTL:
// 	 d = roundf (exitX) + pos[i].XOffset;
// 	 pos[i].XAdvance -= d;
// 	 pos[i].XOffset  -= d;

// 	 pos[j].XAdvance  = roundf (entryX) + pos[j].XOffset;
// 	 break;
// 	   case HB_DIRECTION_TTB:
// 	 pos[i].YAdvance  = roundf (exitY) + pos[i].YOffset;

// 	 d = roundf (entryY) + pos[j].YOffset;
// 	 pos[j].YAdvance -= d;
// 	 pos[j].YOffset  -= d;
// 	 break;
// 	   case HB_DIRECTION_BTT:
// 	 d = roundf (exitY) + pos[i].YOffset;
// 	 pos[i].YAdvance -= d;
// 	 pos[i].YOffset  -= d;

// 	 pos[j].YAdvance  = roundf (entryY);
// 	 break;
// 	   case HB_DIRECTION_INVALID:
// 	   default:
// 	 break;
// 	 }

// 	 /* Cross-direction adjustment */

// 	 /* We attach child to parent (think graph theory and rooted trees whereas
// 	  * the root stays on baseline and each node aligns itself against its
// 	  * parent.
// 	  *
// 	  * Optimize things for the case of RightToLeft, as that's most common in
// 	  * Arabic. */
// 	 unsigned int child  = i;
// 	 unsigned int parent = j;
// 	 Position XOffset = entryX - exitX;
// 	 Position YOffset = entryY - exitY;
// 	 if  (!(c.lookup_props & LookupFlag::RightToLeft))
// 	 {
// 	   unsigned int k = child;
// 	   child = parent;
// 	   parent = k;
// 	   XOffset = -XOffset;
// 	   YOffset = -YOffset;
// 	 }

// 	 /* If child was already connected to someone else, walk through its old
// 	  * chain and reverse the link direction, such that the whole tree of its
// 	  * previous connection now attaches to new parent.  Watch out for case
// 	  * where new parent is on the path from old chain...
// 	  */
// 	 reverseCursiveMinorOffset (pos, child, c.direction, parent);

// 	 pos[child].attach_type = attachTypeCursive;
// 	 pos[child].attach_chain = (int) parent - (int) child;
// 	 buffer.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_GPOS_ATTACHMENT;
// 	 if (likely ((.IsHorizontal().direction)))
// 	   pos[child].YOffset = YOffset;
// 	 else
// 	   pos[child].XOffset = XOffset;

// 	 /* If parent was attached to child, break them free.
// 	  * https://github.com/harfbuzz/harfbuzz/issues/2469
// 	  */
// 	 if (unlikely (pos[parent].attach_chain == -pos[child].attach_chain))
// 	   pos[parent].attach_chain = 0;

// 	 buffer.idx++;
// 	 return true;
//    }

//    template <typename Iterator,
// 		 hb_requires (hb_is_iterator (Iterator))>
//    void serialize (hb_serialize_context_t *c,
// 		   Iterator it,
// 		   const void *src_base,
// 		   const hb_map_t *layout_variation_idx_map)
//    {
// 	 if (unlikely (!c.extend_min ((*this)))) return;
// 	 this.format = 1;
// 	 this.entryExitRecord.len = it.len ();

// 	 for (const EntryExitRecord& entry_record : + it
// 							| hb_map (hb_second))
// 	   c.copy (entry_record, src_base, this, layout_variation_idx_map);

// 	 auto glyphs =
// 	 + it
// 	 | hb_map_retains_sorting (hb_first)
// 	 ;

// 	 coverage.serialize (c, this).serialize (c, glyphs);
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!out)) return false;

// 	 auto it =
// 	 + hb_zip (this+coverage, entryExitRecord)
// 	 | hb_filter (glyphset, hb_first)
// 	 | hb_map_retains_sorting ([&] (hb_pair_t<hb_codepoint_t, const EntryExitRecord&> p) . hb_pair_t<hb_codepoint_t, const EntryExitRecord&>
// 				   { return hb_pair (glyph_map[p.first], p.second);})
// 	 ;

// 	 bool ret = bool (it);
// 	 out.serialize (c.serializer, it, this, c.plan.layout_variation_idx_map);
// 	 return (ret);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (coverage.sanitize (c, this) && entryExitRecord.sanitize (c, this));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    OffsetTo<Coverage>
// 		 coverage;		/* Offset to Coverage table--from
// 					  * beginning of subtable */
//    ArrayOf<EntryExitRecord>
// 		 entryExitRecord;	/* Array of EntryExit records--in
// 					  * Coverage Index order */
//    public:
//    DEFINE_SIZE_ARRAY (6, entryExitRecord);
//  };

//  struct CursivePos
//  {
//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 default:return (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    CursivePosFormat1	format1;
//    } u;
//  };

//  typedef AnchorMatrix BaseArray;		/* base-major--
// 					  * in order of BaseCoverage Index--,
// 					  * mark-minor--
// 					  * ordered by class--zero-based. */

//  static void Markclass_closure_and_remap_indexes (const Coverage  &mark_coverage,
// 						  const MarkArray &mark_array,
// 						  const hb_set_t  &glyphset,
// 						  hb_map_t*        klass_mapping /* INOUT */)
//  {
//    hb_set_t orig_classes;

//    + hb_zip (mark_coverage, mark_array)
//    | hb_filter (glyphset, hb_first)
//    | hb_map (hb_second)
//    | hb_map (&MarkRecord::get_class)
//    | hb_sink (orig_classes)
//    ;

//    unsigned idx = 0;
//    for (auto klass : orig_classes.iter ())
//    {
// 	 if (klass_mapping.has (klass)) continue;
// 	 klass_mapping.set (klass, idx);
// 	 idx++;
//    }
//  }

//  struct MarkBasePosFormat1
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 return (this+markCoverage).intersects (glyphs) &&
// 		(this+baseCoverage).intersects (glyphs);
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 + hb_zip (this+markCoverage, this+markArray)
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_apply ([&] (const MarkRecord& record) { record.collect_variation_indices (c, &(this+markArray)); })
// 	 ;

// 	 hb_map_t klass_mapping;
// 	 Markclass_closure_and_remap_indexes (this+markCoverage, this+markArray, *c.glyph_set, &klass_mapping);

// 	 unsigned basecount = (this+baseArray).rows;
// 	 auto base_iter =
// 	 + hb_zip (this+baseCoverage, hb_range (basecount))
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 ;

// 	 hb_sorted_vector_t<unsigned> base_indexes;
// 	 for (const unsigned row : base_iter)
// 	 {
// 	   + hb_range ((unsigned) classCount)
// 	   | hb_filter (klass_mapping)
// 	   | hb_map ([&] (const unsigned col) { return row * (unsigned) classCount + col; })
// 	   | hb_sink (base_indexes)
// 	   ;
// 	 }
// 	 (this+baseArray).collect_variation_indices (c, base_indexes.iter ());
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 if (unlikely (!(this+markCoverage).collect_coverage (c.input))) return;
// 	 if (unlikely (!(this+baseCoverage).collect_coverage (c.input))) return;
//    }

//    const Coverage &get_coverage () const { return this+markCoverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int mark_index = (this+markCoverage).get_coverage  (buffer.cur().codepoint);
// 	 if (likely (mark_index == NOT_COVERED)) return false;

// 	 /* Now we search backwards for a non-mark glyph */
// 	 hb_ot_apply_context_t::skipping_iterator_t &skippyIter = c.iter_input;
// 	 skippyIter.reset (buffer.idx, 1);
// 	 skippyIter.set_lookup_props (LookupFlag::IgnoreMarks);
// 	 do {
// 	   if (!skippyIter.prev ()) return false;
// 	   /* We only want to attach to the first of a MultipleSubst sequence.
// 		* https://github.com/harfbuzz/harfbuzz/issues/740
// 		* Reject others...
// 		* ...but stop if we find a mark in the MultipleSubst sequence:
// 		* https://github.com/harfbuzz/harfbuzz/issues/1020 */
// 	   if (!_hb_glyph_info_multiplied (&buffer.Info[skippyIter.idx]) ||
// 	   0 == _hb_glyph_info_get_lig_comp (&buffer.Info[skippyIter.idx]) ||
// 	   (skippyIter.idx == 0 ||
// 		_hb_glyph_info_is_mark (&buffer.Info[skippyIter.idx - 1]) ||
// 		_hb_glyph_info_get_lig_id (&buffer.Info[skippyIter.idx]) !=
// 		_hb_glyph_info_get_lig_id (&buffer.Info[skippyIter.idx - 1]) ||
// 		_hb_glyph_info_get_lig_comp (&buffer.Info[skippyIter.idx]) !=
// 		_hb_glyph_info_get_lig_comp (&buffer.Info[skippyIter.idx - 1]) + 1
// 		))
// 	 break;
// 	   skippyIter.reject ();
// 	 } while true;

// 	 /* Checking that matched glyph is actually a base glyph by GDEF is too strong; disabled */
// 	 //if (!_hb_glyph_info_is_base_glyph (&buffer.Info[skippyIter.idx])) { return false; }

// 	 unsigned int baseIndex = (this+baseCoverage).get_coverage  (buffer.Info[skippyIter.idx].codepoint);
// 	 if (baseIndex == NOT_COVERED) return false;

// 	 return ((this+markArray).apply (c, mark_index, baseIndex, this+baseArray, classCount, skippyIter.idx));
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return false;
// 	 out.format = format;

// 	 hb_map_t klass_mapping;
// 	 Markclass_closure_and_remap_indexes (this+markCoverage, this+markArray, glyphset, &klass_mapping);

// 	 if (!klass_mapping.get_population ()) return false;
// 	 out.classCount = klass_mapping.get_population ();

// 	 auto mark_iter =
// 	 + hb_zip (this+markCoverage, this+markArray)
// 	 | hb_filter (glyphset, hb_first)
// 	 ;

// 	 hb_sorted_vector_t<hb_codepoint_t> new_coverage;
// 	 + mark_iter
// 	 | hb_map (hb_first)
// 	 | hb_map (glyph_map)
// 	 | hb_sink (new_coverage)
// 	 ;

// 	 if (!out.markCoverage.serialize (c.serializer, out)
// 			   .serialize (c.serializer, new_coverage.iter ()))
// 	   return false;

// 	 out.markArray.serialize (c.serializer, out)
// 		   .serialize (c.serializer, &klass_mapping, c.plan.layout_variation_idx_map, &(this+markArray), + mark_iter
// 																			| hb_map (hb_second));

// 	 unsigned basecount = (this+baseArray).rows;
// 	 auto base_iter =
// 	 + hb_zip (this+baseCoverage, hb_range (basecount))
// 	 | hb_filter (glyphset, hb_first)
// 	 ;

// 	 new_coverage.reset ();
// 	 + base_iter
// 	 | hb_map (hb_first)
// 	 | hb_map (glyph_map)
// 	 | hb_sink (new_coverage)
// 	 ;

// 	 if (!out.baseCoverage.serialize (c.serializer, out)
// 			   .serialize (c.serializer, new_coverage.iter ()))
// 	   return false;

// 	 hb_sorted_vector_t<unsigned> base_indexes;
// 	 for (const unsigned row : + base_iter
// 				   | hb_map (hb_second))
// 	 {
// 	   + hb_range ((unsigned) classCount)
// 	   | hb_filter (klass_mapping)
// 	   | hb_map ([&] (const unsigned col) { return row * (unsigned) classCount + col; })
// 	   | hb_sink (base_indexes)
// 	   ;
// 	 }
// 	 out.baseArray.serialize (c.serializer, out)
// 		   .serialize (c.serializer, base_iter.len (), &(this+baseArray), c.plan.layout_variation_idx_map, base_indexes.iter ());

// 	 return true;
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this) &&
// 		   markCoverage.sanitize (c, this) &&
// 		   baseCoverage.sanitize (c, this) &&
// 		   markArray.sanitize (c, this) &&
// 		   baseArray.sanitize (c, this, (unsigned int) classCount));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    OffsetTo<Coverage>
// 		 markCoverage;		/* Offset to MarkCoverage table--from
// 					  * beginning of MarkBasePos subtable */
//    OffsetTo<Coverage>
// 		 baseCoverage;		/* Offset to BaseCoverage table--from
// 					  * beginning of MarkBasePos subtable */
//    HBUINT16	classCount;		/* Number of classes defined for marks */
//    OffsetTo<MarkArray>
// 		 markArray;		/* Offset to MarkArray table--from
// 					  * beginning of MarkBasePos subtable */
//    OffsetTo<BaseArray>
// 		 baseArray;		/* Offset to BaseArray table--from
// 					  * beginning of MarkBasePos subtable */
//    public:
//    DEFINE_SIZE_STATIC (12);
//  };

//  struct MarkBasePos
//  {
//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 default:return (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    MarkBasePosFormat1	format1;
//    } u;
//  };

//  typedef AnchorMatrix LigatureAttach;	/* component-major--
// 					  * in order of writing direction--,
// 					  * mark-minor--
// 					  * ordered by class--zero-based. */

//  /* Array of LigatureAttach tables ordered by LigatureCoverage Index */
//  struct LigatureArray : OffsetListOf<LigatureAttach>
//  {
//    template <typename Iterator,
// 		 hb_requires (hb_is_iterator (Iterator))>
//    bool subset (hb_subset_context_t *c,
// 			Iterator		    coverage,
// 			unsigned		    class_count,
// 			const hb_map_t	   *klass_mapping) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();

// 	 auto *out = c.serializer.start_embed (this);
// 	 if (unlikely (!c.serializer.extend_min (out)))  return false;

// 	 unsigned ligature_count = 0;
// 	 for (hb_codepoint_t gid : coverage)
// 	 {
// 	   ligature_count++;
// 	   if (!glyphset.has (gid)) continue;

// 	   auto *matrix = out.serialize_append (c.serializer);
// 	   if (unlikely (!matrix)) return false;

// 	   matrix.serialize_subset (c,
// 				 this.arrayZ[ligature_count - 1],
// 				 this,
// 				 class_count,
// 				 klass_mapping);
// 	 }
// 	 return (this.len);
//    }
//  };

//  struct MarkLigPosFormat1
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 return (this+markCoverage).intersects (glyphs) &&
// 		(this+ligatureCoverage).intersects (glyphs);
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 + hb_zip (this+markCoverage, this+markArray)
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_apply ([&] (const MarkRecord& record) { record.collect_variation_indices (c, &(this+markArray)); })
// 	 ;

// 	 hb_map_t klass_mapping;
// 	 Markclass_closure_and_remap_indexes (this+markCoverage, this+markArray, *c.glyph_set, &klass_mapping);

// 	 unsigned ligcount = (this+ligatureArray).len;
// 	 auto lig_iter =
// 	 + hb_zip (this+ligatureCoverage, hb_range (ligcount))
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 ;

// 	 const LigatureArray& lig_array = this+ligatureArray;
// 	 for (const unsigned i : lig_iter)
// 	 {
// 	   hb_sorted_vector_t<unsigned> lig_indexes;
// 	   unsigned row_count = lig_array[i].rows;
// 	   for (unsigned row : + hb_range (row_count))
// 	   {
// 	 + hb_range ((unsigned) classCount)
// 	 | hb_filter (klass_mapping)
// 	 | hb_map ([&] (const unsigned col) { return row * (unsigned) classCount + col; })
// 	 | hb_sink (lig_indexes)
// 	 ;
// 	   }

// 	   lig_array[i].collect_variation_indices (c, lig_indexes.iter ());
// 	 }
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 if (unlikely (!(this+markCoverage).collect_coverage (c.input))) return;
// 	 if (unlikely (!(this+ligatureCoverage).collect_coverage (c.input))) return;
//    }

//    const Coverage &get_coverage () const { return this+markCoverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int mark_index = (this+markCoverage).get_coverage  (buffer.cur().codepoint);
// 	 if (likely (mark_index == NOT_COVERED)) return false;

// 	 /* Now we search backwards for a non-mark glyph */
// 	 hb_ot_apply_context_t::skipping_iterator_t &skippyIter = c.iter_input;
// 	 skippyIter.reset (buffer.idx, 1);
// 	 skippyIter.set_lookup_props (LookupFlag::IgnoreMarks);
// 	 if (!skippyIter.prev ()) return false;

// 	 /* Checking that matched glyph is actually a ligature by GDEF is too strong; disabled */
// 	 //if (!_hb_glyph_info_is_ligature (&buffer.Info[skippyIter.idx])) { return false; }

// 	 unsigned int j = skippyIter.idx;
// 	 unsigned int ligIndex = (this+ligatureCoverage).get_coverage  (buffer.Info[j].codepoint);
// 	 if (ligIndex == NOT_COVERED) return false;

// 	 const LigatureArray& lig_array = this+ligatureArray;
// 	 const LigatureAttach& ligAttach = lig_array[ligIndex];

// 	 /* Find component to attach to */
// 	 unsigned int compCount = ligAttach.rows;
// 	 if (unlikely (!compCount)) return false;

// 	 /* We must now check whether the ligature ID of the current mark glyph
// 	  * is identical to the ligature ID of the found ligature.  If yes, we
// 	  * can directly use the component index.  If not, we attach the mark
// 	  * glyph to the last component of the ligature. */
// 	 unsigned int compIndex;
// 	 unsigned int ligId = _hb_glyph_info_get_lig_id (&buffer.Info[j]);
// 	 unsigned int markId = _hb_glyph_info_get_lig_id (&buffer.cur());
// 	 unsigned int markComp = _hb_glyph_info_get_lig_comp (&buffer.cur());
// 	 if (ligId && ligId == markId && markComp > 0)
// 	   compIndex = hb_min (compCount, _hb_glyph_info_get_lig_comp (&buffer.cur())) - 1;
// 	 else
// 	   compIndex = compCount - 1;

// 	 return ((this+markArray).apply (c, mark_index, compIndex, ligAttach, classCount, j));
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return false;
// 	 out.format = format;

// 	 hb_map_t klass_mapping;
// 	 Markclass_closure_and_remap_indexes (this+markCoverage, this+markArray, glyphset, &klass_mapping);

// 	 if (!klass_mapping.get_population ()) return false;
// 	 out.classCount = klass_mapping.get_population ();

// 	 auto mark_iter =
// 	 + hb_zip (this+markCoverage, this+markArray)
// 	 | hb_filter (glyphset, hb_first)
// 	 ;

// 	 auto new_mark_coverage =
// 	 + mark_iter
// 	 | hb_map_retains_sorting (hb_first)
// 	 | hb_map_retains_sorting (glyph_map)
// 	 ;

// 	 if (!out.markCoverage.serialize (c.serializer, out)
// 			   .serialize (c.serializer, new_mark_coverage))
// 	   return false;

// 	 out.markArray.serialize (c.serializer, out)
// 		   .serialize (c.serializer,
// 							   &klass_mapping,
// 							   c.plan.layout_variation_idx_map,
// 							   &(this+markArray),
// 							   + mark_iter
// 							   | hb_map (hb_second));

// 	 auto new_ligature_coverage =
// 	 + hb_iter (this + ligatureCoverage)
// 	 | hb_filter (glyphset)
// 	 | hb_map_retains_sorting (glyph_map)
// 	 ;

// 	 if (!out.ligatureCoverage.serialize (c.serializer, out)
// 				   .serialize (c.serializer, new_ligature_coverage))
// 	   return false;

// 	 out.ligatureArray.serialize_subset (c, ligatureArray, this,
// 										  hb_iter (this+ligatureCoverage), classCount, &klass_mapping);

// 	 return true;
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this) &&
// 		   markCoverage.sanitize (c, this) &&
// 		   ligatureCoverage.sanitize (c, this) &&
// 		   markArray.sanitize (c, this) &&
// 		   ligatureArray.sanitize (c, this, (unsigned int) classCount));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    OffsetTo<Coverage>
// 		 markCoverage;		/* Offset to Mark Coverage table--from
// 					  * beginning of MarkLigPos subtable */
//    OffsetTo<Coverage>
// 		 ligatureCoverage;	/* Offset to Ligature Coverage
// 					  * table--from beginning of MarkLigPos
// 					  * subtable */
//    HBUINT16	classCount;		/* Number of defined mark classes */
//    OffsetTo<MarkArray>
// 		 markArray;		/* Offset to MarkArray table--from
// 					  * beginning of MarkLigPos subtable */
//    OffsetTo<LigatureArray>
// 		 ligatureArray;		/* Offset to LigatureArray table--from
// 					  * beginning of MarkLigPos subtable */
//    public:
//    DEFINE_SIZE_STATIC (12);
//  };

//  struct MarkLigPos
//  {
//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 default:return (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    MarkLigPosFormat1	format1;
//    } u;
//  };

//  typedef AnchorMatrix Mark2Array;	/* mark2-major--
// 					  * in order of Mark2Coverage Index--,
// 					  * mark1-minor--
// 					  * ordered by class--zero-based. */

//  struct MarkMarkPosFormat1
//  {
//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 return (this+mark1Coverage).intersects (glyphs) &&
// 		(this+mark2Coverage).intersects (glyphs);
//    }

//    void closure_lookups (hb_closure_lookups_context_t *c) const {}

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 + hb_zip (this+mark1Coverage, this+mark1Array)
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 | hb_apply ([&] (const MarkRecord& record) { record.collect_variation_indices (c, &(this+mark1Array)); })
// 	 ;

// 	 hb_map_t klass_mapping;
// 	 Markclass_closure_and_remap_indexes (this+mark1Coverage, this+mark1Array, *c.glyph_set, &klass_mapping);

// 	 unsigned mark2_count = (this+mark2Array).rows;
// 	 auto mark2_iter =
// 	 + hb_zip (this+mark2Coverage, hb_range (mark2_count))
// 	 | hb_filter (c.glyph_set, hb_first)
// 	 | hb_map (hb_second)
// 	 ;

// 	 hb_sorted_vector_t<unsigned> mark2_indexes;
// 	 for (const unsigned row : mark2_iter)
// 	 {
// 	   + hb_range ((unsigned) classCount)
// 	   | hb_filter (klass_mapping)
// 	   | hb_map ([&] (const unsigned col) { return row * (unsigned) classCount + col; })
// 	   | hb_sink (mark2_indexes)
// 	   ;
// 	 }
// 	 (this+mark2Array).collect_variation_indices (c, mark2_indexes.iter ());
//    }

//    void collect_glyphs (hb_collect_glyphs_context_t *c) const
//    {
// 	 if (unlikely (!(this+mark1Coverage).collect_coverage (c.input))) return;
// 	 if (unlikely (!(this+mark2Coverage).collect_coverage (c.input))) return;
//    }

//    const Coverage &get_coverage () const { return this+mark1Coverage; }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 buffer *Buffer = c.buffer;
// 	 unsigned int mark1_index = (this+mark1Coverage).get_coverage  (buffer.cur().codepoint);
// 	 if (likely (mark1_index == NOT_COVERED)) return false;

// 	 /* now we search backwards for a suitable mark glyph until a non-mark glyph */
// 	 hb_ot_apply_context_t::skipping_iterator_t &skippyIter = c.iter_input;
// 	 skippyIter.reset (buffer.idx, 1);
// 	 skippyIter.set_lookup_props (c.lookup_props & ~LookupFlag::IgnoreFlags);
// 	 if (!skippyIter.prev ()) return false;

// 	 if (!_hb_glyph_info_is_mark (&buffer.Info[skippyIter.idx])) { return false; }

// 	 unsigned int j = skippyIter.idx;

// 	 unsigned int id1 = _hb_glyph_info_get_lig_id (&buffer.cur());
// 	 unsigned int id2 = _hb_glyph_info_get_lig_id (&buffer.Info[j]);
// 	 unsigned int comp1 = _hb_glyph_info_get_lig_comp (&buffer.cur());
// 	 unsigned int comp2 = _hb_glyph_info_get_lig_comp (&buffer.Info[j]);

// 	 if (likely (id1 == id2))
// 	 {
// 	   if (id1 == 0) /* Marks belonging to the same base. */
// 	 goto good;
// 	   else if (comp1 == comp2) /* Marks belonging to the same ligature component. */
// 	 goto good;
// 	 }
// 	 else
// 	 {
// 	   /* If ligature ids don't match, it may be the case that one of the marks
// 		* itself is a ligature.  In which case match. */
// 	   if ((id1 > 0 && !comp1) || (id2 > 0 && !comp2))
// 	 goto good;
// 	 }

// 	 /* Didn't match. */
// 	 return false;

// 	 good:
// 	 unsigned int mark2Index = (this+mark2Coverage).get_coverage  (buffer.Info[j].codepoint);
// 	 if (mark2Index == NOT_COVERED) return false;

// 	 return ((this+mark1Array).apply (c, mark1_index, mark2Index, this+mark2Array, classCount, j));
//    }

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 TRACE_SUBSET (this);
// 	 const hb_set_t &glyphset = *c.plan.glyphset_gsub ();
// 	 const hb_map_t &glyph_map = *c.plan.glyph_map;

// 	 auto *out = c.serializer.start_embed (*this);
// 	 if (unlikely (!c.serializer.extend_min (out))) return false;
// 	 out.format = format;

// 	 hb_map_t klass_mapping;
// 	 Markclass_closure_and_remap_indexes (this+mark1Coverage, this+mark1Array, glyphset, &klass_mapping);

// 	 if (!klass_mapping.get_population ()) return false;
// 	 out.classCount = klass_mapping.get_population ();

// 	 auto mark1_iter =
// 	 + hb_zip (this+mark1Coverage, this+mark1Array)
// 	 | hb_filter (glyphset, hb_first)
// 	 ;

// 	 hb_sorted_vector_t<hb_codepoint_t> new_coverage;
// 	 + mark1_iter
// 	 | hb_map (hb_first)
// 	 | hb_map (glyph_map)
// 	 | hb_sink (new_coverage)
// 	 ;

// 	 if (!out.mark1Coverage.serialize (c.serializer, out)
// 				.serialize (c.serializer, new_coverage.iter ()))
// 	   return false;

// 	 out.mark1Array.serialize (c.serializer, out)
// 			.serialize (c.serializer, &klass_mapping, c.plan.layout_variation_idx_map, &(this+mark1Array), + mark1_iter
// 																			  | hb_map (hb_second));

// 	 unsigned mark2count = (this+mark2Array).rows;
// 	 auto mark2_iter =
// 	 + hb_zip (this+mark2Coverage, hb_range (mark2count))
// 	 | hb_filter (glyphset, hb_first)
// 	 ;

// 	 new_coverage.reset ();
// 	 + mark2_iter
// 	 | hb_map (hb_first)
// 	 | hb_map (glyph_map)
// 	 | hb_sink (new_coverage)
// 	 ;

// 	 if (!out.mark2Coverage.serialize (c.serializer, out)
// 				.serialize (c.serializer, new_coverage.iter ()))
// 	   return false;

// 	 hb_sorted_vector_t<unsigned> mark2_indexes;
// 	 for (const unsigned row : + mark2_iter
// 				   | hb_map (hb_second))
// 	 {
// 	   + hb_range ((unsigned) classCount)
// 	   | hb_filter (klass_mapping)
// 	   | hb_map ([&] (const unsigned col) { return row * (unsigned) classCount + col; })
// 	   | hb_sink (mark2_indexes)
// 	   ;
// 	 }
// 	 out.mark2Array.serialize (c.serializer, out)
// 			.serialize (c.serializer, mark2_iter.len (), &(this+mark2Array), c.plan.layout_variation_idx_map, mark2_indexes.iter ());

// 	 return true;
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    {
// 	 TRACE_SANITIZE (this);
// 	 return (c.check_struct (this) &&
// 		   mark1Coverage.sanitize (c, this) &&
// 		   mark2Coverage.sanitize (c, this) &&
// 		   mark1Array.sanitize (c, this) &&
// 		   mark2Array.sanitize (c, this, (unsigned int) classCount));
//    }

//    protected:
//    HBUINT16	format;			/* Format identifier--format = 1 */
//    OffsetTo<Coverage>
// 		 mark1Coverage;		/* Offset to Combining Mark1 Coverage
// 					  * table--from beginning of MarkMarkPos
// 					  * subtable */
//    OffsetTo<Coverage>
// 		 mark2Coverage;		/* Offset to Combining Mark2 Coverage
// 					  * table--from beginning of MarkMarkPos
// 					  * subtable */
//    HBUINT16	classCount;		/* Number of defined mark classes */
//    OffsetTo<MarkArray>
// 		 mark1Array;		/* Offset to Mark1Array table--from
// 					  * beginning of MarkMarkPos subtable */
//    OffsetTo<Mark2Array>
// 		 mark2Array;		/* Offset to Mark2Array table--from
// 					  * beginning of MarkMarkPos subtable */
//    public:
//    DEFINE_SIZE_STATIC (12);
//  };

//  struct MarkMarkPos
//  {
//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, u.format);
// 	 if (unlikely (!c.may_dispatch (this, &u.format))) return (c.no_dispatch_return_value ());
// 	 switch (u.format) {
// 	 case 1: return (c.dispatch (u.format1, hb_forward<Ts> (ds)...));
// 	 default:return (c.default_return_value ());
// 	 }
//    }

//    protected:
//    union {
//    HBUINT16		format;		/* Format identifier */
//    MarkMarkPosFormat1	format1;
//    } u;
//  };

//  struct ContextPos : Context {};

//  struct ChainContextPos : ChainContext {};

//  struct ExtensionPos : Extension<ExtensionPos>
//  {
//    typedef struct PosLookupSubTable SubTable;
//  };

//  /*
//   * PosLookup
//   */

//  struct PosLookupSubTable
//  {
//    friend struct Lookup;
//    friend struct PosLookup;

//    enum Type {
// 	 Single		= 1,
// 	 Pair		= 2,
// 	 Cursive		= 3,
// 	 MarkBase		= 4,
// 	 MarkLig		= 5,
// 	 MarkMark		= 6,
// 	 Context		= 7,
// 	 ChainContext	= 8,
// 	 Extension		= 9
//    };

//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, unsigned int lookup_type, Ts&&... ds) const
//    {
// 	 TRACE_DISPATCH (this, lookup_type);
// 	 switch (lookup_type) {
// 	 case Single:		return (u.single.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case Pair:			return (u.pair.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case Cursive:		return (u.cursive.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case MarkBase:		return (u.markBase.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case MarkLig:		return (u.markLig.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case MarkMark:		return (u.markMark.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case Context:		return (u.context.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case ChainContext:		return (u.chainContext.dispatch (c, hb_forward<Ts> (ds)...));
// 	 case Extension:		return (u.extension.dispatch (c, hb_forward<Ts> (ds)...));
// 	 default:			return (c.default_return_value ());
// 	 }
//    }

//    bool intersects (const hb_set_t *glyphs, unsigned int lookup_type) const
//    {
// 	 hb_intersects_context_t c (glyphs);
// 	 return dispatch (&c, lookup_type);
//    }

//    protected:
//    union {
//    SinglePos		single;
//    PairPos		pair;
//    CursivePos		cursive;
//    MarkBasePos		markBase;
//    MarkLigPos		markLig;
//    MarkMarkPos		markMark;
//    ContextPos		context;
//    ChainContextPos	chainContext;
//    ExtensionPos		extension;
//    } u;
//    public:
//    DEFINE_SIZE_MIN (0);
//  };

//  struct PosLookup : Lookup
//  {
//    typedef struct PosLookupSubTable SubTable;

//    const SubTable& get_subtable (unsigned int i) const
//    { return Lookup::get_subtable<SubTable> (i); }

//    bool is_reverse () const
//    {
// 	 return false;
//    }

//    bool apply (hb_ot_apply_context_t *c) const
//    {
// 	 TRACE_APPLY (this);
// 	 return (dispatch (c));
//    }

//    bool intersects (const hb_set_t *glyphs) const
//    {
// 	 hb_intersects_context_t c (glyphs);
// 	 return dispatch (&c);
//    }

//    hb_collect_glyphs_context_t::return_t collect_glyphs (hb_collect_glyphs_context_t *c) const
//    { return dispatch (c); }

//    hb_closure_lookups_context_t::return_t closure_lookups (hb_closure_lookups_context_t *c, unsigned this_index) const
//    {
// 	 if (c.is_lookup_visited (this_index))
// 	   return hb_closure_lookups_context_t::default_return_value ();

// 	 c.set_lookup_visited (this_index);
// 	 if (!intersects (c.glyphs))
// 	 {
// 	   c.set_lookup_inactive (this_index);
// 	   return hb_closure_lookups_context_t::default_return_value ();
// 	 }
// 	 c.set_recurse_func (dispatch_closure_lookups_recurse_func);

// 	 hb_closure_lookups_context_t::return_t ret = dispatch (c);
// 	 return ret;
//    }

//    template <typename set_t>
//    void collect_coverage (set_t *glyphs) const
//    {
// 	 hb_collect_coverage_context_t<set_t> c (glyphs);
// 	 dispatch (&c);
//    }

//    static inline bool apply_recurse_func (hb_ot_apply_context_t *c, unsigned int lookup_index);

//    template <typename context_t>
//    static typename context_t::return_t dispatch_recurse_func (context_t *c, unsigned int lookup_index);

//    HB_INTERNAL static hb_closure_lookups_context_t::return_t dispatch_closure_lookups_recurse_func (hb_closure_lookups_context_t *c, unsigned this_index);

//    template <typename context_t, typename ...Ts>
//    typename context_t::return_t dispatch (context_t *c, Ts&&... ds) const
//    { return Lookup::dispatch<SubTable> (c, hb_forward<Ts> (ds)...); }

//    bool subset (hb_subset_context_t *c) const
//    { return Lookup::subset<SubTable> (c); }

//    bool sanitize (hb_sanitize_context_t *c) const
//    { return Lookup::sanitize<SubTable> (c); }
//  };

//  /*
//   * GPOS -- Glyph Positioning
//   * https://docs.microsoft.com/en-us/typography/opentype/spec/gpos
//   */

//  struct GPOS : GSUBGPOS
//  {
//    static constexpr hb_tag_t tableTag = HB_OT_TAG_GPOS;

//    const PosLookup& get_lookup (unsigned int i) const
//    { return static_cast<const PosLookup &> (GSUBGPOS::get_lookup (i)); }

//    static inline void position_finish_advances (hb_font_t *font, buffer *Buffer);
//    static inline void position_finish_offsets (hb_font_t *font, buffer *Buffer);

//    bool subset (hb_subset_context_t *c) const
//    {
// 	 hb_subset_layout_context_t l (c, tableTag, c.plan.gpos_lookups, c.plan.gpos_features);
// 	 return GSUBGPOS::subset<PosLookup> (&l);
//    }

//    bool sanitize (hb_sanitize_context_t *c) const
//    { return GSUBGPOS::sanitize<PosLookup> (c); }

//    HB_INTERNAL bool is_blocklisted (hb_blob_t *blob,
// 					hb_face_t *face) const;

//    void collect_variation_indices (hb_collect_variation_indices_context_t *c) const
//    {
// 	 for (unsigned i = 0; i < GSUBGPOS::get_lookup_count (); i++)
// 	 {
// 	   if (!c.gpos_lookups.has (i)) continue;
// 	   const PosLookup &l = get_lookup (i);
// 	   l.dispatch (c);
// 	 }
//    }

//    void closure_lookups (hb_face_t      *face,
// 			 const hb_set_t *glyphs,
// 			 hb_set_t       *lookup_indexes /* IN/OUT */) const
//    { GSUBGPOS::closure_lookups<PosLookup> (face, glyphs, lookup_indexes); }

//    typedef GSUBGPOS::accelerator_t<GPOS> accelerator_t;
//  };

func positionStartGPOS(buffer *Buffer) {
	for i := range buffer.Pos {
		buffer.Pos[i].attach_chain = 0
		buffer.Pos[i].attach_type = 0
	}
}

func propagateAttachmentOffsets(pos []GlyphPosition, i int, direction Direction) {
	/* Adjusts offsets of attached glyphs (both cursive and mark) to accumulate
	 * offset of glyph they are attached to. */
	chain, type_ := pos[i].attach_chain, pos[i].attach_type
	if chain == 0 {
		return
	}

	pos[i].attach_chain = 0

	j := i + int(chain)

	if j >= len(pos) {
		return
	}

	propagateAttachmentOffsets(pos, j, direction)

	//   assert (!!(type_ & attachTypeMark) ^ !!(type_ & attachTypeCursive));

	if (type_ & attachTypeCursive) != 0 {
		if direction.IsHorizontal() {
			pos[i].YOffset += pos[j].YOffset
		} else {
			pos[i].XOffset += pos[j].XOffset
		}
	} else /*if (type_ & attachTypeMark)*/ {
		pos[i].XOffset += pos[j].XOffset
		pos[i].YOffset += pos[j].YOffset

		// assert (j < i);
		if direction.IsBackward() {
			for _, p := range pos[j:i] {
				pos[i].XOffset -= p.XAdvance
				pos[i].YOffset -= p.YAdvance
			}
		} else {
			for _, p := range pos[j+1 : i+1] {
				pos[i].XOffset += p.XAdvance
				pos[i].YOffset += p.YAdvance
			}
		}
	}
}

func positionFinishOffsetsGPOS(buffer *Buffer) {
	pos := buffer.Pos
	direction := buffer.Props.Direction

	/* Handle attachments */
	if buffer.scratchFlags&HB_BUFFER_SCRATCH_FLAG_HAS_GPOS_ATTACHMENT != 0 {
		for i := range pos {
			propagateAttachmentOffsets(pos, i, direction)
		}
	}
}

//  struct GPOS_accelerator_t : GPOS::accelerator_t {};

//  /* Out-of-class implementation for methods recursing */

//  #ifndef HB_NO_OT_LAYOUT
//  template <typename context_t>
//  /*static*/ typename context_t::return_t PosLookup::dispatch_recurse_func (context_t *c, unsigned int lookup_index)
//  {
//    const PosLookup &l = c.face.table.GPOS.get_relaxed ().table.get_lookup (lookup_index);
//    return l.dispatch (c);
//  }

//  /*static*/ inline hb_closure_lookups_context_t::return_t PosLookup::dispatch_closure_lookups_recurse_func (hb_closure_lookups_context_t *c, unsigned this_index)
//  {
//    const PosLookup &l = c.face.table.GPOS.get_relaxed ().table.get_lookup (this_index);
//    return l.closure_lookups (c, this_index);
//  }

var _ TLookup = lookupGPOS{}

// implements TLookup
type lookupGPOS tt.LookupGPOS

func (l lookupGPOS) collectCoverage(dst *SetDigest) {
	for _, table := range l.Subtables {
		dst.collectCoverage(table.Coverage)
	}
}

func (l lookupGPOS) dispatchSubtables(ctx *hb_get_subtables_context_t) {
	for _, table := range l.Subtables {
		*ctx = append(*ctx, newGPOSApplicable(table))
	}
}

func (l lookupGPOS) dispatchApply(ctx *hb_ot_apply_context_t) bool {
	for _, table := range l.Subtables {
		if gposSubtable(table).apply(ctx) {
			return true
		}
	}
	return false
}

func (lookupGPOS) isReverse() bool { return false }

func apply_recurse_GPOS(c *hb_ot_apply_context_t, lookupIndex uint16) bool {
	_, gpos := c.face.get_gsubgpos_table()
	l := lookupGPOS(gpos.Lookups[lookupIndex])

	// TODO: unify with GSUB
	savedLookupProps := c.lookupProps
	savedLookupIndex := c.lookupIndex

	c.lookupIndex = lookupIndex
	c.set_lookup_props(l.Props())

	ret := l.dispatchApply(c)

	c.lookupIndex = savedLookupIndex
	c.set_lookup_props(savedLookupProps)
	return ret
}

//  implements `hb_apply_func_t`
type gposSubtable tt.GPOSSubtable

// return `true` is the positionning found a match and was applied
func (table gposSubtable) apply(c *hb_ot_apply_context_t) bool {
	buffer := c.buffer
	glyphId := buffer.cur(0).Glyph
	glyphPos := buffer.curPos(0)
	index, ok := table.Coverage.Index(glyphId)
	if !ok {
		return false
	}

	switch data := table.Data.(type) {
	case tt.GPOSSingle1:
		c.applyGPOSValueRecord(data.Format, data.Value, glyphPos)
		buffer.idx++
	case tt.GPOSSingle2:
		c.applyGPOSValueRecord(data.Format, data.Values[index], glyphPos)
		buffer.idx++
	case tt.GPOSPair1:
		skippyIter := &c.iter_input
		skippyIter.reset(buffer.idx, 1)
		if !skippyIter.next() {
			return false
		}
		set := data.Values[index]
		record := set.FindGlyph(buffer.Info[skippyIter.idx].Glyph)
		if record == nil {
			return false
		}
		c.applyGPOSPair(data.Formats, record.Pos, skippyIter.idx)
	case tt.GPOSPair2:
		skippyIter := &c.iter_input
		skippyIter.reset(buffer.idx, 1)
		if !skippyIter.next() {
			return false
		}
		class1, _ := data.First.ClassID(glyphId)
		class2, _ := data.Second.ClassID(buffer.Info[skippyIter.idx].Glyph)
		vals := data.Values[class1][class2]
		c.applyGPOSPair(data.Formats, vals, skippyIter.idx)
	case tt.GPOSCursive1:
		return c.applyGPOSCursive(data, index, table.Coverage)
	case tt.GPOSMarkToBase1:
		return c.applyGPOSMarkToBase(data, index)
	case tt.GPOSMarkToLigature1:
		return c.applyGPOSMarkToLigature(data, index)
	case tt.GPOSMarkToMark1:
		return c.applyGPOSMarkToMark(data, index)
	case tt.GPOSContext1:
		return c.applyLookupContext1(tt.LookupContext1(data), index)
	case tt.GPOSContext2:
		return c.applyLookupContext2(tt.LookupContext2(data), index, glyphId)
	case tt.GPOSContext3:
		return c.applyLookupContext3(tt.LookupContext3(data), index)
	case tt.GPOSChainedContext1:
		return c.applyLookupChainedContext1(tt.LookupChainedContext1(data), index)
	case tt.GPOSChainedContext2:
		return c.applyLookupChainedContext2(tt.LookupChainedContext2(data), index, glyphId)
	case tt.GPOSChainedContext3:
		return c.applyLookupChainedContext3(tt.LookupChainedContext3(data), index)
	}
	return true
}

func (c *hb_ot_apply_context_t) applyGPOSValueRecord(format tt.GPOSValueFormat, v tt.GPOSValueRecord, glyphPos *GlyphPosition) bool {
	var ret bool
	if format == 0 {
		return ret
	}

	font := c.font
	horizontal := c.direction.IsHorizontal()

	if format&tt.XPlacement != 0 {
		glyphPos.XOffset += font.em_scale_x(v.XPlacement)
		ret = ret || v.XPlacement != 0
	}
	if format&tt.YPlacement != 0 {
		glyphPos.YOffset += font.em_scale_y(v.YPlacement)
		ret = ret || v.YPlacement != 0
	}
	if format&tt.XAdvance != 0 {
		if horizontal {
			glyphPos.XAdvance += font.em_scale_x(v.XAdvance)
			ret = ret || v.XAdvance != 0
		}
	}
	/* YAdvance values grow downward but font-space grows upward, hence negation */
	if format&tt.YAdvance != 0 {
		if !horizontal {
			glyphPos.YAdvance -= font.em_scale_y(v.YAdvance)
			ret = ret || v.YAdvance != 0
		}
	}

	if format&tt.Devices == 0 {
		return ret
	}

	useXDevice := font.x_ppem != 0 || len(font.coords) != 0
	useYDevice := font.y_ppem != 0 || len(font.coords) != 0

	if !useXDevice && !useYDevice {
		return ret
	}

	if format&tt.XPlaDevice != 0 && useXDevice {
		glyphPos.XOffset += c.get_x_delta(font, v.XPlaDevice)
		ret = ret || v.XPlaDevice != nil
	}
	if format&tt.YPlaDevice != 0 && useYDevice {
		glyphPos.YOffset += c.get_y_delta(font, v.YPlaDevice)
		ret = ret || v.YPlaDevice != nil
	}
	if format&tt.XAdvDevice != 0 && horizontal && useXDevice {
		glyphPos.XAdvance += c.get_x_delta(font, v.XAdvDevice)
		ret = ret || v.XAdvDevice != nil
	}
	if format&tt.YAdvDevice != 0 && !horizontal && useYDevice {
		/* YAdvance values grow downward but font-space grows upward, hence negation */
		glyphPos.YAdvance -= c.get_y_delta(font, v.YAdvDevice)
		ret = ret || v.YAdvDevice != nil
	}
	return ret
}

func (c *hb_ot_apply_context_t) get_x_delta(font *Font, device tt.GPOSDevice) Position {
	switch device := device.(type) {
	case tt.GPOSDeviceHinting:
		return device.GetDelta(font.x_ppem, font.XScale)
	case tt.GPOSDeviceVariation:
		return font.em_scalef_x(device.GetDelta(font.coords, c.varStore))
	default:
		return 0
	}
}

func (c *hb_ot_apply_context_t) get_y_delta(font *Font, device tt.GPOSDevice) Position {
	switch device := device.(type) {
	case tt.GPOSDeviceHinting:
		return device.GetDelta(font.y_ppem, font.YScale)
	case tt.GPOSDeviceVariation:
		return font.em_scalef_y(device.GetDelta(font.coords, c.varStore))
	default:
		return 0
	}
}

func reverseCursiveMinorOffset(pos []GlyphPosition, i int, direction Direction, new_parent int) {
	chain, type_ := pos[i].attach_chain, pos[i].attach_type
	if chain == 0 || 0 == (type_&attachTypeCursive) {
		return
	}

	pos[i].attach_chain = 0

	j := i + int(chain)

	// stop if we see new parent in the chain
	if j == new_parent {
		return
	}
	reverseCursiveMinorOffset(pos, j, direction, new_parent)

	if direction.IsHorizontal() {
		pos[j].YOffset = -pos[i].YOffset
	} else {
		pos[j].XOffset = -pos[i].XOffset
	}

	pos[j].attach_chain = -chain
	pos[j].attach_type = type_
}

func (c *hb_ot_apply_context_t) applyGPOSPair(formats [2]tt.GPOSValueFormat, values [2]tt.GPOSValueRecord, pos int) {
	buffer := c.buffer

	ap1 := c.applyGPOSValueRecord(formats[0], values[0], buffer.curPos(0))
	ap2 := c.applyGPOSValueRecord(formats[1], values[1], &buffer.Pos[pos])

	if ap1 || ap2 {
		buffer.unsafeToBreak(buffer.idx, pos+1)
	}
	buffer.idx = pos
	if formats[1] != 0 {
		buffer.idx++
	}
}

func (c *hb_ot_apply_context_t) applyGPOSCursive(data tt.GPOSCursive1, covIndex int, cov tt.Coverage) bool {
	buffer := c.buffer

	thisRecord := data[covIndex]
	if thisRecord[0] == nil {
		return false
	}

	skippyIter := &c.iter_input
	skippyIter.reset(buffer.idx, 1)
	if !skippyIter.prev() {
		return false
	}

	prevIndex, ok := cov.Index(buffer.Info[skippyIter.idx].Glyph)
	if !ok {
		return false
	}
	prevRecord := data[prevIndex]
	if prevRecord[1] == nil {
		return false
	}

	i := skippyIter.idx
	j := buffer.idx

	buffer.unsafeToBreak(i, j)
	exitX, exitY := c.getAnchor(prevRecord[1], buffer.Info[i].Glyph)
	entryX, entryY := c.getAnchor(thisRecord[0], buffer.Info[j].Glyph)

	pos := buffer.Pos

	var d Position
	/* Main-direction adjustment */
	switch c.direction {
	case LeftToRight:
		pos[i].XAdvance = roundf(exitX) + pos[i].XOffset

		d = roundf(entryX) + pos[j].XOffset
		pos[j].XAdvance -= d
		pos[j].XOffset -= d
	case RightToLeft:
		d = roundf(exitX) + pos[i].XOffset
		pos[i].XAdvance -= d
		pos[i].XOffset -= d

		pos[j].XAdvance = roundf(entryX) + pos[j].XOffset
	case TopToBottom:
		pos[i].YAdvance = roundf(exitY) + pos[i].YOffset

		d = roundf(entryY) + pos[j].YOffset
		pos[j].YAdvance -= d
		pos[j].YOffset -= d
	case BottomToTop:
		d = roundf(exitY) + pos[i].YOffset
		pos[i].YAdvance -= d
		pos[i].YOffset -= d

		pos[j].YAdvance = roundf(entryY)
	}

	/* Cross-direction adjustment */

	/* We attach child to parent (think graph theory and rooted trees whereas
	 * the root stays on baseline and each node aligns itself against its
	 * parent.
	 *
	 * Optimize things for the case of RightToLeft, as that's most common in
	 * Arabic. */
	child := i
	parent := j
	xOffset := Position(entryX - exitX)
	yOffset := Position(entryY - exitY)
	if uint16(c.lookupProps)&tt.RightToLeft == 0 {
		k := child
		child = parent
		parent = k
		xOffset = -xOffset
		yOffset = -yOffset
	}

	/* If child was already connected to someone else, walk through its old
	 * chain and reverse the link direction, such that the whole tree of its
	 * previous connection now attaches to new parent.  Watch out for case
	 * where new parent is on the path from old chain...
	 */
	reverseCursiveMinorOffset(pos, child, c.direction, parent)

	pos[child].attach_type = attachTypeCursive
	pos[child].attach_chain = int16(parent - child)
	buffer.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_GPOS_ATTACHMENT
	if c.direction.IsHorizontal() {
		pos[child].YOffset = yOffset
	} else {
		pos[child].XOffset = xOffset
	}

	/* If parent was attached to child, break them free.
	 * https://github.com/harfbuzz/harfbuzz/issues/2469 */
	if pos[parent].attach_chain == -pos[child].attach_chain {
		pos[parent].attach_chain = 0
	}

	buffer.idx++
	return true
}

// panic if anchor is nil
func (c *hb_ot_apply_context_t) getAnchor(anchor tt.GPOSAnchor, glyph fonts.GlyphIndex) (x, y float32) {
	font := c.font
	switch anchor := anchor.(type) {
	case tt.GPOSAnchorFormat1:
		return font.em_fscale_x(anchor.X), font.em_fscale_y(anchor.Y)
	case tt.GPOSAnchorFormat2:
		x_ppem, y_ppem := font.x_ppem, font.y_ppem
		var cx, cy Position
		ret := x_ppem != 0 || y_ppem != 0
		if ret {
			cx, cy, ret = font.get_glyph_contour_point_for_origin(glyph, anchor.AnchorPoint, LeftToRight)
		}
		if ret && x_ppem != 0 {
			x = float32(cx)
		} else {
			x = font.em_fscale_x(anchor.X)
		}
		if ret && y_ppem != 0 {
			y = float32(cy)
		} else {
			y = font.em_fscale_y(anchor.Y)
		}
		return x, y
	case tt.GPOSAnchorFormat3:
		x, y = font.em_fscale_x(anchor.X), font.em_fscale_y(anchor.Y)
		if font.x_ppem != 0 || len(font.coords) != 0 {
			x += float32(c.get_x_delta(font, anchor.XDevice))
		}
		if font.y_ppem != 0 || len(font.coords) != 0 {
			y += float32(c.get_y_delta(font, anchor.YDevice))
		}
		return x, y
	default:
		panic("exhaustive switch")
	}
}

func (c *hb_ot_apply_context_t) applyGPOSMarks(marks []tt.GPOSMark, markIndex, glyphIndex int, anchors [][]tt.GPOSAnchor, glyphPos int) bool {
	buffer := c.buffer
	record := &marks[markIndex]
	markClass := record.ClassValue
	markAnchor := record.Anchor

	glyphAnchor := anchors[glyphIndex][markClass]
	/* If this subtable doesn't have an anchor for this base and this class,
	 * return false such that the subsequent subtables have a chance at it. */
	if glyphAnchor == nil {
		return false
	}

	buffer.unsafeToBreak(glyphPos, buffer.idx)
	markX, markY := c.getAnchor(markAnchor, buffer.cur(0).Glyph)
	baseX, baseY := c.getAnchor(glyphAnchor, buffer.Info[glyphPos].Glyph)

	o := buffer.curPos(0)
	o.XOffset = roundf(baseX - markX)
	o.YOffset = roundf(baseY - markY)
	o.attach_type = attachTypeMark
	o.attach_chain = int16(glyphPos - buffer.idx)
	buffer.scratchFlags |= HB_BUFFER_SCRATCH_FLAG_HAS_GPOS_ATTACHMENT

	buffer.idx++
	return true
}

func (c *hb_ot_apply_context_t) applyGPOSMarkToBase(data tt.GPOSMarkToBase1, markIndex int) bool {
	buffer := c.buffer

	// now we search backwards for a non-mark glyph
	skippyIter := &c.iter_input
	skippyIter.reset(buffer.idx, 1)
	skippyIter.matcher.lookupProps = uint32(tt.IgnoreMarks)
	for {
		if !skippyIter.prev() {
			return false
		}
		/* We only want to attach to the first of a MultipleSubst sequence.
		 * https://github.com/harfbuzz/harfbuzz/issues/740
		 * Reject others...
		 * ...but stop if we find a mark in the MultipleSubst sequence:
		 * https://github.com/harfbuzz/harfbuzz/issues/1020 */
		if !buffer.Info[skippyIter.idx].multiplied() || 0 == buffer.Info[skippyIter.idx].getLigComp() ||
			skippyIter.idx == 0 || buffer.Info[skippyIter.idx-1].isMark() ||
			buffer.Info[skippyIter.idx].getLigId() != buffer.Info[skippyIter.idx-1].getLigId() ||
			buffer.Info[skippyIter.idx].getLigComp() != buffer.Info[skippyIter.idx-1].getLigComp()+1 {
			break
		}
		skippyIter.reject()
	}

	/* Checking that matched glyph is actually a base glyph by GDEF is too strong; disabled */
	//if (!_hb_glyph_info_is_base_glyph (&buffer.Info[skippyIter.idx])) { return false; }

	baseIndex, ok := data.BaseCoverage.Index(buffer.Info[skippyIter.idx].Glyph)
	if !ok {
		return false
	}

	return c.applyGPOSMarks(data.Marks, markIndex, baseIndex, data.Bases, skippyIter.idx)
}

func (c *hb_ot_apply_context_t) applyGPOSMarkToLigature(data tt.GPOSMarkToLigature1, markIndex int) bool {
	buffer := c.buffer

	// now we search backwards for a non-mark glyph
	skippyIter := &c.iter_input
	skippyIter.reset(buffer.idx, 1)
	skippyIter.matcher.lookupProps = uint32(tt.IgnoreMarks)
	if !skippyIter.prev() {
		return false
	}

	j := skippyIter.idx
	ligIndex, ok := data.LigatureCoverage.Index(buffer.Info[j].Glyph)
	if !ok {
		return false
	}

	ligAttach := data.Ligatures[ligIndex]

	/* Find component to attach to */
	compCount := len(ligAttach)
	if compCount == 0 {
		return false
	}

	/* We must now check whether the ligature ID of the current mark glyph
	 * is identical to the ligature ID of the found ligature.  If yes, we
	 * can directly use the component index.  If not, we attach the mark
	 * glyph to the last component of the ligature. */
	ligId := buffer.Info[j].getLigId()
	markId := buffer.cur(0).getLigId()
	markComp := buffer.cur(0).getLigComp()
	compIndex := compCount - 1
	if ligId != 0 && ligId == markId && markComp > 0 {
		compIndex = min(compCount, int(buffer.cur(0).getLigComp())) - 1
	}

	return c.applyGPOSMarks(data.Marks, markIndex, compIndex, ligAttach, skippyIter.idx)
}

func (c *hb_ot_apply_context_t) applyGPOSMarkToMark(data tt.GPOSMarkToMark1, mark1Index int) bool {
	buffer := c.buffer

	// now we search backwards for a suitable mark glyph until a non-mark glyph
	skippyIter := &c.iter_input
	skippyIter.reset(buffer.idx, 1)
	skippyIter.matcher.lookupProps = c.lookupProps &^ uint32(ignoreFlags)
	if !skippyIter.prev() {
		return false
	}

	if !buffer.Info[skippyIter.idx].isMark() {
		return false
	}

	j := skippyIter.idx

	id1 := buffer.cur(0).getLigId()
	id2 := buffer.Info[j].getLigId()
	comp1 := buffer.cur(0).getLigComp()
	comp2 := buffer.Info[j].getLigComp()

	if id1 == id2 {
		if id1 == 0 { /* Marks belonging to the same base. */
			goto good
		} else if comp1 == comp2 { /* Marks belonging to the same ligature component. */
			goto good
		}
	} else {
		/* If ligature ids don't match, it may be the case that one of the marks
		* itself is a ligature.  In which case match. */
		if (id1 > 0 && comp1 == 0) || (id2 > 0 && comp2 == 0) {
			goto good
		}
	}

	/* Didn't match. */
	return false

good:
	mark2Index, ok := data.Mark2Coverage.Index(buffer.Info[j].Glyph)
	if !ok {
		return false
	}

	return c.applyGPOSMarks(data.Marks1, mark1Index, mark2Index, data.Marks2, j)
}
