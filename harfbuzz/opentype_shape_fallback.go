package harfbuzz

// ported from harfbuzz/src/hb-ot-shape-fallback.cc Copyright © 2011,2012 Google, Inc. Behdad Esfahbod

const (
	HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW_LEFT  = 200
	HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW       = 202
	HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE       = 214
	HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE_RIGHT = 216
	HB_UNICODE_COMBINING_CLASS_BELOW_LEFT           = 218
	HB_UNICODE_COMBINING_CLASS_BELOW                = 220
	HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT          = 222
	HB_UNICODE_COMBINING_CLASS_LEFT                 = 224
	HB_UNICODE_COMBINING_CLASS_RIGHT                = 226
	HB_UNICODE_COMBINING_CLASS_ABOVE_LEFT           = 228
	HB_UNICODE_COMBINING_CLASS_ABOVE                = 230
	HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT          = 232
	HB_UNICODE_COMBINING_CLASS_DOUBLE_BELOW         = 233
	HB_UNICODE_COMBINING_CLASS_DOUBLE_ABOVE         = 234
)

func recategorize_combining_class(u rune, klass uint8) uint8 {
	if klass >= 200 {
		return klass
	}

	/* Thai / Lao need some per-character work. */
	if (u & ^0xFF) == 0x0E00 {
		if klass == 0 {
			switch u {
			case 0x0E31, 0x0E34, 0x0E35, 0x0E36, 0x0E37, 0x0E47, 0x0E4C, 0x0E4D, 0x0E4E:
				klass = HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT
			case 0x0EB1, 0x0EB4, 0x0EB5, 0x0EB6, 0x0EB7, 0x0EBB, 0x0ECC, 0x0ECD:
				klass = HB_UNICODE_COMBINING_CLASS_ABOVE
			case 0x0EBC:
				klass = HB_UNICODE_COMBINING_CLASS_BELOW
			}
		} else {
			/* Thai virama is below-right */
			if u == 0x0E3A {
				klass = HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT
			}
		}
	}

	switch klass {

	/* Hebrew */

	case 10, /* sheva */
		11, /* hataf segol */
		12, /* hataf patah */
		13, /* hataf qamats */
		14, /* hiriq */
		15, /* tsere */
		16, /* segol */
		17, /* patah */
		18, /* qamats */
		20, /* qubuts */
		22: /* meteg */
		return HB_UNICODE_COMBINING_CLASS_BELOW

	case 23: /* rafe */
		return HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE

	case 24: /* shin dot */
		return HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT

	case 25, /* sin dot */
		19: /* holam */
		return HB_UNICODE_COMBINING_CLASS_ABOVE_LEFT

	case 26: /* point varika */
		return HB_UNICODE_COMBINING_CLASS_ABOVE

	case 21: /* dagesh */

	/* Arabic and Syriac */

	case 27, /* fathatan */
		28, /* dammatan */
		30, /* fatha */
		31, /* damma */
		33, /* shadda */
		34, /* sukun */
		35, /* superscript alef */
		36: /* superscript alaph */
		return HB_UNICODE_COMBINING_CLASS_ABOVE

	case 29, /* kasratan */
		32: /* kasra */
		return HB_UNICODE_COMBINING_CLASS_BELOW

	/* Thai */

	case 103: /* sara u / sara uu */
		return HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT

	case 107: /* mai */
		return HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT

	/* Lao */

	case 118: /* sign u / sign uu */
		return HB_UNICODE_COMBINING_CLASS_BELOW

	case 122: /* mai */
		return HB_UNICODE_COMBINING_CLASS_ABOVE

	/* Tibetan */

	case 129: /* sign aa */
		return HB_UNICODE_COMBINING_CLASS_BELOW

	case 130: /* sign i*/
		return HB_UNICODE_COMBINING_CLASS_ABOVE

	case 132: /* sign u */
		return HB_UNICODE_COMBINING_CLASS_BELOW

	}

	return klass
}

func fallbackMarkPositionRecategorizeMarks(buffer *hb_buffer_t) {
	for i, info := range buffer.info {
		if info.unicode.generalCategory() == nonSpacingMark {
			combining_class := info.getModifiedCombiningClass()
			combining_class = recategorize_combining_class(info.codepoint, combining_class)
			buffer.info[i].setModifiedCombiningClass(combining_class)
		}
	}
}

