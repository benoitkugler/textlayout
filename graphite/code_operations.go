package graphite

import (
	"math"
)

const VARARGS = 0xff

// types or parameters are: (.. is inclusive)
//      number - any byte
//      output_class - 0 .. silf.m_nClass
//      input_class - 0 .. silf.m_nClass
//      sattrnum - 0 .. 29 (gr_slatJWidth) , 55 (gr_slatUserDefn)
//      attrid - 0 .. silf.numUser() where sattrnum == 55; 0..silf.m_iMaxComp where sattrnum == 15 otherwise 0
//      gattrnum - 0 .. face->getGlyphFaceCache->numAttrs()
//      gmetric - 0 .. 11 (kgmetDescent)
//      featidx - 0 .. face.numFeatures()
//      level - any byte
var opcode_table = [MAX_OPCODE + 1]struct {
	impl      [2]instrImpl // indexed by int(constraint)
	name      string
	paramSize uint8 // number of paramerters needed or VARARGS
}{
	{[2]instrImpl{nop, nop}, "NOP", 0},

	{[2]instrImpl{push_byte, push_byte}, "PUSH_BYTE", 1},          // number
	{[2]instrImpl{push_byte_u, push_byte_u}, "PUSH_BYTE_U", 1},    // number
	{[2]instrImpl{push_short, push_short}, "PUSH_SHORT", 2},       // number number
	{[2]instrImpl{push_short_u, push_short_u}, "PUSH_SHORT_U", 2}, // number number
	{[2]instrImpl{push_long, push_long}, "PUSH_LONG", 4},          // number number number number

	{[2]instrImpl{add, add}, "ADD", 0},
	{[2]instrImpl{sub, sub}, "SUB", 0},
	{[2]instrImpl{mul, mul}, "MUL", 0},
	{[2]instrImpl{div_, div_}, "DIV", 0},
	{[2]instrImpl{min_, min_}, "MIN", 0},
	{[2]instrImpl{max_, max_}, "MAX", 0},
	{[2]instrImpl{neg, neg}, "NEG", 0},
	{[2]instrImpl{trunc8, trunc8}, "TRUNC8", 0},
	{[2]instrImpl{trunc16, trunc16}, "TRUNC16", 0},

	{[2]instrImpl{cond, cond}, "COND", 0},
	{[2]instrImpl{and_, and_}, "AND", 0}, // 0x10
	{[2]instrImpl{or_, or_}, "OR", 0},
	{[2]instrImpl{not_, not_}, "NOT", 0},
	{[2]instrImpl{equal, equal}, "EQUAL", 0},
	{[2]instrImpl{not_eq_, not_eq_}, "NOT_EQ", 0},
	{[2]instrImpl{less, less}, "LESS", 0},
	{[2]instrImpl{gtr, gtr}, "GTR", 0},
	{[2]instrImpl{less_eq, less_eq}, "LESS_EQ", 0},
	{[2]instrImpl{gtr_eq, gtr_eq}, "GTR_EQ", 0}, // 0x18

	{[2]instrImpl{next, nil}, "NEXT", 0},
	{[2]instrImpl{nil, nil}, "NEXT_N", 1}, // number <= smap.end - map
	{[2]instrImpl{next, nil}, "COPY_NEXT", 0},
	{[2]instrImpl{put_glyph_8bit_obs, nil}, "PUT_GLYPH_8BIT_OBS", 1}, // output_class
	{[2]instrImpl{put_subs_8bit_obs, nil}, "PUT_SUBS_8BIT_OBS", 3},   // slot input_class output_class
	{[2]instrImpl{put_copy, nil}, "PUT_COPY", 1},                     // slot
	{[2]instrImpl{insert, nil}, "INSERT", 0},
	{[2]instrImpl{delete_, nil}, "DELETE", 0}, // 0x20
	{[2]instrImpl{assoc, nil}, "ASSOC", VARARGS},
	{[2]instrImpl{nil, cntxt_item}, "CNTXT_ITEM", 2}, // slot offset

	{[2]instrImpl{attr_set, nil}, "ATTR_SET", 1},                                       // sattrnum
	{[2]instrImpl{attr_add, nil}, "ATTR_ADD", 1},                                       // sattrnum
	{[2]instrImpl{attr_sub, nil}, "ATTR_SUB", 1},                                       // sattrnum
	{[2]instrImpl{attr_set_slot, nil}, "ATTR_SET_SLOT", 1},                             // sattrnum
	{[2]instrImpl{iattr_set_slot, nil}, "IATTR_SET_SLOT", 2},                           // sattrnum attrid
	{[2]instrImpl{push_slot_attr, push_slot_attr}, "PUSH_SLOT_ATTR", 2},                // sattrnum slot
	{[2]instrImpl{push_glyph_attr_obs, push_glyph_attr_obs}, "PUSH_GLYPH_ATTR_OBS", 2}, // gattrnum slot
	{[2]instrImpl{push_glyph_metric, push_glyph_metric}, "PUSH_GLYPH_METRIC", 3},       // gmetric slot level
	{[2]instrImpl{push_feat, push_feat}, "PUSH_FEAT", 2},                               // featidx slot

	{[2]instrImpl{push_att_to_gattr_obs, push_att_to_gattr_obs}, "PUSH_ATT_TO_GATTR_OBS", 2},          // gattrnum slot
	{[2]instrImpl{push_att_to_glyph_metric, push_att_to_glyph_metric}, "PUSH_ATT_TO_GLYPH_METRIC", 3}, // gmetric slot level
	{[2]instrImpl{push_islot_attr, push_islot_attr}, "PUSH_ISLOT_ATTR", 3},                            // sattrnum slot attrid

	{[2]instrImpl{nil, nil}, "PUSH_IGLYPH_ATTR", 3},

	{[2]instrImpl{pop_ret, pop_ret}, "POP_RET", 0}, // 0x30
	{[2]instrImpl{ret_zero, ret_zero}, "RET_ZERO", 0},
	{[2]instrImpl{ret_true, ret_true}, "RET_TRUE", 0},

	{[2]instrImpl{iattr_set, nil}, "IATTR_SET", 2},                         // sattrnum attrid
	{[2]instrImpl{iattr_add, nil}, "IATTR_ADD", 2},                         // sattrnum attrid
	{[2]instrImpl{iattr_sub, nil}, "IATTR_SUB", 2},                         // sattrnum attrid
	{[2]instrImpl{push_proc_state, push_proc_state}, "PUSH_PROC_STATE", 1}, // dummy
	{[2]instrImpl{push_version, push_version}, "PUSH_VERSION", 0},
	{[2]instrImpl{put_subs, nil}, "PUT_SUBS", 5}, // slot input_class input_class output_class output_class
	{[2]instrImpl{nil, nil}, "PUT_SUBS2", 0},
	{[2]instrImpl{nil, nil}, "PUT_SUBS3", 0},
	{[2]instrImpl{put_glyph, nil}, "PUT_GLYPH", 2},                                              // output_class output_class
	{[2]instrImpl{push_glyph_attr, push_glyph_attr}, "PUSH_GLYPH_ATTR", 3},                      // gattrnum gattrnum slot
	{[2]instrImpl{push_att_to_glyph_attr, push_att_to_glyph_attr}, "PUSH_ATT_TO_GLYPH_ATTR", 3}, // gattrnum gattrnum slot
	{[2]instrImpl{bor, bor}, "BITOR", 0},
	{[2]instrImpl{band, band}, "BITAND", 0},
	{[2]instrImpl{bnot, bnot}, "BITNOT", 0}, // 0x40
	{[2]instrImpl{setbits, setbits}, "BITSET", 4},
	{[2]instrImpl{set_feat, nil}, "SET_FEAT", 2}, // featidx slot
	// private opcodes for internal use only, comes after all other on disk opcodes.
	{[2]instrImpl{temp_copy, nil}, "TEMP_COPY", 0},
}

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
//                      not implemented should also be marked nil in the
//                      opcodes tables this will cause the code class to spot
//                      them in a live code stream and throw a runtime_error
//                      instead.
//    push(n)           Push the value n onto the stack.
//    pop()             Pop the top most value and return it.
//

