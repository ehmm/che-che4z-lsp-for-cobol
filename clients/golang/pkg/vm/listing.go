package vm

import (
	"strings"

	"github.com/code4z/ccf-cli/pkg/model"
)

// Placeholder types to be substituted during resolvePlaceholders
type GotoPlaceholder struct {
	BaseInstruction
	Targets []string
}
func (p *GotoPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type PerformPlaceholder struct {
	BaseInstruction
	Target model.ProcedureName
	Thru   model.ProcedureName
}
func (p *PerformPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type IfBranchPlaceholder struct{ BaseInstruction }
func (p *IfBranchPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type EvaluateBranchPlaceholder struct{ BaseInstruction }
func (p *EvaluateBranchPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type ProgramUnitPlaceholder struct {
	BaseInstruction
}
func (p *ProgramUnitPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type ExitSectionPlaceholder struct {
	BaseInstruction
	SectionPosition int
}
func (p *ExitSectionPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type ExitParagraphPlaceholder struct {
	BaseInstruction
	ParagraphPosition int
}
func (p *ExitParagraphPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type ConditionalBlockPlaceholder struct {
	BaseInstruction
	EndType model.NodeType
}
func (p *ConditionalBlockPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type AlterPlaceholder struct {
	BaseInstruction
	From model.ProcedureName
	To   model.ProcedureName
}
func (p *AlterPlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type XmlParsePlaceholder struct {
	BaseInstruction
	Target model.ProcedureName
	Thru   model.ProcedureName
}
func (p *XmlParsePlaceholder) Execute(ctx *VmContext) []int { panic("Placeholder") }

type PerformInfo struct {
	Position          int
	PerformUntilType  *string
	EndCycleIndex     int
	EndPerformIndex   int
}

type SymbolTable struct {
	sections           map[string]int
	paragraphs         map[string]int
	sectionList        []string // To maintain order of discovery for fallback
	currentSection     string
	currentParagraph   string
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		sections:   make(map[string]int),
		paragraphs: make(map[string]int),
	}
}

func (s *SymbolTable) GetProcedurePositionByName(name model.ProcedureName) (int, bool) {
	n := strings.ToUpper(name.Name)
	if name.InSection != nil {
		pos, ok := s.paragraphs[strings.ToUpper(*name.InSection)+":"+n]
		return pos, ok
	}
	// If it's a section, return its start
	if pos, ok := s.sections[n]; ok {
		return pos, true
	}
	// Fallback to searching in all sections in order of discovery
	for _, sect := range s.sectionList {
		if pos, ok := s.paragraphs[sect+":"+n]; ok {
			return pos, true
		}
	}
	return 0, false
}

func (s *SymbolTable) GetProcedurePositionsByNames(target, thru model.ProcedureName) (int, int, bool) {
	start, ok1 := s.GetProcedurePositionByName(target)
	
	// Complex thru logic: if thru is a section, we want its LAST paragraph's VNCell.
	// But the TS engine often just returns the end of the unit.
	end, ok2 := s.GetProcedurePositionByName(thru)
	
	return start, end, ok1 && ok2
}

type ProgramListing struct {
	Program      *model.CFASTNode
	Instructions []Instruction
	SymbolTable  *SymbolTable
	URIs         map[string]struct{}
}

func NewProgramListing(program *model.CFASTNode) *ProgramListing {
	l := &ProgramListing{
		Program:     program,
		SymbolTable: NewSymbolTable(),
		URIs:        make(map[string]struct{}),
	}
	l.build(program)
	return l
}

func (l *ProgramListing) build(program *model.CFASTNode) {
	l.Instructions = append(l.Instructions, &StartNode{})
	l.addChildren(program)
	l.resolvePlaceholders()
}

func (l *ProgramListing) addChildren(node *model.CFASTNode) {
	node.ID = int32(len(l.Instructions))
	if node.Location.Uri != "" {
		l.URIs[node.Location.Uri] = struct{}{}
	}

	l.addInstructionsForNode(node)

	for _, child := range node.Children {
		l.addChildren(child)
	}

	if node.Type == model.NodeTypeSection || node.Type == model.NodeTypeParagraph {
		vnCell := &VNCell{
			DefaultRedirectPosition: len(l.Instructions) + 1,
		}
		l.Instructions = append(l.Instructions, vnCell)
	}
}

func (l *ProgramListing) addInstructionsForNode(node *model.CFASTNode) {
	switch node.Type {
	case model.NodeTypeProgram:
		l.Instructions = append(l.Instructions, &ProgramUnitPlaceholder{BaseInstruction: BaseInstruction{Node: node}})

	case model.NodeTypeSection:
		name := strings.ToUpper(node.Name)
		l.SymbolTable.sections[name] = len(l.Instructions)
		l.SymbolTable.currentSection = name
		l.SymbolTable.sectionList = append(l.SymbolTable.sectionList, name)
		l.SymbolTable.currentParagraph = ""
		l.Instructions = append(l.Instructions, &ProgramUnitPlaceholder{BaseInstruction: BaseInstruction{Node: node}})

	case model.NodeTypeParagraph:
		name := strings.ToUpper(node.Name)
		l.SymbolTable.currentParagraph = name
		l.SymbolTable.paragraphs[l.SymbolTable.currentSection+":"+name] = len(l.Instructions)
		l.Instructions = append(l.Instructions, &ProgramUnitPlaceholder{BaseInstruction: BaseInstruction{Node: node}})

	case model.NodeTypeGoto:
		var targets []string
		if t, ok := node.TargetName.(string); ok {
			targets = []string{t}
		} else if t, ok := node.TargetName.([]interface{}); ok {
			for _, item := range t {
				targets = append(targets, item.(string))
			}
		}
		l.Instructions = append(l.Instructions, &GotoPlaceholder{BaseInstruction: BaseInstruction{Node: node}, Targets: targets})

	case model.NodeTypePerform:
		target := model.ProcedureName{Name: node.Name}
		if node.TargetName != nil {
			target.Name = node.TargetName.(string)
		}
		if node.TargetSectionName != nil {
			target.InSection = node.TargetSectionName
		}
		thru := target
		if node.ThruName != nil {
			thru.Name = *node.ThruName
		}
		if node.ThruSectionName != nil {
			thru.InSection = node.ThruSectionName
		}
		l.Instructions = append(l.Instructions, &PerformPlaceholder{BaseInstruction: BaseInstruction{Node: node}, Target: target, Thru: thru})
		
		if node.PerformUntilType != nil {
			if *node.PerformUntilType == "UNTIL_CONDITION" {
				l.Instructions = append(l.Instructions, &ConditionEntry{
					SimpleInstruction: SimpleInstruction{BaseInstruction{Node: nil}},
					Starts:             []int{len(l.Instructions) + 1},
					End:                len(l.Instructions) + 2,
				})
				l.Instructions = append(l.Instructions, &JumpInstruction{Position: len(l.Instructions) - 1})
				l.Instructions = append(l.Instructions, &ConditionExit{})
			} else if *node.PerformUntilType == "UNTIL_EXIT" {
				l.Instructions = append(l.Instructions, &JumpInstruction{Position: len(l.Instructions)})
				l.Instructions = append(l.Instructions, &SimpleInstruction{})
			}
		}

	case model.NodeTypeIf:
		l.Instructions = append(l.Instructions, &IfBranchPlaceholder{BaseInstruction: BaseInstruction{Node: node}})
	case model.NodeTypeEvaluate:
		l.Instructions = append(l.Instructions, &EvaluateBranchPlaceholder{BaseInstruction: BaseInstruction{Node: node}})

	case model.NodeTypeStop:
		l.Instructions = append(l.Instructions, &StopRunInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}})
	case model.NodeTypeGoback:
		l.Instructions = append(l.Instructions, &GobackInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}})

	case model.NodeTypeExitSection:
		l.Instructions = append(l.Instructions, &ExitSectionPlaceholder{
			BaseInstruction: BaseInstruction{Node: node},
			SectionPosition: l.SymbolTable.sections[l.SymbolTable.currentSection],
		})
	case model.NodeTypeExitParagraph:
		l.Instructions = append(l.Instructions, &ExitParagraphPlaceholder{
			BaseInstruction:   BaseInstruction{Node: node},
			ParagraphPosition: l.SymbolTable.paragraphs[l.SymbolTable.currentSection+":"+l.SymbolTable.currentParagraph],
		})

	case model.NodeTypeAtEnd:
		l.Instructions = append(l.Instructions, &ConditionalBlockPlaceholder{BaseInstruction: BaseInstruction{Node: node}, EndType: model.NodeTypeAtEndExit})
	case model.NodeTypeOnException, model.NodeTypeOnNotException:
		l.Instructions = append(l.Instructions, &ConditionalBlockPlaceholder{BaseInstruction: BaseInstruction{Node: node}, EndType: model.NodeTypeEndOn})

	case model.NodeTypeAlter:
		l.Instructions = append(l.Instructions, &AlterPlaceholder{BaseInstruction: BaseInstruction{Node: node}, From: *node.From, To: *node.To})

	case model.NodeTypeExecCicsHandle:
		l.Instructions = append(l.Instructions, &CicsHandleAbendInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}, HandleType: *node.HandleType, Size: 3})
		l.Instructions = append(l.Instructions, &RestoreProgramUnit{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: nil}}})
		if *node.HandleType == "LABEL" {
			l.Instructions = append(l.Instructions, &GotoPlaceholder{BaseInstruction: BaseInstruction{Node: node}, Targets: []string{*node.Value}})
		} else {
			l.Instructions = append(l.Instructions, &SimpleInstruction{BaseInstruction: BaseInstruction{Node: node}})
		}

	case model.NodeTypeExecCics:
		l.Instructions = append(l.Instructions, &CicsInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}})
	case model.NodeTypeExecCicsReturn:
		l.Instructions = append(l.Instructions, &CicsReturnInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}})
	case model.NodeTypeExecCicsAbend:
		l.Instructions = append(l.Instructions, &CicsAbendInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}, Cancel: *node.Cancel})

	case model.NodeTypeExecWhenever:
		l.Instructions = append(l.Instructions, &SqlWheneverInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}, WheneverCondition: *node.WheneverCondition, Size: 3})
		l.Instructions = append(l.Instructions, &RestoreProgramUnit{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: nil}}})
		if *node.WheneverType == "GOTO" {
			l.Instructions = append(l.Instructions, &GotoPlaceholder{BaseInstruction: BaseInstruction{Node: node}, Targets: []string{*node.Value}})
		} else {
			l.Instructions = append(l.Instructions, &SimpleInstruction{BaseInstruction: BaseInstruction{Node: node}})
		}

	case model.NodeTypeExecSql:
		l.Instructions = append(l.Instructions, &SqlInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}})

	case model.NodeTypeInput, model.NodeTypeOutput:
		l.Instructions = append(l.Instructions, &PerformPlaceholder{BaseInstruction: BaseInstruction{Node: node}, Target: *node.Target, Thru: *node.Thru})

	case model.NodeTypeXmlParse:
		l.Instructions = append(l.Instructions, &XmlParsePlaceholder{BaseInstruction: BaseInstruction{Node: node}, Target: *node.Target, Thru: *node.Thru})

	case model.NodeTypeCall:
		l.Instructions = append(l.Instructions, &CallInstruction{ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}}}})

	default:
		l.Instructions = append(l.Instructions, &SimpleInstruction{BaseInstruction: BaseInstruction{Node: node}})
	}
}

