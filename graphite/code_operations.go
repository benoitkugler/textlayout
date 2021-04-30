package graphite

import "math"

// This file will be pulled into and integrated into a machine implmentation
// DO NOT build directly and under no circumstances ever #include headers in
// here or you will break the direct_machine.
//
// Implementers' notes
// ==================
// You have access to a few primitives and the full C++ code:
//    declare_params(n) Tells the interpreter how many bytes of parameter
//                      space to claim for this instruction uses and
//                      initialises the param pointer.  You *must* before the
//                      first use of param.
//    use_params(n)     Claim n extra bytes of param space beyond what was
//                      claimed using delcare_param.
//    param             A const byte pointer for the parameter space claimed by
//                      this instruction.
//    binop(op)         Implement a binary operation on the stack using the
//                      specified C++ operator.
//    NOT_IMPLEMENTED   Any instruction body containing this will exit the
//                      program with an assertion error.  Instructions that are
//                      not implemented should also be marked NILOP in the
//                      opcodes tables this will cause the code class to spot
//                      them in a live code stream and throw a runtime_error
//                      instead.
//    push(n)           Push the value n onto the stack.
//    pop()             Pop the top most value and return it.
//
//    You have access to the following named fast 'registers':
//        sp        = The pointer to the current top of stack, the last value
//                    pushed.
//        seg       = A reference to the Segment this code is running over.
//        is        = The current slot index
//        isb       = The original base slot index at the start of this rule
//        isf       = The first positioned slot
//        isl       = The last positioned slot
//        ip        = The current instruction pointer
//        endPos    = Position of advance of last cluster
//        dir       = writing system directionality of the font

// #define NOT_IMPLEMENTED     assert(false)
// #define NOT_IMPLEMENTED

// #define binop(op)           const uint32 a = pop(); *sp = uint32(*sp) op a
// #define sbinop(op)          const int32 a = pop(); *sp = int32(*sp) op a
// #define use_params(n)       dp += n

// #define declare_params(n)   const byte * param = dp; \
//                             use_params(n);

// #define push(n)             { *++sp = n; }
// #define pop()               (*sp--)
// #define slotat(x)           (map[(x)])
// #define DIE                 { is=seg.last; status = Machine::died_early; EXIT(1); }
// #define POSITIONED          1

type regbank struct {
	is   *slot
	smap slotMap
	map_ int // index of the current slot into smap.slots
	// slotref * const map_base;
	// const instr * & ip;
	// uint8           direction;
	// int8            flags;
	status uint8
}

func (r *regbank) slotAt(index int8) *slot {
	return r.smap.slots[r.map_+int(index)+1]
}

const stackMax = 1 << 10

type stack struct {
	registers regbank

	vals [stackMax]int32
	top  int // the top of the stack is at vals[top-1]
}

func (st *stack) push(r int32) {
	st.vals[st.top] = r
	st.top += 1
}

func (st *stack) pop() int32 {
	out := st.vals[st.top-1]
	st.top--
	return out
}

func (st *stack) die() bool {
	st.registers.is = st.registers.smap.segment.last
	st.registers.status = machine_died_early
	st.push(1)
	return false
}

func (st *stack) nop() bool {
	return st.top < stackMax
}

func (st *stack) push_byte(dp []byte) bool {
	// declare_params(1);
	st.push(int32(int8(dp[0])))
	return st.top < stackMax
}

func (st *stack) push_byte_u(dp []byte) bool {
	// declare_params(1)
	st.push(int32(dp[0]))
	return st.top < stackMax
}

func (st *stack) push_short(dp []byte) bool {
	// declare_params(2);
	r := int16(dp[0])<<8 | int16(dp[1])
	st.push(int32(r))
	return st.top < stackMax
}

func (st *stack) push_short_u(dp []byte) bool {
	// declare_params(2);
	r := uint16(dp[0])<<8 | uint16(dp[1])
	st.push(int32(r))
	return st.top < stackMax
}