type regbank struct {
	is        *Slot
	smap      *slotMap
	map_      int // index of the current slot into smap.slots
	mapb      int
	ip        int
	direction bool
	flags     uint8
}

func (r *regbank) slotAt(index int8) *Slot {
	return r.smap.get(r.map_ + int(index))
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

func (st *stack) die(reg *regbank) ([]byte, bool) {
	reg.is = reg.smap.segment.last
	st.push(1)
	return nil, false
}

// Do nothing.
func nop(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	return args, st.top < stackMax
}

// Push the given 8-bit signed number onto the stack.
func push_byte(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.push(int32(int8(args[0])))
	return args[1:], st.top < stackMax
}

// Push the given 8-bit unsigned number onto the stack.
func push_byte_u(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.push(int32(args[0]))
	return args[1:], st.top < stackMax
}

// Treat the two arguments as a 16-bit signed number, with byte1 as the most significant.
// Push the number onto the stack.
func push_short(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	r := int16(uint16(args[0])<<8 | uint16(args[1]))
	st.push(int32(r))
	return args[2:], st.top < stackMax
}

// Treat the two arguments as a 16-bit unsigned number, with byte1 as the most significant.
// Push the number onto the stack.
func push_short_u(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	r := uint16(args[0])<<8 | uint16(args[1])
	st.push(int32(r))
	return args[2:], st.top < stackMax
}

// Treat the four arguments as a 32-bit number, with byte1 as the most significant. Push the
// number onto the stack.
func push_long(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	r := int32(args[0])<<24 | int32(args[1])<<16 | int32(args[2])<<8 | int32(args[3])
	st.push(r)
	return args[4:], st.top < stackMax
}

// Pop the top two items off the stack, add them, and push the result.
func add(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	v := st.pop()
	st.vals[st.top-1] += v
	return args, st.top < stackMax
}

// Pop the top two items off the stack, subtract the first (top-most) from the second, and
// push the result.
func sub(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	v := st.pop()
	st.vals[st.top-1] -= v
	return args, st.top < stackMax
}

// Pop the top two items off the stack, multiply them, and push the result.
func mul(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	v := st.pop()
	st.vals[st.top-1] *= v
	return args, st.top < stackMax
}

// Pop the top two items off the stack, divide the second by the first (top-most), and push
// the result.
func div_(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	b := st.pop()
	a := st.vals[st.top-1]
	if b == 0 || (a == math.MinInt32 && b == -1) {
		return st.die(reg)
	}
	st.vals[st.top-1] = a / b
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push the minimum.
func min_(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	b := st.vals[st.top-1]
	if a < b {
		st.vals[st.top-1] = a
	}
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push the maximum.
func max_(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	b := st.vals[st.top-1]
	if a > b {
		st.vals[st.top-1] = a
	}
	return args, st.top < stackMax
}

// Pop the top item off the stack and push the negation.
func neg(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.vals[st.top-1] = -st.vals[st.top-1]
	return args, st.top < stackMax
}

// Pop the top item off the stack and push the value truncated to 8 bits.
func trunc8(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.vals[st.top-1] = int32(uint8(st.vals[st.top-1]))
	return args, st.top < stackMax
}

// Pop the top item off the stack and push the value truncated to 16 bits.
func trunc16(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.vals[st.top-1] = int32(uint16(st.vals[st.top-1]))
	return args, st.top < stackMax
}

// Pop the top three items off the stack. If the first == 0 (false), push the third back on,
// otherwise push the second back on.
func cond(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	f := st.pop()
	t := st.pop()
	c := st.pop()
	if c != 0 {
		st.push(t)
	} else {
		st.push(f)
	}
	return args, st.top < stackMax
}

func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// Pop the top two items off the stack and push their logical and. Zero is treated as false; all
// other values are treated as true.
func and_(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop() != 0
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] != 0 && a)
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push their logical or. Zero is treated as false; all
// other values are treated as true.
func or_(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop() != 0
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] != 0 || a)
	return args, st.top < stackMax
}

