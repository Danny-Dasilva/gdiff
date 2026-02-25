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
