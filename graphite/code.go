package graphite

/** Used for looking up slot attributes. Most are already available in other functions **/
type attrCode uint8

const (
	/// adjusted glyph advance in x direction in design units
	gr_slatAdvX attrCode = iota
	/// adjusted glyph advance in y direction (usually 0) in design units
	gr_slatAdvY
	/// returns 0. Deprecated.
	gr_slatAttTo
	/// This slot attaches to its parent at the given design units in the x direction
	gr_slatAttX
	/// This slot attaches to its parent at the given design units in the y direction
	gr_slatAttY
	/// This slot attaches to its parent at the given glyph point (not implemented)
	gr_slatAttGpt
	/// x-direction adjustment from the given glyph point (not implemented)
	gr_slatAttXOff
	/// y-direction adjustment from the given glyph point (not implemented)
	gr_slatAttYOff
	/// Where on this glyph should align with the attachment point on the parent glyph in the x-direction.
	gr_slatAttWithX
	/// Where on this glyph should align with the attachment point on the parent glyph in the y-direction
	gr_slatAttWithY
	/// Which glyph point on this glyph should align with the attachment point on the parent glyph (not implemented).
	gr_slatWithGpt
	/// Adjustment to gr_slatWithGpt in x-direction (not implemented)
	gr_slatAttWithXOff
	/// Adjustment to gr_slatWithGpt in y-direction (not implemented)
	gr_slatAttWithYOff
	/// Attach at given nesting level (not implemented)
	gr_slatAttLevel
	/// Line break breakweight for this glyph
	gr_slatBreak
	/// Ligature component reference (not implemented)
	gr_slatCompRef
	/// bidi directionality of this glyph (not implemented)
	gr_slatDir
	/// Whether insertion is allowed before this glyph
	gr_slatInsert
	/// Final positioned position of this glyph relative to its parent in x-direction in pixels
	gr_slatPosX
	/// Final positioned position of this glyph relative to its parent in y-direction in pixels
	gr_slatPosY
	/// Amount to shift glyph by in x-direction design units
	gr_slatShiftX
	/// Amount to shift glyph by in y-direction design units
	gr_slatShiftY
	/// attribute user1
	gr_slatUserDefnV1
	/// not implemented
	gr_slatMeasureSol
	/// not implemented
	gr_slatMeasureEol
	/// Amount this slot can stretch (not implemented)
	gr_slatJStretch
	/// Amount this slot can shrink (not implemented)
	gr_slatJShrink
	/// Granularity by which this slot can stretch or shrink (not implemented)
	gr_slatJStep
	/// Justification weight for this glyph (not implemented)
	gr_slatJWeight
	/// Amount this slot mush shrink or stretch in design units
	gr_slatJWidth
	/// SubSegment split point
	gr_slatSegSplit = gr_slatJStretch + 29
	/// User defined attribute, see subattr for user attr number
	gr_slatUserDefn = 55
	/// Bidi level
	gr_slatBidiLevel = 24 + iota
	/// Collision flags
	gr_slatColFlags
	/// Collision constraint rectangle left (bl.x)
	gr_slatColLimitblx
	/// Collision constraint rectangle lower (bl.y)
	gr_slatColLimitbly
	/// Collision constraint rectangle right (tr.x)
	gr_slatColLimittrx
	/// Collision constraint rectangle upper (tr.y)
	gr_slatColLimittry
	/// Collision shift x
	gr_slatColShiftx
	/// Collision shift y
	gr_slatColShifty
	/// Collision margin
	gr_slatColMargin
	/// Margin cost weight
	gr_slatColMarginWt
	// Additional glyph that excludes movement near this one:
	gr_slatColExclGlyph
	gr_slatColExclOffx
	gr_slatColExclOffy
	// Collision sequence enforcing attributes:
	gr_slatSeqClass
	gr_slatSeqProxClass
	gr_slatSeqOrder
	gr_slatSeqAboveXoff
	gr_slatSeqAboveWt
	gr_slatSeqBelowXlim
	gr_slatSeqBelowWt
	gr_slatSeqValignHt
	gr_slatSeqValignWt

	/// not implemented
	gr_slatMax
	/// not implemented
	gr_slatNoEffect = gr_slatMax + 1
)

type opcode uint8

