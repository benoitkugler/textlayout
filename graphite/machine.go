package graphite

const (
	machine_finished = iota
	machine_stack_underflow
	machine_stack_not_empty
	machine_stack_overflow
	machine_slot_offset_out_bounds
	machine_died_early
)
