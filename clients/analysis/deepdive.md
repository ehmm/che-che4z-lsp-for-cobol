# @code4z/analysis Deep Dive

This package provides the core logic for COBOL Control Flow analysis. It transforms a Control Flow Abstract Syntax Tree (CFAST) into a Control Flow Graph (CFG) and identifies various potential issues like unreachable code or implicit exits.

## Core Concepts

### Virtual Execution
Instead of traditional static analysis, this package uses a **Virtual Execution** approach. It "runs" the COBOL program using a set of Virtual Machines (VMs) to explore all possible execution paths. This is particularly useful for handling COBOL's complex control flow mechanisms like `PERFORM ... THRU`, `GO TO`, and `ALTER`.

### CFAST (Control Flow Abstract Syntax Tree)
The input to the analysis is a `CFAST`. It's a simplified AST focused on control flow statements (Program, Section, Paragraph, If, Evaluate, Perform, Go To, etc.).

### VNCell (Virtual Node Cell)
A key concept in this implementation is the `VNCell`. In COBOL, when a `PERFORM` completes, control returns to the statement following the `PERFORM`. `VNCell`s are virtual instructions placed at the end of paragraphs and sections to handle this return logic dynamically based on the VM state.

## Architecture

### 1. Model (`src/model/`)
Defines the data structures used throughout the analysis:
- `cfast.ts`: Defines the nodes of the input tree.
- `Graph.ts`: Defines the resulting Control Flow Graph.
- `Node.ts` / `GraphDTO.ts`: Data Transfer Objects for exporting the graph.
- `external.ts`: Defines DTOs for diagnostics and locations, compatible with LSP.

### 2. Virtual Machine (`src/vm/`)
The engine that performs the virtual execution:
- `vm.ts`: Contains the `VirtualMachine` class, which holds the execution state (`VmContext`), and `VirtualMachineState`.
- `vp.ts`: The `VirtualProcessor` manages multiple `VirtualMachine` instances. It handles branching (e.g., when an `IF` statement is encountered, it forks the VM).
- `instructions.ts`: Defines various `CobolInstruction` implementations (e.g., `PerformInstruction`, `GotoInstruction`, `ConditionEntry`). Each instruction knows how to update the `VmContext` and where to move the instruction pointer.
- `listing.ts`: Responsible for flattening the `CFAST` into a linear `ProgramListing` of `CobolInstruction`s. It also builds a symbol table for sections and paragraphs.
- `optimizer.ts`: The `IbmOptimizer` prevents infinite loops and redundant path exploration by tracking VM states at each instruction. If a VM reaches the same instruction with the same state, it's stopped.
- `listener.ts`: An interface for callbacks during execution (e.g., reporting control moves or warnings).

### 3. Analysis Logic (`src/`)
- `graphbuilder.ts`: The main entry point. `ControlFlowGraphBuilder` uses the `VirtualProcessor` to build the graph and then uses `DeadCodeCollector` to find unreachable segments.
- `utils.ts`: General utilities for range calculation and other shared logic.
- `consts.ts`: Error messages and other constants.

## Key Features

- **Control Flow Graph Generation**: Produces a visualizable graph of how control moves between programs, sections, and paragraphs.
- **Dead Code Detection**: Identifies `CFAST` nodes that were never visited by any `VirtualMachine` during execution.
- **Fall-through Detection**: Warns about implicit exits from programs (executing the last statement without a `GOBACK` or `STOP RUN`).
- **Complex Statement Support**:
    - **PERFORM ... THRU**: Handled by dynamically redirecting `VNCell`s.
    - **ALTER**: Supported by maintaining an `alterMap` in the `VmContext`.
    - **CICS/SQL**: Specialized instructions handle the unique control flow of CICS (`HANDLE ABEND`) and SQL (`WHENEVER`) statements.

## Usage Flow

1. Flatten `CFAST` into a `ProgramListing` of `CobolInstruction`s (`listing.ts`).
2. Initialize `VirtualProcessor` with the listing.
3. The `VirtualProcessor` runs one or more `VirtualMachine`s.
4. Each `VirtualMachine` steps through instructions, forking when it hits conditions or branching points.
5. The `BuildGraphListener` records links between nodes as the VMs move control.
6. The `IbmOptimizer` ensures the process terminates by pruning redundant states.
7. After execution, `DeadCodeCollector` scans for unvisited instructions.
8. The final `EngineProcessingResult` contains the graph, diagnostics, and events.
