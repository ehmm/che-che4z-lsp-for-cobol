package vm

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"sort"
)

// PathNode is a node in a persistent linked list representing the execution path.
type PathNode struct {
	Instruction Instruction
	Parent      *PathNode
}

func (p *PathNode) ToArray() []Instruction {
	var res []Instruction
	curr := p
	for curr != nil {
		res = append([]Instruction{curr.Instruction}, res...)
		curr = curr.Parent
	}
	return res
}

// VirtualMachineState represents a compact fingerprint of the VM state.
type VirtualMachineState struct {
	Hash uint64
}

func NewVirtualMachineState(redirectMap map[int]int, alterMap map[int]int, stickyMap map[string]int) *VirtualMachineState {
	h := fnv.New64a()
	// Use a buffer to build a deterministic string for hashing
	var buf bytes.Buffer

	// Hash Redirect Map
	buf.WriteString("v:")
	writeSortedMap(&buf, redirectMap)

	// Hash Alter Map
	buf.WriteString("|a:")
	writeSortedMap(&buf, alterMap)

	// Hash Sticky Map
	buf.WriteString("|s:")
	keys := make([]string, 0, len(stickyMap))
	for k := range stickyMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&buf, "%s:%d,", k, stickyMap[k])
	}

	h.Write(buf.Bytes())
	return &VirtualMachineState{Hash: h.Sum64()}
}

func writeSortedMap(buf *bytes.Buffer, m map[int]int) {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Fprintf(buf, "%d:%d,", k, m[k])
	}
}

// VirtualMachine represents a single execution thread.
type VirtualMachine struct {
	Context *VmContext
}

func (vm *VirtualMachine) Step() []*VirtualMachine {
	inst := vm.Context.ProgramListing.GetInstructionByPosition(vm.Context.IC)
	if inst == nil {
		return nil
	}

	nextPositions := inst.Execute(vm.Context)
	if len(nextPositions) == 0 {
		return nil
	}

	var forked []*VirtualMachine
	// If there are multiple paths, clone the VM for the additional ones
	for i := 1; i < len(nextPositions); i++ {
		newCtx := vm.Context.Clone()
		newCtx.IC = nextPositions[i]
		forked = append(forked, &VirtualMachine{Context: newCtx})
	}

	// Update current VM with the first path
	vm.Context.IC = nextPositions[0]
	return forked
}
