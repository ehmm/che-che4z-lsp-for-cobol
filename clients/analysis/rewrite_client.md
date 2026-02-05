# Plan: Rewriting Analysis Engine in Go

This document outlines the plan to migrate the `@code4z/analysis` TypeScript package to a Go-based binary. Go is an ideal choice for this engine due to its superior memory management (lower object overhead), fast execution, and native support for efficient bitwise operations used in state hashing.

---

## Task Phase 1: Data Models & CFAST Parsing

**Goal**: Define the core structures and implement JSON decoding for the CFAST input.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/model/cfast.ts`
*   **Go Task**: Create a `pkg/model` package.
*   **Implementation**: Define structs for `CFASTNode`, `Program`, `Paragraph`, `Section`, `Perform`, etc. Use Go struct tags (e.g., `` `json:"type"` ``) to match the Java-generated JSON.
*   **Optimization**: Use `int32` for IDs and line/character numbers to keep the struct size small.

---

## Task Phase 2: Instruction Set & Flattening

**Goal**: Convert the CFAST tree into a flat slice of executable instructions.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/vm/instructions.ts`
*   **Go Task**: Create a `pkg/vm` package with an `Instruction` interface.
*   **Implementation**: 
    *   Implement Go types for `SimpleInstruction`, `PerformInstruction`, `GotoInstruction`, `VNCell`, and `ConditionEntry`.
    *   Use an `Execute(ctx *VmContext) []int` signature.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/vm/listing.ts`
*   **Go Task**: Implement the `ProgramListing` builder.
*   **Logic**: A recursive function to traverse the `CFASTNode` tree and append to an `[]Instruction` slice.
*   **Symbol Table**: Use a `map[string]int` to store paragraph/section names to instruction pointer mappings.

---

## Task Phase 3: The Virtual Machine (VM)

**Goal**: Implement the core execution logic, path management, and state tracking.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/vm/vm.ts`
*   **VmContext**: A struct containing the current instruction pointer (`IC`), `redirectMap`, `alterMap`, and `stickyMap`.
*   **Path Cloning (Shared Persistent Path)**: 
    *   In Go, implement the `PathNode` linked list.
    *   Cloning a VM in Go is a simple struct copy: `newVm := *oldVm`.
*   **State Fingerprinting**:
    *   Implement the fingerprint string generation.
    *   **Optimization**: Instead of returning a string, use a `bytes.Buffer` to build the state and immediately generate a `uint64` hash (using `hash/fnv` or `github.com/cespare/xxhash`).

---

## Task Phase 4: The Optimizer & Processor

**Goal**: Implement the execution loop and state pruning.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/vm/optimizer.ts`
*   **Implementation**: Use a `map[int]map[uint64]struct{}` to track seen states.
*   **Memory Win**: Go's `map` with `uint64` keys is significantly more memory-efficient than JavaScript's `Set<string>`.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/vm/vp.ts`
*   **Strategy**: Use a Slice as a Stack for **DFS** execution.
*   **Loop**: `for len(vms) > 0 { vm := vms[len(vms)-1]; vms = vms[:len(vms)-1]; ... }`.

---

## Task Phase 5: Streaming JSONL Writer

**Goal**: Implement the write-through mechanism to keep RAM constant.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/model/Graph.ts`
*   **Implementation**: Create a `GraphWriter` that wraps a `*bufio.Writer`.
*   **Logic**: Every time a control move is detected, use `json.NewEncoder(writer).Encode(entry)`.
*   **Edge Tracking**: Use a `map[uint64]struct{}` where the key is `(ParentID << 32) | ChildID` to track seen edges with zero object overhead.

---

## Task Phase 6: CLI & Build Pipeline

**Goal**: Create the entry point and integrate into the Docker workflow.

### Class Reference: `che-che4z-lsp-for-cobol/clients/analysis/src/bin/ccf-cli.ts`
*   **Go Task**: Create `cmd/ccf-cli/main.go`.
*   **Implementation**: Use the `flag` package or `spf13/cobra` for argument parsing.
*   **Docker Integration**: Update `grapher/Dockerfile` to include a Go build stage:
    ```dockerfile
    FROM golang:1.21-alpine AS go-build
    COPY go/ /src/go/
    RUN cd /src/go && go build -o /ccf-cli cmd/ccf-cli/main.go
    ```

---

## Expected Advantages of Go Rewrite

1.  **Heap Stability**: Go's GC is optimized for low latency and handles millions of small pointers much better than V8's "mark-and-sweep."
2.  **Compact State Map**: Storing `uint64` hashes in a Go map uses roughly 1/4 the RAM of JavaScript's `Set<string>`.
3.  **True Parallelism**: If needed, different programs (CFAST files) can be analyzed in parallel using Goroutines with zero overhead.
4.  **Static Binary**: The final output is a single, small binary with no dependency on Node.js or `npm install`.
