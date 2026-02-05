package vm

import (
	"github.com/code4z/ccf-cli/pkg/model"
)

type PerformStorageItem struct {
	VnCellPosition   int
	RedirectPosition int
	ProgramUnit      *model.CFASTNode
	PathHead         *PathNode
}

type VmContext struct {
	IC             int
	ProgramListing *ProgramListing
	GraphWriter    *model.GraphWriter
	RedirectMap    map[int]int
	AlterMap       map[int]int
	StickyMap      map[string]int
	Storage        map[int]PerformStorageItem
	PathHead       *PathNode
	NestedLevel    int
	CurrentUnitID  int32
}

func NewVmContext(listing *ProgramListing, writer *model.GraphWriter) *VmContext {
	return &VmContext{
		IC:             0,
		ProgramListing: listing,
		GraphWriter:    writer,
		RedirectMap:    make(map[int]int),
		AlterMap:       make(map[int]int),
		StickyMap:      make(map[string]int),
		Storage:        make(map[int]PerformStorageItem),
	}
}

func (c *VmContext) Clone() *VmContext {
	newCtx := &VmContext{
		IC:             c.IC,
		ProgramListing: c.ProgramListing,
		GraphWriter:    c.GraphWriter,
		RedirectMap:    make(map[int]int, len(c.RedirectMap)),
		AlterMap:       make(map[int]int, len(c.AlterMap)),
		StickyMap:      make(map[string]int, len(c.StickyMap)),
		Storage:        make(map[int]PerformStorageItem, len(c.Storage)),
		PathHead:       c.PathHead, // Persistent linked list head is shared
		NestedLevel:    c.NestedLevel,
		CurrentUnitID:  c.CurrentUnitID,
	}

	for k, v := range c.RedirectMap {
		newCtx.RedirectMap[k] = v
	}
	for k, v := range c.AlterMap {
		newCtx.AlterMap[k] = v
	}
	for k, v := range c.StickyMap {
		newCtx.StickyMap[k] = v
	}
	for k, v := range c.Storage {
		newCtx.Storage[k] = v
	}

	return newCtx
}

func (c *VmContext) GetInstructionByPosition(ic int) Instruction {
	return c.ProgramListing.GetInstructionByPosition(ic)
}

func (c *VmContext) GetRedirectPosition(defaultPos int) int {
	if pos, ok := c.RedirectMap[defaultPos]; ok {
		return pos
	}
	return defaultPos
}

func (c *VmContext) Redirect(defaultPos, performPos int) int {
	prev := c.GetRedirectPosition(defaultPos)
	c.RedirectMap[defaultPos] = performPos

	c.Storage[performPos] = PerformStorageItem{
		VnCellPosition:   defaultPos - 1,
		RedirectPosition: prev,
		PathHead:         c.PathHead,
	}
	return prev
}

func (c *VmContext) DeactivatePerform(performPos int) {
	if item, ok := c.Storage[performPos]; ok {
		if item.RedirectPosition == item.VnCellPosition+1 {
			delete(c.RedirectMap, item.VnCellPosition+1)
		} else {
			c.RedirectMap[item.VnCellPosition+1] = item.RedirectPosition
		}
		delete(c.Storage, performPos)
		c.PathHead = item.PathHead
	}
}

func (c *VmContext) AddToPath() {
	inst := c.GetInstructionByPosition(c.IC)
	if inst != nil {
		c.PathHead = &PathNode{
			Instruction: inst,
			Parent:      c.PathHead,
		}
	}
}

func (c *VmContext) SetCurrentProgramUnitByPosition(pos int) {
	inst := c.GetInstructionByPosition(pos)
	if inst == nil {
		return
	}
	node := inst.GetInitialNode()
	if node == nil {
		return
	}

	if node.Type == model.NodeTypeParagraph || node.Type == model.NodeTypeSection || node.Type == model.NodeTypeProgram {
		if c.GraphWriter != nil {
			// Write the node
			c.GraphWriter.WriteNode(node.ID, node.Name, string(node.Type), node.Location)
			// Write the edge from the previous unit
			if c.CurrentUnitID != 0 {
				c.GraphWriter.WriteEdge(c.CurrentUnitID, node.ID)
			}
		}
		c.CurrentUnitID = node.ID
	}
}

func (c *VmContext) AddAtler(from, to int) {
	c.AlterMap[from] = to
}

func (c *VmContext) AddHandleAbend() {
	c.StickyMap["%handle_abend%"] = c.IC
}

func (c *VmContext) ResetHandleAbend() {
	delete(c.StickyMap, "%handle_abend%")
}

func (c *VmContext) AddSqlWhenever(condition string) {
	c.StickyMap["%sql_whenever%"+condition] = c.IC
}

func (c *VmContext) GetHandleAbendEntry() int {
	return c.StickyMap["%handle_abend%"]
}

func (c *VmContext) GetSqlWheneverEntries() []int {
	var res []int
	for _, cond := range []string{"NOT_FOUND", "SQLERROR", "SQLWARNING"} {
		if pos, ok := c.StickyMap["%sql_whenever%"+cond]; ok {
			res = append(res, pos)
		}
	}
	return res
}

func (c *VmContext) GetAlterMap() map[int]int {
	return c.AlterMap
}

func (c *VmContext) IncNestedLevel() {
	c.NestedLevel++
}

func (c *VmContext) DecNestedLevel() int {
	c.NestedLevel--
	return c.NestedLevel
}

func (c *VmContext) GetNestedLevel() int {
	return c.NestedLevel
}

func (c *VmContext) GenerateState() *VirtualMachineState {
	return NewVirtualMachineState(c.RedirectMap, c.AlterMap, c.StickyMap)
}