const (
	NOP = iota

	PUSH_BYTE
	PUSH_BYTEU
	PUSH_SHORT
	PUSH_SHORTU
	PUSH_LONG

	ADD
	SUB
	MUL
	DIV

	MIN_
	MAX_

	NEG

	TRUNC8
	TRUNC16

	COND

	AND
	OR
	NOT

	EQUAL
	NOT_EQ

	LESS
	GTR
	LESS_EQ
	GTR_EQ

	NEXT
	NEXT_N
	COPY_NEXT

	PUT_GLYPH_8BIT_OBS
	PUT_SUBS_8BIT_OBS
	PUT_COPY

	INSERT
	DELETE

	ASSOC

	CNTXT_ITEM

	ATTR_SET
	ATTR_ADD
	ATTR_SUB

	ATTR_SET_SLOT

	IATTR_SET_SLOT

	PUSH_SLOT_ATTR
	PUSH_GLYPH_ATTR_OBS

	PUSH_GLYPH_METRIC
	PUSH_FEAT

	PUSH_ATT_TO_GATTR_OBS
	PUSH_ATT_TO_GLYPH_METRIC

	PUSH_ISLOT_ATTR

	PUSH_IGLYPH_ATTR // not implemented

	POP_RET
	RET_ZERO
	RET_TRUE

	IATTR_SET
	IATTR_ADD
	IATTR_SUB

	PUSH_PROC_STATE
	PUSH_VERSION

	PUT_SUBS
	PUT_SUBS2
	PUT_SUBS3

	PUT_GLYPH
	PUSH_GLYPH_ATTR
	PUSH_ATT_TO_GLYPH_ATTR

	BITOR
	BITAND
	BITNOT

	BITSET
	SET_FEAT

	MAX_OPCODE
	// private opcodes for internal use only, comes after all other on disk opcodes
	TEMP_COPY = MAX_OPCODE
)

func (opc opcode) isReturn() bool {
	return opc == POP_RET || opc == RET_ZERO || opc == RET_TRUE
}

type passtype uint8

const (
	PASS_TYPE_UNKNOWN passtype = iota
	PASS_TYPE_LINEBREAK
	PASS_TYPE_SUBSTITUTE
	PASS_TYPE_POSITIONING
	PASS_TYPE_JUSTIFICATION
)

type errorStatusCode uint8

const (
	invalidOpCode errorStatusCode = iota
	unimplementedOpCodeUsed
	outOfRangeData
	jumpPastEnd
	argumentsExhausted
	missingReturn
	nestedContextItem
	underfullStack
)

func (c errorStatusCode) Error() string {
	switch c {
	case invalidOpCode:
		return "invalid opcode"
	case unimplementedOpCodeUsed:
		return "unimplemented opcode used"
	case outOfRangeData:
		return "out of range data"
	case jumpPastEnd:
		return "jump past end"
	case argumentsExhausted:
		return "arguments exhausted"
	case missingReturn:
		return "missing return"
	case nestedContextItem:
		return "nested context item"
	case underfullStack:
		return "underfull stack"
	}
	return "<unknown error code>"
}

// represents loaded graphite stack machine code
type code struct {
	instrs []instrImpl
	args   []byte // concatenated arguments for `instrs`

	// instr *     _code;
	// byte  *     _data;
	// size_t      _data_size,
	// instrCount int
	// byte        dec.max_ref;
	// status                     codeStatus
	constraint, delete, modify bool
	// mutable bool _own;
}

// newCode decodes an input and returns the loaded instructions
// the errors returns are of type errorStatusCode
func newCode(isConstraint bool, bytecode []byte,
	preContext uint8, ruleLength uint16, silf *silfSubtable, face *graphiteFace,
	pt passtype) (*code, error) {
	var out code

	out.constraint = isConstraint

	if len(bytecode) == 0 {
		return &out, nil
	}

	lims := decoderLimits{
		preContext: uint16(preContext),
		ruleLength: ruleLength,
		classes:    silf.classMap.numClasses(),
		glyfAttrs:  face.numAttributes,
		features:   uint16(len(face.feat)),
		attrid: [gr_slatMax]byte{
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 255,
			1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, silf.NumUserDefn, // 0, 0, etc...
		},
	}

	dec := newDecoder(&out, lims, pt)
	lastOpcode, err := dec.load(bytecode)
	if err != nil {
		return nil, err
	}

	// Is this an empty program?
	if len(out.instrs) == 0 {
		return &out, nil
	}

	// When we reach the end check we've terminated it correctly
	if !lastOpcode.isReturn() {
		return nil, err
	}

	// assert((_constraint && immutable()) || !_constraint)
	dec.code.instrs = dec.applyAnalysis(dec.code.instrs)

	// Make this RET_ZERO, we should never reach this but just in case ...
	dec.code.instrs = append(dec.code.instrs, opcode_table[RET_ZERO].impl[boolToInt(out.constraint)])

	return dec.code, nil
}