func (st *stack) push_long(dp []byte) bool {
	// declare_params(4);
	r := int32(dp[0])<<24 | int32(dp[1])<<16 | int32(dp[2])<<8 | int32(dp[3])
	st.push(r)
	return st.top < stackMax
}

func (st *stack) add() bool {
	v := st.pop()
	st.vals[st.top-1] += v
	return st.top < stackMax
}

func (st *stack) sub() bool {
	v := st.pop()
	st.vals[st.top-1] -= v
	return st.top < stackMax
}

func (st *stack) mul() bool {
	v := st.pop()
	st.vals[st.top-1] *= v
	return st.top < stackMax
}

func (st *stack) div_() bool {
	b := st.pop()
	a := st.vals[st.top-1]
	if b == 0 || (a == math.MinInt32 && b == -1) {
		return st.die()
	}
	st.vals[st.top-1] = a / b
	return st.top < stackMax
}

func (st *stack) min_() bool {
	a := st.pop()
	b := st.vals[st.top-1]
	if a < b {
		st.vals[st.top-1] = a
	}
	return st.top < stackMax
}

func (st *stack) max_() bool {
	a := st.pop()
	b := st.vals[st.top-1]
	if a > b {
		st.vals[st.top-1] = a
	}
	return st.top < stackMax
}

func (st *stack) neg() bool {
	st.vals[st.top-1] = -st.vals[st.top-1]
	return st.top < stackMax
}

func (st *stack) trunc8() bool {
	st.vals[st.top-1] = int32(uint8(st.vals[st.top-1]))
	return st.top < stackMax
}

func (st *stack) trunc16() bool {
	st.vals[st.top-1] = int32(uint16(st.vals[st.top-1]))
	return st.top < stackMax
}

func (st *stack) cond() bool {
	f := st.pop()
	t := st.pop()
	c := st.pop()
	if c != 0 {
		st.push(t)
	} else {
		st.push(f)
	}
	return st.top < stackMax
}

func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func (st *stack) and_() bool {
	a := st.pop() != 0
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] != 0 && a)
	return st.top < stackMax
}

func (st *stack) or_() bool {
	a := st.pop() != 0
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] != 0 || a)
	return st.top < stackMax
}

func (st *stack) not_() bool {
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] == 0)
	return st.top < stackMax
}

func (st *stack) equal() bool {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] == a)
	return st.top < stackMax
}

func (st *stack) not_eq_() bool {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] != a)
	return st.top < stackMax
}

func (st *stack) less() bool {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] < a)
	return st.top < stackMax
}

func (st *stack) gtr() bool {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] > a)
	return st.top < stackMax
}

func (st *stack) less_eq() bool {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] <= a)
	return st.top < stackMax
}

func (st *stack) gtr_eq() bool {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] >= a)
	return st.top < stackMax
}

// Move the current slot pointer forward one slot (used after we have finished processing
// that slot).
func (st *stack) next() bool {
	if st.registers.map_ >= st.registers.smap.size {
		return st.die()
	}
	if st.registers.is != nil {
		if st.registers.is == st.registers.smap.highwater {
			st.registers.smap.highpassed = true
		}
		st.registers.is = st.registers.is.next
	}
	st.registers.map_++
	return st.top < stackMax
}

// //func (st *stack) next_n() bool {
// //    use_params(1);
// //    NOT_IMPLEMENTED;
//     //declare_params(1);
//     //const size_t num = uint8(*param);
// //ENDOP

// //func (st *stack) copy_next() bool {
// //     if (is) is = is.next;
// //     ++map;
// //ENDOP

// Put the first glyph of the specified class into the output. Normally used when there is only
// one member of the class, and when inserting.
func (st *stack) put_glyph_8bit_obs(dp []byte) bool {
	// declare_params(1);
	outputClass := dp[0]
	seg := st.registers.smap.segment
	st.registers.is.setGlyph(seg, seg.silf.classMap.getClassGlyph(uint16(outputClass), 0))
	return st.top < stackMax
}

