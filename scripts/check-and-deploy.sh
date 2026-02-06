#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the project root (parent of scripts directory)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root to ensure docker-compose works correctly
cd "$PROJECT_ROOT" || exit 1

# Function to calculate checksum with fallback options
calculate_checksum() {
    local file="$1"
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$file" | cut -d' ' -f1
    elif command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$file" | cut -d' ' -f1
    elif command -v openssl >/dev/null 2>&1; then
        openssl sha256 "$file" | awk '{print $2}'
    else
        echo "Error: No checksum command available (tried shasum, sha256sum, openssl)" >&2
        return 1
    fi
}

# Path to the etags.json file
ETAGS_FILE="scripts/etags.json"
# Path to store the hash
HASH_FILE="scripts/.etags.hash"
# Path to store the git reference
GIT_REF_FILE="scripts/.git.ref"

# make sure we are up to date
git pull >/dev/null 2>&1
/home/lauri/.cargo/bin/uv run scripts/admin.py download >/dev/null 2>&1

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
CURRENT_HASH=$(calculate_checksum "$ETAGS_FILE")
if [ $? -ne 0 ]; then
    exit 1
fi

# Get current git reference
CURRENT_GIT_REF=$(git rev-parse HEAD)

# Read stored hash if it exists
if [ -f "$HASH_FILE" ]; then
    STORED_HASH=$(cat "$HASH_FILE")
else
    STORED_HASH=""
fi

# Read stored git reference if it exists
if [ -f "$GIT_REF_FILE" ]; then
    STORED_GIT_REF=$(cat "$GIT_REF_FILE")
else
    STORED_GIT_REF=""
fi

if [ $VERBOSE -eq 1 ]; then
    echo "Current hash: $CURRENT_HASH"
    echo "Stored hash:  $STORED_HASH"
    echo "Current git ref: $CURRENT_GIT_REF"
# Run docker-compose
if /usr/local/bin/docker-compose up -d --build >/dev/null 2>&1; then
    [ $VERBOSE -eq 1 ] && echo "Docker Compose started successfully"
    # Store the new hash and git reference
    echo "$CURRENT_HASH" > "$HASH_FILE"
    echo "$CURRENT_GIT_REF" > "$GIT_REF_FILE"
    [ $VERBOSE -eq 1 ] && echo "New hash and git reference stored"
    exit 0
else
    echo "Error: docker-compose command failed"
    exit 1
fi
if [ "$CURRENT_GIT_REF" != "$STORED_GIT_REF" ]; then
    GIT_REF_CHANGED=1
fi

if [ $ETAGS_CHANGED -eq 0 ] && [ $GIT_REF_CHANGED -eq 0 ]; then
    [ $VERBOSE -eq 1 ] && echo "No changes detected in $ETAGS_FILE or git reference"
    exit 0
fi

if [ $VERBOSE -eq 1 ]; then
    [ $ETAGS_CHANGED -eq 1 ] && echo "Changes detected in $ETAGS_FILE"
    [ $GIT_REF_CHANGED -eq 1 ] && echo "Changes detected in git reference"
fi
[ $VERBOSE -eq 1 ] && echo "Running docker-compose up -d --build..."

# Run docker-compose
if /usr/local/bin/docker-compose up -d --build >/dev/null 2>&1; then
    [ $VERBOSE -eq 1 ] && echo "Docker Compose started successfully"
    # Store the new hash
    echo "$CURRENT_HASH" > "$HASH_FILE"
    [ $VERBOSE -eq 1 ] && echo "New hash stored"
    exit 0
else
    echo "Error: docker-compose command failed"
    exit 1
fi
