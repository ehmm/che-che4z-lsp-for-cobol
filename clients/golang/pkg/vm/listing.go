package vm

import (
	"strings"

	"github.com/code4z/ccf-cli/pkg/model"
)

type SymbolTable struct {
	sections       map[string]int
	paragraphs     map[string]int
	currentSection string
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		sections:   make(map[string]int),
		paragraphs: make(map[string]int),
	}
}

type ProgramListing struct {
	Program      *model.CFASTNode
	Instructions []Instruction
	SymbolTable  *SymbolTable
}

func NewProgramListing(program *model.CFASTNode) *ProgramListing {
	l := &ProgramListing{
		Program:     program,
		SymbolTable: NewSymbolTable(),
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
	l.addInstructionsForNode(node)

	for _, child := range node.Children {
		l.addChildren(child)
	}

	// Logic for VNCells at end of units
	if node.Type == model.NodeTypeSection || node.Type == model.NodeTypeParagraph {
		vnCell := &VNCell{
			DefaultRedirectPosition: len(l.Instructions) + 1,
		}
		// Update the ProgramUnit instruction with the VNCell index
		if pu, ok := l.Instructions[node.ID].(*ProgramUnit); ok {
			pu.VnCellPosition = len(l.Instructions)
		}
		l.Instructions = append(l.Instructions, vnCell)
	}
}

func (l *ProgramListing) addInstructionsForNode(node *model.CFASTNode) {
	switch node.Type {
	case model.NodeTypeProgram:
		l.Instructions = append(l.Instructions, &ProgramUnit{
			ImportantInstruction: ImportantInstruction{
				SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}},
			},
		})

	case model.NodeTypeSection:
		l.SymbolTable.sections[strings.ToUpper(node.Name)] = len(l.Instructions)
		l.SymbolTable.currentSection = strings.ToUpper(node.Name)
		l.Instructions = append(l.Instructions, &ProgramUnit{
			ImportantInstruction: ImportantInstruction{
				SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}},
			},
		})

	case model.NodeTypeParagraph:
		l.SymbolTable.paragraphs[l.SymbolTable.currentSection+":"+strings.ToUpper(node.Name)] = len(l.Instructions)
		l.Instructions = append(l.Instructions, &ProgramUnit{
			ImportantInstruction: ImportantInstruction{
				SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}},
			},
		})

	case model.NodeTypeStop:
		l.Instructions = append(l.Instructions, &StopRunInstruction{
			SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}},
		})

	case model.NodeTypeGoback:
		l.Instructions = append(l.Instructions, &GobackInstruction{
			SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}},
		})

	case model.NodeTypeCall:
		l.Instructions = append(l.Instructions, &CallInstruction{
			ImportantInstruction: ImportantInstruction{
				SimpleInstruction: SimpleInstruction{BaseInstruction{Node: node}},
			},
		})

	default:
		// Fallback for simple statements
		l.Instructions = append(l.Instructions, &SimpleInstruction{
			BaseInstruction: BaseInstruction{Node: node},
		})
	}
}

func (l *ProgramListing) resolvePlaceholders() {
	// To be expanded with IF/PERFORM logic resolution using SymbolTable
}

func (l *ProgramListing) GetInstructionByPosition(ic int) Instruction {
	if ic < 0 || ic >= len(l.Instructions) {
		return nil
	}
	return l.Instructions[ic]
}