// Pop the top item off the stack and push its logical negation (1 if it equals zero, 0
// otherwise.
func not_(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] == 0)
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push 1 if they are equal, 0 if not.
func equal(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] == a)
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push 0 if they are equal, 1 if not.
func not_eq_(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] != a)
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push 1 if the next-to-the-top is less than the top-
// most; push 0 othewise
func less(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] < a)
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push 1 if the next-to-the-top is greater than the
// top-most; push 0 otherwise.
func gtr(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] > a)
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push 1 if the next-to-the-top is less than or equal
// to the top-most; push 0 otherwise.
func less_eq(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] <= a)
	return args, st.top < stackMax
}

// Pop the top two items off the stack and push 1 if the next-to-the-top is greater than or
// equal to the top-most; push 0 otherwise
func gtr_eq(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = boolToInt(st.vals[st.top-1] >= a)
	return args, st.top < stackMax
}

// Move the current slot pointer forward one slot (used after we have finished processing
// that slot).
func next(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	if reg.map_ >= reg.smap.size {
		return st.die(reg)
	}
	if reg.is != nil {
		if reg.is == reg.smap.highwater {
			reg.smap.highpassed = true
		}
		reg.is = reg.is.Next
	}
	reg.map_++
	return args, st.top < stackMax
}

// //func next_n(reg *regbank, st *stack, args []byte) ([]byte, bool) {
// //    use_params(1);
// //    NOT_IMPLEMENTED;
//     //declare_params(1);
//     //const size_t num = uint8(*param);
// //ENDOP

