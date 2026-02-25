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
