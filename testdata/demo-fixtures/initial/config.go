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