//  static void
//  zero_mark_advances (hb_buffer_t *buffer,
// 			 unsigned int start,
// 			 unsigned int end,
// 			 bool adjust_offsets_when_zeroing)
//  {
//    hb_glyph_info_t *info = buffer.info;
//    for (unsigned int i = start; i < end; i++)
// 	 if (_hb_glyph_info_get_general_category (&info[i]) == HB_UNICODE_GENERAL_CATEGORY_NON_SPACING_MARK)
// 	 {
// 	   if (adjust_offsets_when_zeroing)
// 	   {
// 	 buffer.pos[i].x_offset -= buffer.pos[i].x_advance;
// 	 buffer.pos[i].y_offset -= buffer.pos[i].y_advance;
// 	   }
// 	   buffer.pos[i].x_advance = 0;
// 	   buffer.pos[i].y_advance = 0;
// 	 }
//  }

//  static inline void
//  position_mark (const hb_ot_shape_plan_t *plan HB_UNUSED,
// 			hb_font_t *font,
// 			hb_buffer_t  *buffer,
// 			hb_glyph_extents_t &base_extents,
// 			unsigned int i,
// 			unsigned int combining_class)
//  {
//    hb_glyph_extents_t mark_extents;
//    if (!font.get_glyph_extents (buffer.info[i].codepoint, &mark_extents))
// 	 return;

//    hb_position_t y_gap = font.y_scale / 16;

//    hb_glyph_position_t &pos = buffer.pos[i];
//    pos.x_offset = pos.y_offset = 0;

//    /* We don't position LEFT and RIGHT marks. */

//    /* X positioning */
//    switch (combining_class)
//    {
// 	 case HB_UNICODE_COMBINING_CLASS_DOUBLE_BELOW:
// 	 case HB_UNICODE_COMBINING_CLASS_DOUBLE_ABOVE:
// 	   if (buffer.props.direction == HB_DIRECTION_LTR) {
// 	 pos.x_offset += base_extents.x_bearing + base_extents.width - mark_extents.width / 2 - mark_extents.x_bearing;
// 	 break;
// 	   } else if (buffer.props.direction == HB_DIRECTION_RTL) {
// 	 pos.x_offset += base_extents.x_bearing - mark_extents.width / 2 - mark_extents.x_bearing;
// 	 break;
// 	   }
// 	   HB_FALLTHROUGH;

// 	 default:
// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW:
// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE:
// 	 case HB_UNICODE_COMBINING_CLASS_BELOW:
// 	 case HB_UNICODE_COMBINING_CLASS_ABOVE:
// 	   /* Center align. */
// 	   pos.x_offset += base_extents.x_bearing + (base_extents.width - mark_extents.width) / 2 - mark_extents.x_bearing;
// 	   break;

// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW_LEFT:
// 	 case HB_UNICODE_COMBINING_CLASS_BELOW_LEFT:
// 	 case HB_UNICODE_COMBINING_CLASS_ABOVE_LEFT:
// 	   /* Left align. */
// 	   pos.x_offset += base_extents.x_bearing - mark_extents.x_bearing;
// 	   break;

// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE_RIGHT:
// 	 case HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT:
// 	 case HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT:
// 	   /* Right align. */
// 	   pos.x_offset += base_extents.x_bearing + base_extents.width - mark_extents.width - mark_extents.x_bearing;
// 	   break;
//    }

//    /* Y positioning */
//    switch (combining_class)
//    {
// 	 case HB_UNICODE_COMBINING_CLASS_DOUBLE_BELOW:
// 	 case HB_UNICODE_COMBINING_CLASS_BELOW_LEFT:
// 	 case HB_UNICODE_COMBINING_CLASS_BELOW:
// 	 case HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT:
// 	   /* Add gap, fall-through. */
// 	   base_extents.height -= y_gap;
// 	   HB_FALLTHROUGH;

// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW_LEFT:
// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW:
// 	   pos.y_offset = base_extents.y_bearing + base_extents.height - mark_extents.y_bearing;
// 	   /* Never shift up "below" marks. */
// 	   if ((y_gap > 0) == (pos.y_offset > 0))
// 	   {
// 	 base_extents.height -= pos.y_offset;
// 	 pos.y_offset = 0;
// 	   }
// 	   base_extents.height += mark_extents.height;
// 	   break;

