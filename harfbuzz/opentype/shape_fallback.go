package opentype

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

func fallbackMarkPositionRecategorizeMarks(buffer *cm.Buffer) {
	for i, info := range buffer.Info {
		if info.unicode.GeneralCategory() == nonSpacingMark {
			combining_class := info.GetModifiedCombiningClass()
			combining_class = recategorize_combining_class(info.Codepoint, combining_class)
			buffer.Info[i].SetModifiedCombiningClass(combining_class)
		}
	}
}

func zeroMarkAdvances(buffer *cm.Buffer, start, end int, adjustOffsetsWhenZeroing bool) {
	info := buffer.Info
	for i := start; i < end; i++ {
		if info[i].unicode.GeneralCategory() != nonSpacingMark {
			continue
		}
		if adjustOffsetsWhenZeroing {
			buffer.Pos[i].XOffset -= buffer.Pos[i].XAdvance
			buffer.Pos[i].y_offset -= buffer.Pos[i].y_advance
		}
		buffer.Pos[i].XAdvance = 0
		buffer.Pos[i].y_advance = 0
	}
}

func positionMark(font *cm.Font, buffer *cm.Buffer, baseExtents *hb_glyph_extents_t,
	i int, combiningClass uint8) {
	markExtents, ok := font.face.GetGlyphExtents(buffer.Info[i].Codepoint)
	if !ok {
		return
	}

	yGap := font.y_scale / 16

	pos := &buffer.Pos[i]
	pos.XOffset = 0
	pos.y_offset = 0

	// we don't position LEFT and RIGHT marks.

	// X positioning
	switch combiningClass {
	case HB_UNICODE_COMBINING_CLASS_DOUBLE_BELOW, HB_UNICODE_COMBINING_CLASS_DOUBLE_ABOVE:
		if buffer.Props.Direction == cm.HB_DIRECTION_LTR {
			pos.XOffset += baseExtents.XBearing + baseExtents.Width - markExtents.Width/2 - markExtents.XBearing
			break
		} else if buffer.Props.Direction == cm.HB_DIRECTION_RTL {
			pos.XOffset += baseExtents.XBearing - markExtents.Width/2 - markExtents.XBearing
			break
		}
		fallthrough
	default:
		fallthrough
	case HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW, HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE, HB_UNICODE_COMBINING_CLASS_BELOW, HB_UNICODE_COMBINING_CLASS_ABOVE:
		/* Center align. */
		pos.XOffset += baseExtents.XBearing + (baseExtents.Width-markExtents.Width)/2 - markExtents.XBearing

	case HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW_LEFT, HB_UNICODE_COMBINING_CLASS_BELOW_LEFT, HB_UNICODE_COMBINING_CLASS_ABOVE_LEFT:
		/* Left align. */
		pos.XOffset += baseExtents.XBearing - markExtents.XBearing

	case HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE_RIGHT, HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT, HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT:
		/* Right align. */
		pos.XOffset += baseExtents.XBearing + baseExtents.Width - markExtents.Width - markExtents.XBearing
	}

	/* Y positioning */
	switch combiningClass {
	case HB_UNICODE_COMBINING_CLASS_DOUBLE_BELOW, HB_UNICODE_COMBINING_CLASS_BELOW_LEFT, HB_UNICODE_COMBINING_CLASS_BELOW, HB_UNICODE_COMBINING_CLASS_BELOW_RIGHT:
		/* Add gap, fall-through. */
		baseExtents.Height -= yGap
		fallthrough

	case HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW_LEFT, HB_UNICODE_COMBINING_CLASS_ATTACHED_BELOW:
		pos.y_offset = baseExtents.YBearing + baseExtents.Height - markExtents.YBearing
		/* Never shift up "below" marks. */
		if (yGap > 0) == (pos.y_offset > 0) {
			baseExtents.Height -= pos.y_offset
			pos.y_offset = 0
		}
		baseExtents.Height += markExtents.Height

	case HB_UNICODE_COMBINING_CLASS_DOUBLE_ABOVE, HB_UNICODE_COMBINING_CLASS_ABOVE_LEFT, HB_UNICODE_COMBINING_CLASS_ABOVE, HB_UNICODE_COMBINING_CLASS_ABOVE_RIGHT:
		/* Add gap, fall-through. */
		baseExtents.YBearing += yGap
		baseExtents.Height -= yGap
		fallthrough
	case HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE, HB_UNICODE_COMBINING_CLASS_ATTACHED_ABOVE_RIGHT:
		pos.y_offset = baseExtents.YBearing - (markExtents.YBearing + markExtents.Height)
		/* Don't shift down "above" marks too much. */
		if (yGap > 0) != (pos.y_offset > 0) {
			correction := -pos.y_offset / 2
			baseExtents.YBearing += correction
			baseExtents.Height -= correction
			pos.y_offset += correction
		}
		baseExtents.YBearing -= markExtents.Height
		baseExtents.Height += markExtents.Height
	}
}