func (l *ProgramListing) resolvePlaceholders() {
	utils := ListingUtils{}
	for i := 0; i < len(l.Instructions); i++ {
		inst := l.Instructions[i]
		switch p := inst.(type) {
		case *GotoPlaceholder:
			pos, ok := l.SymbolTable.GetProcedurePositionByName(model.ProcedureName{Name: p.Targets[0]})
			if ok {
				l.Instructions[i] = &GotoInstruction{
					ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: p.Node}}},
					Position:             pos,
				}
			} else {
				l.Instructions[i] = &SimpleInstruction{BaseInstruction: BaseInstruction{Node: p.Node}}
			}

		case *PerformPlaceholder:
			posStart, posEnd, ok := l.SymbolTable.GetProcedurePositionsByNames(p.Target, p.Thru)
			if ok {
				l.Instructions[i] = &PerformInstruction{
					ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: p.Node}}},
					Target:               posStart,
					Thru:                 posEnd,
				}
			} else {
				l.Instructions[i] = &SimpleInstruction{BaseInstruction: BaseInstruction{Node: p.Node}}
			}

		case *IfBranchPlaceholder:
			info := utils.BuildIfConditionInfo(l.Instructions, i)
			l.substituteBranchPlaceholders(i, info)

		case *EvaluateBranchPlaceholder:
			info := utils.BuildEvaluateConditionInfo(l.Instructions, i)
			l.substituteBranchPlaceholders(i, info)

		case *ConditionalBlockPlaceholder:
			info := utils.BuildSimpleConditionInfo(l.Instructions, i, p.Node.Type, p.EndType)
			l.substituteBranchPlaceholders(i, info)

		case *ExitSectionPlaceholder:
			vnCellPos := -1
			for j := p.SectionPosition + 1; j < len(l.Instructions); j++ {
				if _, ok := l.Instructions[j].(*VNCell); ok {
					vnCellPos = j
				} else if pu, ok := l.Instructions[j].(*ProgramUnit); ok {
					if pu.Node.Type == model.NodeTypeSection { break }
				}
			}
			l.Instructions[i] = &ExitSection{
				ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: p.Node}}},
				SectionVnCellPosition: vnCellPos,
			}

		case *ExitParagraphPlaceholder:
			vnCellPos := -1
			for j := p.ParagraphPosition + 1; j < len(l.Instructions); j++ {
				if _, ok := l.Instructions[j].(*VNCell); ok {
					vnCellPos = j
					break
				}
			}
			l.Instructions[i] = &ExitParagraph{
				ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: p.Node}}},
				ParagraphVnCellPosition: vnCellPos,
			}

		case *AlterPlaceholder:
			from, ok1 := l.SymbolTable.GetProcedurePositionByName(p.From)
			to, ok2 := l.SymbolTable.GetProcedurePositionByName(p.To)
			if ok1 && ok2 {
				l.Instructions[i] = &AlterInstruction{
					ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: p.Node}}},
					From:                 from,
					To:                   to,
				}
			} else {
				l.Instructions[i] = &SimpleInstruction{BaseInstruction: BaseInstruction{Node: p.Node}}
			}

		case *XmlParsePlaceholder:
			posStart, posEnd, ok := l.SymbolTable.GetProcedurePositionsByNames(p.Target, p.Thru)
			if ok {
				l.Instructions[i] = &PerformInstruction{
					ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: p.Node}}},
					Target:               posStart,
					Thru:                 posEnd,
				}
			} else {
				l.Instructions[i] = &SimpleInstruction{BaseInstruction: BaseInstruction{Node: p.Node}}
			}

		case *ProgramUnitPlaceholder:
			vnCellPos := -1
			for j := i + 1; j < len(l.Instructions); j++ {
				if _, ok := l.Instructions[j].(*VNCell); ok {
					vnCellPos = j
					if p.Node.Type == model.NodeTypeParagraph { break }
				} else if pu, ok := l.Instructions[j].(*ProgramUnitPlaceholder); ok {
					if pu.Node.Type == model.NodeTypeParagraph || pu.Node.Type == model.NodeTypeSection {
						break
					}
				}
			}
			l.Instructions[i] = &ProgramUnit{
				ImportantInstruction: ImportantInstruction{SimpleInstruction: SimpleInstruction{BaseInstruction{Node: p.Node}}},
				VnCellPosition:       vnCellPos,
			}
		}
	}
}

func (l *ProgramListing) substituteBranchPlaceholders(pos int, info ConditionInfo) {
	l.Instructions[pos] = &ConditionEntry{
		SimpleInstruction: SimpleInstruction{BaseInstruction{Node: l.Instructions[pos].GetInitialNode()}},
		Starts:             info.Starts,
		End:                info.End,
		Closed:             info.Closed,
		BranchEnds:         info.Elses,
	}
	for _, elseIdx := range info.Elses {
		l.Instructions[elseIdx] = &ConditionBranchEnd{
			SimpleInstruction: SimpleInstruction{BaseInstruction{Node: l.Instructions[elseIdx].GetInitialNode()}},
			EndPosition:       info.End,
		}
	}
	l.Instructions[info.End] = &ConditionExit{
		SimpleInstruction: SimpleInstruction{BaseInstruction{Node: l.Instructions[info.End].GetInitialNode()}},
	}
}

func (l *ProgramListing) GetInstructionByPosition(ic int) Instruction {
	if ic < 0 || ic >= len(l.Instructions) {
		return nil
	}
	return l.Instructions[ic]
}