// 	 case HB_UNICODE_COMBINING_CLASS_DOUBLE_ABOVE:
// 	 case HB_UNICODE_COMBINING_CLASS_ABOVE_LEFT:
// 	 case HB_UNICODE_COMBINING_CLASS_ABOVE:
// 	 case HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT:
// 	   /* Add gap, fall-through. */
// 	   base_extents.y_bearing += y_gap;
// 	   base_extents.height -= y_gap;
// 	   HB_FALLTHROUGH;

// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE:
// 	 case HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE_RIGHT:
// 	   pos.y_offset = base_extents.y_bearing - (mark_extents.y_bearing + mark_extents.height);
// 	   /* Don't shift down "above" marks too much. */
// 	   if ((y_gap > 0) != (pos.y_offset > 0))
// 	   {
// 	 int correction = -pos.y_offset / 2;
// 	 base_extents.y_bearing += correction;
// 	 base_extents.height -= correction;
// 	 pos.y_offset += correction;
// 	   }
// 	   base_extents.y_bearing -= mark_extents.height;
// 	   base_extents.height += mark_extents.height;
// 	   break;
//    }
//  }

//  static inline void
//  position_around_base (const hb_ot_shape_plan_t *plan,
// 			   hb_font_t *font,
// 			   hb_buffer_t  *buffer,
// 			   unsigned int base,
// 			   unsigned int end,
// 			   bool adjust_offsets_when_zeroing)
//  {
//    hb_direction_t horiz_dir = HB_DIRECTION_INVALID;

//    buffer.unsafe_to_break (base, end);

//    hb_glyph_extents_t base_extents;
//    if (!font.get_glyph_extents (buffer.info[base].codepoint,
// 				 &base_extents))
//    {
// 	 /* If extents don't work, zero marks and go home. */
// 	 zero_mark_advances (buffer, base + 1, end, adjust_offsets_when_zeroing);
// 	 return;
//    }
//    base_extents.y_bearing += buffer.pos[base].y_offset;
//    /* Use horizontal advance for horizontal positioning.
// 	* Generally a better idea.  Also works for zero-ink glyphs.  See:
// 	* https://github.com/harfbuzz/harfbuzz/issues/1532 */
//    base_extents.x_bearing = 0;
//    base_extents.width = font.get_glyph_h_advance (buffer.info[base].codepoint);

//    unsigned int lig_id = _hb_glyph_info_get_lig_id (&buffer.info[base]);
//    /* Use integer for num_lig_components such that it doesn't convert to unsigned
// 	* when we divide or multiply by it. */
//    int num_lig_components = _hb_glyph_info_get_lig_num_comps (&buffer.info[base]);

//    hb_position_t x_offset = 0, y_offset = 0;
//    if (HB_DIRECTION_IS_FORWARD (buffer.props.direction)) {
// 	 x_offset -= buffer.pos[base].x_advance;
// 	 y_offset -= buffer.pos[base].y_advance;
//    }

//    hb_glyph_extents_t component_extents = base_extents;
//    int last_lig_component = -1;
//    unsigned int last_combining_class = 255;
//    hb_glyph_extents_t cluster_extents = base_extents; /* Initialization is just to shut gcc up. */
//    hb_glyph_info_t *info = buffer.info;
//    for (unsigned int i = base + 1; i < end; i++)
// 	 if (_hb_glyph_info_get_modified_combining_class (&info[i]))
// 	 {
// 	   if (num_lig_components > 1) {
// 	 unsigned int this_lig_id = _hb_glyph_info_get_lig_id (&info[i]);
// 	 int this_lig_component = _hb_glyph_info_get_lig_comp (&info[i]) - 1;
// 	 /* Conditions for attaching to the last component. */
// 	 if (!lig_id || lig_id != this_lig_id || this_lig_component >= num_lig_components)
// 	   this_lig_component = num_lig_components - 1;
// 	 if (last_lig_component != this_lig_component)
// 	 {
// 	   last_lig_component = this_lig_component;
// 	   last_combining_class = 255;
// 	   component_extents = base_extents;
// 	   if (unlikely (horiz_dir == HB_DIRECTION_INVALID)) {
// 		 if (HB_DIRECTION_IS_HORIZONTAL (plan.props.direction))
// 		   horiz_dir = plan.props.direction;
// 		 else
// 		   horiz_dir = hb_script_get_horizontal_direction (plan.props.script);
// 	   }
// 	   if (horiz_dir == HB_DIRECTION_LTR)
// 		 component_extents.x_bearing += (this_lig_component * component_extents.width) / num_lig_components;
// 	   else
// 		 component_extents.x_bearing += ((num_lig_components - 1 - this_lig_component) * component_extents.width) / num_lig_components;
// 	   component_extents.width /= num_lig_components;
// 	 }
// 	   }