func positionAroundBase(plan *hb_ot_shape_plan_t, font *cm.Font, buffer *cm.Buffer,
	base, end int, adjustOffsetsWhenZeroing bool) {

	buffer.UnsafeToBreak(base, end)

	baseExtents, ok := font.Face.GetGlyphExtents(buffer.Info[base].Codepoint)
	if !ok {
		// if extents don't work, zero marks and go home.
		zeroMarkAdvances(buffer, base+1, end, adjustOffsetsWhenZeroing)
		return
	}
	baseExtents.YBearing += buffer.Pos[base].y_offset
	/* Use horizontal advance for horizontal positioning.
	* Generally a better idea.  Also works for zero-ink glyphs.  See:
	* https://github.com/harfbuzz/harfbuzz/issues/1532 */
	baseExtents.XBearing = 0
	baseExtents.Width = font.GetGlyphHAdvance(buffer.Info[base].Codepoint)

	ligId := buffer.Info[base].getLigId()
	numLigComponents := int32(buffer.Info[base].getLigNumComps())

	var x_offset, y_offset Position
	if !buffer.props.direction.isBackward() {
		x_offset -= buffer.Pos[base].XAdvance
		y_offset -= buffer.Pos[base].y_advance
	}

	var horizDir Direction
	componentExtents := baseExtents
	lastLigComponent := int32(-1)
	lastCombiningClass := uint8(255)
	clusterExtents := baseExtents
	info := buffer.Info
	for i := base + 1; i < end; i++ {
		if info[i].GetModifiedCombiningClass() != 0 {
			if numLigComponents > 1 {
				thisLigId := info[i].getLigId()
				thisLigComponent := int32(info[i].GetLigComp() - 1)
				// conditions for attaching to the last component.
				if ligId == 0 || ligId != thisLigId || thisLigComponent >= numLigComponents {
					thisLigComponent = numLigComponents - 1
				}
				if lastLigComponent != thisLigComponent {
					lastLigComponent = thisLigComponent
					lastCombiningClass = 255
					componentExtents = baseExtents
					if horizDir == HB_DIRECTION_INVALID {
						if plan.props.direction.isHorizontal() {
							horizDir = plan.props.direction
						} else {
							horizDir = hb_script_get_horizontal_direction(plan.props.script)
						}
					}
					if horizDir == HB_DIRECTION_LTR {
						componentExtents.XBearing += (thisLigComponent * componentExtents.Width) / numLigComponents
					} else {
						componentExtents.XBearing += ((numLigComponents - 1 - thisLigComponent) * componentExtents.Width) / numLigComponents
					}
					componentExtents.Width /= numLigComponents
				}
			}

			thisCombiningClass := info[i].GetModifiedCombiningClass()
			if lastCombiningClass != thisCombiningClass {
				lastCombiningClass = thisCombiningClass
				clusterExtents = componentExtents
			}

			positionMark(font, buffer, &clusterExtents, i, thisCombiningClass)

			buffer.Pos[i].XAdvance = 0
			buffer.Pos[i].y_advance = 0
			buffer.Pos[i].XOffset += x_offset
			buffer.Pos[i].y_offset += y_offset

		} else {
			if buffer.props.direction.isBackward() {
				x_offset += buffer.Pos[i].XAdvance
				y_offset += buffer.Pos[i].y_advance
			} else {
				x_offset -= buffer.Pos[i].XAdvance
				y_offset -= buffer.Pos[i].y_advance
			}
		}
	}
}