// Determine the index of the glyph that was the input in the given slot within the input
// class, and place the corresponding glyph from the output class in the current slot. The slot number
// is relative to the current input position.
func (st *stack) put_subs_8bit_obs(dp []byte) bool {
	// declare_params(3);
	slotRef := int8(dp[0])
	inputClass := dp[1]
	outputClass := dp[2]
	slot := st.registers.slotAt(slotRef)
	if slot != nil {
		seg := st.registers.smap.segment
		index := seg.silf.classMap.findClassIndex(uint16(inputClass), slot.glyphID)
		st.registers.is.setGlyph(seg, seg.silf.classMap.getClassGlyph(uint16(outputClass), index))
	}
	return st.top < stackMax
}

// Copy the glyph that was in the input in the given slot into the current output slot. The slot
// number is relative to the current input position.
func (st *stack) put_copy(dp []byte) bool {
	// declare_params(1);
	slotRef := int8(dp[0])
	is := st.registers.is
	if is != nil && !is.isDeleted() {
		ref := st.registers.slotAt(slotRef)
		if ref != nil && ref != is {
			tempUserAttrs := is.userAttrs
			if is.parent != nil || is.child != nil {
				return st.die()
			}
			prev := is.prev
			next := is.next

			copy(tempUserAttrs, ref.userAttrs[:st.registers.smap.segment.silf.NumUserDefn])
			*is = *ref
			is.child = nil
			is.sibling = nil
			is.userAttrs = tempUserAttrs
			is.next = next
			is.prev = prev
			if is.parent != nil {
				is.parent.child = is
			}
		}
		is.markCopied(false)
		is.markDeleted(false)
	}
	return st.top < stackMax
}

// Insert a new slot before the current slot and make the new slot the current one.
func (st *stack) insert() bool {
	if st.registers.smap.decMax() <= 0 {
		return st.die()
	}
	seg := st.registers.smap.segment
	newSlot := seg.newSlot()
	if newSlot == nil {
		return st.die()
	}
	iss := st.registers.is
	for iss != nil && iss.isDeleted() {
		iss = iss.next
	}
	if iss == nil {
		if seg.last != nil {
			seg.last.next = newSlot
			newSlot.prev = seg.last
			newSlot.before = seg.last.before
			seg.last = newSlot
		} else {
			seg.first = newSlot
			seg.last = newSlot
		}
	} else if iss.prev != nil {
		iss.prev.next = newSlot
		newSlot.prev = iss.prev
		newSlot.before = iss.prev.after
	} else {
		newSlot.prev = nil
		newSlot.before = iss.before
		seg.first = newSlot
	}
	newSlot.next = iss
	if iss != nil {
		iss.prev = newSlot
		newSlot.original = iss.original
		newSlot.after = iss.before
	} else if newSlot.prev != nil {
		newSlot.original = newSlot.prev.original
		newSlot.after = newSlot.prev.after
	} else {
		newSlot.original = seg.defaultOriginal
	}
	if st.registers.is == st.registers.smap.highwater {
		st.registers.smap.highpassed = false
	}
	st.registers.is = newSlot
	seg.numGlyphs += 1
	if st.registers.map_ != 0 {
		st.registers.map_--
	}
	return st.top < stackMax
}

// func (st *stack) delete_() bool {
//     if (!is || is.isDeleted()) DIE
//     is.markDeleted(true);
//     if (is.prev)
//         is.prev.next(is.next);
//     else
//         seg.first(is.next);

//     if (is.next)
//         is.next.prev(is.prev);
//     else
//         seg.last(is.prev);

//     if (is == smap.highwater())
//             smap.highwater(is.next);
//     if (is.prev)
//         is = is.prev;
//     seg.extendLength(-1);
// return st.top < stackMax
// }

// func (st *stack) assoc() bool {
//     declare_params(1);
//     unsigned int  num = uint8(*param);
//     const int8 *  assocs = reinterpret_cast<const int8 *>(param+1);
//     use_params(num);
//     int max = -1;
//     int min = -1;