// //func copy_next(reg *regbank, st *stack, args []byte) ([]byte, bool) {
// //     if (is) is = is.next;
// //     ++map;
// //ENDOP

// Determine the index of the glyph that was the input in the given slot within the input
// class, and place the corresponding glyph from the output class in the current slot. The slot number
// is relative to the current input position.
func put_subs(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slotRef := int8(args[0])

	inputClass := uint16(args[1])<<8 | uint16(args[2])
	outputClass := uint16(args[3])<<8 | uint16(args[4])
	slot := reg.slotAt(slotRef)
	seg := reg.smap.segment
	if slot != nil {
		index := seg.silf.classMap.findClassIndex(inputClass, slot.glyphID)
		reg.is.setGlyph(seg, seg.silf.classMap.getClassGlyph(outputClass, index))
	}
	return args[5:], st.top < stackMax
}

// #if 0
// func put_subs2(reg *regbank, st *stack, args []byte) ([]byte, bool) { // not implemented
//     NOT_IMPLEMENTED;
// return args, st.top < stackMax
// }

// func put_subs3(reg *regbank, st *stack, args []byte) ([]byte, bool) { // not implemented
//     NOT_IMPLEMENTED;
// return args, st.top < stackMax
// }
// #endif

// Put the first glyph of the specified class into the output. Normally used when there is only
// one member of the class, and when inserting.
func put_glyph(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	outputClass := uint16(args[0])<<8 | uint16(args[1])
	seg := reg.smap.segment
	reg.is.setGlyph(seg, seg.silf.classMap.getClassGlyph(outputClass, 0))
	return args[2:], st.top < stackMax
}

// Put the first glyph of the specified class into the output. Normally used when there is only
// one member of the class, and when inserting.
func put_glyph_8bit_obs(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	outputClass := args[0]
	seg := reg.smap.segment
	reg.is.setGlyph(seg, seg.silf.classMap.getClassGlyph(uint16(outputClass), 0))
	return args[1:], st.top < stackMax
}

