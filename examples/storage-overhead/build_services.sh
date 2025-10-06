#!/bin/bash

# This script compiles the ft-executor and ft-timeoracle binaries
# from their source directories and moves them into the current directory.

# Exit immediately if a command exits with a non-zero status.
set -e

# Get the root directory of the project
# This assumes the script is run from somewhere inside the project
PROJECT_ROOT=$(git rev-parse --show-toplevel)

echo "Project root found at: $PROJECT_ROOT"
echo "---"

# --- Compile ft-timeoracle ---
echo "Compiling ft-timeoracle..."
go build -o ft-timeoracle "$PROJECT_ROOT/ft-timeoracle/"
echo "✓ ft-timeoracle compiled successfully."
echo "---"


# --- Compile ft-executor ---
echo "Compiling ft-executor..."
go build -o ft-executor "$PROJECT_ROOT/ft-executor/"
echo "✓ ft-executor compiled successfully."
echo "---"

echo "✅ All binaries have been compiled and are available in the current directory."
