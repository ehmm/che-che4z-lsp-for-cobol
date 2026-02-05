package vm

import "github.com/code4z/ccf-cli/pkg/model"

type ConditionInfo struct {
	Starts []int
	Elses  []int
	End    int
	Closed bool
}

type ListingUtils struct{}

func (u ListingUtils) BuildIfConditionInfo(instructions []Instruction, position int) ConditionInfo {
	starts := []int{position + 1}
	var elses []int
	var end int
	counter := 1

	for i := position + 1; i < len(instructions); i++ {
		node := instructions[i].GetInitialNode()
		if node == nil {
			continue
		}
		if node.Type == model.NodeTypeIf {
			counter++
			continue
		}
		if node.Type == model.NodeTypeElse && counter == 1 {
			starts = append(starts, i+1)
			elses = append(elses, i)
			continue
		}
		if node.Type == model.NodeTypeEndif {
			counter--
			if counter == 0 {
				end = i
				break
			}
		}
	}
	return ConditionInfo{Starts: starts, Elses: elses, End: end, Closed: len(elses) > 0}
}

func (u ListingUtils) BuildEvaluateConditionInfo(instructions []Instruction, position int) ConditionInfo {
	var starts []int
	var elses []int
	var end int
	closed := false
	counter := 1

	for i := position + 1; i < len(instructions); i++ {
		node := instructions[i].GetInitialNode()
		if node == nil {
			continue
		}
		if node.Type == model.NodeTypeEvaluate {
			counter++
			continue
		}
		if node.Type == model.NodeTypeWhen && counter == 1 {
			starts = append(starts, i+1)
			elses = append(elses, i)
			continue
		}
		if node.Type == model.NodeTypeWhenOther && counter == 1 {
			starts = append(starts, i+1)
			elses = append(elses, i)
			closed = true
			continue
		}
		if node.Type == model.NodeTypeEndEvaluate {
			counter--
			if counter == 0 {
				end = i
				break
			}
		}
	}
	return ConditionInfo{Starts: starts, Elses: elses, End: end, Closed: closed}
}

func (u ListingUtils) BuildSimpleConditionInfo(instructions []Instruction, position int, startType, endType model.NodeType) ConditionInfo {
	starts := []int{position + 1}
	var end int
	counter := 1

	for i := position + 1; i < len(instructions); i++ {
		node := instructions[i].GetInitialNode()
		if node == nil {
			continue
		}
		if node.Type == startType {
			counter++
			continue
		}
		if node.Type == endType {
			counter--
			if counter == 0 {
				end = i
				break
			}
		}
	}
	return ConditionInfo{Starts: starts, Elses: []int{}, End: end, Closed: false}
}