// 	   unsigned int this_combining_class = _hb_glyph_info_get_modified_combining_class (&info[i]);
// 	   if (last_combining_class != this_combining_class)
// 	   {
// 	 last_combining_class = this_combining_class;
// 	 cluster_extents = component_extents;
// 	   }

// 	   position_mark (plan, font, buffer, cluster_extents, i, this_combining_class);

// 	   buffer.pos[i].x_advance = 0;
// 	   buffer.pos[i].y_advance = 0;
// 	   buffer.pos[i].x_offset += x_offset;
// 	   buffer.pos[i].y_offset += y_offset;

// 	 } else {
// 	   if (HB_DIRECTION_IS_FORWARD (buffer.props.direction)) {
// 	 x_offset -= buffer.pos[i].x_advance;
// 	 y_offset -= buffer.pos[i].y_advance;
// 	   } else {
// 	 x_offset += buffer.pos[i].x_advance;
// 	 y_offset += buffer.pos[i].y_advance;
// 	   }
// 	 }
//  }

//  static inline void
//  position_cluster (const hb_ot_shape_plan_t *plan,
// 		   hb_font_t *font,
// 		   hb_buffer_t  *buffer,
// 		   unsigned int start,
// 		   unsigned int end,
// 		   bool adjust_offsets_when_zeroing)
//  {
//    if (end - start < 2)
// 	 return;

//    /* Find the base glyph */
//    hb_glyph_info_t *info = buffer.info;
//    for (unsigned int i = start; i < end; i++)
// 	 if (!_hb_glyph_info_is_unicode_mark (&info[i]))
// 	 {
// 	   /* Find mark glyphs */
// 	   unsigned int j;
// 	   for (j = i + 1; j < end; j++)
// 	 if (!_hb_glyph_info_is_unicode_mark (&info[j]))
// 	   break;

// 	   position_around_base (plan, font, buffer, i, j, adjust_offsets_when_zeroing);

// 	   i = j - 1;
// 	 }
//  }

//  void
//  _hb_ot_shape_fallback_mark_position (const hb_ot_shape_plan_t *plan,
// 					  hb_font_t *font,
// 					  hb_buffer_t  *buffer,
// 					  bool adjust_offsets_when_zeroing)
//  {
//  #ifdef HB_NO_OT_SHAPE_FALLBACK
//    return;
//  #endif

//    _hb_buffer_assert_gsubgpos_vars (buffer);

//    unsigned int start = 0;
//    unsigned int count = buffer.len;
//    hb_glyph_info_t *info = buffer.info;
//    for (unsigned int i = 1; i < count; i++)
// 	 if (likely (!_hb_glyph_info_is_unicode_mark (&info[i]))) {
// 	   position_cluster (plan, font, buffer, start, i, adjust_offsets_when_zeroing);
// 	   start = i;
// 	 }
//    position_cluster (plan, font, buffer, start, count, adjust_offsets_when_zeroing);
//  }

//  #ifndef HB_DISABLE_DEPRECATED
//  struct hb_ot_shape_fallback_kern_driver_t
//  {
//    hb_ot_shape_fallback_kern_driver_t (hb_font_t   *font_,
// 					   hb_buffer_t *buffer) :
// 	 font (font_), direction (buffer.props.direction) {}

//    hb_position_t get_kerning (hb_codepoint_t first, hb_codepoint_t second) const
//    {
// 	 hb_position_t kern = 0;
// 	 font.get_glyph_kerning_for_direction (first, second,
// 						direction,
// 						&kern, &kern);
// 	 return kern;
//    }

//    hb_font_t *font;
//    hb_direction_t direction;
//  };
//  #endif

//  /* Performs font-assisted kerning. */
//  void
//  _hb_ot_shape_fallback_kern (const hb_ot_shape_plan_t *plan,
// 				 hb_font_t *font,
// 				 hb_buffer_t *buffer)
//  {
//  #ifdef HB_NO_OT_SHAPE_FALLBACK
//    return;
//  #endif

