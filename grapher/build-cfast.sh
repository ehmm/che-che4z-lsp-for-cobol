#!/bin/bash
set -e

# Get the absolute path to the project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." &> /dev/null && pwd )"

echo "Building Java CFAST CLI..."

cd "$PROJECT_ROOT/server"
# Using -DskipTests for speed as per common developer workflow
mvn clean package -DskipTests --no-transfer-progress

if [ -f "engine/target/server.jar" ]; then
    echo "CFAST CLI built successfully."
    echo "Executable jar: $PROJECT_ROOT/server/engine/target/server.jar"
else
    echo "Error: server.jar not found in engine/target/"
    exit 1
fi
