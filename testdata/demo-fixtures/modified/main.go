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
