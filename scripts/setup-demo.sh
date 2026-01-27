#!/bin/bash
# Setup test changes for VHS demo recordings
# Run this before recording demo tapes

set -e

DEMO_DIR="/tmp/gdiff-demo"
GDIFF_BIN="/home/danny/Documents/personal/git_diff_tool/gdiff"

# Clean up any existing demo directory
rm -rf "$DEMO_DIR"

# Create fresh demo repo
mkdir -p "$DEMO_DIR"
cd "$DEMO_DIR"
git init
git config user.email "demo@example.com"
git config user.name "Demo User"

# Create initial files
cat > main.go << 'EOF'
package main

import (
    "fmt"
    "os"
)

func main() {
    name := "World"
    if len(os.Args) > 1 {
        name = os.Args[1]
    }
    greeting := getGreeting(name)
    fmt.Println(greeting)
}

func getGreeting(name string) string {
    return "Hello, " + name + "!"
}

func unused() {
    // This function is not used
}
EOF

cat > config.go << 'EOF'
package main

// Config holds application configuration
type Config struct {
    Debug   bool
    Verbose bool
    Output  string
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
    return Config{
        Debug:   false,
        Verbose: false,
        Output:  "stdout",
    }
}
EOF

cat > utils.go << 'EOF'
package main

import "strings"

// ToUpper converts string to uppercase
func ToUpper(s string) string {
    return strings.ToUpper(s)
}

// ToLower converts string to lowercase
func ToLower(s string) string {
    return strings.ToLower(s)
}
EOF

cat > README.md << 'EOF'
# Demo Project

A simple demo project for gdiff.

## Usage

```bash
go run . [name]
```

## Features

- Greeting message
- Command line arguments
EOF

# Commit initial state
git add -A
git commit -m "Initial commit: basic project structure"

# Now make various changes for the demo

# 1. Modify main.go - multiple hunks
cat > main.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "strings"
)

const version = "1.0.0"

func main() {
    name := "World"
    if len(os.Args) > 1 {
        name = strings.TrimSpace(os.Args[1])
    }
    greeting := getGreeting(name)
    fmt.Println(greeting)
    fmt.Printf("Version: %s\n", version)
}

func getGreeting(name string) string {
    if name == "" {
        name = "Anonymous"
    }
    return "Hello, " + name + "!"
}

// Removed unused function
EOF

# 2. Modify config.go - add new field
cat > config.go << 'EOF'
package main

// Config holds application configuration
type Config struct {
    Debug    bool
    Verbose  bool
    Output   string
    LogLevel int
    Timeout  int
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
    return Config{
        Debug:    false,
        Verbose:  false,
        Output:   "stdout",
        LogLevel: 1,
        Timeout:  30,
    }
}

// NewConfig creates a new configuration with custom values
func NewConfig(debug bool, output string) Config {
    cfg := DefaultConfig()
    cfg.Debug = debug
    cfg.Output = output
    return cfg
}
EOF

# 3. Add a new file
cat > logger.go << 'EOF'
package main

import (
    "fmt"
    "time"
)

// Logger handles application logging
type Logger struct {
    prefix string
    level  int
}

// NewLogger creates a new logger instance
func NewLogger(prefix string) *Logger {
    return &Logger{
        prefix: prefix,
        level:  1,
    }
}

// Log outputs a message with timestamp
func (l *Logger) Log(msg string) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    fmt.Printf("[%s] %s: %s\n", timestamp, l.prefix, msg)
}

// Debug outputs a debug message
func (l *Logger) Debug(msg string) {
    if l.level > 0 {
        l.Log("DEBUG: " + msg)
    }
}
EOF

# 4. Modify utils.go - small inline change
cat > utils.go << 'EOF'
package main

import "strings"

// ToUpper converts string to uppercase (trimmed)
func ToUpper(s string) string {
    return strings.ToUpper(strings.TrimSpace(s))
}

// ToLower converts string to lowercase (trimmed)
func ToLower(s string) string {
    return strings.ToLower(strings.TrimSpace(s))
}

// Capitalize capitalizes the first letter
func Capitalize(s string) string {
    if len(s) == 0 {
        return s
    }
    return strings.ToUpper(s[:1]) + s[1:]
}
EOF

# 5. Modify README
cat > README.md << 'EOF'
# Demo Project

A simple demo project for gdiff TUI.

## Installation

```bash
go build -o demo .
```

## Usage

```bash
./demo [name]
```

## Features

- Greeting message with version
- Command line arguments
- Configurable logging
- Utility functions

## Version

Current version: 1.0.0
EOF

echo "Demo repository created at: $DEMO_DIR"
echo "Files with changes:"
cd "$DEMO_DIR" && git status --short

# Copy gdiff binary to /tmp for easy access
cp "$GDIFF_BIN" /tmp/gdiff
chmod +x /tmp/gdiff

echo ""
echo "gdiff binary copied to /tmp/gdiff"
echo "Ready for VHS recording!"
