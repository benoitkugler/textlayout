package graphite

type passtype uint8 

const (
    PASS_TYPE_UNKNOWN passtype = iota
    PASS_TYPE_LINEBREAK
    PASS_TYPE_SUBSTITUTE
    PASS_TYPE_POSITIONING
    PASS_TYPE_JUSTIFICATION
)

type codeStatus uint8 

const (
	loaded codeStatus = iota
	alloc_failed
	invalid_opcode
	unimplemented_opcode_used
	out_of_range_data
	jump_past_end
	arguments_exhausted
	missing_return
	nested_context_item
	underfull_stack
)

// represents loaded graphite stack machine code
type code struct{
	// instr *     _code;
    // byte  *     _data;
    // size_t      _data_size,
    //             _instr_count;
    // byte        _max_ref;
    status codeStatus
    constraint bool
    //             _modify,
    //             _delete;
    // mutable bool _own;
}


func newCode(is_constraint bool, bytecode []byte,
	pre_context uint8,  rule_length uint16, silf *silfSubtable, face *graphiteFace,
	 pt passtype)  {
	var out code 
	out.status =loaded
	out.constraint = is_constraint

if len(bytecode) == 0{
return out
}
const opcode_t *    op_to_fn = Machine::getOpcodeTable();

// Allocate code and data target buffers, these sizes are a worst case
// estimate.  Once we know their real sizes the we'll shrink them.
if (_out)   _code = reinterpret_cast<instr *>(*_out);
else        _code = static_cast<instr *>(malloc(estimateCodeDataOut(bytecode_end-bytecode_begin, 1, is_constraint ? 0 : rule_length)));
_data = reinterpret_cast<byte *>(_code + (bytecode_end - bytecode_begin));

if (!_code || !_data) {
 failure(alloc_failed);
 return;
}

decoder::limits lims = {
 bytecode_end,
 pre_context,
 rule_length,
 silf.numClasses(),
 face.glyphs().numAttrs(),
 face.numFeatures(),
 {1,1,1,1,1,1,1,1,
  1,1,1,1,1,1,1,255,
  1,1,1,1,1,1,1,1,
  1,1,1,1,1,1,0,0,
  0,0,0,0,0,0,0,0,
  0,0,0,0,0,0,0,0,
  0,0,0,0,0,0,0, silf.numUser()}
};

decoder dec(lims, *this, pt);
if(!dec.load(bytecode_begin, bytecode_end))
return;

// Is this an empty program?
if (_instr_count == 0)
{
release_buffers();
::new (this) Code();
return;
}

// When we reach the end check we've terminated it correctly
if (!is_return(_code[_instr_count-1])) {
 failure(missing_return);
 return;
}

assert((_constraint && immutable()) || !_constraint);
dec.apply_analysis(_code, _code + _instr_count);
_max_ref = dec.max_ref();

// Now we know exactly how much code and data the program really needs
// realloc the buffers to exactly the right size so we don't waste any
// memory.
assert((bytecode_end - bytecode_begin) >= ptrdiff_t(_instr_count));
assert((bytecode_end - bytecode_begin) >= ptrdiff_t(_data_size));
memmove(_code + (_instr_count+1), _data, _data_size*sizeof(byte));
size_t const total_sz = ((_instr_count+1) + (_data_size + sizeof(instr)-1)/sizeof(instr))*sizeof(instr);
if (_out)
 *_out += total_sz;
else
{
instr * const old_code = _code;
_code = static_cast<instr *>(realloc(_code, total_sz));
if (!_code) free(old_code);
}
_data = reinterpret_cast<byte *>(_code + (_instr_count+1));

if (!_code)
{
 failure(alloc_failed);
 return;
}

// Make this RET_ZERO, we should never reach this but just in case ...
_code[_instr_count] = op_to_fn[RET_ZERO].impl[_constraint];

#ifdef GRAPHITE2_TELEMETRY
telemetry::count_bytes(_data_size + (_instr_count+1)*sizeof(instr));
#endif
}