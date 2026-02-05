#!/bin/bash
set -e

# Help message
if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <input_cobol_file> <output_json_path> [--max-vms <number>] [--memory <MB>] [--use-go]"
    exit 1
fi

# Resolve absolute paths
INPUT_PATH=$(realpath "$1")
OUTPUT_PATH=$(realpath "$2")
shift 2

MAX_VMS="1000"
MEMORY="2048"
USE_GO=${USE_GO:-"false"}

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --max-vms) MAX_VMS="$2"; shift ;;
        --memory) MEMORY="$2"; shift ;;
        --use-go) USE_GO="true" ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

if [ ! -f "$INPUT_PATH" ]; then
    echo "Error: Input file $INPUT_PATH not found."
    exit 1
fi

INPUT_FILE=$(basename "$INPUT_PATH")

# Create temporary directories for isolated input and output
TEMP_IN_DIR=$(mktemp -d)
TEMP_OUT_DIR=$(mktemp -d)

# Ensure cleanup of the temporary directories on exit
trap 'rm -rf "$TEMP_IN_DIR" "$TEMP_OUT_DIR"' EXIT

# Copy the specific input file to the isolated input directory
cp "$INPUT_PATH" "$TEMP_IN_DIR/"

echo "Starting Control Flow Analysis for $INPUT_FILE (Use Go: $USE_GO)..."

# Run the Docker container
docker run --rm \
  -v "$TEMP_IN_DIR":/data/input \
  -v "$TEMP_OUT_DIR":/data/output \
  -e NODE_OPTIONS="--max-old-space-size=$MEMORY" \
  -e USE_GO_ANALYSIS="$USE_GO" \
  grapher:latest -i "/data/input/$INPUT_FILE" -o "/data/output/result.json" --max-vms "$MAX_VMS"

# Check if the output was generated and copy it to the target path
if [ -f "$TEMP_OUT_DIR/result.json" ]; then
    mkdir -p "$(dirname "$OUTPUT_PATH")"
    cp "$TEMP_OUT_DIR/result.json" "$OUTPUT_PATH"
    echo "Analysis complete. Output saved to: $OUTPUT_PATH"
else
    echo "Error: Analysis failed to produce output."
    exit 1
fi
