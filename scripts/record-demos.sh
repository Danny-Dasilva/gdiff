#!/bin/bash
# Record all VHS demo tapes
# Usage: ./scripts/record-demos.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# Reset demo repo before each recording
reset_demo() {
    echo "Resetting demo repository..."
    "$SCRIPT_DIR/setup-demo.sh" > /dev/null 2>&1
}

# Record a single tape
record_tape() {
    local tape=$1
    echo "Recording $tape..."
    reset_demo
    PATH="/tmp:$PATH" vhs "$tape"
    echo "Done: ${tape%.tape}.gif"
}

# Record all tapes
echo "=== Recording gdiff Demo Tapes ==="
echo ""

record_tape "demo.tape"
record_tape "demo-navigation.tape"
record_tape "demo-staging.tape"
record_tape "demo-character-diff.tape"
record_tape "demo-commit.tape"

echo ""
echo "=== All demos recorded! ==="
echo "Generated GIFs:"
ls -la *.gif
