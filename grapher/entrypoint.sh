#!/bin/bash
set -e

# Usage: ccf-cli -i <input.cbl> -o <output.jsonl> [--max-vms <number>]

# Parse arguments
MAX_VMS="1000"
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -i|--input) INPUT_PATH="$2"; shift ;;
        -o|--output) OUTPUT_PATH="$2"; shift ;;
        --max-vms) MAX_VMS="$2"; shift ;;
        -h|--help) 
            echo "Usage: ccf-cli -i <input.cbl> -o <output.jsonl> [--max-vms <number>]"
            exit 0 
            ;;
        *) echo "Unknown parameter: $1" ;;
    esac
    shift
done

if [ -z "$INPUT_PATH" ] || [ -z "$OUTPUT_PATH" ]; then
    echo "Usage: ccf-cli -i <input.cbl> -o <output.jsonl> [--max-vms <number>]"
    exit 1
fi

JAR_PATH=${JAVA_JAR_PATH:-"/app/bin/server.jar"}
GO_BIN="/app/bin/ccf-cli-go"
TS_BIN="/app/analysis/lib/bin/ccf-cli.js"

INPUT_DIR=$(dirname "$INPUT_PATH")
INPUT_FILE=$(basename "$INPUT_PATH")
# Java CLI generates .cfast.json for all .cbl files in the folder
FILENAME=$(basename "$INPUT_FILE" | sed 's/\.[^.]*$//')
CFAST_PATH="$INPUT_DIR/$FILENAME.cfast.json"

echo "Step 1: Generating CFAST using Java Engine..."
java -jar "$JAR_PATH" cfast --source_folder "$INPUT_DIR"

if [ ! -f "$CFAST_PATH" ]; then
    echo "Error: CFAST generation failed. Expected $CFAST_PATH"
    exit 1
fi

CFAST_SIZE=$(du -h "$CFAST_PATH" | cut -f1)
echo "CFAST generated: $CFAST_PATH ($CFAST_SIZE)"

export CFAST_ALREADY_GENERATED="true"

if [ "$USE_GO_ANALYSIS" = "true" ]; then
    echo "Step 2: Analyzing Control Flow using Go Implementation..."
    "$GO_BIN" -i "$CFAST_PATH" -o "$OUTPUT_PATH" --max-vms "$MAX_VMS"
else
    echo "Step 2: Analyzing Control Flow using TypeScript Implementation..."
    # We call the JS file directly with node. 
    # Note: The TS implementation in ccf-cli.ts also tries to run Phase A, 
    # so we might need a "bypass" or just let it overwrite. 
    # For now, we'll let it use its internal logic but prefer the entrypoint's orchestration.
    node "$TS_BIN" -i "$INPUT_PATH" -o "$OUTPUT_PATH" --max-vms "$MAX_VMS"
fi

echo "Process Complete."
