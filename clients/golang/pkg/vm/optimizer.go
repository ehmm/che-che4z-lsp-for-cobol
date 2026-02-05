package vm

import (
	"github.com/code4z/ccf-cli/pkg/model"
)

// IbmOptimizer stops execution if a VM reaches the same instruction with the same state.
type IbmOptimizer struct {
	// stateMap maps IC (int) -> StateHash (uint64) -> exists (struct{})
	stateMap    map[int]map[uint64]struct{}
	TotalStates int
}

func NewIbmOptimizer() *IbmOptimizer {
	return &IbmOptimizer{
		stateMap: make(map[int]map[uint64]struct{}),
	}
}

// Apply checks if the current VM state has been seen before for the current instruction.
// Returns true if the VM should be stopped.
func (o *IbmOptimizer) Apply(vm *VirtualMachine) bool {
	inst := vm.Context.ProgramListing.GetInstructionByPosition(vm.Context.IC)
	if inst == nil {
		return false
	}

	// Skip optimization for branching/control instructions as per original logic
	switch inst.GetInitialNode().Type {
	case model.NodeTypeGoto, model.NodeTypePerform:
		return false
	}
	if _, ok := inst.(*VNCell); ok {
		return false
	}

	state := vm.Context.GenerateState()
	ic := vm.Context.IC

	states, ok := o.stateMap[ic]
	if !ok {
		states = make(map[uint64]struct{})
		o.stateMap[ic] = states
	}

	if _, seen := states[state.Hash]; seen {
		return true
	}

	states[state.Hash] = struct{}{}
	o.TotalStates++
	return false
}