//     while (num-- > 0)
//     {
//         int sr = *assocs++;
//         slotref ts = slotat(sr);
//         if (ts && (min == -1 || ts.before() < min)) min = ts.before();
//         if (ts && ts.after() > max) max = ts.after();
//     }
//     if (min > -1)   // implies max > -1
//     {
//         is.before(min);
//         is.after(max);
//     }
// return st.top < stackMax
// }

// func (st *stack) cntxt_item() bool {
//     // It turns out this is a cunningly disguised condition forward jump.
//     declare_params(3);
//     const int       is_arg = int8(param[0]);
//     const size_t    iskip  = uint8(param[1]),
//                     dskip  = uint8(param[2]);

//     if (mapb + is_arg != map)
//     {
//         ip += iskip;
//         dp += dskip;
//         push(true);
//     }
// return st.top < stackMax
// }

// func (st *stack) attr_set() bool {
//     declare_params(1);
//     const attrCode      slat = attrCode(uint8(*param));
//     const          int  val  = st.pop();
//     is.setAttr(&seg, slat, 0, val, smap);
// return st.top < stackMax
// }

// func (st *stack) attr_add() bool {
//     declare_params(1);
//     const attrCode      slat = attrCode(uint8(*param));
//     const     uint32_t  val  = st.pop();
//     if ((slat == gr_slatPosX || slat == gr_slatPosY) && (flags & POSITIONED) == 0)
//     {
//         seg.positionSlots(0, *smap.begin(), *(smap.end()-1), seg.currdir());
//         flags |= POSITIONED;
//     }
//     uint32_t res = uint32_t(is.getAttr(&seg, slat, 0));
//     is.setAttr(&seg, slat, 0, int32_t(val + res), smap);
// return st.top < stackMax
// }

// func (st *stack) attr_sub() bool {
//     declare_params(1);
//     const attrCode      slat = attrCode(uint8(*param));
//     const     uint32_t  val  = st.pop();
//     if ((slat == gr_slatPosX || slat == gr_slatPosY) && (flags & POSITIONED) == 0)
//     {
//         seg.positionSlots(0, *smap.begin(), *(smap.end()-1), seg.currdir());
//         flags |= POSITIONED;
//     }
//     uint32_t res = uint32_t(is.getAttr(&seg, slat, 0));
//     is.setAttr(&seg, slat, 0, int32_t(res - val), smap);
// return st.top < stackMax
// }

// func (st *stack) attr_set_slot() bool {
//     declare_params(1);
//     const attrCode  slat   = attrCode(uint8(*param));
//     const int       offset = int(map - smap.begin())*int(slat == gr_slatAttTo);
//     const int       val    = st.pop()  + offset;
//     is.setAttr(&seg, slat, offset, val, smap);
// return st.top < stackMax
// }

// func (st *stack) iattr_set_slot() bool {
//     declare_params(2);
//     const attrCode  slat = attrCode(uint8(param[0]));
//     const uint8     idx  = uint8(param[1]);
//     const int       val  = int(pop()  + (map - smap.begin())*int(slat == gr_slatAttTo));
//     is.setAttr(&seg, slat, idx, val, smap);
// return st.top < stackMax
// }

// func (st *stack) push_slot_attr() bool {
//     declare_params(2);
//     const attrCode      slat     = attrCode(uint8(param[0]));
//     const int           slotRef = int8(param[1]);
//     if ((slat == gr_slatPosX || slat == gr_slatPosY) && (flags & POSITIONED) == 0)
//     {
//         seg.positionSlots(0, *smap.begin(), *(smap.end()-1), seg.currdir());
//         flags |= POSITIONED;
//     }
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         int res = slot.getAttr(&seg, slat, 0);
//         push(res);
//     }
// return st.top < stackMax
// }

// func (st *stack) push_glyph_attr_obs() bool {
//     declare_params(2);
//     const unsigned int  glyph_attr = uint8(param[0]);
//     const int           slotRef   = int8(param[1]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//         push(int32(seg.glyphAttr(slot.gid(), glyph_attr)));
// return st.top < stackMax
// }