// Determine the index of the glyph that was the input in the given slot within the input
// class, and place the corresponding glyph from the output class in the current slot. The slot number
// is relative to the current input position.
func put_subs_8bit_obs(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slotRef := int8(args[0])
	inputClass := args[1]
	outputClass := args[2]
	slot := reg.slotAt(slotRef)
	if slot != nil {
		seg := reg.smap.segment
		index := seg.silf.classMap.findClassIndex(uint16(inputClass), slot.glyphID)
		reg.is.setGlyph(seg, seg.silf.classMap.getClassGlyph(uint16(outputClass), index))
	}
	return args[3:], st.top < stackMax
}

// Copy the glyph that was in the input in the given slot into the current output slot. The slot
// number is relative to the current input position.
func put_copy(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slotRef := int8(args[0])
	is := reg.is
	if is != nil && !is.isDeleted() {
		ref := reg.slotAt(slotRef)
		if ref != nil && ref != is {
			tempUserAttrs := is.userAttrs
			if is.parent != nil || is.child != nil {
				return st.die(reg)
			}
			prev := is.prev
			next := is.Next

			copy(tempUserAttrs, ref.userAttrs)
			*is = *ref
			is.child = nil
			is.sibling = nil
			is.userAttrs = tempUserAttrs
			is.Next = next
			is.prev = prev
			if is.parent != nil {
				is.parent.child = is
			}
		}
		is.markCopied(false)
		is.markDeleted(false)
	}
	return args[1:], st.top < stackMax
}

// Insert a new slot before the current slot and make the new slot the current one.
func insert(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	if reg.smap.decMax() <= 0 {
		return st.die(reg)
	}
	seg := reg.smap.segment
	newSlot := seg.newSlot()
	if newSlot == nil {
		return st.die(reg)
	}
	iss := reg.is
	for iss != nil && iss.isDeleted() {
		iss = iss.Next
	}
	if iss == nil {
		if seg.last != nil {
			seg.last.Next = newSlot
			newSlot.prev = seg.last
			newSlot.Before = seg.last.Before
			seg.last = newSlot
		} else {
			seg.first = newSlot
			seg.last = newSlot
		}
	} else if iss.prev != nil {
		iss.prev.Next = newSlot
		newSlot.prev = iss.prev
		newSlot.Before = iss.prev.After
	} else {
		newSlot.prev = nil
		newSlot.Before = iss.Before
		seg.first = newSlot
	}
	newSlot.Next = iss
	if iss != nil {
		iss.prev = newSlot
		newSlot.original = iss.original
		newSlot.After = iss.Before
	} else if newSlot.prev != nil {
		newSlot.original = newSlot.prev.original
		newSlot.After = newSlot.prev.After
	} else {
		newSlot.original = seg.defaultOriginal
	}
	if reg.is == reg.smap.highwater {
		reg.smap.highpassed = false
	}
	reg.is = newSlot
	seg.numGlyphs += 1
	if reg.map_ != 0 {
		reg.map_--
	}
	return args, st.top < stackMax
}

// Delete the current item in the input stream.
func delete_(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	is := reg.is
	seg := reg.smap.segment
	if is == nil || is.isDeleted() {
		return st.die(reg)
	}
	is.markDeleted(true)
	if is.prev != nil {
		is.prev.Next = is.Next
	} else {
		seg.first = is.Next
	}

	if is.Next != nil {
		is.Next.prev = is.prev
	} else {
		seg.last = is.prev
	}

	if is == reg.smap.highwater {
		reg.smap.highwater = is.Next
	}
	if is.prev != nil {
		is = is.prev
	}
	seg.numGlyphs -= 1
	return args, st.top < stackMax
}

// Set the associations for the current slot to be the given slot(s) in the input. The first
// argument indicates how many slots follow. The slot offsets are relative to the current input slot.
func assoc(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	num := args[0]
	assocs := args[1 : num+1]
	max, min := -1, -1

	for _, sr := range assocs {
		ts := reg.slotAt(int8(sr))
		if ts != nil && (min == -1 || ts.Before < min) {
			min = ts.Before
		}
		if ts != nil && ts.After > max {
			max = ts.After
		}
	}
	if min > -1 { // implies max > -1
		reg.is.Before = min
		reg.is.After = max
	}
	return args[num+1:], st.top < stackMax
}

