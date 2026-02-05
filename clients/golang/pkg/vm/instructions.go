package vm

import (
	"github.com/code4z/ccf-cli/pkg/model"
)

type Instruction interface {
	GetInitialNode() *model.CFASTNode
	Execute(ctx *VmContext) []int
	IsProcessed() bool
	GetLocation() *model.Location
	MarkProcessed()
}

type BaseInstruction struct {
	Node *model.CFASTNode
}

func (b *BaseInstruction) GetInitialNode() *model.CFASTNode {
	return b.Node
}

func (b *BaseInstruction) IsProcessed() bool {
	if b.Node != nil {
		return b.Node.Processed
	}
	return false
}

func (b *BaseInstruction) GetLocation() *model.Location {
	if b.Node != nil {
		return &b.Node.Location
	}
	return nil
}

func (b *BaseInstruction) MarkProcessed() {
	if b.Node != nil {
		b.Node.Processed = true
	}
}

type SimpleInstruction struct {
	BaseInstruction
}

func (s *SimpleInstruction) Execute(ctx *VmContext) []int {
	s.MarkProcessed()
	return []int{ctx.IC + 1}
}

type ImportantInstruction struct {
	SimpleInstruction
}

func (i *ImportantInstruction) Execute(ctx *VmContext) []int {
	if i.Node != nil && i.Node.Type != model.NodeTypeProgram {
		ctx.AddToPath()
	}
	return i.SimpleInstruction.Execute(ctx)
}

type ProgramUnit struct {
	ImportantInstruction
	VnCellPosition int
}

type VNCell struct {
	BaseInstruction
	DefaultRedirectPosition int
}

func (v *VNCell) Execute(ctx *VmContext) []int {
	pos := ctx.GetRedirectPosition(v.DefaultRedirectPosition)
	next := ctx.GetInstructionByPosition(pos)
	if _, ok := next.(*PerformInstruction); ok {
		ctx.DeactivatePerform(pos)
		pos++
	}
	return []int{pos}
}

func (v *VNCell) IsProcessed() bool { return true }

type PerformInstruction struct {
	ImportantInstruction
	Target int
	Thru   int
}

func (p *PerformInstruction) Execute(ctx *VmContext) []int {
	p.ImportantInstruction.Execute(ctx)
	thruInst := ctx.GetInstructionByPosition(p.Thru)
	if pu, ok := thruInst.(*ProgramUnit); ok {
		ctx.Redirect(pu.VnCellPosition+1, ctx.IC)
	}
	ctx.SetCurrentProgramUnitByPosition(p.Target)
	return []int{p.Target}
}

type GotoInstruction struct {
	ImportantInstruction
	Position int
}

func (g *GotoInstruction) Execute(ctx *VmContext) []int {
	g.ImportantInstruction.Execute(ctx)
	alterMap := ctx.GetAlterMap()
	pos := g.Position
	for {
		if next, ok := alterMap[pos]; ok {
			pos = next
		} else {
			break
		}
	}
	ctx.SetCurrentProgramUnitByPosition(pos)
	return []int{pos}
}

type StartNode struct {
	SimpleInstruction
}

func (s *StartNode) IsProcessed() bool { return true }
func (s *StartNode) Execute(ctx *VmContext) []int {
	return []int{1}
}

type StopRunInstruction struct {
	SimpleInstruction
}

func (s *StopRunInstruction) Execute(ctx *VmContext) []int {
	return []int{-1}
}

type GobackInstruction struct {
	SimpleInstruction
}

func (s *GobackInstruction) Execute(ctx *VmContext) []int {
	return []int{-1}
}

type JumpInstruction struct {
	BaseInstruction
	Position int
}

func (j *JumpInstruction) Execute(ctx *VmContext) []int {
	return []int{j.Position}
}

func (j *JumpInstruction) IsProcessed() bool { return true }

type ConditionEntry struct {
	SimpleInstruction
	Starts     []int
	End        int
	Closed     bool
	BranchEnds []int
}

