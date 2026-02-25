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
