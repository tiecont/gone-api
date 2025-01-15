#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Ensure the main binary is built before running
if [ ! -f "./main" ]; then
  echo "Binary './main' not found. Building it now..."
  go build -buildvcs=false -o ./main .
fi

# Start the application with Air
echo "Starting the application with Air..."
air
