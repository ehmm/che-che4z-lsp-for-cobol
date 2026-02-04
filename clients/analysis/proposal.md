# Proposal: COBOL Control Flow CLI (`ccf-cli`)

## Overview
This proposal outlines the creation of a TypeScript-based binary that provides a command-line interface for generating Control Flow Graphs (CFG) from preprocessed COBOL source files. The tool will leverage the existing `@code4z/analysis` logic and the established Java-based CFAST CLI to deliver a standalone JSON representation of a program's control flow.

## Objectives
- Provide a CLI for CI/CD pipelines and external tool integration.
- Export control flow data in a structured JSON format.
- Support "external call" nodes that act as boundaries for the analysis.
- Utilize existing project infrastructure for parsing and CFAST generation.

## Setup and Build Script
To ensure all components are available, a `setup.sh` script will be provided in the project root. This script will:
1.  **Build the Java Engine**: Run `mvn clean package` in the `server` directory to generate `server/engine/target/server.jar`.
2.  **Build the Analysis Package**: Run `npm install` and `npm run compile` in `clients/analysis`.
3.  **Link the CLI**: Use `npm link` to make the `ccf-cli` globally available during development.

## Proposed Architecture

### 1. Command Line Interface
The tool will be implemented as a Node.js binary within the `@code4z/analysis` package.

**Usage Example:**
```bash
ccf-cli -i MYPROG.cbl -o MYPROG.cfg.json --max-vms 1000
```

### 2. Processing Pipeline

#### Phase A: CFAST Generation
The CLI will utilize the existing Java-based CLI available in the `server.jar`:
1.  **Execution**: Invoke `java -jar server/engine/target/server.jar cfast --source_folder <dir>`.
2.  **Output**: The Java CLI generates `.cfast.json` files for each `.cbl` file in the specified directory.
3.  **Consumption**: The `ccf-cli` will read these generated files for further processing.

#### Phase B: Analysis Engine Execution
Using the `@code4z/analysis` package:
1.  Load the `CFAST` JSON.
2.  Instantiate the `ControlFlowGraphBuilder`.
3.  Execute the virtual machine to explore execution paths.
4.  Apply optimizations to ensure termination.

#### Phase C: External Call Handling
To meet the requirement of stopping at external calls:
- **CFAST Extension**: Update the CFAST model to include a `call` node type.
- **Instruction Mapping**: The `listing.ts` component will map COBOL `CALL` statements to a `CallInstruction`.
- **VM Logic**: The `CallInstruction` will be treated as a single node in the graph. The Virtual Machine will not attempt to branch or step "into" the called program, effectively treating it as a black box that returns control to the next instruction.
- **JSON Output**: The resulting node in the output JSON will contain metadata identifying it as an external call and providing the target program name.

#### Phase D: Serialization
The finalized `Graph` object will be normalized into a `GraphDTO` and written to the output file.

## JSON Schema Requirements
The exported JSON should include:
- **Nodes**:
    - `id`: Unique identifier.
    - `type`: `program`, `section`, `paragraph`, or `external_call`.
    - `name`: Identifier of the node (e.g., paragraph name or called program name).
    - `location`: File URI and range (line/character).
- **Edges**:
    - `from`: Source node ID.
    - `to`: Target node ID.
    - `metadata`: (Optional) Transition type (e.g., `fall-thru`, `goto`, `perform`).

## Detailed Implementation Steps

1.  **Setup Script**: Create `setup-ccf.sh` in the root directory.
2.  **Package Setup**:
    - Add `bin/ccf-cli.ts` to `@code4z/analysis`.
    - Configure `package.json` with a `bin` entry.
3.  **Java Integration**:
    - Implement a TypeScript wrapper to execute the `java -jar ... cfast` command.
4.  **Model Updates**:
    - Update `src/model/cfast.ts` to include the `call` type.
    - Update `src/vm/instructions.ts` and `src/vm/listing.ts` to handle the new type.
5.  **CLI Core**:
    - Implement argument parsing and error handling.
    - Add logging for diagnostic purposes (e.g., VM limit reached).
6.  **Testing**:
    - Integration tests using sample COBOL files from the `tests/test_files` directory.

## Benefits
- **Efficiency**: Reuses the high-performance Java parser.
- **Maintainability**: Centralizes control flow logic in the TypeScript package while keeping the heavy-lifting parsing in Java.
- **Accessibility**: Easy setup via a single script.
