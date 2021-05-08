package graphite

import (
	"fmt"
	"sort"
)

type machineStatus uint8

const (
	machine_stack_underflow machineStatus = iota
	machine_stack_not_empty
	machine_stack_overflow
	machine_slot_offset_out_bounds
	machine_died_early
)

func (m machineStatus) Error() string {
	switch m {
	case machine_stack_underflow:
		return "stack_underflow"
	case machine_stack_not_empty:
		return "stack_not_empty"
	case machine_stack_overflow:
		return "stack_overflow"
	case machine_slot_offset_out_bounds:
		return "slot_offset_out_bounds"
	case machine_died_early:
		return "died_early"
	}
	return "<unknown machine status>"
}

const (
	STACK_GUARD = 2
	stackMax    = 1 << 10
)

type stack struct {
	vals [stackMax + 2*STACK_GUARD]int32
	top  int // the top of the stack is at vals[top-1]
}

type machine struct {
	map_  slotMap
	stack stack
}

func newMachine(map_ slotMap) *machine {
	return &machine{
		map_: map_,
	}
}

func (m *machine) run(co *code) (int32, error) {
	if L := co.maxRef + int(m.map_.preContext); m.map_.size <= L || m.map_.get(L) == nil {
		return 1, machine_slot_offset_out_bounds
	}

	// Declare virtual machine registers
	reg := regbank{
		is:        m.map_.slots[0],
		smap:      &m.map_,
		mapb:      1 + int(m.map_.preContext),
		ip:        0,
		direction: m.map_.dir,
		flags:     0,
	}

	// Run the program
	program, args := co.instrs, co.args
	var ok bool
	for ; reg.ip < len(program); reg.ip++ {
		args, ok = program[reg.ip](&reg, &m.stack, args)

		if debugMode >= 2 {
			fmt.Println("FSM: ", co.opCodes[reg.ip], m.stack.vals[:m.stack.top])
		}

		if !ok {
			if co.opCodes[reg.ip].isReturn() {
				break
			} else {
				return 0, machine_died_early
			}
		}
	}

	if err := m.checkFinalStack(); err != nil {
		return 0, err
	}

	var ret int32
	if m.stack.top > 0 {
		ret = m.stack.pop()
	}
	return ret, nil
}

func (m *machine) checkFinalStack() error {
	if m.stack.top < 1 {
		return machine_stack_underflow // This should be impossible now.
	} else if m.stack.top >= stackMax {
		return machine_stack_overflow // So should this.
	} else if m.stack.top != 1 {
		return machine_stack_not_empty
	}
	return nil
}

// may mutate s1 if it has enough capacity
func mergeSortedRuleNumbers(s1, s2 []uint16) []uint16 {
	// TODO: this clearly not the fastest way to do it ..
	s1 = append(s1, s2...)
	sort.Slice(s1, func(i, j int) bool { return s1[i] < s2[i] })
	return s1
}

type FiniteStateMachine struct {
	rules []uint16 // indexes in rule list
	slots slotMap
}

func (fsm *FiniteStateMachine) accumulateRules(rules []uint16) {
	fsm.rules = mergeSortedRuleNumbers(fsm.rules, rules)
}

func (fsm *FiniteStateMachine) reset(slot *Slot, maxPreCtxt uint16) {
	fsm.rules = fsm.rules[:0]
	var ctxt uint16
	for ; ctxt != maxPreCtxt && slot.prev != nil; ctxt, slot = ctxt+1, slot.prev {
	}
	fsm.slots.reset(slot, ctxt)
}