// func (st *stack) push_glyph_metric() bool {
//     declare_params(3);
//     const unsigned int  glyph_attr  = uint8(param[0]);
//     const int           slotRef    = int8(param[1]);
//     const signed int    attr_level  = uint8(param[2]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//         push(seg.getGlyphMetric(slot, glyph_attr, attr_level, dir));
// return st.top < stackMax
// }

// func (st *stack) push_feat() bool {
//     declare_params(2);
//     const unsigned int  feat        = uint8(param[0]);
//     const int           slotRef    = int8(param[1]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         uint8 fid = seg.charinfo(slot.original()).fid();
//         push(seg.getFeature(fid, feat));
//     }
// return st.top < stackMax
// }

// func (st *stack) push_att_to_gattr_obs() bool {
//     declare_params(2);
//     const unsigned int  glyph_attr  = uint8(param[0]);
//     const int           slotRef    = int8(param[1]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         slotref att = slot.attachedTo();
//         if (att) slot = att;
//         push(int32(seg.glyphAttr(slot.gid(), glyph_attr)));
//     }
// return st.top < stackMax
// }

// func (st *stack) push_att_to_glyph_metric() bool {
//     declare_params(3);
//     const unsigned int  glyph_attr  = uint8(param[0]);
//     const int           slotRef    = int8(param[1]);
//     const signed int    attr_level  = uint8(param[2]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         slotref att = slot.attachedTo();
//         if (att) slot = att;
//         push(int32(seg.getGlyphMetric(slot, glyph_attr, attr_level, dir)));
//     }
// return st.top < stackMax
// }

// func (st *stack) push_islot_attr() bool {
//     declare_params(3);
//     const attrCode  slat     = attrCode(uint8(param[0]));
//     const int           slotRef = int8(param[1]),
//                         idx      = uint8(param[2]);
//     if ((slat == gr_slatPosX || slat == gr_slatPosY) && (flags & POSITIONED) == 0)
//     {
//         seg.positionSlots(0, *smap.begin(), *(smap.end()-1), seg.currdir());
//         flags |= POSITIONED;
//     }
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         int res = slot.getAttr(&seg, slat, idx);
//         push(res);
//     }
// return st.top < stackMax
// }

// #if 0
// func (st *stack) push_iglyph_attr() bool { // not implemented
//     NOT_IMPLEMENTED;
// return st.top < stackMax
// }
// #endif

// func (st *stack) pop_ret() bool {
//     const uint32 ret = st.pop();
//     EXIT(ret);
// return st.top < stackMax
// }

// func (st *stack) ret_zero() bool {
//     EXIT(0);
// return st.top < stackMax
// }

// func (st *stack) ret_true() bool {
//     EXIT(1);
// return st.top < stackMax
// }

// func (st *stack) iattr_set() bool {
//     declare_params(2);
//     const attrCode      slat = attrCode(uint8(param[0]));
//     const uint8         idx  = uint8(param[1]);
//     const          int  val  = st.pop();
//     is.setAttr(&seg, slat, idx, val, smap);
// return st.top < stackMax
// }

// func (st *stack) iattr_add() bool {
//     declare_params(2);
//     const attrCode      slat = attrCode(uint8(param[0]));
//     const uint8         idx  = uint8(param[1]);
//     const     uint32_t  val  = st.pop();
//     if ((slat == gr_slatPosX || slat == gr_slatPosY) && (flags & POSITIONED) == 0)
//     {
//         seg.positionSlots(0, *smap.begin(), *(smap.end()-1), seg.currdir());
//         flags |= POSITIONED;
//     }
//     uint32_t res = uint32_t(is.getAttr(&seg, slat, idx));
//     is.setAttr(&seg, slat, idx, int32_t(val + res), smap);
// return st.top < stackMax
// }

// func (st *stack) iattr_sub() bool {
//     declare_params(2);
//     const attrCode      slat = attrCode(uint8(param[0]));
//     const uint8         idx  = uint8(param[1]);
//     const     uint32_t  val  = st.pop();
//     if ((slat == gr_slatPosX || slat == gr_slatPosY) && (flags & POSITIONED) == 0)
//     {
//         seg.positionSlots(0, *smap.begin(), *(smap.end()-1), seg.currdir());
//         flags |= POSITIONED;
//     }
//     uint32_t res = uint32_t(is.getAttr(&seg, slat, idx));
//     is.setAttr(&seg, slat, idx, int32_t(res - val), smap);
// return st.top < stackMax
// }

