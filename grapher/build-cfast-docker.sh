#!/bin/bash
set -e

# Get the absolute path to the project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." &> /dev/null && pwd )"

echo "Building Docker image 'grapher:latest' (Node + Java Multi-stage)..."
cd "$PROJECT_ROOT"
docker build -t grapher:latest -f grapher/Dockerfile .

echo "Verifying the Docker image..."
docker run --rm grapher:latest --help

echo "Testing 'ccf-cli' on a sample file..."
# Use a sample from the engine test resources
SAMPLE_PATH="server/engine/src/test/resources/cfast/case1.cbl"
SAMPLE_FILE="case1.cbl"
SAMPLE_DIR=$(dirname "$SAMPLE_PATH")

# Create a temporary output directory
mkdir -p "$PROJECT_ROOT/grapher/output"

docker run --rm \
  -v "$PROJECT_ROOT/$SAMPLE_DIR":/data/input \
  -v "$PROJECT_ROOT/grapher/output":/data/output \
  -e JAVA_JAR_PATH=/app/bin/server.jar \
  grapher:latest -i /data/input/$SAMPLE_FILE -o /data/output/case1.cfg.json

if [ -f "$PROJECT_ROOT/grapher/output/case1.cfg.json" ]; then
    echo "Successfully generated CFG for case1.cbl"
    echo "Output saved to grapher/output/case1.cfg.json"
    # Show a snippet of the output
    head -n 20 "$PROJECT_ROOT/grapher/output/case1.cfg.json"
else
    echo "Failed to generate CFG for case1.cbl"
    exit 1
fi

echo "Docker image 'ccf-cli' built and verified successfully."
