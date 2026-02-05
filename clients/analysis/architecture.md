# Analysis Engine Architecture

The `@code4z/analysis` package implements a **Virtual Execution Engine** that traces the control flow of COBOL programs. Unlike static analyzers that simply look at syntax, this engine "runs" the program through every possible branch to discover how control actually moves between paragraphs and sections.

---

## 1. Orchestration Layer

### `ControlFlowGraphBuilder` (`src/graphbuilder.ts`)
The main entry point. It orchestrates the entire process:
1.  Loads the **CFAST** tree.
2.  Flattens the tree into a **ProgramListing** (a linear array of instructions).
3.  Initializes the **VirtualProcessor** and **IbmOptimizer**.
4.  Triggers the analysis and handles the resulting diagnostics (like Dead Code).

---

## 2. The Execution Engine (`src/vm/`)

### `VirtualProcessor` (`vp.ts`)
The orchestrator of execution paths. It maintains a **Stack** of active `VirtualMachine` instances.
*   **Strategy**: It uses a **Depth-First Search (DFS)** approach. It pops a VM from the stack, executes one step, and pushes any resulting forks (branches) back onto the stack.
*   **Heartbeat**: Periodically logs metrics (Steps, Active VMs, Heap Usage) to provide visibility during long-running analyses.

### `VirtualMachine` & `VmContext` (`vm.ts`)
A `VirtualMachine` is a lightweight runner. Its state is encapsulated in a `VmContext`:
*   **IC (Instruction Counter)**: Points to the current position in the `ProgramListing`.
*   **Path (Optimized)**: A persistent **Linked List** of previously executed instructions. This allows $O(1)$ cloning; when a VM forks, the new VM simply points to the same "head" of the path as its parent.
*   **Maps**: Tracks the state of `PERFORM` returns (`redirectMap`), `ALTER` statements (`alterMap`), and specialized CICS/SQL triggers.

### `IbmOptimizer` (`optimizer.ts`)
The "brain" that ensures the analysis terminates and doesn't explore redundant code.
*   **State Fingerprinting**: At every step, the optimizer calculates a **64-bit hash** of the VM's current state (including all internal maps).
*   **Pruning**: If a VM reaches an instruction in a state that has been seen before, the optimizer stops that VM immediately. This prevents infinite loops and stops the "State Space Explosion."

---

## 3. Instruction Set (`src/vm/instructions.ts`)

The engine doesn't execute raw COBOL; it executes a set of specialized **Instructions** derived from the CFAST:
*   **Simple Instructions**: Standard linear flow (e.g., `MOVE`, `ADD`).
*   **Branching**: `ConditionEntry` handles `IF` and `EVALUATE`, forking the VM into multiple paths.
*   **Transfer of Control**:
    *   `PerformInstruction`: Handles the complex logic of calling a range of paragraphs.
    *   `VNCell`: A virtual "return cell" placed at the end of paragraphs/sections. It checks the VM state to decide where control should return after a `PERFORM`.
    *   `CallInstruction`: Treats external program calls as discrete nodes (black boxes).

---

## 4. Graph Construction (`src/model/Graph.ts`)

### Write-Through Streaming
To handle massive programs without exhausting memory, the engine uses a **Write-Through** mechanism:
1.  Instead of building a giant graph object in RAM, the `Graph` class accepts a **Writer Callback**.
2.  The moment the engine discovers a link (e.g., `Proc-A` calls `Proc-B`), it emits a JSON object for that node or edge.
3.  The CLI redirects this stream to a **JSONL (JSON Lines)** file on disk.
4.  **Result**: RAM usage stays constant regardless of how many millions of nodes or edges the final graph contains.

---

## 5. Execution Lifecycle Summary

1.  **Flattening**: CFAST tree $ightarrow$ Linear `CobolInstruction` array.
2.  **Initialization**: Single `VirtualMachine` starts at the first instruction.
3.  **The Loop**:
    *   `VirtualProcessor` pops a VM.
    *   `IbmOptimizer` checks if this `(Instruction, StateHash)` has been seen.
    *   If new, the `VirtualMachine` executes the instruction.
    *   If the instruction branches (e.g., `IF`), new VMs are cloned and pushed to the stack.
    *   **Control Moves** (Edges) are emitted to the JSONL stream.
4.  **Completion**: When the stack is empty, the analysis is complete.
5.  **Post-Process**: The `DeadCodeCollector` scans the instruction list for any instruction that was never visited by any VM.
