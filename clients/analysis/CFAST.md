# CFAST (Control Flow Abstract Syntax Tree)

CFAST is a simplified, intermediate representation of a COBOL program's `PROCEDURE DIVISION`. It is specifically designed to filter out data definitions and non-executable logic, focusing entirely on nodes that influence the program's execution path.

## Role in Analysis
The `@code4z/analysis` package does not parse COBOL directly. Instead:
1.  **Java Engine**: Parses the raw COBOL and generates a `CFAST` JSON.
2.  **TypeScript VM**: Consumes this JSON and "runs" it using Virtual Machines to build the final Control Flow Graph.

## Base Node Structure
Every node in the CFAST tree adheres to this base structure:

```json
{
  "type": "string",
  "location": {
    "uri": "string",
    "start": { "line": 1, "character": 1 },
    "end": { "line": 1, "character": 10 }
  },
  "children": [],
  "snippet": "string (optional)"
}
```

*   **type**: The kind of control flow element (see [Node Types](#node-types)).
*   **location**: 1-based coordinates pointing back to the original source.
*   **children**: Nested nodes (e.g., statements inside an `IF` block).
*   **snippet**: A string containing the original source text for this node.

## Node Types

### 1. Structural Units
*   **program**: The root node. Contains the `name` of the program.
*   **section**: A COBOL `SECTION`. Contains the `name`.
*   **paragraph**: A COBOL `PARAGRAPH`. Contains the `name`.

### 2. Branching & Conditionals
*   **if**: Represents the start of an `IF`. Accompanied by an `endif` and optional `else` sibling/child nodes.
*   **evaluate**: Represents a `EVALUATE` block. Contains `when` and `whenother` nodes.
*   **atEnd**: Represents the `AT END` branch of I/O statements.

### 3. Transfer of Control
*   **perform**: An out-of-line `PERFORM`. Includes metadata:
    *   `targetName`: The paragraph/section to run.
    *   `thruName`: The optional end paragraph of the range.
    *   `performUntilType`: `UNTIL_EXIT` or `UNTIL_CONDITION`.
*   **inlineperform**: An inline loop.
*   **goto**: A `GO TO` statement. Contains a list of `targetName` values.
*   **call**: An external program call. Treated as a single node in the graph (does not branch).

### 4. Termination
*   **stop**: A `STOP RUN` statement.
*   **goback**: A `GOBACK` statement.
*   **exit**: An `EXIT` statement.
*   **exitparagraph** / **exitsection**: Explicit exits from a unit.

### 5. Specialized Blocks
*   **execsql** / **execwhenever**: SQL logic.
*   **execcics**: CICS logic.
*   **xmlparse**: XML processing logic.
*   **alter**: The (deprecated) `ALTER ... TO PROCEED TO` logic.

## How CFAST is Built

The tree is built by the Java-based `CFASTBuilderImpl` class. 

### Process:
1.  **Parse**: The full COBOL source is parsed into a comprehensive AST.
2.  **Traverse**: The builder performs a recursive traversal of the `PROCEDURE DIVISION`.
3.  **Filter**: It ignores most "simple" statements (like `MOVE`, `ADD`) unless they are part of a larger control structure.
4.  **Simplify**: Complex statements are mapped to their CFAST equivalents (e.g., mapping both `IF` and `EVALUATE` to a consistent branching pattern).
5.  **Snippet Extraction**: For each relevant node, the builder re-reads the source file to extract a 10-line snippet for debugging and visualization.

## Memory Optimizations
In the standalone CLI mode, we perform **Snippet Stripping** on the CFAST JSON before it reaches the analysis engine. This removes the `snippet` field from every node, which significantly reduces the RAM required to hold the tree in memory for large programs.