const decoderNUMCONTEXTS = 256

type decoderLimits struct {
	preContext                   uint16
	ruleLength                   uint16
	classes, glyfAttrs, features uint16
	attrid                       [gr_slatMax]byte
}

type context struct {
	codeRef             uint8
	changed, referenced bool
}

// responsible for reading a sequence of instructions
type decoder struct {
	code *code // resulting loaded code

	stackDepth          int // the number of element needed in stack
	outIndex, outLength int
	slotRef             int
	maxRef              int
	inCtxtItem          bool
	passtype            passtype

	contexts [decoderNUMCONTEXTS]context

	max decoderLimits
}

func newDecoder(code *code, max decoderLimits, pt passtype) *decoder {
	out := decoder{code: code, max: max, passtype: pt}
	out.outLength = 1
	if !code.constraint {
		out.outIndex = int(max.preContext)
		out.outLength = int(max.ruleLength)
	}
	return &out
}

// parses the input byte code and checks that operations
// and arguments are valid
// note that the instructions are NOT executed
// the last opcode is returned
func (dec *decoder) load(bytecode []byte) (opcode, error) {
	var lastOpcode opcode
	for len(bytecode) != 0 {
		opc, err := dec.fetchOpcode(bytecode)
		if err != nil {
			return 0, err
		}

		bytecode = bytecode[1:]
		dec.analyseOpcode(opc, bytecode)

		bytecode, err = dec.emitOpcode(opc, bytecode)
		if err != nil {
			return 0, err
		}

		lastOpcode = opc
	}

	return lastOpcode, nil
}

func (dec *decoder) validateOpcode(opc opcode, bc []byte) error {
	if opc >= MAX_OPCODE {
		return invalidOpCode
	}
	op := opcode_table[opc]
	if op.impl[boolToInt(dec.code.constraint)] == nil {
		return unimplementedOpCodeUsed
	}
	if op.paramSize == VARARGS && len(bc) == 0 {
		return argumentsExhausted
	}
	paramSize := op.paramSize
	if op.paramSize == VARARGS { // read the number of additional args as first arg
		paramSize = bc[0] + 1
	}
	if len(bc) < int(paramSize) {
		return argumentsExhausted
	}
	return nil
}

