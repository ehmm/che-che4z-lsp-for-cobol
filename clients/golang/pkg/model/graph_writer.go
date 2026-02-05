package model

import (
	"encoding/json"
	"io"
	"sync"
)

// GraphEntry represents a single line in the JSONL output.
type GraphEntry struct {
	Type     string      `json:"type"`
	ID       int32       `json:"id,omitempty"`
	Name     string      `json:"name,omitempty"`
	Kind     string      `json:"kind,omitempty"`
	Location *Location   `json:"location,omitempty"`
	From     int32       `json:"from,omitempty"`
	To       int32       `json:"to,omitempty"`
	Target   interface{} `json:"targetName,omitempty"`
}

// GraphWriter handles streaming graph data to an io.Writer in JSONL format.
type GraphWriter struct {
	writer    io.Writer
	encoder   *json.Encoder
	seenNodes map[int32]struct{}
	seenEdges map[uint64]struct{}
	mu        sync.Mutex
}

func NewGraphWriter(w io.Writer) *GraphWriter {
	return &GraphWriter{
		writer:    w,
		encoder:   json.NewEncoder(w),
		seenNodes: make(map[int32]struct{}),
		seenEdges: make(map[uint64]struct{}),
	}
}

func (gw *GraphWriter) WriteProgram(name string, loc Location) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	return gw.encoder.Encode(GraphEntry{
		Type:     "program",
		Name:     name,
		Location: &loc,
	})
}

func (gw *GraphWriter) WriteNode(id int32, name string, kind string, loc Location) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	if _, seen := gw.seenNodes[id]; seen {
		return nil
	}

	gw.seenNodes[id] = struct{}{}
	return gw.encoder.Encode(GraphEntry{
		Type:     "node",
		ID:       id,
		Name:     name,
		Kind:     kind,
		Location: &loc,
	})
}

func (gw *GraphWriter) WriteEdge(from, to int32) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	// Use a 64-bit key to uniquely identify an edge: (from << 32) | to
	key := (uint64(from) << 32) | uint64(to)
	if _, seen := gw.seenEdges[key]; seen {
		return nil
	}

	gw.seenEdges[key] = struct{}{}
	return gw.encoder.Encode(GraphEntry{
		Type: "edge",
		From: from,
		To:   to,
	})
}
