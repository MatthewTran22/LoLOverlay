#!/bin/bash

# GhostDraft Production Build Script
# This script builds the desktop app with Turso credentials embedded

# Load environment variables from .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Required variables
TURSO_DATABASE_URL="${TURSO_DATABASE_URL:-}"
TURSO_AUTH_TOKEN="${TURSO_AUTH_TOKEN:-}"

# Validate required variables
if [ -z "$TURSO_DATABASE_URL" ]; then
    echo "Error: TURSO_DATABASE_URL is required"
    echo "Set it in .env or as an environment variable"
    exit 1
fi

if [ -z "$TURSO_AUTH_TOKEN" ]; then
    echo "Error: TURSO_AUTH_TOKEN is required"
    echo "Set it in .env or as an environment variable"
    exit 1
fi

echo "Building GhostDraft..."
echo "  Turso URL: $TURSO_DATABASE_URL"
echo "  Auth Token: ${TURSO_AUTH_TOKEN:0:8}..."

# Build with ldflags
wails build -ldflags "\
-X 'ghostdraft/internal/data.TursoURL=$TURSO_DATABASE_URL' \
-X 'ghostdraft/internal/data.TursoAuthToken=$TURSO_AUTH_TOKEN'"

if [ $? -eq 0 ]; then
    echo "Build complete! Output: build/bin/GhostDraft.exe"
else
    echo "Build failed!"
    exit 1
fi