func positionCluster(plan *hb_ot_shape_plan_t, font *cm.Font, buffer *cm.Buffer,
	start, end int, adjustOffsetsWhenZeroing bool) {
	if end-start < 2 {
		return
	}

	// find the base glyph
	info := buffer.Info
	for i := start; i < end; i++ {
		if !info[i].isUnicodeMark() {
			// find mark glyphs
			var j int
			for j = i + 1; j < end; j++ {
				if !info[j].isUnicodeMark() {
					break
				}
			}

			positionAroundBase(plan, font, buffer, i, j, adjustOffsetsWhenZeroing)

			i = j - 1
		}
	}
}

func fallbackMarkPosition(plan *hb_ot_shape_plan_t, font *cm.Font, buffer *cm.Buffer,
	adjustOffsetsWhenZeroing bool) {
	var start int
	info := buffer.Info
	for i := 1; i < len(info); i++ {
		if !info[i].isUnicodeMark() {
			positionCluster(plan, font, buffer, start, i, adjustOffsetsWhenZeroing)
			start = i
		}
	}
	positionCluster(plan, font, buffer, start, len(info), adjustOffsetsWhenZeroing)
}

//  #ifndef HB_DISABLE_DEPRECATED
//  struct hb_ot_shape_fallback_kern_driver_t
//  {
//    hb_ot_shape_fallback_kern_driver_t (Font   *font_,
// 					   buffer *cm.Buffer) :
// 	 font (font_), direction (buffer.props.direction) {}

//    Position get_kerning (hb_codepoint_t first, hb_codepoint_t second) const
//    {
// 	 Position kern = 0;
// 	 font.get_glyph_kerning_for_direction (first, second,
// 						direction,
// 						&kern, &kern);
// 	 return kern;
//    }

//    Font *font;
//    Direction direction;
//  };
//  #endif

//  /* Performs font-assisted kerning. */
//  void
//  _hb_ot_shape_fallback_kern (const hb_ot_shape_plan_t *plan,
// 				 Font *font,
// 				 buffer *cm.Buffer)
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

// adjusts width of various spaces.
func fallbackSpaces(font *cm.Font, buffer *cm.Buffer) {
	info := buffer.Info
	pos := buffer.Pos
	horizontal := buffer.props.direction.isHorizontal()
	for i, inf := range info {
		if !inf.isUnicodeSpace() || inf.isLigated() {
			continue
		}

		space_type := inf.getUnicodeSpaceFallbackType()

		switch space_type {
		case NOT_SPACE, SPACE: // shouldn't happen
		case SPACE_EM, SPACE_EM_2, SPACE_EM_3, SPACE_EM_4, SPACE_EM_5, SPACE_EM_6, SPACE_EM_16:
			if horizontal {
				pos[i].XAdvance = +(font.XScale + int32(space_type)/2) / int32(space_type)
			} else {
				pos[i].y_advance = -(font.y_scale + int32(space_type)/2) / int32(space_type)
			}
		case SPACE_4_EM_18:
			if horizontal {
				pos[i].XAdvance = +font.XScale * 4 / 18
			} else {
				pos[i].y_advance = -font.y_scale * 4 / 18
			}
		case SPACE_FIGURE:
			for u := '0'; u <= '9'; u++ {
				if glyph, ok := font.face.GetNominalGlyph(u); ok {
					if horizontal {
						pos[i].XAdvance = font.GetGlyphHAdvance(glyph)
					} else {
						pos[i].y_advance = font.get_glyph_v_advance(glyph)
					}
				}
			}
		case SPACE_PUNCTUATION:
			glyph, ok := font.face.GetNominalGlyph('.')
			if !ok {
				glyph, ok = font.face.GetNominalGlyph(',')
			}
			if ok {
				if horizontal {
					pos[i].XAdvance = font.GetGlyphHAdvance(glyph)
				} else {
					pos[i].y_advance = font.get_glyph_v_advance(glyph)
				}
			}
		case SPACE_NARROW:
			/* Half-space?
			* Unicode doc https://unicode.org/charts/PDF/U2000.pdf says ~1/4 or 1/5 of EM.
			* However, in my testing, many fonts have their regular space being about that
			* size.  To me, a percentage of the space width makes more sense.  Half is as
			* good as any. */
			if horizontal {
				pos[i].XAdvance /= 2
			} else {
				pos[i].y_advance /= 2
			}
		}
	}
}
