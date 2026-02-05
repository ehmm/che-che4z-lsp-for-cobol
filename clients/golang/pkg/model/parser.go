package model

import (
	"encoding/json"
	"os"
)

// LoadCFAST reads a CFAST JSON file and decodes it into a slice of CFASTNode pointers.
func LoadCFAST(filePath string) ([]*CFASTNode, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var nodes []*CFASTNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}