//  #ifndef HB_DISABLE_DEPRECATED
//    if (HB_DIRECTION_IS_HORIZONTAL (buffer.props.direction) ?
// 	   !font.has_glyph_h_kerning_func () :
// 	   !font.has_glyph_v_kerning_func ())
// 	 return;

//    bool reverse = HB_DIRECTION_IS_BACKWARD (buffer.props.direction);

//    if (reverse)
// 	 buffer.reverse ();

//    hb_ot_shape_fallback_kern_driver_t driver (font, buffer);
//    OT::hb_kern_machine_t<hb_ot_shape_fallback_kern_driver_t> machine (driver);
//    machine.kern (font, buffer, plan.kern_mask, false);

//    if (reverse)
// 	 buffer.reverse ();
//  #endif
//  }

//  /* Adjusts width of various spaces. */
//  void
//  _hb_ot_shape_fallback_spaces (const hb_ot_shape_plan_t *plan HB_UNUSED,
// 				   hb_font_t *font,
// 				   hb_buffer_t  *buffer)
//  {
//    hb_glyph_info_t *info = buffer.info;
//    hb_glyph_position_t *pos = buffer.pos;
//    bool horizontal = HB_DIRECTION_IS_HORIZONTAL (buffer.props.direction);
//    unsigned int count = buffer.len;
//    for (unsigned int i = 0; i < count; i++)
// 	 if (_hb_glyph_info_is_unicode_space (&info[i]) && !_hb_glyph_info_ligated (&info[i]))
// 	 {
// 	   hb_unicode_funcs_t::space_t space_type = _hb_glyph_info_get_unicode_space_fallback_type (&info[i]);
// 	   hb_codepoint_t glyph;
// 	   typedef hb_unicode_funcs_t t;
// 	   switch (space_type)
// 	   {
// 	 case t::NOT_SPACE: /* Shouldn't happen. */
// 	 case t::SPACE:
// 	   break;

// 	 case t::SPACE_EM:
// 	 case t::SPACE_EM_2:
// 	 case t::SPACE_EM_3:
// 	 case t::SPACE_EM_4:
// 	 case t::SPACE_EM_5:
// 	 case t::SPACE_EM_6:
// 	 case t::SPACE_EM_16:
// 	   if (horizontal)
// 		 pos[i].x_advance = +(font.x_scale + ((int) space_type)/2) / (int) space_type;
// 	   else
// 		 pos[i].y_advance = -(font.y_scale + ((int) space_type)/2) / (int) space_type;
// 	   break;

// 	 case t::SPACE_4_EM_18:
// 	   if (horizontal)
// 		 pos[i].x_advance = (int64_t) +font.x_scale * 4 / 18;
// 	   else
// 		 pos[i].y_advance = (int64_t) -font.y_scale * 4 / 18;
// 	   break;

// 	 case t::SPACE_FIGURE:
// 	   for (char u = '0'; u <= '9'; u++)
// 		 if (font.get_nominal_glyph (u, &glyph))
// 		 {
// 		   if (horizontal)
// 		 pos[i].x_advance = font.get_glyph_h_advance (glyph);
// 		   else
// 		 pos[i].y_advance = font.get_glyph_v_advance (glyph);
// 		   break;
// 		 }
// 	   break;

// 	 case t::SPACE_PUNCTUATION:
// 	   if (font.get_nominal_glyph ('.', &glyph) ||
// 		   font.get_nominal_glyph (',', &glyph))
// 	   {
// 		 if (horizontal)
// 		   pos[i].x_advance = font.get_glyph_h_advance (glyph);
// 		 else
// 		   pos[i].y_advance = font.get_glyph_v_advance (glyph);
// 	   }
// 	   break;

// 	 case t::SPACE_NARROW:
// 	   /* Half-space?
// 		* Unicode doc https://unicode.org/charts/PDF/U2000.pdf says ~1/4 or 1/5 of EM.
// 		* However, in my testing, many fonts have their regular space being about that
// 		* size.  To me, a percentage of the space width makes more sense.  Half is as
// 		* good as any. */
// 	   if (horizontal)
// 		 pos[i].x_advance /= 2;
// 	   else
// 		 pos[i].y_advance /= 2;
// 	   break;
// 	   }
// 	 }
//  }