// bc is not empty
func (dec *decoder) fetchOpcode(bc []byte) (opcode, error) {
	opc := opcode(bc[0])
	bc = bc[1:]

	// Do some basic sanity checks based on what we know about the opcode
	if err := dec.validateOpcode(opc, bc); err != nil {
		return 0, err
	}

	// And check its arguments as far as possible
	switch opc {
	case NOP:
	case PUSH_BYTE, PUSH_BYTEU, PUSH_SHORT, PUSH_SHORTU, PUSH_LONG:
		dec.stackDepth++
	case ADD, SUB, MUL, DIV, MIN_, MAX_, AND, OR,
		EQUAL, NOT_EQ, LESS, GTR, LESS_EQ, GTR_EQ, BITOR, BITAND:
		dec.stackDepth--
		if dec.stackDepth <= 0 {
			return 0, underfullStack
		}
	case NEG, TRUNC8, TRUNC16, NOT, BITNOT, BITSET:
		if dec.stackDepth <= 0 {
			return 0, underfullStack
		}
	case COND:
		dec.stackDepth -= 2
		if dec.stackDepth <= 0 {
			return 0, underfullStack
		}
	case NEXT_N:
	// runtime checked
	case NEXT, COPY_NEXT:
		dec.outIndex++
		if dec.outIndex < -1 || dec.outIndex > dec.outLength || dec.slotRef > int(dec.max.ruleLength) {
			return 0, outOfRangeData
		}
	case PUT_GLYPH_8BIT_OBS:
		if err := validUpto(dec.max.classes, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case PUT_SUBS_8BIT_OBS:
		if err := dec.testRef(bc[0]); err != nil {
			return 0, err
		}
		if err := validUpto(dec.max.classes, uint16(bc[1])); err != nil {
			return 0, err
		}
		if err := validUpto(dec.max.classes, uint16(bc[2])); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case PUT_COPY:
		if err := dec.testRef(bc[0]); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case INSERT:
		if dec.passtype >= PASS_TYPE_POSITIONING {
			return 0, invalidOpCode
		}
		dec.outLength++
		if dec.outIndex < 0 {
			dec.outIndex++
		}
		if dec.outIndex < -1 || dec.outIndex >= dec.outLength {
			return 0, outOfRangeData
		}
	case DELETE:
		if dec.passtype >= PASS_TYPE_POSITIONING {
			return 0, invalidOpCode
		}
		if dec.outIndex < int(dec.max.preContext) {
			return 0, outOfRangeData
		}
		dec.outIndex--
		dec.outLength--
		if dec.outIndex < -1 || dec.outIndex > dec.outLength {
			return 0, outOfRangeData
		}
	case ASSOC:
		if bc[0] == 0 {
			return 0, outOfRangeData
		}
		for num := bc[0]; num != 0; num-- {
			if err := dec.testRef(bc[num]); err != nil {
				return 0, err
			}
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case CNTXT_ITEM:
		validUpto(dec.max.ruleLength, uint16(int(dec.max.preContext)+int(int8(bc[0]))))
		if len(bc) < 2+int(bc[1]) {
			return 0, jumpPastEnd
		}
		if dec.inCtxtItem {
			return 0, nestedContextItem
		}
	case ATTR_SET, ATTR_ADD, ATTR_SUB, ATTR_SET_SLOT:
		dec.stackDepth--
		if dec.stackDepth < 0 {
			return 0, underfullStack
		}
		validUpto(gr_slatMax, uint16(bc[0]))
		if attrCode(bc[0]) == gr_slatUserDefn { // use IATTR for user attributes
			return 0, outOfRangeData
		}
		if err := dec.testAttr(attrCode(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case IATTR_SET_SLOT:
		dec.stackDepth--
		if dec.stackDepth < 0 {
			return 0, underfullStack
		}
		if err := validUpto(gr_slatMax, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := validUpto(uint16(dec.max.attrid[bc[0]]), uint16(bc[1])); err != nil {
			return 0, err
		}
		if err := dec.testAttr(attrCode(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case PUSH_SLOT_ATTR:
		dec.stackDepth++
		if err := validUpto(gr_slatMax, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testRef(bc[1]); err != nil {
			return 0, err
		}
		if attrCode(bc[0]) == gr_slatUserDefn { // use IATTR for user attributes
			return 0, outOfRangeData
		}
		if err := dec.testAttr(attrCode(bc[0])); err != nil {
			return 0, err
		}
	case PUSH_GLYPH_ATTR_OBS, PUSH_ATT_TO_GATTR_OBS:
		dec.stackDepth++
		if err := validUpto(dec.max.glyfAttrs, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testRef(bc[1]); err != nil {
			return 0, err
		}
	case PUSH_ATT_TO_GLYPH_METRIC, PUSH_GLYPH_METRIC:
		dec.stackDepth++
		if err := validUpto(kgmetDescent, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testRef(bc[1]); err != nil {
			return 0, err
		}
	// level: dp[2] no check necessary
	case PUSH_FEAT:
		dec.stackDepth++
		if err := validUpto(dec.max.features, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testRef(bc[1]); err != nil {
			return 0, err
		}
	case PUSH_ISLOT_ATTR:
		dec.stackDepth++
		if err := validUpto(gr_slatMax, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testRef(bc[1]); err != nil {
			return 0, err
		}
		if err := validUpto(uint16(dec.max.attrid[bc[0]]), uint16(bc[2])); err != nil {
			return 0, err
		}
		if err := dec.testAttr(attrCode(bc[0])); err != nil {
			return 0, err
		}
	case PUSH_IGLYPH_ATTR:
		// not implemented
		dec.stackDepth++
	case POP_RET:
		dec.stackDepth--
		if dec.stackDepth < 0 {
			return 0, underfullStack
		}
		fallthrough
	case RET_ZERO, RET_TRUE:
	case IATTR_SET, IATTR_ADD, IATTR_SUB:
		dec.stackDepth--
		if dec.stackDepth < 0 {
			return 0, underfullStack
		}
		if err := validUpto(gr_slatMax, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := validUpto(uint16(dec.max.attrid[bc[0]]), uint16(bc[1])); err != nil {
			return 0, err
		}
		if err := dec.testAttr(attrCode(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case PUSH_PROC_STATE, PUSH_VERSION:
		dec.stackDepth++
	case PUT_SUBS:
		if err := dec.testRef(bc[0]); err != nil {
			return 0, err
		}
		if err := validUpto(dec.max.classes, uint16(bc[1])<<8|uint16(bc[2])); err != nil {
			return 0, err
		}
		if err := validUpto(dec.max.classes, uint16(bc[3])<<8|uint16(bc[4])); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case PUT_SUBS2, PUT_SUBS3:
	// not implemented
	case PUT_GLYPH:
		if err := validUpto(dec.max.classes, uint16(bc[0])<<8|uint16(bc[1])); err != nil {
			return 0, err
		}
		if err := dec.testContext(); err != nil {
			return 0, err
		}
	case PUSH_GLYPH_ATTR, PUSH_ATT_TO_GLYPH_ATTR:
		dec.stackDepth++
		if err := validUpto(dec.max.glyfAttrs, uint16(bc[0])<<8|uint16(bc[1])); err != nil {
			return 0, err
		}
		if err := dec.testRef(bc[2]); err != nil {
			return 0, err
		}
	case SET_FEAT:
		if err := validUpto(dec.max.features, uint16(bc[0])); err != nil {
			return 0, err
		}
		if err := dec.testRef(bc[1]); err != nil {
			return 0, err
		}
	default:
		return 0, invalidOpCode
	}

	return opc, nil
}

func validUpto(limit, x uint16) error {
	if (limit != 0) && (x < limit) {
		return nil
	}
	return outOfRangeData
}

func (dec *decoder) testContext() error {
	if dec.outIndex >= dec.outLength || dec.outIndex < 0 || dec.slotRef >= decoderNUMCONTEXTS-1 {
		return outOfRangeData
	}
	return nil
}

func (dec *decoder) testRef(index_ byte) error {
	index := int8(index_)
	if dec.code.constraint && !dec.inCtxtItem {
		if index > 0 || uint16(-index) > dec.max.preContext {
			return outOfRangeData
		}
	} else {
		if L := dec.slotRef + int(dec.max.preContext) + int(index); dec.max.ruleLength == 0 ||
			L >= int(dec.max.ruleLength) || L < 0 {
			return outOfRangeData
		}
	}
	return nil
}

func (dec *decoder) testAttr(attr attrCode) error {
	if dec.passtype < PASS_TYPE_POSITIONING {
		if attr != gr_slatBreak && attr != gr_slatDir && attr != gr_slatUserDefn && attr != gr_slatCompRef {
			return outOfRangeData
		}
	}
	return nil
}

// the length of arg as been checked
func (dec *decoder) analyseOpcode(opc opcode, arg []byte) {
	switch opc {
	case DELETE:
		dec.code.delete = true
	case ASSOC:
		dec.setChanged(0)
		//      for (uint8 num = arg[0]; num; --num)
		//        _analysis.setNoref(num);
	case PUT_GLYPH_8BIT_OBS, PUT_GLYPH:
		dec.code.modify = true
		dec.setChanged(0)
	case ATTR_SET, ATTR_ADD, ATTR_SUB, ATTR_SET_SLOT, IATTR_SET_SLOT, IATTR_SET, IATTR_ADD, IATTR_SUB:
		dec.setNoref(0)
	case NEXT, COPY_NEXT:
		dec.slotRef++
		dec.contexts[dec.slotRef] = context{codeRef: uint8(len(dec.code.instrs) + 1)}
		// if (_analysis.slotRef > _analysis.max_ref) _analysis.max_ref = _analysis.slotRef;
	case INSERT:
		if dec.slotRef >= 0 {
			dec.slotRef--
		}
		dec.code.modify = true
	case PUT_SUBS_8BIT_OBS /* slotRef on 1st parameter */, PUT_SUBS:
		dec.code.modify = true
		dec.setChanged(0)
		fallthrough
	case PUT_COPY:
		if arg[0] != 0 {
			dec.setChanged(0)
			dec.code.modify = true
		}
		dec.setRef(arg[0])
	case PUSH_GLYPH_ATTR_OBS, PUSH_SLOT_ATTR, PUSH_GLYPH_METRIC, PUSH_ATT_TO_GATTR_OBS, PUSH_ATT_TO_GLYPH_METRIC, PUSH_ISLOT_ATTR, PUSH_FEAT, SET_FEAT:
		dec.setRef(arg[1])
	case PUSH_ATT_TO_GLYPH_ATTR, PUSH_GLYPH_ATTR:
		dec.setRef(arg[2])
	}
}

func (dec *decoder) setRef(arg byte) {
	index := int(arg)
	if index+dec.slotRef < 0 || index+dec.slotRef >= decoderNUMCONTEXTS {
		return
	}
	dec.contexts[index+dec.slotRef].referenced = true
	if index+dec.slotRef > dec.maxRef {
		dec.maxRef = index + dec.slotRef
	}
}

func (dec *decoder) setNoref(index int) {
	if index+dec.slotRef < 0 || index+dec.slotRef >= decoderNUMCONTEXTS {
		return
	}
	if index+dec.slotRef > dec.maxRef {
		dec.maxRef = index + dec.slotRef
	}
}

func (dec *decoder) setChanged(index int) {
	if index+dec.slotRef < 0 || index+dec.slotRef >= decoderNUMCONTEXTS {
		return
	}
	dec.contexts[index+dec.slotRef].changed = true
	if index+dec.slotRef > dec.maxRef {
		dec.maxRef = index + dec.slotRef
	}
}

type instrImpl = func(st *stack, data []byte) bool

// length of bc has been checked
// the `code` item will be updated, and the remaining bytecode
// input is returned
func (dec *decoder) emitOpcode(opc opcode, bc []byte) ([]byte, error) {
	op := opcode_table[opc]
	instr := op.impl[boolToInt(dec.code.constraint)]
	if instr == nil {
		return nil, unimplementedOpCodeUsed
	}

	paramSize := op.paramSize
	if op.paramSize == VARARGS {
		paramSize = bc[0] + 1
	}

	// Add this instruction
	dec.code.instrs = append(dec.code.instrs, instr)

	// Grab the parameters
	if paramSize != 0 {
		dec.code.args = append(dec.code.args, bc[:paramSize]...)
		bc = bc[paramSize:]
	}

	// recursively decode a context item so we can split the skip into
	// instruction and data portions.
	if opc == CNTXT_ITEM {
		// assert(_out_index == 0);
		dec.inCtxtItem = true
		dec.slotRef = int(int8(dec.code.args[len(dec.code.args)-2]))
		dec.outIndex = int(dec.max.preContext) + dec.slotRef
		dec.outLength = int(dec.max.ruleLength)

		instrSkip := dec.code.args[len(dec.code.args)-1]

		_, err := dec.load(bc[:instrSkip])
		if err != nil {
			return nil, err
		}
		bc = bc[instrSkip:]

		dec.outLength = 1
		dec.outIndex = 0
		dec.slotRef = 0
		dec.inCtxtItem = false
	}

	return bc, nil
}

func (dec *decoder) applyAnalysis(code []instrImpl) []instrImpl {
	// insert TEMP_COPY commands for slots that need them (that change and are referenced later)
	tempcount := 0
	if dec.code.constraint {
		return code
	}

	temp_copy := opcode_table[TEMP_COPY].impl[0]
	for _, c := range dec.contexts[:dec.slotRef] {
		if !c.referenced || !c.changed {
			continue
		}

		code = append(code, nil)
		tip := code[int(c.codeRef)+tempcount:]
		copy(tip[1:], tip)
		tip[0] = temp_copy
		dec.code.delete = true
	}

	return code
}
