# COBOL Control Flow Grapher (Grapher)

This tool generates a Control Flow Graph (CFG) from preprocessed COBOL source files. It leverages a Java-based parser for CFAST generation and a TypeScript-based Virtual Machine for execution path exploration.

## Features & Optimizations

During this development session, we implemented several critical optimizations to handle massive (20,000+ line) COBOL dispatcher programs that originally caused Out-of-Memory (OOM) errors:

1.  **DFS (Depth-First Search) Strategy**: Switched the Virtual Processor from BFS to DFS. This reduces the number of concurrent active Virtual Machines from 50,000+ to less than 100, drastically lowering RAM usage.
2.  **64-bit State Hashing**: Replaced heavy Java-style Map objects in the Optimizer with compact 64-bit fingerprint hashes. This reduced the memory footprint of execution state tracking by ~90%.
3.  **Shared Persistent Paths**: Refactored VM path management to use a Linked List (Cons-cell) structure. Cloning a VM is now an $O(1)$ operation that shares history with its parent instead of doing a full array copy.
4.  **Write-Through Streaming**: Implemented a streaming JSONL (JSON Lines) writer. Nodes and edges are written to disk the moment they are discovered, eliminating the need to store the entire resulting graph in memory.
5.  **Snippet Stripping**: Added a pre-analysis step to strip heavy code snippets from the AST, saving further heap space.
6.  **External Call Support**: Added logic to treat COBOL `CALL` statements as discrete nodes in the graph without attempting to step into external programs.

---

## How to Build

The tool is packaged as a multi-stage Docker image that includes both the Java parsing engine and the TypeScript analysis binary.

Use the provided build script:
```bash
./grapher/build-cfast-docker.sh
```
This will create a Docker image tagged as `grapher:latest`.

---

## How to Use

The easiest way to run the tool is via the `run_grapher.sh` wrapper script. It handles temporary directory isolation and Docker volume mounting.

### Basic Usage
```bash
./grapher/run_grapher.sh <input_file.cbl> <output_file.jsonl>
```

### Advanced Usage (For Large Programs)
If analyzing an extremely complex program, you can adjust the complexity cap and the memory limit:
```bash
./grapher/run_grapher.sh myprog.CBL output.jsonl --max-vms 2000 --memory 4096
```

*   `--max-vms`: The maximum number of concurrent branches to explore (default: 1000).
*   `--memory`: The RAM (in MB) allocated to the Node.js process (default: 1024).

---

## Output Format (JSONL)

The output is in **JSONL (JSON Lines)** format. Each line is a valid JSON object. This allows for processing massive graphs that would otherwise fail standard JSON parsing.

### 1. Program Header
```json
{"type":"program","name":"MYPROG","location":{"uri":"file:/data/input/myprog.cbl","start":{"line":1,"character":8},"end":{"line":200,"character":12}}}
```

### 2. Nodes
```json
{"id":10, "name":"MAIN-LOGIC", "kind":"paragraph", "location":{...}, "type":"node"}
{"id":15, "targetName":"SUBPROG", "type":"node", "kind":"call"}
```

### 3. Edges
```json
{"type":"edge", "from":10, "to":15}
```

---

## Visualization

You can convert the JSONL output to a Graphviz DOT file for visualization:

```bash
./grapher/json_to_graphviz.py output.jsonl output.dot
```

To generate a PNG image from the DOT file (requires Graphviz installed on your host):
```bash
dot -Tpng output.dot -o output.png
```

Visual Legend:
*   **White Box**: Standard Paragraph.
*   **Grey Box**: COBOL Section.
*   **Light Blue Component**: External Program Call (`CALL`).

---

## Debugging & Tracing

The tool provides a "Heartbeat" log every 5,000 execution steps:
`[Heartbeat] Steps: 50,000 | Active VMs: 12 | Total States: 240,100 | Heap: 850MB`

*   **Steps**: Total instructions processed.
*   **Active VMs**: Number of pending branches in the stack.
*   **Total States**: Number of unique execution fingerprints stored in the cache.
*   **Heap**: Current RAM usage of the analysis engine.
