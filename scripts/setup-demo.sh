#!/bin/bash
# Setup test changes for VHS demo recordings
# Uses version-controlled fixture files instead of heredocs
# Usage: ./scripts/setup-demo.sh [output-dir]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
FIXTURES="$PROJECT_DIR/testdata/demo-fixtures"
DEMO_DIR="${1:-/tmp/gdiff-demo}"

# Deterministic timestamps for reproducible commits
export GIT_COMMITTER_DATE="2025-01-15T10:00:00+00:00"
export GIT_AUTHOR_DATE="2025-01-15T10:00:00+00:00"

# Clean up any existing demo directory
rm -rf "$DEMO_DIR"

# Create fresh demo repo
mkdir -p "$DEMO_DIR"
cd "$DEMO_DIR"
git init
git config user.email "demo@example.com"
git config user.name "Demo User"

# Create initial files from fixtures
cp "$FIXTURES"/initial/* .

# Commit initial state
git add -A
git commit -m "Initial commit: basic project structure"

# Apply modifications from fixtures
cp "$FIXTURES"/modified/* .

echo "Demo repository created at: $DEMO_DIR"
echo "Files with changes:"
git status --short