func (c *ConditionEntry) Execute(ctx *VmContext) []int {
	c.MarkProcessed()
	endInst := ctx.GetInstructionByPosition(c.End)
	if endInst != nil {
		endInst.MarkProcessed()
	}
	ctx.IncNestedLevel()

	for _, be := range c.BranchEnds {
		inst := ctx.GetInstructionByPosition(be)
		if inst != nil {
			inst.MarkProcessed()
		}
	}

	var result []int
	if !c.Closed {
		result = append(result, c.End)
		result = append(result, c.Starts...)
	} else {
		result = append(result, c.Starts[len(c.Starts)-1])
		result = append(result, c.Starts[:len(c.Starts)-1]...)
	}
	return result
}

type ConditionExit struct {
	SimpleInstruction
}

func (c *ConditionExit) Execute(ctx *VmContext) []int {
	c.MarkProcessed()
	ctx.DecNestedLevel()
	return []int{ctx.IC + 1}
}

type ConditionBranchEnd struct {
	SimpleInstruction
	EndPosition int
}

func (c *ConditionBranchEnd) Execute(ctx *VmContext) []int {
	return []int{c.EndPosition}
}

type CallInstruction struct {
	ImportantInstruction
}

type RestoreProgramUnit struct {
	SimpleInstruction
}

func (r *RestoreProgramUnit) Execute(ctx *VmContext) []int {
	node := r.GetInitialNode()
	if node != nil {
		ctx.RestoreCurrentProgramUnit(node.ID)
	}
	return []int{ctx.IC + 1}
}

func (r *RestoreProgramUnit) IsProcessed() bool { return true }

type ExitSection struct {
	ImportantInstruction
	SectionVnCellPosition int
}

func (e *ExitSection) Execute(ctx *VmContext) []int {
	e.ImportantInstruction.Execute(ctx)
	return []int{e.SectionVnCellPosition}
}

type ExitParagraph struct {
	ImportantInstruction
	ParagraphVnCellPosition int
}

func (e *ExitParagraph) Execute(ctx *VmContext) []int {
	e.ImportantInstruction.Execute(ctx)
	return []int{e.ParagraphVnCellPosition}
}

type AlterInstruction struct {
	ImportantInstruction
	From int
	To   int
}

func (a *AlterInstruction) Execute(ctx *VmContext) []int {
	ctx.AddAtler(a.From, a.To)
	return a.ImportantInstruction.Execute(ctx)
}

type CicsHandleAbendInstruction struct {
	SimpleInstruction
	HandleType string
	Size       int
}

func (c *CicsHandleAbendInstruction) Execute(ctx *VmContext) []int {
	if c.HandleType == "LABEL" {
		ctx.AddHandleAbend()
	}
	if c.HandleType == "RESET" {
		ctx.ResetHandleAbend()
	}
	return []int{ctx.IC + c.Size}
}

type CicsInstruction struct {
	SimpleInstruction
}

func (c *CicsInstruction) Execute(ctx *VmContext) []int {
	c.MarkProcessed()
	pos := ctx.GetHandleAbendEntry()
	if pos > 0 {
		return []int{ctx.IC + 1, pos + 1}
	}
	return []int{ctx.IC + 1}
}

type CicsReturnInstruction struct {
	SimpleInstruction
}

func (c *CicsReturnInstruction) Execute(ctx *VmContext) []int {
	c.MarkProcessed()
	pos := ctx.GetHandleAbendEntry()
	if pos > 0 {
		return []int{pos + 1}
	}
	return []int{}
}

type CicsAbendInstruction struct {
	SimpleInstruction
	Cancel bool
}

func (c *CicsAbendInstruction) Execute(ctx *VmContext) []int {
	c.MarkProcessed()
	if c.Cancel {
		return []int{}
	}
	pos := ctx.GetHandleAbendEntry()
	if pos > 0 {
		return []int{pos + 1}
	}
	return []int{}
}

type SqlWheneverInstruction struct {
	SimpleInstruction
	WheneverCondition string
	Size              int
}

func (s *SqlWheneverInstruction) Execute(ctx *VmContext) []int {
	ctx.AddSqlWhenever(s.WheneverCondition)
	return []int{ctx.IC + s.Size}
}

type SqlInstruction struct {
	SimpleInstruction
}

func (s *SqlInstruction) Execute(ctx *VmContext) []int {
	s.MarkProcessed()
	result := []int{ctx.IC + 1}
	entries := ctx.GetSqlWheneverEntries()
	for _, p := range entries {
		result = append(result, p+1)
	}
	return result
}