// If the slot currently being tested is not the slot specified by the <slot-offset> argument
// (relative to the stream position, the first modified item in the rule), skip the given number of bytes
// of stack-machine code. These bytes represent a test that is irrelevant for this slot.
func cntxt_item(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	// It turns out this is a cunningly disguised condition forward jump.
	// declare_params(3);
	is_arg := int8(args[0])
	iskip, dskip := args[1], args[2]
	args = args[3:]
	if reg.mapb+int(is_arg) != reg.map_ {
		reg.ip += int(iskip)
		args = args[dskip:]
		st.push(1)
	}
	return args, st.top < stackMax
}

// Pop the stack and set the value of the given attribute to the resulting numerical value.
func attr_set(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	val := st.pop()
	reg.is.setAttr(reg.smap, slat, 0, int16(val))
	return args[1:], st.top < stackMax
}

// Pop the stack and adjust the value of the given attribute by adding the popped value.
func attr_add(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	val := st.pop()
	smap := reg.smap
	seg := smap.segment
	if (slat == gr_slatPosX || slat == gr_slatPosY) && (reg.flags&POSITIONED) == 0 {
		seg.positionSlots(nil, smap.begin(), smap.endMinus1(), seg.currdir(), true)
		reg.flags |= POSITIONED
	}
	res := int32(reg.is.getAttr(seg, slat, 0))
	reg.is.setAttr(smap, slat, 0, int16(val+res))
	return args[1:], st.top < stackMax
}

// Pop the stack and adjust the value of the given attribute by subtracting the popped value.
func attr_sub(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	val := st.pop()
	smap := reg.smap
	seg := smap.segment
	if (slat == gr_slatPosX || slat == gr_slatPosY) && (reg.flags&POSITIONED) == 0 {
		seg.positionSlots(nil, smap.begin(), smap.endMinus1(), seg.currdir(), true)
		reg.flags |= POSITIONED
	}
	res := int32(reg.is.getAttr(seg, slat, 0))
	reg.is.setAttr(smap, slat, 0, int16(res-val))
	return args[1:], st.top < stackMax
}

// Pop the stack and set the given attribute to the value, which is a reference to another slot,
// making an adjustment for the stream position. The value is relative to the current stream position.
// [Note that corresponding add and subtract operations are not needed since it never makes sense to
// add slot references.]
func attr_set_slot(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])

	offset := int32(reg.map_-1) * boolToInt(slat == gr_slatAttTo)
	val := st.pop() + offset
	reg.is.setAttr(reg.smap, slat, int(offset), int16(val))
	return args[1:], st.top < stackMax
}

// Pop the stack and set the value of the given indexed attribute to the resulting numerical
// value. Not to be used for attributes whose value is a slot reference. [Currently the only non-slot-
// reference indexed slot attributes are userX.]
func iattr_set(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	idx := int(args[1])
	val := st.pop()
	reg.is.setAttr(reg.smap, slat, idx, int16(val))
	return args[2:], st.top < stackMax
}

// Pop the stack and adjust the value of the given indexed slot attribute by adding the
// popped value. Not to be used for attributes whose value is a slot reference. [Currently the only
// non-slot-reference indexed slot attributes are userX.]
func iattr_add(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	idx := int(args[1])
	val := st.pop()
	smap := reg.smap
	seg := smap.segment
	if (slat == gr_slatPosX || slat == gr_slatPosY) && (reg.flags&POSITIONED) == 0 {
		seg.positionSlots(nil, smap.begin(), smap.endMinus1(), seg.currdir(), true)
		reg.flags |= POSITIONED
	}
	res := reg.is.getAttr(seg, slat, idx)
	reg.is.setAttr(smap, slat, idx, int16(val+res))
	return args[2:], st.top < stackMax
}

