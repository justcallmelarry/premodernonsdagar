#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the project root (parent of scripts directory)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root to ensure docker-compose works correctly
cd "$PROJECT_ROOT" || exit 1

# Path to the etags.json file
ETAGS_FILE="scripts/etags.json"
# Path to store the hash
HASH_FILE="scripts/.etags.hash"

# Parse command line arguments
VERBOSE=0
if [ "$1" = "-v" ] || [ "$1" = "--verbose" ]; then
    VERBOSE=1
fi

# Check if etags.json exists
if [ ! -f "$ETAGS_FILE" ]; then
    echo "Error: $ETAGS_FILE not found"
    exit 1
fi

# Calculate current hash of etags.json
CURRENT_HASH=$(shasum -a 256 "$ETAGS_FILE" | cut -d ' ' -f 1)

# Read stored hash if it exists
if [ -f "$HASH_FILE" ]; then
    STORED_HASH=$(cat "$HASH_FILE")
else
    STORED_HASH=""
fi

if [ $VERBOSE -eq 1 ]; then
    echo "Current hash: $CURRENT_HASH"
    echo "Stored hash:  $STORED_HASH"
fi

# Compare hashes
if [ "$CURRENT_HASH" = "$STORED_HASH" ]; then
    [ $VERBOSE -eq 1 ] && echo "No changes detected in $ETAGS_FILE"
    exit 0
fi

[ $VERBOSE -eq 1 ] && echo "Changes detected in $ETAGS_FILE"
[ $VERBOSE -eq 1 ] && echo "Running docker-compose up -d..."

# Run docker-compose
if docker-compose up -d; then
    [ $VERBOSE -eq 1 ] && echo "Docker Compose started successfully"
    # Store the new hash
    echo "$CURRENT_HASH" > "$HASH_FILE"
    [ $VERBOSE -eq 1 ] && echo "New hash stored"
    exit 0
else
    echo "Error: docker-compose command failed"
    exit 1
fi
