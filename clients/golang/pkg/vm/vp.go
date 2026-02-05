package vm

import (
	"fmt"
	"runtime"
	"time"

	"github.com/code4z/ccf-cli/pkg/model"
)

// VirtualProcessor manages the execution of multiple Virtual Machines.
type VirtualProcessor struct {
	vms            []*VirtualMachine
	optimizer      *IbmOptimizer
	maxVmCount     int
	totalSteps     int
	programListing *ProgramListing
	graphWriter    *model.GraphWriter
}

func NewVirtualProcessor(listing *ProgramListing, maxVm int, writer *model.GraphWriter) *VirtualProcessor {
	vp := &VirtualProcessor{
		optimizer:      NewIbmOptimizer(),
		maxVmCount:     maxVm,
		programListing: listing,
		graphWriter:    writer,
	}
	// Initial VM
	vp.vms = []*VirtualMachine{{Context: NewVmContext(listing, writer)}}
	return vp
}

func (vp *VirtualProcessor) Run() {
	lastHeartbeat := time.Now()

	for len(vp.vms) > 0 {
		// DFS: Pop from the end of the slice (stack behavior)
		vm := vp.vms[len(vp.vms)-1]
		vp.vms = vp.vms[:len(vp.vms)-1]

		vp.totalSteps++

		// Heartbeat every 5000 steps or every second
		if vp.totalSteps%5000 == 0 || time.Since(lastHeartbeat) > time.Second {
			vp.logHeartbeat()
			lastHeartbeat = time.Now()
		}

		// Update program unit (paragraph/section tracking) - MUST BE BEFORE OPTIMIZER
		vp.updateProgramUnit(vm)

		if vp.optimizer.Apply(vm) {
			continue
		}

		// Execute one step
		newVms, active := vm.Step()
		
		// If the instruction created branches (forks), add them to the stack
		if len(newVms) > 0 {
			vp.vms = append(vp.vms, newVms...)
		}

		// If the current VM is still active (didn't hit STOP RUN), put it back on the stack
		if active {
			vp.vms = append(vp.vms, vm)
		}

		// Guard condition
		if len(vp.vms) > vp.maxVmCount {
			fmt.Printf("[Warning] Maximum VM count reached: %d\n", vp.maxVmCount)
			return
		}
	}
	fmt.Printf("[Done] Analysis finished. Total Steps: %d, Total States: %d\n", vp.totalSteps, vp.optimizer.TotalStates)
}

func (vp *VirtualProcessor) updateProgramUnit(vm *VirtualMachine) {
	inst := vp.programListing.GetInstructionByPosition(vm.Context.IC)
	if inst == nil {
		return
	}
	if pu, ok := inst.(*ProgramUnit); ok {
		node := pu.GetInitialNode()
		if node != nil && node.ID != vm.Context.CurrentUnitID {
			vm.Context.SetCurrentProgramUnitByPosition(vm.Context.IC)
		}
	}
}

func (vp *VirtualProcessor) logHeartbeat() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	heapAlloc := m.HeapAlloc / 1024 / 1024
	fmt.Printf("[Heartbeat] Steps: %d | Active VMs: %d | Total States: %d | Heap: %dMB\n",
		vp.totalSteps, len(vp.vms), vp.optimizer.TotalStates, heapAlloc)
}