// Pop the stack and adjust the value of the given indexed slot attribute by subtracting the
// popped value. Not to be used for attributes whose value is a slot reference. [Currently the only
// non-slot-reference indexed slot attributes are userX.]
func iattr_sub(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	idx := int(args[1])
	val := st.pop()
	smap := reg.smap
	seg := smap.segment
	if (slat == gr_slatPosX || slat == gr_slatPosY) && (reg.flags&POSITIONED) == 0 {
		seg.positionSlots(nil, smap.begin(), smap.endMinus1(), seg.currdir(), true)
		reg.flags |= POSITIONED
	}
	res := reg.is.getAttr(seg, slat, idx)
	reg.is.setAttr(smap, slat, idx, int16(res-val))
	return args[2:], st.top < stackMax
}

// Pop the stack and set the value of the given indexed attribute to the resulting numerical
// value. Not to be used for attributes whose value is a slot reference. [Currently the only non-slot-
// reference indexed slot attributes are userX.]
func iattr_set_slot(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	idx := args[1]
	val := int(st.pop() + int32(reg.map_-1)*boolToInt(slat == gr_slatAttTo))
	reg.is.setAttr(reg.smap, slat, int(idx), int16(val))
	return args[2:], st.top < stackMax
}

// Look up the value of the given slot attribute of the given slot and push the result on the
// stack. The slot offset is relative to the current input position.
func push_slot_attr(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	slat := attrCode(args[0])
	slotRef := int8(args[1])
	smap := reg.smap
	if (slat == gr_slatPosX || slat == gr_slatPosY) && (reg.flags&POSITIONED) == 0 {
		smap.segment.positionSlots(nil, smap.begin(), smap.endMinus1(), smap.segment.currdir(), true)
		reg.flags |= POSITIONED
	}
	slot := reg.slotAt(slotRef)
	if slot != nil {
		res := slot.getAttr(smap.segment, slat, 0)
		st.push(res)
	}
	return args[2:], st.top < stackMax
}

// Push the value of the indexed slot attribute onto the stack. [The current indexed slot
// attributes are component.X.ref and userX.]
func push_islot_attr(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	// declare_params(3);
	slat := attrCode(args[0])
	slotRef := int8(args[1])
	idx := int(args[2])
	smap := reg.smap
	seg := smap.segment
	if (slat == gr_slatPosX || slat == gr_slatPosY) && (reg.flags&POSITIONED) == 0 {
		seg.positionSlots(nil, smap.begin(), smap.endMinus1(), seg.currdir(), true)
		reg.flags |= POSITIONED
	}
	slot := reg.slotAt(slotRef)
	if slot != nil {
		res := slot.getAttr(seg, slat, idx)
		st.push(res)
	}
	return args[3:], st.top < stackMax
}

// Look up the value of the given glyph attribute of the given slot and push the result on the
// stack. The slot offset is relative to the current input position.
func push_glyph_attr_obs(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	glyphAttr := uint16(args[0])
	slotRef := int8(args[1])
	slot := reg.slotAt(slotRef)
	if slot != nil {
		st.push(int32(reg.smap.segment.face.getGlyphAttr(slot.glyphID, glyphAttr)))
	}
	return args[2:], st.top < stackMax
}

// Look up the value of the given glyph metric of the given slot and push the result on the
// stack. The slot offset is relative to the current input position. The level indicates the attachment
// level for cluster metrics.
func push_glyph_metric(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	glyphAttr := args[0]
	slotRef := int8(args[1])
	attrLevel := args[2]
	slot := reg.slotAt(slotRef)
	if slot != nil {
		st.push(reg.smap.segment.getGlyphMetric(slot, glyphAttr, attrLevel, reg.direction))
	}
	return args[3:], st.top < stackMax
}

// Push the value of the given feature for the current slot onto the stack.
func push_feat(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	feat := args[0]
	slotRef := int8(args[1])
	slot := reg.slotAt(slotRef)
	if slot != nil {
		st.push(reg.smap.segment.getFeature(feat))
	}
	return args[2:], st.top < stackMax
}