// func (st *stack) push_proc_state() bool {
//     use_params(1);
//     push(1);
// return st.top < stackMax
// }

// func (st *stack) push_version() bool {
//     push(0x00030000);
// return st.top < stackMax
// }

// func (st *stack) put_subs() bool {
//     declare_params(5);
//     const int        slotRef     = int8(param[0]);
//     const unsigned int  inputClass  = uint8(param[1]) << 8
//                                      | uint8(param[2]);
//     const unsigned int  outputClass = uint8(param[3]) << 8
//                                      | uint8(param[4]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         int index = seg.findClassIndex(inputClass, slot.gid());
//         is.setGlyph(&seg, seg.getClassGlyph(outputClass, index));
//     }
// return st.top < stackMax
// }

// #if 0
// func (st *stack) put_subs2() bool { // not implemented
//     NOT_IMPLEMENTED;
// return st.top < stackMax
// }

// func (st *stack) put_subs3() bool { // not implemented
//     NOT_IMPLEMENTED;
// return st.top < stackMax
// }
// #endif

// func (st *stack) put_glyph() bool {
//     declare_params(2);
//     const unsigned int outputClass  = uint8(param[0]) << 8
//                                      | uint8(param[1]);
//     is.setGlyph(&seg, seg.getClassGlyph(outputClass, 0));
// return st.top < stackMax
// }

// func (st *stack) push_glyph_attr() bool {
//     declare_params(3);
//     const unsigned int  glyph_attr  = uint8(param[0]) << 8
//                                     | uint8(param[1]);
//     const int           slotRef    = int8(param[2]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//         push(int32(seg.glyphAttr(slot.gid(), glyph_attr)));
// return st.top < stackMax
// }

// func (st *stack) push_att_to_glyph_attr() bool {
//     declare_params(3);
//     const unsigned int  glyph_attr  = uint8(param[0]) << 8
//                                     | uint8(param[1]);
//     const int           slotRef    = int8(param[2]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         slotref att = slot.attachedTo();
//         if (att) slot = att;
//         push(int32(seg.glyphAttr(slot.gid(), glyph_attr)));
//     }
// return st.top < stackMax
// }

// func (st *stack) temp_copy() bool {
//     slotref newSlot = seg.newSlot();
//     if (!newSlot || !is) DIE;
//     int16 *tempUserAttrs = newSlot.userAttrs();
//     memcpy(newSlot, is, sizeof(Slot));
//     memcpy(tempUserAttrs, is.userAttrs(), seg.numAttrs() * sizeof(uint16));
//     newSlot.userAttrs(tempUserAttrs);
//     newSlot.markCopied(true);
//     *map = newSlot;
// return st.top < stackMax
// }

// func (st *stack) band() bool {
//     binop(&);
// return st.top < stackMax
// }

// func (st *stack) bor() bool {
//     binop(|);
// return st.top < stackMax
// }

// func (st *stack) bnot() bool {
//     *sp = ~*sp;
// return st.top < stackMax
// }

// func (st *stack) setbits() bool {
//     declare_params(4);
//     const uint16 m  = uint16(param[0]) << 8
//                     | uint8(param[1]);
//     const uint16 v  = uint16(param[2]) << 8
//                     | uint8(param[3]);
//     *sp = ((*sp) & ~m) | v;
// return st.top < stackMax
// }

// func (st *stack) set_feat() bool {
//     declare_params(2);
//     const unsigned int  feat        = uint8(param[0]);
//     const int           slotRef    = int8(param[1]);
//     slotref slot = slotat(slotRef);
//     if (slot)
//     {
//         uint8 fid = seg.charinfo(slot.original()).fid();
//         seg.setFeature(fid, feat, st.pop());
//     }
// return st.top < stackMax
// }