// Look up the value of the given glyph attribute for the slot indicated by the given slot’s
// attach.to attribute. Push the result on the stack.
func push_att_to_gattr_obs(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	glyphAttr := args[0]
	slotRef := int8(args[1])
	slot := reg.slotAt(slotRef)
	if slot != nil {
		if att := slot.parent; att != nil {
			slot = att
		}
		st.push(int32(reg.smap.segment.face.getGlyphAttr(slot.glyphID, uint16(glyphAttr))))
	}
	return args[2:], st.top < stackMax
}

// Look up the value of the given glyph metric for the slot indicated by the given slot’s
// attach.to attribute. Push the result on the stack.
func push_att_to_glyph_metric(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	glyphAttr := args[0]
	slotRef := int8(args[1])
	attrLevel := args[2]
	slot := reg.slotAt(slotRef)
	if slot != nil {
		if att := slot.parent; att != nil {
			slot = att
		}
		st.push(int32(reg.smap.segment.getGlyphMetric(slot, glyphAttr, attrLevel, reg.direction)))
	}
	return args[3:], st.top < stackMax
}

// #if 0
// func push_iglyph_attr(reg *regbank, st *stack, args []byte) ([]byte, bool) { // not implemented
//     NOT_IMPLEMENTED;
// return args, st.top < stackMax
// }
// #endif

func pop_ret(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	ret := st.pop()
	st.push(ret)
	return args, false
}

func ret_zero(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.push(0)
	return args, false
}

func ret_true(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.push(1)
	return args, false
}

func push_proc_state(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.push(1)
	return args[1:], st.top < stackMax
}

func push_version(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.push(0x00030000)
	return args, st.top < stackMax
}

// Look up the value of the given glyph attribute of the given slot and push the result on the
// stack. The slot offset is relative to the current input position.
func push_glyph_attr(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	glyphAttr := uint16(args[0])<<8 | uint16(args[1])
	slotRef := int8(args[2])
	slot := reg.slotAt(slotRef)
	if slot != nil {
		st.push(int32(reg.smap.segment.face.getGlyphAttr(slot.glyphID, glyphAttr)))
	}
	return args[3:], st.top < stackMax
}

// Look up the value of the given glyph attribute for the slot indicated by the given slot’s
// attach.to attribute. Push the result on the stack.
func push_att_to_glyph_attr(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	glyphAttr := uint16(args[0])<<8 | uint16(args[1])
	slotRef := int8(args[2])
	slot := reg.slotAt(slotRef)
	if slot != nil {
		if att := slot.parent; att != nil {
			slot = att
		}
		st.push(int32(reg.smap.segment.face.getGlyphAttr(slot.glyphID, glyphAttr)))
	}
	return args[3:], st.top < stackMax
}

func temp_copy(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	seg := reg.smap.segment
	newSlot := seg.newSlot()
	is := reg.is
	if newSlot == nil || is == nil {
		return st.die(reg)
	}
	tempUserAttrs := newSlot.userAttrs
	copy(tempUserAttrs, is.userAttrs)
	*newSlot = *is
	newSlot.userAttrs = tempUserAttrs
	newSlot.markCopied(true)
	reg.smap.slots[reg.map_] = newSlot
	return args, st.top < stackMax
}

func band(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = st.vals[st.top-1] & a
	return args, st.top < stackMax
}

func bor(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	a := st.pop()
	st.vals[st.top-1] = st.vals[st.top-1] | a
	return args, st.top < stackMax
}

func bnot(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	st.vals[st.top-1] = ^st.vals[st.top-1]
	return args, st.top < stackMax
}

func setbits(_ *regbank, st *stack, args []byte) ([]byte, bool) {
	m := int32(uint16(args[0])<<8 | uint16(args[1]))
	v := int32(uint16(args[2])<<8 | uint16(args[3]))
	st.vals[st.top-1] = (st.vals[st.top-1] & ^m) | v
	return args[4:], st.top < stackMax
}

func set_feat(reg *regbank, st *stack, args []byte) ([]byte, bool) {
	feat := args[0]
	slotRef := int8(args[1])
	slot := reg.slotAt(slotRef)
	if slot != nil {
		reg.smap.segment.setFeature(feat, int16(st.pop()))
	}
	return args[2:], st.top < stackMax
}
